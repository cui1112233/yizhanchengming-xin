package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"qiantie/backend/internal/novelfetchworkshop"
)

var novelFetchOperations = struct {
	sync.RWMutex
	items map[string]map[string]any
}{items: map[string]map[string]any{}}

// registerNovelFetchLegacyWorkflowRoutes keeps the old workbench API names
// on the Go store. These routes are deliberately small adapters: they return
// the same shapes the workbench expects while preserving the complete raw
// text in each stored document.
func registerNovelFetchLegacyWorkflowRoutes(mux *http.ServeMux, store novelfetchworkshop.Store, usersDir string, rewriteProvider func(*http.Request) (novelfetchworkshop.RewriteProvider, error)) {
	mux.HandleFunc("GET /api/novel-fetch-workshop/logs", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		docs, err := store.ListDocuments(req.Context(), identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"logs": []any{}, "error": "读取日志失败"})
			return
		}
		logs := make([]any, 0)
		for _, doc := range docs {
			logs = append(logs, doc.Logs...)
		}
		writeJSON(w, http.StatusOK, map[string]any{"logs": logs})
	})

	mux.HandleFunc("GET /api/novel-fetch-workshop/batches/current", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		docs, err := store.ListDocuments(req.Context(), identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "读取批次失败"})
			return
		}
		ids := make([]string, 0, len(docs))
		for _, doc := range docs {
			ids = append(ids, doc.BookID)
		}
		writeJSON(w, http.StatusOK, map[string]any{"batch": map[string]any{"id": "current", "taskIds": ids, "bookIds": ids, "updatedAt": time.Now().UTC().Format(time.RFC3339Nano)}})
	})

	mux.HandleFunc("GET /api/novel-fetch-workshop/schedules", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		cfg, _ := store.GetConfig(req.Context(), identity.Username)
		schedules, _ := cfg["novel_fetch_schedules"].([]any)
		if schedules == nil {
			schedules = []any{}
		}
		writeJSON(w, http.StatusOK, map[string]any{"schedules": schedules})
	})
	mux.HandleFunc("POST /api/novel-fetch-workshop/schedules", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		var item map[string]any
		if err := decodeNovelFetchJSON(w, req, &item); err != nil {
			return
		}
		cfg, _ := store.GetConfig(req.Context(), identity.Username)
		schedules, _ := cfg["novel_fetch_schedules"].([]any)
		item["id"] = "schedule-" + time.Now().UTC().Format("20060102150405.000000000")
		item["updatedAt"] = time.Now().UTC().Format(time.RFC3339Nano)
		schedules = append(schedules, item)
		cfg["novel_fetch_schedules"] = schedules
		if err := store.PutConfig(req.Context(), identity.Username, cfg); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "保存定时任务失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"schedule": item})
	})
	mux.HandleFunc("PATCH /api/novel-fetch-workshop/schedules/{id}", func(w http.ResponseWriter, req *http.Request) { updateSchedule(w, req, store, false) })
	mux.HandleFunc("DELETE /api/novel-fetch-workshop/schedules/{id}", func(w http.ResponseWriter, req *http.Request) { updateSchedule(w, req, store, true) })

	mux.HandleFunc("POST /api/novel-fetch-workshop/knowledge/optimize", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		var in map[string]any
		if err := decodeNovelFetchJSON(w, req, &in); err != nil {
			return
		}
		kind, _ := in["kind"].(string)
		if kind == "" {
			kind, _ = in["libraryType"].(string)
		}
		if kind == "" {
			kind, _ = in["library_type"].(string)
		}
		item, _ := in["item"].(map[string]any)
		id, _ := item["id"].(string)
		if id == "" {
			id, _ = in["id"].(string)
		}
		if id == "" {
			id, _ = in["item_id"].(string)
		}
		if kind == "" || id == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "资料库类型或条目ID不能为空"})
			return
		}
		cfg, _ := store.GetConfig(req.Context(), identity.Username)
		var provider novelfetchworkshop.RewriteProvider
		if rewriteProvider != nil {
			provider, _ = rewriteProvider(req)
		}
		out, err := novelfetchworkshop.OptimizeKnowledge(req.Context(), cfg, kind, id, provider)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
			return
		}
		_ = store.PutConfig(req.Context(), identity.Username, cfg)
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "item": out})
	})

	mux.HandleFunc("POST /api/novel-fetch-workshop/tasks/batch-ai-count", func(w http.ResponseWriter, req *http.Request) { updateTaskMeta(w, req, store, "aiCount") })
	mux.HandleFunc("POST /api/novel-fetch-workshop/tasks/apply-rules", func(w http.ResponseWriter, req *http.Request) { applySensitive(w, req, store, false) })
	mux.HandleFunc("POST /api/novel-fetch-workshop/tasks/reprocess-sensitive", func(w http.ResponseWriter, req *http.Request) { applySensitive(w, req, store, true) })
	mux.HandleFunc("POST /api/novel-fetch-workshop/web-submit/operations", func(w http.ResponseWriter, req *http.Request) {
		var input map[string]any
		if err := decodeNovelFetchJSON(w, req, &input); err != nil {
			return
		}
		id := fmt.Sprintf("op-%d", time.Now().UnixNano())
		item := map[string]any{"id": id, "status": "done", "kind": input["kind"], "createdAt": time.Now().UTC().Format(time.RFC3339Nano), "updatedAt": time.Now().UTC().Format(time.RFC3339Nano)}
		if fmt.Sprint(input["kind"]) == "configs" {
			result, err := syncTargetConfigProfiles(req, store, usersDir)
			if err != nil {
				item["status"] = "failed"
				item["error"] = err.Error()
			} else {
				item["result"] = result
			}
		}
		novelFetchOperations.Lock()
		novelFetchOperations.items[id] = item
		novelFetchOperations.Unlock()
		// The legacy workbench reads id/status/result from the top level, while
		// newer callers also accept the nested operation object. Return both
		// shapes so polling never turns a valid operation into /undefined.
		writeJSON(w, http.StatusOK, legacyOperationResponse(item))
	})
	mux.HandleFunc("GET /api/novel-fetch-workshop/web-submit/operations/{id}", func(w http.ResponseWriter, req *http.Request) {
		novelFetchOperations.RLock()
		item := novelFetchOperations.items[req.PathValue("id")]
		novelFetchOperations.RUnlock()
		if item == nil {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "操作不存在"})
			return
		}
		writeJSON(w, http.StatusOK, legacyOperationResponse(item))
	})
}

