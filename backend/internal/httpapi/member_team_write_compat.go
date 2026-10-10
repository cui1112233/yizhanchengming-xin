package httpapi

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"qiantie/backend/internal/auth"
)

func collaborationToken() (string, string, error) {
	raw := make([]byte, 24)
	if _, err := rand.Read(raw); err != nil {
		return "", "", err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	digest := sha256.Sum256([]byte(token))
	return token, fmt.Sprintf("%x", digest[:]), nil
}

var managedAPIScopes = map[string]bool{"text": true, "image": true, "video": true, "tts": true}

func updateManagedProfileFields(profiles *MemberProfileStore, username string, update func(map[string]any)) error {
	profiles.mu.Lock()
	defer profiles.mu.Unlock()
	rows := profiles.readArray("member-profiles.json")
	profile := managedProfileUnlocked(rows, username)
	if profile == nil {
		profile = map[string]any{"username": username}
		rows = append(rows, profile)
	}
	update(profile)
	profile["updatedAt"] = time.Now().UTC().Format(time.RFC3339)
	return profiles.writeArray("member-profiles.json", rows)
}

func updateMemberArchive(profiles *MemberProfileStore, username, actor, reason string, restored bool) (map[string]any, error) {
	profiles.mu.Lock()
	defer profiles.mu.Unlock()
	path := filepath.Join(profiles.SystemDir, "member-archives.json")
	rows := []map[string]any{}
	if raw, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(raw, &rows)
	}
	var archive map[string]any
	for _, row := range rows {
		if row["username"] == username {
			archive = row
			break
		}
	}
	if archive == nil {
		archive = map[string]any{"username": username}
		rows = append(rows, archive)
	}
	if restored {
		archive["restoredAt"], archive["restoredBy"] = time.Now().UTC().Format(time.RFC3339), actor
	} else {
		archive["archivedAt"], archive["archivedBy"], archive["reason"] = time.Now().UTC().Format(time.RFC3339), actor, strings.TrimSpace(reason)
	}
	if err := os.MkdirAll(profiles.SystemDir, 0o755); err != nil {
		return nil, err
	}
	raw, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		return nil, err
	}
	return archive, nil
}

func teamActor(profiles *MemberProfileStore, accounts auth.Store, req *http.Request) (map[string]any, bool) {
	identity, ok := BridgeIdentityFromContext(req.Context())
	if !ok {
		return nil, false
	}
	return profiles.member(identity.Username, accounts), true
}

func teamCanManage(actor map[string]any) bool {
	role := toString(actor["role"])
	return role == "dev" || role == "manager"
}

