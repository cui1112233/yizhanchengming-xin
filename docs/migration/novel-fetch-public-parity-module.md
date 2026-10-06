# 小说获取公网功能补齐 — 模块隔离实现记录

## 当前执行基线

- 目标仓库：`cui1112233/yizhanchengming-xin`
- 固定模块分支：`feat/novel-fetch-public-parity`
- 分支起点：`19df4c4ce8f446df856ec0f0eb1558fd913de51d`
- 新仓根目录未发现 `AGENTS.md`；当前已读取 `README.md` 与 `TASKS.md`。
- 旧仓 `cui1112233/-` 已读取 `AGENTS.md`、`CURRENT_VERSION.md`、小说获取 v88 路由、运行时、版本、知识库、敏感词、排版规则、队列、定时与网络提交边界。
- 历史线索中的公网 SHA `27fa2e12f9401d378d1e6d6298a9d50a27581512` 在旧仓存在，但本任务执行环境无法连接当前公网 `115.190.156.223:3000` 的 `/api/build-info`，因此**不能声称已重新确认当前公网 SHA**。

## 公网/v88 行为契约

本模块按源码对标固定以下语义：

1. 多书城/多平台分组批量 ID，同平台同 ID 去重，不同平台允许相同 ID。
2. 上游必须返回完整原文；`originalRaw` 永久保留完整内容，`maxText` 只限制处理文本，不得把截断文本冒充完整原文。
3. 正文、书名、category、genre、gender、style 等元数据独立持久化。
4. 版本至少支持 `original` 与多 AI 槽位；成功版本不因其他版本失败而丢失，失败重试只补失败/缺失版本，不重复抓已成功原文。
5. 配置在 Run 创建时快照，后续默认模型/版本配置变更不修改已排队 Run。
6. 知识库只把启用条目交给改文模型；敏感词、章节、排版规则是确定性本地处理，模型不可用时不得伪造成功。
7. 定时执行只创建调度消息；真实到点执行必须由统一 Redis Runtime/Worker 接管。
8. 处理记录按 Run/Book/Stage/Attempt 留存，错误必须脱敏。
9. “转入批量工厂”通过边界接口完成，不在小说模块复制批量工厂业务。
10. “提交网络”只创建统一发布意图；真正发布执行由统一发布服务处理。本模块不保存发布密码/Cookie，不实现第二套发布器。

## 本分支模块文件

- `api/internal/novelfetch/types.go`：模块领域模型、Store/Fetcher/TextModel/Dispatcher/BatchFactory/PublishIntent 边界。
- `api/internal/novelfetch/service.go`：多平台批次、完整原文与处理文本分离、多版本改文、规则、知识库、定时派发、失败重试、处理记录、批量工厂与发布意图边界。
- `api/internal/novelfetch/http.go`：可独立挂载的模块级 HTTP handler；**当前未注册到全局路由**。
- `api/internal/novelfetch/*_test.go`：使用模拟上游/模型/派发器/发布意图/批量工厂边界做合成测试。

## 等待任务 A 的共享依赖

以下事项在收到 A 的协作基线前不自行变更：

- 全局 API 路由与鉴权能力映射。
- MySQL 新表/列与 Goose migration 编号。
- 账号/团队归属字段与权限检查的共享类型。
- 后台文本模型目录解析器与凭据获取方式。
- Redis Queue/Lease 的正式 task key/prefix 与 Scheduler/Worker 注册位置。
- TOS 正文/版本大对象的对象 key 与生命周期策略。
- 统一发布服务中“小说版本提交”的正式 Intent 输入结构与 Worker 执行适配。
- 批量工厂创建项目的正式共享接口/类型。
- 前台主导航、路由和共享 API client 接线。
- Docker、部署、ECS、公网切换。

## 真实链路验证状态

当前模块测试仅允许模拟外部依赖：

- 可模拟验证：多平台 ID、正文/元数据写入、完整原文保留、处理文本截断、版本部分成功、失败重试、配置快照、知识库传递、规则处理、调度派发、处理记录、批量工厂边界、发布意图边界。
- 未验证：当前公网实际 SHA 与真实页面行为、真实 121 正文接口、后台真实文本模型、真实 MySQL 持久化、真实 Redis 到点执行、真实 TOS、真实批量工厂接入、真实统一发布执行。
- 不允许用 HTTP 200、队列完成或 mock 返回值替代上述真实链路证据。
