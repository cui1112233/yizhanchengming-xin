# BatchProject 全路由对象归属收口报告

## 结论

- 所有直接包含 BatchProject `{id}` / `{projectId}` 的 HTTP 对象路由，已在业务 service/store 调用前统一经过 `requireBatchProjectAccess`。
- 中间件顺序为：写操作同源 CSRF -> route capability -> BatchProject ownership -> 业务 handler。没有 capability 的调用者不能探测项目归属。
- 认证启用时，checker 缺失或报错统一返回 `503 AUTH_POLICY_UNAVAILABLE` 和 request id；foreign 返回 `403 AUTH_FORBIDDEN` 和 request id。
- `admin` / `owner` 才传入 `elevated=true`；`dev` / `manager` / `member` 均不因角色绕过。普通 owner、同团队和 `teamId=0` 语义继续由现有 MySQL `auth_batch_project_ownership` 查询判定。
- 未新增 schema；未实现或宣称完成列表分页、搜索或软归档。

## 直接 BatchProject 路由矩阵

以下路由均由 `api/internal/httpapi/server.go` 注册，并使用共享 BatchProject 边界。

| 方法 | 路由 | capability | 读取/变更事实 |
| --- | --- | --- | --- |
| GET | `/api/v1/batch-projects/{id}` | `batch.view` | 项目详情与小说列表 |
| PUT | `/api/v1/batch-projects/{projectId}/books/{bookId}/original-text` | `batch.configure` | 原文 |
| GET | `/api/v1/batch-projects/{projectId}/books/{bookId}/storyboard` | `batch.view` | 分镜文档/卡片 |
| POST | `/api/v1/batch-projects/{projectId}/books/{bookId}/storyboard/cards` | `batch.configure` | 新增分镜卡 |
| PUT | `/api/v1/batch-projects/{projectId}/books/{bookId}/storyboard/cards/{cardId}` | `batch.configure` | 修改分镜卡 |
| DELETE | `/api/v1/batch-projects/{projectId}/books/{bookId}/storyboard/cards/{cardId}` | `batch.configure` | 删除分镜卡 |
| POST | `/api/v1/batch-projects/{projectId}/books/{bookId}/storyboard/reorder` | `batch.configure` | 分镜排序 |
| POST | `/api/v1/batch-projects/{projectId}/books/{bookId}/storyboard/recompile` | `batch.execute` | 分镜重编译 |
| GET | `/api/v1/batch-projects/{id}/novel-panel` | `batch.view` | Novel Panel 工作区 |
| PUT | `/api/v1/batch-projects/{id}/novel-panel` | `batch.configure` | Novel Panel 保存 |
| GET | `/api/v1/batch-projects/{id}/novel-panel/history` | `batch.view` | Novel Panel 历史 |
| POST | `/api/v1/batch-projects/{id}/novel-panel/history/{historyId}/restore` | `batch.configure` | Novel Panel 恢复 |
| GET | `/api/v1/batch-projects/{projectId}/books/{bookId}/shuihuo/segments` | `batch.view` | 水火分段 |
| POST | 同上 | `batch.configure` | 新增水火分段 |
| PUT | `/api/v1/batch-projects/{projectId}/books/{bookId}/shuihuo/segments/{segmentId}` | `batch.configure` | 修改水火分段 |
| POST | `/api/v1/batch-projects/{projectId}/books/{bookId}/shuihuo/segments/reorder` | `batch.configure` | 水火分段排序 |
| GET | `/api/v1/batch-projects/{projectId}/books/{bookId}/shuihuo/assets` | `batch.view` | 媒体资产引用 |
| POST | `/api/v1/batch-projects/{projectId}/books/{bookId}/shuihuo/assets/upload` | `batch.configure` | 上传并绑定媒体资产 |
| GET | `/api/v1/batch-projects/{projectId}/books/{bookId}/shuihuo/assets/{assetId}/content` | `batch.view` | 读取媒体内容 |
| GET | `/api/v1/batch-projects/{projectId}/books/{bookId}/shuihuo/media-tasks` | `batch.view` | 媒体任务 |
| POST | 同上 | `batch.execute` | 创建媒体任务 |
| GET | `/api/v1/batch-projects/{projectId}/books/{bookId}/shuihuo/media-tasks/{taskId}/candidates` | `batch.view` | 候选资产 |
| POST | `/api/v1/batch-projects/{projectId}/books/{bookId}/shuihuo/media-tasks/{taskId}/candidates/{candidateId}/select` | `batch.configure` | 选择候选资产 |
| POST | `/api/v1/batch-projects/{projectId}/books/{bookId}/shuihuo/media-tasks/{taskId}/retry` | `batch.execute` | 重试媒体任务 |
| GET | `/api/v1/batch-projects/{id}/settings` | `batch.view` | 统一配置 |
| PUT | `/api/v1/batch-projects/{id}/settings/production` | `batch.configure` | 生产配置 |
| PUT | `/api/v1/batch-projects/{id}/settings/publishing` | `publish.configure` | 发布配置 |
| GET | `/api/v1/batch-projects/{id}/version-profile` | `batch.view` | 版本配置档 |
| PUT | `/api/v1/batch-projects/{id}/version-profile` | `batch.configure` | 保存版本配置档 |
| POST | `/api/v1/batch-projects/{id}/version-profile/sync-121` | `batch.configure` | 同步 121 配置 |
| POST | `/api/v1/batch-projects/{id}/version-profile/sync-style-types` | `batch.configure` | 同步风格类型 |
| GET | `/api/v1/batch-projects/{projectId}/generation` | `batch.view` | 项目生成摘要 |
| POST | 同上 | `batch.execute` | 批量生成 |
| GET | `/api/v1/batch-projects/{projectId}/books/{bookId}/generation` | `batch.view` | 单书生成摘要 |
| POST | 同上 | `batch.execute` | 单书生成 |
| GET | `/api/v1/batch-projects/{projectId}/books/{bookId}/audio-measurement` | `batch.view` | 音频测量 |
| POST | 同上 | `batch.execute` | 执行音频测量 |
| GET | `/api/v1/batch-projects/{projectId}/books/{bookId}/generation/stages/{stage}` | `batch.view` | 阶段输出 |
| POST | `/api/v1/batch-projects/{projectId}/books/{bookId}/generation/stages/{stage}/retry` | `batch.execute` | 阶段重试 |
| GET | `/api/v1/batch-projects/{projectId}/video` | `batch.view` | 项目视频状态 |
| POST | `/api/v1/batch-projects/{projectId}/books/{bookId}/video` | `batch.execute` | 启动视频任务 |
| POST | `/api/v1/batch-projects/{projectId}/books/{bookId}/merge` | `batch.execute` | 启动视频合成 |

