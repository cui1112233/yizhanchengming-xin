package batchfactorytext

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	documents   = "batch_factory_v12_text_documents"
	versions    = "batch_factory_v12_text_versions"
	pagesTable  = "batch_factory_v12_text_pages"
	artifacts   = "batch_factory_v12_text_page_artifacts"
	collections = "batch_factory_v12_text_collections"
	items       = "batch_factory_v12_text_collection_items"
)

type MySQLRepository struct{ db *sql.DB }

func NewMySQLRepository(db *sql.DB) *MySQLRepository { return &MySQLRepository{db: db} }

var _ Repository = (*MySQLRepository)(nil)

type scanner interface{ Scan(...any) error }
type reader interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func readRows[T any](ctx context.Context, db reader, scan func(scanner) (T, error), query string, args ...any) ([]T, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]T, 0)
	for rows.Next() {
		item, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func transaction[T any](ctx context.Context, db *sql.DB, fn func(*sql.Tx) (T, error)) (T, error) {
	return transactionWithOptions(ctx, db, nil, fn)
}
func transactionWithOptions[T any](ctx context.Context, db *sql.DB, options *sql.TxOptions, fn func(*sql.Tx) (T, error)) (T, error) {
	var zero T
	tx, err := db.BeginTx(ctx, options)
	if err != nil {
		return zero, err
	}
	defer tx.Rollback()
	value, err := fn(tx)
	if err != nil {
		return zero, err
	}
	if err = tx.Commit(); err != nil {
		return zero, err
	}
	return value, nil
}
func lookupErr(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}
func revision(expected, current int64) error {
	if expected != current {
		return &ConflictError{expected, current}
	}
	return nil
}
func digest(s string) string           { x := sha256.Sum256([]byte(s)); return hex.EncodeToString(x[:]) }
func validJSON(v json.RawMessage) bool { return len(v) == 0 || json.Valid(v) }
func dbJSON(v json.RawMessage) any {
	if len(v) == 0 {
		return nil
	}
	return string(v)
}

type head struct {
	id, version          string
	pagination, revision int64
}

// All mutations lock the document first, including page edits and result publication.
func lockHead(ctx context.Context, tx *sql.Tx, s Scope) (head, error) {
	var h head
	err := tx.QueryRowContext(ctx, "SELECT id, COALESCE(active_version_id,''), pagination_version, revision FROM "+documents+" WHERE owner_username=? AND batch_id=? AND book_id=? FOR UPDATE", s.Owner, s.BatchID, s.BookID).Scan(&h.id, &h.version, &h.pagination, &h.revision)
	return h, lookupErr(err)
}
func touch(ctx context.Context, tx *sql.Tx, h head) error {
	_, err := tx.ExecContext(ctx, "UPDATE "+documents+" SET revision=revision+1,updated_at=NOW(6) WHERE id=?", h.id)
	return err
}

func getDoc(ctx context.Context, db reader, s Scope) (Document, error) {
	var d Document
	var cfg []byte
	err := db.QueryRowContext(ctx, "SELECT id,original_text,COALESCE(active_version_id,''),pagination_version,config_snapshot_json,COALESCE(cover_artifact_id,''),cover_locked,revision FROM "+documents+" WHERE owner_username=? AND batch_id=? AND book_id=?", s.Owner, s.BatchID, s.BookID).Scan(&d.ID, &d.OriginalText, &d.ActiveVersionID, &d.PaginationVersion, &cfg, &d.CoverArtifactID, &d.CoverLocked, &d.Revision)
	d.ConfigSnapshot = cfg
	return d, lookupErr(err)
}
func (r *MySQLRepository) GetDocument(ctx context.Context, s Scope) (Document, error) {
	return getDoc(ctx, r.db, s)
}

func (r *MySQLRepository) SaveOriginal(ctx context.Context, s Scope, text string) (Document, error) {
	if strings.TrimSpace(text) == "" {
		return Document{}, ErrInvalid
	}
	return transaction(ctx, r.db, func(tx *sql.Tx) (Document, error) { return saveOriginalTx(ctx, tx, s, text, false) })
}
func saveOriginalTx(ctx context.Context, tx *sql.Tx, s Scope, text string, newOnly bool) (Document, error) {
	var book string
	err := tx.QueryRowContext(ctx, "SELECT id FROM batch_factory_v12_books WHERE owner_username=? AND batch_id=? AND id=? FOR UPDATE", s.Owner, s.BatchID, s.BookID).Scan(&book)
	if err != nil {
		return Document{}, lookupErr(err)
	}
	d, err := getDoc(ctx, tx, s)
	if err == nil {
		if newOnly {
			return Document{}, ErrConflict
		}
		if d.OriginalText != text {
			return Document{}, fmt.Errorf("%w：已保存的完整原文不同", ErrConflict)
		}
		return d, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return Document{}, err
	}
	d = Document{ID: rand.Text(), OriginalText: text, Revision: 1}
	_, err = tx.ExecContext(ctx, "INSERT INTO "+documents+" (id,owner_username,batch_id,book_id,original_text,created_at,updated_at) VALUES(?,?,?,?,?,NOW(6),NOW(6))", d.ID, s.Owner, s.BatchID, s.BookID, text)
	return d, err
}

func (r *MySQLRepository) CreateVersion(ctx context.Context, s Scope, in VersionInput) (Version, error) {
	if strings.TrimSpace(in.ProductionText) == "" || !validJSON(in.ChangeTrace) {
		return Version{}, ErrInvalid
	}
	switch in.SensitiveMode {
	case "off", "system", "ai":
	default:
		return Version{}, ErrInvalid
	}
	return transaction(ctx, r.db, func(tx *sql.Tx) (Version, error) { return createVersionTx(ctx, tx, s, in) })
}
func createVersionTx(ctx context.Context, tx *sql.Tx, s Scope, in VersionInput) (Version, error) {
	h, err := lockHead(ctx, tx, s)
	if err != nil {
		return Version{}, err
	}
	if err = revision(in.ExpectedRevision, h.revision); err != nil {
		return Version{}, err
	}
	v := Version{ID: rand.Text(), ProductionText: in.ProductionText, SensitiveMode: in.SensitiveMode, ChangeTrace: in.ChangeTrace, ContentDigest: digest(in.ProductionText)}
	_, err = tx.ExecContext(ctx, "INSERT INTO "+versions+" (id,owner_username,book_id,document_id,production_text,sensitive_mode,change_trace_json,content_digest,created_at) VALUES(?,?,?,?,?,?,?,?,NOW(6))", v.ID, s.Owner, s.BookID, h.id, v.ProductionText, v.SensitiveMode, dbJSON(v.ChangeTrace), v.ContentDigest)
	if err != nil {
		return Version{}, err
	}
	_, err = tx.ExecContext(ctx, "UPDATE "+documents+" SET active_version_id=?,revision=revision+1,updated_at=NOW(6) WHERE id=?", v.ID, h.id)
	return v, err
}

func (r *MySQLRepository) ListVersions(ctx context.Context, s Scope) ([]Version, error) {
	return readRows(ctx, r.db, func(row scanner) (Version, error) {
		var v Version
		var trace []byte
		err := row.Scan(&v.ID, &v.ProductionText, &v.SensitiveMode, &trace, &v.ContentDigest)
		v.ChangeTrace = trace
		return v, err
	}, "SELECT v.id,v.production_text,v.sensitive_mode,v.change_trace_json,v.content_digest FROM "+versions+" v JOIN "+documents+" d ON d.id=v.document_id WHERE d.owner_username=? AND d.batch_id=? AND d.book_id=? ORDER BY v.created_at,v.id", s.Owner, s.BatchID, s.BookID)
}

func validateDrafts(text string, drafts []PageDraft) error {
	runes := []rune(NormalizeLineEndings(text))
	offset := 0
	for i, p := range drafts {
		if p.StartOffset != offset || p.EndOffset <= offset || p.EndOffset > len(runes) || string(runes[offset:p.EndOffset]) != p.Text {
			return fmt.Errorf("%w：第 %d 页与生产文本不符", ErrInvalid, i+1)
		}
		offset = p.EndOffset
	}
	if offset != len(runes) || len(drafts) == 0 {
		return fmt.Errorf("%w：分页未覆盖全部生产文本", ErrInvalid)
	}
	return nil
}
func sameRange(p, q PageDraft) bool {
	return p.StartOffset == q.StartOffset && p.EndOffset == q.EndOffset && p.Text == q.Text
}

// Rune intervals form a unique key within a successful pagination. An
// unchanged system page can therefore keep its ID even if its ordinal moves.
func previousPages(ctx context.Context, tx *sql.Tx, s Scope, drafts []PageDraft) (map[int]Page, error) {
	old, err := readRows(ctx, tx, scanPage, "SELECT "+pageCols+pageJoin+" ORDER BY p.ordinal_no FOR UPDATE", s.Owner, s.BatchID, s.BookID)
	if err != nil {
		return nil, err
	}
	byStart := make(map[int]Page, len(old))
	for _, p := range old {
		if p.Locked && (p.Ordinal <= 0 || p.Ordinal > len(drafts) || !sameRange(p.PageDraft, drafts[p.Ordinal-1])) {
			return nil, fmt.Errorf("%w（第 %d 页）", ErrLocked, p.Ordinal)
		}
		byStart[p.StartOffset] = p
	}
	return byStart, nil
}

func samePtrValue[T comparable](a, b *T) bool { return a == b || (a != nil && b != nil && *a == *b) }

func writeNextPage(ctx context.Context, tx *sql.Tx, s Scope, h head, ordinal int, draft PageDraft, previous Page, oldBook, newBook PageStyle) error {
	next := h.pagination + 1
	if previous.Locked {
		_, err := tx.ExecContext(ctx, "UPDATE "+pagesTable+" SET version_id=?,pagination_version=?,revision=revision+1,updated_at=NOW(6) WHERE id=?", h.version, next, previous.ID)
		return err
	}
	preserved := previous.ID != "" && previous.SourceKind == "system" && sameRange(previous.PageDraft, draft)
	var prior PageDraft
	if preserved {
		prior = previous.PageDraft
	}
	overrides, style, err := ResolvePageStyle(newBook, prior, PagePatch{Layout: draft.Layout, Background: draft.Background, ClearLayout: draft.ClearLayout, ClearBackground: draft.ClearBackground})
	if err != nil {
		return err
	}
	draft.Layout, draft.Background = overrides.Layout, overrides.Background
	if !preserved {
		return insertPage(ctx, tx, s, h.version, next, ordinal, draft)
	}
	_, oldStyle, err := ResolvePageStyle(oldBook, prior, PagePatch{})
	if err != nil {
		return err
	}
	sameStyle := oldStyle == style
	if style.HighlightReuseEnabled && style.HighlightRevision == "" && previous.VersionID != h.version {
		sameStyle = false
	}
	if sameStyle && samePtrValue(prior.Layout, draft.Layout) && samePtrValue(prior.Background, draft.Background) {
		_, err := tx.ExecContext(ctx, "UPDATE "+pagesTable+" SET version_id=?,pagination_version=?,ordinal_no=?,updated_at=NOW(6) WHERE id=?", h.version, next, ordinal, previous.ID)
		return err
	}
	// A raw override can change without changing the pixels (for example
	// clearing a background override equal to the book default).
	change := resetPage
	if sameStyle {
		change = "revision=revision+1,updated_at=NOW(6)"
	}
	layout, err := json.Marshal(draft.Layout)
	if err != nil {
		return err
	}
	background, err := json.Marshal(draft.Background)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, "UPDATE "+pagesTable+" SET version_id=?,pagination_version=?,ordinal_no=?,layout_json=?,background_json=?,"+change+" WHERE id=?", h.version, next, ordinal, string(layout), string(background), previous.ID)
	return err
}

