package screens

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/quantmind-br/shotgun-cli/internal/core/template"
	"github.com/stretchr/testify/assert"
)

func TestNewTemplateSelection(t *testing.T) {
	model := NewTemplateSelection()

	assert.NotNil(t, model)
	assert.True(t, model.loading)
	assert.Nil(t, model.selectedTemplate)
	assert.Equal(t, 0, model.cursor)
}

func TestTemplateSelectionSetSize(t *testing.T) {
	model := NewTemplateSelection()

	model.SetSize(100, 50)

	assert.Equal(t, 100, model.width)
	assert.Equal(t, 50, model.height)
}

func TestTemplateSelectionLoadTemplates(t *testing.T) {
	model := NewTemplateSelection()

	cmd := model.LoadTemplates()
	assert.NotNil(t, cmd)

	// The command should return a TemplatesLoadedMsg or TemplatesErrorMsg
	// We can't fully test this without mocking the template manager,
	// but we can verify the function returns a command
}

func TestTemplateSelectionUpdateLoading(t *testing.T) {
	model := NewTemplateSelection()
	model.loading = true

	cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	assert.Nil(t, model.GetSelected())
	assert.Nil(t, cmd)
}

func TestTemplateSelectionUpdateNoTemplates(t *testing.T) {
	model := &TemplateSelectionModel{
		loading:   false,
		templates: []*template.Template{}, // Empty list
	}

	cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	assert.Nil(t, model.GetSelected())
	assert.Nil(t, cmd)
}

func TestTemplateSelectionUpdateUp(t *testing.T) {
	model := &TemplateSelectionModel{
		loading: false,
		templates: []*template.Template{
			{Name: "t1", Description: "desc1"},
			{Name: "t2", Description: "desc2"},
			{Name: "t3", Description: "desc3"},
		},
		cursor: 1,
	}

	cmd := model.Update(tea.KeyMsg{Type: tea.KeyUp})

	assert.Nil(t, model.GetSelected())
	assert.Nil(t, cmd)
	assert.Equal(t, 0, model.cursor)
}

func TestTemplateSelectionUpdateUpAtTop(t *testing.T) {
	model := &TemplateSelectionModel{
		loading: false,
		templates: []*template.Template{
			{Name: "t1", Description: "desc1"},
			{Name: "t2", Description: "desc2"},
		},
		cursor: 0,
	}

	cmd := model.Update(tea.KeyMsg{Type: tea.KeyUp})

	assert.Nil(t, model.GetSelected())
	assert.Nil(t, cmd)
	assert.Equal(t, 0, model.cursor) // Should stay at 0
}

func TestTemplateSelectionUpdateDown(t *testing.T) {
	model := &TemplateSelectionModel{
		loading: false,
		templates: []*template.Template{
			{Name: "t1", Description: "desc1"},
			{Name: "t2", Description: "desc2"},
			{Name: "t3", Description: "desc3"},
		},
		cursor: 1,
	}

	cmd := model.Update(tea.KeyMsg{Type: tea.KeyDown})

	assert.Nil(t, model.GetSelected())
	assert.Nil(t, cmd)
	assert.Equal(t, 2, model.cursor)
}

func TestTemplateSelectionUpdateDownAtBottom(t *testing.T) {
	model := &TemplateSelectionModel{
		loading: false,
		templates: []*template.Template{
			{Name: "t1", Description: "desc1"},
			{Name: "t2", Description: "desc2"},
		},
		cursor: 1,
	}

	cmd := model.Update(tea.KeyMsg{Type: tea.KeyDown})

	assert.Nil(t, model.GetSelected())
	assert.Nil(t, cmd)
	assert.Equal(t, 1, model.cursor) // Should stay at last
}

func TestTemplateSelectionUpdateEnter(t *testing.T) {
	tmpl := &template.Template{
		Name:        "test-template",
		Description: "Test description",
	}
	model := &TemplateSelectionModel{
		loading: false,
		templates: []*template.Template{
			{Name: "t1", Description: "desc1"},
			tmpl,
			{Name: "t3", Description: "desc3"},
		},
		cursor: 1,
	}

	cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	assert.Equal(t, tmpl, model.GetSelected())
	assert.Nil(t, cmd)
	assert.Equal(t, tmpl, model.selectedTemplate)
}

