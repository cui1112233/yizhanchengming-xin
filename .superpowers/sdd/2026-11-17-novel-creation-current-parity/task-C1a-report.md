# C1a 正式章节导入 API：实施及自审报告

2026-11-17。实施起点 `3d0b8a52af9d5305e8744c8957d3266091321e20`，同目录其他聊天将 HEAD 推进至 `9a39c4f0a8fef0dab331e400fca4113d366369de`；未建分支/工作树，未改或提交其他聊天文件。状态：本后端子任务已实现及验证，等待负责人独立审查；C1 六入口整体、C2 未知原生算法、C5b 表单、4001/UI/像素/公网均未据此标完成。

## 行为与证据边界

- 新增 `POST /api/novel-creation/chapter-imports`，输入为强类型 `bookId? / title? / chapters[{title,content,wordCount?}] / knowledgeBookIds[] / idempotencyKey`。已有书名可省，新书名必填；拒绝未知字段及首个 JSON 后的额外内容。成功返回 `book,chapters,message:"章节已导入"`。
- 收据唯一作用域为 owner+kind+key，保存规范输入 SHA256 指纹及原成功结果。事务先争用唯一收据，再锁项目；相同输入返回原结果（先于已归档作品/已删除文风校验），不同输入返回 409。新书、全部正式章、首版本、作品更新时间、continuity dirty、成功收据一起提交；失败全部回滚。
- 已有作品在项目行锁下按最大章节号加一分配，保留数组源序、原正文和标题；空标题按输入源序补“第 N 章”（无空格）。新章 `version=1,run_id=NULL,audit_status=imported`，版本来源记录 `chapter_import/sourceOrder/title/knowledgeBookIds/styles/wordCountMethod/idempotencyKey`。没有生成审核、事实、记忆或大纲。
- 多个文风逐项校验当前账号，保留有序 ID 及完整 ID/title/content 快照和内容指纹；现有账号文风表没有版本/归档列，使用不可变引用快照，不发明版本号、不合并或分析文风。不存在/跨账号/重复 ID 拒绝。
- 导入后 DirtyFrom 指向最早待刷新章节，Revision 加一；已有更早 dirty 保留。后续写作仍被既有事实/记忆刷新门槛拦住，不能空记忆放行。归档/purging、未结束任务（含暂停/待确认）或活动租约拒绝新增章。
- 新书是普通 novel_projects，使用专用零模型构造：原 payload 可证明 genre=other；modelId 空、targetChapterWords=0、空 model policy 是本站适配现有非空 SQL 列的明确存储值，不是已恢复原生默认。旧 CreateProject 的模型必填校验未放宽；后续发送无模型被原校验拒绝，已选模型但 continuity dirty 也不能创建运行。

原证据已直接读 `_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelImportPage.readable.js:196–217` 的 H：按“第…章”拆分、标题移出正文、去空白的 JS UTF-16 长度。本 API 接受已拆章数组，不新增原文拆章接口，不复用 materials.go 的“片段”解析。客户端 wordCount 不可信，服务端沿用本站手动编辑的 Unicode code point 计数（含正文空白），以 `wordCountMethod=unicode_code_points` 标出与 H 的差异。HTTP 4MiB 及本站输入边界（最多500章、单章20万字元、标题255字元、最多100文风）属于本服务边界，不声称为原版上限。原生落盘、文风合并、自动上下文重建算法仍未知，不实现收费维护或假检查。

## 修改范围

领域新增 `chapter_import.go / chapter_import_store.go / chapter_import_test.go / chapter_import_mysql_test.go`；store.go 引入最小导入接口，memory_store.go 新增收据存储，worker_store.go 提取内存 AtomicCreate 的复制提交逻辑。旧 Atomic 仍必须真实项目且保持归档校验，MySQL 旧 Atomic/租约/fence/journal 完全未改。导入章插入在专用 store 文件，不改既有手动编辑语义。

HTTP 新增 `novel_creation_import_routes.go / novel_creation_import_routes_test.go`，原 routes 仅注册新入口。storage 新增 `novel_submission_receipt_schema.go / novel_creation_parity_schema_test.go`，8800315 仅建收据表；owner+kind+key 主键、owner+project 外键、完成结果成对 CHECK；原迁移 SQL/校验值不变，原固定迁移数量断言8→9。注册仍沿用现有 NovelCreationMigrations 组合路径。未改六张会话关联表，也没有预先实施账号 Agent。

最终自审发现新收据外键需接入既有整作品永久清理。经负责人明确批准，仅在 `backend/internal/projectretention/service.go` 原清理子表列表前加入收据，`novel_test.go` 同步精确 owner/project 删除顺序断言；不改变归档/恢复/保留时长、触发条件、用户设置，也不新增删除入口。回执含正文结果，因此应随作品永久删除，不能另留一份已清理内容。

