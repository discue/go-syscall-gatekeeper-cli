package args

import (
	"fmt"
	"os"
	"syscall"

	"golang.org/x/sys/unix"
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
		fmt.Println(fmt.Sprintf("error stating fd %d: %v", fd, err))
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
	if IsFile(pid, fd) {
		return "file"
	} else if IsSymlink(pid, fd) {
		return "symlink"
	} else if IsBlockDevice(pid, fd) {
		return "block device"
	} else if IsCharDevice(pid, fd) {
		return "char device"
	} else if IsSocket(pid, fd) {
		return "socket"
	} else if IsPipe(pid, fd) {
		return "pipe"
	} else if IsStandardStream(fd) {
		return "standard stream"
	} else if IsDir(pid, fd) {
		return "dir"
	}
	return "unknown"
}
