# 小说恢复与 Writer 注入：补充静态证据

日期：2026-11-20。只读分析；未运行洛水目标、未调用模型、未修改业务代码。

## 来源及方法

- 本轮重算原件 `F:/软件/洛水2026/洛水2026/modules/novels.pyd` SHA-256：`b5b22086c64851e0e94fc70a509cddbaf4ec144ac5beedc64e3a9a8caa40e57d`，与既有不可变快照相同。
- REA 当前子会话 `binary_session` 为 `open:false`；未重复此前全量 import 超时操作。
- 使用既有 `novel-verified-native.gpr`，Ghidra 12.1.4 `-process novels-b5b22086.pyd -readOnly -noanalysis`。已有函数导出脚本每函数限 90 秒；完整日志留在同目录。`0x1802095a0` 单函数 C 导出超时；改用原文件静态反汇编，不再重复。
- `*.fallback.c` 是本轮分析员运行的 Ghidra 结果，**不是新 REA Evidence 记录，也没有冒用 Evidence ID**。`*.assembly.txt` 从经过 SHA 校验的原文件，通过 PE exception table 对应函数范围和 Capstone 只读解码生成。
- `*.annotated.c` 和汇编尾注使用既有 `function-bindings.json` 的模块常量池地址与 `constants/*.json` 的索引映射；尾注是分析员注释。`annotate_evidence.py` 可重现。
- Ghidra 对部分通用属性调用的参数恢复不完整；因此 `read_text`、`strip`、`get` 等动态方法名以汇编 RIP 常量载入为准，不能只读 C。

## 1. Writer 注入：已回答

观察与推导：

1. `_make_augmented_ctx` 构造 provider 闭包。provider (`0x1800d4df0`) 先取得 `self._ctx(agent)`，再比较 `agent == "writer"`。非 writer 分支直接返回原 ctx；只有 writer 分支导入 `dataclasses.replace` 并建立 `_AugmentedClient` 替换 ctx.client。
   - `0x1800d4df0.fallback.annotated.c:78-137`；替换 client 在该文件 `:348-364`。
   - 汇编 `0x1800d4df0.assembly.txt:84-106`，比较后 `jne 0x1800d58e0`。
2. wrapper 的 `chat` 和 `chat_stream` 都先从 messages 调 `_inject`，然后转发给 `_real.chat` / `_real.chat_stream`。
   - `0x1800d6450.assembly.txt:44-84`；`0x1800d6820.assembly.txt:44-84`。
3. `_inject` 的实际判断是：messages 为真，并且 **第一条消息** `.get("role") == "system"`。不是搜索任意 system 消息，也没有正文/审稿阶段名称的判断。
   - `0x1800d5c20.fallback.annotated.c:46-58,246-285`。
   - `0x1800d5c20.assembly.txt:71-117` 证明 `.get("role")` 和 system 比较。
   - 比较辅助函数 `0x1803836d0.assembly.txt:32-36` 的 richcompare 操作数为 `2`（Python 相等比较）。
4. 满足判断后复制消息容器，并用原第一条 content + `"\n\n"` + 闭包 addendum，替换第一条消息。未满足时返回原 messages。不改变消息 role，不给非 writer 客户端注入。
   - `0x1800d5c20.fallback.annotated.c:58-132,190-214,230-240`。
   - 原 content、分隔符、闭包变量的直接汇编载入在 `0x1800d5c20.assembly.txt:188-240`。
5. `_produce` 把此 provider 交给后续 `ShortFictionRunner`，不是只围绕一次正文调用临时建立它。
   - `0x1800cd2d0.assembly.txt:1695-1740`。

含义：原版的边界是后续流水线请求的 **agent 身份 writer + 首条 system**。不能用“只在正文阶段追加”来替代这一规则；也不能因此把补充约束加到 architect/auditor 等其他角色。

尚未证明：ShortFictionRunner 每个具体工作步骤选择哪个 agent，仍需结合其自身角色调用证据；本轮不把所有后续阶段都直接归为 writer。

## 2. 确认恢复时 model 采用规则：已回答到模型选择入口，存在优先顺序

1. API `novels_confirm_pipeline_checkpoint` 从本次请求取 `.get("model")`，作为关键字 `model` 传入 `_agent_service(...).confirm_pipeline_checkpoint(...)`。
   - `0x1800727b0.assembly.txt:613-680`。
   - `0x1800727b0.fallback.annotated.c:380-435`。
2. 服务端 `AgentSessionService.confirm_pipeline_checkpoint` 的仿写分支，调用 pipeline `run_imitation_short`，将恢复载荷作为位置参数，本次传进来的 model 作为关键字 model。
   - `0x1802095a0.assembly.txt:1417-1480`；参数槽 `[rbp+0x50]` 被放入 model 的关键字参数数组。
   - 此函数 C 导出超时，不依赖不完整 C 文本作判断。
3. pipeline 创建 ctx provider 闭包。该 lambda 捕获 model，调用 `_ctx_for_short_fiction(model, agent)`。
   - `0x180124e30.assembly.txt:151-177`（捕获两个值并构造 lambda）；
   - `0x180125650.fallback.annotated.c:79-104` 和 `0x180125650.assembly.txt:92-115`（捕获 model + 请求 agent 作为两个位置参数）。
4. `_ctx_for_short_fiction` 并非直接固定采用本次 model。它计算：
   ```python
   effective = settings.get_model_route("short_fiction", model)
   if not effective:
       effective = settings.get_model_route(agent, model)
   return self._ctx("short", effective, agent=agent)
   ```
   - `0x1800e5280.fallback.annotated.c:53-121`；方法名的汇编在 `0x1800e5280.assembly.txt:48-138`。
