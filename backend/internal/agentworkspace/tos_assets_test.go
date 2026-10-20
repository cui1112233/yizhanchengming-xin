package agentworkspace

import (
	"context"
	"testing"
	"time"
)

func TestTOSAssetReferenceURLUsesInjectedShortLivedHTTPSURL(t *testing.T) {
	store := &TOSAssetObjectStore{Bucket: "agent-assets", referenceURL: func(_ context.Context, asset Asset, ttl time.Duration) (string, error) {
		if asset.ID != "asset-1" || ttl != 15*time.Minute {
			t.Fatalf("asset=%+v ttl=%s", asset, ttl)
		}
		return "https://tos.example/agent-assets/asset-1.png?X-Tos-Signature=test", nil
	}}
	value, err := store.ReferenceURL(context.Background(), Asset{ID: "asset-1", ObjectKey: "agent-assets/asset-1.png"}, 15*time.Minute)
	if err != nil || value == "" {
		t.Fatalf("value=%q err=%v", value, err)
	}
}

func TestTOSAssetReferenceURLRejectsUnavailableStore(t *testing.T) {
	if _, err := (*TOSAssetObjectStore)(nil).ReferenceURL(context.Background(), Asset{ID: "asset-1"}, time.Minute); err != ErrReferenceURLUnavailable {
		t.Fatalf("err=%v", err)
	}
}
