package httpapi

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/aicut"
)

const aicutAssetUploadMaxBytes = int64(2 << 30)

var aicutProxyLocks sync.Map

func RegisterAICutAssetRoutes(mux *http.ServeMux, auth BridgeAuth, store aicut.Store, objects agentworkspace.AssetObjectStore, now func() time.Time) {
	if mux == nil || store == nil {
		return
	}
	if now == nil {
		now = time.Now
	}
	handle := func(pattern string, fn func(http.ResponseWriter, *http.Request, string)) {
		mux.Handle(pattern, auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			identity, ok := BridgeIdentityFromContext(req.Context())
			if !ok || strings.TrimSpace(identity.Username) == "" {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}
			fn(w, req, strings.TrimSpace(identity.Username))
		})))
	}

	handle("GET /api/ai-cut/projects/{id}/assets", func(w http.ResponseWriter, req *http.Request, owner string) {
		items, err := store.ListAssets(req.Context(), owner, req.PathValue("id"))
		if err != nil {
			writeAICutError(w, err)
			return
		}
		for index := range items {
			items[index].URL = aicutSignedAssetContentURL(auth, owner, req.PathValue("id"), items[index].ID, "content", now())
			if items[index].Kind == "font" && objects != nil {
				if family, familyErr := aicutFontFamilyForAsset(req.Context(), items[index], objects); familyErr == nil {
					items[index].FontFamily = family
				}
			}
			if items[index].Kind == "video" && objects != nil && aicutProxyAvailable(req.Context(), items[index], objects) {
				items[index].ProxyURL = aicutSignedAssetContentURL(auth, owner, req.PathValue("id"), items[index].ID, "proxy", now())
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{"assets": items})
	})
	handle("POST /api/ai-cut/projects/{id}/assets", func(w http.ResponseWriter, req *http.Request, owner string) {
		projectID := strings.TrimSpace(req.PathValue("id"))
		if _, err := store.Project(req.Context(), owner, projectID); err != nil {
			writeAICutError(w, err)
			return
		}
		if objects == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "素材存储尚未配置"})
			return
		}
		req.Body = http.MaxBytesReader(w, req.Body, aicutAssetUploadMaxBytes+(1<<20))
		if err := req.ParseMultipartForm(32 << 20); err != nil {
			if isBodyTooLarge(err) {
				writePayloadTooLarge(w, aicutAssetUploadMaxBytes+(1<<20))
				return
			}
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "素材文件无效或超过 2GB 限制"})
			return
		}
		file, header, err := req.FormFile("file")
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请选择素材文件"})
			return
		}
		defer file.Close()
		if header.Size <= 0 || header.Size > aicutAssetUploadMaxBytes {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "素材文件为空或超过 2GB 限制"})
			return
		}
		probe := make([]byte, 512)
		read, readErr := io.ReadFull(file, probe)
		if readErr != nil && !errors.Is(readErr, io.ErrUnexpectedEOF) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "素材文件读取失败"})
			return
		}
		probe = probe[:read]
		contentType, kind := detectedAICutAssetType(header.Header.Get("Content-Type"), header.Filename, probe)
		if kind == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "仅支持图片、视频、音频、SRT 字幕或 TTF/OTF 字体"})
			return
		}
		if kind == "font" && header.Size > aicutFontUploadMaxBytes {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "字体文件不能超过 20MB"})
			return
		}
		assetID := newAgentID("aicut-asset")
		objectKey, err := agentworkspace.AgentAssetObjectKey(owner, projectID, assetID, path.Ext(header.Filename))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "素材文件名无效"})
			return
		}
		objectKey = strings.Replace(objectKey, "agent-assets/", "aicut-assets/", 1)
		asset := aicut.Asset{ID: assetID, ProjectID: projectID, Kind: kind, Name: path.Base(header.Filename), ContentType: contentType, ObjectKey: objectKey, ByteSize: header.Size, CreatedAt: now().UTC()}
		objectAsset := aicutObjectAsset(asset)
		if err := objects.Put(req.Context(), objectAsset, io.MultiReader(bytes.NewReader(probe), file)); err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "素材上传失败"})
			return
		}
		if kind == "font" {
			family, familyErr := aicutFontFamilyForAsset(req.Context(), asset, objects)
			if familyErr != nil {
				_ = objects.Delete(req.Context(), objectAsset)
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "字体文件无效或缺少字体名称"})
				return
			}
			asset.FontFamily = family
		}
		if kind == "video" || kind == "audio" {
			if durationMS, probeErr := probeAICutAssetDuration(req.Context(), asset, objects); probeErr == nil {
				asset.DurationMS = durationMS
			}
		}
		if err := store.CreateAsset(req.Context(), owner, asset); err != nil {
			_ = objects.Delete(req.Context(), objectAsset)
			writeAICutError(w, err)
			return
		}
		asset.URL = aicutSignedAssetContentURL(auth, owner, projectID, asset.ID, "content", now())
		writeJSON(w, http.StatusCreated, map[string]any{"asset": asset})
	})
	serveAssetContent := func(w http.ResponseWriter, req *http.Request, owner string) {
		if objects == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "素材存储尚未配置"})
			return
		}
		asset, err := store.Asset(req.Context(), owner, req.PathValue("id"), req.PathValue("assetID"))
		if err != nil {
			writeAICutError(w, err)
			return
		}
		body, err := objects.Open(req.Context(), aicutObjectAsset(asset))
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "素材读取失败"})
			return
		}
		defer body.Close()
		w.Header().Set("Content-Type", asset.ContentType)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", asset.ByteSize))
		w.Header().Set("Cache-Control", "private, max-age=300")
		_, _ = io.Copy(w, body)
	}
	registerAICutSignedContentRoute(mux, auth, now, "GET /api/ai-cut/projects/{id}/assets/{assetID}/content", "content", serveAssetContent)
	handle("POST /api/ai-cut/projects/{id}/assets/{assetID}/proxy", func(w http.ResponseWriter, req *http.Request, owner string) {
		if objects == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "素材存储尚未配置"})
			return
		}
		projectID := strings.TrimSpace(req.PathValue("id"))
		asset, err := store.Asset(req.Context(), owner, projectID, req.PathValue("assetID"))
		if err != nil {
			writeAICutError(w, err)
			return
		}
		if asset.Kind != "video" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "仅视频素材可以生成代理"})
			return
		}
		lockValue, _ := aicutProxyLocks.LoadOrStore(owner+"\x00"+projectID+"\x00"+asset.ID, &sync.Mutex{})
		lock := lockValue.(*sync.Mutex)
		lock.Lock()
		defer lock.Unlock()
		asset.URL = aicutSignedAssetContentURL(auth, owner, projectID, asset.ID, "content", now())
		asset.ProxyURL = aicutSignedAssetContentURL(auth, owner, projectID, asset.ID, "proxy", now())
		if aicutProxyAvailable(req.Context(), asset, objects) {
			writeJSON(w, http.StatusOK, map[string]any{"asset": asset, "reused": true})
			return
		}
		byteSize, err := generateAICutProxy(req.Context(), asset, objects)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				writeJSON(w, http.StatusRequestTimeout, map[string]string{"error": "代理媒体生成已取消"})
				return
			}
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "代理媒体生成失败：" + err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"asset": asset, "proxyByteSize": byteSize, "reused": false})
	})
	serveProxyContent := func(w http.ResponseWriter, req *http.Request, owner string) {
		if objects == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "素材存储尚未配置"})
			return
		}
		asset, err := store.Asset(req.Context(), owner, req.PathValue("id"), req.PathValue("assetID"))
		if err != nil {
			writeAICutError(w, err)
			return
		}
		if asset.Kind != "video" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "仅视频素材可以使用代理"})
			return
		}
		body, err := objects.Open(req.Context(), aicutProxyObjectAsset(asset))
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "代理媒体尚未生成"})
			return
		}
		defer body.Close()
		w.Header().Set("Content-Type", "video/mp4")
		w.Header().Set("Cache-Control", "private, max-age=3600")
		_, _ = io.Copy(w, body)
	}
	registerAICutSignedContentRoute(mux, auth, now, "GET /api/ai-cut/projects/{id}/assets/{assetID}/proxy/content", "proxy", serveProxyContent)
}

