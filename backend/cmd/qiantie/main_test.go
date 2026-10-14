package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
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
	for _, path := range []string{"/api/config", "/api/agent/tasks", "/api/agent/tasks/demo", "/api/account-recovery/password/request", "/api/member/profile", "/api/novel-panel/projects", "/api/script/director-pipeline"} {
		if !shouldBridgeLegacyAPI(path) {
			t.Fatalf("%s should use the temporary Node bridge", path)
		}
	}
	if shouldBridgeLegacyAPI("/api/script-video") {
		t.Fatal("script video must use Go routes")
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
	if shouldBridgeLegacyAPI("/api/health") || shouldBridgeLegacyAPI("/api/chat") || shouldBridgeLegacyAPI("/api/script/smart-unified-style") || shouldBridgeLegacyAPI("/api/batch-factory/v11/capabilities") || shouldBridgeLegacyAPI("/api/batch-factory/v12/capabilities") || shouldBridgeLegacyAPI("/api/batch-factory/v12/batches") || shouldBridgeLegacyAPI("/api/batch-factory/v12/batches/summary") || shouldBridgeLegacyAPI("/api/batch-factory/v12/fetch-originals") || shouldBridgeLegacyAPI("/api/batch-factory/v12/image-generation") || shouldBridgeLegacyAPI("/api/batch-factory/v12/intakes/demo") || shouldBridgeLegacyAPI("/api/batch-factory/v12/batches/demo/books/book-1/h3/director") || shouldBridgeLegacyAPI("/api/batch-factory/v12/batches/demo/books/book-1/source") || shouldBridgeLegacyAPI("/api/batch-factory/v12/config-versions") || shouldBridgeLegacyAPI("/api/batch-factory/v12/drafts") || shouldBridgeLegacyAPI("/api/batch-factory/v12/automation-presets") || shouldBridgeLegacyAPI("/api/batch-factory/v12/automation-presets/demo") || shouldBridgeLegacyAPI("/api/batch-factory/v12/batches/demo/automation-presets") || shouldBridgeLegacyAPI("/api/batch-factory/v12/batches/demo/automation") || shouldBridgeLegacyAPI("/api/batch-factory/v12/batches/demo/status") || shouldBridgeLegacyAPI("/api/batch-factory/v12/batches/demo/books/book-1/stages") || shouldBridgeLegacyAPI("/api/batch-factory/v12/batches/demo/books/book-1/stages/director") || shouldBridgeLegacyAPI("/api/batch-factory/v12/batches/demo/books/book-1/assets") || shouldBridgeLegacyAPI("/api/batch-factory/v12/batches/demo/books/book-1/assets/asset-1/images/image-1/content") {
		t.Fatal("migrated Go API paths must not use the Node bridge")
	}
	if shouldBridgeLegacyAPI("/api/batch-factory/v12/batches/demo/merge-status") {
		t.Fatal("read-only V12 merge status should use the Go store")
	}
	if shouldBridgeLegacyAPI("/api/batch-factory/v12/batches/demo/runtime-summary") {
		t.Fatal("read-only V12 runtime summary should use Go")
	}
	if shouldBridgeLegacyAPI("/api/batch-factory/v12/video-provider/status") {
		t.Fatal("read-only V12 video provider status should use Go")
	}
	if shouldBridgeLegacyAPI("/api/batch-factory/v12/batches/demo/books/book-1/videos/video-1/effective-settings") || shouldBridgeLegacyAPI("/api/batch-factory/v12/batches/demo/books/book-1/videos/video-1/final-prompt") {
		t.Fatal("read-only V12 compiler previews must use Go")
	}
	if !shouldBridgeLegacyAPIForSlice("/api/batch-factory/v12/batches/demo/books/book-1/h3/director", 1) {
		t.Fatal("H3 execution must stay on the Node bridge before slice 3 is enabled")
	}
	if shouldBridgeLegacyAPIForSlice("/api/batch-factory/v12/batches/demo/books/book-1/h3/director", 3) {
		t.Fatal("H3 execution should use Go once slice 3 is enabled")
	}
	for _, path := range []string{
		"/api/batch-factory/v12/batches/demo/director",
		"/api/batch-factory/v12/batches/demo/books/book-1/director",
		"/api/batch-factory/v12/batches/demo/books/book-1/opening-variants",
		"/api/batch-factory/v12/batches/demo/books/book-1/working-front/viral",
		"/api/batch-factory/v12/batches/demo/books/book-1/hooks/hook-1/approve",
		"/api/batch-factory/v12/batches/demo/books/book-1/production",
		"/api/batch-factory/v12/batches/demo/production",
		"/api/batch-factory/v12/batches/demo/production/cancel",
		"/api/batch-factory/v12/batches/demo/merge",
		"/api/batch-factory/v12/batches/demo/books/book-1/merge",
		"/api/batch-factory/v12/batches/demo/merge-media/artifact-1",
		"/api/batch-factory/v12/publish/121/credential",
		"/api/batch-factory/v12/publish/121/intents",
		"/api/batch-factory/v12/publish/121/intents/intent-1/confirm",
	} {
		if shouldBridgeLegacyAPIForSlice(path, 3) {
			t.Fatalf("slice 3 director path must stay on Go: %s", path)
		}
	}
}

