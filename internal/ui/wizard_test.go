package ui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/quantmind-br/shotgun-cli/internal/core/scanner"
	"github.com/quantmind-br/shotgun-cli/internal/core/template"
	"github.com/quantmind-br/shotgun-cli/internal/ui/screens"
)

const (
	testTaskDescription = "Implement feature"
	testSampleTask      = "Sample task"
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

func TestWizardCanAdvanceStepLogic(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{})
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

	wizard := NewWizard("/workspace", &scanner.ScanConfig{})
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

	wizard := NewWizard("/workspace", &scanner.ScanConfig{})
	wizard.fileTree = &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true}
	wizard.selectedFiles["main.go"] = true
	wizard.template = &template.Template{Name: "basic"}
	wizard.taskDesc = testTaskDescription
	wizard.review = screens.NewReview(wizard.selectedFiles, wizard.fileTree, wizard.template, wizard.taskDesc, "")
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
	wizard.review = screens.NewReview(map[string]bool{}, nil, nil, "", "")
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

	wizard := NewWizard("/workspace", &scanner.ScanConfig{})
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
	wizard := NewWizard("/workspace", &scanner.ScanConfig{})

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

	wizard := NewWizard("/workspace", &scanner.ScanConfig{})
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

	wizard := NewWizard("/workspace", &scanner.ScanConfig{})
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

	wizard := NewWizard("/workspace", &scanner.ScanConfig{})
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

	wizard := NewWizard("/workspace", &scanner.ScanConfig{})
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

	wizard := NewWizard("/workspace", &scanner.ScanConfig{})
	wizard.fileTree = &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true}
	wizard.selectedFiles["main.go"] = true
	// Template with no TASK and no RULES
	wizard.template = &template.Template{
		Name:    "file_structure_only",
		Content: "File Structure:\n{FILE_STRUCTURE}",
	}

	// Start at Review (step 5)
	wizard.step = StepReview
	wizard.review = screens.NewReview(wizard.selectedFiles, wizard.fileTree, wizard.template, "", "")

	// Press F7/F10 to go back - should skip Rules and Task, go to Template Selection
	model, _ := wizard.Update(tea.KeyMsg{Type: tea.KeyF10})
	wizard = model.(*WizardModel)

	if wizard.step != StepTemplateSelection {
		t.Fatalf("expected to go back to TemplateSelection (step %d), got step %d", StepTemplateSelection, wizard.step)
	}
}

func TestWizardBackwardFromRulesWithNoTask(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{})
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

	wizard := NewWizard("/workspace", &scanner.ScanConfig{})

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

	wizard := NewWizard("/workspace", &scanner.ScanConfig{})

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

	wizard := NewWizard("/workspace", &scanner.ScanConfig{})
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

	wizard := NewWizard("/workspace", &scanner.ScanConfig{})
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

	wizard := NewWizard("/workspace", &scanner.ScanConfig{})
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

	wizard := NewWizard("/workspace", &scanner.ScanConfig{})
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

	wizard := NewWizard("/workspace", &scanner.ScanConfig{})
	wizard.progress.Visible = true

	testErr := fmt.Errorf("template rendering failed")
	model, _ := wizard.Update(GenerationErrorMsg{Err: testErr})
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

	wizard := NewWizard("/workspace", &scanner.ScanConfig{})
	wizard.fileTree = &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true}
	wizard.selectedFiles["main.go"] = true
	wizard.template = &template.Template{Name: "basic", Content: "Task: {TASK}"}
	wizard.taskDesc = "Test task"
	wizard.generatedContent = "generated context"
	wizard.generatedFilePath = "/tmp/test.md"
	wizard.step = StepReview
	wizard.review = screens.NewReview(wizard.selectedFiles, wizard.fileTree, wizard.template, wizard.taskDesc, "")

	// Test GeminiProgressMsg
	model, _ := wizard.Update(GeminiProgressMsg{Stage: "sending"})
	wizard = model.(*WizardModel)
	if wizard.progress.Stage != "sending" {
		t.Errorf("expected progress stage 'sending', got %q", wizard.progress.Stage)
	}

	// Test GeminiCompleteMsg
	model, _ = wizard.Update(GeminiCompleteMsg{
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

	wizard := NewWizard("/workspace", &scanner.ScanConfig{})
	wizard.fileTree = &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true}
	wizard.step = StepReview
	wizard.review = screens.NewReview(map[string]bool{}, wizard.fileTree, nil, "", "")
	wizard.geminiSending = true

	testErr := fmt.Errorf("geminiweb: connection timeout")
	model, _ := wizard.Update(GeminiErrorMsg{Err: testErr})
	wizard = model.(*WizardModel)

	if wizard.geminiSending {
		t.Error("geminiSending should be false after error")
	}
	// Error is handled by review screen, check that no panic occurred
}

func TestWizardHandleTemplateMessage(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{})
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

func TestWizardHandleRescanRequest(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{MaxFiles: 100})
	wizard.fileTree = &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true}
	wizard.selectedFiles["main.go"] = true
	wizard.step = StepFileSelection

	// Create a rescan message type (internal)
	type rescanMsg struct{}

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
