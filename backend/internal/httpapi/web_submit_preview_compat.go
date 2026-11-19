package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"qiantie/backend/internal/novelfetchworkshop"
)

// RegisterWebSubmitPreviewCompat keeps the preview operation read-only while
// using the same Go workshop documents that the novel-fetch page displays.
func RegisterWebSubmitPreviewCompat(root *http.ServeMux, middleware BridgeAuth, store novelfetchworkshop.Store) {
	if root == nil || store == nil {
		return
	}
	// Preview is read-only, but it must still identify the caller from a real
	// session instead of caller-supplied signed headers.
	middleware = middleware.SessionOnly()
	previewHandler := middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var body struct {
			TaskIDs  []string `json:"taskIds"`
			IDs      []string `json:"ids"`
			Mode     string   `json:"mode"`
			Versions []string `json:"versions"`
			Force    bool     `json:"force"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, req.Body, 64<<10)).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求内容无效"})
			return
		}
		if len(body.TaskIDs) == 0 {
			body.TaskIDs = body.IDs
		}
		taskIDs, err := resolveWorkshopTaskIDs(req.Context(), store, identity.Username, body.Mode, body.TaskIDs, workshopSubmitFailed)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "读取任务失败"})
			return
		}
		payload, status, message := buildWebSubmitPlanPayload(req.Context(), store, identity.Username, taskIDs, body.Versions, body.Force)
		if status != http.StatusOK {
			writeJSON(w, status, map[string]string{"error": message})
			return
		}
		writeJSON(w, http.StatusOK, payload)
	}))
	root.Handle("POST /api/batch-rewrite/web-submit/preview", previewHandler)
}

// buildWebSubmitPlanPayload builds the read-only submission plan shared by the
// preview endpoint and the dry_run submit mode. 除 groups/skipped 外还返回
// requested/planned/deferred/missing 与 plan_complete，调用方可以据此判断
// 计划是否丢项，而不是只从 groups 是否为空猜测。
func buildWebSubmitPlanPayload(ctx context.Context, store novelfetchworkshop.Store, owner string, taskIDs, requestedVersions []string, force bool) (map[string]any, int, string) {
	settings, err := store.GetConfig(ctx, owner)
	if err != nil {
		return nil, http.StatusInternalServerError, "读取提交配置失败"
	}
	app := mapOrEmpty(settings["appConfig"])
	web := mapOrEmpty(app["web_submit"])
	versions := cleanStrings(requestedVersions)
	if len(versions) == 0 {
		versions = cleanStrings(anyStrings(web["submit_versions"]))
	}
	if len(versions) == 0 {
		versions = []string{"ai1"}
	}
	ids := cleanStrings(taskIDs)
	if len(ids) == 0 {
		return nil, http.StatusBadRequest, "没有收到选中的任务。请在任务列表勾选复选框后重新提交。"
	}
	groups := map[string]map[string]any{}
	order := []string{}
	skipped := []any{}
	requestedItems := []any{}
	requestedKeys := map[string]bool{}
	skipSubmitted := webSubmitSkipSubmitted(web)
	minFileSize := webSubmitMinFileSize(web)
	for _, id := range ids {
		doc, getErr := store.GetDocument(ctx, owner, id)
		if getErr != nil {
			skipped = append(skipped, map[string]any{"id": id, "status": "failed", "error": "任务不存在"})
			continue
		}
		meta := doc.Meta
		if meta == nil {
			meta = map[string]any{}
		}
		done := firstConfiguredValue(meta["siteSubmitDoneVersions"], meta["site_submit_done_versions"])
		pending := mapOrEmpty(meta["siteSubmitPendingConfirmation"])
		executing := anyStrings(meta["siteSubmitExecutingVersions"])
		for _, version := range versions {
			requestedItems = append(requestedItems, map[string]any{"id": id, "version": version})
			requestedKeys[webSubmitManifestKey(id, version)] = true
			if skipSubmitted && !force && containsVersion(done, version) {
				skipped = append(skipped, map[string]any{"id": id, "version": version, "status": "skipped", "reason": "confirmed_result", "error": "该版本已确认提交"})
				continue
			}
			if _, ok := pending[version]; ok {
				skipped = append(skipped, map[string]any{"id": id, "version": version, "status": "unknown", "reason": "remote_confirmation_required", "deferred": true, "error": "上次提交结果未知，需先远端核对；不会重复上传"})
				continue
			}
			if containsVersion(executing, version) {
				skipped = append(skipped, map[string]any{"id": id, "version": version, "status": "executing", "reason": "submission_in_progress", "deferred": true, "error": "该版本已有提交正在执行，已阻止重复派发"})
				continue
			}
			profile := webSubmitProfile(web, version)
			platform := firstText(profile["platform_id"], profile["platformId"], meta["platformId"], meta["platform_id"])
			if gaps := submissionClassificationGaps(platform, submissionResolvedGender(doc), submissionResolvedStyle(meta)); len(gaps) > 0 {
				status := "waiting_config"
				if styleNeedsResolution(meta) {
					status = "waiting_classify"
				}
				skipped = append(skipped, map[string]any{"id": id, "version": version, "status": status, "reason": status, "deferred": true, "error": submissionGapError(gaps)})
				continue
			}
			content, bodyErr := store.GetBody(ctx, owner, id, version)
			if bodyErr != nil || strings.TrimSpace(content.Content) == "" {
				status := "waiting_generate"
				if strings.EqualFold(version, "original") {
					status = "waiting_original"
				}
				skipped = append(skipped, map[string]any{"id": id, "version": version, "status": status, "reason": "file_missing", "deferred": true, "error": "正文尚未生成"})
				continue
			}
			if minFileSize > 0 && len([]byte(content.Content)) < minFileSize {
				status := "waiting_generate"
				if strings.EqualFold(version, "original") {
					status = "waiting_original"
				}
				skipped = append(skipped, map[string]any{"id": id, "version": version, "status": status, "reason": "file_too_small", "deferred": true, "error": fmt.Sprintf("正文不足 %d 字节", minFileSize)})
				continue
			}
			if strings.HasPrefix(strings.ToLower(version), "ai") {
				if err := novelfetchworkshop.ValidateRewriteProse(content.Content); err != nil {
					skipped = append(skipped, map[string]any{"id": id, "version": version, "status": "waiting_generate", "reason": "ai_output_rejected", "deferred": true, "error": "AI 正文复检未通过：" + err.Error()})
					continue
				}
			}
			organization := webSubmitOrganization(web, profile)
			if organization == "" {
				skipped = append(skipped, map[string]any{"id": id, "version": version, "status": "waiting_config", "reason": "organization_missing", "deferred": true, "error": "网站上传配置缺少组织归属，请先同步或重新保存上传配置"})
				continue
			}
			sourceBookID, ok := webSubmitSourceBookID(meta, id)
			if !ok {
				skipped = append(skipped, map[string]any{"id": id, "version": version, "status": "waiting_config", "reason": "invalid_site_book_id", "deferred": true, "error": "任务缺少合法的网站书号（纯数字），无法提交"})
				continue
			}
			gender, style := submissionGenderStyle(doc)
			profileID := "default"
			profileName := "默认配置"
			if len(profile) > 0 {
				profileID = firstText(profile["id"], profileID)
				profileName = firstText(profile["name"], profileName)
			}
			key := strings.Join([]string{platform, gender, style, version, profileID}, "-")
			group, exists := groups[key]
			if !exists {
				group = map[string]any{"group_id": key, "group_key": key, "status": "ready", "version": version, "summary": map[string]any{"platformId": platform, "gender": gender, "style": style, "profile_id": profileID, "profile_name": profileName, "organization": organization}, "advanced": webSubmitAdvanced(web, profile), "items": []any{}}
				groups[key] = group
				order = append(order, key)
			}
			items := group["items"].([]any)
			items = append(items, map[string]any{"id": id, "source_book_id": sourceBookID, "version": version, "size": len([]byte(content.Content)), "profile_id": profileID, "profile_name": profileName, "confirmation_only": false})
			group["items"] = items
		}
	}
	result := make([]any, 0, len(order))
	for _, key := range order {
		result = append(result, groups[key])
	}
	deferredItems := []any{}
	deferredKeys := map[string]bool{}
	for _, raw := range skipped {
		item := mapOrEmpty(raw)
		if item["deferred"] != true {
			continue
		}
		deferredItems = append(deferredItems, item)
		if id, version := firstText(item["id"]), firstText(item["version"]); id != "" && version != "" {
			deferredKeys[webSubmitManifestKey(id, version)] = true
		}
	}
	plannedItems := []any{}
	plannedKeys := map[string]bool{}
	for _, rawGroup := range result {
		group := mapOrEmpty(rawGroup)
		for _, rawItem := range arrayOrEmpty(group["items"]) {
			item := mapOrEmpty(rawItem)
			id, version := firstText(item["id"]), firstText(item["version"])
			plannedItems = append(plannedItems, map[string]any{"id": id, "version": version})
			plannedKeys[webSubmitManifestKey(id, version)] = true
		}
	}
	missingItems := []any{}
	for _, rawItem := range requestedItems {
		item := mapOrEmpty(rawItem)
		key := webSubmitManifestKey(firstText(item["id"]), firstText(item["version"]))
		if plannedKeys[key] || deferredKeys[key] {
			continue
		}
		missingItems = append(missingItems, item)
	}
	planComplete := true
	for key := range requestedKeys {
		if !plannedKeys[key] && !deferredKeys[key] {
			planComplete = false
			break
		}
	}
	plan := map[string]any{"groups": result, "skipped": skipped, "deferred_items": deferredItems, "versions": versions, "requested_items": requestedItems, "planned_items": plannedItems, "missing_items": missingItems, "plan_complete": planComplete}
	return map[string]any{"ok": true, "requested": len(ids), "groups": result, "skipped": skipped, "deferred_items": deferredItems, "versions": versions, "requested_items": requestedItems, "planned_items": plannedItems, "missing_items": missingItems, "plan_complete": planComplete, "plan": plan}, http.StatusOK, ""
}

func anyStrings(value any) []string {
	if values, ok := value.([]string); ok {
		return append([]string(nil), values...)
	}
	result := []string{}
	for _, item := range arrayOrEmpty(value) {
		result = append(result, toString(item))
	}
	return result
}

// cleanStrings / containsVersion 从 HEAD 恢复：工作区的提交预览重构删除了定义，
// 但 novel_fetch_automation / workshop 等文件仍在引用，缺了无法编译。
func cleanStrings(values []string) []string {
	seen := map[string]bool{}
	result := []string{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" && !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	return result
}

func containsVersion(value any, wanted string) bool {
	for _, item := range anyStrings(value) {
		if strings.TrimSpace(item) == wanted {
			return true
		}
	}
	return false
}