func TestAutomationStartIsNativeGo(t *testing.T) {
	path := "/api/batch-factory/v12/batches/demo/automation/start"
	if shouldBridgeLegacyAPI(path) {
		t.Fatalf("automation start must be handled by the Go coordinator")
	}
}

func TestBatchFactoryV12IntakeReadIsNativeGo(t *testing.T) {
	if shouldBridgeLegacyAPI("/api/batch-factory/v12/intakes/intake-1") {
		t.Fatal("V12 intake reads must use the Go batch store")
	}
}

func TestBatchFactoryV12AssetImageGenerationIsNativeGo(t *testing.T) {
	if shouldBridgeLegacyAPI("/api/batch-factory/v12/batches/batch-1/books/book-1/assets/images/generate") {
		t.Fatal("V12 asset image generation must use the Go provider boundary")
	}
}

func TestShuihuoProductionBatchFactoryPathsStayOnGo(t *testing.T) {
	paths := []string{
		"/api/batch-factory/v12/batches/demo",
		"/api/batch-factory/v12/batches/demo/books/book-1/assets",
		"/api/batch-factory/v12/batches/demo/books/book-1/director",
		"/api/batch-factory/v12/batches/demo/books/book-1/production",
		"/api/batch-factory/v12/batches/demo/runtime-summary",
		"/api/batch-factory/v12/publish-121/organizations",
	}
	for _, path := range paths {
		if shouldBridgeLegacyAPIForSlice(path, 3) {
			t.Fatalf("水火生产路径不应转到旧服务: %s", path)
		}
	}
}

func TestShouldBridgeLegacyAPIRequestKeepsH3TraceReadOnGo(t *testing.T) {
	read := httptest.NewRequest(http.MethodGet, "/api/batch-factory/v12/batches/demo/books/book-1/h3/trace", nil)
	if shouldBridgeLegacyAPIRequest(read, 1) {
		t.Fatal("H3 trace read should use Go while H3 writes remain bridged")
	}
	write := httptest.NewRequest(http.MethodPost, "/api/batch-factory/v12/batches/demo/books/book-1/h3/trace", nil)
	if !shouldBridgeLegacyAPIRequest(write, 1) {
		t.Fatal("non-GET H3 trace path should keep the legacy bridge")
	}
}

func TestShouldBridgeLegacyAPIRequestKeepsExternalClassificationAdaptersOnBridge(t *testing.T) {
	for _, path := range []string{
		"/api/batch-factory/v12/batches/demo/classify-fetched-metadata",
		"/api/batch-factory/v12/batches/demo/backfill-121-metadata",
	} {
		req := httptest.NewRequest(http.MethodPost, path, nil)
		if shouldBridgeLegacyAPIRequest(req, 1) {
			t.Fatalf("external adapter %s must use the Go metadata handler", path)
		}
	}
}

