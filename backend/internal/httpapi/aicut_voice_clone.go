package httpapi

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/aicut"
	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/novelfetchworkshop"
)

const (
	aicutVoiceCloneMaxReferenceBytes = 20 << 20
	aicutVoiceCloneMaxRunes          = 1000
)

type aicutVoiceCloneRequest struct {
	AssetID string  `json:"assetId"`
	Text    string  `json:"text"`
	Name    string  `json:"name,omitempty"`
	Speed   float64 `json:"speed,omitempty"`
	Pitch   float64 `json:"pitch,omitempty"`
	Style   string  `json:"style,omitempty"`
}

type aicutVoiceCloner func(context.Context, string, aicut.Asset, []byte, aicutVoiceCloneRequest) ([]byte, string, error)

// RegisterAICutVoiceCloneRoutes synthesizes one project audio asset from a
// project-owned reference recording. The reference remains server-side.
func RegisterAICutVoiceCloneRoutes(mux *http.ServeMux, auth BridgeAuth, store aicut.Store, objects agentworkspace.AssetObjectStore, catalog novelfetchworkshop.Store, cloner aicutVoiceCloner, now func() time.Time) {
	if mux == nil || store == nil {
		return
	}
	if now == nil {
		now = time.Now
	}
	mux.Handle("POST /api/ai-cut/projects/{id}/voice-clone", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok || strings.TrimSpace(identity.Username) == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		owner, projectID := strings.TrimSpace(identity.Username), strings.TrimSpace(req.PathValue("id"))
		if _, err := store.Project(req.Context(), owner, projectID); err != nil {
			writeAICutError(w, err)
			return
		}
		var input aicutVoiceCloneRequest
		if !decodeAICutJSON(w, req, &input) {
			return
		}
		input.AssetID, input.Text, input.Name = strings.TrimSpace(input.AssetID), strings.TrimSpace(input.Text), strings.TrimSpace(input.Name)
		if input.Text == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "克隆音色试听文本不能为空"})
			return
		}
		if len([]rune(input.Text)) > aicutVoiceCloneMaxRunes {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "克隆音色试听文本不能超过 1000 个字符"})
			return
		}
		asset, err := store.Asset(req.Context(), owner, projectID, input.AssetID)
		if err != nil {
			writeAICutError(w, err)
			return
		}
		if asset.Kind != "audio" || !strings.HasPrefix(strings.ToLower(asset.ContentType), "audio/") {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "声音克隆目前仅支持音频素材"})
			return
		}
		if input.Speed == 0 {
			input.Speed = 1
		}
		if input.Speed < .5 || input.Speed > 2 || input.Pitch < -50 || input.Pitch > 50 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "声音克隆语速或音调超出范围"})
			return
		}
		if objects == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "素材存储尚未配置"})
			return
		}
		body, err := objects.Open(req.Context(), aicutObjectAsset(asset))
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "参考音频文件不存在"})
			return
		}
		reference, readErr := io.ReadAll(io.LimitReader(body, aicutVoiceCloneMaxReferenceBytes+1))
		_ = body.Close()
		if readErr != nil || len(reference) == 0 {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "参考音频读取失败"})
			return
		}
		if len(reference) > aicutVoiceCloneMaxReferenceBytes {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "参考音频不能超过 20 MB"})
			return
		}
		activeCloner := cloner
		if activeCloner == nil {
			activeCloner = func(ctx context.Context, owner string, asset aicut.Asset, reference []byte, input aicutVoiceCloneRequest) ([]byte, string, error) {
				return requestConfiguredAICutVoiceCloneBytes(ctx, owner, asset, reference, input, catalog)
			}
		}
		payload, contentType, err := activeCloner(req.Context(), owner, asset, reference, input)
		if err != nil {
			if req.Context().Err() != nil || err == context.Canceled || err == context.DeadlineExceeded {
				writeJSON(w, http.StatusRequestTimeout, map[string]string{"error": "声音克隆已取消"})
				return
			}
			status := http.StatusBadGateway
			if errors.Is(err, batchfactoryv11.ErrUnavailable) {
				status = http.StatusServiceUnavailable
			}
			writeJSON(w, status, map[string]string{"error": "声音克隆失败：" + err.Error()})
			return
		}
		contentType = strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
		if len(payload) == 0 || !strings.HasPrefix(contentType, "audio/") {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "声音克隆失败：服务没有返回有效音频"})
			return
		}
		extension := aicutTTSExtension(contentType)
		createdAt, createdID := now().UTC(), newAgentID("aicut-asset")
		objectKey, err := agentworkspace.AgentAssetObjectKey(owner, projectID, createdID, extension)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "克隆音频素材地址创建失败"})
			return
		}
		name := input.Name
		if name == "" {
			name = strings.TrimSuffix(asset.Name, filepath.Ext(asset.Name)) + "-克隆音色"
		}
		created := aicut.Asset{ID: createdID, ProjectID: projectID, Kind: "audio", Name: aicutTTSAssetName(name, extension, createdAt), ContentType: contentType, ObjectKey: strings.Replace(objectKey, "agent-assets/", "aicut-assets/", 1), ByteSize: int64(len(payload)), CreatedAt: createdAt}
		if err := objects.Put(req.Context(), aicutObjectAsset(created), bytes.NewReader(payload)); err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "克隆音频素材保存失败"})
			return
		}
		if durationMS, probeErr := probeAICutAssetDuration(req.Context(), created, objects); probeErr == nil {
			created.DurationMS = durationMS
		}
		if err := store.CreateAsset(req.Context(), owner, created); err != nil {
			_ = objects.Delete(req.Context(), aicutObjectAsset(created))
			writeAICutError(w, err)
			return
		}
		created.URL = aicutSignedAssetContentURL(auth, owner, projectID, created.ID, "content", now())
		writeJSON(w, http.StatusCreated, map[string]any{"asset": created})
	})))
}

