package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"qiantie/backend/internal/screenplay"
	"strings"
	"testing"
)

func TestScreenplayLibraryCategoriesHTTP(t *testing.T) {
	_, _, _, mux, now := assetLibraryServiceFixture(t)
	path := "/api/comic-screenplays/asset-library/categories"
	get := signedJSONRequest(t, mux, now, "alice", "GET", path, nil)
	if get.Code != 200 {
		t.Fatalf("missing category endpoint: %d %s", get.Code, get.Body.String())
	}
	var state struct {
		Builtin  []string `json:"builtin_categories"`
		Custom   []string `json:"custom_categories"`
		Revision int64    `json:"revision"`
	}
	if err := json.Unmarshal(get.Body.Bytes(), &state); err != nil {
		t.Fatal(err)
	}
	if len(state.Builtin) != 13 || state.Custom == nil || state.Revision != 0 {
		t.Fatal(get.Body.String())
	}
	for _, test := range []struct {
		action, name string
		rev          int64
		status       int
	}{
		{"add", " 自创 ", 0, 200}, {"add", " 自创 ", 0, 200}, {"add", "第二个", 0, 409},
		{"delete", "男主", 1, 400}, {"delete", " 自创 ", 1, 200}, {"delete", " 自创 ", 1, 200},
		{"add", "", 2, 400}, {"bad", "分类", 2, 400}, {"add", "分类", -1, 400},
	} {
		response := signedJSONRequest(t, mux, now, "alice", "POST", path, map[string]any{"action": test.action, "name": test.name, "expectedRevision": test.rev})
		if response.Code != test.status {
			t.Fatalf("%+v: %d %s", test, response.Code, response.Body.String())
		}
	}
	bob := signedJSONRequest(t, mux, now, "bob", "GET", path, nil)
	if err := json.Unmarshal(bob.Body.Bytes(), &state); err != nil || len(state.Custom) != 0 || state.Revision != 0 {
		t.Fatal(bob.Body.String(), err)
	}
	for _, payload := range []map[string]any{{"action": "add", "name": "x"}, {"action": "add", "name": "x", "expectedRevision": 1.2}} {
		res := signedJSONRequest(t, mux, now, "alice", "POST", path, payload)
		if res.Code != 400 {
			t.Fatal(res.Code, res.Body.String())
		}
	}
}

type failingCategoryLibrary struct {
	screenplay.AssetLibraryRepository
}

func (f failingCategoryLibrary) LoadAssetLibraryCategories(context.Context, string) (screenplay.AssetLibraryCategoryState, error) {
	return screenplay.AssetLibraryCategoryState{}, errors.New("database internal secret")
}
func (f failingCategoryLibrary) ChangeAssetLibraryCategory(context.Context, string, string, string, int64) (screenplay.AssetLibraryCategoryState, error) {
	return screenplay.AssetLibraryCategoryState{}, errors.New("database internal secret")
}
func TestScreenplayLibraryCategoriesUnavailableAndStorageErrors(t *testing.T) {
	service, _, _, mux, now := assetLibraryServiceFixture(t)
	path := "/api/comic-screenplays/asset-library/categories"
	original := service.Repository
	service.Repository = &commitLostAssetLibraryRepository{AssetLibraryRepository: original}
	response := signedJSONRequest(t, mux, now, "alice", "GET", path, nil)
	if response.Code != 503 {
		t.Fatal(response.Code, response.Body.String())
	}
	service.Repository = failingCategoryLibrary{original}
	for _, method := range []string{"GET", "POST"} {
		response = signedJSONRequest(t, mux, now, "alice", method, path, map[string]any{"action": "add", "name": "x", "expectedRevision": 0})
		if response.Code != 500 || strings.Contains(response.Body.String(), "secret") || !strings.Contains(response.Body.String(), "请重试") {
			t.Fatal(response.Code, response.Body.String())
		}
	}
}
