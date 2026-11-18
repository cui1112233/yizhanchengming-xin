# CORE-01：无书会话与确认建书

用户已批准按 TASKS.md 顺序执行；此规格细化已批准的 CORE-01，不宣称已完成其余74项。沿用 v88，不新建分支，不推送、不公网发布，不消耗模型测试额度。

## 原版合同与独立实现边界
原 NovelStudioPage.readable.js:290-297、420-460：会话可无book，聊天可能返回action；:645-705 确认动作读取当前模型并切书籍会话；NovelConfirmCard.readable.js:23-44、88-102：确认创建作品与取消/已执行状态。
原 native 存储、系统提示词与具体推理算法未恢复。本轮采用自家 Go/MySQL/模型接口独立实现可验证外部行为，提示词明确是CM自编，放后端prompt模块。取消持久化是本项目可靠性规则，不伪称已恢复原取消实现。

## 本轮完整小范围
先接长篇原创的建书前会话，不以短篇通用章节假装独立story。其他入口仍按登记待修。

输入：新会话创建编号、自然语言创作要求、自家文本模型、思考强度、上下文预算；确认只提交动作编号、当前模型和当前思考要求，客户端不能改服务端建议内容。
输出：未确认时仅保存会话/用户消息/模型回复/待确认动作，无隐藏作品、无生产run。模型真实回复可只是交流，也可建议create_book。确认后在单事务中建作品、生产会话、初始蓝图任务与派发记录，返回固定关联编号。接下来沿用已有蓝图确认和生产恢复能力。
边界：同一创建/发送编号只能对应冻结输入；同会话同一时刻一轮模型请求；所有读写带owner；状态与请求快照持久化；重发原编号返回已有结果，不再收费。列表/返回/刷新不能自动生成。动作取消不建书，重复确认返回同一结果，过期/取消动作拒绝确认。历史消息和轮次有上限，超限明确提示。
失败：输入/配置失败发生在发送前；发出后结果不明保存unknown，不自动重试。原始回复先落盘，再本地解析；已存原文的解析/恢复不再调用模型。保存失败保留原输入和编号；换会话/账号后的迟到响应不得抢导航或覆盖草稿。取消模型建议和停止已发上游请求不是同一语义。

## 接口
POST /studio-sessions → {session}；GET /studio-sessions → {sessions}。
GET /studio-sessions/{id} → StudioSnapshot。
POST /studio-sessions/{id}/messages → StudioSnapshot（正常调用当前配置模型）；客户端遇丢回包只GET核对。
POST /studio-sessions/{id}/actions/confirm → {session,project,productionSession,run}。
POST /studio-sessions/{id}/actions/cancel → StudioSnapshot。
均位于 /api/novel-creation 下；身份来自现有登录，不接受客户端owner。

## 存储与通知
新增独立studio表，避免修改现有project非空约束。利用已有AtomicCreate；确认时锁studio记录，在同一外层事务内创建正式对象和现有dispatch。内部使用quietRuntime，外层提交后才通知Redis；通知失败不得重复建书。

## 登记与验收
- [x] 读取原合同与现有事务保护，完成独立只读审查。
- [x] 后端存储/模型规划/确认/取消/恢复；单元及真实隔离MySQL验证。
- [x] HTTP身份/严格参数/状态码和无隐藏作品回归。
- [x] 页面从空态或长篇入口进入无书会话；建议卡、确认、取消、刷新与迟到响应。
- [x] 构建并核对本地4001身份；无费界面验收，付费生成仍不运行。

本轮真实模型测试预算为0，测试使用明确标注的受控提供方；不把受控文本当作线上生成结果。

## 2026-11-18 执行记录

实现：`backend/internal/novelcreation/studio_*.go`、`backend/internal/storage/novel_studio_schema.go`、`backend/internal/prompt/novelcreation/studio.go`、`backend/internal/httpapi/novel_studio_routes*.go`、`frontend/src/user/pages/novel-creation/StudioChat.*`及页面接线。追加迁移8800316，原迁移不重写。

测试：领域/存储/提示词普通测试通过；HTTP NovelCreation|NovelStudio通过。独立MySQL 8.4.11已验证双迁移、隔离、重开、事务回滚、重复确认和持久派发。三组临时数据库/专用用户均清理，剩余0。最终SQL日志：`_local-archive-20261028/novel-parity-20261117/novelparity202611175d8bc221f865-20261118-173125-test.log`。

浏览器受控接口：65项首次63通过，2项旧入口预期调整后定向2/2通过；新增“讨论确认→正式生产视图→刷新恢复”1/1通过。最终完整重复回归66/66通过，0失败、0跳过（browser-tests.log）。控制接口不等于真实模型生成。前端生产构建通过（保留已有大包提醒）。race测试因CGO未启用不可用；普通并发确认测试通过。

可靠性补修：坏格式保留草稿；明确发送前拒绝解除锁定；无书/已分配会话间迁移完整输入；旧账号/会话/新建迟到回包不抢导航；超长回复保存有标记的预览、原长度和哈希。实际限制上下文预算，确认交接最后冻结的有界输入。思考强度是提示词要求，不伪装成上游原生reasoning参数。

本地已交付：代码提交1886b338。4001新进程48032运行novel-core01-20261118/qiantie-core01.exe，接口git_sha=1886b338，入口/assets/user-BEB_z00c.js。实际空会话→刷新读回通过：原作品1个、新增未关联作品会话1个、模型调用0。截图4001-prebook-restored.jpg、runtime-acceptance.json及build-info.json在同名归档。首轮实测遇本地MySQL handler commit长等，后续已读回成功；未改数据库配置或重启数据库。构建来自共享脏工作区，不能称隔离干净提交构建。

待完成：真实付费规划/正文结果、原版同状态像素与native提示词仍未验证。本轮只交付长篇原创切片，不给其余模式盖章。

## 当前代码定位（提交1886b338）

| 规则 | 文件与行号 |
|---|---|
| 不建隐藏作品的会话创建 | backend/internal/novelcreation/studio_service.go:101 |
| 有界历史输入 | backend/internal/novelcreation/studio_service.go:74 |
| 持久消息与规划状态 | backend/internal/novelcreation/studio_service.go:235 |
| 取消不建作品 | backend/internal/novelcreation/studio_service.go:453 |
| 确认及事务交接 | backend/internal/novelcreation/studio_service.go:483 |
| MySQL固定编号及行锁 | backend/internal/novelcreation/studio_store.go:40、99 |
| 页面确认/取消真实请求 | frontend/src/user/pages/novel-creation/StudioChat.jsx:65 |
| 返回正式作品及生产会话 | frontend/src/user/pages/NovelCreationPage.jsx:299 |
| CM独立系统提示词 | backend/internal/prompt/novelcreation/studio.go:5 |
