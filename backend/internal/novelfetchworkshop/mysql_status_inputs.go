package novelfetchworkshop

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
)

// Read only the fields needed by queue polling. Older documents without a
// durable task state retain the full payload for legacy state derivation.
// The current batch id is required here: queue counts are scoped to the
// active batch so unfinished books from older days cannot inflate today's
// queue banner, and the realtime handler must not need a second config read.
func (s *MySQLStore) AutomationStatusInputs(ctx context.Context, owner string) (map[string]any, []Document, error) {
	var rawSettings []byte
	err := s.DB.QueryRowContext(ctx, `SELECT COALESCE(settings_json,'{}') FROM novel_fetch_workshop_configs WHERE owner_username=?`, owner).Scan(&rawSettings)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, nil, err
	}
	settings := map[string]any{}
	if len(rawSettings) > 0 {
		if decodeErr := json.Unmarshal(rawSettings, &settings); decodeErr != nil {
			return nil, nil, decodeErr
		}
	}
	rows, err := s.DB.QueryContext(ctx, `SELECT CASE WHEN COALESCE(NULLIF(JSON_UNQUOTE(JSON_EXTRACT(document_json,'$.meta.status')),''),NULLIF(JSON_UNQUOTE(JSON_EXTRACT(document_json,'$.meta.queueState')),''),NULLIF(JSON_UNQUOTE(JSON_EXTRACT(document_json,'$.meta.taskStatus')),'')) IS NULL THEN document_json ELSE JSON_OBJECT('bookId',JSON_EXTRACT(document_json,'$.bookId'),'meta',JSON_OBJECT('status',JSON_EXTRACT(document_json,'$.meta.status'),'queueState',JSON_EXTRACT(document_json,'$.meta.queueState'),'taskStatus',JSON_EXTRACT(document_json,'$.meta.taskStatus'),'currentBatchId',JSON_EXTRACT(document_json,'$.meta.currentBatchId'),'batchCreatedAt',JSON_EXTRACT(document_json,'$.meta.batchCreatedAt'),'createdAt',JSON_EXTRACT(document_json,'$.meta.createdAt'),'originalStatus',JSON_EXTRACT(document_json,'$.meta.originalStatus'),'aiStatus',JSON_EXTRACT(document_json,'$.meta.aiStatus'),'siteSubmitStatus',JSON_EXTRACT(document_json,'$.meta.siteSubmitStatus')),'updatedAt',JSON_EXTRACT(document_json,'$.updatedAt')) END FROM novel_fetch_workshop_documents WHERE owner_username=?`, owner)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	documents := []Document{}
	for rows.Next() {
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			return nil, nil, err
		}
		var doc Document
		if err := json.Unmarshal(raw, &doc); err != nil {
			return nil, nil, err
		}
		documents = append(documents, doc)
	}
	return settings, documents, rows.Err()
}
