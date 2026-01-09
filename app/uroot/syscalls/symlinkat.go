//go:build (linux && arm64) || (linux && amd64) || (linux && riscv64)

package syscalls

import (
	"github.com/cuandari/lib/app/runtime"
)

// IsSymlinkAtAllowed checks symlinkat(target, newdirfd, linkpath).
// linkpath is arg 2, newdirfd is arg 1.
func IsSymlinkAtAllowed(s Syscall, isEnter bool) bool {
	writeAllowed := runtime.Get().FileSystemAllowWrite
	if !writeAllowed {
		return false
	}
	if len(runtime.Get().FileSystemAllowedPaths) > 0 {
		return PathIsAllowed(s, 2, 1)
	}
	return true
}
