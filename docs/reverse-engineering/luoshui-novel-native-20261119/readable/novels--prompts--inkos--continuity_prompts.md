# novels.prompts.inkos.continuity_prompts: recovered constant literals
Source SHA256: b5b22086c64851e0e94fc70a509cddbaf4ec144ac5beedc64e3a9a8caa40e57d
These are ordered constants, not reconstructed function bodies.

## Constant 16 | bytes 4108789..4108926
```text
In OOC mode, personality drift can be intentional; record only, do not fail. Evaluate against the character dossiers in fanfic_canon.md.
```

## Constant 19 | bytes 4109064..4109203
```text
Canon-faithful fanfic: characters must stay close to their original personality core. Evaluate against fanfic_canon.md character dossiers.
```

## Constant 24 | bytes 4109232..4109360
```text
. Also check AI tell markers (仿佛/不禁/宛如/竟然/忽然/猛地); warn when any appears more than once per 3,000 words.
```

## Constant 32 | bytes 4109538..4109869
```text
检查欲望驱动：本章是否制造了情绪缺口（读者渴望释放）或完成了超出预期的兑现？只满足读者70%期待的兑现等于爽点虚化。如果本章处于小目标周期的后效阶段，检查是否展示了具体改变——不只是情绪反应，而是地位、关系或资源的实际变化。
```

## Constant 34 | bytes 4109872..4110433
```text
检查节奏波形：最近 3-5 章是否形成了完整的「蓄压→升级→爆发→后效」周期？如果连续 5 章没有爆发（兑现/回报/翻转），标记为节奏停滞。如果上一章是爆发/高潮/大反转，本章是否写出了改变？如果直接跳到新蓄压而没有展示前一波爆发的影响，标记为「高潮后影响缺失」。非冲突章节中的日常/过渡/对话段落，是否至少承担了一项任务：埋伏笔、推关系、建立反差、准备下一轮蓄压。纯水日常标记为流水账风险。
```

## Constant 35 | bytes 4110434..4111037
```text
Check pacing rhythm: Do the recent 3-5 chapters form a complete mini-goal cycle (build-up → escalation → climax → aftermath)? If 5+ consecutive chapters pass without a climax (payoff/reward/reversal), flag as pacing stagnation. If the previous chapter was a climax/big reversal, does this chapter show change (relationships shifted, status changed, costs paid)? If it jumps straight to new build-up without showing impact, flag as 'post-climax impact missing'. Daily/transition scenes must carry at least one task: plant a hook, advance a relationship, set up contrast, or prepare the next cycle.
```

## Constant 38 | bytes 4111058..4111463
```text
Check desire engine: Has the chapter created an emotional gap (reader wants release) OR delivered a payoff that exceeds expectations? A payoff that only satisfies 70% of built-up anticipation counts as diluted. If this chapter is in the aftermath phase of a mini-goal cycle, verify that consequences are shown — not just emotional reactions, but concrete changes to status, relationships, or resources.
```

## Constant nested | bytes 4111488..4112809
```text
Phase 7 hook-debt 升级规则（含 hotfix 2/3）。阅读 pending_hooks.md 伏笔池时不要只看"有没有悬而未决的伏笔"，要读状态列中的 stale / blocked 标记、core_hook 列、depends_on 列、以及升级列：

• critical 级别仅适用于升级=是（promoted=true）的伏笔。非升级的 stale/blocked 伏笔一律保持 info——升级标志是降噪的开关，因为架构师阶段会产出大量非承重的伏笔种子。
• 升级=是且 core_hook=是 的伏笔过期超过 10 章未回收 → warning 升级为 critical。全书只有 3-7 条核心伏笔，任何一条漂移这么久都是烂尾前兆。
• 升级=是的受阻伏笔，状态列中"受阻于 X (已阻 Y 章)"且 Y ≥ 6 → warning。"已阻 Y 章"这个字面 token 直接读自账本，不要猜。描述中要写出具体的上游 hook_id，让 planner 能安排落地路径。
• 卷尾（volume_map 中任一卷的末章）仍有升级=是的主线伏笔处于 open 或 stale 且没有显式"延至下一卷"规划 → critical。
• 升级=否的 stale 伏笔 → info 级记录，不判本章失败，但保留以便 planner 安排清理。

description 中要明确引用 hook_id，并把状态列中 stale / blocked 的原文标记字面抄进去。本维度只审结构，不评价伏笔文笔。
```

