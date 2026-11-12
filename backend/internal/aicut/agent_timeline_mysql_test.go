package aicut

import (
	"github.com/DATA-DOG/go-sqlmock"
	"testing"
	"time"
)

func TestMySQLAgentRunRetainsNamedTargetAcrossClaim(t *testing.T) {
	db, m, e := sqlmock.New()
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	now := time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)
	m.ExpectExec("UPDATE aicut_agent_runs SET state").WithArgs(AgentRunning, now, "run", AgentQueued).WillReturnResult(sqlmock.NewResult(0, 1))
	m.ExpectQuery("SELECT .* FROM aicut_agent_runs WHERE id").WithArgs("run").WillReturnRows(sqlmock.NewRows([]string{"id", "client_request_id", "owner_username", "project_id", "instruction", "mode", "model_id", "active_skill_id", "expected_revision", "state", "turn_count", "error_message", "cancel_requested_at", "created_at", "updated_at", "timeline_id"}).AddRow("run", "request", "alice", "p", "Explain", AgentModeAgent, "model", "", 1, AgentRunning, 0, "", nil, now, now, "named"))
	run, e := NewMySQLStore(db).ClaimAgentRun(t.Context(), "run", now)
	if e != nil || run.TimelineID != "named" || run.Owner != "alice" {
		t.Fatalf("run=%+v %v", run, e)
	}
	if e = m.ExpectationsWereMet(); e != nil {
		t.Fatal(e)
	}
}
