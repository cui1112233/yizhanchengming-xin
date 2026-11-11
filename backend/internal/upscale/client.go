package upscale

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"qiantie/backend/internal/safehttp"
)

// 接口依据火山引擎 AI MediaKit 官方文档：
//   - 提交画质增强（大模型版）任务：POST /api/v1/tools/enhance-video-generative
//   - 查询任务信息：GET /api/v1/tasks/{task_id}
//   - 取消异步任务：POST /api/v1/tasks/terminate
//   - 获取媒体上传地址：POST /api/v1/tools-sync/request-media-upload-url
//
// 鉴权统一为 Authorization: Bearer {API Key}。
const (
	DefaultBaseURL = "https://mediakit.cn-beijing.volces.com"

	enhancePath     = "/api/v1/tools/enhance-video-generative"
	tasksPath       = "/api/v1/tasks/"
	terminatePath   = "/api/v1/tasks/terminate"
	uploadURLPath   = "/api/v1/tools-sync/request-media-upload-url"
	maxResponseSize = 1 << 20
)

// ErrNotConfigured 表示服务端没有配置火山 AI MediaKit 凭证。
var ErrNotConfigured = errors.New("未配置火山凭证")

// ErrInvalidResolution 表示目标分辨率不在画质增强（大模型版）支持范围内。
var ErrInvalidResolution = errors.New("仅支持 1080p 或 2k")

// APIError 是火山返回的业务错误或 HTTP 错误。它不包含任何请求头或凭证内容。
type APIError struct {
	Status  int
	Code    string
	Message string
}

func (e *APIError) Error() string {
	message := strings.TrimSpace(e.Message)
	if message == "" {
		message = "火山 AI MediaKit 请求失败"
	}
	if e.Code != "" {
		message += "（" + e.Code + "）"
	}
	if e.Status >= http.StatusBadRequest {
		return fmt.Sprintf("%s: HTTP %d", message, e.Status)
	}
	return message
}

// Client 是火山 AI MediaKit 的最小 HTTP 客户端，只做请求构造与响应解析。
type Client struct {
	BaseURL    string
	APIKey     string
	HTTP       *http.Client
	UploadHTTP *http.Client
}

func NewClient(baseURL, apiKey string) *Client {
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if base == "" {
		base = DefaultBaseURL
	}
	return &Client{
		BaseURL: base,
		APIKey:  strings.TrimSpace(apiKey),
		// JSON 接口要快速失败；上传请求可能持续很久，不设整体超时，由调用方 context 控制。
		HTTP:       safehttp.NewClientWithTimeout(90 * time.Second),
		UploadHTTP: &http.Client{Transport: safehttp.NewTransport(safehttp.Options{})},
	}
}

func (c *Client) Configured() bool {
	return c != nil && strings.TrimSpace(c.APIKey) != ""
}

func (c *Client) httpClient() *http.Client {
	if c != nil && c.HTTP != nil {
		return c.HTTP
	}
	return safehttp.NewClientWithTimeout(90 * time.Second)
}

func (c *Client) uploadClient() *http.Client {
	if c != nil && c.UploadHTTP != nil {
		return c.UploadHTTP
	}
	return c.httpClient()
}

type errorPayload struct {
	Code    string `json:"code"`
	Type    string `json:"type"`
	Message string `json:"message"`
	Param   string `json:"param"`
}

// UploadTarget 是 request-media-upload-url 返回的直传凭证。
type UploadTarget struct {
	FileID    string
	UploadURL string
	Method    string
	Headers   map[string]string
}

// Task 是查询任务接口里本项目关心的字段。
type Task struct {
	TaskID       string
	Status       string
	VideoURL     string
	ErrorCode    string
	ErrorMessage string
}

// TerminateOutcome 是取消异步任务的批量结果。
type TerminateOutcome struct {
	Terminated []string
	Failed     map[string]string
}

func (c *Client) postJSON(ctx context.Context, path string, payload any) (map[string]any, error) {
	if !c.Configured() {
		return nil, ErrNotConfigured
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("构造火山请求失败：%w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("构造火山请求失败：%w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	return c.execute(req)
}

func (c *Client) execute(req *http.Request) (map[string]any, error) {
	resp, err := c.httpClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("火山 AI MediaKit 请求失败：%w", err)
	}
	defer resp.Body.Close()
	raw, readErr := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if readErr != nil {
		return nil, fmt.Errorf("读取火山响应失败：%w", readErr)
	}
	var payload map[string]any
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &payload)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, apiErrorFromResponse(resp.StatusCode, payload, raw)
	}
	if payload == nil {
		return nil, &APIError{Status: resp.StatusCode, Message: "火山响应不是有效 JSON"}
	}
	if success, ok := payload["success"].(bool); ok && !success {
		return nil, apiErrorFromResponse(resp.StatusCode, payload, raw)
	}
	return payload, nil
}

func apiErrorFromResponse(status int, payload map[string]any, raw []byte) *APIError {
	apiErr := &APIError{Status: status}
	if detail, ok := payload["error"].(map[string]any); ok {
		apiErr.Code = stringValue(detail["code"])
		apiErr.Message = stringValue(detail["message"])
	}
	if apiErr.Message == "" {
		apiErr.Message = truncateText(string(raw), 300)
	}
	if apiErr.Message == "" {
		apiErr.Message = "火山 AI MediaKit 请求失败"
	}
	return apiErr
}

