package novel

import "strings"

type Gender string

const (
	GenderUnknown Gender = ""
	GenderMale    Gender = "male"
	GenderFemale  Gender = "female"
)

type GenderSource string

const (
	GenderSourceManual      GenderSource = "manual"
	GenderSource121Category GenderSource = "121_category"
	GenderSource121Genre    GenderSource = "121_genre"
	GenderSourceAI          GenderSource = "ai"
	GenderSourceUnresolved  GenderSource = "unresolved"
)

type GenderInput struct {
	Manual           Gender
	Category         string
	Genre            string
	VerifiedGenreMap map[string]Gender
}

type GenderResult struct {
	Gender Gender
	Source GenderSource
}

var maleCategoryMarkers = []string{"男生", "男频", "男性", "男向"}
var femaleCategoryMarkers = []string{"女生", "女频", "女性", "女向"}

func ResolveGender(input GenderInput) GenderResult {
	if input.Manual == GenderMale || input.Manual == GenderFemale {
		return GenderResult{Gender: input.Manual, Source: GenderSourceManual}
	}

	category := strings.TrimSpace(input.Category)
	male := containsAny(category, maleCategoryMarkers)
	female := containsAny(category, femaleCategoryMarkers)
	if male != female {
		if male {
			return GenderResult{Gender: GenderMale, Source: GenderSource121Category}
		}
		return GenderResult{Gender: GenderFemale, Source: GenderSource121Category}
	}

	genre := strings.TrimSpace(input.Genre)
	if mapped, ok := input.VerifiedGenreMap[genre]; ok && (mapped == GenderMale || mapped == GenderFemale) {
		return GenderResult{Gender: mapped, Source: GenderSource121Genre}
	}

	return GenderResult{Gender: GenderUnknown, Source: GenderSourceUnresolved}
}

func containsAny(value string, markers []string) bool {
	for _, marker := range markers {
		if strings.Contains(value, marker) {
			return true
		}
	}
	return false
}
