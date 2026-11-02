package batchfactoryv11

import (
	"fmt"
	"strings"
	"unicode"
)

// ComicAssetIdentityIssue reports whether the automatic comic asset set still
// contains labels that are not stable enough for cross-page character
// references. It deliberately focuses on character identity: scene and prop
// names can be descriptive labels, while a character needs a formal name or a
// unique canonical name before the executor reuses it.
func ComicAssetIdentityIssue(book Book) string {
	characterCount := 0
	seen := map[string]string{}
	for _, raw := range book.AssetRecords {
		asset := normalizeLegacyComicBookAsset(raw)
		if strings.TrimSpace(asset.Source) != "comic" || strings.ToLower(strings.TrimSpace(asset.Kind)) != "character" {
			continue
		}
		characterCount++
		name := strings.TrimSpace(asset.Name)
		formalName := strings.TrimSpace(asset.FormalName)
		if name == "" {
			return "存在没有名称的人物资产"
		}
		if formalName == "" && genericComicIdentityLabel(name) {
			return fmt.Sprintf("人物“%s”没有正式人物名", name)
		}
		identity := formalName
		if identity == "" {
			identity = name
		}
		key := normalizeComicIdentityKey(identity)
		if previous, exists := seen[key]; exists {
			return fmt.Sprintf("人物“%s”和“%s”使用了同一个身份名称", previous, name)
		}
		seen[key] = name
	}
	if characterCount == 0 {
		// Older comic asset records did not persist Kind. Keep those records
		// reusable; the next extraction will add the structured kind and formal
		// name fields without making a legacy batch fail only because of storage
		// shape.
		return ""
	}
	return ""
}

func genericComicIdentityLabel(value string) bool {
	key := normalizeComicIdentityKey(value)
	if key == "" {
		return true
	}
	for _, candidate := range []string{
		"我", "他", "她", "你", "我方", "对方", "人物", "角色", "主角", "女主", "男主", "女配", "男配",
		"夫人", "老爷", "少爷", "小姐", "哥哥", "姐姐", "弟弟", "妹妹", "妈妈", "母亲", "爸爸", "父亲",
		"丫鬟", "侍女", "管家", "掌柜", "老人", "小孩", "孩子", "男人", "女人", "男子", "女子", "陌生人",
	} {
		if key == candidate {
			return true
		}
	}
	if strings.HasPrefix(key, "人物") || strings.HasPrefix(key, "角色") {
		rest := strings.TrimPrefix(strings.TrimPrefix(key, "人物"), "角色")
		return rest == "" || allDigits(rest)
	}
	return false
}

func normalizeComicIdentityKey(value string) string {
	var builder strings.Builder
	for _, r := range strings.TrimSpace(value) {
		if unicode.IsSpace(r) || unicode.IsPunct(r) || unicode.IsSymbol(r) {
			continue
		}
		builder.WriteRune(r)
	}
	return strings.ToLower(builder.String())
}

func allDigits(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
