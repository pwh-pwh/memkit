package memory

import (
	"encoding/binary"
	"fmt"
	"unsafe"
)

type PointerChain struct {
	Base    int64
	Offsets []int64
}

type PointerSearchOptions struct {
	MaxDepth   int
	MaxOffset  int64
	MaxResults int
	Align      int
	Ranges     []MapEntry
}

func FindPointerChains(p *Process, target int64, opts PointerSearchOptions) ([]PointerChain, error) {
	if p == nil {
		return nil, fmt.Errorf("process is nil")
	}
	if opts.MaxDepth <= 0 {
		return nil, fmt.Errorf("max depth must be > 0")
	}
	if opts.MaxOffset <= 0 {
		return nil, fmt.Errorf("max offset must be > 0")
	}
	ptrSize := int(unsafe.Sizeof(uintptr(0)))
	align := opts.Align
	if align <= 0 {
		align = ptrSize
	}

	ranges := opts.Ranges
	if len(ranges) == 0 {
		entries, err := p.Maps()
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			if entry.Readable {
				ranges = append(ranges, entry)
			}
		}
	}

	type chainState struct {
		target int64
		chain  PointerChain
	}

	var results []PointerChain
	current := []chainState{{target: target}}

	for depth := 1; depth <= opts.MaxDepth; depth++ {
		if len(current) == 0 {
			break
		}
		var next []chainState
		seen := make(map[int64]struct{})
		for _, state := range current {
			hits, err := scanPointersForTarget(p, ranges, state.target, opts.MaxOffset, align, ptrSize, opts.MaxResults, &results)
			if err != nil {
				return nil, err
			}
			for _, hit := range hits {
				newOffsets := append([]int64(nil), state.chain.Offsets...)
				newOffsets = append([]int64{hit.Offset}, newOffsets...)
				chain := PointerChain{
					Base:    hit.Addr,
					Offsets: newOffsets,
				}
				results = append(results, chain)
				if opts.MaxResults > 0 && len(results) >= opts.MaxResults {
					return results, nil
				}
				if _, ok := seen[hit.Addr]; ok {
					continue
				}
				seen[hit.Addr] = struct{}{}
				next = append(next, chainState{
					target: hit.Addr,
					chain:  chain,
				})
			}
		}
		current = next
	}

	return results, nil
}

type pointerHit struct {
	Addr   int64
	Offset int64
}

func scanPointersForTarget(p *Process, ranges []MapEntry, target int64, maxOffset int64, align, ptrSize int, maxResults int, results *[]PointerChain) ([]pointerHit, error) {
	var hits []pointerHit
	const chunk = 1 << 20
	buf := make([]byte, chunk)

	minVal := target - maxOffset
	maxVal := target

	for _, entry := range ranges {
		if !entry.Readable {
			continue
		}
		for addr := entry.Start; addr < entry.End; {
			remaining := entry.End - addr
			readSize := int64(chunk)
			if remaining < readSize {
				readSize = remaining
			}
			if readSize <= 0 {
				break
			}

			if err := p.Read(addr, buf[:readSize]); err != nil {
				break
			}

			for i := 0; i+ptrSize <= int(readSize); i += align {
				val := decodePtr(buf[i:i+ptrSize], ptrSize)
				v := int64(val)
				if v < minVal || v > maxVal {
					continue
				}
				hits = append(hits, pointerHit{
					Addr:   addr + int64(i),
					Offset: target - v,
				})
				if maxResults > 0 && len(hits)+len(*results) >= maxResults {
					return hits, nil
				}
			}

			addr += readSize
		}
	}

	return hits, nil
}

func decodePtr(buf []byte, size int) uint64 {
	switch size {
	case 4:
		return uint64(binary.LittleEndian.Uint32(buf))
	default:
		return binary.LittleEndian.Uint64(buf)
	}
}
