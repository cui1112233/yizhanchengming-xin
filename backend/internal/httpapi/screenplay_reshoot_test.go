package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/aicut"
	"qiantie/backend/internal/ffmpegbin"
	"qiantie/backend/internal/screenplay"
)

type fakeReshootReferences struct{ calls int }

func (f *fakeReshootReferences) ReferenceURL(_ context.Context, asset agentworkspace.Asset, _ time.Duration) (string, error) {
	f.calls++
	return "https://assets.test/" + asset.ObjectKey, nil
}

type reshootVideoService struct {
	inputs []scriptVideoLocalInput
	poll   scriptVideoStatus
}

func (f *reshootVideoService) Submit(_ context.Context, _ string, in scriptVideoLocalInput) (scriptVideoSubmission, error) {
	f.inputs = append(f.inputs, in)
	return scriptVideoSubmission{TaskID: "reshoot-task-1"}, nil
}

func (f *reshootVideoService) Poll(context.Context, string, string) (scriptVideoStatus, error) {
	return f.poll, nil
}

func reshootWorkbench(t *testing.T, video screenplayVideoService) (*ScreenplayMediaWorkbench, screenplay.Project) {
	t.Helper()
	workbench, p := longVideoTestWorkbench(t, video)
	workbench.ReferencesProvider = &fakeReshootReferences{}
	return workbench, p
}

