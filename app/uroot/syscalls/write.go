//go:build (linux && arm64) || (linux && amd64) || (linux && riscv64)

package syscalls

import (
	"fmt"

	"github.com/discue/go-syscall-gatekeeper/app/runtime"
	"github.com/discue/go-syscall-gatekeeper/app/uroot/syscalls/args"
)

// IsWriteAllowed decides allowance for write-like syscalls using fd context.
// Unified signature: derive fd traits via args helpers using Syscall.TraceePID.
func IsWriteAllowed(s Syscall, isEnter bool) bool {
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
	if runtime.Get().FileSystemAllowWrite && isFile {
		return true
	}
	isPipe := args.IsPipe(s.TraceePID, fd)
	if isPipe {
		return true
	}

	isEventFd := args.FdType(s.TraceePID, fd) == args.FDAnonEvent
	if isEventFd {
		println(fmt.Sprintf("Trying to %s from fd %d which is a anon eventfd %t", "write", fd, isEventFd))
		return true
	}

	return false
}
