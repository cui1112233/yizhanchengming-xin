package safehttp

import (
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestIsBlockedIP(t *testing.T) {
	blocked := []string{
		"127.0.0.1",
		"127.1.2.3",
		"10.0.0.1",
		"172.16.0.1",
		"172.31.255.255",
		"192.168.1.1",
		"169.254.169.254",
		"100.100.100.200",
		"100.64.0.1",
		"0.0.0.0",
		"198.18.0.1",
		"255.255.255.255",
		"::1",
		"fe80::1",
		"fc00::1",
		"::",
	}
	for _, raw := range blocked {
		if !IsBlockedIP(net.ParseIP(raw)) {
			t.Fatalf("expected %s to be blocked", raw)
		}
	}
	allowed := []string{"8.8.8.8", "1.1.1.1", "104.16.0.1", "2001:4860:4860::8888"}
	for _, raw := range allowed {
		if IsBlockedIP(net.ParseIP(raw)) {
			t.Fatalf("expected %s to be allowed", raw)
		}
	}
	if !IsBlockedIP(nil) {
		t.Fatal("nil ip must be blocked")
	}
}

func TestCheckURLRejectsPrivateAndInvalidScheme(t *testing.T) {
	cases := []struct {
		raw   string
		match error
	}{
		{"http://127.0.0.1:4001/api", ErrBlockedAddress},
		{"http://169.254.169.254/latest/meta-data", ErrBlockedAddress},
		{"http://100.100.100.200/latest/meta-data", ErrBlockedAddress},
		{"http://[::1]/health", ErrBlockedAddress},
		{"http://localhost:8080/", ErrBlockedAddress},
		{"file:///etc/passwd", ErrBlockedScheme},
		{"ftp://example.com/file", ErrBlockedScheme},
	}
	for _, item := range cases {
		err := checkURL(item.raw, false)
		if !errors.Is(err, item.match) {
			t.Fatalf("checkURL(%q) = %v, want %v", item.raw, err, item.match)
		}
	}
}

func TestCheckURLAllowsPublicLiteral(t *testing.T) {
	if err := checkURL("https://8.8.8.8/v1/chat", false); err != nil {
		t.Fatalf("public literal rejected: %v", err)
	}
}

func TestRedirectPolicyLimitsHopsAndRechecks(t *testing.T) {
	previous := allowPrivateDestinations.Swap(false)
	defer allowPrivateDestinations.Store(previous)
	policy := redirectPolicy(2)
	via := []*http.Request{{}, {}}
	if err := policy(&http.Request{URL: mustURL(t, "https://8.8.8.8/")}, via); !errors.Is(err, ErrTooManyRedirects) {
		t.Fatalf("expected too many redirects, got %v", err)
	}
	if err := policy(&http.Request{URL: mustURL(t, "http://127.0.0.1/")}, nil); !errors.Is(err, ErrBlockedAddress) {
		t.Fatalf("expected blocked address, got %v", err)
	}
	if err := policy(&http.Request{URL: mustURL(t, "https://8.8.8.8/")}, nil); err != nil {
		t.Fatalf("public hop rejected: %v", err)
	}
}

func TestNewClientPreservesTimeoutAndTransportSettings(t *testing.T) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.ResponseHeaderTimeout = 3 * time.Minute
	client := NewClientWithTransport(6*time.Minute, transport)
	if client.Timeout != 6*time.Minute {
		t.Fatalf("timeout = %s, want 6m", client.Timeout)
	}
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("transport type = %T, want *http.Transport", client.Transport)
	}
	if transport.ResponseHeaderTimeout != 3*time.Minute {
		t.Fatalf("response header timeout = %s, want 3m", transport.ResponseHeaderTimeout)
	}
	if transport.DialContext == nil {
		t.Fatal("dial context must be replaced with guarded dialer")
	}
}

func TestNewClientDefaults(t *testing.T) {
	client := NewClient(Options{})
	if client.Timeout != DefaultTimeout {
		t.Fatalf("timeout = %s, want %s", client.Timeout, DefaultTimeout)
	}
	if client.CheckRedirect == nil {
		t.Fatal("redirect policy must be set")
	}
}

func TestExplicitAllowMatchesOnlyConfiguredAddress(t *testing.T) {
	previous := allowPrivateDestinations.Swap(false)
	defer allowPrivateDestinations.Store(previous)
	t.Setenv(envDoubaoRelayBaseURL, "http://127.0.0.1:8787")

	if err := checkURL("http://127.0.0.1:8787/v1/images/generations", false); err != nil {
		t.Fatalf("configured address rejected: %v", err)
	}
	cases := []string{
		"http://127.0.0.1:8788/v1/images/generations",  // 同机其它端口
		"https://127.0.0.1:8787/v1/images/generations", // 协议不一致
		"http://127.0.0.2:8787/v1/images/generations",  // 同网段其它地址
	}
	for _, raw := range cases {
		if err := checkURL(raw, false); !errors.Is(err, ErrBlockedAddress) {
			t.Fatalf("checkURL(%q) = %v, want blocked", raw, err)
		}
	}
}

