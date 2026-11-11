package httpapi

import (
	"crypto/rand"
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

	"qiantie/backend/internal/safehttp"
)

var referenceAssetFolders = map[string]string{"character": "characters", "scene": "scenes", "prop": "props"}

func validReferenceAssetType(value string) bool {
	_, ok := referenceAssetFolders[strings.ToLower(strings.TrimSpace(value))]
	return ok
}
func validReferenceAssetID(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > 120 || strings.ContainsAny(value, `/\\`) {
		return false
	}
	for _, r := range value {
		if !(r == '_' || r == '-' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9') {
			return false
		}
	}
	return true
}
func validReferenceAssetVariant(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "source" || value == "main" || value == "thumb" {
		return true
	}
	return strings.HasPrefix(value, "source_") || strings.HasPrefix(value, "candidate_")
}

func referenceAssetDir(usersDir, username, assetType string) (string, error) {
	// Keep the existing on-disk location readable after removing the workbench.
	return UserScopedPath(usersDir, username, "novel-panel", "reference_assets", referenceAssetFolders[assetType])
}
func referenceAssetPath(usersDir, username, assetType, assetID, variant string) string {
	dir, err := referenceAssetDir(usersDir, username, assetType)
	if err != nil {
		return ""
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	stem := assetID + "_" + variant
	var best string
	var bestTime time.Time
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".png") && !strings.HasSuffix(strings.ToLower(entry.Name()), ".jpg") && !strings.HasSuffix(strings.ToLower(entry.Name()), ".jpeg") && !strings.HasSuffix(strings.ToLower(entry.Name()), ".webp") {
			continue
		}
		if !strings.HasPrefix(entry.Name(), stem+".") && !(variant == "source" && strings.HasPrefix(entry.Name(), assetID+"_source_")) {
			continue
		}
		info, _ := entry.Info()
		if info != nil && (best == "" || info.ModTime().After(bestTime)) {
			best, bestTime = filepath.Join(dir, entry.Name()), info.ModTime()
		}
	}
	return best
}

func referenceAssetMetadata(usersDir, username, assetType, assetID string) map[string]any {
	result := map[string]any{}
	for _, variant := range []string{"source", "main", "thumb"} {
		path := referenceAssetPath(usersDir, username, assetType, assetID, variant)
		key := variant + "_image"
		result["has_"+key] = path != ""
		if path != "" {
			if info, err := os.Stat(path); err == nil {
				result[variant+"_updated_at"] = info.ModTime().UTC().Format(time.RFC3339Nano)
				result[variant+"_revision"] = fmt.Sprintf("%d", info.ModTime().UnixNano())
			}
		}
	}
	return result
}

func referenceAssetDataURL(value string) ([]byte, string, error) {
	parts := strings.SplitN(strings.TrimSpace(value), ",", 2)
	if len(parts) != 2 || !strings.HasPrefix(strings.ToLower(parts[0]), "data:") || !strings.Contains(strings.ToLower(parts[0]), ";base64") {
		return nil, "", fmt.Errorf("图片数据必须是 base64 data URL")
	}
	mime := strings.TrimPrefix(strings.SplitN(parts[0], ";", 2)[0], "data:")
	if mime == "" {
		mime = "image/png"
	}
	if mime != "image/png" && mime != "image/jpeg" && mime != "image/webp" {
		return nil, "", fmt.Errorf("图片格式不支持")
	}
	payload, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(parts[1], "\n", ""))
	if err != nil || len(payload) == 0 {
		return nil, "", fmt.Errorf("图片内容为空或格式无效")
	}
	return payload, mime, nil
}

func referenceAssetImageExtension(mime string) string {
	if mime == "image/jpeg" {
		return ".jpg"
	}
	if mime == "image/webp" {
		return ".webp"
	}
	return ".png"
}
func referenceAssetRevision(prefix string) string {
	bytes := make([]byte, 6)
	_, _ = rand.Read(bytes)
	return fmt.Sprintf("%s_%d_%x", prefix, time.Now().UnixMilli(), bytes)
}

func writeReferenceAsset(usersDir, username, assetType, assetID, variant string, payload []byte, mime string) (string, error) {
	dir, err := referenceAssetDir(usersDir, username, assetType)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	stem := assetID + "_" + variant
	entries, _ := os.ReadDir(dir)
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), stem+".") {
			_ = os.Remove(filepath.Join(dir, entry.Name()))
		}
	}
	destination := filepath.Join(dir, stem+referenceAssetImageExtension(mime))
	tmp := destination + ".tmp"
	if err := os.WriteFile(tmp, payload, 0o600); err != nil {
		return "", err
	}
	if err := os.Rename(tmp, destination); err != nil {
		return "", err
	}
	return destination, nil
}

