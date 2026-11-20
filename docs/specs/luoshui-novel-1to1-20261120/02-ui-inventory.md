# 02 页面、事件与界面文案清单

该清单按当前安装包23个Novel组件提取。440条记录包含点击、更新事件、禁用条件、占位文案等，**不是440个独立功能或按钮**。事件条目只登记表达式，未冒称已完成调用链核验。

## 组件与导航事件

| 组件 | 发出事件 | 行号 |
|---|---|---|
| NovelStudioSidebar | `create-session` | [NovelStudioSidebar.readable.js:41](source/frontend/NovelStudioSidebar.readable.js#L41) |
| NovelStudioSidebar | `open-session` | [NovelStudioSidebar.readable.js:41](source/frontend/NovelStudioSidebar.readable.js#L41) |
| NovelStudioSidebar | `open-book` | [NovelStudioSidebar.readable.js:41](source/frontend/NovelStudioSidebar.readable.js#L41) |
| NovelStudioSidebar | `toggle-book` | [NovelStudioSidebar.readable.js:41](source/frontend/NovelStudioSidebar.readable.js#L41) |
| NovelStudioSidebar | `create-book-session` | [NovelStudioSidebar.readable.js:41](source/frontend/NovelStudioSidebar.readable.js#L41) |
| NovelStudioSidebar | `rename-session` | [NovelStudioSidebar.readable.js:41](source/frontend/NovelStudioSidebar.readable.js#L41) |
| NovelStudioSidebar | `delete-session` | [NovelStudioSidebar.readable.js:41](source/frontend/NovelStudioSidebar.readable.js#L41) |
| NovelStudioSidebar | `rename-book` | [NovelStudioSidebar.readable.js:41](source/frontend/NovelStudioSidebar.readable.js#L41) |
| NovelStudioSidebar | `delete-book` | [NovelStudioSidebar.readable.js:41](source/frontend/NovelStudioSidebar.readable.js#L41) |
| NovelStudioSidebar | `export-book` | [NovelStudioSidebar.readable.js:41](source/frontend/NovelStudioSidebar.readable.js#L41) |
| NovelStudioSidebar | `launch-mode` | [NovelStudioSidebar.readable.js:41](source/frontend/NovelStudioSidebar.readable.js#L41) |
| NovelStudioSidebar | `switch-workflow` | [NovelStudioSidebar.readable.js:41](source/frontend/NovelStudioSidebar.readable.js#L41) |
| NovelStudioSidebar | `open-import` | [NovelStudioSidebar.readable.js:41](source/frontend/NovelStudioSidebar.readable.js#L41) |
| NovelStudioSidebar | `open-style` | [NovelStudioSidebar.readable.js:41](source/frontend/NovelStudioSidebar.readable.js#L41) |
| NovelStudioSidebar | `open-import-modal` | [NovelStudioSidebar.readable.js:41](source/frontend/NovelStudioSidebar.readable.js#L41) |
| NovelStudioSidebar | `open-imitation-short` | [NovelStudioSidebar.readable.js:41](source/frontend/NovelStudioSidebar.readable.js#L41) |
| NovelStudioSidebar | `open-agent-settings` | [NovelStudioSidebar.readable.js:41](source/frontend/NovelStudioSidebar.readable.js#L41) |
| NovelStudioSidebar | `open-scheduler` | [NovelStudioSidebar.readable.js:41](source/frontend/NovelStudioSidebar.readable.js#L41) |
| NovelStudioSidebar | `open-dashboard` | [NovelStudioSidebar.readable.js:41](source/frontend/NovelStudioSidebar.readable.js#L41) |
| NovelStudioSidebar | `open-memo` | [NovelStudioSidebar.readable.js:41](source/frontend/NovelStudioSidebar.readable.js#L41) |
| NovelConfirmCard | `confirm` | [NovelConfirmCard.readable.js:17](source/frontend/NovelConfirmCard.readable.js#L17) |
| NovelConfirmCard | `dismiss` | [NovelConfirmCard.readable.js:17](source/frontend/NovelConfirmCard.readable.js#L17) |
| NovelToolSteps | `retry` | [NovelToolSteps.readable.js:41](source/frontend/NovelToolSteps.readable.js#L41) |
| NovelToolSteps | `confirm` | [NovelToolSteps.readable.js:41](source/frontend/NovelToolSteps.readable.js#L41) |
| NovelToolSteps | `dismiss` | [NovelToolSteps.readable.js:41](source/frontend/NovelToolSteps.readable.js#L41) |
| NovelThinkingBlock | `update:expanded` | [NovelThinkingBlock.readable.js:24](source/frontend/NovelThinkingBlock.readable.js#L24) |
| NovelQuickActions | `write-next` | [NovelQuickActions.readable.js:11](source/frontend/NovelQuickActions.readable.js#L11) |
| NovelQuickActions | `audit` | [NovelQuickActions.readable.js:11](source/frontend/NovelQuickActions.readable.js#L11) |
| NovelQuickActions | `revise` | [NovelQuickActions.readable.js:11](source/frontend/NovelQuickActions.readable.js#L11) |
| NovelQuickActions | `write-batch` | [NovelQuickActions.readable.js:11](source/frontend/NovelQuickActions.readable.js#L11) |
| NovelStudioChat | `send` | [NovelStudioChat.readable.js:48](source/frontend/NovelStudioChat.readable.js#L48) |
| NovelStudioChat | `confirm-action` | [NovelStudioChat.readable.js:48](source/frontend/NovelStudioChat.readable.js#L48) |
| NovelStudioChat | `dismiss-action` | [NovelStudioChat.readable.js:48](source/frontend/NovelStudioChat.readable.js#L48) |
| NovelStudioChat | `cancel` | [NovelStudioChat.readable.js:48](source/frontend/NovelStudioChat.readable.js#L48) |
| NovelStudioChat | `quick-write-next` | [NovelStudioChat.readable.js:48](source/frontend/NovelStudioChat.readable.js#L48) |
| NovelStudioChat | `quick-write-batch` | [NovelStudioChat.readable.js:48](source/frontend/NovelStudioChat.readable.js#L48) |
| NovelStudioChat | `quick-audit` | [NovelStudioChat.readable.js:48](source/frontend/NovelStudioChat.readable.js#L48) |
| NovelStudioChat | `quick-revise` | [NovelStudioChat.readable.js:48](source/frontend/NovelStudioChat.readable.js#L48) |
| NovelStudioChat | `delete-message` | [NovelStudioChat.readable.js:48](source/frontend/NovelStudioChat.readable.js#L48) |
| NovelStudioChat | `archive-message` | [NovelStudioChat.readable.js:48](source/frontend/NovelStudioChat.readable.js#L48) |
| NovelStudioChat | `update:selectedModel` | [NovelStudioChat.readable.js:48](source/frontend/NovelStudioChat.readable.js#L48) |
| NovelStudioChat | `update:thinkingAutoExpand` | [NovelStudioChat.readable.js:48](source/frontend/NovelStudioChat.readable.js#L48) |
| NovelStudioChat | `clear-context` | [NovelStudioChat.readable.js:48](source/frontend/NovelStudioChat.readable.js#L48) |
| NovelStudioChat | `retry-action` | [NovelStudioChat.readable.js:48](source/frontend/NovelStudioChat.readable.js#L48) |
| NovelPanelArtifactView | `back` | [NovelPanelArtifactView.readable.js:21](source/frontend/NovelPanelArtifactView.readable.js#L21) |
| NovelPanelArtifactView | `save` | [NovelPanelArtifactView.readable.js:21](source/frontend/NovelPanelArtifactView.readable.js#L21) |
| NovelStudioBookPanel | `update:collapsed` | [NovelStudioBookPanel.readable.js:21](source/frontend/NovelStudioBookPanel.readable.js#L21) |
| NovelStudioBookPanel | `saved` | [NovelStudioBookPanel.readable.js:21](source/frontend/NovelStudioBookPanel.readable.js#L21) |
| NovelStudioShortPanel | `update:collapsed` | [NovelStudioShortPanel.readable.js:21](source/frontend/NovelStudioShortPanel.readable.js#L21) |
| NovelStudioShortPanel | `saved` | [NovelStudioShortPanel.readable.js:21](source/frontend/NovelStudioShortPanel.readable.js#L21) |
| NovelStylePanel | `update:selectedStyleIds` | [NovelStylePanel.readable.js:17](source/frontend/NovelStylePanel.readable.js#L17) |
| NovelStylePanel | `back` | [NovelStylePanel.readable.js:17](source/frontend/NovelStylePanel.readable.js#L17) |
| NovelStylePanel | `refresh` | [NovelStylePanel.readable.js:17](source/frontend/NovelStylePanel.readable.js#L17) |
| NovelStylePanel | `applied` | [NovelStylePanel.readable.js:17](source/frontend/NovelStylePanel.readable.js#L17) |
| NovelAgentSettings | `back` | [NovelAgentSettings.readable.js:13](source/frontend/NovelAgentSettings.readable.js#L13) |
| NovelImportPage | `back` | [NovelImportPage.readable.js:29](source/frontend/NovelImportPage.readable.js#L29) |
| NovelImportPage | `close` | [NovelImportPage.readable.js:29](source/frontend/NovelImportPage.readable.js#L29) |
| NovelImportPage | `created` | [NovelImportPage.readable.js:29](source/frontend/NovelImportPage.readable.js#L29) |
| NovelImportPage | `update:tab` | [NovelImportPage.readable.js:29](source/frontend/NovelImportPage.readable.js#L29) |
| NovelImportPage | `submit-start` | [NovelImportPage.readable.js:29](source/frontend/NovelImportPage.readable.js#L29) |
| NovelImportModal | `update:show` | [NovelImportModal.readable.js:26](source/frontend/NovelImportModal.readable.js#L26) |
| NovelImportModal | `update:tab` | [NovelImportModal.readable.js:26](source/frontend/NovelImportModal.readable.js#L26) |
| NovelImportModal | `created` | [NovelImportModal.readable.js:26](source/frontend/NovelImportModal.readable.js#L26) |
| NovelImportModal | `submit-start` | [NovelImportModal.readable.js:26](source/frontend/NovelImportModal.readable.js#L26) |
| NovelImitationShortDialog | `update:show` | [NovelImitationShortDialog.readable.js:22](source/frontend/NovelImitationShortDialog.readable.js#L22) |
| NovelImitationShortDialog | `submit` | [NovelImitationShortDialog.readable.js:22](source/frontend/NovelImitationShortDialog.readable.js#L22) |
| NovelBatchDialog | `update:show` | [NovelBatchDialog.readable.js:13](source/frontend/NovelBatchDialog.readable.js#L13) |
| NovelBatchDialog | `done` | [NovelBatchDialog.readable.js:13](source/frontend/NovelBatchDialog.readable.js#L13) |
| NovelSchedulerPanel | `back` | [NovelSchedulerPanel.readable.js:12](source/frontend/NovelSchedulerPanel.readable.js#L12) |
| NovelCheckpointCard | `confirm` | [NovelCheckpointCard.readable.js:13](source/frontend/NovelCheckpointCard.readable.js#L13) |
| NovelCheckpointCard | `dismiss` | [NovelCheckpointCard.readable.js:13](source/frontend/NovelCheckpointCard.readable.js#L13) |

## 控件和条件逐项登记

| 编号 | 组件/属性 | 表达式 | 来源 | 核验状态 |
|---|---|---|---|---|
| UI-0001 | NovelStudioSidebar / `onClick` | `e => function (e) {<br>  if ("imitation-short" === e.key) return a("switch-workflow", "imitation-short"), void a("open-imitation-short");<br>  a("switch-workflow", e.key), "short" !== e.key && "book-create" !== e.key && a("open-import", e.key);<br>}(t)` | [NovelStudioSidebar.readable.js:237](source/frontend/NovelStudioSidebar.readable.js#L237) | 静态已提取；运行未测 |
| UI-0002 | NovelStudioSidebar / `onClick` | `n[0] \|\| (n[0] = e => o.value = !o.value)` | [NovelStudioSidebar.readable.js:247](source/frontend/NovelStudioSidebar.readable.js#L247) | 静态已提取；运行未测 |
| UI-0003 | NovelStudioSidebar / `onClick` | `e => a("toggle-book", t.id)` | [NovelStudioSidebar.readable.js:259](source/frontend/NovelStudioSidebar.readable.js#L259) | 静态已提取；运行未测 |
| UI-0004 | NovelStudioSidebar / `onClick` | `e => a("open-book", t.id)` | [NovelStudioSidebar.readable.js:270](source/frontend/NovelStudioSidebar.readable.js#L270) | 静态已提取；运行未测 |
| UI-0005 | NovelStudioSidebar / `title` | `"重命名"` | [NovelStudioSidebar.readable.js:276](source/frontend/NovelStudioSidebar.readable.js#L276) | 静态已提取；运行未测 |
| UI-0006 | NovelStudioSidebar / `onClick` | `Zi(e => a("rename-book", t), ["stop"])` | [NovelStudioSidebar.readable.js:277](source/frontend/NovelStudioSidebar.readable.js#L277) | 静态已提取；运行未测 |
| UI-0007 | NovelStudioSidebar / `title` | `"导出"` | [NovelStudioSidebar.readable.js:283](source/frontend/NovelStudioSidebar.readable.js#L283) | 静态已提取；运行未测 |
| UI-0008 | NovelStudioSidebar / `onClick` | `Zi(e => a("export-book", t), ["stop"])` | [NovelStudioSidebar.readable.js:284](source/frontend/NovelStudioSidebar.readable.js#L284) | 静态已提取；运行未测 |
| UI-0009 | NovelStudioSidebar / `title` | `"删除"` | [NovelStudioSidebar.readable.js:290](source/frontend/NovelStudioSidebar.readable.js#L290) | 静态已提取；运行未测 |
| UI-0010 | NovelStudioSidebar / `onClick` | `Zi(e => a("delete-book", t), ["stop"])` | [NovelStudioSidebar.readable.js:291](source/frontend/NovelStudioSidebar.readable.js#L291) | 静态已提取；运行未测 |
| UI-0011 | NovelStudioSidebar / `onClick` | `e => a("open-session", t.id)` | [NovelStudioSidebar.readable.js:302](source/frontend/NovelStudioSidebar.readable.js#L302) | 静态已提取；运行未测 |
| UI-0012 | NovelStudioSidebar / `onClick` | `e => a("create-book-session", t.id)` | [NovelStudioSidebar.readable.js:308](source/frontend/NovelStudioSidebar.readable.js#L308) | 静态已提取；运行未测 |
| UI-0013 | NovelStudioSidebar / `onClick` | `n[1] \|\| (n[1] = e => r.value = !r.value)` | [NovelStudioSidebar.readable.js:315](source/frontend/NovelStudioSidebar.readable.js#L315) | 静态已提取；运行未测 |
| UI-0014 | NovelStudioSidebar / `onClick` | `e => a("open-session", t.id)` | [NovelStudioSidebar.readable.js:329](source/frontend/NovelStudioSidebar.readable.js#L329) | 静态已提取；运行未测 |
| UI-0015 | NovelStudioSidebar / `onClick` | `Zi(e => a("rename-session", t), ["stop"])` | [NovelStudioSidebar.readable.js:335](source/frontend/NovelStudioSidebar.readable.js#L335) | 静态已提取；运行未测 |
| UI-0016 | NovelStudioSidebar / `onClick` | `Zi(e => a("delete-session", t), ["stop"])` | [NovelStudioSidebar.readable.js:341](source/frontend/NovelStudioSidebar.readable.js#L341) | 静态已提取；运行未测 |
| UI-0017 | NovelStudioSidebar / `onClick` | `n[2] \|\| (n[2] = t => a("create-session", e.activeWorkflow))` | [NovelStudioSidebar.readable.js:347](source/frontend/NovelStudioSidebar.readable.js#L347) | 静态已提取；运行未测 |
| UI-0018 | NovelStudioSidebar / `onClick` | `e => {<br>  var n;<br>  "style" === (n = t.key) ? a("open-style") : "import" === n ? a("open-import-modal") : "agent-settings" === n ? a("open-agent-settings") : "scheduler" === n ? a("open-scheduler") : "dashboard" === n ? a("open-dashboard") : "memo" === n && a("open-memo");<br>}` | [NovelStudioSidebar.readable.js:357](source/frontend/NovelStudioSidebar.readable.js#L357) | 静态已提取；运行未测 |
| UI-0019 | NovelConfirmCard / `disabled` | `e.loading \|\| e.streaming` | [NovelConfirmCard.readable.js:95](source/frontend/NovelConfirmCard.readable.js#L95) | 静态已提取；运行未测 |
| UI-0020 | NovelConfirmCard / `onClick` | `n[0] \|\| (n[0] = e => a("confirm"))` | [NovelConfirmCard.readable.js:96](source/frontend/NovelConfirmCard.readable.js#L96) | 静态已提取；运行未测 |
| UI-0021 | NovelConfirmCard / `disabled` | `e.loading \|\| e.streaming` | [NovelConfirmCard.readable.js:100](source/frontend/NovelConfirmCard.readable.js#L100) | 静态已提取；运行未测 |
| UI-0022 | NovelConfirmCard / `onClick` | `n[1] \|\| (n[1] = e => a("dismiss"))` | [NovelConfirmCard.readable.js:101](source/frontend/NovelConfirmCard.readable.js#L101) | 静态已提取；运行未测 |
| UI-0023 | NovelToolSteps / `onClick` | `n[0] \|\| (n[0] = e => o.value = !o.value)` | [NovelToolSteps.readable.js:152](source/frontend/NovelToolSteps.readable.js#L152) | 静态已提取；运行未测 |
| UI-0024 | NovelToolSteps / `onConfirm` | `n[1] \|\| (n[1] = e => a("confirm"))` | [NovelToolSteps.readable.js:180](source/frontend/NovelToolSteps.readable.js#L180) | 静态已提取；运行未测 |
| UI-0025 | NovelToolSteps / `onDismiss` | `n[2] \|\| (n[2] = e => a("dismiss"))` | [NovelToolSteps.readable.js:181](source/frontend/NovelToolSteps.readable.js#L181) | 静态已提取；运行未测 |
| UI-0026 | NovelToolSteps / `onClick` | `n[3] \|\| (n[3] = Zi(e => a("retry"), ["stop"]))` | [NovelToolSteps.readable.js:202](source/frontend/NovelToolSteps.readable.js#L202) | 静态已提取；运行未测 |
| UI-0027 | NovelToolSteps / `onClick` | `n[4] \|\| (n[4] = Zi(e => a("retry"), ["stop"]))` | [NovelToolSteps.readable.js:206](source/frontend/NovelToolSteps.readable.js#L206) | 静态已提取；运行未测 |
| UI-0028 | NovelThinkingBlock / `onClick` | `i` | [NovelThinkingBlock.readable.js:42](source/frontend/NovelThinkingBlock.readable.js#L42) | 静态已提取；运行未测 |
| UI-0029 | NovelQuickActions / `disabled` | `Boolean` | [NovelQuickActions.readable.js:7](source/frontend/NovelQuickActions.readable.js#L7) | 静态已提取；运行未测 |
| UI-0030 | NovelQuickActions / `disabled` | `e.disabled \|\| e.loading` | [NovelQuickActions.readable.js:43](source/frontend/NovelQuickActions.readable.js#L43) | 静态已提取；运行未测 |
| UI-0031 | NovelQuickActions / `onClick` | `e => {<br>  var a;<br>  "write-next" === (a = t.key) ? n("write-next") : "write-batch" === a ? n("write-batch") : "audit" === a ? n("audit") : "revise" === a && n("revise");<br>}` | [NovelQuickActions.readable.js:44](source/frontend/NovelQuickActions.readable.js#L44) | 静态已提取；运行未测 |
| UI-0032 | NovelStudioChat / `onClick` | `e => pe(t, o)` | [NovelStudioChat.readable.js:346](source/frontend/NovelStudioChat.readable.js#L346) | 静态已提取；运行未测 |
| UI-0033 | NovelStudioChat / `onClick` | `e => fe(t)` | [NovelStudioChat.readable.js:365](source/frontend/NovelStudioChat.readable.js#L365) | 静态已提取；运行未测 |
| UI-0034 | NovelStudioChat / `onPositiveClick` | `e => me(o, t)` | [NovelStudioChat.readable.js:375](source/frontend/NovelStudioChat.readable.js#L375) | 静态已提取；运行未测 |
| UI-0035 | NovelStudioChat / `onClick` | `n[0] \|\| (n[0] = Zi(() => {}, ["stop"]))` | [NovelStudioChat.readable.js:384](source/frontend/NovelStudioChat.readable.js#L384) | 静态已提取；运行未测 |
| UI-0036 | NovelStudioChat / `onUpdate:expanded` | `e => t.reasoningExpanded = e` | [NovelStudioChat.readable.js:406](source/frontend/NovelStudioChat.readable.js#L406) | 静态已提取；运行未测 |
| UI-0037 | NovelStudioChat / `onConfirm` | `n[1] \|\| (n[1] = e => a("confirm-action"))` | [NovelStudioChat.readable.js:416](source/frontend/NovelStudioChat.readable.js#L416) | 静态已提取；运行未测 |
| UI-0038 | NovelStudioChat / `onDismiss` | `n[2] \|\| (n[2] = e => a("dismiss-action"))` | [NovelStudioChat.readable.js:417](source/frontend/NovelStudioChat.readable.js#L417) | 静态已提取；运行未测 |
| UI-0039 | NovelStudioChat / `onClick` | `e => pe(t, o)` | [NovelStudioChat.readable.js:427](source/frontend/NovelStudioChat.readable.js#L427) | 静态已提取；运行未测 |
| UI-0040 | NovelStudioChat / `onClick` | `e => fe(t)` | [NovelStudioChat.readable.js:446](source/frontend/NovelStudioChat.readable.js#L446) | 静态已提取；运行未测 |
| UI-0041 | NovelStudioChat / `onPositiveClick` | `e => me(o, t)` | [NovelStudioChat.readable.js:456](source/frontend/NovelStudioChat.readable.js#L456) | 静态已提取；运行未测 |
| UI-0042 | NovelStudioChat / `onClick` | `n[3] \|\| (n[3] = Zi(() => {}, ["stop"]))` | [NovelStudioChat.readable.js:465](source/frontend/NovelStudioChat.readable.js#L465) | 静态已提取；运行未测 |
| UI-0043 | NovelStudioChat / `onUpdate:expanded` | `n[4] \|\| (n[4] = e => k.value = e)` | [NovelStudioChat.readable.js:489](source/frontend/NovelStudioChat.readable.js#L489) | 静态已提取；运行未测 |
| UI-0044 | NovelStudioChat / `onClick` | `he` | [NovelStudioChat.readable.js:504](source/frontend/NovelStudioChat.readable.js#L504) | 静态已提取；运行未测 |
| UI-0045 | NovelStudioChat / `onConfirm` | `n[5] \|\| (n[5] = e => a("confirm-action"))` | [NovelStudioChat.readable.js:525](source/frontend/NovelStudioChat.readable.js#L525) | 静态已提取；运行未测 |
| UI-0046 | NovelStudioChat / `onDismiss` | `n[6] \|\| (n[6] = e => a("dismiss-action"))` | [NovelStudioChat.readable.js:526](source/frontend/NovelStudioChat.readable.js#L526) | 静态已提取；运行未测 |
| UI-0047 | NovelStudioChat / `onMouseenter` | `D` | [NovelStudioChat.readable.js:534](source/frontend/NovelStudioChat.readable.js#L534) | 静态已提取；运行未测 |
| UI-0048 | NovelStudioChat / `onMouseleave` | `L` | [NovelStudioChat.readable.js:535](source/frontend/NovelStudioChat.readable.js#L535) | 静态已提取；运行未测 |
| UI-0049 | NovelStudioChat / `onMouseenter` | `D` | [NovelStudioChat.readable.js:542](source/frontend/NovelStudioChat.readable.js#L542) | 静态已提取；运行未测 |
| UI-0050 | NovelStudioChat / `onMouseleave` | `L` | [NovelStudioChat.readable.js:543](source/frontend/NovelStudioChat.readable.js#L543) | 静态已提取；运行未测 |
| UI-0051 | NovelStudioChat / `onMouseenter` | `t => P.value = e.id` | [NovelStudioChat.readable.js:556](source/frontend/NovelStudioChat.readable.js#L556) | 静态已提取；运行未测 |
| UI-0052 | NovelStudioChat / `onMouseleave` | `n[7] \|\| (n[7] = e => P.value = null)` | [NovelStudioChat.readable.js:557](source/frontend/NovelStudioChat.readable.js#L557) | 静态已提取；运行未测 |
| UI-0053 | NovelStudioChat / `onClick` | `t => function (e) {<br>  var t;<br>  const n = i.value;<br>  if (!n) return;<br>  const a = null == (t = n.$el) ? void 0 : t.querySelector(".n-scrollbar-container");<br>  if (!a) return;<br>  const o = a.querySelector(`.msg-row[data-msg-id="${e}"]`);<br>  if (!o) return;<br>  const r = o.offsetTop - a.clientHeight / 2 + o.offsetHeight / 2;<br>  n.scrollTo({<br>    top: Math.max(0, r),<br>    behavior: "smooth"<br>  }), o.classList.add("msg-nav-highlight"), setTimeout(() => o.classList.remove("msg-nav-highlight"), 1500), _.value = !1, P.value = null;<br>}(e.id)` | [NovelStudioChat.readable.js:558](source/frontend/NovelStudioChat.readable.js#L558) | 静态已提取；运行未测 |
| UI-0054 | NovelStudioChat / `disabled` | `e.isStreaming \|\| e.confirmLoading \|\| !!j.value` | [NovelStudioChat.readable.js:584](source/frontend/NovelStudioChat.readable.js#L584) | 静态已提取；运行未测 |
| UI-0055 | NovelStudioChat / `onWriteNext` | `n[8] \|\| (n[8] = e => a("quick-write-next"))` | [NovelStudioChat.readable.js:587](source/frontend/NovelStudioChat.readable.js#L587) | 静态已提取；运行未测 |
| UI-0056 | NovelStudioChat / `onWriteBatch` | `n[9] \|\| (n[9] = e => a("quick-write-batch"))` | [NovelStudioChat.readable.js:588](source/frontend/NovelStudioChat.readable.js#L588) | 静态已提取；运行未测 |
| UI-0057 | NovelStudioChat / `onAudit` | `n[10] \|\| (n[10] = e => a("quick-audit"))` | [NovelStudioChat.readable.js:589](source/frontend/NovelStudioChat.readable.js#L589) | 静态已提取；运行未测 |
| UI-0058 | NovelStudioChat / `onRevise` | `n[11] \|\| (n[11] = e => a("quick-revise"))` | [NovelStudioChat.readable.js:590](source/frontend/NovelStudioChat.readable.js#L590) | 静态已提取；运行未测 |
| UI-0059 | NovelStudioChat / `onUpdate:modelValue` | `n[12] \|\| (n[12] = e => l.value = e)` | [NovelStudioChat.readable.js:594](source/frontend/NovelStudioChat.readable.js#L594) | 静态已提取；运行未测 |
| UI-0060 | NovelStudioChat / `placeholder` | `oe.value` | [NovelStudioChat.readable.js:596](source/frontend/NovelStudioChat.readable.js#L596) | 静态已提取；运行未测 |
| UI-0061 | NovelStudioChat / `disabled` | `e.isStreaming` | [NovelStudioChat.readable.js:598](source/frontend/NovelStudioChat.readable.js#L598) | 静态已提取；运行未测 |
| UI-0062 | NovelStudioChat / `onKeydown` | `le` | [NovelStudioChat.readable.js:599](source/frontend/NovelStudioChat.readable.js#L599) | 静态已提取；运行未测 |
| UI-0063 | NovelStudioChat / `onCompositionstart` | `n[13] \|\| (n[13] = e => c.value = !0)` | [NovelStudioChat.readable.js:600](source/frontend/NovelStudioChat.readable.js#L600) | 静态已提取；运行未测 |
| UI-0064 | NovelStudioChat / `onCompositionend` | `n[14] \|\| (n[14] = e => c.value = !1)` | [NovelStudioChat.readable.js:601](source/frontend/NovelStudioChat.readable.js#L601) | 静态已提取；运行未测 |
| UI-0065 | NovelStudioChat / `onClick` | `ce` | [NovelStudioChat.readable.js:604](source/frontend/NovelStudioChat.readable.js#L604) | 静态已提取；运行未测 |
| UI-0066 | NovelStudioChat / `title` | `"展开思考"` | [NovelStudioChat.readable.js:620](source/frontend/NovelStudioChat.readable.js#L620) | 静态已提取；运行未测 |
| UI-0067 | NovelStudioChat / `onClick` | `n[15] \|\| (n[15] = e => f.value = !f.value)` | [NovelStudioChat.readable.js:621](source/frontend/NovelStudioChat.readable.js#L621) | 静态已提取；运行未测 |
| UI-0068 | NovelStudioChat / `onClick` | `n[16] \|\| (n[16] = t => {<br>  a("update:thinkingAutoExpand", !e.thinkingAutoExpand), f.value = !1;<br>})` | [NovelStudioChat.readable.js:642](source/frontend/NovelStudioChat.readable.js#L642) | 静态已提取；运行未测 |
| UI-0069 | NovelStudioChat / `onClick` | `n[17] \|\| (n[17] = e => f.value = !1)` | [NovelStudioChat.readable.js:650](source/frontend/NovelStudioChat.readable.js#L650) | 静态已提取；运行未测 |
| UI-0070 | NovelStudioChat / `onClick` | `n[18] \|\| (n[18] = e => h.value = !h.value)` | [NovelStudioChat.readable.js:657](source/frontend/NovelStudioChat.readable.js#L657) | 静态已提取；运行未测 |
| UI-0071 | NovelStudioChat / `onClick` | `ie` | [NovelStudioChat.readable.js:671](source/frontend/NovelStudioChat.readable.js#L671) | 静态已提取；运行未测 |
| UI-0072 | NovelStudioChat / `onClick` | `n[19] \|\| (n[19] = e => m.value = 1e5)` | [NovelStudioChat.readable.js:683](source/frontend/NovelStudioChat.readable.js#L683) | 静态已提取；运行未测 |
| UI-0073 | NovelStudioChat / `onClick` | `n[20] \|\| (n[20] = e => m.value = 2e5)` | [NovelStudioChat.readable.js:685](source/frontend/NovelStudioChat.readable.js#L685) | 静态已提取；运行未测 |
| UI-0074 | NovelStudioChat / `onClick` | `n[21] \|\| (n[21] = e => m.value = 3e5)` | [NovelStudioChat.readable.js:687](source/frontend/NovelStudioChat.readable.js#L687) | 静态已提取；运行未测 |
| UI-0075 | NovelStudioChat / `onClick` | `n[22] \|\| (n[22] = e => m.value = 4e5)` | [NovelStudioChat.readable.js:689](source/frontend/NovelStudioChat.readable.js#L689) | 静态已提取；运行未测 |
| UI-0076 | NovelStudioChat / `onClick` | `n[23] \|\| (n[23] = e => m.value = 1e6)` | [NovelStudioChat.readable.js:691](source/frontend/NovelStudioChat.readable.js#L691) | 静态已提取；运行未测 |
| UI-0077 | NovelStudioChat / `onClick` | `n[24] \|\| (n[24] = e => m.value = van)` | [NovelStudioChat.readable.js:693](source/frontend/NovelStudioChat.readable.js#L693) | 静态已提取；运行未测 |
| UI-0078 | NovelStudioChat / `onClick` | `n[25] \|\| (n[25] = e => h.value = !1)` | [NovelStudioChat.readable.js:700](source/frontend/NovelStudioChat.readable.js#L700) | 静态已提取；运行未测 |
| UI-0079 | NovelStudioChat / `onClick` | `se` | [NovelStudioChat.readable.js:707](source/frontend/NovelStudioChat.readable.js#L707) | 静态已提取；运行未测 |
| UI-0080 | NovelStudioChat / `onClick` | `n[26] \|\| (n[26] = e => a("cancel"))` | [NovelStudioChat.readable.js:718](source/frontend/NovelStudioChat.readable.js#L718) | 静态已提取；运行未测 |
| UI-0081 | NovelStudioChat / `disabled` | `!l.value.trim()` | [NovelStudioChat.readable.js:736](source/frontend/NovelStudioChat.readable.js#L736) | 静态已提取；运行未测 |
| UI-0082 | NovelStudioChat / `onClick` | `re` | [NovelStudioChat.readable.js:737](source/frontend/NovelStudioChat.readable.js#L737) | 静态已提取；运行未测 |
| UI-0083 | NovelStudioChat / `onUpdate:show` | `n[29] \|\| (n[29] = e => u.value = e)` | [NovelStudioChat.readable.js:743](source/frontend/NovelStudioChat.readable.js#L743) | 静态已提取；运行未测 |
| UI-0084 | NovelStudioChat / `onUpdate:modelValue` | `n[27] \|\| (n[27] = e => d.value = e)` | [NovelStudioChat.readable.js:759](source/frontend/NovelStudioChat.readable.js#L759) | 静态已提取；运行未测 |
| UI-0085 | NovelStudioChat / `placeholder` | `"搜索模型..."` | [NovelStudioChat.readable.js:761](source/frontend/NovelStudioChat.readable.js#L761) | 静态已提取；运行未测 |
| UI-0086 | NovelStudioChat / `onClick` | `n[28] \|\| (n[28] = e => u.value = !1)` | [NovelStudioChat.readable.js:766](source/frontend/NovelStudioChat.readable.js#L766) | 静态已提取；运行未测 |
| UI-0087 | NovelStudioChat / `onClick` | `t => {<br>  return n = e.key, void (v.value = v.value === n ? "" : n);<br>  var n;<br>}` | [NovelStudioChat.readable.js:777](source/frontend/NovelStudioChat.readable.js#L777) | 静态已提取；运行未测 |
| UI-0088 | NovelStudioChat / `onClick` | `e => {<br>  return n = t.value, a("update:selectedModel", n), void (u.value = !1);<br>  var n;<br>}` | [NovelStudioChat.readable.js:786](source/frontend/NovelStudioChat.readable.js#L786) | 静态已提取；运行未测 |
| UI-0089 | NovelPanelArtifactView / `title` | `{<br>  type: String,<br>  default: ""<br>}` | [NovelPanelArtifactView.readable.js:7](source/frontend/NovelPanelArtifactView.readable.js#L7) | 静态已提取；运行未测 |
| UI-0090 | NovelPanelArtifactView / `title` | `"返回"` | [NovelPanelArtifactView.readable.js:51](source/frontend/NovelPanelArtifactView.readable.js#L51) | 静态已提取；运行未测 |
| UI-0091 | NovelPanelArtifactView / `onClick` | `n[0] \|\| (n[0] = e => a("back"))` | [NovelPanelArtifactView.readable.js:52](source/frontend/NovelPanelArtifactView.readable.js#L52) | 静态已提取；运行未测 |
| UI-0092 | NovelPanelArtifactView / `title` | `"保存"` | [NovelPanelArtifactView.readable.js:63](source/frontend/NovelPanelArtifactView.readable.js#L63) | 静态已提取；运行未测 |
| UI-0093 | NovelPanelArtifactView / `disabled` | `e.saving` | [NovelPanelArtifactView.readable.js:64](source/frontend/NovelPanelArtifactView.readable.js#L64) | 静态已提取；运行未测 |
| UI-0094 | NovelPanelArtifactView / `onClick` | `s` | [NovelPanelArtifactView.readable.js:65](source/frontend/NovelPanelArtifactView.readable.js#L65) | 静态已提取；运行未测 |
| UI-0095 | NovelPanelArtifactView / `title` | `"取消"` | [NovelPanelArtifactView.readable.js:76](source/frontend/NovelPanelArtifactView.readable.js#L76) | 静态已提取；运行未测 |
| UI-0096 | NovelPanelArtifactView / `onClick` | `i` | [NovelPanelArtifactView.readable.js:77](source/frontend/NovelPanelArtifactView.readable.js#L77) | 静态已提取；运行未测 |
| UI-0097 | NovelPanelArtifactView / `title` | `"编辑"` | [NovelPanelArtifactView.readable.js:85](source/frontend/NovelPanelArtifactView.readable.js#L85) | 静态已提取；运行未测 |
| UI-0098 | NovelPanelArtifactView / `onClick` | `l` | [NovelPanelArtifactView.readable.js:86](source/frontend/NovelPanelArtifactView.readable.js#L86) | 静态已提取；运行未测 |
| UI-0099 | NovelPanelArtifactView / `onUpdate:modelValue` | `n[1] \|\| (n[1] = e => r.value = e)` | [NovelPanelArtifactView.readable.js:97](source/frontend/NovelPanelArtifactView.readable.js#L97) | 静态已提取；运行未测 |
| UI-0100 | NovelStudioBookPanel / `title` | `"故事基石 / 世界观"` | [NovelStudioBookPanel.readable.js:147](source/frontend/NovelStudioBookPanel.readable.js#L147) | 静态已提取；运行未测 |
| UI-0101 | NovelStudioBookPanel / `onClick` | `n[0] \|\| (n[0] = e => a("update:collapsed", !0))` | [NovelStudioBookPanel.readable.js:210](source/frontend/NovelStudioBookPanel.readable.js#L210) | 静态已提取；运行未测 |
| UI-0102 | NovelStudioBookPanel / `onClick` | `n[1] \|\| (n[1] = e => W("chapters"))` | [NovelStudioBookPanel.readable.js:217](source/frontend/NovelStudioBookPanel.readable.js#L217) | 静态已提取；运行未测 |
| UI-0103 | NovelStudioBookPanel / `onClick` | `t => function (e) {<br>  b({<br>    type: "chapter",<br>    chapterNumber: e.number,<br>    title: `第 ${e.number} 章 ${e.title}`<br>  });<br>}(e)` | [NovelStudioBookPanel.readable.js:225](source/frontend/NovelStudioBookPanel.readable.js#L225) | 静态已提取；运行未测 |
| UI-0104 | NovelStudioBookPanel / `title` | ``第 ${e.number} 章 ${e.title}`` | [NovelStudioBookPanel.readable.js:229](source/frontend/NovelStudioBookPanel.readable.js#L229) | 静态已提取；运行未测 |
| UI-0105 | NovelStudioBookPanel / `title` | `"编辑"` | [NovelStudioBookPanel.readable.js:238](source/frontend/NovelStudioBookPanel.readable.js#L238) | 静态已提取；运行未测 |
| UI-0106 | NovelStudioBookPanel / `onClick` | `Zi(t => function (e) {<br>  b({<br>    type: "chapter",<br>    chapterNumber: e.number,<br>    title: `第 ${e.number} 章 ${e.title}`,<br>    autoEdit: !0<br>  });<br>}(e), ["stop"])` | [NovelStudioBookPanel.readable.js:239](source/frontend/NovelStudioBookPanel.readable.js#L239) | 静态已提取；运行未测 |
| UI-0107 | NovelStudioBookPanel / `title` | ``第 ${e.number} 章 ${e.title}`` | [NovelStudioBookPanel.readable.js:243](source/frontend/NovelStudioBookPanel.readable.js#L243) | 静态已提取；运行未测 |
| UI-0108 | NovelStudioBookPanel / `onClick` | `n[2] \|\| (n[2] = e => W("characters"))` | [NovelStudioBookPanel.readable.js:253](source/frontend/NovelStudioBookPanel.readable.js#L253) | 静态已提取；运行未测 |
| UI-0109 | NovelStudioBookPanel / `onClick` | `t => function (e) {<br>  b({<br>    type: "character",<br>    key: e.path.replace(/\.md$/, ""),<br>    title: e.name<br>  });<br>}(e)` | [NovelStudioBookPanel.readable.js:262](source/frontend/NovelStudioBookPanel.readable.js#L262) | 静态已提取；运行未测 |
| UI-0110 | NovelStudioBookPanel / `title` | `e.name` | [NovelStudioBookPanel.readable.js:266](source/frontend/NovelStudioBookPanel.readable.js#L266) | 静态已提取；运行未测 |
| UI-0111 | NovelStudioBookPanel / `onClick` | `n[3] \|\| (n[3] = e => W("foundation"))` | [NovelStudioBookPanel.readable.js:279](source/frontend/NovelStudioBookPanel.readable.js#L279) | 静态已提取；运行未测 |
| UI-0112 | NovelStudioBookPanel / `onClick` | `t => function (e) {<br>  b({<br>    type: "foundation",<br>    key: e.key,<br>    title: e.label<br>  });<br>}(e)` | [NovelStudioBookPanel.readable.js:287](source/frontend/NovelStudioBookPanel.readable.js#L287) | 静态已提取；运行未测 |
| UI-0113 | NovelStudioBookPanel / `title` | `e.label` | [NovelStudioBookPanel.readable.js:291](source/frontend/NovelStudioBookPanel.readable.js#L291) | 静态已提取；运行未测 |
| UI-0114 | NovelStudioBookPanel / `title` | `"编辑"` | [NovelStudioBookPanel.readable.js:300](source/frontend/NovelStudioBookPanel.readable.js#L300) | 静态已提取；运行未测 |
| UI-0115 | NovelStudioBookPanel / `onClick` | `Zi(t => function (e) {<br>  b({<br>    type: "foundation",<br>    key: e.key,<br>    title: e.label,<br>    autoEdit: !0<br>  });<br>}(e), ["stop"])` | [NovelStudioBookPanel.readable.js:301](source/frontend/NovelStudioBookPanel.readable.js#L301) | 静态已提取；运行未测 |
| UI-0116 | NovelStudioBookPanel / `title` | `e.label` | [NovelStudioBookPanel.readable.js:305](source/frontend/NovelStudioBookPanel.readable.js#L305) | 静态已提取；运行未测 |
| UI-0117 | NovelStudioBookPanel / `onClick` | `n[4] \|\| (n[4] = e => W("covers"))` | [NovelStudioBookPanel.readable.js:315](source/frontend/NovelStudioBookPanel.readable.js#L315) | 静态已提取；运行未测 |
| UI-0118 | NovelStudioBookPanel / `onClick` | `n[5] \|\| (n[5] = e => W("worldview"))` | [NovelStudioBookPanel.readable.js:329](source/frontend/NovelStudioBookPanel.readable.js#L329) | 静态已提取；运行未测 |
| UI-0119 | NovelStudioBookPanel / `onClick` | `k` | [NovelStudioBookPanel.readable.js:336](source/frontend/NovelStudioBookPanel.readable.js#L336) | 静态已提取；运行未测 |
| UI-0120 | NovelStudioBookPanel / `title` | `h.value` | [NovelStudioBookPanel.readable.js:341](source/frontend/NovelStudioBookPanel.readable.js#L341) | 静态已提取；运行未测 |
| UI-0121 | NovelStudioBookPanel / `onBack` | `y` | [NovelStudioBookPanel.readable.js:348](source/frontend/NovelStudioBookPanel.readable.js#L348) | 静态已提取；运行未测 |
| UI-0122 | NovelStudioBookPanel / `onSave` | `x` | [NovelStudioBookPanel.readable.js:349](source/frontend/NovelStudioBookPanel.readable.js#L349) | 静态已提取；运行未测 |
| UI-0123 | NovelStudioShortPanel / `title` | `e.label \|\| e.title \|\| "文件"` | [NovelStudioShortPanel.readable.js:161](source/frontend/NovelStudioShortPanel.readable.js#L161) | 静态已提取；运行未测 |
| UI-0124 | NovelStudioShortPanel / `title` | `"短篇大纲"` | [NovelStudioShortPanel.readable.js:169](source/frontend/NovelStudioShortPanel.readable.js#L169) | 静态已提取；运行未测 |
| UI-0125 | NovelStudioShortPanel / `title` | `"推荐"` | [NovelStudioShortPanel.readable.js:175](source/frontend/NovelStudioShortPanel.readable.js#L175) | 静态已提取；运行未测 |
| UI-0126 | NovelStudioShortPanel / `onClick` | `n[0] \|\| (n[0] = e => a("update:collapsed", !0))` | [NovelStudioShortPanel.readable.js:192](source/frontend/NovelStudioShortPanel.readable.js#L192) | 静态已提取；运行未测 |
| UI-0127 | NovelStudioShortPanel / `onClick` | `n[1] \|\| (n[1] = e => W("chapters"))` | [NovelStudioShortPanel.readable.js:205](source/frontend/NovelStudioShortPanel.readable.js#L205) | 静态已提取；运行未测 |
| UI-0128 | NovelStudioShortPanel / `onClick` | `t => x({<br>  key: "final/full.md",<br>  title: e.title<br>})` | [NovelStudioShortPanel.readable.js:213](source/frontend/NovelStudioShortPanel.readable.js#L213) | 静态已提取；运行未测 |
| UI-0129 | NovelStudioShortPanel / `title` | `e.title` | [NovelStudioShortPanel.readable.js:215](source/frontend/NovelStudioShortPanel.readable.js#L215) | 静态已提取；运行未测 |
| UI-0130 | NovelStudioShortPanel / `onClick` | `n[2] \|\| (n[2] = e => W("outline"))` | [NovelStudioShortPanel.readable.js:223](source/frontend/NovelStudioShortPanel.readable.js#L223) | 静态已提取；运行未测 |
| UI-0131 | NovelStudioShortPanel / `onClick` | `w` | [NovelStudioShortPanel.readable.js:231](source/frontend/NovelStudioShortPanel.readable.js#L231) | 静态已提取；运行未测 |
| UI-0132 | NovelStudioShortPanel / `onClick` | `n[3] \|\| (n[3] = e => W("artifacts"))` | [NovelStudioShortPanel.readable.js:235](source/frontend/NovelStudioShortPanel.readable.js#L235) | 静态已提取；运行未测 |
| UI-0133 | NovelStudioShortPanel / `onClick` | `t => x(e)` | [NovelStudioShortPanel.readable.js:243](source/frontend/NovelStudioShortPanel.readable.js#L243) | 静态已提取；运行未测 |
| UI-0134 | NovelStudioShortPanel / `title` | `"编辑"` | [NovelStudioShortPanel.readable.js:250](source/frontend/NovelStudioShortPanel.readable.js#L250) | 静态已提取；运行未测 |
| UI-0135 | NovelStudioShortPanel / `onClick` | `Zi(t => function (e) {<br>  b({<br>    key: e.key,<br>    title: e.label \|\| e.title \|\| "文件",<br>    autoEdit: !0<br>  });<br>}(e), ["stop"])` | [NovelStudioShortPanel.readable.js:251](source/frontend/NovelStudioShortPanel.readable.js#L251) | 静态已提取；运行未测 |
| UI-0136 | NovelStudioShortPanel / `title` | `e.label \|\| e.title \|\| "文件"` | [NovelStudioShortPanel.readable.js:254](source/frontend/NovelStudioShortPanel.readable.js#L254) | 静态已提取；运行未测 |
| UI-0137 | NovelStudioShortPanel / `onClick` | `n[4] \|\| (n[4] = e => W("sales"))` | [NovelStudioShortPanel.readable.js:264](source/frontend/NovelStudioShortPanel.readable.js#L264) | 静态已提取；运行未测 |
| UI-0138 | NovelStudioShortPanel / `onClick` | `_` | [NovelStudioShortPanel.readable.js:271](source/frontend/NovelStudioShortPanel.readable.js#L271) | 静态已提取；运行未测 |
| UI-0139 | NovelStudioShortPanel / `title` | `g.value` | [NovelStudioShortPanel.readable.js:276](source/frontend/NovelStudioShortPanel.readable.js#L276) | 静态已提取；运行未测 |
| UI-0140 | NovelStudioShortPanel / `onBack` | `y` | [NovelStudioShortPanel.readable.js:283](source/frontend/NovelStudioShortPanel.readable.js#L283) | 静态已提取；运行未测 |
| UI-0141 | NovelStudioShortPanel / `onSave` | `k` | [NovelStudioShortPanel.readable.js:284](source/frontend/NovelStudioShortPanel.readable.js#L284) | 静态已提取；运行未测 |
| UI-0142 | NovelStylePanel / `title` | `i.value.trim()` | [NovelStylePanel.readable.js:85](source/frontend/NovelStylePanel.readable.js#L85) | 静态已提取；运行未测 |
| UI-0143 | NovelStylePanel / `onClick` | `n[0] \|\| (n[0] = e => a("back"))` | [NovelStylePanel.readable.js:208](source/frontend/NovelStylePanel.readable.js#L208) | 静态已提取；运行未测 |
| UI-0144 | NovelStylePanel / `onClick` | `p` | [NovelStylePanel.readable.js:216](source/frontend/NovelStylePanel.readable.js#L216) | 静态已提取；运行未测 |
| UI-0145 | NovelStylePanel / `onUpdate:value` | `n[1] \|\| (n[1] = e => i.value = e)` | [NovelStylePanel.readable.js:219](source/frontend/NovelStylePanel.readable.js#L219) | 静态已提取；运行未测 |
| UI-0146 | NovelStylePanel / `placeholder` | `"例如：金庸武侠、知乎盐选、某作者口语风"` | [NovelStylePanel.readable.js:220](source/frontend/NovelStylePanel.readable.js#L220) | 静态已提取；运行未测 |
| UI-0147 | NovelStylePanel / `onUpdate:value` | `n[2] \|\| (n[2] = e => s.value = e)` | [NovelStylePanel.readable.js:223](source/frontend/NovelStylePanel.readable.js#L223) | 静态已提取；运行未测 |
| UI-0148 | NovelStylePanel / `placeholder` | `"直接粘贴样章、段落或写作要求……"` | [NovelStylePanel.readable.js:229](source/frontend/NovelStylePanel.readable.js#L229) | 静态已提取；运行未测 |
| UI-0149 | NovelStylePanel / `onUpdate:value` | `n[3] \|\| (n[3] = e => c.value = e)` | [NovelStylePanel.readable.js:232](source/frontend/NovelStylePanel.readable.js#L232) | 静态已提取；运行未测 |
| UI-0150 | NovelStylePanel / `placeholder` | `"可手写可执行的文风规则；不填则根据参考文本自动生成"` | [NovelStylePanel.readable.js:238](source/frontend/NovelStylePanel.readable.js#L238) | 静态已提取；运行未测 |
| UI-0151 | NovelStylePanel / `onClick` | `h` | [NovelStylePanel.readable.js:241](source/frontend/NovelStylePanel.readable.js#L241) | 静态已提取；运行未测 |
| UI-0152 | NovelStylePanel / `onClick` | `m` | [NovelStylePanel.readable.js:248](source/frontend/NovelStylePanel.readable.js#L248) | 静态已提取；运行未测 |
| UI-0153 | NovelStylePanel / `onUpdate:value` | `n[4] \|\| (n[4] = e => r.value = e)` | [NovelStylePanel.readable.js:255](source/frontend/NovelStylePanel.readable.js#L255) | 静态已提取；运行未测 |
| UI-0154 | NovelStylePanel / `onClick` | `Zi(t => f(e), ["prevent"])` | [NovelStylePanel.readable.js:266](source/frontend/NovelStylePanel.readable.js#L266) | 静态已提取；运行未测 |
| UI-0155 | NovelStylePanel / `onClick` | `t => g(e)` | [NovelStylePanel.readable.js:270](source/frontend/NovelStylePanel.readable.js#L270) | 静态已提取；运行未测 |
| UI-0156 | NovelStylePanel / `onClick` | `W` | [NovelStylePanel.readable.js:278](source/frontend/NovelStylePanel.readable.js#L278) | 静态已提取；运行未测 |
| UI-0157 | NovelAgentSettings / `title` | `e.title` | [NovelAgentSettings.readable.js:108](source/frontend/NovelAgentSettings.readable.js#L108) | 静态已提取；运行未测 |
| UI-0158 | NovelAgentSettings / `title` | `e.title` | [NovelAgentSettings.readable.js:114](source/frontend/NovelAgentSettings.readable.js#L114) | 静态已提取；运行未测 |
| UI-0159 | NovelAgentSettings / `title` | `e.title` | [NovelAgentSettings.readable.js:146](source/frontend/NovelAgentSettings.readable.js#L146) | 静态已提取；运行未测 |
| UI-0160 | NovelAgentSettings / `title` | `"自定义审稿"` | [NovelAgentSettings.readable.js:188](source/frontend/NovelAgentSettings.readable.js#L188) | 静态已提取；运行未测 |
| UI-0161 | NovelAgentSettings / `title` | `"自定义写作要求"` | [NovelAgentSettings.readable.js:211](source/frontend/NovelAgentSettings.readable.js#L211) | 静态已提取；运行未测 |
| UI-0162 | NovelAgentSettings / `placeholder` | `""` | [NovelAgentSettings.readable.js:214](source/frontend/NovelAgentSettings.readable.js#L214) | 静态已提取；运行未测 |
| UI-0163 | NovelAgentSettings / `onClick` | `n[0] \|\| (n[0] = e => a("back"))` | [NovelAgentSettings.readable.js:279](source/frontend/NovelAgentSettings.readable.js#L279) | 静态已提取；运行未测 |
| UI-0164 | NovelAgentSettings / `onUpdate:modelValue` | `e => s.value[t.key] = e` | [NovelAgentSettings.readable.js:291](source/frontend/NovelAgentSettings.readable.js#L291) | 静态已提取；运行未测 |
| UI-0165 | NovelAgentSettings / `placeholder` | `"跟随默认"` | [NovelAgentSettings.readable.js:293](source/frontend/NovelAgentSettings.readable.js#L293) | 静态已提取；运行未测 |
| UI-0166 | NovelAgentSettings / `onUpdate:value` | `e => s.value[t.key] = e` | [NovelAgentSettings.readable.js:301](source/frontend/NovelAgentSettings.readable.js#L301) | 静态已提取；运行未测 |
| UI-0167 | NovelAgentSettings / `placeholder` | `"默认 gemini 出图"` | [NovelAgentSettings.readable.js:303](source/frontend/NovelAgentSettings.readable.js#L303) | 静态已提取；运行未测 |
| UI-0168 | NovelAgentSettings / `onUpdate:value` | `n[1] \|\| (n[1] = e => d.value.maxRetries = e)` | [NovelAgentSettings.readable.js:312](source/frontend/NovelAgentSettings.readable.js#L312) | 静态已提取；运行未测 |
| UI-0169 | NovelAgentSettings / `min` | `0` | [NovelAgentSettings.readable.js:313](source/frontend/NovelAgentSettings.readable.js#L313) | 静态已提取；运行未测 |
| UI-0170 | NovelAgentSettings / `max` | `5` | [NovelAgentSettings.readable.js:314](source/frontend/NovelAgentSettings.readable.js#L314) | 静态已提取；运行未测 |
| UI-0171 | NovelAgentSettings / `onUpdate:value` | `n[2] \|\| (n[2] = e => d.value.passScore = e)` | [NovelAgentSettings.readable.js:318](source/frontend/NovelAgentSettings.readable.js#L318) | 静态已提取；运行未测 |
| UI-0172 | NovelAgentSettings / `min` | `0` | [NovelAgentSettings.readable.js:319](source/frontend/NovelAgentSettings.readable.js#L319) | 静态已提取；运行未测 |
| UI-0173 | NovelAgentSettings / `max` | `100` | [NovelAgentSettings.readable.js:320](source/frontend/NovelAgentSettings.readable.js#L320) | 静态已提取；运行未测 |
| UI-0174 | NovelAgentSettings / `onUpdate:value` | `n[3] \|\| (n[3] = e => d.value.temperatureStep = e)` | [NovelAgentSettings.readable.js:324](source/frontend/NovelAgentSettings.readable.js#L324) | 静态已提取；运行未测 |
| UI-0175 | NovelAgentSettings / `min` | `0` | [NovelAgentSettings.readable.js:325](source/frontend/NovelAgentSettings.readable.js#L325) | 静态已提取；运行未测 |
| UI-0176 | NovelAgentSettings / `max` | `.5` | [NovelAgentSettings.readable.js:326](source/frontend/NovelAgentSettings.readable.js#L326) | 静态已提取；运行未测 |
| UI-0177 | NovelAgentSettings / `onUpdate:value` | `n[4] \|\| (n[4] = e => d.value.baseTemperature = e)` | [NovelAgentSettings.readable.js:331](source/frontend/NovelAgentSettings.readable.js#L331) | 静态已提取；运行未测 |
| UI-0178 | NovelAgentSettings / `min` | `0` | [NovelAgentSettings.readable.js:332](source/frontend/NovelAgentSettings.readable.js#L332) | 静态已提取；运行未测 |
| UI-0179 | NovelAgentSettings / `max` | `2` | [NovelAgentSettings.readable.js:333](source/frontend/NovelAgentSettings.readable.js#L333) | 静态已提取；运行未测 |
| UI-0180 | NovelAgentSettings / `onUpdate:value` | `n[5] \|\| (n[5] = e => d.value.enableLengthNorm = e)` | [NovelAgentSettings.readable.js:338](source/frontend/NovelAgentSettings.readable.js#L338) | 静态已提取；运行未测 |
| UI-0181 | NovelAgentSettings / `onUpdate:value` | `n[6] \|\| (n[6] = e => d.value.enablePostWrite = e)` | [NovelAgentSettings.readable.js:344](source/frontend/NovelAgentSettings.readable.js#L344) | 静态已提取；运行未测 |
| UI-0182 | NovelAgentSettings / `onUpdate:value` | `n[7] \|\| (n[7] = e => d.value.enableAiTells = e)` | [NovelAgentSettings.readable.js:350](source/frontend/NovelAgentSettings.readable.js#L350) | 静态已提取；运行未测 |
| UI-0183 | NovelAgentSettings / `onUpdate:value` | `n[8] \|\| (n[8] = e => d.value.requireSecondaryConfirm = e)` | [NovelAgentSettings.readable.js:356](source/frontend/NovelAgentSettings.readable.js#L356) | 静态已提取；运行未测 |
| UI-0184 | NovelAgentSettings / `onClick` | `n[9] \|\| (n[9] = e => W(!0))` | [NovelAgentSettings.readable.js:363](source/frontend/NovelAgentSettings.readable.js#L363) | 静态已提取；运行未测 |
| UI-0185 | NovelAgentSettings / `onClick` | `n[10] \|\| (n[10] = e => W(!1))` | [NovelAgentSettings.readable.js:370](source/frontend/NovelAgentSettings.readable.js#L370) | 静态已提取；运行未测 |
| UI-0186 | NovelAgentSettings / `onUpdate:checked` | `t => function (e, t) {<br>  c.value = c.value.map(n => n.id === e.id ? {<br>    ...n,<br>    enabled: t<br>  } : n);<br>}(e, t)` | [NovelAgentSettings.readable.js:379](source/frontend/NovelAgentSettings.readable.js#L379) | 静态已提取；运行未测 |
| UI-0187 | NovelAgentSettings / `onClick` | `k` | [NovelAgentSettings.readable.js:390](source/frontend/NovelAgentSettings.readable.js#L390) | 静态已提取；运行未测 |
| UI-0188 | NovelAgentSettings / `onUpdate:checked` | `t => x(e, "enabled", t)` | [NovelAgentSettings.readable.js:399](source/frontend/NovelAgentSettings.readable.js#L399) | 静态已提取；运行未测 |
| UI-0189 | NovelAgentSettings / `placeholder` | `"提示词标题"` | [NovelAgentSettings.readable.js:403](source/frontend/NovelAgentSettings.readable.js#L403) | 静态已提取；运行未测 |
| UI-0190 | NovelAgentSettings / `onUpdate:value` | `t => x(e, "title", t)` | [NovelAgentSettings.readable.js:405](source/frontend/NovelAgentSettings.readable.js#L405) | 静态已提取；运行未测 |
| UI-0191 | NovelAgentSettings / `onClick` | `t => function (e) {<br>  c.value = c.value.filter(t => t.id !== e.id);<br>}(e)` | [NovelAgentSettings.readable.js:409](source/frontend/NovelAgentSettings.readable.js#L409) | 静态已提取；运行未测 |
| UI-0192 | NovelAgentSettings / `placeholder` | `"审稿要求说明…"` | [NovelAgentSettings.readable.js:419](source/frontend/NovelAgentSettings.readable.js#L419) | 静态已提取；运行未测 |
| UI-0193 | NovelAgentSettings / `onUpdate:value` | `t => x(e, "content", t)` | [NovelAgentSettings.readable.js:420](source/frontend/NovelAgentSettings.readable.js#L420) | 静态已提取；运行未测 |
| UI-0194 | NovelAgentSettings / `onClick` | `_` | [NovelAgentSettings.readable.js:426](source/frontend/NovelAgentSettings.readable.js#L426) | 静态已提取；运行未测 |
| UI-0195 | NovelAgentSettings / `onUpdate:checked` | `t => w(e, "enabled", t)` | [NovelAgentSettings.readable.js:435](source/frontend/NovelAgentSettings.readable.js#L435) | 静态已提取；运行未测 |
| UI-0196 | NovelAgentSettings / `placeholder` | `"提示词标题"` | [NovelAgentSettings.readable.js:439](source/frontend/NovelAgentSettings.readable.js#L439) | 静态已提取；运行未测 |
| UI-0197 | NovelAgentSettings / `onUpdate:value` | `t => w(e, "title", t)` | [NovelAgentSettings.readable.js:441](source/frontend/NovelAgentSettings.readable.js#L441) | 静态已提取；运行未测 |
| UI-0198 | NovelAgentSettings / `onUpdate:value` | `t => w(e, "category", t)` | [NovelAgentSettings.readable.js:449](source/frontend/NovelAgentSettings.readable.js#L449) | 静态已提取；运行未测 |
| UI-0199 | NovelAgentSettings / `onClick` | `t => function (e) {<br>  u.value = u.value.filter(t => t.id !== e.id);<br>}(e)` | [NovelAgentSettings.readable.js:453](source/frontend/NovelAgentSettings.readable.js#L453) | 静态已提取；运行未测 |
| UI-0200 | NovelAgentSettings / `placeholder` | `e.placeholder \|\| "写作要求说明……"` | [NovelAgentSettings.readable.js:463](source/frontend/NovelAgentSettings.readable.js#L463) | 静态已提取；运行未测 |
| UI-0201 | NovelAgentSettings / `onUpdate:value` | `t => w(e, "content", t)` | [NovelAgentSettings.readable.js:464](source/frontend/NovelAgentSettings.readable.js#L464) | 静态已提取；运行未测 |
| UI-0202 | NovelAgentSettings / `disabled` | `r.value` | [NovelAgentSettings.readable.js:472](source/frontend/NovelAgentSettings.readable.js#L472) | 静态已提取；运行未测 |
| UI-0203 | NovelAgentSettings / `onClick` | `b` | [NovelAgentSettings.readable.js:473](source/frontend/NovelAgentSettings.readable.js#L473) | 静态已提取；运行未测 |
| UI-0204 | NovelImportPage / `title` | `"仿写创作"` | [NovelImportPage.readable.js:130](source/frontend/NovelImportPage.readable.js#L130) | 静态已提取；运行未测 |
| UI-0205 | NovelImportPage / `title` | `"同人创作"` | [NovelImportPage.readable.js:135](source/frontend/NovelImportPage.readable.js#L135) | 静态已提取；运行未测 |
| UI-0206 | NovelImportPage / `title` | `"番外创作"` | [NovelImportPage.readable.js:140](source/frontend/NovelImportPage.readable.js#L140) | 静态已提取；运行未测 |
| UI-0207 | NovelImportPage / `title` | `"续写创作"` | [NovelImportPage.readable.js:145](source/frontend/NovelImportPage.readable.js#L145) | 静态已提取；运行未测 |
| UI-0208 | NovelImportPage / `title` | `"导入素材"` | [NovelImportPage.readable.js:150](source/frontend/NovelImportPage.readable.js#L150) | 静态已提取；运行未测 |
| UI-0209 | NovelImportPage / `title` | `n` | [NovelImportPage.readable.js:208](source/frontend/NovelImportPage.readable.js#L208) | 静态已提取；运行未测 |
| UI-0210 | NovelImportPage / `title` | `"第1章"` | [NovelImportPage.readable.js:214](source/frontend/NovelImportPage.readable.js#L214) | 静态已提取；运行未测 |
| UI-0211 | NovelImportPage / `title` | `i.value.trim()` | [NovelImportPage.readable.js:317](source/frontend/NovelImportPage.readable.js#L317) | 静态已提取；运行未测 |
| UI-0212 | NovelImportPage / `onClick` | `le` | [NovelImportPage.readable.js:467](source/frontend/NovelImportPage.readable.js#L467) | 静态已提取；运行未测 |
| UI-0213 | NovelImportPage / `onClick` | `le` | [NovelImportPage.readable.js:471](source/frontend/NovelImportPage.readable.js#L471) | 静态已提取；运行未测 |
| UI-0214 | NovelImportPage / `onClick` | `t => Q.value = e.key` | [NovelImportPage.readable.js:478](source/frontend/NovelImportPage.readable.js#L478) | 静态已提取；运行未测 |
| UI-0215 | NovelImportPage / `onUpdate:value` | `n[0] \|\| (n[0] = e => i.value = e)` | [NovelImportPage.readable.js:489](source/frontend/NovelImportPage.readable.js#L489) | 静态已提取；运行未测 |
| UI-0216 | NovelImportPage / `placeholder` | `"留空则自动生成"` | [NovelImportPage.readable.js:490](source/frontend/NovelImportPage.readable.js#L490) | 静态已提取；运行未测 |
| UI-0217 | NovelImportPage / `disabled` | `l.value` | [NovelImportPage.readable.js:491](source/frontend/NovelImportPage.readable.js#L491) | 静态已提取；运行未测 |
| UI-0218 | NovelImportPage / `onUpdate:value` | `n[1] \|\| (n[1] = e => c.value = e)` | [NovelImportPage.readable.js:500](source/frontend/NovelImportPage.readable.js#L500) | 静态已提取；运行未测 |
| UI-0219 | NovelImportPage / `placeholder` | `"粘贴参考小说全文（至少 100 字），或上传 txt / docx 自动填入"` | [NovelImportPage.readable.js:503](source/frontend/NovelImportPage.readable.js#L503) | 静态已提取；运行未测 |
| UI-0220 | NovelImportPage / `disabled` | `l.value` | [NovelImportPage.readable.js:504](source/frontend/NovelImportPage.readable.js#L504) | 静态已提取；运行未测 |
| UI-0221 | NovelImportPage / `disabled` | `l.value \|\| W.value` | [NovelImportPage.readable.js:510](source/frontend/NovelImportPage.readable.js#L510) | 静态已提取；运行未测 |
| UI-0222 | NovelImportPage / `onClick` | `X` | [NovelImportPage.readable.js:511](source/frontend/NovelImportPage.readable.js#L511) | 静态已提取；运行未测 |
| UI-0223 | NovelImportPage / `onClick` | `e => {<br>  return n = t, y.value.splice(n, 1), void (y.value.length \|\| (b.value = [], k.value = []));<br>  var n;<br>}` | [NovelImportPage.readable.js:518](source/frontend/NovelImportPage.readable.js#L518) | 静态已提取；运行未测 |
| UI-0224 | NovelImportPage / `onClick` | `n[2] \|\| (n[2] = e => $(3))` | [NovelImportPage.readable.js:528](source/frontend/NovelImportPage.readable.js#L528) | 静态已提取；运行未测 |
| UI-0225 | NovelImportPage / `onClick` | `n[3] \|\| (n[3] = e => $(5))` | [NovelImportPage.readable.js:534](source/frontend/NovelImportPage.readable.js#L534) | 静态已提取；运行未测 |
| UI-0226 | NovelImportPage / `onClick` | `n[4] \|\| (n[4] = e => $(10))` | [NovelImportPage.readable.js:541](source/frontend/NovelImportPage.readable.js#L541) | 静态已提取；运行未测 |
| UI-0227 | NovelImportPage / `onClick` | `te` | [NovelImportPage.readable.js:547](source/frontend/NovelImportPage.readable.js#L547) | 静态已提取；运行未测 |
| UI-0228 | NovelImportPage / `onUpdate:value` | `n[5] \|\| (n[5] = e => _.value = e)` | [NovelImportPage.readable.js:564](source/frontend/NovelImportPage.readable.js#L564) | 静态已提取；运行未测 |
| UI-0229 | NovelImportPage / `min` | `1e3` | [NovelImportPage.readable.js:565](source/frontend/NovelImportPage.readable.js#L565) | 静态已提取；运行未测 |
| UI-0230 | NovelImportPage / `max` | `5e4` | [NovelImportPage.readable.js:566](source/frontend/NovelImportPage.readable.js#L566) | 静态已提取；运行未测 |
| UI-0231 | NovelImportPage / `placeholder` | `"5000"` | [NovelImportPage.readable.js:569](source/frontend/NovelImportPage.readable.js#L569) | 静态已提取；运行未测 |
| UI-0232 | NovelImportPage / `disabled` | `l.value` | [NovelImportPage.readable.js:570](source/frontend/NovelImportPage.readable.js#L570) | 静态已提取；运行未测 |
| UI-0233 | NovelImportPage / `disabled` | `l.value` | [NovelImportPage.readable.js:578](source/frontend/NovelImportPage.readable.js#L578) | 静态已提取；运行未测 |
| UI-0234 | NovelImportPage / `onClick` | `t => {<br>  return n = e.value, void (_.value = n);<br>  var n;<br>}` | [NovelImportPage.readable.js:579](source/frontend/NovelImportPage.readable.js#L579) | 静态已提取；运行未测 |
| UI-0235 | NovelImportPage / `onUpdate:value` | `n[6] \|\| (n[6] = e => S.value = e)` | [NovelImportPage.readable.js:587](source/frontend/NovelImportPage.readable.js#L587) | 静态已提取；运行未测 |
| UI-0236 | NovelImportPage / `min` | `1` | [NovelImportPage.readable.js:588](source/frontend/NovelImportPage.readable.js#L588) | 静态已提取；运行未测 |
| UI-0237 | NovelImportPage / `max` | `500` | [NovelImportPage.readable.js:589](source/frontend/NovelImportPage.readable.js#L589) | 静态已提取；运行未测 |
| UI-0238 | NovelImportPage / `placeholder` | `"100"` | [NovelImportPage.readable.js:592](source/frontend/NovelImportPage.readable.js#L592) | 静态已提取；运行未测 |
| UI-0239 | NovelImportPage / `disabled` | `l.value` | [NovelImportPage.readable.js:593](source/frontend/NovelImportPage.readable.js#L593) | 静态已提取；运行未测 |
| UI-0240 | NovelImportPage / `disabled` | `l.value` | [NovelImportPage.readable.js:601](source/frontend/NovelImportPage.readable.js#L601) | 静态已提取；运行未测 |
| UI-0241 | NovelImportPage / `onClick` | `t => S.value = e.value` | [NovelImportPage.readable.js:602](source/frontend/NovelImportPage.readable.js#L602) | 静态已提取；运行未测 |
| UI-0242 | NovelImportPage / `onUpdate:value` | `n[7] \|\| (n[7] = e => C.value = e)` | [NovelImportPage.readable.js:607](source/frontend/NovelImportPage.readable.js#L607) | 静态已提取；运行未测 |
| UI-0243 | NovelImportPage / `placeholder` | `"填写题材偏好，如：玄幻、都市"` | [NovelImportPage.readable.js:608](source/frontend/NovelImportPage.readable.js#L608) | 静态已提取；运行未测 |
| UI-0244 | NovelImportPage / `disabled` | `l.value` | [NovelImportPage.readable.js:610](source/frontend/NovelImportPage.readable.js#L610) | 静态已提取；运行未测 |
| UI-0245 | NovelImportPage / `disabled` | `l.value` | [NovelImportPage.readable.js:615](source/frontend/NovelImportPage.readable.js#L615) | 静态已提取；运行未测 |
| UI-0246 | NovelImportPage / `onClick` | `t => function (e) {<br>  const t = C.value.trim();<br>  if (!t) return void (C.value = e);<br>  t.split(/[、,，\s]+/).filter(Boolean).includes(e) \|\| (C.value = `${t}、${e}`);<br>}(e)` | [NovelImportPage.readable.js:616](source/frontend/NovelImportPage.readable.js#L616) | 静态已提取；运行未测 |
| UI-0247 | NovelImportPage / `onUpdate:value` | `n[8] \|\| (n[8] = e => s.value = e)` | [NovelImportPage.readable.js:625](source/frontend/NovelImportPage.readable.js#L625) | 静态已提取；运行未测 |
| UI-0248 | NovelImportPage / `placeholder` | `"剧情、主角、核心冲突（原创剧情）"` | [NovelImportPage.readable.js:628](source/frontend/NovelImportPage.readable.js#L628) | 静态已提取；运行未测 |
| UI-0249 | NovelImportPage / `disabled` | `l.value` | [NovelImportPage.readable.js:629](source/frontend/NovelImportPage.readable.js#L629) | 静态已提取；运行未测 |
| UI-0250 | NovelImportPage / `onUpdate:value` | `n[9] \|\| (n[9] = e => m.value = e)` | [NovelImportPage.readable.js:632](source/frontend/NovelImportPage.readable.js#L632) | 静态已提取；运行未测 |
| UI-0251 | NovelImportPage / `placeholder` | `"留空则默认采用拆书后的方向一；可填「方向二：…」"` | [NovelImportPage.readable.js:635](source/frontend/NovelImportPage.readable.js#L635) | 静态已提取；运行未测 |
| UI-0252 | NovelImportPage / `disabled` | `l.value` | [NovelImportPage.readable.js:636](source/frontend/NovelImportPage.readable.js#L636) | 静态已提取；运行未测 |
| UI-0253 | NovelImportPage / `onUpdate:value` | `n[10] \|\| (n[10] = e => P.value = e)` | [NovelImportPage.readable.js:639](source/frontend/NovelImportPage.readable.js#L639) | 静态已提取；运行未测 |
| UI-0254 | NovelImportPage / `placeholder` | `"填写额外要求，如人设、情节走向、必须出现的桥段、禁忌点等"` | [NovelImportPage.readable.js:642](source/frontend/NovelImportPage.readable.js#L642) | 静态已提取；运行未测 |
| UI-0255 | NovelImportPage / `disabled` | `l.value` | [NovelImportPage.readable.js:643](source/frontend/NovelImportPage.readable.js#L643) | 静态已提取；运行未测 |
| UI-0256 | NovelImportPage / `onUpdate:value` | `n[11] \|\| (n[11] = e => u.value = e)` | [NovelImportPage.readable.js:646](source/frontend/NovelImportPage.readable.js#L646) | 静态已提取；运行未测 |
| UI-0257 | NovelImportPage / `onUpdate:value` | `n[12] \|\| (n[12] = e => h.value = e)` | [NovelImportPage.readable.js:657](source/frontend/NovelImportPage.readable.js#L657) | 静态已提取；运行未测 |
| UI-0258 | NovelImportPage / `placeholder` | `"选择备忘录作为参考素材"` | [NovelImportPage.readable.js:659](source/frontend/NovelImportPage.readable.js#L659) | 静态已提取；运行未测 |
| UI-0259 | NovelImportPage / `multiple` | `""` | [NovelImportPage.readable.js:660](source/frontend/NovelImportPage.readable.js#L660) | 静态已提取；运行未测 |
| UI-0260 | NovelImportPage / `disabled` | `l.value` | [NovelImportPage.readable.js:662](source/frontend/NovelImportPage.readable.js#L662) | 静态已提取；运行未测 |
| UI-0261 | NovelImportPage / `onFocus` | `G` | [NovelImportPage.readable.js:663](source/frontend/NovelImportPage.readable.js#L663) | 静态已提取；运行未测 |
| UI-0262 | NovelImportPage / `onUpdate:value` | `n[13] \|\| (n[13] = e => d.value = e)` | [NovelImportPage.readable.js:668](source/frontend/NovelImportPage.readable.js#L668) | 静态已提取；运行未测 |
| UI-0263 | NovelImportPage / `placeholder` | `"选择已有正传"` | [NovelImportPage.readable.js:670](source/frontend/NovelImportPage.readable.js#L670) | 静态已提取；运行未测 |
| UI-0264 | NovelImportPage / `onUpdate:value` | `n[14] \|\| (n[14] = e => v.value = e)` | [NovelImportPage.readable.js:674](source/frontend/NovelImportPage.readable.js#L674) | 静态已提取；运行未测 |
| UI-0265 | NovelImportPage / `placeholder` | `"选择已有书籍或留空新建"` | [NovelImportPage.readable.js:676](source/frontend/NovelImportPage.readable.js#L676) | 静态已提取；运行未测 |
| UI-0266 | NovelImportPage / `onUpdate:value` | `n[15] \|\| (n[15] = e => i.value = e)` | [NovelImportPage.readable.js:680](source/frontend/NovelImportPage.readable.js#L680) | 静态已提取；运行未测 |
| UI-0267 | NovelImportPage / `placeholder` | `"暂定书名"` | [NovelImportPage.readable.js:681](source/frontend/NovelImportPage.readable.js#L681) | 静态已提取；运行未测 |
| UI-0268 | NovelImportPage / `onUpdate:value` | `n[16] \|\| (n[16] = e => f.value = e)` | [NovelImportPage.readable.js:684](source/frontend/NovelImportPage.readable.js#L684) | 静态已提取；运行未测 |
| UI-0269 | NovelImportPage / `onUpdate:value` | `n[17] \|\| (n[17] = e => s.value = e)` | [NovelImportPage.readable.js:700](source/frontend/NovelImportPage.readable.js#L700) | 静态已提取；运行未测 |
| UI-0270 | NovelImportPage / `placeholder` | `"番外要写什么"` | [NovelImportPage.readable.js:703](source/frontend/NovelImportPage.readable.js#L703) | 静态已提取；运行未测 |
| UI-0271 | NovelImportPage / `onUpdate:value` | `n[18] \|\| (n[18] = e => s.value = e)` | [NovelImportPage.readable.js:706](source/frontend/NovelImportPage.readable.js#L706) | 静态已提取；运行未测 |
| UI-0272 | NovelImportPage / `placeholder` | `"CP、AU、IF 线等"` | [NovelImportPage.readable.js:709](source/frontend/NovelImportPage.readable.js#L709) | 静态已提取；运行未测 |
| UI-0273 | NovelImportPage / `onUpdate:value` | `n[19] \|\| (n[19] = e => p.value = e)` | [NovelImportPage.readable.js:712](source/frontend/NovelImportPage.readable.js#L712) | 静态已提取；运行未测 |
| UI-0274 | NovelImportPage / `onUpdate:value` | `n[20] \|\| (n[20] = e => c.value = e)` | [NovelImportPage.readable.js:728](source/frontend/NovelImportPage.readable.js#L728) | 静态已提取；运行未测 |
| UI-0275 | NovelImportPage / `placeholder` | `"chapters" === Q.value ? "粘贴章节，建议以「第1章」等标题分段" : "粘贴参考原文；也可只选下方文风"` | [NovelImportPage.readable.js:731](source/frontend/NovelImportPage.readable.js#L731) | 静态已提取；运行未测 |
| UI-0276 | NovelImportPage / `onUpdate:value` | `n[21] \|\| (n[21] = e => u.value = e)` | [NovelImportPage.readable.js:734](source/frontend/NovelImportPage.readable.js#L734) | 静态已提取；运行未测 |
| UI-0277 | NovelImportPage / `onClick` | `le` | [NovelImportPage.readable.js:752](source/frontend/NovelImportPage.readable.js#L752) | 静态已提取；运行未测 |
| UI-0278 | NovelImportPage / `onClick` | `ie` | [NovelImportPage.readable.js:760](source/frontend/NovelImportPage.readable.js#L760) | 静态已提取；运行未测 |
| UI-0279 | NovelImportPage / `onUpdate:show` | `n[25] \|\| (n[25] = e => x.value = e)` | [NovelImportPage.readable.js:766](source/frontend/NovelImportPage.readable.js#L766) | 静态已提取；运行未测 |
| UI-0280 | NovelImportPage / `onClick` | `n[22] \|\| (n[22] = e => x.value = !1)` | [NovelImportPage.readable.js:782](source/frontend/NovelImportPage.readable.js#L782) | 静态已提取；运行未测 |
| UI-0281 | NovelImportPage / `onClick` | `ae` | [NovelImportPage.readable.js:786](source/frontend/NovelImportPage.readable.js#L786) | 静态已提取；运行未测 |
| UI-0282 | NovelImportPage / `onClick` | `oe` | [NovelImportPage.readable.js:790](source/frontend/NovelImportPage.readable.js#L790) | 静态已提取；运行未测 |
| UI-0283 | NovelImportPage / `onClick` | `Zi(e => ne(t), ["prevent"])` | [NovelImportPage.readable.js:802](source/frontend/NovelImportPage.readable.js#L802) | 静态已提取；运行未测 |
| UI-0284 | NovelImportPage / `onClick` | `n[23] \|\| (n[23] = Zi(() => {}, ["stop"]))` | [NovelImportPage.readable.js:805](source/frontend/NovelImportPage.readable.js#L805) | 静态已提取；运行未测 |
| UI-0285 | NovelImportPage / `onUpdate:checked` | `e => ne(t)` | [NovelImportPage.readable.js:806](source/frontend/NovelImportPage.readable.js#L806) | 静态已提取；运行未测 |
| UI-0286 | NovelImportPage / `onClick` | `n[24] \|\| (n[24] = e => x.value = !1)` | [NovelImportPage.readable.js:815](source/frontend/NovelImportPage.readable.js#L815) | 静态已提取；运行未测 |
| UI-0287 | NovelImportPage / `onClick` | `re` | [NovelImportPage.readable.js:821](source/frontend/NovelImportPage.readable.js#L821) | 静态已提取；运行未测 |
| UI-0288 | NovelImportModal / `onUpdate:show` | `r[2] \|\| (r[2] = e => n("update:show", e))` | [NovelImportModal.readable.js:42](source/frontend/NovelImportModal.readable.js#L42) | 静态已提取；运行未测 |
| UI-0289 | NovelImportModal / `onUpdate:tab` | `r[0] \|\| (r[0] = e => n("update:tab", e))` | [NovelImportModal.readable.js:55](source/frontend/NovelImportModal.readable.js#L55) | 静态已提取；运行未测 |
| UI-0290 | NovelImportModal / `onClose` | `a` | [NovelImportModal.readable.js:56](source/frontend/NovelImportModal.readable.js#L56) | 静态已提取；运行未测 |
| UI-0291 | NovelImportModal / `onCreated` | `o` | [NovelImportModal.readable.js:57](source/frontend/NovelImportModal.readable.js#L57) | 静态已提取；运行未测 |
| UI-0292 | NovelImportModal / `onSubmitStart` | `r[1] \|\| (r[1] = e => n("submit-start", e))` | [NovelImportModal.readable.js:58](source/frontend/NovelImportModal.readable.js#L58) | 静态已提取；运行未测 |
| UI-0293 | NovelImitationShortDialog / `title` | `g.value.trim() \|\| void 0` | [NovelImitationShortDialog.readable.js:152](source/frontend/NovelImitationShortDialog.readable.js#L152) | 静态已提取；运行未测 |
| UI-0294 | NovelImitationShortDialog / `onUpdate:show` | `n[9] \|\| (n[9] = e => a("update:show", e))` | [NovelImitationShortDialog.readable.js:170](source/frontend/NovelImitationShortDialog.readable.js#L170) | 静态已提取；运行未测 |
| UI-0295 | NovelImitationShortDialog / `disabled` | `r.value` | [NovelImitationShortDialog.readable.js:184](source/frontend/NovelImitationShortDialog.readable.js#L184) | 静态已提取；运行未测 |
| UI-0296 | NovelImitationShortDialog / `onClick` | `O` | [NovelImitationShortDialog.readable.js:185](source/frontend/NovelImitationShortDialog.readable.js#L185) | 静态已提取；运行未测 |
| UI-0297 | NovelImitationShortDialog / `onUpdate:value` | `n[0] \|\| (n[0] = e => g.value = e)` | [NovelImitationShortDialog.readable.js:194](source/frontend/NovelImitationShortDialog.readable.js#L194) | 静态已提取；运行未测 |
| UI-0298 | NovelImitationShortDialog / `placeholder` | `"留空则自动生成"` | [NovelImitationShortDialog.readable.js:195](source/frontend/NovelImitationShortDialog.readable.js#L195) | 静态已提取；运行未测 |
| UI-0299 | NovelImitationShortDialog / `onUpdate:value` | `n[1] \|\| (n[1] = e => i.value = e)` | [NovelImitationShortDialog.readable.js:204](source/frontend/NovelImitationShortDialog.readable.js#L204) | 静态已提取；运行未测 |
| UI-0300 | NovelImitationShortDialog / `placeholder` | `"粘贴参考小说全文（至少 100 字），或上传 txt / docx 自动填入"` | [NovelImitationShortDialog.readable.js:207](source/frontend/NovelImitationShortDialog.readable.js#L207) | 静态已提取；运行未测 |
| UI-0301 | NovelImitationShortDialog / `disabled` | `r.value` | [NovelImitationShortDialog.readable.js:208](source/frontend/NovelImitationShortDialog.readable.js#L208) | 静态已提取；运行未测 |
| UI-0302 | NovelImitationShortDialog / `disabled` | `r.value \|\| l.value` | [NovelImitationShortDialog.readable.js:214](source/frontend/NovelImitationShortDialog.readable.js#L214) | 静态已提取；运行未测 |
| UI-0303 | NovelImitationShortDialog / `onClick` | `N` | [NovelImitationShortDialog.readable.js:215](source/frontend/NovelImitationShortDialog.readable.js#L215) | 静态已提取；运行未测 |
| UI-0304 | NovelImitationShortDialog / `onClick` | `e => {<br>  return n = t, void s.value.splice(n, 1);<br>  var n;<br>}` | [NovelImitationShortDialog.readable.js:222](source/frontend/NovelImitationShortDialog.readable.js#L222) | 静态已提取；运行未测 |
| UI-0305 | NovelImitationShortDialog / `onUpdate:value` | `n[2] \|\| (n[2] = e => u.value = e)` | [NovelImitationShortDialog.readable.js:228](source/frontend/NovelImitationShortDialog.readable.js#L228) | 静态已提取；运行未测 |
| UI-0306 | NovelImitationShortDialog / `min` | `1e3` | [NovelImitationShortDialog.readable.js:229](source/frontend/NovelImitationShortDialog.readable.js#L229) | 静态已提取；运行未测 |
| UI-0307 | NovelImitationShortDialog / `max` | `5e5` | [NovelImitationShortDialog.readable.js:230](source/frontend/NovelImitationShortDialog.readable.js#L230) | 静态已提取；运行未测 |
| UI-0308 | NovelImitationShortDialog / `placeholder` | `"如 13000"` | [NovelImitationShortDialog.readable.js:233](source/frontend/NovelImitationShortDialog.readable.js#L233) | 静态已提取；运行未测 |
| UI-0309 | NovelImitationShortDialog / `disabled` | `r.value` | [NovelImitationShortDialog.readable.js:234](source/frontend/NovelImitationShortDialog.readable.js#L234) | 静态已提取；运行未测 |
| UI-0310 | NovelImitationShortDialog / `disabled` | `r.value` | [NovelImitationShortDialog.readable.js:242](source/frontend/NovelImitationShortDialog.readable.js#L242) | 静态已提取；运行未测 |
| UI-0311 | NovelImitationShortDialog / `onClick` | `t => {<br>  return n = e.value, void (u.value = n);<br>  var n;<br>}` | [NovelImitationShortDialog.readable.js:243](source/frontend/NovelImitationShortDialog.readable.js#L243) | 静态已提取；运行未测 |
| UI-0312 | NovelImitationShortDialog / `onUpdate:value` | `n[3] \|\| (n[3] = e => p.value = e)` | [NovelImitationShortDialog.readable.js:251](source/frontend/NovelImitationShortDialog.readable.js#L251) | 静态已提取；运行未测 |
| UI-0313 | NovelImitationShortDialog / `placeholder` | `"填写读者偏好，如：轻松搞笑、甜宠治愈"` | [NovelImitationShortDialog.readable.js:252](source/frontend/NovelImitationShortDialog.readable.js#L252) | 静态已提取；运行未测 |
| UI-0314 | NovelImitationShortDialog / `disabled` | `r.value` | [NovelImitationShortDialog.readable.js:254](source/frontend/NovelImitationShortDialog.readable.js#L254) | 静态已提取；运行未测 |
| UI-0315 | NovelImitationShortDialog / `disabled` | `r.value` | [NovelImitationShortDialog.readable.js:259](source/frontend/NovelImitationShortDialog.readable.js#L259) | 静态已提取；运行未测 |
| UI-0316 | NovelImitationShortDialog / `onClick` | `t => function (e) {<br>  const t = p.value.trim();<br>  if (!t) return void (p.value = e);<br>  t.split(/[、,，\s]+/).filter(Boolean).includes(e) \|\| (p.value = `${t}、${e}`);<br>}(e)` | [NovelImitationShortDialog.readable.js:260](source/frontend/NovelImitationShortDialog.readable.js#L260) | 静态已提取；运行未测 |
| UI-0317 | NovelImitationShortDialog / `onUpdate:value` | `n[4] \|\| (n[4] = e => d.value = e)` | [NovelImitationShortDialog.readable.js:269](source/frontend/NovelImitationShortDialog.readable.js#L269) | 静态已提取；运行未测 |
| UI-0318 | NovelImitationShortDialog / `min` | `1` | [NovelImitationShortDialog.readable.js:270](source/frontend/NovelImitationShortDialog.readable.js#L270) | 静态已提取；运行未测 |
| UI-0319 | NovelImitationShortDialog / `max` | `50` | [NovelImitationShortDialog.readable.js:271](source/frontend/NovelImitationShortDialog.readable.js#L271) | 静态已提取；运行未测 |
| UI-0320 | NovelImitationShortDialog / `placeholder` | `"1"` | [NovelImitationShortDialog.readable.js:274](source/frontend/NovelImitationShortDialog.readable.js#L274) | 静态已提取；运行未测 |
| UI-0321 | NovelImitationShortDialog / `disabled` | `r.value` | [NovelImitationShortDialog.readable.js:275](source/frontend/NovelImitationShortDialog.readable.js#L275) | 静态已提取；运行未测 |
| UI-0322 | NovelImitationShortDialog / `onUpdate:value` | `n[5] \|\| (n[5] = e => v.value = e)` | [NovelImitationShortDialog.readable.js:281](source/frontend/NovelImitationShortDialog.readable.js#L281) | 静态已提取；运行未测 |
| UI-0323 | NovelImitationShortDialog / `placeholder` | `"填写额外要求，如人设、情节走向、必须出现的桥段、禁忌点等"` | [NovelImitationShortDialog.readable.js:284](source/frontend/NovelImitationShortDialog.readable.js#L284) | 静态已提取；运行未测 |
| UI-0324 | NovelImitationShortDialog / `disabled` | `r.value` | [NovelImitationShortDialog.readable.js:285](source/frontend/NovelImitationShortDialog.readable.js#L285) | 静态已提取；运行未测 |
| UI-0325 | NovelImitationShortDialog / `disabled` | `r.value` | [NovelImitationShortDialog.readable.js:292](source/frontend/NovelImitationShortDialog.readable.js#L292) | 静态已提取；运行未测 |
| UI-0326 | NovelImitationShortDialog / `onClick` | `t => function (e) {<br>  const t = f.value.indexOf(e);<br>  t >= 0 ? f.value.splice(t, 1) : f.value.push(e);<br>}(e)` | [NovelImitationShortDialog.readable.js:293](source/frontend/NovelImitationShortDialog.readable.js#L293) | 静态已提取；运行未测 |
| UI-0327 | NovelImitationShortDialog / `onUpdate:value` | `n[6] \|\| (n[6] = e => c.value = e)` | [NovelImitationShortDialog.readable.js:299](source/frontend/NovelImitationShortDialog.readable.js#L299) | 静态已提取；运行未测 |
| UI-0328 | NovelImitationShortDialog / `onUpdate:value` | `n[7] \|\| (n[7] = e => h.value = e)` | [NovelImitationShortDialog.readable.js:310](source/frontend/NovelImitationShortDialog.readable.js#L310) | 静态已提取；运行未测 |
| UI-0329 | NovelImitationShortDialog / `placeholder` | `"选择备忘录作为参考素材"` | [NovelImitationShortDialog.readable.js:312](source/frontend/NovelImitationShortDialog.readable.js#L312) | 静态已提取；运行未测 |
| UI-0330 | NovelImitationShortDialog / `multiple` | `""` | [NovelImitationShortDialog.readable.js:313](source/frontend/NovelImitationShortDialog.readable.js#L313) | 静态已提取；运行未测 |
| UI-0331 | NovelImitationShortDialog / `disabled` | `r.value` | [NovelImitationShortDialog.readable.js:315](source/frontend/NovelImitationShortDialog.readable.js#L315) | 静态已提取；运行未测 |
| UI-0332 | NovelImitationShortDialog / `onFocus` | `y` | [NovelImitationShortDialog.readable.js:316](source/frontend/NovelImitationShortDialog.readable.js#L316) | 静态已提取；运行未测 |
| UI-0333 | NovelImitationShortDialog / `onUpdate:value` | `n[8] \|\| (n[8] = e => m.value = e)` | [NovelImitationShortDialog.readable.js:319](source/frontend/NovelImitationShortDialog.readable.js#L319) | 静态已提取；运行未测 |
| UI-0334 | NovelImitationShortDialog / `placeholder` | `"关联已有作品"` | [NovelImitationShortDialog.readable.js:321](source/frontend/NovelImitationShortDialog.readable.js#L321) | 静态已提取；运行未测 |
| UI-0335 | NovelImitationShortDialog / `disabled` | `r.value` | [NovelImitationShortDialog.readable.js:323](source/frontend/NovelImitationShortDialog.readable.js#L323) | 静态已提取；运行未测 |
| UI-0336 | NovelImitationShortDialog / `disabled` | `r.value` | [NovelImitationShortDialog.readable.js:328](source/frontend/NovelImitationShortDialog.readable.js#L328) | 静态已提取；运行未测 |
| UI-0337 | NovelImitationShortDialog / `onClick` | `O` | [NovelImitationShortDialog.readable.js:329](source/frontend/NovelImitationShortDialog.readable.js#L329) | 静态已提取；运行未测 |
| UI-0338 | NovelImitationShortDialog / `disabled` | `!C.value` | [NovelImitationShortDialog.readable.js:335](source/frontend/NovelImitationShortDialog.readable.js#L335) | 静态已提取；运行未测 |
| UI-0339 | NovelImitationShortDialog / `onClick` | `G` | [NovelImitationShortDialog.readable.js:336](source/frontend/NovelImitationShortDialog.readable.js#L336) | 静态已提取；运行未测 |
| UI-0340 | NovelBatchDialog / `title` | `"批量写作"` | [NovelBatchDialog.readable.js:104](source/frontend/NovelBatchDialog.readable.js#L104) | 静态已提取；运行未测 |
| UI-0341 | NovelBatchDialog / `onUpdate:show` | `p` | [NovelBatchDialog.readable.js:110](source/frontend/NovelBatchDialog.readable.js#L110) | 静态已提取；运行未测 |
| UI-0342 | NovelBatchDialog / `onClick` | `v` | [NovelBatchDialog.readable.js:130](source/frontend/NovelBatchDialog.readable.js#L130) | 静态已提取；运行未测 |
| UI-0343 | NovelBatchDialog / `onUpdate:value` | `n[0] \|\| (n[0] = e => r.value = e)` | [NovelBatchDialog.readable.js:138](source/frontend/NovelBatchDialog.readable.js#L138) | 静态已提取；运行未测 |
| UI-0344 | NovelBatchDialog / `min` | `1` | [NovelBatchDialog.readable.js:139](source/frontend/NovelBatchDialog.readable.js#L139) | 静态已提取；运行未测 |
| UI-0345 | NovelBatchDialog / `max` | `50` | [NovelBatchDialog.readable.js:140](source/frontend/NovelBatchDialog.readable.js#L140) | 静态已提取；运行未测 |
| UI-0346 | NovelBatchDialog / `onUpdate:value` | `n[1] \|\| (n[1] = e => l.value = e)` | [NovelBatchDialog.readable.js:144](source/frontend/NovelBatchDialog.readable.js#L144) | 静态已提取；运行未测 |
| UI-0347 | NovelBatchDialog / `min` | `1` | [NovelBatchDialog.readable.js:145](source/frontend/NovelBatchDialog.readable.js#L145) | 静态已提取；运行未测 |
| UI-0348 | NovelBatchDialog / `max` | `10` | [NovelBatchDialog.readable.js:146](source/frontend/NovelBatchDialog.readable.js#L146) | 静态已提取；运行未测 |
| UI-0349 | NovelBatchDialog / `onClick` | `d` | [NovelBatchDialog.readable.js:151](source/frontend/NovelBatchDialog.readable.js#L151) | 静态已提取；运行未测 |
| UI-0350 | NovelSchedulerPanel / `onClick` | `n[0] \|\| (n[0] = e => a("back"))` | [NovelSchedulerPanel.readable.js:184](source/frontend/NovelSchedulerPanel.readable.js#L184) | 静态已提取；运行未测 |
| UI-0351 | NovelSchedulerPanel / `onClick` | `h` | [NovelSchedulerPanel.readable.js:195](source/frontend/NovelSchedulerPanel.readable.js#L195) | 静态已提取；运行未测 |
| UI-0352 | NovelSchedulerPanel / `onClick` | `m` | [NovelSchedulerPanel.readable.js:205](source/frontend/NovelSchedulerPanel.readable.js#L205) | 静态已提取；运行未测 |
| UI-0353 | NovelSchedulerPanel / `onUpdate:value` | `n[1] \|\| (n[1] = e => i.value = e)` | [NovelSchedulerPanel.readable.js:211](source/frontend/NovelSchedulerPanel.readable.js#L211) | 静态已提取；运行未测 |
| UI-0354 | NovelSchedulerPanel / `min` | `1` | [NovelSchedulerPanel.readable.js:212](source/frontend/NovelSchedulerPanel.readable.js#L212) | 静态已提取；运行未测 |
| UI-0355 | NovelSchedulerPanel / `max` | `1440` | [NovelSchedulerPanel.readable.js:213](source/frontend/NovelSchedulerPanel.readable.js#L213) | 静态已提取；运行未测 |
| UI-0356 | NovelSchedulerPanel / `onUpdate:value` | `n[2] \|\| (n[2] = e => s.value = e)` | [NovelSchedulerPanel.readable.js:217](source/frontend/NovelSchedulerPanel.readable.js#L217) | 静态已提取；运行未测 |
| UI-0357 | NovelSchedulerPanel / `min` | `1` | [NovelSchedulerPanel.readable.js:218](source/frontend/NovelSchedulerPanel.readable.js#L218) | 静态已提取；运行未测 |
| UI-0358 | NovelSchedulerPanel / `max` | `100` | [NovelSchedulerPanel.readable.js:219](source/frontend/NovelSchedulerPanel.readable.js#L219) | 静态已提取；运行未测 |
| UI-0359 | NovelSchedulerPanel / `onUpdate:value` | `n[3] \|\| (n[3] = e => c.value = e)` | [NovelSchedulerPanel.readable.js:223](source/frontend/NovelSchedulerPanel.readable.js#L223) | 静态已提取；运行未测 |
| UI-0360 | NovelSchedulerPanel / `min` | `1` | [NovelSchedulerPanel.readable.js:224](source/frontend/NovelSchedulerPanel.readable.js#L224) | 静态已提取；运行未测 |
| UI-0361 | NovelSchedulerPanel / `max` | `10` | [NovelSchedulerPanel.readable.js:225](source/frontend/NovelSchedulerPanel.readable.js#L225) | 静态已提取；运行未测 |
| UI-0362 | NovelSchedulerPanel / `onClick` | `f` | [NovelSchedulerPanel.readable.js:231](source/frontend/NovelSchedulerPanel.readable.js#L231) | 静态已提取；运行未测 |
| UI-0363 | NovelHooksPanel / `onClick` | `t => n.value = e` | [NovelHooksPanel.readable.js:37](source/frontend/NovelHooksPanel.readable.js#L37) | 静态已提取；运行未测 |
| UI-0364 | NovelDashboard / `title` | ``第 ${t + 1} 章：${e} 字`` | [NovelDashboard.readable.js:60](source/frontend/NovelDashboard.readable.js#L60) | 静态已提取；运行未测 |
| UI-0365 | NovelCheckpointCard / `disabled` | `e.loading` | [NovelCheckpointCard.readable.js:32](source/frontend/NovelCheckpointCard.readable.js#L32) | 静态已提取；运行未测 |
| UI-0366 | NovelCheckpointCard / `onClick` | `n[0] \|\| (n[0] = e => a("confirm"))` | [NovelCheckpointCard.readable.js:33](source/frontend/NovelCheckpointCard.readable.js#L33) | 静态已提取；运行未测 |
| UI-0367 | NovelCheckpointCard / `disabled` | `e.loading` | [NovelCheckpointCard.readable.js:37](source/frontend/NovelCheckpointCard.readable.js#L37) | 静态已提取；运行未测 |
| UI-0368 | NovelCheckpointCard / `onClick` | `n[1] \|\| (n[1] = e => a("dismiss"))` | [NovelCheckpointCard.readable.js:38](source/frontend/NovelCheckpointCard.readable.js#L38) | 静态已提取；运行未测 |
| UI-0369 | NovelStudioPage / `title` | `a \|\| K9t[r] \|\| "新会话"` | [NovelStudioPage.readable.js:400](source/frontend/NovelStudioPage.readable.js#L400) | 静态已提取；运行未测 |
| UI-0370 | NovelStudioPage / `title` | `t.chapterTitle \|\| t.title` | [NovelStudioPage.readable.js:944](source/frontend/NovelStudioPage.readable.js#L944) | 静态已提取；运行未测 |
| UI-0371 | NovelStudioPage / `title` | `null == (i = e.data) ? void 0 : i.title` | [NovelStudioPage.readable.js:1148](source/frontend/NovelStudioPage.readable.js#L1148) | 静态已提取；运行未测 |
| UI-0372 | NovelStudioPage / `title` | `"选择导出格式"` | [NovelStudioPage.readable.js:1333](source/frontend/NovelStudioPage.readable.js#L1333) | 静态已提取；运行未测 |
| UI-0373 | NovelStudioPage / `onClick` | `() => async function (e, t) {<br>  a.destroyAll();<br>  try {<br>    const a = await async function (e, t = "md") {<br>      var n, a;<br>      const o = "NnW0",<br>        r = "kdlK",<br>        l = "S*tx";<br>      function i(e, t, n, a) {<br>        return Tan(0, 0, n - -140, a);<br>      }<br>      const s = {};<br>      return s[i(0, 0, 161, "ZU!e") + "ss"] = !1, (null == (a = null == (n = Ran()) ? void 0 : n[i(0, 0, 25, o) + i(0, 0, -84, r) + "ort_b" + (c = -33, u = l, Gan(0, u, c - -492))]) ? void 0 : a.call(n, e, t)) ?? s;<br>      var c, u;<br>    }(e.id, t.value);<br>    if (!(null == a ? void 0 : a.success)) return void n.error((null == a ? void 0 : a.message) \|\| "导出失败");<br>    const o = new Blob([a.data.content \|\| ""], {<br>        type: `${t.mime};charset=utf-8`<br>      }),<br>      r = URL.createObjectURL(o),<br>      l = document.createElement("a");<br>    l.href = r, l.download = a.data.filename \|\| `${e.title \|\| "book"}.${t.value}`, l.click(), URL.revokeObjectURL(r), n.success("导出完成");<br>  } catch {<br>    n.error("导出失败");<br>  }<br>}(e, t)` | [NovelStudioPage.readable.js:1338](source/frontend/NovelStudioPage.readable.js#L1338) | 静态已提取；运行未测 |
| UI-0374 | NovelStudioPage / `title` | `"删除作品"` | [NovelStudioPage.readable.js:1375](source/frontend/NovelStudioPage.readable.js#L1375) | 静态已提取；运行未测 |
| UI-0375 | NovelStudioPage / `onPositiveClick` | `async () => {<br>  const t = await async function (e) {<br>    var t, n;<br>    const a = "0UX(",<br>      o = 252,<br>      r = "j^1X",<br>      l = "!jb8";<br>    function i(e, t, n, a) {<br>      return Gan(0, a, e - -443);<br>    }<br>    const s = {};<br>    return s[i(7, 0, 0, "!jb8") + "ss"] = !1, (null == (n = null == (t = Ran()) ? void 0 : t[i(110, 0, 0, a) + i(o, 0, 0, r) + "ete_b" + i(240, 0, 0, l)]) ? void 0 : n.call(t, e)) ?? s;<br>  }(e.id);<br>  (null == t ? void 0 : t.success) ? (delete c[e.id], delete u[e.id], await ye(), W.value === e.id && we(), n.success("作品已删除")) : n.error((null == t ? void 0 : t.message) \|\| "删除失败");<br>}` | [NovelStudioPage.readable.js:1379](source/frontend/NovelStudioPage.readable.js#L1379) | 静态已提取；运行未测 |
| UI-0376 | NovelStudioPage / `onLaunchMode` | `Re` | [NovelStudioPage.readable.js:1537](source/frontend/NovelStudioPage.readable.js#L1537) | 静态已提取；运行未测 |
| UI-0377 | NovelStudioPage / `onSwitchWorkflow` | `_e` | [NovelStudioPage.readable.js:1538](source/frontend/NovelStudioPage.readable.js#L1538) | 静态已提取；运行未测 |
| UI-0378 | NovelStudioPage / `onCreateSession` | `Pe` | [NovelStudioPage.readable.js:1539](source/frontend/NovelStudioPage.readable.js#L1539) | 静态已提取；运行未测 |
| UI-0379 | NovelStudioPage / `onCreateBookSession` | `n[0] \|\| (n[0] = e => Pe("book", e))` | [NovelStudioPage.readable.js:1540](source/frontend/NovelStudioPage.readable.js#L1540) | 静态已提取；运行未测 |
| UI-0380 | NovelStudioPage / `onOpenSession` | `Ce` | [NovelStudioPage.readable.js:1541](source/frontend/NovelStudioPage.readable.js#L1541) | 静态已提取；运行未测 |
| UI-0381 | NovelStudioPage / `onOpenBook` | `Te` | [NovelStudioPage.readable.js:1542](source/frontend/NovelStudioPage.readable.js#L1542) | 静态已提取；运行未测 |
| UI-0382 | NovelStudioPage / `onToggleBook` | `ze` | [NovelStudioPage.readable.js:1543](source/frontend/NovelStudioPage.readable.js#L1543) | 静态已提取；运行未测 |
| UI-0383 | NovelStudioPage / `onRenameSession` | `Pt` | [NovelStudioPage.readable.js:1544](source/frontend/NovelStudioPage.readable.js#L1544) | 静态已提取；运行未测 |
| UI-0384 | NovelStudioPage / `onDeleteSession` | `It` | [NovelStudioPage.readable.js:1545](source/frontend/NovelStudioPage.readable.js#L1545) | 静态已提取；运行未测 |
| UI-0385 | NovelStudioPage / `onRenameBook` | `Rt` | [NovelStudioPage.readable.js:1546](source/frontend/NovelStudioPage.readable.js#L1546) | 静态已提取；运行未测 |
| UI-0386 | NovelStudioPage / `onDeleteBook` | `Ot` | [NovelStudioPage.readable.js:1547](source/frontend/NovelStudioPage.readable.js#L1547) | 静态已提取；运行未测 |
| UI-0387 | NovelStudioPage / `onExportBook` | `Ct` | [NovelStudioPage.readable.js:1548](source/frontend/NovelStudioPage.readable.js#L1548) | 静态已提取；运行未测 |
| UI-0388 | NovelStudioPage / `onOpenImport` | `Oe` | [NovelStudioPage.readable.js:1549](source/frontend/NovelStudioPage.readable.js#L1549) | 静态已提取；运行未测 |
| UI-0389 | NovelStudioPage / `onOpenStyle` | `Ie` | [NovelStudioPage.readable.js:1550](source/frontend/NovelStudioPage.readable.js#L1550) | 静态已提取；运行未测 |
| UI-0390 | NovelStudioPage / `onOpenImportModal` | `Me` | [NovelStudioPage.readable.js:1551](source/frontend/NovelStudioPage.readable.js#L1551) | 静态已提取；运行未测 |
| UI-0391 | NovelStudioPage / `onOpenImitationShort` | `n[1] \|\| (n[1] = e => z.value = !0)` | [NovelStudioPage.readable.js:1552](source/frontend/NovelStudioPage.readable.js#L1552) | 静态已提取；运行未测 |
| UI-0392 | NovelStudioPage / `onOpenAgentSettings` | `Ne` | [NovelStudioPage.readable.js:1553](source/frontend/NovelStudioPage.readable.js#L1553) | 静态已提取；运行未测 |
| UI-0393 | NovelStudioPage / `onOpenScheduler` | `Ge` | [NovelStudioPage.readable.js:1554](source/frontend/NovelStudioPage.readable.js#L1554) | 静态已提取；运行未测 |
| UI-0394 | NovelStudioPage / `onOpenDashboard` | `qe` | [NovelStudioPage.readable.js:1555](source/frontend/NovelStudioPage.readable.js#L1555) | 静态已提取；运行未测 |
| UI-0395 | NovelStudioPage / `onOpenMemo` | `Ae` | [NovelStudioPage.readable.js:1556](source/frontend/NovelStudioPage.readable.js#L1556) | 静态已提取；运行未测 |
| UI-0396 | NovelStudioPage / `onUpdate:selectedModel` | `n[2] \|\| (n[2] = e => P.value = e)` | [NovelStudioPage.readable.js:1571](source/frontend/NovelStudioPage.readable.js#L1571) | 静态已提取；运行未测 |
| UI-0397 | NovelStudioPage / `onUpdate:thinkingAutoExpand` | `n[3] \|\| (n[3] = e => R.value = e)` | [NovelStudioPage.readable.js:1574](source/frontend/NovelStudioPage.readable.js#L1574) | 静态已提取；运行未测 |
| UI-0398 | NovelStudioPage / `onSend` | `De` | [NovelStudioPage.readable.js:1577](source/frontend/NovelStudioPage.readable.js#L1577) | 静态已提取；运行未测 |
| UI-0399 | NovelStudioPage / `onClearContext` | `Qe` | [NovelStudioPage.readable.js:1578](source/frontend/NovelStudioPage.readable.js#L1578) | 静态已提取；运行未测 |
| UI-0400 | NovelStudioPage / `onConfirmAction` | `Ke` | [NovelStudioPage.readable.js:1579](source/frontend/NovelStudioPage.readable.js#L1579) | 静态已提取；运行未测 |
| UI-0401 | NovelStudioPage / `onDismissAction` | `He` | [NovelStudioPage.readable.js:1580](source/frontend/NovelStudioPage.readable.js#L1580) | 静态已提取；运行未测 |
| UI-0402 | NovelStudioPage / `onCancel` | `Je` | [NovelStudioPage.readable.js:1581](source/frontend/NovelStudioPage.readable.js#L1581) | 静态已提取；运行未测 |
| UI-0403 | NovelStudioPage / `onQuickWriteNext` | `yt` | [NovelStudioPage.readable.js:1582](source/frontend/NovelStudioPage.readable.js#L1582) | 静态已提取；运行未测 |
| UI-0404 | NovelStudioPage / `onQuickWriteBatch` | `bt` | [NovelStudioPage.readable.js:1583](source/frontend/NovelStudioPage.readable.js#L1583) | 静态已提取；运行未测 |
| UI-0405 | NovelStudioPage / `onQuickAudit` | `xt` | [NovelStudioPage.readable.js:1584](source/frontend/NovelStudioPage.readable.js#L1584) | 静态已提取；运行未测 |
| UI-0406 | NovelStudioPage / `onQuickRevise` | `St` | [NovelStudioPage.readable.js:1585](source/frontend/NovelStudioPage.readable.js#L1585) | 静态已提取；运行未测 |
| UI-0407 | NovelStudioPage / `onDeleteMessage` | `wt` | [NovelStudioPage.readable.js:1586](source/frontend/NovelStudioPage.readable.js#L1586) | 静态已提取；运行未测 |
| UI-0408 | NovelStudioPage / `onArchiveMessage` | `_t` | [NovelStudioPage.readable.js:1587](source/frontend/NovelStudioPage.readable.js#L1587) | 静态已提取；运行未测 |
| UI-0409 | NovelStudioPage / `onRetryAction` | `Ze` | [NovelStudioPage.readable.js:1588](source/frontend/NovelStudioPage.readable.js#L1588) | 静态已提取；运行未测 |
| UI-0410 | NovelStudioPage / `onClick` | `n[4] \|\| (n[4] = e => ee.value = !1)` | [NovelStudioPage.readable.js:1592](source/frontend/NovelStudioPage.readable.js#L1592) | 静态已提取；运行未测 |
| UI-0411 | NovelStudioPage / `onMousedown` | `ne` | [NovelStudioPage.readable.js:1597](source/frontend/NovelStudioPage.readable.js#L1597) | 静态已提取；运行未测 |
| UI-0412 | NovelStudioPage / `onRetry` | `Ze` | [NovelStudioPage.readable.js:1608](source/frontend/NovelStudioPage.readable.js#L1608) | 静态已提取；运行未测 |
| UI-0413 | NovelStudioPage / `onConfirm` | `Ve` | [NovelStudioPage.readable.js:1613](source/frontend/NovelStudioPage.readable.js#L1613) | 静态已提取；运行未测 |
| UI-0414 | NovelStudioPage / `onDismiss` | `Ue` | [NovelStudioPage.readable.js:1614](source/frontend/NovelStudioPage.readable.js#L1614) | 静态已提取；运行未测 |
| UI-0415 | NovelStudioPage / `onUpdate:collapsed` | `n[5] \|\| (n[5] = e => ee.value = e)` | [NovelStudioPage.readable.js:1620](source/frontend/NovelStudioPage.readable.js#L1620) | 静态已提取；运行未测 |
| UI-0416 | NovelStudioPage / `onSaved` | `ht` | [NovelStudioPage.readable.js:1622](source/frontend/NovelStudioPage.readable.js#L1622) | 静态已提取；运行未测 |
| UI-0417 | NovelStudioPage / `onUpdate:collapsed` | `n[6] \|\| (n[6] = e => ee.value = e)` | [NovelStudioPage.readable.js:1628](source/frontend/NovelStudioPage.readable.js#L1628) | 静态已提取；运行未测 |
| UI-0418 | NovelStudioPage / `onSaved` | `gt` | [NovelStudioPage.readable.js:1630](source/frontend/NovelStudioPage.readable.js#L1630) | 静态已提取；运行未测 |
| UI-0419 | NovelStudioPage / `onUpdate:selectedStyleIds` | `n[7] \|\| (n[7] = e => M.value = e)` | [NovelStudioPage.readable.js:1635](source/frontend/NovelStudioPage.readable.js#L1635) | 静态已提取；运行未测 |
| UI-0420 | NovelStudioPage / `onBack` | `ge` | [NovelStudioPage.readable.js:1637](source/frontend/NovelStudioPage.readable.js#L1637) | 静态已提取；运行未测 |
| UI-0421 | NovelStudioPage / `onRefresh` | `be` | [NovelStudioPage.readable.js:1638](source/frontend/NovelStudioPage.readable.js#L1638) | 静态已提取；运行未测 |
| UI-0422 | NovelStudioPage / `onApplied` | `We` | [NovelStudioPage.readable.js:1639](source/frontend/NovelStudioPage.readable.js#L1639) | 静态已提取；运行未测 |
| UI-0423 | NovelStudioPage / `onBack` | `ge` | [NovelStudioPage.readable.js:1644](source/frontend/NovelStudioPage.readable.js#L1644) | 静态已提取；运行未测 |
| UI-0424 | NovelStudioPage / `onClick` | `ge` | [NovelStudioPage.readable.js:1649](source/frontend/NovelStudioPage.readable.js#L1649) | 静态已提取；运行未测 |
| UI-0425 | NovelStudioPage / `onClick` | `ge` | [NovelStudioPage.readable.js:1659](source/frontend/NovelStudioPage.readable.js#L1659) | 静态已提取；运行未测 |
| UI-0426 | NovelStudioPage / `onUpdate:show` | `n[8] \|\| (n[8] = e => T.value = e)` | [NovelStudioPage.readable.js:1677](source/frontend/NovelStudioPage.readable.js#L1677) | 静态已提取；运行未测 |
| UI-0427 | NovelStudioPage / `onUpdate:tab` | `n[9] \|\| (n[9] = e => A.value = e)` | [NovelStudioPage.readable.js:1679](source/frontend/NovelStudioPage.readable.js#L1679) | 静态已提取；运行未测 |
| UI-0428 | NovelStudioPage / `onCreated` | `Xe` | [NovelStudioPage.readable.js:1684](source/frontend/NovelStudioPage.readable.js#L1684) | 静态已提取；运行未测 |
| UI-0429 | NovelStudioPage / `onSubmitStart` | `Fe` | [NovelStudioPage.readable.js:1685](source/frontend/NovelStudioPage.readable.js#L1685) | 静态已提取；运行未测 |
| UI-0430 | NovelStudioPage / `onUpdate:show` | `n[10] \|\| (n[10] = e => z.value = e)` | [NovelStudioPage.readable.js:1688](source/frontend/NovelStudioPage.readable.js#L1688) | 静态已提取；运行未测 |
| UI-0431 | NovelStudioPage / `onSubmit` | `ut` | [NovelStudioPage.readable.js:1693](source/frontend/NovelStudioPage.readable.js#L1693) | 静态已提取；运行未测 |
| UI-0432 | NovelStudioPage / `onUpdate:show` | `n[12] \|\| (n[12] = e => I.value = e)` | [NovelStudioPage.readable.js:1696](source/frontend/NovelStudioPage.readable.js#L1696) | 静态已提取；运行未测 |
| UI-0433 | NovelStudioPage / `title` | `"book" === q.value ? "重命名作品" : "重命名会话"` | [NovelStudioPage.readable.js:1698](source/frontend/NovelStudioPage.readable.js#L1698) | 静态已提取；运行未测 |
| UI-0434 | NovelStudioPage / `onPositiveClick` | `Mt` | [NovelStudioPage.readable.js:1700](source/frontend/NovelStudioPage.readable.js#L1700) | 静态已提取；运行未测 |
| UI-0435 | NovelStudioPage / `onUpdate:value` | `n[11] \|\| (n[11] = e => N.value = e)` | [NovelStudioPage.readable.js:1704](source/frontend/NovelStudioPage.readable.js#L1704) | 静态已提取；运行未测 |
| UI-0436 | NovelStudioPage / `placeholder` | `"book" === q.value ? "作品标题" : "会话标题"` | [NovelStudioPage.readable.js:1705](source/frontend/NovelStudioPage.readable.js#L1705) | 静态已提取；运行未测 |
| UI-0437 | NovelStudioPage / `onUpdate:show` | `n[13] \|\| (n[13] = e => oe.value = e)` | [NovelStudioPage.readable.js:1710](source/frontend/NovelStudioPage.readable.js#L1710) | 静态已提取；运行未测 |
| UI-0438 | NovelStudioPage / `onDone` | `kt` | [NovelStudioPage.readable.js:1715](source/frontend/NovelStudioPage.readable.js#L1715) | 静态已提取；运行未测 |
| UI-0439 | NovelStudioPage / `onClick` | `n[15] \|\| (n[15] = Zi(e => re.value = !1, ["self"]))` | [NovelStudioPage.readable.js:1724](source/frontend/NovelStudioPage.readable.js#L1724) | 静态已提取；运行未测 |
| UI-0440 | NovelStudioPage / `onUpdate:visible` | `n[14] \|\| (n[14] = e => re.value = e)` | [NovelStudioPage.readable.js:1733](source/frontend/NovelStudioPage.readable.js#L1733) | 静态已提取；运行未测 |

## 原版可见中文文案

| 组件 | 文案（仅静态字面量，动态拼接另查源码） | 来源 |
|---|---|---|
| NovelStudioSidebar | 全部 | [NovelStudioSidebar.readable.js:54](source/frontend/NovelStudioSidebar.readable.js#L54) |
| NovelStudioSidebar | 长篇原创 | [NovelStudioSidebar.readable.js:68](source/frontend/NovelStudioSidebar.readable.js#L68) |
| NovelStudioSidebar | 短篇原创 | [NovelStudioSidebar.readable.js:72](source/frontend/NovelStudioSidebar.readable.js#L72) |
| NovelStudioSidebar | 仿写长篇 | [NovelStudioSidebar.readable.js:76](source/frontend/NovelStudioSidebar.readable.js#L76) |
| NovelStudioSidebar | 仿写短篇 | [NovelStudioSidebar.readable.js:80](source/frontend/NovelStudioSidebar.readable.js#L80) |
| NovelStudioSidebar | 番外创作 | [NovelStudioSidebar.readable.js:84](source/frontend/NovelStudioSidebar.readable.js#L84) |
| NovelStudioSidebar | 续写创作 | [NovelStudioSidebar.readable.js:88](source/frontend/NovelStudioSidebar.readable.js#L88) |
| NovelStudioSidebar | 写作风格 | [NovelStudioSidebar.readable.js:93](source/frontend/NovelStudioSidebar.readable.js#L93) |
| NovelStudioSidebar | 导入素材 | [NovelStudioSidebar.readable.js:97](source/frontend/NovelStudioSidebar.readable.js#L97) |
| NovelStudioSidebar | Agent 设置 | [NovelStudioSidebar.readable.js:101](source/frontend/NovelStudioSidebar.readable.js#L101) |
| NovelStudioSidebar | 定时写作 | [NovelStudioSidebar.readable.js:105](source/frontend/NovelStudioSidebar.readable.js#L105) |
| NovelStudioSidebar | 数据看板 | [NovelStudioSidebar.readable.js:109](source/frontend/NovelStudioSidebar.readable.js#L109) |
| NovelStudioSidebar | 备忘录 | [NovelStudioSidebar.readable.js:113](source/frontend/NovelStudioSidebar.readable.js#L113) |
| NovelStudioSidebar | 新会话 | [NovelStudioSidebar.readable.js:216](source/frontend/NovelStudioSidebar.readable.js#L216) |
| NovelStudioSidebar | 刚刚 | [NovelStudioSidebar.readable.js:224](source/frontend/NovelStudioSidebar.readable.js#L224) |
| NovelStudioSidebar | 开始创作 | [NovelStudioSidebar.readable.js:232](source/frontend/NovelStudioSidebar.readable.js#L232) |
| NovelStudioSidebar | 我的作品 | [NovelStudioSidebar.readable.js:254](source/frontend/NovelStudioSidebar.readable.js#L254) |
| NovelStudioSidebar | 还没有作品 | [NovelStudioSidebar.readable.js:254](source/frontend/NovelStudioSidebar.readable.js#L254) |
| NovelStudioSidebar | 暂无该类型作品 | [NovelStudioSidebar.readable.js:254](source/frontend/NovelStudioSidebar.readable.js#L254) |
| NovelStudioSidebar | 重命名 | [NovelStudioSidebar.readable.js:276](source/frontend/NovelStudioSidebar.readable.js#L276) |
| NovelStudioSidebar | 导出 | [NovelStudioSidebar.readable.js:283](source/frontend/NovelStudioSidebar.readable.js#L283) |
| NovelStudioSidebar | 删除 | [NovelStudioSidebar.readable.js:290](source/frontend/NovelStudioSidebar.readable.js#L290) |
| NovelStudioSidebar |  建书后会话归档到这里  | [NovelStudioSidebar.readable.js:295](source/frontend/NovelStudioSidebar.readable.js#L295) |
| NovelStudioSidebar | 新建会话 | [NovelStudioSidebar.readable.js:312](source/frontend/NovelStudioSidebar.readable.js#L312) |
| NovelStudioSidebar | 会话记录 | [NovelStudioSidebar.readable.js:322](source/frontend/NovelStudioSidebar.readable.js#L322) |
| NovelStudioSidebar | 当前分类暂无会话，点击下方新建 | [NovelStudioSidebar.readable.js:322](source/frontend/NovelStudioSidebar.readable.js#L322) |
| NovelStudioSidebar | 新建会话 | [NovelStudioSidebar.readable.js:351](source/frontend/NovelStudioSidebar.readable.js#L351) |
| NovelConfirmCard | 确认执行 | [NovelConfirmCard.readable.js:30](source/frontend/NovelConfirmCard.readable.js#L30) |
| NovelConfirmCard | 确认创建作品 | [NovelConfirmCard.readable.js:33](source/frontend/NovelConfirmCard.readable.js#L33) |
| NovelConfirmCard | 确认写下一章 | [NovelConfirmCard.readable.js:34](source/frontend/NovelConfirmCard.readable.js#L34) |
| NovelConfirmCard | 确认审稿 | [NovelConfirmCard.readable.js:35](source/frontend/NovelConfirmCard.readable.js#L35) |
| NovelConfirmCard | 确认修订章节 | [NovelConfirmCard.readable.js:36](source/frontend/NovelConfirmCard.readable.js#L36) |
| NovelConfirmCard | 确认生成短篇 | [NovelConfirmCard.readable.js:37](source/frontend/NovelConfirmCard.readable.js#L37) |
| NovelConfirmCard | 确认启动互动世界 | [NovelConfirmCard.readable.js:38](source/frontend/NovelConfirmCard.readable.js#L38) |
| NovelConfirmCard | 确认生成封面 | [NovelConfirmCard.readable.js:39](source/frontend/NovelConfirmCard.readable.js#L39) |
| NovelConfirmCard | 打开仿写入口 | [NovelConfirmCard.readable.js:40](source/frontend/NovelConfirmCard.readable.js#L40) |
| NovelConfirmCard | 打开同人创作 | [NovelConfirmCard.readable.js:41](source/frontend/NovelConfirmCard.readable.js#L41) |
| NovelConfirmCard | 打开番外创作 | [NovelConfirmCard.readable.js:42](source/frontend/NovelConfirmCard.readable.js#L42) |
| NovelConfirmCard | 打开续写导入 | [NovelConfirmCard.readable.js:43](source/frontend/NovelConfirmCard.readable.js#L43) |
| NovelConfirmCard | 确认执行 | [NovelConfirmCard.readable.js:44](source/frontend/NovelConfirmCard.readable.js#L44) |
| NovelConfirmCard | 世界契约 | [NovelConfirmCard.readable.js:64](source/frontend/NovelConfirmCard.readable.js#L64) |
| NovelConfirmCard | 视觉契约 | [NovelConfirmCard.readable.js:67](source/frontend/NovelConfirmCard.readable.js#L67) |
| NovelConfirmCard | 执行中… | [NovelConfirmCard.readable.js:73](source/frontend/NovelConfirmCard.readable.js#L73) |
| NovelConfirmCard | 打开入口 | [NovelConfirmCard.readable.js:73](source/frontend/NovelConfirmCard.readable.js#L73) |
| NovelConfirmCard | 继续执行 | [NovelConfirmCard.readable.js:73](source/frontend/NovelConfirmCard.readable.js#L73) |
| NovelConfirmCard | 已打开 | [NovelConfirmCard.readable.js:77](source/frontend/NovelConfirmCard.readable.js#L77) |
| NovelConfirmCard | 已执行 | [NovelConfirmCard.readable.js:77](source/frontend/NovelConfirmCard.readable.js#L77) |
| NovelConfirmCard | 已取消 | [NovelConfirmCard.readable.js:92](source/frontend/NovelConfirmCard.readable.js#L92) |
| NovelConfirmCard | 取消 | [NovelConfirmCard.readable.js:102](source/frontend/NovelConfirmCard.readable.js#L102) |
| NovelCoverPreview | 封面 | [NovelCoverPreview.readable.js:13](source/frontend/NovelCoverPreview.readable.js#L13) |
| NovelCoverPreview | 图片加载失败 | [NovelCoverPreview.readable.js:119](source/frontend/NovelCoverPreview.readable.js#L119) |
| NovelToolSteps | 执行中 | [NovelToolSteps.readable.js:58](source/frontend/NovelToolSteps.readable.js#L58) |
| NovelToolSteps | 已完成 | [NovelToolSteps.readable.js:58](source/frontend/NovelToolSteps.readable.js#L58) |
| NovelToolSteps | 失败 | [NovelToolSteps.readable.js:58](source/frontend/NovelToolSteps.readable.js#L58) |
| NovelToolSteps | 等待中 | [NovelToolSteps.readable.js:58](source/frontend/NovelToolSteps.readable.js#L58) |
| NovelToolSteps | 任务 | [NovelToolSteps.readable.js:67](source/frontend/NovelToolSteps.readable.js#L67) |
| NovelToolSteps | 思考中 | [NovelToolSteps.readable.js:115](source/frontend/NovelToolSteps.readable.js#L115) |
| NovelToolSteps | AI 正在生成… | [NovelToolSteps.readable.js:184](source/frontend/NovelToolSteps.readable.js#L184) |
| NovelToolSteps | 重试 | [NovelToolSteps.readable.js:203](source/frontend/NovelToolSteps.readable.js#L203) |
| NovelToolSteps | 重试 | [NovelToolSteps.readable.js:207](source/frontend/NovelToolSteps.readable.js#L207) |
| NovelToolSteps | 封面预览 | [NovelToolSteps.readable.js:210](source/frontend/NovelToolSteps.readable.js#L210) |
| NovelThinkingBlock | 思考了  | [NovelThinkingBlock.readable.js:49](source/frontend/NovelThinkingBlock.readable.js#L49) |
| NovelThinkingBlock |  秒 | [NovelThinkingBlock.readable.js:49](source/frontend/NovelThinkingBlock.readable.js#L49) |
| NovelThinkingBlock | 暂无思考内容记录 | [NovelThinkingBlock.readable.js:49](source/frontend/NovelThinkingBlock.readable.js#L49) |
| NovelStreamingIndicator | 思考中…… | [NovelStreamingIndicator.readable.js:20](source/frontend/NovelStreamingIndicator.readable.js#L20) |
| NovelQuickActions | 写下一章 | [NovelQuickActions.readable.js:18](source/frontend/NovelQuickActions.readable.js#L18) |
| NovelQuickActions | 批量写作 | [NovelQuickActions.readable.js:22](source/frontend/NovelQuickActions.readable.js#L22) |
| NovelQuickActions | 审稿 | [NovelQuickActions.readable.js:26](source/frontend/NovelQuickActions.readable.js#L26) |
| NovelQuickActions | 修订 | [NovelQuickActions.readable.js:30](source/frontend/NovelQuickActions.readable.js#L30) |
| NovelStudioChat | 无限 | [NovelStudioChat.readable.js:68](source/frontend/NovelStudioChat.readable.js#L68) |
| NovelStudioChat | 任务 | [NovelStudioChat.readable.js:92](source/frontend/NovelStudioChat.readable.js#L92) |
| NovelStudioChat | 选择模型 | [NovelStudioChat.readable.js:212](source/frontend/NovelStudioChat.readable.js#L212) |
| NovelStudioChat | 输入指令… 可在模型列表中设置思考强度 (Enter 发送，Shift+Enter 换行) | [NovelStudioChat.readable.js:229](source/frontend/NovelStudioChat.readable.js#L229) |
| NovelStudioChat | 上下文已清除 | [NovelStudioChat.readable.js:246](source/frontend/NovelStudioChat.readable.js#L246) |
| NovelStudioChat | 复制失败 | [NovelStudioChat.readable.js:267](source/frontend/NovelStudioChat.readable.js#L267) |
| NovelStudioChat | 请等待当前回复完成 | [NovelStudioChat.readable.js:277](source/frontend/NovelStudioChat.readable.js#L277) |
| NovelStudioChat | 复制失败 | [NovelStudioChat.readable.js:289](source/frontend/NovelStudioChat.readable.js#L289) |
| NovelStudioChat | 告诉我你想写什么——题材、世界观、主角、核心冲突。 | [NovelStudioChat.readable.js:331](source/frontend/NovelStudioChat.readable.js#L331) |
| NovelStudioChat |  复制  | [NovelStudioChat.readable.js:356](source/frontend/NovelStudioChat.readable.js#L356) |
| NovelStudioChat |  存档  | [NovelStudioChat.readable.js:370](source/frontend/NovelStudioChat.readable.js#L370) |
| NovelStudioChat | 删除 | [NovelStudioChat.readable.js:373](source/frontend/NovelStudioChat.readable.js#L373) |
| NovelStudioChat | 取消 | [NovelStudioChat.readable.js:374](source/frontend/NovelStudioChat.readable.js#L374) |
| NovelStudioChat |  删除  | [NovelStudioChat.readable.js:389](source/frontend/NovelStudioChat.readable.js#L389) |
| NovelStudioChat |  确定删除此消息 | [NovelStudioChat.readable.js:394](source/frontend/NovelStudioChat.readable.js#L394) |
| NovelStudioChat | 及对应回复 | [NovelStudioChat.readable.js:394](source/frontend/NovelStudioChat.readable.js#L394) |
| NovelStudioChat |  复制  | [NovelStudioChat.readable.js:437](source/frontend/NovelStudioChat.readable.js#L437) |
| NovelStudioChat |  存档  | [NovelStudioChat.readable.js:451](source/frontend/NovelStudioChat.readable.js#L451) |
| NovelStudioChat | 删除 | [NovelStudioChat.readable.js:454](source/frontend/NovelStudioChat.readable.js#L454) |
| NovelStudioChat | 取消 | [NovelStudioChat.readable.js:455](source/frontend/NovelStudioChat.readable.js#L455) |
| NovelStudioChat |  删除  | [NovelStudioChat.readable.js:470](source/frontend/NovelStudioChat.readable.js#L470) |
| NovelStudioChat |  确定删除这条 AI 回复？  | [NovelStudioChat.readable.js:473](source/frontend/NovelStudioChat.readable.js#L473) |
| NovelStudioChat |  复制  | [NovelStudioChat.readable.js:514](source/frontend/NovelStudioChat.readable.js#L514) |
| NovelStudioChat | 展开思考 | [NovelStudioChat.readable.js:620](source/frontend/NovelStudioChat.readable.js#L620) |
| NovelStudioChat | 展开思考 | [NovelStudioChat.readable.js:638](source/frontend/NovelStudioChat.readable.js#L638) |
| NovelStudioChat |  上下文长度:  | [NovelStudioChat.readable.js:662](source/frontend/NovelStudioChat.readable.js#L662) |
| NovelStudioChat | 上下文长度 | [NovelStudioChat.readable.js:669](source/frontend/NovelStudioChat.readable.js#L669) |
| NovelStudioChat |  清除上下文 (Ctrl+K)  | [NovelStudioChat.readable.js:712](source/frontend/NovelStudioChat.readable.js#L712) |
| NovelStudioChat | 搜索模型... | [NovelStudioChat.readable.js:761](source/frontend/NovelStudioChat.readable.js#L761) |
| NovelStudioChat | 无匹配模型 | [NovelStudioChat.readable.js:804](source/frontend/NovelStudioChat.readable.js#L804) |
| NovelPanelArtifactView | 返回 | [NovelPanelArtifactView.readable.js:51](source/frontend/NovelPanelArtifactView.readable.js#L51) |
| NovelPanelArtifactView | 资料详情 | [NovelPanelArtifactView.readable.js:56](source/frontend/NovelPanelArtifactView.readable.js#L56) |
| NovelPanelArtifactView | 保存 | [NovelPanelArtifactView.readable.js:63](source/frontend/NovelPanelArtifactView.readable.js#L63) |
| NovelPanelArtifactView | 取消 | [NovelPanelArtifactView.readable.js:76](source/frontend/NovelPanelArtifactView.readable.js#L76) |
| NovelPanelArtifactView | 编辑 | [NovelPanelArtifactView.readable.js:85](source/frontend/NovelPanelArtifactView.readable.js#L85) |
| NovelPanelArtifactView | 文件不存在 | [NovelPanelArtifactView.readable.js:95](source/frontend/NovelPanelArtifactView.readable.js#L95) |
| NovelStudioBookPanel | 故事基石 / 世界观 | [NovelStudioBookPanel.readable.js:147](source/frontend/NovelStudioBookPanel.readable.js#L147) |
| NovelStudioBookPanel | 保存失败 | [NovelStudioBookPanel.readable.js:189](source/frontend/NovelStudioBookPanel.readable.js#L189) |
| NovelStudioBookPanel | 已保存并同步真相 | [NovelStudioBookPanel.readable.js:190](source/frontend/NovelStudioBookPanel.readable.js#L190) |
| NovelStudioBookPanel | 已保存 | [NovelStudioBookPanel.readable.js:190](source/frontend/NovelStudioBookPanel.readable.js#L190) |
| NovelStudioBookPanel | 已写  | [NovelStudioBookPanel.readable.js:207](source/frontend/NovelStudioBookPanel.readable.js#L207) |
| NovelStudioBookPanel |  章 | [NovelStudioBookPanel.readable.js:207](source/frontend/NovelStudioBookPanel.readable.js#L207) |
| NovelStudioBookPanel | 章节 | [NovelStudioBookPanel.readable.js:221](source/frontend/NovelStudioBookPanel.readable.js#L221) |
| NovelStudioBookPanel | 第  | [NovelStudioBookPanel.readable.js:235](source/frontend/NovelStudioBookPanel.readable.js#L235) |
| NovelStudioBookPanel |  章  | [NovelStudioBookPanel.readable.js:235](source/frontend/NovelStudioBookPanel.readable.js#L235) |
| NovelStudioBookPanel | 编辑 | [NovelStudioBookPanel.readable.js:238](source/frontend/NovelStudioBookPanel.readable.js#L238) |
| NovelStudioBookPanel | 暂无章节 | [NovelStudioBookPanel.readable.js:250](source/frontend/NovelStudioBookPanel.readable.js#L250) |
| NovelStudioBookPanel | 角色 | [NovelStudioBookPanel.readable.js:257](source/frontend/NovelStudioBookPanel.readable.js#L257) |
| NovelStudioBookPanel | 主要 | [NovelStudioBookPanel.readable.js:274](source/frontend/NovelStudioBookPanel.readable.js#L274) |
| NovelStudioBookPanel | 次要 | [NovelStudioBookPanel.readable.js:274](source/frontend/NovelStudioBookPanel.readable.js#L274) |
| NovelStudioBookPanel | 暂无角色档案 | [NovelStudioBookPanel.readable.js:276](source/frontend/NovelStudioBookPanel.readable.js#L276) |
| NovelStudioBookPanel | 核心文件 | [NovelStudioBookPanel.readable.js:283](source/frontend/NovelStudioBookPanel.readable.js#L283) |
| NovelStudioBookPanel | 编辑 | [NovelStudioBookPanel.readable.js:300](source/frontend/NovelStudioBookPanel.readable.js#L300) |
| NovelStudioBookPanel | 暂无核心文件 | [NovelStudioBookPanel.readable.js:312](source/frontend/NovelStudioBookPanel.readable.js#L312) |
| NovelStudioBookPanel | 封面资源 | [NovelStudioBookPanel.readable.js:319](source/frontend/NovelStudioBookPanel.readable.js#L319) |
| NovelStudioBookPanel | 世界观 | [NovelStudioBookPanel.readable.js:333](source/frontend/NovelStudioBookPanel.readable.js#L333) |
| NovelStudioBookPanel | 暂无世界观摘要 | [NovelStudioBookPanel.readable.js:333](source/frontend/NovelStudioBookPanel.readable.js#L333) |
| NovelStudioBookPanel | 查看完整设定 → | [NovelStudioBookPanel.readable.js:337](source/frontend/NovelStudioBookPanel.readable.js#L337) |
| NovelStudioShortPanel | 已完成 | [NovelStudioShortPanel.readable.js:62](source/frontend/NovelStudioShortPanel.readable.js#L62) |
| NovelStudioShortPanel | 草稿 | [NovelStudioShortPanel.readable.js:62](source/frontend/NovelStudioShortPanel.readable.js#L62) |
| NovelStudioShortPanel | 进行中 | [NovelStudioShortPanel.readable.js:62](source/frontend/NovelStudioShortPanel.readable.js#L62) |
| NovelStudioShortPanel | 保存失败 | [NovelStudioShortPanel.readable.js:151](source/frontend/NovelStudioShortPanel.readable.js#L151) |
| NovelStudioShortPanel | 已保存 | [NovelStudioShortPanel.readable.js:152](source/frontend/NovelStudioShortPanel.readable.js#L152) |
| NovelStudioShortPanel | 文件 | [NovelStudioShortPanel.readable.js:161](source/frontend/NovelStudioShortPanel.readable.js#L161) |
| NovelStudioShortPanel | 短篇大纲 | [NovelStudioShortPanel.readable.js:169](source/frontend/NovelStudioShortPanel.readable.js#L169) |
| NovelStudioShortPanel | 推荐 | [NovelStudioShortPanel.readable.js:175](source/frontend/NovelStudioShortPanel.readable.js#L175) |
| NovelStudioShortPanel |  · 约  | [NovelStudioShortPanel.readable.js:189](source/frontend/NovelStudioShortPanel.readable.js#L189) |
| NovelStudioShortPanel |  字  | [NovelStudioShortPanel.readable.js:189](source/frontend/NovelStudioShortPanel.readable.js#L189) |
| NovelStudioShortPanel | 封面图 | [NovelStudioShortPanel.readable.js:198](source/frontend/NovelStudioShortPanel.readable.js#L198) |
| NovelStudioShortPanel | 封面图 | [NovelStudioShortPanel.readable.js:200](source/frontend/NovelStudioShortPanel.readable.js#L200) |
| NovelStudioShortPanel | 章节结构 | [NovelStudioShortPanel.readable.js:209](source/frontend/NovelStudioShortPanel.readable.js#L209) |
| NovelStudioShortPanel |  字 | [NovelStudioShortPanel.readable.js:220](source/frontend/NovelStudioShortPanel.readable.js#L220) |
| NovelStudioShortPanel | 成稿生成后可查看章节结构 | [NovelStudioShortPanel.readable.js:220](source/frontend/NovelStudioShortPanel.readable.js#L220) |
| NovelStudioShortPanel | 大纲 | [NovelStudioShortPanel.readable.js:227](source/frontend/NovelStudioShortPanel.readable.js#L227) |
| NovelStudioShortPanel | 大纲生成后显示摘要 | [NovelStudioShortPanel.readable.js:227](source/frontend/NovelStudioShortPanel.readable.js#L227) |
| NovelStudioShortPanel |  查看完整大纲 →  | [NovelStudioShortPanel.readable.js:232](source/frontend/NovelStudioShortPanel.readable.js#L232) |
| NovelStudioShortPanel | 成品文件 | [NovelStudioShortPanel.readable.js:239](source/frontend/NovelStudioShortPanel.readable.js#L239) |
| NovelStudioShortPanel | 编辑 | [NovelStudioShortPanel.readable.js:250](source/frontend/NovelStudioShortPanel.readable.js#L250) |
| NovelStudioShortPanel | 文件 | [NovelStudioShortPanel.readable.js:254](source/frontend/NovelStudioShortPanel.readable.js#L254) |
| NovelStudioShortPanel | 暂无成品文件 | [NovelStudioShortPanel.readable.js:261](source/frontend/NovelStudioShortPanel.readable.js#L261) |
| NovelStudioShortPanel | 推荐 | [NovelStudioShortPanel.readable.js:268](source/frontend/NovelStudioShortPanel.readable.js#L268) |
| NovelStudioShortPanel | 查看完整推荐 → | [NovelStudioShortPanel.readable.js:272](source/frontend/NovelStudioShortPanel.readable.js#L272) |
| NovelStylePanel | 加载文风失败 | [NovelStylePanel.readable.js:58](source/frontend/NovelStylePanel.readable.js#L58) |
| NovelStylePanel | 请填写风格名称 | [NovelStylePanel.readable.js:62](source/frontend/NovelStylePanel.readable.js#L62) |
| NovelStylePanel | 请粘贴参考文本，或导入文件 | [NovelStylePanel.readable.js:63](source/frontend/NovelStylePanel.readable.js#L63) |
| NovelStylePanel | 保存失败 | [NovelStylePanel.readable.js:89](source/frontend/NovelStylePanel.readable.js#L89) |
| NovelStylePanel | 风格已保存，可继续新建下一条 | [NovelStylePanel.readable.js:91](source/frontend/NovelStylePanel.readable.js#L91) |
| NovelStylePanel | 风格已更新 | [NovelStylePanel.readable.js:91](source/frontend/NovelStylePanel.readable.js#L91) |
| NovelStylePanel | 导入文件 | [NovelStylePanel.readable.js:130](source/frontend/NovelStylePanel.readable.js#L130) |
| NovelStylePanel | 导入失败 | [NovelStylePanel.readable.js:165](source/frontend/NovelStylePanel.readable.js#L165) |
| NovelStylePanel | 已删除 | [NovelStylePanel.readable.js:190](source/frontend/NovelStylePanel.readable.js#L190) |
| NovelStylePanel | 删除失败 | [NovelStylePanel.readable.js:190](source/frontend/NovelStylePanel.readable.js#L190) |
| NovelStylePanel | ← 返回对话 | [NovelStylePanel.readable.js:209](source/frontend/NovelStylePanel.readable.js#L209) |
| NovelStylePanel | 写作风格 | [NovelStylePanel.readable.js:209](source/frontend/NovelStylePanel.readable.js#L209) |
| NovelStylePanel | 在此管理写作风格知识库：新建名称、粘贴参考文本，或导入 txt / md / docx 等文件（可多选）。 | [NovelStylePanel.readable.js:209](source/frontend/NovelStylePanel.readable.js#L209) |
| NovelStylePanel | 编辑风格 | [NovelStylePanel.readable.js:212](source/frontend/NovelStylePanel.readable.js#L212) |
| NovelStylePanel | 新建风格 | [NovelStylePanel.readable.js:212](source/frontend/NovelStylePanel.readable.js#L212) |
| NovelStylePanel | + 新建 | [NovelStylePanel.readable.js:217](source/frontend/NovelStylePanel.readable.js#L217) |
| NovelStylePanel | 风格名称 | [NovelStylePanel.readable.js:217](source/frontend/NovelStylePanel.readable.js#L217) |
| NovelStylePanel | 例如：金庸武侠、知乎盐选、某作者口语风 | [NovelStylePanel.readable.js:220](source/frontend/NovelStylePanel.readable.js#L220) |
| NovelStylePanel | 参考文本（可粘贴） | [NovelStylePanel.readable.js:221](source/frontend/NovelStylePanel.readable.js#L221) |
| NovelStylePanel | 直接粘贴样章、段落或写作要求…… | [NovelStylePanel.readable.js:229](source/frontend/NovelStylePanel.readable.js#L229) |
| NovelStylePanel | 文风指南（可选，留空则保存时自动分析） | [NovelStylePanel.readable.js:230](source/frontend/NovelStylePanel.readable.js#L230) |
| NovelStylePanel | 可手写可执行的文风规则；不填则根据参考文本自动生成 | [NovelStylePanel.readable.js:238](source/frontend/NovelStylePanel.readable.js#L238) |
| NovelStylePanel | 导入文件 | [NovelStylePanel.readable.js:243](source/frontend/NovelStylePanel.readable.js#L243) |
| NovelStylePanel | 保存 | [NovelStylePanel.readable.js:250](source/frontend/NovelStylePanel.readable.js#L250) |
| NovelStylePanel | 已有风格 | [NovelStylePanel.readable.js:252](source/frontend/NovelStylePanel.readable.js#L252) |
| NovelStylePanel | 无参考文本 | [NovelStylePanel.readable.js:267](source/frontend/NovelStylePanel.readable.js#L267) |
| NovelStylePanel |  个文件 | [NovelStylePanel.readable.js:267](source/frontend/NovelStylePanel.readable.js#L267) |
| NovelStylePanel | 删除 | [NovelStylePanel.readable.js:271](source/frontend/NovelStylePanel.readable.js#L271) |
| NovelStylePanel | 还没有风格。请在上方新建并保存。 | [NovelStylePanel.readable.js:274](source/frontend/NovelStylePanel.readable.js#L274) |
| NovelStylePanel | 已选  | [NovelStylePanel.readable.js:276](source/frontend/NovelStylePanel.readable.js#L276) |
| NovelStylePanel |  项用于对话参考 | [NovelStylePanel.readable.js:276](source/frontend/NovelStylePanel.readable.js#L276) |
| NovelStylePanel | 应用 | [NovelStylePanel.readable.js:280](source/frontend/NovelStylePanel.readable.js#L280) |
| NovelAgentSettings | 跟随对话默认模型 | [NovelAgentSettings.readable.js:37](source/frontend/NovelAgentSettings.readable.js#L37) |
| NovelAgentSettings | 跟随对话默认模型 | [NovelAgentSettings.readable.js:39](source/frontend/NovelAgentSettings.readable.js#L39) |
| NovelAgentSettings | 保存失败 | [NovelAgentSettings.readable.js:161](source/frontend/NovelAgentSettings.readable.js#L161) |
| NovelAgentSettings | 已保存 | [NovelAgentSettings.readable.js:179](source/frontend/NovelAgentSettings.readable.js#L179) |
| NovelAgentSettings | 自定义审稿 | [NovelAgentSettings.readable.js:188](source/frontend/NovelAgentSettings.readable.js#L188) |
| NovelAgentSettings | 自定义写作要求 | [NovelAgentSettings.readable.js:211](source/frontend/NovelAgentSettings.readable.js#L211) |
| NovelAgentSettings | 长篇正文 | [NovelAgentSettings.readable.js:218](source/frontend/NovelAgentSettings.readable.js#L218) |
| NovelAgentSettings | 短篇小说 | [NovelAgentSettings.readable.js:221](source/frontend/NovelAgentSettings.readable.js#L221) |
| NovelAgentSettings | 仿写创作 | [NovelAgentSettings.readable.js:224](source/frontend/NovelAgentSettings.readable.js#L224) |
| NovelAgentSettings | 加载 Agent 设置失败 | [NovelAgentSettings.readable.js:250](source/frontend/NovelAgentSettings.readable.js#L250) |
| NovelAgentSettings | ← 返回对话 | [NovelAgentSettings.readable.js:280](source/frontend/NovelAgentSettings.readable.js#L280) |
| NovelAgentSettings | Agent 设置 | [NovelAgentSettings.readable.js:280](source/frontend/NovelAgentSettings.readable.js#L280) |
| NovelAgentSettings | 为各节点指定模型路由（如写作用好模型、审稿用便宜模型）。审稿维度与自定义提示词会注入审稿节点。 | [NovelAgentSettings.readable.js:280](source/frontend/NovelAgentSettings.readable.js#L280) |
| NovelAgentSettings | 模型路由 | [NovelAgentSettings.readable.js:283](source/frontend/NovelAgentSettings.readable.js#L283) |
| NovelAgentSettings |  留空则使用对话区当前选择的默认模型。封面配图使用出图模型，与对话 LLM 无关。 Codex 出图需先在技能页开启 Codex 渠道并完成验证。  | [NovelAgentSettings.readable.js:285](source/frontend/NovelAgentSettings.readable.js#L285) |
| NovelAgentSettings | 加载中… | [NovelAgentSettings.readable.js:285](source/frontend/NovelAgentSettings.readable.js#L285) |
| NovelAgentSettings | 跟随默认 | [NovelAgentSettings.readable.js:293](source/frontend/NovelAgentSettings.readable.js#L293) |
| NovelAgentSettings | 默认 gemini 出图 | [NovelAgentSettings.readable.js:303](source/frontend/NovelAgentSettings.readable.js#L303) |
| NovelAgentSettings | 写章审稿循环 | [NovelAgentSettings.readable.js:308](source/frontend/NovelAgentSettings.readable.js#L308) |
| NovelAgentSettings | 写下一章时自动审稿-修订循环。未达通过分将尝试修订，修订后仍可能回退到更高分版本。 | [NovelAgentSettings.readable.js:310](source/frontend/NovelAgentSettings.readable.js#L310) |
| NovelAgentSettings | 最大修订轮次 | [NovelAgentSettings.readable.js:310](source/frontend/NovelAgentSettings.readable.js#L310) |
| NovelAgentSettings | 通过分数 | [NovelAgentSettings.readable.js:316](source/frontend/NovelAgentSettings.readable.js#L316) |
| NovelAgentSettings | 升温步长 | [NovelAgentSettings.readable.js:322](source/frontend/NovelAgentSettings.readable.js#L322) |
| NovelAgentSettings | 基础温度 | [NovelAgentSettings.readable.js:329](source/frontend/NovelAgentSettings.readable.js#L329) |
| NovelAgentSettings | 字数归一化 | [NovelAgentSettings.readable.js:340](source/frontend/NovelAgentSettings.readable.js#L340) |
| NovelAgentSettings | 草稿字数偏离目标 >45% 时自动扩写/压缩 | [NovelAgentSettings.readable.js:342](source/frontend/NovelAgentSettings.readable.js#L342) |
| NovelAgentSettings | 后写校验 | [NovelAgentSettings.readable.js:346](source/frontend/NovelAgentSettings.readable.js#L346) |
| NovelAgentSettings | 确定性规则检查（零 LLM 开销） | [NovelAgentSettings.readable.js:348](source/frontend/NovelAgentSettings.readable.js#L348) |
| NovelAgentSettings | AI 味检测 | [NovelAgentSettings.readable.js:352](source/frontend/NovelAgentSettings.readable.js#L352) |
| NovelAgentSettings | 段落等长/套话密度/公式化转折（零 LLM 开销） | [NovelAgentSettings.readable.js:354](source/frontend/NovelAgentSettings.readable.js#L354) |
| NovelAgentSettings | 生产二次确认 | [NovelAgentSettings.readable.js:358](source/frontend/NovelAgentSettings.readable.js#L358) |
| NovelAgentSettings | 大纲/蓝图/基础设定生成后暂停，可在右侧修改文件后再继续 | [NovelAgentSettings.readable.js:360](source/frontend/NovelAgentSettings.readable.js#L360) |
| NovelAgentSettings | 审稿维度（37 项） | [NovelAgentSettings.readable.js:360](source/frontend/NovelAgentSettings.readable.js#L360) |
| NovelAgentSettings |  勾选后即参与审稿；维度说明为内置文案，不可改字，但可自由开关。当前已选  | [NovelAgentSettings.readable.js:360](source/frontend/NovelAgentSettings.readable.js#L360) |
| NovelAgentSettings |  项。  | [NovelAgentSettings.readable.js:360](source/frontend/NovelAgentSettings.readable.js#L360) |
| NovelAgentSettings | 全选 | [NovelAgentSettings.readable.js:365](source/frontend/NovelAgentSettings.readable.js#L365) |
| NovelAgentSettings | 全不选 | [NovelAgentSettings.readable.js:372](source/frontend/NovelAgentSettings.readable.js#L372) |
| NovelAgentSettings | 自定义审稿 | [NovelAgentSettings.readable.js:385](source/frontend/NovelAgentSettings.readable.js#L385) |
| NovelAgentSettings | 勾选后会在审稿时追加到维度列表之后，可自由编写补充要求。 | [NovelAgentSettings.readable.js:387](source/frontend/NovelAgentSettings.readable.js#L387) |
| NovelAgentSettings | + 新增 | [NovelAgentSettings.readable.js:392](source/frontend/NovelAgentSettings.readable.js#L392) |
| NovelAgentSettings | 暂无自定义项，可点击「新增」添加。 | [NovelAgentSettings.readable.js:394](source/frontend/NovelAgentSettings.readable.js#L394) |
| NovelAgentSettings | 提示词标题 | [NovelAgentSettings.readable.js:403](source/frontend/NovelAgentSettings.readable.js#L403) |
| NovelAgentSettings | 删除 | [NovelAgentSettings.readable.js:412](source/frontend/NovelAgentSettings.readable.js#L412) |
| NovelAgentSettings | 审稿要求说明… | [NovelAgentSettings.readable.js:419](source/frontend/NovelAgentSettings.readable.js#L419) |
| NovelAgentSettings | 自定义写作提示词 | [NovelAgentSettings.readable.js:421](source/frontend/NovelAgentSettings.readable.js#L421) |
| NovelAgentSettings | 针对不同写作模式的自定义要求。勾选后会注入到对应写作 Agent 的系统提示中。 | [NovelAgentSettings.readable.js:423](source/frontend/NovelAgentSettings.readable.js#L423) |
| NovelAgentSettings | + 新增 | [NovelAgentSettings.readable.js:428](source/frontend/NovelAgentSettings.readable.js#L428) |
| NovelAgentSettings | 暂无自定义写作提示词。 | [NovelAgentSettings.readable.js:430](source/frontend/NovelAgentSettings.readable.js#L430) |
| NovelAgentSettings | 提示词标题 | [NovelAgentSettings.readable.js:439](source/frontend/NovelAgentSettings.readable.js#L439) |
| NovelAgentSettings | 删除 | [NovelAgentSettings.readable.js:456](source/frontend/NovelAgentSettings.readable.js#L456) |
| NovelAgentSettings | 写作要求说明…… | [NovelAgentSettings.readable.js:463](source/frontend/NovelAgentSettings.readable.js#L463) |
| NovelAgentSettings | 设置保存在本地数据目录，永久生效。 | [NovelAgentSettings.readable.js:469](source/frontend/NovelAgentSettings.readable.js#L469) |
| NovelAgentSettings | 保存设置 | [NovelAgentSettings.readable.js:475](source/frontend/NovelAgentSettings.readable.js#L475) |
| NovelImportPage | 玄幻 | [NovelImportPage.readable.js:61](source/frontend/NovelImportPage.readable.js#L61) |
| NovelImportPage | 都市 | [NovelImportPage.readable.js:61](source/frontend/NovelImportPage.readable.js#L61) |
| NovelImportPage | 仙侠 | [NovelImportPage.readable.js:61](source/frontend/NovelImportPage.readable.js#L61) |
| NovelImportPage | 科幻 | [NovelImportPage.readable.js:61](source/frontend/NovelImportPage.readable.js#L61) |
| NovelImportPage | 历史 | [NovelImportPage.readable.js:61](source/frontend/NovelImportPage.readable.js#L61) |
| NovelImportPage | 悬疑 | [NovelImportPage.readable.js:61](source/frontend/NovelImportPage.readable.js#L61) |
| NovelImportPage | 言情 | [NovelImportPage.readable.js:61](source/frontend/NovelImportPage.readable.js#L61) |
| NovelImportPage | 武侠 | [NovelImportPage.readable.js:61](source/frontend/NovelImportPage.readable.js#L61) |
| NovelImportPage | 奇幻 | [NovelImportPage.readable.js:61](source/frontend/NovelImportPage.readable.js#L61) |
| NovelImportPage | 灵异 | [NovelImportPage.readable.js:61](source/frontend/NovelImportPage.readable.js#L61) |
| NovelImportPage | 末世 | [NovelImportPage.readable.js:61](source/frontend/NovelImportPage.readable.js#L61) |
| NovelImportPage | 系统 | [NovelImportPage.readable.js:61](source/frontend/NovelImportPage.readable.js#L61) |
| NovelImportPage | 1万 | [NovelImportPage.readable.js:78](source/frontend/NovelImportPage.readable.js#L78) |
| NovelImportPage | 仿写 | [NovelImportPage.readable.js:113](source/frontend/NovelImportPage.readable.js#L113) |
| NovelImportPage | 同人 | [NovelImportPage.readable.js:117](source/frontend/NovelImportPage.readable.js#L117) |
| NovelImportPage | 番外 | [NovelImportPage.readable.js:121](source/frontend/NovelImportPage.readable.js#L121) |
| NovelImportPage | 续写 | [NovelImportPage.readable.js:125](source/frontend/NovelImportPage.readable.js#L125) |
| NovelImportPage | 仿写创作 | [NovelImportPage.readable.js:130](source/frontend/NovelImportPage.readable.js#L130) |
| NovelImportPage | 参考原文风格，创作全新原创故事 | [NovelImportPage.readable.js:131](source/frontend/NovelImportPage.readable.js#L131) |
| NovelImportPage | 同人创作 | [NovelImportPage.readable.js:135](source/frontend/NovelImportPage.readable.js#L135) |
| NovelImportPage | 基于原作世界观和角色进行二次创作 | [NovelImportPage.readable.js:136](source/frontend/NovelImportPage.readable.js#L136) |
| NovelImportPage | 番外创作 | [NovelImportPage.readable.js:140](source/frontend/NovelImportPage.readable.js#L140) |
| NovelImportPage | 为已有作品创作外传、前传或平行篇 | [NovelImportPage.readable.js:141](source/frontend/NovelImportPage.readable.js#L141) |
| NovelImportPage | 续写创作 | [NovelImportPage.readable.js:145](source/frontend/NovelImportPage.readable.js#L145) |
| NovelImportPage | 导入已有章节，接续创作后续内容 | [NovelImportPage.readable.js:146](source/frontend/NovelImportPage.readable.js#L146) |
| NovelImportPage | 导入素材 | [NovelImportPage.readable.js:150](source/frontend/NovelImportPage.readable.js#L150) |
| NovelImportPage | 快速创建项目 | [NovelImportPage.readable.js:151](source/frontend/NovelImportPage.readable.js#L151) |
| NovelImportPage | 导入章节 | [NovelImportPage.readable.js:162](source/frontend/NovelImportPage.readable.js#L162) |
| NovelImportPage | 创建作品 | [NovelImportPage.readable.js:162](source/frontend/NovelImportPage.readable.js#L162) |
| NovelImportPage | 第1章 | [NovelImportPage.readable.js:214](source/frontend/NovelImportPage.readable.js#L214) |
| NovelImportPage | 当前环境不支持文件选择 | [NovelImportPage.readable.js:238](source/frontend/NovelImportPage.readable.js#L238) |
| NovelImportPage | 文件 | [NovelImportPage.readable.js:247](source/frontend/NovelImportPage.readable.js#L247) |
| NovelImportPage | 读取文件失败 | [NovelImportPage.readable.js:264](source/frontend/NovelImportPage.readable.js#L264) |
| NovelImportPage | 请粘贴或上传参考原文 | [NovelImportPage.readable.js:303](source/frontend/NovelImportPage.readable.js#L303) |
| NovelImportPage | 参考原文至少需要 100 字 | [NovelImportPage.readable.js:304](source/frontend/NovelImportPage.readable.js#L304) |
| NovelImportPage | 请填写书名 | [NovelImportPage.readable.js:305](source/frontend/NovelImportPage.readable.js#L305) |
| NovelImportPage | 请粘贴原作参考文本 | [NovelImportPage.readable.js:306](source/frontend/NovelImportPage.readable.js#L306) |
| NovelImportPage | 请选择正传书籍 | [NovelImportPage.readable.js:307](source/frontend/NovelImportPage.readable.js#L307) |
| NovelImportPage | 请选择目标书籍或填写新书书名 | [NovelImportPage.readable.js:308](source/frontend/NovelImportPage.readable.js#L308) |
| NovelImportPage | 请粘贴续写章节正文 | [NovelImportPage.readable.js:309](source/frontend/NovelImportPage.readable.js#L309) |
| NovelImportPage | 未能解析出章节，请检查格式 | [NovelImportPage.readable.js:421](source/frontend/NovelImportPage.readable.js#L421) |
| NovelImportPage | 创建失败 | [NovelImportPage.readable.js:439](source/frontend/NovelImportPage.readable.js#L439) |
| NovelImportPage | 章节已导入 | [NovelImportPage.readable.js:440](source/frontend/NovelImportPage.readable.js#L440) |
| NovelImportPage | 书籍已创建 | [NovelImportPage.readable.js:440](source/frontend/NovelImportPage.readable.js#L440) |
| NovelImportPage | 关闭 | [NovelImportPage.readable.js:466](source/frontend/NovelImportPage.readable.js#L466) |
| NovelImportPage | ← 返回对话 | [NovelImportPage.readable.js:472](source/frontend/NovelImportPage.readable.js#L472) |
| NovelImportPage | 导入素材 | [NovelImportPage.readable.js:472](source/frontend/NovelImportPage.readable.js#L472) |
| NovelImportPage | 创建同人、番外、仿写或续写项目，可从文风设置多选参考风格。 | [NovelImportPage.readable.js:472](source/frontend/NovelImportPage.readable.js#L472) |
| NovelImportPage | 暂定标题（可选） | [NovelImportPage.readable.js:487](source/frontend/NovelImportPage.readable.js#L487) |
| NovelImportPage | 留空则自动生成 | [NovelImportPage.readable.js:490](source/frontend/NovelImportPage.readable.js#L490) |
| NovelImportPage | 参考原文 | [NovelImportPage.readable.js:494](source/frontend/NovelImportPage.readable.js#L494) |
| NovelImportPage |  字  | [NovelImportPage.readable.js:498](source/frontend/NovelImportPage.readable.js#L498) |
| NovelImportPage | 粘贴参考小说全文（至少 100 字），或上传 txt / docx 自动填入 | [NovelImportPage.readable.js:503](source/frontend/NovelImportPage.readable.js#L503) |
| NovelImportPage |  粘贴或上传要仿写的参考原文，系统将自动拆解情绪曲线与核心价值观  | [NovelImportPage.readable.js:507](source/frontend/NovelImportPage.readable.js#L507) |
| NovelImportPage | 上传参考文件（可选） | [NovelImportPage.readable.js:507](source/frontend/NovelImportPage.readable.js#L507) |
| NovelImportPage | 正在读取… | [NovelImportPage.readable.js:512](source/frontend/NovelImportPage.readable.js#L512) |
| NovelImportPage | + 选择文件（docx / txt） | [NovelImportPage.readable.js:512](source/frontend/NovelImportPage.readable.js#L512) |
| NovelImportPage | 已读取 ·  | [NovelImportPage.readable.js:515](source/frontend/NovelImportPage.readable.js#L515) |
| NovelImportPage |  文件内容已自动填入上方参考原文  | [NovelImportPage.readable.js:522](source/frontend/NovelImportPage.readable.js#L522) |
| NovelImportPage | 选择参考章节 | [NovelImportPage.readable.js:522](source/frontend/NovelImportPage.readable.js#L522) |
| NovelImportPage | 最近3章 | [NovelImportPage.readable.js:529](source/frontend/NovelImportPage.readable.js#L529) |
| NovelImportPage | 最近5章 | [NovelImportPage.readable.js:535](source/frontend/NovelImportPage.readable.js#L535) |
| NovelImportPage | 最近10章 | [NovelImportPage.readable.js:542](source/frontend/NovelImportPage.readable.js#L542) |
| NovelImportPage | 选择章节 | [NovelImportPage.readable.js:548](source/frontend/NovelImportPage.readable.js#L548) |
| NovelImportPage |  共识别  | [NovelImportPage.readable.js:548](source/frontend/NovelImportPage.readable.js#L548) |
| NovelImportPage |  章 · 已选择  | [NovelImportPage.readable.js:548](source/frontend/NovelImportPage.readable.js#L548) |
| NovelImportPage |  章，共  | [NovelImportPage.readable.js:548](source/frontend/NovelImportPage.readable.js#L548) |
| NovelImportPage |  字  | [NovelImportPage.readable.js:548](source/frontend/NovelImportPage.readable.js#L548) |
| NovelImportPage | 文风统计指纹 | [NovelImportPage.readable.js:548](source/frontend/NovelImportPage.readable.js#L548) |
| NovelImportPage | 平均句长 | [NovelImportPage.readable.js:550](source/frontend/NovelImportPage.readable.js#L550) |
| NovelImportPage | 字 | [NovelImportPage.readable.js:550](source/frontend/NovelImportPage.readable.js#L550) |
| NovelImportPage | 句长波动 | [NovelImportPage.readable.js:552](source/frontend/NovelImportPage.readable.js#L552) |
| NovelImportPage | 段落长度 | [NovelImportPage.readable.js:554](source/frontend/NovelImportPage.readable.js#L554) |
| NovelImportPage | 字 | [NovelImportPage.readable.js:554](source/frontend/NovelImportPage.readable.js#L554) |
| NovelImportPage | 词汇多样性 | [NovelImportPage.readable.js:556](source/frontend/NovelImportPage.readable.js#L556) |
| NovelImportPage | 高频句首： | [NovelImportPage.readable.js:558](source/frontend/NovelImportPage.readable.js#L558) |
| NovelImportPage | 修辞特征： | [NovelImportPage.readable.js:560](source/frontend/NovelImportPage.readable.js#L560) |
| NovelImportPage | 纯算法统计，不消耗 AI 额度；后端将基于 8 维度做深度定性拆解 | [NovelImportPage.readable.js:562](source/frontend/NovelImportPage.readable.js#L562) |
| NovelImportPage | 每章字数 | [NovelImportPage.readable.js:562](source/frontend/NovelImportPage.readable.js#L562) |
| NovelImportPage | 可自由填写每章字数，也可点选常用字数 | [NovelImportPage.readable.js:585](source/frontend/NovelImportPage.readable.js#L585) |
| NovelImportPage | 目标章节数 | [NovelImportPage.readable.js:585](source/frontend/NovelImportPage.readable.js#L585) |
| NovelImportPage | 长篇默认 100 章，可按需调整 | [NovelImportPage.readable.js:605](source/frontend/NovelImportPage.readable.js#L605) |
| NovelImportPage | 读者偏好 | [NovelImportPage.readable.js:605](source/frontend/NovelImportPage.readable.js#L605) |
| NovelImportPage | 填写题材偏好，如：玄幻、都市 | [NovelImportPage.readable.js:608](source/frontend/NovelImportPage.readable.js#L608) |
| NovelImportPage | 可自由填写，点选标签会自动追加 | [NovelImportPage.readable.js:623](source/frontend/NovelImportPage.readable.js#L623) |
| NovelImportPage | 原创故事想法（可选） | [NovelImportPage.readable.js:623](source/frontend/NovelImportPage.readable.js#L623) |
| NovelImportPage | 剧情、主角、核心冲突（原创剧情） | [NovelImportPage.readable.js:628](source/frontend/NovelImportPage.readable.js#L628) |
| NovelImportPage | 仿写方向（可选） | [NovelImportPage.readable.js:630](source/frontend/NovelImportPage.readable.js#L630) |
| NovelImportPage | 留空则默认采用拆书后的方向一；可填「方向二：…」 | [NovelImportPage.readable.js:635](source/frontend/NovelImportPage.readable.js#L635) |
| NovelImportPage | 补充信息（可选） | [NovelImportPage.readable.js:637](source/frontend/NovelImportPage.readable.js#L637) |
| NovelImportPage | 填写额外要求，如人设、情节走向、必须出现的桥段、禁忌点等 | [NovelImportPage.readable.js:642](source/frontend/NovelImportPage.readable.js#L642) |
| NovelImportPage | 文风设置（多选） | [NovelImportPage.readable.js:644](source/frontend/NovelImportPage.readable.js#L644) |
| NovelImportPage | 未分类 | [NovelImportPage.readable.js:653](source/frontend/NovelImportPage.readable.js#L653) |
| NovelImportPage | 暂无可用文风 | [NovelImportPage.readable.js:653](source/frontend/NovelImportPage.readable.js#L653) |
| NovelImportPage | 关联备忘录（可选） | [NovelImportPage.readable.js:655](source/frontend/NovelImportPage.readable.js#L655) |
| NovelImportPage | 选择备忘录作为参考素材 | [NovelImportPage.readable.js:659](source/frontend/NovelImportPage.readable.js#L659) |
| NovelImportPage |  已选  | [NovelImportPage.readable.js:664](source/frontend/NovelImportPage.readable.js#L664) |
| NovelImportPage |  条备忘录，内容将作为参考素材注入  | [NovelImportPage.readable.js:664](source/frontend/NovelImportPage.readable.js#L664) |
| NovelImportPage | 正传书籍 | [NovelImportPage.readable.js:666](source/frontend/NovelImportPage.readable.js#L666) |
| NovelImportPage | 选择已有正传 | [NovelImportPage.readable.js:670](source/frontend/NovelImportPage.readable.js#L670) |
| NovelImportPage | 导入到书籍（可选，不选则新建） | [NovelImportPage.readable.js:672](source/frontend/NovelImportPage.readable.js#L672) |
| NovelImportPage | 选择已有书籍或留空新建 | [NovelImportPage.readable.js:676](source/frontend/NovelImportPage.readable.js#L676) |
| NovelImportPage | 新书书名 | [NovelImportPage.readable.js:678](source/frontend/NovelImportPage.readable.js#L678) |
| NovelImportPage | 书名 | [NovelImportPage.readable.js:678](source/frontend/NovelImportPage.readable.js#L678) |
| NovelImportPage | 暂定书名 | [NovelImportPage.readable.js:681](source/frontend/NovelImportPage.readable.js#L681) |
| NovelImportPage | 番外类型 | [NovelImportPage.readable.js:682](source/frontend/NovelImportPage.readable.js#L682) |
| NovelImportPage | 后传 | [NovelImportPage.readable.js:686](source/frontend/NovelImportPage.readable.js#L686) |
| NovelImportPage | 前传 | [NovelImportPage.readable.js:689](source/frontend/NovelImportPage.readable.js#L689) |
| NovelImportPage | 平行线 | [NovelImportPage.readable.js:692](source/frontend/NovelImportPage.readable.js#L692) |
| NovelImportPage | If 线 | [NovelImportPage.readable.js:695](source/frontend/NovelImportPage.readable.js#L695) |
| NovelImportPage | 番外方向 | [NovelImportPage.readable.js:698](source/frontend/NovelImportPage.readable.js#L698) |
| NovelImportPage | 番外要写什么 | [NovelImportPage.readable.js:703](source/frontend/NovelImportPage.readable.js#L703) |
| NovelImportPage | 同人方向 | [NovelImportPage.readable.js:704](source/frontend/NovelImportPage.readable.js#L704) |
| NovelImportPage | CP、AU、IF 线等 | [NovelImportPage.readable.js:709](source/frontend/NovelImportPage.readable.js#L709) |
| NovelImportPage | 同人模式 | [NovelImportPage.readable.js:710](source/frontend/NovelImportPage.readable.js#L710) |
| NovelImportPage | 尊重原作 (canon) | [NovelImportPage.readable.js:714](source/frontend/NovelImportPage.readable.js#L714) |
| NovelImportPage | AU / 架空 | [NovelImportPage.readable.js:717](source/frontend/NovelImportPage.readable.js#L717) |
| NovelImportPage | OOC / 性格偏离 | [NovelImportPage.readable.js:720](source/frontend/NovelImportPage.readable.js#L720) |
| NovelImportPage | CP / 关系向 | [NovelImportPage.readable.js:723](source/frontend/NovelImportPage.readable.js#L723) |
| NovelImportPage | 章节正文 | [NovelImportPage.readable.js:726](source/frontend/NovelImportPage.readable.js#L726) |
| NovelImportPage | 参考文本（可选） | [NovelImportPage.readable.js:726](source/frontend/NovelImportPage.readable.js#L726) |
| NovelImportPage | 粘贴章节，建议以「第1章」等标题分段 | [NovelImportPage.readable.js:731](source/frontend/NovelImportPage.readable.js#L731) |
| NovelImportPage | 粘贴参考原文；也可只选下方文风 | [NovelImportPage.readable.js:731](source/frontend/NovelImportPage.readable.js#L731) |
| NovelImportPage | 文风设置（多选） | [NovelImportPage.readable.js:732](source/frontend/NovelImportPage.readable.js#L732) |
| NovelImportPage | 未分类 | [NovelImportPage.readable.js:741](source/frontend/NovelImportPage.readable.js#L741) |
| NovelImportPage | 暂无可用文风 | [NovelImportPage.readable.js:741](source/frontend/NovelImportPage.readable.js#L741) |
| NovelImportPage | 取消 | [NovelImportPage.readable.js:754](source/frontend/NovelImportPage.readable.js#L754) |
| NovelImportPage | 选择关联章节 | [NovelImportPage.readable.js:779](source/frontend/NovelImportPage.readable.js#L779) |
| NovelImportPage |  全选  | [NovelImportPage.readable.js:787](source/frontend/NovelImportPage.readable.js#L787) |
| NovelImportPage |  章  | [NovelImportPage.readable.js:787](source/frontend/NovelImportPage.readable.js#L787) |
| NovelImportPage | 清空 | [NovelImportPage.readable.js:791](source/frontend/NovelImportPage.readable.js#L791) |
| NovelImportPage | 字 | [NovelImportPage.readable.js:807](source/frontend/NovelImportPage.readable.js#L807) |
| NovelImportPage |  已选择  | [NovelImportPage.readable.js:811](source/frontend/NovelImportPage.readable.js#L811) |
| NovelImportPage |  章  | [NovelImportPage.readable.js:811](source/frontend/NovelImportPage.readable.js#L811) |
| NovelImportPage |  共  | [NovelImportPage.readable.js:813](source/frontend/NovelImportPage.readable.js#L813) |
| NovelImportPage |  字  | [NovelImportPage.readable.js:813](source/frontend/NovelImportPage.readable.js#L813) |
| NovelImportPage | 取消 | [NovelImportPage.readable.js:817](source/frontend/NovelImportPage.readable.js#L817) |
| NovelImportPage | 确认选择 | [NovelImportPage.readable.js:823](source/frontend/NovelImportPage.readable.js#L823) |
| NovelImitationShortDialog | 轻松搞笑 | [NovelImitationShortDialog.readable.js:56](source/frontend/NovelImitationShortDialog.readable.js#L56) |
| NovelImitationShortDialog | 热血爽文 | [NovelImitationShortDialog.readable.js:56](source/frontend/NovelImitationShortDialog.readable.js#L56) |
| NovelImitationShortDialog | 智斗权谋 | [NovelImitationShortDialog.readable.js:56](source/frontend/NovelImitationShortDialog.readable.js#L56) |
| NovelImitationShortDialog | 悬疑惊悚 | [NovelImitationShortDialog.readable.js:56](source/frontend/NovelImitationShortDialog.readable.js#L56) |
| NovelImitationShortDialog | 甜宠治愈 | [NovelImitationShortDialog.readable.js:56](source/frontend/NovelImitationShortDialog.readable.js#L56) |
| NovelImitationShortDialog | 1.0万 | [NovelImitationShortDialog.readable.js:58](source/frontend/NovelImitationShortDialog.readable.js#L58) |
| NovelImitationShortDialog | 1.1万 | [NovelImitationShortDialog.readable.js:61](source/frontend/NovelImitationShortDialog.readable.js#L61) |
| NovelImitationShortDialog | 1.2万 | [NovelImitationShortDialog.readable.js:64](source/frontend/NovelImitationShortDialog.readable.js#L64) |
| NovelImitationShortDialog | 1.3万 | [NovelImitationShortDialog.readable.js:67](source/frontend/NovelImitationShortDialog.readable.js#L67) |
| NovelImitationShortDialog | 1.5万 | [NovelImitationShortDialog.readable.js:70](source/frontend/NovelImitationShortDialog.readable.js#L70) |
| NovelImitationShortDialog | 2.0万 | [NovelImitationShortDialog.readable.js:73](source/frontend/NovelImitationShortDialog.readable.js#L73) |
| NovelImitationShortDialog | 女频 | [NovelImitationShortDialog.readable.js:76](source/frontend/NovelImitationShortDialog.readable.js#L76) |
| NovelImitationShortDialog | 男频 | [NovelImitationShortDialog.readable.js:76](source/frontend/NovelImitationShortDialog.readable.js#L76) |
| NovelImitationShortDialog | 丧尸 | [NovelImitationShortDialog.readable.js:76](source/frontend/NovelImitationShortDialog.readable.js#L76) |
| NovelImitationShortDialog | 萌娃 | [NovelImitationShortDialog.readable.js:76](source/frontend/NovelImitationShortDialog.readable.js#L76) |
| NovelImitationShortDialog | 萌宠穿越 | [NovelImitationShortDialog.readable.js:76](source/frontend/NovelImitationShortDialog.readable.js#L76) |
| NovelImitationShortDialog | 通古今 | [NovelImitationShortDialog.readable.js:76](source/frontend/NovelImitationShortDialog.readable.js#L76) |
| NovelImitationShortDialog | 直播 | [NovelImitationShortDialog.readable.js:76](source/frontend/NovelImitationShortDialog.readable.js#L76) |
| NovelImitationShortDialog | 历史 | [NovelImitationShortDialog.readable.js:76](source/frontend/NovelImitationShortDialog.readable.js#L76) |
| NovelImitationShortDialog | 系统 | [NovelImitationShortDialog.readable.js:76](source/frontend/NovelImitationShortDialog.readable.js#L76) |
| NovelImitationShortDialog | 修仙 | [NovelImitationShortDialog.readable.js:76](source/frontend/NovelImitationShortDialog.readable.js#L76) |
| NovelImitationShortDialog | 搞笑 | [NovelImitationShortDialog.readable.js:76](source/frontend/NovelImitationShortDialog.readable.js#L76) |
| NovelImitationShortDialog | 重生 | [NovelImitationShortDialog.readable.js:76](source/frontend/NovelImitationShortDialog.readable.js#L76) |
| NovelImitationShortDialog | 穿越 | [NovelImitationShortDialog.readable.js:76](source/frontend/NovelImitationShortDialog.readable.js#L76) |
| NovelImitationShortDialog | 都市 | [NovelImitationShortDialog.readable.js:76](source/frontend/NovelImitationShortDialog.readable.js#L76) |
| NovelImitationShortDialog | 校园 | [NovelImitationShortDialog.readable.js:76](source/frontend/NovelImitationShortDialog.readable.js#L76) |
| NovelImitationShortDialog | 豪门 | [NovelImitationShortDialog.readable.js:76](source/frontend/NovelImitationShortDialog.readable.js#L76) |
| NovelImitationShortDialog | 职场 | [NovelImitationShortDialog.readable.js:76](source/frontend/NovelImitationShortDialog.readable.js#L76) |
| NovelImitationShortDialog | 追妻火葬场 | [NovelImitationShortDialog.readable.js:76](source/frontend/NovelImitationShortDialog.readable.js#L76) |
| NovelImitationShortDialog | 大女主 | [NovelImitationShortDialog.readable.js:76](source/frontend/NovelImitationShortDialog.readable.js#L76) |
| NovelImitationShortDialog | 虐恋 | [NovelImitationShortDialog.readable.js:76](source/frontend/NovelImitationShortDialog.readable.js#L76) |
| NovelImitationShortDialog | 当前环境不支持文件选择 | [NovelImitationShortDialog.readable.js:117](source/frontend/NovelImitationShortDialog.readable.js#L117) |
| NovelImitationShortDialog | 文件 | [NovelImitationShortDialog.readable.js:125](source/frontend/NovelImitationShortDialog.readable.js#L125) |
| NovelImitationShortDialog | 读取文件失败 | [NovelImitationShortDialog.readable.js:138](source/frontend/NovelImitationShortDialog.readable.js#L138) |
| NovelImitationShortDialog | 参考原文至少需要 100 字 | [NovelImitationShortDialog.readable.js:144](source/frontend/NovelImitationShortDialog.readable.js#L144) |
| NovelImitationShortDialog | 仿写短篇 | [NovelImitationShortDialog.readable.js:181](source/frontend/NovelImitationShortDialog.readable.js#L181) |
| NovelImitationShortDialog | 参考原文风格，一键生成知乎盐选风短篇小说 | [NovelImitationShortDialog.readable.js:181](source/frontend/NovelImitationShortDialog.readable.js#L181) |
| NovelImitationShortDialog | 暂定标题（可选） | [NovelImitationShortDialog.readable.js:192](source/frontend/NovelImitationShortDialog.readable.js#L192) |
| NovelImitationShortDialog | 留空则自动生成 | [NovelImitationShortDialog.readable.js:195](source/frontend/NovelImitationShortDialog.readable.js#L195) |
| NovelImitationShortDialog | 参考原文 | [NovelImitationShortDialog.readable.js:198](source/frontend/NovelImitationShortDialog.readable.js#L198) |
| NovelImitationShortDialog |  字  | [NovelImitationShortDialog.readable.js:202](source/frontend/NovelImitationShortDialog.readable.js#L202) |
| NovelImitationShortDialog | 粘贴参考小说全文（至少 100 字），或上传 txt / docx 自动填入 | [NovelImitationShortDialog.readable.js:207](source/frontend/NovelImitationShortDialog.readable.js#L207) |
| NovelImitationShortDialog |  粘贴或上传要仿写的参考原文，系统将自动拆解情绪曲线与核心价值观  | [NovelImitationShortDialog.readable.js:211](source/frontend/NovelImitationShortDialog.readable.js#L211) |
| NovelImitationShortDialog | 上传参考文件（可选） | [NovelImitationShortDialog.readable.js:211](source/frontend/NovelImitationShortDialog.readable.js#L211) |
| NovelImitationShortDialog | 正在读取… | [NovelImitationShortDialog.readable.js:216](source/frontend/NovelImitationShortDialog.readable.js#L216) |
| NovelImitationShortDialog | + 选择文件（docx / txt） | [NovelImitationShortDialog.readable.js:216](source/frontend/NovelImitationShortDialog.readable.js#L216) |
| NovelImitationShortDialog | 已读取 ·  | [NovelImitationShortDialog.readable.js:219](source/frontend/NovelImitationShortDialog.readable.js#L219) |
| NovelImitationShortDialog |  文件内容已自动填入上方参考原文，可继续编辑  | [NovelImitationShortDialog.readable.js:226](source/frontend/NovelImitationShortDialog.readable.js#L226) |
| NovelImitationShortDialog | 目标字数 | [NovelImitationShortDialog.readable.js:226](source/frontend/NovelImitationShortDialog.readable.js#L226) |
| NovelImitationShortDialog | 如 13000 | [NovelImitationShortDialog.readable.js:233](source/frontend/NovelImitationShortDialog.readable.js#L233) |
| NovelImitationShortDialog | 可自由填写字数（字），也可点选常用字数 | [NovelImitationShortDialog.readable.js:249](source/frontend/NovelImitationShortDialog.readable.js#L249) |
| NovelImitationShortDialog | 读者偏好 | [NovelImitationShortDialog.readable.js:249](source/frontend/NovelImitationShortDialog.readable.js#L249) |
| NovelImitationShortDialog | 填写读者偏好，如：轻松搞笑、甜宠治愈 | [NovelImitationShortDialog.readable.js:252](source/frontend/NovelImitationShortDialog.readable.js#L252) |
| NovelImitationShortDialog | 可自由填写，点选标签会自动追加 | [NovelImitationShortDialog.readable.js:267](source/frontend/NovelImitationShortDialog.readable.js#L267) |
| NovelImitationShortDialog | 目标章节 | [NovelImitationShortDialog.readable.js:267](source/frontend/NovelImitationShortDialog.readable.js#L267) |
| NovelImitationShortDialog | 默认 1 章，可按需调整 | [NovelImitationShortDialog.readable.js:279](source/frontend/NovelImitationShortDialog.readable.js#L279) |
| NovelImitationShortDialog | 补充信息（可选） | [NovelImitationShortDialog.readable.js:279](source/frontend/NovelImitationShortDialog.readable.js#L279) |
| NovelImitationShortDialog | 填写额外要求，如人设、情节走向、必须出现的桥段、禁忌点等 | [NovelImitationShortDialog.readable.js:284](source/frontend/NovelImitationShortDialog.readable.js#L284) |
| NovelImitationShortDialog | 灵感标签（可选） | [NovelImitationShortDialog.readable.js:286](source/frontend/NovelImitationShortDialog.readable.js#L286) |
| NovelImitationShortDialog | 文风设置（多选） | [NovelImitationShortDialog.readable.js:297](source/frontend/NovelImitationShortDialog.readable.js#L297) |
| NovelImitationShortDialog | 未分类 | [NovelImitationShortDialog.readable.js:306](source/frontend/NovelImitationShortDialog.readable.js#L306) |
| NovelImitationShortDialog | 暂无可用文风 | [NovelImitationShortDialog.readable.js:306](source/frontend/NovelImitationShortDialog.readable.js#L306) |
| NovelImitationShortDialog | 关联备忘录（可选） | [NovelImitationShortDialog.readable.js:308](source/frontend/NovelImitationShortDialog.readable.js#L308) |
| NovelImitationShortDialog | 选择备忘录作为参考素材 | [NovelImitationShortDialog.readable.js:312](source/frontend/NovelImitationShortDialog.readable.js#L312) |
| NovelImitationShortDialog |  已选  | [NovelImitationShortDialog.readable.js:317](source/frontend/NovelImitationShortDialog.readable.js#L317) |
| NovelImitationShortDialog |  条备忘录，内容将作为参考素材注入  | [NovelImitationShortDialog.readable.js:317](source/frontend/NovelImitationShortDialog.readable.js#L317) |
| NovelImitationShortDialog | 关联存档（可选） | [NovelImitationShortDialog.readable.js:317](source/frontend/NovelImitationShortDialog.readable.js#L317) |
| NovelImitationShortDialog | 关联已有作品 | [NovelImitationShortDialog.readable.js:321](source/frontend/NovelImitationShortDialog.readable.js#L321) |
| NovelImitationShortDialog |  个标签  | [NovelImitationShortDialog.readable.js:326](source/frontend/NovelImitationShortDialog.readable.js#L326) |
| NovelImitationShortDialog | 取消 | [NovelImitationShortDialog.readable.js:331](source/frontend/NovelImitationShortDialog.readable.js#L331) |
| NovelImitationShortDialog |  开始仿写  | [NovelImitationShortDialog.readable.js:338](source/frontend/NovelImitationShortDialog.readable.js#L338) |
| NovelBatchDialog | 已取消 | [NovelBatchDialog.readable.js:82](source/frontend/NovelBatchDialog.readable.js#L82) |
| NovelBatchDialog | 已达目标章数 | [NovelBatchDialog.readable.js:83](source/frontend/NovelBatchDialog.readable.js#L83) |
| NovelBatchDialog | 连续失败过多 | [NovelBatchDialog.readable.js:84](source/frontend/NovelBatchDialog.readable.js#L84) |
| NovelBatchDialog | 批量写作失败 | [NovelBatchDialog.readable.js:88](source/frontend/NovelBatchDialog.readable.js#L88) |
| NovelBatchDialog | 批量写作异常： | [NovelBatchDialog.readable.js:90](source/frontend/NovelBatchDialog.readable.js#L90) |
| NovelBatchDialog | 批量写作 | [NovelBatchDialog.readable.js:104](source/frontend/NovelBatchDialog.readable.js#L104) |
| NovelBatchDialog |  — 正在写第  | [NovelBatchDialog.readable.js:127](source/frontend/NovelBatchDialog.readable.js#L127) |
| NovelBatchDialog |  章  | [NovelBatchDialog.readable.js:127](source/frontend/NovelBatchDialog.readable.js#L127) |
| NovelBatchDialog |  取消  | [NovelBatchDialog.readable.js:132](source/frontend/NovelBatchDialog.readable.js#L132) |
| NovelBatchDialog | 连续写作章数 | [NovelBatchDialog.readable.js:136](source/frontend/NovelBatchDialog.readable.js#L136) |
| NovelBatchDialog | 连续失败停止阈值 | [NovelBatchDialog.readable.js:142](source/frontend/NovelBatchDialog.readable.js#L142) |
| NovelBatchDialog |  开始批量写作  | [NovelBatchDialog.readable.js:153](source/frontend/NovelBatchDialog.readable.js#L153) |
| NovelSchedulerPanel | 运行中 | [NovelSchedulerPanel.readable.js:28](source/frontend/NovelSchedulerPanel.readable.js#L28) |
| NovelSchedulerPanel | 已暂停 | [NovelSchedulerPanel.readable.js:29](source/frontend/NovelSchedulerPanel.readable.js#L29) |
| NovelSchedulerPanel | 已停止 | [NovelSchedulerPanel.readable.js:30](source/frontend/NovelSchedulerPanel.readable.js#L30) |
| NovelSchedulerPanel | 未知 | [NovelSchedulerPanel.readable.js:31](source/frontend/NovelSchedulerPanel.readable.js#L31) |
| NovelSchedulerPanel | 调度器已启动 | [NovelSchedulerPanel.readable.js:104](source/frontend/NovelSchedulerPanel.readable.js#L104) |
| NovelSchedulerPanel | 启动失败 | [NovelSchedulerPanel.readable.js:104](source/frontend/NovelSchedulerPanel.readable.js#L104) |
| NovelSchedulerPanel | 调度器已停止 | [NovelSchedulerPanel.readable.js:129](source/frontend/NovelSchedulerPanel.readable.js#L129) |
| NovelSchedulerPanel | 已恢复 | [NovelSchedulerPanel.readable.js:161](source/frontend/NovelSchedulerPanel.readable.js#L161) |
| NovelSchedulerPanel | 已暂停 | [NovelSchedulerPanel.readable.js:161](source/frontend/NovelSchedulerPanel.readable.js#L161) |
| NovelSchedulerPanel | ← 返回 | [NovelSchedulerPanel.readable.js:185](source/frontend/NovelSchedulerPanel.readable.js#L185) |
| NovelSchedulerPanel | 自动写作调度器 | [NovelSchedulerPanel.readable.js:185](source/frontend/NovelSchedulerPanel.readable.js#L185) |
| NovelSchedulerPanel | 状态： | [NovelSchedulerPanel.readable.js:185](source/frontend/NovelSchedulerPanel.readable.js#L185) |
| NovelSchedulerPanel |  今日  | [NovelSchedulerPanel.readable.js:191](source/frontend/NovelSchedulerPanel.readable.js#L191) |
| NovelSchedulerPanel |  章 \| 累计  | [NovelSchedulerPanel.readable.js:191](source/frontend/NovelSchedulerPanel.readable.js#L191) |
| NovelSchedulerPanel |  章  | [NovelSchedulerPanel.readable.js:191](source/frontend/NovelSchedulerPanel.readable.js#L191) |
| NovelSchedulerPanel |  连续失败  | [NovelSchedulerPanel.readable.js:191](source/frontend/NovelSchedulerPanel.readable.js#L191) |
| NovelSchedulerPanel |  次  | [NovelSchedulerPanel.readable.js:191](source/frontend/NovelSchedulerPanel.readable.js#L191) |
| NovelSchedulerPanel | 恢复 | [NovelSchedulerPanel.readable.js:199](source/frontend/NovelSchedulerPanel.readable.js#L199) |
| NovelSchedulerPanel | 暂停 | [NovelSchedulerPanel.readable.js:199](source/frontend/NovelSchedulerPanel.readable.js#L199) |
| NovelSchedulerPanel | 停止 | [NovelSchedulerPanel.readable.js:207](source/frontend/NovelSchedulerPanel.readable.js#L207) |
| NovelSchedulerPanel | 写作间隔（分钟） | [NovelSchedulerPanel.readable.js:209](source/frontend/NovelSchedulerPanel.readable.js#L209) |
| NovelSchedulerPanel | 每日上限 | [NovelSchedulerPanel.readable.js:215](source/frontend/NovelSchedulerPanel.readable.js#L215) |
| NovelSchedulerPanel | 连续失败暂停阈值 | [NovelSchedulerPanel.readable.js:221](source/frontend/NovelSchedulerPanel.readable.js#L221) |
| NovelSchedulerPanel | 启动调度器 | [NovelSchedulerPanel.readable.js:233](source/frontend/NovelSchedulerPanel.readable.js#L233) |
| NovelSchedulerPanel |  下次写作：约  | [NovelSchedulerPanel.readable.js:235](source/frontend/NovelSchedulerPanel.readable.js#L235) |
| NovelSchedulerPanel |  分钟后  | [NovelSchedulerPanel.readable.js:235](source/frontend/NovelSchedulerPanel.readable.js#L235) |
| NovelHooksPanel | 活跃 | [NovelHooksPanel.readable.js:39](source/frontend/NovelHooksPanel.readable.js#L39) |
| NovelHooksPanel | 已解决 | [NovelHooksPanel.readable.js:40](source/frontend/NovelHooksPanel.readable.js#L40) |
| NovelHooksPanel | 全部 | [NovelHooksPanel.readable.js:41](source/frontend/NovelHooksPanel.readable.js#L41) |
| NovelHooksPanel | 已植入 | [NovelHooksPanel.readable.js:51](source/frontend/NovelHooksPanel.readable.js#L51) |
| NovelHooksPanel | 推进中 | [NovelHooksPanel.readable.js:52](source/frontend/NovelHooksPanel.readable.js#L52) |
| NovelHooksPanel | 已解决 | [NovelHooksPanel.readable.js:53](source/frontend/NovelHooksPanel.readable.js#L53) |
| NovelHooksPanel | 起始 ch | [NovelHooksPanel.readable.js:58](source/frontend/NovelHooksPanel.readable.js#L58) |
| NovelHooksPanel | 最近推进 ch | [NovelHooksPanel.readable.js:58](source/frontend/NovelHooksPanel.readable.js#L58) |
| NovelHooksPanel | 推进  | [NovelHooksPanel.readable.js:58](source/frontend/NovelHooksPanel.readable.js#L58) |
| NovelHooksPanel |  次 | [NovelHooksPanel.readable.js:58](source/frontend/NovelHooksPanel.readable.js#L58) |
| NovelHooksPanel | 暂无伏笔记录 | [NovelHooksPanel.readable.js:60](source/frontend/NovelHooksPanel.readable.js#L60) |
| NovelDashboard | 草稿 | [NovelDashboard.readable.js:28](source/frontend/NovelDashboard.readable.js#L28) |
| NovelDashboard | 通过 | [NovelDashboard.readable.js:29](source/frontend/NovelDashboard.readable.js#L29) |
| NovelDashboard | 待审 | [NovelDashboard.readable.js:30](source/frontend/NovelDashboard.readable.js#L30) |
| NovelDashboard | 未过 | [NovelDashboard.readable.js:31](source/frontend/NovelDashboard.readable.js#L31) |
| NovelDashboard | 降级 | [NovelDashboard.readable.js:32](source/frontend/NovelDashboard.readable.js#L32) |
| NovelDashboard | 总章数 | [NovelDashboard.readable.js:42](source/frontend/NovelDashboard.readable.js#L42) |
| NovelDashboard | 万 | [NovelDashboard.readable.js:42](source/frontend/NovelDashboard.readable.js#L42) |
| NovelDashboard | 总字数 | [NovelDashboard.readable.js:44](source/frontend/NovelDashboard.readable.js#L44) |
| NovelDashboard | 平均字数 | [NovelDashboard.readable.js:46](source/frontend/NovelDashboard.readable.js#L46) |
| NovelDashboard | 章节状态分布 | [NovelDashboard.readable.js:46](source/frontend/NovelDashboard.readable.js#L46) |
| NovelDashboard | 章节字数趋势 | [NovelDashboard.readable.js:54](source/frontend/NovelDashboard.readable.js#L54) |
| NovelDashboard | 暂无章节数据 | [NovelDashboard.readable.js:63](source/frontend/NovelDashboard.readable.js#L63) |
| NovelCheckpointCard | 确认创作蓝图 | [NovelCheckpointCard.readable.js:22](source/frontend/NovelCheckpointCard.readable.js#L22) |
| NovelCheckpointCard | 确认基础设定 | [NovelCheckpointCard.readable.js:22](source/frontend/NovelCheckpointCard.readable.js#L22) |
| NovelCheckpointCard | 确认继续生成 | [NovelCheckpointCard.readable.js:22](source/frontend/NovelCheckpointCard.readable.js#L22) |
| NovelCheckpointCard | 内容已生成，可在右侧修改本地文件，确认后继续下一步。 | [NovelCheckpointCard.readable.js:26](source/frontend/NovelCheckpointCard.readable.js#L26) |
| NovelCheckpointCard | 继续中… | [NovelCheckpointCard.readable.js:28](source/frontend/NovelCheckpointCard.readable.js#L28) |
| NovelCheckpointCard | 确认继续 | [NovelCheckpointCard.readable.js:28](source/frontend/NovelCheckpointCard.readable.js#L28) |
| NovelCheckpointCard |  稍后  | [NovelCheckpointCard.readable.js:39](source/frontend/NovelCheckpointCard.readable.js#L39) |
| NovelStudioPage | 拆书分析 | [NovelStudioPage.readable.js:133](source/frontend/NovelStudioPage.readable.js#L133) |
| NovelStudioPage | 生成仿写方向 | [NovelStudioPage.readable.js:133](source/frontend/NovelStudioPage.readable.js#L133) |
| NovelStudioPage | 构建创作蓝图 | [NovelStudioPage.readable.js:133](source/frontend/NovelStudioPage.readable.js#L133) |
| NovelStudioPage | 生成短篇大纲 | [NovelStudioPage.readable.js:133](source/frontend/NovelStudioPage.readable.js#L133) |
| NovelStudioPage | 审阅大纲 | [NovelStudioPage.readable.js:133](source/frontend/NovelStudioPage.readable.js#L133) |
| NovelStudioPage | 修订大纲 | [NovelStudioPage.readable.js:133](source/frontend/NovelStudioPage.readable.js#L133) |
| NovelStudioPage | 撰写正文 | [NovelStudioPage.readable.js:133](source/frontend/NovelStudioPage.readable.js#L133) |
| NovelStudioPage | 审阅正文 | [NovelStudioPage.readable.js:133](source/frontend/NovelStudioPage.readable.js#L133) |
| NovelStudioPage | 修订正文 | [NovelStudioPage.readable.js:133](source/frontend/NovelStudioPage.readable.js#L133) |
| NovelStudioPage | 编辑审核 | [NovelStudioPage.readable.js:133](source/frontend/NovelStudioPage.readable.js#L133) |
| NovelStudioPage | 打包输出 | [NovelStudioPage.readable.js:133](source/frontend/NovelStudioPage.readable.js#L133) |
| NovelStudioPage | 书籍资料 | [NovelStudioPage.readable.js:180](source/frontend/NovelStudioPage.readable.js#L180) |
| NovelStudioPage | 短篇资料 | [NovelStudioPage.readable.js:180](source/frontend/NovelStudioPage.readable.js#L180) |
| NovelStudioPage | 任务进度 | [NovelStudioPage.readable.js:180](source/frontend/NovelStudioPage.readable.js#L180) |
| NovelStudioPage | 已清除写作风格参考 | [NovelStudioPage.readable.js:222](source/frontend/NovelStudioPage.readable.js#L222) |
| NovelStudioPage | 可用 | [NovelStudioPage.readable.js:249](source/frontend/NovelStudioPage.readable.js#L249) |
| NovelStudioPage | 加载会话失败 | [NovelStudioPage.readable.js:304](source/frontend/NovelStudioPage.readable.js#L304) |
| NovelStudioPage | 新会话 | [NovelStudioPage.readable.js:400](source/frontend/NovelStudioPage.readable.js#L400) |
| NovelStudioPage | 创建会话失败 | [NovelStudioPage.readable.js:402](source/frontend/NovelStudioPage.readable.js#L402) |
| NovelStudioPage | 书籍会话 | [NovelStudioPage.readable.js:434](source/frontend/NovelStudioPage.readable.js#L434) |
| NovelStudioPage | 发送失败 | [NovelStudioPage.readable.js:488](source/frontend/NovelStudioPage.readable.js#L488) |
| NovelStudioPage | 发送失败 | [NovelStudioPage.readable.js:491](source/frontend/NovelStudioPage.readable.js#L491) |
| NovelStudioPage | 创建失败 | [NovelStudioPage.readable.js:565](source/frontend/NovelStudioPage.readable.js#L565) |
| NovelStudioPage | 创建失败 | [NovelStudioPage.readable.js:567](source/frontend/NovelStudioPage.readable.js#L567) |
| NovelStudioPage | 生成短篇大纲 | [NovelStudioPage.readable.js:597](source/frontend/NovelStudioPage.readable.js#L597) |
| NovelStudioPage | 审阅大纲 | [NovelStudioPage.readable.js:597](source/frontend/NovelStudioPage.readable.js#L597) |
| NovelStudioPage | 修订大纲 | [NovelStudioPage.readable.js:597](source/frontend/NovelStudioPage.readable.js#L597) |
| NovelStudioPage | 撰写正文 | [NovelStudioPage.readable.js:597](source/frontend/NovelStudioPage.readable.js#L597) |
| NovelStudioPage | 审阅正文 | [NovelStudioPage.readable.js:597](source/frontend/NovelStudioPage.readable.js#L597) |
| NovelStudioPage | 修订正文 | [NovelStudioPage.readable.js:597](source/frontend/NovelStudioPage.readable.js#L597) |
| NovelStudioPage | 打包输出 | [NovelStudioPage.readable.js:597](source/frontend/NovelStudioPage.readable.js#L597) |
| NovelStudioPage | 继续失败 | [NovelStudioPage.readable.js:655](source/frontend/NovelStudioPage.readable.js#L655) |
| NovelStudioPage | 已确认基础设定 | [NovelStudioPage.readable.js:662](source/frontend/NovelStudioPage.readable.js#L662) |
| NovelStudioPage | 仿写短篇已完成 | [NovelStudioPage.readable.js:662](source/frontend/NovelStudioPage.readable.js#L662) |
| NovelStudioPage | 已继续生成 | [NovelStudioPage.readable.js:662](source/frontend/NovelStudioPage.readable.js#L662) |
| NovelStudioPage | 执行失败 | [NovelStudioPage.readable.js:704](source/frontend/NovelStudioPage.readable.js#L704) |
| NovelStudioPage | 执行失败 | [NovelStudioPage.readable.js:705](source/frontend/NovelStudioPage.readable.js#L705) |
| NovelStudioPage | 已打开导入素材，请补充材料 | [NovelStudioPage.readable.js:708](source/frontend/NovelStudioPage.readable.js#L708) |
| NovelStudioPage | 书籍已创建，已切换到书籍会话 | [NovelStudioPage.readable.js:710](source/frontend/NovelStudioPage.readable.js#L710) |
| NovelStudioPage | 内容已生成，可在右侧修改后确认继续 | [NovelStudioPage.readable.js:717](source/frontend/NovelStudioPage.readable.js#L717) |
| NovelStudioPage | 内容已生成，请查看右侧文件并确认后继续 | [NovelStudioPage.readable.js:718](source/frontend/NovelStudioPage.readable.js#L718) |
| NovelStudioPage | 审稿通过 | [NovelStudioPage.readable.js:722](source/frontend/NovelStudioPage.readable.js#L722) |
| NovelStudioPage | 审稿未通过，请查看卡片中的问题列表 | [NovelStudioPage.readable.js:722](source/frontend/NovelStudioPage.readable.js#L722) |
| NovelStudioPage | 执行完成 | [NovelStudioPage.readable.js:727](source/frontend/NovelStudioPage.readable.js#L727) |
| NovelStudioPage | 书籍已创建，会话已归档到该书下 | [NovelStudioPage.readable.js:776](source/frontend/NovelStudioPage.readable.js#L776) |
| NovelStudioPage | 仿写短篇 | [NovelStudioPage.readable.js:857](source/frontend/NovelStudioPage.readable.js#L857) |
| NovelStudioPage | 仿写短篇 | [NovelStudioPage.readable.js:864](source/frontend/NovelStudioPage.readable.js#L864) |
| NovelStudioPage | 仿写短篇 | [NovelStudioPage.readable.js:877](source/frontend/NovelStudioPage.readable.js#L877) |
| NovelStudioPage | 无法创建会话，请重试 | [NovelStudioPage.readable.js:877](source/frontend/NovelStudioPage.readable.js#L877) |
| NovelStudioPage | 仿写短篇 | [NovelStudioPage.readable.js:878](source/frontend/NovelStudioPage.readable.js#L878) |
| NovelStudioPage | 仿写短篇已开始，可在右侧查看进度 | [NovelStudioPage.readable.js:879](source/frontend/NovelStudioPage.readable.js#L879) |
| NovelStudioPage | 仿写短篇失败 | [NovelStudioPage.readable.js:913](source/frontend/NovelStudioPage.readable.js#L913) |
| NovelStudioPage | 仿写短篇失败 | [NovelStudioPage.readable.js:914](source/frontend/NovelStudioPage.readable.js#L914) |
| NovelStudioPage | 仿写短篇失败 | [NovelStudioPage.readable.js:915](source/frontend/NovelStudioPage.readable.js#L915) |
| NovelStudioPage | 创作蓝图已生成，可在右侧修改后确认继续 | [NovelStudioPage.readable.js:920](source/frontend/NovelStudioPage.readable.js#L920) |
| NovelStudioPage | 创作蓝图已生成，请查看右侧文件并确认后继续 | [NovelStudioPage.readable.js:921](source/frontend/NovelStudioPage.readable.js#L921) |
| NovelStudioPage | 仿写短篇失败 | [NovelStudioPage.readable.js:928](source/frontend/NovelStudioPage.readable.js#L928) |
| NovelStudioPage | 仿写短篇失败 | [NovelStudioPage.readable.js:929](source/frontend/NovelStudioPage.readable.js#L929) |
| NovelStudioPage | 仿写短篇失败 | [NovelStudioPage.readable.js:930](source/frontend/NovelStudioPage.readable.js#L930) |
| NovelStudioPage | 任务完成 | [NovelStudioPage.readable.js:1041](source/frontend/NovelStudioPage.readable.js#L1041) |
| NovelStudioPage | 仿写短篇已完成 | [NovelStudioPage.readable.js:1054](source/frontend/NovelStudioPage.readable.js#L1054) |
| NovelStudioPage | 封面图已生成 | [NovelStudioPage.readable.js:1054](source/frontend/NovelStudioPage.readable.js#L1054) |
| NovelStudioPage | 封面提示词已保存 | [NovelStudioPage.readable.js:1054](source/frontend/NovelStudioPage.readable.js#L1054) |
| NovelStudioPage | 写章失败 | [NovelStudioPage.readable.js:1137](source/frontend/NovelStudioPage.readable.js#L1137) |
| NovelStudioPage | 审稿失败 | [NovelStudioPage.readable.js:1210](source/frontend/NovelStudioPage.readable.js#L1210) |
| NovelStudioPage | 审稿通过 | [NovelStudioPage.readable.js:1226](source/frontend/NovelStudioPage.readable.js#L1226) |
| NovelStudioPage | 审稿未通过，请查看卡片中的问题列表 | [NovelStudioPage.readable.js:1226](source/frontend/NovelStudioPage.readable.js#L1226) |
| NovelStudioPage | 删除失败 | [NovelStudioPage.readable.js:1242](source/frontend/NovelStudioPage.readable.js#L1242) |
| NovelStudioPage | 小说会话 | [NovelStudioPage.readable.js:1250](source/frontend/NovelStudioPage.readable.js#L1250) |
| NovelStudioPage | 已存档到备忘录 | [NovelStudioPage.readable.js:1253](source/frontend/NovelStudioPage.readable.js#L1253) |
| NovelStudioPage | 修订失败 | [NovelStudioPage.readable.js:1297](source/frontend/NovelStudioPage.readable.js#L1297) |
| NovelStudioPage | 修订完成 | [NovelStudioPage.readable.js:1304](source/frontend/NovelStudioPage.readable.js#L1304) |
| NovelStudioPage | 纯文本 (.txt) | [NovelStudioPage.readable.js:1324](source/frontend/NovelStudioPage.readable.js#L1324) |
| NovelStudioPage | 选择导出格式 | [NovelStudioPage.readable.js:1333](source/frontend/NovelStudioPage.readable.js#L1333) |
| NovelStudioPage | 导出失败 | [NovelStudioPage.readable.js:1353](source/frontend/NovelStudioPage.readable.js#L1353) |
| NovelStudioPage | 导出完成 | [NovelStudioPage.readable.js:1359](source/frontend/NovelStudioPage.readable.js#L1359) |
| NovelStudioPage | 导出失败 | [NovelStudioPage.readable.js:1361](source/frontend/NovelStudioPage.readable.js#L1361) |
| NovelStudioPage | 删除作品 | [NovelStudioPage.readable.js:1375](source/frontend/NovelStudioPage.readable.js#L1375) |
| NovelStudioPage | 未命名 | [NovelStudioPage.readable.js:1376](source/frontend/NovelStudioPage.readable.js#L1376) |
| NovelStudioPage | 删除 | [NovelStudioPage.readable.js:1377](source/frontend/NovelStudioPage.readable.js#L1377) |
| NovelStudioPage | 取消 | [NovelStudioPage.readable.js:1378](source/frontend/NovelStudioPage.readable.js#L1378) |
| NovelStudioPage | 作品已删除 | [NovelStudioPage.readable.js:1392](source/frontend/NovelStudioPage.readable.js#L1392) |
| NovelStudioPage | 删除失败 | [NovelStudioPage.readable.js:1392](source/frontend/NovelStudioPage.readable.js#L1392) |
| NovelStudioPage | 书名已更新 | [NovelStudioPage.readable.js:1415](source/frontend/NovelStudioPage.readable.js#L1415) |
| NovelStudioPage | 重命名失败 | [NovelStudioPage.readable.js:1415](source/frontend/NovelStudioPage.readable.js#L1415) |
| NovelStudioPage | 重命名失败 | [NovelStudioPage.readable.js:1436](source/frontend/NovelStudioPage.readable.js#L1436) |
| NovelStudioPage | 删除失败 | [NovelStudioPage.readable.js:1459](source/frontend/NovelStudioPage.readable.js#L1459) |
| NovelStudioPage |  任务执行中，AI 输出请查看左侧对话框  | [NovelStudioPage.readable.js:1631](source/frontend/NovelStudioPage.readable.js#L1631) |
| NovelStudioPage | ← 返回 | [NovelStudioPage.readable.js:1650](source/frontend/NovelStudioPage.readable.js#L1650) |
| NovelStudioPage | 定时写作 | [NovelStudioPage.readable.js:1652](source/frontend/NovelStudioPage.readable.js#L1652) |
| NovelStudioPage | ← 返回 | [NovelStudioPage.readable.js:1660](source/frontend/NovelStudioPage.readable.js#L1660) |
| NovelStudioPage | 数据看板 | [NovelStudioPage.readable.js:1662](source/frontend/NovelStudioPage.readable.js#L1662) |
| NovelStudioPage | 伏笔管理 | [NovelStudioPage.readable.js:1671](source/frontend/NovelStudioPage.readable.js#L1671) |
| NovelStudioPage | 重命名作品 | [NovelStudioPage.readable.js:1698](source/frontend/NovelStudioPage.readable.js#L1698) |
| NovelStudioPage | 重命名会话 | [NovelStudioPage.readable.js:1698](source/frontend/NovelStudioPage.readable.js#L1698) |
| NovelStudioPage | 保存 | [NovelStudioPage.readable.js:1699](source/frontend/NovelStudioPage.readable.js#L1699) |
| NovelStudioPage | 作品标题 | [NovelStudioPage.readable.js:1705](source/frontend/NovelStudioPage.readable.js#L1705) |
| NovelStudioPage | 会话标题 | [NovelStudioPage.readable.js:1705](source/frontend/NovelStudioPage.readable.js#L1705) |
