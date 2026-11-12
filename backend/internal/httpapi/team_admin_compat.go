package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"qiantie/backend/internal/auth"
)

func managedTeamState(profiles *MemberProfileStore) (string, map[string]any, []any) {
	path := adminSystemFile(profiles.SystemDir, "team-collaboration.json")
	state := map[string]any{"teams": []any{}, "invites": []any{}, "notifications": []any{}, "archives": []any{}}
	if raw, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(raw, &state)
	}
	teams, _ := state["teams"].([]any)
	return path, state, teams
}

func saveManagedTeamState(path string, state map[string]any) error {
	raw, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o600)
}

func managedTeamByID(profiles *MemberProfileStore, id string) (string, map[string]any, map[string]any) {
	path, state, teams := managedTeamState(profiles)
	for _, raw := range teams {
		if team, ok := raw.(map[string]any); ok && toString(team["id"]) == id {
			return path, state, team
		}
	}
	if strings.HasPrefix(id, "go-team-") {
		manager := strings.TrimPrefix(id, "go-team-")
		team := map[string]any{"id": id, "managerUsername": manager, "name": manager + "的团队", "coManagers": []any{}}
		teams = append(teams, team)
		state["teams"] = teams
		return path, state, team
	}
	return path, state, nil
}

func teamAdminIdentity(req *http.Request, profiles *MemberProfileStore, accounts auth.Store) (map[string]any, bool) {
	return teamActor(profiles, accounts, req)
}

func teamAdminCanAccess(actor, team map[string]any, ownerOnly bool) bool {
	if actor == nil || team == nil || !teamCanManage(actor) {
		return false
	}
	if toString(actor["role"]) == "dev" {
		return true
	}
	if toString(team["managerUsername"]) == toString(actor["username"]) {
		return true
	}
	return !ownerOnly && containsString(team["coManagers"], toString(actor["username"]))
}

func findAuthAccount(ctx context.Context, accounts auth.Store, username string) (auth.Account, error) {
	return accounts.FindAccount(ctx, username)
}

func managedBool(value any) bool {
	if parsed, ok := value.(bool); ok {
		return parsed
	}
	return strings.EqualFold(toString(value), "true")
}

func readManagedTeams(profiles *MemberProfileStore) []map[string]any {
	path := adminSystemFile(profiles.SystemDir, "team-collaboration.json")
	state := map[string]any{}
	if raw, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(raw, &state)
	}
	rows, _ := state["teams"].([]any)
	teams := make([]map[string]any, 0, len(rows))
	for _, raw := range rows {
		if row, ok := raw.(map[string]any); ok && strings.TrimSpace(toString(row["managerUsername"])) != "" {
			teams = append(teams, row)
		}
	}
	return teams
}

func teamAdminMembers(accounts auth.Store, profiles *MemberProfileStore, manager string, req *http.Request) []map[string]any {
	rows := []map[string]any{}
	lister, ok := accounts.(accountLister)
	if !ok {
		return rows
	}
	accountsList, err := lister.ListAccounts(req.Context())
	if err != nil {
		return rows
	}
	for _, account := range accountsList {
		member := profiles.member(account.Username, accounts)
		if member == nil {
			continue
		}
		if account.Username == manager || (toString(member["role"]) == "member" && toString(member["boundTo"]) == manager) {
			member["usage"] = map[string]any{"month": map[string]any{"totalTokens": 0}}
			rows = append(rows, member)
		}
	}
	return rows
}