func RegisterReferenceAssetRoutes(root *http.ServeMux, middleware BridgeAuth, usersDir string) {
	if root == nil || strings.TrimSpace(usersDir) == "" {
		return
	}
	handler := middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		if req.Method == http.MethodGet {
			assetType, assetID, variant := req.PathValue("assetType"), req.PathValue("assetId"), req.PathValue("variant")
			if !validReferenceAssetType(assetType) || !validReferenceAssetID(assetID) || !validReferenceAssetVariant(variant) {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "参考图参数无效"})
				return
			}
			path := referenceAssetPath(usersDir, identity.Username, strings.ToLower(assetType), assetID, strings.ToLower(variant))
			if path == "" {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "参考图片不存在。", "code": "REFERENCE_ASSET_NOT_FOUND"})
				return
			}
			mime := "image/png"
			switch strings.ToLower(filepath.Ext(path)) {
			case ".jpg", ".jpeg":
				mime = "image/jpeg"
			case ".webp":
				mime = "image/webp"
			}
			w.Header().Set("Cache-Control", "private, max-age=3600")
			w.Header().Set("Content-Type", mime)
			w.Header().Set("X-Content-Type-Options", "nosniff")
			http.ServeFile(w, req, path)
			return
		}
		var input map[string]any
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 16<<20)).Decode(&input) != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求内容无效"})
			return
		}
		assetType, _ := input["asset_type"].(string)
		assetID, _ := input["asset_id"].(string)
		assetType = strings.ToLower(strings.TrimSpace(assetType))
		assetID = strings.TrimSpace(assetID)
		if !validReferenceAssetType(assetType) || !validReferenceAssetID(assetID) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "参考图参数无效"})
			return
		}
		assetPath := req.URL.Path
		if assetPath == "/api/reference-assets/upload" {
			variant, _ := input["variant"].(string)
			if variant != "" && variant != "source" {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "上传图片仅支持 source 类型"})
				return
			}
			payload, mime, err := referenceAssetDataURL(fmt.Sprint(input["data_url"]))
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "保存参考图失败：" + err.Error(), "code": "REFERENCE_ASSET_UPLOAD_FAILED"})
				return
			}
			revision := referenceAssetRevision("source")
			dir, dirErr := referenceAssetDir(usersDir, identity.Username, assetType)
			if dirErr != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "用户名不合法"})
				return
			}
			path := filepath.Join(dir, assetID+"_"+revision+referenceAssetImageExtension(mime))
			if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "保存参考图失败"})
				return
			}
			if err := os.WriteFile(path, payload, 0o600); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "保存参考图失败"})
				return
			}
			meta := referenceAssetMetadata(usersDir, identity.Username, assetType, assetID)
			meta["ok"] = true
			meta["asset_id"] = assetID
			meta["asset_type"] = assetType
			meta["variant"] = revision
			meta["revision"] = revision
			meta["url"] = "/api/reference-assets/file/" + assetType + "/" + assetID + "/" + revision
			writeJSON(w, http.StatusOK, meta)
			return
		}
		if assetPath == "/api/reference-assets/use-source-as-main" {
			source := referenceAssetPath(usersDir, identity.Username, assetType, assetID, "source")
			if source == "" {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "还没有上传参考图", "code": "REFERENCE_ASSET_SOURCE_MISSING"})
				return
			}
			payload, err := os.ReadFile(source)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "读取参考图失败"})
				return
			}
			mime := "image/png"
			if strings.HasSuffix(strings.ToLower(source), ".jpg") || strings.HasSuffix(strings.ToLower(source), ".jpeg") {
				mime = "image/jpeg"
			} else if strings.HasSuffix(strings.ToLower(source), ".webp") {
				mime = "image/webp"
			}
			if _, err := writeReferenceAsset(usersDir, identity.Username, assetType, assetID, "main", payload, mime); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "设置主参考图失败"})
				return
			}
			result := referenceAssetMetadata(usersDir, identity.Username, assetType, assetID)
			result["ok"] = true
			result["asset_id"] = assetID
			result["asset_type"] = assetType
			result["url"] = "/api/reference-assets/file/" + assetType + "/" + assetID + "/main"
			result["main_origin"] = "uploaded"
			writeJSON(w, http.StatusOK, result)
			return
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "参考图接口不存在"})
	}))
	root.Handle("GET /api/reference-assets/file/{assetType}/{assetId}/{variant}", handler)
	root.Handle("POST /api/reference-assets/upload", handler)
	root.Handle("POST /api/reference-assets/use-source-as-main", handler)
	root.Handle("GET /api/reference-assets/legacy", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		target, err := legacyTOSAssetURL(req.URL.Query().Get("url"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "旧图片地址无效，请重新上传。", "code": "REFERENCE_ASSET_LEGACY_INVALID"})
			return
		}
		ctx := req.Context()
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "旧图片无法读取，请重新上传。", "code": "REFERENCE_ASSET_LEGACY_NOT_FOUND"})
			return
		}
		response, err := safehttp.NewClientWithTimeout(15 * time.Second).Do(request)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "旧图片无法读取，请重新上传。", "code": "REFERENCE_ASSET_LEGACY_NOT_FOUND"})
			return
		}
		defer response.Body.Close()
		contentType := strings.ToLower(strings.TrimSpace(strings.Split(response.Header.Get("Content-Type"), ";")[0]))
		if response.StatusCode < 200 || response.StatusCode >= 300 || (contentType != "image/png" && contentType != "image/jpeg" && contentType != "image/webp") {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "旧图片无法读取，请重新上传。", "code": "REFERENCE_ASSET_LEGACY_NOT_FOUND"})
			return
		}
		body, err := io.ReadAll(io.LimitReader(response.Body, 32<<20))
		if err != nil || len(body) == 0 {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "旧图片无法读取，请重新上传。", "code": "REFERENCE_ASSET_LEGACY_NOT_FOUND"})
			return
		}
		w.Header().Set("Cache-Control", "private, max-age=300")
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	})))
}

func legacyTOSAssetURL(raw string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || !strings.HasSuffix(strings.ToLower(parsed.Hostname()), ".volces.com") || !strings.Contains(parsed.Path, "/reference-assets/") {
		return "", fmt.Errorf("unsupported legacy asset url")
	}
	return parsed.String(), nil
}
