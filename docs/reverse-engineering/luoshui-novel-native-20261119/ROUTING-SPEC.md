# 原版自动修订分流规格（2026-11-19）

固定目标 SHA256：b5b22086c64851e0e94fc70a509cddbaf4ec144ac5beedc64e3a9a8caa40e57d。本次是静态推导，尚未进行原版真实模型对照。

## 证据

- `native/0x18004a310.fallback.annotated.c:408-498`：修订入口取得审稿问题，调用 resolve_auto_output_mode，再将其传入自动系统提示词。
- `native/0x1801c29d0.fallback.annotated.c:91-96,229-543,626-839`：空列表、显式范围优先及模式回退。
- `native/0x1801c42c0.fallback.annotated.c:130-180` 与 `0x1801c4990.fallback.annotated.c:130-180`：repairScope 优先于 repair_scope；分别与 structural/local 精确比较，不转小写。
- `native/0x1801c5060.fallback.annotated.c:48-160`、`0x1801c5cd0.fallback.annotated.c:48-160`：category + 空格 + description，逐个正则 search。
- `native/0x1801cb170.fallback.annotated.c:383-857,873-1507`：9条 LOCAL_ONLY_PATTERNS、12条 STRUCTURAL_PATTERNS，均使用忽略大小写。原始表达式在 reviser_prompts 常量146、148-167，不扩写近义词。
- `native/0x1801c8710.fallback.annotated.c:408-475`：rewrite-only追加原常量79，patch-only追加原常量80；不能同时追加两者。
- `native/0x18004eb10.fallback.annotated.c:185-240`：原版按分流选取对应输出；错误模式或无有效正文时原版保留原文。
- 比较辅助函数 `0x180386310.fallback.c:25-38` 使用 Python rich compare 操作3（不等），`0x180383520.fallback.c:25-42`使用操作2（相等）；`0x1803839a0.fallback.c:12-36`确认长度比较是相等。

以上 .fallback 文件为本地 Ghidra 输出，和原有 REA 服务证据分开标注。

## 输入和输出

输入为结构化审稿 issues，保留 severity/category/description，以及可选 repairScope、repair_scope。输出为 patch-only / rewrite-only / allow-full；输出模式进入同一次修订模型请求和结果验证，使用CM已有模型配置。

按原生分支推导的顺序：

1. issues为空：allow-full。
2. `scoped = [i for i in issues if (lower(severity) != 'info' and truthy(repairScope)) or truthy(repair_scope)]`。
3. scoped 中任意有效范围为 structural：rewrite-only。
4. scoped 数量等于非info问题数量，且每项范围为 local：patch-only。
5. 回退到所有非info问题；若为空：patch-only。
6. 任意问题匹配结构规则：rewrite-only；否则全部匹配局部规则：patch-only；否则 allow-full。

注意原版第2步的 and/or 优先级使 info + repair_scope 也可能进入 scoped。这是静态观察到的细节，明确保留，不把它描述为设计意图。repairScope和repair_scope以truthy的前者优先；scope值精确匹配，不自动修正常量大小写。规则只针对category和description，不扫描正文、evidence或附加提示词。

## 边界与失败提示

结构与局部同时出现，按上述顺序选择；未知分类不猜测。新任务冻结review_cycle_version=3；版本0/1/2保持之前协议和请求指纹。选择方式、判断依据与原始输出保存在修订快照。

CM现行明确失败规则继续保留：模型在patch-only返回整章正文，或在rewrite-only只返回补丁，报“本轮要求局部修改/整章重写，模型返回了不匹配的结果，原文未修改”；不自动增加收费重试。这与原版“保留原文继续返回”的处理不同，是当前产品的已登记兼容差异，不能声称异常处理完全等价。旧JSON整章响应只在allow-full/rewrite-only兼容；patch-only不能用整章JSON冒充局部修改。

## 执行登记

- [x] 固定原包及恢复分流函数、原始规则、调用者和消费者
- [x] 按证据形成规格
- [x] 实现分流、保留范围字段、提示词及输出验证
- [x] 优先级、空/混合/未知、错误输出、取消、旧任务恢复测试
- [ ] 实际MySQL保存与本地4001版本验收

兼容范围：CM审稿接口仍使用既有low/medium/high/critical等级；info/warning在本轮仅作为原版分流算法的静态边界测试，不声称整个审稿等级协议已迁移。旧任务0/1/2还须在审稿解码后清除新增scope字段，以保持旧请求指纹；已先测出失败再修复并验证。
