package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"qiantie/backend/internal/novelfetchworkshop"
	"qiantie/backend/internal/videomanagementconnector"
)

func RegisterWebSubmitSubmitCompat(root *http.ServeMux, middleware BridgeAuth, usersDir string, store novelfetchworkshop.Store, genderJudgeResolver func(context.Context, string) (novelfetchworkshop.AIGenderJudge, error), rewriteResolver func(context.Context, string) (novelfetchworkshop.RewriteProvider, error), targets ...videomanagementconnector.Service) {
	if root == nil || store == nil || strings.TrimSpace(usersDir) == "" {
		return
	}
	// Legacy /api/web-submit/submit used to trust signed identity headers.
	// Disable the HMAC path so only a real login session can submit.
	middleware = middleware.SessionOnly()
	submitHandler := middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var body struct {
			TaskIDs         []string `json:"taskIds"`
			IDs             []string `json:"ids"`
			Versions        []string `json:"versions"`
			OperationID     string   `json:"operationId"`
			ReadbackOnly    bool     `json:"readback_only"`
			PreviewRevision string   `json:"previewRevision"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, req.Body, 64<<10)).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求内容无效"})
			return
		}
		if len(body.TaskIDs) == 0 {
			body.TaskIDs = body.IDs
		}
		if body.PreviewRevision != "" {
			revision, err := webSubmitPreviewRevision(req.Context(), store, identity.Username, body.TaskIDs, body.Versions)
			if err != nil || revision != body.PreviewRevision {
				writeJSON(w, http.StatusConflict, map[string]string{"error": "正文、版本或提交配置已变化，请重新生成提交预览后确认"})
				return
			}
		}
		result, err := executeWebSubmit(req.Context(), identity.Username, usersDir, store, webSubmitRequest{TaskIDs: body.TaskIDs, Versions: body.Versions, OperationID: body.OperationID, ReadbackOnly: body.ReadbackOnly}, genderJudgeResolver, targets...)
		if err != nil {
			status := http.StatusInternalServerError
			if requestErr, ok := err.(*webSubmitRequestError); ok {
				status = requestErr.Status
			}
			writeJSON(w, status, map[string]string{"error": err.Error()})
			return
		}
		// 保持 200 以兼容旧前端：非 2xx 会让老客户端直接进入硬失败分支并丢掉逐版本结果。
		// 整体成败通过 ok + result(all_ok/partial/all_failed) + 每个失败项的 code/message 表达。
		writeJSON(w, http.StatusOK, map[string]any{"ok": result.Failed == 0, "result": webSubmitOverallResult(result.Success, result.Failed), "operation_id": result.OperationID, "results": decorateWebSubmitResults(result.Results), "success_groups": result.Success, "failed_groups": result.Failed})
	}))
	// The legacy workbench calls /api/web-submit/submit, while the canonical
	// compatibility route lives under /api/batch-rewrite/web-submit/submit.
	// Keep both paths on the same handler so the Go 4001 page cannot 404.
	root.Handle("POST /api/batch-rewrite/web-submit/submit", submitHandler)
	root.Handle("POST /api/web-submit/submit", submitHandler)
}

func webSubmitSkipSubmitted(web map[string]any) bool {
	value, exists := web["skip_submitted"]
	if !exists {
		return true
	}
	result, ok := value.(bool)
	return !ok || result
}

func webSubmitProfile(web map[string]any, version string) map[string]any {
	profileID := strings.TrimSpace(toString(mapOrEmpty(web["profile_bindings"])[version]))
	if profileID == "" {
		profileID = strings.TrimSpace(toString(mapOrEmpty(web["version_profiles"])[version]))
	}
	if profileID == "" {
		profileID = strings.TrimSpace(toString(web["selected_profile"]))
	}
	profiles := arrayOrEmpty(web["upload_profiles"])
	if profileID == "" && len(profiles) > 0 {
		return mapOrEmpty(profiles[0])
	}
	for _, raw := range profiles {
		profile := mapOrEmpty(raw)
		if profileID == firstText(profile["id"], profile["name"]) || profileID == toString(profile["name"]) {
			return profile
		}
	}
	return map[string]any{}
}

func webSubmitAdvanced(web, profile map[string]any) map[string]any {
	result := cloneAnyMap(mapOrEmpty(profile["advanced"]))
	for key, value := range mapOrEmpty(web["advanced"]) {
		result[key] = value
	}
	return result
}

func webSubmitOrganization(web, profile map[string]any) string {
	return firstText(web["organization"], profile["organization"], mapOrEmpty(profile["advanced"])["organization"], "1")
}

func submissionGenderStyle(doc novelfetchworkshop.Document) (string, string) {
	meta := doc.Meta
	gender := strings.ReplaceAll(firstText(meta["gender"], meta["gender_name"], doc.Gender), "频", "")
	style := firstText(meta["style"], meta["styleName"], meta["style_name"])
	return strings.TrimSpace(gender), style
}

func resolveSubmissionGender(ctx context.Context, doc *novelfetchworkshop.Document, judge novelfetchworkshop.AIGenderJudge) {
	if doc == nil || !genderNeedsResolution(doc.Gender) {
		return
	}
	info := doc.BookInfo
	info.Text = strings.TrimSpace(doc.ProcessedText)
	if info.Text == "" {
		info.Text = strings.TrimSpace(doc.Original)
	}
	decision := novelfetchworkshop.ResolveGender(ctx, "", info, judge)
	doc.Gender, doc.GenderSource, doc.GenderReason = decision.Gender, decision.Source, decision.Reason
	doc.GenderConfidence = decision.Confidence
	syncNovelFetchClassificationMetadata(doc)
}

const webSubmitUnclassifiedStyleError = "该任务尚未识别或选择上传风格，无法提交到视频管理系统。"

func styleNeedsResolution(meta map[string]any) bool {
	style := strings.TrimSpace(firstText(meta["style"], meta["styleName"], meta["style_name"]))
	return style == "" || style == "待判定" || style == "未判定"
}

func uniqueStrings(values []string) []string {
	out := []string{}
	seen := map[string]bool{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" && !seen[value] {
			seen[value] = true
			out = append(out, value)
		}
	}
	return out
}
func removeString(values []string, target string) []string {
	out := []string{}
	for _, value := range values {
		if strings.TrimSpace(value) != target {
			out = append(out, value)
		}
	}
	return out
}

func markSubmissionFailure(meta map[string]any, version, message string) {
	if meta == nil {
		return
	}
	meta["siteSubmitStatus"] = "pending_upload"
	meta["siteSubmitFailedVersions"] = uniqueStrings(append(anyStrings(meta["siteSubmitFailedVersions"]), version))
	errors := mapOrEmpty(meta["siteSubmitFailedErrors"])
	errors[version] = strings.TrimSpace(message)
	meta["siteSubmitFailedErrors"] = errors
}

// webSubmitOverallResult 用稳定字段表达整体结果，避免调用方再从 ok 布尔值猜测。
func webSubmitOverallResult(success, failed int) string {
	switch {
	case failed == 0:
		return "all_ok"
	case success == 0:
		return "all_failed"
	default:
		return "partial"
	}
}

// decorateWebSubmitResults 为失败项补齐 code/message，便于前端逐条展示可操作原因。
func decorateWebSubmitResults(items []any) []any {
	if len(items) == 0 {
		return items
	}
	decorated := make([]any, 0, len(items))
	for _, item := range items {
		row, ok := item.(map[string]any)
		if !ok {
			decorated = append(decorated, item)
			continue
		}
		cloned := make(map[string]any, len(row)+2)
		for key, value := range row {
			cloned[key] = value
		}
		switch strings.ToLower(strings.TrimSpace(toString(cloned["status"]))) {
		case "pending_confirmation":
			if strings.TrimSpace(toString(cloned["code"])) == "" {
				cloned["code"] = "WEB_SUBMIT_PENDING_CONFIRMATION"
			}
			if strings.TrimSpace(toString(cloned["message"])) == "" {
				cloned["message"] = firstText(cloned["error"], "上传已接收，正在等待目标站回读确认")
			}
		case "error", "failed":
			if strings.TrimSpace(toString(cloned["code"])) == "" {
				cloned["code"] = "WEB_SUBMIT_FAILED"
			}
			if strings.TrimSpace(toString(cloned["message"])) == "" {
				cloned["message"] = firstText(cloned["error"], "网站提交失败")
			}
		}
		decorated = append(decorated, cloned)
	}
	return decorated
}
