package video

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	DefaultPersonalVideoCreateURL = "https://ydapi.yadiai.cn/openapi/v1/video/create"
	DefaultPersonalVideoTasksURL  = "https://ydapi.yadiai.cn/openapi/v1/video/tasks"
	DefaultYFAISeedanceCreateURL  = "https://yf.token6688.com/v1/media/generate"
	DefaultYFAISeedanceTasksURL   = "https://yf.token6688.com/v1/tasks"
	DefaultAutoDLH3CreateURL      = "https://autodl.art/api/v1/comfyui/comfyui_workflow/{workflow}"
	DefaultAutoDLH3TasksURL       = "https://autodl.art/api/v1/comfyui/comfyui_workflow/result/{id}"
)

func ProviderForModel(model string) (string, bool) {
	switch model {
	case ModelYD20Mini, "yd2-mini-video":
		return ProviderPersonalAPI, true
	case ModelSeedance20Official:
		return ProviderYFAISeedance, true
	case ModelMiniMaxH3Video:
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
	case ProviderYFAISeedance:
		if config.CreateURL == "" {
			config.CreateURL = DefaultYFAISeedanceCreateURL
		}
		if config.TasksURL == "" {
			config.TasksURL = DefaultYFAISeedanceTasksURL
		}
		return NewYFAISeedanceProvider(config, secret, f.HTTPClient)
	case ProviderAutoDLComfyUI:
		if config.CreateURL == "" {
			config.CreateURL = DefaultAutoDLH3CreateURL
		}
		if config.TasksURL == "" {
			config.TasksURL = DefaultAutoDLH3TasksURL
		}
		return NewAutoDLComfyUIProvider(config, secret, f.HTTPClient)
	default:
		return nil, providerError(ErrorProviderUnavailable, "provider adapter unavailable", nil)
	}
}

func probeHTTP(ctx context.Context, client *http.Client, endpoint, authorization string) error {
	if err := validateProviderURL(endpoint); err != nil {
		return err
	}
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return providerError(ErrorProviderUnavailable, "create provider status request", err)
	}
	if strings.TrimSpace(authorization) != "" {
		req.Header.Set("Authorization", authorization)
	}
	resp, err := client.Do(req)
	if err != nil {
		return providerError(ErrorProviderUnavailable, "provider status request failed", err)
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		return providerError(ErrorProviderAuthFailed, "provider credential rejected", nil)
	}
	if resp.StatusCode >= 500 {
		return providerError(ErrorProviderUnavailable, fmt.Sprintf("provider returned HTTP %d", resp.StatusCode), nil)
	}
	return nil
}
