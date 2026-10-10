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

type NovelFetchBridgeAuth struct {
	Secret    string
	Now       func() time.Time
	MaxSkew   time.Duration
	Users     BridgeUserResolver
	AuthStore auth.Store
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
		if err != nil {
			return BridgeIdentity{}, errors.New("invalid login session")
		}
		account, err := a.AuthStore.FindAccount(req.Context(), session.Username)
		if err != nil || !account.Active {
			return BridgeIdentity{}, errors.New("inactive login account")
		}
		return BridgeIdentity{Username: account.Username, IsOwner: account.IsOwner}, nil
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

func workshopProcessResult(ctx context.Context, store novelfetchworkshop.Store, owner, inputText string, maxTxt int) (map[string]any, error) {
	if strings.TrimSpace(inputText) == "" {
		return nil, errors.New("请输入要处理的书籍清单")
	}
	if maxTxt <= 0 {
		maxTxt = 4000
	}
	documents, parsed, emptyIDs, duplicates := parseWorkshopDocuments(inputText, maxTxt)
	for _, document := range documents {
		if err := store.PutDocument(ctx, owner, document); err != nil {
			return nil, err
		}
	}
	tasks, err := store.ListDocuments(ctx, owner)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"parsed": parsed, "emptyIdCount": emptyIDs, "uniqueTasks": len(documents), "duplicateCount": duplicates,
		"classifyErrors": []any{}, "fetched": 0, "fetchFailed": 0, "generatedAiFiles": 0, "tasks": tasks,
	}, nil
}

func registerNovelFetchWorkshopRoutes(mux *http.ServeMux, store novelfetchworkshop.Store, aiJudge novelfetchworkshop.AIGenderJudge, fetcher *novelfetchworkshop.UpstreamFetcher, rewriteProvider novelfetchworkshop.RewriteProvider) {
	registry := newWorkshopProcessRegistry()
	mux.HandleFunc("POST /api/novel-fetch-workshop/process", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		var input struct {
			InputText string `json:"inputText"`
			MaxTxt    int    `json:"maxTxt"`
		}
		if err := decodeNovelFetchJSON(w, req, &input); err != nil || strings.TrimSpace(input.InputText) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "请输入要处理的书籍清单"})
			return
		}
		if input.MaxTxt <= 0 {
			input.MaxTxt = 4000
		}
		result, err := workshopProcessResult(req.Context(), store, identity.Username, input.InputText, input.MaxTxt)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "保存任务失败"})
			return
		}
		writeJSON(w, http.StatusOK, result)
	})

	mux.HandleFunc("POST /api/novel-fetch-workshop/process/start", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		var input struct {
			InputText string `json:"inputText"`
			MaxTxt    int    `json:"maxTxt"`
		}
		if err := decodeNovelFetchJSON(w, req, &input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "请求内容无效"})
			return
		}
		now := time.Now().UTC().Format(time.RFC3339Nano)
		id := fmt.Sprintf("go-workshop-%d", time.Now().UnixNano())
		job := map[string]any{"id": id, "status": "running", "queue_state": "running", "started_at": now, "updated_at": now, "attempts": 1, "steps": []any{}, "progress": []any{}}
		result, err := workshopProcessResult(req.Context(), store, identity.Username, input.InputText, input.MaxTxt)
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

	mux.HandleFunc("GET /api/novel-fetch-workshop/tasks", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		tasks, err := store.ListDocuments(req.Context(), identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "获取任务列表失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"tasks": tasks})
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
		document.OriginalRaw = result.OriginalRaw
		document.ProcessedText = result.ProcessedText
		document.Original = result.ProcessedText
		document.OriginalRawChars = result.OriginalRawChars
		document.ProcessedChars = result.ProcessedChars
		document.MaxTxt = result.MaxTxt
		document.BookInfo = result.BookInfo
		if document.Meta == nil {
			document.Meta = map[string]any{}
		}
		document.Meta["originalStatus"] = "done"
		document.Meta["originalRawChars"] = result.OriginalRawChars
		document.Meta["processedChars"] = result.ProcessedChars
		if document.Gender == "" {
			decision := novelfetchworkshop.ResolveGender(req.Context(), "", result.BookInfo, aiJudge)
			document.Gender, document.GenderSource, document.GenderReason = decision.Gender, decision.Source, decision.Reason
		}
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
		document, err := novelfetchworkshop.GenerateAIRewrites(req.Context(), store, identity.Username, bookID, input.Count, rewriteProvider)
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
		if rewriteProvider == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": "AI 改文模型未配置"})
			return
		}
		content, err := rewriteProvider.Complete(req.Context(), novelfetchworkshop.RewriteRequest{
			SystemPrompt: "你是接口连通性测试助手。",
			UserPrompt:   "请只回复：ok",
			Temperature:  0,
			MaxTokens:    8,
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
		explicitManualGender := strings.TrimSpace(document.Gender) != ""
		document.BookID = bookID
		existing, err := store.GetDocument(req.Context(), identity.Username, bookID)
		if err != nil && !errors.Is(err, novelfetchworkshop.ErrNotFound) {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "读取任务失败"})
			return
		}
		document = mergeNovelFetchDocument(existing, document)
		if explicitManualGender {
			decision := novelfetchworkshop.ResolveGender(req.Context(), document.Gender, document.BookInfo, aiJudge)
			document.Gender = decision.Gender
			document.GenderSource = decision.Source
			document.GenderReason = decision.Reason
		}
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
		writeJSON(w, http.StatusOK, map[string]any{"platforms": platforms})
	})

	mux.HandleFunc("GET /api/novel-fetch-workshop/config", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		settings, err := store.GetConfig(req.Context(), identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "读取设置失败"})
			return
		}
		appConfig := settings
		if configured, ok := settings["appConfig"].(map[string]any); ok {
			appConfig = configured
		}
		platforms, ok := settings["platforms"].([]any)
		if !ok {
			platforms = []any{}
		}
		styles, ok := settings["styles"].([]any)
		if !ok {
			styles = []any{}
		}
		aiConfig, _ := settings["aiConfig"].(map[string]any)
		if aiConfig == nil {
			if ai, ok := appConfig["ai"].(map[string]any); ok {
				aiConfig = map[string]any{"ai": ai}
			} else {
				aiConfig = map[string]any{}
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"appConfig": appConfig,
			"platforms": platforms,
			"styles":    styles,
			"aiConfig":  aiConfig,
			"settings":  settings,
		})
	})

	mux.HandleFunc("PUT /api/novel-fetch-workshop/config", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
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
		merged := mergeWorkshopConfig(current, settings)
		if err := store.PutConfig(req.Context(), identity.Username, merged); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "保存设置失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "settings": merged})
	})
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
		meta := map[string]any{"bookId": bookID, "bookName": bookName, "maxTxt": maxTxt}
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
