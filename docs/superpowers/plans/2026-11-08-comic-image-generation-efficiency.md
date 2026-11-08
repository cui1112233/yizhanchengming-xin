# Comic Image Generation Efficiency Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让漫画生图可定位慢点、可复用完全相同的成功结果，并避免已计费请求被自动重复提交。

**Architecture:** 在现有图片供应商调用与漫画页落库之间增加轻量生成账本；复用现有 localartifact 与 productionmedia 的本地/TOS读取，不另建文件存储。主机级自适应控制器只调整后续并发。

**Tech Stack:** Go 1.26, MySQL 8, React 18, Ant Design 5, Go embed

## Global Constraints

- 不记录密钥和完整提示词。
- 不自动重发 HTTP 200 后未取回正文的请求。
- 只有输入完全一致才复用；用户明确重新生成必须绕过。
- 不改现有 TOS 迁移、校验和 7 天清理规则。
- 保留脏工作树中所有无关改动，只修改本计划列出的文件。

### Task 1: Add Detailed Attempt Timings

**Files:**
- Modify: `backend/internal/httpapi/shuihuo_projects.go`
- Modify: `backend/internal/httpapi/shuihuo_generated_image_download_test.go`
- Modify: `backend/internal/httpapi/shuihuo_image_attempt_guard_test.go`

- [x] 写失败测试，断言结构化日志包含 queue/header/body/download/save/total。
- [x] 在请求函数逐段记录正文读取、结果解码/下载与落盘耗时。
- [x] 保持 HTTP 200 后断流不在请求层自动重发。
- [x] 运行：`go test ./internal/httpapi -run 'ImageAttempt|GeneratedImageDownload' -count=1`。

### Task 2: Add Durable Generation Cache Schema and Repository

**Files:**
- Modify: `backend/internal/storage/batch_factory_v12_schema.go`
- Modify: `backend/internal/storage/batch_factory_v12_schema_test.go`
- Create: `backend/internal/imagegenerationcache/types.go`
- Create: `backend/internal/imagegenerationcache/mysql_repository.go`
- Create: `backend/internal/imagegenerationcache/mysql_repository_test.go`

- [x] 写失败迁移测试，要求新增缓存表和唯一指纹索引。
- [x] 新增迁移 `1200107`，字段覆盖状态、租约、storage_ref、media_type 和分段耗时。
- [x] 实现 Claim/MarkReady/MarkFailed，Claim 使用租约防并发重复提交。
- [x] 运行：`go test ./internal/storage ./internal/imagegenerationcache -count=1`。

### Task 3: Integrate Exact Fingerprints and Cache Reuse

**Files:**
- Modify: `backend/internal/httpapi/shuihuo_projects.go`
- Modify: `backend/internal/httpapi/router.go`
- Modify: `backend/internal/app/app.go`
- Create: `backend/internal/httpapi/shuihuo_image_cache_test.go`

- [x] 写失败测试：相同输入第二次命中缓存，不触发第二次供应商请求。
- [x] 写失败测试：模型、提示词、画幅、参考图或水印任一变化都会换指纹。
- [x] 写失败测试：显式重新生成绕过缓存。
- [x] 注入缓存仓库；仓库不可用时安全退回旧路径。
- [x] 运行缓存、路由、应用、迁移的聚焦测试。

### Task 4: Add Host-Level Adaptive Concurrency

**Files:**
- Modify: `backend/internal/httpapi/shuihuo_projects.go`
- Create: `backend/internal/httpapi/shuihuo_image_adaptive_concurrency_test.go`

- [x] 写失败测试：连续 reset/stall 后上限降到 2。
- [x] 写失败测试：连续成功后恢复到配置上限，且不超过 4。
- [x] 实现主机级控制器并接入 acquire/release 与结果反馈。
- [x] 运行：`go test ./internal/httpapi -run 'AdaptiveImageConcurrency' -count=1`。

### Task 5: Verify Retry Semantics

**Files:**
- Modify: `backend/internal/httpapi/batch_factory_automation_retry_test.go`
- Modify: `backend/internal/httpapi/automation_failure_test.go`

- [x] 复用现有回归测试确认：已有成功图片的页不会因整本重试再次生成。
- [x] 复用现有回归测试确认：已计费未取回错误保留为独立失败类型。
- [x] 运行：`go test ./internal/httpapi -run 'Retry|BilledReadFailure' -count=1`。

### Task 6: Local Verification and Controlled Release

**Files:**
- Build: `frontend/dist`
- Sync: `backend/web/dist`
- Build: `.release/qiantie-YYYYMMDDHHmm`

- [x] 运行聚焦测试，并额外完成 `go test ./... -count=1 -timeout 300s`。
- [x] 启动脚本完成现有前端资源同步与 Go 本地二进制重建。
- [x] 启动 4001，验证 `/health`、批量工厂页面和迁移 `1200107`。
- [ ] 取得用户对一次受控付费 A/B 的确认后，只生成一张测试图并对账分段耗时。
- [ ] A/B 通过后构建单个 Linux 二进制，保留旧版本，哈希校验后发布并验证线上。
