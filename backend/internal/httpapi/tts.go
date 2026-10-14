package httpapi

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

	"qiantie/backend/internal/novelfetchworkshop"
)

type ttsRequest struct {
	Input string  `json:"input"`
	Voice string  `json:"voice,omitempty"`
	Speed float64 `json:"speed,omitempty"`
	Pitch any     `json:"pitch,omitempty"`
	Style string  `json:"style,omitempty"`
}

// defaultTTSEndpoint is kept as a variable so the shared fallback path can be
// tested without calling the real service. The UI already treats this service
// as the configured default when no catalog audio model is present.
var defaultTTSEndpoint = shuihuoDefaultTTSServiceURL

func registerTTSRoute(root *http.ServeMux, auth BridgeAuth, catalogStore novelfetchworkshop.Store) {
	root.Handle("/api/tts", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		var input ttsRequest
		if !decodeJSON(w, r, &input) {
			return
		}
		text := strings.TrimSpace(input.Input)
		if text == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "输入文本不能为空"})
			return
		}
		if input.Voice == "" {
			input.Voice = "zh-CN-XiaoxiaoNeural"
		}
		if input.Speed == 0 {
			input.Speed = 1.8
		}
		if input.Pitch == nil {
			input.Pitch = "10"
		}
		if input.Style == "" {
			input.Style = "general"
		}
		if identity, ok := BridgeIdentityFromContext(r.Context()); ok && catalogStore != nil {
			if model, found := configuredTTSModel(r.Context(), catalogStore, identity.Username); found {
				serveConfiguredTTS(w, r, input, model)
				return
			}
		}
		body, contentType, err := requestDefaultTTSBytes(r.Context(), input)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(body)))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	})))
}

func requestDefaultTTSBytes(ctx context.Context, input ttsRequest) ([]byte, string, error) {
	payload, err := json.Marshal(map[string]any{"input": input.Input, "voice": input.Voice, "speed": input.Speed, "pitch": input.Pitch, "style": input.Style})
	if err != nil {
		return nil, "", fmt.Errorf("TTS 请求组装失败")
	}
	requestCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(requestCtx, http.MethodPost, defaultTTSEndpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, "", fmt.Errorf("TTS 请求创建失败")
	}
	req.Header.Set("Host", "tts3.121w.com")
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Content-Type", "application/json")
	response, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("TTS 服务请求失败: %w", err)
	}
	defer response.Body.Close()
	body, readErr := io.ReadAll(io.LimitReader(response.Body, 32<<20))
	if readErr != nil || len(body) == 0 || response.StatusCode >= 400 {
		message := strings.TrimSpace(string(body))
		if len(message) > 500 {
			message = message[:500]
		}
		if message == "" {
			message = response.Status
		}
		return nil, "", fmt.Errorf("TTS 服务错误: %s", message)
	}
	if strings.Contains(strings.ToLower(response.Header.Get("Content-Type")), "application/json") {
		return nil, "", fmt.Errorf("TTS 服务返回错误: %s", string(body))
	}
	contentType := strings.TrimSpace(strings.Split(response.Header.Get("Content-Type"), ";")[0])
	if contentType == "" {
		contentType = "audio/mpeg"
	}
	return body, contentType, nil
}

func configuredTTSModel(ctx context.Context, store novelfetchworkshop.Store, owner string) (map[string]any, bool) {
	config, err := store.GetConfig(ctx, owner)
	if err != nil {
		return nil, false
	}
	for _, model := range modelCatalog(config) {
		if !strings.EqualFold(modelString(model, "kind"), "audio") || model["enabled"] != true {
			continue
		}
		if modelString(model, "endpoint") == "" && modelString(model, "baseUrl") == "" {
			continue
		}
		if shuihuoModelCredential(model) == "" || modelString(model, "requestTemplate") == "" {
			continue
		}
		return model, true
	}
	return nil, false
}

func serveConfiguredTTS(w http.ResponseWriter, r *http.Request, input ttsRequest, model map[string]any) {
	payload, contentType, err := requestConfiguredTTSBytes(r.Context(), input, model)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "音频模型请求失败: " + err.Error()})
		return
	}
	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(payload)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(payload)
}