func RegisterTeamAdminCompat(root *http.ServeMux, middleware BridgeAuth, accounts auth.Store, profiles *MemberProfileStore) {
	if root == nil || accounts == nil || profiles == nil {
		return
	}
	root.Handle("GET /api/team-admin/teams", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		actor, ok := teamActor(profiles, accounts, req)
		if !ok || actor == nil {
			managedAccountError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		if !teamCanManage(actor) {
			managedAccountError(w, http.StatusForbidden, "当前身份没有团队管理权限")
			return
		}
		manager := toString(actor["username"])
		teams := make([]map[string]any, 0)
		for _, team := range readManagedTeams(profiles) {
			owner := toString(team["managerUsername"])
			// 联合管理员也必须能看到被委派的团队，否则界面无法进入该团队，
			// “跨团队协作”就只剩后端鉴权、没有任何入口。
			if toString(actor["role"]) != "dev" && owner != manager && !containsString(team["coManagers"], manager) {
				continue
			}
			ownerProfile := profiles.member(owner, accounts)
			if ownerProfile == nil {
				continue
			}
			members := teamAdminMembers(accounts, profiles, owner, req)
			teams = append(teams, map[string]any{"team": team, "manager": ownerProfile, "coManagers": team["coManagers"], "memberCount": len(members) - 1, "usage": map[string]any{"totalTokens": 0}, "governance": map[string]any{"monthlyTokenLimit": ownerProfile["monthlyTokenLimit"]}, "quota": map[string]any{"used": 0, "limit": ownerProfile["monthlyTokenLimit"]}})
		}
		if len(teams) == 0 {
			teams = append(teams, map[string]any{"team": goTeam(actor), "manager": actor, "coManagers": []any{}, "memberCount": 0, "usage": map[string]any{"totalTokens": 0}, "governance": map[string]any{"monthlyTokenLimit": actor["monthlyTokenLimit"]}, "quota": map[string]any{"used": 0, "limit": actor["monthlyTokenLimit"]}})
		}
		writeJSON(w, http.StatusOK, map[string]any{"teams": teams})
	})))
	root.Handle("GET /api/team-admin/teams/{teamId}/members", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		actor, ok := teamActor(profiles, accounts, req)
		if !ok || actor == nil {
			managedAccountError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		var team map[string]any
		for _, candidate := range readManagedTeams(profiles) {
			if toString(candidate["id"]) == req.PathValue("teamId") {
				team = candidate
				break
			}
		}
		manager := toString(actor["username"])
		if team != nil {
			manager = toString(team["managerUsername"])
		}
		if team == nil && req.PathValue("teamId") == "go-team-"+toString(actor["username"]) {
			team = goTeam(actor)
			manager = toString(actor["username"])
		}
		if team == nil || (toString(actor["role"]) != "dev" && manager != toString(actor["username"]) && !containsString(team["coManagers"], toString(actor["username"]))) {
			managedAccountError(w, http.StatusNotFound, "团队不存在")
			return
		}
		if team == nil {
			team = goTeam(actor)
		}
		writeJSON(w, http.StatusOK, map[string]any{"team": team, "members": teamAdminMembers(accounts, profiles, manager, req)})
	})))
	root.Handle("PATCH /api/team-admin/teams/{teamId}/governance", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		actor, ok := teamAdminIdentity(req, profiles, accounts)
		if !ok || actor == nil {
			managedAccountError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		var input struct {
			MonthlyTokenLimit *int64 `json:"monthlyTokenLimit"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 8<<10)).Decode(&input) != nil || input.MonthlyTokenLimit != nil && (*input.MonthlyTokenLimit < 0 || *input.MonthlyTokenLimit > 10000000000) {
			managedAccountError(w, http.StatusBadRequest, "月度额度不正确")
			return
		}
		path, state, team := managedTeamByID(profiles, req.PathValue("teamId"))
		if !teamAdminCanAccess(actor, team, false) {
			managedAccountError(w, http.StatusForbidden, "无权修改该团队")
			return
		}
		manager := toString(team["managerUsername"])
		if err := updateManagedProfileFields(profiles, manager, func(row map[string]any) { row["monthlyTokenLimit"] = input.MonthlyTokenLimit }); err != nil || saveManagedTeamState(path, state) != nil {
			managedAccountError(w, http.StatusInternalServerError, "团队额度保存失败")
			return
		}
		managerProfile := profiles.member(manager, accounts)
		writeJSON(w, http.StatusOK, map[string]any{"team": team, "manager": managerProfile, "governance": map[string]any{"monthlyTokenLimit": input.MonthlyTokenLimit}, "quota": map[string]any{"used": 0, "limit": input.MonthlyTokenLimit}})
	})))
	root.Handle("PUT /api/team-admin/teams/{teamId}/co-managers", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		actor, ok := teamAdminIdentity(req, profiles, accounts)
		if !ok || actor == nil {
			managedAccountError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		var input struct {
			Usernames []string `json:"usernames"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 16<<10)).Decode(&input) != nil {
			managedAccountError(w, http.StatusBadRequest, "请求内容无效")
			return
		}
		path, state, team := managedTeamByID(profiles, req.PathValue("teamId"))
		if !teamAdminCanAccess(actor, team, true) {
			managedAccountError(w, http.StatusForbidden, "仅团队主负责人可设置联合管理员")
			return
		}
		seen := map[string]bool{}
		coManagers := make([]any, 0, len(input.Usernames))
		for _, username := range input.Usernames {
			username = strings.TrimSpace(username)
			if username == "" || seen[username] || username == toString(team["managerUsername"]) {
				continue
			}
			member := profiles.member(username, accounts)
			if member == nil || toString(member["role"]) != "manager" || !managedBool(member["active"]) {
				managedAccountError(w, http.StatusBadRequest, "联合管理员账号不可用")
				return
			}
			seen[username] = true
			coManagers = append(coManagers, username)
		}
		team["coManagers"] = coManagers
		if err := saveManagedTeamState(path, state); err != nil {
			managedAccountError(w, http.StatusInternalServerError, "联合管理员保存失败")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"team": team, "coManagers": coManagers})
	})))
	root.Handle("PUT /api/team-admin/teams/{teamId}/members/{username}/api-scopes", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		actor, ok := teamAdminIdentity(req, profiles, accounts)
		if !ok || actor == nil {
			managedAccountError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		_, _, team := managedTeamByID(profiles, req.PathValue("teamId"))
		if !teamAdminCanAccess(actor, team, false) {
			managedAccountError(w, http.StatusForbidden, "无权管理该团队")
			return
		}
		target := profiles.member(req.PathValue("username"), accounts)
		manager := toString(team["managerUsername"])
		if target == nil || toString(target["role"]) != "member" || toString(target["boundTo"]) != manager {
			managedAccountError(w, http.StatusForbidden, "目标成员不属于该团队")
			return
		}
		var input struct {
			Scopes []string `json:"scopes"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 8<<10)).Decode(&input) != nil {
			managedAccountError(w, http.StatusBadRequest, "请求内容无效")
			return
		}
		seen := map[string]bool{}
		scopes := make([]string, 0, len(input.Scopes))
		for _, scope := range input.Scopes {
			scope = strings.TrimSpace(scope)
			if scope == "" || !managedAPIScopes[scope] || seen[scope] {
				continue
			}
			seen[scope] = true
			scopes = append(scopes, scope)
		}
		if err := updateManagedProfileFields(profiles, toString(target["username"]), func(row map[string]any) { row["apiScopes"], row["apiEnabled"] = scopes, len(scopes) > 0 }); err != nil {
			managedAccountError(w, http.StatusInternalServerError, "权限保存失败")
			return
		}
		updated := profiles.member(req.PathValue("username"), accounts)
		writeJSON(w, http.StatusOK, map[string]any{"member": updated, "changedBy": actor["username"]})
	})))
	root.Handle("POST /api/team-admin/teams/{teamId}/members/{username}/status", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		actor, ok := teamAdminIdentity(req, profiles, accounts)
		if !ok || actor == nil {
			managedAccountError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		_, _, team := managedTeamByID(profiles, req.PathValue("teamId"))
		if !teamAdminCanAccess(actor, team, false) {
			managedAccountError(w, http.StatusForbidden, "无权管理该团队")
			return
		}
		target := profiles.member(req.PathValue("username"), accounts)
		if target == nil || toString(target["role"]) != "member" || toString(target["boundTo"]) != toString(team["managerUsername"]) {
			managedAccountError(w, http.StatusForbidden, "目标成员不属于该团队")
			return
		}
		var input struct {
			Active *bool `json:"active"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 8<<10)).Decode(&input) != nil || input.Active == nil {
			managedAccountError(w, http.StatusBadRequest, "账号状态不正确")
			return
		}
		if err := accounts.SetActive(req.Context(), req.PathValue("username"), *input.Active); err != nil {
			managedAccountError(w, http.StatusBadRequest, "账号状态保存失败")
			return
		}
		updated := profiles.member(req.PathValue("username"), accounts)
		writeJSON(w, http.StatusOK, map[string]any{"member": updated, "changedBy": actor["username"]})
	})))
	root.Handle("POST /api/team-admin/teams/{teamId}/members/{username}/reset-password", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		actor, ok := teamAdminIdentity(req, profiles, accounts)
		if !ok || actor == nil {
			managedAccountError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		_, _, team := managedTeamByID(profiles, req.PathValue("teamId"))
		target := profiles.member(req.PathValue("username"), accounts)
		if !teamAdminCanAccess(actor, team, false) || target == nil || toString(target["role"]) != "member" || toString(target["boundTo"]) != toString(team["managerUsername"]) {
			managedAccountError(w, http.StatusForbidden, "目标成员不属于该团队")
			return
		}
		var input struct {
			Password string `json:"password"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 8<<10)).Decode(&input) != nil || len(input.Password) < 8 {
			managedAccountError(w, http.StatusBadRequest, "新密码至少需要8位")
			return
		}
		hash, algorithm, err := auth.HashPassword(input.Password)
		if err != nil || accounts.UpdatePassword(req.Context(), req.PathValue("username"), hash, algorithm) != nil {
			managedAccountError(w, http.StatusBadRequest, "密码保存失败")
			return
		}
		// 重置密码必须同时作废该成员的全部登录会话，否则旧 token 在过期前
		// 仍然可用，界面提示的“旧会话已撤销”就是假的。
		revoked := int64(0)
		if revoker, ok := accounts.(interface {
			RevokeOtherSessions(context.Context, string, string) (int64, error)
		}); ok {
			revoked, _ = revoker.RevokeOtherSessions(req.Context(), req.PathValue("username"), "")
		}
		writeJSON(w, http.StatusOK, map[string]any{"changed": true, "changedBy": actor["username"], "revokedSessions": revoked})
	})))
}

func containsString(value any, wanted string) bool {
	items, ok := value.([]any)
	if !ok {
		return false
	}
	for _, item := range items {
		if toString(item) == wanted {
			return true
		}
	}
	return false
}
