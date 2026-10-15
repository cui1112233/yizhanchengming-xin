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

	"qiantie/backend/internal/auth"
	"qiantie/backend/internal/novelfetchworkshop"
)

var novelFetchBookIDPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)
var novelFetchVersionIDPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]{1,64}$`)
var novelFetchNumericIDPattern = regexp.MustCompile(`\d{10,25}`)

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

func workshopProcessResult(ctx context.Context, store novelfetchworkshop.Store, owner, inputText, platformID, batchID string, maxTxt int) (map[string]any, error) {
	if strings.TrimSpace(inputText) == "" {
		return nil, errors.New("请输入要处理的书籍清单")
	}
	if maxTxt <= 0 {
		maxTxt = 4000
	}
	documents, parsed, emptyIDs, duplicates := parseWorkshopDocuments(inputText, maxTxt)
	for _, document := range documents {
		if document.Meta == nil {
			document.Meta = map[string]any{}
		}
		document.Meta["platformId"] = platformID
		document.Meta["currentBatchId"] = batchID
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
	return map[string]any{
		"batch_id": batchID, "current_batch_id": batchID,
		"parsed": parsed, "uniqueTasks": len(documents), "unique_tasks": len(documents), "duplicateCount": duplicates, "duplicate_count": duplicates,
		"emptyIdCount": emptyIDs, "empty_id_count": emptyIDs, "classifyErrors": []any{}, "classify_errors": []any{},
		"fetched": 0, "fetchFailed": 0, "fetch_failed": 0, "raw_fetched": 0, "generatedAiFiles": 0, "generated_ai_files": 0, "tasks": tasks,
	}, nil
}

// workshopPreviewResult parses the same input as the start endpoint without
// writing documents. The V78 workbench uses this as a first step to show the
// user how many books were recognised before starting the real run.
func workshopPreviewResult(inputText string, maxTxt int) (map[string]any, error) {
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
	return map[string]any{
		"parsed": parsed, "unique_tasks": len(tasks), "duplicate_count": duplicates,
		"empty_id_count": emptyIDs, "tasks": tasks,
	}, nil
}

func registerNovelFetchWorkshopRoutes(mux *http.ServeMux, store novelfetchworkshop.Store, aiJudge novelfetchworkshop.AIGenderJudge, aiJudgeResolver func(context.Context, string) (novelfetchworkshop.AIGenderJudge, error), fetcher *novelfetchworkshop.UpstreamFetcher, rewriteProvider novelfetchworkshop.RewriteProvider, rewriteResolver func(context.Context, string) (novelfetchworkshop.RewriteProvider, error)) {
	resolveGenderJudge := func(req *http.Request) novelfetchworkshop.AIGenderJudge {
		if aiJudgeResolver == nil {
			return aiJudge
		}
		identity, _ := BridgeIdentityFromContext(req.Context())
		resolved, err := aiJudgeResolver(req.Context(), identity.Username)
		if err == nil && resolved != nil {
			return resolved
		}
		return aiJudge
	}
	resolveRewriteProvider := func(req *http.Request) (novelfetchworkshop.RewriteProvider, error) {
		if rewriteResolver != nil {
			identity, _ := BridgeIdentityFromContext(req.Context())
			resolved, err := rewriteResolver(req.Context(), identity.Username)
			if err != nil {
				return nil, err
			}
			if resolved != nil {
				return resolved, nil
			}
		}
		return rewriteProvider, nil
	}
	registry := newWorkshopProcessRegistry()
	mux.HandleFunc("POST /api/novel-fetch-workshop/process", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		var input struct {
			InputText       string `json:"inputText"`
			InputTextLegacy string `json:"input_text"`
			MaxTxt          int    `json:"maxTxt"`
			MaxTxtLegacy    int    `json:"max_txt"`
			PlatformID      string `json:"platform_id"`
			PlatformIDCamel string `json:"platformId"`
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
			InputText       string `json:"input_text"`
			InputTextLegacy string `json:"inputText"`
			MaxTxt          int    `json:"max_txt"`
			MaxTxtLegacy    int    `json:"maxTxt"`
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
		result, err := workshopPreviewResult(input.InputText, input.MaxTxt)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, result)
	})

	mux.HandleFunc("POST /api/novel-fetch-workshop/process/start", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		var input struct {
			InputText       string `json:"inputText"`
			InputTextLegacy string `json:"input_text"`
			MaxTxt          int    `json:"maxTxt"`
			MaxTxtLegacy    int    `json:"max_txt"`
			PlatformID      string `json:"platform_id"`
			PlatformIDCamel string `json:"platformId"`
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
		now := time.Now().UTC().Format(time.RFC3339Nano)
		id := fmt.Sprintf("go-workshop-%d", time.Now().UnixNano())
		job := map[string]any{"id": id, "status": "running", "queue_state": "running", "started_at": now, "updated_at": now, "attempts": 1, "steps": []any{}, "progress": []any{}}
		batchID := fmt.Sprintf("batch-%d", time.Now().UnixNano())
		result, err := workshopProcessResult(req.Context(), store, identity.Username, input.InputText, input.PlatformID, batchID, input.MaxTxt)
		if err != nil {
			job["status"] = "failed"
			job["queue_state"] = "failed"
			job["error"] = err.Error()
		} else {
			job["status"] = "done"
			job["queue_state"] = "done"
			job["result"] = result
		}
		job["updated_at"] = time.Now().UTC().Format(time.RFC3339Nano)
		registry.put(identity.Username, id, job)
		writeJSON(w, http.StatusOK, job)
	})

	mux.HandleFunc("GET /api/novel-fetch-workshop/process/jobs/latest", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		limit, _ := strconv.Atoi(req.URL.Query().Get("limit"))
		jobs := registry.latest(identity.Username, limit)
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
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "处理任务不存在"})
			return
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
		documents, err := store.ListDocuments(req.Context(), identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "读取任务队列失败"})
			return
		}
		counts := map[string]int{"running": 0, "queued": 0, "waiting_retry": 0, "failed": 0, "stopped": 0}
		for _, document := range documents {
			task := legacyNovelFetchTask(document)
			status := strings.ToLower(firstText(task["ai_status"], document.Meta["status"], document.Meta["taskStatus"], document.Meta["aiStatus"]))
			switch {
			case strings.Contains(status, "running") || strings.Contains(status, "processing"):
				counts["running"]++
			case strings.Contains(status, "queued") || strings.Contains(status, "queue"):
				counts["queued"]++
			case strings.Contains(status, "retry"):
				counts["waiting_retry"]++
			case strings.Contains(status, "failed") || strings.Contains(status, "error"):
				counts["failed"]++
			case strings.Contains(status, "stopped") || strings.Contains(status, "cancel"):
				counts["stopped"]++
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{"counts": counts})
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
			syncNovelFetchClassificationMetadata(&task)
			legacyTasks = append(legacyTasks, legacyNovelFetchTask(task))
		}
		writeJSON(w, http.StatusOK, map[string]any{"tasks": legacyTasks})
	})

	// 批量重试由 Go 直接处理，避免小说获取工作台因为一个重试按钮又回到 Node。
	mux.HandleFunc("POST /api/novel-fetch-workshop/tasks/batch-retry", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		var input struct {
			IDs []string `json:"ids"`
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
				result, fetchErr := fetcher.FetchAndProcess(req.Context(), bookID, platformID, maxTxt)
				if fetchErr != nil {
					failed++
					retryStages["fetch"]++
					continue
				}
				processedText, sensitiveCount, sensitiveItems, _ := applyNovelFetchSensitiveRulesDetailed(req.Context(), store, identity.Username, result.ProcessedText)
				document.OriginalRaw = result.OriginalRaw
				document.Original = processedText
				document.ProcessedText = processedText
				document.OriginalRawChars = result.OriginalRawChars
				document.ProcessedChars = len([]rune(processedText))
				document.MaxTxt = result.MaxTxt
				document.BookInfo = result.BookInfo
				if document.Meta == nil {
					document.Meta = map[string]any{}
				}
				document.Meta["originalStatus"] = "done"
				document.Meta["originalRawChars"] = result.OriginalRawChars
				document.Meta["processedChars"] = document.ProcessedChars
				document.Meta["sensitiveStatus"] = "done"
				document.Meta["sensitiveReplaced"] = sensitiveCount
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
				if _, rewriteErr := novelfetchworkshop.GenerateAIRewrites(req.Context(), store, identity.Username, bookID, 1, provider); rewriteErr != nil {
					failed++
					retryStages["rewrite"]++
					continue
				}
				actionTaken = true
				retryStages["rewrite"]++
			}
			if actionTaken {
				retried++
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
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "retried": retried, "failed": failed, "skipped": skipped, "retry_stage_counts": retryStages, "tasks": tasks})
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
			MaxTxt int `json:"maxTxt"`
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
		result, err := fetcher.FetchAndProcess(req.Context(), bookID, platformID, maxTxt)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
			return
		}
		processedText, sensitiveCount, sensitiveItems, _ := applyNovelFetchSensitiveRulesDetailed(req.Context(), store, identity.Username, result.ProcessedText)
		document.OriginalRaw = result.OriginalRaw
		document.ProcessedText = processedText
		document.Original = processedText
		document.OriginalRawChars = result.OriginalRawChars
		document.ProcessedChars = len([]rune(processedText))
		document.MaxTxt = result.MaxTxt
		document.BookInfo = result.BookInfo
		if document.Meta == nil {
			document.Meta = map[string]any{}
		}
		document.Meta["originalStatus"] = "done"
		document.Meta["originalRawChars"] = result.OriginalRawChars
		document.Meta["processedChars"] = document.ProcessedChars
		document.Meta["sensitiveStatus"] = "done"
		document.Meta["sensitiveReplaced"] = sensitiveCount
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
		document, err := novelfetchworkshop.GenerateAIRewrites(req.Context(), store, identity.Username, bookID, input.Count, provider)
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
			Text string `json:"text"`
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
		items, err := novelfetchworkshop.AnalyzeOpening(req.Context(), provider, input.Text)
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
	inferNovelFetchStyleFailure(document.Meta)
	if style := strings.TrimSpace(firstText(document.Meta["style"], document.Meta["styleName"], document.Meta["style_name"])); style != "" && style != "待判定" && style != "未判定" {
		if strings.TrimSpace(fmt.Sprint(document.Meta["styleStatus"])) == "" {
			document.Meta["styleStatus"] = "classified"
		}
	}
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
	for _, item := range items {
		hits = append(hits, map[string]any{
			"hit_index": item["hit_index"], "keyword": item["keyword"],
			"snippet": item["before"], "status": item["status"],
		})
		fixed = append(fixed, item)
	}
	meta["sensitiveHits"] = map[string]any{"mode": "system_rule", "status": "done", "hit_count": count, "updated_at": updated, "hits": hits}
	meta["sensitiveFixed"] = map[string]any{"mode": "system_rule", "status": "done", "fixed_count": count, "failed_count": 0, "updated_at": updated, "items": fixed}
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
	bookName := metaString("bookName", "book_name")
	platformName := metaString("platformName", "platform_name")
	style := metaString("style", "styleName", "style_name")
	gender := strings.TrimSpace(document.Gender)
	if gender == "" {
		gender = metaString("gender")
	}
	result["id"] = document.BookID
	result["book_id"] = document.BookID
	result["book_name"] = bookName
	result["platform_name"] = platformName
	result["style"] = style
	result["gender"] = gender
	result["classify_status"] = "classify_failed"
	if gender != "" {
		result["classify_status"] = "classified"
	}
	result["original_status"] = metaString("originalStatus", "original_status")
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
		}
	}
	if len(generated) > 0 {
		result["ai_files"] = generated
		result["selected_versions"] = generated
		result["ai_status"] = "ai_done"
	}
	result["site_submit_status"] = metaString("siteSubmitStatus", "site_submit_status")
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
