package screens

import (
	"testing"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestNewTaskInput(t *testing.T) {
	initialValue := "test value"
	model := NewTaskInput(initialValue)

	assert.NotNil(t, model)
	assert.True(t, model.focused)
	assert.Equal(t, initialValue, model.GetValue())
	assert.True(t, model.IsFocused())
}

func TestNewTaskInputEmpty(t *testing.T) {
	model := NewTaskInput("")

	assert.NotNil(t, model)
	assert.True(t, model.focused)
	assert.Equal(t, "", model.GetValue())
}

func TestTaskInputSetSize(t *testing.T) {
	model := NewTaskInput("test")

	model.SetSize(100, 50)

	assert.Equal(t, 100, model.width)
	assert.Equal(t, 50, model.height)
}

func TestTaskInputSetSizeSmallDimensions(t *testing.T) {
	model := NewTaskInput("test")

	// Test with very small dimensions
	model.SetSize(10, 5)

	assert.Equal(t, 10, model.width)
	assert.Equal(t, 5, model.height)
	// Should handle small sizes gracefully
}

func TestTaskInputUpdateEscToBlur(t *testing.T) {
	model := NewTaskInput("test")
	assert.True(t, model.IsFocused())

	cmd := model.Update(tea.KeyMsg{Type: tea.KeyEsc})

	assert.False(t, model.IsFocused())
	assert.Nil(t, cmd)
	assert.Equal(t, "test", model.GetValue())
}

func TestTaskInputUpdateEscToFocus(t *testing.T) {
	model := NewTaskInput("test")
	model.textarea.Blur() // Manually blur

	assert.False(t, model.IsFocused())

	cmd := model.Update(tea.KeyMsg{Type: tea.KeyEsc})

	assert.True(t, model.IsFocused())
	assert.Nil(t, cmd)
}

func TestTaskInputUpdateRegularKey(t *testing.T) {
	model := NewTaskInput("")

	// Type a character
	_ = model.Update(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'h'},
	})

	// The important assertion is that the character was added to the value
	assert.Contains(t, model.GetValue(), "h")
}

func TestTaskInputUpdateMultipleKeys(t *testing.T) {
	model := NewTaskInput("")

	// Type "hello"
	for _, r := range []rune{'h', 'e', 'l', 'l', 'o'} {
		model.Update(tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{r},
		})
	}

	assert.Equal(t, "hello", model.GetValue())
}

func TestTaskInputView(t *testing.T) {
	model := NewTaskInput("test task")
	model.SetSize(100, 50)

	view := model.View()

	assert.Contains(t, view, "Describe Your Task")
	assert.Contains(t, view, "test task")
	assert.Contains(t, view, "Characters:")
	// Just check that shortcuts are present (they may be formatted differently)
	assert.Contains(t, view, "Esc:")
	assert.Contains(t, view, "Help")
}

func TestTaskInputViewEmptyShowsWarning(t *testing.T) {
	model := NewTaskInput("")
	model.SetSize(100, 50)

	view := model.View()

	assert.Contains(t, view, "Describe Your Task")
	assert.Contains(t, view, "required to continue")
}

func TestTaskInputGetValue(t *testing.T) {
	model := NewTaskInput("my value")

	assert.Equal(t, "my value", model.GetValue())
}

func TestTaskInputIsValidWithContent(t *testing.T) {
	model := NewTaskInput("   valid task description   ")

	assert.True(t, model.IsValid())
}

func TestTaskInputIsValidEmpty(t *testing.T) {
	model := NewTaskInput("")

	assert.False(t, model.IsValid())
}

func TestTaskInputIsValidWhitespaceOnly(t *testing.T) {
	model := NewTaskInput("   \n\t   ")

	assert.False(t, model.IsValid())
}

func TestTaskInputIsFocusedTrue(t *testing.T) {
	model := NewTaskInput("test")

	assert.True(t, model.IsFocused())
}

func TestTaskInputIsFocusedFalse(t *testing.T) {
	model := NewTaskInput("test")
	model.textarea.Blur()

	assert.False(t, model.IsFocused())
}

func TestNewRulesInput(t *testing.T) {
	initialValue := "test rules"
	model := NewRulesInput(initialValue)

	assert.NotNil(t, model)
	assert.True(t, model.focused)
	assert.Equal(t, initialValue, model.GetValue())
}

func TestNewRulesInputEmpty(t *testing.T) {
	model := NewRulesInput("")

	assert.NotNil(t, model)
	assert.True(t, model.focused)
	assert.Equal(t, "", model.GetValue())
}

func TestRulesInputSetSize(t *testing.T) {
	model := NewRulesInput("test")

	model.SetSize(100, 50)

	assert.Equal(t, 100, model.width)
	assert.Equal(t, 50, model.height)
}

func TestRulesInputUpdateEsc(t *testing.T) {
	model := NewRulesInput("test")

	cmd := model.Update(tea.KeyMsg{Type: tea.KeyEsc})

	// Should toggle focus state
	assert.False(t, model.IsFocused())
	assert.Nil(t, cmd)
}

func TestRulesInputUpdateRegularKey(t *testing.T) {
	model := NewRulesInput("")

	// Type a character
	_ = model.Update(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'r'},
	})

	assert.Contains(t, model.GetValue(), "r")
}

func TestRulesInputView(t *testing.T) {
	model := NewRulesInput("test rules")
	model.SetSize(100, 50)

	view := model.View()

	// Rules input has different title
	assert.Contains(t, view, "Rules")
	assert.Contains(t, view, "test rules")
	assert.Contains(t, view, "Characters:")
}

func TestRulesInputViewEmpty(t *testing.T) {
	model := NewRulesInput("")
	model.SetSize(100, 50)

	view := model.View()

	// Empty rules input should be valid (it's optional)
	assert.Contains(t, view, "Rules")
	assert.Contains(t, view, "OPTIONAL")
}

func TestRulesInputGetValue(t *testing.T) {
	model := NewRulesInput("my rules")

	assert.Equal(t, "my rules", model.GetValue())
}

func TestRulesInputIsValidWithContent(t *testing.T) {
	model := NewRulesInput("   valid rules   ")

	assert.True(t, model.IsValid())
}

func TestRulesInputIsValidEmpty(t *testing.T) {
	model := NewRulesInput("")

	// Rules input is optional, so empty should be valid
	assert.True(t, model.IsValid())
}

func TestRulesInputIsValidWhitespaceOnly(t *testing.T) {
	model := NewRulesInput("   \n\t   ")

	// Rules is optional - even whitespace is acceptable
	// (users can skip this step)
	assert.True(t, model.IsValid())
}

func TestRulesInputIsFocused(t *testing.T) {
	model := NewRulesInput("test")

	assert.True(t, model.IsFocused())
}

// Test that both input types use textarea correctly
func TestInputModelsUseTextarea(t *testing.T) {
	taskModel := NewTaskInput("task")
	rulesModel := NewRulesInput("rules")

	// Both should have textarea
	assert.NotNil(t, taskModel.textarea)
	assert.NotNil(t, rulesModel.textarea)

	// Both should be textarea.Model type
	assert.IsType(t, textarea.Model{}, taskModel.textarea)
	assert.IsType(t, textarea.Model{}, rulesModel.textarea)
}
