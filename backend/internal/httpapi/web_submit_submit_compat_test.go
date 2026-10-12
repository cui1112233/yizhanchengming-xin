package httpapi

import (
	"testing"

	"qiantie/backend/internal/novelfetchworkshop"
)

func TestSubmissionGenderStyleFallsBackToDocumentAndDefaultStyle(t *testing.T) {
	doc := novelfetchworkshop.Document{
		Gender:   "女频",
		BookInfo: novelfetchworkshop.BookInfo{Genre: "现代"},
		Meta:     map[string]any{},
	}

	gender, style := submissionGenderStyle(doc)
	if gender != "女" {
		t.Fatalf("gender = %q, want 女", gender)
	}
	if style != "现代通用" {
		t.Fatalf("style = %q, want 现代通用", style)
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
