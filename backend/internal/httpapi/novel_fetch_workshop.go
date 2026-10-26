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
	"qiantie/backend/internal/videomanagementconnector"
)

var novelFetchBookIDPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)
var novelFetchVersionIDPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]{1,64}$`)

// The legacy workbench accepts source IDs with a trailing 5-25 digit run.
// Keep that range here so short but valid platform IDs are not discarded
// before the fetch stage gets a chance to handle them.
var novelFetchNumericIDPattern = regexp.MustCompile(`\d{5,25}`)
var errNovelFetchOriginalFetch = errors.New("原文获取失败")

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
	failureMessages := make([]string, 0)
	stopped := false
	completed := 0
	for _, bookID := range activeBookIDs {
		bookName := ""
		if document, err := worker.store.GetDocument(ctx, job.Owner, bookID); err == nil {
			bookName = firstText(document.BookInfo.WorkTitle, document.Meta["bookName"], document.Meta["book_name"])
		}
		progress = append(progress, workshopProgressItem("processing", "开始处理", bookID, bookName))
		worker.registry.put(job.Owner, job.ID, map[string]any{"id": job.ID, "status": "running", "queue_state": "running", "started_at": startedAt, "updated_at": time.Now().UTC().Format(time.RFC3339Nano), "attempts": job.Attempts, "steps": progress, "progress": progress})
		_ = updateWorkshopTaskRuntimeState(ctx, worker.store, job.Owner, bookID, "processing", "")
		bookJob := activeJob
		bookJob.BookIDs = []string{bookID}
		if err := worker.execute(ctx, bookJob); err != nil {
			failedBookIDs = append(failedBookIDs, bookID)
			failureMessages = append(failureMessages, bookID+"："+err.Error())
			status := "failed"
			if errors.Is(err, errNovelFetchOriginalFetch) {
				deferredOriginalFetchRetry = true
				status = "waiting_retry"
				if retryErr := markNovelFetchOriginalFetchRetry(ctx, worker.store, job.Owner, bookID, err, time.Now().UTC()); retryErr != nil {
					_ = updateWorkshopTaskRuntimeState(ctx, worker.store, job.Owner, bookID, status, err.Error())
				}
			} else if retry {
				status = "waiting_retry"
				_ = updateWorkshopTaskRuntimeState(ctx, worker.store, job.Owner, bookID, status, err.Error())
			} else {
				_ = updateWorkshopTaskRuntimeState(ctx, worker.store, job.Owner, bookID, status, err.Error())
			}
			progress = append(progress, workshopProgressItem("error", "处理失败："+err.Error(), bookID, bookName))
			worker.registry.put(job.Owner, job.ID, map[string]any{"id": job.ID, "status": status, "queue_state": "running", "started_at": startedAt, "updated_at": time.Now().UTC().Format(time.RFC3339Nano), "attempts": job.Attempts, "steps": progress, "progress": progress})
			continue
		}
		if worker.bookCancelled(ctx, job.Owner, bookID) {
			stopped = true
			_ = updateWorkshopTaskRuntimeState(ctx, worker.store, job.Owner, bookID, "stopped", "")
			continue
		}
		completed++
		_ = clearNovelFetchOriginalFetchRetry(ctx, worker.store, job.Owner, bookID)
		_ = updateWorkshopTaskRuntimeState(ctx, worker.store, job.Owner, bookID, "completed", "")
		progress = append(progress, workshopProgressItem("success", "处理完成", bookID, bookName))
		worker.registry.put(job.Owner, job.ID, map[string]any{"id": job.ID, "status": "running", "queue_state": "running", "started_at": startedAt, "updated_at": time.Now().UTC().Format(time.RFC3339Nano), "attempts": job.Attempts, "steps": progress, "progress": progress})
	}

	finishedAt := time.Now().UTC().Format(time.RFC3339Nano)
	if len(failedBookIDs) > 0 {
		errorMessage := strings.Join(failureMessages, "；")
		job.LastError = errorMessage
		status := "failed"
		queueState := "failed"
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
		if retry {
			status = "waiting_retry"
			queueState = "queued"
		}
		_ = updateWorkshopRun(worker.store, job, status, errorMessage, startedAt, !retry)
		if !retry {
			_ = updateNovelFetchScheduleRuntimeStatus(ctx, worker.store, job.Owner, job.ScheduleID, "failed", errorMessage)
		}
		documents, _ := worker.store.ListDocuments(ctx, job.Owner)
		result := map[string]any{"batch_id": job.BatchID, "current_batch_id": job.BatchID, "tasks": documents, "completed": completed, "failed": len(failedBookIDs), "failed_book_ids": failedBookIDs}
		progressMessage := "本轮结束，失败任务等待重试"
		if !retry {
			progressMessage = "处理结束，任务失败"
		}
		progress = append(progress, workshopProgressItem("warning", progressMessage, "", ""))
		worker.registry.put(job.Owner, job.ID, map[string]any{"id": job.ID, "status": status, "queue_state": queueState, "started_at": startedAt, "updated_at": finishedAt, "attempts": job.Attempts, "error": errorMessage, "result": result, "steps": progress, "progress": progress})
		retryJob := job
		retryJob.BookIDs = failedBookIDs
		_ = worker.queue.Fail(ctx, retryJob, retry)
		return
	}

	documents, _ := worker.store.ListDocuments(ctx, job.Owner)
	result := map[string]any{"batch_id": job.BatchID, "current_batch_id": job.BatchID, "tasks": documents}
	finalStatus := "done"
	if stopped {
		finalStatus = "stopped"
	}
	_ = updateWorkshopRun(worker.store, job, finalStatus, "", startedAt, true)
	scheduleStatus := "completed"
	scheduleMessage := ""
	if stopped {
		scheduleStatus = "stopped"
	} else if job.AutoSubmit && worker.autoSubmit != nil {
		progress = append(progress, workshopProgressItem("processing", "开始自动上传到视频管理系统", "", ""))
		worker.registry.put(job.Owner, job.ID, map[string]any{"id": job.ID, "status": "running", "queue_state": "running", "started_at": startedAt, "updated_at": time.Now().UTC().Format(time.RFC3339Nano), "attempts": job.Attempts, "steps": progress, "progress": progress})
		if err := worker.autoSubmit(ctx, job); err != nil {
			scheduleStatus = "upload_pending"
			scheduleMessage = err.Error()
			for _, bookID := range activeBookIDs {
				_ = appendWorkshopTaskLog(ctx, worker.store, job.Owner, bookID, "auto_submit_pending", "自动提交未完成："+err.Error())
			}
			progress = append(progress, workshopProgressItem("warning", "自动上传未完成："+err.Error(), "", ""))
		} else {
			for _, bookID := range activeBookIDs {
				_ = appendWorkshopTaskLog(ctx, worker.store, job.Owner, bookID, "auto_submit_completed", "自动提交完成")
			}
			progress = append(progress, workshopProgressItem("success", "自动上传完成", "", ""))
		}
	}
	_ = updateNovelFetchScheduleRuntimeStatus(ctx, worker.store, job.Owner, job.ScheduleID, scheduleStatus, scheduleMessage)
	progress = append(progress, workshopProgressItem("success", "批次处理完成", "", ""))
	worker.registry.put(job.Owner, job.ID, map[string]any{"id": job.ID, "status": finalStatus, "queue_state": finalStatus, "started_at": startedAt, "updated_at": finishedAt, "attempts": job.Attempts, "result": result, "steps": progress, "progress": progress})
	_ = worker.queue.Ack(ctx, job)
}

func workshopProgressItem(level, message, bookID, bookName string) map[string]any {
	return map[string]any{
		"time": time.Now().UTC().Format(time.RFC3339Nano), "level": level,
		"message": message, "book_id": strings.TrimSpace(bookID), "book_name": strings.TrimSpace(bookName),
	}
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
	return strings.EqualFold(firstText(document.Meta["status"], document.Meta["queueState"]), "stopped")
}

func (worker *workshopProcessWorker) execute(ctx context.Context, job novelfetchworkshop.ProcessJob) error {
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
			if worker.fetcher == nil {
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
		if err := ensureWorkshopDocumentClassification(ctx, worker.store, job.Owner, bookID, judge, classificationProvider); err != nil {
			return err
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
		if worker.resolveRewriteProvider == nil {
			return errors.New("未配置文本模型")
		}
		provider, err := worker.resolveRewriteProvider(ctx, job.Owner)
		if err != nil {
			return err
		}
		if _, err := novelfetchworkshop.GenerateAIRewritesForVersionsProcessed(ctx, worker.store, job.Owner, bookID, pendingAI, provider, workshopRewriteProcessor(worker.store, job.Owner, bookID, job.SensitiveAIEnabled, provider)); err != nil {
			return err
		}
	}
	return nil
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
	if genderNeedsResolution(document.Gender) {
		info := document.BookInfo
		info.Text = text
		decision := novelfetchworkshop.ResolveGender(ctx, "", info, genderJudge)
		document.Gender, document.GenderSource, document.GenderReason = decision.Gender, decision.Source, decision.Reason
	}
	if styleNeedsResolution(document.Meta) {
		document.Meta["styleStatus"] = "processing"
		info := document.BookInfo
		info.Text = text
		recordNovelFetchStyleDecision(document.Meta, novelfetchworkshop.ResolveStyle(ctx, info, styleProvider), nil)
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
	document.Logs = append(document.Logs, map[string]any{
		"time":   time.Now().UTC().Format(time.RFC3339Nano),
		"event":  "process_" + status,
		"detail": detail,
	})
	if len(document.Logs) > 200 {
		document.Logs = document.Logs[len(document.Logs)-200:]
	}
	return store.PutDocument(ctx, owner, document)
}

// appendWorkshopTaskLog persists a concise per-book operation log so that the
// task detail remains useful after the in-memory realtime progress expires.
func appendWorkshopTaskLog(ctx context.Context, store novelfetchworkshop.Store, owner, bookID, event, detail string) error {
	document, err := store.GetDocument(ctx, owner, bookID)
	if err != nil {
		return err
	}
	document.Logs = append(document.Logs, map[string]any{
		"time":   time.Now().UTC().Format(time.RFC3339Nano),
		"event":  strings.TrimSpace(event),
		"detail": strings.TrimSpace(detail),
	})
	if len(document.Logs) > 200 {
		document.Logs = document.Logs[len(document.Logs)-200:]
	}
	return store.PutDocument(ctx, owner, document)
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
)

func runNovelFetchScheduleDispatcher(ctx context.Context, store novelfetchworkshop.Store, queue novelfetchworkshop.ProcessQueue, owners novelfetchworkshop.ConfigOwnerLister, registry *workshopProcessRegistry) {
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
	changed := false
	for _, raw := range schedules {
		item, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		status := strings.ToLower(strings.TrimSpace(firstText(item["status"])))
		if status != "" && status != "scheduled" {
			continue
		}
		runAt, err := parseNovelFetchScheduleTime(firstText(item["run_at"], item["runAt"]))
		if err != nil || runAt.After(now) {
			continue
		}
		changed = true
		item["status"] = "dispatching"
		item["updatedAt"] = now.UTC().Format(time.RFC3339Nano)
		if err := store.PutConfig(ctx, owner, settings); err != nil {
			return err
		}

		maxTxt := int(targetNumber(item["max_txt"]))
		if maxTxt <= 0 {
			maxTxt = int(targetNumber(item["maxTxt"]))
		}
		if maxTxt <= 0 {
			maxTxt = 4000
		}
		inputText := firstText(item["input_text"], item["inputText"])
		platformID := firstText(item["platform_id"], item["platformId"], "2")
		jobID := fmt.Sprintf("go-schedule-%d", time.Now().UnixNano())
		batchID := fmt.Sprintf("batch-%d", time.Now().UnixNano())
		_, dispatchErr := workshopProcessResult(ctx, store, owner, inputText, platformID, batchID, maxTxt)
		documents, _, _, _ := parseWorkshopDocuments(inputText, maxTxt)
		bookIDs := make([]string, 0, len(documents))
		for _, document := range documents {
			bookIDs = append(bookIDs, document.BookID)
		}
		sensitiveAIEnabled := novelFetchSensitiveAIEnabled(settings)
		if raw, exists := item["sensitive_ai_enabled"]; exists {
			if explicit, valid := raw.(bool); valid {
				sensitiveAIEnabled = explicit
			}
		}
		job := novelfetchworkshop.ProcessJob{
			ID: jobID, Owner: owner, BatchID: batchID, ScheduleID: firstText(item["id"]), BookIDs: bookIDs, PlatformID: platformID, MaxTxt: maxTxt,
			SensitiveAIEnabled: sensitiveAIEnabled,
			AutoSubmit:         item["auto_submit"] != false,
			SelectedVersions:   scheduleStringSlice(item["selected_versions"]),
			RewriteMethods:     scheduleStringMap(item["rewrite_methods"]),
			TextModelID:        firstText(item["text_model_id"], item["textModelId"], "未指定"),
		}
		if dispatchErr == nil && len(bookIDs) == 0 {
			dispatchErr = errors.New("定时任务没有有效书籍")
		}
		if dispatchErr == nil {
			if cancellationQueue, ok := queue.(novelfetchworkshop.ProcessCancellationQueue); ok {
				dispatchErr = cancellationQueue.ClearCancelledBooks(ctx, owner, bookIDs)
			}
		}
		startedAt := now.UTC().Format(time.RFC3339Nano)
		if dispatchErr == nil {
			dispatchErr = store.PutRun(ctx, owner, novelfetchworkshop.RunRecord{RunID: jobID, BookID: batchID, Stage: "novel_fetch_process", Status: "queued", TextModelID: job.TextModelID, ModelID: "novel-fetch-workshop", ModelDisplayName: "小说获取完整流程", StartedAt: startedAt})
		}
		if dispatchErr == nil {
			dispatchErr = queue.Enqueue(ctx, job)
		}
		if dispatchErr != nil {
			item["status"] = "failed"
			item["error"] = dispatchErr.Error()
		} else {
			item["status"] = "queued"
			item["job_id"] = jobID
			item["batch_id"] = batchID
			delete(item, "error")
		}
		item["updatedAt"] = time.Now().UTC().Format(time.RFC3339Nano)
	}
	if changed {
		return store.PutConfig(ctx, owner, settings)
	}
	return nil
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
	for _, document := range documents {
		if document.Meta == nil {
			document.Meta = map[string]any{}
		}
		document.Meta["platformId"] = platformID
		document.Meta["currentBatchId"] = batchID
		metadata.apply(document.Meta)
		if err := store.PutDocument(ctx, owner, document); err != nil {
			return nil, err
		}
	}
	tasks, err := store.ListDocuments(ctx, owner)
	if err != nil {
		return nil, err
	}
	settings, _ := store.GetConfig(ctx, owner)
	if settings == nil {
		settings = map[string]any{}
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
	}
	for key, value := range metadata.responseFields() {
		result[key] = value
	}
	return result, nil
}

// applySensitiveAIRepair runs only when the user explicitly enables the
// sensitive-word AI repair switch. Deterministic rules still run first; the
// model then refines each hit line and the audit item records the source,
// model and attempt count shown by the workbench.
func applySensitiveAIRepair(ctx context.Context, provider novelfetchworkshop.RewriteProvider, settings map[string]any, bookID, text string, items []map[string]any) (string, []map[string]any) {
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
		items[index]["source"] = "AI鉴组处理"
		items[index]["model"] = model
		items[index]["status"] = "skipped_limit"
		items[index]["error"] = fmt.Sprintf("超过本任务命中上限 %d", maxHits)
	}
	active := items[:maxHits]
	if groupHits && len(active) > 1 {
		outputs := make([]string, 0, len(active))
		for _, item := range active {
			item["source"] = "AI鉴组处理"
			item["model"] = model
			keyword := strings.TrimSpace(fmt.Sprint(item["keyword"]))
			snippet := sensitiveContextSnippet(strings.TrimSpace(fmt.Sprint(item["before"])), keyword, contextChars)
			outputs = append(outputs, fmt.Sprintf("%d. 命中词：%s\n原片段：%s", len(outputs)+1, keyword, snippet))
		}
		userPrompt := strings.NewReplacer("{book_id}", bookID, "{keyword}", "多个命中词", "{snippet}", strings.Join(outputs, "\n\n")).Replace(promptTemplate)
		userPrompt += "\n\n请严格按原顺序返回 JSON 字符串数组，每个元素对应一个片段，只返回数组。"
		var output string
		var lastErr error
		attempts := 0
		if provider == nil {
			lastErr = errors.New("敏感词 AI 修复模型未配置")
		} else {
			for attempts < maxAttempts {
				attempts++
				output, lastErr = provider.Complete(ctx, novelfetchworkshop.RewriteRequest{SystemPrompt: "你是内容合规改写助手。只返回 JSON 字符串数组。", UserPrompt: userPrompt, Temperature: temperature, MaxTokens: 1200})
				output = strings.TrimSpace(output)
				if lastErr == nil && output != "" {
					break
				}
			}
		}
		var repaired []string
		if lastErr == nil {
			lastErr = json.Unmarshal([]byte(output), &repaired)
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
	for _, item := range active {
		item["source"] = "AI鉴组处理"
		item["model"] = model
		before := strings.TrimSpace(fmt.Sprint(item["before"]))
		keyword := strings.TrimSpace(fmt.Sprint(item["keyword"]))
		before = sensitiveContextSnippet(before, keyword, contextChars)
		attempts := 0
		var output string
		var lastErr error
		if provider == nil {
			lastErr = errors.New("敏感词 AI 修复模型未配置")
		} else {
			userPrompt := strings.NewReplacer("{book_id}", bookID, "{keyword}", keyword, "{snippet}", before).Replace(promptTemplate)
			for attempts < maxAttempts {
				attempts++
				output, lastErr = provider.Complete(ctx, novelfetchworkshop.RewriteRequest{SystemPrompt: "你是内容合规改写助手。只返回改写后的小段内容。", UserPrompt: userPrompt, Temperature: temperature, MaxTokens: 1200})
				output = strings.TrimSpace(output)
				if lastErr == nil && output != "" {
					break
				}
			}
		}
		item["attempt"] = attempts
		if lastErr != nil || output == "" || containsSensitiveKeyword(output, keyword) {
			item["status"] = "failed"
			item["error"] = firstText(lastErr, "AI 返回为空或仍包含命中敏感词")
			continue
		}
		item["status"] = "done"
		item["error"] = ""
		text, searchCursor = replaceSensitiveHit(text, searchCursor, item, output)
		item["after"] = output
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
		processed, sensitiveItems = applySensitiveAIRepair(ctx, provider, settings, bookID, processed, sensitiveItems)
		failed := 0
		for _, item := range sensitiveItems {
			if strings.EqualFold(strings.TrimSpace(fmt.Sprint(item["status"])), "failed") {
				failed++
			}
		}
		trace = append(trace, map[string]any{"stage": "sensitive_ai", "title": "敏感词 AI 修复", "changed": before != processed, "chars": len([]rune(processed)), "failed": failed})
		if failed > 0 {
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
				base = &http.Client{}
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
	originalStatus := strings.ToLower(strings.TrimSpace(firstText(document.Meta["originalStatus"], document.Meta["original_status"])))
	if strings.TrimSpace(document.OriginalRaw) != "" && originalStatus == "waiting_sensitive_ai" {
		processed := novelfetchworkshop.ProcessRawText(document.OriginalRaw, maxTxt)
		result = novelfetchworkshop.FetchAndProcessResult{
			OriginalRaw: document.OriginalRaw, ProcessedText: processed.ProcessedText,
			BookInfo: document.BookInfo, OriginalRawChars: processed.OriginalRawChars,
			ProcessedChars: processed.ProcessedChars, MaxTxt: processed.MaxTxt,
		}
	} else {
		platformID := "2"
		if value, ok := document.Meta["platformId"].(string); ok && strings.TrimSpace(value) != "" {
			platformID = strings.TrimSpace(value)
		}
		result, err = fetchNovelWithSettings(ctx, store, owner, fetcher, bookID, platformID, maxTxt)
		if err != nil {
			return fmt.Errorf("%w: %v", errNovelFetchOriginalFetch, err)
		}
		if err := stageNovelFetchRaw(ctx, store, owner, &document, result); err != nil {
			return err
		}
	}
	if limit, enabled := novelFetchShortOriginalLimit(ctx, store, owner); enabled && result.ProcessedChars > 0 && result.ProcessedChars < limit {
		return markNovelFetchShortOriginal(ctx, store, owner, &document, result, limit)
	}
	processedText, rulesTrace, sensitiveItems, err := processWorkshopConfiguredText(ctx, store, owner, bookID, result.ProcessedText, "original", sensitiveAIEnabled, sensitiveProvider)
	if err != nil {
		return err
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
	if err := persistNovelFetchBodies(ctx, store, owner, &document, result.OriginalRaw, processedText); err != nil {
		return err
	}
	if genderNeedsResolution(document.Gender) {
		genderInfo := result.BookInfo
		genderInfo.Text = processedText
		decision := novelfetchworkshop.ResolveGender(ctx, "", genderInfo, genderJudge)
		document.Gender, document.GenderSource, document.GenderReason = decision.Gender, decision.Source, decision.Reason
	}
	if styleNeedsResolution(document.Meta) {
		document.Meta["styleStatus"] = "processing"
		styleInfo := result.BookInfo
		styleInfo.Text = processedText
		style := novelfetchworkshop.ResolveStyle(ctx, styleInfo, styleProvider)
		recordNovelFetchStyleDecision(document.Meta, style, nil)
	}
	return store.PutDocument(ctx, owner, document)
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
	documents, parsed, emptyIDs, duplicates := parseWorkshopDocuments(inputText, maxTxt)
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
	metadata := workshopInputMetadata{}
	if len(sourceMetadata) > 0 {
		metadata = sourceMetadata[0]
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

func registerNovelFetchWorkshopRoutes(mux *http.ServeMux, store novelfetchworkshop.Store, queue novelfetchworkshop.ProcessQueue, aiJudge novelfetchworkshop.AIGenderJudge, aiJudgeResolver func(context.Context, string) (novelfetchworkshop.AIGenderJudge, error), fetcher *novelfetchworkshop.UpstreamFetcher, rewriteProvider novelfetchworkshop.RewriteProvider, rewriteResolver func(context.Context, string) (novelfetchworkshop.RewriteProvider, error), usersDir string, uploadTarget videomanagementconnector.Service) {
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
			result, err := executeWebSubmit(ctx, job.Owner, usersDir, store, webSubmitRequest{TaskIDs: job.BookIDs, Versions: job.SelectedVersions}, func(ctx context.Context, owner string) (novelfetchworkshop.AIGenderJudge, error) {
				return resolveGenderJudgeFor(ctx, owner), nil
			}, uploadTarget)
			if err != nil {
				markNovelFetchAutomationSubmitResult(ctx, store, job.Owner, job.BookIDs, err)
				return err
			}
			if result.Failed > 0 {
				err = fmt.Errorf("%d 个版本待上传", result.Failed)
				markNovelFetchAutomationSubmitResult(ctx, store, job.Owner, job.BookIDs, err)
				return err
			}
			markNovelFetchAutomationSubmitResult(ctx, store, job.Owner, job.BookIDs, nil)
			return nil
		}
		go worker.run(context.Background())
		if owners, ok := store.(novelfetchworkshop.ConfigOwnerLister); ok {
			go runNovelFetchScheduleDispatcher(context.Background(), store, queue, owners, registry)
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
			InputText          string   `json:"input_text"`
			InputTextLegacy    string   `json:"inputText"`
			MaxTxt             int      `json:"max_txt"`
			MaxTxtLegacy       int      `json:"maxTxt"`
			InputFormat        string   `json:"input_format"`
			InputFormatLegacy  string   `json:"inputFormat"`
			ColumnPreset       string   `json:"column_preset"`
			ColumnPresetLegacy string   `json:"columnPreset"`
			ColumnOrder        []string `json:"column_order"`
			ColumnOrderLegacy  []string `json:"columnOrder"`
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
		columnOrder := input.ColumnOrder
		if len(columnOrder) == 0 {
			columnOrder = input.ColumnOrderLegacy
		}
		metadata := workshopInputMetadata{Format: firstText(input.InputFormat, input.InputFormatLegacy), ColumnPreset: firstText(input.ColumnPreset, input.ColumnPresetLegacy), ColumnOrder: columnOrder}
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
			ColumnPreset       string            `json:"column_preset"`
			ColumnPresetLegacy string            `json:"columnPreset"`
			ColumnOrder        []string          `json:"column_order"`
			ColumnOrderLegacy  []string          `json:"columnOrder"`
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
		if queue == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "Redis 任务队列未配置"})
			return
		}
		now := time.Now().UTC().Format(time.RFC3339Nano)
		id := fmt.Sprintf("go-workshop-%d", time.Now().UnixNano())
		batchID := fmt.Sprintf("batch-%d", time.Now().UnixNano())
		jobView := map[string]any{"id": id, "status": "queued", "queue_state": "queued", "started_at": now, "updated_at": now, "attempts": 0, "steps": []any{}, "progress": []any{}}
		textModelID := strings.TrimSpace(input.TextModelID)
		if textModelID == "" {
			textModelID = "未指定"
		}
		columnOrder := input.ColumnOrder
		if len(columnOrder) == 0 {
			columnOrder = input.ColumnOrderLegacy
		}
		metadata := workshopInputMetadata{Format: firstText(input.InputFormat, input.InputFormatLegacy), ColumnPreset: firstText(input.ColumnPreset, input.ColumnPresetLegacy), ColumnOrder: columnOrder}
		result, err := workshopProcessResult(req.Context(), store, identity.Username, input.InputText, input.PlatformID, batchID, input.MaxTxt, metadata)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "保存任务失败"})
			return
		}
		documents, _, _, _ := parseWorkshopDocuments(input.InputText, input.MaxTxt)
		bookIDs := make([]string, 0, len(documents))
		for _, document := range documents {
			bookIDs = append(bookIDs, document.BookID)
			_ = updateWorkshopTaskRuntimeState(req.Context(), store, identity.Username, document.BookID, "queued", "")
		}
		processJob := novelfetchworkshop.ProcessJob{ID: id, Owner: identity.Username, BatchID: batchID, BookIDs: bookIDs, PlatformID: input.PlatformID, MaxTxt: input.MaxTxt, SensitiveAIEnabled: input.SensitiveAIEnabled, AutoSubmit: result["auto_queue_enabled"] == true, SelectedVersions: input.SelectedVersions, RewriteMethods: input.RewriteMethods, TextModelID: textModelID}
		if cancellationQueue, ok := queue.(novelfetchworkshop.ProcessCancellationQueue); ok {
			if err := cancellationQueue.ClearCancelledBooks(req.Context(), identity.Username, bookIDs); err != nil {
				writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "Redis 取消状态清理失败"})
				return
			}
		}
		if err := store.PutRun(req.Context(), identity.Username, novelfetchworkshop.RunRecord{RunID: id, BookID: batchID, Stage: "novel_fetch_process", Status: "queued", TextModelID: textModelID, ModelID: "novel-fetch-workshop", ModelDisplayName: "小说获取完整流程", StartedAt: now}); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "保存任务审计失败"})
			return
		}
		jobView["result"] = result
		registry.put(identity.Username, id, jobView)
		if err := queue.Enqueue(req.Context(), processJob); err != nil {
			_ = updateWorkshopRun(store, processJob, "failed", err.Error(), now, true)
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "Redis 任务队列不可用"})
			return
		}
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
		retried, failed, skipped := 0, 0, 0
		updatedTaskIDs := make([]string, 0, len(ids))
		retryStages := map[string]int{}
		for _, bookID := range ids {
			document, err := store.GetDocument(req.Context(), identity.Username, bookID)
			if errors.Is(err, novelfetchworkshop.ErrNotFound) {
				failed++
				retryStages["failed"]++
				continue
			}
			if err != nil {
				failed++
				retryStages["failed"]++
				continue
			}
			actionTaken := false
			if strings.TrimSpace(fmt.Sprint(document.Meta["originalStatus"])) != "done" && fetcher != nil {
				platformID := "2"
				if value, ok := document.Meta["platformId"].(string); ok && strings.TrimSpace(value) != "" {
					platformID = strings.TrimSpace(value)
				}
				maxTxt := document.MaxTxt
				if maxTxt <= 0 {
					maxTxt = 4000
				}
				result, fetchErr := fetchNovelWithSettings(req.Context(), store, identity.Username, fetcher, bookID, platformID, maxTxt)
				if fetchErr != nil {
					failed++
					retryStages["fetch"]++
					continue
				}
				if err := stageNovelFetchRaw(req.Context(), store, identity.Username, &document, result); err != nil {
					failed++
					retryStages["fetch"]++
					continue
				}
				var sensitiveProvider novelfetchworkshop.RewriteProvider
				if input.SensitiveAIEnabled {
					if rewriteResolver != nil {
						sensitiveProvider, _ = rewriteResolver(req.Context(), identity.Username)
					} else {
						sensitiveProvider = rewriteProvider
					}
				}
				processedText, rulesTrace, sensitiveItems, processErr := processWorkshopConfiguredText(req.Context(), store, identity.Username, bookID, result.ProcessedText, "original", input.SensitiveAIEnabled, sensitiveProvider)
				if processErr != nil {
					failed++
					retryStages["rules"]++
					continue
				}
				sensitiveCount := len(sensitiveItems)
				document.Original = processedText
				document.ProcessedText = processedText
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
					failed++
					retryStages["fetch"]++
					continue
				}
				if err := store.PutDocument(req.Context(), identity.Username, document); err != nil {
					failed++
					retryStages["fetch"]++
					continue
				}
				actionTaken = true
				retryStages["fetch"]++
			}
			if strings.EqualFold(strings.TrimSpace(fmt.Sprint(document.Meta["aiStatus"])), "failed") {
				provider, providerErr := resolveRewriteProvider(req)
				if providerErr != nil || provider == nil {
					failed++
					retryStages["rewrite"]++
					continue
				}
				if _, rewriteErr := novelfetchworkshop.GenerateAIRewritesForVersionsProcessed(req.Context(), store, identity.Username, bookID, []string{"ai1"}, provider, workshopRewriteProcessor(store, identity.Username, bookID, input.SensitiveAIEnabled, provider)); rewriteErr != nil {
					failed++
					retryStages["rewrite"]++
					continue
				}
				actionTaken = true
				retryStages["rewrite"]++
			}
			if actionTaken {
				retried++
				updatedTaskIDs = append(updatedTaskIDs, bookID)
			} else {
				skipped++
				retryStages["skipped"]++
			}
		}
		tasks, err := store.ListDocuments(req.Context(), identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "读取任务列表失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"ok": true, "requested": len(ids), "updated": retried, "task_ids": updatedTaskIDs,
			"retried": retried, "failed": failed, "skipped": skipped, "retry_stage_counts": retryStages, "tasks": tasks,
		})
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
		writeJSON(w, http.StatusOK, document)
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
		if input.Count == 0 {
			input.Count = 1
		}
		provider, providerErr := resolveRewriteProvider(req)
		if providerErr != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": providerErr.Error()})
			return
		}
		versions := make([]string, 0, input.Count)
		for index := 1; index <= input.Count; index++ {
			versions = append(versions, fmt.Sprintf("ai%d", index))
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
		provider, providerErr := resolveRewriteProvider(req)
		if providerErr != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": providerErr.Error()})
			return
		}
		if provider == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": "AI 改文模型未配置"})
			return
		}
		content, err := provider.Complete(req.Context(), novelfetchworkshop.RewriteRequest{
			SystemPrompt: "你是接口连通性测试助手。",
			UserPrompt:   "请只回复：ok",
			Temperature:  0,
			MaxTokens:    64,
		})
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
			return
		}
		content = strings.TrimSpace(content)
		if len(content) > 200 {
			content = content[:200]
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
		content, err := provider.Complete(req.Context(), novelfetchworkshop.RewriteRequest{
			SystemPrompt: "你是小说排版规则设计师。根据文本提出可执行的规则建议。只返回 JSON：{\"summary\":\"...\",\"rules\":[{\"name\":\"...\",\"pattern\":\"...\",\"replacement\":\"...\"}]}，不要解释。",
			UserPrompt:   fmt.Sprintf("规则类型：%s\n目标：%s\n待分析文本：\n%s", input.RuleType, input.Goal, input.Text),
			Temperature:  0.2,
			MaxTokens:    2000,
		})
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "suggestions": strings.TrimSpace(content)})
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
		result, trace := novelfetchworkshop.PreviewRules(input.Text, input.Scope, input.LayoutConfig, input.Knowledge)
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "text": result, "result": result, "trace": trace, "config": input.LayoutConfig})
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
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "kind": req.PathValue("kind"), "data": library, "items": library["items"]})
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
		normalized, err := novelfetchworkshop.SaveKnowledge(settings, req.PathValue("kind"), item)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
			return
		}
		if err := store.PutConfig(req.Context(), identity.Username, settings); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "保存知识库失败"})
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
		removed, err := novelfetchworkshop.DeleteKnowledge(settings, req.PathValue("kind"), req.PathValue("id"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
			return
		}
		if err := store.PutConfig(req.Context(), identity.Username, settings); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "保存知识库失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "removed": removed})
	})

	mux.HandleFunc("POST /api/novel-fetch-workshop/knowledge/{kind}/{id}/optimize", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
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
		item, err := novelfetchworkshop.OptimizeKnowledge(req.Context(), settings, req.PathValue("kind"), req.PathValue("id"), provider)
		if err != nil {
			status := http.StatusBadRequest
			if provider == nil || strings.Contains(err.Error(), "模型未配置") {
				status = http.StatusServiceUnavailable
			}
			writeJSON(w, status, map[string]any{"ok": false, "error": err.Error()})
			return
		}
		if err := store.PutConfig(req.Context(), identity.Username, settings); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "保存知识库失败"})
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
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "保存开头词库失败"})
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
		_ = store.PutConfig(req.Context(), identity.Username, settings)
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
		result, err := store.DeleteDocuments(req.Context(), identity.Username, ids)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "删除任务失败"})
			return
		}
		writeJSON(w, http.StatusOK, result)
	})

	mux.HandleFunc("GET /api/novel-fetch-workshop/platforms", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		settings, err := store.GetConfig(req.Context(), identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "读取平台表失败"})
			return
		}
		platforms := make([]map[string]any, 0)
		if configured, ok := settings["platforms"].([]any); ok {
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
		if len(platforms) == 0 {
			for _, raw := range defaultNovelFetchPlatforms {
				platform := raw.(map[string]any)
				platforms = append(platforms, map[string]any{"id": platform["id"], "name": platform["name"]})
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
		appConfig := novelFetchAppConfig(publicSource)
		platforms, ok := publicSource["platforms"].([]any)
		if !ok {
			platforms = append([]any(nil), defaultNovelFetchPlatforms...)
		}
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
		saveNovelFetchConfig(w, req, store, func(merged map[string]any) map[string]any {
			return map[string]any{"ok": true, "settings": merged}
		})
	}
	// The current React workbench uses POST; keep PUT for the canonical API and
	// accept both methods during the Node-to-Go compatibility period.
	mux.HandleFunc("PUT /api/novel-fetch-workshop/config", saveWorkshopConfig)
	mux.HandleFunc("POST /api/novel-fetch-workshop/config", saveWorkshopConfig)
}

func saveNovelFetchConfig(w http.ResponseWriter, req *http.Request, store novelfetchworkshop.Store, response func(map[string]any) map[string]any) {
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
	current, err := store.GetConfig(req.Context(), identity.Username)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "读取设置失败"})
		return
	}
	merged := mergeWorkshopConfigPreservingSecrets(current, settings)
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
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "保存设置失败"})
		return
	}
	writeJSON(w, http.StatusOK, response(merged))
}

func genderNeedsResolution(value string) bool {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "", "待确认", "未知", "unknown":
		return true
	default:
		return false
	}
}

func selectedNovelFetchAIVersions(values []string) []string {
	selected := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, raw := range values {
		version := strings.ToLower(strings.TrimSpace(raw))
		if version == "original" || version == "" || !strings.HasPrefix(version, "ai") {
			continue
		}
		if !novelFetchVersionIDPattern.MatchString(version) || seen[version] {
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
		document.Meta["gender"] = document.Gender
		if document.GenderSource != "" {
			document.Meta["genderSource"] = string(document.GenderSource)
		}
		if strings.TrimSpace(document.GenderReason) != "" {
			document.Meta["genderReason"] = document.GenderReason
		}
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
	for _, item := range items {
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
	meta["sensitiveHits"] = map[string]any{"mode": "system_rule", "status": "done", "hit_count": count, "updated_at": updated, "hits": hits}
	status := "done"
	if failedCount > 0 {
		status = "failed"
	}
	meta["sensitiveFixed"] = map[string]any{"mode": "system_rule", "status": status, "fixed_count": fixedCount, "failed_count": failedCount, "updated_at": updated, "items": fixed}
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
		if version == "original_raw" {
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
		if version == "original" || seenString(generated, version) || !strings.EqualFold(strings.TrimSpace(body.State), "ready") {
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

func parseWorkshopDocuments(input string, maxTxt int) ([]novelfetchworkshop.Document, int, int, int) {
	lines := strings.Split(strings.ReplaceAll(input, "\r\n", "\n"), "\n")
	documents := make([]novelfetchworkshop.Document, 0)
	seen := make(map[string]struct{})
	parsed, emptyIDs, duplicates := 0, 0, 0
	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.Contains(line, "书籍ID") || strings.Contains(strings.ToLower(line), "book_id") {
			continue
		}
		parsed++
		cells := strings.FieldsFunc(line, func(r rune) bool { return r == '\t' || r == '|' || r == ',' })
		for index := range cells {
			cells[index] = strings.TrimSpace(cells[index])
		}
		bookID := ""
		for _, cell := range cells {
			if match := novelFetchNumericIDPattern.FindString(cell); match != "" {
				bookID = match
				break
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
		bookName, gender := "", ""
		// 常用输入是“ID + 若干空格 + 书名”。先从原始行移除 ID 和
		// 紧随其后的空白/分隔符，避免把 ID 一并保存成书名。
		if idIndex := strings.Index(line, bookID); idIndex >= 0 {
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
		for _, cell := range cells {
			if cell == bookID || strings.Contains(cell, "男频") || strings.Contains(cell, "女频") {
				if strings.Contains(cell, "男频") {
					gender = "男频"
				} else if strings.Contains(cell, "女频") {
					gender = "女频"
				}
				continue
			}
			if bookName == "" {
				bookName = cell
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
