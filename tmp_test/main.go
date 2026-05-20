package main

import (
	"fmt"
	"strings"

	"github.com/quantmind-br/shotgun-cli/internal/core/scanner"
	"github.com/quantmind-br/shotgun-cli/internal/ui/screens"
)

func main() {
	root := &scanner.FileNode{
		Path:    "/tmp/test",
		Name:    "test",
		IsDir:   true,
		Children: []*scanner.FileNode{
			{Path: "/tmp/test/a.go", Name: "a.go", Size: 100},
			{Path: "/tmp/test/b.go", Name: "b.go", Size: 200},
		},
	}

	m := screens.NewFileSelection(root, nil, "10MB")
	m.SetSize(80, 24)
	view := m.View()
	
	lines := strings.Split(view, "\n")
	for i, line := range lines {
		fmt.Printf("%2d: %q\n", i, line)
	}
	fmt.Printf("Total lines: %d\n", len(lines))
}
