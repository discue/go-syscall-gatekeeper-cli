//go:build (linux && arm64) || (linux && amd64) || (linux && riscv64)

package syscalls

import (
	"github.com/cuandari/lib/app/runtime"
)

// IsUnlinkAtAllowed checks unlinkat(dirfd, pathname, flags).
// pathname is arg 1, dirfd is arg 0.
func IsUnlinkAtAllowed(s Syscall, isEnter bool) bool {
	writeAllowed := runtime.Get().FileSystemAllowWrite
	if !writeAllowed {
		return false
	}
	if len(runtime.Get().FileSystemAllowedPaths) > 0 {
		return PathIsAllowed(s, 1, 0)
	}
	return true
}
