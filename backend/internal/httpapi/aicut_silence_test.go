package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/aicut"
	"qiantie/backend/internal/ffmpegbin"
	"testing"
	"time"
)

func TestRunAICutAnalyzeSilenceRealFFmpegAndOwner(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	file := filepath.Join(dir, "tone.wav")
	err := ffmpegbin.Run(ctx, "ffmpeg", "-hide_banner", "-loglevel", "error", "-y", "-f", "lavfi", "-i", `aevalsrc=if(between(t\,1\,2)\,0\,0.4*sin(2*PI*440*t))|-if(between(t\,1\,2)\,0\,0.4*sin(2*PI*440*t)):s=16000:d=3`, "-c:a", "pcm_s16le", file)
	if ffmpegbin.IsMissing(err) {
		t.Skip(err)
	}
	if err != nil {
		t.Fatal(err)
	}
	store := aicut.NewMemoryStore()
	now := time.Now()
	if err := store.CreateProject(ctx, "alice", aicut.Project{ID: "p", Title: "test", AspectRatio: "16:9", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	objects := agentworkspace.NewLocalAssetObjectStore(filepath.Join(dir, "objects"))
	asset := aicut.Asset{ID: "a", ProjectID: "p", Name: "tone.wav", Kind: "audio", ContentType: "audio/wav", ObjectKey: "test/tone.wav", DurationMS: 3000, CreatedAt: now}
	data, _ := os.ReadFile(file)
	if err := objects.Put(ctx, aicutObjectAsset(asset), bytes.NewReader(data)); err != nil {
		t.Fatal(err)
	}
	if err := store.CreateAsset(ctx, "alice", asset); err != nil {
		t.Fatal(err)
	}
	got, err := RunAICutAnalyzeSilence(ctx, store, objects, "alice", "p", "a", aicut.SilenceOptions{PaddingMS: 100})
	if err != nil || got.AllSilent || len(got.Intervals) != 1 || got.Intervals[0].StartMS < 1080 || got.Intervals[0].StartMS > 1120 || got.Intervals[0].EndMS < 1880 || got.Intervals[0].EndMS > 1920 {
		t.Fatalf("%+v %v", got, err)
	}
	if _, err := RunAICutAnalyzeSilence(ctx, store, objects, "bob", "p", "a", aicut.SilenceOptions{}); err == nil {
		t.Fatal("cross-owner asset accepted")
	}
	timeline, err := store.SaveTimeline(ctx, "alice", "p", aicut.Timeline{Script: "保留剧本", Markers: []aicut.TimelineMarker{{ID: "m", Label: "后半段", TimeMS: 2500}}, Tracks: []aicut.Track{{ID: "audio", Kind: aicut.TrackAudio, Clips: []aicut.Clip{{ID: "clip", AssetID: "a", Kind: aicut.ClipAudio, DurationMS: 3000, SourceDurationMS: 3000}}}}}, 0, now)
	if err != nil {
		t.Fatal(err)
	}
	agents := aicut.NewMemoryAgentStore()
	run := aicut.AgentRun{ID: "run", ClientRequestID: "req", Owner: "alice", ProjectID: "p", Instruction: "剪掉真实静音", Mode: aicut.AgentModeAgent, ExpectedRevision: timeline.Revision, State: aicut.AgentQueued, CreatedAt: now, UpdatedAt: now}
	if _, err := agents.CreateAgentRun(ctx, run); err != nil {
		t.Fatal(err)
	}
	executor := aicut.NewAgentToolExecutor(store, agents, func() time.Time { return now })
	executor.ConfigureCompositeTools(aicut.AgentCompositeTools{AnalyzeSilence: func(ctx context.Context, owner, project, asset string, options aicut.SilenceOptions) (aicut.SilenceAnalysis, error) {
		return RunAICutAnalyzeSilence(ctx, store, objects, owner, project, asset, options)
	}})
	call := aicut.ToolCall{ID: "trim", Name: "remove_audio_silence", Arguments: json.RawMessage(`{"assetId":"a","paddingMs":100}`)}
	result, err := executor.Execute(ctx, run, call)
	if err != nil || !result.OK {
		t.Fatalf("actual Agent trim: %+v %v", result, err)
	}
	edited, err := store.Timeline(ctx, "alice", "p")
	if err != nil || edited.Script != "保留剧本" || len(edited.Tracks[0].Clips) != 2 || edited.Tracks[0].Clips[1].SourceStartMS != 1900 || edited.Markers[0].TimeMS != 1700 {
		t.Fatalf("actual saved edit: %+v %v", edited, err)
	}
	replay, err := executor.Execute(ctx, run, call)
	if err != nil || !replay.OK || replay.Revision != result.Revision {
		t.Fatalf("actual receipt replay: %+v %v", replay, err)
	}
}
func TestRunAICutAnalyzeSilenceRejectsNoAudioAndCancel(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	file := filepath.Join(dir, "silent-video.mp4")
	err := ffmpegbin.Run(ctx, "ffmpeg", "-hide_banner", "-loglevel", "error", "-y", "-f", "lavfi", "-i", "color=c=black:s=16x16:d=1", "-an", file)
	if ffmpegbin.IsMissing(err) {
		t.Skip(err)
	}
	if err != nil {
		t.Fatal(err)
	}
	store := aicut.NewMemoryStore()
	now := time.Now()
	if err := store.CreateProject(ctx, "alice", aicut.Project{ID: "p", Title: "test", AspectRatio: "16:9", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	objects := agentworkspace.NewLocalAssetObjectStore(filepath.Join(dir, "objects"))
	asset := aicut.Asset{ID: "v", ProjectID: "p", Name: "silent-video.mp4", Kind: "video", ContentType: "video/mp4", ObjectKey: "test/video.mp4", DurationMS: 1000, CreatedAt: now}
	data, _ := os.ReadFile(file)
	if err := objects.Put(ctx, aicutObjectAsset(asset), bytes.NewReader(data)); err != nil {
		t.Fatal(err)
	}
	if err := store.CreateAsset(ctx, "alice", asset); err != nil {
		t.Fatal(err)
	}
	if _, err := RunAICutAnalyzeSilence(ctx, store, objects, "alice", "p", "v", aicut.SilenceOptions{}); err == nil {
		t.Fatal("video without audio accepted")
	}
	ctx, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := RunAICutAnalyzeSilence(ctx, store, objects, "alice", "p", "v", aicut.SilenceOptions{}); err != context.Canceled {
		t.Fatalf("cancellation: %v", err)
	}
}
func TestRunAICutAnalyzeSilenceRejectsShiftedAudioStream(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	file := filepath.Join(dir, "offset.mkv")
	err := ffmpegbin.Run(ctx, "ffmpeg", "-hide_banner", "-loglevel", "error", "-y", "-f", "lavfi", "-i", "color=c=black:s=16x16:r=25:d=4", "-itsoffset", "1", "-f", "lavfi", "-i", `aevalsrc=if(between(t\,1\,2)\,0\,0.4*sin(2*PI*440*t)):s=16000:d=3`, "-map", "0:v:0", "-map", "1:a:0", "-c:v", "ffv1", "-c:a", "pcm_s16le", file)
	if ffmpegbin.IsMissing(err) {
		t.Skip(err)
	}
	if err != nil {
		t.Fatal(err)
	}
	store := aicut.NewMemoryStore()
	now := time.Now()
	if err := store.CreateProject(ctx, "alice", aicut.Project{ID: "p", Title: "test", AspectRatio: "16:9", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	objects := agentworkspace.NewLocalAssetObjectStore(filepath.Join(dir, "objects"))
	asset := aicut.Asset{ID: "v", ProjectID: "p", Name: "offset.mkv", Kind: "video", ContentType: "video/x-matroska", ObjectKey: "test/offset.mkv", DurationMS: 4000, CreatedAt: now}
	data, _ := os.ReadFile(file)
	if err := objects.Put(ctx, aicutObjectAsset(asset), bytes.NewReader(data)); err != nil {
		t.Fatal(err)
	}
	if err := store.CreateAsset(ctx, "alice", asset); err != nil {
		t.Fatal(err)
	}
	if got, err := RunAICutAnalyzeSilence(ctx, store, objects, "alice", "p", "v", aicut.SilenceOptions{}); err == nil {
		t.Fatalf("offset audio accepted with wrong source times: %+v", got)
	}
}
