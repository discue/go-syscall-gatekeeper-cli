package uroot

var enforced = false
var syscallsBeforeEnforce = make(map[string]int64)
var syscallsAfterEnforce = make(map[string]int64)
var wasForceKilled = false

func enforceGatekeeper() {
	enforced = true
}

func GetIsGatekeeperEnforced() bool {
	return enforced
}

func GetTraceeWasForceKilled() bool {
	return wasForceKilled
}

func SetTraceeWasForceKilled(b bool) {
	wasForceKilled = b
}

func addSyscallToCollection(rax uint64, name string) {
	// key := fmt.Sprintf("%d->%s", rax, name)
	key := name
	if enforced {
		syscallsAfterEnforce[key] = syscallsAfterEnforce[key] + 1
	} else {
		syscallsBeforeEnforce[key] = syscallsBeforeEnforce[key] + 1
	}
}

func GetSyscallsCollectedBeforeEnforce() map[string]int64 {
	return syscallsBeforeEnforce
}

func GetSyscallsCollectedAfterEnforce() map[string]int64 {
	return syscallsAfterEnforce
}
