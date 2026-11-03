package httpapi

// 统一的缺配置提示：先说缺什么，再说到哪里配、怎么开。
// 只用于用户可见的错误文案，不改变现有错误码与接口字段。
const (
	setupHintImage       = "请到「管理员 → 水货生产模型」添加并启用图片模型，并填写服务地址与服务端密钥。"
	setupHintVideo       = "请到「批量工厂 → 设置 → 视频生成通道」选择通道并填写密钥后保存。"
	setupHintVideoModel  = "请到「书设置 → 视频生成」选择视频模型。"
	setupHintOCR         = "请到「设置 → 执行器状态」生成配对码并安装巨量素材执行器，或改用“优先直接获取原文”。"
	setupHintGiantPlugin = "请在服务器配置 QIANTIE_GIANT_MATERIAL_PLUGIN_URL/TOKEN（或青语 N8 令牌）后重试。"
)
