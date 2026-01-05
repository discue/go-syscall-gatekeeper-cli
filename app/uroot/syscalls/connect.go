//go:build (linux && arm64) || (linux && amd64) || (linux && riscv64)

package syscalls

import (
	"fmt"

	"github.com/cuandari/lib/app/runtime"
	"golang.org/x/sys/unix"
)

// IsConnectAllowed determines allowance for connect using the sockaddr family
// decoded directly from the tracee via the Syscall.Reader.
func IsConnectAllowed(s Syscall, isEnter bool) bool {
	addr := s.Args[1].Pointer()
	var family uint16
	if addr != 0 && s.Reader != nil {
		if _, err := s.Reader(addr, &family); err == nil {
			if family == uint16(unix.AF_UNIX) || family == uint16(unix.AF_NETLINK) {
				fmt.Println("connect family:", family, "connect to local socket", runtime.Get().LocalSocketsAllow)
				return runtime.Get().LocalSocketsAllow
			}

			if family == uint16(unix.AF_INET) || family == uint16(unix.AF_INET6) || family == uint16(unix.AF_PACKET) {
				// connect() is a client operation; require client permission
				fmt.Println("connect family", family, "connect to remote socket", runtime.Get().NetworkAllowClient)
				return runtime.Get().NetworkAllowClient
			}

			if family == uint16(unix.AF_UNSPEC) {
				// AF_UNSPEC connect on datagram sockets can “disconnect”; allow only if at least
				// local sockets or network client capability is enabled.
				fmt.Println("connect family", family, "connect to unspeced socket", runtime.Get().LocalSocketsAllow || runtime.Get().NetworkAllowClient)
				return runtime.Get().LocalSocketsAllow || runtime.Get().NetworkAllowClient
			}
		}
	}

	return false
}
