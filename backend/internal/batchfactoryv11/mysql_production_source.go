package batchfactoryv11

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
)

// LockProductionSource reads the source and working draft for another module's
// publication transaction. Reads and source edits must use the same MySQL database.
func (s *MySQLStore) LockProductionSource(ctx context.Context, tx *sql.Tx, owner, batchID, bookID string) (Book, string, error) {
	var b Book
	var metadata []byte
	err := tx.QueryRowContext(ctx, `SELECT b.id, COALESCE(r.source_text,''),COALESCE(r.txt_text,''),r.source_metadata_json
		FROM batch_factory_v12_books b JOIN batch_factory_v12_book_records r ON r.book_id=b.id
		WHERE b.owner_username=? AND b.batch_id=? AND b.id=? FOR UPDATE`, owner, batchID, bookID).Scan(&b.ID, &b.SourceText, &b.TxtText, &metadata)
	if errors.Is(err, sql.ErrNoRows) {
		return b, "", ErrNotFound
	}
	if err != nil {
		return b, "", err
	}
	if err = json.Unmarshal(metadata, &b.SourceMetadata); err != nil {
		return b, "", err
	}
	var draft string
	err = tx.QueryRowContext(ctx, `SELECT COALESCE(content,'') FROM batch_factory_v12_drafts WHERE owner_username=? AND draft_key=? AND kind='working-front-content' AND scope=? FOR UPDATE`, owner, "working-front:"+bookID, batchID).Scan(&draft)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	return b, draft, err
}