func registerAICutSignedContentRoute(mux *http.ServeMux, auth BridgeAuth, now func() time.Time, pattern, variant string, serve func(http.ResponseWriter, *http.Request, string)) {
	authenticated := auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok || strings.TrimSpace(identity.Username) == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		serve(w, req, strings.TrimSpace(identity.Username))
	}))
	mux.Handle(pattern, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if owner, ok := validateAICutAssetAccess(auth.Secret, req, variant, now()); ok {
			serve(w, req, owner)
			return
		}
		authenticated.ServeHTTP(w, req)
	}))
}

func aicutSignedAssetContentURL(auth BridgeAuth, owner, projectID, assetID, variant string, issuedAt time.Time) string {
	base := aicutAssetContentURL(projectID, assetID)
	if variant == "proxy" {
		base = aicutAssetProxyContentURL(projectID, assetID)
	}
	secret := strings.TrimSpace(auth.Secret)
	if secret == "" {
		return base
	}
	expires := issuedAt.UTC().Add(24 * time.Hour).Unix()
	signature := aicutAssetAccessSignature(secret, owner, projectID, assetID, variant, expires)
	query := url.Values{"owner": []string{owner}, "expires": []string{strconv.FormatInt(expires, 10)}, "access": []string{signature}}
	return base + "?" + query.Encode()
}

