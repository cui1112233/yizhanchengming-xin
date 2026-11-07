# Production Media TOS and Retention Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将漫画图片、资产图片、制作音频和合成视频的正式存储迁到 TOS，并默认每 7 天定时清理已验证的服务器副本。

**Architecture:** 保留现有 `storage_ref` 作为稳定逻辑标识，新增 MySQL 媒体迁移账本记录本地文件、TOS object key、大小、SHA256和状态。业务创建文件后只入队一次；低并发 Worker 上传并验证 TOS，已验证才进入本地回收。媒体读取先本地、本地不存在时从 TOS 读取，使现有 API URL 和前端无需重写。

**Tech Stack:** Go 1.26, MySQL 8, Volcengine TOS Go SDK v2, React 18, Ant Design 5, Node test runner, Go embed, systemd

## Global Constraints

- `productionRetentionDays` 默认值是 7，允许值仅为 3/7/14/30。
- 保留时长只删除服务器本地副本，不删除 TOS 正式对象、数据库记录、小说、剧本或项目配置。
- 任何未上传、未验证或校验失败的本地文件绝不自动删除。
- 存量迁移并发数固定为 2，每次只领取 50 条，重启后只继续未完成项。
- 系统盘达到 75% 时提前删除已验证 TOS 的最旧本地副本，直到降到 65%。
- 系统盘达到 85% 且无安全可删副本时，拒绝新的媒体写入并返回明确错误。
- 前端不包含 TOS 凭证，Git、构建产物、运行日志和 API 响应均不输出 AccessKey/SecretKey。
- 发布仍是 React 构建产物通过 Go embed 嵌入单个 Linux 二进制，不使用 GitHub Actions。
- 保留当前未提交工作树；每次只提交当前任务的确切文件。

---

## File Structure

**New backend package**

- `backend/internal/productionmedia/types.go`: 媒体账本状态、远程对象接口、仓库接口。
- `backend/internal/productionmedia/tos_store.go`: 通用 TOS Put/Head/Open/Delete，实现每次读写超时。
- `backend/internal/productionmedia/mysql_repository.go`: 媒体账本入队、领取、状态更新、统计。
- `backend/internal/productionmedia/migrator.go`: 低并发上传、HEAD 验证、SHA256 记录、断点续跑。
- `backend/internal/productionmedia/resolver.go`: 现有 `storage_ref` 的本地优先/TOS 回退读取。
- `backend/internal/productionmedia/purger.go`: 默认 7 天、每小时扫描、75%/65% 水位清理和 85% 写入保护。

**Existing backend integration**

- `backend/internal/config/config.go`: 新增制作媒体 TOS 环境变量。
- `backend/internal/storage/batch_factory_v12_schema.go`: 新增 `1200106` 媒体账本迁移。
- `backend/internal/storage/batch_factory_v12_schema_test.go`: 断言新表、索引和版本。
- `backend/internal/batchfactoryv11/comic_pages_mysql.go`: 漫画页图片入库时幂等入队。
- `backend/internal/batchfactoryv11/book_asset_images_mysql.go`: 资产图片入库时幂等入队。
- `backend/internal/batchfactoryv11/merge_store_mysql.go`: 本地合成结果成功时入队。
- `backend/internal/batchfactoryv11/local_merge_adapter.go`: 保留工作目录立即清理，远程 Output 启用时直传 TOS。
- `backend/internal/localartifact/store.go`: 增加可注入的远程缺失读取器和写入水位检查。
- `backend/internal/app/app.go`: 创建 TOS store、账本、Migrator、Purger 并注入路由。
- `backend/internal/httpapi/user_config.go`: 默认保留时长改为 7。
- `backend/internal/httpapi/storage_status.go`: 新增账号存储统计 API。
- `backend/internal/httpapi/router.go`: 注册存储统计 API 与依赖。

**Frontend**

- `frontend/src/shared/api/batchFactoryV12.js`: 默认保留时长改为 7，新增存储状态请求。
- `frontend/src/user/pages/SettingsPage.jsx`: 默认 7 天，说明真实清理范围，展示存储状态。
- `frontend/src/user/pages/settings-retention-source.test.js`: 更新精确文案和默认值断言。

### Task 1: Change Every Retention Default to 7 Days

**Files:**
- Modify: `backend/internal/httpapi/user_config.go`
- Modify: `backend/internal/batchfactoryv11/local_artifact_purge.go`
- Modify: `frontend/src/shared/api/batchFactoryV12.js`
- Modify: `frontend/src/user/pages/SettingsPage.jsx`
- Test: `backend/internal/httpapi/user_config_test.go`
- Test: `backend/internal/batchfactoryv11/local_artifact_purger_test.go`
- Test: `frontend/src/user/pages/settings-retention-source.test.js`
- Test: `frontend/src/shared/api/batchFactoryV11.test.js`

**Interfaces:**
- Consumes: existing `productionRetentionDays` 3/7/14/30 contract.
- Produces: `DefaultProductionRetentionDays = 7` and matching frontend fallback.

