package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
)

func TestHealthURLNormalizesWildcardListenAddresses(t *testing.T) {
	for _, tc := range []struct {
		listen string
		want   string
	}{
		{listen: ":4000", want: "http://127.0.0.1:4000/health"},
		{listen: "0.0.0.0:4000", want: "http://127.0.0.1:4000/health"},
		{listen: "[::]:4000", want: "http://127.0.0.1:4000/health"},
		{listen: "127.0.0.1:4555", want: "http://127.0.0.1:4555/health"},
	} {
		if got := healthURL(tc.listen); got != tc.want {
			t.Fatalf("healthURL(%q) = %q, want %q", tc.listen, got, tc.want)
		}
	}
}

func TestRunHealthcheckRequiresHTTP200(t *testing.T) {
	okServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			t.Fatalf("healthcheck path = %q, want /health", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer okServer.Close()
	if err := runHealthcheck(okServer.URL + "/health"); err != nil {
		t.Fatalf("runHealthcheck returned error for healthy server: %v", err)
	}

	badServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer badServer.Close()
	if err := runHealthcheck(badServer.URL + "/health"); err == nil {
		t.Fatal("runHealthcheck accepted non-200 response")
	}
}

func TestConfigureDatabasePoolReusesBurstConnections(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.MatchExpectationsInOrder(false)
	configureDatabasePool(db)

	const workers = 8
	for i := 0; i < workers; i++ {
		mock.ExpectPing().WillDelayFor(50 * time.Millisecond)
	}
	start := make(chan struct{})
	errs := make(chan error, workers)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			errs <- db.PingContext(context.Background())
		}()
	}
	close(start)
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	stats := db.Stats()
	if stats.MaxOpenConnections != 24 {
		t.Fatalf("max open connections = %d, want 24", stats.MaxOpenConnections)
	}
	if stats.Idle < workers {
		t.Fatalf("idle connections = %d, want at least %d reusable burst connections", stats.Idle, workers)
	}
}

func TestDatabaseDSNEnablesInterpolationWithoutDroppingExistingOptions(t *testing.T) {
	got, err := databaseDSN("user:pass@tcp(mysql:3306)/qiantie_v88?parseTime=true")
	if err != nil {
		t.Fatalf("databaseDSN() error = %v", err)
	}
	cfg, err := mysql.ParseDSN(got)
	if err != nil {
		t.Fatalf("ParseDSN(%q) error = %v", got, err)
	}
	if !cfg.InterpolateParams {
		t.Fatal("InterpolateParams must be enabled to avoid per-query server prepares")
	}
	if !cfg.ParseTime {
		t.Fatal("databaseDSN() dropped parseTime")
	}
}

func TestAuthProxyForwardsLoginRequestsToNode(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/login" || r.Method != http.MethodPost {
			t.Fatalf("upstream request = %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"用户名或密码错误"}`))
	}))
	defer upstream.Close()

	proxy, err := newAuthProxy(upstream.URL)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/login", nil)
	res := httptest.NewRecorder()
	proxy.ServeHTTP(res, req)
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("proxy status = %d, want %d", res.Code, http.StatusUnauthorized)
	}
}

func TestShouldBridgeLegacyAPI(t *testing.T) {
	for _, path := range []string{"/api/config", "/api/agent/tasks", "/api/agent/tasks/demo", "/api/account-recovery/password/request", "/api/member/profile", "/api/novel-panel/projects", "/api/script-video"} {
		if !shouldBridgeLegacyAPI(path) {
			t.Fatalf("%s should use the temporary Node bridge", path)
		}
	}
	if shouldBridgeLegacyAPI("/api/tts") {
		t.Fatal("/api/tts is implemented by Go and must not use the Node bridge")
	}
	if shouldBridgeLegacyAPI("/api/presets") {
		t.Fatal("/api/presets is implemented by Go and must not use the Node bridge")
	}
	if shouldBridgeLegacyAPI("/api/script-constraint-prompts") {
		t.Fatal("/api/script-constraint-prompts is implemented by Go and must not use the Node bridge")
	}
	for _, path := range []string{"/api/models", "/api/models?kind=text", "/api/config/models", "/api/config/models/text-demo"} {
		if shouldBridgeLegacyAPI(path) {
			t.Fatalf("%s should be handled by Go model catalog", path)
		}
	}
	if shouldBridgeLegacyAPI("/api/health") || shouldBridgeLegacyAPI("/api/chat") || shouldBridgeLegacyAPI("/api/script/smart-unified-style") || shouldBridgeLegacyAPI("/api/batch-factory/v11/capabilities") || shouldBridgeLegacyAPI("/api/batch-factory/v12/capabilities") || shouldBridgeLegacyAPI("/api/batch-factory/v12/batches") || shouldBridgeLegacyAPI("/api/batch-factory/v12/batches/summary") || shouldBridgeLegacyAPI("/api/batch-factory/v12/fetch-originals") || shouldBridgeLegacyAPI("/api/batch-factory/v12/intakes/demo") || shouldBridgeLegacyAPI("/api/batch-factory/v12/batches/demo/books/book-1/h3/director") || shouldBridgeLegacyAPI("/api/batch-factory/v12/batches/demo/books/book-1/source") || shouldBridgeLegacyAPI("/api/batch-factory/v12/config-versions") || shouldBridgeLegacyAPI("/api/batch-factory/v12/drafts") {
		t.Fatal("migrated Go API paths must not use the Node bridge")
	}
	if !shouldBridgeLegacyAPI("/api/batch-factory/v12/batches/demo/status") || !shouldBridgeLegacyAPI("/api/batch-factory/v12/batches/demo/merge-status") || !shouldBridgeLegacyAPI("/api/batch-factory/v12/batches/demo/books/book-1/stages") {
		t.Fatal("slice-gated V12 routes must remain on the compatibility bridge when not enabled locally")
	}
	if !shouldBridgeLegacyAPI("/api/batch-factory/v12/batches/demo/runtime-summary") {
		t.Fatal("remaining V12 compatibility adapters must use the local Node bridge")
	}
}

func TestGoAuthEnabledOnlyWithExplicitOptIn(t *testing.T) {
	t.Setenv("QIANTIE_GO_AUTH_ENABLED", "")
	if goAuthEnabled() {
		t.Fatal("Go auth must remain opt-in before legacy account import")
	}
	t.Setenv("QIANTIE_GO_AUTH_ENABLED", "1")
	if !goAuthEnabled() {
		t.Fatal("Go auth should enable with explicit opt-in")
	}
}
