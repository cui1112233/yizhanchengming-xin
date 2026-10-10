package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"qiantie/backend/internal/auth"
)

type AuthOptions struct {
	Store           auth.Store
	Now             func() time.Time
	SessionResolver func(context.Context, string) (BridgeIdentity, error)
	MFAFile         string
	PasskeyFile     string
}

func RegisterAuthRoutes(root *http.ServeMux, options AuthOptions) {
	now := options.Now
	if now == nil {
		now = time.Now
	}
	root.HandleFunc("POST /api/login", func(w http.ResponseWriter, r *http.Request) {
		if options.Store == nil {
			writeAuthError(w, http.StatusServiceUnavailable, "登录服务暂不可用")
			return
		}
		var input struct {
			Username string `json:"username"`
			Password string `json:"password"`
			Remember bool   `json:"remember"`
			MFACode  string `json:"mfaCode"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10)).Decode(&input); err != nil || strings.TrimSpace(input.Username) == "" || input.Password == "" {
			writeAuthError(w, http.StatusUnauthorized, "用户名或密码错误")
			return
		}
		account, err := options.Store.FindAccount(r.Context(), strings.TrimSpace(input.Username))
		if err != nil || !account.Active || !auth.VerifyPassword(input.Password, account.PasswordHash, account.PasswordAlgorithm) {
			writeAuthError(w, http.StatusUnauthorized, "用户名或密码错误")
			return
		}
		if strings.TrimSpace(options.MFAFile) != "" {
			required, valid, recoveryUsed, verifyErr := verifyMFAForLogin(options.MFAFile, account.Username, input.MFACode, now())
			if verifyErr != nil {
				writeAuthError(w, http.StatusServiceUnavailable, "MFA 服务暂不可用")
				return
			}
			if required && !valid {
				writeJSON(w, http.StatusPreconditionRequired, map[string]string{"error": "请输入 MFA 动态验证码或恢复码", "code": "MFA_REQUIRED"})
				return
			}
			_ = recoveryUsed
		}
		token, err := newSessionToken()
		if err != nil {
			writeAuthError(w, http.StatusInternalServerError, "登录服务暂不可用")
			return
		}
		issued := now().UTC()
		expires := issued.Add(24 * time.Hour)
		if input.Remember {
			expires = issued.Add(30 * 24 * time.Hour)
		}
		if err := options.Store.CreateSession(r.Context(), auth.Session{Token: token, Username: account.Username, IssuedAt: issued, ExpiresAt: expires, IPHint: r.RemoteAddr}); err != nil {
			writeAuthError(w, http.StatusUnauthorized, "用户名或密码错误")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"token": token, "username": account.Username, "active": account.Active, "isOwner": account.IsOwner, "effectivePermissions": []string{}})
	})

	root.HandleFunc("POST /api/login/logout", func(w http.ResponseWriter, r *http.Request) {
		if token := bearerToken(r); token != "" && options.Store != nil {
			_ = options.Store.RevokeSession(r.Context(), token)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	root.HandleFunc("POST /api/login/passkey/options", func(w http.ResponseWriter, r *http.Request) {
		if options.Store == nil || strings.TrimSpace(options.PasskeyFile) == "" {
			writeAuthError(w, http.StatusServiceUnavailable, "Passkey 服务暂不可用")
			return
		}
		var input struct {
			Username string `json:"username"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10)).Decode(&input) != nil || strings.TrimSpace(input.Username) == "" {
			writeAuthError(w, http.StatusBadRequest, "请输入账号")
			return
		}
		account, err := options.Store.FindAccount(r.Context(), strings.TrimSpace(input.Username))
		if err != nil || !account.Active {
			writeAuthError(w, http.StatusNotFound, "当前账号没有可用 Passkey")
			return
		}
		result, err := passkeyAuthenticationOptions(options.PasskeyFile, account.Username, r)
		if err != nil {
			writeAuthError(w, http.StatusNotFound, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, result)
	})

	root.HandleFunc("GET /api/login/session", func(w http.ResponseWriter, r *http.Request) {
		if options.Store == nil {
			writeAuthError(w, http.StatusServiceUnavailable, "登录服务暂不可用")
			return
		}
		token := bearerToken(r)
		session, err := options.Store.FindSession(r.Context(), token, now().UTC())
		if err != nil {
			// A legacy Node request may ask Go to validate a Go-issued token.
			// Do not bounce that failed lookup back to Node, or the two services
			// would recursively validate the same invalid token.
			if options.SessionResolver != nil && r.Header.Get("X-Qiantie-Auth-Bridge") != "1" {
				if identity, resolveErr := options.SessionResolver(r.Context(), token); resolveErr == nil {
					writeJSON(w, http.StatusOK, map[string]any{"username": identity.Username, "active": true, "isOwner": identity.IsOwner, "effectivePermissions": []string{}})
					return
				}
			}
			w.Header().Set("X-Qiantie-Auth-Failure", "session")
			writeAuthError(w, http.StatusUnauthorized, "登录已失效，请重新登录")
			return
		}
		account, err := options.Store.FindAccount(r.Context(), session.Username)
		if err != nil || !account.Active {
			w.Header().Set("X-Qiantie-Auth-Failure", "session")
			writeAuthError(w, http.StatusUnauthorized, "登录已失效，请重新登录")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"username": account.Username, "active": account.Active, "isOwner": account.IsOwner, "effectivePermissions": []string{}})
	})
}

func bearerToken(r *http.Request) string {
	value := strings.TrimSpace(r.Header.Get("Authorization"))
	if len(value) < 8 || !strings.EqualFold(value[:7], "Bearer ") {
		return ""
	}
	return strings.TrimSpace(value[7:])
}

func newSessionToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func writeAuthError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
