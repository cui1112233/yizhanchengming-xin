package publishing

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMySQLAccountVisibilityIsScopedToOwnerOrTeam(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil { t.Fatal(err) }
	defer db.Close()
	store := NewMySQLStore(db)
	now := time.Date(2026, 10, 5, 13, 30, 0, 0, time.UTC)

	query := `SELECT id, owner_user_id, COALESCE(team_id,0), platform, display_name, credential_ref, active, created_at, updated_at FROM publishing_accounts WHERE owner_user_id = ? OR (team_id IS NOT NULL AND team_id = ?) ORDER BY id DESC`
	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(int64(7), int64(3)).
		WillReturnRows(sqlmock.NewRows([]string{"id","owner_user_id","team_id","platform","display_name","credential_ref","active","created_at","updated_at"}).
			AddRow(11, 7, 0, "douyin", "own", "cred-own", true, now, now).
			AddRow(12, 99, 3, "douyin", "team", "cred-team", true, now, now))

	accounts, err := store.ListAccountsVisible(context.Background(), 7, 3, false)
	if err != nil { t.Fatal(err) }
	if len(accounts) != 2 { t.Fatalf("accounts=%d want 2", len(accounts)) }
	if err := mock.ExpectationsWereMet(); err != nil { t.Fatal(err) }
}

func TestMySQLAuditVisibilityCannotCrossOwnerOrTeamBoundary(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil { t.Fatal(err) }
	defer db.Close()
	store := NewMySQLStore(db)
	now := time.Date(2026, 10, 5, 13, 35, 0, 0, time.UTC)

	query := `SELECT a.id, a.intent_id, a.batch_project_id, a.publishing_account_id, a.actor_user_id, a.platform, a.action, a.result, a.error_summary, a.created_at FROM publish_audits a JOIN publishing_accounts p ON p.id = a.publishing_account_id WHERE 1=1 AND a.batch_project_id = ? AND (p.owner_user_id = ? OR (p.team_id IS NOT NULL AND p.team_id = ?)) ORDER BY a.id DESC`
	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(int64(21), int64(7), int64(3)).
		WillReturnRows(sqlmock.NewRows([]string{"id","intent_id","batch_project_id","publishing_account_id","actor_user_id","platform","action","result","error_summary","created_at"}).
			AddRow(91, 73, 21, 11, 7, "douyin", "intent.created", "accepted", "", now))

	audits, err := store.ListAuditsVisible(context.Background(), 7, 3, 21, false)
	if err != nil { t.Fatal(err) }
	if len(audits) != 1 || audits[0].AccountID != 11 { t.Fatalf("audits=%#v", audits) }
	if err := mock.ExpectationsWereMet(); err != nil { t.Fatal(err) }
}
