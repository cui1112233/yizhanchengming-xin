package novelfetchworkshop

import (
	"fmt"
	"strings"
)

var safeSensitiveReplacements = map[string]string{
	"绿帽子": "颜色帽子",
	"绿帽":  "颜色帽",
	"第三者": "关系介入者",
	"小三":  "介入关系的人",
	"出轨":  "关系越界",
	"勾引":  "主动接近",
	"一夜情": "短暂关系",
	"上床":  "私下接触",
	"强奸":  "非法伤害",
	"赌博":  "违规下注",
	"花魁":  "名伶",
	"轮奸":  "多人伤害",
	"奸夫":  "关系越界者",
	"淫妇":  "关系越界者",
	"贱人":  "失礼之人",
	"婊子":  "失礼之人",
	"狗东西": "讨厌家伙",
	"激情视频": "凌晨快乐视频",
	"开房": "外面睡觉",
	"卧槽":  "惊讶感叹",
	"我草":  "惊讶感叹",
	"我操":  "惊讶感叹",
	"我擦":  "惊讶感叹",
	"握草":  "惊讶感叹",
}

var safeSensitiveHomophones = map[rune]rune{
	'奸': '间', '淫': '隐', '色': '瑟', '骚': '燥', '婊': '表', '贱': '见',
	'狗': '够', '杀': '煞', '赌': '堵', '毒': '独', '嫖': '飘', '妓': '技',
	'裸': '落', '乳': '如', '胸': '凶', '性': '幸', '欲': '遇', '插': '叉',
	'操': '超', '妈': '麻', '逼': 'B', '骂': '嘛', '偷': '透', '骗': '片',
}

var safeSensitivePassThrough = map[string]struct{}{
	"玉米": {}, "树林": {}, "小树林": {}, "视频": {}, "服务": {}, "晚上": {}, "深夜": {},
	"直播": {}, "护士": {}, "院长": {}, "男老师": {}, "男女": {}, "睡": {}, "玩": {},
	"第一": {}, "最佳": {}, "首选": {}, "顶尖": {}, "公认": {}, "必选": {}, "最火": {},
}

// EnsureSensitiveRuleReplacements fills empty sensitive-word replacements
// with neutral wording. Existing user replacements always win.
func EnsureSensitiveRuleReplacements(settings map[string]any) int {
	if settings == nil {
		return 0
	}
	changed := 0
	fill := func(raw any) {
		items, ok := raw.([]any)
		if !ok {
			return
		}
		for _, value := range items {
			item, ok := value.(map[string]any)
			if !ok {
				continue
			}
			find := strings.TrimSpace(fmt.Sprint(item["find"]))
			if find == "" || strings.TrimSpace(fmt.Sprint(item["replace"])) != "" {
				continue
			}
			item["replace"] = safeSensitiveReplacement(find)
			changed++
		}
	}
	fillGroups := func(raw any) {
		groups, ok := raw.([]any)
		if !ok {
			return
		}
		for _, rawGroup := range groups {
			if group, ok := rawGroup.(map[string]any); ok {
				fill(group["rules"])
			}
		}
	}
	for _, root := range []map[string]any{settings, mapOrEmpty(settings["appConfig"]), mapOrEmpty(settings["app_config"])} {
		fillGroups(mapOrEmpty(root["sensitive"])["groups"])
		knowledge := mapOrEmpty(root["knowledge"])
		fill(mapOrEmpty(knowledge["layout_rules"])["sensitive_rules"])
		libraries := mapOrEmpty(root["knowledgeLibraries"])
		fill(mapOrEmpty(libraries["layout_rules"])["sensitive_rules"])
	}
	return changed
}

