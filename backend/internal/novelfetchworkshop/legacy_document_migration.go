package novelfetchworkshop

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
)

func hasLegacyDocumentBodies(document Document) bool {
	if document.Original != "" || document.OriginalRaw != "" {
		return true
	}
	for _, value := range document.Versions {
		if _, ok := value.(string); ok {
			return true
		}
	}
	return false
}

func legacyBodyRecords(document Document) []BodyRecord {
	document = normalizeDocument(document)
	records := []BodyRecord{}

	original := document.Original
	if original == "" {
		original = document.OriginalRaw
	}
	if original != "" {
		records = append(records, BodyRecord{
			BookID: document.BookID,
			BodyRef: BodyRef{
				VersionID: "original",
				State:     "ready",
			},
			Content: original,
		})
	}

	versionIDs := make([]string, 0, len(document.Versions))
	for rawVersionID, value := range document.Versions {
		if _, ok := value.(string); !ok {
			continue
		}
		versionID := strings.TrimSpace(rawVersionID)
		if versionID == "" || versionID == "originalRaw" {
			continue
		}
		if versionID == "original" && original != "" {
			continue
		}
		versionIDs = append(versionIDs, versionID)
	}
	sort.Strings(versionIDs)
	for _, versionID := range versionIDs {
		value, ok := document.Versions[versionID].(string)
		if !ok {
			continue
		}
		if versionID == "original" {
			records = append(records, BodyRecord{
				BookID:  document.BookID,
				BodyRef: BodyRef{VersionID: "original", State: "ready"},
				Content: value,
			})
			continue
		}
		records = append(records, BodyRecord{
			BookID:  document.BookID,
			BodyRef: BodyRef{VersionID: versionID, State: "ready"},
			Content: value,
		})
	}
	return records
}

func lightweightLegacyDocument(document Document, bodies []BodyRecord) Document {
	document = normalizeDocument(document)
	document.Original = ""
	document.OriginalRaw = ""

	versions := map[string]any{}
	for versionID, value := range document.Versions {
		if _, isText := value.(string); !isText {
			versions[versionID] = value
		}
	}
	for _, body := range bodies {
		if body.VersionID == "original" {
			continue
		}
		versions[body.VersionID] = map[string]any{
			"versionId": body.VersionID,
			"state":     body.State,
		}
	}
	document.Versions = versions
	return document
}

func (s *MySQLStore) MigrateLegacyDocumentBodies(ctx context.Context, owner, bookID string) error {
	bookID = strings.TrimSpace(bookID)
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var raw []byte
	err = tx.QueryRowContext(ctx, `SELECT document_json FROM novel_fetch_workshop_documents WHERE owner_username=? AND book_id=? FOR UPDATE`, owner, bookID).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}

	var document Document
	if err := json.Unmarshal(raw, &document); err != nil {
		return fmt.Errorf("decode legacy novel fetch document: %w", err)
	}
	document = normalizeDocument(document)
	document.BookID = bookID
	bodies := legacyBodyRecords(document)
	if len(bodies) == 0 {
		return tx.Commit()
	}

	for _, body := range bodies {
		blob, hash, charCount, err := encodeBody(body.Content)
		if err != nil {
			return err
		}
		state := strings.TrimSpace(body.State)
		if state == "" {
			state = "ready"
		}
		if _, err := tx.ExecContext(ctx, `INSERT IGNORE INTO novel_fetch_workshop_bodies(owner_username,book_id,version_id,revision,content_encoding,content_blob,content_hash,char_count,state,created_at,updated_at,last_needed_at,expires_at) VALUES(?,?,?,1,'gzip',?,?,?,?,CURRENT_TIMESTAMP(6),CURRENT_TIMESTAMP(6),CURRENT_TIMESTAMP(6),NULL)`, owner, bookID, body.VersionID, blob, hash, charCount, state); err != nil {
			return err
		}
	}

	lightweight := lightweightLegacyDocument(document, bodies)
	lightweightRaw, err := json.Marshal(lightweight)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE novel_fetch_workshop_documents SET document_json=CAST(? AS CHAR CHARACTER SET utf8mb4), updated_at=CURRENT_TIMESTAMP(6) WHERE owner_username=? AND book_id=?`, string(lightweightRaw), owner, bookID); err != nil {
		return err
	}
	return tx.Commit()
}
