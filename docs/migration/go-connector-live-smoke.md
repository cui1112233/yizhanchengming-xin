# Go 连接器真实验收

`backend/cmd/video-management-smoke` 是不依赖 Node/浏览器 Worker 的真实目标站验收入口。

它按固定顺序执行：

1. 使用环境变量中的账号密码登录视频管理系统；
2. 检查登录会话仍有效；
3. 读取组织目录和网站配置档；
4. 如果提供 `QIANTIE_VIDEO_SMOKE_BOOK_ID`，再读取该书的回读数据。

密码只从进程环境读取，不写文件、不打印。示例：

```powershell
$env:QIANTIE_VIDEO_SMOKE_USERNAME = "..."
$env:QIANTIE_VIDEO_SMOKE_PASSWORD = "..."
$env:QIANTIE_VIDEO_SMOKE_BOOK_ID = "2091850217129763895"
Push-Location backend
go run ./cmd/video-management-smoke
Pop-Location
```

成功输出中的 `runtime` 必须是 `go`，`session` 必须是 `valid`，并且 `book_readback` 为 `true`（提供书号时）。此命令不执行上传，避免未经明确授权修改目标站数据；上传验收必须使用单独的测试书和测试文件。
