//go:build (linux && arm64) || (linux && amd64) || (linux && riscv64)

package syscalls

import (
	"github.com/discue/go-syscall-gatekeeper/app/runtime"
	"golang.org/x/sys/unix"
)

// IsSocketAllowed determines allowance based on socket domain and runtime flags.
// Signature unified to accept Syscall; extracts domain from args.
func IsSocketAllowed(s Syscall) bool {
	domain := int(s.Args[0].Int())

	// Local-only families
	if runtime.Get().LocalSocketsAllow && (domain == unix.AF_UNIX || domain == unix.AF_NETLINK) {
		return true
	}

	// Network families
	if domain == unix.AF_INET || domain == unix.AF_INET6 || domain == unix.AF_PACKET {
		return runtime.Get().NetworkAllowClient || runtime.Get().NetworkAllowServer
	}

	// Default: use configured allow-map fallback
	return runtime.Get().SyscallsAllowMap["socket"]
}
