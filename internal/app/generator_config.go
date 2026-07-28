package app

import (
	"github.com/quantmind-br/shotgun-cli/internal/core/contextgen"
	"github.com/quantmind-br/shotgun-cli/internal/core/scanner"
)

// GeneratorConfigInput is everything a front end knows that the generator needs.
//
// A zero limit means "no limit", matching scanner.ScanConfig. Nothing downstream
// substitutes a different value for it, so a field left unset is visibly unset
// rather than silently replaced.
type GeneratorConfigInput struct {
	Template       string
	TemplateVars   map[string]string
	MaxTotalSize   int64
	MaxFileSize    int64
	MaxFiles       int
	SkipBinary     bool
	IncludeTree    bool
	IncludeSummary bool
	IncludeIgnored bool
}

// BuildGeneratorConfig is the only producer of a contextgen.GenerateConfig.
//
// Both front ends call it. They used to build the struct independently and had
// drifted: the headless path forwarded neither MaxFileSize nor MaxFiles, and the
// TUI forwarded none of the five limits at all, so both inherited generator
// ceilings the user never configured -- `scanner.max-files: 10000` still aborted
// at 1000. Routing every caller through one function is what keeps that from
// silently happening again.
func BuildGeneratorConfig(in GeneratorConfigInput) contextgen.GenerateConfig {
	vars := in.TemplateVars
	if vars == nil {
		vars = make(map[string]string)
	}

	return contextgen.GenerateConfig{
		MaxFileSize:    in.MaxFileSize,
		MaxTotalSize:   in.MaxTotalSize,
		MaxFiles:       in.MaxFiles,
		SkipBinary:     in.SkipBinary,
		TemplateVars:   vars,
		Template:       in.Template,
		IncludeTree:    in.IncludeTree,
		IncludeSummary: in.IncludeSummary,
		IncludeIgnored: in.IncludeIgnored,
	}
}

// ScannerLimits copies the limits a ScanConfig and the generator share, so a
// caller cannot forward one and forget the other. scanner.ScanConfig.MaxFiles is
// an int64 while the generator's is an int; the conversion is clamped rather
// than truncated, because a wrapped negative would read as "no limit".
func ScannerLimits(cfg *scanner.ScanConfig, in GeneratorConfigInput) GeneratorConfigInput {
	if cfg == nil {
		return in
	}

	in.MaxFileSize = cfg.MaxFileSize
	in.MaxFiles = clampToInt(cfg.MaxFiles)
	in.SkipBinary = cfg.SkipBinary
	in.IncludeIgnored = cfg.IncludeIgnored

	return in
}

func clampToInt(v int64) int {
	const maxInt = int64(^uint(0) >> 1)

	switch {
	case v <= 0:
		return 0
	case v > maxInt:
		return int(maxInt)
	default:
		return int(v)
	}
}
