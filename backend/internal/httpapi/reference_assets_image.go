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

	"qiantie/backend/internal/localexecutor"
	"qiantie/backend/internal/novelfetchworkshop"
)

const localReferenceAssetImageTaskPrefix = "go-local-image:"

// RegisterReferenceAssetImageGenerateRoute keeps the older configured image
// API available, while allowing the Script page to use the same paired Doubao
// executor used by the other production surfaces.
func RegisterReferenceAssetImageGenerateRoute(root *http.ServeMux, middleware BridgeAuth, store *UserConfigStore, catalogStores ...novelfetchworkshop.Store) {
	if root == nil || middleware.AuthStore == nil || store == nil {
		return
	}
	var catalogStore novelfetchworkshop.Store
	if len(catalogStores) > 0 {
		catalogStore = catalogStores[0]
	}
	// The wrapper keeps existing callers source-compatible. Router uses the
	// internal registration helper below when the paired executor is available.
	registerReferenceAssetImageGenerateRoute(root, middleware, store, catalogStore, nil)
}

func registerReferenceAssetImageGenerateRoute(root *http.ServeMux, middleware BridgeAuth, store *UserConfigStore, catalogStore novelfetchworkshop.Store, executors *localexecutor.Service) {
	root.Handle("POST /api/reference-assets/generate", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
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
		prompt := strings.TrimSpace(strings.Join([]string{toString(input["name"]), toString(input["description"]), toString(input["generation_guidance"]), toString(input["context"]), toString(input["image_instruction"])}, "\n"))
		if prompt == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "缺少生成内容：请填写人物外形描述或生成引导。", "code": "IMAGE_GENERATION_EMPTY_PROMPT"})
			return
		}
		if strings.EqualFold(strings.TrimSpace(toString(input["imageProvider"])), "doubao_local_executor") || strings.EqualFold(strings.TrimSpace(toString(input["imageModelId"])), "local-doubao-executor-image") {
			if executors == nil {
				writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "豆包本地执行器服务不可用"})
				return
			}
			model := firstText(input["imageModelId"], "Seedream 4.5")
			if strings.EqualFold(model, "local-doubao-executor-image") {
				model = "Seedream 4.5"
			}
			job, err := executors.CreateJob(req.Context(), identity.Username, localexecutor.CreateJobInput{
				SourceTaskID: "reference-asset-image:" + time.Now().UTC().Format("20060102150405.000000000"),
				Platform:     localexecutor.PlatformDoubao,
				Payload: map[string]any{
					"type": "image", "model": model, "prompt": prompt,
					"ratio": referenceAssetImageRatio(input),
				},
			})
			if err != nil {
				writeLocalJobError(w, err)
				return
			}
			writeJSON(w, http.StatusAccepted, map[string]any{"ok": true, "taskId": localReferenceAssetImageTaskPrefix + job.ID, "status": "processing"})
			return
		}
		settings := readReferenceAssetImageSettingsForModel(req.Context(), store, identity.Username, catalogStore, toString(input["imageModelId"]))
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
		timeout := referenceAssetTimeout(settings["timeout_seconds"])
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
		item := referenceAssetImageItem(result)
		image, mime, err := decodeReferenceAssetImageResult(ctx, item)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "图片AI未返回可用的图像内容。", "code": "IMAGE_AI_EMPTY_RESULT"})
			return
		}
		assetType := strings.ToLower(strings.TrimSpace(toString(input["asset_type"])))
		if !validReferenceAssetType(assetType) {
			assetType = "character"
		}
		assetID := strings.TrimSpace(toString(input["asset_id"]))
		if !validReferenceAssetID(assetID) {
			assetID = "gen_" + fmt.Sprint(time.Now().Unix())
		}
		revision := referenceAssetRevision("candidate")
		path := filepath.Join(referenceAssetDir(store.BaseDir, identity.Username, assetType), assetID+"_"+revision+referenceAssetImageExtension(mime))
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "图片保存失败"})
			return
		}
		if err := os.WriteFile(path, image, 0o600); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "图片保存失败"})
			return
		}
		metadata := referenceAssetMetadata(store.BaseDir, identity.Username, assetType, assetID)
		metadata["ok"] = true
		metadata["asset_id"] = assetID
		metadata["asset_type"] = assetType
		metadata["revision"] = revision
		metadata["url"] = "/api/reference-assets/file/" + assetType + "/" + assetID + "/" + revision
		metadata["main_origin"] = "generated"
		writeJSON(w, http.StatusOK, metadata)
	})))
	root.Handle("GET /api/reference-assets/generate/{taskId}", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		taskID := strings.TrimSpace(req.PathValue("taskId"))
		if !strings.HasPrefix(taskID, localReferenceAssetImageTaskPrefix) || executors == nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "图片任务不存在"})
			return
		}
		job, err := executors.GetJob(req.Context(), identity.Username, strings.TrimPrefix(taskID, localReferenceAssetImageTaskPrefix))
		if err != nil {
			writeLocalJobError(w, err)
			return
		}
		result := map[string]any{"ok": true, "taskId": taskID, "status": "processing"}
		switch job.State {
		case localexecutor.JobFailed, localexecutor.JobCancelled:
			result["status"], result["error"] = "failed", job.ErrorMessage
		case localexecutor.JobSucceeded:
			if strings.TrimSpace(job.ArtifactID) == "" {
				result["status"], result["error"] = "failed", "执行器未返回图片结果"
			} else {
				result["status"] = "succeeded"
				result["url"] = "/api/shuihuo-production/local-executor-artifacts/" + url.PathEscape(job.ArtifactID)
			}
		}
		writeJSON(w, http.StatusOK, result)
	})))
}

