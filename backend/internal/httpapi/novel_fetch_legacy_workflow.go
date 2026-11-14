package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"qiantie/backend/internal/novelfetchworkshop"
	"qiantie/backend/internal/videomanagementconnector"
)

var novelFetchOperations = struct {
	sync.RWMutex
	items map[string]map[string]any
}{items: map[string]map[string]any{}}

// registerNovelFetchLegacyWorkflowRoutes keeps the old workbench API names
// on the Go store. These routes are deliberately small adapters: they return
// the same shapes the workbench expects while preserving the complete raw
// text in each stored document.
func registerNovelFetchLegacyWorkflowRoutes(mux *http.ServeMux, store novelfetchworkshop.Store, queue novelfetchworkshop.ProcessQueue, usersDir string, rewriteProvider func(*http.Request) (novelfetchworkshop.RewriteProvider, error), targets ...videomanagementconnector.Service) {
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
		settings, _ := store.GetConfig(req.Context(), identity.Username)
		currentBatchID := strings.TrimSpace(fmt.Sprint(settings["novel_fetch_current_batch_id"]))
		docs, err := store.ListDocuments(req.Context(), identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "读取批次失败"})
			return
		}
		ids := make([]string, 0, len(docs))
		states := make([]map[string]any, 0, len(docs))
		targets := []string{}
		seenTargets := map[string]bool{}
		for _, doc := range docs {
			if currentBatchID == "" || strings.TrimSpace(fmt.Sprint(doc.Meta["currentBatchId"])) != currentBatchID {
				continue
			}
			ids = append(ids, doc.BookID)
			state := legacyNovelFetchTask(doc)
			states = append(states, state)
			if doc.OriginalRawChars > 0 && !seenTargets["original"] {
				targets = append(targets, "original")
				seenTargets["original"] = true
			}
			generatedVersions, _ := state["ai_files"].([]string)
			for _, version := range []string{"ai1", "ai2", "ai3", "ai4", "ai5"} {
				for _, generated := range generatedVersions {
					if generated == version && !seenTargets[version] {
						targets = append(targets, version)
						seenTargets[version] = true
					}
				}
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{"batch": map[string]any{"id": currentBatchID, "status": deriveNovelFetchBatchStatus(filterNovelFetchBatchDocuments(docs, currentBatchID)), "taskIds": ids, "bookIds": ids, "taskStates": states, "settingsSnapshot": map[string]any{"target_versions": targets}, "updatedAt": time.Now().UTC().Format(time.RFC3339Nano)}})
	})

	// V78 的历史批次不是另建一套队列；批次快照直接从 Go 文档的 currentBatchId
	// 归并，避免页面刷新后又回到旧 Node 的内存批次列表。
	mux.HandleFunc("GET /api/novel-fetch-workshop/batches", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		docs, err := store.ListDocuments(req.Context(), identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "读取历史批次失败"})
			return
		}
		groups := map[string][]novelfetchworkshop.Document{}
		for _, doc := range docs {
			batchID := strings.TrimSpace(fmt.Sprint(doc.Meta["currentBatchId"]))
			if batchID == "" {
				continue
			}
			groups[batchID] = append(groups[batchID], doc)
		}
		batches := make([]map[string]any, 0, len(groups))
		for id, rows := range groups {
			ids := make([]string, 0, len(rows))
			states := make([]map[string]any, 0, len(rows))
			for _, row := range rows {
				ids = append(ids, row.BookID)
				states = append(states, legacyNovelFetchTask(row))
			}
			batches = append(batches, map[string]any{
				"id": id, "taskIds": ids, "bookIds": ids, "taskStates": states,
				"createdAt":        firstText(rows[0].Meta["batchCreatedAt"], rows[0].Meta["createdAt"]),
				"status":           deriveNovelFetchBatchStatus(rows),
				"settingsSnapshot": map[string]any{"target_versions": []string{"original"}},
			})
		}
		sort.Slice(batches, func(i, j int) bool { return fmt.Sprint(batches[i]["id"]) > fmt.Sprint(batches[j]["id"]) })
		writeJSON(w, http.StatusOK, map[string]any{"batches": batches})
	})

	mux.HandleFunc("POST /api/novel-fetch-workshop/batches/{batchId}/rerun", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		batchID := strings.TrimSpace(req.PathValue("batchId"))
		var input struct {
			Mode string `json:"mode"`
		}
		_ = decodeNovelFetchJSON(w, req, &input)
		rows, err := store.ListDocuments(req.Context(), identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "读取重跑批次失败"})
			return
		}
		ids := make([]string, 0)
		lines := make([]string, 0)
		for _, row := range rows {
			if strings.TrimSpace(fmt.Sprint(row.Meta["currentBatchId"])) != batchID {
				continue
			}
			failed := strings.Contains(strings.ToLower(fmt.Sprint(row.Meta["status"])), "fail") || strings.Contains(strings.ToLower(fmt.Sprint(row.Meta["aiStatus"])), "fail")
			if strings.EqualFold(input.Mode, "abnormal") && !failed {
				continue
			}
			ids = append(ids, row.BookID)
			name := firstText(row.BookInfo.WorkTitle, row.Meta["bookName"])
			lines = append(lines, row.BookID+func() string {
				if name == "" {
					return ""
				}
				return "\t" + name
			}())
		}
		if len(ids) == 0 {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "批次不存在或没有可重跑任务"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"source_batch_id": batchID, "preselected_book_ids": ids, "input_snapshot": strings.Join(lines, "\n"), "payload": map[string]any{"target_versions": []string{"original"}, "input_text": strings.Join(lines, "\n")}})
	})

	stopSelected := func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		var input struct {
			IDs     []string `json:"ids"`
			BookIDs []string `json:"bookIds"`
		}
		if err := decodeNovelFetchJSON(w, req, &input); err != nil {
			return
		}
		ids := input.IDs
		if len(ids) == 0 {
			ids = input.BookIDs
		}
		rows := make([]novelfetchworkshop.Document, 0, len(ids))
		stoppableIDs := make([]string, 0, len(ids))
		for _, id := range ids {
			row, err := store.GetDocument(req.Context(), identity.Username, strings.TrimSpace(id))
			if err != nil {
				continue
			}
			status := strings.ToLower(firstText(row.Meta["status"], row.Meta["queueState"]))
			if status == "completed" || status == "done" || status == "success" {
				continue
			}
			rows = append(rows, row)
			stoppableIDs = append(stoppableIDs, row.BookID)
		}
		if cancellationQueue, ok := queue.(novelfetchworkshop.ProcessCancellationQueue); ok {
			if err := cancellationQueue.CancelBooks(req.Context(), identity.Username, stoppableIDs); err != nil {
				writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "Redis 取消任务失败"})
				return
			}
		}
		updated := 0
		for _, row := range rows {
			if row.Meta == nil {
				row.Meta = map[string]any{}
			}
			row.Meta["status"] = "stopped"
			row.Meta["queueState"] = "stopped"
			row.Logs = append(row.Logs, map[string]any{"type": "stopped", "at": time.Now().UTC().Format(time.RFC3339Nano)})
			if err := store.PutDocument(req.Context(), identity.Username, row); err == nil {
				updated++
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "updated": updated})
	}
	mux.HandleFunc("POST /api/novel-fetch-workshop/tasks/stop-selected", stopSelected)
	// Keep the old process-level stop action functional. Older pages posted
	// here instead of using the newer task-scoped endpoint.
	mux.HandleFunc("POST /api/novel-fetch-workshop/process/stop", stopSelected)

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
		cfg, err := store.GetConfig(req.Context(), identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "读取定时任务配置失败"})
			return
		}
		if err := normalizeNovelFetchScheduleInput(item, cfg); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		schedules, _ := cfg["novel_fetch_schedules"].([]any)
		item["id"] = "schedule-" + time.Now().UTC().Format("20060102150405.000000000")
		item["status"] = "preparing"
		item["created_at"] = time.Now().UTC().Format(time.RFC3339Nano)
		// 定时处理的既定行为是处理完成后自动上传；普通立即处理不受影响。
		item["auto_submit"] = true
		item["updatedAt"] = time.Now().UTC().Format(time.RFC3339Nano)
		schedules = append(schedules, item)
		cfg["novel_fetch_schedules"] = schedules
		if err := store.PutConfig(req.Context(), identity.Username, cfg); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "保存定时任务失败"})
			return
		}
		if err := prepareNovelFetchScheduleTasks(req.Context(), store, identity.Username, item, time.Now().UTC()); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "定时任务已保存，书籍准备中断，系统会继续恢复：" + err.Error(), "schedule_id": item["id"]})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"schedule": item})
	})
	mux.HandleFunc("PATCH /api/novel-fetch-workshop/schedules/{id}", func(w http.ResponseWriter, req *http.Request) { updateSchedule(w, req, store, false) })
	mux.HandleFunc("DELETE /api/novel-fetch-workshop/schedules/{id}", func(w http.ResponseWriter, req *http.Request) { updateSchedule(w, req, store, true) })
	mux.HandleFunc("POST /api/novel-fetch-workshop/schedules/{id}/cancel", func(w http.ResponseWriter, req *http.Request) { cancelWorkshopSchedule(w, req, store, queue) })

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
		cfg, err := store.GetConfig(req.Context(), identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "读取资料库失败：" + err.Error()})
			return
		}
		var provider novelfetchworkshop.RewriteProvider
		if rewriteProvider != nil {
			provider, _ = rewriteProvider(req)
		}
		out, err := novelfetchworkshop.OptimizeKnowledge(req.Context(), cfg, kind, id, provider)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
			return
		}
		if err := store.PutConfig(req.Context(), identity.Username, cfg); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "优化结果保存失败：" + err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "item": out})
	})

	mux.HandleFunc("POST /api/novel-fetch-workshop/tasks/batch-ai-count", func(w http.ResponseWriter, req *http.Request) { updateTaskMeta(w, req, store, "aiCount") })
	mux.HandleFunc("POST /api/novel-fetch-workshop/tasks/apply-rules", func(w http.ResponseWriter, req *http.Request) { applySensitive(w, req, store, false, rewriteProvider) })
	mux.HandleFunc("POST /api/novel-fetch-workshop/tasks/reprocess-sensitive", func(w http.ResponseWriter, req *http.Request) { applySensitive(w, req, store, true, rewriteProvider) })
	mux.HandleFunc("POST /api/novel-fetch-workshop/web-submit/operations", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		var input map[string]any
		if err := decodeNovelFetchJSON(w, req, &input); err != nil {
			return
		}
		if toString(input["kind"]) != "configs" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "不支持该操作；同步风格请使用风格目录同步入口"})
			return
		}
		id := fmt.Sprintf("op-%d", time.Now().UnixNano())
		item := map[string]any{"id": id, "status": "done", "kind": input["kind"], "createdAt": time.Now().UTC().Format(time.RFC3339Nano), "updatedAt": time.Now().UTC().Format(time.RFC3339Nano)}
		if fmt.Sprint(input["kind"]) == "configs" {
			result, err := syncTargetConfigProfiles(req, store, usersDir, targets...)
			if err != nil {
				item["status"] = "failed"
				item["error"] = err.Error()
			} else {
				item["result"] = result
			}
		}
		novelFetchOperations.Lock()
		// Expire old polling snapshots instead of retaining every operation forever.
		for key, previous := range novelFetchOperations.items {
			created, err := time.Parse(time.RFC3339Nano, toString(previous["createdAt"]))
			if err != nil || time.Since(created) > time.Hour {
				delete(novelFetchOperations.items, key)
			}
		}
		novelFetchOperations.items[identity.Username+"\x00"+id] = item
		novelFetchOperations.Unlock()
		// The legacy workbench reads id/status/result from the top level, while
		// newer callers also accept the nested operation object. Return both
		// shapes so polling never turns a valid operation into /undefined.
		writeJSON(w, http.StatusOK, legacyOperationResponse(item))
	})
	mux.HandleFunc("GET /api/novel-fetch-workshop/web-submit/operations/{id}", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		novelFetchOperations.RLock()
		item := novelFetchOperations.items[identity.Username+"\x00"+req.PathValue("id")]
		novelFetchOperations.RUnlock()
		if item == nil {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "操作不存在"})
			return
		}
		writeJSON(w, http.StatusOK, legacyOperationResponse(item))
	})
}

