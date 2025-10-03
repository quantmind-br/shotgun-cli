package assets

import "embed"

// Templates contains the embedded template files
//
//go:embed templates/*.md
var Templates embed.FS
