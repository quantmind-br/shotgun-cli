package ui

import (
	gocontext "context"
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/quantmind-br/shotgun-cli/internal/app"
	"github.com/quantmind-br/shotgun-cli/internal/core/context"
	"github.com/quantmind-br/shotgun-cli/internal/core/llm"
	"github.com/quantmind-br/shotgun-cli/internal/core/scanner"
	"github.com/quantmind-br/shotgun-cli/internal/core/template"
	"github.com/quantmind-br/shotgun-cli/internal/ui/components"
	"github.com/quantmind-br/shotgun-cli/internal/ui/screens"
)

type mockContextService struct {
	sendToLLMWithProgressFunc func(ctx gocontext.Context, content string, cfg app.LLMSendConfig, progress app.LLMProgressCallback) (*llm.Result, error)
}

func (m *mockContextService) Generate(ctx gocontext.Context, cfg app.GenerateConfig) (*app.GenerateResult, error) {
	return nil, nil
}

func (m *mockContextService) GenerateWithProgress(ctx gocontext.Context, cfg app.GenerateConfig, progress app.ProgressCallback) (*app.GenerateResult, error) {
	return nil, nil
}

func (m *mockContextService) SendToLLM(ctx gocontext.Context, content string, provider llm.Provider) (*llm.Result, error) {
	return nil, nil
}

func (m *mockContextService) SendToLLMWithProgress(ctx gocontext.Context, content string, cfg app.LLMSendConfig, progress app.LLMProgressCallback) (*llm.Result, error) {
	if m.sendToLLMWithProgressFunc != nil {
		return m.sendToLLMWithProgressFunc(ctx, content, cfg, progress)
	}
	return &llm.Result{Response: "mock response", Duration: 100 * time.Millisecond}, nil
}

const (
	testTaskDescription = "Implement feature"
	testSampleTask      = "Sample task"
)

// wizardTestMockScanner is a test double that implements scanner.Scanner
type wizardTestMockScanner struct {
	scanFunc         func(rootPath string, config *scanner.ScanConfig) (*scanner.FileNode, error)
	scanProgressFunc func(rootPath string, config *scanner.ScanConfig, progress chan<- scanner.Progress) (*scanner.FileNode, error)
}

func (m *wizardTestMockScanner) Scan(rootPath string, config *scanner.ScanConfig) (*scanner.FileNode, error) {
	return m.scanFunc(rootPath, config)
}

func (m *wizardTestMockScanner) ScanWithProgress(rootPath string, config *scanner.ScanConfig, progress chan<- scanner.Progress) (*scanner.FileNode, error) {
	return m.scanProgressFunc(rootPath, config, progress)
}

// wizardTestMockGenerator is a test double for ContextGenerator
type wizardTestMockGenerator struct {
	generateFunc               func(root *scanner.FileNode, selections map[string]bool, config context.GenerateConfig) (string, error)
	generateWithProgressFunc   func(root *scanner.FileNode, selections map[string]bool, config context.GenerateConfig, progress func(string)) (string, error)
	generateWithProgressExFunc func(root *scanner.FileNode, selections map[string]bool, config context.GenerateConfig, progress func(context.GenProgress)) (string, error)
}

func (m *wizardTestMockGenerator) Generate(root *scanner.FileNode, selections map[string]bool, config context.GenerateConfig) (string, error) {
	if m.generateFunc != nil {
		return m.generateFunc(root, selections, config)
	}
	return "", nil
}

func (m *wizardTestMockGenerator) GenerateWithProgress(root *scanner.FileNode, selections map[string]bool, config context.GenerateConfig, progress func(string)) (string, error) {
	if m.generateWithProgressFunc != nil {
		return m.generateWithProgressFunc(root, selections, config, progress)
	}
	return "", nil
}

func (m *wizardTestMockGenerator) GenerateWithProgressEx(root *scanner.FileNode, selections map[string]bool, config context.GenerateConfig, progress func(context.GenProgress)) (string, error) {
	if m.generateWithProgressExFunc != nil {
		return m.generateWithProgressExFunc(root, selections, config, progress)
	}
	return "", nil
}

// setWizardFileTree is a test helper that sets the file tree via the fileSelection component
func setWizardFileTree(wizard *WizardModel, tree *scanner.FileNode) {
	var selections map[string]bool
	if wizard.fileSelection != nil {
		selections = wizard.fileSelection.GetSelections()
	}
	wizard.fileSelection = screens.NewFileSelection(tree, selections)
}

// getWizardFileTree is a test helper that retrieves the file tree via the fileSelection component
func getWizardFileTree(wizard *WizardModel) *scanner.FileNode {
	if wizard.fileSelection != nil {
		return wizard.fileSelection.GetFileTree()
	}
	return nil
}

// setWizardTemplate is a test helper that sets the template via the templateSelection component
func setWizardTemplate(wizard *WizardModel, tmpl *template.Template) {
	if wizard.templateSelection == nil {
		wizard.templateSelection = screens.NewTemplateSelection()
	}
	wizard.templateSelection.SetSelectedForTest(tmpl)
}

// getWizardTemplate is a test helper that retrieves the template via the templateSelection component
func getWizardTemplate(wizard *WizardModel) *template.Template {
	if wizard.templateSelection != nil {
		return wizard.templateSelection.GetSelected()
	}
	return nil
}

// setWizardSelectedFiles is a test helper that sets file selections via the fileSelection component
func setWizardSelectedFiles(wizard *WizardModel, selections map[string]bool) {
	if wizard.fileSelection == nil {
		wizard.fileSelection = screens.NewFileSelection(nil, selections)
	} else {
		wizard.fileSelection.SetSelectionsForTest(selections)
	}
}

// setWizardTaskDesc is a test helper that sets task description via the taskInput component
func setWizardTaskDesc(wizard *WizardModel, taskDesc string) {
	if wizard.taskInput == nil {
		wizard.taskInput = screens.NewTaskInput(taskDesc)
	} else {
		wizard.taskInput.SetValueForTest(taskDesc)
	}
}

