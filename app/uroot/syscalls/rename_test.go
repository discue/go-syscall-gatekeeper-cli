//go:build (linux && arm64) || (linux && amd64) || (linux && riscv64)

package syscalls

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/cuandari/lib/app/runtime"
)

func TestRenameAllowed(t *testing.T) {
	td := t.TempDir()
	runtime.Get().FileSystemAllowedPaths = []string{td}
	runtime.Get().FileSystemAllowWrite = true

	old := filepath.Join(td, "old.txt")
	newp := filepath.Join(td, "new.txt")
	var s Syscall
	b1 := uintptr(0x6000)
	b2 := uintptr(0x7000)
	s.Args[0] = SyscallArgument{Value: b1}
	s.Args[1] = SyscallArgument{Value: b2}
	s.Reader = func(addr Addr, v interface{}) (int, error) {
		if bb, ok := v.(*[1]byte); ok {
			ap := uintptr(addr)
			if ap >= b1 && ap < b1+uintptr(len(old)+1) {
				off := int(ap - b1)
				b := []byte(old + "\x00")
				bb[0] = b[off]
				return 1, nil
			}
			if ap >= b2 && ap < b2+uintptr(len(newp)+1) {
				off := int(ap - b2)
				b := []byte(newp + "\x00")
				bb[0] = b[off]
				return 1, nil
			}
			return 0, fmt.Errorf("out of range")
		}
		return 0, fmt.Errorf("unsupported v type")
	}

	if !IsRenameAllowed(s, true) {
		t.Fatalf("expected rename allowed for %s -> %s", old, newp)
	}
}

func TestRenameNotAllowed(t *testing.T) {
	td := t.TempDir()
	other := t.TempDir()
	runtime.Get().FileSystemAllowedPaths = []string{td}
	runtime.Get().FileSystemAllowWrite = true

	old := filepath.Join(td, "old.txt")
	newp := filepath.Join(other, "new.txt")
	var s Syscall
	b1 := uintptr(0x7010)
	b2 := uintptr(0x8010)
	s.Args[0] = SyscallArgument{Value: b1}
	s.Args[1] = SyscallArgument{Value: b2}
	s.Reader = func(addr Addr, v interface{}) (int, error) {
		if bb, ok := v.(*[1]byte); ok {
			ap := uintptr(addr)
			if ap >= b1 && ap < b1+uintptr(len(old)+1) {
				off := int(ap - b1)
				b := []byte(old + "\x00")
				bb[0] = b[off]
				return 1, nil
			}
			if ap >= b2 && ap < b2+uintptr(len(newp)+1) {
				off := int(ap - b2)
				b := []byte(newp + "\x00")
				bb[0] = b[off]
				return 1, nil
			}
			return 0, fmt.Errorf("out of range")
		}
		return 0, fmt.Errorf("unsupported v type")
	}

	if IsRenameAllowed(s, true) {
		t.Fatalf("expected rename NOT allowed for %s -> %s", old, newp)
	}
}
