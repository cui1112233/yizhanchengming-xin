package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/shuihuo"
)

type ShuihuoService interface {
	CreateProject(context.Context, shuihuo.Actor, shuihuo.CreateProjectInput) (shuihuo.ReadModel, error)
	ListProjects(context.Context, shuihuo.Actor) ([]shuihuo.Project, error)
	GetProject(context.Context, shuihuo.Actor, int64) (shuihuo.ReadModel, error)
	ReplaceSource(context.Context, shuihuo.Actor, int64, string) (shuihuo.ReadModel, error)
	DeleteProject(context.Context, shuihuo.Actor, int64) error
	ParagraphSegmentation(context.Context, shuihuo.Actor, int64, shuihuo.SegmentationInput) ([]shuihuo.Candidate, error)
	FixedSegmentation(context.Context, shuihuo.Actor, int64, shuihuo.FixedSegmentationInput) ([]shuihuo.Candidate, error)
	ImportSegmentation(context.Context, shuihuo.Actor, int64, shuihuo.SegmentationInput) ([]shuihuo.Candidate, error)
	SmartSegmentation(context.Context, shuihuo.Actor, int64, shuihuo.SegmentationInput) ([]shuihuo.Candidate, error)
	ConfirmSegmentation(context.Context, shuihuo.Actor, int64, []shuihuo.Candidate) (shuihuo.ReadModel, error)
	CreateSegment(context.Context, shuihuo.Actor, int64, shuihuo.SegmentInput) (shuihuo.Segment, error)
	UpdateSegment(context.Context, shuihuo.Actor, int64, shuihuo.SegmentInput) (shuihuo.Segment, error)
	DeleteSegment(context.Context, shuihuo.Actor, int64) error
	ReorderSegments(context.Context, shuihuo.Actor, int64, []int64) (shuihuo.ReadModel, error)
}

type shuihuoSourceRequest struct {
	SourceText string `json:"sourceText"`
}

type shuihuoConfirmRequest struct {
	Candidates []shuihuo.Candidate `json:"candidates"`
}

type shuihuoReorderRequest struct {
	SegmentIDs []int64 `json:"segmentIds"`
}

func (h handler) shuihuoActor(r *http.Request) shuihuo.Actor {
	if user, ok := authn.CurrentUser(r.Context()); ok {
		role := strings.ToLower(strings.TrimSpace(user.Role))
		return shuihuo.Actor{
			UserID:          user.ID,
			TeamID:          user.TeamID,
			BypassOwnership: role == "admin" || role == "owner",
		}
	}
	if h.deps.Auth == nil {
		return shuihuo.Actor{BypassOwnership: true}
	}
	return shuihuo.Actor{}
}

func (h handler) requireShuihuoService(w http.ResponseWriter, r *http.Request) bool {
	if h.deps.Shuihuo != nil {
		return true
	}
	h.writeServiceError(w, r, http.StatusServiceUnavailable, "SHUIHUO_UNAVAILABLE", "水货生产服务暂不可用", "shuihuo", "service", nil)
	return false
}

func (h handler) writeShuihuoError(w http.ResponseWriter, r *http.Request, operation string, err error) {
	switch {
	case errors.Is(err, shuihuo.ErrInvalid):
		writeJSON(w, http.StatusBadRequest, map[string]any{"code": "SHUIHUO_INVALID", "message": "水货生产请求参数无效", "request_id": requestIDFromRequest(r)})
	case errors.Is(err, shuihuo.ErrNotFound):
		writeJSON(w, http.StatusNotFound, map[string]any{"code": "SHUIHUO_NOT_FOUND", "message": "水货项目或分镜不存在", "request_id": requestIDFromRequest(r)})
	case errors.Is(err, shuihuo.ErrForbidden):
		writeJSON(w, http.StatusForbidden, map[string]any{"code": "AUTH_FORBIDDEN", "message": "你没有访问该水货项目的权限", "request_id": requestIDFromRequest(r)})
	case errors.Is(err, shuihuo.ErrSmartUnavailable):
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"code": "SHUIHUO_SMART_UNAVAILABLE", "message": "智能分段所需文本模型暂不可用", "request_id": requestIDFromRequest(r)})
	default:
		h.writeServiceError(w, r, http.StatusInternalServerError, "SHUIHUO_OPERATION_FAILED", "水货生产操作失败", "shuihuo", operation, err)
	}
}

func (h handler) listShuihuoProjects(w http.ResponseWriter, r *http.Request) {
	if !h.requireShuihuoService(w, r) {
		return
	}
	rows, err := h.deps.Shuihuo.ListProjects(r.Context(), h.shuihuoActor(r))
	if err != nil {
		h.writeShuihuoError(w, r, "list_projects", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"projects": rows})
}

func (h handler) createShuihuoProject(w http.ResponseWriter, r *http.Request) {
	if !h.requireShuihuoService(w, r) {
		return
	}
	var input shuihuo.CreateProjectInput
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	result, err := h.deps.Shuihuo.CreateProject(r.Context(), h.shuihuoActor(r), input)
	if err != nil {
		h.writeShuihuoError(w, r, "create_project", err)
		return
	}
	writeJSON(w, http.StatusCreated, result.Project)
}

