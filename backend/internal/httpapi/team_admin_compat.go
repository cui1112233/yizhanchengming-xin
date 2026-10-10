package httpapi

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"qiantie/backend/internal/auth"
)

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
			if toString(actor["role"]) != "dev" && owner != manager {
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
