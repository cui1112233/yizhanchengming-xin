package httpapi

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"qiantie/backend/internal/videomanagementconnector"
)

// RegisterNovelFetchUploadSessionCompat moves the read-only target-site
// session check to Go while preserving the legacy upload-target.json format.
// The cookie is still verified against the real target site; a local file alone
// is never treated as proof of a live login.
func RegisterNovelFetchUploadSessionCompat(root *http.ServeMux, middleware BridgeAuth, usersDir string, targets ...videomanagementconnector.Service) {
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
		if len(targets) == 0 || targets[0] == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "loggedIn": false, "error": "视频管理系统 Go 连接器未配置"})
			return
		}
		if err := targets[0].CheckSession(req.Context(), videomanagementconnector.Session{Cookie: cookie, TargetUsername: toString(session["targetUsername"])}); err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"ok": false, "loggedIn": false, "notLoggedIn": true, "status": "expired", "lastVerifiedAt": nil, "error": "目标站登录已失效，请重新登录", "code": "SESSION_EXPIRED"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "loggedIn": true, "status": "ready", "lastVerifiedAt": time.Now().UTC().Format(time.RFC3339)})
	})))
	root.Handle("POST /api/novel-fetch-upload/upload-login", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 16<<10)).Decode(&input) != nil || strings.TrimSpace(input.Username) == "" || input.Password == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "请输入账号和密码"})
			return
		}
		if len(targets) == 0 || targets[0] == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": "视频管理系统 Go 连接器未配置"})
			return
		}
		connected, err := targets[0].Login(req.Context(), videomanagementconnector.LoginInput{Username: strings.TrimSpace(input.Username), Password: input.Password})
		if err != nil || strings.TrimSpace(connected.Cookie) == "" {
			writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "error": "121 登录失败"})
			return
		}
		if err := saveUploadTargetSession(usersDir, identity.Username, map[string]any{"mode": "go_connector", "cookie": connected.Cookie, "sessionKey": connected.Cookie, "targetUsername": connected.TargetUsername, "baseUrl": "http://two.121w.com/tttadmin", "status": "ready", "loginAt": time.Now().UTC().Format(time.RFC3339)}); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "上传会话保存失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "username": strings.TrimSpace(input.Username), "status": "ready"})
	})))
}

func saveUploadTargetSession(usersDir, username string, value map[string]any) error {
	directory := filepath.Join(usersDir, username)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(directory, "upload-target.json"), raw, 0o600)
}
