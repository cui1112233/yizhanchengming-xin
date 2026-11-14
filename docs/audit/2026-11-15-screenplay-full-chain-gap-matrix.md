# CM漫剧全流程差距总表（重新核查）

结论：没有全部写完，也没有全部对齐。之前“按钮存在、测试通过”的汇报只能覆盖限定功能；不应当作为全软件闭环证明。

范围是已发布原版新编辑器的三阶段及二级窗口。439条原点击模板含隐藏分支，并不等于439个当前按钮。案卷full-chain-20261115/original-button-verification.csv保留每条源动作、行列和条件；尚未逐条建立运行验收映射，不把它们标记完成。

证据等级：本地数据闭环（已实际保存重读）、已接入但模型未实跑、部分实现、未完整实现、本轮补通：参数/模拟验证，待真实模型。没有供应商预算就明确未实跑，不用模拟结果冒充真实服务输出。

| 阶段 | 功能/链路 | 当前状态 | 前端证据位置 | 后端 | 缺口或验证限制 |
|---|---|---|---|---|---|
| 作品库 | 新建、打开、搜索、归档、恢复 | 已闭环：本地数据 | frontend/src/user/pages/ComicScreenplayPage.jsx:207 | backend/internal/screenplay/manual_document.go | 版本与归档已做实际HTTP；永久删除未实删 |
| 输入 | 直接输入原文、TXT/DOCX读取保存 | 已实现，部分实际验收 | frontend/src/user/pages/ComicScreenplayPage.jsx:8 | backend/internal/httpapi/screenplay_routes.go | 保存原文不等于已经生成章节或剧本 |
| 输入 | 原文建立章节，不先生成软件剧本 | 本轮补通：参数/模拟验证，待真实模型 | frontend/src/user/pages/ComicScreenplayStoryboardStage.jsx:27 | backend/internal/httpapi/screenplay_authoring_routes.go | 之前只有剧本列表；新增显式建立原文章节，不虚构剧本 |
| 剧本 | 选择提示词/画风/自己模型进入最终请求 | 已接入：模型未实跑 | frontend/src/user/pages/ComicScreenplayPage.jsx:8 | backend/internal/screenplay/screenplay_service.go | 已有参数测试与实际选择保存；不可当真实生成对照 |
| 剧本 | 逐章生成、手工编辑、版本恢复 | 部分闭环 | frontend/src/user/pages/ComicScreenplayAuthoringPanels.jsx:5 | backend/internal/screenplay/authoring_service.go | 手工保存/恢复实际验收；逐章AI生成只模拟 |
| 剧本 | AI故事摘要、标签与原版全部字段 | 部分实现 | frontend/src/user/pages/ComicScreenplayStoryStage.jsx:6 | backend/internal/screenplay/screenplay_service.go | 展示已有字段；不等于原版所有生成规则相同 |
| 模型配置 | 比例、分辨率、模板、核心指令、独立生图改图 | 已闭环：配置数据 | frontend/src/user/pages/ComicScreenplayModelConfig.jsx:9 | backend/internal/screenplay/creative_config.go | 真实配置读回与清空/冲突验证；模型用自己的目录 |
| 资产 | 人物/场景/道具增删改与解析程度 | 已闭环：数据；AI待实跑 | frontend/src/user/pages/ComicScreenplayAssetStage.jsx:13 | backend/internal/screenplay/assets_service.go | 手工CRUD真实落库；AI资产解析只做模拟 |
| 资产 | 角色设计六类字段及提示词保存 | 已实现，参数已验收 | frontend/src/user/pages/ComicCharacterDesignDialog.jsx:8 | backend/internal/screenplay/assets_service.go | 选择保存真实字段；不以字段数冒充图像质量 |
| 资产 | 批量遗漏资产检查 | 部分实现 | frontend/src/user/pages/ComicScreenplayBatchTools.jsx:30 | backend/internal/screenplay/assets_service.go | 只补缺失保旧卡；顺序执行，离页后不继续后续项 |
| 资产 | 检查任务后台持续、关闭重开续看 | 未实现于检查入口 | frontend/src/user/pages/ComicScreenplayBatchTools.jsx:15 | backend/internal/screenplay/production_runtime.go | 生产后台运行不等于检查入口拥有后台任务 |
| 资产 | 跨项目普通资产库、导入到当前库 | 未完整实现 | frontend/src/user/pages/ComicScreenplayAssetStage.jsx:13 | backend/internal/screenplay/assets_service.go | 音色跨项目复制不等于通用资产库 |
| 资产 | 多视图生成与各视图独立管理 | 未完整实现 | frontend/src/user/pages/ComicScreenplayAssetDialog.jsx:19 | backend/internal/screenplay/media_store.go | 现有角色模板不等于原版多视图操作链 |
| 图片 | 自己模型、生图/改图、参考图和选中图 | 已接入：本轮未实跑供应商 | frontend/src/user/pages/ComicScreenplayAssetDialog.jsx:19 | backend/internal/screenplay/media_store.go | 模型选择/参数/重试保护有测试，真实图像输出需另验 |
| 图片 | 批量图片、单项失败恢复 | 部分闭环 | frontend/src/user/pages/ComicScreenplayAssetBatchDialog.jsx:14 | backend/internal/screenplay/media_store.go | 有逐项状态和恢复；不冒充原版并行算法 |
| 站位 | 新增编辑、分析、参考资产、图片保存 | 部分闭环 | frontend/src/user/pages/ComicScreenplayPositioningStage.jsx:12 | backend/internal/screenplay/media_store.go | 结构/图片绑定与标注有检查；分析和出图待模型实跑 |
| 站位 | 画布标注、编辑图片、保存标注PNG | 已实现：部分本地文件验收 | frontend/src/user/pages/ComicScreenplayPositioningStage.jsx:12 | backend/internal/screenplay/media_store.go | PNG有离线浏览器验证；不等于原版所有画布手势 |
| 音色 | 上传、真实库选择、试听与保存 | 部分闭环 | frontend/src/user/pages/ComicVoiceSwitchDialog.jsx:12 | backend/internal/screenplay/media_store.go | 原音色设计专用服务未配置时禁用，不能用普通TTS冒充 |
| 分镜设置 | 收藏/最近/前缀、四密度和两个开关 | 已闭环：设置数据 | frontend/src/user/pages/ComicScreenplayStoryboardSettings.jsx:12 | backend/internal/screenplay/storyboard_preferences.go | 偏好MySQL实读写与模拟最终请求；非原native算法证明 |
| 分镜规划 | 新规划/已有规划/无规划、来源绑定 | 已接入：模型未实跑 | frontend/src/user/pages/ComicScreenplayStoryboardPrep.jsx:5 | backend/internal/screenplay/storyboard_planning.go | 有独立规划与来源摘要/版本/恢复；收费输出待实跑 |
| 分镜 | 章节原文到指定集号、无剧本单集生成 | 本轮补通：参数/模拟验证，待真实模型 | frontend/src/user/pages/comic-screenplay-storyboard-input.js:6 | backend/internal/screenplay/storyboard_preparation.go | 之前前端默认第一集；改为真实chapterId+目标号，后端读取原文 |
| 分镜 | 无剧本批量原文章节 | 本轮补通：参数/模拟验证，待真实模型 | frontend/src/user/pages/ComicScreenplayBatchTools.jsx:11 | backend/internal/screenplay/storyboard_preparation.go | 之前列表为空；改从真实原文章节构建批量目标，保留原顺序号 |
| 分镜 | 按集持久请求、回包丢失核对和续跑 | 已接入：模拟与数据验证 | frontend/src/user/pages/comic-screenplay-storyboard-attempts.js:16 | backend/internal/screenplay/storyboard_generation.go | 按集幂等/partial/unknown有测试；真实上游断网需另验 |
| 分镜 | 逐场景并发/失败场景自动一次重试 | 未完整实现 | frontend/src/user/pages/ComicScreenplayStoryboardPrep.jsx:29 | backend/internal/screenplay/storyboard_service.go | 当前每集一次模型调用，不是原版逐场景调度 |
| 分镜 | 失败场景列表、单场景/全部失败重试 | 未实现同粒度 | frontend/src/user/pages/ComicScreenplayBatchTools.jsx:58 | backend/internal/screenplay/storyboard_generation.go | 当前失败记录按集，不能说等同原版场景级 |
| 分镜 | 原文→场景自动识别→逐场景生成 | 部分/待补 | frontend/src/user/pages/ComicScreenplayStoryboardStage.jsx:24 | backend/internal/screenplay/storyboard_service.go | 本轮补原文直接按集；场景拆分策略与独立场景任务未完整接通 |
| 分镜编辑 | 新增编辑、拆分合并、批量时长 | 已闭环：本地数据 | frontend/src/user/pages/ComicScreenplayPage.jsx:11 | backend/internal/screenplay/storyboard_editor.go | 真实17组HTTP及页面保存版本递增已验收 |
| 分镜检查 | AI衔接Agent独立报告与后续修正 | 未完整对齐 | frontend/src/user/pages/ComicScreenplayProductionPanel.jsx:11 | backend/internal/screenplay/production_doctor.go | 生产质检不等于原版独立衔接检查窗口 |
| Agent | 预填草稿、发送、方案应用 | 部分实现 | frontend/src/user/pages/ComicScreenplayAuthoringPanels.jsx:36 | backend/internal/httpapi/screenplay_authoring_routes.go | 预填不自动发送；只支持已有动作范围，不是全Skills |
| Agent | 原版全Skills、原生工具、复杂调度 | 未完整实现 | frontend/src/user/pages/ComicScreenplayAuthoringPanels.jsx:36 | backend/internal/httpapi/screenplay_authoring_routes.go | 不能以一个对话入口声称原版所有服务已接 |
| 媒体 | 图片/音频/视频/字幕面板及文件保存 | 部分闭环 | frontend/src/user/pages/ComicScreenplayMediaPanel.jsx:13 | backend/internal/screenplay/media_store.go | 文件/参数/状态有检查；真实供应商制作未全跑 |
| 媒体 | TTS、ASR、视频生成轮询与失败处理 | 已接入：真实模型未全验收 | frontend/src/user/pages/ComicScreenplayMediaTools.jsx:6 | backend/internal/screenplay/media_store.go | 依赖真实配置，未配置项不能算缺UI或假可用 |
| 交付 | 时间线建立并进入AI-CUT | 部分闭环 | frontend/src/user/pages/ComicScreenplayMediaPanel.jsx:8 | backend/internal/screenplay/media_store.go | 数据构建与导航有测试；完整素材到可导出成片仍需真实验收 |
| 交付 | 媒体包、剪映兼容草稿与成片下载 | 部分闭环 | frontend/src/user/pages/ComicScreenplayMediaPanel.jsx:4 | backend/internal/screenplay/media_store.go | 本地文件/模拟检查存在；当前剪映真实导入未完成 |
| 自动生产 | 九步骤、质检、暂停、定时、重建、恢复 | 部分闭环 | frontend/src/user/pages/ComicScreenplayProductionPanel.jsx:11 | backend/internal/screenplay/production_runtime.go | 持久调度有测试，真实整片供应商全流程没有本轮验收 |
| UI | 布局、图标、弹窗、原版像素对比 | 部分对齐 | frontend/src/user/pages/ComicScreenplayIcon.jsx（仅模块边界） | backend/internal/screenplay/creative_config.go | 已引入原图标与若干原结构；截图超时范围和所有编辑器仍缺 |

