package agentworkspace

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"regexp"
	"strings"
	"time"
)

var (
	ErrAssetNotFound           = errors.New("agent asset not found")
	ErrAssetExists             = errors.New("agent asset already exists")
	ErrInvalidAsset            = errors.New("invalid agent asset")
	ErrReferenceURLUnavailable = errors.New("agent reference URL unavailable")
)

var assetSafeID = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,95}$`)
var assetExtension = regexp.MustCompile(`^[a-z0-9]{1,10}$`)

type Asset struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"projectId"`
	Kind        string    `json:"kind"`
	Name        string    `json:"name"`
	ContentType string    `json:"contentType"`
	ObjectKey   string    `json:"-"`
	URL         string    `json:"url"`
	ByteSize    int64     `json:"byteSize"`
	CreatedAt   time.Time `json:"createdAt"`
}

type AssetStore interface {
	ListAssets(context.Context, string, string, string) ([]Asset, error)
	CreateAsset(context.Context, string, Asset) error
	Asset(context.Context, string, string, string) (Asset, error)
}

type AssetObjectStore interface {
	Put(context.Context, Asset, io.Reader) error
	Open(context.Context, Asset) (io.ReadCloser, error)
	Delete(context.Context, Asset) error
}

// ReferenceURLProvider exposes a short-lived read URL only to a server-side
// model submission. It must never expose object-storage credentials to users.
type ReferenceURLProvider interface {
	ReferenceURL(context.Context, Asset, time.Duration) (string, error)
}

func ValidAssetKind(kind string) bool {
	return kind == "image" || kind == "video" || kind == "audio"
}

func ValidateAsset(asset Asset) error {
	if !assetSafeID.MatchString(asset.ID) || !assetSafeID.MatchString(asset.ProjectID) || !ValidAssetKind(asset.Kind) {
		return ErrInvalidAsset
	}
	if strings.TrimSpace(asset.Name) == "" || len(asset.Name) > 191 || strings.TrimSpace(asset.ObjectKey) == "" || asset.ByteSize < 0 {
		return ErrInvalidAsset
	}
	contentType := strings.ToLower(strings.TrimSpace(asset.ContentType))
	if (asset.Kind == "image" && !strings.HasPrefix(contentType, "image/")) ||
		(asset.Kind == "video" && !strings.HasPrefix(contentType, "video/")) ||
		(asset.Kind == "audio" && !strings.HasPrefix(contentType, "audio/")) {
		return ErrInvalidAsset
	}
	return nil
}

func AgentAssetObjectKey(owner, projectID, assetID, extension string) (string, error) {
	if strings.TrimSpace(owner) == "" || !assetSafeID.MatchString(projectID) || !assetSafeID.MatchString(assetID) {
		return "", ErrInvalidAsset
	}
	extension = strings.ToLower(strings.TrimPrefix(strings.TrimSpace(extension), "."))
	if !assetExtension.MatchString(extension) {
		return "", ErrInvalidAsset
	}
	sum := sha256.Sum256([]byte(strings.TrimSpace(owner)))
	return "agent-assets/" + hex.EncodeToString(sum[:12]) + "/" + projectID + "/" + assetID + "." + extension, nil
}
