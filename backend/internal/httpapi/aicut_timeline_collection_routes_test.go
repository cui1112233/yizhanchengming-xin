package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"qiantie/backend/internal/aicut"
	"testing"
	"time"
)

func collectionRouteFixture(t *testing.T) (*http.ServeMux, *aicut.MemoryStore, *aicut.TimelineCollectionStore, time.Time) {
	t.Helper()
	now := time.Date(2026, 11, 12, 9, 0, 0, 0, time.UTC)
	base := aicut.NewMemoryStore()
	if e := base.CreateProject(t.Context(), "alice", aicut.Project{ID: "p", Title: "P", CreatedAt: now, UpdatedAt: now}); e != nil {
		t.Fatal(e)
	}
	s := aicut.NewMemoryTimelineCollectionStore(base)
	mux := http.NewServeMux()
	RegisterAICutTimelineCollectionRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, s, func() time.Time { return now })
	return mux, base, s, now
}
func createRouteTimeline(t *testing.T, mux http.Handler, now time.Time, title string) aicut.TimelineMetadata {
	t.Helper()
	r := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/ai-cut/projects/p/timelines", map[string]any{"title": title})
	if r.Code != http.StatusCreated {
		t.Fatalf("create=%d %s", r.Code, r.Body)
	}
	var out struct {
		Timeline aicut.TimelineMetadata `json:"timeline"`
	}
	if e := json.Unmarshal(r.Body.Bytes(), &out); e != nil {
		t.Fatal(e)
	}
	if out.Timeline.ID == "" {
		t.Fatal("missing server ID")
	}
	return out.Timeline
}
func TestAICutTimelineCollectionRoutesIsolationHistoryAndAuth(t *testing.T) {
	mux, base, _, now := collectionRouteFixture(t)
	a := createRouteTimeline(t, mux, now, "A")
	path := "/api/ai-cut/projects/p/timelines/" + a.ID + "/snapshot"
	save := func(rev int64) *httptest.ResponseRecorder {
		return signedJSONRequest(t, mux, now, "alice", http.MethodPut, path, map[string]any{"expectedRevision": rev, "timeline": map[string]any{"script": "named", "tracks": []any{}}})
	}
	r := save(0)
	if r.Code != 200 {
		t.Fatalf("save=%d %s", r.Code, r.Body)
	}
	if r = save(0); r.Code != 409 {
		t.Fatalf("conflict=%d", r.Code)
	}
	legacy, _ := base.Timeline(t.Context(), "alice", "p")
	if legacy.Script != "" || legacy.Revision != 0 {
		t.Fatalf("legacy=%+v", legacy)
	}
	r = signedJSONRequest(t, mux, now, "bob", "GET", path, nil)
	if r.Code != 404 {
		t.Fatalf("owner=%d", r.Code)
	}
	unauth := httptest.NewRecorder()
	mux.ServeHTTP(unauth, httptest.NewRequest("GET", path, nil))
	if unauth.Code != 401 {
		t.Fatalf("unauth=%d", unauth.Code)
	}
	hp := "/api/ai-cut/projects/p/timelines/" + a.ID + "/history"
	r = signedJSONRequest(t, mux, now, "alice", "POST", hp+"/undo", map[string]any{"expectedRevision": 1})
	if r.Code != 200 || !containsText(r.Body.String(), `"canRedo":true`) {
		t.Fatalf("undo=%d %s", r.Code, r.Body)
	}
	r = signedJSONRequest(t, mux, now, "alice", "POST", hp+"/redo", map[string]any{"expectedRevision": 2})
	if r.Code != 200 || !containsText(r.Body.String(), `"script":"named"`) {
		t.Fatalf("redo=%d %s", r.Code, r.Body)
	}
}
func TestAICutTimelineCollectionRoutesCRUDReorderLimits(t *testing.T) {
	mux, _, s, now := collectionRouteFixture(t)
	a := createRouteTimeline(t, mux, now, "A")
	root := "/api/ai-cut/projects/p/timelines"
	r := signedJSONRequest(t, mux, now, "alice", "PATCH", root+"/"+a.ID, map[string]any{"title": "Renamed"})
	if r.Code != 200 {
		t.Fatalf("rename=%d %s", r.Code, r.Body)
	}
	r = signedJSONRequest(t, mux, now, "alice", "POST", root+"/reorder", map[string]any{"timelineIds": []string{a.ID, "default"}})
	if r.Code != 200 {
		t.Fatalf("reorder=%d %s", r.Code, r.Body)
	}
	r = signedJSONRequest(t, mux, now, "alice", "GET", root, nil)
	if r.Code != 200 || !containsText(r.Body.String(), `"title":"Renamed"`) {
		t.Fatalf("list=%d %s", r.Code, r.Body)
	}
	for _, method := range []string{"PATCH", "DELETE"} {
		r = signedJSONRequest(t, mux, now, "alice", method, root+"/default", map[string]any{"title": "Changed"})
		if r.Code != 400 {
			t.Fatalf("default %s=%d", method, r.Code)
		}
	}
	r = signedJSONRequest(t, mux, now, "alice", "POST", root+"/reorder", map[string]any{"timelineIds": []string{a.ID, a.ID}})
	if r.Code != 400 {
		t.Fatalf("bad order=%d", r.Code)
	}
	r = signedJSONRequest(t, mux, now, "alice", "DELETE", root+"/"+a.ID, nil)
	if r.Code != 204 {
		t.Fatalf("delete=%d", r.Code)
	}
	r = signedJSONRequest(t, mux, now, "alice", "GET", root+"/"+a.ID+"/snapshot", nil)
	if r.Code != 404 {
		t.Fatalf("deleted=%d", r.Code)
	}
	for i := 1; i < 50; i++ {
		if _, e := s.Create(t.Context(), "alice", "p", "More"); e != nil {
			t.Fatal(e)
		}
	}
	r = signedJSONRequest(t, mux, now, "alice", "POST", root, map[string]any{"title": "Overflow"})
	if r.Code != 409 {
		t.Fatalf("limit=%d %s", r.Code, r.Body)
	}
}
func TestAICutTimelineCollectionRoutesDuplicatePreservesDefault(t *testing.T) {
	mux, base, _, now := collectionRouteFixture(t)
	_, e := base.SaveTimeline(t.Context(), "alice", "p", aicut.Timeline{Script: "legacy", Markers: []aicut.TimelineMarker{{ID: "original", Label: "Cue"}}}, 0, now)
	if e != nil {
		t.Fatal(e)
	}
	r := signedJSONRequest(t, mux, now, "alice", "POST", "/api/ai-cut/projects/p/timelines/default/duplicate", map[string]any{"title": "Copy"})
	if r.Code != 201 {
		t.Fatalf("duplicate=%d %s", r.Code, r.Body)
	}
	var p struct {
		Timeline aicut.TimelineMetadata `json:"timeline"`
	}
	_ = json.Unmarshal(r.Body.Bytes(), &p)
	r = signedJSONRequest(t, mux, now, "alice", "GET", "/api/ai-cut/projects/p/timelines/"+p.Timeline.ID+"/snapshot", nil)
	var snapshot struct {
		Timeline aicut.Timeline `json:"timeline"`
	}
	_ = json.Unmarshal(r.Body.Bytes(), &snapshot)
	if r.Code != 200 || snapshot.Timeline.Script != "legacy" || snapshot.Timeline.Revision != 0 || len(snapshot.Timeline.Markers) != 1 || snapshot.Timeline.Markers[0].ID == "original" {
		t.Fatalf("copy=%d %s", r.Code, r.Body)
	}
	v, _ := base.Timeline(t.Context(), "alice", "p")
	if v.Revision != 1 || v.Markers[0].ID != "original" {
		t.Fatalf("default=%+v", v)
	}
}