func validateAICutAssetAccess(secret string, req *http.Request, variant string, now time.Time) (string, bool) {
	secret = strings.TrimSpace(secret)
	owner := strings.TrimSpace(req.URL.Query().Get("owner"))
	expires, err := strconv.ParseInt(req.URL.Query().Get("expires"), 10, 64)
	provided := strings.TrimSpace(req.URL.Query().Get("access"))
	if secret == "" || owner == "" || err != nil || provided == "" || now.UTC().Unix() > expires {
		return "", false
	}
	expected := aicutAssetAccessSignature(secret, owner, req.PathValue("id"), req.PathValue("assetID"), variant, expires)
	providedBytes, decodeErr := hex.DecodeString(provided)
	expectedBytes, _ := hex.DecodeString(expected)
	if decodeErr != nil || !hmac.Equal(providedBytes, expectedBytes) {
		return "", false
	}
	return owner, true
}

func aicutAssetAccessSignature(secret, owner, projectID, assetID, variant string, expires int64) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(owner + "\n" + projectID + "\n" + assetID + "\n" + variant + "\n" + strconv.FormatInt(expires, 10)))
	return hex.EncodeToString(mac.Sum(nil))
}

func generateAICutProxy(ctx context.Context, asset aicut.Asset, objects agentworkspace.AssetObjectStore) (int64, error) {
	dir, err := os.MkdirTemp("", "aicut-proxy-*")
	if err != nil {
		return 0, err
	}
	defer os.RemoveAll(dir)
	extension := strings.ToLower(path.Ext(asset.Name))
	if extension == "" || len(extension) > 12 {
		extension = ".video"
	}
	input := filepath.Join(dir, "source"+extension)
	output := filepath.Join(dir, "proxy.mp4")
	if err := copyAICutObjectToFile(ctx, objects, asset, input); err != nil {
		return 0, fmt.Errorf("读取原视频失败")
	}
	renderCtx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()
	args := []string{
		"-hide_banner", "-y", "-i", input,
		"-map", "0:v:0", "-map", "0:a?", "-sn", "-map_metadata", "-1",
		"-vf", "scale=w='if(gt(iw,ih),min(iw,960),-2)':h='if(gt(iw,ih),-2,min(ih,960))':flags=lanczos",
		"-c:v", "libx264", "-preset", "veryfast", "-crf", "30", "-maxrate", "1500k", "-bufsize", "3000k", "-pix_fmt", "yuv420p",
		"-c:a", "aac", "-b:a", "96k", "-movflags", "+faststart", output,
	}
	if err := runAICutFFmpeg(renderCtx, "ffmpeg", args...); err != nil {
		if renderCtx.Err() != nil {
			return 0, renderCtx.Err()
		}
		return 0, err
	}
	info, err := os.Stat(output)
	if err != nil || info.Size() <= 0 {
		return 0, fmt.Errorf("代理文件为空")
	}
	file, err := os.Open(output)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	proxy := aicutProxyObjectAsset(asset)
	proxy.ByteSize = info.Size()
	if err := objects.Put(ctx, proxy, file); err != nil {
		return 0, fmt.Errorf("保存代理文件失败")
	}
	return info.Size(), nil
}