func filterNovelFetchBatchDocuments(documents []novelfetchworkshop.Document, batchID string) []novelfetchworkshop.Document {
	result := make([]novelfetchworkshop.Document, 0, len(documents))
	for _, document := range documents {
		if strings.TrimSpace(fmt.Sprint(document.Meta["currentBatchId"])) == batchID {
			result = append(result, document)
		}
	}
	return result
}

func deriveNovelFetchBatchStatus(documents []novelfetchworkshop.Document) string {
	if len(documents) == 0 {
		return "pending"
	}
	counts := map[string]int{}
	for _, document := range documents {
		counts[novelfetchworkshop.DeriveTaskStatus(document)]++
	}
	if counts[novelfetchworkshop.TaskStatusFailed] > 0 {
		return "failed"
	}
	if counts[novelfetchworkshop.TaskStatusProcessing] > 0 {
		return "processing"
	}
	if counts[novelfetchworkshop.TaskStatusQueued] > 0 {
		return "queued"
	}
	if counts[novelfetchworkshop.TaskStatusCompleted] == len(documents) {
		return "done"
	}
	return "pending"
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
func syncTargetConfigProfiles(req *http.Request, store novelfetchworkshop.Store, usersDir string, targets ...videomanagementconnector.Service) (map[string]any, error) {
	identity, ok := BridgeIdentityFromContext(req.Context())
	if !ok {
		return nil, fmt.Errorf("unauthorized")
	}
	session := readUploadTargetSession(usersDir, identity.Username)
	cookie := strings.TrimSpace(toString(session["cookie"]))
	if cookie == "" {
		return nil, fmt.Errorf("请先保存账号密码并登录目标站")
	}
	var payload map[string]any
	if len(targets) > 0 && targets[0] != nil {
		reader, ok := targets[0].(interface {
			RefreshProfiles(context.Context, videomanagementconnector.Session) (any, error)
		})
		if !ok {
			return nil, fmt.Errorf("视频管理系统 Go 连接器不支持配置档读取")
		}
		value, err := reader.RefreshProfiles(req.Context(), videomanagementconnector.Session{Cookie: cookie, TargetUsername: toString(session["targetUsername"])})
		if err != nil {
			return nil, fmt.Errorf("目标站配置档读取失败")
		}
		payload = mapOrEmpty(value)
	} else {
		return nil, fmt.Errorf("视频管理系统 Go 连接器未配置")
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
		if err := decodeNovelFetchJSON(w, req, &patch); err != nil {
			return
		}
	}
	for _, v := range raw {
		m, _ := v.(map[string]any)
		if m["id"] == id {
			if remove {
				if !workshopScheduleTerminal(firstText(m["status"])) {
					writeJSON(w, http.StatusConflict, map[string]any{"error": "请先取消或停止定时任务，再删除记录；删除记录不会删除书目"})
					return
				}
				found = m
				continue
			}
			// Validate the edited time before changing the saved item. Rejected
			// changes must leave the original schedule executable.
			if _, snake := patch["run_at"]; snake || patch["runAt"] != nil {
				runAt, err := parseNovelFetchScheduleTime(strings.TrimSpace(firstText(patch["run_at"], patch["runAt"])))
				if err != nil {
					writeJSON(w, http.StatusBadRequest, map[string]any{"error": "请选择有效的定时处理时间"})
					return
				}
				patch["run_at"] = runAt.UTC().Format(time.RFC3339Nano)
			}
			m = cloneAnyMap(m)
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
	if err := store.PutConfig(req.Context(), identity.Username, cfg); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "保存定时任务失败"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"schedule": found})
}

func workshopScheduleTerminal(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "cancelled", "canceled", "completed", "done", "success", "failed", "stopped", "skipped_existing":
		return true
	}
	return false
}

