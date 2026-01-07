package ui

import (
	"fmt"
	"testing"
	"time"

	"github.com/quantmind-br/shotgun-cli/internal/core/context"
	"github.com/quantmind-br/shotgun-cli/internal/core/scanner"
	"github.com/quantmind-br/shotgun-cli/internal/core/template"
)

// mockGenerator is a test double for ContextGenerator
type mockGenerator struct {
	generateFunc               func(root *scanner.FileNode, selections map[string]bool, config context.GenerateConfig) (string, error)
	generateWithProgressFunc   func(root *scanner.FileNode, selections map[string]bool, config context.GenerateConfig, progress func(string)) (string, error)
	generateWithProgressExFunc func(root *scanner.FileNode, selections map[string]bool, config context.GenerateConfig, progress func(context.GenProgress)) (string, error)
}

func (m *mockGenerator) Generate(root *scanner.FileNode, selections map[string]bool, config context.GenerateConfig) (string, error) {
	if m.generateFunc != nil {
		return m.generateFunc(root, selections, config)
	}
	return "", nil
}

func (m *mockGenerator) GenerateWithProgress(root *scanner.FileNode, selections map[string]bool, config context.GenerateConfig, progress func(string)) (string, error) {
	if m.generateWithProgressFunc != nil {
		return m.generateWithProgressFunc(root, selections, config, progress)
	}
	return "", nil
}

func (m *mockGenerator) GenerateWithProgressEx(root *scanner.FileNode, selections map[string]bool, config context.GenerateConfig, progress func(context.GenProgress)) (string, error) {
	if m.generateWithProgressExFunc != nil {
		return m.generateWithProgressExFunc(root, selections, config, progress)
	}
	return "", nil
}

func TestGenerateCoordinator_New(t *testing.T) {
	t.Parallel()

	mockGen := &mockGenerator{}
	coord := NewGenerateCoordinator(mockGen)

	if coord.generator != mockGen {
		t.Error("generator not set correctly")
	}
	if coord.started {
		t.Error("started should be false initially")
	}
	if coord.content != "" {
		t.Error("content should be empty initially")
	}
}

func TestGenerateCoordinator_Start(t *testing.T) {
	t.Parallel()

	mockGen := &mockGenerator{
		generateFunc: func(root *scanner.FileNode, selections map[string]bool, config context.GenerateConfig) (string, error) {
			return "generated content", nil
		},
	}

	coord := NewGenerateCoordinator(mockGen)
	cfg := &GenerateConfig{
		FileTree: &scanner.FileNode{Name: "root"},
		Template: &template.Template{},
	}

	cmd := coord.Start(cfg)

	if cmd == nil {
		t.Fatal("Start should return a command")
	}
	if coord.progressCh == nil {
		t.Error("progressCh should be initialized")
	}
	if coord.done == nil {
		t.Error("done channel should be initialized")
	}
}

func TestGenerateCoordinator_Poll_BeforeStart(t *testing.T) {
	t.Parallel()

	coord := NewGenerateCoordinator(&mockGenerator{})
	cmd := coord.Poll()

	if cmd != nil {
		t.Error("Poll before Start should return nil")
	}
}

func TestGenerateCoordinator_Poll_DuringGeneration(t *testing.T) {
	t.Parallel()

	mockGen := &mockGenerator{}
	coord := NewGenerateCoordinator(mockGen)

	coord.progressCh = make(chan context.GenProgress, 10)
	coord.done = make(chan bool)

	go func() {
		coord.progressCh <- context.GenProgress{Stage: "testing", Message: "working"}
	}()

	cmd := coord.Poll()
	if cmd == nil {
		t.Fatal("Poll should return command when progress available")
	}
}

func TestGenerateCoordinator_Result_Success(t *testing.T) {
	t.Parallel()

	expectedContent := "final result"
	mockGen := &mockGenerator{
		generateFunc: func(root *scanner.FileNode, selections map[string]bool, config context.GenerateConfig) (string, error) {
			return expectedContent, nil
		},
	}

	coord := NewGenerateCoordinator(mockGen)
	cfg := &GenerateConfig{
		FileTree: &scanner.FileNode{Name: "root"},
		Template: &template.Template{},
	}

	cmd := coord.Start(cfg)
	msg := cmd()
	_ = msg

	for i := 0; i < 10; i++ {
		if coord.IsComplete() {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	content, err := coord.Result()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if content != expectedContent {
		t.Errorf("got content %q, want %q", content, expectedContent)
	}
}

func TestGenerateCoordinator_Result_Error(t *testing.T) {
	t.Parallel()

	expectedErr := fmt.Errorf("generation failed")
	mockGen := &mockGenerator{
		generateFunc: func(root *scanner.FileNode, selections map[string]bool, config context.GenerateConfig) (string, error) {
			return "", expectedErr
		},
	}

	coord := NewGenerateCoordinator(mockGen)
	cfg := &GenerateConfig{
		FileTree: &scanner.FileNode{Name: "root"},
		Template: &template.Template{},
	}

	cmd := coord.Start(cfg)
	cmd() // Start generation

	// Wait loop
	for i := 0; i < 10; i++ {
		if coord.IsComplete() {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	content, err := coord.Result()
	if err != expectedErr {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
	if content != "" {
		t.Errorf("got content %q, want empty", content)
	}
}

func TestGenerateCoordinator_Reset(t *testing.T) {
	t.Parallel()

	coord := NewGenerateCoordinator(&mockGenerator{})
	coord.Start(&GenerateConfig{})

	coord.Reset()

	if coord.started {
		t.Error("started should be false after reset")
	}
	if coord.config != nil {
		t.Error("config should be nil after reset")
	}
	if coord.progressCh != nil {
		t.Error("progressCh should be nil after reset")
	}
}
