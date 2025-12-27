//go:build (linux && arm64) || (linux && amd64) || (linux && riscv64)

package syscalls

import (
	"github.com/discue/go-syscall-gatekeeper/app/runtime"
	"golang.org/x/sys/unix"
)

// IsConnectAllowed determines allowance for connect using the sockaddr family
// decoded directly from the tracee via the Syscall.Reader.
func IsConnectAllowed(s Syscall) bool {
	allow := false
	addr := s.Args[1].Pointer()
	var family uint16
	if addr != 0 && s.Reader != nil {
		if _, err := s.Reader(addr, &family); err == nil {
			if runtime.Get().LocalSocketsAllow && (family == uint16(unix.AF_UNIX) || family == uint16(unix.AF_NETLINK)) {
				allow = true
			} else if family == uint16(unix.AF_INET) || family == uint16(unix.AF_INET6) || family == uint16(unix.AF_PACKET) {
				allow = runtime.Get().NetworkAllowClient || runtime.Get().NetworkAllowServer
			}
		}
	}
	return allow
}
