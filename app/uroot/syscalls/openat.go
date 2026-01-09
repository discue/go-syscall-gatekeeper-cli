package syscalls

import (
	"github.com/cuandari/lib/app/runtime"
	"golang.org/x/sys/unix"
)

// IsOpenAtReadOnly checks read-only intent for openat.
// Unified signature: extract flags from Syscall.Args.
func IsOpenAtReadOnly(s Syscall, isEnter bool) bool {
	flags := int(s.Args[3].Uint())
	writeAccMask := unix.O_WRONLY | unix.O_RDWR
	if flags&writeAccMask == 0 {
		return (flags&(unix.O_CREAT|unix.O_TRUNC|unix.O_APPEND) == 0)
	}
	return false
}

func IsOpenAtAllowed(s Syscall, isEnter bool) bool {
	readAllowed := runtime.Get().FileSystemAllowRead
	writeAllowed := runtime.Get().FileSystemAllowWrite
	isReadOnlySyscall := IsOpenAtReadOnly(s, isEnter)

	if isReadOnlySyscall && readAllowed {
		return PathIsAllowed(s, 1, 0)
	}

	if !isReadOnlySyscall && writeAllowed {
		return PathIsAllowed(s, 1, 0)
	}

	return false
}
