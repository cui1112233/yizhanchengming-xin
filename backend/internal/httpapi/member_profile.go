package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"qiantie/backend/internal/auth"
)

type MemberProfileStore struct{ SystemDir string }

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
}
