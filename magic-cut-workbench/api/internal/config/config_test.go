package config

import "testing"

func TestLoadDefaultsToIsolatedDevelopmentRuntime(t *testing.T) {
	t.Setenv("MAGIC_CUT_LISTEN_ADDR", "")
	t.Setenv("MAGIC_CUT_AUTH_BASE_URL", "")
	t.Setenv("MAGIC_CUT_DATABASE_DSN", "")
	t.Setenv("MAGIC_CUT_REDIS_ADDR", "")
	t.Setenv("MAGIC_CUT_REDIS_PREFIX", "")
	t.Setenv("MAGIC_CUT_STORAGE_MODE", "")
	t.Setenv("MAGIC_CUT_STORAGE_ROOT", "")
	t.Setenv("MAGIC_CUT_TOS_PREFIX", "")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	if cfg.ListenAddr != ":4040" {
		t.Fatalf("ListenAddr = %q, want :4040", cfg.ListenAddr)
	}
	if cfg.AuthBaseURL != "http://127.0.0.1:4001" {
		t.Fatalf("AuthBaseURL = %q", cfg.AuthBaseURL)
	}
	if cfg.DatabaseName != "qiantie_magic_cut_dev" {
		t.Fatalf("DatabaseName = %q", cfg.DatabaseName)
	}
	if cfg.RedisPrefix != "magic_cut:dev:" {
		t.Fatalf("RedisPrefix = %q", cfg.RedisPrefix)
	}
	if cfg.TOSPrefix != "magic-cut/dev" {
		t.Fatalf("TOSPrefix = %q", cfg.TOSPrefix)
	}
}

func TestLoadRejectsSharedBusinessDatabase(t *testing.T) {
	t.Setenv("MAGIC_CUT_DATABASE_NAME", "qiantie")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() accepted the shared business database")
	}
}
