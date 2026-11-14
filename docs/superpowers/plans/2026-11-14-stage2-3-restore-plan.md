# 阶段 2 / 阶段 3 代码修改计划

- 日期：2026-11-14
- 基线：`v88` @ `3c6653ac`（生产 `qiantie-202611141140`）
- 上游计划：[2026-11-14-automation-reliability-followups.md](2026-11-14-automation-reliability-followups.md)
- 阶段 0 结论：
  - 调用方取消后，`Store.Open` 仍执行完整回源（实测 1504ms 不受影响）。
  - `MarkRehydrated` 失败时本地文件已存在，二次访问会重复下载（remote open 1 → 2）。
  - TOS 客户端：连接 30s、请求/读写 5min、重试 3 次，不存在无限挂起。
- 约束：不改并发算法、不改媒体状态机语义、每阶段独立提交并可回滚。

## 阶段 2：回源取消与超时边界

拆成 2a（基础设施，必做）和 2b（请求路径替换，分批）。

### 2a. 基础设施（1 个提交）

改动文件：`backend/internal/localartifact/store.go`、`backend/internal/app/app.go`、`backend/internal/localartifact/store_restore_test.go`

1. 新增错误与默认值

```go
var ErrRestoreTimeout = errors.New("artifact restore timed out")

const defaultRestoreTimeout = 2 * time.Minute
const maxRestoreTimeout = 10 * time.Minute
```

2. `Store` 增加字段与设置方法

```go
type Store struct {
    // ...existing fields
    restoreTimeout time.Duration
}

func NewStore(root string, maxBytes int64) *Store          // restoreTimeout 默认 2m
func (s *Store) SetRestoreTimeout(d time.Duration)        // <=0 用默认；>10m 截断为 10m
```

3. 拆分上下文版本，旧方法退化为“有界后台上下文”

```go
func (s *Store) Open(storageRef string) (*os.File, error) {
    ctx, cancel := context.WithTimeout(context.Background(), s.restoreTimeout)
    defer cancel()
    return s.OpenContext(ctx, storageRef)
}

func (s *Store) OpenContext(ctx context.Context, storageRef string) (*os.File, error)
func (s *Store) PathContext(ctx context.Context, storageRef string) (string, error)
```

4. 错误与日志

- 回源返回前判断：`errors.Is(err, context.DeadlineExceeded)` → 包装为 `ErrRestoreTimeout`；`context.Canceled` 原样返回。
- 仅在真正触发回源时打日志：`storageRef`、耗时、错误。
- 本地命中路径不打日志，避免放大日志量。

5. 配置注入

`backend/internal/app/app.go` 在 `artifactStore.SetMissingRestorer(resolver)` 附近读取：

```go
QIANTIE_ARTIFACT_RESTORE_TIMEOUT   // Go duration，默认 2m，最大 10m
```

非法值或超界时回退默认并打印一次告警。

6. 测试

- `TestOpenContextStopsWhenCallerCancels`：ctx 取消后 100ms 内返回 `context.Canceled`，且无可见文件。
- `TestOpenAppliesRestoreTimeout`：restorer 阻塞 500ms，超时设 50ms，返回 `ErrRestoreTimeout`，无半成品文件。
- `TestOpenStillReadsExistingLocalFile`：本地已存在时不触发回源。

验收：`go test ./internal/localartifact/...` 全绿；现有 `TestOpenRestoresMissingVerifiedArtifactIntoLocalCache` 等用例保持通过。

风险：把所有 `Open/Path` 调用点默认变成 2 分钟上限。大文件回源可能被截断，需要时用 `OpenContext` 传入更长 ctx 或调高环境变量。

### 2b. 请求路径替换（分 2 个提交）

只替换“HTTP 请求内会阻塞用户”的调用点，其余后台/worker 路径保留 `Open` 包裹的有界行为。

第一批（媒体读取与下载）：

| 文件 | 位置 | 改法 |
|---|---|---|
| `agent_batch_factory_bridge.go` | `openMerge(ctx,...)` 47 行 | `Local.OpenContext(ctx, ...)` |
| `local_executor_artifacts.go` | 98、129 行 | 使用 `r.Context()` |
| `batch_factory_comic_pages.go` | 2326、2583、2639、2770、3125、3199 行 | zip/预览路径统一传 `ctx` |
| `batch_factory_comic_thumbnails.go` | `handleComicPageThumbnail` 及 helper | helper 增加 `ctx` 参数 |
| `batch_factory_v11_merge.go` | 180、192、260 行 | handler ctx |
| `shuihuo_projects.go` | 2404、2502、2569、3072、3754、3776 行 | handler ctx |

第二批（其余素材与资产读取）：

`batch_factory_comic_reference.go`、`batch_factory_comic_reference_data.go`、`comic_materials.go`、`batch_factory_v11_assets.go`、`agent_video_asset_ingest.go`、`agent_batch_factory_audio.go`、`agent_image_generator.go`。

