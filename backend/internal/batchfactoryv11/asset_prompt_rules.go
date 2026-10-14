package batchfactoryv11

// staticAssetPromptRule keeps reusable asset cards separate from storyboard
// actions. The same rule is used by the one-call extractor and H3 appearance
// compiler so a later appearance pass cannot reintroduce camera actions.
const staticAssetPromptRule = `
【资产提示词边界】
人物、场景、道具提示词只写静态视觉锚点，而且这些锚点必须可跨镜头复用；不得写镜头动作、剧情动作、对白或一次性情绪表演。
人物必须详细展开描述身份、时代、性别、年龄阶段、脸型五官、肤色与皮肤细节、发型、体态基准、服装材质与颜色、固定配饰和稳定气质，至少覆盖这些维度，不能只返回“古装女性”“漂亮女人”这类短标签；人物资产不得描述坐在哪里、手放在哪里、正在做什么、临时表情、镜头景别、景深、电影感、光晕或一次性室内光线。场景只允许描述空间、时段、陈设、材质、光线和固定氛围；道具只允许描述材质、形状、颜色、状态和固定细节。
不得写“端坐”“抬手”“走动”“转身”“凝视”等镜头动作，也不得写拿取、递交、放下、说话、奔跑、进入、离开、被踩碎等剧情动作。不要把动作词包装成外形提示词。若需要表达稳定姿态，只写“肩背平直、体态端庄、站姿基准稳定”等静态体态锚点。
动作、表情变化、场景内位置、临时光影、景别、景深、运镜、时长和剧情结果只能出现在分镜视频提示词中。人物资产中不得出现“端坐于雕花红木太师椅上”“双手搭在椅扶”“室内微黄光晕勾勒面部轮廓”“明暗层次丰富”“电影级景深”“神情深不可测”这类镜头化句子。`
