package batchfactoryv11

import (
	"encoding/json"
	"strings"
	"testing"
)

func validDirectorJSON() string {
	return `{
  "characters": [{"name":"林晚","prompt":"18岁中国女性，黑色长发。"}],
  "scenes": [{"name":"林家客厅","prompt":"现代中式客厅。"}],
  "props": [{"name":"玻璃杯","prompt":"透明厚底玻璃杯。"}],
  "storyboard": [
    {
      "id": 1,
      "scene_id": 1,
      "duration_sec": 9,
      "characters": ["林晚"],
      "props": ["玻璃杯"],
      "scene": "林家客厅",
      "prefix_key": "modern_conflict",
      "shots": [
        {"start_sec":0,"end_sec":3,"shot_type":"中景","camera":"缓慢推轨","description":"林晚进入客厅"},
        {"start_sec":3,"end_sec":9,"shot_type":"特写","camera":"固定","description":"她握紧玻璃杯"}
      ],
      "video_desc": "林晚进入客厅并握紧玻璃杯。"
    }
  ]
}`
}

func twoDirectorVideosJSON(t *testing.T) string {
	t.Helper()
	var root map[string]any
	if err := json.Unmarshal([]byte(validDirectorJSON()), &root); err != nil {
		t.Fatalf("unmarshal validDirectorJSON: %v", err)
	}
	storyboard, ok := root["storyboard"].([]any)
	if !ok || len(storyboard) != 1 {
		t.Fatalf("unexpected storyboard fixture: %#v", root["storyboard"])
	}
	first, ok := storyboard[0].(map[string]any)
	if !ok {
		t.Fatalf("unexpected first storyboard: %#v", storyboard[0])
	}
	second := map[string]any{}
	for key, value := range first {
		second[key] = value
	}
	second["id"] = 2
	root["storyboard"] = append(storyboard, second)
	encoded, err := json.Marshal(root)
	if err != nil {
		t.Fatalf("marshal two-video fixture: %v", err)
	}
	return string(encoded)
}

