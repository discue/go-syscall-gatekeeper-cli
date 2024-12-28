package main

import (
	"os"
	"testing"

	"github.com/discue/go-syscall-gatekeeper/app/runtime"
	"github.com/stretchr/testify/assert"
)

func TestNoImplicitAllow(t *testing.T) {
	a := assert.New(t)
	os.Args = []string{"", "run", "--no-implicit-allow", "true"}

	configureAndParseArgs()
	a.Empty(runtime.Get().SyscallsAllowList)
	a.True(runtime.Get().SyscallsKillTargetIfNotAllowed)
}

func TestAllowFileSystemReadAccess(t *testing.T) {
	a := assert.New(t)
	os.Args = []string{"", "run", "--allow-file-system-read", "ls", "-l"}

	configureAndParseArgs()
	a.Contains(runtime.Get().SyscallsAllowList, "read")
	a.True(runtime.Get().SyscallsKillTargetIfNotAllowed)
}

func TestAllowFileSystemWriteAccess(t *testing.T) {
	a := assert.New(t)
	os.Args = []string{"", "run", "--allow-file-system-write", "ls", "-l"}

	configureAndParseArgs()
	a.Contains(runtime.Get().SyscallsAllowList, "write")
	a.True(runtime.Get().SyscallsKillTargetIfNotAllowed)
}

func TestAllowFileSystemAccess(t *testing.T) {
	a := assert.New(t)
	os.Args = []string{"", "run", "--allow-file-system", "ls", "-l"}

	configureAndParseArgs()
	a.Contains(runtime.Get().SyscallsAllowList, "openat2")
	a.True(runtime.Get().SyscallsKillTargetIfNotAllowed)
}

func TestAllowProcessManagement(t *testing.T) {
	a := assert.New(t)
	os.Args = []string{"", "run", "--allow-process-management", "ps", "-ef"}

	configureAndParseArgs()
	a.Contains(runtime.Get().SyscallsAllowList, "fork")
	a.True(runtime.Get().SyscallsKillTargetIfNotAllowed)
}

func TestAllowNetworkClient(t *testing.T) {
	a := assert.New(t)
	os.Args = []string{"", "run", "--allow-network-client", "curl", "https://google.com"}

	configureAndParseArgs()
	a.Contains(runtime.Get().SyscallsAllowList, "connect")
	a.True(runtime.Get().SyscallsKillTargetIfNotAllowed)
}

func TestAllowNetworkServer(t *testing.T) {
	a := assert.New(t)
	os.Args = []string{"", "run", "--allow-network-server", "true"}

	configureAndParseArgs()
	a.Contains(runtime.Get().SyscallsAllowList, "accept")
	a.True(runtime.Get().SyscallsKillTargetIfNotAllowed)
}

func TestAllowNetworking(t *testing.T) {
	a := assert.New(t)
	os.Args = []string{"", "run", "--allow-networking", "curl", "https://google.com"}

	configureAndParseArgs()
	a.Contains(runtime.Get().SyscallsAllowList, "accept")
	a.True(runtime.Get().SyscallsKillTargetIfNotAllowed)
}

func TestAllowMemoryManagement(t *testing.T) {
	a := assert.New(t)
	os.Args = []string{"", "run", "--allow-memory-management", "true"}

	configureAndParseArgs()
	a.Contains(runtime.Get().SyscallsAllowList, "mmap")
	a.True(runtime.Get().SyscallsKillTargetIfNotAllowed)
}

func TestAllowSignals(t *testing.T) {
	a := assert.New(t)
	os.Args = []string{"", "run", "--allow-signals", "true"}

	configureAndParseArgs()
	a.Contains(runtime.Get().SyscallsAllowList, "rt_sigaction")
	a.True(runtime.Get().SyscallsKillTargetIfNotAllowed)
}

func TestAllowTimersAndClocksManagement(t *testing.T) {
	a := assert.New(t)
	os.Args = []string{"", "run", "--allow-timers-and-clocks-management", "true"}

	configureAndParseArgs()
	a.Contains(runtime.Get().SyscallsAllowList, "timer_create")
	a.True(runtime.Get().SyscallsKillTargetIfNotAllowed)
}

