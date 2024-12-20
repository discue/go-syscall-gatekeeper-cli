package uroot

import (
	"github.com/discue/go-syscall-gatekeeper/app/runtime"
)

func allowSyscall(name string) bool {
	if runtime.Get().SyscallsAllowMap[name] == false {
		return false
	}

	return true
}