// RequestUploadURL 申请本地文件直传地址。返回的 file_id 可直接作为 video_url 使用。
func (c *Client) RequestUploadURL(ctx context.Context) (UploadTarget, error) {
	payload, err := c.postJSON(ctx, uploadURLPath, map[string]any{})
	if err != nil {
		return UploadTarget{}, err
	}
	result, _ := payload["result"].(map[string]any)
	target := UploadTarget{
		FileID:    stringValue(result["file_id"]),
		UploadURL: stringValue(result["upload_url"]),
		Method:    strings.ToUpper(stringValue(result["method"])),
		Headers:   map[string]string{},
	}
	if items, ok := result["upload_headers"].([]any); ok {
		for _, item := range items {
			entry, ok := item.(map[string]any)
			if !ok {
				continue
			}
			key := stringValue(entry["key"])
			if key == "" {
				continue
			}
			target.Headers[key] = stringValue(entry["value"])
		}
	}
	if target.FileID == "" || target.UploadURL == "" {
		return UploadTarget{}, &APIError{Status: 502, Message: "火山未返回上传地址或 file_id"}
	}
	if target.Method == "" {
		target.Method = http.MethodPut
	}
	return target, nil
}

// UploadMedia 把文件二进制流 PUT 到预签名地址。必须使用原始二进制，不能用 multipart。
func (c *Client) UploadMedia(ctx context.Context, target UploadTarget, body io.Reader, size int64, contentType string) error {
	method := strings.ToUpper(strings.TrimSpace(target.Method))
	if method == "" {
		method = http.MethodPut
	}
	req, err := http.NewRequestWithContext(ctx, method, target.UploadURL, body)
	if err != nil {
		return fmt.Errorf("构造火山上传请求失败：%w", err)
	}
	if size > 0 {
		req.ContentLength = size
	}
	for key, value := range target.Headers {
		req.Header.Set(key, value)
	}
	if contentType != "" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", contentType)
	}
	resp, err := c.uploadClient().Do(req)
	if err != nil {
		return fmt.Errorf("上传视频到火山失败：%w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		message := truncateText(string(raw), 300)
		if message == "" {
			message = "上传视频到火山失败"
		}
		return &APIError{Status: resp.StatusCode, Message: message}
	}
	return nil
}

// SubmitEnhance 提交画质增强（大模型版）任务并返回 task_id。
func (c *Client) SubmitEnhance(ctx context.Context, videoURL, resolution string) (string, error) {
	payload, err := c.postJSON(ctx, enhancePath, map[string]any{
		"video_url":  strings.TrimSpace(videoURL),
		"resolution": strings.TrimSpace(resolution),
	})
	if err != nil {
		return "", err
	}
	taskID := stringValue(payload["task_id"])
	if taskID == "" {
		return "", &APIError{Status: 502, Message: "火山未返回 task_id"}
	}
	return taskID, nil
}

// QueryTask 查询异步任务状态与结果。火山不返回百分比进度，只有 running/completed/failed。
func (c *Client) QueryTask(ctx context.Context, taskID string) (Task, error) {
	if !c.Configured() {
		return Task{}, ErrNotConfigured
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+tasksPath+url.PathEscape(strings.TrimSpace(taskID)), nil)
	if err != nil {
		return Task{}, fmt.Errorf("构造火山查询请求失败：%w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Accept", "application/json")
	payload, err := c.execute(req)
	if err != nil {
		return Task{}, err
	}
	task := Task{TaskID: taskID, Status: stringValue(payload["status"])}
	if result, ok := payload["result"].(map[string]any); ok {
		task.VideoURL = stringValue(result["video_url"])
	}
	if detail, ok := payload["error"].(map[string]any); ok {
		task.ErrorCode = stringValue(detail["code"])
		task.ErrorMessage = stringValue(detail["message"])
	}
	return task, nil
}

// Terminate 批量取消异步任务。火山只允许取消排队中或准备阶段的任务。
func (c *Client) Terminate(ctx context.Context, taskIDs []string) (TerminateOutcome, error) {
	outcome := TerminateOutcome{Failed: map[string]string{}}
	if len(taskIDs) == 0 {
		return outcome, nil
	}
	payload, err := c.postJSON(ctx, terminatePath, map[string]any{"task_ids": taskIDs})
	if err != nil {
		return outcome, err
	}
	result, _ := payload["result"].(map[string]any)
	if items, ok := result["terminated_tasks"].([]any); ok {
		for _, item := range items {
			entry, ok := item.(map[string]any)
			if !ok {
				continue
			}
			if id := stringValue(entry["task_id"]); id != "" {
				outcome.Terminated = append(outcome.Terminated, id)
			}
		}
	}
	if items, ok := result["terminate_failed_tasks"].([]any); ok {
		for _, item := range items {
			entry, ok := item.(map[string]any)
			if !ok {
				continue
			}
			id := stringValue(entry["task_id"])
			if id == "" {
				continue
			}
			message := "火山拒绝取消"
			if detail, ok := entry["error"].(map[string]any); ok {
				if text := stringValue(detail["message"]); text != "" {
					message = text
				}
			}
			outcome.Failed[id] = message
		}
	}
	return outcome, nil
}

// NormalizeFileID 统一成 mediakit://{file_id}；接口已经带前缀时原样返回。
func NormalizeFileID(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	if strings.HasPrefix(value, "mediakit://") {
		return value
	}
	return "mediakit://" + value
}

func stringValue(value any) string {
	text, _ := value.(string)
	return strings.TrimSpace(text)
}

func truncateText(value string, limit int) string {
	text := strings.TrimSpace(value)
	if len(text) <= limit {
		return text
	}
	return text[:limit]
}
