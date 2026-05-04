package assets

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestEmbeddedTemplatesExist verifies that all required templates are properly embedded
func TestEmbeddedTemplatesExist(t *testing.T) {
	t.Parallel()

	templates := []string{
		"templates/prompt_projectManager.md",
		"templates/prompt_makePlan.md",
		"templates/prompt_makeDiffGitFormat.md",
		"templates/prompt_analyzeBug.md",
	}

	for _, tmpl := range templates {
		content, err := Templates.ReadFile(tmpl)
		require.NoError(t, err, "template %s should be embedded", tmpl)
		require.NotEmpty(t, content, "template %s should have content", tmpl)
	}
}

// TestEmbeddedTemplatesContent verifies that embedded templates have non-empty content
func TestEmbeddedTemplatesContent(t *testing.T) {
	t.Parallel()

	templates := []string{
		"templates/prompt_projectManager.md",
		"templates/prompt_makePlan.md",
		"templates/prompt_makeDiffGitFormat.md",
		"templates/prompt_analyzeBug.md",
	}

	for _, tmpl := range templates {
		content, err := Templates.ReadFile(tmpl)
		require.NoError(t, err, "template %s should be embedded", tmpl)
		require.NotEmpty(t, content, "template %s should have non-empty content", tmpl)
		require.Greater(t, len(content), 100, "template %s should have reasonable length", tmpl)
	}
}
