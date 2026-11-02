package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"regexp"
	"strings"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/localartifact"
	"qiantie/backend/internal/mergeworker"
	"qiantie/backend/internal/videomanagementconnector"
)

var publish121BookID = regexp.MustCompile(`^(?:\d{1,20}|[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12})$`)

// BatchFactory121Publisher is the Go owner of the Batch Factory upload path.
// It deliberately consumes the persisted merge artifact and upload-target
// session; browser supplied media URLs are never trusted.
type BatchFactory121Publisher struct {
	Store       batchfactoryv11.Store
	UsersDir    string
	LocalFiles  *localartifact.Store
	MergedFiles mergeworker.ObjectReader
	Connector   videomanagementconnector.Service
}

func selectPublishMergeJobs(batch batchfactoryv11.Batch, book batchfactoryv11.Book, jobs []batchfactoryv11.MergeJob) ([]batchfactoryv11.MergeJob, error) {
	requiredCount := 1
	settings := batchfactoryv11.ResolveSettings(batch.SettingsState.Patch, book.SettingsState.Patch)
	var openingEnabled bool
	if raw := settings["openingEnabled"]; len(raw) > 0 {
		_ = json.Unmarshal(raw, &openingEnabled)
	}
	if openingEnabled && len(book.Videos) >= 2 {
		requiredCount = 4
		if raw := settings["openingCount"]; len(raw) > 0 {
			if err := json.Unmarshal(raw, &requiredCount); err != nil {
				return nil, fmt.Errorf("换开头数量配置无效")
			}
		}
		if requiredCount < 1 || requiredCount > 8 {
			return nil, fmt.Errorf("换开头数量必须在 1 到 8 之间")
		}
	}

	newestByOpening := make(map[int]batchfactoryv11.MergeJob, requiredCount)
	for _, job := range jobs {
		if job.BookID != book.ID || job.Status != batchfactoryv11.MergeSucceeded || strings.TrimSpace(job.OutputURL) == "" {
			continue
		}
		current, exists := newestByOpening[job.OpeningVariantIndex]
		if !exists || job.UpdatedAt.After(current.UpdatedAt) {
			newestByOpening[job.OpeningVariantIndex] = job
		}
	}

	selected := make([]batchfactoryv11.MergeJob, 0, requiredCount)
	for index := 0; index < requiredCount; index++ {
		job, ok := newestByOpening[index]
		if !ok {
			if index == 0 {
				return nil, fmt.Errorf("当前书没有可上传的原开头最终合成视频")
			}
			return nil, fmt.Errorf("当前书缺少换开头 %d 的最终合成视频", index)
		}
		selected = append(selected, job)
	}
	return selected, nil
}

func publish121Content(book batchfactoryv11.Book, workingFrontContent string) []byte {
	source := book.SourceText
	if source == "" {
		source = book.TxtText
	}
	editedLines := make([]string, 0)
	for _, line := range strings.Split(workingFrontContent, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			editedLines = append(editedLines, line)
		}
	}
	if len(editedLines) == 0 {
		return []byte(source)
	}

	limit := 5
	switch value := book.SourceMetadata["contentRangeLines"].(type) {
	case float64:
		limit = int(value)
	case int:
		limit = value
	case int64:
		limit = int(value)
	case json.Number:
		if parsed, err := value.Int64(); err == nil {
			limit = int(parsed)
		}
	}
	if limit < 1 {
		limit = 1
	} else if limit > 500 {
		limit = 500
	}

	sourceLines := strings.Split(source, "\n")
	replacement, seen := 0, 0
	for index, line := range sourceLines {
		if strings.TrimSpace(line) == "" || seen >= limit {
			continue
		}
		seen++
		if replacement >= len(editedLines) {
			continue
		}
		sourceLines[index] = editedLines[replacement]
		replacement++
	}
	if replacement == 0 {
		return []byte(strings.Join(editedLines, "\n"))
	}
	return []byte(strings.Join(sourceLines, "\n"))
}

// errPublishUnconfirmed 表示视频与正文已上传到视频管理系统，但回读确认失败。
// 调用方应把该书写成“已上传待确认”（stage=upload_confirm），重试时优先回读，
// 绝不无条件重传，避免在目标站产生重复作品。
var errPublishUnconfirmed = errors.New("视频管理系统已上传但回读确认失败")

// submissionConfirmedForBook 判断回读列表里是否真的存在该书。GetSubmission 的
// 传输层成功不代表书已入库，空列表同样需要重新提交。
func submissionConfirmedForBook(raw any, bookID string) bool {
	if raw == nil || strings.TrimSpace(bookID) == "" {
		return false
	}
	encoded, err := json.Marshal(raw)
	if err != nil {
		return false
	}
	return strings.Contains(string(encoded), bookID)
}

