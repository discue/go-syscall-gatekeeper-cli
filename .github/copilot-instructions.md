# AI Coding Agent Instructions for go-syscall-gatekeeper

These instructions capture the essential, repository-specific knowledge an AI coding agent needs to be productive quickly.

## Big picture
- **Purpose:** A library + small runtime that starts/traces a target process and enforces syscall permissions using ptrace/seccomp primitives adapted from u-root's strace.
- **Library API:** The project exposes a public API in `package gatekeeper` (`gatekeeper.Start` / `gatekeeper.StartAndWait`) so a separate CLI project can construct a `runtime.Config` and invoke the library to start/monitor a tracee.
- **Tracing & enforcement:** The tracing loop lives in `app/uroot/tracer.go` and consults policy in `runtime.Config` and `SyscallsAllowMap`. Per-syscall helpers live under `app/uroot/syscalls/`.
- **Config:** Runtime configuration is primarily environment-driven (via `envconfig`) but can be injected programmatically via `runtime.Set(cfg)` and cleared with `runtime.Reset()` (useful for tests).

## Important files / where to look
- Runtime & policy: `app/runtime/config.go`, `app/runtime/syscall_map.go` (syscall groups)
- Tracer logic: `app/uroot/tracer.go`, `app/uroot/syscall-gatekeeper.go`
- Per-syscall helpers: `app/uroot/syscalls/*.go` (add syscall-specific gating here)
- Public API: `gatekeeper/gatekeeper.go` (Start / StartAndWait)
- Process exec & strace wrapper: `app/uroot/uroot.go` (Exec, Trace, Strace)
- E2E scenarios & runner: `test-e2e/` and `test-e2e/run.sh`

## Key concepts & conventions
- **Central source of truth:** `runtime.Get()` reads the active configuration. Use `runtime.Set(cfg)` to inject a programmatic config for calls originating from the external CLI or for tests.
- **Allow-map semantics:** `CreateSyscallAllowMap()` iterates syscall numbers and builds a name->bool map. If the allow-list is empty, defaults are permissive; provided names flip entries accordingly.
- **FD heuristics:** Tracer uses fd-type helpers (in `app/uroot/syscall-argument.go` and `app/uroot/syscalls/*`) to decide permission for socket/file/pipe access — prefer adding checks in the syscall helper rather than directly modifying tracer logic.
- **Cancellation & exit model:** The code uses `context.WithCancelCause` and surface-specific causes as `uroot.ExitEventError`. `uroot.Exec` returns an exit context you can wait on.
- **Platform constraints:** The tracing code is linux-only (see build tags in `app/uroot/uroot.go`). Many tests and E2E scenarios require a Linux kernel with ptrace/seccomp; use Docker or WSL2 on non-Linux hosts.

## Developer workflows
- Build (Linux):
  - `go build ./...` (or `go build -o gatekeeper .` inside a module that provides a main)
  - The repo no longer contains a user-facing CLI; the external CLI project should import `github.com/cuandari/lib/gatekeeper` and call `Start`/`StartAndWait`.
- Unit tests: `go test ./...` or use `bash ./test.sh`. Some tests require ptrace/seccomp and will fail on non-Linux environments.
- E2E tests: `test-e2e/run.sh` — the runner expects a gatekeeper binary present; when working locally on non-Linux, run the tests inside Docker using the provided Dockerfile or use WSL2.

## How to extend the project (common tasks)
- Add a syscall-specific gate: create `app/uroot/syscalls/<name>.go` with helper functions and call it from `tracer.go` switch-case (see existing patterns for `socket`, `connect`, `openat`, `read`, `write`).
- Add a new capability group: extend `app/runtime/syscall_map.go` and add wiring in the external CLI to set flags that flip these groups via `runtime.Config`.
- Add config values: prefer adding typed fields to `runtime.Config` and ensure `CreateSyscallAllowMap` memberships remain consistent.

## Integration points
- seccomp name/number mapping: `github.com/seccomp/libseccomp-golang` (used when building syscall map)
- OS primitives: `golang.org/x/sys/unix` for ptrace/wait
- Upstream inspiration: u-root's `strace` implementation (see `app/uroot/README.md`)

## Quick example (how an external CLI should call this repo)
```go
cfg := &runtime.Config{ /* set flags or load via env */ }
runtime.Set(cfg)
code, err := gatekeeper.StartAndWait(context.Background(), cfg, "curl", []string{"-v","https://google.com"})
if err != nil { /* handle start error */ }
fmt.Printf("exit code: %d\n", code)
```