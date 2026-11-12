package httpapi

import (
	"context"
	"net/http"
	"qiantie/backend/internal/aicut"
	"strings"
	"time"
)

type aicutTimelineScopeKey struct{}

func aicutTimelineIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(aicutTimelineScopeKey{}).(string)
	if id == "" {
		return aicut.DefaultTimelineID
	}
	return id
}
func withAICutTimelineScope(next http.Handler, auth BridgeAuth) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSpace(r.Header.Get("X-AICut-Timeline-Id"))
		if id == "" {
			id = aicut.DefaultTimelineID
		}
		r = r.WithContext(context.WithValue(r.Context(), aicutTimelineScopeKey{}, id))
		const prefix = "/api/ai-cut/projects/"
		if id == aicut.DefaultTimelineID || !strings.HasPrefix(r.URL.Path, prefix) {
			next.ServeHTTP(w, r)
			return
		}
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, prefix), "/")
		suffix := parts[1:]
		allowed := false
		if len(suffix) == 0 {
			allowed = r.Method == http.MethodGet || r.Method == http.MethodPatch
		} else {
			switch suffix[0] {
			case "timelines":
				allowed = true
			case "asset-folders":
				allowed = true
			case "creative-canvas":
				allowed = r.Method == http.MethodPost && (len(suffix) == 1 || (len(suffix) == 2 && (suffix[1] == "assets" || suffix[1] == "script")))
			case "assets":
				allowed = (len(suffix) == 3 && suffix[2] == "folder" && r.Method == http.MethodPatch) || r.Method == http.MethodGet || (len(suffix) == 1 && r.Method == http.MethodPost) || (len(suffix) == 3 && suffix[2] == "proxy" && r.Method == http.MethodPost)
			case "export-jobs", "asr-jobs", "agent-runs", "edit-plans":
				allowed = r.Method == http.MethodGet || r.Method == http.MethodPost
			case "agent-messages":
				allowed = len(suffix) == 1 && r.Method == http.MethodGet
			case "audio-analysis":
				allowed = len(suffix) == 1 && r.Method == http.MethodGet
			case "smart-split", "extract-audio", "consolidate", "filler-trim", "oral-clean", "multicam-sync":
				allowed = len(suffix) == 1 && r.Method == http.MethodPost
			case "subtitles":
				allowed = len(suffix) == 2 && suffix[1] == "import" && r.Method == http.MethodPost
			case "captions":
				allowed = len(suffix) == 2 && suffix[1] == "correct" && r.Method == http.MethodPost
			case "visual-frames", "tracking-frames":
				allowed = len(suffix) == 1 && r.Method == http.MethodPost
			}
		}
		if allowed {
			next.ServeHTTP(w, r)
			return
		}
		auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "此工具暂不支持当前时间线，请切换到默认时间线后使用"})
		})).ServeHTTP(w, r)
	})
}

// This adapter is confined to export registration. Timeline selection lives in
// each request context; no request changes shared stores or service state.
type aicutExportTimelineStore struct {
	aicut.Store
	timelines *aicut.TimelineCollectionStore
}

func (s aicutExportTimelineStore) Timeline(ctx context.Context, o, p string) (aicut.Timeline, error) {
	id := aicutTimelineIDFromContext(ctx)
	if id == aicut.DefaultTimelineID {
		return s.Store.Timeline(ctx, o, p)
	}
	if s.timelines == nil {
		return aicut.Timeline{}, aicut.ErrTimelineJobUnsupported
	}
	return s.timelines.Timeline(ctx, o, p, id)
}

// Synchronous tools share project assets but read and save only the immutable
// timeline ID carried by this request. Agent and saved edit plans are excluded.
type aicutRequestTimelineStore struct{ aicutExportTimelineStore }

func (s aicutRequestTimelineStore) UpdateProjectRatio(ctx context.Context, o, p, ratio string, now time.Time) (aicut.Project, error) {
	updater, ok := s.Store.(interface {
		UpdateProjectRatio(context.Context, string, string, string, time.Time) (aicut.Project, error)
	})
	if !ok {
		return aicut.Project{}, aicut.ErrInvalidProject
	}
	return updater.UpdateProjectRatio(ctx, o, p, ratio, now)
}

func (s aicutRequestTimelineStore) SaveTimeline(ctx context.Context, o, p string, v aicut.Timeline, rev int64, now time.Time) (aicut.Timeline, error) {
	id := aicutTimelineIDFromContext(ctx)
	if id == aicut.DefaultTimelineID {
		return s.Store.SaveTimeline(ctx, o, p, v, rev, now)
	}
	if s.timelines == nil {
		return aicut.Timeline{}, aicut.ErrTimelineJobUnsupported
	}
	return s.timelines.Save(ctx, o, p, id, v, rev, now)
}
func (s aicutRequestTimelineStore) UpdateProjectAppearance(ctx context.Context, o, p, mode, color string, now time.Time) (aicut.Project, error) {
	updater, ok := s.Store.(interface {
		UpdateProjectAppearance(context.Context, string, string, string, string, time.Time) (aicut.Project, error)
	})
	if !ok {
		return aicut.Project{}, aicut.ErrInvalidProject
	}
	return updater.UpdateProjectAppearance(ctx, o, p, mode, color, now)
}
