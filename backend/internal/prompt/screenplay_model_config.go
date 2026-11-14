package prompt

// Recovered verbatim from model-config fragments 056-Aga and 057-zga.
// These defaults are server owned; explicitly saved empty text remains empty.
const ScreenplayImageCoreDefault = "虚幻引擎渲染效果、pbr材质，柔和的漫射光、 左侧 1/3 区域为超大高清面部特写，右侧 2/3 区域整齐排布角色三张全身三视图，包含角色的正面 (Front)、侧面 (Profile) 及背面 (Back) 三个维度的全身站姿视图。纯白背景，干净极简，无阴影，8K高清，高清细节，质感锐利"

const ScreenplayAssetProcessDefault = "请根据参考图，人物长相，人物长相不要改变，穿着不要任何改变，风格改成是【虚幻引擎渲染效果、pbr材质，柔和的漫射光、 干净极简，无阴影，8K高清，高清细节，质感锐利】，固定该人物已确认的人脸、眼型、五官比例、发型轮廓与体型。人物设定卡，浅灰渐变背景，白色圆角外框与分隔线，无顶部名称。上方为正面头肩肖像，眼部覆盖玫红色横条-横条不要透明注意，完全不透明的横条，右上嵌入同一人物双眼局部放大图，下方仅标注“角色眼部区域”；下半部为同一套服装正面、左侧面、背面完整全身三视图，三视图身体比例、服装颜色、材质、鞋型和发型保持一致，双臂自然下垂。风格避免真人摄影感；无场景、无手持物、无品牌标志、无水印。"

// Independently authored policy: the original western extraction backend was not recovered.
const ScreenplayWesternAssetExtraction = `按海外人物、场景和道具的文化语境提取资产。保留正文原有的姓名、族裔、服饰、建筑、地理、宗教和时代特征，不要本土化改写，不凭空添加东方化特征。输出结构与基础资产提取规则完全一致。`

// Independent compatibility rules for legacy image modes, not recovered template bodies.
func ScreenplayLegacyCharacterImage(mode string) string {
	switch mode {
	case "single":
		return "生成单个人物的完整全身照，画面只出现该人物一次，突出人物外貌及服装细节。"
	case "three_view":
		return "同一人物全身三视图：正面、侧面、背面；三个人物视图外貌、服装、比例保持一致。"
	case "three_view_realistic_sketch":
		return "将同一角色表现为真人写实的正面、侧面、背面全身三视图，保持线稿参考中的外貌、服装及比例一致。"
	case "all_around":
		return "人物全方位设定图，多角度展示同一人物的面部、全身和服装细节，各视角保持一致。"
	}
	return ""
}
