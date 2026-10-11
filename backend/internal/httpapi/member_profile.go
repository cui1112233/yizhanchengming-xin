package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"qiantie/backend/internal/auth"
)

type MemberProfileStore struct {
	SystemDir string
	mu        sync.Mutex
}

func (s *MemberProfileStore) mfaStatus(username string) map[string]any {
	status := map[string]any{"enabled": false, "enabledAt": nil, "recoveryCodesRemaining": 0, "setupPending": false}
	if s == nil || strings.TrimSpace(s.SystemDir) == "" {
		return status
	}
	raw, err := os.ReadFile(filepath.Join(s.SystemDir, "mfa.json"))
	if err != nil {
		return status
	}
	var rows []struct {
		Username       string   `json:"username"`
		Enabled        bool     `json:"enabled"`
		Secret         string   `json:"secret"`
		EnabledAt      string   `json:"enabledAt"`
		RecoveryHashes []string `json:"recoveryHashes"`
		PendingSecret  string   `json:"pendingSecret"`
	}
	if json.Unmarshal(raw, &rows) != nil {
		return status
	}
	for _, row := range rows {
		if row.Username != username {
			continue
		}
		status["enabled"] = row.Enabled && strings.TrimSpace(row.Secret) != ""
		if row.EnabledAt != "" {
			status["enabledAt"] = row.EnabledAt
		}
		status["recoveryCodesRemaining"] = len(row.RecoveryHashes)
		status["setupPending"] = strings.TrimSpace(row.PendingSecret) != ""
		break
	}
	return status
}

func NewMemberProfileStore(systemDir string) *MemberProfileStore {
	return &MemberProfileStore{SystemDir: strings.TrimSpace(systemDir)}
}

func (s *MemberProfileStore) readArray(name string) []map[string]any {
	if s == nil || s.SystemDir == "" {
		return nil
	}
	raw, err := os.ReadFile(filepath.Join(s.SystemDir, name))
	if err != nil {
		return nil
	}
	var rows []map[string]any
	if json.Unmarshal(raw, &rows) != nil {
		return nil
	}
	return rows
}

func (s *MemberProfileStore) writeArray(name string, rows []map[string]any) error {
	if err := os.MkdirAll(s.SystemDir, 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		return err
	}
	tmp := filepath.Join(s.SystemDir, name+".tmp")
	if err := os.WriteFile(tmp, raw, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, filepath.Join(s.SystemDir, name))
}

func textPatch(value any, fallback string, max int) (string, error) {
	if value == nil {
		return fallback, nil
	}
	text, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("字段格式不正确")
	}
	text = strings.TrimSpace(text)
	if len(text) > max {
		return "", fmt.Errorf("内容不能超过 %d 个字符", max)
	}
	return text, nil
}

func (s *MemberProfileStore) update(username string, input map[string]any, accounts auth.Store) (map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	member := s.member(username, accounts)
	if member == nil {
		return nil, auth.ErrNotFound
	}
	profiles := s.readArray("member-profiles.json")
	var profile map[string]any
	for _, row := range profiles {
		if row["username"] == username {
			profile = row
			break
		}
	}
	if profile == nil {
		profile = map[string]any{"username": username, "displayName": username, "avatarUrl": nil, "role": "member", "boundTo": nil, "monthlyTokenLimit": nil}
		profiles = append(profiles, profile)
	}
	if value, exists := input["displayName"]; exists {
		next, err := textPatch(value, fmt.Sprint(profile["displayName"]), 40)
		if err != nil || next == "" {
			if err == nil {
				err = fmt.Errorf("显示名称不能为空")
			}
			return nil, err
		}
		profile["displayName"] = next
	}
	if _, exists := input["bio"]; exists || input["phone"] != nil || input["email"] != nil || input["teamTitle"] != nil {
		details := s.readArray("member-profile-details.json")
		var row map[string]any
		for _, item := range details {
			if item["username"] == username {
				row = item
				break
			}
		}
		if row == nil {
			row = map[string]any{"username": username, "bio": nil, "phone": nil, "email": nil, "teamTitle": nil}
			details = append(details, row)
		}
		for _, field := range []struct {
			name string
			max  int
		}{{"bio", 240}, {"phone", 32}, {"email", 120}, {"teamTitle", 60}} {
			if raw, ok := input[field.name]; ok {
				next, err := textPatch(raw, fmt.Sprint(row[field.name]), field.max)
				if err != nil {
					return nil, err
				}
				if next == "" {
					row[field.name] = nil
				} else {
					row[field.name] = next
				}
			}
		}
		row["updatedAt"] = time.Now().UTC().Format(time.RFC3339)
		if err := s.writeArray("member-profile-details.json", details); err != nil {
			return nil, err
		}
	}
	profile["updatedAt"] = time.Now().UTC().Format(time.RFC3339)
	if err := s.writeArray("member-profiles.json", profiles); err != nil {
		return nil, err
	}
	return s.member(username, accounts), nil
}

