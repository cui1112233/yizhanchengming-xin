package batchfactoryv11

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

const (
	VideoProviderPersonalAPI  = "personal_api"
	VideoProviderDoubaoLocal  = "doubao_local_executor"
	VideoProviderAutoDLH3     = "autodl_comfyui"
	VideoProviderYFAISeedance = "yfai_seedance"
	// VideoProviderYFAI 是“视频引擎”下拉里的 YFAI 官方直连分组。执行层仍复用
	// VideoProviderYFAISeedance 的适配器与密钥，两者只区分配置分组口径。
	VideoProviderYFAI             = "yfai"
	PersonalVideoProviderID       = "yd_video"
	DefaultPersonalVideoModel     = "yd2.0-mini"
	DefaultPersonalVideoCreateURL = "https://ydapi.yadiai.cn/openapi/v1/video/create"
	DefaultPersonalVideoTasksURL  = "https://ydapi.yadiai.cn/openapi/v1/video/tasks"
)

// ErrorCodeVideoProviderNotConfigured 是视频生成通道缺配置的稳定错误码；
// httpapi 会把它透传到响应 code 字段，前端据此显示设置入口。
const ErrorCodeVideoProviderNotConfigured = "VIDEO_PROVIDER_NOT_CONFIGURED"

type notConfiguredError struct{ err error }

func (e notConfiguredError) Error() string     { return e.err.Error() }
func (e notConfiguredError) Unwrap() error     { return e.err }
func (e notConfiguredError) ErrorCode() string { return ErrorCodeVideoProviderNotConfigured }

// wrapNotConfigured 保留原错误文本与 errors.Is 语义，只附加稳定错误码。
func wrapNotConfigured(err error) error {
	if err == nil {
		return nil
	}
	return notConfiguredError{err: err}
}

type VideoProviderConfig struct {
	Provider   string
	APIKey     string
	Model      string
	CreateURL  string
	TasksURL   string
	ResultURL  string
	WorkflowID string
}

type VideoProviderRegistry interface {
	Put(context.Context, string, VideoProviderConfig) error
	Resolve(context.Context, string, string) (VideoProviderConfig, error)
	View(context.Context, string, string) (VideoProviderConfigView, error)
}

type VideoProviderConfigView struct {
	Provider   string `json:"provider"`
	Model      string `json:"model"`
	Configured bool   `json:"configured"`
}

type MemoryVideoProviderRegistry struct {
	mu     sync.RWMutex
	values map[string]VideoProviderConfig
}

func NewMemoryVideoProviderRegistry() *MemoryVideoProviderRegistry {
	return &MemoryVideoProviderRegistry{values: map[string]VideoProviderConfig{}}
}

func (r *MemoryVideoProviderRegistry) Put(_ context.Context, owner string, cfg VideoProviderConfig) error {
	if r == nil {
		return ErrUnavailable
	}
	owner = strings.TrimSpace(owner)
	if owner == "" {
		return fmt.Errorf("%w: owner is required", ErrInvalid)
	}
	normalized, err := normalizeVideoProviderConfig(cfg)
	if err != nil {
		return err
	}
	r.putNormalized(owner, normalized)
	return nil
}

