package batchfactoryv11

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

// seedOpeningBook builds a batch with opening replacement enabled and, when a
// storyboard is given, a persisted director revision whose SD cards drive the
// book's VIDEO records.
func seedOpeningBook(t *testing.T, openingEnabled bool, openingCount int, storyboard []DirectorVideo) (*MemoryStore, Batch, Book) {
	t.Helper()
	store := NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", CreateBatchInput{Title: "B", Books: []CreateBookInput{{Title: "K", SourceText: "她被当众羞辱后沉默离开。"}}})
	if err != nil {
		t.Fatal(err)
	}
	patch := SettingsPatch{"openingEnabled": rawSetting(t, openingEnabled), "openingCount": rawSetting(t, openingCount)}
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeBatch, BatchID: batch.ID}, SettingsUpdate{Patch: patch, ExpectedRevision: batch.Revision}); err != nil {
		t.Fatal(err)
	}
	batch, err = store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	if len(storyboard) > 0 {
		if _, err := store.PersistDirectorRevision(context.Background(), "alice", book, DirectorSnapshot{Mode: "original", MaxVideoDuration: 10, AspectRatio: "9:16"}, "opening-test-digest", "", DirectorResult{Storyboard: storyboard}); err != nil {
			t.Fatal(err)
		}
		batch, err = store.GetBatch(context.Background(), "alice", batch.ID)
		if err != nil {
			t.Fatal(err)
		}
		book = batch.Books[0]
	}
	return store, batch, book
}

func openingVariantsFromPatch(t *testing.T, store *MemoryStore, batchID, bookID, videoID string) []OpeningVariant {
	t.Helper()
	patch := store.DebugPatch(ScopeRef{Kind: ScopeVideo, BatchID: batchID, BookID: bookID, VideoID: videoID})
	raw, ok := patch["openingVariants"]
	if !ok {
		return nil
	}
	var variants []OpeningVariant
	if err := json.Unmarshal(raw, &variants); err != nil {
		t.Fatalf("openingVariants patch is not valid JSON: %v", err)
	}
	return variants
}

func TestRunOpeningVariantsSkipsWhenDisabledOrSingleStoryboard(t *testing.T) {
	sdStoryboard := []DirectorVideo{
		{DurationSec: 10, FinalPrompt: "原分镜一：新娘低头抚平嫁衣袖口。"},
		{DurationSec: 10, FinalPrompt: "分镜二：陆沉推门而入。"},
	}
	t.Run("opening disabled", func(t *testing.T) {
		store, batch, book := seedOpeningBook(t, false, 4, sdStoryboard)
		provider := &queuedDirectorProvider{}
		variants, err := (&DirectorService{Store: store, Provider: provider}).RunOpeningVariants(context.Background(), "alice", batch.ID, book.ID, PresetSnapshot{})
		if err != nil {
			t.Fatal(err)
		}
		if len(variants) != 0 {
			t.Fatalf("variants = %d, want 0", len(variants))
		}
		if len(provider.calls) != 0 {
			t.Fatalf("provider must not be called when opening is disabled, calls=%d", len(provider.calls))
		}
	})
	t.Run("single storyboard", func(t *testing.T) {
		store, batch, book := seedOpeningBook(t, true, 4, sdStoryboard[:1])
		provider := &queuedDirectorProvider{}
		variants, err := (&DirectorService{Store: store, Provider: provider}).RunOpeningVariants(context.Background(), "alice", batch.ID, book.ID, PresetSnapshot{})
		if err != nil {
			t.Fatal(err)
		}
		if len(variants) != 0 {
			t.Fatalf("variants = %d, want 0", len(variants))
		}
		if len(provider.calls) != 0 {
			t.Fatalf("provider must not be called for a single storyboard, calls=%d", len(provider.calls))
		}
	})
}