func insertPage(ctx context.Context, tx *sql.Tx, s Scope, version string, pagination int64, ordinal int, d PageDraft) error {
	layout, err := json.Marshal(d.Layout)
	if err != nil {
		return err
	}
	background, err := json.Marshal(d.Background)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, "INSERT INTO "+pagesTable+" (id,owner_username,book_id,version_id,pagination_version,ordinal_no,start_offset,end_offset,page_text,content_digest,layout_json,background_json,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,NOW(6),NOW(6))", rand.Text(), s.Owner, s.BookID, version, pagination, ordinal, d.StartOffset, d.EndOffset, d.Text, digest(d.Text), string(layout), string(background))
	return err
}
func (r *MySQLRepository) ReplaceSystemPages(ctx context.Context, s Scope, in PaginationUpdate) (int64, error) {
	if !validJSON(in.ConfigSnapshot) {
		return 0, ErrInvalid
	}
	return transaction(ctx, r.db, func(tx *sql.Tx) (int64, error) { return replacePagesTx(ctx, tx, s, in) })
}
func replacePagesTx(ctx context.Context, tx *sql.Tx, s Scope, in PaginationUpdate) (int64, error) {
	h, err := lockHead(ctx, tx, s)
	if err != nil {
		return 0, err
	}
	if err = revision(in.ExpectedRevision, h.revision); err != nil {
		return 0, err
	}
	if h.version == "" || in.VersionID != h.version {
		return 0, ErrConflict
	}
	var text string
	var priorSnapshot []byte
	err = tx.QueryRowContext(ctx, "SELECT v.production_text,d.config_snapshot_json FROM "+versions+" v JOIN "+documents+" d ON d.id=v.document_id WHERE v.owner_username=? AND v.book_id=? AND v.id=?", s.Owner, s.BookID, h.version).Scan(&text, &priorSnapshot)
	if err != nil {
		return 0, lookupErr(err)
	}
	if err = validateDrafts(text, in.Pages); err != nil {
		return 0, err
	}
	previous, err := previousPages(ctx, tx, s, in.Pages)
	if err != nil {
		return 0, err
	}
	oldBook, err := PageStyleFromSnapshot(priorSnapshot)
	if err != nil {
		return 0, err
	}
	if len(in.ConfigSnapshot) == 0 {
		in.ConfigSnapshot = priorSnapshot
	}
	newBook, err := PageStyleFromSnapshot(in.ConfigSnapshot)
	if err != nil {
		return 0, err
	}
	next := h.pagination + 1
	for i, d := range in.Pages {
		ordinal := i + 1
		err = writeNextPage(ctx, tx, s, h, ordinal, d, previous[d.StartOffset], oldBook, newBook)
		if err != nil {
			return 0, err
		}
	}
	_, err = tx.ExecContext(ctx, "DELETE FROM "+pagesTable+" WHERE owner_username=? AND book_id=? AND pagination_version=? AND locked=FALSE", s.Owner, s.BookID, h.pagination)
	if err != nil {
		return 0, err
	}
	_, err = tx.ExecContext(ctx, "UPDATE "+documents+" SET pagination_version=?,config_snapshot_json=?,revision=revision+1,updated_at=NOW(6) WHERE id=?", next, dbJSON(in.ConfigSnapshot), h.id)
	return next, err
}

