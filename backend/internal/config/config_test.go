package config

import (
	"reflect"
	"testing"
	"time"
)

func TestFromEnvRequiresMySQLAndBridgeSecret(t *testing.T) {
	t.Setenv("QIANTIE_MYSQL_DSN", "")
	t.Setenv("QIANTIE_BRIDGE_SECRET", "")
	if _, err := FromEnv(); err == nil {
		t.Fatal("expected missing DSN error")
	}
	t.Setenv("QIANTIE_MYSQL_DSN", "user:pass@tcp(mysql:3306)/qiantie")
	if _, err := FromEnv(); err == nil {
		t.Fatal("expected missing bridge secret error")
	}
}

// keepUnstagedSlice pins the slice to 0 for tests that only exercise unrelated
// config fields; the shipped default is 3 (see TestFromEnvDefaultsToSliceThree).
func keepUnstagedSlice(t *testing.T) {
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_SLICE", "0")
}

func TestFromEnvReadsOptionalGiantMaterialPlugin(t *testing.T) {
	keepUnstagedSlice(t)
	t.Setenv("QIANTIE_MYSQL_DSN", "mysql")
	t.Setenv("QIANTIE_BRIDGE_SECRET", "bridge")
	t.Setenv("QIANTIE_GIANT_MATERIAL_PLUGIN_URL", "http://127.0.0.1:18180")
	t.Setenv("QIANTIE_GIANT_MATERIAL_PLUGIN_TOKEN", "plugin-token")
	cfg, err := FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.GiantMaterialPluginURL != "http://127.0.0.1:18180" || cfg.GiantMaterialPluginToken != "plugin-token" {
		t.Fatalf("giant material plugin config = %#v", cfg)
	}
}

func TestFromEnvReadsN8ComicOpenAPIConfig(t *testing.T) {
	keepUnstagedSlice(t)
	t.Setenv("QIANTIE_MYSQL_DSN", "mysql")
	t.Setenv("QIANTIE_BRIDGE_SECRET", "bridge")
	t.Setenv("QIANTIE_N8_COMIC_BASE_URL", "")
	t.Setenv("QIANTIE_N8_COMIC_OPEN_API_KEY", "server-only-key")
	cfg, err := FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.N8ComicBaseURL != "https://n8-view.hnqingyuwen.top/center-api" || cfg.N8ComicOpenAPIKey != "server-only-key" {
		t.Fatalf("n8 comic config = %#v", cfg)
	}
}

func TestFromEnvReadsUniqueN8ComicProducerAccounts(t *testing.T) {
	keepUnstagedSlice(t)
	t.Setenv("QIANTIE_MYSQL_DSN", "mysql")
	t.Setenv("QIANTIE_BRIDGE_SECRET", "bridge")
	t.Setenv("QIANTIE_N8_COMIC_PRODUCER_ACCOUNTS", " 崔家铭,张三，崔家铭\n李四 ")
	cfg, err := FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"崔家铭", "张三", "李四"}
	if !reflect.DeepEqual(cfg.N8ComicProducerAccounts, want) {
		t.Fatalf("producer accounts = %#v, want %#v", cfg.N8ComicProducerAccounts, want)
	}
}

func TestFromEnvDefaultsToSliceThree(t *testing.T) {
	t.Setenv("QIANTIE_MYSQL_DSN", "user:pass@tcp(mysql:3306)/qiantie")
	t.Setenv("QIANTIE_BRIDGE_SECRET", "secret")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_SLICE", "")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_TEXT_ENDPOINT", "https://example.com/v1/chat/completions")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_TEXT_API_KEY", "text-key")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_TEXT_MODEL", "text-model")
	cfg, err := FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Slice != 3 {
		t.Fatalf("default slice=%d, want 3", cfg.Slice)
	}
}

func TestFromEnvRejectsUnreleasedSliceNumber(t *testing.T) {
	t.Setenv("QIANTIE_MYSQL_DSN", "user:pass@tcp(mysql:3306)/qiantie")
	t.Setenv("QIANTIE_BRIDGE_SECRET", "secret")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_SLICE", "4")
	if _, err := FromEnv(); err == nil {
		t.Fatal("expected invalid slice error")
	}
}

