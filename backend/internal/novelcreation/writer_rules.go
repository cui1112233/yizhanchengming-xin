package novelcreation

import "strings"

// 洛水 novels.prompts.inkos.writer_prompts 的写作规则块移植。
// 原文区块只进入 system prompt，不进入正文；输出仍保持 JSON 合同，
// 避免把 Markdown 表格区块混入 content 字段。
//
// 版本由 WriterRuleVersion 门控：旧任务没有该快照字段，重放时不追加任何块，
// 保证历史请求指纹不变。

type WriterRuleInput struct {
	Genre         string
	ChapterNumber int
	TargetWords   int
	HasHookDebt   bool
	HasRetrieval  bool
}

// genreWritingBodies 对应洛水 GENRE_BODIES 的题材正文描述。
var genreWritingBodies = map[string]string{
	"xuanhuan": "玄幻升级流：力量体系清晰、节奏明快、爽点密集，主角成长线贯穿全书。",
	"玄幻":       "玄幻升级流：力量体系清晰、节奏明快、爽点密集，主角成长线贯穿全书。",
	"xianxia":  "仙侠修真：境界体系、宗门势力、天地规则，注重修炼逻辑与因果。",
	"仙侠":       "仙侠修真：境界体系、宗门势力、天地规则，注重修炼逻辑与因果。",
	"urban":    "都市现实：贴近当代生活，冲突来自社会关系、职场、家庭与阶层。",
	"都市":       "都市现实：贴近当代生活，冲突来自社会关系、职场、家庭与阶层。",
	"romance":  "言情：情感推进为核心，关系张力与情绪回报是主发动机。",
	"言情":       "言情：情感推进为核心，关系张力与情绪回报是主发动机。",
	"suspense": "悬疑：线索递进、误导与反转并重，每章结尾必须留下悬念钩子。",
	"mystery":  "悬疑：线索递进、误导与反转并重，每章结尾必须留下悬念钩子。",
	"悬疑":       "悬疑：线索递进、误导与反转并重，每章结尾必须留下悬念钩子。",
	"history":  "历史：朝堂、战争与民间生活交织，时代常识与礼法不可出错。",
	"历史":       "历史：朝堂、战争与民间生活交织，时代常识与礼法不可出错。",
	"scifi":    "科幻：核心设定自洽，每章只集中展示一个关键设定。",
	"科幻":       "科幻：核心设定自洽，每章只集中展示一个关键设定。",
	"game":     "游戏：副本、升级与数值体系前后一致，战斗节奏明快。",
	"游戏":       "游戏：副本、升级与数值体系前后一致，战斗节奏明快。",
	"horror":   "恐怖：恐惧递进，不一次用尽最恐怖元素，章末留悬念。",
	"恐怖":       "恐怖：恐惧递进，不一次用尽最恐怖元素，章末留悬念。",
}

// BuildWriterRuleBlocks 返回应追加到 writer system prompt 的规则块。
func BuildWriterRuleBlocks(input WriterRuleInput) []string {
	blocks := []string{writerInputGovernance}
	if body := genreWritingBodies[strings.ToLower(strings.TrimSpace(input.Genre))]; body != "" {
		blocks = append(blocks, "## 题材写作要求\n\n"+body)
	} else if body := genreWritingBodies[strings.TrimSpace(input.Genre)]; body != "" {
		blocks = append(blocks, "## 题材写作要求\n\n"+body)
	}
	blocks = append(blocks, writerCraftRules, writerDensityAndHook, writerAntiAITells, writerProseExecution)
	if input.ChapterNumber > 0 && input.ChapterNumber <= 3 {
		blocks = append(blocks, writerOpeningThreeChapters)
	}
	if input.HasHookDebt {
		blocks = append(blocks, writerHookDebt)
	}
	blocks = append(blocks, `## 输出纪律

本任务的 JSON 在既有 content 合同之上扩展两个字段，不得改动 content 的必填要求：

- preWriteCheck：数组，每项 {"item":"检查项","record":"本章记录","note":"备注"}；只放写前检查表，表格内容不得混入 content。
- title：字符串，只写章节名，不含「第X章」；与已有标题重复时系统会回退为章节计划标题。
- content：本章正文，不写标题、不写解释、不写状态卡、伏笔池或摘要；这些由后续结算步骤处理。`)
	return blocks
}

