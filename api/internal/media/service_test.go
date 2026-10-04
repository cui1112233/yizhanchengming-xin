package media

import (
	"context"
	"database/sql"
	"testing"
)

type readerFake struct {
	asset Asset
	err error
}
func (f readerFake) Get(context.Context, string, string) (Asset, error) { return f.asset, f.err }

type signerFake struct {
	bucket string
	key string
	expires int64
	url string
	err error
}
func (f *signerFake) Bucket() string { return f.bucket }
func (f *signerFake) SignedGetURL(key string, expires int64) (string, error) {
	f.key, f.expires = key, expires
	return f.url, f.err
}

func TestServiceResolveReturnsSignedPreviewForOwnerAsset(t *testing.T) {
	signer := &signerFake{bucket: "prod-media", url: "https://signed.example/image"}
	service := Service{
		Assets: readerFake{asset: Asset{ID: "asset_1", Owner: "owner", MediaType: TypeImage, TOSBucket: "prod-media", TOSKey: "images/a.png", MimeType: "image/png"}},
		Signer: signer,
	}
	resolved, err := service.Resolve(context.Background(), "owner", "asset_1")
	if err != nil { t.Fatal(err) }
	if resolved.Asset.ID != "asset_1" || resolved.PreviewURL != "https://signed.example/image" { t.Fatalf("resolved=%#v", resolved) }
	if signer.key != "images/a.png" || signer.expires != DefaultPreviewTTLSeconds { t.Fatalf("signer=%#v", signer) }
}

func TestServiceResolveRejectsBucketMismatch(t *testing.T) {
	service := Service{
		Assets: readerFake{asset: Asset{ID: "asset_1", TOSBucket: "other", TOSKey: "a.png"}},
		Signer: &signerFake{bucket: "prod-media", url: "https://signed.example"},
	}
	if _, err := service.Resolve(context.Background(), "owner", "asset_1"); err == nil { t.Fatal("expected bucket mismatch") }
}

func TestServiceResolveMapsMissingAsset(t *testing.T) {
	service := Service{Assets: readerFake{err: sql.ErrNoRows}, Signer: &signerFake{bucket: "b"}}
	if _, err := service.Resolve(context.Background(), "owner", "missing"); err != ErrNotFound { t.Fatalf("err=%v", err) }
}
