package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"time"

	"qiantie/backend/internal/auth"
)

type accountLister interface {
	ListAccounts(context.Context) ([]auth.Account, error)
}

var managedUsernamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{1,31}$`)

func managedAccountError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func managedProfile(profiles *MemberProfileStore, username string) map[string]any {
	for _, row := range profiles.readArray("member-profiles.json") {
		if row["username"] == username {
			return row
		}
	}
	return nil
}

func saveManagedProfile(profiles *MemberProfileStore, username, displayName, role string, monthly any) error {
	profiles.mu.Lock()
	defer profiles.mu.Unlock()
	rows := profiles.readArray("member-profiles.json")
	profile := managedProfileUnlocked(rows, username)
	if profile == nil {
		profile = map[string]any{"username": username}
		rows = append(rows, profile)
	}
	profile["displayName"], profile["role"] = displayName, role
	if _, exists := profile["boundTo"]; !exists {
		profile["boundTo"] = nil
	}
	profile["monthlyTokenLimit"] = monthly
	profile["updatedAt"] = time.Now().UTC().Format(time.RFC3339)
	return profiles.writeArray("member-profiles.json", rows)
}

func managedProfileUnlocked(rows []map[string]any, username string) map[string]any {
	for _, row := range rows {
		if row["username"] == username {
			return row
		}
	}
	return nil
}

func managedAccountRow(account auth.Account, profile map[string]any) map[string]any {
	displayName, role := account.Username, "member"
	var boundTo, monthly any
	if profile != nil {
		if value, ok := profile["displayName"].(string); ok && strings.TrimSpace(value) != "" {
			displayName = value
		}
		if value, ok := profile["role"].(string); ok && value != "" {
			role = value
		}
		boundTo, monthly = profile["boundTo"], profile["monthlyTokenLimit"]
	}
	if account.IsOwner {
		role = "dev"
	}
	return map[string]any{"username": account.Username, "displayName": displayName, "role": role, "active": account.Active, "boundTo": boundTo, "teamOwner": nil, "apiScopes": []any{}, "monthlyTokenLimit": monthly, "presence": map[string]any{"online": false}, "usage": map[string]any{"month": map[string]any{}}}
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
			row := managedAccountRow(account, managedProfile(profiles, account.Username))
			memberRole := strings.ToLower(strings.TrimSpace(toString(row["role"])))
			if role != "" && memberRole != role {
				continue
			}
			displayName := toString(row["displayName"])
			if query != "" && !strings.Contains(strings.ToLower(account.Username+" "+displayName), query) {
				continue
			}
			row["role"] = memberRole
			result = append(result, row)
		}
		writeJSON(w, http.StatusOK, map[string]any{"accounts": result})
	})))
	root.Handle("POST /api/account-admin/accounts", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			managedAccountError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		self := profiles.member(identity.Username, accounts)
		if self == nil || self["role"] != "dev" {
			managedAccountError(w, http.StatusForbidden, "仅 DEV 可以管理全局账号")
			return
		}
		var input struct {
			Username    string `json:"username"`
			DisplayName string `json:"displayName"`
			Password    string `json:"password"`
			Role        string `json:"role"`
			Monthly     any    `json:"monthlyTokenLimit"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 32<<10)).Decode(&input) != nil {
			managedAccountError(w, http.StatusBadRequest, "请求内容无效")
			return
		}
		input.Username, input.DisplayName, input.Role = strings.TrimSpace(input.Username), strings.TrimSpace(input.DisplayName), strings.ToLower(strings.TrimSpace(input.Role))
		if !managedUsernamePattern.MatchString(input.Username) || input.DisplayName == "" || len(input.Password) < 8 || (input.Role != "member" && input.Role != "manager") {
			managedAccountError(w, http.StatusBadRequest, "账号、显示名称、角色或密码格式不正确")
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
		if err := saveManagedProfile(profiles, input.Username, input.DisplayName, input.Role, input.Monthly); err != nil {
			managedAccountError(w, http.StatusInternalServerError, "账号资料保存失败")
			return
		}
		account, _ := accounts.FindAccount(req.Context(), input.Username)
		writeJSON(w, http.StatusCreated, map[string]any{"member": managedAccountRow(account, managedProfile(profiles, input.Username))})
	})))
	root.Handle("PATCH /api/account-admin/accounts/{username}", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			managedAccountError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		self := profiles.member(identity.Username, accounts)
		if self == nil || self["role"] != "dev" {
			managedAccountError(w, http.StatusForbidden, "仅 DEV 可以管理全局账号")
			return
		}
		username := req.PathValue("username")
		account, err := accounts.FindAccount(req.Context(), username)
		if err != nil {
			managedAccountError(w, http.StatusNotFound, "目标账号不存在")
			return
		}
		if account.IsOwner {
			managedAccountError(w, http.StatusBadRequest, "当前开发者账号不可编辑")
			return
		}
		var input struct {
			DisplayName string `json:"displayName"`
			Role        string `json:"role"`
			Active      *bool  `json:"active"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 16<<10)).Decode(&input) != nil {
			managedAccountError(w, http.StatusBadRequest, "请求内容无效")
			return
		}
		input.DisplayName, input.Role = strings.TrimSpace(input.DisplayName), strings.ToLower(strings.TrimSpace(input.Role))
		if input.DisplayName == "" || (input.Role != "member" && input.Role != "manager") {
			managedAccountError(w, http.StatusBadRequest, "显示名称或角色格式不正确")
			return
		}
		old := managedProfile(profiles, username)
		var monthly any
		if old != nil {
			monthly = old["monthlyTokenLimit"]
		}
		if err := saveManagedProfile(profiles, username, input.DisplayName, input.Role, monthly); err != nil {
			managedAccountError(w, http.StatusInternalServerError, "账号资料保存失败")
			return
		}
		if input.Active != nil {
			if err := accounts.SetActive(req.Context(), username, *input.Active); err != nil {
				managedAccountError(w, http.StatusInternalServerError, "账号状态保存失败")
				return
			}
		}
		account, _ = accounts.FindAccount(req.Context(), username)
		writeJSON(w, http.StatusOK, map[string]any{"member": managedAccountRow(account, managedProfile(profiles, username))})
	})))
}
