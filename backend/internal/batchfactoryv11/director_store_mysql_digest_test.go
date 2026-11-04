package batchfactoryv11

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func originalModePatch() SettingsPatch {
	return SettingsPatch{"mode": json.RawMessage(`"original"`)}
}

func TestPersistDirectorRevisionReusesVideoSlotsWhenDigestIsUnchanged(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	book := Book{ID: "book-1", BatchID: "batch-1"}
	snapshot := DirectorSnapshot{Mode: "original", Effective: originalModePatch()}
	output := DirectorResult{Storyboard: []DirectorVideo{{DurationSec: 9}}}

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT revision FROM batch_factory_v12_books`).
		WithArgs("book-1", "batch-1", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"revision"}).AddRow(3))
	mock.ExpectQuery(`SELECT id,revision,snapshot_id,mode,source_digest,hook_revision_id,output_json,created_at FROM batch_factory_v12_director_revisions`).
		WithArgs("batch-1", "book-1", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"id", "revision", "snapshot_id", "mode", "source_digest", "hook_revision_id", "output_json", "created_at"}).
			AddRow("director-old", 1, "snapshot-old", "original", "digest-abc", nil, `{"storyboard":[]}`, time.Now()))
	mock.ExpectQuery(`SELECT effective_json FROM batch_factory_v12_config_snapshots`).
		WithArgs("snapshot-old", "alice", "batch-1", "book-1").
		WillReturnRows(sqlmock.NewRows([]string{"effective_json"}).AddRow(`{"mode":"original"}`))
	mock.ExpectQuery(`SELECT v.id,COALESCE`).
		WithArgs("batch-1", "book-1", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"id", "label", "video_prompt", "visual_prompt", "duration_seconds", "ordinal", "revision"}).
			AddRow("video-keep", "VIDEO 01", "保持的提示词", "视觉", 9, 0, 1))
	mock.ExpectCommit()

	revision, err := store.PersistDirectorRevision(context.Background(), "alice", book, snapshot, "digest-abc", "", output)
	if err != nil {
		t.Fatal(err)
	}
	// 相同 digest + 相同冻结配置：复用修订，绝不轮换槽位。
	if revision.ID != "director-old" || revision.Revision != 1 {
		t.Fatalf("revision = %s/%d, want the existing one", revision.ID, revision.Revision)
	}
	if len(revision.Videos) != 1 || revision.Videos[0].ID != "video-keep" {
		t.Fatalf("videos = %+v, want the existing video slot untouched", revision.Videos)
	}
	if revision.Videos[0].CompatibilityState != "active" {
		t.Fatalf("compatibility state = %q, want active", revision.Videos[0].CompatibilityState)
	}
	if len(revision.OrphanedOverrides) != 0 {
		t.Fatalf("orphaned overrides = %+v, want none", revision.OrphanedOverrides)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("no rotation or insert may run when the plan is unchanged: %v", err)
	}
}

func TestReusableDirectorRevisionRejectsAChangedDigest(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	book := Book{ID: "book-1", BatchID: "batch-1"}

	mock.ExpectQuery(`SELECT id,revision,snapshot_id,mode,source_digest,hook_revision_id,output_json,created_at FROM batch_factory_v12_director_revisions`).
		WithArgs("batch-1", "book-1", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"id", "revision", "snapshot_id", "mode", "source_digest", "hook_revision_id", "output_json", "created_at"}).
			AddRow("director-old", 1, "snapshot-old", "original", "digest-old", nil, `{"storyboard":[]}`, time.Now()))

	reused, ok, err := store.reusableDirectorRevision(context.Background(), db, "alice", book,
		DirectorSnapshot{Mode: "original", Effective: originalModePatch()}, "digest-new", "", `{"mode":"original"}`)
	if err != nil {
		t.Fatal(err)
	}
	if ok || reused.ID != "" {
		t.Fatalf("reused = %+v ok=%v, want the normal rotation path for a changed digest", reused, ok)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestReusableDirectorRevisionRejectsChangedSettingsWithTheSameDigest(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	book := Book{ID: "book-1", BatchID: "batch-1"}

	mock.ExpectQuery(`SELECT id,revision,snapshot_id,mode,source_digest,hook_revision_id,output_json,created_at FROM batch_factory_v12_director_revisions`).
		WithArgs("batch-1", "book-1", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"id", "revision", "snapshot_id", "mode", "source_digest", "hook_revision_id", "output_json", "created_at"}).
			AddRow("director-old", 1, "snapshot-old", "original", "digest-abc", nil, `{"storyboard":[]}`, time.Now()))
	mock.ExpectQuery(`SELECT effective_json FROM batch_factory_v12_config_snapshots`).
		WithArgs("snapshot-old", "alice", "batch-1", "book-1").
		WillReturnRows(sqlmock.NewRows([]string{"effective_json"}).AddRow(`{"mode":"original","storyboardDurationLimit":10}`))

	reused, ok, err := store.reusableDirectorRevision(context.Background(), db, "alice", book,
		DirectorSnapshot{Mode: "original", Effective: originalModePatch()}, "digest-abc", "", `{"mode":"original"}`)
	if err != nil {
		t.Fatal(err)
	}
	if ok || reused.ID != "" {
		t.Fatalf("reused = %+v ok=%v, digest alone must not suppress a settings change", reused, ok)
	}
}

func TestCanonicalJSONEqualIgnoresFormatting(t *testing.T) {
	left := `{"b":2,"a":1}`
	right := `{ "a": 1.0, "b": 2 }`
	if !canonicalJSONEqual(left, right) {
		t.Fatal("canonical comparison must ignore key order and number formatting")
	}
	if canonicalJSONEqual(`{"a":1}`, `{"a":2}`) {
		t.Fatal("different settings must not compare equal")
	}
}

// 输入 digest 变化时必须走原有轮换：旧槽位 orphaned，新槽位重建。
func TestPersistDirectorRevisionRotatesSlotsWhenDigestChanges(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	book := Book{ID: "book-1", BatchID: "batch-1"}
	snapshot := DirectorSnapshot{Mode: "original", Effective: originalModePatch()}

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT revision FROM batch_factory_v12_books`).
		WithArgs("book-1", "batch-1", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"revision"}).AddRow(3))
	mock.ExpectQuery(`SELECT id,revision,snapshot_id,mode,source_digest,hook_revision_id,output_json,created_at FROM batch_factory_v12_director_revisions`).
		WithArgs("batch-1", "book-1", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"id", "revision", "snapshot_id", "mode", "source_digest", "hook_revision_id", "output_json", "created_at"}).
			AddRow("director-old", 1, "snapshot-old", "original", "digest-old", nil, `{"storyboard":[]}`, time.Now()))
	mock.ExpectExec(`INSERT INTO batch_factory_v12_config_snapshots`).
		WithArgs(sqlmock.AnyArg(), "alice", "batch-1", "book-1", string(`{"mode":"original"}`), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT COALESCE\(MAX\(revision\),0\) FROM batch_factory_v12_director_revisions`).
		WithArgs("batch-1", "book-1", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"revision"}).AddRow(1))
	mock.ExpectExec(`INSERT INTO batch_factory_v12_director_revisions`).
		WithArgs(sqlmock.AnyArg(), "batch-1", "book-1", "alice", sqlmock.AnyArg(), sqlmock.AnyArg(), "original", "digest-new", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT id,ordinal FROM batch_factory_v12_videos`).
		WithArgs("batch-1", "book-1", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"id", "ordinal"}).
			AddRow("video-old", 0))
	mock.ExpectQuery(`SELECT patch_json FROM batch_factory_v12_settings_patches`).
		WithArgs("video", "video-old", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"patch_json"}))
	mock.ExpectExec(`UPDATE batch_factory_v12_videos SET compatibility_state='orphaned'`).
		WithArgs(sqlmock.AnyArg(), "batch-1", "book-1", "alice").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO batch_factory_v12_videos`).
		WithArgs(sqlmock.AnyArg(), "batch-1", "book-1", "alice", 0, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO batch_factory_v12_video_records`).
		WithArgs(sqlmock.AnyArg(), "VIDEO 01", sqlmock.AnyArg(), sqlmock.AnyArg(), 9).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO batch_factory_v12_director_video_links`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), 0).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT id,kind,name FROM batch_factory_v12_book_assets`).
		WithArgs("batch-1", "book-1", "alice").
		WillReturnRows(sqlmock.NewRows([]string{"id", "kind", "name"}))
	mock.ExpectExec(`UPDATE batch_factory_v12_books SET revision=`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "book-1", "batch-1", "alice").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	output := DirectorResult{Storyboard: []DirectorVideo{{DurationSec: 9}}}
	revision, err := store.PersistDirectorRevision(context.Background(), "alice", book, snapshot, "digest-new", "", output)
	if err != nil {
		t.Fatal(err)
	}
	if revision.ID == "" || revision.ID == "director-old" || revision.SourceDigest != "digest-new" {
		t.Fatalf("revision = %+v, want a brand new revision for a changed digest", revision)
	}
	if len(revision.Videos) != 1 || revision.Videos[0].ID == "video-old" {
		t.Fatalf("videos = %+v, want a freshly rotated slot", revision.Videos)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("a changed digest must run the normal rotation path: %v", err)
	}
}

