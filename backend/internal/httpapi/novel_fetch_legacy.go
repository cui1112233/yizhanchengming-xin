package httpapi

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"qiantie/backend/internal/novelfetchworkshop"
)

var legacyNovelFetchBookID = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)

var legacyNovelFetchPlatforms = map[int]string{
	1: "黑岩付费", 2: "番茄付费", 3: "七猫付费", 4: "点众付费", 7: "番茄免费",
	15: "知乎付费", 20: "掌阅付费", 26: "卓越付费", 29: "九州书城", 31: "掌文付费",
}

func writeLegacyNovelFetchFile(root, bookID, text string) bool {
	if strings.TrimSpace(root) == "" || !legacyNovelFetchBookID.MatchString(bookID) {
		return false
	}
	dir := filepath.Join(root, storageNovelFetch)
	if os.MkdirAll(dir, 0o755) != nil {
		return false
	}
	return os.WriteFile(filepath.Join(dir, bookID+".txt"), []byte(text), 0o600) == nil
}

func RegisterLegacyNovelFetchRoutes(root *http.ServeMux, middleware BridgeAuth, store novelfetchworkshop.Store, fetcher *novelfetchworkshop.UpstreamFetcher, configs *UserConfigStore) {
	if root == nil || store == nil {
		return
	}
	authHandler := func(next http.Handler) http.Handler {
		return middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if _, ok := BridgeIdentityFromContext(req.Context()); !ok {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}
			next.ServeHTTP(w, req)
		}))
	}
	root.Handle("POST /api/novel-fetch", authHandler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		var input struct {
			Platform     any      `json:"platform"`
			BookIDs      []string `json:"bookIds"`
			MaxTxt       int      `json:"maxTxt"`
			SaveToFolder bool     `json:"saveToFolder"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 64<<10)).Decode(&input) != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求内容无效"})
			return
		}
		platformText := strings.TrimSpace(strconv.FormatInt(int64(toLegacyPlatform(input.Platform)), 10))
		platformID := toLegacyPlatform(input.Platform)
		if _, ok := legacyNovelFetchPlatforms[platformID]; !ok {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "无效的平台 ID"})
			return
		}
		if len(input.BookIDs) == 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请提供书籍 ID 列表"})
			return
		}
		if len(input.BookIDs) > 50 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "一次最多获取 50 本书"})
			return
		}
		maxTxt := input.MaxTxt
		if maxTxt < 100 || maxTxt > 100000 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "字数需为 100–100000 的整数"})
			return
		}
		if fetcher == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "原文抓取服务未配置"})
			return
		}
		storageRoot := ""
		if input.SaveToFolder {
			storageRoot = configuredStorageRoot(configs, identity.Username)
		}
		results := make([]map[string]any, 0, len(input.BookIDs))
		seen := map[string]bool{}
		for _, rawID := range input.BookIDs {
			bookID := strings.TrimSpace(rawID)
			if bookID == "" || seen[bookID] {
				continue
			}
			seen[bookID] = true
			base := map[string]any{"bookId": bookID, "platform": platformID, "platformName": legacyNovelFetchPlatforms[platformID]}
			if !legacyNovelFetchBookID.MatchString(bookID) {
				base["status"], base["data"], base["error"], base["length"] = "error", nil, "书籍 ID 格式不正确："+bookID, 0
				results = append(results, base)
				continue
			}
			fetched, err := fetcher.FetchAndProcess(req.Context(), bookID, platformText, maxTxt)
			if err != nil {
				base["status"], base["data"], base["error"], base["length"] = "error", nil, err.Error(), 0
				results = append(results, base)
				continue
			}
			document := novelfetchworkshop.Document{BookID: bookID, OriginalRaw: fetched.OriginalRaw, Original: fetched.ProcessedText, ProcessedText: fetched.ProcessedText, OriginalRawChars: fetched.OriginalRawChars, ProcessedChars: fetched.ProcessedChars, MaxTxt: fetched.MaxTxt, BookInfo: fetched.BookInfo, Meta: map[string]any{"bookId": bookID, "platformId": platformText, "platformName": legacyNovelFetchPlatforms[platformID], "originalStatus": "done", "originalRawChars": fetched.OriginalRawChars, "processedChars": fetched.ProcessedChars, "maxTxt": fetched.MaxTxt, "updatedAt": time.Now().UTC().Format(time.RFC3339)}}
			if err := persistNovelFetchBodies(req.Context(), store, identity.Username, &document, fetched.OriginalRaw, fetched.ProcessedText); err != nil {
				base["status"], base["data"], base["error"], base["length"] = "error", nil, "保存原文正文失败", 0
				results = append(results, base)
				continue
			}
			if err := store.PutDocument(req.Context(), identity.Username, document); err != nil {
				base["status"], base["data"], base["error"], base["length"] = "error", nil, "保存原文失败", 0
				results = append(results, base)
				continue
			}
			saved := input.SaveToFolder && writeLegacyNovelFetchFile(storageRoot, bookID, fetched.ProcessedText)
			base["status"], base["data"], base["error"], base["length"], base["savedToFolder"] = "ok", fetched.ProcessedText, nil, fetched.ProcessedChars, saved
			results = append(results, base)
		}
		writeJSON(w, http.StatusOK, map[string]any{"results": results})
	})))
	root.Handle("GET /api/novel-fetch/saved", authHandler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		documents, err := store.ListDocuments(req.Context(), identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "读取已保存小说失败"})
			return
		}
		books := make([]map[string]any, 0, len(documents))
		for _, document := range documents {
			books = append(books, map[string]any{"bookId": document.BookID, "text": document.ProcessedText, "original": document.Original, "meta": document.Meta, "updatedAt": document.UpdatedAt})
		}
		writeJSON(w, http.StatusOK, map[string]any{"books": books})
	})))
	root.Handle("POST /api/novel-fetch/save", authHandler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		var input struct {
			BookID string         `json:"bookId"`
			Text   string         `json:"text"`
			Meta   map[string]any `json:"meta"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 256<<10)).Decode(&input) != nil || !legacyNovelFetchBookID.MatchString(strings.TrimSpace(input.BookID)) || strings.TrimSpace(input.Text) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "书籍 ID 或正文无效"})
			return
		}
		document, err := store.GetDocument(req.Context(), identity.Username, input.BookID)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "该书尚未处理，无法保存"})
			return
		}
		document.Original, document.ProcessedText = input.Text, input.Text
		document.ProcessedChars = len([]rune(input.Text))
		if input.Meta != nil {
			document.Meta = mergeConfigMap(document.Meta, input.Meta)
		}
		if err := store.PutDocument(req.Context(), identity.Username, document); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "保存失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "savedAt": time.Now().UTC().Format(time.RFC3339)})
	})))
}

func toLegacyPlatform(value any) int {
	switch typed := value.(type) {
	case float64:
		return int(typed)
	case string:
		parsed, _ := strconv.Atoi(strings.TrimSpace(typed))
		return parsed
	default:
		return 0
	}
}
