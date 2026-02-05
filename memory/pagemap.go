package memory

import (
	"encoding/binary"
	"fmt"
	"os"
)

type PageMapEntry struct {
	PFN       uint64
	SoftDirty bool
	FilePage  bool
	Swapped   bool
	Present   bool
}

func ReadPageMapEntry(p *Process, vaddr int64) (PageMapEntry, error) {
	if p == nil {
		return PageMapEntry{}, fmt.Errorf("process is nil")
	}
	path := fmt.Sprintf("/proc/%d/pagemap", p.PID)
	file, err := os.Open(path)
	if err != nil {
		return PageMapEntry{}, err
	}
	defer file.Close()

	pageSize := int64(os.Getpagesize())
	index := (uint64(vaddr) / uint64(pageSize)) * 8
	buf := make([]byte, 8)
	if _, err := file.ReadAt(buf, int64(index)); err != nil {
		return PageMapEntry{}, err
	}

	entry := binary.LittleEndian.Uint64(buf)
	return PageMapEntry{
		PFN:       entry & ((1 << 55) - 1),
		SoftDirty: (entry>>55)&1 == 1,
		FilePage:  (entry>>61)&1 == 1,
		Swapped:   (entry>>62)&1 == 1,
		Present:   (entry>>63)&1 == 1,
	}, nil
}

func VirtToPhys(p *Process, vaddr int64) (int64, error) {
	entry, err := ReadPageMapEntry(p, vaddr)
	if err != nil {
		return 0, err
	}
	if !entry.Present {
		return 0, fmt.Errorf("page not present")
	}
	pageSize := int64(os.Getpagesize())
	phys := int64(entry.PFN)*pageSize + (vaddr % pageSize)
	return phys, nil
}
