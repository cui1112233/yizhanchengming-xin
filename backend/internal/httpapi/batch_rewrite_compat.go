package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"sort"
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
	root.Handle("POST /api/batch-rewrite/config", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		saveNovelFetchConfig(w, req, store, func(merged map[string]any) map[string]any {
			return map[string]any{"ok": true, "config": legacyBatchRewriteConfig(merged, false)}
		})
	})))
	issuesHandler := middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		entries, err := legacyBatchRewriteIssues(req.Context(), store, identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "读取问题日志失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"entries": entries})
	}))
	root.Handle("GET /api/batch-rewrite/issues", issuesHandler)
	// Older novel-fetch pages used the workshop namespace for this same log.
	// Keep the alias on Go so those pages do not fall back to the old Node app.
	root.Handle("GET /api/novel-fetch-workshop/issues", issuesHandler)
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
			tasks = append(tasks, legacyBatchRewriteTask(document))
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
		document, err = hydrateNovelFetchDocument(req.Context(), store, identity.Username, document)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "读取正文失败"})
			return
		}
		meta := legacyBatchRewriteTask(document)
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
	diagnostic := middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
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
		// The legacy workbench reads snake_case fields from diagnostic payloads,
		// while Go documents keep the original camelCase metadata. Reuse the
		// same adapter as the task table so the detail view cannot silently show
		// an all-empty card.
		meta := legacyBatchRewriteTask(document)
		switch {
		case strings.HasSuffix(req.URL.Path, "/sensitive-log"):
			hits := toInt(meta["sensitive_hit_count"])
			fixed := toInt(meta["sensitive_fixed_count"])
			hitsData := mapOrEmpty(meta["sensitiveHits"])
			fixedData := mapOrEmpty(meta["sensitiveFixed"])
			if len(hitsData) == 0 {
				hitsData = map[string]any{"mode": meta["sensitive_mode"], "status": meta["sensitive_status"], "hit_count": hits, "updated_at": meta["updated_at"], "hits": []any{}}
			}
			if len(fixedData) == 0 {
				fixedData = map[string]any{"mode": meta["sensitive_mode"], "status": meta["sensitive_status"], "fixed_count": fixed, "updated_at": meta["updated_at"], "items": []any{}}
			}
			writeJSON(w, http.StatusOK, map[string]any{
				"meta":            meta,
				"sensitive_hits":  hitsData,
				"sensitive_fixed": fixedData,
				"logs":            document.Logs,
			})
		case strings.HasSuffix(req.URL.Path, "/site-submit-log"):
			logs := []any{}
			if raw, ok := meta["websiteSubmitHistory"].([]any); ok {
				logs = raw
			}
			var result any = map[string]any{}
			if len(logs) > 0 {
				result = logs[len(logs)-1]
			}
			writeJSON(w, http.StatusOK, map[string]any{"meta": meta, "result": result, "logs": logs})
		default:
			writeJSON(w, http.StatusOK, map[string]any{"meta": meta, "stages": []any{}})
		}
	}))
	root.Handle("GET /api/batch-rewrite/tasks/{id}/sensitive-log", diagnostic)
	root.Handle("GET /api/batch-rewrite/tasks/{id}/rules-trace", diagnostic)
	root.Handle("GET /api/batch-rewrite/tasks/{id}/site-submit-log", diagnostic)
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
		webSubmit := mapOrEmpty(appConfig["web_submit"])
		// 小说获取和批量工厂使用同一份目标站会话。小说获取先登录时，
		// 批量工厂也要直接显示这个账号，而不是要求再次录入密码。
		if strings.TrimSpace(toString(webSubmit["username"])) == "" {
			session := readUploadTargetSession(usersDir, identity.Username)
			if targetUsername := strings.TrimSpace(toString(session["targetUsername"])); targetUsername != "" {
				webSubmit["username"] = targetUsername
				webSubmit["password_masked"] = true
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{"settings": webSubmit})
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
		settingsInput := mapOrEmpty(body.Settings)
		if password := strings.TrimSpace(toString(settingsInput["password"])); password != "" {
			username := strings.TrimSpace(toString(settingsInput["username"]))
			if username == "" {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请输入账号和密码"})
				return
			}
			cookie, loginErr := loginTargetSite(req, username, password)
			if loginErr != nil {
				writeJSON(w, http.StatusBadGateway, map[string]string{"error": loginErr.Error()})
				return
			}
			if err := saveUploadTargetSession(usersDir, identity.Username, map[string]any{"mode": "http_cookie", "cookie": cookie, "sessionKey": cookie, "targetUsername": username, "baseUrl": "http://two.121w.com/tttadmin", "status": "ready", "loginAt": time.Now().UTC().Format(time.RFC3339)}); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "上传会话保存失败"})
				return
			}
			body.Settings["password_masked"] = true
			delete(body.Settings, "password")
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
		cookie := strings.TrimSpace(toString(session["cookie"]))
		loggedIn := false
		detail := "未登录"
		if cookie != "" {
			check, requestErr := http.NewRequestWithContext(req.Context(), http.MethodGet, "http://two.121w.com/tttadmin/zidingyi.php", nil)
			if requestErr != nil {
				detail = "小说获取共享会话检查失败"
			} else {
				check.Header.Set("Cookie", cookie)
				response, probeErr := (&http.Client{Timeout: 30 * time.Second}).Do(check)
				if probeErr != nil {
					detail = "小说获取共享会话连接失败"
				} else {
					body, readErr := io.ReadAll(io.LimitReader(response.Body, 1<<20))
					response.Body.Close()
					if readErr != nil || response.StatusCode < 200 || response.StatusCode >= 300 {
						detail = "小说获取共享会话检查失败"
					} else if !targetPageLooksAuthenticated(string(body)) {
						detail = "共享登录会话已失效，请先到小说获取重新登录"
					} else {
						loggedIn = true
						detail = "已登录（共享小说获取会话）"
					}
				}
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": loggedIn, "checks": []any{map[string]any{"name": "目标站登录会话", "ok": loggedIn, "detail": detail}, map[string]any{"name": "上传接口", "ok": true, "detail": "/tttadmin/api/zbooklist_upload.php"}}})
	})))
	root.Handle("GET /api/batch-rewrite/web-submit/history", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		documents, err := store.ListDocuments(req.Context(), identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "读取提交历史失败"})
			return
		}
		records := make([]map[string]any, 0)
		for _, document := range documents {
			bookName := firstText(document.Meta["bookName"], document.Meta["book_name"], document.Meta["title"])
			platformName := firstText(document.Meta["platformName"], document.Meta["platform_name"])
			for _, raw := range document.Logs {
				item, ok := raw.(map[string]any)
				if !ok || toString(item["event"]) != "site_submit" {
					continue
				}
				data := mapOrEmpty(item["data"])
				receipt := mapOrEmpty(item["remote_receipt"])
				remote := mapOrEmpty(receipt["remote_record"])
				status := toString(item["status"])
				result := firstText(item["error"], receipt["upload_error"])
				verified, _ := receipt["verified"].(bool)
				if result == "" && status == "submitted" && verified {
					result = fmt.Sprintf("视频管理系统已确认接收 %d 个文件", intValue(receipt["upload_success_count"]))
					if detail := toString(remote["detail"]); detail != "" {
						result += "；" + detail
					}
				} else if result == "" && status == "accepted_pending" {
					result = firstText(remote["detail"], "视频管理系统已接收，等待后台确认")
				}
				records = append(records, map[string]any{
					"book_id": document.BookID, "book_name": bookName, "platform_name": platformName,
					"version": item["version"], "status": status, "group_id": item["group_id"],
					"material_allocation": item["material_allocation"], "result": result,
					"error": firstText(item["error"], receipt["upload_error"]),
					"time":  firstText(item["time"], item["updated_at"]), "data": data,
				})
			}
		}
		sort.SliceStable(records, func(i, j int) bool { return toString(records[i]["time"]) > toString(records[j]["time"]) })
		if len(records) > 200 {
			records = records[:200]
		}
		groups := []any{}
		if settings, configErr := store.GetConfig(req.Context(), identity.Username); configErr == nil {
			groups = append(groups, arrayOrEmpty(settings["site_submit_groups"])...)
			app := mapOrEmpty(settings["appConfig"])
			groups = append(groups, arrayOrEmpty(app["site_submit_groups"])...)
		}
		sort.SliceStable(groups, func(i, j int) bool {
			return toString(mapOrEmpty(groups[i])["updated_at"]) > toString(mapOrEmpty(groups[j])["updated_at"])
		})
		if len(groups) > 50 {
			groups = groups[:50]
		}
		writeJSON(w, http.StatusOK, map[string]any{"records": records, "groups": groups})
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
		items, hasItems := targetConfigItems(payload)
		if !hasItems || len(items) == 0 {
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
			organization := firstText(data["organization_id"], data["organizationId"], data["organization"], data["org_id"], data["orgId"], source["organization_id"], source["organizationId"], source["organization"], source["org_id"], source["orgId"])
			if organization != "" {
				profile["organization"] = organization
			}
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

// targetConfigItems accepts the response shapes used by different releases of
// the target backend: data[], data.list[], result[] and items[].
func targetConfigItems(payload map[string]any) ([]any, bool) {
	var find func(any) ([]any, bool)
	find = func(value any) ([]any, bool) {
		if items, ok := value.([]any); ok {
			return items, true
		}
		object, ok := value.(map[string]any)
		if !ok {
			return nil, false
		}
		for _, key := range []string{"list", "rows", "records", "items", "configs", "data", "result"} {
			if items, found := find(object[key]); found {
				return items, true
			}
		}
		return nil, false
	}
	for _, key := range []string{"data", "result", "configs", "items"} {
		if items, ok := find(payload[key]); ok {
			return items, true
		}
	}
	return nil, false
}

// legacyBatchRewriteTask converts the Go document (camelCase fields and
// nested book metadata) to the flat snake_case shape used by the V78 table.
// Without this adapter the API had the data, but every visible column read an
// empty legacy field name.
func legacyBatchRewriteTask(document novelfetchworkshop.Document) map[string]any {
	task := map[string]any{}
	for key, value := range document.Meta {
		task[key] = value
	}
	platformID := firstText(task["platform_id"], task["platformId"])
	platformName := firstText(task["platform_name"], task["platformName"])
	if platformName == "" {
		if id, err := strconv.Atoi(platformID); err == nil {
			platformName = legacyNovelFetchPlatforms[id]
		}
	}
	bookName := firstText(task["book_name"], task["bookName"], task["title"])
	style := firstText(task["style"], task["style_name"], task["styleName"], document.BookInfo.Genre, document.BookInfo.Category)
	if style == "" {
		style = "待判定"
	}
	gender := firstText(task["gender"], strings.TrimSpace(document.Gender))
	sensitiveReplaced := toInt(firstText(task["sensitive_hit_count"], task["sensitiveReplaced"]))
	if sensitiveReplaced == 0 {
		sensitiveReplaced = toInt(task["sensitiveReplaced"])
	}
	originalStatus := firstText(task["original_status"], task["originalStatus"])
	aiStatus := firstText(task["ai_status"], task["aiStatus"])
	classifyStatus := firstText(task["classify_status"], task["classifyStatus"])
	if classifyStatus == "" && gender != "" {
		classifyStatus = "done"
	}
	task["id"] = document.BookID
	task["book_id"] = document.BookID
	task["book_name"] = bookName
	task["platform_id"] = platformID
	task["platform_name"] = platformName
	task["style"] = style
	task["gender"] = gender
	task["classify_status"] = classifyStatus
	task["original_status"] = originalStatus
	task["ai_status"] = aiStatus
	// A previous worker can leave aiStatus at ai_processing even though the
	// generated body has already been durably stored. The V78 processing view
	// consumes this compatibility shape, so derive completion from the stored
	// versions rather than presenting a stale in-flight state.
	generatedVersions := make([]string, 0, 5)
	for index := 1; index <= 5; index++ {
		versionID := fmt.Sprintf("ai%d", index)
		value, exists := document.Versions[versionID]
		if !exists || strings.TrimSpace(fmt.Sprint(value)) == "" || fmt.Sprint(value) == "<nil>" {
			continue
		}
		generatedVersions = append(generatedVersions, versionID)
	}
	if len(generatedVersions) > 0 {
		task["ai_files"] = generatedVersions
		task["selected_versions"] = generatedVersions
		task["ai_status"] = "ai_done"
	}
	task["original_chars"] = document.ProcessedChars
	task["raw_original_chars"] = document.OriginalRawChars
	task["sensitive_status"] = firstText(task["sensitive_status"], task["sensitiveStatus"])
	if task["sensitive_status"] == "" {
		task["sensitive_status"] = "done"
	}
	task["sensitive_mode"] = "系统规则"
	task["sensitive_hit_count"] = sensitiveReplaced
	task["sensitive_fixed_count"] = sensitiveReplaced
	task["sensitive_failed_count"] = toInt(task["sensitive_failed_count"])
	task["ai_generated_count"] = toInt(firstText(task["ai_generated_count"], task["aiGeneratedCount"]))
	if len(generatedVersions) > task["ai_generated_count"].(int) {
		task["ai_generated_count"] = len(generatedVersions)
	}
	// The legacy V78 table reads flat snake_case fields. Keep the real submit
	// receipt visible after the Go submit handler stores it in task metadata.
	submitStatus := firstText(task["site_submit_status"], task["siteSubmitStatus"])
	confirmedVersions := anyStrings(task["site_submit_confirmed_versions"])
	if len(confirmedVersions) == 0 {
		confirmedVersions = anyStrings(task["siteSubmitDoneVersions"])
	}
	failedVersions := anyStrings(task["site_submit_failed_versions"])
	if len(failedVersions) == 0 {
		failedVersions = anyStrings(task["siteSubmitFailedVersions"])
	}
	task["site_submit_status"] = submitStatus
	task["siteSubmitStatus"] = submitStatus
	task["site_submit_confirmed_versions"] = confirmedVersions
	task["siteSubmitConfirmedVersions"] = confirmedVersions
	task["site_submit_failed_versions"] = failedVersions
	task["siteSubmitFailedVersions"] = failedVersions
	createdAt := firstText(task["created_at"], task["createdAt"], task["created_time"], task["createdTime"])
	if createdAt == "" {
		// Older records did not persist a separate creation field. UpdatedAt is
		// still a real task timestamp and is preferable to showing a blank cell.
		createdAt = document.UpdatedAt
	}
	task["created_at"] = createdAt
	task["createdAt"] = createdAt
	task["updated_at"] = firstText(task["updated_at"], task["updatedAt"], document.UpdatedAt)
	if document.MaxTxt > 0 {
		task["max_txt"] = document.MaxTxt
	}
	return task
}

// legacyBatchRewriteIssues keeps the old issue-log response shape while using
// the Go workshop documents as the source of truth. It deliberately mirrors
// the legacy Node filtering rules: only failed/error events and partial runs
// become visible issues.
func legacyBatchRewriteIssues(ctx context.Context, store novelfetchworkshop.Store, owner string) ([]map[string]any, error) {
	documents, err := store.ListDocuments(ctx, owner)
	if err != nil {
		return nil, err
	}
	entries := make([]map[string]any, 0)
	for _, document := range documents {
		taskName := firstText(document.Meta["bookName"], document.Meta["book_name"], document.Meta["title"])
		if taskError := firstText(document.Meta["ai_error"], document.Meta["aiError"], document.Meta["error"]); taskError != "" {
			event := "task_failed"
			if firstText(document.Meta["ai_error"], document.Meta["aiError"]) != "" {
				event = "ai_generate_failed"
			}
			entries = append(entries, batchRewriteIssueEntry(document.BookID, taskName, event, document.UpdatedAt, map[string]any{"error": taskError, "status": "failed"}, 0))
		}
		for index, raw := range document.Logs {
			item, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			if entry, ok := batchRewriteIssueEntryFromLog(document.BookID, taskName, item, index); ok {
				entries = append(entries, entry)
			}
		}
	}
	sort.SliceStable(entries, func(i, j int) bool {
		return toString(entries[i]["at"]) > toString(entries[j]["at"])
	})
	if len(entries) > 200 {
		entries = entries[:200]
	}
	return entries, nil
}

func batchRewriteIssueEntryFromLog(bookID, bookName string, item map[string]any, index int) (map[string]any, bool) {
	event := toString(item["event"])
	if event == "" {
		event = "task_error"
	}
	data := mapOrEmpty(item["data"])
	status := firstText(data["status"], item["status"])
	detail := firstText(data["error"], data["upload_error"], item["error"])
	failedCount := toInt(data["failed_count"])
	if !regexp.MustCompile(`(?i)(failed|error|失败|错误)`).MatchString(event+" "+status) && failedCount <= 0 && status != "partial" {
		return nil, false
	}
	return batchRewriteIssueEntry(bookID, bookName, event, firstText(item["time"], item["updated_at"]), map[string]any{"error": detail, "status": status, "items": data["items"]}, index), true
}

func toInt(value any) int {
	switch v := value.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case json.Number:
		n, _ := strconv.Atoi(string(v))
		return n
	default:
		n, _ := strconv.Atoi(strings.TrimSpace(toString(value)))
		return n
	}
}

