package memory

import (
	"fmt"
	"io"
	"os"
	"unsafe"
)

type MemoryMode int

const (
	ModeMemFile MemoryMode = iota
	ModeSyscall
	ModeDirect
)

type Process struct {
	PID  int
	mode MemoryMode
	mem  *os.File
}

func NewProcess(pid int) *Process {
	return &Process{
		PID:  pid,
		mode: ModeMemFile,
	}
}

func (p *Process) SetMode(mode MemoryMode) error {
	p.mode = mode
	if mode == ModeMemFile {
		return p.openMem()
	}
	return nil
}

func (p *Process) Close() error {
	if p.mem == nil {
		return nil
	}
	err := p.mem.Close()
	p.mem = nil
	return err
}

func (p *Process) Read(addr int64, buf []byte) error {
	switch p.mode {
	case ModeMemFile:
		if err := p.openMem(); err != nil {
			return err
		}
		n, err := p.mem.ReadAt(buf, addr)
		if err != nil && err != io.EOF {
			return fmt.Errorf("failed to read memory: %w", err)
		}
		if n != len(buf) {
			return fmt.Errorf("failed to read memory: short read %d/%d", n, len(buf))
		}
		return nil
	case ModeSyscall:
		return readBySyscall(p.PID, addr, buf)
	case ModeDirect:
		copy(buf, unsafe.Slice((*byte)(unsafe.Pointer(uintptr(addr))), len(buf)))
		return nil
	default:
		return fmt.Errorf("unknown memory mode")
	}
}

func (p *Process) Write(addr int64, data []byte) error {
	switch p.mode {
	case ModeMemFile:
		if err := p.openMem(); err != nil {
			return err
		}
		n, err := p.mem.WriteAt(data, addr)
		if err != nil {
			return fmt.Errorf("failed to write memory: %w", err)
		}
		if n != len(data) {
			return fmt.Errorf("failed to write memory: short write %d/%d", n, len(data))
		}
		return nil
	case ModeSyscall:
		return writeBySyscall(p.PID, addr, data)
	case ModeDirect:
		copy(unsafe.Slice((*byte)(unsafe.Pointer(uintptr(addr))), len(data)), data)
		return nil
	default:
		return fmt.Errorf("unknown memory mode")
	}
}

func (p *Process) Maps() ([]MapEntry, error) {
	return ParseMaps(p.PID)
}

func (p *Process) FilteredMaps(filter MapFilter) ([]MapEntry, error) {
	entries, err := p.Maps()
	if err != nil {
		return nil, err
	}
	return FilterMaps(entries, filter), nil
}

func (p *Process) ModuleBase(name string) (int64, error) {
	return GetModuleBase(p.PID, name)
}

func (p *Process) openMem() error {
	if p.mem != nil {
		return nil
	}
	path := fmt.Sprintf("/proc/%d/mem", p.PID)
	file, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("failed to open memory file: %w", err)
	}
	p.mem = file
	return nil
}

func ReadValFromProcess[T any](p *Process, addr int64) (T, error) {
	size := unsafe.Sizeof(*new(T))
	value := new(T)
	buf := unsafe.Slice((*byte)(unsafe.Pointer(value)), size)
	if err := p.Read(addr, buf); err != nil {
		return *new(T), err
	}
	return *value, nil
}

func WriteValToProcess[T any](p *Process, addr int64, val T) error {
	size := unsafe.Sizeof(val)
	buf := unsafe.Slice((*byte)(unsafe.Pointer(&val)), size)
	return p.Write(addr, buf)
}