func TestFromEnvRequiresDirectorProviderForSliceTwo(t *testing.T) {
	t.Setenv("QIANTIE_MYSQL_DSN", "user:pass@tcp(mysql:3306)/qiantie")
	t.Setenv("QIANTIE_BRIDGE_SECRET", "secret")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_SLICE", "2")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_TEXT_ENDPOINT", "")
	if _, err := FromEnv(); err == nil {
		t.Fatal("expected missing Director provider error")
	}
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_TEXT_ENDPOINT", "https://example.com/v1/chat/completions")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_TEXT_API_KEY", "test-key")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_TEXT_MODEL", "test-model")
	if _, err := FromEnv(); err != nil {
		t.Fatalf("unexpected configured Slice 2 error: %v", err)
	}
}

func TestFromEnvAcceptsDynamicTextProviderForSliceTwo(t *testing.T) {
	t.Setenv("QIANTIE_MYSQL_DSN", "mysql")
	t.Setenv("QIANTIE_BRIDGE_SECRET", "bridge")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_SLICE", "2")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_TEXT_ENDPOINT", "")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_TEXT_API_KEY", "")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_TEXT_MODEL", "")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_DYNAMIC_TEXT_PROVIDER", "1")
	cfg, err := FromEnv()
	if err != nil {
		t.Fatalf("unexpected dynamic Slice 2 error: %v", err)
	}
	if !cfg.DynamicTextProvider {
		t.Fatal("expected dynamic text provider to be enabled")
	}
}

func TestFromEnvAcceptsProductionSliceFourOnlyWithVideoProvider(t *testing.T) {
	t.Setenv("QIANTIE_MYSQL_DSN", "user:pass@tcp(mysql:3306)/qiantie")
	t.Setenv("QIANTIE_BRIDGE_SECRET", "secret")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_SLICE", "4")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_TEXT_ENDPOINT", "https://example.com/v1/chat/completions")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_TEXT_API_KEY", "text-key")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_TEXT_MODEL", "text-model")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_VIDEO_ENDPOINT", "https://video.example/v1/generate")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_VIDEO_POLL_ENDPOINT", "https://video.example/v1/tasks/{id}")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_VIDEO_API_KEY", "video-key")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_VIDEO_MODEL", "video-model")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_PRODUCTION_ENABLED", "1")
	cfg, err := FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Slice != 4 || !cfg.ProductionEnabled || cfg.VideoModel != "video-model" {
		t.Fatalf("cfg=%+v", cfg)
	}
}

func TestFromEnvAcceptsMergeSliceFiveOnlyWithMergeProvider(t *testing.T) {
	t.Setenv("QIANTIE_MYSQL_DSN", "user:pass@tcp(mysql:3306)/qiantie")
	t.Setenv("QIANTIE_BRIDGE_SECRET", "secret")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_SLICE", "5")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_TEXT_ENDPOINT", "https://example.com/v1/chat/completions")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_TEXT_API_KEY", "text-key")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_TEXT_MODEL", "text-model")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_VIDEO_ENDPOINT", "https://video.example/v1/generate")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_VIDEO_POLL_ENDPOINT", "https://video.example/v1/tasks/{id}")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_VIDEO_API_KEY", "video-key")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_VIDEO_MODEL", "video-model")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_PRODUCTION_ENABLED", "1")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_MERGE_ENDPOINT", "https://merge.example/v1/merge")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_MERGE_POLL_ENDPOINT", "https://merge.example/v1/tasks/{id}")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_MERGE_API_KEY", "merge-key")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_MERGE_ENABLED", "1")
	cfg, err := FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Slice != 5 || !cfg.MergeEnabled || cfg.MergeEndpoint == "" {
		t.Fatalf("cfg=%+v", cfg)
	}
}

func TestFromEnvAcceptsExplicitLocalMergeForSliceFiveWithoutRemoteProvider(t *testing.T) {
	t.Setenv("QIANTIE_MYSQL_DSN", "user:pass@tcp(mysql:3306)/qiantie")
	t.Setenv("QIANTIE_BRIDGE_SECRET", "secret")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_SLICE", "5")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_DYNAMIC_TEXT_PROVIDER", "1")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_PRODUCTION_ENABLED", "1")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_LOCAL_MERGE_ENABLED", "1")
	cfg, err := FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.LocalMergeEnabled || !cfg.MergeEnabled || cfg.MergeEndpoint != "" {
		t.Fatalf("cfg=%+v", cfg)
	}
}

