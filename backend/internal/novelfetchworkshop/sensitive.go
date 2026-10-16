package novelfetchworkshop

import (
	"fmt"
	"strings"
)

var safeSensitiveReplacements = map[string]string{
	"绿帽子":     "颜色帽子",
	"绿帽":      "颜色帽",
	"家暴":      "家庭伤害",
	"家庭暴力":    "家庭伤害",
	"虐童":      "儿童伤害",
	"虐待":      "伤害行为",
	"龌龊":      "不雅表达",
	"荡妇":      "失礼称呼",
	"双修":      "共同修习",
	"青楼":      "风月场所",
	"屌丝":      "普通青年",
	"碧莲":      "不当称呼",
	"特么的":     "情绪化用语",
	"草泥马":     "情绪化用语",
	"装B":      "卖弄表现",
	"装逼":      "卖弄表现",
	"黑鬼":      "不当称呼",
	"小日本":     "东邻称呼",
	"下药":      "投放药物",
	"下了药":     "投放药物",
	"守活寡":     "长期独居",
	"红灯区":     "特殊场所",
	"私生子":     "非婚生子女",
	"人贩子":     "非法拐卖者",
	"玩了多少小明星": "与多位公众人物交往",
	"知三当三":    "明知介入关系",
	"龟儿子":     "失礼称呼",
	"狗男女":     "关系失当者",
	"劈腿":      "关系越界",
	"外遇":      "关系越界",
	"婚外情":     "关系越界",
	"恋尸癖":     "异常癖好",
	"搞大肚子":    "意外怀孕",
	"私通":      "私下交往",
	"强暴":      "强行伤害",
	"露沟":      "服装暴露",
	"走光":      "服装暴露",
	"露点":      "服装暴露",
	"乱伦":      "不当关系",
	"恋童":      "不当关系",
	"恋童癖":     "异常癖好",
	"催情":      "特殊药物",
	"延时喷剂":    "特殊药物",
	"春药":      "特殊药物",
	"迷药":      "特殊药物",
	"媚药":      "特殊药物",
	"爆菊":      "不雅行为",
	"撸管":      "不雅行为",
	"潮吹":      "不雅行为",
	"口交":      "不雅行为",
	"群交":      "不雅行为",
	"足交":      "不雅行为",
	"第三者":     "关系介入者",
	"小三":      "介入关系的人",
	"出轨":      "关系越界",
	"勾引":      "主动接近",
	"一夜情":     "短暂关系",
	"上床":      "私下接触",
	"强奸":      "非法伤害",
	"赌博":      "违规下注",
	"花魁":      "名伶",
	"轮奸":      "多人伤害",
	"奸夫":      "关系越界者",
	"淫妇":      "关系越界者",
	"贱人":      "失礼之人",
	"婊子":      "失礼之人",
	"狗东西":     "讨厌家伙",
	"激情视频":    "凌晨快乐视频",
	"开房":      "外面睡觉",
	"捉奸":      "发现关系越界",
	"爬上床":     "深夜留宿",
	"避孕套":     "防护用品",
	"强上":      "强行靠近",
	"禁脔":      "专属对象",
	"骚货":      "轻浮称呼",
	"私会":      "私下见面",
	"搞我闺蜜":    "介入朋友关系",
	"拉皮条":     "私下牵线",
	"床上技术":    "亲密相处方式",
	"鬼混":      "混乱交往",
	"与多人有染":   "同时与多人交往",
	"潜规则":     "隐性约定",
	"伟哥":      "活力药物",
	"把妹":      "主动结识女生",
	"约炮":      "临时约会",
	"裸聊":      "私密聊天",
	"卧槽":      "惊讶感叹",
	"我草":      "惊讶感叹",
	"我操":      "惊讶感叹",
	"我擦":      "惊讶感叹",
	"握草":      "惊讶感叹",
	// 规则库中这些词是明确的违规表达；单独列出是为了让替换结果保留
	// 原意，而不是被宽泛的“相关内容”吞掉。普通同名词仍由白名单保留。
	"仙人板板": "失礼称呼", "高丽棒子": "地域歧视称呼", "东亚病夫": "地域歧视称呼",
	"衣冠禽兽": "失德之人", "杂种": "失礼称呼", "万人骑": "不雅称呼", "千人骑万人乘": "不雅称呼",
	"乱搞": "混乱交往", "孽种": "失礼称呼", "黄谣": "不实传言", "擦边": "暧昧内容",
	"事业线": "服装线条", "黑丝": "深色长袜", "骨科": "不当亲属关系", "听话水": "诱导性药物",
	"外围女": "陪同人员", "楼凤": "特殊服务称呼", "绿奴": "关系角色", "字母圈": "兴趣圈层", "SM": "特殊互动",
	"捆绑": "限制行动", "幼女": "未成年女孩", "萝莉控": "未成年偏好", "破处": "首次亲密经历",
	"大尺度": "高强度内容", "车震": "车内亲密行为", "野战": "户外冲突", "肉番": "成人向作品",
	"香艳": "暧昧氛围", "呻吟": "痛苦声音", "玉体": "身体", "胴体": "身体",
	"炉鼎": "修炼对象", "采补": "夺取功力", "弄死": "造成死亡", "灭门": "家族覆灭",
	"活埋": "掩埋伤害", "凌迟": "残酷刑罚", "腰斩": "残酷刑罚", "自残": "伤害自己",
	"轻生": "结束生命", "同归于尽": "共同承受毁灭", "报复社会": "极端报复",
	"摇头丸": "违禁药物", "大麻": "违禁药物", "六合彩": "违规博彩", "走私": "非法运输",
	"保护伞": "违规庇护", "游行": "公开集会", "贩卖人口": "非法拐卖", "电诈": "网络诈骗",
	"地域歧视": "地域偏见", "种族歧视": "群体偏见", "调教": "关系训练", "主奴": "支配关系",
	"榨干": "过度消耗", "弄死你": "伤害威胁", "真丝吊带": "细肩带服装", "强制爱": "强迫关系",
	"囚禁": "限制自由", "锁链": "束缚物", "挂空挡": "车辆空挡", "窥视": "暗中观察",
	"渔网袜": "网格长袜", "兔女郎": "主题服饰", "跪下": "屈身动作", "碎尸": "肢体破坏",
	"抛尸": "遗弃遗体", "毁容": "面部伤害", "割喉": "颈部伤害", "撕票": "伤害人质",
	"挖眼": "眼部伤害", "剁手指": "手部伤害", "割腕": "腕部伤害", "跳楼": "高处坠落",
	"灭口": "消除证人", "校园霸凌": "校园欺凌", "PUA": "情感操控", "嘎腰子": "非法取器官",
	"缅北": "境外高风险地区", "高利贷": "违规借贷", "洗脑": "强行灌输", "脑残": "失礼称呼",
	"孤儿": "失去双亲的孩子", "猪猡": "失礼称呼", "蛆虫": "失礼称呼", "下三滥": "低俗手段",
	"笑气": "特殊气体", "针孔摄像": "隐蔽摄像", "卖血": "有偿献血", "地下钱庄": "非法资金渠道",
	"监听": "暗中监听", "跟踪": "尾随观察", "威胁": "施压行为", "恐吓": "惊吓威逼",
	"尸块": "遗体碎片", "剥皮": "皮肤伤害", "抽筋": "肌肉痉挛", "扇耳光": "掌掴动作",
	"跪地求饶": "屈身求情", "软禁": "限制自由", "阉割": "身体伤害", "乖乖水": "诱导性药物",
	"焚尸": "焚烧遗体", "尸检": "遗体检查", "乱葬岗": "遗体安置地", "胯下": "身体下方",
	"挺进": "向前推进", "顶弄": "推撞动作", "吮吸": "吸取动作", "亵玩": "不当玩弄",
	"惹火": "引发冲突", "蹂躏": "严重伤害", "挺弄": "推撞动作", "摩挲": "轻抚动作",
	"翻云覆雨": "关系纠葛", "巫山云雨": "亲密往来", "颠鸾倒凤": "关系混乱", "殉情": "共同赴死",
	"私生": "非婚生育", "雌竞": "女性竞争", "卖肾": "非法取器官", "被讹": "遭遇敲诈",
	"师生恋": "师生关系越界", "胁迫": "强迫行为", "强迫": "施加强制", "死刑": "极刑判决",
	"夜袭": "夜间突袭", "千术": "博彩骗术", "QJ": "非法伤害", "侵犯": "越界伤害",
	"野鸡": "非正规对象", "拐仔": "非法拐带者", "拐女": "非法拐带者", "霸凌": "欺凌行为",
	"网恋": "网络交往", "立案": "司法立案", "犯": "违规者", "撞死": "造成死亡",
	"刑侦": "刑事调查", "包庇": "违规袒护", "寡嫂": "失伴嫂嫂", "私生女": "非婚生女儿",
	"戴套": "使用防护用品", "戴驴套": "使用防护用品", "牛马": "辛苦劳动者", "后宫": "内宅群像",
	"诅咒": "恶意祝愿", "咒我": "恶意针对", "污蔑": "不实指控", "祸乱": "混乱局面",
	"暧昧": "模糊关系", "吐口水": "吐液体动作", "造假": "伪造行为", "需求旺盛": "需求较多",
	"涉黄": "涉不雅内容", "顶罪": "替人担责", "牢底": "长期服刑", "冲喜": "婚俗安排",
	"下头": "令人不适", "虾头": "失礼称呼", "脏病": "传染性疾病", "非法": "违规行为",
	"飞刀": "投掷利器", "顶替": "冒名替代", "内衣": "贴身衣物", "丝袜": "长袜",
	"政权": "治理体系", "体制": "制度体系", "意识形态": "观念体系", "公务员": "公职人员",
	"下海": "转行谋生", "私密": "私人内容", "成人": "成年人",
	"乃子": "胸部", "奶子": "胸部", "大奶": "丰满身材", "轮大米": "多人侵犯",
	"误睡": "意外同宿", "隐私照": "私密照片", "三奶": "关系角色", "二奶": "关系角色",
	"无情道": "清冷修行路", "替身": "替代身份", "为闺蜜": "为朋友", "报了警": "已报警",
	"罚款": "行政处罚", "碰瓷": "故意讹诈", "越狱": "逃离监管", "外网": "外部网络",
	"拆迁": "搬迁安置", "缴获": "查扣物品", "无名单": "未列名", "观赛": "观看比赛",
	"人亡": "人员离世", "失联": "失去联系", "流水": "资金流转", "销户": "注销账户",
	"烈士": "牺牲人员", "纸人陪他": "纸人陪伴", "成人向": "成熟向",
	"揭穿": "揭示真相", "曝光": "公开信息", "权倾朝野": "影响力很大", "暗通款曲": "私下往来",
	"渣男": "失信男性", "牺牲": "付出代价", "毛贼": "小偷", "提审": "提讯调查",
	"徒刑": "刑期", "无期": "长期刑期", "火警": "火灾警报", "侵": "越界行为",
	"警报": "风险提示", "祖安": "网络争执", "祖安文化": "网络争执文化", "矮矬穷": "外貌攻击称呼",
	"贩子": "非法交易者", "人贩": "非法拐卖者", "号召": "公开倡议", "大群": "多人群组",
	"决赛": "最终比赛", "带煞": "带有凶险意味",
	"引魂人": "引导亡者的人", "大吏": "高级官员", "封疆": "边地管理", "书记": "记录人员",
	"主任": "负责人", "民族": "族群", "央视": "电视媒体", "大出血": "大量失血",
	"睡了": "同宿过", "进酒店": "进入住处", "复印件": "复制文件", "身份证": "身份凭证",
	"顶级": "高等级", "渣夫": "失信丈夫",
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
