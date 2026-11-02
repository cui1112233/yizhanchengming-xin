# 小说获取提交预检与失败日志实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在小说获取提交到视频管理系统前明确阻止未识别风格，并让本地或外部提交失败可在提交日志中追溯。

**Architecture:** Go 提交预览接口返回结构化的不可提交项；Go 提交接口对每个版本写入 `site_submit` 日志。前端复用现有预览和提交日志呈现，不改变 121 登录或上传协议。

**Tech Stack:** Go、`net/http`、现有小说获取工作台存储、现有批量改文前端。

## Global Constraints

- 正式源码只改 `v88`。
- 不改变 121 登录、Cookie 管理或外部上传协议。
- 不自动上传测试任务。
- 风格仍由既有分类流程或用户配置获得，不自动猜测替换。
- 不提交、删除或修改现有无关任务和构建产物。

---

### Task 1: 增加提交预览的风格预检

**Files:**
- Modify: `backend/internal/httpapi/web_submit_preview_compat.go:52-98`
- Modify: `backend/internal/httpapi/web_submit_submit_compat_test.go`

**Interfaces:**
- Consumes: `submissionGenderStyle(document)`、`styleNeedsResolution(meta)`。
- Produces: preview `skipped` 项 `{id, version, status: "blocked", error: "该任务尚未识别或选择上传风格，无法提交到视频管理系统。"}`。

- [ ] **Step 1: 写失败测试**

在 `web_submit_submit_compat_test.go` 增加预览测试，存入一个 `style: "待判定"` 且拥有 `ai1` 正文的任务，调用预览接口并断言：

```go
assert.Equal(t, "blocked", skipped[0]["status"])
assert.Equal(t, "该任务尚未识别或选择上传风格，无法提交到视频管理系统。", skipped[0]["error"])
```

- [ ] **Step 2: 运行测试确认失败**

Run: `go test ./internal/httpapi -run TestWebSubmitPreviewBlocksUnresolvedStyle -count=1`

Expected: FAIL，因为当前预览把“待判定”风格标为可提交分组。

- [ ] **Step 3: 实现最小预检**

在 `RegisterWebSubmitPreviewCompat` 中，在读取正文后、创建分组前加入：

```go
if styleNeedsResolution(meta) {
    skipped = append(skipped, map[string]any{
        "id": id, "version": version, "status": "blocked",
        "error": "该任务尚未识别或选择上传风格，无法提交到视频管理系统。",
    })
    continue
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `go test ./internal/httpapi -run TestWebSubmitPreviewBlocksUnresolvedStyle -count=1`

Expected: PASS。

### Task 2: 提交失败写入可见日志

**Files:**
- Modify: `backend/internal/httpapi/web_submit_submit_compat.go:109-185`
- Modify: `backend/internal/httpapi/web_submit_submit_compat_test.go`

**Interfaces:**
- Consumes: `markSubmissionFailure(meta, version, message)`。
- Produces: 每个失败版本写入 `Document.Logs` 的 `site_submit` 事件，包含 `version`、`status`、`error`、`time`、`externalRequested`。

- [ ] **Step 1: 写失败测试**

创建拥有 AI1 正文、性别有效而风格为“待判定”的任务，调用提交接口后断言：

```go
assert.Equal(t, "ai1", anyStrings(doc.Meta["siteSubmitFailedVersions"])[0])
assert.Equal(t, "site_submit", log["event"])
assert.Equal(t, "failed", log["status"])
assert.Equal(t, false, log["externalRequested"])
assert.Equal(t, "该任务尚未识别或选择上传风格，无法提交到视频管理系统。", log["error"])
```

- [ ] **Step 2: 运行测试确认失败**

Run: `go test ./internal/httpapi -run TestWebSubmitPersistsUnresolvedStyleFailureLog -count=1`

Expected: FAIL，因为当前只写失败元数据，不追加日志。

- [ ] **Step 3: 实现统一失败记录**

将提交处理中的 `failedVersion` 改为同时：

```go
markSubmissionFailure(meta, version, message)
doc.Logs = append(doc.Logs, map[string]any{
    "event": "site_submit", "version": version, "status": "failed",
    "error": message, "time": time.Now().UTC().Format(time.RFC3339Nano),
    "externalRequested": externalRequested,
})
```

本地失败传 `false`；`http.Client.Do` 已执行后发生网络或响应失败传 `true`；上传请求创建失败传 `false`。

- [ ] **Step 4: 将待判定风格错误统一化**

在调用 `buildUploadFieldsCompat` 前检查 `styleNeedsResolution(meta)`，走统一失败记录，并使用固定文案：

```go
"该任务尚未识别或选择上传风格，无法提交到视频管理系统。"
```

- [ ] **Step 5: 运行定向测试确认通过**

Run: `go test ./internal/httpapi -run 'TestWebSubmit.*(PreviewBlocksUnresolvedStyle|PersistsUnresolvedStyleFailureLog)' -count=1`

Expected: PASS。

### Task 3: 回归验证和提交

**Files:**
- Modify: `backend/internal/httpapi/web_submit_preview_compat.go`
- Modify: `backend/internal/httpapi/web_submit_submit_compat.go`
- Modify: `backend/internal/httpapi/web_submit_submit_compat_test.go`

- [ ] **Step 1: 运行提交相关完整测试**

Run: `go test ./internal/httpapi -run 'TestWebSubmit|TestNovelFetch' -count=1`

Expected: PASS。

- [ ] **Step 2: 运行 httpapi 全量测试**

Run: `go test ./internal/httpapi -count=1`

Expected: PASS。

- [ ] **Step 3: 审核差异范围**

Run: `git diff -- backend/internal/httpapi/web_submit_preview_compat.go backend/internal/httpapi/web_submit_submit_compat.go backend/internal/httpapi/web_submit_submit_compat_test.go`

Expected: 只含预检、日志记录和测试。

- [ ] **Step 4: 只暂存本次文件并提交**

Run:

```powershell
git add backend/internal/httpapi/web_submit_preview_compat.go backend/internal/httpapi/web_submit_submit_compat.go backend/internal/httpapi/web_submit_submit_compat_test.go docs/superpowers/specs/2026-10-13-novel-fetch-submit-preflight-design.md docs/superpowers/plans/2026-10-13-novel-fetch-submit-preflight.md
git commit -m "fix: surface novel submit validation failures"
```

- [ ] **Step 5: 记录提交 SHA**

Run: `git status --short --branch; git log -1 --oneline`

Expected: 只提交本次文件；既有未跟踪文件保持未修改。
