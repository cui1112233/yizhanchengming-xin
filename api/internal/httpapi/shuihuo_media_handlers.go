package httpapi

import (
	"encoding/json"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/shuihuo"
	"net/http"
	"strconv"
)

func scope(r *http.Request) (int64, int64, error) {
	p, e := strconv.ParseInt(r.PathValue("projectId"), 10, 64)
	if e != nil || p < 1 {
		return 0, 0, e
	}
	b, e := strconv.ParseInt(r.PathValue("bookId"), 10, 64)
	return p, b, e
}
func (h handler) shuihuoSegments(w http.ResponseWriter, r *http.Request) {
	if h.deps.ShuihuoMedia == nil {
		writeError(w, 503, "shuihuo media unavailable")
		return
	}
	p, b, e := scope(r)
	if e != nil || b < 1 {
		writeError(w, 400, "invalid scope")
		return
	}
	if r.Method == http.MethodGet {
		x, e := h.deps.ShuihuoMedia.ListSegments(r.Context(), p, b)
		if e != nil {
			writeError(w, 422, e.Error())
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
	x, e := h.deps.ShuihuoMedia.CreateSegment(r.Context(), shuihuo.CreateSegmentInput{BatchProjectID: p, BookID: b, Position: v.Position, Text: v.Text, EditRevision: v.EditRevision})
	if e != nil {
		writeError(w, 422, e.Error())
		return
	}
	writeJSON(w, 201, x)
}
func (h handler) shuihuoAssets(w http.ResponseWriter, r *http.Request) {
	if h.deps.ShuihuoMedia == nil {
		writeError(w, 503, "shuihuo media unavailable")
		return
	}
	p, b, e := scope(r)
	if e != nil || b < 1 {
		writeError(w, 400, "invalid scope")
		return
	}
	if r.Method == http.MethodGet {
		x, e := h.deps.ShuihuoMedia.ListAssets(r.Context(), p, b, 0)
		if e != nil {
			writeError(w, 422, e.Error())
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
	x, e := h.deps.ShuihuoMedia.CreateAsset(r.Context(), shuihuo.CreateAssetInput{BatchProjectID: p, BookID: b, SegmentID: v.SegmentID, Type: v.Type, Bucket: v.Bucket, ObjectKey: v.ObjectKey, Metadata: v.Metadata})
	if e != nil {
		writeError(w, 422, e.Error())
		return
	}
	writeJSON(w, 201, x)
}
func (h handler) shuihuoMediaTasks(w http.ResponseWriter, r *http.Request) {
	if h.deps.ShuihuoMedia == nil {
		writeError(w, 503, "shuihuo media unavailable")
		return
	}
	p, b, e := scope(r)
	if e != nil || b < 1 {
		writeError(w, 400, "invalid scope")
		return
	}
	if r.Method == http.MethodGet {
		x, e := h.deps.ShuihuoMedia.ListMediaTasks(r.Context(), p, b)
		if e != nil {
			writeError(w, 422, e.Error())
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
	x, e := h.deps.ShuihuoMedia.CreateMediaTask(r.Context(), shuihuo.CreateMediaTaskInput{BatchProjectID: p, BookID: b, SegmentID: v.SegmentID, SourceAssetID: v.SourceAssetID, ProductionTaskID: v.ProductionTaskID, Kind: v.Kind, Provider: v.Provider, Model: v.Model, RequestID: v.RequestID})
	if e != nil {
		writeError(w, 422, e.Error())
		return
	}
	writeJSON(w, 202, x)
}