func TestAICutTimelineCollectionRouterMemoryFallback(t *testing.T) {
	_, base, _, now := collectionRouteFixture(t)
	router := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, AICut: base})
	r := signedJSONRequest(t, router, now, "alice", "GET", "/api/ai-cut/projects/p/timelines", nil)
	if r.Code != 200 || !containsText(r.Body.String(), `"id":"default"`) {
		t.Fatalf("router=%d %s", r.Code, r.Body)
	}
	a := createRouteTimeline(t, router, now, "Persisted memory")
	r = signedJSONRequest(t, router, now, "alice", "GET", "/api/ai-cut/projects/p/timelines/"+a.ID+"/snapshot", nil)
	if r.Code != 200 {
		t.Fatalf("fallback persistence=%d %s", r.Code, r.Body)
	}
}

func TestAICutNamedTimelineExportHTTPUsesSelectedSnapshot(t *testing.T) {
	_, base, collection, now := collectionRouteFixture(t)
	a, _ := collection.Create(t.Context(), "alice", "p", "A")
	_, _ = collection.Save(t.Context(), "alice", "p", a.ID, aicut.Timeline{Script: "selected"}, 0, now)
	svc := NewAICutExportService(base, exportTestObjects{}, &exportTestNotifier{}, func() time.Time { return now })
	router := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, AICut: base, AICutTimelines: collection, AICutExport: svc})
	r := signedJSONRequestWithTimeline(t, router, now, "POST", "/api/ai-cut/projects/p/export-jobs", a.ID, map[string]any{"format": "mp3", "expectedRevision": 1})
	if r.Code != 202 {
		t.Fatalf("submit=%d %s", r.Code, r.Body)
	}
	var out struct {
		Job aicut.ExportJob `json:"job"`
	}
	_ = json.Unmarshal(r.Body.Bytes(), &out)
	persisted, e := base.ExportJob(t.Context(), "alice", "p", out.Job.ID)
	if e != nil || persisted.Options.TimelineID != a.ID || persisted.TimelineSnapshot.Script != "selected" {
		t.Fatalf("persisted=%+v %v", persisted, e)
	}
	r = signedJSONRequestWithTimeline(t, router, now, "GET", "/api/ai-cut/projects/p/export-jobs/"+out.Job.ID, "default", nil)
	if r.Code != 404 {
		t.Fatalf("default reads named=%d", r.Code)
	}
	r = signedJSONRequestWithTimeline(t, router, now, "POST", "/api/ai-cut/projects/p/unsupported-tool", a.ID, map[string]any{})
	if r.Code != 409 {
		t.Fatalf("agent guard=%d", r.Code)
	}
}

