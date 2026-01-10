
<p align="center">
<picture>
  <!-- <source media="(prefers-color-scheme: dark)" srcset="https://avatars.githubusercontent.com/u/252919145?s=200&v=4"> -->
  <img alt="Cuandari Logo featuring a medieval helmet" src="https://avatars.githubusercontent.com/u/252919145?s=200&v=4" width="200" height="200" style="border-radius: 2.5rem">
</picture>
</p>

<br/>
<div align="center">

[![contributions - welcome](https://img.shields.io/badge/contributions-welcome-blue/green)](/CONTRIBUTING.md "Go to contributions doc")
[![GitHub License](https://img.shields.io/github/license/cuandari/lib-oss.svg)](https://github.com/cuandari/lib-oss/blob/main/LICENSE)
<br/>
[![Go Report Card](https://goreportcard.com/badge/github.com/cuandari/lib-oss)](https://goreportcard.com/report/github.com/cuandari/lib-oss)
[![Go](https://img.shields.io/github/go-mod/go-version/cuandari/lib-oss
)](https://github.com/cuandari/lib-oss/blob/main/go.mod)
<br/>
[![lints](https://github.com/cuandari/lib-oss/actions/workflows/lints.yml/badge.svg)](https://github.com/cuandari/lib-oss/actions/workflows/lints.yml)
[![tests](https://github.com/cuandari/lib-oss/actions/workflows/tests.yml/badge.svg)](https://github.com/cuandari/lib-oss/actions/workflows/tests.yml)
</div>

<br/>

# cuandari/lib - Process Manager with Privilege Restrictions
Go process manager that can be used to
- start other processes and control their lifecycle,
- watch the status of the started process and return appropriate exit codes,
- and, most importantly, **trace and limit the syscalls of the started process**.

This allows you to start trusted and untrusted applications (e.g., Go, Python, Node.js) and limit their access to the file system or the network. With simple command-line flags you can easily grant permissions to the started process.

## Use Cases
- **Securely run untrusted code**: Limit what trusted and untrusted applications can do on your system.
- **Sandboxing**: Create lightweight sandboxes for applications without the overhead of full VMs or containers.
- **Testing and debugging**: Trace syscalls to understand application behavior and identify potential issues.
- **Compliance and auditing**: Enforce strict policies on application behavior for regulatory compliance

## 🤝 Examples
This section shows examples of how processes can be started with different levels of permissions and success. See below how the `curl` command fails until both filesystem and network permissions are granted.

While it's obvious why `curl` needs network permissions, filesystem permissions are necessary to read, e.g., configuration files and shared libraries.

### ❌ No filesystem permissions
In this case, `curl` is only started with a default set of permissions. The command fails because access to the filesystem is denied.
```bash
$ gatekeeper run -- curl -v google.com
[...]
Syscall not allowed: access
enter [pid 4855] access (/etc/ld.so.preload)
PID 4855 exited from signal SIGKILL (killed) (9)
Exiting with code 111
exit status 111
```

### ❌ With filesystem permissions, but no permission to access the network
In this second case, `curl` is started with a default set of permissions and **read access for the file system**. The command still fails because access to the network-related socket syscall gets denied.
```bash
$ gatekeeper run \
  --allow-file-system-read \
  -- \
  curl -v google.com
[...]
Syscall not allowed: socket
enter [pid 4996] socket
PID 4996 exited from signal SIGKILL (killed) (9)
Exiting with code 111
exit status 111
```

### ✅ With filesystem and network permissions
In this case, `curl` is started with read access to the filesystem **and** network. The command then exits with success.
```bash
$ gatekeeper run \
  --allow-file-system-read \
  --allow-network-client \
  --allow-network-local-sockets \
  -- \
  curl -v google.com
[...]
<HTML><HEAD><meta http-equiv="content-type" content="text/html;charset=utf-8">
<TITLE>301 Moved</TITLE></HEAD><BODY>
<H1>301 Moved</H1>
The document has moved
<A HREF="http://www.google.com/">here</A>.
</BODY></HTML>
[...]
PID 5255 exited from exit status 0 (code = 0)
Exiting with code 0
```

### ✅ With filesystem and network permissions
In this case, `curl` is started with read access to only specific folders **and** network. The command then exits with success.
```bash
$ gatekeeper run \
  --allow-file-system-read \
  --allow-network-client \
  --allow-network-local-sockets \
  --allow-file-system-path=/etc \
  --allow-file-system-path=/lib/x86_64-linux-gnu \
  --allow-file-system-path=/usr/lib \
  --allow-file-system-path=/usr/share \
  --allow-file-system-path=/proc/sys/crypto \
  --allow-file-system-path=/home/stfsy \
  -- \
  curl -v google.com
[...]
<HTML><HEAD><meta http-equiv="content-type" content="text/html;charset=utf-8">
<TITLE>301 Moved</TITLE></HEAD><BODY>
<H1>301 Moved</H1>
The document has moved
<A HREF="http://www.google.com/">here</A>.
</BODY></HTML>
[...]
PID 5255 exited from exit status 0 (code = 0)
Exiting with code 0
```

## 📦 Installation
Install the package:

```bash
go get https://github.com/cuandari/lib-oss
```

## 🔣 Usage
```bash
./gatekeeper [run|trace] [flags] -- [binary] [args...]
```

### 🤺 Permissions
You can pass the following flags.

- Triggers & verbosity:
  - `--trigger-enforce-on-log-match` — Enable enforcement when trace output contains this string (use with `--enforce-on-startup=false`).
  - `--trigger-enforce-on-signal` — Enable enforcement upon receiving this signal (name or number, use with `--enforce-on-startup=false`).
  - `--verbose` — Enable verbose decision logging from the tracer.

- Filesystem:
  - `--allow-file-system-read` — Allow read-only filesystem access (open O_RDONLY, read, stat, list).
  - `--allow-file-system-write` — Allow modifying the filesystem (create, write, rename, unlink, truncate).
  - `--allow-file-system` — Alias for `--allow-file-system-write` (full read/write filesystem access).
  - `--allow-file-system-permissions` — Allow changing file ownership and permissions (chmod/chown/fchmod/fchown*).
  - `--allow-file-system-path` — Allow whitelisting specific filesystem paths (repeatable); **paths should be absolute**. Example: `--allow-file-system-path=/etc` `--allow-file-system-path=/lib`. When provided, access is restricted to the listed directories (useful to grant minimal read access without enabling broad filesystem permissions).

- Network & sockets:
  - `--allow-network-client` — Allow outbound network connections (socket/connect/send/recv).
  - `--allow-network-server` — Allow listening sockets and incoming connections (socket/bind/listen/accept).
  - `--allow-network-local-sockets` — Allow local-only sockets (AF_UNIX, AF_NETLINK) for client use.
  - `--allow-networking` — Enable both client and server networking capabilities.

- Process & runtime:
  - `--allow-process-management` — Allow process/thread creation and lifecycle control (exec/fork/clone/wait).
  - `--allow-memory-management` — Allow memory mapping and related syscalls (mmap/mprotect/mremap/brk).
  - `--allow-signals` — Allow setting and handling POSIX signals (rt_sig*, sigaltstack).
  - `--allow-timers-and-clocks-management` — Allow timers and clocks (clock_gettime, timerfd_*, nanosleep).
  - `--allow-security-and-permissions` — Allow identity/capability changes and seccomp (setuid/setgid/capset/seccomp). Risky; enable only if needed.
  - `--allow-system-information` — Allow system information and rlimit operations (uname/sysinfo/getrlimit/setrlimit).
  - `--allow-process-communication` — Allow IPC mechanisms (SysV shared memory, semaphores, mqueue).
  - `--allow-process-synchronization` — Allow synchronization primitives (futex/flock/robust list).
  - `--allow-misc` — Allow miscellaneous syscalls (includes ioctl, splice, vmsplice).

- Enforcement / baseline / action:
  - `--enforce-on-startup` (default true) — Start with enforcement enabled on startup.
  - `--allow-implicit-commands` (default true) — Enable the safe baseline implicit permissions (enabled by default).
  - `--on-syscall-denied {kill|error}` — Action when a syscall is denied: `kill` (SIGKILL) or `error` (simulate EPERM via SIGSYS).



## Baseline
By default (unless you pass `--allow-implicit-commands=false`), gatekeeper enables a safe baseline including process management, memory, synchronization, signals, basic time queries and sleep (`clock_gettime`, `gettimeofday`, `nanosleep`), miscellaneous, security, and system information. This avoids breaking common applications that need time functions without requiring extra flags. Use `--allow-timers-and-clocks-management` for the full timers/clock set (e.g., `timerfd_*`, `setitimer`), or keep the default minimal set for tighter policies. If you only need to permit access to a small set of directories (for example, `/etc` or `/lib`), prefer `--allow-file-system-path` to whitelist those paths instead of granting broad filesystem read/write permissions. To explicitly disable the implicit baseline, pass `--allow-implicit-commands=false`. (Note: this flag replaces older `--no-implicit-allow`-style usage.)

#### Dynamically allow individual syscalls
In addition to grouped permissions, you can enable specific syscalls directly from the CLI without modifying configuration files. This is useful for targeted exceptions.

- `--allow-syscall-<name>`: allow a single syscall by name.
- `--allow-syscall=<name>`: equivalent form using `=`.

### 🔎 Trace
The `trace` subcommand runs the given binary and traces its syscalls. For example:

```bash
./gatekeeper trace ls -l
```

## 🧪 Running Unit Tests
To run tests, run the following command

```bash
./test.sh
```

# 🚧 Running E2E Tests
To run the end-to-end tests, run the following command

```bash
./test-e2e.sh
```
This will run all the end-to-end tests located in the `test-e2e` directory.

## 📄 License
[BSD 3-Clause](https://choosealicense.com/licenses/bsd-3-clause/)