- [ ] **Step 1: Write failing backend default tests**

Add assertions equivalent to:

```go
func TestDefaultProductionRetentionDaysIsSeven(t *testing.T) {
	config := cloneConfigMap(defaultUserConfig)
	if got := int(config["productionRetentionDays"].(int)); got != 7 {
		t.Fatalf("productionRetentionDays=%d want=7", got)
	}
	if normalizeLocalMergedArtifactRetentionDays(0) != 7 {
		t.Fatal("invalid retention must fall back to seven days")
	}
}
```

- [ ] **Step 2: Write failing frontend source/default tests**

Require all three settings fallbacks and the API fallback to use 7:

```js
assert.match(source, /productionRetentionDays:\s*7/);
assert.match(source, /\? Number\(config\.productionRetentionDays\) : 7/);
assert.match(apiSource, /includes\(days\) \? days : 7/);
```

- [ ] **Step 3: Run focused tests and confirm failure**

Run:

```powershell
cd F:\脚本测试\一战晟铭-v88\backend
go test ./internal/httpapi ./internal/batchfactoryv11 -run Retention -count=1
cd ..\frontend
node --test src/user/pages/settings-retention-source.test.js src/shared/api/batchFactoryV11.test.js
```

Expected: tests fail because current defaults are 3.

- [ ] **Step 4: Implement the single default constant**

Use this backend constant and make all invalid values resolve to it:

```go
const DefaultProductionRetentionDays = 7
```

Set `defaultUserConfig["productionRetentionDays"]`, frontend form load fallback, save fallback, initial value and `productionRetentionDays()` API fallback to 7.

- [ ] **Step 5: Run focused tests and confirm pass**

Run the Step 3 commands. Expected: PASS.

- [ ] **Step 6: Commit only Task 1 files**

```powershell
git add -- backend/internal/httpapi/user_config.go backend/internal/httpapi/user_config_test.go backend/internal/batchfactoryv11/local_artifact_purge.go backend/internal/batchfactoryv11/local_artifact_purger_test.go frontend/src/shared/api/batchFactoryV12.js frontend/src/shared/api/batchFactoryV11.test.js frontend/src/user/pages/SettingsPage.jsx frontend/src/user/pages/settings-retention-source.test.js
git commit -m "fix: default production retention to seven days"
```

### Task 2: Add Production Media TOS Configuration

**Files:**
- Modify: `backend/internal/config/config.go`
- Modify: `backend/internal/config/config_test.go`
- Create: `backend/internal/productionmedia/types.go`

**Interfaces:**
- Consumes: process environment.
- Produces: `productionmedia.TOSConfig` and validated `config.Config` fields.

- [ ] **Step 1: Write failing configuration tests**

Add tests for complete configuration, empty configuration and partial configuration rejection:

```go
func TestFromEnvReadsProductionMediaTOS(t *testing.T) {
	t.Setenv("QIANTIE_PRODUCTION_MEDIA_TOS_ENDPOINT", "https://tos.example.com")
	t.Setenv("QIANTIE_PRODUCTION_MEDIA_TOS_REGION", "cn-test")
	t.Setenv("QIANTIE_PRODUCTION_MEDIA_TOS_BUCKET", "qiantie-media")
	t.Setenv("QIANTIE_PRODUCTION_MEDIA_TOS_ACCESS_KEY", "access")
	t.Setenv("QIANTIE_PRODUCTION_MEDIA_TOS_SECRET_KEY", "secret")
	t.Setenv("QIANTIE_PRODUCTION_MEDIA_TOS_PREFIX", "production-media")
	cfg, err := FromEnv()
	if err != nil { t.Fatal(err) }
	if cfg.ProductionMediaTOSBucket != "qiantie-media" || cfg.ProductionMediaTOSPrefix != "production-media" {
		t.Fatalf("production media config=%+v", cfg)
	}
}
```

- [ ] **Step 2: Run the focused config test**

Run: `go test ./internal/config -run ProductionMediaTOS -count=1`

Expected: FAIL because fields do not exist.

- [ ] **Step 3: Add exact configuration fields and validation**

Add these fields:

```go
ProductionMediaTOSEndpoint  string
ProductionMediaTOSRegion    string
ProductionMediaTOSBucket    string
ProductionMediaTOSAccessKey string
ProductionMediaTOSSecretKey string
ProductionMediaTOSPrefix    string
```

Rules: all five connection fields are either all empty or all non-empty; prefix defaults to `production-media` and is trimmed of leading/trailing `/`.

- [ ] **Step 4: Define storage contracts**

In `types.go`, define:

```go
package productionmedia

type TOSConfig struct {
	Endpoint, Region, Bucket, AccessKey, SecretKey, Prefix string
}

type RemoteObject struct {
	Key, MediaType, SHA256 string
	ByteSize int64
}

type ObjectStore interface {
	Put(context.Context, RemoteObject, io.Reader) error
	Head(context.Context, string) (RemoteObject, error)
	Open(context.Context, string) (io.ReadCloser, error)
	Delete(context.Context, string) error
}
```

