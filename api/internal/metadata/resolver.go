package metadata

import "strings"

const (
	SourceManual   = "manual"
	SourceCategory = "category"
	SourceGenre    = "genre"
	SourceProvider = "provider"
	SourceAI       = "ai"
)

var verifiedGenreGender = map[string]string{
	"现代言情": "女频",
	"古代言情": "女频",
	"现言脑洞": "女频",
	"古言脑洞": "女频",
	"宫斗宅斗": "女频",
	"青春甜宠": "女频",
	"豪门总裁": "女频",
	"职场婚恋": "女频",
	"都市":   "男频",
	"都市脑洞": "男频",
	"传统玄幻": "男频",
	"玄幻":   "男频",
	"历史":   "男频",
	"科幻":   "男频",
	"游戏体育": "男频",
}

func ResolveGender(manual, category, genre, ai string) (string, string) {
	if value := normalizeGender(manual); value != "" {
		return value, SourceManual
	}
	if value := normalizeGender(category); value != "" {
		return value, SourceCategory
	}
	if value := verifiedGenreGender[strings.TrimSpace(genre)]; value != "" {
		return value, SourceGenre
	}
	if value := normalizeGender(ai); value != "" {
		return value, SourceAI
	}
	return "", ""
}

func ResolveStyle(manual, provider, ai string) (string, string) {
	if value := strings.TrimSpace(manual); value != "" {
		return value, SourceManual
	}
	if value := strings.TrimSpace(provider); value != "" {
		return value, SourceProvider
	}
	if value := strings.TrimSpace(ai); value != "" {
		return value, SourceAI
	}
	return "", ""
}

func normalizeGender(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	switch value {
	case "男频", "男生", "男性向", "男", "male", "m":
		return "男频"
	case "女频", "女生", "女性向", "女", "female", "f":
		return "女频"
	default:
		return ""
	}
}
