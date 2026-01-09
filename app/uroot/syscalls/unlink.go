//go:build (linux && arm64) || (linux && amd64) || (linux && riscv64)

package syscalls

import (
	"github.com/cuandari/lib/app/runtime"
)

// IsUnlinkAllowed checks unlink(pathname).
// pathname is arg 0.
func IsUnlinkAllowed(s Syscall, isEnter bool) bool {
	writeAllowed := runtime.Get().FileSystemAllowWrite
	if !writeAllowed {
		return false
	}
	if len(runtime.Get().FileSystemAllowedPaths) > 0 {
		return PathIsAllowed(s, 0, -1)
	}
	return true
}
