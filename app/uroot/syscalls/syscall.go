package syscalls

// Syscall bundles arguments and minimal tracee context so helpers share one signature.
type Syscall struct {
	// Args are the arguments to the syscall.
	Args SyscallArguments

	// TraceePID is the pid of the process performing the syscall.
	TraceePID int

	// Reader allows helpers to read tracee memory (e.g., sockaddr, open_how).
	// Provide as a function to avoid cross-package type coupling.
	Reader func(addr Addr, v interface{}) (int, error)
}
