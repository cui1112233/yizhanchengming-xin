# AI-CUT 实际链路与开源对照审计

日期：2026-11-12。范围：当前 `F:/脚本测试/一战晟铭-v88`，v88 分支，审计开始 HEAD `40bd0ace` 及现有未提交改动。本轮核查页面、请求、处理、保存、预览与导出代码；没有修改产品逻辑、重启程序、调用付费模型或改动用户工程。报告不是134项功能全部现场验收，也不假定正在运行的二进制与未提交源码完全一致。

## 结论

存在控件空转、能力描述超出实际范围，以及预览/导出不一致。不能再以“已接线”“20项工具”“原测试通过”代表完整可用。下列问题中，5项已用临时测试复现，其余标明为源码核查；并非全部都是空按钮。

## 已复现的五项问题

| 编号 | 功能和触发 | 实际结果 | 证据 | 优先级 |
| --- | --- | --- | --- | --- |
| F01 | 已导入剧本、有标记，再导入SRT或识别字幕 | 新时间线只保留轨道，Script变成空，Markers变成0个；ASR成功后会保存这个结果 | `backend/internal/aicut/srt.go:117`；`asr_service.go:200`；诊断 `TestAuditSubtitleImportPreservesScriptAndMarkers` | P1，数据丢失 |
| F02 | 音频片段音量设0后导出，且无覆盖音量的关键帧 | 预览音量0，导出生成 `volume=1.000`，会恢复声音 | `backend/internal/httpapi/aicut_export.go:655`；`frontend/src/user/pages/aicutPreview.js:69`；`TestAuditZeroVolumeRemainsSilentInExport` | P1，成片错误 |
| F03 | 4秒片段速度从0.5连续变到2倍 | 预览第1秒对应源688ms、瞬时0.875倍；导出整段采用1.25倍，对应源1250ms | `backend/internal/httpapi/aicut_export.go:731`；`aicutPreview.js:54`；`TestAuditContinuousSpeedCurveNotFlattenedToAverage` | P2，时间映射不一致 |
| F04 | 10秒视频，特效仅放4–5秒 | 导出把滤镜套在整个0–10秒片段上，没有特效时间窗口 | `backend/internal/httpapi/aicut_export.go:517`、`:778`；`TestAuditEffectLayerRetainsItsTimeRange` | P2，作用范围错误 |
| F05 | 自动成片关闭字幕，模型仍返回启用的add_caption | 后端接受并应用字幕操作，没有按用户开关拦截 | `backend/internal/aicut/edit_plan_apply.go:219`；`TestAuditAutoPlanRespectsDisabledCaptionOption` | P2，选项约束不生效 |

上述导出诊断检查的是程序实际生成的FFmpeg处理参数，没有渲染用户素材或逐帧/逐样本测量成片。预览源位置与音量另通过直接调用当前JS纯函数核实。

## 源码确认的断点、空转与能力边界

| 编号 | 界面/能力 | 真实链路 | 证据与判断 |
| --- | --- | --- | --- |
| F06 | 自动成片：选择刚上传素材，期待从空工程生成视频 | 模型收到时间线、素材ID和选项，没有素材名称/时长等资料，也没有读取声音/视频帧；允许操作不能新增视频/图片片段，仅背景音乐可新增媒体。空时间线无法完成自动组装 | `backend/internal/httpapi/aicut_edit_plan_routes.go:112`；`backend/internal/prompt/aicut_edit_plan.go:25`、`:42`；`edit_plan_apply.go:239`。属于已有时间线修改草案，完整自动成片未完成 |
| F07 | ASR字幕与剪辑后视频对齐 | 转写整份原素材，直接以源字幕时间作为时间线时间；不对应素材实例的开始位置、源裁剪点和倍速 | `backend/internal/aicut/asr_service.go:185`；`srt.go:111`。视频放第10秒，原视频第1秒对白仍落工程第1秒。手动导入SRT本就可能使用工程时间，问题特指ASR映射 |
| F08 | Agent“确认可用ASR模型后识别字幕” | 工具强制要求ASR模型ID，项目快照只给素材/时间线；没有查询模型的工具，前端ASR目录也没交给Agent | `frontend/src/user/components/AICutAgentPanel.jsx:14`；`backend/internal/aicut/agent_tools.go:157`、`:647`。用户手工提供ID可执行，自动选择模型链未完成 |
| F09 | 已添加的独立特效轨实时预览 | 特效片段无assetId，播放器查不到素材直接跳过；导出有真实滤镜 | `frontend/src/user/pages/aicutPreview.js:131`；`AICutPage.jsx:3494`；`aicut_export.go:778`。直接调用预览函数，加入生效中的鱼眼特效前后返回值完全相同。不是无导出实现，而是预览缺失 |
| F10 | 资源滤镜强度 | 浏览器写catalogParameters，后端Clip无该字段；导出强制强度0.5，并在VisualEffect后再叠一遍目录滤镜 | `frontend/src/user/pages/AICutPage.jsx:3515`；`backend/internal/aicut/types.go:208`；`aicut_export.go:774`。控件、保存和导出不同步；非所有滤镜都只表现为同一种偏差 |
| F11 | 文字资源库的字号和不透明度 | 资源面板可调整参数，应用文字分支却只用固定templateStyle，没有使用parameters | `frontend/src/user/pages/aicut/editorCatalog.js:109`；`ResourceLibrary.jsx:155`；`AICutPage.jsx:3543`、`:3558`。这两个资源参数属于控件空转，不等于整个文字属性编辑器都无效 |
| F12 | 移动物体先跟踪再消除 | 跟踪修改整个视频的位置关键帧；消除区域还是固定矩形，导出是静态delogo，二者没有绑定 | `frontend/src/user/pages/aicut/InspectorPanel.jsx:447`；`aicutCommands.js:878`、`:1157`；`aicut_export.go:1019`。固定水印修复是真实能力，移动物体消除的引导超出实际能力 |
| F13 | 渐变标题 | 网页CSS为多色渐变，导出ASS仅单一紫色 | `frontend/src/user/pages/aicut-page.css:90`；`backend/internal/httpapi/aicut_export.go:1949`、`:1999`。预览/成片样式不一致 |
| F14 | 曝光等基础调色 | 网页曝光用RGB乘2^exposure；导出改用eq亮度exposure/4，色温等也不是相同运算 | `frontend/src/user/pages/aicutVisualProcessing.js:535`；`backend/internal/httpapi/aicut_export.go:1345`。HSL/RGB曲线/色轮确有LUT处理，不能泛化为整个调色都是假功能 |
| F15 | 草案审核与预计时长 | warnings保存但界面不显示；审核项缺裁剪时间、字幕文字和实际参数；预计时长直接显示模型自报，勾选后不重算 | `backend/internal/aicut/edit_plan.go:92`、`:321`；`frontend/src/user/components/AICutAutoEditDrawer.jsx:136`。审核界面不够完整，预计时长不能充当已测量结果；目标时长作为软目标本身不算缺陷 |
| F16 | 10/12-bit输出 | 有真实编码参数，但合成中间经过rgba及yuv420p，再转换成高位深输出 | `backend/internal/httpapi/aicut_export.go:823`、`:889`、`:937`。不能据输出位深宣称10-bit全流程。当前没找到明确HDR界面宣称，HDR缺失不计为虚假按钮；本轮未做色阶像素测量 |

