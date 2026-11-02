package novelfetchworkshop

import "testing"

func TestParseGenderAIResultKeepsValidClassificationAtLowConfidence(t *testing.T) {
	good := parseGenderAIResult(`{"gender":"女频","confidence":0.91,"reason":"主线围绕女主婚恋关系"}`)
	if good.Gender != "女频" || good.Confidence != 0.91 || good.Reason == "" {
		t.Fatalf("good=%+v", good)
	}
	low := parseGenderAIResult(`{"gender":"女频","confidence":0.62,"reason":"线索不充分"}`)
	if low.Gender != "女频" || low.Confidence != 0.62 {
		t.Fatalf("valid low-confidence classification was discarded: %+v", low)
	}
	for _, response := range []string{`女频`} {
		got := parseGenderAIResult(response)
		if got.Gender != "女频" || got.Confidence <= 0 {
			t.Errorf("plain gender response %q should be accepted: %+v", response, got)
		}
	}
	withoutScore := parseGenderAIResult(`{"gender":"女频","reason":"女主情感线为主"}`)
	if withoutScore.Gender != "女频" {
		t.Fatalf("valid label without score was discarded: %+v", withoutScore)
	}
	withoutReason := parseGenderAIResult(`{"gender":"女频"}`)
	if withoutReason.Gender != "女频" || withoutReason.Confidence <= 0 {
		t.Fatalf("valid label without reason was discarded: %+v", withoutReason)
	}
}

func TestDecideGenderManualOverridesBookInfoAndAI(t *testing.T) {
	got := DecideGender("女频", BookInfo{Category: "男频", Genre: "男生"}, "男频")
	if got.Gender != "女频" || got.Source != GenderManual {
		t.Fatalf("manual decision = %#v, want 女频/manual", got)
	}
	if got.Reason == "" {
		t.Fatal("manual decision reason is empty")
	}
}

func TestDecideGenderUsesExplicitCategory(t *testing.T) {
	got := DecideGender("", BookInfo{Category: "男频"}, "女频")
	if got.Gender != "男频" || got.Source != GenderBookInfo {
		t.Fatalf("category decision = %#v, want 男频/bookinfo", got)
	}
}

func TestDecideGenderUsesExplicitGenre(t *testing.T) {
	got := DecideGender("", BookInfo{Genre: "女性向"}, "男频")
	if got.Gender != "女频" || got.Source != GenderBookInfo {
		t.Fatalf("genre decision = %#v, want 女频/bookinfo", got)
	}
}

func TestDecideGenderConflictUsesValidAI(t *testing.T) {
	got := DecideGender("", BookInfo{Category: "男频", Genre: "女频"}, "女频")
	if got.Gender != "女频" || got.Source != GenderAI {
		t.Fatalf("AI conflict decision = %#v, want 女频/ai", got)
	}
	if got.Reason == "" {
		t.Fatal("AI decision reason is empty")
	}
}

func TestDecideGenderBookInfoConflictWithoutAIIsUnknown(t *testing.T) {
	got := DecideGender("", BookInfo{Category: "男频女频"}, "")
	if got.Gender != "待确认" || got.Source != GenderUnknown {
		t.Fatalf("bookinfo conflict decision = %#v, want 待确认/unknown", got)
	}
}

func TestDecideGenderBookInfoConflictWithConflictingAIIsUnknown(t *testing.T) {
	got := DecideGender("", BookInfo{Genre: "男女向"}, "男频女频")
	if got.Gender != "待确认" || got.Source != GenderUnknown {
		t.Fatalf("conflicting AI decision = %#v, want 待确认/unknown", got)
	}
}

func TestDecideGenderAmbiguousManualDoesNotGuess(t *testing.T) {
	got := DecideGender("男女向", BookInfo{}, "")
	if got.Gender != "待确认" || got.Source != GenderUnknown {
		t.Fatalf("ambiguous manual decision = %#v, want 待确认/unknown", got)
	}
}
