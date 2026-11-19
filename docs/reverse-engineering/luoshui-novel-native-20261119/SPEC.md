# 小说写作链路对齐规格：原包证据先行

版本：novels.pyd b5b22086c64851e0e94fc70a509cddbaf4ec144ac5beedc64e3a9a8caa40e57d。

证据等级：A=REA观察到的原生函数体/构造参数，B=完整解码的常量及原位置，C=当前CM源码；动态调用和最终提示词若未恢复完整，明确为待确认。所有native伪代码不是原Python源码。本文是实施合同，未完成项不能标为已上线。

## 1. 写作提示词的真实组装

来源：`native/0x1801dfdd0.annotated.c:208`起（输出格式分支）、`:464`起（题材开头）、`:530`起（核心规则）、`:589`起（受约束输入）；函数由`writer-module.pseudocode.c`中的原生构造绑定确认为build_writer_system_prompt。完整REA：ev_3bdfb90ae4fa90bfa017ecabdb88e0b1655a8f50489ba5dc4b9dd19e35958102。

- 输入：语言、题材配置、当前写作模式及相关书籍规则，实际参数顺序按反编译继续还原，不用CM变量名冒充原签名。
- 输出：按分支拼接的写作系统提示词；已找到25个builder，包括题材、核心规则、创作宪法、代入感、黄金开篇、叙事人称、文风、写前检查和输出格式。不能仅用一条“你是小说作家”替代。
- 边界：中文/英文、创作/常规等输出格式有分支；只加载匹配分支。缺少资料不能编造已有文风或规则。
- 失败：无法构造必要上下文时应明确缺哪项，不把缺失状态当模型已写完。原版完整异常文案仍待对应调用方补证。
- CM差距：`backend/internal/prompt/novelcreation/definitions.go`现有ChapterDraft为自写JSON合同；原25个builder尚未逐项移植。必须先建立输出解析兼容层，再替换提示词。

## 2. 输出区块与剧情结算

来源：`readable/novels--prompts--inkos--writer_prompts.md`（每个Constant标题有原字节范围），明确存在PRE_WRITE_CHECK、正文输出和POST_SETTLEMENT要求；核心格式builder构造地址见`native-index.json`。

- 输入：本章memo、伏笔账、人物与世界规则、前文及目标篇幅。
- 输出：写前检查、正文和对应结算内容；各模式精确区块名称/可选条件须由build_output_format与build_creative_output_format补证。禁止将所有模式强行统一成猜测的四个区块。
- 边界：写前分析术语不能漏进正文；伏笔ID和数值结算不能漏进叙事正文；全员追踪启用时结算需增加人物清单与关系变化。
- 失败：缺失必要区块、重复区块、空正文、区块误序均需可解释错误，禁止把原始格式全文当正文保存。
- CM差距：当前先单独写正文，再另调用settlement/memory，不等于原版输出与分工；需分别保留原始输出、解析正文及状态增量。

## 3. 审稿循环

来源：`native/0x1800b9330.annotated.c:368`定义assess，`:409`读取audit，`:1920`及`:3113`关联pass_score，`:2415`起比较length_in_range并选择best。REA：ev_684a0100194ec31e6b8494e8e69aaaa339cea0a178c463f433a1c4aa3d4aedbd。常量`constants/novels--pipeline--chapter_review_cycle.json`含max_review_iterations、temperature_step、NET_IMPROVEMENT_EPSILON及退出/回退信息。

- 输入：正文、目标字数、审核函数、修订函数、可选后写/AI味/敏感检查回调、通过分、修订轮数、温度步进和取消检查。
- 输出：final_content、final_word_count、revised、rolled_back、iterations和snapshots；不是只有最后一个分数。
- 边界：空正文报“章节内容为空，无法审稿”；存在无新正文、达到通过线、无净提升、回退最高分等退出路径。通过与最佳候选还关联字数范围；具体比较运算需小函数证据及assess完整体补齐，不能仅凭字符串确定阈值。
- 失败：没有改善不能无限重复调用；取消应中断后续调用。原版敏感检查与确定性检查的合并字段、分数处理待assess/merge函数补证。
- CM差距：当前主要按模型分数/阻塞问题择优，本地检查放在审核循环之后，尚未按原版assess合并。此项登记为待修，不能因为37维已进入请求就视为完整审稿。

