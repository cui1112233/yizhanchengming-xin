# Go embed 前端：单二进制发布规范

## 结论

本项目的可维护前端源码在 `frontend/src`。发布前先构建 React，再把 `frontend/dist` 复制到 `backend/web/dist`，最后由 `backend/web/assets.go` 使用 Go `embed.FS` 编进 `backend/cmd/qiantie` 生成的 Go 二进制。

因此，Go 服务可以用一个二进制同时提供：

- `/api/*` 和 `/health` 等 Go 接口；
- React 用户端、管理端页面和静态资源；
- SPA 未命中页面的 `index.html` 回退。

`frontend/dist` 和 `backend/web/dist` 都是构建产物，不是源码，禁止把它们反向复制回 `frontend/src`。

## 当前代码布局

```text
frontend/src/             React + Ant Design 源码
frontend/dist/            Vite 构建结果
backend/web/assets.go     //go:embed dist/*，导出嵌入文件系统
backend/web/dist/         构建时复制进 Go 包的静态文件
backend/cmd/qiantie/      Go 主服务入口
scripts/build-go-local.ps1  前端产物同步与 Go 编译
```

当前不是 Gin 双静态目录模式，也不需要为用户端和管理端各维护一个独立的 `embed.FS`。管理端通过 `/admin` 路由选择 `admin.html`，用户端和其他 React 页面使用 `index.html`，具体回退逻辑在 `backend/internal/web/static.go`。

## Go embed 实现要点

```go
// backend/web/assets.go
//go:embed dist/*
var Dist embed.FS
```

`Handler()` 先通过 `fs.Sub(Dist, "dist")` 去掉顶层目录，再交给静态路由处理。API 路由先于静态兜底注册：

1. `/health` 和 `/api/*` 进入 Go API；
2. `/admin` 和 `/admin/*` 读取 `admin.html`；
3. 其他页面优先读取真实静态文件；
4. 没有对应文件的前端路径回退到 `index.html`，交给 React 路由处理；
5. `/batch-rewrite/v2` 只显式保留旧 V78 对照页；
6. `/novel-panel/workbench` 明确拒绝，避免旧编译工作台重新成为默认入口。

这样可以避免 API 请求被 SPA 页面吞掉，也避免浏览器刷新 React 子路由得到 404。

## 构建顺序（不可颠倒）

必须先生成前端 `dist`，再同步到 Go embed 目录，最后编译 Go。项目脚本已经固化这条顺序：

```powershell
Set-Location F:\脚本测试\一战晟铭-v88
npm --prefix frontend install
npm --prefix frontend run build
.\scripts\build-go-local.ps1
```

`build-go-local.ps1` 会完成以下工作：

1. 确认 `frontend/dist/index.html` 存在；
2. 清理并复制 `frontend/dist` 到 `backend/web/dist`；
3. 复制需要保留的品牌和宠物静态资源；
4. 编译 `backend/cmd/qiantie`；
5. 按需编译本地执行器并记录构建元数据。

只同步前端、不重新编译 Go 时：

```powershell
.\scripts\build-go-local.ps1 -SkipGoBuild
```

不要先执行 `go build` 再构建前端，否则二进制会嵌入旧页面或空目录。`embed` 目录不存在或为空时，构建也会失败。

## 启动与验证

```powershell
.\scripts\start-local-go-review.ps1 -UseYcmMysql
```

启动脚本读取 `.env.v88-review`，连接本地审核用 MySQL/Redis，并启动 Go 二进制。确认运行时没有误启 Node：

```powershell
.\scripts\verify-no-node-runtime.ps1
```

验证不能只看构建是否成功。至少检查：

- `/health` 返回正常；
- `/novel-fetch`、`/batch-factory`、`/shuihuo-production`、`/admin/presets` 都返回正确 React 页面；
- `/api/*` 请求确实进入 Go API；
- 刷新 React 子路由不会丢页面；
- 任务、数据库、Redis、对象存储和外部接口的实际链路可用。

## 发布边界

单二进制是应用发布包，但不代表所有基础设施都被打进二进制。MySQL、Redis、FFmpeg、TOS 和外部视频管理系统仍是运行依赖，必须按环境提供。

本地 Go review 运行可以由一个 Go 二进制直接启动。线上是否已经完全移除 Nginx、Docker 或其他混合运行时，必须以目标环境的实际部署记录和公网验证为准，不能仅凭本地代码推断。

每次发布必须记录：源码 Git SHA、构建时间、目标环境、二进制校验值、服务状态和实际页面/API 验证结果。未经用户明确指令，不向远端 Git 仓库推送。
