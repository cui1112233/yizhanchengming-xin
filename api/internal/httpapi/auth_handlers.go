package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
)

const (
	AccessCookieName  = "ycm_access"
	RefreshCookieName = "ycm_refresh"
)

func (h handler) login(w http.ResponseWriter, r *http.Request) {
	if h.deps.Auth == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"code": "AUTH_UNAVAILABLE", "message": "登录服务暂不可用"})
		return
	}
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil || strings.TrimSpace(input.Username) == "" || input.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"code": "AUTH_INVALID_REQUEST", "message": "用户名和密码不能为空"})
		return
	}
	input.Username = strings.TrimSpace(input.Username)
	key := limiterKey(r.RemoteAddr, input.Username)
	if h.deps.LoginLimiter != nil && !h.deps.LoginLimiter.Allow(key) {
		w.Header().Set("Retry-After", "300")
		writeJSON(w, http.StatusTooManyRequests, map[string]any{"code": "AUTH_RATE_LIMITED", "message": "登录尝试过于频繁，请稍后再试"})
		return
	}

	credentials, user, err := h.deps.Auth.Login(r.Context(), input.Username, input.Password)
	if err != nil {
		if errors.Is(err, authn.ErrUnauthenticated) {
			if h.deps.LoginLimiter != nil {
				h.deps.LoginLimiter.RecordFailure(key)
			}
			writeJSON(w, http.StatusUnauthorized, map[string]any{"code": "AUTH_UNAUTHENTICATED", "message": "用户名或密码错误"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]any{"code": "AUTH_LOGIN_FAILED", "message": "登录失败，请稍后重试"})
		return
	}
	if h.deps.LoginLimiter != nil {
		h.deps.LoginLimiter.Reset(key)
	}
	h.setAuthCookies(w, credentials)
	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

func (h handler) refreshAuth(w http.ResponseWriter, r *http.Request) {
	if h.deps.Auth == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"code": "AUTH_UNAVAILABLE", "message": "登录服务暂不可用"})
		return
	}
	cookie, err := r.Cookie(RefreshCookieName)
	if err != nil || cookie.Value == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"code": "AUTH_UNAUTHENTICATED", "message": "登录状态无效或已过期"})
		return
	}
	credentials, user, err := h.deps.Auth.Refresh(r.Context(), cookie.Value)
	if err != nil {
		// Do not emit cookie deletion here. A concurrent browser tab may already
		// have rotated the same old refresh token and installed newer cookies;
		// a late 401 must not erase that valid session.
		if errors.Is(err, authn.ErrUnauthenticated) {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"code": "AUTH_UNAUTHENTICATED", "message": "登录状态无效或已过期"})
			return
		}
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"code": "AUTH_UNAVAILABLE", "message": "登录服务暂不可用"})
		return
	}
	h.setAuthCookies(w, credentials)
	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

func (h handler) currentUser(w http.ResponseWriter, r *http.Request) {
	user, ok := authn.CurrentUser(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"code": "AUTH_UNAUTHENTICATED", "message": "登录状态无效或已过期"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

func (h handler) logout(w http.ResponseWriter, r *http.Request) {
	var accessToken, refreshToken string
	if cookie, err := r.Cookie(AccessCookieName); err == nil {
		accessToken = cookie.Value
	}
	if cookie, err := r.Cookie(RefreshCookieName); err == nil {
		refreshToken = cookie.Value
	}
	if h.deps.Auth != nil {
		if err := h.deps.Auth.Logout(r.Context(), accessToken, refreshToken); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"code": "AUTH_UNAVAILABLE", "message": "退出登录暂不可用，请稍后重试"})
			return
		}
	}
	h.clearAuthCookies(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h handler) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.deps.Auth == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"code": "AUTH_UNAVAILABLE", "message": "登录服务暂不可用"})
			return
		}
		cookie, err := r.Cookie(AccessCookieName)
		if err != nil || cookie.Value == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"code": "AUTH_UNAUTHENTICATED", "message": "登录状态无效或已过期"})
			return
		}
		user, err := h.deps.Auth.AuthenticateAccess(r.Context(), cookie.Value)
		if err != nil {
			if errors.Is(err, authn.ErrUnauthenticated) {
				writeJSON(w, http.StatusUnauthorized, map[string]any{"code": "AUTH_UNAUTHENTICATED", "message": "登录状态无效或已过期"})
				return
			}
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"code": "AUTH_UNAVAILABLE", "message": "登录服务暂不可用"})
			return
		}
		next.ServeHTTP(w, r.WithContext(authn.WithCurrentUser(r.Context(), user)))
	})
}

func (h handler) setAuthCookies(w http.ResponseWriter, credentials authn.Credentials) {
	secure := h.deps.SecureCookies
	http.SetCookie(w, &http.Cookie{
		Name: AccessCookieName, Value: credentials.AccessToken, Path: "/", HttpOnly: true,
		Secure: secure, SameSite: http.SameSiteLaxMode, MaxAge: int((15 * time.Minute).Seconds()),
	})
	http.SetCookie(w, &http.Cookie{
		Name: RefreshCookieName, Value: credentials.RefreshToken, Path: "/api/auth", HttpOnly: true,
		Secure: secure, SameSite: http.SameSiteLaxMode, MaxAge: int((30 * 24 * time.Hour).Seconds()),
	})
}

func (h handler) clearAuthCookies(w http.ResponseWriter) {
	secure := h.deps.SecureCookies
	for _, cookie := range []*http.Cookie{
		{Name: AccessCookieName, Path: "/"},
		{Name: RefreshCookieName, Path: "/api/auth"},
	} {
		cookie.Value = ""
		cookie.HttpOnly = true
		cookie.Secure = secure
		cookie.SameSite = http.SameSiteLaxMode
		cookie.MaxAge = -1
		cookie.Expires = time.Unix(1, 0).UTC()
		http.SetCookie(w, cookie)
	}
}
