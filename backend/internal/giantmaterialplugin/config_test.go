package giantmaterialplugin

import "testing"

func TestConfigFromEnvRequiresSeparatePluginAndUpstreamTokens(t *testing.T) {
	_, err := ConfigFromEnv(func(string) string { return "" })
	if err == nil {
		t.Fatal("expected missing plugin token error")
	}

	values := map[string]string{
		"QIANTIE_GIANT_MATERIAL_PLUGIN_TOKEN": "plugin-token",
		"QINGYU_N8_ADMIN_TOKEN":               "upstream-token",
	}
	cfg, err := ConfigFromEnv(func(key string) string { return values[key] })
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ListenAddr != "127.0.0.1:18180" || cfg.UpstreamEndpoint == "" {
		t.Fatalf("config=%#v", cfg)
	}
}
