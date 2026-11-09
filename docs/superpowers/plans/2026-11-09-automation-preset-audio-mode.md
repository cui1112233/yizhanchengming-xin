# Automation Preset Audio Mode Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 修复自动化预设 MySQL 字段歧义，并让批量统一配置和自动化预设只使用按风格自动匹配音频。

**Architecture:** 后端只修正预设列表查询的字段来源，不改表结构；前端增加一个批量配置归一化函数，在显示、保存和载入预设时统一清理手动音频字段。单本配置通过现有 `singleBook` 边界保留手动能力。

**Tech Stack:** Go、MySQL、React、Ant Design、Node test runner、Vite

## Global Constraints

- 不新建分支。
- 不删除或重建当前 15 本小说。
- 批量统一配置与自动化预设固定使用“按风格自动匹配”。
- 单本配置继续允许手动指定音频。
- 先完成本地测试和构建，不直接发布线上。

---

### Task 1: 修复自动化预设列表 SQL

**Files:**
- Modify: `backend/internal/httpapi/automation_presets_mysql.go`
- Test: `backend/internal/httpapi/automation_presets_mysql_test.go`

**Interfaces:**
- Consumes: `MySQLAutomationPresetStore.List(owner string)`
- Produces: 可在 MySQL 联表中无歧义执行的预设列表查询

- [ ] **Step 1: 写失败测试**

在 `automation_presets_mysql_test.go` 中让 SQL mock 只接受以下带别名的查询：

```go
mock.ExpectQuery(`SELECT p\.id,p\.name,p\.current_revision,r\.config_json,p\.created_at,p\.updated_at FROM batch_factory_v12_automation_presets p JOIN batch_factory_v12_automation_preset_revisions r`).
    WithArgs("alice")
```

- [ ] **Step 2: 运行测试确认失败**

Run: `go test ./internal/httpapi -run TestMySQLAutomationPresetStoreList -count=1`

Expected: FAIL，因为当前查询使用未限定的 `id,name,...`。

- [ ] **Step 3: 最小修复查询**

把 `List` 查询改为：

```sql
SELECT p.id,p.name,p.current_revision,r.config_json,p.created_at,p.updated_at
FROM batch_factory_v12_automation_presets p
JOIN batch_factory_v12_automation_preset_revisions r
  ON r.preset_id=p.id AND r.revision=p.current_revision
WHERE p.owner_username=?
```

- [ ] **Step 4: 运行专项测试**

Run: `go test ./internal/httpapi -run TestMySQLAutomationPresetStoreList -count=1`

Expected: PASS

### Task 2: 批量配置固定自动音频模式

**Files:**
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryUnifiedSettingsModal.jsx`
- Test: `frontend/src/user/pages/shuihuo/batchFactoryAudioCategory.test.js`

**Interfaces:**
- Consumes: `comicPublishSettings` 与 `singleBook`
- Produces: `normalizeBatchAutoAudioConfig(config)`，返回清理过手动音频字段的配置副本

- [ ] **Step 1: 写失败测试**

测试断言：批量 UI 不再包含“手动指定”选项；自动化预设保存和载入调用归一化函数；单本 `指定音频` 代码仍存在。

```js
assert.doesNotMatch(batchAudioSection, /label: '手动指定'/)
assert.match(source, /normalizeBatchAutoAudioConfig\(value\)/)
assert.match(source, /singleBook \? <div className="batch-factory-engine-field"><span><b>指定音频<\/b>/)
```

- [ ] **Step 2: 运行测试确认失败**

Run: `node --test src/user/pages/shuihuo/batchFactoryAudioCategory.test.js`

Expected: FAIL，因为批量 UI 当前仍显示手动选项，预设也保存原始配置。

- [ ] **Step 3: 实现批量归一化**

新增并使用以下行为等价的函数：

```js
export function normalizeBatchAutoAudioConfig(config = {}) {
  const next = clone(config || {});
  const publish = next.comicPublishSettings || {};
  next.comicPublishSettings = {
    ...publish,
    audioCategorySelectionMode: 'auto',
    audioCategoryId: null,
    audioCategoryName: '',
    audioCandidates: [],
    specifiedAudioId: '',
    specifiedAudioName: ''
  };
  return next;
}
```

批量显示只渲染“按风格自动匹配”；单本继续渲染现有自动/手动切换、分类和指定音频。

- [ ] **Step 4: 在预设边界应用归一化**

保存新预设时使用 `normalizeBatchAutoAudioConfig(value)`；载入预设时使用 `normalizeBatchAutoAudioConfig(selectedPreset.config)`；统一配置保存前对批量草稿做相同处理。

- [ ] **Step 5: 运行专项测试**

Run: `node --test src/user/pages/shuihuo/batchFactoryAudioCategory.test.js src/user/pages/shuihuo/BatchFactoryUnifiedSettingsModal.source.test.js`

Expected: PASS

### Task 3: 本地完整验证

**Files:**
- Verify: `backend/internal/httpapi`
- Verify: `frontend`

**Interfaces:**
- Consumes: Task 1 和 Task 2 的实现
- Produces: 可本地构建的嵌入式前端与后端源代码

- [ ] **Step 1: 后端回归测试**

Run: `go test ./internal/httpapi ./internal/batchfactoryv11`

Expected: PASS

- [ ] **Step 2: 前端专项与完整测试**

Run: `npm test -- --runInBand`

Expected: PASS，或只存在与本次修改无关且已有记录的失败。

- [ ] **Step 3: 前端构建**

Run: `npm run build`

Expected: PASS，并生成可供 Go embed 使用的前端资源。

- [ ] **Step 4: 检查变更范围**

Run: `git diff --check` 和 `git status --short`

Expected: 无空白错误；只包含本方案文件和必要的构建产物。

- [ ] **Step 5: 提交本地修复**

```bash
git add backend/internal/httpapi/automation_presets_mysql.go backend/internal/httpapi/automation_presets_mysql_test.go frontend/src/user/pages/shuihuo/BatchFactoryUnifiedSettingsModal.jsx frontend/src/user/pages/shuihuo/batchFactoryAudioCategory.test.js docs/superpowers/specs/2026-11-09-automation-preset-audio-mode-design.md docs/superpowers/plans/2026-11-09-automation-preset-audio-mode.md
git commit -m "fix(batch-factory): repair preset audio defaults"
```
