package batchfactoryv11

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

type repairRecordingProvider struct {
	calls int
	reply string
	err   error
}

func (p *repairRecordingProvider) Complete(context.Context, TextCompletionRequest) (string, error) {
	p.calls++
	return p.reply, p.err
}

func TestRewriteRejectedVideoPromptDoesNotMisreportProviderFailureAsFormatError(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	batch, err := store.CreateBatch(ctx, "alice", CreateBatchInput{Title: "b", Books: []CreateBookInput{{
		Title:      "book",
		SourceText: "正文",
		Videos:     []CreateVideoInput{{Label: "VIDEO01", VideoPrompt: "人物站在桌前。"}},
	}}})
	if err != nil {
		t.Fatal(err)
	}
	providerErr := &DirectorHTTPError{StatusCode: 502, Model: "gpt-test", Body: `{"error":{"message":"upstream unavailable"}}`}
	provider := &repairRecordingProvider{err: providerErr}
	service := &BookStageService{Store: store, Director: &DirectorService{Store: store, Provider: provider}}

	_, err = service.rewriteRejectedVideoPrompt(ctx, "alice", batch.ID, batch.Books[0].ID, "", "内容不符合平台规范")
	if err == nil {
		t.Fatal("want the provider failure")
	}
	if strings.Contains(err.Error(), "文本模型输出不符合当前格式要求") {
		t.Fatalf("provider failure was misreported as a format error: %v", err)
	}
	if !errors.Is(err, ErrUpstream) {
		t.Fatalf("provider failure must preserve ErrUpstream, got %v", err)
	}

	fallback, err := service.rewriteRejectedVideoPrompt(ctx, "alice", batch.ID, batch.Books[0].ID, "林倩当众撕毁身份证和户口本，破损证件留在桌面。", "内容不符合平台规范")
	if err != nil {
		t.Fatalf("transient channel outage must use deterministic safety fallback: %v", err)
	}
	if strings.Contains(fallback, "撕毁") || strings.Contains(fallback, "破损证件") {
		t.Fatalf("unsafe fallback=%q", fallback)
	}

	secondPass, err := service.rewriteRejectedVideoPrompt(ctx, "alice", batch.ID, batch.Books[0].ID, fallback, "内容不符合平台规范")
	if err != nil {
		t.Fatalf("an already sanitized prompt must remain usable while the text channel is down: %v", err)
	}
	if secondPass != fallback {
		t.Fatalf("second fallback=%q, want stable prompt %q", secondPass, fallback)
	}
}

