package memory

import (
	"fmt"
	"io"
	"os"
	"unsafe"
)

func ReadVal[T any](pid int, addr int64) (T, error) {
	memPath := fmt.Sprintf("/proc/%d/mem", pid)
	file, err := os.OpenFile(memPath, os.O_RDONLY, 0)
	if err != nil {
		return *new(T), fmt.Errorf("failed to open memory file: %w", err)
	}
	defer file.Close()
	tSize := unsafe.Sizeof(*new(T))
	rT := new(T)
	buf := unsafe.Slice((*byte)(unsafe.Pointer(rT)), tSize)
	n, err := file.ReadAt(buf, addr)
	if err != nil && err != io.EOF {
		return *new(T), fmt.Errorf("failed to read memory: %w", err)
	}
	if n != int(tSize) {
		return *new(T), fmt.Errorf("failed to read memory: short read %d/%d", n, tSize)
	}
	return *rT, nil
}

func WriteVal[T any](pid int, addr int64, val T) error {
	memPath := fmt.Sprintf("/proc/%d/mem", pid)
	file, err := os.OpenFile(memPath, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("failed to open memory file: %w", err)
	}
	defer file.Close()
	tSize := unsafe.Sizeof(val)
	buf := unsafe.Slice((*byte)(unsafe.Pointer(&val)), tSize)
	n, err := file.WriteAt(buf, addr)
	if err != nil {
		return fmt.Errorf("failed to write memory: %w", err)
	}
	if n != int(tSize) {
		return fmt.Errorf("failed to write memory: short write %d/%d", n, tSize)
	}
	return nil
}
