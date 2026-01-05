// Copyright 2018 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build (linux && arm64) || (linux && amd64) || (linux && riscv64)

// Package strace traces Linux process events.
//
// An straced process will emit events for syscalls, signals, exits, and new
// children.
package uroot

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	runtimeConfig "github.com/cuandari/lib/app/runtime"
	"github.com/cuandari/lib/app/uroot/stdout"
	"github.com/cuandari/lib/app/utils"

	"golang.org/x/sys/unix"
)

// ExitEvent is emitted when the process exits regularly using exit_group(2).
type ExitEvent struct {
	// WaitStatus is the exit status.
	WaitStatus unix.WaitStatus
}

// NewChildEvent is emitted when a clone/fork/vfork syscall is done.
type NewChildEvent struct {
	PID int
}

// EventCallback is a function called on each event while the subject process
// is stopped.
type EventCallback func(t Task, record *TraceRecord)

// Task is a Linux process.
type Task interface {
	// Read reads from the process at Addr to the interface{}
	// and returns a byte count and error.
	Read(addr Addr, v interface{}) (int, error)

	// Name is a human-readable process identifier. E.g. PID or argv[0].
	Name() string
}

func Exec(ctx context.Context, bin string, args []string) (*exec.Cmd, context.Context, error) {
	executable, err := exec.LookPath(bin)
	if err != nil {
		cancelledContext, cancel := context.WithCancelCause(context.Background())
		cancel(&ExitEventError{
			ExitCode: 12,
		})
		return nil, cancelledContext, fmt.Errorf("unable to find executable %s: %w", bin, err)
	}
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Env = getEnv(os.Getpid())
	cmd.WaitDelay = 5 * time.Second
	cmd.Cancel = func() error {
		return syscall.Kill(cmd.Process.Pid, syscall.SIGTERM)
	}

	if runtimeConfig.Get().ExecutionMode == runtimeConfig.EXECUTION_MODE_TRACE {
		// nolint:gosimple
		go func() {
			<-ctx.Done()
			f, _ := os.Create("gk-syscalls-before-enforce.txt")
			for k := range syscallsBeforeEnforce {
				_, _ = f.WriteString(k)
				_, _ = f.WriteString("\n")
			}
			f, _ = os.Create("gk-syscalls-after-enforce.txt")
			for k := range syscallsAfterEnforce {
				_, _ = f.WriteString(k)
				_, _ = f.WriteString("\n")
			}
		}()
	}

	// setup goroutines to read and print stdout
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		cancelledContext, cancel := context.WithCancelCause(context.Background())
		cancel(&ExitEventError{
			ExitCode: 13,
		})
		return nil, cancelledContext, fmt.Errorf("error creating stdout pipe: %w", err)
	}

	if runtimeConfig.Get().EnforceOnStartup {
		enforceGatekeeper()

		stdout.PipeStdOut(ctx, stdoutPipe)
		// if we should enable the gatekeeper via log search string
		// create another goroutine that keeps monitoring stdout
	} else if runtimeConfig.Get().TriggerEnforceLogMatch != "" {
		go func() {
			scanner := bufio.NewScanner(stdoutPipe)
			for scanner.Scan() {
				if scanner.Err() != nil {
					break
				}

				brkLoop := false
				select {
				case <-ctx.Done():
					utils.SafeClose(stdoutPipe, "stdout pipe (trigger-enforce log match)")
					brkLoop = true
				default:
					t := scanner.Text()
					_, _ = os.Stdout.WriteString(t)
					_, _ = os.Stdout.WriteString("\n")

					if strings.Contains(t, runtimeConfig.Get().TriggerEnforceLogMatch) {
						println("Enabling gatekeeper now because log search string was detected.")
						enforceGatekeeper()
						brkLoop = true
					}
				}

				if brkLoop {
					break
				}
			}

			stdout.PipeStdOut(ctx, stdoutPipe)
		}()
	} else if runtimeConfig.Get().TriggerEnforceSignal != "" {
		go func() {
			signalChan := make(chan os.Signal, 1)
			signal.Notify(signalChan, unix.SignalNum(runtimeConfig.Get().TriggerEnforceSignal))

			<-signalChan
			println("Enabling gatekeeper now because signal was detected.")
			enforceGatekeeper()

			stdout.PipeStdOut(ctx, stdoutPipe)
		}()
	} else {
		stdout.PipeStdOut(ctx, stdoutPipe)
	}

	// setup goroutines to read and print errout
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		cancelledContext, cancel := context.WithCancelCause(context.Background())
		cancel(&ExitEventError{
			ExitCode: 14,
		})
		return nil, cancelledContext, fmt.Errorf("error creating stderr pipe: %w", err)
	}
	stdout.PipeStdErr(ctx, stderrPipe)

	exitContext, cancel := context.WithCancelCause(ctx)
	go func() {
		Strace(cmd, cancel, os.Stdout)
	}()

	return cmd, exitContext, nil
}

