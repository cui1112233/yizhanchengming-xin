package batchfactoryv11

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

	"qiantie/backend/internal/safehttp"
)

type YFAISeedanceAdapter struct {
	BaseURL     string
	APIKey      string
	Model       string
	Client      *http.Client
	ValidateURL func(string) (*url.URL, error)
}

var _ ProductionAdapter = (*YFAISeedanceAdapter)(nil)
var _ ProductionPoller = (*YFAISeedanceAdapter)(nil)

func (a *YFAISeedanceAdapter) endpoint(path string) (*url.URL, error) {
	raw := strings.TrimRight(strings.TrimSpace(a.BaseURL), "/") + path
	if a.ValidateURL != nil {
		return a.ValidateURL(raw)
	}
	u, err := validateProductionURL(raw)
	if err != nil {
		return nil, err
	}
	if strings.ToLower(u.Hostname()) != "yf.token6688.com" {
		return nil, fmt.Errorf("endpoint host must be yf.token6688.com")
	}
	return u, nil
}

func (a *YFAISeedanceAdapter) Validate() error {
	if strings.TrimSpace(a.APIKey) == "" {
		return fmt.Errorf("YFAI Seedance API key is required")
	}
	if strings.TrimSpace(a.Model) == "" {
		return fmt.Errorf("YFAI Seedance model is required")
	}
	_, err := a.endpoint("/v1/media/generate")
	return err
}

func (a *YFAISeedanceAdapter) Submit(ctx context.Context, model FrozenVideoModel, prompt FinalPrompt) (ProviderTaskRef, error) {
	if err := a.Validate(); err != nil {
		return ProviderTaskRef{}, err
	}
	if model.ID != a.Model {
		return ProviderTaskRef{}, fmt.Errorf("frozen video model does not match YFAI Seedance model")
	}
	duration := int(prompt.DurationSeconds)
	if duration < 4 {
		duration = 4
	}
	if duration > 15 {
		return ProviderTaskRef{}, fmt.Errorf("Seedance duration exceeds 15 seconds")
	}
	values := prompt.EffectiveSettings.Values
	images := append([]string(nil), prompt.ReferenceImageURLs...)
	params := map[string]any{"mode": "text-to-video", "duration": strconv.Itoa(duration), "resolution": EffectiveVideoResolution(values), "aspect_ratio": EffectiveVideoAspectRatio(values), "quality": "mini", "count": 1, "return_last_frame": false}
	if len(images) > 0 {
		params["mode"] = "reference"
		params["images"] = images
	}
	payload := map[string]any{"model": a.Model, "prompt": prompt.CompiledPrompt, "params": params}
	endpoint, _ := a.endpoint("/v1/media/generate")
	reply, err := a.do(ctx, http.MethodPost, endpoint.String(), payload)
	if err != nil {
		return ProviderTaskRef{}, err
	}
	taskID := firstVideoValue(reply, "task_id", "taskId")
	if taskID == "" {
		reason := yadiProviderFailureReason(reply)
		if reason == "" {
			reason = firstVideoValue(reply, "error", "errorMessage", "error_message")
		}
		return ProviderTaskRef{}, fmt.Errorf("YFAI Seedance did not create a task: %s", reason)
	}
	return ProviderTaskRef{ProviderTaskID: taskID, State: ProductionQueued, RequestedDurationSeconds: float64(duration)}, nil
}

