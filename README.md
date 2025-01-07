
<p align="center"><a href="https://www.discue.io/" target="_blank" rel="noopener noreferrer"><img width="256" src="https://www.discue.io/icons-fire-no-badge-square/web/icon-192.png" alt="Vue logo"></a></p>

<br/>
<div align="center">

[![contributions - welcome](https://img.shields.io/badge/contributions-welcome-blue/green)](/CONTRIBUTING.md "Go to contributions doc")
[![GitHub License](https://img.shields.io/github/license/discue/go-syscall-gatekeeper.svg)](https://github.com/discue/go-syscall-gatekeeper-cli/blob/master/LICENSE)
<br/>
[![Go Report Card](https://goreportcard.com/badge/github.com/discue/go-syscall-gatekeeper)](https://goreportcard.com/report/github.com/discue/go-syscall-gatekeeper)
[![Go](https://img.shields.io/github/go-mod/go-version/discue/go-syscall-gatekeeper-cli
)](https://github.com/discue/go-syscall-gatekeeper-cli/blob/main/go.mod)
<br/>
[![lints](https://github.com/discue/go-syscall-gatekeeper-cli/actions/workflows/lints.yml/badge.svg)](https://github.com/discue/go-syscall-gatekeeper-cli/actions/workflows/lints.yml)
[![tests](https://github.com/discue/go-syscall-gatekeeper-cli/actions/workflows/tests.yml/badge.svg)](https://github.com/discue/go-syscall-gatekeeper-cli/actions/workflows/tests.yml)
</div>

<br/>

# go-syscall-gatekeeper
Go process manager that can be used to 
- start other processes and control their lifecycle,
- watch the status of the started process and return appropriate exit codes,
- and, most importantly, **trace and limit the syscalls of the started process**. 

This allows you to start trusted and untrusted applications e.g. go, python, node apps and limit their access to the file system, or to the network. With simple command line flags you can easily grant permissions to the started process.

## 🤝 Examples
This section shows some examples of how processes can be started with different level of permissions and... success. See below, how the `curl` command is failing until both filesystem and network permissions are granted.

While it's obvious, why `curl` needs network permissions, the filesystem permissions are necessary to read e.g. configuration files and shared libraries.

### ❌ No filesystem permissions
In this case, `curl` is only started with a default set of permissions. The command fails because, access to the filesystem gets denied.
```bash
$ gatekeeper run curl -v google.com
[...]
Syscall not allowed: access
enter [pid 4855] access (/etc/ld.so.preload)
PID 4855 exited from signal SIGKILL (killed) (9)
Exiting with code 111
exit status 111
```

### ❌ With filesystem permissions, but no permission to access network
In this second case, `curl` is started with a default set of permissions and **read access for the file system**. The command still fails because access to the network-related socket syscall gets denied.
```bash
$ gatekeeper run --allow-file-system-read curl -v google.com
[...]
Syscall not allowed: socket
enter [pid 4996] socket
PID 4996 exited from signal SIGKILL (killed) (9)
Exiting with code 111
exit status 111
```

### ✅ With filesystem and network permissions
In this final case, `curl` is started with read access to the filesystem **and** network. The command then exits with success.
```bash
$ gatekeeper run --allow-file-system-read --allow-network-client curl -v google.com
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
go get https://github.com/discue/go-syscall-gatekeeper
```

## 🔣 Usage
```bash
./gatekeeper [run|trace] [binary] [args...]
```
### 🚀 Run
The `run` subcommand runs the given command without any syscall restrictions. This is as good as calling the target program directly.

```bash
./gatekeeper run ls -l
```

### 🤺 Permissions
You can pass the following flags:
- `--allow-file-system-read` to allow the started process to read from the file system,
- `--allow-file-system-write` to allow the started process to write to the file system,
- `--allow-network-client` to allow the started process to open sockets and open connections to other servers,
- `--allow-network-server` to allow the started process to listen on ports and accept incoming connections.

### 🔎 Trace
The `trace` subcommand run the given binary and traces the syscalls. In this case, the `gatekeeper` will 

```bash
./gatekeeper trace ls -l
```

## 🧪 Running Tests
To run tests, run the following command

```bash
./test.sh
```

## 📄 License
[BSD 3-Clause](https://choosealicense.com/licenses/bsd-3-clause/)
