package uroot

import (
	"github.com/cuandari/lib/app/runtime"
)

func allowSyscall(name string) bool {
	return runtime.Get().SyscallsAllowMap[name]
}
