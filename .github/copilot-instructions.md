# AI Coding Agent Instructions for go-syscall-gatekeeper

These instructions capture project-specific architecture, workflows, and patterns so an AI agent can be productive quickly in this repo.

**Big Picture**
- **Purpose:** A CLI "gatekeeper" that starts/traces a target process and enforces syscall permissions via ptrace/seccomp.
- **Entrypoint:** [main.go](../main.go) parses flags, builds `runtime.Config`, starts the tracee via `uroot.Exec()` and coordinates shutdown.
- **Tracing & Enforcement:** [app/uroot/tracer.go](../app/uroot/tracer.go) drives a ptrace loop, classifies events, inspects syscall args, and enforces based on runtime config and allow-maps.
- **Policy:** [app/runtime/config.go](../app/runtime/config.go) defines `Config` (env-driven via `envconfig`) and builds an allow-map of syscalls; [app/runtime/syscall_map.go](../app/runtime/syscall_map.go) groups syscalls by capability (filesystem, network, process mgmt, etc.).
- **Gatekeeping:** [app/uroot/syscall-gatekeeper.go](../app/uroot/syscall-gatekeeper.go) checks `runtime.Get().SyscallsAllowMap[name]` to allow/deny.
- **u-root basis:** Many tracing components derive from u-root’s strace; see [app/uroot/README.md](../app/uroot/README.md).

**Execution Modes & Flags**
- **Modes:** `trace` vs `run` set in [configureAndParseArgs()](../main.go) and `GATEKEEPER_EXECUTION_MODE` env. `trace` enforces and logs, `run` executes without restrictions.
- **Permission flags:** Examples in [README.md](../README.md). Flags map to grouped allow-lists: `--allow-file-system-read`, `--allow-file-system-write`, `--allow-network-client`, `--allow-network-server`, plus broader categories (memory, signals, IPC, etc.).
- **Implicit defaults:** Unless `--no-implicit-allow` is set, baseline categories (process mgmt, memory, sync, signals, misc, security, system info) are enabled.
- **Denied action:** `--on-syscall-denied {kill|error}` chooses SIGKILL vs simulated error (SIGSYS / EPERM) on disallowed syscalls.
- **Delayed enforcement:** Use `--no-enforce-on-startup` with one trigger: `--trigger-enforce-on-log-match` or `--trigger-enforce-on-signal`. See e2e config examples under [test-e2e/configuration](../test-e2e/configuration).

**Runtime Config via Environment**
- **Source:** Loaded by `envconfig` in [app/runtime/config.go](../app/runtime/config.go). Prefix `GATEKEEPER_` with split_words.
- **Key vars:** `GATEKEEPER_SYSCALLS_ALLOW_LIST`, `GATEKEEPER_SYS_CALLS_KILL_TARGET_IF_NOT_ALLOWED`, `GATEKEEPER_SYS_CALLS_DENY_TARGET_IF_NOT_ALLOWED`, `GATEKEEPER_FILE_SYSTEM_ALLOW_READ`, `GATEKEEPER_FILE_SYSTEM_ALLOW_WRITE`, `GATEKEEPER_NETWORK_ALLOW_CLIENT`, `GATEKEEPER_NETWORK_ALLOW_SERVER`, `GATEKEEPER_ENFORCE_ON_STARTUP`, `GATEKEEPER_EXECUTION_MODE` (TRACE|RUN), `GATEKEEPER_TRIGGER_ENFORCE_LOG_MATCH`, `GATEKEEPER_TRIGGER_ENFORCE_SIGNAL`, `GATEKEEPER_VERBOSE_LOG`.
- **Allow-map creation:** `CreateSyscallAllowMap()` initializes a full syscall map defaulting to deny/allow based on list emptiness, then flips entries for any specified allow-list names.

**Syscall Handling Patterns**
- **FD-type heuristics:** In [app/uroot/tracer.go](../app/uroot/tracer.go), read/write syscalls consult FD type via `args.IsStandardStream`, `args.IsSocket`, `args.IsFile`, `args.IsPipe`, `args.FdType` to grant access when the corresponding category is enabled.
- **Filesystem reads vs writes:** For `openat`, read-only allowance is derived from `O_RDONLY` vs write flags when `FileSystemAllowRead` is true and `FileSystemAllowWrite` is false.
- **Per-syscall logic:** Add custom handling under [app/uroot/syscalls](../app/uroot/syscalls). Keep enforcement decisions in tracer and the allow-map consistent.

**Developer Workflows**
- **Build (Linux preferred):**
  ```bash
  go build -o gatekeeper .
  ./gatekeeper trace --allow-file-system-read --allow-network-client curl -v google.com
  ```
- **Unit tests:** `go test ./...` or `bash ./test.sh`. Linux kernel with ptrace/seccomp required for tracing logic.
- **E2E tests:** See [test-e2e/run.sh](../test-e2e/run.sh) and grouped scenarios under [test-e2e](../test-e2e) (filesystem, network, configuration, process-management). Use WSL2 or Docker on Windows.
- **Docker:** Dockerfile exists; use it to run tests on Linux kernel if host OS lacks seccomp/ptrace.

**Conventions & Tips**
- **Central source of truth:** Policy toggles live in `runtime.Config`; tracer reads from `runtime.Get()` only.
- **Grouping:** Extend capability sets in [app/runtime/syscall_map.go](../app/runtime/syscall_map.go) when adding features; wire flags in [main.go](../main.go) to those groups via `allowList` methods.
- **Logging:** Tracer prints concise decision logs (syscall name, FD type, allowance). Keep additions similarly minimal and actionable.
- **Exit codes:** Aggregated in `waitForShutdown()`; signal exits map to code 111.

**Integration Points**
- **libseccomp:** [github.com/seccomp/libseccomp-golang](https://github.com/seccomp/libseccomp-golang) used for syscall name/number mapping and policy concepts.
- **u-root:** Tracing primitives adapted from u-root’s strace implementation.
- **x/sys/unix:** For ptrace and wait operations.
