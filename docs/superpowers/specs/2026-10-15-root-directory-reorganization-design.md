# 根目录业务分区整理设计

日期：2026-10-15
分支：v88

## 目标

让根目录只保留少量入口和配置，按业务职责组织代码，避免 Go 接口、React 页面、外部 Worker、执行器、部署文件和运行数据混在一起。

## 目标结构

```text
backend/       Go 接口、业务逻辑、Go embed 静态发布
frontend/      React 前台、管理后台、兼容静态页面和前端资源
services/      外部服务 Worker（保留视频管理系统浏览器 Worker）
executors/     执行器实现（保留现有 Go 执行器；旧 Node 执行器单独标记）
deploy/        发布与部署配置
scripts/       本地启动、构建、检查脚本
docs/          设计、迁移、审计文档
tests/         测试
data/          本地运行数据
```

## 迁移范围

1. 删除根目录 `node_modules`，并移除根 `package.json` / `package-lock.json` 中已经没有调用方的旧 Node 后端依赖；前端依赖仍由 `frontend/package.json` 管理。
2. 将 `local-executor` 和 `giant-material-executor` 归入 `executors/`，同步更新构建脚本、测试、部署清单和文档引用。
3. 将根 `public`、`pets`、`branding`、根 `dist` 中仍被 Go embed 或构建脚本使用的资源归入 `frontend/` 的明确静态源目录；删除已经没有调用方的旧根静态副本。
4. 保留 `backend/web/dist` 作为 Go embed 的构建产物目录，不把生成产物重新散落到根目录。
5. 保留 `services/121-browser-worker`，因为它仍是外部浏览器 Worker，不属于已删除的 Node 主服务。
6. 保留 `.github`、历史 Docker/发布审计资料、`data` 和 `.local-review` 运行数据；这些在当前整合阶段仍有审计或运行用途。
7. 保留 `AGENTS.md`、`CURRENT_VERSION.md`、`.env.v88-review.example` 等根级规则和示例配置；真实 `.env.v88-review` 只保留本地，不入库。

## 本轮执行边界

本轮将两个执行器实体移动到 `executors/`，并同步当前脚本、测试和文档入口。
根 `public/`、`pets/`、`branding/` 仍是 Go embed 构建脚本的兼容输入，先保持原路径，
由构建脚本统一生成 `backend/web/dist`；待静态资源来源完成独立迁移验证后再删除旧副本。

## 兼容原则

- 所有移动先更新引用，再删除旧路径。
- Go `embed` 的最终输入仍由 `backend/web/dist` 统一提供。
- `/batch-factory`、`/novel-fetch`、`/novel-panel` 等页面路径不改变。
- 不修改业务接口，不改变数据库/Redis/TOS 配置。
- 不删除当前运行数据和备份 tag。

## 验收

- `go test ./...`（`backend`）通过。
- Go 本地二进制可直接启动并访问根页面及主要业务页面。
- 前端构建通过，`backend/web/dist` 能重新生成。
- 构建脚本、部署脚本、测试中不再引用已删除的根 Node 后端路径。
- 根目录仅保留约定的入口、配置和业务目录。