func TestNovelFetchCompatibilitySurfaceUsesGoHandler(t *testing.T) {
	for _, tc := range []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/batch-rewrite/knowledge"},
		{http.MethodGet, "/api/batch-rewrite/knowledge/summary"},
		{http.MethodPost, "/api/batch-rewrite/knowledge"},
		{http.MethodDelete, "/api/batch-rewrite/knowledge/item-1"},
	} {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		if !isGoNovelFetchCompatRequest(req) || isLegacyNovelFetchBridgeRequest(req) {
			t.Fatalf("%s %s must use the Go knowledge handler", tc.method, tc.path)
		}
	}
	for _, path := range []string{
		"/api/batch-rewrite/process/jobs/latest",
		"/api/batch-rewrite/rules/ai-suggest",
	} {
		method := http.MethodGet
		if strings.Contains(path, "rules/") {
			method = http.MethodPost
		}
		req := httptest.NewRequest(method, path, nil)
		if !isGoNovelFetchCompatRequest(req) || isLegacyNovelFetchBridgeRequest(req) {
			t.Fatalf("%s %s must use the Go native compatibility handler", method, path)
		}
	}
	for _, tc := range []struct{ method, path string }{
		{http.MethodPost, "/api/batch-rewrite/opening/analyze"},
		{http.MethodPost, "/api/batch-rewrite/rules/preview"},
		{http.MethodPost, "/api/batch-rewrite/ai/test"},
		{http.MethodPost, "/api/batch-rewrite/process"},
		{http.MethodPost, "/api/batch-rewrite/process/start"},
		{http.MethodPost, "/api/batch-rewrite/tasks/book-1/fetch"},
		{http.MethodPost, "/api/batch-rewrite/tasks/book-1/generate-ai"},
		{http.MethodPost, "/api/batch-rewrite/tasks/batch-retry"},
		{http.MethodPost, "/api/batch-rewrite/config"},
		{http.MethodPost, "/api/batch-rewrite/knowledge/optimize"},
		{http.MethodGet, "/api/batch-rewrite/logs"},
		{http.MethodGet, "/api/batch-rewrite/schedules"},
		{http.MethodPost, "/api/batch-rewrite/schedules"},
		{http.MethodPatch, "/api/batch-rewrite/schedules/schedule-1"},
		{http.MethodDelete, "/api/batch-rewrite/schedules/schedule-1"},
		{http.MethodGet, "/api/batch-rewrite/batches/current"},
		{http.MethodPost, "/api/batch-rewrite/tasks/apply-rules"},
		{http.MethodPost, "/api/batch-rewrite/tasks/reprocess-sensitive"},
		{http.MethodPost, "/api/batch-rewrite/tasks/batch-ai-count"},
		{http.MethodPost, "/api/batch-rewrite/web-submit/operations"},
		{http.MethodGet, "/api/batch-rewrite/web-submit/operations/op-1"},
	} {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		if !isGoNovelFetchCompatRequest(req) || isLegacyNovelFetchBridgeRequest(req) {
			t.Fatalf("%s %s must use the Go native compatibility handler", tc.method, tc.path)
		}
	}
	for _, tc := range []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/batch-rewrite/issues"},
		{http.MethodGet, "/api/batch-rewrite/web-submit/history"},
		{http.MethodPost, "/api/batch-rewrite/web-submit/submit"},
		{http.MethodPost, "/api/batch-rewrite/tasks/batch-delete"},
		{http.MethodPost, "/api/batch-rewrite/work-form"},
	} {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		if !isGoNovelFetchCompatRequest(req) || isLegacyNovelFetchBridgeRequest(req) {
			t.Fatalf("%s %s must use the Go compatibility handler", tc.method, tc.path)
		}
	}
}

