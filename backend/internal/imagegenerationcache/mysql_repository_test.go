package imagegenerationcache

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMySQLRepositoryClaimCreatesOneGenerationLease(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := NewMySQLRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT IGNORE INTO batch_factory_v12_image_generation_cache").
		WithArgs("alice", "fingerprint", "gpt-image-2", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(7, 1))
	mock.ExpectCommit()

	entry, outcome, err := repo.Claim(context.Background(), "alice", "fingerprint", "gpt-image-2", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if outcome != ClaimAcquired || entry.ID != 7 || entry.State != StateGenerating {
		t.Fatalf("claim=(%+v,%s), want acquired id=7", entry, outcome)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLRepositoryClaimReusesReadyResult(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := NewMySQLRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT IGNORE INTO batch_factory_v12_image_generation_cache").
		WithArgs("alice", "fingerprint", "gpt-image-2", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT id,state").WithArgs("alice", "fingerprint").
		WillReturnRows(sqlmock.NewRows([]string{"id", "state", "storage_ref", "media_type", "lease_until", "queue_ms", "header_ms", "body_ms", "download_ms", "save_ms", "total_ms"}).
			AddRow(9, StateReady, "lea_result.png", "image/png", nil, 10, 20, 30, 40, 50, 150))
	mock.ExpectCommit()

	entry, outcome, err := repo.Claim(context.Background(), "alice", "fingerprint", "gpt-image-2", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if outcome != ClaimReady || entry.StorageRef != "lea_result.png" || entry.Timing.Total != 150*time.Millisecond {
		t.Fatalf("claim=(%+v,%s), want ready cached result", entry, outcome)
	}
}

func TestMySQLRepositoryMarkReadyPersistsTimings(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := NewMySQLRepository(db)
	timing := Timing{Queue: 10 * time.Millisecond, Header: 20 * time.Millisecond, Body: 30 * time.Millisecond, Download: 40 * time.Millisecond, Save: 50 * time.Millisecond, Total: 150 * time.Millisecond}
	mock.ExpectExec("UPDATE batch_factory_v12_image_generation_cache SET state='ready'").
		WithArgs("lea_result.png", "image/png", int64(10), int64(20), int64(30), int64(40), int64(50), int64(150), "alice", "fingerprint").
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := repo.MarkReady(context.Background(), "alice", "fingerprint", "lea_result.png", "image/png", timing); err != nil {
		t.Fatal(err)
	}
}
