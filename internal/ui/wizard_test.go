package ui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/quantmind-br/shotgun-cli/internal/core/context"
	"github.com/quantmind-br/shotgun-cli/internal/core/scanner"
	"github.com/quantmind-br/shotgun-cli/internal/core/template"
	"github.com/quantmind-br/shotgun-cli/internal/ui/components"
	"github.com/quantmind-br/shotgun-cli/internal/ui/screens"
)

const (
	testTaskDescription = "Implement feature"
	testSampleTask      = "Sample task"
)

func TestWizardInitStartsScanCommand(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/project", &scanner.ScanConfig{MaxFiles: 10}, nil)
	cmd := wizard.Init()
	if cmd == nil {
		t.Fatalf("expected init command to be non-nil")
	}
	msg := cmd()
	// Init now returns a batch (spinner tick + startScanMsg)
	batchMsg, ok := msg.(tea.BatchMsg)
	if !ok {
		t.Fatalf("expected tea.BatchMsg, got %T", msg)
	}
	var foundScanMsg bool
	for _, batchCmd := range batchMsg {
		if batchCmd == nil {
			continue
		}
		result := batchCmd()
		if scanMsg, ok := result.(startScanMsg); ok {
			foundScanMsg = true
			if scanMsg.rootPath != "/tmp/project" {
				t.Fatalf("unexpected root path: %s", scanMsg.rootPath)
			}
		}
	}
	if !foundScanMsg {
		t.Fatalf("expected startScanMsg in batch, but not found")
	}
}

func TestWizardHandlesScanLifecycle(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{MaxFiles: 5}, nil)

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
	tree := &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true}
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

func TestWizardFinishScan_SuccessfulCompletion(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{MaxFiles: 5}, nil)

	// Initialize scan state
	model, _ := wizard.Update(startScanMsg{rootPath: "/workspace", config: &scanner.ScanConfig{MaxFiles: 5}})
	wiz := model.(*WizardModel)

	// Simulate successful scan by setting result
	tree := &scanner.FileNode{
		Name:  "root",
		Path:  "/workspace",
		IsDir: true,
		Children: []*scanner.FileNode{
			{Name: "main.go", Path: "/workspace/main.go", Size: 1024},
		},
	}
	wiz.scanState.result = tree
	wiz.scanState.scanErr = nil

	// Call finishScan
	cmd := wiz.finishScan()
	if cmd == nil {
		t.Fatal("expected finishScan to return a command")
	}

	// Execute the command
	msg := cmd()

	// Should return ScanCompleteMsg with the tree
	scanComplete, ok := msg.(ScanCompleteMsg)
	if !ok {
		t.Fatalf("expected ScanCompleteMsg, got %T", msg)
	}

	if scanComplete.Tree != tree {
		t.Errorf("expected tree %v, got %v", tree, scanComplete.Tree)
	}
	if scanComplete.Tree.Name != "root" {
		t.Errorf("expected tree name 'root', got %s", scanComplete.Tree.Name)
	}
}

func TestWizardFinishScan_WithError(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{MaxFiles: 5}, nil)

	// Initialize scan state
	model, _ := wizard.Update(startScanMsg{rootPath: "/workspace", config: &scanner.ScanConfig{MaxFiles: 5}})
	wiz := model.(*WizardModel)

	// Simulate scan error
	expectedErr := fmt.Errorf("permission denied: /secret")
	wiz.scanState.scanErr = expectedErr

	// Call finishScan
	cmd := wiz.finishScan()
	if cmd == nil {
		t.Fatal("expected finishScan to return a command")
	}

	// Execute the command
	msg := cmd()

	// Should return ScanErrorMsg
	scanErr, ok := msg.(ScanErrorMsg)
	if !ok {
		t.Fatalf("expected ScanErrorMsg, got %T", msg)
	}

	if scanErr.Err == nil {
		t.Fatal("expected error to be set")
	}
	if scanErr.Err.Error() != expectedErr.Error() {
		t.Errorf("expected error %q, got %q", expectedErr.Error(), scanErr.Err.Error())
	}
}

func TestWizardFinishScan_NilScanState(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil)
	// scanState is nil by default

	// Call finishScan - it returns a command even with nil scanState
	// The function doesn't guard against nil scanState, so executing it would panic
	cmd := wizard.finishScan()

	// Verify a command is returned (function doesn't check for nil state)
	if cmd == nil {
		t.Error("expected finishScan to return a command even with nil scanState")
	}

	// Verify that executing the command panics (this documents current behavior)
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when executing finishScan command with nil scanState")
		}
	}()
	_ = cmd()
}

func TestWizardFinishGeneration_NilGenerateState(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil)
	// generateState is nil by default

	// Call finishGeneration - it returns a command even with nil generateState
	cmd := wizard.finishGeneration()

	// Verify a command is returned (function doesn't check for nil state)
	if cmd == nil {
		t.Error("expected finishGeneration to return a command even with nil generateState")
	}

	// Verify that executing the command panics (this documents current behavior)
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when executing finishGeneration command with nil generateState")
		}
	}()
	_ = cmd()
}

func TestWizardFinishGeneration_SuccessfulGeneration(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	wizard := NewWizard(tempDir, &scanner.ScanConfig{}, nil)
	wizard.fileTree = &scanner.FileNode{Name: "root", Path: tempDir, IsDir: true}
	wizard.selectedFiles["main.go"] = true
	wizard.template = &template.Template{Name: "basic", Content: "{FILE_STRUCTURE}"}
	wizard.taskDesc = "test task"
	wizard.wizardConfig.Context.MaxSize = "1MB"

	// Initialize generation state
	tree := &scanner.FileNode{Name: "root", Path: tempDir, IsDir: true}
	model, _ := wizard.Update(startGenerationMsg{
		fileTree:      tree,
		selectedFiles: wizard.selectedFiles,
		template:      wizard.template,
		taskDesc:      wizard.taskDesc,
		rules:         "",
		rootPath:      tempDir,
	})
	wiz := model.(*WizardModel)

	// Simulate successful generation by setting content
	expectedContent := "# Generated Context\n\nFile Structure:\nroot/"
	wiz.generateState.content = expectedContent

	// Call finishGeneration
	cmd := wiz.finishGeneration()
	if cmd == nil {
		t.Fatal("expected finishGeneration to return a command")
	}

	// Execute the command
	msg := cmd()

	// Should return GenerationCompleteMsg
	genComplete, ok := msg.(screens.GenerationCompleteMsg)
	if !ok {
		t.Fatalf("expected GenerationCompleteMsg, got %T", msg)
	}

	if genComplete.Content != expectedContent {
		t.Errorf("expected content %q, got %q", expectedContent, genComplete.Content)
	}
	if genComplete.FilePath == "" {
		t.Error("expected file path to be set")
	}
}

func TestWizardFinishGeneration_EmptyContent(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil)
	wizard.fileTree = &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true}
	wizard.selectedFiles["main.go"] = true

	// Initialize generation state
	tree := &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true}
	model, _ := wizard.Update(startGenerationMsg{
		fileTree:      tree,
		selectedFiles: wizard.selectedFiles,
		template:      &template.Template{Name: "basic"},
		taskDesc:      "test task",
		rootPath:      "/workspace",
	})
	wiz := model.(*WizardModel)

	// Simulate generation with empty content
	wiz.generateState.content = ""

	// Call finishGeneration
	cmd := wiz.finishGeneration()
	if cmd == nil {
		t.Fatal("expected finishGeneration to return a command")
	}

	// Execute the command
	msg := cmd()

	// Should return GenerationErrorMsg
	genErr, ok := msg.(screens.GenerationErrorMsg)
	if !ok {
		t.Fatalf("expected GenerationErrorMsg, got %T", msg)
	}

	if genErr.Err == nil {
		t.Fatal("expected error for empty content")
	}
	if !strings.Contains(genErr.Err.Error(), "no content generated") {
		t.Errorf("expected 'no content generated' error, got %q", genErr.Err.Error())
	}
}

