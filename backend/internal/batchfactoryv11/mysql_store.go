package batchfactoryv11

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

type MySQLStore struct{ db *sql.DB }

var _ Store = (*MySQLStore)(nil)

func NewMySQLStore(db *sql.DB) *MySQLStore { return &MySQLStore{db: db} }

func newID(prefix string) (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return prefix + "_" + hex.EncodeToString(buf), nil
}

func (s *MySQLStore) CreateIntake(ctx context.Context, owner string, input NovelFetchIntakeInput) (Intake, error) {
	input = normalizeNovelFetchIntake(input)
	id, err := newID("intake")
	if err != nil {
		return Intake{}, err
	}
	payload, err := json.Marshal(input)
	if err != nil {
		return Intake{}, ErrInvalid
	}
	now := time.Now().UTC()
	if _, err := s.db.ExecContext(ctx, `INSERT INTO batch_factory_v12_intakes(id,owner_username,payload_json,created_at) VALUES(?,?,?,?)`, id, owner, string(payload), now); err != nil {
		return Intake{}, err
	}
	return Intake{ID: id, Owner: owner, Payload: payload, CreatedAt: now}, nil
}

func (s *MySQLStore) GetIntake(ctx context.Context, owner, id string) (Intake, error) {
	var v Intake
	var payload []byte
	var consumed sql.NullTime
	err := s.db.QueryRowContext(ctx, `SELECT id,payload_json,consumed_at,created_at FROM batch_factory_v12_intakes WHERE id=? AND owner_username=?`, id, owner).Scan(&v.ID, &payload, &consumed, &v.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Intake{}, ErrNotFound
	}
	if err != nil {
		return Intake{}, err
	}
	v.Owner, v.Payload = owner, json.RawMessage(payload)
	if consumed.Valid {
		t := consumed.Time
		v.ConsumedAt = &t
	}
	return v, nil
}

func (s *MySQLStore) CreateBatchFromIntake(ctx context.Context, owner, intakeID string, input CreateBatchInput) (Batch, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Batch{}, err
	}
	defer tx.Rollback()
	var payload []byte
	var consumed sql.NullTime
	err = tx.QueryRowContext(ctx, `SELECT payload_json,consumed_at FROM batch_factory_v12_intakes WHERE id=? AND owner_username=? FOR UPDATE`, intakeID, owner).Scan(&payload, &consumed)
	if errors.Is(err, sql.ErrNoRows) {
		return Batch{}, ErrNotFound
	}
	if err != nil {
		return Batch{}, err
	}
	if consumed.Valid {
		return Batch{}, ErrConflict
	}
	var intake NovelFetchIntakeInput
	if err := json.Unmarshal(payload, &intake); err != nil {
		return Batch{}, err
	}
	if len(input.Books) == 0 {
		input.Books = intake.Books
	}
	if strings.TrimSpace(input.Title) == "" {
		input.Title = "Novel Fetch Batch"
	}
	batch, err := createBatchTx(ctx, tx, owner, input, intakeID)
	if err != nil {
		return Batch{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE batch_factory_v12_intakes SET consumed_at=CURRENT_TIMESTAMP(6) WHERE id=? AND owner_username=?`, intakeID, owner); err != nil {
		return Batch{}, err
	}
	if err := tx.Commit(); err != nil {
		return Batch{}, err
	}
	return batch, nil
}

// AppendBooksFromIntake keeps the selected Novel Fetch content version as a
// new book record in an existing batch. It never replaces previously produced
// book data, and it consumes the one-time intake only after the whole append
// transaction succeeds.
func (s *MySQLStore) AppendBooksFromIntake(ctx context.Context, owner, batchID, intakeID string, allowDuplicate bool) (Batch, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Batch{}, err
	}
	defer tx.Rollback()
	var payload []byte
	var consumed sql.NullTime
	if err = tx.QueryRowContext(ctx, `SELECT payload_json,consumed_at FROM batch_factory_v12_intakes WHERE id=? AND owner_username=? FOR UPDATE`, intakeID, owner).Scan(&payload, &consumed); errors.Is(err, sql.ErrNoRows) {
		return Batch{}, ErrNotFound
	} else if err != nil {
		return Batch{}, err
	}
	if consumed.Valid {
		return Batch{}, ErrConflict
	}
	var batchRevision int64
	if err = tx.QueryRowContext(ctx, `SELECT revision FROM batch_factory_v12_batches WHERE id=? AND owner_username=? FOR UPDATE`, batchID, owner).Scan(&batchRevision); errors.Is(err, sql.ErrNoRows) {
		return Batch{}, ErrNotFound
	} else if err != nil {
		return Batch{}, err
	}
	var intake NovelFetchIntakeInput
	if err = json.Unmarshal(payload, &intake); err != nil {
		return Batch{}, ErrInvalid
	}
	books := intake.Books
	if !allowDuplicate {
		existing := map[string]bool{}
		rows, err := tx.QueryContext(ctx, `SELECT COALESCE(source_book_id,''),COALESCE(platform,'') FROM batch_factory_v12_book_records r JOIN batch_factory_v12_books b ON b.id=r.book_id WHERE b.batch_id=? AND b.owner_username=? FOR UPDATE`, batchID, owner)
		if err != nil {
			return Batch{}, err
		}
		for rows.Next() {
			var sourceID, platform string
			if err := rows.Scan(&sourceID, &platform); err != nil {
				rows.Close()
				return Batch{}, err
			}
			if key := batchBookKey(CreateBookInput{BookID: sourceID, Platform: platform}); key != "" {
				existing[key] = true
			}
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return Batch{}, err
		}
		rows.Close()
		books = uniqueBatchBooks(intake.Books, existing)
	}
	var nextOrdinal int
	if err = tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(ordinal)+1,0) FROM batch_factory_v12_books WHERE batch_id=? AND owner_username=?`, batchID, owner).Scan(&nextOrdinal); err != nil {
		return Batch{}, err
	}
	now := time.Now().UTC()
	for _, bookInput := range books {
		bookID, err := newID("book")
		if err != nil {
			return Batch{}, err
		}
		sourceID := sourceBookID(bookInput)
		if sourceID == "" {
			sourceID = bookID
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO batch_factory_v12_books(id,batch_id,owner_username,ordinal,revision,created_at,updated_at) VALUES(?,?,?,?,1,?,?)`, bookID, batchID, owner, nextOrdinal, now, now); err != nil {
			return Batch{}, err
		}
		nextOrdinal++
		var metadata any
		if len(bookInput.SourceMetadata) > 0 {
			var encodedMetadata []byte
			encodedMetadata, err = json.Marshal(bookInput.SourceMetadata)
			if err != nil {
				return Batch{}, ErrInvalid
			}
			// Bind JSON as text. MySQL rejects driver-bound []byte as binary
			// when inserting into a JSON column.
			metadata = string(encodedMetadata)
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO batch_factory_v12_book_records(book_id,source_book_id,source_task_id,platform,title,source_text,txt_text,txt_file_name,source_metadata_json) VALUES(?,?,?,?,?,?,?,?,?)`, bookID, nullableString(sourceID), nullableString(bookInput.SourceTaskID), nullableString(bookInput.Platform), bookInput.Title, nullableString(bookInput.SourceText), nullableString(bookInput.TxtText), nullableString(bookInput.TxtFileName), metadata); err != nil {
			return Batch{}, err
		}
	}
	if len(books) > 0 {
		if _, err = tx.ExecContext(ctx, `UPDATE batch_factory_v12_batches SET revision=?,updated_at=? WHERE id=? AND owner_username=?`, batchRevision+1, now, batchID, owner); err != nil {
			return Batch{}, err
		}
	}
	if _, err = tx.ExecContext(ctx, `UPDATE batch_factory_v12_intakes SET consumed_at=? WHERE id=? AND owner_username=?`, now, intakeID, owner); err != nil {
		return Batch{}, err
	}
	if err = tx.Commit(); err != nil {
		return Batch{}, err
	}
	return s.GetBatch(ctx, owner, batchID)
}

func (s *MySQLStore) CreateBatch(ctx context.Context, owner string, input CreateBatchInput) (Batch, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Batch{}, err
	}
	defer tx.Rollback()
	batch, err := createBatchTx(ctx, tx, owner, input, "")
	if err != nil {
		return Batch{}, err
	}
	if err := tx.Commit(); err != nil {
		return Batch{}, err
	}
	return batch, nil
}