- [ ] **Step 5: Run tests**

Run: `go test ./internal/config ./internal/productionmedia -count=1`

Expected: PASS.

- [ ] **Step 6: Commit**

```powershell
git add -- backend/internal/config/config.go backend/internal/config/config_test.go backend/internal/productionmedia/types.go
git commit -m "feat: add production media TOS configuration"
```

### Task 3: Add the Durable Media Migration Ledger

**Files:**
- Modify: `backend/internal/storage/batch_factory_v12_schema.go`
- Modify: `backend/internal/storage/batch_factory_v12_schema_test.go`
- Create: `backend/internal/productionmedia/mysql_repository.go`
- Create: `backend/internal/productionmedia/mysql_repository_test.go`

**Interfaces:**
- Consumes: MySQL connection and stable logical `storage_ref`.
- Produces: migration `1200106`, `productionmedia.Repository`, idempotent enqueue and claim/update operations.

- [ ] **Step 1: Write the failing migration test**

Require migration `1200106` and these columns/indexes:

```go
for _, required := range []string{
	"batch_factory_v12_media_objects", "logical_ref", "owner_username",
	"media_kind", "local_storage_ref", "tos_object_key", "byte_size",
	"sha256", "media_type", "state", "retention_days", "verified_at",
	"purge_after", "attempt_count", "last_error", "idx_bfv12_media_claim",
} {
	if !strings.Contains(ddl, required) { t.Fatalf("missing %s", required) }
}
```

- [ ] **Step 2: Run migration test and confirm failure**

Run: `go test ./internal/storage -run ProductionMedia -count=1`

Expected: FAIL because migration is absent.

- [ ] **Step 3: Add migration 1200106**

Create this table through `V12ProductionMediaStatements()`:

```sql
CREATE TABLE IF NOT EXISTS batch_factory_v12_media_objects (
  logical_ref VARCHAR(255) NOT NULL PRIMARY KEY,
  owner_username VARCHAR(191) NOT NULL,
  media_kind VARCHAR(32) NOT NULL,
  local_storage_ref VARCHAR(255) NOT NULL,
  tos_object_key VARCHAR(768) NULL,
  byte_size BIGINT NOT NULL DEFAULT 0,
  sha256 CHAR(64) NULL,
  media_type VARCHAR(128) NOT NULL,
  state VARCHAR(24) NOT NULL,
  retention_days INT NOT NULL DEFAULT 7,
  verified_at DATETIME(6) NULL,
  purge_after DATETIME(6) NULL,
  attempt_count INT NOT NULL DEFAULT 0,
  last_error VARCHAR(512) NULL,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  deleted_at DATETIME(6) NULL,
  KEY idx_bfv12_media_claim (state,updated_at,logical_ref),
  KEY idx_bfv12_media_purge (state,purge_after,logical_ref),
  KEY idx_bfv12_media_owner (owner_username,media_kind,state)
) ENGINE=InnoDB
```

- [ ] **Step 4: Write failing repository tests**

Cover:

```go
repo.Enqueue(ctx, MediaObject{LogicalRef: "lea_a.png", Owner: "alice", Kind: "comic_image", RetentionDays: 7})
repo.Enqueue(ctx, sameObject) // must not reset verified/committed rows
repo.ClaimUploads(ctx, 50)    // uses FOR UPDATE SKIP LOCKED
repo.MarkVerified(ctx, ref, objectKey, size, sha256, verifiedAt, purgeAfter)
repo.MarkRetry(ctx, ref, cause)
```

- [ ] **Step 5: Implement repository with bounded retries**

Use `INSERT ... ON DUPLICATE KEY UPDATE` only for rows in `pending/failed`; claim `pending/failed` rows whose retry time is due, set `uploading`, increment attempt count, and reclaim `uploading` rows older than 15 minutes.

- [ ] **Step 6: Run tests**

Run: `go test ./internal/storage ./internal/productionmedia -run 'ProductionMedia|Repository' -count=1`

Expected: PASS.

- [ ] **Step 7: Commit**

```powershell
git add -- backend/internal/storage/batch_factory_v12_schema.go backend/internal/storage/batch_factory_v12_schema_test.go backend/internal/productionmedia/mysql_repository.go backend/internal/productionmedia/mysql_repository_test.go
git commit -m "feat: add production media migration ledger"
```

### Task 4: Implement TOS Upload, Verification, and Remote Read

**Files:**
- Create: `backend/internal/productionmedia/tos_store.go`
- Create: `backend/internal/productionmedia/tos_store_test.go`

**Interfaces:**
- Consumes: `TOSConfig`, `RemoteObject`.
- Produces: `NewTOSStore(TOSConfig) (*TOSStore, error)` implementing `ObjectStore`.

- [ ] **Step 1: Write failing tests using injected SDK functions**

Test that:

```go
object := RemoteObject{Key: "production-media/comic/images/lea_a.png", MediaType: "image/png", ByteSize: 123, SHA256: strings.Repeat("a", 64)}
// Put passes content type and checksum metadata.
// Head returns the exact size and checksum.
// A size/checksum mismatch is an error.
// Open returns the object body.
// Every object key remains under the configured prefix.
```

- [ ] **Step 2: Run test and confirm failure**

Run: `go test ./internal/productionmedia -run TOS -count=1`

Expected: FAIL because the TOS store is absent.

- [ ] **Step 3: Implement key confinement and TOS operations**

Use this constructor contract:

```go
func NewTOSStore(cfg TOSConfig) (*TOSStore, error)
```

Reject empty/absolute/traversal keys, require HTTPS endpoint, set metadata `qiantie-sha256`, use a 30-second per-request context for metadata calls, and preserve the caller's longer upload context for streaming writes.

- [ ] **Step 4: Run tests**

Run: `go test ./internal/productionmedia -run TOS -count=1`

Expected: PASS.

- [ ] **Step 5: Commit**

```powershell
git add -- backend/internal/productionmedia/tos_store.go backend/internal/productionmedia/tos_store_test.go
git commit -m "feat: add verified production media TOS store"
```

### Task 5: Implement the Resumable Upload Worker

**Files:**
- Create: `backend/internal/productionmedia/migrator.go`
- Create: `backend/internal/productionmedia/migrator_test.go`

**Interfaces:**
- Consumes: `Repository`, `ObjectStore`, `*localartifact.Store`.
- Produces: `Migrator.RunOnce(context.Context) (MigrationSummary, error)` and `Migrator.Start(context.Context, time.Duration)`.

- [ ] **Step 1: Write failing worker tests**

Required cases:

```go
func TestMigratorUploadsVerifiesThenCommits(t *testing.T)
func TestMigratorKeepsLocalFileWhenPutFails(t *testing.T)
func TestMigratorKeepsLocalFileWhenHeadMismatch(t *testing.T)
func TestMigratorDoesNotReuploadVerifiedRow(t *testing.T)
func TestMigratorProcessesAtMostFiftyWithTwoWorkers(t *testing.T)
```

The success test must assert state order `pending -> uploading -> verified`, exact byte size/SHA256, and no local deletion.

- [ ] **Step 2: Run tests and confirm failure**

Run: `go test ./internal/productionmedia -run Migrator -count=1`

Expected: FAIL because Migrator is absent.

- [ ] **Step 3: Implement deterministic object keys**

Use:

```go
func ObjectKey(kind, logicalRef string) (string, error) {
	// comic_image -> comic/images/<logicalRef>
	// asset_image -> comic/assets/<logicalRef>
	// audio       -> audio/<logicalRef>
	// video       -> video/clips/<logicalRef>
	// merge_video -> video/merged/<logicalRef>
}
```

No date component is used, so a retry always targets the same idempotent key.

- [ ] **Step 4: Implement upload and verification order**

For each claimed row: open the local file once to calculate SHA256 and exact byte size, close it, reopen it for upload with the already-known SHA256 metadata, call Head, require exact `ByteSize` and SHA256 metadata, then call `MarkVerified` with `purge_after = verified_at + retentionDays`. A changed local size or hash between the two reads fails the row and never marks it verified.

- [ ] **Step 5: Start at boot and every minute**

`Start` runs one immediate pass, then every minute. It uses a semaphore of size 2 and one 50-row claim per pass.

- [ ] **Step 6: Run tests**

Run: `go test ./internal/productionmedia -run Migrator -count=1 -race`

Expected: PASS with no race report.

- [ ] **Step 7: Commit**

```powershell
git add -- backend/internal/productionmedia/migrator.go backend/internal/productionmedia/migrator_test.go
git commit -m "feat: migrate production media to TOS safely"
```

### Task 6: Add Local-First, TOS-Fallback Reads

**Files:**
- Modify: `backend/internal/localartifact/store.go`
- Create: `backend/internal/localartifact/store_test.go`
- Create: `backend/internal/productionmedia/resolver.go`
- Create: `backend/internal/productionmedia/resolver_test.go`

**Interfaces:**
- Consumes: logical ref, ledger repository, local store, TOS store.
- Produces: `MissingRestorer.RestoreStorageRef(context.Context, string) error`, atomic local restoration, and a compatible `localartifact.Store.Open(string) (*os.File, error)`.

- [ ] **Step 1: Write failing resolver tests**

Test this exact order:

```go
// local exists -> read local; repository and TOS are not called
// local missing + state verified/cleanup_pending/cleaned -> download TOS object,
// verify size/hash, atomically restore the exact logical ref, then open local file
// local missing + state pending/uploading/failed -> return os.ErrNotExist
// unknown logical ref -> return os.ErrNotExist
// interrupted or corrupt download -> no partial file is visible
```

- [ ] **Step 2: Add a narrow fallback interface to localartifact**