func TestTemplateSelectionUpdateSpace(t *testing.T) {
	tmpl := &template.Template{
		Name:        "test-template",
		Description: "Test description",
	}
	model := &TemplateSelectionModel{
		loading: false,
		templates: []*template.Template{
			{Name: "t1", Description: "desc1"},
			tmpl,
			{Name: "t3", Description: "desc3"},
		},
		cursor: 1,
	}

	cmd := model.Update(tea.KeyMsg{Type: tea.KeySpace})

	assert.Equal(t, tmpl, model.GetSelected())
	assert.Nil(t, cmd)
}

func TestTemplateSelectionUpdateHome(t *testing.T) {
	model := &TemplateSelectionModel{
		loading: false,
		templates: []*template.Template{
			{Name: "t1", Description: "desc1"},
			{Name: "t2", Description: "desc2"},
			{Name: "t3", Description: "desc3"},
		},
		cursor: 2,
	}

	cmd := model.Update(tea.KeyMsg{Type: tea.KeyHome})

	assert.Nil(t, model.GetSelected())
	assert.Nil(t, cmd)
	assert.Equal(t, 0, model.cursor)
}

func TestTemplateSelectionUpdateEnd(t *testing.T) {
	model := &TemplateSelectionModel{
		loading: false,
		templates: []*template.Template{
			{Name: "t1", Description: "desc1"},
			{Name: "t2", Description: "desc2"},
			{Name: "t3", Description: "desc3"},
		},
		cursor: 0,
	}

	cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnd})

	assert.Nil(t, model.GetSelected())
	assert.Nil(t, cmd)
	assert.Equal(t, 2, model.cursor)
}

func TestTemplateSelectionHandleMessageTemplatesLoaded(t *testing.T) {
	model := NewTemplateSelection()
	model.loading = true

	templates := []*template.Template{
		{Name: "t1", Description: "desc1"},
		{Name: "t2", Description: "desc2"},
	}

	cmd := model.HandleMessage(TemplatesLoadedMsg{Templates: templates})

	assert.Nil(t, cmd)
	assert.False(t, model.loading)
	assert.Equal(t, templates, model.templates)
	assert.Equal(t, 0, model.cursor)
	assert.Nil(t, model.err)
}

func TestTemplateSelectionHandleMessageTemplatesError(t *testing.T) {
	model := NewTemplateSelection()
	model.loading = true

	testErr := assert.AnError
	cmd := model.HandleMessage(TemplatesErrorMsg{Err: testErr})

	assert.Nil(t, cmd)
	assert.False(t, model.loading)
	assert.Equal(t, testErr, model.err)
	assert.Nil(t, model.templates)
}

func TestTemplateSelectionViewLoading(t *testing.T) {
	model := &TemplateSelectionModel{
		loading: true,
	}

	view := model.View()

	assert.Contains(t, view, "Choose Template")
	assert.Contains(t, view, "Loading templates...")
}

func TestTemplateSelectionViewError(t *testing.T) {
	model := &TemplateSelectionModel{
		loading: false,
		err:     assert.AnError,
	}

	view := model.View()

	assert.Contains(t, view, "Choose Template")
	assert.Contains(t, view, "Error loading templates")
}

func TestTemplateSelectionViewNoTemplates(t *testing.T) {
	model := &TemplateSelectionModel{
		loading:   false,
		templates: []*template.Template{},
	}

	view := model.View()

	assert.Contains(t, view, "Choose Template")
	assert.Contains(t, view, "No templates found")
}

func TestTemplateSelectionViewWithTemplates(t *testing.T) {
	model := &TemplateSelectionModel{
		loading: false,
		templates: []*template.Template{
			{Name: "t1", Description: "desc1"},
			{Name: "t2", Description: "desc2"},
		},
		cursor: 0,
		width:  100,
		height: 50,
	}

	view := model.View()

	assert.Contains(t, view, "Choose Template")
	assert.Contains(t, view, "t1")
	assert.Contains(t, view, "t2")
}

