package shuihuo

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMySQLStoreCreateLinkedMediaTaskRejectsProductionTaskOutsideScope(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectExec(`(?s)INSERT INTO shuihuo_media_tasks.*SELECT.*video_production_tasks.*video_production_jobs`).
		WithArgs(int64(7), int64(8), nil, nil, int64(31), MediaVideo, "volc", "model", "req-1", int64(31), int64(7), int64(8)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	_, err = NewMySQLStore(db).CreateMediaTask(context.Background(), CreateMediaTaskInput{
		BatchProjectID: 7, BookID: 8, ProductionTaskID: 31, Kind: MediaVideo,
		Provider: "volc", Model: "model", RequestID: "req-1",
	})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("err=%v, want ErrNotFound for cross-scope production task", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
