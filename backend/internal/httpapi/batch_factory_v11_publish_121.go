package httpapi

import (
	"context"
	"encoding/json"
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

func (p *BatchFactory121Publisher) Submit(ctx context.Context, owner, batchID, bookID string, input map[string]any) (map[string]any, error) {
	if p == nil || p.Store == nil || p.Connector == nil {
		return nil, fmt.Errorf("视频管理系统上传服务未配置")
	}
	session := readUploadTargetSession(p.UsersDir, owner)
	cookie := strings.TrimSpace(toString(session["cookie"]))
	if cookie == "" {
		return nil, fmt.Errorf("请先保存账号密码并登录目标站")
	}
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
	content := []byte(book.SourceText)
	if len(content) == 0 {
		content = []byte(book.TxtText)
	}
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
	var selected batchfactoryv11.MergeJob
	for _, job := range jobs {
		if job.BookID == bookID && job.Status == batchfactoryv11.MergeSucceeded && strings.TrimSpace(job.OutputURL) != "" && (selected.ID == "" || job.UpdatedAt.After(selected.UpdatedAt)) {
			selected = job
		}
	}
	if selected.ID == "" {
		return nil, fmt.Errorf("当前书没有可上传的最终合成视频")
	}
	video, err := p.readMerge(ctx, selected.OutputURL)
	if err != nil {
		return nil, fmt.Errorf("读取最终合成视频失败：%w", err)
	}
	defer video.Close()
	videoBytes, err := io.ReadAll(io.LimitReader(video, 512<<20))
	if err != nil || len(videoBytes) == 0 {
		return nil, fmt.Errorf("最终合成视频为空")
	}
	videoUploader, ok := p.Connector.(videomanagementconnector.VideoUploadService)
	if !ok {
		return nil, fmt.Errorf("视频管理系统 Go 连接器不支持视频上传")
	}
	targetSession := videomanagementconnector.Session{Cookie: cookie, TargetUsername: toString(session["targetUsername"])}
	head, err := videoUploader.UploadVideo(ctx, targetSession, book.BookID+".mp4", videoBytes)
	if err != nil {
		return nil, err
	}
	fields["jieya_ai_head_video"] = fmt.Sprintf("[%s]", head)
	fields["jieya_ai_head_video_mode"] = "0"
	fields["jieya_ai_head_horizontal_flip"] = "0"
	submission, err := p.Connector.Submit(ctx, targetSession, videomanagementconnector.SubmissionInput{BookID: book.BookID, Filename: book.BookID + ".txt", Content: content, FormFields: fields})
	if err != nil {
		return nil, fmt.Errorf("视频管理系统上传失败：%w", err)
	}
	remote := mapOrEmpty(submission.ReadbackRaw)
	confirmed, err := p.Connector.GetSubmission(ctx, targetSession, book.BookID)
	if err != nil {
		return nil, fmt.Errorf("视频管理系统回读失败：%w", err)
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
		"status":      "uploaded",
		"phase":       "completed",
		"detail":      "视频管理系统已接收并回读确认",
		"submittedAt": uploadedAt,
		"bookId":      book.BookID,
	})
	metadata["websiteSubmitHistory"] = history
	if _, err := p.Store.UpdateBookMetadata(ctx, owner, batchID, bookID, batchfactoryv11.UpdateBookMetadataInput{Metadata: metadata, ExpectedRevision: book.Revision}); err != nil {
		return nil, fmt.Errorf("上传已成功，但保存上传状态失败：%w", err)
	}
	return map[string]any{"status": "uploaded", "bookId": book.BookID, "internalBookId": book.ID, "sourceTextFile": book.BookID + ".txt", "aiHeadVideoFile": book.BookID + ".mp4", "uploadedAt": uploadedAt, "response": remote, "readback": readback}, nil
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
