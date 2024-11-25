//go:build windows
// +build windows

package uroot

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
)

func Exec(c context.Context, bin string, args []string) (*exec.Cmd, error) {
	cmd := exec.CommandContext(c, bin, args...)

	// setup goroutines to read and print stdout
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("error creating stdout pipe: %w", err)
	}

	go func() {
		scanner := bufio.NewScanner(stdoutPipe)
		// after we broke the first loop we create another without the if statement
		for scanner.Scan() {
			logger.Info(fmt.Sprintf(">> %s", scanner.Text())) // Print to parent's stdout
		}
	}()

	// setup goroutines to read and print errout
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("error creating stderr pipe: %w", err)
	}

	go func() {
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			logger.Error(scanner.Text()) // Print to parent's stderr
		}
	}()

	err = cmd.Start()
	if err != nil {
		logger.Error(fmt.Sprintf("error starting %s: %s", bin, err.Error()))
		return nil, err
	}

	logger.Info(fmt.Sprintf("%s started. Enabling gatekeeper now", bin))
	enforceGatekeeper()
	logger.Info(fmt.Sprintf("Gatekeeper enabled %t", GetIsGatekeeperEnforced()))
	return cmd, nil
}
