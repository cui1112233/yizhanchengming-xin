# 09 当前版接口调用合同补证

本轮基于当前bundle `7c252056…`，独立重写白名单字面量解释器，不执行原bundle/helper、不导入目标模块。解码表zan共565项，Qan下标减267，旋转278次后原校验式=299251。原式、UTF16/UTF8范围和调用者均保留。

## 一、名称存在与实际调用分开

扫描得到57个novels_属性访问，覆盖44个名称。其中8个支路的调用参数含未解析`_0x…`占位，另1个非异步脚手架分支暂不算有效调用。其余48个异步候选包含文件选择的guard和调用重复，不能称为48个独立功能。全部46个公开API均通过IDA构造参数绑定到真实原生地址；不代表46个函数体都已读完。

当前Qan相关未解析属性为0，不等于整个17MB前端所有动态属性都已解完；其他字符串表和动态对象关系不在该结论范围。

| 原接口 | 前端组件/共享函数 | 参数原式 | IDA入口 | 完整调用片段 | 状态 |
|---|---|---|---|---|---|
| `novels_list_books` | NovelStudioPage | `()` | `0x18006cde0` | [原式](source/bridge-contracts/novels_list_books.inspection.js) | 前端异步候选＋原生入口已绑定 |
| `novels_get_book` | 共享桥接函数 Oan | `(e)` | `0x18006d730` | [原式](source/bridge-contracts/novels_get_book.inspection.js) | 前端异步候选＋原生入口已绑定 |
| `novels_rename_book` | NovelStudioPage | `(e, t)` | `0x18006dee0` | [原式](source/bridge-contracts/novels_rename_book.inspection.js) | 前端异步候选＋原生入口已绑定 |
| `novels_delete_book` | NovelStudioPage | `(e)` | `0x18006e800` | [原式](source/bridge-contracts/novels_delete_book.inspection.js) | 前端异步候选＋原生入口已绑定 |
| `novels_list_sessions` | 共享桥接函数 Ian | `(e = null)` | `0x18006ed60` | [原式](source/bridge-contracts/novels_list_sessions.inspection.js) | 前端异步候选＋原生入口已绑定 |
| `novels_get_session` | 共享桥接函数 Nan | `(e)` | `0x18006f6f0` | [原式](source/bridge-contracts/novels_get_session.inspection.js) | 前端异步候选＋原生入口已绑定 |
| `novels_get_short_story` | 共享桥接函数 Man | `(e)` | `0x180071030` | [原式](source/bridge-contracts/novels_get_short_story.inspection.js) | 前端异步候选＋原生入口已绑定 |
| `novels_read_short_story_file` | NovelStudioShortPanel | `(e, t)` | `0x1800717e0` | [原式](source/bridge-contracts/novels_read_short_story_file.inspection.js) | 前端异步候选＋原生入口已绑定 |
| `novels_save_short_story_file` | NovelStudioShortPanel | `(e, t, n)` | `0x180071f90` | [原式](source/bridge-contracts/novels_save_short_story_file.inspection.js) | 前端异步候选＋原生入口已绑定 |
| `novels_confirm_pipeline_checkpoint` | NovelStudioPage | `(e)` | `0x1800727b0` | [原式](source/bridge-contracts/novels_confirm_pipeline_checkpoint.inspection.js) | 前端异步候选＋原生入口已绑定 |
| `novels_create_session` | NovelStudioPage | `(e = {})` | `0x180073da0` | [原式](source/bridge-contracts/novels_create_session.inspection.js) | 前端异步候选＋原生入口已绑定 |
| `novels_rename_session` | NovelStudioPage | `(e, t)` | `0x180074f40` | [原式](source/bridge-contracts/novels_rename_session.inspection.js) | 前端异步候选＋原生入口已绑定 |
| `novels_bind_session_book` | NovelStudioPage | `(e, t)` | `0x180075870` | [原式](source/bridge-contracts/novels_bind_session_book.inspection.js) | 前端异步候选＋原生入口已绑定 |
| `novels_delete_session` | NovelStudioPage | `(e)` | `0x1800761f0` | [原式](source/bridge-contracts/novels_delete_session.inspection.js) | 前端异步候选＋原生入口已绑定 |
| `novels_update_session_messages` | 共享桥接函数 qan | `(e, t)` | `0x180076750` | [原式](source/bridge-contracts/novels_update_session_messages.inspection.js) | 前端异步候选＋原生入口已绑定 |
| `novels_list_style_sources` | NovelStudioPage | `()` | `0x180077180` | [原式](source/bridge-contracts/novels_list_style_sources.inspection.js) | 前端异步候选＋原生入口已绑定 |
| `novels_get_style_profile` | NovelStylePanel | `(e)` | `0x180077890` | [原式](source/bridge-contracts/novels_get_style_profile.inspection.js) | 前端异步候选＋原生入口已绑定 |
| `novels_save_style_profile` | NovelStylePanel | `(e)` | `0x1800780b0` | [原式](source/bridge-contracts/novels_save_style_profile.inspection.js) | 前端异步候选＋原生入口已绑定 |
| `novels_delete_style_profile` | NovelStylePanel | `(e)` | `0x180079700` | [原式](source/bridge-contracts/novels_delete_style_profile.inspection.js) | 前端异步候选＋原生入口已绑定 |
| `novels_import_style_files` | NovelStylePanel | `(e)` | `0x180079ef0` | [原式](source/bridge-contracts/novels_import_style_files.inspection.js) | 前端异步候选＋原生入口已绑定 |
| `novels_pick_style_files` | NovelImitationShortDialog / NovelImportPage / NovelStylePanel | `()` | `0x18007b680` | [原式](source/bridge-contracts/novels_pick_style_files.inspection.js) | 前端异步候选＋原生入口已绑定 |
| `novels_init_imitation` | NovelStudioPage | `(e)` | `0x18007be90` | [原式](source/bridge-contracts/novels_init_imitation.inspection.js) | 前端异步候选＋原生入口已绑定 |
| `novels_run_imitation_short` | NovelStudioPage | `(e)` | `0x18007c840` | [原式](source/bridge-contracts/novels_run_imitation_short.inspection.js) | 前端异步候选＋原生入口已绑定 |
| `novels_init_fanfic` | NovelImportPage | `(e)` | `0x18007f270` | [原式](source/bridge-contracts/novels_init_fanfic.inspection.js) | 前端异步候选＋原生入口已绑定 |
| `novels_init_spinoff` | NovelImportPage | `(e)` | `0x18007fc20` | [原式](source/bridge-contracts/novels_init_spinoff.inspection.js) | 前端异步候选＋原生入口已绑定 |
| `novels_import_chapters` | NovelImportPage | `(e)` | `0x1800805d0` | [原式](source/bridge-contracts/novels_import_chapters.inspection.js) | 前端异步候选＋原生入口已绑定 |
| `novels_get_agent_settings` | NovelAgentSettings | `()` | `0x180081b90` | [原式](source/bridge-contracts/novels_get_agent_settings.inspection.js) | 前端异步候选＋原生入口已绑定 |
| `novels_save_agent_settings` | NovelAgentSettings | `(e)` | `0x1800823b0` | [原式](source/bridge-contracts/novels_save_agent_settings.inspection.js) | 前端异步候选＋原生入口已绑定 |
| `novels_send_message` | NovelStudioPage | `(e)` | `0x180082cb0` | [原式](source/bridge-contracts/novels_send_message.inspection.js) | 前端异步候选＋原生入口已绑定 |
| `novels_confirm_action` | NovelStudioPage | `(e)` | `0x180084690` | [原式](source/bridge-contracts/novels_confirm_action.inspection.js) | 前端异步候选＋原生入口已绑定 |
| `novels_cancel_operation` | 共享桥接函数 Aan | `(e)` | `0x180085ca0` | [原式](source/bridge-contracts/novels_cancel_operation.inspection.js) | 前端异步候选＋原生入口已绑定 |
| `novels_get_chapter` | NovelStudioBookPanel | `(e, t)` | `0x1800863b0` | [原式](source/bridge-contracts/novels_get_chapter.inspection.js) | 前端异步候选＋原生入口已绑定 |
| `novels_save_chapter` | NovelStudioBookPanel | `(e, t, n, a = {})` | `0x180086d40` | [原式](source/bridge-contracts/novels_save_chapter.inspection.js) | 前端异步候选＋原生入口已绑定 |
| `novels_read_story_file` | NovelStudioBookPanel | `(e, t)` | `0x180087910` | [原式](source/bridge-contracts/novels_read_story_file.inspection.js) | 前端异步候选＋原生入口已绑定 |
| `novels_save_story_file` | NovelStudioBookPanel | `(e, t, n)` | `0x1800880c0` | [原式](source/bridge-contracts/novels_save_story_file.inspection.js) | 前端异步候选＋原生入口已绑定 |
| `novels_get_book_rules` | 未观察到 | `未恢复前端签名` | `0x1800888e0` | [原式](source/bridge-contracts/novels_get_book_rules.inspection.js) | 未观察到前端调用；有原生函数体 |
| `novels_save_book_rules` | 未观察到 | `未恢复前端签名` | `0x180089120` | [原式](source/bridge-contracts/novels_save_book_rules.inspection.js) | 未观察到前端调用；有原生函数体 |
| `novels_write_next` | NovelStudioPage | `(e)` | `0x1800899f0` | [原式](source/bridge-contracts/novels_write_next.inspection.js) | 前端异步候选＋原生入口已绑定 |
| `novels_write_batch` | NovelBatchDialog | `(e)` | `0x18008a790` | [原式](source/bridge-contracts/novels_write_batch.inspection.js) | 前端异步候选＋原生入口已绑定 |
| `novels_scheduler_start` | NovelSchedulerPanel | `(e)` | `0x18008c8a0` | [原式](source/bridge-contracts/novels_scheduler_start.inspection.js) | 前端异步候选＋原生入口已绑定 |
| `novels_scheduler_stop` | NovelSchedulerPanel | `(e = {})` | `0x18008dad0` | [原式](source/bridge-contracts/novels_scheduler_stop.inspection.js) | 前端异步候选＋原生入口已绑定 |
| `novels_scheduler_pause` | NovelSchedulerPanel | `(e = {})` | `0x18008e1b0` | [原式](source/bridge-contracts/novels_scheduler_pause.inspection.js) | 前端异步候选＋原生入口已绑定 |
| `novels_scheduler_status` | NovelSchedulerPanel | `(e = {})` | `0x18008e9d0` | [原式](source/bridge-contracts/novels_scheduler_status.inspection.js) | 前端异步候选＋原生入口已绑定 |
| `novels_audit_chapter` | NovelStudioPage | `(e)` | `0x18008ef80` | [原式](source/bridge-contracts/novels_audit_chapter.inspection.js) | 前端异步候选＋原生入口已绑定 |
| `novels_revise_chapter` | NovelStudioPage | `(e)` | `0x18008fb70` | [原式](source/bridge-contracts/novels_revise_chapter.inspection.js) | 前端异步候选＋原生入口已绑定 |
| `novels_export_book` | NovelStudioPage | `(e, t = "md")` | `0x180090e60` | [原式](source/bridge-contracts/novels_export_book.inspection.js) | 前端异步候选＋原生入口已绑定 |