// normalizeVideoProviderConfig applies provider defaults and validates the
// minimum credentials required before a provider can be used. It is shared by
// the in-memory registry and the persistent registry so both paths accept the
// same input.
func normalizeVideoProviderConfig(cfg VideoProviderConfig) (VideoProviderConfig, error) {
	cfg.Provider = normalizeVideoProvider(cfg.Provider)
	switch cfg.Provider {
	case VideoProviderPersonalAPI:
		cfg.Model = strings.TrimSpace(cfg.Model)
		if cfg.Model == "" {
			cfg.Model = DefaultPersonalVideoModel
		}
		cfg.APIKey = strings.TrimSpace(cfg.APIKey)
		if cfg.APIKey == "" {
			return cfg, fmt.Errorf("%w: personal video API key is required", ErrInvalid)
		}
		if strings.TrimSpace(cfg.CreateURL) == "" {
			cfg.CreateURL = DefaultPersonalVideoCreateURL
		}
		if strings.TrimSpace(cfg.TasksURL) == "" {
			cfg.TasksURL = DefaultPersonalVideoTasksURL
		}
		if strings.TrimSpace(cfg.ResultURL) == "" {
			cfg.ResultURL = strings.TrimRight(cfg.TasksURL, "/") + "/{id}/result"
		}
	case VideoProviderDoubaoLocal:
		cfg.APIKey = ""
		cfg.Model = strings.TrimSpace(cfg.Model)
		if cfg.Model == "" {
			cfg.Model = "doubao-seedance"
		}
	case VideoProviderAutoDLH3:
		cfg.APIKey = strings.TrimSpace(cfg.APIKey)
		if cfg.APIKey == "" {
			return cfg, fmt.Errorf("%w: AutoDL H3 API key is required", ErrInvalid)
		}
		cfg.Model = strings.TrimSpace(cfg.Model)
		if cfg.Model == "" {
			cfg.Model = AutoDLH3Model
		}
		if strings.TrimSpace(cfg.CreateURL) == "" {
			cfg.CreateURL = DefaultAutoDLH3CreateURL
		}
		if strings.TrimSpace(cfg.TasksURL) == "" {
			cfg.TasksURL = DefaultAutoDLH3TasksURL
		}
		cfg.WorkflowID = strings.TrimSpace(cfg.WorkflowID)
		if cfg.WorkflowID != "" && !validAutoDLH3WorkflowID(cfg.WorkflowID) {
			return cfg, fmt.Errorf("%w: AutoDL H3 workflow id is invalid", ErrInvalid)
		}
	case VideoProviderYFAISeedance:
		cfg.APIKey = strings.TrimSpace(cfg.APIKey)
		if cfg.APIKey == "" {
			return cfg, fmt.Errorf("%w: YFAI Seedance API key is required", ErrInvalid)
		}
		cfg.Model = strings.TrimSpace(cfg.Model)
		if cfg.Model == "" {
			cfg.Model = "seedance-2-0-official"
		}
		if strings.TrimSpace(cfg.CreateURL) == "" {
			cfg.CreateURL = "https://yf.token6688.com"
		}
	default:
		return cfg, fmt.Errorf("%w: unsupported video provider", ErrInvalid)
	}
	return cfg, nil
}

func (r *MemoryVideoProviderRegistry) putNormalized(owner string, cfg VideoProviderConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.values[owner+"\x00"+cfg.Provider] = cfg
}

func (r *MemoryVideoProviderRegistry) has(owner, provider string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.values[owner+"\x00"+provider]
	return ok
}

func (r *MemoryVideoProviderRegistry) Resolve(_ context.Context, owner, provider string) (VideoProviderConfig, error) {
	if r == nil {
		return VideoProviderConfig{}, ErrUnavailable
	}
	owner = strings.TrimSpace(owner)
	provider = normalizeVideoProvider(provider)
	if owner == "" {
		return VideoProviderConfig{}, fmt.Errorf("%w: owner is required", ErrInvalid)
	}
	if provider == "" {
		provider = VideoProviderPersonalAPI
	}
	r.mu.RLock()
	cfg, ok := r.values[owner+"\x00"+provider]
	r.mu.RUnlock()
	if !ok {
		return VideoProviderConfig{}, wrapNotConfigured(fmt.Errorf("%w: 视频生成通道未配置；请到「批量工厂 → 设置 → 视频生成通道」选择通道并填写密钥后保存。", ErrUnavailable))
	}
	return cfg, nil
}

func (r *MemoryVideoProviderRegistry) View(_ context.Context, owner, provider string) (VideoProviderConfigView, error) {
	cfg, err := r.Resolve(context.Background(), owner, provider)
	if err != nil {
		return VideoProviderConfigView{Provider: normalizeVideoProvider(provider)}, err
	}
	return VideoProviderConfigView{Provider: cfg.Provider, Model: cfg.Model, Configured: strings.TrimSpace(cfg.APIKey) != "" || cfg.Provider == VideoProviderDoubaoLocal}, nil
}

