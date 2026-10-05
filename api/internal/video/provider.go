package video

import "net/http"

const (
	DefaultPersonalVideoCreateURL = "https://ydapi.yadiai.cn/openapi/v1/video/create"
	DefaultPersonalVideoTasksURL  = "https://ydapi.yadiai.cn/openapi/v1/video/tasks"
)

func ProviderForModel(model string) (string, bool) {
	switch model {
	case ModelYD20Mini, "yd2-mini-video":
		return ProviderPersonalAPI, true
	case "seedance-2-0-official":
		return ProviderYFAISeedance, true
	case "minimax-h3-video":
		return ProviderAutoDLComfyUI, true
	case "local-doubao-executor-video", "doubao-seedance":
		return ProviderDoubaoLocalExecutor, true
	default:
		return "", false
	}
}

type DefaultProviderFactory struct {
	HTTPClient *http.Client
}

func (f DefaultProviderFactory) Build(config ProviderConfig, secret string) (Provider, error) {
	switch config.ProviderKey {
	case ProviderPersonalAPI:
		if config.CreateURL == "" {
			config.CreateURL = DefaultPersonalVideoCreateURL
		}
		if config.TasksURL == "" {
			config.TasksURL = DefaultPersonalVideoTasksURL
		}
		return NewPersonalAPIProvider(config, secret, f.HTTPClient)
	default:
		return nil, providerError(ErrorProviderUnavailable, "provider adapter unavailable", nil)
	}
}