```go
type MissingRestorer interface {
	RestoreStorageRef(context.Context, string) error
}

func (s *Store) SetMissingRestorer(restorer MissingRestorer)
func (s *Store) Restore(storageRef string, source io.Reader, maxBytes int64) error
```

Do not change the public `Open(string) (*os.File, error)` contract. It first performs the existing confined local open. Only on `os.ErrNotExist` does it call the configured restorer once and repeat the local open. `Restore` validates the logical ref with the existing confinement rules, writes to a sibling temporary file with a hard byte limit, fsyncs, closes and atomically renames it into place. This keeps all existing HTTP range and seek behavior intact and never exposes a partial TOS download.

- [ ] **Step 3: Implement resolver**

Resolver reads the ledger by logical ref and restores only rows whose remote object was verified. It never constructs a TOS key from untrusted route input without a ledger lookup. It downloads through `tosStore.Open`, verifies the recorded byte size and SHA256 while writing through `localartifact.Store.Restore`, and marks the row rehydrated so its next local purge is no earlier than the account retention window.

- [ ] **Step 4: Run focused and package tests**

Run:

```powershell
cd F:\脚本测试\一战晟铭-v88\backend
go test ./internal/localartifact ./internal/productionmedia ./internal/httpapi -run 'Artifact|Comic|MergeMedia|Resolver' -count=1
```

Expected: PASS; range requests, thumbnails, comic content and merge media still work.

- [ ] **Step 5: Commit**

```powershell
git add -- backend/internal/localartifact/store.go backend/internal/localartifact/store_test.go backend/internal/productionmedia/resolver.go backend/internal/productionmedia/resolver_test.go backend/internal/httpapi
git commit -m "feat: read migrated media from TOS fallback"
```

### Task 7: Enqueue Images, Audio, Video Clips, and Merge Outputs Exactly Once

**Files:**
- Modify: `backend/internal/batchfactoryv11/comic_pages_mysql.go`
- Modify: `backend/internal/batchfactoryv11/book_asset_images_mysql.go`
- Modify: `backend/internal/batchfactoryv11/merge_store_mysql.go`
- Modify: `backend/internal/batchfactoryv11/types.go`
- Modify: `backend/internal/httpapi/local_executor_artifacts.go`
- Test: corresponding existing `*_test.go` files

**Interfaces:**
- Consumes: domain rows containing local `storage_ref` or local merge output URL.
- Produces: `MediaQueue.Enqueue(context.Context, MediaObject) error` calls inside the same business transaction where possible.

- [ ] **Step 1: Define the queue interface**

```go
type MediaObject struct {
	LogicalRef, Owner, Kind, LocalStorageRef, MediaType string
	RetentionDays int
}

type Queue interface {
	Enqueue(context.Context, MediaObject) error
}
```

- [ ] **Step 2: Write failing idempotency tests**

Cover one test for each kind:

```text
comic_image, asset_image, audio, video, merge_video
```

Each test creates the same domain record twice or retries after an ambiguous commit and asserts exactly one media ledger row with the same logical ref.

- [ ] **Step 3: Enqueue domain records**

- `CreateComicPageImage`: enqueue non-empty local `StorageRef` as `comic_image`.
- `CreateBookAssetImage`: enqueue non-empty local `StorageRef` as `asset_image`.
- local executor artifact completion: enqueue MP3 as `audio`, MP4 as `video`, images only when attached to a comic/asset domain row.
- merge success: parse only the existing safe `/merge-media/merge_*` route and enqueue `<artifactID>.mp4` as `merge_video`.

Do not enqueue remote HTTP URLs or `tos://` values.

- [ ] **Step 4: Use owner retention settings**

Add a `RetentionResolver` adapter around `UserConfigStore.GetConfig`. Valid 3/7/14/30 values are returned; missing/invalid values return 7.

- [ ] **Step 5: Run tests**

Run:

```powershell
go test ./internal/batchfactoryv11 ./internal/httpapi -run 'ComicPageImage|BookAssetImage|LocalExecutorArtifact|Merge.*Media|Retention' -count=1
```

Expected: PASS and retry tests show no duplicate upload rows.

- [ ] **Step 6: Commit**

Stage only the files actually changed by this task, then:

```powershell
git commit -m "feat: queue completed production media for TOS"
```

### Task 8: Implement Timed Cleanup and Disk-Watermark Protection

**Files:**
- Create: `backend/internal/productionmedia/purger.go`
- Create: `backend/internal/productionmedia/purger_test.go`
- Modify: `backend/internal/localartifact/store.go`
- Modify: `backend/internal/localartifact/store_test.go`

**Interfaces:**
- Consumes: verified ledger rows, local files, injectable disk usage function.
- Produces: hourly local cleanup, 75%/65% pressure cleanup, 85% media-write guard.

- [ ] **Step 1: Write failing purger tests**

Required cases:

