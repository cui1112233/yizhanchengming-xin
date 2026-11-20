# 原版短篇补写与四消息修订合同

日期：2026-11-20。只读、未运行目标、未调用模型、未修改业务源码、未提交。原件 SHA-256 为 `b5b22086c64851e0e94fc70a509cddbaf4ec144ac5beedc64e3a9a8caa40e57d`；前端原 bundle SHA 为 `7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6`。

## 来源与覆盖

复用已有 Ghidra 12.1.4 项目，`-readOnly -noanalysis`，不重新 import。原 `_produce` C 导出本轮超时，未重复；改为一次导出已有 Listing 的整个函数 body。`0x180165140.full.assembly.txt:2` 明确完整 body 为 `[0x180165140,0x18016b90d]`。较小 helpers 的完整 C 已成功导出。均为分析员本地结果，不是新 REA Evidence，不新增/冒用 REA ID。

注意：初始 `*.assembly.txt` 来自 PE 单个 unwind 范围，部分 helper 会提前结束。存在 `*.full.assembly.txt` 时以该文件为准；它来自保存的 Ghidra 函数完整 body。C 的动态方法参数可能被恢复器省略，方法名通过完整 Listing 的 RIP 常量载入确认。

## 1. 原首稿、缺章与补写

### 请求输入与顺序

首稿保存 `writer_sys = self._writer_sys_prompt()`，随后 writer 请求两条消息：system 为该字符串，user 为 `build_short_fiction_writer_user_prompt(draft_input)`。draft_input 含 direction、outline_markdown、chapter_count、chars_per_chapter。首稿 temperature=0.58，max_tokens 由 `estimate_short_fiction_max_tokens(chapter_count, chars_per_chapter)` 给出。

证据：`0x180165140.full.assembly.txt:502-721`，关键词同一条调用中明确传给 `_chat`，其实际 agent=writer。

### 解析形态

`parse_short_fiction_batch_draft(raw, expected_chapters=N)` (`0x180153e50`) 返回 `ShortFictionBatchDraft(story_title, opening_hook, chapters, raw_content)`。它按 1..N 逐章构造条目，不是把 LLM 输出只当一个当前段 JSON。

- 标题来自 SHORT_FICTION_TITLE block，回退 Markdown 首标题，最终默认“未命名短篇”；开篇 hook 接受 SHORT_FICTION_OPENING_HOOK，再回退 OPENING_HOOK。
- 逐章标题：CHAPTER n TITLE block，回退 `_extract_markdown_chapter_title(raw,n)`；再由 `_normalize_chapter_title` 规范化。
- 逐章正文：调用 `_extract_last_nonempty_tagged_block(raw,"CHAPTER n CONTENT")`，空则尝试 `_extract_duplicate_title_tagged_chapter_content(raw,n)`，再尝试 `_extract_markdown_chapter_content(raw,n)`，最后空字符串；该结果送 `_sanitize_chapter_content`，并记录 `_count_zh_chars`。
- raw_content 留存原始聚合文本，后续拼接补写和修订 assistant 使用它，而非仅使用渲染后的 chapter 文本。

证据：`0x180153e50.fallback.annotated.c:113-239,333-601,606-1030,1080-1230,1265-1420`；完整指令另见 `0x180153e50.full.assembly.txt`。

限定：上述确认了实际调用顺序与分支。四个 extraction/sanitize helper 的全部正则边缘行为本轮没有进一步恢复，不能把“按名字猜测的全部去标签规则、重复 block 策略”当完成的实现合同。

### 缺章判定

`find_empty_short_fiction_chapters(draft)` (`0x180156ff0`) 遍历 draft.chapters，检查 `chapter.content.strip()`；为空就收集该 chapter.number。不是按字数少于目标、置信度或审稿评分触发补写。

证据：`0x180156ff0.assembly.txt:64-204`、`0x180156ff0.fallback.c:145-219`。

### 次数、补写请求与合并

`SHORT_FICTION_DRAFT_COMPLETION_ATTEMPTS=3` 有原件模块赋值证据：`parser-module-selected-bindings.assembly.txt:2-5`。首稿之外，最多 3 次缺章补写；每次都重新检查 missing，已补齐就跳出。

补写输入对象 `ShortFictionDraftContinuationPromptInput`：

- direction：原 direction。
- outline_markdown：原 outline_text。
- chapter_count / chars_per_chapter：原目标。
- existing_draft_markdown：`render_short_fiction_draft_markdown(draft_v1)`。
- missing_chapters：当前 missing 编号列表。

消息为 `[system(writer_sys), user(build_short_fiction_draft_continuation_user_prompt(cont_input))]`，agent=writer，temperature=0.68；max_tokens 使用 `estimate_short_fiction_max_tokens(len(missing), chars_per_chapter)`。2026-11-20后续聚焦复核已更正前轮误判：0x18016682f 的 helper 0x180375370 对当前missing取长度，再作为estimate第一参，证据见 `../novel-full-draft-20261120/0x180375370.fallback.c`；补写用户提示里的chapter_count仍是原总章节数。

响应处理：

```python
merged_raw = draft_v1.raw_content.strip() + "\n\n" + cont_raw.strip()
draft_v1 = parse_short_fiction_batch_draft(merged_raw, expected_chapters=N)
missing = find_empty_short_fiction_chapters(draft_v1)
```

证据：`0x180165140.full.assembly.txt:1053-1110,1399-1474,1559-1788,1856-1990`；输入关键字 tuple 来自已解码常量 `novels.pipeline.short_fiction_runner#138`。