// Strace traces and prints process events for `c` and its children to `out`.
func Strace(c *exec.Cmd, cancelFunc context.CancelCauseFunc, out io.Writer) {
	Trace(c, cancelFunc, PrintTraces(out))
}

// Trace traces `c` and any children c clones.
//
// Only one trace can be active per process.
//
// recordCallback is called every time a process event happens with the process
// in a stopped state.
func Trace(c *exec.Cmd, cancelFunc context.CancelCauseFunc, recordCallback ...EventCallback) {
	if c.SysProcAttr == nil {
		c.SysProcAttr = &syscall.SysProcAttr{}
	}
	c.SysProcAttr.Ptrace = true

	// Because the go runtime forks traced processes with PTRACE_TRACEME
	// we need to maintain the parent-child relationship for ptrace to work.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := c.Start(); err != nil {
		fmt.Printf("Unable to start process %s\n", err.Error())
		cancelFunc(&ExitEventError{
			ExitCode: 2,
		})
	}

	tracer := &tracer{
		processes: make(map[int]*process),
		callback:  recordCallback,
	}

	// Start will fork, set PTRACE_TRACEME, and then execve. Once that
	// happens, we should be stopped at the execve "exit". This wait will
	// return at that exit point.
	//
	// The new task image has been loaded at this point, with us just about
	// to jump into _start.
	//
	// It'd make sense to assume, but this stop is NOT a syscall-exit-stop
	// of the execve. It is a signal-stop triggered at the end of execve,
	// within the confines of the new task image.  This means the execve
	// syscall args are not in their registers, and we can't print the
	// exit.
	//
	// NOTE(chrisko): we could make it such that we can read the args of
	// the execve. If we were to signal ourselves between PTRACE_TRACEME
	// and execve, we'd stop before the execve and catch execve as a
	// syscall-stop after. To do so, we have 3 options: (1) write a copy of
	// stdlib exec.Cmd.Start/os.StartProcess with the change, or (2)
	// upstreaming a change that would make it into the next Go version, or
	// (3) use something other than *exec.Cmd as the API.
	//
	// A copy of the StartProcess logic would be tedious, an upstream
	// change would take a while to get into Go, and we want this API to be
	// easily usable. I think it's ok to sacrifice the execve for now.
	if _, ws, err := wait(c.Process.Pid); err != nil {
		fmt.Printf("Received error while waiting for pid %d to stop for ptrace injection: %s\n", c.Process.Pid, err.Error())
		cancelFunc(&ExitEventError{
			ExitCode: 2,
		})
		return
	} else if ws.TrapCause() != 0 {
		fmt.Printf("Expected pid %d to be stopped but got %#+v\n", c.Process.Pid, ws)
		cancelFunc(&ExitEventError{
			ExitCode: 2,
		})
		return
	}
	tracer.addProcess(c.Process.Pid, SyscallExit)

	if err := unix.PtraceSetOptions(c.Process.Pid,
		// Tells ptrace to generate a SIGTRAP signal immediately before a new program is executed with the execve system call.
		unix.PTRACE_O_TRACEEXEC|
			// Make it easy to distinguish syscall-stops from other SIGTRAPS.
			unix.PTRACE_O_TRACESYSGOOD|
			// Kill tracee if tracer exits.
			unix.PTRACE_O_EXITKILL|
			// Automatically trace fork(2)'d, clone(2)'d, and vfork(2)'d children.
			unix.PTRACE_O_TRACECLONE|unix.PTRACE_O_TRACEFORK|unix.PTRACE_O_TRACEVFORK); err != nil {

		fmt.Printf("Unable to set ptrace options %s\n", err.Error())
		cancelFunc(&ExitEventError{
			ExitCode: 2,
		})
		return
	}

	// Start the process back up.
	if err := unix.PtraceSyscall(c.Process.Pid, 0); err != nil {
		fmt.Printf("Unable to resume process %d: %s\n", c.Process.Pid, err.Error())
		cancelFunc(&ExitEventError{
			ExitCode: 2,
		})
		return
	}

	tracer.runLoop(cancelFunc)
}

func getEnv(pid int) []string {
	envVar := fmt.Sprintf("GATEKEEPER_PID=%d", pid)
	env := os.Environ()
	env = append(env, envVar)
	return env
}
