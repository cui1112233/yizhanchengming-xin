package media

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

type ObjectUploader interface {
	Put(context.Context, string, string, int64, io.Reader) (string, string, error)
	Delete(context.Context, string) error
}

type AssetWriter interface {
	Create(context.Context, CreateInput) (Asset, error)
}

type IngestInput struct {
	Owner         string
	MediaType     string
	ContentType   string
	ContentLength int64
	Body          io.Reader
	ObjectKey     string
	WidthPx       uint
	HeightPx      uint
	DurationMs    uint64
	SourceTaskID  string
	Metadata      json.RawMessage
}

type IngestService struct {
	Uploader ObjectUploader
	Assets   AssetWriter
	Now      func() time.Time
}

func (s IngestService) Ingest(ctx context.Context, input IngestInput) (Asset, error) {
	if s.Uploader == nil || s.Assets == nil {
		return Asset{}, errors.New("media ingest service unavailable")
	}
	input.Owner = strings.TrimSpace(input.Owner)
	input.ContentType = canonicalContentType(input.ContentType)
	if input.Owner == "" { return Asset{}, errors.New("media owner is required") }
	if input.Body == nil { return Asset{}, errors.New("media body is required") }
	if err := validateMediaContentType(input.MediaType, input.ContentType); err != nil { return Asset{}, err }
	if len(input.Metadata) > 0 && !json.Valid(input.Metadata) { return Asset{}, errors.New("media metadata must be valid json") }

	key := strings.TrimLeft(strings.TrimSpace(input.ObjectKey), "/")
	if key == "" {
		generated, err := s.generateObjectKey(input.MediaType, input.ContentType)
		if err != nil { return Asset{}, err }
		key = generated
	}
	if err := validateObjectKey(key); err != nil { return Asset{}, err }
	contentLength := input.ContentLength
	if contentLength <= 0 { contentLength = -1 }
	bucket, storedKey, err := s.Uploader.Put(ctx, key, input.ContentType, contentLength, input.Body)
	if err != nil { return Asset{}, fmt.Errorf("upload media to TOS: %w", err) }
	bucket = strings.TrimSpace(bucket)
	storedKey = strings.TrimLeft(strings.TrimSpace(storedKey), "/")
	if bucket == "" || storedKey == "" {
		_ = s.Uploader.Delete(ctx, key)
		return Asset{}, errors.New("TOS upload returned empty bucket or key")
	}

	asset, persistErr := s.Assets.Create(ctx, CreateInput{
		Owner: input.Owner,
		MediaType: input.MediaType,
		TOSBucket: bucket,
		TOSKey: storedKey,
		MimeType: input.ContentType,
		SizeBytes: nonNegativeUint64(input.ContentLength),
		WidthPx: input.WidthPx,
		HeightPx: input.HeightPx,
		DurationMs: input.DurationMs,
		SourceTaskID: strings.TrimSpace(input.SourceTaskID),
		Metadata: append(json.RawMessage(nil), input.Metadata...),
	})
	if persistErr == nil { return asset, nil }
	cleanupErr := s.Uploader.Delete(ctx, storedKey)
	if cleanupErr != nil {
		return Asset{}, fmt.Errorf("persist media asset: %v; cleanup uploaded TOS object: %w", persistErr, cleanupErr)
	}
	return Asset{}, fmt.Errorf("persist media asset: %w", persistErr)
}

func (s IngestService) generateObjectKey(mediaType, contentType string) (string, error) {
	extension, err := mediaExtension(contentType)
	if err != nil { return "", err }
	now := time.Now().UTC()
	if s.Now != nil { now = s.Now().UTC() }
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil { return "", err }
	return fmt.Sprintf("%s/%04d/%02d/%02d/%s%s", mediaType, now.Year(), now.Month(), now.Day(), hex.EncodeToString(buf), extension), nil
}

func canonicalContentType(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if index := strings.Index(value, ";"); index >= 0 { value = strings.TrimSpace(value[:index]) }
	return value
}

func validateMediaContentType(mediaType, contentType string) error {
	switch mediaType {
	case TypeImage:
		if !strings.HasPrefix(contentType, "image/") { return errors.New("image asset requires an image MIME type") }
	case TypeVideo:
		if !strings.HasPrefix(contentType, "video/") { return errors.New("video asset requires a video MIME type") }
	default:
		return errors.New("media_type must be image or video")
	}
	_, err := mediaExtension(contentType)
	return err
}

func mediaExtension(contentType string) (string, error) {
	switch canonicalContentType(contentType) {
	case "image/png": return ".png", nil
	case "image/jpeg", "image/jpg": return ".jpg", nil
	case "image/webp": return ".webp", nil
	case "image/gif": return ".gif", nil
	case "video/mp4": return ".mp4", nil
	case "video/webm": return ".webm", nil
	case "video/quicktime": return ".mov", nil
	default: return "", fmt.Errorf("unsupported media MIME type %q", contentType)
	}
}

func validateObjectKey(key string) error {
	if key == "" { return errors.New("TOS object key is required") }
	if strings.Contains(key, "\\") || strings.Contains(key, "../") || strings.HasPrefix(key, "../") || strings.ContainsAny(key, "\r\n\x00") {
		return errors.New("invalid TOS object key")
	}
	return nil
}

func nonNegativeUint64(value int64) uint64 {
	if value <= 0 { return 0 }
	return uint64(value)
}
