# 回滚说明（20261104-04-6db3b5ad）

本版 = HEAD `6db3b5ad6a8d3ebc6b0350ac459725a75a9457f8` + 工作区 diff（工作区改动 158 条，未跟踪 46 条）。

## A. 线上回滚（二进制 + presets）

```powershell
# 查看回滚点
.\scripts\release-v88.ps1 -ListBackups -TargetHost <host> -RemoteDataDir <prod-data>

# 回滚到指定批次（远端备份目录 .rollback/20261104-04-6db3b5ad）
.\scripts\release-v88.ps1 -Rollback -BackupId 20261104-04-6db3b5ad -TargetHost <host> -RemoteDataDir <prod-data>
```

远端备份位置：`/opt/qiantie/releases/v88/.rollback/20261104-04-6db3b5ad`（含 `qiantie-server` 与 `presets.json`）。

## B. 本机 4001 回滚

```powershell
# 不要直接覆盖运行中的 exe；先停止 4001 对应进程，再复制本批次二进制
Copy-Item ".release/v88/20261104-04-6db3b5ad/qiantie-local.exe" "backend/qiantie-local.exe" -Force
.\scripts\start-local-go.ps1 -SkipBuild
```

本轮构建的二进制位于 `.release/v88/20261104-04-6db3b5ad/qiantie-local.exe`；使用 `-UpdateLocalBinary` 时
旧二进制会备份为 `.release/v88/20261104-04-6db3b5ad/previous-qiantie-local.exe`。

## C. 源码回滚

```powershell
git checkout -- .
git apply --3way "docs/releases/20261104-04-6db3b5ad/changes.diff"
# 未跟踪文件内容见 docs/releases/20261104-04-6db3b5ad/untracked-snapshot/
```

注意：`changes.diff` 只覆盖已跟踪文件的未暂存改动；未跟踪源码在
`untracked-snapshot/` 内，清单见 `untracked-snapshot-manifest.txt`。

## D. 数据回滚

```powershell
# 远端 presets 备份目录：/opt/qiantie/releases/v88/.rollback/20261104-04-6db3b5ad
#   其中 presets.json 为上传前的线上画风库，qiantie-server 为上传前的二进制
# 本地原始数据仍在 data/system/presets.json（本脚本不修改本地数据）
```

## E. 本批实际切换记录（本机 4001）

| 项目 | 旧 | 新 |
| --- | --- | --- |
| PID | 38352（2026-11-04 13:45:11） | 5612（2026-11-04 14:46:33） |
| 二进制 SHA256 | `7870f4bbf16c492f1d2c46c8e6a6c9f8f09e5decb5a6ef3841babc3243ccc709` | `f8ef9dd2d5672d3ea152aacf4b8dca953a4c8d9ce0b832d676f3f958d8655a52` |
| 前端入口 | `/assets/user-CEHz_Uc5.js`（批次 03） | `/assets/user-M4kZML8e.js`（与线上一致） |
| Slice | 3（enabled=9 / disabled=5） | 5（enabled=13 / disabled=[publish.yadi]） |

旧二进制备份：`.release/v88/20261104-04-6db3b5ad/previous-qiantie-local.exe`。

回滚到批次 03 运行时（含恢复 Slice 3 行为需同时回退脚本改动）：

```powershell
Stop-Process -Id 5612 -Force
Copy-Item ".release/v88/20261104-04-6db3b5ad/previous-qiantie-local.exe" "backend/qiantie-local.exe" -Force
# 同步回退 .local-review/runtime-manifest.json 的 frontendEntry 为 /assets/user-CEHz_Uc5.js
# 如需回到 Slice 3，请先撤销 scripts/start-local-go.ps1 与 start-local-go-review.ps1 的 Slice=5 改动
.\scripts\start-local-go-review.ps1 -SkipBuild
curl.exe -s -o NUL -w "%{http_code}" http://127.0.0.1:4001/health
```

本批只切换本机 4001；远端 `/opt/www/qiantie/qiantie` 未由本批变更。