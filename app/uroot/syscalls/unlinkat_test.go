//go:build (linux && arm64) || (linux && amd64) || (linux && riscv64)

package syscalls

import (
	"path/filepath"
	"testing"

	"github.com/cuandari/lib/app/runtime"
)

func TestUnlinkAtAllowed(t *testing.T) {
	td := t.TempDir()
	runtime.Get().FileSystemAllowedPaths = []string{td}
	runtime.Get().FileSystemAllowWrite = true

	path := filepath.Join(td, "file.txt")
	var s Syscall
	base := uintptr(0x5200)
	s.Args[1] = SyscallArgument{Value: base} // pathname
	s.Args[0] = SyscallArgument{Value: uintptr(0)}
	s.Reader = makeReaderFor(path, base)

	if !IsUnlinkAtAllowed(s, true) {
		t.Fatalf("expected unlinkat allowed for %s", path)
	}
}

func TestUnlinkAtNotAllowed(t *testing.T) {
	td := t.TempDir()
	other := t.TempDir()
	runtime.Get().FileSystemAllowedPaths = []string{td}
	runtime.Get().FileSystemAllowWrite = true

	path := filepath.Join(other, "file.txt")
	var s Syscall
	base := uintptr(0x5201)
	s.Args[1] = SyscallArgument{Value: base} // pathname
	s.Args[0] = SyscallArgument{Value: uintptr(0)}
	s.Reader = makeReaderFor(path, base)

	if IsUnlinkAtAllowed(s, true) {
		t.Fatalf("expected unlinkat NOT allowed for %s", path)
	}
}
