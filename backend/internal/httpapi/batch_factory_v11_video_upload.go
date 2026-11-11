package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/localartifact"
)

type videoUploadOutput interface {
	PutUploadedMP4(context.Context, string, string) (string, error)
}

func registerVideoUploadRoutes(mux *http.ServeMux, service *batchfactoryv11.ProductionService, mergeService *batchfactoryv11.MergeService, files *localartifact.Store, output videoUploadOutput) {
	mux.HandleFunc("POST /api/batch-factory/v11/batches/{batchId}/books/{bookId}/videos/{videoId}/upload", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		if service == nil || files == nil || output == nil {
			writeStoreError(w, batchfactoryv11.ErrUnavailable)
			return
		}
		if !limitRequestBody(w, r, maxVideoUploadBodyBytes+maxMultipartOverheadBytes) {
			return
		}
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			if isBodyTooLarge(err) {
				writePayloadTooLarge(w, maxVideoUploadBodyBytes+maxMultipartOverheadBytes)
				return
			}
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid video upload"})
			return
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "MP4 file is required"})
			return
		}
		defer file.Close()
		if !strings.EqualFold(strings.TrimSpace(header.Header.Get("Content-Type")), "video/mp4") && !strings.HasSuffix(strings.ToLower(header.Filename), ".mp4") {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "only MP4 files are accepted"})
			return
		}
		uploadID, err := newVideoUploadID()
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "create upload id failed"})
			return
		}
		artifact, err := files.SaveMP4(uploadID, file)
		if err != nil {
			writeVideoUploadStorageError(w, err)
			return
		}
		defer files.Remove(artifact.StorageRef)
		localPath, err := files.Path(artifact.StorageRef)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "video storage failed"})
			return
		}
		mediaURL, err := output.PutUploadedMP4(r.Context(), uploadID, localPath)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "upload video to TOS failed"})
			return
		}
		task, err := service.RegisterUploadedVideo(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"), r.PathValue("videoId"), uploadID, mediaURL)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"task": task})
	})
	if mergeService != nil && files != nil && output != nil {
		mux.HandleFunc("POST /api/batch-factory/v11/batches/{batchId}/books/{bookId}/merge/upload", func(w http.ResponseWriter, r *http.Request) {
			owner, ok := bridgeOwner(r)
			if !ok {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}
			if !limitRequestBody(w, r, maxVideoUploadBodyBytes+maxMultipartOverheadBytes) {
				return
			}
			if err := r.ParseMultipartForm(32 << 20); err != nil {
				if isBodyTooLarge(err) {
					writePayloadTooLarge(w, maxVideoUploadBodyBytes+maxMultipartOverheadBytes)
					return
				}
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid video upload"})
				return
			}
			file, header, err := r.FormFile("file")
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "MP4 file is required"})
				return
			}
			defer file.Close()
			if !strings.EqualFold(strings.TrimSpace(header.Header.Get("Content-Type")), "video/mp4") && !strings.HasSuffix(strings.ToLower(header.Filename), ".mp4") {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "only MP4 files are accepted"})
				return
			}
			uploadID, err := newVideoUploadID()
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "create upload id failed"})
				return
			}
			artifact, err := files.SaveMP4(uploadID, file)
			if err != nil {
				writeVideoUploadStorageError(w, err)
				return
			}
			defer files.Remove(artifact.StorageRef)
			localPath, err := files.Path(artifact.StorageRef)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "video storage failed"})
				return
			}
			mediaURL, err := output.PutUploadedMP4(r.Context(), uploadID, localPath)
			if err != nil {
				writeJSON(w, http.StatusBadGateway, map[string]string{"error": "upload video to TOS failed"})
				return
			}
			job, err := mergeService.RegisterUploadedMerge(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"), "manual-upload:"+uploadID, mediaURL)
			if err != nil {
				writeStoreError(w, err)
				return
			}
			writeJSON(w, http.StatusCreated, map[string]any{"job": job})
		})
	}
}

func newVideoUploadID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "upload_" + hex.EncodeToString(bytes), nil
}

func writeVideoUploadStorageError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, localartifact.ErrInvalidMP4), errors.Is(err, localartifact.ErrInvalidID):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid MP4 upload"})
	case errors.Is(err, localartifact.ErrTooLarge):
		writeJSON(w, http.StatusRequestEntityTooLarge, map[string]string{"error": "video exceeds size limit"})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "video storage failed"})
	}
}
