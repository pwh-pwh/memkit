package memory

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"unsafe"
)

const defaultChunkSize = 1 << 20

type Searcher struct {
	Process   *Process
	Filter    MapFilter
	ChunkSize int
}

func NewSearcher(p *Process) *Searcher {
	return &Searcher{
		Process: p,
		Filter: MapFilter{
			RequireReadable: true,
		},
		ChunkSize: defaultChunkSize,
	}
}

func (s *Searcher) SearchBytes(target []byte) ([]int64, error) {
	if s.Process == nil {
		return nil, fmt.Errorf("process is nil")
	}
	if len(target) == 0 {
		return nil, fmt.Errorf("target is empty")
	}
	chunkSize := s.ChunkSize
	if chunkSize <= 0 {
		chunkSize = defaultChunkSize
	}

	entries, err := s.Process.FilteredMaps(s.Filter)
	if err != nil {
		return nil, err
	}

	var results []int64
	overlap := len(target) - 1

	for _, entry := range entries {
		if !entry.Readable {
			continue
		}

		var prev []byte
		for addr := entry.Start; addr < entry.End; {
			remaining := entry.End - addr
			readSize := int64(chunkSize)
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

			searchBuf := buf
			baseAddr := addr
			if len(prev) > 0 {
				combined := make([]byte, len(prev)+len(buf))
				copy(combined, prev)
				copy(combined[len(prev):], buf)
				searchBuf = combined
				baseAddr = addr - int64(len(prev))
			}

			for i := 0; ; {
				idx := bytes.Index(searchBuf[i:], target)
				if idx < 0 {
					break
				}
				pos := i + idx
				matchAddr := baseAddr + int64(pos)
				results = append(results, matchAddr)
				i = pos + 1
			}

			if overlap > 0 {
				if len(buf) >= overlap {
					prev = append(prev[:0], buf[len(buf)-overlap:]...)
				} else {
					prev = append(prev[:0], buf...)
				}
			}

			addr += readSize
		}
	}

	return results, nil
}

func SearchValue[T any](s *Searcher, val T) ([]int64, error) {
	size := unsafe.Sizeof(val)
	buf := unsafe.Slice((*byte)(unsafe.Pointer(&val)), size)
	return s.SearchBytes(buf)
}

type CompareOp int

const (
	OpEQ CompareOp = iota
	OpNE
	OpGT
	OpGE
	OpLT
	OpLE
)

func SearchNumber[T any](s *Searcher, target T, op CompareOp) ([]int64, error) {
	return searchNumberWithOp(s, target, op, nil, nil)
}

func SearchNumberRange[T any](s *Searcher, min, max T, includeMin, includeMax bool) ([]int64, error) {
	return searchNumberWithOp(s, min, rangeMinOp(&includeMin), &max, &includeMax)
}

func searchNumberWithOp[T any](s *Searcher, target T, op CompareOp, max *T, includeMax *bool) ([]int64, error) {
	if s == nil || s.Process == nil {
		return nil, fmt.Errorf("process is nil")
	}

	typeSize, err := numericSize(target)
	if err != nil {
		return nil, err
	}
	if max != nil {
		maxSize, err := numericSize(*max)
		if err != nil {
			return nil, err
		}
		if maxSize != typeSize {
			return nil, fmt.Errorf("range type mismatch")
		}
		cmp, err := compareValues(target, *max)
		if err != nil {
			return nil, err
		}
		if cmp > 0 {
			return nil, fmt.Errorf("min greater than max")
		}
	}

	entries, err := s.Process.FilteredMaps(s.Filter)
	if err != nil {
		return nil, err
	}

	var results []int64
	for _, entry := range entries {
		if !entry.Readable {
			continue
		}
		for addr := entry.Start; addr < entry.End; {
			remaining := entry.End - addr
			readSize := int64(s.chunkSize(typeSize))
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

			for i := 0; i+typeSize <= len(buf); i += typeSize {
				match, err := matchNumber(buf[i:i+typeSize], target, op)
				if err != nil {
					return nil, err
				}
				if match && max != nil {
					// Range check: value <= max (or < max)
					rmatch, err := matchNumber(buf[i:i+typeSize], *max, rangeMaxOp(includeMax))
					if err != nil {
						return nil, err
					}
					if !rmatch {
						continue
					}
				}
				if match {
					results = append(results, addr+int64(i))
				}
			}

			addr += readSize
		}
	}
	return results, nil
}

func (s *Searcher) chunkSize(typeSize int) int {
	if s.ChunkSize <= 0 {
		return defaultChunkSize
	}
	if s.ChunkSize < typeSize {
		return typeSize
	}
	return s.ChunkSize
}

