package videogen

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type YFAIProvider struct {
	Config      Config
	Client      *http.Client
	ValidateURL URLValidator
}

func (p YFAIProvider) Submit(ctx context.Context, request Request) (Task, error) {
	cfg, err := NormalizeConfig(p.Config)
	if err != nil { return Task{}, err }
	if ProviderForModel(request.Model, cfg.Provider) != ProviderYFAISeedance || !ModelMatchesProviderModel(request.Model, cfg.Model, cfg.Provider) {
		return Task{}, errors.New("selected video model does not match YFAI Seedance provider")
	}
	if strings.TrimSpace(request.Prompt) == "" { return Task{}, errors.New("video prompt is required") }
	duration := request.Duration
	if duration < 4 { duration = 4 }
	if duration > 15 { return Task{}, errors.New("Seedance duration exceeds 15 seconds") }
	resolution := strings.TrimSpace(request.Resolution)
	if resolution == "" { resolution = "720p" }
	aspect := strings.TrimSpace(request.AspectRatio)
	if aspect == "" { aspect = "9:16" }
	params := map[string]any{"mode":"text-to-video", "duration":strconv.Itoa(duration), "resolution":resolution, "aspect_ratio":aspect, "quality":"mini", "count":1, "return_last_frame":false}
	if len(request.ReferenceURLs) > 0 {
		images := make([]string, 0, len(request.ReferenceURLs))
		for _, raw := range request.ReferenceURLs {
			raw = strings.TrimSpace(raw)
			if raw == "" { return Task{}, errors.New("YFAI reference image URL is empty") }
			if err := p.validateMediaURL(raw); err != nil { return Task{}, err }
			images = append(images, raw)
		}
		params["mode"] = "reference"
		params["images"] = images
	}
	payload := map[string]any{"model":cfg.Model, "prompt":request.Prompt, "params":params}
	endpoint := strings.TrimRight(cfg.CreateURL, "/") + "/v1/media/generate"
	reply, err := p.do(ctx, http.MethodPost, endpoint, cfg.APIKey, payload)
	if err != nil { return Task{}, err }
	taskID := firstString(reply, "task_id", "taskId")
	if taskID == "" { return Task{}, fmt.Errorf("YFAI Seedance did not create a task: %s", providerFailureReason(reply)) }
	return Task{ID:taskID, State:StateQueued}, nil
}

func (p YFAIProvider) Poll(ctx context.Context, task Task) (Task, error) {
	cfg, err := NormalizeConfig(p.Config)
	if err != nil { return Task{}, err }
	id := strings.TrimSpace(task.ID)
	if id == "" { return Task{}, errors.New("YFAI Seedance task id is required") }
	endpoint := strings.TrimRight(cfg.CreateURL, "/") + "/v1/tasks/" + url.PathEscape(id)
	reply, err := p.do(ctx, http.MethodGet, endpoint, cfg.APIKey, nil)
	if err != nil { return Task{}, err }
	state := strings.ToLower(firstString(reply, "status", "state"))
	switch state {
	case "failed", "error", "cancelled", "canceled":
		return Task{ID:id, State:StateFailed}, nil
	case "completed", "succeeded", "success", "done":
		mediaURL := firstString(reply, "output_url", "url", "result_url")
		if mediaURL == "" { return Task{ID:id, State:StateRunning}, nil }
		if err := p.validateMediaURL(mediaURL); err != nil { return Task{}, fmt.Errorf("YFAI Seedance returned unsafe media URL: %w", err) }
		return Task{ID:id, State:StateSucceeded, MediaURL:mediaURL}, nil
	default:
		return Task{ID:id, State:StateRunning}, nil
	}
}

func (p YFAIProvider) do(ctx context.Context, method, endpoint, apiKey string, payload any) ([]byte, error) {
	if err := p.validateEndpoint(endpoint); err != nil { return nil, err }
	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil { return nil, err }
		body = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil { return nil, err }
	req.Header.Set("Authorization", "Bearer "+apiKey)
	if payload != nil { req.Header.Set("Content-Type", "application/json") }
	client := p.Client
	if client == nil { client = &http.Client{Timeout:120*time.Second} }
	resp, err := client.Do(req)
	if err != nil { return nil, errors.New("YFAI Seedance request failed") }
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil { return nil, err }
	if resp.StatusCode < 200 || resp.StatusCode >= 300 { return nil, fmt.Errorf("YFAI Seedance returned HTTP %d", resp.StatusCode) }
	return data,nil
}

func (p YFAIProvider) validateEndpoint(raw string) error {
	if p.ValidateURL != nil { return p.ValidateURL(raw) }
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Scheme != "https" || strings.ToLower(u.Hostname()) != "yf.token6688.com" { return errors.New("YFAI endpoint must use https://yf.token6688.com") }
	return nil
}

func (p YFAIProvider) validateMediaURL(raw string) error {
	if p.ValidateURL != nil { return p.ValidateURL(raw) }
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || (u.Scheme!="https" && u.Scheme!="http") || u.Hostname()=="" { return errors.New("invalid YFAI media URL") }
	return nil
}
