package batchfactoryv11

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

// HTTPVideoAdapter is the one server-side video provider adapter used by V11.
// Its wire format is intentionally small and provider-neutral; a provider can
// expose a compatible endpoint without reintroducing the legacy Node route.
type HTTPVideoAdapter struct {
	Endpoint     string
	PollEndpoint string
	APIKey       string
	Model        string
	Client       *http.Client
	ValidateURL  func(string) (*url.URL, error)
}

var _ ProductionAdapter = (*HTTPVideoAdapter)(nil)
var _ ProductionPoller = (*HTTPVideoAdapter)(nil)

type ProductionPoller interface {
	Poll(context.Context, FrozenVideoModel, ProviderTaskRef) (ProviderTaskRef, error)
}

// ProductionAdapterPoller is the complete provider contract used by the
// script-video compatibility routes: submit a task and poll its result.
// Keeping the combined interface here avoids duplicating that contract in the
// HTTP layer while retaining the smaller interfaces for production services.
type ProductionAdapterPoller interface {
	ProductionAdapter
	ProductionPoller
}

func (a *HTTPVideoAdapter) Validate() error {
	if _, err := a.validatedURL(strings.TrimSpace(a.Endpoint)); err != nil {
		return fmt.Errorf("video provider endpoint: %w", err)
	}
	if strings.TrimSpace(a.APIKey) == "" {
		return fmt.Errorf("video provider API key is required")
	}
	if strings.TrimSpace(a.Model) == "" {
		return fmt.Errorf("video provider model is required")
	}
	if strings.TrimSpace(a.PollEndpoint) != "" {
		if _, err := a.validatedURL(strings.TrimSpace(a.PollEndpoint)); err != nil {
			return fmt.Errorf("video provider poll endpoint: %w", err)
		}
	}
	return nil
}

func validateProductionURL(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil {
		return nil, fmt.Errorf("must be an absolute HTTPS URL without userinfo")
	}
	return u, nil
}

func (a *HTTPVideoAdapter) validatedURL(raw string) (*url.URL, error) {
	if a.ValidateURL != nil {
		return a.ValidateURL(raw)
	}
	return validateProductionURL(raw)
}

func (a *HTTPVideoAdapter) Submit(ctx context.Context, model FrozenVideoModel, prompt FinalPrompt) (ProviderTaskRef, error) {
	if err := a.Validate(); err != nil {
		return ProviderTaskRef{}, err
	}
	endpoint, err := a.validatedURL(a.Endpoint)
	if err != nil {
		return ProviderTaskRef{}, fmt.Errorf("validate video endpoint: %w", err)
	}
	if strings.TrimSpace(model.ID) == "" || model.ID != a.Model {
		return ProviderTaskRef{}, fmt.Errorf("frozen video model does not match configured provider model")
	}
	values := prompt.EffectiveSettings.Values
	duration := prompt.DurationSeconds
	if duration <= 0 {
		duration = rawInt(values, "duration", 0)
	}
	if duration <= 0 {
		duration = model.MaxDuration
	}
	payload := map[string]any{
		"model":       a.Model,
		"prompt":      prompt.CompiledPrompt,
		"duration":    duration,
		"aspectRatio": EffectiveVideoAspectRatio(values),
	}
	if len(prompt.ReferenceImageURLs) > 0 {
		payload["referenceImageUrls"] = prompt.ReferenceImageURLs
	}
	if resolution := EffectiveVideoResolution(values); resolution != "" {
		payload["resolution"] = resolution
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return ProviderTaskRef{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), bytes.NewReader(body))
	if err != nil {
		return ProviderTaskRef{}, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+a.APIKey)
	response, err := a.doJSON(request)
	if err != nil {
		return ProviderTaskRef{}, err
	}
	providerTaskID := firstVideoValue(response, "providerTaskId", "provider_task_id", "taskId", "task_id", "id")
	mediaURL := firstVideoValue(response, "mediaUrl", "media_url", "videoUrl", "video_url", "url")
	actualDuration := firstVideoFloat(response, "actualDurationSeconds", "actual_duration_seconds", "durationSeconds", "duration_seconds", "duration")
	if mediaURL != "" {
		if _, err := a.validatedURL(mediaURL); err != nil {
			return ProviderTaskRef{}, fmt.Errorf("provider returned unsafe media URL: %w", err)
		}
		return ProviderTaskRef{ProviderTaskID: providerTaskID, State: ProductionSucceeded, MediaURL: mediaURL, RequestedDurationSeconds: float64(duration), ActualDurationSeconds: actualDuration}, nil
	}
	if providerTaskID == "" {
		return ProviderTaskRef{}, fmt.Errorf("video provider response did not contain a task id or media URL")
	}
	return ProviderTaskRef{ProviderTaskID: providerTaskID, State: ProductionQueued, RequestedDurationSeconds: float64(duration)}, nil
}

