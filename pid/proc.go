package pid

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type ProcStat struct {
	PID   int
	Comm  string
	State byte
	PPID  int
}

func GetPIDList() ([]int, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}
	var pids []int
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}
		pids = append(pids, pid)
	}
	return pids, nil
}

func GetProcessList() ([]ProcStat, error) {
	pids, err := GetPIDList()
	if err != nil {
		return nil, err
	}
	var procs []ProcStat
	for _, pid := range pids {
		stat, err := readProcStat(pid)
		if err != nil {
			continue
		}
		procs = append(procs, stat)
	}
	return procs, nil
}

func FindProcess(keyword string) (int, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return -1, fmt.Errorf("keyword is empty")
	}
	pids, err := GetPIDList()
	if err != nil {
		return -1, err
	}
	for _, pid := range pids {
		cmdline, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "cmdline"))
		if err == nil && strings.Contains(string(cmdline), keyword) {
			return pid, nil
		}
		stat, err := readProcStat(pid)
		if err == nil && strings.Contains(stat.Comm, keyword) {
			return pid, nil
		}
	}
	return -1, fmt.Errorf("process not found")
}

func readProcStat(pid int) (ProcStat, error) {
	path := filepath.Join("/proc", strconv.Itoa(pid), "stat")
	file, err := os.Open(path)
	if err != nil {
		return ProcStat{}, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	line, err := reader.ReadString('\n')
	if err != nil && len(line) == 0 {
		return ProcStat{}, err
	}

	lparen := strings.Index(line, "(")
	rparen := strings.LastIndex(line, ")")
	if lparen == -1 || rparen == -1 || rparen <= lparen {
		return ProcStat{}, fmt.Errorf("invalid stat format")
	}

	comm := line[lparen+1 : rparen]
	rest := strings.Fields(line[rparen+1:])
	if len(rest) < 3 {
		return ProcStat{}, fmt.Errorf("invalid stat fields")
	}
	state := rest[0][0]
	ppid, err := strconv.Atoi(rest[1])
	if err != nil {
		return ProcStat{}, err
	}

	return ProcStat{PID: pid, Comm: comm, State: state, PPID: ppid}, nil
}
