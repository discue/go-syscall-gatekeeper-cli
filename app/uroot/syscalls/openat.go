package syscalls

import (
	"golang.org/x/sys/unix"
)

// IsOpenAtReadOnly checks read-only intent for openat.
// Unified signature: extract flags from Syscall.Args.
func IsOpenAtReadOnly(s Syscall) bool {
	flags := int(s.Args[3].Uint())
	writeAccMask := unix.O_WRONLY | unix.O_RDWR
	if flags&writeAccMask == 0 {
		return (flags&(unix.O_CREAT|unix.O_TRUNC|unix.O_APPEND) == 0)
	}
	return false
}
