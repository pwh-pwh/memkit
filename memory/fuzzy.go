package memory

import (
	"encoding/binary"
	"fmt"
	"math"
)

type FuzzyMode int

const (
	FuzzyUnchanged FuzzyMode = iota
	FuzzyIncreased
	FuzzyDecreased
)

type FuzzySnapshot[T any] struct {
	TypeSize int
	Values   map[int64]T
}

func CaptureFuzzySnapshot[T any](s *Searcher) (*FuzzySnapshot[T], error) {
	if s == nil || s.Process == nil {
		return nil, fmt.Errorf("process is nil")
	}
	size, err := numericSize(*new(T))
	if err != nil {
		return nil, err
	}

	entries, err := s.Process.FilteredMaps(s.Filter)
	if err != nil {
		return nil, err
	}

	snap := &FuzzySnapshot[T]{
		TypeSize: size,
		Values:   make(map[int64]T),
	}
	total := totalBytes(entries)
	for _, entry := range entries {
		if !entry.Readable {
			continue
		}
		for addr := entry.Start; addr < entry.End; {
			remaining := entry.End - addr
			readSize := int64(s.chunkSize(size))
			if remaining < readSize {
				readSize = remaining
			}
			if readSize <= 0 {
				break
			}

			buf := make([]byte, int(readSize))
			if err := s.Process.Read(addr, buf); err != nil {
				break
			}
			s.reportProgress(readSize, total)

			for i := 0; i+size <= len(buf); i += size {
				val, err := decodeNumber[T](buf[i : i+size])
				if err != nil {
					return nil, err
				}
				snap.Values[addr+int64(i)] = val
			}
			addr += readSize
		}
	}

	return snap, nil
}

func FilterFuzzySnapshot[T any](s *Searcher, snap *FuzzySnapshot[T], mode FuzzyMode) ([]int64, error) {
	if s == nil || s.Process == nil {
		return nil, fmt.Errorf("process is nil")
	}
	if snap == nil {
		return nil, fmt.Errorf("snapshot is nil")
	}
	if snap.TypeSize <= 0 {
		return nil, fmt.Errorf("invalid snapshot type size")
	}
	var results []int64
	buf := make([]byte, snap.TypeSize)
	for addr, oldVal := range snap.Values {
		if err := s.Process.Read(addr, buf); err != nil {
			continue
		}
		newVal, err := decodeNumber[T](buf)
		if err != nil {
			return nil, err
		}
		cmp, err := compareValues(newVal, oldVal)
		if err != nil {
			return nil, err
		}
		switch mode {
		case FuzzyUnchanged:
			if cmp == 0 {
				results = append(results, addr)
			}
		case FuzzyIncreased:
			if cmp > 0 {
				results = append(results, addr)
			}
		case FuzzyDecreased:
			if cmp < 0 {
				results = append(results, addr)
			}
		}
	}
	return results, nil
}

func decodeNumber[T any](buf []byte) (T, error) {
	var zero T
	switch any(zero).(type) {
	case int8:
		return any(int8(buf[0])).(T), nil
	case uint8:
		return any(uint8(buf[0])).(T), nil
	case int16:
		v := int16(binary.LittleEndian.Uint16(buf))
		return any(v).(T), nil
	case uint16:
		v := binary.LittleEndian.Uint16(buf)
		return any(v).(T), nil
	case int32:
		v := int32(binary.LittleEndian.Uint32(buf))
		return any(v).(T), nil
	case uint32:
		v := binary.LittleEndian.Uint32(buf)
		return any(v).(T), nil
	case int64:
		v := int64(binary.LittleEndian.Uint64(buf))
		return any(v).(T), nil
	case uint64:
		v := binary.LittleEndian.Uint64(buf)
		return any(v).(T), nil
	case float32:
		v := math.Float32frombits(binary.LittleEndian.Uint32(buf))
		return any(v).(T), nil
	case float64:
		v := math.Float64frombits(binary.LittleEndian.Uint64(buf))
		return any(v).(T), nil
	default:
		return zero, fmt.Errorf("unsupported numeric type")
	}
}