本轮已修其中post-write一项：`native/0x1800bf180.fallback.annotated.c`确认assess调用检查回调；`native/0x1800b7c40.fallback.annotated.c:654`起将error提升为critical，`:153`将blocked审核的passed置false，`:707`记录source=post_write。当前CM对新版任务在每次审核后合并本地校验问题，阻止假通过并驱动修订；不编造扣分。AI味/敏感检查尚未按原版完整合并，候选字数范围择优、无净提升提前退出及局部PATCHES链路继续待修。这里不代表审稿循环全部等价。

## 4. 局部修订与整章修订

来源：`readable/novels--prompts--inkos--reviser_prompts.md` Constant70/71/75/76，分别给出结构问题整章、局部问题PATCHES、混合问题整章的规则。工厂REA证据ev_121d5878896af8131d1f77a05fcd573773157be9653a5e956a49b6e6b652b5ef证明build_auto_system_prompt对应0x1801c8710。函数体完整补证仍在进行，不能仅凭工厂推定运行分支。

- 输入：分级审核问题、当前正文、规则/状态/伏笔上下文及修订路由。
- 输出：局部问题产出PATCHES；结构、篇幅、主线/人设/时间线等问题产出REVISED_CONTENT。混合问题用整章结果。原版FIXED_ISSUES、UPDATED_STATE、UPDATED_HOOKS等内容需由相应格式和解析函数逐项验证。
- 边界：局部补丁保留未改动文字；不能用补丁假装完成整章压缩。无法安全修改时不能虚构已修复。原版补丁定位、重复文本和冲突处理还需spot_fix_patches函数补证。
- 失败：无新内容退出、补丁无法匹配不允许静默全文替换。后一句为CM必需的验收保护，尚未声称原版完全一致。
- CM差距：现有chapterRevisionOutput只接strategy/content/resolvedIssueIDs，尚未接原版PATCHES及原格式状态变更。必须先解析与验证，再连保存和重试，不能只换system prompt。

## 5. 长篇仿写、短篇、番外、续写不能合并成一个换词流程

来源：95模块中分别存在`novels.pipeline.imitation_short_runner`、`short_fiction_runner`、`pipeline_runner`；已有前端/桥接证据另见`2026-11-17-luoshui-novel-bridge-contracts.md`。模块存在只证明分工，不证明完整调用次序。

- 输入：每种模式的原表单字段与选择项；长篇仿写有referenceChunks/styleDimensions/selectedDirection，短篇、番外及导入续写使用各自入口。
- 输出：模式专用作品对象、资料和执行状态，禁止假装全都等于长篇的一章。
- 边界：全局Agent设置继承、关联作品、原文/文风选择及取消/重开均需逐入口对照；已存在CM实现不能充当原版证据。
- 失败：未补证的内部生成顺序、提示词动态拼接及网络参数标未知。正常/空/重复/错误/取消输入分别验收。
- CM差距：短篇设置继承、番外专用状态、完整资料和模型分工仍未等价；本轮不能标完成。

## 实施顺序与验收条件

1. 完成输出解析及修订路由规格，固定真实输入/输出样本（不能自造样本后称原版）。
2. 在后端prompt模块迁移已证实的builder和原文片段，保留来源/版本；使用自家模型配置。
3. 按原版状态流接入审稿、修订、记忆、伏笔及结果保存；旧任务保护另列兼容行为，不能改旧快照。
4. 测正常、空值、重复、错误、取消；断言请求与保存结果，不只断言按钮存在。
5. 原/CM相同输入实测后登记等价；无预算时只做离线链路测试，明确不能证明真实产出质量。
