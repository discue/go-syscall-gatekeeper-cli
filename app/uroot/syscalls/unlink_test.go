//go:build (linux && arm64) || (linux && amd64) || (linux && riscv64)

package syscalls

import (
	"path/filepath"
	"testing"

	"github.com/cuandari/lib/app/runtime"
)

func TestUnlinkAllowed(t *testing.T) {
	td := t.TempDir()
	runtime.Get().FileSystemAllowedPaths = []string{td}
	runtime.Get().FileSystemAllowWrite = true

	path := filepath.Join(td, "file.txt")
	var s Syscall
	base := uintptr(0x5000)
	s.Args[0] = SyscallArgument{Value: base}
	s.Reader = makeReaderFor(path, base)

	if !IsUnlinkAllowed(s, true) {
		t.Fatalf("expected unlink allowed for %s", path)
	}
}

func TestUnlinkNotAllowed(t *testing.T) {
	td := t.TempDir()
	other := t.TempDir()
	runtime.Get().FileSystemAllowedPaths = []string{td}
	runtime.Get().FileSystemAllowWrite = true

	path := filepath.Join(other, "file.txt")
	var s Syscall
	base := uintptr(0x5001)
	s.Args[0] = SyscallArgument{Value: base}
	s.Reader = makeReaderFor(path, base)

	if IsUnlinkAllowed(s, true) {
		t.Fatalf("expected unlink NOT allowed for %s", path)
	}
}
