# 水火生产 Go 路由修复实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 修复 `http://localhost:4001/shuihuo-production` 使用批量工厂时被错误转发并返回“该功能尚未迁移到 Go 服务”的问题。

**Architecture:** 保留已经存在的 Go 批量工厂和水火生产接口，只调整 4001 Go 入口的请求分流规则，让水火生产实际使用的 V12 接口优先进入 Go；对 Go 尚未实现的接口继续保留兼容转发，不伪造成功。通过入口路由测试和 Go 单元测试确认分流结果。

**Tech Stack:** Go、`net/http`、现有 Go API、现有 React 水火生产页面。

## Global Constraints

- 正式源码只改 `v88` 主线。
- 不改运行中的服务器或 4001 运行目录作为最终修复。
- 不删除历史分支、旧 Workflow 或现有 Docker 基础设施。
- 只提交本次涉及的源码和测试文件，不提交已有未跟踪构建产物。
- 用户可见的 `121` 继续显示为“视频管理系统”。

---

### Task 1: 固定当前路由问题

**Files:**
- Modify: `backend/cmd/qiantie/main_test.go`
- Modify: `backend/cmd/qiantie/main.go`

**Interfaces:**
- Consumes: `shouldBridgeLegacyAPIRequest(*http.Request, int)` 和 `shouldBridgeLegacyAPIForSlice(string, int)`。
- Produces: 水火生产页面所需的 V12 路由分流规则。

- [ ] **Step 1: 写失败测试**

在 `main_test.go` 增加测试，明确验证水火生产批量工厂常用接口在当前 Go slice 已启用时不应走旧 Node 兼容桥，包括：

```go
func TestShuihuoProductionBatchFactoryPathsStayOnGo(t *testing.T) {
    paths := []string{
        "/api/batch-factory/v12/batches/demo",
        "/api/batch-factory/v12/batches/demo/books/book-1/assets",
        "/api/batch-factory/v12/batches/demo/books/book-1/director",
        "/api/batch-factory/v12/batches/demo/books/book-1/production",
        "/api/batch-factory/v12/batches/demo/runtime-summary",
    }
    for _, path := range paths {
        if shouldBridgeLegacyAPIForSlice(path, 3) {
            t.Fatalf("水火生产路径不应转到旧服务: %s", path)
        }
    }
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `go test ./cmd/qiantie`

Expected: 新增路由断言失败，说明至少一个水火生产接口仍被分流到旧服务。

- [ ] **Step 3: 找出最小路由修复**

检查 `shouldBridgeLegacyAPIForSlice` 与 `isGoV12AliasPath` 的路径拆分规则，补齐当前 Go 已注册的水火生产批量工厂接口；不要把尚未注册的接口一并标成 Go 已实现。

- [ ] **Step 4: 运行入口测试确认通过**

Run: `go test ./cmd/qiantie`

Expected: PASS。

- [ ] **Step 5: 检查本次变更范围**

Run: `git diff -- backend/cmd/qiantie/main.go backend/cmd/qiantie/main_test.go`

Expected: 只包含 V12 路由判断和对应测试，不出现构建产物或运行目录文件。

---

### Task 2: 验证 Go 接口和页面调用链

**Files:**
- Test: `backend/internal/httpapi`
- Test: `frontend/src/shared/api/batchFactoryV11.js`
- Test: `frontend/src/user/pages/ShuihuoProductionPage.jsx`

**Interfaces:**
- Consumes: Task 1 修复后的 Go 路由入口。
- Produces: 页面访问批量工厂时不再收到“尚未迁移到 Go 服务”的错误。

- [ ] **Step 1: 运行 Go 批量工厂相关测试**

Run: `go test ./internal/httpapi ./internal/batchfactoryv11/...`

Expected: PASS；若失败，按实际缺失接口修复对应 Go 路由，不修改前端提示来掩盖接口错误。

- [ ] **Step 2: 运行前端已有批量工厂测试**

Run: `npm test -- --runInBand`

Working directory: `frontend`

Expected: 相关批量工厂和水火生产测试 PASS；不得把已有无关失败当成此次修复成功。

- [ ] **Step 3: 构建前端**

Run: `npm run build`

Working directory: `frontend`

Expected: 构建成功。构建生成物只用于验证，不纳入提交，除非仓库明确跟踪该目录。

- [ ] **Step 4: 最后确认 Git 工作区**

Run: `git status --short`

Expected: 只看到本次源码、测试和计划文件；已有无关未提交文件不得被修改或加入提交。

---

### Task 3: 提交可追溯修复

**Files:**
- Commit only the exact files changed by Tasks 1-2.

- [ ] **Step 1: 查看差异和最近提交格式**

Run: `git diff --stat; git diff -- backend/cmd/qiantie/main.go backend/cmd/qiantie/main_test.go; git log -5 --oneline`

- [ ] **Step 2: 只暂存本次文件**

Run: `git add backend/cmd/qiantie/main.go backend/cmd/qiantie/main_test.go docs/superpowers/plans/2026-10-13-shuihuo-production-go-routing.md`

- [ ] **Step 3: 提交修复**

Run: `git commit -m "fix: route Shuihuo production batch factory through Go"`

- [ ] **Step 4: 记录提交 SHA 并确认干净范围**

Run: `git status --short --branch; git log -1 --oneline`

Expected: 提交成功，提交 SHA 可追溯；其他已有未提交文件仍保持原状。
