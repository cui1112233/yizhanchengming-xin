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
	// Force 与参考实现一致：跳过“已确认提交”的保护，但不会绕过
	// executing/unknown 的回读保护，避免重复派发。
	Force bool
	// Manifest 在上传前落盘，逐项记录 ready/executing/unknown 等状态。
	Manifest *webSubmitManifest
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
	// 不再设置批次级总超时：单个外部请求由连接器的 120 秒超时兜底，
	// 崩溃恢复由锁心跳与 manifest stale 判定负责。
	if request.Queue == nil {
		request.Queue, _ = ctx.Value(submitQueueKey{}).(novelfetchworkshop.ProcessQueue)
	}
	operationID := strings.TrimSpace(request.OperationID)
	if operationID == "" {
		operationID = fmt.Sprintf("web-submit-%d", time.Now().UTC().UnixNano())
	}
	manifest := request.Manifest
	manifestItem := func(bookID, version, state, reason, message string) {
		if manifest != nil && strings.TrimSpace(version) != "" {
			manifest.setItem(bookID, version, state, reason, message)
		}
	}
	result := webSubmitExecution{OperationID: operationID, Results: []any{}}
	session := readUploadTargetSession(usersDir, owner)
	release, guardErr := acquireWorkshopSubmit(ctx, request.Queue, workshopExternalSubmitScope(session), "")
	if guardErr != nil {
		recordWorkshopSubmitBlocked(ctx, store, owner, request.TaskIDs, operationID, "queued", guardErr.Error())
		return result, &webSubmitRequestError{Status: http.StatusConflict, Message: guardErr.Error()}
	}
	defer release()
	// 拿到提交锁说明当前没有并发提交，可以安全收尾上次崩溃遗留的 executing 清单。
	recoverStaleWebSubmitManifests(usersDir, owner)
	settings, err := store.GetConfig(ctx, owner)
	if err != nil {
		return result, &webSubmitRequestError{Status: http.StatusInternalServerError, Message: "读取提交配置失败"}
	}
	web := mapOrEmpty(mapOrEmpty(settings["appConfig"])["web_submit"])
	if web["enabled"] != true {
		recordWorkshopSubmitBlocked(ctx, store, owner, request.TaskIDs, operationID, "waiting_configuration", "网站提交尚未启用；请先保存账号并勾选启用网站提交")
		return result, &webSubmitRequestError{Status: http.StatusBadRequest, Message: "网站提交尚未启用；请先保存账号并勾选启用网站提交"}
	}
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
		recordWorkshopSubmitBlocked(ctx, store, owner, request.TaskIDs, operationID, workshopSubmitBlockedStatus(err), err.Error())
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
	minFileSize := webSubmitMinFileSize(web)
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
		blockContract := func() error {
			meta["siteSubmitStatus"] = "waiting_configuration"
			meta["siteSubmitPauseReason"] = "site_contract_changed"
			meta["status"], meta["queueState"] = "waiting_configuration", "waiting_configuration"
			meta["siteSubmitExecutingVersions"] = []string{}
			doc.Meta = meta
			appendNovelFetchDocumentLog(&doc, "upload_contract_changed", workshopContractChangedMessage, time.Now())
			if err := store.PutDocument(ctx, owner, doc); err != nil {
				return err
			}
			if err := pauseWorkshopSubmitContract(ctx, store, owner, session); err != nil {
				return err
			}
			return workshopContractBlockedError()
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
		// 内部任务 ID 可能带账号前缀，上传/回读必须使用网站书号。
		uploadBookID, hasSiteBookID := webSubmitSourceBookID(meta, bookID)
		confirmedHashes := map[string][]string{}
		if hasSiteBookID && !request.Force {
			confirmedHashes = webSubmitConfirmedContentHashes(ctx, store, owner, bookID, done)
		}
		deferredStatus := ""
		for _, version := range versions {
			row := map[string]any{"bookId": bookID, "version": version, "status": "error", "operation_id": operationID, "retryable": false, "readback_status": "not_started"}
			clearExecuting := func() {
				meta["siteSubmitExecutingVersions"] = removeString(anyStrings(meta["siteSubmitExecutingVersions"]), version)
			}
			// markDeferred 与参考实现一致：分类/配置/正文未就绪属于“延后”，
			// 不是提交失败，不计入 failed_groups，也不会自动重复派发。
			markDeferred := func(status, reason, message string) {
				row["status"] = status
				row["reason"] = reason
				row["deferred"] = true
				row["error"] = message
				row["retryable"] = status != "waiting_generate" && status != "waiting_original" && status != "waiting_classify"
				if deferredStatus == "" {
					deferredStatus = status
				}
				meta["siteSubmitStatus"] = status
				doc.Logs = append(doc.Logs, map[string]any{"event": "site_submit", "operation_id": operationID, "version": version, "status": status, "reason": reason, "error": message, "time": time.Now().UTC().Format(time.RFC3339Nano), "externalRequested": false})
				result.Results = append(result.Results, row)
				manifestItem(bookID, version, status, reason, message)
			}
			if !hasSiteBookID {
				markDeferred("waiting_config", "invalid_site_book_id", "任务缺少合法的网站书号（纯数字），无法提交")
				continue
			}
			failedVersion := func(message string, externalRequested bool) {
				row["error"] = message
				row["retryable"] = true
				clearExecuting()
				markSubmissionFailure(meta, version, message)
				doc.Logs = append(doc.Logs, map[string]any{"event": "site_submit", "version": version, "status": "failed", "error": message, "time": time.Now().UTC().Format(time.RFC3339), "externalRequested": externalRequested})
				result.Results = append(result.Results, row)
				result.Failed++
				manifestItem(bookID, version, "failed", "", message)
			}
			uncertainVersion := func(message, submissionID string) {
				row["status"] = "pending_confirmation"
				row["error"] = message
				row["readback_status"] = "pending_confirmation"
				clearExecuting()
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
				manifestItem(bookID, version, "unknown", "remote_confirmation_required", message)
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
				clearExecuting()
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
				manifestItem(bookID, version, "submitted", "confirmed_receipt", "")
			}
			if skipSubmitted && !request.Force && containsVersion(done, version) {
				row["status"] = "skipped"
				row["reason"] = "confirmed_result"
				row["error"] = "该版本已确认提交"
				result.Results = append(result.Results, row)
				manifestItem(bookID, version, "skipped", "confirmed_result", "该版本已确认提交")
				continue
			}
			targetSession := videomanagementconnector.Session{Cookie: cookie, TargetUsername: toString(session["targetUsername"])}
			_, needsReadback := mapOrEmpty(meta["siteSubmitPendingConfirmation"])[version]
			if !needsReadback {
				// executing 表示上次上传已派发但没有留下结果；只能回读，不能重复上传。
				needsReadback = containsVersion(anyStrings(meta["siteSubmitExecutingVersions"]), version)
			}
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
				confirmed, readbackErr := targets[0].GetSubmission(ctx, targetSession, uploadBookID)
				if readbackErr != nil {
					uncertainVersion("上传已接收，正在等待目标站回读确认", firstText(mapOrEmpty(meta["siteSubmitPendingConfirmation"])[version]))
					if errors.Is(readbackErr, videomanagementconnector.ErrSiteContractChanged) {
						return result, blockContract()
					}
					continue
				}
				attempt := mapOrEmpty(mapOrEmpty(meta["siteSubmitAttempts"])[version])
				if !submissionMatchesAttempt(confirmed, firstText(mapOrEmpty(meta["siteSubmitPendingConfirmation"])[version], uploadBookID), firstText(attempt["started_at"])) {
					// 与批量原文改文系统源码对齐：目标站已存在该书的提交记录
					// 即视为已提交；时间戳和回执号只作为证据，不再阻塞任务。
					if strings.TrimSpace(firstText(confirmed.ID, mapOrEmpty(confirmed.ReadbackRaw)["bookid"])) == "" {
						uncertainVersion("只查到旧记录，尚未确认本次提交；不会重复上传", firstText(mapOrEmpty(meta["siteSubmitPendingConfirmation"])[version], uploadBookID))
						continue
					}
				}
				recordSuccess(confirmed, confirmed)
				continue
			}
			profile := webSubmitProfile(web, version)
			row["profile"] = firstText(profile["id"], profile["name"], mapOrEmpty(web["version_profiles"])[version], mapOrEmpty(web["profile_bindings"])[version], web["selected_profile"])
			platformText := firstText(profile["platform_id"], profile["platformId"], meta["platformId"], meta["platform_id"])
			if gaps := submissionClassificationGaps(platformText, submissionResolvedGender(doc), submissionResolvedStyle(meta)); len(gaps) > 0 {
				status := "waiting_config"
				if styleNeedsResolution(meta) {
					status = "waiting_classify"
				}
				markDeferred(status, status, submissionGapError(gaps))
				continue
			}
			content, bodyErr := store.GetBody(ctx, owner, bookID, version)
			if bodyErr != nil || strings.TrimSpace(content.Content) == "" {
				status := "waiting_generate"
				if strings.EqualFold(version, "original") {
					status = "waiting_original"
				}
				markDeferred(status, "file_missing", "未找到正文版本")
				continue
			}
			if minFileSize > 0 && len([]byte(content.Content)) < minFileSize {
				status := "waiting_generate"
				if strings.EqualFold(version, "original") {
					status = "waiting_original"
				}
				markDeferred(status, "file_too_small", fmt.Sprintf("正文不足 %d 字节", minFileSize))
				continue
			}
			if strings.HasPrefix(strings.ToLower(version), "ai") {
				if err := novelfetchworkshop.ValidateRewriteProse(content.Content); err != nil {
					// 与参考实现一致：AI 正文复检失败转入待生成，等待重新生成，而不是记为提交失败。
					markDeferred("waiting_generate", "ai_output_rejected", "AI 正文复检未通过："+err.Error()+"；请重新生成该版本后提交")
					continue
				}
			}
			// 与参考实现一致：正文（规范化换行后）与其他已确认版本完全相同则跳过，
			// 避免同一内容以不同版本号重复上传。
			if !request.Force {
				if candidates := confirmedHashes[webSubmitCanonicalContentHash(content.Content)]; len(candidates) > 0 {
					// 与参考实现一致：只与其他已确认版本比较，不把当前版本算作重复。
					matchedVersion := ""
					for _, candidate := range candidates {
						if candidate != version {
							matchedVersion = candidate
							break
						}
					}
					if matchedVersion != "" {
						row["status"] = "skipped"
						row["reason"] = "duplicate_content"
						message := fmt.Sprintf("内容与已确认版本 %s 相同，跳过重复上传", matchedVersion)
						row["error"] = message
						result.Results = append(result.Results, row)
						manifestItem(bookID, version, "skipped", "duplicate_content", message)
						continue
					}
				}
			}
			organization := webSubmitOrganization(web, profile)
			if organization == "" {
				markDeferred("waiting_config", "organization_missing", "网站上传配置缺少组织归属，请先同步或重新保存上传配置")
				continue
			}
			platform, _ := strconv.Atoi(platformText)
			gender, style := submissionGenderStyle(doc)
			fields, fieldErr := buildUploadFieldsCompat(platform, gender, style, webSubmitAdvanced(web, profile), map[string]any{"organization": organization})
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
			// 上传前先写入 executing，崩溃后可以识别“已派发但结果未知”，
			// 下一次提交只会回读确认，不会重复上传。
			meta["siteSubmitExecutingVersions"] = uniqueStrings(append(anyStrings(meta["siteSubmitExecutingVersions"]), version))
			meta["siteSubmitStatus"] = "executing"
			meta["siteSubmitFailedVersions"] = uniqueStrings(append(anyStrings(meta["siteSubmitFailedVersions"]), version))
			pendingErrors := mapOrEmpty(meta["siteSubmitFailedErrors"])
			pendingErrors[version] = "提交请求已开始，等待目标站确认"
			meta["siteSubmitFailedErrors"] = pendingErrors
			pendingReceipts := mapOrEmpty(meta["siteSubmitPendingConfirmation"])
			pendingReceipts[version] = uploadBookID
			meta["siteSubmitPendingConfirmation"] = pendingReceipts
			manifestItem(bookID, version, "executing", "submission_in_progress", "提交请求已开始，等待目标站确认")
			if err := store.PutDocument(ctx, owner, doc); err != nil {
				return result, fmt.Errorf("提交前保存尝试记录失败: %w", err)
			}
			var submission videomanagementconnector.Submission
			var submitErr error
			for attempt := 0; attempt <= retryTimes; attempt++ {
				submission, submitErr = targets[0].Submit(ctx, targetSession, videomanagementconnector.SubmissionInput{BookID: uploadBookID, Filename: uploadBookID + ".txt", Content: []byte(content.Content), FormFields: fields})
				if submitErr == nil {
					break
				}
				if !strings.Contains(strings.ToLower(submitErr.Error()), "rejected submission") {
					break
				}
			}
			if submitErr != nil {
				if errors.Is(submitErr, videomanagementconnector.ErrSiteContractChanged) {
					uncertainVersion("接口结构变化，提交结果待回读确认；不会重复上传", uploadBookID)
					return result, blockContract()
				}
				if errors.Is(submitErr, videomanagementconnector.ErrSessionExpired) || errors.Is(submitErr, videomanagementconnector.ErrSessionMissing) {
					delete(mapOrEmpty(meta["siteSubmitPendingConfirmation"]), version)
					clearExecuting()
					markSubmissionFailure(meta, version, workshopLoginRequiredMessage)
					meta["siteSubmitStatus"] = "waiting_login"
					doc.Meta = meta
					doc.Logs = append(doc.Logs, map[string]any{"event": "site_submit", "operation_id": operationID, "version": version, "status": "waiting_login", "error": workshopLoginRequiredMessage, "externalRequested": true, "time": time.Now().UTC().Format(time.RFC3339Nano)})
					manifestItem(bookID, version, "failed", "login_required", workshopLoginRequiredMessage)
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
			// 与批量原文改文系统源码对齐：上传接口返回成功即视为已提交；
			// 回读只用来补充证据，回读失败或只有旧记录都不再阻塞任务。
			confirmed := submission
			if readback, readbackErr := targets[0].GetSubmission(ctx, targetSession, bookID); readbackErr == nil {
				confirmed = readback
			} else if errors.Is(readbackErr, videomanagementconnector.ErrSiteContractChanged) {
				clearExecuting()
				manifestItem(bookID, version, "unknown", "site_contract_changed", "接口结构变化，提交结果待回读确认")
				return result, blockContract()
			}
			recordSuccess(submission, confirmed)
		}
		if deferredStatus != "" {
			meta["siteSubmitStatus"] = deferredStatus
		} else if len(mapOrEmpty(meta["siteSubmitPendingConfirmation"])) > 0 {
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
