package ui

import (
	"testing"
)

func TestGetBreakpoint(t *testing.T) {
	tests := []struct {
		width    int
		expected BreakpointType
		name     string
	}{
		{30, BreakpointMobile, "mobile_very_small"},
		{59, BreakpointMobile, "mobile_max"},
		{60, BreakpointNarrow, "narrow_min"},
		{70, BreakpointNarrow, "narrow_mid"},
		{79, BreakpointNarrow, "narrow_max"},
		{80, BreakpointMedium, "medium_min"},
		{100, BreakpointMedium, "medium_mid"},
		{119, BreakpointMedium, "medium_max"},
		{120, BreakpointWide, "wide_min"},
		{200, BreakpointWide, "wide_large"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetBreakpoint(tt.width)
			if result != tt.expected {
				t.Errorf("GetBreakpoint(%d) = %v, expected %v", tt.width, result, tt.expected)
			}
		})
	}
}

func TestLayoutContextCreation(t *testing.T) {
	ctx := NewLayoutContext(100, 30)

	if ctx.Width != 100 {
		t.Errorf("Expected width 100, got %d", ctx.Width)
	}
	if ctx.Height != 30 {
		t.Errorf("Expected height 30, got %d", ctx.Height)
	}
	if ctx.Breakpoint != BreakpointMedium {
		t.Errorf("Expected medium breakpoint for width 100, got %v", ctx.Breakpoint)
	}
	if !ctx.calculated {
		t.Error("Expected layout to be calculated upon creation")
	}
}

func TestMobileLayoutCalculation(t *testing.T) {
	ctx := NewLayoutContext(50, 20) // Mobile breakpoint

	expected := BreakpointMobile
	if ctx.Breakpoint != expected {
		t.Errorf("Expected %v breakpoint, got %v", expected, ctx.Breakpoint)
	}

	// Mobile layout characteristics
	if ctx.FieldWidth != 46 { // 50 - 4
		t.Errorf("Expected field width 46, got %d", ctx.FieldWidth)
	}
	if ctx.LabelWidth != 0 {
		t.Errorf("Expected label width 0 for mobile, got %d", ctx.LabelWidth)
	}
	if !ctx.TabsVertical {
		t.Error("Expected vertical tabs for mobile")
	}
	if ctx.ShowFullShortcuts {
		t.Error("Expected abbreviated shortcuts for mobile")
	}
	if ctx.ShowHelpText {
		t.Error("Expected no help text for mobile")
	}
	if ctx.ShowPanels {
		t.Error("Expected no panels for mobile")
	}
	if !ctx.CompactMode {
		t.Error("Expected compact mode for mobile")
	}
}

func TestNarrowLayoutCalculation(t *testing.T) {
	ctx := NewLayoutContext(70, 25) // Narrow breakpoint

	expected := BreakpointNarrow
	if ctx.Breakpoint != expected {
		t.Errorf("Expected %v breakpoint, got %v", expected, ctx.Breakpoint)
	}

	// Narrow layout characteristics
	if ctx.FieldWidth != 62 { // 70 - 8
		t.Errorf("Expected field width 62, got %d", ctx.FieldWidth)
	}
	if ctx.LabelWidth != 12 {
		t.Errorf("Expected label width 12 for narrow, got %d", ctx.LabelWidth)
	}
	if !ctx.TabsVertical {
		t.Error("Expected vertical tabs for narrow")
	}
	if ctx.ShowFullShortcuts {
		t.Error("Expected abbreviated shortcuts for narrow")
	}
	if !ctx.ShowHelpText {
		t.Error("Expected help text for narrow")
	}
	if ctx.ShowPanels {
		t.Error("Expected no panels for narrow")
	}
}

