package runtime

type SyscallAllowList struct {
	Syscalls []string
}

func NewSyscallAllowList() *SyscallAllowList {
	return &SyscallAllowList{}
}

func (sal *SyscallAllowList) AllowAllFileSystemReadAccess() {
	sal.Syscalls = append(sal.Syscalls, syscallMap["File Read Operations"]...)
	// Opening files is required before read operations; gated by tracer flags for O_RDONLY/O_WRONLY
	sal.Syscalls = append(sal.Syscalls, syscallMap["File Open"]...)
	sal.Syscalls = append(sal.Syscalls, syscallMap["Basic File Descriptor Operations"]...)
}

func (sal *SyscallAllowList) AllowAllFileSystemWriteAccess() {
	// Grant raw IO write operations (write*, pwrite*, fallocate, copy_file_range, sync_file_range)
	sal.Syscalls = append(sal.Syscalls, syscallMap["File Write Operations"]...)
	// Also grant file creation and metadata-changing operations (rename*, link*, mkdir*, symlink*, umask, unlink*, utime*)
	sal.Syscalls = append(sal.Syscalls, syscallMap["File Create/Metadata"]...)
	// Opening files is required before write operations; gated by tracer flags for O_RDONLY/O_WRONLY
	sal.Syscalls = append(sal.Syscalls, syscallMap["File Open"]...)
	sal.Syscalls = append(sal.Syscalls, syscallMap["Basic File Descriptor Operations"]...)
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
	sal.AllowNetworkClient()
	sal.AllowNetworkServer()
}

func (sal *SyscallAllowList) AllowNetworkClient() {
	sal.Syscalls = append(sal.Syscalls, syscallMap["Basic File Descriptor Operations"]...)
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
