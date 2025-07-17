package memory

import (
	"fmt"
	"os"
	"unsafe"
)

func ReadVal[T any](pid int, addr int64) (T, error) {
	memPath := fmt.Sprintf("/proc/%d/mem", pid)
	file, err := os.OpenFile(memPath, os.O_RDONLY, 0)
	if err != nil {
		return *new(T), fmt.Errorf("failed to open memory file: %w", err)
	}
	file.Seek(addr, 0)
	tSize := unsafe.Sizeof(*new(T))
	rT := new(T)
	_, err = file.Read(unsafe.Slice((*byte)(unsafe.Pointer(rT)), tSize))
	if err != nil {
		return *new(T), fmt.Errorf("failed to read memory: %w", err)
	}
	return *rT, nil
}
