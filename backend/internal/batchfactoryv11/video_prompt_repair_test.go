package batchfactoryv11

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

type repairRecordingProvider struct {
	calls int
	reply string
}

func (p *repairRecordingProvider) Complete(context.Context, TextCompletionRequest) (string, error) {
	p.calls++
	return p.reply, nil
}

func TestRepairRejectedVideoPromptRewritesOverrideOnce(t *testing.T) {
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

	if _, err := service.RepairRejectedVideoPrompts(ctx, "alice", batch.ID, book.ID, ""); !errors.Is(err, ErrVideoPromptRepairExhausted) {
		t.Fatalf("second repair error=%v, want ErrVideoPromptRepairExhausted", err)
	}
	if provider.calls != 1 {
		t.Fatalf("exhausted repair must not call the model again, calls=%d", provider.calls)
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