func createBatchTx(ctx context.Context, tx *sql.Tx, owner string, input CreateBatchInput, sourceIntakeID string) (Batch, error) {
	batchID, err := newID("batch")
	if err != nil {
		return Batch{}, err
	}
	now := time.Now().UTC()
	title := strings.TrimSpace(input.Title)
	if title == "" {
		title = "Untitled Batch"
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO batch_factory_v12_batches(id,owner_username,revision,created_at,updated_at) VALUES(?,?,1,?,?)`, batchID, owner, now, now); err != nil {
		return Batch{}, err
	}
	var source any
	if sourceIntakeID != "" {
		source = sourceIntakeID
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO batch_factory_v12_batch_records(batch_id,title,source_intake_id) VALUES(?,?,?)`, batchID, title, source); err != nil {
		return Batch{}, err
	}
	batch := Batch{ID: batchID, Title: title, SourceIntakeID: sourceIntakeID, Revision: 1, Books: []Book{}, CreatedAt: now, UpdatedAt: now}
	for bookOrdinal, bi := range uniqueBatchBooks(input.Books, map[string]bool{}) {
		if len(input.GiantAutomationPlan) > 0 {
			bi.SourceMetadata = copySourceMetadataWithGiantPlan(bi.SourceMetadata, input.GiantAutomationPlan)
		}
		bookID, err := newID("book")
		if err != nil {
			return Batch{}, err
		}
		sourceID := sourceBookID(bi)
		if sourceID == "" {
			sourceID = bookID
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO batch_factory_v12_books(id,batch_id,owner_username,ordinal,revision,created_at,updated_at) VALUES(?,?,?,?,1,?,?)`, bookID, batchID, owner, bookOrdinal, now, now); err != nil {
			return Batch{}, err
		}
		metadata := any(nil)
		if len(bi.SourceMetadata) > 0 {
			encodedMetadata, marshalErr := json.Marshal(bi.SourceMetadata)
			if marshalErr != nil {
				return Batch{}, ErrInvalid
			}
			metadata = string(encodedMetadata)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO batch_factory_v12_book_records(book_id,source_book_id,source_task_id,platform,title,source_text,txt_text,txt_file_name,source_metadata_json) VALUES(?,?,?,?,?,?,?,?,?)`, bookID, nullableString(sourceID), nullableString(bi.SourceTaskID), nullableString(bi.Platform), bi.Title, nullableString(bi.SourceText), nullableString(bi.TxtText), nullableString(bi.TxtFileName), metadata); err != nil {
			return Batch{}, err
		}
		book := Book{ID: bookID, BookID: sourceID, BatchID: batchID, Title: bi.Title, SourceText: bi.SourceText, SourceTaskID: bi.SourceTaskID, Platform: bi.Platform, TxtText: bi.TxtText, TxtFileName: bi.TxtFileName, SourceMetadata: bi.SourceMetadata, Revision: 1, Videos: []Video{}}
		for videoOrdinal, vi := range bi.Videos {
			videoID, err := newID("video")
			if err != nil {
				return Batch{}, err
			}
			if _, err := tx.ExecContext(ctx, `INSERT INTO batch_factory_v12_videos(id,batch_id,book_id,owner_username,ordinal,compatibility_state,revision,created_at,updated_at) VALUES(?,?,?,?,?,'active',1,?,?)`, videoID, batchID, bookID, owner, videoOrdinal, now, now); err != nil {
				return Batch{}, err
			}
			if _, err := tx.ExecContext(ctx, `INSERT INTO batch_factory_v12_video_records(video_id,label,video_prompt,visual_prompt,duration_seconds) VALUES(?,?,?,?,?)`, videoID, vi.Label, nullableString(vi.VideoPrompt), nullableString(vi.VisualPrompt), nullableFloat(vi.DurationSeconds)); err != nil {
				return Batch{}, err
			}
			book.Videos = append(book.Videos, Video{ID: videoID, BatchID: batchID, BookID: bookID, Label: vi.Label, VideoPrompt: vi.VideoPrompt, VisualPrompt: vi.VisualPrompt, DurationSeconds: vi.DurationSeconds, CompatibilityState: "active", Revision: 1})
		}
		batch.Books = append(batch.Books, book)
	}
	if len(input.InitialBatchSettings) > 0 {
		encoded, err := json.Marshal(input.InitialBatchSettings)
		if err != nil {
			return Batch{}, ErrInvalid
		}
		const initialRevision int64 = 2
		if _, err := tx.ExecContext(ctx, `UPDATE batch_factory_v12_batches SET revision=?,updated_at=? WHERE id=? AND owner_username=?`, initialRevision, now, batchID, owner); err != nil {
			return Batch{}, err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO batch_factory_v12_settings_patches(scope_type,scope_id,owner_username,batch_id,book_id,video_id,patch_json,revision,updated_at) VALUES(?,?,?,?,?,?,?,?,?)`, string(ScopeBatch), batchID, owner, batchID, nil, nil, string(encoded), initialRevision, now); err != nil {
			return Batch{}, err
		}
		batch.Revision = initialRevision
		batch.SettingsState = SettingsState{Patch: cloneSettingsPatch(input.InitialBatchSettings), Revision: initialRevision}
	}
	return batch, nil
}

func copySourceMetadataWithGiantPlan(metadata map[string]any, plan map[string]any) map[string]any {
	copyMetadata := make(map[string]any, len(metadata)+1)
	for key, value := range metadata {
		copyMetadata[key] = value
	}
	copyPlan := make(map[string]any, len(plan))
	for key, value := range plan {
		copyPlan[key] = value
	}
	copyMetadata["giantAutomationPlan"] = copyPlan
	return copyMetadata
}

func cloneSettingsPatch(patch SettingsPatch) SettingsPatch {
	copyPatch := make(SettingsPatch, len(patch))
	for key, value := range patch {
		copyPatch[key] = append(json.RawMessage(nil), value...)
	}
	return copyPatch
}

func (s *MySQLStore) UpdateBookMetadata(ctx context.Context, owner, batchID, bookID string, input UpdateBookMetadataInput) (Book, error) {
	if strings.TrimSpace(owner) == "" || strings.TrimSpace(batchID) == "" || strings.TrimSpace(bookID) == "" || input.ExpectedRevision < 1 {
		return Book{}, ErrInvalid
	}
	sourceBookID := strings.TrimSpace(input.SourceBookID)
	if sourceBookID != "" && !validSourceBookID(sourceBookID) {
		return Book{}, fmt.Errorf("%w：%s", ErrInvalid, SourceBookIDValidationMessage)
	}
	metadata, err := json.Marshal(input.Metadata)
	if err != nil {
		return Book{}, ErrInvalid
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Book{}, err
	}
	defer tx.Rollback()
	var current int64
	err = tx.QueryRowContext(ctx, `SELECT revision FROM batch_factory_v12_books WHERE id=? AND batch_id=? AND owner_username=? FOR UPDATE`, bookID, batchID, owner).Scan(&current)
	if errors.Is(err, sql.ErrNoRows) {
		return Book{}, ErrNotFound
	}
	if err != nil {
		return Book{}, err
	}
	if current != input.ExpectedRevision {
		return Book{}, ErrConflict
	}
	title := strings.TrimSpace(input.Title)
	if sourceBookID != "" {
		// sourceLine 需要“书号\t书名”，用户只改书号时书名取记录里的当前值。
		if title == "" {
			if err = tx.QueryRowContext(ctx, `SELECT COALESCE(title,'') FROM batch_factory_v12_book_records WHERE book_id=? FOR UPDATE`, bookID).Scan(&title); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return Book{}, ErrNotFound
				}
				return Book{}, err
			}
		}
		metadataValues := map[string]any{}
		if string(metadata) != "null" {
			if err = json.Unmarshal(metadata, &metadataValues); err != nil {
				return Book{}, ErrInvalid
			}
		}
		applySourceBookIDMetadata(metadataValues, sourceBookID, title)
		if metadata, err = json.Marshal(metadataValues); err != nil {
			return Book{}, ErrInvalid
		}
	}
	// MySQL JSON rejects driver-bound []byte as _binary. Bind JSON as UTF-8 text
	// so metadata updates (including direct-read and OCR state) are durable.
	switch {
	case title != "" && sourceBookID != "":
		if _, err = tx.ExecContext(ctx, `UPDATE batch_factory_v12_book_records SET title=?,source_book_id=?,source_metadata_json=? WHERE book_id=?`, title, sourceBookID, string(metadata), bookID); err != nil {
			return Book{}, err
		}
	case title != "":
		if _, err = tx.ExecContext(ctx, `UPDATE batch_factory_v12_book_records SET title=?,source_metadata_json=? WHERE book_id=?`, title, string(metadata), bookID); err != nil {
			return Book{}, err
		}
	case sourceBookID != "":
		if _, err = tx.ExecContext(ctx, `UPDATE batch_factory_v12_book_records SET source_book_id=?,source_metadata_json=? WHERE book_id=?`, sourceBookID, string(metadata), bookID); err != nil {
			return Book{}, err
		}
	default:
		if _, err = tx.ExecContext(ctx, `UPDATE batch_factory_v12_book_records SET source_metadata_json=? WHERE book_id=?`, string(metadata), bookID); err != nil {
			return Book{}, err
		}
	}
	if _, err = tx.ExecContext(ctx, `UPDATE batch_factory_v12_books SET revision=revision+1,updated_at=? WHERE id=? AND batch_id=? AND owner_username=?`, time.Now().UTC(), bookID, batchID, owner); err != nil {
		return Book{}, err
	}
	if err = tx.Commit(); err != nil {
		return Book{}, err
	}
	batch, err := s.GetBatch(ctx, owner, batchID)
	if err != nil {
		return Book{}, err
	}
	for _, book := range batch.Books {
		if book.ID == bookID {
			return book, nil
		}
	}
	return Book{}, ErrNotFound
}

func (s *MySQLStore) CaptureBookSource(ctx context.Context, owner, batchID, bookID string, input CaptureBookSourceInput) (Book, error) {
	sourceText, validSourceText := normalizeCapturedSourceText(input.SourceText)
	rawSourceText, validRawSourceText := normalizeCapturedSourceText(input.RawSourceText)
	if strings.TrimSpace(owner) == "" || strings.TrimSpace(batchID) == "" || strings.TrimSpace(bookID) == "" || input.ExpectedRevision < 1 || (!validSourceText && !validRawSourceText) {
		return Book{}, ErrInvalid
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Book{}, err
	}
	defer tx.Rollback()
	var current int64
	err = tx.QueryRowContext(ctx, `SELECT revision FROM batch_factory_v12_books WHERE id=? AND batch_id=? AND owner_username=? FOR UPDATE`, bookID, batchID, owner).Scan(&current)
	if errors.Is(err, sql.ErrNoRows) {
		return Book{}, ErrNotFound
	}
	if err != nil {
		return Book{}, err
	}
	if current != input.ExpectedRevision {
		return Book{}, ErrConflict
	}
	var source, currentTitle, sourceBookID sql.NullString
	var rawMetadata []byte
	err = tx.QueryRowContext(ctx, `SELECT source_text,title,source_book_id,source_metadata_json FROM batch_factory_v12_book_records WHERE book_id=? FOR UPDATE`, bookID).Scan(&source, &currentTitle, &sourceBookID, &rawMetadata)
	if errors.Is(err, sql.ErrNoRows) {
		return Book{}, ErrNotFound
	}
	if err != nil {
		return Book{}, err
	}
	if strings.TrimSpace(source.String) != "" && !input.ReplaceSource {
		return Book{}, ErrConflict
	}
	metadata := decodeSourceMetadata(rawMetadata)
	if metadata == nil {
		metadata = map[string]any{}
	}
	for key, value := range input.SourceMetadata {
		metadata[key] = value
	}
	encodedMetadata, err := json.Marshal(metadata)
	if err != nil {
		return Book{}, ErrInvalid
	}
	nextTitle := currentTitle.String
	if title := strings.TrimSpace(input.SourceTitle); title != "" && (input.ReplaceTitle || shouldReplaceGeneratedBookTitle(nextTitle, sourceBookID.String, title)) {
		nextTitle = title
	}
	nextSourceBookID := sourceBookID.String
	if candidate := strings.TrimSpace(input.SourceBookID); candidate != "" {
		nextSourceBookID = candidate
	}
	nextSourceText := source.String
	if validSourceText || input.ReplaceSource {
		nextSourceText = sourceText
	}
	if !validRawSourceText && validSourceText {
		rawSourceText = nextSourceText
	}
	// See UpdateBookMetadata: this column is JSON, not a binary blob.
	if _, err = tx.ExecContext(ctx, `UPDATE batch_factory_v12_book_records SET title=?,source_book_id=?,source_text=?,txt_text=?,source_metadata_json=? WHERE book_id=?`, nextTitle, nullableString(nextSourceBookID), nextSourceText, rawSourceText, string(encodedMetadata), bookID); err != nil {
		return Book{}, err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE batch_factory_v12_books SET revision=revision+1,updated_at=? WHERE id=? AND batch_id=? AND owner_username=?`, time.Now().UTC(), bookID, batchID, owner); err != nil {
		return Book{}, err
	}
	if err = tx.Commit(); err != nil {
		return Book{}, err
	}
	batch, err := s.GetBatch(ctx, owner, batchID)
	if err != nil {
		return Book{}, err
	}
	for _, book := range batch.Books {
		if book.ID == bookID {
			return book, nil
		}
	}
	return Book{}, ErrNotFound
}

func decodeSourceMetadata(raw []byte) map[string]any {
	if len(raw) == 0 {
		return nil
	}
	var metadata map[string]any
	if err := json.Unmarshal(raw, &metadata); err != nil {
		return nil
	}
	return metadata
}

func nullableString(v string) any {
	if v == "" {
		return nil
	}
	return v
}
func nullableFloat(v float64) any {
	if v == 0 {
		return nil
	}
	return v
}

func (s *MySQLStore) ListBatches(ctx context.Context, owner string) ([]Batch, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT b.id,r.title,COALESCE(r.source_intake_id,''),b.revision,b.created_at,b.updated_at FROM batch_factory_v12_batches b JOIN batch_factory_v12_batch_records r ON r.batch_id=b.id WHERE b.owner_username=? AND EXISTS (SELECT 1 FROM batch_factory_v12_books vb WHERE vb.batch_id=b.id AND vb.owner_username=b.owner_username AND vb.hidden_at IS NULL) ORDER BY b.created_at DESC,b.id DESC`, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Batch{}
	byID := map[string]int{}
	for rows.Next() {
		var batch Batch
		if err := rows.Scan(&batch.ID, &batch.Title, &batch.SourceIntakeID, &batch.Revision, &batch.CreatedAt, &batch.UpdatedAt); err != nil {
			return nil, err
		}
		batch.Books = []Book{}
		byID[batch.ID] = len(out)
		out = append(out, batch)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	rows.Close()
	if len(out) == 0 {
		return out, nil
	}

	books, err := s.db.QueryContext(ctx, `SELECT b.batch_id,b.id,r.title,COALESCE(r.source_book_id,''),COALESCE(r.platform,''),b.revision FROM batch_factory_v12_books b JOIN batch_factory_v12_book_records r ON r.book_id=b.id WHERE b.owner_username=? AND b.hidden_at IS NULL ORDER BY b.batch_id,b.ordinal,b.id`, owner)
	if err != nil {
		return nil, err
	}
	defer books.Close()
	for books.Next() {
		var batchID string
		var book Book
		if err := books.Scan(&batchID, &book.ID, &book.Title, &book.BookID, &book.Platform, &book.Revision); err != nil {
			return nil, err
		}
		index, ok := byID[batchID]
		if !ok {
			continue
		}
		book.BatchID = batchID
		if book.BookID == "" {
			book.BookID = book.ID
		}
		book.Videos = []Video{}
		out[index].Books = append(out[index].Books, book)
	}
	if err := books.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// ListBatchSummaryIndex returns only the fields rendered by the project-card
// list. It deliberately skips source text, settings patches, director
// revisions, assets and videos; those belong to the single-batch detail path.
// Keeping this as two bulk queries prevents the library page from turning N
// historical batches into thousands of SQL statements.
func (s *MySQLStore) ListBatchSummaryIndex(ctx context.Context, owner string) ([]Batch, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT b.id,r.title,b.revision,b.created_at,b.updated_at FROM batch_factory_v12_batches b JOIN batch_factory_v12_batch_records r ON r.batch_id=b.id WHERE b.owner_username=? AND EXISTS (SELECT 1 FROM batch_factory_v12_books vb WHERE vb.batch_id=b.id AND vb.owner_username=b.owner_username AND vb.hidden_at IS NULL) ORDER BY b.created_at DESC,b.id DESC`, owner)
	if err != nil {
		return nil, err
	}
	batches := []Batch{}
	batchIndexByID := map[string]int{}
	for rows.Next() {
		var batch Batch
		if err := rows.Scan(&batch.ID, &batch.Title, &batch.Revision, &batch.CreatedAt, &batch.UpdatedAt); err != nil {
			rows.Close()
			return nil, err
		}
		batch.Books = []Book{}
		batches = append(batches, batch)
		batchIndexByID[batch.ID] = len(batches) - 1
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()

	bookRows, err := s.db.QueryContext(ctx, `SELECT b.batch_id,b.id,COALESCE(r.source_book_id,''),COALESCE(r.title,''),COALESCE(r.platform,'') FROM batch_factory_v12_books b JOIN batch_factory_v12_book_records r ON r.book_id=b.id WHERE b.owner_username=? AND b.hidden_at IS NULL ORDER BY b.batch_id,b.ordinal,b.id`, owner)
	if err != nil {
		return nil, err
	}
	defer bookRows.Close()
	for bookRows.Next() {
		var book Book
		if err := bookRows.Scan(&book.BatchID, &book.ID, &book.BookID, &book.Title, &book.Platform); err != nil {
			return nil, err
		}
		if index, ok := batchIndexByID[book.BatchID]; ok {
			batches[index].Books = append(batches[index].Books, book)
		}
	}
	if err := bookRows.Err(); err != nil {
		return nil, err
	}
	bookRows.Close()

	// A project card needs at most one already-generated cover. Reading that
	// choice here in one bulk query prevents the browser from polling full
	// production and merge histories for every old project after each reload.
	coverRows, err := s.db.QueryContext(ctx, `SELECT j.batch_id,j.output_url FROM batch_factory_v12_merge_jobs j LEFT JOIN batch_factory_v12_books b ON b.id=j.book_id AND b.owner_username=j.owner_username WHERE j.owner_username=? AND j.status='succeeded' AND COALESCE(j.output_url,'')<>'' AND (b.id IS NULL OR b.hidden_at IS NULL) ORDER BY j.batch_id,COALESCE(b.ordinal,2147483647),j.updated_at DESC,j.id DESC`, owner)
	if err != nil {
		return nil, err
	}
	defer coverRows.Close()
	seenCover := map[string]bool{}
	for coverRows.Next() {
		var batchID, outputURL string
		if err := coverRows.Scan(&batchID, &outputURL); err != nil {
			return nil, err
		}
		artifactID := projectCoverArtifactID(batchID, outputURL)
		if artifactID == "" {
			continue
		}
		if seenCover[batchID] {
			continue
		}
		seenCover[batchID] = true
		if index, ok := batchIndexByID[batchID]; ok {
			batches[index].ProjectCoverArtifactID = artifactID
		}
	}
	if err := coverRows.Err(); err != nil {
		return nil, err
	}

	// Not every batch has been merged into a final video yet. Reuse the first
	// completed storyboard video as its project-card cover so the library can
	// show real work without issuing one production-status request per card.
	shotRows, err := s.db.QueryContext(ctx, `SELECT j.batch_id,t.media_url FROM batch_factory_v12_production_tasks t JOIN batch_factory_v12_production_jobs j ON j.id=t.job_id AND j.owner_username=t.owner_username LEFT JOIN batch_factory_v12_books b ON b.id=j.book_id AND b.owner_username=j.owner_username LEFT JOIN batch_factory_v12_videos v ON v.id=t.video_id AND v.owner_username=t.owner_username WHERE t.owner_username=? AND t.status='succeeded' AND COALESCE(t.media_url,'')<>'' ORDER BY j.batch_id,COALESCE(b.ordinal,2147483647),COALESCE(v.ordinal,2147483647),t.updated_at DESC,t.id DESC`, owner)
	if err != nil {
		return nil, err
	}
	defer shotRows.Close()
	seenShot := map[string]bool{}
	for shotRows.Next() {
		var batchID, mediaURL string
		if err := shotRows.Scan(&batchID, &mediaURL); err != nil {
			return nil, err
		}
		if seenShot[batchID] {
			continue
		}
		seenShot[batchID] = true
		if index, ok := batchIndexByID[batchID]; ok {
			batches[index].ProjectCoverMediaURL = mediaURL
		}
	}
	if err := shotRows.Err(); err != nil {
		return nil, err
	}

	// 漫画项目没有分镜成片，用首个漫画页主图作为卡片封面。封面必须来自一次
	// 批量查询，否则项目库会退化成按项目逐个发图片请求。
	comicRows, err := s.db.QueryContext(ctx, `SELECT p.batch_id,p.book_id,p.id,i.id FROM batch_factory_v12_comic_page_images i JOIN batch_factory_v12_comic_pages p ON p.id=i.comic_page_id AND p.owner_username=i.owner_username LEFT JOIN batch_factory_v12_books b ON b.id=p.book_id AND b.owner_username=p.owner_username LEFT JOIN batch_factory_v12_videos v ON v.id=p.video_id AND v.owner_username=p.owner_username WHERE i.owner_username=? AND i.is_primary=TRUE ORDER BY p.batch_id,COALESCE(b.ordinal,2147483647),COALESCE(v.ordinal,2147483647),i.created_at DESC,i.id DESC`, owner)
	if err != nil {
		return nil, err
	}
	seenComic := map[string]bool{}
	for comicRows.Next() {
		var batchID, bookID, comicPageID, imageID string
		if err := comicRows.Scan(&batchID, &bookID, &comicPageID, &imageID); err != nil {
			comicRows.Close()
			return nil, err
		}
		if seenComic[batchID] {
			continue
		}
		seenComic[batchID] = true
		if index, ok := batchIndexByID[batchID]; ok && batches[index].ProjectCoverArtifactID == "" && batches[index].ProjectCoverMediaURL == "" {
			batches[index].ProjectCoverImageURL = projectCoverComicPageImageURL(batchID, bookID, comicPageID, imageID)
		}
	}
	if err := comicRows.Err(); err != nil {
		comicRows.Close()
		return nil, err
	}
	comicRows.Close()

	// 还没生成漫画页时，退到书籍资产主图，避免项目卡片长期是空白色块。
	assetRows, err := s.db.QueryContext(ctx, `SELECT a.batch_id,a.book_id,a.id,i.id FROM batch_factory_v12_book_asset_images i JOIN batch_factory_v12_book_assets a ON a.id=i.asset_id AND a.owner_username=i.owner_username LEFT JOIN batch_factory_v12_books b ON b.id=a.book_id AND b.owner_username=a.owner_username WHERE i.owner_username=? AND i.is_primary=TRUE ORDER BY a.batch_id,COALESCE(b.ordinal,2147483647),i.created_at DESC,i.id DESC`, owner)
	if err != nil {
		return nil, err
	}
	seenAsset := map[string]bool{}
	for assetRows.Next() {
		var batchID, bookID, assetID, imageID string
		if err := assetRows.Scan(&batchID, &bookID, &assetID, &imageID); err != nil {
			assetRows.Close()
			return nil, err
		}
		if seenAsset[batchID] {
			continue
		}
		seenAsset[batchID] = true
		if index, ok := batchIndexByID[batchID]; ok && batches[index].ProjectCoverImageURL == "" && batches[index].ProjectCoverArtifactID == "" && batches[index].ProjectCoverMediaURL == "" {
			batches[index].ProjectCoverImageURL = projectCoverAssetImageURL(batchID, bookID, assetID, imageID)
		}
	}
	if err := assetRows.Err(); err != nil {
		assetRows.Close()
		return nil, err
	}
	assetRows.Close()

	// 项目卡片上的“漫画/视频”标记由批次或书籍设置里的 productionMode 决定。
	modeRows, err := s.db.QueryContext(ctx, `SELECT batch_id,patch_json FROM batch_factory_v12_settings_patches WHERE owner_username=? AND scope_type IN ('batch','book')`, owner)
	if err != nil {
		return nil, err
	}
	seenComicMode := map[string]bool{}
	for modeRows.Next() {
		var batchID, patchJSON string
		if err := modeRows.Scan(&batchID, &patchJSON); err != nil {
			modeRows.Close()
			return nil, err
		}
		var patch struct {
			ProductionMode string `json:"productionMode"`
		}
		if json.Unmarshal([]byte(patchJSON), &patch) == nil && strings.EqualFold(patch.ProductionMode, "comic") {
			seenComicMode[batchID] = true
		}
	}
	if err := modeRows.Err(); err != nil {
		modeRows.Close()
		return nil, err
	}
	modeRows.Close()
	for batchID := range seenComicMode {
		if index, ok := batchIndexByID[batchID]; ok {
			batches[index].ProductionMode = "comic"
		}
	}
	for index := range batches {
		if batches[index].ProductionMode == "" {
			batches[index].ProductionMode = "video"
		}
	}
	return batches, nil
}

// ListShuihuoProjectBatchIDs returns the batches whose batch-scope settings
// mark them as Shuihuo projects. The V12 batch list used to call GetBatch for
// every batch only to read this single flag, which hydrated settings, drafts,
// assets and videos N times. One patch query keeps the JSON contract unchanged.
func (s *MySQLStore) ListShuihuoProjectBatchIDs(ctx context.Context, owner string) (map[string]bool, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT batch_id,patch_json FROM batch_factory_v12_settings_patches WHERE owner_username=? AND scope_type='batch'`, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	flags := map[string]bool{}
	for rows.Next() {
		var batchID, patchJSON string
		if err := rows.Scan(&batchID, &patchJSON); err != nil {
			return nil, err
		}
		var patch struct {
			ShuihuoProject json.RawMessage `json:"shuihuoProject"`
		}
		if json.Unmarshal([]byte(patchJSON), &patch) == nil && strings.TrimSpace(string(patch.ShuihuoProject)) == "true" {
			flags[batchID] = true
		}
	}
	return flags, rows.Err()
}

func projectCoverComicPageImageURL(batchID, bookID, comicPageID, imageID string) string {
	return "/api/batch-factory/v12/batches/" + url.PathEscape(batchID) + "/books/" + url.PathEscape(bookID) + "/comic-pages/" + url.PathEscape(comicPageID) + "/images/" + url.PathEscape(imageID) + "/content"
}

func projectCoverAssetImageURL(batchID, bookID, assetID, imageID string) string {
	return "/api/batch-factory/v12/batches/" + url.PathEscape(batchID) + "/books/" + url.PathEscape(bookID) + "/assets/" + url.PathEscape(assetID) + "/images/" + url.PathEscape(imageID) + "/content"
}

func projectCoverArtifactID(batchID, outputURL string) string {
	for _, version := range []string{"v11", "v12"} {
		prefix := "/api/batch-factory/" + version + "/batches/" + batchID + "/merge-media/"
		if artifactID, ok := strings.CutPrefix(strings.TrimSpace(outputURL), prefix); ok && localMergedArtifactID.MatchString(artifactID) {
			return artifactID
		}
	}
	return ""
}

func (s *MySQLStore) GetBatch(ctx context.Context, owner, id string) (Batch, error) {
	return loadBatch(ctx, s.db, owner, id)
}

// BatchOwner 只按批次 id 读取归属账号，不做 owner 过滤。
// 详情接口用它把“批次属于其他账号”从裸 404 里区分出来，返回可读的 403。
func (s *MySQLStore) BatchOwner(ctx context.Context, id string) (string, error) {
	var owner string
	err := s.db.QueryRowContext(ctx, `SELECT b.owner_username FROM batch_factory_v12_batches b JOIN batch_factory_v12_batch_records r ON r.batch_id=b.id WHERE b.id=?`, id).Scan(&owner)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	return owner, nil
}

// GetBatchRuntimeIndex returns only the identity data needed by the live
// runtime summary. The full GetBatch response hydrates every settings patch,
// director revision, asset and video; doing that every few seconds turns a
// status poll into thousands of prepared statements on mature batches.
func (s *MySQLStore) GetBatchRuntimeIndex(ctx context.Context, owner, id string) (Batch, error) {
	var batch Batch
	err := s.db.QueryRowContext(ctx, `SELECT b.id,b.revision,b.created_at,b.updated_at FROM batch_factory_v12_batches b WHERE b.id=? AND b.owner_username=?`, id, owner).
		Scan(&batch.ID, &batch.Revision, &batch.CreatedAt, &batch.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Batch{}, ErrNotFound
	}
	if err != nil {
		return Batch{}, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id,revision FROM batch_factory_v12_books WHERE batch_id=? AND owner_username=? AND hidden_at IS NULL ORDER BY ordinal,id`, id, owner)
	if err != nil {
		return Batch{}, err
	}
	defer rows.Close()
	batch.Books = []Book{}
	for rows.Next() {
		var book Book
		if err := rows.Scan(&book.ID, &book.Revision); err != nil {
			return Batch{}, err
		}
		book.BatchID = batch.ID
		videoRows, err := s.db.QueryContext(ctx, `SELECT id,ordinal FROM batch_factory_v12_videos WHERE batch_id=? AND book_id=? AND owner_username=? AND compatibility_state='active' ORDER BY ordinal,id`, id, book.ID, owner)
		if err != nil {
			return Batch{}, err
		}
		for videoRows.Next() {
			var videoID string
			var ordinal int
			if err := videoRows.Scan(&videoID, &ordinal); err != nil {
				videoRows.Close()
				return Batch{}, err
			}
			book.Videos = append(book.Videos, Video{ID: videoID, BatchID: batch.ID, BookID: book.ID, Ordinal: ordinal, CompatibilityState: "active"})
		}
		if err := videoRows.Err(); err != nil {
			videoRows.Close()
			return Batch{}, err
		}
		if err := videoRows.Close(); err != nil {
			return Batch{}, err
		}
		// 轮询同样需要历史 ordinal 与可播信息，否则前端在状态刷新后会丢掉
		// 「上一版」的挂载点，或只能列出槽位却无法播放。
		historyRows, err := s.db.QueryContext(ctx, `SELECT v.id,v.ordinal,v.compatibility_state,COALESCE(t.id,''),COALESCE(t.media_url,''),COALESCE(t.provider,''),COALESCE(t.status,''),t.created_at FROM batch_factory_v12_videos v LEFT JOIN batch_factory_v12_production_tasks t ON t.video_id=v.id AND t.owner_username=v.owner_username AND t.status='succeeded' AND COALESCE(t.media_url,'')<>'' AND NOT EXISTS (SELECT 1 FROM batch_factory_v12_hidden_production_tasks h WHERE h.task_id=t.id AND h.owner_username=t.owner_username) WHERE v.batch_id=? AND v.book_id=? AND v.owner_username=? AND v.compatibility_state<>'active' ORDER BY v.ordinal,v.updated_at DESC,v.id,t.updated_at DESC,t.created_at DESC,t.id`, id, book.ID, owner)
		if err != nil {
			return Batch{}, err
		}
		historySeen := map[string]bool{}
		for historyRows.Next() {
			var video Video
			var taskID, mediaURL, provider, taskStatus string
			var generatedAt sql.NullTime
			if err := historyRows.Scan(&video.ID, &video.Ordinal, &video.CompatibilityState, &taskID, &mediaURL, &provider, &taskStatus, &generatedAt); err != nil {
				historyRows.Close()
				return Batch{}, err
			}
			if historySeen[video.ID] {
				continue
			}
			historySeen[video.ID] = true
			video.BatchID, video.BookID = batch.ID, book.ID
			video.ProductionTaskID, video.MediaURL, video.Provider = taskID, mediaURL, provider
			video.Succeeded = taskStatus == string(ProductionSucceeded)
			if generatedAt.Valid {
				video.GeneratedAt = generatedAt.Time
			}
			book.VideoHistory = append(book.VideoHistory, video)
		}
		if err := historyRows.Err(); err != nil {
			historyRows.Close()
			return Batch{}, err
		}
		historyRows.Close()
		batch.Books = append(batch.Books, book)
	}
	if err := rows.Err(); err != nil {
		return Batch{}, err
	}
	return batch, nil
}

