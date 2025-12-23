package args

import (
	"net"
	"os"
	"testing"

	iouring_syscall "github.com/iceber/iouring-go/syscall"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sys/unix"
)

// cleanup helpers ensure defer cleanups check errors without failing tests
type closer interface{ Close() error }

func cleanupRemove(t *testing.T, path string) {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		t.Logf("cleanup remove %s: %v", path, err)
	}
}

func cleanupRemoveAll(t *testing.T, path string) {
	if err := os.RemoveAll(path); err != nil {
		t.Logf("cleanup removeall %s: %v", path, err)
	}
}

func cleanupClose(t *testing.T, c closer) {
	if err := c.Close(); err != nil {
		t.Logf("cleanup close: %v", err)
	}
}

func cleanupUnixClose(t *testing.T, fd int) {
	if err := unix.Close(fd); err != nil {
		t.Logf("cleanup unix close fd %d: %v", fd, err)
	}
}

func TestIsStdIn(t *testing.T) {
	a := assert.New(t)
	a.True(IsStdIn(int32(os.Stdin.Fd())))
	a.False(IsStdIn(int32(os.Stdout.Fd())))
	a.False(IsStdIn(int32(os.Stderr.Fd())))
}

func TestIsStdOut(t *testing.T) {
	a := assert.New(t)
	a.True(IsStdOut(int32(os.Stdout.Fd())))
	a.False(IsStdOut(int32(os.Stdin.Fd())))
	a.False(IsStdOut(int32(os.Stderr.Fd())))
}

func TestIsStdErr(t *testing.T) {
	a := assert.New(t)
	a.True(IsStdErr(int32(os.Stderr.Fd())))
	a.False(IsStdErr(int32(os.Stdin.Fd())))
	a.False(IsStdErr(int32(os.Stdout.Fd())))
}

func TestIsNotAStandardStream(t *testing.T) {
	a := assert.New(t)
	a.False(IsStandardStream(100))
}

func TestStdInIsStandardStream(t *testing.T) {
	a := assert.New(t)
	a.True(IsStandardStream(int32(os.Stdin.Fd())))
}

func TestStdOutIsStandardStream(t *testing.T) {
	a := assert.New(t)
	a.True(IsStandardStream(int32(os.Stdout.Fd())))
}

func TestStdErrIsStandardStream(t *testing.T) {
	a := assert.New(t)
	a.True(IsStandardStream(int32(os.Stderr.Fd())))
}

func TestIsFile(t *testing.T) {
	a := assert.New(t)

	f, err := os.CreateTemp("", "test-file")
	a.NoError(err)

	defer cleanupRemove(t, f.Name())
	a.True(IsFile(os.Getpid(), int32(f.Fd())))
}

func TestIsDir(t *testing.T) {
	a := assert.New(t)

	dir, err := os.MkdirTemp("", "test-dir")
	a.NoError(err)
	defer cleanupRemoveAll(t, dir)

	f, _ := os.OpenFile(dir, os.O_RDONLY, 0)
	defer cleanupClose(t, f)

	a.True(IsDir(os.Getpid(), int32(f.Fd())))
}

func TestIsSymlink(t *testing.T) {
	a := assert.New(t)

	// Create a temporary directory ensuring the path has no symlinks
	tempDir, err := os.MkdirTemp("", "test-dir") // Use a consistent non-symlink path
	a.NoError(err)
	defer cleanupRemoveAll(t, tempDir)

	// Create a direct symlink
	link := tempDir + "/test-link"
	err = os.Symlink(tempDir, link) //  'dir' should be a simple direct path.
	a.NoError(err)
	defer cleanupRemove(t, link)

	fd, err := unix.Open(link, unix.O_PATH|unix.O_NOFOLLOW, 0)
	a.NoError(err)
	defer cleanupUnixClose(t, fd)
	a.True(IsSymlink(os.Getpid(), int32(fd)))
}

func TestIsBlockDevice(t *testing.T) {
	// This test requires a block device to be available.
	// We'll skip it.
	t.Skip()
}

func TestIsCharDevice(t *testing.T) {
	// This test requires a character device to be available.
	// We'll skip it.
	t.Skip()
}

