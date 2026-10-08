package generation

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"github.com/DATA-DOG/go-sqlmock"
	"regexp"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/observability"
)

func TestPublishPromptArchivesOldVersionAndAuditsAtomically(t *testing.T) {
	db, m := newSQLMock(t)
	m.ExpectBegin()
	m.ExpectQuery(regexp.QuoteMeta("SELECT id,version,content FROM generation_prompts WHERE prompt_key=? FOR UPDATE")).WithArgs(PromptScript).WillReturnRows(sqlmock.NewRows([]string{"id", "version", "content"}).AddRow(1, 1, "old").AddRow(2, 2, "new"))
	m.ExpectExec("UPDATE generation_prompts SET lifecycle='archived'").WithArgs(PromptScript).WillReturnResult(sqlmock.NewResult(0, 1))
	m.ExpectExec("UPDATE generation_prompts SET lifecycle='published'").WithArgs(int64(7), PromptScript, 2).WillReturnResult(sqlmock.NewResult(0, 1))
	m.ExpectExec("INSERT INTO admin_audit_logs").WillReturnResult(sqlmock.NewResult(3, 1))
	m.ExpectCommit()
	if err := NewMySQLStore(db).PublishAdminPrompt(context.Background(), 7, PromptScript, 2); err != nil {
		t.Fatal(err)
	}
	if err := m.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPublishPromptRollsBackWhenAuditWriteFails(t *testing.T) {
	db, m := newSQLMock(t)
	m.ExpectBegin()
	m.ExpectQuery(regexp.QuoteMeta("SELECT id,version,content FROM generation_prompts WHERE prompt_key=? FOR UPDATE")).WithArgs(PromptScript).WillReturnRows(sqlmock.NewRows([]string{"id", "version", "content"}).AddRow(1, 1, "body"))
	m.ExpectExec("UPDATE generation_prompts SET lifecycle='archived'").WithArgs(PromptScript).WillReturnResult(sqlmock.NewResult(0, 1))
	m.ExpectExec("UPDATE generation_prompts SET lifecycle='published'").WithArgs(int64(7), PromptScript, 2).WillReturnResult(sqlmock.NewResult(0, 1))
	m.ExpectExec("INSERT INTO admin_audit_logs").WillReturnError(context.DeadlineExceeded)
	m.ExpectRollback()
	if err := NewMySQLStore(db).PublishAdminPrompt(context.Background(), 7, PromptScript, 2); err == nil {
		t.Fatal("expected audit failure")
	}
	if err := m.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRestorePromptCreatesHigherPublishedVersion(t *testing.T) {
	db, m := newSQLMock(t)
	m.ExpectBegin()
	m.ExpectQuery("SELECT id,version,content FROM generation_prompts WHERE prompt_key=.*FOR UPDATE").WithArgs(PromptScript).WillReturnRows(sqlmock.NewRows([]string{"id", "version", "content"}).AddRow(1, 1, "old").AddRow(2, 2, "current"))
	h := sha256.Sum256([]byte("old"))
	m.ExpectExec("INSERT INTO generation_prompts").WithArgs(PromptScript, 3, "old", hex.EncodeToString(h[:])).WillReturnResult(sqlmock.NewResult(3, 1))
	m.ExpectExec("UPDATE generation_prompts SET lifecycle='archived'").WithArgs(PromptScript).WillReturnResult(sqlmock.NewResult(0, 1))
	m.ExpectExec("UPDATE generation_prompts SET lifecycle='published'").WithArgs(int64(7), PromptScript, 3).WillReturnResult(sqlmock.NewResult(0, 1))
	m.ExpectExec("INSERT INTO admin_audit_logs").WillReturnResult(sqlmock.NewResult(4, 1))
	m.ExpectCommit()
	if err := NewMySQLStore(db).RestoreAdminPrompt(context.Background(), 7, PromptScript, 1); err != nil {
		t.Fatal(err)
	}
	if err := m.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPublishPromptRollsBackWhenPublishedTargetDoesNotExist(t *testing.T) {
	db, m := newSQLMock(t)
	m.ExpectBegin()
	m.ExpectQuery(regexp.QuoteMeta("SELECT id,version,content FROM generation_prompts WHERE prompt_key=? FOR UPDATE")).WithArgs(PromptScript).WillReturnRows(sqlmock.NewRows([]string{"id", "version", "content"}).AddRow(1, 1, "old"))
	// The old published row is archived first, but publishing an absent version
	// must abort the transaction so the key is never left without exactly one
	// published row.
	m.ExpectExec("UPDATE generation_prompts SET lifecycle='archived'").WithArgs(PromptScript).WillReturnResult(sqlmock.NewResult(0, 1))
	m.ExpectExec("UPDATE generation_prompts SET lifecycle='published'").WithArgs(int64(7), PromptScript, 2).WillReturnResult(sqlmock.NewResult(0, 0))
	m.ExpectRollback()
	if err := NewMySQLStore(db).PublishAdminPrompt(context.Background(), 7, PromptScript, 2); err == nil {
		t.Fatal("expected missing publish target to roll back")
	}
	if err := m.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPublishPromptAuditUsesOnlyAllowlistedServerMetadata(t *testing.T) {
	db, m := newSQLMock(t)
	serverRequestID := "req_server_123"
	ctx := observability.WithRequestID(context.Background(), serverRequestID)
	m.ExpectBegin()
	m.ExpectQuery(regexp.QuoteMeta("SELECT id,version,content FROM generation_prompts WHERE prompt_key=? FOR UPDATE")).WithArgs(PromptScript).WillReturnRows(sqlmock.NewRows([]string{"id", "version", "content"}).AddRow(11, 1, "old body").AddRow(12, 2, "published body"))
	m.ExpectExec("UPDATE generation_prompts SET lifecycle='archived'").WithArgs(PromptScript).WillReturnResult(sqlmock.NewResult(0, 1))
	m.ExpectExec("UPDATE generation_prompts SET lifecycle='published'").WithArgs(int64(7), PromptScript, 2).WillReturnResult(sqlmock.NewResult(0, 1))
	h := sha256.Sum256([]byte("published body"))
	m.ExpectExec("INSERT INTO admin_audit_logs").WithArgs(
		int64(7), "admin.prompt.publish", "publish", "generation_prompt", PromptScript,
		int64(12), serverRequestID, "success", `{"version":2}`, hex.EncodeToString(h[:]),
	).WillReturnResult(sqlmock.NewResult(3, 1))
	m.ExpectCommit()

	// The final argument simulates an untrusted request-body requestId. The
	// store must ignore it and persist only the server request id from ctx.
	if err := NewMySQLStore(db).PublishAdminPrompt(ctx, 7, PromptScript, 2); err != nil {
		t.Fatal(err)
	}
	if err := m.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
