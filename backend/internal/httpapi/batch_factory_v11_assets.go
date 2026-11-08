package httpapi

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"os"
	"strings"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/localartifact"
	"qiantie/backend/internal/localexecutor"
)

func registerBookAssetRoutes(mux *http.ServeMux, store batchfactoryv11.Store, files *localartifact.Store) {
	const collection = "/api/batch-factory/v11/batches/{batchId}/books/{bookId}/assets"
	const imageCollection = collection + "/{assetId}/images"
	const imageUpload = imageCollection + "/upload"
	const imageContent = imageCollection + "/{imageId}/content"
	const imageThumbnail = imageCollection + "/{imageId}/thumbnail"
	mux.HandleFunc("GET "+collection, func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		assets, err := store.ListBookAssets(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"assets": assets})
	})
	mux.HandleFunc("POST "+collection, func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input batchfactoryv11.CreateBookAssetInput
		if !decodeJSON(w, r, &input) {
			return
		}
		asset, err := store.CreateBookAsset(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"), input)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"asset": asset})
	})
	mux.HandleFunc("PATCH "+collection+"/{assetId}", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input batchfactoryv11.UpdateBookAssetInput
		if !decodeJSON(w, r, &input) {
			return
		}
		asset, err := store.UpdateBookAsset(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"), r.PathValue("assetId"), input)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"asset": asset})
	})
	mux.HandleFunc("GET "+imageCollection, func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		images, err := store.ListBookAssetImages(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"), r.PathValue("assetId"))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"images": bookAssetImageResponses(images, r.PathValue("batchId"), r.PathValue("bookId"), r.PathValue("assetId"))})
	})
	mux.HandleFunc("POST "+imageCollection, func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input batchfactoryv11.CreateBookAssetImageInput
		if !decodeJSON(w, r, &input) {
			return
		}
		if strings.EqualFold(strings.TrimSpace(input.Source), "upload") {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "uploaded image bytes must use the upload endpoint"})
			return
		}
		image, err := store.CreateBookAssetImage(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"), r.PathValue("assetId"), input)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"image": bookAssetImageResponse(image, r.PathValue("batchId"), r.PathValue("bookId"), r.PathValue("assetId"))})
	})
	mux.HandleFunc("POST "+imageUpload, func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		if files == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "asset image storage is unavailable"})
			return
		}
		var input struct {
			DataURL string `json:"dataUrl"`
		}
		if !decodeJSON(w, r, &input) {
			return
		}
		mediaType, payload, err := decodeImageDataURL(input.DataURL)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid image upload"})
			return
		}
		artifactID, err := localexecutor.NewArtifactID()
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "image id generation failed"})
			return
		}
		saved, err := files.SaveImage(artifactID, mediaType, bytes.NewReader(payload))
		if err != nil {
			writeBookAssetImageStorageError(w, err)
			return
		}
		image, err := store.CreateBookAssetImage(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"), r.PathValue("assetId"), batchfactoryv11.CreateBookAssetImageInput{StorageRef: saved.StorageRef, MediaType: saved.MediaType, Source: "upload"})
		if err != nil {
			_ = files.Remove(saved.StorageRef)
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"image": bookAssetImageResponse(image, r.PathValue("batchId"), r.PathValue("bookId"), r.PathValue("assetId"))})
	})
	mux.HandleFunc("PUT "+imageCollection+"/{imageId}/primary", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		image, err := store.SetPrimaryBookAssetImage(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"), r.PathValue("assetId"), r.PathValue("imageId"))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"image": bookAssetImageResponse(image, r.PathValue("batchId"), r.PathValue("bookId"), r.PathValue("assetId"))})
	})
	mux.HandleFunc("GET "+imageContent, func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		if files == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "asset image storage is unavailable"})
			return
		}
		images, err := store.ListBookAssetImages(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"), r.PathValue("assetId"))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		for _, image := range images {
			if image.ID != r.PathValue("imageId") {
				continue
			}
			if image.StorageRef == "" {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "image bytes are not stored locally"})
				return
			}
			file, err := files.Open(image.StorageRef)
			if err != nil {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "image file not found"})
				return
			}
			defer file.Close()
			w.Header().Set("Content-Type", image.MediaType)
			w.Header().Set("X-Content-Type-Options", "nosniff")
			// Each image ID is an immutable version. Allow the signed-in browser
			// to keep its own copy so revisiting a batch does not download the
			// same multi-megabyte asset again. `private` prevents shared caches
			// from mixing images between accounts.
			if info, statErr := file.Stat(); statErr == nil {
				w.Header().Set("ETag", fmt.Sprintf("\"%x-%x\"", info.Size(), info.ModTime().UTC().UnixNano()))
			}
			w.Header().Set("Cache-Control", "private, max-age=604800, immutable")
			http.ServeContent(w, r, image.ID, image.CreatedAt, file)
			return
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "image not found"})
	})
	mux.HandleFunc("GET "+imageThumbnail, handleBookAssetThumbnail(store, files))
}