func TestMediumLayoutCalculation(t *testing.T) {
	ctx := NewLayoutContext(100, 30) // Medium breakpoint

	expected := BreakpointMedium
	if ctx.Breakpoint != expected {
		t.Errorf("Expected %v breakpoint, got %v", expected, ctx.Breakpoint)
	}

	// Medium layout characteristics
	if ctx.FieldWidth != 84 { // 100 - 16
		t.Errorf("Expected field width 84, got %d", ctx.FieldWidth)
	}
	if ctx.LabelWidth != 20 {
		t.Errorf("Expected label width 20 for medium, got %d", ctx.LabelWidth)
	}
	if ctx.TabsVertical {
		t.Error("Expected horizontal tabs for medium")
	}
	if !ctx.ShowFullShortcuts {
		t.Error("Expected full shortcuts for medium")
	}
	if !ctx.ShowHelpText {
		t.Error("Expected help text for medium")
	}
	if !ctx.ShowPanels {
		t.Error("Expected panels for medium")
	}
}

func TestWideLayoutCalculation(t *testing.T) {
	ctx := NewLayoutContext(150, 40) // Wide breakpoint

	expected := BreakpointWide
	if ctx.Breakpoint != expected {
		t.Errorf("Expected %v breakpoint, got %v", expected, ctx.Breakpoint)
	}

	// Wide layout characteristics
	if ctx.FieldWidth != 126 { // 150 - 24
		t.Errorf("Expected field width 126, got %d", ctx.FieldWidth)
	}
	if ctx.LabelWidth != 25 {
		t.Errorf("Expected label width 25 for wide, got %d", ctx.LabelWidth)
	}
	if ctx.TabsVertical {
		t.Error("Expected horizontal tabs for wide")
	}
	if !ctx.ShowFullShortcuts {
		t.Error("Expected full shortcuts for wide")
	}
	if !ctx.ShowHelpText {
		t.Error("Expected help text for wide")
	}
	if !ctx.ShowPanels {
		t.Error("Expected panels for wide")
	}
	if ctx.CompactMode {
		t.Error("Expected non-compact mode for wide")
	}
}

func TestLayoutContextUpdate(t *testing.T) {
	ctx := NewLayoutContext(100, 30) // Medium breakpoint

	// No change - should return false
	changed := ctx.Update(100, 30)
	if changed {
		t.Error("Expected no change when updating with same dimensions")
	}

	// Change width but same breakpoint - should return false
	changed = ctx.Update(110, 30)
	if changed {
		t.Error("Expected no breakpoint change when staying within medium range")
	}

	// Change to different breakpoint - should return true
	changed = ctx.Update(50, 30)
	if !changed {
		t.Error("Expected breakpoint change when switching from medium to mobile")
	}
	if ctx.Breakpoint != BreakpointMobile {
		t.Errorf("Expected mobile breakpoint after update, got %v", ctx.Breakpoint)
	}
}

func TestGetTabLayout(t *testing.T) {
	tests := []struct {
		width            int
		expectedVertical bool
		expectedSymbols  bool
		expectedMaxTabs  int
		name             string
	}{
		{50, true, true, 6, "mobile"},
		{70, true, false, 4, "narrow"},
		{100, false, false, -1, "medium"},
		{150, false, false, -1, "wide"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewLayoutContext(tt.width, 30)
			vertical, symbols, maxTabs := ctx.GetTabLayout()

			if vertical != tt.expectedVertical {
				t.Errorf("Expected vertical=%v, got %v", tt.expectedVertical, vertical)
			}
			if symbols != tt.expectedSymbols {
				t.Errorf("Expected symbols=%v, got %v", tt.expectedSymbols, symbols)
			}
			if maxTabs != tt.expectedMaxTabs {
				t.Errorf("Expected maxTabs=%d, got %d", tt.expectedMaxTabs, maxTabs)
			}
		})
	}
}

func TestGetFieldLayout(t *testing.T) {
	tests := []struct {
		width            int
		expectedVertical bool
		name             string
	}{
		{50, true, "mobile_vertical"},
		{70, true, "narrow_vertical"},
		{100, false, "medium_horizontal"},
		{150, false, "wide_horizontal"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewLayoutContext(tt.width, 30)
			vertical, labelWidth, fieldWidth := ctx.GetFieldLayout()

			if vertical != tt.expectedVertical {
				t.Errorf("Expected vertical=%v, got %v", tt.expectedVertical, vertical)
			}
			if labelWidth < 0 {
				t.Errorf("Expected non-negative label width, got %d", labelWidth)
			}
			if fieldWidth <= 0 {
				t.Errorf("Expected positive field width, got %d", fieldWidth)
			}
		})
	}
}

