package metadata

import "testing"

func TestResolveGenderPriority(t *testing.T) {
	tests := []struct {
		name     string
		manual   string
		category string
		genre    string
		ai       string
		want     string
		source   string
	}{
		{
			name: "manual wins over every lower priority source", manual: "男频", category: "女频", genre: "现代言情", ai: "女频", want: "男频", source: SourceManual,
		},
		{
			name: "121 category wins over genre and ai", category: "女频", genre: "都市", ai: "男频", want: "女频", source: SourceCategory,
		},
		{
			name: "verified genre mapping is used before ai", genre: "现代言情", ai: "男频", want: "女频", source: SourceGenre,
		},
		{
			name: "verified male genre mapping is supported", genre: "都市脑洞", ai: "女频", want: "男频", source: SourceGenre,
		},
		{
			name: "unknown genre falls back to ai", genre: "自定义新题材", ai: "女频", want: "女频", source: SourceAI,
		},
		{
			name: "category aliases normalize before ai", category: "女生", ai: "男频", want: "女频", source: SourceCategory,
		},
		{
			name: "invalid values do not become authoritative", manual: "未知", category: "其他", genre: "未收录题材", ai: "男频", want: "男频", source: SourceAI,
		},
		{
			name: "empty inputs stay unresolved", want: "", source: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, source := ResolveGender(tc.manual, tc.category, tc.genre, tc.ai)
			if got != tc.want || source != tc.source {
				t.Fatalf("ResolveGender() = (%q, %q), want (%q, %q)", got, source, tc.want, tc.source)
			}
		})
	}
}

func TestResolveGenderTrimsInputs(t *testing.T) {
	got, source := ResolveGender("  女频  ", "男频", "都市", "男频")
	if got != "女频" || source != SourceManual {
		t.Fatalf("ResolveGender() = (%q, %q)", got, source)
	}
}

func TestResolveStylePriority(t *testing.T) {
	tests := []struct {
		name     string
		manual   string
		provider string
		ai       string
		want     string
		source   string
	}{
		{name: "manual style wins", manual: "爽文", provider: "情感", ai: "悬疑", want: "爽文", source: SourceManual},
		{name: "provider style wins over ai", provider: "情感", ai: "悬疑", want: "情感", source: SourceProvider},
		{name: "ai is fallback", ai: "悬疑", want: "悬疑", source: SourceAI},
		{name: "empty remains empty", want: "", source: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, source := ResolveStyle(tc.manual, tc.provider, tc.ai)
			if got != tc.want || source != tc.source {
				t.Fatalf("ResolveStyle() = (%q, %q), want (%q, %q)", got, source, tc.want, tc.source)
			}
		})
	}
}