func (a *YFAISeedanceAdapter) Poll(ctx context.Context, _ FrozenVideoModel, task ProviderTaskRef) (ProviderTaskRef, error) {
	endpoint, err := a.endpoint("/v1/tasks/" + url.PathEscape(task.ProviderTaskID))
	if err != nil {
		return ProviderTaskRef{}, err
	}
	reply, err := a.do(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return ProviderTaskRef{}, err
	}
	state := strings.ToLower(firstVideoValue(reply, "status", "state"))
	isFinal := firstVideoBool(reply, "is_final", "isFinal", "final")
	result := ProviderTaskRef{ProviderTaskID: task.ProviderTaskID, State: ProductionRunning}
	if state == "failed" || state == "error" || state == "cancelled" || state == "canceled" {
		result.State = ProductionFailed
		result.ErrorMessage = firstVideoValue(reply, "error", "errorMessage", "error_message", "message", "msg")
		return result, nil
	}
	// Upstream asks callers to close the loop on is_final first and to accept
	// status=completed/failed as a fallback. A completed task without an
	// output_url is not a result yet (output may still be copied to the CDN).
	if isFinal || state == "completed" || state == "succeeded" || state == "success" || state == "done" {
		media := firstVideoValue(reply, "output_url", "url", "result_url")
		if media != "" {
			result.State, result.MediaURL = ProductionSucceeded, media
			return result, nil
		}
	}
	// A non-terminal task that carries an error is a failure; do not keep
	// polling it forever.
	if message := firstVideoValue(reply, "error", "errorMessage", "error_message"); message != "" {
		result.State, result.ErrorMessage = ProductionFailed, message
	}
	return result, nil
}

func (a *YFAISeedanceAdapter) do(ctx context.Context, method, endpoint string, payload any) ([]byte, error) {
	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+a.APIKey)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	client := a.Client
	if client == nil {
		client = safehttp.NewClientWithTimeout(120 * time.Second)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("YFAI Seedance request failed")
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, newYFAIUpstreamError(resp.StatusCode, resp.Header.Get("Retry-After"))
	}
	return data, nil
}

// YFAIUpstreamError keeps the upstream HTTP status visible to callers so the
// UI can tell an invalid key (401), missing balance (402) and rate limiting
// (429) apart instead of collapsing everything into "unknown error".
type YFAIUpstreamError struct {
	StatusCode int
	RetryAfter time.Duration
	Message    string
}

func (e *YFAIUpstreamError) Error() string { return e.Message }

const yfaiDefaultRetryAfter = 30 * time.Second

func newYFAIUpstreamError(statusCode int, retryAfterHeader string) *YFAIUpstreamError {
	err := &YFAIUpstreamError{StatusCode: statusCode}
	switch statusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		err.Message = "YFAI Seedance returned HTTP 401（密钥无效或无权限，请检查 API Key）"
		if statusCode == http.StatusForbidden {
			err.Message = "YFAI Seedance returned HTTP 403（密钥无权限，请检查 API Key）"
		}
	case http.StatusPaymentRequired:
		err.Message = "YFAI Seedance returned HTTP 402（账户余额不足，请充值）"
	case http.StatusTooManyRequests:
		err.RetryAfter = parseYFAIRetryAfter(retryAfterHeader)
		err.Message = fmt.Sprintf("YFAI Seedance returned HTTP 429（接口限流，请 %.0f 秒后重试）", err.RetryAfter.Seconds())
	default:
		err.Message = fmt.Sprintf("YFAI Seedance returned HTTP %d", statusCode)
	}
	return err
}

// parseYFAIRetryAfter accepts both the delta-seconds and HTTP-date forms of
// Retry-After and falls back to a conservative default so an upstream 429 can
// still be backed off when the header is missing or malformed.
func parseYFAIRetryAfter(header string) time.Duration {
	value := strings.TrimSpace(header)
	if value == "" {
		return yfaiDefaultRetryAfter
	}
	if seconds, err := strconv.Atoi(value); err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}
	if at, err := http.ParseTime(value); err == nil {
		if wait := time.Until(at); wait > 0 {
			return wait
		}
		return yfaiDefaultRetryAfter
	}
	return yfaiDefaultRetryAfter
}

func firstVideoBool(raw []byte, keys ...string) bool {
	var payload any
	if json.Unmarshal(raw, &payload) != nil {
		return false
	}
	for _, object := range videoObjects(payload) {
		for _, key := range keys {
			switch typed := object[key].(type) {
			case bool:
				return typed
			case string:
				switch strings.ToLower(strings.TrimSpace(typed)) {
				case "true", "1", "yes":
					return true
				}
			case float64:
				return typed != 0
			}
		}
	}
	return false
}
