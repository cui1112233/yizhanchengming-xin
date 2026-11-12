package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"qiantie/backend/internal/aicut"
)

const aicutJSONLimit = 2 << 20

func RegisterAICutRoutes(mux *http.ServeMux, auth BridgeAuth, store aicut.Store, now func() time.Time) {
	if mux == nil || store == nil {
		return
	}
	if now == nil {
		now = time.Now
	}
	handle := func(pattern string, fn func(http.ResponseWriter, *http.Request, string)) {
		mux.Handle(pattern, auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			identity, ok := BridgeIdentityFromContext(req.Context())
			if !ok || strings.TrimSpace(identity.Username) == "" {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}
			fn(w, req, strings.TrimSpace(identity.Username))
		})))
	}

	handle("GET /api/ai-cut/projects", func(w http.ResponseWriter, req *http.Request, owner string) {
		projects, err := store.ListProjects(req.Context(), owner)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "剪辑项目读取失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"projects": projects})
	})
	handle("POST /api/ai-cut/projects", func(w http.ResponseWriter, req *http.Request, owner string) {
		var input struct {
			Title       string `json:"title"`
			AspectRatio string `json:"aspectRatio"`
		}
		if !decodeAICutJSON(w, req, &input) {
			return
		}
		createdAt := now().UTC()
		project := aicut.Project{ID: newAgentID("aicut"), Title: input.Title, AspectRatio: input.AspectRatio, CreatedAt: createdAt, UpdatedAt: createdAt}
		if err := store.CreateProject(req.Context(), owner, project); err != nil {
			writeAICutError(w, err)
			return
		}
		project, _ = store.Project(req.Context(), owner, project.ID)
		writeJSON(w, http.StatusCreated, map[string]any{"project": project})
	})
	handle("GET /api/ai-cut/projects/{id}", func(w http.ResponseWriter, req *http.Request, owner string) {
		project, err := store.Project(req.Context(), owner, req.PathValue("id"))
		if err != nil {
			writeAICutError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"project": project})
	})
	handle("PATCH /api/ai-cut/projects/{id}", func(w http.ResponseWriter, req *http.Request, owner string) {
		var input struct {
			Title           *string `json:"title"`
			BackgroundMode  *string `json:"backgroundMode"`
			BackgroundColor *string `json:"backgroundColor"`
			AspectRatio     *string `json:"aspectRatio"`
		}
		if !decodeAICutJSON(w, req, &input) {
			return
		}
		if input.Title == nil && input.BackgroundMode == nil && input.BackgroundColor == nil && input.AspectRatio == nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "没有可更新的项目设置"})
			return
		}
		if input.Title != nil && strings.TrimSpace(*input.Title) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "项目名称不能为空"})
			return
		}
		project, err := store.Project(req.Context(), owner, req.PathValue("id"))
		if input.AspectRatio != nil && *input.AspectRatio != "9:16" && *input.AspectRatio != "16:9" && *input.AspectRatio != "1:1" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "画面比例仅支持 9:16、16:9 或 1:1"})
			return
		}
		if err == nil && input.AspectRatio != nil {
			updater, ok := store.(interface {
				UpdateProjectRatio(context.Context, string, string, string, time.Time) (aicut.Project, error)
			})
			if !ok {
				writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "画面比例设置暂不可用"})
				return
			}
			project, err = updater.UpdateProjectRatio(req.Context(), owner, req.PathValue("id"), *input.AspectRatio, now().UTC())
		}
		if err == nil && input.Title != nil {
			project, err = store.RenameProject(req.Context(), owner, req.PathValue("id"), *input.Title, now().UTC())
		}
		if err == nil && (input.BackgroundMode != nil || input.BackgroundColor != nil) {
			updater, ok := store.(interface {
				UpdateProjectAppearance(context.Context, string, string, string, string, time.Time) (aicut.Project, error)
			})
			if !ok {
				writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "画布背景设置暂不可用"})
				return
			}
			mode, color := project.BackgroundMode, project.BackgroundColor
			if input.BackgroundMode != nil {
				mode = *input.BackgroundMode
			}
			if input.BackgroundColor != nil {
				color = *input.BackgroundColor
			}
			project, err = updater.UpdateProjectAppearance(req.Context(), owner, req.PathValue("id"), mode, color, now().UTC())
			if errors.Is(err, aicut.ErrInvalidProject) {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "画布背景仅支持背景色或透明，颜色必须为 #RRGGBB"})
				return
			}
		}
		if err != nil {
			writeAICutError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"project": project})
	})
	handle("DELETE /api/ai-cut/projects/{id}", func(w http.ResponseWriter, req *http.Request, owner string) {
		if err := store.DeleteProject(req.Context(), owner, req.PathValue("id"), now().UTC()); err != nil {
			writeAICutError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	handle("GET /api/ai-cut/projects/{id}/timeline", func(w http.ResponseWriter, req *http.Request, owner string) {
		timeline, err := store.Timeline(req.Context(), owner, req.PathValue("id"))
		if err != nil {
			writeAICutError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"timeline": timeline})
	})
	handle("PUT /api/ai-cut/projects/{id}/timeline", func(w http.ResponseWriter, req *http.Request, owner string) {
		var input struct {
			ExpectedRevision int64          `json:"expectedRevision"`
			Timeline         aicut.Timeline `json:"timeline"`
		}
		if !decodeAICutJSON(w, req, &input) {
			return
		}
		timeline, err := store.SaveTimeline(req.Context(), owner, req.PathValue("id"), input.Timeline, input.ExpectedRevision, now().UTC())
		if err != nil {
			writeAICutError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"timeline": timeline})
	})
	handle("GET /api/ai-cut/projects/{id}/timeline/history", func(w http.ResponseWriter, req *http.Request, owner string) {
		history, err := store.TimelineHistory(req.Context(), owner, req.PathValue("id"))
		if err != nil {
			writeAICutError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"history": history})
	})
	moveHistory := func(w http.ResponseWriter, req *http.Request, owner string, delta int) {
		var input struct {
			ExpectedRevision int64 `json:"expectedRevision"`
		}
		if !decodeAICutJSON(w, req, &input) {
			return
		}
		var timeline aicut.Timeline
		var history aicut.TimelineHistoryState
		var err error
		if delta < 0 {
			timeline, history, err = store.UndoTimeline(req.Context(), owner, req.PathValue("id"), input.ExpectedRevision, now().UTC())
		} else {
			timeline, history, err = store.RedoTimeline(req.Context(), owner, req.PathValue("id"), input.ExpectedRevision, now().UTC())
		}
		if err != nil {
			writeAICutError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"timeline": timeline, "history": history})
	}
	handle("POST /api/ai-cut/projects/{id}/timeline/history/undo", func(w http.ResponseWriter, req *http.Request, owner string) {
		moveHistory(w, req, owner, -1)
	})
	handle("POST /api/ai-cut/projects/{id}/timeline/history/redo", func(w http.ResponseWriter, req *http.Request, owner string) {
		moveHistory(w, req, owner, 1)
	})
	handle("POST /api/ai-cut/projects/{id}/subtitles/import", func(w http.ResponseWriter, req *http.Request, owner string) {
		var input struct {
			Filename         string `json:"filename"`
			SRT              string `json:"srt"`
			ExpectedRevision int64  `json:"expectedRevision"`
		}
		if !decodeAICutJSON(w, req, &input) {
			return
		}
		cues, err := aicut.ParseSRT(input.SRT)
		if err != nil {
			writeAICutError(w, err)
			return
		}
		current, err := store.Timeline(req.Context(), owner, req.PathValue("id"))
		if err != nil {
			writeAICutError(w, err)
			return
		}
		if current.Revision != input.ExpectedRevision {
			writeAICutError(w, aicut.ErrRevisionConflict)
			return
		}
		prepared, trackID, err := aicut.ApplySRTCues(current, input.Filename, cues, newAgentID)
		if err != nil {
			writeAICutError(w, err)
			return
		}
		timeline, err := store.SaveTimeline(req.Context(), owner, req.PathValue("id"), prepared, input.ExpectedRevision, now().UTC())
		if err != nil {
			writeAICutError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"timeline": timeline, "trackId": trackID, "imported": len(cues)})
	})
}

