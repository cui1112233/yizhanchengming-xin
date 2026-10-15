package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"qiantie/giant-material-executor/internal/agent"
)

func TestLoopbackServerRejectsNonLoopbackBind(t *testing.T) {
	if _, err := NewServer(ServerConfig{Addr: "0.0.0.0:17861", Origin: "https://example.com", Nonce: "nonce"}); err == nil {
		t.Fatal("expected non-loopback bind rejection")
	}
}

func TestLoopbackServerRejectsUnknownOrigin(t *testing.T) {
	server, err := NewServer(ServerConfig{Addr: "127.0.0.1:17861", Origin: "https://example.com", Nonce: "nonce"})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/v1/health", nil)
	req.Header.Set("Origin", "https://evil.example")
	req.Header.Set("X-Giant-Executor-Nonce", "nonce")
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestHealthShowsMissingModelWithoutLeakingCredentials(t *testing.T) {
	server, err := NewServer(ServerConfig{Addr: "127.0.0.1:17861", Origin: "https://example.com", Nonce: "nonce", Version: "0.1.0", Snapshot: func() agent.Snapshot {
		return agent.Snapshot{State: agent.StateDownloadingModel, BindingState: agent.BindingConnecting, WorkerResident: true}
	}})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/v1/health", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("X-Giant-Executor-Nonce", "nonce")
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"modelReady":false`) || !strings.Contains(body, `"state":"downloading_model"`) || !strings.Contains(body, `"bindingState":"connecting"`) {
		t.Fatalf("body=%s", body)
	}
	if strings.Contains(strings.ToLower(body), "token") || strings.Contains(strings.ToLower(body), "cookie") {
		t.Fatalf("credentials leaked: %s", body)
	}
}

func TestPairingCanBootstrapNonceAndReturnsItToBrowser(t *testing.T) {
	server, err := NewServer(ServerConfig{Addr: "127.0.0.1:17861", Origin: "https://example.com", Nonce: "nonce", Callbacks: Callbacks{
		Pair: func(context.Context, string) error { return nil },
	}})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/v1/pair", strings.NewReader(`{"code":"one-time-code"}`))
	req.Header.Set("Origin", "https://example.com")
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"nonce":"nonce"`) {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestSetupPageIsServedFromLoopback(t *testing.T) {
	server, err := NewServer(ServerConfig{Addr: "127.0.0.1:17861", Origin: "https://example.com", Nonce: "nonce"})
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/setup", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	for _, want := range []string{"巨量素材执行器", `id="publicURL"`, `id="code"`, "fetch('/v1/pair'", `id="applyUpdate"`, "/v1/update/apply"} {
		if !strings.Contains(rec.Body.String(), want) {
			t.Fatalf("setup page missing %q", want)
		}
	}
}

func TestSetupPageRequiresExplicitLocalUpdateAction(t *testing.T) {
	server, err := NewServer(ServerConfig{Addr: "127.0.0.1:17861", Origin: "https://example.com", Nonce: "nonce"})
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/setup", nil))
	body := rec.Body.String()
	if strings.Contains(body, `id="applyUpdate" hidden`) {
		t.Fatal("setup page hides the explicit local update action")
	}
	if strings.Contains(body, "正在下载替换，执行器会自动重启") {
		t.Fatal("setup page still claims that update checks replace the app automatically")
	}
	if !strings.Contains(body, "请点击“立即更新”") {
		t.Fatal("setup page does not explain the required local update confirmation")
	}
}

func TestLocalSetupOriginCanPair(t *testing.T) {
	var pairedCode string
	server, err := NewServer(ServerConfig{Addr: "127.0.0.1:17861", Origin: "https://example.com", Nonce: "nonce", Callbacks: Callbacks{
		Pair: func(_ context.Context, code string) error {
			pairedCode = code
			return nil
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/v1/pair", strings.NewReader(`{"code":"one-time-code"}`))
	req.Header.Set("Origin", "http://127.0.0.1:17861")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || pairedCode != "one-time-code" {
		t.Fatalf("status=%d paired=%q body=%s", rec.Code, pairedCode, rec.Body.String())
	}
}

func TestLocalSetupPairAllowsSameLoopbackRequestWithoutOriginHeader(t *testing.T) {
	var pairedCode string
	server, err := NewServer(ServerConfig{Addr: "127.0.0.1:17861", Origin: "https://example.com", Nonce: "nonce", Callbacks: Callbacks{
		Pair: func(_ context.Context, code string) error {
			pairedCode = code
			return nil
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:17861/v1/pair", strings.NewReader(`{"code":"same-loopback-code"}`))
	req.Host = "127.0.0.1:17861"
	req.RemoteAddr = "127.0.0.1:54321"
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || pairedCode != "same-loopback-code" {
		t.Fatalf("status=%d paired=%q body=%s", rec.Code, pairedCode, rec.Body.String())
	}
}

func TestLocalSetupCanApplyUpdateOnlyThroughLoopback(t *testing.T) {
	applied := false
	server, err := NewServer(ServerConfig{Addr: "127.0.0.1:17861", Origin: "https://example.com", Nonce: "nonce", Callbacks: Callbacks{
		ApplyUpdate: func(context.Context) (string, error) { applied = true; return "0.6.0", nil },
	}})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:17861/v1/update/apply", nil)
	req.Host = "127.0.0.1:17861"
	req.RemoteAddr = "127.0.0.1:54321"
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !applied || !strings.Contains(rec.Body.String(), `"version":"0.6.0"`) {
		t.Fatalf("status=%d applied=%v body=%s", rec.Code, applied, rec.Body.String())
	}
}

func TestPairingPersistsBrowserSelectedPublicURLBeforePair(t *testing.T) {
	var savedURL string
	var pairedURL string
	server, err := NewServer(ServerConfig{Addr: "127.0.0.1:17861", Origin: "https://example.com", Nonce: "nonce", Callbacks: Callbacks{
		SetPublicURL: func(value string) error {
			savedURL = value
			return nil
		},
		Pair: func(_ context.Context, _ string) error {
			pairedURL = savedURL
			return nil
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/v1/pair", strings.NewReader(`{"publicURL":"https://factory.example.com","code":"one-time-code"}`))
	req.Header.Set("Origin", "http://127.0.0.1:17861")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || savedURL != "https://factory.example.com" || pairedURL != savedURL {
		t.Fatalf("status=%d saved=%q paired=%q body=%s", rec.Code, savedURL, pairedURL, rec.Body.String())
	}
}
