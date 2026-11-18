package httpapi

import (
	"context"
	"encoding/json"
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
			Versions []string `json:"versions"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, req.Body, 64<<10)).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求内容无效"})
			return
		}
		payload, status, message := buildWebSubmitPlanPayload(req.Context(), store, identity.Username, body.TaskIDs, body.Versions)
		if status != http.StatusOK {
			writeJSON(w, status, map[string]string{"error": message})
			return
		}
		writeJSON(w, http.StatusOK, payload)
	}))
	root.Handle("POST /api/batch-rewrite/web-submit/preview", previewHandler)
}

// buildWebSubmitPlanPayload builds the read-only submission plan shared by the
// preview endpoint and the dry_run submit mode.
func buildWebSubmitPlanPayload(ctx context.Context, store novelfetchworkshop.Store, owner string, taskIDs, requestedVersions []string) (map[string]any, int, string) {
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
	if len(taskIDs) == 0 {
		return nil, http.StatusBadRequest, "没有收到选中的任务。请在任务列表勾选复选框后重新提交。"
	}
	groups := map[string]map[string]any{}
	order := []string{}
	skipped := []any{}
	skipSubmitted := webSubmitSkipSubmitted(web)
	minTextChars := intValue(web["min_text_chars"])
	for _, rawID := range taskIDs {
		id := strings.TrimSpace(rawID)
		if id == "" {
			continue
		}
		doc, getErr := store.GetDocument(ctx, owner, id)
		if getErr != nil {
			skipped = append(skipped, map[string]any{"id": id, "status": "failed", "error": "任务不存在"})
			continue
		}
		meta := doc.Meta
		for _, version := range versions {
			if skipSubmitted && (containsVersion(meta["siteSubmitDoneVersions"], version) || containsVersion(meta["site_submit_done_versions"], version)) {
				skipped = append(skipped, map[string]any{"id": id, "version": version, "status": "skipped", "error": "该版本已确认提交"})
				continue
			}
			if styleNeedsResolution(meta) {
				skipped = append(skipped, map[string]any{"id": id, "version": version, "status": "blocked", "error": webSubmitUnclassifiedStyleError})
				continue
			}
			content, bodyErr := store.GetBody(ctx, owner, id, version)
			if bodyErr != nil || strings.TrimSpace(content.Content) == "" {
				skipped = append(skipped, map[string]any{"id": id, "version": version, "status": "skipped", "error": "未找到正文版本"})
				continue
			}
			if minTextChars > 0 && len([]rune(content.Content)) < minTextChars {
				skipped = append(skipped, map[string]any{"id": id, "version": version, "status": "skipped", "error": "正文少于最少字数"})
				continue
			}
			profile := webSubmitProfile(web, version)
			platform := firstText(profile["platform_id"], profile["platformId"], meta["platformId"], meta["platform_id"])
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
				group = map[string]any{"group_id": key, "group_key": key, "status": "ready", "version": version, "summary": map[string]any{"platformId": platform, "gender": gender, "style": style, "profile_id": profileID, "profile_name": profileName, "organization": webSubmitOrganization(web, profile)}, "advanced": webSubmitAdvanced(web, profile), "items": []any{}}
				groups[key] = group
				order = append(order, key)
			}
			items := group["items"].([]any)
			items = append(items, map[string]any{"id": id, "version": version, "size": len([]rune(content.Content)), "profile_id": profileID, "profile_name": profileName, "confirmation_only": false})
			group["items"] = items
		}
	}
	result := make([]any, 0, len(order))
	for _, key := range order {
		result = append(result, groups[key])
	}
	plan := map[string]any{"material_limit": 8, "groups": result, "skipped": skipped}
	revision, err := webSubmitPreviewRevision(ctx, store, owner, taskIDs, requestedVersions)
	if err != nil {
		return nil, http.StatusInternalServerError, "生成提交预览校验失败"
	}
	plan["revision"] = revision
	return map[string]any{"material_limit": 8, "groups": result, "skipped": skipped, "plan": plan}, http.StatusOK, ""
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
// 但 novel_fetch_automation / workshop / preview_revision 等文件仍在引用，缺了无法编译。
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