// ListBatchRecoveryIndex returns the smallest payload needed by the periodic
// giant-material recovery sweep. In particular it does not hydrate videos,
// director revisions or assets for every historical batch on every sweep.
func (s *MySQLStore) ListBatchRecoveryIndex(ctx context.Context, owner string) ([]Batch, error) {
	// 两步查询：source_text / 草稿正文可能是整本小说，直接参与 ORDER BY 的
	// filesort 会撑爆 sort_buffer_size（MySQL ERROR 1038）。先只对 book id
	// 排序，再按 id 分批读取负载，顺序仍是 createdAt DESC, id DESC, ordinal, id。
	const idQuery = `SELECT b.id,b.batch_id FROM batch_factory_v12_books b JOIN batch_factory_v12_book_records r ON r.book_id=b.id JOIN batch_factory_v12_batches ba ON ba.id=b.batch_id AND ba.owner_username=b.owner_username WHERE b.owner_username=? AND b.hidden_at IS NULL AND JSON_UNQUOTE(JSON_EXTRACT(r.source_metadata_json,'$.sourceMode'))='giant_material' ORDER BY ba.created_at DESC,ba.id DESC,b.ordinal,b.id`
	idRows, err := s.db.QueryContext(ctx, idQuery, owner)
	if err != nil {
		return nil, err
	}
	type recoveryKey struct{ bookID, batchID string }
	keys := []recoveryKey{}
	for idRows.Next() {
		var key recoveryKey
		if err := idRows.Scan(&key.bookID, &key.batchID); err != nil {
			idRows.Close()
			return nil, err
		}
		keys = append(keys, key)
	}
	if err := idRows.Err(); err != nil {
		idRows.Close()
		return nil, err
	}
	idRows.Close()

	out := []Batch{}
	batchIndexes := map[string]int{}
	decodePatch := func(raw []byte) (SettingsPatch, error) {
		patch := SettingsPatch{}
		if len(raw) == 0 {
			return patch, nil
		}
		if err := json.Unmarshal(raw, &patch); err != nil {
			return nil, err
		}
		return normalizeLegacyNestedPatch(patch), nil
	}
	type recoveryRow struct {
		batchID                     string
		batchRevision, bookRevision int64
		createdAt, updatedAt        time.Time
		batchPatchRaw, metadataRaw  []byte
		bookPatchRaw                []byte
		bookID, sourceText, working string
	}
	const recoveryPageSize = 200
	const payloadQuery = `SELECT ba.id,ba.revision,ba.created_at,ba.updated_at,COALESCE(bsp.patch_json,JSON_OBJECT()),b.id,b.revision,COALESCE(r.source_text,''),r.source_metadata_json,COALESCE(d.content,''),COALESCE(ksp.patch_json,JSON_OBJECT()) FROM batch_factory_v12_books b JOIN batch_factory_v12_book_records r ON r.book_id=b.id JOIN batch_factory_v12_batches ba ON ba.id=b.batch_id AND ba.owner_username=b.owner_username LEFT JOIN batch_factory_v12_settings_patches bsp ON bsp.scope_type='batch' AND bsp.scope_id=ba.id AND bsp.owner_username=ba.owner_username LEFT JOIN batch_factory_v12_settings_patches ksp ON ksp.scope_type='book' AND ksp.scope_id=b.id AND ksp.owner_username=b.owner_username LEFT JOIN batch_factory_v12_drafts d ON d.owner_username=b.owner_username AND d.draft_key=CONCAT('working-front:',b.id) AND d.kind='working-front-content' AND d.scope=ba.id WHERE b.owner_username=? AND b.id IN (`
	for start := 0; start < len(keys); start += recoveryPageSize {
		end := start + recoveryPageSize
		if end > len(keys) {
			end = len(keys)
		}
		page := keys[start:end]
		placeholders := make([]string, 0, len(page))
		args := make([]any, 0, len(page)+1)
		args = append(args, owner)
		for _, key := range page {
			placeholders = append(placeholders, "?")
			args = append(args, key.bookID)
		}
		rows, err := s.db.QueryContext(ctx, payloadQuery+strings.Join(placeholders, ",")+")", args...)
		if err != nil {
			return nil, err
		}
		rowsByBook := make(map[string]recoveryRow, len(page))
		for rows.Next() {
			var row recoveryRow
			if err := rows.Scan(&row.batchID, &row.batchRevision, &row.createdAt, &row.updatedAt, &row.batchPatchRaw, &row.bookID, &row.bookRevision, &row.sourceText, &row.metadataRaw, &row.working, &row.bookPatchRaw); err != nil {
				rows.Close()
				return nil, err
			}
			rowsByBook[row.bookID] = row
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, err
		}
		rows.Close()
		for _, key := range page {
			row, ok := rowsByBook[key.bookID]
			if !ok {
				continue
			}
			index, ok := batchIndexes[row.batchID]
			if !ok {
				batchPatch, err := decodePatch(row.batchPatchRaw)
				if err != nil {
					return nil, err
				}
				index = len(out)
				batchIndexes[row.batchID] = index
				out = append(out, Batch{
					ID: row.batchID, Revision: row.batchRevision, CreatedAt: row.createdAt, UpdatedAt: row.updatedAt,
					SettingsState: SettingsState{Patch: batchPatch, Revision: row.batchRevision}, Books: []Book{},
				})
			}
			bookPatch, err := decodePatch(row.bookPatchRaw)
			if err != nil {
				return nil, err
			}
			out[index].Books = append(out[index].Books, Book{
				ID: row.bookID, BatchID: row.batchID, Revision: row.bookRevision, SourceText: row.sourceText,
				WorkingFrontContent: row.working, SourceMetadata: decodeSourceMetadata(row.metadataRaw),
				SettingsState: SettingsState{Patch: bookPatch, Revision: row.bookRevision},
			})
		}
	}
	return out, nil
}

