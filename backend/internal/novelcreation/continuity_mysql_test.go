package novelcreation

import (
	"context"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"testing"
	"time"
)

func TestMySQLContinuityMemoryIsBoundToOwnerProjectAndChapterVersion(t *testing.T) {
	for _, kind := range []string{"valid", "wrong-memory", "missing", "read-error"} {
		t.Run(kind, func(t *testing.T) {
			store, mock := newNovelCreationMockStore(t)
			now := time.Now().UTC()
			mock.ExpectQuery("SELECT project_id, title").WithArgs("alice", "project").WillReturnRows(sqlmock.NewRows([]string{"id", "title", "genre", "world", "protagonist", "conflict", "model", "words", "status", "created", "updated", "deleted"}).AddRow("project", "书", "悬疑", nil, nil, nil, "text", 2000, "draft", now, now, nil))
			query := mock.ExpectQuery(`(?s)SELECT m.memory_id.*v.version_no=c.version_no AND BINARY v.content=BINARY c.content.*r.run_id=v.run_id.*WHERE m.owner_username=\? AND m.project_id=\? AND m.chapter_id=\?`).WithArgs("alice", "project", "chapter")
			rows := sqlmock.NewRows([]string{"id", "project", "chapter", "snapshot", "created", "updated", "version", "runtime"})
			if kind == "read-error" {
				query.WillReturnError(errors.New("database failed"))
			} else {
				if kind != "missing" {
					memory := `{"summary":"前文"}`
					if kind == "wrong-memory" {
						memory = `{"summary":"旧版"}`
					}
					rows.AddRow("memory", "project", "chapter", memory, now, now, 3, `{"truthSettlement":{"facts":["事实"]},"memoryIndex":{"summary":"前文"}}`)
				}
				query.WillReturnRows(rows)
			}
			memory, found, err := store.GetChapterMemory(context.Background(), "alice", "project", "chapter")
			switch kind {
			case "valid":
				if err != nil || !found || memory.ChapterVersion != 3 || string(memory.PreviousTruth) != `{"facts":["事实"]}` {
					t.Fatalf("memory=%+v found=%v err=%v", memory, found, err)
				}
			case "missing":
				if err != nil || found {
					t.Fatalf("found=%v err=%v", found, err)
				}
			default:
				if err == nil {
					t.Fatal("invalid continuity accepted")
				}
			}
			if err = mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
func TestContinuityQueuedChapterEditStopsBeforePipeline(t *testing.T) {
	f := newServiceFixture(t, nil)
	continuityFirstChapter(t, f)
	run, err := f.service.SendMessage(context.Background(), "alice", f.sendInput("second"))
	if err != nil {
		t.Fatal(err)
	}
	key := memoryKey("alice", f.project.ID, f.project.ID+"-chapter-1")
	b := f.store.chapterBundles[key]
	b.Chapter.Version = 2
	f.store.chapterBundles[key] = b
	if _, err = f.service.Run(context.Background(), "alice", run.ID); err == nil {
		t.Fatal("edited predecessor accepted")
	}
	if f.runner.callCount() != 0 {
		t.Fatal("pipeline called after predecessor edited")
	}
}
func TestMySQLContinuityRetrySavesImmutableNewVersion(t *testing.T) {
	store, mock := newNovelCreationMockStore(t)
	ctx := context.Background()
	now := time.Now().UTC()
	b := validMySQLChapterBundle()
	b.RunID = "retry-run"
	b.Chapter.Version = 2
	b.ChapterVersion.ID = "retry-version"
	b.ChapterVersion.Version = 2
	b.ChapterVersion.RunID = b.RunID
	b.Foundation = Foundation{}
	b.Audits = nil
	b.Revisions = nil
	b.Memories = []Memory{{ID: "retry-memory", Snapshot: []byte(`{"summary":"重试记忆"}`)}}
	b.RuntimeSnapshot = []byte(`{"truthSettlement":{"facts":["重试事实"]},"memoryIndex":{"summary":"重试记忆"}}`)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT 1 FROM novel_projects").WithArgs("alice", b.ProjectID).WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectQuery("SELECT 1 FROM novel_runs").WithArgs("alice", b.RunID, b.ProjectID).WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectExec("INSERT INTO novel_chapters").WillReturnResult(sqlmock.NewResult(1, 1))
	// The historical row must never be updated on a duplicate chapter/version key.
	mock.ExpectExec(`(?s)^INSERT INTO novel_chapter_versions .*VALUES \(\?, \?, \?, \?, \?, \?, \?, \?, \?\)$`).WithArgs(b.ChapterVersion.ID, "alice", b.ProjectID, b.Chapter.ID, b.RunID, 2, b.Chapter.Content, []byte(b.ChapterVersion.Snapshot), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO novel_memories").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE novel_runs SET status").WithArgs(b.FinalRunStatus, []byte(b.RuntimeSnapshot), sqlmock.AnyArg(), sqlmock.AnyArg(), "alice", b.RunID, b.ProjectID).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	if err := store.SaveChapterBundle(ctx, "alice", b); err != nil {
		t.Fatal(err)
	}
	mock.ExpectQuery("SELECT project_id, title").WithArgs("alice", b.ProjectID).WillReturnRows(sqlmock.NewRows([]string{"id", "title", "genre", "world", "protagonist", "conflict", "model", "words", "status", "created", "updated", "deleted"}).AddRow(b.ProjectID, "书", "悬疑", nil, nil, nil, "text", 2000, "draft", now, now, nil))
	mock.ExpectQuery(`(?s)SELECT m.memory_id.*v.version_no=c.version_no.*r.run_id=v.run_id`).WithArgs("alice", b.ProjectID, b.Chapter.ID).WillReturnRows(sqlmock.NewRows([]string{"id", "project", "chapter", "snapshot", "created", "updated", "version", "runtime"}).AddRow("retry-memory", b.ProjectID, b.Chapter.ID, string(b.Memories[0].Snapshot), now, now, 2, string(b.RuntimeSnapshot)))
	memory, found, err := store.GetChapterMemory(ctx, "alice", b.ProjectID, b.Chapter.ID)
	if err != nil || !found || memory.ChapterVersion != 2 || string(memory.PreviousTruth) != `{"facts":["重试事实"]}` {
		t.Fatalf("next chapter lost mysql retry continuity: %+v %v", memory, err)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
