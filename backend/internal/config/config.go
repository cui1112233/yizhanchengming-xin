package config

import (
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ListenAddr                    string
	ReleaseSHA                    string
	MySQLDSN                      string
	BridgeSecret                  string
	RedisURL                      string
	NovelFetchTOSEndpoint         string
	NovelFetchTOSRegion           string
	NovelFetchTOSBucket           string
	NovelFetchTOSAccessKey        string
	NovelFetchTOSSecretKey        string
	AgentAssetTOSEndpoint         string
	AgentAssetTOSRegion           string
	AgentAssetTOSBucket           string
	AgentAssetTOSAccessKey        string
	AgentAssetTOSSecretKey        string
	AgentAssetLocalDir            string
	Slice                         int
	DirectorEndpoint              string
	DirectorAPIKey                string
	DirectorModel                 string
	DynamicTextProvider           bool
	VideoEndpoint                 string
	VideoPollEndpoint             string
	VideoAPIKey                   string
	VideoModel                    string
	ProductionEnabled             bool
	MergeEndpoint                 string
	MergePollEndpoint             string
	MergeAPIKey                   string
	MergeEnabled                  bool
	LocalMergeEnabled             bool
	MergeTOSEndpoint              string
	MergeTOSRegion                string
	MergeTOSBucket                string
	MergeTOSAccessKey             string
	MergeTOSSecretKey             string
	MergeTOSPublicBaseURL         string
	ExternalCredentialsKey        []byte
	External121Enabled            bool
	External121Endpoint           string
	External121APIKey             string
	NovelFetch121Endpoint         string
	NovelFetch121Timeout          time.Duration
	ExternalYadiEnabled           bool
	ExternalYadiEndpoint          string
	ExternalYadiAPIKey            string
	LocalExecutorArtifactDir      string
	LocalExecutorArtifactMaxBytes int64
	LocalExecutorPublicBaseURL    string
	LocalExecutorDownloadsDir     string
	QingyuN8AdminToken            string
	QingyuMaterialSelectURL       string
	GiantMaterialPluginURL        string
	GiantMaterialPluginToken      string
	N8ComicBaseURL                string
	N8ComicOpenAPIKey             string
	N8AdminBaseURL                string
	N8ComicProducerAccounts       []string
	LegacyAccountsPath            string
	PresetPath                    string
	AutomationPresetPath          string
	AutomationStatePath           string
	PersonalPromptPath            string
	ScriptHistoryDir              string
	MemberSystemDir               string
	MemberAvatarDir               string
}

