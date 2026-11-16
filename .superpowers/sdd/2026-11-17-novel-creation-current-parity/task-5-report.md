# Task 5：Agent控制、模型路由和有界处理

日期：2026-11-17。范围仅Task5；直接在`v88`实施，没有新分支/worktree、付费模型调用、4001重启、推送或部署。基线为`83dab59691d99fcfeb67bf547b3e7d041920dfe9`。源码和本报告同一Task5本地提交，提交SHA由最终交接返回。

## 已实现

- 新提交冻结`agent_contract_version: 1`和完整有效`project_settings.agent`；每阶段按显式`modelRoutes`选模型，空路由采用本次提交选定的模型。缓存命中先复用已完成调用；只有新外呼才重新验证当前账号Active、管理员绑定、所选模型存在/启用和当前合法配置。不回退其他模型。
- 通过分默认80，范围0–100；基础温度默认0.65，范围0–2；修订增量默认0.15，范围0–0.5；内容修订默认3轮，范围0–5。JSON缺省/null修订数采用3，显式0就是0；旧已保存非零轮数保留。目标字数、思考设置继续生效。数值界限沿用设置归一化方式。
- `enableLengthNorm`、`enablePostWrite`、`enableAiTells`默认true，分别控制真正的长度整理、正文后处理和表达整理调用；任一关闭即无对应外呼。顺序为草稿→这些已启用步骤→审核/有界修订→事实→记忆，后续均接收最终选中的正文。
- 审核使用有效通过分；新审核提示词为同一ID的v2，明确`passed`与本轮分数/高危问题的关系；保留v1原文字节用于旧journal恢复。修订温度为`min(2, baseTemperature + round * temperatureStep)`，首轮修订round=1。其余正常步骤采用基础温度；结构修复独立采用原有0.1温度并单独记录。
- JSON结构修复对每个步骤最多额外1次；内容修订和结构修复分别计数、分别使用稳定key。达到内容修订上限仍有审核问题时保留有效正文，状态为`succeeded_with_warnings`，不无限重试。
- 真实提供方支持可选`supportsTemperature`/`supportsMaxTokens`声明；缺省未知保持旧payload，显式false才省略对应参数。两种真实HTTP payload builder均接通。自定义渠道保存只对白名单中的这两个可选bool进行清洗/保留；旧配置和预置渠道按Key可用的语义保留。
- 每个实际调用的journal保存选中模型ID、实际模型/渠道安全ID、实际目录账号、配置修订或`safeConfigDigest`、提示词版本、请求温度/上限和参数是否发送。检查点/最终`runtimeSnapshot`汇总这些记录及有效Agent设置。无密钥、认证信息或endpoint原文；安全digest不包含凭据、URL用户信息、path/query，因此不冒充密钥轮换版本。

## 二次确认与Task7接口

设置仍通过现有`GET/PUT /api/novel-creation/projects/{projectID}/settings`的`agent`对象读写。新增字段：

| 字段 | JSON类型/含义 |
| --- | --- |
| `version` | 后端有效设置版本，当前1；UI不必自行指定 |
| `modelRoutes` | `{阶段键: 当前授权模型ID}`，空值采用运行选中的模型 |
| `passScore` | number，0–100，默认80 |
| `baseTemperature` | number，0–2，默认0.65，显式0保留 |
| `temperatureStep` | number，0–0.5，默认0.15，显式0保留 |
| `maxRevisionRounds` | integer 0–5；缺省/null采用3，显式0禁止内容修订 |
| `enableLengthNorm` / `enablePostWrite` / `enableAiTells` | bool，缺省true |
| `requireSecondaryConfirm` | bool，缺省false |
| `targetChapterWords` / `thinkingLevel` | 沿用已有合同 |

阶段键：`foundation-plan`、`chapter-plan`、`chapter-draft`、`length-normalize`、`post-write`、`ai-tells`、`chapter-audit`、`chapter-revision`、`truth-settlement`、`memory-index`。结构修复复用对应阶段路由，journal key另加`:repair`；审核/修订轮次也有独立key。

`requireSecondaryConfirm=true`时，首个外呼由显式提交/蓝图确认授权；后续每一个新外呼（包括结构修复）都先暂停。已有completed结果的回放不重复确认、不重新调用。

暂停时`run.status=awaiting_confirmation`，`run.runtimeSnapshot`包含：

```json
{
  "confirmationKind": "next_model_call",
  "confirmationId": "绑定本次run/key/请求指纹/安全模型配置的确认编号",
  "callKey": "novel-creation.chapter-draft",
  "nextModelCall": {"chosenModelId": "…", "provider": {}, "promptId": "…", "promptVersion": 1},
  "agentSettings": {},
  "modelCalls": {}
}
```

确认使用现有`POST /api/novel-creation/runs/{runID}/confirm`，提交`{"confirmationId":"当前编号"}`，返回202和**同一run ID**。不修改不可变输入、不创建后继运行，不重做历史步骤。缺编号/错误编号拒绝；重复已经确认的旧编号只返回当前运行，即使下一检查点已经产生，也绝不批准新的外呼。编号绑定安全配置，配置变化不会让旧编号授权另一个请求。

蓝图确认仍提交原有`{"blueprint":...}`或`{}`，并保持原来创建唯一后继run的语义。UI必须用`confirmationKind`区分；不要把模型调用确认当作编辑蓝图。模型调用确认等待对应自动任务轮次状态为`awaiting_model_confirmation`，等待文案是“请确认后继续下一次模型调用”；蓝图等待仍为`awaiting_confirmation`和“请确认创作蓝图后继续”。两种等待均不增加成功章数/失败计数。

## 红→绿与实际请求证据

所有请求均为本地替身/回环HTTP服务器，无真实提供方或收费调用。

