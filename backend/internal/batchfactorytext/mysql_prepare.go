package batchfactorytext

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
)

// PreparePages publishes the version and its pages in one transaction. A locked
// page conflict leaves the previously active version and all its links intact.
func (r *MySQLRepository) PreparePages(ctx context.Context, s Scope, v VersionInput, p PaginationUpdate) (int64, error) {
	if strings.TrimSpace(v.ProductionText) == "" || (v.SensitiveMode != "system" && v.SensitiveMode != "off" && v.SensitiveMode != "ai") || !validJSON(v.ChangeTrace) || !validJSON(p.ConfigSnapshot) {
		return 0, ErrInvalid
	}
	// Lock the working-draft key range even when no draft exists yet.
	return transactionWithOptions(ctx, r.db, &sql.TxOptions{Isolation: sql.LevelRepeatableRead}, func(tx *sql.Tx) (int64, error) {
		if p.VerifyProduction != nil {
			if err := p.VerifyProduction(ctx, tx); err != nil {
				return 0, err
			}
		}
		if p.ExpectedRevision == 0 && strings.TrimSpace(p.OriginalText) != "" {
			doc, err := saveOriginalTx(ctx, tx, s, p.OriginalText, true)
			if err != nil {
				return 0, err
			}
			p.ExpectedRevision = doc.Revision
		}
		h, err := lockHead(ctx, tx, s)
		if err != nil {
			return 0, err
		}
		if err = revision(p.ExpectedRevision, h.revision); err != nil {
			return 0, err
		}
		var content, mode string
		var trace []byte
		if h.version != "" {
			if err = tx.QueryRowContext(ctx, "SELECT content_digest,sensitive_mode,change_trace_json FROM "+versions+" WHERE owner_username=? AND book_id=? AND id=?", s.Owner, s.BookID, h.version).Scan(&content, &mode, &trace); err != nil {
				return 0, err
			}
		}
		p.VersionID = h.version
		var stored, requested any
		_ = json.Unmarshal(trace, &stored)
		_ = json.Unmarshal(v.ChangeTrace, &requested)
		storedJSON, _ := json.Marshal(stored)
		requestedJSON, _ := json.Marshal(requested)
		if content != digest(v.ProductionText) || mode != v.SensitiveMode || string(storedJSON) != string(requestedJSON) {
			v.ExpectedRevision = h.revision
			created, err := createVersionTx(ctx, tx, s, v)
			if err != nil {
				return 0, err
			}
			p.VersionID = created.ID
			p.ExpectedRevision++
		}
		old, err := getDoc(ctx, tx, s)
		if err != nil {
			return 0, err
		}
		pagination, err := replacePagesTx(ctx, tx, s, p)
		if err != nil {
			return 0, err
		}
		if old.CoverArtifactID != "" && !old.CoverLocked && coverFingerprint(old.ConfigSnapshot) != coverFingerprint(p.ConfigSnapshot) {
			if _, err = tx.ExecContext(ctx, "UPDATE "+documents+" SET cover_artifact_id=NULL WHERE id=? AND cover_locked=FALSE", h.id); err != nil {
				return 0, err
			}
		}
		return pagination, nil
	})
}

func (r *MySQLRepository) GetVersion(ctx context.Context, s Scope, id string) (Version, error) {
	var v Version
	var trace []byte
	err := r.db.QueryRowContext(ctx, "SELECT v.id,v.production_text,v.sensitive_mode,v.change_trace_json,v.content_digest FROM "+versions+" v JOIN "+documents+" d ON d.id=v.document_id WHERE d.owner_username=? AND d.batch_id=? AND d.book_id=? AND v.id=?", s.Owner, s.BatchID, s.BookID, id).Scan(&v.ID, &v.ProductionText, &v.SensitiveMode, &trace, &v.ContentDigest)
	v.ChangeTrace = trace
	return v, lookupErr(err)
}
func coverFingerprint(snapshot json.RawMessage) string {
	var s struct {
		Fingerprint string `json:"coverFingerprint"`
	}
	_ = json.Unmarshal(snapshot, &s)
	return s.Fingerprint
}
