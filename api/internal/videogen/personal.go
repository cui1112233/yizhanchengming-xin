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

const DefaultPersonalFirstFrameURL = "https://tvmao-public.tos-cn-beijing.volces.com/tapnow/empty.png"

type URLValidator func(string) error

type PersonalProvider struct {
	Config      Config
	Client      *http.Client
	ValidateURL URLValidator
}

func (p PersonalProvider) Submit(ctx context.Context, request Request) (Task, error) {
	cfg, err := NormalizeConfig(p.Config)
	if err != nil { return Task{}, err }
	if ProviderForModel(request.Model, cfg.Provider) != ProviderPersonalAPI || !ModelMatchesProviderModel(request.Model, cfg.Model, cfg.Provider) {
		return Task{}, errors.New("selected video model does not match personal API provider")
	}
	if strings.TrimSpace(request.Prompt) == "" { return Task{}, errors.New("video prompt is required") }
	if request.Duration <= 0 { return Task{}, errors.New("video duration is required") }
	if len(request.ReferenceURLs) > 3 { return Task{}, errors.New("personal video supports at most 3 reference images") }
	images := []string{DefaultPersonalFirstFrameURL}
	for _, raw := range request.ReferenceURLs {
		raw = strings.TrimSpace(raw)
		if raw == "" { return Task{}, errors.New("personal video reference image URL is empty") }
		if err := p.validateMediaURL(raw); err != nil { return Task{}, fmt.Errorf("invalid personal video reference image URL: %w", err) }
		images = append(images, raw)
	}
	aspect := strings.TrimSpace(request.AspectRatio)
	if aspect == "" { aspect = "9:16" }
	resolution := strings.TrimSpace(request.Resolution)
	if resolution == "" { resolution = "720p" }
	payload := map[string]any{
		"model": cfg.Model,
		"prompt": request.Prompt,
		"image_urls": images,
		"duration": strconv.Itoa(request.Duration),
		"aspect_ratio": aspect,
		"resolution": resolution,
	}
	reply, err := p.doJSON(ctx, http.MethodPost, cfg.CreateURL, cfg.APIKey, payload)
	if err != nil { return Task{}, err }
	if mediaURL := firstString(reply, "url", "video_url", "videoUrl", "mediaUrl", "media_url"); mediaURL != "" {
		if err := p.validateMediaURL(mediaURL); err != nil { return Task{}, fmt.Errorf("personal video returned unsafe media URL: %w", err) }
		return Task{State: StateSucceeded, MediaURL: mediaURL, DurationMs: uint64(request.Duration) * 1000}, nil
	}
	taskID := firstString(reply, "task_id", "taskId", "id", "providerTaskId", "provider_task_id")
	if taskID == "" {
		reason := providerFailureReason(reply)
		if reason != "" { return Task{}, fmt.Errorf("personal video provider did not create a task: %s", reason) }
		return Task{}, errors.New("personal video response did not contain a task id")
	}
	return Task{ID: taskID, State: StateQueued}, nil
}

func (p PersonalProvider) Poll(ctx context.Context, task Task) (Task, error) {
	cfg, err := NormalizeConfig(p.Config)
	if err != nil { return Task{}, err }
	taskID := strings.TrimSpace(task.ID)
	if taskID == "" { return Task{}, errors.New("personal video task id is required") }
	statusURL := strings.TrimRight(cfg.TasksURL, "/") + "/" + url.PathEscape(taskID)
	reply, err := p.doJSON(ctx, http.MethodGet, statusURL, cfg.APIKey, nil)
	if err != nil { return Task{}, err }
	state := strings.ToUpper(firstString(reply, "status", "state"))
	switch state {
	case "FAILED", "ERROR", "CANCELLED", "CANCELED":
		return Task{ID: taskID, State: StateFailed}, nil
	case "SUCCESS", "SUCCEEDED", "COMPLETED", "DONE":
		resultURL := strings.ReplaceAll(cfg.ResultURL, "{id}", url.PathEscape(taskID))
		result, err := p.doJSON(ctx, http.MethodGet, resultURL, cfg.APIKey, nil)
		if err != nil { return Task{}, err }
		mediaURL := resultMediaURL(result)
		if mediaURL == "" { return Task{}, errors.New("personal video completed without media URL") }
		if err := p.validateMediaURL(mediaURL); err != nil { return Task{}, fmt.Errorf("personal video returned unsafe media URL: %w", err) }
		duration := firstFloat(result, "actualDurationSeconds", "actual_duration_seconds", "durationSeconds", "duration_seconds")
		return Task{ID: taskID, State: StateSucceeded, MediaURL: mediaURL, DurationMs: uint64(duration * 1000)}, nil
	default:
		return Task{ID: taskID, State: StateRunning}, nil
	}
}

