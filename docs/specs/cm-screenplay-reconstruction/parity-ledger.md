# 原版逐项复刻验收台账

更新时间：2026-11-15T00:48:23.737165+00:00。当前任务：**G11 资产检查后台持续、离页、关闭重开续看同一任务**。

保留36组、695处交互位置、330处接口调用位置。当前完整对照通过 **0/36组**；此数只表示已有台账裁定，不是功能完成率。4组本地数据闭环不能当作原版复刻完成。

## 状态口径

| 状态 | 含义 |
|---|---|
| `implemented` | 有实施记录；尚未通过完整原版对照 |
| `unverified` | 未核实绑定或行为；不能以名称/代码存在作为通过 |
| `partial` | 部分实现或本地验证，仍有明确缺口 |
| `missing` | 已记录缺失能力；实现进行中也不能提前关闭 |
| `compared_passed` | 所有适用门槛具备同项原版/CM对照证据；允许记录未授权付费用例阻塞但不得把整项判通过 |

## 每项必须补齐的证据

每组及每个交互/调用位置继承 `full_parity_v1`，默认全部待验证：输入输出、边界、错误、取消与晚到结果、重复与恢复、界面/图标/几何、最终请求、保存重读，以及同项原版/CM对照。JSON中空 `checks` 表示继承全部未验门槛，不代表免验。

每条证据记录文件路径、准确位置、文件摘要、对应源身份、原版/CM归属及支持的结论；最终请求和落库重读要能串到同一请求标识。未适用项必须附原版依据和原因。付费测试未经授权保留为 `blocked_paid_test`，不能删除需求或标为通过。

规划 `succeeded` 只表示独立规划请求已保存，不等于已恢复原生续写完成规则。

## 36组固定顺序

| ID | 分组 | 验收项 | 当前实施状态 | 原版完整对照 |
|---|---|---|---|---|
| G01 | 作品库 | 新建、打开、搜索、归档、恢复 | partial | unverified |
| G02 | 输入 | 直接输入原文、TXT/DOCX读取保存 | implemented | unverified |
| G03 | 输入 | 原文建立章节，不先生成软件剧本 | implemented | unverified |
| G04 | 剧本 | 选择提示词/画风/自己模型进入最终请求 | implemented | unverified |
| G05 | 剧本 | 逐章生成、手工编辑、版本恢复 | partial | unverified |
| G06 | 剧本 | AI故事摘要、标签与原版全部字段 | partial | unverified |
| G07 | 模型配置 | 比例、分辨率、模板、核心指令、独立生图改图 | partial | unverified |
| G08 | 资产 | 人物/场景/道具增删改与解析程度 | partial | unverified |
| G09 | 资产 | 角色设计六类字段及提示词保存 | implemented | unverified |
| G10 | 资产 | 批量遗漏资产检查 | partial | unverified |
| G11 | 资产 | 检查任务后台持续、关闭重开续看 | partial | unverified |
| G12 | 资产 | 跨项目普通资产库、导入到当前库 | partial | unverified |
| G13 | 资产 | 多视图生成与各视图独立管理 | partial | unverified |
| G14 | 图片 | 自己模型、生图/改图、参考图和选中图 | implemented | unverified |
| G15 | 图片 | 批量图片、单项失败恢复 | partial | unverified |
| G16 | 站位 | 新增编辑、分析、参考资产、图片保存 | partial | unverified |
| G17 | 站位 | 画布标注、编辑图片、保存标注PNG | partial | unverified |
| G18 | 音色 | 上传、真实库选择、试听与保存 | partial | unverified |
| G19 | 分镜设置 | 收藏/最近/前缀、四密度和两个开关 | partial | unverified |
| G20 | 分镜规划 | 新规划/已有规划/无规划、来源绑定 | partial | unverified |
| G21 | 分镜 | 章节原文到指定集号、无剧本单集生成 | implemented | unverified |
| G22 | 分镜 | 无剧本批量原文章节 | implemented | unverified |
| G23 | 分镜 | 按集持久请求、回包丢失核对和续跑 | implemented | unverified |
| G24 | 分镜 | 逐场景并发/失败场景自动一次重试 | partial | unverified |
| G25 | 分镜 | 失败场景列表、单场景/全部失败重试 | partial | unverified |
| G26 | 分镜 | 原文→场景自动识别→逐场景生成 | partial | unverified |
| G27 | 分镜编辑 | 新增编辑、拆分合并、批量时长 | partial | unverified |
| G28 | 分镜检查 | AI衔接Agent独立报告与后续修正 | partial | unverified |
| G29 | Agent | 预填草稿、发送、方案应用 | partial | unverified |
| G30 | Agent | 原版全Skills、原生工具、复杂调度 | partial | unverified |
| G31 | 媒体 | 图片/音频/视频/字幕面板及文件保存 | partial | unverified |
| G32 | 媒体 | TTS、ASR、视频生成轮询与失败处理 | implemented | unverified |
| G33 | 交付 | 时间线建立并进入AI-CUT | partial | unverified |
| G34 | 交付 | 媒体包、剪映兼容草稿与成片下载 | partial | unverified |
| G35 | 自动生产 | 九步骤、质检、暂停、定时、重建、恢复 | partial | unverified |
| G36 | UI | 布局、图标、弹窗、原版像素对比 | partial | unverified |

