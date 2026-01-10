//go:build (linux && arm64) || (linux && amd64) || (linux && riscv64)

package syscalls

import "github.com/cuandari/lib/app/runtime"

// IsAccessAllowed checks access(pathname, mode). It requires read permission and
// that the pathname is allowed via PathIsAllowed(pathArgIndex=0, dirfd=-1).
func IsAccessAllowed(s Syscall, isEnter bool) bool {
	if !runtime.Get().FileSystemAllowRead {
		return false
	}
	return PathIsAllowed(s, 0, -1)
}