func TestAICutNamedTimelineExportHTTPRejectsMissingTarget(t *testing.T) {
	_, base, collection, now := collectionRouteFixture(t)
	svc := NewAICutExportService(base, exportTestObjects{}, &exportTestNotifier{}, func() time.Time { return now }, collection)
	router := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, AICut: base, AICutTimelines: collection, AICutExport: svc})
	r := signedJSONRequestWithTimeline(t, router, now, "POST", "/api/ai-cut/projects/p/export-jobs", "missing", map[string]any{"format": "mp3", "expectedRevision": 0})
	if r.Code != 404 {
		t.Fatalf("missing target=%d %s", r.Code, r.Body)
	}
}

func TestAICutNamedSynchronousToolsImportAndCorrectSelectedTimeline(t *testing.T) {
	_, base, collection, now := collectionRouteFixture(t)
	a, _ := collection.Create(t.Context(), "alice", "p", "A")
	router := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, AICut: base, AICutTimelines: collection})
	r := signedJSONRequestWithTimeline(t, router, now, "POST", "/api/ai-cut/projects/p/subtitles/import", a.ID, map[string]any{"filename": "caption.srt", "srt": "1\n00:00:00,000 --> 00:00:01,000\nBefore\n", "expectedRevision": 0})
	if r.Code != 200 {
		t.Fatalf("import=%d %s", r.Code, r.Body)
	}
	v, _ := collection.Timeline(t.Context(), "alice", "p", a.ID)
	if len(v.Tracks) != 1 {
		t.Fatal(v)
	}
	clip := v.Tracks[0].Clips[0]
	r = signedJSONRequestWithTimeline(t, router, now, "POST", "/api/ai-cut/projects/p/captions/correct", a.ID, map[string]any{"expectedRevision": 1, "corrections": []map[string]any{{"clipId": clip.ID, "text": "After"}}})
	if r.Code != 200 {
		t.Fatalf("correct=%d %s", r.Code, r.Body)
	}
	v, _ = collection.Timeline(t.Context(), "alice", "p", a.ID)
	d, _ := base.Timeline(t.Context(), "alice", "p")
	if v.Revision != 2 || v.Tracks[0].Clips[0].Text != "After" || d.Revision != 0 || len(d.Tracks) != 0 {
		t.Fatalf("named=%+v default=%+v", v, d)
	}
	r = signedJSONRequestWithTimeline(t, router, now, "PATCH", "/api/ai-cut/projects/p", a.ID, map[string]any{"backgroundMode": "transparent"})
	if r.Code != 200 {
		t.Fatalf("shared appearance=%d %s", r.Code, r.Body)
	}
}