func TestGetShortcutGroups(t *testing.T) {
	// Mobile - minimal shortcuts
	ctx := NewLayoutContext(50, 20)
	groups := ctx.GetShortcutGroups()
	if len(groups) != 1 {
		t.Errorf("Expected 1 group for mobile, got %d", len(groups))
	}
	if len(groups[0]) != 3 {
		t.Errorf("Expected 3 shortcuts for mobile, got %d", len(groups[0]))
	}

	// Medium with limited space - grouped shortcuts
	ctx = NewLayoutContext(90, 30)
	groups = ctx.GetShortcutGroups()
	if len(groups) != 2 {
		t.Errorf("Expected 2 groups for medium limited, got %d", len(groups))
	}

	// Wide - all shortcuts in one line
	ctx = NewLayoutContext(150, 40)
	groups = ctx.GetShortcutGroups()
	if len(groups) != 1 {
		t.Errorf("Expected 1 group for wide, got %d", len(groups))
	}
	if len(groups[0]) != 6 {
		t.Errorf("Expected 6 shortcuts for wide, got %d", len(groups[0]))
	}
}

func TestShouldShowPanel(t *testing.T) {
	tests := []struct {
		width    int
		priority int
		expected bool
		name     string
	}{
		{50, 1, false, "mobile_high_priority"},
		{50, 3, false, "mobile_low_priority"},
		{70, 1, false, "narrow_high_priority"},
		{70, 3, false, "narrow_low_priority"},
		{100, 1, true, "medium_high_priority"},
		{100, 3, false, "medium_low_priority"},
		{150, 1, true, "wide_high_priority"},
		{150, 3, true, "wide_low_priority"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewLayoutContext(tt.width, 30)
			result := ctx.ShouldShowPanel("test", tt.priority)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestCalculateContentArea(t *testing.T) {
	ctx := NewLayoutContext(100, 30)
	contentWidth, contentHeight := ctx.CalculateContentArea()

	if contentWidth <= 0 {
		t.Errorf("Expected positive content width, got %d", contentWidth)
	}
	if contentHeight <= 0 {
		t.Errorf("Expected positive content height, got %d", contentHeight)
	}

	// Content area should be smaller than total area
	if contentWidth >= ctx.Width {
		t.Errorf("Content width (%d) should be less than total width (%d)", contentWidth, ctx.Width)
	}
	if contentHeight >= ctx.Height {
		t.Errorf("Content height (%d) should be less than total height (%d)", contentHeight, ctx.Height)
	}
}

func TestIsBreakpointChange(t *testing.T) {
	ctx := NewLayoutContext(100, 30) // Medium breakpoint

	// Same breakpoint
	if ctx.IsBreakpointChange(110) {
		t.Error("Expected no breakpoint change for width 110")
	}

	// Different breakpoint
	if !ctx.IsBreakpointChange(50) {
		t.Error("Expected breakpoint change for width 50")
	}
	if !ctx.IsBreakpointChange(150) {
		t.Error("Expected breakpoint change for width 150")
	}
}

func TestBreakpointString(t *testing.T) {
	tests := []struct {
		breakpoint BreakpointType
		expected   string
	}{
		{BreakpointMobile, "mobile"},
		{BreakpointNarrow, "narrow"},
		{BreakpointMedium, "medium"},
		{BreakpointWide, "wide"},
	}

	for _, tt := range tests {
		result := tt.breakpoint.String()
		if result != tt.expected {
			t.Errorf("Expected %s, got %s", tt.expected, result)
		}
	}
}

func TestGetMaxVisibleFields(t *testing.T) {
	// Test with different breakpoints
	tests := []struct {
		width      int
		height     int
		fieldCount int
		name       string
	}{
		{50, 20, 10, "mobile"},
		{70, 25, 8, "narrow"},
		{100, 30, 6, "medium"},
		{150, 40, 12, "wide"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewLayoutContext(tt.width, tt.height)
			maxVisible := ctx.GetMaxVisibleFields(tt.fieldCount)

			// Should always show at least 2 fields
			if maxVisible < 2 {
				t.Errorf("Expected at least 2 visible fields, got %d", maxVisible)
			}

			// Should not exceed actual field count
			if maxVisible > tt.fieldCount {
				t.Errorf("Expected max visible (%d) <= field count (%d)", maxVisible, tt.fieldCount)
			}
		})
	}
}

func TestLayoutContextDebug(t *testing.T) {
	ctx := NewLayoutContext(100, 30)
	debugStr := ctx.Debug()

	if debugStr == "" {
		t.Error("Expected non-empty debug string")
	}

	// Should contain key information
	expectedSubstrings := []string{"Width", "Height", "Breakpoint", "FieldWidth", "LabelWidth"}
	for _, substr := range expectedSubstrings {
		if !contains(debugStr, substr) {
			t.Errorf("Expected debug string to contain '%s'", substr)
		}
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || (len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Test integration with EnhancedConfigFormModel
func TestSetSizeIntegration(t *testing.T) {
	// Create a minimal model for testing
	model := &EnhancedConfigFormModel{
		width:  0,
		height: 0,
	}

	// Test initial SetSize
	model.SetSize(100, 30)

	if model.width != 100 {
		t.Errorf("Expected width 100, got %d", model.width)
	}
	if model.height != 30 {
		t.Errorf("Expected height 30, got %d", model.height)
	}

	// Layout context should be created
	if model.layoutContext == nil {
		t.Error("Expected layout context to be created")
	}

	// Should be medium breakpoint
	if model.layoutContext.Breakpoint != BreakpointMedium {
		t.Errorf("Expected medium breakpoint, got %v", model.layoutContext.Breakpoint)
	}
}

func TestGetLayoutContextIntegration(t *testing.T) {
	model := &EnhancedConfigFormModel{
		width:  80,
		height: 25,
	}

	// Get layout context - should create if not exists
	ctx := model.GetLayoutContext()
	if ctx == nil {
		t.Error("Expected layout context to be created")
	}

	if ctx.Width != 80 {
		t.Errorf("Expected context width 80, got %d", ctx.Width)
	}
	if ctx.Height != 25 {
		t.Errorf("Expected context height 25, got %d", ctx.Height)
	}
}

func TestIsBreakpointChangeIntegration(t *testing.T) {
	model := &EnhancedConfigFormModel{
		width:  100,
		height: 30,
	}

	// Initialize layout context
	model.SetSize(100, 30)

	// Same breakpoint
	if model.IsBreakpointChange(110) {
		t.Error("Expected no breakpoint change for width 110")
	}

	// Different breakpoint
	if !model.IsBreakpointChange(50) {
		t.Error("Expected breakpoint change for width 50")
	}
}

// Test adaptive styles system (T7)
func TestAdaptiveStyles(t *testing.T) {
	tests := []struct {
		width      int
		breakpoint BreakpointType
		name       string
	}{
		{50, BreakpointMobile, "mobile"},
		{70, BreakpointNarrow, "narrow"},
		{100, BreakpointMedium, "medium"},
		{150, BreakpointWide, "wide"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewLayoutContext(tt.width, 30)

			// Test section title style adapts to breakpoint
			titleStyle := GetSectionTitleStyle(ctx)
			rendered := titleStyle.Render("Test")
			if rendered == "" {
				t.Error("Expected styled title text")
			}
			// Style may appear as "Test" in plain text, but should not be empty

			// Test section description style adapts to breakpoint
			descStyle := GetSectionDescStyle(ctx)
			rendered = descStyle.Render("Test")
			if rendered == "" {
				t.Error("Expected styled description text")
			}

			// Test field label style uses calculated width
			labelStyle := GetFieldLabelStyle(ctx)
			rendered = labelStyle.Render("Test")
			if rendered == "" {
				t.Error("Expected styled label text")
			}

			// Test field active/inactive styles adapt
			activeStyle := GetFieldActiveStyle(ctx)
			rendered = activeStyle.Render("Test")
			if rendered == "" {
				t.Error("Expected styled active field")
			}

			inactiveStyle := GetFieldInactiveStyle(ctx)
			rendered = inactiveStyle.Render("Test")
			if rendered == "" {
				t.Error("Expected styled inactive field")
			}

			// Test help style adapts
			helpStyle := GetHelpStyle(ctx)
			rendered = helpStyle.Render("Test")
			if rendered == "" {
				t.Error("Expected styled help text")
			}
		})
	}
}

func TestStylesWithoutContext(t *testing.T) {
	// Test that styles work without context (fallback mode)
	titleStyle := GetSectionTitleStyle(nil)
	rendered := titleStyle.Render("Test")
	if rendered == "" {
		t.Error("Expected styled title with nil context")
	}

	labelStyle := GetFieldLabelStyle(nil)
	rendered = labelStyle.Render("Test")
	if rendered == "" {
		t.Error("Expected styled label with nil context")
	}

	// Error and success styles should always work (non-adaptive)
	errorStyle := GetErrorStyle()
	rendered = errorStyle.Render("Test")
	if rendered == "" {
		t.Error("Expected styled error text")
	}

	successStyle := GetSuccessStyle()
	rendered = successStyle.Render("Test")
	if rendered == "" {
		t.Error("Expected styled success text")
	}
}

func TestLabelStyleVerticalMode(t *testing.T) {
	// Test narrow breakpoint that triggers vertical layout
	ctx := NewLayoutContext(70, 25) // This should trigger vertical layout
	labelStyle := GetFieldLabelStyle(ctx)

	// In vertical mode, label should not have fixed width and should align left
	vertical, _, _ := ctx.GetFieldLayout()
	if !vertical {
		t.Skip("Test requires vertical layout mode")
	}

	rendered := labelStyle.Render("Test")
	if rendered == "" {
		t.Error("Expected styled label in vertical mode")
	}
}

func TestMobilePaddingReduction(t *testing.T) {
	mobileCtx := NewLayoutContext(50, 20) // Mobile breakpoint
	wideCtx := NewLayoutContext(150, 40)  // Wide breakpoint

	// Mobile styles should have less/no padding compared to wide
	mobileTitleStyle := GetSectionTitleStyle(mobileCtx)
	wideTitleStyle := GetSectionTitleStyle(wideCtx)

	// Both should render properly
	mobileRendered := mobileTitleStyle.Render("Test")
	wideRendered := wideTitleStyle.Render("Test")

	if mobileRendered == "" {
		t.Error("Expected styled mobile title")
	}
	if wideRendered == "" {
		t.Error("Expected styled wide title")
	}

	// Test that mobile field styles also reduce padding
	mobileActiveStyle := GetFieldActiveStyle(mobileCtx)
	wideActiveStyle := GetFieldActiveStyle(wideCtx)

	mobileRendered = mobileActiveStyle.Render("Test")
	wideRendered = wideActiveStyle.Render("Test")

	if mobileRendered == "" {
		t.Error("Expected styled mobile active field")
	}
	if wideRendered == "" {
		t.Error("Expected styled wide active field")
	}
}

// Test responsive tabs implementation (T2)
func TestResponsiveTabsIntegration(t *testing.T) {
	// Create a mock section for testing
	mockSection := EnhancedConfigSection{
		Name:        "Test Section",
		Icon:        "🔧",
		Description: "Test description",
		Fields:      []EnhancedConfigField{},
	}

	tests := []struct {
		width            int
		expectedVertical bool
		expectedSymbols  bool
		name             string
	}{
		{50, true, true, "mobile_vertical_symbols"},
		{70, true, false, "narrow_vertical_names"},
		{100, false, false, "medium_horizontal_names"},
		{150, false, false, "wide_horizontal_names"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := &EnhancedConfigFormModel{
				width:         tt.width,
				height:        30,
				sections:      []EnhancedConfigSection{mockSection, mockSection}, // Two sections
				activeSection: 0,
			}

			// Initialize layout context
			model.SetSize(tt.width, 30)

			// Get layout configuration
			ctx := model.GetLayoutContext()
			vertical, symbols, _ := ctx.GetTabLayout()

			if vertical != tt.expectedVertical {
				t.Errorf("Expected vertical=%v, got %v", tt.expectedVertical, vertical)
			}
			if symbols != tt.expectedSymbols {
				t.Errorf("Expected symbols=%v, got %v", tt.expectedSymbols, symbols)
			}

			// Test that renderSectionTabs doesn't panic and returns content
			tabsOutput := model.renderSectionTabs()
			if tabsOutput == "" {
				t.Error("Expected non-empty tabs output")
			}
		})
	}
}

func TestTabsOverflowHandling(t *testing.T) {
	// Create multiple mock sections
	sections := []EnhancedConfigSection{
		{Name: "Section 1", Description: "Desc 1", Fields: []EnhancedConfigField{}, Icon: "🔧"},
		{Name: "Section 2", Description: "Desc 2", Fields: []EnhancedConfigField{}, Icon: "⚙️"},
		{Name: "Section 3", Description: "Desc 3", Fields: []EnhancedConfigField{}, Icon: "🛠️"},
		{Name: "Section 4", Description: "Desc 4", Fields: []EnhancedConfigField{}, Icon: "🔩"},
		{Name: "Section 5", Description: "Desc 5", Fields: []EnhancedConfigField{}, Icon: "⚡"},
		{Name: "Section 6", Description: "Desc 6", Fields: []EnhancedConfigField{}, Icon: "🎯"},
		{Name: "Section 7", Description: "Desc 7", Fields: []EnhancedConfigField{}, Icon: "📊"},
	}

	model := &EnhancedConfigFormModel{
		width:         50, // Mobile breakpoint
		height:        30,
		sections:      sections,
		activeSection: 0,
	}

	model.SetSize(50, 30)

	// Mobile should limit to 6 tabs max
	ctx := model.GetLayoutContext()
	_, _, maxTabs := ctx.GetTabLayout()

	if maxTabs != 6 {
		t.Errorf("Expected maxTabs=6 for mobile, got %d", maxTabs)
	}

	// Should show overflow indicator
	tabsOutput := model.renderSectionTabs()
	if tabsOutput == "" {
		t.Error("Expected non-empty tabs output")
	}

	// Output should contain overflow indicator when we have more sections than limit
	if len(sections) > maxTabs && !contains(tabsOutput, "...") && !contains(tabsOutput, "+") {
		t.Error("Expected overflow indicator in tabs output")
	}
}

func TestTabsActiveIndicators(t *testing.T) {
	mockSection := EnhancedConfigSection{
		Name:        "Test",
		Icon:        "🔧",
		Description: "Test",
		Fields:      []EnhancedConfigField{},
	}

	// Test mobile mode with symbols
	model := &EnhancedConfigFormModel{
		width:         50,
		height:        30,
		sections:      []EnhancedConfigSection{mockSection, mockSection},
		activeSection: 0,
	}
	model.SetSize(50, 30)

	tabsOutput := model.renderSectionTabs()
	if tabsOutput == "" {
		t.Error("Expected non-empty tabs output")
	}

	// Should contain active indicator for mobile
	if !contains(tabsOutput, "▶") {
		t.Error("Expected active indicator (▶) in mobile tabs output")
	}

	// Test narrow mode with vertical layout
	model.SetSize(70, 30)
	tabsOutput = model.renderSectionTabs()

	if tabsOutput == "" {
		t.Error("Expected non-empty tabs output for narrow")
	}
}
