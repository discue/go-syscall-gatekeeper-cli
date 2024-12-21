package args

import (
	"os"

	"golang.org/x/sys/unix"
)

const (
	// from https://github.com/golang/go/blob/669d87a935536eb14cb2db311a83345359189924/src/archive/tar/common.go#L627
	ISDIR  = 040000  // Directory
	ISFIFO = 010000  // FIFO
	ISREG  = 0100000 // Regular file
	ISLNK  = 0120000 // Symbolic link
	ISBLK  = 060000  // Block special file
	ISCHR  = 020000  // Character special file
	ISSOCK = 0140000 // Socket
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

func statFd(fd int32) (unix.Stat_t, error) {
	var stat unix.Stat_t
	err := unix.Fstat(int(fd), &stat) // Fstat takes int, not int32
	return stat, err
}

func isFdType(fd int32, fdConstant uint32) bool {
	stat, err := statFd(fd)
	if err != nil {
		return false
	}
	return stat.Mode&fdConstant == fdConstant
}

func IsFile(fd int32) bool {
	return isFdType(fd, ISREG)
}

func IsDir(fd int32) bool {
	return isFdType(fd, ISDIR)
}

func IsSymlink(fd int32) bool {
	return isFdType(fd, ISLNK)
}

func IsBlockDevice(fd int32) bool {
	return isFdType(fd, ISBLK)
}

func IsCharDevice(fd int32) bool {
	return isFdType(fd, ISCHR)
}

func IsSocket(fd int32) bool {
	return isFdType(fd, ISSOCK)
}

func IsPipe(fd int32) bool {
	return isFdType(fd, ISFIFO)
}
