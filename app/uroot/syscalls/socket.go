//go:build (linux && arm64) || (linux && amd64) || (linux && riscv64)

package syscalls

import (
	"fmt"

	"github.com/cuandari/lib/app/runtime"
	"golang.org/x/sys/unix"
)

// IsSocketAllowed determines allowance based on socket domain and runtime flags.
// Signature unified to accept Syscall; extracts domain from args.
func IsSocketAllowed(s Syscall, isEnter bool) bool {
	domain := int(s.Args[0].Int())

	// Local-only families
	if domain == unix.AF_UNIX || domain == unix.AF_NETLINK {
		fmt.Println("socket domain:", domain, "allowed as local socket", runtime.Get().LocalSocketsAllow)
		return runtime.Get().LocalSocketsAllow
	}

	// Network families
	if domain == unix.AF_INET || domain == unix.AF_INET6 || domain == unix.AF_PACKET {
		fmt.Println("socket domain:", domain, "allowed as network socket", runtime.Get().NetworkAllowClient || runtime.Get().NetworkAllowServer)
		return runtime.Get().NetworkAllowClient || runtime.Get().NetworkAllowServer
	}

	fmt.Println("socket domain:", domain, "not explicitly allowed")
	// print socket constants for debugging uncommon domains
	fmt.Println("AF_UNIX:", unix.AF_UNIX, "AF_NETLINK:", unix.AF_NETLINK, "AF_INET:", unix.AF_INET, "AF_INET6:", unix.AF_INET6, "AF_PACKET:", unix.AF_PACKET)

	return false
}