const pageCols = "p.id,p.version_id,p.pagination_version,p.ordinal_no,p.start_offset,p.end_offset,p.page_text,p.content_digest,p.layout_json,p.background_json,p.source_kind,p.locked,p.render_version,COALESCE(p.active_artifact_id,''),p.status,COALESCE(p.failed_stage,''),COALESCE(p.error_type,''),COALESCE(p.last_error,''),p.revision"

func scanPage(row scanner) (Page, error) {
	var p Page
	var layout, background []byte
	err := row.Scan(&p.ID, &p.VersionID, &p.PaginationVersion, &p.Ordinal, &p.StartOffset, &p.EndOffset, &p.Text, &p.ContentDigest, &layout, &background, &p.SourceKind, &p.Locked, &p.RenderVersion, &p.ActiveArtifactID, &p.Status, &p.FailedStage, &p.ErrorType, &p.LastError, &p.Revision)
	if err == nil && len(layout) > 0 {
		err = json.Unmarshal(layout, &p.Layout)
	}
	if err == nil && len(background) > 0 {
		err = json.Unmarshal(background, &p.Background)
	}
	return p, lookupErr(err)
}

// Keep the last committed pagination visible while the new version is being
// paginated. In particular, a person must still be able to release a locked
// upload whose source range disappeared from the new text.
const pageJoin = " FROM " + pagesTable + " p JOIN " + documents + " d ON d.owner_username=p.owner_username AND d.book_id=p.book_id AND d.pagination_version=p.pagination_version WHERE d.owner_username=? AND d.batch_id=? AND d.book_id=?"

