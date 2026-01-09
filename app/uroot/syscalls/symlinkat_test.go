//go:build (linux && arm64) || (linux && amd64) || (linux && riscv64)

package syscalls

import (
	"path/filepath"
	"testing"

	"github.com/cuandari/lib/app/runtime"
)

func TestSymlinkAtAllowed(t *testing.T) {
	td := t.TempDir()
	runtime.Get().FileSystemAllowedPaths = []string{td}
	runtime.Get().FileSystemAllowWrite = true

	linkpath := filepath.Join(td, "lnk")
	var s Syscall
	b1 := uintptr(0x7120) // target
	b2 := uintptr(0x8128) // linkpath
	// symlinkat(target, newdirfd, linkpath)
	s.Args[2] = SyscallArgument{Value: b2}
	s.Args[1] = SyscallArgument{Value: uintptr(0)}
	s.Args[0] = SyscallArgument{Value: b1}
	s.Reader = func(addr Addr, v interface{}) (int, error) {
		if bb, ok := v.(*[1]byte); ok {
			ap := uintptr(addr)
			target := "/tmp/some"
			if ap >= b1 && ap < b1+uintptr(len(target)+1) {
				off := int(ap - b1)
				b := []byte(target + "\x00")
				bb[0] = b[off]
				return 1, nil
			}
			if ap >= b2 && ap < b2+uintptr(len(linkpath)+1) {
				off := int(ap - b2)
				b := []byte(linkpath + "\x00")
				bb[0] = b[off]
				return 1, nil
			}
			return 0, nil
		}
		return 0, nil
	}

	if !IsSymlinkAtAllowed(s, true) {
		t.Fatalf("expected symlinkat allowed for %s", linkpath)
	}
}
