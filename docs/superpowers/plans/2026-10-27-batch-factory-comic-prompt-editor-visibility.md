# Batch Factory Comic Prompt Editor Visibility Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让漫画页的两段页面描述都可编辑，并把最终完整出图提示词严格限制为开发者可见。

**Architecture:** 复用现有账号会话和 `role === 'dev'` 规则。前端负责按角色决定是否渲染、是否请求预览；后端负责最终权限校验。漫画页保存仍走现有视频覆盖接口，但一次提交两个漫画字段。

**Tech Stack:** React 18、Ant Design、Go `net/http`、现有 Batch Factory V12 API、Node test、Go test。

## Global Constraints

- 不新建 Git 分支，不提交，不清理工作区已有修改。
- 系统提示词继续由 Go 后端拼装，前端不硬编码完整系统提示词。
- 开发者身份只按 `role === 'dev'` 判断。

---

### Task 1: 锁定前端漫画编辑行为

**Files:**
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryComicFullPrompt.source.test.js`
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryNovelList.jsx`

**Interfaces:**
- Consumes: `getCurrentAccount(): Promise<Account>`、`saveVideoOverride(..., { patch, expectedRevision })`
- Produces: `PromptPanel({ isDeveloper })`，漫画保存补丁包含 `comicPrompt` 与 `comicImagePrompt`

- [ ] 写失败测试，要求漫画编辑区存在两个独立字段、保存两个字段、普通用户不加载开发者预览。
- [ ] 运行测试，确认当前源码因单字段编辑和无权限分支而失败。
- [ ] 增加独立的 `comicDirectorPrompt` 状态，编辑、取消、切页和保存都同步处理两个字段。
- [ ] 在父组件读取账号角色，将 `isDeveloper` 传入提示词面板，并仅对开发者请求和渲染完整预览。
- [ ] 重跑前端定向测试。

### Task 2: 保护后端完整提示词接口

**Files:**
- Modify: `backend/internal/httpapi/batch_factory_comic_pages_test.go`
- Modify: `backend/internal/httpapi/batch_factory_comic_pages.go`
- Modify: `backend/internal/httpapi/router.go`

**Interfaces:**
- Consumes: `auth.Store.FindAccount`、`MemberProfileStore.member`
- Produces: `POST .../prompt-preview` 对开发者返回 `200`，对其他账号返回 `403`

- [ ] 写失败测试，分别以普通账号和开发者账号调用预览接口。
- [ ] 运行测试，确认普通账号当前错误地得到 `200`。
- [ ] 把账号仓库和成员资料仓库传入漫画路由，仅允许角色为 `dev` 的账号读取完整预览。
- [ ] 重跑后端定向测试。

### Task 3: 构建与运行验收

**Files:**
- Verify: `frontend/dist`
- Verify: `backend/web/dist`
- Verify: `backend/qiantie-local.exe`

**Interfaces:**
- Consumes: 前两项实现
- Produces: `http://localhost:4001/batch-factory` 的可见运行结果

- [ ] 运行前端定向测试和相关 Go 测试。
- [ ] 运行前端构建，再运行 `scripts/build-go-local.ps1` 生成嵌入式本地程序。
- [ ] 重启本地服务并检查健康接口与运行版本。
- [ ] 在当前登录开发者账号中打开漫画提示词，确认两个编辑框与开发者完整预览均符合设计。