func (s *MySQLStore) DeleteBook(ctx context.Context, owner, batchID, bookID string) error {
	return s.DeleteBookWithRetention(ctx, owner, batchID, bookID, DefaultLocalMergedArtifactRetentionDays)
}

func (s *MySQLStore) DeleteBookWithRetention(ctx context.Context, owner, batchID, bookID string, retentionDays int) error {
	if strings.TrimSpace(owner) == "" || strings.TrimSpace(batchID) == "" || strings.TrimSpace(bookID) == "" {
		return ErrInvalid
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var one int
	if err = tx.QueryRowContext(ctx, `SELECT 1 FROM batch_factory_v12_books WHERE id=? AND batch_id=? AND owner_username=? FOR UPDATE`, bookID, batchID, owner).Scan(&one); errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	} else if err != nil {
		return err
	}
	purgeAfter := time.Now().UTC().AddDate(0, 0, normalizeLocalMergedArtifactRetentionDays(retentionDays))
	if err = queueLocalMergedArtifactPurges(ctx, tx, owner, batchID, bookID, purgeAfter); err != nil {
		return err
	}
	if err = deleteBookTx(ctx, tx, owner, batchID, bookID); err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `UPDATE batch_factory_v12_batches SET revision=revision+1,updated_at=CURRENT_TIMESTAMP(6) WHERE id=? AND owner_username=?`, batchID, owner)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s *MySQLStore) DeleteBatch(ctx context.Context, owner, batchID string) error {
	return s.DeleteBatchWithRetention(ctx, owner, batchID, DefaultLocalMergedArtifactRetentionDays)
}

