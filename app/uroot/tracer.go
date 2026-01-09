package uroot

import (
	"context"
	"fmt"
	"strings"
	"syscall"
	"time"

	"github.com/cuandari/lib/app/runtime"
	"github.com/cuandari/lib/app/uroot/syscalls"
	"github.com/cuandari/lib/app/uroot/syscalls/args"
	sec "github.com/seccomp/libseccomp-golang"
	"golang.org/x/sys/unix"
)

type ExitEventError struct {
	ExitCode int
	Signal   string
}

func (e *ExitEventError) Error() string {
	return fmt.Sprintf("Exited with status %d and signal %s", e.ExitCode, e.Signal)
}

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

func (t *tracer) call(p *process, rec *TraceRecord) {
	for _, c := range t.callback {
		c(p, rec)
	}
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

func (t *tracer) runLoop(cancelFunc context.CancelCauseFunc) {
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
			fmt.Printf("All watched processes died: %s. This is not related to the gatekeeper.\n", err.Error())
			cancelFunc(&ExitEventError{
				ExitCode: 3,
			})
		} else if err != nil {
			fmt.Printf("Unable to wait for processed to stop and intercept syscall: %s. Exiting\n", err.Error())
			cancelFunc(&ExitEventError{
				ExitCode: 3,
			})
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
				if !GetIsGatekeeperEnforced() {
					break
				}

				if err := rec.syscallStop(p); err != nil {
					if strings.Contains(err.Error(), "no such process") {
						println(fmt.Sprintf("Error trying to continue pid %d: %s", p.pid, err.Error()))
						// race condition during shutdown of the tracee. do nothing. When calling wait again
						// at the beginning of this loop we will receive the actual exit status
						break

					} else {
						fmt.Printf("Unable to read syscall params and args of pid %d: %s. Exiting\n", p.pid, err.Error())
						cancelFunc(&ExitEventError{
							ExitCode: 3,
						})
					}
				}

				rax := rec.Syscall.Regs.Orig_rax
				name, err := sec.ScmpSyscall(rec.Syscall.Sysno).GetName()
				isExitEvent := rec.Event == SyscallExit
				if err != nil {
					// ending up here, we were not able to get the name of the syscall
					// suspicious because the process might not have been stopped by a syscall
					// but something else, so printing here for now while keeping the tracee
					// up and running
					fmt.Printf("Unknown syscall detected for exit->%t: %s %d\n", isExitEvent, err.Error(), rax)
				} else {
					addSyscallToCollection(rax, name)

					allow := allowSyscall(name)

					// Family-aware gating for socket/connect via helpers
					// Build unified syscall context for helpers
					var sargs syscalls.SyscallArguments
					for i := 0; i < len(sargs); i++ {
						sargs[i] = syscalls.SyscallArgument{Value: rec.Syscall.Args[i].Value}
					}
					s := syscalls.Syscall{
						Args:      sargs,
						TraceePID: p.pid,
						Reader: func(addr syscalls.Addr, v interface{}) (int, error) {
							return p.Read(Addr(addr), v)
						},
					}
					switch {
					case name == "socket":
						allow = syscalls.IsSocketAllowed(s, rec.Event == SyscallEnter)
					case name == "connect":
						allow = syscalls.IsConnectAllowed(s, rec.Event == SyscallEnter)
					// Enforce path-level whitelist for open/openat/openat2 when configured.
					case name == "openat2" &&
						len(runtime.Get().FileSystemAllowedPaths) > 0:
						// pathname is arg 1, dirfd is arg 0
						if !syscalls.PathIsAllowed(s, 1, 0) {
							allow = false
						}
					case name == "openat" &&
						len(runtime.Get().FileSystemAllowedPaths) > 0:
						// pathname is arg 1, dirfd is arg 0
						if !syscalls.PathIsAllowed(s, 1, 0) {
							allow = false
						}
					case name == "openat2" &&
						runtime.Get().FileSystemAllowRead &&
						!runtime.Get().FileSystemAllowWrite:
						// Gate file open syscalls when only read access is allowed.
						allow = syscalls.IsOpenAt2ReadOnly(s, rec.Event == SyscallEnter)
					case name == "openat" &&
						runtime.Get().FileSystemAllowRead &&
						!runtime.Get().FileSystemAllowWrite:
						// Gate file open syscalls when only read access is allowed.
						allow = syscalls.IsOpenAtReadOnly(s, rec.Event == SyscallEnter)
					case name == "open" &&
						runtime.Get().FileSystemAllowRead &&
						!runtime.Get().FileSystemAllowWrite:
						// Gate file open syscalls when only read access is allowed.
						allow = syscalls.IsOpenReadOnly(s, rec.Event == SyscallEnter)
					case name == "write" || name == "writev" || name == "send" || name == "sendmsg" || name == "sendmmsg" || name == "sendto":
						syscallArgs := rec.Syscall.Args
						fd := syscallArgs[0].Int()

						allow = syscalls.IsWriteAllowed(s, rec.Event == SyscallEnter)

						// {
						//     isEventFd := args.FdType(p.pid, fd) == args.FDAnonEvent
						//     println(fmt.Sprintf("Trying to %s from fd %d which is a anon eventfd %t", name, fd, isEventFd))
						//     allow = isEventFd
						// }

						if !allow {
							fdType := args.FdType(p.pid, fd)
							println(fmt.Sprintf("Trying to write to fd %d which is of type %s", fd, fdType))
						}

					case name == "read" || name == "readv" || name == "recv" || name == "recvfrom" || name == "recvmsg" || name == "recvmmsg":
						syscallArgs := rec.Syscall.Args
						fd := syscallArgs[0].Int()

						allow = syscalls.IsReadAllowed(s, rec.Event == SyscallEnter)

						// if !allow {
						//     isEventFd := args.FdType(p.pid, fd) == args.FDAnonEvent
						//     println(fmt.Sprintf("Trying to %s from fd %d which is a anon eventfd %t", name, fd, isEventFd))
						//     allow = isEventFd
						// }

						if !allow {
							fdType := args.FdType(p.pid, fd)
							println(fmt.Sprintf("Trying to read from fd %d which is of type %s", fd, fdType))
						}
					case name == "shutdown":
						// shutdown(int sockfd, int how)
						syscallArgs := rec.Syscall.Args
						fd := syscallArgs[0].Int()

						allow = syscalls.IsShutdownAllowed(s, rec.Event == SyscallEnter)

						if !allow {
							fdType := args.FdType(p.pid, fd)
							println(fmt.Sprintf("Trying to shutdown fd %d which is of type %s", fd, fdType))
						}
					case name == "close":
						// close(int fd)
						syscallArgs := rec.Syscall.Args
						fd := syscallArgs[0].Int()

						allow = syscalls.IsCloseAllowed(s, rec.Event == SyscallEnter)

						if !allow {
							fdType := args.FdType(p.pid, fd)
							println(fmt.Sprintf("Trying to close fd %d which is of type %s", fd, fdType))
						}
					}

					if !allow {
						fmt.Println("Syscall not allowed:", name)
						if runtime.Get().SyscallsDenyTargetIfNotAllowed {
							fmt.Println("Syscall not allowed. However we don't have permission to kill")

							// https://stackoverflow.com/a/6469069/13163094
							switch rec.Event {
							case SyscallEnter:
								// Make sure the syscall is not valid anymore by changing the value that identifies it
								rec.Syscall.Regs.Orig_rax = ^uint64(0)
								rec.Syscall.Regs.Rax = ^uint64(0)
								// In the context of seccomp, SIGSYS is the primary signal used to indicate a policy violation.
								// When a seccomp filter is in place and a process attempts a disallowed system call, the kernel
								// intercepts the call and sends SIGSYS to the process, preventing the system call from executing.
								injectSignal = unix.SIGSYS
							case SyscallExit:
								// Syscall Exit: The kernel sets register rax to the result of the syscall. This is typically 0 for success or -1 (represented as the maximum unsigned integer value) for an error.
								rec.Syscall.Regs.Orig_rax = ^uint64(0)
								rec.Syscall.Regs.Rax = ^uint64(0)
								// Syscall Exit: If Rax indicates an error (-1), Rdx will typically contain the specific error code (the errno) explaining the reason for the failure.
								rec.Syscall.Regs.Rdx = uint64(unix.EPERM) // Set errno
								// In the context of seccomp, SIGSYS is the primary signal used to indicate a policy violation.
								// When a seccomp filter is in place and a process attempts a disallowed system call, the kernel
								// intercepts the call and sends SIGSYS to the process, preventing the system call from executing.
								injectSignal = unix.SIGSYS
							default:
								// Don't send SIGKILL; let the process continue with the simulated error return
								injectSignal = 0
							}

							// Set registers before continuing with the syscall exit.
							if err := unix.PtraceSetRegs(p.pid, &rec.Syscall.Regs); err != nil {
								fmt.Printf("Unable to set syscall params and args: %s. Exiting\n", err.Error())
								cancelFunc(&ExitEventError{
									ExitCode: 3,
								})
							}

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
			case unix.SIGCHLD:

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
						fmt.Printf("Unable to get event message: %s. Exiting\n", err.Error())
						cancelFunc(&ExitEventError{
							ExitCode: 3,
						})
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

		t.call(p, rec)

		if rec.Event == Exit {
			delete(t.processes, pid)
			if len(t.processes) < 1 {
				cancelFunc(&ExitEventError{
					ExitCode: rec.Exit.WaitStatus.ExitStatus(),
				})
				return
			}
		}

		if rec.Event == SignalExit {
			delete(t.processes, pid)
			if len(t.processes) < 1 {
				cancelFunc(&ExitEventError{
					Signal: signalString(rec.SignalExit.Signal),
				})
				return
			}
		}

		if err := p.cont(injectSignal); err != nil {
			if strings.Contains(err.Error(), "no such process") {
				println(fmt.Sprintf("Error trying to continue pid %d: %s", p.pid, err.Error()))
				// race condition during shutdown of the tracee. do nothing. When calling wait again
				// at the beginning of this loop we will receive the actual exit status
			} else {
				fmt.Printf("Unable to continue process with pid %d: %s. Exiting\n", p.pid, err.Error())
				cancelFunc(&ExitEventError{
					ExitCode: 3,
				})
			}
		}

		// if another child was started, we need to continue the child, too
		if rec.Event == NewChild {
			// Which process was stopped?
			p, ok := t.processes[rec.NewChild.PID]
			if !ok {
				continue
			}

			e := p.cont(0)
			if e != nil {
				fmt.Printf("Error continuing child process %d: %s\n", pid, e.Error())
			}
		}
	}
}

func wait(pid int) (int, unix.WaitStatus, error) {
	var w unix.WaitStatus
	pid, err := unix.Wait4(pid, &w, 0, nil)
	return pid, w, err
}