func (r *MySQLRepository) ListPages(ctx context.Context, s Scope) ([]Page, error) {
	return readRows(ctx, r.db, scanPage, "SELECT "+pageCols+pageJoin+" ORDER BY p.ordinal_no", s.Owner, s.BatchID, s.BookID)
}
func currentPage(ctx context.Context, tx *sql.Tx, s Scope, id string) (Page, error) {
	return scanPage(tx.QueryRowContext(ctx, "SELECT "+pageCols+pageJoin+" AND p.id=? FOR UPDATE", s.Owner, s.BatchID, s.BookID, id))
}
func mutatePage[T any](ctx context.Context, db *sql.DB, s Scope, id string, expected int64, allowReleaseOfOldUpload bool, fn func(*sql.Tx, Page) (T, error)) (T, error) {
	return transaction(ctx, db, func(tx *sql.Tx) (T, error) {
		var zero T
		h, err := lockHead(ctx, tx, s)
		if err != nil {
			return zero, err
		}
		p, err := currentPage(ctx, tx, s, id)
		if err != nil {
			return zero, err
		}
		if err = revision(expected, p.Revision); err != nil {
			return zero, err
		}
		if p.VersionID != h.version && !allowReleaseOfOldUpload {
			return zero, ErrConflict
		}
		out, err := fn(tx, p)
		if err != nil {
			return zero, err
		}
		return out, touch(ctx, tx, h)
	})
}

