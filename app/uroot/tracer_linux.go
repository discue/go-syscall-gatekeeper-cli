package uroot

import (
	"os"
	"syscall"
	"time"

	"github.com/discue/go-syscall-gatekeeper/app/runtime"
	sec "github.com/seccomp/libseccomp-golang"
	"golang.org/x/sys/unix"
)

type tracer struct {
	processes map[int]*process
	callback  []EventCallback
	stop      bool
}

func (t *tracer) terminate() {
	// for pid, _ := range t.processes {
	// 	logger.Info(fmt.Sprintf("Terminating pid %d", pid))
	// 	syscall.Kill(pid, syscall.SIGINT)
	// }
	t.stop = true
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

func (t *tracer) runLoop() error {
	for t.stop == false {
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
				name, _ := sec.ScmpSyscall(rax).GetName()

				addSyscallToCollection(rax, name)
				if runtime.Get().SyscallsKillTargetIfNotAllowed {
					if runtime.Get().SyscallsAllowMap[name] == false {
						injectSignal = syscall.SIGKILL
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

		if rec.Event == SignalExit || rec.Event == Exit {
			delete(t.processes, pid)
			continue
		}

		if err := p.cont(injectSignal); err != nil {
			return err
		}
	}

	return nil
}
