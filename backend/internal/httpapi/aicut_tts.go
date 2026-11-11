package httpapi

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"
	"unicode"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/aicut"
	"qiantie/backend/internal/novelfetchworkshop"
)

const aicutTTSMaxRunes = 5000

type aicutTTSRequest struct {
	Text  string  `json:"text"`
	Name  string  `json:"name,omitempty"`
	Voice string  `json:"voice,omitempty"`
	Speed float64 `json:"speed,omitempty"`
	Pitch float64 `json:"pitch,omitempty"`
	Style string  `json:"style,omitempty"`
}

// RegisterAICutTTSRoutes turns TTS output into a durable AI-CUT project asset.
// It deliberately reuses the maintained TTS adapters and project object store;
// the browser never has to keep a temporary Blob alive for project playback.
func RegisterAICutTTSRoutes(mux *http.ServeMux, auth BridgeAuth, store aicut.Store, objects agentworkspace.AssetObjectStore, catalog novelfetchworkshop.Store, now func() time.Time) {
	if mux == nil || store == nil {
		return
	}
	if now == nil {
		now = time.Now
	}
	mux.Handle("POST /api/ai-cut/projects/{id}/tts", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok || strings.TrimSpace(identity.Username) == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		owner := strings.TrimSpace(identity.Username)
		projectID := strings.TrimSpace(req.PathValue("id"))
		if _, err := store.Project(req.Context(), owner, projectID); err != nil {
			writeAICutError(w, err)
			return
		}
		if objects == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "素材存储尚未配置"})
			return
		}
		var input aicutTTSRequest
		if !decodeAICutJSON(w, req, &input) {
			return
		}
		text := strings.TrimSpace(input.Text)
		if text == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "配音文本不能为空"})
			return
		}
		if len([]rune(text)) > aicutTTSMaxRunes {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "配音文本不能超过 5000 个字符"})
			return
		}
		if input.Speed == 0 {
			input.Speed = 1.8
		}
		if input.Speed < .5 || input.Speed > 2 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "语速必须在 0.5 到 2 之间"})
			return
		}
		if input.Pitch < -50 || input.Pitch > 50 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "音调必须在 -50 到 50 之间"})
			return
		}
		if strings.TrimSpace(input.Voice) == "" {
			input.Voice = "zh-CN-XiaoxiaoNeural"
		}
		if strings.TrimSpace(input.Style) == "" {
			input.Style = "general"
		}

		ttsInput := ttsRequest{Input: text, Voice: strings.TrimSpace(input.Voice), Speed: input.Speed, Pitch: input.Pitch, Style: strings.TrimSpace(input.Style)}
		payload, contentType, err := requestAICutTTSBytes(req.Context(), owner, ttsInput, catalog)
		if err != nil {
			if req.Context().Err() != nil || err == context.Canceled || err == context.DeadlineExceeded {
				writeJSON(w, http.StatusRequestTimeout, map[string]string{"error": "配音生成已取消"})
				return
			}
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "配音生成失败：" + err.Error()})
			return
		}
		contentType = strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
		if len(payload) == 0 || !strings.HasPrefix(contentType, "audio/") {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "配音生成失败：服务没有返回有效音频"})
			return
		}
		extension := aicutTTSExtension(contentType)
		assetID := newAgentID("aicut-asset")
		objectKey, keyErr := agentworkspace.AgentAssetObjectKey(owner, projectID, assetID, extension)
		if keyErr != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "配音素材地址创建失败"})
			return
		}
		objectKey = strings.Replace(objectKey, "agent-assets/", "aicut-assets/", 1)
		asset := aicut.Asset{
			ID: assetID, ProjectID: projectID, Kind: "audio",
			Name: aicutTTSAssetName(input.Name, extension, now()), ContentType: contentType,
			ObjectKey: objectKey, ByteSize: int64(len(payload)), CreatedAt: now().UTC(),
		}
		objectAsset := aicutObjectAsset(asset)
		if err := objects.Put(req.Context(), objectAsset, bytes.NewReader(payload)); err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "配音素材保存失败"})
			return
		}
		if durationMS, probeErr := probeAICutAssetDuration(req.Context(), asset, objects); probeErr == nil {
			asset.DurationMS = durationMS
		}
		if err := store.CreateAsset(req.Context(), owner, asset); err != nil {
			_ = objects.Delete(req.Context(), objectAsset)
			writeAICutError(w, err)
			return
		}
		asset.URL = aicutSignedAssetContentURL(auth, owner, projectID, asset.ID, "content", now())
		writeJSON(w, http.StatusCreated, map[string]any{"asset": asset})
	})))
}

func requestAICutTTSBytes(ctx context.Context, owner string, input ttsRequest, catalog novelfetchworkshop.Store) ([]byte, string, error) {
	if catalog != nil {
		if model, found := configuredTTSModel(ctx, catalog, owner); found {
			return requestConfiguredTTSBytes(ctx, input, model)
		}
	}
	return requestDefaultTTSBytes(ctx, input)
}

func aicutTTSExtension(contentType string) string {
	switch strings.ToLower(strings.TrimSpace(contentType)) {
	case "audio/wav", "audio/wave", "audio/x-wav":
		return ".wav"
	case "audio/ogg":
		return ".ogg"
	case "audio/webm":
		return ".webm"
	case "audio/flac", "audio/x-flac":
		return ".flac"
	case "audio/mp4", "audio/x-m4a":
		return ".m4a"
	default:
		return ".mp3"
	}
}

func aicutTTSAssetName(value, extension string, now time.Time) string {
	name := strings.TrimSpace(value)
	name = strings.TrimSuffix(name, path.Ext(name))
	name = strings.Map(func(r rune) rune {
		if r == '/' || r == '\\' || unicode.IsControl(r) {
			return '-'
		}
		return r
	}, name)
	runes := []rune(strings.Trim(name, " .-"))
	if len(runes) > 64 {
		runes = runes[:64]
	}
	name = strings.TrimSpace(string(runes))
	if name == "" {
		name = fmt.Sprintf("AI配音-%s", now.Format("20060102-150405"))
	}
	return name + extension
}
