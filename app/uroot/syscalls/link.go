//go:build (linux && arm64) || (linux && amd64) || (linux && riscv64)

package syscalls

import (
	"github.com/cuandari/lib/app/runtime"
)

// IsLinkAllowed checks link(oldpath, newpath).
func IsLinkAllowed(s Syscall, isEnter bool) bool {
	writeAllowed := runtime.Get().FileSystemAllowWrite
	if !writeAllowed {
		return false
	}
	if !PathIsAllowed(s, 0, -1) {
		return false
	}
	if !PathIsAllowed(s, 1, -1) {
		return false
	}
	return true
}
