// Package diff provides utilities for processing and splitting git diff files.
package diff

import (
	"strconv"
	"strings"
)

// Chunk represents a portion of a diff that has been split.
type Chunk struct {
	// Lines contains the diff content lines for this chunk.
	Lines []string

	// FileCount is the number of files represented in this chunk.
	FileCount int

	// StartLine is the original line number where this chunk starts (1-indexed).
	// When a chunk repeats the header of a file continued from the previous
	// chunk, StartLine points at the first original (non-repeated) line.
	StartLine int
}

// SplitConfig configures how diffs are split.
type SplitConfig struct {
	// ApproxLines is the target number of lines per chunk.
	// The actual chunk size may vary to preserve file boundaries.
	ApproxLines int
}

// DefaultSplitConfig returns a default split configuration.
func DefaultSplitConfig() SplitConfig {
	return SplitConfig{
		ApproxLines: 500,
	}
}

// hunk is a single "@@" section of a file diff.
type hunk struct {
	lines []string
	// at is the 1-indexed original line number of the hunk header.
	at int
}

// fileSection is one file's diff: its header block plus its hunks.
type fileSection struct {
	header []string
	// at is the 1-indexed original line number of the first header line.
	at    int
	hunks []hunk
}

func (f *fileSection) size() int {
	n := len(f.header)
	for i := range f.hunks {
		n += len(f.hunks[i].lines)
	}

	return n
}

// IntelligentSplit splits a diff into chunks while preserving file boundaries.
//
// The algorithm parses the diff into file sections and hunks, tracking the line
// budget declared by each "@@" header so that content lines are never mistaken
// for structure. Chunks are cut only between files; a file whose diff exceeds
// ApproxLines on its own is cut between its hunks, and every continuation chunk
// repeats that file's header so each chunk remains an applyable patch.
//
// Parameters:
//   - lines: The diff content as a slice of strings (one per line)
//   - config: Configuration for the split operation
//
// Returns a slice of Chunk, each containing a valid portion of the diff.
func IntelligentSplit(lines []string, config SplitConfig) []Chunk {
	if len(lines) == 0 {
		return []Chunk{{Lines: lines, FileCount: 0, StartLine: 1}}
	}

	if config.ApproxLines <= 0 {
		config.ApproxLines = 500
	}

	preamble, files := parseSections(lines)

	b := &chunkBuilder{limit: config.ApproxLines}
	b.addPreamble(preamble)
	for i := range files {
		b.addFile(&files[i])
	}

	return b.done(lines)
}

// chunkBuilder accumulates lines into chunks that respect the line budget.
type chunkBuilder struct {
	limit     int
	chunks    []Chunk
	current   Chunk
	fileCount int
	// hunksSinceHeader counts hunks emitted after the last header of the file
	// currently being written, so a repeated header never lands alone.
	hunksSinceHeader int
}

func (b *chunkBuilder) add(lines []string, at int) {
	if len(b.current.Lines) == 0 {
		b.current.StartLine = at
	}
	b.current.Lines = append(b.current.Lines, lines...)
}

func (b *chunkBuilder) flush() {
	if len(b.current.Lines) == 0 {
		return
	}
	b.current.FileCount = b.fileCount
	b.chunks = append(b.chunks, b.current)
	b.current = Chunk{}
	b.fileCount = 0
	b.hunksSinceHeader = 0
}

// addPreamble emits the lines preceding the first file section (for example a
// cover letter from git format-patch), cutting them on the line budget.
func (b *chunkBuilder) addPreamble(preamble []string) {
	for i, line := range preamble {
		if len(b.current.Lines) >= b.limit {
			b.flush()
		}
		b.add([]string{line}, i+1)
	}
}

func (b *chunkBuilder) addFile(f *fileSection) {
	if len(b.current.Lines) > 0 && len(b.current.Lines)+f.size() > b.limit {
		b.flush()
	}

	b.add(f.header, f.at)
	b.fileCount++
	b.hunksSinceHeader = 0

	for i := range f.hunks {
		h := &f.hunks[i]
		if b.hunksSinceHeader > 0 && len(b.current.Lines)+len(h.lines) > b.limit {
			b.flush()
			// Repeat the file header so the continuation chunk stays applyable.
			b.add(f.header, h.at)
			b.fileCount++
		}
		b.add(h.lines, h.at)
		b.hunksSinceHeader++
	}
}

func (b *chunkBuilder) done(lines []string) []Chunk {
	b.flush()
	if len(b.chunks) == 0 {
		return []Chunk{{Lines: lines, FileCount: CountFiles(lines), StartLine: 1}}
	}

	return b.chunks
}