首次发现 missing 后会写 `drafts/v001-partial.md`；补写阶段会上报“补写缺失章节：…（第…次）”。循环后调用 `validate_short_fiction_draft_for_final(...expected_chapters=N)`：目标章节数不符或仍有空章都会抛 `ShortFictionDraftValidationError`，不继续打包成品。验证证据：`0x180157690.fallback.annotated.c` 的 count 检查和 find_empty 分支；调用点 `0x180165140.full.assembly.txt:2101-2130`。

## 2. 正文修订的四消息合同

原版此处是一次“第二版完整正文”修订，不是每章片段修订，也没有多轮修订循环。角色 writer，temperature=0.45，max_tokens 仍按 chapter_count/每章目标估算。

消息严格按以下顺序构造：

1. system：同一个 writer_sys。
2. user：重新调用首稿 `build_short_fiction_writer_user_prompt(draft_input)`，输入仍为原首稿 draft_input。
3. assistant：`draft_v1.raw_content.strip()`，是首稿经过最多 3 次缺章补写、合并重解析后得到的 draft_v1。不是未补齐的最早 draft_raw，也不是修订 response。
4. user：`build_short_fiction_draft_revision_followup(revision_input)`。

`revision_input` 构造字段为 direction、outline_markdown、chapter_count、chars_per_chapter、review。review 为此前正文审阅文本。

证据：`0x180165140.full.assembly.txt:2854-3268`。assistant 原文读取在 `:3021-3057`，首稿用户 builder 在 `:2979`，最后 user builder 在 `:3148`。该区域仅有一次修订请求；新结果解析成 draft_v2 后验证；失败时保留 v001，并记录“第二轮改稿未采用”“系统没有用不完整或解析失败的改稿覆盖完整首稿。”，不是让无效第二稿覆盖首稿。证据 `:3312-3859`。

### 最后一条 user 的完整构造

`0x1801d8140` 原 builder 实际只读取 `revision_input.review.strip()` 与 `revision_input.chapter_count`；direction、outline_markdown、chars_per_chapter 在此 followup 不再次拼入。

自然语言模板在同目录 `revision-followup-template.txt`，保留全部固定行和两个动态占位：review.strip()、_chapter_output_blocks(chapter_count)。由常量 72-80 加输出格式常量组成，最后 `"\n".join`。完整 C 证据：`0x1801d8140.fallback.annotated.c:37-229`；动态字段直接指令证据 `0x1801d8140.full.assembly.txt:85-95,238-265`。

CM 若保留按章分段 JSON：可以适配“原 user + assistant 当前首稿 + 明确审稿 followup”的会话关系，但输出格式、作品/片段边界必须明确属于 CM 适配。原版是全篇 Markdown blocks，不能把 CM 内部 revision preset 的整个 JSON 包装冒充原版 review 文本，也不能宣称原版支持“每一轮始终首稿”之类未存在的多轮逻辑。

## 3. 传输重试与取消的边界

- `_chat` (`0x180160790`) 把闭包传 `retry_short_fiction_call(...label=stage_label...)`。闭包每次用捕获的 agent 取得 ctx，再 `.chat(messages, stage_label=..., **kwargs)`；messages/kwargs 不在传输重试之间改成另一个生成任务。
- `retry_short_fiction_call` 是从 utils.llm_retry 导入 retry_llm_call 后别名绑定；见 `parser-module-selected-bindings.assembly.txt:43-87`。
- 通用 retry_llm_call (`0x18035cbe0`) 对非暂时错误直接重新抛出；达到 max_attempts 也抛出。默认 max_attempts=3、base_delay=2.0 的构造字典在 `novels.utils.llm_retry` 常量#38。暂时错误分类来自 is_transient_llm_error，不是所有异常都重试。sleep 前记录 attempt/max_attempts/delay。
- 这“3次传输尝试”和上面的“首稿后最多3次缺章补写”是两层不同限制，不能混为一个最多3次模型调用。
- `_produce` 的完整 Listing 中显式 `_cancelled` 检查仅在入口/outline阶段后、开始首稿前两处（`:101-145,356-398`）；对应 OperationCancelledError。未在补写循环、四消息修订处看到新的本层显式检查。

未知：ctx/AI client 内部是否在每次 chat 或传输重试中额外检查取消，以及 sleep 的精确退避公式，本轮未继续扩大分析。不能声称原版取消必然实时生效或不存在内部取消；CM 应保留自己现有取消和付费保护。

## 4. 保存后是否有独立短篇重新终审

有界结论：**本轮检查的原版 ShortPanel、Studio saved 回调和 NovelsApi 直接方法库存里，没有找到独立“编辑稿件后重新 editorial review”的入口。** 不新增该原版功能。

- `original-short-panel.saved.txt:125-156`：保存调用短篇文件 API，成功后只更新显示内容、提示“已保存”、emit saved。
- 同文件 `:278-304`：预览组件只有 onBack/onSave；章节、成品、推荐的按钮是打开/编辑文件，没有重审按钮。
- `original-studio-saved-excerpts.txt` 原文件行 `1624-1630` 将 saved 交 gt；原文件行 `1094-1098` 的 gt 仅重新读取 shortStory 数据，不发起模型工作。
- `original-api-inventory.json` 是已保存原件函数构造绑定里 `novels.api.novels_api` 的全部直接方法。短篇相关只有 get/read/save、run_imitation_short、confirm_pipeline_checkpoint；通用 novels_audit_chapter 存在，但不是已证明的独立短篇全篇 editorial 重审入口。
- 两个前端恢复文件的来源、SHA256、行数记在 `frontend-source-identities.json`。这些是同一已确认 bundle 的静态恢复片段，外部依赖与其他隐藏/动态入口仍可能未覆盖。

本轮未运行 UI、未执行目标代码，因此不把有界未找到表述为整个产品永远不能重新审稿。
