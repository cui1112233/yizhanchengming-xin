# novels.prompts.inkos.reviser_prompts: recovered constant literals
Source SHA256: b5b22086c64851e0e94fc70a509cddbaf4ec144ac5beedc64e3a9a8caa40e57d
These are ordered constants, not reconstructed function bodies.

## Constant 56 | bytes 4213971..4214098
```text
【LANGUAGE OVERRIDE】ALL output (FIXED_ISSUES, PATCHES, REVISED_CONTENT, UPDATED_STATE, UPDATED_HOOKS) MUST be in English.


```

## Constant 70 | bytes 4214478..4214829
```text


ROUTING: The reviewer's blocking issues are structural / semantic (character collapse, mainline drift, missing payoff, timeline break, unpaid hook, memo drift, etc.). You MUST output REVISED_CONTENT — do not emit PATCHES, they cannot fix this class of problem. If you cannot safely rewrite, say so in FIXED_ISSUES and leave REVISED_CONTENT empty.
```

## Constant 71 | bytes 4214830..4215084
```text


ROUTING: The reviewer's blocking issues are local (wording, paragraph shape, fatigue word, information boundary, knowledge pollution). You MUST output PATCHES only — do not rewrite the whole chapter. If patches are not possible, leave PATCHES empty.
```

## Constant 75 | bytes 4215194..4215858
```text


PATCHES and REVISED_CONTENT serve different problems — choose by problem type, not preference:

PATCHES — for local text issues (wording, dialogue, AI-tell phrases, small continuity errors).
  Each PATCH quotes the passage to change (a sentence, a paragraph, or multiple paragraphs) and provides a replacement. Untouched text stays exactly as-is.

REVISED_CONTENT — for whole-chapter issues (length compression, structural rewrite, pacing restructure, major plot realignment).
  Outputs the full revised chapter. When Critical issues include length or structural problems, you must use REVISED_CONTENT — patches cannot compress or restructure a chapter.
```

## Constant 76 | bytes 4215859..4216063
```text


If Critical issues include both local and whole-chapter problems, use REVISED_CONTENT (it addresses everything in one pass).

Revision principles:
1. Fix root causes — do not apply superficial polish
```

## Constant 77 | bytes 4216064..4217734
```text

2. Hook status must stay in sync with the hooks board. If hook debt briefs are provided, preserve hook payoff scenes
3. Do not alter the plot direction or core conflicts
4. Preserve the original language style, rhythm, and pacing — do not compress transitional scenes or remove breathing room
5. Emotion through action (never "he felt angry" — show it). Values through behavior, not slogans
6. Different characters speak differently. No "everyone gasped in unison"
7. Escalate: bad things stack, each worse than the last

Cycle-aware revision:
- If this chapter should be "aftermath" but is still escalating tension, rewrite the densest conflict passage into a change-showing passage — who lost what, whose attitude shifted, what the new normal is
- If this chapter should be "climax" but has no clear payoff, find the closest scene to a reward and amplify it — make the promised release exceed reader expectations
- Daily passages that don't serve the main line: rewrite as "bait" — add a detail pointing to the future, a hint, a character reaction that seeds curiosity

Output format:

=== FIXED_ISSUES ===
(List each fix on its own line; if a safe local fix is not possible, explain here)

=== PATCHES ===
(Output local patches if applicable. Omit this section entirely if using REVISED_CONTENT)
--- PATCH 1 ---
TARGET_TEXT:
(Exact quote from the original that identifies the passage to change)
REPLACEMENT_TEXT:
(Replacement text for this passage)
--- END PATCH ---

=== REVISED_CONTENT ===
(Full revised chapter content — only when PATCHES cannot solve the problem. Omit this section if using PATCHES)

=== UPDATED_STATE ===
(Full updated state card)

```

## Constant 79 | bytes 4217786..4218122
```text


分流指令：reviewer 报告的阻塞问题属于结构/语义错（人设崩、主线偏、爽点缺、时间线错、伏笔未收、memo 偏离等）。你必须输出 REVISED_CONTENT——禁止输出 PATCHES，这类问题不能靠补丁修复。如果无法安全重写，在 FIXED_ISSUES 里说明并留空 REVISED_CONTENT。
```

## Constant 80 | bytes 4218123..4218352
```text


分流指令：reviewer 报告的阻塞问题属于局部错（措辞、段落形状、疲劳词、信息越界、知识污染）。你必须只输出 PATCHES——不要整章改写。如果做不出补丁，留空 PATCHES。
```

## Constant 83 | bytes 4218462..4219090
```text


PATCHES 和 REVISED_CONTENT 分别处理不同类型的问题——按问题类型选择，不是按偏好：

PATCHES——处理局部文字问题（措辞、对话、AI痕迹、小的连续性错误）。
  每个 PATCH 引用要修改的原文段落（一句、一段或多段皆可），给出替换文本。未涉及的内容保持原样。

REVISED_CONTENT——处理全章级问题（字数压缩、结构重组、节奏重排、重大剧情偏离）。
  输出修正后的完整正文。当 Critical 问题包含字数或结构性问题时，必须使用 REVISED_CONTENT——PATCHES 无法压缩或重构整章。
```