func (p *BatchFactory121Publisher) Submit(ctx context.Context, owner, batchID, bookID string, input map[string]any) (map[string]any, error) {
	if p == nil || p.Store == nil || p.Connector == nil {
		return nil, fmt.Errorf("视频管理系统上传服务未配置")
	}
	session := readUploadTargetSession(p.UsersDir, owner)
	cookie := strings.TrimSpace(toString(session["cookie"]))
	if cookie == "" {
		return nil, fmt.Errorf("请先保存账号密码并登录目标站")
	}
	targetSession := videomanagementconnector.Session{Cookie: cookie, TargetUsername: toString(session["targetUsername"])}
	batch, err := p.Store.GetBatch(ctx, owner, batchID)
	if err != nil {
		return nil, err
	}
	var book batchfactoryv11.Book
	for _, candidate := range batch.Books {
		if candidate.ID == bookID {
			book = candidate
			break
		}
	}
	if book.ID == "" {
		return nil, batchfactoryv11.ErrNotFound
	}
	if strings.TrimSpace(book.BookID) == "" {
		return nil, fmt.Errorf("当前小说缺少来源书号")
	}
	if !publish121BookID.MatchString(strings.TrimSpace(book.BookID)) {
		return nil, fmt.Errorf("121 上传文件名只支持纯数字书号或 UUID 格式")
	}
	metadata := book.SourceMetadata
	// 发布幂等三态：uploaded 直接短路；upload_pending_confirm 先回读，确认成功
	// 补写 uploaded，确认不存在才带已持久化的视频 heads 重新提交正文。
	force := input["force"] == true || strings.EqualFold(toString(input["force"]), "true")
	submitStatus := strings.TrimSpace(toString(metadata["websiteSubmitStatus"]))
	if submitStatus == "uploaded" && !force {
		return map[string]any{"status": "uploaded", "bookId": book.BookID, "internalBookId": book.ID, "skipped": true}, nil
	}
	pendingHeads := []string{}
	if submitStatus == "upload_pending_confirm" && !force {
		confirmed, readErr := p.Connector.GetSubmission(ctx, targetSession, book.BookID)
		if readErr != nil {
			return nil, fmt.Errorf("%w：%v", errPublishUnconfirmed, readErr)
		}
		if submissionConfirmedForBook(confirmed.ReadbackRaw, book.BookID) {
			uploadedAt := time.Now().UTC().Format(time.RFC3339Nano)
			confirmedMetadata := clonePublishMetadata(metadata)
			confirmedMetadata["websiteSubmitStatus"] = "uploaded"
			confirmedMetadata["publishStatus"] = "uploaded"
			confirmedMetadata["websiteSubmitProgress"] = map[string]any{
				"status":    "uploaded",
				"phase":     "completed",
				"message":   "视频管理系统回读确认成功（重试补齐）",
				"updatedAt": uploadedAt,
			}
			if _, err := p.Store.UpdateBookMetadata(ctx, owner, batchID, bookID, batchfactoryv11.UpdateBookMetadataInput{Metadata: confirmedMetadata, ExpectedRevision: book.Revision}); err != nil {
				return nil, fmt.Errorf("回读已成功，但保存上传状态失败：%w", err)
			}
			return map[string]any{
				"status":         "uploaded",
				"bookId":         book.BookID,
				"internalBookId": book.ID,
				"uploadedAt":     uploadedAt,
				"readback":       mapOrEmpty(confirmed.ReadbackRaw),
			}, nil
		}
		if raw, ok := metadata["websiteSubmitHeads"].([]any); ok {
			for _, item := range raw {
				pendingHeads = append(pendingHeads, fmt.Sprint(item))
			}
		}
	}
	platform := firstText(metadata["platformId"], metadata["platform_id"], book.Platform)
	if mapped := map[string]string{"黑岩付费": "1", "番茄付费": "2", "七猫付费": "3", "点众付费": "4", "阅文付费": "6", "番茄免费": "7", "知乎付费": "15", "掌阅付费": "20", "卓越付费": "26", "九州书城": "29", "掌文付费": "31"}[platform]; mapped != "" {
		platform = mapped
	}
	gender := strings.TrimSuffix(firstText(metadata["gender"], metadata["gender_name"]), "频")
	style := firstText(metadata["style"], metadata["styleName"], metadata["style_name"])
	if platform == "" || gender == "" || style == "" || style == "待判定" || style == "未判定" {
		return nil, fmt.Errorf("当前小说尚未完成书城、男女频和风格识别")
	}
	genderID := map[string]string{"男": "1", "女": "2"}[gender]
	styleID := map[string]string{"古风虐文": "101", "古风甜文": "102", "古风通用": "103", "年代虐文": "201", "年代甜文": "202", "年代通用": "203", "现代虐文": "301", "现代甜文": "302", "现代悬疑": "303", "现代通用": "304", "男频都市": "305", "现代女主": "306", "玄幻": "307", "历史": "308", "爆款BGM": "309", "家庭奇葩": "310", "家庭伤感": "311", "职场打脸": "312"}[style]
	if genderID == "" || styleID == "" {
		return nil, fmt.Errorf("视频管理系统不支持当前男女频或风格：%s/%s", gender, style)
	}
	settings := batchfactoryv11.ResolveSettings(batch.SettingsState.Patch, book.SettingsState.Patch)
	advanced := map[string]string{"tl5": "0", "ziti": "1", "zitidx": "62", "biaohong": "", "biaohong_reuse": "", "keywords": "", "jieya_num": "4", "jieya_ai_head": "3", "jieya_speed": "1.7", "jieya_pitch": "0", "font_color_styles": "[1]", "gunping_num": "4", "gunping_speed": "1"}
	if raw, ok := settings["publishSettings"]; ok {
		var publish map[string]any
		_ = json.Unmarshal(raw, &publish)
		for key, targetKey := range map[string]string{"jieyaNum": "jieya_num", "jieyaSpeed": "jieya_speed", "jieyaPitch": "jieya_pitch", "gunpingNum": "gunping_num", "gunpingSpeed": "gunping_speed", "tl5": "tl5", "ziti": "ziti", "zitidx": "zitidx", "jieyaAiHead": "jieya_ai_head"} {
			if value, ok := publish[key]; ok {
				advanced[targetKey] = fmt.Sprint(value)
			}
		}
	}
	organization := publishOrganization(input, settings)
	if organization == "" {
		return nil, fmt.Errorf("请选择组织归属")
	}
	fields := map[string]string{"platform_id": platform, "organization": organization, "gender": genderID, "style": styleID}
	for key, value := range advanced {
		fields[key] = value
	}
	workingFrontContent := book.WorkingFrontContent
	draft, err := p.Store.GetDraft(ctx, owner, "working-front:"+book.ID, "working-front-content", batchID)
	if err == nil {
		workingFrontContent = draft.Content
	} else if !errors.Is(err, batchfactoryv11.ErrNotFound) {
		return nil, err
	}
	content := publish121Content(book, workingFrontContent)
	if len(content) == 0 {
		return nil, fmt.Errorf("当前小说没有可上传的正文")
	}
	mergeJobs, ok := p.Store.(batchfactoryv11.MergeRepository)
	if !ok {
		return nil, fmt.Errorf("最终合成记录服务未配置")
	}
	jobs, err := mergeJobs.ListMergeJobs(ctx, owner, batchID)
	if err != nil {
		return nil, err
	}
	selectedJobs, err := selectPublishMergeJobs(batch, book, jobs)
	if err != nil {
		return nil, err
	}
	videoUploader, ok := p.Connector.(videomanagementconnector.VideoUploadService)
	if !ok {
		return nil, fmt.Errorf("视频管理系统 Go 连接器不支持视频上传")
	}
	type publishVideo struct {
		job      batchfactoryv11.MergeJob
		filename string
		content  []byte
	}
	videos := make([]publishVideo, 0, len(selectedJobs))
	videoFiles := make([]string, 0, len(selectedJobs))
	mergeJobIDs := make([]string, 0, len(selectedJobs))
	openingVariantIndexes := make([]int, 0, len(selectedJobs))
	for _, job := range selectedJobs {
		video, readErr := p.readMerge(ctx, job.OutputURL)
		if readErr != nil {
			return nil, fmt.Errorf("读取开头 %d 最终合成视频失败：%w", job.OpeningVariantIndex, readErr)
		}
		videoBytes, readErr := io.ReadAll(io.LimitReader(video, 512<<20))
		closeErr := video.Close()
		if readErr != nil || closeErr != nil || len(videoBytes) == 0 {
			return nil, fmt.Errorf("开头 %d 最终合成视频为空", job.OpeningVariantIndex)
		}
		filename := book.BookID + ".mp4"
		if len(selectedJobs) > 1 {
			filename = fmt.Sprintf("%s-opening-%d.mp4", book.BookID, job.OpeningVariantIndex)
		}
		videos = append(videos, publishVideo{job: job, filename: filename, content: videoBytes})
		videoFiles = append(videoFiles, filename)
		mergeJobIDs = append(mergeJobIDs, job.ID)
		openingVariantIndexes = append(openingVariantIndexes, job.OpeningVariantIndex)
	}
	heads := make([]string, 0, len(videos))
	if len(pendingHeads) == len(videos) {
		// 上一轮已上传成功、只是回读失败：直接复用持久化的 heads，不重复上传。
		heads = append(heads, pendingHeads...)
	} else {
		for _, video := range videos {
			head, uploadErr := videoUploader.UploadVideo(ctx, targetSession, video.filename, video.content)
			if uploadErr != nil {
				return nil, uploadErr
			}
			heads = append(heads, head)
		}
		// 视频全部上传成功后先落盘“待确认”，再提交正文容器；此后任何失败重试
		// 都会走回读优先路径，不会无条件重传。
		pendingMetadata := clonePublishMetadata(metadata)
		pendingMetadata["websiteSubmitStatus"] = "upload_pending_confirm"
		pendingMetadata["websiteSubmitHeads"] = heads
		pendingMetadata["websiteSubmitProgress"] = map[string]any{
			"status":    "uploading",
			"phase":     "confirm_pending",
			"message":   "视频已上传，等待视频管理系统回读确认",
			"updatedAt": time.Now().UTC().Format(time.RFC3339Nano),
		}
		if _, err := p.Store.UpdateBookMetadata(ctx, owner, batchID, bookID, batchfactoryv11.UpdateBookMetadataInput{Metadata: pendingMetadata, ExpectedRevision: book.Revision}); err != nil {
			return nil, fmt.Errorf("视频管理系统上传进度保存失败：%w", err)
		}
		book.Revision++
	}
	headVideoJSON, err := marshalPublishedVideoHeads(heads)
	if err != nil {
		return nil, err
	}
	fields["jieya_ai_head_video"] = headVideoJSON
	fields["jieya_ai_head_video_mode"] = "0"
	fields["jieya_ai_head_horizontal_flip"] = "0"
	submission, err := p.Connector.Submit(ctx, targetSession, videomanagementconnector.SubmissionInput{BookID: book.BookID, Filename: book.BookID + ".txt", Content: content, FormFields: fields})
	if err != nil {
		return nil, fmt.Errorf("视频管理系统上传失败：%w", err)
	}
	remote := mapOrEmpty(submission.ReadbackRaw)
	confirmed, err := p.Connector.GetSubmission(ctx, targetSession, book.BookID)
	if err != nil {
		return nil, fmt.Errorf("%w：%v", errPublishUnconfirmed, err)
	}
	readback := mapOrEmpty(confirmed.ReadbackRaw)
	uploadedAt := time.Now().UTC().Format(time.RFC3339Nano)
	// The Go publisher is the single upload owner. Persist the confirmed
	// response back to the book so the workbench can show the same success
	// state after a refresh and will not treat an already uploaded book as new.
	metadata = clonePublishMetadata(metadata)
	metadata["websiteSubmitStatus"] = "uploaded"
	metadata["publishStatus"] = "uploaded"
	metadata["websiteSubmitProgress"] = map[string]any{
		"status":    "uploaded",
		"phase":     "completed",
		"message":   "视频管理系统已接收并回读确认",
		"updatedAt": uploadedAt,
	}
	history := publishHistory(metadata["websiteSubmitHistory"])
	history = append(history, map[string]any{
		"status":                "uploaded",
		"phase":                 "completed",
		"detail":                "视频管理系统已接收并回读确认",
		"submittedAt":           uploadedAt,
		"bookId":                book.BookID,
		"videoCount":            len(videoFiles),
		"openingVariantIndexes": openingVariantIndexes,
		"mergeJobIds":           mergeJobIDs,
		"videoFiles":            videoFiles,
	})
	metadata["websiteSubmitHistory"] = history
	if _, err := p.Store.UpdateBookMetadata(ctx, owner, batchID, bookID, batchfactoryv11.UpdateBookMetadataInput{Metadata: metadata, ExpectedRevision: book.Revision}); err != nil {
		return nil, fmt.Errorf("上传已成功，但保存上传状态失败：%w", err)
	}
	return map[string]any{
		"status":                "uploaded",
		"bookId":                book.BookID,
		"internalBookId":        book.ID,
		"sourceTextFile":        book.BookID + ".txt",
		"aiHeadVideoFile":       videoFiles[0],
		"aiHeadVideoFiles":      videoFiles,
		"videoCount":            len(videoFiles),
		"openingVariantIndexes": openingVariantIndexes,
		"mergeJobIds":           mergeJobIDs,
		"uploadedAt":            uploadedAt,
		"response":              remote,
		"readback":              readback,
	}, nil
}

