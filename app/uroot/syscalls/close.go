//go:build (linux && arm64) || (linux && amd64) || (linux && riscv64)

package syscalls

func IsCloseAllowed(s Syscall, isEnter bool) bool {
	// Always allow close syscall
	return true
}