func (s *MemberProfileStore) member(ctxUsername string, accounts auth.Store) map[string]any {
	account, err := accounts.FindAccount(context.Background(), ctxUsername)
	if err != nil {
		return nil
	}
	profile := map[string]any{}
	for _, row := range s.readArray("member-profiles.json") {
		if row["username"] == ctxUsername {
			profile = row
			break
		}
	}
	role := "member"
	if account.IsOwner {
		role = "dev"
	} else if value, ok := profile["role"].(string); ok && (value == "dev" || value == "manager") {
		role = value
	}
	member := map[string]any{"username": account.Username, "active": account.Active, "isOwner": account.IsOwner, "createdAt": account.CreatedAt, "updatedAt": account.UpdatedAt, "displayName": account.Username, "avatarUrl": nil, "role": role, "boundTo": nil, "monthlyTokenLimit": nil, "apiScopes": []any{}, "apiEnabled": false, "bio": nil, "phone": nil, "email": nil, "teamTitle": nil}
	for _, key := range []string{"displayName", "avatarUrl", "boundTo", "monthlyTokenLimit", "apiEnabled", "apiScopes"} {
		if value, ok := profile[key]; ok {
			member[key] = value
		}
	}
	for _, row := range s.readArray("member-profile-details.json") {
		if row["username"] != ctxUsername {
			continue
		}
		for _, key := range []string{"bio", "phone", "email", "teamTitle"} {
			if value, ok := row[key]; ok {
				member[key] = value
			}
		}
		break
	}
	if role != "member" {
		member["boundTo"] = nil
	}
	return member
}

func emptyMemberUsage() map[string]any {
	empty := func() map[string]any {
		return map[string]any{"calls": 0, "successfulCalls": 0, "billableCalls": 0, "knownUsageCalls": 0, "pricedCalls": 0, "inputTokens": 0, "outputTokens": 0, "totalTokens": 0, "byFeature": map[string]any{}, "callsByFeature": map[string]any{}, "byModel": map[string]any{}, "costByCurrency": map[string]any{}, "estimatedCost": nil, "costCurrency": nil}
	}
	return map[string]any{"day": empty(), "month": empty(), "recent": []any{}}
}

func memberQuota(limit any, used int) map[string]any {
	quota := map[string]any{"used": used, "limit": limit, "remaining": nil, "percent": nil, "level": "unlimited", "threshold": nil}
	if limit == nil {
		return quota
	}
	value, ok := limit.(float64)
	if !ok {
		value, _ = strconv.ParseFloat(strings.TrimSpace(fmt.Sprint(limit)), 64)
	}
	if value <= 0 {
		return quota
	}
	remaining := value - float64(used)
	if remaining < 0 {
		remaining = 0
	}
	percent := float64(used) / value * 100
	level := "normal"
	if percent >= 100 {
		level = "exhausted"
	} else if percent >= 90 {
		level = "critical"
	} else if percent >= 70 {
		level = "warning"
	}
	quota["remaining"] = remaining
	quota["percent"] = percent
	quota["level"] = level
	return quota
}

