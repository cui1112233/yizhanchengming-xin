# 原生规划正文、续写与完成判断：恢复规格

2026-11-16。承接现有规划缺口。本项恢复自当前安装版，不能据此宣称用户ZIP中不同原件也具有相同行为。没有拿到完整作者源码；以下行号是原编译函数声明元数据，前端行号则属于恢复后的JS。

证据主目录：`07-parity-execution/planning-native-20261116/`。脚本在隔离目录使用自造章节；限制原账户资料读取，网络和外部进程均拒绝。模块初始化尝试的握手被拒绝，没有放宽鉴权或读取原账号。正常生成实验仅将模型客户端工厂和自造提示词读取换成测试替身，其余使用原章节写入、规划生成、校验及章节读回方法。不是供应商生成结果。

## 入口与输入

- 恢复JS442503–442523：`su`调用`screenplay_generate_chapter_plan`，传project_path、chapter_index、prompt_id、prompt_name、model、planDensity；返回complete=false时提醒打开“查看规划”核对。
- 原Api包装入口：`api.py:37612`，委托`screenplay.creation_agent.generate_chapter_plan`，声明`creation_agent.py:3166`。
- 模型客户端工厂来自`agent.llm_client.create_openai_client`，声明102。CM继续用自己的模型配置，不复制原服务地址或账号。
- 缺项目路径/章节返回“缺少项目路径或章节”；缺配置在生成前报错。自造章节由原Api.save_screenplay_chapters保存到隔离项目的chapters.json，再由原规划链读取。

## 提示词和实际请求

`entry-and-invalid-input.json`保存四档的`get_storyboard_plan_guide`和`fill_plan_guide`工具式/聊天式全文，各档不是仅替换名称。规划六章为总览、资产绑定核对、分节明细、生成前必须处理的风险点、待确认项、跨集承接与批处理。普通档以不超过15秒的小节为基线；SD2.5档明确25.0–30.0秒、推荐26–30秒。

`planning-generation-boundary.json`捕获原生成函数实际构造的请求：流式请求；首轮system为独立短指令，user由聊天式规划全文、所选分镜词全文、上集规划说明及本集剧本组成。不能继续拿CM目前独立通用词作为原版恢复正文，也不能把整个规划正文随意塞进system后声称相同。

后续轮次保留对话历史，并加原`_plan_continue_user_message`（声明3102）生成的续写要求：当前字数、缺少章节、已输出末尾，要求从断开处继续，不重开标题、不重写全文。累计拼接不能自行宣称百分百去重；原helper短文本重复实验也会保留重复内容。

## 实际轮次和输出

六组正常结构的自造实验：

| 输入/替身返回 | 实际调用次数 | 原返回 |
|---|---:|---|
| 默认，始终只给第一章 | 8 | success=true，has_plan=true，complete=false，rounds=8 |
| auto_continue=false，同上 | 1 | 保留35字，complete=false |
| max_rounds=2，同上 | 2 | 保留70字，complete=false |
| 符合原校验的2500字六章规划 | 1 | complete=true，包含完整原因 |
| 首轮空返回 | 1 | success=false，“规划生成为空” |
| 模型客户端报错 | 3次控制调用 | success=false，错误可见；这三次来自原调用层重试，不能称三轮剧情续写 |

原章节读回已保存到每个实验记录。planning-generation-summary.json已逐字比较空值/报错后的规划，确认保留先前完整稿，10项读回/调用次数检查全部通过。取消、已有规划显式continue、跨集摘要、轮次上下界和复杂重复内容仍需补实验。

## 完成判断

原`works.episode_coverage.evaluate_plan`声明185；`planning-rule-cases.json`保存28种基础组合、四档长度边界、10种拼接、4种续写提示输入输出及模块常量。

- 六章是否齐全和标题是否属于对应类别。
- 第一章含剧本字数等预算信息；“写了六个标题”不代表完整。
- 有分节表及声明的总规划节数。
- 原常量：普通档最低2500字符、SD2.5最低1800字符，最后一章最低80字符，行数比例0.9。常量记录与控制实验分开，不能仅凭常量推断所有分支。
- 已实测相同规范文本：普通档2499拒绝、2500通过；SD2.5档1799拒绝、1800通过。这只是完成器的格式阈值，不是剧本或规划质量保证。
- 原返回ok、chars、headings、missing、wrong、rows、declared、reason；UI“可能仍不完整”不能变成完整成功。

## CM需修正的对应规则

当前`backend/internal/screenplay/storyboard_planning.go`只请求一轮，非空文字即PlanningStatus=succeeded；`backend/internal/prompt/screenplay_planning_reference.go`明确仍是独立通用指令。与上面的原编译实验不一致，本项不能标为已复刻。

实施时系统预设正文放后端prompt模块；保存所选分镜词和自己的模型映射。首轮与续写请求按原组包、正文和完成器对照，逐轮保留规划和来源版本；已有文字、请求成功、规划完整分别表示。到达轮次上限仍保留部分稿并说明缺什么，不将部分稿误作完整。失败/空值/取消不得覆盖旧稿或重复收费；已授权测试一律控制生成器，不再消耗真实额度。

验收必须补齐正常完整/多轮部分/空值/重复/输入变更/错误/取消，以及原版和CM的同输入最终请求、保存后读回和页面状态。138分镜正文的20%–30%续写仍是另一链，不能用本规划续写通过来代替它。

后续检查：[35 完整性规则220组对照](35-planning-coverage-results.md)。其中修复的只是完成判断与显示，8轮续写仍未接入。

2026-11-16后续：[37 规划正文及默认8轮独立接入](37-planning-continuation-results.md)。本文的“当前未接入”段落属于恢复规格时的历史检查点，以37及实际运行证据为准。
