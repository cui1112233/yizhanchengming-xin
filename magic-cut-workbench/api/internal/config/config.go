package config

import (
	"errors"
	"os"
	"strings"
)

const isolatedDatabaseName = "qiantie_magic_cut_dev"

type Config struct {
	ListenAddr   string
	AuthBaseURL  string
	DatabaseDSN  string
	DatabaseName string
	RedisAddr    string
	RedisPrefix  string
	StorageMode  string
	StorageRoot  string
	TOSPrefix    string
}

func Load() (Config, error) {
	cfg := Config{
		ListenAddr:   envOrDefault("MAGIC_CUT_LISTEN_ADDR", ":4040"),
		AuthBaseURL:  envOrDefault("MAGIC_CUT_AUTH_BASE_URL", "http://127.0.0.1:4001"),
		DatabaseDSN:  strings.TrimSpace(os.Getenv("MAGIC_CUT_DATABASE_DSN")),
		DatabaseName: envOrDefault("MAGIC_CUT_DATABASE_NAME", isolatedDatabaseName),
		RedisAddr:    envOrDefault("MAGIC_CUT_REDIS_ADDR", "127.0.0.1:6379"),
		RedisPrefix:  envOrDefault("MAGIC_CUT_REDIS_PREFIX", "magic_cut:dev:"),
		StorageMode:  envOrDefault("MAGIC_CUT_STORAGE_MODE", "local"),
		StorageRoot:  envOrDefault("MAGIC_CUT_STORAGE_ROOT", "./data"),
		TOSPrefix:    envOrDefault("MAGIC_CUT_TOS_PREFIX", "magic-cut/dev"),
	}

	if cfg.DatabaseName != isolatedDatabaseName {
		return Config{}, errors.New("MAGIC_CUT_DATABASE_NAME must be qiantie_magic_cut_dev in the isolated workbench")
	}
	if !strings.HasPrefix(cfg.RedisPrefix, "magic_cut:dev:") {
		return Config{}, errors.New("MAGIC_CUT_REDIS_PREFIX must stay inside magic_cut:dev:")
	}
	if cfg.TOSPrefix != "magic-cut/dev" && !strings.HasPrefix(cfg.TOSPrefix, "magic-cut/dev/") {
		return Config{}, errors.New("MAGIC_CUT_TOS_PREFIX must stay inside magic-cut/dev")
	}

	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
