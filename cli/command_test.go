package cli_test

import (
	"testing"

	"bytes"

	"github.com/cuandari/lib/cli"
	"github.com/stretchr/testify/assert"
)

func TestNewCommandDefaults(t *testing.T) {
	a := assert.New(t)
	c := cli.NewCommand()
	a.NotNil(c)
	a.NotNil(c.FlagSet())

	// Default values before parsing
	a.False(*c.Verbose)
	a.False(*c.AllowFileSystemReadAccess)
	a.False(*c.AllowNetworkClient)
	a.False(*c.NoImplicitAllow)
	a.Equal("", *c.TriggerEnforceOnLogMatch)
	a.Equal("", *c.TriggerEnforceOnSignal)
}

func TestParseKnownFlags(t *testing.T) {
	a := assert.New(t)
	c := cli.NewCommand()

	// Parse a selection of flags and trailing args
	err := c.Parse([]string{
		"--verbose",
		"--allow-file-system-read",
		"--allow-timers-and-clocks-management",
		"--allow-network-client",
		"--trigger-enforce-on-log-match", "needle",
		"--trigger-enforce-on-signal", "SIGUSR1",
		"binary", "arg1",
	})
	a.NoError(err)

	a.True(*c.Verbose)
	a.True(*c.AllowFileSystemReadAccess)
	a.True(*c.AllowTimersAndClocksManagement)
	a.True(*c.AllowNetworkClient)
	a.Equal("needle", *c.TriggerEnforceOnLogMatch)
	a.Equal("SIGUSR1", *c.TriggerEnforceOnSignal)

	args := c.Args()
	a.Equal([]string{"binary", "arg1"}, args)
}

func TestOnSyscallDeniedFlag(t *testing.T) {
	a := assert.New(t)
	c := cli.NewCommand()

	err := c.Parse([]string{"--on-syscall-denied", "error"})
	a.NoError(err)
	a.Equal(cli.ErrorAction, c.Action)

	c = cli.NewCommand()
	err = c.Parse([]string{"--on-syscall-denied", "kill"})
	a.NoError(err)
	a.Equal(cli.KillAction, c.Action)
}

func TestSyscallDeniedActionSet(t *testing.T) {
	a := assert.New(t)
	var s cli.SyscallDeniedAction

	a.NoError(s.Set("kill"))
	a.Equal(cli.KillAction, s)

	a.NoError(s.Set("error"))
	a.Equal(cli.ErrorAction, s)

	a.Error(s.Set("invalid"))
}

func TestAllExpectedFlagsPresent(t *testing.T) {
	a := assert.New(t)
	c := cli.NewCommand()

	expected := []string{
		"trigger-enforce-on-log-match",
		"trigger-enforce-on-signal",
		"verbose",
		"allow-file-system-read",
		"allow-file-system-write",
		"allow-file-system",
		"allow-file-system-permissions",
		"allow-network-client",
		"allow-network-server",
		"allow-local-sockets",
		"allow-process-management",
		"allow-networking",
		"allow-memory-management",
		"allow-signals",
		"allow-timers-and-clocks-management",
		"allow-security-and-permissions",
		"allow-system-information",
		"allow-process-communication",
		"allow-process-synchronization",
		"allow-misc",
		"no-enforce-on-startup",
		"no-implicit-allow",
		"on-syscall-denied",
	}

	for _, name := range expected {
		a.NotNil(c.FlagSet().Lookup(name), "missing flag: %s", name)
	}
}

func TestDefaultsRemainFalseUntilParsed(t *testing.T) {
	a := assert.New(t)
	c := cli.NewCommand()

	a.False(*c.AllowFileSystemReadAccess)
	a.False(*c.AllowFileSystemWriteAccess)
	a.False(*c.AllowNetworkClient)
	a.False(*c.AllowNetworkServer)
	a.False(*c.AllowLocalSockets)
	a.False(*c.AllowProcessManagement)
	a.False(*c.AllowNetworking)
	a.False(*c.AllowMemoryManagement)
	a.False(*c.AllowSignals)
	a.False(*c.AllowTimersAndClocksManagement)
	a.False(*c.AllowSecurityAndPermissions)
	a.False(*c.AllowSystemInformation)
	a.False(*c.AllowProcessCommunication)
	a.False(*c.AllowProcessSynchronization)
	a.False(*c.AllowMisc)
	a.False(*c.NoEnforceOnStart)
	a.False(*c.NoImplicitAllow)
}

func TestPrintDefaultsContainsKeyFlags(t *testing.T) {
	a := assert.New(t)
	c := cli.NewCommand()
	var buf bytes.Buffer
	c.FlagSet().SetOutput(&buf)
	c.FlagSet().PrintDefaults()
	out := buf.String()
	a.Contains(out, "allow-file-system-read")
	a.Contains(out, "allow-network-client")
	a.Contains(out, "on-syscall-denied")
}
