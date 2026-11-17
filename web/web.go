// Package web serves the application's frontend from the binary.
package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var assets embed.FS

// FS returns the frontend tree rooted at the directory the pages are served from.
func FS() (fs.FS, error) {
	return fs.Sub(assets, "dist")
}
