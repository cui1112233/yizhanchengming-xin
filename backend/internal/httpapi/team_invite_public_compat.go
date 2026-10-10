package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"qiantie/backend/internal/auth"
)

var publicInviteUsernamePattern = regexp.MustCompile(`^[A-Za-z0-9_-]{3,32}$`)

func RegisterPublicTeamInviteCompat(root *http.ServeMux, systemDir string, accounts auth.Store, profiles *MemberProfileStore) {
	if root == nil || strings.TrimSpace(systemDir) == "" || accounts == nil || profiles == nil {
		return
	}
	root.Handle("GET /api/login/invite/{token}", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		token := strings.TrimSpace(req.PathValue("token"))
		digestBytes := sha256.Sum256([]byte(token))
		digest := hex.EncodeToString(digestBytes[:])
		state := map[string]any{"teams": []any{}, "invites": []any{}}
		if raw, err := os.ReadFile(adminSystemFile(systemDir, "team-collaboration.json")); err == nil {
			_ = json.Unmarshal(raw, &state)
		}
		var found map[string]any
		for _, item := range arrayOrEmpty(state["invites"]) {
			row := mapOrEmpty(item)
			if toString(row["tokenDigest"]) == digest {
				found = row
				break
			}
		}
		if found == nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "邀请不存在"})
			return
		}
		expires, _ := time.Parse(time.RFC3339, toString(found["expiresAt"]))
		expired := !expires.IsZero() && time.Now().UTC().After(expires)
		used := strings.TrimSpace(toString(found["usedAt"])) != ""
		managerName := toString(found["managerUsername"])
		manager := profiles.member(managerName, accounts)
		teamName := ""
		for _, item := range arrayOrEmpty(state["teams"]) {
			team := mapOrEmpty(item)
			if toString(team["id"]) == toString(found["teamId"]) {
				teamName = toString(team["name"])
				break
			}
		}
		invite := map[string]any{"id": found["id"], "teamId": found["teamId"], "teamName": teamName, "managerUsername": managerName, "createdAt": found["createdAt"], "expiresAt": found["expiresAt"], "monthlyTokenLimit": found["monthlyTokenLimit"], "apiScopes": found["apiScopes"], "usedAt": found["usedAt"], "usedBy": found["usedBy"], "available": !expired && !used, "expired": expired}
		var managerPublic map[string]any
		if manager != nil {
			managerPublic = map[string]any{"username": managerName, "displayName": manager["displayName"]}
		}
		writeJSON(w, http.StatusOK, map[string]any{"invite": invite, "manager": managerPublic})
	}))
	root.Handle("POST /api/login/invite/{token}", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var input struct {
			Username    string `json:"username"`
			Password    string `json:"password"`
			DisplayName string `json:"displayName"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 16<<10)).Decode(&input) != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求内容无效"})
			return
		}
		input.Username, input.DisplayName = strings.TrimSpace(input.Username), strings.TrimSpace(input.DisplayName)
		if !publicInviteUsernamePattern.MatchString(input.Username) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "登录账号需为 3-32 位字母、数字、下划线或短横线"})
			return
		}
		if len(input.Password) < 8 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "密码至少需要 8 位"})
			return
		}
		invite, team, manager, err := claimPublicInvite(profiles, accounts, tokenFromRequest(req), input.Username)
		if err != nil {
			writeJSON(w, inviteErrorStatus(err), map[string]string{"error": err.Error()})
			return
		}
		claimed := true
		hash, algorithm, err := auth.HashPassword(input.Password)
		if err == nil {
			now := time.Now().UTC()
			err = accounts.CreateAccount(req.Context(), auth.Account{Username: input.Username, PasswordHash: hash, PasswordAlgorithm: algorithm, Active: true, CreatedAt: now, UpdatedAt: now})
		}
		if err == nil {
			err = saveManagedProfile(profiles, input.Username, input.DisplayName, "member", invite["monthlyTokenLimit"])
		}
		if err != nil {
			if claimed {
				releasePublicInvite(profiles, tokenFromRequest(req), input.Username)
			}
			writeJSON(w, http.StatusConflict, map[string]string{"error": "接受邀请失败"})
			return
		}
		_ = updateManagedProfileFields(profiles, input.Username, func(row map[string]any) {
			row["boundTo"] = manager["username"]
			row["apiEnabled"] = false
			row["apiScopes"] = []string{}
		})
		if scopes, ok := invite["apiScopes"].([]any); ok {
			parsed := make([]string, 0, len(scopes))
			for _, scope := range scopes {
				if managedAPIScopes[toString(scope)] {
					parsed = append(parsed, toString(scope))
				}
			}
			_ = updateManagedProfileFields(profiles, input.Username, func(row map[string]any) { row["apiScopes"], row["apiEnabled"] = parsed, len(parsed) > 0 })
		}
		account, _ := accounts.FindAccount(req.Context(), input.Username)
		member := managedAccountRow(account, managedProfile(profiles, input.Username))
		writeJSON(w, http.StatusCreated, map[string]any{"member": member, "team": team})
	}))
}

func tokenFromRequest(req *http.Request) string { return strings.TrimSpace(req.PathValue("token")) }

func claimPublicInvite(profiles *MemberProfileStore, accounts auth.Store, token, username string) (map[string]any, map[string]any, map[string]any, error) {
	profiles.mu.Lock()
	defer profiles.mu.Unlock()
	state := profiles.readCollaborationState()
	digestBytes := sha256.Sum256([]byte(token))
	digest := hex.EncodeToString(digestBytes[:])
	var invite map[string]any
	for _, item := range arrayOrEmpty(state["invites"]) {
		row := mapOrEmpty(item)
		if toString(row["tokenDigest"]) == digest {
			invite = row
			break
		}
	}
	if invite == nil {
		return nil, nil, nil, fmt.Errorf("邀请不存在")
	}
	expires, _ := time.Parse(time.RFC3339, toString(invite["expiresAt"]))
	if !expires.IsZero() && time.Now().UTC().After(expires) {
		return nil, nil, nil, fmt.Errorf("邀请已过期")
	}
	if strings.TrimSpace(toString(invite["usedAt"])) != "" {
		return nil, nil, nil, fmt.Errorf("邀请已被使用")
	}
	if _, err := accounts.FindAccount(context.Background(), username); err == nil {
		return nil, nil, nil, fmt.Errorf("账号已存在")
	}
	manager := profiles.member(toString(invite["managerUsername"]), accounts)
	active, _ := manager["active"].(bool)
	if manager == nil || !active || (toString(manager["role"]) != "dev" && toString(manager["role"]) != "manager") {
		return nil, nil, nil, fmt.Errorf("邀请所属团队当前不可用")
	}
	invite["usedAt"], invite["usedBy"] = time.Now().UTC().Format(time.RFC3339), username
	if err := profiles.writeCollaborationState(state); err != nil {
		return nil, nil, nil, fmt.Errorf("接受邀请失败")
	}
	team := goTeam(manager)
	for _, item := range arrayOrEmpty(state["teams"]) {
		row := mapOrEmpty(item)
		if toString(row["id"]) == toString(invite["teamId"]) {
			team = row
			break
		}
	}
	return invite, team, manager, nil
}

func releasePublicInvite(profiles *MemberProfileStore, token, username string) {
	profiles.mu.Lock()
	defer profiles.mu.Unlock()
	state := profiles.readCollaborationState()
	digestBytes := sha256.Sum256([]byte(token))
	digest := hex.EncodeToString(digestBytes[:])
	for _, item := range arrayOrEmpty(state["invites"]) {
		row := mapOrEmpty(item)
		if toString(row["tokenDigest"]) == digest && toString(row["usedBy"]) == username {
			row["usedAt"], row["usedBy"] = nil, nil
			_ = profiles.writeCollaborationState(state)
			return
		}
	}
}

func inviteErrorStatus(err error) int {
	if strings.Contains(err.Error(), "不存在") {
		return http.StatusNotFound
	}
	if strings.Contains(err.Error(), "已") || strings.Contains(err.Error(), "账号") || strings.Contains(err.Error(), "团队") {
		return http.StatusConflict
	}
	return http.StatusBadRequest
}