func normalizeVideoProvider(provider string) string {
	normalized := strings.ToLower(strings.TrimSpace(provider))
	// 梵客视频的具体型号保存为 fanke-video-api::<modelId>，执行层统一走梵客适配器。
	if strings.HasPrefix(normalized, "fanke-video-api") {
		return VideoProviderYFAISeedance
	}
	switch normalized {
	case "", "personal", "personal_api", "yd_video", "yadi":
		return VideoProviderPersonalAPI
	case "doubao", "doubao_local", "doubao_local_executor", "local-doubao-executor-video":
		return VideoProviderDoubaoLocal
	case "h3", "minimax_h3", "autodl", "autodl_comfyui", "autodl_comfyui_video", "minimax-h3-video":
		return VideoProviderAutoDLH3
	case "yfai", "yfai_seedance", "seedance-2-0-official":
		return VideoProviderYFAISeedance
	default:
		return normalized
	}
}

const fanKeVideoSelectionPrefix = "fanke-video-api::"

// FanKeModelFromSelection 从 "fanke-video-api::<modelId>" 里取出具体型号；
// 旧值（裸 fanke-video-api）返回空，由调用方回退到目录里的默认型号。
func FanKeModelFromSelection(modelID string) string {
	raw := strings.TrimSpace(modelID)
	if !strings.HasPrefix(strings.ToLower(raw), fanKeVideoSelectionPrefix) {
		return ""
	}
	return strings.TrimSpace(raw[len(fanKeVideoSelectionPrefix):])
}

// VideoCatalogEngines is the complete "视频引擎" grouping enum for book
// settings. YFAI is added alongside the four existing groups; nothing is
// removed or renamed.
var VideoCatalogEngines = []string{
	VideoProviderPersonalAPI,
	VideoProviderDoubaoLocal,
	VideoProviderAutoDLH3,
	VideoProviderYFAI,
	VideoProviderYFAISeedance,
}

// CatalogEngineForModel maps a catalog model to its engine group. It only
// decides the settings grouping; the execution provider is still chosen by
// VideoProviderForModel/normalizeVideoProvider.
func CatalogEngineForModel(modelID, fallback string) string {
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(modelID)), "fanke-video-api") {
		// 梵客聚合站自己的型号（含它的 Seedance 变体）保留在梵客分组。
		return VideoProviderYFAISeedance
	}
	switch strings.ToLower(strings.TrimSpace(modelID)) {
	case "seedance-2-0-official":
		return VideoProviderYFAI
	case "fanke-video-api":
		// 梵客聚合站自己的型号（含它的 Seedance 变体）保留在梵客分组。
		return VideoProviderYFAISeedance
	case "minimax-h3-video":
		return VideoProviderAutoDLH3
	case "yd2-mini-video", "yd2.0-mini":
		return VideoProviderPersonalAPI
	case "local-doubao-executor-video", "doubao-seedance":
		return VideoProviderDoubaoLocal
	}
	if strings.EqualFold(strings.TrimSpace(fallback), VideoProviderYFAI) {
		return VideoProviderYFAI
	}
	return normalizeVideoProvider(fallback)
}

// VideoProviderForModel keeps the saved model selection and the runtime
// provider coherent. Catalog IDs intentionally differ from a few provider
// model IDs, so the model selection is the authoritative choice whenever it
// is one of the built-in video models.
func VideoProviderForModel(modelID, fallback string) string {
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(modelID)), "fanke-video-api") {
		return VideoProviderYFAISeedance
	}
	switch strings.ToLower(strings.TrimSpace(modelID)) {
	case "minimax-h3-video":
		return VideoProviderAutoDLH3
	case "yd2-mini-video", "yd2.0-mini":
		return VideoProviderPersonalAPI
	case "seedance-2-0-official":
		return VideoProviderYFAISeedance
	case "local-doubao-executor-video", "doubao-seedance":
		return VideoProviderDoubaoLocal
	default:
		return normalizeVideoProvider(fallback)
	}
}

