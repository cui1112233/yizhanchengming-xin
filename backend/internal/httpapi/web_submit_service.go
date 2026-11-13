package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"qiantie/backend/internal/novelfetchworkshop"
	"qiantie/backend/internal/videomanagementconnector"
)

type webSubmitRequest struct {
	Queue        novelfetchworkshop.ProcessQueue
	TaskIDs      []string
	Versions     []string
	OperationID  string
	ReadbackOnly bool
}

type webSubmitExecution struct {
	OperationID string
	Results     []any
	Success     int
	Failed      int
}

type webSubmitRequestError struct {
	Status  int
	Message string
	Code    string
}

func (err *webSubmitRequestError) Error() string { return err.Message }

// A book may already have uploads from previous days. An HTTP success or a
// matching book alone cannot confirm the current send attempt.
func submissionMatchesAttempt(receipt videomanagementconnector.Submission, expectedID, startedAt string) bool {
	record := mapOrEmpty(receipt.ReadbackRaw)
	bookID := firstText(record["bookid"])
	if bookID == "" {
		return strings.TrimSpace(receipt.ID) != ""
	} // Other connectors return an explicit confirmed receipt.
	if expectedID != "" && expectedID != bookID {
		return receipt.ID == expectedID
	}
	started, err := time.Parse(time.RFC3339Nano, startedAt)
	if err != nil {
		return false
	}
	created, err := time.ParseInLocation(time.DateTime, firstText(record["addtime"]), time.FixedZone("Asia/Shanghai", 8*60*60))
	return err == nil && !created.Before(started.Truncate(time.Second))
}

