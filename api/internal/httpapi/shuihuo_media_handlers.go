package httpapi

import (
	"encoding/json"
	"errors"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/shuihuo"
	"net/http"
	"strconv"
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
	var v struct {
		SegmentID int64             `json:"segmentId"`
		Type      shuihuo.AssetType `json:"type"`
		Bucket    string            `json:"bucket"`
		ObjectKey string            `json:"objectKey"`
		Metadata  json.RawMessage   `json:"metadata"`
	}
	if json.NewDecoder(r.Body).Decode(&v) != nil {
		writeError(w, 400, "invalid JSON")
		return
	}
	x, e := s.CreateAsset(r.Context(), shuihuo.CreateAssetInput{BatchProjectID: p, BookID: b, SegmentID: v.SegmentID, Type: v.Type, Bucket: v.Bucket, ObjectKey: v.ObjectKey, Metadata: v.Metadata})
	if e != nil {
		swerr(w, e)
		return
	}
	writeJSON(w, 201, x)
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
