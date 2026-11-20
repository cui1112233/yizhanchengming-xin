# 全篇短篇缺章补写：已确认合同与待补证边界

状态：本轮已按下述合同实现，正在做数据库、页面和本地4001验收。原版完整helper、模板和函数行号见 `docs/verification/novel-full-draft-20261120/FINDINGS.md`；旧取证见 `docs/verification/novel-short-recovery-20261120/FINDINGS.md`。本合同的补写是实际缺章补写，与旧JSON结构修复分开。

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
- 补写仍复用writer system及仿写附加约束；本轮进一步反编译len包装helper确认：max_tokens按当前缺章数量和每章目标字数估算。此前按总章数的记录错误，本轮已更正。首稿和第二稿按总章数。
- 预算公式max(12288, ceil(章数×每章目标字数×2.2)+4096)。参数：首稿temperature=0.58，补写0.68，第二版修订0.45；正文审稿0.3/8192 tokens，包装0.45/4096 tokens。本轮新协议固定这些已取证参数，阶段模型仍用我们自己的配置。
- 通用临时网络错误默认最多3次传输尝试，是另一层限制，不等于最多3次模型调用。本轮没有把CM现有明确结果重试/未知结果暂停保护改成无条件自动重试。
- 第二版修订采用补齐后的完整首稿作assistant上下文。新协议只有一次第二版；旧分段任务及长篇原有多轮设置继续有效，页面明确范围。

## 失败与保留

- 最后仍缺章时进入最终验证失败，不能包装为完整成品。
- 不完整或解析失败的第二版不能覆盖完整首稿；保留v001、第二稿原文和未采用原因。已知无效候选继续用首稿包装；取消、未知结果、鉴权/限流、持久化失败继续明确失败或暂停，不吞掉所有异常标成功。这是与原版宽泛Exception捕获的明确安全差异。
- 解析采用已恢复helper规则：书名、开篇钩子、章标题取首块（含首空块）；章正文取最后非空块；再按重复TITLE和Markdown回退、清理。原版宽松Markdown回退可能重复首章，按证据保留并测试；正常模板要求CHAPTER标记，不冒称该回退能精确识别所有Markdown分章。
- 取消检查在已取证的入口和首稿前有显式调用；不能凭本轮静态范围声称原版所有步骤实时取消。CM应保留现有更细粒度取消检查。

## CM输入映射、兼容和实际产物

- 仅新建short_imitation快照写short_full_draft_version=1，旧缺省0维持旧分段协议，未知版本在调用前拒绝。
- 保留CM已有蓝图确认和JSON策划/大纲审核；已确认蓝图+JSON大纲转换为outline_markdown，所选方向+用户创作资料映射到direction。这是明确的CM规划适配，未宣称原版此处也走相同JSON步骤。
- 写作system由后台冻结的完整原版writer模板、仿写约束及用户short_fiction/imitation写作要求组成。37维度与自定义审稿只注入实际文字审稿/修订用途；通过分、分段字数整理及多轮设置不强加给整篇正文。
- 所有原版完整system/user模板在后端prompt/novelcreation/short_full，由后台系统预设管理；新增9项，合计39项。发布值冻结到任务，归档模板不偷偷恢复。未知占位符在发模型前报错；用户资料里的花括号是原样文本，不递归当模板执行。
- 持久化fullDraft包括currentVersion、revisionStatus、revisionWarning、review、completionAttempts、missingChapters。drafting/pending/adopted/rejected分别表示首稿阶段、首稿保留待修订、采用第二稿、第二稿未采用；运行是否已结束仍以任务/作品状态为准。
- 保存drafts/v001.md、drafts/v002.md或v002-rejected.md，缺章失败保留v001-incomplete.md；review及package原始文本分别保存，版本文件只读。
- 原版包装title/intro/selling_points/cover aliases和逐行去一个-/ *规则已实现；CM保持最终采用正文标题，简介/卖点进入推荐文案，封面提示词独立保存。
- 每次已成功调用按原有数据库调用记录恢复，第二稿/包装失败重试不重新生成已成功的正文。完整首稿持久化失败时立即停止，不继续花费后续调用。
- 范围保护：章节1..50、UTF-8、2M UTF-16文本上限沿用CM资源限制。计字按原版实际行为统计全部非空白字符，并非只数中文。

## 验收方法和剩余差异

正常、空输出、重复标记、缺章补齐/三次仍缺、无效第二稿、保存失败、取消、限流、未知付费结果、服务重开分别覆盖；本轮使用固定假回复、真实本地HTTP传输、隔离MySQL和浏览器，未调用付费模型，不能据此声称真实模型文章质量已验收。全软件UI像素和其他创作模式不是本轮完成范围。
