package client

import (
	"fmt"
	"strings"
)

// Response represents the result of executing a command on the server.
type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Success  bool
}

// String returns a formatted string representation of the response.
func (r *Response) String() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Exit Code: %d\n", r.ExitCode))
	sb.WriteString(fmt.Sprintf("Success: %t\n", r.Success))

	if r.Stdout != "" {
		sb.WriteString("\n--- STDOUT ---\n")
		sb.WriteString(r.Stdout)
	}

	if r.Stderr != "" {
		sb.WriteString("\n--- STDERR ---\n")
		sb.WriteString(r.Stderr)
	}

	return sb.String()
}

// Error returns an error if the response indicates failure.
func (r *Response) Error() error {
	if r.Success {
		return nil
	}

	return fmt.Errorf("command failed with exit code %d: %s", r.ExitCode, r.Stderr)
}