func referenceAssetImageItem(result map[string]any) map[string]any {
	const maxDepth = 4
	var walk func(any, int) map[string]any
	walk = func(value any, depth int) map[string]any {
		if depth > maxDepth {
			return nil
		}
		switch current := value.(type) {
		case map[string]any:
			for _, key := range []string{"b64_json", "base64", "image_base64", "url", "image_url"} {
				if strings.TrimSpace(toString(current[key])) != "" {
					return current
				}
			}
			for _, key := range []string{"data", "images", "output", "result"} {
				if item := walk(current[key], depth+1); item != nil {
					return item
				}
			}
			for _, nested := range current {
				if item := walk(nested, depth+1); item != nil {
					return item
				}
			}
		case []any:
			for _, nested := range current {
				if item := walk(nested, depth+1); item != nil {
					return item
				}
			}
		}
		return nil
	}
	return walk(result, 0)
}

func decodeReferenceAssetImageResult(ctx context.Context, result map[string]any) ([]byte, string, error) {
	if result == nil {
		return nil, "", fmt.Errorf("image result is empty")
	}
	for _, key := range []string{"b64_json", "base64", "image_base64"} {
		if encoded := strings.TrimSpace(toString(result[key])); encoded != "" {
			if strings.HasPrefix(encoded, "data:") {
				if comma := strings.Index(encoded, ","); comma >= 0 {
					encoded = encoded[comma+1:]
				}
			}
			image, err := base64.StdEncoding.DecodeString(encoded)
			if err != nil || len(image) == 0 {
				if err == nil {
					err = fmt.Errorf("decoded image is empty")
				}
				return nil, "", err
			}
			return image, "image/png", nil
		}
	}
	imageURL := ""
	for _, key := range []string{"url", "image_url"} {
		if imageURL = strings.TrimSpace(toString(result[key])); imageURL != "" {
			break
		}
	}
	if imageURL == "" {
		return nil, "", fmt.Errorf("image result has no image data")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return nil, "", err
	}
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, "", fmt.Errorf("image URL returned status %d", response.StatusCode)
	}
	mime := strings.ToLower(strings.TrimSpace(strings.Split(response.Header.Get("Content-Type"), ";")[0]))
	if mime != "image/jpeg" && mime != "image/webp" && mime != "image/png" {
		return nil, "", fmt.Errorf("unsupported image content type %q", mime)
	}
	image, err := io.ReadAll(io.LimitReader(response.Body, 64<<20))
	if err != nil {
		return nil, "", err
	}
	if len(image) == 0 {
		return nil, "", fmt.Errorf("image response is empty")
	}
	return image, mime, nil
}

func referenceAssetImageRatio(input map[string]any) string {
	ratio := strings.TrimSpace(toString(input["aspect_ratio"]))
	if ratio == "16:9" || ratio == "9:16" || ratio == "1:1" {
		return ratio
	}
	return "9:16"
}
