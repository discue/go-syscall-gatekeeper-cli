package uroot

import "fmt"

// TraceError is returned when something failed on a specific process.
type TraceError struct {
	// PID is the process ID associated with the error.
	PID int
	Err error
}

func (t *TraceError) Error() string {
	return fmt.Sprintf("trace error on pid %d: %v", t.PID, t.Err)
}
