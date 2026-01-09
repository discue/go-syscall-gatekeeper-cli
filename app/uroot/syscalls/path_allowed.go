//go:build (linux && arm64) || (linux && amd64) || (linux && riscv64)

package syscalls

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cuandari/lib/app/runtime"
	"golang.org/x/sys/unix"
)

// readPath reads a NUL-terminated string from tracee memory starting at addr.
// It reads up to maxLen bytes.
func readPath(s Syscall, addr Addr, maxLen int) (string, error) {
	var b [1]byte
	path := ""
	for i := 0; i < maxLen; i++ {
		if _, err := s.Reader(addr+Addr(i), &b); err != nil {
			return "", fmt.Errorf("unable to read path: %w", err)
		}
		if b[0] == 0 {
			break
		}
		path += string(b[:])
	}
	return path, nil
}

// PathIsAllowed checks whether the path argument (at pathArgIndex) of the
// provided Syscall falls under any of the configured allowed paths. If
// dirfdArgIndex >= 0 it is used to resolve relative paths (as in openat).
// If runtime config's FileSystemAllowedPaths is empty, PathIsAllowed returns true.
func PathIsAllowed(s Syscall, pathArgIndex int, dirfdArgIndex int) bool {
	allowed := runtime.Get().FileSystemAllowedPaths
	if len(allowed) == 0 {
		// No path-level restriction configured
		return true
	}

	pathAddr := s.Args[pathArgIndex].Pointer()
	path, err := readPath(s, pathAddr, 4096)
	if err != nil || path == "" {
		return false
	}

	// Resolve to absolute path (handle relative paths and dirfd)
	var absPath string
	if strings.HasPrefix(path, "/") {
		absPath = filepath.Clean(path)
	} else {
		// Need to resolve based on dirfd (if provided) or tracee cwd
		var base string
		if dirfdArgIndex >= 0 {
			dirfd := s.Args[dirfdArgIndex].Int()
			if dirfd == unix.AT_FDCWD {
				// tracee current working directory
				cwdPath := fmt.Sprintf("/proc/%d/cwd", s.TraceePID)
				if resolved, err := os.Readlink(cwdPath); err == nil {
					base = resolved
				} else {
					return false
				}
			} else {
				fdPath := fmt.Sprintf("/proc/%d/fd/%d", s.TraceePID, dirfd)
				if resolved, err := os.Readlink(fdPath); err == nil {
					base = resolved
				} else {
					return false
				}
			}
		} else {
			// No dirfd provided; interpret relative to tracee cwd
			cwdPath := fmt.Sprintf("/proc/%d/cwd", s.TraceePID)
			if resolved, err := os.Readlink(cwdPath); err == nil {
				base = resolved
			} else {
				return false
			}
		}

		joined := filepath.Join(base, path)
		absPath = filepath.Clean(joined)
	}

	// If the path doesn't exist yet (e.g., create), check parent directory
	checkPath := absPath
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		checkPath = filepath.Dir(absPath)
	}

	for _, a := range allowed {
		ap := filepath.Clean(a)
		// If allowed path is relative, canonicalize via Abs
		if !strings.HasPrefix(ap, "/") {
			if ap2, err := filepath.Abs(ap); err == nil {
				ap = ap2
			}
		}
		if ap == checkPath {
			return true
		}
		rel, err := filepath.Rel(ap, checkPath)
		if err == nil && rel != "" && !strings.HasPrefix(rel, "..") {
			return true
		}
	}

	// print path that is not allowed
	fmt.Printf("path %s is not allowed\n", checkPath)

	return false
}