const resetPage = "active_artifact_id=NULL,status='pending',failed_stage=NULL,error_type=NULL,last_error=NULL,revision=revision+1,updated_at=NOW(6)"

func (r *MySQLRepository) PatchPage(ctx context.Context, s Scope, id string, expected int64, in PagePatch) (Page, error) {
	if in.Layout == nil && in.Background == nil && !in.ClearLayout && !in.ClearBackground {
		return Page{}, ErrInvalid
	}
	return mutatePage(ctx, r.db, s, id, expected, false, func(tx *sql.Tx, p Page) (Page, error) {
		if p.Locked {
			return Page{}, ErrLocked
		}
		overrides, _, err := ResolvePageStyle(PageStyle{}, p.PageDraft, in)
		if err != nil {
			return Page{}, err
		}
		p.Layout, p.Background = overrides.Layout, overrides.Background
		layout, err := json.Marshal(p.Layout)
		if err != nil {
			return Page{}, err
		}
		background, err := json.Marshal(p.Background)
		if err != nil {
			return Page{}, err
		}
		_, err = tx.ExecContext(ctx, "UPDATE "+pagesTable+" SET layout_json=?,background_json=?,"+resetPage+" WHERE id=?", string(layout), string(background), p.ID)
		if err != nil {
			return Page{}, err
		}
		return currentPage(ctx, tx, s, id)
	})
}
func (r *MySQLRepository) RestoreSystemPage(ctx context.Context, s Scope, id string, expected int64) (Page, error) {
	return mutatePage(ctx, r.db, s, id, expected, true, func(tx *sql.Tx, p Page) (Page, error) {
		if !p.Locked {
			return Page{}, ErrConflict
		}
		_, err := tx.ExecContext(ctx, "UPDATE "+pagesTable+" SET locked=FALSE,source_kind='system',"+resetPage+" WHERE id=?", p.ID)
		if err != nil {
			return Page{}, err
		}
		return currentPage(ctx, tx, s, id)
	})
}

