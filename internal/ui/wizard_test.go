package ui

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/quantmind-br/shotgun-cli/internal/core/scanner"
	"github.com/quantmind-br/shotgun-cli/internal/core/template"
	"github.com/quantmind-br/shotgun-cli/internal/ui/screens"
)

const (
	testTaskDescription = "Implement feature"
)

func TestWizardInitStartsScanCommand(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/project", &scanner.ScanConfig{MaxFiles: 10})
	cmd := wizard.Init()
	if cmd == nil {
		t.Fatalf("expected init command to be non-nil")
	}
	msg := cmd()
	scanMsg, ok := msg.(startScanMsg)
	if !ok {
		t.Fatalf("expected startScanMsg, got %T", msg)
	}
	if scanMsg.rootPath != "/tmp/project" {
		t.Fatalf("unexpected root path: %s", scanMsg.rootPath)
	}
}

func TestWizardHandlesScanLifecycle(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{MaxFiles: 5})

	var model tea.Model
	var cmd tea.Cmd
	model, cmd = wizard.Update(startScanMsg{rootPath: "/workspace", config: &scanner.ScanConfig{MaxFiles: 5}})
	wiz := model.(*WizardModel)
	if wiz.scanState == nil {
		t.Fatalf("expected scan state to be initialized")
	}
	if cmd == nil {
		t.Fatalf("expected iterative scan command to be scheduled")
	}

	// Simulate progress
	model, _ = wiz.Update(ScanProgressMsg{Current: 1, Total: 10, Stage: "scanning"})
	wiz = model.(*WizardModel)
	if !wiz.progress.Visible {
		t.Fatalf("expected progress visibility during scan")
	}

	// Simulate completion
	tree := &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true, Selected: true}
	model, _ = wiz.Update(ScanCompleteMsg{Tree: tree})
	wiz = model.(*WizardModel)
	if wiz.fileTree != tree {
		t.Fatalf("expected file tree to be set")
	}
	if wiz.fileSelection == nil {
		t.Fatalf("expected file selection model to be created")
	}
	if wiz.progress.Visible {
		t.Fatalf("progress should be hidden after completion")
	}
}

func TestWizardCanAdvanceStepLogic(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{})
	wizard.fileTree = &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true, Selected: true}
	wizard.selectedFiles["main.go"] = true
	wizard.template = &template.Template{Name: "basic"}
	wizard.taskDesc = testTaskDescription

	if !wizard.canAdvanceStep() {
		t.Fatalf("expected step 1 to advance when files selected")
	}

	wizard.step = StepTemplateSelection
	if !wizard.canAdvanceStep() {
		t.Fatalf("expected template selection to advance when template chosen")
	}

	wizard.step = StepTaskInput
	if !wizard.canAdvanceStep() {
		t.Fatalf("expected task input to advance when description provided")
	}

	wizard.step = StepRulesInput
	if !wizard.canAdvanceStep() {
		t.Fatalf("rules input should always be allowed to advance")
	}
}

func TestWizardGenerateContextCommand(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{})
	wizard.fileTree = &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true, Selected: true}
	wizard.selectedFiles["main.go"] = true
	wizard.template = &template.Template{Name: "basic"}
	wizard.taskDesc = testTaskDescription

	cmd := wizard.generateContext()
	if cmd == nil {
		t.Fatalf("expected generateContext to return command")
	}
	msg := cmd()
	startGen, ok := msg.(startGenerationMsg)
	if !ok {
		t.Fatalf("expected startGenerationMsg, got %T", msg)
	}
	if startGen.rootPath != "/workspace" {
		t.Fatalf("unexpected root path: %s", startGen.rootPath)
	}
}