func (s *MySQLStore) DeleteBatchWithRetention(ctx context.Context, owner, batchID string, retentionDays int) error {
	if strings.TrimSpace(owner) == "" || strings.TrimSpace(batchID) == "" {
		return ErrInvalid
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var one int
	if err = tx.QueryRowContext(ctx, `SELECT 1 FROM batch_factory_v12_batches WHERE id=? AND owner_username=? FOR UPDATE`, batchID, owner).Scan(&one); errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	} else if err != nil {
		return err
	}
	purgeAfter := time.Now().UTC().AddDate(0, 0, normalizeLocalMergedArtifactRetentionDays(retentionDays))
	if err = queueLocalMergedArtifactPurges(ctx, tx, owner, batchID, "", purgeAfter); err != nil {
		return err
	}
	rows, err := tx.QueryContext(ctx, `SELECT id FROM batch_factory_v12_books WHERE batch_id=? AND owner_username=? FOR UPDATE`, batchID, owner)
	if err != nil {
		return err
	}
	var bookIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		bookIDs = append(bookIDs, id)
	}
	if err := rows.Close(); err != nil {
		return err
	}
	for _, bookID := range bookIDs {
		if err := deleteBookTx(ctx, tx, owner, batchID, bookID); err != nil {
			return err
		}
	}
	if err := deleteBatchTx(ctx, tx, owner, batchID); err != nil {
		return err
	}
	return tx.Commit()
}

var deleteBatchStmts = []string{
	`DELETE a FROM batch_factory_v12_external_audits a JOIN batch_factory_v12_external_intents i ON i.id=a.intent_id WHERE i.owner_username=? AND i.batch_id=?`,
	`DELETE FROM batch_factory_v12_external_intents WHERE owner_username=? AND batch_id=?`,
	`DELETE FROM batch_factory_v12_merge_sources WHERE owner_username=? AND job_id IN (SELECT id FROM batch_factory_v12_merge_jobs WHERE owner_username=? AND batch_id=?)`,
	`DELETE FROM batch_factory_v12_merge_jobs WHERE owner_username=? AND batch_id=?`,
	`DELETE FROM batch_factory_v12_settings_patches WHERE owner_username=? AND batch_id=?`,
	`DELETE FROM batch_factory_v12_config_snapshots WHERE owner_username=? AND batch_id=?`,
	`DELETE FROM batch_factory_v12_drafts WHERE owner_username=? AND scope=?`,
	`DELETE FROM batch_factory_v12_batch_records WHERE batch_id=?`,
	`DELETE FROM batch_factory_v12_batches WHERE id=? AND owner_username=?`,
}

func deleteBatchTx(ctx context.Context, tx *sql.Tx, owner, batchID string) error {
	for _, stmt := range deleteBatchStmts {
		if err := execDeleteBatchStmt(ctx, tx, stmt, owner, batchID); err != nil {
			return err
		}
	}
	return nil
}

func execDeleteBatchStmt(ctx context.Context, tx *sql.Tx, stmt, owner, batchID string) error {
	args := []any{owner, batchID}
	if strings.Contains(stmt, "merge_sources") {
		args = []any{owner, owner, batchID}
	}
	if strings.Contains(stmt, "batch_records") {
		args = []any{batchID}
	}
	if strings.Contains(stmt, "batches WHERE") {
		args = []any{batchID, owner}
	}
	_, err := tx.ExecContext(ctx, stmt, args...)
	return err
}

type deleteBookStmt struct {
	query string
	args  func(owner, batchID, bookID string) []any
}

func deleteBookDefaultArgs(owner, batchID, bookID string) []any {
	return []any{owner, batchID, bookID}
}