func insertArtifact(ctx context.Context, tx *sql.Tx, s Scope, a PageArtifact) error {
	if strings.TrimSpace(a.StorageRef) == "" || utf8.RuneCountInString(a.StorageRef) > 512 || a.ByteSize <= 0 || strings.TrimSpace(a.MediaType) == "" || len(a.SHA256) != 64 {
		return ErrInvalid
	}
	if _, err := hex.DecodeString(a.SHA256); err != nil {
		return ErrInvalid
	}
	_, err := tx.ExecContext(ctx, "INSERT INTO "+artifacts+" (id,owner_username,book_id,page_id,artifact_kind,source_kind,render_version,storage_ref,media_type,byte_size,sha256,created_at) VALUES(?,?,?,NULLIF(?,''),?,?,?,?,?,?,?,NOW(6))", a.ID, s.Owner, s.BookID, a.PageID, a.Kind, a.SourceKind, a.RenderVersion, a.StorageRef, a.MediaType, a.ByteSize, a.SHA256)
	return err
}
func sourceKind(uploaded bool) string {
	if uploaded {
		return "user_upload"
	}
	return "system"
}
func (r *MySQLRepository) savePageArtifact(ctx context.Context, s Scope, id string, expected int64, file ArtifactFile, uploaded bool) (PageArtifact, error) {
	return mutatePage(ctx, r.db, s, id, expected, false, func(tx *sql.Tx, p Page) (PageArtifact, error) {
		if p.Locked && !uploaded {
			return PageArtifact{}, ErrLocked
		}
		a := PageArtifact{ID: rand.Text(), PageID: p.ID, Kind: "page", SourceKind: sourceKind(uploaded), RenderVersion: p.RenderVersion + 1, ArtifactFile: file}
		if err := insertArtifact(ctx, tx, s, a); err != nil {
			return PageArtifact{}, err
		}
		_, err := tx.ExecContext(ctx, "UPDATE "+pagesTable+" SET active_artifact_id=?,locked=?,source_kind=?,render_version=?,status='ready',failed_stage=NULL,error_type=NULL,last_error=NULL,revision=revision+1,updated_at=NOW(6) WHERE id=?", a.ID, uploaded, a.SourceKind, a.RenderVersion, p.ID)
		return a, err
	})
}
func (r *MySQLRepository) LockPageArtifact(ctx context.Context, s Scope, id string, expected int64, file ArtifactFile) (PageArtifact, error) {
	return r.savePageArtifact(ctx, s, id, expected, file, true)
}
func (r *MySQLRepository) SaveSystemPageArtifact(ctx context.Context, s Scope, id string, expected int64, file ArtifactFile) (PageArtifact, error) {
	return r.savePageArtifact(ctx, s, id, expected, file, false)
}

