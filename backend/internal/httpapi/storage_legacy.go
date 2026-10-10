package httpapi

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	storageScriptDir  = "剧本生成"
	storageNovelFetch = "小说获取"
	storageNovelAdapt = "改编小说"
	storageProjectDir = "制作工程"
)

func configuredStorageRoot(configs *UserConfigStore, username string) string {
	if configs == nil {
		return ""
	}
	config := configs.read(username)
	root, _ := config["storageRoot"].(string)
	return strings.TrimSpace(root)
}

func scanStorageFiles(dir string) []map[string]any {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return []map[string]any{}
	}
	result := make([]map[string]any, 0)
	for _, entry := range entries {
		if !entry.Type().IsRegular() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		result = append(result, map[string]any{"name": entry.Name(), "size": info.Size()})
	}
	return result
}

func scanConfiguredStorage(root string) map[string]any {
	projects := make([]map[string]any, 0)
	for _, entry := range scanSubdirectories(filepath.Join(root, storageProjectDir)) {
		projects = append(projects, map[string]any{
			"name":  entry,
			"files": scanStorageFiles(filepath.Join(root, storageProjectDir, entry)),
		})
	}
	return map[string]any{
		"scriptResults": scanStorageFiles(filepath.Join(root, storageScriptDir)),
		"novelFetch":    scanStorageFiles(filepath.Join(root, storageNovelFetch)),
		"novelAdapt":    scanStorageFiles(filepath.Join(root, storageNovelAdapt)),
		"projects":      projects,
	}
}

func scanSubdirectories(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return []string{}
	}
	result := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			result = append(result, entry.Name())
		}
	}
	return result
}

func emptyStorageReport() map[string]any {
	return map[string]any{
		"scriptResults": map[string]any{"found": 0, "added": 0},
		"novelFetch":    map[string]any{"found": 0},
		"novelAdapt":    map[string]any{"found": 0},
		"projects":      map[string]any{"found": 0},
		"errors":        []string{},
	}
}

func restoreConfiguredStorage(ctx context.Context, root, username string, history *ScriptHistoryStore) (map[string]any, error) {
	report := emptyStorageReport()
	if root == "" {
		return report, nil
	}
	scriptDir := filepath.Join(root, storageScriptDir)
	files := scanStorageFiles(scriptDir)
	found := 0
	for _, file := range files {
		name, _ := file["name"].(string)
		if strings.HasSuffix(strings.ToLower(name), ".md") {
			found++
		}
	}
	if history == nil {
		return nil, fmt.Errorf("history store is unavailable")
	}
	history.mu.Lock()
	defer history.mu.Unlock()
	data, dir, err := history.load(username)
	if err != nil {
		return nil, err
	}
	entries, _ := data["entries"].([]any)
	added := 0
	errors := []string{}
	for _, file := range files {
		name, _ := file["name"].(string)
		if !strings.HasSuffix(strings.ToLower(name), ".md") {
			continue
		}
		id := strings.TrimSuffix(name, filepath.Ext(name))
		if !validScriptHistoryID(id) || findHistoryEntry(data, id) != nil {
			continue
		}
		raw, readErr := os.ReadFile(filepath.Join(scriptDir, name))
		if readErr != nil {
			errors = append(errors, storageScriptDir+"/"+name+": "+readErr.Error())
			continue
		}
		info, _ := os.Stat(filepath.Join(scriptDir, name))
		createdAt := any(nil)
		if info != nil {
			createdAt = info.ModTime().UnixMilli()
		}
		title := strings.TrimSpace(strings.SplitN(string(raw), "\n", 2)[0])
		title = strings.TrimSpace(strings.TrimPrefix(title, "#"))
		if title == "" {
			title = id
		}
		entry := map[string]any{"id": id, "filename": name, "title": title, "output": string(raw), "restoredFrom": "local", "createdAt": createdAt}
		entries = append([]any{entry}, entries...)
		added++
	}
	if added > 0 {
		if len(entries) > 50 {
			entries = entries[:50]
		}
		data["entries"] = entries
		if err := saveScriptHistory(dir, data); err != nil {
			return nil, err
		}
	}
	result := map[string]any{
		"scriptResults": map[string]any{"found": found, "added": added},
		"novelFetch":    map[string]any{"found": len(scanStorageFiles(filepath.Join(root, storageNovelFetch)))},
		"novelAdapt":    map[string]any{"found": len(scanStorageFiles(filepath.Join(root, storageNovelAdapt)))},
		"projects":      map[string]any{"found": len(scanSubdirectories(filepath.Join(root, storageProjectDir)))},
		"errors":        errors,
	}
	_ = ctx
	return result, nil
}

func RegisterStorageRoutes(root *http.ServeMux, middleware BridgeAuth, configs *UserConfigStore, history *ScriptHistoryStore) {
	if root == nil || configs == nil || history == nil {
		return
	}
	authHandler := func(next http.Handler) http.Handler {
		return middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if _, ok := BridgeIdentityFromContext(req.Context()); !ok {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}
			next.ServeHTTP(w, req)
		}))
	}
	root.Handle("GET /api/storage/list", authHandler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		storageRoot := configuredStorageRoot(configs, identity.Username)
		if storageRoot == "" {
			writeJSON(w, http.StatusOK, map[string]any{"scriptResults": []any{}, "novelFetch": []any{}, "novelAdapt": []any{}, "projects": []any{}})
			return
		}
		writeJSON(w, http.StatusOK, scanConfiguredStorage(storageRoot))
	})))
	root.Handle("POST /api/storage/restore", authHandler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		storageRoot := configuredStorageRoot(configs, identity.Username)
		result, err := restoreConfiguredStorage(req.Context(), storageRoot, identity.Username, history)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "恢复本地存储失败"})
			return
		}
		writeJSON(w, http.StatusOK, result)
	})))
	root.Handle("POST /api/storage/save-media", authHandler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		var input struct {
			ProjectName string `json:"projectName"`
			Category    string `json:"category"`
			Filename    string `json:"filename"`
			DataURL     string `json:"dataUrl"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 64<<20)).Decode(&input) != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求内容无效"})
			return
		}
		rootDir := configuredStorageRoot(configs, identity.Username)
		if rootDir == "" {
			writeJSON(w, http.StatusOK, map[string]any{"saved": false, "reason": "未配置本地存储文件夹"})
			return
		}
		categoryDir := map[string]string{"image": "图片", "video": "视频", "audio": "配音"}[input.Category]
		if categoryDir == "" || strings.TrimSpace(input.ProjectName) == "" || strings.TrimSpace(input.Filename) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "projectName/category/filename/dataUrl 均必填"})
			return
		}
		prefix, encoded, ok := strings.Cut(input.DataURL, ",")
		if !ok || !strings.HasPrefix(prefix, "data:") || !strings.HasSuffix(prefix, ";base64") {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "dataUrl 必须是 base64 data URL"})
			return
		}
		data, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "dataUrl 必须是 base64 data URL"})
			return
		}
		projectName := strings.Map(func(r rune) rune {
			if strings.ContainsRune(`\\/:*?"<>|`, r) {
				return '_'
			}
			return r
		}, strings.TrimSpace(input.ProjectName))
		if projectName == "" {
			projectName = "未命名项目"
		}
		dir := filepath.Join(rootDir, storageProjectDir, projectName, categoryDir)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "文件夹创建失败"})
			return
		}
		filename := filepath.Base(input.Filename)
		if err := os.WriteFile(filepath.Join(dir, filename), data, 0o600); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "文件保存失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"saved": true, "path": filepath.Join(dir, filename)})
	})))
}
