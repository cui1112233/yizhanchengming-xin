package httpapi

import (
	"testing"

	"qiantie/backend/internal/batchfactoryv11"
)

func TestComicAudioCategoryMapsNovelStyleAndGender(t *testing.T) {
	tests := []struct {
		style  string
		gender string
		want   string
	}{
		{style: "现代悬疑", gender: "女频", want: "现代悬疑"},
		{style: "男频都市", gender: "男频", want: "男频都市"},
		{style: "古风通用", gender: "女频", want: "古风通用"},
		{style: "玄幻", gender: "男频", want: "古风通用"},
		{style: "历史", gender: "男频", want: "古风通用"},
		{style: "古风虐文", gender: "女频", want: "古风虐文"},
		{style: "古风甜文", gender: "女频", want: "古风甜文"},
		{style: "现代虐文", gender: "女频", want: "现代虐文"},
		{style: "年代虐文", gender: "女频", want: "现代虐文"},
		{style: "家庭伤感", gender: "女频", want: "现代虐文"},
		{style: "现代甜文", gender: "女频", want: "现代通用"},
		{style: "现代通用", gender: "女频", want: "现代通用"},
		{style: "年代甜文", gender: "女频", want: "现代通用"},
		{style: "年代通用", gender: "女频", want: "现代通用"},
		{style: "现代女主", gender: "女频", want: "现代通用"},
		{style: "家庭奇葩", gender: "女频", want: "现代通用"},
		{style: "职场打脸", gender: "男频", want: "男频都市"},
		{style: "职场打脸", gender: "女频", want: "现代通用"},
		{style: "未知", gender: "男频", want: "男频都市"},
		{style: "未知", gender: "女频", want: "现代通用"},
	}
	for _, tt := range tests {
		t.Run(tt.style+tt.gender, func(t *testing.T) {
			book := batchfactoryv11.Book{SourceMetadata: map[string]any{
				"style": tt.style, "gender": tt.gender, "classifyStatus": "classified",
			}}
			got, err := comicAudioCategoryName(book)
			if err != nil || got != tt.want {
				t.Fatalf("comicAudioCategoryName() = %q, %v; want %q", got, err, tt.want)
			}
		})
	}
}

func TestComicAudioCategoryRejectsMissingClassification(t *testing.T) {
	book := batchfactoryv11.Book{SourceMetadata: map[string]any{"style": "未知", "classifyStatus": "classified"}}
	if _, err := comicAudioCategoryName(book); err == nil {
		t.Fatal("comicAudioCategoryName() error = nil, want missing classification error")
	}
}
