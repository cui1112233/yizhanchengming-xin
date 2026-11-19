package batchfactoryv11

import (
	"context"
	"slices"
	"strings"
	"testing"
)

func TestParseOpeningVariants(t *testing.T) {
	raw := strings.Join([]string{
		"===VARIANT 1===",
		"时长：10秒",
		"画面故事脚本：摔东西争吵的开场，茶盏砸落大理石地面碎裂。",
		"",
		"===VARIANT 2===",
		"时长：10秒",
		"画面故事脚本：雨夜推门而入，两人对视沉默后爆发争吵。",
	}, "\n")
	variants := parseOpeningVariants(raw, 15, 10, 2)
	if len(variants) != 2 {
		t.Fatalf("variants = %d, want 2", len(variants))
	}
	if variants[0].Index != 1 || variants[0].Status != "success" || variants[0].DurationSec != 10 {
		t.Fatalf("variant[0] = %#v", variants[0])
	}
	if !strings.Contains(variants[1].Prompt, "雨夜推门而入") {
		t.Fatalf("variant[1].prompt = %q", variants[1].Prompt)
	}
	if variants[0].Label != "分镜一 | 换开头1" {
		t.Fatalf("label = %q", variants[0].Label)
	}
}

func TestParseOpeningVariantsAcceptsInlineMarkerAndDuration(t *testing.T) {
	variants := parseOpeningVariants("===VARIANT 1=== 时长：10秒\n开头正文", 15, 10, 1)
	if len(variants) != 1 {
		t.Fatalf("variants = %d, want 1", len(variants))
	}
	if variants[0].Status != "success" || variants[0].DurationSec != 10 || variants[0].Prompt != "开头正文" {
		t.Fatalf("variant = %#v", variants[0])
	}
}

func TestParseOpeningVariantsAcceptsCommonModelMarkerDrift(t *testing.T) {
	raw := strings.Join([]string{
		"### VARIANT 1",
		"时长：10秒",
		"变体一正文",
		"",
		"**变体 2：**",
		"变体二正文",
		"",
		"=== 换开头 3 ===",
		"变体三正文",
	}, "\n")
	variants := parseOpeningVariants(raw, 15, 10, 3)
	for index, variant := range variants {
		if variant.Status != "success" {
			t.Fatalf("variant[%d] = %#v, want success", index, variant)
		}
	}
	if !strings.Contains(variants[1].Prompt, "变体二正文") {
		t.Fatalf("variant[1].prompt = %q", variants[1].Prompt)
	}
}

func TestParseOpeningVariantsFillsMissingAsFailed(t *testing.T) {
	raw := "===VARIANT 1===\n时长：10秒\n正文"
	variants := parseOpeningVariants(raw, 15, 10, 3)
	if len(variants) != 3 {
		t.Fatalf("variants = %d, want 3", len(variants))
	}
	if variants[0].Status != "success" {
		t.Fatalf("variant[0].status = %q", variants[0].Status)
	}
	for _, variant := range variants[1:] {
		if variant.Status != "failed" || variant.Prompt != "" {
			t.Fatalf("missing variant = %#v", variant)
		}
	}
}

func TestParseOpeningVariantsRejectsBadDuration(t *testing.T) {
	raw := "===VARIANT 1===\n时长：99秒\n正文"
	variants := parseOpeningVariants(raw, 15, 10, 1)
	if variants[0].Status != "failed" {
		t.Fatalf("expected failed variant, got %#v", variants[0])
	}
}

func TestParseOpeningVariantsInheritsOriginalDurationWhenModelOmitsDurationLine(t *testing.T) {
	raw := "===VARIANT 1===\n**换开头画面**\n茶盏砸落大理石地面碎裂，争吵爆发。"
	variants := parseOpeningVariants(raw, 15, 10, 1)
	if len(variants) != 1 {
		t.Fatalf("variants = %d, want 1", len(variants))
	}
	if variants[0].Status != "success" || variants[0].DurationSec != 10 {
		t.Fatalf("variant = %#v", variants[0])
	}
	if !strings.Contains(variants[0].Prompt, "茶盏砸落") {
		t.Fatalf("variant prompt = %q", variants[0].Prompt)
	}
}

func TestParseOpeningVariantsAcceptsChineseNumeralMarkers(t *testing.T) {
	raw := strings.Join([]string{
		"【变体一】",
		"时长：10秒",
		"茶盏砸落大理石地面碎裂。",
		"",
		"换开头二：",
		"雨夜推门而入，两人对视沉默。",
	}, "\n")
	variants := parseOpeningVariants(raw, 15, 10, 2)
	for index, variant := range variants {
		if variant.Status != "success" {
			t.Fatalf("variant[%d] = %#v, want success", index, variant)
		}
	}
	if !strings.Contains(variants[1].Prompt, "雨夜推门而入") {
		t.Fatalf("variant[1].prompt = %q", variants[1].Prompt)
	}
}

