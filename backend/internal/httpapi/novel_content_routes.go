package httpapi

import (
	"net/http"
	"net/url"
	"path/filepath"
	"qiantie/backend/internal/documenttext"
	"qiantie/backend/internal/novelcreation"
	"strings"
)

func registerNovelContentRoutes(handle func(string, func(http.ResponseWriter, *http.Request, string)), decode func(http.ResponseWriter, *http.Request, any) bool, s *novelcreation.Service) {
	result := func(w http.ResponseWriter, v any, e error) {
		if e != nil {
			writeNovelCreationError(w, e)
			return
		}
		writeJSON(w, http.StatusOK, v)
	}
	handle("GET /api/novel-creation/documents/licenses", func(w http.ResponseWriter, r *http.Request, o string) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte(documenttext.LicenseText()))
	})
	handle("GET /api/novel-creation/documents/availability", func(w http.ResponseWriter, r *http.Request, o string) { result(w, s.DocumentAvailability(), nil) })
	handle("GET /api/novel-creation/projects/{projectID}/documents", func(w http.ResponseWriter, r *http.Request, o string) {
		v, e := s.ListDocuments(r.Context(), o, r.PathValue("projectID"))
		result(w, map[string]any{"documents": v}, e)
	})
	handle("POST /api/novel-creation/projects/{projectID}/documents/paste", func(w http.ResponseWriter, r *http.Request, o string) {
		var in struct {
			Title   string `json:"title"`
			Content string `json:"content"`
		}
		if !decode(w, r, &in) {
			return
		}
		v, e := s.PasteDocument(r.Context(), o, r.PathValue("projectID"), in.Title, in.Content)
		result(w, map[string]any{"document": v}, e)
	})
	handle("POST /api/novel-creation/projects/{projectID}/documents/upload", func(w http.ResponseWriter, r *http.Request, o string) {
		r.Body = http.MaxBytesReader(w, r.Body, 8*documenttext.MaxFileBytes+(1<<20))
		if e := r.ParseMultipartForm(1 << 20); e != nil {
			if r.MultipartForm != nil {
				defer r.MultipartForm.RemoveAll()
			}
			writeJSON(w, 400, map[string]string{"error": "上传内容无效或超过限制（每批8个，每个20MiB）"})
			return
		}
		defer r.MultipartForm.RemoveAll()
		files := r.MultipartForm.File["files"]
		count := 0
		for _, list := range r.MultipartForm.File {
			count += len(list)
		}
		if len(files) < 1 || len(files) > 8 || count != len(files) {
			writeJSON(w, 400, map[string]string{"error": "请用files字段上传1到8个文件"})
			return
		}
		for _, f := range files {
			ext := strings.ToLower(filepath.Ext(f.Filename))
			if f.Size > documenttext.MaxFileBytes || (ext != ".txt" && ext != ".docx" && ext != ".pdf") {
				writeJSON(w, 400, map[string]string{"error": "仅支持20MiB以内的TXT、DOCX或PDF"})
				return
			}
		}
		jobs := []novelcreation.DocumentJob{}
		failures := []map[string]string{}
		for _, f := range files {
			reader, e := f.Open()
			if e != nil {
				failures = append(failures, map[string]string{"filename": f.Filename, "error": "无法读取上传文件"})
				continue
			}
			j, e := s.UploadDocument(r.Context(), o, r.PathValue("projectID"), f.Filename, reader)
			reader.Close()
			if j.ID != "" {
				jobs = append(jobs, j)
			}
			if e != nil {
				failures = append(failures, map[string]string{"filename": f.Filename, "error": e.Error()})
			}
		}
		code := http.StatusAccepted
		if len(failures) > 0 {
			code = http.StatusMultiStatus
		}
		writeJSON(w, code, map[string]any{"documents": jobs, "failures": failures})
	})
	handle("GET /api/novel-creation/projects/{projectID}/documents/{documentID}", func(w http.ResponseWriter, r *http.Request, o string) {
		v, e := s.GetDocument(r.Context(), o, r.PathValue("projectID"), r.PathValue("documentID"))
		result(w, map[string]any{"document": v}, e)
	})
	handle("POST /api/novel-creation/projects/{projectID}/documents/{documentID}/retry", func(w http.ResponseWriter, r *http.Request, o string) {
		v, e := s.RetryDocument(r.Context(), o, r.PathValue("projectID"), r.PathValue("documentID"))
		result(w, map[string]any{"document": v}, e)
	})
	handle("POST /api/novel-creation/projects/{projectID}/documents/{documentID}/selection", func(w http.ResponseWriter, r *http.Request, o string) {
		var in novelcreation.DocumentSelection
		if !decode(w, r, &in) {
			return
		}
		v, e := s.SelectDocument(r.Context(), o, r.PathValue("projectID"), r.PathValue("documentID"), in)
		result(w, map[string]any{"selection": v}, e)
	})
	handle("PUT /api/novel-creation/projects/{projectID}/documents/{documentID}/chapters", func(w http.ResponseWriter, r *http.Request, o string) {
		var in novelcreation.DocumentSegments
		if !decode(w, r, &in) {
			return
		}
		v, e := s.ResegmentDocument(r.Context(), o, r.PathValue("projectID"), r.PathValue("documentID"), in)
		result(w, map[string]any{"document": v}, e)
	})
	handle("DELETE /api/novel-creation/projects/{projectID}/documents/{documentID}", func(w http.ResponseWriter, r *http.Request, o string) {
		result(w, map[string]bool{"deleted": true}, s.DeleteDocument(r.Context(), o, r.PathValue("projectID"), r.PathValue("documentID")))
	})
	handle("GET /api/novel-creation/projects/{projectID}/styles", func(w http.ResponseWriter, r *http.Request, o string) {
		v, e := s.ListStyleSources(r.Context(), o, r.PathValue("projectID"))
		result(w, map[string]any{"styles": v}, e)
	})
	handle("POST /api/novel-creation/projects/{projectID}/styles", func(w http.ResponseWriter, r *http.Request, o string) {
		var in struct {
			DocumentID  string `json:"documentId"`
			Title       string `json:"title"`
			Description string `json:"description"`
		}
		if !decode(w, r, &in) {
			return
		}
		v, e := s.CreateStyleFromDocument(r.Context(), o, r.PathValue("projectID"), in.DocumentID, in.Title, in.Description)
		result(w, map[string]any{"style": v}, e)
	})
	handle("POST /api/novel-creation/projects/{projectID}/maintenance", func(w http.ResponseWriter, r *http.Request, o string) {
		var in novelcreation.MaintenanceInput
		if !decode(w, r, &in) {
			return
		}
		v, e := s.QueueMaintenance(r.Context(), o, r.PathValue("projectID"), in)
		if e != nil {
			writeNovelCreationError(w, e)
			return
		}
		writeJSON(w, http.StatusAccepted, map[string]any{"run": v})
	})
	handle("GET /api/novel-creation/projects/{projectID}/continuity", func(w http.ResponseWriter, r *http.Request, o string) {
		v, e := s.GetContinuityState(r.Context(), o, r.PathValue("projectID"))
		result(w, v, e)
	})
	handle("PATCH /api/novel-creation/projects/{projectID}", func(w http.ResponseWriter, r *http.Request, o string) {
		var in struct {
			Title string `json:"title"`
		}
		if !decode(w, r, &in) {
			return
		}
		v, e := s.RenameProject(r.Context(), o, r.PathValue("projectID"), in.Title)
		result(w, map[string]any{"project": v}, e)
	})
	archive := func(w http.ResponseWriter, r *http.Request, o string) {
		result(w, map[string]bool{"archived": true}, s.ArchiveProject(r.Context(), o, r.PathValue("projectID")))
	}
	handle("DELETE /api/novel-creation/projects/{projectID}", archive)
	handle("POST /api/novel-creation/projects/{projectID}/archive", archive)
	handle("POST /api/novel-creation/projects/{projectID}/restore", func(w http.ResponseWriter, r *http.Request, o string) {
		result(w, map[string]bool{"restored": true}, s.RestoreProject(r.Context(), o, r.PathValue("projectID")))
	})
	handle("PATCH /api/novel-creation/projects/{projectID}/sessions/{sessionID}", func(w http.ResponseWriter, r *http.Request, o string) {
		var in struct {
			Action string `json:"action"`
			Title  string `json:"title"`
		}
		if !decode(w, r, &in) {
			return
		}
		v, e := s.ManageSession(r.Context(), o, r.PathValue("projectID"), r.PathValue("sessionID"), in.Action, in.Title)
		result(w, map[string]any{"session": v}, e)
	})
	handle("GET /api/novel-creation/projects/{projectID}/chapters/{chapterID}/versions", func(w http.ResponseWriter, r *http.Request, o string) {
		v, e := s.ListChapterVersions(r.Context(), o, r.PathValue("projectID"), r.PathValue("chapterID"))
		result(w, map[string]any{"versions": v}, e)
	})
	handle("PUT /api/novel-creation/projects/{projectID}/chapters/{chapterID}", func(w http.ResponseWriter, r *http.Request, o string) {
		var in novelcreation.ChapterEdit
		if !decode(w, r, &in) {
			return
		}
		v, e := s.EditChapter(r.Context(), o, r.PathValue("projectID"), r.PathValue("chapterID"), in)
		result(w, map[string]any{"chapter": v}, e)
	})
	handle("POST /api/novel-creation/projects/{projectID}/chapters/{chapterID}/restore", func(w http.ResponseWriter, r *http.Request, o string) {
		var in struct {
			ExpectedVersion int `json:"expectedVersion"`
			Version         int `json:"version"`
		}
		if !decode(w, r, &in) {
			return
		}
		v, e := s.RestoreChapter(r.Context(), o, r.PathValue("projectID"), r.PathValue("chapterID"), in.ExpectedVersion, in.Version)
		result(w, map[string]any{"chapter": v}, e)
	})
	handle("POST /api/novel-creation/projects/{projectID}/exports", func(w http.ResponseWriter, r *http.Request, o string) {
		var in struct {
			Format string `json:"format"`
		}
		if !decode(w, r, &in) {
			return
		}
		v, e := s.PersistExport(r.Context(), o, r.PathValue("projectID"), in.Format)
		result(w, v, e)
	})
	download := func(w http.ResponseWriter, v novelcreation.ExportDocument, e error) {
		if e != nil {
			writeNovelCreationError(w, e)
			return
		}
		w.Header().Set("Content-Type", v.ContentType)
		w.Header().Set("Content-Disposition", "attachment; filename*=UTF-8''"+url.PathEscape(v.Filename))
		w.Write([]byte(v.Content))
	}
	handle("GET /api/novel-creation/projects/{projectID}/exports/{exportID}", func(w http.ResponseWriter, r *http.Request, o string) {
		v, e := s.DownloadExport(r.Context(), o, r.PathValue("projectID"), r.PathValue("exportID"))
		download(w, v, e)
	})
	handle("GET /api/novel-creation/runs/{runID}/short-result", func(w http.ResponseWriter, r *http.Request, o string) {
		v, e := s.GetShortResult(r.Context(), o, r.PathValue("runID"))
		result(w, map[string]any{"result": v}, e)
	})
	handle("GET /api/novel-creation/runs/{runID}/short-result/export", func(w http.ResponseWriter, r *http.Request, o string) {
		v, e := s.ExportShortResult(r.Context(), o, r.PathValue("runID"), r.URL.Query().Get("format"))
		download(w, v, e)
	})
}
