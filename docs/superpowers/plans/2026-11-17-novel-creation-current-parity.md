# 小说创作链路修复实施计划

> **2026-11-17纠偏：当前UI被用户明确拒绝，原Task1–7勾选仅保留历史实施事实，不代表洛水字面对齐通过。** 先依据已纠正规格和两份原版差异审计修订任务，不再按下文旧解释继续堆功能。重点撤回：通用新建表单前置、仿写仅文风输入、六模式共用单章产物、逐请求二次确认、本地检查模型化、六工具统一Drawer。当前不重启4001、不发起模型请求或公网发布。

纠偏待办（原版合同映射，不是新布局方案）：

- [ ] C1 六入口与原专用表单逐字段映射至HTTP/Go/持久实体；分清聊天、建书、短篇和章节导入。
- [ ] C2 Agent原17路由/共享范围/37维及自定义项对照；原生本地检查和归一化算法先补证，未知不得自创冒充。
- [ ] C3 撤回两项本地检查的额外模型步骤，按内容里程碑实现二次确认，保留调用恢复与去重保护。
- [ ] C4 独立短篇产物/大纲/文件及长篇正式导入链；总字数、每章字数、目标章数分开。
- [ ] C5 原左侧六项及底部工具、主工作区切换、专用弹窗、条件右栏；不得只改CSS或统一Drawer。
- [ ] C6 分别验字段边界、零模型路径、实际持久产物和同状态原版画面；收费与公网验证继续单列。

原始证据及当前代码定位见`docs/audit/2026-11-17-novel-original-ui-correction.md`、`docs/audit/2026-11-17-novel-original-agent-correction.md`。当前只完成纠偏分析和规格写回，以上尚未实施，不以旧测试绿灯抵销缺口。

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox syntax for tracking.

**Goal:** 按已批准规格补齐小说创作四块闭环，在本地4001留下真实可核验结果。

**Architecture:** 保留现有Go/React实现；MySQL持久任务、输入快照、租约和内容版本，Redis通知与取消。手动/定时/批量统一生成合同，页面从后端恢复正式状态。

**Tech Stack:** Go、React/Antd、MySQL、Redis、TOS、Go embed。

## Global Constraints

### 当前纠偏执行顺序（用户已指令执行）

先执行C5a原布局框架，随后C1/C2后端合同、C3处理语义、C4专用产物、C5b专用表单与右侧资料、C6总验收。C5a不是整体复刻验收，不触发4001更新；专用表单不能只显示而丢参数。

**C5a文件**：`frontend/src/user/pages/NovelCreationPage.jsx`、`NovelCreationPage.css`；交互测试沿用`novel-creation`现有浏览器测试框架。

C5a实施及审查完成：28fb3d38、修复2c59fe39；完整历史39/39，导航补丁后9/9，独立审查R1后6/6。只表示布局框架，不表示专用表单/语义/像素/4001已验收。

**C1a下一实施任务：正式章节导入API闭环**。依据原`NovelImportPage.readable.js:196–217,438–468`，详见SDD `task-C1-contract-mapping.md`。输入`bookId?、title?、chapters[{title,content,wordCount?}]、knowledgeBookIds[]、idempotencyKey`；已有作品不强制新书名，新作品名必填；仅导入正式章节，绝不创建模型run或后台分析。

文件范围：新增`backend/internal/novelcreation/chapter_import.go`、`chapter_import_store.go`及领域/实库测试，现有store/memory/content-management接口最小扩展；新增`backend/internal/httpapi/novel_creation_import_routes.go`及路由测试、原routes注册；storage新增唯一导入回执迁移及schema测试。不得顺带改六表nullable或账号Agent，也不得部署当前日常数据库。

