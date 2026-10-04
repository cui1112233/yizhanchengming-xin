package metadata

import "testing"

func TestResolveGenderFrom121CategorySemantics(t *testing.T) {
	tests := []struct {
		category string
		want     string
	}{
		{category: "男生生活", want: "男频"},
		{category: "女生言情", want: "女频"},
		{category: "男性向都市", want: "男频"},
		{category: "女性向悬疑", want: "女频"},
		{category: "都市生活", want: ""},
		{category: "男女双强", want: ""},
	}

	for _, tc := range tests {
		got, source := ResolveGender("", tc.category, "8", "")
		if got != tc.want {
			t.Fatalf("ResolveGender(category=%q) = %q, want %q", tc.category, got, tc.want)
		}
		if tc.want != "" && source != SourceCategory {
			t.Fatalf("source = %q, want %q", source, SourceCategory)
		}
		if tc.want == "" && source != "" {
			t.Fatalf("ambiguous category source = %q, want empty", source)
		}
	}
}
