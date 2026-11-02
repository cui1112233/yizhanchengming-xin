package batchfactoryv11

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

func ensureBookAssetOwnership(ctx context.Context, q interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, owner, batchID, bookID, assetID string) error {
	var one int
	err := q.QueryRowContext(ctx, `SELECT 1 FROM batch_factory_v12_book_assets WHERE id=? AND owner_username=? AND batch_id=? AND book_id=?`, assetID, owner, batchID, bookID).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func (s *MySQLStore) ListBookAssetImages(ctx context.Context, owner, batchID, bookID, assetID string) ([]BookAssetImage, error) {
	if err := ensureBookOwnership(ctx, s.db, owner, batchID, bookID); err != nil {
		return nil, err
	}
	if err := ensureBookAssetOwnership(ctx, s.db, owner, batchID, bookID, assetID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id,url,COALESCE(storage_ref,''),media_type,source,is_primary,revision,created_at FROM batch_factory_v12_book_asset_images WHERE owner_username=? AND asset_id=? ORDER BY created_at,id`, owner, assetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	images := []BookAssetImage{}
	for rows.Next() {
		var image BookAssetImage
		if err := rows.Scan(&image.ID, &image.URL, &image.StorageRef, &image.MediaType, &image.Source, &image.IsPrimary, &image.Revision, &image.CreatedAt); err != nil {
			return nil, err
		}
		image.AssetID = assetID
		images = append(images, image)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return images, nil
}

func (s *MySQLStore) CreateBookAssetImage(ctx context.Context, owner, batchID, bookID, assetID string, input CreateBookAssetImageInput) (BookAssetImage, error) {
	input, err := validBookAssetImageInput(input)
	if err != nil {
		return BookAssetImage{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return BookAssetImage{}, err
	}
	defer tx.Rollback()
	if err := ensureBookOwnership(ctx, tx, owner, batchID, bookID); err != nil {
		return BookAssetImage{}, err
	}
	if err := ensureBookAssetOwnership(ctx, tx, owner, batchID, bookID, assetID); err != nil {
		return BookAssetImage{}, err
	}
	var primaryCount int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM batch_factory_v12_book_asset_images WHERE owner_username=? AND asset_id=? AND is_primary=TRUE FOR UPDATE`, owner, assetID).Scan(&primaryCount); err != nil {
		return BookAssetImage{}, err
	}
	id, err := newID("asset-image")
	if err != nil {
		return BookAssetImage{}, err
	}
	now := time.Now().UTC()
	image := BookAssetImage{ID: id, AssetID: assetID, URL: input.URL, StorageRef: input.StorageRef, MediaType: input.MediaType, Source: input.Source, IsPrimary: primaryCount == 0, Revision: 1, CreatedAt: now}
	if _, err := tx.ExecContext(ctx, `INSERT INTO batch_factory_v12_book_asset_images(id,owner_username,asset_id,url,storage_ref,media_type,source,is_primary,revision,created_at) VALUES(?,?,?,?,?,?,?,?,1,?)`, image.ID, owner, image.AssetID, image.URL, nullableString(image.StorageRef), image.MediaType, image.Source, image.IsPrimary, now); err != nil {
		return BookAssetImage{}, err
	}
	if err := tx.Commit(); err != nil {
		return BookAssetImage{}, err
	}
	return image, nil
}

func (s *MySQLStore) SetPrimaryBookAssetImage(ctx context.Context, owner, batchID, bookID, assetID, imageID string) (BookAssetImage, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return BookAssetImage{}, err
	}
	defer tx.Rollback()
	if err := ensureBookOwnership(ctx, tx, owner, batchID, bookID); err != nil {
		return BookAssetImage{}, err
	}
	if err := ensureBookAssetOwnership(ctx, tx, owner, batchID, bookID, assetID); err != nil {
		return BookAssetImage{}, err
	}
	var image BookAssetImage
	err = tx.QueryRowContext(ctx, `SELECT url,COALESCE(storage_ref,''),media_type,source,is_primary,revision,created_at FROM batch_factory_v12_book_asset_images WHERE id=? AND owner_username=? AND asset_id=? FOR UPDATE`, imageID, owner, assetID).Scan(&image.URL, &image.StorageRef, &image.MediaType, &image.Source, &image.IsPrimary, &image.Revision, &image.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return BookAssetImage{}, ErrNotFound
	}
	if err != nil {
		return BookAssetImage{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE batch_factory_v12_book_asset_images SET is_primary=FALSE,revision=revision+1 WHERE owner_username=? AND asset_id=? AND is_primary=TRUE AND id<>?`, owner, assetID, imageID); err != nil {
		return BookAssetImage{}, err
	}
	if !image.IsPrimary {
		if _, err := tx.ExecContext(ctx, `UPDATE batch_factory_v12_book_asset_images SET is_primary=TRUE,revision=revision+1 WHERE id=? AND owner_username=? AND asset_id=?`, imageID, owner, assetID); err != nil {
			return BookAssetImage{}, err
		}
		image.IsPrimary, image.Revision = true, image.Revision+1
	}
	image.ID, image.AssetID = imageID, assetID
	if err := tx.Commit(); err != nil {
		return BookAssetImage{}, err
	}
	return image, nil
}
