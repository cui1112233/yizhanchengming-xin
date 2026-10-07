package novelpanel

import (
	"errors"
	"testing"
)

func TestParseForcedRosterPreservesBracketHintsAndDeduplicates(t *testing.T) {
	input := "沈清月（青年）, 顾川（少年，回忆）\n沈清月（青年）|林姨[中年]"
	first := ParseForcedRoster(input)
	second := ParseForcedRoster(input)
	if len(first) != 3 {
		t.Fatalf("expected 3 cards, got %d: %#v", len(first), first)
	}
	if first[0].BaseName != "沈清月" || first[0].StageHint != "青年" {
		t.Fatalf("unexpected first card: %#v", first[0])
	}
	if first[1].BaseName != "顾川" || first[1].StageHint != "少年，回忆" {
		t.Fatalf("bracket delimiter should stay inside one roster entry: %#v", first[1])
	}
	if first[2].BaseName != "林姨" || first[2].StageHint != "中年" {
		t.Fatalf("square bracket hint not parsed: %#v", first[2])
	}
	if first[0].ID != second[0].ID || first[1].ID != second[1].ID {
		t.Fatal("forced roster ids must be stable for the same ordered roster")
	}
}

func TestMergeForcedCharactersKeepsEditedCardFieldsButDropsUnlistedCharacters(t *testing.T) {
	forced := ParseForcedRoster("甲（青年）,乙")
	submitted := []CharacterCard{
		{ID: "legacy-a", SourceEntry: "甲（青年）", BaseName: "甲", Appearance: "黑色长发", CardNote: "主角", Gender: "女", AssetRefs: []string{"asset:character:1"}},
		{ID: "legacy-extra", BaseName: "丙", DisplayName: "丙", Appearance: "不应保留"},
	}
	got := MergeForcedCharacters(forced, submitted)
	if len(got) != 2 {
		t.Fatalf("forced roster should remain authoritative: %#v", got)
	}
	if got[0].Appearance != "黑色长发" || got[0].CardNote != "主角" || got[0].Gender != "女" || len(got[0].AssetRefs) != 1 {
		t.Fatalf("edited card fields were not preserved: %#v", got[0])
	}
}

func TestValidateRelationshipsRejectsUnknownOrSelfEndpoints(t *testing.T) {
	characters := []CharacterCard{{ID: "a"}, {ID: "b"}}
	if err := ValidateRelationships(characters, []CharacterRelationship{{FromID: "a", ToID: "b", Type: "朋友"}}); err != nil {
		t.Fatalf("valid relationship rejected: %v", err)
	}
	for _, relationships := range [][]CharacterRelationship{
		{{FromID: "a", ToID: "missing", Type: "朋友"}},
		{{FromID: "a", ToID: "a", Type: "自我"}},
	} {
		if err := ValidateRelationships(characters, relationships); !errors.Is(err, ErrInvalid) {
			t.Fatalf("expected ErrInvalid, got %v", err)
		}
	}
}
