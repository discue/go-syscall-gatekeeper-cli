package runtime

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
	sec "github.com/seccomp/libseccomp-golang"
)

type EXECUTION_MODE int32

const (
	EXECUTION_MODE_TRACE EXECUTION_MODE = iota
	EXECUTION_MODE_RUN   EXECUTION_MODE = iota + 1
)

type SyscallConfig struct {
	SyscallsAllowList              []string `split_words:"true"`
	SyscallsAllowMap               map[string]bool
	SyscallsKillTargetIfNotAllowed bool `split_words:"true" default:"true"`
	SyscallsDenyTargetIfNotAllowed bool `split_words:"true" default:"false"`
}

type FsConfig struct {
	FileSystemAllowRead  bool `split_words:"true" default:"false"`
	FileSystemAllowWrite bool `split_words:"true" default:"false"`
}

type NetworkConfig struct {
	NetworkAllowClient bool `split_words:"true" default:"false"`
	NetworkAllowServer bool `split_words:"true" default:"false"`
}

type GatekeeperConfig struct {
	EnforceOnStartup bool           `split_words:"true" default:"true"`
	ExecutionMode    EXECUTION_MODE `env:"EXECUTION_MODE,enum=TRACE,RUN"`
	LogSearchString  string         `split_words:"true" default:"true"`
	VerboseLog       bool           `split_words:"true" default:"false"`
}

type Config struct {
	FsConfig
	GatekeeperConfig
	NetworkConfig
	SyscallConfig
}

var c *Config

func Load() {
	var s Config
	err := envconfig.Process("GATEKEEPER", &s)
	if err != nil {
		panic(fmt.Sprintf("unable to read environment configuration %s", err.Error()))
	}

	s.SyscallsAllowMap = CreateSyscallAllowMap(s.SyscallsAllowList)
	c = &s
}

func Get() *Config {
	if c == nil {
		Load()
	}

	return c
}

func CreateSyscallAllowMap(syscallAllowList []string) map[string]bool {
	defaultAllowDeny := len(syscallAllowList) == 0
	syscalls := make(map[string]bool)

	// Iterate from 0 to 500 (inclusive)
	for i := 0; i <= 500; i++ {
		// Get syscall name from number
		syscallName, err := sec.ScmpSyscall(int32(i)).GetName()

		// Handle errors gracefully
		if err != nil {
			continue
		}

		// Add to map with value set to true
		syscalls[syscallName] = defaultAllowDeny
	}

	for _, syscall := range syscallAllowList {
		syscalls[syscall] = !defaultAllowDeny
	}

	return syscalls
}
