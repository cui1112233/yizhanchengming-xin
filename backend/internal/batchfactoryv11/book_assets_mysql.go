package batchfactoryv11

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

func ensureBookOwnership(ctx context.Context, q interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, owner, batchID, bookID string) error {
	var one int
	err := q.QueryRowContext(ctx, `SELECT 1 FROM batch_factory_v11_books WHERE id=? AND batch_id=? AND owner_username=?`, bookID, batchID, owner).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func (s *MySQLStore) ListBookAssets(ctx context.Context, owner, batchID, bookID string) ([]BookAsset, error) {
	if err := ensureBookOwnership(ctx, s.db, owner, batchID, bookID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id,kind,name,COALESCE(formal_name,''),COALESCE(facts,''),COALESCE(appearance,''),COALESCE(relationships,''),prompt,COALESCE(image_model_id,''),COALESCE(image_aspect_ratio,''),source,COALESCE(extraction_preset_id,''),COALESCE(extraction_preset_version,0),revision,created_at,updated_at FROM batch_factory_v11_book_assets WHERE owner_username=? AND batch_id=? AND book_id=? ORDER BY kind,name,id`, owner, batchID, bookID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	assets := []BookAsset{}
	for rows.Next() {
		var asset BookAsset
		if err := rows.Scan(&asset.ID, &asset.Kind, &asset.Name, &asset.FormalName, &asset.Facts, &asset.Appearance, &asset.Relationships, &asset.Prompt, &asset.ImageModelID, &asset.ImageAspectRatio, &asset.Source, &asset.ExtractionPresetID, &asset.ExtractionPresetVersion, &asset.Revision, &asset.CreatedAt, &asset.UpdatedAt); err != nil {
			return nil, err
		}
		if asset.Appearance == "" {
			asset.Appearance = asset.Prompt
		}
		asset.BatchID, asset.BookID = batchID, bookID
		assets = append(assets, asset)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return assets, nil
}

func (s *MySQLStore) CreateBookAsset(ctx context.Context, owner, batchID, bookID string, input CreateBookAssetInput) (BookAsset, error) {
	input, err := normalizeBookAssetInput(input)
	if err != nil {
		return BookAsset{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return BookAsset{}, err
	}
	defer tx.Rollback()
	if err := ensureBookOwnership(ctx, tx, owner, batchID, bookID); err != nil {
		return BookAsset{}, err
	}
	id, err := newID("asset")
	if err != nil {
		return BookAsset{}, err
	}
	now := time.Now().UTC()
	asset := BookAsset{ID: id, BatchID: batchID, BookID: bookID, Kind: input.Kind, Name: input.Name, FormalName: input.FormalName, Facts: input.Facts, Appearance: input.Appearance, Relationships: input.Relationships, Prompt: input.Prompt, ImageModelID: input.ImageModelID, ImageAspectRatio: input.ImageAspectRatio, Source: "manual", Revision: 1, CreatedAt: now, UpdatedAt: now}
	if _, err := tx.ExecContext(ctx, `INSERT INTO batch_factory_v11_book_assets(id,owner_username,batch_id,book_id,kind,name,formal_name,facts,appearance,relationships,prompt,image_model_id,image_aspect_ratio,source,revision,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,'manual',1,?,?)`, asset.ID, owner, asset.BatchID, asset.BookID, asset.Kind, asset.Name, asset.FormalName, asset.Facts, asset.Appearance, asset.Relationships, asset.Prompt, nullableString(asset.ImageModelID), nullableString(asset.ImageAspectRatio), now, now); err != nil {
		return BookAsset{}, err
	}
	if err := tx.Commit(); err != nil {
		return BookAsset{}, err
	}
	return asset, nil
}

// PersistExtractedBookAssets replaces the generated asset set with the result
// for the current frozen video source. Manual assets are deliberately retained;
// otherwise a five-line refresh would keep every scene and prop generated from
// an earlier, longer source and silently leak them into the new workbench.
func (s *MySQLStore) PersistExtractedBookAssets(ctx context.Context, owner string, book Book, snapshot DirectorSnapshot, assets DirectorAssets) ([]BookAsset, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if err := ensureBookOwnership(ctx, tx, owner, book.BatchID, book.ID); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM batch_factory_v11_book_assets WHERE owner_username=? AND batch_id=? AND book_id=? AND source='director'`, owner, book.BatchID, book.ID); err != nil {
		return nil, err
	}
	output := DirectorResult{Characters: assets.Characters, Scenes: assets.Scenes, Props: assets.Props}
	if err := persistDirectorBookAssets(ctx, tx, owner, book, snapshot, output, time.Now().UTC()); err != nil {
		return nil, err
	}
	// Older records may have been manually saved before the structured asset
	// fields were added. Keep their hand-written prompt, but fill only empty
	// structured fields from the new extraction result.
	now := time.Now().UTC()
	for _, seed := range directorBookAssets(book, snapshot, output) {
		if _, err := tx.ExecContext(ctx, `UPDATE batch_factory_v11_book_assets SET
			formal_name=CASE WHEN TRIM(COALESCE(formal_name,''))='' THEN ? ELSE formal_name END,
			facts=CASE WHEN TRIM(COALESCE(facts,''))='' THEN ? ELSE facts END,
			appearance=CASE WHEN TRIM(COALESCE(appearance,''))='' THEN ? ELSE appearance END,
			relationships=CASE WHEN TRIM(COALESCE(relationships,''))='' THEN ? ELSE relationships END,
			updated_at=?
			WHERE owner_username=? AND batch_id=? AND book_id=? AND source='manual' AND kind=? AND name=?`,
			seed.FormalName, seed.Facts, seed.Appearance, seed.Relationships, now, owner, book.BatchID, book.ID, seed.Kind, seed.Name); err != nil {
			return nil, err
		}
	}
	rows, err := tx.QueryContext(ctx, `SELECT id,kind,name,COALESCE(formal_name,''),COALESCE(facts,''),COALESCE(appearance,''),COALESCE(relationships,''),prompt,COALESCE(image_model_id,''),COALESCE(image_aspect_ratio,''),source,COALESCE(extraction_preset_id,''),COALESCE(extraction_preset_version,0),revision,created_at,updated_at FROM batch_factory_v11_book_assets WHERE owner_username=? AND batch_id=? AND book_id=? ORDER BY kind,name,id`, owner, book.BatchID, book.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []BookAsset{}
	for rows.Next() {
		var asset BookAsset
		if err := rows.Scan(&asset.ID, &asset.Kind, &asset.Name, &asset.FormalName, &asset.Facts, &asset.Appearance, &asset.Relationships, &asset.Prompt, &asset.ImageModelID, &asset.ImageAspectRatio, &asset.Source, &asset.ExtractionPresetID, &asset.ExtractionPresetVersion, &asset.Revision, &asset.CreatedAt, &asset.UpdatedAt); err != nil {
			return nil, err
		}
		if asset.Appearance == "" {
			asset.Appearance = asset.Prompt
		}
		asset.BatchID, asset.BookID = book.BatchID, book.ID
		result = append(result, asset)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *MySQLStore) UpdateBookAsset(ctx context.Context, owner, batchID, bookID, assetID string, input UpdateBookAssetInput) (BookAsset, error) {
	input.Name, input.Prompt = strings.TrimSpace(input.Name), strings.TrimSpace(input.Prompt)
	input.FormalName, input.Facts, input.Appearance, input.Relationships, input.ImageModelID = strings.TrimSpace(input.FormalName), strings.TrimSpace(input.Facts), strings.TrimSpace(input.Appearance), strings.TrimSpace(input.Relationships), strings.TrimSpace(input.ImageModelID)
	if input.Name == "" || input.Prompt == "" || len(input.Name) > 255 || !validBookAssetImageAspectRatio(input.ImageAspectRatio) {
		return BookAsset{}, ErrInvalid
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return BookAsset{}, err
	}
	defer tx.Rollback()
	if err := ensureBookOwnership(ctx, tx, owner, batchID, bookID); err != nil {
		return BookAsset{}, err
	}
	now := time.Now().UTC()
	result, err := tx.ExecContext(ctx, `UPDATE batch_factory_v11_book_assets SET name=?,formal_name=?,facts=?,appearance=?,relationships=?,prompt=?,image_model_id=?,image_aspect_ratio=?,source='manual',revision=revision+1,updated_at=? WHERE id=? AND owner_username=? AND batch_id=? AND book_id=? AND revision=?`, input.Name, input.FormalName, input.Facts, input.Appearance, input.Relationships, input.Prompt, nullableString(input.ImageModelID), nullableString(input.ImageAspectRatio), now, assetID, owner, batchID, bookID, input.ExpectedRevision)
	if err != nil {
		return BookAsset{}, err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		var revision int64
		err := tx.QueryRowContext(ctx, `SELECT revision FROM batch_factory_v11_book_assets WHERE id=? AND owner_username=? AND batch_id=? AND book_id=?`, assetID, owner, batchID, bookID).Scan(&revision)
		if errors.Is(err, sql.ErrNoRows) {
			return BookAsset{}, ErrNotFound
		}
		if err != nil {
			return BookAsset{}, err
		}
		return BookAsset{}, ErrConflict
	}
	var asset BookAsset
	err = tx.QueryRowContext(ctx, `SELECT kind,name,COALESCE(formal_name,''),COALESCE(facts,''),COALESCE(appearance,''),COALESCE(relationships,''),prompt,COALESCE(image_model_id,''),COALESCE(image_aspect_ratio,''),source,COALESCE(extraction_preset_id,''),COALESCE(extraction_preset_version,0),revision,created_at,updated_at FROM batch_factory_v11_book_assets WHERE id=? AND owner_username=? AND batch_id=? AND book_id=?`, assetID, owner, batchID, bookID).Scan(&asset.Kind, &asset.Name, &asset.FormalName, &asset.Facts, &asset.Appearance, &asset.Relationships, &asset.Prompt, &asset.ImageModelID, &asset.ImageAspectRatio, &asset.Source, &asset.ExtractionPresetID, &asset.ExtractionPresetVersion, &asset.Revision, &asset.CreatedAt, &asset.UpdatedAt)
	if err != nil {
		return BookAsset{}, err
	}
	asset.ID, asset.BatchID, asset.BookID = assetID, batchID, bookID
	if err := tx.Commit(); err != nil {
		return BookAsset{}, err
	}
	return asset, nil
}

// DeleteBookAsset removes one manually managed asset and its image versions.
// The legacy Shuihuo route uses this narrow optional method so the shared
// Store interface does not have to change for existing callers.
func (s *MySQLStore) DeleteBookAsset(ctx context.Context, owner, batchID, bookID, assetID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := ensureBookAssetOwnership(ctx, tx, owner, batchID, bookID, assetID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM batch_factory_v11_book_asset_images WHERE asset_id=? AND owner_username=?`, assetID, owner); err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, `DELETE FROM batch_factory_v11_book_assets WHERE id=? AND owner_username=? AND batch_id=? AND book_id=?`, assetID, owner, batchID, bookID)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return ErrNotFound
	}
	return tx.Commit()
}

func persistDirectorBookAssets(ctx context.Context, tx *sql.Tx, owner string, book Book, snapshot DirectorSnapshot, output DirectorResult, now time.Time) error {
	desired := directorBookAssets(book, snapshot, output)
	// No structured assets in this run (e.g. SD plain-text mode, where assets
	// are saved separately by the extraction stage). An empty list means "no
	// new information", never "delete everything": leave the ledger untouched.
	if len(desired) == 0 {
		return nil
	}

	rows, err := tx.QueryContext(ctx, `SELECT id,kind,name,source FROM batch_factory_v11_book_assets WHERE owner_username=? AND batch_id=? AND book_id=?`, owner, book.BatchID, book.ID)
	if err != nil {
		return err
	}
	existing := []existingAssetRow{}
	for rows.Next() {
		var row existingAssetRow
		if err := rows.Scan(&row.ID, &row.Kind, &row.Name, &row.Source); err != nil {
			rows.Close()
			return err
		}
		existing = append(existing, row)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()

	hasImage := map[string]bool{}
	imageRows, err := tx.QueryContext(ctx, `SELECT DISTINCT asset_id FROM batch_factory_v11_book_asset_images WHERE owner_username=?`, owner)
	if err != nil {
		return err
	}
	for imageRows.Next() {
		var assetID string
		if err := imageRows.Scan(&assetID); err != nil {
			imageRows.Close()
			return err
		}
		hasImage[assetID] = true
	}
	if err := imageRows.Err(); err != nil {
		imageRows.Close()
		return err
	}
	imageRows.Close()

	plan := reconcileDirectorAssets(existing, desired, hasImage)

	// Delete stale rows first so later inserts never collide with freed names.
	for _, id := range plan.Deletes {
		if _, err := tx.ExecContext(ctx, `DELETE FROM batch_factory_v11_book_assets WHERE id=? AND owner_username=? AND source='director'`, id, owner); err != nil {
			return err
		}
	}
	// Matched rows are frozen on purpose: no UPDATE, their saved prompt wins.
	for _, asset := range plan.Inserts {
		id, err := newID("asset")
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO batch_factory_v11_book_assets(id,owner_username,batch_id,book_id,kind,name,formal_name,facts,appearance,relationships,prompt,image_model_id,image_aspect_ratio,source,extraction_preset_id,extraction_preset_version,revision,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,'director',?,?,1,?,?)`,
			id, owner, book.BatchID, book.ID, asset.Kind, asset.Name, asset.FormalName, asset.Facts, asset.Appearance, asset.Relationships, asset.Prompt, nil, nil, nullableString(asset.ExtractionPresetID), nullableInt64(asset.ExtractionPresetVersion), now, now); err != nil {
			return err
		}
	}
	return nil
}

func nullableInt64(v int64) any {
	if v == 0 {
		return nil
	}
	return v
}
