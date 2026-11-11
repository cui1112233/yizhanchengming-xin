package aicut

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"
)

type fixedASRProvider struct {
	cues    []SRTCue
	called  chan struct{}
	release chan struct{}
}

func (p *fixedASRProvider) Transcribe(ctx context.Context, _ ASRInput) ([]SRTCue, error) {
	if p.called != nil {
		close(p.called)
	}
	if p.release != nil {
		select {
		case <-p.release:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return p.cues, nil
}

func createASRFixture(t *testing.T) (*MemoryStore, time.Time) {
	t.Helper()
	store := NewMemoryStore()
	now := time.Date(2026, 11, 9, 11, 0, 0, 0, time.UTC)
	if err := store.CreateProject(context.Background(), "alice", Project{ID: "cut-1", Title: "测试", AspectRatio: "9:16", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	if err := store.CreateAsset(context.Background(), "alice", Asset{ID: "asset-1", ProjectID: "cut-1", Kind: "audio", Name: "voice.mp3", ContentType: "audio/mpeg", ObjectKey: "voice.mp3", ByteSize: 3, CreatedAt: now}); err != nil {
		t.Fatal(err)
	}
	return store, now
}

func TestASRServiceProcessesQueuedJobIntoSubtitleTrack(t *testing.T) {
	store, now := createASRFixture(t)
	provider := &fixedASRProvider{cues: []SRTCue{{StartMS: 0, EndMS: 1200, Text: "第一句"}, {StartMS: 1200, EndMS: 2400, Text: "第二句"}}}
	service := NewASRService(ASRServiceOptions{Store: store, OpenObject: func(context.Context, Asset) (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewBufferString("mp3")), nil
	}, ResolveProvider: func(context.Context, string, string) (ASRProvider, error) { return provider, nil }, Now: func() time.Time { return now.Add(time.Minute) }})
	job, err := service.Submit(context.Background(), "alice", ASRSubmit{ProjectID: "cut-1", AssetID: "asset-1", ModelID: "asr-main", Language: "zh", ExpectedRevision: 0})
	if err != nil {
		t.Fatal(err)
	}
	if err := service.ProcessNext(context.Background()); err != nil {
		t.Fatal(err)
	}
	completed, err := store.ASRJob(context.Background(), "alice", "cut-1", job.ID)
	if err != nil || completed.State != ASRSucceeded || completed.Imported != 2 || completed.TrackID == "" {
		t.Fatalf("job=%#v err=%v", completed, err)
	}
	timeline, _ := store.Timeline(context.Background(), "alice", "cut-1")
	if timeline.Revision != 1 || len(timeline.Tracks) != 1 || len(timeline.Tracks[0].Clips) != 2 {
		t.Fatalf("timeline=%#v", timeline)
	}
}

func TestASRServiceCancellationDoesNotWriteTimeline(t *testing.T) {
	store, now := createASRFixture(t)
	provider := &fixedASRProvider{cues: []SRTCue{{StartMS: 0, EndMS: 1000, Text: "不应写入"}}, called: make(chan struct{}), release: make(chan struct{})}
	service := NewASRService(ASRServiceOptions{Store: store, OpenObject: func(context.Context, Asset) (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewBufferString("mp3")), nil
	}, ResolveProvider: func(context.Context, string, string) (ASRProvider, error) { return provider, nil }, Now: func() time.Time { return now.Add(time.Minute) }})
	job, err := service.Submit(context.Background(), "alice", ASRSubmit{ProjectID: "cut-1", AssetID: "asset-1", ModelID: "asr-main", ExpectedRevision: 0})
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- service.ProcessNext(context.Background()) }()
	<-provider.called
	if _, err := service.Cancel(context.Background(), "alice", "cut-1", job.ID); err != nil {
		t.Fatal(err)
	}
	close(provider.release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	timeline, _ := store.Timeline(context.Background(), "alice", "cut-1")
	if timeline.Revision != 0 || len(timeline.Tracks) != 0 {
		t.Fatalf("canceled job changed timeline: %#v", timeline)
	}
	canceled, _ := store.ASRJob(context.Background(), "alice", "cut-1", job.ID)
	if canceled.State != ASRCanceled {
		t.Fatalf("state=%s", canceled.State)
	}
}

func TestASRServiceRejectsNonMediaAndEmptyModel(t *testing.T) {
	store, now := createASRFixture(t)
	if err := store.CreateAsset(context.Background(), "alice", Asset{ID: "subtitle-1", ProjectID: "cut-1", Kind: "subtitle", Name: "a.srt", ContentType: "application/x-subrip", ObjectKey: "a.srt", ByteSize: 2, CreatedAt: now}); err != nil {
		t.Fatal(err)
	}
	service := NewASRService(ASRServiceOptions{Store: store})
	if _, err := service.Submit(context.Background(), "alice", ASRSubmit{ProjectID: "cut-1", AssetID: "asset-1"}); err == nil {
		t.Fatal("empty model must fail")
	}
	if _, err := service.Submit(context.Background(), "alice", ASRSubmit{ProjectID: "cut-1", AssetID: "subtitle-1", ModelID: "asr-main"}); err == nil {
		t.Fatal("subtitle asset must fail")
	}
}

func TestASRServiceRejectsDuplicateActiveJobForSameAsset(t *testing.T) {
	store, now := createASRFixture(t)
	service := NewASRService(ASRServiceOptions{Store: store, Now: func() time.Time { return now }})
	first, err := service.Submit(context.Background(), "alice", ASRSubmit{ProjectID: "cut-1", AssetID: "asset-1", ModelID: "asr-main"})
	if err != nil {
		t.Fatalf("first Submit() error = %v", err)
	}
	if first.State != ASRQueued {
		t.Fatalf("first state = %s", first.State)
	}
	if _, err = service.Submit(context.Background(), "alice", ASRSubmit{ProjectID: "cut-1", AssetID: "asset-1", ModelID: "asr-main"}); err == nil || !strings.Contains(err.Error(), "正在识别") {
		t.Fatalf("duplicate Submit() error = %v, want active duplicate error", err)
	}
}
