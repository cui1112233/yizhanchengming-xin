package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"qiantie/backend/internal/auth"
)

func RegisterAccountTransferCompat(root *http.ServeMux, middleware BridgeAuth, accounts auth.Store, profiles *MemberProfileStore) {
	if root == nil || accounts == nil || profiles == nil {
		return
	}
	root.Handle("GET /api/account-admin/accounts/{username}/transfer-preview", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if !accountTransferDev(w, req, accounts, profiles) {
			return
		}
		preview, err := accountTransferPreview(req.PathValue("username"), req.URL.Query().Get("boundTo"), req.URL.Query().Get("resetMonthlyTokenLimit") == "true", accounts, profiles)
		if err != nil {
			managedAccountError(w, transferStatus(err), err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"preview": preview})
	})))
	root.Handle("POST /api/account-admin/accounts/{username}/transfer", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			managedAccountError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		self := profiles.member(identity.Username, accounts)
		if self == nil || toString(self["role"]) != "dev" || self["active"] != true {
			managedAccountError(w, http.StatusForbidden, "仅 DEV 可以管理全局账号")
			return
		}
		var input struct {
			BoundTo                string `json:"boundTo"`
			ResetMonthlyTokenLimit bool   `json:"resetMonthlyTokenLimit"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 16<<10)).Decode(&input) != nil {
			managedAccountError(w, http.StatusBadRequest, "请求内容无效")
			return
		}
		preview, err := accountTransferPreview(req.PathValue("username"), input.BoundTo, input.ResetMonthlyTokenLimit, accounts, profiles)
		if err != nil {
			managedAccountError(w, transferStatus(err), err.Error())
			return
		}
		if err := updateManagedProfileFields(profiles, req.PathValue("username"), func(row map[string]any) {
			row["boundTo"] = strings.TrimSpace(input.BoundTo)
			row["apiScopes"] = []string{}
			row["apiEnabled"] = false
			if input.ResetMonthlyTokenLimit {
				row["monthlyTokenLimit"] = nil
			}
		}); err != nil {
			managedAccountError(w, http.StatusInternalServerError, "成员转移失败")
			return
		}
		member := profiles.member(req.PathValue("username"), accounts)
		recipients := accountTransferNotify(profiles, preview, member)
		writeJSON(w, http.StatusOK, map[string]any{"preview": preview, "member": member, "notifications": map[string]any{"recipients": recipients}})
	})))
}

func accountTransferDev(w http.ResponseWriter, req *http.Request, accounts auth.Store, profiles *MemberProfileStore) bool {
	identity, ok := BridgeIdentityFromContext(req.Context())
	if !ok {
		managedAccountError(w, http.StatusUnauthorized, "unauthorized")
		return false
	}
	self := profiles.member(identity.Username, accounts)
	if self == nil || toString(self["role"]) != "dev" || self["active"] != true {
		managedAccountError(w, http.StatusForbidden, "仅 DEV 可以管理全局账号")
		return false
	}
	return true
}

func accountTransferOwner(username string, accounts auth.Store, profiles *MemberProfileStore) map[string]any {
	if strings.TrimSpace(username) == "" {
		return nil
	}
	owner := profiles.member(username, accounts)
	if owner == nil || (toString(owner["role"]) != "dev" && toString(owner["role"]) != "manager") || owner["active"] != true {
		return nil
	}
	team := goTeam(owner)
	return map[string]any{"username": owner["username"], "displayName": owner["displayName"], "role": owner["role"], "team": map[string]any{"id": team["id"], "name": team["name"]}, "monthlyTokenLimit": owner["monthlyTokenLimit"]}
}

func accountTransferPreview(username, boundTo string, reset bool, accounts auth.Store, profiles *MemberProfileStore) (map[string]any, error) {
	member := profiles.member(strings.TrimSpace(username), accounts)
	if member == nil {
		return nil, fmt.Errorf("目标账号不存在")
	}
	if member["active"] != true {
		return nil, fmt.Errorf("停用账号不能转移团队")
	}
	if toString(member["role"]) != "member" {
		return nil, fmt.Errorf("仅 MEMBER 可以转移团队")
	}
	destination := accountTransferOwner(boundTo, accounts, profiles)
	if destination == nil {
		return nil, fmt.Errorf("目标团队负责人不存在或不可用")
	}
	if toString(member["boundTo"]) == toString(destination["username"]) {
		return nil, fmt.Errorf("成员已属于该团队")
	}
	from := accountTransferOwner(toString(member["boundTo"]), accounts, profiles)
	fromWrap := any(nil)
	if from != nil {
		fromWrap = map[string]any{"owner": from, "team": from["team"], "teamMonthlyTokenLimit": from["monthlyTokenLimit"]}
	}
	return map[string]any{"member": map[string]any{"username": member["username"], "displayName": member["displayName"]}, "from": fromWrap, "to": map[string]any{"owner": destination, "team": destination["team"], "teamMonthlyTokenLimit": destination["monthlyTokenLimit"]}, "oldScopes": member["apiScopes"], "personalMonthlyTokenLimit": map[string]any{"before": member["monthlyTokenLimit"], "after": map[bool]any{true: nil, false: member["monthlyTokenLimit"]}[reset], "reset": reset}, "requiresReauthorization": true}, nil
}

func accountTransferNotify(profiles *MemberProfileStore, preview, member map[string]any) []string {
	recipients := []string{toString(member["username"])}
	if from, ok := preview["from"].(map[string]any); ok {
		if owner, ok := from["owner"].(map[string]any); ok && toString(owner["username"]) != "" {
			recipients = append(recipients, toString(owner["username"]))
		}
	}
	if to, ok := preview["to"].(map[string]any); ok {
		if owner, ok := to["owner"].(map[string]any); ok && toString(owner["username"]) != "" {
			recipients = append(recipients, toString(owner["username"]))
		}
	}
	seen := map[string]bool{}
	unique := make([]string, 0, len(recipients))
	for _, username := range recipients {
		if seen[username] {
			continue
		}
		seen[username] = true
		unique = append(unique, username)
	}
	profiles.mu.Lock()
	defer profiles.mu.Unlock()
	state := profiles.readCollaborationState()
	rows := arrayOrEmpty(state["notifications"])
	now := time.Now().UTC().Format(time.RFC3339)
	for _, username := range unique {
		isMember := username == toString(member["username"])
		title := "团队成员归属已变更"
		message := fmt.Sprintf("成员 @%s 已调整团队，需要重新授予 API 能力。", member["username"])
		kind := "team.member_transferred"
		if isMember {
			title = "所属团队已变更，等待重新授权"
			message = "你所属的团队已变更，原 API 权限已清除，请等待新团队重新授权。"
			kind = "team.transferred"
		}
		rows = append(rows, map[string]any{"id": fmt.Sprintf("transfer-%d-%s", time.Now().UnixNano(), username), "username": username, "type": kind, "title": title, "message": message, "metadata": map[string]any{"member": member["username"], "requiresReauthorization": true}, "createdAt": now, "readAt": nil})
	}
	state["notifications"] = rows
	_ = profiles.writeCollaborationState(state)
	return unique
}

func transferStatus(err error) int {
	if strings.Contains(err.Error(), "不存在") {
		return http.StatusNotFound
	}
	if strings.Contains(err.Error(), "停用") || strings.Contains(err.Error(), "已属于") {
		return http.StatusConflict
	}
	return http.StatusBadRequest
}
