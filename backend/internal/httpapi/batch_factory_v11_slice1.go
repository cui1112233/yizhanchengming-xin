package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/prompt"
)

type batchRuntimeIndexReader interface {
	GetBatchRuntimeIndex(context.Context, string, string) (batchfactoryv11.Batch, error)
}

type batchRecoveryIndexLister interface {
	ListBatchRecoveryIndex(context.Context, string) ([]batchfactoryv11.Batch, error)
}

type batchSummaryIndexLister interface {
	ListBatchSummaryIndex(context.Context, string) ([]batchfactoryv11.Batch, error)
}

func deleteRetentionDays(r *http.Request) int {
	value, err := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("retentionDays")))
	if err != nil {
		return batchfactoryv11.DefaultLocalMergedArtifactRetentionDays
	}
	switch value {
	case 3, 7, 14, 30:
		return value
	default:
		return batchfactoryv11.DefaultLocalMergedArtifactRetentionDays
	}
}

func deleteBookWithRetention(ctx context.Context, store batchfactoryv11.Store, owner, batchID, bookID string, days int) error {
	if retentionStore, ok := store.(batchfactoryv11.RetentionDeletingStore); ok {
		return retentionStore.DeleteBookWithRetention(ctx, owner, batchID, bookID, days)
	}
	return store.DeleteBook(ctx, owner, batchID, bookID)
}

func deleteBatchWithRetention(ctx context.Context, store batchfactoryv11.Store, owner, batchID string, days int) error {
	if retentionStore, ok := store.(batchfactoryv11.RetentionDeletingStore); ok {
		return retentionStore.DeleteBatchWithRetention(ctx, owner, batchID, days)
	}
	return store.DeleteBatch(ctx, owner, batchID)
}

