package novelfetchworkshop

import (
	"fmt"
	"strings"
)

var safeSensitiveReplacements = map[string]string{
	"绿帽子":   "颜色帽子",
	"绿帽":    "颜色帽",
	"家暴":    "家庭伤害",
	"家庭暴力":  "家庭伤害",
	"虐童":    "儿童伤害",
	"虐待":    "伤害行为",
	"龌龊":    "不雅表达",
	"荡妇":    "失礼称呼",
	"双修":    "共同修习",
	"青楼":    "风月场所",
	"屌丝":    "普通青年",
	"碧莲":    "不当称呼",
	"特么的":  "情绪化用语",
	"草泥马":  "情绪化用语",
	"装B":    "卖弄表现",
	"装逼":    "卖弄表现",
	"黑鬼":    "不当称呼",
	"小日本":  "东邻称呼",
	"下药":    "投放药物",
	"下了药":  "投放药物",
	"守活寡":  "长期独居",
	"红灯区":  "特殊场所",
	"私生子":  "非婚生子女",
	"人贩子":  "非法拐卖者",
	"劈腿":    "关系越界",
	"外遇":    "关系越界",
	"婚外情":  "关系越界",
	"恋尸癖":  "异常癖好",
	"搞大肚子": "意外怀孕",
	"私通":    "私下交往",
	"强暴":    "强行伤害",
	"露沟":    "服装暴露",
	"走光":    "服装暴露",
	"露点":    "服装暴露",
	"乱伦":    "不当关系",
	"恋童":    "不当关系",
	"恋童癖":  "异常癖好",
	"催情":    "特殊药物",
	"延时喷剂": "特殊药物",
	"春药":    "特殊药物",
	"迷药":    "特殊药物",
	"媚药":    "特殊药物",
	"爆菊":    "不雅行为",
	"撸管":    "不雅行为",
	"潮吹":    "不雅行为",
	"口交":    "不雅行为",
	"群交":    "不雅行为",
	"足交":    "不雅行为",
	"第三者":   "关系介入者",
	"小三":    "介入关系的人",
	"出轨":    "关系越界",
	"勾引":    "主动接近",
	"一夜情":   "短暂关系",
	"上床":    "私下接触",
	"强奸":    "非法伤害",
	"赌博":    "违规下注",
	"花魁":    "名伶",
	"轮奸":    "多人伤害",
	"奸夫":    "关系越界者",
	"淫妇":    "关系越界者",
	"贱人":    "失礼之人",
	"婊子":    "失礼之人",
	"狗东西":   "讨厌家伙",
	"激情视频":  "凌晨快乐视频",
	"开房":    "外面睡觉",
	"捉奸":    "发现关系越界",
	"爬上床":   "深夜留宿",
	"避孕套":   "防护用品",
	"强上":    "强行靠近",
	"禁脔":    "专属对象",
	"骚货":    "轻浮称呼",
	"私会":    "私下见面",
	"搞我闺蜜":  "介入朋友关系",
	"拉皮条":   "私下牵线",
	"床上技术":  "亲密相处方式",
	"鬼混":    "混乱交往",
	"与多人有染": "同时与多人交往",
	"潜规则":   "隐性约定",
	"伟哥":    "活力药物",
	"把妹":    "主动结识女生",
	"约炮":    "临时约会",
	"裸聊":    "私密聊天",
	"卧槽":    "惊讶感叹",
	"我草":    "惊讶感叹",
	"我操":    "惊讶感叹",
	"我擦":    "惊讶感叹",
	"握草":    "惊讶感叹",
}