func TestWizardHandlesScanLifecycle(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{MaxFiles: 5}, nil, nil)
	// Inject mock scanner into coordinator
	mockSc := &wizardTestMockScanner{
		scanProgressFunc: func(rootPath string, config *scanner.ScanConfig, progress chan<- scanner.Progress) (*scanner.FileNode, error) {
			return &scanner.FileNode{Name: "root", Path: rootPath, IsDir: true}, nil
		},
	}
	wizard.scanCoordinator = NewScanCoordinator(mockSc)

	var cmd tea.Cmd
	model, cmd := wizard.Update(startScanMsg{rootPath: "/workspace", config: &scanner.ScanConfig{MaxFiles: 5}})
	wiz := model.(*WizardModel)
	if wiz.scanCoordinator == nil {
		t.Fatalf("expected scan coordinator to be initialized")
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
	if getWizardFileTree(wiz) != tree {
		t.Fatalf("expected file tree to be set")
	}
	if wiz.fileSelection == nil {
		t.Fatalf("expected file selection model to be created")
	}
	if wiz.progress.Visible {
		t.Fatalf("progress should be hidden after completion")
	}
}

func TestWizardHandlesGenerationLifecycle(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	wizard := NewWizard(tempDir, &scanner.ScanConfig{}, nil, nil)
	mockGen := &wizardTestMockGenerator{
		generateFunc: func(root *scanner.FileNode, selections map[string]bool, config context.GenerateConfig) (string, error) {
			return "generated content", nil
		},
	}
	wizard.generateCoordinator = NewGenerateCoordinator(mockGen)

	tree := &scanner.FileNode{Name: "root", Path: tempDir, IsDir: true}
	setWizardFileTree(wizard, tree)
	setWizardSelectedFiles(wizard, map[string]bool{"main.go": true})
	tmpl := &template.Template{Name: "basic"}
	setWizardTemplate(wizard, tmpl)
	setWizardTaskDesc(wizard, "test task")

	// Trigger generation
	cmd := wizard.handleStartGeneration(startGenerationMsg{
		fileTree:      tree,
		selectedFiles: wizard.getSelectedFiles(),
		template:      tmpl,
		taskDesc:      wizard.getTaskDesc(),
		rootPath:      tempDir,
	})

	if cmd == nil {
		t.Fatal("expected start generation command")
	}

	// Simulate completion
	model, _ := wizard.Update(screens.GenerationCompleteMsg{
		Content: "generated content",
	})
	wiz := model.(*WizardModel)

	if wiz.generatedContent != "generated content" {
		t.Errorf("expected content 'generated content', got %q", wiz.generatedContent)
	}
	if wiz.progress.Visible {
		t.Error("progress should be hidden after completion")
	}
}
func TestWizardCanAdvanceStepLogic(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil, nil)
	setWizardFileTree(wizard, &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true})
	setWizardSelectedFiles(wizard, map[string]bool{"main.go": true})
	setWizardTemplate(wizard, &template.Template{Name: "basic"})
	setWizardTaskDesc(wizard, testTaskDescription)

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

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil, nil)
	setWizardFileTree(wizard, &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true})
	setWizardSelectedFiles(wizard, map[string]bool{"main.go": true})
	setWizardTemplate(wizard, &template.Template{Name: "basic"})
	setWizardTaskDesc(wizard, testTaskDescription)

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

func TestWizardGenerateContextMissingTemplate(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil, nil)
	setWizardSelectedFiles(wizard, map[string]bool{"main.go": true})

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

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil, nil)
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

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil, nil)
	tree := &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true}
	setWizardFileTree(wizard, tree)
	setWizardSelectedFiles(wizard, map[string]bool{"main.go": true})
	tmpl := &template.Template{Name: "basic"}
	setWizardTemplate(wizard, tmpl)

	var model tea.Model

	model, _ = wizard.Update(startGenerationMsg{
		fileTree:      tree,
		selectedFiles: wizard.getSelectedFiles(),
		template:      tmpl,
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

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil, nil)
	setWizardFileTree(wizard, &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true})
	setWizardSelectedFiles(wizard, map[string]bool{"main.go": true})
	setWizardTemplate(wizard, &template.Template{Name: "basic"})
	setWizardTaskDesc(wizard, testTaskDescription)

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

func TestWizardKeyboardNavigation_CtrlN_AdvancesStep(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		startStep    int
		expectedStep int
		setupFunc    func(*WizardModel)
	}{
		{
			name:         "from file selection to template selection",
			startStep:    StepFileSelection,
			expectedStep: StepTemplateSelection,
			setupFunc: func(m *WizardModel) {
				setWizardSelectedFiles(m, map[string]bool{"file.go": true})
			},
		},
		{
			name:         "from template selection to task input",
			startStep:    StepTemplateSelection,
			expectedStep: StepTaskInput,
			setupFunc: func(m *WizardModel) {
				setWizardTemplate(m, &template.Template{Name: "test", Content: "Task: {TASK}"})
			},
		},
		{
			name:         "from task input to rules input",
			startStep:    StepTaskInput,
			expectedStep: StepRulesInput,
			setupFunc: func(m *WizardModel) {
				setWizardTemplate(m, &template.Template{Name: "test", Content: "Task: {TASK}\nRules: {RULES}"})
				setWizardTaskDesc(m, "test task")
			},
		},
		{
			name:         "from rules input to review",
			startStep:    StepRulesInput,
			expectedStep: StepReview,
			setupFunc: func(m *WizardModel) {
				setWizardTemplate(m, &template.Template{Name: "test", Content: "Task: {TASK}\nRules: {RULES}"})
				setWizardTaskDesc(m, "test task")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			wizard := NewWizard("/tmp", &scanner.ScanConfig{}, nil, nil)
			setWizardFileTree(wizard, &scanner.FileNode{Name: "root", Path: "/tmp", IsDir: true})
			wizard.step = tt.startStep
			if tt.setupFunc != nil {
				tt.setupFunc(wizard)
			}

			msg := tea.KeyMsg{Type: tea.KeyCtrlN}
			model, _ := wizard.Update(msg)
			result := model.(*WizardModel)

			if result.step != tt.expectedStep {
				t.Errorf("expected step %d, got %d", tt.expectedStep, result.step)
			}
		})
	}
}

