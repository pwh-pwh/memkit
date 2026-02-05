package memory

import (
	"encoding/binary"
	"fmt"
	"unsafe"
)

func ResolvePointerChain(p *Process, base int64, offsets []int64) (int64, error) {
	if p == nil {
		return 0, fmt.Errorf("process is nil")
	}
	if len(offsets) == 0 {
		return base, nil
	}

	addr := base
	ptrSize := int(unsafe.Sizeof(uintptr(0)))
	for _, off := range offsets {
		ptr, err := readPointerValue(p, addr, ptrSize)
		if err != nil {
			return 0, err
		}
		addr = int64(ptr) + off
	}
	return addr, nil
}

func readPointerValue(p *Process, addr int64, size int) (uint64, error) {
	buf := make([]byte, size)
	if err := p.Read(addr, buf); err != nil {
		return 0, err
	}
	switch size {
	case 4:
		return uint64(binary.LittleEndian.Uint32(buf)), nil
	case 8:
		return binary.LittleEndian.Uint64(buf), nil
	default:
		return 0, fmt.Errorf("unsupported pointer size: %d", size)
	}
}
