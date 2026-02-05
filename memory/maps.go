package memory

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type MapEntry struct {
	Start      int64
	End        int64
	Perms      string
	Readable   bool
	Writable   bool
	Executable bool
	Private    bool
	Offset     int64
	Device     string
	Inode      uint64
	Path       string
	Range      RangeType
}

type RangeType int

const (
	RangeUnknown RangeType = iota
	RangeStack
	RangeHeap
	RangeJava
	RangeAnon
	RangeAshmem
	RangeDevice
	RangeCode
	RangeFile
	RangeVdso
	RangeVvar
)

func (r RangeType) String() string {
	switch r {
	case RangeStack:
		return "stack"
	case RangeHeap:
		return "heap"
	case RangeJava:
		return "java"
	case RangeAnon:
		return "anon"
	case RangeAshmem:
		return "ashmem"
	case RangeDevice:
		return "device"
	case RangeCode:
		return "code"
	case RangeFile:
		return "file"
	case RangeVdso:
		return "vdso"
	case RangeVvar:
		return "vvar"
	default:
		return "unknown"
	}
}

type MapFilter struct {
	RequireReadable   bool
	RequireWritable   bool
	RequireExecutable bool
	RequirePrivate    bool
	PathContains      string
	MinStart          int64
	MaxEnd            int64
}

func ParseMaps(pid int) ([]MapEntry, error) {
	path := fmt.Sprintf("/proc/%d/maps", pid)
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open maps: %w", err)
	}
	defer file.Close()

	var entries []MapEntry
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 5 {
			return nil, fmt.Errorf("invalid maps line: %q", line)
		}

		start, end, err := parseAddressRange(fields[0])
		if err != nil {
			return nil, fmt.Errorf("invalid address range %q: %w", fields[0], err)
		}

		perms := fields[1]
		offset, err := parseHexInt64(fields[2])
		if err != nil {
			return nil, fmt.Errorf("invalid offset %q: %w", fields[2], err)
		}

		inode, err := strconv.ParseUint(fields[4], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid inode %q: %w", fields[4], err)
		}

		pathname := ""
		if len(fields) > 5 {
			pathname = strings.Join(fields[5:], " ")
		}

		entry := MapEntry{
			Start:      start,
			End:        end,
			Perms:      perms,
			Readable:   len(perms) > 0 && perms[0] == 'r',
			Writable:   len(perms) > 1 && perms[1] == 'w',
			Executable: len(perms) > 2 && perms[2] == 'x',
			Private:    len(perms) > 3 && perms[3] == 'p',
			Offset:     offset,
			Device:     fields[3],
			Inode:      inode,
			Path:       pathname,
		}
		entry.Range = DetermineRange(entry)
		entries = append(entries, entry)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read maps: %w", err)
	}
	return entries, nil
}

func GetModuleBase(pid int, name string) (int64, error) {
	if name == "" {
		return -1, fmt.Errorf("module name is empty")
	}
	maps, err := ParseMaps(pid)
	if err != nil {
		return -1, err
	}
	for _, entry := range maps {
		if strings.Contains(entry.Path, name) {
			return entry.Start, nil
		}
	}
	return -1, fmt.Errorf("module not found: %s", name)
}

func DetermineRange(entry MapEntry) RangeType {
	path := entry.Path
	switch {
	case path == "[stack]":
		return RangeStack
	case path == "[heap]":
		return RangeHeap
	case path == "[vdso]":
		return RangeVdso
	case path == "[vvar]":
		return RangeVvar
	}

	if strings.Contains(path, "dalvik") || strings.Contains(path, "art") || strings.Contains(path, "jit-cache") || strings.Contains(path, "zygote") {
		return RangeJava
	}
	if strings.Contains(path, "/dev/ashmem") {
		return RangeAshmem
	}
	if strings.HasPrefix(path, "/dev/") {
		return RangeDevice
	}
	if path == "" {
		return RangeAnon
	}
	if entry.Executable {
		return RangeCode
	}
	return RangeFile
}

func FilterMaps(entries []MapEntry, filter MapFilter) []MapEntry {
	if len(entries) == 0 {
		return nil
	}
	var out []MapEntry
	for _, entry := range entries {
		if filter.RequireReadable && !entry.Readable {
			continue
		}
		if filter.RequireWritable && !entry.Writable {
			continue
		}
		if filter.RequireExecutable && !entry.Executable {
			continue
		}
		if filter.RequirePrivate && !entry.Private {
			continue
		}
		if filter.PathContains != "" && !strings.Contains(entry.Path, filter.PathContains) {
			continue
		}
		if filter.MinStart != 0 && entry.Start < filter.MinStart {
			continue
		}
		if filter.MaxEnd != 0 && entry.End > filter.MaxEnd {
			continue
		}
		out = append(out, entry)
	}
	return out
}

func parseAddressRange(addr string) (int64, int64, error) {
	parts := strings.Split(addr, "-")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid address range")
	}
	start, err := parseHexInt64(parts[0])
	if err != nil {
		return 0, 0, err
	}
	end, err := parseHexInt64(parts[1])
	if err != nil {
		return 0, 0, err
	}
	return start, end, nil
}

func parseHexInt64(val string) (int64, error) {
	return strconv.ParseInt(val, 16, 64)
}
