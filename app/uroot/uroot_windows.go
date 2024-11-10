//go:build windows
// +build windows

package uroot

import (
	"bufio"
	"fmt"
	"os/exec"
)

func Exec(bin string, args []string) error {
	cmd := exec.Command(bin, args...)

	// setup goroutines to read and print stdout
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error creating stdout pipe: %w", err)
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
		return fmt.Errorf("error creating stderr pipe: %w", err)
	}
	go func() {
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			logger.Error(scanner.Text()) // Print to parent's stderr
		}
	}()

	cmd.Start()
	enforceGatekeeper()

	return cmd.Wait()
}
