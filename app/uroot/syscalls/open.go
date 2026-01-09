//go:build (linux && arm64) || (linux && amd64) || (linux && riscv64)

package syscalls

import (
	"github.com/cuandari/lib/app/runtime"
	"golang.org/x/sys/unix"
)

// IsOpenReadOnly checks open(pathname, flags, mode) for read-only intent.
// Unified signature: extract flags from Syscall.Args.
func IsOpenReadOnly(s Syscall, isEnter bool) bool {
	flags := int(s.Args[1].Uint())
	writeAccMask := unix.O_WRONLY | unix.O_RDWR
	if flags&writeAccMask == 0 {
		// Even with O_RDONLY, certain flags imply write intent
		return (flags&(unix.O_CREAT|unix.O_TRUNC|unix.O_APPEND) == 0)
	}
	return false
}

func IsOpenAllowed(s Syscall, isEnter bool) bool {
	readAllowed := runtime.Get().FileSystemAllowRead
	writeAllowed := runtime.Get().FileSystemAllowWrite

	isReadOnlySyscall := IsOpenReadOnly(s, isEnter)

	if isReadOnlySyscall && readAllowed {
		return PathIsAllowed(s, 0, -1)
	}

	if !isReadOnlySyscall && writeAllowed {
		return PathIsAllowed(s, 0, -1)
	}

	return false
}
