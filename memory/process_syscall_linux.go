//go:build linux || android

package memory

import (
	"fmt"
	"syscall"
	"unsafe"
)

func readBySyscall(pid int, addr int64, buf []byte) error {
	if len(buf) == 0 {
		return nil
	}
	local := []syscall.Iovec{{
		Base: &buf[0],
		Len:  uint64(len(buf)),
	}}
	remote := []syscall.Iovec{{
		Base: (*byte)(unsafe.Pointer(uintptr(addr))),
		Len:  uint64(len(buf)),
	}}
	n, _, errno := syscall.Syscall6(
		syscall.SYS_PROCESS_VM_READV,
		uintptr(pid),
		uintptr(unsafe.Pointer(&local[0])),
		uintptr(len(local)),
		uintptr(unsafe.Pointer(&remote[0])),
		uintptr(len(remote)),
		0,
	)
	if errno != 0 {
		return fmt.Errorf("process_vm_readv failed: %w", errno)
	}
	if n != uintptr(len(buf)) {
		return fmt.Errorf("process_vm_readv short read %d/%d", n, len(buf))
	}
	return nil
}

func writeBySyscall(pid int, addr int64, data []byte) error {
	if len(data) == 0 {
		return nil
	}
	local := []syscall.Iovec{{
		Base: &data[0],
		Len:  uint64(len(data)),
	}}
	remote := []syscall.Iovec{{
		Base: (*byte)(unsafe.Pointer(uintptr(addr))),
		Len:  uint64(len(data)),
	}}
	n, _, errno := syscall.Syscall6(
		syscall.SYS_PROCESS_VM_WRITEV,
		uintptr(pid),
		uintptr(unsafe.Pointer(&local[0])),
		uintptr(len(local)),
		uintptr(unsafe.Pointer(&remote[0])),
		uintptr(len(remote)),
		0,
	)
	if errno != 0 {
		return fmt.Errorf("process_vm_writev failed: %w", errno)
	}
	if n != uintptr(len(data)) {
		return fmt.Errorf("process_vm_writev short write %d/%d", n, len(data))
	}
	return nil
}