func decodeAICutJSON(w http.ResponseWriter, req *http.Request, target any) bool {
	decoder := json.NewDecoder(io.LimitReader(req.Body, aicutJSONLimit))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求内容无效"})
		return false
	}
	return true
}

func writeAICutError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, aicut.ErrTimelineNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "时间线不存在"})
	case errors.Is(err, aicut.ErrProjectNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "剪辑项目不存在"})
	case errors.Is(err, aicut.ErrProjectExists):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "剪辑项目已存在"})
	case errors.Is(err, aicut.ErrRevisionConflict):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "时间线已更新，请刷新后重试"})
	case errors.Is(err, aicut.ErrTimelineHistoryBoundary):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "没有可撤销或重做的操作"})
	case errors.Is(err, aicut.ErrInvalidProject):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "剪辑项目信息无效"})
	case errors.Is(err, aicut.ErrInvalidTimeline):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "时间线内容无效"})
	case errors.Is(err, aicut.ErrInvalidSRT):
		message := strings.TrimSpace(strings.TrimPrefix(err.Error(), aicut.ErrInvalidSRT.Error()+":"))
		if message == "" {
			message = "字幕文件内容无效"
		}
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": message})
	case errors.Is(err, aicut.ErrAssetNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "素材不存在"})
	case errors.Is(err, aicut.ErrAssetExists):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "素材已存在"})
	case errors.Is(err, aicut.ErrInvalidAsset):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "素材信息无效"})
	case errors.Is(err, aicut.ErrASRJobNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "语音识别任务不存在"})
	case errors.Is(err, aicut.ErrInvalidASRJob):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "语音识别任务信息无效"})
	case errors.Is(err, aicut.ErrAgentRunNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "AI-CUT Agent 任务不存在"})
	case errors.Is(err, aicut.ErrInvalidAgentRun), errors.Is(err, aicut.ErrInvalidAgentTransition):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "AI-CUT Agent 任务状态无效"})
	default:
		log.Printf("AI-CUT request failed: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "AI-CUT 保存失败"})
	}
}