// parseSections splits a diff into the leading preamble and its file sections.
func parseSections(lines []string) (preamble []string, files []fileSection) {
	var oldLeft, newLeft int
	inHunk := false

	for i, line := range lines {
		if inHunk && consumeHunkLine(line, &oldLeft, &newLeft) {
			f := &files[len(files)-1]
			h := &f.hunks[len(f.hunks)-1]
			h.lines = append(h.lines, line)
			inHunk = oldLeft > 0 || newLeft > 0

			continue
		}
		inHunk = false

		switch {
		case IsGitDiffHeader(line):
			files = append(files, fileSection{header: []string{line}, at: i + 1})
		case isHunkHeader(line):
			if len(files) == 0 {
				// A hunk with no file header: keep it as an anonymous section.
				files = append(files, fileSection{at: i + 1})
			}
			f := &files[len(files)-1]
			f.hunks = append(f.hunks, hunk{lines: []string{line}, at: i + 1})
			oldLeft, newLeft, _ = parseHunkRanges(line)
			inHunk = oldLeft > 0 || newLeft > 0
		case startsPlainFile(lines, i) && sectionIsComplete(files):
			files = append(files, fileSection{header: []string{line}, at: i + 1})
		case len(files) == 0:
			preamble = append(preamble, line)
		default:
			appendToSection(&files[len(files)-1], line)
		}
	}

	return preamble, files
}

// sectionIsComplete reports whether the last parsed section can take no more
// header lines, so a "---" line must open a new file rather than extend it.
func sectionIsComplete(files []fileSection) bool {
	return len(files) == 0 || len(files[len(files)-1].hunks) > 0
}

// startsPlainFile reports whether lines[i] opens a unified diff file section
// that has no "diff --git" line, i.e. a "--- x" immediately followed by "+++ y"
// outside of any hunk.
func startsPlainFile(lines []string, i int) bool {
	if !strings.HasPrefix(lines[i], "--- ") {
		return false
	}

	return i+1 < len(lines) && strings.HasPrefix(lines[i+1], "+++ ")
}

// appendToSection puts a non-structural line in the section's header while no
// hunk has been seen, and in the last hunk afterwards (mail signatures and the
// like trailing a file diff).
func appendToSection(f *fileSection, line string) {
	if len(f.hunks) == 0 {
		f.header = append(f.header, line)

		return
	}
	h := &f.hunks[len(f.hunks)-1]
	h.lines = append(h.lines, line)
}

// consumeHunkLine reports whether line is a hunk body line and, if so, debits
// it from the remaining old/new line budget declared by the hunk header.
func consumeHunkLine(line string, oldLeft, newLeft *int) bool {
	debit := func(n *int) {
		if *n > 0 {
			*n--
		}
	}

	if line == "" {
		// A context line whose single leading space was stripped.
		debit(oldLeft)
		debit(newLeft)

		return true
	}

	switch line[0] {
	case ' ':
		debit(oldLeft)
		debit(newLeft)
	case '-':
		debit(oldLeft)
	case '+':
		debit(newLeft)
	case '\\':
		// "\ No newline at end of file" belongs to the hunk but counts for neither side.
	default:
		return false
	}

	return true
}

// isHunkHeader reports whether the line is a well-formed "@@ -a,b +c,d @@" header.
func isHunkHeader(line string) bool {
	_, _, ok := parseHunkRanges(line)

	return ok
}

// parseHunkRanges extracts the old and new line counts from a hunk header.
func parseHunkRanges(line string) (oldCount, newCount int, ok bool) {
	if !strings.HasPrefix(line, "@@ ") {
		return 0, 0, false
	}
	rest := line[len("@@ "):]
	end := strings.Index(rest, " @@")
	if end < 0 {
		return 0, 0, false
	}
	parts := strings.Fields(rest[:end])
	if len(parts) != 2 || !strings.HasPrefix(parts[0], "-") || !strings.HasPrefix(parts[1], "+") {
		return 0, 0, false
	}

	oldCount, okOld := parseRangeCount(parts[0][1:])
	newCount, okNew := parseRangeCount(parts[1][1:])
	if !okOld || !okNew {
		return 0, 0, false
	}

	return oldCount, newCount, true
}

// parseRangeCount reads the length of a "start,length" range, defaulting to 1.
func parseRangeCount(s string) (int, bool) {
	start, length, hasLength := strings.Cut(s, ",")
	if _, err := strconv.Atoi(start); err != nil {
		return 0, false
	}
	if !hasLength {
		return 1, true
	}
	n, err := strconv.Atoi(length)
	if err != nil || n < 0 {
		return 0, false
	}

	return n, true
}

// IsDiffHeader returns true if the line is a unified diff file header
// ("--- old" or "+++ new"). The trailing space is required: without it a
// content line such as "+++i;" (the addition of "++i;") would be mistaken for
// a header. Structural decisions use the hunk-aware parser, not this helper.
func IsDiffHeader(line string) bool {
	return strings.HasPrefix(line, "--- ") || strings.HasPrefix(line, "+++ ")
}

// IsGitDiffHeader returns true if the line is a git diff header.
func IsGitDiffHeader(line string) bool {
	return strings.HasPrefix(line, "diff --git")
}

// CountFiles counts the number of files in a diff.
func CountFiles(lines []string) int {
	_, files := parseSections(lines)

	return len(files)
}

// TotalLines returns the total number of lines across all chunks.
func TotalLines(chunks []Chunk) int {
	total := 0
	for _, chunk := range chunks {
		total += len(chunk.Lines)
	}

	return total
}

// TotalFiles returns the total number of files across all chunks. A file split
// across chunks is counted once per chunk it appears in.
func TotalFiles(chunks []Chunk) int {
	total := 0
	for _, chunk := range chunks {
		total += chunk.FileCount
	}

	return total
}