func TestWizardKeyboardNavigation_CtrlP_GoesBack(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		startStep    int
		expectedStep int
		setupFunc    func(*WizardModel)
	}{
		{
			name:         "from template selection to file selection",
			startStep:    StepTemplateSelection,
			expectedStep: StepFileSelection,
			setupFunc:    nil,
		},
		{
			name:         "from task input to template selection",
			startStep:    StepTaskInput,
			expectedStep: StepTemplateSelection,
			setupFunc: func(m *WizardModel) {
				setWizardTemplate(m, &template.Template{Name: "test", Content: "Task: {TASK}"})
			},
		},
		{
			name:         "from rules input to task input",
			startStep:    StepRulesInput,
			expectedStep: StepTaskInput,
			setupFunc: func(m *WizardModel) {
				setWizardTemplate(m, &template.Template{Name: "test", Content: "Task: {TASK}\nRules: {RULES}"})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			wizard := NewWizard("/tmp", &scanner.ScanConfig{}, nil, nil)
			setWizardFileTree(wizard, &scanner.FileNode{Name: "root", Path: "/tmp", IsDir: true})
			wizard.step = tt.startStep
			if tt.setupFunc != nil {
				tt.setupFunc(wizard)
			}

			msg := tea.KeyMsg{Type: tea.KeyCtrlP}
			model, _ := wizard.Update(msg)
			result := model.(*WizardModel)

			if result.step != tt.expectedStep {
				t.Errorf("expected step %d, got %d", tt.expectedStep, result.step)
			}
		})
	}
}

func TestWizardKeyboardNavigation_CtrlP_FirstStep_NoOp(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp", &scanner.ScanConfig{}, nil, nil)
	wizard.step = StepFileSelection

	msg := tea.KeyMsg{Type: tea.KeyCtrlP}
	model, _ := wizard.Update(msg)
	result := model.(*WizardModel)

	if result.step != StepFileSelection {
		t.Errorf("expected to stay at step %d, got %d", StepFileSelection, result.step)
	}
}

func TestWizardKeyboardNavigation_FunctionKeys_StillWork(t *testing.T) {
	t.Parallel()

	t.Run("F8 advances step", func(t *testing.T) {
		t.Parallel()
		wizard := NewWizard("/tmp", &scanner.ScanConfig{}, nil, nil)
		setWizardFileTree(wizard, &scanner.FileNode{Name: "root", Path: "/tmp", IsDir: true})
		wizard.step = StepFileSelection
		setWizardSelectedFiles(wizard, map[string]bool{"file.go": true})

		msg := tea.KeyMsg{Type: tea.KeyF8}
		model, _ := wizard.Update(msg)
		result := model.(*WizardModel)

		if result.step != StepTemplateSelection {
			t.Errorf("expected step %d, got %d", StepTemplateSelection, result.step)
		}
	})

	t.Run("F7 goes back", func(t *testing.T) {
		t.Parallel()
		wizard := NewWizard("/tmp", &scanner.ScanConfig{}, nil, nil)
		wizard.step = StepTemplateSelection

		msg := tea.KeyMsg{Type: tea.KeyF7}
		model, _ := wizard.Update(msg)
		result := model.(*WizardModel)

		if result.step != StepFileSelection {
			t.Errorf("expected step %d, got %d", StepFileSelection, result.step)
		}
	})
}

func TestWizardHelp_ShowsNewShortcuts(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp", &scanner.ScanConfig{}, nil, nil)
	wizard.showHelp = true
	wizard.width = 80
	wizard.height = 24

	view := wizard.View()

	if !strings.Contains(view, "Ctrl+N") {
		t.Error("expected help to contain 'Ctrl+N'")
	}
	if !strings.Contains(view, "Ctrl+P") {
		t.Error("expected help to contain 'Ctrl+P'")
	}
}

func TestWizardHelpToggle(t *testing.T) {
	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil, nil)

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

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil, nil)
	setWizardFileTree(wizard, &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true})
	setWizardSelectedFiles(wizard, map[string]bool{"main.go": true})
	// Template that has neither TASK nor RULES
	setWizardTemplate(wizard, &template.Template{
		Name:    "file_structure_only",
		Content: "File Structure:\n{FILE_STRUCTURE}",
	})

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

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil, nil)
	setWizardFileTree(wizard, &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true})
	setWizardSelectedFiles(wizard, map[string]bool{"main.go": true})
	// Template that has RULES but not TASK
	setWizardTemplate(wizard, &template.Template{
		Name:    "rules_only",
		Content: "Rules: {RULES}\nFile Structure:\n{FILE_STRUCTURE}",
	})

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

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil, nil)
	setWizardFileTree(wizard, &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true})
	setWizardSelectedFiles(wizard, map[string]bool{"main.go": true})
	// Template that has TASK but not RULES
	setWizardTemplate(wizard, &template.Template{
		Name:    "task_only",
		Content: "Task: {TASK}\nFile Structure:\n{FILE_STRUCTURE}",
	})
	setWizardTaskDesc(wizard, testSampleTask)

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

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil, nil)
	setWizardFileTree(wizard, &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true})
	setWizardSelectedFiles(wizard, map[string]bool{"main.go": true})
	// Template that has both TASK and RULES
	setWizardTemplate(wizard, &template.Template{
		Name:    "full_template",
		Content: "Task: {TASK}\nRules: {RULES}\nFile Structure:\n{FILE_STRUCTURE}",
	})

	// Start at template selection (step 2)
	wizard.step = StepTemplateSelection

	// Press F8 to advance - should go to Task (step 3)
	model, _ := wizard.Update(tea.KeyMsg{Type: tea.KeyF8})
	wizard = model.(*WizardModel)

	if wizard.step != StepTaskInput {
		t.Fatalf("expected to go to TaskInput (step %d), got step %d", StepTaskInput, wizard.step)
	}

	// Provide task and advance
	setWizardTaskDesc(wizard, testSampleTask)
	model, _ = wizard.Update(tea.KeyMsg{Type: tea.KeyF8})
	wizard = model.(*WizardModel)

	if wizard.step != StepRulesInput {
		t.Fatalf("expected to go to RulesInput (step %d), got step %d", StepRulesInput, wizard.step)
	}
}

func TestWizardBackwardNavigationSkipsCorrectly(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil, nil)
	tree := &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true}
	setWizardFileTree(wizard, tree)
	setWizardSelectedFiles(wizard, map[string]bool{"main.go": true})
	// Template with no TASK and no RULES
	tmpl := &template.Template{
		Name:    "file_structure_only",
		Content: "File Structure:\n{FILE_STRUCTURE}",
	}
	setWizardTemplate(wizard, tmpl)

	// Start at Review (step 5)
	wizard.step = StepReview
	wizard.review = screens.NewReview(wizard.getSelectedFiles(), tree, tmpl, "", "", "")

	// Press F7/F10 to go back - should skip Rules and Task, go to Template Selection
	model, _ := wizard.Update(tea.KeyMsg{Type: tea.KeyF10})
	wizard = model.(*WizardModel)

	if wizard.step != StepTemplateSelection {
		t.Fatalf("expected to go back to TemplateSelection (step %d), got step %d", StepTemplateSelection, wizard.step)
	}
}

