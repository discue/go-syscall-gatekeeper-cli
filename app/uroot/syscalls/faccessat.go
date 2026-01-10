//go:build (linux && arm64) || (linux && amd64) || (linux && riscv64)

package syscalls

import "github.com/cuandari/lib/app/runtime"

// IsFaccessAtAllowed checks faccessat/faccessat2(dirfd, pathname, ...).
// It requires read permission and that the pathname is allowed via
// PathIsAllowed(pathArgIndex=1, dirfdArgIndex=0).
func IsFaccessAtAllowed(s Syscall, isEnter bool) bool {
	if !runtime.Get().FileSystemAllowRead {
		return false
	}
	return PathIsAllowed(s, 1, 0)
}