5. `get_model_route(route, fallback)` 优先取 `load().get("modelRoutes", {}).get(route)`，转字符串并 strip。该值非空就先返回；只有为空时才检查 fallback，fallback 也需 strip 后非空，否则返回 None。
   - `0x1802239f0.fallback.annotated.c:47-165`（配置路由优先返回）、`:167-248`（fallback）。
   - `0x1802239f0.assembly.txt` 保留方法名载入；不存在“只要请求带 model 就越过已配路由”的分支。

含义：恢复请求的当前 model **确实进入后端恢复链并作为模型路由的后备值**，但已有 `short_fiction` 路由会优先。当前 model 非空时，第一次 get_model_route 通常就返回该后备值，此时不会再查 agent 路由。只有第一次结果为空，才继续查 agent 路由。因此不能宣称恢复后必然采用前端当前下拉模型；应实现同样的配置优先与后备关系。

尚未证明：最终 AI 客户端发出的请求里实际 model 名称；本轮未执行目标，也未消费额度。`self._ctx` 内部服务商别名转换不属于本轮已证明范围。

## 3. 编辑后蓝图重新读取：已回答

`_resume_or_run_blueprint` 每次进入均执行：
```python
text = FileRepository.read_text(book_path / "imitation" / "blueprint.md", "").strip()
if text:
    # 记录已缓存、阶段完成
    return text
# 空文件或纯空白则走 architect 重新生成
```

- `0x1800d22e0.assembly.txt:65-156`：FileRepository、read_text、imitation、blueprint.md、默认空字符串、strip。
- `0x1800d22e0.fallback.annotated.c:65-139`：读取后判断非空；`:503-547`：非空分支记录 `Resuming: blueprint already exists`、已缓存完成并返回刚读出的内容。
- 这证明恢复函数从磁盘拿当前内容，不是只取暂停时内存中的蓝图。编辑内容必须已保存到同一个路径；本轮未证明编辑器保存 API 与路径权限的完整链路。
- 空内容行为是重新生成，不是明确报错；不能把“原版会停下报蓝图空”写成事实。

## 未扩大结论的地方

- 以上均为静态代码观察和推导，不能称为实际恢复执行记录。
- 重复确认/并发恢复/取消边界未在本轮证明。
- 额外导出的 scheduler.resume (`0x1802a61a0`) 属其他调度模块，未用于解释流水线确认。
- 所有新增输出仅位于本目录；不修改原件、Ghidra 原项目、业务文件，不提交。

## 4. 最后补查：修订正文、缺失章节补写的真实角色

此项直接查看 `ShortFictionRunner._produce`（`0x180165140`）原件汇编，不根据函数名字猜角色。

| 工作 | 实际角色及消息 | 本轮文件行号 |
|---|---|---|
| 首稿正文 | `_chat` 的第一个位置参数明确为 `writer`，消息从 system 开始 | `0x180165140.assembly.txt:538-590` |
| 缺失/空章节补写 | `_chat` 的第一个位置参数明确为 `writer`，消息从 system 开始；同分支使用 `build_short_fiction_draft_continuation_user_prompt` | `0x180165140.assembly.txt:1399-1442,1854-1988` |
| 修订正文 | 阶段“修订正文”后 `_chat` 的第一个位置参数明确为 `writer`，第一条 system，后面原始 user、assistant 首稿、user revision followup | `0x180165140.assembly.txt:2852-2977,3010-3146` |
| 审阅正文 | `_chat` 明确载入 `auditor`，第一条 system | `0x180165140.assembly.txt:2494-2532` |

角色参数确实用于取得上下文，而不只是日志：`_chat` 创建的重试回调 `0x180160f00` 取 `self._ctx`，用捕获的 `agent` 调它，再取 `.chat`，传 messages、stage_label、kwargs。直接证据见 `0x180160f00.assembly.txt:78-127,153-258`。

因此：正文修订由 writer 执行，不能把它当成 reviser 而排除仿写 addendum。首稿、缺失章节补写、正文修订都会满足前述 writer 包装规则，并具有首条 system。正文审阅是 auditor，不应加此 addendum。

对 CM 当前“ShortBodyPrompt + ShortBodyRevision 加 addon”的对应判断：正文与正文修订这两处角色边界吻合。空章节补写也必须沿用带 addon 的 writer system；如果另走未加 addon 的续写 system，就会漏注入。本段仅给出原版行为，未检查或修改 CM 业务代码。

**独立长度整理仍未证实。** 本轮仅在 `novels.pipeline.short_fiction_runner` 已完整解码常量池以及 `run`、`_produce` 的直接 RIP 常量调用范围检查 `normalize` / `length` / `reviser`，没有独立长度整理的入口；发现的是目标每章字数、token 估算及缺失章节补写。模块另有通用 `novels.services.length_normalizer.normalize_chapter_length` 不等于本条仿写短篇路径会调用它，本轮不把其他路径的身份套到这里。不能据此增加一个“原版 writer 长度整理”规则，也不能把这个有界检查写成整个原版不存在长度整理。

归档说明：Git版本仅清理行尾空白和文件末尾多余空行，不改内部行号、指令地址或内容。原始导出保存在 `_local-archive-20261028/novel-blueprint-repair-20261120/native-raw/`；原始及归档SHA见 `normalization-index.json`。