func aicutProxyAvailable(ctx context.Context, asset aicut.Asset, objects agentworkspace.AssetObjectStore) bool {
	body, err := objects.Open(ctx, aicutProxyObjectAsset(asset))
	if err != nil {
		return false
	}
	_ = body.Close()
	return true
}

func aicutProxyObjectAsset(asset aicut.Asset) agentworkspace.Asset {
	return agentworkspace.Asset{
		ID: asset.ID, ProjectID: asset.ProjectID, Kind: "video",
		Name:        strings.TrimSuffix(asset.Name, path.Ext(asset.Name)) + "-代理.mp4",
		ContentType: "video/mp4", ObjectKey: asset.ObjectKey + ".proxy.mp4", CreatedAt: asset.CreatedAt,
	}
}

func detectedAICutAssetType(requestType, filename string, probe []byte) (string, string) {
	ext := strings.ToLower(path.Ext(filename))
	if ext == ".srt" {
		return "application/x-subrip", "subtitle"
	}
	if ext == ".ttf" && len(probe) >= 4 && (bytes.Equal(probe[:4], []byte{0, 1, 0, 0}) || string(probe[:4]) == "true") {
		return "font/ttf", "font"
	}
	if ext == ".otf" && len(probe) >= 4 && string(probe[:4]) == "OTTO" {
		return "font/otf", "font"
	}
	declared := strings.ToLower(strings.TrimSpace(strings.Split(requestType, ";")[0]))
	for _, contentType := range []string{http.DetectContentType(probe), declared, mime.TypeByExtension(ext)} {
		contentType = strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
		if strings.HasPrefix(contentType, "image/") {
			return contentType, "image"
		}
		if strings.HasPrefix(contentType, "video/") {
			return contentType, "video"
		}
		if strings.HasPrefix(contentType, "audio/") {
			return contentType, "audio"
		}
	}
	return "", ""
}

func aicutFontFamilyForAsset(ctx context.Context, asset aicut.Asset, objects agentworkspace.AssetObjectStore) (string, error) {
	body, err := objects.Open(ctx, aicutObjectAsset(asset))
	if err != nil {
		return "", err
	}
	defer body.Close()
	data, err := io.ReadAll(io.LimitReader(body, aicutFontUploadMaxBytes+1))
	if err != nil || int64(len(data)) > aicutFontUploadMaxBytes {
		return "", fmt.Errorf("字体文件读取失败")
	}
	return parseAICutFontFamily(data)
}

func aicutObjectAsset(asset aicut.Asset) agentworkspace.Asset {
	return agentworkspace.Asset{ID: asset.ID, ProjectID: asset.ProjectID, Kind: asset.Kind, Name: asset.Name, ContentType: asset.ContentType, ObjectKey: asset.ObjectKey, ByteSize: asset.ByteSize, CreatedAt: asset.CreatedAt}
}

func aicutAssetContentURL(projectID, assetID string) string {
	return "/api/ai-cut/projects/" + projectID + "/assets/" + assetID + "/content"
}

func aicutAssetProxyContentURL(projectID, assetID string) string {
	return "/api/ai-cut/projects/" + projectID + "/assets/" + assetID + "/proxy/content"
}
