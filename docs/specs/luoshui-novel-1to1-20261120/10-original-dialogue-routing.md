# 10 原创对话、确认和长短篇分流：本轮原生补证

## 已完成和未完成分开

本轮从API层一直追到`AgentSessionService.send_message/confirm_action`的真实函数体。原版前端与原生的名字、入参和可见分流已取得证据；后面的`PipelineRunner.create_book/run_short_fiction`全部内部阶段尚未读完。因此G02/G03推进为“确认入口及分流已核验，完整生产与恢复仍待补”，不标全流程完成。

原生源文件身份：`novels.pyd` SHA256 `b5b22086…`。原始IDA导出保留在`source/ida/`，分析员按常量槽加的注释单独放`source/ida-annotated/`；注释内的“Original IDA line”用于回到原导出行。不是恢复的可运行Python工程。

## 1. 入口和函数绑定

| 层 | 原名 | 原生入口 | 证据 |
|---|---|---|---|
| 前端发送 | novels_send_message | 0x180082cb0 | `source/bridge-contracts/novels_send_message.inspection.js:1`；API构造绑定 |
| 前端确认 | novels_confirm_action | 0x180084690 | 同名inspection；API构造绑定 |
| 服务发送 | AgentSessionService.send_message | 0x1801f5cb0 | 会话模块初始化函数`0x180217350`的构造指针与名称槽 |
| 服务确认 | AgentSessionService.confirm_action | 0x1801fd690 | 同上；`source/native-entrypoint-bindings.json` |
| 检查点确认 | AgentSessionService.confirm_pipeline_checkpoint | 0x1802095a0 | 目前已绑定入口，未在本轮读取完整函数体 |

## 2. 对话请求到底传什么

前端`NovelStudioPage.readable.js:442–492`先确保会话存在，再发：sessionId、message、model、knowledgeBookIds、thinkingMode=true、codexReasoningEffort、claudeReasoningEffort，可选contextLength。失败/取消/成功分别处理，不是点发送就自动建作品。

API原生体`source/ida/0x180082cb0.c:220–345`读取sessionId、message、model并做前置判断；`:351–449`加载会话、清取消标记并确定作品上下文；`:497–584`读取contextLength及camelCase/snake_case的渠道思考参数；`:595–921`取服务send_message并构造关键字参数，成功包成success/data；`:1043–1047`可见cancelled字段的异常返回分支。精确kwargs、异常文字保留在原常量和注释，不把前端字段名直接当服务参数名。

API常量#216的关键字列表为model、knowledge_book_ids、thinking_mode、context_length、codex_reasoning_effort、claude_reasoning_effort，见`constants/novels--api--novels_api.json`原件字节[3928321,3928430)。这说明**前端思考设置有进入服务调用**，但不能据此保证所有上游模型都会接受同样参数。

## 3. 对话模型先给建议或动作提案

`source/ida/0x1801f5cb0.c`可见：

| 原导出行号 | 行为和证据范围 |
|---|---|
| 1072附近 | `_resolve_chat_model`，随后读取session.kind。模型解析规则的完整函数体本轮未导出。 |
| 1970–2007 | 调`build_system_prompt`并使用kind；不是各模式共用一句“你是小说作家”。 |
| 2125附近 | knowledge_bridge相关上下文载入。具体选中资料注入顺序仍需继续追。 |
| 2679–3007 | user/assistant消息处理、用户消息持久化。精确裁剪和去重分支尚未逐个还原。 |
| 3183–3358 | 根据`_should_expose_propose_tool`选择提案工具，调用`chat_with_tools`，包含取消回调。 |
| 3953–4075 | 构造action/toolName=propose_action，将动作写session.pendingAction，并保存会话。 |
| 4413–4530、5090附近 | propose_action规范化和validate_proposed_action。并非模型吐出任何文字都可直接执行。 |

上表证明这些可见路径存在及先后位置，不声称所有请求必然走全部路径；无工具、无提案、流式输出和失败分支仍需逐分支读。文档不能将send_message等同于完整生产器。

## 4. 确认卡提交的是哪份动作

前端`NovelStudioPage.readable.js:669–706`的确认载荷仅包含当前sessionId、model及渠道思考设置，**没有把整份action重新提交**。真正动作来自服务已保存的pendingAction。

