# 漫画出图跨任务限流与续跑修复实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 多个漫画自动化任务同时运行时，全服务图片请求仍受同一个并发上限约束；失败重试只补未完成页，并让当前批次能继续恢复。

**Architecture:** 保留现有“每本书并发”和“单任务阶段公平”两层控制，再增加一个进程级共享总闸门，所有 GPT 图片调用都必须先经过它。页面续跑继续以已落库主图为完成依据，不重新生成成功页；部署后只观察自动恢复，不主动触发新的付费重试。

**Tech Stack:** Go、MySQL 持久化任务、React 嵌入资源、systemd 单二进制部署。

## Global Constraints

- 所有源码修改先落在本地真实仓库，再构建本地运行并验证，最后才发布线上。
- 不清理、不覆盖工作树中其他人的大量未提交改动。
- 不主动点击重试或新增付费生图调用。
- 只发布 Go embed 生成的单个 Linux 二进制。

---

### Task 1: 进程级图片并发总闸门

**Files:**
- Modify: `backend/internal/httpapi/shuihuo_projects.go`
- Test: `backend/internal/httpapi/shuihuo_image_semaphore_test.go`

**Interfaces:**
- Consumes: `acquireShuihuoGPTImageGenerationSlot(context.Context, string)`
- Produces: 所有任务共享的图片请求总并发约束，并保留原有任务内阶段公平。

- [ ] **Step 1: 写失败测试**：创建两个互不相同的任务上下文，同时请求图片槽位，断言合计在用数不超过进程上限。
- [ ] **Step 2: 运行聚焦测试并确认失败**：失败原因必须是第二个任务绕过共享上限。
- [ ] **Step 3: 最小实现**：增加进程级共享信号量，领取顺序为进程总闸门、任务阶段闸门、单书闸门；任一领取失败时按逆序释放。
- [ ] **Step 4: 运行聚焦测试并确认通过**。

### Task 2: 续跑与旧状态验证

**Files:**
- Verify: `backend/internal/httpapi/batch_factory_comic_pages.go`
- Verify: `backend/internal/httpapi/batch_factory_automation.go`
- Test: `backend/internal/httpapi/batch_factory_automation_reuse_test.go`

**Interfaces:**
- Consumes: 已落库漫画页、主图及自动化书级状态。
- Produces: 已完成页跳过、在途页不重复提交、失败页可在下一轮续跑。

- [ ] **Step 1: 运行已有“只补未完成页”聚焦测试**。
- [ ] **Step 2: 若覆盖不足，先补失败测试，再做最小修复**。
- [ ] **Step 3: 确认当前批次无需人工再次点击也不会重跑成功页**。

### Task 3: 本地构建与线上发布

**Files:**
- Modify: `backend/current-release.json`
- Create: `backend/internal/releasenotes/notes/<release>.json`

**Interfaces:**
- Consumes: 通过测试的 Go 源码和现有前端嵌入资源。
- Produces: 本地 4001 可验证运行、Linux 单二进制和线上运行版本。

- [ ] **Step 1: 运行 Go 聚焦测试和后端构建**。
- [ ] **Step 2: 启动本地 4001，核验 `/health`、`/api/runtime-build-info` 与页面入口**。
- [ ] **Step 3: 构建 Linux 二进制并核对 SHA256**。
- [ ] **Step 4: 上传新文件、切换软链接、重启服务并核验线上版本和当前批次日志**。