func TestExplicitAllowUnsetBlocksLoopback(t *testing.T) {
	previous := allowPrivateDestinations.Swap(false)
	defer allowPrivateDestinations.Store(previous)
	t.Setenv(envDoubaoRelayBaseURL, "")
	t.Setenv(envOutboundAllowURLs, "")

	if err := checkURL("http://127.0.0.1:8787/v1/images/generations", false); !errors.Is(err, ErrBlockedAddress) {
		t.Fatalf("unset config must keep loopback blocked, got %v", err)
	}
}

func TestExplicitAllowListEnvMatchesOnlyListedAddresses(t *testing.T) {
	previous := allowPrivateDestinations.Swap(false)
	defer allowPrivateDestinations.Store(previous)
	t.Setenv(envDoubaoRelayBaseURL, "")
	t.Setenv(envOutboundAllowURLs, "http://127.0.0.1:18900, https://127.0.0.1:18901")

	if err := checkURL("http://127.0.0.1:18900/v1/chat/completions", false); err != nil {
		t.Fatalf("listed address rejected: %v", err)
	}
	if err := checkURL("https://127.0.0.1:18901/v1/chat/completions", false); err != nil {
		t.Fatalf("second listed address rejected: %v", err)
	}
	cases := []string{
		"http://127.0.0.1:18902/v1/chat/completions",  // 未列出的端口
		"https://127.0.0.1:18900/v1/chat/completions", // 协议不一致
		"http://127.0.0.2:18900/v1/chat/completions",  // 未列出的地址
	}
	for _, raw := range cases {
		if err := checkURL(raw, false); !errors.Is(err, ErrBlockedAddress) {
			t.Fatalf("checkURL(%q) = %v, want blocked", raw, err)
		}
	}
}

func TestExplicitAllowDomainChecksResolvedAddresses(t *testing.T) {
	previous := allowPrivateDestinations.Swap(false)
	defer allowPrivateDestinations.Store(previous)
	t.Setenv(envDoubaoRelayBaseURL, "http://localhost:8787")

	if err := checkURL("http://localhost:8787/v1/images/generations", false); err != nil {
		t.Fatalf("configured domain rejected: %v", err)
	}
	// 域名放行不等于同机字面量地址放行：host 必须完全一致。
	if err := checkURL("http://127.0.0.1:8787/v1/images/generations", false); !errors.Is(err, ErrBlockedAddress) {
		t.Fatalf("literal loopback must not inherit domain allow, got %v", err)
	}
}

func TestExplicitAllowAppliesToDialAndRedirect(t *testing.T) {
	previous := allowPrivateDestinations.Swap(false)
	defer allowPrivateDestinations.Store(previous)
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})
	allowed := httptest.NewServer(handler)
	defer allowed.Close()
	other := httptest.NewServer(handler)
	defer other.Close()
	t.Setenv(envDoubaoRelayBaseURL, allowed.URL)

	// 请求级与拨号级都应只放行配置地址。
	client := NewClient(Options{Timeout: 5 * time.Second})
	response, err := client.Get(allowed.URL)
	if err != nil {
		t.Fatalf("configured address request failed: %v", err)
	}
	response.Body.Close()
	if _, err := client.Get(other.URL); !errors.Is(err, ErrBlockedAddress) {
		t.Fatalf("same host other port = %v, want blocked", err)
	}

	// 专用 Transport 不套请求级校验，只能靠拨号白名单；仍须拒绝其它端口。
	dialOnly := NewClientWithTransport(5*time.Second, http.DefaultTransport.(*http.Transport).Clone())
	response, err = dialOnly.Get(allowed.URL)
	if err != nil {
		t.Fatalf("dial allow list rejected configured address: %v", err)
	}
	response.Body.Close()
	if _, err := dialOnly.Get(other.URL); err == nil {
		t.Fatal("dial allow list must reject other loopback port")
	}

	// 重定向跳到白名单外的端口必须被逐跳校验拒绝。
	redirector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, other.URL, http.StatusFound)
	}))
	defer redirector.Close()
	t.Setenv(envDoubaoRelayBaseURL, redirector.URL)
	if _, err := NewClient(Options{Timeout: 5 * time.Second}).Get(redirector.URL); !errors.Is(err, ErrBlockedAddress) {
		t.Fatalf("redirect outside allow list = %v, want blocked", err)
	}
}

func mustURL(t *testing.T, raw string) *url.URL {
	t.Helper()
	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parse %q: %v", raw, err)
	}
	return parsed
}