`source/ida/0x1801fd690.c:1629–1663`读取pendingAction；为空时用常量#117返回失败。读取action字段后再做分派。必须保留这条绑定：会话→待确认动作→用户确认→执行；不能前端临时凑一个动作就直接生成，也不能“确认”和“蓝图后继续”共用一份不区分类型的处理。

## 5. 长篇与短篇的确认分支不同

| 分支 | 输入/动作 | 可见输出及边界 | 原导出证据 |
|---|---|---|---|
| create_book | 后台pendingAction，选定model | 取pipeline.create_book；成功后改session.kind=book、挂书籍并保存；给前端switchToBook等结果。**该分支不是全书正文已写完**。 | `0x1801fd690.c:2449–2501`调用/改kind；`:3017`附近保存；返回字段见后续分支与前端 |
| short_run | pendingAction中的direction/instruction/summary/title，现有或生成的storyId、model | 构造短篇载荷，设置storyId，调用pipeline.run_short_fiction；短篇有自己的实体，不伪装成长篇第一章。 | `:3177–3668`；`:3640–3668`为storyId/调用 |
| short_run暂停 | 内层生产返回awaitingConfirmation | 保存short_fiction_run等待状态，返回checkpoint；前端显示检查点卡。 | `:3841–4393`；前端`:712–722` |
| write_next | 待确认写章动作 | 调write_next_chapter；章节返回后刷新作品。 | `:6392–6424`及后续分支 |
| audit_chapter | 待确认审稿动作 | 调audit_chapter，按审稿对象返回，不能只显示“已执行”。 | `:7383–7437,7613–8185` |
| revise_chapter | 待确认修订动作 | 独立修订分支及结果保存，不能当写下一章。 | `:8367–8433,8942–9383` |
| assistedAction | 引导至素材入口的动作 | 返回assistedAction/targetRoute，前端打开仿写/同人/番外/续写表单；不是马上生成。 | `:2242–2285`；前端`:706–710` |

分支列表还不是这个10779行函数所有路径的穷举；封面、批量、互动世界和异常收尾需继续逐项核验。不能因为已知几类动作就标“全部确认逻辑相同”。

## 6. 前端如何消费真实返回

来源：`NovelStudioPage.readable.js:699–728`。

- success=false：进度卡error、显示message、重新加载会话；不是通用“执行完成”。
- assistedAction：按targetRoute打开对应导入表单并提示补材料。
- switchToBook+session：切换到真实书籍会话，清待确认卡，刷新作品/会话。
- awaitingConfirmation：显示新检查点，暂停在可编辑文件状态，不把它当生成成功。
- shortFiction：显示短篇结果；audit按passed显示通过/未通过；chapter区分新章和修订。
- 只有上述业务结果均未出现才走普通执行完成提示。不能用一个success toast替所有返回处理。

## 7. 当前CM已确认的差别（只限本次读到的入口）

| 项目 | CM当前源码 | 与原版的差别 |
|---|---|---|
| 对话返回协议 | `backend/internal/novelcreation/studio_service.go:168–205`只接受action.type=create_book，其他类型标invalid_output | 原版对话服务可提案多种动作；此CM入口不能接short_run/write_next等同等动作，差异已确认。不能由此推定CM所有其他短篇入口都不存在。 |
| 确认生产 | `studio_service.go:491–580`核对ActionID与已保存r.Action，防重复；随后建project和long_original会话，再发生产指令 | 待确认绑定有实现，但此分支固定长篇；与原版同一确认服务按create_book/short_run等分派不同。 |
| 防重复 | 同文件已confirmed返回既有transition，事务中再次核对；生产使用studio-confirm:actionID | 这是CM已有保护；补多动作分流时必须保留，不能为复刻把它删掉。 |

CM摘要与文件哈希见`source/cm-studio-routing-snapshot.json`。这是源代码对照，不是4001当前运行状态验收。

## 8. 尚需补证，不能提前写成已修复

1. PipelineRunner.create_book、run_short_fiction完整函数入口、调用与各阶段输出。
2. 原版全部提案字段验证和默认值；失败、取消、重复确认的每条出口。
3. 建书、短篇、审稿、修订、批量、封面的完整类型与UI反馈。
4. 将确认的新协议差异转为CM实施规格，再连同原有防重复/权限/持久化一起实现和验证。本轮没有改产品代码，也没用收费模型测试。
