//go:build (linux && arm64) || (linux && amd64) || (linux && riscv64)

package syscalls

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cuandari/lib/app/runtime"
	"github.com/cuandari/lib/app/utils"
	"golang.org/x/sys/unix"
)

func TestFaccessAtAllowedAbsolute(t *testing.T) {
	d := t.TempDir()
	runtime.Get().FileSystemAllowedPaths = []string{d}
	runtime.Get().FileSystemAllowRead = true

	path := filepath.Join(d, "file.txt")
	var s Syscall
	base := uintptr(0x9000)
	// dirfd at arg 0 (not used for absolute path)
	s.Args[0] = SyscallArgument{Value: uintptr(0)}
	// path at arg 1
	s.Args[1] = SyscallArgument{Value: base}
	s.Reader = makeReaderFor(path, base)

	if !IsFaccessAtAllowed(s, true) {
		t.Fatalf("expected faccessat allowed for absolute path %s", path)
	}
}

func TestFaccessAtNotAllowedAbsolute(t *testing.T) {
	allowed := t.TempDir()
	runtime.Get().FileSystemAllowedPaths = []string{allowed}
	runtime.Get().FileSystemAllowRead = true

	another := t.TempDir()
	path := filepath.Join(another, "file.txt")
	var s Syscall
	base := uintptr(0xA000)
	s.Args[0] = SyscallArgument{Value: uintptr(0)}
	s.Args[1] = SyscallArgument{Value: base}
	s.Reader = makeReaderFor(path, base)

	if IsFaccessAtAllowed(s, true) {
		t.Fatalf("expected faccessat NOT allowed for absolute path %s", path)
	}
}

func TestFaccessAtAllowedWithATFDCWD(t *testing.T) {
	d := t.TempDir()
	runtime.Get().FileSystemAllowedPaths = []string{d}
	runtime.Get().FileSystemAllowRead = true

	relPath := "subdir/file.txt"
	// Use current working directory as tracee cwd via AT_FDCWD
	var s Syscall
	base := uintptr(0xB000)
	// dirfd at arg 0 is AT_FDCWD
	v := unix.AT_FDCWD
	s.Args[0] = SyscallArgument{Value: uintptr(v)}
	// path at arg 1
	s.Args[1] = SyscallArgument{Value: base}
	// Reader reads the relative path
	s.Reader = makeReaderFor(relPath, base)
	// Make the process cwd point to our temp dir by changing the test process cwd
	// Save and restore original cwd
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chdir(orig)
	}()
	if err := os.Chdir(d); err != nil {
		t.Fatal(err)
	}
	// Ensure PathIsAllowed can resolve AT_FDCWD via /proc/<pid>/cwd
	s.TraceePID = os.Getpid()

	if !IsFaccessAtAllowed(s, true) {
		t.Fatalf("expected faccessat allowed for relative path %s with AT_FDCWD", relPath)
	}
}

func TestFaccessAtDeniedWhenReadNotAllowed(t *testing.T) {
	d := t.TempDir()
	runtime.Get().FileSystemAllowedPaths = []string{d}
	runtime.Get().FileSystemAllowRead = false

	path := filepath.Join(d, "file.txt")
	var s Syscall
	base := uintptr(0xC000)
	s.Args[0] = SyscallArgument{Value: uintptr(0)}
	s.Args[1] = SyscallArgument{Value: base}
	s.Reader = makeReaderFor(path, base)

	if IsFaccessAtAllowed(s, true) {
		t.Fatalf("expected faccessat DENIED when FileSystemAllowRead is false for %s", path)
	}
}

func TestFaccessAtAllowedRelativeWithDirfdAndFdResolution(t *testing.T) {
	// This is essentially the same as the earlier test but verifies that
	// resolving /proc/<pid>/fd/<fd> works as expected.
	d := t.TempDir()
	runtime.Get().FileSystemAllowedPaths = []string{d}
	runtime.Get().FileSystemAllowRead = true

	relPath := "subdir/file.txt"
	f, err := os.Open(d)
	if err != nil {
		t.Fatal(err)
	}
	defer utils.SafeClose(f, "test-faccess-fd")
	fd := int(f.Fd())

	var s Syscall
	base := uintptr(0xD000)
	// dirfd at arg 0
	s.Args[0] = SyscallArgument{Value: uintptr(fd)}
	// path at arg 1
	s.Args[1] = SyscallArgument{Value: base}
	s.Reader = makeReaderFor(relPath, base)
	s.TraceePID = os.Getpid()

	if !IsFaccessAtAllowed(s, true) {
		t.Fatalf("expected faccessat allowed for relative path %s with dirfd %d", relPath, fd)
	}
}
