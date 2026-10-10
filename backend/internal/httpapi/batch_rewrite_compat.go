package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"qiantie/backend/internal/novelfetchworkshop"
)

// RegisterBatchRewriteConfigCompat exposes the read-only configuration shape
// expected by the legacy novel-fetch workbench. The actual settings are still
// stored in the Go workshop store, so the page no longer needs Node just to
// bootstrap its form.
func RegisterBatchRewriteConfigCompat(root *http.ServeMux, middleware BridgeAuth, store novelfetchworkshop.Store, usersDir string) {
	if root == nil || store == nil {
		return
	}
	read := func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		settings, err := store.GetConfig(req.Context(), identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "读取设置失败"})
			return
		}
		writeJSON(w, http.StatusOK, legacyBatchRewriteConfig(settings, req.URL.Path == "/api/batch-rewrite/bootstrap"))
	}
	handler := middleware.Middleware(http.HandlerFunc(read))
	root.Handle("GET /api/batch-rewrite/config", handler)
	root.Handle("GET /api/batch-rewrite/bootstrap", handler)
	root.Handle("GET /api/batch-rewrite/tasks", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		documents, err := store.ListDocuments(req.Context(), identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "读取任务失败"})
			return
		}
		tasks := make([]map[string]any, 0, len(documents))
		for _, document := range documents {
			task := map[string]any{}
			for key, value := range document.Meta {
				task[key] = value
			}
			task["id"] = document.BookID
			task["book_id"] = document.BookID
			if _, ok := task["updated_at"]; !ok {
				task["updated_at"] = document.UpdatedAt
			}
			if _, ok := task["original_chars"]; !ok {
				task["original_chars"] = document.ProcessedChars
			}
			tasks = append(tasks, task)
		}
		writeJSON(w, http.StatusOK, map[string]any{"tasks": tasks})
	})))
	root.Handle("GET /api/batch-rewrite/tasks/{id}", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		bookID := req.PathValue("id")
		document, err := store.GetDocument(req.Context(), identity.Username, bookID)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "任务不存在"})
			return
		}
		meta := map[string]any{}
		for key, value := range document.Meta {
			meta[key] = value
		}
		meta["id"] = document.BookID
		meta["book_id"] = document.BookID
		meta["updated_at"] = document.UpdatedAt
		if document.MaxTxt > 0 {
			meta["max_txt"] = document.MaxTxt
		}
		aiTexts := make([]map[string]any, 0, len(document.Versions))
		for version, raw := range document.Versions {
			item := map[string]any{"name": version, "version": version}
			if value, ok := raw.(map[string]any); ok {
				for key, field := range value {
					item[key] = field
				}
			} else if text, ok := raw.(string); ok {
				item["text"] = text
			}
			aiTexts = append(aiTexts, item)
		}
		original := document.Original
		if original == "" {
			original = document.ProcessedText
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"meta":             meta,
			"original":         original,
			"original_raw":     document.OriginalRaw,
			"ai_texts":         aiTexts,
			"has_original_raw": document.OriginalRaw != "",
			"logs":             document.Logs,
		})
	})))
	root.Handle("POST /api/batch-rewrite/tasks/batch-delete", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var body struct {
			IDs []string `json:"ids"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, req.Body, 64<<10)).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求内容无效"})
			return
		}
		ids := make([]string, 0, len(body.IDs))
		seen := map[string]bool{}
		for _, raw := range body.IDs {
			id := strings.TrimSpace(raw)
			if id != "" && !seen[id] {
				seen[id] = true
				ids = append(ids, id)
			}
		}
		result, err := store.DeleteDocuments(req.Context(), identity.Username, ids)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "删除任务失败"})
			return
		}
		documents, _ := store.ListDocuments(req.Context(), identity.Username)
		writeJSON(w, http.StatusOK, map[string]any{"requested": result.Requested, "deleted": result.Deleted, "failed": 0, "tasks": documents})
	})))
	root.Handle("POST /api/batch-rewrite/work-form", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var body struct {
			State map[string]any `json:"state"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, req.Body, 128<<10)).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求内容无效"})
			return
		}
		settings, err := store.GetConfig(req.Context(), identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "读取设置失败"})
			return
		}
		if body.State == nil {
			body.State = map[string]any{}
		}
		appConfig, ok := settings["appConfig"].(map[string]any)
		if !ok {
			appConfig = map[string]any{}
		}
		appConfig["work_form"] = body.State
		settings["appConfig"] = appConfig
		if err := store.PutConfig(req.Context(), identity.Username, settings); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "保存设置失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "state": body.State})
	})))
	root.Handle("GET /api/batch-rewrite/web-submit/config", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		settings, err := store.GetConfig(req.Context(), identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "读取提交配置失败"})
			return
		}
		appConfig := mapOrEmpty(settings["appConfig"])
		writeJSON(w, http.StatusOK, map[string]any{"settings": mapOrEmpty(appConfig["web_submit"])})
	})))
	root.Handle("POST /api/batch-rewrite/web-submit/config", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var body struct {
			Settings map[string]any `json:"settings"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, req.Body, 128<<10)).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求内容无效"})
			return
		}
		settings, err := store.GetConfig(req.Context(), identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "读取提交配置失败"})
			return
		}
		appConfig := mapOrEmpty(settings["appConfig"])
		appConfig["web_submit"] = body.Settings
		settings["appConfig"] = appConfig
		if err := store.PutConfig(req.Context(), identity.Username, settings); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "保存提交配置失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "settings": body.Settings, "tasks": []any{}})
	})))
	root.Handle("GET /api/batch-rewrite/web-submit/environment", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		session := readUploadTargetSession(usersDir, identity.Username)
		loggedIn := strings.TrimSpace(toString(session["cookie"])) != ""
		detail := "未登录"
		if loggedIn {
			detail = "已保存登录会话"
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": loggedIn, "checks": []any{map[string]any{"name": "目标站登录会话", "ok": loggedIn, "detail": detail}, map[string]any{"name": "上传接口", "ok": true, "detail": "/tttadmin/api/zbooklist_upload.php"}}})
	})))
}

func legacyBatchRewriteConfig(settings map[string]any, lightweight bool) map[string]any {
	appConfig := settings
	if configured, ok := settings["appConfig"].(map[string]any); ok {
		appConfig = configured
	}
	platforms := arrayOrEmpty(settings["platforms"])
	styles := arrayOrEmpty(settings["styles"])
	knowledge := arrayOrEmptyMap(appConfig["knowledge"])
	if lightweight {
		knowledge = map[string]any{}
	}
	workForm := mapOrEmpty(appConfig["work_form"])
	if len(workForm) == 0 {
		workForm = mapOrEmpty(settings["work_form"])
	}
	return map[string]any{
		"app_config":        appConfig,
		"platforms":         platforms,
		"styles":            styles,
		"sensitive":         mapOrEmpty(settings["sensitive"]),
		"web_submit":        mapOrEmpty(appConfig["web_submit"]),
		"knowledge":         knowledge,
		"knowledge_loaded":  !lightweight,
		"knowledge_summary": map[string]any{},
		"parse_modes": []any{
			map[string]any{"id": "smart", "name": "智能解析"},
			map[string]any{"id": "strict", "name": "严格解析"},
		},
		"column_presets": []any{},
		"work_form":      workForm,
	}
}

func arrayOrEmpty(value any) []any {
	if items, ok := value.([]any); ok {
		return items
	}
	return []any{}
}

func arrayOrEmptyMap(value any) map[string]any {
	if items, ok := value.(map[string]any); ok {
		return items
	}
	return map[string]any{}
}

func mapOrEmpty(value any) map[string]any {
	if items, ok := value.(map[string]any); ok {
		return items
	}
	return map[string]any{}
}
