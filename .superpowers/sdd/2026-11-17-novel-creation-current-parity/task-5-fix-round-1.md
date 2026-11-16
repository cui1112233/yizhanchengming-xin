# Task5 Fix Round 1：确认编号配置往返卡住

日期：2026-11-17。修复基线：`e9191a2d01e9f2646a6c9c316c9788a9554fa527`。仅处理审查的Important/P2：待调用配置A→确认A→改B→确认B→回A时，确定性编号回到旧A，历史批准记录吞掉当前确认。

## 红测

先新增真实服务/持久调用记录/请求捕获测试，执行：

`go test ./internal/novelcreation -run TestAgentConfirmationGeneration -count=1`

原实现明确失败：

- `TestAgentConfirmationGenerationConfigAToBToA`：`A1/B1/A2 must be unique: first==third=true`。
- `TestAgentConfirmationGenerationRepairsLegacyApprovedPending`：`legacy approved pending was neither refreshed nor kept unsent`。

均使用假故事、提供方替身，不调用真实模型。

## 修复

- `CallRecord`新增可选持久`ConfirmationGeneration`和`ConfirmationFingerprint`，继续存于现有JSON，不需要SQL迁移。
- 编号仍绑定run、call key、请求指纹和安全模型配置，并增加该调用的持久代次。只有真正产生新pending才递增；同一pending的再次观察/重启保持原编号。A1/B1/A2因此分别唯一。
- 保留旧已批准编号的幂等早返回；旧A1/B1不能批准A2，也不能批准已经出现的下一调用。没有通过删除早返回来恢复旧编号授权。
- 兼容没有新字段的旧记录：用已记录的安全metadata补充绑定身份；普通旧pending/已确认记录仍可复用。若发现“当前pending编号也在历史批准记录中、当前call仍prepared且未确认”的旧卡死形态，用户提交旧编号时原子换出fresh pending编号，状态仍为`awaiting_confirmation`，返回新编号后必须再次明确确认。此过程不外呼、不自动授权，旧编号随后重发不再改变新pending。

## 绿测及边界

- `go test ./internal/novelcreation -run 'TestAgentConfirmationGeneration|TestAgentSecondaryConfirmation' -count=1`：通过，0.050s。
- `go test ./internal/novelcreation ./internal/prompt/novelcreation -count=1`：完整通过，0.247s / 0.020s。
- `go test ./internal/httpapi -run 'TestNovelConfirmationHTTP|TestNovelSavedChannelCapabilities|TestNovelResolver|TestNovelIntegrationSampling' -count=1`：通过，0.194s。

新增测试同时证明：A→B→A当前A2可确认续行；旧A1/B1和随后迟到的A2均不能批准下一调用；当前确认重复请求与重启后仍只实际发送一次；待确认先落盘、检查点状态保存故障后重启再次观察编号不变；无新代次字段的真实旧确定性编号卡死seed可安全换号，仍需要fresh确认。A/B切换期间外呼数保持1；有效A2确认后仅增加到2。既有late-token、蓝图、repair、自动任务计数与真实BASE零新增外呼恢复测试在完整domain套件中继续通过。

文件仅：`worker.go`、`worker_store.go`、`model_confirmation.go`、`confirmation_generation_test.go`和本报告。确认HTTP/API字段不变，Task7继续提交当前`confirmationId`；返回仍待确认时读取返回运行的新编号即可。

没有临时目录删除尝试、真实提供方调用、4001重启、推送或部署；没有扩大权限入口或其他功能。仅本轮文件提交，提交SHA由最终交接返回。