```go
func TestPurgerDeletesOnlyVerifiedDueLocalCopy(t *testing.T)
func TestPurgerNeverDeletesPendingUploadingOrFailed(t *testing.T)
func TestPurgerPressureModeOldestFirstUntilBelowSixtyFivePercent(t *testing.T)
func TestWriteGuardRejectsAtEightyFivePercentWithoutSafeCandidate(t *testing.T)
func TestWriteGuardAllowsAtEightyFivePercentAfterSafeCleanup(t *testing.T)
```

- [ ] **Step 2: Run and confirm failure**

Run: `go test ./internal/productionmedia ./internal/localartifact -run 'Purger|WriteGuard' -count=1`

Expected: FAIL because purger and guard are absent.

- [ ] **Step 3: Implement disk usage injection**

```go
type DiskUsageFunc func(path string) (usedPercent float64, err error)

type Purger struct {
	Repository Repository
	Local *localartifact.Store
	DiskUsage DiskUsageFunc
	Root string
	Now func() time.Time
}
```

Production implementation uses `syscall.Statfs`; tests use deterministic percentages.

- [ ] **Step 4: Implement cleanup order**

Normal mode claims `verified/cleanup_pending` rows whose `purge_after <= now`. Pressure mode starts at `usedPercent >= 75`, claims only verified local copies ordered by `verified_at, logical_ref`, deletes until the next disk sample is below 65, then stops. A missing local file is marked `cleaned`, not retried forever.

- [ ] **Step 5: Add media write guard**

Before `SaveImage`, `SaveMP3`, `SaveMP4` and `SaveMedia` create a new file, call the injected guard. At 85%, run one safe pressure-clean pass; if usage remains at least 85%, return `localartifact.ErrStoragePressure` without creating a partial file.

- [ ] **Step 6: Start cleanup at boot and hourly**

The worker runs immediately on start and every hour. Stale `qiantie-local-merge-*` directories older than 30 minutes remain handled by `reclaimStaleLocalMergeWorkspaces`.

- [ ] **Step 7: Run race and package tests**

Run:

```powershell
go test ./internal/productionmedia ./internal/localartifact ./internal/batchfactoryv11 -run 'Purger|WriteGuard|LocalMerge' -count=1 -race
```

Expected: PASS with no race report.

- [ ] **Step 8: Commit**

```powershell
git add -- backend/internal/productionmedia/purger.go backend/internal/productionmedia/purger_test.go backend/internal/localartifact/store.go backend/internal/localartifact/store_test.go
git commit -m "feat: enforce media retention and disk watermarks"
```

### Task 9: Wire Workers and Add Storage Status API

**Files:**
- Modify: `backend/internal/app/app.go`
- Modify: `backend/internal/app/app_test.go`
- Create: `backend/internal/httpapi/storage_status.go`
- Create: `backend/internal/httpapi/storage_status_test.go`
- Modify: `backend/internal/httpapi/router.go`

**Interfaces:**
- Consumes: Config, DB, UserConfigStore, LocalArtifacts, TOS store.
- Produces: running migrator/purger/resolver and `GET /api/storage/production-media/status`.

- [ ] **Step 1: Write failing app wiring tests**

Verify:

```go
// complete TOS config -> migrator, resolver and purger are created
// empty TOS config -> app starts in local-only compatibility mode and no media is deleted
// partial TOS config -> config validation fails before serving
```

- [ ] **Step 2: Write failing API test**

Expected response shape:

```json
{
  "defaultRetentionDays": 7,
  "localBytes": 4084811622,
  "diskUsedPercent": 72.5,
  "tosVerified": 1200,
  "pendingUpload": 50,
  "pendingCleanup": 10,
  "failed": 1,
  "nextScanAt": "2026-11-07T18:00:00Z",
  "writesPaused": false
}
```

The route requires an authenticated user and returns owner-scoped row counts; disk percentage is host-wide.

- [ ] **Step 3: Wire one shared UserConfigStore instance**

Create `userConfigStore := httpapi.NewUserConfigStore(cfg.ScriptHistoryDir)` once, pass it both to RouterOptions and the retention resolver. Do not construct two independent stores.

- [ ] **Step 4: Wire production media components**

When TOS config is complete: construct TOS store, MySQL repository, resolver, migrator and purger; set the local store's missing opener and write guard; start migrator every minute and purger every hour. When config is empty: keep the existing local behavior and disable deletion of non-merge media.

- [ ] **Step 5: Register and test status route**

Run:

```powershell
go test ./internal/app ./internal/httpapi -run 'ProductionMedia|StorageStatus' -count=1
```

Expected: PASS.

- [ ] **Step 6: Commit**

```powershell
git add -- backend/internal/app/app.go backend/internal/app/app_test.go backend/internal/httpapi/storage_status.go backend/internal/httpapi/storage_status_test.go backend/internal/httpapi/router.go
git commit -m "feat: run production media storage workers"
```

### Task 10: Update Settings UI and Explain the Real Cleanup Scope

