package httpapi

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"qiantie/backend/internal/auth"
	"qiantie/backend/internal/novelfetchworkshop"
	"qiantie/backend/internal/safehttp"
	"qiantie/backend/internal/videomanagementconnector"
)

var novelFetchBookIDPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)
var novelFetchVersionIDPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]{1,64}$`)

// novelFetchMaxTxtLimit matches the legacy API validation. The new workshop
// endpoints must enforce the same ceiling, otherwise an oversized maxTxt sends
// an entire novel into the rewrite prompt and amplifies AI cost/timeouts.
const novelFetchMaxTxtLimit = 200000

func novelFetchMaxTxtInvalid(value int) bool {
	return value > novelFetchMaxTxtLimit
}

// The legacy workbench accepts source IDs with a trailing 5-25 digit run.
// Keep that range here so short but valid platform IDs are not discarded
// before the fetch stage gets a chance to handle them.
var novelFetchNumericIDPattern = regexp.MustCompile(`\d{5,25}`)

// 参考实现 safe_book_id 的提取顺序：先识别 u_<scope>_<digits> 复合书号，
// 再从链接参数 bookid/book_id/bookId 中取值，最后才回退到任意 5-25 位数字。
// 直接取第一段数字会把链接里的章节号等参数误当书号。
var novelFetchBookIDURLPattern = regexp.MustCompile(`(?i)(?:bookid|book_id|bookId)\s*=\s*(\d{5,25})`)
var novelFetchScopedBookIDPattern = regexp.MustCompile(`^u_[A-Za-z0-9_-]+_\d{5,25}$`)

func novelFetchSafeBookID(value string) string {
	text := strings.TrimSpace(value)
	if text == "" {
		return ""
	}
	if novelFetchScopedBookIDPattern.MatchString(text) {
		return text
	}
	if match := novelFetchBookIDURLPattern.FindStringSubmatch(text); len(match) > 1 {
		return match[1]
	}
	return novelFetchNumericIDPattern.FindString(text)
}

var novelFetchBearerSecretPattern = regexp.MustCompile(`(?i)(authorization\s*:\s*bearer\s+)[^\s,;]+`)
var novelFetchCookieSecretPattern = regexp.MustCompile(`(?i)(cookie\s*:\s*)[^;\r\n]+`)
var novelFetchNamedSecretPattern = regexp.MustCompile(`(?i)((?:api[_-]?key|access[_-]?token|refresh[_-]?token|password|secret)\s*[:=]\s*)[^,\s;]+`)
var errNovelFetchOriginalFetch = errors.New("原文获取失败")
var errNovelFetchClassificationPending = errors.New("性别或题材风格尚未确认，自动改文已暂停；确认分类后可重试")

var defaultNovelFetchPlatforms = []any{
	map[string]any{"id": "1", "name": "黑岩付费"}, map[string]any{"id": "2", "name": "番茄付费"},
	map[string]any{"id": "3", "name": "七猫付费"}, map[string]any{"id": "4", "name": "点众付费"},
	map[string]any{"id": "7", "name": "番茄免费"}, map[string]any{"id": "15", "name": "知乎付费"},
	map[string]any{"id": "20", "name": "掌阅付费"}, map[string]any{"id": "26", "name": "卓越付费"},
	map[string]any{"id": "29", "name": "九州书城"}, map[string]any{"id": "31", "name": "掌文付费"},
}

var defaultNovelFetchStyles = []any{"古风虐文", "古风甜文", "古风通用", "年代虐文", "年代甜文", "年代通用", "现代虐文", "现代甜文", "现代悬疑", "现代通用", "男频都市", "现代女主", "玄幻", "历史", "爆款BGM", "家庭奇葩", "家庭伤感", "职场打脸"}

type NovelFetchBridgeAuth struct {
	Secret          string
	Now             func() time.Time
	MaxSkew         time.Duration
	Users           BridgeUserResolver
	AuthStore       auth.Store
	SessionResolver func(context.Context, string) (BridgeIdentity, error)
	IdentityScope   func(context.Context, BridgeIdentity, string) (BridgeIdentity, error)
}

func novelFetchCanonicalPayload(username, issuedAt, owner, method, pathname string) string {
	return strings.Join([]string{username, issuedAt, owner, method, pathname}, "\n")
}

func novelFetchSignature(secret, payload string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

func (a NovelFetchBridgeAuth) authenticate(req *http.Request) (BridgeIdentity, error) {
	if token := bearerToken(req); token != "" && a.AuthStore != nil {
		now := time.Now
		if a.Now != nil {
			now = a.Now
		}
		session, err := a.AuthStore.FindSession(req.Context(), token, now().UTC())
		if err == nil {
			account, accountErr := a.AuthStore.FindAccount(req.Context(), session.Username)
			if accountErr == nil && account.Active {
				return BridgeIdentity{Username: account.Username, IsOwner: account.IsOwner}, nil
			}
		}
		if a.SessionResolver != nil {
			if identity, resolveErr := a.SessionResolver(req.Context(), token); resolveErr == nil {
				return identity, nil
			}
		}
		return BridgeIdentity{}, errors.New("invalid login session")
	}
	if token := bearerToken(req); token != "" && a.SessionResolver != nil {
		if identity, err := a.SessionResolver(req.Context(), token); err == nil {
			return identity, nil
		}
		return BridgeIdentity{}, errors.New("invalid login session")
	}
	if strings.TrimSpace(a.Secret) == "" {
		return BridgeIdentity{}, errors.New("bridge secret is not configured")
	}
	username := strings.TrimSpace(req.Header.Get(HeaderUsername))
	ownerRaw := req.Header.Get(HeaderIsOwner)
	issuedRaw := req.Header.Get(HeaderIssuedAt)
	signature := strings.TrimSpace(req.Header.Get(HeaderSignature))
	if username == "" || ownerRaw == "" || issuedRaw == "" || signature == "" {
		return BridgeIdentity{}, errors.New("missing signed bridge headers")
	}
	isOwner, err := strconv.ParseBool(ownerRaw)
	if err != nil {
		return BridgeIdentity{}, errors.New("invalid owner header")
	}
	issuedUnix, err := strconv.ParseInt(issuedRaw, 10, 64)
	if err != nil {
		return BridgeIdentity{}, errors.New("invalid issued-at header")
	}
	now := time.Now
	if a.Now != nil {
		now = a.Now
	}
	maxSkew := 5 * time.Minute
	if a.MaxSkew > 0 {
		maxSkew = a.MaxSkew
	}
	delta := now().Sub(time.Unix(issuedUnix, 0))
	if delta < 0 {
		delta = -delta
	}
	if delta > maxSkew {
		return BridgeIdentity{}, errors.New("expired bridge signature")
	}
	expected := novelFetchSignature(a.Secret, novelFetchCanonicalPayload(username, issuedRaw, ownerRaw, req.Method, req.URL.Path))
	provided, err := hex.DecodeString(signature)
	if err != nil {
		return BridgeIdentity{}, errors.New("malformed bridge signature")
	}
	expectedBytes, _ := hex.DecodeString(expected)
	if !hmac.Equal(provided, expectedBytes) {
		return BridgeIdentity{}, errors.New("invalid bridge signature")
	}
	if a.Users != nil {
		if err := a.Users.ResolveBridgeUser(req.Context(), username, isOwner); err != nil {
			return BridgeIdentity{}, err
		}
	}
	return BridgeIdentity{Username: username, IsOwner: isOwner}, nil
}

func (a NovelFetchBridgeAuth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, err := a.authenticate(req)
		if err != nil {
			if bearerToken(req) != "" {
				w.Header().Set("X-Qiantie-Auth-Failure", "session")
			}
			writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
			return
		}
		if a.IdentityScope != nil {
			identity, err = a.IdentityScope(req.Context(), identity, req.Header.Get(HeaderNovelFetchAccount))
			if err != nil {
				writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
				return
			}
		}
		next.ServeHTTP(w, req.WithContext(context.WithValue(req.Context(), bridgeIdentityKey{}, identity)))
	})
}

func novelFetchBodyPathValues(w http.ResponseWriter, req *http.Request) (string, string, bool) {
	bookID := strings.TrimSpace(req.PathValue("bookId"))
	if !novelFetchBookIDPattern.MatchString(bookID) {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "非法的书籍ID"})
		return "", "", false
	}
	versionID := strings.TrimSpace(req.PathValue("versionId"))
	if versionID != "" && !novelFetchVersionIDPattern.MatchString(versionID) {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "非法的版本ID"})
		return "", "", false
	}
	return bookID, versionID, true
}

type workshopProcessRegistry struct {
	mu   sync.RWMutex
	jobs map[string]map[string]map[string]any
}

func newWorkshopProcessRegistry() *workshopProcessRegistry {
	return &workshopProcessRegistry{jobs: make(map[string]map[string]map[string]any)}
}

func (r *workshopProcessRegistry) put(owner, id string, job map[string]any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.jobs[owner] == nil {
		r.jobs[owner] = make(map[string]map[string]any)
	}
	r.jobs[owner][id] = job
}

func (r *workshopProcessRegistry) get(owner, id string) (map[string]any, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	job, ok := r.jobs[owner][id]
	if !ok {
		return nil, false
	}
	copyJob := map[string]any{}
	for key, value := range job {
		copyJob[key] = value
	}
	return copyJob, true
}

func (r *workshopProcessRegistry) latest(owner string, limit int) []map[string]any {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]map[string]any, 0, len(r.jobs[owner]))
	for _, job := range r.jobs[owner] {
		copyJob := map[string]any{}
		for key, value := range job {
			copyJob[key] = value
		}
		items = append(items, copyJob)
	}
	sort.Slice(items, func(i, j int) bool {
		return fmt.Sprint(items[i]["updated_at"]) > fmt.Sprint(items[j]["updated_at"])
	})
	if limit < 1 {
		limit = 1
	}
	if limit > 20 {
		limit = 20
	}
	if len(items) > limit {
		items = items[:limit]
	}
	return items
}

type workshopProcessWorker struct {
	store                  novelfetchworkshop.Store
	queue                  novelfetchworkshop.ProcessQueue
	registry               *workshopProcessRegistry
	fetcher                *novelfetchworkshop.UpstreamFetcher
	resolveGenderJudge     func(context.Context, string) novelfetchworkshop.AIGenderJudge
	resolveRewriteProvider func(context.Context, string) (novelfetchworkshop.RewriteProvider, error)
	autoSubmit             func(context.Context, novelfetchworkshop.ProcessJob) error
}

func (worker *workshopProcessWorker) run(ctx context.Context) {
	for ctx.Err() == nil {
		job, err := worker.queue.Claim(ctx, 30*time.Second)
		if err != nil {
			continue
		}
		worker.process(ctx, job)
	}
}

func (worker *workshopProcessWorker) process(ctx context.Context, job novelfetchworkshop.ProcessJob) {
	job.Attempts++
	startedAt := time.Now().UTC().Format(time.RFC3339Nano)
	progress := []any{workshopProgressItem("info", "任务开始", "", "")}
	activeBookIDs := make([]string, 0, len(job.BookIDs))
	for _, bookID := range job.BookIDs {
		if !worker.bookCancelled(ctx, job.Owner, bookID) {
			activeBookIDs = append(activeBookIDs, bookID)
		}
	}
	if len(activeBookIDs) == 0 {
		_ = updateWorkshopRun(worker.store, job, "stopped", "任务已停止", startedAt, true)
		_ = updateNovelFetchScheduleRuntimeStatus(ctx, worker.store, job.Owner, job.ScheduleID, "stopped", "")
		worker.registry.put(job.Owner, job.ID, map[string]any{"id": job.ID, "status": "stopped", "queue_state": "stopped", "started_at": startedAt, "updated_at": startedAt, "attempts": job.Attempts, "steps": []any{}, "progress": []any{}})
		_ = worker.queue.Ack(ctx, job)
		return
	}
	activeJob := job
	activeJob.BookIDs = activeBookIDs
	_ = updateWorkshopRun(worker.store, job, "running", "", startedAt, false)
	_ = updateNovelFetchScheduleRuntimeStatus(ctx, worker.store, job.Owner, job.ScheduleID, "running", "")
	worker.registry.put(job.Owner, job.ID, map[string]any{"id": job.ID, "status": "running", "queue_state": "running", "started_at": startedAt, "updated_at": startedAt, "attempts": job.Attempts, "steps": progress, "progress": progress})
	retry := job.Attempts < 3
	deferredOriginalFetchRetry := false
	failedBookIDs := make([]string, 0)
	retryBookIDs := make([]string, 0)
	classificationPendingCount := 0
	failureMessages := make([]string, 0)
	stopped := false
	completed := 0
	uploadFailures := make([]string, 0)
	loginBlockedCount := 0
	blockedUploadState := "waiting_login"
	type bookProcessResult struct {
		bookID   string
		bookName string
		err      error
		stopped  bool
	}
	concurrency := novelFetchConfiguredConcurrency(ctx, worker.store, job.Owner, len(activeBookIDs))
	semaphore := make(chan struct{}, concurrency)
	results := make(chan bookProcessResult, len(activeBookIDs))
	var processing sync.WaitGroup
	for _, bookID := range activeBookIDs {
		bookName := ""
		if document, err := worker.store.GetDocument(ctx, job.Owner, bookID); err == nil {
			bookName = firstText(document.BookInfo.WorkTitle, document.Meta["bookName"], document.Meta["book_name"])
		}
		progress = append(progress, workshopProgressItem("processing", "开始处理", bookID, bookName))
		bookID, bookName := bookID, bookName
		_ = updateWorkshopTaskRuntimeState(ctx, worker.store, job.Owner, bookID, "processing", "")
		processing.Add(1)
		go func() {
			defer processing.Done()
			select {
			case semaphore <- struct{}{}:
			case <-ctx.Done():
				results <- bookProcessResult{bookID: bookID, bookName: bookName, err: ctx.Err()}
				return
			}
			defer func() { <-semaphore }()
			bookJob := activeJob
			bookJob.BookIDs = []string{bookID}
			err := worker.execute(ctx, bookJob)
			results <- bookProcessResult{bookID: bookID, bookName: bookName, err: err, stopped: worker.bookCancelled(ctx, job.Owner, bookID)}
		}()
	}
	worker.registry.put(job.Owner, job.ID, map[string]any{"id": job.ID, "status": "running", "queue_state": "running", "started_at": startedAt, "updated_at": time.Now().UTC().Format(time.RFC3339Nano), "attempts": job.Attempts, "steps": progress, "progress": progress})
	go func() {
		processing.Wait()
		close(results)
	}()
	for result := range results {
		bookID, bookName := result.bookID, result.bookName
		if result.err != nil {
			failedBookIDs = append(failedBookIDs, bookID)
			failureMessages = append(failureMessages, bookID+"："+result.err.Error())
			status := "failed"
			if errors.Is(result.err, novelfetchworkshop.ErrRewriteValidation) {
				_ = updateWorkshopTaskRuntimeState(ctx, worker.store, job.Owner, bookID, "failed", result.err.Error())
			} else if errors.Is(result.err, errNovelFetchClassificationPending) {
				classificationPendingCount++
				status = "waiting_classification"
				// 记录分类阶段的等待状态，确认分类后由重试队列统一清理。
				_ = scheduleNovelFetchDocumentRetry(ctx, worker.store, job.Owner, bookID, novelFetchRetryStageClassify, result.err, time.Now().UTC())
			} else if errors.Is(result.err, errNovelFetchOriginalFetch) {
				retryBookIDs = append(retryBookIDs, bookID)
				deferredOriginalFetchRetry = true
				status = "waiting_retry"
				if retryErr := markNovelFetchOriginalFetchRetry(ctx, worker.store, job.Owner, bookID, result.err, time.Now().UTC()); retryErr != nil {
					_ = updateWorkshopTaskRuntimeState(ctx, worker.store, job.Owner, bookID, status, result.err.Error())
				}
			} else if novelFetchAIManualRetryRequired(result.err) {
				// 对照参考实现：模型确定性失败写 manual_ai_retry_required，
				// 不再自动重试。同一段超长文本反复提交只会重复截断。
				if markErr := markNovelFetchAIManualRetry(ctx, worker.store, job.Owner, bookID, result.err, time.Now().UTC()); markErr != nil {
					_ = updateWorkshopTaskRuntimeState(ctx, worker.store, job.Owner, bookID, status, result.err.Error())
				}
			} else if retry {
				retryBookIDs = append(retryBookIDs, bookID)
				status = "waiting_retry"
				_ = scheduleNovelFetchDocumentRetry(ctx, worker.store, job.Owner, bookID, novelFetchRetryStageForError(result.err), result.err, time.Now().UTC())
				_ = updateWorkshopTaskRuntimeState(ctx, worker.store, job.Owner, bookID, status, result.err.Error())
			} else {
				// 队列自身重试次数用尽后仍然写阶段退避时间，交给自动扫描器在
				// 到期后重新接管；否则状态 failed 会被立刻重新入队，形成死循环。
				_ = scheduleNovelFetchDocumentRetry(ctx, worker.store, job.Owner, bookID, novelFetchRetryStageForError(result.err), result.err, time.Now().UTC())
				_ = updateWorkshopTaskRuntimeState(ctx, worker.store, job.Owner, bookID, status, result.err.Error())
			}
			progress = append(progress, workshopProgressItem("error", "处理失败："+result.err.Error(), bookID, bookName))
			worker.registry.put(job.Owner, job.ID, map[string]any{"id": job.ID, "status": status, "queue_state": "running", "started_at": startedAt, "updated_at": time.Now().UTC().Format(time.RFC3339Nano), "attempts": job.Attempts, "steps": progress, "progress": progress})
			continue
		}
		if result.stopped {
			stopped = true
			_ = updateWorkshopTaskRuntimeState(ctx, worker.store, job.Owner, bookID, "stopped", "")
			continue
		}
		completed++
		_ = clearNovelFetchOriginalFetchRetry(ctx, worker.store, job.Owner, bookID)
		_ = updateWorkshopTaskRuntimeState(ctx, worker.store, job.Owner, bookID, "completed", "")
		progress = append(progress, workshopProgressItem("success", "处理完成", bookID, bookName))
		worker.registry.put(job.Owner, job.ID, map[string]any{"id": job.ID, "status": "running", "queue_state": "running", "started_at": startedAt, "updated_at": time.Now().UTC().Format(time.RFC3339Nano), "attempts": job.Attempts, "steps": progress, "progress": progress})
		if job.AutoSubmit && worker.autoSubmit != nil && !worker.bookCancelled(ctx, job.Owner, bookID) {
			bookJob := job
			bookJob.BookIDs = []string{bookID}
			if document, err := worker.store.GetDocument(ctx, job.Owner, bookID); err == nil {
				document.Meta["autoSubmitReady"] = true
				document.Meta["autoSubmitReadyVersions"] = append([]string{}, job.SelectedVersions...)
				document.Meta["autoSubmitReadyFingerprint"] = novelFetchUploadReadyFingerprint(ctx, worker.store, job.Owner, document, job.SelectedVersions)
				if err := worker.store.PutDocument(ctx, job.Owner, document); err != nil {
					uploadFailures = append(uploadFailures, bookID+"：提交前保存状态失败："+err.Error())
					failedBookIDs = append(failedBookIDs, bookID)
					failureMessages = append(failureMessages, bookID+"：提交前保存状态失败："+err.Error())
					if retry {
						retryBookIDs = append(retryBookIDs, bookID)
					}
					continue
				}
			}
			_ = appendWorkshopTaskLog(ctx, worker.store, job.Owner, bookID, "auto_submit_started", "本书处理完成，开始提交到视频管理系统")
			progress = append(progress, workshopProgressItem("processing", "本书开始上传", bookID, bookName))
			worker.registry.put(job.Owner, job.ID, map[string]any{"id": job.ID, "status": "running", "queue_state": "running", "started_at": startedAt, "updated_at": time.Now().UTC().Format(time.RFC3339Nano), "attempts": job.Attempts, "steps": progress, "progress": progress})
			if err := worker.autoSubmit(ctx, bookJob); err != nil {
				uploadFailures = append(uploadFailures, bookID+"："+err.Error())
				failedBookIDs = append(failedBookIDs, bookID)
				failureMessages = append(failureMessages, bookID+"：上传未完成："+err.Error())
				if blockedStatus := workshopSubmitBlockedStatus(err); blockedStatus != "" {
					loginBlockedCount++
					blockedUploadState = blockedStatus
					_ = updateWorkshopTaskRuntimeState(ctx, worker.store, job.Owner, bookID, blockedStatus, err.Error())
				} else if retry {
					retryBookIDs = append(retryBookIDs, bookID)
				}
				if workshopSubmitBlockedStatus(err) == "" {
					_ = updateWorkshopTaskRuntimeState(ctx, worker.store, job.Owner, bookID, "upload_pending", err.Error())
				}
				_ = appendWorkshopTaskLog(ctx, worker.store, job.Owner, bookID, "auto_submit_pending", "上传未完成，重试仅继续上传："+err.Error())
				progress = append(progress, workshopProgressItem("warning", "上传未完成："+err.Error(), bookID, bookName))
			} else {
				_ = appendWorkshopTaskLog(ctx, worker.store, job.Owner, bookID, "auto_submit_completed", "视频管理系统已确认提交完成")
				progress = append(progress, workshopProgressItem("success", "本书上传完成", bookID, bookName))
			}
		}
	}

	finishedAt := time.Now().UTC().Format(time.RFC3339Nano)
	if len(failedBookIDs) > 0 {
		errorMessage := strings.Join(failureMessages, "；")
		job.LastError = errorMessage
		status := "failed"
		queueState := "failed"
		if len(retryBookIDs) == 0 && classificationPendingCount > 0 {
			status = "waiting_classification"
			queueState = "waiting_classification"
		}
		if deferredOriginalFetchRetry {
			status = "waiting_retry"
			queueState = "waiting_retry"
			_ = updateWorkshopRun(worker.store, job, status, errorMessage, startedAt, true)
			_ = updateNovelFetchScheduleRuntimeStatus(ctx, worker.store, job.Owner, job.ScheduleID, status, errorMessage)
			documents, _ := worker.store.ListDocuments(ctx, job.Owner)
			result := map[string]any{"batch_id": job.BatchID, "current_batch_id": job.BatchID, "tasks": documents, "completed": completed, "failed": len(failedBookIDs), "failed_book_ids": failedBookIDs}
			progress = append(progress, workshopProgressItem("warning", "原文未抓到，等待下次自动抓取", "", ""))
			worker.registry.put(job.Owner, job.ID, map[string]any{"id": job.ID, "status": status, "queue_state": queueState, "started_at": startedAt, "updated_at": finishedAt, "attempts": job.Attempts, "error": errorMessage, "result": result, "steps": progress, "progress": progress})
			_ = worker.queue.Fail(ctx, job, false)
			return
		}
		if retry && len(retryBookIDs) > 0 {
			status = "waiting_retry"
			queueState = "queued"
		}
		if len(uploadFailures) == len(failedBookIDs) {
			status = "upload_pending"
		}
		if loginBlockedCount == len(failedBookIDs) {
			status = blockedUploadState
			queueState = blockedUploadState
		}
		retryEnabled := retry && len(retryBookIDs) > 0
		_ = updateWorkshopRun(worker.store, job, status, errorMessage, startedAt, !retryEnabled)
		if !retryEnabled {
			_ = updateNovelFetchScheduleRuntimeStatus(ctx, worker.store, job.Owner, job.ScheduleID, status, errorMessage)
		}
		documents, _ := worker.store.ListDocuments(ctx, job.Owner)
		result := map[string]any{"batch_id": job.BatchID, "current_batch_id": job.BatchID, "tasks": documents, "completed": completed, "failed": len(failedBookIDs), "failed_book_ids": failedBookIDs}
		progressMessage := "本轮结束，失败任务等待重试"
		if status == "waiting_classification" {
			progressMessage = "改文已暂停，等待确认分类后重试"
		} else if status == "waiting_login" {
			progressMessage = workshopLoginRequiredMessage
		} else if status == "waiting_configuration" {
			progressMessage = workshopContractChangedMessage
		} else if status == "upload_pending" {
			progressMessage = "正文已完成，上传未完成；重试仅继续上传"
		} else if !retryEnabled {
			progressMessage = "处理结束，任务失败"
		}
		progress = append(progress, workshopProgressItem("warning", progressMessage, "", ""))
		worker.registry.put(job.Owner, job.ID, map[string]any{"id": job.ID, "status": status, "queue_state": queueState, "started_at": startedAt, "updated_at": finishedAt, "attempts": job.Attempts, "error": errorMessage, "result": result, "steps": progress, "progress": progress})
		retryJob := job
		retryJob.BookIDs = retryBookIDs
		_ = worker.queue.Fail(ctx, retryJob, retryEnabled)
		return
	}

	documents, _ := worker.store.ListDocuments(ctx, job.Owner)
	result := map[string]any{"batch_id": job.BatchID, "current_batch_id": job.BatchID, "tasks": documents}
	finalStatus := "done"
	if stopped {
		finalStatus = "stopped"
	}
	scheduleStatus := "completed"
	scheduleMessage := ""
	if stopped {
		scheduleStatus = "stopped"
	} else if len(uploadFailures) > 0 {
		scheduleStatus = "upload_pending"
		scheduleMessage = strings.Join(uploadFailures, "；")
	}
	if scheduleStatus == "upload_pending" {
		finalStatus = "upload_pending"
		_ = updateWorkshopRun(worker.store, job, finalStatus, scheduleMessage, startedAt, false)
	} else {
		_ = updateWorkshopRun(worker.store, job, finalStatus, "", startedAt, true)
	}
	_ = updateNovelFetchScheduleRuntimeStatus(ctx, worker.store, job.Owner, job.ScheduleID, scheduleStatus, scheduleMessage)
	if finalStatus == "done" {
		progress = append(progress, workshopProgressItem("success", "批次处理完成", "", ""))
	} else if finalStatus == "upload_pending" {
		progress = append(progress, workshopProgressItem("warning", "内容处理完成，等待自动上传重试", "", ""))
	}
	worker.registry.put(job.Owner, job.ID, map[string]any{"id": job.ID, "status": finalStatus, "queue_state": finalStatus, "started_at": startedAt, "updated_at": finishedAt, "attempts": job.Attempts, "result": result, "steps": progress, "progress": progress})
	_ = worker.queue.Ack(ctx, job)
}

func workshopProgressItem(level, message, bookID, bookName string) map[string]any {
	return map[string]any{
		"time": time.Now().UTC().Format(time.RFC3339Nano), "level": level,
		"message": message, "book_id": strings.TrimSpace(bookID), "book_name": strings.TrimSpace(bookName),
	}
}

func novelFetchConfiguredConcurrency(ctx context.Context, store novelfetchworkshop.Store, owner string, taskCount int) int {
	concurrency := 4
	if settings, err := store.GetConfig(ctx, owner); err == nil {
		if configured := toInt(mapOrEmpty(settings["fetch"])["concurrency"]); configured > 0 {
			concurrency = configured
		}
	}
	if concurrency > 20 {
		concurrency = 20
	}
	if taskCount > 0 && concurrency > taskCount {
		concurrency = taskCount
	}
	if concurrency < 1 {
		return 1
	}
	return concurrency
}

func (worker *workshopProcessWorker) bookCancelled(ctx context.Context, owner, bookID string) bool {
	if cancellationQueue, ok := worker.queue.(novelfetchworkshop.ProcessCancellationQueue); ok {
		cancelled, err := cancellationQueue.IsBookCancelled(ctx, owner, bookID)
		if err == nil && cancelled {
			return true
		}
	}
	document, err := worker.store.GetDocument(ctx, owner, bookID)
	if err != nil {
		return false
	}
	return novelFetchStopped(firstText(document.Meta["status"], document.Meta["queueState"]))
}

func (worker *workshopProcessWorker) execute(ctx context.Context, job novelfetchworkshop.ProcessJob) error {
	ctx = novelfetchworkshop.WithProcessTextModel(ctx, job.TextModelID)
	if job.AutoSubmit && len(job.BookIDs) == 1 {
		if document, err := worker.store.GetDocument(ctx, job.Owner, job.BookIDs[0]); err == nil && document.Meta["autoSubmitReady"] == true && firstText(document.Meta["autoSubmitReadyFingerprint"]) == novelFetchUploadReadyFingerprint(ctx, worker.store, job.Owner, document, job.SelectedVersions) {
			_ = appendWorkshopTaskLog(ctx, worker.store, job.Owner, job.BookIDs[0], "upload_resume", "正文已完成，跳过原文和改文生成，仅继续上传")
			return nil
		}
	}
	var sensitiveProvider novelfetchworkshop.RewriteProvider
	if job.SensitiveAIEnabled && worker.resolveRewriteProvider != nil {
		provider, err := worker.resolveRewriteProvider(ctx, job.Owner)
		if err != nil {
			return err
		}
		sensitiveProvider = provider
	}
	selectedAI := selectedNovelFetchAIVersions(job.SelectedVersions)
	for _, bookID := range job.BookIDs {
		if worker.bookCancelled(ctx, job.Owner, bookID) {
			continue
		}
		document, documentErr := worker.store.GetDocument(ctx, job.Owner, bookID)
		originalReady := documentErr == nil && strings.EqualFold(firstText(document.Meta["originalStatus"], document.Meta["original_status"]), "done") && strings.TrimSpace(firstText(document.ProcessedText, document.Original)) != ""
		if !originalReady {
			if worker.fetcher == nil && strings.TrimSpace(document.OriginalRaw) == "" {
				// Preserve the legacy no-fetch worker behaviour used by recovery
				// jobs; a ready original still proceeds to classification below.
				continue
			}
			var judge novelfetchworkshop.AIGenderJudge
			if worker.resolveGenderJudge != nil {
				judge = worker.resolveGenderJudge(ctx, job.Owner)
			}
			if err := executeWorkshopDocumentFetch(ctx, worker.store, job.Owner, bookID, worker.fetcher, job.MaxTxt, job.SensitiveAIEnabled, sensitiveProvider, judge, sensitiveProvider); err != nil {
				return err
			}
		}
		document, documentErr = worker.store.GetDocument(ctx, job.Owner, bookID)
		if documentErr == nil && strings.EqualFold(firstText(document.Meta["originalStatus"], document.Meta["original_status"]), "too_short") {
			continue
		}
		var judge novelfetchworkshop.AIGenderJudge
		if worker.resolveGenderJudge != nil {
			judge = worker.resolveGenderJudge(ctx, job.Owner)
		}
		var classificationProvider novelfetchworkshop.RewriteProvider
		if worker.resolveRewriteProvider != nil {
			classificationProvider, _ = worker.resolveRewriteProvider(ctx, job.Owner)
		}
		if document, getErr := worker.store.GetDocument(ctx, job.Owner, bookID); getErr == nil && novelFetchClassificationResolved(document) {
			if originalReady {
				_ = appendWorkshopTaskLog(ctx, worker.store, job.Owner, bookID, "classification_skipped", "已有分类信息，跳过判断")
			}
		} else {
			_ = appendWorkshopTaskLog(ctx, worker.store, job.Owner, bookID, "classification_started", "开始判断男女频和题材风格")
			if err := ensureWorkshopDocumentClassification(ctx, worker.store, job.Owner, bookID, judge, classificationProvider); err != nil {
				_ = appendWorkshopTaskLog(ctx, worker.store, job.Owner, bookID, "classification_failed", err.Error())
				return err
			}
			_ = appendWorkshopTaskLog(ctx, worker.store, job.Owner, bookID, "classification_completed", "男女频和题材风格判断完成")
		}
		if worker.bookCancelled(ctx, job.Owner, bookID) {
			continue
		}
		if len(selectedAI) == 0 {
			continue
		}
		document, documentErr = worker.store.GetDocument(ctx, job.Owner, bookID)
		if documentErr != nil {
			return documentErr
		}
		pendingAI := make([]string, 0, len(selectedAI))
		for _, versionID := range selectedAI {
			if strings.TrimSpace(toString(document.Versions[versionID])) != "" {
				continue
			}
			if body, bodyErr := worker.store.GetBody(ctx, job.Owner, bookID, versionID); bodyErr == nil && strings.TrimSpace(body.Content) != "" {
				continue
			}
			pendingAI = append(pendingAI, versionID)
		}
		if len(pendingAI) == 0 {
			continue
		}
		if !novelFetchClassificationResolved(document) {
			document.Meta["status"] = "waiting_classification"
			document.Meta["queueState"] = "waiting_classification"
			document.Meta["error"] = "性别或题材风格尚未确认，自动改文已暂停；确认分类后可重试"
			if err := worker.store.PutDocument(ctx, job.Owner, document); err != nil {
				return err
			}
			return errNovelFetchClassificationPending
		}
		if worker.resolveRewriteProvider == nil {
			return errors.New("未配置文本模型")
		}
		provider, err := worker.resolveRewriteProvider(ctx, job.Owner)
		if err != nil {
			return err
		}
		_ = appendWorkshopTaskLog(ctx, worker.store, job.Owner, bookID, "rewrite_started", "开始生成 AI 改写版本")
		if _, err := novelfetchworkshop.GenerateAIRewritesForVersionsProcessed(ctx, worker.store, job.Owner, bookID, pendingAI, provider, workshopRewriteProcessor(worker.store, job.Owner, bookID, job.SensitiveAIEnabled, provider), job.RewriteMethods); err != nil {
			_ = appendWorkshopTaskLog(ctx, worker.store, job.Owner, bookID, "rewrite_failed", err.Error())
			return err
		}
		_ = appendWorkshopTaskLog(ctx, worker.store, job.Owner, bookID, "rewrite_completed", "AI 改写版本已保存")
	}
	return nil
}

func novelFetchUploadReadyFingerprint(ctx context.Context, store novelfetchworkshop.Store, owner string, document novelfetchworkshop.Document, versions []string) string {
	parts := []string{document.Original, document.ProcessedText, strings.Join(versions, "\x00")}
	for _, version := range versions {
		if body, err := store.GetBody(ctx, owner, document.BookID, version); err == nil {
			parts = append(parts, version, body.Content)
		} else {
			parts = append(parts, version, firstText(document.Versions[version]))
		}
	}
	hash := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return hex.EncodeToString(hash[:])
}

// ensureWorkshopDocumentClassification repairs ready originals created by
// older runs. It intentionally does not fetch or rewrite the body again.
func ensureWorkshopDocumentClassification(ctx context.Context, store novelfetchworkshop.Store, owner, bookID string, genderJudge novelfetchworkshop.AIGenderJudge, styleProvider novelfetchworkshop.RewriteProvider) error {
	document, err := store.GetDocument(ctx, owner, bookID)
	if err != nil {
		return err
	}
	if document.Meta == nil {
		document.Meta = map[string]any{}
	}
	text := strings.TrimSpace(firstText(document.ProcessedText, document.Original))
	if text == "" {
		return nil
	}
	needsGender := genderNeedsResolution(document.Gender)
	needsStyle := styleNeedsResolution(document.Meta)
	var classification novelfetchworkshop.ClassificationResult
	if needsGender || needsStyle {
		info := document.BookInfo
		info.Text = text
		classification = novelfetchworkshop.ClassifyBook(ctx, "", info, genderJudge, styleProvider)
	}
	if needsGender {
		decision := classification.Gender
		document.Gender, document.GenderSource, document.GenderReason = decision.Gender, decision.Source, decision.Reason
		document.GenderConfidence = decision.Confidence
	}
	if needsStyle {
		document.Meta["styleStatus"] = "processing"
		recordNovelFetchStyleDecision(document.Meta, classification.Style, nil)
	}
	syncNovelFetchClassificationMetadata(&document)
	return store.PutDocument(ctx, owner, document)
}

func updateWorkshopTaskRuntimeState(ctx context.Context, store novelfetchworkshop.Store, owner, bookID, status, message string) error {
	document, err := store.GetDocument(ctx, owner, bookID)
	if err != nil {
		return err
	}
	if document.Meta == nil {
		document.Meta = map[string]any{}
	}
	document.Meta["status"] = status
	document.Meta["queueState"] = status
	if strings.TrimSpace(message) == "" {
		delete(document.Meta, "error")
	} else {
		document.Meta["error"] = message
	}
	detail := "处理状态：" + status
	if strings.TrimSpace(message) != "" {
		detail += "；" + strings.TrimSpace(message)
	}
	appendNovelFetchDocumentLog(&document, "process_"+status, detail, time.Now().UTC())
	return store.PutDocument(ctx, owner, document)
}

func setNovelFetchBookQueueState(ctx context.Context, store novelfetchworkshop.Store, owner, bookID, jobID, batchID, status string) error {
	document, err := store.GetDocument(ctx, owner, bookID)
	if err != nil {
		return err
	}
	if document.Meta == nil {
		document.Meta = map[string]any{}
	}
	document.Meta["status"] = status
	document.Meta["queueState"] = status
	document.Meta["currentProcessJobId"] = jobID
	if strings.TrimSpace(batchID) != "" {
		document.Meta["currentBatchId"] = batchID
	}
	delete(document.Meta, "error")
	appendNovelFetchDocumentLog(&document, "queue_"+status, "任务已进入小说获取队列", time.Now().UTC())
	return store.PutDocument(ctx, owner, document)
}

func markNovelFetchTaskQueued(ctx context.Context, store novelfetchworkshop.Store, owner, bookID, jobID, batchID string, at time.Time) error {
	document, err := store.GetDocument(ctx, owner, bookID)
	if err != nil {
		return err
	}
	if document.Meta == nil {
		document.Meta = map[string]any{}
	}
	status := strings.ToLower(strings.TrimSpace(firstText(document.Meta["status"], document.Meta["queueState"])))
	if status == "running" || status == "processing" || novelFetchStopped(status) {
		return nil
	}
	currentJobID := firstText(document.Meta["currentProcessJobId"], document.Meta["current_process_job_id"])
	if currentJobID != "" && currentJobID != jobID && status == "queued" {
		return nil
	}
	document.Meta["status"] = "queued"
	document.Meta["queueState"] = "queued"
	delete(document.Meta, "scheduledHold")
	document.Meta["currentProcessJobId"] = jobID
	if strings.TrimSpace(batchID) != "" {
		document.Meta["currentBatchId"] = batchID
	}
	delete(document.Meta, "error")
	// 入队成功说明 Redis 已恢复，清掉入队失败留下的退避字段；否则下一次
	// 扫描会被旧的 automationNextRetryAt 继续挡住。
	delete(document.Meta, "automationEnqueueAttempts")
	delete(document.Meta, "automationNextRetryAt")
	appendNovelFetchDocumentLog(&document, "queue_queued", "任务已进入小说获取队列", at)
	return store.PutDocument(ctx, owner, document)
}

func activeNovelFetchJobForBooks(ctx context.Context, store novelfetchworkshop.Store, owner string, bookIDs []string) map[string]any {
	return activeNovelFetchJobForBooksWithQueue(ctx, store, nil, owner, bookIDs)
}

// activeNovelFetchJobForBooksWithQueue reports the first book that must not be
// deleted. It derives the effective status from the same rule as the task list
// (queueState wins over a stale terminal meta.status) and, when a queue is
// available, also checks whether the referenced job still exists in Redis.
func activeNovelFetchJobForBooksWithQueue(ctx context.Context, store novelfetchworkshop.Store, queue novelfetchworkshop.ProcessQueue, owner string, bookIDs []string) map[string]any {
	for _, bookID := range bookIDs {
		document, err := store.GetDocument(ctx, owner, bookID)
		if err != nil {
			continue
		}
		jobID := firstText(document.Meta["currentProcessJobId"], document.Meta["current_process_job_id"])
		status := novelfetchworkshop.DeriveTaskStatus(document)
		active := status == novelfetchworkshop.TaskStatusQueued || status == novelfetchworkshop.TaskStatusProcessing
		if !active && jobID != "" && queue != nil {
			if exists, existsErr := novelFetchProcessJobExists(ctx, queue, jobID); existsErr == nil && exists {
				active = true
			}
		}
		if !active {
			continue
		}
		return map[string]any{
			"id":          jobID,
			"bookId":      bookID,
			"status":      status,
			"queue_state": status,
		}
	}
	return nil
}

func novelFetchProcessJobExists(ctx context.Context, queue novelfetchworkshop.ProcessQueue, jobID string) (bool, error) {
	type processJobExistence interface {
		ProcessJobExists(context.Context, string) (bool, error)
	}
	if checker, ok := queue.(processJobExistence); ok {
		return checker.ProcessJobExists(ctx, jobID)
	}
	return false, nil
}

// appendWorkshopTaskLog persists a concise per-book operation log so that the
// task detail remains useful after the in-memory realtime progress expires.
func appendWorkshopTaskLog(ctx context.Context, store novelfetchworkshop.Store, owner, bookID, event, detail string) error {
	document, err := store.GetDocument(ctx, owner, bookID)
	if err != nil {
		return err
	}
	appendNovelFetchDocumentLog(&document, event, detail, time.Now().UTC())
	return store.PutDocument(ctx, owner, document)
}

func appendNovelFetchDocumentLog(document *novelfetchworkshop.Document, event, detail string, at time.Time) {
	event = strings.TrimSpace(event)
	if event == "" {
		event = "operation"
	}
	detail = strings.TrimSpace(detail)
	detail = novelFetchBearerSecretPattern.ReplaceAllString(detail, `${1}[已隐藏]`)
	detail = novelFetchCookieSecretPattern.ReplaceAllString(detail, `${1}[已隐藏]`)
	detail = novelFetchNamedSecretPattern.ReplaceAllString(detail, `${1}[已隐藏]`)
	if len([]rune(detail)) > 500 {
		detail = string([]rune(detail)[:500]) + "…"
	}
	if document.Logs == nil {
		document.Logs = []any{}
	}
	entry := map[string]any{
		"time":   at.UTC().Format(time.RFC3339Nano),
		"event":  event,
		"detail": detail,
	}
	if stage, status := novelFetchLogStageStatus(event); stage != "" {
		entry["stage"] = stage
		entry["status"] = status
	}
	document.Logs = append(document.Logs, entry)
	if len(document.Logs) > 200 {
		document.Logs = document.Logs[len(document.Logs)-200:]
	}
}

func novelFetchLogStageStatus(event string) (string, string) {
	for _, stage := range []string{"original_raw", "sensitive_processing", "rules_processing", "classification", "rewrite", "fetch", "auto_submit", "schedule_dispatch", "queue", "process"} {
		prefix := stage + "_"
		if strings.HasPrefix(event, prefix) {
			return stage, strings.TrimPrefix(event, prefix)
		}
	}
	return "", ""
}

func updateWorkshopRun(store novelfetchworkshop.Store, job novelfetchworkshop.ProcessJob, status, message, startedAt string, finished bool) error {
	record := novelfetchworkshop.RunRecord{RunID: job.ID, BookID: job.BatchID, Stage: "novel_fetch_process", Status: status, Attempts: job.Attempts, TextModelID: job.TextModelID, ModelID: "novel-fetch-workshop", ModelDisplayName: "小说获取完整流程", ErrorMessage: message, StartedAt: startedAt}
	if finished {
		record.FinishedAt = time.Now().UTC().Format(time.RFC3339Nano)
	}
	return store.PutRun(context.Background(), job.Owner, record)
}

func updateNovelFetchScheduleRuntimeStatus(ctx context.Context, store novelfetchworkshop.Store, owner, scheduleID, status, message string) error {
	if strings.TrimSpace(scheduleID) == "" {
		return nil
	}
	settings, err := store.GetConfig(ctx, owner)
	if err != nil {
		return err
	}
	schedules, _ := settings["novel_fetch_schedules"].([]any)
	found := false
	for _, raw := range schedules {
		item, ok := raw.(map[string]any)
		if !ok || firstText(item["id"]) != scheduleID {
			continue
		}
		if firstText(item["status"]) == "cancelled" {
			return nil
		}
		if item["partial_dispatch"] == true {
			// Child results do not settle a parent that still owns other books.
			return nil
		}
		item["status"] = status
		item["updatedAt"] = time.Now().UTC().Format(time.RFC3339Nano)
		if strings.TrimSpace(message) == "" {
			delete(item, "error")
		} else {
			item["error"] = message
		}
		found = true
		break
	}
	if !found {
		return nil
	}
	settings["novel_fetch_schedules"] = schedules
	return store.PutConfig(ctx, owner, settings)
}

const (
	novelFetchScheduleDispatchInterval = 5 * time.Second
	novelFetchAutomationScanInterval   = 20 * time.Second
	novelFetchScheduleRecoveryTimeout  = 30 * time.Second
)

func runNovelFetchScheduleDispatcher(ctx context.Context, store novelfetchworkshop.Store, queue novelfetchworkshop.ProcessQueue, owners novelfetchworkshop.ConfigOwnerLister, registry *workshopProcessRegistry, recoverUpload ...func(context.Context, string)) {
	dispatchSchedules := func() {
		ownerNames, err := owners.ListConfigOwners(ctx)
		if err != nil {
			return
		}
		now := time.Now()
		for _, owner := range ownerNames {
			_ = dispatchDueNovelFetchSchedules(ctx, store, queue, owner, now)
		}
	}
	dispatchAutomation := func() {
		ownerNames, err := owners.ListConfigOwners(ctx)
		if err != nil {
			return
		}
		now := time.Now()
		for _, owner := range ownerNames {
			if len(recoverUpload) > 0 && recoverUpload[0] != nil {
				recoverUpload[0](ctx, owner)
			}
			job, dispatchErr := dispatchNovelFetchAutomationOwner(ctx, store, queue, owner, now)
			if dispatchErr != nil || job == nil {
				continue
			}
			updatedAt := time.Now().UTC().Format(time.RFC3339Nano)
			registry.put(owner, job.ID, map[string]any{"id": job.ID, "status": "queued", "queue_state": "queued", "started_at": updatedAt, "updated_at": updatedAt, "steps": []any{}, "progress": []any{}})
		}
	}

	// Recover already-due work at startup, then keep each source on its own
	// cadence. UI polling remains independent and can stay responsive.
	dispatchSchedules()
	dispatchAutomation()
	scheduleTicker := time.NewTicker(novelFetchScheduleDispatchInterval)
	automationTicker := time.NewTicker(novelFetchAutomationScanInterval)
	defer scheduleTicker.Stop()
	defer automationTicker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-scheduleTicker.C:
			dispatchSchedules()
		case <-automationTicker.C:
			dispatchAutomation()
		}
	}
}

func dispatchDueNovelFetchSchedules(ctx context.Context, store novelfetchworkshop.Store, queue novelfetchworkshop.ProcessQueue, owner string, now time.Time) error {
	settings, err := store.GetConfig(ctx, owner)
	if err != nil {
		return err
	}
	schedules, _ := settings["novel_fetch_schedules"].([]any)
	for _, raw := range schedules {
		item, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		status := strings.ToLower(strings.TrimSpace(firstText(item["status"])))
		if len(scheduleStringSlice(item["pending_book_ids"])) > 0 && !novelFetchStopped(status) {
			if enabled, explicit := mapOrEmpty(settings[novelFetchAutomationConfigKey])["enabled"].(bool); !explicit || enabled {
				if err := dispatchNovelFetchSchedulePending(ctx, store, queue, owner, item, now); err != nil {
					return err
				}
			}
			continue
		}
		if handled, err := reconcileNovelFetchScheduleRetry(ctx, store, owner, item, settings, now); err != nil {
			return err
		} else if handled {
			continue
		}
		recovering := status == "dispatching"
		if status != "" && status != "scheduled" && status != "preparing" && status != "paused" && !recovering {
			continue
		}
		runAt, err := parseNovelFetchScheduleTime(firstText(item["run_at"], item["runAt"]))
		if err != nil {
			item["status"] = "failed"
			item["error"] = "定时任务时间无效，请重新设置时间后再执行"
			if recovering {
				item["error"] = "定时派发时间信息不完整，已停止自动恢复"
			}
			item["updatedAt"] = now.UTC().Format(time.RFC3339Nano)
			if err := saveNovelFetchScheduleItem(ctx, store, owner, item); err != nil {
				return err
			}
			continue
		}
		if !recovering && item["tasks_prepared"] != true {
			if err := prepareNovelFetchScheduleTasks(ctx, store, owner, item, now); err != nil {
				return err
			}
			status = firstText(item["status"])
			if status != "scheduled" {
				continue
			}
		}
		if recovering {
			updatedAt, parseErr := time.Parse(time.RFC3339Nano, strings.TrimSpace(firstText(item["updatedAt"], item["updated_at"])))
			if parseErr != nil {
				item["status"] = "failed"
				item["error"] = "定时派发缺少更新时间，已停止自动恢复"
				item["updatedAt"] = now.UTC().Format(time.RFC3339Nano)
				if err := saveNovelFetchScheduleItem(ctx, store, owner, item); err != nil {
					return err
				}
				continue
			}
			if now.Sub(updatedAt) < novelFetchScheduleRecoveryTimeout {
				continue
			}
		} else if runAt.After(now) {
			continue
		}
		if enabled, explicit := mapOrEmpty(settings[novelFetchAutomationConfigKey])["enabled"].(bool); explicit && !enabled && !recovering {
			if status != "paused" {
				item["status"] = "paused"
				item["error"] = "自动队列已暂停，开启全自动后会继续执行"
				item["updatedAt"] = now.UTC().Format(time.RFC3339Nano)
				if err := saveNovelFetchScheduleItem(ctx, store, owner, item); err != nil {
					return err
				}
			}
			continue
		}

		maxTxt := int(targetNumber(item["max_txt"]))
		if maxTxt <= 0 {
			maxTxt = int(targetNumber(item["maxTxt"]))
		}
		if maxTxt <= 0 {
			maxTxt = 4000
		}
		platformID := firstText(item["platform_id"], item["platformId"], "2")
		scheduleID := firstText(item["id"])
		jobID := firstText(item["job_id"])
		batchID := firstText(item["batch_id"])
		bookIDs := scheduleStringSlice(item["book_ids"])
		var dispatchErr error
		if scheduleID == "" {
			dispatchErr = errors.New("定时任务缺少唯一编号，无法安全派发")
		}
		if !recovering && dispatchErr == nil {
			var allComplete bool
			bookIDs, allComplete, dispatchErr = filterDueScheduleBooks(ctx, store, owner, item)
			if dispatchErr == nil && len(bookIDs) == 0 {
				item["status"], item["error"] = "skipped_existing", "没有可派发任务，请查看每本书的跳过原因"
				if allComplete {
					item["status"] = "completed"
					delete(item, "error")
				}
				item["book_ids"], item["updatedAt"] = bookIDs, now.UTC().Format(time.RFC3339Nano)
				if err := saveNovelFetchScheduleItem(ctx, store, owner, item); err != nil {
					return err
				}
				continue
			}
			newJobID, newBatchID := novelFetchScheduleDispatchIDs(owner, scheduleID, runAt)
			jobID, batchID = newJobID, firstText(batchID, newBatchID)
			item["status"] = "dispatching"
			item["dispatch_id"] = "dispatch-" + strings.TrimPrefix(jobID, "go-schedule-")
			item["job_id"] = jobID
			item["batch_id"] = batchID
			item["book_ids"] = append([]string(nil), bookIDs...)
			item["dispatch_started_at"] = now.UTC().Format(time.RFC3339Nano)
			item["updatedAt"] = now.UTC().Format(time.RFC3339Nano)
			if err := saveNovelFetchScheduleItem(ctx, store, owner, item); err != nil {
				if errors.Is(err, novelfetchworkshop.ErrConfigConflict) {
					return nil
				}
				return err
			}
		} else if dispatchErr == nil && (jobID == "" || batchID == "" || firstText(item["dispatch_id"]) == "" || len(bookIDs) == 0) {
			dispatchErr = errors.New("定时派发记录不完整，已停止自动恢复")
		}
		reconciled := false
		if dispatchErr == nil {
			runs, listErr := store.ListRuns(ctx, owner, "")
			if listErr != nil {
				dispatchErr = listErr
			} else {
				for _, run := range runs {
					if run.RunID != jobID {
						continue
					}
					switch strings.ToLower(strings.TrimSpace(run.Status)) {
					case "running":
						item["status"] = "queued"
						reconciled = true
					case "done", "completed", "succeeded":
						item["status"] = "completed"
						reconciled = true
					case "upload_pending", "pending_verification", "pending_confirmation":
						item["status"] = "pending_verification"
						item["error"] = firstText(run.ErrorMessage, "提交结果待核验，系统不会自动重复提交")
						reconciled = true
					case "failed", "stopped":
						item["status"] = run.Status
						item["error"] = firstText(run.ErrorMessage, "任务已结束，未自动重复执行")
						reconciled = true
					}
					if reconciled {
						break
					}
				}
			}
		}
		if reconciled {
			item["updatedAt"] = now.UTC().Format(time.RFC3339Nano)
			if err := saveNovelFetchScheduleItem(ctx, store, owner, item); err != nil {
				return err
			}
			continue
		}
		sensitiveAIEnabled := novelFetchSensitiveAIEnabled(settings)
		if raw, exists := item["sensitive_ai_enabled"]; exists {
			if explicit, valid := raw.(bool); valid {
				sensitiveAIEnabled = explicit
			}
		}
		job := novelfetchworkshop.ProcessJob{
			ID: jobID, Owner: owner, BatchID: batchID, ScheduleID: scheduleID, BookIDs: bookIDs, PlatformID: platformID, MaxTxt: maxTxt,
			SensitiveAIEnabled: sensitiveAIEnabled,
			AutoSubmit:         item["auto_submit"] != false,
			SelectedVersions:   scheduleStringSlice(item["selected_versions"]),
			RewriteMethods:     scheduleStringMap(item["rewrite_methods"]),
			TextModelID:        firstText(item["text_model_id"], item["textModelId"]),
		}
		if dispatchErr == nil && len(bookIDs) == 0 {
			dispatchErr = errors.New("定时任务没有有效书籍")
		}
		if dispatchErr == nil && recovering {
			dispatchErr = ensureScheduleRecoveryDocuments(ctx, store, owner, item)
		}
		if dispatchErr == nil {
			dispatchErr = novelFetchSnapshotBeforeQueue(ctx, store, job)
		}
		// Automatic dispatch never revokes a manual stop. Only an explicit
		// user retry may clear cancellation flags.
		startedAt := now.UTC().Format(time.RFC3339Nano)
		if dispatchErr == nil {
			dispatchErr = store.PutRun(ctx, owner, novelfetchworkshop.RunRecord{RunID: jobID, BookID: batchID, Stage: "novel_fetch_process", Status: "queued", TextModelID: job.TextModelID, ModelID: "novel-fetch-workshop", ModelDisplayName: "小说获取完整流程", StartedAt: startedAt})
		}
		if dispatchErr == nil {
			uniqueQueue, supported := queue.(novelfetchworkshop.ProcessBookDedupQueue)
			if !supported {
				dispatchErr = errors.New("任务队列不支持防重复派发")
			} else {
				queued, existingID, enqueueErr := uniqueQueue.EnqueueUniqueBooks(ctx, job, 7*24*time.Hour)
				dispatchErr = enqueueErr
				if dispatchErr == nil && !queued && existingID != "" && existingID != job.ID {
					item["status"] = "waiting_retry"
					item["pending_book_ids"] = bookIDs
					item["partial_dispatch"] = true
					item["retry_tracking"] = true
					item["existing_job_id"] = existingID
					item["job_id"], item["batch_id"] = jobID, batchID
					_ = updateWorkshopRun(store, job, "duplicate_skipped", "书目已由其他有效任务处理", startedAt, true)
					if err := saveNovelFetchScheduleItem(ctx, store, owner, item); err != nil {
						return err
					}
					if err := dispatchNovelFetchSchedulePending(ctx, store, queue, owner, item, now); err != nil {
						return err
					}
					continue
				}
			}
		}
		if dispatchErr != nil {
			item["status"] = "failed"
			item["error"] = dispatchErr.Error()
		} else if firstText(item["status"]) != "skipped_existing" {
			item["status"] = "queued"
			item["job_id"] = jobID
			item["batch_id"] = batchID
			delete(item, "error")
		}
		item["updatedAt"] = now.UTC().Format(time.RFC3339Nano)
		if dispatchErr == nil {
			for _, bookID := range bookIDs {
				if firstText(item["status"]) == "queued" {
					_ = markNovelFetchTaskQueued(ctx, store, owner, bookID, job.ID, batchID, now)
				} else {
					_ = appendWorkshopTaskLog(ctx, store, owner, bookID, "schedule_dispatch_duplicate_skipped", "书目已由其他有效任务处理")
				}
			}
		}
		if err := saveNovelFetchScheduleItem(ctx, store, owner, item); err != nil {
			return err
		}
		continue
	}
	return nil
}

func novelFetchScheduleDispatchIDs(owner, scheduleID string, runAt time.Time) (string, string) {
	identity := strings.TrimSpace(owner) + "\x00" + strings.TrimSpace(scheduleID) + "\x00" + runAt.UTC().Format(time.RFC3339Nano)
	hash := sha256.Sum256([]byte(identity))
	suffix := hex.EncodeToString(hash[:12])
	return "go-schedule-" + suffix, "batch-schedule-" + suffix
}

func parseNovelFetchScheduleTime(value string) (time.Time, error) {
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed, nil
	}
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		location = time.FixedZone("CST", 8*60*60)
	}
	return time.ParseInLocation("2006-01-02T15:04", value, location)
}

func scheduleStringSlice(value any) []string {
	items, _ := value.([]any)
	result := make([]string, 0, len(items))
	for _, item := range items {
		if text := strings.TrimSpace(fmt.Sprint(item)); text != "" {
			result = append(result, text)
		}
	}
	if direct, ok := value.([]string); ok {
		return append([]string(nil), direct...)
	}
	return result
}

func scheduleStringMap(value any) map[string]string {
	result := map[string]string{}
	items, _ := value.(map[string]any)
	for key, item := range items {
		result[key] = fmt.Sprint(item)
	}
	if direct, ok := value.(map[string]string); ok {
		return direct
	}
	return result
}

type workshopInputMetadata struct {
	Format       string
	ColumnPreset string
	ColumnOrder  []string
	// Normalized 表示 input_text 已由新版前端按书籍ID/书名/男女频/理由
	// 的固定四列序列化，后端不能再按原始列顺序二次映射。
	Normalized bool
}

func (metadata workshopInputMetadata) apply(meta map[string]any) {
	if value := strings.TrimSpace(metadata.Format); value != "" {
		meta["inputFormat"] = value
	}
	if value := strings.TrimSpace(metadata.ColumnPreset); value != "" {
		meta["columnPreset"] = value
	}
	if len(metadata.ColumnOrder) > 0 {
		meta["columnOrder"] = append([]string(nil), metadata.ColumnOrder...)
	}
}

func (metadata workshopInputMetadata) responseFields() map[string]any {
	result := map[string]any{}
	if value := strings.TrimSpace(metadata.Format); value != "" {
		result["input_format"] = value
	}
	if value := strings.TrimSpace(metadata.ColumnPreset); value != "" {
		result["column_preset"] = value
	}
	if len(metadata.ColumnOrder) > 0 {
		result["column_order"] = append([]string(nil), metadata.ColumnOrder...)
	}
	return result
}

func workshopProcessResult(ctx context.Context, store novelfetchworkshop.Store, owner, inputText, platformID, batchID string, maxTxt int, sourceMetadata ...workshopInputMetadata) (map[string]any, error) {
	if strings.TrimSpace(inputText) == "" {
		return nil, errors.New("请输入要处理的书籍清单")
	}
	if maxTxt <= 0 {
		maxTxt = 4000
	}
	metadata := workshopInputMetadata{}
	if len(sourceMetadata) > 0 {
		metadata = sourceMetadata[0]
	}
	documents, parsed, emptyIDs, duplicates := parseWorkshopDocuments(inputText, maxTxt)
	result, _, _, err := workshopProcessParsedResult(ctx, store, owner, documents, parsed, emptyIDs, duplicates, platformID, batchID, metadata)
	return result, err
}

// workshopProcessParsedResult 复用已解析的书单，避免 /process/start 先解析一次、
// 再在保存阶段重复解析和查库。已有任务是否跳过与定时任务使用同一套判定：
// 只有正在执行、需要人工处理或所选版本已全部完成时才跳过，终态任务可以重新处理。
func workshopProcessParsedResult(ctx context.Context, store novelfetchworkshop.Store, owner string, documents []novelfetchworkshop.Document, parsed, emptyIDs, duplicates int, platformID, batchID string, metadata workshopInputMetadata, selectedVersions ...[]string) (map[string]any, []string, []string, error) {
	settings, _ := store.GetConfig(ctx, owner)
	if settings == nil {
		settings = map[string]any{}
	}
	processingFingerprint := novelFetchProcessingFingerprint(settings)
	versions := novelFetchAutomationVersions(settings)
	if len(selectedVersions) > 0 && len(selectedVersions[0]) > 0 {
		versions = cleanStrings(selectedVersions[0])
	}
	skipItem := map[string]any{"id": "", "selected_versions": versions, "auto_submit": novelFetchAutomationEnabled(settings)}
	processBookIDs := make([]string, 0, len(documents))
	skippedBookIDs := make([]string, 0)
	alreadyExists := 0
	for _, document := range documents {
		existing, err := store.GetDocument(ctx, owner, document.BookID)
		if err == nil {
			alreadyExists++
			if existing.Meta == nil {
				existing.Meta = map[string]any{}
			}
			existing.Meta["lastImportedBatchId"] = batchID
			existing.Meta["lastImportedAt"] = time.Now().UTC().Format(time.RFC3339Nano)
			if reason := scheduleBookSkipReason(ctx, store, owner, existing, skipItem); reason != "" {
				skippedBookIDs = append(skippedBookIDs, document.BookID)
				if err := store.PutDocument(ctx, owner, existing); err != nil {
					return nil, nil, nil, err
				}
				continue
			}
			existing.Meta["platformId"] = platformID
			existing.Meta["currentBatchId"] = batchID
			metadata.apply(existing.Meta)
			if err := store.PutDocument(ctx, owner, existing); err != nil {
				return nil, nil, nil, err
			}
			processBookIDs = append(processBookIDs, document.BookID)
			continue
		}
		if !errors.Is(err, novelfetchworkshop.ErrNotFound) {
			return nil, nil, nil, err
		}
		if document.Meta == nil {
			document.Meta = map[string]any{}
		}
		document.Meta["platformId"] = platformID
		document.Meta["currentBatchId"] = batchID
		document.Meta["processingFingerprint"] = processingFingerprint
		document.Meta["processingConfigRevision"] = "novel-fetch-rules-v1"
		metadata.apply(document.Meta)
		if err := store.PutDocument(ctx, owner, document); err != nil {
			return nil, nil, nil, err
		}
		processBookIDs = append(processBookIDs, document.BookID)
	}
	tasks, err := store.ListDocuments(ctx, owner)
	if err != nil {
		return nil, nil, nil, err
	}
	settings["novel_fetch_current_batch_id"] = batchID
	settings["novel_fetch_current_batch_at"] = time.Now().UTC().Format(time.RFC3339Nano)
	_ = store.PutConfig(ctx, owner, settings)
	result := map[string]any{
		"batch_id": batchID, "current_batch_id": batchID,
		"auto_queue_enabled": novelFetchAutomationEnabled(settings),
		"parsed":             parsed, "uniqueTasks": len(documents), "unique_tasks": len(documents), "duplicateCount": duplicates, "duplicate_count": duplicates,
		"emptyIdCount": emptyIDs, "empty_id_count": emptyIDs, "classifyErrors": []any{}, "classify_errors": []any{},
		"fetched": 0, "fetchFailed": 0, "fetch_failed": 0, "raw_fetched": 0, "generatedAiFiles": 0, "generated_ai_files": 0, "tasks": tasks,
		"alreadyExists": alreadyExists, "already_exists": alreadyExists,
		"skippedExistingBookIds": skippedBookIDs, "skipped_existing_book_ids": skippedBookIDs,
	}
	for key, value := range metadata.responseFields() {
		result[key] = value
	}
	return result, processBookIDs, skippedBookIDs, nil
}

func novelFetchProcessingFingerprint(settings map[string]any) string {
	app := mapOrEmpty(settings["appConfig"])
	rules := map[string]any{
		"sensitive":          firstConfiguredValue(settings["sensitive"], app["sensitive"]),
		"knowledge":          firstConfiguredValue(settings["knowledgeLibraries"], settings["knowledge"], app["knowledge"]),
		"layout":             firstConfiguredValue(settings["layout"], app["layout"]),
		"processing_version": "novel-fetch-rules-v1",
	}
	encoded, _ := json.Marshal(rules)
	digest := sha256.Sum256(encoded)
	return hex.EncodeToString(digest[:])
}

func firstConfiguredValue(values ...any) any {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
}

func decodeSensitiveRepairArray(output string, repaired *[]string) error {
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), repaired); err == nil {
		return nil
	}
	// A few OpenAI-compatible providers prepend labels such as "Text response:"
	// even when JSON mode is requested. Accept the first complete string array
	// while keeping the existing length and sensitive-word checks below.
	for offset, char := range output {
		if char != '[' {
			continue
		}
		candidate := []string{}
		decoder := json.NewDecoder(strings.NewReader(output[offset:]))
		if err := decoder.Decode(&candidate); err == nil {
			*repaired = candidate
			return nil
		}
	}
	return json.Unmarshal([]byte(strings.TrimSpace(output)), repaired)
}

// 与批量原文改文系统源码保持一致的敏感词系统提示。
const sensitiveRepairSystemPrompt = "你是内容合规改写助手。只改写用户给出的命中片段，保持原意，去掉违规、擦边、色情、低俗表达。"

const sensitiveGroupRepairSystemPrompt = "你是内容合规改写助手。请批量处理用户给出的敏感词命中片段。" +
	"每条只改写对应 snippet，不要增加新剧情，不要改变人物关系和原意。" +
	"必须只返回 JSON，不要解释。"

// decodeSensitiveRepairText 兼容源码接受的两种单条返回：纯文本片段，
// 或只包含一个改写字段的 JSON 对象（fixed_text/text/content/result/output）。
// 返回空字符串表示这次输出不可用，交给上层继续重试。
func decodeSensitiveRepairText(output string) string {
	trimmed := strings.TrimSpace(output)
	if strings.HasPrefix(trimmed, "```") {
		trimmed = strings.TrimPrefix(trimmed, "```")
		if index := strings.Index(trimmed, "\n"); index >= 0 {
			trimmed = trimmed[index+1:]
		}
		trimmed = strings.TrimSuffix(strings.TrimSpace(trimmed), "```")
		trimmed = strings.TrimSpace(trimmed)
	}
	if !strings.HasPrefix(trimmed, "{") {
		return trimmed
	}
	var payload map[string]any
	if json.Unmarshal([]byte(trimmed), &payload) != nil {
		return trimmed
	}
	values := make([]string, 0, 1)
	for _, key := range []string{"fixed_text", "text", "content", "result", "output"} {
		if text, ok := payload[key].(string); ok && strings.TrimSpace(text) != "" {
			values = append(values, strings.TrimSpace(text))
		}
	}
	if len(values) != 1 {
		return ""
	}
	return values[0]
}

