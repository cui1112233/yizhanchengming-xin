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
		if catalog != nil {
			raw := settings["aiPromptConfig"]
			if len(raw) == 0 {
				raw = json.RawMessage(`{"comic":{"enabled":true}}`)
			}
			hydrated, hydrateErr := hydratePrivateCustomizationComicPromptConfig(catalog, raw)
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

// hydratePrivateCustomizationComicPromptConfig keeps the single-book studio
// independent from Batch Factory presets while still using the backend-owned
// published prompts required by the comic asset and page stages.
func hydratePrivateCustomizationComicPromptConfig(catalog *prompt.Catalog, raw json.RawMessage) (json.RawMessage, error) {
	if catalog == nil {
		return raw, nil
	}
	var config map[string]any
	if err := json.Unmarshal(raw, &config); err != nil || config == nil {
		return nil, fmt.Errorf("漫画提示词配置格式不正确")
	}
	comic, ok := config["comic"].(map[string]any)
	if !ok || comic == nil {
		comic = map[string]any{}
		config["comic"] = comic
	}
	comic["enabled"] = true
	if strings.TrimSpace(stringValue(comic["presetId"])) == "" {
		comic["presetId"] = "batch-comic-page"
	}
	assets, ok := comic["assets"].(map[string]any)
	if !ok || assets == nil {
		assets = map[string]any{}
		comic["assets"] = assets
	}
	if strings.TrimSpace(stringValue(assets["presetId"])) == "" {
		assets["presetId"] = "batch-comic-assets"
	}
	encoded, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("保存漫画提示词配置失败")
	}
	return hydrateBatchFactoryPromptConfig(catalog, encoded)
}

func isPrivateCustomizationComicBatch(batch batchfactoryv11.Batch) bool {
	for _, book := range batch.Books {
		if strings.TrimSpace(book.Platform) == "private_customization" || strings.TrimSpace(stringValue(book.SourceMetadata["sourceMode"])) == "private_customization" {
			return true
		}
	}
	return false
}
