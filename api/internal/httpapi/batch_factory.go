package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/batchfactory"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/pipeline"
)

type BatchStarter interface {
	Start(context.Context, batchfactory.StartInput) (pipeline.Job, error)
}

type batchFactoryHandler struct { starter BatchStarter }

type createJobRequest struct {
	IntakeID string `json:"intake_id"`
	RunAt    string `json:"run_at"`
}

func NewBatchFactoryHandler(starter BatchStarter) http.Handler { return &batchFactoryHandler{starter: starter} }

func (h *batchFactoryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { w.WriteHeader(http.StatusMethodNotAllowed); return }
	if h.starter == nil { http.Error(w, "batch starter unavailable", http.StatusServiceUnavailable); return }

	var input createJobRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil { http.Error(w, "invalid request body", http.StatusBadRequest); return }
	if strings.TrimSpace(input.IntakeID) == "" { http.Error(w, "intake_id is required", http.StatusBadRequest); return }

	var runAt time.Time
	if strings.TrimSpace(input.RunAt) != "" {
		parsed, err := time.Parse(time.RFC3339, input.RunAt)
		if err != nil { http.Error(w, "run_at must be RFC3339", http.StatusBadRequest); return }
		runAt = parsed
	}
	job, err := h.starter.Start(r.Context(), batchfactory.StartInput{IntakeID: input.IntakeID, RunAt: runAt})
	if err != nil { http.Error(w, err.Error(), http.StatusBadRequest); return }

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{"job_id": job.ID, "status": job.Status, "run_at": job.RunAt})
}
