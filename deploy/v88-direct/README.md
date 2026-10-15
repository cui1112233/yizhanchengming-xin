# V88 Go 二进制直部署

这是当前唯一的应用发布入口，不使用 Docker、Compose、GHCR 或 GitHub
Actions 编译/发布应用。

## 本地打包

在仓库根目录执行：

```powershell
powershell -ExecutionPolicy Bypass -File deploy/v88-direct/build-direct-release.ps1
```

脚本会构建 React 页面，把页面同步到 Go embed 目录，再交叉编译
Linux amd64 的 `qiantie`，最后生成：

```text
release/v88-direct-<完整 Git SHA>.tar.gz
```

## 服务器激活

```powershell
powershell -ExecutionPolicy Bypass -File deploy/v88-direct/send-direct-release.ps1 `
  -ComputerName <服务器地址>
```

服务器端由 `qiantie.service` 直接启动：

```text
/opt/qiantie/v88/direct/current/go/qiantie
```

激活过程会切换版本软链、重启 systemd 服务、检查 `/health` 和
`/api/runtime-build-info`，失败时自动恢复 `previous` 版本。

MySQL、Redis、TOS、FFmpeg 等已有基础设施不由应用发布脚本删除或重建。