**Files:**
- Modify: `frontend/src/shared/api/batchFactoryV12.js`
- Modify: `frontend/src/user/pages/SettingsPage.jsx`
- Modify: `frontend/src/user/pages/settings-retention-source.test.js`
- Create: `frontend/src/user/pages/productionMediaStorage.js`
- Create: `frontend/src/user/pages/productionMediaStorage.test.js`

**Interfaces:**
- Consumes: storage status API.
- Produces: default 7-day selector, accurate explanation and visible storage health summary.

- [ ] **Step 1: Write failing UI/source tests**

Require these exact Chinese ideas in rendered/source content:

```text
默认 7 天
只清理已上传并验证 TOS 可读的服务器本地副本
不删除 TOS 正式文件、数据库、项目配置和未上传成功的文件
合成中间文件在任务结束后立即清理
```

- [ ] **Step 2: Write formatter tests**

`productionMediaStorage.js` exports:

```js
export function formatStorageBytes(bytes) {}
export function storagePressureTone(percent, paused) {}
export function summarizeProductionMediaStatus(status) {}
```

Test 0 bytes, 3.9GB, 74.9%, 75%, 85% and paused states.

- [ ] **Step 3: Run tests and confirm failure**

Run:

```powershell
cd F:\脚本测试\一战晟铭-v88\frontend
node --test src/user/pages/settings-retention-source.test.js src/user/pages/productionMediaStorage.test.js
```

Expected: FAIL because formatter/status panel is absent and old copy is inaccurate.

- [ ] **Step 4: Implement status request and settings panel**

Load `/api/storage/production-media/status` when the storage section is visible. Show local size, disk percentage, verified/pending/failed counts, next scan and whether new media writes are paused. A status-request failure must show a retry action and must not block saving retention days.

- [ ] **Step 5: Run focused and full frontend tests**

Run:

```powershell
node --test src/user/pages/settings-retention-source.test.js src/user/pages/productionMediaStorage.test.js src/shared/api/batchFactoryV11.test.js
npm test
npm run build
```

Expected: focused tests pass; full tests/build pass or any pre-existing unrelated failures are recorded with an old/new comparison.

- [ ] **Step 6: Commit**

```powershell
git add -- frontend/src/shared/api/batchFactoryV12.js frontend/src/user/pages/SettingsPage.jsx frontend/src/user/pages/settings-retention-source.test.js frontend/src/user/pages/productionMediaStorage.js frontend/src/user/pages/productionMediaStorage.test.js
git commit -m "feat: show production media retention status"
```

### Task 11: Migrate Existing Media Without Deleting It

**Files:**
- Create: `backend/cmd/production-media-migrate/main.go`
- Create: `backend/cmd/production-media-migrate/main_test.go`
- Modify: `backend/internal/productionmedia/mysql_repository.go`
- Modify: `backend/internal/productionmedia/mysql_repository_test.go`

**Interfaces:**
- Consumes: existing comic image, asset image, merge job and referenced audio/video rows.
- Produces: idempotent pending ledger rows and a dry-run count report.

- [ ] **Step 1: Write failing discovery tests**

Use SQL mock rows to cover:

```text
1075 comic page image refs -> comic_image
222 book asset image refs -> asset_image
local /merge-media/merge_* URLs -> merge_video
local executor MP3/MP4 refs still referenced by active jobs -> audio/video
remote HTTP/TOS URLs -> skipped
missing local files -> reported, not queued as verified
```

- [ ] **Step 2: Implement dry-run and enqueue modes**

CLI contract:

```text
production-media-migrate -dry-run
production-media-migrate -enqueue -limit 50
production-media-migrate -status
```

Default is dry-run. There is no command that deletes local files.

- [ ] **Step 3: Run command tests**

Run:

```powershell
go test ./cmd/production-media-migrate ./internal/productionmedia -run 'Discover|MigrateCommand' -count=1
```

Expected: PASS.

- [ ] **Step 4: Run local dry-run against the configured review database**

Run the command with the existing local review environment. Compare discovered logical refs to actual files; require no duplicate logical refs and report missing-file rows separately.

- [ ] **Step 5: Commit**

```powershell
git add -- backend/cmd/production-media-migrate backend/internal/productionmedia/mysql_repository.go backend/internal/productionmedia/mysql_repository_test.go
git commit -m "feat: discover existing media for TOS migration"
```

### Task 12: Full Verification, Linux Binary Build, Canary Deployment, and Migration

**Files:**
- Build: `frontend/dist`
- Sync mechanically: `backend/web/dist`
- Build: `.release/qiantie-YYYYMMDDHHmm`
- Deploy remotely: `/opt/www/qiantie/qiantie-YYYYMMDDHHmm`
- Modify remotely: `/opt/www/qiantie/qiantie.env`

**Interfaces:**
- Consumes: all completed source tasks and production TOS credentials.
- Produces: verified single Linux binary, one-book canary migration, then bounded production migration.

- [ ] **Step 1: Run backend verification**

Run:

