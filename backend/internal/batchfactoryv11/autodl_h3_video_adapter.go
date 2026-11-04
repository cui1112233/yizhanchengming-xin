package batchfactoryv11

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	AutoDLH3Model              = "minimax-h3-video"
	AutoDLH3NoImageWorkflow    = "minimax_h3_lightx2v_no_pic"
	AutoDLH3ReferenceWorkflow  = "minimax_h3_lightx2v_v5_15s"
	DefaultAutoDLH3CreateURL   = "https://autodl.art/api/v1/comfyui/comfyui_workflow/{workflow}"
	DefaultAutoDLH3TasksURL    = "https://autodl.art/api/v1/comfyui/comfyui_workflow/result/{id}"
	MaxAutoDLH3ReferenceImages = 9
	// The reference workflow requires a ref_image_0 entry. When a retry loses
	// its reference images this transparent 1×1 PNG keeps the workflow/payload
	// consistent. It is served by this application, never by an external CDN.
	AutoDLH3FallbackImagePath = "/assets/empty-1x1.png"
	// QIANTIE_AUTODL_H3_FALLBACK_REFERENCE_IMAGE overrides the absolute URL.
	AutoDLH3FallbackImageEnv = "QIANTIE_AUTODL_H3_FALLBACK_REFERENCE_IMAGE"
)

// AutoDLH3FallbackImageURL resolves the absolute URL sent to the provider. The
// default points at this application's public origin because AutoDL fetches the
// image from its own network.
func AutoDLH3FallbackImageURL() string {
	if value := strings.TrimSpace(os.Getenv(AutoDLH3FallbackImageEnv)); value != "" {
		return value
	}
	return "https://yzcm.site" + AutoDLH3FallbackImagePath
}

// AutoDLH3VideoAdapter implements the two-step ComfyUI workflow API. The
// provider token stays server-side and is deliberately sent as the raw
// Authorization value, matching AutoDL.Art's API contract.
type AutoDLH3VideoAdapter struct {
	CreateURL   string
	TasksURL    string
	APIKey      string
	Model       string
	WorkflowID  string
	Client      *http.Client
	ValidateURL func(string) (*url.URL, error)
	// FallbackReferenceImageURL is injected only when the chosen workflow is a
	// reference workflow but the task carries no reference images.
	FallbackReferenceImageURL string
}

var _ ProductionAdapter = (*AutoDLH3VideoAdapter)(nil)
var _ ProductionPoller = (*AutoDLH3VideoAdapter)(nil)

func (a *AutoDLH3VideoAdapter) Validate() error {
	if _, err := a.validateTemplate(a.CreateURL, "{workflow}"); err != nil {
		return fmt.Errorf("AutoDL H3 create endpoint: %w", err)
	}
	if _, err := a.validateTemplate(a.TasksURL, "{id}"); err != nil {
		return fmt.Errorf("AutoDL H3 result endpoint: %w", err)
	}
	if strings.TrimSpace(a.APIKey) == "" {
		return fmt.Errorf("AutoDL H3 API key is required")
	}
	if strings.TrimSpace(a.Model) == "" {
		return fmt.Errorf("AutoDL H3 model is required")
	}
	return nil
}

func (a *AutoDLH3VideoAdapter) validatedURL(raw string) (*url.URL, error) {
	if a.ValidateURL != nil {
		return a.ValidateURL(raw)
	}
	return validateProductionURL(raw)
}

func (a *AutoDLH3VideoAdapter) validateTemplate(raw, placeholder string) (*url.URL, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("endpoint is required")
	}
	if !strings.Contains(raw, placeholder) {
		return nil, fmt.Errorf("endpoint must contain %s", placeholder)
	}
	testValue := strings.Replace(raw, placeholder, "placeholder", 1)
	u, err := a.validatedURL(testValue)
	if err != nil {
		return nil, err
	}
	if a.ValidateURL == nil && strings.ToLower(u.Hostname()) != "autodl.art" {
		return nil, fmt.Errorf("endpoint host must be autodl.art")
	}
	return u, nil
}