// VideoModelMatchesProviderModel accepts catalog/provider aliases for the
// same capability while still rejecting a model selected for another provider.
func VideoModelMatchesProviderModel(selectedModel, providerModel, provider string) bool {
	selectedModel = strings.ToLower(strings.TrimSpace(selectedModel))
	providerModel = strings.ToLower(strings.TrimSpace(providerModel))
	if selectedModel == "" || providerModel == "" {
		return true
	}
	if selectedModel == providerModel {
		return true
	}
	return VideoProviderForModel(selectedModel, provider) == normalizeVideoProvider(provider) &&
		VideoProviderForModel(providerModel, provider) == normalizeVideoProvider(provider)
}

// NormalizeVideoProviderForHTTP normalizes provider names accepted by HTTP clients.
func NormalizeVideoProviderForHTTP(provider string) string { return normalizeVideoProvider(provider) }

// DefaultVideoModelFor returns the built-in model id used when a provider has
// no explicitly saved or catalog-configured model.
func DefaultVideoModelFor(provider string) string {
	switch normalizeVideoProvider(provider) {
	case VideoProviderPersonalAPI:
		return DefaultPersonalVideoModel
	case VideoProviderAutoDLH3:
		return AutoDLH3Model
	case VideoProviderYFAISeedance:
		return "seedance-2-0-official"
	case VideoProviderDoubaoLocal:
		return "doubao-seedance"
	default:
		return ""
	}
}

// IsConfigured reports whether the channel holds an API credential. The local
// doubao channel is judged by the executor status instead, matching the legacy
// provider-status response.
func (cfg VideoProviderConfig) IsConfigured() bool {
	return strings.TrimSpace(cfg.APIKey) != ""
}

// CatalogIDForProvider maps a provider id to the model catalog id used by the
// admin catalog page. Providers without a catalog entry return an empty id.
func CatalogIDForProvider(provider string) string { return catalogIDForProvider(provider) }

// catalogIDForProvider maps the stable provider id to the model catalog id used
// by the admin catalog page.
func catalogIDForProvider(provider string) string {
	switch normalizeVideoProvider(provider) {
	case VideoProviderPersonalAPI:
		return "yd2-mini-video"
	case VideoProviderAutoDLH3:
		return "minimax-h3-video"
	case VideoProviderYFAISeedance:
		return "seedance-2-0-official"
	default:
		return ""
	}
}

// ResolveVideoProviderSetting derives the effective provider setting from an
// account config document using the same priority as the runtime registry:
// explicitly saved videoProviders > managed model catalog > legacy video keys.
// It is read-only and never exposes credentials.
func ResolveVideoProviderSetting(config map[string]any, rawProvider string) VideoProviderConfig {
	provider := normalizeVideoProvider(rawProvider)
	cfg := VideoProviderConfig{Provider: provider, Model: DefaultVideoModelFor(provider)}
	if config == nil {
		return cfg
	}
	if providers, ok := config["videoProviders"].(map[string]any); ok {
		if entry, ok := providers[provider].(map[string]any); ok {
			cfg.APIKey = configString(entry, "apiKey")
			if model := configString(entry, "model"); model != "" {
				cfg.Model = model
			}
			return cfg
		}
	}
	catalogVersion := configNumber(config["modelCatalogVersion"])
	if catalogID := catalogIDForProvider(provider); catalogID != "" && catalogVersion >= 1 {
		models, _ := config["modelCatalog"].([]any)
		for _, item := range models {
			entry, ok := item.(map[string]any)
			if !ok || configString(entry, "id") != catalogID {
				continue
			}
			cfg.APIKey = configString(entry, "credential")
			if model := configString(entry, "modelId"); model != "" {
				cfg.Model = model
			}
			return cfg
		}
	}
	video, _ := config["video"].(map[string]any)
	switch provider {
	case VideoProviderPersonalAPI:
		cfg.APIKey = configString(video, "ydApiKey")
		if cfg.APIKey == "" {
			cfg.APIKey = configString(video, "apiKey")
		}
	case VideoProviderAutoDLH3:
		cfg.APIKey = configString(video, "h3ApiKey")
		if cfg.APIKey == "" {
			cfg.APIKey = configString(video, "apiKey")
		}
	}
	return cfg
}
