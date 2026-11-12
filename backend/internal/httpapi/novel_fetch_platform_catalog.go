package httpapi

import (
	"fmt"
	"strconv"
	"strings"
)

// Only restore destroyed labels for known IDs. Valid custom names and visibility
// are operator data; unknown IDs must never be assigned a guessed book source.
func novelFetchPlatformCatalog(settings map[string]any) []any {
	source := settings
	if nested, ok := settings["settings"].(map[string]any); ok {
		source = nested
	}
	rows, ok := source["platforms"].([]any)
	if !ok || len(rows) == 0 {
		rows = defaultNovelFetchPlatforms
	}
	known := map[string]string{}
	for _, raw := range defaultNovelFetchPlatforms {
		p := raw.(map[string]any)
		known[fmt.Sprint(p["id"])] = fmt.Sprint(p["name"])
	}
	result := make([]any, 0, len(rows))
	for _, raw := range rows {
		p, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		id := strings.TrimSpace(fmt.Sprint(p["id"]))
		if id == "" || id == "<nil>" {
			continue
		}
		name, _ := p["name"].(string)
		name = strings.TrimSpace(name)
		if strings.Trim(name, "?？� \t\r\n") == "" {
			name = known[id]
		}
		// 历史数据里出现过被剥掉反斜杠的 \uXXXX 转义（如 u9ed1u5ca9u4ed8u8d39）。
		// 这类名字不是运营自定义，而是编码损坏，读时还原；还原不出时回退到内置名。
		if decoded, ok := decodeLegacyUnicodeEscapes(name); ok {
			name = decoded
		}
		if name == "" {
			name = known[id]
		}
		if name == "" {
			continue
		}
		copy := make(map[string]any, len(p))
		for key, value := range p {
			copy[key] = value
		}
		copy["id"] = id
		copy["name"] = name
		result = append(result, copy)
	}
	return result
}

// decodeLegacyUnicodeEscapes 还原被剥掉反斜杠的 \uXXXX 序列，例如
// "u9ed1u5ca9u4ed8u8d39" → "黑岩付费"。只有整串都由 u+4 位十六进制组成时
// 才处理，避免误伤正常的英文或运营自定义名称。
func decodeLegacyUnicodeEscapes(name string) (string, bool) {
	value := strings.TrimPrefix(strings.TrimSpace(name), "\\")
	if len(value) < 10 || value[0] != 'u' {
		return "", false
	}
	var builder strings.Builder
	for i := 0; i < len(value); {
		if value[i] != 'u' || i+5 > len(value) {
			return "", false
		}
		code, err := strconv.ParseUint(value[i+1:i+5], 16, 32)
		if err != nil {
			return "", false
		}
		builder.WriteRune(rune(code))
		i += 5
	}
	decoded := builder.String()
	if decoded == "" || decoded == name {
		return "", false
	}
	return decoded, true
}
