package cli

import (
	"flag"
	"fmt"
	"strings"

	sec "github.com/seccomp/libseccomp-golang"
)

// stringSlice is a repeatable flag type that collects each flag occurrence
// into a slice. Implements flag.Value.
type stringSlice []string

func (s *stringSlice) String() string { return strings.Join([]string(*s), ",") }

func (s *stringSlice) Set(v string) error {
	if v == "" {
		return nil
	}
	*s = append(*s, strings.TrimSpace(v))
	return nil
}

// SyscallDeniedAction implements flag.Value for the on-syscall-denied flag.
type SyscallDeniedAction string

const (
	KillAction  SyscallDeniedAction = "kill"
	ErrorAction SyscallDeniedAction = "error"
)

func (s *SyscallDeniedAction) String() string {
	return string(*s)
}

func (s *SyscallDeniedAction) Set(value string) error {
	switch value {
	case "kill":
		*s = KillAction
	case "error":
		*s = ErrorAction
	default:
		return fmt.Errorf("invalid value for on-syscall-denied: %s.  Must be 'kill' or 'error'", value)
	}
	return nil
}

// Command groups the FlagSet and pointers to defined flags.
type Command struct {
	flagSet *flag.FlagSet

	// Triggers & verbosity
	TriggerEnforceOnLogMatch *string
	TriggerEnforceOnSignal   *string
	Verbose                  *bool

	// Permissions
	AllowFileSystemReadAccess        *bool
	AllowFileSystemWriteAccess       *bool
	AllowFileSystemAccess            *bool
	AllowFileSystemPermissionsAccess *bool
	// AllowFileSystemPath supports specifying the flag multiple times to
	// build a whitelist. Example: --allow-file-system-paths=/etc --allow-file-system-paths=/var
	AllowFileSystemPath *stringSlice
	// Derived list (synonym) populated after Parse()
	AllowFileSystemPathsList []string

	AllowNetworkClient             *bool
	AllowNetworkServer             *bool
	AllowNetworkLocalSockets       *bool
	AllowProcessManagement         *bool
	AllowNetworking                *bool
	AllowMemoryManagement          *bool
	AllowSignals                   *bool
	AllowTimersAndClocksManagement *bool
	AllowSecurityAndPermissions    *bool
	AllowSystemInformation         *bool
	AllowProcessCommunication      *bool
	AllowProcessSynchronization    *bool
	AllowMisc                      *bool

	EnforceOnStartup      *bool
	AllowImplicitCommands *bool

	Action SyscallDeniedAction
}

