# 短篇修订上下文实施与补写对照

日期：2026-11-20。执行计划：`docs/plans/2026-11-20-novel-short-writer-recovery.md`。本轮不调用收费模型，不运行原版，不发布公网。

## 先保存证据，再写规格

原件novels.pyd SHA256 b5b22086c64851e0e94fc70a509cddbaf4ec144ac5beedc64e3a9a8caa40e57d。只读复用既有Ghidra项目，不重新全量import。函数完整body、较小helper伪代码、常量赋值与前端回调均已保存：

- [原版规则与文件行号](novel-short-recovery-20261120/FINDINGS.md)。原 `_produce` C 导出超时，以完整函数Listing确认分支；不是假称C恢复成功。
- [原版跟进模板](novel-short-recovery-20261120/revision-followup-template.txt)。
- [输入、输出、边界和失败合同](../specs/novel-creation-alignment/2026-11-20-short-writer-conversation-contract.md)。

新增Ghidra导出是分析员本地证据，不冒用REA Evidence ID。Git归档仅规范行尾空白和末尾换行；原始副本保存在_local-archive-20261028/novel-short-writer-recovery-20261120/native-raw/，对应SHA见normalization-index.json。

## 已实现的修复

之前CM正文修订是一个独立system加user请求，缺少原写作对话。本轮新建短篇仿写任务冻结 `short_writer_conversation_version=1`，发送：

1. 原首稿实际使用的writer system，包含冻结写作要求和仿写附加约束。
2. 原首稿实际user输入。
3. 有效首稿的原始回复。
4. 本轮修订跟进要求、原配置修订预设和本轮审核输入。

后台新增 **短篇写作·修订跟进要求（四消息）**，ID `novel-creation.short-writer-revision-followup`，预设数29→30。修订重点有原版常量72–80及函数构造证据；全篇Markdown输出明确改为CM当前章/段JSON适配。原文来源和适配说明在 `backend/internal/prompt/novelcreation/short_writer/provenance.json`。后台实际发布正文冻结进任务，前端没有硬编码这段提示词。

每段独立保存首稿上下文；格式修复成功后的有效回复可作为首稿。后续修订不覆盖它。修订格式无效时保留四条消息，再追加坏回复和结构修复要求，共六条。实际messages参与调用指纹及持久化metadata。429恢复从已有成功调用重建上下文，不重新生成首稿；取消和结果不明继续遵守现有保护。

旧任务未开启该字段时，继续发送原两消息请求及原格式修复请求，不改变原日志指纹。短篇原创不自动开启本轮新字段。未知协议版本在调用前停止。

## 本轮没有冒称实现的差异

| 对照项 | 原版证据 | CM当前情况 |
| --- | --- | --- |
| 缺章补写 | 首稿后最多3次；缺失以content.strip为空判断；每次合并raw、重新解析、重新检查，最终验证 | 仍为按章分段JSON，不等于原版整篇Markdown缺章解析器；本轮完成spec取证，尚未完整移植 |
| 正文修订次数 | 可见路径一次完整第二版，assistant为补写完成后的首稿 | 保留CM用户可配置多轮；只恢复四消息上下文关系，不称次数完全相同 |
| 无效改稿回退 | 原版不让不完整/解析失败第二稿覆盖完整首稿，保留v001 | CM仍有自己的审核失败与停止规则；本轮没有吞掉所有修订异常后标成功，回退展示及后续流程需另行对齐 |
| 修改后独立终审 | 已检查ShortPanel、Studio saved回调及API直接方法库存中未找到独立入口 | 保留旧终审过期提示，不添加假原版按钮；此结论只覆盖已查范围 |
| 提示词和输出 | 原版全篇带标记Markdown，自然语言审稿 | 本轮跟进要求有来源并实际使用，但按段JSON、评分循环和部分动态builder仍是CM适配 |

## 验证登记

- 四消息及修订格式修复测试先失败后通过；冻结writer/addon/原修订/新增跟进预设均实际进入对应消息。
- 两轮修订保持有效首稿、分段隔离、取消、缺上下文停止、未知版本、旧协议、自家模型分路及429恢复已通过。
- 完整novelcreation、prompt/novelcreation、storage通过；小说HTTP定向回归通过。
- 本地HTTP模拟渠道验证实际发送的四/六消息、顺序及配置模型，未联系任何外部模型服务。
- 独立静态审查未发现阻断问题，未把原版全篇规则与CM分段适配混称一致。
- 独立MySQL8.4通过（178.233s）：原有蓝图编辑恢复、终审429复用19次模拟调用；旧/新对话协议分别12次模拟调用。新协议重开后实际发出四消息且消息存入metadata，旧协议仍两消息，无重复首稿。总计43次固定回复模拟调用，零付费。临时库和账号已清理。
- 精简日志见[验收附件](novel-short-writer-recovery-20261120/)，原始测试日志在_local-archive-20261028/novel-short-writer-recovery-20261120/。
- React构建通过（1m47s），Go embed准备和Go编译通过。源码提交 `d8fdef38`；运行PID30572，`_local-archive-20261028/novel-writer-d8fdef38/qiantie-novel-writer.exe`；git_sha与runtime_manifest_sha均为d8fdef38，health.ok=true，小说入口200。
- 实际小说chunk与嵌入构建相同（NovelCreationPage-C1UnCqa5.js，SHA256 57ecc75875c2cff8d07df9c41b7cf63fba3cb9601514243daad038ba0319f4f4）；本轮无小说前端业务改动。入口脚本仍为user-Bj7mg_zH.js。
- 本地系统预设文件已核验新ID状态published，名称“短篇写作·修订跟进要求（四消息）”。现有已创建任务不会被新预设覆盖。
- 本地包包含工作区现有其他并行改动，它们没有混入本轮提交；不声称干净检出的全站发布包。公网未部署、未推送，实际收费模型调用0。