func TestTemplateSelectionViewWithSelectedTemplate(t *testing.T) {
	model := &TemplateSelectionModel{
		loading: false,
		templates: []*template.Template{
			{Name: "t1", Description: "desc1"},
			{Name: "t2", Description: "desc2"},
		},
		cursor:           1,
		selectedTemplate: &template.Template{Name: "t1", Description: "desc1"},
		width:            100,
		height:           50,
	}

	view := model.View()

	// Selected template should show checkmark
	assert.Contains(t, view, "✓")
}

func TestTemplateSelectionCalculateScrollBounds(t *testing.T) {
	// Test with fewer templates than available height
	templates := make([]*template.Template, 3)
	for i := 0; i < 3; i++ {
		templates[i] = &template.Template{Name: "t" + string(rune('1'+i))}
	}

	model := &TemplateSelectionModel{
		templates: templates,
		cursor:    1,
		height:    50,
	}

	start, end := model.calculateScrollBounds()

	assert.Equal(t, 0, start)
	assert.Equal(t, 3, end) // All templates visible
}

func TestTemplateSelectionFormatTemplateLine(t *testing.T) {
	tmpl := &template.Template{Name: "test", Description: "desc"}

	// Test cursor position
	model := &TemplateSelectionModel{
		templates: []*template.Template{tmpl},
		cursor:    0,
	}

	line := model.formatTemplateLine(0)
	assert.Contains(t, line, "test")
	assert.Contains(t, line, "▶")

	// Test selected template
	model.selectedTemplate = tmpl
	line = model.formatTemplateLine(0)
	assert.Contains(t, line, "test")
	assert.Contains(t, line, "✓")

	// Test non-cursor, non-selected
	model.cursor = 1
	line = model.formatTemplateLine(0)
	assert.Contains(t, line, "test")
}

func TestTemplateSelectionRenderFooter(t *testing.T) {
	model := &TemplateSelectionModel{}

	footer := model.renderFooter()

	assert.Contains(t, footer, "Navigate")
	assert.Contains(t, footer, "Select")
	assert.Contains(t, footer, "Help")
	assert.Contains(t, footer, "v: View full")
}

func TestTemplateModalState(t *testing.T) {
	t.Run("v key opens modal", func(t *testing.T) {
		model := &TemplateSelectionModel{
			loading:   false,
			templates: []*template.Template{{Name: "test", Content: "content"}},
			cursor:    0,
			width:     80,
			height:    24,
		}

		model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})

		assert.True(t, model.showingFullPreview)
		assert.Equal(t, 0, model.previewScrollY)
	})

	t.Run("esc closes modal", func(t *testing.T) {
		model := &TemplateSelectionModel{
			loading:            false,
			templates:          []*template.Template{{Name: "test", Content: "content"}},
			cursor:             0,
			showingFullPreview: true,
			previewScrollY:     5,
		}

		model.Update(tea.KeyMsg{Type: tea.KeyEscape})

		assert.False(t, model.showingFullPreview)
		assert.Equal(t, 0, model.previewScrollY)
	})

	t.Run("q closes modal", func(t *testing.T) {
		model := &TemplateSelectionModel{
			loading:            false,
			templates:          []*template.Template{{Name: "test", Content: "content"}},
			cursor:             0,
			showingFullPreview: true,
		}

		model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

		assert.False(t, model.showingFullPreview)
	})

	t.Run("modal preserves cursor position", func(t *testing.T) {
		model := &TemplateSelectionModel{
			loading: false,
			templates: []*template.Template{
				{Name: "t1", Content: "c1"},
				{Name: "t2", Content: "c2"},
				{Name: "t3", Content: "c3"},
			},
			cursor: 2,
		}

		model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
		assert.True(t, model.showingFullPreview)

		model.Update(tea.KeyMsg{Type: tea.KeyEscape})

		assert.Equal(t, 2, model.cursor)
	})
}