## 间接资源路由（保留先解析所属项目再授权）

| 方法 | 路由 | 解析来源 |
| --- | --- | --- |
| POST | `/api/v1/batch-projects/{projectId}/book-runs/{bookRunId}/retry` | 先授权 URL project，再用 `ProjectIDForBookRun` 校验真实 project 一致 |
| POST | `/api/v1/video-tasks/{taskId}/poll` | `ProjectIDForProductionTask` |
| POST | `/api/v1/video-tasks/{taskId}/cancel` | `ProjectIDForProductionTask` |
| POST | `/api/v1/video-tasks/{taskId}/retry` | `ProjectIDForProductionTask` |
| GET | `/api/v1/video-merge-jobs/{jobId}` | `ProjectIDForMergeJob` |
| POST | `/api/v1/video-merge-attempts/{attemptId}/retry` | `ProjectIDForMergeAttempt` |

这些路由不把资源 ID 当作 BatchProject ID；resolver 失败或归属 checker 失败均 fail-closed。BookRun 因 URL 已含 project，先授权 URL project，只有授权通过才解析 BookRun 归属并校验一致。VIDEO task/merge 的 URL 不含 project，先确认 policy 已接线，再解析真实 project；解析到 foreign project 与 resource 不存在使用完全相同的 `404 VIDEO_NOT_FOUND` envelope，避免资源存在性 oracle。

## 发布路由核验（服务层 defense-in-depth）

- `POST /api/v1/publishing/intents`：body 内 BatchProject ID，由 `publishing.Service.CreateIntent` 同时校验 BatchProject 归属与发布账号归属。
- `GET /api/v1/publishing/intents/{id}`：先读取 intent，再按 intent 的 BatchProject 和账号归属校验。
- `GET /api/v1/publishing/audits?batchProjectId=...`：项目过滤存在时先校验 BatchProject，再执行 owner/team 范围审计查询。
- 这些 API 不含 BatchProject path parameter，未错误套用 path middleware，也未建立第二套授权规则。

## 明确保留的 Intake 边界

`POST /api/v1/intakes/{id}/batch-projects` 继续使用 `requireIntakeAccess("id", ...)`，在创建 BatchProject 前校验 Intake 归属；没有错误套用尚未存在的 BatchProject 归属。

## 测试证据

- 新增 `batch_project_ownership_routes_test.go`：42 个直接 BatchProject 对象路由的行为矩阵；其中 22 个本轮新增保护路由逐项验证 foreign 403、request id、checker nil/error 503 和业务 fake 零调用。
- 验证 capability 在 ownership 之前、跨源 CSRF 在 ownership 之前。
- 验证仅 `admin/owner` 传入 elevated；`dev/manager/member` 不绕过。
- 既有 `publishing/visibility_test.go` 覆盖普通 owner、非零同团队、foreign、`teamId=0` 和 elevated 但项目不存在的 MySQL 查询语义。
- 既有 runtime、VIDEO resource、publishing service 测试继续通过。
- `go test ./internal/httpapi -count=1`：通过。
- `go test ./... -count=1`：全部包通过，退出码 0。

## 未完成

