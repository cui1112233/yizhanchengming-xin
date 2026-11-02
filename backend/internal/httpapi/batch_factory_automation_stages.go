package httpapi

// 阶段契约单一来源：frontend/src/user/pages/shuihuo/automationStages.json 是
// 唯一真相文件，本文件的常量必须与它逐项一致，由 TestAutomationStagesContract
// 守护。前端时间线与后端自动化运行阶段共用同一份定义，任何改名/新增都必须
// 双侧同步，否则测试失败。
type automationStageDefinition struct {
	ID       string `json:"id"`
	Mode     string `json:"mode"`
	Label    string `json:"label"`
	Terminal bool   `json:"terminal"`
}

var automationStageContract = []automationStageDefinition{
	{ID: "source", Mode: "video", Label: "原文已就绪"},
	{ID: "assets", Mode: "video", Label: "资产（人物/场景/图片）已就绪"},
	{ID: "director", Mode: "video", Label: "文本提示词已就绪"},
	{ID: "video", Mode: "video", Label: "视频生成"},
	{ID: "upload", Mode: "video", Label: "上传网络"},
	{ID: "complete", Mode: "video", Label: "完成", Terminal: true},
	{ID: "source", Mode: "comic", Label: "原文已就绪"},
	{ID: "assets", Mode: "comic", Label: "漫画资产已就绪"},
	{ID: "director", Mode: "comic", Label: "漫画分页与提示词已就绪"},
	{ID: "video", Mode: "comic", Label: "漫画图片生成"},
	{ID: "complete", Mode: "comic", Label: "漫画合集", Terminal: true},
	{ID: "preflight", Mode: "runtime", Label: "预检"},
	{ID: "fetch_original", Mode: "runtime", Label: "获取原文"},
	{ID: "storyboards", Mode: "runtime", Label: "分镜生成"},
	{ID: "openings", Mode: "runtime", Label: "片头生成"},
	{ID: "audio", Mode: "runtime", Label: "音频生成"},
	{ID: "merge", Mode: "runtime", Label: "视频合成"},
	{ID: "submit", Mode: "runtime", Label: "定时提交"},
	{ID: "comic_assets", Mode: "runtime", Label: "漫画资产提取"},
	{ID: "comic_bootstrap", Mode: "runtime", Label: "漫画资产初始化"},
	{ID: "comic_images", Mode: "runtime", Label: "整本漫画出图"},
	{ID: "comic_collection", Mode: "runtime", Label: "漫画合集整理"},
	{ID: "comic_publish", Mode: "runtime", Label: "漫画发布"},
	{ID: "upload_confirm", Mode: "runtime", Label: "已上传待确认"},
	{ID: "published", Mode: "runtime", Label: "发布完成", Terminal: true},
	{ID: "completed", Mode: "runtime", Label: "已完成", Terminal: true},
}