func numericSize[T any](val T) (int, error) {
	switch any(val).(type) {
	case int8, uint8:
		return 1, nil
	case int16, uint16:
		return 2, nil
	case int32, uint32, float32:
		return 4, nil
	case int64, uint64, float64:
		return 8, nil
	default:
		return 0, fmt.Errorf("unsupported numeric type")
	}
}

func rangeMinOp(include *bool) CompareOp {
	if include == nil {
		return OpGE
	}
	if *include {
		return OpGE
	}
	return OpGT
}

func rangeMaxOp(include *bool) CompareOp {
	if include == nil {
		return OpLE
	}
	if *include {
		return OpLE
	}
	return OpLT
}

func matchNumber[T any](buf []byte, target T, op CompareOp) (bool, error) {
	switch t := any(target).(type) {
	case int8:
		return compareInt64(int64(int8(buf[0])), int64(t), op), nil
	case uint8:
		return compareUint64(uint64(buf[0]), uint64(t), op), nil
	case int16:
		v := int16(binary.LittleEndian.Uint16(buf))
		return compareInt64(int64(v), int64(t), op), nil
	case uint16:
		v := binary.LittleEndian.Uint16(buf)
		return compareUint64(uint64(v), uint64(t), op), nil
	case int32:
		v := int32(binary.LittleEndian.Uint32(buf))
		return compareInt64(int64(v), int64(t), op), nil
	case uint32:
		v := binary.LittleEndian.Uint32(buf)
		return compareUint64(uint64(v), uint64(t), op), nil
	case int64:
		v := int64(binary.LittleEndian.Uint64(buf))
		return compareInt64(v, t, op), nil
	case uint64:
		v := binary.LittleEndian.Uint64(buf)
		return compareUint64(v, t, op), nil
	case float32:
		v := math.Float32frombits(binary.LittleEndian.Uint32(buf))
		return compareFloat64(float64(v), float64(t), op), nil
	case float64:
		v := math.Float64frombits(binary.LittleEndian.Uint64(buf))
		return compareFloat64(v, t, op), nil
	default:
		return false, fmt.Errorf("unsupported numeric type")
	}
}

func compareInt64(a, b int64, op CompareOp) bool {
	switch op {
	case OpEQ:
		return a == b
	case OpNE:
		return a != b
	case OpGT:
		return a > b
	case OpGE:
		return a >= b
	case OpLT:
		return a < b
	case OpLE:
		return a <= b
	default:
		return false
	}
}

func compareUint64(a, b uint64, op CompareOp) bool {
	switch op {
	case OpEQ:
		return a == b
	case OpNE:
		return a != b
	case OpGT:
		return a > b
	case OpGE:
		return a >= b
	case OpLT:
		return a < b
	case OpLE:
		return a <= b
	default:
		return false
	}
}

func compareFloat64(a, b float64, op CompareOp) bool {
	switch op {
	case OpEQ:
		return a == b
	case OpNE:
		return a != b
	case OpGT:
		return a > b
	case OpGE:
		return a >= b
	case OpLT:
		return a < b
	case OpLE:
		return a <= b
	default:
		return false
	}
}

func compareValues[T any](a, b T) (int, error) {
	switch av := any(a).(type) {
	case int8:
		bv := any(b).(int8)
		return compareOrderInt64(int64(av), int64(bv)), nil
	case uint8:
		bv := any(b).(uint8)
		return compareOrderUint64(uint64(av), uint64(bv)), nil
	case int16:
		bv := any(b).(int16)
		return compareOrderInt64(int64(av), int64(bv)), nil
	case uint16:
		bv := any(b).(uint16)
		return compareOrderUint64(uint64(av), uint64(bv)), nil
	case int32:
		bv := any(b).(int32)
		return compareOrderInt64(int64(av), int64(bv)), nil
	case uint32:
		bv := any(b).(uint32)
		return compareOrderUint64(uint64(av), uint64(bv)), nil
	case int64:
		bv := any(b).(int64)
		return compareOrderInt64(av, bv), nil
	case uint64:
		bv := any(b).(uint64)
		return compareOrderUint64(av, bv), nil
	case float32:
		bv := any(b).(float32)
		if math.IsNaN(float64(av)) || math.IsNaN(float64(bv)) {
			return 0, fmt.Errorf("nan not supported")
		}
		return compareOrderFloat64(float64(av), float64(bv)), nil
	case float64:
		bv := any(b).(float64)
		if math.IsNaN(av) || math.IsNaN(bv) {
			return 0, fmt.Errorf("nan not supported")
		}
		return compareOrderFloat64(av, bv), nil
	default:
		return 0, fmt.Errorf("unsupported numeric type")
	}
}

