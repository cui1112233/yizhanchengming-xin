# 回滚说明（20261104-06-6db3b5ad）

本版 = HEAD `6db3b5ad6a8d3ebc6b0350ac459725a75a9457f8` + 工作区 diff（工作区改动 182 条，未跟踪 52 条）。

## A. 线上回滚（二进制 + presets）

```powershell
# 查看回滚点
.\scripts\release-v88.ps1 -ListBackups -TargetHost <host> -RemoteDataDir <prod-data>

# 回滚到指定批次（远端备份目录 .rollback/20261104-06-6db3b5ad）
.\scripts\release-v88.ps1 -Rollback -BackupId 20261104-06-6db3b5ad -TargetHost <host> -RemoteDataDir <prod-data>
```

远端备份位置：`/opt/qiantie/releases/v88/.rollback/20261104-06-6db3b5ad`（含 `qiantie-server` 与 `presets.json`）。

## B. 本机 4001 回滚

```powershell
# 不要直接覆盖运行中的 exe；先停止 4001 对应进程，再复制本批次二进制
Copy-Item ".release/v88/20261104-06-6db3b5ad/qiantie-local.exe" "backend/qiantie-local.exe" -Force
.\scripts\start-local-go.ps1 -SkipBuild
```

本轮构建的二进制位于 `.release/v88/20261104-06-6db3b5ad/qiantie-local.exe`；使用 `-UpdateLocalBinary` 时
旧二进制会备份为 `.release/v88/20261104-06-6db3b5ad/previous-qiantie-local.exe`。

## C. 源码回滚

```powershell
git checkout -- .
git apply --3way "docs/releases/20261104-06-6db3b5ad/changes.diff"
# 未跟踪文件内容见 docs/releases/20261104-06-6db3b5ad/untracked-snapshot/
```

注意：`changes.diff` 只覆盖已跟踪文件的未暂存改动；未跟踪源码在
`untracked-snapshot/` 内，清单见 `untracked-snapshot-manifest.txt`。

## D. 数据回滚

```powershell
# 远端 presets 备份目录：/opt/qiantie/releases/v88/.rollback/20261104-06-6db3b5ad
#   其中 presets.json 为上传前的线上画风库，qiantie-server 为上传前的二进制
# 本地原始数据仍在 data/system/presets.json（本脚本不修改本地数据）
```

## E. 本批实际执行的回滚点（2026-11-04 真实部署）

| 环境 | 旧 SHA256 | 新 SHA256 |
| --- | --- | --- |
| 本机 4001 | `e0a82f13c93dbe6bc1bf1f0453bfed5f4f30c69e0e48848044593537db6c1e48`（PID 25300） | `28b7019165b60ac152260d35947707abe472d481bb3f1f9a47bd6f2a0dbbb661`（PID 32800） |
| 线上 | `2fbea42470ca1e8d1148c6489d33e4f5f10676245f93d979071cd2a6d373f663` | `c1785f00ba3e55911e469a9f5c1584369353687d71b5131ff445f0ce09f2e7f9` |

线上备份：

- `/root/backup-20261104/qiantie.pre-20261104-165530`
- `/root/backup-20261104/runtime-release.pre-20261104-165530.json`

上线回滚（旧二进制 + 旧 runtime-release.json + 重启）：

```bash
cp -p /root/backup-20261104/qiantie.pre-20261104-165530 /opt/www/qiantie/qiantie
cp -p /root/backup-20261104/runtime-release.pre-20261104-165530.json /opt/www/qiantie/runtime-release.json
systemctl restart qiantie.service
curl -s http://127.0.0.1:4000/api/runtime-build-info
# 预期 frontend_entry 回到 /assets/user-BdLg_lUI.js
```

本机回滚（旧二进制在 `.backups/release-06/qiantie-local-pre-20261104-06.exe`）：

```powershell
Stop-Process -Id 32800 -Force
Copy-Item ".backups/release-06/qiantie-local-pre-20261104-06.exe" "backend/qiantie-local.exe" -Force
# 把 .local-review/runtime-manifest.json 的 frontendEntry 改回 /assets/user-BdLg_lUI.js
.\scripts\start-local-go.ps1 -SkipBuild
```

数据：本批未上传线上 presets，线上画风库未被覆盖，无需数据回滚。