func registerSliceOneRoutes(mux *http.ServeMux, store batchfactoryv11.Store, catalog *prompt.Catalog) {
	mux.HandleFunc("POST /api/batch-factory/v11/intakes/manual", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input batchfactoryv11.ManualIntakeInput
		if !decodeJSON(w, r, &input) {
			return
		}
		if input.ContentRangeLines <= 0 {
			input.ContentRangeLines = 5
		}
		if input.ContentRangeLines > 500 {
			writeStoreError(w, batchfactoryv11.ErrInvalid)
			return
		}
		if input.ComicAnalysisCharacters <= 0 {
			input.ComicAnalysisCharacters = 2000
		}
		if input.ComicAnalysisCharacters < 100 || input.ComicAnalysisCharacters > 100000 {
			writeStoreError(w, batchfactoryv11.ErrInvalid)
			return
		}
		if input.ContentCaptureCharacters <= 0 {
			input.ContentCaptureCharacters = 4000
		}
		if input.ContentCaptureCharacters < 100 || input.ContentCaptureCharacters > 100000 {
			writeStoreError(w, batchfactoryv11.ErrInvalid)
			return
		}
		queueStatus := "manual_pending"
		if strings.TrimSpace(input.ScheduledAt) != "" {
			if _, err := time.Parse(time.RFC3339, input.ScheduledAt); err != nil {
				writeStoreError(w, batchfactoryv11.ErrInvalid)
				return
			}
			queueStatus = "scheduled_waiting"
		}
		grouped := len(input.Groups) > 0
		var books []batchfactoryv11.CreateBookInput
		var err error
		if grouped {
			books, err = batchfactoryv11.ParseGroupedManualBookLists(input)
		} else {
			books, err = batchfactoryv11.ParseManualBookList(input)
		}
		if err != nil {
			writeStoreError(w, err)
			return
		}
		if len(books) > 50 {
			writeStoreError(w, batchfactoryv11.ErrInvalid)
			return
		}
		for index := range books {
			if books[index].SourceMetadata == nil {
				books[index].SourceMetadata = map[string]any{}
			}
			books[index].SourceMetadata["contentRangeLines"] = input.ContentRangeLines
			books[index].SourceMetadata["comicAnalysisCharacters"] = input.ComicAnalysisCharacters
			books[index].SourceMetadata["contentCaptureCharacters"] = input.ContentCaptureCharacters
			books[index].SourceMetadata["scheduledAt"] = strings.TrimSpace(input.ScheduledAt)
			books[index].SourceMetadata["queueStatus"] = queueStatus
			if !grouped {
				books[index].SourceMetadata["platformName"] = strings.TrimSpace(input.PlatformName)
			}
		}
		intake, err := store.CreateIntake(r.Context(), owner, batchfactoryv11.NovelFetchIntakeInput{
			Books: books,
			Metadata: map[string]any{
				"sourceMode": "manual_original", "platformId": input.PlatformID, "platformName": strings.TrimSpace(input.PlatformName), "parseMode": input.ParseMode,
				"columnPresetId": input.ColumnPresetID, "columnOrder": input.ColumnOrder, "inputText": input.InputText,
				"contentRangeLines": input.ContentRangeLines, "comicAnalysisCharacters": input.ComicAnalysisCharacters, "contentCaptureCharacters": input.ContentCaptureCharacters, "scheduledAt": strings.TrimSpace(input.ScheduledAt), "queueStatus": queueStatus,
				"groupCount": len(input.Groups),
			},
		})
		if err != nil {
			writeStoreError(w, err)
			return
		}
		batch, err := store.CreateBatchFromIntake(r.Context(), owner, intake.ID, batchfactoryv11.CreateBatchInput{Title: input.Title, InitialBatchSettings: input.InitialBatchSettings})
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"batch": batch, "intake": intake})
	})
	mux.HandleFunc("POST /api/batch-factory/v11/intakes/novel-fetch", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input batchfactoryv11.NovelFetchIntakeInput
		if !decodeJSON(w, r, &input) {
			return
		}
		intake, err := store.CreateIntake(r.Context(), owner, input)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"intake": intake})
	})
	mux.HandleFunc("GET /api/batch-factory/v11/intakes/{intakeId}", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		intake, err := store.GetIntake(r.Context(), owner, r.PathValue("intakeId"))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"intake": intake})
	})
	mux.HandleFunc("POST /api/batch-factory/v11/intakes/{intakeId}/batches", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input batchfactoryv11.CreateBatchInput
		if !decodeJSON(w, r, &input) {
			return
		}
		batch, err := store.CreateBatchFromIntake(r.Context(), owner, r.PathValue("intakeId"), input)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"batch": batch})
	})
	mux.HandleFunc("POST /api/batch-factory/v11/batches/{batchId}/intakes/{intakeId}/books", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input struct {
			AllowDuplicate bool `json:"allowDuplicate"`
		}
		if !decodeJSON(w, r, &input) {
			return
		}
		batch, err := store.AppendBooksFromIntake(r.Context(), owner, r.PathValue("batchId"), r.PathValue("intakeId"), input.AllowDuplicate)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"batch": batch})
	})
	mux.HandleFunc("GET /api/batch-factory/v11/batches", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		batches, err := store.ListBatches(r.Context(), owner)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"batches": batches})
	})
	mux.HandleFunc("GET /api/batch-factory/v11/batches/summary-index", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var batches []batchfactoryv11.Batch
		var err error
		if lister, ok := store.(batchSummaryIndexLister); ok {
			batches, err = lister.ListBatchSummaryIndex(r.Context(), owner)
		} else {
			batches, err = store.ListBatches(r.Context(), owner)
		}
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"batches": batches})
	})
	mux.HandleFunc("POST /api/batch-factory/v11/batches", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input batchfactoryv11.CreateBatchInput
		if !decodeJSON(w, r, &input) {
			return
		}
		batch, err := store.CreateBatch(r.Context(), owner, input)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"batch": batch})
	})
	mux.HandleFunc("GET /api/batch-factory/v11/batches/recovery-index", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var batches []batchfactoryv11.Batch
		var err error
		if lister, ok := store.(batchRecoveryIndexLister); ok {
			batches, err = lister.ListBatchRecoveryIndex(r.Context(), owner)
		} else {
			batches, err = store.ListBatches(r.Context(), owner)
		}
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"batches": batches})
	})
	mux.HandleFunc("GET /api/batch-factory/v11/batches/{batchId}/runtime-index", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var batch batchfactoryv11.Batch
		var err error
		if reader, ok := store.(batchRuntimeIndexReader); ok {
			batch, err = reader.GetBatchRuntimeIndex(r.Context(), owner, r.PathValue("batchId"))
		} else {
			batch, err = store.GetBatch(r.Context(), owner, r.PathValue("batchId"))
		}
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"batch": batch})
	})
	mux.HandleFunc("GET /api/batch-factory/v11/batches/{batchId}", func(w http.ResponseWriter, r *http.Request) {
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
		writeJSON(w, http.StatusOK, map[string]any{"batch": batch})
	})
	mux.HandleFunc("DELETE /api/batch-factory/v11/batches/{batchId}/books/{bookId}", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		if err := deleteBookWithRetention(r.Context(), store, owner, r.PathValue("batchId"), r.PathValue("bookId"), deleteRetentionDays(r)); err != nil {
			writeStoreError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("DELETE /api/batch-factory/v11/batches/{batchId}", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		if err := deleteBatchWithRetention(r.Context(), store, owner, r.PathValue("batchId"), deleteRetentionDays(r)); err != nil {
			writeStoreError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("PUT /api/batch-factory/v11/batches/{batchId}/settings", settingsHandler(store, catalog, batchfactoryv11.ScopeBatch))
	mux.HandleFunc("PUT /api/batch-factory/v11/batches/{batchId}/books/{bookId}/metadata", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input struct {
			Metadata         map[string]any `json:"metadata"`
			MetadataPatch    map[string]any `json:"metadataPatch,omitempty"`
			ExpectedRevision int64          `json:"expectedRevision"`
		}
		if !decodeJSON(w, r, &input) {
			return
		}
		if input.MetadataPatch != nil {
			batch, err := store.GetBatch(r.Context(), owner, r.PathValue("batchId"))
			if err != nil {
				writeStoreError(w, err)
				return
			}
			var current *batchfactoryv11.Book
			for index := range batch.Books {
				if batch.Books[index].ID == r.PathValue("bookId") {
					current = &batch.Books[index]
					break
				}
			}
			if current == nil {
				writeStoreError(w, batchfactoryv11.ErrNotFound)
				return
			}
			input.Metadata = make(map[string]any, len(current.SourceMetadata)+len(input.MetadataPatch))
			for key, value := range current.SourceMetadata {
				input.Metadata[key] = value
			}
			for key, value := range input.MetadataPatch {
				input.Metadata[key] = value
			}
		}
		book, err := store.UpdateBookMetadata(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"), batchfactoryv11.UpdateBookMetadataInput{
			Metadata:         input.Metadata,
			ExpectedRevision: input.ExpectedRevision,
		})
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"book": book})
	})
	mux.HandleFunc("PUT /api/batch-factory/v11/batches/{batchId}/books/{bookId}/source", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input batchfactoryv11.CaptureBookSourceInput
		if !decodeJSON(w, r, &input) {
			return
		}
		book, err := store.CaptureBookSource(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"), input)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"book": book})
	})
	mux.HandleFunc("PUT /api/batch-factory/v11/batches/{batchId}/books/{bookId}/override", settingsHandler(store, catalog, batchfactoryv11.ScopeBook))
	mux.HandleFunc("PUT /api/batch-factory/v11/batches/{batchId}/books/{bookId}/videos/{videoId}/override", settingsHandler(store, catalog, batchfactoryv11.ScopeVideo))
	mux.HandleFunc("POST /api/batch-factory/v11/batches/{batchId}/change-impact", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input batchfactoryv11.SettingsUpdate
		if !decodeJSON(w, r, &input) {
			return
		}
		impact, err := store.ChangeImpact(r.Context(), owner, r.PathValue("batchId"), input)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, impact)
	})
	mux.HandleFunc("GET /api/batch-factory/v11/config-versions", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		versions, err := store.ConfigVersions(r.Context(), owner)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"configVersions": versions})
	})
	mux.HandleFunc("POST /api/batch-factory/v11/config-versions", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input batchfactoryv11.ConfigVersion
		if !decodeJSON(w, r, &input) {
			return
		}
		version, err := store.CreateConfigVersion(r.Context(), owner, input)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"configVersion": version})
	})
	mux.HandleFunc("PUT /api/batch-factory/v11/config-versions/{versionId}", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input batchfactoryv11.RenameConfigVersionInput
		if !decodeJSON(w, r, &input) {
			return
		}
		version, err := store.RenameConfigVersion(r.Context(), owner, r.PathValue("versionId"), input.Name)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"configVersion": version})
	})
	mux.HandleFunc("GET /api/batch-factory/v11/prompts", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		prompts, err := store.ListPrompts(r.Context(), owner, r.URL.Query().Get("kind"))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"prompts": prompts})
	})
	mux.HandleFunc("POST /api/batch-factory/v11/prompts", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input batchfactoryv11.Prompt
		if !decodeJSON(w, r, &input) {
			return
		}
		if strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.Content) == "" {
			writeStoreError(w, batchfactoryv11.ErrInvalid)
			return
		}
		prompt, err := store.CreatePrompt(r.Context(), owner, input)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"prompt": prompt})
	})
	mux.HandleFunc("GET /api/batch-factory/v11/drafts", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		q := r.URL.Query()
		draft, err := store.GetDraft(r.Context(), owner, q.Get("key"), q.Get("kind"), q.Get("scope"))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"draft": draft})
	})
	mux.HandleFunc("PUT /api/batch-factory/v11/drafts", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input batchfactoryv11.Draft
		if !decodeJSON(w, r, &input) {
			return
		}
		if strings.TrimSpace(input.Key) == "" {
			writeStoreError(w, batchfactoryv11.ErrInvalid)
			return
		}
		draft, err := store.SaveDraft(r.Context(), owner, input)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"draft": draft})
	})
}

