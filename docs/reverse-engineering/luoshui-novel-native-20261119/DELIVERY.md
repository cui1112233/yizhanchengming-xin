# 本轮实际交付与待办

产品提交a5383e1f，已并入主开发分支v88。没有推送或发布公网。

## 已完成

1. 原文件副本：`F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-native-20261119/novels-b5b22086.pyd`，SHA见manifest。当前原包路径未改动。
2. `constants/`95个模块、12662项顶层常量完整解析；`readable/`含原提示词片段和原字节范围。不是95个模块完整Python源码。
3. REA真实调用open_binary、search_strings、xrefs、analyze_function、export_evidence_bundle。原生证据保存在同archive的rea-native-module-evidence.json、rea-native-chain-evidence.json。回包超限均使用完整导出，不将截断文本当分析结果。
4. 13份REA函数证据渲染及离线Ghidra选择函数补证；773条静态构造绑定覆盖82模块，24处参数未知。离线全局分析180秒限时，不宣称完整调用图；fallback文件明确不是REA签发的证据。
5. SPEC记录写作组装、输出区块、审稿循环、局部/整章修订及六入口的输入、输出、边界、失败和CM差距。
6. 首处基于原函数的纠错已实施：post-write硬错误并入每次审核，error提升critical、passed=false、source=post_write，不修改模型分数。修订能收到这些问题；警告适配既有CM审核schema为medium，不作为硬阻塞。
7. 新任务冻结review_cycle_version=1；无该字段的历史任务维持原执行行为。关闭后写校验不合并、修订预算0时不自动多调用并保留warning结果。

## 验证与运行

- TestNativePostWriteErrorParticipatesInReviewCycle：模型92分通过但本地4硬错误→实际进入修订→复审→保存干净正文，测试模型8次请求。
- TestLegacyFrozenRunKeepsReportingOnlyPostWrite：旧任务保持6次请求和旧行为。
- TestNativePostWriteSwitchAndZeroRevisionBudget：关闭不拦截，预算0不超限。
- TestNativePostWriteReviewRealMySQL：隔离MySQL8.4实测通过，实际合并后的首轮审核passed=false持久化，最终正文已修改，57.77秒，0真实收费调用。
- Go novelcreation/prompt整包通过，HTTP小说相关定向测试通过，Go嵌入构建通过。
- 4001运行a5383e1f，入口`/assets/user-DMxUgrd5.js`，PID61128（观察时），health正常、小说页200。二进制archive/qiantie-native-review.exe，清单runtime-manifest.json。进程可能被其他并行工作更新，后续操作先重新检查。

## 没完成的明确登记

- 全部原生提示词动态分支的组装与迁移。
- 原版输出区块解析、PATCHES定位/冲突/撤销及整章修订完整链路。
- AI味/敏感检查的原版合并、篇幅择优、无净改善提前退出等完整审稿细节。
- 六入口的全局Agent继承、原版状态/记忆/伏笔全部文件与检索链路。
- 原/CM相同输入的真实生成质量对比及所有UI逐项验收。

以上不能因为本轮一个纠错通过而标为全部对齐。后续逐项按WORKLOG登记，禁止自行补写缺少依据的业务规则后声称复刻。

## 第二步交付补充（2026-11-19）

已接入新建长篇任务的PATCHES/REVISED_CONTENT分段修订。后端文件：native_revision.go；后端prompt文件：prompt/novelcreation/native_revision.go。提示词原始片段83、85来自固定目标常量，适配说明与剩余差异见PATCH-SPEC.md。
测试：小说整包、prompt整包、小说HTTP定向通过；实际隔离MySQL验证补丁→复审→正文及原始修订保存→完成任务重复运行不再请求模型。未消耗真实模型额度。
剩余事项中的“原版输出区块/PATCHES”现在仅局部及整章正文分段这一部分已接入；完整强制路由、原版状态回写、撤销路径和其它修订协议仍不能标记全部一致。
