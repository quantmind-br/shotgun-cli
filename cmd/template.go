package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/quantmind-br/shotgun-cli/internal/core/template"
)

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Template management",
	Long:  "Commands for listing, rendering, and managing prompt templates",
}

var templateListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available templates",
	Long: `List all available embedded templates with their names and descriptions.

Templates are embedded resources that provide structured prompts for different
use cases such as code review, documentation generation, refactoring, etc.

Example:
  shotgun-cli template list`,

	RunE: func(cmd *cobra.Command, args []string) error {
		manager, err := template.NewManager()
		if err != nil {
			return fmt.Errorf("failed to initialize template manager: %w", err)
		}

		templates, err := manager.ListTemplates()
		if err != nil {
			return fmt.Errorf("failed to list templates: %w", err)
		}

		if len(templates) == 0 {
			fmt.Println("No templates available.")
			return nil
		}

		// Print header
		fmt.Println("Available Templates:")
		fmt.Println("==================")
		fmt.Println()

		// Calculate column width for alignment including source
		maxNameWidth := 0
		for _, tmpl := range templates {
			// Include source in width calculation: "name (source)"
			nameWithSource := fmt.Sprintf("%s (%s)", tmpl.Name, tmpl.Source)
			if len(nameWithSource) > maxNameWidth {
				maxNameWidth = len(nameWithSource)
			}
		}

		// Print templates in a formatted table
		for _, tmpl := range templates {
			nameWithSource := fmt.Sprintf("%s (%s)", tmpl.Name, tmpl.Source)
			nameFormatted := fmt.Sprintf("%-*s", maxNameWidth, nameWithSource)
			fmt.Printf("  %s  %s\n", nameFormatted, tmpl.Description)
		}

		fmt.Printf("\nTotal: %d templates\n", len(templates))
		fmt.Println("\nUse 'shotgun-cli template render <name>' to render a specific template.")

		return nil
	},
}

var templateRenderCmd = &cobra.Command{
	Use:   "render [template-name]",
	Short: "Render a template with variables",
	Long: `Render a specific template with variable substitution.

Templates can contain variables in the format {{.VariableName}} which will be
replaced with values provided via the --var flag. All required variables must
be provided for successful rendering. The command validates required variables
before rendering and will fail with a helpful error if any are missing.

Examples:
  shotgun-cli template render code-review
  shotgun-cli template render refactor --var language=go --var style=functional
  shotgun-cli template render documentation --var project=myapp --output docs.md
  shotgun-cli template render bug-fix --var severity=high --var component=auth`,

	Args: cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		manager, err := template.NewManager()
		if err != nil {
			log.Debug().Err(err).Msg("Failed to initialize template manager for completion")
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		templates, err := manager.ListTemplates()
		if err != nil {
			log.Debug().Err(err).Msg("Failed to list templates for completion")
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		completions := make([]string, 0, len(templates))
		for _, tmpl := range templates {
			if strings.HasPrefix(tmpl.Name, toComplete) {
				// Include description in completion for better UX
				completions = append(completions, fmt.Sprintf("%s\t%s", tmpl.Name, tmpl.Description))
			}
		}

		return completions, cobra.ShellCompDirectiveNoFileComp
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		templateName := args[0]

		// Verify template exists
		manager, err := template.NewManager()
		if err != nil {
			return fmt.Errorf("failed to initialize template manager: %w", err)
		}
		templates, err := manager.ListTemplates()
		if err != nil {
			return fmt.Errorf("failed to verify template existence: %w", err)
		}

		templateExists := false
		for _, tmpl := range templates {
			if tmpl.Name == templateName {
				templateExists = true
				break
			}
		}

		if !templateExists {
			availableNames := make([]string, len(templates))
			for i, tmpl := range templates {
				availableNames[i] = tmpl.Name
			}
			return fmt.Errorf("template '%s' not found. Available templates: %s",
				templateName, strings.Join(availableNames, ", "))
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		templateName := args[0]
		variables, _ := cmd.Flags().GetStringToString("var")
		output, _ := cmd.Flags().GetString("output")

		log.Debug().
			Str("template", templateName).
			Interface("variables", variables).
			Str("output", output).
			Msg("Rendering template")

		if err := renderTemplate(templateName, variables, output); err != nil {
			return fmt.Errorf("template rendering failed: %w", err)
		}

		// Success message
		if output != "" {
			fmt.Printf("✅ Template '%s' rendered successfully to: %s\n", templateName, output)
		} else {
			log.Info().Str("template", templateName).Msg("Template rendered to stdout")
		}

		return nil
	},
}

func renderTemplate(templateName string, variables map[string]string, output string) error {
	manager, err := template.NewManager()
	if err != nil {
		return fmt.Errorf("failed to initialize template manager: %w", err)
	}

	// Pre-validate required template variables
	requiredVars, err := manager.GetRequiredVariables(templateName)
	if err != nil {
		return fmt.Errorf("template '%s' not found", templateName)
	}

	if len(requiredVars) > 0 {
		var missingVars []string
		for _, requiredVar := range requiredVars {
			if _, exists := variables[requiredVar]; !exists {
				missingVars = append(missingVars, requiredVar)
			}
		}

		if len(missingVars) > 0 {
			return fmt.Errorf("missing required variables for '%s': %s. Provide via --var key=value",
				templateName, strings.Join(missingVars, ", "))
		}
	}

	// Render template
	content, err := manager.RenderTemplate(templateName, variables)
	if err != nil {
		return fmt.Errorf("failed to render template '%s': %w", templateName, err)
	}

	// Handle output
	if output != "" {
		// Write to file
		if err := os.WriteFile(output, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write output file '%s': %w", output, err)
		}
	} else {
		// Write to stdout
		fmt.Print(content)
	}

	return nil
}

var templateImportCmd = &cobra.Command{
	Use:   "import <file> [name]",
	Short: "Import a template file to user template directory",
	Long: `Import a template file from the filesystem to the user's template directory.

The template file should be in markdown format (.md) and contain valid template
variables in the format {VARIABLE_NAME}. If a name is not provided, it will be
extracted from the filename.

The template will be copied to ~/.config/shotgun-cli/templates/ and will be
available for use immediately.

Examples:
  shotgun-cli template import /path/to/mytemplate.md
  shotgun-cli template import /path/to/template.md custom-name`,

	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		var templateName string

		if len(args) == 2 {
			templateName = args[1]
		}

		// Read template file
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read template file: %w", err)
		}

		// Validate template content by parsing it
		fileName := filepath.Base(filePath)
		if templateName == "" {
			// Extract name from filename
			templateName = strings.TrimSuffix(fileName, filepath.Ext(fileName))
			// Remove "prompt_" prefix if present
			if strings.HasPrefix(templateName, "prompt_") {
				templateName = strings.TrimPrefix(templateName, "prompt_")
			}
		}

		// Use internal parseTemplate function to validate
		// We can't directly call parseTemplate as it's not exported, so we'll do basic validation
		contentStr := string(content)
		if contentStr == "" {
			return fmt.Errorf("template content is empty")
		}

		// Get user templates directory
		userTemplatesDir := filepath.Join(xdg.ConfigHome, "shotgun-cli", "templates")
		if err := os.MkdirAll(userTemplatesDir, 0755); err != nil {
			return fmt.Errorf("failed to create user templates directory: %w", err)
		}

		// Determine destination filename
		destFileName := templateName + ".md"
		destPath := filepath.Join(userTemplatesDir, destFileName)

		// Check if file already exists
		if _, err := os.Stat(destPath); err == nil {
			// File exists, ask for confirmation
			fmt.Printf("Template '%s' already exists. Overwrite? (y/N): ", templateName)
			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read user input: %w", err)
			}

			response = strings.TrimSpace(strings.ToLower(response))
			if response != "y" && response != "yes" {
				fmt.Println("Import cancelled.")
				return nil
			}
		}

		// Write template to user directory
		if err := os.WriteFile(destPath, content, 0644); err != nil {
			return fmt.Errorf("failed to write template file: %w", err)
		}

		fmt.Printf("✅ Template '%s' imported successfully to: %s\n", templateName, destPath)
		fmt.Println("Use 'shotgun-cli template list' to see all available templates.")

		return nil
	},
}

