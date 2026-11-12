package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"qiantie/backend/internal/aicut"
	"strings"
	"testing"
	"time"
)

type scopedASRProvider func(context.Context, aicut.ASRInput) ([]aicut.SRTCue, error)

func (f scopedASRProvider) Transcribe(c context.Context, in aicut.ASRInput) ([]aicut.SRTCue, error) {
	return f(c, in)
}
func TestAICutNamedASRHTTPImportsAndPollsSelectedTimeline(t *testing.T) {
	_, base, collection, now := collectionRouteFixture(t)
	a, _ := collection.Create(t.Context(), "alice", "p", "A")
	_ = base.CreateAsset(t.Context(), "alice", aicut.Asset{ID: "audio", ProjectID: "p", Kind: "audio", Name: "voice.mp3", ContentType: "audio/mpeg", ObjectKey: "audio", ByteSize: 3, CreatedAt: now})
	service := aicut.NewASRService(aicut.ASRServiceOptions{Store: base, Timelines: collection, Now: func() time.Time { return now }, OpenObject: func(context.Context, aicut.Asset) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader("audio")), nil
	}, ResolveProvider: func(context.Context, string, string) (aicut.ASRProvider, error) {
		return scopedASRProvider(func(context.Context, aicut.ASRInput) ([]aicut.SRTCue, error) {
			return []aicut.SRTCue{{StartMS: 0, EndMS: 1000, Text: "Named ASR"}}, nil
		}), nil
	}})
	router := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, AICut: base, AICutTimelines: collection, AICutASR: service})
	root := "/api/ai-cut/projects/p/asr-jobs"
	r := signedJSONRequestWithTimeline(t, router, now, "POST", root, a.ID, map[string]any{"assetId": "audio", "modelId": "test", "expectedRevision": 0})
	if r.Code != 202 {
		t.Fatalf("submit=%d %s", r.Code, r.Body)
	}
	var p struct {
		Job aicut.ASRJob `json:"job"`
	}
	_ = json.Unmarshal(r.Body.Bytes(), &p)
	if p.Job.TimelineID != a.ID {
		t.Fatalf("job=%+v", p.Job)
	}
	r = signedJSONRequestWithTimeline(t, router, now, "POST", root+"/"+p.Job.ID+"/cancel", "default", map[string]any{})
	if r.Code != 404 {
		t.Fatalf("default cancel=%d", r.Code)
	}
	if e := service.ProcessNext(t.Context()); e != nil {
		t.Fatal(e)
	}
	r = signedJSONRequestWithTimeline(t, router, now, "GET", root, a.ID, nil)
	if r.Code != 200 || !containsText(r.Body.String(), `"state":"succeeded"`) {
		t.Fatalf("named poll=%d %s", r.Code, r.Body)
	}
	r = signedJSONRequestWithTimeline(t, router, now, "GET", root, "default", nil)
	if r.Code != 200 || containsText(r.Body.String(), p.Job.ID) {
		t.Fatalf("default poll=%d %s", r.Code, r.Body)
	}
	v, _ := collection.Timeline(t.Context(), "alice", "p", a.ID)
	d, _ := base.Timeline(t.Context(), "alice", "p")
	if len(v.Tracks) != 1 || v.Tracks[0].Clips[0].Text != "Named ASR" || d.Revision != 0 {
		t.Fatalf("named=%+v default=%+v", v, d)
	}
}
