# 全篇短篇缺章补写：已确认合同与待补证边界

状态：本轮已取证，尚未完整移植。原版证据和函数行号见 `docs/verification/novel-short-recovery-20261120/FINDINGS.md`。不能把CM当前JSON结构修复称为已完成此合同。

## 输入

- 原方向、完整Markdown大纲、目标章节数chapter_count、每章目标chars_per_chapter。
- 首稿原始回复，原writer system及writer上下文。
- 每次补写使用当前已解析首稿的Markdown渲染结果existing_draft_markdown和当前missing_chapters编号列表。

## 输出

- parse_short_fiction_batch_draft(raw, expected_chapters=N)输出story_title、opening_hook、chapters、raw_content；按1..N构造章节。
- 每章的标题/正文有带标记block与Markdown回退路径；保存聚合raw_content，供后续拼接和第二版修订使用。
- 空章节由chapter.content.strip()为空判定，与审稿分数或少于目标字数不同。

## 边界

- 首稿之外最多3次缺章补写。补齐即停止，不固定执行满3次。
- 每次响应以当前raw_content.strip()+两个换行+cont_raw.strip()合并，再对完整聚合文本重新解析、重新计算missing。
- 补写仍复用writer system及仿写附加约束；max_tokens按原总章节数和目标字数估算，不把missing数量冒充总章节数。
- 参数：首稿temperature=0.58，补写0.68，第二版修订0.45。CM当前采用用户配置温度，不能称这些参数已逐项还原。
- 通用临时网络错误默认最多3次传输尝试，是另一层限制，不等于最多3次模型调用。本轮没有把CM现有明确结果重试/未知结果暂停保护改成无条件自动重试。
- 第二版修订采用补齐后的完整首稿作assistant上下文。原版可见路径只有一次第二版；CM可配置多轮是独立适配。

## 失败与保留

- 最后仍缺章时进入最终验证失败，不能包装为完整成品。
- 原版不让不完整或解析失败的第二版覆盖完整首稿；保留v001并记录未采用改稿。CM该回退及可见性仍需对照实施，不能吞掉所有异常后直接标成功。
- 原版的精确用户报错文案、各提取/清理helper的全部正则边界尚未完整恢复，不编造原版错误字符串。
- 取消检查在已取证的入口和首稿前有显式调用；不能凭本轮静态范围声称原版所有步骤实时取消。CM应保留现有更细粒度取消检查。

## 下一次实施前必须补证

1. 重复CONTENT/TITLE block、纯Markdown章节、空白/缺号/重复章号的确切提取策略。
2. sanitize内容清理与中文字符计数规则，不按函数名猜实现。
3. 无效改稿回退和传输错误在外层任务状态上的具体分支。
4. 全篇输出与CM现有按章持久化、版本记录和恢复的迁移方案；不能直接替换旧任务请求导致已完成章节重复生成。
