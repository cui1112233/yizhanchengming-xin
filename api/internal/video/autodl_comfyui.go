package video

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	autoDLNoImageWorkflow   = "minimax_h3_lightx2v_no_pic"
	autoDLReferenceWorkflow = "minimax_h3_lightx2v_v5_15s"
	maxAutoDLReferenceImages = 9
)

type autoDLComfyUIProvider struct {
	config ProviderConfig
	secret string
	client *http.Client
}

func NewAutoDLComfyUIProvider(config ProviderConfig, secret string, client *http.Client) (Provider, error) {
	config.ProviderKey = strings.TrimSpace(config.ProviderKey)
	config.Model = strings.TrimSpace(config.Model)
	config.CreateURL = strings.TrimSpace(config.CreateURL)
	config.TasksURL = strings.TrimSpace(config.TasksURL)
	secret = strings.TrimSpace(secret)
	if config.ProviderKey != ProviderAutoDLComfyUI || config.Model != ModelMiniMaxH3Video {
		return nil, providerError(ErrorProviderUnavailable, "autodl_comfyui configuration does not match minimax-h3-video", nil)
	}
	if secret == "" {
		return nil, providerError(ErrorProviderUnconfigured, "autodl_comfyui secret is not configured", nil)
	}
	if !strings.Contains(config.CreateURL, "{workflow}") || !strings.Contains(config.TasksURL, "{id}") {
		return nil, providerError(ErrorProviderUnconfigured, "autodl_comfyui endpoint templates are invalid", nil)
	}
	if err := validateProviderURL(strings.ReplaceAll(config.CreateURL, "{workflow}", "probe")); err != nil {
		return nil, err
	}
	if err := validateProviderURL(strings.ReplaceAll(config.TasksURL, "{id}", "probe")); err != nil {
		return nil, err
	}
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	return &autoDLComfyUIProvider{config: config, secret: secret, client: client}, nil
}

func (p *autoDLComfyUIProvider) Submit(ctx context.Context, input SubmitRequest) (SubmitResult, error) {
	if input.Model != "" && input.Model != p.config.Model {
		return SubmitResult{}, providerError(ErrorProviderUnavailable, "requested model is not configured for provider", nil)
	}
	if len(input.ReferenceImageURLs) > maxAutoDLReferenceImages {
		return SubmitResult{}, providerError(ErrorProviderRequestFailed, fmt.Sprintf("autodl_comfyui supports at most %d reference images", maxAutoDLReferenceImages), nil)
	}
	duration := input.DurationSeconds
	if duration <= 0 {
		duration = 5
	}
	resolution := strings.TrimSpace(input.Resolution)
	if resolution == "" {
		resolution = "480p竖"
	}
	workflow := autoDLNoImageWorkflow
	payload := map[string]any{"prompt": input.Prompt, "duration": duration, "resolution": resolution}
	if len(input.ReferenceImageURLs) > 0 {
		workflow = autoDLReferenceWorkflow
	}
	for index, raw := range input.ReferenceImageURLs {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return SubmitResult{}, providerError(ErrorProviderRequestFailed, fmt.Sprintf("autodl_comfyui reference image %d is empty", index), nil)
		}
		if err := validateProviderURL(raw); err != nil {
			return SubmitResult{}, err
		}
		payload[fmt.Sprintf("ref_image_%d", index)] = raw
	}
	endpoint := strings.ReplaceAll(p.config.CreateURL, "{workflow}", workflow)
	decoded, err := p.doJSON(ctx, http.MethodPost, endpoint, payload)
	if err != nil {
		return SubmitResult{}, err
	}
	providerJobID := firstString(decoded, "task_id", "taskId", "id")
	if providerJobID == "" {
		return SubmitResult{}, providerError(ErrorProviderInvalidResponse, "autodl_comfyui response missing task id", nil)
	}
	status := autoDLTaskStatus(firstString(decoded, "status", "state"))
	if status == "" {
		status = TaskQueued
	}
	return SubmitResult{ProviderJobID: providerJobID, Status: status}, nil
}