func settingsHandler(store batchfactoryv11.Store, catalog *prompt.Catalog, kind batchfactoryv11.ScopeKind) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input batchfactoryv11.SettingsUpdate
		if !decodeJSON(w, r, &input) {
			return
		}
		if raw, exists := input.Patch["aiPromptConfig"]; exists {
			hydrated, err := hydrateBatchFactoryPromptConfig(catalog, raw)
			if err != nil {
				writeStoreError(w, fmt.Errorf("%w: %v", batchfactoryv11.ErrInvalid, err))
				return
			}
			input.Patch["aiPromptConfig"] = hydrated
		}
		if raw, exists := input.Patch["productionMode"]; exists {
			var mode string
			if err := json.Unmarshal(raw, &mode); err != nil {
				writeStoreError(w, fmt.Errorf("%w: 生产模式格式不正确", batchfactoryv11.ErrInvalid))
				return
			}
			normalized, err := normalizeBatchFactoryProductionMode(mode)
			if err != nil {
				writeStoreError(w, fmt.Errorf("%w: %v", batchfactoryv11.ErrInvalid, err))
				return
			}
			encoded, _ := json.Marshal(normalized)
			input.Patch["productionMode"] = encoded
		}
		ref := batchfactoryv11.ScopeRef{Kind: kind, BatchID: r.PathValue("batchId"), BookID: r.PathValue("bookId"), VideoID: r.PathValue("videoId")}
		result, err := store.SaveSettings(r.Context(), owner, ref, input)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, result)
	}
}

