//go:build !linux && !android

package memory

import "fmt"

func readBySyscall(pid int, addr int64, buf []byte) error {
	return fmt.Errorf("syscall mode not supported on this platform")
}

func writeBySyscall(pid int, addr int64, data []byte) error {
	return fmt.Errorf("syscall mode not supported on this platform")
}
