package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"qiantie/backend/internal/auth"
)

func RegisterMFARoutes(root *http.ServeMux, middleware BridgeAuth, profiles *MemberProfileStore, accounts auth.Store) {
	if root == nil || profiles == nil || accounts == nil || strings.TrimSpace(profiles.SystemDir) == "" {
		return
	}
	path := filepath.Join(profiles.SystemDir, "mfa.json")
	withIdentity := func(handler http.HandlerFunc) http.Handler {
		return middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if _, ok := BridgeIdentityFromContext(req.Context()); !ok {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}
			handler(w, req)
		}))
	}
	currentPasswordOK := func(req *http.Request, username, password string) bool {
		account, err := accounts.FindAccount(req.Context(), username)
		return err == nil && auth.VerifyPassword(password, account.PasswordHash, account.PasswordAlgorithm)
	}
	root.Handle("POST /api/member/security/mfa/setup", withIdentity(func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		var input struct {
			CurrentPassword string `json:"currentPassword"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 16<<10)).Decode(&input) != nil || !currentPasswordOK(req, identity.Username, input.CurrentPassword) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "当前密码不正确"})
			return
		}
		rows, err := readMFAFile(path)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "MFA 配置读取失败"})
			return
		}
		secret, err := base32MFASecret()
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "MFA 密钥生成失败"})
			return
		}
		rows, row := mfaRowFor(rows, identity.Username)
		row.PendingSecret = secret
		if err := writeMFAFile(path, rows); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "MFA 设置保存失败"})
			return
		}
		label := url.QueryEscape("一战晟铭:" + identity.Username)
		writeJSON(w, http.StatusOK, map[string]string{"secret": secret, "otpauthUri": fmt.Sprintf("otpauth://totp/%s?secret=%s&issuer=%s&digits=6&period=30", label, secret, url.QueryEscape("一战晟铭"))})
	}))
	root.Handle("POST /api/member/security/mfa/enable", withIdentity(func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		var input struct {
			Code string `json:"code"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 16<<10)).Decode(&input) != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求内容无效"})
			return
		}
		rows, err := readMFAFile(path)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "MFA 配置读取失败"})
			return
		}
		rows, row := mfaRowFor(rows, identity.Username)
		if strings.TrimSpace(row.PendingSecret) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请先开始 MFA 设置", "code": "MFA_SETUP_REQUIRED"})
			return
		}
		if !mfaTOTP(row.PendingSecret, input.Code, time.Now()) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "动态验证码不正确", "code": "MFA_CODE_INVALID"})
			return
		}
		recoveryCodes, err := generateMFARecoveryCodes()
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "恢复码生成失败"})
			return
		}
		row.Secret, row.PendingSecret, row.Enabled, row.EnabledAt = row.PendingSecret, "", true, time.Now().UTC().Format(time.RFC3339)
		row.RecoveryHashes = make([]string, 0, len(recoveryCodes))
		for _, code := range recoveryCodes {
			row.RecoveryHashes = append(row.RecoveryHashes, mfaRecoveryHash(code))
		}
		if err := writeMFAFile(path, rows); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "MFA 设置保存失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"status": mfaStatusRow(row), "recoveryCodes": recoveryCodes})
	}))
	root.Handle("POST /api/member/security/mfa/disable", withIdentity(func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		var input struct {
			CurrentPassword string `json:"currentPassword"`
			Code            string `json:"code"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 16<<10)).Decode(&input) != nil || !currentPasswordOK(req, identity.Username, input.CurrentPassword) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "当前密码不正确"})
			return
		}
		required, valid, _, err := verifyMFAForLogin(path, identity.Username, input.Code, time.Now())
		if err != nil || !required || !valid {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "动态验证码或恢复码不正确", "code": "MFA_CODE_INVALID"})
			return
		}
		rows, _ := readMFAFile(path)
		rows, row := mfaRowFor(rows, identity.Username)
		row.Enabled, row.Secret, row.PendingSecret, row.RecoveryHashes = false, "", "", nil
		if err := writeMFAFile(path, rows); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "MFA 关闭失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"status": mfaStatusRow(row)})
	}))
	root.Handle("POST /api/member/security/mfa/recovery-codes", withIdentity(func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		var input struct {
			CurrentPassword string `json:"currentPassword"`
			Code            string `json:"code"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 16<<10)).Decode(&input) != nil || !currentPasswordOK(req, identity.Username, input.CurrentPassword) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "当前密码不正确"})
			return
		}
		required, valid, _, err := verifyMFAForLogin(path, identity.Username, input.Code, time.Now())
		if err != nil || !required || !valid {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "动态验证码或恢复码不正确", "code": "MFA_CODE_INVALID"})
			return
		}
		recoveryCodes, err := generateMFARecoveryCodes()
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "恢复码生成失败"})
			return
		}
		rows, _ := readMFAFile(path)
		rows, row := mfaRowFor(rows, identity.Username)
		row.RecoveryHashes = make([]string, 0, len(recoveryCodes))
		for _, code := range recoveryCodes {
			row.RecoveryHashes = append(row.RecoveryHashes, mfaRecoveryHash(code))
		}
		if err := writeMFAFile(path, rows); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "恢复码保存失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"recoveryCodes": recoveryCodes, "recoveryCodesRemaining": len(recoveryCodes)})
	}))
}

func mfaRowFor(rows []mfaFileRow, username string) ([]mfaFileRow, *mfaFileRow) {
	for index := range rows {
		if rows[index].Username == username {
			return rows, &rows[index]
		}
	}
	rows = append(rows, mfaFileRow{Username: username, RecoveryHashes: []string{}})
	return rows, &rows[len(rows)-1]
}