func TestRunOpeningVariantsGeneratesForStructuredFirstStoryboard(t *testing.T) {
	// 结构化分镜用镜头/动作描述而不是整段 FinalPrompt；换开头必须照样生成，
	// 契约里把分镜一渲染成“镜头画面”文本喂给元提示词模型。
	structured := []DirectorVideo{
		{DurationSec: 10, VideoDesc: "林晚进入客厅", Shots: []DirectorShot{{StartSec: 0, EndSec: 10, ShotType: "中景", Camera: "缓慢推轨", Description: "林晚进入客厅"}}},
		{DurationSec: 10, VideoDesc: "陆沉推门而入", Shots: []DirectorShot{{StartSec: 0, EndSec: 10, ShotType: "全景", Camera: "固定", Description: "陆沉推门而入"}}},
	}
	store, batch, book := seedOpeningBook(t, true, 4, structured)
	reply := strings.Join([]string{
		"===VARIANT 1===",
		"时长：10秒",
		"变体一：茶盏砸落大理石地面碎裂，争吵爆发。",
		"",
		"===VARIANT 2===",
		"时长：10秒",
		"变体二：雨夜推门而入，两人对视沉默。",
		"",
		"===VARIANT 3===",
		"时长：10秒",
		"变体三：庭院灯影摇晃，她攥紧婚书。",
	}, "\n")
	provider := &queuedDirectorProvider{values: []string{reply}}
	variants, err := (&DirectorService{Store: store, Provider: provider}).RunOpeningVariants(context.Background(), "alice", batch.ID, book.ID, PresetSnapshot{})
	if err != nil {
		t.Fatal(err)
	}
	if len(variants) != 3 {
		t.Fatalf("variants = %d, want 3", len(variants))
	}
	if variants[0].Status != "success" || variants[0].DurationSec != 10 || !strings.Contains(variants[0].Prompt, "茶盏砸落") {
		t.Fatalf("variant[0] = %#v", variants[0])
	}
	if len(provider.calls) != 1 {
		t.Fatalf("provider calls = %d, want 1", len(provider.calls))
	}
	call := provider.calls[0]
	if !strings.Contains(call.UserPrompt, "镜头画面") || !strings.Contains(call.UserPrompt, "林晚进入客厅") {
		t.Fatalf("structured VIDEO01 must be rendered into the contract user prompt:\n%s", call.UserPrompt)
	}
	if !strings.Contains(call.UserPrompt, "陆沉推门而入") {
		t.Fatalf("follow-up storyboard must stay in the contract user prompt:\n%s", call.UserPrompt)
	}
	saved := openingVariantsFromPatch(t, store, batch.ID, book.ID, book.Videos[0].ID)
	if len(saved) != 3 || saved[0].Status != "success" || !strings.Contains(saved[0].Prompt, "茶盏砸落") {
		t.Fatalf("VIDEO01 patch must persist openingVariants, got %#v", saved)
	}
}

func TestRunOpeningVariantsSkipsBlankFirstStoryboard(t *testing.T) {
	// 分镜一既没有整段提示词，也没有任何镜头/动作描述（异常空数据）时，
	// 无法构造换开头契约，必须静默跳过且不调用模型。
	store, batch, book := seedOpeningBook(t, true, 4, []DirectorVideo{
		{DurationSec: 10},
		{DurationSec: 10, FinalPrompt: "分镜二：陆沉推门而入。"},
	})
	provider := &queuedDirectorProvider{}
	variants, err := (&DirectorService{Store: store, Provider: provider}).RunOpeningVariants(context.Background(), "alice", batch.ID, book.ID, PresetSnapshot{})
	if err != nil {
		t.Fatal(err)
	}
	if len(variants) != 0 {
		t.Fatalf("variants = %d, want 0", len(variants))
	}
	if len(provider.calls) != 0 {
		t.Fatalf("provider must not be called for a blank first storyboard, calls=%d", len(provider.calls))
	}
}