```powershell
cd F:\脚本测试\一战晟铭-v88\backend
go test ./internal/config ./internal/storage ./internal/productionmedia ./internal/localartifact ./internal/batchfactoryv11 ./internal/httpapi ./internal/app ./cmd/production-media-migrate -count=1
go test ./internal/productionmedia ./internal/localartifact ./internal/batchfactoryv11 -race -count=1
```

Expected: PASS.

- [ ] **Step 2: Run frontend verification and embed assets**

Run:

```powershell
cd F:\脚本测试\一战晟铭-v88\frontend
npm test
npm run build
robocopy dist ..\backend\web\dist /MIR
if ($LASTEXITCODE -ge 8) { throw "robocopy failed: $LASTEXITCODE" }
```

Expected: tests/build pass and embedded asset directory matches `frontend/dist`.

- [ ] **Step 3: Build one Linux amd64 binary**

Run:

```powershell
$buildId = 'qiantie-' + (Get-Date -Format 'yyyyMMddHHmm')
$env:GOOS='linux'; $env:GOARCH='amd64'; $env:CGO_ENABLED='0'
go -C F:\脚本测试\一战晟铭-v88\backend build -trimpath -ldflags "-s -w" -o "F:\脚本测试\一战晟铭-v88\.release\$buildId" ./cmd/qiantie
Get-FileHash "F:\脚本测试\一战晟铭-v88\.release\$buildId" -Algorithm SHA256
```

Expected: one ELF amd64 binary and a recorded SHA256.

- [ ] **Step 4: Back up production env and add TOS variables without printing secrets**

Copy the already-configured reference-asset TOS endpoint, region, bucket and credentials into these new production-media and merge variable names inside the protected production env. Read and write the values only on the server; do not print them to the terminal or plan output:

```text
QIANTIE_PRODUCTION_MEDIA_TOS_ENDPOINT
QIANTIE_PRODUCTION_MEDIA_TOS_REGION
QIANTIE_PRODUCTION_MEDIA_TOS_BUCKET
QIANTIE_PRODUCTION_MEDIA_TOS_ACCESS_KEY
QIANTIE_PRODUCTION_MEDIA_TOS_SECRET_KEY
QIANTIE_PRODUCTION_MEDIA_TOS_PREFIX=production-media
QIANTIE_BATCH_FACTORY_V11_MERGE_TOS_ENDPOINT
QIANTIE_BATCH_FACTORY_V11_MERGE_TOS_REGION
QIANTIE_BATCH_FACTORY_V11_MERGE_TOS_BUCKET
QIANTIE_BATCH_FACTORY_V11_MERGE_TOS_ACCESS_KEY
QIANTIE_BATCH_FACTORY_V11_MERGE_TOS_SECRET_KEY
```

Expected: env remains mode 600; logs show only key names and `SET`, never values.

- [ ] **Step 5: Upload, hash-verify, switch symlink, and restart**

Upload to a `.new` path, compare local/remote SHA256, chmod 755, rename to the final timestamped name, then atomically switch `/opt/www/qiantie/qiantie`. Restart `qiantie` and verify `/health`; on failure switch the symlink back to `qiantie-202611071612` and restart.

- [ ] **Step 6: Verify migration/schema startup before enqueueing data**

Run:

```bash
systemctl is-active --quiet qiantie
curl --fail --silent http://127.0.0.1:4000/health
mysql -Nse "SELECT version FROM schema_migrations WHERE version=1200106; DESCRIBE batch_factory_v12_media_objects;" qiantie_v88
```

Expected: migration 1200106 exists and no upload rows have been deleted.

- [ ] **Step 7: Canary one book**

Select one non-critical book with comic images and a merge output. Enqueue only its refs, wait for `verified`, confirm browser preview/original/download/merge playback, then wait through one manual cleanup invocation and confirm the same URLs still work through TOS fallback.

- [ ] **Step 8: Enqueue the remaining stock in bounded batches**

Run repeated `-enqueue -limit 50` passes. Keep worker concurrency at 2. After each batch, record pending/uploading/verified/failed counts, service latency, disk percentage and TOS errors. Stop on rising failure count, HTTP latency regression or disk >=85%.

- [ ] **Step 9: Observe for 24 hours before deleting migrated stock**

Do not manually change `purge_after` to shorten the observation. Confirm hourly purger runs, new files use default 7 days, and pressure cleanup only targets verified rows.

- [ ] **Step 10: Final authenticated browser acceptance**

Verify:

```text
https://yzcm.site/settings
https://yzcm.site/batch-factory
comic thumbnail and original content routes
asset image content routes
comic collection download
merged video playback/range requests
new comic image generation
new merged video generation
```

Expected: all succeed after at least one local file has been cleaned and served from TOS.

- [ ] **Step 11: Record release evidence**

Record build ID, Git HEAD plus dirty-worktree archive identity, binary SHA256, runtime build info, test results, migration counts, TOS verification counts, disk before/after and rollback target. Do not claim completion until TOS fallback is visibly proven after local deletion.
