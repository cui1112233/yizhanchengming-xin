# 回滚说明（20261104-03-6db3b5ad）

本版 = HEAD `6db3b5ad6a8d3ebc6b0350ac459725a75a9457f8` + 工作区 diff（工作区改动 146 条，未跟踪 39 条）。

## A. 线上回滚（二进制 + presets）

```powershell
# 查看回滚点
.\scripts\release-v88.ps1 -ListBackups -TargetHost <host> -RemoteDataDir <prod-data>

# 回滚到指定批次（远端备份目录 .rollback/20261104-03-6db3b5ad）
.\scripts\release-v88.ps1 -Rollback -BackupId 20261104-03-6db3b5ad -TargetHost <host> -RemoteDataDir <prod-data>
```

远端备份位置：`/opt/qiantie/releases/v88/.rollback/20261104-03-6db3b5ad`（含 `qiantie-server` 与 `presets.json`）。

## B. 本机 4001 回滚

```powershell
# 不要直接覆盖运行中的 exe；先停止 4001 对应进程，再复制本批次二进制
Copy-Item ".release/v88/20261104-03-6db3b5ad/qiantie-local.exe" "backend/qiantie-local.exe" -Force
.\scripts\start-local-go.ps1 -SkipBuild
```

本轮构建的二进制位于 `.release/v88/20261104-03-6db3b5ad/qiantie-local.exe`；使用 `-UpdateLocalBinary` 时
旧二进制会备份为 `.release/v88/20261104-03-6db3b5ad/previous-qiantie-local.exe`。

## C. 源码回滚

```powershell
git checkout -- .
git apply --3way "docs/releases/20261104-03-6db3b5ad/changes.diff"
# 未跟踪文件内容见 docs/releases/20261104-03-6db3b5ad/untracked-snapshot/
```

注意：`changes.diff` 只覆盖已跟踪文件的未暂存改动；未跟踪源码在
`untracked-snapshot/` 内，清单见 `untracked-snapshot-manifest.txt`。

## D. 数据回滚

```powershell
# 远端 presets 备份目录：/opt/qiantie/releases/v88/.rollback/20261104-03-6db3b5ad
#   其中 presets.json 为上传前的线上画风库，qiantie-server 为上传前的二进制
# 本地原始数据仍在 data/system/presets.json（本脚本不修改本地数据）
```

## E. 本批实际切换记录（本机 4001）

| 项目 | 旧 | 新 |
| --- | --- | --- |
| PID | 37896（2026-11-04 12:39:13） | 38352（2026-11-04 13:45:11） |
| 二进制 SHA256 | `89c82e424c9945c2a9422b5943fa215c4cc464fed0dfc9154519241a7468807c` | `7870f4bbf16c492f1d2c46c8e6a6c9f8f09e5decb5a6ef3841babc3243ccc709` |
| 前端入口 | `/assets/user-CKV6--Bb.js`（批次 02） | `/assets/user-CEHz_Uc5.js` |
| buildId | 20261104-02-6db3b5ad | 20261104-03-6db3b5ad |

旧二进制已备份为 `.release/v88/20261104-03-6db3b5ad/previous-qiantie-local.exe`
（与 `.release/v88/20261104-02-6db3b5ad/qiantie-local.exe` 内容一致）。

回滚到批次 02 运行时：

```powershell
Stop-Process -Id 38352 -Force
Copy-Item ".release/v88/20261104-03-6db3b5ad/previous-qiantie-local.exe" "backend/qiantie-local.exe" -Force
# 同步回退构建标记与运行清单（否则启动脚本会认为二进制与清单不符）
#   backend/current-release.json -> buildId 20261104-02-6db3b5ad
#   .local-review/runtime-manifest.json -> frontendEntry /assets/user-CKV6--Bb.js
.\scripts\start-local-go-review.ps1 -SkipBuild
curl.exe -s -o NUL -w "%{http_code}" http://127.0.0.1:4001/health
```

注意：本批只切换本机 4001，远端 Deploy/Rollback 未执行、未验证。