func TestRepairRejectedVideoPromptRewritesWithinBudget(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	batch, err := store.CreateBatch(ctx, "alice", CreateBatchInput{Title: "b", Books: []CreateBookInput{{
		Title:      "book",
		SourceText: "正文",
		Videos:     []CreateVideoInput{{Label: "VIDEO01", VideoPrompt: "林晚举刀冲向对方，鲜血直流。"}},
	}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	video := book.Videos[0]
	original, _ := json.Marshal("林晚举刀冲向对方，鲜血直流。")
	if _, err := store.SaveSettings(ctx, "alice", ScopeRef{Kind: ScopeVideo, BatchID: batch.ID, BookID: book.ID, VideoID: video.ID}, SettingsUpdate{Patch: SettingsPatch{"videoPrompt": original}, ExpectedRevision: video.Revision}); err != nil {
		t.Fatal(err)
	}
	repository, err := (&ProductionService{Store: store}).repository()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repository.CreateProductionJob(ctx, ProductionJob{
		Owner: "alice", BatchID: batch.ID, BookID: book.ID, RequestID: "req-1", DirectorRevisionID: "rev-1",
		Status: ProductionFailed,
		Tasks:  []ProductionTask{{VideoID: video.ID, Status: ProductionFailed, ErrorMessage: "视频任务失败：内容不符合平台规范"}},
	}); err != nil {
		t.Fatal(err)
	}
	provider := &repairRecordingProvider{reply: "林晚握紧手中物件，与对方对峙。"}
	service := &BookStageService{Store: store, Director: &DirectorService{Store: store, Provider: provider}, Production: &ProductionService{Store: store}}

	repaired, err := service.RepairRejectedVideoPrompts(ctx, "alice", batch.ID, book.ID, "")
	if err != nil {
		t.Fatal(err)
	}
	if !repaired {
		t.Fatal("want the rejected prompt to be repaired")
	}
	if provider.calls != 1 {
		t.Fatalf("provider calls=%d, want 1", provider.calls)
	}
	latest, err := store.GetBatch(ctx, "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	settings := latest.Books[0].Videos[0].SettingsState.Patch
	var prompt string
	if err := json.Unmarshal(settings["videoPrompt"], &prompt); err != nil {
		t.Fatal(err)
	}
	if prompt != "林晚握紧手中物件，与对方对峙。" {
		t.Fatalf("videoPrompt=%q, want the rewritten prompt", prompt)
	}
	var counts map[string]int
	if err := json.Unmarshal(settings[videoPromptRepairCountsKey], &counts); err != nil {
		t.Fatal(err)
	}
	if counts["0"] != 1 {
		t.Fatalf("repair counts=%v, want variant 0 repaired once", counts)
	}

	// 第二轮改写仍然允许，直到 videoPromptRepairMaxRounds 用尽。
	if repaired, err := service.RepairRejectedVideoPrompts(ctx, "alice", batch.ID, book.ID, ""); err != nil || !repaired {
		t.Fatalf("second repair=%v err=%v, want another rewrite", repaired, err)
	}
	if provider.calls != 2 {
		t.Fatalf("provider calls=%d, want 2 rounds", provider.calls)
	}
	if _, err := service.RepairRejectedVideoPrompts(ctx, "alice", batch.ID, book.ID, ""); !errors.Is(err, ErrVideoPromptRepairExhausted) {
		t.Fatalf("third repair error=%v, want ErrVideoPromptRepairExhausted", err)
	}
	if provider.calls != 2 {
		t.Fatalf("exhausted repair must not call the model again, calls=%d", provider.calls)
	}
}

func TestRepairRejectedVideoPromptAllowsOneCompiledSafetyMigrationRetry(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	batch, err := store.CreateBatch(ctx, "alice", CreateBatchInput{Title: "b", Books: []CreateBookInput{{
		Title: "book", SourceText: "正文", Videos: []CreateVideoInput{{Label: "VIDEO01", VideoPrompt: "原提示词"}},
	}}})
	if err != nil {
		t.Fatal(err)
	}
	book, video := batch.Books[0], batch.Books[0].Videos[0]
	overrides, _ := json.Marshal(map[string]string{"1": "人物把证件收起，保持克制。"})
	counts, _ := json.Marshal(map[string]int{"1": videoPromptRepairMaxRounds})
	markers, _ := json.Marshal(map[string]string{"1": promptDigest("人物把证件收起，保持克制。")})
	_, err = store.SaveSettings(ctx, "alice", ScopeRef{Kind: ScopeVideo, BatchID: batch.ID, BookID: book.ID, VideoID: video.ID}, SettingsUpdate{
		Patch: SettingsPatch{openingVariantPromptsKey: overrides, videoPromptRepairCountsKey: counts, videoPromptRepairMarkersKey: markers}, ExpectedRevision: video.Revision,
	})
	if err != nil {
		t.Fatal(err)
	}
	updatedBatch, err := store.GetBatch(ctx, "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	video = updatedBatch.Books[0].Videos[0]
	repository, err := (&ProductionService{Store: store}).repository()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repository.CreateProductionJob(ctx, ProductionJob{Owner: "alice", BatchID: batch.ID, BookID: book.ID, RequestID: "failed", DirectorRevisionID: "director", Tasks: []ProductionTask{{
		VideoID: video.ID, OpeningVariantIndex: 1, Status: ProductionFailed, ErrorMessage: "内容审核拒绝",
	}}}); err != nil {
		t.Fatal(err)
	}
	service := &BookStageService{Store: store, Production: &ProductionService{Store: store}}
	if repaired, err := service.RepairRejectedVideoPrompts(ctx, "alice", batch.ID, book.ID, video.ID); err != nil || !repaired {
		t.Fatalf("migration repair=%v err=%v, want one full-compiled retry", repaired, err)
	}
	if repaired, err := service.RepairRejectedVideoPrompts(ctx, "alice", batch.ID, book.ID, video.ID); repaired || !errors.Is(err, ErrVideoPromptRepairExhausted) {
		t.Fatalf("second migration repair=%v err=%v, want exhausted", repaired, err)
	}
}

func TestRepairRejectedOpeningVariantUsesVariantOverride(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	batch, err := store.CreateBatch(ctx, "alice", CreateBatchInput{Title: "b", Books: []CreateBookInput{{
		Title:      "book",
		SourceText: "正文",
		Videos:     []CreateVideoInput{{Label: "VIDEO01", VideoPrompt: "林晚推门而入。"}, {Label: "VIDEO02"}},
	}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	video := book.Videos[0]
	variants, _ := json.Marshal([]OpeningVariant{{Index: 1, Label: "分镜一 | 换开头1", Prompt: "林晚掐住对方脖子，鲜血直流。", Status: "success"}})
	if _, err := store.SaveSettings(ctx, "alice", ScopeRef{Kind: ScopeVideo, BatchID: batch.ID, BookID: book.ID, VideoID: video.ID}, SettingsUpdate{Patch: SettingsPatch{"openingVariants": variants}, ExpectedRevision: video.Revision}); err != nil {
		t.Fatal(err)
	}
	repository, err := (&ProductionService{Store: store}).repository()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repository.CreateProductionJob(ctx, ProductionJob{
		Owner: "alice", BatchID: batch.ID, BookID: book.ID, RequestID: "req-2", DirectorRevisionID: "rev-1",
		Status: ProductionFailed,
		Tasks:  []ProductionTask{{VideoID: video.ID, OpeningVariantIndex: 1, Status: ProductionFailed, ErrorMessage: "内容审核拒绝"}},
	}); err != nil {
		t.Fatal(err)
	}
	provider := &repairRecordingProvider{reply: "林晚逼近对方，与对方对峙。"}
	service := &BookStageService{Store: store, Director: &DirectorService{Store: store, Provider: provider}, Production: &ProductionService{Store: store}}

	repaired, err := service.RepairRejectedVideoPrompts(ctx, "alice", batch.ID, book.ID, "")
	if err != nil {
		t.Fatal(err)
	}
	if !repaired {
		t.Fatal("want the rejected opening variant to be repaired")
	}
	latest, err := store.GetBatch(ctx, "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	settings := latest.Books[0].Videos[0].SettingsState.Patch
	if raw, ok := settings["videoPrompt"]; ok && len(raw) > 0 {
		t.Fatalf("variant repair must not write the VIDEO01 override: %s", raw)
	}
	var overrides map[string]string
	if err := json.Unmarshal(settings[openingVariantPromptsKey], &overrides); err != nil {
		t.Fatal(err)
	}
	if overrides["1"] != "林晚逼近对方，与对方对峙。" {
		t.Fatalf("variant overrides=%v, want the rewritten prompt", overrides)
	}
	var storedVariants []OpeningVariant
	if err := json.Unmarshal(settings["openingVariants"], &storedVariants); err != nil {
		t.Fatal(err)
	}
	if len(storedVariants) != 1 || storedVariants[0].Prompt != "林晚逼近对方，与对方对峙。" {
		t.Fatalf("openingVariants=%+v, want the visible prompt updated", storedVariants)
	}
}

func TestVideoPromptOverrideDoesNotLeakIntoOpeningVariants(t *testing.T) {
	video := Video{SettingsState: SettingsState{Patch: SettingsPatch{
		"videoPrompt":            rawSetting(t, "VIDEO01 覆盖"),
		openingVariantPromptsKey: rawSetting(t, map[string]string{"1": "换开头1 覆盖"}),
	}}}
	if value, ok := videoPromptOverrideFor(video, 0); !ok || value != "VIDEO01 覆盖" {
		t.Fatalf("variant 0 override=%q ok=%v", value, ok)
	}
	if value, ok := videoPromptOverrideFor(video, 1); !ok || value != "换开头1 覆盖" {
		t.Fatalf("variant 1 override=%q ok=%v", value, ok)
	}
	videoOnly := Video{SettingsState: SettingsState{Patch: SettingsPatch{"videoPrompt": rawSetting(t, "VIDEO01 覆盖")}}}
	if value, ok := videoPromptOverrideFor(videoOnly, 1); ok {
		t.Fatalf("videoPrompt must not leak into opening variants, got %q", value)
	}
}

// 用户手动改过提示词后，1 轮自动改写预算不能把人永久拦住。
func TestRepairLetsManualPromptEditBypassExhausted(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	batch, err := store.CreateBatch(ctx, "alice", CreateBatchInput{Title: "b", Books: []CreateBookInput{{
		Title:      "book",
		SourceText: "正文",
		Videos:     []CreateVideoInput{{Label: "VIDEO01", VideoPrompt: "原始提示词"}},
	}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	video := book.Videos[0]
	if _, err := store.SaveSettings(ctx, "alice", ScopeRef{Kind: ScopeVideo, BatchID: batch.ID, BookID: book.ID, VideoID: video.ID}, SettingsUpdate{Patch: SettingsPatch{"videoPrompt": rawSetting(t, "原始提示词")}, ExpectedRevision: video.Revision}); err != nil {
		t.Fatal(err)
	}
	repository, err := (&ProductionService{Store: store}).repository()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repository.CreateProductionJob(ctx, ProductionJob{
		Owner: "alice", BatchID: batch.ID, BookID: book.ID, RequestID: "req-3", DirectorRevisionID: "rev-1",
		Status: ProductionFailed,
		Tasks:  []ProductionTask{{VideoID: video.ID, Status: ProductionFailed, ErrorMessage: "内容审核拒绝"}},
	}); err != nil {
		t.Fatal(err)
	}
	provider := &repairRecordingProvider{reply: "自动改写后的提示词"}
	service := &BookStageService{Store: store, Director: &DirectorService{Store: store, Provider: provider}, Production: &ProductionService{Store: store}}
	if repaired, err := service.RepairRejectedVideoPrompts(ctx, "alice", batch.ID, book.ID, ""); err != nil || !repaired {
		t.Fatalf("first repair=%v err=%v", repaired, err)
	}
	if repaired, err := service.RepairRejectedVideoPrompts(ctx, "alice", batch.ID, book.ID, ""); err != nil || !repaired {
		t.Fatalf("second repair=%v err=%v", repaired, err)
	}
	if _, err := service.RepairRejectedVideoPrompts(ctx, "alice", batch.ID, book.ID, ""); !errors.Is(err, ErrVideoPromptRepairExhausted) {
		t.Fatalf("same rewritten prompt must be exhausted, got %v", err)
	}
	latest, err := store.GetBatch(ctx, "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	edited := latest.Books[0].Videos[0]
	if _, err := store.SaveSettings(ctx, "alice", ScopeRef{Kind: ScopeVideo, BatchID: batch.ID, BookID: book.ID, VideoID: video.ID}, SettingsUpdate{Patch: SettingsPatch{"videoPrompt": rawSetting(t, "用户手动修改后的提示词")}, ExpectedRevision: edited.Revision}); err != nil {
		t.Fatal(err)
	}
	repaired, err := service.RepairRejectedVideoPrompts(ctx, "alice", batch.ID, book.ID, "")
	if err != nil {
		t.Fatalf("manual edit must not be blocked: %v", err)
	}
	if repaired {
		t.Fatal("manual edit must be submitted as-is, not rewritten again")
	}
	if provider.calls != 2 {
		t.Fatalf("manual edit must not call the model again, calls=%d", provider.calls)
	}
}

// H3 冻结编译也必须服从 videoPrompt 覆盖，否则审核拒绝后的改写不生效。
func TestCompileH3ForProductionAppliesVideoPromptOverride(t *testing.T) {
	ctx := context.Background()
	store, batch, book := seedH3DirectorRevision(t)
	document := *book.DirectorRevision.Output.H3Director
	timeline := mustH3Timeline(t, document, 7420)
	timeline.DirectorRevisionID = book.DirectorRevision.ID
	persistedTimeline, err := store.PersistH3CanonicalTimeline(ctx, "alice", batch.ID, book.ID, timeline)
	if err != nil {
		t.Fatal(err)
	}
	input := completeH3CompileInput(document, timeline)
	input.TimelineID = persistedTimeline.ID
	compilation, err := CompileH3VideoSegments(input)
	if err != nil {
		t.Fatal(err)
	}
	compiled, err := store.PersistH3VideoCompilation(ctx, "alice", batch.ID, book.ID, compilation)
	if err != nil {
		t.Fatal(err)
	}
	videos, err := store.ApplyH3CompilationVideos(ctx, "alice", batch.ID, book.ID, compiled)
	if err != nil {
		t.Fatal(err)
	}
	book.Videos = videos
	compiler := &PromptCompilerService{Store: store}
	videoID := book.Videos[0].ID
	frozen, err := compiler.CompileH3ForProduction(ctx, "alice", batch.ID, book.ID, videoID, compiled.ID, 3)
	if err != nil {
		t.Fatal(err)
	}
	if frozen.CompiledPrompt != compiled.Compilation.Segments[0].CompiledPrompt {
		t.Fatal("frozen compilation prompt must be used when no override exists")
	}
	video, ok := videoByID(book, videoID)
	if !ok {
		t.Fatal("video not found")
	}
	if _, err := store.SaveSettings(ctx, "alice", ScopeRef{Kind: ScopeVideo, BatchID: batch.ID, BookID: book.ID, VideoID: videoID}, SettingsUpdate{Patch: SettingsPatch{"videoPrompt": rawSetting(t, "去敏后的 H3 提示词")}, ExpectedRevision: video.Revision}); err != nil {
		t.Fatal(err)
	}
	overridden, err := compiler.CompileH3ForProduction(ctx, "alice", batch.ID, book.ID, videoID, compiled.ID, 3)
	if err != nil {
		t.Fatal(err)
	}
	if overridden.CompiledPrompt != "去敏后的 H3 提示词" || overridden.DisplayPrompt != "去敏后的 H3 提示词" {
		t.Fatalf("override not applied: compiled=%q display=%q", overridden.CompiledPrompt, overridden.DisplayPrompt)
	}
	if overridden.SnapshotHash == frozen.SnapshotHash {
		t.Fatal("override must change the snapshot hash")
	}
}
