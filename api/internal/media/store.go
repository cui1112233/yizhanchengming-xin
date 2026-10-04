package media

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	TypeImage = "image"
	TypeVideo = "video"

	StatusReady = "ready"
)

type Asset struct {
	ID           string          `json:"id"`
	Owner        string          `json:"-"`
	MediaType    string          `json:"media_type"`
	Status       string          `json:"status"`
	TOSBucket    string          `json:"-"`
	TOSKey       string          `json:"-"`
	MimeType     string          `json:"mime_type,omitempty"`
	SizeBytes    uint64          `json:"size_bytes"`
	WidthPx      uint            `json:"width_px"`
	HeightPx     uint            `json:"height_px"`
	DurationMs   uint64          `json:"duration_ms"`
	SourceTaskID string          `json:"source_task_id,omitempty"`
	Metadata     json.RawMessage `json:"metadata,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

type Reference struct {
	ID         string `json:"id"`
	MediaType  string `json:"media_type"`
	MimeType   string `json:"mime_type,omitempty"`
	StorageKey string `json:"storage_key"`
}

func (a Asset) Reference() Reference {
	storageKey := ""
	if strings.TrimSpace(a.TOSBucket) != "" && strings.TrimSpace(a.TOSKey) != "" {
		storageKey = "tos://" + strings.TrimSpace(a.TOSBucket) + "/" + strings.TrimLeft(strings.TrimSpace(a.TOSKey), "/")
	}
	return Reference{ID: a.ID, MediaType: a.MediaType, MimeType: a.MimeType, StorageKey: storageKey}
}

type CreateInput struct {
	Owner        string
	MediaType    string
	TOSBucket    string
	TOSKey       string
	MimeType     string
	SizeBytes    uint64
	WidthPx      uint
	HeightPx     uint
	DurationMs   uint64
	SourceTaskID string
	Metadata     json.RawMessage
}

type executor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

type SQLStore struct {
	db   *sql.DB
	exec executor
}

func NewSQLStore(db *sql.DB) *SQLStore { return newSQLStoreWithExecutor(db) }

func newSQLStoreWithExecutor(exec executor) *SQLStore { return &SQLStore{exec: exec} }

func validateCreateInput(input CreateInput) error {
	if strings.TrimSpace(input.Owner) == "" {
		return errors.New("media owner is required")
	}
	if input.MediaType != TypeImage && input.MediaType != TypeVideo {
		return errors.New("media_type must be image or video")
	}
	if strings.TrimSpace(input.TOSBucket) == "" {
		return errors.New("TOS bucket is required")
	}
	if strings.TrimSpace(input.TOSKey) == "" {
		return errors.New("TOS key is required")
	}
	if len(input.Metadata) > 0 && !json.Valid(input.Metadata) {
		return errors.New("media metadata must be valid json")
	}
	return nil
}

func newAssetID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil { return "", err }
	return "asset_" + hex.EncodeToString(buf), nil
}

func (s *SQLStore) Create(ctx context.Context, input CreateInput) (Asset, error) {
	if s == nil || s.exec == nil {
		return Asset{}, errors.New("media store unavailable")
	}
	if err := validateCreateInput(input); err != nil { return Asset{}, err }
	id, err := newAssetID()
	if err != nil { return Asset{}, err }
	now := time.Now().UTC()
	asset := Asset{
		ID: id,
		Owner: strings.TrimSpace(input.Owner),
		MediaType: input.MediaType,
		Status: StatusReady,
		TOSBucket: strings.TrimSpace(input.TOSBucket),
		TOSKey: strings.TrimSpace(input.TOSKey),
		MimeType: strings.TrimSpace(input.MimeType),
		SizeBytes: input.SizeBytes,
		WidthPx: input.WidthPx,
		HeightPx: input.HeightPx,
		DurationMs: input.DurationMs,
		SourceTaskID: strings.TrimSpace(input.SourceTaskID),
		Metadata: append(json.RawMessage(nil), input.Metadata...),
		CreatedAt: now,
		UpdatedAt: now,
	}
	var sourceTask any
	if asset.SourceTaskID != "" { sourceTask = asset.SourceTaskID }
	var metadata any
	if len(asset.Metadata) > 0 { metadata = []byte(asset.Metadata) }
	result, err := s.exec.ExecContext(ctx, `INSERT INTO media_assets
(id, owner, media_type, tos_bucket, mime_type, tos_key, status, size_bytes, width_px, height_px, duration_ms, source_task_id, metadata_json)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		asset.ID, asset.Owner, asset.MediaType, asset.TOSBucket, asset.MimeType, asset.TOSKey, asset.Status,
		asset.SizeBytes, asset.WidthPx, asset.HeightPx, asset.DurationMs, sourceTask, metadata)
	if err != nil { return Asset{}, err }
	rows, err := result.RowsAffected()
	if err != nil { return Asset{}, err }
	if rows != 1 { return Asset{}, fmt.Errorf("media asset insert affected %d rows", rows) }
	return asset, nil
}

func (s *SQLStore) Get(ctx context.Context, owner, id string) (Asset, error) {
	if s == nil || s.db == nil { return Asset{}, errors.New("media store unavailable") }
	var asset Asset
	var metadata []byte
	var sourceTask sql.NullString
	err := s.db.QueryRowContext(ctx, `SELECT id, owner, media_type, status, tos_bucket, tos_key, mime_type, size_bytes, width_px, height_px, duration_ms, source_task_id, metadata_json, created_at, updated_at
FROM media_assets WHERE owner=? AND id=? LIMIT 1`, strings.TrimSpace(owner), strings.TrimSpace(id)).Scan(
		&asset.ID, &asset.Owner, &asset.MediaType, &asset.Status, &asset.TOSBucket, &asset.TOSKey, &asset.MimeType,
		&asset.SizeBytes, &asset.WidthPx, &asset.HeightPx, &asset.DurationMs, &sourceTask, &metadata, &asset.CreatedAt, &asset.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) { return Asset{}, sql.ErrNoRows }
	if err != nil { return Asset{}, err }
	if sourceTask.Valid { asset.SourceTaskID = sourceTask.String }
	if len(metadata) > 0 { asset.Metadata = append(json.RawMessage(nil), metadata...) }
	return asset, nil
}
