package web

import (
	"embed"
	"io/fs"
)

//go:embed dist/*
var distFS embed.FS

// GetFS returns the filesystem pointing to the embedded dist directory.
func GetFS() (fs.FS, error) {
	return fs.Sub(distFS, "dist")
}