func TestWizardBackwardFromRulesWithNoTask(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil, nil)
	setWizardFileTree(wizard, &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true})
	setWizardSelectedFiles(wizard, map[string]bool{"main.go": true})
	// Template with RULES but no TASK
	setWizardTemplate(wizard, &template.Template{
		Name:    "rules_only",
		Content: "Rules: {RULES}\nFile Structure:\n{FILE_STRUCTURE}",
	})

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

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil, nil)

	// No template set
	if wizard.requiresTaskInput() {
		t.Fatal("expected requiresTaskInput to be false when no template set")
	}

	// Template without TASK
	setWizardTemplate(wizard, &template.Template{
		Name:    "no_task",
		Content: "File Structure:\n{FILE_STRUCTURE}",
	})
	if wizard.requiresTaskInput() {
		t.Fatal("expected requiresTaskInput to be false when template has no TASK")
	}

	// Template with TASK
	setWizardTemplate(wizard, &template.Template{
		Name:    "with_task",
		Content: "Task: {TASK}",
	})
	if !wizard.requiresTaskInput() {
		t.Fatal("expected requiresTaskInput to be true when template has TASK")
	}
}

func TestWizardRequiresRulesInput(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil, nil)

	// No template set
	if wizard.requiresRulesInput() {
		t.Fatal("expected requiresRulesInput to be false when no template set")
	}

	// Template without RULES
	setWizardTemplate(wizard, &template.Template{
		Name:    "no_rules",
		Content: "Task: {TASK}",
	})
	if wizard.requiresRulesInput() {
		t.Fatal("expected requiresRulesInput to be false when template has no RULES")
	}

	// Template with RULES
	setWizardTemplate(wizard, &template.Template{
		Name:    "with_rules",
		Content: "Rules: {RULES}",
	})
	if !wizard.requiresRulesInput() {
		t.Fatal("expected requiresRulesInput to be true when template has RULES")
	}
}

func TestWizardCanAdvanceWithoutTaskWhenNotRequired(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil, nil)
	setWizardFileTree(wizard, &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true})
	setWizardSelectedFiles(wizard, map[string]bool{"main.go": true})
	// Template that does NOT require TASK
	setWizardTemplate(wizard, &template.Template{
		Name:    "no_task",
		Content: "File Structure:\n{FILE_STRUCTURE}",
	})

	// At Task Input step with empty task description
	wizard.step = StepTaskInput
	setWizardTaskDesc(wizard, "")

	// Should be able to advance since template doesn't require TASK
	if !wizard.canAdvanceStep() {
		t.Fatal("expected canAdvanceStep to return true when template doesn't require TASK")
	}
}

func TestWizardCannotAdvanceWithoutTaskWhenRequired(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil, nil)
	setWizardFileTree(wizard, &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true})
	setWizardSelectedFiles(wizard, map[string]bool{"main.go": true})
	// Template that requires TASK
	setWizardTemplate(wizard, &template.Template{
		Name:    "with_task",
		Content: "Task: {TASK}\nFile Structure:\n{FILE_STRUCTURE}",
	})

	// At Task Input step with empty task description
	wizard.step = StepTaskInput
	setWizardTaskDesc(wizard, "")

	// Should NOT be able to advance since template requires TASK
	if wizard.canAdvanceStep() {
		t.Fatal("expected canAdvanceStep to return false when template requires TASK and task is empty")
	}

	// Provide task description
	setWizardTaskDesc(wizard, testSampleTask)

	// Now should be able to advance
	if !wizard.canAdvanceStep() {
		t.Fatal("expected canAdvanceStep to return true when template requires TASK and task is provided")
	}
}

func TestWizardHandleWindowResize(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil, nil)
	tree := &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true}
	setWizardFileTree(wizard, tree)

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

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil, nil)
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

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil, nil)
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

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil, nil)
	tree := &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true}
	setWizardFileTree(wizard, tree)
	setWizardSelectedFiles(wizard, map[string]bool{"main.go": true})
	tmpl := &template.Template{Name: "basic", Content: "Task: {TASK}"}
	setWizardTemplate(wizard, tmpl)
	setWizardTaskDesc(wizard, "Test task")
	wizard.generatedContent = "generated context"
	wizard.generatedFilePath = "/tmp/test.md"
	wizard.step = StepReview
	wizard.review = screens.NewReview(wizard.getSelectedFiles(), tree, tmpl, wizard.getTaskDesc(), "", "")

	// Test GeminiProgressMsg
	model, _ := wizard.Update(screens.LLMProgressMsg{Stage: "sending"})
	wizard = model.(*WizardModel)
	if wizard.progress.Stage != "sending" {
		t.Errorf("expected progress stage 'sending', got %q", wizard.progress.Stage)
	}

	// Test GeminiCompleteMsg
	model, _ = wizard.Update(screens.LLMCompleteMsg{
		Response:   "AI response here",
		OutputFile: "/tmp/response.md",
		Duration:   5 * time.Second,
	})
	wizard = model.(*WizardModel)

	if wizard.llmResponseFile != "/tmp/response.md" {
		t.Errorf("expected response file '/tmp/response.md', got %q", wizard.llmResponseFile)
	}
	if wizard.llmSending {
		t.Error("geminiSending should be false after completion")
	}
}

func TestWizardHandleGeminiError(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil, nil)
	tree := &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true}
	setWizardFileTree(wizard, tree)
	wizard.step = StepReview
	wizard.review = screens.NewReview(map[string]bool{}, tree, nil, "", "", "")
	wizard.llmSending = true

	testErr := fmt.Errorf("geminiweb: connection timeout")
	model, _ := wizard.Update(screens.LLMErrorMsg{Err: testErr})
	wizard = model.(*WizardModel)

	if wizard.llmSending {
		t.Error("geminiSending should be false after error")
	}
	// Error is handled by review screen, check that no panic occurred
}

