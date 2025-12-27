package syscalls

import (
	"golang.org/x/sys/unix"
)

// A processIo is used to implement io.Reader and io.Writer.
// it contains a pid, which is unchanging; and an
// addr and byte count which change as IO proceeds.
type processIo struct {
	pid   int
	addr  uintptr
	bytes int
}

// Read implements io.Read for a processIo.
func (p *processIo) Read(b []byte) (int, error) {
	n, err := unix.PtracePeekData(p.pid, p.addr, b)
	if err != nil {
		return n, err
	}
	p.addr += uintptr(n)
	p.bytes += n
	return n, nil
}

// newProcReader returns an io.Reader for a processIo.
func newProcReader(pid int, addr uintptr) *processIo {
	return &processIo{pid: pid, addr: addr}
}
