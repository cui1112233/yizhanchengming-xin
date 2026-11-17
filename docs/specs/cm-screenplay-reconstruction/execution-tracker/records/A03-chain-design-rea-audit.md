# A03 链路设计重新核查：REA与原版对照（2026-11-18）

结论：现有实现具有真实后端和保存逻辑，但不能称作固定ZIP的完整流程复刻。问题不止上游503；主调度、资产提示词及模型参数存在实际差异。此前恢复证据和局部实现不等于这些差异已关闭。

## 来源、工具及证据边界

原件F:/软件/洛水2026.zip，SHA934e43c03115587777050970332deeb0cf653e1c0bfb80b180c9e386f89bcf8c。可读前端readable-supplied-frontend/pretty/index-CnO3GE3B.js，SHAfc78d66444269ecb4af9b6114c649e735c11c7770dbfa0cc1d262fd2f3168335，以下O行号均指它。full-extracted/洛水2026仍保留原文件，readable-native-agent-prompts保留从原生模型边界截获的内置system。

本轮实际调用REA analyze_javascript_application：完整大文件60秒超时（ev_ff28d8a3e39bebed56d542a6589de4b5535d04d69aa1100f9ea7aa3c4a2f92dd）；改用AST精确提取的at117206-117426和tt117091-117185，无改写函数正文，两个模块共14132字节成功解析，2942个AST节点。源范围与摘要见source-map.json。REA父证据ev_610c037d2726eef682a92baaed0dddb72b9475b789e1facf940796ce924cfb42，摘要ev_46eac32bf2f6237e40e52d252099e80ea526e271d523cd3a2b686f0b6e96698f；静态语义有586项未知，不能冒充完整自动调用图。精确property追踪ev_24ec2a1d64684c22301c7fd98837e7746828a8860eecb7930dfeed71cb59bdfa返回partial且无种子匹配，未用它证明动态桥接已接通。具体桥接字段来自原文件逐行读取和以前原生受控捕获，不是REA自动证明运行结果。

本轮自己的最终请求捕获：Go overlay在4001来源冻结副本加入单独测试，不修改原快照或产品代码。真实执行AssetService.Extract到TextGenerator边界后由替身停止，不调用网络、不生成成果。三类资产×四档共12组，cm-asset-requests.json保留完整消息。原始对比读取JSON字段名曾错用snake_case得到空文本，纠正为SystemPrompt/UserPrompt并断言非空后重新计算；只有asset-comparison.json最终非空结果为依据。模型/画风用OWN_*，没有凭据。原生结果仍是此前受控记录，本轮没有重复运行原版或消费模型额度。

## 原版实际设计与CM差异

| 编号 | 原版规则及位置 | CM当前对应位置 | 本轮结论和修复要求 |
|---|---|---|---|
| CD01 主Agent步骤 | O118972-119011传pipeline和content/segments；O118794起读取initialPipeline；原生agent-pipeline-v2 normal/reverse分别以所选顺序执行四项 | backend/internal/screenplay/state.go:23固定0_ingest到8_finalize；batch_executor.go:171按固定步骤执行；production_runtime.go:370调用RunNext | 固定自动制作不是原版可选顺序的主Agent。此对比范围为该执行器，不推断全站没有其他Agent。应补原版步骤ID映射、顺序、每项输入输出/状态与持久化，自己的服务承接 |
| CD02 资产解析system | 原生characters/scenes/props system分别17476/11965/7578字符，zip-baseline/13及readable-native-agent-prompts/manifest.json | assets_service.go:84按type/degree选Registry；:354四档映射仍有TODO需运行确认 | 12个实捕请求与相应原版单独解析system均不相同。不是缺12个功能，也不是证明四档都该用同一个system；须先确定每个入口/档位的真实原版规则，再注册到后端prompt并比较最终消息 |
| CD03 资产正文/画风与模型 | 原版单独重解析从作品ai_config读模型；本组style_desc没有拼进提取system/user（原生asset-parse-v3/v4，zip-baseline/13） | assets_service.go:118传styledAssetContent；:548拼画风要求；:266使用req.Model，空值默认gpt-4o-mini | 本轮OWN_STYLE实捕证明确实多了画风段。项目已存模型与请求模型不同时采用请求值，不能与原版入口混为同一规则。保留自家模型目录，但各入口取值和解析/出图的画风作用需分开 |
| CD04 分镜请求内容 | O117118-117139手动单场带原文、一个场景、all_scenes、模板和画风；批量入口O117206起 | storyboard_scenes.go:110-155已组三段用户消息及模板变量，storyboard_scene_reference.go取章节原文 | 这部分已有实现与原生捕获对照，不能称所有按钮都是空壳；仍需逐模板协议和真实成功回复验收 |
| CD05 分镜传输参数 | 原生storyboard-matrix-v1最终模型边界temperature1/max_tokens20000/streamfalse；zip-baseline/11:实际模型消息 | storyboard_scenes.go:155未设Temperature/MaxOutputTokens；httpapi/screenplay_routes.go:42默认16000并转发零值温度 | 默认参数仍不同。先明确自己模型能力上限与原版语义，再给恢复模板设置显式后端策略；不直接照搬原版多层自动重试导致重复费用 |
| CD06 生成与保存分离 | O117353调用save_chapter_storyboard_result，再se保存chapters；原生原回复日志和可编辑作品是两份；zip-baseline/12正常重开记录 | storyboard_generation.go检查点与storyboard_service.go保存存在；本地C03仍0个生成分镜 | 原回复、解析结果、保存版本、重新打开应逐层验收，503失败不能当成后续层通过。已有受控保存证据不等于真实整链输出 |
| CD07 媒体和导出 | 原视频队列/供应商请求与导出分别调查；zip-baseline/18只验证入队及协议，19未验证合法原生最终产物 | 本计划ZG17-21仍待逐项对照 | 未完成配音绑定→视频落盘/入库→编辑→导出的成品链路对照，不能由按钮存在推导接通 |

