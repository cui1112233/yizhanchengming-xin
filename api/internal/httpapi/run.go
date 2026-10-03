package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/batchfactory"
)

type RunCreator interface {
	CreateAndStart(context.Context, batchfactory.CreateRunInput) (batchfactory.RunResult, error)
}

type runHandler struct {
	creator RunCreator
	owner   OwnerResolver
}

type createRunRequest struct {
	Title  string               `json:"title"`
	RunAt  *time.Time           `json:"run_at,omitempty"`
	Groups []intakeGroupRequest `json:"groups"`
}

func NewRunHandler(creator RunCreator, owner OwnerResolver) http.Handler {
	return &runHandler{creator: creator, owner: owner}
}

func (h *runHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if h.creator == nil || h.owner == nil {
		http.Error(w, "run service unavailable", http.StatusServiceUnavailable)
		return
	}
	owner, err := h.owner(r)
	if err != nil || strings.TrimSpace(owner) == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var request createRunRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	groups, err := convertIntakeGroups(request.Groups)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	input := batchfactory.CreateRunInput{
		Intake: batchfactory.CreateIntakeInput{
			Owner:  strings.TrimSpace(owner),
			Title:  request.Title,
			Groups: groups,
		},
	}
	if request.RunAt != nil {
		input.RunAt = request.RunAt.UTC()
	}
	result, err := h.creator.CreateAndStart(r.Context(), input)
	if err != nil {
		if result.IntakeID != "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(struct {
				batchfactory.RunResult
				Error string `json:"error"`
			}{RunResult: result, Error: err.Error()})
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(result)
}
