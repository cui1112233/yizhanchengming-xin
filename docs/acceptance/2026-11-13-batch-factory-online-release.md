# 批量工厂线上发布验收

用户在询问是否上线后明确回复“执行”，授权本次线上更新。没有推送 Git、修改数据库记录、调用真实模型或提交真实生产/重试/发布任务。

- 发布地址：https://yzcm.site。
- 发布源码基点：9cd2fd570105940fe9f4f1f310a64a55a1793431。
- 发布编号：qiantie-202611131955-9cd2fd57。
- 使用 scripts/deploy-go-embedded.ps1 构建 Go embed 单文件，上传至已有服务 qiantie.service 的版本目录并切换软链接。
- Linux 文件 SHA256：eb425e6ac56f2c48470131a7fbe2e23c9b586343f4dc8a48832286ae8822766d。
- 前端入口：/assets/user-D05QZakN.js，为此前已验收的前端产物。没有使用其他会话正在修改的前端重新构建。
- 后端取本地冻结快照，包含工作区既有未提交修改；不声称仅由干净 HEAD 构建。快照及逐文件 Go 源码散列保存在本地忽略目录 _local-archive-20261028/release-20261113-automation/candidate-20261113-195519。

## 发布前检查

冻结快照的 BatchFactory、BatchAutomation、ComicPage、CustomCopy 相关 Go 检查通过；补充 TestCustomCopies、TestProcessJobExistsUsesExplicitRedisExistence 通过。前端菜单、徽标、文案选择相关 12 项检查通过。Windows、Linux 构建通过；同一快照在本地 4012 实际运行，health 正常，runtime-build-info 的完整源码标识和前端入口相符。

线上发布前没有 running/queued/scheduled 自动化任务，没有 generating 漫画页。保留原有历史阶段记录和 paused/needs_attention 等状态。

## 线上结果

- 公网 /health 返回 ok=true。
- 公网 /api/runtime-build-info 的 git_sha 和 runtime_manifest_sha 均为完整 9cd2fd570105940fe9f4f1f310a64a55a1793431，frontend_entry 与候选一致。
- 公网前端脚本返回 HTTP 200。
- 已登录公网批次 batch_aa22e54ddc24063eb4635a789af03310 刷新后，实际菜单显示“暂停任务（可继续）”“继续未完成任务”“结束本轮任务”，并展示暂停/结束的行为说明。实际失败步骤数 2，可重试入口可用；未点击收费重试。
- 小说列表的 20 个行内操作均为“查看”，没有重复的“执行这本”或“插队”；生产表工具栏仍保留插队。验收后返回生产表。
- 截图：.local-review/screenshots/online-automation-controls-20261113.jpg。

## 回退点

旧版本 /opt/www/qiantie/qiantie-202611131753 保留。服务器侧回退目录 /opt/www/qiantie/rollback-qiantie-202611131955-9cd2fd57 保存旧链接目标、旧二进制散列、旧版本说明与环境配置备份；环境配置备份不下载、不提交。发布只更新版本标识配置，不改账号、模型、数据库或存储配置。切换脚本在服务或健康/版本检查失败时自动恢复旧链接与备份。