func legacyOperationResponse(item map[string]any) map[string]any {
	response := map[string]any{"ok": true, "operation": item}
	for key, value := range item {
		response[key] = value
	}
	return response
}

// syncTargetConfigProfiles is shared by the old operation-based workbench and
// the newer direct sync endpoint. The old page used to mark the operation done
// without ever calling the target configuration API, which made the UI report
// zero profiles even when the target account had profiles.
func syncTargetConfigProfiles(req *http.Request, store novelfetchworkshop.Store, usersDir string) (map[string]any, error) {
	identity, ok := BridgeIdentityFromContext(req.Context())
	if !ok {
		return nil, fmt.Errorf("unauthorized")
	}
	session := readUploadTargetSession(usersDir, identity.Username)
	cookie := strings.TrimSpace(toString(session["cookie"]))
	if cookie == "" {
		return nil, fmt.Errorf("请先保存账号密码并登录目标站")
	}
	response, err := targetGET(req, "http://two.121w.com/tttadmin/api/zdy_config.php?action=list", cookie)
	if err != nil {
		return nil, fmt.Errorf("目标站配置档读取失败")
	}
	var payload map[string]any
	if json.Unmarshal(response, &payload) != nil {
		return nil, fmt.Errorf("目标站配置档返回格式无效")
	}
	items, hasItems := targetConfigItems(payload)
	if !hasItems || len(items) == 0 {
		return nil, fmt.Errorf("目标站未返回可用配置档")
	}
	settings, err := store.GetConfig(req.Context(), identity.Username)
	if err != nil {
		return nil, fmt.Errorf("读取提交配置失败")
	}
	app := mapOrEmpty(settings["appConfig"])
	web := mapOrEmpty(app["web_submit"])
	local := []any{}
	for _, item := range arrayOrEmpty(web["upload_profiles"]) {
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
		groups = append(groups, map[string]any{"id": "121-" + id, "name": firstText(source["config_name"], "121 配置档 "+id), "enabled": true, "source": "121", "config_id": id, "is_default": targetNumber(source["is_default"]) == 1, "platform_id": firstText(data["platform_id"], source["platform_id"]), "gender": strings.ReplaceAll(gender, "频", ""), "style": firstText(data["style_name"], source["style_name"], data["style"]), "advanced": data})
	}
	web["upload_profiles"] = append(local, groups...)
	app["web_submit"] = web
	settings["appConfig"] = app
	if err := store.PutConfig(req.Context(), identity.Username, settings); err != nil {
		return nil, fmt.Errorf("保存配置档失败")
	}
	return map[string]any{"ok": true, "settings": web, "groups": groups, "output": []string{fmt.Sprintf("已从视频管理系统同步 %d 个配置档", len(groups))}}, nil
}

