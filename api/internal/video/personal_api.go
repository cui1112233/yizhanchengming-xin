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

type personalAPIProvider struct {
	config ProviderConfig
	secret string
	client *http.Client
}

func NewPersonalAPIProvider(config ProviderConfig, secret string, client *http.Client) (Provider, error) {
	config.ProviderKey = strings.TrimSpace(config.ProviderKey)
	config.Model = strings.TrimSpace(config.Model)
	config.CreateURL = strings.TrimSpace(config.CreateURL)
	config.TasksURL = strings.TrimSpace(config.TasksURL)
	config.ResultURL = strings.TrimSpace(config.ResultURL)
	secret = strings.TrimSpace(secret)
	if config.ProviderKey != ProviderPersonalAPI || config.Model != ModelYD20Mini {
		return nil, providerError(ErrorProviderUnavailable, "personal_api configuration does not match yd2.0-mini", nil)
	}
	if secret == "" {
		return nil, providerError(ErrorProviderUnconfigured, "personal_api secret is not configured", nil)
	}
	if config.CreateURL == "" || config.TasksURL == "" {
		return nil, providerError(ErrorProviderUnconfigured, "personal_api endpoints are not configured", nil)
	}
	for _, raw := range []string{config.CreateURL, config.TasksURL} {
		if err := validateProviderURL(raw); err != nil {
			return nil, err
		}
	}
	if config.ResultURL == "" {
		config.ResultURL = strings.TrimRight(config.TasksURL, "/") + "/{id}/result"
	}
	if err := validateProviderURL(strings.ReplaceAll(config.ResultURL, "{id}", "probe")); err != nil {
		return nil, err
	}
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	client = hardenedHTTPClient(client, true)
	return &personalAPIProvider{config: config, secret: secret, client: client}, nil
}

func validateProviderURL(raw string) error {
	if _, err := validateRemoteMediaURL(raw, true); err != nil {
		return providerError(ErrorProviderUnconfigured, "provider endpoint is invalid or targets a blocked address", err)
	}
	return nil
}

