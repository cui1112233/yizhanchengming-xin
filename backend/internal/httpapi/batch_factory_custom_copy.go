package httpapi

import (
	"context"
	"errors"
	"net/http"
	"qiantie/backend/internal/videomanagementconnector"
	"strings"
	"time"
)

func registerBatchFactoryCustomCopy(root *http.ServeMux, auth BridgeAuth, usersDir string, targets ...videomanagementconnector.Service) {
	root.Handle("GET /api/batch-factory/v12/custom-copies/{bookId}", auth.SessionOnly().Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity, ok := BridgeIdentityFromContext(r.Context())
		if !ok {
			writeJSON(w, 401, map[string]string{"error": "请先登录"})
			return
		}
		if len(targets) == 0 {
			writeJSON(w, 503, map[string]string{"error": "文案后台未配置"})
			return
		}
		reader, ok := targets[0].(videomanagementconnector.CustomCopyReader)
		if !ok {
			writeJSON(w, 503, map[string]string{"error": "文案后台暂不支持读取"})
			return
		}
		saved := readUploadTargetSession(usersDir, identity.Username)
		session := videomanagementconnector.Session{Cookie: strings.TrimSpace(toString(saved["cookie"])), TargetUsername: toString(saved["targetUsername"])}
		ctx, cancel := context.WithTimeout(r.Context(), 55*time.Second)
		defer cancel()
		copies, err := reader.CustomCopies(ctx, session, r.PathValue("bookId"))
		if err != nil {
			message := err.Error()
			status := http.StatusBadGateway
			if errors.Is(err, videomanagementconnector.ErrSessionMissing) || errors.Is(err, videomanagementconnector.ErrSessionExpired) {
				message = "文案后台尚未登录或登录已失效，请在上传设置中登录视频管理后台"
				status = 409
			}
			writeJSON(w, status, map[string]string{"error": message})
			return
		}
		writeJSON(w, 200, map[string]any{"copies": copies})
	})))
}
