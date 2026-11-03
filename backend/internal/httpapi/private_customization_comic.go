package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/prompt"
)

type privateCustomizationComicInput struct {
	Title                string                        `json:"title"`
	SourceText           string                        `json:"sourceText"`
	InitialBatchSettings batchfactoryv11.SettingsPatch `json:"initialBatchSettings,omitempty"`
}

func registerPrivateCustomizationComicRoute(mux *http.ServeMux, store batchfactoryv11.Store, catalog *prompt.Catalog) {
	mux.HandleFunc("POST /api/batch-factory/v11/private-comic/tasks", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input privateCustomizationComicInput
		if !decodeJSON(w, r, &input) {
			return
		}
		sourceText := strings.TrimSpace(input.SourceText)
		if sourceText == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "sourceText is required"})
			return
		}
		if len([]rune(sourceText)) < 20 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "漫画正文至少需要 20 个字"})
			return
		}
		title := strings.TrimSpace(input.Title)
		if title == "" {
			title = "私人定制漫画"
		}
		settings := input.InitialBatchSettings
		if settings == nil {
			settings = batchfactoryv11.SettingsPatch{}
		}
		settings["productionMode"] = json.RawMessage(`"comic"`)
		if raw, ok := settings["aiPromptConfig"]; ok && catalog != nil {
			hydrated, hydrateErr := hydrateBatchFactoryPromptConfig(catalog, raw)
			if hydrateErr != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": hydrateErr.Error()})
				return
			}
			settings["aiPromptConfig"] = hydrated
		}
		bookID := fmt.Sprintf("private-comic-%d", time.Now().UnixNano())
		intake, err := store.CreateIntake(r.Context(), owner, batchfactoryv11.NovelFetchIntakeInput{
			Books: []batchfactoryv11.CreateBookInput{{
				ID: bookID, BookID: bookID, Title: title, Platform: "private_customization",
				SourceText: sourceText, TxtText: sourceText,
				SourceMetadata: map[string]any{"sourceMode": "private_customization", "queueStatus": "manual_pending"},
			}},
			Metadata: map[string]any{"sourceMode": "private_customization", "title": title},
		})
		if err != nil {
			writeStoreError(w, err)
			return
		}
		batch, err := store.CreateBatchFromIntake(r.Context(), owner, intake.ID, batchfactoryv11.CreateBatchInput{Title: title, InitialBatchSettings: settings})
		if err != nil {
			writeStoreError(w, err)
			return
		}
		if len(batch.Books) != 1 {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "private comic task must contain exactly one book"})
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"batch": batch, "book": batch.Books[0]})
	})
}