func requestConfiguredTTSBytes(ctx context.Context, input ttsRequest, model map[string]any) ([]byte, string, error) {
	endpoint := modelString(model, "endpoint")
	if endpoint == "" {
		endpoint = modelString(model, "baseUrl")
	}
	method, headers, body, err := configuredTTSRequest(modelString(model, "requestTemplate"), input, shuihuoModelCredential(model))
	if err != nil {
		return nil, "", fmt.Errorf("请求模板无效: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, "", fmt.Errorf("音频模型地址无效")
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if req.Header.Get("Authorization") == "" {
		req.Header.Set("Authorization", "Bearer "+shuihuoModelCredential(model))
	}
	response, err := (&http.Client{Timeout: 60 * time.Second}).Do(req)
	if err != nil {
		return nil, "", err
	}
	defer response.Body.Close()
	payload, readErr := io.ReadAll(io.LimitReader(response.Body, 32<<20))
	if readErr != nil || len(payload) == 0 || response.StatusCode >= 400 {
		return nil, "", fmt.Errorf("音频模型返回状态 %d", response.StatusCode)
	}
	if strings.Contains(strings.ToLower(response.Header.Get("Content-Type")), "application/json") {
		audioURL, resolveErr := configuredTTSAudioURL(payload, modelString(model, "responseMapping"))
		if resolveErr != nil {
			return nil, "", fmt.Errorf("音频模型返回了 JSON，而不是音频文件: %w", resolveErr)
		}
		return fetchConfiguredTTSAudio(ctx, audioURL)
	}
	contentType := strings.TrimSpace(strings.Split(response.Header.Get("Content-Type"), ";")[0])
	if contentType == "" {
		contentType = "audio/mpeg"
	}
	return payload, contentType, nil
}

func configuredTTSAudioURL(payload []byte, responseMapping string) (string, error) {
	var mapping map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(responseMapping)), &mapping); err != nil {
		return "", fmt.Errorf("返回映射无效")
	}
	if len(mapping) == 0 {
		return "", fmt.Errorf("未配置音频 URL 返回映射")
	}
	var response any
	if err := json.Unmarshal(payload, &response); err != nil {
		return "", fmt.Errorf("返回内容不是有效 JSON")
	}
	for _, rawPath := range mapping {
		path, ok := rawPath.(string)
		if !ok || strings.TrimSpace(path) == "" {
			continue
		}
		value, ok := jsonPathValue(response, path)
		if !ok {
			continue
		}
		candidate, ok := value.(string)
		if !ok {
			continue
		}
		parsed, err := url.Parse(strings.TrimSpace(candidate))
		if err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != "" {
			return parsed.String(), nil
		}
	}
	return "", fmt.Errorf("未找到有效的音频 URL")
}

func jsonPathValue(root any, path string) (any, bool) {
	current := root
	for _, part := range strings.Split(strings.Trim(path, " ."), ".") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		object, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = object[part]
		if !ok {
			return nil, false
		}
	}
	return current, true
}

func fetchConfiguredTTSAudio(ctx context.Context, audioURL string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, audioURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("音频 URL 无效")
	}
	response, err := (&http.Client{Timeout: 60 * time.Second}).Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("音频 URL 请求失败: %w", err)
	}
	defer response.Body.Close()
	payload, readErr := io.ReadAll(io.LimitReader(response.Body, 32<<20))
	if readErr != nil || len(payload) == 0 || response.StatusCode >= 400 {
		return nil, "", fmt.Errorf("音频 URL 返回状态 %d", response.StatusCode)
	}
	contentType := strings.TrimSpace(strings.Split(response.Header.Get("Content-Type"), ";")[0])
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = "audio/mpeg"
	}
	return payload, contentType, nil
}

func configuredTTSRequest(template string, input ttsRequest, credential string) (string, map[string]string, []byte, error) {
	replacements := map[string]string{
		"{{input}}": input.Input, "{{prompt}}": input.Input, "{{voice}}": input.Voice,
		"{{speech_rate}}": fmt.Sprint(input.Speed), "{{speed}}": fmt.Sprint(input.Speed),
		"{{pitch}}": fmt.Sprint(input.Pitch), "{{style}}": input.Style, "{{credential}}": credential,
	}
	raw := template
	for from, to := range replacements {
		raw = strings.ReplaceAll(raw, from, to)
	}
	var envelope map[string]any
	if err := json.Unmarshal([]byte(raw), &envelope); err != nil {
		return "", nil, nil, err
	}
	method := strings.ToUpper(strings.TrimSpace(fmt.Sprint(envelope["method"])))
	if method == "" || method == "<nil>" {
		method = http.MethodPost
	}
	headers := map[string]string{}
	if rawHeaders, ok := envelope["headers"].(map[string]any); ok {
		for key, value := range rawHeaders {
			headers[key] = fmt.Sprint(value)
		}
	}
	bodyValue, ok := envelope["body"]
	if !ok {
		bodyValue = envelope
	}
	body, err := json.Marshal(bodyValue)
	return method, headers, body, err
}