## 二、输入与返回原式

每项inspection文件保留原桥接函数、真实调用者全文和参数表达式；以下表仅做索引，不把局部变量名擅自解释成后台签名。返回字段是调用者语法候选，需看完整success/cancelled分支。

| 接口 | 调用实参原式 | 出现的响应字段候选 |
|---|---|---|
| `novels_list_books` | 无显式实参 | data, message, success |
| `novels_get_book` | e<br>t<br>s.bookId<br>W.value | bookId, data, message, storyId, success |
| `novels_rename_book` | G.value.id; N.value | bookId, message, success |
| `novels_delete_book` | e.id | message, success |
| `novels_list_sessions` | null<br>e | data, success |
| `novels_get_session` | e | bookId, data, message, pendingAction, success |
| `novels_get_short_story` | n<br>e.storyId<br>s.storyId<br>y.value | bookId, data, message, storyId, success |
| `novels_read_short_story_file` | n.shortStoryId; e.key | data, success |
| `novels_save_short_story_file` | n.shortStoryId; l.value.key; e | message, success |
| `novels_confirm_pipeline_checkpoint` | {<br>  sessionId: r.sessionId \|\| d.value,<br>  tool: r.tool,<br>  checkpoint: r.checkpoint,<br>  storyId: r.storyId,<br>  bookId: r.bookId,<br>  message: r.message,<br>  model: P.value \|\| void 0,<br>  codexReasoningEffort: l,<br>  claudeReasoningEffort: l<br>} | awaitingConfirmation, bookId, data, message, storyId, success |
| `novels_create_session` | {<br>  kind: o,<br>  bookId: t,<br>  workflow: r,<br>  title: a \|\| K9t[r] \|\| "新会话"<br>} | bookId, data, message, success |
| `novels_rename_session` | G.value.id; N.value | bookId, message, success |
| `novels_bind_session_book` | d.value; e.id | bookId, success |
| `novels_delete_session` | e.id | bookId, message, success |
| `novels_update_session_messages` | d.value; n<br>e; o<br>d.value; o | data, message, success |
| `novels_list_style_sources` | 无显式实参 | data, success |
| `novels_get_style_profile` | e.id | data, message, success |
| `novels_save_style_profile` | {<br>  id: l.value \|\| void 0,<br>  title: i.value.trim(),<br>  sourceText: s.value,<br>  styleGuide: c.value<br>} | data, message, success |
| `novels_delete_style_profile` | e.id | message, success |
| `novels_import_style_files` | {<br>  id: l.value,<br>  files: n<br>} | data, message, success |
| `novels_pick_style_files` | 无显式实参 | data, message, success |
| `novels_init_imitation` | t | data, message, success |
| `novels_run_imitation_short` | {<br>  ...e,<br>  sessionId: d.value,<br>  model: P.value \|\| void 0<br>} | awaitingConfirmation, cancelled, data, message, success |
| `novels_init_fanfic` | {<br>  ...n,<br>  fanficMode: p.value<br>} | bookId, data, message, success |
| `novels_init_spinoff` | {<br>  ...n,<br>  parentBookId: d.value,<br>  spinoffType: f.value<br>} | bookId, data, message, success |
| `novels_import_chapters` | {<br>  ...n,<br>  bookId: v.value \|\| void 0,<br>  chapters: a<br>} | bookId, data, message, success |
| `novels_get_agent_settings` | 无显式实参 | data, message, success |
| `novels_save_agent_settings` | {<br>  modelRoutes: {<br>    ...s.value<br>  },<br>  reviewPrompts: c.value.map(y),<br>  customWritePrompts: u.value.map(e => ({<br>    id: e.id,<br>    category: e.category,<br>    title: e.title,<br>    content: e.content,<br>    enabled: Boolean(e.enabled)<br>  })),<br>  reviewCycle: {<br>    maxRetries: Math.max(0, Math.min(5, Number(d.value.maxRetries) \|\| 3)),<br>    passScore: Math.max(0, Math.min(100, Number(d.value.passScore) \|\| 80)),<br>    temperatureStep: Math.max(0, Math.min(.5, Number(d.value.temperatureStep) \|\| .15)),<br>    baseTemperature: Math.max(0, Math.min(2, Number(d.value.baseTemperature) \|\| .65)),<br>    enableLengthNorm: Boolean(d.value.enableLengthNorm),<br>    enablePostWrite: Boolean(d.value.enablePostWrite),<br>    enableAiTells: Boolean(d.value.enableAiTells),<br>    requireSecondaryConfirm: Boolean(d.value.requireSecondaryConfirm)<br>  }<br>} | data, message, success |
| `novels_send_message` | s | action, bookId, cancelled, data, message, success |
| `novels_confirm_action` | {<br>  sessionId: d.value,<br>  model: P.value \|\| void 0,<br>  codexReasoningEffort: A,<br>  claudeReasoningEffort: A<br>} | action, awaitingConfirmation, bookId, data, message, success |
| `novels_cancel_operation` | n.bookId<br>t | bookId |
| `novels_get_chapter` | n.bookId; e.chapterNumber | bookId, data, success |
| `novels_save_chapter` | n.bookId; l.value.chapterNumber; e; {<br>  syncTruth: !0,<br>  replayFrom: !0<br>} | bookId, data, message, success |
| `novels_read_story_file` | n.bookId; e.key | bookId, data, success |
| `novels_save_story_file` | n.bookId; l.value.key; e | bookId, data, message, success |
| `novels_get_book_rules` | 前端未观察到；原生体另查 | 未观察到/动态访问 |
| `novels_save_book_rules` | 前端未观察到；原生体另查 | 未观察到/动态访问 |
| `novels_write_next` | {<br>  bookId: W.value,<br>  sessionId: d.value,<br>  model: P.value \|\| void 0<br>} | bookId, cancelled, data, message, success |
| `novels_write_batch` | {<br>  bookId: n.bookId,<br>  sessionId: n.sessionId,<br>  model: n.model,<br>  count: r.value,<br>  maxConsecutiveFailures: l.value<br>} | bookId, cancelled, data, message, success |
| `novels_scheduler_start` | {<br>  bookId: n.bookId,<br>  sessionId: n.sessionId,<br>  model: n.model,<br>  intervalSeconds: 60 * i.value,<br>  dailyChapterCap: s.value,<br>  maxConsecutiveFailures: c.value<br>} | bookId, data, message, success |
| `novels_scheduler_stop` | 无显式实参 | data, success |
| `novels_scheduler_pause` | {<br>  action: t<br>} | action, data, success |
| `novels_scheduler_status` | 无显式实参 | data, success |
| `novels_audit_chapter` | {<br>  bookId: W.value,<br>  sessionId: d.value,<br>  model: P.value \|\| void 0<br>} | bookId, data, message, success |
| `novels_revise_chapter` | {<br>  bookId: W.value,<br>  sessionId: d.value,<br>  model: P.value \|\| void 0,<br>  mode: "polish"<br>} | bookId, cancelled, data, message, success |
| `novels_export_book` | e.id; t.value | data, message, success |

