//go:build (linux && arm64) || (linux && amd64) || (linux && riscv64)

package syscalls

import (
	"fmt"
	"os"
	"unsafe"
)

// Addr is an address for use in strace I/O
type Addr uintptr

// SyscallArgument represents a syscall argument; access via typed methods.
type SyscallArgument struct {
	Value uintptr
}

// SyscallArguments represents the set of arguments passed to a syscall.
type SyscallArguments [6]SyscallArgument

// Pointer returns the Addr representation of a pointer argument.
func (a SyscallArgument) Pointer() Addr {
	return Addr(a.Value)
}

func (a SyscallArgument) String() string {
	strPtr := (*string)(unsafe.Pointer(a.Value))
	return fmt.Sprint(strPtr)
}

// Int returns the int32 representation of a 32-bit signed integer argument.
func (a SyscallArgument) Int() int32 { return int32(a.Value) }

// Uint returns the uint32 representation of a 32-bit unsigned integer argument.
func (a SyscallArgument) Uint() uint32 { return uint32(a.Value) }

// Int64 returns the int64 representation of a 64-bit signed integer argument.
func (a SyscallArgument) Int64() int64 { return int64(a.Value) }

// Uint64 returns the uint64 representation of a 64-bit unsigned integer argument.
func (a SyscallArgument) Uint64() uint64 { return uint64(a.Value) }

// SizeT returns the uint representation of a size_t argument.
func (a SyscallArgument) SizeT() uint { return uint(a.Value) }

// ModeT returns the uint representation of a mode_t argument.
func (a SyscallArgument) ModeT() uint { return uint(uint16(a.Value)) }

// Path returns the resolved path for a file descriptor argument.
func (a SyscallArgument) Path() string {
	filePath := fmt.Sprintf("/proc/self/fd/%d", a.Int())
	realPath, err := os.Readlink(filePath)
	if err != nil {
		return "<unable to read file descriptor"
	}
	return realPath
}
