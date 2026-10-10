package novelfetchworkshop

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type MySQLStore struct {
	DB              *sql.DB
	LegacyConfigDir string
}

func NewMySQLStore(db *sql.DB) *MySQLStore { return &MySQLStore{DB: db} }

func (s *MySQLStore) GetDocument(ctx context.Context, owner, bookID string) (Document, error) {
	var raw []byte
	var updated time.Time
	err := s.DB.QueryRowContext(ctx, `SELECT document_json, updated_at FROM novel_fetch_workshop_documents WHERE owner_username=? AND book_id=?`, owner, bookID).Scan(&raw, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return Document{}, ErrNotFound
	}
	if err != nil {
		return Document{}, err
	}
	var document Document
	if err := json.Unmarshal(raw, &document); err != nil {
		return Document{}, fmt.Errorf("decode novel fetch document: %w", err)
	}
	document = normalizeDocument(document)
	document.BookID = bookID
	if hasLegacyDocumentBodies(document) {
		if err := s.MigrateLegacyDocumentBodies(ctx, owner, bookID); err != nil {
			return Document{}, err
		}
		return s.GetDocument(ctx, owner, bookID)
	}
	document.UpdatedAt = updated.UTC().Format(time.RFC3339Nano)
	return document, nil
}

func (s *MySQLStore) PutDocument(ctx context.Context, owner string, document Document) error {
	document = normalizeDocument(document)
	if document.BookID == "" {
		return errors.New("book id is required")
	}
	raw, err := json.Marshal(document)
	if err != nil {
		return err
	}
	_, err = s.DB.ExecContext(ctx, `INSERT INTO novel_fetch_workshop_documents(owner_username,book_id,document_json,created_at,updated_at) VALUES(?,?,?,CURRENT_TIMESTAMP(6),CURRENT_TIMESTAMP(6)) ON DUPLICATE KEY UPDATE document_json=VALUES(document_json), updated_at=CURRENT_TIMESTAMP(6)`, owner, document.BookID, string(raw))
	return err
}

