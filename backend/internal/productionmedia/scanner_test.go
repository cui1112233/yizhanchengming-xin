package productionmedia

import (
	"bytes"
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"qiantie/backend/internal/localartifact"
)

type recordingEnqueuer struct{ items []Object }

func (r *recordingEnqueuer) Enqueue(_ context.Context, item Object) error {
	r.items = append(r.items, item)
	return nil
}

type countingRestorer struct{ calls int }

func (r *countingRestorer) RestoreStorageRef(context.Context, string) error {
	r.calls++
	return nil
}

func TestScannerDiscoversExistingComicImageWithoutDuplicatingFile(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	files := localartifact.NewStore(t.TempDir(), 1<<20)
	restorer := &countingRestorer{}
	files.SetMissingRestorer(restorer)
	artifact, err := files.SaveImage("comic_123", "image/png", pngFixture())
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectQuery("SELECT i.owner_username").WillReturnRows(sqlmock.NewRows([]string{"owner_username", "storage_ref", "media_type"}).AddRow("alice", artifact.StorageRef, "image/png").AddRow("alice", "missing.png", "image/png"))
	mock.ExpectQuery("SELECT i.owner_username").WillReturnRows(sqlmock.NewRows([]string{"owner_username", "storage_ref", "media_type"}))
	mock.ExpectQuery("SELECT owner_username").WillReturnRows(sqlmock.NewRows([]string{"owner_username", "storage_ref", "media_type"}))
	mock.ExpectQuery("SELECT owner_username").WillReturnRows(sqlmock.NewRows([]string{"owner_username", "output_url"}))
	recorder := &recordingEnqueuer{}
	scanner := Scanner{DB: db, Repository: recorder, Local: files, Prefix: "production-media", RetentionDays: func(string) int { return 7 }}
	summary, err := scanner.RunOnce(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if summary.Enqueued != 1 || summary.Skipped != 1 || len(recorder.items) != 1 {
		t.Fatalf("summary=%+v items=%+v", summary, recorder.items)
	}
	if recorder.items[0].MediaKind != KindComicImage || recorder.items[0].ObjectKey != "production-media/comic/images/"+artifact.StorageRef {
		t.Fatalf("item=%+v", recorder.items[0])
	}
	if restorer.calls != 0 {
		t.Fatalf("scanner rehydrated %d cleaned objects", restorer.calls)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func pngFixture() *bytes.Reader {
	return bytes.NewReader([]byte("\x89PNG\r\n\x1a\n12345678"))
}
