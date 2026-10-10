# 本地 Go 运行入口

前端仍然用 npm 打包，正式运行入口改由 Go 提供。Go 会同时返回 API 和打包后的 React 页面。

## 第一次准备

1. 设置本机环境变量 `QIANTIE_MYSQL_DSN` 和 `QIANTIE_BRIDGE_SECRET`。
2. 如需队列，再设置 `QIANTIE_REDIS_URL`；没设置时只代表队列不可用，不会影响 API 骨架启动。
3. 执行 `powershell -File scripts/build-go-local.ps1`。

## 启动与检查

执行 `powershell -File scripts/start-local-go.ps1` 启动 Go。

另开窗口执行 `powershell -File scripts/check-local-go.ps1`，它会检查 `/health` 和首页。

Node/Vite 仍可用于前端开发，但不要让它和 Go 同时占用同一个端口。