func batchRewriteIssueEntry(bookID, bookName, event, at string, data map[string]any, index int) map[string]any {
	detail := firstText(data["error"], data["upload_error"])
	kind := "batch-rewrite.processing"
	label := "原文处理"
	if event == "site_submit" {
		kind, label = "batch-rewrite.submit", "网站提交"
	} else if strings.Contains(event, "sensitive") {
		kind, label = "batch-rewrite.sensitive", "敏感词处理"
	} else if strings.Contains(event, "ai") {
		kind, label = "batch-rewrite.ai", "AI 文案处理"
	}
	if detail == "" {
		detail = "书籍 " + bookID + " 的" + label + "失败"
	}
	if at == "" {
		at = time.Unix(0, 0).UTC().Format(time.RFC3339)
	}
	entry := map[string]any{
		"id":   fmt.Sprintf("batch-rewrite:%s:%s:%s", bookID, event, firstText(at, strconv.Itoa(index))),
		"kind": kind,
		"message": fmt.Sprintf("书籍 %s%s：%s", bookID, func() string {
			if bookName != "" {
				return "《" + bookName + "》"
			}
			return ""
		}(), detail),
		"at":     at,
		"path":   "/novel-fetch",
		"method": event,
	}
	if strings.Contains(detail, "未登录") || strings.Contains(detail, "会话") || strings.Contains(detail, "认证") {
		entry["status"] = 401
	}
	return entry
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
	// The legacy workbench reads the catalog from appConfig.  Older saved
	// settings can contain only empty placeholders (or the newer snake_case
	// shape), so normalize them here before building the response.  Otherwise
	// the three knowledge bases render as present but all zero entries.
	settings = novelfetchworkshop.EnsureWorkshopDefaults(settings)
	appConfig := cloneAnyMap(settings)
	if configured, ok := settings["appConfig"].(map[string]any); ok {
		appConfig = cloneAnyMap(configured)
	}
	platforms := arrayOrEmpty(settings["platforms"])
	styles := arrayOrEmpty(settings["styles"])
	if len(platforms) == 0 {
		platforms = append([]any(nil), defaultNovelFetchPlatforms...)
	}
	if len(styles) == 0 {
		styles = append([]any(nil), defaultNovelFetchStyles...)
	}
	// 旧工作台会直接读取这些字段来初始化表单。Go 端不能只返回空壳，
	// 否则页面虽然能打开，但“文本模型/解析模式/工作流”都会回到空值。
	applyBatchRewriteDefaults(appConfig)
	if catalog, exists := settings["modelCatalog"]; exists {
		if raw, ok := appConfig["modelCatalog"].([]any); !ok || len(raw) == 0 {
			appConfig["modelCatalog"] = catalog
		}
	}
	if strings.TrimSpace(toString(appConfig["text_model_id"])) == "" {
		for _, key := range []string{"text_model_id", "textModelId", "selectedTextModelId"} {
			if value := strings.TrimSpace(toString(settings[key])); value != "" {
				appConfig["text_model_id"] = value
				break
			}
		}
	}
	knowledge := arrayOrEmptyMap(appConfig["knowledge"])
	if lightweight {
		knowledge = map[string]any{}
	} else if knowledgeCatalogEmpty(knowledge) {
		// Some migrated accounts have a complete-looking knowledge object whose
		// every list is empty.  Keep the bundled V78 catalog visible in that
		// case; an empty placeholder is not a deliberate deletion.
		defaults := novelfetchworkshop.DefaultWorkshopSettings()
		defaultApp := mapOrEmpty(defaults["appConfig"])
		knowledge = mapOrEmpty(defaultApp["knowledge"])
	}
	workForm := mapOrEmpty(appConfig["work_form"])
	if len(workForm) == 0 {
		workForm = mapOrEmpty(settings["work_form"])
	}
	// The legacy page renders the top-level `sensitive` field, while the
	// migrated Go config keeps the same groups under `appConfig.sensitive`.
	// Expose the canonical app config here so the page cannot manufacture an
	// empty "默认敏感词组" while the worker uses the real rules.
	sensitive := mapOrEmpty(settings["sensitive"])
	if configured := mapOrEmpty(appConfig["sensitive"]); len(configured) > 0 {
		sensitive = configured
	} else if len(sensitive) > 0 {
		appConfig["sensitive"] = sensitive
	}
	return map[string]any{
		"app_config":        publicBatchAppConfig(appConfig),
		"platforms":         platforms,
		"styles":            styles,
		"sensitive":         sensitive,
		"web_submit":        publicBatchWebSubmit(appConfig["web_submit"]),
		"knowledge":         knowledge,
		"knowledge_loaded":  !lightweight,
		"knowledge_summary": batchKnowledgeSummary(knowledge, lightweight),
		"parse_modes":       batchRewriteParseModes,
		"column_presets":    batchRewriteColumnPresets,
		"work_form":         workForm,
	}
}