func handleBookAssetThumbnail(store batchfactoryv11.Store, files *localartifact.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		if files == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "asset image storage is unavailable"})
			return
		}
		images, err := store.ListBookAssetImages(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"), r.PathValue("assetId"))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		for _, image := range images {
			if image.ID != r.PathValue("imageId") {
				continue
			}
			if image.StorageRef == "" {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "image bytes are not stored locally"})
				return
			}
			original, err := files.Open(image.StorageRef)
			if err != nil {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "image file not found"})
				return
			}
			defer original.Close()
			info, err := original.Stat()
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "image file is not readable"})
				return
			}
			key := imageThumbnailCacheKey(image.StorageRef, image.MediaType, info)
			path, err := ensureImageThumbnail(files, original, key)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "thumbnail generation failed"})
				return
			}
			thumbnail, err := os.Open(path)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "thumbnail is not readable"})
				return
			}
			defer thumbnail.Close()
			w.Header().Set("Content-Type", "image/jpeg")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("Cache-Control", "private, max-age=604800, immutable")
			w.Header().Set("ETag", `"`+key[:32]+`"`)
			http.ServeContent(w, r, key[:32]+".jpg", time.Time{}, thumbnail)
			return
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "image not found"})
	}
}

func decodeImageDataURL(value string) (string, []byte, error) {
	parts := strings.SplitN(strings.TrimSpace(value), ",", 2)
	if len(parts) != 2 || !strings.HasPrefix(parts[0], "data:") || !strings.HasSuffix(parts[0], ";base64") {
		return "", nil, errors.New("invalid data url")
	}
	mediaType, _, err := mime.ParseMediaType(strings.TrimSuffix(strings.TrimPrefix(parts[0], "data:"), ";base64"))
	if err != nil {
		return "", nil, err
	}
	bytes, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil || len(bytes) == 0 {
		return "", nil, errors.New("invalid base64")
	}
	return strings.ToLower(mediaType), bytes, nil
}

func bookAssetImageResponse(image batchfactoryv11.BookAssetImage, batchID, bookID, assetID string) batchfactoryv11.BookAssetImage {
	if image.StorageRef != "" {
		base := "/api/batch-factory/v12/batches/" + batchID + "/books/" + bookID + "/assets/" + assetID + "/images/" + image.ID
		image.URL = base + "/content"
		image.ThumbnailURL = base + "/thumbnail"
	}
	return image
}

func bookAssetImageResponses(images []batchfactoryv11.BookAssetImage, batchID, bookID, assetID string) []batchfactoryv11.BookAssetImage {
	for index := range images {
		images[index] = bookAssetImageResponse(images[index], batchID, bookID, assetID)
	}
	return images
}

func writeBookAssetImageStorageError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, localartifact.ErrInvalidImage), errors.Is(err, localartifact.ErrInvalidID):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid image upload"})
	case errors.Is(err, localartifact.ErrTooLarge):
		writeJSON(w, http.StatusRequestEntityTooLarge, map[string]string{"error": "image exceeds size limit"})
	case errors.Is(err, localartifact.ErrAlreadyExists):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "image already exists"})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "image storage failed"})
	}
}