func safeSensitiveReplacement(find string) string {
	if replacement, ok := safeSensitiveReplacements[find]; ok {
		return replacement
	}
	if _, ok := safeSensitivePassThrough[find]; ok {
		return find
	}
	// Keep the replacement understandable even when a newly imported rule is
	// not in the hand-curated table above. These broad buckets are intentionally
	// neutral and preserve the scene's meaning better than deleting the word.
	switch {
	case containsAny(find, "卧槽", "我草", "我操", "我擦", "握草", "靠"):
		return "惊讶感叹"
	case containsAny(find, "强奸", "轮奸", "奸杀", "先奸", "强上", "绑架", "虐待"):
		return "伤害行为"
	case containsAny(find, "第三者", "小三", "出轨", "情妇", "偷情", "私会", "私奔", "绿帽", "绿茶"):
		return "关系介入者"
	case containsAny(find, "赌博", "赌", "博彩", "下注"):
		return "违规下注"
	case containsAny(find, "毒品", "吸毒", "冰毒", "海洛因", "鸦片", "涉毒"):
		return "违禁品"
	case containsAny(find, "警察", "公安", "民警", "武警", "刑警", "警局", "派出所", "执法"):
		return "执法人员"
	case containsAny(find, "法院", "法官", "检察院", "司法", "审判", "审判"):
		return "司法机构"
	case containsAny(find, "政府", "中央", "市委", "县委", "常委", "政委", "纪委", "民政", "教育局", "税务局", "交通局", "卫健委", "招生办"):
		return "公共机构"
	case containsAny(find, "军队", "部队", "军区", "特战", "军校", "当兵", "军车", "国防"):
		return "相关队伍"
	case containsAny(find, "丈夫", "老公", "妻子", "老婆", "情人", "对象", "婆婆", "嫂子", "妹夫", "小叔", "舍友男友"):
		return "人物关系"
	case containsAny(find, "贷款", "征信", "合同", "收费", "红包", "中奖", "提现", "回扣", "代付", "担保", "债"):
		return "相关事务"
	case containsAny(find, "佛", "菩萨", "活佛", "庇佑", "冥婚", "邪教"):
		return "宗教相关称谓"
	case containsAny(find, "癌", "脑瘤", "肾虚", "精神病", "萎缩", "确诊", "疾病", "贫困"):
		return "健康状况"
	case containsAny(find, "诈骗", "偷税", "受贿", "贿赂", "传销", "拐卖", "造谣", "网暴", "勒索", "敲诈", "洗钱", "偷渡", "骗婚", "酒驾"):
		return "违规行为"
	case containsAny(find, "奸", "淫", "骚", "妓", "嫖", "裸", "乳", "胸", "性", "色", "欲", "精", "射", "插", "上床", "床上", "约炮", "一夜"):
		if replacement := neutralizeByHomophone(find); replacement != find {
			return replacement
		}
		return "亲密行为"
	case containsAny(find, "杀", "砍", "打", "暴力", "暴打", "行凶", "伤害", "血腥", "凶杀"):
		return "伤害行为"
	case containsAny(find, "傻", "蠢", "贱", "狗", "畜", "王八", "废物", "垃圾", "婊", "逼", "妈的", "他妈"):
		if replacement := neutralizeByHomophone(find); replacement != find {
			return replacement
		}
		return "不当称呼"
	default:
		if replacement := neutralizeByHomophone(find); replacement != find {
			return replacement
		}
		return "中性说法"
	}
}

func neutralizeByHomophone(find string) string {
	var builder strings.Builder
	changed := false
	for _, char := range find {
		if replacement, ok := safeSensitiveHomophones[char]; ok {
			builder.WriteRune(replacement)
			changed = true
			continue
		}
		builder.WriteRune(char)
	}
	if !changed {
		return find
	}
	return builder.String()
}

func containsAny(value string, parts ...string) bool {
	for _, part := range parts {
		if strings.Contains(value, part) {
			return true
		}
	}
	return false
}

// ApplySensitiveRules applies the saved sensitive-word replacements to the
// working text. It never touches Document.OriginalRaw; callers decide which
// processed field to save. Both the current map form and the legacy
// "@{enabled=True; find=x; replace=y}" form are accepted.
func ApplySensitiveRules(text string, settings map[string]any) (string, int) {
	result, changes, _ := ApplySensitiveRulesWithDetails(text, settings)
	return result, changes
}