func TestFromEnvUsesConfigurableNovelFetchUpstream(t *testing.T) {
	keepUnstagedSlice(t)
	t.Setenv("QIANTIE_MYSQL_DSN", "mysql")
	t.Setenv("QIANTIE_BRIDGE_SECRET", "bridge")
	t.Setenv("QIANTIE_NOVEL_FETCH_121_ENDPOINT", "")
	t.Setenv("QIANTIE_NOVEL_FETCH_121_TIMEOUT_SECONDS", "")
	cfg, err := FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.NovelFetch121Endpoint != "https://txt.121w.com/api.php" || cfg.NovelFetch121Timeout != 20*time.Second {
		t.Fatalf("default novel upstream config = %#v", cfg)
	}

	t.Setenv("QIANTIE_NOVEL_FETCH_121_ENDPOINT", "http://127.0.0.1:8080/api.php")
	t.Setenv("QIANTIE_NOVEL_FETCH_121_TIMEOUT_SECONDS", "5")
	cfg, err = FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.NovelFetch121Endpoint != "http://127.0.0.1:8080/api.php" || cfg.NovelFetch121Timeout != 5*time.Second {
		t.Fatalf("configured novel upstream config = %#v", cfg)
	}
}

func TestFromEnvReadsOptionalRedisRuntimeAddress(t *testing.T) {
	keepUnstagedSlice(t)
	t.Setenv("QIANTIE_MYSQL_DSN", "mysql")
	t.Setenv("QIANTIE_BRIDGE_SECRET", "bridge")
	t.Setenv("QIANTIE_REDIS_URL", "redis://localhost:6379")
	cfg, err := FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.RedisURL != "redis://localhost:6379" {
		t.Fatalf("redis url = %q", cfg.RedisURL)
	}
}

func TestFromEnvReadsNovelFetchTOSConfig(t *testing.T) {
	keepUnstagedSlice(t)
	t.Setenv("QIANTIE_MYSQL_DSN", "mysql")
	t.Setenv("QIANTIE_BRIDGE_SECRET", "bridge")
	t.Setenv("QIANTIE_NOVEL_FETCH_TOS_ENDPOINT", "https://tos.example.com")
	t.Setenv("QIANTIE_NOVEL_FETCH_TOS_REGION", "cn-test")
	t.Setenv("QIANTIE_NOVEL_FETCH_TOS_BUCKET", "novel-bodies")
	t.Setenv("QIANTIE_NOVEL_FETCH_TOS_ACCESS_KEY", "access")
	t.Setenv("QIANTIE_NOVEL_FETCH_TOS_SECRET_KEY", "secret")
	cfg, err := FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.NovelFetchTOSBucket != "novel-bodies" || cfg.NovelFetchTOSEndpoint != "https://tos.example.com" || cfg.NovelFetchTOSRegion != "cn-test" {
		t.Fatalf("novel fetch TOS config=%#v", cfg)
	}
}

func TestFromEnvReadsOptionalLocalAgentAssetDirectory(t *testing.T) {
	keepUnstagedSlice(t)
	t.Setenv("QIANTIE_MYSQL_DSN", "mysql")
	t.Setenv("QIANTIE_BRIDGE_SECRET", "bridge")
	t.Setenv("QIANTIE_AGENT_ASSET_LOCAL_DIR", `F:\local-review\agent-assets`)
	cfg, err := FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AgentAssetLocalDir != `F:\local-review\agent-assets` {
		t.Fatalf("agent asset local dir=%q", cfg.AgentAssetLocalDir)
	}
}

func TestFromEnvUsesLocalAgentAssetDirectoryByDefault(t *testing.T) {
	keepUnstagedSlice(t)
	t.Setenv("QIANTIE_MYSQL_DSN", "mysql")
	t.Setenv("QIANTIE_BRIDGE_SECRET", "bridge")
	t.Setenv("QIANTIE_AGENT_ASSET_LOCAL_DIR", "")
	cfg, err := FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AgentAssetLocalDir != "data/agent-assets" {
		t.Fatalf("default agent asset local dir=%q", cfg.AgentAssetLocalDir)
	}
}

func TestFromEnvUsesSharedTOSConfigForNovelBodies(t *testing.T) {
	keepUnstagedSlice(t)
	t.Setenv("QIANTIE_MYSQL_DSN", "mysql")
	t.Setenv("QIANTIE_BRIDGE_SECRET", "bridge")
	t.Setenv("QIANTIE_NOVEL_FETCH_TOS_ENDPOINT", "")
	t.Setenv("QIANTIE_NOVEL_FETCH_TOS_REGION", "")
	t.Setenv("QIANTIE_NOVEL_FETCH_TOS_BUCKET", "")
	t.Setenv("QIANTIE_NOVEL_FETCH_TOS_ACCESS_KEY", "")
	t.Setenv("QIANTIE_NOVEL_FETCH_TOS_SECRET_KEY", "")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_MERGE_TOS_ENDPOINT", "https://shared-tos.example.com")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_MERGE_TOS_REGION", "cn-shared")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_MERGE_TOS_BUCKET", "shared-bucket")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_MERGE_TOS_ACCESS_KEY", "shared-access")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_MERGE_TOS_SECRET_KEY", "shared-secret")
	cfg, err := FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.NovelFetchTOSEndpoint != "https://shared-tos.example.com" || cfg.NovelFetchTOSBucket != "shared-bucket" {
		t.Fatalf("novel fetch did not reuse shared TOS config: %#v", cfg)
	}
}

