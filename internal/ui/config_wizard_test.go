package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/quantmind-br/shotgun-cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfigWizard(t *testing.T) {
	t.Parallel()

	wizard := NewConfigWizard()

	require.NotNil(t, wizard)
	assert.Equal(t, 0, wizard.ActiveCategory())
	assert.False(t, wizard.ShowingHelp())
	assert.False(t, wizard.ConfirmingQuit())
	assert.Empty(t, wizard.SavedMessage())
	assert.Empty(t, wizard.ErrorMessage())
	assert.Len(t, wizard.categories, 6)
	assert.Len(t, wizard.categoryScreens, 6)
}

func TestConfigWizard_Init(t *testing.T) {
	t.Parallel()

	wizard := NewConfigWizard()
	cmd := wizard.Init()
	assert.Nil(t, cmd)
}

func TestConfigWizard_HandleWindowSize(t *testing.T) {
	t.Parallel()

	wizard := NewConfigWizard()
	model, cmd := wizard.Update(tea.WindowSizeMsg{Width: 100, Height: 50})

	assert.Nil(t, cmd)
	require.NotNil(t, model)

	w := model.(*ConfigWizardModel)
	assert.Equal(t, 100, w.width)
	assert.Equal(t, 50, w.height)
}

func TestConfigWizard_HandleWindowSize_Small(t *testing.T) {
	t.Parallel()

	wizard := NewConfigWizard()
	model, _ := wizard.Update(tea.WindowSizeMsg{Width: 20, Height: 10})

	require.NotNil(t, model)
}

func TestConfigWizard_CategoryNavigation(t *testing.T) {
	t.Parallel()

	wizard := NewConfigWizard()
	wizard.Update(tea.WindowSizeMsg{Width: 100, Height: 50})

	assert.Equal(t, 0, wizard.ActiveCategory())

	wizard.Update(tea.KeyMsg{Type: tea.KeyTab})
	assert.Equal(t, 1, wizard.ActiveCategory())

	wizard.Update(tea.KeyMsg{Type: tea.KeyTab})
	assert.Equal(t, 2, wizard.ActiveCategory())

	wizard.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	assert.Equal(t, 1, wizard.ActiveCategory())

	wizard.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	assert.Equal(t, 0, wizard.ActiveCategory())
}

func TestConfigWizard_CategoryNavigation_Wrap(t *testing.T) {
	t.Parallel()

	wizard := NewConfigWizard()
	wizard.Update(tea.WindowSizeMsg{Width: 100, Height: 50})

	wizard.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	assert.Equal(t, 5, wizard.ActiveCategory())

	wizard.Update(tea.KeyMsg{Type: tea.KeyTab})
	assert.Equal(t, 0, wizard.ActiveCategory())

	wizard.activeCategory = 5
	wizard.Update(tea.KeyMsg{Type: tea.KeyTab})
	assert.Equal(t, 0, wizard.ActiveCategory())
}

func TestConfigWizard_HelpToggle(t *testing.T) {
	t.Parallel()

	wizard := NewConfigWizard()
	wizard.Update(tea.WindowSizeMsg{Width: 100, Height: 50})

	assert.False(t, wizard.ShowingHelp())

	wizard.Update(tea.KeyMsg{Type: tea.KeyF1})
	assert.True(t, wizard.ShowingHelp())

	wizard.Update(tea.KeyMsg{Type: tea.KeyF1})
	assert.False(t, wizard.ShowingHelp())
}

func TestConfigWizard_HelpCloseWithEsc(t *testing.T) {
	t.Parallel()

	wizard := NewConfigWizard()
	wizard.showHelp = true

	wizard.Update(tea.KeyMsg{Type: tea.KeyEsc})
	assert.False(t, wizard.ShowingHelp())
}

func TestConfigWizard_HelpCloseWithQ(t *testing.T) {
	t.Parallel()

	wizard := NewConfigWizard()
	wizard.showHelp = true

	wizard.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	assert.False(t, wizard.ShowingHelp())
}

