package httpapi

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
	"net/http"
)

func (h handler) workspaceSettings(w http.ResponseWriter, r *http.Request) {
	u, ok := authn.CurrentUser(r.Context())
	if !ok || h.deps.Database == nil {
		writeJSON(w, 503, map[string]any{"code": "SETTINGS_UNAVAILABLE", "message": "设置服务暂不可用"})
		return
	}
	if r.Method == http.MethodPut {
		var in struct {
			Theme                string `json:"theme"`
			NotificationsEnabled *bool  `json:"notificationsEnabled"`
			StoragePreference    string `json:"storagePreference"`
		}
		if json.NewDecoder(r.Body).Decode(&in) != nil || (in.Theme != "dark" && in.Theme != "light") || (in.StoragePreference != "tos" && in.StoragePreference != "local_executor") {
			writeJSON(w, 400, map[string]any{"code": "SETTINGS_INVALID", "message": "设置内容无效"})
			return
		}
		enabled := true
		if in.NotificationsEnabled != nil {
			enabled = *in.NotificationsEnabled
		}
		_, err := h.deps.Database.ExecContext(r.Context(), `INSERT INTO user_workspace_preferences(user_id,theme,notifications_enabled,storage_preference) VALUES(?,?,?,?) ON DUPLICATE KEY UPDATE theme=VALUES(theme),notifications_enabled=VALUES(notifications_enabled),storage_preference=VALUES(storage_preference)`, u.ID, in.Theme, enabled, in.StoragePreference)
		if err != nil {
			h.writeServiceError(w, r, 500, "SETTINGS_SAVE_FAILED", "保存设置失败", "workspace_settings", "save", err)
			return
		}
	}
	var theme, storage string
	var notify bool
	err := h.deps.Database.QueryRowContext(r.Context(), `SELECT theme,notifications_enabled,storage_preference FROM user_workspace_preferences WHERE user_id=?`, u.ID).Scan(&theme, &notify, &storage)
	if errors.Is(err, sql.ErrNoRows) {
		theme = "dark"
		storage = "tos"
		notify = true
	} else if err != nil {
		h.writeServiceError(w, r, http.StatusInternalServerError, "SETTINGS_LOAD_FAILED", "读取设置失败", "workspace_settings", "load", err)
		return
	}
	executors := []any{}
	executorStatus := map[string]any{"available": true}
	if h.deps.VideoLocalExecutor != nil {
		if rows, e := h.deps.VideoLocalExecutor.List(r.Context()); e == nil {
			for _, x := range rows {
				executors = append(executors, x)
			}
		} else {
			executorStatus = map[string]any{"available": false, "reason": "执行器状态暂不可用，请稍后刷新。"}
		}
	} else {
		executorStatus = map[string]any{"available": false, "reason": "当前环境未配置本地执行器服务。"}
	}
	writeJSON(w, 200, map[string]any{"settings": map[string]any{"theme": theme, "notificationsEnabled": notify, "storagePreference": storage}, "executors": executors, "executorStatus": executorStatus})
}
