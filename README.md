# 🚀 memkit

![Go](https://img.shields.io/badge/language-Go-blue.svg)
![Platform](https://img.shields.io/badge/platform-Android-green.svg)
![License](https://img.shields.io/badge/license-MIT-yellow.svg)

**English | [中文文档](./README_zh.md)**

memkit is a lightweight and efficient memory framework written in Go, designed for Android memory reading, editing, and manipulation tasks. It provides developers with an easy-to-use API for direct memory operations on Android processes.

---

## ✨ Features

- 🟦 **Pure Go Implementation**: No dependency on other languages or complex environments.
- 📱 **Android Support**: Tailored for Android (Linux `/proc` based).
- ⚡ **Multiple Read/Write Modes**: `/proc/<pid>/mem` and `process_vm_readv/writev` (with fallback).
- 🧭 **Maps Parsing + Range Tags**: Parse `/proc/<pid>/maps`, classify ranges (heap/stack/java/etc.).
- 🔎 **Search Toolkit**: Bytes/AOB search, typed number search, and range filters.
- 🚀 **Concurrent Scanning**: Worker-based scanning with progress callbacks.

---

## 🧩 Main Functions

- Read and write memory from specified processes
- Parse and filter memory maps; query module base addresses
- Search values/patterns; refine results; set operations on result sets
- Resolve pointer chains for dynamic addresses

---

## 📦 Installation

```bash
go get github.com/pwh-pwh/memkit
```

---

## 🚀 Quick Start

```go
import "github.com/pwh-pwh/memkit/memory"

func main() {
    pid := 1234
    proc := memory.NewProcess(pid)
    defer proc.Close()

    // Read a value
    v, _ := memory.ReadValFromProcess[int32](proc, 0x12345678)
    _ = v

    // Get module base
    base, _ := proc.ModuleBase("libil2cpp.so")
    _ = base

    // Search bytes / AOB
    s := memory.NewSearcher(proc)
    s.Workers = 4
    addrs, _ := s.SearchPattern("12 34 ?? 56")
    _ = addrs
}
```

---

## 🎯 Use Cases

- Android game memory editing
- App reverse engineering
- Automated testing and debugging
- System tool development

---

## 🤝 Contributing

Feel free to submit issues or pull requests to help improve memkit!

---

## 📄 License

MIT License

---

For more documentation or usage examples, please refer to the source code or open an issue.