var batchRewriteParseModes = []any{
	map[string]any{"id": "smart", "name": "smart"},
	map[string]any{"id": "header", "name": "header"},
	map[string]any{"id": "multi_header", "name": "multi_header"},
	map[string]any{"id": "fixed_full_11", "name": "fixed_full_11"},
	map[string]any{"id": "fixed_from_b", "name": "fixed_from_b"},
	map[string]any{"id": "fixed_paid_basic", "name": "fixed_paid_basic"},
	map[string]any{"id": "custom", "name": "custom"},
}

var batchRewriteColumnPresets = []any{
	map[string]any{"id": "paid_name_reason", "name": "付费ID/书名/推荐理由", "columns": []any{"paid_book_id", "book_name", "reason"}},
	map[string]any{"id": "paid_name_gender_reason", "name": "付费ID/书名/男女频/推荐理由", "columns": []any{"paid_book_id", "book_name", "gender", "reason"}},
	map[string]any{"id": "free_paid_name_gender_reason", "name": "免费ID/付费ID/书名/男女频/推荐理由", "columns": []any{"free_book_id", "paid_book_id", "book_name", "gender", "reason"}},
	map[string]any{"id": "sample_input", "name": "书籍ID/书名/推荐理由/男女频/标签/评级", "columns": []any{"book_id", "book_name", "reason", "gender", "tags", "rating"}},
	map[string]any{"id": "full_11", "name": "完整11列", "columns": []any{"ignore", "free_book_id", "paid_book_id", "book_name", "rating", "ignore", "ignore", "ignore", "ignore", "gender", "reason"}},
}

