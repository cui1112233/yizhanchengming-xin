package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
		payload, err := json.Marshal(map[string]any{"input": text, "voice": input.Voice, "speed": input.Speed, "pitch": input.Pitch, "style": input.Style})
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "TTS 请求组装失败"})
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://tts3.121w.com/v1/audio/speech", bytes.NewReader(payload))
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "TTS 请求创建失败"})
			return
		}
		req.Header.Set("Host", "tts3.121w.com")
		req.Header.Set("User-Agent", "Mozilla/5.0")
		req.Header.Set("Accept", "*/*")
		req.Header.Set("Content-Type", "application/json")
		response, err := (&http.Client{}).Do(req)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "TTS 服务请求失败: " + err.Error()})
			return
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
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": fmt.Sprintf("TTS 服务错误: %s", message)})
			return
		}
		if strings.Contains(strings.ToLower(response.Header.Get("Content-Type")), "application/json") {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "TTS 服务返回错误: " + string(body)})
			return
		}
		if contentType := response.Header.Get("Content-Type"); contentType != "" {
			w.Header().Set("Content-Type", contentType)
		}
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(body)))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	})))
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
	endpoint := modelString(model, "endpoint")
	if endpoint == "" {
		endpoint = modelString(model, "baseUrl")
	}
	method, headers, body, err := configuredTTSRequest(modelString(model, "requestTemplate"), input, shuihuoModelCredential(model))
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "音频模型请求模板无效: " + err.Error()})
		return
	}
	req, err := http.NewRequestWithContext(r.Context(), method, endpoint, bytes.NewReader(body))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "音频模型地址无效"})
		return
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
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "音频模型请求失败: " + err.Error()})
		return
	}
	defer response.Body.Close()
	payload, readErr := io.ReadAll(io.LimitReader(response.Body, 32<<20))
	if readErr != nil || len(payload) == 0 || response.StatusCode >= 400 {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": fmt.Sprintf("音频模型返回状态 %d", response.StatusCode)})
		return
	}
	if strings.Contains(strings.ToLower(response.Header.Get("Content-Type")), "application/json") {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "音频模型返回了 JSON，而不是音频文件"})
		return
	}
	if contentType := response.Header.Get("Content-Type"); contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(payload)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(payload)
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
