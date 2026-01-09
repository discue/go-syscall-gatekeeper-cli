# Changelog

## 1.0.0 (2026-01-09)


### Features

* add additional checks for write and openat syscalls ([0433912](https://github.com/cuandari/lib-oss/commit/0433912fd183bc934ccd5234245e135c6b0237db))
* add additional logic for shutdown and close syscalls ([458aae5](https://github.com/cuandari/lib-oss/commit/458aae547d2f74791932815189d65e23474354a3))
* add and use build and runtime configuration ([8ac2607](https://github.com/cuandari/lib-oss/commit/8ac2607544e5703da780905d5cf4810fdc294b9b))
* add functions that enable validation of file descriptors ([88ba881](https://github.com/cuandari/lib-oss/commit/88ba881d0286e20704c41d8f313ce633735c8648))
* add gatekeeper proxy and liveness server ([6f33829](https://github.com/cuandari/lib-oss/commit/6f338293c57493722912d442ce549da56c574795))
* add more finegrained allow levels for file system ([b1f41c2](https://github.com/cuandari/lib-oss/commit/b1f41c2ed38f58ebd1c1248e57ae074c6f20d615))
* add more finegrained syscall arg checks for write and read ([e6bf26c](https://github.com/cuandari/lib-oss/commit/e6bf26cf3c964ba66d74e62692c68d9b146f9876))
* add new cli flag to only error on syscalls instead of killing tracee ([448c3eb](https://github.com/cuandari/lib-oss/commit/448c3eb3bab0abce33edc2010295a285891fddb4))
* add new streamlined cli flags for delaying and triggering syscall checks ([1c97039](https://github.com/cuandari/lib-oss/commit/1c970394c406597d833c40fce1cff57c27ce0fa1))
* add permissions for network clients and/or servers ([6de96a2](https://github.com/cuandari/lib-oss/commit/6de96a27d1f0d79964985c6f55e585342d17ceba))
* allow configured allowed file system paths ([fad254c](https://github.com/cuandari/lib-oss/commit/fad254c59c089206ff59772b20b07549afdbea9d))
* allow more syscalls, add more groups ([7a3da54](https://github.com/cuandari/lib-oss/commit/7a3da54d36ddffe2d35e152aa0880a84e0d18b98))
* allow only local sockets if requested ([eaf1b80](https://github.com/cuandari/lib-oss/commit/eaf1b80ebbef529a9fc96f7a080634de92bbf281))
* allow passing additional syscalls ([c7e4046](https://github.com/cuandari/lib-oss/commit/c7e4046cbf64e7555ca65646df13ee780accf9cc))
* always allow write if target is a standard stream ([5079d4e](https://github.com/cuandari/lib-oss/commit/5079d4ec2a1c2269c4e1455ada39bb72c596b606))
* catch exit code tracee and return it ([b0dd36a](https://github.com/cuandari/lib-oss/commit/b0dd36a95496e81c2e5329d4ac82da79e024b285))
* collect amount of calls per syscall ([752b06c](https://github.com/cuandari/lib-oss/commit/752b06c819d8388daac47275da4fd8e53b23d379))
* exit asap if tracee cannot be started ([d70e423](https://github.com/cuandari/lib-oss/commit/d70e423861150fa2d7d473aac6c2c77e3feb5e2a))
* exit early if not enough params passed ([a55f1cc](https://github.com/cuandari/lib-oss/commit/a55f1ccab0da800fb5ff308f97e8db3ceaf1c30f))
* handle also anonymous fd types ([fad51e0](https://github.com/cuandari/lib-oss/commit/fad51e077477f3fc807e74a7442ee7ae78c4a163))
* improve error and exit handling in tracer ([98cb18e](https://github.com/cuandari/lib-oss/commit/98cb18e9d599cc70a9fce94a3c22cf146bb11c88))
* improve error handling during ptrace setup ([16009c8](https://github.com/cuandari/lib-oss/commit/16009c8aa9292d67dbcf57a17288ff53c0409569))
* improve error handling in tracer.go ([3dc56ee](https://github.com/cuandari/lib-oss/commit/3dc56ee86749f774362e2b8564c79b1a2457c73f))
* inject SIGSYS intro tracee if syscall was denied ([a42ca17](https://github.com/cuandari/lib-oss/commit/a42ca170fe1d3ccac0960986a64dc2464a1374a5))
* read fds for tracee pid instead of our own ([d8aa689](https://github.com/cuandari/lib-oss/commit/d8aa689303cdf4fee3369c2137695bbd586183d3))
* reduce default permissions ([d8c8cef](https://github.com/cuandari/lib-oss/commit/d8c8cef34bcd164f76d89eae6a356112ad948dda))
* return exit code 111 if tracee was called because of not allowed syscall ([a20036f](https://github.com/cuandari/lib-oss/commit/a20036f512b251b0869ce5bebf140a87d13e749d))
* set GATEKEEPER_PID var in tracee env ([c71fd30](https://github.com/cuandari/lib-oss/commit/c71fd307d9647b48c82065b8646063f84d455040))
* stopped traced service when main services is stopped ([a1a246b](https://github.com/cuandari/lib-oss/commit/a1a246b7b5aa11857f7bd72700aa8fcc374f5c4e))
* store whether general fs access is allowed or not ([7ca8a39](https://github.com/cuandari/lib-oss/commit/7ca8a399bf0945422d8013d8698bdd92efb4ee4d))
* **tracer:** checking whether files are opened for write access ([b6c9286](https://github.com/cuandari/lib-oss/commit/b6c928642112b7d8e6ad965abb00c306da94a1f7))
* update cli flag descriptions ([b5abfbd](https://github.com/cuandari/lib-oss/commit/b5abfbdd22420d18531230a9ca5d831862424b57))
* update syscall map and default permissions ([0964d15](https://github.com/cuandari/lib-oss/commit/0964d15c0f6207937d124eb6c94efb304f9c1333))
* **uroot:** lookup path of binary ([a483001](https://github.com/cuandari/lib-oss/commit/a48300153e7a6f80d539f636050a321f199fe9bf))


### Bug Fixes

* add missing parameter ([4696337](https://github.com/cuandari/lib-oss/commit/46963370796c5e962da62ec0467e6ec8cbd126bc))
* set implicit commands if requested ([25ea137](https://github.com/cuandari/lib-oss/commit/25ea137407a7e13c00812e5398da761bdd9cfb70))
