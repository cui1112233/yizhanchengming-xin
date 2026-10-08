package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/unifiedsettings"
)

type UnifiedSettingsService interface {
	GetCurrent(context.Context, int64) (unifiedsettings.Current, error)
	SaveProduction(context.Context, int64, map[string]any) (unifiedsettings.Current, error)
	SavePublishing(context.Context, int64, map[string]any) (unifiedsettings.Current, error)
	SaveProfile(context.Context, int64, unifiedsettings.VersionProfile) (unifiedsettings.Current, error)
	Sync121(context.Context, int64) (unifiedsettings.Current, error)
	SyncStyleTypes(context.Context, int64) (unifiedsettings.Current, error)
}

func projectID(r *http.Request) (int64, error) { return strconv.ParseInt(r.PathValue("id"), 10, 64) }

func (h handler) getUnifiedSettings(w http.ResponseWriter, r *http.Request) {
	id, err := projectID(r); if err != nil { writeError(w, http.StatusBadRequest, "invalid batch project id"); return }
	if h.deps.UnifiedSettings == nil { writeError(w, http.StatusServiceUnavailable, "unified settings unavailable"); return }
	current, err := h.deps.UnifiedSettings.GetCurrent(r.Context(), id); if err != nil { h.writeServiceError(w, r, http.StatusInternalServerError, "SETTINGS_READ_FAILED", "读取统一配置失败", "unified_settings", "get", err); return }
	writeJSON(w, http.StatusOK, current)
}

func decodeMap(r *http.Request) (map[string]any, error) {
	var value map[string]any
	err := json.NewDecoder(r.Body).Decode(&value)
	return value, err
}

func (h handler) saveProductionSettings(w http.ResponseWriter, r *http.Request) {
	id, err := projectID(r); if err != nil { writeError(w, http.StatusBadRequest, "invalid batch project id"); return }
	value, err := decodeMap(r); if err != nil { writeError(w, http.StatusBadRequest, "invalid production settings"); return }
	current, err := h.deps.UnifiedSettings.SaveProduction(r.Context(), id, value); if err != nil { h.writeServiceError(w, r, http.StatusInternalServerError, "SETTINGS_SAVE_FAILED", "保存统一配置失败", "unified_settings", "save_production", err); return }
	writeJSON(w, http.StatusOK, current)
}

func (h handler) savePublishingSettings(w http.ResponseWriter, r *http.Request) {
	id, err := projectID(r); if err != nil { writeError(w, http.StatusBadRequest, "invalid batch project id"); return }
	value, err := decodeMap(r); if err != nil { writeError(w, http.StatusBadRequest, "invalid publishing settings"); return }
	current, err := h.deps.UnifiedSettings.SavePublishing(r.Context(), id, value); if err != nil { h.writeServiceError(w, r, http.StatusInternalServerError, "SETTINGS_SAVE_FAILED", "保存统一配置失败", "unified_settings", "save_publishing", err); return }
	writeJSON(w, http.StatusOK, current)
}

func (h handler) getVersionProfile(w http.ResponseWriter, r *http.Request) {
	id, err := projectID(r); if err != nil { writeError(w, http.StatusBadRequest, "invalid batch project id"); return }
	current, err := h.deps.UnifiedSettings.GetCurrent(r.Context(), id); if err != nil { h.writeServiceError(w, r, http.StatusInternalServerError, "SETTINGS_READ_FAILED", "读取统一配置失败", "unified_settings", "get_profile", err); return }
	writeJSON(w, http.StatusOK, map[string]any{"projectId": id, "profile": current.Profile})
}

func (h handler) saveVersionProfile(w http.ResponseWriter, r *http.Request) {
	id, err := projectID(r); if err != nil { writeError(w, http.StatusBadRequest, "invalid batch project id"); return }
	var profile unifiedsettings.VersionProfile
	if err := json.NewDecoder(r.Body).Decode(&profile); err != nil { writeError(w, http.StatusBadRequest, "invalid version profile"); return }
	current, err := h.deps.UnifiedSettings.SaveProfile(r.Context(), id, profile); if err != nil { h.writeServiceError(w, r, http.StatusInternalServerError, "SETTINGS_SAVE_FAILED", "保存统一配置失败", "unified_settings", "save_profile", err); return }
	writeJSON(w, http.StatusOK, current)
}

func (h handler) sync121Settings(w http.ResponseWriter, r *http.Request) {
	id, err := projectID(r); if err != nil { writeError(w, http.StatusBadRequest, "invalid batch project id"); return }
	current, err := h.deps.UnifiedSettings.Sync121(r.Context(), id); if err != nil { h.writeServiceError(w, r, http.StatusInternalServerError, "SETTINGS_SYNC_FAILED", "同步统一配置失败", "unified_settings", "sync_121", err); return }
	writeJSON(w, http.StatusOK, current)
}

func (h handler) syncStyleTypes(w http.ResponseWriter, r *http.Request) {
	id, err := projectID(r); if err != nil { writeError(w, http.StatusBadRequest, "invalid batch project id"); return }
	current, err := h.deps.UnifiedSettings.SyncStyleTypes(r.Context(), id); if err != nil { h.writeServiceError(w, r, http.StatusInternalServerError, "SETTINGS_SYNC_FAILED", "同步统一配置失败", "unified_settings", "sync_style_types", err); return }
	writeJSON(w, http.StatusOK, current)
}
