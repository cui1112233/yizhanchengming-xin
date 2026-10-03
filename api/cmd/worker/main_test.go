package main

import (
	"os"
	"testing"
)

func TestConfigFromEnvRequiresMySQLAndRedisAndDefaults(t *testing.T) {
	t.Setenv("MYSQL_DSN", "user:pass@tcp(mysql:3306)/app")
	t.Setenv("REDIS_URL", "redis://redis:6379/0")
	_ = os.Unsetenv("PIPELINE_QUEUE_KEY")
	_ = os.Unsetenv("121_FETCH_ENDPOINT")

	cfg, err := configFromEnv()
	if err != nil { t.Fatal(err) }
	if cfg.QueueKey != "qiantie:pipeline:ready" { t.Fatalf("queue=%q", cfg.QueueKey) }
	if cfg.FetchEndpoint != "https://txt.121w.com/api.php" { t.Fatalf("endpoint=%q", cfg.FetchEndpoint) }
}