func updateSchedule(w http.ResponseWriter, req *http.Request, store novelfetchworkshop.Store, remove bool) {
	identity, _ := BridgeIdentityFromContext(req.Context())
	cfg, _ := store.GetConfig(req.Context(), identity.Username)
	raw, _ := cfg["novel_fetch_schedules"].([]any)
	id := req.PathValue("id")
	out := make([]any, 0, len(raw))
	var found any
	var patch map[string]any
	if !remove {
		_ = decodeNovelFetchJSON(w, req, &patch)
	}
	for _, v := range raw {
		m, _ := v.(map[string]any)
		if m["id"] == id {
			if remove {
				found = m
				continue
			}
			for k, val := range patch {
				m[k] = val
			}
			m["updatedAt"] = time.Now().UTC().Format(time.RFC3339Nano)
			found = m
		}
		out = append(out, m)
	}
	if found == nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "定时任务不存在"})
		return
	}
	cfg["novel_fetch_schedules"] = out
	_ = store.PutConfig(req.Context(), identity.Username, cfg)
	writeJSON(w, http.StatusOK, map[string]any{"schedule": found})
}

func updateTaskMeta(w http.ResponseWriter, req *http.Request, store novelfetchworkshop.Store, key string) {
	identity, _ := BridgeIdentityFromContext(req.Context())
	var in struct {
		IDs     []string `json:"ids"`
		BookIDs []string `json:"bookIds"`
		Count   int      `json:"count"`
	}
	if err := decodeNovelFetchJSON(w, req, &in); err != nil {
		return
	}
	ids := in.IDs
	if len(ids) == 0 {
		ids = in.BookIDs
	}
	changed := 0
	for _, id := range ids {
		d, err := store.GetDocument(req.Context(), identity.Username, id)
		if err != nil {
			continue
		}
		if d.Meta == nil {
			d.Meta = map[string]any{}
		}
		d.Meta[key] = in.Count
		_ = store.PutDocument(req.Context(), identity.Username, d)
		changed++
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "updated": changed})
}

func applySensitive(w http.ResponseWriter, req *http.Request, store novelfetchworkshop.Store, restore bool) {
	identity, _ := BridgeIdentityFromContext(req.Context())
	var in struct {
		IDs     []string `json:"ids"`
		BookIDs []string `json:"bookIds"`
		Text    string   `json:"text"`
	}
	if err := decodeNovelFetchJSON(w, req, &in); err != nil {
		return
	}
	ids := in.IDs
	if len(ids) == 0 {
		ids = in.BookIDs
	}
	settings, _ := store.GetConfig(req.Context(), identity.Username)
	changed := 0
	replacements := 0
	for _, id := range ids {
		d, err := store.GetDocument(req.Context(), identity.Username, id)
		if err != nil {
			continue
		}
		d, err = hydrateNovelFetchDocument(req.Context(), store, identity.Username, d)
		if err != nil {
			continue
		}
		if restore && d.OriginalRaw != "" {
			d.Original = d.OriginalRaw
			d.ProcessedText = d.OriginalRaw
		}
		if d.ProcessedText == "" {
			d.ProcessedText = d.Original
		}
		processed, count, sensitiveItems := novelfetchworkshop.ApplySensitiveRulesWithDetails(d.ProcessedText, settings)
		d.ProcessedText = processed
		d.Original = processed
		replacements += count
		if d.Meta == nil {
			d.Meta = map[string]any{}
		}
		d.Meta["sensitiveProcessed"] = true
		d.Meta["sensitiveReplacementCount"] = count
		d.Meta["sensitiveStatus"] = "done"
		d.Meta["sensitiveReplaced"] = count
		setSensitiveDetails(d.Meta, sensitiveItems, count)
		d.Logs = append(d.Logs, map[string]any{"type": "sensitive_reprocessed", "at": time.Now().UTC().Format(time.RFC3339Nano), "replacements": count})
		d.ProcessedChars = len([]rune(d.ProcessedText))
		if err := persistNovelFetchBodies(req.Context(), store, identity.Username, &d, d.OriginalRaw, d.ProcessedText); err != nil {
			continue
		}
		_ = store.PutDocument(req.Context(), identity.Username, d)
		changed++
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "processed": changed, "applied": changed, "restored": restore, "replacements": replacements})
}
