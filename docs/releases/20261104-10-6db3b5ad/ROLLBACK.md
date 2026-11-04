# 回滚说明（20261104-10-6db3b5ad）

本版 = HEAD `6db3b5ad6a8d3ebc6b0350ac459725a75a9457f8` + 工作区 diff（工作区改动 237 条，未跟踪 67 条）。

## A. 线上回滚（二进制 + presets）

```powershell
# 查看回滚点
.\scripts\release-v88.ps1 -ListBackups -TargetHost <host> -RemoteDataDir <prod-data>

# 回滚到指定批次（远端备份目录 .rollback/20261104-10-6db3b5ad）
.\scripts\release-v88.ps1 -Rollback -BackupId 20261104-10-6db3b5ad -TargetHost <host> -RemoteDataDir <prod-data>
```

远端备份位置：`/opt/qiantie/releases/v88/.rollback/20261104-10-6db3b5ad`（含 `qiantie-server` 与 `presets.json`）。

## B. 本机 4001 回滚

```powershell
# 不要直接覆盖运行中的 exe；先停止 4001 对应进程，再复制本批次二进制
Copy-Item ".release/v88/20261104-10-6db3b5ad/qiantie-local.exe" "backend/qiantie-local.exe" -Force
.\scripts\start-local-go.ps1 -SkipBuild
```

本轮构建的二进制位于 `.release/v88/20261104-10-6db3b5ad/qiantie-local.exe`；使用 `-UpdateLocalBinary` 时
旧二进制会备份为 `.release/v88/20261104-10-6db3b5ad/previous-qiantie-local.exe`。

## C. 源码回滚

```powershell
git checkout -- .
git apply --3way "docs/releases/20261104-10-6db3b5ad/changes.diff"
# 未跟踪文件内容见 docs/releases/20261104-10-6db3b5ad/untracked-snapshot/
```

注意：`changes.diff` 只覆盖已跟踪文件的未暂存改动；未跟踪源码在
`untracked-snapshot/` 内，清单见 `untracked-snapshot-manifest.txt`。

## D. 数据回滚

```powershell
# 远端 presets 备份目录：/opt/qiantie/releases/v88/.rollback/20261104-10-6db3b5ad
#   其中 presets.json 为上传前的线上画风库，qiantie-server 为上传前的二进制
# 本地原始数据仍在 data/system/presets.json（本脚本不修改本地数据）
```

## E. 实际部署路径（2026-11-04 23:34 CST，已执行）

线上路径 `/opt/www/qiantie`，备份在 `/root/backup-20261104/`（TS=20261104-2332）：

| 文件 | 字节 | SHA256 |
|---|---|---|
| `qiantie.prev-20261104-2332`（09 线上二进制） | 70,668,450 | `0785707c…1827e61` |
| `qiantie.service-20261104-2332.txt` | 397 | `4f6634e8…cf81ec257` |
| `qiantie-db-pre-20261104-2332.sql.gz` | 69,762,622 | `1225033d…b303224e8` |
| `qiantie-data-pre-20261104-2332.tgz` | 164,560,835 | `ff2a5f26…0dd8a855` |

回滚（只换二进制 + 重启，不动数据）：

```bash
cp -p /root/backup-20261104/qiantie.prev-20261104-2332 /opt/www/qiantie/qiantie.new-rollback
chmod +x /opt/www/qiantie/qiantie.new-rollback
mv -f /opt/www/qiantie/qiantie.new-rollback /opt/www/qiantie/qiantie   # 运行中二进制不能直接 cp 覆盖（Text file busy）
systemctl restart qiantie.service
curl -s -o /dev/null -w '%{http_code}\n' http://127.0.0.1:4000/health   # 期望 200
curl -s -o /dev/null -w '%{http_code}\n' http://127.0.0.1:4000/assets/user-DmXbWs5q.js  # 期望 200（09 入口恢复）
```

`runtime-release.json` 需同步回填为 09 版（`buildId=20261104-09-6db3b5ad`、`frontendEntry=/assets/user-DmXbWs5q.js`、`binarySha256=0785707c…1827e61`）。本批未改线上数据。