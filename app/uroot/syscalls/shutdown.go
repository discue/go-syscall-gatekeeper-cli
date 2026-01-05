//go:build (linux && arm64) || (linux && amd64) || (linux && riscv64)

package syscalls

import (
	"github.com/cuandari/lib/app/runtime"
	"github.com/cuandari/lib/app/uroot/syscalls/args"
)

// IsShutdownAllowed returns whether shutdown should be permitted.
// Unified signature: derive fd traits via args helpers using Syscall.TraceePID.
func IsShutdownAllowed(s Syscall, isEnter bool) bool {
	fd := s.Args[0].Int()
	isSocket := args.IsSocket(s.TraceePID, fd)
	if (runtime.Get().NetworkAllowServer || runtime.Get().NetworkAllowClient || runtime.Get().LocalSocketsAllow) && isSocket {
		return true
	}
	return args.IsStandardStream(fd)
}
