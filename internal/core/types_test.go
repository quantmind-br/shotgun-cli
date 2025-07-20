package core

import (
	"testing"
)

func TestSelectionState(t *testing.T) {
	ss := NewSelectionState()
	
	// Test initial state
	if !ss.IsFileIncluded("test.go") {
		t.Error("Files should be included by default")
	}
	
	// Test exclusion
	ss.ExcludeFile("test.go")
	if ss.IsFileIncluded("test.go") {
		t.Error("File should be excluded after ExcludeFile")
	}
	
	// Test inclusion
	ss.IncludeFile("test.go")
	if !ss.IsFileIncluded("test.go") {
		t.Error("File should be included after IncludeFile")
	}
	
	// Test toggle
	ss.ToggleFile("test.go")
	if ss.IsFileIncluded("test.go") {
		t.Error("File should be excluded after toggle")
	}
	
	ss.ToggleFile("test.go")
	if !ss.IsFileIncluded("test.go") {
		t.Error("File should be included after second toggle")
	}
	
	// Test reset
	ss.ExcludeFile("test1.go")
	ss.ExcludeFile("test2.go")
	ss.Reset()
	
	if !ss.IsFileIncluded("test1.go") || !ss.IsFileIncluded("test2.go") {
		t.Error("All files should be included after reset")
	}
}

func TestTemplateProcessor(t *testing.T) {
	tp := NewSimpleTemplateProcessor()
	
	// Test simple template
	tp.templates["test"] = "Hello {TASK}!"
	
	data := TemplateData{
		Task: "World",
	}
	
	result, err := tp.GeneratePrompt("test", data)
	if err != nil {
		t.Errorf("GeneratePrompt failed: %v", err)
	}
	
	expected := "Hello World!"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestTemplateInfo(t *testing.T) {
	// Test getting template info
	info, exists := GetTemplateInfo(TemplateDevKey)
	if !exists {
		t.Error("Dev template should exist")
	}
	
	if info.Key != TemplateDevKey {
		t.Errorf("Expected key %q, got %q", TemplateDevKey, info.Key)
	}
	
	// Test non-existent template
	_, exists = GetTemplateInfo("nonexistent")
	if exists {
		t.Error("Non-existent template should not exist")
	}
}