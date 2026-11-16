package httpapi

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

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/safehttp"
)

type nativeScriptVideoAdapter struct {
	cfg    batchfactoryv11.VideoProviderConfig
	Client *http.Client
}

func (a *nativeScriptVideoAdapter) Validate() error {
	if !isNativeScriptVideoProvider(a.cfg.Provider) || strings.TrimSpace(a.cfg.APIKey) == "" || strings.TrimSpace(a.cfg.Model) == "" {
		return fmt.Errorf("原生视频配置缺少协议、模型或密钥")
	}
	tasksURL, err := a.tasksEndpoint()
	if err != nil {
		return err
	}
	for _, raw := range []string{a.cfg.CreateURL, tasksURL} {
		u, err := url.Parse(raw)
		if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.Fragment != "" {
			return fmt.Errorf("原生视频需要配置HTTPS创建和查询地址")
		}
		if err := validateNativeMedia(raw); err != nil {
			return fmt.Errorf("原生视频接口必须使用公网HTTPS地址")
		}
	}
	return nil
}
func (a *nativeScriptVideoAdapter) Submit(ctx context.Context, model batchfactoryv11.FrozenVideoModel, p batchfactoryv11.FinalPrompt) (batchfactoryv11.ProviderTaskRef, error) {
	if err := a.Validate(); err != nil {
		return batchfactoryv11.ProviderTaskRef{}, err
	}
	if model.ID != a.cfg.Model {
		return batchfactoryv11.ProviderTaskRef{}, fmt.Errorf("视频模型与配置不符")
	}
	if p.DurationSeconds < 1 || p.DurationSeconds > 15 {
		return batchfactoryv11.ProviderTaskRef{}, fmt.Errorf("视频时长需在1至15秒内")
	}
	ratio := batchfactoryv11.EffectiveVideoAspectRatio(p.EffectiveSettings.Values)
	resolution := batchfactoryv11.EffectiveVideoResolution(p.EffectiveSettings.Values)
	if a.cfg.Provider == scriptVideoProviderSora && len(p.SubjectReferenceURIs) > 0 {
		return batchfactoryv11.ProviderTaskRef{}, fmt.Errorf("Sora协议不能携带主体图片引用")
	}
	unique := map[string]bool{}
	for _, raw := range append(append([]string{}, p.SubjectReferenceURIs...), p.ReferenceImageURLs...) {
		unique[raw] = true
	}
	if len(unique) > 9 {
		return batchfactoryv11.ProviderTaskRef{}, fmt.Errorf("参考图片数量不能超过9张")
	}
	var payload map[string]any
	if a.cfg.Provider == scriptVideoProviderArk {
		content := []any{map[string]any{"type": "text", "text": p.CompiledPrompt}}
		seen := map[string]bool{}
		for _, raw := range append(append([]string{}, p.SubjectReferenceURIs...), p.ReferenceImageURLs...) {
			if err := validateNativeMedia(raw); err != nil {
				return batchfactoryv11.ProviderTaskRef{}, err
			}
			if !seen[raw] {
				content = append(content, map[string]any{"type": "image_url", "image_url": map[string]string{"url": raw}, "role": "reference_image"})
				seen[raw] = true
			}
		}
		payload = map[string]any{"model": a.cfg.Model, "content": content, "ratio": ratio, "duration": p.DurationSeconds, "resolution": resolution, "watermark": false}
	} else {
		images := append([]string{}, p.ReferenceImageURLs...)
		orientation := "landscape"
		if ratio == "9:16" {
			orientation = "portrait"
		}
		payload = map[string]any{"images": images, "model": a.cfg.Model, "orientation": orientation, "prompt": p.CompiledPrompt, "size": "small", "duration": p.DurationSeconds, "watermark": false}
	}
	reply, err := a.do(ctx, http.MethodPost, a.cfg.CreateURL, payload)
	if err != nil {
		return batchfactoryv11.ProviderTaskRef{}, err
	}
	id := nativeString(reply, "id")

	if id == "" || len(id) > nativeScriptTaskIDMaxBytes {
		return batchfactoryv11.ProviderTaskRef{}, fmt.Errorf("视频响应未返回可用任务ID，提交状态未知，请核查后再操作")
	}
	return batchfactoryv11.ProviderTaskRef{ProviderTaskID: id, State: batchfactoryv11.ProductionQueued}, nil
}
func nativeString(m map[string]any, key string) string {
	v, _ := m[key].(string)
	return strings.TrimSpace(v)
}
func (a *nativeScriptVideoAdapter) do(ctx context.Context, method, endpoint string, payload any) (map[string]any, error) {
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
		return nil, fmt.Errorf("视频请求地址无效")
	}
	req.Header.Set("Authorization", "Bearer "+a.cfg.APIKey)
	req.Header.Set("Accept", "application/json")
	if payload != nil || a.cfg.Provider == scriptVideoProviderSora {
		req.Header.Set("Content-Type", "application/json")
	}
	var transport http.RoundTripper
	if a.Client != nil {
		transport = a.Client.Transport
	}
	timeout := 120 * time.Second
	if method == http.MethodGet {
		timeout = 30 * time.Second
	} else if a.cfg.Provider == scriptVideoProviderSora {
		timeout = 600 * time.Second
	}
	client := safehttp.NewClient(safehttp.Options{Timeout: timeout, Transport: transport})
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return fmt.Errorf("视频接口不允许重定向") }
	resp, err := client.Do(req)
	if err != nil {
		if errors.Is(err, safehttp.ErrBlockedAddress) || errors.Is(err, safehttp.ErrBlockedScheme) || errors.Is(err, safehttp.ErrResolveFailed) {
			return nil, &scriptVideoPreflightError{Err: fmt.Errorf("视频接口地址在发送前被安全检查拒绝，请检查公网地址")}
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("视频接口请求未确认，请核查任务后再操作")
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, (1<<20)+1))
	if err != nil || len(data) > 1<<20 {
		return nil, fmt.Errorf("视频接口响应过大或读取失败")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("视频接口返回HTTP %d", resp.StatusCode)
	}
	var reply map[string]any
	if json.Unmarshal(data, &reply) != nil {
		return nil, fmt.Errorf("视频接口响应不是有效JSON")
	}
	return reply, nil
}
func (a *nativeScriptVideoAdapter) tasksEndpoint() (string, error) {
	if strings.TrimSpace(a.cfg.TasksURL) != "" {
		return a.cfg.TasksURL, nil
	}
	u, err := url.Parse(a.cfg.CreateURL)
	if err != nil {
		return "", fmt.Errorf("视频创建地址无效")
	}
	if a.cfg.Provider == scriptVideoProviderSora && strings.HasSuffix(u.Path, "/v1/video/create") {
		u.Path = strings.TrimSuffix(u.Path, "/create") + "/query"
		u.RawPath = ""
		return u.String(), nil
	}
	if a.cfg.Provider == scriptVideoProviderArk && strings.HasSuffix(strings.TrimRight(u.Path, "/"), "/contents/generations/tasks") {
		u.Path = strings.TrimRight(u.Path, "/") + "/{id}"
		u.RawPath = ""
		return strings.ReplaceAll(u.String(), "%7Bid%7D", "{id}"), nil
	}
	return "", fmt.Errorf("请为此视频地址配置查询地址")
}
func nativeScriptResponseFields(provider string, reply map[string]any) (status, video, message string) {
	status = nativeString(reply, "status")
	video = nativeString(reply, "video_url")
	if errObject, ok := reply["error"].(map[string]any); ok {
		message = nativeString(errObject, "message")
	}
	if provider == scriptVideoProviderArk {
		if content, ok := reply["content"].(map[string]any); ok {
			if value := nativeString(content, "video_url"); value != "" {
				video = value
			}
		}
		if data, ok := reply["data"].(map[string]any); ok {
			if status == "" {
				status = nativeString(data, "status")
			}
			if video == "" {
				video = nativeString(data, "video_url")
				if video == "" {
					video = nativeString(data, "url")
				}
			}
		}
	}
	return
}
func (a *nativeScriptVideoAdapter) Poll(ctx context.Context, _ batchfactoryv11.FrozenVideoModel, task batchfactoryv11.ProviderTaskRef) (batchfactoryv11.ProviderTaskRef, error) {
	if err := ctx.Err(); err != nil {
		return batchfactoryv11.ProviderTaskRef{}, err
	}
	if err := a.Validate(); err != nil {
		return batchfactoryv11.ProviderTaskRef{}, err
	}
	if strings.TrimSpace(task.ProviderTaskID) == "" || len(task.ProviderTaskID) > nativeScriptTaskIDMaxBytes {
		return batchfactoryv11.ProviderTaskRef{}, fmt.Errorf("视频任务ID无效")
	}
	endpoint, _ := a.tasksEndpoint()
	hasPlaceholder := strings.Contains(endpoint, "{id}")
	endpoint = strings.ReplaceAll(endpoint, "{id}", url.PathEscape(task.ProviderTaskID))
	u, err := url.Parse(endpoint)
	if err != nil {
		return batchfactoryv11.ProviderTaskRef{}, fmt.Errorf("视频查询地址无效")
	}
	if a.cfg.Provider == scriptVideoProviderSora {
		query := u.Query()
		query.Set("id", task.ProviderTaskID)
		u.RawQuery = query.Encode()
	} else if !hasPlaceholder {
		basePath := strings.TrimRight(u.Path, "/")
		escapedBase := strings.TrimRight(u.EscapedPath(), "/")
		u.Path = basePath + "/" + task.ProviderTaskID
		u.RawPath = escapedBase + "/" + url.PathEscape(task.ProviderTaskID)
	}
	reply, err := a.do(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return batchfactoryv11.ProviderTaskRef{}, err
	}
	status, video, message := nativeScriptResponseFields(a.cfg.Provider, reply)
	out := batchfactoryv11.ProviderTaskRef{ProviderTaskID: task.ProviderTaskID, State: batchfactoryv11.ProductionRunning}
	switch strings.ToLower(status) {
	case "failed", "error", "cancelled", "canceled":
		out.State = batchfactoryv11.ProductionFailed
		message = strings.ReplaceAll(message, a.cfg.APIKey, "[已隐藏]")
		message = strings.Map(func(r rune) rune {
			if r < ' ' {
				return ' '
			}
			return r
		}, message)
		runes := []rune(message)
		if len(runes) > 512 {
			message = string(runes[:512])
		}
		if message == "" {
			message = "视频生成失败或已取消"
		}
		out.ErrorMessage = message
	case "completed", "succeeded":
		if video != "" {
			if strings.HasPrefix(video, "asset://") {
				return batchfactoryv11.ProviderTaskRef{}, fmt.Errorf("视频结果地址无效")
			}
			if err = validateNativeMedia(video); err != nil {
				return batchfactoryv11.ProviderTaskRef{}, fmt.Errorf("视频结果地址无效")
			}
			out.State = batchfactoryv11.ProductionSucceeded
			out.MediaURL = video
		}
	}
	return out, nil
}
