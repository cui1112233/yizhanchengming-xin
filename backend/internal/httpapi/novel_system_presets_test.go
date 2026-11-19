package httpapi

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/auth"
	"qiantie/backend/internal/prompt"
	novelprompt "qiantie/backend/internal/prompt/novelcreation"
)

func TestNovelSystemPresetsAdminPublicationAndPermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "presets.json")
	if e := os.WriteFile(path, []byte(`[]`), 0600); e != nil {
		t.Fatal(e)
	}
	if e := novelprompt.EnsureAdminPresets(path); e != nil {
		t.Fatal(e)
	}
	catalog := prompt.NewCatalog(path)
	store := &governanceAuthStore{accounts: map[string]auth.Account{
		"owner":  {Username: "owner", Active: true, IsOwner: true},
		"member": {Username: "member", Active: true},
	}}
	now := time.Now().UTC()
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, AuthStore: store, MemberProfileStore: NewMemberProfileStore(t.TempDir()), PresetCatalog: catalog})
	for _, url := range []string{"/api/admin/presets?module=novel-creation", "/api/admin/preset-slots?module=novel-creation"} {
		if r := signedJSONRequest(t, api, now, "member", http.MethodGet, url, nil); r.Code != http.StatusForbidden || strings.Contains(r.Body.String(), novelprompt.ChapterDraft.SystemPrompt) {
			t.Fatal("member could read admin prompts", r.Code)
		}
		r := signedJSONRequest(t, api, now, "owner", http.MethodGet, url, nil)
		if r.Code != http.StatusOK || !strings.Contains(r.Body.String(), "novel.") {
			t.Fatal(url, r.Code, r.Body.String())
		}
	}
	body, e := prompt.NewAdminStore(path).Get(novelprompt.ChapterDraft.ID, 1)
	if e != nil {
		t.Fatal(e)
	}
	body["body"] = "ADMIN_HTTP_PUBLISHED"
	r := signedJSONRequest(t, api, now, "owner", http.MethodPost, "/api/admin/presets/draft", body)
	if r.Code != http.StatusCreated {
		t.Fatal(r.Code, r.Body.String())
	}
	var response struct {
		Preset struct {
			Version int `json:"version"`
		} `json:"preset"`
	}
	if e = json.Unmarshal(r.Body.Bytes(), &response); e != nil {
		t.Fatal(e)
	}
	r = signedJSONRequest(t, api, now, "owner", http.MethodPost, "/api/admin/presets/"+novelprompt.ChapterDraft.ID+"/publish", map[string]any{"version": response.Preset.Version})
	if r.Code != http.StatusOK {
		t.Fatal(r.Code, r.Body.String())
	}
	rows, e := catalog.PublishedBodies(novelprompt.AdminModule)
	if e != nil {
		t.Fatal(e)
	}
	for _, p := range rows {
		if p.ID == novelprompt.ChapterDraft.ID && p.Body == "ADMIN_HTTP_PUBLISHED" && p.Version == response.Preset.Version {
			return
		}
	}
	t.Fatal("admin API publication did not reach runtime catalog")
}