func (r *MySQLRepository) SaveCover(ctx context.Context, s Scope, expected int64, file ArtifactFile, uploaded bool) (PageArtifact, error) {
	return transaction(ctx, r.db, func(tx *sql.Tx) (PageArtifact, error) {
		h, err := lockHead(ctx, tx, s)
		if err != nil {
			return PageArtifact{}, err
		}
		if err = revision(expected, h.revision); err != nil {
			return PageArtifact{}, err
		}
		d, err := getDoc(ctx, tx, s)
		if err != nil {
			return PageArtifact{}, err
		}
		if d.CoverLocked && !uploaded {
			return PageArtifact{}, ErrLocked
		}
		a := PageArtifact{ID: rand.Text(), Kind: "cover", SourceKind: sourceKind(uploaded), RenderVersion: h.revision + 1, ArtifactFile: file}
		if err = insertArtifact(ctx, tx, s, a); err != nil {
			return PageArtifact{}, err
		}
		_, err = tx.ExecContext(ctx, "UPDATE "+documents+" SET cover_artifact_id=?,cover_locked=?,revision=revision+1,updated_at=NOW(6) WHERE id=?", a.ID, uploaded, h.id)
		return a, err
	})
}
func (r *MySQLRepository) RestoreCover(ctx context.Context, s Scope, expected int64) error {
	_, err := transaction(ctx, r.db, func(tx *sql.Tx) (struct{}, error) {
		var z struct{}
		h, err := lockHead(ctx, tx, s)
		if err != nil {
			return z, err
		}
		if err = revision(expected, h.revision); err != nil {
			return z, err
		}
		res, err := tx.ExecContext(ctx, "UPDATE "+documents+" SET cover_artifact_id=NULL,cover_locked=FALSE,revision=revision+1,updated_at=NOW(6) WHERE id=? AND cover_locked=TRUE", h.id)
		if err != nil {
			return z, err
		}
		n, err := res.RowsAffected()
		if err == nil && n == 0 {
			return z, ErrConflict
		}
		return z, err
	})
	return err
}
func (r *MySQLRepository) ListArtifacts(ctx context.Context, s Scope) ([]PageArtifact, error) {
	return readRows(ctx, r.db, func(row scanner) (PageArtifact, error) {
		var a PageArtifact
		err := row.Scan(&a.ID, &a.PageID, &a.Kind, &a.SourceKind, &a.RenderVersion, &a.StorageRef, &a.MediaType, &a.ByteSize, &a.SHA256)
		return a, err
	}, "SELECT a.id,COALESCE(a.page_id,''),a.artifact_kind,a.source_kind,a.render_version,a.storage_ref,a.media_type,a.byte_size,a.sha256 FROM "+artifacts+" a JOIN "+documents+" d ON a.owner_username=d.owner_username AND a.book_id=d.book_id WHERE d.owner_username=? AND d.batch_id=? AND d.book_id=? ORDER BY a.created_at,a.id", s.Owner, s.BatchID, s.BookID)
}

func collectionPages(ctx context.Context, tx *sql.Tx, s Scope, h head, ids []string) ([]CollectionItem, error) {
	selected := make(map[string]bool, len(ids))
	for _, id := range ids {
		if id == "" || selected[id] {
			return nil, ErrInvalid
		}
		selected[id] = true
	}
	all, err := readRows(ctx, tx, func(row scanner) (CollectionItem, error) {
		var i CollectionItem
		err := row.Scan(&i.Ordinal, &i.PageID, &i.ArtifactID, &i.StorageRef, &i.MediaType, &i.SHA256)
		return i, err
	}, "SELECT p.ordinal_no,p.id,COALESCE(a.id,''),COALESCE(a.storage_ref,''),COALESCE(a.media_type,''),COALESCE(a.sha256,'') FROM "+pagesTable+" p LEFT JOIN "+artifacts+" a ON a.id=p.active_artifact_id AND a.page_id=p.id AND a.owner_username=p.owner_username AND a.book_id=p.book_id AND p.status='ready' WHERE p.owner_username=? AND p.book_id=? AND p.version_id=? AND p.pagination_version=? ORDER BY p.ordinal_no", s.Owner, s.BookID, h.version, h.pagination)
	if err != nil {
		return nil, err
	}
	out := make([]CollectionItem, 0, len(all))
	for _, i := range all {
		if len(ids) > 0 && !selected[i.PageID] {
			continue
		}
		if i.ArtifactID == "" {
			return nil, fmt.Errorf("%w（第 %d 页）", ErrNotReady, i.Ordinal)
		}
		out = append(out, i)
	}
	if len(ids) > 0 && len(out) != len(ids) {
		return nil, ErrNotFound
	}
	if len(out) == 0 {
		return nil, ErrNotReady
	}
	return out, nil
}

