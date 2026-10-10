package httpapi

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// RegisterNovelFetchUploadSessionCompat moves the read-only target-site
// session check to Go while preserving the legacy upload-target.json format.
// The cookie is still verified against the real target site; a local file alone
// is never treated as proof of a live login.
func RegisterNovelFetchUploadSessionCompat(root *http.ServeMux, middleware BridgeAuth, usersDir string) {
	if root == nil || middleware.AuthStore == nil || strings.TrimSpace(usersDir) == "" {
		return
	}
	root.Handle("GET /api/novel-fetch-upload/upload-session", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		path := filepath.Join(usersDir, identity.Username, "upload-target.json")
		session := map[string]any{}
		raw, err := os.ReadFile(path)
		if err != nil || json.Unmarshal(raw, &session) != nil || strings.TrimSpace(toString(session["cookie"])) == "" {
			writeJSON(w, http.StatusOK, map[string]any{"ok": true, "loggedIn": false, "status": "missing", "lastVerifiedAt": nil})
			return
		}
		cookie := strings.TrimSpace(toString(session["cookie"]))
		check, err := http.NewRequestWithContext(req.Context(), http.MethodGet, "http://two.121w.com/tttadmin/zidingyi.php", nil)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "error": "121 会话验证失败"})
			return
		}
		check.Header.Set("Cookie", cookie)
		client := &http.Client{Timeout: 30 * time.Second}
		response, err := client.Do(check)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "loggedIn": false, "error": "121 会话验证失败"})
			return
		}
		defer response.Body.Close()
		if response.StatusCode == http.StatusUnauthorized || response.StatusCode == http.StatusForbidden {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"ok": false, "loggedIn": false, "notLoggedIn": true, "status": "expired", "lastVerifiedAt": nil, "error": "目标站登录已失效，请重新登录", "code": "SESSION_EXPIRED"})
			return
		}
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "loggedIn": false, "error": "121 会话验证失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "loggedIn": true, "status": "ready", "lastVerifiedAt": time.Now().UTC().Format(time.RFC3339)})
	})))
}