- BatchProject 列表服务端分页、搜索与筛选。
- BatchProject 软归档、恢复、归档可见性和删除/恢复语义。
- 真实 MySQL、登录态浏览器、刷新恢复和公网/ECS 验收。

## Review Fix Round 1

独立审查发现路由级归属边界之外还有四类资源关系缺口，本轮已补齐：

1. `shuihuo_media_tasks` 两条旧写入分支合并为同一条 `INSERT ... SELECT`。写入同时要求：book 属于 BatchProject intake；可选 segment 与 source asset 属于同一 project/book；可选 VIDEO production task 的 job 也属于同一 project/book。`RowsAffected != 1` 返回 `ErrNotFound`，service 在 store 拒绝后不会调用 QueueMediaTask 或 Redis enqueue。
2. `UploadAsset` 在读取请求体、写临时文件或调用 TOS 前，先用只读 MySQL 查询验证 project/book/可选 segment scope。上传后 `CreateAsset` 仍执行原子 scope 二次校验；若该插入失败，只删除本次随机生成并刚上传的实际 object key，不删除任何已有对象。TOS/bucket 未配置仍先返回 `ErrStorageUnavailable`。
3. Publishing `CreateIntent` 对正 `bookId` 调用 `BookBelongsToBatchProject`，以 `batch_projects.intake_id = books.intake_id` 为事实关系；关系不成立返回 `ErrNotFound`，不写 intent/audit。`bookId=0` 的项目级发布不执行多余 book 查询，既有项目与发布账号授权保持不变。
4. BookRun retry 先授权 URL 中的 project，再解析 `bookRunId` 所属项目并校验一致；foreign URL project 与缺失 checker 均不会调用 resolver。VIDEO task/merge 这类无 project URL 的资源路由，也在 resolver 前检查 checker 已接线，再按 resolver 结果授权。

新增/加强测试覆盖：

- Shuihuo 跨 project 的 book、segment、source asset、production task/job 组合拒绝，`RowsAffected=0 -> ErrNotFound`，拒绝后 Queue/Redis enqueue 零调用。
- Upload scope 拒绝时 body read、CreateAsset、TOS Put/Delete 全部零调用；DB 插入失败时 TOS Put 一次、对同一新 key Delete 一次；未配置对象存储继续返回真实 unavailable。
- Publishing 跨项目 book 拒绝且 intent/audit 零写；MySQL 查询明确使用 BatchProject 与 Book 的共享 intake 关系。
- 间接 resource 的 missing checker resolver 零调用、resolver 安全错误、foreign actual project、URL mismatch 和业务 mutation 零调用；错误响应含 request id 且不泄露底层错误。
- 受影响包与 `go test ./... -count=1` 退出码 0。

## Review Fix Round 2

1. `MySQLStore.CreateAsset` 改为单一事务：`Begin -> INSERT ... SELECT -> 同 tx 回读/Scan -> Commit`。INSERT、RowsAffected、LastInsertId 或回读失败均先 Rollback；只有 rollback 成功，错误才标记为“确定未持久化”。Commit 或 Rollback 本身失败属于提交结果不确定，不授权删除 TOS 对象。
2. `UploadAsset` 只在 store 明确返回“确定未持久化”时补偿删除本次随机生成并刚上传的 key。删除使用 `context.WithoutCancel` 保留 request values、脱离原请求取消，再叠加 5 秒 timeout；cleanup 失败时返回值仍 `errors.Is` 原业务错误，同时在安全错误文本中明确 cleanup failure，不删除任何既有 object key。
3. BookRun 在 URL project 授权后，将 resolver not-found 与 actual-project mismatch 收敛为完全相同的 `404 RUNTIME_NOT_FOUND / BookRun 不存在` envelope。VIDEO 无 project URL 的资源，将 resolver not-found 与 foreign resolved project 收敛为相同的 `404 VIDEO_NOT_FOUND`；policy 缺失或 checker error 仍为 `503 AUTH_POLICY_UNAVAILABLE`，授权成功才进入业务 handler。

新增/加强测试覆盖：

- `CreateAsset` 插入成功但同事务回读失败会 Rollback；成功路径在同一事务回读后 Commit；Commit error 标记为结果不确定且不会触发 TOS 删除。
- 原请求在 DB 写失败时已取消，补偿 Delete 仍收到未取消且有限 deadline 的独立 context；Delete cleanup 失败保留原 persistence error，并明确 cleanup failure 边界。
- BookRun not-found 与 URL mismatch 使用固定 request id 验证完整响应 body 完全一致，且 retry 零调用。
- VIDEO resolver not-found 与 foreign actual project 使用固定 request id 验证完整响应 body 完全一致；补齐 nil policy resolver 零调用、checker error 503、authorized 进入业务和所有拒绝 business 零调用。
- `go test ./internal/shuihuo ./internal/httpapi -count=1`：通过。
- `go test ./... -count=1`：全部包通过，退出码 0。
