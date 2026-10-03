package main

import "testing"

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
