package httpapi

import (
	"context"
	"net/http"
	"strings"

	"qiantie/backend/internal/auth"
)

type accountLister interface {
	ListAccounts(context.Context) ([]auth.Account, error)
}

// RegisterAccountAdminReadCompat keeps the account directory contract on Go.
// Mutating account operations remain on the old service until their complete
// audit/session/team behavior is ported.
func RegisterAccountAdminReadCompat(root *http.ServeMux, middleware BridgeAuth, accounts auth.Store, profiles *MemberProfileStore) {
	if root == nil || accounts == nil || profiles == nil {
		return
	}
	root.Handle("GET /api/account-admin/accounts", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		self := profiles.member(identity.Username, accounts)
		if self == nil || self["role"] != "dev" {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "仅 DEV 可以管理全局账号"})
			return
		}
		lister, ok := accounts.(accountLister)
		if !ok {
			writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "账号目录尚未接入 Go 存储"})
			return
		}
		rows, err := lister.ListAccounts(req.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "账号目录读取失败"})
			return
		}
		role := strings.ToLower(strings.TrimSpace(req.URL.Query().Get("role")))
		query := strings.ToLower(strings.TrimSpace(req.URL.Query().Get("query")))
		result := make([]map[string]any, 0, len(rows))
		for _, account := range rows {
			member := profiles.member(account.Username, accounts)
			if member == nil {
				continue
			}
			memberRole := strings.ToLower(strings.TrimSpace(toString(member["role"])))
			if role != "" && memberRole != role {
				continue
			}
			displayName := toString(member["displayName"])
			if query != "" && !strings.Contains(strings.ToLower(account.Username+" "+displayName), query) {
				continue
			}
			result = append(result, map[string]any{
				"username":          account.Username,
				"displayName":       displayName,
				"role":              memberRole,
				"active":            account.Active,
				"boundTo":           member["boundTo"],
				"teamOwner":         nil,
				"apiScopes":         member["apiScopes"],
				"monthlyTokenLimit": member["monthlyTokenLimit"],
				"presence":          map[string]any{"online": false},
				"usage":             map[string]any{"month": map[string]any{}},
			})
		}
		writeJSON(w, http.StatusOK, map[string]any{"accounts": result})
	})))
}
