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
	registerBatchFactoryCustomCopy(root, middleware, usersDir, targets...)
	sessionHandler := middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		path, pathErr := UserScopedPath(usersDir, identity.Username, "upload-target.json")
		if pathErr != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "用户名不合法"})
			return
		}
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
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "loggedIn": true, "status": "ready", "username": strings.TrimSpace(toString(session["targetUsername"])), "lastVerifiedAt": time.Now().UTC().Format(time.RFC3339)})
	}))
	root.Handle("GET /api/novel-fetch-upload/upload-session", sessionHandler)
	root.Handle("GET /api/novel-fetch-workshop/upload/session", sessionHandler)

	loginHandler := middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
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
			writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "error": videoManagementLoginError(err, connected.Cookie)})
			return
		}
		if err := saveUploadTargetSession(usersDir, identity.Username, map[string]any{"mode": "go_connector", "cookie": connected.Cookie, "sessionKey": connected.Cookie, "targetUsername": connected.TargetUsername, "baseUrl": "https://two.121w.com/tttadmin", "status": "ready", "loginAt": time.Now().UTC().Format(time.RFC3339)}); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "上传会话保存失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "username": strings.TrimSpace(input.Username), "status": "ready"})
	}))
	root.Handle("POST /api/novel-fetch-upload/upload-login", loginHandler)
	root.Handle("POST /api/novel-fetch-workshop/upload/login", loginHandler)
}

func videoManagementLoginError(err error, cookie string) string {
	if err == nil && strings.TrimSpace(cookie) == "" {
		return "视频管理系统未返回登录会话，请稍后重试"
	}
	message := strings.ToLower(strings.TrimSpace(err.Error()))
	if strings.Contains(message, "login page") && strings.Contains(message, "http 404") {
		return "视频管理系统登录入口不可用，请检查后台地址"
	}
	if strings.Contains(message, "login rejected") {
		return "视频管理系统账号或密码错误"
	}
	if strings.Contains(message, "connection refused") || strings.Contains(message, "connectex") || strings.Contains(message, "no such host") || strings.Contains(message, "timeout") || strings.Contains(message, "deadline exceeded") {
		return "暂时无法连接视频管理系统，请稍后重试"
	}
	return "视频管理系统登录暂时不可用，请稍后重试"
}

func saveUploadTargetSession(usersDir, username string, value map[string]any) error {
	directory, err := UserScopedPath(usersDir, username)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(directory, "upload-target.json"), raw, 0o600)
}