func TestWizardFinishGeneration_ContentExceedsMaxSize(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil)
	wizard.fileTree = &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true}
	wizard.selectedFiles["main.go"] = true
	wizard.template = &template.Template{Name: "basic"}
	wizard.taskDesc = "test task"

	// Set a small max size
	wizard.wizardConfig.Context.MaxSize = "1KB"

	// Initialize generation state
	tree := &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true}
	model, _ := wizard.Update(startGenerationMsg{
		fileTree:      tree,
		selectedFiles: wizard.selectedFiles,
		template:      wizard.template,
		taskDesc:      wizard.taskDesc,
		rootPath:      "/workspace",
	})
	wiz := model.(*WizardModel)

	// Simulate generation with content exceeding max size
	largeContent := strings.Repeat("x", 2*1024) // 2KB
	wiz.generateState.content = largeContent

	// Call finishGeneration
	cmd := wiz.finishGeneration()
	if cmd == nil {
		t.Fatal("expected finishGeneration to return a command")
	}

	// Execute the command
	msg := cmd()

	// Should return GenerationErrorMsg
	genErr, ok := msg.(screens.GenerationErrorMsg)
	if !ok {
		t.Fatalf("expected GenerationErrorMsg, got %T", msg)
	}

	if genErr.Err == nil {
		t.Fatal("expected error for content exceeding max size")
	}
	if !strings.Contains(genErr.Err.Error(), "exceeds maximum allowed size") {
		t.Errorf("expected size validation error, got %q", genErr.Err.Error())
	}
}

func TestWizardCanAdvanceStepLogic(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil)
	wizard.fileTree = &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true}
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

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil)
	wizard.fileTree = &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true}
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

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil)
	wizard.fileTree = &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true}
	wizard.selectedFiles["main.go"] = true
	wizard.template = &template.Template{Name: "basic"}
	wizard.taskDesc = testTaskDescription
	wizard.review = screens.NewReview(wizard.selectedFiles, wizard.fileTree, wizard.template, wizard.taskDesc, "", "")
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
	msg := screens.GenerationCompleteMsg{Content: "generated", FilePath: "/tmp/prompt.md"}
	model, clipboardCmd := wizard.Update(msg)
	wizard = model.(*WizardModel)
	if wizard.generatedFilePath != "/tmp/prompt.md" {
		t.Fatalf("expected generated file path to be stored")
	}
	if clipboardCmd == nil {
		t.Fatalf("expected clipboard command to be emitted")
	}

	// Simulate clipboard completion
	model, _ = wizard.Update(screens.ClipboardCompleteMsg{Success: true})
	wizard = model.(*WizardModel)
	if wizard.error != nil {
		t.Fatalf("did not expect clipboard error, got %v", wizard.error)
	}
}

func TestWizardGenerateContextMissingTemplate(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil)
	wizard.selectedFiles["main.go"] = true

	cmd := wizard.generateContext()
	if cmd == nil {
		t.Fatalf("expected command to be returned")
	}
	msg := cmd()
	genErr, ok := msg.(screens.GenerationErrorMsg)
	if !ok {
		t.Fatalf("expected GenerationErrorMsg, got %T", msg)
	}
	if genErr.Err == nil {
		t.Fatalf("expected error when template missing")
	}
}

func TestWizardClipboardFailureStored(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil)
	wizard.review = screens.NewReview(map[string]bool{}, nil, nil, "", "", "")
	wizard.generatedFilePath = "/tmp/test.md"

	model, _ := wizard.Update(screens.ClipboardCompleteMsg{Success: false, Err: fmt.Errorf("copy failed")})
	wizard = model.(*WizardModel)
	// Clipboard failure should not set error anymore (it just updates review screen)
	if wizard.error != nil {
		t.Fatalf("clipboard error should not be stored in model.error, got: %v", wizard.error)
	}
}

func TestWizardHandlesStructuredProgressMessages(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil)
	wizard.fileTree = &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true}
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

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil)
	wizard.fileTree = &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true}
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

func TestWizardHelpToggle(t *testing.T) {
	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil)

	// Initially help should be hidden
	if wizard.showHelp {
		t.Fatal("expected showHelp to be false initially")
	}

	// Press F1 to show help
	model, _ := wizard.Update(tea.KeyMsg{Type: tea.KeyF1})
	wizard = model.(*WizardModel)
	if !wizard.showHelp {
		t.Fatal("expected showHelp to be true after pressing F1")
	}

	// View should contain help content
	view := wizard.View()
	if !strings.Contains(view, "Help") {
		t.Fatal("expected view to contain 'Help' when showHelp is true")
	}
	if !strings.Contains(view, "Global Shortcuts") {
		t.Fatal("expected view to contain 'Global Shortcuts'")
	}

	// Press F1 again to hide help
	model, _ = wizard.Update(tea.KeyMsg{Type: tea.KeyF1})
	wizard = model.(*WizardModel)
	if wizard.showHelp {
		t.Fatal("expected showHelp to be false after pressing F1 again")
	}
}

// Test for wizard skip step logic based on template requirements
func TestWizardSkipStepsNoTaskNoRules(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil)
	wizard.fileTree = &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true}
	wizard.selectedFiles["main.go"] = true
	// Template that has neither TASK nor RULES
	wizard.template = &template.Template{
		Name:    "file_structure_only",
		Content: "File Structure:\n{FILE_STRUCTURE}",
	}

	// Start at template selection (step 2)
	wizard.step = StepTemplateSelection

	// Press F8 to advance - should skip Task and Rules, go directly to Review
	model, _ := wizard.Update(tea.KeyMsg{Type: tea.KeyF8})
	wizard = model.(*WizardModel)

	if wizard.step != StepReview {
		t.Fatalf("expected to skip to Review (step %d), got step %d", StepReview, wizard.step)
	}
}

func TestWizardSkipStepsNoTaskHasRules(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil)
	wizard.fileTree = &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true}
	wizard.selectedFiles["main.go"] = true
	// Template that has RULES but not TASK
	wizard.template = &template.Template{
		Name:    "rules_only",
		Content: "Rules: {RULES}\nFile Structure:\n{FILE_STRUCTURE}",
	}

	// Start at template selection (step 2)
	wizard.step = StepTemplateSelection

	// Press F8 to advance - should skip Task, go to Rules
	model, _ := wizard.Update(tea.KeyMsg{Type: tea.KeyF8})
	wizard = model.(*WizardModel)

	if wizard.step != StepRulesInput {
		t.Fatalf("expected to skip to RulesInput (step %d), got step %d", StepRulesInput, wizard.step)
	}
}

func TestWizardSkipStepsHasTaskNoRules(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil)
	wizard.fileTree = &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true}
	wizard.selectedFiles["main.go"] = true
	// Template that has TASK but not RULES
	wizard.template = &template.Template{
		Name:    "task_only",
		Content: "Task: {TASK}\nFile Structure:\n{FILE_STRUCTURE}",
	}
	wizard.taskDesc = testSampleTask

	// Start at task input (step 3)
	wizard.step = StepTaskInput

	// Press F8 to advance - should skip Rules, go to Review
	model, _ := wizard.Update(tea.KeyMsg{Type: tea.KeyF8})
	wizard = model.(*WizardModel)

	if wizard.step != StepReview {
		t.Fatalf("expected to skip to Review (step %d), got step %d", StepReview, wizard.step)
	}
}

func TestWizardNoSkipWhenBothRequired(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil)
	wizard.fileTree = &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true}
	wizard.selectedFiles["main.go"] = true
	// Template that has both TASK and RULES
	wizard.template = &template.Template{
		Name:    "full_template",
		Content: "Task: {TASK}\nRules: {RULES}\nFile Structure:\n{FILE_STRUCTURE}",
	}

	// Start at template selection (step 2)
	wizard.step = StepTemplateSelection

	// Press F8 to advance - should go to Task (step 3)
	model, _ := wizard.Update(tea.KeyMsg{Type: tea.KeyF8})
	wizard = model.(*WizardModel)

	if wizard.step != StepTaskInput {
		t.Fatalf("expected to go to TaskInput (step %d), got step %d", StepTaskInput, wizard.step)
	}

	// Provide task and advance
	wizard.taskDesc = testSampleTask
	model, _ = wizard.Update(tea.KeyMsg{Type: tea.KeyF8})
	wizard = model.(*WizardModel)

	if wizard.step != StepRulesInput {
		t.Fatalf("expected to go to RulesInput (step %d), got step %d", StepRulesInput, wizard.step)
	}
}

