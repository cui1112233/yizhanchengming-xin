package httpapi

import (
	"encoding/json"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
)

func automationSettings(values map[string]any) batchfactoryv11.SettingsPatch {
	patch := batchfactoryv11.SettingsPatch{}
	for key, item := range values {
		raw, err := json.Marshal(item)
		if err != nil {
			panic(err)
		}
		patch[key] = raw
	}
	return patch
}

func automationTask(id, videoID string, variant int, status batchfactoryv11.ProductionState, media string, at time.Time) batchfactoryv11.ProductionTask {
	return batchfactoryv11.ProductionTask{ID: id, VideoID: videoID, OpeningVariantIndex: variant, Status: status, MediaURL: media, UpdatedAt: at}
}

// 已完成的分镜视频可能分散在多个 production job 里（每轮只补没做完的分镜），
// 自动化必须按 ordinal 汇总判断，不能因为“没有一条 job 全部成功”就整轮重跑。
func TestAutomationSucceededProductionAggregatesCompletedVideosAcrossJobs(t *testing.T) {
	batch := batchfactoryv11.Batch{ID: "batch-1"}
	book := batchfactoryv11.Book{
		ID:               "book-1",
		Videos:           []batchfactoryv11.Video{{ID: "video-0", Ordinal: 0}, {ID: "video-1", Ordinal: 1}, {ID: "video-2", Ordinal: 2}},
		DirectorRevision: &batchfactoryv11.DirectorRevision{ID: "director-1"},
	}
	jobs := []batchfactoryv11.ProductionJob{
		{ID: "job-a", BookID: "book-1", DirectorRevisionID: "director-1", Tasks: []batchfactoryv11.ProductionTask{
			automationTask("task-0", "video-0", 0, batchfactoryv11.ProductionSucceeded, "https://media.example/0.mp4", time.Unix(10, 0)),
		}},
		{ID: "job-b", BookID: "book-1", DirectorRevisionID: "director-1", Tasks: []batchfactoryv11.ProductionTask{
			automationTask("task-1", "video-1", 0, batchfactoryv11.ProductionSucceeded, "https://media.example/1.mp4", time.Unix(20, 0)),
			automationTask("task-2", "video-2", 0, batchfactoryv11.ProductionSucceeded, "https://media.example/2.mp4", time.Unix(30, 0)),
		}},
	}
	production, complete := automationSucceededProduction(jobs, batch, book)
	if !complete {
		t.Fatalf("3 succeeded videos across jobs must count as complete: %+v", production.Tasks)
	}
	if len(production.Tasks) != 3 {
		t.Fatalf("production tasks = %d, want 3 reused succeeded videos", len(production.Tasks))
	}
}

// 同一 ordinal 的 orphaned 槽位里已经有成功成片时，视频阶段同样视为完成，
// 再次触发不应产生新的视频任务（新任务是 orphaned 与额度浪费的来源）。
func TestAutomationSucceededProductionReusesOrphanedSlot(t *testing.T) {
	batch := batchfactoryv11.Batch{ID: "batch-1"}
	book := batchfactoryv11.Book{
		ID:               "book-1",
		Videos:           []batchfactoryv11.Video{{ID: "video-active-0", Ordinal: 0}, {ID: "video-active-1", Ordinal: 1}},
		DirectorRevision: &batchfactoryv11.DirectorRevision{ID: "director-2"},
		VideoHistory: []batchfactoryv11.Video{
			{ID: "video-old-0", Ordinal: 0, CompatibilityState: "orphaned"},
		},
	}
	jobs := []batchfactoryv11.ProductionJob{
		{ID: "job-old", BookID: "book-1", DirectorRevisionID: "director-1", Tasks: []batchfactoryv11.ProductionTask{
			automationTask("task-old-0", "video-old-0", 0, batchfactoryv11.ProductionSucceeded, "https://media.example/old-0.mp4", time.Unix(10, 0)),
		}},
		{ID: "job-new", BookID: "book-1", DirectorRevisionID: "director-2", Tasks: []batchfactoryv11.ProductionTask{
			automationTask("task-1", "video-active-1", 0, batchfactoryv11.ProductionSucceeded, "https://media.example/1.mp4", time.Unix(20, 0)),
		}},
	}
	if _, complete := automationSucceededProduction(jobs, batch, book); !complete {
		t.Fatal("orphaned succeeded video of ordinal 0 must be reused instead of resubmitting video-0")
	}
}

func TestAutomationSucceededProductionRequiresEveryOrdinal(t *testing.T) {
	batch := batchfactoryv11.Batch{ID: "batch-1"}
	book := batchfactoryv11.Book{
		ID:     "book-1",
		Videos: []batchfactoryv11.Video{{ID: "video-0", Ordinal: 0}, {ID: "video-1", Ordinal: 1}},
	}
	jobs := []batchfactoryv11.ProductionJob{
		{ID: "job-a", BookID: "book-1", Tasks: []batchfactoryv11.ProductionTask{
			automationTask("task-0", "video-0", 0, batchfactoryv11.ProductionSucceeded, "https://media.example/0.mp4", time.Unix(10, 0)),
		}},
	}
	if _, complete := automationSucceededProduction(jobs, batch, book); complete {
		t.Fatal("a book missing the second storyboard video must not be treated as complete")
	}
}

