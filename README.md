
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
Go application that can be used to watch and limit syscalls of other processes.

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
You can pass the following flags, to allow:
- `--allow-file-system-read`: To allow the started process to read from the file system
- `--allow-file-system-write`: To allow the started process to write to the file system
- `--allow-network-client`: To allow the started process to open sockets and open connections to other servers
- `--allow-network-server`: To allow the started process to listen on ports and accept incoming connections

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
