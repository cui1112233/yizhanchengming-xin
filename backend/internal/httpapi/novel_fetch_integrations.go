package httpapi

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"qiantie/backend/internal/novelfetchworkshop"
	"qiantie/backend/internal/safehttp"
)

// 参考实现暴露的三个集成接口：
//   POST /api/integrations/workbench/push
//   POST /api/integrations/changdu/bookid-fetch/run
//   GET  /api/integrations/changdu/bookid-fetch/status
// 这里复用工作台现有的建任务 + 入队逻辑，不单独维护第二套处理管线。

type novelFetchBookIDFetchState struct {
	mu            sync.Mutex
	lastRunAt     string
	lastStatus    string
	lastError     string
	lastFetched   int
	lastAccepted  int
	lastSkipped   int
	lastFailed    int
	lastSourceURL string
}

var novelFetchBookIDFetchStateStore = &novelFetchBookIDFetchState{}

type novelFetchIntegrationBook struct {
	BookID    string `json:"book_id"`
	BookIDAlt string `json:"bookId"`
	BookName  string `json:"book_name"`
	Name      string `json:"name"`
	Author    string `json:"author"`
	Category  string `json:"category"`
}

type novelFetchWorkbenchPushRequest struct {
	Source       string                      `json:"source"`
	SourceRunID  string                      `json:"source_run_id"`
	SourceRunAlt string                      `json:"sourceRunId"`
	TriggerMode  string                      `json:"trigger_mode"`
	PlatformID   string                      `json:"platform_id"`
	PlatformAlt  string                      `json:"platformId"`
	Books        []novelFetchIntegrationBook `json:"books"`
}

type novelFetchBookIDFetchRunRequest struct {
	URL         string `json:"url"`
	PlatformID  string `json:"platform_id"`
	PlatformAlt string `json:"platformId"`
	InputText   string `json:"input_text"`
	InputAlt    string `json:"inputText"`
	Force       bool   `json:"force"`
}

// integrationPlatformID 与参考实现一致：未指定时默认平台 2。
func integrationPlatformID(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return "2"
}

// integrationCreateDocuments 只创建任务元数据；真正入队交给 enqueueWorkshopRetries，
// 与工作台重试走同一条经过测试的处理链路。
func integrationCreateDocuments(ctx context.Context, store novelfetchworkshop.Store, owner string, books []novelFetchIntegrationBook, platformID string, maxTxt int) (created []string, duplicates int, failed int) {
	seen := map[string]bool{}
	for _, book := range books {
		bookID := novelFetchSafeBookID(firstText(book.BookID, book.BookIDAlt))
		if bookID == "" || !novelFetchBookIDPattern.MatchString(bookID) {
			failed++
			continue
		}
		if seen[bookID] {
			duplicates++
			continue
		}
		seen[bookID] = true
		if existing, err := store.GetDocument(ctx, owner, bookID); err == nil && existing.BookID != "" {
			duplicates++
			continue
		}
		bookName := strings.TrimSpace(firstText(book.BookName, book.Name))
		meta := map[string]any{
			"platformId":   platformID,
			"bookName":     bookName,
			"name":         bookName,
			"importSource": "integration",
			"importedAt":   time.Now().UTC().Format(time.RFC3339Nano),
			"status":       "queued",
			"queueState":   "queued",
		}
		if author := strings.TrimSpace(book.Author); author != "" {
			meta["author"] = author
		}
		if category := strings.TrimSpace(book.Category); category != "" {
			meta["category"] = category
		}
		document := novelfetchworkshop.Document{BookID: bookID, Meta: meta, MaxTxt: maxTxt, UpdatedAt: time.Now().UTC().Format(time.RFC3339Nano)}
		if err := store.PutDocument(ctx, owner, document); err != nil {
			failed++
			continue
		}
		created = append(created, bookID)
	}
	return created, duplicates, failed
}