func TestRunOpeningVariantsGeneratesAndPersistsVariants(t *testing.T) {
	store, batch, book := seedOpeningBook(t, true, 4, []DirectorVideo{
		{DurationSec: 10, FinalPrompt: "原分镜一：新娘低头抚平嫁衣袖口。"},
		{DurationSec: 10, FinalPrompt: "分镜二：陆沉推门而入。"},
	})
	reply := strings.Join([]string{
		"===VARIANT 1===",
		"时长：10秒",
		"变体一：茶盏砸落大理石地面碎裂，争吵爆发。",
		"",
		"===VARIANT 2===",
		"时长：10秒",
		"变体二：雨夜推门而入，两人对视沉默。",
		"",
		"===VARIANT 3===",
		"时长：10秒",
		"变体三：庭院灯影摇晃，她攥紧婚书。",
	}, "\n")
	provider := &queuedDirectorProvider{values: []string{reply}}
	variants, err := (&DirectorService{Store: store, Provider: provider}).RunOpeningVariants(context.Background(), "alice", batch.ID, book.ID, PresetSnapshot{})
	if err != nil {
		t.Fatal(err)
	}
	if len(variants) != 3 {
		t.Fatalf("variants = %d, want 3 (2 success + 1 failed)", len(variants))
	}
	if variants[0].Status != "success" || variants[0].DurationSec != 10 || !strings.Contains(variants[0].Prompt, "茶盏砸落") {
		t.Fatalf("variant[0] = %#v", variants[0])
	}
	if variants[1].Status != "success" || !strings.Contains(variants[1].Prompt, "雨夜推门而入") {
		t.Fatalf("variant[1] = %#v", variants[1])
	}
	if variants[2].Status != "success" || !strings.Contains(variants[2].Prompt, "庭院灯影") {
		t.Fatalf("variant[2] = %#v", variants[2])
	}
	if len(provider.calls) != 1 {
		t.Fatalf("provider calls = %d, want 1", len(provider.calls))
	}
	call := provider.calls[0]
	if !strings.Contains(call.UserPrompt, "原分镜一：新娘低头抚平嫁衣袖口。") || !strings.Contains(call.UserPrompt, "分镜二：陆沉推门而入。") {
		t.Fatalf("contract user prompt must carry the original VIDEO01 and follow-up storyboard:\n%s", call.UserPrompt)
	}
	saved := openingVariantsFromPatch(t, store, batch.ID, book.ID, book.Videos[0].ID)
	if len(saved) != 3 || saved[0].Status != "success" || !strings.Contains(saved[0].Prompt, "茶盏砸落") {
		t.Fatalf("VIDEO01 patch must persist openingVariants, got %#v", saved)
	}
}

func TestRunOpeningVariantsFailsWhenARequiredVariantIsMissing(t *testing.T) {
	store, batch, book := seedOpeningBook(t, true, 4, []DirectorVideo{
		{DurationSec: 10, FinalPrompt: "原分镜一：新娘低头抚平嫁衣袖口。"},
		{DurationSec: 10, FinalPrompt: "分镜二：陆沉推门而入。"},
	})
	provider := &queuedDirectorProvider{values: []string{strings.Join([]string{
		"===VARIANT 1===", "时长：10秒", "变体一：茶盏砸落大理石地面碎裂。", "",
		"===VARIANT 2===", "时长：10秒", "变体二：雨夜推门而入。",
	}, "\n")}}
	variants, err := (&DirectorService{Store: store, Provider: provider}).RunOpeningVariants(context.Background(), "alice", batch.ID, book.ID, PresetSnapshot{})
	if err == nil {
		t.Fatal("missing required variant must fail the opening stage")
	}
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("missing model variant must be retryable, got %v", err)
	}
	if len(variants) != 3 || variants[2].Status != "failed" || variants[2].FailureReason == "" {
		t.Fatalf("variants = %#v", variants)
	}
	saved := openingVariantsFromPatch(t, store, batch.ID, book.ID, book.Videos[0].ID)
	if len(saved) != 3 || saved[2].Status != "failed" || saved[2].FailureReason == "" {
		t.Fatalf("saved variants = %#v", saved)
	}
}

func TestBuildOpeningVariantsContractAppendsCanonicalOutputFormat(t *testing.T) {
	contract := BuildOpeningVariantsContract(
		Book{Title: "K", SourceText: "正文"},
		DirectorVideo{DurationSec: 10, FinalPrompt: "分镜一"},
		[]DirectorVideo{{DurationSec: 10, FinalPrompt: "分镜二"}},
		PresetSnapshot{Body: "旧版规则要求把时长与标记写在一行。"},
		3,
		10,
	)
	if !strings.Contains(contract.SystemPrompt, "===VARIANT 1===\n时长：10秒\n可直接替换的分镜一提示词正文") {
		t.Fatalf("contract missing canonical machine format:\n%s", contract.SystemPrompt)
	}
}