var templateExportCmd = &cobra.Command{
	Use:   "export <name> <file>",
	Short: "Export a template to a file",
	Long: `Export an existing template to a file on the filesystem.

This command can be used to backup templates, share them, or create copies
for modification. Both embedded and custom templates can be exported.

Examples:
  shotgun-cli template export analyzeBug /tmp/exported.md
  shotgun-cli template export mycustom ~/backup/custom-template.md`,

	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		templateName := args[0]
		outputPath := args[1]

		// Get template manager
		manager, err := template.NewManager()
		if err != nil {
			return fmt.Errorf("failed to initialize template manager: %w", err)
		}

		// Get template
		tmpl, err := manager.GetTemplate(templateName)
		if err != nil {
			return fmt.Errorf("template '%s' not found", templateName)
		}

		// Check if output file exists
		if _, err := os.Stat(outputPath); err == nil {
			// File exists, ask for confirmation
			fmt.Printf("File '%s' already exists. Overwrite? (y/N): ", outputPath)
			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read user input: %w", err)
			}

			response = strings.TrimSpace(strings.ToLower(response))
			if response != "y" && response != "yes" {
				fmt.Println("Export cancelled.")
				return nil
			}
		}

		// Create directory if it doesn't exist
		dir := filepath.Dir(outputPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		// Write template content to file
		if err := os.WriteFile(outputPath, []byte(tmpl.Content), 0644); err != nil {
			return fmt.Errorf("failed to write template file: %w", err)
		}

		fmt.Printf("✅ Template '%s' exported successfully to: %s\n", templateName, outputPath)

		return nil
	},
}

func init() {
	// Template render flags
	templateRenderCmd.Flags().StringToString("var", nil, "Template variables (key=value pairs)")
	templateRenderCmd.Flags().StringP("output", "o", "", "Output file (default: stdout)")

	// Add subcommands
	templateCmd.AddCommand(templateListCmd)
	templateCmd.AddCommand(templateRenderCmd)
	templateCmd.AddCommand(templateImportCmd)
	templateCmd.AddCommand(templateExportCmd)
	rootCmd.AddCommand(templateCmd)
}