func (s *MySQLStore) ListDocuments(ctx context.Context, owner string) ([]Document, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT book_id, document_json, updated_at FROM novel_fetch_workshop_documents WHERE owner_username=? ORDER BY updated_at DESC, book_id ASC`, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []Document{}
	for rows.Next() {
		var bookID string
		var raw []byte
		var updated time.Time
		if err := rows.Scan(&bookID, &raw, &updated); err != nil {
			return nil, err
		}
		var document Document
		if err := json.Unmarshal(raw, &document); err != nil {
			return nil, fmt.Errorf("decode novel fetch document %s: %w", bookID, err)
		}
		document = normalizeDocument(document)
		document.BookID = bookID
		if hasLegacyDocumentBodies(document) {
			document = lightweightLegacyDocument(document, legacyBodyRecords(document))
		}
		document.UpdatedAt = updated.UTC().Format(time.RFC3339Nano)
		result = append(result, document)
	}
	return result, rows.Err()
}

func (s *MySQLStore) DeleteDocuments(ctx context.Context, owner string, ids []string) (DeleteResult, error) {
	result := DeleteResult{Requested: len(ids), Results: make([]DeleteRecord, 0, len(ids))}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return result, err
	}
	defer tx.Rollback()
	for _, rawID := range ids {
		id := strings.TrimSpace(rawID)
		if id == "" {
			result.Results = append(result.Results, DeleteRecord{BookID: rawID, Deleted: false})
			continue
		}
		res, err := tx.ExecContext(ctx, `DELETE FROM novel_fetch_workshop_documents WHERE owner_username=? AND book_id=?`, owner, id)
		if err != nil {
			return result, err
		}
		rows, err := res.RowsAffected()
		if err != nil {
			return result, err
		}
		deleted := rows > 0
		if deleted {
			result.Deleted++
		}
		result.Results = append(result.Results, DeleteRecord{BookID: id, Deleted: deleted})
	}
	if err := tx.Commit(); err != nil {
		return result, err
	}
	return result, nil
}

func (s *MySQLStore) GetConfig(ctx context.Context, owner string) (map[string]any, error) {
	var raw []byte
	err := s.DB.QueryRowContext(ctx, `SELECT settings_json FROM novel_fetch_workshop_configs WHERE owner_username=?`, owner).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		// The old Node frontend stores the model catalog in each user's
		// api-config.json. Read it as a local compatibility fallback until the
		// user saves the settings through the Go frontend. Secrets stay local and
		// are never committed to the repository.
		settings := map[string]any{}
		if strings.TrimSpace(s.LegacyConfigDir) != "" {
			raw, readErr := os.ReadFile(filepath.Join(s.LegacyConfigDir, owner, "api-config.json"))
			if readErr == nil {
				var legacy map[string]any
				if json.Unmarshal(raw, &legacy) == nil {
					if catalog, ok := legacy["modelCatalog"]; ok {
						settings["modelCatalog"] = catalog
					}
					if version, ok := legacy["modelCatalogVersion"]; ok {
						settings["modelCatalogVersion"] = version
					}
				}
			}
		}
		return settings, nil
	}
	if err != nil {
		return nil, err
	}
	var settings map[string]any
	if err := json.Unmarshal(raw, &settings); err != nil {
		return nil, fmt.Errorf("decode novel fetch config: %w", err)
	}
	if settings == nil {
		settings = map[string]any{}
	}
	legacyCatalog := s.readLegacyModelCatalog(owner)
	if catalog, ok := settings["modelCatalog"].([]any); !ok || len(catalog) == 0 {
		if len(legacyCatalog) > 0 {
			settings["modelCatalog"] = legacyCatalog
		}
	}
	return settings, nil
}

func (s *MySQLStore) readLegacyModelCatalog(owner string) []any {
	if s == nil || strings.TrimSpace(s.LegacyConfigDir) == "" {
		return nil
	}
	raw, err := os.ReadFile(filepath.Join(s.LegacyConfigDir, owner, "api-config.json"))
	if err != nil {
		return nil
	}
	var legacy map[string]any
	if json.Unmarshal(raw, &legacy) != nil {
		return nil
	}
	catalog, _ := legacy["modelCatalog"].([]any)
	return catalog
}

func (s *MySQLStore) PutConfig(ctx context.Context, owner string, settings map[string]any) error {
	if settings == nil {
		settings = map[string]any{}
	}
	raw, err := json.Marshal(settings)
	if err != nil {
		return err
	}
	_, err = s.DB.ExecContext(ctx, `INSERT INTO novel_fetch_workshop_configs(owner_username,settings_json,created_at,updated_at) VALUES(?,?,CURRENT_TIMESTAMP(6),CURRENT_TIMESTAMP(6)) ON DUPLICATE KEY UPDATE settings_json=VALUES(settings_json), updated_at=CURRENT_TIMESTAMP(6)`, owner, string(raw))
	return err
}

func (s *MySQLStore) PutRun(ctx context.Context, owner string, record RunRecord) error {
	record = normalizeRunRecord(record)
	if record.RunID == "" || record.BookID == "" || record.Stage == "" || record.Status == "" || record.TextModelID == "" || record.ModelID == "" {
		return errors.New("run audit fields are required")
	}
	_, err := s.DB.ExecContext(ctx, `INSERT INTO novel_fetch_workshop_runs(run_id,owner_username,book_id,stage,status,attempts,text_model_id,model_id,model_display_name,request_id,error_message,started_at,finished_at,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,CURRENT_TIMESTAMP(6)) ON DUPLICATE KEY UPDATE book_id=VALUES(book_id),stage=VALUES(stage),status=VALUES(status),attempts=VALUES(attempts),text_model_id=VALUES(text_model_id),model_id=VALUES(model_id),model_display_name=VALUES(model_display_name),request_id=VALUES(request_id),error_message=VALUES(error_message),started_at=VALUES(started_at),finished_at=VALUES(finished_at),updated_at=CURRENT_TIMESTAMP(6)`, record.RunID, owner, record.BookID, record.Stage, record.Status, record.Attempts, record.TextModelID, record.ModelID, record.ModelDisplayName, nullableRunString(record.RequestID), nullableRunString(record.ErrorMessage), nullableTime(record.StartedAt), nullableTime(record.FinishedAt), time.Now().UTC())
	return err
}

func (s *MySQLStore) ListRuns(ctx context.Context, owner, bookID string) ([]RunRecord, error) {
	query := `SELECT run_id,book_id,stage,status,attempts,text_model_id,model_id,model_display_name,COALESCE(request_id,''),COALESCE(error_message,''),started_at,finished_at,updated_at FROM novel_fetch_workshop_runs WHERE owner_username=?`
	args := []any{owner}
	if bookID != "" {
		query += ` AND book_id=?`
		args = append(args, bookID)
	}
	query += ` ORDER BY updated_at DESC, run_id ASC`
	rows, err := s.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []RunRecord{}
	for rows.Next() {
		var record RunRecord
		var startedAt, finishedAt sql.NullTime
		var updatedAt time.Time
		if err := rows.Scan(&record.RunID, &record.BookID, &record.Stage, &record.Status, &record.Attempts, &record.TextModelID, &record.ModelID, &record.ModelDisplayName, &record.RequestID, &record.ErrorMessage, &startedAt, &finishedAt, &updatedAt); err != nil {
			return nil, err
		}
		if startedAt.Valid {
			record.StartedAt = startedAt.Time.UTC().Format(time.RFC3339Nano)
		}
		if finishedAt.Valid {
			record.FinishedAt = finishedAt.Time.UTC().Format(time.RFC3339Nano)
		}
		record.UpdatedAt = updatedAt.UTC().Format(time.RFC3339Nano)
		result = append(result, record)
	}
	return result, rows.Err()
}

func nullableTime(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return nil
	}
	return parsed.UTC()
}

func nullableRunString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func (s *MySQLStore) PutBody(ctx context.Context, owner string, body BodyRecord) (BodyRef, error) {
	bookID := strings.TrimSpace(body.BookID)
	versionID := strings.TrimSpace(body.VersionID)
	if bookID == "" {
		return BodyRef{}, errors.New("book id is required")
	}
	if versionID == "" {
		return BodyRef{}, errors.New("version id is required")
	}
	state := strings.TrimSpace(body.State)
	if state == "" {
		state = "ready"
	}
	blob, hash, charCount, err := encodeBody(body.Content)
	if err != nil {
		return BodyRef{}, err
	}
	_, err = s.DB.ExecContext(ctx, `INSERT INTO novel_fetch_workshop_bodies(owner_username,book_id,version_id,revision,content_encoding,content_blob,content_hash,char_count,state,created_at,updated_at,last_needed_at,releasable_at,expires_at) VALUES(?,?,?,1,'gzip',?,?,?,?,CURRENT_TIMESTAMP(6),CURRENT_TIMESTAMP(6),CURRENT_TIMESTAMP(6),NULL,NULL) ON DUPLICATE KEY UPDATE revision=revision+1, content_encoding=VALUES(content_encoding), content_blob=VALUES(content_blob), content_hash=VALUES(content_hash), char_count=VALUES(char_count), state=VALUES(state), updated_at=CURRENT_TIMESTAMP(6), last_needed_at=CURRENT_TIMESTAMP(6), releasable_at=NULL, expires_at=NULL`, owner, bookID, versionID, blob, hash, charCount, state)
	if err != nil {
		return BodyRef{}, err
	}

	ref := BodyRef{VersionID: versionID}
	var updated time.Time
	err = s.DB.QueryRowContext(ctx, `SELECT revision, content_hash, char_count, state, updated_at FROM novel_fetch_workshop_bodies WHERE owner_username=? AND book_id=? AND version_id=?`, owner, bookID, versionID).Scan(&ref.Revision, &ref.ContentHash, &ref.CharCount, &ref.State, &updated)
	if err != nil {
		return BodyRef{}, err
	}
	ref.UpdatedAt = updated.UTC().Format(time.RFC3339Nano)
	return ref, nil
}

func (s *MySQLStore) GetBody(ctx context.Context, owner, bookID, versionID string) (BodyRecord, error) {
	bookID = strings.TrimSpace(bookID)
	versionID = strings.TrimSpace(versionID)
	var ref BodyRef
	ref.VersionID = versionID
	var encoding string
	var blob []byte
	var updated time.Time
	err := s.DB.QueryRowContext(ctx, `SELECT revision, content_encoding, content_blob, content_hash, char_count, state, updated_at FROM novel_fetch_workshop_bodies WHERE owner_username=? AND book_id=? AND version_id=?`, owner, bookID, versionID).Scan(&ref.Revision, &encoding, &blob, &ref.ContentHash, &ref.CharCount, &ref.State, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return BodyRecord{}, ErrNotFound
	}
	if err != nil {
		return BodyRecord{}, err
	}
	content, err := decodeBody(blob, encoding)
	if err != nil {
		return BodyRecord{}, fmt.Errorf("decode novel fetch body: %w", err)
	}
	if _, err := s.DB.ExecContext(ctx, `UPDATE novel_fetch_workshop_bodies SET last_needed_at=CURRENT_TIMESTAMP(6) WHERE owner_username=? AND book_id=? AND version_id=?`, owner, bookID, versionID); err != nil {
		return BodyRecord{}, err
	}
	ref.UpdatedAt = updated.UTC().Format(time.RFC3339Nano)
	return BodyRecord{BodyRef: ref, BookID: bookID, Content: content}, nil
}

func (s *MySQLStore) ListBodyRefs(ctx context.Context, owner, bookID string) ([]BodyRef, error) {
	bookID = strings.TrimSpace(bookID)
	rows, err := s.DB.QueryContext(ctx, `SELECT version_id, revision, content_hash, char_count, state, updated_at FROM novel_fetch_workshop_bodies WHERE owner_username=? AND book_id=? ORDER BY version_id ASC`, owner, bookID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	refs := []BodyRef{}
	for rows.Next() {
		var ref BodyRef
		var updated time.Time
		if err := rows.Scan(&ref.VersionID, &ref.Revision, &ref.ContentHash, &ref.CharCount, &ref.State, &updated); err != nil {
			return nil, err
		}
		ref.UpdatedAt = updated.UTC().Format(time.RFC3339Nano)
		refs = append(refs, ref)
	}
	return refs, rows.Err()
}

func (s *MySQLStore) DeleteBody(ctx context.Context, owner, bookID, versionID string) (bool, error) {
	bookID = strings.TrimSpace(bookID)
	versionID = strings.TrimSpace(versionID)
	result, err := s.DB.ExecContext(ctx, `DELETE FROM novel_fetch_workshop_bodies WHERE owner_username=? AND book_id=? AND version_id=?`, owner, bookID, versionID)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}
