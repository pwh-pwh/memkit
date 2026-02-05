package memory

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

func DumpMaps(pid int, w io.Writer) error {
	entries, err := ParseMaps(pid)
	if err != nil {
		return err
	}
	bw := bufio.NewWriter(w)
	for _, entry := range entries {
		_, err := fmt.Fprintf(bw, "%#x-%#x %s %s %d %s\n",
			entry.Start,
			entry.End,
			entry.Perms,
			entry.Path,
			entry.Inode,
			entry.Range.String(),
		)
		if err != nil {
			return err
		}
	}
	return bw.Flush()
}

func DumpMemory(p *Process, start, end int64, w io.Writer) error {
	if p == nil {
		return fmt.Errorf("process is nil")
	}
	if start >= end {
		return fmt.Errorf("invalid range")
	}
	size := end - start
	const chunk = 1 << 20
	buf := make([]byte, chunk)
	for offset := int64(0); offset < size; {
		remain := size - offset
		n := int64(chunk)
		if remain < n {
			n = remain
		}
		if err := p.Read(start+offset, buf[:n]); err != nil {
			return err
		}
		if _, err := w.Write(buf[:n]); err != nil {
			return err
		}
		offset += n
	}
	return nil
}

func DumpMemoryToFile(p *Process, start, end int64, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return DumpMemory(p, start, end, file)
}
