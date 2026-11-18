package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"qiantie/backend/internal/novelcreation"
)

const longImitationJSONLimit = 16 << 20

func decodeLongImitation(w http.ResponseWriter, r *http.Request, v any) bool {
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, longImitationJSONLimit))
	d.DisallowUnknownFields()
	if e := d.Decode(v); e != nil {
		writeJSON(w, 400, map[string]string{"error": "请求内容无效或超过16MiB"})
		return false
	}
	if d.Decode(new(any)) != io.EOF {
		writeJSON(w, 400, map[string]string{"error": "请求只能包含一个JSON对象"})
		return false
	}
	return true
}
func registerNovelLongImitationRoutes(handle func(string, func(http.ResponseWriter, *http.Request, string)), s *novelcreation.Service, starter NovelCreationRunStarter) {
	handle("POST /api/novel-creation/long-imitation", func(w http.ResponseWriter, r *http.Request, owner string) {
		var in novelcreation.LongImitationInput
		if !decodeLongImitation(w, r, &in) {
			return
		}
		out, e := s.StartLongImitation(r.Context(), owner, in)
		if e != nil {
			writeNovelCreationError(w, e)
			return
		}
		if starter != nil {
			starter(context.Background(), owner, out.Run.ID)
		}
		writeJSON(w, http.StatusAccepted, out)
	})
	handle("GET /api/novel-creation/long-imitation/requests/{key}", func(w http.ResponseWriter, r *http.Request, owner string) {
		out, e := s.GetLongImitationRequest(r.Context(), owner, r.PathValue("key"))
		if e != nil {
			writeNovelCreationError(w, e)
			return
		}
		writeJSON(w, 200, out)
	})
	handle("GET /api/novel-creation/projects/{projectID}/long-imitation", func(w http.ResponseWriter, r *http.Request, owner string) {
		out, e := s.GetLongImitation(r.Context(), owner, r.PathValue("projectID"))
		if e != nil {
			writeNovelCreationError(w, e)
			return
		}
		if r.URL.Query().Get("summary") == "1" {
			writeJSON(w, 200, map[string]any{"creation": novelcreation.LongImitationSummary(out)})
			return
		}
		writeJSON(w, 200, map[string]any{"creation": out})
	})
	handle("GET /api/novel-creation/long-imitation/memos", func(w http.ResponseWriter, r *http.Request, owner string) {
		out, truncated, e := s.LongImitationMemos(r.Context(), owner)
		if e != nil {
			writeNovelCreationError(w, e)
			return
		}
		writeJSON(w, 200, map[string]any{"memos": out, "truncated": truncated, "limit": 500})
	})
	handle("POST /api/novel-creation/long-imitation/preview", func(w http.ResponseWriter, r *http.Request, owner string) {
		var in struct {
			ReferenceText string `json:"referenceText"`
		}
		if !decodeLongImitation(w, r, &in) {
			return
		}
		out, e := s.PreviewLongImitation(r.Context(), owner, in.ReferenceText)
		if e != nil {
			writeNovelCreationError(w, e)
			return
		}
		writeJSON(w, 200, out)
	})
	handle("POST /api/novel-creation/long-imitation/parse", func(w http.ResponseWriter, r *http.Request, owner string) {
		r.Body = http.MaxBytesReader(w, r.Body, 161<<20)
		reader, e := r.MultipartReader()
		if e != nil {
			writeJSON(w, 400, map[string]string{"error": "请以multipart上传TXT或DOCX"})
			return
		}
		type file struct {
			Filename string                          `json:"filename"`
			Chapters []novelcreation.MaterialChapter `json:"chapters"`
		}
		type failure struct {
			Filename string `json:"filename"`
			Error    string `json:"error"`
		}
		files := []file{}
		failures := []failure{}
		count := 0
		for {
			part, e := reader.NextPart()
			if e == io.EOF {
				break
			}
			if e != nil {
				writeJSON(w, 400, map[string]string{"error": "上传内容无效或超过总大小限制"})
				return
			}
			count++
			if count > 8 || part.FormName() != "files" || part.FileName() == "" {
				part.Close()
				writeJSON(w, 400, map[string]string{"error": "最多上传8个文件，字段名须为files"})
				return
			}
			name := filepath.Base(part.FileName())
			data, readErr := io.ReadAll(io.LimitReader(part, (20<<20)+1))
			part.Close()
			if readErr != nil {
				failures = append(failures, failure{name, "读取文件失败"})
				continue
			}
			chapters, parseErr := s.ParseLongImitationReference(r.Context(), owner, name, data)
			if parseErr != nil {
				failures = append(failures, failure{name, parseErr.Error()})
				continue
			}
			files = append(files, file{name, chapters})
		}
		if count == 0 {
			writeJSON(w, 400, map[string]string{"error": "请至少选择一个文件"})
			return
		}
		writeJSON(w, 200, map[string]any{"files": files, "failures": failures})
	})
}
