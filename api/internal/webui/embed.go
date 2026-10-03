package webui

import (
	"embed"
	"io/fs"
	"net/http"
)

// embeddedDist is replaced by the real React build output in CI before the
// production server binary is compiled.
//
//go:embed all:dist
var embeddedDist embed.FS

func Handler() (http.Handler, error) {
	root, err := fs.Sub(embeddedDist, "dist")
	if err != nil {
		return nil, err
	}
	return newSPAHandler(root)
}
