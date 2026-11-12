package httpapi

import (
	"context"
	"encoding/json"
	"qiantie/backend/internal/aicut"
	"qiantie/backend/internal/batchfactoryv11"
	"testing"
	"time"
)

func TestAICutNamedAgentHTTPRunMessagesEventsAndCancelAreScoped(t *testing.T) {
	_, base, c, now := collectionRouteFixture(t)
	a, _ := c.Create(t.Context(), "alice", "p", "A")
	agents := aicut.NewMemoryAgentStore()
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, AICut: base, AICutTimelines: c, AICutAgentStore: agents})
	root := "/api/ai-cut/projects/p/agent-runs"
	body := map[string]any{"instruction": "Explain", "mode": "ask", "clientRequestId": "request", "expectedRevision": 0}
	r := signedJSONRequestWithTimeline(t, api, now, "POST", root, a.ID, body)
	if r.Code != 202 {
		t.Fatalf("create=%d %s", r.Code, r.Body)
	}
	var p struct {
		Run aicut.AgentRun `json:"run"`
	}
	_ = json.Unmarshal(r.Body.Bytes(), &p)
	if p.Run.TimelineID != a.ID {
		t.Fatalf("run=%+v", p.Run)
	}
	_ = agents.AppendAgentMessage(t.Context(), aicut.AgentMessage{ID: "m", RunID: p.Run.ID, Owner: "alice", ProjectID: "p", Role: aicut.AgentRoleUser, Content: "Named message", CreatedAt: now})
	_, _ = agents.AppendAgentEvent(t.Context(), aicut.AgentEvent{RunID: p.Run.ID, Owner: "alice", ProjectID: "p", Type: aicut.AgentEventThinking, CreatedAt: now})
	for _, suffix := range []string{"", "/events", "/cancel"} {
		method := "GET"
		if suffix == "/cancel" {
			method = "POST"
		}
		r = signedJSONRequestWithTimeline(t, api, now, method, root+"/"+p.Run.ID+suffix, "default", map[string]any{})
		if r.Code != 404 {
			t.Fatalf("cross-scope %s=%d %s", suffix, r.Code, r.Body)
		}
	}
	for _, id := range []string{"default", a.ID} {
		r = signedJSONRequestWithTimeline(t, api, now, "GET", "/api/ai-cut/projects/p/agent-messages", id, nil)
		if r.Code != 200 || containsText(r.Body.String(), "Named message") != (id == a.ID) {
			t.Fatalf("messages scope=%s %d %s", id, r.Code, r.Body)
		}
	}
	r = signedJSONRequestWithTimeline(t, api, now, "POST", root+"/"+p.Run.ID+"/cancel", a.ID, map[string]any{})
	if r.Code != 200 {
		t.Fatalf("cancel=%d %s", r.Code, r.Body)
	}
}
func TestAICutNamedEditPlanHTTPPlanApplyAndReplayAreScoped(t *testing.T) {
	_, base, c, now := collectionRouteFixture(t)
	a, _ := c.Create(t.Context(), "alice", "p", "A")
	_, _ = c.Save(t.Context(), "alice", "p", a.ID, aicut.Timeline{Tracks: []aicut.Track{{ID: "subtitle-1", Kind: aicut.TrackSubtitle, Clips: []aicut.Clip{}}}}, 0, now)
	provider := &editPlanTestProvider{result: `{"summary":{"sourceDurationMs":1000,"estimatedDurationMs":1000},"operations":[{"id":"caption","type":"add_caption","trackId":"subtitle-1","startMs":0,"durationMs":1000,"text":"Named planned caption","reason":"Caption","enabled":true}],"warnings":[]}`}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, AICut: base, AICutTimelines: c, TextProviderResolver: func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) { return provider, nil }})
	root := "/api/ai-cut/projects/p/edit-plans"
	body := map[string]any{"clientRequestId": "plan-request", "modelId": "model", "expectedRevision": 1, "assetIds": []string{"asset"}, "goal": "Caption", "targetDurationMs": 1000, "aspectRatio": "9:16"}
	r := signedJSONRequestWithTimeline(t, api, now, "POST", root, a.ID, body)
	if r.Code != 201 {
		t.Fatalf("create=%d %s", r.Code, r.Body)
	}
	var p struct {
		Plan aicut.EditPlan `json:"plan"`
	}
	_ = json.Unmarshal(r.Body.Bytes(), &p)
	if p.Plan.TimelineID != a.ID {
		t.Fatalf("plan=%+v", p.Plan)
	}
	for _, suffix := range []string{"", "/cancel", "/apply"} {
		method := "GET"
		if suffix != "" {
			method = "POST"
		}
		r = signedJSONRequestWithTimeline(t, api, now, method, root+"/"+p.Plan.ID+suffix, "default", map[string]any{})
		if r.Code != 404 {
			t.Fatalf("cross-scope %s=%d %s", suffix, r.Code, r.Body)
		}
	}
	r = signedJSONRequestWithTimeline(t, api, now, "POST", root+"/"+p.Plan.ID+"/apply", a.ID, map[string]any{})
	if r.Code != 200 {
		t.Fatalf("apply=%d %s", r.Code, r.Body)
	}
	v, _ := c.Timeline(t.Context(), "alice", "p", a.ID)
	legacy, _ := base.Timeline(t.Context(), "alice", "p")
	if v.Revision != 2 || v.Tracks[0].Clips[0].Text != "Named planned caption" || legacy.Revision != 0 {
		t.Fatalf("named=%+v legacy=%+v", v, legacy)
	}
	r = signedJSONRequestWithTimeline(t, api, now, "POST", root, a.ID, body)
	if r.Code != 200 || provider.calls != 1 {
		t.Fatalf("replay=%d calls=%d %s", r.Code, provider.calls, r.Body)
	}
}