func sensitiveServiceUnavailable(err error) bool {
	if err == nil {
		return false
	}
	text := strings.ToLower(err.Error())
	for _, marker := range []string{"502", "503", "504", "timeout", "deadline exceeded", "temporarily unavailable", "暂不可用", "connection reset"} {
		if strings.Contains(text, marker) {
			return true
		}
	}
	return false
}

// applySensitiveAIRepair runs only when the user explicitly enables the
// sensitive-word AI repair switch. Deterministic rules still run first; the
// model then refines each hit line and the audit item records the source,
// model and attempt count shown by the workbench.
func applySensitiveAIRepair(ctx context.Context, provider novelfetchworkshop.RewriteProvider, settings map[string]any, bookID, text string, items []map[string]any, bookVars map[string]any) (string, []map[string]any) {
	if len(items) == 0 {
		return text, items
	}
	if settings == nil {
		settings = map[string]any{}
	}
	app := mapOrEmpty(settings["appConfig"])
	ai := mapOrEmpty(settings["sensitive_ai"])
	if len(ai) == 0 {
		ai = mapOrEmpty(app["sensitive_ai"])
	}
	enabled, _ := ai["enabled"].(bool)
	if !enabled {
		return text, items
	}
	model := firstText(ai["model"], ai["model_id"], ai["modelId"], settings["text_model_id"], settings["textModelId"], "未指定")
	promptTemplate := strings.TrimSpace(fmt.Sprint(ai["prompt"]))
	if promptTemplate == "" || promptTemplate == "<nil>" {
		promptTemplate = "你是内容合规改写助手。只改写下面命中敏感词的小段内容，保留原剧情、人物关系和情绪，不增加新剧情。只返回改写后的小段，不要解释。\n\n任务ID：{book_id}\n命中词：{keyword}\n\n原片段：\n{snippet}"
	}
	maxAttempts := toInt(ai["retries"]) + 1
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	if maxAttempts > 3 {
		maxAttempts = 3
	}
	contextChars := toInt(ai["context_chars"])
	if contextChars < 0 {
		contextChars = 0
	}
	maxHits := toInt(ai["max_hits_per_task"])
	if maxHits <= 0 {
		maxHits = len(items)
	}
	if maxHits > len(items) {
		maxHits = len(items)
	}
	groupHits, _ := ai["group_hits"].(bool)
	temperature := 0.2
	if value, ok := ai["temperature"].(float64); ok {
		temperature = value
	}
	for index := maxHits; index < len(items); index++ {
		items[index]["source"] = "ai_group"
		items[index]["model"] = model
		items[index]["status"] = "skipped_limit"
		items[index]["error"] = fmt.Sprintf("超过本任务命中上限 %d", maxHits)
	}
	active := items[:maxHits]
	if groupHits && len(active) > 1 {
		outputs := make([]string, 0, len(active))
		for _, item := range active {
			item["source"] = "ai_group"
			item["model"] = model
			keyword := strings.TrimSpace(fmt.Sprint(item["keyword"]))
			snippet := sensitiveContextSnippet(strings.TrimSpace(fmt.Sprint(item["before"])), keyword, contextChars)
			outputs = append(outputs, fmt.Sprintf("%d. 命中词：%s\n原片段：%s", len(outputs)+1, keyword, snippet))
		}
		userPrompt := strings.NewReplacer(
			"{book_id}", bookID,
			"{book_name}", firstText(bookVars["book_name"], bookVars["bookName"]),
			"{style}", firstText(bookVars["style"]),
			"{gender}", firstText(bookVars["gender"]),
			"{keyword}", "多个命中词",
			"{snippet}", strings.Join(outputs, "\n\n"),
			// 整组模式下没有单个命中序号，避免模板里残留占位符。
			"{hit_index}", "",
		).Replace(promptTemplate)
		userPrompt += "\n\n请严格按原顺序返回 JSON 字符串数组，每个元素对应一个片段，只返回数组。"
		var output string
		var lastErr error
		attempts := 0
		if provider == nil {
			lastErr = errors.New("敏感词 AI 修复模型未配置")
		} else {
			for attempts < maxAttempts {
				attempts++
				output, lastErr = provider.Complete(ctx, novelfetchworkshop.RewriteRequest{SystemPrompt: sensitiveGroupRepairSystemPrompt, UserPrompt: userPrompt, Temperature: temperature, MaxTokens: 1200})
				output = strings.TrimSpace(output)
				if lastErr != nil {
					// 服务不可用时立即停止，避免同一故障连续烧掉全部重试次数。
					if sensitiveServiceUnavailable(lastErr) {
						break
					}
					continue
				}
				if output == "" {
					userPrompt += "\n\n上次返回为空。只返回完整 JSON 数组，不要解释。"
					continue
				}
				break
			}
		}
		var repaired []string
		if lastErr == nil {
			lastErr = decodeSensitiveRepairArray(output, &repaired)
			if lastErr == nil && len(repaired) != len(active) {
				lastErr = fmt.Errorf("AI 整组返回数量不完整：%d/%d", len(repaired), len(active))
			}
		}
		if lastErr == nil {
			for index, repairedText := range repaired {
				if strings.TrimSpace(repairedText) == "" || containsSensitiveKeyword(repairedText, strings.TrimSpace(fmt.Sprint(active[index]["keyword"]))) {
					lastErr = errors.New("AI 整组结果仍包含命中敏感词")
					break
				}
			}
		}
		if lastErr != nil {
			fallbackSafe := strings.TrimSpace(output) != ""
			for _, item := range active {
				after := strings.TrimSpace(fmt.Sprint(item["after"]))
				keyword := strings.TrimSpace(fmt.Sprint(item["keyword"]))
				if after == "" || containsSensitiveKeyword(after, keyword) {
					fallbackSafe = false
					break
				}
			}
			if fallbackSafe {
				for _, item := range active {
					item["attempt"] = attempts
					item["source"] = "deterministic_fallback"
					item["status"] = "done"
					item["error"] = lastErr.Error()
				}
				return text, items
			}
			for _, item := range active {
				item["attempt"] = attempts
				item["status"] = "failed"
				item["error"] = lastErr.Error()
			}
			return text, items
		}
		searchCursor := 0
		for index, item := range active {
			item["attempt"] = attempts
			item["status"] = "done"
			item["error"] = ""
			text, searchCursor = replaceSensitiveHit(text, searchCursor, item, repaired[index])
			item["after"] = repaired[index]
		}
		return text, items
	}
	searchCursor := 0
	concurrency := toInt(ai["concurrency"])
	if concurrency <= 0 {
		concurrency = 1
	}
	if concurrency > 32 {
		concurrency = 32
	}
	if concurrency > len(active) {
		concurrency = len(active)
	}
	outputs := make([]string, len(active))
	repairOne := func(index int, item map[string]any) {
		item["source"] = "ai"
		item["model"] = model
		keyword := strings.TrimSpace(fmt.Sprint(item["keyword"]))
		before := strings.TrimSpace(fmt.Sprint(item["before"]))
		before = sensitiveContextSnippet(before, keyword, contextChars)
		userPrompt := strings.NewReplacer(
			"{book_id}", bookID,
			"{book_name}", firstText(bookVars["book_name"], bookVars["bookName"]),
			"{style}", firstText(bookVars["style"]),
			"{gender}", firstText(bookVars["gender"]),
			"{keyword}", keyword,
			"{snippet}", before,
			"{hit_index}", firstText(item["hit_index"], item["hitIndex"]),
		).Replace(promptTemplate)
		attempts := 0
		var output string
		var lastErr error
		if provider == nil {
			lastErr = errors.New("敏感词 AI 修复模型未配置")
		} else {
			for attempts < maxAttempts {
				attempts++
				output, lastErr = provider.Complete(ctx, novelfetchworkshop.RewriteRequest{SystemPrompt: sensitiveRepairSystemPrompt, UserPrompt: userPrompt, Temperature: temperature, MaxTokens: 1200})
				output = strings.TrimSpace(output)
				if lastErr != nil {
					// 服务不可用时立即停止，避免同一故障连续烧掉全部重试次数。
					if sensitiveServiceUnavailable(lastErr) {
						break
					}
					continue
				}
				if output == "" {
					userPrompt += "\n\n上次返回为空。只返回改写后的短片段，不要解释。"
					continue
				}
				// 源码同时接受纯文本和 {"fixed_text":"..."} 两种返回；
				// 解码后仍无法得到唯一片段时按不合格重试。
				output = decodeSensitiveRepairText(output)
				if output == "" {
					userPrompt += "\n\n上次返回不合格。只返回改写后的短片段或 {\"fixed_text\":\"...\"}，不要解释。"
					continue
				}
				if containsSensitiveKeyword(output, keyword) {
					userPrompt += "\n\n上次返回仍包含命中词。必须彻底改写该词，只返回改写后的短片段。"
					continue
				}
				break
			}
		}
		item["attempt"] = attempts
		if lastErr != nil || output == "" || containsSensitiveKeyword(output, keyword) {
			item["status"] = "failed"
			// firstText 只识别字符串，error 必须显式取文本，否则真实失败原因会被吞掉。
			message := "AI 返回为空或仍包含命中敏感词"
			if lastErr != nil {
				message = lastErr.Error()
			}
			item["error"] = message
			return
		}
		item["status"] = "done"
		item["error"] = ""
		outputs[index] = output
	}
	if concurrency == 1 {
		for index, item := range active {
			repairOne(index, item)
		}
	} else {
		semaphore := make(chan struct{}, concurrency)
		var repairs sync.WaitGroup
		for index, item := range active {
			index := index
			item := item
			repairs.Add(1)
			go func() {
				defer repairs.Done()
				semaphore <- struct{}{}
				defer func() { <-semaphore }()
				repairOne(index, item)
			}()
		}
		repairs.Wait()
	}
	for index, item := range active {
		if !strings.EqualFold(strings.TrimSpace(fmt.Sprint(item["status"])), "done") {
			continue
		}
		text, searchCursor = replaceSensitiveHit(text, searchCursor, item, outputs[index])
		item["after"] = outputs[index]
	}
	return text, items
}

