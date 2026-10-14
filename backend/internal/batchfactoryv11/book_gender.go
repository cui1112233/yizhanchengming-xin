package batchfactoryv11

import (
	"fmt"
	"strings"
)

// bookGenderForPrompt reads the already-resolved novel classification. The
// fetched novel metadata is the authority here; an ambiguous category/genre
// pair is deliberately ignored instead of making the model guess.
func bookGenderForPrompt(book Book) string {
	metadata := book.SourceMetadata
	if len(metadata) == 0 {
		return ""
	}
	for _, key := range []string{"gender", "bookGender", "book_gender"} {
		if gender := normalizeBookGender(metadata[key]); gender != "" {
			return gender
		}
	}
	if classification, ok := metadata["classification"].(map[string]any); ok {
		for _, key := range []string{"gender", "bookGender", "book_gender"} {
			if gender := normalizeBookGender(classification[key]); gender != "" {
				return gender
			}
		}
	}
	category := normalizeBookGender(metadata["category"])
	genre := normalizeBookGender(metadata["genre"])
	if category != "" && genre != "" && category != genre {
		return ""
	}
	if category != "" {
		return category
	}
	return genre
}

func normalizeBookGender(value any) string {
	text := strings.ToLower(strings.TrimSpace(fmt.Sprint(value)))
	if text == "" || text == "<nil>" {
		return ""
	}
	male := strings.Contains(text, "男")
	female := strings.Contains(text, "女")
	if male == female {
		return ""
	}
	if male {
		return "男频"
	}
	return "女频"
}

func isFirstPersonCharacter(name string) bool {
	name = strings.TrimSpace(name)
	return name == "我" || strings.Contains(name, "（我）") || strings.Contains(name, "(我)")
}

func containsAnyGenderMarker(value, marker string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	marker = strings.ToLower(strings.TrimSpace(marker))
	if value == "" || marker == "" {
		return false
	}
	for _, token := range []string{"男性", "女性", "男人", "女人", "男孩", "女孩", "少年", "少女", "男子", "女子", "男生", "女生", "丈夫", "妻子", "父亲", "母亲", "公子", "小姐", "先生", "夫人"} {
		if strings.Contains(value, token) && strings.Contains(token, marker) {
			return true
		}
	}
	return false
}

func hasExplicitCharacterGender(value string) bool {
	return containsAnyGenderMarker(value, "男") || containsAnyGenderMarker(value, "女")
}

// applyBookGenderFallback is a business rule, not an AI instruction. It only
// enriches an ambiguous first-person character; explicit facts always win.
func applyBookGenderFallback(book Book, assets DirectorAssets) DirectorAssets {
	gender := bookGenderForPrompt(book)
	if gender == "" {
		return assets
	}
	assets.Characters = append([]NamedPrompt(nil), assets.Characters...)
	for index, character := range assets.Characters {
		if !isFirstPersonCharacter(character.Name) && !isFirstPersonCharacter(character.FormalName) {
			continue
		}
		combined := strings.Join([]string{character.Facts, character.Appearance, character.Relationships, character.Prompt}, "\n")
		if hasExplicitCharacterGender(combined) {
			continue
		}
		marker := "男性"
		if gender == "女频" {
			marker = "女性"
		}
		character.Facts = appendGenderFallback(character.Facts, marker)
		character.Appearance = appendGenderFallback(character.Appearance, marker)
		character.Prompt = appendGenderFallback(character.Prompt, marker)
		assets.Characters[index] = character
	}
	return assets
}

func appendGenderFallback(value, marker string) string {
	value = strings.TrimSpace(value)
	addition := "性别锚点：" + marker + "（依据本书男女频分类补判，仅在原文未明确时使用）。"
	if value == "" {
		return addition
	}
	return value + " " + addition
}

func firstPersonGenderRule(book Book) string {
	gender := bookGenderForPrompt(book)
	if gender == "" {
		return ""
	}
	marker := "男性"
	if gender == "女频" {
		marker = "女性"
	}
	return fmt.Sprintf("【第一人称角色性别规则】本书男女频已确定为%s。若人物名为“我”或“（我）”且原文/资产没有明确性别，按%s输出；原文明确性别时以原文事实优先。", gender, marker)
}
