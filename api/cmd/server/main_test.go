package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/agent"
)

func setBaseServerEnv(t *testing.T) {
	t.Helper()
	t.Setenv("MYSQL_DSN", "user:pass@tcp(127.0.0.1:3306)/qiantie?parseTime=true")
	t.Setenv("REDIS_URL", "redis://127.0.0.1:6379/0")
	t.Setenv("SYSTEM_OWNER", "user-1")
	t.Setenv("LISTEN_ADDR", ":8080")
	t.Setenv("PIPELINE_QUEUE_KEY", "")
	t.Setenv("AGENT_TOOL_QUEUE_KEY", "")
	t.Setenv("AI_BASE_URL", "")
	t.Setenv("AI_MODEL", "")
	t.Setenv("AI_API_KEY", "")
	t.Setenv("TOS_ENDPOINT", "")
	t.Setenv("TOS_REGION", "")
	t.Setenv("TOS_ACCESS_KEY", "")
	t.Setenv("TOS_SECRET_KEY", "")
	t.Setenv("TOS_BUCKET", "")
}

func TestConfigFromEnvRequiresMySQLRedisAndOwner(t *testing.T) {
	setBaseServerEnv(t)
	cfg, err := configFromEnv()
	if err != nil { t.Fatal(err) }
	if cfg.MySQLDSN == "" || cfg.RedisURL == "" || cfg.SystemOwner != "user-1" || cfg.ListenAddr != ":8080" { t.Fatalf("cfg=%#v", cfg) }
	if cfg.AgentToolQueueKey != "qiantie:agent:tools" { t.Fatalf("agent queue=%q", cfg.AgentToolQueueKey) }
}

func TestConfigFromEnvAcceptsCompleteAgentAIConfig(t *testing.T) {
	setBaseServerEnv(t)
	t.Setenv("AI_BASE_URL", "https://api.example.com/v1")
	t.Setenv("AI_MODEL", "agent-model")
	t.Setenv("AI_API_KEY", "secret")
	cfg, err := configFromEnv()
	if err != nil { t.Fatal(err) }
	if cfg.AIBaseURL != "https://api.example.com/v1" || cfg.AIModel != "agent-model" || cfg.AIAPIKey != "secret" { t.Fatalf("AI cfg=%#v", cfg) }
}

func TestConfigFromEnvRejectsPartialAgentAIConfig(t *testing.T) {
	setBaseServerEnv(t)
	t.Setenv("AI_BASE_URL", "https://api.example.com/v1")
	if _, err := configFromEnv(); err == nil { t.Fatal("expected partial AI config error") }
}

func TestConfigFromEnvAcceptsCompleteTOSConfig(t *testing.T) {
	setBaseServerEnv(t)
	t.Setenv("TOS_ENDPOINT", "tos-cn-guangzhou.volces.com")
	t.Setenv("TOS_REGION", "cn-guangzhou")
	t.Setenv("TOS_ACCESS_KEY", "ak")
	t.Setenv("TOS_SECRET_KEY", "sk")
	t.Setenv("TOS_BUCKET", "prod-media")
	cfg, err := configFromEnv()
	if err != nil { t.Fatal(err) }
	if cfg.TOSBucket != "prod-media" || cfg.TOSRegion != "cn-guangzhou" { t.Fatalf("TOS cfg=%#v", cfg) }
}

func TestConfigFromEnvRejectsPartialTOSConfig(t *testing.T) {
	setBaseServerEnv(t)
	t.Setenv("TOS_BUCKET", "prod-media")
	if _, err := configFromEnv(); err == nil { t.Fatal("expected partial TOS config error") }
}

func TestConfigFromEnvRejectsMissingRedis(t *testing.T) {
	setBaseServerEnv(t)
	t.Setenv("REDIS_URL", "")
	if _, err := configFromEnv(); err == nil { t.Fatal("expected missing redis error") }
}

func TestConfigFromEnvRejectsMissingOwner(t *testing.T) {
	setBaseServerEnv(t)
	t.Setenv("SYSTEM_OWNER", "")
	if _, err := configFromEnv(); err == nil { t.Fatal("expected missing owner error") }
}

func TestNewAgentServiceUsesDeterministicResponderWithoutAI(t *testing.T) {
	service, err := newAgentService(nil, serverConfig{}, nil)
	if err != nil { t.Fatal(err) }
	if service == nil || service.Store == nil || service.Responder == nil { t.Fatalf("agent service not fully wired: %#v", service) }
	if _, ok := service.Responder.(*agent.Responder); !ok { t.Fatalf("expected deterministic responder, got %T", service.Responder) }
}

func TestNewAgentServiceUsesHybridResponderWithAI(t *testing.T) {
	service, err := newAgentService(nil, serverConfig{AIBaseURL:"https://api.example.com",AIModel:"agent-model",AIAPIKey:"secret"}, nil)
	if err != nil { t.Fatal(err) }
	if _, ok := service.Responder.(*agent.HybridResponder); !ok { t.Fatalf("expected hybrid responder, got %T", service.Responder) }
}

func TestComposeHTTPHandlerKeepsAPIRoutesSeparateFromFrontend(t *testing.T) {
	api := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("api")) })
	frontend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("frontend")) })
	handler := composeHTTPHandler(api, frontend)
	cases := []struct{ path, want string }{
		{"/api/batch-factory/intakes","api"},{"/api/batch-factory/jobs","api"},{"/api/agent/threads","api"},{"/api/media/assets/asset_1","api"},{"/healthz","api"},{"/batch-factory","frontend"},{"/agent","frontend"},{"/","frontend"},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tc.path, nil))
			if recorder.Body.String() != tc.want { t.Fatalf("path=%s body=%q",tc.path,recorder.Body.String()) }
		})
	}
}
