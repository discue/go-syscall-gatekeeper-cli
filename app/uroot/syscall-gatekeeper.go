package uroot

import (
	"github.com/discue/go-syscall-gatekeeper/app/runtime"
)

func allowSyscall(name string) bool {
	return runtime.Get().SyscallsAllowMap[name]
}
