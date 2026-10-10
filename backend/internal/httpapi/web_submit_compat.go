package httpapi

import (
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
}
