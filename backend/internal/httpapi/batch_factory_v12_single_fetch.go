package httpapi

import (
	"net/http"
	"strconv"
	"strings"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/novelfetchworkshop"
)

func registerV12SingleFetchRoute(mux *http.ServeMux, store batchfactoryv11.Store, fetcher *novelfetchworkshop.UpstreamFetcher) {
	if mux == nil || store == nil || fetcher == nil {
		return
	}
	mux.HandleFunc("POST /api/batch-factory/v12/batches/{batchId}/books/{bookId}/fetch-original", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		batch, err := store.GetBatch(r.Context(), owner, r.PathValue("batchId"))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		var book *batchfactoryv11.Book
		for index := range batch.Books {
			if batch.Books[index].ID == r.PathValue("bookId") {
				book = &batch.Books[index]
				break
			}
		}
		if book == nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "小说不存在"})
			return
		}
		maxTxt := 4000
		var input struct {
			MaxTxt   int    `json:"maxTxt"`
			Platform string `json:"platform"`
		}
		if r.ContentLength != 0 && !decodeJSON(w, r, &input) {
			return
		}
		if input.MaxTxt != 0 {
			maxTxt = input.MaxTxt
		}
		if maxTxt < 100 || maxTxt > 100000 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "字数参数无效"})
			return
		}
		platform := strings.TrimSpace(input.Platform)
		if platform == "" {
			platform = strings.TrimSpace(book.Platform)
		}
		if platform == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "缺少平台参数"})
			return
		}
		bookID := strings.TrimSpace(book.BookID)
		if bookID == "" {
			bookID = strings.TrimSpace(book.ID)
		}
		if bookID == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "缺少小说编号"})
			return
		}
		result, err := fetcher.FetchAndProcess(r.Context(), bookID, platform, maxTxt)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		metadata := map[string]any{}
		if result.BookInfo.Category != "" {
			metadata["category"] = result.BookInfo.Category
		}
		if result.BookInfo.Genre != "" {
			metadata["genre"] = result.BookInfo.Genre
		}
		decision := novelfetchworkshop.ResolveGender(r.Context(), "", result.BookInfo, nil)
		if decision.Gender == "男频" || decision.Gender == "女频" {
			metadata["gender"], metadata["genderSource"] = decision.Gender, string(decision.Source)
		}
		updated, err := store.CaptureBookSource(r.Context(), owner, batch.ID, book.ID, batchfactoryv11.CaptureBookSourceInput{SourceText: result.ProcessedText, ReplaceSource: true, SourceTitle: book.Title, SourceMetadata: metadata, ExpectedRevision: book.Revision})
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "status": "done", "book": updated, "rawData": result.OriginalRaw, "length": result.ProcessedChars, "originalLength": result.OriginalRawChars, "maxTxt": result.MaxTxt, "platform": parsePlatformNumber(platform), "bookinfo": result.BookInfo, "sourceMetadata": metadata})
	})
}

func parsePlatformNumber(value string) any {
	if number, err := strconv.Atoi(value); err == nil {
		return number
	}
	return value
}