func compareOrderInt64(a, b int64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

func compareOrderUint64(a, b uint64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

func compareOrderFloat64(a, b float64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

func RefineAddresses(addrs []int64, fn func(int64) (bool, error)) ([]int64, error) {
	if len(addrs) == 0 {
		return nil, nil
	}
	if fn == nil {
		return nil, fmt.Errorf("refine function is nil")
	}
	var out []int64
	for _, addr := range addrs {
		ok, err := fn(addr)
		if err != nil {
			return nil, err
		}
		if ok {
			out = append(out, addr)
		}
	}
	return out, nil
}

func RefineNumber[T any](s *Searcher, addrs []int64, target T, op CompareOp) ([]int64, error) {
	if s == nil || s.Process == nil {
		return nil, fmt.Errorf("process is nil")
	}
	size, err := numericSize(target)
	if err != nil {
		return nil, err
	}
	return RefineAddresses(addrs, func(addr int64) (bool, error) {
		buf := make([]byte, size)
		if err := s.Process.Read(addr, buf); err != nil {
			return false, nil
		}
		return matchNumber(buf, target, op)
	})
}

func RefineNumberRange[T any](s *Searcher, addrs []int64, min, max T, includeMin, includeMax bool) ([]int64, error) {
	if s == nil || s.Process == nil {
		return nil, fmt.Errorf("process is nil")
	}
	minSize, err := numericSize(min)
	if err != nil {
		return nil, err
	}
	maxSize, err := numericSize(max)
	if err != nil {
		return nil, err
	}
	if minSize != maxSize {
		return nil, fmt.Errorf("range type mismatch")
	}
	cmp, err := compareValues(min, max)
	if err != nil {
		return nil, err
	}
	if cmp > 0 {
		return nil, fmt.Errorf("min greater than max")
	}

	return RefineAddresses(addrs, func(addr int64) (bool, error) {
		buf := make([]byte, minSize)
		if err := s.Process.Read(addr, buf); err != nil {
			return false, nil
		}
		okMin, err := matchNumber(buf, min, rangeMinOp(&includeMin))
		if err != nil || !okMin {
			return okMin, err
		}
		return matchNumber(buf, max, rangeMaxOp(&includeMax))
	})
}

func RefineBytes(s *Searcher, addrs []int64, target []byte) ([]int64, error) {
	if s == nil || s.Process == nil {
		return nil, fmt.Errorf("process is nil")
	}
	if len(target) == 0 {
		return nil, fmt.Errorf("target is empty")
	}
	return RefineAddresses(addrs, func(addr int64) (bool, error) {
		buf := make([]byte, len(target))
		if err := s.Process.Read(addr, buf); err != nil {
			return false, nil
		}
		return bytes.Equal(buf, target), nil
	})
}

func AddressesUnion(a, b []int64) []int64 {
	if len(a) == 0 {
		return append([]int64(nil), b...)
	}
	if len(b) == 0 {
		return append([]int64(nil), a...)
	}
	seen := make(map[int64]struct{}, len(a)+len(b))
	var out []int64
	for _, addr := range a {
		if _, ok := seen[addr]; ok {
			continue
		}
		seen[addr] = struct{}{}
		out = append(out, addr)
	}
	for _, addr := range b {
		if _, ok := seen[addr]; ok {
			continue
		}
		seen[addr] = struct{}{}
		out = append(out, addr)
	}
	return out
}

func AddressesIntersect(a, b []int64) []int64 {
	if len(a) == 0 || len(b) == 0 {
		return nil
	}
	setB := make(map[int64]struct{}, len(b))
	for _, addr := range b {
		setB[addr] = struct{}{}
	}
	var out []int64
	seen := make(map[int64]struct{}, len(a))
	for _, addr := range a {
		if _, ok := setB[addr]; !ok {
			continue
		}
		if _, ok := seen[addr]; ok {
			continue
		}
		seen[addr] = struct{}{}
		out = append(out, addr)
	}
	return out
}

func AddressesDiff(a, b []int64) []int64 {
	if len(a) == 0 {
		return nil
	}
	if len(b) == 0 {
		return append([]int64(nil), a...)
	}
	setB := make(map[int64]struct{}, len(b))
	for _, addr := range b {
		setB[addr] = struct{}{}
	}
	var out []int64
	for _, addr := range a {
		if _, ok := setB[addr]; ok {
			continue
		}
		out = append(out, addr)
	}
	return out
}
