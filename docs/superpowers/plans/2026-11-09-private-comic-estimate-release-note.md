# 私人定制漫画预计张数公告实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为本地 4001 增加 v44 产品更新公告，并让公告明确展示私人定制漫画的预计张数功能。

**Architecture:** 沿用现有 Go embed 公告目录和 `releaseId` 选择机制，不修改弹窗组件。新增一份与本地发布指针匹配的公告 JSON，重建嵌入式 Go 二进制后通过接口和真实浏览器验收。

**Tech Stack:** Go、JSON、React 现有产品更新弹窗、PowerShell 本地启动脚本

## Global Constraints

- 新公告编号固定为 `qiantie-202611090148`，版本号固定为 `44`。
- 公告只描述已经完成的私人定制漫画预计张数功能。
- 保持“一版只弹一次”逻辑不变。
- 只更新本地 4001，不构建或部署线上版本。
- 不新建分支，不推送远端。

---

### Task 1: 增加 v44 公告契约

**Files:**
- Modify: `backend/internal/httpapi/product_updates_test.go`
- Create: `backend/internal/releasenotes/notes/qiantie-202611090148.json`
- Modify: `backend/current-release.json`

**Interfaces:**
- Consumes: `GET /api/product-updates/latest` 和现有 `releasenotes.Default()` 目录加载机制。
- Produces: `releaseId=qiantie-202611090148`、`version=44` 且含“预计张数”条目的公告响应。

- [ ] **Step 1: 写失败接口测试**

在 `product_updates_test.go` 增加测试：

```go
func TestProductUpdatesLatestReturnsV44PrivateComicEstimateNotes(t *testing.T) {
	handler := NewRouter(RouterOptions{RuntimeManifestPath: writeReleaseManifest(t, "qiantie-202611090148")})
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/api/product-updates/latest", nil))
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	var body struct {
		ReleaseID string `json:"releaseId"`
		Version string `json:"version"`
		Sections []struct { Items []struct { Title string `json:"title"` } `json:"items"` } `json:"sections"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil { t.Fatal(err) }
	if body.ReleaseID != "qiantie-202611090148" || body.Version != "44" { t.Fatalf("body = %#v", body) }
	if len(body.Sections) != 1 || len(body.Sections[0].Items) != 1 || body.Sections[0].Items[0].Title != "漫画生成前可查看预计张数" { t.Fatalf("sections = %#v", body.Sections) }
}
```

- [ ] **Step 2: 运行测试并确认因公告缺失失败**

Run: `go test ./internal/httpapi -run TestProductUpdatesLatestReturnsV44PrivateComicEstimateNotes -count=1`

Expected: FAIL，接口返回 404，因为 `qiantie-202611090148.json` 尚不存在。

- [ ] **Step 3: 新增最小公告并更新本地指针**

新增公告 JSON：

```json
{
  "releaseId": "qiantie-202611090148",
  "version": "44",
  "releasedAt": "2026-11-09",
  "title": "产品更新 · 2026-11-09",
  "summary": "私人定制漫画现在会在生成前实时显示预计张数。",
  "sections": [{
    "type": "feature",
    "label": "新增",
    "items": [{
      "title": "漫画生成前可查看预计张数",
      "desc": "粘贴正文后直接显示预计生成范围，例如 2000 字预计约 17～23 张；实际张数仍以生成结果为准。",
      "tags": ["私人定制", "漫画生成", "预计张数"]
    }]
  }]
}
```

同时把 `backend/current-release.json` 的 `releaseId` 改为 `qiantie-202611090148`，其它归档字段保持不变。

- [ ] **Step 4: 运行公告测试**

Run: `go test ./internal/releasenotes ./internal/httpapi -count=1`

Expected: PASS。

- [ ] **Step 5: 提交公告改动**

```powershell
git add -- backend/internal/httpapi/product_updates_test.go backend/internal/releasenotes/notes/qiantie-202611090148.json backend/current-release.json
git commit -m "feat: announce private comic page estimate"
```

### Task 2: 重建并验收本地 4001

**Files:**
- Build output: `backend/qiantie-local.exe`
- Runtime asset directory: `backend/web/dist`

**Interfaces:**
- Consumes: Task 1 的 v44 公告和现有 `/script` 预计张数实现。
- Produces: 本地 4001 返回 v44 公告，并在浏览器显示公告和预计张数。

- [ ] **Step 1: 运行前端相关测试和构建**

```powershell
cd frontend
node --test src/user/pages/PrivateComicWorkspace.source.test.js src/user/pages/shuihuo/batchFactoryContentRange.test.js
npm run build
```

Expected: 23 项测试全部通过，Vite 构建退出码为 0。

- [ ] **Step 2: 核对并停止本地 4001 进程**

确认监听进程路径必须为 `F:\脚本测试\一战晟铭-v88\backend\qiantie-local.exe`，路径不符时停止操作；路径一致时只停止该 PID。

- [ ] **Step 3: 重建并启动本地二进制**

```powershell
.\scripts\build-go-local.ps1
.\scripts\start-local-go-review.ps1 -UseYcmMysql -SkipBuild -Detached
```

Expected: 4001 由新 `qiantie-local.exe` 监听。

- [ ] **Step 4: 验证接口和运行版本**

核对：

- `/health` 返回 200 和 `{"ok":true}`。
- `/api/runtime-build-info` 的 Git SHA 等于当前 HEAD。
- `/api/product-updates/latest` 返回 `releaseId=qiantie-202611090148`、`version=44`。
- `/script` 返回 200。

- [ ] **Step 5: 用内置浏览器验收可见结果**

刷新本地 `/script`，等待更新弹窗出现并确认标题为“产品更新 · 2026-11-09”；关闭弹窗后进入“漫画生成”，确认正文下方出现 `预计约`。

- [ ] **Step 6: 检查工作树**

Run: `git status --short`

Expected: 无输出；不执行任何线上发布命令。