func applyBatchRewriteDefaults(app map[string]any) {
	defaults := map[string]any{
		"workflow":   map[string]any{"auto_classify_missing": true, "auto_fetch_original": true, "auto_rewrite_after_fetch": true, "auto_submit_after_rewrite": false, "auto_submit_confirmed": false},
		"fetch":      map[string]any{"endpoint": "https://txt.121w.com/api.php", "default_max_txt": 4000, "timeout_seconds": 30, "concurrency": 4, "retries": 1, "auto_detect_platform": false},
		"ai":         map[string]any{"base_url": "", "api_key": "", "model": "", "timeout_seconds": 180, "max_concurrency": 6, "retry_times": 2, "max_tokens": 1200, "temperature": 0.45, "top_p": 0.9, "stream": false, "json_mode": false, "enable_thinking": false, "disable_thinking": false, "extra_body_json": ""},
		"ai_presets": []any{}, "ai_assignments": map[string]any{"classifier": "__current__", "rewrite": "__current__", "sensitive_fix": "__current__"},
		"rewrite":    map[string]any{"default_ai_count": 1, "max_ai_count": 5, "process_line_count": 5, "anchor_line_count": 5, "temperature": 0.45, "strategy": "instruction", "method_sequence": []any{"high_imitation", "opening_instruction", "instruction"}, "prompt": "你是短视频小说正文改写师。请只根据原文、风格类型和男女频改写开头部分，保留故事事实、人物关系、时代背景和后续衔接。输出正文，不要解释。"},
		"storage":    map[string]any{"cleanup_enabled": true, "retention_days": 7},
		"parser":     map[string]any{"default_parse_mode": "smart", "default_column_preset_id": "sample_input", "custom_column_order": "书籍ID,书名,推荐理由,男女频,标签,评级"},
		"web_submit": map[string]any{"enabled": false, "username": "", "password_masked": false, "skip_submitted": true, "min_text_chars": 0, "retry_times": 1, "submit_mode": "version", "selected_profile": "", "advanced": map[string]any{}, "upload_profiles": []any{}, "profile_bindings": map[string]any{}, "style_catalog": []any{}, "target": "two.121w.com"},
	}
	for key, value := range defaults {
		if _, exists := app[key]; !exists || app[key] == nil {
			app[key] = value
		} else if dm, ok := value.(map[string]any); ok {
			if current, ok := app[key].(map[string]any); ok {
				for child, fallback := range dm {
					if _, exists := current[child]; !exists || current[child] == nil {
						current[child] = fallback
					}
				}
			} else {
				app[key] = dm
			}
		}
	}
	if _, ok := app["modelCatalog"]; !ok {
		app["modelCatalog"] = []any{}
	}
	if _, ok := app["text_model_id"]; !ok {
		app["text_model_id"] = ""
	}
}