func containsSensitiveKeyword(output, keyword string) bool {
	return strings.TrimSpace(keyword) != "" && strings.Contains(strings.ToLower(output), strings.ToLower(keyword))
}

func sensitiveContextSnippet(text, keyword string, contextChars int) string {
	if contextChars <= 0 || len([]rune(text)) <= contextChars*2 || keyword == "" {
		return text
	}
	index := strings.Index(strings.ToLower(text), strings.ToLower(keyword))
	if index < 0 {
		return string([]rune(text)[:sensitiveMinInt(len([]rune(text)), contextChars*2)])
	}
	runes := []rune(text)
	center := len([]rune(text[:index]))
	start := center - contextChars
	if start < 0 {
		start = 0
	}
	end := center + len([]rune(keyword)) + contextChars
	if end > len(runes) {
		end = len(runes)
	}
	return string(runes[start:end])
}

func replaceSensitiveHit(text string, cursor int, item map[string]any, output string) (string, int) {
	target := fmt.Sprint(item["after"])
	if target == "" {
		target = fmt.Sprint(item["before"])
	}
	index := strings.Index(text[cursor:], target)
	if index >= 0 {
		index += cursor
		return text[:index] + output + text[index+len(target):], index + len(output)
	}
	if index := strings.Index(text, target); index >= 0 {
		return text[:index] + output + text[index+len(target):], index + len(output)
	}
	return text, cursor
}

func sensitiveMinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func processWorkshopConfiguredText(ctx context.Context, store novelfetchworkshop.Store, owner, bookID, text, scope string, sensitiveAIEnabled bool, provider novelfetchworkshop.RewriteProvider) (string, []map[string]any, []map[string]any, error) {
	settings, err := store.GetConfig(ctx, owner)
	if err != nil {
		return text, nil, nil, err
	}
	processed, trace, sensitiveItems := novelfetchworkshop.ProcessConfiguredText(text, scope, settings)
	if sensitiveAIEnabled && len(sensitiveItems) > 0 {
		ai := mapOrEmpty(settings["sensitive_ai"])
		if len(ai) == 0 {
			ai = mapOrEmpty(mapOrEmpty(settings["appConfig"])["sensitive_ai"])
		}
		if len(ai) == 0 {
			ai = map[string]any{}
		}
		ai["enabled"] = true
		settings["sensitive_ai"] = ai
		before := processed
		bookVars := map[string]any{}
		if document, docErr := store.GetDocument(ctx, owner, bookID); docErr == nil {
			bookVars["book_name"] = firstText(document.Meta["bookName"], document.Meta["book_name"], document.Meta["title"])
			bookVars["style"] = firstText(document.Meta["style"])
			bookVars["gender"] = document.Gender
		}
		processed, sensitiveItems = applySensitiveAIRepair(ctx, provider, settings, bookID, processed, sensitiveItems, bookVars)
		failed := 0
		for _, item := range sensitiveItems {
			if strings.EqualFold(strings.TrimSpace(fmt.Sprint(item["status"])), "failed") {
				failed++
			}
		}
		trace = append(trace, map[string]any{"stage": "sensitive_ai", "title": "敏感词 AI 修复", "changed": before != processed, "chars": len([]rune(processed)), "failed": failed})
		if failed > 0 {
			resolved := 0
			masked := 0
			for _, item := range sensitiveItems {
				if !strings.EqualFold(strings.TrimSpace(fmt.Sprint(item["status"])), "failed") {
					continue
				}
				keyword := strings.TrimSpace(fmt.Sprint(item["keyword"]))
				if keyword == "" || keyword == "<nil>" {
					continue
				}
				if strings.Contains(processed, keyword) {
					processed = strings.ReplaceAll(processed, keyword, strings.Repeat("＊", len([]rune(keyword))))
					item["source"] = "mask"
					item["masked"] = true
					masked++
				} else {
					item["source"] = "fallback"
				}
				item["status"] = "done"
				resolved++
			}
			if resolved == failed {
				trace = append(trace, map[string]any{"stage": "sensitive_fallback", "title": "敏感词自动兜底", "changed": true, "chars": len([]rune(processed)), "resolved": resolved, "masked": masked})
				return processed, trace, sensitiveItems, nil
			}
			return processed, trace, sensitiveItems, fmt.Errorf("敏感词 AI 修复失败 %d 处，已停止后续处理", failed)
		}
	}
	return processed, trace, sensitiveItems, nil
}

func workshopRewriteProcessor(store novelfetchworkshop.Store, owner, bookID string, sensitiveAIEnabled bool, provider novelfetchworkshop.RewriteProvider) novelfetchworkshop.RewriteTextProcessor {
	return func(ctx context.Context, versionID, text string) (string, []map[string]any, []map[string]any, error) {
		return processWorkshopConfiguredText(ctx, store, owner, bookID, text, "ai", sensitiveAIEnabled, provider)
	}
}

// executeWorkshopDocumentFetch is the shared Go-side fetch stage used by the
// full-workflow start endpoint. It keeps the sensitive-rule pass and optional
// AI repair in the same order as the manual fetch/reprocess buttons.
func stageNovelFetchRaw(ctx context.Context, store novelfetchworkshop.Store, owner string, document *novelfetchworkshop.Document, result novelfetchworkshop.FetchAndProcessResult) error {
	if document == nil {
		return errors.New("小说任务不存在")
	}
	document.OriginalRaw = result.OriginalRaw
	document.OriginalRawChars = result.OriginalRawChars
	document.MaxTxt = result.MaxTxt
	document.BookInfo = result.BookInfo
	if document.Meta == nil {
		document.Meta = map[string]any{}
	}
	document.Meta["originalStatus"] = "waiting_sensitive_ai"
	document.Meta["originalRawStatus"] = "done"
	document.Meta["originalRawChars"] = result.OriginalRawChars
	document.Meta["queueStage"] = "sensitive_ai"
	document.Meta["sensitiveStatus"] = "pending"
	if err := persistNovelFetchBodies(ctx, store, owner, document, result.OriginalRaw, ""); err != nil {
		return err
	}
	if latest, latestErr := store.GetDocument(ctx, owner, document.BookID); latestErr == nil {
		document.Logs = append([]any(nil), latest.Logs...)
	}
	return store.PutDocument(ctx, owner, *document)
}

func novelFetchShortOriginalLimit(ctx context.Context, store novelfetchworkshop.Store, owner string) (int, bool) {
	settings, err := store.GetConfig(ctx, owner)
	if err != nil {
		return 0, false
	}
	workflow := mapOrEmpty(settings["workflow"])
	if len(workflow) == 0 {
		return 0, false
	}
	if value, ok := workflow["skip_short_original"].(bool); ok && !value {
		return 0, false
	}
	limit := intValue(workflow["min_original_chars"])
	if limit <= 0 {
		limit = 2000
	}
	return limit, true
}

func markNovelFetchShortOriginal(ctx context.Context, store novelfetchworkshop.Store, owner string, document *novelfetchworkshop.Document, result novelfetchworkshop.FetchAndProcessResult, limit int) error {
	document.Original = result.ProcessedText
	document.ProcessedText = result.ProcessedText
	document.ProcessedChars = result.ProcessedChars
	if document.Meta == nil {
		document.Meta = map[string]any{}
	}
	document.Meta["originalStatus"] = "too_short"
	document.Meta["originalTooShort"] = true
	document.Meta["originalMinChars"] = limit
	document.Meta["status"] = "original_too_short"
	document.Meta["aiStatus"] = "skipped_short_original"
	document.Meta["sensitiveStatus"] = "skipped"
	document.Meta["queueStage"] = "completed"
	if err := persistNovelFetchBodies(ctx, store, owner, document, result.OriginalRaw, result.ProcessedText); err != nil {
		return err
	}
	if latest, latestErr := store.GetDocument(ctx, owner, document.BookID); latestErr == nil {
		document.Logs = append([]any(nil), latest.Logs...)
	}
	return store.PutDocument(ctx, owner, *document)
}

// fetchNovelWithSettings applies the per-account fetch settings that the
// legacy workbench exposed. The process queue remains the outer recovery
// mechanism; these retries are the short, immediate request retries inside a
// single fetch attempt.
func fetchNovelWithSettings(ctx context.Context, store novelfetchworkshop.Store, owner string, fetcher *novelfetchworkshop.UpstreamFetcher, bookID, platformID string, maxTxt int) (novelfetchworkshop.FetchAndProcessResult, error) {
	if fetcher == nil {
		return novelfetchworkshop.FetchAndProcessResult{}, errors.New("原文获取器未配置")
	}
	configured := *fetcher
	retries := 0
	if settings, err := store.GetConfig(ctx, owner); err == nil {
		fetchConfig := mapOrEmpty(settings["fetch"])
		if len(fetchConfig) == 0 {
			fetchConfig = mapOrEmpty(mapOrEmpty(settings["appConfig"])["fetch"])
		}
		if endpoint := strings.TrimSpace(fmt.Sprint(fetchConfig["endpoint"])); endpoint != "" && endpoint != "<nil>" {
			// Keep an explicitly injected endpoint (used by local/test upstreams)
			// when the stored value is only the legacy default.
			if strings.TrimSpace(configured.Endpoint) == "" || endpoint != novelfetchworkshop.DefaultUpstreamEndpoint {
				configured.Endpoint = endpoint
			}
		}
		if seconds := intValue(fetchConfig["timeout_seconds"]); seconds > 0 {
			base := configured.Client
			if base == nil {
				base = safehttp.NewClient(safehttp.Options{})
			}
			client := *base
			client.Timeout = time.Duration(seconds) * time.Second
			configured.Client = &client
		}
		retries = intValue(fetchConfig["retries"])
		if retries < 0 {
			retries = 0
		}
		if retries > 5 {
			retries = 5
		}
	}
	var result novelfetchworkshop.FetchAndProcessResult
	var err error
	for attempt := 0; attempt <= retries; attempt++ {
		result, err = configured.FetchAndProcess(ctx, bookID, platformID, maxTxt)
		if err == nil || ctx.Err() != nil || attempt == retries {
			break
		}
	}
	return result, err
}

// novelFetchSensitiveStageError 标记失败发生在敏感词处理阶段，
// 让上层按 sensitive 阶段记录退避，而不是重复记成原文抓取失败。
type novelFetchSensitiveStageError struct{ err error }

func (e *novelFetchSensitiveStageError) Error() string { return e.err.Error() }
func (e *novelFetchSensitiveStageError) Unwrap() error { return e.err }

func executeWorkshopDocumentFetch(ctx context.Context, store novelfetchworkshop.Store, owner, bookID string, fetcher *novelfetchworkshop.UpstreamFetcher, maxTxt int, sensitiveAIEnabled bool, sensitiveProvider novelfetchworkshop.RewriteProvider, genderJudge novelfetchworkshop.AIGenderJudge, styleProvider novelfetchworkshop.RewriteProvider) error {
	document, err := store.GetDocument(ctx, owner, bookID)
	if err != nil {
		return err
	}
	if maxTxt <= 0 {
		maxTxt = document.MaxTxt
	}
	if maxTxt <= 0 {
		maxTxt = 4000
	}
	var result novelfetchworkshop.FetchAndProcessResult
	if strings.TrimSpace(document.OriginalRaw) != "" {
		_ = appendWorkshopTaskLog(ctx, store, owner, bookID, "rules_processing_started", "使用已保存原文继续处理")
		processed := novelfetchworkshop.ProcessRawText(document.OriginalRaw, maxTxt)
		result = novelfetchworkshop.FetchAndProcessResult{
			OriginalRaw: document.OriginalRaw, ProcessedText: processed.ProcessedText,
			BookInfo: document.BookInfo, OriginalRawChars: processed.OriginalRawChars,
			ProcessedChars: processed.ProcessedChars, MaxTxt: processed.MaxTxt,
		}
	} else {
		_ = appendWorkshopTaskLog(ctx, store, owner, bookID, "fetch_started", "开始获取原文")
		// 平台 ID 可能只以名称形式存在（历史导入/批量工厂），这里按平台目录
		// 自动换算，避免把任务发到默认书城后拿到错误的 400。
		fetchSettings, _ := store.GetConfig(ctx, owner)
		platformID := novelFetchResolvedPlatformID(fetchSettings, document.Meta)
		result, err = fetchNovelWithSettings(ctx, store, owner, fetcher, bookID, platformID, maxTxt)
		if err != nil {
			_ = appendWorkshopTaskLog(ctx, store, owner, bookID, "fetch_failed", err.Error())
			return fmt.Errorf("%w: %v", errNovelFetchOriginalFetch, err)
		}
		if err := stageNovelFetchRaw(ctx, store, owner, &document, result); err != nil {
			_ = appendWorkshopTaskLog(ctx, store, owner, bookID, "original_raw_save_failed", err.Error())
			return err
		}
		_ = appendWorkshopTaskLog(ctx, store, owner, bookID, "fetch_completed", "原文获取完成")
		_ = appendWorkshopTaskLog(ctx, store, owner, bookID, "original_raw_saved", "完整原文已单独保存")
	}
	if limit, enabled := novelFetchShortOriginalLimit(ctx, store, owner); enabled && result.ProcessedChars > 0 && result.ProcessedChars < limit {
		return markNovelFetchShortOriginal(ctx, store, owner, &document, result, limit)
	}
	_ = appendWorkshopTaskLog(ctx, store, owner, bookID, "rules_processing_started", "开始执行正文规则")
	_ = appendWorkshopTaskLog(ctx, store, owner, bookID, "sensitive_processing_started", "开始处理敏感词")
	processedText, rulesTrace, sensitiveItems, err := processWorkshopConfiguredText(ctx, store, owner, bookID, result.ProcessedText, "original", sensitiveAIEnabled, sensitiveProvider)
	if err != nil {
		_ = appendWorkshopTaskLog(ctx, store, owner, bookID, "sensitive_processing_failed", err.Error())
		// 与源码一致：敏感词处理失败单独记在 sensitive 阶段，
		// 不和原文抓取失败共用同一段退避时间。
		_ = scheduleNovelFetchDocumentRetry(ctx, store, owner, bookID, novelFetchRetryStageSensitive, err, time.Now().UTC())
		return &novelFetchSensitiveStageError{err: err}
	}
	_ = appendWorkshopTaskLog(ctx, store, owner, bookID, "rules_processing_completed", "正文规则处理完成")
	sensitiveCount := len(sensitiveItems)
	document.ProcessedText = processedText
	document.Original = processedText
	document.ProcessedChars = len([]rune(processedText))
	if document.Meta == nil {
		document.Meta = map[string]any{}
	}
	document.Meta["originalStatus"] = "done"
	document.Meta["originalRawChars"] = result.OriginalRawChars
	document.Meta["processedChars"] = document.ProcessedChars
	document.Meta["sensitiveStatus"] = "done"
	document.Meta["sensitiveReplaced"] = sensitiveCount
	document.Meta["rulesTrace"] = rulesTrace
	document.Meta["rulesTraceByVersion"] = map[string]any{"original": rulesTrace}
	setSensitiveDetails(document.Meta, sensitiveItems, sensitiveCount)
	if err := persistNovelFetchBodies(ctx, store, owner, &document, result.OriginalRaw, processedText); err != nil {
		return err
	}
	_ = appendWorkshopTaskLog(ctx, store, owner, bookID, "sensitive_processing_completed", "敏感词处理结果已保存")
	needsGender := genderNeedsResolution(document.Gender)
	needsStyle := styleNeedsResolution(document.Meta)
	var classification novelfetchworkshop.ClassificationResult
	if needsGender || needsStyle {
		_ = appendWorkshopTaskLog(ctx, store, owner, bookID, "classification_started", "开始判断男女频和题材风格")
		classificationInfo := result.BookInfo
		classificationInfo.Text = processedText
		classification = novelfetchworkshop.ClassifyBook(ctx, "", classificationInfo, genderJudge, styleProvider)
	}
	if needsGender {
		decision := classification.Gender
		document.Gender, document.GenderSource, document.GenderReason = decision.Gender, decision.Source, decision.Reason
		document.GenderConfidence = decision.Confidence
	}
	if needsStyle {
		document.Meta["styleStatus"] = "processing"
		recordNovelFetchStyleDecision(document.Meta, classification.Style, nil)
	}
	if latest, latestErr := store.GetDocument(ctx, owner, bookID); latestErr == nil {
		document.Logs = append([]any(nil), latest.Logs...)
	}
	if err := store.PutDocument(ctx, owner, document); err != nil {
		return err
	}
	if needsGender || needsStyle {
		_ = appendWorkshopTaskLog(ctx, store, owner, bookID, "classification_completed", "男女频和题材风格判断完成")
	} else {
		_ = appendWorkshopTaskLog(ctx, store, owner, bookID, "classification_skipped", "已有分类信息，跳过判断")
	}
	return nil
}

