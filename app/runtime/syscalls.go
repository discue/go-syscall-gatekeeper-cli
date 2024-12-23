package runtime

type SyscallAllowList struct {
	Syscalls []string
}

func NewSyscallAllowList() *SyscallAllowList {
	return &SyscallAllowList{}
}

func (sal *SyscallAllowList) AllowAllFileSystemReadAccess() {
	sal.Syscalls = append(sal.Syscalls, syscallMap["File Read Operations"]...)
}

func (sal *SyscallAllowList) AllowAllFileSystemWriteAccess() {
	sal.Syscalls = append(sal.Syscalls, syscallMap["File Write Operations"]...)
}

func (sal *SyscallAllowList) AllowAllFilePermissions() {
	sal.Syscalls = append(sal.Syscalls, syscallMap["File Permissions"]...)
}

func (sal *SyscallAllowList) AllowAllFileDescriptors() {
	sal.Syscalls = append(sal.Syscalls, syscallMap["File Descriptor Operations"]...)
}

func (sal *SyscallAllowList) AllowAllFileSystemAccess() {
	sal.AllowAllFileSystemReadAccess()
	sal.AllowAllFileSystemWriteAccess()
	sal.AllowAllFilePermissions()
	sal.AllowAllFileDescriptors()
}

// func (sal *SyscallAllowList) AllowFileRead() {
// 	sal.syscalls = append(sal.syscalls, syscallMap["File Management - Read Operations"]...)
// }

// func (sal *SyscallAllowList) AllowFileWrite() {
// 	sal.syscalls = append(sal.syscalls, syscallMap["File Management - Other Operations"]...)
// }

func (sal *SyscallAllowList) AllowProcessManagement() {
	sal.Syscalls = append(sal.Syscalls, syscallMap["Process Management"]...)
}

// func (sal *SyscallAllowList) AllowProcessCommunication() {
// 	sal.syscalls = append(sal.syscalls, syscallMap["Interprocess Communication"]...)

// }

func (sal *SyscallAllowList) AllowNetworking() {
	sal.Syscalls = append(sal.Syscalls, syscallMap["Networking Client"]...)
	sal.Syscalls = append(sal.Syscalls, syscallMap["Networking Server"]...)
}

func (sal *SyscallAllowList) AllowNetworkClient() {
	sal.Syscalls = append(sal.Syscalls, syscallMap["Networking Client"]...)
}

func (sal *SyscallAllowList) AllowNetworkServer() {
	sal.Syscalls = append(sal.Syscalls, syscallMap["Networking Server"]...)
}

func (sal *SyscallAllowList) AllowMemoryManagement() {
	sal.Syscalls = append(sal.Syscalls, syscallMap["Memory Management"]...)
}

func (sal *SyscallAllowList) AllowSignals() {
	sal.Syscalls = append(sal.Syscalls, syscallMap["Signals"]...)

}

func (sal *SyscallAllowList) AllowTimersAndClocksManagement() {
	sal.Syscalls = append(sal.Syscalls, syscallMap["Timers and Clocks"]...)
}

func (sal *SyscallAllowList) AllowSecurityAndPermissions() {
	sal.Syscalls = append(sal.Syscalls, syscallMap["Security and Permissions"]...)
}

func (sal *SyscallAllowList) AllowSystemInformation() {
	sal.Syscalls = append(sal.Syscalls, syscallMap["System Information"]...)
}

func (sal *SyscallAllowList) AllowProcessCommunication() {
	sal.Syscalls = append(sal.Syscalls, syscallMap["IPC"]...)
}

func (sal *SyscallAllowList) AllowProcessSynchronization() {
	sal.Syscalls = append(sal.Syscalls, syscallMap["Synchronization"]...)
}

func (sal *SyscallAllowList) AllowMisc() {
	sal.Syscalls = append(sal.Syscalls, syscallMap["Miscellaneous"]...)
}
