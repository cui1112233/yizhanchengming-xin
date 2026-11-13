package batchfactoryv11

import "testing"

func TestComicBootstrapCanReuseOnlyWhenSourceAndPagesAreComplete(t *testing.T) {
	book := Book{
		SourceText:       "正文",
		AssetRecords:     []BookAsset{{Source: "comic", Prompt: "角色提示"}},
		DirectorRevision: &DirectorRevision{SourceDigest: sourceDigest("正文")},
		Videos: []Video{{
			ComicPrompt:      "分镜提示词",
			ComicImagePrompt: "出图提示词",
			ComicSourceSegment: &ComicSourceSegment{
				SourceText:   "正文",
				SourceDigest: sourceDigest("正文"),
			},
		}},
	}
	if !comicBootstrapCanReuse(book, true) {
		t.Fatal("complete comic result should be reusable")
	}

	book.Videos[0].ComicSourceSegment.SourceDigest = sourceDigest("新正文")
	if comicBootstrapCanReuse(book, true) {
		t.Fatal("stale comic result must not be reused")
	}
}

func TestComicAssetStageCanReuseExistingComicAssets(t *testing.T) {
	book := Book{AssetRecords: []BookAsset{{Source: "comic", Prompt: "角色提示"}}}
	if !comicAssetStageCanReuse(book) {
		t.Fatal("existing comic asset should satisfy the missing-only asset stage")
	}

	book.AssetRecords[0].Source = "manual"
	if comicAssetStageCanReuse(book) {
		t.Fatal("manual-only assets must not satisfy the comic asset stage")
	}
}

func TestComicAssetIdentityIssueDetectsGenericCharacterWithoutFormalName(t *testing.T) {
	book := Book{AssetRecords: []BookAsset{
		{Kind: "character", Name: "我", Prompt: "年轻女性，黑色长发", Source: "comic"},
		{Kind: "character", Name: "夫人", Prompt: "端庄女性，深色长裙", Source: "comic"},
	}}
	if issue := ComicAssetIdentityIssue(book); issue == "" {
		t.Fatal("generic character labels without formal names must trigger automatic identity recovery")
	}
}

func TestComicAssetIdentityIssueAllowsLifeStageVariantsOfSameFormalName(t *testing.T) {
	book := Book{AssetRecords: []BookAsset{
		{Kind: "character", Name: "林岁安少年", FormalName: "林岁安", Prompt: "十五岁少年，清瘦", Source: "comic"},
		{Kind: "character", Name: "林岁安幼儿", FormalName: "林岁安", Prompt: "五岁幼童，圆脸", Source: "comic"},
	}}
	if issue := ComicAssetIdentityIssue(book); issue != "" {
		t.Fatalf("life-stage variants with distinct asset names must not block production, got issue %q", issue)
	}
}

func TestComicAssetIdentityIssueStillRejectsDuplicatedAssetNames(t *testing.T) {
	book := Book{AssetRecords: []BookAsset{
		{Kind: "character", Name: "林岁安", FormalName: "林岁安", Prompt: "十五岁少年", Source: "comic"},
		{Kind: "character", Name: "林岁安", FormalName: "林岁安", Prompt: "五岁幼童", Source: "comic"},
	}}
	if issue := ComicAssetIdentityIssue(book); issue == "" {
		t.Fatal("two assets with the same asset name must still be treated as a duplicated identity")
	}
}

func TestComicAssetIdentityIssueAllowsDistinctFormalNames(t *testing.T) {
	book := Book{AssetRecords: []BookAsset{
		{Kind: "character", Name: "我", FormalName: "沈知微", Prompt: "年轻女性，黑色长发", Source: "comic"},
		{Kind: "character", Name: "我", FormalName: "陆沉", Prompt: "成年男性，短黑发", Source: "comic"},
	}}
	if issue := ComicAssetIdentityIssue(book); issue != "" {
		t.Fatalf("formal names should disambiguate characters, got issue %q", issue)
	}
}
