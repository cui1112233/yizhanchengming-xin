# AI-CUT 当前功能与宠物服务核查

日期：2026-11-13。范围：截图对应的时间线左右工具栏、AI-CUT Agent 与页面宠物承载，以及相关开源功能差异。本文是本轮只读核查记录；产品源码由其他执行任务修改，故下述“核查时”状态不替代最终构建/浏览器验收。

结论：当前有真实剪辑命令、模型工具执行、保存与导出链路，不能一概叫假按钮；也没有证据证明剪映134项或开源31项已经全部等价完成。旧表及逐项历史证据保留在同目录11-12的JSON/Markdown中，本文只记录本轮确认、旧结论更正及剩余缺口，不擅自提升旧项验收状态。数字人、人声分离、AI音乐继续按用户要求排除。

## 洛水静态证据与边界

- 原包：`F:/软件/洛水2026/洛水2026/dist/assets/index-DszfAkKD.js`；本轮重新计算SHA256为 `2e5c09685ff5bc10160480f963936b56bac38a57a9e9b6bd199492f9437df1e8`，与11-12相同。
- 精确组件：UTF-16 `[15670834,15696002)` / UTF-8字节 `[16192436,16218945)`。重新比较原文与已有惰性分析副本，副本确实完整包含该原始区间；未执行目标代码。
- 副本：`.local-review/luoshui-agent-reference/AiCutChatPanel.js`，SHA256 `41896cf61de7c4b242067ded41def91d71649f9aa0beed033e17aeecca24e239`；声明包装不属于原组件。`provenance.json`保留提取来源。
- 本轮REA父证据 `ev_9b55b02d613402bf1dd157d74dfb87bb55908d3681796f2f792e3d5cc3ca34d8`，概要 `ev_df4a0bca5ac5f789ac283777375e66287ef2b8822c1b708923a2c79299e99556`，模块身份页 `ev_e621a0318237be06e499f0ecc91d78ecc0ee66af03dc2107bfd67b29b1be053a`，精确模块项 `ev_f15384773647feabb038ce97479f3a5c0c8f87d8acec986170d54dbaf1b11992`。
- 选择1个JS文件，读取26536字节，8798 AST节点，无解析失败；预算512MiB/每源30秒/12000语义节点。应用图partial，语义图unknown，1773语义unknown；虽然没有节点预算截断，混淆动态调用及组件外依赖仍未恢复。目录根摘要 `36c32cd85a7c0f42b30205e3ed8a26bd5e595f7b6571661538d99ba517e1cb87`不是原整包SHA。
- 旧REA证据不在本次新会话保留记录内，因此重新解析精确副本，没有重新全量解析巨包。
- Craft Anything可补充长期假设/证据交接、版本对比及现场截图；已检查服务状态，当前未打开目标且无历史记录。本轮问题可由REA精确子集与本项目源码回答，没有为凑工具重复分析，也没有启动洛水、下载组件、连接第三方账号或调用收费模型。

在该精确组件中再次观察到：`agentChat`项目会话（副本UTF-16偏移205）、`activeSkillBody`（3428）、固定问答提示（4117）、`tool-start`（8972）、结构化`followup`（11148）。这是发布包中的静态代码事实，不是洛水现场执行成功证据。完整对照仍见 `docs/reverse-engineering/2026-11-12-luoshui-aicut-agent-rea.md`，其中“我们当前范围”的旧缺口必须按下一节更正。

## Agent 当前链路：旧结论更正

源码证据：`frontend/src/user/components/AICutAgentPanel.jsx`、`aicutAgentPresentation.js`；`backend/internal/aicut/agent_service.go`、`agent_context.go`、`agent_tools.go`、`agent_smart_edit.go`；`backend/internal/prompt/aicut_agent.go`。