func (p PersonalProvider) doJSON(ctx context.Context, method, endpoint, apiKey string, payload any) ([]byte, error) {
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
	if client == nil { client = &http.Client{Timeout: 120 * time.Second} }
	resp, err := client.Do(req)
	if err != nil { return nil, errors.New("personal video provider request failed") }
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil { return nil, errors.New("read personal video provider response") }
	if resp.StatusCode < 200 || resp.StatusCode >= 300 { return nil, fmt.Errorf("personal video provider returned HTTP %d", resp.StatusCode) }
	return data, nil
}

func (p PersonalProvider) validateEndpoint(raw string) error {
	if p.ValidateURL != nil { return p.ValidateURL(raw) }
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Scheme != "https" || strings.ToLower(u.Hostname()) != "ydapi.yadiai.cn" { return errors.New("personal video endpoint must use https://ydapi.yadiai.cn") }
	return nil
}

func (p PersonalProvider) validateMediaURL(raw string) error {
	if p.ValidateURL != nil { return p.ValidateURL(raw) }
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Hostname() == "" { return errors.New("media URL must be http or https") }
	host := strings.ToLower(u.Hostname())
	if host == "localhost" || strings.HasSuffix(host, ".localhost") { return errors.New("local media URL is forbidden") }
	return nil
}

func providerFailureReason(raw []byte) string {
	reason := strings.TrimSpace(firstString(raw, "error_message", "errorMessage", "message", "msg", "detail"))
	if reason == "" || strings.EqualFold(reason, "success") || strings.EqualFold(reason, "ok") { return "" }
	runes := []rune(reason)
	if len(runes) > 240 { reason = string(runes[:240]) }
	return reason
}

func resultMediaURL(raw []byte) string {
	if value := firstString(raw, "url", "video_url", "videoUrl", "mediaUrl", "media_url", "output_url", "result_url"); value != "" { return value }
	var value any
	if json.Unmarshal(raw, &value) != nil { return "" }
	return findURLInValue(value)
}

func findURLInValue(value any) string {
	switch item := value.(type) {
	case map[string]any:
		for _, key := range []string{"url", "video_url", "videoUrl", "media_url", "mediaUrl", "output_url", "result_url"} {
			if raw, ok := item[key].(string); ok && strings.TrimSpace(raw) != "" { return strings.TrimSpace(raw) }
		}
		for _, key := range []string{"urls", "outputs", "results", "data", "result", "output"} {
			if nested, ok := item[key]; ok { if found := findURLInValue(nested); found != "" { return found } }
		}
	case []any:
		for _, nested := range item { if found := findURLInValue(nested); found != "" { return found } }
	case string:
		if strings.HasPrefix(strings.TrimSpace(item), "http://") || strings.HasPrefix(strings.TrimSpace(item), "https://") { return strings.TrimSpace(item) }
	}
	return ""
}

func firstString(raw []byte, keys ...string) string {
	var value any
	if json.Unmarshal(raw, &value) != nil { return "" }
	return findString(value, keys)
}

func findString(value any, keys []string) string {
	switch item := value.(type) {
	case map[string]any:
		for _, key := range keys {
			if raw, ok := item[key]; ok {
				switch v := raw.(type) {
				case string: if strings.TrimSpace(v) != "" { return strings.TrimSpace(v) }
				case float64: return strconv.FormatFloat(v, 'f', -1, 64)
				}
			}
		}
		for _, nested := range item { if found := findString(nested, keys); found != "" { return found } }
	case []any:
		for _, nested := range item { if found := findString(nested, keys); found != "" { return found } }
	}
	return ""
}

func firstFloat(raw []byte, keys ...string) float64 {
	value := firstString(raw, keys...)
	result, _ := strconv.ParseFloat(value, 64)
	return result
}