// ApplySensitiveRulesWithDetails applies the same deterministic replacements
// as ApplySensitiveRules and also returns the before/after line for every hit.
// The detail records are used by the task detail page and are safe to persist
// without exposing the untouched raw document.
func ApplySensitiveRulesWithDetails(text string, settings map[string]any) (string, int, []map[string]any) {
	if text == "" || settings == nil {
		return text, 0, []map[string]any{}
	}
	rules := make([]any, 0)
	collectSensitiveRules(settings, &rules)
	result := text
	changes := 0
	items := make([]map[string]any, 0)
	for _, raw := range rules {
		find, replace, enabled := sensitiveRule(raw)
		if !enabled || find == "" || find == replace {
			continue
		}
		for {
			index := strings.Index(result, find)
			if index < 0 {
				break
			}
			before := sensitiveLine(result, index)
			result = result[:index] + replace + result[index+len(find):]
			after := sensitiveLine(result, index)
			changes++
			items = append(items, map[string]any{
				"hit_index": len(items) + 1,
				"keyword":   find,
				"before":    before,
				"after":     after,
				"status":    "done",
				"source":    "system_rule",
				"error":     "",
			})
		}
	}
	return result, changes, items
}

func sensitiveLine(text string, index int) string {
	start := strings.LastIndex(text[:index], "\n") + 1
	end := strings.IndexByte(text[index:], '\n')
	if end < 0 {
		end = len(text)
	} else {
		end += index
	}
	return text[start:end]
}

func collectSensitiveRules(value any, out *[]any) {
	switch typed := value.(type) {
	case map[string]any:
		for _, key := range []string{"sensitive_rules", "sensitiveRules", "rules"} {
			if raw, ok := typed[key]; ok {
				if key == "rules" {
					// Only groups under a sensitive section are meaningful; plain
					// knowledge item rules are handled by their own features.
					if _, hasSensitive := typed["groups"]; !hasSensitive {
						continue
					}
				}
				collectSensitiveRules(raw, out)
			}
		}
		if groups, ok := typed["groups"]; ok {
			collectSensitiveRules(groups, out)
		}
		for _, key := range []string{"knowledge", "app_config", "appConfig", "layout_rules", "layout", "sensitive"} {
			if nested, ok := typed[key]; ok {
				collectSensitiveRules(nested, out)
			}
		}
	case []any:
		*out = append(*out, typed...)
	case []map[string]any:
		for _, item := range typed {
			*out = append(*out, item)
		}
	}
}

func sensitiveRule(raw any) (string, string, bool) {
	if item, ok := raw.(map[string]any); ok {
		find := strings.TrimSpace(stringValue(item["find"], item["pattern"], item["source"]))
		replace := stringValue(item["replace"], item["replacement"], item["target"])
		enabled := true
		if value, exists := item["enabled"]; exists {
			enabled = boolValue(value)
		}
		return find, replace, enabled
	}
	value := strings.TrimSpace(stringValue(raw))
	if !strings.HasPrefix(value, "@{") || !strings.HasSuffix(value, "}") {
		return "", "", false
	}
	body := strings.TrimSuffix(strings.TrimPrefix(value, "@{"), "}")
	fields := map[string]string{}
	for _, part := range strings.Split(body, ";") {
		pieces := strings.SplitN(part, "=", 2)
		if len(pieces) == 2 {
			fields[strings.ToLower(strings.TrimSpace(pieces[0]))] = strings.TrimSpace(pieces[1])
		}
	}
	enabled := fields["enabled"] == "" || strings.EqualFold(fields["enabled"], "true") || fields["enabled"] == "1"
	return fields["find"], fields["replace"], enabled
}

func stringValue(values ...any) string {
	for _, value := range values {
		if value == nil {
			continue
		}
		if text, ok := value.(string); ok {
			if strings.TrimSpace(text) != "" {
				return text
			}
			continue
		}
		return strings.TrimSpace(fmt.Sprint(value))
	}
	return ""
}

func boolValue(value any) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		return strings.EqualFold(strings.TrimSpace(typed), "true") || strings.TrimSpace(typed) == "1"
	default:
		return false
	}
}
