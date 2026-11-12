# AI-CUT 页面宠物服务与功能对照实施

目标以已批准用户要求为准，设计见同日期pet-parity-design。按可独立验收的任务推进，不把历史清单当完成证明。

## A 重新核实洛水与开源差异

只读核查本机洛水前端包，REA分析精确组件/片段并记录哈希与覆盖；访问OpenCut、OpenChatCut、Kdenlive/Shotcut等原仓库/官方资料。更新`docs/audit/2026-11-13-aicut-current-parity.md`：截图按钮映射、当前已完成与部分、仍缺失、已排除；每一条指出真实处理函数和验收方法。不能从字符串存在推断成功。把可立即独立修复的虚入口反馈执行者。

## B 宠物页面服务宿主

新增`frontend/src/shared/pet/pageService.js`及测试，服务接口`registerPetPageService({id,scope,label,render}) -> unregister`、`getPetPageService()`、`subscribePetPageService(listener)`、`openPetPageService(id)`；注册器只允许当前页面自身安装/卸载，旧cleanup不得清空新服务，通知不泄露模型/密钥。

修改`CmPenguinCompanion.jsx`和宠物CSS，在当前专属服务存在时显示其真实面板，保留关闭与宠物操作；旧通用问答不能吞掉剪辑指令。面板关闭后保留挂载状态以保持运行事件，离开页面销毁。组件变化与API无关，任何模型操作沿用后端已有鉴权。

## C 工具栏与无效入口核查

检查`TimelineToolbar.jsx`、`timelineTools.js`与相应测试。提供可复用主工具栏动作描述器，主执行者在AICutPage接入。只输出真实已绑定命令，考虑选中/锁定/异步处理中禁用；把剪映截图位置映射到现有命令，补缺失的直接入口与明确文字提示。新增动作如需后端变更先报告，不假设已接入。

## D 页面接入与运行验收

主执行者在AICutPage通过页面服务注册器暴露现有AICutAgentPanel（相同project/timeline/revision/model/assets/callback）。增加独立宠物宿主小组件复用账号、宠物显示配置；原按钮改为打开宠物当前服务，避免两套Agent。其他页面的CM bridge继续正常。工程变化/时间线变化必须清理旧注册并更新scope；自动成片忙碌与未保存timeline校验保留。

逐任务先测试失败后实现，复查功能与代码质量，再整体检查；浏览器实测宠物进入剪辑服务、切换页面恢复相应能力、实际工具不会仅返回成功文字、截图工具栏操作保存/撤销。构建React和Goembed、验证health/runtime-build-info；仅提交本轮路径，记录仍未实现的差异。
