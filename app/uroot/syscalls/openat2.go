//go:build (linux && arm64) || (linux && amd64) || (linux && riscv64)

package syscalls

import (
	"github.com/cuandari/lib/app/runtime"
	"golang.org/x/sys/unix"
)

// IsOpenAt2ReadOnly checks read-only intent for openat2 by decoding open_how
// from tracee memory via Syscall.Reader.
func IsOpenAt2ReadOnly(s Syscall, isEnter bool) bool {
	type openHow struct {
		Flags   uint64
		Mode    uint64
		Resolve uint64
	}
	addr := s.Args[2].Pointer()
	var how openHow
	if s.Reader == nil {
		return false
	}
	if _, err := s.Reader(addr, &how); err != nil {
		return false
	}
	flags := int(how.Flags)
	writeAccMask := unix.O_WRONLY | unix.O_RDWR
	if flags&writeAccMask == 0 {
		return (flags&(unix.O_CREAT|unix.O_TRUNC|unix.O_APPEND) == 0)
	}
	return false
}

func IsOpenAt2Allowed(s Syscall, isEnter bool) bool {
	readAllowed := runtime.Get().FileSystemAllowRead
	writeAllowed := runtime.Get().FileSystemAllowWrite

	isReadOnlySyscall := IsOpenAt2ReadOnly(s, isEnter)
	if isReadOnlySyscall && readAllowed {
		return PathIsAllowed(s, 1, 0)
	}

	if !isReadOnlySyscall && writeAllowed {
		return PathIsAllowed(s, 1, 0)
	}

	return false
}
