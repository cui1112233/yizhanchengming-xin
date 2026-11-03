package httpapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"qiantie/backend/internal/videomanagementconnector"
)

// 视频管理系统凭证的四种可区分状态。状态接口、环境自检、同步配置档/风格目录
// 都只读取 resolveTargetCredential 的结果，避免出现“页面显示已登录并验证、
// 同步却报登录已失效”的两套判定。
const (
	TargetCredentialVerified    = "verified"
	TargetCredentialNotLoggedIn = "not_logged_in"
	TargetCredentialIncomplete  = "incomplete"
	TargetCredentialExpired     = "expired"
	TargetCredentialUnavailable = "unavailable"
)

type targetCredential struct {
	State     string                           `json:"state"`
	Code      string                           `json:"code,omitempty"`
	Message   string                           `json:"message"`
	Username  string                           `json:"username,omitempty"`
	CheckedAt string                           `json:"checkedAt"`
	Session   videomanagementconnector.Session `json:"-"`
}

func (c targetCredential) verified() bool { return c.State == TargetCredentialVerified }

// resolveTargetCredential 是目标站凭证解析的唯一入口：读取同一份
// upload-target.json 会话，调用连接器同一个 CheckSession，并把结果归类为
// 未登录 / 凭证不全 / 已过期 / 不可用 / 已验证。调用方不得再自行判断
// cookie 或登录页。
func resolveTargetCredential(ctx context.Context, session map[string]any, accountUsername string, targets ...videomanagementconnector.Service) targetCredential {
	cookie := strings.TrimSpace(toString(session["cookie"]))
	targetUsername := strings.TrimSpace(toString(session["targetUsername"]))
	if strings.TrimSpace(accountUsername) == "" {
		accountUsername = targetUsername
	}
	result := targetCredential{Username: strings.TrimSpace(accountUsername), CheckedAt: time.Now().UTC().Format(time.RFC3339)}
	if cookie == "" {
		if result.Username != "" {
			result.State = TargetCredentialIncomplete
			result.Code = ErrorCodeTargetCredentialIncomplete
			result.Message = MessageTargetCredentialIncomplete
		} else {
			result.State = TargetCredentialNotLoggedIn
			result.Code = ErrorCodeTargetNotLoggedIn
			result.Message = MessageTargetNotLoggedIn
		}
		return result
	}
	result.Session = videomanagementconnector.Session{Cookie: cookie, TargetUsername: targetUsername}
	if len(targets) == 0 || targets[0] == nil {
		result.State = TargetCredentialUnavailable
		result.Code = ErrorCodeTargetConnectorUnavailable
		result.Message = MessageTargetConnectorUnavailable
		return result
	}
	err := targets[0].CheckSession(ctx, result.Session)
	switch {
	case err == nil:
		result.State = TargetCredentialVerified
		result.Message = MessageTargetSessionVerified
	case errors.Is(err, videomanagementconnector.ErrSessionExpired):
		result.State = TargetCredentialExpired
		result.Code = ErrorCodeTargetSessionExpired
		result.Message = MessageTargetSessionExpired
	case errors.Is(err, videomanagementconnector.ErrSessionMissing):
		result.State = TargetCredentialNotLoggedIn
		result.Code = ErrorCodeTargetNotLoggedIn
		result.Message = MessageTargetNotLoggedIn
	default:
		result.State = TargetCredentialUnavailable
		result.Code = ErrorCodeTargetConnectorUnavailable
		result.Message = fmt.Sprintf("%s（%v）", MessageTargetConnectorUnavailable, err)
	}
	return result
}

// writeTargetCredentialError 让状态接口之外的同步类接口复用同一套状态文案与
// 稳定错误码，前端不会再收到一句笼统的“登录已失效”。
func writeTargetCredentialError(w http.ResponseWriter, credential targetCredential) {
	status := http.StatusBadRequest
	switch credential.State {
	case TargetCredentialExpired:
		status = http.StatusUnauthorized
	case TargetCredentialUnavailable:
		status = http.StatusBadGateway
	}
	writeJSON(w, status, map[string]any{"ok": false, "error": credential.Message, "code": credential.Code, "credential": credential})
}

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
		credential := resolveTargetCredential(req.Context(), session, "", targets...)
		valid := credential.verified()
		writeJSON(w, http.StatusOK, map[string]any{"ok": valid, "credential": credential, "output": []any{}, "result": map[string]any{"checks": map[string]any{"login_session": valid, "upload_endpoint": "/tttadmin/api/zbooklist_upload.php"}}})
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
