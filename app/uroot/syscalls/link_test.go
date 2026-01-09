//go:build (linux && arm64) || (linux && amd64) || (linux && riscv64)

package syscalls

import (
	"path/filepath"
	"testing"

	"github.com/cuandari/lib/app/runtime"
)

func TestLinkNotAllowed(t *testing.T) {
	td := t.TempDir()
	other := t.TempDir()
	runtime.Get().FileSystemAllowedPaths = []string{td}
	runtime.Get().FileSystemAllowWrite = true

	old := filepath.Join(td, "old")
	newp := filepath.Join(other, "new")
	var s Syscall
	b1 := uintptr(0x7030)
	b2 := uintptr(0x8038)
	s.Args[0] = SyscallArgument{Value: b1}
	s.Args[1] = SyscallArgument{Value: b2}
	s.Reader = func(addr Addr, v interface{}) (int, error) {
		if bb, ok := v.(*[1]byte); ok {
			b := []byte(old + "\x00" + newp + "\x00")
			off := int(uintptr(addr) - b1)
			if off < 0 {
				off = int(uintptr(addr) - b2)
			}
			if off < 0 || off >= len(b) {
				return 0, nil
			}
			bb[0] = b[off]
			return 1, nil
		}
		return 0, nil
	}

	if IsLinkAllowed(s, true) {
		t.Fatalf("expected link NOT allowed for %s -> %s", old, newp)
	}
}