func (a *HTTPVideoAdapter) Poll(ctx context.Context, _ FrozenVideoModel, task ProviderTaskRef) (ProviderTaskRef, error) {
	if err := a.Validate(); err != nil {
		return ProviderTaskRef{}, err
	}
	if strings.TrimSpace(task.ProviderTaskID) == "" {
		return ProviderTaskRef{}, fmt.Errorf("provider task id is required")
	}
	pollTemplate := strings.TrimSpace(a.PollEndpoint)
	if pollTemplate == "" {
		return ProviderTaskRef{}, fmt.Errorf("video provider poll endpoint is required")
	}
	rawEndpoint := strings.ReplaceAll(pollTemplate, "{id}", url.PathEscape(task.ProviderTaskID))
	endpoint, err := a.validatedURL(rawEndpoint)
	if err != nil {
		return ProviderTaskRef{}, fmt.Errorf("validate video poll endpoint: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return ProviderTaskRef{}, err
	}
	request.Header.Set("Authorization", "Bearer "+a.APIKey)
	response, err := a.doJSON(request)
	if err != nil {
		return ProviderTaskRef{}, err
	}
	status := strings.ToLower(firstVideoValue(response, "status", "state"))
	mediaURL := firstVideoValue(response, "mediaUrl", "media_url", "videoUrl", "video_url", "url")
	result := ProviderTaskRef{ProviderTaskID: task.ProviderTaskID, MediaURL: mediaURL, ActualDurationSeconds: firstVideoFloat(response, "actualDurationSeconds", "actual_duration_seconds", "durationSeconds", "duration_seconds", "duration")}
	switch status {
	case "succeeded", "success", "completed", "done":
		if mediaURL == "" {
			return ProviderTaskRef{}, fmt.Errorf("completed video response did not contain media URL")
		}
		if _, err := a.validatedURL(mediaURL); err != nil {
			return ProviderTaskRef{}, fmt.Errorf("provider returned unsafe media URL: %w", err)
		}
		result.State = ProductionSucceeded
	case "failed", "error", "cancelled", "canceled":
		result.State = ProductionFailed
	default:
		result.State = ProductionRunning
	}
	return result, nil
}

func (a *HTTPVideoAdapter) doJSON(request *http.Request) ([]byte, error) {
	client := a.Client
	if client == nil {
		client = &http.Client{Timeout: 120 * time.Second}
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("video provider request failed")
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, 8<<20))
	if err != nil {
		return nil, fmt.Errorf("read video provider response: %w", err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("video provider returned HTTP %d", response.StatusCode)
	}
	return body, nil
}

func firstVideoValue(raw []byte, keys ...string) string {
	var payload any
	if json.Unmarshal(raw, &payload) != nil {
		return ""
	}
	for _, object := range videoObjects(payload) {
		for _, key := range keys {
			if value, ok := object[key]; ok {
				switch typed := value.(type) {
				case string:
					if strings.TrimSpace(typed) != "" {
						return strings.TrimSpace(typed)
					}
				case float64:
					return fmt.Sprintf("%.0f", typed)
				}
			}
		}
	}
	return ""
}

func firstVideoFloat(raw []byte, keys ...string) float64 {
	var payload any
	if json.Unmarshal(raw, &payload) != nil {
		return 0
	}
	for _, object := range videoObjects(payload) {
		for _, key := range keys {
			switch value := object[key].(type) {
			case float64:
				if value > 0 {
					return value
				}
			case string:
				var parsed float64
				if _, err := fmt.Sscan(strings.TrimSpace(value), &parsed); err == nil && parsed > 0 {
					return parsed
				}
			}
		}
	}
	return 0
}

func videoObjects(value any) []map[string]any {
	// Providers do not agree on the envelope around a task.  Some return a
	// task directly in data, while others put it in items/outputs arrays.  Walk
	// every JSON container and let callers choose only the explicit fields they
	// understand; this keeps parsing tolerant without accepting arbitrary data
	// as a task ID or media URL.
	switch typed := value.(type) {
	case map[string]any:
		objects := []map[string]any{typed}
		for _, nested := range typed {
			objects = append(objects, videoObjects(nested)...)
		}
		return objects
	case []any:
		objects := []map[string]any{}
		for _, nested := range typed {
			objects = append(objects, videoObjects(nested)...)
		}
		return objects
	default:
		return nil
	}
}