## 测试与真实持久结果

- HTTP 红测最初实际 404；实现后同请求得到201。额外 JSON 红测实际得到201，专用严格解码后400且无书写入。
- 新收据迁移红测明确缺少8800315，追加迁移后通过。内存 Atomic 提取自审发现空项目会绕过旧校验，补红测实际失败后改为显式 requireProject，恢复原保护。
- 三包完整检查：`go test ./internal/novelcreation ./internal/httpapi ./internal/storage -count=1` 全过：3.285s / 105.629s / .075s。内存 Atomic 修正后领域全包再次通过3.536s；最终严格 JSON 改动后全部新 HTTP 用例通过 .066s。包括既有 unknown/journal/save_pending/immutable commit/BASE 回放用例；这些只证明本站保护未回归。
- `go build ./...`（backend）通过，未写运行包或更新4001。
- 新收据导致永久清理的真实红测明确返回1451 FK限制；补收据删除后真实清理通过。另用测试专属子表外键在中途阻止章删除，证明先删的回执也随事务回滚恢复、正文保留；移除测试阻挡后同作品重试清理成功，项目/章节/首版本/收据全零，其他作品11章及旧JSON完全未改。归档期间可重读原回执，恢复也不会再导入章节。最后领域全包3.364s、保留模块全包.050s通过，最终Go全包构建通过。
- 独立 MySQL 8.4.11：完整 container ID、image ID、名称、tmpfs、2GiB限制、loopback `127.0.0.1:63678` 与既有 resource.json 一致；每次 schema/user/nonce 均校验，显式拒绝3306。仅测试专属 schema。
- 新鲜 schema `novelparity2026111740c6c85fa21b` 在旧迁移上先写旧项目、run 快照、旧版本，再应用新迁移两遍；原 JSON 读回字节不变，9项小说迁移的 ledger checksum 全核对。实际表约束验证包括重复 owner+kind+key 拒绝、跨账号 project FK 拒绝、半结果 CHECK 拒绝，以及不同 owner/kind 可以复用 key（临时行事务回滚）。
- 实库已有书最大号8→正式9/10、version1/NULL run、dirty9/revision1、新 service 重新读取和 TXT/Markdown 一致；8个已完成收据并发重放、4个不同 key 并发追加连续号，8个首次提交并发建书返回同一书且仅2章；归档后原回执可重放而新导入拒绝。第二章用测试专属 CHECK 强制真实 SQL 失败，确认不留空书/失败收据。新零模型书无 run/dispatch。
- 实际签名 HTTP→MySQL→重建 router/store→原结果重放及 TXT/Markdown 已过，读得2条version1/NULL run与0个run；继续追加无需新书名，跨账号请求404。实际HTTP持久书 `project-7355987072d3e3583df2004d`。

实库原始日志在 `_local-archive-20261028/novel-parity-20261117/`：`novelparity2026111740c6c85fa21b-20261117-153145-test.log`（新鲜旧→新迁移）、`novelparity2026111740c6c85fa21b-20261117-153914-test.log`（最终首次并发/归档）、`C1a-http-real-mysql.log`。首次实库跑出的章列名错误已按现有 number_no 修正；触发器因最小用户权限拒绝，改用专属 CHECK 注入，未调整服务器全局配置。多次新 schema 留存使512MiB测试tmpfs满，最新 `27b839e61fab` 在旧迁移阶段失败；日志保留，校验该本轮 schema 的 owner nonce 后仅移除这一未完成 schema，再复用40c的已核验库，最终新增并发/归档用例全部通过1.119s。没有将这些失败日志改称绿灯。

永久清理实库红/绿日志分别为同目录 `novelparity2026111740c6c85fa21b-20261117-154339-test.log` / `novelparity2026111740c6c85fa21b-20261117-154409-test.log`；最终包括上述全部实库用例的运行2.000s通过。

## 自审与交接

核对了旧迁移 diff 只有尾部新增8800315、旧 SQL字节没有变化；确认旧 Atomic 项目保护、MySQL旧事务、run恢复逻辑未改。新路径不依赖 provider/文档/TOS；HTTP测试的生成执行器调用即 panic，所有导入保持零run，不启动后台分析。只提交上述明确代码路径与本报告/进度追加；未动前端、提示词、运行包、日常/生产数据库、公网、其他脏文件、被拒绝删除的 `.tmp-task5-base-capture`，未推送。

待负责人独立审查后再决定后续任务；本次不能替代当前UI、原生算法、全入口或正式运行环境验收。
