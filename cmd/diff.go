package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Diff management tools",
	Long:  "Commands for splitting and managing large diff files",
}

var diffSplitCmd = &cobra.Command{
	Use:   "split",
	Short: "Split large diff into chunks",
	Long: `Split a large diff file into smaller, manageable chunks while preserving diff context.

This command intelligently splits diff files by keeping related changes together,
ensuring that each chunk maintains proper diff headers and context. It avoids
breaking in the middle of file changes and includes appropriate chunk headers.

By default, each chunk includes metadata headers. Use --no-header to omit headers
for better patch applicability with standard patch tools.

Examples:
  shotgun-cli diff split --input large-diff.patch
  shotgun-cli diff split --input changes.diff --output-dir chunks --approx-lines 1000
  shotgun-cli diff split -i feature-branch.diff -o review-chunks --approx-lines 300
  shotgun-cli diff split --input patch.diff --no-header  # For patch tool compatibility`,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		input, _ := cmd.Flags().GetString("input")
		if input == "" {
			return fmt.Errorf("input file is required")
		}

		// Check if input file exists
		if _, err := os.Stat(input); os.IsNotExist(err) {
			return fmt.Errorf("input file does not exist: %s", input)
		}

		// Check if input file is readable
		file, err := os.Open(input)
		if err != nil {
			return fmt.Errorf("cannot read input file '%s': %w", input, err)
		}
		_ = file.Close()

		// Validate approx-lines
		approxLines, _ := cmd.Flags().GetInt("approx-lines")
		if approxLines <= 0 {
			return fmt.Errorf("approx-lines must be positive, got: %d", approxLines)
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		input, _ := cmd.Flags().GetString("input")
		outputDir, _ := cmd.Flags().GetString("output-dir")
		approxLines, _ := cmd.Flags().GetInt("approx-lines")
		noHeader, _ := cmd.Flags().GetBool("no-header")

		log.Info().
			Str("input", input).
			Str("outputDir", outputDir).
			Int("approxLines", approxLines).
			Bool("noHeader", noHeader).
			Msg("Starting diff split")

		if err := splitDiffFile(input, outputDir, approxLines, noHeader); err != nil {
			return fmt.Errorf("failed to split diff file: %w", err)
		}

		fmt.Printf("✅ Diff file split successfully!\n")
		fmt.Printf("📁 Input file: %s\n", input)
		fmt.Printf("📂 Output directory: %s\n", outputDir)

		return nil
	},
}

type DiffChunk struct {
	Lines     []string
	FileCount int
	StartLine int
}

func splitDiffFile(inputPath, outputDir string, approxLines int, noHeader bool) error {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Open input file
	file, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer file.Close()

	// Read and parse diff
	scanner := bufio.NewScanner(file)
	var allLines []string
	for scanner.Scan() {
		allLines = append(allLines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	if len(allLines) == 0 {
		return fmt.Errorf("input file is empty")
	}

	log.Info().Int("totalLines", len(allLines)).Msg("Read diff file")

	// Split into chunks
	chunks := intelligentSplitDiff(allLines, approxLines)

	log.Info().Int("chunks", len(chunks)).Msg("Split into chunks")

	// Write chunks to files
	inputBasename := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))

	for i, chunk := range chunks {
		chunkFilename := fmt.Sprintf("%s-chunk-%02d.diff", inputBasename, i+1)
		chunkPath := filepath.Join(outputDir, chunkFilename)

		if err := writeChunk(chunkPath, chunk, i+1, len(chunks), noHeader); err != nil {
			return fmt.Errorf("failed to write chunk %d: %w", i+1, err)
		}

		fmt.Printf("📄 Chunk %d: %s (%d lines, %d files)\n",
			i+1, chunkFilename, len(chunk.Lines), chunk.FileCount)
	}

	return nil
}

