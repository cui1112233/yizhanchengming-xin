# novels.prompts.imitation_short: recovered constant literals
Source SHA256: b5b22086c64851e0e94fc70a509cddbaf4ec144ac5beedc64e3a9a8caa40e57d
These are ordered constants, not reconstructed function bodies.

## Constant 4 | bytes 4016745..4016981
```text
你是一位资深知乎盐选短篇小说拆书师。

你的任务是对参考原文进行深度拆解，产出一份六维《拆书报告》。
拆的不是剧情，而是「为什么读者会上头」。

## 六维拆书方法论


```

## Constant 6 | bytes 4016983..4018078
```text


## 输出要求

严格按照以下 Markdown 模板输出完整拆书报告（不要跳过任何维度）：

```
# 《原作标题》拆书报告

## 一、核心情绪曲线
- 一句话概括全篇情绪走向：____
- 逐拍情绪图谱：
  | 阶段 | 关键事件 | 主导情绪 | 强度(1-10) | 情绪引擎 |
  |------|---------|---------|-----------|---------|
  | ... | ... | ... | ... | ... |
- 情绪最低点 / 最高点 / 最关键反转点
- 「意难平 / 爽 / 暖」的核心来源

## 二、人物关系与人设
- 主角：欲望 / 致命缺陷 / 人物弧光 / 人设标签
- 关键对手或关系核心
- 重要配角及功能
- 中心关系张力

## 三、核心价值观（仿写之魂）
- 一句话价值观
- 表层主题 vs 深层价值观
- 戳中读者的「相信」

## 四、大众底层情绪渴望
- 主渴望 + 次渴望
- 读者画像与代入位

## 五、底层核心逻辑
- 结构骨架（用→连接）
- 核心矛盾引擎
- 钩子与悬念链

## 六、金句摘录（仅作手法参考）
- 逐条摘录 5-10 句，标注手法与情绪功能
```
```

## Constant 11 | bytes 4018348..4018938
```text


## 要求

1. 三个方向在题材/时代/设定上拉开差距（如：现代现实、古代架空、脑洞重生）
2. 共享同一核心价值观
3. 每个方向给出：
   - 题材类型
   - 一句话设定（高概念钩子）
   - 故事概要（150-250字，可看出起承转合与情绪爆点）
   - 如何沿用核心价值观、如何放大情绪曲线
   - 预期主爽点 / 主虐点

## 输出格式

=== DIRECTION 1 ===
题材：____
设定：____
概要：____
价值观沿用：____
主爽点：____
主虐点：____

=== DIRECTION 2 ===
...

=== DIRECTION 3 ===
...
```

## Constant 23 | bytes 4019371..4020131
```text


## 蓝图要求

输出以下内容：

### 1. 放大后的情绪曲线
在原作每个情绪节点上标注「如何比原作更狠/更爽/更上头」

### 2. 新人物关系
相对原作做了哪些微调

### 3. 新主线结构（事件链）——情绪功能→新事件映射表
| 情绪功能（来自拆书）| 原作用什么事件实现 | 新作改用什么事件实现 |
|---|---|---|

### 4. 新金句清单
5-8 条全新金句，标注落点位置

### 5. 角色与命名
按去 AI 味规范起名，避开禁用清单

### 6. 分段规划
按 5-8 个场景/小节切分，每节预估字数，合计落在目标区间

### 7. 防抄袭对照
列出原作的关键元素（人名/地名/年代/事件/台词），以及新作中的替代方案
```

## Constant 33 | bytes 4020317..4020796
```text
1. 核心价值观必须与原作相同
2. 情绪曲线要比原作更饱满
3. 绝不抄袭——不允许与原文连续相同的字词句
4. 主线结构要不同——保留情绪功能链，换掉事件
5. 人物关系有变化
6. 金句全部重写
7. 全文禁用破折号（——）
8. 禁用AI味名字（林墨/沈墨/陈默/苏婉等）
9. 知乎风：第一人称唠嗑感、大白话短句、show not tell
10. 禁对仗腔、禁排比三连、禁段尾升华点题
```

## Constant 42 | bytes 4021145..4021753
```text


## 审核要求

1. 逐维度审稿，产出《编辑审核意见单》
2. 总评级：可过审 / 需修改 / 打回
3. 重点维度：开篇留存、追读性、爽虐兑现、情绪冲击、逻辑自洽、价值观导向、内容合规、AI 味检查（含破折号/禁用名）、原创查重、完整度字数
4. 附 2-3 个知乎风标题建议 + 一句话简介
5. 具体定位到段落/章节给出修改建议

## 输出格式

=== REVIEW START ===
## 总评级
____

## 逐维度审核
...

## 标题建议
1. ____
2. ____
3. ____

## 一句话简介
____

## 修改清单
...
=== REVIEW END ===
```

## Constant 55 | bytes 4021920..4021994
```text
Load all reference constants combined, for injection into system prompts.
```

## Constant 56 | bytes 4021995..4022251
```text
Imitation short-fiction prompt builders.

Compiles the zhihu-fiction-rewrite skill's methodology into structured
prompts for the imitation short fiction pipeline.  All reference content
lives in references.py as compiled Python constants (not .md files).

```

## Constant 65 | bytes 4022346..4022370
```text
prompts\imitation_short
```

## Constant nested | bytes 4022373..4022403
```text
NUITKA_PACKAGE_novels_prompts
```

## Constant nested | bytes 4022438..4022484
```text
NUITKA_PACKAGE_novels_prompts_imitation_short
```

## Constant nested | bytes 4022676..4022689
```text
SKILL_PROMPT
```

## Constant 77 | bytes 4022690..4022703
```text
SKILL_PROMPT
```

## Constant 80 | bytes 4022761..4022796
```text
build_deconstruction_system_prompt
```

## Constant 82 | bytes 4022829..4022862
```text
build_deconstruction_user_prompt
```

## Constant 83 | bytes 4022863..4022893
```text
build_direction_system_prompt
```

## Constant 86 | bytes 4023057..4023085
```text
build_direction_user_prompt
```

## Constant 87 | bytes 4023086..4023116
```text
build_blueprint_system_prompt
```

## Constant 90 | bytes 4023304..4023332
```text
build_blueprint_user_prompt
```

## Constant 93 | bytes 4023442..4023479
```text
build_editorial_review_system_prompt
```

## Constant 95 | bytes 4023512..4023547
```text
build_editorial_review_user_prompt
```

## Constant 99 | bytes 4023626..4023669
```text
novels\prompts\imitation_short\__init__.py
```

## Constant 100 | bytes 4023670..4023710
```text
<module novels.prompts.imitation_short>
```
