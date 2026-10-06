package sharedplatform

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMySQLIntakeAccessStoreClaimsAndChecksOwner(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewMySQLIntakeAccessStore(db)

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO auth_intake_ownership (intake_id, owner_user_id, team_id) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE intake_id = VALUES(intake_id)")).
		WithArgs(int64(51), int64(7), int64(3)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := store.ClaimIntake(context.Background(), 51, 7, 3); err != nil {
		t.Fatal(err)
	}

	mock.ExpectQuery("SELECT EXISTS").
		WithArgs(int64(51), int64(7), int64(3)).
		WillReturnRows(sqlmock.NewRows([]string{"allowed"}).AddRow(true))
	allowed, err := store.CanAccessIntake(context.Background(), 51, 7, 3, false)
	if err != nil || !allowed {
		t.Fatalf("allowed=%v err=%v", allowed, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