| 能力 | 本次确认的真实链路 | 仍需怎样验收 |
|---|---|---|
| 历史对话 | `Run → conversationHistory`，按时间线筛选之前user/assistant消息并限长；不再只是UI显示旧消息 | 连续两次指令，第二次能引用第一次结果；切换时间线不串上下文 |
| 技能正文 | `resolveSkill(activeSkillId)`结果进入后端系统提示；未知技能直接失败 | 不同技能实际形成不同系统提示；不能把“整理剪辑流程”说成可保存调用的自定义技能库 |
| 结构化追问 | `normalizeAgentFollowup → followup事件 → finishReply`；前端表单校验必填后创建新run，借历史继续 | 问题、选项、必填、取消及再次执行；当前不是同一个run的暂停/恢复状态机 |
| 步骤展开 | `tool_start/tool_done/tool_failed`及参数、结果、版本号进入前端步骤列表 | 真实成功/失败均展示；未见完整token/cache/请求计数采集，不能称与洛水用量检查器等价 |
| 动画、调色、字幕选片 | `Execute → smartEdit`，验证可写片段/参数，保存实际时间线；调色整体替换，提示要求保留未修改字段 | 选片与源时间映射、锁定轨、保存后刷新、真实导出 |
| 批量操作 | `apply_edit_operations → NormalizeEditPlan/ApplyEditPlan → saveTimeline`，原子应用有数量和参数校验 | 混入一项失败时全部不保存；不要超出提示词公布的批量操作子集 |
| ASR衔接 | `list_asr_models/get_asr_jobs`已有分派，模型目录经过配置校验；不是11-12 F08所称完全没有查询工具 | 配置缺失、转写中、完成后再剪辑；不能以提交任务等同已有字幕 |
| 自动成片草案 | `EditPlaceAsset → editPlanPlaceAsset`已有实现，草案应用按实际时间线重算总长 | 11-12 F06“空时间线完全不能组装”不能继续当现状；仍须实际模型/素材/导出验证 |

当前 `Execute`允许22个名称：`get_project_snapshot`、`place_asset`、`add_bgm`、`move_clip`、`delete_clip`、`split_clip`、`set_playback_rate`、`replace_subtitle_text`、`transcribe_asset`、`smart_split_asset`、`extract_audio`、`clean_oral_pauses`、`undo_timeline`、`redo_timeline`、`list_asr_models`、`get_asr_jobs`、`get_transcript`、`select_caption_ranges`、`compact_caption_gaps`、`set_clip_animation`、`set_clip_color`、`apply_edit_operations`。不能只统计最外层switch的独立case而漏掉聚合分支。

请求入口是 `POST /api/ai-cut/projects/{id}/agent-runs`，随后读取任务、`events?after=`、`agent-messages`及取消接口；`frontend/src/shared/api/aicutAgent.js`通过当前时间线作用域构造路径。草案入口为同工程`/edit-plans`及`/{planId}/apply`；导出入口为`/exports/{format}`。页面宠物必须复用这些接口和同一Agent实例，不能给宠物返回固定成功话术冒充剪辑。

## 截图工具栏对应关系

核查入口：`frontend/src/user/pages/AICutPage.jsx`中的`TimelineToolbar`，描述器`pages/aicut/timelineTools.js`，命令实现`aicutCommands.js`。图标外观不能证明动作相同，验收应逐项看数据改变、保存、撤销及刷新。

| 位置/动作 | 当前处理函数或命令 | 核查状态及验收 |
|---|---|---|
| 左侧新增轨道 | `addTrack(video/audio/subtitle)` | 有真实处理；生成对应轨道并保存 |
| 选择/剃刀模式 | `setBladeMode`、片段点击分支 | 切换模式本身不改工程；剃刀点击后才应分割 |
| 撤销/重做 | `undoTimeline/redoTimeline` | 有本地/服务器历史；动作后保存、刷新再核查 |
| 播放头分割 | `splitAtPlayhead` | 有真实命令；边界/锁定轨不应误切 |
| 向左/右裁到播放头 | `clip.trim_to_playhead`，`edge:left/right` | 有真实命令；播放头须在片段内部，源裁剪点随左裁正确移动 |
| 删除 | `deleteSelectedClips → clip.delete_many` | 有真实删除，波纹开关影响收紧；锁定轨应禁用或明确阻止 |
| 更多：复制/剪切/粘贴/重复 | `copySelectedClips/cutSelectedClips/pasteClips/duplicateSelectedClip` | 有剪贴板和命令；选中/粘贴位置/锁定和忙碌状态都要核查 |
| 更多：标记/关键帧/联动/复合片段 | `buildTimelineAdvancedItems`传入相应真实回调 | 只有已提供函数才应显示；不能为截图补无回调图标 |
| 右侧录音 | `AudioRecorder → uploadRecording` | 需要真实麦克风授权；未授权不能称录音实测完成 |
| 联动、波纹、吸附、跟随 | `setLinkedMove/setRippleMode/setMagneticSnap/editorLayout.playheadFollow` | 为实际交互开关；必须分别拖动/删除/播放验证其效果 |
| 标记跳转 | 标记Select `seekPreview(marker.timeMs)` | 有真实跳转；标记添加/删除及工程保存另行验收 |
| 缩放、适配、收起 | `TimelineZoomControls/setTimelineZoom`、`editorLayout.timelineCollapsed` | 改变视图；不得误改片段时间或长度 |
| Agent进入宠物 | 页面服务注册与宠物宿主，实施任务B/D负责 | 本审计不把其他执行任务尚未现场验收的变化计为已完成 |