func TestRunOpeningVariantsKeepsStoredSuccessesWhenReplyIsInvalid(t *testing.T) {
	store, batch, book := seedOpeningBook(t, true, 4, []DirectorVideo{
		{DurationSec: 10, FinalPrompt: "原分镜一：新娘低头抚平嫁衣袖口。"},
		{DurationSec: 10, FinalPrompt: "分镜二：陆沉推门而入。"},
	})
	good := strings.Join([]string{
		"===VARIANT 1===",
		"时长：10秒",
		"变体一：茶盏砸落大理石地面碎裂，争吵爆发。",
		"",
		"===VARIANT 2===",
		"时长：10秒",
		"变体二：雨夜推门而入，两人对视沉默。",
		"",
		"===VARIANT 3===",
		"时长：10秒",
		"变体三：庭院灯影摇晃，她攥紧婚书。",
	}, "\n")
	service := &DirectorService{Store: store, Provider: &queuedDirectorProvider{values: []string{good}}}
	if _, err := service.RunOpeningVariants(context.Background(), "alice", batch.ID, book.ID, PresetSnapshot{}); err != nil {
		t.Fatal(err)
	}
	service.Provider = &queuedDirectorProvider{values: []string{"这不是合法的变体输出，没有任何分段标记。"}}
	variants, err := service.RunOpeningVariants(context.Background(), "alice", batch.ID, book.ID, PresetSnapshot{})
	if err != nil {
		t.Fatal(err)
	}
	if len(variants) != 3 {
		t.Fatalf("variants = %d, want 3", len(variants))
	}
	for i, variant := range variants {
		if variant.Status != "success" {
			t.Fatalf("variant[%d].status = %q, want success", i, variant.Status)
		}
	}
	saved := openingVariantsFromPatch(t, store, batch.ID, book.ID, book.Videos[0].ID)
	if len(saved) != 3 {
		t.Fatalf("saved variants = %d, want 3", len(saved))
	}
	if saved[0].Status != "success" || !strings.Contains(saved[0].Prompt, "茶盏砸落") {
		t.Fatalf("stored success variant 1 must survive a failed rerun, got %#v", saved[0])
	}
	if saved[1].Status != "success" || !strings.Contains(saved[1].Prompt, "雨夜推门而入") {
		t.Fatalf("stored success variant 2 must survive a failed rerun, got %#v", saved[1])
	}
}

func TestRunOpeningVariantsRetriesOnceAfterTruncatedOutput(t *testing.T) {
	store, batch, book := seedOpeningBook(t, true, 4, []DirectorVideo{
		{DurationSec: 10, FinalPrompt: "原分镜一：新娘低头抚平嫁衣袖口。"},
		{DurationSec: 10, FinalPrompt: "分镜二：陆沉推门而入。"},
	})
	reply := strings.Join([]string{
		"===VARIANT 1===", "时长：10秒", "变体一：茶盏砸落大理石地面碎裂。", "",
		"===VARIANT 2===", "时长：10秒", "变体二：雨夜推门而入。", "",
		"===VARIANT 3===", "时长：10秒", "变体三：庭院灯影摇晃，她攥紧婚书。",
	}, "\n")
	provider := &truncatingDirectorProvider{values: []string{"truncated", reply}}
	variants, err := (&DirectorService{Store: store, Provider: provider}).RunOpeningVariants(context.Background(), "alice", batch.ID, book.ID, PresetSnapshot{})
	if err != nil {
		t.Fatalf("opening variants should recover from one truncated response: %v", err)
	}
	if len(provider.calls) != 2 {
		t.Fatalf("provider calls = %d, want exactly one retry", len(provider.calls))
	}
	if !strings.Contains(provider.calls[1].SystemPrompt, "更紧凑") {
		t.Fatalf("retry must ask for compact output: %q", provider.calls[1].SystemPrompt)
	}
	if len(variants) != 3 || variants[2].Status != "success" || !strings.Contains(variants[2].Prompt, "庭院灯影") {
		t.Fatalf("variants = %#v", variants)
	}
}

func TestRunOpeningVariantsFailsAfterSecondTruncation(t *testing.T) {
	store, batch, book := seedOpeningBook(t, true, 4, []DirectorVideo{
		{DurationSec: 10, FinalPrompt: "原分镜一：新娘低头抚平嫁衣袖口。"},
		{DurationSec: 10, FinalPrompt: "分镜二：陆沉推门而入。"},
	})
	provider := &alwaysTruncatingDirectorProvider{}
	_, err := (&DirectorService{Store: store, Provider: provider}).RunOpeningVariants(context.Background(), "alice", batch.ID, book.ID, PresetSnapshot{})
	if err == nil || !strings.Contains(err.Error(), "finish_reason=length") {
		t.Fatalf("err=%v, want the provider truncation reason preserved", err)
	}
	if provider.calls != 2 {
		t.Fatalf("provider calls = %d, want the retry to stay bounded at 2 attempts", provider.calls)
	}
}