func marshalPublishedVideoHeads(heads []string) (string, error) {
	values := make([]map[string]any, 0, len(heads))
	for _, head := range heads {
		var value map[string]any
		if err := json.Unmarshal([]byte(strings.TrimSpace(head)), &value); err != nil || len(value) == 0 {
			return "", fmt.Errorf("视频管理系统返回了无效的视频对象")
		}
		values = append(values, value)
	}
	raw, err := json.Marshal(values)
	if err != nil {
		return "", fmt.Errorf("整理视频上传结果失败：%w", err)
	}
	return string(raw), nil
}

func clonePublishMetadata(source map[string]any) map[string]any {
	out := make(map[string]any, len(source)+5)
	for key, value := range source {
		out[key] = value
	}
	return out
}

func publishHistory(value any) []any {
	switch history := value.(type) {
	case []any:
		return append([]any(nil), history...)
	case []map[string]any:
		out := make([]any, 0, len(history))
		for _, item := range history {
			out = append(out, item)
		}
		return out
	default:
		return []any{}
	}
}

func publishOrganization(input map[string]any, settings map[string]json.RawMessage) string {
	organization := strings.TrimSpace(toString(input["organization"]))
	if organization != "" {
		return organization
	}
	if raw, ok := settings["publishSettings"]; ok {
		var publish map[string]any
		if json.Unmarshal(raw, &publish) == nil {
			for _, value := range []any{publish["organization"], publish["organizationId"], publish["organization_id"], publish["orgId"], publish["org_id"]} {
				if value != nil {
					organization = strings.TrimSpace(fmt.Sprint(value))
					if organization != "" {
						break
					}
				}
			}
		}
	}
	return organization
}

