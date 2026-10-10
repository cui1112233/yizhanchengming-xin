package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"qiantie/backend/internal/auth"
)

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
}
