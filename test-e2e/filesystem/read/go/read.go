package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func readResolvConf() (string, error) {
	// Open the file.
	f, err := os.Open("/etc/resolv.conf")
	if err != nil {
		if os.IsNotExist(err) { // Check specifically for file not found
			return "", fmt.Errorf("error: /etc/resolv.conf not found")
		}
		return "", fmt.Errorf("error opening /etc/resolv.conf: %w", err) // Wrap the error
	}
	defer func() {
		_ = f.Close()
	}()

	// Read the file contents.
	contents, err := io.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf("error reading /etc/resolv.conf: %w", err) // Wrap the error
	}
	return string(contents), nil
}

func main() {
	contents, err := readResolvConf()
	if err != nil {
		fmt.Fprintln(os.Stderr, err) // Print errors to stderr
		os.Exit(1)
	}

	if strings.Contains(contents, "nameserver") {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}
