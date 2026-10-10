package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"qiantie/backend/internal/auth"
)

type MemberProfileStore struct {
	SystemDir string
	mu        sync.Mutex
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
	member := map[string]any{"username": account.Username, "active": account.Active, "isOwner": account.IsOwner, "createdAt": account.CreatedAt, "updatedAt": account.UpdatedAt, "displayName": account.Username, "avatarUrl": nil, "role": role, "boundTo": nil, "monthlyTokenLimit": nil, "apiScopes": []any{}, "apiEnabled": account.IsOwner}
	for _, key := range []string{"displayName", "avatarUrl", "boundTo", "monthlyTokenLimit"} {
		if value, ok := profile[key]; ok {
			member[key] = value
		}
	}
	if role != "member" {
		member["boundTo"] = nil
	}
	return member
}

func RegisterMemberProfileRoutes(root *http.ServeMux, authMiddleware BridgeAuth, profiles *MemberProfileStore, accounts auth.Store) {
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
		writeJSON(w, http.StatusOK, map[string]any{"member": member, "presence": map[string]any{"online": true}, "manager": nil, "team": nil, "notificationSummary": map[string]any{"unread": 0, "recent": []any{}}, "memberQuota": map[string]any{"used": 0, "limit": member["monthlyTokenLimit"]}, "teamGovernance": nil, "usage": map[string]any{"day": map[string]any{}, "month": map[string]any{}, "recent": []any{}}})
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
}
