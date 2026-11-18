package novelcreation

import (
	"context"
	"encoding/json"
	"github.com/DATA-DOG/go-sqlmock"
	"strings"
	"testing"
)

// Opt-in: studioOwnedDB verifies an isolated nonce/schema/user and MySQL 8.4,
// and applies NovelCreationMigrations only. Every model response is a local fake.
func TestLongImitationRealMySQLOwnerCreateRollbackReplaySummary(t *testing.T) {
	db := studioOwnedDB(t)
	ctx := context.Background()
	owner := novelID("imitation-sql")
	st := NewMySQLStore(db)
	provider := &queuedPipelineProvider{responses: []string{imitationAnalysisJSON(), foundationJSON}}
	s := NewService(st, NewMemoryRuntime(), pipelineWithProvider(provider))
	in := imitationInput("sql-create")
	out, e := s.StartLongImitation(ctx, owner, in)
	if e != nil {
		t.Fatal(e)
	}
	if len(provider.requests) != 0 || out.Project.TargetChapterWords != 50000 {
		t.Fatal("creation called a model or clamped requested target")
	}
	config, e := s.GetLongImitation(ctx, owner, out.Project.ID)
	if e != nil || config.Input.StoryIdea != in.StoryIdea || len(config.ReferenceChunks) != 1 {
		t.Fatal("full SQL configuration round trip", e)
	}
	summary, _ := json.Marshal(LongImitationSummary(config))
	if strings.Contains(string(summary), in.ReferenceText) || !strings.Contains(string(summary), `"hasReference":true`) {
		t.Fatal("summary contains raw reference or lacks metadata")
	}
	ex, ok, e := st.GetExecution(ctx, owner, out.Run.ID)
	if e != nil || !ok || ex.State != "ready" {
		t.Fatal("durable SQL dispatch missing", e)
	}
	again, e := s.StartLongImitation(ctx, owner, in)
	if e != nil || again.Run.ID != out.Run.ID || again.Project.ID != out.Project.ID {
		t.Fatal("POST replay changed SQL identity", e)
	}
	recovered, e := s.GetLongImitationRequest(ctx, owner, in.IdempotencyKey)
	if e != nil || recovered.Run.ID != out.Run.ID {
		t.Fatal("read-only request recovery", e)
	}
	if _, e = s.GetLongImitationRequest(ctx, owner+"other", in.IdempotencyKey); !IsCode(e, ErrProjectNotFound) {
		t.Fatal("cross owner receipt exposed", e)
	}
	fail := true
	faulted := NewService(imitationCommitFault{Store: NewMySQLStore(db), fail: &fail}, NewMemoryRuntime(), nil)
	failedInput := in
	failedInput.IdempotencyKey = "sql-rollback"
	failed, e := faulted.StartLongImitation(ctx, owner, failedInput)
	if e == nil {
		t.Fatal("expected final receipt failure")
	}
	if _, e = st.GetProject(ctx, owner, failed.Project.ID); !IsCode(e, ErrProjectNotFound) {
		t.Fatal("project survived SQL rollback", e)
	}
	if _, found, e := st.GetExecution(ctx, owner, failed.Run.ID); e != nil || found {
		t.Fatal("dispatch survived SQL rollback", e)
	}
	if _, e = s.GetLongImitationRequest(ctx, owner, failedInput.IdempotencyKey); !IsCode(e, ErrProjectNotFound) {
		t.Fatal("receipt survived SQL rollback", e)
	}
	projects, e := st.ListProjects(ctx, owner)
	if e != nil || len(projects) != 1 {
		t.Fatal("duplicate/orphan SQL project", e)
	}
	checkpoint, e := s.Run(ctx, owner, out.Run.ID)
	if e != nil || checkpoint.Status != RunAwaitingConfirmation {
		t.Fatalf("fake-provider SQL blueprint failed status=%s err=%v", checkpoint.Status, e)
	}
	project, e := st.GetProject(ctx, owner, out.Project.ID)
	if e != nil || project.Title != "雨幕记者" {
		t.Fatal("SQL automatic book naming did not persist", e)
	}
	session, e := st.GetSession(ctx, owner, out.Session.ID)
	if e != nil || session.Title != "雨幕记者" {
		t.Fatal("SQL automatic session naming did not persist", e)
	}
	config, e = s.GetLongImitation(ctx, owner, project.ID)
	if e != nil || config.Analysis == nil || len(config.Analysis.Dimensions) != 8 || len(provider.requests) != 2 {
		t.Fatal("SQL analysis/config recovery missing or repeated model stage", e)
	}
	t.Log("isolated SQL create/config/summary/replay/rollback/analysis/automatic-name verified; local fake provider calls=2, network model calls=0")
	t.Run("full-length external key", func(t *testing.T) {
		longInput := in
		longInput.IdempotencyKey = strings.Repeat("键", 128)
		created, e := s.StartLongImitation(ctx, owner, longInput)
		if e != nil || len(created.Run.IdempotencyKey) > 96 {
			t.Fatal("long external key SQL creation failed", e)
		}
		saved, e := s.GetLongImitationRequest(ctx, owner, longInput.IdempotencyKey)
		if e != nil || saved.Run.ID != created.Run.ID {
			t.Fatal("long-key SQL receipt lookup failed", e)
		}
		stored, e := st.GetRun(ctx, owner, created.Run.ID)
		if e != nil || stored.IdempotencyKey != created.Run.IdempotencyKey {
			t.Fatal("long-key SQL run did not retain hash", e)
		}
	})
}

