package giantmaterialplugin

import (
	"fmt"
	"strings"

	"qiantie/backend/internal/giantmaterialresolver"
)

type Config struct {
	ListenAddr       string
	PluginToken      string
	UpstreamEndpoint string
	UpstreamToken    string
}

func ConfigFromEnv(getenv func(string) string) (Config, error) {
	if getenv == nil {
		return Config{}, fmt.Errorf("environment reader is required")
	}
	cfg := Config{
		ListenAddr:       strings.TrimSpace(getenv("QIANTIE_GIANT_MATERIAL_PLUGIN_LISTEN_ADDR")),
		PluginToken:      strings.TrimSpace(getenv("QIANTIE_GIANT_MATERIAL_PLUGIN_TOKEN")),
		UpstreamEndpoint: strings.TrimSpace(getenv("QINGYU_MATERIAL_SELECT_URL")),
		UpstreamToken:    strings.TrimSpace(getenv("QINGYU_N8_ADMIN_TOKEN")),
	}
	if cfg.ListenAddr == "" {
		cfg.ListenAddr = "127.0.0.1:18180"
	}
	if cfg.UpstreamEndpoint == "" {
		cfg.UpstreamEndpoint = giantmaterialresolver.DefaultMaterialURL
	}
	if cfg.PluginToken == "" {
		return Config{}, fmt.Errorf("QIANTIE_GIANT_MATERIAL_PLUGIN_TOKEN is required")
	}
	if cfg.UpstreamToken == "" {
		return Config{}, fmt.Errorf("QINGYU_N8_ADMIN_TOKEN is required")
	}
	return cfg, nil
}