1. `TestAgentSettingsPresenceAndApprovedDefaults`先失败：`{}`归一化只有旧默认2轮，没有通过分等字段；`TestAgentPipelineRoutesZeroRevisionAndTemperature`先错误进入reviser并结构修复失败。实现后均绿，实际零温度保留、0轮无reviser，6个必需阶段选中各自模型。
2. `TestAgentOptionalStepsUseFinalContentAndEachToggle`先失败：直接进入audit、未执行可选步骤。实现后all/on/off逐项通过，实际调用数分别9或8或6；最终正文进入审核、事实和记忆。另有全部10种阶段路由的11次实际调用用例（含复审）。
3. 两个`TestNovel*SamplingOmissionPreservesOtherDefaults`先捕获到不应发送的`temperature/max_tokens`；实现后false声明对应字段均缺席，普通模块默认请求字段仍在。
4. `TestNovelSavedChannelCapabilitiesReachActualPayload`先失败：保存清洗丢失能力声明，实际HTTP仍收到温度/上限；白名单保存后从保存→渠道resolver→真实pipeline→HTTP body整链绿。
5. `TestDefinitions*`扩展为10个定义时先失败“7, want 10”，登记后绿。新增系统提示词及结构修复说明全部在`backend/internal/prompt/novelcreation`。
6. 二次确认初始红测暴露缺少`ConfirmModelCall`接口（编译期红）；实现后`TestAgentSecondaryConfirmationSameRunAndStaleToken`、HTTP归属/token测试、repair gate、重启恢复及蓝图source→child自动计数测试绿。等待文案补测先明确失败“请确认创作蓝图后继续”，区分模型检查点后绿。
7. 权限变化测试确认：已有foundation completed保留，下一步仍prepared、没有`may_have_sent`，外呼总数1，确定配置错误而非unknown，批量立即停止。HTTP resolver覆盖账号停用、管理员绑定变化、管理员禁用/删除指定模型。
8. 本地事件保存失败后恢复，已完成步骤不再次解析提供方，旧metadata保持v1，后续新调用采用当前v2安全配置；实际省略标志和“不支持，未发送”一起保存。

## 真实BASE旧记录A/B

没有伪造“新包生成的旧记录”。只读`git archive 83dab596… backend`到ignored目录`.tmp-task5-base-capture`，运行**原BASE完整后端**的`Service.Run`，仅新增捕获测试：旧假故事先结构repair一次、内容revision一次，共9次真实替身请求。对保存bundle前注入本地存储失败，得到原始`indexing`运行、`ready`执行和9条completed调用。

捕获文件：`backend/internal/novelcreation/testdata/task5-base-agent-replay.json`，包含来源SHA、原PromptSnapshot、调用原响应/指纹及实际请求。PromptSnapshot单独导出，因原Run类型对它设置`json:"-"`。全部为既有假故事，无凭据。

原包输出：`BASE actual requests=9 completed journal=9 state=ready status=indexing`。

新包`TestAgentActualBASEJournalReplaysWithoutNewDefaultsOrResolver`输出：`actual BASE requests=9; fingerprints equal=9; current replay provider=0; resolver=0; optional additions=0`。同一运行成功保存，原9条指纹/响应不变，未追加新默认步骤。新增request字段采用omitempty，旧请求序列化指纹保持相等。

无新SQL迁移：设置与执行记录原本已采用JSON，使用兼容读取/保存归一化完成字段升级；未占用新的迁移编号，也未改动最新8800313。旧运行缺少`agent_contract_version`时保留BASE旧步骤、原提示词与温度/阈值，旧明确轮数保留、旧缺省按2处理；新项目/新提交才采用新默认行为。

## 验证结果

- `go test ./internal/novelcreation ./internal/prompt/... -run 'Novel|Pipeline|Agent|Revision' -count=1`：通过。
- `go test ./internal/novelcreation ./internal/prompt/novelcreation -count=1`：完整domain和小说prompt测试通过。
- `go test ./internal/batchfactoryv11 ./internal/httpapi ./internal/app -run 'Novel|Sampling|IntegrationChannel|PublicManagedModelCapabilities|ConfiguredTextProvider|ResolveNovelFetch' -count=1`：通过（batchfactoryv11 0.048s、httpapi 10.033s、app 0.133s）。
- 对全部本任务Go文件执行gofmt；仅检查本任务暂存文件的diff，不修复其他聊天的脏文件。全仓`git diff --check`曾报告已有其他前端改动的空白问题，未修改它们。

## 文件与限制

实现：`novelcreation/{types,validation,pipeline,service,worker,worker_store,automation_state,model_confirmation}.go`；后端prompt定义；最小`TextCompletionRequest`/两种provider payload；渠道能力保存/公开能力；Novel专用权限resolver及app接线。测试：Agent行为、真实BASE fixture、prompt合同、两builder真实请求体、实际渠道保存链、HTTP确认/账号归属、动态权限；旧已有测试仅随新增真实步骤调整其替身序列或明确使用旧合同。

Task7前端尚未接入；未跑4001、未重建Go embed、未验收原版真实模型行为或像素一比一。原版高级`reviewPrompts/customWritePrompts`不在批准Task5范围，本次未还原。安全digest不等价于包含所有私密配置的完整版本，也不用于承诺密钥轮换追踪。继承的storage测试116/115迁移计数问题按Task8保留，本次未扩大修复。

临时目录清理限制：对`F:\脚本测试\一战晟铭-v88\.tmp-task5-base-capture`的两次原生PowerShell删除请求被自动审批拒绝，工具仅返回`rejected: blocked by policy`，没有更详细原因。绝对路径已核验；停止删除尝试，没有换工具绕过。ignored临时目录因此保留，未暂存/提交，不影响功能测试。
