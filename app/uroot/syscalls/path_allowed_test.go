//go:build (linux && arm64) || (linux && amd64) || (linux && riscv64)

package syscalls

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/cuandari/lib/app/runtime"
	"github.com/cuandari/lib/app/utils"
)

func makeReaderFor(path string, baseAddr uintptr) func(Addr, interface{}) (int, error) {
	b := []byte(path + "\x00")
	return func(addr Addr, v interface{}) (int, error) {
		off := int(uintptr(addr) - baseAddr)
		if off < 0 || off >= len(b) {
			return 0, fmt.Errorf("out of range read: %d", off)
		}
		// v is expected to be *[1]byte in our reader usage
		if bb, ok := v.(*[1]byte); ok {
			bb[0] = b[off]
			return 1, nil
		}
		return 0, fmt.Errorf("unsupported v type")
	}
}

func TestPathIsAllowedAbsolute(t *testing.T) {
	// create a temp dir to allow
	td := t.TempDir()
	runtime.Get().FileSystemAllowedPaths = []string{td}

	path := filepath.Join(td, "file.txt")
	var s Syscall
	base := uintptr(0x1000)
	// set path argument at arg 0
	s.Args[0] = SyscallArgument{Value: base}
	s.Reader = makeReaderFor(path, base)
	// TraceePID doesn't matter for absolute paths
	allowed := PathIsAllowed(s, 0, -1)
	if !allowed {
		t.Fatalf("expected allowed for path %s", path)
	}
}

func TestPathIsNotAllowedAbsolute(t *testing.T) {
	// allowed path is /tmp/allowed, we test with another dir
	td := t.TempDir()
	another := t.TempDir()
	runtime.Get().FileSystemAllowedPaths = []string{td}

	path := filepath.Join(another, "file.txt")
	var s Syscall
	base := uintptr(0x2000)
	s.Args[0] = SyscallArgument{Value: base}
	s.Reader = makeReaderFor(path, base)

	allowed := PathIsAllowed(s, 0, -1)
	if allowed {
		t.Fatalf("expected not allowed for path %s", path)
	}
}

func TestPathIsAllowedRelativeWithDirfd(t *testing.T) {
	// allowed to the temp dir
	td := t.TempDir()
	runtime.Get().FileSystemAllowedPaths = []string{td}

	// create a file relative path inside td
	relPath := "sub/file.txt"

	// open td and get fd
	f, err := os.Open(td)
	if err != nil {
		t.Fatal(err)
	}
	defer utils.SafeClose(f, "test-path")
	fd := int(f.Fd())

	var s Syscall
	base := uintptr(0x3000)
	// dirfd at arg 0
	s.Args[0] = SyscallArgument{Value: uintptr(fd)}
	// path at arg 1
	s.Args[1] = SyscallArgument{Value: base}
	// Reader reads the relative path
	s.Reader = makeReaderFor(relPath, base)
	// Use current process PID so /proc/<pid>/fd/<fd> is valid
	s.TraceePID = os.Getpid()

	allowed := PathIsAllowed(s, 1, 0)
	if !allowed {
		t.Fatalf("expected allowed for relative path %s with dirfd %d", relPath, fd)
	}
}

func TestPathIsAllowedNewFileUnderAllowedParent(t *testing.T) {
	// allowed root
	td := t.TempDir()
	runtime.Get().FileSystemAllowedPaths = []string{td}

	path := filepath.Join(td, "newfile-that-does-not-exist.txt")
	var s Syscall
	base := uintptr(0x4000)
	s.Args[0] = SyscallArgument{Value: base}
	s.Reader = makeReaderFor(path, base)

	if PathIsAllowed(s, 0, -1) != true {
		t.Fatalf("expected allowed for creating new file under allowed parent")
	}
}

func TestPathIsNotAllowedParentWhenOnlyChildAllowed(t *testing.T) {
	// If only a child path is allowed (e.g., /etc/os-release), the parent
	// directory (/etc) should NOT be allowed for checks such as access().
	td := t.TempDir()
	child := filepath.Join(td, "childdir")
	runtime.Get().FileSystemAllowedPaths = []string{child}

	path := td
	var s Syscall
	base := uintptr(0x4500)
	s.Args[0] = SyscallArgument{Value: base}
	s.Reader = makeReaderFor(path, base)

	if PathIsAllowed(s, 0, -1) {
		t.Fatalf("expected parent path %s NOT to be allowed when only child %s is allowed", path, child)
	}
}