// seedShotVideo 生成一条真实可解析的测试视频并挂到分镜上。
func seedShotVideo(t *testing.T, workbench *ScreenplayMediaWorkbench, p screenplay.Project, shotID string, seconds float64) string {
	t.Helper()
	if _, err := ffmpegbin.ResolveFFmpeg(); err != nil {
		t.Skipf("本机没有 ffmpeg：%v", err)
	}
	directory := t.TempDir()
	path := filepath.Join(directory, "shot.mp4")
	runner := &ffmpegRunnerForTest{}
	if err := runner.build(t, path, seconds); err != nil {
		t.Fatalf("生成测试视频失败：%v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return seedVideoBytes(t, workbench, p, shotID, raw, "video/mp4")
}

type ffmpegRunnerForTest struct{}

func (r *ffmpegRunnerForTest) build(t *testing.T, path string, seconds float64) error {
	t.Helper()
	ctx := context.Background()
	command, _, err := ffmpegbin.Command(ctx, "", "-hide_banner", "-loglevel", "error", "-y",
		"-f", "lavfi", "-i", fmt.Sprintf("color=c=navy:s=320x568:d=%.1f:r=24", seconds),
		"-f", "lavfi", "-i", fmt.Sprintf("sine=frequency=330:duration=%.1f", seconds),
		"-shortest", "-c:v", "libx264", "-pix_fmt", "yuv420p", "-c:a", "aac", path)
	if err != nil {
		return err
	}
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", string(output))
	}
	return nil
}

func seedVideoBytes(t *testing.T, workbench *ScreenplayMediaWorkbench, p screenplay.Project, shotID string, raw []byte, contentType string) string {
	t.Helper()
	ctx := context.Background()
	if err := workbench.ensureCut(ctx, p); err != nil {
		t.Fatal(err)
	}
	assetID := newAgentID("asset")
	key := "aicut-assets/" + p.ID + "/" + assetID + ".mp4"
	asset := agentworkspace.Asset{ID: assetID, ProjectID: p.ID, Kind: "video", Name: "镜头视频", ContentType: contentType, ObjectKey: key, ByteSize: int64(len(raw)), CreatedAt: time.Now().UTC()}
	objects := workbench.Objects.(*memoryImageToolObjects)
	if err := objects.Put(ctx, asset, bytes.NewReader(raw)); err != nil {
		t.Fatal(err)
	}
	cutAsset := aicut.Asset{ID: assetID, ProjectID: p.ID, Kind: "video", Name: "镜头视频", ContentType: contentType, ObjectKey: key, ByteSize: int64(len(raw)), CreatedAt: time.Now().UTC()}
	if err := workbench.Cut.CreateAsset(ctx, p.OwnerUsername, cutAsset); err != nil {
		t.Fatal(err)
	}
	item := screenplay.MediaItem{ID: "item-" + assetID, AssetID: assetID, Name: "镜头视频", Kind: "video", Status: "succeeded", TargetID: shotID, ShotID: shotID, CreatedAt: time.Now().UTC()}
	if _, err := workbench.Media.MutateMedia(ctx, p.OwnerUsername, p.ID, func(state *screenplay.MediaState) error {
		state.Items = append(state.Items, item)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	return item.ID
}

func TestReshootCandidatesOnlyListsShotsWithVideo(t *testing.T) {
	ctx := context.Background()
	workbench, p := reshootWorkbench(t, nil)
	seedVideoBytes(t, workbench, p, "shot-2", []byte("fake"), "video/mp4")
	view, err := workbench.reshootCandidates(ctx, BridgeAuth{}, p, 1, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(view.Candidates) != 1 || view.Candidates[0].ShotID != "shot-2" || view.Candidates[0].Seq != 2 {
		t.Fatalf("候选应为第 2 镜：%+v", view.Candidates)
	}
	if view.Capabilities["generate"] == nil || view.Capabilities["apply"] == nil {
		t.Fatalf("应返回生成/拼接能力：%+v", view.Capabilities)
	}
	empty, err := workbench.reshootCandidates(ctx, BridgeAuth{}, p, 2, "")
	if err != nil || len(empty.Candidates) != 0 {
		t.Fatalf("第 2 集没有视频应为空：%+v %v", empty, err)
	}
}

func TestReshootFramesExtractsTwentyFrames(t *testing.T) {
	ctx := context.Background()
	workbench, p := reshootWorkbench(t, nil)
	seedShotVideo(t, workbench, p, "shot-1", 6)
	view, err := workbench.reshootFrames(ctx, p, reshootRequest{ShotID: "shot-1"})
	if err != nil {
		t.Fatal(err)
	}
	if view.Duration < 5.4 || view.Duration > 6.6 {
		t.Fatalf("视频时长异常：%.2f", view.Duration)
	}
	if len(view.Frames) != screenplay.ReshootFrameCount {
		t.Fatalf("时间轴帧数应为 20，实际 %d", len(view.Frames))
	}
	if len(view.FirstFrame) < 100 || len(view.LastFrame) < 100 {
		t.Fatal("首尾帧应为 base64 图片")
	}
	if view.Frames[0][:23] != "data:image/jpeg;base64," {
		t.Fatalf("帧格式异常：%s", view.Frames[0][:30])
	}
}

func TestReshootGenerateUsesFramePromptAndSavedModel(t *testing.T) {
	ctx := context.Background()
	service := &reshootVideoService{poll: scriptVideoStatus{Status: "failed", Error: "供应商离线"}}
	workbench, p := reshootWorkbench(t, service)
	seedShotVideo(t, workbench, p, "shot-1", 6)
	view, err := workbench.startReshoot(ctx, p, reshootRequest{ShotID: "shot-1", Start: 1, End: 3.4, Duration: 0, Prompt: "把外套换成红色"})
	if err != nil {
		t.Fatal(err)
	}
	if view.Status != screenplay.ReshootStatusGenerating || view.ProviderID != "reshoot-task-1" {
		t.Fatalf("任务应处于生成中：%+v", view)
	}
	if len(service.inputs) != 1 {
		t.Fatalf("应提交一次任务，实际 %d", len(service.inputs))
	}
	input := service.inputs[0]
	if input.ModelKey != "seedance-2-0-official" || input.GenerationMode != "image_to_video" {
		t.Fatalf("模型或模式不对：%+v", input)
	}
	if len(input.ImageURLs) != 2 {
		t.Fatalf("首尾帧应作为两张参考图：%+v", input.ImageURLs)
	}
	if input.Prompt != screenplay.ReshootFramePromptPrefix+"，把外套换成红色" {
		t.Fatalf("提示词应带首尾帧说明：%q", input.Prompt)
	}
	if fmt.Sprint(input.Duration) != "2" {
		t.Fatalf("生成时长应按框选四舍五入：%v", input.Duration)
	}
	status, err := workbench.reshootStatus(ctx, BridgeAuth{}, p, "shot-1")
	if err != nil {
		t.Fatal(err)
	}
	if status.Status != screenplay.ReshootStatusFailed || status.Error != "供应商离线" {
		t.Fatalf("轮询失败应落到失败状态：%+v", status)
	}
}

func TestReshootApplySplicesAndSavesShotVideo(t *testing.T) {
	ctx := context.Background()
	workbench, p := reshootWorkbench(t, nil)
	seedShotVideo(t, workbench, p, "shot-1", 6)
	// 新片段 2 秒，预先标成已完成，模拟供应商结果已导入。
	insertDirectory := t.TempDir()
	insertPath := filepath.Join(insertDirectory, "insert.mp4")
	runner := &ffmpegRunnerForTest{}
	if err := runner.build(t, insertPath, 2); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(insertPath)
	if err != nil {
		t.Fatal(err)
	}
	segmentID := seedVideoBytes(t, workbench, p, "", raw, "video/mp4")
	task := screenplay.ReshootTask{ShotID: "shot-1", Seq: 1, Episode: 1, Start: 2, End: 4, Duration: 2, Mode: screenplay.ReshootModeFrames, Status: screenplay.ReshootStatusDone, Progress: 100, ProviderID: "task-1", ItemID: segmentID}
	if _, err = workbench.Media.MutateMedia(ctx, p.OwnerUsername, p.ID, func(state *screenplay.MediaState) error {
		state.PutReshoot(task)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	view, err := workbench.applyReshoot(ctx, BridgeAuth{}, p, "shot-1")
	if err != nil {
		t.Fatal(err)
	}
	if view.Status != screenplay.ReshootStatusApplied || view.AppliedID == "" || view.AppliedURL == "" {
		t.Fatalf("应生成并登记拼接成片：%+v", view)
	}
	state, err := workbench.Media.LoadMedia(ctx, p.OwnerUsername, p.ID)
	if err != nil {
		t.Fatal(err)
	}
	var saved screenplay.MediaItem
	for _, item := range state.Items {
		if item.ID == view.AppliedID {
			saved = item
			break
		}
	}
	if saved.Kind != "video" || saved.ShotID != "shot-1" || saved.DurationMS == 0 {
		t.Fatalf("成片应作为这条分镜的视频素材保存：%+v", saved)
	}
	if saved.DurationMS < 5000 || saved.DurationMS > 7000 {
		t.Fatalf("拼接后时长异常：%dms", saved.DurationMS)
	}
	payload, _ := json.Marshal(view)
	if len(payload) == 0 {
		t.Fatal("状态应可序列化")
	}
}