// deleteBookStmts pairs each DELETE statement with exactly the arguments it
// needs. Keep them next to each other so the number of placeholders can never
// drift away from the number of arguments again.
var deleteBookStmts = []deleteBookStmt{
	{query: `DELETE a FROM batch_factory_v12_external_audits a JOIN batch_factory_v12_external_intents i ON i.id=a.intent_id WHERE i.owner_username=? AND i.batch_id=? AND i.book_id=?`, args: deleteBookDefaultArgs},
	{query: `DELETE FROM batch_factory_v12_external_intents WHERE owner_username=? AND batch_id=? AND book_id=?`, args: deleteBookDefaultArgs},
	{query: `DELETE FROM batch_factory_v12_hidden_production_tasks WHERE owner_username=? AND batch_id=? AND book_id=?`, args: deleteBookDefaultArgs},
	{query: `DELETE e FROM batch_factory_v12_production_events e JOIN batch_factory_v12_production_jobs j ON j.id=e.job_id WHERE j.owner_username=? AND j.batch_id=? AND j.book_id=?`, args: deleteBookDefaultArgs},
	{query: `DELETE t FROM batch_factory_v12_production_tasks t JOIN batch_factory_v12_production_jobs j ON j.id=t.job_id WHERE j.owner_username=? AND j.batch_id=? AND j.book_id=?`, args: deleteBookDefaultArgs},
	{query: `DELETE FROM batch_factory_v12_production_jobs WHERE owner_username=? AND batch_id=? AND book_id=?`, args: deleteBookDefaultArgs},
	{query: `DELETE FROM batch_factory_v12_book_stage_runs WHERE owner_username=? AND batch_id=? AND book_id=?`, args: deleteBookDefaultArgs},
	{query: `DELETE FROM batch_factory_v12_settings_patches WHERE owner_username=? AND batch_id=? AND book_id=?`, args: deleteBookDefaultArgs},
	// V12 H3 records have restrictive foreign keys back to both the book and
	// its director revision. They must be removed before their V11 parents;
	// otherwise deleting an H3-enabled batch fails and the transaction rolls
	// back with a generic internal error.
	{query: `DELETE FROM batch_factory_v12_video_compilations WHERE owner_username=? AND batch_id=? AND book_id=?`, args: deleteBookDefaultArgs},
	{query: `DELETE FROM batch_factory_v12_canonical_timelines WHERE owner_username=? AND batch_id=? AND book_id=?`, args: deleteBookDefaultArgs},
	{query: `DELETE FROM batch_factory_v12_audio_measurements WHERE owner_username=? AND batch_id=? AND book_id=?`, args: deleteBookDefaultArgs},
	{query: `DELETE o FROM batch_factory_v12_orphaned_overrides o JOIN batch_factory_v12_director_revisions d ON d.id=o.director_revision_id WHERE d.owner_username=? AND d.batch_id=? AND d.book_id=?`, args: deleteBookDefaultArgs},
	{query: `DELETE l FROM batch_factory_v12_director_video_links l JOIN batch_factory_v12_director_revisions d ON d.id=l.director_revision_id WHERE d.owner_username=? AND d.batch_id=? AND d.book_id=?`, args: deleteBookDefaultArgs},
	{query: `DELETE FROM batch_factory_v12_director_revisions WHERE owner_username=? AND batch_id=? AND book_id=?`, args: deleteBookDefaultArgs},
	{query: `DELETE FROM batch_factory_v12_hook_revisions WHERE owner_username=? AND batch_id=? AND book_id=?`, args: deleteBookDefaultArgs},
	{query: `DELETE i FROM batch_factory_v12_book_asset_images i JOIN batch_factory_v12_book_assets a ON a.id=i.asset_id WHERE a.owner_username=? AND a.batch_id=? AND a.book_id=?`, args: deleteBookDefaultArgs},
	{query: `DELETE FROM batch_factory_v12_book_assets WHERE owner_username=? AND batch_id=? AND book_id=?`, args: deleteBookDefaultArgs},
	{
		query: `DELETE FROM batch_factory_v12_drafts WHERE owner_username=? AND draft_key=? AND scope=?`,
		args:  func(owner, batchID, bookID string) []any { return []any{owner, "working-front:" + bookID, batchID} },
	},
	{query: `DELETE FROM batch_factory_v12_config_snapshots WHERE owner_username=? AND batch_id=? AND book_id=?`, args: deleteBookDefaultArgs},
	{query: `DELETE FROM batch_factory_v12_video_records WHERE video_id IN (SELECT id FROM batch_factory_v12_videos WHERE owner_username=? AND batch_id=? AND book_id=?)`, args: deleteBookDefaultArgs},
	{query: `DELETE FROM batch_factory_v12_videos WHERE owner_username=? AND batch_id=? AND book_id=?`, args: deleteBookDefaultArgs},
	{
		query: `DELETE FROM batch_factory_v12_book_records WHERE book_id=?`,
		args:  func(_, _, bookID string) []any { return []any{bookID} },
	},
	{
		query: `DELETE FROM batch_factory_v12_books WHERE id=? AND batch_id=? AND owner_username=?`,
		args:  func(owner, batchID, bookID string) []any { return []any{bookID, batchID, owner} },
	},
}

func deleteBookTx(ctx context.Context, tx *sql.Tx, owner, batchID, bookID string) error {
	for _, stmt := range deleteBookStmts {
		if _, err := tx.ExecContext(ctx, stmt.query, stmt.args(owner, batchID, bookID)...); err != nil {
			return err
		}
	}
	return nil
}

type batchQueryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func loadBatch(ctx context.Context, q batchQueryer, owner, id string) (Batch, error) {
	var b Batch
	err := q.QueryRowContext(ctx, `SELECT b.id,r.title,COALESCE(r.source_intake_id,''),b.revision,b.created_at,b.updated_at FROM batch_factory_v12_batches b JOIN batch_factory_v12_batch_records r ON r.batch_id=b.id WHERE b.id=? AND b.owner_username=?`, id, owner).Scan(&b.ID, &b.Title, &b.SourceIntakeID, &b.Revision, &b.CreatedAt, &b.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Batch{}, ErrNotFound
	}
	if err != nil {
		return Batch{}, err
	}
	b.Books = []Book{}
	// 两步查询：source_text/txt_text 存整本小说，参与 ORDER BY 的 filesort 会
	// 触发 MySQL ERROR 1038。先只排序 book id，再按 id 分批读取正文并保持顺序。
	idRows, err := q.QueryContext(ctx, `SELECT id FROM batch_factory_v12_books WHERE batch_id=? AND owner_username=? AND hidden_at IS NULL ORDER BY ordinal,id`, b.ID, owner)
	if err != nil {
		return Batch{}, err
	}
	bookIDs := []string{}
	for idRows.Next() {
		var bookID string
		if err := idRows.Scan(&bookID); err != nil {
			idRows.Close()
			return Batch{}, err
		}
		bookIDs = append(bookIDs, bookID)
	}
	if err := idRows.Err(); err != nil {
		idRows.Close()
		return Batch{}, err
	}
	idRows.Close()
	const bookPageSize = 200
	for start := 0; start < len(bookIDs); start += bookPageSize {
		end := start + bookPageSize
		if end > len(bookIDs) {
			end = len(bookIDs)
		}
		page := bookIDs[start:end]
		placeholders := make([]string, 0, len(page))
		args := make([]any, 0, len(page)+2)
		args = append(args, b.ID, owner)
		for _, bookID := range page {
			placeholders = append(placeholders, "?")
			args = append(args, bookID)
		}
		rows, err := q.QueryContext(ctx, `SELECT b.id,r.title,COALESCE(r.source_text,''),COALESCE(r.source_book_id,''),COALESCE(r.source_task_id,''),COALESCE(r.platform,''),COALESCE(r.txt_text,''),COALESCE(r.txt_file_name,''),r.source_metadata_json,b.revision FROM batch_factory_v12_books b JOIN batch_factory_v12_book_records r ON r.book_id=b.id WHERE b.batch_id=? AND b.owner_username=? AND b.id IN (`+strings.Join(placeholders, ",")+`)`, args...)
		if err != nil {
			return Batch{}, err
		}
		booksByID := make(map[string]Book, len(page))
		for rows.Next() {
			var book Book
			var metadata []byte
			if err := rows.Scan(&book.ID, &book.Title, &book.SourceText, &book.BookID, &book.SourceTaskID, &book.Platform, &book.TxtText, &book.TxtFileName, &metadata, &book.Revision); err != nil {
				rows.Close()
				return Batch{}, err
			}
			book.SourceMetadata = decodeSourceMetadata(metadata)
			if book.BookID == "" {
				book.BookID = book.ID
			}
			book.BatchID = b.ID
			book.Videos = []Video{}
			booksByID[book.ID] = book
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return Batch{}, err
		}
		rows.Close()
		for _, bookID := range page {
			if book, ok := booksByID[bookID]; ok {
				b.Books = append(b.Books, book)
			}
		}
	}
	// Release the book-list connection before loading per-book drafts. If many
	// clients open the same batch together, holding one rows cursor per request
	// while each request asks the pool for another connection can exhaust the
	// pool and deadlock every authenticated batch read.
	for i := range b.Books {
		var working sql.NullString
		workingErr := q.QueryRowContext(ctx, `SELECT content FROM batch_factory_v12_drafts WHERE owner_username=? AND draft_key=? AND kind='working-front-content' AND scope=?`, owner, "working-front:"+b.Books[i].ID, b.ID).Scan(&working)
		if workingErr != nil && !errors.Is(workingErr, sql.ErrNoRows) {
			return Batch{}, workingErr
		}
		if working.Valid {
			b.Books[i].WorkingFrontContent = working.String
		}
	}
	for i := range b.Books {
		vrows, err := q.QueryContext(ctx, `SELECT v.id,r.label,COALESCE(r.video_prompt,''),COALESCE(r.visual_prompt,''),COALESCE(r.duration_seconds,0),v.ordinal,v.compatibility_state,v.revision FROM batch_factory_v12_videos v JOIN batch_factory_v12_video_records r ON r.video_id=v.id WHERE v.batch_id=? AND v.book_id=? AND v.owner_username=? AND v.compatibility_state='active' ORDER BY v.ordinal,v.id`, b.ID, b.Books[i].ID, owner)
		if err != nil {
			return Batch{}, err
		}
		for vrows.Next() {
			var v Video
			if err := vrows.Scan(&v.ID, &v.Label, &v.VideoPrompt, &v.VisualPrompt, &v.DurationSeconds, &v.Ordinal, &v.CompatibilityState, &v.Revision); err != nil {
				vrows.Close()
				return Batch{}, err
			}
			v.BatchID = b.ID
			v.BookID = b.Books[i].ID
			b.Books[i].Videos = append(b.Books[i].Videos, v)
		}
		if err := vrows.Err(); err != nil {
			vrows.Close()
			return Batch{}, err
		}
		vrows.Close()
		// 轮换会把上一轮视频改成 orphaned，但行和媒体都还在。这里单独取一份
		// 历史记录（含最近成功任务的可播地址），让前端能把旧的成功视频继续
		// 列为「上一版」并直接预览/播放。
		hrows, err := q.QueryContext(ctx, `SELECT v.id,COALESCE(r.label,''),v.ordinal,v.compatibility_state,v.revision,COALESCE(t.id,''),COALESCE(t.media_url,''),COALESCE(t.provider,''),COALESCE(t.status,''),t.created_at FROM batch_factory_v12_videos v LEFT JOIN batch_factory_v12_video_records r ON r.video_id=v.id LEFT JOIN batch_factory_v12_production_tasks t ON t.video_id=v.id AND t.owner_username=v.owner_username AND t.status='succeeded' AND COALESCE(t.media_url,'')<>'' AND NOT EXISTS (SELECT 1 FROM batch_factory_v12_hidden_production_tasks h WHERE h.task_id=t.id AND h.owner_username=t.owner_username) WHERE v.batch_id=? AND v.book_id=? AND v.owner_username=? AND v.compatibility_state<>'active' ORDER BY v.ordinal,v.updated_at DESC,v.id,t.updated_at DESC,t.created_at DESC,t.id`, b.ID, b.Books[i].ID, owner)
		if err != nil {
			return Batch{}, err
		}
		historySeen := map[string]bool{}
		for hrows.Next() {
			var v Video
			var taskID, mediaURL, provider, taskStatus string
			var generatedAt sql.NullTime
			if err := hrows.Scan(&v.ID, &v.Label, &v.Ordinal, &v.CompatibilityState, &v.Revision, &taskID, &mediaURL, &provider, &taskStatus, &generatedAt); err != nil {
				hrows.Close()
				return Batch{}, err
			}
			if historySeen[v.ID] {
				continue
			}
			historySeen[v.ID] = true
			v.BatchID = b.ID
			v.BookID = b.Books[i].ID
			v.ProductionTaskID, v.MediaURL, v.Provider = taskID, mediaURL, provider
			v.Succeeded = taskStatus == string(ProductionSucceeded)
			if generatedAt.Valid {
				v.GeneratedAt = generatedAt.Time
			}
			b.Books[i].VideoHistory = append(b.Books[i].VideoHistory, v)
		}
		if err := hrows.Err(); err != nil {
			hrows.Close()
			return Batch{}, err
		}
		hrows.Close()
	}
	return b, nil
}