func (r *MySQLRepository) CreateCollection(ctx context.Context, s Scope, in CollectionInput) (Collection, error) {
	if !validJSON(in.ConfigSnapshot) {
		return Collection{}, ErrInvalid
	}
	return transaction(ctx, r.db, func(tx *sql.Tx) (Collection, error) {
		h, err := lockHead(ctx, tx, s)
		if err != nil {
			return Collection{}, err
		}
		if in.VersionID != h.version || in.PaginationVersion != h.pagination || (in.ExpectedRevision > 0 && in.ExpectedRevision != h.revision) {
			return Collection{}, ErrConflict
		}
		if h.version == "" || h.pagination == 0 {
			return Collection{}, ErrNotReady
		}
		var cover string
		err = tx.QueryRowContext(ctx, "SELECT a.id FROM "+documents+" d JOIN "+artifacts+" a ON a.id=d.cover_artifact_id AND a.owner_username=d.owner_username AND a.book_id=d.book_id AND a.artifact_kind='cover' WHERE d.id=?", h.id).Scan(&cover)
		if errors.Is(err, sql.ErrNoRows) {
			return Collection{}, fmt.Errorf("%w（封面）", ErrNotReady)
		}
		if err != nil {
			return Collection{}, err
		}
		frozen, err := collectionPages(ctx, tx, s, h, in.PageIDs)
		if err != nil {
			return Collection{}, err
		}
		c := Collection{ID: rand.Text(), VersionID: h.version, PaginationVersion: h.pagination, CoverArtifactID: cover, ConfigSnapshot: in.ConfigSnapshot, Status: "ready", Items: frozen}
		_, err = tx.ExecContext(ctx, "INSERT INTO "+collections+" (id,owner_username,batch_id,book_id,version_id,pagination_version,cover_artifact_id,status,config_snapshot_json,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,NOW(6),NOW(6))", c.ID, s.Owner, s.BatchID, s.BookID, c.VersionID, c.PaginationVersion, c.CoverArtifactID, c.Status, dbJSON(c.ConfigSnapshot))
		if err != nil {
			return Collection{}, err
		}
		for _, i := range c.Items {
			_, err = tx.ExecContext(ctx, "INSERT INTO "+items+" (collection_id,ordinal_no,page_id,artifact_id,storage_ref,media_type,sha256) VALUES(?,?,?,?,?,?,?)", c.ID, i.Ordinal, i.PageID, i.ArtifactID, i.StorageRef, i.MediaType, i.SHA256)
			if err != nil {
				return Collection{}, err
			}
		}
		return c, nil
	})
}

const collectionSelect = "SELECT id,version_id,pagination_version,cover_artifact_id,status,COALESCE(storage_ref,''),config_snapshot_json FROM " + collections + " WHERE owner_username=? AND batch_id=? AND book_id=?"

func scanCollection(row scanner) (Collection, error) {
	var c Collection
	var cfg []byte
	err := row.Scan(&c.ID, &c.VersionID, &c.PaginationVersion, &c.CoverArtifactID, &c.Status, &c.StorageRef, &cfg)
	c.ConfigSnapshot = cfg
	return c, lookupErr(err)
}
func (r *MySQLRepository) ListCollections(ctx context.Context, s Scope) ([]Collection, error) {
	return readRows(ctx, r.db, scanCollection, collectionSelect+" ORDER BY created_at DESC,id DESC", s.Owner, s.BatchID, s.BookID)
}
func (r *MySQLRepository) GetCollection(ctx context.Context, s Scope, id string) (Collection, error) {
	c, err := scanCollection(r.db.QueryRowContext(ctx, collectionSelect+" AND id=?", s.Owner, s.BatchID, s.BookID, id))
	if err != nil {
		return Collection{}, err
	}
	c.Items, err = readRows(ctx, r.db, func(row scanner) (CollectionItem, error) {
		var i CollectionItem
		err := row.Scan(&i.Ordinal, &i.PageID, &i.ArtifactID, &i.StorageRef, &i.MediaType, &i.SHA256)
		return i, err
	}, "SELECT i.ordinal_no,i.page_id,i.artifact_id,i.storage_ref,i.media_type,i.sha256 FROM "+items+" i JOIN "+collections+" c ON c.id=i.collection_id WHERE c.owner_username=? AND c.batch_id=? AND c.book_id=? AND c.id=? ORDER BY i.ordinal_no", s.Owner, s.BatchID, s.BookID, id)
	return c, err
}
