package publishing

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func TestMySQLTransactionsRollbackAndRetry(t *testing.T) {
	dsn := os.Getenv("TASK15_MYSQL_DSN")
	if dsn == "" {
		t.Skip("TASK15_MYSQL_DSN not configured")
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	now := time.Date(2026, 10, 5, 13, 0, 0, 0, time.UTC)

	userResult, err := db.ExecContext(ctx, `INSERT INTO auth_users (username, display_name, password_hash, role, active) VALUES (?, ?, ?, 'member', TRUE)`, "task15-atomic-user", "Task15 Atomic User", "integration-only-hash")
	if err != nil {
		t.Fatal(err)
	}
	userID, err := userResult.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	defer db.ExecContext(ctx, `DELETE FROM auth_users WHERE id = ?`, userID)

	store := NewMySQLStore(db)
	credential := EncryptedCredential{
		Ref: CredentialRef{ID: "cred_task15_atomic", Platform: "douyin", Name: "atomic", CreatedAt: now, UpdatedAt: now},
		OwnerUserID: userID,
		KeyID: "integration-key-id",
		Nonce: []byte("123456789012"),
		Ciphertext: []byte("integration-ciphertext-only"),
	}
	account := Account{
		OwnerUserID: userID,
		Platform: "douyin",
		DisplayName: strings.Repeat("x", 192),
		CredentialRefID: credential.Ref.ID,
		Active: true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if _, err := store.CreateAccountWithCredential(ctx, account, credential); err == nil {
		t.Fatal("expected oversized account display name to fail after credential insert")
	}
	assertCount(t, db, `SELECT COUNT(*) FROM publishing_credentials WHERE ref = ?`, 0, credential.Ref.ID)
	assertCount(t, db, `SELECT COUNT(*) FROM publishing_accounts WHERE credential_ref = ?`, 0, credential.Ref.ID)

	account.DisplayName = "Task15 Atomic Account"
	createdAccount, err := store.CreateAccountWithCredential(ctx, account, credential)
	if err != nil {
		t.Fatalf("retry account create after rollback: %v", err)
	}
	defer db.ExecContext(ctx, `DELETE FROM publishing_credentials WHERE ref = ?`, credential.Ref.ID)
	defer db.ExecContext(ctx, `DELETE FROM publishing_accounts WHERE id = ?`, createdAccount.ID)
	assertCount(t, db, `SELECT COUNT(*) FROM publishing_credentials WHERE ref = ?`, 1, credential.Ref.ID)
	assertCount(t, db, `SELECT COUNT(*) FROM publishing_accounts WHERE id = ?`, 1, createdAccount.ID)

	intakeResult, err := db.ExecContext(ctx, `INSERT INTO intakes (name, status) VALUES (?, 'completed')`, "task15-publish-atomic")
	if err != nil {
		t.Fatal(err)
	}
	intakeID, err := intakeResult.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	defer db.ExecContext(ctx, `DELETE FROM intakes WHERE id = ?`, intakeID)

	projectResult, err := db.ExecContext(ctx, `INSERT INTO batch_projects (intake_id, name) VALUES (?, ?)`, intakeID, "task15-publish-project")
	if err != nil {
		t.Fatal(err)
	}
	projectID, err := projectResult.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}

	intent := Intent{
		BatchProjectID: projectID,
		PublishingAccountID: createdAccount.ID,
		RequestedByUserID: userID,
		Platform: "douyin",
		Status: IntentStatusPending,
		RequestedAt: now,
		UpdatedAt: now,
	}
	audit := Audit{
		BatchProjectID: projectID,
		AccountID: createdAccount.ID,
		ActorUserID: userID,
		Platform: "douyin",
		Action: strings.Repeat("a", 65),
		Result: AuditResultAccepted,
		CreatedAt: now,
	}

	if _, err := store.CreateIntentWithAudit(ctx, intent, audit); err == nil {
		t.Fatal("expected oversized audit action to fail after intent insert")
	}
	assertCount(t, db, `SELECT COUNT(*) FROM publish_intents WHERE batch_project_id = ? AND publishing_account_id = ?`, 0, projectID, createdAccount.ID)
	assertCount(t, db, `SELECT COUNT(*) FROM publish_audits WHERE batch_project_id = ? AND publishing_account_id = ?`, 0, projectID, createdAccount.ID)

	audit.Action = "intent.created"
	createdIntent, err := store.CreateIntentWithAudit(ctx, intent, audit)
	if err != nil {
		t.Fatalf("retry intent create after rollback: %v", err)
	}
	defer db.ExecContext(ctx, `DELETE FROM publish_intents WHERE id = ?`, createdIntent.ID)
	assertCount(t, db, `SELECT COUNT(*) FROM publish_intents WHERE id = ?`, 1, createdIntent.ID)
	assertCount(t, db, `SELECT COUNT(*) FROM publish_audits WHERE intent_id = ?`, 1, createdIntent.ID)
}

func assertCount(t *testing.T, db *sql.DB, query string, want int, args ...any) {
	t.Helper()
	var got int
	if err := db.QueryRow(query, args...).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("count for %q = %d, want %d", query, got, want)
	}
}