func TestNovelFetchNamespaceNeverFallsBackToNode(t *testing.T) {
	for _, path := range []string{
		"/api/batch-rewrite/future-endpoint",
		"/api/novel-fetch-workshop/future-endpoint",
		"/api/novel-fetch-upload/future-endpoint",
		"/api/novel-fetch/future-endpoint",
	} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		if !isGoNovelFetchCompatRequest(req) || isLegacyNovelFetchBridgeRequest(req) {
			t.Fatalf("%s must stay on Go and never use the Node bridge", path)
		}
	}
}

func TestNovelFetchSavedSurfaceUsesGoHandler(t *testing.T) {
	for _, tc := range []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/api/novel-fetch/saved"},
		{method: http.MethodPost, path: "/api/novel-fetch/save"},
		{method: http.MethodPost, path: "/api/novel-fetch"},
		{method: http.MethodGet, path: "/api/batch-rewrite/tasks"},
		{method: http.MethodGet, path: "/api/batch-rewrite/tasks/book-1"},
		{method: http.MethodGet, path: "/api/batch-rewrite/config"},
		{method: http.MethodGet, path: "/api/batch-rewrite/bootstrap"},
	} {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		if !isGoNovelFetchCompatRequest(req) {
			t.Fatalf("%s %s must use the Go saved-novel handler", tc.method, tc.path)
		}
	}
}

func TestNovelFetchNestedTaskDiagnosticsUseGoHandler(t *testing.T) {
	for _, path := range []string{
		"/api/batch-rewrite/tasks/book-1/sensitive-log",
		"/api/batch-rewrite/tasks/book-1/rules-trace",
		"/api/batch-rewrite/tasks/book-1/site-submit-log",
	} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		if !isGoNovelFetchCompatRequest(req) {
			t.Fatalf("nested task endpoint %s must use the Go compatibility handler", path)
		}
	}
}

func TestGoAuthEnabledDefaultsSafelyWithoutNodeAuthTarget(t *testing.T) {
	for _, tc := range []struct {
		name     string
		nodeURL  string
		switchOn string
		want     bool
	}{
		{name: "no Node URL and no switch", want: true},
		{name: "Node URL and no switch", nodeURL: "http://127.0.0.1:3001", want: false},
		{name: "explicit true", switchOn: "true", want: true},
		{name: "explicit one", switchOn: "1", want: true},
		{name: "no Node URL and explicit false", switchOn: "false", want: true},
		{name: "Node URL and explicit false", nodeURL: "http://127.0.0.1:3001", switchOn: "false", want: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("QIANTIE_GO_AUTH_ENABLED", tc.switchOn)
			if got := goAuthEnabled(tc.nodeURL); got != tc.want {
				t.Fatalf("goAuthEnabled(%q) = %t, want %t", tc.nodeURL, got, tc.want)
			}
		})
	}
}

func TestAdminPresetReadsUseGoHandler(t *testing.T) {
	for _, path := range []string{"/api/admin/presets", "/api/admin/preset-slots"} {
		if !isGoAdminReadRequest(httptest.NewRequest(http.MethodGet, path, nil)) {
			t.Fatalf("%s must use the Go admin read handler", path)
		}
	}
	for _, item := range []struct{ method, path string }{
		{http.MethodGet, "/api/admin/presets/demo/1"},
		{http.MethodPost, "/api/admin/presets/draft"},
		{http.MethodPut, "/api/admin/presets/demo/1/draft"},
		{http.MethodPost, "/api/admin/presets/demo/publish"},
		{http.MethodPost, "/api/admin/presets/demo/rollback"},
	} {
		if !isGoAdminReadRequest(httptest.NewRequest(item.method, item.path, nil)) {
			t.Fatalf("%s %s must use the Go admin preset handler", item.method, item.path)
		}
	}
	if isGoAdminReadRequest(httptest.NewRequest(http.MethodPost, "/api/admin/presets", nil)) {
		t.Fatal("unrecognized admin preset route must not use the Go handler")
	}
}
