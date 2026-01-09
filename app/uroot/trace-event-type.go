package uroot

// EventType describes a process event.
type EventType int

const (
	// Unknown is for events we do not know how to interpret.
	Unknown EventType = 0x0

	// SyscallEnter is the event for a process calling a syscall.  Event
	// Args will contain the arguments sent by the userspace process.
	//
	// ptrace calls this a syscall-enter-stop.
	SyscallEnter EventType = 0x2

	// SyscallExit is the event for the kernel returning a syscall. Args
	// will contain the arguments as returned by the kernel.
	//
	// ptrace calls this a syscall-exit-stop.
	SyscallExit EventType = 0x3

	// SignalExit means the process has been terminated by a signal.
	SignalExit EventType = 0x4

	// Exit means the process has exited with an exit code.
	Exit EventType = 0x5

	// SignalStop means the process was stopped by a signal.
	//
	// ptrace calls this a signal-delivery-stop.
	SignalStop EventType = 0x6

	// NewChild means the process created a new child thread or child
	// process via fork, clone, or vfork.
	//
	// ptrace calls this a PTRACE_EVENT_(FORK|CLONE|VFORK).
	NewChild EventType = 0x7
)