func TestWizardGenerationFlowUpdatesState(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{})
	wizard.fileTree = &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true, Selected: true}
	wizard.selectedFiles["main.go"] = true
	wizard.template = &template.Template{Name: "basic"}
	wizard.taskDesc = testTaskDescription
	wizard.review = screens.NewReview(wizard.selectedFiles, wizard.template, wizard.taskDesc, "")
	var model tea.Model

	// Initialize generation state
	model, _ = wizard.Update(startGenerationMsg{
		fileTree:      wizard.fileTree,
		selectedFiles: wizard.selectedFiles,
		template:      wizard.template,
		taskDesc:      wizard.taskDesc,
		rules:         "",
		rootPath:      "/workspace",
	})
	wizard = model.(*WizardModel)

	if wizard.generateState == nil {
		t.Fatalf("expected generation state to be initialized")
	}

	// Progress message
	model, _ = wizard.Update(GenerationProgressMsg{Stage: "render", Message: "rendering"})
	wizard = model.(*WizardModel)
	if !wizard.progress.Visible {
		t.Fatalf("expected progress visible during generation")
	}

	// Completion message should schedule clipboard command
	msg := GenerationCompleteMsg{Content: "generated", FilePath: "/tmp/prompt.md"}
	model, clipboardCmd := wizard.Update(msg)
	wizard = model.(*WizardModel)
	if wizard.generatedFilePath != "/tmp/prompt.md" {
		t.Fatalf("expected generated file path to be stored")
	}
	if clipboardCmd == nil {
		t.Fatalf("expected clipboard command to be emitted")
	}

	// Simulate clipboard completion
	model, _ = wizard.Update(ClipboardCompleteMsg{Success: true})
	wizard = model.(*WizardModel)
	if wizard.error != nil {
		t.Fatalf("did not expect clipboard error, got %v", wizard.error)
	}
}

func TestWizardGenerateContextMissingTemplate(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{})
	wizard.selectedFiles["main.go"] = true

	cmd := wizard.generateContext()
	if cmd == nil {
		t.Fatalf("expected command to be returned")
	}
	msg := cmd()
	genErr, ok := msg.(GenerationErrorMsg)
	if !ok {
		t.Fatalf("expected GenerationErrorMsg, got %T", msg)
	}
	if genErr.Err == nil {
		t.Fatalf("expected error when template missing")
	}
}

func TestWizardClipboardFailureStored(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{})
	wizard.review = screens.NewReview(map[string]bool{}, nil, "", "")
	wizard.generatedFilePath = "/tmp/test.md"

	model, _ := wizard.Update(ClipboardCompleteMsg{Success: false, Err: fmt.Errorf("copy failed")})
	wizard = model.(*WizardModel)
	// Clipboard failure should not set error anymore (it just updates review screen)
	if wizard.error != nil {
		t.Fatalf("clipboard error should not be stored in model.error, got: %v", wizard.error)
	}
}

func TestWizardHandlesStructuredProgressMessages(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{})
	wizard.fileTree = &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true, Selected: true}
	wizard.selectedFiles["main.go"] = true
	wizard.template = &template.Template{Name: "basic"}

	var model tea.Model

	model, _ = wizard.Update(startGenerationMsg{
		fileTree:      wizard.fileTree,
		selectedFiles: wizard.selectedFiles,
		template:      wizard.template,
		taskDesc:      testTaskDescription,
		rootPath:      "/workspace",
	})
	wizard = model.(*WizardModel)

	stage := "collect"
	model, _ = wizard.Update(GenerationProgressMsg{Stage: stage, Message: "Collecting"})
	wizard = model.(*WizardModel)
	if wizard.progress.Stage != stage {
		t.Fatalf("expected progress stage %s, got %s", stage, wizard.progress.Stage)
	}
}

func TestWizardKeyboardNavigation(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{})
	wizard.fileTree = &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true, Selected: true}
	wizard.selectedFiles["main.go"] = true
	wizard.template = &template.Template{Name: "basic"}
	wizard.taskDesc = testTaskDescription

	wizard.step = StepFileSelection
	model, _ := wizard.Update(tea.KeyMsg{Type: tea.KeyF8})
	wizard = model.(*WizardModel)
	if wizard.step != StepTemplateSelection {
		t.Fatalf("expected to advance to template selection")
	}

	model, _ = wizard.Update(tea.KeyMsg{Type: tea.KeyF10})
	wizard = model.(*WizardModel)
	if wizard.step != StepFileSelection {
		t.Fatalf("expected to move back to file selection")
	}
}