// 开启换开头后，即使原版分镜视频都已存在，只要配置要求的换开头提示词还没生成，
// 自动化就不能把视频阶段当成可复用完成，否则会跳过换开头直接合成并在上传时失败。
func TestAutomationSucceededProductionWaitsForConfiguredOpeningVariants(t *testing.T) {
	batch := batchfactoryv11.Batch{ID: "batch-1", SettingsState: batchfactoryv11.SettingsState{Patch: automationSettings(map[string]any{"openingEnabled": true, "openingCount": 2})}}
	book := batchfactoryv11.Book{
		ID:           "book-1",
		Videos:       []batchfactoryv11.Video{{ID: "video-0", Ordinal: 0}, {ID: "video-1", Ordinal: 1}},
		VideoHistory: []batchfactoryv11.Video{{ID: "video-old-0", Ordinal: 0, CompatibilityState: "orphaned"}},
	}
	jobs := []batchfactoryv11.ProductionJob{
		{ID: "job-a", BookID: "book-1", Tasks: []batchfactoryv11.ProductionTask{
			automationTask("task-0", "video-0", 0, batchfactoryv11.ProductionSucceeded, "https://media.example/0.mp4", time.Unix(10, 0)),
			automationTask("task-1", "video-1", 0, batchfactoryv11.ProductionSucceeded, "https://media.example/1.mp4", time.Unix(20, 0)),
		}},
	}
	if _, complete := automationSucceededProduction(jobs, batch, book); complete {
		t.Fatal("configured opening variant must be generated before completed base videos can be reused")
	}
	variants, err := json.Marshal([]batchfactoryv11.OpeningVariant{{Index: 1, Label: "换开头1", Status: "success", Prompt: "雨夜推门而入。"}})
	if err != nil {
		t.Fatal(err)
	}
	book.Videos[0].SettingsState.Patch = batchfactoryv11.SettingsPatch{"openingVariants": variants}
	if _, complete := automationSucceededProduction(jobs, batch, book); complete {
		t.Fatal("a generated opening variant without a video must keep the video stage incomplete")
	}
	jobs[0].Tasks = append(jobs[0].Tasks, automationTask("task-opening-1", "video-old-0", 1, batchfactoryv11.ProductionSucceeded, "https://media.example/opening-1.mp4", time.Unix(30, 0)))
	if _, complete := automationSucceededProduction(jobs, batch, book); !complete {
		t.Fatal("a generated opening variant with a succeeded video must complete the video stage")
	}
}

func TestAutomationResumeStageReopensMissingOpeningBeforeVideo(t *testing.T) {
	batch := batchfactoryv11.Batch{ID: "batch-1", SettingsState: batchfactoryv11.SettingsState{Patch: automationSettings(map[string]any{"openingEnabled": true, "openingCount": 2})}}
	book := batchfactoryv11.Book{ID: "book-1", Videos: []batchfactoryv11.Video{{ID: "video-0", Ordinal: 0}, {ID: "video-1", Ordinal: 1}}}

	if got := automationResumeStageForBook("video", batch, book); got != "opening" {
		t.Fatalf("resume stage = %q, want opening so the missing opening prompt is generated first", got)
	}

	variants, err := json.Marshal([]batchfactoryv11.OpeningVariant{{Index: 1, Label: "换开头1", Status: "success", Prompt: "雨夜推门而入。"}})
	if err != nil {
		t.Fatal(err)
	}
	book.Videos[0].SettingsState.Patch = batchfactoryv11.SettingsPatch{"openingVariants": variants}
	if got := automationResumeStageForBook("video", batch, book); got != "video" {
		t.Fatalf("resume stage = %q, want video after opening prompts are complete", got)
	}
}

func TestAutomationBookUploadedSkipsRepeatedUpload(t *testing.T) {
	if automationBookUploaded(batchfactoryv11.Book{}) {
		t.Fatal("a book without an upload receipt must not be treated as uploaded")
	}
	uploaded := batchfactoryv11.Book{SourceMetadata: map[string]any{"websiteSubmitStatus": "uploaded"}}
	if !automationBookUploaded(uploaded) {
		t.Fatal("uploaded books must skip a second upload")
	}
	published := batchfactoryv11.Book{SourceMetadata: map[string]any{"publishStatus": "uploaded"}}
	if !automationBookUploaded(published) {
		t.Fatal("publishStatus=uploaded must also count as uploaded")
	}
}

func TestAutomationAssetsCompletion(t *testing.T) {
	off := automationSettings(map[string]any{"assetImagesEnabled": false})
	on := automationSettings(map[string]any{"assetImagesEnabled": true})
	book := batchfactoryv11.Book{ID: "book-1"}
	if !automationAssetsComplete(book, off) {
		t.Fatal("assets switch off must count as complete")
	}
	if automationAssetsComplete(book, on) {
		t.Fatal("assets switch on without any asset record must stay incomplete")
	}
	book.AssetRecords = []batchfactoryv11.BookAsset{{Prompt: "角色描述"}}
	if !automationAssetsComplete(book, on) {
		t.Fatal("existing asset records must count as complete")
	}
}

func TestAutomationDirectorCompletionFollowsSourceDigest(t *testing.T) {
	digest := batchfactoryv11.SourceInputDigest("正文内容")
	book := batchfactoryv11.Book{ID: "book-1", SourceText: "正文内容"}
	if automationDirectorComplete(book) {
		t.Fatal("book without a director revision must stay incomplete")
	}
	book.Videos = []batchfactoryv11.Video{{ID: "video-0", Ordinal: 0}}
	book.DirectorRevision = &batchfactoryv11.DirectorRevision{ID: "director-1", SourceDigest: digest}
	if !automationDirectorComplete(book) {
		t.Fatal("same input digest must count as complete")
	}
	book.SourceText = "正文内容（已修改）"
	if automationDirectorComplete(book) {
		t.Fatal("changed source text must force the director stage to run again")
	}
}
