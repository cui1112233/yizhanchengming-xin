# 漫剧内容与画风 Implementation Plan

> Execute task by task with superpowers:subagent-driven-development. User approved design 2026-11-13-comic-screenplay-input-style-design.md. Work directly on v88; preserve unrelated changes. No paid generation, Git push or public deployment.

**Goal:** 新建漫剧支持直接输入及预览导入，解析档位明显可辨，使用后端画风库并传入真实资产和分镜链路。

**Architecture:** 后端提供项目无关的文件解析及可用画风列表，创建时一并写入原文，按画风预设标识解析生成描述。前端独立输入草稿、可见选择卡和图片画风库，沿用作品保存与生成服务。

**Tech Stack:** Go/MySQL、React/Antd、现有预设模块、Go embed。

## Global Constraints

- 不调用付费模型，不复制洛水专有素材和提示词；前端不硬编码生成提示词。
- 不建分支，不覆盖、清理或提交无关工作区改动。
- 先本地 4001 真实页面验收。本次不发布公网。

### Task 1: 后端输入与画风闭环

Files: backend/internal/httpapi/*screenplay*.go; backend/internal/screenplay/{mysql_store,memory_store,assets_service,storyboard_service}.go; focused tests.

- [ ] 先写失败测试：无项目解析 TXT/DOCX；创建即持久化 sourceText；画风列表使用现有预设；选中标识解析后端描述并在资产和分镜请求中出现。
- [ ] 实现 GET /api/comic-screenplays/styles，返回 styles[{id,name,previewUrl}]；POST /api/comic-screenplays/source/parse，返回既有解析格式。
- [ ] 创建/更新 selectedStyle 为预设 ID，styleDesc 由后端解析，保留历史名称兼容；确保生成请求包含选中画风。
- [ ] 相关后端测试通过，提供前端接口契约和字段。

### Task 2: 新建与设置界面

Files: frontend/src/shared/api/comicScreenplay.js; frontend/src/user/pages/ComicScreenplayPage.jsx; comic-screenplay.css; helpers and tests.

- [ ] 先写行为测试：文本校验、导入前不创建、草稿与现有作品原文隔离、选择画风持久化字段。
- [ ] 直接输入/上传切换，默认直接输入；上传后可编辑、字数提示、失败不建空项目，防止旧解析结果覆盖新草稿。
- [ ] 四档有名称、真实范围说明、勾选及明显选中边框，支持键盘、深浅主题。
- [ ] 画风库真实图片、当前选择、清除、搜索、空/失败反馈；项目设置也能更新画风。
- [ ] 相关测试及前端生产构建通过。

### Task 3: 审查与本地交付

- [ ] 独立审查边界及持久化和生成链路，修复发现。
- [ ] 只提交本次源码；冻结当前基线与构建资源，生成 Go embed 本地程序。
- [ ] 确认本地任务空闲后启动并验证版本，真实页面检查新建窗口、选择和持久化；不用付费生成。
- [ ] 保存截图及验收记录，向用户展示可见结果及公网状态。
