package main

import (
	"os"
	"testing"
)

func setRequiredWorkerEnv(t *testing.T) {
	t.Helper()
	t.Setenv("MYSQL_DSN", "user:pass@tcp(mysql:3306)/app")
	t.Setenv("REDIS_URL", "redis://redis:6379/0")
	_ = os.Unsetenv("PIPELINE_QUEUE_KEY")
	_ = os.Unsetenv("121_FETCH_ENDPOINT")
	_ = os.Unsetenv("AI_BASE_URL")
	_ = os.Unsetenv("AI_MODEL")
	_ = os.Unsetenv("AI_API_KEY")
}

func TestConfigFromEnvRequiresMySQLAndRedisAndDefaults(t *testing.T) {
	setRequiredWorkerEnv(t)
	cfg, err := configFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.QueueKey != "qiantie:pipeline:ready" {
		t.Fatalf("queue=%q", cfg.QueueKey)
	}
	if cfg.FetchEndpoint != "https://txt.121w.com/api.php" {
		t.Fatalf("endpoint=%q", cfg.FetchEndpoint)
	}
	if cfg.AIBaseURL != "" || cfg.AIModel != "" || cfg.AIAPIKey != "" {
		t.Fatalf("AI should remain optional when completely unset: %+v", cfg)
	}
}

func TestConfigFromEnvAcceptsCompleteAIConfig(t *testing.T) {
	setRequiredWorkerEnv(t)
	t.Setenv("AI_BASE_URL", "https://example.ai/v1")
	t.Setenv("AI_MODEL", "text-model")
	t.Setenv("AI_API_KEY", "secret")

	cfg, err := configFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AIBaseURL != "https://example.ai/v1" || cfg.AIModel != "text-model" || cfg.AIAPIKey != "secret" {
		t.Fatalf("unexpected AI config: %+v", cfg)
	}
}

func TestConfigFromEnvRejectsPartialAIConfig(t *testing.T) {
	setRequiredWorkerEnv(t)
	t.Setenv("AI_BASE_URL", "https://example.ai/v1")
	if _, err := configFromEnv(); err == nil {
		t.Fatal("expected partial AI config error")
	}
}
