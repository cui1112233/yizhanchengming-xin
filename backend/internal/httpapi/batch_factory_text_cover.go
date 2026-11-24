package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	txt "qiantie/backend/internal/batchfactorytext"
)

// Cards always have a readable cover, including books not yet produced. This
// read-only route renders the book title without creating pages or calling AI.
func (a textAPI) displayCover(w http.ResponseWriter, r *http.Request, x textRequest) error {
	d, err := a.repo.GetDocument(x, x.scope)
	if err != nil && !errors.Is(err, txt.ErrNotFound) {
		return err
	}
	if d.CoverLocked && d.CoverArtifactID != "" {
		r.SetPathValue("artifactId", d.CoverArtifactID)
		return a.content(w, r, x)
	}
	snapshot, err := bookTextSnapshot(x)
	if err != nil {
		return err
	}
	d.ConfigSnapshot, err = json.Marshal(snapshot)
	if err != nil {
		return err
	}
	input, renderer, err := a.renderInput(x, d, txt.Page{}, true, txt.PagePatch{})
	if err != nil {
		return err
	}
	data, mime, err := renderer.Render(x, input)
	if err != nil {
		return textLayoutError(err)
	}
	w.Header().Set("Content-Type", mime)
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	_, err = w.Write(data)
	return err
}
