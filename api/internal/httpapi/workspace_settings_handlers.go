package httpapi

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
	"net/http"
	"regexp"
)

var workspacePetID = regexp.MustCompile(`^[a-z0-9_-]{1,64}$`)

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
			PetID                string `json:"petId"`
			SoundVolume          int    `json:"soundVolume"`
			PetVisible           *bool  `json:"petVisible"`
			CompanionActive      *bool  `json:"companionActive"`
		}
		if json.NewDecoder(r.Body).Decode(&in) != nil || (in.Theme != "dark" && in.Theme != "light") || (in.StoragePreference != "tos" && in.StoragePreference != "local_executor") || !workspacePetID.MatchString(in.PetID) || in.SoundVolume < 0 || in.SoundVolume > 100 {
			writeJSON(w, 400, map[string]any{"code": "SETTINGS_INVALID", "message": "设置内容无效"})
			return
		}
		enabled := true
		if in.NotificationsEnabled != nil {
			enabled = *in.NotificationsEnabled
		}
		petVisible, companionActive := true, false
		if in.PetVisible != nil {
			petVisible = *in.PetVisible
		}
		if in.CompanionActive != nil {
			companionActive = *in.CompanionActive
		}
		_, err := h.deps.Database.ExecContext(r.Context(), `INSERT INTO user_workspace_preferences(user_id,theme,notifications_enabled,storage_preference,pet_id,sound_volume,pet_visible,companion_active) VALUES(?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE theme=VALUES(theme),notifications_enabled=VALUES(notifications_enabled),storage_preference=VALUES(storage_preference),pet_id=VALUES(pet_id),sound_volume=VALUES(sound_volume),pet_visible=VALUES(pet_visible),companion_active=VALUES(companion_active)`, u.ID, in.Theme, enabled, in.StoragePreference, in.PetID, in.SoundVolume, petVisible, companionActive)
		if err != nil {
			h.writeServiceError(w, r, 500, "SETTINGS_SAVE_FAILED", "保存设置失败", "workspace_settings", "save", err)
			return
		}
	}
	var theme, storage, petID string
	var notify, petVisible, companionActive bool
	var soundVolume int
	err := h.deps.Database.QueryRowContext(r.Context(), `SELECT theme,notifications_enabled,storage_preference,pet_id,sound_volume,pet_visible,companion_active FROM user_workspace_preferences WHERE user_id=?`, u.ID).Scan(&theme, &notify, &storage, &petID, &soundVolume, &petVisible, &companionActive)
	if errors.Is(err, sql.ErrNoRows) {
		theme = "dark"
		storage = "tos"
		notify = true
		petID, soundVolume, petVisible, companionActive = "default", 60, true, false
	} else if err != nil {
		h.writeServiceError(w, r, http.StatusInternalServerError, "SETTINGS_LOAD_FAILED", "读取设置失败", "workspace_settings", "load", err)
		return
	}
	executors := []any{}
	// This is a user-scoped projection, never a browser inference. A missing
	// service or an empty owner-scoped result is explicitly unavailable.
	executorStatus := map[string]any{"status": "unavailable", "reasonCode": "not_configured", "displayName": "本地执行器", "online": false}
	if h.deps.VideoLocalExecutor != nil {
		if rows, e := h.deps.VideoLocalExecutor.ListForOwner(r.Context(), u.ID); e == nil {
			for _, x := range rows {
				executors = append(executors, x)
			}
			if len(rows) > 0 {
				online := false
				for _, x := range rows {
					if x.Online {
						online = true
						break
					}
				}
				if online {
					executorStatus = map[string]any{"status": "available", "reasonCode": "available", "displayName": "本地执行器", "online": true}
				} else {
					executorStatus = map[string]any{"status": "degraded", "reasonCode": "offline", "displayName": "本地执行器", "online": false}
				}
			}
		} else {
			executorStatus = map[string]any{"status": "degraded", "reasonCode": "status_unavailable", "displayName": "本地执行器", "online": false}
		}
	}
	writeJSON(w, 200, map[string]any{"settings": map[string]any{"theme": theme, "notificationsEnabled": notify, "storagePreference": storage, "petId": petID, "soundVolume": soundVolume, "petVisible": petVisible, "companionActive": companionActive}, "executors": executors, "executorStatus": executorStatus})
}