func cancelWorkshopSchedule(w http.ResponseWriter, req *http.Request, store novelfetchworkshop.Store, queue novelfetchworkshop.ProcessQueue) {
	identity, _ := BridgeIdentityFromContext(req.Context())
	cfg, err := store.GetConfig(req.Context(), identity.Username)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "读取定时任务失败"})
		return
	}
	schedules, _ := cfg["novel_fetch_schedules"].([]any)
	for _, raw := range schedules {
		item, _ := raw.(map[string]any)
		if firstText(item["id"]) != req.PathValue("id") {
			continue
		}
		status := firstText(item["status"])
		if workshopScheduleTerminal(status) {
			writeJSON(w, http.StatusOK, map[string]any{"ok": true, "schedule": item, "deleted_count": 0})
			return
		}
		if status == "dispatching" {
			writeJSON(w, http.StatusConflict, map[string]any{"error": "任务正在入队，请稍后刷新再取消"})
			return
		}
		ids := scheduleStringSlice(item["book_ids"])
		if len(ids) > 0 {
			cancellation, ok := queue.(novelfetchworkshop.ProcessCancellationQueue)
			if !ok {
				writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "任务队列暂不支持取消，请稍后重试"})
				return
			}
			if err := cancellation.CancelBooks(req.Context(), identity.Username, ids); err != nil {
				writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "取消队列任务失败，请重试"})
				return
			}
		}
		// Running books retain their bodies: cancellation is cooperative. Only
		// a never-dispatched schedule may clean explicitly owned created_ids.
		deleted := 0
		if status == "scheduled" || status == "preparing" || status == "paused" || status == "" {
			owned := []string{}
			createdIDs := scheduleStringSlice(item["created_ids"])
			for _, id := range uniqueStrings(append(scheduleStringSlice(item["book_ids"]), createdIDs...)) {
				doc, getErr := store.GetDocument(req.Context(), identity.Username, id)
				if getErr == nil && firstText(doc.Meta["schedule_id"], doc.Meta["scheduleId"]) == firstText(item["id"]) {
					if !seenString(createdIDs, id) || strings.TrimSpace(doc.OriginalRaw) != "" || strings.TrimSpace(doc.Original) != "" || strings.TrimSpace(doc.ProcessedText) != "" || len(doc.Versions) > 0 || len(doc.BodyRefs) > 0 {
						delete(doc.Meta, "scheduledHold")
						doc.Meta["status"], doc.Meta["queueState"] = "stopped", "stopped"
						if err := store.PutDocument(req.Context(), identity.Username, doc); err != nil {
							writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "保留定时书籍状态失败，请重试"})
							return
						}
						continue
					}
					owned = append(owned, id)
				}
			}
			if len(owned) > 0 {
				result, deleteErr := store.DeleteDocuments(req.Context(), identity.Username, owned)
				if deleteErr != nil {
					writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "清理定时创建的书目失败，请重试"})
					return
				}
				deleted = result.Deleted
			}
		}
		item["status"] = "cancelled"
		item["cancelled_at"] = time.Now().UTC().Format(time.RFC3339Nano)
		item["updatedAt"] = item["cancelled_at"]
		item["message"] = "定时任务已取消；已生成内容保留"
		cfg["novel_fetch_schedules"] = schedules
		if err := store.PutConfig(req.Context(), identity.Username, cfg); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "保存取消状态失败，请重试"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "schedule": item, "deleted_count": deleted})
		return
	}
	writeJSON(w, http.StatusNotFound, map[string]any{"error": "定时任务不存在"})
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
	uniqueIDs := uniqueNovelFetchTaskIDs(ids)
	if in.Count < 1 || in.Count > 20 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "AI 数量须为 1 到 20"})
		return
	}
	cfg, err := store.GetConfig(req.Context(), identity.Username)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "读取 AI 数量配置失败"})
		return
	}
	changed, failed := 0, 0
	updatedIDs := make([]string, 0, len(uniqueIDs))
	for _, id := range uniqueIDs {
		d, err := store.GetDocument(req.Context(), identity.Username, id)
		if err != nil {
			failed++
			continue
		}
		if d.Meta == nil {
			d.Meta = map[string]any{}
		}
		d.Meta[key] = workshopAIRequestedCount(d, cfg, in.Count)
		d.Meta["ai_count"] = d.Meta[key]
		if err := store.PutDocument(req.Context(), identity.Username, d); err != nil {
			failed++
			continue
		}
		changed++
		updatedIDs = append(updatedIDs, id)
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": failed == 0, "requested": len(uniqueIDs), "updated": changed, "skipped": 0, "failed": failed, "task_ids": updatedIDs, "message": "已保存，下次生成 AI 改文时只补缺少的版本；当前运行任务和已有版本不受影响"})
}

