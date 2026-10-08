package webui

import (
	"embed"
	"io/fs"
)

// dist is populated by scripts/build-embedded-ui.sh before the server binary is
// compiled. placeholder.txt keeps normal Go test/build commands compilable in a
// source checkout before a frontend build has happened.
//
//go:embed all:dist
var dist embed.FS

func EmbeddedFiles() fs.FS {
	files, err := fs.Sub(dist, "dist")
	if err != nil {
		panic("embedded user interface is unavailable: " + err.Error())
	}
	return files
}
