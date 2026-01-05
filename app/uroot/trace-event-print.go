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
	"os"

	// "github.com/cuandari/lib/app/uroot/syscalls/args"
	"github.com/cuandari/lib/app/uroot/syscalls/args"
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
			// _, _ = fmt.Fprintln(w, SysCallEnter(t, record.Syscall))
		case SyscallExit:
			// _, _ = fmt.Fprintln(w, SysCallExit(t, record.Syscall))
		case SignalExit:
			_, _ = fmt.Fprintf(w, "PID %d exited from signal %s\n", record.PID, signalString(record.SignalExit.Signal))
		case Exit:
			_, _ = fmt.Fprintf(w, "PID %d exited from exit status %d (code = %d)\n", record.PID, record.Exit.WaitStatus, record.Exit.WaitStatus.ExitStatus())
		case SignalStop:
			_, _ = fmt.Fprintf(w, "PID %d got signal %s\n", record.PID, signalString(record.SignalStop.Signal))
		case NewChild:
			_, _ = fmt.Fprintf(w, "PID %d spawned new child %d\n", record.PID, record.NewChild.PID)
		}
	}
}

// SysCallEnter is called each time a system call enter event happens.
func SysCallEnter(t Task, s *SyscallEvent) string {
	pid := t.(*process).pid
	name, _ := sec.ScmpSyscall(s.Regs.Orig_rax).GetName()
	if name == "access" {
		pathAddr := s.Args[0].Pointer()
		var path string
		var b [1]byte
		for len(path) < 1024 {
			if _, err := t.Read(pathAddr, b[:]); err != nil {
				break
			}
			if b[0] == 0 {
				break
			}
			path = path + string(b[:])
			pathAddr++
		}

		return fmt.Sprintf("enter %s %s (%s)", t.Name(), name, path)

	} else if name == "fstat" || name == "close" || name == "read" || name == "ioctl" || name == "pread64" || name == "getdents" || name == "getdents64" {
		dirfd := s.Args[0].Int()
		var fdPath string
		fdPath = fmt.Sprintf("/proc/%d/fd/%d", t.(*process).pid, dirfd)

		// Get path from /proc
		if dirfd == unix.AT_FDCWD {
			fdPath, _ = os.Getwd()
		} else {
			pid := t.(*process).pid
			if args.IsSocket(pid, dirfd) {
				println("Closing network socket with fd %d", dirfd)
				fdPath = "socket"
			} else if args.IsFile(pid, dirfd) || args.IsDir(pid, dirfd) {
				if resolvedPath, err := os.Readlink(fdPath); err != nil {
					fdPath = resolvedPath
				}
			}
		}

		return fmt.Sprintf("enter %s %s (%s)", t.Name(), name, fdPath)

	} else if name == "openat" {
		dirfd := s.Args[0].Int()
		pathAddr := s.Args[1].Pointer()
		flags := s.Args[2].Uint() // or Int() depending on your usage
		mode := s.Args[3].ModeT() // Assuming mode_t is represented as uint

		var path string
		var b [1]byte
		for len(path) < 1024 {
			if _, err := t.Read(pathAddr, b[:]); err != nil {
				break
			}
			if b[0] == 0 {
				break
			}
			path = path + string(b[:])
			pathAddr++
		}

		accessMode := ""
		if flags&unix.O_RDONLY == unix.O_RDONLY {
			accessMode = "read"
		}
		if flags&unix.O_WRONLY == unix.O_WRONLY {
			accessMode = "write"
		}
		if flags&unix.O_RDWR == unix.O_RDWR {
			accessMode = "write"
		}
		if flags&unix.O_RDWR == unix.O_APPEND {
			accessMode = "write"
		}
		if flags&unix.O_RDWR == unix.O_CREAT {
			accessMode = "write"
		}
		if flags&unix.O_RDWR == unix.O_TRUNC {
			accessMode = "write"
		}

		fdType := "fd"

		// Get path from /proc
		var fdPath string
		if dirfd == unix.AT_FDCWD {
			fdPath, _ = os.Getwd()
		} else {
			fdPath = fmt.Sprintf("/proc/%d/fd/%d", t.(*process).pid, dirfd)
			if resolvedPath, err := os.Readlink(fdPath); err == nil {
				fdPath = resolvedPath
			} else {
				println(fmt.Printf("error reading fd path: %v\n", err))
			}
		}

		return fmt.Sprintf("enter %s %s(%d:%s,%s, %q, %s, %#o)", t.Name(), name, dirfd, fdType, fdPath, path, accessMode, mode)
	} else if name == "write" {
		dirfd := s.Args[0].Int()

		fdType := "fd"
		if args.IsStdIn(dirfd) {
			fdType = " (stdin)"
		} else if args.IsStdOut(dirfd) {
			fdType = " (stdout)"
		} else if args.IsStdErr(dirfd) {
			fdType = " (stderr)"
		} else if args.IsFile(pid, dirfd) {
			fdType = " (file)"
		} else if args.IsDir(pid, dirfd) {
			fdType = " (directory)"
		} else if args.IsSocket(pid, dirfd) {
			fdType = " (socket)"
		} else if args.IsPipe(pid, dirfd) {
			fdType = " (pipe)"
		} else if args.IsBlockDevice(pid, dirfd) {
			fdType = " (block)"
		} else if args.IsCharDevice(pid, dirfd) {
			fdType = " (char)"
		}

		return fmt.Sprintf("enter %s %s(%d, %s)", t.Name(), name, dirfd, fdType)
	}
	return fmt.Sprintf("enter %s %s", t.Name(), name)
}

// SysCallExit is called each time a system call exit event happens.
func SysCallExit(t Task, s *SyscallEvent) string {
	name, _ := sec.ScmpSyscall(s.Sysno).GetName()
	return fmt.Sprintf("exit %s %s", t.Name(), name)
}