func TestConfigWizard_ConfirmQuit_Cancel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		key  string
	}{
		{"n cancels", "n"},
		{"N cancels", "N"},
		{"esc cancels", "esc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			wizard := NewConfigWizard()
			wizard.confirmQuit = true

			if tt.key == "esc" {
				wizard.Update(tea.KeyMsg{Type: tea.KeyEsc})
			} else {
				wizard.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)})
			}

			assert.False(t, wizard.ConfirmingQuit())
		})
	}
}

func TestConfigWizard_QuitWithoutChanges(t *testing.T) {
	t.Parallel()

	wizard := NewConfigWizard()
	wizard.Update(tea.WindowSizeMsg{Width: 100, Height: 50})

	_, cmd := wizard.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	assert.NotNil(t, cmd)
}

func TestConfigWizard_View_Main(t *testing.T) {
	t.Parallel()

	wizard := NewConfigWizard()
	wizard.Update(tea.WindowSizeMsg{Width: 100, Height: 50})

	view := wizard.View()
	assert.NotEmpty(t, view)
	assert.Contains(t, view, "Configuration Settings")
	assert.Contains(t, view, "Categories")
	assert.Contains(t, view, string(config.CategoryScanner))
}

func TestConfigWizard_View_Help(t *testing.T) {
	t.Parallel()

	wizard := NewConfigWizard()
	wizard.showHelp = true

	view := wizard.View()
	assert.NotEmpty(t, view)
	assert.Contains(t, view, "Help")
	assert.Contains(t, view, "Navigation")
	assert.Contains(t, view, "Editing")
}

func TestConfigWizard_View_ConfirmQuit(t *testing.T) {
	t.Parallel()

	wizard := NewConfigWizard()
	wizard.confirmQuit = true

	view := wizard.View()
	assert.NotEmpty(t, view)
	assert.Contains(t, view, "Unsaved Changes")
	assert.Contains(t, view, "[S]")
	assert.Contains(t, view, "[Y]")
	assert.Contains(t, view, "[N]")
}

func TestConfigWizard_View_SavedMessage(t *testing.T) {
	t.Parallel()

	wizard := NewConfigWizard()
	wizard.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	wizard.savedMessage = "Test saved message"

	view := wizard.View()
	assert.Contains(t, view, "Test saved message")
}

func TestConfigWizard_View_ErrorMessage(t *testing.T) {
	t.Parallel()

	wizard := NewConfigWizard()
	wizard.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	wizard.errorMessage = "Test error message"

	view := wizard.View()
	assert.Contains(t, view, "Test error message")
}

func TestConfigWizard_ConfigSavedMsg(t *testing.T) {
	t.Parallel()

	wizard := NewConfigWizard()
	model, cmd := wizard.Update(ConfigSavedMsg{})

	w := model.(*ConfigWizardModel)
	assert.Equal(t, "Configuration saved successfully!", w.SavedMessage())
	assert.Empty(t, w.ErrorMessage())
	assert.NotNil(t, cmd)
}

func TestConfigWizard_ConfigSavedMsg_QuitAfterSave(t *testing.T) {
	t.Parallel()

	wizard := NewConfigWizard()
	wizard.quitAfterSave = true
	_, cmd := wizard.Update(ConfigSavedMsg{})

	assert.NotNil(t, cmd)
	msg := cmd()
	_, isQuit := msg.(tea.QuitMsg)
	assert.True(t, isQuit, "should return tea.QuitMsg when quitAfterSave is true")
}

func TestConfigWizard_ConfigSaveErrorMsg(t *testing.T) {
	t.Parallel()

	wizard := NewConfigWizard()
	model, cmd := wizard.Update(ConfigSaveErrorMsg{Err: assert.AnError})

	w := model.(*ConfigWizardModel)
	assert.Contains(t, w.ErrorMessage(), "Error saving")
	assert.Empty(t, w.SavedMessage())
	assert.NotNil(t, cmd)
}

