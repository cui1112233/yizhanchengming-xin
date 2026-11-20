# 01 版本与证据规则

当前前端：`F:/软件/洛水2026/洛水2026/dist/assets/index-DgTjggsH.js`，SHA256 `7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6`。

当前原生：`F:/软件/洛水2026/洛水2026/modules/novels.pyd`，SHA256 `b5b22086c64851e0e94fc70a509cddbaf4ec144ac5beedc64e3a9a8caa40e57d`。

前端经过 Babel 仅解析、格式化，没有执行目标程序；文件头记录原 UTF16 范围。行号指恢复片段行号，不是原作者工程行号。

原生证据区分 IDA 伪代码、旧有 Ghidra 伪代码、常量。常量出现不证明调用发生；静态事件不证明按钮可用；文档覆盖不证明 CM 已对齐。

旧前端 SHA `5278a6d1b525152a6e9542c9d245bd90c658b84db3d94be9198ee2ac98306d4a` 只作历史参考，不能替代当前前端。

## 当前组件清单

| 组件 | 恢复文件 | 原 UTF16 范围 | SHA256 |
|---|---|---|---|
| NovelCreateIcon | [NovelCreateIcon.readable.js:1](source/frontend/NovelCreateIcon.readable.js#L1) | [8506080, 8512308] | `2be65998791f00af1b77b2c69b8ff1a5bf53face684e1bc12eb92d0592724f71` |
| NovelStudioSidebar | [NovelStudioSidebar.readable.js:1](source/frontend/NovelStudioSidebar.readable.js#L1) | [8542215, 8550130] | `9c18f2fed68af50d9b513fcd9c9370f96b6b361ad7b849a492e28d8a5ff418b0` |
| NovelConfirmCard | [NovelConfirmCard.readable.js:1](source/frontend/NovelConfirmCard.readable.js#L1) | [8550523, 8552978] | `b31b63a49ff35288edf2fab89b145611dcdeceb557bd7d67e6a516c2c9b17ad9` |
| NovelCoverPreview | [NovelCoverPreview.readable.js:1](source/frontend/NovelCoverPreview.readable.js#L1) | [8573753, 8575641] | `e53aeaef80441a23ade1b41df219d9fdc2d5e98d55c9cc1e5c4a0b3d2c58fc7b` |
| NovelToolSteps | [NovelToolSteps.readable.js:1](source/frontend/NovelToolSteps.readable.js#L1) | [8619062, 8625267] | `b9fff331a5454865dd946500aae15ea63447bf613637f0918df68dd884913880` |
| NovelThinkingBlock | [NovelThinkingBlock.readable.js:1](source/frontend/NovelThinkingBlock.readable.js#L1) | [8625439, 8626316] | `38a525599d765b2edaba753a4c46af96d7cbda4332e2b8e7951b5650abce9484` |
| NovelStreamingIndicator | [NovelStreamingIndicator.readable.js:1](source/frontend/NovelStreamingIndicator.readable.js#L1) | [8626359, 8626704] | `16a717fb61c15567079da1abbf45b098ce44f4dde11f755b663d2a22c441e071` |
| NovelQuickActions | [NovelQuickActions.readable.js:1](source/frontend/NovelQuickActions.readable.js#L1) | [8626774, 8627496] | `85a9257c134a4759e3cb6fce947e5d5a8c3d27049eb1c94a59012ab6037212b4` |
| NovelStudioChat | [NovelStudioChat.readable.js:1](source/frontend/NovelStudioChat.readable.js#L1) | [8629071, 8650803] | `51fdacdc6b113b35494270077376a06e056090800b0358607d1dda88ad23284a` |
| NovelPanelArtifactView | [NovelPanelArtifactView.readable.js:1](source/frontend/NovelPanelArtifactView.readable.js#L1) | [8651068, 8652874] | `261b1e20575a37a6042ade6336e77b304f74248d14d334f4c84256eb840fc8a0` |
| NovelStudioBookPanel | [NovelStudioBookPanel.readable.js:1](source/frontend/NovelStudioBookPanel.readable.js#L1) | [8683553, 8691226] | `51188e2d5a0f6293de371ddc98f6be4cc96b32f7e988e79bd3c05f008adc193e` |
| NovelStudioShortPanel | [NovelStudioShortPanel.readable.js:1](source/frontend/NovelStudioShortPanel.readable.js#L1) | [8691913, 8697920] | `a9ed21ad6f8e734a776afdaa9888bdbd1b55664f16acd0f584f0cf7fb7dd8085` |
| NovelStylePanel | [NovelStylePanel.readable.js:1](source/frontend/NovelStylePanel.readable.js#L1) | [8698363, 8704788] | `dd26414007e8102269542588154d62333c28aafa030db67d9b1efc71f71cae5e` |
| NovelAgentSettings | [NovelAgentSettings.readable.js:1](source/frontend/NovelAgentSettings.readable.js#L1) | [8744349, 8756515] | `e7e5b76ee59a26a294e55d98301beea0a95c5dd48f1a1834d5524f8094b78ad3` |
| NovelImportPage | [NovelImportPage.readable.js:1](source/frontend/NovelImportPage.readable.js#L1) | [8809779, 8830298] | `03bf6417209734c8a3df4932b958932f2007d6949ca8c02fb5eb2cfb0b3ef0e3` |
| NovelImportModal | [NovelImportModal.readable.js:1](source/frontend/NovelImportModal.readable.js#L1) | [8830352, 8831305] | `9d429c0f33b4786b33da0a55a2f310288c4d763d6e7f24b5a366a9fe5e0fb5ff` |
| NovelImitationShortDialog | [NovelImitationShortDialog.readable.js:1](source/frontend/NovelImitationShortDialog.readable.js#L1) | [8832348, 8840980] | `6b0d791550053430d284b6be4f31e3f5554eba0e789b892e342da3b3f8175b78` |
| NovelBatchDialog | [NovelBatchDialog.readable.js:1](source/frontend/NovelBatchDialog.readable.js#L1) | [8841139, 8844178] | `08118acd6c7465ccbdf853e4ba67ddd61921c8b613e994d02470d8870c70eed3` |
| NovelSchedulerPanel | [NovelSchedulerPanel.readable.js:1](source/frontend/NovelSchedulerPanel.readable.js#L1) | [8844545, 8849386] | `922344e0d84b5245d2ebced6dc0798c6e3ec74a5de30234429f4cea798c9fc31` |
| NovelHooksPanel | [NovelHooksPanel.readable.js:1](source/frontend/NovelHooksPanel.readable.js#L1) | [8849652, 8851209] | `0e498527ee940af7409ccdb376567aa895c2c9a04331930a4da40b399551870f` |
| NovelDashboard | [NovelDashboard.readable.js:1](source/frontend/NovelDashboard.readable.js#L1) | [8851662, 8853325] | `f713313ba854a9b2cdf7a571337ad203fd9ea47d23d6bb6dff40791c8775cd8e` |
| NovelCheckpointCard | [NovelCheckpointCard.readable.js:1](source/frontend/NovelCheckpointCard.readable.js#L1) | [9221037, 9221822] | `50bc411dc383aa5c919570974fcc46396a514fd584ddbaa26d9e8fd5b11b1528` |
| NovelStudioPage | [NovelStudioPage.readable.js:1](source/frontend/NovelStudioPage.readable.js#L1) | [9248264, 9294153] | `b789cdac45ec085fdb353a0ceadcaa94d54756069f4c2ba169a8f6a97fc2e6d3` |

## 原生分析位置

- IDA 数据库：`_local-archive-20261028/luoshui-ida-20261120/native/novels-sdk.pyd.i64`；工作副本 `C:/rs-ida/novels-sdk.pyd.i64`。本轮新增 IDA 导出见 `source/ida/`。
- 已解码常量：`docs/reverse-engineering/luoshui-novel-native-20261119/constants/`（95模块；完整常量不等于完整函数体）。
- 原生函数绑定：同目录 `function-bindings.json`；已有函数体覆盖见 `native-index.json`。
- 既有短篇完整体/Listing：`docs/verification/novel-full-draft-20261120/`；生产确认与终审：`docs/verification/novel-rea-followup-20261120/`。
- 证据文件逐个哈希见 [evidence-index.json](source/evidence-index.json)。