## Constant 85 | bytes 4219262..4220703
```text

2. 伏笔状态必须与伏笔池同步。如果提供了 Hook Debt 简报，必须保留伏笔兑现段落
3. 不改变剧情走向和核心冲突
4. 保持原文的语言风格、节奏和呼吸——不要压缩过渡段、不要删掉减速段
5. 情绪用动作外化（不写"他感到愤怒"，写动作）。价值观通过行为传达
6. 不同角色说话方式必须不同。禁止"众人齐声惊呼"
7. 坏事叠坏事，每层比上一层过分

小目标周期修稿指引：
- 如果本章应该是"后效"阶段但仍在加压，把最密集的冲突段落改写为展示改变的段落——谁失去了什么、谁的态度变了、新的常态是什么
- 如果本章应该是"爆发"阶段但没有明确兑现，找到最接近回报的场景并放大它——让承诺的释放超过读者预期
- 日常段落如果不服务主线，改写为"饵"：加入一个指向未来的细节、一句暗示、一个角色反应

输出格式：

=== FIXED_ISSUES ===
(逐条说明修正了什么)

=== PATCHES ===
(局部补丁——仅用于局部文字问题。有全章级问题时省略此区块)
--- PATCH 1 ---
TARGET_TEXT:
(从原文中精确引用要修改的段落)
REPLACEMENT_TEXT:
(替换后的文本)
--- END PATCH ---

=== REVISED_CONTENT ===
(修正后的完整正文——用于字数/结构/节奏等全章级问题。仅局部问题时省略此区块)

=== UPDATED_STATE ===
(更新后的完整状态卡)

```

## Constant 90 | bytes 4220795..4221150
```text
=== FIXED_ISSUES ===
(逐条说明修正了什么，一行一条；如果无法安全定点修复，也在这里说明)

=== PATCHES ===
--- PATCH 1 ---
TARGET_TEXT:
(必须从原文中精确复制、且能唯一命中的原句或原段)
REPLACEMENT_TEXT:
(替换后的局部文本)
--- END PATCH ---

=== UPDATED_STATE ===
(更新后的完整状态卡)

```

## Constant 91 | bytes 4221151..4221323
```text
=== FIXED_ISSUES ===
(逐条说明修正了什么，一行一条)

=== REVISED_CONTENT ===
(修正后的完整正文)

=== UPDATED_STATE ===
(更新后的完整状态卡)

```

## Constant 92 | bytes 4221324..4221523
```text

9. spot-fix 只能输出局部补丁，禁止输出整章改写；TARGET_TEXT 必须能在原文中唯一命中
10. 如果需要大面积改写，说明无法安全 spot-fix，并让 PATCHES 留空
```

## Constant nested | bytes 4222585..4222831
```text
润色：只改表达、节奏、段落呼吸，不改事实与剧情结论。禁止：增删段落、改变人名/地名/物品名、增加新情节或新对话、改变因果关系。只允许：替换用词、调整句序、修改标点节奏
```

## Constant nested | bytes 4222832..4223097
```text
改写：允许重组问题段落、调整画面和叙述力度，但优先保留原文的绝大部分句段。除非问题跨越整章，否则禁止整章推倒重写；只能围绕问题段落及其直接上下文改写，同时保留核心事实与人物动机
```

## Constant nested | bytes 4223184..4224144
```text
反检测改写：在保持剧情不变的前提下，降低AI生成可检测性。

改写手法（附正例）：
1. 打破句式规律：连续短句 → 长短交替，句式不可预测
2. 口语化替代：✗"然而事情并没有那么简单" → ✓"哪有那么便宜的事"
3. 减少"了"字密度：✗"他走了过去，拿了杯子" → ✓"他走过去，端起杯子"
4. 转折词降频：✗"虽然…但是…" → ✓ 用角色内心吐槽或直接动作切换
5. 情绪外化：✗"他感到愤怒" → ✓"他捏碎了茶杯，滚烫的茶水流过指缝"
6. 删掉叙述者结论：✗"这一刻他终于明白了力量" → ✓ 只写行动，让读者自己感受
7. 群像反应具体化：✗"全场震惊" → ✓"老陈的烟掉在裤子上，烫得他跳起来"
8. 段落长度差异化：不再等长段落，有的段只有一句话，有的段七八行
9. 消灭"不禁""仿佛""宛如"等AI标记词：换成具体感官描写
```

## Constant 178 | bytes 4225753..4225778
```text
build_auto_system_prompt
```

## Constant 181 | bytes 4225971..4225998
```text
build_legacy_system_prompt
```

## Constant 185 | bytes 4226023..4226047
```text
ReviserUserPromptParams
```

## Constant nested | bytes 4226090..4226114
```text
ReviserUserPromptParams
```

## Constant 190 | bytes 4226120..4226146
```text
build_reviser_user_prompt
```

## Constant 191 | bytes 4226147..4226187
```text
novels\prompts\inkos\reviser_prompts.py
```

## Constant 197 | bytes 4226281..4226327
```text
<module novels.prompts.inkos.reviser_prompts>
```