## 三、原生构造绑定的可重复证据

API模块常量池基址`0x1803aff50`；IDA初始化函数`0x180091e60`调用构造器`0x1803609a0`。从每个构造器的函数指针、名称槽、限定名称槽，按`(槽地址-模块池基址)/8`取已解码常量，得到真实方法名和原生入口。每项保留常量索引、原文件字节范围、初始化C行号，见`source/native-entrypoint-bindings.json`。这比搜索到一个方法字符串多了一层真实函数绑定，仍不证明运行调用成功。

会话服务同法绑定13个方法；pipeline_runner暂未匹配此构造器形态，不能推定该模块没有方法。

## 四、两个规则接口的结论

- `novels_get_book_rules`绑定`0x1800888e0`，IDA函数体490行：加载models.book_rules的load_book_rules，取_state().story_dir(book_id)，成功返回success/data并调用to_dict。原始C行96–150，注释版`source/ida-annotated/0x1800888e0.c`。
- `novels_save_book_rules`绑定`0x180089120`，IDA函数体482行：BookRules.from_dict后传给save_book_rules，并使用作品story_dir。原始C行95–166，注释版同目录。
- 当前已解析前端范围未观察到这两个属性调用。应登记“后端存在、前端调用未知”，不能判为原版假功能或自行新增两个按钮。
- 参数与错误传播的完全恢复仍需结合两个函数体全部异常出口，尚不能仅按这些load名称宣称规则完整生效。

## 五、本阶段验收边界

G01的名称/前端参数原式/调用者/原生入口对照已经落盘，实际后端签名、全部异常和运行验收不在“静态覆盖通过”中。后续G02/G03继续从send_message和confirm_action追进AgentSessionService；G08再按最终请求核对提示词实际使用。
