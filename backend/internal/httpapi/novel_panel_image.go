package httpapi

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"qiantie/backend/internal/novelfetchworkshop"
)

func RegisterNovelPanelImageGenerateCompat(root *http.ServeMux, middleware BridgeAuth, store *UserConfigStore, catalogStores ...novelfetchworkshop.Store) {
	if root == nil || middleware.AuthStore == nil || store == nil {
		return
	}
	root.Handle("POST /api/novel-panel/reference-assets/generate", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input map[string]any
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 16<<20)).Decode(&input) != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求内容无效"})
			return
		}
		var catalogStore novelfetchworkshop.Store
		if len(catalogStores) > 0 {
			catalogStore = catalogStores[0]
		}
		settings := readPanelImageSettingsWithCatalog(req.Context(), store, identity.Username, catalogStore)
		missing := []string{}
		for _, key := range []string{"base_url", "model", "api_key"} {
			if strings.TrimSpace(toString(settings[key])) == "" {
				missing = append(missing, key)
			}
		}
		if len(missing) > 0 {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "图片AI未配置：缺少 " + strings.Join(missing, "、") + "。", "code": "IMAGE_SETTINGS_INCOMPLETE"})
			return
		}
		prompt := strings.TrimSpace(strings.Join([]string{toString(input["name"]), toString(input["description"]), toString(input["generation_guidance"]), toString(input["context"]), toString(input["image_instruction"])}, "\n"))
		if prompt == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "缺少生成内容：请填写人物外形描述或生成引导。", "code": "IMAGE_GENERATION_EMPTY_PROMPT"})
			return
		}
		base := strings.TrimRight(toString(settings["base_url"]), "/")
		pathValue := strings.TrimSpace(toString(settings["generate_path"]))
		if pathValue == "" {
			pathValue = "/images/generations"
		}
		parsed, err := url.Parse(base)
		if err != nil || parsed.Host == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "图片 API 地址格式不正确。", "code": "IMAGE_AI_URL_INVALID"})
			return
		}
		if parsed.Path == "" || parsed.Path == "/" {
			base += "/v1"
		}
		endpoint := strings.TrimRight(base, "/") + "/" + strings.TrimLeft(pathValue, "/")
		payload := map[string]any{"model": settings["model"], "prompt": prompt, "n": 1, "size": settings["size"], "response_format": "b64_json"}
		if extra, ok := settings["extra_json"].(map[string]any); ok {
			for key, value := range extra {
				payload[key] = value
			}
		}
		body, _ := json.Marshal(payload)
		timeout := panelTimeout(settings["timeout_seconds"])
		if timeout < 30 {
			timeout = 30
		}
		ctx, cancel := context.WithTimeout(req.Context(), time.Duration(timeout)*time.Second)
		defer cancel()
		upstreamReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(string(body)))
		upstreamReq.Header.Set("Content-Type", "application/json")
		upstreamReq.Header.Set("Authorization", "Bearer "+toString(settings["api_key"]))
		response, err := http.DefaultClient.Do(upstreamReq)
		if err != nil {
			if ctx.Err() != nil {
				writeJSON(w, http.StatusGatewayTimeout, map[string]string{"error": fmt.Sprintf("图片AI请求超时（超过 %d 秒）", timeout), "code": "IMAGE_AI_TIMEOUT"})
				return
			}
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "图片AI请求失败", "code": "IMAGE_AI_REQUEST_FAILED"})
			return
		}
		defer response.Body.Close()
		raw, _ := io.ReadAll(io.LimitReader(response.Body, 64<<20))
		var result map[string]any
		_ = json.Unmarshal(raw, &result)
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			writeJSON(w, http.StatusBadGateway, map[string]any{"error": "图片AI请求失败", "code": "IMAGE_AI_REQUEST_FAILED", "status": response.StatusCode})
			return
		}
		data, _ := result["data"].([]any)
		if len(data) == 0 {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "图片AI未返回可用的图像内容。", "code": "IMAGE_AI_EMPTY_RESULT"})
			return
		}
		item, _ := data[0].(map[string]any)
		var image []byte
		mime := "image/png"
		if encoded := strings.TrimSpace(toString(item["b64_json"])); encoded != "" {
			image, err = base64.StdEncoding.DecodeString(encoded)
		} else if imageURL := strings.TrimSpace(toString(item["url"])); imageURL != "" {
			var imageResponse *http.Response
			imageResponse, err = http.Get(imageURL)
			if err == nil {
				defer imageResponse.Body.Close()
				image, err = io.ReadAll(io.LimitReader(imageResponse.Body, 64<<20))
				contentType := strings.ToLower(imageResponse.Header.Get("Content-Type"))
				if strings.Contains(contentType, "jpeg") {
					mime = "image/jpeg"
				} else if strings.Contains(contentType, "webp") {
					mime = "image/webp"
				}
			}
		}
		if err != nil || len(image) == 0 {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "图片AI未返回可用的图像内容。", "code": "IMAGE_AI_EMPTY_RESULT"})
			return
		}
		assetType := strings.ToLower(strings.TrimSpace(toString(input["asset_type"])))
		if !validPanelAssetType(assetType) {
			assetType = "character"
		}
		assetID := strings.TrimSpace(toString(input["asset_id"]))
		if !validPanelAssetID(assetID) {
			assetID = "gen_" + fmt.Sprint(time.Now().Unix())
		}
		revision := panelRevision("candidate")
		path := filepath.Join(panelAssetDir(store.BaseDir, identity.Username, assetType), assetID+"_"+revision+panelImageExtension(mime))
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "图片保存失败"})
			return
		}
		if err := os.WriteFile(path, image, 0o600); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "图片保存失败"})
			return
		}
		metadata := panelAssetMetadata(store.BaseDir, identity.Username, assetType, assetID)
		metadata["ok"] = true
		metadata["asset_id"] = assetID
		metadata["asset_type"] = assetType
		metadata["revision"] = revision
		metadata["url"] = "/api/reference-assets/file/" + assetType + "/" + assetID + "/" + revision
		metadata["main_origin"] = "generated"
		writeJSON(w, http.StatusOK, metadata)
	})))
}
