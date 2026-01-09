//go:build (linux && arm64) || (linux && amd64) || (linux && riscv64)

package syscalls

import (
	"github.com/cuandari/lib/app/runtime"
)

// IsRenameAtAllowed checks renameat(olddirfd, oldpath, newdirfd, newpath).
func IsRenameAtAllowed(s Syscall, isEnter bool) bool {
	writeAllowed := runtime.Get().FileSystemAllowWrite
	if !writeAllowed {
		return false
	}
	if !PathIsAllowed(s, 1, 0) {
		return false
	}
	if !PathIsAllowed(s, 3, 2) {
		return false
	}
	return true
}
