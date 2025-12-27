//go:build (linux && arm64) || (linux && amd64) || (linux && riscv64)

package syscalls

import "golang.org/x/sys/unix"

// IsOpenReadOnly checks open(pathname, flags, mode) for read-only intent.
// Unified signature: extract flags from Syscall.Args.
func IsOpenReadOnly(s Syscall) bool {
	flags := int(s.Args[1].Uint())
	writeAccMask := unix.O_WRONLY | unix.O_RDWR
	if flags&writeAccMask == 0 {
		// Even with O_RDONLY, certain flags imply write intent
		return (flags&(unix.O_CREAT|unix.O_TRUNC|unix.O_APPEND) == 0)
	}
	return false
}
