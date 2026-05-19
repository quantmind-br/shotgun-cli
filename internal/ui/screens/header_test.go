package screens

import (
	"strings"
	"testing"
)

func TestTaskInput_View_HasStepHeader(t *testing.T) {
	m := NewTaskInput("")
	m.SetSize(80, 24)
	view := m.View()
	if !strings.Contains(view, "Step 3/5") {
		t.Errorf("TaskInput.View() missing 'Step 3/5' header\nView output:\n%s", view)
	}
}

func TestRulesInput_View_HasStepHeader(t *testing.T) {
	m := NewRulesInput("")
	m.SetSize(80, 24)
	view := m.View()
	if !strings.Contains(view, "Step 4/5") {
		t.Errorf("RulesInput.View() missing 'Step 4/5' header\nView output:\n%s", view)
	}
}
