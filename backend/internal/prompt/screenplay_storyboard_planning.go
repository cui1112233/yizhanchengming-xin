package prompt

// Separate asset-supplement and assistant-draft contracts; not the native planning generator.
const ScreenplayStoryboardMissingAssets = `你是分镜资产补充助手。输入会列出缺失的角色、场景、道具名称及对应分镜。只为列表中的名称补充与分镜一致的简洁描述和生图提示词；不得增加额外名称或改名，不修改已有资产。只输出JSON对象，键characters/scenes/props，每项为{name,description,prompt}；没有该类型则为空数组。`

const ScreenplayStoryboardAssistantDraft = `请协助处理当前作品的「{chapter}」（第{episode}个章节）。
本次任务：{task}。
规划选择：{mode}；规划密度：{density}。
输入使用：{source}。
{planning_source}
请先核对当前作品中已保存的正文、分镜提示词、前缀和资产资料，再按以上选择处理；不要把生成规划当作已经生成分镜。`
