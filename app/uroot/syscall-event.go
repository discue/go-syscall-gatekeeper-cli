package uroot

import (
	"time"

	"golang.org/x/sys/unix"
)

// SyscallEvent is populated for both SyscallEnter and SyscallExit event types.
type SyscallEvent struct {
	// Regs are the process's registers as they were when the event was
	// recorded.
	Regs unix.PtraceRegs

	// Sysno is the syscall number.
	Sysno int

	// Args are the arguments to the syscall.
	Args SyscallArguments

	// Ret is the return value of the syscall. Only populated on
	// SyscallExit.
	Ret [2]SyscallArgument

	// Errno is an errno, if there was on in Ret. Only populated on
	// SyscallExit.
	Errno unix.Errno

	// Duration is the duration from enter to exit for this particular
	// syscall. Only populated on SyscallExit.
	Duration time.Duration
}

// FillArgs pulls the correct registers to populate system call arguments
// and the system call number into a TraceRecord. Note that the system
// call number is not technically an argument. This is good, in a sense,
// since it makes the function arguments end up in "the right place"
// from the point of view of the caller. The performance improvement is
// negligible, as you can see by a look at the GNU runtime.
func (s *SyscallEvent) FillArgs() {
	s.Args = SyscallArguments{
		{uintptr(s.Regs.Rdi)},
		{uintptr(s.Regs.Rsi)},
		{uintptr(s.Regs.Rdx)},
		{uintptr(s.Regs.R10)},
		{uintptr(s.Regs.R8)},
		{uintptr(s.Regs.R9)},
	}
	s.Sysno = int(uint32(s.Regs.Orig_rax))
}

// FillRet fills the TraceRecord with the result values from the registers.
func (s *SyscallEvent) FillRet() {
	s.Ret = [2]SyscallArgument{{uintptr(s.Regs.Rax)}, {uintptr(s.Regs.Rdx)}}
	if errno := int(s.Regs.Rax); errno < 0 {
		s.Errno = unix.Errno(-errno)
	}
}
