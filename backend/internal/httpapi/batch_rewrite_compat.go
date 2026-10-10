package httpapi

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

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
	root.Handle("POST /api/batch-rewrite/web-submit/sync-styles", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var body struct {
			Persist *bool `json:"persist"`
		}
		_ = json.NewDecoder(http.MaxBytesReader(w, req.Body, 16<<10)).Decode(&body)
		session := readUploadTargetSession(usersDir, identity.Username)
		cookie := strings.TrimSpace(toString(session["cookie"]))
		if cookie == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请先保存账号密码并登录目标站"})
			return
		}
		response, err := targetGET(req, "http://two.121w.com/tttadmin/zidingyi.php", cookie)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "目标站风格目录读取失败"})
			return
		}
		styles := parseTargetStyles(response)
		if len(styles) == 0 {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "目标站未返回可用风格目录"})
			return
		}
		settings, err := store.GetConfig(req.Context(), identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "读取提交配置失败"})
			return
		}
		app := mapOrEmpty(settings["appConfig"])
		web := mapOrEmpty(app["web_submit"])
		old := stringArray(settings["styles"])
		web["style_catalog"] = styles
		app["web_submit"] = web
		settings["appConfig"] = app
		if body.Persist == nil || *body.Persist {
			settings["styles"] = styleNames(styles)
			if err := store.PutConfig(req.Context(), identity.Username, settings); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "保存风格目录失败"})
				return
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "settings": web, "styles": styleNames(styles), "style_sync": map[string]any{"old_count": len(old), "new_count": len(styles), "added": styleNames(styles), "removed": []any{}}, "output": []string{"已从视频管理系统同步风格目录"}})
	})))
	root.Handle("POST /api/batch-rewrite/web-submit/sync-configs", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var body struct {
			Persist *bool `json:"persist"`
		}
		_ = json.NewDecoder(http.MaxBytesReader(w, req.Body, 16<<10)).Decode(&body)
		session := readUploadTargetSession(usersDir, identity.Username)
		cookie := strings.TrimSpace(toString(session["cookie"]))
		if cookie == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请先保存账号密码并登录目标站"})
			return
		}
		response, err := targetGET(req, "http://two.121w.com/tttadmin/api/zdy_config.php?action=list", cookie)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "目标站配置档读取失败"})
			return
		}
		var payload map[string]any
		if json.Unmarshal(response, &payload) != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "目标站配置档返回格式无效"})
			return
		}
		items, _ := payload["data"].([]any)
		if len(items) == 0 {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "目标站未返回可用配置档"})
			return
		}
		settings, err := store.GetConfig(req.Context(), identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "读取提交配置失败"})
			return
		}
		app := mapOrEmpty(settings["appConfig"])
		web := mapOrEmpty(app["web_submit"])
		existing := mapOrEmpty(web)
		local := []any{}
		for _, item := range arrayOrEmpty(existing["upload_profiles"]) {
			if strings.ToLower(toString(mapOrEmpty(item)["source"])) != "121" {
				local = append(local, item)
			}
		}
		groups := make([]any, 0, len(items))
		for i, item := range items {
			source := mapOrEmpty(item)
			data := parseObject(source["config_data"])
			id := toString(source["id"])
			if id == "" {
				id = fmt.Sprintf("%d", i+1)
			}
			gender := toString(data["gender"])
			if gender == "1" {
				gender = "男"
			} else if gender == "2" {
				gender = "女"
			}
			profile := map[string]any{"id": "121-" + id, "name": firstText(source["config_name"], "121 配置档 "+id), "enabled": true, "source": "121", "config_id": id, "is_default": targetNumber(source["is_default"]) == 1, "platform_id": firstText(data["platform_id"], source["platform_id"]), "gender": strings.ReplaceAll(gender, "频", ""), "style": firstText(data["style_name"], source["style_name"], data["style"]), "advanced": data}
			groups = append(groups, profile)
		}
		web["upload_profiles"] = append(local, groups...)
		app["web_submit"] = web
		settings["appConfig"] = app
		if body.Persist == nil || *body.Persist {
			if err := store.PutConfig(req.Context(), identity.Username, settings); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "保存配置档失败"})
				return
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "settings": web, "groups": groups, "output": []string{fmt.Sprintf("已从视频管理系统同步 %d 个配置档", len(groups))}})
	})))
}

func targetGET(req *http.Request, url, cookie string) ([]byte, error) {
	check, err := http.NewRequestWithContext(req.Context(), http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	check.Header.Set("Cookie", cookie)
	response, err := (&http.Client{Timeout: 30 * time.Second}).Do(check)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, 4<<20))
	if err != nil || response.StatusCode < 200 || response.StatusCode >= 300 || strings.Contains(string(body), "管理员登录") {
		return nil, fmt.Errorf("target request failed")
	}
	return body, nil
}
func parseTargetStyles(body []byte) []map[string]any {
	re := regexp.MustCompile(`(?is)<select[^>]*\bid=["']style["'][^>]*>(.*?)</select>`)
	match := re.FindSubmatch(body)
	if len(match) < 2 {
		return nil
	}
	option := regexp.MustCompile(`(?is)<option[^>]*value=["']([^"']+)["'][^>]*>(.*?)</option>`)
	seen := map[string]bool{}
	out := []map[string]any{}
	for _, item := range option.FindAllSubmatch(match[1], -1) {
		id, name := strings.TrimSpace(string(item[1])), strings.TrimSpace(regexp.MustCompile(`<[^>]+>`).ReplaceAllString(string(item[2]), ""))
		name = html.UnescapeString(name)
		if id != "" && name != "" && !strings.Contains(name, "请选择") && !seen[id] {
			seen[id] = true
			out = append(out, map[string]any{"id": id, "name": name})
		}
	}
	return out
}
func styleNames(styles []map[string]any) []string {
	out := make([]string, 0, len(styles))
	for _, item := range styles {
		out = append(out, toString(item["name"]))
	}
	return out
}
func stringArray(value any) []string {
	out := []string{}
	for _, item := range arrayOrEmpty(value) {
		out = append(out, toString(item))
	}
	return out
}
func parseObject(value any) map[string]any {
	if object, ok := value.(map[string]any); ok {
		return object
	}
	var parsed map[string]any
	if json.Unmarshal([]byte(toString(value)), &parsed) == nil && parsed != nil {
		return parsed
	}
	return map[string]any{}
}
func firstText(values ...any) string {
	for _, value := range values {
		if text := strings.TrimSpace(toString(value)); text != "" {
			return text
		}
	}
	return ""
}
func targetNumber(value any) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case json.Number:
		n, _ := v.Float64()
		return n
	}
	n, _ := strconv.ParseFloat(strings.TrimSpace(toString(value)), 64)
	return n
}

func legacyBatchRewriteConfig(settings map[string]any, lightweight bool) map[string]any {
	appConfig := settings
	if configured, ok := settings["appConfig"].(map[string]any); ok {
		appConfig = configured
	}
	platforms := arrayOrEmpty(settings["platforms"])
	styles := arrayOrEmpty(settings["styles"])
	if len(platforms) == 0 {
		platforms = append([]any(nil), defaultNovelFetchPlatforms...)
	}
	if len(styles) == 0 {
		styles = append([]any(nil), defaultNovelFetchStyles...)
	}
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
