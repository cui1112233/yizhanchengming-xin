package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"qiantie/backend/internal/aicut"
	"strings"
	"testing"
	"time"
)

func personalTemplateInput() map[string]any {
	return map[string]any{"title": "我的模板", "layout": map[string]any{"version": 1, "slots": []any{map[string]any{"id": "slot-1", "kind": "video", "label": "素材 1"}}, "tracks": []any{map[string]any{"id": "track-1", "kind": "video", "name": "画面", "order": 0, "clips": []any{map[string]any{"id": "clip-1", "kind": "video", "name": "片段", "assetId": "slot-1", "startMs": 0, "durationMs": 1000, "sourceStartMs": 0, "sourceDurationMs": 2000}}}}}}
}
func TestAICutPersonalTemplatesOwnerPersistenceAndAuth(t *testing.T) {
	now := time.Date(2026, 11, 12, 1, 0, 0, 0, time.UTC)
	store := aicut.NewMemoryPersonalTemplateStore()
	mux := http.NewServeMux()
	auth := BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}
	RegisterAICutPersonalTemplateRoutes(mux, auth, store)
	root := "/api/ai-cut/personal-templates"
	r := signedJSONRequest(t, mux, now, "alice", "POST", root, personalTemplateInput())
	if r.Code != 201 {
		t.Fatalf("create=%d %s", r.Code, r.Body)
	}
	var created struct {
		Template aicut.PersonalTemplate `json:"template"`
	}
	if e := json.Unmarshal(r.Body.Bytes(), &created); e != nil {
		t.Fatal(e)
	}
	if created.Template.ID == "" {
		t.Fatal("missing ID")
	}
	mux2 := http.NewServeMux()
	RegisterAICutPersonalTemplateRoutes(mux2, auth, store)
	for _, owner := range []string{"alice", "bob"} {
		r = signedJSONRequest(t, mux2, now, owner, "GET", root, nil)
		if r.Code != 200 || strings.Contains(r.Body.String(), created.Template.ID) != (owner == "alice") {
			t.Fatalf("isolation %s=%d %s", owner, r.Code, r.Body)
		}
	}
	if r = signedJSONRequest(t, mux, now, "bob", "DELETE", root+"/"+created.Template.ID, nil); r.Code != 404 {
		t.Fatalf("foreign delete=%d", r.Code)
	}
	r = signedJSONRequest(t, mux, now, "alice", "GET", root, nil)
	if !strings.Contains(r.Body.String(), created.Template.ID) {
		t.Fatal("foreign delete changed record")
	}
	unauth := httptest.NewRecorder()
	mux.ServeHTTP(unauth, httptest.NewRequest("GET", root, nil))
	if unauth.Code != 401 {
		t.Fatalf("auth=%d", unauth.Code)
	}
	if r = signedJSONRequest(t, mux, now, "alice", "DELETE", root+"/"+created.Template.ID, nil); r.Code != 204 {
		t.Fatalf("delete=%d", r.Code)
	}
}
func TestAICutPersonalTemplatesRejectPrivateReferences(t *testing.T) {
	now := time.Now().UTC()
	mux := http.NewServeMux()
	RegisterAICutPersonalTemplateRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, aicut.NewMemoryPersonalTemplateStore())
	for _, field := range []string{"url", "assetId", "linkGroupId", "compoundId", "transitionTrackId"} {
		input := personalTemplateInput()
		layout := input["layout"].(map[string]any)
		clip := layout["tracks"].([]any)[0].(map[string]any)["clips"].([]any)[0].(map[string]any)
		clip[field] = "private-reference"
		r := signedJSONRequest(t, mux, now, "alice", "POST", "/api/ai-cut/personal-templates", input)
		if r.Code != 400 {
			t.Fatalf("%s=%d %s", field, r.Code, r.Body)
		}
	}
}
