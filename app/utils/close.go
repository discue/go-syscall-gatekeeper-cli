package utils

import (
	"fmt"
	"io"
)

// SafeClose closes an io.Closer and logs any error. It returns nothing.
func SafeClose(c io.Closer, label string) {
	if c == nil {
		return
	}
	if err := c.Close(); err != nil {
		fmt.Printf("Error closing %s: %s\n", label, err.Error())
	}
}