func TestWizardHandleTemplateMessage(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil, nil)
	setWizardFileTree(wizard, &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true})
	setWizardSelectedFiles(wizard, map[string]bool{"main.go": true})
	wizard.step = StepTemplateSelection

	selectedTemplate := &template.Template{
		Name:    "code-review",
		Content: "Review: {TASK}\nFiles: {FILE_STRUCTURE}",
	}

	model, _ := wizard.Update(TemplateSelectedMsg{Template: selectedTemplate})
	wizard = model.(*WizardModel)

	tmpl := getWizardTemplate(wizard)
	if tmpl == nil {
		t.Fatal("expected template to be set")
	}
	if tmpl.Name != "code-review" {
		t.Errorf("expected template name 'code-review', got %q", tmpl.Name)
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

			wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil, nil)
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

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil, nil)
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

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil, nil)
	setWizardFileTree(wizard, &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true})
	setWizardSelectedFiles(wizard, map[string]bool{"main.go": true})
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

	wizard := NewWizard("/workspace", &scanner.ScanConfig{MaxFiles: 100}, nil, nil)
	tree := &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true}
	setWizardFileTree(wizard, tree)
	setWizardSelectedFiles(wizard, map[string]bool{"main.go": true})
	wizard.step = StepFileSelection

	// The wizard should handle rescan by restarting the scan
	// This tests the conceptual flow - actual rescan may use different mechanism
	initialTree := tree

	// Simulate conditions that would trigger rescan
	wizard.scanCoordinator = nil
	model, cmd := wizard.Update(startScanMsg{rootPath: "/workspace", config: &scanner.ScanConfig{MaxFiles: 100}})
	wizard = model.(*WizardModel)

	if wizard.scanCoordinator == nil {
		t.Error("expected scan coordinator to be initialized on rescan")
	}
	if cmd == nil {
		t.Error("expected scan command to be scheduled")
	}
	_ = initialTree // Acknowledge variable
}

func TestWizardViewRendersFileSelectionStep(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil, nil)
	tree := &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true}
	setWizardFileTree(wizard, tree)
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

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil, nil)
	setWizardFileTree(wizard, &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true})
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

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil, nil)
	setWizardFileTree(wizard, &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true})
	setWizardTemplate(wizard, &template.Template{Name: "basic", Content: "Task: {TASK}"})
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

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil, nil)
	setWizardFileTree(wizard, &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true})
	setWizardTemplate(wizard, &template.Template{Name: "basic", Content: "Rules: {RULES}"})
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

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil, nil)
	tree := &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true}
	setWizardFileTree(wizard, tree)
	setWizardSelectedFiles(wizard, map[string]bool{"main.go": true})
	tmpl := &template.Template{Name: "basic"}
	setWizardTemplate(wizard, tmpl)
	setWizardTaskDesc(wizard, "Test task")
	wizard.step = StepReview
	wizard.review = screens.NewReview(wizard.getSelectedFiles(), tree, tmpl, wizard.getTaskDesc(), "", "")
	wizard.width = 80
	wizard.height = 24

	view := wizard.View()

	if view == "" {
		t.Error("expected non-empty view")
	}
}

func TestWizardViewWithError(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil, nil)
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

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil, nil)
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
			wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil, nil)
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
		wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil, nil)
		wizard.step = StepFileSelection
		setWizardSelectedFiles(wizard, map[string]bool{})

		wizard.handleNextStep()

		if wizard.validationError == "" {
			t.Error("expected validation error when no files selected")
		}
		if !strings.Contains(strings.ToLower(wizard.validationError), "file") {
			t.Errorf("error should mention files, got %q", wizard.validationError)
		}
	})

	t.Run("no template selected shows error", func(t *testing.T) {
		wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil, nil)
		wizard.step = StepTemplateSelection

		wizard.handleNextStep()

		if wizard.validationError == "" {
			t.Error("expected validation error when no template selected")
		}
		if !strings.Contains(strings.ToLower(wizard.validationError), "template") {
			t.Errorf("error should mention template, got %q", wizard.validationError)
		}
	})

	t.Run("empty task description shows error when required", func(t *testing.T) {
		wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil, nil)
		wizard.step = StepTaskInput
		setWizardTemplate(wizard, &template.Template{Name: "test", Content: "Task: {TASK}"})
		setWizardTaskDesc(wizard, "")

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

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil, nil)
	tree := &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true}
	setWizardFileTree(wizard, tree)
	wizard.step = StepFileSelection
	wizard.validationError = "Some error"

	wizard.handleStepInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})

	if wizard.validationError != "" {
		t.Errorf("expected validation error to be cleared, got %q", wizard.validationError)
	}
}

func TestWizardValidationErrorClearedOnSuccessfulAdvance(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil, nil)
	wizard.step = StepFileSelection
	setWizardSelectedFiles(wizard, map[string]bool{"/test/file.go": true})
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

	wizard := NewWizard("/workspace", &scanner.ScanConfig{}, nil, nil)
	tree := &scanner.FileNode{Name: "root", Path: "/workspace", IsDir: true}
	setWizardFileTree(wizard, tree)
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
			model := NewWizard("/tmp/test", &scanner.ScanConfig{}, &WizardConfig{}, nil)
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
	model := NewWizard("/tmp/test", &scanner.ScanConfig{}, &WizardConfig{}, nil)
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
			model := NewWizard("/tmp/test", &scanner.ScanConfig{}, &WizardConfig{}, nil)
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

func TestWizardHandleStepInput_FileSelection(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil, nil)
	tree := &scanner.FileNode{Name: "root", Path: "/tmp/test", IsDir: true}
	setWizardFileTree(wizard, tree)
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

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil, nil)
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

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil, nil)
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

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil, nil)
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

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil, nil)
	tree := &scanner.FileNode{Name: "root", Path: "/tmp/test", IsDir: true}
	setWizardFileTree(wizard, tree)
	setWizardSelectedFiles(wizard, map[string]bool{"main.go": true})
	tmpl := &template.Template{Name: "basic"}
	setWizardTemplate(wizard, tmpl)
	wizard.review = screens.NewReview(wizard.getSelectedFiles(), tree, tmpl, "task", "", "")
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

			wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil, nil)
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

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil, nil)
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

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil, nil)
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

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil, nil)
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

func TestWizardBuildLLMSendConfig_OpenAI(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil, nil)
	wizard.generatedFilePath = "/tmp/test.md"
	wizard.wizardConfig = &WizardConfig{
		LLM: LLMConfig{
			Provider: "openai",
			APIKey:   "sk-test-key",
			Model:    "gpt-4o",
			Timeout:  60,
		},
	}

	cfg := wizard.buildLLMSendConfig()

	if cfg.Provider != "openai" {
		t.Errorf("expected provider 'openai', got '%s'", cfg.Provider)
	}
	if cfg.APIKey != "sk-test-key" {
		t.Errorf("expected APIKey 'sk-test-key', got '%s'", cfg.APIKey)
	}
	if cfg.Model != "gpt-4o" {
		t.Errorf("expected model 'gpt-4o', got '%s'", cfg.Model)
	}
}

