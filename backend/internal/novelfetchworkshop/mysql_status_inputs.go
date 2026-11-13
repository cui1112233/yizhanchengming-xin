package novelfetchworkshop

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
)

// Read only the fields needed by queue polling. Older documents without a
// durable task state retain the full payload for legacy state derivation.
func (s *MySQLStore) AutomationStatusInputs(ctx context.Context, owner string) (map[string]any, []Document, error) {
	var enabled string
	err := s.DB.QueryRowContext(ctx, `SELECT COALESCE(JSON_UNQUOTE(JSON_EXTRACT(settings_json,'$.novel_fetch_automation.enabled')),'false') FROM novel_fetch_workshop_configs WHERE owner_username=?`, owner).Scan(&enabled)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, nil, err
	}
	settings := map[string]any{"novel_fetch_automation": map[string]any{"enabled": enabled == "true"}}
	rows, err := s.DB.QueryContext(ctx, `SELECT CASE WHEN COALESCE(NULLIF(JSON_UNQUOTE(JSON_EXTRACT(document_json,'$.meta.status')),''),NULLIF(JSON_UNQUOTE(JSON_EXTRACT(document_json,'$.meta.queueState')),''),NULLIF(JSON_UNQUOTE(JSON_EXTRACT(document_json,'$.meta.taskStatus')),'')) IS NULL THEN document_json ELSE JSON_OBJECT('meta',JSON_OBJECT('status',JSON_EXTRACT(document_json,'$.meta.status'),'queueState',JSON_EXTRACT(document_json,'$.meta.queueState'),'taskStatus',JSON_EXTRACT(document_json,'$.meta.taskStatus'))) END FROM novel_fetch_workshop_documents WHERE owner_username=?`, owner)
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