func TestPersistDirectorRevisionKeepsSlotsAcrossConsecutiveRetries(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	book := Book{ID: "book-1", BatchID: "batch-1"}
	// 自动化重试：先跑 director 再跑 video。两次 director 输入完全相同，
	// 第二次不得把第一次之后仍然 active 的成功成片废掉。
	for attempt := 0; attempt < 2; attempt++ {
		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT revision FROM batch_factory_v12_books`).
			WithArgs("book-1", "batch-1", "alice").
			WillReturnRows(sqlmock.NewRows([]string{"revision"}).AddRow(3))
		mock.ExpectQuery(`SELECT id,revision,snapshot_id,mode,source_digest,hook_revision_id,output_json,created_at FROM batch_factory_v12_director_revisions`).
			WithArgs("batch-1", "book-1", "alice").
			WillReturnRows(sqlmock.NewRows([]string{"id", "revision", "snapshot_id", "mode", "source_digest", "hook_revision_id", "output_json", "created_at"}).
				AddRow("director-old", 1, "snapshot-old", "original", "digest-abc", nil, `{"storyboard":[]}`, time.Now()))
		mock.ExpectQuery(`SELECT effective_json FROM batch_factory_v12_config_snapshots`).
			WithArgs("snapshot-old", "alice", "batch-1", "book-1").
			WillReturnRows(sqlmock.NewRows([]string{"effective_json"}).AddRow(`{"mode":"original"}`))
		mock.ExpectQuery(`SELECT v.id,COALESCE`).
			WithArgs("batch-1", "book-1", "alice").
			WillReturnRows(sqlmock.NewRows([]string{"id", "label", "video_prompt", "visual_prompt", "duration_seconds", "ordinal", "revision"}).
				AddRow("video-keep-1", "VIDEO 01", "提示词 1", "视觉 1", 9, 0, 1).
				AddRow("video-keep-2", "VIDEO 02", "提示词 2", "视觉 2", 9, 1, 1))
		mock.ExpectCommit()

		snapshot := DirectorSnapshot{Mode: "original", Effective: originalModePatch()}
		revision, err := store.PersistDirectorRevision(context.Background(), "alice", book, snapshot, "digest-abc", "", DirectorResult{Storyboard: []DirectorVideo{{DurationSec: 9}, {DurationSec: 9}}})
		if err != nil {
			t.Fatal(err)
		}
		if len(revision.Videos) != 2 || revision.Videos[0].ID != "video-keep-1" || revision.Videos[1].ID != "video-keep-2" {
			t.Fatalf("attempt %d videos = %+v, want the same active slots", attempt, revision.Videos)
		}
		for _, video := range revision.Videos {
			if video.CompatibilityState != "active" {
				t.Fatalf("attempt %d video %s state = %q, want active", attempt, video.ID, video.CompatibilityState)
			}
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("consecutive identical retries must not rotate slots: %v", err)
	}
}

