package video

import (
	"context"
	"crypto/sha256"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMySQLStoreListLocalExecutorsForOwnerFiltersInSQL(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	now := time.Date(2026, 10, 8, 9, 0, 0, 0, time.UTC)
	tokenHash := sha256.Sum256([]byte("owner-only-token"))
	rows := sqlmock.NewRows([]string{"id", "owner_user_id", "name", "provider_key", "model", "capabilities_json", "token_hash", "last_seen_at", "created_at", "updated_at"}).
		AddRow("lex_owner", int64(11), "owner mac", ProviderDoubaoLocalExecutor, ModelDoubaoSeedance, `[]`, tokenHash[:], now, now, now)
	mock.ExpectQuery(regexp.QuoteMeta(localExecutorSelect + ` WHERE owner_user_id=? ORDER BY created_at ASC, id ASC`)).
		WithArgs(int64(11)).
		WillReturnRows(rows)

	got, err := NewMySQLStore(db).ListLocalExecutorsForOwner(context.Background(), 11)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != "lex_owner" || got[0].OwnerUserID != 11 {
		t.Fatalf("owner-local executor result = %#v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMySQLStoreListLocalExecutorsForOwnerRejectsMissingOwner(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := NewMySQLStore(db).ListLocalExecutorsForOwner(context.Background(), 0); err != ErrLocalExecutorUnauthorized {
		t.Fatalf("error = %v, want unauthorized", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
