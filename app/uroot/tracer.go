package uroot

import (
	"context"
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/discue/go-syscall-gatekeeper/app/runtime"
	"github.com/discue/go-syscall-gatekeeper/app/uroot/syscalls/args"
	sec "github.com/seccomp/libseccomp-golang"
	"golang.org/x/sys/unix"
)

type ExitEventError struct {
	ExitEvent *ExitEvent
}

func (e *ExitEventError) Error() string {
	return fmt.Sprintf("Exited with status %d", e.ExitEvent.WaitStatus.ExitStatus())
}

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

type tracer struct {
	processes map[int]*process
	callback  []EventCallback
}

// SignalEvent is a signal that was delivered to the process.
type SignalEvent struct {
	// Signal is the signal number.
	Signal unix.Signal

	// TODO: Add other siginfo_t stuff
}

func (t *tracer) call(p *process, rec *TraceRecord) error {
	for _, c := range t.callback {
		if err := c(p, rec); err != nil {
			return err
		}
	}
	return nil
}

func (t *tracer) addProcess(pid int, event EventType) {
	t.processes[pid] = &process{
		pid: pid,
		lastSyscallStop: &TraceRecord{
			Event: event,
			Time:  time.Now(),
		},
	}
}

func (t *tracer) runLoop(cancelFunc context.CancelCauseFunc) error {
	for {
		// TODO: we cannot have any other children. I'm not sure this
		// is actually solvable: if we used a session or process group,
		// a tracee process's usage of them would mess up our accounting.
		//
		// If we just ignored wait's of processes that we're not
		// tracing, we'll be messing up other stuff in this program
		// waiting on those.
		//
		// To actually encapsulate this library in a packge, we could
		// do one of two things:
		//
		//   1) fork from the parent in order to be able to trace
		//      children correctly. Then, a user of this library could
		//      actually independently trace two different processes.
		//      I don't know if that's worth doing.
		//   2) have one goroutine per process, and call wait4
		//      individually on each process we expect. We gotta check
		//      if each has to be tied to an OS thread or not.
		//
		// The latter option seems much nicer.
		pid, status, err := wait(-1)
		if err == unix.ECHILD {
			// All our children are gone.
			return nil
		} else if err != nil {
			return os.NewSyscallError("wait4", err)
		}

		// Which process was stopped?
		p, ok := t.processes[pid]
		if !ok {
			continue
		}

		rec := &TraceRecord{
			PID:  p.pid,
			Time: time.Now(),
		}

		var injectSignal unix.Signal
		if status.Exited() {
			rec.Event = Exit
			rec.Exit = &ExitEvent{
				WaitStatus: status,
			}
		} else if status.Signaled() {
			rec.Event = SignalExit
			rec.SignalExit = &SignalEvent{
				Signal: status.Signal(),
			}
		} else if status.Stopped() {
			// Ptrace stops kinds.
			switch signal := status.StopSignal(); signal {
			// Syscall-stop.
			//
			// Setting PTRACE_O_TRACESYSGOOD means StopSignal ==
			// SIGTRAP|0x80 (0x85) for syscall-stops.
			//
			// It allows us to distinguish syscall-stops from regular
			// SIGTRAPs (e.g. sent by tkill(2)).
			case syscall.SIGTRAP | 0x80:
				if err := rec.syscallStop(p); err != nil {
					return err
				}

				rax := rec.Syscall.Regs.Orig_rax
				name, err := sec.ScmpSyscall(rec.Syscall.Sysno).GetName()
				if err != nil {
					// ending up here, we were not able to get the name of the syscall
					// suspicious because the process might not have been stopped by a syscall
					// but something else, so printing here for now while keeping the tracee
					// up and running
					fmt.Printf("Unknown syscall detected: %s %d\n", err.Error(), rax)
				} else {
					addSyscallToCollection(rax, name)

					allow := allowSyscall(name)

					if !allow {
						if name == "write" {
							syscallArgs := rec.Syscall.Args
							fd := syscallArgs[0].Int()
							isStdStream := args.IsStandardStream(fd)
							println(fmt.Sprintf("Trying to write to fd %d which is std stream %t", fd, isStdStream))
							if isStdStream {
								// we allow writing to standard streams
								allow = true
							}
						}
					} else {
						if name == "openat" &&
							runtime.Get().FileSystemAllowRead &&
							!runtime.Get().FileSystemAllowWrite {
							// if runtime.Get().FilesystemAllowRead && !runtime.Get().FilesystemAllowWrite {

							args := rec.Syscall.Args
							mode := args[3].ModeT() // Assuming mode_t is represented as uint

							accessMode := ""
							if mode&unix.O_RDONLY == unix.O_RDONLY {
								accessMode = "read"
							}
							if mode&unix.O_WRONLY == unix.O_WRONLY {
								accessMode = "write"
							}
							if mode&unix.O_RDWR == unix.O_RDWR {
								accessMode = "write"
							}
							if mode&unix.O_RDWR == unix.O_APPEND {
								accessMode = "write"
							}
							if mode&unix.O_RDWR == unix.O_CREAT {
								accessMode = "write"
							}
							if mode&unix.O_RDWR == unix.O_TRUNC {
								accessMode = "write"
							}

							println(fmt.Printf("access mode is %s\n", accessMode))

							if accessMode != "read" {
								allow = false
							}
						} else if name == "writev" || name == "sendto" || name == "recvmsg" || name == "recvfrom" {
							if runtime.Get().NetworkAllowServer || runtime.Get().NetworkAllowClient {
								syscallArgs := rec.Syscall.Args
								fd := syscallArgs[0].Int()

								isSocket := args.IsSocket(p.pid, fd)
								println(fmt.Sprintf("Trying to writev to fd %d which is socket %t", fd, isSocket))
								if !isSocket {
									// we allow writing to sockets
									allow = false
								}
							} else if runtime.Get().FileSystemAllowWrite {
								syscallArgs := rec.Syscall.Args
								fd := syscallArgs[0].Int()

								isFile := args.IsFile(p.pid, fd)
								println(fmt.Sprintf("Trying to writev to fd %d which is socket %t", fd, isFile))
								if isFile {
									// we allow writing to sockets
									allow = true
								}
							}
						}
					}

					if !allow {
						fmt.Println("Syscall not allowed:", name)
						if runtime.Get().SyscallsDenyTargetIfNotAllowed {
							fmt.Println("Syscall not allowed. However we don't have permission to kill")

							if rec.Event == SyscallEnter {
								// Set the return value registers to indicate an error (-1 and errno)
								rec.Syscall.Regs.Rax = ^uint64(0) // Return -1 for error
								// Choose an appropriate errno (e.g., EPERM for permission denied)
								// rec.Syscall.Regs.Rax = uint64(unix.EPERM) // Set errno
							} else if rec.Event == SyscallExit {
								// Syscall Exit: The kernel sets register rax to the result of the syscall. This is typically 0 for success or -1 (represented as the maximum unsigned integer value) for an error.
								rec.Syscall.Regs.Rax = ^uint64(0) // Set errno
								// Syscall Exit: If Rax indicates an error (-1), Rdx will typically contain the specific error code (the errno) explaining the reason for the failure.
								rec.Syscall.Regs.Rdx = uint64(unix.EPERM) // Set errno
							}

							// Set registers before continuing with the syscall exit.
							if err := unix.PtraceSetRegs(p.pid, &rec.Syscall.Regs); err != nil {
								return &TraceError{PID: p.pid, Err: fmt.Errorf("failed to set registers: %v", err)}
							}

							// Don't send SIGKILL; let the process continue with the simulated error return
							injectSignal = 0

						} else {
							injectSignal = syscall.SIGKILL
						}
					}
				}

			// Group-stop, but also a special stop: first stop after
			// fork/clone/vforking a new task.
			//
			// TODO: is that different than a group-stop, or the same?
			case syscall.SIGSTOP:
				// TODO: have a list of expected children SIGSTOPs, and
				// make events only for all the unexpected ones.
				fallthrough

			// Group-stop.
			//
			// TODO: do something.
			case syscall.SIGTSTP, syscall.SIGTTOU, syscall.SIGTTIN:
				rec.Event = SignalStop
				injectSignal = signal
				rec.SignalStop = &SignalEvent{
					Signal: signal,
				}

				// TODO: Do we have to use PTRACE_LISTEN to
				// restart the task in order to keep the task
				// in stopped state, as expected by whomever
				// sent the stop signal?

			// Either a regular signal-delivery-stop, or a PTRACE_EVENT stop.
			case syscall.SIGTRAP:
				switch tc := status.TrapCause(); tc {
				// This is a PTRACE_EVENT stop.
				case unix.PTRACE_EVENT_CLONE, unix.PTRACE_EVENT_FORK, unix.PTRACE_EVENT_VFORK:
					childPID, err := unix.PtraceGetEventMsg(pid)
					if err != nil {
						return &TraceError{
							PID: pid,
							Err: os.NewSyscallError("ptrace(PTRACE_GETEVENTMSG)", err),
						}
					}
					// The first event will be an Enter syscall, so
					// set the last event to an exit.
					t.addProcess(int(childPID), SyscallExit)

					rec.Event = NewChild
					rec.NewChild = &NewChildEvent{
						PID: int(childPID),
					}

				// Regular signal-delivery-stop.
				default:
					rec.Event = SignalStop
					rec.SignalStop = &SignalEvent{
						Signal: signal,
					}
					injectSignal = signal
				}

			// Signal-delivery-stop.
			default:
				rec.Event = SignalStop
				rec.SignalStop = &SignalEvent{
					Signal: signal,
				}
				injectSignal = signal
			}
		} else {
			rec.Event = Unknown
		}

		if err := t.call(p, rec); err != nil {
			return err
		}

		if rec.Event == Exit {
			delete(t.processes, pid)
			if len(t.processes) < 1 {
				cancelFunc(&ExitEventError{
					ExitEvent: rec.Exit,
				})
			}
			continue
		}

		if rec.Event == SignalExit {
			delete(t.processes, pid)
			if len(t.processes) < 1 {
				cancelFunc(&ExitEventError{
					ExitEvent: &ExitEvent{
						Signal: signalString(rec.SignalExit.Signal),
					},
				})
			}
			continue
		}

		if err := p.cont(injectSignal); err != nil {
			return err
		}
	}
}

func wait(pid int) (int, unix.WaitStatus, error) {
	var w unix.WaitStatus
	pid, err := unix.Wait4(pid, &w, 0, nil)
	return pid, w, err
}
