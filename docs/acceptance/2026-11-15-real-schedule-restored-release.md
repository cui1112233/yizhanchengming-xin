# 模型恢复后的真实定时测试与上线验收

2026-11-15，承接用户“测试成功上线”和“可以了，测试”的授权。此记录取代同日失败阶段的发布阻断结论。

## 真实结果

- 候选程序实际生成新的AI1正文3674字，先前不存在该版本。模型为账号已配置的gemini-3.7-flash-high。执行证据显示初始预算1200、实际恢复预算12000、调用两次后成功；正文已保存。
- 找到提交阻碍：顶层web_submit.enabled已经为true，但执行配置appConfig.web_submit缺少对应有效项。通过配置接口保存提交开关后，两处均验证为true，保留原账号和凭据。
- 缺失的提交通过明确重试继续，没有重新生成AI1。目标系统确认回执289167，verified=true。
- 实际发布版本：2a979f907451a3a209bfbc888ef1461f8b1e1899。公网health正常，git_sha和runtime_manifest_sha均匹配。
- 发布后新建定时任务schedule-20261115154418.159903599，选择此前不存在的AI2，明确使用普通改文方式和自动提交。
- 该任务到点执行，实际生成并保存AI2正文3730字，目标系统回执289168且verified=true。定时任务completed、aiStatus=done、siteSubmitStatus=submitted，已确认版本同时包含ai1和ai2。

## 发布范围和回退

生产源码基点d45d2a4b，仅带入批量测书相关后端修复以及2a979f90的截断预算调整。保留原生产前端186个嵌入文件，未把其他模块正在开发的界面一起发布。

截断恢复最多三次，预算按1200→12000→24000有上限增加；非截断错误不会因此重复调用。源码核心包检查通过，生产候选相关HTTP回归通过，新预算额外流程回归通过。

Linux二进制SHA256：db0c3547db718a9c3e3a1cb3522be86ca38ed6aa337cc3973dbc254ed84b3b56。

服务版本路径：/opt/www/qiantie/qiantie-20261115-schedule-2a979f90。回退资料：/opt/www/qiantie/rollback-qiantie-20261115-schedule-2a979f90。切换脚本在健康、版本或入口校验失败时恢复旧链接与环境配置。

本地4001在测试期间由另一项工作更新为95e06d2e，自动保护检查阻止了覆盖该进程。本次上线结论针对https://yzcm.site，不把本地4001等同于生产。

本地原始记录：F:/脚本测试/一战晟铭-v88/_local-archive-20261028/schedule-live-release-20261115/。生产测试资料：/opt/www/qiantie/diagnostics/schedule-live-20261115/。可提交的脱敏验收证据位于docs/verification/schedule-restored-release-2026-11-15/。