func TestTemplateModalScrolling(t *testing.T) {
	createLongTemplate := func() *TemplateSelectionModel {
		lines := make([]string, 50)
		for i := range lines {
			lines[i] = "Line content"
		}
		return &TemplateSelectionModel{
			loading:            false,
			templates:          []*template.Template{{Name: "long", Content: "Line1\nLine2\nLine3\nLine4\nLine5\nLine6\nLine7\nLine8\nLine9\nLine10\nLine11\nLine12\nLine13\nLine14\nLine15\nLine16\nLine17\nLine18\nLine19\nLine20"}},
			cursor:             0,
			showingFullPreview: true,
			height:             20,
			width:              80,
		}
	}

	t.Run("j scrolls down", func(t *testing.T) {
		model := createLongTemplate()
		model.previewScrollY = 0

		model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})

		assert.Equal(t, 1, model.previewScrollY)
	})

	t.Run("k scrolls up", func(t *testing.T) {
		model := createLongTemplate()
		model.previewScrollY = 5

		model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})

		assert.Equal(t, 4, model.previewScrollY)
	})

	t.Run("k does not scroll below 0", func(t *testing.T) {
		model := createLongTemplate()
		model.previewScrollY = 0

		model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})

		assert.Equal(t, 0, model.previewScrollY)
	})

	t.Run("g jumps to top", func(t *testing.T) {
		model := createLongTemplate()
		model.previewScrollY = 10

		model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})

		assert.Equal(t, 0, model.previewScrollY)
	})

	t.Run("G jumps to bottom", func(t *testing.T) {
		model := createLongTemplate()
		model.previewScrollY = 0

		model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})

		assert.Greater(t, model.previewScrollY, 0)
	})
}

func TestTemplateModalRendering(t *testing.T) {
	t.Run("renders template name in header", func(t *testing.T) {
		model := &TemplateSelectionModel{
			loading:            false,
			templates:          []*template.Template{{Name: "my-template", Content: "content"}},
			cursor:             0,
			showingFullPreview: true,
			width:              80,
			height:             24,
		}

		view := model.renderFullPreviewModal()

		assert.Contains(t, view, "my-template")
	})

	t.Run("renders complete content", func(t *testing.T) {
		model := &TemplateSelectionModel{
			loading:            false,
			templates:          []*template.Template{{Name: "test", Content: "Line1\nLine2\nLine3\nLine4\nLine5"}},
			cursor:             0,
			showingFullPreview: true,
			width:              80,
			height:             50,
		}

		view := model.renderFullPreviewModal()

		assert.Contains(t, view, "Line1")
		assert.Contains(t, view, "Line5")
	})

	t.Run("shows close hint in footer", func(t *testing.T) {
		model := &TemplateSelectionModel{
			loading:            false,
			templates:          []*template.Template{{Name: "test", Content: "content"}},
			cursor:             0,
			showingFullPreview: true,
			width:              80,
			height:             24,
		}

		view := model.renderFullPreviewModal()

		assert.Contains(t, view, "Esc")
		assert.Contains(t, view, "Close")
	})
}

func TestTemplateKeyRouting(t *testing.T) {
	t.Run("normal keys routed to list when modal closed", func(t *testing.T) {
		model := &TemplateSelectionModel{
			loading: false,
			templates: []*template.Template{
				{Name: "t1", Content: "c1"},
				{Name: "t2", Content: "c2"},
			},
			cursor:             0,
			showingFullPreview: false,
		}

		model.Update(tea.KeyMsg{Type: tea.KeyDown})

		assert.Equal(t, 1, model.cursor)
	})

	t.Run("keys routed to modal when open", func(t *testing.T) {
		model := &TemplateSelectionModel{
			loading: false,
			templates: []*template.Template{
				{Name: "t1", Content: "c1"},
				{Name: "t2", Content: "c2"},
			},
			cursor:             0,
			showingFullPreview: true,
			height:             24,
		}

		model.Update(tea.KeyMsg{Type: tea.KeyDown})

		assert.Equal(t, 0, model.cursor)
	})
}

func TestViewShowsModalWhenOpen(t *testing.T) {
	model := &TemplateSelectionModel{
		loading:            false,
		templates:          []*template.Template{{Name: "test", Content: "content"}},
		cursor:             0,
		showingFullPreview: true,
		width:              80,
		height:             24,
	}

	view := model.View()

	assert.Contains(t, view, "test")
	assert.Contains(t, view, "content")
	assert.Contains(t, view, "Close")
}