func TestParseOpeningVariantsAcceptsJSONReply(t *testing.T) {
	raw := "```json\n{\"variants\":[{\"index\":1,\"prompt\":\"茶盏砸落\"},{\"index\":2,\"content\":\"雨夜推门而入\"}]}\n```"
	variants := parseOpeningVariants(raw, 15, 10, 2)
	if len(variants) != 2 || variants[0].Status != "success" || variants[1].Status != "success" {
		t.Fatalf("variants = %#v", variants)
	}
	if !strings.Contains(variants[0].Prompt, "茶盏砸落") || !strings.Contains(variants[1].Prompt, "雨夜推门而入") {
		t.Fatalf("variants = %#v", variants)
	}
}

func TestParseOpeningVariantsAcceptsMarkerlessSingleVariant(t *testing.T) {
	raw := "时长：10秒\n茶盏砸落大理石地面碎裂，争吵爆发。"
	variants := parseOpeningVariants(raw, 15, 10, 1)
	if len(variants) != 1 || variants[0].Status != "success" {
		t.Fatalf("variants = %#v", variants)
	}
	if variants[0].Prompt != "茶盏砸落大理石地面碎裂，争吵爆发。" {
		t.Fatalf("variant prompt = %q", variants[0].Prompt)
	}
}

// 视频阶段的完成口径只能要求“已成功生成的变体”，不能按配置的
// openingCount 硬算 1..N：否则缺一个变体就会让整本永远停在“等待生成视频”。
func TestRequiredOpeningVariantIndicesUsesSuccessfullyGeneratedVariants(t *testing.T) {
	ctx := context.Background()
	store, batch, book := seedOpeningBook(t, true, 4, []DirectorVideo{
		{DurationSec: 10, FinalPrompt: "原开场"},
		{DurationSec: 10, FinalPrompt: "分镜二"},
	})
	if len(book.Videos) < 2 {
		t.Fatalf("videos = %d, want 2", len(book.Videos))
	}
	video := book.Videos[0]
	variants := []OpeningVariant{
		{Index: 1, Label: "换开头1", Status: "success", Prompt: "茶盏砸落大理石地面碎裂。"},
		{Index: 2, Label: "换开头2", Status: "failed", FailureReason: "模型未输出该变体分段"},
		{Index: 3, Label: "换开头3", Status: "success", Prompt: "雨夜推门而入。"},
	}
	if _, err := store.SaveSettings(ctx, "alice", ScopeRef{Kind: ScopeVideo, BatchID: batch.ID, BookID: book.ID, VideoID: video.ID}, SettingsUpdate{Patch: SettingsPatch{"openingVariants": rawSetting(t, variants)}, ExpectedRevision: video.Revision}); err != nil {
		t.Fatal(err)
	}
	batch, err := store.GetBatch(ctx, "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	book = batch.Books[0]
	indices := RequiredOpeningVariantIndices(batch, book)
	if want := []int{0, 1, 3}; !slices.Equal(indices, want) {
		t.Fatalf("indices = %v, want %v (failed 换开头2 must not block the video stage)", indices, want)
	}
}

// Novel Fetch 追加的书在导演前没有 VIDEO 槽位。自动化用批次快照判断时，
// 0 个 VIDEO 只能说明分镜还没生成，绝不能被当成“换开头已完成”；否则新书
// 会跳过换开头，最终合成和上传都缺 opening 1。
func TestOpeningStageCompleteTreatsMissingVideosAsIncomplete(t *testing.T) {
	batch := Batch{ID: "batch-1", SettingsState: SettingsState{Patch: SettingsPatch{
		"openingEnabled": rawSetting(t, true),
		"openingCount":   rawSetting(t, 2),
	}}}
	book := Book{ID: "book-1"}
	if OpeningStageComplete(batch, book) {
		t.Fatal("a book whose director has not created VIDEO slots must stay incomplete")
	}
	book.Videos = []Video{{ID: "video-1", Ordinal: 0}}
	if !OpeningStageComplete(batch, book) {
		t.Fatal("a single-storyboard book has no separate opening variant and must count as complete")
	}
	book.Videos = append(book.Videos, Video{ID: "video-2", Ordinal: 1})
	if OpeningStageComplete(batch, book) {
		t.Fatal("multi-storyboard book without opening variants must stay incomplete")
	}
	book.Videos[0].SettingsState.Patch = SettingsPatch{"openingVariants": rawSetting(t, []OpeningVariant{{Index: 1, Status: "success", Prompt: "雨夜推门而入。"}})}
	if !OpeningStageComplete(batch, book) {
		t.Fatal("a successfully generated opening variant must complete the stage")
	}
}
