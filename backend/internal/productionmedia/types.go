package productionmedia

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"
)

const (
	KindComicImage = "comic_image"
	KindAssetImage = "asset_image"
	KindAudio      = "audio"
	KindVideo      = "video"
	KindMergeVideo = "merge_video"

	StatePending        = "pending"
	StateUploading      = "uploading"
	StateVerified       = "verified"
	StateCleanupPending = "cleanup_pending"
	StateCleaned        = "cleaned"
	StateFailed         = "failed"
)

type Object struct {
	ID               int64
	Owner            string
	MediaKind        string
	MediaType        string
	LocalStorageRef  string
	ObjectKey        string
	ByteSize         int64
	SHA256           string
	State            string
	RetentionDays    int
	RemoteVerifiedAt time.Time
	PurgeAfter       time.Time
}

type RemoteMetadata struct {
	ByteSize  int64
	SHA256    string
	MediaType string
}

type RemoteStore interface {
	PutFile(context.Context, string, string, string, string, int64) error
	Head(context.Context, string) (RemoteMetadata, error)
	Open(context.Context, string) (io.ReadCloser, error)
}

type Repository interface {
	ClaimUpload(context.Context) (Object, bool, error)
	MarkVerified(context.Context, Object) error
	MarkUploadFailed(context.Context, int64, error) error
	FindVerifiedByStorageRef(context.Context, string) (Object, bool, error)
	MarkRehydrated(context.Context, int64, time.Time) error
	ClaimDuePurge(context.Context, bool) (Object, bool, error)
	MarkLocalDeleted(context.Context, int64) error
	MarkPurgeRetry(context.Context, int64, error) error
}

func SHA256Bytes(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

func NormalizeRetentionDays(days int) int {
	switch days {
	case 3, 7, 14, 30:
		return days
	default:
		return 7
	}
}

func ObjectKey(prefix, kind, storageRef string) (string, error) {
	prefix = strings.Trim(strings.TrimSpace(prefix), "/")
	storageRef = strings.TrimSpace(storageRef)
	if prefix == "" || storageRef == "" || filepath.Base(storageRef) != storageRef || strings.Contains(storageRef, "\\") {
		return "", fmt.Errorf("invalid production media object key")
	}
	var folder string
	switch kind {
	case KindComicImage:
		folder = "comic/images"
	case KindAssetImage:
		folder = "comic/assets"
	case KindAudio:
		folder = "audio"
	case KindVideo:
		folder = "video/clips"
	case KindMergeVideo:
		folder = "video/merged"
	default:
		return "", fmt.Errorf("invalid production media kind %q", kind)
	}
	return prefix + "/" + folder + "/" + storageRef, nil
}