func publicBatchAppConfig(app map[string]any) map[string]any {
	out := cloneAnyMap(app)
	if raw, ok := app["modelCatalog"].([]any); ok {
		models := make([]any, 0, len(raw))
		for _, item := range raw {
			if model, ok := item.(map[string]any); ok {
				models = append(models, publicManagedModel(model))
			}
		}
		out["modelCatalog"] = models
	}
	if raw, ok := app["modelCatalog"].([]map[string]any); ok {
		models := make([]any, 0, len(raw))
		for _, model := range raw {
			models = append(models, publicManagedModel(model))
		}
		out["modelCatalog"] = models
	}
	if raw, ok := app["ai"].(map[string]any); ok {
		ai := cloneAnyMap(raw)
		delete(ai, "api_key")
		out["ai"] = ai
	}
	return out
}

func publicBatchWebSubmit(value any) map[string]any {
	cfg := mapOrEmpty(value)
	out := map[string]any{"enabled": cfg["enabled"] == true, "username": firstText(cfg["username"]), "password": "", "password_masked": cfg["password_masked"] == true, "skip_submitted": cfg["skip_submitted"] != false, "min_text_chars": targetNumber(cfg["min_text_chars"]), "retry_times": targetNumber(cfg["retry_times"]), "submit_mode": "version", "selected_profile": firstText(cfg["selected_profile"]), "advanced": mapOrEmpty(cfg["advanced"]), "upload_profiles": arrayOrEmpty(cfg["upload_profiles"]), "profile_bindings": mapOrEmpty(cfg["profile_bindings"]), "style_catalog": arrayOrEmpty(cfg["style_catalog"]), "target": "two.121w.com"}
	if out["min_text_chars"].(float64) < 0 {
		out["min_text_chars"] = float64(0)
	}
	if out["retry_times"].(float64) < 1 {
		out["retry_times"] = float64(1)
	}
	return out
}

