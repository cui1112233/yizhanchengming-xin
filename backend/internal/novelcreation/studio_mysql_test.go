package novelcreation

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
	"os"
	"qiantie/backend/internal/storage"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestStudioMySQLCreationReservesWithoutProjectReceipt(t *testing.T) {
	st, m := newNovelCreationMockStore(t)
	ctx := context.Background()
	now := time.Now().UTC()
	r := studioRecord{CreateKey: "new", CreateFingerprint: "digest", Session: StudioSession{ID: "s", Mode: "long_original", CreatedAt: now, UpdatedAt: now}, Messages: []StudioMessage{}}
	raw, _ := json.Marshal(r)
	m.ExpectBegin()
	m.ExpectExec("INSERT INTO novel_studio_sessions .*ON DUPLICATE KEY UPDATE create_key=create_key").WithArgs("alice", "s", "new", "digest", sqlmock.AnyArg(), now, now).WillReturnResult(sqlmock.NewResult(1, 1))
	m.ExpectQuery("SELECT create_fingerprint,record_json FROM novel_studio_sessions WHERE owner_username=\\? AND create_key=\\? FOR UPDATE").WithArgs("alice", "new").WillReturnRows(sqlmock.NewRows([]string{"create_fingerprint", "record_json"}).AddRow("digest", raw))
	m.ExpectCommit()
	e := st.AtomicCreate(ctx, "alice", func(tx Store) error {
		saved, e := tx.ReserveStudioCreation(ctx, "alice", "new", "digest", raw)
		if e != nil {
			return e
		}
		var got studioRecord
		if e = json.Unmarshal(saved, &got); e != nil {
			return e
		}
		if got.Session.ProjectID != "" {
			t.Fatal("hidden project")
		}
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
	if e = m.ExpectationsWereMet(); e != nil {
		t.Fatal(e)
	}
}
func TestStudioMySQLRowLockAndRollback(t *testing.T) {
	st, m := newNovelCreationMockStore(t)
	ctx := context.Background()
	r := studioRecord{Session: StudioSession{ID: "s"}, Messages: []StudioMessage{}}
	raw, _ := json.Marshal(r)
	m.ExpectBegin()
	m.ExpectQuery("SELECT record_json FROM novel_studio_sessions WHERE owner_username=\\? AND studio_id=\\? FOR UPDATE").WithArgs("alice", "s").WillReturnRows(sqlmock.NewRows([]string{"record_json"}).AddRow(raw))
	m.ExpectExec("UPDATE novel_studio_sessions SET record_json=\\?,updated_at=\\? WHERE owner_username=\\? AND studio_id=\\?").WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "alice", "s").WillReturnResult(sqlmock.NewResult(0, 1))
	m.ExpectRollback()
	e := st.AtomicCreate(ctx, "alice", func(tx Store) error {
		got, e := readStudio(ctx, tx, "alice", "s", true)
		if e != nil {
			return e
		}
		got.Session.Title = "changed"
		if e = writeStudio(ctx, tx, "alice", got, false); e != nil {
			return e
		}
		return conflict()
	})
	if !IsCode(e, ErrIdempotencyConflict) {
		t.Fatal(e)
	}
	if e = m.ExpectationsWereMet(); e != nil {
		t.Fatal(e)
	}
}
func TestStudioMySQLWrongCreationFingerprintRollsBack(t *testing.T) {
	st, m := newNovelCreationMockStore(t)
	r := studioRecord{Session: StudioSession{ID: "s"}}
	raw, _ := json.Marshal(r)
	m.ExpectBegin()
	m.ExpectExec("INSERT INTO novel_studio_sessions").WillReturnResult(sqlmock.NewResult(0, 0))
	m.ExpectQuery("SELECT create_fingerprint,record_json.*FOR UPDATE").WithArgs("alice", "new").WillReturnRows(sqlmock.NewRows([]string{"create_fingerprint", "record_json"}).AddRow("old-digest", raw))
	m.ExpectRollback()
	e := st.AtomicCreate(context.Background(), "alice", func(tx Store) error {
		_, e := tx.ReserveStudioCreation(context.Background(), "alice", "new", "new-digest", raw)
		return e
	})
	if !IsCode(e, ErrIdempotencyConflict) {
		t.Fatal(e)
	}
	if e = m.ExpectationsWereMet(); e != nil {
		t.Fatal(e)
	}
}

// Real MySQL coverage uses only the existing opt-in, nonce-guarded disposable
// schema. No app database and no network model calls are involved.
func studioOwnedDB(t *testing.T) *sql.DB {
	t.Helper()
	raw := os.Getenv("QIANTIE_TEST_NOVEL_MYSQL_DSN")
	if raw == "" {
		if os.Getenv("QIANTIE_TEST_NOVEL_REQUIRE") == "1" {
			t.Fatal("owned MySQL DSN required")
		}
		t.Skip("opt-in owned studio MySQL acceptance")
	}
	config, e := mysql.ParseDSN(raw)
	schema := os.Getenv("QIANTIE_TEST_NOVEL_SCHEMA")
	addr := os.Getenv("QIANTIE_TEST_NOVEL_MYSQL_ADDR")
	nonce := os.Getenv("QIANTIE_TEST_NOVEL_NONCE")
	if e != nil || !regexp.MustCompile(`^novelparity20261117[a-f0-9]{12}$`).MatchString(schema) || !regexp.MustCompile(`^127\.0\.0\.1:[0-9]{4,5}$`).MatchString(addr) || !regexp.MustCompile(`^[a-f0-9]{32}$`).MatchString(nonce) || config.DBName != schema || config.Net != "tcp" || config.Addr != addr || config.User != "np"+strings.TrimPrefix(schema, "novelparity20261117") {
		t.Fatal("refusing non-owned MySQL studio target")
	}
	db, e := sql.Open("mysql", config.FormatDSN())
	if e != nil {
		t.Fatal("open owned MySQL failed")
	}
	db.SetMaxOpenConns(5)
	t.Cleanup(func() { db.Close() })
	var actual, version string
	if e = db.QueryRow("SELECT DATABASE(),VERSION()").Scan(&actual, &version); e != nil || actual != schema || !strings.HasPrefix(version, "8.4.") {
		t.Fatal("wrong owned MySQL database/version")
	}
	// This marker is the existing runner's cleanup ownership proof. No app or
	// unrelated feature schema is needed for studio/novel-domain persistence.
	if _, e = db.Exec(`CREATE TABLE IF NOT EXISTS novel_parity_fixture_owner (task_id VARCHAR(80) PRIMARY KEY, nonce VARCHAR(100) NOT NULL, source_sha VARCHAR(64) NOT NULL, created_at DATETIME(6) NOT NULL)`); e != nil {
		t.Fatal(e)
	}
	if _, e = db.Exec(`INSERT IGNORE INTO novel_parity_fixture_owner VALUES ('novel-parity-task8',?,?,UTC_TIMESTAMP(6))`, nonce, os.Getenv("QIANTIE_TEST_NOVEL_SOURCE_SHA")); e != nil {
		t.Fatal(e)
	}
	var savedNonce string
	if e = db.QueryRow(`SELECT nonce FROM novel_parity_fixture_owner WHERE task_id='novel-parity-task8'`).Scan(&savedNonce); e != nil || savedNonce != nonce {
		t.Fatal("owner nonce mismatch")
	}
	for pass := 1; pass <= 2; pass++ {
		if e = storage.RunMigrations(context.Background(), db, storage.NovelCreationMigrations()); e != nil {
			t.Fatal(e)
		}
		t.Logf("owned studio schema=%s MySQL=%s novel-only migration-pass=%d", schema, version, pass)
	}
	return db
}
func TestStudioRealMySQLOwnerIsolationAndRestart(t *testing.T) {
	db := studioOwnedDB(t)
	st := NewMySQLStore(db)
	p := &studioTestProvider{reply: `{"message":"先确认设定","action":{"type":"create_book","book":{"title":"隔离验收作品","genre":"幻想","world":"测试世界","protagonist":"主角","coreConflict":"冲突","targetChapterWords":2000}}}`}
	s := NewService(st, NewMemoryRuntime(), studioTestPipeline{provider: p})
	ctx := context.Background()
	key := novelID("studio-test")
	ss, e := s.CreateStudioSession(ctx, key, StudioCreateSessionInput{IdempotencyKey: "create", Mode: "long_original"})
	if e != nil {
		t.Fatal(e)
	}
	same, e := s.CreateStudioSession(ctx, key, StudioCreateSessionInput{IdempotencyKey: "create", Mode: "long_original"})
	if e != nil || same.ID != ss.ID {
		t.Fatal("creation replay", e)
	}
	if _, e = s.GetStudioSnapshot(ctx, key+"-other", ss.ID); e == nil {
		t.Fatal("cross owner read")
	}
	in := StudioSendInput{IdempotencyKey: "chat", Content: "讨论世界设定", ModelID: "test-only"}
	if _, e = s.SendStudioMessage(ctx, key, ss.ID, in); e != nil {
		t.Fatal(e)
	}
	reopened := NewService(NewMySQLStore(db), NewMemoryRuntime(), studioTestPipeline{provider: p})
	snap, e := reopened.SendStudioMessage(ctx, key, ss.ID, in)
	if e != nil || snap.Turn.State != "completed" || p.calls != 1 {
		t.Fatalf("restart: %+v %v", snap, e)
	}
	projects, e := st.ListProjects(ctx, key)
	if e != nil || len(projects) != 0 {
		t.Fatal("hidden project", e)
	}
	if snap.PendingAction == nil {
		t.Fatal("pending action missing after database reopen")
	}
	fail := true
	faulted := NewService(studioFaultStore{Store: NewMySQLStore(db), state: "confirmed", fail: &fail}, NewMemoryRuntime(), studioTestPipeline{provider: p})
	a := StudioActionInput{ActionID: snap.PendingAction.ID, ModelID: "test-only"}
	if _, e = faulted.ConfirmStudioAction(ctx, key, ss.ID, a); e == nil {
		t.Fatal("expected final save failure")
	}
	projects, e = st.ListProjects(ctx, key)
	if e != nil || len(projects) != 0 {
		t.Fatal("real SQL rollback left a hidden project", e)
	}
	out, e := reopened.ConfirmStudioAction(ctx, key, ss.ID, a)
	if e != nil {
		t.Fatal(e)
	}
	third := NewService(NewMySQLStore(db), NewMemoryRuntime(), studioTestPipeline{provider: p})
	again, e := third.ConfirmStudioAction(ctx, key, ss.ID, a)
	if e != nil || again.Project.ID != out.Project.ID || again.Run.ID != out.Run.ID {
		t.Fatal("confirmation replay after restart", e)
	}
	projects, e = st.ListProjects(ctx, key)
	if e != nil || len(projects) != 1 {
		t.Fatal("duplicate book", e)
	}
	ex, ok, e := st.GetExecution(ctx, key, out.Run.ID)
	if e != nil || !ok || ex.State != "ready" {
		t.Fatal("missing durable queued production", e)
	}
	if p.calls != 1 {
		t.Fatal("real SQL recovery rebilled planner")
	}
}
