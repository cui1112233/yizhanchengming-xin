# ZIP第五批：主Agent调度与模型映射

2026-11-17继续计划41。本批未改产品代码、未部署或重启4001，没有付费文字/图片/视频生成。采用固定ZIP原生接口、原调度器和真实文件；文字客户端模拟，图片入口明确拒绝生成。

## 完成的验证

原主流程15组：正常与换序、重复步骤、未知步骤、仅输入、空内容、缺路径、部分失败、取消、剧本生成、编号入口组合、重复提交，以及文字解析/剧本/图片步骤衔接。记录最终文字消息、各步状态、进度和文件。

原前端$r等函数回放13组真实原生进度；正常/部分失败/剧本三类再走原Jo/ya、真实原生保存及新进程重开。26项记录校验通过。并运行我们现有模型和媒体配置的六项定向Go测试，6项通过。这些是受控链路和本地代码证据，不是完整真实窗口或供应商验收。

## 新确认的规则和缺陷

- 主流程按pipeline顺序执行，使用本次model；不能套用资产独立重解析的读作品配置规则。
- 中间失败后后续仍可继续，保留成功结果。空原文导致四步全失败，但外层仍success=true/所有Agent执行完成；CM必须逐步汇总，不能只读外层成功。
- 未知步骤被跳过仍报完成；重复步骤执行两次；重复提交整个流程又发完整请求。CM要区分合法重复编排和同一任务重复提交。
- 取消停止后续请求，已完成摘要文件仍保留；没有证明真实在途HTTP终止。
- 剧本后图片步骤在明确出图失败时仍被标completed；这会使没有有效图片的流程看起来完成，需要修正。
- 原生返回的各步result并非统一data包装。实际进度回调直接更新正确数据；完整复刻必须连同事件和保存一起对应。

已恢复故事摘要原生system，1191字符，正文在readable-native-agent-prompts/synopsis-system.txt，摘要/来源见analysis/synopsis-system-provenance.json。与前三份资产system分开记录。

## 模型映射的边界

05-model-mapping已写字段与选择规则，要求使用本账号目录和自己的模型ID。现有图片生成/编辑配置与配音/音色需分开。此次4001型号目录请求返回401，所以实际可用型号名单仍未确认，不能说目录为空或模型都能用。没有修改认证配置。

4001只读版本快照git_sha=runtime_manifest_sha=9f28ee75、PID26264；不是本批部署。我们源码和运行版本分别记录，不将源码测试当作该二进制验收。

## 证据位置

固定证据根：F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-screenplay-reconstruction-20261115/07-parity-execution/zip-reconstruction-20261116/

analysis/native-api-isolated/agent-pipeline-v2/v3.json；agent-pipeline-reopen-v2.json；analysis/agent-progress-replay-v2.json；agent-ui-save-reopen-v1.json；agent-pipeline-validation.json；cm-model-mapping-focused-tests.jsonl；cm-model-mapping-readonly-v2-20261117.json。

REA实际分析$r、ka、ya、xa四份原函数：1128节点、无解析失败。摘要ev_106bd0be0092e2e75f08eae3020f5470ff6fad3c9b615f9904a9dee39c6d2718，覆盖只限切片，不代表整包语义分析完整。

## 剩余

长原文/跨段处理、全部Agent组合、音色与配音、视频及成品输出、该ZIP规划/续写、站位和场景多视图入口、真实窗口/图标/点击，以及实际可用模型目录仍需补齐。完整规格后实施，再到4001做成套对照。不得据本批宣布一比一完成。