- [ ] 写领域与HTTP红测，覆盖已有/新书、跨账号、归档/活动任务拒绝、同key重放/改内容冲突、全部章及首版本、失败整批回滚、续写记忆dirty、重读导出，provider调用即失败。
- [ ] receipt(owner,kind,key)唯一键与规范输入指纹；回执/新书/全部章节/首版本/continuity dirty同一事务，失败不留空书。已有书持项目锁分配max+1，不覆盖历史正文。客户端wordCount非权威；采用本站计字并注明非已恢复原算法。
- [ ] 选中文风先校验归属并保存引用，不能不支持却默默丢弃；导入不分析文风、不生成事实。新书使用专用无模型构造，不放宽旧CreateProject或伪造模型ID。未知helper不自创冒充，入口阶段不做收费维护。
- [ ] 保持现有未知结果/租约/保存恢复测试；新迁移不改旧SQL或校验值。先独立测试库验证实际DDL/事务再作完成结论，禁止日常库迁移或运行包更新；仅提交本任务文件并独立审查。

- [ ] 先写失败交互测试：六入口连续且无自加三组；六工具只在左侧下部；Agent/风格/定时/看板切主工作区并能返回对话；导入为Modal、备忘录为独立浮层；切工具和返回保留当前草稿/会话。
- [ ] 按原`NovelStudioSidebar/NovelStudioPage`恢复位置与容器，不改生成接口或把尚未支持的表单伪接通。保留原资料组件、权限校验、作用域和维护提交保护，移除统一工具Drawer。无作品下现有项目级后端限制明确报告，不能偷建作品、伪造全局配置或触发模型。
- [ ] 空会话无作品时不常驻自加当前资料面板；已有作品的真实章节/进度暂保留，待C5b替换为原长短篇产物面板，不称完整复刻。
- [ ] 运行新测试及既有小说浏览器回归；红绿证据保存SDD目录；仅提交本任务文件，独立审查后再进入下一项。原版截图仍未验收。

- 已批准规格：`docs/superpowers/specs/2026-11-17-novel-creation-current-parity-design.md`；N01–N12逐项留证。
- 基准是当前洛水安装版，旧ZIP只作历史参考；恢复片段只供分析，不复制到产品。
- 用户要求直接在v88维护、不建分支；保留无关脏文件。按此要求覆盖技能的新分支/工作树建议。
- 每项红测→最小实现→绿测→审查；实施代理串行，审查只读。不得多代理同时改共享文件。
- 不发起收费模型调用或公网发布；真实收费验收独立记录为待授权，不伪称已通过。
- 仅本地4001为交付入口，最终源码和运行包须一致。提示词只在后端prompt模块。
- 临时证据与构建放`_local-archive-20261028/novel-parity-20261117/`；本计划进度账本放`.superpowers/sdd/2026-11-17-novel-creation-current-parity/progress.md`。

## Task 1: 可信前文与会话恢复 API

**Files:** `backend/internal/novelcreation/{store,service,mysql_store,memory_store}.go`；新增`continuity.go`、`continuity_test.go`；`backend/internal/httpapi/novel_creation_routes.go`及其测试。

**Interfaces:** Store新增`GetChapterMemory(ctx,owner,projectID,chapterID) (Memory,bool,error)`；Service新增`GetSessionSnapshot(ctx,owner,sessionID) (Snapshot,error)`；GET `/api/novel-creation/sessions/{sessionID}/snapshot`。生成快照新增`previous_chapters`（章ID、版本、结尾）和版本化事实/记忆。第一章为空，后续章失败关闭。

- [x] 红测：真实保存第一章及其记忆后发送第二章，捕获Pipeline.RunContext检查PreviousMemory/PreviousTruth与第一章一致；注入记忆读取错误，确认Pipeline未调用。

```go
if string(got.PreviousMemory) == "{}" || len(got.PreviousMemory) == 0 { t.Fatal("lost persisted chapter memory") }
```

- [x] 实现：MySQL按owner/project/chapter关联查询最新记忆；内存库等价；从最新已保存章节和蓝图结算提取输入，不用空对象替代读取失败；按会话返回消息/最新run/章节，空会话可正常返回。
- [x] 验证：`go test ./internal/novelcreation ./internal/httpapi -run 'Continuity|SessionSnapshot' -count=1`，再运行两包相关旧测试。确认跨账号404、无运行空状态、已有正文版本一致。审查后提交。

## Task 2: 完整自动执行输入合同

**Files:** `backend/internal/novelcreation/{types,validation,scheduler,batch,mysql_store,memory_store}.go`；`backend/internal/storage/novel_creation_schema.go`；HTTP routes及测试。

