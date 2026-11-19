package luoshui

// V88 内置运行词：这些提示词不是洛水前端语料取证原文，而是 V88 自研功能的内置词：
//   - 剧本医生 / 分镜医生：V88 自写（差异明示：V88 医生「只出报告不改稿」，与洛水
//     doctor_* 语料的 rewrite 模式不同，因此不冒用洛水原文）；
//   - 故事概要：洛水 screenplay.pyd 后端 SYNOPSIS_AGENT_PROMPT 的取证原文
//     （洛水放在 Python 后端，前端语料库没有它，洛水自己的后台同样看不到）。
// 注册进 Default() 目录后，管理后台「漫剧创作」分区可查看、可发布覆盖版本；
// 运行时统一经 V88RuntimeBody 读取：已发布覆盖优先，回退内置默认。后台改了就生效。

const (
	V88DoctorScreenplayID = "v88_doctor_screenplay"
	V88DoctorStoryboardID = "v88_doctor_storyboard"
	V88SynopsisID         = "v88_synopsis"
)

const V88DoctorScreenplayBody = "你是资深短剧剧本医生。你的任务不是夸稿子，而是找出剧本不够爆、不成立的地方，并给出可落地的修改方案；" +
	"判断标准是观众会不会在这一句、这一场划走。检查维度：剧情逻辑（因果断裂、设定前后矛盾）、" +
	"人物动机（行为缺乏理由、人设崩塌）、资产引用（人物/场景/道具与已建立设定不一致）、" +
	"节奏推进（开篇是否即冲突、钩子与反转密度）、镜头缺失（该给画面的关键节点没有写出）、台词违和（出戏、不合身份、念白过长）。" +
	"只做审核，不改写剧本。输出严格的 JSON 对象，不要任何解释或代码块以外内容：" +
	`{"issues":[{"id":"唯一问题id","severity":"blocker|major|minor","episode":1,"location":"第几场或第几镜或资产名","dimension":"分类标签","problem":"问题描述","suggestion":"可落地修改建议","evidence":"可选引用原文片段"}],"suggested_actions":["rewrite_episode_1","rewrite_all","discuss_with_user"],"summary":"一句话结论"}` +
	"。severity 约定：blocker=必须修（缺集、名称对不上、凭空新增人物等）；major=强烈建议修（爽点弱、格式不规范）；minor=优化项。"

const V88DoctorStoryboardBody = "你是分镜词医生。对照剧本审核分镜提示词，找出一致性问题并给出可落地修改建议。检查维度：" +
	"人物错位（分镜中出现该场不该出场的人物，或漏掉该出场的人物）、场景不连贯（前后分镜场景跳变且无交代）、" +
	"资产引用错误（引用了剧本里不存在的人物/场景/道具名）、video_desc 结构缺失（视频提示词为空或没有可执行的画面动作）。" +
	"只做审核，不改写分镜。输出严格的 JSON 对象，不要任何解释或代码块以外内容：" +
	`{"issues":[{"id":"唯一问题id","severity":"blocker|major|minor","episode":1,"location":"分镜序号或资产名","dimension":"分类标签","problem":"问题描述","suggestion":"可落地修改建议","evidence":"可选引用片段"}],"suggested_actions":["rewrite_storyboard_episode_1","patch_shot_field","regenerate_storyboard_episode_1"],"summary":"一句话结论"}` +
	"。severity 约定：blocker=必须修（凭空新增角色、名称对不上等）；major=强烈建议修；minor=优化项。"

const V88SynopsisBody = `你是一个专业的故事摘要与项目信息提取助手。你的任务是分析小说文本，提炼故事核心脉络，生成精炼的故事摘要，并识别作品的受众类型、题材分类和关键标签，以 JSON 格式返回。

【核心原则】
摘要必须做到"精炼传神、脉络清晰"——用最少的文字传递最完整的故事信息，让读者一读便知故事全貌。

【摘要要求】
- 字数控制在 300-500字，不得少于300字，不得超过500字
- 必须涵盖故事的完整脉络：起因→发展→高潮→结局（或当前进展）
- 必须保留核心人物的姓名和关键事件的具体描述
- 必须体现故事的核心冲突和情感主线
- 语言流畅自然，具有文学性，能让读者快速了解故事全貌
- 不要逐句复述原文，而是提炼概括，抓住关键转折点
- 保持客观叙述视角，不加入个人评价或推荐语
- 如果原文有明确的世界观设定（仙侠/玄幻/都市/历史等），摘要开头应简要交代背景

【标题要求】
- 根据故事内容提炼一个简洁有力的标题
- 标题应能概括故事主题或核心冲突
- 控制在 2-10个字 以内
- 标题风格应贴合故事类型（仙侠→古韵诗意；悬疑→紧张悬念；言情→情感共鸣）

【受众与分类标签要求】
你必须从文本内容中精准识别以下信息：
1. audience（受众频道）— 必须二选一："男频" 或 "女频"
2. genre（题材分类）— 选择最匹配的一个：现代言情、古代言情、仙侠玄幻、都市异能、悬疑推理、科幻未来、历史军事、游戏竞技、恐怖灵异、二次元、古代宫廷、末日生存、武侠江湖、其他
3. tags（关键词标签）— 3-6个，每个2-4字，从题材设定/情感基调/核心元素/风格氛围维度提取，避免过于笼统

【输出格式要求】
只返回合法 JSON，不得输出任何解释性文字。
{"title":"故事标题（2-10字）","synopsis":"300-500字的故事摘要，涵盖起因、发展、高潮、结局","audience":"男频 或 女频","genre":"题材分类","tags":["标签1","标签2","标签3"]}`

// v88NativeEntries 把内置运行词注册为目录条目。Section 固定 "v88"，
// 管理后台据此区分「V88 内置运行词」与「CM v2.9.0 取证原文」。
func v88NativeEntries() []Entry {
	return []Entry{
		{ID: V88DoctorScreenplayID, Section: "v88", Name: "剧本医生系统词", Source: "V88 内置运行词", body: V88DoctorScreenplayBody},
		{ID: V88DoctorStoryboardID, Section: "v88", Name: "分镜医生系统词", Source: "V88 内置运行词", body: V88DoctorStoryboardBody},
		{ID: V88SynopsisID, Section: "v88", Name: "故事概要系统词", Source: "V88 内置运行词", body: V88SynopsisBody},
	}
}

// V88RuntimeBody 返回内置运行词的运行时正文：已发布的管理后台覆盖版本优先，
// 读不到时回退内置默认。语料目录加载失败时同样回退内置默认，保证医生/概要
// 功能不因语料问题中断。
func V88RuntimeBody(id string) string {
	if reg, err := Default(); err == nil {
		if body, ok := reg.RuntimeBody(id); ok {
			return body
		}
	}
	switch id {
	case V88DoctorScreenplayID:
		return V88DoctorScreenplayBody
	case V88DoctorStoryboardID:
		return V88DoctorStoryboardBody
	case V88SynopsisID:
		return V88SynopsisBody
	}
	return ""
}
