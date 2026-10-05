package video

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type yfaiSeedanceProvider struct {
	config ProviderConfig
	secret string
	client *http.Client
}

func NewYFAISeedanceProvider(config ProviderConfig, secret string, client *http.Client) (Provider, error) {
	config.ProviderKey = strings.TrimSpace(config.ProviderKey)
	config.Model = strings.TrimSpace(config.Model)
	config.CreateURL = strings.TrimSpace(config.CreateURL)
	config.TasksURL = strings.TrimSpace(config.TasksURL)
	secret = strings.TrimSpace(secret)
	if config.ProviderKey != ProviderYFAISeedance || config.Model != ModelSeedance20Official {
		return nil, providerError(ErrorProviderUnavailable, "yfai_seedance configuration does not match seedance-2-0-official", nil)
	}
	if secret == "" {
		return nil, providerError(ErrorProviderUnconfigured, "yfai_seedance secret is not configured", nil)
	}
	if config.CreateURL == "" || config.TasksURL == "" {
		return nil, providerError(ErrorProviderUnconfigured, "yfai_seedance endpoints are not configured", nil)
	}
	if err := validateProviderURL(config.CreateURL); err != nil {
		return nil, err
	}
	if err := validateProviderURL(config.TasksURL); err != nil {
		return nil, err
	}
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	return &yfaiSeedanceProvider{config: config, secret: secret, client: client}, nil
}

func (p *yfaiSeedanceProvider) Submit(ctx context.Context, input SubmitRequest) (SubmitResult, error) {
	if input.Model != "" && input.Model != p.config.Model {
		return SubmitResult{}, providerError(ErrorProviderUnavailable, "requested model is not configured for provider", nil)
	}
	duration := input.DurationSeconds
	if duration < 4 {
		duration = 4
	}
	if duration > 15 {
		return SubmitResult{}, providerError(ErrorProviderRequestFailed, "seedance duration exceeds 15 seconds", nil)
	}
	resolution := strings.TrimSpace(input.Resolution)
	if resolution == "" {
		resolution = "720p"
	}
	aspectRatio := strings.TrimSpace(input.AspectRatio)
	if aspectRatio == "" {
		aspectRatio = "9:16"
	}
	params := map[string]any{
		"mode":              "text-to-video",
		"duration":          strconv.Itoa(duration),
		"resolution":        resolution,
		"aspect_ratio":      aspectRatio,
		"quality":           "mini",
		"count":             1,
		"return_last_frame": false,
	}
	if len(input.ReferenceImageURLs) > 0 {
		params["mode"] = "reference"
		params["images"] = append([]string(nil), input.ReferenceImageURLs...)
	}
	decoded, err := p.doJSON(ctx, http.MethodPost, p.config.CreateURL, map[string]any{
		"model": p.config.Model,
		"prompt": input.Prompt,
		"params": params,
	})
	if err != nil {
		return SubmitResult{}, err
	}
	providerJobID := firstString(decoded, "task_id", "taskId")
	if providerJobID == "" {
		return SubmitResult{}, providerError(ErrorProviderInvalidResponse, "yfai_seedance response missing task id", nil)
	}
	return SubmitResult{ProviderJobID: providerJobID, Status: TaskQueued}, nil
}

func (p *yfaiSeedanceProvider) Poll(ctx context.Context, providerJobID string) (PollResult, error) {
	providerJobID = strings.TrimSpace(providerJobID)
	if providerJobID == "" {
		return PollResult{}, providerError(ErrorProviderInvalidResponse, "provider job id is required", nil)
	}
	decoded, err := p.doJSON(ctx, http.MethodGet, strings.TrimRight(p.config.TasksURL, "/")+"/"+url.PathEscape(providerJobID), nil)
	if err != nil {
		return PollResult{}, err
	}
	state := strings.ToLower(firstString(decoded, "status", "state"))
	switch state {
	case "queued", "submitted", "pending":
		return PollResult{Status: TaskQueued}, nil
	case "running", "processing", "generating":
		return PollResult{Status: TaskRunning}, nil
	case "failed", "error":
		return PollResult{Status: TaskFailed, ErrorCode: ErrorProviderRequestFailed, ErrorMessage: "yfai_seedance task failed"}, nil
	case "cancelled", "canceled":
		return PollResult{Status: TaskCancelled}, nil
	case "completed", "succeeded", "success", "done":
		artifactURL := firstString(decoded, "output_url", "url", "result_url", "media_url", "video_url")
		if artifactURL == "" {
			return PollResult{}, providerError(ErrorProviderInvalidResponse, "yfai_seedance completed without artifact URL", nil)
		}
		return PollResult{Status: TaskSucceeded, ArtifactURL: artifactURL}, nil
	default:
		return PollResult{}, providerError(ErrorProviderInvalidResponse, "yfai_seedance returned unknown task status", nil)
	}
}

func (p *yfaiSeedanceProvider) Cancel(context.Context, string) (CancelResult, error) {
	return CancelResult{Accepted: false}, providerError(ErrorProviderCancelUnsupported, "yfai_seedance cancel is not supported", nil)
}

func (p *yfaiSeedanceProvider) Probe(ctx context.Context) error {
	return probeHTTP(ctx, p.client, strings.TrimRight(p.config.TasksURL, "/"), "Bearer "+p.secret)
}

func (p *yfaiSeedanceProvider) doJSON(ctx context.Context, method, endpoint string, payload any) (any, error) {
	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return nil, providerError(ErrorProviderRequestFailed, "encode yfai_seedance request", err)
		}
		body = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, providerError(ErrorProviderRequestFailed, "create yfai_seedance request", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.secret)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, providerError(ErrorProviderUnavailable, "yfai_seedance request failed", err)
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, providerError(ErrorProviderInvalidResponse, "read yfai_seedance response", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		code := ErrorProviderRequestFailed
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			code = ErrorProviderAuthFailed
		} else if resp.StatusCode >= 500 {
			code = ErrorProviderUnavailable
		}
		return nil, providerError(code, fmt.Sprintf("yfai_seedance returned HTTP %d", resp.StatusCode), nil)
	}
	var decoded any
	if err := json.Unmarshal(responseBody, &decoded); err != nil {
		return nil, providerError(ErrorProviderInvalidResponse, "decode yfai_seedance response", err)
	}
	return decoded, nil
}
