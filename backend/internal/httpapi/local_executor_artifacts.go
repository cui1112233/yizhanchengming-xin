package httpapi

import (
	"errors"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"time"

	"qiantie/backend/internal/localartifact"
	"qiantie/backend/internal/localexecutor"
)

func RegisterLocalExecutorArtifactRoutes(root *http.ServeMux, auth BridgeAuth, service *localexecutor.Service, files *localartifact.Store) {
	if root == nil || service == nil || files == nil {
		return
	}

	root.HandleFunc("POST /api/local-executor/v1/jobs/{id}/artifact", func(w http.ResponseWriter, req *http.Request) {
		token, ok := executorBearerToken(req)
		if !ok {
			writeExecutorError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		mediaType, _, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
		mediaType = strings.ToLower(mediaType)
		if err != nil || !supportedExecutorArtifactMediaType(mediaType) {
			writeExecutorError(w, http.StatusBadRequest, "artifact must be video/mp4, image/png, image/jpeg, or image/webp")
			return
		}
		generation, err := strconv.ParseInt(strings.TrimSpace(req.Header.Get("X-Lease-Generation")), 10, 64)
		if err != nil || generation < 1 {
			writeExecutorError(w, http.StatusBadRequest, "invalid lease generation")
			return
		}
		lease := localexecutor.LeaseCredential{
			Token:      strings.TrimSpace(req.Header.Get("X-Lease-Token")),
			Generation: generation,
		}
		if lease.Token == "" {
			writeExecutorError(w, http.StatusBadRequest, "missing lease token")
			return
		}
		jobID := req.PathValue("id")
		if err := service.RecordProgress(req.Context(), token, jobID, lease, localexecutor.JobUploading); err != nil {
			writeLocalJobError(w, err)
			return
		}

		artifactID, err := localexecutor.NewArtifactID()
		if err != nil {
			writeExecutorError(w, http.StatusInternalServerError, "artifact id generation failed")
			return
		}
		var saved localartifact.Artifact
		if mediaType == "video/mp4" {
			saved, err = files.SaveMP4(artifactID, req.Body)
		} else {
			saved, err = files.SaveImage(artifactID, mediaType, req.Body)
		}
		if err != nil {
			writeArtifactStoreError(w, err)
			return
		}

		bound, err := service.RecordArtifact(req.Context(), token, jobID, lease, localexecutor.ArtifactInput{
			ID: saved.ID, MediaType: saved.MediaType, ByteSize: saved.ByteSize,
			SHA256: saved.SHA256, StorageRef: saved.StorageRef,
		})
		if err != nil {
			_ = files.Remove(saved.StorageRef)
			writeArtifactDomainError(w, err)
			return
		}
		status := http.StatusCreated
		if bound.ID != saved.ID {
			_ = files.Remove(saved.StorageRef)
			status = http.StatusOK
		}
		writeExecutorJSON(w, status, bound)
	})

	// Remote merge/publish providers cannot send a user cookie. Accept only a
	// short-lived HMAC token bound to the owner and artifact id for this media
	// path; the normal authenticated player route remains unchanged.
	root.HandleFunc("GET /api/local-executor/v1/artifacts/{id}", func(w http.ResponseWriter, req *http.Request) {
		owner, err := localexecutor.VerifyArtifactPublicToken(auth.Secret, req.PathValue("id"), req.URL.Query().Get("token"), time.Now().UTC())
		if err != nil {
			writeExecutorError(w, http.StatusForbidden, "invalid or expired artifact token")
			return
		}
		artifact, err := service.GetArtifact(req.Context(), owner, req.PathValue("id"))
		if err != nil {
			writeArtifactDomainError(w, err)
			return
		}
		file, err := files.Open(artifact.StorageRef)
		if err != nil {
			if errors.Is(err, localartifact.ErrInvalidID) {
				writeExecutorError(w, http.StatusInternalServerError, "invalid artifact storage reference")
				return
			}
			writeExecutorError(w, http.StatusNotFound, "artifact file not found")
			return
		}
		defer file.Close()
		contentType, filename, err := artifactResponseMetadata(artifact.ID, artifact.MediaType)
		if err != nil {
			writeExecutorError(w, http.StatusInternalServerError, "invalid artifact media type")
			return
		}
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		http.ServeContent(w, req, filename, artifact.CreatedAt, file)
	})

	root.Handle("GET /api/shuihuo-production/local-executor-artifacts/{id}", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeExecutorError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		artifact, err := service.GetArtifact(req.Context(), identity.Username, req.PathValue("id"))
		if err != nil {
			writeArtifactDomainError(w, err)
			return
		}
		file, err := files.Open(artifact.StorageRef)
		if err != nil {
			if errors.Is(err, localartifact.ErrInvalidID) {
				writeExecutorError(w, http.StatusInternalServerError, "invalid artifact storage reference")
				return
			}
			writeExecutorError(w, http.StatusNotFound, "artifact file not found")
			return
		}
		defer file.Close()
		contentType, filename, err := artifactResponseMetadata(artifact.ID, artifact.MediaType)
		if err != nil {
			writeExecutorError(w, http.StatusInternalServerError, "invalid artifact media type")
			return
		}
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		http.ServeContent(w, req, filename, artifact.CreatedAt, file)
	})))
}

func supportedExecutorArtifactMediaType(mediaType string) bool {
	switch strings.ToLower(strings.TrimSpace(mediaType)) {
	case "video/mp4", "image/png", "image/jpeg", "image/webp":
		return true
	default:
		return false
	}
}

func artifactResponseMetadata(id, mediaType string) (string, string, error) {
	mediaType = strings.ToLower(strings.TrimSpace(mediaType))
	extension := ""
	switch mediaType {
	case "video/mp4":
		extension = ".mp4"
	case "image/png":
		extension = ".png"
	case "image/jpeg":
		extension = ".jpg"
	case "image/webp":
		extension = ".webp"
	default:
		return "", "", errors.New("unsupported artifact media type")
	}
	return mediaType, id + extension, nil
}

func writeArtifactDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, localexecutor.ErrArtifactNotFound):
		writeExecutorError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, localexecutor.ErrArtifactConflict):
		writeExecutorError(w, http.StatusConflict, err.Error())
	default:
		writeLocalJobError(w, err)
	}
}

func writeArtifactStoreError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, localartifact.ErrInvalidMP4), errors.Is(err, localartifact.ErrInvalidImage), errors.Is(err, localartifact.ErrInvalidID):
		writeExecutorError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, localartifact.ErrTooLarge):
		writeExecutorError(w, http.StatusRequestEntityTooLarge, err.Error())
	case errors.Is(err, localartifact.ErrAlreadyExists):
		writeExecutorError(w, http.StatusConflict, err.Error())
	default:
		writeExecutorError(w, http.StatusInternalServerError, "artifact storage error")
	}
}
