package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
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
			Mode            string   `json:"mode"`
			Versions        []string `json:"versions"`
			OperationID     string   `json:"operationId"`
			ReadbackOnly    bool     `json:"readback_only"`
			PreviewRevision string   `json:"previewRevision"`
			// 参考实现接收 force/grouped/dry_run；其中 force 与 grouped
			// 在提交计划里没有独立分支，dry_run 只生成计划不执行上传。
			Force   bool `json:"force"`
			Grouped bool `json:"grouped"`
			DryRun  bool `json:"dry_run"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, req.Body, 64<<10)).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求内容无效"})
			return
		}
		if len(body.TaskIDs) == 0 {
			body.TaskIDs = body.IDs
		}
		// 参考实现支持 mode=selected|failed|all 的提交范围；
		// all 表示当前账号全部可见任务，不受页面日期/搜索过滤影响。
		resolvedIDs, err := resolveWorkshopTaskIDs(req.Context(), store, identity.Username, body.Mode, body.TaskIDs, workshopSubmitFailed)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "读取任务失败"})
			return
		}
		body.TaskIDs = resolvedIDs
		// 参考实现的提交不依赖预览令牌，预览只用于查看。这里保留字段兼容旧前端，
		// 但不再因为后台 meta 变化而拒绝提交；提交时按最新正文和配置重新生成计划。
		planPayload, status, message := buildWebSubmitPlanPayload(req.Context(), store, identity.Username, body.TaskIDs, body.Versions)
		if status != http.StatusOK {
			writeJSON(w, status, map[string]string{"error": message})
			return
		}
		// 无论 dry_run 还是真实提交，都先把计划落盘成 manifest；
		// 崩溃后可以据此识别 executing/unknown，避免重复派发。
		runID := newWebSubmitRunID()
		manifest, manifestErr := newWebSubmitManifest(usersDir, identity.Username, runID, map[bool]string{true: "dry_run", false: "manual_submit"}[body.DryRun], planPayload)
		if manifestErr != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "提交计划落盘失败，请检查数据目录后重试"})
			return
		}
		if body.DryRun {
			manifest.finish("dry_run")
			writeJSON(w, http.StatusOK, map[string]any{
				"ok":                true,
				"dry_run":           true,
				"requested":         planPayload["requested"],
				"group_count":       len(arrayOrEmpty(planPayload["groups"])),
				"success_groups":    0,
				"failed_groups":     0,
				"results":           []any{},
				"skipped":           planPayload["skipped"],
				"deferred_items":    planPayload["deferred_items"],
				"requested_items":   planPayload["requested_items"],
				"planned_items":     planPayload["planned_items"],
				"missing_items":     planPayload["missing_items"],
				"versions":          planPayload["versions"],
				"plan_complete":     planPayload["plan_complete"],
				"submission_run_id": runID,
				"groups":            planPayload["groups"],
				"plan":              planPayload["plan"],
				"material_limit":    planPayload["material_limit"],
			})
			return
		}
		result, err := executeWebSubmit(req.Context(), identity.Username, usersDir, store, webSubmitRequest{TaskIDs: body.TaskIDs, Versions: body.Versions, OperationID: body.OperationID, ReadbackOnly: body.ReadbackOnly, Manifest: manifest}, genderJudgeResolver, targets...)
		if err != nil {
			manifest.finish("aborted")
			status := http.StatusInternalServerError
			if requestErr, ok := err.(*webSubmitRequestError); ok {
				status = requestErr.Status
			}
			writeJSON(w, status, map[string]string{"error": err.Error()})
			return
		}
		manifest.finish("completed")
		// 保持 200 以兼容旧前端：非 2xx 会让老客户端直接进入硬失败分支并丢掉逐版本结果。
		// 整体成败通过 ok + result(all_ok/partial/all_failed) + 每个失败项的 code/message 表达。
		writeJSON(w, http.StatusOK, map[string]any{
			"ok":                result.Failed == 0,
			"result":            webSubmitOverallResult(result.Success, result.Failed),
			"operation_id":      result.OperationID,
			"submission_run_id": runID,
			"requested":         len(body.TaskIDs),
			"group_count":       len(arrayOrEmpty(planPayload["groups"])),
			"success_groups":    result.Success,
			"failed_groups":     result.Failed,
			"skipped":           planPayload["skipped"],
			"deferred_items":    planPayload["deferred_items"],
			"results":           decorateWebSubmitResults(result.Results),
		})
	}))
	// The legacy workbench calls /api/web-submit/submit, while the canonical
	// compatibility route lives under /api/batch-rewrite/web-submit/submit.
	// Keep both paths on the same handler so the Go 4001 page cannot 404.
	root.Handle("POST /api/batch-rewrite/web-submit/submit", submitHandler)
	root.Handle("POST /api/web-submit/submit", submitHandler)
	// 提交清单读取：用于提交历史、排障和崩溃后确认上次轮次的落盘状态。
	manifestHandler := middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		recoverStaleWebSubmitManifests(usersDir, identity.Username)
		limit, _ := strconv.Atoi(req.URL.Query().Get("limit"))
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "runs": listWebSubmitManifests(usersDir, identity.Username, limit)})
	}))
	root.Handle("GET /api/batch-rewrite/web-submit/manifests", manifestHandler)
	root.Handle("GET /api/web-submit/manifests", manifestHandler)
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
	// 与参考实现一致：配置里没有组织归属时不再默认 "1" 静默提交，
	// 而是由提交计划阻塞为 waiting_config，提示用户先同步组织。
	return firstText(web["organization"], profile["organization"], mapOrEmpty(profile["advanced"])["organization"])
}

// webSubmitMinFileSize 与参考实现一致：默认 3KB，避免把空文件或明显过短的正文推到网站。
func webSubmitMinFileSize(web map[string]any) int {
	if value, ok := web["min_file_size"]; ok {
		return intValue(value)
	}
	return 3 * 1024
}

func submissionResolvedGender(doc novelfetchworkshop.Document) string {
	gender, _ := submissionGenderStyle(doc)
	if genderNeedsResolution(gender) {
		return ""
	}
	return strings.TrimSpace(gender)
}

func submissionResolvedStyle(meta map[string]any) string {
	if styleNeedsResolution(meta) {
		return ""
	}
	return strings.TrimSpace(firstText(meta["style"], meta["styleName"], meta["style_name"]))
}

// submissionClassificationGaps 对齐参考实现：平台、男女频、风格类型缺失时
// 都不允许提交，而不是只阻塞风格。
func submissionClassificationGaps(platform, gender, style string) []string {
	gaps := []string{}
	if strings.TrimSpace(platform) == "" {
		gaps = append(gaps, "平台")
	}
	if strings.TrimSpace(gender) == "" {
		gaps = append(gaps, "男女频")
	}
	if strings.TrimSpace(style) == "" {
		gaps = append(gaps, "风格类型")
	}
	return gaps
}

func submissionGapError(gaps []string) string {
	return "无法映射网站字段：" + strings.Join(gaps, "、")
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

// webSubmitStatusText 把历史状态（pending_upload/error/pending_confirmation）
// 和新状态（waiting_config/unknown 等）收敛成同一套中文词汇，接口和页面共用。
func webSubmitStatusText(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "submitted", "success", "done", "uploaded":
		return "已提交"
	case "pending_confirmation", "pending_verification":
		return "待远端核验"
	case "unknown":
		return "结果未知"
	case "partial":
		return "部分接收"
	case "pending_upload", "pending", "pending_submit", "queued":
		return "待提交"
	case "waiting_configuration", "waiting_config":
		return "等待配置"
	case "waiting_login":
		return "等待登录"
	case "waiting_classification", "waiting_classify":
		return "等待确认分类"
	case "waiting_generate", "pending_generate":
		return "待生成"
	case "waiting_original":
		return "待原文"
	case "blocked":
		return "已阻塞"
	case "executing", "processing", "running":
		return "执行中"
	case "deferred":
		return "已延后"
	case "skipped":
		return "已跳过"
	case "failed", "error":
		return "失败"
	case "stopped":
		return "已停止"
	default:
		return strings.TrimSpace(status)
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
		cloned := make(map[string]any, len(row)+3)
		for key, value := range row {
			cloned[key] = value
		}
		if status := toString(cloned["status"]); strings.TrimSpace(status) != "" {
			cloned["status_label"] = webSubmitStatusText(status)
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