func TestDirectorParseFencedJSON(t *testing.T) {
	raw, err := ParseDirectorJSON("```json\n" + validDirectorJSON() + "\n```")
	if err != nil {
		t.Fatalf("ParseDirectorJSON returned error: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("parsed JSON is invalid: %v", err)
	}
	if _, ok := decoded["storyboard"]; !ok {
		t.Fatalf("storyboard missing from parsed JSON")
	}
}

func TestNormalizeAssetExtractionOutputKeepsCharacterStructure(t *testing.T) {
	assets, err := NormalizeAssetExtractionOutput(json.RawMessage(`{
		"characters":[{"name":"赵姨娘","formal_name":"赵姨娘","facts":"侯府侍妾，做侍妾二十年","appearance":"瘦削瓜子脸，黛青暗纹袄裙","relationships":"与侯门主母为主仆关系","prompt":"赵姨娘，侯府侍妾，瘦削瓜子脸，黛青暗纹袄裙"}],
		"scenes":[],"props":[]
	}`))
	if err != nil {
		t.Fatalf("NormalizeAssetExtractionOutput returned error: %v", err)
	}
	if len(assets.Characters) != 1 {
		t.Fatalf("characters=%+v", assets.Characters)
	}
	character := assets.Characters[0]
	if character.FormalName != "赵姨娘" || character.Facts == "" || character.Appearance == "" || character.Relationships == "" {
		t.Fatalf("structured character fields were lost: %+v", character)
	}
}

func TestSDDirectorTextPreservesStructuredAssetReferences(t *testing.T) {
	result, err := parseSDDirectorText(`===VIDEO 01===
时长：10秒
资产引用：人物=妈妈、我；场景=家中餐厅；道具=孕检单
镜头画面：妈妈把孕检单放到家中餐厅的餐桌上，我站在一旁。`, 15)
	if err != nil {
		t.Fatalf("parseSDDirectorText returned error: %v", err)
	}
	if len(result.Storyboard) != 1 {
		t.Fatalf("unexpected storyboard: %#v", result.Storyboard)
	}
	video := result.Storyboard[0]
	if got, want := video.Characters, []string{"妈妈", "我"}; !equalStrings(got, want) {
		t.Fatalf("characters = %#v, want %#v", got, want)
	}
	if video.Scene != "家中餐厅" {
		t.Fatalf("scene = %q, want %q", video.Scene, "家中餐厅")
	}
	if got, want := video.Props, []string{"孕检单"}; !equalStrings(got, want) {
		t.Fatalf("props = %#v, want %#v", got, want)
	}
	if strings.Contains(video.FinalPrompt, "资产引用：") {
		t.Fatalf("internal asset-reference header leaked into prompt: %q", video.FinalPrompt)
	}
}

func equalStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

func TestDirectorNormalizeValidOutput(t *testing.T) {
	result, err := NormalizeDirectorOutput(json.RawMessage(validDirectorJSON()), DirectorSettings{
		MaxVideoDuration: 15,
		AspectRatio:      "9:16",
		AllowedPrefixKeys: []string{
			"modern_conflict",
		},
	})
	if err != nil {
		t.Fatalf("NormalizeDirectorOutput returned error: %v", err)
	}
	if len(result.Characters) != 1 || result.Characters[0].Name != "林晚" {
		t.Fatalf("unexpected characters: %#v", result.Characters)
	}
	if len(result.Storyboard) != 1 || result.Storyboard[0].DurationSec != 9 {
		t.Fatalf("unexpected storyboard: %#v", result.Storyboard)
	}
	if !result.SourceCoverage.SourceComplete || result.SourceCoverage.HasRemainingSource {
		t.Fatalf("unexpected default source coverage: %#v", result.SourceCoverage)
	}
}

func TestDirectorRequiresAndPreservesH3CanonicalMetadata(t *testing.T) {
	var root map[string]any
	if err := json.Unmarshal([]byte(validDirectorJSON()), &root); err != nil {
		t.Fatal(err)
	}
	storyboard := root["storyboard"].([]any)
	video := storyboard[0].(map[string]any)
	video["scene_memory"] = "林晚站在客厅门口，右手握着玻璃杯，门仍半开。"
	shots := video["shots"].([]any)
	shots[0].(map[string]any)["visual_context"] = "林家客厅，傍晚的窗边冷光照在茶几上。"
	shots[0].(map[string]any)["lighting"] = "侧向冷光勾勒林晚面部，暗部保留木材纹理。"
	shots[0].(map[string]any)["rhythm"] = "克制铺陈"
	shots[0].(map[string]any)["audio"] = "门锁轻响，林晚（低声）：“我回来了。”"
	shots[1].(map[string]any)["visual_context"] = "林家客厅，玻璃杯仍在茶几边缘。"
	shots[1].(map[string]any)["lighting"] = "窗外冷光落在杯沿，反射短暂闪动。"
	shots[1].(map[string]any)["rhythm"] = "骤然收紧"
	shots[1].(map[string]any)["audio"] = "无"
	raw, err := json.Marshal(root)
	if err != nil {
		t.Fatal(err)
	}
	result, err := NormalizeDirectorOutput(raw, DirectorSettings{
		MaxVideoDuration:  15,
		AspectRatio:       "9:16",
		AllowedPrefixKeys: []string{"modern_conflict"},
		RequireH3Metadata: true,
	})
	if err != nil {
		t.Fatalf("H3 metadata should normalize: %v", err)
	}
	videoResult := result.Storyboard[0]
	if videoResult.SceneMemory == "" || videoResult.Shots[0].Rhythm != "克制铺陈" || videoResult.Shots[0].Audio == "" {
		t.Fatalf("H3 metadata was not persisted: %#v", videoResult)
	}

	delete(video, "scene_memory")
	missingMemory, err := json.Marshal(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NormalizeDirectorOutput(missingMemory, DirectorSettings{MaxVideoDuration: 15, AspectRatio: "9:16", AllowedPrefixKeys: []string{"modern_conflict"}, RequireH3Metadata: true}); err == nil {
		t.Fatal("H3 metadata validation must reject missing scene_memory")
	}
}

func TestStoryboardVideoPromptUsesDirectorShotTimelineInsteadOfSummary(t *testing.T) {
	result, err := NormalizeDirectorOutput(json.RawMessage(validDirectorJSON()), DirectorSettings{
		MaxVideoDuration:  15,
		AspectRatio:       "9:16",
		AllowedPrefixKeys: []string{"modern_conflict"},
	})
	if err != nil {
		t.Fatal(err)
	}
	prompt := storyboardVideoPrompt(result.Storyboard[0])
	for _, expected := range []string{
		"镜头画面：",
		"00:00-00:03 | 中景｜缓慢推轨 | 林晚进入客厅",
		"00:03-00:09 | 特写｜固定 | 她握紧玻璃杯",
	} {
		if !strings.Contains(prompt, expected) {
			t.Fatalf("storyboard prompt is missing %q:\n%s", expected, prompt)
		}
	}
	if strings.Contains(prompt, result.Storyboard[0].VideoDesc) {
		t.Fatalf("summary video_desc must not replace the storyboard prompt:\n%s", prompt)
	}
}

func TestDirectorRequiresStoryboardTotalToMatchAudioPlan(t *testing.T) {
	settings := DirectorSettings{
		MaxVideoDuration:   10,
		AudioTargetSeconds: 9,
		AspectRatio:        "9:16",
		AllowedPrefixKeys:  []string{"modern_conflict"},
	}
	if _, err := NormalizeDirectorOutput(json.RawMessage(validDirectorJSON()), settings); err != nil {
		t.Fatalf("matching audio plan rejected: %v", err)
	}
	settings.AudioTargetSeconds = 10
	if _, err := NormalizeDirectorOutput(json.RawMessage(validDirectorJSON()), settings); err == nil || !strings.Contains(err.Error(), "总时长必须严格等于配音规划时长") {
		t.Fatalf("expected exact audio-total rejection, got %v", err)
	}
}

func TestDirectorFixedSingleVideoRules(t *testing.T) {
	value, err := NormalizeDirectorOutput(json.RawMessage(twoDirectorVideosJSON(t)), DirectorSettings{
		MaxVideoDuration:  15,
		FixedSingleVideo:  true,
		AspectRatio:       "9:16",
		AllowedPrefixKeys: []string{"modern_conflict"},
	})
	if err != nil || len(value.Storyboard) != 2 {
		t.Fatalf("fixed opening must preserve the full storyboard for review: value=%+v err=%v", value, err)
	}
}

func TestDirectorRejectsFractionalDuration(t *testing.T) {
	fractional := strings.Replace(validDirectorJSON(), `"duration_sec": 9`, `"duration_sec": 9.5`, 1)
	_, err := NormalizeDirectorOutput(json.RawMessage(fractional), DirectorSettings{
		MaxVideoDuration:  15,
		AspectRatio:       "9:16",
		AllowedPrefixKeys: []string{"modern_conflict"},
	})
	if err == nil || !strings.Contains(err.Error(), "必须是整数") {
		t.Fatalf("expected integer error, got %v", err)
	}
}

func TestDirectorRoundsProviderShotTimesToIntegerSeconds(t *testing.T) {
	raw := strings.Replace(validDirectorJSON(), `"start_sec":0,"end_sec":3`, `"start_sec":0.0,"end_sec":2.7`, 1)
	raw = strings.Replace(raw, `"start_sec":3,"end_sec":9`, `"start_sec":3.2,"end_sec":9.0`, 1)
	result, err := NormalizeDirectorOutput(json.RawMessage(raw), DirectorSettings{
		MaxVideoDuration: 15, AspectRatio: "9:16", AllowedPrefixKeys: []string{"modern_conflict"},
	})
	if err != nil {
		t.Fatalf("expected decimal provider times to normalize, got %v", err)
	}
	shots := result.Storyboard[0].Shots
	if shots[0].StartSec != 0 || shots[0].EndSec != 3 || shots[1].StartSec != 3 || shots[1].EndSec != 9 {
		t.Fatalf("unexpected normalized shots: %#v", shots)
	}
}

func TestDirectorRejectsDiscontinuousShots(t *testing.T) {
	broken := strings.Replace(validDirectorJSON(), `"start_sec":3,"end_sec":9`, `"start_sec":4,"end_sec":9`, 1)
	_, err := NormalizeDirectorOutput(json.RawMessage(broken), DirectorSettings{
		MaxVideoDuration:  15,
		AspectRatio:       "9:16",
		AllowedPrefixKeys: []string{"modern_conflict"},
	})
	if err == nil || !strings.Contains(err.Error(), "必须连续") {
		t.Fatalf("expected continuity error, got %v", err)
	}
}

func TestDirectorNormalizesProviderShotDescriptionAliases(t *testing.T) {
	for _, field := range []string{"shot_description", "画面描述"} {
		t.Run(field, func(t *testing.T) {
			output := strings.Replace(validDirectorJSON(), `"description":"林晚进入客厅"`, `"`+field+`":"林晚进入客厅"`, 1)
			result, err := NormalizeDirectorOutput(json.RawMessage(output), DirectorSettings{
				MaxVideoDuration:  15,
				AspectRatio:       "9:16",
				AllowedPrefixKeys: []string{"modern_conflict"},
			})
			if err != nil {
				t.Fatalf("NormalizeDirectorOutput returned error: %v", err)
			}
			if got := result.Storyboard[0].Shots[0].Description; got != "林晚进入客厅" {
				t.Fatalf("description=%q", got)
			}
		})
	}
}

func TestDirectorRejectsUnknownReferencesAndPrefix(t *testing.T) {
	unknownCharacter := strings.Replace(validDirectorJSON(), `"characters": ["林晚"]`, `"characters": ["陌生人"]`, 1)
	_, err := NormalizeDirectorOutput(json.RawMessage(unknownCharacter), DirectorSettings{
		MaxVideoDuration:  15,
		AspectRatio:       "9:16",
		AllowedPrefixKeys: []string{"modern_conflict"},
	})
	if err == nil || !strings.Contains(err.Error(), "未知人物") {
		t.Fatalf("expected unknown character error, got %v", err)
	}

	unknownPrefix := strings.Replace(validDirectorJSON(), `"prefix_key": "modern_conflict"`, `"prefix_key": "unknown"`, 1)
	_, err = NormalizeDirectorOutput(json.RawMessage(unknownPrefix), DirectorSettings{
		MaxVideoDuration:  15,
		AspectRatio:       "9:16",
		AllowedPrefixKeys: []string{"modern_conflict"},
	})
	if err == nil || !strings.Contains(err.Error(), "未知前缀类型") {
		t.Fatalf("expected unknown prefix error, got %v", err)
	}
}

func TestDirectorFixedSingleDefaultSourceCoverage(t *testing.T) {
	fixed := strings.Replace(validDirectorJSON(), `"duration_sec": 9`, `"duration_sec": 15`, 1)
	fixed = strings.Replace(fixed, `"end_sec":9`, `"end_sec":15`, 1)
	result, err := NormalizeDirectorOutput(json.RawMessage(fixed), DirectorSettings{
		MaxVideoDuration:  15,
		FixedSingleVideo:  true,
		AspectRatio:       "9:16",
		AllowedPrefixKeys: []string{"modern_conflict"},
	})
	if err != nil {
		t.Fatalf("NormalizeDirectorOutput returned error: %v", err)
	}
	if !result.SourceCoverage.SourceComplete || result.SourceCoverage.HasRemainingSource {
		t.Fatalf("fixed-single must still analyze the full current production text: %#v", result.SourceCoverage)
	}
}
