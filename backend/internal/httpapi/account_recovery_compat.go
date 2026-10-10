package httpapi

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"qiantie/backend/internal/auth"
)

// RegisterAccountRecoveryStatusCompat serves the harmless read-only portion of
// the recovery contract from the Go-owned profile files. Mail delivery,
// password reset, and passkey operations remain on their existing service.
func RegisterAccountRecoveryStatusCompat(root *http.ServeMux, middleware BridgeAuth, systemDir string) {
	if root == nil || middleware.AuthStore == nil || systemDir == "" {
		return
	}
	root.Handle("GET /api/account-recovery/email/status", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		email := recoveryProfileEmail(systemDir, identity.Username)
		verified, verifiedAt := false, any(nil)
		if raw, err := os.ReadFile(filepath.Join(systemDir, "account-recovery.json")); err == nil {
			var state map[string]any
			if json.Unmarshal(raw, &state) == nil {
				for _, value := range recoveryVerifiedEmails(state) {
					record, ok := value.(map[string]any)
					if !ok || fmt.Sprint(record["username"]) != identity.Username || email == "" || fmt.Sprint(record["email"]) != email {
						continue
					}
					verified = true
					verifiedAt = record["verifiedAt"]
					break
				}
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{"email": email, "verified": verified, "verifiedAt": verifiedAt, "deliveryConfigured": strings.TrimSpace(os.Getenv("QIANTIE_MAIL_ENDPOINT")) != "" && strings.TrimSpace(os.Getenv("QIANTIE_MAIL_FROM")) != ""})
	})))
	root.Handle("POST /api/account-recovery/email/request", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		endpoint := strings.TrimSpace(os.Getenv("QIANTIE_MAIL_ENDPOINT"))
		from := strings.TrimSpace(os.Getenv("QIANTIE_MAIL_FROM"))
		if endpoint == "" || from == "" {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "邮件发送通道尚未配置", "code": "MAIL_UNAVAILABLE"})
			return
		}
		email := recoveryProfileEmail(systemDir, identity.Username)
		if email == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请先在个人资料中填写邮箱"})
			return
		}
		token, expiresAt, err := issueRecoveryToken(filepath.Join(systemDir, "account-recovery.json"), "email", identity.Username, email)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "验证链接生成失败"})
			return
		}
		verifyURL := recoveryAbsoluteURL(req, "/recover?verify="+url.QueryEscape(token))
		if err := sendRecoveryMail(req.Context(), endpoint, os.Getenv("QIANTIE_MAIL_TOKEN"), from, email, "验证你的一战晟铭邮箱", "请在 30 分钟内打开以下链接完成邮箱验证：\n"+verifyURL+"\n\n如果不是你本人操作，请忽略此邮件。", ""); err != nil {
			_, _ = revokeRecoveryToken(filepath.Join(systemDir, "account-recovery.json"), "email", identity.Username)
			status := http.StatusBadGateway
			if err.Error() == "MAIL_UNAVAILABLE" {
				status = http.StatusServiceUnavailable
			}
			writeJSON(w, status, map[string]string{"error": "邮件服务请求失败", "code": "MAIL_DELIVERY_FAILED"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"sent": true, "email": email, "expiresAt": expiresAt})
	})))
	root.Handle("POST /api/account-recovery/password/request", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		endpoint := strings.TrimSpace(os.Getenv("QIANTIE_MAIL_ENDPOINT"))
		from := strings.TrimSpace(os.Getenv("QIANTIE_MAIL_FROM"))
		if endpoint == "" || from == "" {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "邮件发送通道尚未配置", "code": "MAIL_UNAVAILABLE"})
			return
		}
		var input struct {
			Email string `json:"email"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 16<<10)).Decode(&input) != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求内容无效"})
			return
		}
		email := strings.TrimSpace(strings.ToLower(input.Email))
		if email == "" {
			writeJSON(w, http.StatusAccepted, map[string]any{"accepted": true, "message": "如果该邮箱已验证并绑定账号，重置邮件将会发送。"})
			return
		}
		state := map[string]any{}
		path := filepath.Join(systemDir, "account-recovery.json")
		if raw, err := os.ReadFile(path); err == nil {
			_ = json.Unmarshal(raw, &state)
		}
		username := ""
		for _, value := range recoveryVerifiedEmails(state) {
			row, ok := value.(map[string]any)
			if ok && strings.EqualFold(fmt.Sprint(row["email"]), email) {
				username = fmt.Sprint(row["username"])
				break
			}
		}
		if username != "" {
			account, accountErr := middleware.AuthStore.FindAccount(req.Context(), username)
			if accountErr == nil && account.Active && recoveryProfileEmail(systemDir, username) == email {
				token, _, issueErr := issueRecoveryToken(path, "password", username, email)
				if issueErr != nil {
					writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "重置链接生成失败"})
					return
				}
				resetURL := recoveryAbsoluteURL(req, "/recover?reset="+url.QueryEscape(token))
				if err := sendRecoveryMail(req.Context(), endpoint, os.Getenv("QIANTIE_MAIL_TOKEN"), from, email, "重置你的一战晟铭密码", "请在 30 分钟内打开以下链接设置新密码：\n"+resetURL+"\n\n该链接只能使用一次。如果不是你本人申请，请忽略此邮件。", ""); err != nil {
					_, _ = revokeRecoveryToken(path, "password", username)
					writeJSON(w, http.StatusBadGateway, map[string]string{"error": "邮件服务请求失败", "code": "MAIL_DELIVERY_FAILED"})
					return
				}
			}
		}
		writeJSON(w, http.StatusAccepted, map[string]any{"accepted": true, "message": "如果该邮箱已验证并绑定账号，重置邮件将会发送。"})
	}))
	root.Handle("POST /api/account-recovery/email/verify", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var input struct {
			Token string `json:"token"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 16<<10)).Decode(&input) != nil || strings.TrimSpace(input.Token) == "" {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "验证链接无效"})
			return
		}
		path := filepath.Join(systemDir, "account-recovery.json")
		state := map[string]any{}
		if raw, err := os.ReadFile(path); err == nil {
			if json.Unmarshal(raw, &state) != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "账号找回状态损坏"})
				return
			}
		}
		digest := sha256.Sum256([]byte(strings.TrimSpace(input.Token)))
		digestText := hex.EncodeToString(digest[:])
		now := time.Now().UTC()
		var matched map[string]any
		for _, value := range recoveryEmailTokens(state) {
			row, ok := value.(map[string]any)
			if !ok || fmt.Sprint(row["tokenDigest"]) != digestText {
				continue
			}
			if recoveryTokenUsed(row["usedAt"]) {
				writeJSON(w, http.StatusConflict, map[string]string{"error": "验证链接已使用"})
				return
			}
			expires, err := time.Parse(time.RFC3339, fmt.Sprint(row["expiresAt"]))
			if err != nil || !expires.After(now) {
				writeJSON(w, http.StatusConflict, map[string]string{"error": "验证链接已过期"})
				return
			}
			matched = row
			break
		}
		if matched == nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "验证链接无效"})
			return
		}
		username := fmt.Sprint(matched["username"])
		email := fmt.Sprint(matched["email"])
		matched["usedAt"] = now.Format(time.RFC3339)
		verified := make([]any, 0)
		for _, value := range recoveryVerifiedEmails(state) {
			row, ok := value.(map[string]any)
			if !ok || fmt.Sprint(row["username"]) != username {
				verified = append(verified, value)
			}
		}
		verified = append(verified, map[string]any{"username": username, "email": email, "verifiedAt": now.Format(time.RFC3339)})
		state["verifiedEmails"] = verified
		if err := writeRecoveryState(path, state); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "邮箱验证状态保存失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"verified": true, "email": email, "verifiedAt": now.Format(time.RFC3339)})
	}))
	root.Handle("POST /api/account-recovery/password/reset", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var input struct {
			Token    string `json:"token"`
			Password string `json:"password"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 16<<10)).Decode(&input) != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求内容无效"})
			return
		}
		if len(input.Password) < 8 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "新密码至少需要 8 位"})
			return
		}
		path := filepath.Join(systemDir, "account-recovery.json")
		state := map[string]any{}
		if raw, err := os.ReadFile(path); err == nil {
			if json.Unmarshal(raw, &state) != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "账号找回状态损坏"})
				return
			}
		}
		digest := sha256.Sum256([]byte(strings.TrimSpace(input.Token)))
		digestText := hex.EncodeToString(digest[:])
		now := time.Now().UTC()
		var matched map[string]any
		for _, value := range recoveryPasswordTokens(state) {
			row, ok := value.(map[string]any)
			if !ok || fmt.Sprint(row["tokenDigest"]) != digestText {
				continue
			}
			if recoveryTokenUsed(row["usedAt"]) {
				writeJSON(w, http.StatusConflict, map[string]string{"error": "重置链接已使用"})
				return
			}
			expires, err := time.Parse(time.RFC3339, fmt.Sprint(row["expiresAt"]))
			if err != nil || !expires.After(now) {
				writeJSON(w, http.StatusConflict, map[string]string{"error": "重置链接已过期"})
				return
			}
			matched = row
			break
		}
		if matched == nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "重置链接无效"})
			return
		}
		username := fmt.Sprint(matched["username"])
		account, err := middleware.AuthStore.FindAccount(req.Context(), username)
		if err != nil || !account.Active {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "账号当前不可用"})
			return
		}
		hash, algorithm, err := auth.HashPassword(input.Password)
		if err != nil || middleware.AuthStore.UpdatePassword(req.Context(), username, hash, algorithm) != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "密码保存失败"})
			return
		}
		matched["usedAt"] = now.Format(time.RFC3339)
		if err := writeRecoveryState(path, state); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "重置状态保存失败"})
			return
		}
		var revoked int64
		if revoker, ok := middleware.AuthStore.(interface {
			RevokeOtherSessions(context.Context, string, string) (int64, error)
		}); ok {
			revoked, _ = revoker.RevokeOtherSessions(req.Context(), username, "")
		}
		writeJSON(w, http.StatusOK, map[string]any{"reset": true, "revoked": revoked})
	}))
}

func recoveryVerifiedEmails(state map[string]any) []any {
	values, _ := state["verifiedEmails"].([]any)
	return values
}

func recoveryProfileEmail(systemDir, username string) string {
	raw, err := os.ReadFile(filepath.Join(systemDir, "member-profile-details.json"))
	if err != nil {
		return ""
	}
	var rows []map[string]any
	if json.Unmarshal(raw, &rows) != nil {
		return ""
	}
	for _, row := range rows {
		if fmt.Sprint(row["username"]) == username {
			return strings.TrimSpace(strings.ToLower(fmt.Sprint(row["email"])))
		}
	}
	return ""
}

func recoveryAbsoluteURL(req *http.Request, path string) string {
	scheme := strings.TrimSpace(req.Header.Get("X-Forwarded-Proto"))
	if scheme != "http" && scheme != "https" {
		scheme = "http"
	}
	return scheme + "://" + req.Host + path
}

func issueRecoveryToken(path, kind, username, email string) (string, string, error) {
	state := map[string]any{}
	if raw, err := os.ReadFile(path); err == nil {
		if json.Unmarshal(raw, &state) != nil {
			return "", "", fmt.Errorf("invalid recovery state")
		}
	}
	if state["emailTokens"] == nil {
		state["emailTokens"] = []any{}
	}
	if state["passwordResetTokens"] == nil {
		state["passwordResetTokens"] = []any{}
	}
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", "", err
	}
	token := base64.RawURLEncoding.EncodeToString(tokenBytes)
	digest := sha256.Sum256([]byte(token))
	now := time.Now().UTC()
	record := map[string]any{"id": fmt.Sprintf("go-%d", now.UnixNano()), "username": username, "email": email, "tokenDigest": hex.EncodeToString(digest[:]), "createdAt": now.Format(time.RFC3339), "expiresAt": now.Add(30 * time.Minute).Format(time.RFC3339), "usedAt": nil}
	key := "emailTokens"
	if kind == "password" {
		key = "passwordResetTokens"
	}
	rows, _ := state[key].([]any)
	next := make([]any, 0, len(rows)+1)
	for _, value := range rows {
		row, ok := value.(map[string]any)
		if ok && fmt.Sprint(row["username"]) == username && !recoveryTokenUsed(row["usedAt"]) {
			continue
		}
		next = append(next, value)
	}
	state[key] = append(next, record)
	if err := writeRecoveryState(path, state); err != nil {
		return "", "", err
	}
	return token, record["expiresAt"].(string), nil
}

func revokeRecoveryToken(path, kind, username string) (int, error) {
	state := map[string]any{}
	if raw, err := os.ReadFile(path); err == nil {
		if json.Unmarshal(raw, &state) != nil {
			return 0, err
		}
	}
	key := "emailTokens"
	if kind == "password" {
		key = "passwordResetTokens"
	}
	rows, _ := state[key].([]any)
	next := make([]any, 0, len(rows))
	removed := 0
	for _, value := range rows {
		row, ok := value.(map[string]any)
		if ok && fmt.Sprint(row["username"]) == username && !recoveryTokenUsed(row["usedAt"]) {
			removed++
			continue
		}
		next = append(next, value)
	}
	if removed > 0 {
		state[key] = next
		return removed, writeRecoveryState(path, state)
	}
	return 0, nil
}

func sendRecoveryMail(ctx context.Context, endpoint, token, from, to, subject, body, html string) error {
	if strings.TrimSpace(endpoint) == "" || strings.TrimSpace(from) == "" {
		return fmt.Errorf("MAIL_UNAVAILABLE")
	}
	payload, _ := json.Marshal(map[string]any{"from": from, "to": to, "subject": subject, "text": body, "html": html})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if strings.TrimSpace(token) != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(token))
	}
	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("MAIL_DELIVERY_FAILED")
	}
	return nil
}

func recoveryEmailTokens(state map[string]any) []any {
	values, _ := state["emailTokens"].([]any)
	return values
}

func recoveryPasswordTokens(state map[string]any) []any {
	values, _ := state["passwordResetTokens"].([]any)
	return values
}

func recoveryTokenUsed(value any) bool {
	if value == nil {
		return false
	}
	text := strings.TrimSpace(fmt.Sprint(value))
	return text != "" && text != "<nil>"
}

func writeRecoveryState(path string, state map[string]any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o600)
}
