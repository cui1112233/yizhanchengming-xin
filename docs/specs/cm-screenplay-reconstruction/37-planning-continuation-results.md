# 规划正文与自动续写：原编译对照及独立接入

2026-11-16。承接34–36，只说明内置规划生成这一项的本次实现。完整36组仍需继续。

## 原件及规则证据

原安装版`screenplay.creation_agent.generate_chapter_plan`声明`creation_agent.py:3166`，规划正文来自`works.storyboard_plan_guide.fill_plan_guide:498`；续写、拼接、上集摘要分别为`creation_agent.py:3102/3081/3072`。这些是编译声明元数据，不是恢复后的作者源码。原函数在隔离自造项目执行，模型工厂使用测试替身，账户资料/网络/子进程限制仍在。安装版和用户ZIP版本不同，以下结论不跨版本混用。

主案`07-parity-execution/planning-native-20261116/`：

- `entry-and-invalid-input.json`：四档、工具式与聊天式8份完整体例；`planning-generation-boundary.json`：首轮system/user实际组包及默认8轮请求。
- `planning-generation-extended.json`：中途空、错、第二轮完成、负值/零/1/10/100轮边界。最大8；负值收为1；零按默认8。连续两次空续写停止，本次已收到部分稿保留。中途错误也保留本次部分稿；首轮空或错保留之前保存的规划。
- `planning-generation-plain-append.json`：普通文本两段拼接实为201字，不能直接字符串拼接成200字。
- `planning-continuation-tail.json`：续写末尾使用最后16行；60种拼接样例含重叠、Markdown粘连与偏移边界。不能声称彻底消除所有重复。
- `planning-previous-excerpts.json`：上集末段取2200字符，只读紧邻上集。初次实验只改chapters.json，没有改变原helper实际读取的`chapter_0_storyboard.json`，不作有效长短对照；修正自造文件后重新验证，并捕获带上集摘录的实际生成请求。

原版后续请求保持4条消息：原system、原user、assistant累计全文、user本次续写要求。不是不断追加全部历史，也不能只发“请继续”。中途空返回时不把空assistant写入历史。原模型客户端报错重试与规划轮次是两层，不能将4次客户端调用说成4轮规划。

## CM实现

后端prompt模块嵌入恢复的8份全文，生成入口使用对应聊天式体例、所选分镜词全文、上集末段和本集正文。四档全文、首轮消息、16种续写消息及有上集的首轮请求逐字对照通过。前端没有硬编码模型提示词。

`storyboard_plan_rounds.go`接入完成器、默认最多8轮、完整即停、连续空停止和错误保留部分稿。`planning_rounds`及续写中断原因随正文保存；部分稿仍显示partial。原请求标识重读不会再次调用；上集末段、模型、所选分镜词、正文、体例和轮次上限都参与输入核对。取消返回旧正文和旧完整性信息；不会自动续跑已取消请求。

`TextRequest`与自身文本通道契约加入消息列表；普通与自定义OpenAI兼容通道都保留4条消息，参数回退也不丢历史。模型选择继续解析自己的配置入口，不使用原软件账号。已移除不再使用的通用规划提示词，独立的资产补充/助手草稿规则继续保留，不能误删其它入口。

网页显示生成轮次、中断原因和部分稿提醒。规划接口不再套用共享10分钟等待上限，由通道自身超时并保留用户明确取消信号，避免8轮尚未完成时被网页计时器切断。

## 已验证与边界

- 15组原编译生成场景：最终正文、轮次、字数、完整性和每轮消息一致；客户端错误重试在自身通道层测试，未冒称供应商错误重试次数完全相同。
- 8份体例、16份续写消息、60种拼接和既有220组完成器；全部来自原实验期望，而非由CM生成期望。
- screenplay/prompt全包及batchfactoryv11全包通过，相关Screenplay HTTP测试通过；普通通道参数回退和自定义通道最终HTTP请求保留历史。
- 第二轮取消保留旧完整稿，取消后同请求不自动重跑；改轮次或改上集规划不能复用原请求；空值保留旧稿。浏览器核对部分稿读回后的实际轮次数、中断原因及再次点击不重复规划。多轮取消信号和等待配置测试通过。

所有模型调用使用受控替身/最终HTTP传输，无新增真实额度消耗。自己的通道目前保持原有非流式HTTP方式；本次对齐的是文本、消息、业务轮次、结果和保存，不代表原版流式事件、逐轮实时推送或桌面像素也已对齐。已有规划的显式继续入口、真实项目更复杂的资产上下文及138分镜续写仍需单独核查。本轮实际4001交付结果随后附记，不能从测试通过推断已经更新。

## 实际4001交付及联合检查

源码提交`6a2e7ca2d96a62e7ac7b2883c53966b74b1c6304`，从提交快照独立构建。替换前核对批量工厂、小说获取和其它任务均空闲；保留13b20876运行包回滚。实际4001 PID31028，运行SHA及manifest均6a2e7ca2，入口`/assets/user-Tp3H7XdY.js`及`ComicScreenplayPage-pktjHMLo.js`与构建字节一致。旧回放作品原始回包逐字一致。

`07-parity-execution/planning-continuation-runtime-20261116/`保存`runtime-verification.json`与`mysql-http-result.json`。联合测试使用真实MySQL、AuthoringService、screenplayTextGenerator、自定义模型通道及最终请求解析，只替换最后HTTP传输。9项检查通过：8轮完整历史、自己的模型入口绑定、重复请求不重发、实际4001签名GET读回相同正文/完整性/轮次/partial状态。独立测试作品清理完成，真实供应商调用为0。

继续发现：原`_emit_plan_progress:3121`对自造window调用`window.onScreenplayAIProgress`，包含type=plan、chapter_index、round、chars、complete、missing，总轮数参数8。证据`planning-native-20261116/planning-progress-events.json`，对应恢复JS`index-DszfAkKD.js:440145/440515`等订阅点。这一实时显示尚未接入CM，下一步需对照真实字段及取消/切页时的迟到通知，不能以最终轮次已显示代替实时进度已对齐。
