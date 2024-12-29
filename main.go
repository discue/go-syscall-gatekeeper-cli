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

	tracee := startTracee(mainCtx)
	waitForShutdown(cancel, tracee)
}

func startTracee(c context.Context) context.Context {
	args := configureAndParseArgs()

	_, exitContext, err := uroot.Exec(c, args[0], args[1:])
	if err != nil {
		fmt.Println(err.Error())
	}

	return exitContext
}

func configureAndParseArgs() []string {
	conf := runtime.Get()
	mode := os.Args[1]

	runCmd := flag.NewFlagSet("mode", flag.ExitOnError)
	triggerEnforceOnLogMatch := runCmd.String("trigger-enforce-on-log-match", "", "a string")
	triggerEnforceOnSignal := runCmd.String("trigger-enforce-on-signal", "", "a string")
	verbose := runCmd.Bool("verbose", false, "a bool")

	// permissions
	allowFileSystemReadAccess := runCmd.Bool("allow-file-system-read", false, "a bool")
	allowFileSystemWriteAccess := runCmd.Bool("allow-file-system-write", false, "a bool")
	allowNetworkClient := runCmd.Bool("allow-network-client", false, "a bool")
	allowNetworkServer := runCmd.Bool("allow-network-server", false, "a bool")
	allowFileSystemAccess := runCmd.Bool("allow-file-system", false, "a bool")
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

	// parse flags now
	err := runCmd.Parse(os.Args[2:])
	if err != nil {
		fmt.Println(err.Error())
		runCmd.Usage()
	}

	allowList := runtime.NewSyscallAllowList()

	if *allowFileSystemWriteAccess {
		allowList.AllowAllFileSystemWriteAccess()
		allowList.AllowAllFileSystemReadAccess()
		allowList.AllowAllFileDescriptors()
		runtime.Get().FsConfig.FileSystemAllowRead = true
		runtime.Get().FsConfig.FileSystemAllowWrite = true
	} else if *allowFileSystemReadAccess {
		allowList.AllowAllFileSystemReadAccess()
		allowList.AllowAllFileDescriptors()
		runtime.Get().FsConfig.FileSystemAllowRead = true
		runtime.Get().FsConfig.FileSystemAllowWrite = false
	} else if *allowFileSystemAccess {
		allowList.AllowAllFileSystemAccess()
	}

	if *allowProcessManagement {
		allowList.AllowProcessManagement()
	}

	if *allowNetworkClient {
		allowList.AllowNetworkClient()
		conf.NetworkAllowClient = true
	}

	if *allowNetworkServer {
		allowList.AllowNetworkServer()
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

	if mode == "trace" {
		conf.ExecutionMode = runtime.EXECUTION_MODE_TRACE
	} else if mode == "run" {
		conf.ExecutionMode = runtime.EXECUTION_MODE_RUN
	} else {
		runCmd.Usage()
		os.Exit(1)
	}

	return runCmd.Args()
}

func waitForShutdown(cancel context.CancelFunc, tracee context.Context) {
	signal, stop := signal.NotifyContext(context.Background(), syscall.SIGKILL, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-signal.Done():
		break
	case <-tracee.Done():
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
	os.Exit(exitCode)
}
