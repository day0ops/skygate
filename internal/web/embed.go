package web

import "embed"

// Static holds the bundled dashboard's HTML/CSS/JS under the "static"
// subdirectory, embedded at compile time.
//
//go:embed static
var Static embed.FS
