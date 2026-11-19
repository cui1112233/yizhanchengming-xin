# 状态、伏笔和记忆：原版责任链与CM适配（2026-11-19）

固定目标SHA256与其它规格相同。静态反编译不等于原版生成质量或运行时验收。

## 已恢复的责任链

- `native/0x18004a310.fallback.annotated.c:2138`：修订Agent返回提取后的正文；不是直接把修订的UPDATED_STATE写入状态。
- `native/0x180105e90.fallback.annotated.c:221,427,642`：独立调用 `_settle_validate_write_truth`，随后更新章节记忆索引和时序记忆。
- `native/0x180103dc0.fallback.annotated.c:168,342,354,852,950`：读取旧状态→Settler→StateValidator→写truth文件→写章节运行状态。
- `native/0x180052020.fallback.annotated.c:1490-1592`：结算响应的POST_SETTLEMENT、RUNTIME_STATE_DELTA和旧格式状态区块分开解析。
- `native/0x180298040.fallback.annotated.c:68-90`：生成增量应用结果后保存结构化状态。
- `native/0x1802d4080.fallback.annotated.c:120-298,602-974,1053-1153`：在旧快照上应用增量；检查章节倒退、摘要章节不匹配/重复；依次处理hook、currentState和summary。
- `native/0x1802db600.fallback.c:39-80`：旧/新任一状态为resolved，合并状态仍为resolved；否则处理推进状态。原始常量对应state_reducer#45/#71。
- `constants/novels--prompts--inkos--settler_prompts.json`：#6明确只输出增量；#27明确不能删除与本章无关的旧hooks，不能把尚未发生的事件写成事实。

所有fallback为本地Ghidra，不冒充REA服务证据。

## 已确认的CM差异与本步修复范围

当前CM的truthSettlement是最近一次模型响应；continuity.go下一章直接使用它。模型只返回本章新增内容时，上章未提及的旧记录会丢失。修订阶段的UPDATED_STATE仍只保存证据，由独立结算处理实际状态。

本步保留CM现有JSON字段，做明确标记的增量适配：facts按CM事实id合并，人物变化与时间事件按相同记录去重并保留历史，foreshadowing按CM thread合并；未提及的条目不删除，resolved不可被open重开。空数组表示无新增，不清空历史。同一增量中重复事实id或伏笔thread拒绝，避免顺序决定冲突结果。合并结果供记忆阶段、持久化和下一章读取；保留本次原始增量及适配版本作为证据。旧0/1/2/3版本不改变行为，新任务版本4启用。

这是CM既有数据结构的兼容适配，不是原版RUNTIME_STATE_DELTA全部字段的迁移。原版currentState字段/事实有效期、hookId及候选准入、章节摘要去重、支线/情感/人物矩阵、数值账本及StateValidator模型协议仍需逐项恢复实现，不能标为完成。

## 输入、输出、边界、失败

输入：上一章保存的CM truthSettlement及本章已校验的结算增量。输出：合并后的相同JSON结构，附原始增量证据。章节正文必须先通过现有审稿/修订，再结算。
空增量保留旧记录；同id事实只更新该事实，其它不变；同thread伏笔更新且resolved保持终态；重复增量应用不重复追加历史；旧快照中不存在的数组按空数组兼容。
前文JSON无法解析、前文条目无效、本章重复标识或合并校验失败均明确失败，不写正文/状态，不自动调用模型重试。取消仍在写入前检查，沿用章节与记忆的数据库事务和版本绑定。

- [x] 恢复责任链并形成规格
- [x] 重现未提及记录丢失
- [x] 实现版本化兼容适配与原始增量保存
- [x] 测试空/重复/冲突/旧任务与跨章持久化
- [ ] 更新本地4001

原版写入顺序补充：Settler在0x180052020.fallback.annotated.c:2011调用apply_settler_runtime_delta；0x180298040.fallback.annotated.c:68-90会保存结构化快照。这发生在外层StateValidator之前，不能从外层函数名推断原版全流程是原子提交。CM继续沿用MySQL正文/记忆事务，不复制这一中途写入时序。