func executeWebSubmit(ctx context.Context, owner, usersDir string, store novelfetchworkshop.Store, request webSubmitRequest, genderJudgeResolver func(context.Context, string) (novelfetchworkshop.AIGenderJudge, error), targets ...videomanagementconnector.Service) (webSubmitExecution, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Minute)
	defer cancel()
	if request.Queue == nil {
		request.Queue, _ = ctx.Value(submitQueueKey{}).(novelfetchworkshop.ProcessQueue)
	}
	operationID := strings.TrimSpace(request.OperationID)
	if operationID == "" {
		operationID = fmt.Sprintf("web-submit-%d", time.Now().UTC().UnixNano())
	}
	result := webSubmitExecution{OperationID: operationID, Results: []any{}}
	release, guardErr := acquireWorkshopSubmit(ctx, request.Queue, owner, "")
	if guardErr != nil {
		recordWorkshopSubmitBlocked(ctx, store, owner, request.TaskIDs, operationID, "queued", guardErr.Error())
		return result, &webSubmitRequestError{Status: http.StatusConflict, Message: guardErr.Error()}
	}
	defer release()
	settings, err := store.GetConfig(ctx, owner)
	if err != nil {
		return result, &webSubmitRequestError{Status: http.StatusInternalServerError, Message: "读取提交配置失败"}
	}
	web := mapOrEmpty(mapOrEmpty(settings["appConfig"])["web_submit"])
	if web["enabled"] != true {
		recordWorkshopSubmitBlocked(ctx, store, owner, request.TaskIDs, operationID, "waiting_configuration", "网站提交尚未启用；请先保存账号并勾选启用网站提交")
		return result, &webSubmitRequestError{Status: http.StatusBadRequest, Message: "网站提交尚未启用；请先保存账号并勾选启用网站提交"}
	}
	session := readUploadTargetSession(usersDir, owner)
	cookie := strings.TrimSpace(toString(session["cookie"]))
	if cookie == "" {
		recordWorkshopSubmitBlocked(ctx, store, owner, request.TaskIDs, operationID, "waiting_login", "视频管理系统未登录，请先登录后再提交")
		return result, &webSubmitRequestError{Status: http.StatusBadRequest, Message: "请先在网站提交页保存账号密码并登录目标站", Code: "UPLOAD_LOGIN_REQUIRED"}
	}
	versions := cleanStrings(request.Versions)
	var pauseTarget videomanagementconnector.Service
	if len(targets) > 0 {
		pauseTarget = targets[0]
	}
	if err := checkWorkshopSubmitPause(ctx, store, owner, session, pauseTarget); err != nil {
		recordWorkshopSubmitBlocked(ctx, store, owner, request.TaskIDs, operationID, "waiting_login", err.Error())
		return result, err
	}
	if len(versions) == 0 {
		versions = cleanStrings(anyStrings(web["submit_versions"]))
	}
	if len(versions) == 0 {
		versions = []string{"ai1"}
	}
	if len(request.TaskIDs) == 0 {
		return result, &webSubmitRequestError{Status: http.StatusBadRequest, Message: "没有可提交的任务，请先勾选任务"}
	}
	skipSubmitted := webSubmitSkipSubmitted(web)
	minTextChars := intValue(web["min_text_chars"])
	retryTimes := intValue(web["retry_times"])
	if retryTimes < 0 {
		retryTimes = 0
	}
	if retryTimes > 10 {
		retryTimes = 10
	}
	for _, rawID := range request.TaskIDs {
		bookID := strings.TrimSpace(rawID)
		doc, docErr := store.GetDocument(ctx, owner, bookID)
		if docErr != nil {
			result.Results = append(result.Results, map[string]any{"bookId": bookID, "status": "error", "error": "任务不存在"})
			result.Failed++
			continue
		}
		meta := doc.Meta
		if meta == nil {
			meta = map[string]any{}
		}
		if genderNeedsResolution(doc.Gender) && genderJudgeResolver != nil {
			if judge, judgeErr := genderJudgeResolver(ctx, owner); judgeErr == nil {
				resolveSubmissionGender(ctx, &doc, judge)
				meta = doc.Meta
			}
		}
		done := anyStrings(meta["siteSubmitDoneVersions"])
		if len(done) == 0 {
			done = anyStrings(meta["site_submit_done_versions"])
		}
		for _, version := range versions {
			row := map[string]any{"bookId": bookID, "version": version, "status": "error", "operation_id": operationID, "retryable": false, "readback_status": "not_started"}
			failedVersion := func(message string, externalRequested bool) {
				row["error"] = message
				row["retryable"] = true
				markSubmissionFailure(meta, version, message)
				doc.Logs = append(doc.Logs, map[string]any{"event": "site_submit", "version": version, "status": "failed", "error": message, "time": time.Now().UTC().Format(time.RFC3339), "externalRequested": externalRequested})
				result.Results = append(result.Results, row)
				result.Failed++
			}
			uncertainVersion := func(message, submissionID string) {
				row["status"] = "pending_confirmation"
				row["error"] = message
				row["readback_status"] = "pending_confirmation"
				meta["siteSubmitStatus"] = "pending_confirmation"
				meta["siteSubmitFailedVersions"] = uniqueStrings(append(anyStrings(meta["siteSubmitFailedVersions"]), version))
				errors := mapOrEmpty(meta["siteSubmitFailedErrors"])
				errors[version] = message
				meta["siteSubmitFailedErrors"] = errors
				receipts := mapOrEmpty(meta["siteSubmitPendingConfirmation"])
				receipts[version] = submissionID
				meta["siteSubmitPendingConfirmation"] = receipts
				doc.Logs = append(doc.Logs, map[string]any{"event": "site_submit", "operation_id": operationID, "version": version, "status": "pending_confirmation", "error": message, "time": time.Now().UTC().Format(time.RFC3339), "externalRequested": true, "submissionId": submissionID})
				result.Results = append(result.Results, row)
				result.Failed++
			}
			recordSuccess := func(submission, confirmed videomanagementconnector.Submission) {
				row["status"] = "submitted"
				row["submissionId"] = firstText(confirmed.ID, submission.ID)
				row["verified"] = true
				row["retryable"] = false
				row["readback_status"] = "confirmed"
				row["message"] = firstText(confirmed.Message, "视频管理系统已回读确认")
				row["error"] = nil
				result.Results = append(result.Results, row)
				result.Success++
				meta["siteSubmitStatus"] = "submitted"
				done = append(done, version)
				meta["siteSubmitDoneVersions"] = uniqueStrings(done)
				meta["siteSubmitFailedVersions"] = removeString(anyStrings(meta["siteSubmitFailedVersions"]), version)
				if errors := mapOrEmpty(meta["siteSubmitFailedErrors"]); len(errors) > 0 {
					delete(errors, version)
					meta["siteSubmitFailedErrors"] = errors
				}
				if receipts := mapOrEmpty(meta["siteSubmitPendingConfirmation"]); len(receipts) > 0 {
					delete(receipts, version)
					meta["siteSubmitPendingConfirmation"] = receipts
				}
				doc.Logs = append(doc.Logs, map[string]any{"event": "site_submit", "operation_id": operationID, "version": version, "status": "submitted", "time": time.Now().UTC().Format(time.RFC3339), "externalRequested": true, "remote_receipt": map[string]any{"verified": true, "submission_id": firstText(confirmed.ID, submission.ID), "upload_success_count": 1, "remote_record": confirmed.ReadbackRaw}})
			}
			if skipSubmitted && containsVersion(done, version) {
				row["status"] = "skipped"
				row["error"] = "该版本已确认提交"
				result.Results = append(result.Results, row)
				continue
			}
			targetSession := videomanagementconnector.Session{Cookie: cookie, TargetUsername: toString(session["targetUsername"])}
			_, needsReadback := mapOrEmpty(meta["siteSubmitPendingConfirmation"])[version]
			if !needsReadback && len(mapOrEmpty(meta["siteSubmitPendingConfirmation"])) == 0 {
				needsReadback = strings.EqualFold(firstText(meta["siteSubmitStatus"], meta["site_submit_status"]), "pending_confirmation") && containsVersion(anyStrings(meta["siteSubmitFailedVersions"]), version)
			}
			if request.ReadbackOnly && !needsReadback {
				row["error"] = "该版本没有待核对的提交记录"
				result.Results = append(result.Results, row)
				result.Failed++
				continue
			}
			if needsReadback {
				if len(targets) == 0 || targets[0] == nil {
					failedVersion("视频管理系统 Go 连接器未执行回读确认", false)
					continue
				}
				confirmed, readbackErr := targets[0].GetSubmission(ctx, targetSession, bookID)
				if readbackErr != nil {
					uncertainVersion("上传已接收，正在等待目标站回读确认", firstText(mapOrEmpty(meta["siteSubmitPendingConfirmation"])[version]))
					continue
				}
				attempt := mapOrEmpty(mapOrEmpty(meta["siteSubmitAttempts"])[version])
				if !submissionMatchesAttempt(confirmed, firstText(mapOrEmpty(meta["siteSubmitPendingConfirmation"])[version], bookID), firstText(attempt["started_at"])) {
					uncertainVersion("只查到旧记录，尚未确认本次提交；不会重复上传", firstText(mapOrEmpty(meta["siteSubmitPendingConfirmation"])[version], bookID))
					continue
				}
				recordSuccess(confirmed, confirmed)
				continue
			}
			if styleNeedsResolution(meta) {
				failedVersion(webSubmitUnclassifiedStyleError, false)
				continue
			}
			content, bodyErr := store.GetBody(ctx, owner, bookID, version)
			if bodyErr != nil || strings.TrimSpace(content.Content) == "" {
				failedVersion("未找到正文版本", false)
				continue
			}
			if minTextChars > 0 && len([]rune(content.Content)) < minTextChars {
				failedVersion("正文少于最少字数", false)
				continue
			}
			if strings.HasPrefix(strings.ToLower(version), "ai") {
				if err := novelfetchworkshop.ValidateRewriteProse(content.Content); err != nil {
					failedVersion("AI 正文复检未通过："+err.Error()+"；请重新生成该版本后提交", false)
					row["retryable"] = false
					continue
				}
			}
			profile := webSubmitProfile(web, version)
			row["profile"] = firstText(profile["id"], profile["name"], mapOrEmpty(web["version_profiles"])[version], mapOrEmpty(web["profile_bindings"])[version], web["selected_profile"])
			platform, _ := strconv.Atoi(firstText(profile["platform_id"], profile["platformId"], meta["platformId"], meta["platform_id"]))
			gender, style := submissionGenderStyle(doc)
			fields, fieldErr := buildUploadFieldsCompat(platform, gender, style, webSubmitAdvanced(web, profile), map[string]any{"organization": webSubmitOrganization(web, profile)})
			if fieldErr != nil {
				failedVersion(fieldErr.Error(), false)
				continue
			}
			if len(targets) == 0 || targets[0] == nil {
				failedVersion("视频管理系统 Go 连接器未执行上传", false)
				continue
			}
			contentHash := sha256.Sum256([]byte(content.Content))
			attempts := mapOrEmpty(meta["siteSubmitAttempts"])
			attempts[version] = map[string]any{"operation_id": operationID, "content_sha256": hex.EncodeToString(contentHash[:]), "started_at": time.Now().UTC().Format(time.RFC3339Nano)}
			meta["siteSubmitAttempts"] = attempts
			meta["siteSubmitStatus"] = "pending_confirmation"
			meta["siteSubmitFailedVersions"] = uniqueStrings(append(anyStrings(meta["siteSubmitFailedVersions"]), version))
			pendingErrors := mapOrEmpty(meta["siteSubmitFailedErrors"])
			pendingErrors[version] = "提交请求已开始，等待目标站确认"
			meta["siteSubmitFailedErrors"] = pendingErrors
			pendingReceipts := mapOrEmpty(meta["siteSubmitPendingConfirmation"])
			pendingReceipts[version] = bookID
			meta["siteSubmitPendingConfirmation"] = pendingReceipts
			if err := store.PutDocument(ctx, owner, doc); err != nil {
				return result, fmt.Errorf("提交前保存尝试记录失败: %w", err)
			}
			var submission videomanagementconnector.Submission
			var submitErr error
			for attempt := 0; attempt <= retryTimes; attempt++ {
				submission, submitErr = targets[0].Submit(ctx, targetSession, videomanagementconnector.SubmissionInput{BookID: bookID, Filename: bookID + ".txt", Content: []byte(content.Content), FormFields: fields})
				if submitErr == nil {
					break
				}
				if !strings.Contains(strings.ToLower(submitErr.Error()), "rejected submission") {
					break
				}
			}
			if submitErr != nil {
				if errors.Is(submitErr, videomanagementconnector.ErrSessionExpired) || errors.Is(submitErr, videomanagementconnector.ErrSessionMissing) {
					delete(mapOrEmpty(meta["siteSubmitPendingConfirmation"]), version)
					markSubmissionFailure(meta, version, workshopLoginRequiredMessage)
					meta["siteSubmitStatus"] = "waiting_login"
					doc.Meta = meta
					doc.Logs = append(doc.Logs, map[string]any{"event": "site_submit", "operation_id": operationID, "version": version, "status": "waiting_login", "error": workshopLoginRequiredMessage, "externalRequested": true, "time": time.Now().UTC().Format(time.RFC3339Nano)})
					if err := store.PutDocument(ctx, owner, doc); err != nil {
						return result, err
					}
					if err := pauseWorkshopSubmitAccount(ctx, store, owner, session); err != nil {
						return result, err
					}
					result.Failed++
					return result, &webSubmitRequestError{Status: http.StatusConflict, Message: workshopLoginRequiredMessage, Code: "UPLOAD_LOGIN_REQUIRED"}
				}
				uncertainVersion("提交结果不确定，等待目标站回读确认", bookID)
				continue
			}
			confirmed, readbackErr := targets[0].GetSubmission(ctx, targetSession, bookID)
			if readbackErr != nil {
				uncertainVersion("上传已接收，但目标站回读确认失败", submission.ID)
				continue
			}
			if !submissionMatchesAttempt(confirmed, firstText(submission.ID, bookID), firstText(mapOrEmpty(attempts[version])["started_at"])) {
				uncertainVersion("只查到旧记录，尚未确认本次提交；不会重复上传", firstText(submission.ID, bookID))
				continue
			}
			recordSuccess(submission, confirmed)
		}
		if len(mapOrEmpty(meta["siteSubmitPendingConfirmation"])) > 0 {
			meta["siteSubmitStatus"] = "pending_confirmation"
		} else if len(anyStrings(meta["siteSubmitFailedVersions"])) > 0 {
			meta["siteSubmitStatus"] = "pending_upload"
		}
		doc.Meta = meta
		if err := store.PutDocument(ctx, owner, doc); err != nil {
			return result, fmt.Errorf("保存提交结果失败: %w", err)
		}
	}
	return result, nil
}
