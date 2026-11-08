package httpapi

import (
	"net/http"

	"qiantie/backend/internal/releasenotes"
)

// registerProductUpdateRoute exposes the read-only changelog for the release
// this binary actually belongs to. It is intentionally unauthenticated: the
// notes are public product information and the shell must be able to show them
// before a session is restored.
//
// A missing or mismatched release returns 404 so the frontend stays silent
// instead of showing another release's notes.
func registerProductUpdateRoute(mux *http.ServeMux, runtimeManifestPath string, catalog *releasenotes.Catalog) {
	if mux == nil || catalog == nil {
		return
	}
	mux.HandleFunc("GET /api/product-updates/latest", func(w http.ResponseWriter, _ *http.Request) {
		releaseID := releasenotes.CurrentReleaseID(runtimeManifestPath)
		note, ok := catalog.ForReleaseID(releaseID)
		if !ok {
			// Do not cache a negative answer: once the new binary replaces the
			// old one, the same URL must be able to return the new notes.
			w.Header().Set("Cache-Control", "no-store")
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "暂无更新说明"})
			return
		}
		// Revalidate on every request instead of a long max-age so a release
		// never leaves browsers showing the previous release's changelog.
		w.Header().Set("Cache-Control", "no-cache, must-revalidate")
		writeJSON(w, http.StatusOK, note)
	})
}
