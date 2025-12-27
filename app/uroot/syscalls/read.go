//go:build (linux && arm64) || (linux && amd64) || (linux && riscv64)

package syscalls

import (
	"github.com/discue/go-syscall-gatekeeper/app/runtime"
	"github.com/discue/go-syscall-gatekeeper/app/uroot/syscalls/args"
)

// IsReadAllowed decides allowance for read-like syscalls using fd context.
// Unified signature: derive fd traits via args helpers using Syscall.TraceePID.
func IsReadAllowed(s Syscall) bool {
	fd := s.Args[0].Int()
	isStdStream := args.IsStandardStream(fd)
	if isStdStream {
		return true
	}
	isSocket := args.IsSocket(s.TraceePID, fd)
	if (runtime.Get().NetworkAllowServer || runtime.Get().NetworkAllowClient || runtime.Get().LocalSocketsAllow) && isSocket {
		return true
	}
	isFile := args.IsFile(s.TraceePID, fd)
	if runtime.Get().FileSystemAllowRead && isFile {
		return true
	}
	isPipe := args.IsPipe(s.TraceePID, fd)
	if isPipe {
		return true
	}
	return runtime.Get().SyscallsAllowMap["read"]
}