func (s *MySQLStore) SaveSettings(ctx context.Context, owner string, ref ScopeRef, update SettingsUpdate) (SettingsResult, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return SettingsResult{}, err
	}
	defer tx.Rollback()
	current, err := lockScopeRevision(ctx, tx, owner, ref)
	if err != nil {
		return SettingsResult{}, err
	}
	if current != update.ExpectedRevision {
		return SettingsResult{}, ErrConflict
	}
	versionID, selectsVersion, err := configVersionIDFromUpdate(ref, update)
	if err != nil {
		return SettingsResult{}, err
	}
	if selectsVersion {
		var config json.RawMessage
		err = tx.QueryRowContext(ctx, `SELECT config_json FROM batch_factory_v12_config_versions WHERE id=? AND (owner_username IS NULL OR owner_username=?)`, versionID, owner).Scan(&config)
		if errors.Is(err, sql.ErrNoRows) {
			return SettingsResult{}, ErrNotFound
		}
		if err != nil {
			return SettingsResult{}, err
		}
		if isAIReasoningPresetConfig(config) {
			return SettingsResult{}, ErrInvalid
		}
	}
	patch, err := loadPatch(ctx, tx, owner, ref)
	if err != nil {
		return SettingsResult{}, err
	}
	if ref.Kind == ScopeVideo {
		var preEditPrompt string
		if err := tx.QueryRowContext(ctx, `SELECT COALESCE(video_prompt,'') FROM batch_factory_v12_video_records WHERE video_id=?`, ref.VideoID).Scan(&preEditPrompt); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return SettingsResult{}, ErrNotFound
			}
			return SettingsResult{}, err
		}
		update = initializeAIOriginalVideoPrompt(update, patch, preEditPrompt)
	}
	patch = ApplySparseUpdate(patch, update)
	next := current + 1
	if err := updateScopeRevision(ctx, tx, owner, ref, next); err != nil {
		return SettingsResult{}, err
	}
	encoded, err := json.Marshal(patch)
	if err != nil {
		return SettingsResult{}, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO batch_factory_v12_settings_patches(scope_type,scope_id,owner_username,batch_id,book_id,video_id,patch_json,revision,updated_at) VALUES(?,?,?,?,?,?,?,?,CURRENT_TIMESTAMP(6)) ON DUPLICATE KEY UPDATE patch_json=VALUES(patch_json),revision=VALUES(revision),updated_at=VALUES(updated_at)`, string(ref.Kind), scopeID(ref), owner, ref.BatchID, nullIfEmpty(ref.BookID), nullIfEmpty(ref.VideoID), string(encoded), next)
	if err != nil {
		return SettingsResult{}, err
	}
	effective, err := effectiveSettings(ctx, tx, owner, ref)
	if err != nil {
		return SettingsResult{}, err
	}
	snapshotID, err := newID("snapshot")
	if err != nil {
		return SettingsResult{}, err
	}
	effJSON, err := json.Marshal(effective)
	if err != nil {
		return SettingsResult{}, err
	}
	now := time.Now().UTC()
	if _, err := tx.ExecContext(ctx, `INSERT INTO batch_factory_v12_config_snapshots(id,owner_username,batch_id,book_id,video_id,effective_json,created_at) VALUES(?,?,?,?,?,?,?)`, snapshotID, owner, ref.BatchID, nullIfEmpty(ref.BookID), nullIfEmpty(ref.VideoID), string(effJSON), now); err != nil {
		return SettingsResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return SettingsResult{}, err
	}
	snap := ConfigSnapshot{ID: snapshotID, BatchID: ref.BatchID, BookID: ref.BookID, VideoID: ref.VideoID, Effective: effective, CreatedAt: now}
	return SettingsResult{Scope: ref, Patch: patch, Revision: next, Snapshot: &snap}, nil
}

// AcceptDirectorPromptCandidate updates a selected VIDEO and removes that
// candidate entry while holding both scopes in one MySQL transaction.
func (s *MySQLStore) AcceptDirectorPromptCandidate(ctx context.Context, owner, batchID, bookID, videoID, candidateID string, expectedVideoRevision int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	bookRef := ScopeRef{Kind: ScopeBook, BatchID: batchID, BookID: bookID}
	videoRef := ScopeRef{Kind: ScopeVideo, BatchID: batchID, BookID: bookID, VideoID: videoID}
	bookRevision, err := lockScopeRevision(ctx, tx, owner, bookRef)
	if err != nil {
		return err
	}
	videoRevision, err := lockScopeRevision(ctx, tx, owner, videoRef)
	if err != nil {
		return err
	}
	if videoRevision != expectedVideoRevision {
		return ErrConflict
	}
	bookPatch, err := loadPatch(ctx, tx, owner, bookRef)
	if err != nil {
		return err
	}
	candidate, err := directorPromptCandidateFromPatch(bookPatch)
	if err != nil {
		return err
	}
	if candidate.ID != strings.TrimSpace(candidateID) {
		return fmt.Errorf("%w: candidate no longer matches", ErrConflict)
	}
	prompt, exists := candidate.VideoPrompts[videoID]
	if !exists || strings.TrimSpace(prompt) == "" {
		return ErrNotFound
	}
	videoPatch, err := loadPatch(ctx, tx, owner, videoRef)
	if err != nil {
		return err
	}
	encodedPrompt, err := json.Marshal(prompt)
	if err != nil {
		return err
	}
	videoPatch = ApplySparseUpdate(videoPatch, SettingsUpdate{Patch: SettingsPatch{
		"aiOriginalVideoPrompt": encodedPrompt,
		"videoPrompt":           encodedPrompt,
	}})
	delete(candidate.VideoPrompts, videoID)
	if len(candidate.VideoPrompts) == 0 {
		delete(bookPatch, "directorPromptCandidate")
	} else {
		encodedCandidate, err := json.Marshal(candidate)
		if err != nil {
			return err
		}
		bookPatch["directorPromptCandidate"] = encodedCandidate
	}
	if err := updateScopeRevision(ctx, tx, owner, videoRef, videoRevision+1); err != nil {
		return err
	}
	if err := updateScopeRevision(ctx, tx, owner, bookRef, bookRevision+1); err != nil {
		return err
	}
	if err := upsertSettingsPatch(ctx, tx, owner, videoRef, videoPatch, videoRevision+1); err != nil {
		return err
	}
	if err := upsertSettingsPatch(ctx, tx, owner, bookRef, bookPatch, bookRevision+1); err != nil {
		return err
	}
	return tx.Commit()
}

func upsertSettingsPatch(ctx context.Context, tx *sql.Tx, owner string, ref ScopeRef, patch SettingsPatch, revision int64) error {
	encoded, err := json.Marshal(patch)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO batch_factory_v12_settings_patches(scope_type,scope_id,owner_username,batch_id,book_id,video_id,patch_json,revision,updated_at) VALUES(?,?,?,?,?,?,?,?,CURRENT_TIMESTAMP(6)) ON DUPLICATE KEY UPDATE patch_json=VALUES(patch_json),revision=VALUES(revision),updated_at=VALUES(updated_at)`, string(ref.Kind), scopeID(ref), owner, ref.BatchID, nullIfEmpty(ref.BookID), nullIfEmpty(ref.VideoID), string(encoded), revision)
	return err
}

func scopeID(ref ScopeRef) string {
	switch ref.Kind {
	case ScopeBatch:
		return ref.BatchID
	case ScopeBook:
		return ref.BookID
	case ScopeVideo:
		return ref.VideoID
	}
	return ""
}
func nullIfEmpty(v string) any {
	if v == "" {
		return nil
	}
	return v
}

