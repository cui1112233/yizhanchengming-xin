package batchfactoryv11

import "testing"

func TestParseManualBookListUsesPresetAndRetainsNovelFetchMetadata(t *testing.T) {
	books, err := ParseManualBookList(ManualIntakeInput{
		PlatformID:     "15",
		PlatformName:   "番茄小说",
		ParseMode:      "smart",
		ColumnPresetID: "sample_input",
		InputText:      "2080989285751305136\t重生书\t推荐理由\t女频\t重生,爽文\tS",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(books) != 1 || books[0].BookID != "2080989285751305136" || books[0].Title != "重生书" {
		t.Fatalf("books=%+v", books)
	}
	metadata := books[0].SourceMetadata
	if metadata["platformName"] != "番茄小说" || metadata["gender"] != "女频" || metadata["rating"] != "S" || metadata["tags"] != "重生,爽文" || metadata["reason"] != "推荐理由" {
		t.Fatalf("metadata=%#v", metadata)
	}
	if metadata["sourceMode"] != "manual_original" || metadata["sourceLine"] == "" {
		t.Fatalf("source metadata=%#v", metadata)
	}
}

func TestParseManualBookListRetainsFullMetadataPreset(t *testing.T) {
	books, err := ParseManualBookList(ManualIntakeInput{
		PlatformID:     "15",
		PlatformName:   "知乎付费",
		ParseMode:      "smart",
		ColumnPresetID: "full_metadata",
		InputText:      "2080989285751305136\t重生书\t女频\t现代虐文\t重生,爽文\t女主逆袭\tS+",
	})
	if err != nil {
		t.Fatal(err)
	}
	metadata := books[0].SourceMetadata
	if metadata["gender"] != "女频" || metadata["style"] != "现代虐文" || metadata["tags"] != "重生,爽文" || metadata["reason"] != "女主逆袭" || metadata["rating"] != "S+" {
		t.Fatalf("metadata=%#v", metadata)
	}
}

func TestParseManualBookListPersists121MetadataWithoutOverwritingManualGender(t *testing.T) {
	books, err := ParseManualBookList(ManualIntakeInput{
		PlatformID: "2", PlatformName: "番茄付费", ParseMode: "smart", ColumnPresetID: "full_metadata",
		InputText: "7673480334440139800\t手工女频书\t女频",
		SourceMetadataByBookID: map[string]map[string]any{
			"7673480334440139800": {"category": "男频-都市", "genre": "都市", "gender": "男频", "genderSource": "121_category"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	metadata := books[0].SourceMetadata
	if metadata["category"] != "男频-都市" || metadata["genre"] != "都市" {
		t.Fatalf("121 metadata was not persisted: %#v", metadata)
	}
	if metadata["gender"] != "女频" || metadata["genderSource"] != "manual" {
		t.Fatalf("manual gender must win over 121 metadata: %#v", metadata)
	}
}

func TestParseManualBookListSeparatesBookIDAndTitleWithSingleSpace(t *testing.T) {
	books, err := ParseManualBookList(ManualIntakeInput{
		PlatformID:     "15",
		PlatformName:   "知乎付费",
		ParseMode:      "smart",
		ColumnPresetID: "full_metadata",
		InputText:      "2085148147785918287 阿芙",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(books) != 1 || books[0].BookID != "2085148147785918287" || books[0].Title != "阿芙" {
		t.Fatalf("books=%+v", books)
	}
}

func TestParseManualBookListAcceptsShortQimaoBookIDs(t *testing.T) {
	books, err := ParseManualBookList(ManualIntakeInput{
		PlatformID:     "3",
		PlatformName:   "七猫付费",
		ParseMode:      "smart",
		ColumnPresetID: "full_metadata",
		InputText:      "567168 晚风惊扰旧梦\n705142 她指指方向，我未对路",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(books) != 2 || books[0].BookID != "567168" || books[0].Title != "晚风惊扰旧梦" || books[1].BookID != "705142" || books[1].Title != "她指指方向，我未对路" {
		t.Fatalf("books=%+v", books)
	}
	if books[0].Platform != "3" || books[0].SourceMetadata["platformName"] != "七猫付费" {
		t.Fatalf("book=%+v", books[0])
	}
}

func TestParseManualBookListAcceptsHorizontallyPastedBookIDs(t *testing.T) {
	books, err := ParseManualBookList(ManualIntakeInput{
		PlatformID:     "2",
		PlatformName:   "番茄付费",
		ParseMode:      "smart",
		ColumnPresetID: "full_metadata",
		InputText:      "7707226373706088457\t7705762208613335078 7706177708078759962",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(books) != 3 {
		t.Fatalf("horizontally pasted IDs were not split into books: %+v", books)
	}
	for index, want := range []string{"7707226373706088457", "7705762208613335078", "7706177708078759962"} {
		if books[index].BookID != want {
			t.Fatalf("books[%d].BookID = %q, want %q", index, books[index].BookID, want)
		}
	}
}

func TestParseManualBookListSeparatesPastedBookIDAndTitleAndKeepsFetchedOriginal(t *testing.T) {
	const bookID = "7655922169695718462"
	const title = "未婚妻拿我当踏板，我加入剧组她却哭了"
	books, err := ParseManualBookList(ManualIntakeInput{
		PlatformID:         "2",
		PlatformName:       "番茄付费",
		ParseMode:          "smart",
		ColumnPresetID:     "full_metadata",
		InputText:          bookID + title,
		SourceTextByBookID: map[string]string{bookID: "已抓取的小说正文"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(books) != 1 || books[0].BookID != bookID || books[0].Title != title {
		t.Fatalf("books=%+v", books)
	}
	if books[0].SourceText != "已抓取的小说正文" {
		t.Fatalf("source text was not matched: %+v", books[0])
	}
}

func TestParseManualBookListUnderstandsHeaderAndDeduplicatesBookID(t *testing.T) {
	books, err := ParseManualBookList(ManualIntakeInput{
		PlatformID: "zhihu-paid",
		ParseMode:  "header",
		InputText:  "书籍ID\t书名\t男女频\t类型\n100000000001\t书A\t男频\t都市\n100000000001\t书A新版\t男频\t都市",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(books) != 1 || books[0].BookID != "100000000001" || books[0].Title != "书A新版" {
		t.Fatalf("books=%+v", books)
	}
	if books[0].SourceMetadata["style"] != "都市" || books[0].Platform != "zhihu-paid" {
		t.Fatalf("book=%+v", books[0])
	}
}