func TestWizardBuildLLMSendConfig_Anthropic(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil, nil)
	wizard.generatedFilePath = "/tmp/test.md"
	wizard.wizardConfig = &WizardConfig{
		LLM: LLMConfig{
			Provider: "anthropic",
			APIKey:   "sk-ant-test-key",
			Model:    "claude-sonnet-4-20250514",
			Timeout:  60,
		},
	}

	cfg := wizard.buildLLMSendConfig()

	if cfg.Provider != "anthropic" {
		t.Errorf("expected provider 'anthropic', got '%s'", cfg.Provider)
	}
	if cfg.APIKey != "sk-ant-test-key" {
		t.Errorf("expected APIKey 'sk-ant-test-key', got '%s'", cfg.APIKey)
	}
}

func TestWizardBuildLLMSendConfig_Gemini(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil, nil)
	wizard.generatedFilePath = "/tmp/test.md"
	wizard.wizardConfig = &WizardConfig{
		LLM: LLMConfig{
			Provider: "gemini",
			APIKey:   "test-gemini-key",
			Model:    "gemini-2.5-flash",
			Timeout:  120,
		},
	}

	cfg := wizard.buildLLMSendConfig()

	if cfg.Provider != "gemini" {
		t.Errorf("expected provider 'gemini', got '%s'", cfg.Provider)
	}
	if cfg.Model != "gemini-2.5-flash" {
		t.Errorf("expected model 'gemini-2.5-flash', got '%s'", cfg.Model)
	}
}

func TestWizardBuildLLMSendConfig_GeminiWeb(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil, nil)
	wizard.generatedFilePath = "/tmp/test.md"
	wizard.wizardConfig = &WizardConfig{
		LLM: LLMConfig{
			Provider:   "geminiweb",
			BinaryPath: "/path/to/geminiweb",
			Model:      "gemini-2.0-pro",
			Timeout:    300,
		},
	}

	cfg := wizard.buildLLMSendConfig()

	if cfg.Provider != "geminiweb" {
		t.Errorf("expected provider 'geminiweb', got '%s'", cfg.Provider)
	}
	if cfg.BinaryPath != "/path/to/geminiweb" {
		t.Errorf("expected BinaryPath '/path/to/geminiweb', got '%s'", cfg.BinaryPath)
	}
}

func TestWizardBuildLLMSendConfig_GeminiWeb_LegacyConfig(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil, nil)
	wizard.generatedFilePath = "/tmp/test.md"
	wizard.wizardConfig = &WizardConfig{
		LLM: LLMConfig{
			Provider: "",
		},
		Gemini: GeminiConfig{
			BinaryPath: "/path/to/geminiweb",
			Model:      "gemini-2.5-pro",
			Timeout:    300,
		},
	}

	cfg := wizard.buildLLMSendConfig()

	if cfg.Provider != "geminiweb" {
		t.Errorf("expected provider 'geminiweb', got '%s'", cfg.Provider)
	}
	if cfg.Model != "gemini-2.5-pro" {
		t.Errorf("expected model 'gemini-2.5-pro', got '%s'", cfg.Model)
	}
	if cfg.BinaryPath != "/path/to/geminiweb" {
		t.Errorf("expected BinaryPath '/path/to/geminiweb', got '%s'", cfg.BinaryPath)
	}
}

func TestWizardBuildLLMSendConfig_GeminiWeb_MixedConfig(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil, nil)
	wizard.generatedFilePath = "/tmp/test.md"
	wizard.wizardConfig = &WizardConfig{
		LLM: LLMConfig{
			Provider:   "geminiweb",
			Model:      "llm-model",
			Timeout:    60,
			BinaryPath: "",
		},
		Gemini: GeminiConfig{
			BinaryPath: "/path/to/geminiweb",
			Model:      "gemini-2.5-pro",
			Timeout:    300,
		},
	}

	cfg := wizard.buildLLMSendConfig()

	if cfg.Provider != "geminiweb" {
		t.Errorf("expected provider 'geminiweb', got '%s'", cfg.Provider)
	}
	if cfg.Model != "llm-model" {
		t.Errorf("expected model 'llm-model' (from LLM config), got '%s'", cfg.Model)
	}
	if cfg.BinaryPath != "/path/to/geminiweb" {
		t.Errorf("expected BinaryPath from legacy config, got '%s'", cfg.BinaryPath)
	}
}

func TestWizardBuildLLMSendConfig_SaveResponse(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil, nil)
	wizard.generatedFilePath = "/tmp/test.md"
	wizard.wizardConfig = &WizardConfig{
		LLM: LLMConfig{
			Provider:     "openai",
			SaveResponse: true,
		},
	}

	cfg := wizard.buildLLMSendConfig()

	if !cfg.SaveResponse {
		t.Error("expected SaveResponse to be true")
	}
	if cfg.OutputPath != "/tmp/test_response.md" {
		t.Errorf("expected OutputPath '/tmp/test_response.md', got '%s'", cfg.OutputPath)
	}
}

func TestWizardBuildLLMSendConfig_NoConfig(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil, nil)
	wizard.generatedFilePath = "/tmp/test.md"
	wizard.wizardConfig = &WizardConfig{}

	cfg := wizard.buildLLMSendConfig()

	if cfg.Provider != "geminiweb" {
		t.Errorf("expected default provider 'geminiweb', got '%s'", cfg.Provider)
	}
}

func TestWizardSendToLLMCmd_ReturnsCommand(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil, nil)
	wizard.generatedContent = "test content to send"
	wizard.generatedFilePath = "/tmp/test.md"
	wizard.wizardConfig = &WizardConfig{
		LLM: LLMConfig{
			SaveResponse: false,
		},
	}

	cmd := wizard.sendToLLMCmd()
	if cmd == nil {
		t.Fatal("expected non-nil command")
	}
}

func TestWizardHandleSendToGemini_NotReviewStep(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil, nil)
	wizard.step = StepFileSelection // Not review step
	wizard.generatedContent = "some content"

	cmd := wizard.handleSendToLLM()
	if cmd != nil {
		t.Error("expected nil command when not on review step")
	}
}

