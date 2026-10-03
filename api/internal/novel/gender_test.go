package novel

import "testing"

func TestResolveGenderManualWins(t *testing.T) {
	result := ResolveGender(GenderInput{
		Manual:   GenderFemale,
		Category: "男生生活",
	})
	if result.Gender != GenderFemale || result.Source != GenderSourceManual {
		t.Fatalf("expected manual female to win, got %#v", result)
	}
}

func TestResolveGenderFromMaleCategory(t *testing.T) {
	result := ResolveGender(GenderInput{Category: "男生生活"})
	if result.Gender != GenderMale || result.Source != GenderSource121Category {
		t.Fatalf("expected male from category, got %#v", result)
	}
}

func TestResolveGenderFromFemaleCategory(t *testing.T) {
	result := ResolveGender(GenderInput{Category: "女生言情"})
	if result.Gender != GenderFemale || result.Source != GenderSource121Category {
		t.Fatalf("expected female from category, got %#v", result)
	}
}

func TestResolveGenderAmbiguousCategoryFallsThrough(t *testing.T) {
	result := ResolveGender(GenderInput{Category: "男生女生都爱看"})
	if result.Gender != GenderUnknown || result.Source != GenderSourceUnresolved {
		t.Fatalf("expected ambiguous category to remain unresolved, got %#v", result)
	}
}

func TestResolveGenderUsesOnlyProvidedVerifiedGenreMapping(t *testing.T) {
	result := ResolveGender(GenderInput{
		Genre: "verified-male-genre",
		VerifiedGenreMap: map[string]Gender{
			"verified-male-genre": GenderMale,
		},
	})
	if result.Gender != GenderMale || result.Source != GenderSource121Genre {
		t.Fatalf("expected verified genre fallback, got %#v", result)
	}
}

func TestResolveGenderUnknownGenreDoesNotGuess(t *testing.T) {
	result := ResolveGender(GenderInput{Genre: "8"})
	if result.Gender != GenderUnknown || result.Source != GenderSourceUnresolved {
		t.Fatalf("expected unknown genre to remain unresolved, got %#v", result)
	}
}
