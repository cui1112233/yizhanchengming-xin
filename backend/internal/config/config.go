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
	QingyuN8AdminToken            string
	QingyuMaterialSelectURL       string
	LegacyAccountsPath            string
	PresetPath                    string
	PersonalPromptPath            string
	ScriptHistoryDir              string
	MemberSystemDir               string
	NodeAuthBaseURL               string
}

func FromEnv() (Config, error) {
	rawNovelTimeout := envOr("QIANTIE_NOVEL_FETCH_121_TIMEOUT_SECONDS", "20")
	novelTimeoutSeconds, err := strconv.Atoi(rawNovelTimeout)
	if err != nil || novelTimeoutSeconds <= 0 {
		return Config{}, fmt.Errorf("invalid QIANTIE_NOVEL_FETCH_121_TIMEOUT_SECONDS")
	}
	cfg := Config{ListenAddr: envOr("QIANTIE_GO_LISTEN_ADDR", ":4000"), ReleaseSHA: strings.TrimSpace(os.Getenv("QIANTIE_RELEASE_SHA")), MySQLDSN: strings.TrimSpace(os.Getenv("QIANTIE_MYSQL_DSN")), BridgeSecret: strings.TrimSpace(os.Getenv("QIANTIE_BRIDGE_SECRET")), NodeAuthBaseURL: envOr("QIANTIE_NODE_AUTH_BASE_URL", "http://127.0.0.1:3001"), NovelFetch121Endpoint: envOr("QIANTIE_NOVEL_FETCH_121_ENDPOINT", "https://txt.121w.com/api.php"), NovelFetch121Timeout: time.Duration(novelTimeoutSeconds) * time.Second, DirectorEndpoint: strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_TEXT_ENDPOINT")), DirectorAPIKey: strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_TEXT_API_KEY")), DirectorModel: strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_TEXT_MODEL")), DynamicTextProvider: strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_DYNAMIC_TEXT_PROVIDER")) == "1", VideoEndpoint: strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_VIDEO_ENDPOINT")), VideoPollEndpoint: strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_VIDEO_POLL_ENDPOINT")), VideoAPIKey: strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_VIDEO_API_KEY")), VideoModel: strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_VIDEO_MODEL")), ProductionEnabled: strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_PRODUCTION_ENABLED")) == "1", MergeEndpoint: strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_MERGE_ENDPOINT")), MergePollEndpoint: strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_MERGE_POLL_ENDPOINT")), MergeAPIKey: strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_MERGE_API_KEY")), LocalMergeEnabled: strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_LOCAL_MERGE_ENABLED")) == "1", MergeTOSEndpoint: strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_MERGE_TOS_ENDPOINT")), MergeTOSRegion: strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_MERGE_TOS_REGION")), MergeTOSBucket: strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_MERGE_TOS_BUCKET")), MergeTOSAccessKey: strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_MERGE_TOS_ACCESS_KEY")), MergeTOSSecretKey: strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_MERGE_TOS_SECRET_KEY")), MergeTOSPublicBaseURL: strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_MERGE_TOS_PUBLIC_BASE_URL")), External121Enabled: strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_121_ENABLED")) == "1", External121Endpoint: strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_121_ENDPOINT")), External121APIKey: strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_121_API_KEY")), ExternalYadiEnabled: strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_YADI_ENABLED")) == "1", ExternalYadiEndpoint: strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_YADI_ENDPOINT")), ExternalYadiAPIKey: strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_YADI_API_KEY")), LocalExecutorArtifactDir: envOr("QIANTIE_LOCAL_EXECUTOR_ARTIFACT_DIR", "data/local-executor-artifacts"), LocalExecutorPublicBaseURL: strings.TrimRight(strings.TrimSpace(os.Getenv("QIANTIE_LOCAL_EXECUTOR_PUBLIC_BASE_URL")), "/"), QingyuN8AdminToken: strings.TrimSpace(os.Getenv("QINGYU_N8_ADMIN_TOKEN")), QingyuMaterialSelectURL: envOr("QINGYU_MATERIAL_SELECT_URL", "https://n8.hnqingyuwen.top/center-api/material/video/select")}
	cfg.MergeEnabled = strings.TrimSpace(os.Getenv("QIANTIE_BATCH_FACTORY_V11_MERGE_ENABLED")) == "1" || cfg.LocalMergeEnabled
	cfg.LegacyAccountsPath = strings.TrimSpace(os.Getenv("QIANTIE_LEGACY_ACCOUNTS_PATH"))
	cfg.PresetPath = envOr("QIANTIE_PRESET_PATH", "data/system/presets.json")
	cfg.PersonalPromptPath = envOr("QIANTIE_PERSONAL_PROMPT_PATH", "data/system/script-constraint-prompts.json")
	cfg.ScriptHistoryDir = envOr("QIANTIE_SCRIPT_HISTORY_DIR", "data/users")
	cfg.MemberSystemDir = envOr("QIANTIE_MEMBER_SYSTEM_DIR", "data/system")
	cfg.RedisURL = strings.TrimSpace(os.Getenv("QIANTIE_REDIS_URL"))
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
