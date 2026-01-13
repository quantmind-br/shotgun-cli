package geminiweb

import (
	"context"
	"errors"
	"os/exec"
	"sync"
)

// MockCommandRunner is a mock implementation of CommandRunner for testing.
type MockCommandRunner struct {
	mu              sync.Mutex
	lookPathFunc    func(file string) (string, error)
	lookPathCalls   []string
	commandFunc     func(ctx context.Context, name string, args ...string) *exec.Cmd
	commandCalls    []CommandCall
	executeCommands []*exec.Cmd
}

// CommandCall records a call to CommandContext.
type CommandCall struct {
	Name string
	Args []string
}

// NewMockRunner creates a new mock runner with default behavior.
// By default, it simulates a binary that is not found.
func NewMockRunner() *MockCommandRunner {
	return &MockCommandRunner{
		lookPathFunc: func(file string) (string, error) {
			return "", errors.New("binary not found")
		},
		commandFunc: func(ctx context.Context, name string, args ...string) *exec.Cmd {
			return exec.CommandContext(ctx, name, args...)
		},
	}
}

// NewMockRunnerAvailable creates a mock that simulates an available binary.
func NewMockRunnerAvailable() *MockCommandRunner {
	return &MockCommandRunner{
		lookPathFunc: func(file string) (string, error) {
			return "/usr/bin/geminiweb", nil
		},
		commandFunc: func(ctx context.Context, name string, args ...string) *exec.Cmd {
			return exec.CommandContext(ctx, name, args...)
		},
	}
}

// LookPath implements CommandRunner.
func (m *MockCommandRunner) LookPath(file string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lookPathCalls = append(m.lookPathCalls, file)
	if m.lookPathFunc != nil {
		return m.lookPathFunc(file)
	}
	return "", errors.New("not found")
}

// CommandContext implements CommandRunner.
func (m *MockCommandRunner) CommandContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.commandCalls = append(m.commandCalls, CommandCall{
		Name: name,
		Args: args,
	})
	m.executeCommands = append(m.executeCommands, exec.CommandContext(ctx, name, args...))
	if m.commandFunc != nil {
		return m.commandFunc(ctx, name, args...)
	}
	return exec.CommandContext(ctx, name, args...)
}

// SetLookPathFunc sets the function to use for LookPath calls.
func (m *MockCommandRunner) SetLookPathFunc(fn func(file string) (string, error)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lookPathFunc = fn
}

// GetLookPathCalls returns all recorded LookPath calls.
func (m *MockCommandRunner) GetLookPathCalls() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]string{}, m.lookPathCalls...)
}

// GetCommandCalls returns all recorded CommandContext calls.
func (m *MockCommandRunner) GetCommandCalls() []CommandCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	calls := make([]CommandCall, len(m.commandCalls))
	copy(calls, m.commandCalls)
	return calls
}

// WasLookPathCalled returns true if LookPath was called with the given file.
func (m *MockCommandRunner) WasLookPathCalled(file string) bool {
	for _, f := range m.GetLookPathCalls() {
		if f == file {
			return true
		}
	}
	return false
}

// WasCommandCalled returns true if CommandContext was called with the given name.
func (m *MockCommandRunner) WasCommandCalled(name string) bool {
	for _, call := range m.GetCommandCalls() {
		if call.Name == name {
			return true
		}
	}
	return false
}

// Reset clears all recorded calls.
func (m *MockCommandRunner) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lookPathCalls = nil
	m.commandCalls = nil
	m.executeCommands = nil
}
