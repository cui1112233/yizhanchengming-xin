package httpapi

import (
	"net/http"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/localartifact"
	"qiantie/backend/internal/novelfetchworkshop"
)

// registerV12AssetGenerationRoute keeps the asset image action on the same Go
// store/provider boundary as the Shuihuo project page. The old Node handler
// used to receive this one action even when every other batch operation was on
// Go, which made the asset step the last hidden runtime split.
func registerV12AssetGenerationRoute(mux *http.ServeMux, store batchfactoryv11.Store, files *localartifact.Store, catalogStore novelfetchworkshop.Store) {
	if mux == nil || store == nil || files == nil {
		return
	}
	mux.HandleFunc("POST /api/batch-factory/v12/batches/{batchId}/books/{bookId}/assets/images/generate", func(w http.ResponseWriter, r *http.Request) {
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
		bookID := r.PathValue("bookId")
		var selected *batchfactoryv11.Book
		for index := range batch.Books {
			if batch.Books[index].ID == bookID {
				selected = &batch.Books[index]
				break
			}
		}
		if selected == nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "小说不存在"})
			return
		}
		var input struct {
			AssetIDs    []string `json:"assetIds"`
			ModelID     string   `json:"modelId"`
			AspectRatio string   `json:"imageAspectRatio"`
		}
		if !decodeJSON(w, r, &input) {
			return
		}
		selectedBatch := batch
		selectedBatch.Books = []batchfactoryv11.Book{*selected}
		images, err := generateShuihuoAssetImages(r.Context(), store, files, owner, selectedBatch, input.AssetIDs, input.ModelID, input.AspectRatio, catalogStore)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		for index := range images {
			images[index] = bookAssetImageResponse(images[index], batch.ID, bookID, images[index].AssetID)
		}
		writeJSON(w, http.StatusCreated, map[string]any{"images": images})
	})
}
