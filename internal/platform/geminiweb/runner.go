package geminiweb

import (
	"context"
	"os/exec"
)

// CommandRunner is an interface for executing commands.
// It enables testing by allowing mock implementations.
type CommandRunner interface {
	// LookPath searches for an executable named file in the directories
	// named by the PATH environment variable.
	LookPath(file string) (string, error)

	// CommandContext returns a Cmd struct to execute the named program with
	// the given arguments. The context can be used to cancel the command.
	CommandContext(ctx context.Context, name string, args ...string) *exec.Cmd
}

// DefaultCommandRunner is the production implementation that uses os/exec.
type DefaultCommandRunner struct{}

// LookPath implements CommandRunner by calling exec.LookPath.
func (r *DefaultCommandRunner) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

// CommandContext implements CommandRunner by calling exec.CommandContext.
func (r *DefaultCommandRunner) CommandContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, name, args...)
}

// defaultRunner is the global default runner used by the package.
var defaultRunner CommandRunner = &DefaultCommandRunner{}

// SetDefaultRunner sets the global default command runner.
// This is primarily used for testing.
func SetDefaultRunner(runner CommandRunner) {
	defaultRunner = runner
}

// GetDefaultRunner returns the current global default command runner.
func GetDefaultRunner() CommandRunner {
	return defaultRunner
}
