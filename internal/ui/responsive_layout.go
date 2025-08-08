package ui

import (
	"fmt"
)

// BreakpointType defines the different responsive breakpoints
type BreakpointType int

const (
	// BreakpointMobile represents screens < 60 characters wide
	BreakpointMobile BreakpointType = iota
	// BreakpointNarrow represents screens 60-79 characters wide
	BreakpointNarrow
	// BreakpointMedium represents screens 80-119 characters wide
	BreakpointMedium
	// BreakpointWide represents screens >= 120 characters wide
	BreakpointWide
)

// Breakpoint width thresholds
const (
	MobileMaxWidth = 59  // < 60 chars
	NarrowMaxWidth = 79  // 60-79 chars
	MediumMaxWidth = 119 // 80-119 chars
	WideMinWidth   = 120 // >= 120 chars
)

// LayoutContext holds information about the current layout state and calculations
type LayoutContext struct {
	Width      int
	Height     int
	Breakpoint BreakpointType

	// Calculated layout values
	FieldWidth        int
	LabelWidth        int
	ContentPadding    int
	HelpIndentation   int
	TabsVertical      bool
	ShowFullShortcuts bool
	ShowHelpText      bool
	ShowPanels        bool
	CompactMode       bool

	// Cached calculations to avoid recomputation
	calculated bool
}

// GetBreakpoint determines the breakpoint based on width
func GetBreakpoint(width int) BreakpointType {
	switch {
	case width <= MobileMaxWidth:
		return BreakpointMobile
	case width <= NarrowMaxWidth:
		return BreakpointNarrow
	case width <= MediumMaxWidth:
		return BreakpointMedium
	default:
		return BreakpointWide
	}
}

// String returns the string representation of a breakpoint
func (b BreakpointType) String() string {
	switch b {
	case BreakpointMobile:
		return "mobile"
	case BreakpointNarrow:
		return "narrow"
	case BreakpointMedium:
		return "medium"
	case BreakpointWide:
		return "wide"
	default:
		return "unknown"
	}
}

// NewLayoutContext creates a new layout context and calculates all layout values
func NewLayoutContext(width, height int) *LayoutContext {
	ctx := &LayoutContext{
		Width:      width,
		Height:     height,
		Breakpoint: GetBreakpoint(width),
	}

	ctx.calculateLayout()
	return ctx
}

// calculateLayout computes all layout-dependent values based on the current breakpoint
func (ctx *LayoutContext) calculateLayout() {
	if ctx.calculated {
		return
	}

	switch ctx.Breakpoint {
	case BreakpointMobile:
		ctx.calculateMobileLayout()
	case BreakpointNarrow:
		ctx.calculateNarrowLayout()
	case BreakpointMedium:
		ctx.calculateMediumLayout()
	case BreakpointWide:
		ctx.calculateWideLayout()
	}

	ctx.calculated = true
}

// calculateMobileLayout configures layout for mobile breakpoint (< 60 chars)
func (ctx *LayoutContext) calculateMobileLayout() {
	ctx.FieldWidth = ctx.Width - 4 // Minimal padding
	ctx.LabelWidth = 0             // No fixed label width
	ctx.ContentPadding = 1         // Minimal padding
	ctx.HelpIndentation = 2        // Minimal indentation
	ctx.TabsVertical = true        // Stack tabs vertically
	ctx.ShowFullShortcuts = false  // Abbreviated shortcuts only
	ctx.ShowHelpText = false       // Hide help text
	ctx.ShowPanels = false         // Hide optional panels
	ctx.CompactMode = true         // Ultra-compact mode
}

// calculateNarrowLayout configures layout for narrow breakpoint (60-79 chars)
func (ctx *LayoutContext) calculateNarrowLayout() {
	ctx.FieldWidth = ctx.Width - 8 // Small padding
	ctx.LabelWidth = 12            // Shorter label width
	ctx.ContentPadding = 2         // Small padding
	ctx.HelpIndentation = 4        // Small indentation
	ctx.TabsVertical = true        // Stack tabs vertically
	ctx.ShowFullShortcuts = false  // Abbreviated shortcuts
	ctx.ShowHelpText = true        // Show help but minimal
	ctx.ShowPanels = false         // Hide optional panels
	ctx.CompactMode = true         // Compact mode
}

// calculateMediumLayout configures layout for medium breakpoint (80-119 chars)
func (ctx *LayoutContext) calculateMediumLayout() {
	ctx.FieldWidth = ctx.Width - 16 // Moderate padding
	ctx.LabelWidth = 20             // Standard label width
	ctx.ContentPadding = 4          // Standard padding
	ctx.HelpIndentation = 8         // Moderate indentation
	ctx.TabsVertical = false        // Horizontal tabs
	ctx.ShowFullShortcuts = true    // Full shortcuts with wrapping
	ctx.ShowHelpText = true         // Full help text
	ctx.ShowPanels = true           // Show panels but collapsible
	ctx.CompactMode = false         // Standard mode
}

// calculateWideLayout configures layout for wide breakpoint (>= 120 chars)
func (ctx *LayoutContext) calculateWideLayout() {
	ctx.FieldWidth = ctx.Width - 24 // Full padding
	ctx.LabelWidth = 25             // Full label width
	ctx.ContentPadding = 6          // Full padding
	ctx.HelpIndentation = 12        // Full indentation
	ctx.TabsVertical = false        // Horizontal tabs
	ctx.ShowFullShortcuts = true    // Full shortcuts
	ctx.ShowHelpText = true         // Full help text
	ctx.ShowPanels = true           // Show all panels
	ctx.CompactMode = false         // Full mode
}

