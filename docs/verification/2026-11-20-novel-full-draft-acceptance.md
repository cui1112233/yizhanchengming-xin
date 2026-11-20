# 短篇全篇、缺章补写及二稿保留验收

代码提交：`d421ccb1`，v88原开发分支，无新分支、无推送、无公网发布。原件固定为 `F:/软件/洛水2026/洛水2026/modules/novels.pyd`，SHA256 `b5b22086c64851e0e94fc70a509cddbaf4ec144ac5beedc64e3a9a8caa40e57d`。

## 证据 → 规格 → 实现

- 只读反编译项目：`_local-archive-20261028/luoshui-novel-native-20261119/ghidra-project/novel-verified-native.gpr`。
- 完整helper、原模板、C和Listing及文件行号：`docs/verification/novel-full-draft-20261120/FINDINGS.md`。本轮实际使用Craft调查计划和交接；原版函数分析复用已存Ghidra项目，本轮导出不是新REA Evidence ID。
- 规格：`docs/specs/novel-creation-alignment/2026-11-20-short-full-draft-completion-contract.md`。
- 原样导出备份和规范化哈希索引：`_local-archive-20261028/novel-full-short-repair-20261120/native-raw/`；只去行尾空白和多余EOF换行，未删除内部行。
- 后端：`backend/internal/novelcreation/short_full_parser.go`、`short_full_pipeline.go`；模板均在 `backend/internal/prompt/novelcreation/short_full/`，其provenance记录来源、模板SHA和CM适配。

本轮已修复并独立复查的实际差异：书名/章名取首块而正文取末次非空块；补写预算按len(missing)；完整首稿在后续审稿前落库；一次二稿四消息；无效二稿保留首稿；跨模式损坏快照在模型调用前拒绝；编辑后同步当前完整下载、保留历史原稿；页面使用真实writing状态。

## 实际链路

新建短篇仿写才启用版本1：专用拆书/方向/蓝图确认 → CM策划大纲适配 → 整篇Markdown首稿 → 若缺章最多补写3次、每次重新检查 → 持久化完整v001 → 文字审稿 → 一次完整二稿 → 有效采用v002或保留v001并记录拒绝原因 → 包装 → 已有仿写终审 → 完成。

正文审稿不再套旧分段评分、多轮修订或±45%字数门槛。37维度、自定义写作/审稿和自家阶段模型继续进入真实请求。系统新增9条后台预设，总39条；发布内容冻结到任务，旧缺省版本0任务继续原协议。

已保存的首稿、二稿、未采用二稿、审稿、包装原文和当前完整正文均有实际文件入口。历史版本只读；只在真实文件存在时显示入口。失败/取消时不会继续显示“正在”。

取消、限流、鉴权、未知付费结果、持久化失败会停止或暂停；只有确定无效的第二稿才回退并继续包装。这是对原版宽泛Exception捕获的明确安全适配。

## 验证结果

| 检查 | 结果 |
| --- | --- |
| Go novelcreation、prompt/novelcreation、storage全包 | 通过；root最后全包31.482s / 0.959s / 0.355s；员工新增边界后全包再通过 |
| 本轮focused全篇及预设测试 | 两包通过，0.445s / 0.323s |
| HTTP小说相关既有回归 | 通过，1.338s |
| 实际integrationTextProvider→本地HTTP | 首稿Markdown无response_format、修订四消息、自家模型与参数均通过；旧四/六消息也通过 |
| 新全篇真实隔离MySQL8.4 | 通过131.56s：缺章补齐、无效二稿保留、包装429、服务重开只重试包装，12次假回复 |
| 旧writer两协议隔离MySQL | 通过186.73s：旧0/四消息1均保存并恢复，各12次假回复 |
| 旧蓝图/确认/重开/终审/429隔离MySQL | 初次组合执行续跑发生context canceled；没有放宽超时或改worker。单独新nonce库重跑通过300.04s，19次假回复。首次失败日志保留，未据单次重跑推断取消根因 |
| 新真实浏览器 | 最终6/6通过69.56s；真实writing首稿、缺章、pending，采用/拒绝、只读文件及失败取消状态 |
| 既有文件编辑/确认真实浏览器 | 6/6通过32.95s |
| 前端生产构建 | 通过1m55s，有既有大chunk提示 |
| Go embed与二进制 | 静态资源检查及Go构建通过 |
| 独立只读代码复查 | 未发现新的可复现严重错链；已确认两处跨模式/编辑同步修复 |

浏览器红绿、数据库、构建及本地身份记录均在 `_local-archive-20261028/novel-full-short-repair-20261120/`。所有模型测试使用固定假回复或本地HTTP；本轮没有收费模型测试、没有执行原版程序。

## 本地4001验收

- PID：52100，二进制：`_local-archive-20261028/novel-full-d421ccb1/qiantie-novel-full.exe`，相邻 `runtime-manifest.json`。
- `/health`：`ok:true`；`/api/runtime-build-info` 的git_sha和runtime_manifest_sha均为d421ccb1。
- `/novel-creation`：HTTP200；入口 `/assets/user-BUNeHBtC.js`。
- 新小说chunk：`NovelCreationPage-f6-enl-2.js`，SHA256 `0fb1a57dc3e39de024997b5680e49ac4c8738e655d6198bb71e4ea9290e7e168`。服务返回文件与本轮生成文件相同。
- 运行后 `data/system/presets.json` 实查9条 `novel-creation.short-full-*` 全部published，包含完整system/user/补写/二稿/包装/章节模板。
- 启动日志明确未启用开机恢复批量任务；本轮未新建真实生成任务。

## 仍不能混称完成的范围

本轮完成的是短篇仿写全篇正文及补写、回退这条明确合同。CM保留JSON策划大纲、配置和安全保护；原版宽松Markdown回退的已知分章缺陷也明确保留并测试，正常模板要求CHAPTER标记。真实模型文章质量、原版实际运行和全软件逐像素一致性未在本轮验收。工作区仍包含其他任务的修改，本地整站构建包含其现有文件，不等于干净公网发布包。
