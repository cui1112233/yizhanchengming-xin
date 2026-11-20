package httpapi

import (
	"encoding/json"
	"io"
	"net/http"

	"qiantie/backend/internal/novelcreation"
)

func registerNovelImportRoutes(handle func(string, func(http.ResponseWriter, *http.Request, string)), service *novelcreation.Service) {
	handle("POST /api/novel-creation/spinoffs", func(w http.ResponseWriter, r *http.Request, owner string) {
		var in novelcreation.SpinoffInput
		d := json.NewDecoder(http.MaxBytesReader(w, r.Body, novelCreationJSONLimit))
		d.DisallowUnknownFields()
		if e := d.Decode(&in); e != nil {
			writeJSON(w, 400, map[string]string{"error": "番外请求内容无效"})
			return
		}
		if e := d.Decode(new(any)); e != io.EOF {
			writeJSON(w, 400, map[string]string{"error": "番外请求内容无效"})
			return
		}
		out, e := service.StartSpinoff(r.Context(), owner, in)
		if e != nil {
			writeNovelCreationError(w, e)
			return
		}
		writeJSON(w, http.StatusCreated, out)
	})
	handle("GET /api/novel-creation/projects/{projectID}/spinoff", func(w http.ResponseWriter, r *http.Request, owner string) {
		out, e := service.GetSpinoff(r.Context(), owner, r.PathValue("projectID"))
		if e != nil {
			writeNovelCreationError(w, e)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"creation": out})
	})
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