## 开源实际链路对照

- [FunClip videoclipper.py](https://github.com/modelscope/FunClip/blob/main/funclip/videoclipper.py)：本轮直接读取源码，`video_recog → recog` 从实际声音得到识别结果与时间戳；`video_clip` 用时间戳执行subclip，拼接视频，同时借助generate_srt_clip与累计偏移重排字幕，最后write_videofile。说明“文字选片”需要真实内容、时间映射与成片，不只是一段规划提示词。该项目自身也有尚未完成的计划项，本轮不把README的计划当成实现。
- [auto-editor](https://github.com/WyattBlue/auto-editor)：官方说明按音频响度/画面运动检测活动区间，再按margin保留自然边界。我们的字幕间隔压缩只读取字幕起止点，没有分析声音，不能等价称为真实静音检测或完整自动去废片。
- [Shotcut 官方功能与处理范围](https://www.shotcut.org/features/)：其公开列出滤镜关键帧、时间重映射、轨道混合、字幕导入/编辑/渲染、10-bit端到端、MLT保存/加载/导出等。这里采用的是链路验收标准，不假定它每个滤镜都与我们算法相同，也不因其有功能就反推我们存在假按钮。
- 之前文档提及的OpenChatCut固定链接本轮浏览工具未能读取，未把它作为此次新问题的证据。

## 真实实现，但有范围限制

- 人物抠像：MediaPipe轻量模型、生成蒙版、上传并写入工程、导出使用蒙版都有实现；不代表复杂发丝/多人场景质量已验证。
- 区域跟踪：真实抽帧和像素匹配，生成整个视频的位移关键帧；不等于AI语义跟踪、移动消除或任意贴纸自动绑定。
- 贴纸：自有图片作为独立画面轨，不是第三方贴纸商城。
- 工程模板/个人排布模板：实际生成轨道、绑定自有素材并保存；不是丰富的在线模板体系。
- 音频降噪、EQ、声道、响度、变声等：导出确实进入FFmpeg。部分试听是近似处理，界面已有说明。
- 智能切镜：使用FFmpeg画面变化检测，不等于理解剧情与镜头好坏。
- Agent字幕选片、字幕间停顿压缩、动画、调色、批量操作：均有实际执行和保存；它们仍依赖字幕质量及被调用的底层处理正确性。
- 自定义流程整理：目前是可复制的说明，尚不是能保存/再次调用的自定义技能库；当前提示词已有如实限定。

## 验证与复现

临时诊断文件位于 `_local-archive-20261028/aicut-chain-audit/`，通过Go overlay把测试映射进包，未把测试或修复写入产品源码：

```powershell
go test -overlay F:/脚本测试/一战晟铭-v88/_local-archive-20261028/aicut-chain-audit/overlay.json ./internal/aicut ./internal/httpapi -run '^TestAudit' -count=1 -vet=off
```

结果：五项“应有行为”断言全部失败，分别证明F01–F05。完整输出在该目录 `diagnostic-tests.txt`。这是问题复现，不应包装成“新测试通过”。

另用Node直接调用当前 `buildAICutPreviewFrame`，得到 `sourceTimeMs=688`、`playbackRate=.875`、`volume=0`，且新增正在生效的鱼眼特效后画面描述不变。现有相关Go测试仍通过，说明旧用例没有覆盖这些契约，部分速度测试还把均速近似写成了期望。

## 建议修复次序

1. 先修剧本/标记丢失、音量0恢复声音及ASR时间映射。
2. 统一预览与导出的曲线变速、效果作用区间、滤镜强度与文字样式。
3. 补齐自动成片的真实素材解析/放置/时间映射，后端强制执行开关，审核展示完整参数与警告并重算结果时长。
4. 补ASR模型目录和任务状态查询，使Agent能可靠衔接“识别完成后再剪辑”。
5. 对尚未实现的移动物体消除、自定义技能库、高位深全流程明确标注范围，避免名称超出能力。
