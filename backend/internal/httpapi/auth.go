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

	root.HandleFunc("GET /api/login/session", func(w http.ResponseWriter, r *http.Request) {
		if options.Store == nil {
			writeAuthError(w, http.StatusServiceUnavailable, "登录服务暂不可用")
			return
		}
		token := bearerToken(r)
		session, err := options.Store.FindSession(r.Context(), token, now().UTC())
		if err != nil {
			if options.SessionResolver != nil {
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
