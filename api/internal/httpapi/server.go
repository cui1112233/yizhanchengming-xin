package httpapi

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/intake"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/pipeline"
)

type IntakeService interface {
	CreateIntake(ctx context.Context, input intake.CreateIntakeInput) (intake.Intake, []intake.Book, error)
	ExecuteIntake(ctx context.Context, intakeID int64, maxText int) (intake.ExecuteResult, error)
}

type IntakeReader interface {
	ListIntakes(ctx context.Context) ([]intake.Intake, error)
	ListBooks(ctx context.Context, intakeID int64) ([]intake.Book, error)
}

type BatchProjectReader interface {
	ListBatchProjects(ctx context.Context) ([]intake.BatchProject, error)
}

type BatchProjectDetailReader interface {
	GetBatchProject(ctx context.Context, id int64) (intake.BatchProject, error)
	ListBooks(ctx context.Context, intakeID int64) ([]intake.Book, error)
}

type PipelineService interface {
	Create(ctx context.Context, request pipeline.CreateRequest) (pipeline.CreateResult, error)
}

type Dependencies struct {
	Intakes             IntakeService
	Reader              IntakeReader
	BatchProjects       BatchProjectReader
	BatchProjectDetails BatchProjectDetailReader
	Pipeline            PipelineService
}

func NewHandler(values ...Dependencies) http.Handler {
	var deps Dependencies
	if len(values) > 0 {
		deps = values[0]
	}

	api := handler{deps: deps}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("POST /api/v1/intakes", api.createIntake)
	mux.HandleFunc("GET /api/v1/intakes", api.listIntakes)
	mux.HandleFunc("POST /api/v1/intakes/{id}/execute", api.executeIntake)
	mux.HandleFunc("GET /api/v1/intakes/{id}/books", api.listBooks)
	mux.HandleFunc("POST /api/v1/intakes/{id}/batch-projects", api.createBatchProject)
	mux.HandleFunc("GET /api/v1/batch-projects", api.listBatchProjects)
	mux.HandleFunc("GET /api/v1/batch-projects/{id}", api.getBatchProject)
	return mux
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