## 原始位置、去重与绑定

原始695处交互（439点击/67弹窗/189表单更新）及330处调用在 JSON 中逐条保留原始记录、CSV行、原件坐标和稳定ID。当前按完全相同静态描述得到 690 个UI描述候选组、254 个调用描述候选组，其中 47 组含多个位置。**这些只是待核对的去重候选，不是独立功能总数。**

尚未批准任何仅凭名称、表达式或共用组件产生的行为合并。候选仍保存所有成员坐标。原先未绑定CM的位置保持 `unbound/unverified`；36组和位置之间的归属须有调用/交互证据后填写，不根据标签猜配。

## 源身份与更新办法

当前源码身份：`SRC-0e8cf2180356835e`；分支 `v88`，HEAD `25ccb869d27e587b763f397208a160b7574be52f`。工作区含未提交改动，JSON同时保存相关源文件SHA256和工作区差异摘要。HEAD不能替代当前源码文件摘要；源码身份不能替代4001运行二进制身份。运行身份仍待逐次验收证据绑定。

完整台账：[parity-ledger.json](parity-ledger.json)。生成脚本：`_local-archive-20261028/luoshui-screenplay-reconstruction-20261115/07-parity-execution/generate_parity_ledger.py`。重跑保留已填写的结果、证据、具体绑定、用例、阻塞和历史；不得在补证据前批量提高状态。

每关闭一项：先补原版出处、明确CM绑定和输入用例，再完成所需实现；核对界面/图标/几何、实际最终请求与保存重读，执行错误/取消/重复用例，写入证据并重新裁定；全部适用门槛通过才使用 `compared_passed`，随后切换下一项。当前G11只标正在工作，没有提前通过。

## 来源与限制

- `_local-archive-20261028/luoshui-screenplay-contract-20261114/full-chain-20261115/feature-chain-matrix.json`（SHA256 `b92cc2b1ad5b1e8816edd52536ed22c4723ccf245cd44e224b0e2774bb01ea41`）
- `_local-archive-20261028/luoshui-screenplay-contract-20261114/full-chain-20261115/all-ui-verification-obligations.csv`（SHA256 `ef14ca5beb66d52e4bdc46a2ed4b555bee921dd8ea625c5eac5feb7de15244ef`）
- `_local-archive-20261028/luoshui-screenplay-contract-20261114/full-chain-20261115/original-api-call-obligations.csv`（SHA256 `349d7ce3b8fed47ea808902b271c69315b1911c4aa22691ae10ef51104d2f326`）
- `docs/specs/cm-screenplay-reconstruction/04-implementation-20261115.md`（SHA256 `74a7903663d594c34df95b4a65a38890c6e62750789c58f58b0de8774a8fe7ad`）
- `docs/specs/cm-screenplay-reconstruction/05-remaining-scope-audit-20261115.md`（SHA256 `51480ce3da194498cedc8b2c99cacb803594fafccc805f15efc04e03953b830e`）

04/05实施与核账是最新范围依据；36组矩阵保留历史行号，不把历史位置宣称为当前运行绑定。没有删除未实现、原生未知、供应商未验证或受费用阻塞的义务。
