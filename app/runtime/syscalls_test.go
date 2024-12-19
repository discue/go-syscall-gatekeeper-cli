package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAllowAllFileSystemAccess(t *testing.T) {
	a := assert.New(t)
	sal := NewSyscallAllowList()
	sal.AllowAllFileSystemAccess()

	a.NotEmpty(sal.Syscalls)
	a.Contains(sal.Syscalls, "access")

}

func TestAllowProcessManagement(t *testing.T) {
	a := assert.New(t)
	sal := NewSyscallAllowList()
	sal.AllowProcessManagement()

	a.NotEmpty(sal.Syscalls)
	a.Contains(sal.Syscalls, "fork")
}

func TestAllowNetworking(t *testing.T) {
	a := assert.New(t)
	sal := NewSyscallAllowList()
	sal.AllowNetworking()

	a.NotEmpty(sal.Syscalls)
	a.Contains(sal.Syscalls, "bind")

}

func TestAllowMemoryManagement(t *testing.T) {
	a := assert.New(t)
	sal := NewSyscallAllowList()
	sal.AllowMemoryManagement()

	a.NotEmpty(sal.Syscalls)
	a.Contains(sal.Syscalls, "mmap")
}

func TestAllowSignals(t *testing.T) {
	a := assert.New(t)
	sal := NewSyscallAllowList()
	sal.AllowSignals()

	a.NotEmpty(sal.Syscalls)
	a.Contains(sal.Syscalls, "rt_sigaction")
}

func TestAllowTimersAndClocksManagement(t *testing.T) {
	a := assert.New(t)
	sal := NewSyscallAllowList()
	sal.AllowTimersAndClocksManagement()

	a.NotEmpty(sal.Syscalls)
	a.Contains(sal.Syscalls, "timer_create")

}

func TestAllowSecurityAndPermissions(t *testing.T) {
	a := assert.New(t)
	sal := NewSyscallAllowList()
	sal.AllowSecurityAndPermissions()

	a.NotEmpty(sal.Syscalls)
	a.Contains(sal.Syscalls, "setresuid")
}

func TestAllowSystemInformation(t *testing.T) {
	a := assert.New(t)
	sal := NewSyscallAllowList()
	sal.AllowSystemInformation()

	a.NotEmpty(sal.Syscalls)
	a.Contains(sal.Syscalls, "uname")
}

func TestAllowProcessCommunication(t *testing.T) {
	a := assert.New(t)
	sal := NewSyscallAllowList()
	sal.AllowProcessCommunication()

	a.NotEmpty(sal.Syscalls)
	a.Contains(sal.Syscalls, "mq_open")
}

func TestAllowProcessSynchronization(t *testing.T) {
	a := assert.New(t)
	sal := NewSyscallAllowList()
	sal.AllowProcessSynchronization()

	a.NotEmpty(sal.Syscalls)
	a.Contains(sal.Syscalls, "futex")
}

func TestAllowMisc(t *testing.T) {
	a := assert.New(t)
	sal := NewSyscallAllowList()
	sal.AllowMisc()

	a.NotEmpty(sal.Syscalls)
	a.Contains(sal.Syscalls, "sync")
}