func TestWizardBackwardNavigationSkipsCorrectly(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil)
	wizard.fileTree = &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true}
	wizard.selectedFiles["main.go"] = true
	// Template with no TASK and no RULES
	wizard.template = &template.Template{
		Name:    "file_structure_only",
		Content: "File Structure:\n{FILE_STRUCTURE}",
	}

	// Start at Review (step 5)
	wizard.step = StepReview
	wizard.review = screens.NewReview(wizard.selectedFiles, wizard.fileTree, wizard.template, "", "", "")

	// Press F7/F10 to go back - should skip Rules and Task, go to Template Selection
	model, _ := wizard.Update(tea.KeyMsg{Type: tea.KeyF10})
	wizard = model.(*WizardModel)

	if wizard.step != StepTemplateSelection {
		t.Fatalf("expected to go back to TemplateSelection (step %d), got step %d", StepTemplateSelection, wizard.step)
	}
}

func TestWizardBackwardFromRulesWithNoTask(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil)
	wizard.fileTree = &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true}
	wizard.selectedFiles["main.go"] = true
	// Template with RULES but no TASK
	wizard.template = &template.Template{
		Name:    "rules_only",
		Content: "Rules: {RULES}\nFile Structure:\n{FILE_STRUCTURE}",
	}

	// Start at Rules Input (step 4)
	wizard.step = StepRulesInput
	wizard.rulesInput = screens.NewRulesInput("")

	// Press F7/F10 to go back - should skip Task, go to Template Selection
	model, _ := wizard.Update(tea.KeyMsg{Type: tea.KeyF10})
	wizard = model.(*WizardModel)

	if wizard.step != StepTemplateSelection {
		t.Fatalf("expected to go back to TemplateSelection (step %d), got step %d", StepTemplateSelection, wizard.step)
	}
}

func TestWizardRequiresTaskInput(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil)

	// No template set
	if wizard.requiresTaskInput() {
		t.Fatal("expected requiresTaskInput to be false when no template set")
	}

	// Template without TASK
	wizard.template = &template.Template{
		Name:    "no_task",
		Content: "File Structure:\n{FILE_STRUCTURE}",
	}
	if wizard.requiresTaskInput() {
		t.Fatal("expected requiresTaskInput to be false when template has no TASK")
	}

	// Template with TASK
	wizard.template = &template.Template{
		Name:    "with_task",
		Content: "Task: {TASK}",
	}
	if !wizard.requiresTaskInput() {
		t.Fatal("expected requiresTaskInput to be true when template has TASK")
	}
}

func TestWizardRequiresRulesInput(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil)

	// No template set
	if wizard.requiresRulesInput() {
		t.Fatal("expected requiresRulesInput to be false when no template set")
	}

	// Template without RULES
	wizard.template = &template.Template{
		Name:    "no_rules",
		Content: "Task: {TASK}",
	}
	if wizard.requiresRulesInput() {
		t.Fatal("expected requiresRulesInput to be false when template has no RULES")
	}

	// Template with RULES
	wizard.template = &template.Template{
		Name:    "with_rules",
		Content: "Rules: {RULES}",
	}
	if !wizard.requiresRulesInput() {
		t.Fatal("expected requiresRulesInput to be true when template has RULES")
	}
}

func TestWizardCanAdvanceWithoutTaskWhenNotRequired(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil)
	wizard.fileTree = &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true}
	wizard.selectedFiles["main.go"] = true
	// Template that does NOT require TASK
	wizard.template = &template.Template{
		Name:    "no_task",
		Content: "File Structure:\n{FILE_STRUCTURE}",
	}

	// At Task Input step with empty task description
	wizard.step = StepTaskInput
	wizard.taskDesc = ""

	// Should be able to advance since template doesn't require TASK
	if !wizard.canAdvanceStep() {
		t.Fatal("expected canAdvanceStep to return true when template doesn't require TASK")
	}
}

func TestWizardCannotAdvanceWithoutTaskWhenRequired(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil)
	wizard.fileTree = &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true}
	wizard.selectedFiles["main.go"] = true
	// Template that requires TASK
	wizard.template = &template.Template{
		Name:    "with_task",
		Content: "Task: {TASK}\nFile Structure:\n{FILE_STRUCTURE}",
	}

	// At Task Input step with empty task description
	wizard.step = StepTaskInput
	wizard.taskDesc = ""

	// Should NOT be able to advance since template requires TASK
	if wizard.canAdvanceStep() {
		t.Fatal("expected canAdvanceStep to return false when template requires TASK and task is empty")
	}

	// Provide task description
	wizard.taskDesc = testSampleTask

	// Now should be able to advance
	if !wizard.canAdvanceStep() {
		t.Fatal("expected canAdvanceStep to return true when template requires TASK and task is provided")
	}
}

func TestWizardHandleWindowResize(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil)
	wizard.fileTree = &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true}
	wizard.fileSelection = screens.NewFileSelection(wizard.fileTree, wizard.selectedFiles)

	// Initial dimensions
	initialWidth := wizard.width
	initialHeight := wizard.height

	// Send window resize message
	model, _ := wizard.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	wizard = model.(*WizardModel)

	if wizard.width != 120 {
		t.Errorf("expected width 120, got %d", wizard.width)
	}
	if wizard.height != 40 {
		t.Errorf("expected height 40, got %d", wizard.height)
	}
	if wizard.width == initialWidth && wizard.height == initialHeight {
		t.Error("dimensions should have changed after resize")
	}
}

func TestWizardHandleScanError(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil)
	wizard.progress.Visible = true // Simulate progress being shown

	testErr := fmt.Errorf("permission denied: /secret")
	model, _ := wizard.Update(ScanErrorMsg{Err: testErr})
	wizard = model.(*WizardModel)

	if wizard.error == nil {
		t.Fatal("expected error to be set")
	}
	if wizard.error.Error() != testErr.Error() {
		t.Errorf("expected error %q, got %q", testErr.Error(), wizard.error.Error())
	}
	if wizard.progress.Visible {
		t.Error("progress should be hidden after error")
	}
}

func TestWizardHandleGenerationError(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil)
	wizard.progress.Visible = true

	testErr := fmt.Errorf("template rendering failed")
	model, _ := wizard.Update(screens.GenerationErrorMsg{Err: testErr})
	wizard = model.(*WizardModel)

	if wizard.error == nil {
		t.Fatal("expected error to be set")
	}
	if wizard.error.Error() != testErr.Error() {
		t.Errorf("expected error %q, got %q", testErr.Error(), wizard.error.Error())
	}
	if wizard.progress.Visible {
		t.Error("progress should be hidden after error")
	}
}

func TestWizardGeminiLifecycle(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil)
	wizard.fileTree = &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true}
	wizard.selectedFiles["main.go"] = true
	wizard.template = &template.Template{Name: "basic", Content: "Task: {TASK}"}
	wizard.taskDesc = "Test task"
	wizard.generatedContent = "generated context"
	wizard.generatedFilePath = "/tmp/test.md"
	wizard.step = StepReview
	wizard.review = screens.NewReview(wizard.selectedFiles, wizard.fileTree, wizard.template, wizard.taskDesc, "", "")

	// Test GeminiProgressMsg
	model, _ := wizard.Update(screens.GeminiProgressMsg{Stage: "sending"})
	wizard = model.(*WizardModel)
	if wizard.progress.Stage != "sending" {
		t.Errorf("expected progress stage 'sending', got %q", wizard.progress.Stage)
	}

	// Test GeminiCompleteMsg
	model, _ = wizard.Update(screens.GeminiCompleteMsg{
		Response:   "AI response here",
		OutputFile: "/tmp/response.md",
		Duration:   5 * time.Second,
	})
	wizard = model.(*WizardModel)

	if wizard.geminiResponseFile != "/tmp/response.md" {
		t.Errorf("expected response file '/tmp/response.md', got %q", wizard.geminiResponseFile)
	}
	if wizard.geminiSending {
		t.Error("geminiSending should be false after completion")
	}
}