const writerInputGovernance = `## 输入治理契约

- 本章具体写什么，以 chapter_plan 和已提供的上下文为准。
- 卷纲是默认规划，不是全局最高规则；当本章目标与卷纲冲突时，先完成本章目标。
- 真正不能突破的硬护栏只有：世界设定、连续性事实、显式禁令。
- 如果提供了 hook_health 或 retrieved_chapters，把它们当作连续性约束；缺失的检查项不代表通过。
- 多角色场景至少给出一轮带阻力的直接交锋，不要把人物关系写成解释或总结。`

const writerCraftRules = `## 写作铁律

- 情绪用动作外化：不写「他感到愤怒」，写具体动作与身体反应。
- 价值观靠行为传达，不喊口号。
- 每个配角都有自己的利益和反击能力，不是工具人。
- 场景里放入 1-2 种五感细节，顺手带过，不写成大段天气描写。
- 具体化：不写「大城市」，写读者摸得到的地面细节。
- 少用「虽然/但是/然而/因此」等转折词，用角色反应替代。
- 制造情绪缺口，释放必须超过读者预期；满足七成等于失败。
- 每个角色行动前自问：为什么做？符合人设吗？读者会觉得突兀吗？
- 不同角色说话方式必须有差异：用词、句长、口头禅。
- 禁止资料卡式介绍角色、一次引入三个以上新角色、群体齐声惊呼。
- 坏事分层叠加，每层比上一层更过分。
- 角色只能基于自己已掌握的信息行动。`

const writerDensityAndHook = `## 看点密度与断章

- 每 300 字至少一个可感知的小看点；每 500 字至少抛出一个「接下来怎样」的小悬念。
- 出现连续 300 字以上纯环境、回忆、议论、心理独白而不推进主线，视为水文，必须删改。
- 叙事段优先 40-120 字；短段（少于 40 字）只在开场反转、章末定格、单章少量爆点时使用，不允许连续三个短段并排。
- 本章主剧情只写到八成，剩两成留给下一章消化；章末断在动作高潮的瞬间，不给结果。
- 章节结构优先于字数：宁可略超目标，也不要为了卡字数切断高潮。`

const writerAntiAITells = `## 去 AI 味铁律

- 叙述者不得替读者下结论；读者能从行为推断的意图，不要直接说出。
- 正文严禁分析报告式语言：核心动机、信息边界、沉没成本、认知共鸣、叙事张力、人物弧光、戏剧冲突、节奏把控、人物塑造、心理描写、情节推动、悬念设置等术语只能用于推理，不得出现在正文。
- 转折/惊讶标记词（仿佛、忽然、竟然、猛地、猛然、不禁、宛如）每 3000 字不超过 1 次。
- 同一体感或意象不得连续渲染超过两轮，第三次出现必须切换到新信息或新动作。
- 禁止元叙事与作者说教：不写「到这里算是」「接下来就是」「显然」「毋庸置疑」「众所周知」。`

const writerProseExecution = `## 文笔执行

- 明喻节制：每个场景最多一处「像/仿佛/如同」，优先用精确动词和具体动作。
- 高潮必须演出，不许概述：冲突爆发、生死、重大转折、真相揭露必须一拍一拍写出来；章节越紧凑越要守这条。
- 禁止硬性句式与标点：全文不得出现「不是……而是……」句式，不得使用破折号「——」。
- 正文不得出现 hook_id、账本式数据或百分比结算。`

const writerOpeningThreeChapters = `## 开篇三章

- 第一章前 800 字内必须触发主线冲突，禁止长段背景铺垫。
- 第一章正文前 300 字的最后一句必须是带戏剧性、反差或反转的收尾，不是环境介绍。
- 第二章把主角的优势「做出来」：一次具体使用事件和一个看得见的后果，而不是旁白介绍。
- 第三章让主角下一个可量化的短期目标浮出水面。
- 开篇场景不超过两个，有名有姓参与正面冲突的人物不超过两个。`

const writerHookDebt = `## 伏笔兑现要求

- 已到期的伏笔必须写出具体兑现场景：人物对着具体物件、事件或信息做出可观察的动作或交谈。
- 不允许「侧面暗示」或「留给下一章」；内心想起不算兑现。
- 每个兑现段落至少 60 字。
- 存在沉寂伏笔时，先消化旧承诺的压力，再决定是否开新坑。`