func (p *autoDLComfyUIProvider) Poll(ctx context.Context, providerJobID string) (PollResult, error) {
	providerJobID = strings.TrimSpace(providerJobID)
	if providerJobID == "" {
		return PollResult{}, providerError(ErrorProviderInvalidResponse, "provider job id is required", nil)
	}
	endpoint := strings.ReplaceAll(p.config.TasksURL, "{id}", url.PathEscape(providerJobID))
	decoded, err := p.doJSON(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return PollResult{}, err
	}
	status := autoDLTaskStatus(firstString(decoded, "status", "state"))
	switch status {
	case TaskSucceeded:
		artifactURL := autoDLResultURL(decoded)
		if artifactURL == "" {
			return PollResult{}, providerError(ErrorProviderInvalidResponse, "autodl_comfyui completed without artifact URL", nil)
		}
		if err := validateProviderURL(artifactURL); err != nil {
			return PollResult{}, providerError(ErrorProviderInvalidResponse, "autodl_comfyui returned invalid artifact URL", err)
		}
		return PollResult{Status: TaskSucceeded, ArtifactURL: artifactURL}, nil
	case TaskFailed:
		return PollResult{Status: TaskFailed, ErrorCode: ErrorProviderRequestFailed, ErrorMessage: "autodl_comfyui task failed"}, nil
	case TaskCancelled:
		return PollResult{Status: TaskCancelled}, nil
	case TaskQueued, TaskRunning:
		return PollResult{Status: status}, nil
	default:
		return PollResult{Status: TaskRunning}, nil
	}
}

func (p *autoDLComfyUIProvider) Cancel(context.Context, string) (CancelResult, error) {
	return CancelResult{Accepted: false}, providerError(ErrorProviderCancelUnsupported, "autodl_comfyui cancel is not supported", nil)
}

func (p *autoDLComfyUIProvider) Probe(ctx context.Context) error {
	endpoint := strings.ReplaceAll(p.config.TasksURL, "{id}", "__status__")
	return probeHTTP(ctx, p.client, endpoint, p.secret)
}

func (p *autoDLComfyUIProvider) doJSON(ctx context.Context, method, endpoint string, payload any) (any, error) {
	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return nil, providerError(ErrorProviderRequestFailed, "encode autodl_comfyui request", err)
		}
		body = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, providerError(ErrorProviderRequestFailed, "create autodl_comfyui request", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", p.secret)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, providerError(ErrorProviderUnavailable, "autodl_comfyui request failed", err)
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, providerError(ErrorProviderInvalidResponse, "read autodl_comfyui response", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		code := ErrorProviderRequestFailed
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			code = ErrorProviderAuthFailed
		} else if resp.StatusCode >= 500 {
			code = ErrorProviderUnavailable
		}
		return nil, providerError(code, fmt.Sprintf("autodl_comfyui returned HTTP %d", resp.StatusCode), nil)
	}
	var decoded any
	if err := json.Unmarshal(responseBody, &decoded); err != nil {
		return nil, providerError(ErrorProviderInvalidResponse, "decode autodl_comfyui response", err)
	}
	return decoded, nil
}

func autoDLTaskStatus(value string) TaskStatus {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "success", "succeeded", "completed", "done":
		return TaskSucceeded
	case "failed", "error":
		return TaskFailed
	case "cancelled", "canceled":
		return TaskCancelled
	case "queued", "submitted", "pending":
		return TaskQueued
	case "running", "processing", "generating":
		return TaskRunning
	default:
		return ""
	}
}

func autoDLResultURL(value any) string {
	var walk func(any) string
	walk = func(current any) string {
		switch v := current.(type) {
		case string:
			if strings.HasPrefix(strings.TrimSpace(v), "http://") || strings.HasPrefix(strings.TrimSpace(v), "https://") {
				return strings.TrimSpace(v)
			}
		case []any:
			for _, child := range v {
				if found := walk(child); found != "" {
					return found
				}
			}
		case map[string]any:
			for _, key := range []string{"results", "url", "video_url", "videoUrl", "media_url", "mediaUrl"} {
				if child, ok := v[key]; ok {
					if found := walk(child); found != "" {
						return found
					}
				}
			}
		}
		return ""
	}
	return walk(value)
}
