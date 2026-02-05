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