func requestConfiguredAICutVoiceCloneBytes(ctx context.Context, owner string, asset aicut.Asset, reference []byte, input aicutVoiceCloneRequest, catalog novelfetchworkshop.Store) ([]byte, string, error) {
	if catalog == nil {
		return nil, "", fmt.Errorf("%w: 未配置支持声音克隆的音频模型", batchfactoryv11.ErrUnavailable)
	}
	config, err := catalog.GetConfig(ctx, owner)
	if err != nil {
		return nil, "", fmt.Errorf("%w: 无法读取音频模型配置", batchfactoryv11.ErrUnavailable)
	}
	var selected map[string]any
	for _, model := range modelCatalog(config) {
		capabilities, _ := model["capabilities"].(map[string]any)
		if !strings.EqualFold(modelString(model, "kind"), "audio") || model["enabled"] != true || capabilities["supportsVoiceClone"] != true || modelString(model, "voiceCloneRequestTemplate") == "" || shuihuoModelCredential(model) == "" {
			continue
		}
		selected = model
		break
	}
	if selected == nil {
		return nil, "", fmt.Errorf("%w: 未配置支持声音克隆的音频模型", batchfactoryv11.ErrUnavailable)
	}
	template := modelString(selected, "voiceCloneRequestTemplate")
	template = strings.ReplaceAll(template, "{{reference_audio_base64}}", base64.StdEncoding.EncodeToString(reference))
	template = strings.ReplaceAll(template, "{{reference_audio_media_type}}", strings.TrimSpace(strings.Split(asset.ContentType, ";")[0]))
	cloneInput := ttsRequest{Input: input.Text, Voice: "cloned-reference", Speed: input.Speed, Pitch: input.Pitch, Style: input.Style}
	modelCopy := make(map[string]any, len(selected)+2)
	for key, value := range selected {
		modelCopy[key] = value
	}
	modelCopy["requestTemplate"] = template
	if mapping := modelString(selected, "voiceCloneResponseMapping"); mapping != "" {
		modelCopy["responseMapping"] = mapping
	}
	return requestConfiguredTTSBytes(ctx, cloneInput, modelCopy)
}