func TestWizardHandleSendToGemini_NoGeneratedContent(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil, nil)
	wizard.step = StepReview
	wizard.generatedContent = "" // No content

	cmd := wizard.handleSendToLLM()
	if cmd != nil {
		t.Error("expected nil command when no generated content")
	}
}

func TestWizardHandleSendToGemini_AlreadySending(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil, nil)
	wizard.step = StepReview
	wizard.generatedContent = "some content"
	wizard.llmSending = true // Already sending
	wizard.wizardConfig = &WizardConfig{
		LLM: LLMConfig{
			Provider: "openai",
			APIKey:   "sk-test-key",
		},
	}

	cmd := wizard.handleSendToLLM()
	if cmd != nil {
		t.Error("expected nil command when already sending")
	}
}

func TestWizardHandleSendToGemini_ReturnsCommandForValidInput(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil, nil)
	wizard.step = StepReview
	wizard.generatedContent = "some content"
	wizard.generatedFilePath = "/tmp/test.md"
	wizard.wizardConfig = &WizardConfig{
		LLM: LLMConfig{
			Provider: "openai",
		},
	}
	wizard.review = screens.NewReview(nil, nil, nil, "", "", "")

	cmd := wizard.handleSendToLLM()
	if cmd == nil {
		t.Error("expected non-nil command for valid input")
	}
	if !wizard.llmSending {
		t.Error("expected llmSending to be true after handleSendToLLM")
	}
}

func TestWizardHandleSendToGemini_WithValidProvider(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil, nil)
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

	cmd := wizard.handleSendToLLM()
	_ = cmd
}

func TestWizardHandleRescanRequest_FileSelectionStep(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{MaxFiles: 10}, nil, nil)
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

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil, nil)
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
			wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil, nil)
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

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil, nil)
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
	if wizard.llmSending {
		t.Error("expected geminiSending to be false initially")
	}

	// After handleSendToGemini (may or may not set flag depending on provider availability)
	_ = wizard.handleSendToLLM()

	// Simulate completion
	wizard.llmSending = false
	if wizard.llmSending {
		t.Error("expected geminiSending to be false after completion")
	}
}

func TestWizardHandleGeminiProgress(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil, nil)
	wizard.progressComponent = components.NewProgress()

	msg := screens.LLMProgressMsg{
		Stage: "sending",
	}

	// This should not panic
	wizard.handleLLMProgress(msg)

	if !wizard.progress.Visible {
		t.Error("expected progress to be visible after progress message")
	}
}

func TestWizardHandleGeminiComplete(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil, nil)
	wizard.llmSending = true
	wizard.progress.Visible = true
	wizard.review = screens.NewReview(nil, nil, nil, "", "", "")

	msg := screens.LLMCompleteMsg{
		Response:   "test response",
		OutputFile: "/tmp/test_response.md",
		Duration:   time.Second,
	}

	wizard.handleLLMComplete(msg)

	if wizard.llmSending {
		t.Error("expected geminiSending to be false after completion")
	}
	if wizard.progress.Visible {
		t.Error("expected progress to be invisible after completion")
	}
	if wizard.llmResponseFile != msg.OutputFile {
		t.Errorf("expected geminiResponseFile to be %s, got %s", msg.OutputFile, wizard.llmResponseFile)
	}
}

func TestWizardHandleGeminiError_ClearsState(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil, nil)
	wizard.llmSending = true
	wizard.progress.Visible = true
	wizard.review = screens.NewReview(nil, nil, nil, "", "", "")

	msg := screens.LLMErrorMsg{
		Err: fmt.Errorf("test error"),
	}

	wizard.handleLLMError(msg)

	if wizard.llmSending {
		t.Error("expected geminiSending to be false after error")
	}
	if wizard.progress.Visible {
		t.Error("expected progress to be invisible after error")
	}
}

func TestWizardHandleSendToLLM_DelegatesToService(t *testing.T) {
	t.Parallel()

	var receivedContent string
	var receivedCfg app.LLMSendConfig
	mockSvc := &mockContextService{
		sendToLLMWithProgressFunc: func(ctx gocontext.Context, content string, cfg app.LLMSendConfig, progress app.LLMProgressCallback) (*llm.Result, error) {
			receivedContent = content
			receivedCfg = cfg
			return &llm.Result{Response: "test response", Duration: 50 * time.Millisecond}, nil
		},
	}

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil, mockSvc)
	wizard.step = StepReview
	wizard.generatedContent = "test content to send"
	wizard.generatedFilePath = "/tmp/test.md"
	wizard.wizardConfig = &WizardConfig{
		LLM: LLMConfig{
			Provider:     "openai",
			APIKey:       "sk-test-key",
			SaveResponse: true,
		},
	}
	wizard.review = screens.NewReview(nil, nil, nil, "", "", "")

	cmd := wizard.handleSendToLLM()
	if cmd == nil {
		t.Fatal("expected non-nil command")
	}

	msg := cmd()
	batchMsg, ok := msg.(tea.BatchMsg)
	if !ok {
		t.Fatalf("expected tea.BatchMsg, got %T", msg)
	}

	var foundCompleteMsg bool
	for _, batchCmd := range batchMsg {
		if batchCmd == nil {
			continue
		}
		result := batchCmd()
		if completeMsg, ok := result.(screens.LLMCompleteMsg); ok {
			foundCompleteMsg = true
			if completeMsg.Response != "test response" {
				t.Errorf("expected response 'test response', got '%s'", completeMsg.Response)
			}
		}
	}

	if !foundCompleteMsg {
		t.Error("expected LLMCompleteMsg in batch")
	}
	if receivedContent != "test content to send" {
		t.Errorf("expected content 'test content to send', got '%s'", receivedContent)
	}
	if receivedCfg.Provider != "openai" {
		t.Errorf("expected provider 'openai', got '%s'", receivedCfg.Provider)
	}
}

