package generation

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/observability"
)

func TestCreateAdminPromptDraftAuditsInSameTransaction(t *testing.T) {
	db, mock := newSQLMock(t)
	ctx := observability.WithRequestID(context.Background(), "req-draft-create")
	content := "draft-body"
	h := sha256.Sum256([]byte(content))
	sha := hex.EncodeToString(h[:])
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT version FROM generation_prompts WHERE prompt_key=? FOR UPDATE")).WithArgs(PromptScript).WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow(1).AddRow(3))
	mock.ExpectExec("INSERT INTO generation_prompts").WithArgs(PromptScript, 4, content, sha).WillReturnResult(sqlmock.NewResult(22, 1))
	mock.ExpectExec("INSERT INTO admin_audit_logs").WithArgs(int64(7), "admin.prompt.edit", "draft.create", "generation_prompt", PromptScript, int64(22), "req-draft-create", "success", `{"version":4}`, sha).WillReturnResult(sqlmock.NewResult(23, 1))
	mock.ExpectCommit()

	if err := NewMySQLStore(db).CreateAdminPromptDraft(ctx, 7, PromptScript, content); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateAdminPromptDraftRejectsPublishedVersion(t *testing.T) {
	db, mock := newSQLMock(t)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id,lifecycle FROM generation_prompts WHERE prompt_key=? AND version=? FOR UPDATE")).WithArgs(PromptScript, 3).WillReturnRows(sqlmock.NewRows([]string{"id", "lifecycle"}).AddRow(11, "published"))
	mock.ExpectRollback()

	if err := NewMySQLStore(db).UpdateAdminPromptDraft(context.Background(), 7, PromptScript, 3, "new-body"); err != ErrConflict {
		t.Fatalf("err=%v want=%v", err, ErrConflict)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateAdminPromptDraftRollsBackWhenAuditFails(t *testing.T) {
	db, mock := newSQLMock(t)
	ctx := observability.WithRequestID(context.Background(), "req-draft-update")
	content := "updated-body"
	h := sha256.Sum256([]byte(content))
	sha := hex.EncodeToString(h[:])
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id,lifecycle FROM generation_prompts WHERE prompt_key=? AND version=? FOR UPDATE")).WithArgs(PromptScript, 4).WillReturnRows(sqlmock.NewRows([]string{"id", "lifecycle"}).AddRow(22, "draft"))
	mock.ExpectExec("UPDATE generation_prompts SET content=").WithArgs(content, sha, PromptScript, 4).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO admin_audit_logs").WillReturnError(context.DeadlineExceeded)
	mock.ExpectRollback()

	if err := NewMySQLStore(db).UpdateAdminPromptDraft(ctx, 7, PromptScript, 4, content); err == nil {
		t.Fatal("expected audit failure")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetAdminPromptReturnsPersistedLifecycleMetadata(t *testing.T) {
	db, mock := newSQLMock(t)
	now := time.Date(2026, 10, 8, 8, 0, 0, 0, time.UTC)
	query := regexp.QuoteMeta(adminPromptSelect + " WHERE prompt_key=? AND version=?")
	mock.ExpectQuery(query).WithArgs(PromptScript, 4).WillReturnRows(sqlmock.NewRows([]string{"id", "prompt_key", "version", "content", "enabled", "lifecycle", "seed_source", "content_sha256", "published_at", "created_at", "updated_at"}).AddRow(22, PromptScript, 4, "draft-body", false, "draft", "admin", "sha", nil, now, now))
	prompt, err := NewMySQLStore(db).GetAdminPrompt(context.Background(), PromptScript, 4)
	if err != nil {
		t.Fatal(err)
	}
	if prompt.Lifecycle != "draft" || prompt.SeedSource != "admin" || prompt.Content != "draft-body" || prompt.Enabled {
		t.Fatalf("prompt=%#v", prompt)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
