package memory

import (
	"bytes"
	"fmt"
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
