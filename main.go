// Package main is the entry point for the k6 CLI application. It assembles all the crucial components for the running.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/discue/go-syscall-gatekeeper/app/runtime"
	"github.com/discue/go-syscall-gatekeeper/app/uroot"
)

var osExit = os.Exit // Assign exit to a variable to allow mocking in unit tests
func exit(code int) {
	osExit(code)
}

type syscallDeniedAction string

const (
	KillAction  syscallDeniedAction = "kill"
	ErrorAction syscallDeniedAction = "error"
)

func (s *syscallDeniedAction) String() string {
	return string(*s)
}

func (s *syscallDeniedAction) Set(value string) error {
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

func main() {
	mainCtx, cancel := context.WithCancel(context.Background())
	println(fmt.Sprintf("gatekeeper started with %#+v", os.Args))

	tracee := startTracee(mainCtx)
	waitForShutdown(cancel, tracee)
}

func startTracee(c context.Context) context.Context {
	args := configureAndParseArgs()
	program := args[0]
	programArgs := args[1:]
	println(fmt.Sprintf("starting %s with args %#+v", program, programArgs))

	_, exitContext, err := uroot.Exec(c, program, programArgs)
	if err != nil {
		fmt.Println(err.Error())
	}

	return exitContext
}

func configureAndParseArgs() []string {
	conf := runtime.Get()

	runCmd := flag.NewFlagSet("mode", flag.ExitOnError)
	triggerEnforceOnLogMatch := runCmd.String("trigger-enforce-on-log-match", "", "a string")
	triggerEnforceOnSignal := runCmd.String("trigger-enforce-on-signal", "", "a string")
	verbose := runCmd.Bool("verbose", false, "a bool")

	// permissions
	allowFileSystemReadAccess := runCmd.Bool("allow-file-system-read", false, "a bool")
	allowFileSystemWriteAccess := runCmd.Bool("allow-file-system-write", false, "a bool")
	allowFileSystemAccess := runCmd.Bool("allow-file-system", false, "alias for --allow-file-system-write")
	allowFileSystemPermissionsAccess := runCmd.Bool("allow-file-system-permissions", false, "a bool")

	allowNetworkClient := runCmd.Bool("allow-network-client", false, "a bool")
	allowNetworkServer := runCmd.Bool("allow-network-server", false, "a bool")
	allowProcessManagement := runCmd.Bool("allow-process-management", false, "a bool")
	allowNetworking := runCmd.Bool("allow-networking", false, "a bool")
	allowMemoryManagement := runCmd.Bool("allow-memory-management", false, "a bool")
	allowSignals := runCmd.Bool("allow-signals", false, "a bool")
	allowTimersAndClocksManagement := runCmd.Bool("allow-timers-and-clocks-management", false, "a bool")
	allowSecurityAndPermissions := runCmd.Bool("allow-security-and-permissions", false, "a bool")
	allowSystemInformation := runCmd.Bool("allow-system-information", false, "a bool")
	allowProcessCommunication := runCmd.Bool("allow-process-communication", false, "a bool")
	allowProcessSynchronization := runCmd.Bool("allow-process-synchronization", false, "a bool")
	allowMisc := runCmd.Bool("allow-misc", false, "a bool")
	noEnforceOnStart := runCmd.Bool("no-enforce-on-startup", false, "a bool")
	noImplicitAllow := runCmd.Bool("no-implicit-allow", false, "a bool")
	var action syscallDeniedAction // Custom flag variable
	runCmd.Var(&action, "on-syscall-denied", "Action to take when a syscall is denied: 'kill' or 'error'")

	if len(os.Args) < 3 {
		fmt.Println("Error: You did not provide enough parameters and flags.")
		runCmd.Usage()
		exit(100)
	}

	mode := os.Args[1]

	// parse flags now
	err := runCmd.Parse(os.Args[2:])
	if err != nil {
		fmt.Println(err.Error())
		runCmd.Usage()
		exit(100)
	}

	allowList := runtime.NewSyscallAllowList()

	if *allowFileSystemWriteAccess || *allowFileSystemAccess {
		allowList.AllowAllFileSystemWriteAccess()
		allowList.AllowAllFileSystemReadAccess()
		allowList.AllowAllFileDescriptors()
		runtime.Get().FileSystemAllowRead = true
		runtime.Get().FileSystemAllowWrite = true
	} else if *allowFileSystemReadAccess {
		allowList.AllowAllFileSystemReadAccess()
		allowList.AllowAllFileDescriptors()
		runtime.Get().FileSystemAllowRead = true
		runtime.Get().FileSystemAllowWrite = false
	}

	if *allowFileSystemPermissionsAccess {
		allowList.AllowAllFilePermissions()
	}

	if *allowProcessManagement {
		allowList.AllowProcessManagement()
	}

	if *allowNetworkClient {
		allowList.AllowNetworkClient()
		allowList.AllowAllFileDescriptors()
		conf.NetworkAllowClient = true
	}

	if *allowNetworkServer {
		allowList.AllowNetworkServer()
		allowList.AllowAllFileDescriptors()
		conf.NetworkAllowServer = true
	}

	if *allowNetworking {
		allowList.AllowNetworking()
	}

	if *allowMemoryManagement {
		allowList.AllowMemoryManagement()
	}

	if *allowSignals {
		allowList.AllowSignals()
	}

	if *allowTimersAndClocksManagement {
		allowList.AllowTimersAndClocksManagement()
	}

	if *allowSecurityAndPermissions {
		allowList.AllowSecurityAndPermissions()
	}

	if *allowSystemInformation {
		allowList.AllowSystemInformation()
	}

	if *allowProcessCommunication {
		allowList.AllowProcessCommunication()
	}

	if *allowProcessSynchronization {
		allowList.AllowProcessSynchronization()
	}

	if *allowMisc {
		allowList.AllowMisc()
	}

	if !*noImplicitAllow {
		allowList.AllowProcessManagement()
		allowList.AllowMemoryManagement()
		allowList.AllowProcessSynchronization()
		allowList.AllowSignals()
		allowList.AllowMisc()
		allowList.AllowSecurityAndPermissions()
		allowList.AllowSystemInformation()
	}

	conf.VerboseLog = *verbose

	if *noEnforceOnStart {
		if *triggerEnforceOnLogMatch == "" && *triggerEnforceOnSignal == "" {
			fmt.Println("Error: To delay the enforcement of seccomp policies, please also specify either --trigger-enforce-on-log-match or --trigger-enforce-on-signal.")
			runCmd.Usage()
			exit(100)
		} else {
			conf.EnforceOnStartup = false
		}
	} else {
		conf.EnforceOnStartup = true
	}

	if *triggerEnforceOnLogMatch != "" {
		conf.TriggerEnforceLogMatch = *triggerEnforceOnLogMatch
	} else if *triggerEnforceOnSignal != "" {
		conf.TriggerEnforceSignal = *triggerEnforceOnSignal
	}

	if len(allowList.Syscalls) > 0 {
		conf.SyscallsAllowList = allowList.Syscalls
		conf.SyscallsAllowMap = runtime.CreateSyscallAllowMap(conf.SyscallsAllowList)
	}

	if action == ErrorAction {
		conf.SyscallsKillTargetIfNotAllowed = false
		conf.SyscallsDenyTargetIfNotAllowed = true
	} else {
		conf.SyscallsKillTargetIfNotAllowed = true
		conf.SyscallsDenyTargetIfNotAllowed = false
	}

	switch mode {
	case "trace":
		conf.ExecutionMode = runtime.EXECUTION_MODE_TRACE
	case "run":
		conf.ExecutionMode = runtime.EXECUTION_MODE_RUN
	default:
		runCmd.Usage()
		exit(100)
	}

	return runCmd.Args()
}

func waitForShutdown(cancel context.CancelFunc, tracee context.Context) {
	signal, stop := signal.NotifyContext(context.Background(), syscall.SIGKILL, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-signal.Done():
		println("Received signal from outside")
		break
	case <-tracee.Done():
		println("Tracee got cancelled")
		break
	}

	cancel()
	time.Sleep(1 * time.Second)

	stop()
	<-tracee.Done()

	// collect exit code of tracee
	traceeCancelCause := context.Cause(tracee)
	e := &uroot.ExitEventError{}
	errors.As(traceeCancelCause, &e)

	exitCode := 0
	if e.Signal != "" {
		exitCode = 111
	} else if e.ExitCode != 0 {
		exitCode = e.ExitCode
	}

	println(fmt.Sprintf("Exiting with code %d", exitCode))
	// exit
	exit(exitCode)
}