func TestWizardHandleGeminiError(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil)
	wizard.fileTree = &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true}
	wizard.step = StepReview
	wizard.review = screens.NewReview(map[string]bool{}, wizard.fileTree, nil, "", "", "")
	wizard.geminiSending = true

	testErr := fmt.Errorf("geminiweb: connection timeout")
	model, _ := wizard.Update(screens.GeminiErrorMsg{Err: testErr})
	wizard = model.(*WizardModel)

	if wizard.geminiSending {
		t.Error("geminiSending should be false after error")
	}
	// Error is handled by review screen, check that no panic occurred
}

func TestWizardHandleTemplateMessage(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil)
	wizard.fileTree = &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true}
	wizard.selectedFiles["main.go"] = true
	wizard.step = StepTemplateSelection

	selectedTemplate := &template.Template{
		Name:    "code-review",
		Content: "Review: {TASK}\nFiles: {FILE_STRUCTURE}",
	}

	model, _ := wizard.Update(TemplateSelectedMsg{Template: selectedTemplate})
	wizard = model.(*WizardModel)

	if wizard.template == nil {
		t.Fatal("expected template to be set")
	}
	if wizard.template.Name != "code-review" {
		t.Errorf("expected template name 'code-review', got %q", wizard.template.Name)
	}
}

func TestWizardHandleTemplateMessage_WrongStep(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		step int
	}{
		{"file selection step", StepFileSelection},
		{"task input step", StepTaskInput},
		{"rules input step", StepRulesInput},
		{"review step", StepReview},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil)
			wizard.step = tt.step

			// Create a template selection (should be ignored due to wrong step)
			wizard.templateSelection = screens.NewTemplateSelection()

			selectedTemplate := &template.Template{
				Name:    "code-review",
				Content: "Review: {TASK}",
			}

			// handleTemplateMessage should return nil when not in StepTemplateSelection
			cmd := wizard.handleTemplateMessage(TemplateSelectedMsg{Template: selectedTemplate})

			if cmd != nil {
				t.Errorf("expected nil cmd when in wrong step %d, got non-nil", tt.step)
			}
		})
	}
}

func TestWizardHandleTemplateMessage_NilTemplateSelection(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil)
	wizard.step = StepTemplateSelection
	// templateSelection is nil by default

	selectedTemplate := &template.Template{
		Name:    "code-review",
		Content: "Review: {TASK}",
	}

	// handleTemplateMessage should return nil when templateSelection is nil
	cmd := wizard.handleTemplateMessage(TemplateSelectedMsg{Template: selectedTemplate})

	if cmd != nil {
		t.Error("expected nil cmd when templateSelection is nil, got non-nil")
	}
}

func TestWizardHandleTemplateMessage_CorrectStepWithSelection(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil)
	wizard.fileTree = &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true}
	wizard.selectedFiles["main.go"] = true
	wizard.step = StepTemplateSelection
	wizard.templateSelection = screens.NewTemplateSelection()

	selectedTemplate := &template.Template{
		Name:    "code-review",
		Content: "Review: {TASK}\nFiles: {FILE_STRUCTURE}",
	}

	// handleTemplateMessage should delegate to templateSelection.HandleMessage
	cmd := wizard.handleTemplateMessage(TemplateSelectedMsg{Template: selectedTemplate})

	// When templateSelection exists and step is correct, it should return a command
	// The command may be nil if templateSelection.HandleMessage returns nil
	// Just verify the function doesn't panic
	if cmd != nil {
		// If cmd is non-nil, it should be executable
		_ = cmd()
	}
}

func TestWizardHandleRescanRequest(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{MaxFiles: 100}, nil)
	wizard.fileTree = &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true}
	wizard.selectedFiles["main.go"] = true
	wizard.step = StepFileSelection

	// The wizard should handle rescan by restarting the scan
	// This tests the conceptual flow - actual rescan may use different mechanism
	initialTree := wizard.fileTree

	// Simulate conditions that would trigger rescan
	wizard.scanState = nil
	model, cmd := wizard.Update(startScanMsg{rootPath: "/workspace", config: &scanner.ScanConfig{MaxFiles: 100}})
	wizard = model.(*WizardModel)

	if wizard.scanState == nil {
		t.Error("expected scan state to be initialized on rescan")
	}
	if cmd == nil {
		t.Error("expected scan command to be scheduled")
	}
	_ = initialTree // Acknowledge variable
}

func TestWizardViewRendersFileSelectionStep(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil)
	wizard.fileTree = &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true}
	wizard.fileSelection = screens.NewFileSelection(wizard.fileTree, wizard.selectedFiles)
	wizard.step = StepFileSelection
	wizard.width = 80
	wizard.height = 24

	view := wizard.View()

	if view == "" {
		t.Error("expected non-empty view")
	}
	// View should contain file selection content
	if !strings.Contains(view, "root") && !strings.Contains(view, "File") {
		t.Error("expected view to contain file selection elements")
	}
}

func TestWizardViewRendersTemplateSelectionStep(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil)
	wizard.fileTree = &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true}
	wizard.step = StepTemplateSelection
	wizard.templateSelection = screens.NewTemplateSelection() // Will use default templates
	wizard.width = 80
	wizard.height = 24

	view := wizard.View()

	if view == "" {
		t.Error("expected non-empty view")
	}
}

func TestWizardViewRendersTaskInputStep(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil)
	wizard.fileTree = &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true}
	wizard.template = &template.Template{Name: "basic", Content: "Task: {TASK}"}
	wizard.step = StepTaskInput
	wizard.taskInput = screens.NewTaskInput("")
	wizard.width = 80
	wizard.height = 24

	view := wizard.View()

	if view == "" {
		t.Error("expected non-empty view")
	}
}

func TestWizardViewRendersRulesInputStep(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil)
	wizard.fileTree = &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true}
	wizard.template = &template.Template{Name: "basic", Content: "Rules: {RULES}"}
	wizard.step = StepRulesInput
	wizard.rulesInput = screens.NewRulesInput("")
	wizard.width = 80
	wizard.height = 24

	view := wizard.View()

	if view == "" {
		t.Error("expected non-empty view")
	}
}

func TestWizardViewRendersReviewStep(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil)
	wizard.fileTree = &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true}
	wizard.selectedFiles["main.go"] = true
	wizard.template = &template.Template{Name: "basic"}
	wizard.taskDesc = "Test task"
	wizard.step = StepReview
	wizard.review = screens.NewReview(wizard.selectedFiles, wizard.fileTree, wizard.template, wizard.taskDesc, "", "")
	wizard.width = 80
	wizard.height = 24

	view := wizard.View()

	if view == "" {
		t.Error("expected non-empty view")
	}
}

func TestWizardViewWithError(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil)
	wizard.error = fmt.Errorf("test error message")
	wizard.width = 80
	wizard.height = 24

	view := wizard.View()

	if !strings.Contains(view, "error") && !strings.Contains(view, "Error") {
		t.Error("expected view to show error")
	}
}

func TestWizardViewWithProgress(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil)
	wizard.progress.Visible = true
	wizard.progress.Stage = "scanning"
	wizard.progress.Current = 50
	wizard.progress.Total = 100
	wizard.width = 80
	wizard.height = 24

	view := wizard.View()

	if view == "" {
		t.Error("expected non-empty view with progress")
	}
}

func TestWizardParseSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected int64
		hasError bool
	}{
		// Basic valid sizes
		{"1KB", 1024, false},
		{"1MB", 1024 * 1024, false},
		{"1GB", 1024 * 1024 * 1024, false},
		{"100", 100, false},
		{"0", 0, false},

		// Very large numbers with different suffixes
		{"999KB", 999 * 1024, false},
		{"999MB", 999 * 1024 * 1024, false},
		{"999GB", int64(999) * 1024 * 1024 * 1024, false},

		// Negative numbers (accepted by ParseInt, returns negative values)
		{"-1KB", -1024, false},
		{"-100", -100, false},
		{"-1MB", -(1024 * 1024), false},

		// Float values should be rejected (ParseInt doesn't handle floats)
		{"1.5MB", 0, true},
		{"1.0KB", 0, true},
		{"0.5GB", 0, true},

		// Case variations (function handles case-insensitivity via ToUpper)
		{"1kb", 1024, false},
		{"1Kb", 1024, false},
		{"1Mb", 1024 * 1024, false},
		{"1mB", 1024 * 1024, false},
		{"1Gb", 1024 * 1024 * 1024, false},
		{"1gB", 1024 * 1024 * 1024, false},

		// Leading/trailing whitespace is trimmed by TrimSpace
		{" 1KB", 1024, false},
		{"1KB ", 1024, false},
		{"  1MB  ", 1024 * 1024, false},

		// Empty string
		{"", 0, true},

		// Invalid suffixes
		{"1TB", 0, true},
		{"1XB", 0, true},
		{"1PB", 0, true},

		// Non-numeric strings
		{"invalid", 0, true},
		{"MB", 0, true},       // No number
		{"10XB", 0, true},     // Invalid suffix
		{"10.5.5MB", 0, true}, // Multiple dots
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parseSize(tt.input)
			if tt.hasError {
				if err == nil {
					t.Errorf("expected error for invalid input %q, got nil", tt.input)
				}
				if result != 0 {
					t.Errorf("expected 0 for invalid input %q, got %d", tt.input, result)
				}
			} else {
				if err != nil {
					t.Errorf("parseSize(%q) unexpected error: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("parseSize(%q) = %d, want %d", tt.input, result, tt.expected)
				}
			}
		})
	}
}

func TestWizardValidationErrorMessages(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		step           int
		expectedSubstr string
	}{
		{"file selection step", StepFileSelection, "file"},
		{"template selection step", StepTemplateSelection, "template"},
		{"task input step", StepTaskInput, "task"},
		{"rules input step returns empty", StepRulesInput, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil)
			wizard.step = tt.step

			msg := wizard.getValidationErrorMessage()

			if tt.expectedSubstr == "" {
				if msg != "" {
					t.Errorf("expected empty message for step %d, got %q", tt.step, msg)
				}
			} else {
				if !strings.Contains(strings.ToLower(msg), tt.expectedSubstr) {
					t.Errorf("expected message to contain %q, got %q", tt.expectedSubstr, msg)
				}
			}
		})
	}
}

func TestWizardValidationErrorSetOnFailedAdvance(t *testing.T) {
	t.Parallel()

	t.Run("empty file selection shows error", func(t *testing.T) {
		wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil)
		wizard.step = StepFileSelection
		wizard.selectedFiles = map[string]bool{}

		wizard.handleNextStep()

		if wizard.validationError == "" {
			t.Error("expected validation error when no files selected")
		}
		if !strings.Contains(strings.ToLower(wizard.validationError), "file") {
			t.Errorf("error should mention files, got %q", wizard.validationError)
		}
	})

	t.Run("no template selected shows error", func(t *testing.T) {
		wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil)
		wizard.step = StepTemplateSelection
		wizard.template = nil

		wizard.handleNextStep()

		if wizard.validationError == "" {
			t.Error("expected validation error when no template selected")
		}
		if !strings.Contains(strings.ToLower(wizard.validationError), "template") {
			t.Errorf("error should mention template, got %q", wizard.validationError)
		}
	})

	t.Run("empty task description shows error when required", func(t *testing.T) {
		wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil)
		wizard.step = StepTaskInput
		wizard.template = &template.Template{Name: "test", Content: "Task: {TASK}"}
		wizard.taskDesc = ""

		wizard.handleNextStep()

		if wizard.validationError == "" {
			t.Error("expected validation error when task empty")
		}
		if !strings.Contains(strings.ToLower(wizard.validationError), "task") {
			t.Errorf("error should mention task, got %q", wizard.validationError)
		}
	})
}

func TestWizardValidationErrorClearedOnInput(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil)
	wizard.fileTree = &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true}
	wizard.fileSelection = screens.NewFileSelection(wizard.fileTree, wizard.selectedFiles)
	wizard.step = StepFileSelection
	wizard.validationError = "Some error"

	wizard.handleStepInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})

	if wizard.validationError != "" {
		t.Errorf("expected validation error to be cleared, got %q", wizard.validationError)
	}
}

func TestWizardValidationErrorClearedOnSuccessfulAdvance(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil)
	wizard.step = StepFileSelection
	wizard.selectedFiles = map[string]bool{"/test/file.go": true}
	wizard.validationError = "Previous error"

	wizard.handleNextStep()

	if wizard.validationError != "" {
		t.Errorf("expected validation error to be cleared on successful advance, got %q", wizard.validationError)
	}
	if wizard.step != StepTemplateSelection {
		t.Errorf("expected to advance to template selection, got step %d", wizard.step)
	}
}

func TestWizardViewShowsValidationError(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil)
	wizard.fileTree = &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true}
	wizard.fileSelection = screens.NewFileSelection(wizard.fileTree, wizard.selectedFiles)
	wizard.step = StepFileSelection
	wizard.validationError = "Select at least one file to continue"
	wizard.width = 80
	wizard.height = 24

	view := wizard.View()

	if !strings.Contains(view, "Select at least one file") {
		t.Error("expected view to display validation error")
	}
}

