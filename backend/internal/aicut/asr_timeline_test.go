package aicut

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"
)

func TestASRNamedTimelinePersistsAcrossWorkerRestartAndRetry(t *testing.T) {
	ctx := context.Background()
	base, now := createASRFixture(t)
	collection := NewMemoryTimelineCollectionStore(base)
	a, _ := collection.Create(ctx, "alice", "cut-1", "A")
	b, _ := collection.Create(ctx, "alice", "cut-1", "B")
	calls := 0
	provider := ASRProviderFuncForTest(func(context.Context, ASRInput) ([]SRTCue, error) {
		calls++
		if calls == 1 {
			return nil, errors.New("temporary")
		}
		return []SRTCue{{StartMS: 0, EndMS: 1200, Text: "Named caption"}}, nil
	})
	options := ASRServiceOptions{Store: base, Timelines: collection, OpenObject: func(context.Context, Asset) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader("audio")), nil
	}, ResolveProvider: func(context.Context, string, string) (ASRProvider, error) { return provider, nil }, Now: func() time.Time { return now }}
	svc := NewASRService(options)
	input := ASRSubmit{ProjectID: "cut-1", TimelineID: a.ID, AssetID: "asset-1", ModelID: "model", ExpectedRevision: 0}
	job, e := svc.Submit(ctx, "alice", input)
	if e != nil {
		t.Fatal(e)
	}
	if e = NewASRService(options).ProcessNext(ctx); e == nil {
		t.Fatal("expected first provider failure")
	}
	failed, _ := base.ASRJob(ctx, "alice", "cut-1", job.ID)
	if failed.State != ASRFailed || failed.TimelineID != a.ID {
		t.Fatalf("failed=%+v", failed)
	}
	job, e = NewASRService(options).Submit(ctx, "alice", input)
	if e != nil {
		t.Fatal(e)
	}
	if e = NewASRService(options).ProcessNext(ctx); e != nil {
		t.Fatal(e)
	}
	done, _ := base.ASRJob(ctx, "alice", "cut-1", job.ID)
	v, _ := collection.Timeline(ctx, "alice", "cut-1", a.ID)
	if done.State != ASRSucceeded || done.TimelineID != a.ID || v.Revision != 1 || len(v.Tracks) != 1 || v.Tracks[0].Clips[0].Text != "Named caption" {
		t.Fatalf("done=%+v timeline=%+v", done, v)
	}
	for _, id := range []string{DefaultTimelineID, b.ID} {
		v, _ := collection.Timeline(ctx, "alice", "cut-1", id)
		if v.Revision != 0 || len(v.Tracks) != 0 {
			t.Fatalf("wrong target %s=%+v", id, v)
		}
		jobs, e := svc.ListForTimeline(ctx, "alice", "cut-1", id)
		if e != nil || len(jobs) != 0 {
			t.Fatalf("jobs scope %s=%+v %v", id, jobs, e)
		}
		if _, e = svc.JobForTimeline(ctx, "alice", "cut-1", job.ID, id); !errors.Is(e, ErrASRJobNotFound) {
			t.Fatalf("foreign scope=%v", e)
		}
		if _, e = svc.CancelForTimeline(ctx, "alice", "cut-1", job.ID, id); !errors.Is(e, ErrASRJobNotFound) {
			t.Fatalf("foreign cancel=%v", e)
		}
	}
}

type ASRProviderFuncForTest func(context.Context, ASRInput) ([]SRTCue, error)

func (f ASRProviderFuncForTest) Transcribe(ctx context.Context, in ASRInput) ([]SRTCue, error) {
	return f(ctx, in)
}
func TestASRNamedDuplicateDetectionIsPerTimeline(t *testing.T) {
	base, now := createASRFixture(t)
	c := NewMemoryTimelineCollectionStore(base)
	a, _ := c.Create(t.Context(), "alice", "cut-1", "A")
	s := NewASRService(ASRServiceOptions{Store: base, Timelines: c, Now: func() time.Time { return now }})
	in := ASRSubmit{ProjectID: "cut-1", AssetID: "asset-1", ModelID: "model"}
	if _, e := s.Submit(t.Context(), "alice", in); e != nil {
		t.Fatal(e)
	}
	in.TimelineID = a.ID
	if _, e := s.Submit(t.Context(), "alice", in); e != nil {
		t.Fatalf("other timeline duplicate=%v", e)
	}
	if _, e := s.Submit(t.Context(), "alice", in); e == nil {
		t.Fatal("same timeline duplicate allowed")
	}
}

func TestASRStoredTimelineTargetCannotBeChangedByJobUpdate(t *testing.T) {
	base, now := createASRFixture(t)
	job := ASRJob{ID: "job", Owner: "alice", ProjectID: "cut-1", TimelineID: "original", AssetID: "asset-1", ModelID: "model", State: ASRQueued, CreatedAt: now, UpdatedAt: now}
	if e := base.CreateASRJob(t.Context(), "alice", job); e != nil {
		t.Fatal(e)
	}
	job.TimelineID = "other"
	if e := base.UpdateASRJob(t.Context(), job); !errors.Is(e, ErrInvalidASRJob) {
		t.Fatalf("retarget=%v", e)
	}
	stored, _ := base.ASRJob(t.Context(), "alice", "cut-1", "job")
	if stored.TimelineID != "original" {
		t.Fatal("persisted target changed")
	}
}