var safeSensitiveHomophones = map[rune]rune{
	'奸': '间', '淫': '隐', '色': '瑟', '骚': '燥', '婊': '表', '贱': '见',
	'狗': '够', '杀': '煞', '赌': '堵', '毒': '独', '嫖': '飘', '妓': '技',
	'裸': '落', '乳': '如', '胸': '凶', '性': '幸', '欲': '遇', '插': '叉',
	'操': '超', '妈': '麻', '逼': 'B', '骂': '嘛', '偷': '透', '骗': '片',
	'怂': '松', '蛋': '弹', '猥': '委', '琐': '锁', '老': '佬', '资': '姿',
	'丫': '呀', '房': '方', '床': '创', '湿': '诗', '喘': '船',
	'舔': '添', '吻': '稳', '娇': '焦', '暴': '抱', '力': '利', '虐': '略',
	'龌': '握', '龊': '错', '荡': '当', '妇': '服', '屌': '调', '莲': '联',
	'伦': '论', '童': '同', '癖': '僻', '药': '要', '露': '路', '沟': '勾',
	'点': '典', '器': '气', '官': '关',
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
			if find == "" {
				continue
			}
			existing := strings.TrimSpace(fmt.Sprint(item["replace"]))
			replacement := safeSensitiveReplacement(find)
			// A previous version could persist the original word as its own
			// replacement. That makes the rule look enabled while leaving the
			// prohibited expression untouched, so only that broken value is
			// repaired; a real user-authored replacement still wins.
			if existing != "" && existing != find {
				continue
			}
			if replacement == find {
				// Unknown/benign catalog entries keep their readable original
				// wording when no reviewed neutral phrase exists yet. This is
				// only for an empty value; a persisted same-as-find value is
				// left alone until its term gets an explicit mapping.
				if existing != "" {
					continue
				}
				item["replace"] = find
				changed++
				continue
			}
			item["replace"] = replacement
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
		// The rules editor reads the canonical knowledge.sensitive catalog. The
		// older locations above are kept for compatibility, but leaving this
		// view untouched makes the same words appear without replacements.
		fillGroups(mapOrEmpty(knowledge["sensitive"])["groups"])
		fill(mapOrEmpty(knowledge["layout_rules"])["sensitive_rules"])
		libraries := mapOrEmpty(root["knowledgeLibraries"])
		fillGroups(mapOrEmpty(libraries["sensitive"])["groups"])
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
	case containsAny(find, "家暴", "家庭暴力", "虐童", "虐待老人", "殴打家人"):
		return "家庭伤害"
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
	case containsAny(find, "奸", "淫", "骚", "妓", "嫖", "裸", "乳", "胸", "性", "色", "欲", "精", "射", "插", "喘", "湿", "阴", "阳", "床", "房", "洞房", "交配", "配种", "情趣", "情欲", "亲热", "热吻", "爱液", "液", "叫", "娇", "裙", "裤", "丝袜", "内衣", "诱惑", "春宵", "发情", "欢", "交欢", "侍寝", "承欢", "调情", "挑逗", "撩", "酥", "痉挛", "颤", "缠", "滴蜡", "炮友", "换妻", "陪睡", "包夜", "全套", "援交", "卖身", "包养", "房事", "上过床", "爬床", "滚床", "床单", "堕胎", "流产", "取卵", "避孕", "捐精", "借种", "上床", "床上", "约炮", "一夜"):
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
		// Do not replace an unknown phrase with an opaque label. Keeping the
		// original is more readable for ordinary false-positive entries and
		// leaves the term available for a later, reviewed dictionary entry.
		return find
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
		// Scan a stable snapshot once. Repeatedly searching the mutated result
		// can loop forever when a replacement contains the original find text
		// (for example, find="A", replace="A说明").
		source := result
		var builder strings.Builder
		cursor := 0
		for {
			index := strings.Index(source[cursor:], find)
			if index < 0 {
				builder.WriteString(source[cursor:])
				break
			}
			index += cursor
			builder.WriteString(source[cursor:index])
			before := sensitiveLine(source, index)
			builder.WriteString(replace)
			after := strings.Replace(before, find, replace, 1)
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
			cursor = index + len(find)
		}
		result = builder.String()
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