**Interfaces:** 定时/批量输入增加ModeInputs、KnowledgeIDs、NoteIDs、StyleProfileID及相应JSON持久字段；统一输入构造函数服务于SendMessage和调度。迁移使用新增版本，不改写旧迁移。

- [x] 红测：四种附加输入模式创建调度/批次后立即执行，断言发送完整材料；创建缺必填资料时拒绝，不能保存空计划。

```go
if got.ModeInputs["referenceText"] != "参考片段" { t.Fatal("scheduled run dropped reference") }
```

- [x] 实现：复用同一输入校验与材料归属选择；序列化输入快照；旧缺输入计划暂停并显示配置需调整，不能静默猜值。
- [x] 验证：`go test ./internal/novelcreation ./internal/httpapi -run 'Schedule|Batch|ModeInput' -count=1`，覆盖重试/继续快照不变、引用失效与其他账号材料。审查后提交。

## Task 3: 可靠任务领取、租约和中断恢复

**Files:** 新增`backend/internal/novelcreation/{worker,worker_store,worker_test}.go`；修改Store及两实现、Service/Runtime、`backend/internal/app/app.go`、迁移和HTTP starter。

**Interfaces:** 持久待派发、运行租约及调用状态；`StartWorker(ctx,service,interval)`扫描与领取。MySQL条件更新领取/续租/提交；90秒租约、30秒续租。HTTP starter只通知，不直接生成。

- [x] 红测：两个worker竞争同一个run只调用一次；过期token不能提交；已发送结果未知重启后暂停；明确未发送可恢复；已完成正文只重试保存。

```go
if runner.callCount() != 1 { t.Fatal("duplicate upstream dispatch") }
```

- [x] 实现：用户消息/运行/待派发事务；Redis通知失效仍可从数据库扫描；记录发送前/已发送/已完成及步骤产物；取消可打断上下文并阻止后续步骤；删除直接后台Run链及未消费队列路径。
- [x] 验证：`go test ./internal/novelcreation ./internal/httpapi ./internal/app -run 'Novel|Worker|Lease|Recovery' -count=1`，聚焦并发疑点使用race；记录迁移与重启恢复证据。审查后提交。

## Task 4: 定时和批量状态、阈值与日志

**Files:** `backend/internal/novelcreation/{scheduler,batch,types,validation,mysql_store,memory_store}.go`，必要时拆`automation_state.go`；HTTP路由、迁移和测试。

**Interfaces:** Schedule增加DailyChapterCap、MaxConsecutiveFailures、状态/成功计数/日计数/逐轮记录；Batch增加MaxConsecutiveFailures、计数和StoppedReason。定时轮次以计划ID+到期时刻唯一。

- [x] 红测：上海时区跨日、日限10、失败阈值3、作品占用不推进、暂停跨日不恢复、两进程去重、立即执行不改原时刻；批次失败不增加完成章数、恢复不重复成功章。

```go
if after.CompletedChapters != before.CompletedChapters { t.Fatal("failure counted as completed chapter") }
```

- [x] 实现：间隔1–1440分钟、日限1–100、失败阈值1–10；成功保存才计数；未知结果/确定配置错误立即暂停；临时失败有界；停机只补一个轮次，下一次从处理完加间隔计算。
- [x] 验证：自动时间替身测试和MySQL唯一领取测试；运行`go test ./internal/novelcreation ./internal/httpapi -run 'Schedule|Batch|Automation' -count=1`。审查后提交。

## Task 5: Agent模型路由与有界处理

**Files:** `backend/internal/novelcreation/{types,validation,pipeline,service}.go`、pipeline tests、`backend/internal/prompt/novelcreation/`、settings迁移；前端设置面板随后接入。

**Interfaces:** AgentSettings增加ModelRoutes、PassScore、BaseTemperature、TemperatureStep、可空MaxRevisionRounds以区分旧默认和显式0、EnableLengthNorm、EnablePostWrite、EnableAiTells、RequireSecondaryConfirm；所有生效参数写入快照。

- [x] 红测：每个阶段选对模型；0次修订不调用reviser；审核分数与温度生效；关闭步骤无调用；不支持参数不发送；蓝图确认幂等。

