package videogen

import (
	"errors"
	"strings"
)

const (
	ProviderPersonalAPI  = "personal_api"
	ProviderDoubaoLocal  = "doubao_local_executor"
	ProviderAutoDLH3     = "autodl_comfyui"
	ProviderYFAISeedance = "yfai_seedance"

	DefaultPersonalModel     = "yd2.0-mini"
	DefaultPersonalCreateURL = "https://ydapi.yadiai.cn/openapi/v1/video/create"
	DefaultPersonalTasksURL  = "https://ydapi.yadiai.cn/openapi/v1/video/tasks"

	DefaultAutoDLH3Model     = "minimax-h3-video"
	DefaultAutoDLH3CreateURL = "https://autodl.art/api/v1/comfyui/comfyui_workflow/{workflow}"
	DefaultAutoDLH3TasksURL  = "https://autodl.art/api/v1/comfyui/comfyui_workflow/result/{id}"
)

type Config struct {
	Provider  string
	APIKey    string
	Model     string
	CreateURL string
	TasksURL  string
	ResultURL string
}

func NormalizeProvider(provider string) string {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "", "personal", "personal_api", "yd_video", "yadi":
		return ProviderPersonalAPI
	case "doubao", "doubao_local", "doubao_local_executor", "local-doubao-executor-video":
		return ProviderDoubaoLocal
	case "h3", "minimax_h3", "autodl", "autodl_comfyui", "autodl_comfyui_video", "minimax-h3-video":
		return ProviderAutoDLH3
	case "yfai", "yfai_seedance", "seedance-2-0-official":
		return ProviderYFAISeedance
	default:
		return strings.ToLower(strings.TrimSpace(provider))
	}
}

func ProviderForModel(modelID, fallback string) string {
	switch strings.ToLower(strings.TrimSpace(modelID)) {
	case "minimax-h3-video":
		return ProviderAutoDLH3
	case "yd2-mini-video", "yd2.0-mini":
		return ProviderPersonalAPI
	case "seedance-2-0-official":
		return ProviderYFAISeedance
	case "local-doubao-executor-video", "doubao-seedance":
		return ProviderDoubaoLocal
	default:
		return NormalizeProvider(fallback)
	}
}

func ModelMatchesProviderModel(selectedModel, providerModel, provider string) bool {
	selectedModel = strings.ToLower(strings.TrimSpace(selectedModel))
	providerModel = strings.ToLower(strings.TrimSpace(providerModel))
	if selectedModel == "" || providerModel == "" { return true }
	if selectedModel == providerModel { return true }
	provider = NormalizeProvider(provider)
	return ProviderForModel(selectedModel, provider) == provider && ProviderForModel(providerModel, provider) == provider
}

func NormalizeConfig(cfg Config) (Config, error) {
	cfg.Provider = NormalizeProvider(cfg.Provider)
	cfg.APIKey = strings.TrimSpace(cfg.APIKey)
	cfg.Model = strings.TrimSpace(cfg.Model)
	cfg.CreateURL = strings.TrimSpace(cfg.CreateURL)
	cfg.TasksURL = strings.TrimSpace(cfg.TasksURL)
	cfg.ResultURL = strings.TrimSpace(cfg.ResultURL)
	switch cfg.Provider {
	case ProviderPersonalAPI:
		if cfg.APIKey == "" { return Config{}, errors.New("personal video API key is required") }
		if cfg.Model == "" { cfg.Model = DefaultPersonalModel }
		if cfg.CreateURL == "" { cfg.CreateURL = DefaultPersonalCreateURL }
		if cfg.TasksURL == "" { cfg.TasksURL = DefaultPersonalTasksURL }
		if cfg.ResultURL == "" { cfg.ResultURL = strings.TrimRight(cfg.TasksURL, "/") + "/{id}/result" }
	case ProviderDoubaoLocal:
		cfg.APIKey = ""
		if cfg.Model == "" { cfg.Model = "doubao-seedance" }
	case ProviderAutoDLH3:
		if cfg.APIKey == "" { return Config{}, errors.New("AutoDL H3 API key is required") }
		if cfg.Model == "" { cfg.Model = DefaultAutoDLH3Model }
		if cfg.CreateURL == "" { cfg.CreateURL = DefaultAutoDLH3CreateURL }
		if cfg.TasksURL == "" { cfg.TasksURL = DefaultAutoDLH3TasksURL }
	case ProviderYFAISeedance:
		if cfg.APIKey == "" { return Config{}, errors.New("YFAI Seedance API key is required") }
		if cfg.Model == "" { cfg.Model = "seedance-2-0-official" }
		if cfg.CreateURL == "" { cfg.CreateURL = "https://yf.token6688.com" }
	default:
		return Config{}, errors.New("unsupported video provider")
	}
	return cfg, nil
}
