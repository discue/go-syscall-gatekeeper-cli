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
	"github.com/discue/go-syscall-gatekeeper/app/utils"
)

var logger = utils.NewLogger("main")

func main() {
	mainCtx, cancel := context.WithCancel(context.Background())

	tracee := startTracee(mainCtx)
	waitForShutdown(cancel, tracee)
}

func startTracee(c context.Context) context.Context {
	traceeCtx, _ := context.WithCancel(c)
	args := configureAndParseArgs()

	_, exitContext, err := uroot.Exec(traceeCtx, args[0], args[1:])
	if err != nil {
		fmt.Println(err.Error())
	}

	return exitContext
}

func configureAndParseArgs() []string {
	conf := runtime.Get()
	mode := os.Args[1]

	runCmd := flag.NewFlagSet("mode", flag.ExitOnError)
	runLogSearchString := runCmd.String("log-search-string", "", "a string")
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
	noImplicitAllow := runCmd.Bool("no-implicit-allow", false, "a bool")

	// parse flags now
	runCmd.Parse(os.Args[2:])

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

	if *runLogSearchString != "" {
		conf.LogSearchString = *runLogSearchString
		conf.EnforceOnStartup = false
	}

	if len(allowList.Syscalls) > 0 {
		conf.SyscallsAllowList = allowList.Syscalls
		conf.SyscallsAllowMap = runtime.CreateSyscallAllowMap(conf.SyscallsAllowList)
		conf.SyscallsKillTargetIfNotAllowed = true
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
	if e.ExitEvent != nil {
		if e.ExitEvent.Signal != "" {
			exitCode = 111
		} else {
			exitCode = e.ExitEvent.WaitStatus.ExitStatus()
		}
	}

	println(fmt.Sprintf("Exiting with code %d", exitCode))
	// exit
	os.Exit(exitCode)
}
