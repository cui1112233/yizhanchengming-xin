package httpapi

import (
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
	root.Handle("POST /api/batch-rewrite/web-submit/preview", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
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
		settings, err := store.GetConfig(req.Context(), identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "读取提交配置失败"})
			return
		}
		app := mapOrEmpty(settings["appConfig"])
		web := mapOrEmpty(app["web_submit"])
		versions := cleanStrings(body.Versions)
		if len(versions) == 0 {
			versions = cleanStrings(anyStrings(web["submit_versions"]))
		}
		if len(versions) == 0 {
			versions = []string{"ai1"}
		}
		if len(body.TaskIDs) == 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "没有收到选中的任务。请在任务列表勾选复选框后重新提交。"})
			return
		}
		groups := map[string]map[string]any{}
		order := []string{}
		skipped := []any{}
		for _, rawID := range body.TaskIDs {
			id := strings.TrimSpace(rawID)
			if id == "" {
				continue
			}
			doc, getErr := store.GetDocument(req.Context(), identity.Username, id)
			if getErr != nil {
				skipped = append(skipped, map[string]any{"id": id, "status": "failed", "error": "任务不存在"})
				continue
			}
			meta := doc.Meta
			for _, version := range versions {
				if containsVersion(meta["siteSubmitDoneVersions"], version) || containsVersion(meta["site_submit_done_versions"], version) {
					skipped = append(skipped, map[string]any{"id": id, "version": version, "status": "skipped", "error": "该版本已确认提交"})
					continue
				}
				content, bodyErr := store.GetBody(req.Context(), identity.Username, id, version)
				if bodyErr != nil || strings.TrimSpace(content.Content) == "" {
					skipped = append(skipped, map[string]any{"id": id, "version": version, "status": "skipped", "error": "未找到正文版本"})
					continue
				}
				platform := firstText(meta["platformId"], meta["platform_id"])
				gender := firstText(meta["gender"])
				style := firstText(meta["style"])
				profileID := "default"
				profileName := "默认配置"
				if profiles := arrayOrEmpty(web["upload_profiles"]); len(profiles) > 0 {
					profile := mapOrEmpty(profiles[0])
					profileID = firstText(profile["id"], profileID)
					profileName = firstText(profile["name"], profileName)
				}
				key := strings.Join([]string{platform, gender, style, version, profileID}, "-")
				group, exists := groups[key]
				if !exists {
					group = map[string]any{"group_id": key, "group_key": key, "status": "ready", "version": version, "summary": map[string]any{"platformId": platform, "gender": gender, "style": style, "profile_id": profileID, "profile_name": profileName}, "advanced": mapOrEmpty(web["advanced"]), "items": []any{}}
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
		writeJSON(w, http.StatusOK, map[string]any{"material_limit": 8, "groups": result, "skipped": skipped})
	})))
}

func anyStrings(value any) []string {
	result := []string{}
	for _, item := range arrayOrEmpty(value) {
		result = append(result, toString(item))
	}
	return result
}

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
	for _, item := range arrayOrEmpty(value) {
		if strings.TrimSpace(toString(item)) == wanted {
			return true
		}
	}
	return false
}
