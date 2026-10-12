package httpapi

import (
	"testing"

	"qiantie/backend/internal/novelfetchworkshop"
)

func TestSubmissionGenderStyleFallsBackToDocumentWithoutInventingStyle(t *testing.T) {
	doc := novelfetchworkshop.Document{
		Gender:   "女频",
		BookInfo: novelfetchworkshop.BookInfo{Genre: "现代"},
		Meta:     map[string]any{},
	}

	gender, style := submissionGenderStyle(doc)
	if gender != "女" {
		t.Fatalf("gender = %q, want 女", gender)
	}
	if style != "" {
		t.Fatalf("style = %q, want empty until classification exists", style)
	}
}

func TestSubmissionGenderStylePrefersSavedMetadata(t *testing.T) {
	doc := novelfetchworkshop.Document{
		Gender: "男频",
		Meta: map[string]any{
			"gender": "女频",
			"style":  "现代悬疑",
		},
	}

	gender, style := submissionGenderStyle(doc)
	if gender != "女" || style != "现代悬疑" {
		t.Fatalf("metadata was not preferred: gender=%q style=%q", gender, style)
	}
}

func TestSyncNovelFetchClassificationMetadataPersistsResolvedGender(t *testing.T) {
	doc := novelfetchworkshop.Document{Gender: "女频", GenderSource: novelfetchworkshop.GenderAI, GenderReason: "AI 判断", Meta: map[string]any{}}
	syncNovelFetchClassificationMetadata(&doc)
	if got := doc.Meta["gender"]; got != "女频" {
		t.Fatalf("meta gender = %#v, want 女频", got)
	}
	if got := doc.Meta["genderSource"]; got != "ai" {
		t.Fatalf("meta genderSource = %#v, want ai", got)
	}
}
