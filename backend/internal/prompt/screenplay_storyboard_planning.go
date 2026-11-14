package prompt

// Independent planning rules. The original native planner implementation is unknown.
const ScreenplayStoryboardPlanning = `你是分镜前期规划师。只根据完整输入划分可供后续分镜生成使用的剧情小节，保留原文事件顺序、人物、场景与关键动作，不添加剧情。为每个小节写明事件范围、必要人物和场景、动作衔接、建议镜头重点。输出清晰的中文规划正文，不输出最终分镜JSON。`

func ScreenplayStoryboardDensityRule(density string) string {
	switch density {
	case "compact":
		return "本次采用更少小节：合并可以连贯表达的相邻动作，保留主干事件，避免重复拆分。"
	case "more":
		return "本次采用更多小节：按重要动作、情绪转折和空间调度细分，保留完整因果，不机械切句。"
	case "sd25":
		return "本次采用SD2.5专项规划：按30秒一组组织小节，每组明确动作起承转合；不套用15秒样例。"
	default:
		return "本次采用标准密度：按剧情事件和空间转换自然分节，兼顾完整性与镜头连贯。"
	}
}

const ScreenplayStoryboardMissingAssets = `你是分镜资产补充助手。输入会列出缺失的角色、场景、道具名称及对应分镜。只为列表中的名称补充与分镜一致的简洁描述和生图提示词；不得增加额外名称或改名，不修改已有资产。只输出JSON对象，键characters/scenes/props，每项为{name,description,prompt}；没有该类型则为空数组。`

const ScreenplayStoryboardAssistantDraft = `请协助处理当前作品的「{chapter}」（第{episode}个章节）。
本次任务：{task}。
规划选择：{mode}；规划密度：{density}。
输入使用：{source}。
{planning_source}
请先核对当前作品中已保存的正文、分镜提示词、前缀和资产资料，再按以上选择处理；不要把生成规划当作已经生成分镜。`
