package httpapi

import (
	"net/http"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/batchfactory"
)

func NewRouter(starter BatchStarter) http.Handler {
	return NewRouterWithAgent(starter, nil, nil, nil)
}

func NewRouterWithIntakes(starter BatchStarter, intakeCreator IntakeCreator, ownerResolver OwnerResolver) http.Handler {
	return NewRouterWithAgent(starter, intakeCreator, ownerResolver, nil)
}

func NewRouterWithAgent(starter BatchStarter, intakeCreator IntakeCreator, ownerResolver OwnerResolver, agentAPI AgentAPI) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/api/batch-factory/jobs", NewBatchFactoryHandler(starter))
	mux.Handle("/api/batch-factory/intakes", NewIntakeHandler(intakeCreator, ownerResolver))
	runs := batchfactory.RunService{Intakes: intakeCreator, Starter: starter}
	mux.Handle("/api/batch-factory/runs", NewRunHandler(runs, ownerResolver))
	if agentAPI != nil {
		agentHandler := NewAgentHandler(agentAPI, ownerResolver)
		mux.Handle("/api/agent/threads", agentHandler)
		mux.Handle("/api/agent/threads/", agentHandler)
		mux.Handle("/api/agent/tasks", agentHandler)
	}
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	return mux
}
