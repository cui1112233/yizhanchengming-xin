# 漫画手机工作区实施计划

> **For agentic workers:** Use superpowers:executing-plans to implement task-by-task in the existing v88 checkout as requested.

**Goal:** 落地已确认的正文、人物场景、漫画页图三屏及其工具面板。

**Architecture:** 保留 ComicPrivatePage 的配置快照、创建、轮询与恢复逻辑；新编辑器组件承接视图和局部工具状态；资产/页图编辑使用网页现有 V12 接口并回读。

**Tech Stack:** Taro React、TypeScript、SCSS、Vitest。

## Constraints

390px 基准，360–430px 兼容；深灰底、青色操作、无背景图；真实字段与权限；不自动发起生成或部署。直接在 v88 工作，保留无关改动。

### 1. 服务操作

- [x] 新建 comicEditing.ts：资产 PATCH 带 expectedRevision，图片上传 dataUrl，主图 PUT，页提示词 PUT；各操作刷新读回。
- [x] 添加服务请求回归：校验地址、book/asset 范围、revision、双提示词校验与图片来源。

### 2. 统一工作区

- [x] 在 ComicSettingsStage、ComicAssetsStage、ComicPageWorkspace 和 editor.scss 中落地三屏；主页面传入真实状态和既有回调。
- [x] 正文使用专属画风、画幅面板及既有模型/提示词选择组件；三项内容切换保持同一项目。
- [x] 人物场景列表点选并独立编辑/生图/选版本；页图大图使用 aspectFit，缩略页条可切页，工具独立展开。
- [x] 保留配置确认与自动运行范围提示、进行中/失败/空态；禁用理由可读。

### 3. 验证与交付

- [x] 跑相关服务/页面测试与类型检查，修复行为回归而非删除用例。
- [x] 跑完整测试和 weapp 构建、WXSS 兼容检查。
- [x] 在微信开发者工具核对主屏与工具面板；无可验收任务时不伪造结果。
- [x] 记录源码修改与实际验证范围，本地提交仅本轮明确文件。

验证边界及类型检查的范围外错误详见 docs/design/2026-11-24-comic-mobile-editor-qa.md。真实生成后的资产/页图画面尚待有效项目验收。
