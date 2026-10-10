package webassets

import (
	"embed"
	"io/fs"
	"net/http"

	staticweb "qiantie/backend/internal/web"
)

// Dist contains the frontend build copied by scripts/build-go-local.ps1.
// Keeping the embed here avoids Go's restriction on embedding parent folders.
//
//go:embed dist/*
var Dist embed.FS

func Handler() http.Handler {
	files, err := fs.Sub(Dist, "dist")
	if err != nil {
		return staticweb.StaticHandler(nil)
	}
	return staticweb.StaticHandler(files)
}