func batchKnowledgeSummary(knowledge map[string]any, lightweight bool) map[string]any {
	if lightweight {
		return map[string]any{}
	}
	count := func(section, key string) int { return len(arrayOrEmpty(mapOrEmpty(knowledge[section])[key])) }
	return map[string]any{"opening_phrases": count("opening_phrases", "items"), "opening_styles": count("opening_phrases", "styles"), "rewrite_profiles": count("rewrite_templates", "profiles"), "temporary_instructions": count("rewrite_templates", "temporary_instructions"), "high_imitation_prompts": count("high_imitation", "prompts"), "high_imitation_references": count("high_imitation", "references"), "sensitive_rules": count("layout_rules", "sensitive_rules"), "symbol_rules": count("symbol_rules", "symbol_rules"), "pair_fill_rules": count("symbol_rules", "pair_fill_rules"), "chapter_exact_rules": count("chapter_rules", "chapter_exact_rules"), "chapter_inline_rules": count("chapter_rules", "chapter_inline_rules")}
}

func knowledgeCatalogEmpty(knowledge map[string]any) bool {
	if len(knowledge) == 0 {
		return true
	}
	for _, section := range []string{"opening_phrases", "high_imitation", "rewrite_templates"} {
		value := mapOrEmpty(knowledge[section])
		for _, key := range []string{"items", "styles", "prompts", "references", "profiles", "temporary_instructions"} {
			if len(arrayOrEmpty(value[key])) > 0 {
				return false
			}
		}
	}
	return true
}

func cloneAnyMap(input map[string]any) map[string]any {
	out := map[string]any{}
	for key, value := range input {
		out[key] = value
	}
	return out
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
