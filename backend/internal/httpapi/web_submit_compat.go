package httpapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func RegisterWebSubmitTargetCheckCompat(root *http.ServeMux, middleware BridgeAuth, usersDir string) {
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
		check, err := http.NewRequestWithContext(req.Context(), http.MethodGet, "http://two.121w.com/tttadmin/zidingyi.php", nil)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "验证失败"})
			return
		}
		check.Header.Set("Cookie", cookie)
		response, err := (&http.Client{Timeout: 30 * time.Second}).Do(check)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "目标站连接失败"})
			return
		}
		defer response.Body.Close()
		body, _ := io.ReadAll(io.LimitReader(response.Body, 1<<20))
		page := string(body)
		valid := response.StatusCode >= 200 && response.StatusCode < 300 && strings.Contains(page, "自定义文案") && !strings.Contains(page, "管理员登录")
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
		body, err := targetGET(req, "http://two.121w.com/tttadmin/api/organization.php", cookie)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "error": "视频管理系统组织目录读取失败"})
			return
		}
		var payload any
		if err := json.Unmarshal(body, &payload); err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "error": "视频管理系统组织目录返回了无效数据"})
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
