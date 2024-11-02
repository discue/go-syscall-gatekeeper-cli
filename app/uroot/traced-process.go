package uroot

import (
	"encoding/binary"
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

// process is a Linux thread.
type process struct {
	pid int

	// ptrace does not tell you whether a syscall-stop is a
	// syscall-enter-stop or syscall-exit-stop. You gotta keep track of
	// that shit your own self.
	lastSyscallStop *TraceRecord
}

// Name implements Task.Name.
func (p *process) Name() string {
	return fmt.Sprintf("[pid %d]", p.pid)
}

// Read reads from the process at Addr to the interface{}
// and returns a byte count and error.
func (p *process) Read(addr Addr, v interface{}) (int, error) {
	r := newProcReader(p.pid, uintptr(addr))
	err := binary.Read(r, binary.NativeEndian, v)
	return r.bytes, err
}

func (p *process) cont(signal unix.Signal) error {
	// Event has been processed. Restart 'em.
	if err := unix.PtraceSyscall(p.pid, int(signal)); err != nil {
		return os.NewSyscallError("ptrace(PTRACE_SYSCALL)", fmt.Errorf("on pid %d: %w", p.pid, err))
	}
	return nil
}