## 原版流程用大白话说明

1. 新建作品：保存原文、章节和作品配置（模型、剧本提示词、画风、比例）。ZIP入口是TXT，CM直接输入/DOCX是用户明确要求的扩展，继续保留。
2. 主Agent：读取用户选的流程并按顺序处理；可先做摘要、角色/场景/道具分析，或只生成剧本等。完整原文与分段分别传入，各步产物独立存在。
3. 剧本产物：结构化场景供后续分镜，资产解析有独立内置提示词。解析文字和资产出图是不同操作，不应把画风作用混成一条。
4. 分镜：读取选择的分镜词，用整章原文、当前单场景和前后场上下文组请求；显示每场结果；部分失败保留成功场并定位失败场。
5. 保存：原回复日志只是记录，最终还需保存可编辑分镜及章节；打开作品需读回相同结果。
6. 后续媒体：资产参考图/音色绑定、配音、视频任务、结果文件和导出分别有自己的状态与保存，不能跳过后几层就称成片。

原版也有假成功、重试第二场错覆盖第一场、保存失败仍报完成等缺陷，沿既有缺陷例外不复制。所谓对齐是用户可用流程和输入输出规则一致，而不是复制缺陷或原版账号密钥。

## 下一批顺序及验收门槛

P0先封闭CD01-CD03：恢复主Agent流程合同→明确每步使用的后端提示词与字段→结果保存/重开→同样输入和同样选项在原版与CM比较。正常、空值、重复步骤/重复提交、部分失败、取消逐项登记。不能只给固定9步换成原版名称。

P1补CD05及全部模板输出协议，保持自己的模型、单次传输与防重复收费；旧结果不被覆盖，失败不擅自自动补试。

P2再补CD06-CD07媒体及实际产物；真实生成预算仍为0，继续免费核查不阻塞，但新的真实调用需单独预算。

本轮为调查登记，没有修改产品、重启4001、切换模型或调用收费服务。以此报告替代“都对齐”的口头结论；A03/A06/B04/B09/C03未完成项不关闭。

## 12组消息对比（字符数仅用于识别，不代表质量）

| 资产 | 档位 | 当前提示词ID | 原版system字符 | CM实际system字符 | 逐字相同 |
|---|---|---|---:|---:|---|
| character | standard | asset_character_standard_v5 | 17476 | 10240 | 否 |
| character | more_diff | asset_character_lite | 17476 | 5035 | 否 |
| character | master | asset_character_v6_3d | 17476 | 13661 | 否 |
| character | master_real | asset_character_real_v7 | 17476 | 10934 | 否 |
| scene | standard | asset_scene_v3 | 11965 | 21229 | 否 |
| scene | more_diff | asset_scene_v3 | 11965 | 21229 | 否 |
| scene | master | asset_scene_v4 | 11965 | 24516 | 否 |
| scene | master_real | asset_scene_v4 | 11965 | 24516 | 否 |
| prop | standard | asset_prop_v3 | 7578 | 5799 | 否 |
| prop | more_diff | asset_prop_v3 | 7578 | 5799 | 否 |
| prop | master | asset_prop_v4 | 7578 | 9809 | 否 |
| prop | master_real | asset_prop_v4 | 7578 | 9809 | 否 |