// NewCommand constructs the CLI FlagSet and returns a Command with pointers to all flags.
func NewCommand() *Command {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	c := &Command{flagSet: fs}

	// Triggers & verbosity
	c.TriggerEnforceOnLogMatch = fs.String("trigger-enforce-on-log-match", "", "Enable enforcement when trace output contains this string (use with -enforce-on-startup=false)")
	c.TriggerEnforceOnSignal = fs.String("trigger-enforce-on-signal", "", "Enable enforcement upon receiving this signal (name or number, use with -enforce-on-startup=false)")
	c.Verbose = fs.Bool("verbose", false, "Enable verbose decision logging from the tracer")

	// Permissions
	c.AllowFileSystemReadAccess = fs.Bool("allow-file-system-read", false, "Allow read-only filesystem access (open O_RDONLY, read, stat, list)")
	c.AllowFileSystemWriteAccess = fs.Bool("allow-file-system-write", false, "Allow modifying the filesystem (create, write, rename, unlink, truncate)")
	c.AllowFileSystemAccess = fs.Bool("allow-file-system", false, "Alias for --allow-file-system-write (full read/write filesystem access)")
	c.AllowFileSystemPermissionsAccess = fs.Bool("allow-file-system-permissions", false, "Allow changing file ownership and permissions (chmod/chown/fchmod/fchown*)")
	var allowFileSystemPaths stringSlice
	fs.Var(&allowFileSystemPaths, "allow-file-system-path", "Allow filesystem path (repeatable); example: --allow-file-system-path=/etc --allow-file-system-path=/var")
	c.AllowFileSystemPath = &allowFileSystemPaths // will be populated during Parse()

	c.AllowNetworkClient = fs.Bool("allow-network-client", false, "Allow outbound network connections (socket/connect/send/recv)")
	c.AllowNetworkServer = fs.Bool("allow-network-server", false, "Allow listening sockets and incoming connections (socket/bind/listen/accept)")
	c.AllowNetworkLocalSockets = fs.Bool("allow-network-local-sockets", false, "Allow local-only sockets (AF_UNIX, AF_NETLINK) for client use")
	c.AllowProcessManagement = fs.Bool("allow-process-management", false, "Allow process/thread creation and lifecycle control (exec/fork/clone/wait)")
	c.AllowNetworking = fs.Bool("allow-networking", false, "Allow both client and server networking capabilities")
	c.AllowMemoryManagement = fs.Bool("allow-memory-management", false, "Allow memory mapping and related syscalls (mmap/mprotect/mremap/brk)")
	c.AllowSignals = fs.Bool("allow-signals", false, "Allow setting and handling POSIX signals (rt_sig*, sigaltstack)")
	c.AllowTimersAndClocksManagement = fs.Bool("allow-timers-and-clocks-management", false, "Allow timers and clock syscalls (clock_gettime, timerfd_*, nanosleep)")
	c.AllowSecurityAndPermissions = fs.Bool("allow-security-and-permissions", false, "Allow identity/capability changes and seccomp (setuid/setgid/capset/seccomp). Risky; enable only if needed.")
	c.AllowSystemInformation = fs.Bool("allow-system-information", false, "Allow system information and rlimit operations (uname/sysinfo/getrlimit/setrlimit)")
	c.AllowProcessCommunication = fs.Bool("allow-process-communication", false, "Allow IPC mechanisms (SysV shared memory, semaphores, message queues, POSIX mqueue)")
	c.AllowProcessSynchronization = fs.Bool("allow-process-synchronization", false, "Allow synchronization primitives (futex/flock/robust list)")
	c.AllowMisc = fs.Bool("allow-misc", false, "Allow miscellaneous syscalls (includes ioctl, splice, vmsplice).")
	c.EnforceOnStartup = fs.Bool("enforce-on-startup", true, "Start with enforcement enabled on startup (default)")
	c.AllowImplicitCommands = fs.Bool("allow-implicit-commands", true, "Enable baseline implicit permissions; allow additional commands by default")

	// Custom action flag
	fs.Var(&c.Action, "on-syscall-denied", "Action when a syscall is denied: 'kill' (SIGKILL) or 'error' (simulate EPERM via SIGSYS)")

	return c
}

// PreScanDynamicSyscalls scans raw args for dynamic syscall flags and returns filtered args and validated syscall names.
// Supported forms:
//
//	--allow-syscall-<name>
//	--allow-syscall=<name>
func (c *Command) PreScanDynamicSyscalls(rawArgs []string) (filteredArgs []string, dynamicSyscalls []string) {
	filteredArgs = make([]string, 0, len(rawArgs))
	dynamicSyscalls = make([]string, 0)
	for _, a := range rawArgs {
		if len(a) >= 16 && a[:16] == "--allow-syscall-" { // fast path for prefix
			name := a[16:]
			if idx := indexByte(name, '='); idx >= 0 {
				name = name[:idx]
			}
			if name != "" {
				if _, err := sec.GetSyscallFromName(name); err == nil {
					dynamicSyscalls = append(dynamicSyscalls, name)
				}
			}
			continue
		}
		if len(a) >= 16 && a[:16] == "--allow-syscall=" {
			name := a[16:]
			if name != "" {
				if _, err := sec.GetSyscallFromName(name); err == nil {
					dynamicSyscalls = append(dynamicSyscalls, name)
				}
			}
			continue
		}
		filteredArgs = append(filteredArgs, a)
	}
	return
}

// indexByte is a small helper avoiding bytes import; returns -1 if not found.
func indexByte(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

// Parse parses CLI flags using the command's internal flagSet and derives helper fields.
func (c *Command) Parse(args []string) error {
	if err := c.flagSet.Parse(args); err != nil {
		return err
	}
	if c.AllowFileSystemPath != nil && len(*c.AllowFileSystemPath) > 0 {
		c.AllowFileSystemPathsList = make([]string, len(*c.AllowFileSystemPath))
		copy(c.AllowFileSystemPathsList, *c.AllowFileSystemPath)
	}
	return nil
}

// Args returns trailing non-flag arguments.
func (c *Command) Args() []string { return c.flagSet.Args() }

// Usage prints usage via the command's flagSet.
func (c *Command) Usage() { c.flagSet.Usage() }

// FlagSet returns the internal flag set (read-only access).
func (c *Command) FlagSet() *flag.FlagSet { return c.flagSet }