func TestConfigWizard_ClearSavedMsgTick(t *testing.T) {
	t.Parallel()

	wizard := NewConfigWizard()
	wizard.savedMessage = "Test"
	wizard.errorMessage = "Error"

	model, cmd := wizard.Update(clearSavedMsgTick{})

	w := model.(*ConfigWizardModel)
	assert.Empty(t, w.SavedMessage())
	assert.Empty(t, w.ErrorMessage())
	assert.Nil(t, cmd)
}

func TestConfigWizard_CurrentScreen(t *testing.T) {
	t.Parallel()

	wizard := NewConfigWizard()
	wizard.Update(tea.WindowSizeMsg{Width: 100, Height: 50})

	screen := wizard.currentScreen()
	require.NotNil(t, screen)
	assert.Equal(t, config.CategoryScanner, screen.Category())
}

func TestConfigWizard_CurrentScreen_OutOfBounds(t *testing.T) {
	t.Parallel()

	wizard := NewConfigWizard()
	wizard.activeCategory = -1
	assert.Nil(t, wizard.currentScreen())

	wizard.activeCategory = 100
	assert.Nil(t, wizard.currentScreen())
}

func TestConfigWizard_HasUnsavedChanges_Empty(t *testing.T) {
	t.Parallel()

	wizard := NewConfigWizard()
	assert.False(t, wizard.hasUnsavedChanges())
}

func TestConfigWizard_DelegateToScreen(t *testing.T) {
	t.Parallel()

	wizard := NewConfigWizard()
	wizard.Update(tea.WindowSizeMsg{Width: 100, Height: 50})

	wizard.Update(tea.KeyMsg{Type: tea.KeyDown})
}

func TestConfigWizard_RenderSidebar_WithChanges(t *testing.T) {
	t.Parallel()

	wizard := NewConfigWizard()
	wizard.Update(tea.WindowSizeMsg{Width: 100, Height: 50})

	sidebar := wizard.renderSidebar()
	assert.NotEmpty(t, sidebar)
	assert.Contains(t, sidebar, "Categories")
}

func TestConfigWizard_RenderHeader(t *testing.T) {
	t.Parallel()

	wizard := NewConfigWizard()
	wizard.width = 100

	header := wizard.renderHeader()
	assert.NotEmpty(t, header)
	assert.Contains(t, header, "Configuration Settings")
}

func TestConfigWizard_RenderFooter(t *testing.T) {
	t.Parallel()

	wizard := NewConfigWizard()

	footer := wizard.renderFooter()
	assert.NotEmpty(t, footer)
	assert.Contains(t, footer, "Tab")
	assert.Contains(t, footer, "Save")
}

func TestConfigWizard_ConfirmQuit_Save(t *testing.T) {
	t.Parallel()

	wizard := NewConfigWizard()
	wizard.confirmQuit = true

	model, cmd := wizard.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})

	w := model.(*ConfigWizardModel)
	assert.False(t, w.ConfirmingQuit())
	assert.NotNil(t, cmd)
}

func TestConfigWizard_CtrlC_WithChanges(t *testing.T) {
	t.Parallel()

	wizard := NewConfigWizard()
	wizard.Update(tea.WindowSizeMsg{Width: 100, Height: 50})

	model, _ := wizard.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	w := model.(*ConfigWizardModel)
	assert.False(t, w.ConfirmingQuit())
}

func TestConfigWizard_CtrlQ_NoChanges(t *testing.T) {
	t.Parallel()

	wizard := NewConfigWizard()
	wizard.Update(tea.WindowSizeMsg{Width: 100, Height: 50})

	_, cmd := wizard.Update(tea.KeyMsg{Type: tea.KeyCtrlQ})
	assert.NotNil(t, cmd)
}

func TestConfigWizard_CtrlS(t *testing.T) {
	t.Parallel()

	wizard := NewConfigWizard()
	wizard.Update(tea.WindowSizeMsg{Width: 100, Height: 50})

	_, cmd := wizard.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	assert.NotNil(t, cmd)
}

func TestConfigWizard_UnknownMsg(t *testing.T) {
	t.Parallel()

	wizard := NewConfigWizard()

	type unknownMsg struct{}
	model, cmd := wizard.Update(unknownMsg{})

	assert.NotNil(t, model)
	assert.Nil(t, cmd)
}
