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
	"runtime"
	"strings"
	"sync/atomic"
	"syscall"

	config "github.com/discue/go-syscall-gatekeeper/app/buildtime-config"
	sec "github.com/seccomp/libseccomp-golang"
	"golang.org/x/sys/unix"
)

var traceActive uint32

type iovec struct {
	P Addr   /* Starting address */
	S uint32 /* Number of bytes to transfer */
}

// Addr is an address for use in strace I/O
type Addr uintptr

func wait(pid int) (int, unix.WaitStatus, error) {
	var w unix.WaitStatus
	pid, err := unix.Wait4(pid, &w, 0, nil)
	return pid, w, err
}

// SignalEvent is a signal that was delivered to the process.
type SignalEvent struct {
	// Signal is the signal number.
	Signal unix.Signal

	// TODO: Add other siginfo_t stuff
}

// ExitEvent is emitted when the process exits regularly using exit_group(2).
type ExitEvent struct {
	// WaitStatus is the exit status.
	WaitStatus unix.WaitStatus
}

// NewChildEvent is emitted when a clone/fork/vfork syscall is done.
type NewChildEvent struct {
	PID int
}

// Trace traces `c` and any children c clones.
//
// Only one trace can be active per process.
//
// recordCallback is called every time a process event happens with the process
// in a stopped state.
func Trace(c *exec.Cmd, recordCallback ...EventCallback) error {
	if !atomic.CompareAndSwapUint32(&traceActive, 0, 1) {
		return fmt.Errorf("a process trace is already active in this process")
	}
	defer func() {
		atomic.StoreUint32(&traceActive, 0)
	}()

	if c.SysProcAttr == nil {
		c.SysProcAttr = &syscall.SysProcAttr{}
	}
	c.SysProcAttr.Ptrace = true

	// Because the go runtime forks traced processes with PTRACE_TRACEME
	// we need to maintain the parent-child relationship for ptrace to work.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := c.Start(); err != nil {
		return err
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
		return err
	} else if ws.TrapCause() != 0 {
		return fmt.Errorf("wait(pid=%d): got %v, want stopped process", c.Process.Pid, ws)
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
		return &TraceError{
			PID: c.Process.Pid,
			Err: os.NewSyscallError("ptrace(PTRACE_SETOPTIONS)", err),
		}
	}

	// Start the process back up.
	if err := unix.PtraceSyscall(c.Process.Pid, 0); err != nil {
		return &TraceError{
			PID: c.Process.Pid,
			Err: fmt.Errorf("failed to resume: %w", err),
		}
	}

	return tracer.runLoop()
}

// EventCallback is a function called on each event while the subject process
// is stopped.
type EventCallback func(t Task, record *TraceRecord) error

// Task is a Linux process.
type Task interface {
	// Read reads from the process at Addr to the interface{}
	// and returns a byte count and error.
	Read(addr Addr, v interface{}) (int, error)

	// Name is a human-readable process identifier. E.g. PID or argv[0].
	Name() string
}

// RecordTraces sends each event on c.
func RecordTraces(c chan<- *TraceRecord) EventCallback {
	return func(t Task, record *TraceRecord) error {
		c <- record
		return nil
	}
}

func signalString(s unix.Signal) string {
	if 0 <= s && int(s) < len(signals) {
		return fmt.Sprintf("%s (%d)", signals[s], int(s))
	}
	return fmt.Sprintf("signal %d", int(s))
}

// PrintTraces prints every trace event to w.
func PrintTraces(w io.Writer) EventCallback {
	return func(t Task, record *TraceRecord) error {
		switch record.Event {
		case SyscallEnter:
			fmt.Fprintln(w, SysCallEnter(t, record.Syscall))
		case SyscallExit:
			fmt.Fprintln(w, SysCallExit(t, record.Syscall))
		case SignalExit:
			fmt.Fprintf(w, "PID %d exited from signal %s\n", record.PID, signalString(record.SignalExit.Signal))
		case Exit:
			fmt.Fprintf(w, "PID %d exited from exit status %d (code = %d)\n", record.PID, record.Exit.WaitStatus, record.Exit.WaitStatus.ExitStatus())
		case SignalStop:
			fmt.Fprintf(w, "PID %d got signal %s\n", record.PID, signalString(record.SignalStop.Signal))
		case NewChild:
			fmt.Fprintf(w, "PID %d spawned new child %d\n", record.PID, record.NewChild.PID)
		}
		return nil
	}
}