func TestFromEnvUsesNovelFetchTOSForAgentAssetsWhenUnset(t *testing.T) {
	keepUnstagedSlice(t)
	t.Setenv("QIANTIE_MYSQL_DSN", "mysql")
	t.Setenv("QIANTIE_BRIDGE_SECRET", "bridge")
	t.Setenv("QIANTIE_NOVEL_FETCH_TOS_ENDPOINT", "https://shared-tos.example.com")
	t.Setenv("QIANTIE_NOVEL_FETCH_TOS_REGION", "cn-shared")
	t.Setenv("QIANTIE_NOVEL_FETCH_TOS_BUCKET", "shared-bucket")
	t.Setenv("QIANTIE_NOVEL_FETCH_TOS_ACCESS_KEY", "shared-access")
	t.Setenv("QIANTIE_NOVEL_FETCH_TOS_SECRET_KEY", "shared-secret")
	t.Setenv("QIANTIE_AGENT_ASSET_TOS_ENDPOINT", "")
	t.Setenv("QIANTIE_AGENT_ASSET_TOS_REGION", "")
	t.Setenv("QIANTIE_AGENT_ASSET_TOS_BUCKET", "")
	t.Setenv("QIANTIE_AGENT_ASSET_TOS_ACCESS_KEY", "")
	t.Setenv("QIANTIE_AGENT_ASSET_TOS_SECRET_KEY", "")

	cfg, err := FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AgentAssetTOSEndpoint != "https://shared-tos.example.com" || cfg.AgentAssetTOSRegion != "cn-shared" || cfg.AgentAssetTOSBucket != "shared-bucket" || cfg.AgentAssetTOSAccessKey != "shared-access" || cfg.AgentAssetTOSSecretKey != "shared-secret" {
		t.Fatalf("agent assets did not reuse Novel Fetch TOS config: %#v", cfg)
	}
}

func TestFromEnvRejectsInvalidNovelFetchUpstreamTimeout(t *testing.T) {
	t.Setenv("QIANTIE_MYSQL_DSN", "mysql")
	t.Setenv("QIANTIE_BRIDGE_SECRET", "bridge")
	t.Setenv("QIANTIE_NOVEL_FETCH_121_TIMEOUT_SECONDS", "0")
	if _, err := FromEnv(); err == nil {
		t.Fatal("expected invalid novel upstream timeout error")
	}
}

func TestFromEnvAcceptsSliceSixWithEncryptedExternalPublishConfig(t *testing.T) {
	t.Setenv("QIANTIE_MYSQL_DSN", "user:pass@tcp(mysql:3306)/qiantie")
	t.Setenv("QIANTIE_BRIDGE_SECRET", "secret")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_SLICE", "6")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_TEXT_ENDPOINT", "https://example.com/v1/chat/completions")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_TEXT_API_KEY", "text-key")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_TEXT_MODEL", "text-model")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_VIDEO_ENDPOINT", "https://video.example/v1/generate")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_VIDEO_POLL_ENDPOINT", "https://video.example/v1/tasks/{id}")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_VIDEO_API_KEY", "video-key")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_VIDEO_MODEL", "video-model")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_PRODUCTION_ENABLED", "1")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_MERGE_ENDPOINT", "https://merge.example/v1/merge")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_MERGE_API_KEY", "merge-key")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_MERGE_ENABLED", "1")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_121_ENABLED", "1")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_121_ENDPOINT", "https://121-gateway.example/submit")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_121_API_KEY", "121-key")
	t.Setenv("QIANTIE_BATCH_FACTORY_V11_CREDENTIALS_KEY", "12345678901234567890123456789012")
	cfg, err := FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Slice != 6 || !cfg.External121Enabled || len(cfg.ExternalCredentialsKey) != 32 {
		t.Fatalf("cfg=%+v", cfg)
	}
}
