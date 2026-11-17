package novelcreation

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/go-sql-driver/mysql"
	"qiantie/backend/internal/projectretention"
	"qiantie/backend/internal/storage"
)

// An independent, opt-in old-schema -> new-schema fixture. It accepts only the
// nonce-labelled, loopback, task-owned schema, never the daily 3306 service.
func chapterImportOwnedDB(t *testing.T) *sql.DB {
	t.Helper()
	raw := os.Getenv("QIANTIE_TEST_NOVEL_MYSQL_DSN")
	if raw == "" {
		t.Skip("opt-in owned MySQL acceptance")
	}
	c, e := mysql.ParseDSN(raw)
	schema := os.Getenv("QIANTIE_TEST_NOVEL_SCHEMA")
	addr := os.Getenv("QIANTIE_TEST_NOVEL_MYSQL_ADDR")
	nonce := os.Getenv("QIANTIE_TEST_NOVEL_NONCE")
	if e != nil || !regexp.MustCompile(`^novelparity20261117[a-f0-9]{12}$`).MatchString(schema) || !regexp.MustCompile(`^127\.0\.0\.1:[0-9]{4,5}$`).MatchString(addr) || strings.HasSuffix(addr, ":3306") || c.Net != "tcp" || c.Addr != addr || c.DBName != schema || c.User != "np"+strings.TrimPrefix(schema, "novelparity20261117") || !regexp.MustCompile(`^[a-f0-9]{32}$`).MatchString(nonce) {
		t.Fatal("refusing non-owned MySQL target")
	}
	db, e := sql.Open("mysql", raw)
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { db.Close() })
	db.SetMaxOpenConns(8)
	var actual, version string
	if e = db.QueryRow(`SELECT DATABASE(),VERSION()`).Scan(&actual, &version); e != nil || actual != schema || !strings.HasPrefix(version, "8.4.") {
		t.Fatal("wrong owned database/version")
	}
	if _, e = db.Exec(`CREATE TABLE IF NOT EXISTS novel_parity_fixture_owner (task_id VARCHAR(80) PRIMARY KEY, nonce VARCHAR(100) NOT NULL, source_sha VARCHAR(64) NOT NULL, created_at DATETIME(6) NOT NULL)`); e != nil {
		t.Fatal(e)
	}
	if _, e = db.Exec(`INSERT IGNORE INTO novel_parity_fixture_owner VALUES ('novel-parity-task8',?,?,UTC_TIMESTAMP(6))`, nonce, os.Getenv("QIANTIE_TEST_NOVEL_SOURCE_SHA")); e != nil {
		t.Fatal(e)
	}
	var saved string
	if e = db.QueryRow(`SELECT nonce FROM novel_parity_fixture_owner WHERE task_id='novel-parity-task8'`).Scan(&saved); e != nil || saved != nonce {
		t.Fatal("owner nonce mismatch")
	}
	t.Logf("owned-schema=%s MySQL=%s address=%s nonce-verified", schema, version, addr)
	return db
}

