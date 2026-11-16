# 小说创作链路与洛水证据审计

日期：2026-11-17（北京时间）。范围：原包身份、历史分析/规格追溯、当前源码静态检查、本地4001只读页面、现有专项测试。未调用收费模型，未创建业务作品，未修改业务代码，未发布。此报告不是完整功能对齐验收。

## 1. 先回答当前状态

已有分析文档、规格、计划和Go/React实现；4001能打开小说创作。但是完整链路仍有明确缺口，不能宣称六模式、长篇记忆、定时、多章续写、刷新恢复全部可用，也不能宣称与洛水一比一。

## 2. 原包、安装版、历史恢复目录

| 对象 | 实际位置和本轮证据 |
| --- | --- |
| 用户ZIP | `F:/软件/洛水2026.zip`，344737065字节，SHA256 `934e43c03115587777050970332deeb0cf653e1c0bfb80b180c9e386f89bcf8c`，12257项；直接检查目录无APK、无`modules/novels.pyd`，前端有`index-CnO3GE3B.js`和`index-DALB2Co1.js`。 |
| 当前安装目录 | `F:/软件/洛水2026/洛水2026`；当前`dist/index.html:7`加载`index-C1aA08SG.js`；`modules/novels.pyd`当前SHA256 `0471b2a77ee1aad06f4dd30e58b2515ed89b3a20d780ad5ae0f64ab3e4552853`。这是Windows x86_64库，依赖python312.dll，不是Android APK。 |
| 旧报告版本 | `docs/reverse-engineering/2026-11-12-luoshui-novel-studio-rea.md:29`引用`index-DszfAkKD.js`；该文件仍存在，但不是当前HTML入口；该报告`:32`模块哈希`2a685a81...`也与当前模块不同。不能混为同一基线。 |
| 旧REA工作目录 | `G:/Backup/Documents/ChatGPT/v88/work/luoshui2026-rea/`，含scope、timeline、workitems、evidence/E-001至E-006；它是证据工作目录，不是完整反编译源码目录。 |
| 旧恢复产物 | 旧小说报告`:7-16`只写`recovered/frontend/...`及`recovered/nbc/...`相对路径。本轮在源码目录、G盘工作目录、F软件目录检索指定`novels_api.nbc`、`pipeline_runner.nbc`、`state_manager.nbc`未定位。不能声称它们已完整留档，也不能由此断言从未生成过。 |
| 本轮留档 | `F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/`，`manifest.json`记录身份，`zip-original/`和`installed-original/`分别保存7个相关原始文件。**原始文件快照不是恢复源码**，未复制账户数据库、浏览器数据或密钥。 |

Craft证据：ZIP清单`ev_6e301cfebb114581804a49790282c7629e4f8866feeede5c0c1d73884fca3100`（受读取字节上限约束，部分文件未哈希；嵌套包未展开）；当前模块身份`ev_aced9d81a9b608eaed13aa38750788f5bd624848ae8c727698e61e2b180c862d`。

## 3. 有没有调用Craft / REA

- 历史文档`docs/plan/2026-11-12-novel-creation-rewrite.md:15`说某轮Craft未连接；后来`docs/reverse-engineering/2026-11-12-luoshui-novel-studio-rea.md:20`写入Craft证据`ev_6ea52f27...`，并明确控件树不等于像素验收。这两条属于不同阶段，不能扩大为完整对齐证明；本轮尚未重新导入并核验旧Craft bundle。
- 本轮确实调用Craft：investigation/session、brief、plan、target_open、triage、archive、finding、unknown、handoff及证据导出。开始时没有打开项目、旧发现或旧证据；这仅说明本轮会话初始为空，不能证明历史没有使用。
- 本轮确实调用REA：打开旧前端包，执行`analyze_javascript_application`。父证据`ev_cb09dd1004f99793a150dc49bf6ae367b3c0d0f21a36dcc4d482d6db9b82f7e9`，摘要`ev_5a28b5c1950f509be89acbf8f28bf4219a4da41f05a403921f852a7d24283102`。512MiB分析进程在解析时耗尽内存，覆盖partial、parsed_javascript_files=0；不能把这次分析称作成功恢复调用图。
- 用户提供的APK技能已评估，但本目标没有APK，正确路线是Windows/Python原生模块与Web前端分析。

## 4. 规格到实现的现状

已有规格：`docs/superpowers/specs/2026-11-11-novel-creation-clean-room-parity-spec.md`和`docs/requirements/2026-11-12-novel-creation-spec.md`；实施记录：`docs/plan/2026-11-12-novel-creation-rewrite.md`。

已观察到的代码主链：`NovelCreationPage.send` → `shared/api/novelCreation.js` → `novel_creation_routes.go:116` → `Service.SendMessage:72`保存消息/运行快照并入队 → 路由启动后台Run → `Service.Run:170` → `Pipeline.Run:195`规划/正文/审核/修订/事实/记忆 → `SaveChapterBundle` → `GetSnapshot` → 页面1.5秒轮询。

这证明主干有实现，不证明真实模型和数据库的整链路验收成功。系统提示词确实从后端`internal/prompt/novelcreation`接入。工具面板也存在settings/materials/notes/dashboard/schedules/batches接口，不能简单称为全是假按钮。

## 5. 本轮确认的缺口与可验证规则

以下是当前源码证据及影响推导，不冒充原版运行结果。

