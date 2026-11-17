package novelcreation

// Opt-in acceptance tests. Never point these at an application database.
// The runner must create a fresh, task-owned schema/user and supply the nonce.
import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/projectretention"
	novelprompt "qiantie/backend/internal/prompt/novelcreation"
	"qiantie/backend/internal/storage"
)

var parityInit sync.Once
var parityInitErr error

func parityDB(t *testing.T) *sql.DB {
	t.Helper()
	raw := os.Getenv("QIANTIE_TEST_NOVEL_MYSQL_DSN")
	if raw == "" {
		if os.Getenv("QIANTIE_TEST_NOVEL_REQUIRE") == "1" {
			t.Fatal("owned MySQL DSN required")
		}
		t.Skip("opt-in owned MySQL acceptance")
	}
	c, e := mysql.ParseDSN(raw)
	n := os.Getenv("QIANTIE_TEST_NOVEL_SCHEMA")
	addr := os.Getenv("QIANTIE_TEST_NOVEL_MYSQL_ADDR")
	if addr == "" {
		addr = "127.0.0.1:3306"
	}
	if e != nil || !regexp.MustCompile(`^127\.0\.0\.1:[0-9]{4,5}$`).MatchString(addr) || !regexp.MustCompile(`^novelparity20261117[a-f0-9]{12}$`).MatchString(n) || c.DBName != n || c.Net != "tcp" || c.Addr != addr || c.User != "np"+strings.TrimPrefix(n, "novelparity20261117") || len(os.Getenv("QIANTIE_TEST_NOVEL_NONCE")) < 32 {
		t.Fatal("refusing non-owned MySQL target")
	}
	db, e := sql.Open("mysql", c.FormatDSN())
	if e != nil {
		t.Fatal("open owned MySQL failed")
	}
	db.SetMaxOpenConns(5)
	t.Cleanup(func() { db.Close() })
	var actual, version string
	if e = db.QueryRow("SELECT DATABASE(),VERSION()").Scan(&actual, &version); e != nil || actual != n || !strings.HasPrefix(version, "8.4.") {
		t.Fatal("wrong owned MySQL database/version")
	}
	parityInit.Do(func() {
		ctx := context.Background()
		_, parityInitErr = db.Exec(`CREATE TABLE IF NOT EXISTS novel_parity_fixture_owner (task_id VARCHAR(80) PRIMARY KEY, nonce VARCHAR(100) NOT NULL, source_sha VARCHAR(64) NOT NULL, created_at DATETIME(6) NOT NULL)`)
		if parityInitErr != nil {
			return
		}
		_, parityInitErr = db.Exec(`INSERT IGNORE INTO novel_parity_fixture_owner VALUES ('novel-parity-task8',?,?,UTC_TIMESTAMP(6))`, os.Getenv("QIANTIE_TEST_NOVEL_NONCE"), os.Getenv("QIANTIE_TEST_NOVEL_SOURCE_SHA"))
		if parityInitErr != nil {
			return
		}
		var nonce string
		parityInitErr = db.QueryRow(`SELECT nonce FROM novel_parity_fixture_owner WHERE task_id='novel-parity-task8'`).Scan(&nonce)
		if parityInitErr != nil {
			return
		}
		if nonce != os.Getenv("QIANTIE_TEST_NOVEL_NONCE") {
			parityInitErr = fmt.Errorf("owner nonce mismatch")
			return
		}
		ms := append(storage.AppMigrations(), storage.NovelFetchWorkshopMigrations()...)
		ms = append(ms, storage.AgentProxyMigrations()...)
		ms = append(ms, storage.AuthMigrations()...)
		ms = append(ms, storage.ScreenplayMigrations()...)
		ms = append(ms, storage.NovelCreationMigrations()...)
		for pass := 1; pass <= 2; pass++ {
			if parityInitErr = storage.RunMigrations(ctx, db, ms); parityInitErr != nil {
				return
			}
			t.Logf("owned-schema=%s MySQL=%s migration-pass=%d", n, version, pass)
		}
		for _, m := range ms {
			var checksum string
			if parityInitErr = db.QueryRow(`SELECT checksum FROM schema_migrations WHERE version=?`, m.Version).Scan(&checksum); parityInitErr != nil {
				return
			}
			if checksum != storage.ChecksumFor(m) {
				parityInitErr = fmt.Errorf("migration checksum %d", m.Version)
				return
			}
		}
		uniqueVersions := map[int]string{}
		for _, m := range ms {
			sum := storage.ChecksumFor(m)
			if prior, ok := uniqueVersions[m.Version]; ok && prior != sum {
				parityInitErr = fmt.Errorf("duplicate migration version with different checksum %d", m.Version)
				return
			}
			uniqueVersions[m.Version] = sum
		}
		var ledgerCount int
		if parityInitErr = db.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&ledgerCount); parityInitErr != nil {
			return
		}
		if ledgerCount != len(uniqueVersions) {
			parityInitErr = fmt.Errorf("migration ledger count=%d expected=%d", ledgerCount, len(uniqueVersions))
			return
		}
		t.Logf("all %d composed entries / %d unique migration checksums and ledger rows verified", len(ms), ledgerCount)
		for _, v := range []int{8800311, 8800312, 8800313, 8800314, 8800401} {
			var count int
			if parityInitErr = db.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE version=?`, v).Scan(&count); parityInitErr != nil {
				return
			}
			if count != 1 {
				parityInitErr = fmt.Errorf("missing migration %d", v)
				return
			}
		}
	})
	if parityInitErr != nil {
		t.Fatal(parityInitErr)
	}
	return db
}

type parityHTTP struct {
	mu       sync.Mutex
	payloads []map[string]any
	stages   []string
	block    chan struct{}
	entered  chan struct{}
	once     sync.Once
	reject   bool
	server   *httptest.Server
	client   *http.Client
}

func newParityHTTP(t *testing.T) *parityHTTP {
	t.Helper()
	h := &parityHTTP{}
	h.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			Messages []struct{ Role, Content string }
		}
		if e := json.NewDecoder(r.Body).Decode(&in); e != nil {
			http.Error(w, "bad JSON", 400)
			return
		}
		var system, user string
		for _, m := range in.Messages {
			if m.Role == "system" {
				system = m.Content
			}
			if m.Role == "user" {
				user = m.Content
			}
		}
		var payload map[string]any
		_ = json.Unmarshal([]byte(user), &payload)
		stage, response := "", ""
		switch {
		case strings.Contains(system, novelprompt.FoundationPlan.OutputContract):
			stage, response = "foundation", foundationJSON
		case strings.Contains(system, novelprompt.ChapterPlan.OutputContract):
			stage = "plan"
			n := int(payload["chapter_number"].(float64))
			response = strings.Replace(chapterPlanJSON, `"chapterNumber":1`, fmt.Sprintf(`"chapterNumber":%d`, n), 1)
		case strings.Contains(system, novelprompt.ChapterAudit.OutputContract):
			stage, response = "audit", passedAuditJSON
		case strings.Contains(system, novelprompt.TruthSettlement.OutputContract):
			stage, response = "truth", settlementJSON
		case strings.Contains(system, novelprompt.MemoryIndex.OutputContract):
			stage, response = "memory", memoryJSON
		default:
			stage, response = "draft", draftJSON
		}
		h.mu.Lock()
		h.payloads = append(h.payloads, payload)
		h.stages = append(h.stages, stage)
		block, entered, reject := h.block, h.entered, h.reject
		h.mu.Unlock()
		if entered != nil {
			h.once.Do(func() { close(entered) })
		}
		if block != nil {
			select {
			case <-block:
			case <-r.Context().Done():
				return
			}
		}
		if reject {
			http.Error(w, "rate limited", 429)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"message": map[string]any{"content": response}}}})
	}))
	addr := h.server.Listener.Addr().String()
	transport := &http.Transport{Proxy: nil, DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
		if address != addr {
			return nil, fmt.Errorf("non-loopback fixture address refused")
		}
		return (&net.Dialer{}).DialContext(ctx, network, address)
	}}
	h.client = &http.Client{Transport: transport, Timeout: 10 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	t.Cleanup(func() { h.server.Close(); transport.CloseIdleConnections() })
	return h
}
func (h *parityHTTP) count() int { h.mu.Lock(); defer h.mu.Unlock(); return len(h.stages) }
func (h *parityHTTP) pipeline() *Pipeline {
	return NewPipeline(func(_ context.Context, _, m string) (batchfactoryv11.DirectorProvider, error) {
		if m != "np-fake" {
			return nil, fmt.Errorf("fixture model refused")
		}
		return &batchfactoryv11.OpenAICompatibleProvider{Endpoint: h.server.URL, APIKey: "local-acceptance-only", Model: m, Client: h.client}, nil
	})
}

type parityFixture struct {
	db      *sql.DB
	store   Store
	service *Service
	owner   string
	project Project
	session Session
	http    *parityHTTP
}

func parityFixtureFor(t *testing.T, mode string) parityFixture {
	t.Helper()
	db := parityDB(t)
	h := newParityHTTP(t)
	st := NewMySQLStore(db)
	s := NewService(st, NewMemoryRuntime(), h.pipeline())
	retention := projectretention.NewService(db, nil, nil, nil, nil)
	s.ConfigureRetention(retention.ArchiveNovel, func(ctx context.Context, o, p string) error {
		return retention.Restore(ctx, o, projectretention.ModuleNovelCreation, p)
	})
	o := novelID("acceptance")
	p, e := s.CreateProject(context.Background(), o, CreateProjectInput{Title: "验收雨幕记者", Genre: "都市悬疑", World: "旧城暴雨", Protagonist: "林澜", CoreConflict: "调查匿名信", ModelID: "np-fake", TargetChapterWords: 2000})
	if e != nil {
		t.Fatal(e)
	}
	var a AgentSettings
	_ = json.Unmarshal([]byte(`{"maxRevisionRounds":0,"enableLengthNorm":false,"enablePostWrite":false,"enableAiTells":false,"requireSecondaryConfirm":false}`), &a)
	if _, e = s.UpdateProjectSettings(context.Background(), o, p.ID, ProjectSettings{Agent: a}); e != nil {
		t.Fatal(e)
	}
	se, e := s.CreateSession(context.Background(), o, CreateSessionInput{ProjectID: p.ID, Mode: mode, Title: "验收会话"})
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("owner=%s project=%s session=%s mode=%s", o, p.ID, se.ID, mode)
	return parityFixture{db, st, s, o, p, se, h}
}
func (f parityFixture) input(key string) SendMessageInput {
	return SendMessageInput{ProjectID: f.project.ID, SessionID: f.session.ID, Content: "写一章完整故事，以匿名信开场。", ModelID: "np-fake", IdempotencyKey: key}
}
func (f parityFixture) other(t *testing.T) *Service {
	t.Helper()
	return NewService(NewMySQLStore(parityDB(t)), NewMemoryRuntime(), f.http.pipeline())
}
func parityRun(t *testing.T, f parityFixture, in SendMessageInput) Run {
	t.Helper()
	r, e := f.service.SendMessage(context.Background(), f.owner, in)
	if e != nil {
		t.Fatal(e)
	}
	r, e = f.service.Run(context.Background(), f.owner, r.ID)
	if e != nil {
		t.Fatal(e)
	}
	if r.Status == RunAwaitingConfirmation {
		r, e = f.service.ConfirmRun(context.Background(), f.owner, r.ID, json.RawMessage(foundationJSON))
		if e != nil {
			t.Fatal(e)
		}
		r, e = f.service.Run(context.Background(), f.owner, r.ID)
		if e != nil {
			t.Fatal(e)
		}
	}
	if r.Status != RunSucceeded {
		t.Fatalf("run=%s status=%s", r.ID, r.Status)
	}
	t.Logf("run=%s status=%s HTTP=%d", r.ID, r.Status, f.http.count())
	return r
}
func parityCount(t *testing.T, f parityFixture, table, where string, args ...any) int {
	t.Helper()
	allowed := map[string]bool{"novel_runs": true, "novel_messages": true, "novel_run_dispatch": true, "novel_run_events": true, "novel_chapter_versions": true, "novel_automation_rounds": true, "novel_run_steps": true}
	if !allowed[table] {
		t.Fatal("invalid count table")
	}
	var n int
	if e := f.db.QueryRow("SELECT COUNT(*) FROM "+table+" WHERE "+where, args...).Scan(&n); e != nil {
		t.Fatal(e)
	}
	return n
}

func TestNovelParityRealMySQLModes(t *testing.T) {
	for _, mode := range []string{"long_original", "short_original", "long_imitation", "short_imitation", "fanfiction", "continuation"} {
		t.Run(mode, func(t *testing.T) {
			f := parityFixtureFor(t, mode)
			in := f.input("valid")
			if mode == "long_imitation" || mode == "short_imitation" || mode == "fanfiction" || mode == "continuation" {
				if _, e := f.service.SendMessage(context.Background(), f.owner, in); e == nil {
					t.Fatal("missing mode input accepted")
				}
			}
			bad := in
			bad.Content = ""
			if _, e := f.service.SendMessage(context.Background(), f.owner, bad); e == nil {
				t.Fatal("empty input accepted")
			}
			switch mode {
			case "long_imitation", "short_imitation":
				p, e := f.service.SaveStyleProfile(context.Background(), f.owner, StyleProfile{Title: "验收文风", Content: "短句，克制的环境描写"})
				if e != nil {
					t.Fatal(e)
				}
				in.StyleProfileID = p.ID
			case "fanfiction":
				in.ModeInputs = map[string]string{"existingSetting": "林澜是调查记者，旧城连续暴雨"}
			case "continuation":
				in.ModeInputs = map[string]string{"previousText": "林澜收到匿名信，决定前往旧报社。"}
			}
			foreign, e := f.service.SaveStyleProfile(context.Background(), "foreign", StyleProfile{Title: "外账号", Content: "不可用"})
			if e != nil {
				t.Fatal(e)
			}
			bad = in
			bad.StyleProfileID = foreign.ID
			if _, e = f.service.SendMessage(context.Background(), f.owner, bad); e == nil {
				t.Fatal("foreign style accepted")
			}
			bad = in
			bad.NoteIDs = []string{"missing-note"}
			if _, e = f.service.SendMessage(context.Background(), f.owner, bad); e == nil {
				t.Fatal("missing reference accepted")
			}
			if f.http.count() != 0 {
				t.Fatal("invalid input called HTTP")
			}
			r := parityRun(t, f, in)
			fresh := f.other(t)
			cs, e := fresh.store.ListChapters(context.Background(), f.owner, f.project.ID)
			if e != nil || len(cs) != 1 || cs[0].Content == "" {
				t.Fatalf("not persisted: %+v %v", cs, e)
			}
			if parityCount(t, f, "novel_chapter_versions", "run_id=?", r.ID) != 1 {
				t.Fatal("version count")
			}
		})
	}
}

func TestNovelParityRealMySQLSchema(t *testing.T) {
	db := parityDB(t)
	for _, c := range []struct{ table, name, kind string }{
		{"novel_runs", "uq_novel_runs_idempotency", "UNIQUE"},
		{"novel_automation_rounds", "PRIMARY", "PRIMARY KEY"},
		{"novel_automation_rounds", "uq_novel_automation_run", "UNIQUE"},
		{"novel_automation_rounds", "fk_novel_automation_run", "FOREIGN KEY"},
		{"novel_run_dispatch", "fk_novel_dispatch_run", "FOREIGN KEY"},
		{"novel_project_slots", "fk_novel_slot_project", "FOREIGN KEY"},
		{"novel_chapter_versions", "uq_novel_chapter_versions_number", "UNIQUE"},
		{"novel_chapter_versions", "fk_novel_chapter_versions_chapter", "FOREIGN KEY"},
	} {
		var n int
		e := db.QueryRow(`SELECT COUNT(*) FROM information_schema.TABLE_CONSTRAINTS WHERE CONSTRAINT_SCHEMA=DATABASE() AND TABLE_NAME=? AND CONSTRAINT_NAME=? AND CONSTRAINT_TYPE=?`, c.table, c.name, c.kind).Scan(&n)
		if e != nil || n != 1 {
			t.Fatalf("constraint %s.%s count=%d %v", c.table, c.name, n, e)
		}
	}
	var cols string
	if e := db.QueryRow(`SELECT GROUP_CONCAT(COLUMN_NAME ORDER BY ORDINAL_POSITION) FROM information_schema.KEY_COLUMN_USAGE WHERE CONSTRAINT_SCHEMA=DATABASE() AND TABLE_NAME='novel_automation_rounds' AND CONSTRAINT_NAME='PRIMARY'`).Scan(&cols); e != nil || cols != "owner_username,kind,plan_id,round_key" {
		t.Fatalf("round primary=%s %v", cols, e)
	}
	t.Log("actual information_schema verifies run/round/version unique keys and dispatch/slot/round/version foreign keys")
}

func TestNovelParityRealMySQLContent(t *testing.T) {
	f := parityFixtureFor(t, "short_original")
	ctx := context.Background()
	j, e := f.service.PasteDocument(ctx, f.owner, f.project.ID, "验收粘贴", "第一章 初见\n林澜收信。\n第二章 重逢\n林澜赴约。")
	if e != nil {
		t.Fatal(e)
	}
	j, e = f.other(t).GetDocument(ctx, f.owner, f.project.ID, j.ID)
	if e != nil || j.Status != "ready" || len(j.Chapters) != 2 {
		t.Fatalf("parse=%+v %v", j, e)
	}
	selected, e := f.service.SelectDocument(ctx, f.owner, f.project.ID, j.ID, DocumentSelection{ExpectedVersion: j.Version, ChapterIDs: []string{j.Chapters[1].ID, j.Chapters[0].ID}})
	if e != nil || !strings.HasPrefix(selected.Content, "第二章") {
		t.Fatalf("selection=%+v %v", selected, e)
	}
	source, e := f.service.CreateStyleFromDocument(ctx, f.owner, f.project.ID, j.ID, "验收纯统计", "短句")
	if e != nil {
		t.Fatal(e)
	}
	note, e := f.store.CreateNote(ctx, f.owner, f.project.ID, ProjectNoteInput{Title: "验收备忘录", Content: "信封无寄件人"})
	if e != nil {
		t.Fatal(e)
	}
	if f.http.count() != 0 {
		t.Fatal("paste/statistics called provider")
	}
	in := f.input("content")
	in.KnowledgeIDs = []string{"docref-" + j.ID}
	in.NoteIDs = []string{note.ID}
	r := parityRun(t, f, in)
	savedSnapshot, e := f.other(t).GetSnapshot(ctx, f.owner, r.ID, 0)
	if e != nil || !savedSnapshot.CanExportShort {
		t.Fatalf("saved short capability=%t %v", savedSnapshot.CanExportShort, e)
	}
	raw, _ := json.Marshal(f.http.payloads)
	if !strings.Contains(string(raw), "信封无寄件人") || !strings.Contains(string(raw), "林澜赴约") {
		t.Fatal("materials not injected into actual HTTP")
	}
	cs, e := f.store.ListChapters(ctx, f.owner, f.project.ID)
	if e != nil || len(cs) != 1 {
		t.Fatal(e)
	}
	ch := cs[0]
	edited, e := f.service.EditChapter(ctx, f.owner, f.project.ID, ch.ID, ChapterEdit{ExpectedVersion: ch.Version, Title: ch.Title, Content: "验收手动编辑后的正文。"})
	if e != nil || edited.Version != 2 {
		t.Fatalf("edit=%+v %v", edited, e)
	}
	savedSnapshot, e = f.other(t).GetSnapshot(ctx, f.owner, r.ID, 0)
	if e != nil || !savedSnapshot.CanExportShort {
		t.Fatalf("edited short lost historical export fact %v", e)
	}
	originalExport, e := f.other(t).ExportShortResult(ctx, f.owner, r.ID, "txt")
	if e != nil || strings.TrimSpace(originalExport.Content) != ch.Content {
		t.Fatal("short export did not retain original saved run version")
	}
	if _, e = f.service.EditChapter(ctx, f.owner, f.project.ID, ch.ID, ChapterEdit{ExpectedVersion: 1, Content: "旧窗口"}); e == nil {
		t.Fatal("stale edit accepted")
	}
	restored, e := f.service.RestoreChapter(ctx, f.owner, f.project.ID, ch.ID, 2, 1)
	if e != nil || restored.Version != 3 || restored.Content != ch.Content {
		t.Fatalf("restore=%+v %v", restored, e)
	}
	versions, e := f.other(t).ListChapterVersions(ctx, f.owner, f.project.ID, ch.ID)
	if e != nil || len(versions) != 3 {
		t.Fatal("version history missing")
	}
	savedSnapshot, e = f.other(t).GetSessionSnapshot(ctx, f.owner, f.session.ID)
	if e != nil || !savedSnapshot.CanExportShort {
		t.Fatalf("historical short version capability=%t %v", savedSnapshot.CanExportShort, e)
	}
	for _, format := range []string{"txt", "markdown"} {
		doc, e := f.service.ExportShortResult(ctx, f.owner, r.ID, format)
		if e != nil || len(doc.Content) == 0 {
			t.Fatalf("short export %s %v", format, e)
		}
	}
	before := f.http.count()
	if e = f.service.ArchiveProject(ctx, f.owner, f.project.ID); e != nil {
		t.Fatal(e)
	}
	if e = f.service.RestoreProject(ctx, f.owner, f.project.ID); e != nil {
		t.Fatal(e)
	}
	if f.http.count() != before {
		t.Fatal("archive/restore generated")
	}
	t.Logf("document=%s style=%s run=%s chapter=%s versions=3 ordered selection + pure statistics + memo + TXT/Markdown + archive/restore; no real TOS", j.ID, source.Profile.ID, r.ID, ch.ID)
}

func TestNovelParityRealMySQLAtomicAndClaim(t *testing.T) {
	f := parityFixtureFor(t, "short_original")
	other := f.other(t)
	in := f.input("same-key")
	var wg sync.WaitGroup
	ids := make(chan string, 12)
	start := make(chan struct{})
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			s := f.service
			if i%2 == 1 {
				s = other
			}
			r, e := s.SendMessage(context.Background(), f.owner, in)
			if e != nil {
				t.Error(e)
			}
			ids <- r.ID
		}(i)
	}
	close(start)
	wg.Wait()
	close(ids)
	id := ""
	for got := range ids {
		if id == "" {
			id = got
		}
		if got == "" || got != id {
			t.Fatal("duplicate run")
		}
	}
	for _, check := range []struct {
		table, where string
		args         []any
	}{{"novel_runs", "project_id=?", []any{f.project.ID}}, {"novel_messages", "session_id=? AND role='user'", []any{f.session.ID}}, {"novel_run_dispatch", "run_id=?", []any{id}}} {
		if n := parityCount(t, f, check.table, check.where, check.args...); n != 1 {
			t.Fatalf("%s=%d", check.table, n)
		}
	}
	in.Content = "不同正文"
	if _, e := other.SendMessage(context.Background(), f.owner, in); !IsCode(e, ErrIdempotencyConflict) {
		t.Fatalf("different input=%v", e)
	}
	f.http.block = make(chan struct{})
	t.Cleanup(func() {
		select {
		case <-f.http.block:
		default:
			close(f.http.block)
		}
	})
	f.http.entered = make(chan struct{})
	done := make(chan error, 1)
	go func() { _, e := f.service.Run(context.Background(), f.owner, id); done <- e }()
	<-f.http.entered
	if _, e := other.Run(context.Background(), f.owner, id); !IsCode(e, ErrProjectLocked) {
		t.Fatalf("second executor=%v", e)
	}
	in = f.input("different-key")
	second, err := other.SendMessage(context.Background(), f.owner, in)
	if err != nil || second.ID == id || second.ID == "" {
		t.Fatalf("distinct send=%s %v", second.ID, err)
	}
	if _, err = other.Run(context.Background(), f.owner, second.ID); !IsCode(err, ErrProjectLocked) {
		t.Fatalf("distinct send overlapped=%v", err)
	}
	close(f.http.block)
	if e := <-done; e != nil {
		t.Fatal(e)
	}
	if f.http.count() != 6 {
		t.Fatalf("HTTP=%d want6", f.http.count())
	}
	if _, err = other.Run(context.Background(), f.owner, second.ID); err == nil {
		t.Fatal("stale second input overwrote first chapter")
	}
	if f.http.count() != 6 || parityCount(t, f, "novel_chapter_versions", "project_id=?", f.project.ID) != 1 {
		t.Fatal("stale second send called provider or overwrote result")
	}
	t.Logf("run=%s 12 requests=>1 user/run/dispatch; 2 pools=>6 HTTP", id)
}

func TestNovelParityRealMySQLContinuity(t *testing.T) {
	f := parityFixtureFor(t, "long_original")
	first := parityRun(t, f, f.input("chapter1"))
	second := parityRun(t, f, f.input("chapter2"))
	f.http.mu.Lock()
	var secondPlan map[string]any
	for i, p := range f.http.payloads {
		if f.http.stages[i] == "plan" && p["chapter_number"] == float64(2) {
			secondPlan = p
		}
	}
	f.http.mu.Unlock()
	raw, _ := json.Marshal(secondPlan)
	for _, part := range []string{"previous_chapters", "version", "林澜拿到带有失踪者指纹的信", "林澜收到匿名信并找到指纹", "雨水敲着窗"} {
		if !strings.Contains(string(raw), part) {
			t.Fatalf("second HTTP lost %s", part)
		}
	}
	if parityCount(t, f, "novel_chapter_versions", "run_id IN (?,?)", first.ID, second.ID) != 2 {
		t.Fatal("two versions missing")
	}
	before := f.http.count()
	f.service.store = continuityFailStore{f.store}
	if _, e := f.service.SendMessage(context.Background(), f.owner, f.input("read-fail")); e == nil {
		t.Fatal("read error ignored")
	}
	if f.http.count() != before {
		t.Fatal("read failure called provider")
	}
	t.Logf("first=%s second=%s chapter2 real HTTP carries persisted version/truth/memory; failed read HTTP delta=0", first.ID, second.ID)
}

func TestNovelParityRealMySQLRecovery(t *testing.T) {
	for _, kind := range []string{"ready", "event", "completed", "save_pending"} {
		t.Run(kind, func(t *testing.T) {
			f := parityFixtureFor(t, "short_original")
			r, e := f.service.SendMessage(context.Background(), f.owner, f.input(kind))
			if e != nil {
				t.Fatal(e)
			}
			switch kind {
			case "event", "completed":
				f.service.store = faultJournalStore{f.store, &onceWriteFault{kind: kind}}
			case "save_pending":
				f.service.store = flakySaveStore{f.store, &saveFailure{left: 1}}
			}
			if kind != "ready" {
				if _, e = f.service.Run(context.Background(), f.owner, r.ID); e == nil {
					t.Fatal("fault did not fire")
				}
			}
			before := f.http.count()
			ex, _, e := f.store.GetExecution(context.Background(), f.owner, r.ID)
			if e != nil {
				t.Fatal(e)
			}
			other := f.other(t)
			if kind == "completed" {
				if ex.State != "paused" {
					t.Fatalf("unknown=%s", ex.State)
				}
				if _, e = other.Run(context.Background(), f.owner, r.ID); e == nil {
					t.Fatal("unknown result was executable")
				}
				persisted, _ := other.store.GetRun(context.Background(), f.owner, r.ID)
				if persisted.Status != RunPaused || persisted.ErrorCode != ErrResultUnknown {
					t.Fatal("unknown result lost persisted reason")
				}
				if _, e = other.RetryRun(context.Background(), f.owner, r.ID); !IsCode(e, ErrResultUnknown) {
					t.Fatal(e)
				}
				if f.http.count() != before {
					t.Fatal("unknown resent")
				}
				return
			}
			if kind == "event" && ex.State != "ready" {
				t.Fatalf("completed raw state=%s", ex.State)
			}
			if kind == "save_pending" && (ex.State != "save_pending" || ex.Bundle == nil) {
				t.Fatal("bundle missing")
			}
			if kind == "ready" {
				ctx, cancel := context.WithCancel(context.Background())
				StartWorker(ctx, other, 20*time.Millisecond)
				deadline := time.Now().Add(8 * time.Second)
				for time.Now().Before(deadline) {
					got, _ := f.store.GetRun(context.Background(), f.owner, r.ID)
					if got.Status == RunSucceeded {
						break
					}
					time.Sleep(20 * time.Millisecond)
				}
				cancel()
			} else {
				if _, e = other.Run(context.Background(), f.owner, r.ID); e != nil {
					t.Fatal(e)
				}
			}
			got, e := other.store.GetRun(context.Background(), f.owner, r.ID)
			if e != nil || got.Status != RunSucceeded {
				t.Fatalf("restart=%s %v", got.Status, e)
			}
			if f.http.count() != 6 {
				t.Fatalf("HTTP=%d want6", f.http.count())
			}
			if kind == "save_pending" && f.http.count() != before {
				t.Fatal("save regenerated")
			}
			if _, e = other.Run(context.Background(), f.owner, r.ID); e != nil {
				t.Fatal(e)
			}
			if parityCount(t, f, "novel_chapter_versions", "run_id=?", r.ID) != 1 {
				t.Fatal("duplicate version")
			}
			t.Logf("run=%s kind=%s before=%d after=%d", r.ID, kind, before, f.http.count())
		})
	}
}

func TestNovelParityRealMySQLFenceAndCancel(t *testing.T) {
	t.Run("fence", func(t *testing.T) {
		f := parityFixtureFor(t, "short_original")
		r, e := f.service.SendMessage(context.Background(), f.owner, f.input("fence"))
		if e != nil {
			t.Fatal(e)
		}
		l, _, e := f.service.claim(context.Background(), f.owner, r.ID)
		if e != nil {
			t.Fatal(e)
		}
		e = f.store.Atomic(context.Background(), f.owner, f.project.ID, func(tx Store) error {
			now, e := tx.Clock(context.Background())
			if e != nil {
				return e
			}
			ex, _, e := tx.GetExecution(context.Background(), f.owner, r.ID)
			if e != nil {
				return e
			}
			ex.LeaseUntil = now.Add(-time.Second)
			if e = tx.PutExecution(context.Background(), ex); e != nil {
				return e
			}
			slot, e := tx.GetSlot(context.Background(), f.owner, f.project.ID)
			if e != nil {
				return e
			}
			slot.LeaseUntil = ex.LeaseUntil
			return tx.PutSlot(context.Background(), f.owner, f.project.ID, slot)
		})
		if e != nil {
			t.Fatal(e)
		}
		next, _, e := f.other(t).claim(context.Background(), f.owner, r.ID)
		if e != nil || next.Fence <= l.Fence {
			t.Fatalf("fence=%+v %v", next, e)
		}
		old := &guardedStore{Store: f.store, lease: l}
		if _, e = old.AppendRunEvent(context.Background(), f.owner, RunEvent{RunID: r.ID}); !IsCode(e, ErrLeaseLost) {
			t.Fatal(e)
		}
		if e = old.SaveFoundation(context.Background(), f.owner, Foundation{ID: "old", ProjectID: f.project.ID, Snapshot: json.RawMessage(foundationJSON)}); !IsCode(e, ErrLeaseLost) {
			t.Fatal(e)
		}
		if e = old.SaveChapterBundle(context.Background(), f.owner, serviceTestBundle(RunContext{Run: r, Project: f.project}, RunSucceeded)); !IsCode(e, ErrLeaseLost) {
			t.Fatal(e)
		}
		_, _ = f.service.CancelRun(context.Background(), f.owner, r.ID)
		t.Logf("run=%s fence=%d=>%d; old writes rejected", r.ID, l.Fence, next.Fence)
	})
	t.Run("cancel", func(t *testing.T) {
		f := parityFixtureFor(t, "short_original")
		f.http.block = make(chan struct{})
		f.http.entered = make(chan struct{})
		r, e := f.service.SendMessage(context.Background(), f.owner, f.input("cancel"))
		if e != nil {
			t.Fatal(e)
		}
		done := make(chan error, 1)
		go func() { _, e := f.service.Run(context.Background(), f.owner, r.ID); done <- e }()
		<-f.http.entered
		if _, e = f.other(t).CancelRun(context.Background(), f.owner, r.ID); e != nil {
			t.Fatal(e)
		}
		close(f.http.block)
		<-done
		if f.http.count() != 1 {
			t.Fatal("cancel dispatched later stages")
		}
		if parityCount(t, f, "novel_chapter_versions", "run_id=?", r.ID) != 0 {
			t.Fatal("cancel saved body")
		}
		t.Logf("run=%s cancel after1 received HTTP; later HTTP=0 versions=0", r.ID)
	})
}

func TestNovelParityRealMySQLCreationIntents(t *testing.T) {
	for _, kind := range []string{"schedule", "batch"} {
		t.Run(kind, func(t *testing.T) {
			f := parityFixtureFor(t, "short_original")
			other := f.other(t)
			ctx := context.Background()
			note, e := f.store.CreateNote(ctx, f.owner, f.project.ID, ProjectNoteInput{Content: "原始备忘录"})
			if e != nil {
				t.Fatal(e)
			}
			si := ScheduleInput{IdempotencyKey: "same-create", SessionID: f.session.ID, Prompt: "原请求", Enabled: true, NoteIDs: []string{note.ID}}
			bi := BatchInput{IdempotencyKey: "same-create", SessionID: f.session.ID, Prompt: "原请求", NoteIDs: []string{note.ID}}
			create := func(s *Service) (string, error) {
				if kind == "schedule" {
					p, e := s.CreateSchedule(ctx, f.owner, f.project.ID, si)
					return p.ID, e
				}
				p, e := s.CreateBatch(ctx, f.owner, f.project.ID, bi)
				return p.ID, e
			}
			var wg sync.WaitGroup
			ids := make(chan string, 8)
			start := make(chan struct{})
			for i := 0; i < 8; i++ {
				wg.Add(1)
				go func(i int) {
					defer wg.Done()
					<-start
					s := f.service
					if i%2 == 1 {
						s = other
					}
					id, e := create(s)
					if e != nil {
						t.Error(e)
					}
					ids <- id
				}(i)
			}
			close(start)
			wg.Wait()
			close(ids)
			id := ""
			for got := range ids {
				if id == "" {
					id = got
				}
				if got == "" || got != id {
					t.Fatal("creation duplicated")
				}
			}
			if e = f.store.DeleteNote(ctx, f.owner, f.project.ID, note.ID); e != nil {
				t.Fatal(e)
			}
			var fingerprint string
			if kind == "schedule" {
				p, e := f.store.GetSchedule(ctx, f.owner, id)
				if e != nil {
					t.Fatal(e)
				}
				fingerprint = p.Creation.Fingerprint
				p.Prompt = "改后"
				p.Status = "stopped"
				p.Enabled = false
				p.TotalSavedChapters = 7
				if _, e = f.store.UpdateSchedule(ctx, f.owner, p); e != nil {
					t.Fatal(e)
				}
			} else {
				p, e := f.store.GetBatch(ctx, f.owner, id)
				if e != nil {
					t.Fatal(e)
				}
				fingerprint = p.Creation.Fingerprint
				p.Prompt = "改后"
				p.Status = BatchStopped
				p.CompletedChapters = 2
				if _, e = f.store.UpdateBatch(ctx, f.owner, p); e != nil {
					t.Fatal(e)
				}
			}
			if e = f.service.ArchiveProject(ctx, f.owner, f.project.ID); e != nil {
				t.Fatal(e)
			}
			before, e := f.store.GetProject(ctx, f.owner, f.project.ID)
			if e != nil {
				t.Fatal(e)
			}
			for i := 0; i < 8; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					got, e := create(other)
					if e != nil || got != id {
						t.Errorf("archived replay=%s %v", got, e)
					}
				}()
			}
			wg.Wait()
			after, _ := f.store.GetProject(ctx, f.owner, f.project.ID)
			if !reflect.DeepEqual(before, after) {
				t.Fatal("replay touched project")
			}
			if kind == "schedule" {
				ps, e := other.ListSchedules(ctx, f.owner, f.project.ID)
				if e != nil || len(ps) != 1 || ps[0].Prompt != "改后" || ps[0].TotalSavedChapters != 7 || ps[0].Creation.Fingerprint != fingerprint || ps[0].Enabled {
					t.Fatalf("schedule reset %+v %v", ps, e)
				}
				si.Prompt = "冲突"
				if _, e = create(other); !IsCode(e, ErrIdempotencyConflict) {
					t.Fatalf("conflict=%v", e)
				}
				si.IdempotencyKey = "new"
				if _, e = create(other); e == nil {
					t.Fatal("archived new allowed")
				}
			} else {
				ps, e := other.ListBatches(ctx, f.owner, f.project.ID)
				if e != nil || len(ps) != 1 || ps[0].Prompt != "改后" || ps[0].CompletedChapters != 2 || ps[0].Creation.Fingerprint != fingerprint {
					t.Fatalf("batch reset %+v %v", ps, e)
				}
				bi.Prompt = "冲突"
				if _, e = create(other); !IsCode(e, ErrIdempotencyConflict) {
					t.Fatal(e)
				}
				bi.IdempotencyKey = "new"
				if _, e = create(other); e == nil {
					t.Fatal("archived new allowed")
				}
			}
			if f.http.count() != 0 {
				t.Fatal("creation invoked model")
			}
			t.Logf("%s id=%s 8 concurrent initial=>1; edited/stopped/progressed/deleted-ref/archived replay unchanged HTTP=0", kind, id)
		})
	}
}

func TestNovelParityRealMySQLAutomation(t *testing.T) {
	t.Run("schedule", func(t *testing.T) {
		f := parityFixtureFor(t, "short_original")
		ctx := context.Background()
		other := f.other(t)
		p, e := f.service.CreateSchedule(ctx, f.owner, f.project.ID, ScheduleInput{SessionID: f.session.ID, Prompt: "写作", Enabled: true, IntervalMinutes: 30})
		if e != nil {
			t.Fatal(e)
		}
		if f.http.count() != 0 {
			t.Fatal("save called provider")
		}
		due := time.Now().UTC().Add(-time.Minute)
		p.NextRunAt = due
		if _, e = f.store.UpdateSchedule(ctx, f.owner, p); e != nil {
			t.Fatal(e)
		}
		var wg sync.WaitGroup
		for _, s := range []*Service{f.service, other} {
			wg.Add(1)
			go func(s *Service) {
				defer wg.Done()
				_, e := s.DispatchDueSchedules(ctx, time.Now().UTC())
				if e != nil {
					t.Error(e)
				}
			}(s)
		}
		wg.Wait()
		ps, e := f.store.GetSchedule(ctx, f.owner, p.ID)
		if e != nil {
			t.Fatal(e)
		}
		runs, e := f.store.ListRuns(ctx, f.owner, f.project.ID)
		if e != nil || len(runs) != 1 {
			t.Fatalf("runs=%d %v", len(runs), e)
		}
		if _, e = f.service.Run(ctx, f.owner, runs[0].ID); e != nil {
			t.Fatal(e)
		}
		for _, s := range []*Service{other, f.service} {
			if _, e = s.DispatchDueSchedules(ctx, time.Now().UTC()); e != nil {
				t.Fatal(e)
			}
		}
		ps, e = f.store.GetSchedule(ctx, f.owner, p.ID)
		if e != nil || ps.TotalSavedChapters != 1 {
			t.Fatalf("counter=%+v %v", ps, e)
		}
		if parityCount(t, f, "novel_automation_rounds", "plan_id=?", p.ID) != 1 {
			t.Fatal("duplicate round")
		}
		if _, e = f.service.ControlSchedule(ctx, f.owner, f.project.ID, p.ID, "pause"); e != nil {
			t.Fatal(e)
		}
		t.Logf("schedule=%s run=%s saved=1 rounds=1 HTTP=%d", p.ID, runs[0].ID, f.http.count())
	})
	t.Run("batch", func(t *testing.T) {
		f := parityFixtureFor(t, "short_original")
		ctx := context.Background()
		p, e := f.service.CreateBatch(ctx, f.owner, f.project.ID, BatchInput{SessionID: f.session.ID, Prompt: "两章", TotalChapters: 2})
		if e != nil {
			t.Fatal(e)
		}
		other := f.other(t)
		for i := 0; i < 2; i++ {
			var wg sync.WaitGroup
			for _, s := range []*Service{f.service, other} {
				wg.Add(1)
				go func(s *Service) {
					defer wg.Done()
					_, e := s.AdvanceBatches(ctx)
					if e != nil {
						t.Error(e)
					}
				}(s)
			}
			wg.Wait()
			plan, e := f.store.GetBatch(ctx, f.owner, p.ID)
			if e != nil {
				t.Fatal(e)
			}
			if _, e = f.service.Run(ctx, f.owner, plan.CurrentRunID); e != nil {
				t.Fatal(e)
			}
		}
		for i := 0; i < 3; i++ {
			if _, e = other.AdvanceBatches(ctx); e != nil {
				t.Fatal(e)
			}
		}
		plan, e := f.store.GetBatch(ctx, f.owner, p.ID)
		if e != nil || plan.Status != BatchCompleted || plan.CompletedChapters != 2 {
			t.Fatalf("batch=%+v %v", plan, e)
		}
		cs, e := other.store.ListChapters(ctx, f.owner, f.project.ID)
		if e != nil || len(cs) != 2 || cs[0].Number == cs[1].Number {
			t.Fatal("two distinct chapters missing")
		}
		before := f.http.count()
		if _, e = other.ResumeBatch(ctx, f.owner, f.project.ID, p.ID); e == nil {
			t.Fatal("completed resumed")
		}
		other.AdvanceBatches(ctx)
		if f.http.count() != before {
			t.Fatal("repeat generated")
		}
		t.Logf("batch=%s saved=2 chapters=2 HTTP=%d repeat delta=0", p.ID, before)
	})
	t.Run("rejection", func(t *testing.T) {
		f := parityFixtureFor(t, "short_original")
		f.http.reject = true
		ctx := context.Background()
		p, e := f.service.CreateBatch(ctx, f.owner, f.project.ID, BatchInput{SessionID: f.session.ID, Prompt: "失败", TotalChapters: 2, MaxConsecutiveFailures: 1})
		if e != nil {
			t.Fatal(e)
		}
		runs, e := f.service.AdvanceBatches(ctx)
		if e != nil || len(runs) != 1 {
			t.Fatal(e)
		}
		if _, e = f.service.Run(ctx, f.owner, runs[0].Run.ID); e == nil {
			t.Fatal("429 accepted")
		}
		f.service.AdvanceBatches(ctx)
		p, e = f.store.GetBatch(ctx, f.owner, p.ID)
		if e != nil || p.CompletedChapters != 0 || p.ConsecutiveFailures != 1 || (p.Status != BatchStopped && p.Status != BatchFailed) {
			t.Fatalf("failure counted %+v %v", p, e)
		}
		if f.http.count() != 1 {
			t.Fatal("429 repeated")
		}
		t.Logf("batch=%s HTTP429=1 saved=0 failures=1 status=%s", p.ID, p.Status)
	})
}

// Real child processes both contend on the same database, with no shared Go heap.
func TestNovelParityRealMySQLCreationProcessHelper(t *testing.T) {
	if os.Getenv("QIANTIE_TEST_NOVEL_CHILD") != "1" {
		t.Skip("child process only")
	}
	db := parityDB(t)
	s := NewService(NewMySQLStore(db), NewMemoryRuntime(), nil)
	p, e := s.CreateSchedule(context.Background(), os.Getenv("QIANTIE_TEST_NOVEL_OWNER"), os.Getenv("QIANTIE_TEST_NOVEL_PROJECT"), ScheduleInput{SessionID: os.Getenv("QIANTIE_TEST_NOVEL_SESSION"), Prompt: "跨进程创建", IdempotencyKey: "process-shared"})
	if e != nil {
		t.Fatal(e)
	}
	fmt.Printf("OWNED_PID=%d OWNED_PLAN=%s\n", os.Getpid(), p.ID)
}
func TestNovelParityRealMySQLCrossProcessCreation(t *testing.T) {
	f := parityFixtureFor(t, "short_original")
	exe, e := os.Executable()
	if e != nil {
		t.Fatal(e)
	}
	var cmds []*exec.Cmd
	for i := 0; i < 2; i++ {
		cmd := exec.Command(exe, "-test.run=^TestNovelParityRealMySQLCreationProcessHelper$", "-test.v")
		cmd.Env = append(os.Environ(), "QIANTIE_TEST_NOVEL_CHILD=1", "QIANTIE_TEST_NOVEL_OWNER="+f.owner, "QIANTIE_TEST_NOVEL_PROJECT="+f.project.ID, "QIANTIE_TEST_NOVEL_SESSION="+f.session.ID)
		cmds = append(cmds, cmd)
	}
	type result struct {
		b []byte
		e error
	}
	out := make(chan result, 2)
	for _, cmd := range cmds {
		go func(cmd *exec.Cmd) { b, e := cmd.CombinedOutput(); out <- result{b, e} }(cmd)
	}
	for i := 0; i < 2; i++ {
		r := <-out
		if r.e != nil {
			t.Fatalf("child failed: %s", r.b)
		}
		t.Log(strings.TrimSpace(string(r.b)))
	}
	ps, e := f.store.ListSchedules(context.Background(), f.owner, f.project.ID)
	if e != nil || len(ps) != 1 {
		t.Fatalf("cross-process plans=%d %v", len(ps), e)
	}
	t.Logf("2 OS children same DB => schedule=%s count=1 HTTP=0", ps[0].ID)
}

func TestNovelParityRealMySQLArchivedBatchesDoNotStarveActive(t *testing.T) {
	f := parityFixtureFor(t, "short_original")
	ctx := context.Background()
	for i := 0; i < 20; i++ {
		p, e := f.service.CreateProject(ctx, f.owner, CreateProjectInput{Title: fmt.Sprintf("验收归档批次%d", i), Genre: "悬疑", ModelID: "np-fake"})
		if e != nil {
			t.Fatal(e)
		}
		se, e := f.service.CreateSession(ctx, f.owner, CreateSessionInput{ProjectID: p.ID, Mode: "short_original"})
		if e != nil {
			t.Fatal(e)
		}
		b, e := f.service.CreateBatch(ctx, f.owner, p.ID, BatchInput{SessionID: se.ID, TotalChapters: 1})
		if e != nil {
			t.Fatal(e)
		}
		if e = f.service.ArchiveProject(ctx, f.owner, p.ID); e != nil {
			t.Fatal(e)
		}
		if i%2 == 1 {
			if _, e = f.db.Exec(`UPDATE novel_projects SET status='purging' WHERE owner_username=? AND project_id=?`, f.owner, p.ID); e != nil {
				t.Fatal(e)
			}
		}
		if _, e = f.db.Exec(`UPDATE novel_batches SET updated_at='2000-01-01 00:00:00' WHERE owner_username=? AND batch_id=?`, f.owner, b.ID); e != nil {
			t.Fatal(e)
		}
	}
	normal, e := f.service.CreateBatch(ctx, f.owner, f.project.ID, BatchInput{SessionID: f.session.ID, TotalChapters: 1})
	if e != nil {
		t.Fatal(e)
	}
	dispatched, e := f.service.AdvanceBatches(ctx)
	if e != nil {
		t.Fatalf("archived batches obstruct active scanner: %v", e)
	}
	found := false
	for _, d := range dispatched {
		if d.Batch.ID == normal.ID {
			found = true
		}
	}
	if !found {
		t.Fatal("21st active batch starved behind20 archived/purging rows")
	}
	p, e := f.store.GetBatch(ctx, f.owner, normal.ID)
	if e != nil || p.CurrentRunID == "" {
		t.Fatal("normal batch not durably queued")
	}
	if _, e = f.service.Run(ctx, f.owner, p.CurrentRunID); e != nil {
		t.Fatal(e)
	}
	if _, e = f.service.AdvanceBatches(ctx); e != nil {
		t.Fatal(e)
	}
	t.Logf("20 archived/purging rows skipped; active batch=%s run=%s saved exactly once", normal.ID, p.CurrentRunID)
}
