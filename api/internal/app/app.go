package app

import (
	"database/sql"
	"net/http"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/httpapi"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/intake"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/pipeline"
)

// NewHandler wires the production intake and pipeline services to the same
// MySQL-backed store. Immediate and scheduled runs therefore share the same
// source of truth; only pipeline.CreateRequest.RunAt differs.
func NewHandler(db *sql.DB, fetcher intake.Fetcher, classifier intake.Classifier, now pipeline.Clock) http.Handler {
	store := intake.NewMySQLStore(db)
	intakeService := intake.NewService(store, fetcher, classifier)
	pipelineService := pipeline.NewService(store, now)

	return httpapi.NewHandler(httpapi.Dependencies{
		Intakes:             intakeService,
		Reader:              store,
		Pipeline:            pipelineService,
		BatchProjects:       store,
		BatchProjectDetails: store,
	})
}
