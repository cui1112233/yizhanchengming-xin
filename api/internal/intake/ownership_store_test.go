package intake

import (
	"context"
	"database/sql/driver"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMySQLStoreCreatesIntakeAndOwnershipInOneTransaction(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO intakes (name, status) VALUES (?, ?)")).
		WithArgs("安全批次", StatusPending).
		WillReturnResult(sqlmock.NewResult(41, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO auth_intake_ownership (intake_id, owner_user_id, team_id) VALUES (?, ?, ?)")).
		WithArgs(int64(41), int64(7), int64(3)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	created, err := NewMySQLStore(db).CreateOwnedIntake(context.Background(), "安全批次", ActorScope{UserID: 7, TeamID: 3})
	if err != nil {
		t.Fatalf("CreateOwnedIntake: %v", err)
	}
	if created.ID != 41 || created.Name != "安全批次" || created.Status != StatusPending {
		t.Fatalf("created = %+v", created)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreRollsBackIntakeWhenOwnershipInsertFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO intakes (name, status) VALUES (?, ?)")).
		WithArgs("不可孤立", StatusPending).
		WillReturnResult(sqlmock.NewResult(42, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO auth_intake_ownership (intake_id, owner_user_id, team_id) VALUES (?, ?, ?)")).
		WithArgs(int64(42), int64(7), nil).
		WillReturnError(errors.New("ownership unavailable"))
	mock.ExpectRollback()

	if _, err := NewMySQLStore(db).CreateOwnedIntake(context.Background(), "不可孤立", ActorScope{UserID: 7}); err == nil {
		t.Fatal("CreateOwnedIntake succeeded without ownership")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreIntakeAccessMatrix(t *testing.T) {
	tests := []struct {
		name     string
		userID   int64
		teamID   int64
		elevated bool
		query    string
		args     []driver.Value
		allowed  bool
	}{
		{name: "owner", userID: 7, teamID: 0, query: "SELECT EXISTS(SELECT 1 FROM auth_intake_ownership WHERE intake_id = ? AND (owner_user_id = ? OR (? > 0 AND team_id IS NOT NULL AND team_id = ?)))", args: []driver.Value{int64(41), int64(7), int64(0), int64(0)}, allowed: true},
		{name: "same team", userID: 8, teamID: 3, query: "SELECT EXISTS(SELECT 1 FROM auth_intake_ownership WHERE intake_id = ? AND (owner_user_id = ? OR (? > 0 AND team_id IS NOT NULL AND team_id = ?)))", args: []driver.Value{int64(41), int64(8), int64(3), int64(3)}, allowed: true},
		{name: "foreign", userID: 8, teamID: 9, query: "SELECT EXISTS(SELECT 1 FROM auth_intake_ownership WHERE intake_id = ? AND (owner_user_id = ? OR (? > 0 AND team_id IS NOT NULL AND team_id = ?)))", args: []driver.Value{int64(41), int64(8), int64(9), int64(9)}, allowed: false},
		{name: "zero team never matches", userID: 8, teamID: 0, query: "SELECT EXISTS(SELECT 1 FROM auth_intake_ownership WHERE intake_id = ? AND (owner_user_id = ? OR (? > 0 AND team_id IS NOT NULL AND team_id = ?)))", args: []driver.Value{int64(41), int64(8), int64(0), int64(0)}, allowed: false},
		{name: "elevated sees legacy unowned", userID: 1, teamID: 0, elevated: true, query: "SELECT EXISTS(SELECT 1 FROM intakes WHERE id = ?)", args: []driver.Value{int64(41)}, allowed: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			mock.ExpectQuery(regexp.QuoteMeta(tc.query)).WithArgs(tc.args...).
				WillReturnRows(sqlmock.NewRows([]string{"allowed"}).AddRow(tc.allowed))
			allowed, err := NewMySQLStore(db).CanAccessIntake(context.Background(), 41, tc.userID, tc.teamID, tc.elevated)
			if err != nil {
				t.Fatal(err)
			}
			if allowed != tc.allowed {
				t.Fatalf("allowed = %v, want %v", allowed, tc.allowed)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestMySQLStoreListsVisibleIntakesInSQL(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	now := time.Date(2026, 10, 9, 0, 0, 0, 0, time.UTC)
	query := "SELECT i.id, i.name, i.status, i.created_at, i.updated_at FROM intakes i JOIN auth_intake_ownership o ON o.intake_id = i.id WHERE o.owner_user_id = ? OR (? > 0 AND o.team_id IS NOT NULL AND o.team_id = ?) ORDER BY i.id DESC LIMIT 100"
	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(int64(8), int64(3), int64(3)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "status", "created_at", "updated_at"}).AddRow(41, "团队批次", StatusCompleted, now, now))

	rows, err := NewMySQLStore(db).ListVisibleIntakes(context.Background(), 8, 3, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].ID != 41 {
		t.Fatalf("rows = %+v", rows)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreElevatedListIncludesLegacyUnownedIntakes(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	query := "SELECT id, name, status, created_at, updated_at FROM intakes ORDER BY id DESC LIMIT 100"
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "status", "created_at", "updated_at"}))
	if _, err := NewMySQLStore(db).ListVisibleIntakes(context.Background(), 1, 0, true); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
