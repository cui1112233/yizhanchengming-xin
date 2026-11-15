package batchfactoryv11

import (
	"context"
	"database/sql/driver"
	"github.com/DATA-DOG/go-sqlmock"
	"os"
	"path/filepath"
	"qiantie/backend/internal/localartifact"
	"testing"
	"time"
)

type retentionTimeArg func(driver.Value) bool

func (f retentionTimeArg) Match(v driver.Value) bool { return f(v) }

func TestRetentionWholeBatchGetsHiddenWithFullGrace(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	old := time.Date(2026, 10, 10, 13, 2, 17, 0, time.UTC)
	mock.ExpectQuery("SELECT b.id").WithArgs("alice").WillReturnRows(sqlmock.NewRows([]string{"id", "completed_at", "existing"}).AddRow("batch-old", old, nil))
	mock.ExpectBegin()
	// 旧版本按书隐藏的数据先恢复，统一改由作品集计时。
	mock.ExpectExec("UPDATE batch_factory_v12_books SET hidden_at=NULL").WithArgs("alice").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("UPDATE batch_factory_v12_batches SET completed_at").WithArgs(old, old.AddDate(0, 0, 3), "batch-old", "alice").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE batch_factory_v12_batches SET hidden_at=NULL").WithArgs("alice", sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 0))
	var hidden time.Time
	nowArg := retentionTimeArg(func(v driver.Value) bool { var ok bool; hidden, ok = v.(time.Time); return ok })
	graceArg := retentionTimeArg(func(v driver.Value) bool { t, ok := v.(time.Time); return ok && t.Sub(hidden) == 24*time.Hour })
	mock.ExpectExec("UPDATE batch_factory_v12_batches SET hidden_at=\\?,purge_after=\\?").WithArgs(nowArg, graceArg, "alice", sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	result, err := NewMySQLStore(db).RecomputeBatchRetention(context.Background(), "alice", 3)
	if err != nil || result.Hidden != 1 {
		t.Fatalf("hide old batch: %+v %v", result, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRetentionSweepRunsAtMidnightOnly(t *testing.T) {
	day := func(hour, minute int) time.Time {
		return time.Date(2026, 11, 16, hour, minute, 0, 0, time.Local)
	}
	cases := []struct{ now, want time.Time }{
		{day(10, 0), time.Date(2026, 11, 17, 0, 0, 0, 0, time.Local)},
		{day(0, 0), time.Date(2026, 11, 17, 0, 0, 0, 0, time.Local)},
		{time.Date(2026, 11, 15, 23, 59, 0, 0, time.Local), time.Date(2026, 11, 16, 0, 0, 0, 0, time.Local)},
	}
	for _, tc := range cases {
		if got := nextRetentionSweep(0, 0, tc.now); !got.Equal(tc.want) {
			t.Fatalf("nextRetentionSweep(%s)=%s want=%s", tc.now, got, tc.want)
		}
	}
}

func TestRetentionMissingTOSKeepsLocalFileAndLedger(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	root := t.TempDir()
	file := filepath.Join(root, "keep.mp4")
	if err := os.WriteFile(file, []byte("saved media"), 0600); err != nil {
		t.Fatal(err)
	}
	mock.ExpectQuery("SELECT object_key").WithArgs("alice", "keep.mp4").WillReturnRows(sqlmock.NewRows([]string{"object_key"}).AddRow("production-media/keep.mp4"))
	err = NewMySQLStore(db).purgeStorageRef(context.Background(), "alice", "keep.mp4", localartifact.NewStore(root, 1<<20), nil)
	if err == nil {
		t.Fatal("missing TOS must stop deletion")
	}
	if _, err := os.Stat(file); err != nil {
		t.Fatal("local file deleted before cloud deletion was possible")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
