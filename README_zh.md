# 🚀 memkit

![Go](https://img.shields.io/badge/language-Go-blue.svg)
![Platform](https://img.shields.io/badge/platform-Android-green.svg)
![License](https://img.shields.io/badge/license-MIT-yellow.svg)

memkit 是一个用 Go 语言编写的安卓内存操作框架。它旨在为开发者提供**简洁、高效**的内存读取与修改能力，适用于需要直接操作安卓进程内存的各种场景。

---

## ✨ 特性

- 🟦 **纯 Go 实现**：无需依赖其他语言或复杂环境。
- 📱 **安卓支持**：基于 Linux `/proc` 机制。
- ⚡ **多种读写模式**：`/proc/<pid>/mem` 与 `process_vm_readv/writev`（失败可回退）。
- 🧭 **maps 解析 + 区段标注**：解析 `/proc/<pid>/maps`，区分 heap/stack/java 等。
- 🔎 **搜索工具**：字节/AOB 搜索、数值类型搜索与范围过滤。
- 🚀 **并发扫描**：支持 worker 并行扫描与进度回调。
- 🧩 **高级能力**：pagemap、指针链搜索、AOB patch、模糊扫描。

---

## 🧩 主要功能

- 读取/写入指定进程内存
- 解析与筛选 maps，支持模块基址查询
- 搜索与精炼结果，支持集合操作
- 指针链解析（动态地址）
- pagemap 查询（虚拟地址 -> 物理地址）
- 扫描结果保存/加载（JSON/CSV）

---

## 📦 安装

```bash
go get github.com/pwh-pwh/memkit
```

---

## 🚀 快速开始

```go
import "github.com/pwh-pwh/memkit/memory"

func main() {
    pid := 1234
    proc := memory.NewProcess(pid)
    defer proc.Close()

    // 读取值
    v, _ := memory.ReadValFromProcess[int32](proc, 0x12345678)
    _ = v

    // 获取模块基址
    base, _ := proc.ModuleBase("libil2cpp.so")
    _ = base

    // AOB 搜索
    s := memory.NewSearcher(proc)
    s.Workers = 4
    addrs, _ := s.SearchPattern("12 34 ?? 56")
    _ = addrs
}
```

---

## 📌 版本

- `v0.5.0`：pagemap、指针链搜索、AOB patch、模糊扫描、滑动搜索、结果持久化。

---

## 🎯 用途场景

- 安卓游戏内存修改
- 应用逆向分析
- 自动化测试及调试
- 系统工具开发

---

## 🤝 贡献

欢迎提交 issue 或 pull request，共同完善 memkit！

---

## 📄 许可证

MIT License

---

如需更多文档或使用示例，请参见项目源码或提交问题。