## 先补的真实断链
原文已经保存且没有软件剧本时，前端分镜目标只读document，导致章节原文不能进入单集/批量入口。调用方还把不存在目标回退到第1集，可能将第二章写入第1集。后端本来支持chapterId绑定原文，因此这不是缺按钮，而是列表/输入/目标编号未对接。本轮修复这三处，并补显式从已存原文建立一个章节的入口；它不会伪造软件剧本，也不假装AI已经自动分集。

## 下一阶段仍需要的工作
1. 逐场景任务和真实失败场景恢复；不能继续以按集失败记录替代。
2. 检查入口独立后台持久任务、关闭重开/暂停/恢复；不能借生产队列存在就宣布检查后台完成。
3. 通用跨作品资产库、多视图、AI衔接检查与完整Agent Skills逐项实现和验收。
4. 用实际可用模型验证原文到资产、分镜、音视频、时间线与最终文件；单独记录次数、费用和输出。本轮没有新收费请求授权。

## 本轮回归记录

原文路径新增4项纯逻辑检查、完整Page原文建立/取消/第二章单集与批量/原文分镜编辑检查通过，原有阶段与准备窗口20项相关浏览器检查通过。Go专项用假模型确认没有软件剧本时按chapterId读取第二章真实原文，只把分镜写到第二集，不创建假剧本。

完整前端155项首轮154通过，唯一失败是Chrome网络进程崩溃导致音色确认生命周期测试未完成；该用例独立重跑通过。Go screenplay/storage整包通过，httpapi整包唯一失败为无关NovelFetch后台启动计时（0.39秒），独立重跑通过。没有修改这些无关模块或放宽断言。

测试期间默认Go模块缓存有依赖文件消失，改用本项目独立缓存和已验证可达的包镜像后恢复；未清理全局缓存或关闭校验。没有新供应商调用。首轮失败日志与重跑记录分别保留，不把重跑说成首轮全绿。