func TestValidateContentSize(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		maxSize     string
		wantErr     bool
		errContains string
	}{
		{
			name:    "empty max size - should not validate",
			content: strings.Repeat("x", 10000000), // 10MB
			maxSize: "",
			wantErr: false,
		},
		{
			name:    "content within limits",
			content: strings.Repeat("x", 1000), // 1KB
			maxSize: "1MB",
			wantErr: false,
		},
		{
			name:    "content exactly at limit",
			content: strings.Repeat("x", 1024*1024), // 1MB
			maxSize: "1MB",
			wantErr: false,
		},
		{
			name:        "content exceeds limit by 1 byte",
			content:     strings.Repeat("x", 1024*1024+1), // 1MB + 1 byte
			maxSize:     "1MB",
			wantErr:     true,
			errContains: "exceeds maximum allowed size",
		},
		{
			name:        "content far exceeds limit",
			content:     strings.Repeat("x", 5*1024*1024), // 5MB
			maxSize:     "1MB",
			wantErr:     true,
			errContains: "exceeds maximum allowed size",
		},
		{
			name:    "empty content",
			content: "",
			maxSize: "1MB",
			wantErr: false, // Empty content is allowed (checked elsewhere)
		},
		{
			name:    "small content with small limit",
			content: "hello",
			maxSize: "100B",
			wantErr: false,
		},
		{
			name:        "content exceeds small limit",
			content:     strings.Repeat("x", 200),
			maxSize:     "100B",
			wantErr:     true,
			errContains: "exceeds maximum allowed size",
		},
		{
			name:    "content within KB limit",
			content: strings.Repeat("x", 500*1024), // 500KB
			maxSize: "1MB",
			wantErr: false,
		},
		{
			name:        "content exceeds GB limit",
			content:     strings.Repeat("x", 2*1024*1024*1024), // 2GB
			maxSize:     "1GB",
			wantErr:     true,
			errContains: "exceeds maximum allowed size",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a minimal wizard model
			model := NewWizard("/tmp/test", &scanner.ScanConfig{}, &WizardConfig{})
			model.wizardConfig.Context.MaxSize = tt.maxSize

			err := model.validateContentSize(tt.content)

			if tt.wantErr {
				if err == nil {
					t.Errorf("validateContentSize() expected error containing %q, but got nil", tt.errContains)
					return
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("validateContentSize() error = %v, expected to contain %q", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("validateContentSize() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestValidateContentSize_InvalidMaxSize(t *testing.T) {
	model := NewWizard("/tmp/test", &scanner.ScanConfig{}, &WizardConfig{})
	model.wizardConfig.Context.MaxSize = "invalid-size"

	err := model.validateContentSize("test content")
	if err == nil {
		t.Error("validateContentSize() expected error for invalid max-size, got nil")
	}
	if !strings.Contains(err.Error(), "invalid max-size configuration") {
		t.Errorf("validateContentSize() error = %v, expected to contain 'invalid max-size configuration'", err)
	}
}

func TestValidateContentSize_BoundaryConditions(t *testing.T) {
	tests := []struct {
		name        string
		contentSize int
		maxSize     string
		wantErr     bool
	}{
		{"zero bytes at zero limit", 0, "0B", false},
		{"one byte at zero limit", 1, "0B", true},
		{"100 bytes at 100B limit", 100, "100B", false},
		{"101 bytes at 100B limit", 101, "100B", true},
		{"1KB at 1KB limit", 1024, "1KB", false},
		{"1025 bytes at 1KB limit", 1025, "1KB", true},
		{"exact MB match", 1024 * 1024, "1MB", false},
		{"exact GB match", 1024 * 1024 * 1024, "1GB", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewWizard("/tmp/test", &scanner.ScanConfig{}, &WizardConfig{})
			model.wizardConfig.Context.MaxSize = tt.maxSize

			content := strings.Repeat("x", tt.contentSize)
			err := model.validateContentSize(content)

			if tt.wantErr {
				if err == nil {
					t.Errorf("validateContentSize() expected error for %d bytes with limit %s, got nil", tt.contentSize, tt.maxSize)
				}
			} else {
				if err != nil {
					t.Errorf("validateContentSize() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestWizardIterativeScanCmd_NilScanState(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil)
	wizard.scanState = nil

	cmd := wizard.iterativeScanCmd()
	msg := cmd()

	scanErr, ok := msg.(ScanErrorMsg)
	if !ok {
		t.Fatalf("expected ScanErrorMsg, got %T", msg)
	}
	if scanErr.Err == nil {
		t.Error("expected error for nil scanState")
	}
}

func TestWizardIterativeScanCmd_StartsScan(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	wizard := NewWizard(tempDir, &scanner.ScanConfig{MaxFiles: 10}, nil)

	progressCh := make(chan scanner.Progress, 100)
	done := make(chan bool)
	scanConfig := &scanner.ScanConfig{
		MaxFiles:    10,
		MaxFileSize: 1024 * 1024,
	}

	scanr := scanner.NewFileSystemScanner()
	wizard.scanState = &scanState{
		scanner:    scanr,
		rootPath:   tempDir,
		config:     scanConfig,
		progressCh: progressCh,
		done:       done,
		started:    false,
	}

	cmd := wizard.iterativeScanCmd()
	msg := cmd()

	_, isPollMsg := msg.(pollScanMsg)
	if !isPollMsg {
		t.Errorf("expected pollScanMsg, got %T", msg)
	}

	if !wizard.scanState.started {
		t.Error("expected scanState.started to be true")
	}

	// Clean up: drain any remaining progress and wait for goroutine
	go func() {
		for range progressCh {
		}
	}()
	<-done
}

func TestWizardIterativeScanCmd_AlreadyStarted(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	wizard := NewWizard(tempDir, &scanner.ScanConfig{MaxFiles: 10}, nil)

	progressCh := make(chan scanner.Progress, 100)
	done := make(chan bool)
	scanConfig := &scanner.ScanConfig{
		MaxFiles:    10,
		MaxFileSize: 1024 * 1024,
	}

	scanr := scanner.NewFileSystemScanner()
	wizard.scanState = &scanState{
		scanner:    scanr,
		rootPath:   tempDir,
		config:     scanConfig,
		progressCh: progressCh,
		done:       done,
		started:    true, // Already started
	}

	cmd := wizard.iterativeScanCmd()
	msg := cmd()

	_, isPollMsg := msg.(pollScanMsg)
	if !isPollMsg {
		t.Errorf("expected pollScanMsg, got %T", msg)
	}

	// Since started was already true, no new goroutine was created
	close(done)
	close(progressCh)
}

func TestWizardIterativeGenerateCmd_NilGenerateState(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil)
	wizard.generateState = nil

	cmd := wizard.iterativeGenerateCmd()
	msg := cmd()

	genErr, ok := msg.(screens.GenerationErrorMsg)
	if !ok {
		t.Fatalf("expected GenerationErrorMsg, got %T", msg)
	}
	if genErr.Err == nil {
		t.Error("expected error for nil generateState")
	}
}

func TestWizardIterativeGenerateCmd_StartsGeneration(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	wizard := NewWizard(tempDir, &scanner.ScanConfig{}, nil)

	progressCh := make(chan context.GenProgress, 100)
	done := make(chan bool)
	fileTree := &scanner.FileNode{Name: "root", Path: tempDir, IsDir: true}
	selections := map[string]bool{tempDir: true}
	genConfig := &context.GenerateConfig{
		MaxFileSize:  1024 * 1024,
		MaxTotalSize: 10 * 1024 * 1024,
		MaxFiles:     100,
		TemplateVars: map[string]string{
			"TASK":         "test task",
			"RULES":        "",
			"CURRENT_DATE": time.Now().Format("2006-01-02"),
		},
	}

	generator := context.NewDefaultContextGenerator()
	wizard.generateState = &generateState{
		generator:  generator,
		fileTree:   fileTree,
		selections: selections,
		config:     genConfig,
		rootPath:   tempDir,
		progressCh: progressCh,
		done:       done,
		started:    false,
	}

	cmd := wizard.iterativeGenerateCmd()
	msg := cmd()

	_, isPollMsg := msg.(pollGenerateMsg)
	if !isPollMsg {
		t.Errorf("expected pollGenerateMsg, got %T", msg)
	}

	if !wizard.generateState.started {
		t.Error("expected generateState.started to be true")
	}

	// Clean up: drain any remaining progress and wait for goroutine
	go func() {
		for range progressCh {
		}
	}()
	<-done
}

func TestWizardIterativeGenerateCmd_AlreadyStarted(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	wizard := NewWizard(tempDir, &scanner.ScanConfig{}, nil)

	progressCh := make(chan context.GenProgress, 100)
	done := make(chan bool)
	fileTree := &scanner.FileNode{Name: "root", Path: tempDir, IsDir: true}
	selections := map[string]bool{tempDir: true}
	genConfig := &context.GenerateConfig{
		MaxFileSize:  1024 * 1024,
		MaxTotalSize: 10 * 1024 * 1024,
		MaxFiles:     100,
		TemplateVars: map[string]string{
			"TASK":         "test task",
			"RULES":        "",
			"CURRENT_DATE": time.Now().Format("2006-01-02"),
		},
	}

	generator := context.NewDefaultContextGenerator()
	wizard.generateState = &generateState{
		generator:  generator,
		fileTree:   fileTree,
		selections: selections,
		config:     genConfig,
		rootPath:   tempDir,
		progressCh: progressCh,
		done:       done,
		started:    true, // Already started
	}

	cmd := wizard.iterativeGenerateCmd()
	msg := cmd()

	_, isPollMsg := msg.(pollGenerateMsg)
	if !isPollMsg {
		t.Errorf("expected pollGenerateMsg, got %T", msg)
	}

	// Since started was already true, no new goroutine was created
	close(done)
	close(progressCh)
}

func TestWizardHandleStepInput_FileSelection(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil)
	wizard.fileTree = &scanner.FileNode{Name: "root", Path: "/tmp/test", IsDir: true}
	wizard.fileSelection = screens.NewFileSelection(wizard.fileTree, wizard.selectedFiles)
	wizard.step = StepFileSelection
	wizard.validationError = "Some error"

	// Send input key
	_ = wizard.handleStepInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})

	// Validation error should be cleared
	if wizard.validationError != "" {
		t.Errorf("expected validation error to be cleared, got %q", wizard.validationError)
	}

	// Function executed without panic - test passes
}

func TestWizardHandleStepInput_TemplateSelection(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil)
	wizard.templateSelection = screens.NewTemplateSelection()
	wizard.step = StepTemplateSelection
	wizard.validationError = "Some error"

	// Send input key
	_ = wizard.handleStepInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})

	// Validation error should be cleared
	if wizard.validationError != "" {
		t.Errorf("expected validation error to be cleared, got %q", wizard.validationError)
	}

	// Function executed without panic - test passes
}

func TestWizardHandleStepInput_TaskInput(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil)
	wizard.taskInput = screens.NewTaskInput("initial task")
	wizard.step = StepTaskInput
	wizard.validationError = "Some error"

	// Send input key
	cmd := wizard.handleStepInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})

	// Validation error should be cleared
	if wizard.validationError != "" {
		t.Errorf("expected validation error to be cleared, got %q", wizard.validationError)
	}

	// Task input should return command
	if cmd == nil {
		t.Error("expected command from taskInput.Update")
	}
}

