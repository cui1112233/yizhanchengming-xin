package httpapi

import (
	"encoding/json"
	"errors"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/shuihuo"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func scope(r *http.Request) (int64, int64, error) {
	p, e := strconv.ParseInt(r.PathValue("projectId"), 10, 64)
	if e != nil || p < 1 {
		return 0, 0, errors.New("invalid project")
	}
	b, e := strconv.ParseInt(r.PathValue("bookId"), 10, 64)
	if e != nil || b < 1 {
		return 0, 0, errors.New("invalid book")
	}
	return p, b, nil
}
func sid(r *http.Request, k string) (int64, error) {
	v, e := strconv.ParseInt(r.PathValue(k), 10, 64)
	if e != nil || v < 1 {
		return 0, errors.New("invalid " + k)
	}
	return v, nil
}
func swerr(w http.ResponseWriter, e error) {
	if errors.Is(e, shuihuo.ErrStorageUnavailable) {
		writeJSON(w, 503, map[string]string{"error": "storage_unavailable", "message": "TOS storage is not configured"})
		return
	}
	if errors.Is(e, shuihuo.ErrInvalidUpload) {
		writeJSON(w, 422, map[string]string{"error": "invalid_upload", "message": "file type, content, or size is not allowed"})
		return
	}
	if errors.Is(e, shuihuo.ErrConflict) {
		writeJSON(w, 409, map[string]string{"error": "version_conflict", "message": "segment changed; refresh then retry"})
		return
	}
	if errors.Is(e, shuihuo.ErrNotFound) {
		writeJSON(w, 404, map[string]string{"error": "not_found"})
		return
	}
	if errors.Is(e, shuihuo.ErrInvalidReorder) {
		writeJSON(w, 409, map[string]string{"error": "invalid_reorder"})
		return
	}
	writeError(w, 422, e.Error())
}
func (h handler) sm(w http.ResponseWriter) (ShuihuoMediaService, bool) {
	if h.deps.ShuihuoMedia == nil {
		writeError(w, 503, "shuihuo media unavailable")
		return nil, false
	}
	return h.deps.ShuihuoMedia, true
}
func (h handler) shuihuoSegments(w http.ResponseWriter, r *http.Request) {
	s, ok := h.sm(w)
	if !ok {
		return
	}
	p, b, e := scope(r)
	if e != nil {
		writeError(w, 400, e.Error())
		return
	}
	if r.Method == http.MethodGet {
		x, e := s.ListSegments(r.Context(), p, b)
		if e != nil {
			swerr(w, e)
			return
		}
		writeJSON(w, 200, x)
		return
	}
	var v struct {
		Position     int    `json:"position"`
		Text         string `json:"text"`
		EditRevision string `json:"editRevision"`
	}
	if json.NewDecoder(r.Body).Decode(&v) != nil {
		writeError(w, 400, "invalid JSON")
		return
	}
	x, e := s.CreateSegment(r.Context(), shuihuo.CreateSegmentInput{BatchProjectID: p, BookID: b, Position: v.Position, Text: v.Text, EditRevision: v.EditRevision})
	if e != nil {
		swerr(w, e)
		return
	}
	writeJSON(w, 201, x)
}
func (h handler) updateShuihuoSegment(w http.ResponseWriter, r *http.Request) {
	s, ok := h.sm(w)
	if !ok {
		return
	}
	p, b, e := scope(r)
	id, x := sid(r, "segmentId")
	if e != nil || x != nil {
		writeError(w, 400, "invalid scope or segment")
		return
	}
	var v struct {
		Text         string `json:"text"`
		EditRevision string `json:"editRevision"`
		Version      int    `json:"version"`
	}
	if json.NewDecoder(r.Body).Decode(&v) != nil {
		writeError(w, 400, "invalid JSON")
		return
	}
	out, e := s.UpdateSegment(r.Context(), p, b, id, shuihuo.UpdateSegmentInput{Text: v.Text, EditRevision: v.EditRevision, Version: v.Version})
	if e != nil {
		swerr(w, e)
		return
	}
	writeJSON(w, 200, out)
}
func (h handler) reorderShuihuoSegments(w http.ResponseWriter, r *http.Request) {
	s, ok := h.sm(w)
	if !ok {
		return
	}
	p, b, e := scope(r)
	if e != nil {
		writeError(w, 400, e.Error())
		return
	}
	var v struct {
		SegmentIDs []int64 `json:"segmentIds"`
	}
	if json.NewDecoder(r.Body).Decode(&v) != nil {
		writeError(w, 400, "invalid JSON")
		return
	}
	out, e := s.ReorderSegments(r.Context(), p, b, v.SegmentIDs)
	if e != nil {
		swerr(w, e)
		return
	}
	writeJSON(w, 200, out)
}
func (h handler) shuihuoAssets(w http.ResponseWriter, r *http.Request) {
	s, ok := h.sm(w)
	if !ok {
		return
	}
	p, b, e := scope(r)
	if e != nil {
		writeError(w, 400, e.Error())
		return
	}
	if r.Method == http.MethodGet {
		x, e := s.ListAssets(r.Context(), p, b, 0)
		if e != nil {
			swerr(w, e)
			return
		}
		writeJSON(w, 200, x)
		return
	}
	writeError(w, http.StatusMethodNotAllowed, "use the controlled multipart upload endpoint")
}
func (h handler) uploadShuihuoAsset(w http.ResponseWriter, r *http.Request) {
	s, ok := h.sm(w)
	if !ok {
		return
	}
	p, b, err := scope(r)
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	// The cap is a request-level guard; type-specific limits are applied by the
	// service while copying to an untrusted temporary file.
	r.Body = http.MaxBytesReader(w, r.Body, (2<<30)+1024)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeError(w, 400, "invalid multipart upload")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, 400, "file is required")
		return
	}
	defer file.Close()
	typeValue := shuihuo.AssetType(strings.TrimSpace(r.FormValue("type")))
	segmentID := int64(0)
	if raw := strings.TrimSpace(r.FormValue("segmentId")); raw != "" {
		segmentID, err = strconv.ParseInt(raw, 10, 64)
		if err != nil || segmentID < 1 {
			writeError(w, 400, "invalid segmentId")
			return
		}
	}
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	asset, err := s.UploadAsset(r.Context(), shuihuo.UploadAssetInput{BatchProjectID: p, BookID: b, SegmentID: segmentID, Type: typeValue, Filename: header.Filename, ContentType: contentType, Body: file})
	if err != nil {
		swerr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, asset)
}
func (h handler) readShuihuoAsset(w http.ResponseWriter, r *http.Request) {
	s, ok := h.sm(w)
	if !ok {
		return
	}
	p, b, err := scope(r)
	id, idErr := sid(r, "assetId")
	if err != nil || idErr != nil {
		writeError(w, 400, "invalid asset scope")
		return
	}
	asset, body, err := s.OpenAsset(r.Context(), p, b, id)
	if err != nil {
		swerr(w, err)
		return
	}
	defer body.Close()
	contentType := "application/octet-stream"
	var metadata struct {
		ContentType string `json:"contentType"`
	}
	_ = json.Unmarshal(asset.Metadata, &metadata)
	if metadata.ContentType != "" {
		contentType = metadata.ContentType
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	_, _ = io.Copy(w, body)
}
func (h handler) shuihuoMediaTasks(w http.ResponseWriter, r *http.Request) {
	s, ok := h.sm(w)
	if !ok {
		return
	}
	p, b, e := scope(r)
	if e != nil {
		writeError(w, 400, e.Error())
		return
	}
	if r.Method == http.MethodGet {
		x, e := s.ListMediaTasks(r.Context(), p, b)
		if e != nil {
			swerr(w, e)
			return
		}
		writeJSON(w, 200, x)
		return
	}
	var v struct {
		SegmentID        int64             `json:"segmentId"`
		SourceAssetID    int64             `json:"sourceAssetId"`
		ProductionTaskID int64             `json:"productionTaskId"`
		Kind             shuihuo.MediaKind `json:"kind"`
		Provider         string            `json:"provider"`
		Model            string            `json:"model"`
		RequestID        string            `json:"requestId"`
	}
	if json.NewDecoder(r.Body).Decode(&v) != nil {
		writeError(w, 400, "invalid JSON")
		return
	}
	x, e := s.CreateMediaTask(r.Context(), shuihuo.CreateMediaTaskInput{BatchProjectID: p, BookID: b, SegmentID: v.SegmentID, SourceAssetID: v.SourceAssetID, ProductionTaskID: v.ProductionTaskID, Kind: v.Kind, Provider: v.Provider, Model: v.Model, RequestID: v.RequestID})
	if e != nil {
		swerr(w, e)
		return
	}
	writeJSON(w, 202, x)
}
func (h handler) listShuihuoCandidates(w http.ResponseWriter, r *http.Request) {
	s, ok := h.sm(w)
	if !ok {
		return
	}
	p, b, e := scope(r)
	task, x := sid(r, "taskId")
	if e != nil || x != nil {
		writeError(w, 400, "invalid scope or task")
		return
	}
	out, e := s.ListCandidates(r.Context(), p, b, task)
	if e != nil {
		swerr(w, e)
		return
	}
	writeJSON(w, 200, out)
}
func (h handler) selectShuihuoCandidate(w http.ResponseWriter, r *http.Request) {
	s, ok := h.sm(w)
	if !ok {
		return
	}
	p, b, e := scope(r)
	task, x := sid(r, "taskId")
	candidate, y := sid(r, "candidateId")
	if e != nil || x != nil || y != nil {
		writeError(w, 400, "invalid scope or candidate")
		return
	}
	out, e := s.SelectCandidate(r.Context(), p, b, task, candidate)
	if e != nil {
		swerr(w, e)
		return
	}
	writeJSON(w, 200, out)
}
func (h handler) retryShuihuoMediaTask(w http.ResponseWriter, r *http.Request) {
	s, ok := h.sm(w)
	if !ok {
		return
	}
	p, b, e := scope(r)
	task, x := sid(r, "taskId")
	if e != nil || x != nil {
		writeError(w, 400, "invalid scope or task")
		return
	}
	out, e := s.RetryMediaTask(r.Context(), p, b, task)
	if e != nil {
		swerr(w, e)
		return
	}
	writeJSON(w, 202, out)
}