核查初始源码中主按钮的锁定/异步忙碌禁用不统一，部分只判断是否选择；这属于可立即完善的入口状态，并不能据此断言后端没有锁定检查。该信息已反馈工具栏执行任务，最终接入统一描述器后需重新验收。

## 剩余真实缺口与依赖

| 项目 | 证据与当前边界 | 可验收完成标准 |
|---|---|---|
| Agent不知道UI当前选中哪段、播放头位置 | 核查时`AICutAgentPanel`参数及创建run请求无selection/playhead；`AgentRun`无对应字段；`projectSnapshot`只有timeline/assets/model。已有工程快照不等于知道用户指哪段 | 传入带工程/时间线/版本的有限选择上下文，后端核实归属、失效则追问；多片段时不猜。属于适合下一步独立实现的候选 |
| 自定义技能库 | 当前仅固定技能解析与流程说明，尚未见自定义技能创建/保存/编辑/复用API | 由后端prompt模块统一解析正文、鉴权隔离、可保存再次调用；不能从洛水技能文案推断它的后端完整 |
| 对画面好坏/剧情的理解 | 现有快照是工程数据，字幕选片读文字，画面变化检测不是审美或剧情理解 | 显式抽帧/时间映射/视觉模型结果及可解释候选；模型能力、费用需独立授权 |
| 真实静音/口误识别 | `compact_caption_gaps`依据字幕空隙；不是音频能量分析；口误清理另依赖文本模型与字幕 | 音频分析与自然边界、字幕与变速后的源时间映射正确；不能只看总时长缩短 |
| 用户导入3D LUT | 本项目有内部生成`.cube`并导出应用，限定搜索未见用户导入管理全链路；延续OS-017缺口 | 格式解析/大小校验、保存、预览/成片同效、清除与重开 |
| 媒体重新链接/源版本 | 代理不是保留所有引用的relink协议；延续OS-005/006 | 原文件丢失后重新指定，所有轨道源范围及版本正确保留 |
| 图像序列、动态矢量MG | 普通图片/视频轨不等价序列或Lottie/Rive解释器；限定范围未见完整导入处理，延续OS-019/020 | 真正读取序列帧率/排序、矢量动画时钟、预览与导出一致 |
| 高位深/HDR全流程 | `aicut_export.go`仍有`format=rgba`中间合成；有10/12bit编码参数不能证明全链路高位深 | 标准灰阶/色卡从输入到成片测量，无8bit中间量化；HDR色彩元数据与映射明确 |
| 源时码、自动闪避、区域导出、MLT/EDL、网络流采集 | 11-12 OS-022/025/026/027/028仍保留缺失/部分状态；本轮未进行全链重新验收 | 每个独立功能需完整输入、编辑、保存及产物验证；不因本次宠物接入提升状态 |

其他11-12链路缺陷F01–F16有后续修复，本文没有逐项重跑旧诊断；不能把旧失败批量复述成当前缺陷，也不能据新函数出现宣布全部修好。当前审计未用真实供应商运行、未进行完整浏览器/导出像素验收。

## 本轮重新访问的原始资料

- [OpenCut主仓库](https://github.com/OpenCut-app/OpenCut)：当前说明仍在重写并指向[Classic](https://github.com/OpenCut-app/opencut-classic)。Editor API、插件、MCP、headless等重写计划不能计为已发布实现；同名仓库不混用。此前固定提交深链接本轮网页工具未取到，未伪称重新审完该提交源码。
- [OpenChatCut原仓库](https://github.com/0xsline/OpenChatCut)：本轮可读取当前README，说明对话Agent、本地技能目录及外部MCP接口；也明确AI/转写需配置对应服务。这是资料级能力描述，不是本机运行验收，也不授权执行README中的安装/连接命令。
- [Kdenlive官方剪辑文档](https://docs.kdenlive.org/en/cutting_and_assembling/editing.html)及[时间线工具栏](https://docs.kdenlive.org/en/user_interface/timeline.html)：用于区分选择/剃刀/波纹/滑移等行为。本项目有波纹删除开关不自动等价完整波纹/滑移工具。
- [Shotcut官方功能表](https://www.shotcut.org/features/)：公开列出图像序列、Lottie/Rive、字幕格式、高位深处理、工程交换等；只作差异基线，不把其全部实现或质量作为本轮已实测事实。
- [auto-editor官方Edit Reference](https://auto-editor.com/ref/edit)：音频活动检测读取样本，运动检测比较灰度帧；字幕匹配是独立方法。因此不能把字幕间隙压缩换名为音频静音分析。

本轮只新增本报告，不改产品源码、不改既有134/31项状态、不提交、不推送。交给实施任务的优先候选：统一工具栏真实回调/禁用状态，页面宠物复用同一Agent，再补选中片段/播放头上下文。完整功能对齐仍需按旧清单逐项关闭剩余缺口。