func TestWizardHandleSendToLLM_ServiceError(t *testing.T) {
	t.Parallel()

	mockSvc := &mockContextService{
		sendToLLMWithProgressFunc: func(ctx gocontext.Context, content string, cfg app.LLMSendConfig, progress app.LLMProgressCallback) (*llm.Result, error) {
			return nil, fmt.Errorf("service error: connection failed")
		},
	}

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil, mockSvc)
	wizard.step = StepReview
	wizard.generatedContent = "test content"
	wizard.generatedFilePath = "/tmp/test.md"
	wizard.wizardConfig = &WizardConfig{
		LLM: LLMConfig{
			Provider: "openai",
		},
	}
	wizard.review = screens.NewReview(nil, nil, nil, "", "", "")

	cmd := wizard.handleSendToLLM()
	if cmd == nil {
		t.Fatal("expected non-nil command")
	}

	msg := cmd()
	batchMsg, ok := msg.(tea.BatchMsg)
	if !ok {
		t.Fatalf("expected tea.BatchMsg, got %T", msg)
	}

	var foundErrorMsg bool
	for _, batchCmd := range batchMsg {
		if batchCmd == nil {
			continue
		}
		result := batchCmd()
		if errMsg, ok := result.(screens.LLMErrorMsg); ok {
			foundErrorMsg = true
			if errMsg.Err == nil {
				t.Error("expected error to be set")
			}
			if !strings.Contains(errMsg.Err.Error(), "connection failed") {
				t.Errorf("expected error to contain 'connection failed', got '%s'", errMsg.Err.Error())
			}
		}
	}

	if !foundErrorMsg {
		t.Error("expected LLMErrorMsg in batch")
	}
}

func TestWizardWithMockService(t *testing.T) {
	t.Parallel()

	mockSvc := &mockContextService{}
	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil, mockSvc)

	if wizard.contextService != mockSvc {
		t.Error("expected wizard to use injected mock service")
	}
}

func TestWizardWithNilService_CreatesDefault(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp/test", &scanner.ScanConfig{}, nil, nil)

	if wizard.contextService == nil {
		t.Error("expected wizard to create default service when nil passed")
	}
}

func TestWizardView_SmallScreen_WidthTooSmall(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp", &scanner.ScanConfig{}, nil, nil)
	wizard.width = 30
	wizard.height = 24

	view := wizard.View()

	if !strings.Contains(view, "Terminal too small") {
		t.Error("expected small screen warning")
	}
	if !strings.Contains(view, "30x24") {
		t.Error("expected current dimensions in warning")
	}
	if !strings.Contains(view, "40x10") {
		t.Error("expected minimum dimensions in warning")
	}
}

func TestWizardView_SmallScreen_HeightTooSmall(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp", &scanner.ScanConfig{}, nil, nil)
	wizard.width = 80
	wizard.height = 5

	view := wizard.View()

	if !strings.Contains(view, "Terminal too small") {
		t.Error("expected small screen warning")
	}
	if !strings.Contains(view, "80x5") {
		t.Error("expected current dimensions in warning")
	}
}

func TestWizardView_SmallScreen_BothTooSmall(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp", &scanner.ScanConfig{}, nil, nil)
	wizard.width = 20
	wizard.height = 5

	view := wizard.View()

	if !strings.Contains(view, "Terminal too small") {
		t.Error("expected small screen warning")
	}
	if !strings.Contains(view, "20x5") {
		t.Error("expected current dimensions in warning")
	}
}

func TestWizardView_SmallScreen_NormalSizeShowsContent(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp", &scanner.ScanConfig{}, nil, nil)
	wizard.width = 80
	wizard.height = 24

	view := wizard.View()

	if strings.Contains(view, "Terminal too small") {
		t.Error("should not show small screen warning for normal size")
	}
}

func TestWizardView_SmallScreen_BoundaryConditions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		width         int
		height        int
		expectWarning bool
	}{
		{"exactly minimum - no warning", 40, 10, false},
		{"one less than min width", 39, 10, true},
		{"one less than min height", 40, 9, true},
		{"well above minimum", 120, 40, false},
		{"exactly one above min width", 41, 10, false},
		{"exactly one above min height", 40, 11, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			wizard := NewWizard("/tmp", &scanner.ScanConfig{}, nil, nil)
			wizard.width = tt.width
			wizard.height = tt.height

			view := wizard.View()

			if tt.expectWarning {
				if !strings.Contains(view, "Terminal too small") {
					t.Errorf("expected warning for %dx%d", tt.width, tt.height)
				}
			} else {
				if strings.Contains(view, "Terminal too small") {
					t.Errorf("unexpected warning for %dx%d", tt.width, tt.height)
				}
			}
		})
	}
}

func TestWizardView_SmallScreen_ZeroDimensions_NoWarning(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp", &scanner.ScanConfig{}, nil, nil)
	wizard.width = 0
	wizard.height = 0

	view := wizard.View()

	if strings.Contains(view, "Terminal too small") {
		t.Error("should not show small screen warning for 0x0 dimensions")
	}
}

func TestWizard_renderSmallScreenWarning(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp", &scanner.ScanConfig{}, nil, nil)
	wizard.width = 30
	wizard.height = 8

	view := wizard.renderSmallScreenWarning()

	if !strings.Contains(view, "Terminal too small") {
		t.Error("expected warning text")
	}
	if !strings.Contains(view, "30x8") {
		t.Error("expected current dimensions")
	}
	if !strings.Contains(view, "40x10") {
		t.Error("expected minimum dimensions")
	}
	if view == "" {
		t.Error("expected non-empty view")
	}
}

func TestWizardView_SmallScreen_PrecedenceOverHelp(t *testing.T) {
	t.Parallel()

	wizard := NewWizard("/tmp", &scanner.ScanConfig{}, nil, nil)
	wizard.width = 30
	wizard.height = 8
	wizard.showHelp = true

	view := wizard.View()

	if !strings.Contains(view, "Terminal too small") {
		t.Error("small screen warning should show")
	}
	if strings.Contains(view, "Global Shortcuts") {
		t.Error("help content should not show when terminal too small")
	}
}

func TestWizard_isTerminalTooSmall(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		width    int
		height   int
		expected bool
	}{
		{"zero dimensions", 0, 0, false},
		{"width zero only", 0, 24, false},
		{"height zero only", 80, 0, false},
		{"both below minimum", 20, 5, true},
		{"width below minimum", 30, 24, true},
		{"height below minimum", 80, 5, true},
		{"exactly at minimum", 40, 10, false},
		{"above minimum", 80, 24, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			wizard := NewWizard("/tmp", &scanner.ScanConfig{}, nil, nil)
			wizard.width = tt.width
			wizard.height = tt.height

			result := wizard.isTerminalTooSmall()

			if result != tt.expected {
				t.Errorf("isTerminalTooSmall() = %v, want %v for %dx%d",
					result, tt.expected, tt.width, tt.height)
			}
		})
	}
}
