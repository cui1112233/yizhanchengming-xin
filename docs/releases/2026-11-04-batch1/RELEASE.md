# V88 第一批发布存档（2026-11-04-batch1）

> 本版 = **HEAD `6db3b5ad6a8d3ebc6b0350ac459725a75a9457f8` + `git-diff-worktree.patch` + `untracked/` 目录**
> （工作区同时存在多路会话未提交改动，本批未做 `git commit` / `git add` / `git stash`，
> 因此以"源码快照 + 二进制 SHA"等价替代 git tag，实现可追溯与可回滚。）

## 1. 存档内容

| 文件 | 内容 |
| --- | --- |
| `head-sha.txt` | `git rev-parse HEAD` 原始输出 |
| `git-status-porcelain.txt` | `git status --porcelain` 全量（118 条改动路径） |
| `git-diff-worktree.patch` | `git diff` 全量补丁（已跟踪文件的未暂存改动） |
| `git-diff-stat.txt` | `git diff --stat` 变更统计 |
| `untracked-files.txt` | 未跟踪路径清单（`git status --porcelain` 的 24 条，含目录聚合） |
| `untracked/` | **47 个未跟踪文件的原始内容**，保持仓库相对路径（2,471,768 bytes） |
| `untracked-sha256.txt` | `untracked/` 内 47 个文件的 SHA256 + 相对路径 + 字节数 |
| `build-info.txt` | 构建命令、耗时、二进制大小与 SHA256 |
| `runtime-manifest.json` | 构建脚本生成的运行清单（gitSha / frontendEntry / builtAt） |

## 2. 本批带上去的功能

1. 七档画幅：16:9 / 9:16 / 1:1 / 4:3 / 3:4 / 3:2 / 2:3。
2. 76 条画风/预设种子（`backend/internal/prompt/defaults/admin_presets.json`，含
   `batch-comic-style-13` v2「Q版豆丁2D漫风」）。
3. 宠物默认值收敛为「小家伙」+「（默认）」标记 + CM 回落 bug 修复（不含老用户改写）。
4. 后端 S1：capabilities 派生 + JSON 404 + `comicReadiness` 权威字段 + Slice 默认 3。
5. 私人定制阶段 1：外壳与三栏结构（`.private-studio-*`）。

## 3. 构建

- `npm.cmd run build`：exit 0，16.91s，入口 `dist/assets/user-bhhBrVux.js`。
- `scripts/build-go-local.ps1`：exit 0，同步 `frontend/dist` → `backend/web/dist` 并编译。
- 新二进制：`backend/qiantie-local.exe`，77,315,584 bytes，
  SHA256 `5EBE4FDDD68B4E11D29DEDFC9A94CC54ED28111E52DE31D8CAAF2F6A8A9C93F4`。
- **4001 启动时触发重建**（`frontend/src/user/pages/ProfilePage.jsx` 于 11:59:36 被其它会话修改），
  实际部署二进制为 **77,317,632 bytes / SHA256 `E2A4D76F…788DC` / 入口 `/assets/user-2gZgH99m.js`**，
  与 4002 烟雾验证所用的 11:56 版（入口 `/assets/user-bhhBrVux.js`）不同；Go 侧同源。

## 4. 回滚步骤

**A. 二进制回滚（最快）**

```powershell
Stop-Process -Id <当前 4001 PID>
Copy-Item "$env:TEMP\v88-backup-<时间戳>\qiantie-local.exe" "f:\脚本测试\一战晟铭-v88\backend\qiantie-local.exe" -Force
& "f:\脚本测试\一战晟铭-v88\scripts\start-local-go.ps1" -SkipBuild
```

注意：`start-local-go.ps1 -SkipBuild` 只复用现有二进制，不重建前端/Go。
若回滚到更早的二进制，需同时确认 `backend/web/dist` 与二进制版本匹配。

**B. 源码快照回滚（从快照重新构建）**

```powershell
# 前提：当前工作区改动已暂存/另存，避免快照覆盖丢失
git -C "f:\脚本测试\一战晟铭-v88" checkout -- .
git -C "f:\脚本测试\一战晟铭-v88" apply --3way "docs/releases/2026-11-04-batch1/git-diff-worktree.patch"
# 再恢复未跟踪文件：把 untracked/ 下的内容按相对路径复制回仓库根目录
Copy-Item -Path "docs/releases/2026-11-04-batch1/untracked/*" -Destination "f:\脚本测试\一战晟铭-v88" -Recurse -Force
```

**C. 更快回滚点**

上一篇已知可用二进制备份：`%TEMP%\v88-backup-20261104-103339\qiantie-local.exe`
（SHA256 `C72A2CA231ED53B18A513615A29A2F3E87BFC0FB18B9F861838BDF972B6EB7C7`）。

## 5. 已知限制（重要）

- 本批**没有 git commit**：仓库 HEAD 仍是 `6db3b5ad`，`git log` 无法区分本批与之前版本。
  追溯只能依赖本目录快照 + 二进制 SHA256。
- `git-diff-worktree.patch` 只覆盖**已跟踪文件**的未暂存改动；未跟踪文件的内容已完整保存在
  `untracked/`（47 个文件 / 2,471,768 bytes），恢复时直接复制回对应相对路径即可。
- **已排除（未归档）的未跟踪文件**，原因如下：
  - `.backups/mysql-full-20261103T114757Z.sql.gz`（60.5 MB 全库备份，含用户数据/凭据，不入库存档）
  - `.backups/pet-default-migration/20261104-114526/choushiyiguai/api-config.json`（单用户配置备份）
  - `docs/releases/2026-11-04-batch1/*`（存档自身，避免递归包含）
  上述文件不参与本版重建；若需还原对应运行态数据，请使用各自原始备份位置。
- 发布前 4001 运行中的二进制（PID 21332，11:10:46 启动）被本次构建在原路径覆盖，
  其原始字节未留存，二进制级回滚请使用上面的 A/C 方案。
