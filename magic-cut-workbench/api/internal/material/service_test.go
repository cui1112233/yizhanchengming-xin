package material

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	worktask "yizhanchengming/magic-cut-workbench/internal/task"
)

type importRecorder struct {
	material Material
	task     worktask.Task
}

func (r *importRecorder) SaveImport(_ context.Context, material Material, task worktask.Task) error {
	r.material = material
	r.task = task
	return nil
}

type probeFunc func(context.Context, string) (Metadata, error)

func (fn probeFunc) Probe(ctx context.Context, path string) (Metadata, error) {
	return fn(ctx, path)
}

func TestClassifyMediaRejectsUnsupportedFilesWithStableCode(t *testing.T) {
	_, err := ClassifyMedia("setup.exe")
	if ErrorCode(err) != "FORMAT_NOT_SUPPORTED" {
		t.Fatalf("code = %q; err=%v", ErrorCode(err), err)
	}
	for name, want := range map[string]string{
		"clip.MP4":  "video",
		"voice.wav": "audio",
		"cover.png": "image",
	} {
		got, classifyErr := ClassifyMedia(name)
		if classifyErr != nil || got != want {
			t.Fatalf("ClassifyMedia(%q) = %q, %v; want %q", name, got, classifyErr, want)
		}
	}
}

func TestParseProbeJSONExtractsDurationDimensionsAndCodecs(t *testing.T) {
	metadata, err := ParseProbeJSON([]byte(`{
		"format":{"duration":"12.345"},
		"streams":[
			{"codec_type":"video","codec_name":"h264","width":1920,"height":1080,"r_frame_rate":"25/1"},
			{"codec_type":"audio","codec_name":"aac"}
		]
	}`))
	if err != nil {
		t.Fatal(err)
	}
	if metadata.DurationMs != 12345 || metadata.Width != 1920 || metadata.Height != 1080 || metadata.FrameRate != 25 || metadata.VideoCodec != "h264" || metadata.AudioCodec != "aac" {
		t.Fatalf("metadata = %+v", metadata)
	}
}

func TestImportUsesGeneratedSafePathAndPersistsPendingProbeTask(t *testing.T) {
	root := t.TempDir()
	recorder := &importRecorder{}
	service := NewService(ServiceOptions{
		StorageRoot: root,
		MaxBytes:    1024,
		Probe: probeFunc(func(_ context.Context, path string) (Metadata, error) {
			if strings.Contains(path, "original name") || strings.Contains(path, "..") {
				t.Fatalf("unsafe storage path: %s", path)
			}
			return Metadata{DurationMs: 2500, Width: 1280, Height: 720, VideoCodec: "h264"}, nil
		}),
		Repository:    recorder,
		NewMaterialID: func() string { return "mcm_safe123" },
		NewTaskID:     func() string { return "mct_probe123" },
		Now:           func() time.Time { return time.Date(2026, 11, 20, 8, 0, 0, 0, time.UTC) },
	})

	temp, err := os.CreateTemp(t.TempDir(), "upload-*.mp4")
	if err != nil {
		t.Fatal(err)
	}
	defer temp.Close()
	if _, err = io.Copy(temp, bytes.NewBufferString("video fixture")); err != nil {
		t.Fatal(err)
	}
	if _, err = temp.Seek(0, io.SeekStart); err != nil {
		t.Fatal(err)
	}
	header := &multipart.FileHeader{Filename: "../original name.mp4", Size: int64(len("video fixture"))}

	got, queued, err := service.Import(context.Background(), "alice@example.com", "mcp_project", temp, header)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "mcm_safe123" || queued.ID != "mct_probe123" || got.Status != "ready" || queued.Status != worktask.StatusSucceeded {
		t.Fatalf("material=%+v task=%+v", got, queued)
	}
	wantPath := filepath.Join(root, "alice_example.com", "mcp_project", "mcm_safe123.mp4")
	if got.StorageKey != wantPath {
		t.Fatalf("storageKey=%q want %q", got.StorageKey, wantPath)
	}
	if _, err = os.Stat(wantPath); err != nil {
		t.Fatal(err)
	}
	if recorder.material.ID != got.ID || recorder.task.ID != queued.ID {
		t.Fatalf("persisted material=%+v task=%+v", recorder.material, recorder.task)
	}
}

func TestImportRejectsOversizedFileBeforeWriting(t *testing.T) {
	service := NewService(ServiceOptions{StorageRoot: t.TempDir(), MaxBytes: 3})
	header := &multipart.FileHeader{Filename: "clip.mp4", Size: 4}
	temp, createErr := os.CreateTemp(t.TempDir(), "oversized-*.mp4")
	if createErr != nil {
		t.Fatal(createErr)
	}
	defer temp.Close()
	if _, createErr = temp.WriteString("1234"); createErr != nil {
		t.Fatal(createErr)
	}
	if _, createErr = temp.Seek(0, io.SeekStart); createErr != nil {
		t.Fatal(createErr)
	}
	_, _, err := service.Import(context.Background(), "alice", "mcp_project", temp, header)
	if ErrorCode(err) != "FILE_TOO_LARGE" {
		t.Fatalf("code=%q err=%v", ErrorCode(err), err)
	}
}