func (p *BatchFactory121Publisher) readMerge(ctx context.Context, output string) (io.ReadCloser, error) {
	output = strings.TrimSpace(output)
	if strings.HasPrefix(output, "http://") || strings.HasPrefix(output, "https://") {
		response, err := (&http.Client{Timeout: 120 * time.Second}).Get(output)
		if err != nil {
			return nil, err
		}
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			response.Body.Close()
			return nil, fmt.Errorf("视频地址返回 HTTP %d", response.StatusCode)
		}
		return response.Body, nil
	}
	artifactID := path.Base(output)
	artifactID = strings.TrimSuffix(artifactID, ".mp4")
	if p.MergedFiles != nil {
		if reader, err := p.MergedFiles.Open(ctx, artifactID); err == nil {
			return reader, nil
		}
	}
	if p.LocalFiles != nil {
		return p.LocalFiles.Open(artifactID + ".mp4")
	}
	return nil, fmt.Errorf("找不到合成视频文件")
}

func publish121OrganizationsHandler(root *http.ServeMux, middleware BridgeAuth, usersDir string, targets ...videomanagementconnector.Service) {
	root.Handle("GET /api/batch-factory/v11/publish-121/organizations", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"ok": false, "error": "unauthorized"})
			return
		}
		session := readUploadTargetSession(usersDir, identity.Username)
		cookie := strings.TrimSpace(toString(session["cookie"]))
		if cookie == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "请先保存账号密码并登录目标站"})
			return
		}
		if len(targets) == 0 || targets[0] == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": "视频管理系统 Go 连接器未配置"})
			return
		}
		catalogs, err := targets[0].RefreshCatalogs(req.Context(), videomanagementconnector.Session{Cookie: cookie, TargetUsername: toString(session["targetUsername"])})
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "error": "视频管理系统组织目录读取失败"})
			return
		}
		payload := any(catalogs.Raw)
		organizations := normalizeOrganizationOptions(payload)
		if len(organizations) == 0 {
			writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "error": "视频管理系统未返回可用组织目录"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "organizations": organizations})
	})))
}

func registerBatchFactory121Publish(root *http.ServeMux, middleware BridgeAuth, publisher *BatchFactory121Publisher) {
	if root == nil || publisher == nil {
		return
	}
	handler := middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input map[string]any
		if !decodeJSON(w, req, &input) {
			return
		}
		result, err := publisher.Submit(req.Context(), identity.Username, req.PathValue("batchId"), req.PathValue("bookId"), input)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "result": result})
	}))
	root.Handle("POST /api/batch-factory/v11/batches/{batchId}/books/{bookId}/publish-121", handler)
	root.Handle("POST /api/batch-factory/v12/batches/{batchId}/books/{bookId}/publish-121", handler)
}