func (h handler) getShuihuoProject(w http.ResponseWriter, r *http.Request) {
	if !h.requireShuihuoService(w, r) {
		return
	}
	id, err := parsePositiveID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	result, err := h.deps.Shuihuo.GetProject(r.Context(), h.shuihuoActor(r), id)
	if err != nil {
		h.writeShuihuoError(w, r, "get_project", err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h handler) deleteShuihuoProject(w http.ResponseWriter, r *http.Request) {
	if !h.requireShuihuoService(w, r) {
		return
	}
	id, err := parsePositiveID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.deps.Shuihuo.DeleteProject(r.Context(), h.shuihuoActor(r), id); err != nil {
		h.writeShuihuoError(w, r, "delete_project", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h handler) replaceShuihuoSource(w http.ResponseWriter, r *http.Request) {
	if !h.requireShuihuoService(w, r) {
		return
	}
	id, err := parsePositiveID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	var input shuihuoSourceRequest
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	result, err := h.deps.Shuihuo.ReplaceSource(r.Context(), h.shuihuoActor(r), id, input.SourceText)
	if err != nil {
		h.writeShuihuoError(w, r, "replace_source", err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h handler) shuihuoParagraphSegmentation(w http.ResponseWriter, r *http.Request) {
	if !h.requireShuihuoService(w, r) {
		return
	}
	id, err := parsePositiveID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	var input shuihuo.SegmentationInput
	if r.ContentLength != 0 {
		if err := decodeJSON(w, r, &input); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	rows, err := h.deps.Shuihuo.ParagraphSegmentation(r.Context(), h.shuihuoActor(r), id, input)
	if err != nil {
		h.writeShuihuoError(w, r, "paragraph_segmentation", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"candidates": rows})
}

func (h handler) shuihuoFixedSegmentation(w http.ResponseWriter, r *http.Request) {
	if !h.requireShuihuoService(w, r) {
		return
	}
	id, err := parsePositiveID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	var input shuihuo.FixedSegmentationInput
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	rows, err := h.deps.Shuihuo.FixedSegmentation(r.Context(), h.shuihuoActor(r), id, input)
	if err != nil {
		h.writeShuihuoError(w, r, "fixed_segmentation", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"candidates": rows})
}

func (h handler) shuihuoImportSegmentation(w http.ResponseWriter, r *http.Request) {
	if !h.requireShuihuoService(w, r) {
		return
	}
	id, err := parsePositiveID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	var input shuihuo.SegmentationInput
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	rows, err := h.deps.Shuihuo.ImportSegmentation(r.Context(), h.shuihuoActor(r), id, input)
	if err != nil {
		h.writeShuihuoError(w, r, "import_segmentation", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"candidates": rows})
}

func (h handler) shuihuoSmartSegmentation(w http.ResponseWriter, r *http.Request) {
	if !h.requireShuihuoService(w, r) {
		return
	}
	id, err := parsePositiveID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	var input shuihuo.SegmentationInput
	if r.ContentLength != 0 {
		if err := decodeJSON(w, r, &input); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	rows, err := h.deps.Shuihuo.SmartSegmentation(r.Context(), h.shuihuoActor(r), id, input)
	if err != nil {
		h.writeShuihuoError(w, r, "smart_segmentation", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"candidates": rows})
}

func (h handler) confirmShuihuoSegmentation(w http.ResponseWriter, r *http.Request) {
	if !h.requireShuihuoService(w, r) {
		return
	}
	id, err := parsePositiveID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	var input shuihuoConfirmRequest
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	result, err := h.deps.Shuihuo.ConfirmSegmentation(r.Context(), h.shuihuoActor(r), id, input.Candidates)
	if err != nil {
		h.writeShuihuoError(w, r, "confirm_segmentation", err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h handler) createShuihuoSegment(w http.ResponseWriter, r *http.Request) {
	if !h.requireShuihuoService(w, r) {
		return
	}
	projectID, err := parsePositiveID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	var input shuihuo.SegmentInput
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	segment, err := h.deps.Shuihuo.CreateSegment(r.Context(), h.shuihuoActor(r), projectID, input)
	if err != nil {
		h.writeShuihuoError(w, r, "create_segment", err)
		return
	}
	writeJSON(w, http.StatusCreated, segment)
}

func (h handler) updateShuihuoSegment(w http.ResponseWriter, r *http.Request) {
	if !h.requireShuihuoService(w, r) {
		return
	}
	segmentID, err := parsePositiveID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	var input shuihuo.SegmentInput
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	segment, err := h.deps.Shuihuo.UpdateSegment(r.Context(), h.shuihuoActor(r), segmentID, input)
	if err != nil {
		h.writeShuihuoError(w, r, "update_segment", err)
		return
	}
	writeJSON(w, http.StatusOK, segment)
}

func (h handler) deleteShuihuoSegment(w http.ResponseWriter, r *http.Request) {
	if !h.requireShuihuoService(w, r) {
		return
	}
	segmentID, err := parsePositiveID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.deps.Shuihuo.DeleteSegment(r.Context(), h.shuihuoActor(r), segmentID); err != nil {
		h.writeShuihuoError(w, r, "delete_segment", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h handler) reorderShuihuoSegments(w http.ResponseWriter, r *http.Request) {
	if !h.requireShuihuoService(w, r) {
		return
	}
	projectID, err := parsePositiveID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	var input shuihuoReorderRequest
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	result, err := h.deps.Shuihuo.ReorderSegments(r.Context(), h.shuihuoActor(r), projectID, input.SegmentIDs)
	if err != nil {
		h.writeShuihuoError(w, r, "reorder_segments", err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}
