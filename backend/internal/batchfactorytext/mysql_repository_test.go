package batchfactorytext

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"os"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/go-sql-driver/mysql"
	"qiantie/backend/internal/storage"
)

func previousPageRows() *sqlmock.Rows {
	return sqlmock.NewRows(strings.Fields("id version pagination ordinal start end text digest layout background source locked render artifact status failedStage errorType lastError revision"))
}

// A failed insert must roll back before the document's pagination pointer advances.
func TestRepositoryReplaceRollsBackOnPageInsert(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := NewMySQLRepository(db)
	s := Scope{"alice", "batch", "book"}
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id, .* FOR UPDATE").WithArgs("alice", "batch", "book").WillReturnRows(sqlmock.NewRows([]string{"id", "active_version_id", "pagination_version", "revision"}).AddRow("doc", "version", 2, 5))
	mock.ExpectQuery("SELECT v.production_text,d.config_snapshot_json FROM").WithArgs("alice", "book", "version").WillReturnRows(sqlmock.NewRows([]string{"production_text", "snapshot"}).AddRow("一。", nil))
	mock.ExpectQuery("SELECT p.id,.* ORDER BY p.ordinal_no FOR UPDATE").WithArgs("alice", "batch", "book").WillReturnRows(previousPageRows())
	fail := errors.New("page insert failed")
	mock.ExpectExec("INSERT INTO batch_factory_v12_text_pages").WillReturnError(fail)
	mock.ExpectRollback()
	_, err = repo.ReplaceSystemPages(context.Background(), s, PaginationUpdate{VersionID: "version", ExpectedRevision: 5, Pages: []PageDraft{{Text: "一。", StartOffset: 0, EndOffset: 2}}})
	if !errors.Is(err, fail) {
		t.Fatalf("got %v", err)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

// TEXT_REPO_TEST_DSN must point at a dedicated, disposable MySQL test schema.
func mysqlBook(t *testing.T) (*MySQLRepository, Scope, *sql.DB) {
	t.Helper()
	dsn := os.Getenv("TEXT_REPO_TEST_DSN")
	if dsn == "" {
		t.Skip("set TEXT_REPO_TEST_DSN to run against an isolated MySQL test schema")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	ctx := context.Background()
	exec := func(q string, args ...any) {
		t.Helper()
		if _, err := db.ExecContext(ctx, q, args...); err != nil {
			t.Fatal(err)
		}
	}
	exec("CREATE TABLE IF NOT EXISTS batch_factory_v12_books(id VARCHAR(64) PRIMARY KEY, owner_username VARCHAR(191) NOT NULL, batch_id VARCHAR(64) NOT NULL) ENGINE=InnoDB")
	for _, ddl := range storage.V12TextProductionStatements() {
		exec(ddl)
	}
	var id string
	if err := db.QueryRow("SELECT UUID()").Scan(&id); err != nil {
		t.Fatal(err)
	}
	s := Scope{"alice", "batch", id}
	exec("INSERT INTO batch_factory_v12_books VALUES(?,?,?)", id, s.Owner, s.BatchID)
	return NewMySQLRepository(db), s, db
}

func TestRepositoryMySQLPageLifecycle(t *testing.T) {
	ctx := context.Background()
	repo, s, db := mysqlBook(t)
	wrong := s
	wrong.Owner = "bob"
	wrongBatch := s
	wrongBatch.BatchID = "other"
	for _, scope := range []Scope{wrong, wrongBatch} {
		if _, err := repo.SaveOriginal(ctx, scope, "伪"); !errors.Is(err, ErrNotFound) {
			t.Fatalf("foreign original: %v", err)
		}
	}
	const raw = "甲。乙。丙。"
	doc, err := repo.SaveOriginal(ctx, s, raw)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.SaveOriginal(ctx, s, "替换"); !errors.Is(err, ErrConflict) {
		t.Fatalf("original replaced: %v", err)
	}
	reused, err := repo.SaveOriginal(ctx, s, raw)
	if err != nil || reused.ID != doc.ID || reused.Revision != doc.Revision {
		t.Fatalf("original repeat: %+v %v", reused, err)
	}
	v, err := repo.CreateVersion(ctx, s, VersionInput{ExpectedRevision: doc.Revision, ProductionText: "甲。乙。", SensitiveMode: "system"})
	if err != nil {
		t.Fatal(err)
	}
	doc, err = repo.GetDocument(ctx, s)
	if err != nil {
		t.Fatal(err)
	}
	drafts := []PageDraft{{Text: "甲。", EndOffset: 2}, {Text: "乙。", StartOffset: 2, EndOffset: 4}}
	replace := func(p []PageDraft) (int64, error) {
		return repo.ReplaceSystemPages(ctx, s, PaginationUpdate{ExpectedRevision: doc.Revision, VersionID: v.ID, Pages: p})
	}
	// The second draft fails after the first page has already been inserted.
	broken := append([]PageDraft(nil), drafts...)
	broken[1].Layout = &LayoutConfig{FontSize: math.Inf(1)}
	if _, err := replace(broken); err == nil {
		t.Fatal("accepted invalid layout")
	}
	still, err := repo.GetDocument(ctx, s)
	if err != nil || still.Revision != doc.Revision || still.PaginationVersion != 0 {
		t.Fatalf("published failed pagination: %+v %v", still, err)
	}
	var leftover int
	if err = db.QueryRow("SELECT COUNT(*) FROM batch_factory_v12_text_pages WHERE owner_username=? AND book_id=?", s.Owner, s.BookID).Scan(&leftover); err != nil || leftover != 0 {
		t.Fatalf("partial write survived rollback: %d %v", leftover, err)
	}
	if n, err := replace(drafts); err != nil || n != 1 {
		t.Fatalf("first pages %d %v", n, err)
	}
	pages, err := repo.ListPages(ctx, s)
	if err != nil || len(pages) != 2 {
		t.Fatalf("list %+v %v", pages, err)
	}
	hash := strings.Repeat("a", 64)
	file := ArtifactFile{StorageRef: "test/upload.png", MediaType: "image/png", ByteSize: 8, SHA256: hash}
	a, err := repo.LockPageArtifact(ctx, s, pages[1].ID, pages[1].Revision, file)
	if err != nil {
		t.Fatal(err)
	}
	pages, err = repo.ListPages(ctx, s)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.PatchPage(ctx, s, pages[1].ID, pages[1].Revision, PagePatch{Layout: &LayoutConfig{FontSize: 45}}); !errors.Is(err, ErrLocked) {
		t.Fatalf("locked patch: %v", err)
	}
	if _, err := repo.PatchPage(ctx, s, pages[0].ID, pages[0].Revision+1, PagePatch{Layout: &LayoutConfig{FontSize: 45}}); !errors.Is(err, ErrConflict) {
		t.Fatalf("stale patch: %v", err)
	}
	if _, err := repo.PatchPage(ctx, s, pages[0].ID, pages[0].Revision, PagePatch{Layout: &LayoutConfig{FontSize: 45}}); err != nil {
		t.Fatal(err)
	}
	doc, err = repo.GetDocument(ctx, s)
	if err != nil || doc.OriginalText != raw {
		t.Fatalf("edited original: %+v %v", doc, err)
	}
	if _, err := replace([]PageDraft{{Text: "甲。乙。", EndOffset: 4}}); !errors.Is(err, ErrLocked) {
		t.Fatalf("moved locked page: %v", err)
	}
	if n, err := replace(drafts); err != nil || n != 2 {
		t.Fatalf("repagination %d %v", n, err)
	}
	active, err := repo.ListPages(ctx, s)
	if err != nil || len(active) != 2 {
		t.Fatalf("active pages: %+v %v", active, err)
	}
	if active[1].ID != pages[1].ID || !active[1].Locked || active[1].ActiveArtifactID != a.ID {
		t.Fatalf("lost locked page: %+v", active[1])
	}
	if _, err := repo.CreateCollection(ctx, s, CollectionInput{VersionID: v.ID, PaginationVersion: 2}); !errors.Is(err, ErrNotReady) {
		t.Fatalf("collection with missing artifacts: %v", err)
	}
	if _, err := repo.SaveSystemPageArtifact(ctx, s, active[0].ID, active[0].Revision, file); err != nil {
		t.Fatal(err)
	}
	doc, err = repo.GetDocument(ctx, s)
	if err != nil {
		t.Fatal(err)
	}
	cover, err := repo.SaveCover(ctx, s, doc.Revision, file, true)
	if err != nil {
		t.Fatal(err)
	}
	c, err := repo.CreateCollection(ctx, s, CollectionInput{VersionID: v.ID, PaginationVersion: 2})
	if err != nil {
		t.Fatal(err)
	}
	if c.CoverArtifactID != cover.ID || len(c.Items) != 2 || c.Items[1].ArtifactID != a.ID {
		t.Fatalf("bad snapshot %+v", c)
	}
	doc, err = repo.GetDocument(ctx, s)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = repo.SaveCover(ctx, s, doc.Revision, file, false); !errors.Is(err, ErrLocked) {
		t.Fatalf("automation overwrote cover: %v", err)
	}
	newVersion, err := repo.CreateVersion(ctx, s, VersionInput{ExpectedRevision: doc.Revision, ProductionText: "甲。", SensitiveMode: "off"})
	if err != nil {
		t.Fatal(err)
	}
	doc, err = repo.GetDocument(ctx, s)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.ReplaceSystemPages(ctx, s, PaginationUpdate{ExpectedRevision: doc.Revision, VersionID: newVersion.ID, Pages: drafts[:1]}); !errors.Is(err, ErrLocked) {
		t.Fatalf("shorter new version silently lost locked page: %v", err)
	}
	if _, err := repo.SaveSystemPageArtifact(ctx, s, active[1].ID, active[1].Revision, file); !errors.Is(err, ErrConflict) {
		t.Fatalf("stale render: %v", err)
	}
	visible, err := repo.ListPages(ctx, s)
	if err != nil || len(visible) != 2 || !visible[1].Locked {
		t.Fatalf("locked conflict hidden from editor: %+v %v", visible, err)
	}
	if _, err := repo.RestoreSystemPage(ctx, s, active[1].ID, active[1].Revision); err != nil {
		t.Fatal(err)
	}
	active, err = repo.ListPages(ctx, s)
	if err != nil || active[1].Locked || active[1].ActiveArtifactID != "" {
		t.Fatalf("restore %+v %v", active, err)
	}
	doc, err = repo.GetDocument(ctx, s)
	if err != nil {
		t.Fatal(err)
	}
	if n, err := repo.ReplaceSystemPages(ctx, s, PaginationUpdate{ExpectedRevision: doc.Revision, VersionID: newVersion.ID, Pages: drafts[:1]}); err != nil || n != 3 {
		t.Fatalf("retry after unlock: %d %v", n, err)
	}
	frozen, err := repo.GetCollection(ctx, s, c.ID)
	if err != nil || frozen.Items[1].ArtifactID != a.ID || frozen.Items[1].StorageRef != file.StorageRef {
		t.Fatalf("lost frozen page %+v %v", frozen, err)
	}
	if _, err := repo.GetCollection(ctx, wrongBatch, c.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("wrong batch: %v", err)
	}
	if _, err := repo.GetCollection(ctx, wrong, c.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("foreign collection: %v", err)
	}
	if versions, err := repo.ListVersions(ctx, s); err != nil || len(versions) != 2 || versions[0].ProductionText != "甲。乙。" {
		t.Fatalf("versions: %+v %v", versions, err)
	}
	if list, err := repo.ListCollections(ctx, s); err != nil || len(list) != 1 || list[0].ID != c.ID {
		t.Fatalf("collections: %+v %v", list, err)
	}
	if files, err := repo.ListArtifacts(ctx, s); err != nil || len(files) != 3 {
		t.Fatalf("artifacts: %+v %v", files, err)
	}
	doc, err = repo.GetDocument(ctx, s)
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.RestoreCover(ctx, s, doc.Revision); err != nil {
		t.Fatal(err)
	}
	doc, err = repo.GetDocument(ctx, s)
	if err != nil || doc.CoverLocked || doc.CoverArtifactID != "" || doc.OriginalText != raw {
		t.Fatalf("restored cover: %+v %v", doc, err)
	}
}

func TestRepositoryLockedPageCannotMoveEvenIfUploadIsNotCurrentlyActive(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id, .* FOR UPDATE").WithArgs("alice", "batch", "book").WillReturnRows(sqlmock.NewRows([]string{"id", "active_version_id", "pagination_version", "revision"}).AddRow("doc", "v2", 1, 7))
	mock.ExpectQuery("SELECT v.production_text,d.config_snapshot_json FROM").WithArgs("alice", "book", "v2").WillReturnRows(sqlmock.NewRows([]string{"production_text", "snapshot"}).AddRow("甲。", nil))
	mock.ExpectQuery("SELECT p.id,.* ORDER BY p.ordinal_no FOR UPDATE").WithArgs("alice", "batch", "book").WillReturnRows(previousPageRows().AddRow("upload", "v1", 1, 2, 2, 4, "乙。", "abc", nil, nil, "user_upload", true, 2, "file", "ready", "", "", "", 2))
	mock.ExpectRollback()
	_, err = NewMySQLRepository(db).ReplaceSystemPages(context.Background(), Scope{"alice", "batch", "book"}, PaginationUpdate{ExpectedRevision: 7, VersionID: "v2", Pages: []PageDraft{{Text: "甲。", EndOffset: 2}}})
	if !errors.Is(err, ErrLocked) {
		t.Fatalf("locked page discarded: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRepositoryPaginationAcceptsNormalizedNewlinesAndExactRunes(t *testing.T) {
	drafts := []PageDraft{{Text: "甲。\n", EndOffset: 3}, {Text: "🙂。\n乙。", StartOffset: 3, EndOffset: 8}}
	if err := validateDrafts("甲。\r\n🙂。\r乙。", drafts); err != nil {
		t.Fatalf("normalized pages rejected: %v", err)
	}
	drafts[1].Text = "🙂。\n丙。"
	if err := validateDrafts("甲。\r\n🙂。\r乙。", drafts); !errors.Is(err, ErrInvalid) {
		t.Fatalf("changed content accepted: %v", err)
	}
}

func TestRepositoryMySQLReusesOnlyUnchangedPages(t *testing.T) {
	ctx := context.Background()
	repo, s, _ := mysqlBook(t)
	doc, err := repo.SaveOriginal(ctx, s, "甲。乙。丙。")
	if err != nil {
		t.Fatal(err)
	}
	refresh := func() {
		t.Helper()
		doc, err = repo.GetDocument(ctx, s)
		if err != nil {
			t.Fatal(err)
		}
	}
	version := func(text string) Version {
		t.Helper()
		refresh()
		v, e := repo.CreateVersion(ctx, s, VersionInput{ExpectedRevision: doc.Revision, ProductionText: text, SensitiveMode: "off"})
		if e != nil {
			t.Fatal(e)
		}
		return v
	}
	v := version("甲。乙。")
	drafts := []PageDraft{{Text: "甲。", EndOffset: 2}, {Text: "乙。", StartOffset: 2, EndOffset: 4}}
	config := []byte(`{"layout":{"fontSize":30},"background":{"kind":"color","color":"#101010"}}`)
	replace := func(drafts []PageDraft) error {
		t.Helper()
		refresh()
		_, e := repo.ReplaceSystemPages(ctx, s, PaginationUpdate{ExpectedRevision: doc.Revision, VersionID: v.ID, ConfigSnapshot: config, Pages: drafts})
		return e
	}
	list := func() []Page {
		t.Helper()
		p, e := repo.ListPages(ctx, s)
		if e != nil || len(p) != 2 {
			t.Fatalf("pages %+v %v", p, e)
		}
		return p
	}
	saveFile := func(p Page) PageArtifact {
		t.Helper()
		a, e := repo.SaveSystemPageArtifact(ctx, s, p.ID, p.Revision, ArtifactFile{StorageRef: "test/" + p.ID, MediaType: "image/png", ByteSize: 8, SHA256: strings.Repeat("e", 64)})
		if e != nil {
			t.Fatal(e)
		}
		return a
	}
	if err = replace(drafts); err != nil {
		t.Fatal(err)
	}
	p := list()
	bg := Background{Kind: "color", Color: "#A12345"}
	layout := LayoutConfig{FontSize: 40}
	p[0], err = repo.PatchPage(ctx, s, p[0].ID, p[0].Revision, PagePatch{Layout: &layout, Background: &bg})
	if err != nil {
		t.Fatal(err)
	}
	firstArtifact := saveFile(p[0])
	saveFile(p[1])
	refresh()
	_, err = repo.SaveCover(ctx, s, doc.Revision, firstArtifact.ArtifactFile, false)
	if err != nil {
		t.Fatal(err)
	}
	before := list()
	v = version("甲。丙。")
	drafts[1].Text = "丙。"
	broken := append([]PageDraft(nil), drafts...)
	broken[1].Layout = &LayoutConfig{FontSize: math.NaN()}
	if err = replace(broken); err == nil {
		t.Fatal("bad layout unexpectedly succeeded")
	}
	rolledBack := list()
	if rolledBack[0].Revision != before[0].Revision || rolledBack[0].PaginationVersion != before[0].PaginationVersion || rolledBack[0].ActiveArtifactID != firstArtifact.ID || rolledBack[1].ActiveArtifactID != before[1].ActiveArtifactID {
		t.Fatalf("partial repagination: %+v", rolledBack)
	}
	if err = replace(drafts); err != nil {
		t.Fatal(err)
	}
	after := list()
	if after[0].ID != before[0].ID || after[0].Revision != before[0].Revision || after[0].Status != "ready" || after[0].ActiveArtifactID != firstArtifact.ID || after[0].Background == nil || *after[0].Background != bg || after[0].Layout == nil || *after[0].Layout != layout {
		t.Fatalf("replaced unchanged first page: %+v", after[0])
	}
	if after[1].ActiveArtifactID != "" || after[1].Status != "pending" {
		t.Fatalf("reused changed second page: %+v", after[1])
	}
	snapshot, err := repo.CreateCollection(ctx, s, CollectionInput{VersionID: v.ID, PaginationVersion: after[0].PaginationVersion, PageIDs: []string{after[0].ID}})
	if err != nil || len(snapshot.Items) != 1 || snapshot.Items[0].ArtifactID != firstArtifact.ID {
		t.Fatalf("unchanged page not usable: %+v %v", snapshot, err)
	}
	saveFile(after[1])
	config = []byte(`{"layout":{"fontSize":53},"background":{"kind":"color","color":"#FFFFFF"}}`)
	if err = replace(drafts); err != nil {
		t.Fatal(err)
	}
	after = list()
	if after[0].ActiveArtifactID != firstArtifact.ID || after[1].ActiveArtifactID != "" {
		t.Fatalf("incorrect book style invalidation: %+v", after)
	}
	drafts[0].ClearBackground = true
	if err = replace(drafts); err != nil {
		t.Fatal(err)
	}
	after = list()
	if after[0].Background != nil || after[0].ActiveArtifactID != "" || after[0].Layout == nil || *after[0].Layout != layout {
		t.Fatalf("explicit clear lost or discarded other override: %+v", after[0])
	}
	rendered := saveFile(after[0])
	// Omitting the book snapshot on a targeted repagination keeps the
	// existing book defaults. An explicit {} resets them.
	config = nil
	if err = replace(drafts); err != nil {
		t.Fatal(err)
	}
	after = list()
	if after[0].ActiveArtifactID != rendered.ID {
		t.Fatalf("omitting the book snapshot reset its background: %+v", after[0])
	}
	config = []byte(`{"layout":{"fontSize":53},"background":{"kind":"color","color":"#EFEFEF"}}`)
	drafts[0].ClearBackground = false
	if err = replace(drafts); err != nil {
		t.Fatal(err)
	}
	after = list()
	if after[0].ActiveArtifactID == rendered.ID || after[0].Status != "pending" {
		t.Fatalf("inherited background still stale: %+v", after[0])
	}
}

func TestRepositoryResolvePageStyleInheritanceAndExplicitClear(t *testing.T) {
	book := PageStyle{Layout: LayoutConfig{FontID: "noto", FontSize: 32}, Background: Background{Kind: "color", Color: "#111111"}}
	layout := LayoutConfig{FontID: "xiangHai", FontSize: 48}
	bg := Background{Kind: "color", Color: "#658933"}
	prior := PageDraft{Layout: &layout, Background: &bg}
	p, style, err := ResolvePageStyle(book, prior, PagePatch{})
	if err != nil || p.Layout == nil || p.Background == nil || style.Layout != layout || style.Background != bg {
		t.Fatalf("retain: %+v %+v %v", p, style, err)
	}
	p, style, err = ResolvePageStyle(book, prior, PagePatch{ClearLayout: true})
	if err != nil || p.Layout != nil || p.Background == nil || style.Layout != book.Layout || style.Background != bg {
		t.Fatalf("clear layout: %+v %+v %v", p, style, err)
	}
	_, _, err = ResolvePageStyle(book, prior, PagePatch{Layout: &layout, ClearLayout: true})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("ambiguous update %v", err)
	}
	_, _, err = ResolvePageStyle(book, prior, PagePatch{Layout: &LayoutConfig{}})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("empty override %v", err)
	}
}

func TestRepositoryReusesSystemPageWithoutOverwritingItsRevisionOrArtifact(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id, .* FOR UPDATE").WithArgs("alice", "batch", "book").WillReturnRows(sqlmock.NewRows([]string{"id", "version_id", "pagination_version", "revision"}).AddRow("doc", "v2", 2, 4))
	mock.ExpectQuery("SELECT v.production_text,d.config_snapshot_json FROM").WithArgs("alice", "book", "v2").WillReturnRows(sqlmock.NewRows([]string{"text", "snapshot"}).AddRow("甲。", nil))
	mock.ExpectQuery("SELECT p.id,.* ORDER BY p.ordinal_no FOR UPDATE").WithArgs("alice", "batch", "book").WillReturnRows(previousPageRows().AddRow("p", "v1", 2, 1, 0, 2, "甲。", "digest", nil, nil, "system", false, 2, "artifact", "ready", "", "", "", 7))
	mock.ExpectExec("UPDATE batch_factory_v12_text_pages SET version_id=.*,pagination_version=.*,ordinal_no=.*,updated_at=NOW.* WHERE id=").WithArgs("v2", 3, 1, "p").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM batch_factory_v12_text_pages").WithArgs("alice", "book", 2).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("UPDATE batch_factory_v12_text_documents SET pagination_version=").WithArgs(3, nil, "doc").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	pagination, err := NewMySQLRepository(db).ReplaceSystemPages(context.Background(), Scope{"alice", "batch", "book"}, PaginationUpdate{ExpectedRevision: 4, VersionID: "v2", Pages: []PageDraft{{Text: "甲。", EndOffset: 2}}})
	if err != nil || pagination != 3 {
		t.Fatalf("pagination=%d, err=%v", pagination, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
