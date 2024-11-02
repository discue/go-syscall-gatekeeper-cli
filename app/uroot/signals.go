package uroot

import (
	"golang.org/x/sys/unix"
)

// Signal table (taken from Go runtime, but adding all-caps signal names)
var signals = [...]string{
	unix.SIGHUP:  "SIGHUP (hangup)",
	unix.SIGINT:  "SIGINT (interrupt)",
	unix.SIGQUIT: "SIGQUIT (quit)",
	unix.SIGILL:  "SIGILL (illegal instruction)",
	unix.SIGTRAP: "SIGTRAP (trace/breakpoint trap)",
	unix.SIGABRT: "SIGABRT (aborted)",
	unix.SIGBUS:  "SIGBUS (bus error)",
	unix.SIGFPE:  "SIGFPE (floating point exception)",
	unix.SIGKILL: "SIGKILL (killed)",
	unix.SIGUSR1: "SIGUSR1 (user defined signal 1)",
	unix.SIGSEGV: "SIGSEGV (segmentation fault)",
	unix.SIGUSR2: "SIGUSR2 (user defined signal 2)",
	unix.SIGPIPE: "SIGPIPE (broken pipe)",
	unix.SIGALRM: "SIGALRM (alarm clock)",
	unix.SIGTERM: "SIGTERM (terminated)",
	16:           "SIGSTKFLT (stack fault)",
	17:           "SIGCHLD (child exited)",
	18:           "SIGCONT (continued)",
	19:           "SIGSTOP (stopped)",
	20:           "SIGTSTP (stopped)",
	21:           "SIGTTIN (stopped - tty input)",
	22:           "SIGTTOU (stopped - tty output)",
	23:           "SIGURG (urgent I/O condition)",
	24:           "SIGXCPU (CPU time limit exceeded)",
	25:           "SIGXFSZ (file size limit exceeded)",
	26:           "SIGVTALRM (virtual timer expired)",
	27:           "SIGPROF (profiling timer expired)",
	28:           "SIGWINCH (window changed)",
	29:           "SIGPOLL (I/O possible)",
	30:           "SIGPWR (power failure)",
	31:           "SIGSYS (bad system call)",
}
