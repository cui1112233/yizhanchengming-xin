package media

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

const DefaultPreviewTTLSeconds int64 = 900

var ErrNotFound = errors.New("media asset not found")

type AssetReader interface {
	Get(context.Context, string, string) (Asset, error)
}

type PreviewSigner interface {
	Bucket() string
	SignedGetURL(string, int64) (string, error)
}

type ResolvedAsset struct {
	Asset      Reference `json:"asset"`
	PreviewURL string    `json:"preview_url"`
}

type Service struct {
	Assets AssetReader
	Signer PreviewSigner
}

func (s Service) Resolve(ctx context.Context, owner, id string) (ResolvedAsset, error) {
	if s.Assets == nil || s.Signer == nil {
		return ResolvedAsset{}, errors.New("media service unavailable")
	}
	owner = strings.TrimSpace(owner)
	id = strings.TrimSpace(id)
	if owner == "" || id == "" {
		return ResolvedAsset{}, errors.New("owner and media asset id are required")
	}
	asset, err := s.Assets.Get(ctx, owner, id)
	if errors.Is(err, sql.ErrNoRows) {
		return ResolvedAsset{}, ErrNotFound
	}
	if err != nil { return ResolvedAsset{}, err }
	if strings.TrimSpace(asset.TOSBucket) != strings.TrimSpace(s.Signer.Bucket()) {
		return ResolvedAsset{}, errors.New("media asset belongs to a different TOS bucket")
	}
	previewURL, err := s.Signer.SignedGetURL(asset.TOSKey, DefaultPreviewTTLSeconds)
	if err != nil { return ResolvedAsset{}, err }
	return ResolvedAsset{Asset: asset.Reference(), PreviewURL: previewURL}, nil
}