func memberTeamGovernance(manager map[string]any) map[string]any {
	usage := emptyMemberUsage()
	monthlyLimit := manager["monthlyTokenLimit"]
	return map[string]any{"managerUsername": manager["username"], "monthlyTokenLimit": monthlyLimit, "thresholds": []any{70, 90, 100}, "updatedAt": nil, "updatedBy": nil, "usage": usage["month"], "quota": memberQuota(monthlyLimit, 0)}
}

func (s *MemberProfileStore) updateAvatar(username, dataURL, avatarDir string, accounts auth.Store) (map[string]any, error) {
	const maxBytes = 2 * 1024 * 1024
	value := strings.TrimSpace(dataURL)
	prefix, encoded, ok := strings.Cut(value, ",")
	if !ok || !strings.HasPrefix(prefix, "data:image/") || !strings.HasSuffix(prefix, ";base64") {
		return nil, fmt.Errorf("头像仅支持 PNG、JPG 或 WebP")
	}
	ext := "png"
	if strings.Contains(prefix, "image/jpeg") {
		ext = "jpg"
	} else if strings.Contains(prefix, "image/webp") {
		ext = "webp"
	} else if !strings.Contains(prefix, "image/png") {
		return nil, fmt.Errorf("头像仅支持 PNG、JPG 或 WebP")
	}
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(decoded) == 0 || len(decoded) > maxBytes {
		return nil, fmt.Errorf("头像大小不能超过 2MB")
	}
	if strings.TrimSpace(avatarDir) == "" {
		return nil, fmt.Errorf("头像存储未配置")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.member(username, accounts) == nil {
		return nil, auth.ErrNotFound
	}
	if err := os.MkdirAll(avatarDir, 0o700); err != nil {
		return nil, err
	}
	for _, other := range []string{"png", "jpg", "webp"} {
		if other != ext {
			_ = os.Remove(filepath.Join(avatarDir, username+"."+other))
		}
	}
	if err := os.WriteFile(filepath.Join(avatarDir, username+"."+ext), decoded, 0o600); err != nil {
		return nil, err
	}
	profiles := s.readArray("member-profiles.json")
	var profile map[string]any
	for _, row := range profiles {
		if row["username"] == username {
			profile = row
			break
		}
	}
	if profile == nil {
		profile = map[string]any{"username": username}
		profiles = append(profiles, profile)
	}
	profile["avatarUrl"] = "/user-content/avatars/" + username + "." + ext
	profile["updatedAt"] = time.Now().UTC().Format(time.RFC3339)
	if err := s.writeArray("member-profiles.json", profiles); err != nil {
		return nil, err
	}
	return s.member(username, accounts), nil
}

func (s *MemberProfileStore) readCollaborationState() map[string]any {
	state := map[string]any{"notifications": []any{}}
	if s == nil || s.SystemDir == "" {
		return state
	}
	raw, err := os.ReadFile(filepath.Join(s.SystemDir, "team-collaboration.json"))
	if err != nil || json.Unmarshal(raw, &state) != nil {
		return map[string]any{"notifications": []any{}}
	}
	if _, ok := state["notifications"].([]any); !ok {
		state["notifications"] = []any{}
	}
	return state
}

func (s *MemberProfileStore) writeCollaborationState(state map[string]any) error {
	if err := os.MkdirAll(s.SystemDir, 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	tmp := filepath.Join(s.SystemDir, "team-collaboration.json.tmp")
	if err := os.WriteFile(tmp, raw, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, filepath.Join(s.SystemDir, "team-collaboration.json"))
}

func (s *MemberProfileStore) listNotifications(username string, limit int) []map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	state := s.readCollaborationState()
	rows, _ := state["notifications"].([]any)
	result := make([]map[string]any, 0)
	for index := len(rows) - 1; index >= 0; index-- {
		row, ok := rows[index].(map[string]any)
		if !ok || row["username"] != username {
			continue
		}
		result = append(result, row)
		if limit > 0 && len(result) >= limit {
			break
		}
	}
	return result
}

func (s *MemberProfileStore) markAllNotificationsRead(username string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	state := s.readCollaborationState()
	rows, _ := state["notifications"].([]any)
	changed := 0
	now := time.Now().UTC().Format(time.RFC3339)
	for _, raw := range rows {
		row, ok := raw.(map[string]any)
		if !ok || row["username"] != username || !notificationUnread(row["readAt"]) {
			continue
		}
		row["readAt"] = now
		changed++
	}
	if changed == 0 {
		return 0, nil
	}
	return changed, s.writeCollaborationState(state)
}

func (s *MemberProfileStore) markNotificationRead(username, id string) (map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	state := s.readCollaborationState()
	rows, _ := state["notifications"].([]any)
	for _, raw := range rows {
		row, ok := raw.(map[string]any)
		if !ok || row["username"] != username || fmt.Sprint(row["id"]) != id {
			continue
		}
		if notificationUnread(row["readAt"]) {
			row["readAt"] = time.Now().UTC().Format(time.RFC3339)
			if err := s.writeCollaborationState(state); err != nil {
				return nil, err
			}
		}
		return row, nil
	}
	return nil, auth.ErrNotFound
}

func notificationLimit(raw string) int {
	limit, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || limit <= 0 {
		return 50
	}
	if limit > 5000 {
		return 5000
	}
	return limit
}

func notificationUnread(value any) bool {
	if value == nil {
		return true
	}
	return strings.TrimSpace(fmt.Sprint(value)) == ""
}

func persistedSessionHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func changeMemberPassword(ctx context.Context, username, currentPassword, newPassword, currentToken string, accounts auth.Store) (int64, error) {
	account, err := accounts.FindAccount(ctx, username)
	if err != nil {
		return 0, err
	}
	if !auth.VerifyPassword(currentPassword, account.PasswordHash, account.PasswordAlgorithm) {
		return 0, fmt.Errorf("当前密码不正确")
	}
	if len(newPassword) < 8 {
		return 0, fmt.Errorf("新密码至少需要 8 位")
	}
	if currentPassword == newPassword {
		return 0, fmt.Errorf("新密码不能与当前密码相同")
	}
	hash, algorithm, err := auth.HashPassword(newPassword)
	if err != nil {
		return 0, err
	}
	if err := accounts.UpdatePassword(ctx, username, hash, algorithm); err != nil {
		return 0, err
	}
	if revoker, ok := accounts.(interface {
		RevokeOtherSessions(context.Context, string, string) (int64, error)
	}); ok {
		return revoker.RevokeOtherSessions(ctx, username, currentToken)
	}
	return 0, nil
}

func RegisterMemberProfileRoutes(root *http.ServeMux, authMiddleware BridgeAuth, profiles *MemberProfileStore, accounts auth.Store, avatarDir string) {
	if root == nil || profiles == nil || accounts == nil {
		return
	}
	root.Handle("GET /api/member/me", authMiddleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		member := profiles.member(identity.Username, accounts)
		if member == nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "账号不存在"})
			return
		}
		manager := map[string]any(nil)
		team := map[string]any(nil)
		if toString(member["role"]) == "dev" || toString(member["role"]) == "manager" {
			team = goTeam(member)
		} else if bound := strings.TrimSpace(toString(member["boundTo"])); bound != "" {
			if found, err := accounts.FindAccount(req.Context(), bound); err == nil {
				manager = profiles.member(found.Username, accounts)
				if manager != nil {
					team = goTeam(manager)
				}
			}
		}
		notifications := profiles.listNotifications(identity.Username, 8)
		unread := 0
		for _, notification := range notifications {
			if notificationUnread(notification["readAt"]) {
				unread++
			}
		}
		usage := emptyMemberUsage()
		governanceManager := member
		if manager != nil {
			governanceManager = manager
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"member":              member,
			"presence":            map[string]any{"online": true},
			"manager":             manager,
			"team":                team,
			"notificationSummary": map[string]any{"unread": unread, "recent": notifications},
			"memberQuota":         memberQuota(member["monthlyTokenLimit"], 0),
			"teamGovernance":      memberTeamGovernance(governanceManager),
			"usage":               usage,
		})
	})))
	root.Handle("PATCH /api/member/profile", authMiddleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input map[string]any
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 64<<10)).Decode(&input) != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求内容无效"})
			return
		}
		member, err := profiles.update(identity.Username, input, accounts)
		if err != nil {
			status := http.StatusBadRequest
			if err == auth.ErrNotFound {
				status = http.StatusNotFound
			}
			writeJSON(w, status, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"member": member})
	})))
	root.Handle("POST /api/member/avatar", authMiddleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input struct {
			AvatarDataURL string `json:"avatarDataUrl"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 3<<20)).Decode(&input) != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求内容无效"})
			return
		}
		member, err := profiles.updateAvatar(identity.Username, input.AvatarDataURL, avatarDir, accounts)
		if err != nil {
			status := http.StatusBadRequest
			if err == auth.ErrNotFound {
				status = http.StatusNotFound
			}
			writeJSON(w, status, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"member": member})
	})))
	root.Handle("POST /api/member/security/password", authMiddleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input struct {
			CurrentPassword string `json:"currentPassword"`
			NewPassword     string `json:"newPassword"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 16<<10)).Decode(&input) != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求内容无效"})
			return
		}
		revoked, err := changeMemberPassword(req.Context(), identity.Username, input.CurrentPassword, input.NewPassword, bearerToken(req), accounts)
		if err != nil {
			status := http.StatusBadRequest
			if err == auth.ErrNotFound {
				status = http.StatusNotFound
			}
			writeJSON(w, status, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"changed": true, "revoked": revoked})
	})))
	root.Handle("GET /api/member/notifications", authMiddleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		entries := profiles.listNotifications(identity.Username, notificationLimit(req.URL.Query().Get("limit")))
		unread := 0
		for _, entry := range entries {
			if notificationUnread(entry["readAt"]) {
				unread++
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{"entries": entries, "unread": unread})
	})))
	root.Handle("GET /api/member/security", authMiddleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		member := profiles.member(identity.Username, accounts)
		if member == nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "账号不存在"})
			return
		}
		sessions := []map[string]any{}
		if listing, ok := accounts.(interface {
			ListSessionsForUser(context.Context, string) ([]auth.Session, error)
		}); ok {
			if persisted, err := listing.ListSessionsForUser(req.Context(), identity.Username); err == nil {
				currentHash := persistedSessionHash(bearerToken(req))
				for _, session := range persisted {
					sessions = append(sessions, map[string]any{
						"id":       session.Token[:minInt(10, len(session.Token))],
						"current":  bearerToken(req) != "" && session.Token == currentHash,
						"issuedAt": session.IssuedAt,
						"browser":  session.Browser,
						"os":       session.OS,
						"ipHint":   session.IPHint,
					})
				}
			}
		}
		other := 0
		for _, session := range sessions {
			if current, _ := session["current"].(bool); !current {
				other++
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"account":            member,
			"mfa":                profiles.mfaStatus(identity.Username),
			"sessions":           sessions,
			"persistentSessions": []any{},
			"otherSessionCount":  other,
		})
	})))
	root.Handle("POST /api/member/security/sessions/revoke-others", authMiddleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		revoked := int64(0)
		if revoker, ok := accounts.(interface {
			RevokeOtherSessions(context.Context, string, string) (int64, error)
		}); ok {
			count, err := revoker.RevokeOtherSessions(req.Context(), identity.Username, bearerToken(req))
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "退出其他设备失败"})
				return
			}
			revoked = count
		}
		writeJSON(w, http.StatusOK, map[string]any{"revoked": revoked})
	})))
	root.Handle("POST /api/member/notifications/read-all", authMiddleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		changed, err := profiles.markAllNotificationsRead(identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "通知状态保存失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"changed": changed})
	})))
	root.Handle("POST /api/member/notifications/{id}/read", authMiddleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		entry, err := profiles.markNotificationRead(identity.Username, req.PathValue("id"))
		if err != nil {
			status := http.StatusBadRequest
			if err == auth.ErrNotFound {
				status = http.StatusNotFound
			}
			writeJSON(w, status, map[string]string{"error": "通知不存在"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"notification": entry})
	})))
}
