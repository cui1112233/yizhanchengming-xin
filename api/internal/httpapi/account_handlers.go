package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
)

// accountProfile is deliberately a safe session projection. Sensitive auth
// material is never queried from auth_sessions or credential tables.
func (h handler) accountProfile(w http.ResponseWriter, r *http.Request) {
	u, ok := authn.CurrentUser(r.Context())
	if !ok || u.ID <= 0 || h.deps.Database == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"code": "AUTH_UNAUTHENTICATED", "message": "登录状态无效或已过期"})
		return
	}
	if r.Method == http.MethodPut {
		var in struct {
			DisplayName string `json:"displayName"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxRequestBody)).Decode(&in); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"code": "PROFILE_INVALID", "message": "资料内容无效"})
			return
		}
		in.DisplayName = strings.TrimSpace(in.DisplayName)
		if in.DisplayName == "" || len([]rune(in.DisplayName)) > 64 {
			writeJSON(w, http.StatusBadRequest, map[string]any{"code": "PROFILE_INVALID", "message": "显示名称长度无效"})
			return
		}
		if _, err := h.deps.Database.ExecContext(r.Context(), `UPDATE auth_users SET display_name=? WHERE id=? AND active=TRUE`, in.DisplayName, u.ID); err != nil {
			h.writeServiceError(w, r, http.StatusInternalServerError, "PROFILE_SAVE_FAILED", "保存资料失败", "account_profile", "save", err)
			return
		}
		u.DisplayName = in.DisplayName
	}
	writeJSON(w, http.StatusOK, safeAccountProjection(u))
}

func (h handler) memberCenter(w http.ResponseWriter, r *http.Request) {
	u, ok := authn.CurrentUser(r.Context())
	if !ok || u.ID <= 0 {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"code": "AUTH_UNAUTHENTICATED", "message": "登录状态无效或已过期"})
		return
	}
	// Usage, activities and notifications are intentionally unavailable until a
	// durable per-user projection exists; no browser-generated statistics.
	p := safeAccountProjection(u)
	p["usage"] = map[string]any{"availability": "unavailable"}
	p["activity"] = []any{}
	p["notifications"] = []any{}
	writeJSON(w, http.StatusOK, p)
}

func safeAccountProjection(u authn.User) map[string]any {
	return map[string]any{"profile": map[string]any{"id": u.ID, "username": u.Username, "displayName": u.DisplayName, "avatar": map[string]any{"availability": "unavailable"}}, "membership": map[string]any{"role": u.Role, "teamId": u.TeamID, "capabilities": u.Capabilities}, "security": map[string]any{"emailVerification": "unavailable", "mfa": "unavailable"}}
}