// SysCallEnter is called each time a system call enter event happens.
func SysCallEnter(t Task, s *SyscallEvent) string {
	name, _ := sec.ScmpSyscall(s.Regs.Orig_rax).GetName()
	return fmt.Sprintf("enter %s %s", t.Name(), name)
}

// SysCallExit is called each time a system call exit event happens.
func SysCallExit(t Task, s *SyscallEvent) string {
	name, _ := sec.ScmpSyscall(s.Sysno).GetName()
	return fmt.Sprintf("exit %s %s", t.Name(), name)
}

func Exec(c context.Context, bin string, args []string) error {
	cmd := exec.Command(bin, args...)

	// setup goroutines to read and print stdout
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error creating stdout pipe: %w", err)
	}

	if config.SyscallsDelayEnforceUntilCheck == false {
		enforceGatekeeper()

		// if we should enable the gatekeeper via log search string
		// create another goroutine that keeps monitoring stdout
	} else if config.GatekeeperLivenessCheckLogEnabled {
		go func() {
			scanner := bufio.NewScanner(stdoutPipe)
			for scanner.Scan() {
				t := scanner.Text()
				logger.Info(fmt.Sprintf(">> %s", t)) // Print to parent's stdout

				if strings.Contains(t, config.GatekeeperLivenessCheckLogSearchString) {
					logger.Info("Enabling gatekeeper now because log search string was detected.")
					enforceGatekeeper()
					break
				}
			}

			// after we broke the first loop we create another without the if statement
			for scanner.Scan() {
				logger.Info(fmt.Sprintf(">> %s", scanner.Text())) // Print to parent's stdout
			}
		}()
	} else {
		go func() {
			scanner := bufio.NewScanner(stdoutPipe)
			for scanner.Scan() {
				logger.Info(fmt.Sprintf(">> %s", scanner.Text())) // Print to parent's stdout
			}
		}()
	}

	// setup goroutines to read and print errout
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("error creating stderr pipe: %w", err)
	}
	go func() {
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			logger.Error(scanner.Text()) // Print to parent's stderr
		}
	}()

	return Strace(cmd, os.Stdout)
}

// Strace traces and prints process events for `c` and its children to `out`.
func Strace(c *exec.Cmd, out io.Writer) error {
	return Trace(c, PrintTraces(out))
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

// ReadString reads a null-terminated string from the process
// at Addr and any errors.
func ReadString(t Task, addr Addr, max int) (string, error) {
	if addr == 0 {
		return "<nil>", nil
	}
	var s string
	var b [1]byte
	for len(s) < max {
		if _, err := t.Read(addr, b[:]); err != nil {
			return "", err
		}
		if b[0] == 0 {
			break
		}
		s = s + string(b[:])
		addr++
	}
	return s, nil
}

// ReadStringVector takes an address, max string size, and max number of string to read,
// and returns a string slice or error.
func ReadStringVector(t Task, addr Addr, maxsize, maxno int) ([]string, error) {
	var v []Addr
	if addr == 0 {
		return []string{}, nil
	}

	// Read in a maximum of maxno addresses
	for len(v) < maxno {
		var a uint64
		n, err := t.Read(addr, &a)
		if err != nil {
			return nil, fmt.Errorf("could not read vector element at %#x: %w", addr, err)
		}
		if a == 0 {
			break
		}
		addr += Addr(n)
		v = append(v, Addr(a))
	}
	var vs []string
	for _, a := range v {
		s, err := ReadString(t, a, maxsize)
		if err != nil {
			return vs, fmt.Errorf("could not read string at %#x: %w", a, err)
		}
		vs = append(vs, s)
	}
	return vs, nil
}

// CaptureAddress pulls a socket address from the process as a byte slice.
// It returns any errors.
func CaptureAddress(t Task, addr Addr, addrlen uint32) ([]byte, error) {
	b := make([]byte, addrlen)
	if _, err := t.Read(addr, b); err != nil {
		return nil, err
	}
	return b, nil
}
