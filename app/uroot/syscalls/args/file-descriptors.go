package args

import (
	"fmt"
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

const (
	FDUnknown       = "unknown"
	FDFile          = "file"
	FDDir           = "directory"
	FDSymlink       = "symlink"
	FDCharDevice    = "char device"
	FDBlockDevice   = "block device"
	FDSocket        = "socket"
	FDPipe          = "pipe"
	FDAnonEvent     = "anon eventfd"
	FDAnonEventPoll = "anon eventpoll"
	FDAnonIoUring   = "anon io_uring"
)

var (
	// store fds for standard streams for faster access
	stdIn  = os.Stdin.Fd()
	stdOut = os.Stdout.Fd()
	stdErr = os.Stderr.Fd()
)

func IsStdIn(fd int32) bool {
	return fd == int32(stdIn)
}

func IsStdOut(fd int32) bool {
	return fd == int32(stdOut)
}

func IsStdErr(fd int32) bool {
	return fd == int32(stdErr)
}

func IsStandardStream(fd int32) bool {
	return IsStdIn(fd) || IsStdOut(fd) || IsStdErr(fd)
}

func LstatFd(pid int, fd int32) (unix.Stat_t, error) {
	path := fmt.Sprintf("/proc/%d/fd/%d", pid, fd)
	var stat unix.Stat_t
	err := unix.Stat(path, &stat) // Fstat takes int, not int32
	return stat, err
}

func isFdType(pid int, fd int32, fdConstant uint32) bool {
	stat, err := LstatFd(pid, fd)
	if err != nil {
		fmt.Printf("error stating fd %d: %v\n", fd, err)
		return false
	}
	return stat.Mode&fdConstant == fdConstant
}

func IsFile(pid int, fd int32) bool {
	return isFdType(pid, fd, syscall.S_IFREG)
}

func IsDir(pid int, fd int32) bool {
	return isFdType(pid, fd, syscall.S_IFDIR)
}

func IsSymlink(pid int, fd int32) bool {
	return isFdType(pid, fd, syscall.S_IFLNK)
}

func IsBlockDevice(pid int, fd int32) bool {
	return isFdType(pid, fd, syscall.S_IFBLK)
}

func IsCharDevice(pid int, fd int32) bool {
	return isFdType(pid, fd, syscall.S_IFCHR)
}

func IsSocket(pid int, fd int32) bool {
	return isFdType(pid, fd, syscall.S_IFSOCK)
}

func IsPipe(pid int, fd int32) bool {
	return isFdType(pid, fd, syscall.S_IFIFO)
}

func FdType(pid int, fd int32) string {
	stat, err := LstatFd(pid, fd)
	if err != nil {
		fmt.Printf("error stating fd %d: %v\n", fd, err)
		return FDUnknown
	}

	// Check file type using stat.Mode
	switch stat.Mode & syscall.S_IFMT {
	case syscall.S_IFREG:
		return FDFile
	case syscall.S_IFDIR:
		return FDDir
	case syscall.S_IFLNK:
		return FDSymlink
	case syscall.S_IFCHR:
		return FDCharDevice
	case syscall.S_IFBLK:
		return FDBlockDevice
	case syscall.S_IFSOCK:
		return FDSocket
	case syscall.S_IFIFO:
		return FDPipe
	}

	// Fallback: Use readlink to determine the type
	filePath := fmt.Sprintf("/proc/%d/fd/%d", pid, fd)
	link, err := os.Readlink(filePath)
	if err != nil {
		fmt.Printf("error reading link for fd %d: %v\n", fd, err)
		return FDUnknown
	}

	// Check for specific anon_inode types
	switch link {
	case "anon_inode:[eventfd]":
		return FDAnonEvent
	case "anon_inode:[eventpoll]":
		return FDAnonEventPoll
	case "anon_inode:[io_uring]":
		return FDAnonIoUring
	}

	return fmt.Sprintf("unknown (%s)", link)
}
