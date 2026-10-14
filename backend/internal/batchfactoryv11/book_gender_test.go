package batchfactoryv11

import (
	"strings"
	"testing"
)

func TestApplyBookGenderFallbackToAmbiguousFirstPersonCharacter(t *testing.T) {
	book := Book{SourceMetadata: map[string]any{"gender": "女频"}}
	assets := DirectorAssets{Characters: []NamedPrompt{{
		Name:       "妾室（我）",
		FormalName: "妾室（我）",
		Facts:      "身份：宅邸侍妾。",
		Appearance: "发髻简单，身着素色襦裙。",
		Prompt:     "古代宅邸侍妾，神态恭顺。",
	}}}

	got := applyBookGenderFallback(book, assets)
	character := got.Characters[0]
	if character.Facts == assets.Characters[0].Facts || character.Appearance == assets.Characters[0].Appearance || character.Prompt == assets.Characters[0].Prompt {
		t.Fatalf("ambiguous first-person character was not enriched: %#v", character)
	}
	for _, value := range []string{character.Facts, character.Appearance, character.Prompt} {
		if !containsAnyGenderMarker(value, "女") {
			t.Fatalf("female fallback marker missing from %q", value)
		}
	}
}

func TestApplyBookGenderFallbackDoesNotOverrideExplicitGender(t *testing.T) {
	book := Book{SourceMetadata: map[string]any{"category": "男频"}}
	assets := DirectorAssets{Characters: []NamedPrompt{{
		Name:       "我",
		Facts:      "性别：女性。",
		Appearance: "成年女性，长发。",
		Prompt:     "女性角色，现代服装。",
	}}}

	got := applyBookGenderFallback(book, assets)
	if got.Characters[0].Facts != assets.Characters[0].Facts || got.Characters[0].Prompt != assets.Characters[0].Prompt {
		t.Fatalf("explicit gender was overwritten: %#v", got.Characters[0])
	}
}

func TestBookGenderForPromptUsesResolvedBookClassification(t *testing.T) {
	if got := bookGenderForPrompt(Book{SourceMetadata: map[string]any{"classification": map[string]any{"gender": "男频"}}}); got != "男频" {
		t.Fatalf("nested classification gender=%q, want 男频", got)
	}
	if got := bookGenderForPrompt(Book{SourceMetadata: map[string]any{"category": "男频", "genre": "女频"}}); got != "" {
		t.Fatalf("conflicting category/genre gender=%q, want empty", got)
	}
}

func TestBuildAssetExtractionContractCarriesBookGenderFallback(t *testing.T) {
	contract, err := BuildAssetExtractionContract(Book{
		ID:             "book-gender",
		Title:          "女频测试书",
		SourceText:     "我走进屋里。",
		SourceMetadata: map[string]any{"gender": "女频"},
	}, DirectorSnapshot{})
	if err != nil {
		t.Fatal(err)
	}
	if !containsAnyGenderMarker(contract.SystemPrompt, "女") || !strings.Contains(contract.UserPrompt, `"book_gender": "女频"`) {
		t.Fatalf("book gender fallback missing from contract: system=%q user=%q", contract.SystemPrompt, contract.UserPrompt)
	}
}

func TestAssetExtractionContractSeparatesStaticAppearanceFromStoryActions(t *testing.T) {
	contract, err := BuildAssetExtractionContract(Book{ID: "book-static", Title: "静态资产", SourceText: "她走进厅堂。"}, DirectorSnapshot{})
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{"只写静态", "必须详细展开", "肤色与皮肤细节", "服装材质与颜色", "不得写镜头动作", "端坐", "抬手", "走动", "电影级景深", "神情深不可测"} {
		if !strings.Contains(contract.SystemPrompt, required) {
			t.Fatalf("asset contract missing static-only rule %q: %s", required, contract.SystemPrompt)
		}
	}
}
