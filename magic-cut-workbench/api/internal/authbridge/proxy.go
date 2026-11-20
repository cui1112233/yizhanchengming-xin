package authbridge

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const accountCookieName = "magic_cut_account"

var allowedPaths = map[string]struct{}{
	"/api/login":                 {},
	"/api/login/session":         {},
	"/api/login/logout":          {},
	"/api/login/passkey/options": {},
	"/api/login/passkey/verify":  {},
}

type proxy struct {
	baseURL *url.URL
	client  *http.Client
}

func New(baseURL string, client *http.Client) http.Handler {
	parsed, err := url.Parse(strings.TrimRight(baseURL, "/"))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			writeError(w, http.StatusServiceUnavailable, "ACCOUNT_SERVICE_UNAVAILABLE", "账号服务地址无效")
		})
	}
	if client == nil {
		client = &http.Client{}
	}
	clientCopy := *client
	if clientCopy.Timeout == 0 {
		clientCopy.Timeout = 15 * time.Second
	}
	return &proxy{baseURL: parsed, client: &clientCopy}
}

func (p *proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if _, ok := allowedPaths[r.URL.Path]; !ok {
		writeError(w, http.StatusNotFound, "AUTH_BRIDGE_PATH_NOT_ALLOWED", "该路径不属于账号桥接")
		return
	}

	target := *p.baseURL
	target.Path = r.URL.Path
	target.RawQuery = r.URL.RawQuery
	body := io.LimitReader(r.Body, 1<<20)
	request, err := http.NewRequestWithContext(r.Context(), r.Method, target.String(), body)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "ACCOUNT_SERVICE_UNAVAILABLE", "账号服务暂不可用")
		return
	}
	copyRequestHeader(request.Header, r.Header, "Authorization")
	if request.Header.Get("Authorization") == "" {
		if cookie, cookieErr := r.Cookie(accountCookieName); cookieErr == nil && strings.TrimSpace(cookie.Value) != "" {
			request.Header.Set("Authorization", "Bearer "+strings.TrimSpace(cookie.Value))
		}
	}
	copyRequestHeader(request.Header, r.Header, "Accept")
	copyRequestHeader(request.Header, r.Header, "Content-Type")
	copyRequestHeader(request.Header, r.Header, "Origin")
	request.Header.Set("X-Qiantie-Auth-Bridge", "1")

	response, err := p.client.Do(request)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "ACCOUNT_SERVICE_UNAVAILABLE", "无法连接一战晟铭账号服务，请确认 4001 已启动")
		return
	}
	defer response.Body.Close()

	copyResponseHeader(w.Header(), response.Header, "Content-Type")
	copyResponseHeader(w.Header(), response.Header, "Cache-Control")
	copyResponseHeader(w.Header(), response.Header, "Set-Cookie")
	copyResponseHeader(w.Header(), response.Header, "X-Qiantie-Auth-Failure")
	if r.URL.Path == "/api/login/logout" && response.StatusCode < http.StatusBadRequest {
		http.SetCookie(w, accountCookie("", -1, r.TLS != nil))
	}
	w.WriteHeader(response.StatusCode)
	_, _ = io.Copy(w, io.LimitReader(response.Body, 2<<20))
}

func NewBootstrap(baseURL string, client *http.Client) http.Handler {
	accountProxy, ok := New(baseURL, client).(*proxy)
	if !ok {
		return New(baseURL, client)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin, allowed := bootstrapOrigin(r)
		if !allowed {
			writeError(w, http.StatusForbidden, "AUTH_BRIDGE_ORIGIN_NOT_ALLOWED", "只允许从一战晟铭进入魔剪")
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Vary", "Origin")
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "请求方式不允许")
			return
		}
		token := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
		if token == "" || token == r.Header.Get("Authorization") {
			writeError(w, http.StatusUnauthorized, "AUTH_REQUIRED", "请先在一战晟铭登录")
			return
		}

		target := *accountProxy.baseURL
		target.Path = "/api/login/session"
		request, err := http.NewRequestWithContext(r.Context(), http.MethodGet, target.String(), nil)
		if err != nil {
			writeError(w, http.StatusServiceUnavailable, "ACCOUNT_SERVICE_UNAVAILABLE", "账号服务暂不可用")
			return
		}
		request.Header.Set("Authorization", "Bearer "+token)
		request.Header.Set("X-Qiantie-Auth-Bridge", "1")
		response, err := accountProxy.client.Do(request)
		if err != nil {
			writeError(w, http.StatusServiceUnavailable, "ACCOUNT_SERVICE_UNAVAILABLE", "无法连接一战晟铭账号服务，请确认 4001 已启动")
			return
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			copyResponseHeader(w.Header(), response.Header, "Content-Type")
			copyResponseHeader(w.Header(), response.Header, "X-Qiantie-Auth-Failure")
			w.WriteHeader(response.StatusCode)
			_, _ = io.Copy(w, io.LimitReader(response.Body, 2<<20))
			return
		}

		http.SetCookie(w, accountCookie(token, 30*24*60*60, r.TLS != nil))
		w.WriteHeader(http.StatusNoContent)
	})
}

func bootstrapOrigin(r *http.Request) (string, bool) {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	parsed, err := url.Parse(origin)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Port() != "4001" {
		return "", false
	}
	requestHost := r.Host
	if host, _, splitErr := net.SplitHostPort(r.Host); splitErr == nil {
		requestHost = host
	}
	return origin, strings.EqualFold(parsed.Hostname(), strings.Trim(requestHost, "[]"))
}

func accountCookie(value string, maxAge int, secure bool) *http.Cookie {
	return &http.Cookie{
		Name:     accountCookieName,
		Value:    value,
		Path:     "/api/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	}
}

func copyRequestHeader(destination, source http.Header, name string) {
	if value := source.Get(name); value != "" {
		destination.Set(name, value)
	}
}

func copyResponseHeader(destination, source http.Header, name string) {
	for _, value := range source.Values(name) {
		destination.Add(name, value)
	}
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"code": code, "error": message})
}
