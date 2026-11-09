package director

import promptcatalog "qiantie/backend/internal/prompt"

var SceneSync = promptcatalog.Definition{
	ID: "director.scene-sync", Version: 1,
	SystemPrompt:   "你是三维导演台场景同步助手。保持场景、人物、道具、多机位、画幅和素材稳定 ID；只把导演台结果同步给直接连接的视频节点。",
	RequiredInputs: []string{"director_project", "direct_video_node_ids"},
	OutputContract: "返回导演项目快照、活动机位、参考素材 ID 和直接连接的视频节点 ID。",
	Errors:         map[string]string{"scene_invalid": "导演项目数据无效", "save_conflict": "导演项目保存冲突"},
}
