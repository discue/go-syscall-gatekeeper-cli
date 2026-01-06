// Package main is the entry point for the k6 CLI application. It assembles all the crucial components for the running.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cuandari/lib/app/runtime"
	"github.com/cuandari/lib/app/uroot"
	"github.com/cuandari/lib/cli"
)

var osExit = os.Exit // Assign exit to a variable to allow mocking in unit tests
func exit(code int) {
	osExit(code)
}

// CLI flag type and constants moved to cli package.

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

	c := cli.NewCommand()

	if len(os.Args) < 3 {
		fmt.Println("Error: You did not provide enough parameters and flags.")
		c.Usage()
		exit(100)
	}

	mode := os.Args[1]

	// Pre-scan for dynamic syscall allow flags and filter them out before parsing
	// Supported forms:
	//  - --allow-syscall-<name>
	//  - --allow-syscall=<name>
	rawArgs := os.Args[2:]
	filteredArgs, dynamicSyscalls := c.PreScanDynamicSyscalls(rawArgs)

	// parse known flags now
	err := c.Parse(filteredArgs)
	if err != nil {
		fmt.Println(err.Error())
		c.Usage()
		exit(100)
	}

	allowList := runtime.NewSyscallAllowList()

	if *c.AllowFileSystemWriteAccess || *c.AllowFileSystemAccess {
		allowList.AllowAllFileSystemWriteAccess()
		allowList.AllowAllFileSystemReadAccess()
		allowList.AllowAllFileDescriptors()
		runtime.Get().FileSystemAllowRead = true
		runtime.Get().FileSystemAllowWrite = true
	} else if *c.AllowFileSystemReadAccess {
		allowList.AllowAllFileSystemReadAccess()
		allowList.AllowAllFileDescriptors()
		runtime.Get().FileSystemAllowRead = true
		runtime.Get().FileSystemAllowWrite = false
	}

	if *c.AllowFileSystemPermissionsAccess {
		allowList.AllowAllFilePermissions()
	}

	if *c.AllowProcessManagement {
		allowList.AllowProcessManagement()
	}

	if *c.AllowNetworkClient {
		allowList.AllowNetworkClient()
		allowList.AllowAllFileDescriptors()
		conf.NetworkAllowClient = true
	}

	if *c.AllowNetworkServer {
		allowList.AllowNetworkServer()
		allowList.AllowAllFileDescriptors()
		conf.NetworkAllowServer = true
	}

	if *c.AllowNetworkLocalSockets {
		allowList.AllowLocalSockets()
		// allowList.AllowAllFileDescriptors()
		conf.LocalSocketsAllow = true
	}

	if *c.AllowNetworking {
		allowList.AllowNetworking()
	}

	if *c.AllowMemoryManagement {
		allowList.AllowMemoryManagement()
	}

	if *c.AllowSignals {
		allowList.AllowSignals()
	}

	if *c.AllowTimersAndClocksManagement {
		allowList.AllowTimersAndClocksManagement()
	}

	if *c.AllowSecurityAndPermissions {
		allowList.AllowSecurityAndPermissions()
	}

	if *c.AllowSystemInformation {
		allowList.AllowSystemInformation()
	}

	if *c.AllowProcessCommunication {
		allowList.AllowProcessCommunication()
	}

	if *c.AllowProcessSynchronization {
		allowList.AllowProcessSynchronization()
	}

	if *c.AllowMisc {
		allowList.AllowMisc()
	}

	// Append dynamically allowed syscalls collected from CLI
	if len(dynamicSyscalls) > 0 {
		allowList.Syscalls = append(allowList.Syscalls, dynamicSyscalls...)
	}

	if *c.AllowImplicitCommands {
		allowList.AllowProcessManagement()
		allowList.AllowMemoryManagement()
		allowList.AllowProcessSynchronization()
		allowList.AllowSignals()
		// Basic time queries and sleep are broadly required and safe
		allowList.AllowBasicTime()
		allowList.AllowMisc()
		allowList.AllowSecurityAndPermissions()
		allowList.AllowSystemInformation()
	}

	conf.VerboseLog = *c.Verbose

	if !*c.EnforceOnStartup {
		if *c.TriggerEnforceOnLogMatch == "" && *c.TriggerEnforceOnSignal == "" {
			fmt.Println("Error: To delay the enforcement of seccomp policies, please also specify either --trigger-enforce-on-log-match or --trigger-enforce-on-signal.")
			c.Usage()
			exit(100)
		} else {
			conf.EnforceOnStartup = false
		}
	} else {
		conf.EnforceOnStartup = true
	}

	if *c.TriggerEnforceOnLogMatch != "" {
		conf.TriggerEnforceLogMatch = *c.TriggerEnforceOnLogMatch
	} else if *c.TriggerEnforceOnSignal != "" {
		conf.TriggerEnforceSignal = *c.TriggerEnforceOnSignal
	}

	if len(allowList.Syscalls) > 0 {
		conf.SyscallsAllowList = allowList.Syscalls
		conf.SyscallsAllowMap = runtime.CreateSyscallAllowMap(conf.SyscallsAllowList)
	}

	if c.Action == cli.ErrorAction {
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
		c.Usage()
		exit(100)
	}

	return c.Args()
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
