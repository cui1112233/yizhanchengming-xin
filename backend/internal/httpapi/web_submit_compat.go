package httpapi

import (
	"fmt"
	"net/http"
	"strings"

	"qiantie/backend/internal/videomanagementconnector"
)

func RegisterWebSubmitTargetCheckCompat(root *http.ServeMux, middleware BridgeAuth, usersDir string, targets ...videomanagementconnector.Service) {
	if root == nil || middleware.AuthStore == nil || strings.TrimSpace(usersDir) == "" {
		return
	}
	root.Handle("POST /api/batch-rewrite/web-submit/test-visible", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"ok": false, "error": "unauthorized"})
			return
		}
		session := readUploadTargetSession(usersDir, identity.Username)
		cookie := strings.TrimSpace(toString(session["cookie"]))
		if cookie == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "请先保存账号密码并登录目标站"})
			return
		}
		valid := false
		if len(targets) > 0 && targets[0] != nil {
			err := targets[0].CheckSession(req.Context(), videomanagementconnector.Session{Cookie: cookie, TargetUsername: toString(session["targetUsername"])})
			valid = err == nil
		} else {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": "视频管理系统 Go 连接器未配置"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": valid, "output": []any{}, "result": map[string]any{"checks": map[string]any{"login_session": valid, "upload_endpoint": "/tttadmin/api/zbooklist_upload.php"}}})
	})))
	root.Handle("GET /api/batch-factory/v12/publish-121/organizations", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"ok": false, "error": "unauthorized"})
			return
		}
		session := readUploadTargetSession(usersDir, identity.Username)
		cookie := strings.TrimSpace(toString(session["cookie"]))
		if cookie == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "请先保存账号密码并登录目标站"})
			return
		}
		var payload any
		if len(targets) > 0 && targets[0] != nil {
			catalogs, err := targets[0].RefreshCatalogs(req.Context(), videomanagementconnector.Session{Cookie: cookie, TargetUsername: toString(session["targetUsername"])})
			if err != nil {
				writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "error": "视频管理系统组织目录读取失败"})
				return
			}
			payload = catalogs.Raw
		} else {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": "视频管理系统 Go 连接器未配置"})
			return
		}
		organizations := normalizeOrganizationOptions(payload)
		if len(organizations) == 0 {
			writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "error": "视频管理系统未返回可用组织目录"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "organizations": organizations})
	})))
}

func targetPageLooksAuthenticated(page string) bool {
	text := strings.TrimSpace(page)
	if text == "" {
		return false
	}
	for _, marker := range []string{"管理员登录", "登录后操作", "login.php", "name=login"} {
		if strings.Contains(strings.ToLower(text), strings.ToLower(marker)) {
			return false
		}
	}
	return true
}

func normalizeOrganizationOptions(payload any) []map[string]any {
	root := mapOrEmpty(payload)
	var value any
	if list, ok := payload.([]any); ok {
		value = list
	} else {
		value = root["data"]
	}
	if value == nil {
		value = root["organizations"]
	}
	if value == nil {
		value = root["items"]
	}
	if list, ok := value.([]any); ok {
		out := make([]map[string]any, 0, len(list))
		seen := map[string]bool{}
		for _, item := range list {
			row := mapOrEmpty(item)
			id := organizationText(row["id"], row["organization_id"], row["value"])
			name := organizationText(row["name"], row["organization_name"], row["label"])
			if id == "" || name == "" || seen[id] {
				continue
			}
			seen[id] = true
			out = append(out, map[string]any{"id": id, "name": name, "level": firstText(row["level"], row["level_name"])})
		}
		return out
	}
	return nil
}

func organizationText(values ...any) string {
	for _, value := range values {
		if value == nil {
			continue
		}
		text := strings.TrimSpace(fmt.Sprint(value))
		if text != "" && text != "<nil>" {
			return text
		}
	}
	return ""
}
