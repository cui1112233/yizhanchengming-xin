package generation

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/observability"
)

// AdminPrompt is the persisted generation_prompts version projected for the
// Admin API. Content is populated only for a capability-protected detail read;
// HTTP list handlers must use the metadata projection.
type AdminPrompt struct {
	ID            int64      `json:"id"`
	Key           string     `json:"key"`
	Version       int        `json:"version"`
	Content       string     `json:"content,omitempty"`
	Enabled       bool       `json:"enabled"`
	Lifecycle     string     `json:"lifecycle"`
	SeedSource    string     `json:"seedSource"`
	ContentSHA256 string     `json:"contentSha256"`
	PublishedAt   *time.Time `json:"publishedAt,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

const adminPromptSelect = `SELECT id,prompt_key,version,content,enabled,lifecycle,seed_source,content_sha256,published_at,created_at,updated_at FROM generation_prompts`

type adminPromptScanner interface {
	Scan(...any) error
}

func scanAdminPrompt(scanner adminPromptScanner) (AdminPrompt, error) {
	var prompt AdminPrompt
	var publishedAt sql.NullTime
	if err := scanner.Scan(&prompt.ID, &prompt.Key, &prompt.Version, &prompt.Content, &prompt.Enabled, &prompt.Lifecycle, &prompt.SeedSource, &prompt.ContentSHA256, &publishedAt, &prompt.CreatedAt, &prompt.UpdatedAt); err != nil {
		return AdminPrompt{}, err
	}
	if publishedAt.Valid {
		value := publishedAt.Time
		prompt.PublishedAt = &value
	}
	return prompt, nil
}

func (s *MySQLStore) ListAdminPrompts(ctx context.Context, key string) ([]AdminPrompt, error) {
	query := adminPromptSelect + ` ORDER BY prompt_key,version DESC`
	args := []any{}
	if key = strings.TrimSpace(key); key != "" {
		query = adminPromptSelect + ` WHERE prompt_key=? ORDER BY version DESC`
		args = append(args, key)
	}
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]AdminPrompt, 0)
	for rows.Next() {
		prompt, err := scanAdminPrompt(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, prompt)
	}
	return out, rows.Err()
}

func (s *MySQLStore) GetAdminPrompt(ctx context.Context, key string, version int) (AdminPrompt, error) {
	if strings.TrimSpace(key) == "" || version <= 0 {
		return AdminPrompt{}, ErrInvalid
	}
	row := s.db.QueryRowContext(ctx, adminPromptSelect+` WHERE prompt_key=? AND version=?`, key, version)
	prompt, err := scanAdminPrompt(row)
	if err != nil {
		return AdminPrompt{}, noRows(err)
	}
	return prompt, nil
}

func (s *MySQLStore) CreateAdminPromptDraft(ctx context.Context, actor int64, key, content string) error {
	key = strings.TrimSpace(key)
	if actor <= 0 || key == "" || strings.TrimSpace(content) == "" {
		return ErrInvalid
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	rows, err := tx.QueryContext(ctx, "SELECT version FROM generation_prompts WHERE prompt_key=? FOR UPDATE", key)
	if err != nil {
		return err
	}
	maxVersion := 0
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			rows.Close()
			return err
		}
		if version > maxVersion {
			maxVersion = version
		}
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}
	version := maxVersion + 1
	h := sha256.Sum256([]byte(content))
	contentSHA := hex.EncodeToString(h[:])
	result, err := tx.ExecContext(ctx, "INSERT INTO generation_prompts(prompt_key,version,content,enabled,lifecycle,seed_source,content_sha256) VALUES(?,?,?,0,'draft','admin',?)", key, version, content, contentSHA)
	if err != nil {
		return err
	}
	promptID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	if err := insertAdminPromptAudit(ctx, tx, actor, "admin.prompt.edit", "draft.create", key, promptID, contentSHA, fmt.Sprintf(`{"version":%d}`, version)); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *MySQLStore) UpdateAdminPromptDraft(ctx context.Context, actor int64, key string, version int, content string) error {
	key = strings.TrimSpace(key)
	if actor <= 0 || key == "" || version <= 0 || strings.TrimSpace(content) == "" {
		return ErrInvalid
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var promptID int64
	var lifecycle string
	if err := tx.QueryRowContext(ctx, "SELECT id,lifecycle FROM generation_prompts WHERE prompt_key=? AND version=? FOR UPDATE", key, version).Scan(&promptID, &lifecycle); err != nil {
		return noRows(err)
	}
	if lifecycle != "draft" {
		return ErrConflict
	}
	h := sha256.Sum256([]byte(content))
	contentSHA := hex.EncodeToString(h[:])
	result, err := tx.ExecContext(ctx, "UPDATE generation_prompts SET content=?,content_sha256=?,enabled=0 WHERE prompt_key=? AND version=? AND lifecycle='draft'", content, contentSHA, key, version)
	if err != nil {
		return err
	}
	if affected, err := result.RowsAffected(); err != nil {
		return err
	} else if affected != 1 {
		return ErrConflict
	}
	if err := insertAdminPromptAudit(ctx, tx, actor, "admin.prompt.edit", "draft.update", key, promptID, contentSHA, fmt.Sprintf(`{"version":%d}`, version)); err != nil {
		return err
	}
	return tx.Commit()
}

func insertAdminPromptAudit(ctx context.Context, tx *sql.Tx, actor int64, capability, action, resourceID string, promptVersionID int64, contentSHA, summary string) error {
	_, err := tx.ExecContext(ctx, "INSERT INTO admin_audit_logs(actor_user_id,capability,action,resource_type,resource_id,prompt_version_id,request_id,result,summary_json,content_sha256) VALUES(?,?,?,?,?,?,?,?,?,?)", actor, capability, action, "generation_prompt", resourceID, promptVersionID, observability.RequestID(ctx), "success", summary, contentSHA)
	return err
}
