package novelfetchworkshop

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

// Summary reads must not trigger legacy body migrations or object downloads.
func (s *MySQLStore) GetDocumentSummary(ctx context.Context, owner, bookID string) (Document, error) {
	var raw []byte
	var updated time.Time
	err := s.DB.QueryRowContext(ctx, `SELECT JSON_REMOVE(document_json, '$.original', '$.originalRaw', '$.processedText', '$.versions', '$.logs'), updated_at FROM novel_fetch_workshop_documents WHERE owner_username=? AND book_id=?`, owner, bookID).Scan(&raw, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return Document{}, ErrNotFound
	}
	if err != nil {
		return Document{}, err
	}
	var document Document
	if err = json.Unmarshal(raw, &document); err != nil {
		return Document{}, err
	}
	document.BookID = bookID
	document.UpdatedAt = updated.UTC().Format(time.RFC3339Nano)
	return document, nil
}

func ReadDocumentSummary(ctx context.Context, store Store, owner, bookID string) (Document, error) {
	if reader, ok := store.(interface {
		GetDocumentSummary(context.Context, string, string) (Document, error)
	}); ok {
		return reader.GetDocumentSummary(ctx, owner, bookID)
	}
	document, err := store.GetDocument(ctx, owner, bookID)
	if err != nil {
		return Document{}, err
	}
	document.Original = ""
	document.OriginalRaw = ""
	document.ProcessedText = ""
	document.Versions = nil
	document.Logs = nil
	return document, nil
}
