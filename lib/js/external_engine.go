package js

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// ExternalEngine uses an external command-line JS runtime (node, deno, bun).
type ExternalEngine struct {
	Command string
}

// NewExternalEngine creates a new engine that shells out to an external command.
func NewExternalEngine(command string) (*ExternalEngine, error) {
	// Check if the command exists in the system's PATH.
	if _, err := exec.LookPath(command); err != nil {
		return nil, fmt.Errorf("javascript runtime '%s' not found in PATH: %w", command, err)
	}
	return &ExternalEngine{Command: command}, nil
}

// Run executes a script by piping it to the external runtime's stdin.
func (e *ExternalEngine) Run(script string) (string, error) {
	cmd := exec.Command(e.Command)
	cmd.Stdin = strings.NewReader(script)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("external js runtime '%s' failed with exit error: %w. Stderr: %s", e.Command, err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}