func intelligentSplitDiff(lines []string, approxLines int) []DiffChunk {
	var chunks []DiffChunk
	var currentChunk DiffChunk
	currentFileLines := make([]string, 0, approxLines)
	fileCount := 0
	inFileSection := false

	for i, line := range lines {
		currentChunk.Lines = append(currentChunk.Lines, line)
		currentFileLines = append(currentFileLines, line)

		// Track file boundaries
		if isDiffHeader(line) || isGitDiffHeader(line) {
			if inFileSection && len(currentFileLines) > 1 {
				// We've finished the previous file
				fileCount++
			}
			inFileSection = true
			currentFileLines = []string{line}
		}

		// Check if we should create a new chunk
		shouldSplit := len(currentChunk.Lines) >= approxLines &&
			canSplitHere(lines, i, inFileSection)

		if shouldSplit {
			currentChunk.FileCount = fileCount
			if inFileSection && len(currentFileLines) > 1 {
				currentChunk.FileCount++
			}
			chunks = append(chunks, currentChunk)

			// Start new chunk
			currentChunk = DiffChunk{StartLine: i + 1}
			fileCount = 0
			inFileSection = false
			currentFileLines = []string{}
		}
	}

	// Add final chunk if it has content
	if len(currentChunk.Lines) > 0 {
		if inFileSection && len(currentFileLines) > 1 {
			fileCount++
		}
		currentChunk.FileCount = fileCount
		chunks = append(chunks, currentChunk)
	}

	// Ensure we have at least one chunk
	if len(chunks) == 0 {
		chunks = append(chunks, DiffChunk{
			Lines:     lines,
			FileCount: countFiles(lines),
			StartLine: 1,
		})
	}

	return chunks
}

func isDiffHeader(line string) bool {
	return strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++")
}

func isGitDiffHeader(line string) bool {
	return strings.HasPrefix(line, "diff --git")
}

func isIndexLine(line string) bool {
	return strings.HasPrefix(line, "index ")
}

func canSplitHere(lines []string, index int, inFileSection bool) bool {
	if index >= len(lines)-1 {
		return false
	}

	nextLine := lines[index+1]

	// Good places to split:
	// 1. Before a new file diff starts
	if isGitDiffHeader(nextLine) || isDiffHeader(nextLine) {
		return true
	}

	// 2. At the end of a complete file section (after context lines)
	if !inFileSection {
		return true
	}

	// 3. Between files but not in the middle of a file change
	if strings.HasPrefix(nextLine, "@@") {
		// This is a hunk header, safe to split before it
		return true
	}

	return false
}

func countFiles(lines []string) int {
	count := 0
	for _, line := range lines {
		if isGitDiffHeader(line) || (isDiffHeader(line) && strings.HasPrefix(line, "---")) {
			count++
		}
	}
	return count
}

func writeChunk(path string, chunk DiffChunk, chunkNum, totalChunks int, noHeader bool) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Conditionally write chunk header
	if !noHeader {
		fmt.Fprintf(file, "# Diff Chunk %d of %d\n", chunkNum, totalChunks)
		fmt.Fprintf(file, "# Generated by shotgun-cli diff split\n")
		fmt.Fprintf(file, "# Files in this chunk: %d\n", chunk.FileCount)
		fmt.Fprintf(file, "# Lines in this chunk: %d\n", len(chunk.Lines))
		fmt.Fprintf(file, "#\n")
	}

	// Write diff content
	for _, line := range chunk.Lines {
		fmt.Fprintln(file, line)
	}

	return nil
}

func init() {
	// Diff split flags
	diffSplitCmd.Flags().StringP("input", "i", "", "Input diff file (required)")
	diffSplitCmd.Flags().StringP("output-dir", "o", "chunks", "Output directory for chunks")
	diffSplitCmd.Flags().Int("approx-lines", 500, "Approximate lines per chunk")
	diffSplitCmd.Flags().Bool("no-header", false, "Omit metadata headers for better patch tool compatibility")

	// Mark input as required
	_ = diffSplitCmd.MarkFlagRequired("input")

	diffCmd.AddCommand(diffSplitCmd)
	rootCmd.AddCommand(diffCmd)
}
