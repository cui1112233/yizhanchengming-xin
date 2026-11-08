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

func TestRuntimeHandlerKeepsAllAPIRequestsOnGo(t *testing.T) {
	api := http.NewServeMux()
	api.HandleFunc("/api/known", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	static := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	handler := newRuntimeHandler(api, static)

	for _, path := range []string{"/api/known", "/api/unknown", "/health"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code == http.StatusTeapot {
			t.Fatalf("API path %s was sent to the static handler", path)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/index.html", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTeapot {
		t.Fatalf("page path reached API handler, status=%d", rec.Code)
	}
}

func TestRuntimeHandlerKeepsLocalExecutorDownloadsOnGo(t *testing.T) {
	api := http.NewServeMux()
	api.HandleFunc("GET /downloads/local-executor/manifest.json", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	static := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	handler := newRuntimeHandler(api, static)

	req := httptest.NewRequest(http.MethodGet, "/downloads/local-executor/manifest.json", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("local executor download reached status %d, want Go handler status %d", rec.Code, http.StatusNoContent)
	}
}

func TestRuntimeHandlerKeepsMemberAvatarsOnGo(t *testing.T) {
	api := http.NewServeMux()
	api.HandleFunc("GET /user-content/avatars/{file}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	static := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	handler := newRuntimeHandler(api, static)

	req := httptest.NewRequest(http.MethodGet, "/user-content/avatars/alice.png", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("avatar path reached status %d, want Go handler status %d", rec.Code, http.StatusNoContent)
	}
}

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

func TestConfigureDatabasePoolSupportsBurstAndCapsIdleConnections(t *testing.T) {
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
	const maxIdle = 4
	if stats.Idle != maxIdle {
		t.Fatalf("idle connections = %d, want %d after the burst", stats.Idle, maxIdle)
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