func lockScopeRevision(ctx context.Context, tx *sql.Tx, owner string, ref ScopeRef) (int64, error) {
	var rev int64
	var err error
	switch ref.Kind {
	case ScopeBatch:
		err = tx.QueryRowContext(ctx, `SELECT revision FROM batch_factory_v12_batches WHERE id=? AND owner_username=? FOR UPDATE`, ref.BatchID, owner).Scan(&rev)
	case ScopeBook:
		err = tx.QueryRowContext(ctx, `SELECT revision FROM batch_factory_v12_books WHERE id=? AND batch_id=? AND owner_username=? FOR UPDATE`, ref.BookID, ref.BatchID, owner).Scan(&rev)
	case ScopeVideo:
		err = tx.QueryRowContext(ctx, `SELECT revision FROM batch_factory_v12_videos WHERE id=? AND book_id=? AND batch_id=? AND owner_username=? FOR UPDATE`, ref.VideoID, ref.BookID, ref.BatchID, owner).Scan(&rev)
	default:
		return 0, ErrInvalid
	}
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNotFound
	}
	return rev, err
}
func updateScopeRevision(ctx context.Context, tx *sql.Tx, owner string, ref ScopeRef, next int64) error {
	var res sql.Result
	var err error
	switch ref.Kind {
	case ScopeBatch:
		res, err = tx.ExecContext(ctx, `UPDATE batch_factory_v12_batches SET revision=?,updated_at=CURRENT_TIMESTAMP(6) WHERE id=? AND owner_username=?`, next, ref.BatchID, owner)
	case ScopeBook:
		res, err = tx.ExecContext(ctx, `UPDATE batch_factory_v12_books SET revision=?,updated_at=CURRENT_TIMESTAMP(6) WHERE id=? AND batch_id=? AND owner_username=?`, next, ref.BookID, ref.BatchID, owner)
	case ScopeVideo:
		res, err = tx.ExecContext(ctx, `UPDATE batch_factory_v12_videos SET revision=?,updated_at=CURRENT_TIMESTAMP(6) WHERE id=? AND book_id=? AND batch_id=? AND owner_username=?`, next, ref.VideoID, ref.BookID, ref.BatchID, owner)
	default:
		return ErrInvalid
	}
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return ErrNotFound
	}
	return nil
}
func loadPatch(ctx context.Context, q batchQueryer, owner string, ref ScopeRef) (SettingsPatch, error) {
	var raw []byte
	err := q.QueryRowContext(ctx, `SELECT patch_json FROM batch_factory_v12_settings_patches WHERE scope_type=? AND scope_id=? AND owner_username=?`, string(ref.Kind), scopeID(ref), owner).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return SettingsPatch{}, nil
	}
	if err != nil {
		return nil, err
	}
	out := SettingsPatch{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return normalizeLegacyNestedPatch(out), nil
}
func effectiveSettings(ctx context.Context, q batchQueryer, owner string, ref ScopeRef) (SettingsPatch, error) {
	layers := []SettingsPatch{}
	for _, r := range []ScopeRef{{Kind: ScopeBatch, BatchID: ref.BatchID}, {Kind: ScopeBook, BatchID: ref.BatchID, BookID: ref.BookID}, {Kind: ScopeVideo, BatchID: ref.BatchID, BookID: ref.BookID, VideoID: ref.VideoID}} {
		if r.Kind == ScopeBook && ref.BookID == "" {
			continue
		}
		if r.Kind == ScopeVideo && ref.VideoID == "" {
			continue
		}
		p, err := loadPatch(ctx, q, owner, r)
		if err != nil {
			return nil, err
		}
		layers = append(layers, p)
	}
	return ResolveSettings(layers...), nil
}

func (s *MySQLStore) ConfigVersions(ctx context.Context, owner string) ([]ConfigVersion, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,name,config_json,created_at FROM batch_factory_v12_config_versions WHERE owner_username IS NULL OR owner_username=? ORDER BY created_at,id`, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []ConfigVersion{}
	for rows.Next() {
		var v ConfigVersion
		var raw []byte
		if err := rows.Scan(&v.ID, &v.Name, &raw, &v.CreatedAt); err != nil {
			return nil, err
		}
		v.Config = json.RawMessage(raw)
		out = append(out, v)
	}
	return out, rows.Err()
}
func validConfigVersionJSON(raw json.RawMessage) bool {
	if len(raw) == 0 || !json.Valid(raw) {
		return false
	}
	var value map[string]any
	return json.Unmarshal(raw, &value) == nil
}
func isAIReasoningPresetConfig(raw json.RawMessage) bool {
	var value struct {
		Type string `json:"type"`
	}
	return json.Unmarshal(raw, &value) == nil && value.Type == "ai-reasoning-preset"
}
func (s *MySQLStore) CreateConfigVersion(ctx context.Context, owner string, version ConfigVersion) (ConfigVersion, error) {
	if strings.TrimSpace(owner) == "" || strings.TrimSpace(version.Name) == "" || !validConfigVersionJSON(version.Config) {
		return ConfigVersion{}, ErrInvalid
	}
	id, err := newID("config")
	if err != nil {
		return ConfigVersion{}, err
	}
	version.ID, version.Name, version.CreatedAt = id, strings.TrimSpace(version.Name), time.Now().UTC()
	if _, err := s.db.ExecContext(ctx, `INSERT INTO batch_factory_v12_config_versions(id,owner_username,name,config_json,created_at) VALUES(?,?,?,?,?)`, version.ID, owner, version.Name, version.Config, version.CreatedAt); err != nil {
		return ConfigVersion{}, err
	}
	return version, nil
}
func (s *MySQLStore) RenameConfigVersion(ctx context.Context, owner, id, name string) (ConfigVersion, error) {
	if strings.TrimSpace(name) == "" {
		return ConfigVersion{}, ErrInvalid
	}
	result, err := s.db.ExecContext(ctx, `UPDATE batch_factory_v12_config_versions SET name=? WHERE id=? AND owner_username=?`, strings.TrimSpace(name), id, owner)
	if err != nil {
		return ConfigVersion{}, err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return ConfigVersion{}, err
	}
	if count != 1 {
		return ConfigVersion{}, ErrNotFound
	}
	var version ConfigVersion
	var raw []byte
	if err := s.db.QueryRowContext(ctx, `SELECT id,name,config_json,created_at FROM batch_factory_v12_config_versions WHERE id=? AND owner_username=?`, id, owner).Scan(&version.ID, &version.Name, &raw, &version.CreatedAt); err != nil {
		return ConfigVersion{}, err
	}
	version.Config = json.RawMessage(raw)
	return version, nil
}
func (s *MySQLStore) ChangeImpact(ctx context.Context, owner, batchID string, update SettingsUpdate) (ChangeImpact, error) {
	b, err := s.GetBatch(ctx, owner, batchID)
	if err != nil {
		return ChangeImpact{}, err
	}
	videos := 0
	for _, book := range b.Books {
		videos += len(book.Videos)
	}
	return ChangeImpact{AffectedBooks: len(b.Books), AffectedVideos: videos, InvalidatesDirector: false, PreservesOverrides: true, ChangedKeys: ChangedKeys(update)}, nil
}
func (s *MySQLStore) ListPrompts(ctx context.Context, owner, kind string) ([]Prompt, error) {
	query := `SELECT d.id,v.id,d.name,d.kind,v.content,v.revision,v.created_at FROM batch_factory_v12_prompt_definitions d JOIN batch_factory_v12_prompt_versions v ON v.prompt_id=d.id AND v.revision=(SELECT MAX(v2.revision) FROM batch_factory_v12_prompt_versions v2 WHERE v2.prompt_id=d.id AND v2.owner_username=d.owner_username) WHERE d.owner_username=?`
	args := []any{owner}
	if kind != "" {
		query += ` AND d.kind=?`
		args = append(args, kind)
	}
	query += ` ORDER BY d.created_at,d.id`
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Prompt{}
	for rows.Next() {
		var p Prompt
		if err := rows.Scan(&p.ID, &p.VersionID, &p.Name, &p.Kind, &p.Content, &p.Revision, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
func (s *MySQLStore) CreatePrompt(ctx context.Context, owner string, p Prompt) (Prompt, error) {
	if strings.TrimSpace(p.Name) == "" || strings.TrimSpace(p.Content) == "" {
		return Prompt{}, ErrInvalid
	}
	id, err := newID("prompt")
	if err != nil {
		return Prompt{}, err
	}
	vid, err := newID("promptv")
	if err != nil {
		return Prompt{}, err
	}
	now := time.Now().UTC()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Prompt{}, err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `INSERT INTO batch_factory_v12_prompt_definitions(id,owner_username,name,kind,created_at) VALUES(?,?,?,?,?)`, id, owner, p.Name, p.Kind, now); err != nil {
		return Prompt{}, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO batch_factory_v12_prompt_versions(id,prompt_id,owner_username,revision,content,created_at) VALUES(?,?,?,1,?,?)`, vid, id, owner, p.Content, now); err != nil {
		return Prompt{}, err
	}
	if err := tx.Commit(); err != nil {
		return Prompt{}, err
	}
	p.ID = id
	p.VersionID = vid
	p.Revision = 1
	p.CreatedAt = now
	return p, nil
}
func (s *MySQLStore) GetDraft(ctx context.Context, owner, key, kind, scope string) (Draft, error) {
	var d Draft
	err := s.db.QueryRowContext(ctx, `SELECT draft_key,kind,scope,COALESCE(content,''),revision,updated_at FROM batch_factory_v12_drafts WHERE owner_username=? AND draft_key=? AND kind=? AND scope=?`, owner, key, kind, scope).Scan(&d.Key, &d.Kind, &d.Scope, &d.Content, &d.Revision, &d.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Draft{}, ErrNotFound
	}
	return d, err
}
func (s *MySQLStore) SaveDraft(ctx context.Context, owner string, d Draft) (Draft, error) {
	if strings.TrimSpace(d.Key) == "" {
		return Draft{}, ErrInvalid
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Draft{}, err
	}
	defer tx.Rollback()
	var current int64
	err = tx.QueryRowContext(ctx, `SELECT revision FROM batch_factory_v12_drafts WHERE owner_username=? AND draft_key=? AND kind=? AND scope=? FOR UPDATE`, owner, d.Key, d.Kind, d.Scope).Scan(&current)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return Draft{}, err
	}
	next := current + 1
	if _, err := tx.ExecContext(ctx, `INSERT INTO batch_factory_v12_drafts(owner_username,draft_key,kind,scope,content,revision,updated_at) VALUES(?,?,?,?,?,?,CURRENT_TIMESTAMP(6)) ON DUPLICATE KEY UPDATE content=VALUES(content),revision=VALUES(revision),updated_at=VALUES(updated_at)`, owner, d.Key, d.Kind, d.Scope, nullableString(d.Content), next); err != nil {
		return Draft{}, err
	}
	if err := tx.Commit(); err != nil {
		return Draft{}, err
	}
	d.Revision = next
	d.UpdatedAt = time.Now().UTC()
	return d, nil
}

func (s *MySQLStore) String() string { return fmt.Sprintf("BatchFactoryV11MySQLStore(%p)", s.db) }
