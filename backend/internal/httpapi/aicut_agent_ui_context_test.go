package httpapi

import (
	"context"
	"errors"
	"net/http"
	"qiantie/backend/internal/aicut"
	"testing"
	"time"
)

type failingUIEventStore struct{ aicut.AgentStore }

func (s failingUIEventStore) AppendAgentEvent(context.Context, aicut.AgentEvent) (aicut.AgentEvent, error) {
	return aicut.AgentEvent{}, errors.New("write failed")
}
func TestAICutAgentUIContextBeforeQueueAndFailure(t *testing.T) {
	for _, kind := range []string{"valid", "invalid", "write-failure"} {
		t.Run(kind, func(t *testing.T) {
			now := time.Now()
			timelines := aicut.NewMemoryStore()
			_ = timelines.CreateProject(t.Context(), "alice", aicut.Project{ID: "cut-1", Title: "test"})
			base := aicut.NewMemoryAgentStore()
			var agents aicut.AgentStore = base
			if kind == "write-failure" {
				agents = failingUIEventStore{base}
			}
			mux := http.NewServeMux()
			started := 0
			RegisterAICutAgentRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, timelines, agents, func(ctx context.Context, id string) error {
				started++
				events, _ := base.AgentEvents(ctx, "alice", "cut-1", id, 0)
				if len(events) != 1 || events[0].Type != "ui_context" || !containsText(events[0].Message, `"playheadMs":1234`) {
					t.Errorf("context not persisted before queue: %#v", events)
				}
				return nil
			}, func() time.Time { return now })
			clips := []any{}
			if kind == "invalid" {
				clips = append(clips, map[string]any{"trackId": "other", "clipId": "missing"})
			}
			input := map[string]any{"instruction": "剪这段", "expectedRevision": 0, "clientRequestId": "ui-1", "uiContext": map[string]any{"selectedClips": clips, "playheadMs": 1234}}
			response := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/ai-cut/projects/cut-1/agent-runs", input)
			want := 202
			count := 1
			if kind == "invalid" {
				want = 400
				count = 0
			}
			if kind == "write-failure" {
				want = 503
				count = 0
			}
			if response.Code != want || started != count {
				t.Fatalf("status=%d started=%d body=%s", response.Code, started, response.Body.String())
			}
			if kind == "valid" {
				input["uiContext"] = map[string]any{"selectedClips": []any{}, "playheadMs": 999}
				replay := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/ai-cut/projects/cut-1/agent-runs", input)
				if replay.Code != 202 || started != 1 {
					t.Fatalf("bad replay: %d %d", replay.Code, started)
				}
				stored, _ := base.CreateAgentRun(t.Context(), aicut.AgentRun{ID: "probe", ClientRequestID: "ui-1", Owner: "alice", ProjectID: "cut-1", Instruction: "probe", Mode: aicut.AgentModeAgent})
				events, _ := base.AgentEvents(t.Context(), "alice", "cut-1", stored.ID, 0)
				if len(events) != 1 || !containsText(events[0].Message, `"playheadMs":1234`) {
					t.Fatalf("replay rewrote context: %#v", events)
				}
			}
			if kind == "write-failure" {
				stored, _ := base.CreateAgentRun(t.Context(), aicut.AgentRun{ID: "probe", ClientRequestID: "ui-1", Owner: "alice", ProjectID: "cut-1", Instruction: "probe", Mode: aicut.AgentModeAgent})
				if stored.State != aicut.AgentFailed {
					t.Fatalf("run not failed: %#v", stored)
				}
			}
		})
	}
}

func TestAICutAgentUIContextUsesOnlyRequestedTimeline(t *testing.T) {
	_, base, collection, now := collectionRouteFixture(t)
	named, _ := collection.Create(t.Context(), "alice", "p", "A")
	_, err := collection.Save(t.Context(), "alice", "p", named.ID, aicut.Timeline{Tracks: []aicut.Track{{ID: "t", Kind: aicut.TrackSubtitle, Clips: []aicut.Clip{{ID: "named-clip", Kind: aicut.ClipSubtitle, DurationMS: 1000, Text: "named"}}}}}, 0, now)
	if err != nil {
		t.Fatal(err)
	}
	agents := aicut.NewMemoryAgentStore()
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, AICut: base, AICutTimelines: collection, AICutAgentStore: agents})
	input := map[string]any{"instruction": "剪这段", "expectedRevision": 1, "clientRequestId": "selection", "uiContext": map[string]any{"selectedClips": []any{map[string]any{"trackId": "t", "clipId": "named-clip"}}, "playheadMs": 200}}
	response := signedJSONRequestWithTimeline(t, api, now, "POST", "/api/ai-cut/projects/p/agent-runs", named.ID, input)
	if response.Code != 202 {
		t.Fatalf("named create: %d %s", response.Code, response.Body)
	}
	input["expectedRevision"] = 0
	input["clientRequestId"] = "other"
	response = signedJSONRequestWithTimeline(t, api, now, "POST", "/api/ai-cut/projects/p/agent-runs", "default", input)
	if response.Code != 400 {
		t.Fatalf("cross timeline accepted: %d %s", response.Code, response.Body)
	}
	input["expectedRevision"] = 0
	input["clientRequestId"] = "stale"
	response = signedJSONRequestWithTimeline(t, api, now, "POST", "/api/ai-cut/projects/p/agent-runs", named.ID, input)
	if response.Code != 409 {
		t.Fatalf("stale selection accepted: %d %s", response.Code, response.Body)
	}
}
