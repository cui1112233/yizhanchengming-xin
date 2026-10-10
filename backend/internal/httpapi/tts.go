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
)

type ttsRequest struct {
	Input string  `json:"input"`
	Voice string  `json:"voice,omitempty"`
	Speed float64 `json:"speed,omitempty"`
	Pitch any     `json:"pitch,omitempty"`
	Style string  `json:"style,omitempty"`
}

func registerTTSRoute(root *http.ServeMux, auth BridgeAuth) {
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
