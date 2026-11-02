package httpapi

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"qiantie/backend/internal/batchfactoryv11"
)

func comicSourceDigest(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func registerComicSourceSegmentRoutes(mux *http.ServeMux, store batchfactoryv11.Store, director *batchfactoryv11.DirectorService) {
	segmentStore, ok := store.(batchfactoryv11.ComicSourceSegmentStore)
	if mux == nil || !ok {
		return
	}
	const base = "/api/batch-factory/v12/batches/{batchId}/books/{bookId}/comic-source-segments"
	mux.HandleFunc("GET "+base, func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		result, err := segmentStore.GetComicSourceSegmentation(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"segmentation": result})
	})
	mux.HandleFunc("POST "+base+"/segment", func(w http.ResponseWriter, r *http.Request) {
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
			writeStoreError(w, batchfactoryv11.ErrNotFound)
			return
		}
		if director != nil && director.Provider != nil {
			result, err := director.RunComicSourceSegmentation(r.Context(), owner, batch.ID, book.ID)
			if err != nil {
				writeStoreError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"segmentation": result})
			return
		}
		source := book.SourceText
		if book.WorkingFrontContent != "" {
			source = book.WorkingFrontContent
		}
		boundaries := batchfactoryv11.EnumerateComicSourceBoundaries(source)
		drafts, err := batchfactoryv11.FallbackComicSourceSegments(source, boundaries)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		result, err := segmentStore.ReplaceComicSourceSegments(r.Context(), owner, batch.ID, book.ID, comicSourceDigest(source), drafts)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"segmentation": result})
	})
	mux.HandleFunc("POST "+base+"/{videoId}/split", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input batchfactoryv11.ComicSourceSplitInput
		if !decodeJSON(w, r, &input) {
			return
		}
		result, err := segmentStore.SplitComicSourceSegment(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"), r.PathValue("videoId"), input)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"segmentation": result})
	})
}
