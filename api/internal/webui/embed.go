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
	if files, err := fs.Sub(dist, "dist/user"); err == nil {
		return files
	}
	// Keep source-checkout builds useful before the embed script has populated
	// the user/admin subdirectories. The release build always uses dist/user.
	files, err := fs.Sub(dist, "dist")
	if err != nil {
		panic("embedded user interface is unavailable: " + err.Error())
	}
	return files
}

// EmbeddedAdminFiles returns the separately built administrator application.
// A source checkout without a frontend build gets an empty FS and therefore a
// safe 503 from the HTTP handler instead of accidentally serving the user SPA.
func EmbeddedAdminFiles() fs.FS {
	files, err := fs.Sub(dist, "dist/admin")
	if err != nil {
		return emptyFS{}
	}
	return files
}

type emptyFS struct{}

func (emptyFS) Open(string) (fs.File, error) { return nil, fs.ErrNotExist }