需要给没有 `ctx` 参数的 helper 增加参数，调用链逐层传入；不得在 helper 内部新建 `context.Background()`。

保留不动的路径：

- `batch_factory_automation.go:3222`（任务 worker）
- `batchfactoryv11/local_merge_adapter.go:111`（合并执行器）
- `Store.Path` 的 worker 调用点（`batch_factory_v11_merge.go:70/413`、`batch_factory_v11_video_upload.go:62/118`）：如确认在请求内，可改 `PathContext`，否则保持。

验收：

- `go test ./internal/httpapi/` 相关用例通过。
- 本地 4001 打开漫画页预览、执行器产物下载、合并封面读取，均正常返回。
- 本地断开请求后日志出现 `context canceled`，不再等待完整回源。

风险：helper 签名变更会触达多个调用点，逐个编译验证；不做跨包接口改动。

## 阶段 3：`MarkRehydrated` 状态协调

改动文件：`backend/internal/productionmedia/resolver.go`、`backend/internal/productionmedia/resolver_test.go`（新增）

### 现状行为

```go
body, err := r.Remote.Open(ctx, item.ObjectKey)      // 每次都下载
local.Restore(...)
return r.Repository.MarkRehydrated(ctx, item.ID, ...) // 失败即整体失败
```

### 目标行为

1. 本地已有同尺寸文件时跳过 TOS 下载，只重试仓库标记。
2. 标记失败不再让读取失败：文件已可用，返回成功并记录日志。
3. 本地缺失或尺寸不一致时仍走完整回源。

### 实现

```go
func (r *Resolver) RestoreStorageRef(ctx context.Context, storageRef string) error {
    // ...FindVerifiedByStorageRef 与状态校验不变

    localReady := false
    if path, statErr := r.Local.LocalPath(storageRef); statErr == nil {
        if info, infoErr := os.Stat(path); infoErr == nil && info.Size() == item.ByteSize {
            localReady = true
        }
    }
    if !localReady {
        body, err := r.Remote.Open(ctx, item.ObjectKey)
        // ...错误与长度校验不变
        if err := r.Local.Restore(storageRef, body, item.ByteSize, item.SHA256); err != nil {
            return fmt.Errorf("restore production media %s: %w", storageRef, err)
        }
    }
    if err := r.Repository.MarkRehydrated(ctx, item.ID, purgeAfter); err != nil {
        // 本地文件已可用，读取不应失败；记录以便仓库状态最终收敛。
        log.Printf("production media rehydrate mark failed ref=%s id=%d err=%v", storageRef, item.ID, err)
        return nil
    }
    return nil
}
```

说明：

- 尺寸校验代替重复 SHA 计算；文件由 `Restore` 原子写入且不与外部编辑共享，尺寸一致即视为同一对象。
- 标记失败返回 `nil`：调用方 `Store.OpenContext` 随后能正常打开文件；下一次访问会再次尝试标记，直到数据库恢复。
- 不改变 `verified/cleanup_pending/cleaned` 才允许恢复的前置条件。

### 测试

- `TestResolverSkipsDownloadWhenLocalCopyMatchesSize`：预置同尺寸文件，断言 `Remote.Open` 未被调用、`MarkRehydrated` 调用 1 次。
- `TestResolverMarkFailureKeepsReadWorking`：`MarkRehydrated` 返回错误，断言 `RestoreStorageRef` 返回 nil、本地文件存在；第二次调用不再下载，标记被重试。
- `TestResolverRedownloadsWhenLocalCopyMissing`：删除本地文件后再次调用，断言重新下载。
- `TestResolverRejectsUnverifiedState`：非允许状态仍返回 `os.ErrNotExist`。

验收：`go test ./internal/productionmedia/...` 全绿。

风险：标记长期失败时仓库状态停留在 `cleaned`，但读取不受影响；日志需可检索，后续可加计数指标。

## 提交与发布安排

1. `feat(artifact): 回源支持上下文与超时` —— 阶段 2a。
2. `refactor(httpapi): 媒体读取使用请求上下文` —— 阶段 2b 第一批。
3. `refactor(httpapi): 素材读取使用请求上下文` —— 阶段 2b 第二批。
4. `fix(media): 本地已恢复时跳过重复回源` —— 阶段 3。

每个提交后执行：

```text
go build ./...
go test ./internal/localartifact/... ./internal/productionmedia/...
go test ./internal/httpapi/ -run "Comic|Media|Artifact|Merge|Executor"
```

阶段全部完成后：本地 4001 实际验证 → `scripts/deploy-go-embedded.ps1 -SkipFrontendBuild` → 生产发布并记录 SHA、二进制 SHA256、回滚点。

## 回滚策略

- 阶段 2a 出问题：回退该提交即可恢复“无超时上限”的旧行为，不涉及数据结构。
- 阶段 2b 出问题：逐文件回退，不影响 2a。
- 阶段 3 出问题：回退 resolver 提交；数据库状态机与清理逻辑未改动，无需数据修复。