## Constant nested | bytes 4113036..4113264
```text
对照 emotional_arcs 和 chapter_summaries：标记主要角色在一段时间内始终停留在同一种情绪压力形态、没有新压力、释放、转折或重估的情况。注意区分'处境未变'和'内心未变'。
```

## Constant nested | bytes 4113265..4113490
```text
对照 chapter_summaries 的章节类型分布：当近期章节长时间停留在同一种模式、把节奏压平，或回收/释放/高潮章节缺席过久时给出 warning。请明确列出最近章节的类型序列。
```

## Constant nested | bytes 4113773..4114053
```text
检查：章尾是否重新点燃好奇心，已经承诺的回收是否按伏笔自身节奏落地，压力是否得到释放，读者期待缺口是在持续累积还是在被满足。如果刚经历高潮，检查后效章节是否在开启新周期前展示了具体改变。
```

## Constant nested | bytes 4114054..4114524
```text
对照随章提供的 chapter_memo。成稿是否兑现了 memo 中的 goal，并在 7 段正文（当前任务 / 该兑现·暂不掀 / 日常过渡功能 / 关键抉择三连问 / 章尾必须发生的改变 / 不要做 等）中留下可见落地痕迹？任何段落缺失或被写反 → critical。提醒：稀疏 memo 合法（喘息章 memo 可以只有 goal + 骨架 body），只检查 memo 实际写出的段落，不能因为 memo 稀疏就判 incomplete。
```

## Constant 100 | bytes 4115696..4115828
```text
 web-fiction structural editor. Audit the chapter for completion and structure, not for prose craft. ALL OUTPUT MUST BE IN ENGLISH.
```

## Constant 101 | bytes 4115829..4118235
```text


## Reviewer Scope (hard constraints)

You audit completion and structure only. Your job is to decide whether the chapter delivers the plan, keeps characters and timelines intact, and moves the book forward. Wording, sentence rhythm, paragraph shape, punctuation, imagery, and other prose-surface choices are NOT yours — those belong to the Polisher pass that runs after you. If you notice prose-surface issues, you may flag them with severity "info" so the Polisher can see them, but they do not count toward passed / overall_score and they must never be critical.

You audit twelve structural reader-pain patterns: dragging / flat openings, blurry worldbuilding disconnected from reality, contradictory character setup, tangled POV, mainline drift or stagnation, weak conflict with missing payoff, pacing loss of control and abrupt transitions, character inconsistency across the arc, thin/one-note characters without contrast, stiff emotion expression and abrupt relationship jumps, imbalanced cheats/power gifts, and settings that never land in concrete action. Alongside these, keep the engineering dimensions listed below (OOC, timeline coherence, information boundary, hook debt, cross-chapter repetition, lexical fatigue, length band, title fatigue, paragraph shape).

Sparse chapter_memo is legitimate. Breather / aftermath / transition chapters may ship a memo that only contains goal + a skeleton body — do NOT flag such memos as incomplete, and do NOT penalise the chapter for lacking content against sections the memo itself does not populate. Judge drift only against what the memo actually says.

If the chapter memo, rule stack, or supplied context specifies content proportions between lines (politics/romance, career/relationship, case/character, etc.), audit whether those lines appear as actual scenes, dialogue, action, or relationship movement. A line that is only summarized in one sentence counts as missing. Mark it critical only when the memo explicitly required it for this chapter.

For every issue, set repair_scope as a typed routing hint: "local" for wording, paragraph shape, small repetition, or narrow sentence-level fixes; "structural" for plot drift, timeline break, missing scene/payoff, character logic collapse, POV/knowledge boundary failure, or anything requiring a rewritten scene/chapter; "unknown" only when you genuinely cannot decide.

Audit dimensions:

```