func (p *personalAPIProvider) Submit(ctx context.Context, input SubmitRequest) (SubmitResult, error) {
	if input.Model != "" && input.Model != p.config.Model {
		return SubmitResult{}, providerError(ErrorProviderUnavailable, "requested model is not configured for provider", nil)
	}
	payload := map[string]any{
		"model": p.config.Model,
		"prompt": input.Prompt,
	}
	if len(input.ReferenceImageURLs) > 0 {
		payload["image_urls"] = input.ReferenceImageURLs
	}
	if input.DurationSeconds > 0 {
		payload["duration"] = fmt.Sprintf("%d", input.DurationSeconds)
	}
	if input.AspectRatio != "" {
		payload["aspect_ratio"] = input.AspectRatio
	}
	if input.Resolution != "" {
		payload["resolution"] = input.Resolution
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return SubmitResult{}, providerError(ErrorProviderRequestFailed, "encode provider request", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.config.CreateURL, bytes.NewReader(body))
	if err != nil {
		return SubmitResult{}, providerError(ErrorProviderRequestFailed, "create provider request", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.secret)
	resp, err := p.client.Do(req)
	if err != nil {
		return SubmitResult{}, providerError(ErrorProviderUnavailable, "personal_api request failed", err)
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return SubmitResult{}, providerError(ErrorProviderInvalidResponse, "read provider response", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		code := ErrorProviderRequestFailed
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			code = ErrorProviderAuthFailed
		}
		return SubmitResult{}, providerError(code, fmt.Sprintf("personal_api returned HTTP %d", resp.StatusCode), nil)
	}
	var decoded any
	if err := json.Unmarshal(responseBody, &decoded); err != nil {
		return SubmitResult{}, providerError(ErrorProviderInvalidResponse, "decode provider response", err)
	}
	if artifactURL := firstString(decoded, "url", "media_url", "video_url", "result_url"); artifactURL != "" {
		return SubmitResult{Status: TaskSucceeded, ArtifactURL: artifactURL}, nil
	}
	providerJobID := firstString(decoded, "task_id", "taskId", "id", "providerTaskId", "provider_task_id")
	if providerJobID == "" {
		return SubmitResult{}, providerError(ErrorProviderInvalidResponse, "personal_api response missing task id", nil)
	}
	return SubmitResult{ProviderJobID: providerJobID, Status: TaskQueued}, nil
}

func (p *personalAPIProvider) Poll(ctx context.Context, providerJobID string) (PollResult, error) {
	providerJobID = strings.TrimSpace(providerJobID)
	if providerJobID == "" {
		return PollResult{}, providerError(ErrorProviderInvalidResponse, "provider job id is required", nil)
	}
	decoded, err := p.getJSON(ctx, strings.TrimRight(p.config.TasksURL, "/")+"/"+url.PathEscape(providerJobID))
	if err != nil {
		return PollResult{}, err
	}
	remoteStatus := strings.ToLower(firstString(decoded, "status", "state"))
	switch remoteStatus {
	case "queued", "submitted", "pending":
		return PollResult{Status: TaskQueued}, nil
	case "running", "processing", "generating":
		return PollResult{Status: TaskRunning}, nil
	case "failed", "error", "cancelled", "canceled":
		return PollResult{Status: TaskFailed, ErrorCode: ErrorProviderRequestFailed, ErrorMessage: "personal_api task failed"}, nil
	case "success", "succeeded", "completed", "done":
		artifactURL := firstString(decoded, "url", "media_url", "video_url", "result_url", "output_url")
		if artifactURL == "" {
			resultURL := strings.ReplaceAll(p.config.ResultURL, "{id}", url.PathEscape(providerJobID))
			result, err := p.getJSON(ctx, resultURL)
			if err != nil {
				return PollResult{}, err
			}
			artifactURL = firstString(result, "url", "media_url", "video_url", "result_url", "output_url")
		}
		if artifactURL == "" {
			return PollResult{}, providerError(ErrorProviderInvalidResponse, "personal_api completed without artifact URL", nil)
		}
		return PollResult{Status: TaskSucceeded, ArtifactURL: artifactURL}, nil
	default:
		return PollResult{}, providerError(ErrorProviderInvalidResponse, "personal_api returned unknown task status", nil)
	}
}

func (p *personalAPIProvider) Cancel(context.Context, string) (CancelResult, error) {
	return CancelResult{Accepted: false}, providerError(ErrorProviderCancelUnsupported, "personal_api cancel is not implemented in phase 1", nil)
}

func (p *personalAPIProvider) getJSON(ctx context.Context, endpoint string) (any, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, providerError(ErrorProviderRequestFailed, "create provider poll request", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.secret)
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, providerError(ErrorProviderUnavailable, "personal_api poll failed", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, providerError(ErrorProviderInvalidResponse, "read provider poll response", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		code := ErrorProviderRequestFailed
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			code = ErrorProviderAuthFailed
		}
		return nil, providerError(code, fmt.Sprintf("personal_api returned HTTP %d", resp.StatusCode), nil)
	}
	var decoded any
	if err := json.Unmarshal(body, &decoded); err != nil {
		return nil, providerError(ErrorProviderInvalidResponse, "decode provider poll response", err)
	}
	return decoded, nil
}

func firstString(value any, keys ...string) string {
	wanted := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		wanted[key] = struct{}{}
	}
	var walk func(any) string
	walk = func(current any) string {
		switch v := current.(type) {
		case map[string]any:
			for _, key := range keys {
				if raw, ok := v[key]; ok {
					if text, ok := raw.(string); ok && strings.TrimSpace(text) != "" {
						return strings.TrimSpace(text)
					}
				}
			}
			for key, child := range v {
				if _, isWanted := wanted[key]; isWanted {
					continue
				}
				if found := walk(child); found != "" {
					return found
				}
			}
		case []any:
			for _, child := range v {
				if found := walk(child); found != "" {
					return found
				}
			}
		}
		return ""
	}
	return walk(value)
}
