package httpapi

import (
	"encoding/json"
	"io"
	"net/http"

	"qiantie/backend/internal/novelcreation"
)

func registerNovelImportRoutes(handle func(string, func(http.ResponseWriter, *http.Request, string)), service *novelcreation.Service) {
	handle("GET /api/novel-creation/style-profiles", func(w http.ResponseWriter, r *http.Request, owner string) {
		styles, err := service.ListStyleProfiles(r.Context(), owner)
		if err != nil {
			writeNovelCreationError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"styles": styles})
	})
	handle("GET /api/novel-creation/projects/{projectID}/chapters", func(w http.ResponseWriter, r *http.Request, owner string) {
		chapters, err := service.ListProjectChapters(r.Context(), owner, r.PathValue("projectID"))
		if err != nil {
			writeNovelCreationError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"chapters": chapters})
	})
	handle("POST /api/novel-creation/chapter-imports", func(w http.ResponseWriter, r *http.Request, owner string) {
		var in novelcreation.ImportChaptersInput
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, novelCreationJSONLimit))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&in); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求内容无效，请检查后再试"})
			return
		}
		if err := decoder.Decode(new(any)); err != io.EOF {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求内容无效，请检查后再试"})
			return
		}
		out, err := service.ImportChapters(r.Context(), owner, in)
		if err != nil {
			writeNovelCreationError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, out)
	})
}
