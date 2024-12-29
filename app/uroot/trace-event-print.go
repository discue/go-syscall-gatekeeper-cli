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
	"fmt"
	"io"

	sec "github.com/seccomp/libseccomp-golang"
	"golang.org/x/sys/unix"
)

func signalString(s unix.Signal) string {
	if 0 <= s && int(s) < len(signals) {
		return fmt.Sprintf("%s (%d)", signals[s], int(s))
	}
	return fmt.Sprintf("signal %d", int(s))
}

// PrintTraces prints every trace event to w.
func PrintTraces(w io.Writer) EventCallback {
	return func(t Task, record *TraceRecord) {
		switch record.Event {
		case SyscallEnter:
			fmt.Fprintln(w, SysCallEnter(t, record.Syscall))
		case SyscallExit:
			// fmt.Fprintln(w, SysCallExit(t, record.Syscall))
		case SignalExit:
			fmt.Fprintf(w, "PID %d exited from signal %s\n", record.PID, signalString(record.SignalExit.Signal))
		case Exit:
			fmt.Fprintf(w, "PID %d exited from exit status %d (code = %d)\n", record.PID, record.Exit.WaitStatus, record.Exit.WaitStatus.ExitStatus())
		case SignalStop:
			fmt.Fprintf(w, "PID %d got signal %s\n", record.PID, signalString(record.SignalStop.Signal))
		case NewChild:
			fmt.Fprintf(w, "PID %d spawned new child %d\n", record.PID, record.NewChild.PID)
		}
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
