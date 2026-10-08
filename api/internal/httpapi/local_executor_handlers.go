package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/video"
)

func (h handler) registerVideoLocalExecutor(w http.ResponseWriter, r *http.Request) {
	// Bootstrap authentication is retained only for rollout compatibility. It
	// cannot establish an owner, so new registrations must redeem a pairing.
	writeJSON(w, http.StatusGone, map[string]any{"error": "local_executor_pairing_required", "message": "请通过安全配对意图注册本地执行器"})
}

func (h handler) listVideoLocalExecutors(w http.ResponseWriter, r *http.Request) {
	if h.deps.VideoLocalExecutor == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "local_executor_unavailable", "message": "local executor service unavailable"})
		return
	}
	user, ok := authn.CurrentUser(r.Context())
	if !ok || user.ID <= 0 {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"code": "AUTH_UNAUTHENTICATED", "message": "登录状态无效或已过期"})
		return
	}
	items, err := h.deps.VideoLocalExecutor.ListForOwner(r.Context(), user.ID)
	if err != nil {
		writeLocalExecutorError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"executors": items})
}

func (h handler) createVideoLocalExecutorPairing(w http.ResponseWriter, r *http.Request) {
	if h.deps.VideoLocalExecutor == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "local_executor_unavailable", "message": "local executor service unavailable"})
		return
	}
	user, ok := authn.CurrentUser(r.Context())
	if !ok || user.ID <= 0 {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"code": "AUTH_UNAUTHENTICATED", "message": "登录状态无效或已过期"})
		return
	}
	pairing, err := h.deps.VideoLocalExecutor.CreatePairingIntent(r.Context(), user.ID)
	if err != nil {
		writeLocalExecutorError(w, err)
		return
	}
	// The browser receives a QR/deep-link payload only. It never receives an
	// executor credential and should render this value without exposing text.
	writeJSON(w, http.StatusCreated, map[string]any{"deepLink": "ycm-executor://pair?intent=" + pairing.Payload, "expiresAt": pairing.ExpiresAt})
}

func (h handler) redeemVideoLocalExecutorPairing(w http.ResponseWriter, r *http.Request) {
	if h.deps.VideoLocalExecutor == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "local_executor_unavailable", "message": "local executor service unavailable"})
		return
	}
	defer r.Body.Close()
	var input struct {
		PairingPayload string   `json:"pairingPayload"`
		Name           string   `json:"name"`
		ProviderKey    string   `json:"providerKey"`
		Model          string   `json:"model"`
		Capabilities   []string `json:"capabilities"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxRequestBody)).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_request", "message": "invalid JSON body"})
		return
	}
	result, err := h.deps.VideoLocalExecutor.RedeemPairingIntent(r.Context(), input.PairingPayload, video.LocalExecutorRegistrationInput{Name: input.Name, ProviderKey: input.ProviderKey, Model: input.Model, Capabilities: input.Capabilities})
	if err != nil {
		writeLocalExecutorError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

func (h handler) unbindVideoLocalExecutor(w http.ResponseWriter, r *http.Request) {
	if h.deps.VideoLocalExecutor == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "local_executor_unavailable", "message": "local executor service unavailable"})
		return
	}
	user, ok := authn.CurrentUser(r.Context())
	if !ok || user.ID <= 0 {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"code": "AUTH_UNAUTHENTICATED", "message": "登录状态无效或已过期"})
		return
	}
	if err := h.deps.VideoLocalExecutor.UnbindForOwner(r.Context(), user.ID, r.PathValue("executorId")); err != nil {
		writeLocalExecutorError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h handler) getVideoLocalExecutorIdentity(w http.ResponseWriter, r *http.Request) {
	if h.deps.VideoLocalExecutor == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "local_executor_unavailable", "message": "local executor service unavailable"})
		return
	}
	identity, err := h.deps.VideoLocalExecutor.Identity(r.Context(), executorBearerToken(r))
	if err != nil {
		writeLocalExecutorError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, identity)
}

func (h handler) heartbeatVideoLocalExecutor(w http.ResponseWriter, r *http.Request) {
	if h.deps.VideoLocalExecutor == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "local_executor_unavailable", "message": "local executor service unavailable"})
		return
	}
	defer r.Body.Close()
	var input video.LocalExecutorHeartbeatInput
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxRequestBody)).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_request", "message": "invalid JSON body"})
		return
	}
	if err := h.deps.VideoLocalExecutor.Heartbeat(r.Context(), executorBearerToken(r), input); err != nil {
		writeLocalExecutorError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h handler) completeVideoLocalExecutorTask(w http.ResponseWriter, r *http.Request) {
	if h.deps.VideoLocalExecutor == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "local_executor_unavailable", "message": "local executor service unavailable"})
		return
	}
	defer r.Body.Close()
	var input video.LocalExecutorCompleteInput
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxRequestBody)).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_request", "message": "invalid JSON body"})
		return
	}
	if err := h.deps.VideoLocalExecutor.CompleteTask(r.Context(), executorBearerToken(r), r.PathValue("taskId"), input); err != nil {
		writeLocalExecutorError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h handler) failVideoLocalExecutorTask(w http.ResponseWriter, r *http.Request) {
	if h.deps.VideoLocalExecutor == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "local_executor_unavailable", "message": "local executor service unavailable"})
		return
	}
	defer r.Body.Close()
	var input video.LocalExecutorFailInput
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxRequestBody)).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_request", "message": "invalid JSON body"})
		return
	}
	if err := h.deps.VideoLocalExecutor.FailTask(r.Context(), executorBearerToken(r), r.PathValue("taskId"), input); err != nil {
		writeLocalExecutorError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func executorBearerToken(r *http.Request) string {
	value := strings.TrimSpace(r.Header.Get("Authorization"))
	parts := strings.SplitN(value, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return strings.TrimSpace(parts[1])
	}
	return value
}

func writeLocalExecutorError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, video.ErrLocalExecutorUnauthorized):
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "local_executor_unauthorized", "message": "local executor credential is invalid"})
	case errors.Is(err, video.ErrLocalExecutorInvalid):
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "local_executor_invalid", "message": "invalid local executor request"})
	case errors.Is(err, video.ErrLocalExecutorTaskNotFound):
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "local_executor_task_not_found", "message": "local executor task not found"})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "local_executor_internal", "message": "local executor operation failed"})
	}
}
