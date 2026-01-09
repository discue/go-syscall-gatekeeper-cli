//go:build (linux && arm64) || (linux && amd64) || (linux && riscv64)

package syscalls

import "github.com/cuandari/lib/app/runtime"

// IsMkdirAllowed checks mkdir(pathname) semantics against runtime config.
// pathname is arg 0.
func IsMkdirAllowed(s Syscall, isEnter bool) bool {
	writeAllowed := runtime.Get().FileSystemAllowWrite
	if !writeAllowed {
		return false
	}
	// If a path whitelist is configured, ensure the pathname is allowed.
	return PathIsAllowed(s, 0, -1)
}

// IsMkdirAtAllowed checks mkdirat(dirfd, pathname, mode) semantics against runtime config.
// pathname is arg 1, dirfd is arg 0.
func IsMkdirAtAllowed(s Syscall, isEnter bool) bool {
	writeAllowed := runtime.Get().FileSystemAllowWrite
	if !writeAllowed {
		return false
	}
	return PathIsAllowed(s, 1, 0)
}