func TestIsSocket(t *testing.T) {
	a := assert.New(t)

	tcpConn, err := net.ListenTCP("tcp", &net.TCPAddr{})
	a.NoError(err)
	defer cleanupClose(t, tcpConn)

	file, err := tcpConn.File() // Get the *os.File
	a.NoError(err)
	defer cleanupClose(t, file) // Important: Close the file to release resources

	a.True(IsSocket(os.Getpid(), int32(file.Fd()))) // Now you can check the FD
}

func TestIsPipe(t *testing.T) {
	a := assert.New(t)

	r, w, err := os.Pipe()
	a.NoError(err)
	defer cleanupClose(t, r)
	defer cleanupClose(t, w)

	a.True(IsPipe(os.Getpid(), int32(r.Fd())))
	a.True(IsPipe(os.Getpid(), int32(w.Fd())))
}
func TestFdTypeFile(t *testing.T) {
	a := assert.New(t)

	f, err := os.CreateTemp("", "test-file")
	a.NoError(err)
	defer cleanupRemove(t, f.Name())
	a.Equal(FDFile, FdType(os.Getpid(), int32(f.Fd())))
}

func TestFdTypeDir(t *testing.T) {
	a := assert.New(t)

	dir, err := os.MkdirTemp("", "test-dir")
	a.NoError(err)
	defer cleanupRemoveAll(t, dir)

	f, _ := os.OpenFile(dir, os.O_RDONLY, 0)
	defer cleanupClose(t, f)

	a.Equal(FDDir, FdType(os.Getpid(), int32(f.Fd())))
}

func TestFdTypeSymlink(t *testing.T) {
	a := assert.New(t)

	// Create a temporary directory ensuring the path has no symlinks
	tempDir, err := os.MkdirTemp("", "test-dir") // Use a consistent non-symlink path
	a.NoError(err)
	defer cleanupRemoveAll(t, tempDir)

	// Create a direct symlink
	link := tempDir + "/test-link"
	err = os.Symlink(tempDir, link) //  'dir' should be a simple direct path.
	a.NoError(err)
	defer cleanupRemove(t, link)

	fd, err := unix.Open(link, unix.O_PATH|unix.O_NOFOLLOW, 0)
	a.NoError(err)
	defer cleanupUnixClose(t, fd)

	a.Equal(FDSymlink, FdType(os.Getpid(), int32(fd)))
}

func TestFdTypeSocket(t *testing.T) {
	a := assert.New(t)

	tcpConn, err := net.ListenTCP("tcp", &net.TCPAddr{})
	a.NoError(err)
	defer cleanupClose(t, tcpConn)

	file, err := tcpConn.File()
	a.NoError(err)
	defer cleanupClose(t, file)

	a.Equal(FDSocket, FdType(os.Getpid(), int32(file.Fd())))
}

func TestFdTypePipe(t *testing.T) {
	a := assert.New(t)

	r, w, err := os.Pipe()
	a.NoError(err)
	defer cleanupClose(t, r)
	defer cleanupClose(t, w)

	a.Equal(FDPipe, FdType(os.Getpid(), int32(r.Fd())))
	a.Equal(FDPipe, FdType(os.Getpid(), int32(w.Fd())))
}

func TestFdTypeUnknown(t *testing.T) {
	a := assert.New(t)
	a.Equal(FDUnknown, FdType(os.Getpid(), 1000000))
}

func TestFdTypeAnonEvent(t *testing.T) {
	a := assert.New(t)

	fd, err := unix.Eventfd(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanupUnixClose(t, fd)

	fdType := FdType(os.Getpid(), int32(fd))
	a.Equal(FDAnonEvent, fdType)
}

func TestFdTypeAnonEventPoll(t *testing.T) {
	a := assert.New(t)

	fd, err := unix.EpollCreate1(0)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanupUnixClose(t, fd)

	fdType := FdType(os.Getpid(), int32(fd))
	a.Equal(FDAnonEventPoll, fdType)
}

func TestFdTypeAnonIoUring(t *testing.T) {
	a := assert.New(t)

	fd, err := iouring_syscall.IOURingSetup(1, &iouring_syscall.IOURingParams{})
	a.NoError(err)
	defer cleanupUnixClose(t, int(fd))

	fdType := FdType(os.Getpid(), int32(fd))
	a.Equal(FDAnonIoUring, fdType)
}