```go
if reviseCalls != 0 { t.Fatal("explicit zero revision limit ignored") }
```

- [x] 实现：保持旧作品已保存值，新设置默认采用批准规格；结构修复和内容修订分开限额；保留有审核提醒的有效正文，不无限生成。
- [x] 验证：`go test ./internal/novelcreation ./internal/prompt/... -run 'Novel|Pipeline|Agent|Revision' -count=1`，红/绿记录和实际请求参数替身记录。审查后提交。

## Task 6: 资料导入、文风、备忘录与内容版本

**Files:** 新增`backend/internal/novelcreation/{materials,content_management}.go`及测试；Store/MySQL/MemoryStore、迁移与HTTP routes；复用项目已有上传/TOS和文档解析能力。

**Interfaces:** 文件上传→解析任务→章节列表→选中范围；StyleProfile和Note引用快照；作品/会话Rename/Archive/Restore；ChapterVersion列表及显式恢复；短篇独立结果查询；TXT/Markdown导出。

- [x] 红测：合法TXT/DOCX/PDF、扫描件/加密PDF/超20MiB/超8文件拒绝；无标题单片段；文风或参考二选一；备注引用生效；跨账号/删除引用/章节改版拒绝旧快照覆盖。

```go
if len(snapshot.SelectedChapterIDs) != 2 { t.Fatal("selected chapter range lost") }
```

- [x] 实现：TOS只放原文件/导出文件，MySQL存解析和版本；纯文风统计不调用AI；归档复用全站机制；编辑/恢复均新增版本并标记记忆刷新，不覆盖历史。
- [x] 验证：上传和解析服务聚焦测试、HTTP/存储事务测试；资料失败状态可查且不产生空可用内容；审查后提交。

## Task 7: 前端交互与恢复

**Files:** `frontend/src/user/pages/NovelCreationPage.jsx/.css`；拆分`novel-creation/`下会话、调度、资料、Agent面板和输入合同；`frontend/src/shared/api/novelCreation.js`；新增真正交互测试。

**Interfaces:** 使用Task1–6正式API；会话快照恢复、游标轮询、AbortController与单在途请求；草稿键按账号/作品/会话/模式隔离。

- [x] 红测：点击新建产生新会话；刷新恢复run与正文；慢旧请求不能串作品；IME Enter不发送；定时/批量传材料；所有阈值/Agent字段实际发送并显示保存结果。

```js
assert.equal(createdSession.id === previousSession.id, false);
assert.equal(outgoing.modeInputs.referenceText, '参考片段');
```

- [x] 实现：恢复/新建/草稿逻辑，六模式不同资料视图，导入章节/风格/备注，管理操作和版本列表；不叠透明弹窗，不暴露密钥或内部错误；状态与后端一致。
- [x] 验证：前端交互测试加构建；不能用仅匹配源码字符串的4个测试代替浏览器事件。审查后提交。

## Task 8: 同状态对照、本地4001交付与总审

**Files:** 新增`docs/acceptance/2026-11-17-novel-creation-current-parity.md`与验证证据；只使用已有本地Go embed启动脚本。

- [ ] 记录当前源码SHA、当前4001进程/二进制路径、前端入口和备份，检查其他聊天是否有待运行变更；不混装不同源码的前端和后端。
- [ ] 运行相关Go测试、前端交互测试/构建、Go embed构建；只重启确认的4001进程。
- [ ] 浏览器N01–N11逐项验收，与原版相同尺寸和状态对照；需要真实模型的N12单独标记预算待授权，不伪造成功样例。
- [ ] 每项写明输入/运行ID/持久结果/截图/失败提示/边界，记录仍未解决的原生静态与像素证据缺口。
- [ ] 全范围只读审查，修复重大问题，再核对运行SHA与源码。达到规格完成门槛才报告整体完成；不自动上线。

## 执行检查

规格1–2对应全局约束及Task8；规格3–4对应Task1/2/6/7；规格5对应Task1/5；规格6对应Task3/7；规格7对应Task2/4；规格8对应Task6/7；规格9对应Task7/8；规格10–12由各任务测试和Task8验收覆盖。按此顺序执行，不在任务间重复询问是否继续。
