package pid

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type PIDFetcher interface {
	GetPID(packageName string) (int, error)
}

type PidofPidFetcher struct {
}

type PsPidFetcher struct {
}

type ProcPidFetcher struct {
}

func (p PidofPidFetcher) GetPID(packageName string) (int, error) {
	if packageName == "" {
		return -1, errors.New("package name is empty")
	}
	output, err := exec.Command("pidof", packageName).CombinedOutput()
	if err != nil {
		return -1, err
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return -1, err
	}
	return pid, nil
}

func (p PsPidFetcher) GetPID(packageName string) (int, error) {
	if packageName == "" {
		return -1, errors.New("package name is empty")
	}
	output, err := exec.Command("ps", "-a", packageName).CombinedOutput()
	if err != nil {
		return -1, err
	}
	strOut := string(output)
	if strOut == "" {
		return -1, errors.New("package not found")
	}
	for item := range strings.Lines(strOut) {
		fields := strings.Fields(item)
		if len(fields) != 4 {
			continue
		}
		pidStr := fields[0]
		name := fields[3]
		if strings.Contains(name, packageName) {
			pid, err := strconv.Atoi(pidStr)
			if err != nil {
				return -1, err
			}
			return pid, nil
		}
	}
	return -1, errors.New("package not found")
}

func (p ProcPidFetcher) GetPID(packageName string) (int, error) {
	dirEntries, err := os.ReadDir("/proc")
	if err != nil {
		return -1, err
	}
	for _, entry := range dirEntries {
		if entry.IsDir() {
			pid, err := strconv.Atoi(entry.Name())
			if err != nil {
				continue
			}
			cmdLinePath := filepath.Join("/proc", entry.Name(), "cmdline")
			cmdLine, err := os.ReadFile(cmdLinePath)
			if err != nil {
				continue
			}
			if strings.Contains(string(cmdLine), packageName) {
				return pid, nil
			}
		}
	}
	return -1, errors.New("package not found")
}
