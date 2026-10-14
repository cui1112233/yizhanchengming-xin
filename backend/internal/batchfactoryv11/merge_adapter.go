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

// HTTPMergeAdapter speaks the provider-neutral merge contract used by V11.
// The provider receives an ordered source list and must return either a final
// media URL or a task id that can be polled later.
type HTTPMergeAdapter struct {
	Endpoint     string
	PollEndpoint string
	APIKey       string
	Client       *http.Client
	ValidateURL  func(string) (*url.URL, error)
}

var _ MergeAdapter = (*HTTPMergeAdapter)(nil)
var _ MergePoller = (*HTTPMergeAdapter)(nil)

func (a *HTTPMergeAdapter) Validate() error {
	if _, err := a.validatedURL(strings.TrimSpace(a.Endpoint)); err != nil {
		return fmt.Errorf("merge provider endpoint: %w", err)
	}
	if strings.TrimSpace(a.APIKey) == "" {
		return fmt.Errorf("merge provider API key is required")
	}
	if strings.TrimSpace(a.PollEndpoint) != "" {
		if _, err := a.validatedURL(strings.TrimSpace(a.PollEndpoint)); err != nil {
			return fmt.Errorf("merge provider poll endpoint: %w", err)
		}
	}
	return nil
}

func (a *HTTPMergeAdapter) validatedURL(raw string) (*url.URL, error) {
	if a.ValidateURL != nil {
		return a.ValidateURL(raw)
	}
	return validateProductionURL(raw)
}

func (a *HTTPMergeAdapter) Submit(ctx context.Context, batchID string, sources []MergeMedia, options MergeOptions) (MergeJob, error) {
	if err := a.Validate(); err != nil {
		return MergeJob{}, err
	}
	endpoint, err := a.validatedURL(strings.TrimSpace(a.Endpoint))
	if err != nil {
		return MergeJob{}, err
	}
	if strings.TrimSpace(batchID) == "" || len(sources) == 0 {
		return MergeJob{}, fmt.Errorf("%w: merge batch and sources are required", ErrInvalid)
	}
	for index, source := range sources {
		if strings.TrimSpace(source.ProductionJobID) == "" || strings.TrimSpace(source.VideoID) == "" || strings.TrimSpace(source.MediaURL) == "" {
			return MergeJob{}, fmt.Errorf("%w: merge source production job, video identity and media URL are required", ErrInvalid)
		}
		if source.Order != index {
			return MergeJob{}, fmt.Errorf("%w: merge source order must be contiguous", ErrInvalid)
		}
		if _, err := a.validatedURL(strings.TrimSpace(source.MediaURL)); err != nil {
			return MergeJob{}, fmt.Errorf("unsafe merge source URL: %w", err)
		}
	}
	payload := map[string]any{
		"batchId":              batchID,
		"sources":              sources,
		"timingMode":           options.TimingMode,
		"audioSource":          options.AudioSource,
		"audioArtifactId":      options.AudioArtifactID,
		"speed":                options.Speed,
		"ttsSpeed":             options.TTSSpeed,
		"audioDurationSeconds": options.AudioDurationSeconds,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return MergeJob{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), bytes.NewReader(body))
	if err != nil {
		return MergeJob{}, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+a.APIKey)
	response, err := a.doJSON(request)
	if err != nil {
		return MergeJob{}, err
	}
	providerTaskID := firstVideoValue(response, "providerTaskId", "provider_task_id", "taskId", "task_id", "id")
	outputURL := firstVideoValue(response, "outputUrl", "output_url", "mediaUrl", "media_url", "videoUrl", "video_url", "url")
	if outputURL != "" {
		if _, err := a.validatedURL(outputURL); err != nil {
			return MergeJob{}, fmt.Errorf("provider returned unsafe merge media URL: %w", err)
		}
		return MergeJob{ProviderTaskID: providerTaskID, Status: MergeSucceeded, OutputURL: outputURL}, nil
	}
	if providerTaskID == "" {
		return MergeJob{}, fmt.Errorf("merge provider response did not contain a task id or output URL")
	}
	return MergeJob{ProviderTaskID: providerTaskID, Status: MergeQueued}, nil
}

func (a *HTTPMergeAdapter) Poll(ctx context.Context, batchID string, job MergeJob) (MergeJob, error) {
	if err := a.Validate(); err != nil {
		return MergeJob{}, err
	}
	if strings.TrimSpace(job.ProviderTaskID) == "" {
		return MergeJob{}, fmt.Errorf("merge provider task id is required")
	}
	if strings.TrimSpace(a.PollEndpoint) == "" {
		return MergeJob{}, fmt.Errorf("merge provider poll endpoint is required")
	}
	rawEndpoint := strings.ReplaceAll(strings.TrimSpace(a.PollEndpoint), "{id}", url.PathEscape(job.ProviderTaskID))
	rawEndpoint = strings.ReplaceAll(rawEndpoint, "{batchId}", url.PathEscape(batchID))
	endpoint, err := a.validatedURL(rawEndpoint)
	if err != nil {
		return MergeJob{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return MergeJob{}, err
	}
	request.Header.Set("Authorization", "Bearer "+a.APIKey)
	response, err := a.doJSON(request)
	if err != nil {
		return MergeJob{}, err
	}
	result := MergeJob{ProviderTaskID: job.ProviderTaskID, Status: MergeRunning}
	if id := firstVideoValue(response, "providerTaskId", "provider_task_id", "taskId", "task_id", "id"); id != "" {
		result.ProviderTaskID = id
	}
	result.OutputURL = firstVideoValue(response, "outputUrl", "output_url", "mediaUrl", "media_url", "videoUrl", "video_url", "url")
	status := strings.ToLower(firstVideoValue(response, "status", "state"))
	switch status {
	case "succeeded", "success", "completed", "done":
		if result.OutputURL == "" {
			return MergeJob{}, fmt.Errorf("completed merge response did not contain output URL")
		}
		if _, err := a.validatedURL(result.OutputURL); err != nil {
			return MergeJob{}, fmt.Errorf("provider returned unsafe merge media URL: %w", err)
		}
		result.Status = MergeSucceeded
	case "failed", "error", "cancelled", "canceled":
		result.Status = MergeFailed
		result.ErrorMessage = firstVideoValue(response, "error", "errorMessage", "error_message", "message")
	}
	return result, nil
}

func (a *HTTPMergeAdapter) doJSON(request *http.Request) ([]byte, error) {
	client := a.Client
	if client == nil {
		client = &http.Client{Timeout: 120 * time.Second}
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("merge provider request failed")
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, 8<<20))
	if err != nil {
		return nil, fmt.Errorf("read merge provider response: %w", err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("merge provider returned HTTP %d", response.StatusCode)
	}
	return body, nil
}