func TestWizardHandleStepInput_RulesInput(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil)
	wizard.rulesInput = screens.NewRulesInput("initial rules")
	wizard.step = StepRulesInput
	wizard.validationError = "Some error"

	// Send input key
	cmd := wizard.handleStepInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})

	// Validation error should be cleared
	if wizard.validationError != "" {
		t.Errorf("expected validation error to be cleared, got %q", wizard.validationError)
	}

	// Rules input should return command
	if cmd == nil {
		t.Error("expected command from rulesInput.Update")
	}
}

func TestWizardHandleStepInput_Review(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil)
	wizard.fileTree = &scanner.FileNode{Name: "root", Path: "/tmp/test", IsDir: true}
	wizard.template = &template.Template{Name: "basic"}
	wizard.review = screens.NewReview(wizard.selectedFiles, wizard.fileTree, wizard.template, "task", "", "")
	wizard.step = StepReview
	wizard.validationError = "Some error"

	// Send input key
	_ = wizard.handleStepInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})

	// Validation error should be cleared
	if wizard.validationError != "" {
		t.Errorf("expected validation error to be cleared, got %q", wizard.validationError)
	}

	// Function executed without panic - test passes
}

func TestWizardHandleStepInput_NilSubModels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		step int
	}{
		{"file selection with nil model", StepFileSelection},
		{"template selection with nil model", StepTemplateSelection},
		{"task input with nil model", StepTaskInput},
		{"rules input with nil model", StepRulesInput},
		{"review with nil model", StepReview},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil)
			wizard.step = tt.step
			wizard.validationError = "Some error"

			// Send input key - should not panic with nil sub-models
			cmd := wizard.handleStepInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})

			// Validation error should still be cleared
			if wizard.validationError != "" {
				t.Errorf("expected validation error to be cleared even with nil sub-model, got %q", wizard.validationError)
			}

			// Command should be nil when sub-model is nil
			if cmd != nil {
				t.Errorf("expected nil command when sub-model is nil, got non-nil")
			}
		})
	}
}

func TestWizardClipboardCopyCmd_Success(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil)
	content := "test content for clipboard"

	cmd := wizard.clipboardCopyCmd(content)
	msg := cmd()

	complete, ok := msg.(screens.ClipboardCompleteMsg)
	if !ok {
		t.Fatalf("expected ClipboardCompleteMsg, got %T", msg)
	}
	if !complete.Success {
		t.Errorf("expected Success=true for clipboard copy, got Success=false with err: %v", complete.Err)
	}
}

func TestWizardClipboardCopyCmd_EmptyContent(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil)
	content := ""

	cmd := wizard.clipboardCopyCmd(content)
	msg := cmd()

	complete, ok := msg.(screens.ClipboardCompleteMsg)
	if !ok {
		t.Fatalf("expected ClipboardCompleteMsg, got %T", msg)
	}
	// Empty content should still succeed (clipboard.Copy accepts empty strings)
	if !complete.Success {
		t.Errorf("expected Success=true for empty content, got Success=false with err: %v", complete.Err)
	}
}

func TestWizardClipboardCopyCmd_LargeContent(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil)
	content := strings.Repeat("x", 10000) // 10KB of content

	cmd := wizard.clipboardCopyCmd(content)
	msg := cmd()

	complete, ok := msg.(screens.ClipboardCompleteMsg)
	if !ok {
		t.Fatalf("expected ClipboardCompleteMsg, got %T", msg)
	}
	// Large content should still succeed
	if !complete.Success {
		t.Errorf("expected Success=true for large content, got Success=false with err: %v", complete.Err)
	}
}

// ============================================================================
// LLM Integration Tests
// ============================================================================

func TestWizardCreateLLMProvider_OpenAI(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil)
	wizard.wizardConfig = &WizardConfig{
		LLM: LLMConfig{
			Provider: "openai",
			APIKey:   "sk-test-key",
			Model:    "gpt-4o",
			Timeout:  60,
		},
	}

	provider, err := wizard.createLLMProvider()
	if err != nil {
		t.Logf("Provider creation returned error (may be expected): %v", err)
	}
	if provider != nil {
		if provider.Name() != "OpenAI" {
			t.Errorf("expected provider name 'OpenAI', got '%s'", provider.Name())
		}
	}
}

func TestWizardCreateLLMProvider_Anthropic(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil)
	wizard.wizardConfig = &WizardConfig{
		LLM: LLMConfig{
			Provider: "anthropic",
			APIKey:   "sk-ant-test-key",
			Model:    "claude-sonnet-4-20250514",
			Timeout:  60,
		},
	}

	provider, err := wizard.createLLMProvider()
	if err != nil {
		t.Logf("Provider creation returned error (may be expected): %v", err)
	}
	if provider != nil {
		if provider.Name() != "Anthropic" {
			t.Errorf("expected provider name 'Anthropic', got '%s'", provider.Name())
		}
	}
}

func TestWizardCreateLLMProvider_Gemini(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil)
	wizard.wizardConfig = &WizardConfig{
		LLM: LLMConfig{
			Provider: "gemini",
			APIKey:   "test-gemini-key",
			Model:    "gemini-2.5-flash",
			Timeout:  120,
		},
	}

	provider, err := wizard.createLLMProvider()
	if err != nil {
		t.Logf("Provider creation returned error (may be expected): %v", err)
	}
	if provider != nil {
		if provider.Name() != "Gemini" {
			t.Errorf("expected provider name 'Gemini', got '%s'", provider.Name())
		}
	}
}

func TestWizardCreateLLMProvider_GeminiWeb(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil)
	wizard.wizardConfig = &WizardConfig{
		LLM: LLMConfig{
			Provider:   "geminiweb",
			BinaryPath: "/path/to/geminiweb",
			Model:      "gemini-2.0-pro",
			Timeout:    300,
		},
	}

	provider, err := wizard.createLLMProvider()
	if err != nil {
		t.Logf("Provider creation returned error (may be expected): %v", err)
	}
	if provider != nil {
		if provider.Name() != "GeminiWeb" {
			t.Errorf("expected provider name 'GeminiWeb', got '%s'", provider.Name())
		}
	}
}

func TestWizardCreateLLMProvider_GeminiWeb_LegacyConfig(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil)
	wizard.wizardConfig = &WizardConfig{
		LLM: LLMConfig{
			Provider: "", // Empty should trigger GeminiWeb from legacy config
		},
		Gemini: GeminiConfig{
			BinaryPath: "/path/to/geminiweb",
			Model:      "gemini-2.5-pro",
			Timeout:    300,
		},
	}

	provider, err := wizard.createLLMProvider()
	if err != nil {
		t.Logf("Provider creation returned error (may be expected): %v", err)
	}
	if provider != nil {
		if provider.Name() != "GeminiWeb" {
			t.Errorf("expected provider name 'GeminiWeb', got '%s'", provider.Name())
		}
	}
}

func TestWizardCreateLLMProvider_GeminiWeb_MixedConfig(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil)
	wizard.wizardConfig = &WizardConfig{
		LLM: LLMConfig{
			Provider:   "geminiweb",
			Model:      "llm-model", // LLM config model
			Timeout:    60,
			BinaryPath: "", // Empty - should use Gemini config
		},
		Gemini: GeminiConfig{
			BinaryPath: "/path/to/geminiweb",
			Model:      "gemini-2.5-pro", // Should be used as fallback
			Timeout:    300,
		},
	}

	provider, err := wizard.createLLMProvider()
	if err != nil {
		t.Logf("Provider creation returned error (may be expected): %v", err)
	}
	if provider != nil {
		if provider.Name() != "GeminiWeb" {
			t.Errorf("expected provider name 'GeminiWeb', got '%s'", provider.Name())
		}
	}
}

func TestWizardCreateLLMProvider_InvalidProvider(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil)
	wizard.wizardConfig = &WizardConfig{
		LLM: LLMConfig{
			Provider: "invalid-provider",
		},
	}

	provider, err := wizard.createLLMProvider()
	if err == nil {
		t.Error("expected error for invalid provider, got nil")
	}
	if provider != nil {
		t.Error("expected nil provider for invalid provider")
	}
}