func TestLongImitationMySQLReceivesBoundedHashForFullLengthExternalKey(t *testing.T) {
	memory := NewService(NewMemoryStore(), NewMemoryRuntime(), nil)
	input := imitationInput(strings.Repeat("键", 128))
	created, e := memory.StartLongImitation(context.Background(), "alice", input)
	if e != nil {
		t.Fatal(e)
	}
	run := created.Run
	if len(run.IdempotencyKey) > 96 {
		t.Fatal("unbounded key passed to SQL")
	}
	st, m := newNovelCreationMockStore(t)
	m.ExpectExec("INSERT INTO novel_runs").WithArgs(run.ID, "alice", run.ProjectID, run.SessionID, run.IdempotencyKey, run.Status, run.ModelID, run.ThinkingLevel, sqlmock.AnyArg(), nil, "", nil, sqlmock.AnyArg(), sqlmock.AnyArg(), nil, "alice", run.ProjectID, "alice", run.SessionID, run.ProjectID).WillReturnResult(sqlmock.NewResult(1, 1))
	m.ExpectExec("UPDATE novel_sessions SET current_run_id").WithArgs(run.ID, sqlmock.AnyArg(), "alice", run.SessionID, run.ProjectID).WillReturnResult(sqlmock.NewResult(1, 1))
	saved, e := st.CreateRun(context.Background(), "alice", run)
	if e != nil || saved.IdempotencyKey != run.IdempotencyKey {
		t.Fatal("hash not delivered to SQL insert", e)
	}
	if e = m.ExpectationsWereMet(); e != nil {
		t.Fatal(e)
	}
}

func TestLongImitationMySQLCreatePreservesFiftyThousandTarget(t *testing.T) {
	st, m := newNovelCreationMockStore(t)
	m.ExpectExec("INSERT INTO novel_projects").WithArgs(sqlmock.AnyArg(), "alice", "待蓝图命名", "other", "own", `{"creationMode":"long_imitation"}`, 50000, "draft", sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	p, e := st.CreateLongImitationProject(context.Background(), "alice", "待蓝图命名", "own", 50000)
	if e != nil || p.TargetChapterWords != 50000 {
		t.Fatal("special mode word target lost", e)
	}
	if e = m.ExpectationsWereMet(); e != nil {
		t.Fatal(e)
	}
}
func TestLongImitationMySQLRecoveryLookupIsReadOnlyOwnerScoped(t *testing.T) {
	st, m := newNovelCreationMockStore(t)
	raw, _ := json.Marshal(LongImitationResult{Project: Project{ID: "p"}, Run: Run{ID: "r"}})
	m.ExpectQuery("SELECT fingerprint,project_id,result_json FROM novel_submission_receipts WHERE owner_username=\\? AND kind=\\? AND idempotency_key=\\?").WithArgs("alice", "long_imitation", "request").WillReturnRows(sqlmock.NewRows([]string{"fingerprint", "project_id", "result_json"}).AddRow("f", "p", raw))
	s := NewService(st, nil, nil)
	out, e := s.GetLongImitationRequest(context.Background(), "alice", "request")
	if e != nil || out.Run.ID != "r" {
		t.Fatal(e)
	}
	if e = m.ExpectationsWereMet(); e != nil {
		t.Fatal(e)
	}
}