| 优先级 | 缺口、证据 | 输入 / 输出 / 边界 / 失败验收规则 |
| --- | --- | --- |
| P0 | 前文事实和记忆未回读。`service.go:115-142`虽读取章节数和蓝图，却把`previous_truth`、`previous_memory`固定为空；`:223`直接传入流水线。Store接口也没有读取已保存记忆的方法。 | 第N章输入必须携带截至N-1章的事实、摘要、人物变化和伏笔；输出保留与来源章版本关联的快照。第一章允许为空，后续章读取失败须明确停止，不能默默当空。 |
| P0 | Redis队列入队但正常执行走即时后台调用。`service.go:160`调用Enqueue；`app.go:361-368`直接启动Run；非测试代码未找到Dequeue消费调用，也未找到小说运行启动恢复扫描。`runtime.go:211`仅定义LPOP。 | 提交成功后必须由可靠工作进程领取；服务重启后待执行/运行中任务可恢复或明确失败。必须验证中断后不会永远排队、不重复生成或收费。当前只能判断有明显恢复缺口风险，未做进程中断实测。 |
| P0 | 定时和批量丢失模式必填输入。`validation.go:79-83,96-116`要求仿写referenceText、番外existingSetting、续写previousText；`scheduler.go:152-156`与`batch.go:234-241`调用SendMessage却不传ModeInputs/KnowledgeIDs。 | 新建、立即执行、到点、重试共用同一不可变输入合同。缺必填输入应在保存时拒绝并指出字段；不能保存成功后到点才失败。四种有额外输入的模式当前存在确定的校验冲突。 |
| P1 | 页面刷新/切会话缺恢复。`NovelCreationPage.jsx:184-203`只读会话列表；Run轮询依赖内存中的currentRun；API模块只有已知runId的Get，无按会话恢复快照调用。 | 重新进入同一作品/会话可读取历史消息、最新运行与章节；无运行时显示空状态；读失败应给重试入口。当前4001账号作品为0，因此未做历史内容实测。 |
| P1 | 新建会话按钮没有真正创建会话。`NovelCreationPage.jsx:541-550`仅清空选择和草稿；`:236-239`又自动复用该模式已有会话。 | 用户点新建后应创建新的会话ID；发送必须进入新会话，旧消息不丢失。创建失败保留草稿。 |
| P1 | 锁15分钟，运行允许45分钟，未见续租。`service.go:12,182`、`runtime.go:225-253`固定TTL；`app.go:363`运行超时45分钟。 | 慢任务仍须保持排他或通过数据库租约/版本拒绝并发；锁失效停止新调用并清晰提示。当前为并发风险推导，未做45分钟实测。 |
| P1 | 调度先推进下次时间，失败也可能错过本轮。`scheduler.go:152-165`SendMessage后不论其是否失败均更新LastRunAt/NextRunAt；幂等键使用本次Unix秒，无跨进程原子领取。 | 每个计划到期时刻有唯一执行记录；临时占用应等待/重试，确定失败才终止；每次失败有独立可读记录。跨进程去重需专门验收。 |
| P1 | 模式输入切换直接清空，未见按会话/模式草稿保存。`NovelCreationPage.jsx:672`setModeInputs({})；旧clean-room规格4.3要求保存草稿。 | 切换模式后返回应保留未发送文本和附加输入；放弃时明确提示。 |
| P2 | 规格自身矛盾。`requirements/...:6`仍写六种横向模式；REA报告`:20`已纠正为左栏两列三行。验收记录`plan/...:65,84,100`写4011，与同文`:51`要求4001不一致。 | 一个明确原版版本、一个已确认UI合同、4001实际运行身份、逐项证据；不得用旧设计或另一个端口的测试宣告完成。 |

## 6. 本轮实际验证

- 本地4001真实账号导航中有小说创作，点击打开`/novel-creation`；看到六模式、六工具、模型选择、参考文本框及开始按钮；作品数0。只读观察，没有创建或生成。
- runtime-build-info观察：git_sha/runtime_manifest_sha均`9f28ee75`，frontend_entry=`/assets/user-Ca5fZt50.js`；审计时源码HEAD已为`ed1fd7d349bb3f72097151d8769cd940971c9020`。源码和当前运行包不能视为完全相同。
- `go test ./internal/novelcreation ./internal/httpapi -run 'NovelCreation|Novel|Pipeline|Schedule|Batch' -count=1`：两包通过（0.029s、29.577s）。范围受-run过滤，不是全量测试；Provider多为测试替身。
- `node --test src/user/pages/NovelCreationPage.source.test.js`：4/4通过。这四项只检查源码文案/路由/引用，不是浏览器操作测试，不能覆盖上述缺口。
- 未完成：真实模型生成、六模式同输入对照、长篇连续两章事实回读、真实到点调度、重启恢复、慢任务并发、原版当前页面逐状态与同尺寸截图。禁止将这些写成通过。

## 7. 后续设计候选（未实施）

建议选择A：以明确版本的洛水小说创作为基准，先补可追溯原包/恢复记录，再修“输入快照→可靠领取→正文/事实/记忆保存→页面恢复”的闭环，最后补六模式、定时/批量和页面逐状态对照。保留现有有效实现，逐缺口增加失败场景验收。

B：只修当前暴露的输入和页面问题，成本低，但恢复/并发/长篇连续性仍无保证。C：整体推倒重写，成本最高，且仍需先固定原版版本和规格。

开始修改业务行为前，应确认对标基准：用户ZIP，还是已更新的当前安装版。建议使用实际需要对齐的安装版，同时保留ZIP作历史基线；之后按用户调用的brainstorming流程确认修复设计，再写实施计划。
