# Product Update Release Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为 `qiantie-202611071037` 补齐 v27 更新公告，并在本地验收后发布到线上。

**Architecture:** 更新公告以 JSON 放入 Go embed 目录，由运行版本清单中的 `releaseId` 选择。发布脚本生成匹配的二进制和清单；线上通过软链接切换二进制，并同步替换清单。

**Tech Stack:** Go、React、PowerShell、systemd、nginx

## Global Constraints

- 必须先完成本地测试和本地候选二进制验收，再上传线上。
- 不重置、不清理、不覆盖工作区中的其他改动。
- 不提交、不推送、不创建分支。
- 保留当前线上二进制和清单作为回滚点。

---

### Task 1: 公告匹配门禁

**Files:**
- Modify: `backend/internal/httpapi/product_updates_test.go`
- Create: `backend/internal/releasenotes/notes/qiantie-202611071037.json`

**Interfaces:**
- Consumes: `GET /api/product-updates/latest`
- Produces: `releaseId=qiantie-202611071037`、`version=27` 的公告响应

- [ ] 在接口测试中新增 `qiantie-202611071037` 的断言。
- [ ] 运行测试并确认因公告缺失得到 404。
- [ ] 新增 v27 公告 JSON，内容限定为断点续跑、失败项补生成、进行中任务保护。
- [ ] 重跑 `go test ./internal/httpapi ./internal/releasenotes -count=1` 并确认通过。

### Task 2: 本地发布产物

**Files:**
- Modify: `backend/current-release.json`（由发布脚本生成）
- Create: `.release/v88/20261107-01-e8c86d6b/qiantie-server-linux-amd64`
- Create: `docs/releases/20261107-01-e8c86d6b/runtime-manifest.json`

**Interfaces:**
- Consumes: `scripts/release-v88.ps1 -ReleaseId qiantie-202611071037`
- Produces: 已嵌入公告的 Linux 二进制、运行版本清单和本地验收报告

- [ ] 运行发布脚本但不使用 `-Deploy`。
- [ ] 检查本地前置验收中的 health、公告 releaseId、前端入口和运行 SHA。
- [ ] 记录 Linux 文件大小与 SHA256。

### Task 3: 线上原子发布

**Files:**
- Create: `/opt/www/qiantie/qiantie-202611071037`
- Modify: `/opt/www/qiantie/runtime-release.json`
- Modify: `/opt/www/qiantie/qiantie` 软链接

**Interfaces:**
- Consumes: Task 2 的 Linux 二进制和 `runtime-manifest.json`
- Produces: 线上运行 `qiantie-202611071037` 并返回 v27 公告

- [ ] 上传二进制和清单到临时文件。
- [ ] 比对本地与远端 SHA256。
- [ ] 备份旧清单，原子替换清单和软链接，重启 `qiantie.service`。
- [ ] 验证进程实际文件、health、公告接口、批量工厂页面、警告日志和 systemd 重启次数。
- [ ] 任一核心验收失败时恢复旧软链接和旧清单并重启。