func (a *AutoDLH3VideoAdapter) Submit(ctx context.Context, model FrozenVideoModel, prompt FinalPrompt) (ProviderTaskRef, error) {
	if err := a.Validate(); err != nil {
		return ProviderTaskRef{}, err
	}
	if strings.TrimSpace(model.ID) == "" || model.ID != a.Model {
		return ProviderTaskRef{}, fmt.Errorf("frozen video model does not match AutoDL H3 model")
	}
	values := prompt.EffectiveSettings.Values
	duration := prompt.DurationSeconds
	if duration <= 0 {
		duration = rawInt(values, "duration", 0)
	}
	if duration <= 0 {
		duration = model.MaxDuration
	}
	if duration <= 0 {
		duration = 5
	}
	if model.MaxDuration > 0 && duration > model.MaxDuration {
		return ProviderTaskRef{}, fmt.Errorf("AutoDL H3 duration exceeds model maximum")
	}
	resolution := autoDLH3Resolution(rawString(values, "videoResolution", rawString(values, "resolution", "768p竖")))
	images := append([]string(nil), prompt.ReferenceImageURLs...)
	if len(images) == 0 {
		images = rawStrings(values, "imageUrls")
	}
	if len(images) == 0 {
		images = rawStrings(values, "referenceImages")
	}
	if len(images) > MaxAutoDLH3ReferenceImages {
		return ProviderTaskRef{}, fmt.Errorf("AutoDL H3 supports at most %d reference images", MaxAutoDLH3ReferenceImages)
	}
	for index, image := range images {
		image = strings.TrimSpace(image)
		if image == "" {
			return ProviderTaskRef{}, fmt.Errorf("AutoDL H3 reference image %d is empty", index)
		}
		parsed, err := a.validatedURL(image)
		if err != nil {
			return ProviderTaskRef{}, fmt.Errorf("invalid AutoDL H3 reference image: %w", err)
		}
		images[index] = parsed.String()
	}
	workflow := strings.TrimSpace(a.WorkflowID)
	if workflow == "" {
		workflow = AutoDLH3NoImageWorkflow
		if len(images) > 0 {
			workflow = AutoDLH3ReferenceWorkflow
		}
	}
	// The provider rejects a submission whose payload carries no ref_image_0,
	// even when the named no-image workflow is called (live evidence:
	// "缺少必填参数：ref_image_0"). A book with zero reference images therefore
	// still submits one self-hosted transparent PNG, and the workflow moves to
	// the reference variant so the payload and workflow agree.
	if len(images) == 0 {
		fallback := strings.TrimSpace(a.FallbackReferenceImageURL)
		if fallback == "" {
			fallback = AutoDLH3FallbackImageURL()
		}
		parsed, err := a.validatedURL(fallback)
		if err != nil {
			return ProviderTaskRef{}, fmt.Errorf("invalid AutoDL H3 fallback reference image: %w", err)
		}
		images = []string{parsed.String()}
		if workflow == AutoDLH3NoImageWorkflow {
			workflow = AutoDLH3ReferenceWorkflow
		}
	}
	if !validAutoDLH3WorkflowID(workflow) {
		return ProviderTaskRef{}, fmt.Errorf("AutoDL H3 workflow id is invalid")
	}
	payload := map[string]any{
		"prompt":     prompt.CompiledPrompt,
		"duration":   duration,
		"resolution": resolution,
	}
	for index, image := range images {
		payload[fmt.Sprintf("ref_image_%d", index)] = image
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return ProviderTaskRef{}, err
	}
	endpoint, err := a.validatedURL(strings.ReplaceAll(a.CreateURL, "{workflow}", workflow))
	if err != nil {
		return ProviderTaskRef{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), bytes.NewReader(body))
	if err != nil {
		return ProviderTaskRef{}, err
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", strings.TrimSpace(a.APIKey))
	reply, err := a.doJSON(request)
	if err != nil {
		return ProviderTaskRef{}, err
	}
	taskID := firstVideoValue(reply, "task_id", "taskId", "id")
	if taskID == "" {
		if rejection := autoDLH3ProviderRejection(reply, a.APIKey); rejection != "" {
			return ProviderTaskRef{}, fmt.Errorf("AutoDL H3 rejected submission: %s", rejection)
		}
		return ProviderTaskRef{}, fmt.Errorf("AutoDL H3 response did not contain a task id")
	}
	return ProviderTaskRef{ProviderTaskID: taskID, State: autoDLH3State(firstVideoValue(reply, "status", "state")), RequestedDurationSeconds: float64(duration)}, nil
}

// autoDLH3ProviderRejection returns only bounded, explicitly named response
// fields.  It is intentionally not a raw-body dump: provider replies can be
// arbitrary and must never leak the configured credential into task history.
func autoDLH3ProviderRejection(reply []byte, apiKey string) string {
	parts := make([]string, 0, 3)
	if code := autoDLH3SafeDetail(firstVideoValue(reply, "code", "error_code"), apiKey); code != "" {
		parts = append(parts, "code="+code)
	}
	if message := autoDLH3SafeDetail(firstVideoValue(reply, "msg", "message", "error", "detail"), apiKey); message != "" {
		parts = append(parts, "message="+message)
	}
	if requestID := autoDLH3SafeDetail(firstVideoValue(reply, "request_id", "requestId", "trace_id", "traceId"), apiKey); requestID != "" {
		parts = append(parts, "request_id="+requestID)
	}
	return strings.Join(parts, " ")
}

func autoDLH3SafeDetail(value, apiKey string) string {
	value = strings.Join(strings.Fields(value), " ")
	if apiKey = strings.TrimSpace(apiKey); apiKey != "" {
		value = strings.ReplaceAll(value, apiKey, "[redacted]")
	}
	const maxLength = 240
	if len(value) > maxLength {
		value = value[:maxLength] + "…"
	}
	return value
}

// autoDLH3Resolution translates the editor's generic legacy labels into the
// exact values accepted by the H3 image/audio workflow.  In particular 720p
// is not an AutoDL option there; 768p is its supported equivalent.
func autoDLH3Resolution(value string) string {
	normalized := strings.ToLower(strings.ReplaceAll(strings.Join(strings.Fields(value), ""), "×", "x"))
	switch normalized {
	case "480p横", "480phorizontal", "480plandscape":
		return "480p横"
	case "768p横", "720p横", "720phorizontal", "720plandscape", "768phorizontal", "768plandscape":
		return "768p横"
	case "480p", "480p竖", "480pvertical", "480pportrait":
		return "480p竖"
	case "768p", "720p", "768p竖", "720p竖", "720pvertical", "720pportrait", "768pvertical", "768pportrait", "":
		return "768p竖"
	default:
		return "768p竖"
	}
}

func validAutoDLH3WorkflowID(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > 120 {
		return false
	}
	for _, character := range value {
		if !((character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') || (character >= '0' && character <= '9') || character == '_' || character == '-') {
			return false
		}
	}
	return true
}

func (a *AutoDLH3VideoAdapter) Poll(ctx context.Context, _ FrozenVideoModel, task ProviderTaskRef) (ProviderTaskRef, error) {
	if err := a.Validate(); err != nil {
		return ProviderTaskRef{}, err
	}
	taskID := strings.TrimSpace(task.ProviderTaskID)
	if taskID == "" {
		return ProviderTaskRef{}, fmt.Errorf("AutoDL H3 task id is required")
	}
	endpoint, err := a.validatedURL(strings.ReplaceAll(a.TasksURL, "{id}", url.PathEscape(taskID)))
	if err != nil {
		return ProviderTaskRef{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return ProviderTaskRef{}, err
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Authorization", strings.TrimSpace(a.APIKey))
	reply, err := a.doJSON(request)
	if err != nil {
		return ProviderTaskRef{}, err
	}
	state := autoDLH3State(firstVideoValue(reply, "status", "state"))
	result := ProviderTaskRef{ProviderTaskID: taskID, State: state, ActualDurationSeconds: firstVideoFloat(reply, "actualDurationSeconds", "actual_duration_seconds", "durationSeconds", "duration_seconds")}
	if state == ProductionSucceeded {
		mediaURL := autoDLH3MediaURL(reply)
		if mediaURL == "" {
			return ProviderTaskRef{}, fmt.Errorf("AutoDL H3 completed response did not contain a media URL")
		}
		if _, err := a.validatedURL(mediaURL); err != nil {
			return ProviderTaskRef{}, fmt.Errorf("AutoDL H3 returned unsafe media URL: %w", err)
		}
		result.MediaURL = mediaURL
	}
	return result, nil
}

func autoDLH3State(value string) ProductionState {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "SUCCESS", "SUCCEEDED", "COMPLETED", "DONE":
		return ProductionSucceeded
	case "FAILED", "ERROR", "CANCELLED", "CANCELED":
		return ProductionFailed
	case "QUEUED", "SUBMITTED", "PENDING":
		return ProductionQueued
	default:
		return ProductionRunning
	}
}

func autoDLH3MediaURL(raw []byte) string {
	var payload any
	if json.Unmarshal(raw, &payload) != nil {
		return ""
	}
	for _, object := range videoObjects(payload) {
		if results, ok := object["results"].([]any); ok {
			for _, result := range results {
				if value, ok := result.(string); ok && strings.TrimSpace(value) != "" {
					return strings.TrimSpace(value)
				}
				if resultObject, ok := result.(map[string]any); ok {
					for _, key := range []string{"url", "video_url", "videoUrl", "media_url", "mediaUrl"} {
						if value, ok := resultObject[key].(string); ok && strings.TrimSpace(value) != "" {
							return strings.TrimSpace(value)
						}
					}
				}
			}
		}
	}
	return firstVideoValue(raw, "url", "video_url", "videoUrl", "media_url", "mediaUrl")
}

func (a *AutoDLH3VideoAdapter) doJSON(request *http.Request) ([]byte, error) {
	client := a.Client
	if client == nil {
		client = &http.Client{Timeout: 120 * time.Second}
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("AutoDL H3 provider request failed")
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, 8<<20))
	if err != nil {
		return nil, fmt.Errorf("read AutoDL H3 provider response")
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("AutoDL H3 provider returned HTTP %d", response.StatusCode)
	}
	return body, nil
}