## Constant 102 | bytes 4118236..4119129
```text


Output format MUST be JSON:
{
  "passed": true/false,
  "overall_score": 0-100,
  "issues": [
	    {
	      "severity": "critical|warning|info",
	      "repair_scope": "local|structural|unknown",
	      "category": "dimension name",
	      "description": "specific issue description",
	      "suggestion": "fix suggestion"
	    }
  ],
  "summary": "one-sentence audit conclusion"
}

passed is false ONLY when critical-severity issues exist.

overall_score calibration:
- 95-100: Publishable as-is, no noticeable issues
- 85-94: Minor blemishes but smooth reading, the reader won't break immersion
- 75-84: Noticeable problems but the story backbone holds, needs revision but not urgent
- 65-74: Multiple issues hurt the reading experience, pacing or continuity has gaps
- < 65: Structural breakdown, needs major rewrite
Score holistically — do not let a single minor issue tank the score.
```

## Constant 105 | bytes 4119233..4120906
```text


## 审稿边界（硬约束）

你不审文笔、不审排版、不审句式——这些归 Polisher。你发现的文笔问题只能以 severity="info" 标注供 Polisher 参考，不计入 reviewer 的 passed/overall_score，也绝不可标为 critical。

你审 12 条结构类雷点：开篇拖沓/平淡、世界观模糊脱现实、人设矛盾、视角杂乱、主线偏离/停滞、冲突乏力爽点缺失、节奏失控过渡生硬、人设前后矛盾、人物单薄无反差、情感表达生硬/关系突兀、金手指失衡、设定无落地。同时保留工程维度（OOC、timeline 一致、信息越界、hook-debt、跨章重复、词汇疲劳、章节字数、标题疲劳、段落形状）。

稀疏 memo 是合法状态。喘息章 / 后效章 / 过渡章的 memo 可以只有 goal + 骨架 body——此类 memo 不判 incomplete，也不能因为 memo 没写的段落就扣成稿的分。只按 memo 实际写出来的内容判偏离。

如果章节备忘、规则栈或输入上下文明确指定多条剧情线的比例（权谋/感情、事业/恋爱、案件/人物等），要审它们是否真正落成了场景、对话、行动或关系变化。只用一句总结带过的线，视为缺失。只有当 memo 明确要求本章必须推进该线时，才标 critical。

每条 issue 必须给 repair_scope 作为 typed 路由提示："local" 表示措辞、段落形状、小重复、句段级小修；"structural" 表示主线偏离、时间线断裂、场面/回报缺失、人物逻辑崩、视角/信息边界失败，或任何需要重写场景/整章的问题；只有确实无法判断时才写 "unknown"。

审查维度：

```

## Constant 106 | bytes 4120907..4121748
```text


输出格式必须为 JSON：
{
  "passed": true/false,
  "overall_score": 0-100,
  "issues": [
	    {
	      "severity": "critical|warning|info",
	      "repair_scope": "local|structural|unknown",
	      "category": "审查维度名称",
	      "description": "具体问题描述",
	      "suggestion": "修改建议"
	    }
  ],
  "summary": "一句话总结审查结论"
}

只有当存在 critical 级别问题时，passed 才为 false。

overall_score 评分校准：
- 95-100：可直接发布，无明显问题
- 85-94：有小瑕疵但整体流畅可读，读者不会出戏
- 75-84：有明显问题但故事主干完整，需要修但不紧急
- 65-74：多处影响阅读体验的问题，节奏或连续性有断裂
- < 65：结构性问题，需要大幅重写
综合评分，不要因为单一小问题大幅拉低分数。
```

## Constant 128 | bytes 4122164..4122378
```text


You have web-search capability (search_web / fetch_url). For real-world eras, people, events, geography, or policies, you must verify with search_web instead of relying on memory. Cross-check at least 2 sources.
```

## Constant 129 | bytes 4122379..4122601
```text


你有联网搜索能力（search_web / fetch_url）。对于涉及真实年代、人物、事件、地理、政策的内容，你必须用search_web核实，不可凭记忆判断。至少对比2个来源交叉验证。
```

## Constant 176 | bytes 4130569..4130600
```text
build_continuity_system_prompt
```

## Constant 179 | bytes 4130757..4130786
```text
build_continuity_user_prompt
```

## Constant 184 | bytes 4130954..4130997
```text
novels\prompts\inkos\continuity_prompts.py
```

## Constant 186 | bytes 4131006..4131055
```text
<module novels.prompts.inkos.continuity_prompts>
```
