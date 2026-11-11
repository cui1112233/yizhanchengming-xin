package httpapi

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/aicut"
	"qiantie/backend/internal/ffmpegbin"
)

func consolidationTimeline(track aicut.Track) aicut.Timeline {
	return aicut.Timeline{Revision: 7, Tracks: []aicut.Track{track}}
}

func TestRunAICutConsolidateRendersPersistentProjectAsset(t *testing.T) {
	now := time.Date(2026, 11, 11, 8, 30, 0, 0, time.UTC)
	temp := t.TempDir()
	sourcePath := filepath.Join(temp, "source.mp4")
	if err := ffmpegbin.Run(t.Context(), "ffmpeg", "-hide_banner", "-loglevel", "error", "-f", "lavfi", "-i", "testsrc2=s=320x180:r=30:d=0.6", "-c:v", "libx264", "-pix_fmt", "yuv420p", sourcePath); err != nil {
		t.Skipf("ffmpeg unavailable: %v", err)
	}
	store := aicut.NewMemoryStore()
	project := aicut.Project{ID: "cut-consolidate", Title: "片段合并", AspectRatio: "16:9", CreatedAt: now, UpdatedAt: now}
	if err := store.CreateProject(t.Context(), "alice", project); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	source := aicut.Asset{ID: "source-video", ProjectID: project.ID, Kind: "video", Name: "source.mp4", ContentType: "video/mp4", ObjectKey: "aicut-assets/alice/cut-consolidate/source.mp4", ByteSize: info.Size(), DurationMS: 600, CreatedAt: now}
	if err := store.CreateAsset(t.Context(), "alice", source); err != nil {
		t.Fatal(err)
	}
	objects := agentworkspace.NewLocalAssetObjectStore(filepath.Join(temp, "objects"))
	file, err := os.Open(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	if err := objects.Put(t.Context(), aicutObjectAsset(source), file); err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	_ = file.Close()
	timeline := aicut.Timeline{Revision: 0, Tracks: []aicut.Track{{ID: "video-1", Kind: aicut.TrackVideo, Name: "画面轨道", Clips: []aicut.Clip{
		{ID: "clip-a", AssetID: source.ID, Kind: aicut.ClipVideo, Name: "前半段", StartMS: 0, DurationMS: 300, SourceStartMS: 0, SourceDurationMS: 600, PlaybackRate: 1, Volume: 1},
		{ID: "clip-b", AssetID: source.ID, Kind: aicut.ClipVideo, Name: "后半段", StartMS: 300, DurationMS: 300, SourceStartMS: 300, SourceDurationMS: 600, PlaybackRate: 1, Volume: 1},
	}}}}
	saved, err := store.SaveTimeline(t.Context(), "alice", project.ID, timeline, 0, now)
	if err != nil {
		t.Fatal(err)
	}
	created, err := RunAICutConsolidate(t.Context(), store, objects, func() time.Time { return now }, "alice", project.ID, []string{"clip-a", "clip-b"}, saved.Revision)
	if err != nil {
		t.Fatalf("consolidate: %v", err)
	}
	if created.Kind != "video" || created.ContentType != "video/mp4" || created.DurationMS != 600 || created.ByteSize <= 0 {
		t.Fatalf("created=%+v", created)
	}
	stored, err := store.Asset(t.Context(), "alice", project.ID, created.ID)
	if err != nil || stored.ObjectKey != created.ObjectKey {
		t.Fatalf("stored=%+v err=%v", stored, err)
	}
	body, err := objects.Open(t.Context(), aicutObjectAsset(created))
	if err != nil {
		t.Fatal(err)
	}
	defer body.Close()
	mergedPath := filepath.Join(temp, "merged.mp4")
	merged, err := os.Create(mergedPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.Copy(merged, body); err != nil {
		_ = merged.Close()
		t.Fatal(err)
	}
	_ = merged.Close()
	probe, _, err := ffmpegbin.Command(t.Context(), "ffprobe", "-v", "error", "-select_streams", "v:0", "-show_entries", "stream=codec_name", "-of", "default=noprint_wrappers=1:nokey=1", mergedPath)
	if err != nil {
		t.Fatal(err)
	}
	codec, err := probe.Output()
	if err != nil || strings.TrimSpace(string(codec)) != "h264" {
		t.Fatalf("codec=%q err=%v", codec, err)
	}
}

func TestPrepareAICutConsolidationTimelineRebasesSelectedClips(t *testing.T) {
	timeline := consolidationTimeline(aicut.Track{
		ID: "video-1", Kind: aicut.TrackVideo,
		Clips: []aicut.Clip{
			{ID: "before", Kind: aicut.ClipVideo, StartMS: 0, DurationMS: 500},
			{ID: "clip-b", Kind: aicut.ClipVideo, StartMS: 2000, DurationMS: 1200, SourceStartMS: 300},
			{ID: "clip-a", Kind: aicut.ClipVideo, StartMS: 1000, DurationMS: 1000, SourceStartMS: 50},
		},
	})

	prepared, durationMS, err := prepareAICutConsolidationTimeline(timeline, []string{"clip-b", "clip-a"})
	if err != nil {
		t.Fatalf("prepare consolidation: %v", err)
	}
	if durationMS != 2200 {
		t.Fatalf("duration=%d want 2200", durationMS)
	}
	if prepared.Revision != 7 || len(prepared.Tracks) != 1 || len(prepared.Tracks[0].Clips) != 2 {
		t.Fatalf("prepared=%+v", prepared)
	}
	first, second := prepared.Tracks[0].Clips[0], prepared.Tracks[0].Clips[1]
	if first.ID != "clip-a" || first.StartMS != 0 || first.SourceStartMS != 50 {
		t.Fatalf("first=%+v", first)
	}
	if second.ID != "clip-b" || second.StartMS != 1000 || second.SourceStartMS != 300 {
		t.Fatalf("second=%+v", second)
	}
	if timeline.Tracks[0].Clips[1].StartMS != 2000 || timeline.Tracks[0].Clips[2].StartMS != 1000 {
		t.Fatal("source timeline was modified")
	}
}

func TestPrepareAICutConsolidationTimelineRejectsInvalidSelection(t *testing.T) {
	base := consolidationTimeline(aicut.Track{ID: "video-1", Kind: aicut.TrackVideo, Clips: []aicut.Clip{
		{ID: "a", Kind: aicut.ClipVideo, StartMS: 0, DurationMS: 1000},
		{ID: "b", Kind: aicut.ClipVideo, StartMS: 1000, DurationMS: 1000},
	}})
	tests := []struct {
		name     string
		timeline aicut.Timeline
		ids      []string
		want     string
	}{
		{name: "one clip", timeline: base, ids: []string{"a"}, want: "2 到 50"},
		{name: "duplicate", timeline: base, ids: []string{"a", "a"}, want: "不能重复"},
		{name: "missing", timeline: base, ids: []string{"a", "missing"}, want: "已经不存在"},
		{name: "locked", timeline: consolidationTimeline(aicut.Track{ID: "video-1", Kind: aicut.TrackVideo, Locked: true, Clips: base.Tracks[0].Clips}), ids: []string{"a", "b"}, want: "锁定轨道"},
		{name: "too long", timeline: consolidationTimeline(aicut.Track{ID: "video-1", Kind: aicut.TrackVideo, Clips: []aicut.Clip{{ID: "a", Kind: aicut.ClipVideo, StartMS: 0, DurationMS: 1000}, {ID: "b", Kind: aicut.ClipVideo, StartMS: 30 * 60 * 1000, DurationMS: 1000}}}), ids: []string{"a", "b"}, want: "30 分钟"},
		{name: "wrong kind", timeline: consolidationTimeline(aicut.Track{ID: "audio-1", Kind: aicut.TrackAudio, Clips: []aicut.Clip{{ID: "a", Kind: aicut.ClipAudio, StartMS: 0, DurationMS: 1000}, {ID: "b", Kind: aicut.ClipAudio, StartMS: 1000, DurationMS: 1000}}}), ids: []string{"a", "b"}, want: "只支持同一画面轨道"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, _, err := prepareAICutConsolidationTimeline(test.timeline, test.ids)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("err=%v want contains %q", err, test.want)
			}
		})
	}
}

func TestPrepareAICutConsolidationTimelineAllowsOverlappingComposition(t *testing.T) {
	timeline := consolidationTimeline(aicut.Track{ID: "video-1", Kind: aicut.TrackVideo, Clips: []aicut.Clip{
		{ID: "a", Kind: aicut.ClipVideo, StartMS: 1000, DurationMS: 1200},
		{ID: "b", Kind: aicut.ClipVideo, StartMS: 1800, DurationMS: 1000},
	}})
	prepared, durationMS, err := prepareAICutConsolidationTimeline(timeline, []string{"a", "b"})
	if err != nil {
		t.Fatalf("overlap should render as the current composition: %v", err)
	}
	if durationMS != 1800 || prepared.Tracks[0].Clips[0].StartMS != 0 || prepared.Tracks[0].Clips[1].StartMS != 800 {
		t.Fatalf("duration=%d clips=%+v", durationMS, prepared.Tracks[0].Clips)
	}
}