func TestChapterImportRealMySQL(t *testing.T) {
	db := chapterImportOwnedDB(t)
	ctx := context.Background()
	all := append(storage.AppMigrations(), storage.NovelFetchWorkshopMigrations()...)
	all = append(all, storage.AgentProxyMigrations()...)
	all = append(all, storage.AuthMigrations()...)
	all = append(all, storage.ScreenplayMigrations()...)
	novel := storage.NovelCreationMigrations()
	old := append(all, novel[:len(novel)-1]...)
	if e := storage.RunMigrations(ctx, db, old); e != nil {
		t.Fatal(e)
	}
	st := NewMySQLStore(db)
	s := chapterImportService(st)
	owner := novelID("import-owner")
	p, e := st.CreateProject(ctx, owner, CreateProjectInput{Title: "旧书", Genre: "other", ModelID: "no-network-model"})
	if e != nil {
		t.Fatal(e)
	}
	se, e := st.CreateSession(ctx, owner, CreateSessionInput{ProjectID: p.ID, Mode: "long_original"})
	if e != nil {
		t.Fatal(e)
	}
	legacyPrompt := json.RawMessage(`{"legacy":"  unchanged 中文", "version":1}`)
	legacyRuntime := json.RawMessage(`{"journal":"unknown", "zero":0}`)
	r, e := st.CreateRun(ctx, owner, Run{ID: novelID("run"), ProjectID: p.ID, SessionID: se.ID, IdempotencyKey: "legacy", Status: RunSucceeded, PromptSnapshot: legacyPrompt, RuntimeSnapshot: legacyRuntime})
	if e != nil {
		t.Fatal(e)
	}
	b := serviceTestBundle(RunContext{Run: r, Project: p, Session: se, ChapterNumber: 1}, RunSucceeded)
	b.Chapter.Number = 8
	b.Chapter.ID = novelID("chapter")
	b.ChapterVersion.ChapterID = b.Chapter.ID
	if e = st.SaveChapterBundle(ctx, owner, b); e != nil {
		t.Fatal(e)
	}
	readLegacy := func() string {
		var prompt, runtime, version []byte
		if e := db.QueryRow(`SELECT prompt_snapshot_json,runtime_snapshot_json FROM novel_runs WHERE owner_username=? AND run_id=?`, owner, r.ID).Scan(&prompt, &runtime); e != nil {
			t.Fatal(e)
		}
		if e := db.QueryRow(`SELECT snapshot_json FROM novel_chapter_versions WHERE owner_username=? AND chapter_id=?`, owner, b.Chapter.ID).Scan(&version); e != nil {
			t.Fatal(e)
		}
		return string(prompt) + string(runtime) + string(version)
	}
	before := readLegacy()
	all = append(all, novel...)
	for pass := 1; pass <= 2; pass++ {
		if e = storage.RunMigrations(ctx, db, all); e != nil {
			t.Fatal(e)
		}
		if readLegacy() != before {
			t.Fatal("migration mutated legacy JSON")
		}
		t.Logf("migration-pass=%d legacy-JSON-bytes-unchanged", pass)
	}
	for _, m := range novel {
		var sum string
		if e = db.QueryRow(`SELECT checksum FROM schema_migrations WHERE version=?`, m.Version).Scan(&sum); e != nil || sum != storage.ChecksumFor(m) {
			t.Fatalf("checksum %d %v", m.Version, e)
		}
	}
	in := chapterImportInput("actual")
	in.BookID = p.ID
	in.Title = ""
	got, e := s.ImportChapters(ctx, owner, in)
	if e != nil {
		t.Fatal(e)
	}
	if len(got.Chapters) != 2 || got.Chapters[0].Number != 9 || got.Chapters[1].Number != 10 {
		t.Fatalf("%+v", got)
	}
	reopened := chapterImportService(NewMySQLStore(db))
	listed, e := reopened.store.ListChapters(ctx, owner, p.ID)
	if e != nil || len(listed) != 3 {
		t.Fatalf("reopen=%+v %v", listed, e)
	}
	state, e := reopened.GetContinuityState(ctx, owner, p.ID)
	if e != nil || state.DirtyFrom != 9 || state.Revision != 1 {
		t.Fatalf("continuity=%+v %v", state, e)
	}
	for _, c := range got.Chapters {
		vs, e := st.ListChapterVersions(ctx, owner, p.ID, c.ID)
		if e != nil || len(vs) != 1 || vs[0].Version != 1 || vs[0].RunID != "" || vs[0].Content != c.Content {
			t.Fatalf("versions=%+v %v", vs, e)
		}
	}
	for _, format := range []string{"txt", "md"} {
		doc, e := reopened.ExportProject(ctx, owner, p.ID, format)
		if e != nil || !strings.Contains(doc.Content, "故人等候。") {
			t.Fatalf("export=%+v %v", doc, e)
		}
	}
	if readLegacy() != before {
		t.Fatal("import mutated legacy JSON")
	}
	// Concurrent exact replays must resolve to the original receipt; distinct
	// keys append under the same project lock and never allocate duplicate numbers.
	var wg sync.WaitGroup
	errs := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, e := reopened.ImportChapters(ctx, owner, in)
			if e == nil && fingerprint(v) != fingerprint(got) {
				e = fmt.Errorf("different replay")
			}
			errs <- e
		}()
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		if e != nil {
			t.Fatal(e)
		}
	}
	in.Chapters = append([]ImportedChapterInput(nil), in.Chapters...)
	// Independent keys contend for the same project and allocate contiguous IDs.
	errs = make(chan error, 4)
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			request := chapterImportInput(fmt.Sprintf("append-%d", i))
			request.BookID = p.ID
			request.Title = ""
			_, e := reopened.ImportChapters(ctx, owner, request)
			errs <- e
		}(i)
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		if e != nil {
			t.Fatal(e)
		}
	}
	listed, e = st.ListChapters(ctx, owner, p.ID)
	if e != nil || len(listed) != 11 {
		t.Fatalf("distinct key chapters=%d %v", len(listed), e)
	}
	for i, c := range listed {
		if c.Number != 8+i {
			t.Fatalf("non-contiguous number %d at %d", c.Number, i)
		}
	}
	// Actual constraints, with test-only placeholder rows rolled back afterward.
	constraintTx, e := db.BeginTx(ctx, nil)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = constraintTx.Exec(`INSERT INTO novel_submission_receipts (owner_username,kind,idempotency_key,fingerprint,created_at) VALUES (?,'chapter_import','actual',?,UTC_TIMESTAMP(6))`, owner, strings.Repeat("a", 64)); e == nil {
		t.Fatal("receipt unique constraint missing")
	}
	if _, e = constraintTx.Exec(`INSERT INTO novel_submission_receipts (owner_username,kind,idempotency_key,fingerprint,project_id,result_json,created_at) VALUES (?,'chapter_import','foreign-fk',?,?,'{}',UTC_TIMESTAMP(6))`, "other-"+owner, strings.Repeat("a", 64), p.ID); e == nil {
		t.Fatal("receipt owner/project FK missing")
	}
	if _, e = constraintTx.Exec(`INSERT INTO novel_submission_receipts (owner_username,kind,idempotency_key,fingerprint,project_id,created_at) VALUES (?,'chapter_import','half-result',?,?,UTC_TIMESTAMP(6))`, owner, strings.Repeat("a", 64), p.ID); e == nil {
		t.Fatal("receipt completion check missing")
	}
	if _, e = constraintTx.Exec(`INSERT INTO novel_submission_receipts (owner_username,kind,idempotency_key,fingerprint,created_at) VALUES (?,'other_kind','actual',?,UTC_TIMESTAMP(6))`, owner, strings.Repeat("a", 64)); e != nil {
		t.Fatalf("kind not independent: %v", e)
	}
	if _, e = constraintTx.Exec(`INSERT INTO novel_submission_receipts (owner_username,kind,idempotency_key,fingerprint,created_at) VALUES (?,'chapter_import','actual',?,UTC_TIMESTAMP(6))`, "other-"+owner, strings.Repeat("a", 64)); e != nil {
		t.Fatalf("owner not independent: %v", e)
	}
	if e = constraintTx.Rollback(); e != nil {
		t.Fatal(e)
	}
	in.Chapters[0].Content = "冲突"
	if _, e = reopened.ImportChapters(ctx, owner, in); !IsCode(e, ErrIdempotencyConflict) {
		t.Fatalf("conflict=%v", e)
	}
	// Force a real SQL failure after chapter 1 and before chapter 2 commits.
	if _, e = db.Exec(`ALTER TABLE novel_chapters ADD CONSTRAINT c1a_import_failure CHECK (title <> 'C1a forced SQL failure')`); e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { db.Exec(`ALTER TABLE novel_chapters DROP CHECK c1a_import_failure`) })
	failed := chapterImportInput("sql-rollback")
	failed.Chapters[1].Title = "C1a forced SQL failure"
	var projectCount int
	db.QueryRow(`SELECT COUNT(*) FROM novel_projects WHERE owner_username=?`, owner).Scan(&projectCount)
	if _, e = reopened.ImportChapters(ctx, owner, failed); !IsCode(e, ErrPersistenceFailed) {
		t.Fatalf("forced failure=%v", e)
	}
	var count int
	db.QueryRow(`SELECT COUNT(*) FROM novel_projects WHERE owner_username=?`, owner).Scan(&count)
	if count != projectCount {
		t.Fatal("empty book committed")
	}
	db.QueryRow(`SELECT COUNT(*) FROM novel_submission_receipts WHERE owner_username=? AND idempotency_key='sql-rollback'`, owner).Scan(&count)
	if count != 0 {
		t.Fatal("failed receipt committed")
	}
	// A new model-free book remains normal stored content, with no run/dispatch.
	fresh, e := reopened.ImportChapters(ctx, owner, chapterImportInput("fresh"))
	if e != nil || fresh.Book.ModelID != "" {
		t.Fatalf("fresh=%+v %v", fresh, e)
	}
	db.QueryRow(`SELECT COUNT(*) FROM novel_runs WHERE owner_username=? AND project_id=?`, owner, fresh.Book.ID).Scan(&count)
	if count != 0 {
		t.Fatal("import created run")
	}
	db.QueryRow(`SELECT COUNT(*) FROM novel_run_dispatch WHERE owner_username=? AND project_id=?`, owner, fresh.Book.ID).Scan(&count)
	if count != 0 {
		t.Fatal("import dispatched work")
	}
	// Race the first submission, rather than only racing cached replays.
	var winner ImportChaptersResult
	var winnerMu sync.Mutex
	errs = make(chan error, 8)
	firstInput := chapterImportInput("race-first")
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, e := reopened.ImportChapters(ctx, owner, firstInput)
			if e == nil {
				winnerMu.Lock()
				if winner.Book.ID == "" {
					winner = v
				} else if fingerprint(winner) != fingerprint(v) {
					e = fmt.Errorf("first submissions returned different books")
				}
				winnerMu.Unlock()
			}
			errs <- e
		}()
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		if e != nil {
			t.Fatal(e)
		}
	}
	if e = db.QueryRow(`SELECT COUNT(*) FROM novel_chapters WHERE owner_username=? AND project_id=?`, owner, winner.Book.ID).Scan(&count); e != nil || count != 2 {
		t.Fatalf("first-race chapters=%d %v", count, e)
	}
	if _, e = db.Exec(`UPDATE novel_projects SET status='archived' WHERE owner_username=? AND project_id=?`, owner, winner.Book.ID); e != nil {
		t.Fatal(e)
	}
	if replay, e := reopened.ImportChapters(ctx, owner, firstInput); e != nil || fingerprint(replay) != fingerprint(winner) {
		t.Fatalf("archived receipt replay=%+v %v", replay, e)
	}
	firstInput.BookID = winner.Book.ID
	firstInput.IdempotencyKey = "archived-new"
	if _, e = reopened.ImportChapters(ctx, owner, firstInput); e == nil {
		t.Fatal("archived book accepted new chapters")
	}
	// Exercise the existing whole-project lifecycle with the new receipt FK.
	ret := projectretention.NewService(db, nil, nil, nil, nil)
	retIn := chapterImportInput("retention")
	retBook, e := reopened.ImportChapters(ctx, owner, retIn)
	if e != nil {
		t.Fatal(e)
	}
	if e = ret.ArchiveNovel(ctx, owner, retBook.Book.ID); e != nil {
		t.Fatal(e)
	}
	if cached, e := reopened.ImportChapters(ctx, owner, retIn); e != nil || fingerprint(cached) != fingerprint(retBook) {
		t.Fatalf("retention archived replay=%+v %v", cached, e)
	}
	if e = ret.Restore(ctx, owner, projectretention.ModuleNovelCreation, retBook.Book.ID); e != nil {
		t.Fatal(e)
	}
	if cached, e := reopened.ImportChapters(ctx, owner, retIn); e != nil || fingerprint(cached) != fingerprint(retBook) {
		t.Fatalf("restored receipt replay=%+v %v", cached, e)
	}
	if e = db.QueryRow(`SELECT COUNT(*) FROM novel_chapters WHERE owner_username=? AND project_id=?`, owner, retBook.Book.ID).Scan(&count); e != nil || count != 2 {
		t.Fatal("restore reran import")
	}
	if e = ret.ArchiveNovel(ctx, owner, retBook.Book.ID); e != nil {
		t.Fatal(e)
	}
	// A task-owned FK blocks a later child delete after receipts are deleted.
	// The rollback must restore both receipts and content for a safe retry.
	if _, e = db.Exec(`CREATE TABLE c1a_purge_guard (owner_username VARCHAR(191) NOT NULL,chapter_id VARCHAR(96) NOT NULL,PRIMARY KEY(owner_username,chapter_id),CONSTRAINT fk_c1a_purge_guard FOREIGN KEY(owner_username,chapter_id) REFERENCES novel_chapters(owner_username,chapter_id)) ENGINE=InnoDB`); e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { db.Exec(`DROP TABLE IF EXISTS c1a_purge_guard`) })
	if _, e = db.Exec(`INSERT INTO c1a_purge_guard VALUES (?,?)`, owner, retBook.Chapters[0].ID); e != nil {
		t.Fatal(e)
	}
	if e = ret.Purge(ctx, owner, projectretention.ModuleNovelCreation, retBook.Book.ID); e == nil {
		t.Fatal("forced purge failure not returned")
	}
	if e = db.QueryRow(`SELECT COUNT(*) FROM novel_submission_receipts WHERE owner_username=? AND project_id=?`, owner, retBook.Book.ID).Scan(&count); e != nil || count != 1 {
		t.Fatalf("failed purge lost receipt=%d %v", count, e)
	}
	if e = db.QueryRow(`SELECT COUNT(*) FROM novel_chapters WHERE owner_username=? AND project_id=?`, owner, retBook.Book.ID).Scan(&count); e != nil || count != 2 {
		t.Fatalf("failed purge lost content=%d %v", count, e)
	}
	if _, e = db.Exec(`DROP TABLE c1a_purge_guard`); e != nil {
		t.Fatal(e)
	}
	if e = ret.Purge(ctx, owner, projectretention.ModuleNovelCreation, retBook.Book.ID); e != nil {
		t.Fatalf("retry whole-project purge=%v", e)
	}
	for _, table := range []string{"novel_projects", "novel_chapters", "novel_chapter_versions", "novel_submission_receipts"} {
		if e = db.QueryRow(`SELECT COUNT(*) FROM `+table+` WHERE owner_username=? AND project_id=?`, owner, retBook.Book.ID).Scan(&count); e != nil || count != 0 {
			t.Fatalf("purge left %s=%d %v", table, count, e)
		}
	}
	if e = db.QueryRow(`SELECT COUNT(*) FROM novel_chapters WHERE owner_username=? AND project_id=?`, owner, p.ID).Scan(&count); e != nil || count != 11 || readLegacy() != before {
		t.Fatal("purge affected another project")
	}
	t.Logf("persisted book=%s chapters=%d firstNumber=%d replay-workers=8 real-SQL-rollback=verified", got.Book.ID, len(got.Chapters), got.Chapters[0].Number)
}
