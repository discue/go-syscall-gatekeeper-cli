//go:build (linux && arm64) || (linux && amd64) || (linux && riscv64)

package syscalls

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cuandari/lib/app/runtime"
	"github.com/cuandari/lib/app/utils"
)

func TestIsAccessAllowedAbsolute(t *testing.T) {
	d := t.TempDir()
	runtime.Get().FileSystemAllowedPaths = []string{d}
	runtime.Get().FileSystemAllowRead = true

	path := filepath.Join(d, "file.txt")
	var s Syscall
	base := uintptr(0x5000)
	s.Args[0] = SyscallArgument{Value: base}
	s.Reader = makeReaderFor(path, base)

	if !IsAccessAllowed(s, true) {
		t.Fatalf("expected access allowed for %s", path)
	}
}

func TestIsAccessNotAllowedAbsolute(t *testing.T) {
	allowed := t.TempDir()
	runtime.Get().FileSystemAllowedPaths = []string{allowed}
	runtime.Get().FileSystemAllowRead = true

	another := t.TempDir()
	path := filepath.Join(another, "file.txt")
	var s Syscall
	base := uintptr(0x6000)
	s.Args[0] = SyscallArgument{Value: base}
	s.Reader = makeReaderFor(path, base)

	if IsAccessAllowed(s, true) {
		t.Fatalf("expected access NOT allowed for %s", path)
	}
}

func TestIsFaccessAtAllowedRelativeWithDirfd(t *testing.T) {
	d := t.TempDir()
	runtime.Get().FileSystemAllowedPaths = []string{d}
	runtime.Get().FileSystemAllowRead = true

	relPath := "subdir/file.txt"
	f, err := os.Open(d)
	if err != nil {
		t.Fatal(err)
	}
	defer utils.SafeClose(f, "test-faccess")
	fd := int(f.Fd())

	var s Syscall
	base := uintptr(0x7000)
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

func TestIsAccessDeniedWhenReadNotAllowed(t *testing.T) {
	d := t.TempDir()
	runtime.Get().FileSystemAllowedPaths = []string{d}
	runtime.Get().FileSystemAllowRead = false

	path := filepath.Join(d, "file.txt")
	var s Syscall
	base := uintptr(0x8000)
	s.Args[0] = SyscallArgument{Value: base}
	s.Reader = makeReaderFor(path, base)

	if IsAccessAllowed(s, true) {
		t.Fatalf("expected access DENIED when FileSystemAllowRead is false for %s", path)
	}
}