// workshopPreviewResult parses the same input as the start endpoint without
// writing documents. The V78 workbench uses this as a first step to show the
// user how many books were recognised before starting the real run.
func workshopPreviewResult(inputText string, maxTxt int, sourceMetadata ...workshopInputMetadata) (map[string]any, error) {
	if strings.TrimSpace(inputText) == "" {
		return nil, errors.New("请输入要处理的书籍清单")
	}
	if maxTxt <= 0 {
		maxTxt = 4000
	}
	metadata := workshopInputMetadata{}
	if len(sourceMetadata) > 0 {
		metadata = sourceMetadata[0]
	}
	documents, parsed, emptyIDs, duplicates := parseWorkshopDocuments(inputText, maxTxt, metadata)
	tasks := make([]map[string]any, 0, len(documents))
	for _, document := range documents {
		meta := map[string]any{}
		for key, value := range document.Meta {
			meta[key] = value
		}
		meta["id"] = document.BookID
		meta["book_id"] = document.BookID
		tasks = append(tasks, meta)
	}
	result := map[string]any{
		"parsed": parsed, "unique_tasks": len(tasks), "duplicate_count": duplicates,
		"empty_id_count": emptyIDs, "tasks": tasks,
	}
	for key, value := range metadata.responseFields() {
		result[key] = value
	}
	return result, nil
}

