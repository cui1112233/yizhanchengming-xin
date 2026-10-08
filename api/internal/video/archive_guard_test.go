package video

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestProductionAndMergeCreationLockAndRejectArchivedProject(t *testing.T) {
	for _, tc := range []struct {
		name string
		call func(*MySQLStore) error
	}{
		{name: "production", call: func(store *MySQLStore) error {
			_, _, err := store.CreateOrGetProductionJob(context.Background(), ProductionJob{BatchProjectID: 51, BookID: 21})
			return err
		}},
		{name: "merge", call: func(store *MySQLStore) error {
			_, err := store.CreateMergeJob(context.Background(), MergeJob{BatchProjectID: 51, BookID: 21})
			return err
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			archivedAt := time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)
			mock.ExpectBegin()
			mock.ExpectQuery(regexp.QuoteMeta("SELECT archived_at FROM batch_projects WHERE id=? FOR UPDATE")).WithArgs(int64(51)).WillReturnRows(sqlmock.NewRows([]string{"archived_at"}).AddRow(archivedAt))
			mock.ExpectRollback()
			if err := tc.call(NewMySQLStore(db)); !errors.Is(err, ErrProjectArchived) {
				t.Fatalf("err=%v, want ErrProjectArchived", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestRetryAttemptCreationLocksAndRejectsArchivedProject(t *testing.T) {
	for _, tc := range []struct {
		name, query string
		call        func(*MySQLStore) error
	}{
		{name: "production task", query: "SELECT p.archived_at FROM video_production_jobs j JOIN batch_projects p ON p.id=j.batch_project_id WHERE j.id=? FOR UPDATE", call: func(store *MySQLStore) error {
			_, err := store.CreateProductionTask(context.Background(), ProductionTask{ProductionJobID: 71, Attempt: 2})
			return err
		}},
		{name: "merge attempt", query: "SELECT p.archived_at FROM video_merge_jobs j JOIN batch_projects p ON p.id=j.batch_project_id WHERE j.id=? FOR UPDATE", call: func(store *MySQLStore) error {
			_, err := store.CreateMergeAttempt(context.Background(), MergeAttempt{MergeJobID: 71, Attempt: 2})
			return err
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			archivedAt := time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)
			mock.ExpectBegin()
			mock.ExpectQuery(regexp.QuoteMeta(tc.query)).WithArgs(int64(71)).WillReturnRows(sqlmock.NewRows([]string{"archived_at"}).AddRow(archivedAt))
			mock.ExpectRollback()
			if err := tc.call(NewMySQLStore(db)); !errors.Is(err, ErrProjectArchived) {
				t.Fatalf("err=%v, want ErrProjectArchived", err)
			}
		})
	}
}