func bridgeOwner(r *http.Request) (string, bool) {
	identity, ok := BridgeIdentityFromContext(r.Context())
	return identity.Username, ok && identity.Username != ""
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	if r.Body == nil {
		return true
	}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(target); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return false
	}
	return true
}

func writeStoreError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, batchfactoryv11.ErrNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
	case errors.Is(err, batchfactoryv11.ErrConflict):
		message := strings.TrimSpace(err.Error())
		if message == "" || message == batchfactoryv11.ErrConflict.Error() {
			message = "revision conflict"
		}
		writeJSON(w, http.StatusConflict, map[string]string{"error": message})
	case errors.Is(err, batchfactoryv11.ErrInvalid):
		message := strings.TrimSpace(err.Error())
		if message == "" || message == batchfactoryv11.ErrInvalid.Error() {
			message = "invalid input"
		}
		// Keep validation failures actionable without returning an unbounded
		// provider/config payload to the browser.
		if len([]rune(message)) > 240 {
			runes := []rune(message)
			message = string(runes[:240]) + "..."
		}
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": message})
	case errors.Is(err, batchfactoryv11.ErrUnavailable):
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "capability unavailable"})
	case errors.Is(err, batchfactoryv11.ErrUpstream):
		message := strings.TrimSpace(err.Error())
		if message == "" || message == batchfactoryv11.ErrUpstream.Error() {
			message = "temporary upstream failure"
		}
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": message})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
	}
}
