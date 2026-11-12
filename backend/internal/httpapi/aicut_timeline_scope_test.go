package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAICutNamedTimelineGuardBlocksLegacyTools(t *testing.T) {
	now := time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)
	hit := false
	h := withAICutTimelineScope(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { hit = true; w.WriteHeader(204) }), BridgeAuth{Secret: "secret", Now: func() time.Time { return now }})
	for _, suffix := range []string{"timeline", "timeline/history/undo", "exports/mp4", "unknown-future-tool"} {
		t.Run(suffix, func(t *testing.T) {
			hit = false
			r := signedJSONRequestWithTimeline(t, h, now, "POST", "/api/ai-cut/projects/p/"+suffix, "named", nil)
			if r.Code != 409 || hit {
				t.Fatalf("guard=%d hit=%v %s", r.Code, hit, r.Body)
			}
		})
	}
	req := httptest.NewRequest("POST", "/api/ai-cut/projects/p/unsupported-tool", nil)
	req.Header.Set("X-AICut-Timeline-Id", "named")
	out := httptest.NewRecorder()
	h.ServeHTTP(out, req)
	if out.Code != 401 {
		t.Fatalf("unauth=%d", out.Code)
	}
}
func TestAICutNamedTimelineGuardAllowsExplicitSharedResources(t *testing.T) {
	now := time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)
	h := withAICutTimelineScope(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if aicutTimelineIDFromContext(r.Context()) != "named" {
			t.Error("missing immutable context")
		}
		w.WriteHeader(204)
	}), BridgeAuth{Secret: "secret", Now: func() time.Time { return now }})
	for _, tc := range []struct{ method, path string }{{"GET", "/api/ai-cut/projects/p"}, {"PATCH", "/api/ai-cut/projects/p"}, {"GET", "/api/ai-cut/projects/p/assets"}, {"POST", "/api/ai-cut/projects/p/assets"}, {"GET", "/api/ai-cut/projects/p/assets/a/content"}, {"GET", "/api/ai-cut/projects/p/audio-analysis"}, {"POST", "/api/ai-cut/projects/p/visual-frames"}, {"POST", "/api/ai-cut/projects/p/tracking-frames"}, {"POST", "/api/ai-cut/projects/p/multicam-sync"}, {"POST", "/api/ai-cut/projects/p/asr-jobs"}, {"GET", "/api/ai-cut/projects/p/timelines/named/snapshot"}, {"POST", "/api/ai-cut/projects/p/export-jobs"}} {
		r := signedJSONRequestWithTimeline(t, h, now, tc.method, tc.path, "named", nil)
		if r.Code != 204 {
			t.Fatalf("%s %s=%d", tc.method, tc.path, r.Code)
		}
	}
}
func signedJSONRequestWithTimeline(t *testing.T, h http.Handler, now time.Time, method, path, id string, body any) *httptest.ResponseRecorder {
	t.Helper()
	wrapped := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("X-AICut-Timeline-Id", id)
		h.ServeHTTP(w, r)
	})
	return signedJSONRequest(t, wrapped, now, "alice", method, path, body)
}
