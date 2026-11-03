package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"qiantie/backend/internal/giantmaterialexecutor"
)

const giantMaterialExecutorBodyLimit = giantmaterialexecutor.MaxResultBytes + (64 << 10)

// giantClaimLongPollWait 是 claim 接口的长轮询等待时长；用变量便于接口测试
// 即时返回。
var giantClaimLongPollWait = giantmaterialexecutor.ClaimLongPollWait

func RegisterGiantMaterialExecutorRoutes(root *http.ServeMux, auth BridgeAuth, service *giantmaterialexecutor.Service) {
	if root == nil || service == nil {
		return
	}

	root.Handle("GET /api/shuihuo-production/giant-material-executors", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeGiantExecutorError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		executors, err := service.ListExecutors(req.Context(), identity.Username)
		if err != nil {
			writeGiantExecutorServiceError(w, err)
			return
		}
		writeGiantExecutorJSON(w, http.StatusOK, map[string]any{"executors": executors})
	})))

	root.Handle("DELETE /api/shuihuo-production/giant-material-executors/{id}", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeGiantExecutorError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		if err := service.DeleteExecutor(req.Context(), identity.Username, req.PathValue("id")); err != nil {
			writeGiantExecutorServiceError(w, err)
			return
		}
		writeGiantExecutorJSON(w, http.StatusOK, map[string]any{"ok": true})
	})))

	root.Handle("GET /api/shuihuo-production/giant-material-executor/preference", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeGiantExecutorError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		preference, err := service.GetPreference(req.Context(), identity.Username)
		if err != nil {
			writeGiantExecutorServiceError(w, err)
			return
		}
		writeGiantExecutorJSON(w, http.StatusOK, preference)
	})))

	root.Handle("PUT /api/shuihuo-production/giant-material-executor/preference", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeGiantExecutorError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		var input struct {
			PreferredOS string `json:"preferredOs"`
		}
		if err := decodeGiantExecutorJSON(w, req, &input); err != nil {
			writeGiantExecutorError(w, http.StatusBadRequest, "invalid request")
			return
		}
		preference, err := service.SavePreference(req.Context(), identity.Username, input.PreferredOS)
		if err != nil {
			writeGiantExecutorServiceError(w, err)
			return
		}
		writeGiantExecutorJSON(w, http.StatusOK, preference)
	})))

	root.Handle("POST /api/shuihuo-production/giant-material-executor/pairings", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeGiantExecutorError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		var input struct {
			Platform string `json:"platform"`
		}
		if err := decodeGiantExecutorJSON(w, req, &input); err != nil {
			writeGiantExecutorError(w, http.StatusBadRequest, "invalid request")
			return
		}
		pairing, err := service.CreatePairing(req.Context(), identity.Username, input.Platform)
		if err != nil {
			writeGiantExecutorServiceError(w, err)
			return
		}
		writeGiantExecutorJSON(w, http.StatusCreated, pairing)
	})))

	root.Handle("POST /api/shuihuo-production/giant-material-jobs", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeGiantExecutorError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		var input giantmaterialexecutor.CreateJobInput
		if err := decodeGiantExecutorJSON(w, req, &input); err != nil {
			writeGiantExecutorError(w, http.StatusBadRequest, "invalid request")
			return
		}
		job, err := service.CreateJob(req.Context(), identity.Username, input)
		if err != nil {
			writeGiantExecutorServiceError(w, err)
			return
		}
		writeGiantExecutorJSON(w, http.StatusCreated, map[string]any{"job": job})
	})))

	root.Handle("GET /api/shuihuo-production/giant-material-jobs/{id}", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeGiantExecutorError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		job, err := service.GetJob(req.Context(), identity.Username, req.PathValue("id"))
		if err != nil {
			writeGiantExecutorServiceError(w, err)
			return
		}
		writeGiantExecutorJSON(w, http.StatusOK, map[string]any{"job": job})
	})))

	root.Handle("PUT /api/shuihuo-production/giant-material-jobs/{id}/cancel", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeGiantExecutorError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		job, err := service.CancelJob(req.Context(), identity.Username, req.PathValue("id"))
		if err != nil {
			writeGiantExecutorServiceError(w, err)
			return
		}
		writeGiantExecutorJSON(w, http.StatusOK, map[string]any{"job": job})
	})))

	root.Handle("POST /api/shuihuo-production/giant-material-jobs/{id}/retry", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeGiantExecutorError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		job, err := service.RetryJob(req.Context(), identity.Username, req.PathValue("id"))
		if err != nil {
			writeGiantExecutorServiceError(w, err)
			return
		}
		writeGiantExecutorJSON(w, http.StatusOK, map[string]any{"job": job})
	})))

	root.HandleFunc("POST /api/giant-material-executor/v1/pair", func(w http.ResponseWriter, req *http.Request) {
		var input giantmaterialexecutor.PairInput
		if err := decodeGiantExecutorJSON(w, req, &input); err != nil {
			writeGiantExecutorError(w, http.StatusBadRequest, "invalid request")
			return
		}
		result, err := service.Pair(req.Context(), input)
		if err != nil {
			writeGiantExecutorServiceError(w, err)
			return
		}
		writeGiantExecutorJSON(w, http.StatusOK, result)
	})

	root.HandleFunc("POST /api/giant-material-executor/v1/heartbeat", func(w http.ResponseWriter, req *http.Request) {
		token, ok := executorBearerToken(req)
		if !ok {
			writeGiantExecutorError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		var input giantmaterialexecutor.HeartbeatInput
		if err := decodeGiantExecutorJSON(w, req, &input); err != nil {
			writeGiantExecutorError(w, http.StatusBadRequest, "invalid request")
			return
		}
		if err := service.Heartbeat(req.Context(), token, input); err != nil {
			writeGiantExecutorServiceError(w, err)
			return
		}
		writeGiantExecutorJSON(w, http.StatusOK, map[string]any{"ok": true, "heartbeatIntervalSeconds": giantmaterialexecutor.HeartbeatIntervalSeconds})
	})

	root.HandleFunc("POST /api/giant-material-executor/v1/jobs/claim", func(w http.ResponseWriter, req *http.Request) {
		token, ok := executorBearerToken(req)
		if !ok {
			writeGiantExecutorError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		// 长轮询：没活时在服务器侧等待，老版本执行器无需升级即可自动降噪。
		claim, err := service.ClaimWhenAvailable(req.Context(), token, giantClaimLongPollWait)
		if errors.Is(err, context.Canceled) {
			return
		}
		if errors.Is(err, giantmaterialexecutor.ErrNoClaimableJob) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if err != nil {
			writeGiantExecutorServiceError(w, err)
			return
		}
		writeGiantExecutorJSON(w, http.StatusOK, claim)
	})

	root.HandleFunc("POST /api/giant-material-executor/v1/jobs/{id}/renew", func(w http.ResponseWriter, req *http.Request) {
		token, ok := executorBearerToken(req)
		if !ok {
			writeGiantExecutorError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		var input giantmaterialexecutor.LeaseCredential
		if err := decodeGiantExecutorJSON(w, req, &input); err != nil {
			writeGiantExecutorError(w, http.StatusBadRequest, "invalid request")
			return
		}
		lease, err := service.Renew(req.Context(), token, req.PathValue("id"), input)
		if err != nil {
			writeGiantExecutorServiceError(w, err)
			return
		}
		writeGiantExecutorJSON(w, http.StatusOK, lease)
	})

	root.HandleFunc("POST /api/giant-material-executor/v1/jobs/{id}/progress", func(w http.ResponseWriter, req *http.Request) {
		token, ok := executorBearerToken(req)
		if !ok {
			writeGiantExecutorError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		var input struct {
			giantmaterialexecutor.LeaseCredential
			State     giantmaterialexecutor.JobState `json:"state"`
			Completed int                            `json:"completed"`
			Total     int                            `json:"total"`
			Percent   int                            `json:"percent"`
		}
		if err := decodeGiantExecutorJSON(w, req, &input); err != nil {
			writeGiantExecutorError(w, http.StatusBadRequest, "invalid request")
			return
		}
		if err := service.Progress(req.Context(), token, req.PathValue("id"), input.LeaseCredential, input.State, giantmaterialexecutor.ProgressInput{Completed: input.Completed, Total: input.Total, Percent: input.Percent}); err != nil {
			writeGiantExecutorServiceError(w, err)
			return
		}
		writeGiantExecutorJSON(w, http.StatusOK, map[string]any{"ok": true})
	})

	root.HandleFunc("POST /api/giant-material-executor/v1/jobs/{id}/result", func(w http.ResponseWriter, req *http.Request) {
		token, ok := executorBearerToken(req)
		if !ok {
			writeGiantExecutorError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		var input struct {
			giantmaterialexecutor.LeaseCredential
			Text      string `json:"text"`
			WordCount int    `json:"wordCount"`
		}
		if err := decodeGiantExecutorJSON(w, req, &input); err != nil {
			writeGiantExecutorError(w, http.StatusBadRequest, "invalid request")
			return
		}
		job, err := service.Complete(req.Context(), token, req.PathValue("id"), input.LeaseCredential, giantmaterialexecutor.ResultInput{Text: input.Text, WordCount: input.WordCount})
		if err != nil {
			writeGiantExecutorServiceError(w, err)
			return
		}
		writeGiantExecutorJSON(w, http.StatusOK, map[string]any{"job": job})
	})

	root.HandleFunc("POST /api/giant-material-executor/v1/jobs/{id}/fail", func(w http.ResponseWriter, req *http.Request) {
		token, ok := executorBearerToken(req)
		if !ok {
			writeGiantExecutorError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		var input struct {
			giantmaterialexecutor.LeaseCredential
			Code    string `json:"code"`
			Message string `json:"message"`
		}
		if err := decodeGiantExecutorJSON(w, req, &input); err != nil {
			writeGiantExecutorError(w, http.StatusBadRequest, "invalid request")
			return
		}
		if err := service.Fail(req.Context(), token, req.PathValue("id"), input.LeaseCredential, giantmaterialexecutor.FailureInput{Code: input.Code, Message: input.Message}); err != nil {
			writeGiantExecutorServiceError(w, err)
			return
		}
		writeGiantExecutorJSON(w, http.StatusOK, map[string]any{"ok": true})
	})
}

func decodeGiantExecutorJSON(w http.ResponseWriter, req *http.Request, target any) error {
	req.Body = http.MaxBytesReader(w, req.Body, giantMaterialExecutorBodyLimit)
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("multiple json values")
	}
	return nil
}

func writeGiantExecutorServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, giantmaterialexecutor.ErrInvalidPlatform), errors.Is(err, giantmaterialexecutor.ErrInvalidInput), errors.Is(err, giantmaterialexecutor.ErrResultTooLarge):
		writeGiantExecutorError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, giantmaterialexecutor.ErrPairingInvalid):
		writeGiantExecutorError(w, http.StatusConflict, err.Error())
	case errors.Is(err, giantmaterialexecutor.ErrExecutorUnauthorized):
		writeGiantExecutorError(w, http.StatusUnauthorized, "unauthorized")
	case errors.Is(err, giantmaterialexecutor.ErrJobNotFound):
		writeGiantExecutorError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, giantmaterialexecutor.ErrExecutorNotFound):
		writeGiantExecutorError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, giantmaterialexecutor.ErrExecutorOffline):
		writeGiantExecutorJSON(w, http.StatusServiceUnavailable, map[string]any{"error": err.Error(), "code": ErrorCodeExecutorNotReady})
	case errors.Is(err, giantmaterialexecutor.ErrNoClaimableJob), errors.Is(err, giantmaterialexecutor.ErrStaleLease), errors.Is(err, giantmaterialexecutor.ErrJobCancelled), errors.Is(err, giantmaterialexecutor.ErrInvalidJobState), errors.Is(err, giantmaterialexecutor.ErrJobConflict):
		writeGiantExecutorError(w, http.StatusConflict, err.Error())
	default:
		writeGiantExecutorError(w, http.StatusInternalServerError, "giant material executor service error")
	}
}

func writeGiantExecutorError(w http.ResponseWriter, status int, message string) {
	writeGiantExecutorJSON(w, status, map[string]any{"error": message})
}

func writeGiantExecutorJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
