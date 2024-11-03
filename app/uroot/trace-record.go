package uroot

import (
	"os"
	"time"

	"golang.org/x/sys/unix"
)

// TraceRecord has information about a process event.
type TraceRecord struct {
	PID   int
	Time  time.Time
	Event EventType

	// Poor man's union. One of the following five will be populated
	// depending on the Event.

	Syscall    *SyscallEvent
	SignalExit *SignalEvent
	SignalStop *SignalEvent
	Exit       *ExitEvent
	NewChild   *NewChildEvent
}

func (t *TraceRecord) syscallStop(p *process) error {
	t.Syscall = &SyscallEvent{}

	if err := unix.PtraceGetRegs(p.pid, &t.Syscall.Regs); err != nil {
		return &TraceError{
			PID: p.pid,
			Err: os.NewSyscallError("ptrace(PTRACE_GETREGS)", err),
		}
	}

	// name, _ := sec.ScmpSyscall(t.Syscall.Regs.Orig_rax).GetName()
	t.Syscall.FillArgs()

	// if name == "openat" {
	// 	var s string
	// 	var b [1]byte
	// 	addr := t.Syscall.Args[1].Pointer()
	// 	for len(s) < 1024 {
	// 		if _, err := p.Read(addr, b[:]); err != nil {
	// 			break
	// 		}
	// 		if b[0] == 0 {
	// 			break
	// 		}
	// 		s = s + string(b[:])
	// 		addr++
	// 	}

	// 	fmt.Printf("name %s arg %s", name, s)
	// } else {
	// 	fmt.Printf("name %s args %#v", name, t.Syscall.Args)
	// }

	// TODO: the ptrace man page mentions that seccomp can inject a
	// syscall-exit-stop without a preceding syscall-enter-stop. Detect
	// that here, however you'd detect it...
	if p.lastSyscallStop.Event == SyscallEnter {
		t.Event = SyscallExit
		t.Syscall.FillRet()
		t.Syscall.Duration = time.Since(p.lastSyscallStop.Time)
	} else {
		t.Event = SyscallEnter
	}
	p.lastSyscallStop = t
	return nil
}
