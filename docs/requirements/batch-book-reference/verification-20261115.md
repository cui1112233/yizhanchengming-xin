# 2026-11-15 本轮验证记录

## 原包与分析材料

- 压缩包 SHA256：`5b872ae25ab46ba33d2dafc6d2bd6cd5e727c00a3f962e4de6ca9309c4911ba0`。
- 原包 43 个普通文件与独立解包目录的 43 个文件全部匹配。详见 archive-verification.json。
- 生成脚本运行成功；全部分析 JSON 可解析。707 个函数、184 个 HTML 控件、62 个 API 路由。171 个具有 ID 的控件都找到了 JS 文本引用；这只是静态定位，不能替代动态行为验收。
- 62 个原接口中，42 个找到同名主入口或显式兼容入口；剩余项含统一账号替代、前缀转发和未确认缺失。此比例不是完成率。
- 原系统的接口调用可达图包含直接函数、文件读写与网络边界；动态分派和分支顺序仍保留限制说明。

## 测试先重现，再最小修复

新增 novel_fetch_schedule_input_test.go。第一次运行时确认以下缺陷均存在：空书单、无效时间、没有有效书籍 ID、未知平台仍返回 200 并保存；字段别名没有统一；损坏时间卡片一直保持 scheduled。调整时间测试还确认：无效编辑会覆盖原时间。

修复后：

- 创建前检查输入，错误返回 400，未写入定时任务。
- 时间以北京时间解析后统一保存带时区时间；保留所选模型、版本及改文方式。
- 调整时间先校验，拒绝的修改不改变原任务；runAt 别名生效。
- 历史无效时间标 failed、保存清楚原因，不入队。
- 没有改动重复书策略、任务实例、收费生成或实际上传行为。

## 已执行并通过

在 backend 目录：

```text
go test ./internal/httpapi -run 'TestNovelFetchScheduleTimeEditValidatesBeforeChangingSavedTask|Schedule|Scheduled|WorkshopProcessWorkerAutomaticallyUploads|Workshop.*Retry|NovelFetch.*Retry|WebSubmit.*(Readback|Login)' -count=1
ok qiantie/backend/internal/httpapi 27.021s

go test ./internal/novelfetchworkshop ./internal/videomanagementconnector -count=1
ok qiantie/backend/internal/novelfetchworkshop 3.408s
ok qiantie/backend/internal/videomanagementconnector 0.086s
```

新增的创建校验、时间规范、历史损坏时间用例此前也单独执行并通过。未宣称全仓库测试通过。测试使用隔离存储和受控服务，没有实际书籍收费调用或真实发布。

## 运行版本边界

本轮只读查看本地 4001 的 `/api/runtime-build-info`，返回 git_sha=`23899a82`；进程来自另一个工作中的候选目录。没有停止它或覆盖服务，没有部署线上。本轮修复的验证是源码测试结果，不把现有 4001 当成已运行本轮修复。

仍须完成：保存即入库、任务实例与续做、输入列格式、摘要读取性能、真实目标站链路以及动态页面验收。下一轮按 execution-plan.md 和 v88-behavior-gaps.md 接续，不另起一套实现。