// Update recalculates the layout if width or height has changed
func (ctx *LayoutContext) Update(width, height int) bool {
	if ctx.Width == width && ctx.Height == height {
		return false // No change
	}

	oldBreakpoint := ctx.Breakpoint
	ctx.Width = width
	ctx.Height = height
	ctx.Breakpoint = GetBreakpoint(width)
	ctx.calculated = false

	ctx.calculateLayout()

	// Return true if breakpoint changed (triggers re-render)
	return oldBreakpoint != ctx.Breakpoint
}

// GetTabLayout returns configuration for tab rendering
func (ctx *LayoutContext) GetTabLayout() (vertical bool, useSymbols bool, maxTabs int) {
	vertical = ctx.TabsVertical

	switch ctx.Breakpoint {
	case BreakpointMobile:
		useSymbols = true
		maxTabs = 6
	case BreakpointNarrow:
		useSymbols = false
		maxTabs = 4
	default:
		useSymbols = false
		maxTabs = -1 // No limit
	}

	return vertical, useSymbols, maxTabs
}

// GetFieldLayout returns configuration for field rendering
func (ctx *LayoutContext) GetFieldLayout() (vertical bool, labelWidth int, fieldWidth int) {
	vertical = ctx.Breakpoint <= BreakpointNarrow && ctx.LabelWidth < 15
	return vertical, ctx.LabelWidth, ctx.FieldWidth
}

// GetShortcutGroups returns organized shortcuts based on available space
func (ctx *LayoutContext) GetShortcutGroups() [][]string {
	if !ctx.ShowFullShortcuts {
		// Mobile/narrow - essential shortcuts only
		return [][]string{
			{"Ctrl+S", "Enter", "?"},
		}
	}

	if ctx.Width < 100 {
		// Medium breakpoint with limited space - group shortcuts
		return [][]string{
			{"Ctrl+S: Save", "Enter: Edit", "?: Help"},
			{"Ctrl+T: Test", "Ctrl+R: Reset", "Tab: Navigate"},
		}
	}

	// Wide breakpoint - all shortcuts in one line
	return [][]string{
		{"Ctrl+S: Save", "Ctrl+T: Test", "Ctrl+R: Reset", "?: Help", "Tab: Navigate", "Enter: Edit"},
	}
}

// ShouldShowPanel determines if a panel should be shown based on priority and available space
func (ctx *LayoutContext) ShouldShowPanel(panelType string, priority int) bool {
	if !ctx.ShowPanels {
		return false
	}

	// Panel priority: 1=highest, 3=lowest
	switch ctx.Breakpoint {
	case BreakpointMobile:
		return false
	case BreakpointNarrow:
		return false
	case BreakpointMedium:
		return priority == 1 // Only highest priority panels
	case BreakpointWide:
		return true // Show all panels
	}

	return false
}

// GetHelpTextLayout returns configuration for help text rendering
func (ctx *LayoutContext) GetHelpTextLayout() (show bool, indentation int, iconOnly bool) {
	show = ctx.ShowHelpText
	indentation = ctx.HelpIndentation
	iconOnly = ctx.Breakpoint == BreakpointMobile

	return show, indentation, iconOnly
}

// CalculateContentArea returns the available content area after headers and footers
func (ctx *LayoutContext) CalculateContentArea() (contentWidth, contentHeight int) {
	// Reserve space for header (title + tabs), status bar, and padding
	headerLines := 3 // Title + subtitle + spacing
	if ctx.TabsVertical {
		headerLines += 4 // Additional lines for vertical tabs
	} else {
		headerLines += 2 // Single line for horizontal tabs + spacing
	}

	statusLines := 2 // Status bar + spacing
	if !ctx.ShowFullShortcuts {
		statusLines = 1 // Compact status bar
	}

	contentWidth = ctx.Width - (ctx.ContentPadding * 2)
	contentHeight = ctx.Height - headerLines - statusLines

	// Ensure minimum content area
	if contentWidth < 20 {
		contentWidth = 20
	}
	if contentHeight < 5 {
		contentHeight = 5
	}

	return contentWidth, contentHeight
}

// GetMaxVisibleFields calculates how many fields can be displayed given the available space
func (ctx *LayoutContext) GetMaxVisibleFields(fieldsCount int) int {
	_, contentHeight := ctx.CalculateContentArea()

	// Each field takes 1-2 lines depending on layout
	linesPerField := 1
	if vertical, _, _ := ctx.GetFieldLayout(); vertical {
		linesPerField = 2 // Label above field
	}
	if ctx.ShowHelpText {
		linesPerField++ // Additional line for help text
	}

	maxVisible := contentHeight / linesPerField

	// Always show at least 2 fields, even if cramped
	if maxVisible < 2 {
		maxVisible = 2
	}

	// Don't exceed actual field count
	if maxVisible > fieldsCount {
		maxVisible = fieldsCount
	}

	return maxVisible
}

// IsBreakpointChange returns true if changing to the given size would trigger a breakpoint change
func (ctx *LayoutContext) IsBreakpointChange(newWidth int) bool {
	return GetBreakpoint(newWidth) != ctx.Breakpoint
}

// Debug returns a string representation of the layout context for debugging
func (ctx *LayoutContext) Debug() string {
	return fmt.Sprintf(
		"LayoutContext{Width: %d, Height: %d, Breakpoint: %s, FieldWidth: %d, LabelWidth: %d, TabsVertical: %t, CompactMode: %t}",
		ctx.Width, ctx.Height, ctx.Breakpoint.String(), ctx.FieldWidth, ctx.LabelWidth, ctx.TabsVertical, ctx.CompactMode,
	)
}