func TestAllowSecurityAndPermissions(t *testing.T) {
	a := assert.New(t)
	os.Args = []string{"", "run", "--allow-security-and-permissions", "true"}

	configureAndParseArgs()
	a.Contains(runtime.Get().SyscallsAllowList, "setresuid")
	a.True(runtime.Get().SyscallsKillTargetIfNotAllowed)
}

func TestAllowSystemInformation(t *testing.T) {
	a := assert.New(t)
	os.Args = []string{"", "run", "--allow-system-information", "true"}
	configureAndParseArgs()
	a.Contains(runtime.Get().SyscallsAllowList, "uname")
	a.True(runtime.Get().SyscallsKillTargetIfNotAllowed)
}

func TestAllowProcessCommunication(t *testing.T) {
	a := assert.New(t)
	os.Args = []string{"", "run", "--allow-process-communication", "true"}

	configureAndParseArgs()
	a.Contains(runtime.Get().SyscallsAllowList, "mq_open")
	a.True(runtime.Get().SyscallsKillTargetIfNotAllowed)
}

func TestAllowProcessSynchronization(t *testing.T) {
	a := assert.New(t)
	os.Args = []string{"", "run", "--allow-process-synchronization", "true"}

	configureAndParseArgs()
	a.Contains(runtime.Get().SyscallsAllowList, "futex")
	a.True(runtime.Get().SyscallsKillTargetIfNotAllowed)
}

func TestAllowMisc(t *testing.T) {
	a := assert.New(t)
	os.Args = []string{"", "run", "--allow-misc", "true"}

	configureAndParseArgs()
	a.Contains(runtime.Get().SyscallsAllowList, "sync")
	a.True(runtime.Get().SyscallsKillTargetIfNotAllowed)
}

func TestLogSearchString(t *testing.T) {
	a := assert.New(t)
	os.Args = []string{"", "run", "--log-search-string", "test", "true"}

	configureAndParseArgs()
	a.Equal("test", runtime.Get().LogSearchString)
	a.False(runtime.Get().EnforceOnStartup)
}

func TestLogSearchStringKillTarget(t *testing.T) {
	a := assert.New(t)
	os.Args = []string{"", "run", "--log-search-string", "test", "true"}

	configureAndParseArgs()
	a.Equal("test", runtime.Get().LogSearchString)
	a.True(runtime.Get().SyscallsKillTargetIfNotAllowed)
}

func TestDenyTarget(t *testing.T) {
	a := assert.New(t)
	os.Args = []string{"", "run", "--on-syscall-denied", "error", "true"}

	configureAndParseArgs()
	a.False(runtime.Get().SyscallsKillTargetIfNotAllowed)
	a.True(runtime.Get().SyscallsDenyTargetIfNotAllowed)
}

func TestKillTarget(t *testing.T) {
	a := assert.New(t)
	os.Args = []string{"", "run", "--on-syscall-denied", "kill", "true"}

	configureAndParseArgs()
	a.True(runtime.Get().SyscallsKillTargetIfNotAllowed)
	a.False(runtime.Get().SyscallsDenyTargetIfNotAllowed)
}

func TestVerboseLog(t *testing.T) {
	a := assert.New(t)
	os.Args = []string{"", "run", "--verbose", "true"}

	configureAndParseArgs()
	a.True(runtime.Get().VerboseLog)
}

func TestRunMode(t *testing.T) {
	a := assert.New(t)
	os.Args = []string{"", "run", "true"}

	configureAndParseArgs()
	a.Equal(runtime.EXECUTION_MODE_RUN, runtime.Get().ExecutionMode)
}

func TestTraceMode(t *testing.T) {
	a := assert.New(t)
	os.Args = []string{"", "trace", "true"}

	configureAndParseArgs()
	a.Equal(runtime.EXECUTION_MODE_TRACE, runtime.Get().ExecutionMode)
}
