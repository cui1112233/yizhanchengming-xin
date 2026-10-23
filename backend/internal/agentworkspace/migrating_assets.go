package agentworkspace

import (
	"context"
	"io"
	"strings"
	"time"
)

// MigratingAssetObjectStore keeps legacy local project assets usable after the
// primary store changes to TOS. A reference request copies a missing legacy
// object once so external generation providers receive a signed TOS URL.
type MigratingAssetObjectStore struct {
	Primary    AssetObjectStore
	References ReferenceURLProvider
	Legacy     AssetObjectStore
}

func NewMigratingAssetObjectStore(primary AssetObjectStore, references ReferenceURLProvider, legacy AssetObjectStore) *MigratingAssetObjectStore {
	return &MigratingAssetObjectStore{Primary: primary, References: references, Legacy: legacy}
}

func (s *MigratingAssetObjectStore) Put(ctx context.Context, asset Asset, body io.Reader) error {
	if s == nil || s.Primary == nil {
		return ErrInvalidAsset
	}
	return s.Primary.Put(ctx, asset, body)
}

func (s *MigratingAssetObjectStore) Open(ctx context.Context, asset Asset) (io.ReadCloser, error) {
	if s == nil || s.Primary == nil {
		return nil, ErrInvalidAsset
	}
	body, err := s.Primary.Open(ctx, asset)
	if err == nil || s.Legacy == nil {
		return body, err
	}
	return s.Legacy.Open(ctx, asset)
}

func (s *MigratingAssetObjectStore) Delete(ctx context.Context, asset Asset) error {
	if s == nil || s.Primary == nil {
		return ErrInvalidAsset
	}
	primaryErr := s.Primary.Delete(ctx, asset)
	if s.Legacy == nil {
		return primaryErr
	}
	legacyErr := s.Legacy.Delete(ctx, asset)
	if primaryErr != nil {
		return primaryErr
	}
	return legacyErr
}

func (s *MigratingAssetObjectStore) ReferenceURL(ctx context.Context, asset Asset, ttl time.Duration) (string, error) {
	if s == nil || s.Primary == nil || s.References == nil || strings.TrimSpace(asset.ObjectKey) == "" {
		return "", ErrReferenceURLUnavailable
	}
	body, err := s.Primary.Open(ctx, asset)
	if err == nil {
		_ = body.Close()
		return s.References.ReferenceURL(ctx, asset, ttl)
	}
	if s.Legacy == nil {
		return "", ErrReferenceURLUnavailable
	}
	legacyBody, legacyErr := s.Legacy.Open(ctx, asset)
	if legacyErr != nil {
		return "", ErrReferenceURLUnavailable
	}
	defer legacyBody.Close()
	if putErr := s.Primary.Put(ctx, asset, legacyBody); putErr != nil {
		return "", putErr
	}
	return s.References.ReferenceURL(ctx, asset, ttl)
}

var _ AssetObjectStore = (*MigratingAssetObjectStore)(nil)
var _ ReferenceURLProvider = (*MigratingAssetObjectStore)(nil)
