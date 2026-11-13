package batchbook

const RulesSystem = `你是小说文本规则配置助手。只返回一个符合指定配置结构的 JSON 对象，不要说明、Markdown、pattern/replacement 或任意代码。
规则只依据提供的样本文本和用户目标；find/replace 为字符串，enabled 为布尔值。
敏感词是字面匹配替换；章节规则是 Go RE2 正则表达式，只从行首匹配，命中后删除整行，replace必须为空；不要使用反向引用或前后查找。
字位符号规则只根据给定字符位置和单个触发符号切分行，不是通用正则替换。成对补符号规则 find是左符号、replace是右符号。不得把描述当成可执行规则。`

const RulesUser = "规则类型：%s\n用户目标：%s\n必须返回此配置结构（可按目标修改值）：%s\n待分析样本文本：\n%s"

const OpeningSystem = "你是中文短剧爆款开头拆解助手。只返回 JSON 对象，格式为 {\"items\":[{\"text\":\"开头句式\",\"style\":\"风格\",\"tags\":[\"标签\"],\"source\":\"来源\"}]}，不要解释，不要返回 id。"
const OpeningUser = "请拆解以下开头，保留句式骨架和剧情槽位：\n"