func registerNovelFetchWorkshopRoutes(mux *http.ServeMux, store novelfetchworkshop.Store, queue novelfetchworkshop.ProcessQueue, aiJudge novelfetchworkshop.AIGenderJudge, aiJudgeResolver func(context.Context, string) (novelfetchworkshop.AIGenderJudge, error), fetcher *novelfetchworkshop.UpstreamFetcher, rewriteProvider novelfetchworkshop.RewriteProvider, rewriteResolver func(context.Context, string) (novelfetchworkshop.RewriteProvider, error), usersDir string, uploadTarget videomanagementconnector.Service, profiles *MemberProfileStore, accounts auth.Store) {
	resolveGenderJudgeFor := func(ctx context.Context, owner string) novelfetchworkshop.AIGenderJudge {
		if aiJudgeResolver == nil {
			return aiJudge
		}
		resolved, err := aiJudgeResolver(ctx, owner)
		if err == nil && resolved != nil {
			return resolved
		}
		return aiJudge
	}
	resolveGenderJudge := func(req *http.Request) novelfetchworkshop.AIGenderJudge {
		identity, _ := BridgeIdentityFromContext(req.Context())
		return resolveGenderJudgeFor(req.Context(), identity.Username)
	}
	resolveRewriteProviderFor := func(ctx context.Context, owner string) (novelfetchworkshop.RewriteProvider, error) {
		if rewriteResolver != nil {
			resolved, err := rewriteResolver(ctx, owner)
			if err != nil {
				return nil, err
			}
			if resolved != nil {
				return resolved, nil
			}
		}
		return rewriteProvider, nil
	}
	resolveRewriteProvider := func(req *http.Request) (novelfetchworkshop.RewriteProvider, error) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		return resolveRewriteProviderFor(req.Context(), identity.Username)
	}
	registry := newWorkshopProcessRegistry()
	if queue != nil {
		worker := &workshopProcessWorker{store: store, queue: queue, registry: registry, fetcher: fetcher, resolveGenderJudge: resolveGenderJudgeFor, resolveRewriteProvider: resolveRewriteProviderFor}
		worker.autoSubmit = func(ctx context.Context, job novelfetchworkshop.ProcessJob) error {
			settings, err := store.GetConfig(ctx, job.Owner)
			if err != nil {
				return err
			}
			batchSize := intValue(mapOrEmpty(settings["workflow"])["auto_submit_batch_size"])
			if batchSize <= 0 {
				batchSize = 12
			}
			if batchSize > 1000 {
				batchSize = 1000
			}
			for _, ids := range novelFetchSubmitChunks(job.BookIDs, batchSize) {
				result, submitErr := executeWebSubmit(ctx, job.Owner, usersDir, store, webSubmitRequest{TaskIDs: ids, Versions: job.SelectedVersions, Queue: queue}, func(ctx context.Context, owner string) (novelfetchworkshop.AIGenderJudge, error) {
					return resolveGenderJudgeFor(ctx, owner), nil
				}, uploadTarget)
				if submitErr == nil && result.Failed > 0 {
					submitErr = fmt.Errorf("%d 个版本待上传", result.Failed)
				}
				if submitErr != nil {
					markNovelFetchAutomationSubmitResult(ctx, store, job.Owner, ids, submitErr)
					return submitErr
				}
			}
			markNovelFetchAutomationSubmitResult(ctx, store, job.Owner, job.BookIDs, nil)
			return nil
		}
		go worker.run(context.Background())
		if owners, ok := store.(novelfetchworkshop.ConfigOwnerLister); ok {
			go runNovelFetchScheduleDispatcher(context.Background(), store, queue, owners, registry, func(ctx context.Context, owner string) {
				recoverWorkshopSubmitAccount(ctx, store, queue, owner, usersDir, uploadTarget)
			})
		}
	}
	mux.HandleFunc("POST /api/novel-fetch-workshop/automation/start", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		if queue == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "Redis 任务队列未配置"})
			return
		}
		if err := saveNovelFetchAutomationEnabled(req.Context(), store, identity.Username, true); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "保存全自动状态失败"})
			return
		}
		job, err := dispatchNovelFetchAutomationOwner(req.Context(), store, queue, identity.Username, time.Now())
		if err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "自动队列启动失败：" + err.Error()})
			return
		}
		if job != nil {
			now := time.Now().UTC().Format(time.RFC3339Nano)
			registry.put(identity.Username, job.ID, map[string]any{"id": job.ID, "status": "queued", "queue_state": "queued", "started_at": now, "updated_at": now, "steps": []any{}, "progress": []any{}})
		}
		status, err := novelFetchAutomationStatus(req.Context(), store, registry, identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "读取全自动状态失败"})
			return
		}
		writeJSON(w, http.StatusOK, status)
	})
	mux.HandleFunc("POST /api/novel-fetch-workshop/automation/pause", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		if err := saveNovelFetchAutomationEnabled(req.Context(), store, identity.Username, false); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "保存全自动状态失败"})
			return
		}
		status, err := novelFetchAutomationStatus(req.Context(), store, registry, identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "读取全自动状态失败"})
			return
		}
		writeJSON(w, http.StatusOK, status)
	})
	mux.HandleFunc("POST /api/novel-fetch-workshop/process", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		var input struct {
			InputText          string `json:"inputText"`
			InputTextLegacy    string `json:"input_text"`
			MaxTxt             int    `json:"maxTxt"`
			MaxTxtLegacy       int    `json:"max_txt"`
			PlatformID         string `json:"platform_id"`
			PlatformIDCamel    string `json:"platformId"`
			SensitiveAIEnabled bool   `json:"sensitive_ai_enabled"`
		}
		if err := decodeNovelFetchJSON(w, req, &input); err != nil {
			return
		}
		if strings.TrimSpace(input.InputText) == "" {
			input.InputText = input.InputTextLegacy
		}
		if input.MaxTxt <= 0 {
			input.MaxTxt = input.MaxTxtLegacy
		}
		if strings.TrimSpace(input.PlatformID) == "" {
			input.PlatformID = input.PlatformIDCamel
		}
		if strings.TrimSpace(input.PlatformID) == "" {
			input.PlatformID = "2"
		}
		if strings.TrimSpace(input.InputText) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "请输入要处理的书籍清单"})
			return
		}
		if input.MaxTxt <= 0 {
			input.MaxTxt = 4000
		}
		if novelFetchMaxTxtInvalid(input.MaxTxt) {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": fmt.Sprintf("maxTxt 不能超过 %d", novelFetchMaxTxtLimit)})
			return
		}
		batchID := fmt.Sprintf("batch-%d", time.Now().UnixNano())
		result, err := workshopProcessResult(req.Context(), store, identity.Username, input.InputText, input.PlatformID, batchID, input.MaxTxt)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "保存任务失败"})
			return
		}
		writeJSON(w, http.StatusOK, result)
	})

	mux.HandleFunc("POST /api/novel-fetch-workshop/process/preview", func(w http.ResponseWriter, req *http.Request) {
		var input struct {
			InputText          string          `json:"input_text"`
			InputTextLegacy    string          `json:"inputText"`
			MaxTxt             int             `json:"max_txt"`
			MaxTxtLegacy       int             `json:"maxTxt"`
			ParseMode          string          `json:"parse_mode"`
			InputFormat        string          `json:"input_format"`
			InputFormatLegacy  string          `json:"inputFormat"`
			ColumnPresetID     string          `json:"column_preset_id"`
			ColumnPreset       string          `json:"column_preset"`
			ColumnPresetLegacy string          `json:"columnPreset"`
			ColumnOrderRaw     json.RawMessage `json:"column_order"`
			ColumnOrderLegacy  json.RawMessage `json:"columnOrder"`
			InputNormalized    bool            `json:"input_normalized"`
		}
		if err := decodeNovelFetchJSON(w, req, &input); err != nil {
			return
		}
		if strings.TrimSpace(input.InputText) == "" {
			input.InputText = input.InputTextLegacy
		}
		if input.MaxTxt <= 0 {
			input.MaxTxt = input.MaxTxtLegacy
		}
		if novelFetchMaxTxtInvalid(input.MaxTxt) {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": fmt.Sprintf("max_txt 不能超过 %d", novelFetchMaxTxtLimit)})
			return
		}
		columnOrder := workshopColumnOrderRaw(input.ColumnOrderRaw)
		if len(columnOrder) == 0 {
			columnOrder = workshopColumnOrderRaw(input.ColumnOrderLegacy)
		}
		normalized := input.InputNormalized || strings.TrimSpace(firstText(input.InputFormat, input.InputFormatLegacy)) != ""
		metadata := workshopInputMetadata{Format: firstText(input.ParseMode, input.InputFormat, input.InputFormatLegacy), ColumnPreset: firstText(input.ColumnPresetID, input.ColumnPreset, input.ColumnPresetLegacy), ColumnOrder: columnOrder, Normalized: normalized}
		result, err := workshopPreviewResult(input.InputText, input.MaxTxt, metadata)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, result)
	})

	mux.HandleFunc("POST /api/novel-fetch-workshop/process/start", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		var input struct {
			InputText          string            `json:"inputText"`
			InputTextLegacy    string            `json:"input_text"`
			MaxTxt             int               `json:"maxTxt"`
			MaxTxtLegacy       int               `json:"max_txt"`
			PlatformID         string            `json:"platform_id"`
			PlatformIDCamel    string            `json:"platformId"`
			SensitiveAIEnabled bool              `json:"sensitive_ai_enabled"`
			SelectedVersions   []string          `json:"selected_versions"`
			RewriteMethods     map[string]string `json:"rewrite_methods"`
			TextModelID        string            `json:"text_model_id"`
			InputFormat        string            `json:"input_format"`
			InputFormatLegacy  string            `json:"inputFormat"`
			ParseMode          string            `json:"parse_mode"`
			ColumnPresetID     string            `json:"column_preset_id"`
			ColumnPreset       string            `json:"column_preset"`
			ColumnPresetLegacy string            `json:"columnPreset"`
			ColumnOrderRaw     json.RawMessage   `json:"column_order"`
			ColumnOrderLegacy  json.RawMessage   `json:"columnOrder"`
			InputNormalized    bool              `json:"input_normalized"`
		}
		if err := decodeNovelFetchJSON(w, req, &input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "请求内容无效"})
			return
		}
		if strings.TrimSpace(input.InputText) == "" {
			input.InputText = input.InputTextLegacy
		}
		if input.MaxTxt <= 0 {
			input.MaxTxt = input.MaxTxtLegacy
		}
		if strings.TrimSpace(input.PlatformID) == "" {
			input.PlatformID = input.PlatformIDCamel
		}
		if strings.TrimSpace(input.PlatformID) == "" {
			input.PlatformID = "2"
		}
		if input.MaxTxt <= 0 {
			input.MaxTxt = 4000
		}
		if novelFetchMaxTxtInvalid(input.MaxTxt) {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": fmt.Sprintf("maxTxt 不能超过 %d", novelFetchMaxTxtLimit)})
			return
		}
		normalizedVersions, err := normalizeNovelFetchAIVersions(input.SelectedVersions)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		input.SelectedVersions = normalizedVersions
		if queue == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "Redis 任务队列未配置"})
			return
		}
		now := time.Now().UTC().Format(time.RFC3339Nano)
		id := fmt.Sprintf("go-workshop-%d", time.Now().UnixNano())
		batchID := fmt.Sprintf("batch-%d", time.Now().UnixNano())
		jobView := map[string]any{"id": id, "status": "queued", "queue_state": "queued", "started_at": now, "updated_at": now, "attempts": 0, "steps": []any{}, "progress": []any{}}
		// 空选择表示使用账号默认文本模型；不能写入“未指定”占位符，
		// 否则解析器会把它当成显式模型而跳过默认回退，AI 阶段必然失败。
		textModelID := strings.TrimSpace(input.TextModelID)
		columnOrder := workshopColumnOrderRaw(input.ColumnOrderRaw)
		if len(columnOrder) == 0 {
			columnOrder = workshopColumnOrderRaw(input.ColumnOrderLegacy)
		}
		normalized := input.InputNormalized || strings.TrimSpace(firstText(input.InputFormat, input.InputFormatLegacy)) != ""
		metadata := workshopInputMetadata{Format: firstText(input.ParseMode, input.InputFormat, input.InputFormatLegacy), ColumnPreset: firstText(input.ColumnPresetID, input.ColumnPreset, input.ColumnPresetLegacy), ColumnOrder: columnOrder, Normalized: normalized}
		documents, parsed, emptyIDs, duplicates := parseWorkshopDocuments(input.InputText, input.MaxTxt, metadata)
		result, bookIDs, skippedBookIDs, err := workshopProcessParsedResult(req.Context(), store, identity.Username, documents, parsed, emptyIDs, duplicates, input.PlatformID, batchID, metadata, input.SelectedVersions)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "保存任务失败"})
			return
		}
		result["skippedExistingBookIds"] = skippedBookIDs
		result["skipped_existing_book_ids"] = skippedBookIDs
		if len(bookIDs) == 0 {
			if existing := activeNovelFetchJobForBooks(req.Context(), store, identity.Username, skippedBookIDs); existing != nil {
				jobView["id"] = firstText(existing["id"])
				jobView["status"] = firstText(existing["status"], "queued")
				jobView["queue_state"] = firstText(existing["queue_state"], existing["status"], "queued")
				jobView["existing"] = true
				jobView["result"] = result
				writeJSON(w, http.StatusAccepted, jobView)
				return
			}
			jobView["status"] = "done"
			jobView["queue_state"] = "done"
			jobView["result"] = result
			writeJSON(w, http.StatusOK, jobView)
			return
		}
		processJob := novelfetchworkshop.ProcessJob{ID: id, Owner: identity.Username, BatchID: batchID, BookIDs: bookIDs, PlatformID: input.PlatformID, MaxTxt: input.MaxTxt, SensitiveAIEnabled: input.SensitiveAIEnabled, AutoSubmit: result["auto_queue_enabled"] == true, SelectedVersions: input.SelectedVersions, RewriteMethods: input.RewriteMethods, TextModelID: textModelID}
		if err := novelFetchSnapshotBeforeQueue(req.Context(), store, processJob); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "保存执行参数失败"})
			return
		}
		if cancellationQueue, ok := queue.(novelfetchworkshop.ProcessCancellationQueue); ok {
			if err := cancellationQueue.ClearCancelledBooks(req.Context(), identity.Username, bookIDs); err != nil {
				writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "Redis 取消状态清理失败"})
				return
			}
		}
		jobView["result"] = result
		uniqueQueue, supported := queue.(novelfetchworkshop.ProcessBookDedupQueue)
		if !supported {
			_ = updateWorkshopRun(store, processJob, "failed", "任务队列不支持按书籍防重复入队", now, true)
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "任务队列不支持防重复入队"})
			return
		}
		// 队列入队是 all-or-nothing：只要一本书已被占用，整批都不会入队。
		// 这里先剔除被占用的书，只把剩余的书籍作为同一批次入队，避免其余
		// 书籍被错误标记为“排队中”却永远没有任务处理。
		conflicts := map[string]string{}
		if conflictQueue, ok := queue.(novelfetchworkshop.ProcessBookConflictQueue); ok {
			detected, err := conflictQueue.ConflictingBookIDs(req.Context(), processJob)
			if err != nil {
				writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "Redis 队列状态读取失败"})
				return
			}
			conflicts = detected
		}
		queuedBookIDs := make([]string, 0, len(bookIDs))
		conflictBookIDs := make([]string, 0, len(conflicts))
		for _, bookID := range bookIDs {
			existingID, taken := conflicts[bookID]
			if !taken {
				queuedBookIDs = append(queuedBookIDs, bookID)
				continue
			}
			conflictBookIDs = append(conflictBookIDs, bookID)
			_ = setNovelFetchBookQueueState(req.Context(), store, identity.Username, bookID, existingID, "", "waiting_retry")
			_ = updateWorkshopTaskRuntimeState(req.Context(), store, identity.Username, bookID, "waiting_retry", "该书已有任务在运行，已跳过本次开始")
		}
		if len(queuedBookIDs) == 0 {
			_ = updateWorkshopRun(store, processJob, "duplicate_skipped", "同一书籍已有有效队列任务", now, true)
			jobView["id"] = id
			for _, bookID := range conflictBookIDs {
				if existingID := strings.TrimSpace(conflicts[bookID]); existingID != "" {
					jobView["id"] = existingID
					break
				}
			}
			jobView["existing"] = true
			jobView["conflicting_book_ids"] = conflictBookIDs
			registry.put(identity.Username, id, jobView)
			writeJSON(w, http.StatusAccepted, jobView)
			return
		}
		processJob.BookIDs = queuedBookIDs
		for _, bookID := range queuedBookIDs {
			if err := setNovelFetchBookQueueState(req.Context(), store, identity.Username, bookID, id, batchID, "queued"); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "保存任务排队状态失败"})
				return
			}
		}
		if err := store.PutRun(req.Context(), identity.Username, novelfetchworkshop.RunRecord{RunID: id, BookID: batchID, Stage: "novel_fetch_process", Status: "queued", TextModelID: textModelID, ModelID: "novel-fetch-workshop", ModelDisplayName: "小说获取完整流程", StartedAt: now}); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "保存任务审计失败"})
			return
		}
		queued, existingID, enqueueErr := uniqueQueue.EnqueueUniqueBooks(req.Context(), processJob, 24*time.Hour)
		if enqueueErr != nil {
			_ = updateWorkshopRun(store, processJob, "failed", enqueueErr.Error(), now, true)
			for _, bookID := range queuedBookIDs {
				document, getErr := store.GetDocument(req.Context(), identity.Username, bookID)
				if getErr == nil && document.Meta != nil {
					delete(document.Meta, "currentProcessJobId")
					_ = store.PutDocument(req.Context(), identity.Username, document)
				}
				_ = updateWorkshopTaskRuntimeState(req.Context(), store, identity.Username, bookID, "waiting_retry", enqueueErr.Error())
			}
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "Redis 任务队列不可用"})
			return
		}
		if !queued {
			_ = updateWorkshopRun(store, processJob, "duplicate_skipped", "同一书籍已有有效队列任务", now, true)
			if strings.TrimSpace(existingID) == "" {
				existingID = id
			}
			for _, bookID := range queuedBookIDs {
				_ = setNovelFetchBookQueueState(req.Context(), store, identity.Username, bookID, existingID, "", "waiting_retry")
				_ = updateWorkshopTaskRuntimeState(req.Context(), store, identity.Username, bookID, "waiting_retry", "同一书籍已有有效队列任务，已跳过本次开始")
			}
			jobView["id"] = existingID
			jobView["existing"] = true
			jobView["conflicting_book_ids"] = append(conflictBookIDs, queuedBookIDs...)
			registry.put(identity.Username, id, jobView)
			writeJSON(w, http.StatusAccepted, jobView)
			return
		}
		registry.put(identity.Username, id, jobView)
		writeJSON(w, http.StatusAccepted, jobView)
	})

	mux.HandleFunc("GET /api/novel-fetch-workshop/process/jobs/latest", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		limit, _ := strconv.Atoi(req.URL.Query().Get("limit"))
		jobs := registry.latest(identity.Username, limit)
		if len(jobs) == 0 {
			runs, err := store.ListRuns(req.Context(), identity.Username, "")
			if err == nil {
				for _, run := range runs {
					if run.Stage != "novel_fetch_process" {
						continue
					}
					jobs = append(jobs, persistedWorkshopRun(run))
					if limit > 0 && len(jobs) >= limit {
						break
					}
				}
			}
		}
		latest := map[string]any{}
		if len(jobs) > 0 {
			latest = jobs[0]
		}
		writeJSON(w, http.StatusOK, map[string]any{"latest": latest, "jobs": jobs})
	})

	mux.HandleFunc("GET /api/novel-fetch-workshop/process/jobs/{jobId}", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		job, ok := registry.get(identity.Username, req.PathValue("jobId"))
		if !ok {
			runs, err := store.ListRuns(req.Context(), identity.Username, "")
			for _, run := range runs {
				if run.RunID == req.PathValue("jobId") && run.Stage == "novel_fetch_process" {
					job = persistedWorkshopRun(run)
					ok = true
					break
				}
			}
			if !ok || err != nil {
				writeJSON(w, http.StatusNotFound, map[string]any{"error": "处理任务不存在"})
				return
			}
		}
		writeJSON(w, http.StatusOK, job)
	})

	mux.HandleFunc("GET /api/novel-fetch-workshop/runs", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		bookID := strings.TrimSpace(req.URL.Query().Get("bookId"))
		if bookID != "" && !novelFetchBookIDPattern.MatchString(bookID) {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "非法的书籍ID"})
			return
		}
		runs, err := store.ListRuns(req.Context(), identity.Username, bookID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "获取执行审计失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"runs": runs})
	})

	mux.HandleFunc("POST /api/novel-fetch-workshop/runs", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		var record novelfetchworkshop.RunRecord
		if err := decodeNovelFetchJSON(w, req, &record); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "请求内容无效"})
			return
		}
		if !novelFetchBookIDPattern.MatchString(strings.TrimSpace(record.BookID)) {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "非法的书籍ID"})
			return
		}
		if err := store.PutRun(req.Context(), identity.Username, record); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "执行审计字段无效"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})

	// The current Novel Fetch page reads its queue summary from the canonical
	// workshop route. Keep this route beside the rest of the canonical handlers
	// so a refresh does not fall through to the default 404 page.
	mux.HandleFunc("GET /api/novel-fetch-workshop/realtime/status", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		status, err := novelFetchAutomationStatus(req.Context(), store, registry, identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "读取任务队列失败"})
			return
		}
		status["queue_available"] = queue != nil
		writeJSON(w, http.StatusOK, status)
	})

	mux.HandleFunc("GET /api/novel-fetch-workshop/tasks", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		tasks, err := store.ListDocuments(req.Context(), identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "获取任务列表失败"})
			return
		}
		// 小说获取旧工作台仍使用 snake_case 字段。返回新字段的同时补齐
		// 兼容字段，避免页面把已经保存的元数据误显示为空。
		legacyTasks := make([]map[string]any, 0, len(tasks))
		for _, task := range tasks {
			if backfillNovelFetchClassification(&task) {
				if err := store.PutDocument(req.Context(), identity.Username, task); err != nil {
					writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "保存分类信息失败"})
					return
				}
			}
			syncNovelFetchClassificationMetadata(&task)
			legacyTasks = append(legacyTasks, legacyNovelFetchTask(task))
		}
		writeJSON(w, http.StatusOK, map[string]any{"tasks": legacyTasks})
	})

	// 批量重试由 Go 直接处理，避免小说获取工作台因为一个重试按钮又回到 Node。
	mux.HandleFunc("POST /api/novel-fetch-workshop/tasks/batch-retry", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		var input struct {
			IDs                []string `json:"ids"`
			SensitiveAIEnabled bool     `json:"sensitive_ai_enabled"`
		}
		if err := decodeNovelFetchJSON(w, req, &input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "请求内容无效"})
			return
		}
		ids := make([]string, 0, len(input.IDs))
		seen := map[string]bool{}
		for _, raw := range input.IDs {
			id := strings.TrimSpace(raw)
			if !novelFetchBookIDPattern.MatchString(id) {
				writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "非法的书籍ID"})
				return
			}
			if !seen[id] {
				seen[id] = true
				ids = append(ids, id)
			}
		}
		result, err := enqueueWorkshopRetries(req.Context(), store, queue, identity.Username, ids, input.SensitiveAIEnabled)
		if err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": err.Error(), "result": result})
			return
		}
		status := http.StatusOK
		if result.Queued > 0 || result.AlreadyRunning > 0 {
			status = http.StatusAccepted
		}
		writeJSON(w, status, result)
	})

	mux.HandleFunc("GET /api/novel-fetch-workshop/tasks/{bookId}/bodies", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		bookID, _, ok := novelFetchBodyPathValues(w, req)
		if !ok {
			return
		}
		bodies, err := store.ListBodyRefs(req.Context(), identity.Username, bookID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "读取正文列表失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"bodies": bodies})
	})

	mux.HandleFunc("GET /api/novel-fetch-workshop/tasks/{bookId}/bodies/{versionId}", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		bookID, versionID, ok := novelFetchBodyPathValues(w, req)
		if !ok {
			return
		}
		body, err := store.GetBody(req.Context(), identity.Username, bookID, versionID)
		if errors.Is(err, novelfetchworkshop.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "正文不存在"})
			return
		}
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "读取正文失败"})
			return
		}
		writeJSON(w, http.StatusOK, body)
	})

	mux.HandleFunc("PUT /api/novel-fetch-workshop/tasks/{bookId}/bodies/{versionId}", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		bookID, versionID, ok := novelFetchBodyPathValues(w, req)
		if !ok {
			return
		}
		var input struct {
			Content string `json:"content"`
			State   string `json:"state"`
		}
		if err := decodeNovelFetchJSON(w, req, &input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "请求内容无效"})
			return
		}
		ref, err := store.PutBody(req.Context(), identity.Username, novelfetchworkshop.BodyRecord{
			BookID:  bookID,
			BodyRef: novelfetchworkshop.BodyRef{VersionID: versionID, State: input.State},
			Content: input.Content,
		})
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "保存正文失败"})
			return
		}
		writeJSON(w, http.StatusOK, ref)
	})

	mux.HandleFunc("DELETE /api/novel-fetch-workshop/tasks/{bookId}/bodies/{versionId}", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		bookID, versionID, ok := novelFetchBodyPathValues(w, req)
		if !ok {
			return
		}
		deleted, err := store.DeleteBody(req.Context(), identity.Username, bookID, versionID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "删除正文失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"deleted": deleted})
	})

	registerNovelFetchSummary(mux, store)
	mux.HandleFunc("GET /api/novel-fetch-workshop/tasks/{bookId}", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		bookID := strings.TrimSpace(req.PathValue("bookId"))
		if !novelFetchBookIDPattern.MatchString(bookID) {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "非法的书籍ID"})
			return
		}
		document, err := store.GetDocument(req.Context(), identity.Username, bookID)
		if errors.Is(err, novelfetchworkshop.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "任务不存在"})
			return
		}
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "读取任务失败"})
			return
		}
		document, err = hydrateNovelFetchDocument(req.Context(), store, identity.Username, document)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "读取正文失败"})
			return
		}
		if backfillNovelFetchClassification(&document) {
			if err := store.PutDocument(req.Context(), identity.Username, document); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "保存分类信息失败"})
				return
			}
		}
		syncNovelFetchClassificationMetadata(&document)
		writeJSON(w, http.StatusOK, novelFetchTaskDetailPayload(document))
	})

	mux.HandleFunc("POST /api/novel-fetch-workshop/tasks/{bookId}/fetch", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		if fetcher == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "原文抓取服务未配置"})
			return
		}
		bookID := strings.TrimSpace(req.PathValue("bookId"))
		if !novelFetchBookIDPattern.MatchString(bookID) {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "非法的书籍ID"})
			return
		}
		document, err := store.GetDocument(req.Context(), identity.Username, bookID)
		if errors.Is(err, novelfetchworkshop.ErrNotFound) {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "任务不存在"})
			return
		}
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "读取任务失败"})
			return
		}
		var input struct {
			MaxTxt             int  `json:"maxTxt"`
			SensitiveAIEnabled bool `json:"sensitive_ai_enabled"`
		}
		if req.ContentLength != 0 {
			if err := decodeNovelFetchJSON(w, req, &input); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]any{"error": "请求内容无效"})
				return
			}
		}
		maxTxt := input.MaxTxt
		if maxTxt <= 0 {
			maxTxt = document.MaxTxt
		}
		if maxTxt <= 0 {
			maxTxt = 4000
		}
		platformID := "2"
		if value, ok := document.Meta["platformId"].(string); ok && strings.TrimSpace(value) != "" {
			platformID = strings.TrimSpace(value)
		}
		result, err := fetchNovelWithSettings(req.Context(), store, identity.Username, fetcher, bookID, platformID, maxTxt)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
			return
		}
		if err := stageNovelFetchRaw(req.Context(), store, identity.Username, &document, result); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "保存原文正文失败"})
			return
		}
		var sensitiveProvider novelfetchworkshop.RewriteProvider
		if input.SensitiveAIEnabled {
			sensitiveProvider, _ = resolveRewriteProvider(req)
		}
		processedText, rulesTrace, sensitiveItems, processErr := processWorkshopConfiguredText(req.Context(), store, identity.Username, bookID, result.ProcessedText, "original", input.SensitiveAIEnabled, sensitiveProvider)
		if processErr != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": processErr.Error()})
			return
		}
		sensitiveCount := len(sensitiveItems)
		document.ProcessedText = processedText
		document.Original = processedText
		document.ProcessedChars = len([]rune(processedText))
		if document.Meta == nil {
			document.Meta = map[string]any{}
		}
		document.Meta["originalStatus"] = "done"
		document.Meta["originalRawChars"] = result.OriginalRawChars
		document.Meta["processedChars"] = document.ProcessedChars
		document.Meta["sensitiveStatus"] = "done"
		document.Meta["sensitiveReplaced"] = sensitiveCount
		document.Meta["rulesTrace"] = rulesTrace
		document.Meta["rulesTraceByVersion"] = map[string]any{"original": rulesTrace}
		setSensitiveDetails(document.Meta, sensitiveItems, sensitiveCount)
		if err := persistNovelFetchBodies(req.Context(), store, identity.Username, &document, result.OriginalRaw, processedText); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "保存原文正文失败"})
			return
		}
		if genderNeedsResolution(document.Gender) {
			genderInfo := result.BookInfo
			genderInfo.Text = processedText
			decision := novelfetchworkshop.ResolveGender(req.Context(), "", genderInfo, resolveGenderJudge(req))
			document.Gender, document.GenderSource, document.GenderReason = decision.Gender, decision.Source, decision.Reason
			document.GenderConfidence = decision.Confidence
		}
		if styleNeedsResolution(document.Meta) {
			document.Meta["styleStatus"] = "processing"
			styleInfo := result.BookInfo
			styleInfo.Text = processedText
			provider, providerErr := resolveRewriteProvider(req)
			if providerErr != nil {
				recordNovelFetchStyleDecision(document.Meta, novelfetchworkshop.StyleDecision{}, providerErr)
			} else {
				style := novelfetchworkshop.ResolveStyle(req.Context(), styleInfo, provider)
				recordNovelFetchStyleDecision(document.Meta, style, nil)
			}
		}
		syncNovelFetchClassificationMetadata(&document)
		if err := store.PutDocument(req.Context(), identity.Username, document); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "保存原文失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"task": document})
	})

	mux.HandleFunc("POST /api/novel-fetch-workshop/tasks/{bookId}/restore-original", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		bookID := strings.TrimSpace(req.PathValue("bookId"))
		if !novelFetchBookIDPattern.MatchString(bookID) {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "非法的书籍ID"})
			return
		}
		document, err := store.GetDocument(req.Context(), identity.Username, bookID)
		if errors.Is(err, novelfetchworkshop.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "任务不存在"})
			return
		}
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "读取任务失败"})
			return
		}
		document, err = hydrateNovelFetchDocument(req.Context(), store, identity.Username, document)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "读取正文失败"})
			return
		}
		if strings.TrimSpace(document.OriginalRaw) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "没有可恢复的原始正文"})
			return
		}
		processed := novelfetchworkshop.ProcessRawText(document.OriginalRaw, document.MaxTxt)
		document.Original = processed.ProcessedText
		document.ProcessedText = processed.ProcessedText
		document.OriginalRawChars = processed.OriginalRawChars
		document.ProcessedChars = processed.ProcessedChars
		document.MaxTxt = processed.MaxTxt
		if document.Meta == nil {
			document.Meta = map[string]any{}
		}
		document.Meta["originalStatus"] = "restored"
		if err := persistNovelFetchBodies(req.Context(), store, identity.Username, &document, document.OriginalRaw, document.ProcessedText); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "保存恢复正文失败"})
			return
		}
		if err := store.PutDocument(req.Context(), identity.Username, document); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "保存恢复结果失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"task": document})
	})

	mux.HandleFunc("POST /api/novel-fetch-workshop/tasks/{bookId}/generate-ai", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		bookID := strings.TrimSpace(req.PathValue("bookId"))
		if !novelFetchBookIDPattern.MatchString(bookID) {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "非法的书籍ID"})
			return
		}
		var input struct {
			Count int `json:"count"`
		}
		if err := decodeNovelFetchJSON(w, req, &input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "请求内容无效"})
			return
		}
		if input.Count < 0 || input.Count > 20 {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "AI 数量须为 1 到 20"})
			return
		}
		if activeNovelFetchJobForBooks(req.Context(), store, identity.Username, []string{bookID}) != nil {
			writeJSON(w, http.StatusConflict, map[string]any{"ok": false, "error": "该书已有任务在运行，请完成后再生成；已保存的数量会在下次生成时生效"})
			return
		}
		existing, err := store.GetDocument(req.Context(), identity.Username, bookID)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]any{"ok": false, "error": "任务不存在或无法读取"})
			return
		}
		cfg, err := store.GetConfig(req.Context(), identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "读取 AI 数量配置失败"})
			return
		}
		versions := workshopMissingAIVersions(existing, workshopAIRequestedCount(existing, cfg, input.Count))
		if len(versions) == 0 {
			writeJSON(w, http.StatusOK, map[string]any{"ok": true, "task": existing, "message": "所需 AI 版本已齐，无需重复生成"})
			return
		}
		provider, providerErr := resolveRewriteProvider(req)
		if providerErr != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": providerErr.Error()})
			return
		}
		document, err := novelfetchworkshop.GenerateAIRewritesForVersionsProcessed(req.Context(), store, identity.Username, bookID, versions, provider, workshopRewriteProcessor(store, identity.Username, bookID, false, provider))
		if err != nil {
			status := http.StatusBadRequest
			if rewriteProvider == nil || strings.Contains(err.Error(), "模型未配置") {
				status = http.StatusServiceUnavailable
			}
			writeJSON(w, status, map[string]any{"ok": false, "error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "task": document})
	})

	mux.HandleFunc("POST /api/novel-fetch-workshop/ai/test", func(w http.ResponseWriter, req *http.Request) {
		var input struct {
			Purpose   string `json:"purpose"`
			Model     string `json:"model"`
			TextModel string `json:"textModel"`
		}
		if req.ContentLength != 0 {
			if err := decodeNovelFetchJSON(w, req, &input); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "请求内容无效"})
				return
			}
		}
		// 连通性测试必须使用当前选择的模型，否则测试通过但实际执行仍可能失败。
		identity, _ := BridgeIdentityFromContext(req.Context())
		ctx := novelfetchworkshop.WithProcessTextModel(req.Context(), firstText(input.Model, input.TextModel))
		provider, providerErr := resolveRewriteProviderFor(ctx, identity.Username)
		if providerErr != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": providerErr.Error()})
			return
		}
		if provider == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": "AI 改文模型未配置"})
			return
		}
		content, err := provider.Complete(ctx, novelfetchworkshop.RewriteRequest{
			SystemPrompt: "你是接口连通性测试助手。",
			UserPrompt:   "请只回复：ok",
			Temperature:  0,
			// 推理模型会先消耗思考 token；预算过小会在返回正文前被截断，
			// 导致连通性测试误报模型不可用。这里留出足够余量。
			MaxTokens: 2048,
		})
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
			return
		}
		content = strings.TrimSpace(content)
		if content == "" {
			writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "error": "文本模型返回为空，请检查模型或上游通道"})
			return
		}
		if runes := []rune(content); len(runes) > 200 {
			content = string(runes[:200])
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "content": content})
	})

	mux.HandleFunc("POST /api/novel-fetch-workshop/rules/suggest", func(w http.ResponseWriter, req *http.Request) {
		provider, providerErr := resolveRewriteProvider(req)
		if providerErr != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": providerErr.Error()})
			return
		}
		if provider == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": "AI 改文模型未配置"})
			return
		}
		var input struct {
			Text     string `json:"text"`
			RuleType string `json:"ruleType"`
			Goal     string `json:"goal"`
		}
		if err := decodeNovelFetchJSON(w, req, &input); err != nil || strings.TrimSpace(input.Text) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "请输入待分析文本"})
			return
		}
		content, err := novelfetchworkshop.SuggestRules(req.Context(), provider, input.RuleType, input.Goal, input.Text)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "suggestions": content})
	})

	mux.HandleFunc("POST /api/novel-fetch-workshop/rules/preview", func(w http.ResponseWriter, req *http.Request) {
		var input struct {
			Text         string         `json:"text"`
			Scope        string         `json:"scope"`
			LayoutConfig map[string]any `json:"layoutConfig"`
			Knowledge    map[string]any `json:"knowledge"`
		}
		if err := decodeNovelFetchJSON(w, req, &input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "请求内容无效"})
			return
		}
		if strings.TrimSpace(input.Text) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "请输入待预览文本"})
			return
		}
		identity, _ := BridgeIdentityFromContext(req.Context())
		settings, err := store.GetConfig(req.Context(), identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "读取预览规则失败"})
			return
		}
		// The page sends text/scope by default. Preview the same saved account
		// configuration used by processing, with optional unsaved preview overrides.
		app := mapOrEmpty(settings["appConfig"])
		if input.LayoutConfig != nil {
			app["layout"] = input.LayoutConfig
		}
		if input.Knowledge != nil {
			knowledge := mapOrEmpty(settings["knowledgeLibraries"])
			if len(knowledge) == 0 {
				knowledge = mapOrEmpty(app["knowledge"])
			}
			for kind, rules := range input.Knowledge {
				if patch, ok := rules.(map[string]any); ok {
					merged := cloneAnyMap(mapOrEmpty(knowledge[kind]))
					for key, value := range patch {
						merged[key] = value
					}
					knowledge[kind] = merged
				} else {
					knowledge[kind] = rules
				}
			}
			settings["knowledgeLibraries"] = knowledge
		}
		settings["appConfig"] = app
		result, trace, _ := novelfetchworkshop.ProcessConfiguredText(input.Text, input.Scope, settings)
		layout := mapOrEmpty(app["layout"])
		if len(layout) == 0 {
			layout = mapOrEmpty(settings["layout"])
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "text": result, "result": result, "trace": trace, "config": layout})
	})

	mux.HandleFunc("POST /api/novel-fetch-workshop/knowledge/repair-imported-sensitive", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		settings, err := store.GetConfig(req.Context(), identity.Username)
		if err != nil {
			writeJSON(w, 500, map[string]any{"error": "读取规则失败"})
			return
		}
		count, err := novelfetchworkshop.RepairImportedSensitive(settings)
		if err == nil && count > 0 {
			err = store.PutConfig(req.Context(), identity.Username, settings)
		}
		if err != nil {
			writeJSON(w, 409, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, 200, map[string]any{"ok": true, "repaired": count, "configRevision": settings[novelfetchworkshop.ConfigRevisionKey]})
	})
	mux.HandleFunc("POST /api/novel-fetch-workshop/knowledge/repair-imported", func(w http.ResponseWriter, req *http.Request) {
		var body struct {
			Apply          bool `json:"apply"`
			ConfigRevision any  `json:"configRevision"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, req.Body, 4096)).Decode(&body); err != nil {
			writeJSON(w, 400, map[string]any{"error": "修复参数无效"})
			return
		}
		identity, _ := BridgeIdentityFromContext(req.Context())
		settings, err := store.GetConfig(req.Context(), identity.Username)
		if err != nil {
			writeJSON(w, 500, map[string]any{"error": "读取资料失败"})
			return
		}
		if body.Apply && (body.ConfigRevision == nil || fmt.Sprint(body.ConfigRevision) != fmt.Sprint(settings[novelfetchworkshop.ConfigRevisionKey])) {
			writeJSON(w, 409, map[string]any{"error": "配置已变化，请重新检查后再修复"})
			return
		}
		report := novelfetchworkshop.RepairImportedKnowledge(settings, body.Apply)
		if body.Apply && report.Repaired > 0 {
			if err = store.PutConfig(req.Context(), identity.Username, settings); err != nil {
				writeJSON(w, 409, map[string]any{"error": "保存修复失败，请重新检查：" + err.Error()})
				return
			}
		}
		writeJSON(w, 200, map[string]any{"ok": true, "applied": body.Apply, "report": report, "configRevision": settings[novelfetchworkshop.ConfigRevisionKey]})
	})
	mux.HandleFunc("GET /api/novel-fetch-workshop/knowledge/summary", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		settings, err := store.GetConfig(req.Context(), identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "读取知识库失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "summary": novelfetchworkshop.KnowledgeSummary(settings)})
	})

	mux.HandleFunc("GET /api/novel-fetch-workshop/knowledge/{kind}", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		settings, err := store.GetConfig(req.Context(), identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "读取知识库失败"})
			return
		}
		library, err := novelfetchworkshop.ListKnowledge(settings, req.PathValue("kind"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "kind": req.PathValue("kind"), "data": library, "items": library["items"], "configRevision": settings[novelfetchworkshop.ConfigRevisionKey]})
	})

	mux.HandleFunc("POST /api/novel-fetch-workshop/knowledge/{kind}", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		var input map[string]any
		if err := decodeNovelFetchJSON(w, req, &input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "请求内容无效"})
			return
		}
		item, _ := input["item"].(map[string]any)
		if item == nil {
			item = input
		}
		settings, err := store.GetConfig(req.Context(), identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "读取知识库失败"})
			return
		}
		if !checkWorkshopConfigRevision(w, settings, input[novelfetchworkshop.ConfigRevisionKey]) {
			return
		}
		normalized, err := novelfetchworkshop.SaveKnowledge(settings, req.PathValue("kind"), item)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
			return
		}
		if err := store.PutConfig(req.Context(), identity.Username, settings); err != nil {
			writeWorkshopConfigSaveError(w, err, "保存知识库失败")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "item": normalized})
	})

	mux.HandleFunc("DELETE /api/novel-fetch-workshop/knowledge/{kind}/{id}", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		settings, err := store.GetConfig(req.Context(), identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "读取知识库失败"})
			return
		}
		if revision := req.URL.Query().Get("configRevision"); revision != "" && !checkWorkshopConfigRevision(w, settings, revision) {
			return
		}
		removed, err := novelfetchworkshop.DeleteKnowledge(settings, req.PathValue("kind"), req.PathValue("id"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
			return
		}
		if err := store.PutConfig(req.Context(), identity.Username, settings); err != nil {
			writeWorkshopConfigSaveError(w, err, "保存知识库失败")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "removed": removed})
	})

	mux.HandleFunc("POST /api/novel-fetch-workshop/knowledge/{kind}/{id}/optimize", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		input := map[string]any{}
		if req.Body != nil && req.ContentLength != 0 {
			if err := decodeNovelFetchJSON(w, req, &input); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "优化要求格式无效"})
				return
			}
		}
		provider, providerErr := resolveRewriteProvider(req)
		if providerErr != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": providerErr.Error()})
			return
		}
		settings, err := store.GetConfig(req.Context(), identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "读取知识库失败"})
			return
		}
		if !checkWorkshopConfigRevision(w, settings, input[novelfetchworkshop.ConfigRevisionKey]) {
			return
		}
		item, err := novelfetchworkshop.OptimizeKnowledge(req.Context(), settings, req.PathValue("kind"), req.PathValue("id"), provider, input)
		if err != nil {
			status := http.StatusBadRequest
			if provider == nil || strings.Contains(err.Error(), "模型未配置") {
				status = http.StatusServiceUnavailable
			}
			writeJSON(w, status, map[string]any{"ok": false, "error": err.Error()})
			return
		}
		if err := store.PutConfig(req.Context(), identity.Username, settings); err != nil {
			writeWorkshopConfigSaveError(w, err, "保存知识库失败")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "item": item})
	})

	mux.HandleFunc("POST /api/novel-fetch-workshop/opening/analyze", func(w http.ResponseWriter, req *http.Request) {
		var input struct {
			Text  string `json:"text"`
			Style string `json:"style"`
			Extra string `json:"extra"`
		}
		if err := decodeNovelFetchJSON(w, req, &input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "请求内容无效"})
			return
		}
		provider, providerErr := resolveRewriteProvider(req)
		if providerErr != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": providerErr.Error()})
			return
		}
		openingText := input.Text
		if style := strings.TrimSpace(input.Style); style != "" {
			openingText = "适用风格：" + style + "\n" + openingText
		}
		if extra := strings.TrimSpace(input.Extra); extra != "" {
			openingText = "补充要求：" + extra + "\n" + openingText
		}
		items, err := novelfetchworkshop.AnalyzeOpening(req.Context(), provider, openingText)
		if err != nil {
			status := http.StatusBadRequest
			if provider == nil {
				status = http.StatusServiceUnavailable
			}
			writeJSON(w, status, map[string]any{"ok": false, "error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "items": items})
	})

	mux.HandleFunc("POST /api/novel-fetch-workshop/opening/save", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		var input map[string]any
		if err := decodeNovelFetchJSON(w, req, &input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "请求内容无效"})
			return
		}
		item, _ := input["item"].(map[string]any)
		if item == nil {
			item = input
		}
		settings, err := store.GetConfig(req.Context(), identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "读取开头词库失败"})
			return
		}
		normalized, err := novelfetchworkshop.SaveKnowledge(settings, "opening_phrases", item)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
			return
		}
		if err := store.PutConfig(req.Context(), identity.Username, settings); err != nil {
			writeWorkshopConfigSaveError(w, err, "保存开头词库失败")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "item": normalized})
	})

	mux.HandleFunc("POST /api/novel-fetch-workshop/opening/normalize", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		settings, err := store.GetConfig(req.Context(), identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "读取开头词库失败"})
			return
		}
		library, _ := novelfetchworkshop.ListKnowledge(settings, "opening_phrases")
		items := make([]map[string]any, 0)
		seen := map[string]bool{}
		removed := 0
		for _, raw := range novelfetchworkshop.NormalizeKnowledgeItemsForHTTP(library["items"]) {
			text := strings.TrimSpace(fmt.Sprint(raw["text"]))
			if text == "" || seen[text] {
				removed++
				continue
			}
			seen[text] = true
			items = append(items, raw)
		}
		library["items"] = items
		libraries, _ := settings["knowledgeLibraries"].(map[string]any)
		if libraries == nil {
			libraries = map[string]any{}
		}
		libraries["opening_phrases"] = library
		settings["knowledgeLibraries"] = libraries
		novelfetchworkshop.SyncKnowledgeConfigViews(settings, map[string]any{"knowledgeLibraries": map[string]any{"opening_phrases": library}})
		if err := store.PutConfig(req.Context(), identity.Username, settings); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "保存规范化开头词失败"})
			return
		}
		grouped := map[string][]map[string]any{}
		for _, item := range items {
			style := strings.TrimSpace(fmt.Sprint(item["style"]))
			if style == "" {
				style = "未分类"
			}
			grouped[style] = append(grouped[style], item)
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "removed": removed, "grouped": grouped})
	})

	mux.HandleFunc("PUT /api/novel-fetch-workshop/tasks/{bookId}", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		bookID := strings.TrimSpace(req.PathValue("bookId"))
		if !novelFetchBookIDPattern.MatchString(bookID) {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "非法的书籍ID"})
			return
		}
		var document novelfetchworkshop.Document
		if err := decodeNovelFetchJSON(w, req, &document); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "请求内容无效"})
			return
		}
		explicitManualGender := !genderNeedsResolution(document.Gender)
		document.BookID = bookID
		existing, err := store.GetDocument(req.Context(), identity.Username, bookID)
		if err != nil && !errors.Is(err, novelfetchworkshop.ErrNotFound) {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "读取任务失败"})
			return
		}
		document = mergeNovelFetchDocument(existing, document)
		if explicitManualGender {
			genderInfo := document.BookInfo
			genderInfo.Text = document.ProcessedText
			decision := novelfetchworkshop.ResolveGender(req.Context(), document.Gender, genderInfo, resolveGenderJudge(req))
			document.Gender = decision.Gender
			document.GenderSource = decision.Source
			document.GenderReason = decision.Reason
			document.GenderConfidence = decision.Confidence
		} else if genderNeedsResolution(document.Gender) {
			// A sparse edit of an older task must still apply the deterministic
			// category mapping.  Pending values are not manual choices and must
			// not remain stuck just because the user edited another field.
			genderInfo := document.BookInfo
			genderInfo.Text = document.ProcessedText
			decision := novelfetchworkshop.ResolveGender(req.Context(), "", genderInfo, resolveGenderJudge(req))
			document.Gender = decision.Gender
			document.GenderSource = decision.Source
			document.GenderReason = decision.Reason
			document.GenderConfidence = decision.Confidence
		}
		if styleNeedsResolution(document.Meta) && strings.TrimSpace(document.ProcessedText) != "" {
			if document.Meta == nil {
				document.Meta = map[string]any{}
			}
			document.Meta["styleStatus"] = "processing"
			styleInfo := document.BookInfo
			styleInfo.Text = document.ProcessedText
			provider, providerErr := resolveRewriteProvider(req)
			if providerErr != nil {
				recordNovelFetchStyleDecision(document.Meta, novelfetchworkshop.StyleDecision{}, providerErr)
			} else {
				style := novelfetchworkshop.ResolveStyle(req.Context(), styleInfo, provider)
				recordNovelFetchStyleDecision(document.Meta, style, nil)
			}
		}
		syncNovelFetchClassificationMetadata(&document)
		if err := store.PutDocument(req.Context(), identity.Username, document); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "保存任务失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})

	mux.HandleFunc("DELETE /api/novel-fetch-workshop/tasks", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		var body struct {
			IDs []string `json:"ids"`
		}
		if err := decodeNovelFetchJSON(w, req, &body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "请求内容无效"})
			return
		}
		ids := make([]string, 0, len(body.IDs))
		for _, raw := range body.IDs {
			id := strings.TrimSpace(raw)
			if !novelFetchBookIDPattern.MatchString(id) {
				writeJSON(w, http.StatusBadRequest, map[string]any{"error": "非法的书籍ID"})
				return
			}
			ids = append(ids, id)
		}
		// 运行中的任务必须跳过删除，否则 worker 处理完会把文档重新写回，
		// 已删除任务会“复活”或覆盖删除后的新数据。
		skipped := make([]map[string]any, 0)
		deletable := make([]string, 0, len(ids))
		for _, id := range ids {
			if active := activeNovelFetchJobForBooksWithQueue(req.Context(), store, queue, identity.Username, []string{id}); active != nil {
				skipped = append(skipped, active)
				continue
			}
			deletable = append(deletable, id)
		}
		result, err := store.DeleteDocuments(req.Context(), identity.Username, deletable)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "删除任务失败"})
			return
		}
		response := map[string]any{
			"requested": len(ids),
			"deleted":   result.Deleted,
			"results":   result.Results,
			"skipped":   skipped,
		}
		writeJSON(w, http.StatusOK, response)
	})

	mux.HandleFunc("GET /api/novel-fetch-workshop/platforms", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		settings, err := store.GetConfig(req.Context(), identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "读取平台表失败"})
			return
		}
		platforms := make([]map[string]any, 0)
		if configured := novelFetchPlatformCatalog(settings); len(configured) > 0 {
			for _, raw := range configured {
				platform, ok := raw.(map[string]any)
				if !ok || platform["visible"] == false {
					continue
				}
				id, idOK := platform["id"].(string)
				name, nameOK := platform["name"].(string)
				if idOK && strings.TrimSpace(id) != "" && nameOK && strings.TrimSpace(name) != "" {
					platforms = append(platforms, map[string]any{"id": id, "name": name})
				}
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{"platforms": platforms})
	})

	mux.HandleFunc("GET /api/novel-fetch-workshop/config", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		settings, err := store.GetConfig(req.Context(), identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "读取设置失败"})
			return
		}
		// The migrated store can contain a compatibility wrapper under
		// `settings`.  The browser contract is the same flat config returned by
		// 3001, so unwrap it before selecting appConfig/platforms/aiConfig.
		publicSource := settings
		if nested, ok := settings["settings"].(map[string]any); ok {
			publicSource = nested
		}
		publicSource = novelfetchworkshop.EnsureWorkshopDefaults(publicSource)
		// 普通成员只读所属管理员的模型目录：成员自己保存的旧副本不再展示，
		// 避免“页面看到旧 Key、预检按管理员目录判断”的配置漂移。
		if canManage, _ := configManagementState(identity, profiles, accounts); !canManage {
			if owner := effectiveModelCatalogOwner(identity, profiles, accounts); owner != "" {
				if managed, err := store.GetConfig(req.Context(), owner); err == nil {
					if catalog, ok := managed["modelCatalog"].([]any); ok && len(catalog) > 0 {
						publicSource["modelCatalog"] = catalog
					}
				}
			}
		}
		appConfig := novelFetchAppConfig(publicSource)
		platforms := novelFetchPlatformCatalog(publicSource)
		publicSource["platforms"] = platforms
		styles, ok := publicSource["styles"].([]any)
		if !ok {
			styles = append([]any(nil), defaultNovelFetchStyles...)
		}
		aiConfig, _ := publicSource["aiConfig"].(map[string]any)
		if aiConfig == nil {
			if ai, ok := appConfig["ai"].(map[string]any); ok {
				aiConfig = map[string]any{"ai": ai}
			} else {
				aiConfig = map[string]any{}
			}
		}
		publicAppConfig := publicNovelFetchConfig(appConfig)
		publicAIConfig := publicNovelFetchConfig(aiConfig)
		publicSettings := publicNovelFetchConfig(publicSource)
		// The legacy embedded workbench reads the catalog and its counters from
		// the top-level response.  Keep that view in the same response as the
		// canonical appConfig so both pages consume one Go-owned configuration
		// instead of silently falling back to an empty client-side placeholder.
		knowledge := mapOrEmpty(appConfig["knowledge"])
		if len(knowledge) == 0 {
			knowledge = mapOrEmpty(appConfig["knowledgeLibraries"])
		}
		knowledge = publicNovelFetchConfig(knowledge)
		knowledgeSummary := batchKnowledgeSummary(knowledge, false)
		writeJSON(w, http.StatusOK, map[string]any{
			"appConfig":         publicAppConfig,
			"platforms":         platforms,
			"styles":            styles,
			"aiConfig":          publicAIConfig,
			"settings":          publicSettings,
			"knowledge":         knowledge,
			"knowledge_loaded":  true,
			"knowledge_summary": knowledgeSummary,
		})
	})

	saveWorkshopConfig := func(w http.ResponseWriter, req *http.Request) {
		stripCatalog := false
		if identity, ok := BridgeIdentityFromContext(req.Context()); ok {
			if canManage, _ := configManagementState(identity, profiles, accounts); !canManage {
				stripCatalog = true
			}
		}
		saveNovelFetchConfig(w, req, store, stripCatalog, func(merged map[string]any) map[string]any {
			return map[string]any{"ok": true, "settings": merged}
		})
	}
	// The current React workbench uses POST; keep PUT for the canonical API and
	// accept both methods during the Node-to-Go compatibility period.
	mux.HandleFunc("PUT /api/novel-fetch-workshop/config", saveWorkshopConfig)
	mux.HandleFunc("POST /api/novel-fetch-workshop/config", saveWorkshopConfig)
}

func novelFetchSubmitChunks(bookIDs []string, batchSize int) [][]string {
	if batchSize <= 0 {
		batchSize = 12
	}
	if batchSize > 1000 {
		batchSize = 1000
	}
	chunks := make([][]string, 0, (len(bookIDs)+batchSize-1)/batchSize)
	for start := 0; start < len(bookIDs); start += batchSize {
		end := min(start+batchSize, len(bookIDs))
		chunks = append(chunks, append([]string(nil), bookIDs[start:end]...))
	}
	return chunks
}

func saveNovelFetchConfig(w http.ResponseWriter, req *http.Request, store novelfetchworkshop.Store, stripModelCatalog bool, response func(map[string]any) map[string]any) {
	identity, ok := BridgeIdentityFromContext(req.Context())
	if !ok || identity.Username == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
		return
	}
	var patch map[string]any
	if err := decodeNovelFetchJSON(w, req, &patch); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "请求内容无效"})
		return
	}
	settings, _ := patch["settings"].(map[string]any)
	if settings == nil {
		settings = patch
	}
	// Both legacy flat forms and newer nested forms edit the same sections.
	// Copy only supplied fields into the nested patch so an old appConfig
	// cannot silently override a newly saved flat setting at execution time.
	settings = cloneAnyMap(settings)
	appPatch := cloneAnyMap(mapOrEmpty(settings["appConfig"]))
	for _, key := range []string{"rewrite", "layout", "workflow", "fetch", "sensitive_ai", "storage", "parser", "work_form", "web_submit", "text_model_id", "processing_rule_prompt", "processing_rule_scope"} {
		if value, supplied := settings[key]; supplied {
			if flat, ok := value.(map[string]any); ok {
				appPatch[key] = mergeWorkshopConfig(flat, mapOrEmpty(appPatch[key]))
			} else if _, explicit := appPatch[key]; !explicit {
				appPatch[key] = value
			}
		}
	}
	settings["appConfig"] = appPatch
	// 普通成员不能覆盖模型目录：目录统一由所属管理员维护，成员侧只读。
	if stripModelCatalog {
		delete(settings, "modelCatalog")
		delete(settings, "modelCatalogVersion")
	}
	current, err := store.GetConfig(req.Context(), identity.Username)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "读取设置失败"})
		return
	}
	merged := mergeWorkshopConfigPreservingSecrets(current, settings)
	if _, supplied := settings["platforms"]; supplied {
		merged["platforms"] = novelFetchPlatformCatalog(merged)
	}
	// The React workbench sends the switch as a flat field. Keep the canonical
	// nested setting in sync so manual, scheduled, and automatic runs agree.
	if raw, ok := settings["sensitive_ai_enabled"]; ok {
		if enabled, valid := raw.(bool); valid {
			ai := mapOrEmpty(merged["sensitive_ai"])
			ai["enabled"] = enabled
			merged["sensitive_ai"] = ai
		}
	}
	if raw, ok := settings["sensitiveAiEnabled"]; ok {
		if enabled, valid := raw.(bool); valid {
			ai := mapOrEmpty(merged["sensitive_ai"])
			ai["enabled"] = enabled
			merged["sensitive_ai"] = ai
		}
	}
	// Rule edits arrive under appConfig.knowledge while the knowledge APIs use
	// knowledgeLibraries. Persist both views together so replacements survive
	// a restart and are visible from either workbench.
	novelfetchworkshop.SyncKnowledgeConfigViews(merged, settings)
	novelfetchworkshop.EnsureSensitiveRuleReplacements(merged)
	if err := store.PutConfig(req.Context(), identity.Username, merged); err != nil {
		if errors.Is(err, novelfetchworkshop.ErrConfigConflict) {
			writeJSON(w, http.StatusConflict, map[string]any{"error": err.Error(), "code": "CONFIG_CONFLICT"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "保存设置失败"})
		return
	}
	writeJSON(w, http.StatusOK, response(merged))
}

func genderNeedsResolution(value string) bool {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "", "待确认", "待判定", "待分类", "未判定", "未知", "unknown", "pending":
		return true
	default:
		return false
	}
}

func novelFetchClassificationResolved(document novelfetchworkshop.Document) bool {
	return !genderNeedsResolution(document.Gender) && !styleNeedsResolution(document.Meta)
}

// canonicalNovelFetchAIVersion maps aliases such as ai01/ai001 to the
// canonical ai1..ai20 identifier. Anything outside that range is rejected so a
// single job cannot request an unbounded number of AI rewrites per book.
func canonicalNovelFetchAIVersion(value string) (string, bool) {
	name := strings.ToLower(strings.TrimSpace(value))
	if !strings.HasPrefix(name, "ai") {
		return "", false
	}
	index, err := strconv.Atoi(strings.TrimPrefix(name, "ai"))
	if err != nil || index < 1 || index > 20 {
		return "", false
	}
	return "ai" + strconv.Itoa(index), true
}

// normalizeNovelFetchAIVersions validates user supplied versions and collapses
// zero-padded aliases (ai01, ai001, ...) to the same canonical version.
func normalizeNovelFetchAIVersions(values []string) ([]string, error) {
	normalized := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, raw := range values {
		value := strings.TrimSpace(raw)
		if value == "" {
			continue
		}
		if strings.EqualFold(value, "original") {
			if !seen["original"] {
				seen["original"] = true
				normalized = append(normalized, "original")
			}
			continue
		}
		canonical, ok := canonicalNovelFetchAIVersion(value)
		if !ok {
			return nil, fmt.Errorf("不支持的AI版本：%s（仅支持 ai1 到 ai20）", value)
		}
		if seen[canonical] {
			continue
		}
		seen[canonical] = true
		normalized = append(normalized, canonical)
	}
	return normalized, nil
}

func selectedNovelFetchAIVersions(values []string) []string {
	selected := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, raw := range values {
		version, ok := canonicalNovelFetchAIVersion(raw)
		if !ok || seen[version] {
			continue
		}
		seen[version] = true
		selected = append(selected, version)
	}
	sort.Slice(selected, func(i, j int) bool {
		left, _ := strconv.Atoi(strings.TrimPrefix(selected[i], "ai"))
		right, _ := strconv.Atoi(strings.TrimPrefix(selected[j], "ai"))
		return left < right
	})
	return selected
}

func seenString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func persistedWorkshopRun(run novelfetchworkshop.RunRecord) map[string]any {
	job := map[string]any{
		"id": run.RunID, "status": run.Status, "queue_state": run.Status,
		"started_at": run.StartedAt, "updated_at": run.UpdatedAt,
		"attempts": run.Attempts, "steps": []any{}, "progress": []any{},
	}
	if run.ErrorMessage != "" {
		job["error"] = run.ErrorMessage
	}
	return job
}

// syncNovelFetchClassificationMetadata keeps the legacy flat metadata view in
// step with the canonical document fields. The old submit/workbench pages
// read meta.gender, while the workshop API reads document.gender.
func syncNovelFetchClassificationMetadata(document *novelfetchworkshop.Document) {
	if document == nil {
		return
	}
	if document.Meta == nil {
		document.Meta = map[string]any{}
	}
	if strings.TrimSpace(document.Gender) != "" && !genderNeedsResolution(document.Gender) {
		if document.GenderConfidence <= 0 && (document.GenderSource == novelfetchworkshop.GenderManual || document.GenderSource == novelfetchworkshop.GenderBookInfo) {
			document.GenderConfidence = 1
		}
		document.Meta["gender"] = document.Gender
		if document.GenderSource != "" {
			document.Meta["genderSource"] = string(document.GenderSource)
		}
		if strings.TrimSpace(document.GenderReason) != "" {
			document.Meta["genderReason"] = document.GenderReason
		}
		document.Meta["genderConfidence"] = document.GenderConfidence
	}
	if strings.TrimSpace(document.BookInfo.Category) != "" {
		document.Meta["category"] = document.BookInfo.Category
	}
	if strings.TrimSpace(document.BookInfo.Genre) != "" {
		document.Meta["genre"] = document.BookInfo.Genre
	}
	if title := strings.TrimSpace(document.BookInfo.WorkTitle); title != "" {
		document.Meta["bookName"] = title
	}
	// Platform metadata often contains a complete style label.  Resolve that
	// deterministic value here so old tasks do not stay at “待判定” and never
	// require a second AI call merely because the list page was opened.
	if styleNeedsResolution(document.Meta) {
		style := novelfetchworkshop.ResolveStyle(context.Background(), document.BookInfo, nil)
		if strings.TrimSpace(style.Style) != "" {
			recordNovelFetchStyleDecision(document.Meta, style, nil)
		}
	}
	inferNovelFetchStyleFailure(document.Meta)
	if style := strings.TrimSpace(firstText(document.Meta["style"], document.Meta["styleName"], document.Meta["style_name"])); style != "" && style != "待判定" && style != "未判定" {
		if strings.TrimSpace(fmt.Sprint(document.Meta["styleStatus"])) == "" {
			document.Meta["styleStatus"] = "classified"
		}
	}
}

// backfillNovelFetchClassification repairs tasks created by older flows.  The
// old worker stored the upstream book info but did not copy the deterministic
// platform/gender result into the document, leaving the list stuck at pending.
// This helper deliberately performs no AI call; AI fallback remains part of
// the fetch pipeline and is only needed when book metadata is ambiguous.
func backfillNovelFetchClassification(document *novelfetchworkshop.Document) bool {
	if document == nil {
		return false
	}
	changed := false
	if document.Meta == nil {
		document.Meta = map[string]any{}
	}
	if strings.TrimSpace(fmt.Sprint(document.Meta["platformName"])) == "" {
		platformID := strings.TrimSpace(fmt.Sprint(document.Meta["platformId"]))
		if number, err := strconv.Atoi(platformID); err == nil {
			if platformName := legacyNovelFetchPlatforms[number]; platformName != "" {
				document.Meta["platformName"] = platformName
				changed = true
			}
		}
	}
	if genderNeedsResolution(document.Gender) {
		decision := novelfetchworkshop.ResolveGender(context.Background(), "", document.BookInfo, nil)
		if !genderNeedsResolution(decision.Gender) {
			document.Gender = decision.Gender
			document.GenderSource = decision.Source
			document.GenderReason = decision.Reason
			document.GenderConfidence = decision.Confidence
			changed = true
		}
	}
	return changed
}

func applyNovelFetchSensitiveRules(ctx context.Context, store novelfetchworkshop.Store, owner, text string) (string, int, error) {
	processed, replaced, _, err := applyNovelFetchSensitiveRulesDetailed(ctx, store, owner, text)
	return processed, replaced, err
}

func applyNovelFetchSensitiveRulesDetailed(ctx context.Context, store novelfetchworkshop.Store, owner, text string) (string, int, []map[string]any, error) {
	if store == nil {
		return text, 0, []map[string]any{}, nil
	}
	settings, err := store.GetConfig(ctx, owner)
	if err != nil {
		return text, 0, []map[string]any{}, err
	}
	processed, replaced, items := novelfetchworkshop.ApplySensitiveRulesWithDetails(text, settings)
	return processed, replaced, items, nil
}

func setSensitiveDetails(meta map[string]any, items []map[string]any, count int) {
	if meta == nil {
		return
	}
	updated := time.Now().UTC().Format(time.RFC3339Nano)
	hits := make([]any, 0, len(items))
	fixed := make([]any, 0, len(items))
	failedCount := 0
	mode := "system_rule"
	for _, item := range items {
		itemSource := strings.ToLower(strings.TrimSpace(fmt.Sprint(item["source"])))
		if itemSource == "ai_group" || itemSource == "group" {
			mode = "ai_group"
		} else if itemSource == "ai" || itemSource == "ai_each" {
			mode = "ai_each"
		}
		hits = append(hits, map[string]any{
			"hit_index": item["hit_index"], "keyword": item["keyword"],
			"start": item["start"], "end": item["end"],
			"snippet": item["before"], "status": item["status"],
		})
		fixed = append(fixed, item)
		if strings.EqualFold(strings.TrimSpace(fmt.Sprint(item["status"])), "failed") {
			failedCount++
		}
	}
	fixedCount := count - failedCount
	if fixedCount < 0 {
		fixedCount = 0
	}
	meta["sensitiveHits"] = map[string]any{"mode": mode, "status": "done", "hit_count": count, "updated_at": updated, "hits": hits}
	status := "done"
	if failedCount > 0 {
		status = "failed"
	}
	meta["sensitiveFixed"] = map[string]any{"mode": mode, "status": status, "fixed_count": fixedCount, "failed_count": failedCount, "updated_at": updated, "items": fixed}
}

// novelFetchAppConfig restores the flat config contract used by the 3001
// workbench.  MySQL rows created by the migration keep the real fetch/rewrite
// sections at the root while appConfig contains only the newer UI fields.
// Merge both maps instead of returning the small compatibility fragment.
func novelFetchAppConfig(source map[string]any) map[string]any {
	base := map[string]any{}
	for _, key := range []string{
		"workflow", "fetch", "ai", "ai_presets", "ai_assignments", "rewrite",
		"storage", "parser", "knowledge", "knowledgeLibraries", "modelCatalog",
		"modelCatalogVersion", "sensitive_ai", "text_model_id", "web_submit", "work_form",
		"processing_rule_prompt", "processing_rule_scope",
	} {
		if value, ok := source[key]; ok {
			base[key] = value
		}
	}
	if configured, ok := source["appConfig"].(map[string]any); ok {
		base = mergeWorkshopConfig(base, configured)
	}
	return base
}

// publicNovelFetchConfig removes credentials before configuration reaches the browser.
// Empty api_key values are intentionally retained by the save path so a user can
// edit other settings without accidentally clearing the stored key.
func publicNovelFetchConfig(input map[string]any) map[string]any {
	out := map[string]any{}
	for key, value := range input {
		switch key {
		case "importedSensitiveRepairBackup", "importedKnowledgeRepairBackup":
			continue
		case "credential", "api_key", "password", "secret", "token":
			if key == "api_key" || key == "password" {
				out[key+"_configured"] = strings.TrimSpace(fmt.Sprint(value)) != "" && fmt.Sprint(value) != "<nil>"
			}
			continue
		}
		if nested, ok := value.(map[string]any); ok {
			out[key] = publicNovelFetchConfig(nested)
			continue
		}
		if list, ok := value.([]any); ok {
			items := make([]any, 0, len(list))
			for _, item := range list {
				if nested, ok := item.(map[string]any); ok {
					items = append(items, publicNovelFetchConfig(nested))
				} else {
					items = append(items, item)
				}
			}
			out[key] = items
			continue
		}
		out[key] = value
	}
	return out
}

func novelFetchTaskDetailPayload(document novelfetchworkshop.Document) map[string]any {
	raw, _ := json.Marshal(document)
	payload := map[string]any{}
	_ = json.Unmarshal(raw, &payload)
	status := novelfetchworkshop.DeriveTaskStatus(document)
	meta := mapOrEmpty(payload["meta"])
	meta["status"] = status
	meta["effectiveStatus"] = status
	meta["effective_status"] = status
	payload["meta"] = meta
	payload["status"] = status
	payload["effectiveStatus"] = status
	payload["effective_status"] = status
	return payload
}

func legacyNovelFetchTask(document novelfetchworkshop.Document) map[string]any {
	raw, _ := json.Marshal(document)
	result := map[string]any{}
	_ = json.Unmarshal(raw, &result)
	meta := document.Meta
	metaString := func(keys ...string) string {
		for _, key := range keys {
			if value, ok := meta[key]; ok {
				if text := strings.TrimSpace(fmt.Sprint(value)); text != "" && text != "<nil>" {
					return text
				}
			}
		}
		return ""
	}
	bookName := firstText(metaString("bookName", "book_name"), document.BookInfo.WorkTitle)
	platformName := metaString("platformName", "platform_name")
	if platformName == "" {
		if platformID, err := strconv.Atoi(metaString("platformId", "platform_id")); err == nil {
			platformName = legacyNovelFetchPlatforms[platformID]
		}
	}
	style := metaString("style", "styleName", "style_name")
	gender := strings.TrimSpace(document.Gender)
	if genderNeedsResolution(gender) {
		if decision := novelfetchworkshop.ResolveGender(context.Background(), "", document.BookInfo, nil); !genderNeedsResolution(decision.Gender) {
			gender = decision.Gender
		}
	}
	if gender == "" || genderNeedsResolution(gender) {
		gender = metaString("gender")
	}
	result["id"] = document.BookID
	result["book_id"] = document.BookID
	result["book_name"] = bookName
	result["platform_name"] = platformName
	result["style"] = style
	result["gender"] = gender
	result["classify_status"] = "classify_failed"
	if gender != "" && !genderNeedsResolution(gender) {
		result["classify_status"] = "classified"
	}
	result["original_status"] = metaString("originalStatus", "original_status")
	result["original_fetch_retry_count"] = intValue(meta["originalFetchRetryAttempts"])
	result["original_fetch_next_retry_at"] = metaString("originalFetchNextRetryAt", "original_fetch_next_retry_at")
	result["original_fetch_last_error"] = metaString("originalFetchLastError", "original_fetch_last_error")
	if result["original_status"] == "" && document.OriginalRawChars > 0 {
		result["original_status"] = "done"
	}
	result["original_raw_chars"] = document.OriginalRawChars
	result["processed_chars"] = document.ProcessedChars
	result["max_txt"] = document.MaxTxt
	result["ai_status"] = metaString("aiStatus", "ai_status")
	if result["ai_status"] == "" {
		result["ai_status"] = "waiting_ai_config"
	}
	generated := make([]string, 0, len(document.Versions))
	for version, value := range document.Versions {
		if !strings.HasPrefix(version, "ai") {
			continue
		}
		if state, ok := value.(map[string]any); ok && strings.EqualFold(strings.TrimSpace(fmt.Sprint(state["state"])), "ready") {
			generated = append(generated, version)
			continue
		}
		// GenerateAIRewrites stores the body text in Versions and the durable
		// readiness state in BodyRefs. Reading only the old map-shaped state made
		// completed AI versions disappear from the V78 task table after refresh.
		if _, ok := document.BodyRefs[version]; ok && strings.EqualFold(strings.TrimSpace(document.BodyRefs[version].State), "ready") {
			generated = append(generated, version)
		}
	}
	for version, body := range document.BodyRefs {
		if !strings.HasPrefix(version, "ai") || seenString(generated, version) || !strings.EqualFold(strings.TrimSpace(body.State), "ready") {
			continue
		}
		generated = append(generated, version)
	}
	sort.Strings(generated)
	if len(generated) > 0 {
		result["ai_files"] = generated
		result["selected_versions"] = generated
		result["ai_status"] = "ai_done"
	}
	result["site_submit_status"] = metaString("siteSubmitStatus", "site_submit_status")
	for _, field := range []struct {
		output string
		keys   []string
	}{
		{"site_submit_done_versions", []string{"siteSubmitDoneVersions", "site_submit_done_versions"}},
		{"site_submit_failed_versions", []string{"siteSubmitFailedVersions", "site_submit_failed_versions"}},
		{"site_submit_partial_versions", []string{"siteSubmitPartialVersions", "site_submit_partial_versions"}},
		{"site_submit_extra_done_versions", []string{"siteSubmitExtraDoneVersions", "site_submit_extra_done_versions"}},
		{"site_submit_pending_versions", []string{"siteSubmitPendingVersions", "site_submit_pending_versions"}},
		{"site_submit_waiting_generate_versions", []string{"siteSubmitWaitingGenerateVersions", "site_submit_waiting_generate_versions"}},
		{"site_submit_ready_versions", []string{"siteSubmitReadyVersions", "site_submit_ready_versions"}},
		{"site_submit_expected_versions", []string{"siteSubmitExpectedVersions", "site_submit_expected_versions"}},
	} {
		for _, key := range field.keys {
			if value, ok := meta[key]; ok {
				result[field.output] = value
				break
			}
		}
	}
	result["status"] = novelfetchworkshop.DeriveTaskStatus(document)
	// effective_status 是后端唯一有效状态；status 为兼容旧页面保留同值。
	result["effective_status"] = result["status"]
	result["effectiveStatus"] = result["status"]
	result["queue_state"] = metaString("queueState", "queue_state")
	createdAt := metaString("createdAt", "created_at", "createdTime", "created_time")
	if createdAt == "" {
		createdAt = document.UpdatedAt
	}
	result["created_at"] = createdAt
	result["createdAt"] = createdAt
	return result
}

func mergeWorkshopConfigPreservingSecrets(current, patch map[string]any) map[string]any {
	merged := mergeWorkshopConfig(current, patch)
	currentAI, _ := current["aiConfig"].(map[string]any)
	patchAI, _ := patch["aiConfig"].(map[string]any)
	if currentAI != nil && patchAI != nil {
		if strings.TrimSpace(fmt.Sprint(patchAI["api_key"])) == "" {
			if key := strings.TrimSpace(fmt.Sprint(currentAI["api_key"])); key != "" && key != "<nil>" {
				if ai, ok := merged["aiConfig"].(map[string]any); ok {
					ai["api_key"] = key
				}
			}
		}
	}
	currentApp, _ := current["appConfig"].(map[string]any)
	patchApp, _ := patch["appConfig"].(map[string]any)
	if currentApp != nil && patchApp != nil {
		currentAI, _ := currentApp["ai"].(map[string]any)
		patchAI, _ := patchApp["ai"].(map[string]any)
		if currentAI != nil && patchAI != nil && strings.TrimSpace(fmt.Sprint(patchAI["api_key"])) == "" {
			if key := strings.TrimSpace(fmt.Sprint(currentAI["api_key"])); key != "" && key != "<nil>" {
				if app, ok := merged["appConfig"].(map[string]any); ok {
					if ai, ok := app["ai"].(map[string]any); ok {
						ai["api_key"] = key
					}
				}
			}
		}
	}
	return merged
}

var workshopMultiSpacePattern = regexp.MustCompile(`\s{2,}`)

// 与参考实现的列预设保持一致；旧版页面传预设 ID，新版页面传展示名。
var workshopColumnPresetColumns = map[string][]string{
	"paid_name_reason":             {"paid_book_id", "book_name", "reason"},
	"paid_name_gender_reason":      {"paid_book_id", "book_name", "gender", "reason"},
	"free_paid_name_gender_reason": {"free_book_id", "paid_book_id", "book_name", "gender", "reason"},
	"sample_input":                 {"book_id", "book_name", "reason", "gender", "tags", "rating"},
	"full_11":                      {"ignore", "free_book_id", "paid_book_id", "book_name", "rating", "ignore", "ignore", "ignore", "ignore", "gender", "reason"},
}

var workshopColumnPresetAliases = map[string]string{
	"付费id/书名/推荐理由":           "paid_name_reason",
	"付费id/书名/男女频/推荐理由":       "paid_name_gender_reason",
	"免费id/付费id/书名/男女频/推荐理由":  "free_paid_name_gender_reason",
	"书籍id/书名/推荐理由/男女频/标签/评级": "sample_input",
	"完整11列": "full_11",
}

var workshopFixedColumnOrders = map[string][]string{
	"fixed_full_11":    workshopColumnPresetColumns["full_11"],
	"fixed_from_b":     workshopColumnPresetColumns["full_11"][1:],
	"fixed_paid_basic": {"paid_book_id", "book_name", "gender", "reason"},
}

var workshopColumnAliases = map[string]string{}

func init() {
	for alias, field := range map[string]string{
		"id": "book_id", "bookid": "book_id", "书籍id": "book_id", "小说id": "book_id", "作品id": "book_id",
		"书号": "book_id", "书籍编号": "book_id", "小说编号": "book_id", "作品编号": "book_id",
		"freebookid": "free_book_id", "免费id": "free_book_id", "免费书籍id": "free_book_id",
		"paidbookid": "paid_book_id", "付费id": "paid_book_id", "付费书籍id": "paid_book_id",
		"bookname": "book_name", "书名": "book_name", "作品名称": "book_name", "小说名称": "book_name",
		"书籍名称": "book_name", "名称": "book_name", "title": "book_name",
		"gender": "gender", "男女频": "gender", "性别": "gender", "频道": "gender", "频道类型": "gender",
		"reason": "reason", "推荐理由": "reason", "推荐原因": "reason", "入选理由": "reason", "理由": "reason", "简介": "reason",
		"tags": "tags", "标签": "tags", "rating": "rating", "评级": "rating",
		"ignore": "ignore", "忽略": "ignore", "跳过": "ignore", "空": "ignore",
	} {
		workshopColumnAliases[alias] = field
	}
}

func workshopNormalizeColumnToken(token string) string {
	normalized := strings.ToLower(strings.TrimSpace(token))
	normalized = strings.NewReplacer(" ", "", "\t", "", "_", "", "：", "", ":", "", "*", "", "（", "", "）", "", "(", "", ")", "", "【", "", "】", "", "[", "", "]", "").Replace(normalized)
	if field, ok := workshopColumnAliases[normalized]; ok {
		return field
	}
	return ""
}

func workshopColumnOrderFromText(text string) []string {
	order := make([]string, 0)
	for _, token := range strings.FieldsFunc(text, func(r rune) bool {
		return r == '\n' || r == '\r' || r == '\t' || r == ',' || r == '，' || r == '、' || r == '/' || r == '|'
	}) {
		if field := workshopNormalizeColumnToken(token); field != "" {
			order = append(order, field)
		}
	}
	return order
}

func workshopColumnOrderFromValues(values []string) []string {
	order := make([]string, 0, len(values))
	for _, value := range values {
		if field := workshopNormalizeColumnToken(value); field != "" {
			order = append(order, field)
		}
	}
	return order
}

// workshopColumnOrderRaw 兼容新版数组和旧版逗号/顿号分隔字符串。
// 返回值保留原始列名用于回显，实际解析时再由 workshopResolveColumnOrder 映射。
func workshopColumnOrderRaw(raw json.RawMessage) []string {
	if len(raw) == 0 {
		return nil
	}
	var values []string
	if err := json.Unmarshal(raw, &values); err == nil {
		return cleanStrings(values)
	}
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		order := make([]string, 0)
		for _, token := range strings.FieldsFunc(text, func(r rune) bool {
			return r == '\n' || r == '\r' || r == '\t' || r == ',' || r == '，' || r == '、' || r == '/' || r == '|'
		}) {
			if token = strings.TrimSpace(token); token != "" {
				order = append(order, token)
			}
		}
		return order
	}
	return nil
}

func workshopResolveColumnOrder(metadata workshopInputMetadata) []string {
	format := strings.ToLower(strings.TrimSpace(metadata.Format))
	if fixed, ok := workshopFixedColumnOrders[format]; ok {
		return append([]string(nil), fixed...)
	}
	preset := strings.ToLower(strings.TrimSpace(metadata.ColumnPreset))
	if alias, ok := workshopColumnPresetAliases[preset]; ok {
		preset = alias
	}
	if columns, ok := workshopColumnPresetColumns[preset]; ok {
		return append([]string(nil), columns...)
	}
	if format == "custom" && len(metadata.ColumnOrder) > 0 {
		return workshopColumnOrderFromValues(metadata.ColumnOrder)
	}
	if len(metadata.ColumnOrder) > 0 {
		return workshopColumnOrderFromValues(metadata.ColumnOrder)
	}
	return nil
}

func workshopSplitLine(line string) []string {
	split := func(parts []string) []string {
		values := make([]string, 0, len(parts))
		for _, part := range parts {
			values = append(values, strings.TrimSpace(part))
		}
		return values
	}
	if strings.Contains(line, "\t") {
		return split(strings.Split(line, "\t"))
	}
	if strings.Contains(line, "|") {
		return split(strings.Split(line, "|"))
	}
	if workshopMultiSpacePattern.MatchString(line) {
		return split(workshopMultiSpacePattern.Split(line, -1))
	}
	if strings.Contains(line, ",") {
		return split(strings.Split(line, ","))
	}
	return []string{strings.TrimSpace(line)}
}

func workshopHeaderRowScore(cells []string) int {
	fields := map[string]struct{}{}
	for _, cell := range cells {
		if field := workshopNormalizeColumnToken(cell); field != "" && field != "ignore" {
			fields[field] = struct{}{}
		}
	}
	return len(fields)
}

// workshopDetectHeader 返回合并后的列顺序和需要跳过的表头行数。
func workshopDetectHeader(rows [][]string, maxScan int) ([]string, int) {
	firstHeader := -1
	for index := 0; index < len(rows) && index < maxScan; index++ {
		if workshopHeaderRowScore(rows[index]) >= 2 {
			firstHeader = index
			break
		}
	}
	if firstHeader < 0 {
		return nil, 0
	}
	lastHeader := firstHeader
	limit := len(rows)
	if limit > maxScan {
		limit = maxScan
	}
	for lastHeader+1 < limit && workshopHeaderRowScore(rows[lastHeader+1]) >= 2 {
		lastHeader++
	}
	width := 0
	for _, row := range rows[:lastHeader+1] {
		if len(row) > width {
			width = len(row)
		}
	}
	order := make([]string, width)
	fields := 0
	for _, row := range rows[:lastHeader+1] {
		for index, cell := range row {
			if field := workshopNormalizeColumnToken(cell); field != "" {
				order[index] = field
			}
		}
	}
	for _, field := range order {
		if field != "" {
			fields++
		}
	}
	if fields < 2 {
		return nil, 0
	}
	return order, lastHeader + 1
}

func workshopGenderValue(value string) string {
	if strings.Contains(value, "男频") {
		return "男频"
	}
	if strings.Contains(value, "女频") {
		return "女频"
	}
	return ""
}

func workshopRowByColumnOrder(cells, order []string) (bookID, bookName, gender string) {
	values := map[string]string{}
	for index, field := range order {
		if index >= len(cells) {
			break
		}
		cell := strings.TrimSpace(cells[index])
		if field == "" || field == "ignore" {
			continue
		}
		if field == "reason" || field == "tags" {
			// 最后一列经常包含被分隔符切开的推荐理由，重新合并剩余单元格。
			if index == len(order)-1 {
				parts := make([]string, 0, len(cells)-index)
				for _, item := range cells[index:] {
					if strings.TrimSpace(item) != "" {
						parts = append(parts, strings.TrimSpace(item))
					}
				}
				values["reason"] = strings.Join(parts, " ")
				continue
			}
			if values["reason"] != "" {
				values["reason"] += " " + cell
			} else {
				values["reason"] = cell
			}
			continue
		}
		if values[field] == "" {
			values[field] = cell
		}
	}
	bookID = novelFetchSafeBookID(firstText(values["paid_book_id"], values["book_id"], values["free_book_id"]))
	bookName = values["book_name"]
	gender = workshopGenderValue(values["gender"])
	return bookID, bookName, gender
}

func parseWorkshopDocuments(input string, maxTxt int, options ...workshopInputMetadata) ([]novelfetchworkshop.Document, int, int, int) {
	metadata := workshopInputMetadata{}
	if len(options) > 0 {
		metadata = options[0]
	}
	format := strings.ToLower(strings.TrimSpace(metadata.Format))
	lines := strings.Split(strings.ReplaceAll(input, "\r\n", "\n"), "\n")
	rawRows := make([][]string, 0, len(lines))
	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			continue
		}
		rawRows = append(rawRows, workshopSplitLine(line))
	}
	headerOrder := []string(nil)
	headerRows := 0
	// 新前端的输入已经序列化为固定四列，不能再按原始表头/列顺序重排。
	if !metadata.Normalized && (format == "" || format == "smart" || format == "header" || format == "multi_header") {
		maxScan := 5
		if format == "multi_header" {
			maxScan = 8
		}
		headerOrder, headerRows = workshopDetectHeader(rawRows, maxScan)
	}
	order := headerOrder
	if len(order) == 0 {
		order = workshopResolveColumnOrder(metadata)
	}
	if metadata.Normalized {
		// 新版页面已按书籍ID/书名/男女频/理由序列化，忽略展示用的列顺序。
		order = []string{"book_id", "book_name", "gender", "reason"}
	}
	documents := make([]novelfetchworkshop.Document, 0)
	seen := make(map[string]struct{})
	parsed, emptyIDs, duplicates := 0, 0, 0
	for index, cells := range rawRows {
		if index < headerRows {
			continue
		}
		line := strings.Join(cells, "\t")
		// 表头需要跳过，但带 book_id=123456 的链接本身不是表头，不能一起丢掉。
		if workshopHeaderRowScore(cells) >= 2 && novelFetchBookIDURLPattern.FindString(line) == "" {
			continue
		}
		parsed++
		bookID, bookName, gender := "", "", ""
		if len(order) > 0 {
			bookID, bookName, gender = workshopRowByColumnOrder(cells, order)
		} else if format == "bookid_only" {
			for _, cell := range cells {
				if candidate := novelFetchSafeBookID(cell); candidate != "" {
					bookID = candidate
					break
				}
			}
		}
		idFromURL := novelFetchBookIDURLPattern.MatchString(line)
		if bookID == "" && !idFromURL {
			for _, cell := range cells {
				if candidate := novelFetchSafeBookID(cell); candidate != "" {
					bookID = candidate
					break
				}
			}
		}
		if bookID == "" {
			emptyIDs++
			continue
		}
		if _, exists := seen[bookID]; exists {
			duplicates++
			continue
		}
		seen[bookID] = struct{}{}
		if bookName == "" {
			// 常用输入是“ID + 若干空格 + 书名”。先从原始行移除 ID 和
			// 紧随其后的空白/分隔符，避免把 ID 一并保存成书名。链接输入不参与
			// 书名推断，否则查询参数会被误存成书名。
			if idIndex := strings.Index(line, bookID); idIndex >= 0 && !idFromURL {
				remainder := strings.TrimLeftFunc(line[idIndex+len(bookID):], func(r rune) bool {
					return unicode.IsSpace(r) || r == '|' || r == ','
				})
				if remainder != "" {
					nameParts := strings.FieldsFunc(remainder, func(r rune) bool { return r == '\t' || r == '|' || r == ',' })
					if len(nameParts) > 0 {
						candidate := strings.TrimSpace(nameParts[0])
						if !strings.Contains(candidate, "男频") && !strings.Contains(candidate, "女频") {
							bookName = candidate
						}
					}
				}
			}
		}
		if gender == "" {
			for _, cell := range cells {
				if strings.Contains(cell, bookID) {
					continue
				}
				if value := workshopGenderValue(cell); value != "" {
					gender = value
					break
				}
			}
		}
		if bookName == "" && len(order) == 0 {
			for _, cell := range cells {
				if strings.Contains(cell, bookID) || strings.Contains(cell, "男频") || strings.Contains(cell, "女频") {
					continue
				}
				bookName = cell
				break
			}
		}
		// The V78 preview hides the textarea after parsing and later rebuilds
		// input_text from each selected task's sourceLine. Preserve the original
		// line here; without it the second request starts with an empty list.
		meta := map[string]any{"bookId": bookID, "bookName": bookName, "maxTxt": maxTxt, "sourceLine": line, "createdAt": time.Now().UTC().Format(time.RFC3339Nano)}
		if gender != "" {
			meta["gender"] = gender
			meta["genderSource"] = "input"
		}
		document := novelfetchworkshop.Document{BookID: bookID, Gender: gender, MaxTxt: maxTxt, Meta: meta}
		if gender != "" {
			document.GenderSource = novelfetchworkshop.GenderManual
			document.GenderReason = "采用输入清单中的男女频"
		}
		documents = append(documents, document)
	}
	return documents, parsed, emptyIDs, duplicates
}

func mergeWorkshopConfig(current, patch map[string]any) map[string]any {
	merged := map[string]any{}
	for key, value := range current {
		merged[key] = value
	}
	for key, value := range patch {
		currentMap, currentOK := merged[key].(map[string]any)
		patchMap, patchOK := value.(map[string]any)
		if currentOK && patchOK {
			merged[key] = mergeWorkshopConfig(currentMap, patchMap)
			continue
		}
		merged[key] = value
	}
	return merged
}

func mergeNovelFetchDocument(existing, incoming novelfetchworkshop.Document) novelfetchworkshop.Document {
	if strings.TrimSpace(incoming.Gender) == "" {
		incoming.Gender = existing.Gender
	}
	if incoming.GenderSource == "" {
		incoming.GenderSource = existing.GenderSource
	}
	if strings.TrimSpace(incoming.GenderReason) == "" {
		incoming.GenderReason = existing.GenderReason
	}
	if strings.TrimSpace(incoming.BookInfo.Category) == "" {
		incoming.BookInfo.Category = existing.BookInfo.Category
	}
	if strings.TrimSpace(incoming.BookInfo.Genre) == "" {
		incoming.BookInfo.Genre = existing.BookInfo.Genre
	}
	if incoming.Meta == nil {
		incoming.Meta = existing.Meta
	}
	if incoming.BodyRefs == nil {
		incoming.BodyRefs = existing.BodyRefs
	}
	if strings.TrimSpace(incoming.Original) == "" {
		incoming.Original = existing.Original
	}
	if strings.TrimSpace(incoming.OriginalRaw) == "" {
		incoming.OriginalRaw = existing.OriginalRaw
	}
	if strings.TrimSpace(incoming.ProcessedText) == "" {
		incoming.ProcessedText = existing.ProcessedText
	}
	if incoming.OriginalRawChars == 0 {
		incoming.OriginalRawChars = existing.OriginalRawChars
	}
	if incoming.ProcessedChars == 0 {
		incoming.ProcessedChars = existing.ProcessedChars
	}
	if incoming.MaxTxt == 0 {
		incoming.MaxTxt = existing.MaxTxt
	}
	if incoming.Versions == nil {
		incoming.Versions = existing.Versions
	}
	if incoming.Logs == nil {
		incoming.Logs = existing.Logs
	}
	return incoming
}

func decodeNovelFetchJSON(w http.ResponseWriter, req *http.Request, target any) error {
	req.Body = http.MaxBytesReader(w, req.Body, 64<<20)
	decoder := json.NewDecoder(req.Body)
	return decoder.Decode(target)
}
