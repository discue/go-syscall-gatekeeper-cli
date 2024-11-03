package uroot

var enforced = false
var syscallsBeforeEnforce = make(map[string]int)
var syscallsAfterEnforce = make(map[string]int)

func enforceGatekeeper() {
	enforced = true
}

func GetIsGatekeeperEnforced() bool {
	return enforced
}

func addSyscallToCollection(_rax int, name string) {
	if enforced {
		syscallsAfterEnforce[name] = syscallsAfterEnforce[name] + 1
	} else {
		syscallsBeforeEnforce[name] = syscallsBeforeEnforce[name] + 1
	}
}

func GetSyscallsCollectedBeforeEnforce() map[string]int {
	return syscallsBeforeEnforce
}

func GetSyscallsCollectedAfterEnforce() map[string]int {
	return syscallsAfterEnforce
}