func applySensitive(w http.ResponseWriter, req *http.Request, store novelfetchworkshop.Store, restore bool, resolveProvider func(*http.Request) (novelfetchworkshop.RewriteProvider, error)) {
	identity, _ := BridgeIdentityFromContext(req.Context())
	var in struct {
		IDs                []string `json:"ids"`
		BookIDs            []string `json:"bookIds"`
		Text               string   `json:"text"`
		SensitiveAIEnabled bool     `json:"sensitive_ai_enabled"`
		Scope              string   `json:"scope"`
	}
	if err := decodeNovelFetchJSON(w, req, &in); err != nil {
		return
	}
	ids := in.IDs
	if len(ids) == 0 {
		ids = in.BookIDs
	}
	settings, _ := store.GetConfig(req.Context(), identity.Username)
	var provider novelfetchworkshop.RewriteProvider
	if in.SensitiveAIEnabled && resolveProvider != nil {
		provider, _ = resolveProvider(req)
	}
	uniqueIDs := uniqueNovelFetchTaskIDs(ids)
	changed, failed := 0, 0
	updatedIDs := make([]string, 0, len(uniqueIDs))
	replacements := 0
	scope := strings.ToLower(strings.TrimSpace(in.Scope))
	if scope != "original" && scope != "ai" && scope != "both" {
		scope = "both"
	}
	for _, id := range uniqueIDs {
		d, err := store.GetDocument(req.Context(), identity.Username, id)
		if err != nil {
			failed++
			continue
		}
		d, err = hydrateNovelFetchDocument(req.Context(), store, identity.Username, d)
		if err != nil {
			failed++
			continue
		}
		if d.Meta == nil {
			d.Meta = map[string]any{}
		}
		if scope == "original" || scope == "both" {
			base := d.ProcessedText
			if base == "" {
				base = d.Original
			}
			if restore && d.OriginalRaw != "" {
				base = novelfetchworkshop.ProcessRawText(d.OriginalRaw, d.MaxTxt).ProcessedText
			}
			var processed string
			var sensitiveItems []map[string]any
			var trace []map[string]any
			if restore {
				processed, _, sensitiveItems = novelfetchworkshop.ApplySensitiveRulesForScopeWithDetails(base, settings, "original")
				processed, sensitiveItems = applySensitiveAIRepair(req.Context(), provider, settings, d.BookID, processed, sensitiveItems)
			} else {
				processed, trace, sensitiveItems, err = processWorkshopConfiguredText(req.Context(), store, identity.Username, d.BookID, base, "original", in.SensitiveAIEnabled, provider)
				if err != nil {
					continue
				}
				d.Meta["rulesTrace"] = trace
			}
			count := len(sensitiveItems)
			d.ProcessedText = processed
			d.Original = processed
			replacements += count
			d.Meta["sensitiveReplacementCount"] = count
			d.Meta["sensitiveReplaced"] = count
			setSensitiveDetails(d.Meta, sensitiveItems, count)
		}
		if !restore && (scope == "ai" || scope == "both") {
			traceByVersion := mapOrEmpty(d.Meta["rulesTraceByVersion"])
			for versionID, raw := range d.Versions {
				if !strings.HasPrefix(strings.ToLower(versionID), "ai") {
					continue
				}
				content := strings.TrimSpace(fmt.Sprint(raw))
				if content == "" {
					continue
				}
				processed, trace, sensitiveItems, processErr := processWorkshopConfiguredText(req.Context(), store, identity.Username, d.BookID, content, "ai", in.SensitiveAIEnabled, provider)
				if processErr != nil {
					failed++
					continue
				}
				d.Versions[versionID] = processed
				traceByVersion[versionID] = trace
				replacements += len(sensitiveItems)
				if ref, bodyErr := store.PutBody(req.Context(), identity.Username, novelfetchworkshop.BodyRecord{BookID: d.BookID, BodyRef: novelfetchworkshop.BodyRef{VersionID: versionID, State: "ready"}, Content: processed}); bodyErr == nil {
					d.BodyRefs[versionID] = ref
				}
			}
			d.Meta["rulesTraceByVersion"] = traceByVersion
		}
		d.Meta["sensitiveProcessed"] = true
		d.Meta["sensitiveStatus"] = "done"
		d.Logs = append(d.Logs, map[string]any{"type": "rules_processed", "at": time.Now().UTC().Format(time.RFC3339Nano), "scope": scope})
		d.ProcessedChars = len([]rune(d.ProcessedText))
		if err := persistNovelFetchBodies(req.Context(), store, identity.Username, &d, d.OriginalRaw, d.ProcessedText); err != nil {
			failed++
			continue
		}
		if err := store.PutDocument(req.Context(), identity.Username, d); err != nil {
			failed++
			continue
		}
		changed++
		updatedIDs = append(updatedIDs, id)
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "requested": len(uniqueIDs), "updated": changed, "skipped": 0, "failed": failed, "task_ids": updatedIDs, "processed": changed, "applied": changed, "restored": restore, "replacements": replacements})
}

func uniqueNovelFetchTaskIDs(ids []string) []string {
	result := make([]string, 0, len(ids))
	seen := make(map[string]bool, len(ids))
	for _, raw := range ids {
		id := strings.TrimSpace(raw)
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		result = append(result, id)
	}
	return result
}