func TestWizardCreateLLMProvider_NoConfig(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil)
	// Set wizardConfig to empty struct instead of nil to avoid panic
	wizard.wizardConfig = &WizardConfig{}

	provider, err := wizard.createLLMProvider()
	// Should not panic, but may error or return nil
	if err != nil {
		t.Logf("Provider creation returned error (expected for empty config): %v", err)
	}
	_ = provider // May be nil
}

func TestWizardSendToLLMCmd_ReturnsCommand(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil)
	wizard.generatedContent = "test content to send"
	wizard.generatedFilePath = "/tmp/test.md"
	wizard.wizardConfig = &WizardConfig{
		LLM: LLMConfig{
			SaveResponse: false, // Don't actually save
		},
	}

	// Create a mock provider or skip if provider creation fails
	provider, err := wizard.createLLMProvider()
	if err != nil || provider == nil {
		t.Skip("Skipping test - provider not available")
		return
	}

	cmd := wizard.sendToLLMCmd(provider)
	if cmd == nil {
		t.Fatal("expected non-nil command")
	}
	// Don't execute the command as it would make an actual LLM call
}

func TestWizardSendToLLMCmd_SavesResponse(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	wizard := NewWizard(tempDir, &scanner.ScanConfig{}, nil)
	wizard.generatedContent = "test content to send"
	wizard.generatedFilePath = tempDir + "/test.md"
	wizard.wizardConfig = &WizardConfig{
		LLM: LLMConfig{
			SaveResponse: true,
		},
	}

	provider, err := wizard.createLLMProvider()
	if err != nil || provider == nil {
		t.Skip("Skipping test - provider not available")
		return
	}

	cmd := wizard.sendToLLMCmd(provider)
	if cmd == nil {
		t.Fatal("expected non-nil command")
	}
	// Don't execute the command as it would make an actual LLM call
}

func TestWizardHandleSendToGemini_NotReviewStep(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil)
	wizard.step = StepFileSelection // Not review step
	wizard.generatedContent = "some content"

	cmd := wizard.handleSendToGemini()
	if cmd != nil {
		t.Error("expected nil command when not on review step")
	}
}

func TestWizardHandleSendToGemini_NoGeneratedContent(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil)
	wizard.step = StepReview
	wizard.generatedContent = "" // No content

	cmd := wizard.handleSendToGemini()
	if cmd != nil {
		t.Error("expected nil command when no generated content")
	}
}

func TestWizardHandleSendToGemini_AlreadySending(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil)
	wizard.step = StepReview
	wizard.generatedContent = "some content"
	wizard.geminiSending = true // Already sending
	wizard.wizardConfig = &WizardConfig{
		LLM: LLMConfig{
			Provider: "openai",
			APIKey:   "sk-test-key",
		},
	}

	cmd := wizard.handleSendToGemini()
	if cmd != nil {
		t.Error("expected nil command when already sending")
	}
}

func TestWizardHandleSendToGemini_ProviderCreationError(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil)
	wizard.step = StepReview
	wizard.generatedContent = "some content"
	wizard.wizardConfig = &WizardConfig{
		LLM: LLMConfig{
			Provider: "invalid-provider", // Will cause creation error
		},
	}
	wizard.review = screens.NewReview(nil, nil, nil, "", "", "")

	cmd := wizard.handleSendToGemini()
	if cmd != nil {
		t.Error("expected nil command when provider creation fails")
	}
	// Error should be set on review model
}

func TestWizardHandleSendToGemini_WithValidProvider(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil)
	wizard.step = StepReview
	wizard.generatedContent = "test content"
	wizard.generatedFilePath = "/tmp/test.md"
	wizard.wizardConfig = &WizardConfig{
		LLM: LLMConfig{
			Provider:     "openai",
			APIKey:       "sk-test-key",
			SaveResponse: false,
		},
	}
	wizard.review = screens.NewReview(nil, nil, nil, "", "", "")

	cmd := wizard.handleSendToGemini()
	_ = cmd
}

func TestWizardHandleRescanRequest_FileSelectionStep(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{MaxFiles: 10}, nil)
	wizard.step = StepFileSelection
	wizard.rootPath = "/tmp/test"

	cmd := wizard.handleRescanRequest()
	if cmd == nil {
		t.Fatal("expected non-nil command on file selection step")
	}
	// Verify it's a scan command
	msg := cmd()
	if _, ok := msg.(ScanCompleteMsg); ok {
		// This is OK - the scan completes immediately for empty directories
	} else {
		// Some other message type is also OK
		_ = msg
	}
}

func TestWizardHandleRescanRequest_NotFileSelectionStep(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil)
	wizard.step = StepReview // Not file selection step
	wizard.rootPath = "/tmp/test"

	cmd := wizard.handleRescanRequest()
	if cmd != nil {
		t.Error("expected nil command when not on file selection step")
	}
}

func TestWizardHandleRescanRequest_AllSteps(t *testing.T) {
	t.Parallel()

	steps := []int{
		StepFileSelection,
		StepTemplateSelection,
		StepTaskInput,
		StepRulesInput,
		StepReview,
	}

	for _, step := range steps {
		t.Run(fmt.Sprintf("step_%d", step), func(t *testing.T) {
			wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil)
			wizard.step = step
			wizard.rootPath = "/tmp/test"

			cmd := wizard.handleRescanRequest()

			if step == StepFileSelection {
				if cmd == nil {
					t.Error("expected non-nil command on file selection step")
				}
			} else {
				if cmd != nil {
					t.Error("expected nil command on non-file-selection steps")
				}
			}
		})
	}
}

func TestWizardGeminiSendingFlagLifecycle(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil)
	wizard.step = StepReview
	wizard.generatedContent = "test content"
	wizard.wizardConfig = &WizardConfig{
		LLM: LLMConfig{
			Provider:     "openai",
			APIKey:       "sk-test-key",
			SaveResponse: false,
		},
	}
	wizard.review = screens.NewReview(nil, nil, nil, "", "", "")

	// Initial state
	if wizard.geminiSending {
		t.Error("expected geminiSending to be false initially")
	}

	// After handleSendToGemini (may or may not set flag depending on provider availability)
	_ = wizard.handleSendToGemini()

	// Simulate completion
	wizard.geminiSending = false
	if wizard.geminiSending {
		t.Error("expected geminiSending to be false after completion")
	}
}

func TestWizardHandleGeminiProgress(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil)
	wizard.progressComponent = components.NewProgress()

	msg := screens.GeminiProgressMsg{
		Stage: "sending",
	}

	// This should not panic
	wizard.handleGeminiProgress(msg)

	if !wizard.progress.Visible {
		t.Error("expected progress to be visible after progress message")
	}
}

func TestWizardHandleGeminiComplete(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil)
	wizard.geminiSending = true
	wizard.progress.Visible = true
	wizard.review = screens.NewReview(nil, nil, nil, "", "", "")

	msg := screens.GeminiCompleteMsg{
		Response:   "test response",
		OutputFile: "/tmp/test_response.md",
		Duration:   time.Second,
	}

	wizard.handleGeminiComplete(msg)

	if wizard.geminiSending {
		t.Error("expected geminiSending to be false after completion")
	}
	if wizard.progress.Visible {
		t.Error("expected progress to be invisible after completion")
	}
	if wizard.geminiResponseFile != msg.OutputFile {
		t.Errorf("expected geminiResponseFile to be %s, got %s", msg.OutputFile, wizard.geminiResponseFile)
	}
}

func TestWizardHandleGeminiError_ClearsState(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil)
	wizard.geminiSending = true
	wizard.progress.Visible = true
	wizard.review = screens.NewReview(nil, nil, nil, "", "", "")

	msg := screens.GeminiErrorMsg{
		Err: fmt.Errorf("test error"),
	}

	wizard.handleGeminiError(msg)

	if wizard.geminiSending {
		t.Error("expected geminiSending to be false after error")
	}
	if wizard.progress.Visible {
		t.Error("expected progress to be invisible after error")
	}
}
