# 本轮免费核验

代码未修改；没有调用真实模型，没有启动定时/批量生成任务，没有执行删除。

2026-11-18，backend目录实际执行：

```powershell
go test ./internal/novelcreation -run 'TestNormalizeCreateProjectInput|TestNormalizeModeInputs|TestServiceRejectsMissingModeInputBeforePersistingMessage|TestStudioNoHiddenProjectAndDurableConfirmation' -count=1 -v
```

结果exit0，7个顶层测试通过，其中模式必填项含4个子例。测试核对：作品字数默认/边界、额外模式字段过滤、缺失输入不保存消息、无书讨论确认的持久行为。这证明当前实现规则，不证明规则与原版一致。

实际4001运行信息读回：git_sha=c4054404；frontend_entry=/assets/user-LoEy2Yrp.js。源代码HEAD=8ab7e06f，后者是文档提交；不需要重启来做本轮只读核查。

当前UI：六入口均已点击，长篇仿写及续写弹窗打开后取消；六工具已进入，备忘录与导入弹层关闭。现场输入为空，没有提交生成或保存新作品。原有作品数为1。

现场发现：短篇原创/仿写短篇/番外只切模式；短仿写参考控件是普通展开文本框；Agent仅8个模型路径；看板只有4个数字；通用文件上传因未配置TOS禁用，说明文字真实显示原因。

原版窗口洛水v2.9.4已运行，无障碍树可读并保存original-live-ax.txt；点击报coordinate input geometry is unavailable，截图window capture timed out。Raise和Escape没有关闭原弹窗。原版其余点击及同状态像素仍未完成，不据此登记通过。

ui-observations.json混合保存完整状态及相邻差异状态：单条差异状态要结合前文读取，最终短仿写为完整状态。不是133个原按钮或全部当前按钮的执行报告。
