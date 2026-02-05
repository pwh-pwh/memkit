# Repository Guidelines

## Project Structure & Module Organization
- `memory/` contains the generic memory reader (`ReadVal`) and its tests.
- `pid/` contains PID lookup helpers (`PidofPidFetcher`, `PsPidFetcher`, `ProcPidFetcher`) and tests.
- `doc.md` is a small command-note file (currently Android/system commands).
- `go.mod` defines the module: `github.com/pwh-pwh/memkit` and Go version `1.24.5`.

## Build, Test, and Development Commands
- `go test ./...` runs all package tests.
- `go test ./pid -run TestPidofPidFetcher_GetPID` runs a single PID test.
- `go test ./memory -run TestForReadVal` runs the memory read test.

Note: tests currently expect a real, running process (e.g., `test_cli`) and a valid memory address. Adjust test inputs or mark tests as integration when needed.

## Coding Style & Naming Conventions
- Use standard Go formatting via `gofmt` (tabs for indentation).
- Package names are short and lowercase (`memory`, `pid`).
- Exported types/functions use `PascalCase`; unexported use `camelCase`.
- Keep error messages short and actionable (see `memory/memory.go`).

## Testing Guidelines
- Framework: Go’s built-in `testing` package.
- File naming: `*_test.go`; test naming: `TestXxx`.
- Prefer deterministic tests. For integration-style tests that require system state, document prerequisites in the test or skip when missing.

## Commit & Pull Request Guidelines
- Commit history shows short, descriptive messages, often in Chinese (e.g., “实现获取pid并增加测试方法”).
- Keep commits focused on a single change or feature.
- Pull requests should include:
  - Summary of changes and rationale.
  - How tests were run (or why not).
  - Any required environment details (e.g., Linux `/proc` access or Android tooling).

## Security & Configuration Notes
- `memory.ReadVal` reads from `/proc/<pid>/mem`; this typically requires elevated privileges and Linux/Android support.
- PID fetchers rely on system commands (`pidof`, `ps`) or `/proc` filesystem; ensure these are available in your environment.