func FromEnv() (Config, error) {
	rawNovelTimeout := envOr("QIANTIE_NOVEL_FETCH_121_TIMEOUT_SECONDS", "20")
	novelTimeoutSeconds, err := strconv.Atoi(rawNovelTimeout)
	if err != nil || novelTimeoutSeconds <= 0 {
		return Config{}, fmt.Errorf("invalid QIANTIE_NOVEL_FETCH_121_TIMEOUT_SECONDS")
	}
	cfg := Config{ListenAddr: envOr("QIANTIE_GO_LISTEN_ADDR", ":4000"), ReleaseSHA: strings.TrimSpace(os.Getenv("QIANTIE_RELEASE_SHA")), MySQLDSN: strings.TrimSpace(os.Getenv("QIANTIE_MYSQL_DSN")), BridgeSecret: strings.TrimSpace(os.Getenv("QIANTIE_BRIDGE_SECRET")), NovelFetch121Endpoint: envOr("QIANTIE_NOVEL_FETCH_121_ENDPOINT", "https://txt.121w.com/api.php"), NovelFetch121Timeout: time.Duration(novelTimeoutSeconds) * time.Second, DirectorEndpoint: strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_TEXT_ENDPOINT")), DirectorAPIKey: strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_TEXT_API_KEY")), DirectorModel: strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_TEXT_MODEL")), DynamicTextProvider: strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_DYNAMIC_TEXT_PROVIDER")) == "1", VideoEndpoint: strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_VIDEO_ENDPOINT")), VideoPollEndpoint: strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_VIDEO_POLL_ENDPOINT")), VideoAPIKey: strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_VIDEO_API_KEY")), VideoModel: strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_VIDEO_MODEL")), ProductionEnabled: strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_PRODUCTION_ENABLED")) == "1", MergeEndpoint: strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_MERGE_ENDPOINT")), MergePollEndpoint: strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_MERGE_POLL_ENDPOINT")), MergeAPIKey: strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_MERGE_API_KEY")), LocalMergeEnabled: strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_LOCAL_MERGE_ENABLED")) == "1", MergeTOSEndpoint: strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_MERGE_TOS_ENDPOINT")), MergeTOSRegion: strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_MERGE_TOS_REGION")), MergeTOSBucket: strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_MERGE_TOS_BUCKET")), MergeTOSAccessKey: strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_MERGE_TOS_ACCESS_KEY")), MergeTOSSecretKey: strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_MERGE_TOS_SECRET_KEY")), MergeTOSPublicBaseURL: strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_MERGE_TOS_PUBLIC_BASE_URL")), External121Enabled: strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_121_ENABLED")) == "1", External121Endpoint: strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_121_ENDPOINT")), External121APIKey: strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_121_API_KEY")), ExternalYadiEnabled: strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_YADI_ENABLED")) == "1", ExternalYadiEndpoint: strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_YADI_ENDPOINT")), ExternalYadiAPIKey: strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_YADI_API_KEY")), LocalExecutorArtifactDir: envOr("QIANTIE_LOCAL_EXECUTOR_ARTIFACT_DIR", "data/local-executor-artifacts"), LocalExecutorPublicBaseURL: strings.TrimRight(strings.TrimSpace(os.Getenv("QIANTIE_LOCAL_EXECUTOR_PUBLIC_BASE_URL")), "/"), QingyuN8AdminToken: strings.TrimSpace(os.Getenv("QINGYU_N8_ADMIN_TOKEN")), QingyuMaterialSelectURL: envOr("QINGYU_MATERIAL_SELECT_URL", "https://n8.hnqingyuwen.top/center-api/material/video/select")}
	cfg.GiantMaterialPluginURL = strings.TrimRight(strings.TrimSpace(os.Getenv("QIANTIE_GIANT_MATERIAL_PLUGIN_URL")), "/")
	cfg.GiantMaterialPluginToken = strings.TrimSpace(os.Getenv("QIANTIE_GIANT_MATERIAL_PLUGIN_TOKEN"))
	cfg.N8ComicBaseURL = strings.TrimRight(envOr("QIANTIE_N8_COMIC_BASE_URL", "https://n8-view.hnqingyuwen.top/center-api"), "/")
	cfg.N8ComicOpenAPIKey = strings.TrimSpace(os.Getenv("QIANTIE_N8_COMIC_OPEN_API_KEY"))
	cfg.N8AdminBaseURL = strings.TrimRight(envOr("QIANTIE_N8_ADMIN_BASE_URL", "https://n8-view.hnqingyuwen.top/center-api/auth"), "/")
	cfg.N8ComicProducerAccounts = uniqueEnvValues(os.Getenv("QIANTIE_N8_COMIC_PRODUCER_ACCOUNTS"))
	cfg.LocalExecutorDownloadsDir = envOr("QIANTIE_LOCAL_EXECUTOR_DOWNLOADS_DIR", "data/downloads")
	cfg.MergeEnabled = strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_MERGE_ENABLED")) == "1" || cfg.LocalMergeEnabled
	cfg.LegacyAccountsPath = strings.TrimSpace(os.Getenv("QIANTIE_LEGACY_ACCOUNTS_PATH"))
	cfg.PresetPath = envOr("QIANTIE_PRESET_PATH", "data/system/presets.json")
	cfg.AutomationPresetPath = envOr("QIANTIE_BATCH_FACTORY_AUTOMATION_PRESETS_PATH", "data/batch-factory-v11-automation-presets.json")
	cfg.AutomationStatePath = envOr("QIANTIE_BATCH_FACTORY_AUTOMATION_STATE_PATH", "data/batch-factory-v11-automation.json")
	cfg.PersonalPromptPath = envOr("QIANTIE_PERSONAL_PROMPT_PATH", "data/system/script-constraint-prompts.json")
	cfg.ScriptHistoryDir = envOr("QIANTIE_SCRIPT_HISTORY_DIR", "data/users")
	cfg.MemberSystemDir = envOr("QIANTIE_MEMBER_SYSTEM_DIR", "data/system")
	cfg.MemberAvatarDir = envOr("QIANTIE_MEMBER_AVATAR_DIR", "data/avatars")
	cfg.RedisURL = strings.TrimSpace(os.Getenv("QIANTIE_REDIS_URL"))
	cfg.NovelFetchTOSEndpoint = strings.TrimSpace(os.Getenv("QIANTIE_NOVEL_FETCH_TOS_ENDPOINT"))
	cfg.NovelFetchTOSRegion = strings.TrimSpace(os.Getenv("QIANTIE_NOVEL_FETCH_TOS_REGION"))
	cfg.NovelFetchTOSBucket = strings.TrimSpace(os.Getenv("QIANTIE_NOVEL_FETCH_TOS_BUCKET"))
	cfg.NovelFetchTOSAccessKey = strings.TrimSpace(os.Getenv("QIANTIE_NOVEL_FETCH_TOS_ACCESS_KEY"))
	cfg.NovelFetchTOSSecretKey = strings.TrimSpace(os.Getenv("QIANTIE_NOVEL_FETCH_TOS_SECRET_KEY"))
	cfg.AgentAssetTOSEndpoint = strings.TrimSpace(os.Getenv("QIANTIE_AGENT_ASSET_TOS_ENDPOINT"))
	cfg.AgentAssetTOSRegion = strings.TrimSpace(os.Getenv("QIANTIE_AGENT_ASSET_TOS_REGION"))
	cfg.AgentAssetTOSBucket = strings.TrimSpace(os.Getenv("QIANTIE_AGENT_ASSET_TOS_BUCKET"))
	cfg.AgentAssetTOSAccessKey = strings.TrimSpace(os.Getenv("QIANTIE_AGENT_ASSET_TOS_ACCESS_KEY"))
	cfg.AgentAssetTOSSecretKey = strings.TrimSpace(os.Getenv("QIANTIE_AGENT_ASSET_TOS_SECRET_KEY"))
	// Agent Canvas must always have a durable local target for generated media.
	// Without this fallback, a valid local executor is rejected before it receives
	// work whenever a launcher omits QIANTIE_AGENT_ASSET_LOCAL_DIR.
	cfg.AgentAssetLocalDir = envOr("QIANTIE_AGENT_ASSET_LOCAL_DIR", "data/agent-assets")
	if cfg.NovelFetchTOSEndpoint == "" {
		cfg.NovelFetchTOSEndpoint = cfg.MergeTOSEndpoint
	}
	if cfg.NovelFetchTOSRegion == "" {
		cfg.NovelFetchTOSRegion = cfg.MergeTOSRegion
	}
	if cfg.NovelFetchTOSBucket == "" {
		cfg.NovelFetchTOSBucket = cfg.MergeTOSBucket
	}
	if cfg.NovelFetchTOSAccessKey == "" {
		cfg.NovelFetchTOSAccessKey = cfg.MergeTOSAccessKey
	}
	if cfg.NovelFetchTOSSecretKey == "" {
		cfg.NovelFetchTOSSecretKey = cfg.MergeTOSSecretKey
	}
	if cfg.AgentAssetTOSEndpoint == "" {
		cfg.AgentAssetTOSEndpoint = cfg.NovelFetchTOSEndpoint
	}
	if cfg.AgentAssetTOSRegion == "" {
		cfg.AgentAssetTOSRegion = cfg.NovelFetchTOSRegion
	}
	if cfg.AgentAssetTOSBucket == "" {
		cfg.AgentAssetTOSBucket = cfg.NovelFetchTOSBucket
	}
	if cfg.AgentAssetTOSAccessKey == "" {
		cfg.AgentAssetTOSAccessKey = cfg.NovelFetchTOSAccessKey
	}
	if cfg.AgentAssetTOSSecretKey == "" {
		cfg.AgentAssetTOSSecretKey = cfg.NovelFetchTOSSecretKey
	}
	if cfg.MySQLDSN == "" {
		return Config{}, fmt.Errorf("QIANTIE_MYSQL_DSN is required")
	}
	if cfg.BridgeSecret == "" {
		return Config{}, fmt.Errorf("QIANTIE_BRIDGE_SECRET is required")
	}
	rawSlice := envOr("QIANTIE_BATCH_FACTORY_V11_SLICE", "0")
	slice, err := strconv.Atoi(rawSlice)
	if err != nil || slice < 0 || slice > 6 {
		return Config{}, fmt.Errorf("invalid QIANTIE_BATCH_FACTORY_V11_SLICE")
	}
	cfg.Slice = slice
	rawArtifactMax := envOr("QIANTIE_LOCAL_EXECUTOR_ARTIFACT_MAX_BYTES", "1073741824")
	artifactMax, err := strconv.ParseInt(rawArtifactMax, 10, 64)
	if err != nil || artifactMax <= 0 {
		return Config{}, fmt.Errorf("invalid QIANTIE_LOCAL_EXECUTOR_ARTIFACT_MAX_BYTES")
	}
	cfg.LocalExecutorArtifactMaxBytes = artifactMax
	if cfg.Slice >= 2 && !cfg.DynamicTextProvider && (cfg.DirectorEndpoint == "" || cfg.DirectorAPIKey == "" || cfg.DirectorModel == "") {
		return Config{}, fmt.Errorf("Slice 2+ requires the V11 text provider endpoint, API key, and model")
	}
	if cfg.Slice >= 4 && !cfg.ProductionEnabled {
		return Config{}, fmt.Errorf("Slice 4 requires QIANTIE_BATCH_FACTORY_V11_PRODUCTION_ENABLED=1")
	}
	if cfg.Slice >= 5 && (!cfg.MergeEnabled || (!cfg.LocalMergeEnabled && (cfg.MergeEndpoint == "" || cfg.MergeAPIKey == ""))) {
		return Config{}, fmt.Errorf("Slice 5 requires explicit merge provider configuration and QIANTIE_BATCH_FACTORY_V11_MERGE_ENABLED=1")
	}
	if cfg.Slice >= 6 {
		if cfg.External121Enabled && (cfg.External121Endpoint == "" || cfg.External121APIKey == "") {
			return Config{}, fmt.Errorf("Slice 6 121 requires explicit endpoint and API key")
		}
		if cfg.ExternalYadiEnabled && (cfg.ExternalYadiEndpoint == "" || cfg.ExternalYadiAPIKey == "") {
			return Config{}, fmt.Errorf("Slice 6 Yadi requires explicit endpoint and API key")
		}
		if cfg.External121Enabled || cfg.ExternalYadiEnabled {
			rawKey := strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_CREDENTIALS_KEY"))
			decoded, decodeErr := decodeCredentialKey(rawKey)
			if decodeErr != nil {
				return Config{}, decodeErr
			}
			cfg.ExternalCredentialsKey = decoded
		}
	}
	return cfg, nil
}

func uniqueEnvValues(raw string) []string {
	seen := map[string]struct{}{}
	values := make([]string, 0)
	for _, item := range strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == '，' || r == ';' || r == '；' || r == '\n' || r == '\r'
	}) {
		value := strings.TrimSpace(item)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		values = append(values, value)
	}
	return values
}

func decodeCredentialKey(raw string) ([]byte, error) {
	if raw == "" {
		return nil, fmt.Errorf("Slice 6 external publish requires QIANTIE_BATCH_FACTORY_V11_CREDENTIALS_KEY")
	}
	if decoded, err := base64.StdEncoding.DecodeString(raw); err == nil && len(decoded) == 32 {
		return decoded, nil
	}
	if len(raw) == 32 {
		return []byte(raw), nil
	}
	return nil, fmt.Errorf("QIANTIE_BATCH_FACTORY_V11_CREDENTIALS_KEY must decode to 32 bytes")
}
func envOr(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
