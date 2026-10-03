package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestConfigFromEnvRequiresMySQLRedisAndOwner(t *testing.T) {
	t.Setenv("MYSQL_DSN", "user:pass@tcp(127.0.0.1:3306)/qiantie?parseTime=true")
	t.Setenv("REDIS_URL", "redis://127.0.0.1:6379/0")
	t.Setenv("SYSTEM_OWNER", "user-1")
	t.Setenv("LISTEN_ADDR", ":8080")

	cfg, err := configFromEnv()
	if err != nil { t.Fatal(err) }
	if cfg.MySQLDSN == "" || cfg.RedisURL == "" || cfg.SystemOwner != "user-1" || cfg.ListenAddr != ":8080" { t.Fatalf("cfg=%#v", cfg) }
}

func TestConfigFromEnvRejectsMissingRedis(t *testing.T) {
	t.Setenv("MYSQL_DSN", "dsn")
	t.Setenv("REDIS_URL", "")
	t.Setenv("SYSTEM_OWNER", "user-1")
	if _, err := configFromEnv(); err == nil { t.Fatal("expected missing redis error") }
}

func TestConfigFromEnvRejectsMissingOwner(t *testing.T) {
	t.Setenv("MYSQL_DSN", "dsn")
	t.Setenv("REDIS_URL", "redis://127.0.0.1:6379/0")
	t.Setenv("SYSTEM_OWNER", "")
	if _, err := configFromEnv(); err == nil { t.Fatal("expected missing owner error") }
}

func TestNewAgentServiceWiresStoreAndResponder(t *testing.T) {
	service := newAgentService(nil)
	if service == nil || service.Store == nil || service.Responder == nil {
		t.Fatalf("agent service not fully wired: %#v", service)
	}
}

func TestComposeHTTPHandlerKeepsAPIRoutesSeparateFromFrontend(t *testing.T) {
	api := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("api"))
	})
	frontend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("frontend"))
	})
	handler := composeHTTPHandler(api, frontend)

	cases := []struct {
		path string
		want string
	}{
		{path: "/api/batch-factory/intakes", want: "api"},
		{path: "/api/batch-factory/jobs", want: "api"},
		{path: "/api/agent/threads", want: "api"},
		{path: "/healthz", want: "api"},
		{path: "/batch-factory", want: "frontend"},
		{path: "/agent", want: "frontend"},
		{path: "/", want: "frontend"},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tc.path, nil))
			if recorder.Body.String() != tc.want { t.Fatalf("path=%s body=%q", tc.path, recorder.Body.String()) }
		})
	}
}
