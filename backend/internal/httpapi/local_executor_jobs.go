package httpapi

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"qiantie/backend/internal/localexecutor"
)

func RegisterLocalExecutorJobRoutes(root *http.ServeMux, auth BridgeAuth, service *localexecutor.Service) {
	if root == nil || service == nil {
		return
	}

	root.Handle("POST /api/shuihuo-production/local-executor-jobs", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeExecutorError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		var input localexecutor.CreateJobInput
		if err := decodeExecutorJSON(w, req, &input); err != nil {
			writeExecutorError(w, http.StatusBadRequest, "invalid request")
			return
		}
		job, err := service.CreateJob(req.Context(), identity.Username, input)
		if err != nil {
			writeLocalJobError(w, err)
			return
		}
		writeExecutorJSON(w, http.StatusCreated, job)
	})))

	root.Handle("GET /api/shuihuo-production/local-executor-jobs/{id}", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeExecutorError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		job, err := service.GetJob(req.Context(), identity.Username, req.PathValue("id"))
		if err != nil {
			writeLocalJobError(w, err)
			return
		}
		writeExecutorJSON(w, http.StatusOK, job)
	})))

	root.Handle("PUT /api/shuihuo-production/local-executor-jobs/{id}/cancel", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeExecutorError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		job, err := service.CancelJob(req.Context(), identity.Username, req.PathValue("id"))
		if err != nil {
			writeLocalJobError(w, err)
			return
		}
		writeExecutorJSON(w, http.StatusOK, job)
	})))

	root.Handle("POST /api/shuihuo-production/local-executor-tests", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeExecutorError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		var input struct {
			Type     string `json:"type"`
			Prompt   string `json:"prompt"`
			Ratio    string `json:"ratio"`
			Duration int    `json:"duration"`
		}
		if err := decodeExecutorJSON(w, req, &input); err != nil {
			writeExecutorError(w, http.StatusBadRequest, "invalid request")
			return
		}
		input.Type = strings.ToLower(strings.TrimSpace(input.Type))
		input.Prompt = strings.TrimSpace(input.Prompt)
		if (input.Type != "image" && input.Type != "video") || input.Prompt == "" {
			writeExecutorError(w, http.StatusBadRequest, "type must be image or video and prompt is required")
			return
		}
		if strings.TrimSpace(input.Ratio) == "" {
			if input.Type == "image" {
				input.Ratio = "1:1"
			} else {
				input.Ratio = "16:9"
			}
		}
		payload := map[string]any{
			"type": input.Type, "prompt": input.Prompt, "ratio": input.Ratio,
		}
		if input.Type == "image" {
			payload["model"] = "doubao-image-4.5"
		} else {
			if input.Duration <= 0 {
				input.Duration = 5
			}
			payload["model"] = "doubao-video-fast"
			payload["duration"] = input.Duration
		}
		job, err := service.CreateJob(req.Context(), identity.Username, localexecutor.CreateJobInput{
			SourceTaskID: "settings-test:" + input.Type + ":" + time.Now().UTC().Format("20060102150405.000000000"),
			Platform:     localexecutor.PlatformDoubao,
			Payload:      payload,
		})
		if err != nil {
			writeLocalJobError(w, err)
			return
		}
		writeExecutorJSON(w, http.StatusAccepted, map[string]any{"jobId": job.ID, "type": input.Type, "status": job.State})
	})))

	root.Handle("GET /api/shuihuo-production/local-executor-tests/{id}", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeExecutorError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		job, err := service.GetJob(req.Context(), identity.Username, req.PathValue("id"))
		if err != nil {
			writeLocalJobError(w, err)
			return
		}
		result := map[string]any{"jobId": job.ID, "status": job.State}
		if job.State == localexecutor.JobFailed || job.State == localexecutor.JobCancelled {
			result["error"] = job.ErrorMessage
		}
		if job.State == localexecutor.JobSucceeded && job.ArtifactID != "" {
			artifact, artifactErr := service.GetArtifact(req.Context(), identity.Username, job.ArtifactID)
			if artifactErr != nil {
				writeArtifactDomainError(w, artifactErr)
				return
			}
			result["mediaType"] = artifact.MediaType
			result["resultUrl"] = "/api/shuihuo-production/local-executor-artifacts/" + artifact.ID
		}
		writeExecutorJSON(w, http.StatusOK, result)
	})))

	root.HandleFunc("POST /api/local-executor/v1/jobs/claim", func(w http.ResponseWriter, req *http.Request) {
		token, ok := executorBearerToken(req)
		if !ok {
			writeExecutorError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		wait := time.Duration(0)
		if raw := req.URL.Query().Get("waitSeconds"); raw != "" {
			seconds, parseErr := strconv.Atoi(raw)
			if parseErr != nil || seconds < 1 || seconds > 30 {
				writeExecutorError(w, http.StatusBadRequest, "waitSeconds must be between 1 and 30")
				return
			}
			wait = time.Duration(seconds) * time.Second
		}
		claim, err := service.ClaimJobWait(req.Context(), token, wait)
		if errors.Is(err, localexecutor.ErrNoClaimableJob) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if err != nil {
			writeLocalJobError(w, err)
			return
		}
		writeExecutorJSON(w, http.StatusOK, claim)
	})

	root.HandleFunc("POST /api/local-executor/v1/jobs/{id}/renew", func(w http.ResponseWriter, req *http.Request) {
		token, ok := executorBearerToken(req)
		if !ok {
			writeExecutorError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		var input localexecutor.LeaseCredential
		if err := decodeExecutorJSON(w, req, &input); err != nil {
			writeExecutorError(w, http.StatusBadRequest, "invalid request")
			return
		}
		lease, err := service.RenewJob(req.Context(), token, req.PathValue("id"), input)
		if err != nil {
			writeLocalJobError(w, err)
			return
		}
		writeExecutorJSON(w, http.StatusOK, lease)
	})

	root.HandleFunc("POST /api/local-executor/v1/jobs/{id}/progress", func(w http.ResponseWriter, req *http.Request) {
		token, ok := executorBearerToken(req)
		if !ok {
			writeExecutorError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		var input struct {
			localexecutor.LeaseCredential
			State localexecutor.JobState `json:"state"`
		}
		if err := decodeExecutorJSON(w, req, &input); err != nil {
			writeExecutorError(w, http.StatusBadRequest, "invalid request")
			return
		}
		if err := service.RecordProgress(req.Context(), token, req.PathValue("id"), input.LeaseCredential, input.State); err != nil {
			writeLocalJobError(w, err)
			return
		}
		writeExecutorJSON(w, http.StatusOK, map[string]any{"ok": true})
	})

	root.HandleFunc("POST /api/local-executor/v1/jobs/{id}/acceptance", func(w http.ResponseWriter, req *http.Request) {
		token, ok := executorBearerToken(req)
		if !ok {
			writeExecutorError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		var input struct {
			localexecutor.LeaseCredential
			AccountID    string `json:"accountId"`
			SubmissionID string `json:"submissionId"`
		}
		if err := decodeExecutorJSON(w, req, &input); err != nil {
			writeExecutorError(w, http.StatusBadRequest, "invalid request")
			return
		}
		err := service.RecordAcceptance(req.Context(), token, req.PathValue("id"), input.LeaseCredential, localexecutor.AcceptanceInput{AccountID: input.AccountID, SubmissionID: input.SubmissionID})
		if err != nil {
			writeLocalJobError(w, err)
			return
		}
		writeExecutorJSON(w, http.StatusOK, map[string]any{"ok": true})
	})

	root.HandleFunc("POST /api/local-executor/v1/jobs/{id}/release", func(w http.ResponseWriter, req *http.Request) {
		token, ok := executorBearerToken(req)
		if !ok {
			writeExecutorError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		var input struct {
			localexecutor.LeaseCredential
			Reason string `json:"reason"`
		}
		if err := decodeExecutorJSON(w, req, &input); err != nil {
			writeExecutorError(w, http.StatusBadRequest, "invalid request")
			return
		}
		if err := service.ReleaseJob(req.Context(), token, req.PathValue("id"), input.LeaseCredential, input.Reason); err != nil {
			writeLocalJobError(w, err)
			return
		}
		writeExecutorJSON(w, http.StatusOK, map[string]any{"ok": true})
	})

	root.HandleFunc("POST /api/local-executor/v1/jobs/{id}/fail", func(w http.ResponseWriter, req *http.Request) {
		token, ok := executorBearerToken(req)
		if !ok {
			writeExecutorError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		var input struct {
			localexecutor.LeaseCredential
			Code    string `json:"code"`
			Message string `json:"message"`
		}
		if err := decodeExecutorJSON(w, req, &input); err != nil {
			writeExecutorError(w, http.StatusBadRequest, "invalid request")
			return
		}
		err := service.FailJob(req.Context(), token, req.PathValue("id"), input.LeaseCredential, localexecutor.FailureInput{Code: input.Code, Message: input.Message})
		if err != nil {
			writeLocalJobError(w, err)
			return
		}
		writeExecutorJSON(w, http.StatusOK, map[string]any{"ok": true})
	})

	root.HandleFunc("POST /api/local-executor/v1/jobs/{id}/result", func(w http.ResponseWriter, req *http.Request) {
		token, ok := executorBearerToken(req)
		if !ok {
			writeExecutorError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		var input struct {
			localexecutor.LeaseCredential
			ArtifactID string `json:"artifactId"`
		}
		if err := decodeExecutorJSON(w, req, &input); err != nil {
			writeExecutorError(w, http.StatusBadRequest, "invalid request")
			return
		}
		if err := service.CompleteJob(req.Context(), token, req.PathValue("id"), input.LeaseCredential, localexecutor.ResultInput{ArtifactID: input.ArtifactID}); err != nil {
			writeLocalJobError(w, err)
			return
		}
		writeExecutorJSON(w, http.StatusOK, map[string]any{"ok": true})
	})
}

func writeLocalJobError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, localexecutor.ErrInvalidPlatform), errors.Is(err, localexecutor.ErrInvalidInput):
		writeExecutorError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, localexecutor.ErrExecutorUnauthorized):
		writeExecutorError(w, http.StatusUnauthorized, "unauthorized")
	case errors.Is(err, localexecutor.ErrJobNotFound):
		writeExecutorError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, localexecutor.ErrStaleLease), errors.Is(err, localexecutor.ErrJobCancelled), errors.Is(err, localexecutor.ErrAcceptedJobPinned), errors.Is(err, localexecutor.ErrInvalidJobState), errors.Is(err, localexecutor.ErrJobConflict):
		writeExecutorError(w, http.StatusConflict, err.Error())
	default:
		writeExecutorError(w, http.StatusInternalServerError, "local executor job service error")
	}
}
