package novelfetch

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type ModuleHandler struct {
	service *Service
}

func NewModuleHandler(service *Service) http.Handler {
	handler := &ModuleHandler{service: service}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /batches", handler.createBatch)
	mux.HandleFunc("POST /batches/{id}/runs", handler.startRun)
	mux.HandleFunc("POST /runs/{id}/execute", handler.executeRun)
	mux.HandleFunc("POST /runs/{id}/books/{bookKey}/retry", handler.retryBook)
	mux.HandleFunc("GET /runs/{id}/records", handler.records)
	mux.HandleFunc("POST /runs/{id}/handoff", handler.handoff)
	mux.HandleFunc("POST /runs/{id}/books/{bookKey}/submit-intents", handler.submitIntent)
	mux.HandleFunc("GET /config", handler.getConfig)
	mux.HandleFunc("PUT /config", handler.saveConfig)
	mux.HandleFunc("GET /knowledge", handler.listKnowledge)
	mux.HandleFunc("PUT /knowledge/{id}", handler.upsertKnowledge)
	mux.HandleFunc("DELETE /knowledge/{id}", handler.deleteKnowledge)
	mux.HandleFunc("POST /rules/preview", handler.previewRules)
	return mux
}

func (h *ModuleHandler) createBatch(w http.ResponseWriter, r *http.Request) {
	var input CreateBatchInput
	if err := decodeBody(w, r, &input); err != nil {
		writeModuleError(w, err)
		return
	}
	batch, books, err := h.service.CreateBatch(r.Context(), input)
	if err != nil {
		writeModuleError(w, err)
		return
	}
	writeModuleJSON(w, http.StatusCreated, map[string]any{"batch": batch, "books": books})
}

func (h *ModuleHandler) startRun(w http.ResponseWriter, r *http.Request) {
	var body struct{ RunAt string `json:"runAt"` }
	if r.ContentLength != 0 {
		if err := decodeBody(w, r, &body); err != nil {
			writeModuleError(w, err)
			return
		}
	}
	var runAt time.Time
	if strings.TrimSpace(body.RunAt) != "" {
		parsed, err := time.Parse(time.RFC3339, body.RunAt)
		if err != nil {
			writeModuleError(w, ErrInvalid)
			return
		}
		runAt = parsed
	}
	run, err := h.service.StartRun(r.Context(), r.PathValue("id"), runAt)
	if err != nil {
		writeModuleError(w, err)
		return
	}
	writeModuleJSON(w, http.StatusCreated, map[string]any{"run": run})
}

func (h *ModuleHandler) executeRun(w http.ResponseWriter, r *http.Request) {
	run, err := h.service.ExecuteRun(r.Context(), r.PathValue("id"))
	if err != nil {
		writeModuleError(w, err)
		return
	}
	writeModuleJSON(w, http.StatusOK, map[string]any{"run": run})
}

func (h *ModuleHandler) retryBook(w http.ResponseWriter, r *http.Request) {
	bookKey, _ := url.PathUnescape(r.PathValue("bookKey"))
	book, err := h.service.RetryBook(r.Context(), r.PathValue("id"), bookKey)
	if err != nil {
		writeModuleError(w, err)
		return
	}
	writeModuleJSON(w, http.StatusOK, map[string]any{"book": book})
}

func (h *ModuleHandler) records(w http.ResponseWriter, r *http.Request) {
	records, err := h.service.Records(r.Context(), r.PathValue("id"))
	if err != nil {
		writeModuleError(w, err)
		return
	}
	writeModuleJSON(w, http.StatusOK, map[string]any{"records": records})
}

func (h *ModuleHandler) handoff(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.HandoffToBatchFactory(r.Context(), r.PathValue("id"))
	if err != nil {
		writeModuleError(w, err)
		return
	}
	writeModuleJSON(w, http.StatusAccepted, result)
}

func (h *ModuleHandler) submitIntent(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Version             string `json:"version"`
		PublishingAccountID int64  `json:"publishingAccountId"`
	}
	if err := decodeBody(w, r, &body); err != nil {
		writeModuleError(w, err)
		return
	}
	bookKey, _ := url.PathUnescape(r.PathValue("bookKey"))
	result, err := h.service.RequestNetworkSubmit(r.Context(), SubmitIntentRequest{
		RunID: r.PathValue("id"), BookKey: bookKey, Version: body.Version,
		PublishingAccountID: body.PublishingAccountID,
	})
	if err != nil {
		writeModuleError(w, err)
		return
	}
	writeModuleJSON(w, http.StatusAccepted, result)
}

func (h *ModuleHandler) getConfig(w http.ResponseWriter, r *http.Request) {
	config, err := h.service.GetConfig(r.Context())
	if err != nil {
		writeModuleError(w, err)
		return
	}
	writeModuleJSON(w, http.StatusOK, map[string]any{"config": config})
}

func (h *ModuleHandler) saveConfig(w http.ResponseWriter, r *http.Request) {
	var config Config
	if err := decodeBody(w, r, &config); err != nil {
		writeModuleError(w, err)
		return
	}
	config, err := h.service.SaveConfig(r.Context(), config)
	if err != nil {
		writeModuleError(w, err)
		return
	}
	writeModuleJSON(w, http.StatusOK, map[string]any{"config": config})
}

func (h *ModuleHandler) listKnowledge(w http.ResponseWriter, r *http.Request) {
	entries, err := h.service.ListKnowledge(r.Context(), r.URL.Query().Get("kind"))
	if err != nil {
		writeModuleError(w, err)
		return
	}
	writeModuleJSON(w, http.StatusOK, map[string]any{"items": entries})
}

func (h *ModuleHandler) upsertKnowledge(w http.ResponseWriter, r *http.Request) {
	var entry KnowledgeEntry
	if err := decodeBody(w, r, &entry); err != nil {
		writeModuleError(w, err)
		return
	}
	entry.ID = r.PathValue("id")
	entry, err := h.service.UpsertKnowledge(r.Context(), entry)
	if err != nil {
		writeModuleError(w, err)
		return
	}
	writeModuleJSON(w, http.StatusOK, map[string]any{"item": entry})
}

func (h *ModuleHandler) deleteKnowledge(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteKnowledge(r.Context(), r.URL.Query().Get("kind"), r.PathValue("id")); err != nil {
		writeModuleError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ModuleHandler) previewRules(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Text   string `json:"text"`
		Config Config `json:"config"`
	}
	if err := decodeBody(w, r, &body); err != nil {
		writeModuleError(w, err)
		return
	}
	writeModuleJSON(w, http.StatusOK, map[string]any{"text": h.service.PreviewRules(r.Context(), body.Text, body.Config)})
}

func decodeBody(w http.ResponseWriter, r *http.Request, target any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 2<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return ErrInvalid
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return ErrInvalid
	}
	return nil
}

func writeModuleError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, ErrInvalid):
		status = http.StatusBadRequest
	case errors.Is(err, ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, ErrNotDue):
		status = http.StatusConflict
	case errors.Is(err, ErrRuntimeUnavailable), errors.Is(err, ErrModelUnavailable), errors.Is(err, ErrPublishUnavailable), errors.Is(err, ErrHandoffUnavailable):
		status = http.StatusServiceUnavailable
	}
	writeModuleJSON(w, status, map[string]string{"error": safeError(err), "status": strconv.Itoa(status)})
}

func writeModuleJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