func RegisterMemberTeamWriteCompat(root *http.ServeMux, middleware BridgeAuth, accounts auth.Store, profiles *MemberProfileStore) {
	if root == nil || accounts == nil || profiles == nil {
		return
	}
	root.Handle("POST /api/member/team/invites", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		actor, ok := teamActor(profiles, accounts, req)
		if !ok {
			managedAccountError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		if actor == nil || !teamCanManage(actor) {
			managedAccountError(w, http.StatusForbidden, "无权创建团队邀请")
			return
		}
		var input struct {
			ManagerUsername   string   `json:"managerUsername"`
			ExpiresInHours    float64  `json:"expiresInHours"`
			MonthlyTokenLimit *int64   `json:"monthlyTokenLimit"`
			APIScopes         []string `json:"apiScopes"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, req.Body, 16<<10)).Decode(&input); err != nil {
			managedAccountError(w, http.StatusBadRequest, "请求内容无效")
			return
		}
		manager := toString(actor["username"])
		if toString(actor["role"]) == "dev" && strings.TrimSpace(input.ManagerUsername) != "" {
			manager = strings.TrimSpace(input.ManagerUsername)
		}
		expiresInHours := input.ExpiresInHours
		if expiresInHours == 0 {
			expiresInHours = 72
		}
		if expiresInHours < 1 || expiresInHours > 720 {
			managedAccountError(w, http.StatusBadRequest, "邀请有效期需在1到720小时之间")
			return
		}
		if input.MonthlyTokenLimit != nil && (*input.MonthlyTokenLimit < 0 || *input.MonthlyTokenLimit > 10000000000) {
			managedAccountError(w, http.StatusBadRequest, "月度额度不正确")
			return
		}
		scopes := make([]string, 0, len(input.APIScopes))
		seenScopes := map[string]bool{}
		for _, scope := range input.APIScopes {
			scope = strings.TrimSpace(scope)
			if scope != "" && managedAPIScopes[scope] && !seenScopes[scope] {
				scopes = append(scopes, scope)
				seenScopes[scope] = true
			}
		}
		managerProfile := profiles.member(manager, accounts)
		if managerProfile == nil || (toString(managerProfile["role"]) != "dev" && toString(managerProfile["role"]) != "manager") {
			managedAccountError(w, http.StatusBadRequest, "目标账号不是团队负责人")
			return
		}
		path := adminSystemFile(profiles.SystemDir, "team-collaboration.json")
		state := map[string]any{"teams": []any{}, "invites": []any{}, "notifications": []any{}, "archives": []any{}}
		if raw, err := os.ReadFile(path); err == nil {
			_ = json.Unmarshal(raw, &state)
		}
		teams, _ := state["teams"].([]any)
		var team map[string]any
		for _, raw := range teams {
			if row, ok := raw.(map[string]any); ok && toString(row["managerUsername"]) == manager {
				team = row
				break
			}
		}
		if team == nil {
			team = map[string]any{"id": "go-team-" + manager, "name": toString(managerProfile["displayName"]) + "的团队", "managerUsername": manager, "coManagers": []any{}, "createdAt": time.Now().UTC().Format(time.RFC3339)}
			teams = append(teams, team)
			state["teams"] = teams
		}
		token, digest, err := collaborationToken()
		if err != nil {
			managedAccountError(w, http.StatusInternalServerError, "邀请生成失败")
			return
		}
		expires := time.Now().UTC().Add(time.Duration(expiresInHours * float64(time.Hour)))
		invite := map[string]any{"id": token[:12], "tokenDigest": digest, "teamId": team["id"], "managerUsername": manager, "createdBy": toString(actor["username"]), "createdAt": time.Now().UTC().Format(time.RFC3339), "expiresAt": expires.Format(time.RFC3339), "monthlyTokenLimit": input.MonthlyTokenLimit, "apiScopes": scopes, "usedAt": nil, "usedBy": nil}
		invites, _ := state["invites"].([]any)
		invites = append(invites, invite)
		state["invites"] = invites
		raw, _ := json.MarshalIndent(state, "", "  ")
		if err := os.WriteFile(path, raw, 0o600); err != nil {
			managedAccountError(w, http.StatusInternalServerError, "邀请保存失败")
			return
		}
		public := map[string]any{"id": invite["id"], "teamId": team["id"], "teamName": team["name"], "managerUsername": manager, "createdAt": invite["createdAt"], "expiresAt": invite["expiresAt"], "monthlyTokenLimit": input.MonthlyTokenLimit, "apiScopes": scopes, "usedAt": nil, "usedBy": nil}
		writeJSON(w, http.StatusCreated, map[string]any{"token": token, "invite": public, "inviteUrl": "/invite/" + token})
	})))
	root.Handle("POST /api/member/team/members", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		actor, ok := teamActor(profiles, accounts, req)
		if !ok {
			managedAccountError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		if actor == nil || !teamCanManage(actor) {
			managedAccountError(w, http.StatusForbidden, "当前身份没有成员创建权限")
			return
		}
		var input struct {
			Username    string `json:"username"`
			DisplayName string `json:"displayName"`
			Password    string `json:"password"`
			BoundTo     string `json:"boundTo"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 32<<10)).Decode(&input) != nil {
			managedAccountError(w, http.StatusBadRequest, "请求内容无效")
			return
		}
		input.Username, input.DisplayName = strings.TrimSpace(input.Username), strings.TrimSpace(input.DisplayName)
		if !managedUsernamePattern.MatchString(input.Username) || input.DisplayName == "" || len(input.Password) < 8 {
			managedAccountError(w, http.StatusBadRequest, "账号、显示名称或密码格式不正确")
			return
		}
		boundTo := strings.TrimSpace(input.BoundTo)
		if toString(actor["role"]) == "manager" {
			boundTo = toString(actor["username"])
		}
		if boundTo == "" {
			managedAccountError(w, http.StatusBadRequest, "请先绑定团队负责人")
			return
		}
		owner, err := accounts.FindAccount(req.Context(), boundTo)
		if err != nil {
			managedAccountError(w, http.StatusBadRequest, "团队负责人不存在")
			return
		}
		ownerProfile := profiles.member(owner.Username, accounts)
		if ownerProfile == nil || (toString(ownerProfile["role"]) != "dev" && toString(ownerProfile["role"]) != "manager") {
			managedAccountError(w, http.StatusBadRequest, "目标账号不是团队负责人")
			return
		}
		if _, err := accounts.FindAccount(req.Context(), input.Username); err == nil {
			managedAccountError(w, http.StatusConflict, "账号已存在")
			return
		} else if err != auth.ErrNotFound {
			managedAccountError(w, http.StatusInternalServerError, "账号检查失败")
			return
		}
		hash, algorithm, err := auth.HashPassword(input.Password)
		if err != nil {
			managedAccountError(w, http.StatusInternalServerError, "密码处理失败")
			return
		}
		now := time.Now().UTC()
		if err := accounts.CreateAccount(req.Context(), auth.Account{Username: input.Username, PasswordHash: hash, PasswordAlgorithm: algorithm, Active: true, CreatedAt: now, UpdatedAt: now}); err != nil {
			managedAccountError(w, http.StatusConflict, "账号创建失败")
			return
		}
		if err := saveManagedProfile(profiles, input.Username, input.DisplayName, "member", nil); err != nil {
			managedAccountError(w, http.StatusInternalServerError, "成员资料保存失败")
			return
		}
		profiles.mu.Lock()
		rows := profiles.readArray("member-profiles.json")
		profile := managedProfileUnlocked(rows, input.Username)
		profile["boundTo"] = boundTo
		_ = profiles.writeArray("member-profiles.json", rows)
		profiles.mu.Unlock()
		account, _ := accounts.FindAccount(req.Context(), input.Username)
		writeJSON(w, http.StatusCreated, map[string]any{"member": managedAccountRow(account, profile)})
	})))
	root.Handle("PATCH /api/member/team/governance/{username}", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		actor, ok := teamActor(profiles, accounts, req)
		if !ok {
			managedAccountError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		if actor == nil || !teamCanManage(actor) {
			managedAccountError(w, http.StatusForbidden, "当前身份没有团队额度管理权限")
			return
		}
		username := req.PathValue("username")
		manager := profiles.member(username, accounts)
		if manager == nil || (toString(manager["role"]) != "dev" && toString(manager["role"]) != "manager") {
			managedAccountError(w, http.StatusBadRequest, "目标账号不是团队负责人")
			return
		}
		if toString(actor["role"]) == "manager" && toString(actor["username"]) != username {
			managedAccountError(w, http.StatusForbidden, "MANAGER 只能设置自己团队的总额度")
			return
		}
		var input struct {
			MonthlyTokenLimit any `json:"monthlyTokenLimit"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 8<<10)).Decode(&input) != nil {
			managedAccountError(w, http.StatusBadRequest, "请求内容无效")
			return
		}
		if input.MonthlyTokenLimit != nil {
			if value, ok := input.MonthlyTokenLimit.(float64); !ok || value < 0 || value != float64(int64(value)) {
				managedAccountError(w, http.StatusBadRequest, "额度必须是非负整数")
				return
			}
		}
		if err := updateManagedProfileFields(profiles, username, func(row map[string]any) { row["monthlyTokenLimit"] = input.MonthlyTokenLimit }); err != nil {
			managedAccountError(w, http.StatusInternalServerError, "团队额度保存失败")
			return
		}
		updated := profiles.member(username, accounts)
		writeJSON(w, http.StatusOK, map[string]any{"manager": updated, "team": goTeam(updated), "monthlyTokenLimit": updated["monthlyTokenLimit"], "used": 0})
	})))
	root.Handle("PATCH /api/member/team/members/{username}", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		actor, ok := teamActor(profiles, accounts, req)
		if !ok {
			managedAccountError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		if actor == nil || !teamCanManage(actor) {
			managedAccountError(w, http.StatusForbidden, "无权修改团队成员")
			return
		}
		username := req.PathValue("username")
		account, err := accounts.FindAccount(req.Context(), username)
		if err != nil {
			managedAccountError(w, http.StatusNotFound, "成员不存在")
			return
		}
		if account.IsOwner {
			managedAccountError(w, http.StatusBadRequest, "开发者账号不可作为团队成员编辑")
			return
		}
		current := profiles.member(username, accounts)
		if current == nil || toString(current["role"]) != "member" {
			managedAccountError(w, http.StatusBadRequest, "目标账号不是团队成员")
			return
		}
		if toString(actor["role"]) == "manager" && toString(current["boundTo"]) != toString(actor["username"]) {
			managedAccountError(w, http.StatusForbidden, "不能修改其他团队成员")
			return
		}
		var input struct {
			DisplayName string `json:"displayName"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 16<<10)).Decode(&input) != nil || strings.TrimSpace(input.DisplayName) == "" {
			managedAccountError(w, http.StatusBadRequest, "显示名称不能为空")
			return
		}
		if err := saveManagedProfile(profiles, username, strings.TrimSpace(input.DisplayName), "member", current["monthlyTokenLimit"]); err != nil {
			managedAccountError(w, http.StatusInternalServerError, "成员资料保存失败")
			return
		}
		account, _ = accounts.FindAccount(req.Context(), username)
		writeJSON(w, http.StatusOK, map[string]any{"member": managedAccountRow(account, managedProfile(profiles, username))})
	})))
	root.Handle("POST /api/member/team/members/{username}/status", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		actor, ok := teamActor(profiles, accounts, req)
		if !ok {
			managedAccountError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		if actor == nil || !teamCanManage(actor) {
			managedAccountError(w, http.StatusForbidden, "无权修改成员状态")
			return
		}
		username := req.PathValue("username")
		member := profiles.member(username, accounts)
		if member == nil || toString(member["role"]) != "member" {
			managedAccountError(w, http.StatusNotFound, "成员不存在")
			return
		}
		if toString(actor["role"]) == "manager" && toString(member["boundTo"]) != toString(actor["username"]) {
			managedAccountError(w, http.StatusForbidden, "不能修改其他团队成员")
			return
		}
		var input struct {
			Active bool `json:"active"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 8<<10)).Decode(&input) != nil {
			managedAccountError(w, http.StatusBadRequest, "请求内容无效")
			return
		}
		if err := accounts.SetActive(req.Context(), username, input.Active); err != nil {
			managedAccountError(w, http.StatusInternalServerError, "成员状态保存失败")
			return
		}
		account, _ := accounts.FindAccount(req.Context(), username)
		writeJSON(w, http.StatusOK, map[string]any{"member": managedAccountRow(account, managedProfile(profiles, username))})
	})))
	root.Handle("POST /api/member/team/members/{username}/reset-password", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		actor, ok := teamActor(profiles, accounts, req)
		if !ok {
			managedAccountError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		if actor == nil || !teamCanManage(actor) {
			managedAccountError(w, http.StatusForbidden, "无权重置成员密码")
			return
		}
		username := req.PathValue("username")
		member := profiles.member(username, accounts)
		if member == nil || toString(member["role"]) != "member" {
			managedAccountError(w, http.StatusNotFound, "成员不存在")
			return
		}
		if toString(actor["role"]) == "manager" && toString(member["boundTo"]) != toString(actor["username"]) {
			managedAccountError(w, http.StatusForbidden, "不能修改其他团队成员")
			return
		}
		var input struct {
			Password string `json:"password"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 8<<10)).Decode(&input) != nil || len(input.Password) < 8 {
			managedAccountError(w, http.StatusBadRequest, "密码至少需要 8 位")
			return
		}
		hash, algorithm, err := auth.HashPassword(input.Password)
		if err != nil {
			managedAccountError(w, http.StatusInternalServerError, "密码处理失败")
			return
		}
		if err := accounts.UpdatePassword(req.Context(), username, hash, algorithm); err != nil {
			managedAccountError(w, http.StatusInternalServerError, "成员密码保存失败")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"reset": true, "username": username})
	})))
	root.Handle("POST /api/member/team/members/{username}/archive", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		actor, ok := teamActor(profiles, accounts, req)
		if !ok {
			managedAccountError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		username := req.PathValue("username")
		member := profiles.member(username, accounts)
		if actor == nil || !teamCanManage(actor) {
			managedAccountError(w, http.StatusForbidden, "无权归档成员")
			return
		}
		if member == nil || toString(member["role"]) != "member" {
			managedAccountError(w, http.StatusNotFound, "成员不存在")
			return
		}
		if toString(actor["role"]) == "manager" && toString(member["boundTo"]) != toString(actor["username"]) {
			managedAccountError(w, http.StatusForbidden, "不能归档其他团队成员")
			return
		}
		var input struct {
			Reason string `json:"reason"`
		}
		_ = json.NewDecoder(http.MaxBytesReader(w, req.Body, 8<<10)).Decode(&input)
		if err := accounts.SetActive(req.Context(), username, false); err != nil {
			managedAccountError(w, http.StatusInternalServerError, "成员归档失败")
			return
		}
		archive, err := updateMemberArchive(profiles, username, toString(actor["username"]), input.Reason, false)
		if err != nil {
			managedAccountError(w, http.StatusInternalServerError, "归档记录保存失败")
			return
		}
		account, _ := accounts.FindAccount(req.Context(), username)
		writeJSON(w, http.StatusOK, map[string]any{"member": managedAccountRow(account, managedProfile(profiles, username)), "archive": archive, "revoked": 0})
	})))
	root.Handle("POST /api/member/team/members/{username}/api", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		actor, ok := teamActor(profiles, accounts, req)
		if !ok {
			managedAccountError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		username := req.PathValue("username")
		member := profiles.member(username, accounts)
		if actor == nil || !teamCanManage(actor) {
			managedAccountError(w, http.StatusForbidden, "无权修改成员 API 权限")
			return
		}
		if member == nil || toString(member["role"]) != "member" {
			managedAccountError(w, http.StatusNotFound, "成员不存在")
			return
		}
		if toString(actor["role"]) == "manager" && toString(member["boundTo"]) != toString(actor["username"]) {
			managedAccountError(w, http.StatusForbidden, "不能修改其他团队成员")
			return
		}
		var input struct {
			Enabled bool   `json:"enabled"`
			Scope   string `json:"scope"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 8<<10)).Decode(&input) != nil {
			managedAccountError(w, http.StatusBadRequest, "请求内容无效")
			return
		}
		scope := strings.TrimSpace(input.Scope)
		if scope == "" {
			scope = "*"
		}
		if scope != "*" && !managedAPIScopes[scope] {
			managedAccountError(w, http.StatusBadRequest, "API scope 不合法")
			return
		}
		if err := updateManagedProfileFields(profiles, username, func(row map[string]any) {
			row["apiEnabled"] = input.Enabled
			if input.Enabled && scope != "*" {
				row["apiScopes"] = []string{scope}
			} else if !input.Enabled {
				row["apiScopes"] = []string{}
			}
		}); err != nil {
			managedAccountError(w, http.StatusInternalServerError, "成员权限保存失败")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"member": profiles.member(username, accounts)})
	})))
	root.Handle("PUT /api/member/team/members/{username}/api-scopes", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		actor, ok := teamActor(profiles, accounts, req)
		if !ok {
			managedAccountError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		username := req.PathValue("username")
		member := profiles.member(username, accounts)
		if actor == nil || !teamCanManage(actor) {
			managedAccountError(w, http.StatusForbidden, "无权修改成员 API 权限")
			return
		}
		if member == nil || toString(member["role"]) != "member" {
			managedAccountError(w, http.StatusNotFound, "成员不存在")
			return
		}
		if toString(actor["role"]) == "manager" && toString(member["boundTo"]) != toString(actor["username"]) {
			managedAccountError(w, http.StatusForbidden, "不能修改其他团队成员")
			return
		}
		var input struct {
			Scopes []string `json:"scopes"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 16<<10)).Decode(&input) != nil {
			managedAccountError(w, http.StatusBadRequest, "请求内容无效")
			return
		}
		seen := map[string]bool{}
		scopes := make([]string, 0, len(input.Scopes))
		for _, scope := range input.Scopes {
			scope = strings.TrimSpace(scope)
			if scope == "" || seen[scope] {
				continue
			}
			if !managedAPIScopes[scope] {
				managedAccountError(w, http.StatusBadRequest, "API scopes 不合法")
				return
			}
			seen[scope] = true
			scopes = append(scopes, scope)
		}
		if err := updateManagedProfileFields(profiles, username, func(row map[string]any) { row["apiScopes"] = scopes; row["apiEnabled"] = len(scopes) > 0 }); err != nil {
			managedAccountError(w, http.StatusInternalServerError, "成员权限保存失败")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"member": profiles.member(username, accounts)})
	})))
	root.Handle("POST /api/member/team/members/{username}/restore", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		actor, ok := teamActor(profiles, accounts, req)
		if !ok {
			managedAccountError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		username := req.PathValue("username")
		member := profiles.member(username, accounts)
		if actor == nil || !teamCanManage(actor) {
			managedAccountError(w, http.StatusForbidden, "无权恢复成员")
			return
		}
		if member == nil || toString(member["role"]) != "member" {
			managedAccountError(w, http.StatusNotFound, "成员不存在")
			return
		}
		if toString(actor["role"]) == "manager" && toString(member["boundTo"]) != toString(actor["username"]) {
			managedAccountError(w, http.StatusForbidden, "不能恢复其他团队成员")
			return
		}
		if err := accounts.SetActive(req.Context(), username, true); err != nil {
			managedAccountError(w, http.StatusInternalServerError, "成员恢复失败")
			return
		}
		archive, err := updateMemberArchive(profiles, username, toString(actor["username"]), "", true)
		if err != nil {
			managedAccountError(w, http.StatusInternalServerError, "归档记录保存失败")
			return
		}
		account, _ := accounts.FindAccount(req.Context(), username)
		writeJSON(w, http.StatusOK, map[string]any{"member": managedAccountRow(account, managedProfile(profiles, username)), "archive": archive})
	})))
}
