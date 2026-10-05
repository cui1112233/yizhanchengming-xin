package httpapi

import "net/http"

func (h handler) listBatchProjects(w http.ResponseWriter, _ *http.Request) {
	if h.deps.BatchProjects == nil {
		writeError(w, http.StatusServiceUnavailable, "batch project reader unavailable")
		return
	}
	writeError(w, http.StatusServiceUnavailable, "batch project list not implemented")
}
