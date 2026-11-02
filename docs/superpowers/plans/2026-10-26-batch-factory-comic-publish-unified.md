# Batch Factory Comic Publish Unified Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在漫画模式的统一配置中提供简单、可保存的漫画发布默认设置。

**Architecture:** 复用现有统一配置草稿和批次设置保存接口，新建一个只负责漫画发布字段的 React 表单。以 `productionMode` 条件渲染视频或漫画发布表单，并把漫画值隔离保存在 `comicPublishSettings`。

**Tech Stack:** React, Ant Design, Node source tests, Vite, Go embed

## Global Constraints

- 仅修改 `v88` 当前源码，不覆盖或清理工作区中的其他改动。
- 本次不调用外部漫画上传接口。
- 密钥、签名和接口地址不进入前端。

---

### Task 1: 漫画发布统一表单

**Files:**
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryUnifiedSettingsModal.jsx`
- Test: `frontend/src/user/pages/shuihuo/BatchFactoryUnifiedSettingsModal.source.test.js`

**Interfaces:**
- Consumes: `draftPatch.productionMode` and `draftPatch.comicPublishSettings`.
- Produces: `BatchFactoryComicPublishSettingsForm({ value, onChange })` and a mode-specific publish tab.

- [ ] **Step 1: Write the failing source test**

Assert that comic mode renders `BatchFactoryComicPublishSettingsForm`, stores `producerAdminName`, `audioCategoryId`, and `workMatchMode` in `comicPublishSettings`, while video mode keeps `BatchFactoryPublishSettingsForm`.

- [ ] **Step 2: Run the focused test and confirm it fails**

Run: `node --test frontend/src/user/pages/shuihuo/BatchFactoryUnifiedSettingsModal.source.test.js`

- [ ] **Step 3: Implement the minimal form**

Add the three user-facing controls, automatic-processing explanation, and conditional rendering based on normalized production mode.

- [ ] **Step 4: Run focused frontend tests**

Run the unified-settings and production-mode source tests and confirm zero failures.

### Task 2: Build and live acceptance

**Files:**
- Generated: `frontend/dist/**`
- Generated: local Go embedded binary

**Interfaces:**
- Consumes: tested React source.
- Produces: the local `:4001` runtime showing the comic publish form.

- [ ] **Step 1: Build the frontend**

Run the repository frontend build command and confirm exit code 0.

- [ ] **Step 2: Rebuild and restart the local Go embedded runtime**

Use the repository's approved local Go script without deploying publicly.

- [ ] **Step 3: Verify in the browser**

Open the current batch, enter unified configuration, switch to comic mode, and confirm the three fields are visible and video publication controls are absent.
