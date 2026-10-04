package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/media"
)

type mediaResolverFake struct {
	owner string
	id string
	result media.ResolvedAsset
	err error
}
func (f *mediaResolverFake) Resolve(_ context.Context, owner, id string) (media.ResolvedAsset, error) {
	f.owner, f.id = owner, id
	return f.result, f.err
}

func TestMediaHandlerReturnsSignedPreview(t *testing.T) {
	resolver := &mediaResolverFake{result: media.ResolvedAsset{
		Asset: media.Reference{ID: "asset_1", MediaType: media.TypeImage, MimeType: "image/png", StorageKey: "tos://bucket/images/a.png"},
		PreviewURL: "https://signed.example/a.png",
	}}
	handler := NewMediaHandler(resolver, testOwner)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/media/assets/asset_1", nil))
	if rec.Code != http.StatusOK { t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String()) }
	if resolver.owner != "owner" || resolver.id != "asset_1" { t.Fatalf("owner=%q id=%q", resolver.owner, resolver.id) }
	if body := rec.Body.String(); body == "" || !containsAll(body, "asset_1", "https://signed.example/a.png") { t.Fatalf("body=%s", body) }
}

func TestMediaHandlerMapsNotFound(t *testing.T) {
	handler := NewMediaHandler(&mediaResolverFake{err: media.ErrNotFound}, testOwner)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/media/assets/missing", nil))
	if rec.Code != http.StatusNotFound { t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String()) }
}

func TestMediaHandlerRejectsUnauthorizedRequest(t *testing.T) {
	handler := NewMediaHandler(&mediaResolverFake{}, func(*http.Request) (string, error) { return "", errors.New("no") })
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/media/assets/asset_1", nil))
	if rec.Code != http.StatusUnauthorized { t.Fatalf("status=%d", rec.Code) }
}

func containsAll(value string, parts ...string) bool {
	for _, part := range parts {
		if !strings.Contains(value, part) { return false }
	}
	return true
}
