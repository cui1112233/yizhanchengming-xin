# 真实短篇仿写链路测试

结论：本轮真实模型链路未走通，卡在首步拆书分析，未生成正文；不能用固定回复回归通过替代真实成品验收。

## 授权及预算

用户明确允许最多12次文本请求，禁止图片/视频，并允许临时改为同渠道anti-2模型、测试后恢复。实际最多4次：两次UI正常流程起步，两次独立单次通道检查。单次检查和业务worker都用SingleAttempt，无自动传输重试。查询模型列表和网络HEAD检查不生成文本。

为了限制消耗，在第二次UI起步前临时启用账号默认的逐次模型确认、将旧策划修订轮数设0；数据库验证设置true/0后才创建任务。原值false/3已恢复。预算记录 `_local-archive-20261028/novel-live-chain-20261120/budget.json`，Used=4、Limit=12，已停止进一步收费测试。

## 真实页面与结果

- 已登录用户：choushiyiguai，测试地址 `http://127.0.0.1:4001/novel-creation`。
- 空参考原文时开始仿写禁用；填写自写326字参考后可用。取消弹窗不建立作品，重开保留本地草稿。
- 第一任务：`project-17abe20b7ccbf749bf9fad75`，run `run-e7658eca7c094264b967b196`，暂定名“链路实测-1120-渡口来信”，目标1000字、1章、开启蓝图后暂停。新full协议1已实查。
- 实际调用记录停在 `novel-creation.short-imitation-deconstruction:chunk:0`，State=may_have_sent，任务paused/novel_upstream_result_unknown，0章，尚无蓝图或正文。
- 单次配置通道检查使用同一个实际resolver及真实API，明确返回HTTP503，JSON error.code=model_not_found，实际模型 `[aistudio]gemini-3.8-flash`。上游请求ID `202611200710187167980228268d9d6jNnY4o0g`。
- 同origin的只读模型列表没有aistudio，包含anti-2/anti/vertex的gemini-3.8-flash。用户批准后只改原渠道的modelId/displayName为anti-2，保留URL、凭据和其他配置。
- 第二任务：`project-0f4b66996788063728968745`，run `run-d8ebe63d28db9ed9f8b85a1c`，暂定名“链路实测-1120-anti2”。实际快照逐次确认true，仍在首个拆书调用暂停，没有继续产生更多模型调用。
- anti-2的独立单次检查也返回HTTP503/model_not_found，请求ID `202611200717136315878868268d9d62fEiXKyB`。**列表包含模型不等于当前通道可执行。**
- 两个UI调用的原始错误被旧worker隐藏为unknown，未把独立检查的响应冒充原UI请求的逐字响应。两个任务保持真实暂停记录，未人为改成成功，也未重复提交未知调用。
- 原模型及显示名称已恢复为aistudio，Agent默认项恢复false/3，SQL条件更新只恢复本轮明确修改的字段，其他并发配置保留；证据 `config-restored.txt`。

## 发现并修复的问题

1. 原worker只将400/401/403/404/422/429归入明确拒绝；即使503的结构化错误明确为model_not_found，也被显示为“上游结果待确认”。新规则仅对**503且JSON error.code严格等于model_not_found**标记rejected/novel_model_unavailable，提示检查配置或更换模型；普通503、非结构化网关错误仍保持unknown保护，不自动重试。
2. 失败/暂停的结束事件带progress100，RunCard取所有事件最大值，导致首步失败却出现绿色满格。新进度逻辑排除失败、取消、暂停、待确认的结束事件，保留最近已执行阶段进度；只有succeeded/succeeded_with_warnings显示100和成功色。

后端回归先明确失败（known503被归为may_have_sent），修复后通过；同时验证普通503仍unknown和重复调用不重发。进度回归覆盖首步失败、暂停、待确认、取消和真正成功。

## 不收费回归与运行更新

- Go novelcreation/prompt/storage三包通过（4.829s/0.425s/0.213s）。
- 进度测试2/2通过，实际旧事件序列“阶段5→暂停100”显示5%、非成功状态。
- 新全篇隔离MySQL、HTTP回归、构建和本地页面复核结果在完成后登记。
- 本轮模型检查的临时测试程序在archive目录，正常测试和产品中不包含收费探测入口；未保存/输出真实API凭据。

证据目录：`_local-archive-20261028/novel-live-chain-20261120/`。当前程序最初SHA256 `62404758bbb709becae80494e207dbf095d70164e36cdf5fee9dceafc25a66b9`，4001的git信息为空，未沿用上一轮d421ccb1身份声称此时实际二进制相同。后续修复版本会单独登记。
