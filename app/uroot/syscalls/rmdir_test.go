//go:build (linux && arm64) || (linux && amd64) || (linux && riscv64)

package syscalls

import (
	"path/filepath"
	"testing"

	"github.com/cuandari/lib/app/runtime"
)

func TestRmdirAllowed(t *testing.T) {
	td := t.TempDir()
	runtime.Get().FileSystemAllowedPaths = []string{td}
	runtime.Get().FileSystemAllowWrite = true

	path := filepath.Join(td, "dir")
	var s Syscall
	base := uintptr(0x5100)
	s.Args[0] = SyscallArgument{Value: base}
	s.Reader = makeReaderFor(path, base)

	if !IsRmdirAllowed(s, true) {
		t.Fatalf("expected rmdir allowed for %s", path)
	}
}

func TestRmdirNotAllowed(t *testing.T) {
	td := t.TempDir()
	other := t.TempDir()
	runtime.Get().FileSystemAllowedPaths = []string{td}
	runtime.Get().FileSystemAllowWrite = true

	path := filepath.Join(other, "dir")
	var s Syscall
	base := uintptr(0x5101)
	s.Args[0] = SyscallArgument{Value: base}
	s.Reader = makeReaderFor(path, base)

	if IsRmdirAllowed(s, true) {
		t.Fatalf("expected rmdir NOT allowed for %s", path)
	}
}