func RegisterNovelFetchIntegrations(root *http.ServeMux, middleware NovelFetchBridgeAuth, store novelfetchworkshop.Store, queue novelfetchworkshop.ProcessQueue) {
	if root == nil || store == nil {
		return
	}

	pushHandler := middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"ok": false, "error": "unauthorized"})
			return
		}
		var input novelFetchWorkbenchPushRequest
		if err := decodeNovelFetchJSON(w, req, &input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "请求内容无效"})
			return
		}
		settings := mustGetNovelFetchConfig(req, store, identity.Username)
		if len(input.Books) == 0 {
			writeJSON(w, http.StatusOK, map[string]any{"ok": true, "status": "skipped", "accepted": 0, "created": 0, "skipped_duplicate": 0, "failed": 0, "results": []any{}, "auto_queue_enabled": novelFetchAutomationEnabled(settings)})
			return
		}
		platformID := integrationPlatformID(input.PlatformID, input.PlatformAlt, toString(mapOrEmpty(settings["fetch"])["default_platform_id"]))
		maxTxt := intValue(mapOrEmpty(settings["fetch"])["default_max_txt"])
		if maxTxt <= 0 {
			maxTxt = 4000
		}
		created, duplicates, failed := integrationCreateDocuments(req.Context(), store, identity.Username, input.Books, platformID, maxTxt)
		queued := 0
		if len(created) > 0 {
			if queue == nil {
				writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "status": "failed", "error": "Redis 任务队列未配置", "created": len(created)})
				return
			}
			result, err := enqueueWorkshopRetries(req.Context(), store, queue, identity.Username, created, false)
			if err != nil {
				writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "status": "failed", "error": err.Error(), "created": len(created), "failed": result.Failed})
				return
			}
			queued = result.Queued
			failed += result.Failed
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"ok": true, "status": "ok", "job_id": "", "source": firstText(input.Source, "changdu-workbench"),
			"accepted": queued, "created": len(created), "skipped_duplicate": duplicates, "failed": failed,
			"results": []any{}, "poll_url": "/api/process/jobs/latest",
			"auto_queue_enabled": novelFetchAutomationEnabled(settings),
		})
	}))

	bookIDFetchHandler := middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"ok": false, "error": "unauthorized"})
			return
		}
		var input novelFetchBookIDFetchRunRequest
		if err := decodeNovelFetchJSON(w, req, &input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "请求内容无效"})
			return
		}
		settings := mustGetNovelFetchConfig(req, store, identity.Username)
		sourceURL := strings.TrimSpace(input.URL)
		if sourceURL == "" {
			sourceURL = strings.TrimSpace(toString(mapOrEmpty(settings["integrations"])["bookid_fetch_url"]))
		}
		text := strings.TrimSpace(firstText(input.InputText, input.InputAlt))
		if text == "" && sourceURL == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "status": "disabled", "error": "未配置书城 ID 接口地址", "accepted": 0, "skipped_duplicate": 0, "failed": 0, "results": []any{}})
			return
		}
		novelFetchBookIDFetchStateStore.mu.Lock()
		novelFetchBookIDFetchStateStore.lastRunAt = time.Now().UTC().Format(time.RFC3339Nano)
		novelFetchBookIDFetchStateStore.lastSourceURL = sourceURL
		novelFetchBookIDFetchStateStore.mu.Unlock()
		if text == "" {
			if err := safehttp.CheckURL(sourceURL); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "status": "failed", "error": "书城 ID 接口地址不可用：" + err.Error()})
				return
			}
			client := safehttp.NewClientWithTimeout(30 * time.Second)
			response, err := client.Get(sourceURL)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "status": "failed", "error": "拉取书城 ID 失败：" + err.Error()})
				return
			}
			defer response.Body.Close()
			body, err := io.ReadAll(io.LimitReader(response.Body, 5<<20))
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "status": "failed", "error": "读取书城 ID 响应失败"})
				return
			}
			text = string(body)
		}
		rawIDs := novelFetchNumericIDPattern.FindAllString(text, -1)
		books := make([]novelFetchIntegrationBook, 0, len(rawIDs))
		seen := map[string]bool{}
		for _, id := range rawIDs {
			if seen[id] {
				continue
			}
			seen[id] = true
			books = append(books, novelFetchIntegrationBook{BookID: id})
		}
		platformID := integrationPlatformID(input.PlatformID, input.PlatformAlt, toString(mapOrEmpty(settings["fetch"])["default_platform_id"]))
		maxTxt := intValue(mapOrEmpty(settings["fetch"])["default_max_txt"])
		if maxTxt <= 0 {
			maxTxt = 4000
		}
		created, duplicates, failed := integrationCreateDocuments(req.Context(), store, identity.Username, books, platformID, maxTxt)
		queued := 0
		if len(created) > 0 {
			if queue == nil {
				writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "status": "failed", "error": "Redis 任务队列未配置", "accepted": len(created)})
				return
			}
			result, err := enqueueWorkshopRetries(req.Context(), store, queue, identity.Username, created, false)
			if err != nil {
				writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "status": "failed", "error": err.Error(), "accepted": result.Queued, "failed": result.Failed})
				return
			}
			queued = result.Queued
			failed += result.Failed
		}
		skipped := duplicates + len(rawIDs) - len(books)
		if skipped < 0 {
			skipped = duplicates
		}
		novelFetchBookIDFetchStateStore.mu.Lock()
		novelFetchBookIDFetchStateStore.lastStatus = "ok"
		novelFetchBookIDFetchStateStore.lastError = ""
		novelFetchBookIDFetchStateStore.lastFetched = len(rawIDs)
		novelFetchBookIDFetchStateStore.lastAccepted = queued
		novelFetchBookIDFetchStateStore.lastSkipped = skipped
		novelFetchBookIDFetchStateStore.lastFailed = failed
		novelFetchBookIDFetchStateStore.mu.Unlock()
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "status": "ok", "fetched": len(rawIDs), "unique": len(books), "accepted": queued, "skipped_duplicate": skipped, "failed": failed, "results": []any{}, "source": sourceURL})
	}))

	statusHandler := middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"ok": false, "error": "unauthorized"})
			return
		}
		settings := mustGetNovelFetchConfig(req, store, identity.Username)
		enabled := strings.TrimSpace(toString(mapOrEmpty(settings["integrations"])["bookid_fetch_url"])) != ""
		novelFetchBookIDFetchStateStore.mu.Lock()
		payload := map[string]any{
			"ok": true, "enabled": enabled, "running": false,
			"last_run_at":            novelFetchBookIDFetchStateStore.lastRunAt,
			"last_status":            novelFetchBookIDFetchStateStore.lastStatus,
			"last_error":             novelFetchBookIDFetchStateStore.lastError,
			"last_fetched":           novelFetchBookIDFetchStateStore.lastFetched,
			"last_accepted":          novelFetchBookIDFetchStateStore.lastAccepted,
			"last_skipped_duplicate": novelFetchBookIDFetchStateStore.lastSkipped,
			"last_failed":            novelFetchBookIDFetchStateStore.lastFailed,
			"source_url":             novelFetchBookIDFetchStateStore.lastSourceURL,
		}
		novelFetchBookIDFetchStateStore.mu.Unlock()
		writeJSON(w, http.StatusOK, payload)
	}))

	root.Handle("POST /api/integrations/workbench/push", pushHandler)
	root.Handle("POST /api/integrations/changdu/bookid-fetch/run", bookIDFetchHandler)
	root.Handle("GET /api/integrations/changdu/bookid-fetch/status", statusHandler)
}

func mustGetNovelFetchConfig(req *http.Request, store novelfetchworkshop.Store, owner string) map[string]any {
	settings, err := store.GetConfig(req.Context(), owner)
	if err != nil || settings == nil {
		return map[string]any{}
	}
	return settings
}
