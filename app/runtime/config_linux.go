package runtime

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
	sec "github.com/seccomp/libseccomp-golang"
)

type SyscallConfig struct {
	SyscallsAllowList              []string `split_words:"true"`
	SyscallsAllowMap               map[string]bool
	SyscallsDelayEnforceUntilCheck bool `split_words:"true" default:"true"`
	SyscallsKillTargetIfNotAllowed bool `split_words:"true" default:"true"`
}

type GatekeeperConfig struct {
	LivenessCheckHttpEnabled              bool   `split_words:"true" default:"false"`
	LivenessCheckHttpProbeIntervalSeconds int    `split_words:"true" default:"1"`
	LivenessCheckLogEnabled               bool   `split_words:"true" default:"true"`
	LivenessCheckLogSearchString          string `split_words:"true" default:"Server running at"`
}

type GatekeeperServerConfig struct {
	ServerEnabled bool `split_words:"true" default:"true"`
	ServerPort    int  `split_words:"true" default:"8081"`
}

type LivenessProxyConfig struct {
	LivenessTargetProxyEnabled bool   `split_words:"true" default:"false"`
	LivenessTargetProtocol     string `split_words:"true" default:"http"`
	LivenessTargetHostname     string `split_words:"true" default:"127.0.0.1"`
	LivenessTargetPort         int    `split_words:"true" default:"8080"`
	LivenessTargetPath         string `split_words:"true" default:"/live"`
}

type HealthProxyConfig struct {
	HealthTargetProxyEnabled bool   `split_words:"true" default:"false"`
	HealthTargetProtocol     string `split_words:"true" default:"http"`
	HealthTargetHostname     string `split_words:"true" default:"127.0.0.1"`
	HealthTargetPort         int    `split_words:"true" default:"8080"`
	HealthTargetPath         string `split_words:"true" default:"/health"`
}

type Config struct {
	GatekeeperConfig
	SyscallConfig
	HealthProxyConfig
	LivenessProxyConfig
	GatekeeperServerConfig
}

var c *Config

func Load() {
	var s Config
	err := envconfig.Process("GATEKEEPER", &s)
	if err != nil {
		panic(fmt.Sprintf("unable to read environment configuration %s", err.Error()))
	}

	s.SyscallsAllowMap = createSyscallMap(s.SyscallsAllowList)
	c = &s
}

func Get() Config {
	if c == nil {
		Load()
	}

	return *c
}

func reset() {
	c = nil
}

func createSyscallMap(syscallAllowList []string) map[string]bool {
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
