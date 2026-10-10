package httpapi

import (
	"net/http"
	"strings"

	"qiantie/backend/internal/auth"
)

func goTeam(manager map[string]any) map[string]any {
	username := toString(manager["username"])
	name := toString(manager["displayName"])
	if name == "" {
		name = username
	}
	return map[string]any{"id": "go-team-" + username, "name": name + "的团队", "managerUsername": username, "coManagers": []any{}}
}

func RegisterMemberTeamReadCompat(root *http.ServeMux, middleware BridgeAuth, accounts auth.Store, profiles *MemberProfileStore) {
	if root == nil || accounts == nil || profiles == nil {
		return
	}
	root.Handle("GET /api/member/team/meta", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		self := profiles.member(identity.Username, accounts)
		if self == nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "账号不存在"})
			return
		}
		if toString(self["role"]) == "dev" {
			owners := []map[string]any{self}
			if lister, ok := accounts.(accountLister); ok {
				if rows, err := lister.ListAccounts(req.Context()); err == nil {
					for _, account := range rows {
						member := profiles.member(account.Username, accounts)
						if member != nil && toString(member["role"]) == "manager" {
							owners = append(owners, member)
						}
					}
				}
			}
			teams := make([]map[string]any, 0, len(owners))
			for _, owner := range owners {
				teams = append(teams, map[string]any{"team": goTeam(owner), "manager": owner})
			}
			writeJSON(w, http.StatusOK, map[string]any{"team": goTeam(self), "teams": teams})
			return
		}
		manager := self
		if bound := strings.TrimSpace(toString(self["boundTo"])); bound != "" {
			if found, err := accounts.FindAccount(req.Context(), bound); err == nil {
				manager = profiles.member(found.Username, accounts)
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{"team": goTeam(manager), "manager": manager})
	})))
	root.Handle("GET /api/member/team", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		self := profiles.member(identity.Username, accounts)
		if self == nil || (toString(self["role"]) != "dev" && toString(self["role"]) != "manager") {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "当前身份没有团队管理权限"})
			return
		}
		owner := self
		lister, ok := accounts.(accountLister)
		if !ok {
			writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "团队数据尚未接入 Go 存储"})
			return
		}
		rows, err := lister.ListAccounts(req.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "团队读取失败"})
			return
		}
		members := make([]map[string]any, 0)
		for _, account := range rows {
			member := profiles.member(account.Username, accounts)
			if member == nil {
				continue
			}
			role := toString(member["role"])
			bound := strings.TrimSpace(toString(member["boundTo"]))
			if account.Username == toString(owner["username"]) || (role == "member" && bound == toString(owner["username"])) {
				member["presence"] = map[string]any{"online": false}
				member["archive"] = nil
				member["quota"] = map[string]any{"used": 0, "limit": member["monthlyTokenLimit"]}
				member["usage"] = map[string]any{"day": map[string]any{}, "month": map[string]any{}}
				members = append(members, member)
			}
		}
		response := map[string]any{"team": goTeam(owner), "teamGovernance": map[string]any{"monthlyTokenLimit": owner["monthlyTokenLimit"]}, "members": members}
		if toString(self["role"]) == "dev" {
			response["teams"] = []any{goTeam(owner)}
		}
		writeJSON(w, http.StatusOK, response)
	})))
	root.Handle("GET /api/member/team/governance", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		self := profiles.member(identity.Username, accounts)
		if self == nil || (toString(self["role"]) != "dev" && toString(self["role"]) != "manager") {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "当前身份没有团队额度管理权限"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"teams": []any{map[string]any{"manager": self, "team": goTeam(self), "monthlyTokenLimit": self["monthlyTokenLimit"], "used": 0}}})
	})))
	root.Handle("GET /api/member/team/members/{username}/usage", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		self := profiles.member(identity.Username, accounts)
		target := profiles.member(req.PathValue("username"), accounts)
		if self == nil || target == nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "成员不存在"})
			return
		}
		if toString(self["role"]) == "manager" && toString(target["boundTo"]) != toString(self["username"]) && toString(target["username"]) != toString(self["username"]) {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "无权查看该成员用量"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"day": map[string]any{}, "month": map[string]any{"totalTokens": 0}, "quota": map[string]any{"used": 0, "limit": target["monthlyTokenLimit"]}, "recent": []any{}})
	})))
}
