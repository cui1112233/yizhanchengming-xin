# 08 当前23组件的输入、输出和边界原式

每个组件按实际props、emits和控件条件登记。表中表达式尚未全部手工解混淆；只代表静态合同，不是运行验收。无emits的展示组件不会被凭空加保存/生成按钮。

## NovelCreateIcon

来源：`source/frontend/NovelCreateIcon.readable.js:1`。原bundle范围：[8506080, 8512308]。

### 输入

| 属性 | 类型/默认表达式 | 行号 |
|---|---|---|
| name | `{<br>  type: String,<br>  required: !0<br>}` | 7 |
| size | `{<br>  type: Number,<br>  default: 18<br>}` | 11 |

### 输出

没有显式emits声明；不推定有后台写入。

### 边界和动作原式

| 编号 | 属性 | 表达式 | 行号 |
|---|---|---|---|

### 失败、空态和取消文案

无独立失败文案字面量；父层/动态分支见对应源码，不能编造。

运行状态：未实测；完整语义与剩余缺口见03/04/06/07。

## NovelStudioSidebar

来源：`source/frontend/NovelStudioSidebar.readable.js:1`。原bundle范围：[8542215, 8550130]。

### 输入

| 属性 | 类型/默认表达式 | 行号 |
|---|---|---|
| books | `{<br>  type: Array,<br>  default: () => []<br>}` | 7 |
| projectSessions | `{<br>  type: Array,<br>  default: () => []<br>}` | 11 |
| modeSessions | `{<br>  type: Array,<br>  default: () => []<br>}` | 15 |
| sessionProgress | `{<br>  type: Object,<br>  default: () => ({})<br>}` | 19 |
| activeWorkflow | `{<br>  type: String,<br>  default: "book-create"<br>}` | 23 |
| activeSessionId | `String` | 27 |
| activeBookId | `String` | 28 |
| activeMode | `String` | 29 |
| activeTool | `String` | 30 |
| expandedBooks | `{<br>  type: Object,<br>  default: () => ({})<br>}` | 31 |
| bookSessions | `{<br>  type: Object,<br>  default: () => ({})<br>}` | 35 |
| isDarkMode | `Boolean` | 39 |

### 输出

`create-session` (:41), `open-session` (:41), `open-book` (:41), `toggle-book` (:41), `create-book-session` (:41), `rename-session` (:41), `delete-session` (:41), `rename-book` (:41), `delete-book` (:41), `export-book` (:41), `launch-mode` (:41), `switch-workflow` (:41), `open-import` (:41), `open-style` (:41), `open-import-modal` (:41), `open-imitation-short` (:41), `open-agent-settings` (:41), `open-scheduler` (:41), `open-dashboard` (:41), `open-memo` (:41)

### 边界和动作原式

| 编号 | 属性 | 表达式 | 行号 |
|---|---|---|---|
| UI-0001 | onClick | `e => function (e) {<br>  if ("imitation-short" === e.key) return a("switch-workflow", "imitation-short"), void a("open-imitation-short");<br>  a("switch-workflow", e.key), "short" !== e.key && "book-create" !== e.key && a("open-import", e.key);<br>}(t)` | 237 |
| UI-0002 | onClick | `n[0] \|\| (n[0] = e => o.value = !o.value)` | 247 |
| UI-0003 | onClick | `e => a("toggle-book", t.id)` | 259 |
| UI-0004 | onClick | `e => a("open-book", t.id)` | 270 |
| UI-0005 | title | `"重命名"` | 276 |
| UI-0006 | onClick | `Zi(e => a("rename-book", t), ["stop"])` | 277 |
| UI-0007 | title | `"导出"` | 283 |
| UI-0008 | onClick | `Zi(e => a("export-book", t), ["stop"])` | 284 |
| UI-0009 | title | `"删除"` | 290 |
| UI-0010 | onClick | `Zi(e => a("delete-book", t), ["stop"])` | 291 |
| UI-0011 | onClick | `e => a("open-session", t.id)` | 302 |
| UI-0012 | onClick | `e => a("create-book-session", t.id)` | 308 |
| UI-0013 | onClick | `n[1] \|\| (n[1] = e => r.value = !r.value)` | 315 |
| UI-0014 | onClick | `e => a("open-session", t.id)` | 329 |
| UI-0015 | onClick | `Zi(e => a("rename-session", t), ["stop"])` | 335 |
| UI-0016 | onClick | `Zi(e => a("delete-session", t), ["stop"])` | 341 |
| UI-0017 | onClick | `n[2] \|\| (n[2] = t => a("create-session", e.activeWorkflow))` | 347 |
| UI-0018 | onClick | `e => {<br>  var n;<br>  "style" === (n = t.key) ? a("open-style") : "import" === n ? a("open-import-modal") : "agent-settings" === n ? a("open-agent-settings") : "scheduler" === n ? a("open-scheduler") : "dashboard" === n ? a("open-dashboard") : "memo" === n && a("open-memo");<br>}` | 357 |

### 失败、空态和取消文案

- `:254` "暂无该类型作品"
- `:322` "当前分类暂无会话，点击下方新建"

运行状态：未实测；完整语义与剩余缺口见03/04/06/07。

## NovelConfirmCard

来源：`source/frontend/NovelConfirmCard.readable.js:1`。原bundle范围：[8550523, 8552978]。

### 输入

| 属性 | 类型/默认表达式 | 行号 |
|---|---|---|
| action | `{<br>  type: Object,<br>  required: !0<br>}` | 7 |
| loading | `Boolean` | 11 |
| locked | `Boolean` | 12 |
| embedded | `Boolean` | 13 |
| streaming | `Boolean` | 14 |
| rejected | `Boolean` | 15 |

### 输出

`confirm` (:17), `dismiss` (:17)

### 边界和动作原式

| 编号 | 属性 | 表达式 | 行号 |
|---|---|---|---|
| UI-0019 | disabled | `e.loading \|\| e.streaming` | 95 |
| UI-0020 | onClick | `n[0] \|\| (n[0] = e => a("confirm"))` | 96 |
| UI-0021 | disabled | `e.loading \|\| e.streaming` | 100 |
| UI-0022 | onClick | `n[1] \|\| (n[1] = e => a("dismiss"))` | 101 |

### 失败、空态和取消文案

- `:92` "已取消"
- `:102` "取消"

运行状态：未实测；完整语义与剩余缺口见03/04/06/07。

## NovelCoverPreview

来源：`source/frontend/NovelCoverPreview.readable.js:1`。原bundle范围：[8573753, 8575641]。

### 输入

| 属性 | 类型/默认表达式 | 行号 |
|---|---|---|
| imagePath | `{<br>  type: String,<br>  default: ""<br>}` | 7 |
| alt | `{<br>  type: String,<br>  default: "封面"<br>}` | 11 |
| compact | `Boolean` | 15 |

### 输出

没有显式emits声明；不推定有后台写入。

### 边界和动作原式

| 编号 | 属性 | 表达式 | 行号 |
|---|---|---|---|

### 失败、空态和取消文案

- `:119` "图片加载失败"

运行状态：未实测；完整语义与剩余缺口见03/04/06/07。

## NovelToolSteps

来源：`source/frontend/NovelToolSteps.readable.js:1`。原bundle范围：[8619062, 8625267]。

### 输入

| 属性 | 类型/默认表达式 | 行号 |
|---|---|---|
| execution | `{<br>  type: Object,<br>  default: null<br>}` | 7 |
| bookTitleFallback | `{<br>  type: String,<br>  default: ""<br>}` | 11 |
| isDarkMode | `{<br>  type: Boolean,<br>  default: !1<br>}` | 15 |
| embeddedAction | `{<br>  type: Object,<br>  default: null<br>}` | 19 |
| actionLocked | `Boolean` | 23 |
| actionRejected | `Boolean` | 24 |
| confirmLoading | `Boolean` | 25 |
| isStreaming | `Boolean` | 26 |
| streamingText | `{<br>  type: String,<br>  default: ""<br>}` | 27 |
| streamingReasoning | `{<br>  type: String,<br>  default: ""<br>}` | 31 |
| streamStage | `{<br>  type: String,<br>  default: ""<br>}` | 35 |
| hideStreamOutput | `Boolean` | 39 |

### 输出

`retry` (:41), `confirm` (:41), `dismiss` (:41)

### 边界和动作原式

| 编号 | 属性 | 表达式 | 行号 |
|---|---|---|---|
| UI-0023 | onClick | `n[0] \|\| (n[0] = e => o.value = !o.value)` | 152 |
| UI-0024 | onConfirm | `n[1] \|\| (n[1] = e => a("confirm"))` | 180 |
| UI-0025 | onDismiss | `n[2] \|\| (n[2] = e => a("dismiss"))` | 181 |
| UI-0026 | onClick | `n[3] \|\| (n[3] = Zi(e => a("retry"), ["stop"]))` | 202 |
| UI-0027 | onClick | `n[4] \|\| (n[4] = Zi(e => a("retry"), ["stop"]))` | 206 |

### 失败、空态和取消文案

- `:58` "失败"
- `:58` "等待中"

运行状态：未实测；完整语义与剩余缺口见03/04/06/07。

## NovelThinkingBlock

来源：`source/frontend/NovelThinkingBlock.readable.js:1`。原bundle范围：[8625439, 8626316]。

### 输入

| 属性 | 类型/默认表达式 | 行号 |
|---|---|---|
| seconds | `{<br>  type: Number,<br>  default: 0<br>}` | 7 |
| reasoning | `{<br>  type: String,<br>  default: ""<br>}` | 11 |
| defaultOpen | `{<br>  type: Boolean,<br>  default: !1<br>}` | 15 |
| thinkingMode | `{<br>  type: Boolean,<br>  default: !1<br>}` | 19 |

### 输出

`update:expanded` (:24)

### 边界和动作原式

| 编号 | 属性 | 表达式 | 行号 |
|---|---|---|---|
| UI-0028 | onClick | `i` | 42 |

### 失败、空态和取消文案

- `:49` "暂无思考内容记录"

运行状态：未实测；完整语义与剩余缺口见03/04/06/07。

## NovelStreamingIndicator

来源：`source/frontend/NovelStreamingIndicator.readable.js:1`。原bundle范围：[8626359, 8626704]。

### 输入

| 属性 | 类型/默认表达式 | 行号 |
|---|---|---|
| active | `{<br>  type: Boolean,<br>  default: !0<br>}` | 7 |
| isDarkMode | `Boolean` | 11 |

### 输出

没有显式emits声明；不推定有后台写入。

### 边界和动作原式

| 编号 | 属性 | 表达式 | 行号 |
|---|---|---|---|

### 失败、空态和取消文案

无独立失败文案字面量；父层/动态分支见对应源码，不能编造。

运行状态：未实测；完整语义与剩余缺口见03/04/06/07。

## NovelQuickActions

来源：`source/frontend/NovelQuickActions.readable.js:1`。原bundle范围：[8626774, 8627496]。

### 输入

| 属性 | 类型/默认表达式 | 行号 |
|---|---|---|
| disabled | `Boolean` | 7 |
| loading | `Boolean` | 8 |
| isDarkMode | `Boolean` | 9 |

### 输出

`write-next` (:11), `audit` (:11), `revise` (:11), `write-batch` (:11)

### 边界和动作原式

| 编号 | 属性 | 表达式 | 行号 |
|---|---|---|---|
| UI-0029 | disabled | `Boolean` | 7 |
| UI-0030 | disabled | `e.disabled \|\| e.loading` | 43 |
| UI-0031 | onClick | `e => {<br>  var a;<br>  "write-next" === (a = t.key) ? n("write-next") : "write-batch" === a ? n("write-batch") : "audit" === a ? n("audit") : "revise" === a && n("revise");<br>}` | 44 |

### 失败、空态和取消文案

无独立失败文案字面量；父层/动态分支见对应源码，不能编造。

运行状态：未实测；完整语义与剩余缺口见03/04/06/07。

## NovelStudioChat

来源：`source/frontend/NovelStudioChat.readable.js:1`。原bundle范围：[8629071, 8650803]。

### 输入

| 属性 | 类型/默认表达式 | 行号 |
|---|---|---|
| session | `{<br>  type: Object,<br>  default: null<br>}` | 7 |
| isStreaming | `Boolean` | 11 |
| streamingText | `String` | 12 |
| streamingReasoning | `String` | 13 |
| pipelineStreamStage | `{<br>  type: String,<br>  default: ""<br>}` | 14 |
| toolExecution | `{<br>  type: Object,<br>  default: null<br>}` | 18 |
| pendingAction | `{<br>  type: Object,<br>  default: null<br>}` | 22 |
| confirmLoading | `Boolean` | 26 |
| showQuickActions | `Boolean` | 27 |
| quickActionLoading | `Boolean` | 28 |
| modelOptions | `{<br>  type: Array,<br>  default: () => []<br>}` | 29 |
| selectedModel | `String` | 33 |
| thinkingMode | `{<br>  type: Boolean,<br>  default: !0<br>}` | 34 |
| thinkingAutoExpand | `{<br>  type: Boolean,<br>  default: !0<br>}` | 38 |
| isDarkMode | `Boolean` | 42 |
| bookTitleFallback | `{<br>  type: String,<br>  default: ""<br>}` | 43 |

### 输出

`send` (:48), `confirm-action` (:48), `dismiss-action` (:48), `cancel` (:48), `quick-write-next` (:48), `quick-write-batch` (:48), `quick-audit` (:48), `quick-revise` (:48), `delete-message` (:48), `archive-message` (:48), `update:selectedModel` (:48), `update:thinkingAutoExpand` (:48), `clear-context` (:48), `retry-action` (:48)

### 边界和动作原式

| 编号 | 属性 | 表达式 | 行号 |
|---|---|---|---|
| UI-0032 | onClick | `e => pe(t, o)` | 346 |
| UI-0033 | onClick | `e => fe(t)` | 365 |
| UI-0034 | onPositiveClick | `e => me(o, t)` | 375 |
| UI-0035 | onClick | `n[0] \|\| (n[0] = Zi(() => {}, ["stop"]))` | 384 |
| UI-0036 | onUpdate:expanded | `e => t.reasoningExpanded = e` | 406 |
| UI-0037 | onConfirm | `n[1] \|\| (n[1] = e => a("confirm-action"))` | 416 |
| UI-0038 | onDismiss | `n[2] \|\| (n[2] = e => a("dismiss-action"))` | 417 |
| UI-0039 | onClick | `e => pe(t, o)` | 427 |
| UI-0040 | onClick | `e => fe(t)` | 446 |
| UI-0041 | onPositiveClick | `e => me(o, t)` | 456 |
| UI-0042 | onClick | `n[3] \|\| (n[3] = Zi(() => {}, ["stop"]))` | 465 |
| UI-0043 | onUpdate:expanded | `n[4] \|\| (n[4] = e => k.value = e)` | 489 |
| UI-0044 | onClick | `he` | 504 |
| UI-0045 | onConfirm | `n[5] \|\| (n[5] = e => a("confirm-action"))` | 525 |
| UI-0046 | onDismiss | `n[6] \|\| (n[6] = e => a("dismiss-action"))` | 526 |
| UI-0047 | onMouseenter | `D` | 534 |
| UI-0048 | onMouseleave | `L` | 535 |
| UI-0049 | onMouseenter | `D` | 542 |
| UI-0050 | onMouseleave | `L` | 543 |
| UI-0051 | onMouseenter | `t => P.value = e.id` | 556 |
| UI-0052 | onMouseleave | `n[7] \|\| (n[7] = e => P.value = null)` | 557 |
| UI-0053 | onClick | `t => function (e) {<br>  var t;<br>  const n = i.value;<br>  if (!n) return;<br>  const a = null == (t = n.$el) ? void 0 : t.querySelector(".n-scrollbar-container");<br>  if (!a) return;<br>  const o = a.querySelector(`.msg-row[data-msg-id="${e}"]`);<br>  if (!o) return;<br>  const r = o.offsetTop - a.clientHeight / 2 + o.offsetHeight / 2;<br>  n.scrollTo({<br>    top: Math.max(0, r),<br>    behavior: "smooth"<br>  }), o.classList.add("msg-nav-highlight"), setTimeout(() => o.classList.remove("msg-nav-highlight"), 1500), _.value = !1, P.value = null;<br>}(e.id)` | 558 |
| UI-0054 | disabled | `e.isStreaming \|\| e.confirmLoading \|\| !!j.value` | 584 |
| UI-0055 | onWriteNext | `n[8] \|\| (n[8] = e => a("quick-write-next"))` | 587 |
| UI-0056 | onWriteBatch | `n[9] \|\| (n[9] = e => a("quick-write-batch"))` | 588 |
| UI-0057 | onAudit | `n[10] \|\| (n[10] = e => a("quick-audit"))` | 589 |
| UI-0058 | onRevise | `n[11] \|\| (n[11] = e => a("quick-revise"))` | 590 |
| UI-0059 | onUpdate:modelValue | `n[12] \|\| (n[12] = e => l.value = e)` | 594 |
| UI-0060 | placeholder | `oe.value` | 596 |
| UI-0061 | disabled | `e.isStreaming` | 598 |
| UI-0062 | onKeydown | `le` | 599 |
| UI-0063 | onCompositionstart | `n[13] \|\| (n[13] = e => c.value = !0)` | 600 |
| UI-0064 | onCompositionend | `n[14] \|\| (n[14] = e => c.value = !1)` | 601 |
| UI-0065 | onClick | `ce` | 604 |
| UI-0066 | title | `"展开思考"` | 620 |
| UI-0067 | onClick | `n[15] \|\| (n[15] = e => f.value = !f.value)` | 621 |
| UI-0068 | onClick | `n[16] \|\| (n[16] = t => {<br>  a("update:thinkingAutoExpand", !e.thinkingAutoExpand), f.value = !1;<br>})` | 642 |
| UI-0069 | onClick | `n[17] \|\| (n[17] = e => f.value = !1)` | 650 |
| UI-0070 | onClick | `n[18] \|\| (n[18] = e => h.value = !h.value)` | 657 |
| UI-0071 | onClick | `ie` | 671 |
| UI-0072 | onClick | `n[19] \|\| (n[19] = e => m.value = 1e5)` | 683 |
| UI-0073 | onClick | `n[20] \|\| (n[20] = e => m.value = 2e5)` | 685 |
| UI-0074 | onClick | `n[21] \|\| (n[21] = e => m.value = 3e5)` | 687 |
| UI-0075 | onClick | `n[22] \|\| (n[22] = e => m.value = 4e5)` | 689 |
| UI-0076 | onClick | `n[23] \|\| (n[23] = e => m.value = 1e6)` | 691 |
| UI-0077 | onClick | `n[24] \|\| (n[24] = e => m.value = van)` | 693 |
| UI-0078 | onClick | `n[25] \|\| (n[25] = e => h.value = !1)` | 700 |
| UI-0079 | onClick | `se` | 707 |
| UI-0080 | onClick | `n[26] \|\| (n[26] = e => a("cancel"))` | 718 |
| UI-0081 | disabled | `!l.value.trim()` | 736 |
| UI-0082 | onClick | `re` | 737 |
| UI-0083 | onUpdate:show | `n[29] \|\| (n[29] = e => u.value = e)` | 743 |
| UI-0084 | onUpdate:modelValue | `n[27] \|\| (n[27] = e => d.value = e)` | 759 |
| UI-0085 | placeholder | `"搜索模型..."` | 761 |
| UI-0086 | onClick | `n[28] \|\| (n[28] = e => u.value = !1)` | 766 |
| UI-0087 | onClick | `t => {<br>  return n = e.key, void (v.value = v.value === n ? "" : n);<br>  var n;<br>}` | 777 |
| UI-0088 | onClick | `e => {<br>  return n = t.value, a("update:selectedModel", n), void (u.value = !1);<br>  var n;<br>}` | 786 |

### 失败、空态和取消文案

- `:267` "复制失败"
- `:277` "请等待当前回复完成"
- `:289` "复制失败"
- `:374` "取消"
- `:455` "取消"

运行状态：未实测；完整语义与剩余缺口见03/04/06/07。

## NovelPanelArtifactView

来源：`source/frontend/NovelPanelArtifactView.readable.js:1`。原bundle范围：[8651068, 8652874]。

### 输入

| 属性 | 类型/默认表达式 | 行号 |
|---|---|---|
| title | `{<br>  type: String,<br>  default: ""<br>}` | 7 |
| content | `{<br>  type: String,<br>  default: ""<br>}` | 11 |
| loading | `Boolean` | 15 |
| saving | `Boolean` | 16 |
| readOnly | `Boolean` | 17 |
| autoEdit | `Boolean` | 18 |
| isDarkMode | `Boolean` | 19 |

### 输出

`back` (:21), `save` (:21)

### 边界和动作原式

| 编号 | 属性 | 表达式 | 行号 |
|---|---|---|---|
| UI-0089 | title | `{<br>  type: String,<br>  default: ""<br>}` | 7 |
| UI-0090 | title | `"返回"` | 51 |
| UI-0091 | onClick | `n[0] \|\| (n[0] = e => a("back"))` | 52 |
| UI-0092 | title | `"保存"` | 63 |
| UI-0093 | disabled | `e.saving` | 64 |
| UI-0094 | onClick | `s` | 65 |
| UI-0095 | title | `"取消"` | 76 |
| UI-0096 | onClick | `i` | 77 |
| UI-0097 | title | `"编辑"` | 85 |
| UI-0098 | onClick | `l` | 86 |
| UI-0099 | onUpdate:modelValue | `n[1] \|\| (n[1] = e => r.value = e)` | 97 |

### 失败、空态和取消文案

- `:76` "取消"
- `:95` "文件不存在"

运行状态：未实测；完整语义与剩余缺口见03/04/06/07。

## NovelStudioBookPanel

来源：`source/frontend/NovelStudioBookPanel.readable.js:1`。原bundle范围：[8683553, 8691226]。

### 输入

| 属性 | 类型/默认表达式 | 行号 |
|---|---|---|
| book | `{<br>  type: Object,<br>  default: null<br>}` | 7 |
| bookId | `{<br>  type: String,<br>  default: ""<br>}` | 11 |
| collapsed | `{<br>  type: Boolean,<br>  default: !1<br>}` | 15 |
| isDarkMode | `Boolean` | 19 |

### 输出

`update:collapsed` (:21), `saved` (:21)

### 边界和动作原式

| 编号 | 属性 | 表达式 | 行号 |
|---|---|---|---|
| UI-0100 | title | `"故事基石 / 世界观"` | 147 |
| UI-0101 | onClick | `n[0] \|\| (n[0] = e => a("update:collapsed", !0))` | 210 |
| UI-0102 | onClick | `n[1] \|\| (n[1] = e => W("chapters"))` | 217 |
| UI-0103 | onClick | `t => function (e) {<br>  b({<br>    type: "chapter",<br>    chapterNumber: e.number,<br>    title: `第 ${e.number} 章 ${e.title}`<br>  });<br>}(e)` | 225 |
| UI-0104 | title | ``第 ${e.number} 章 ${e.title}`` | 229 |
| UI-0105 | title | `"编辑"` | 238 |
| UI-0106 | onClick | `Zi(t => function (e) {<br>  b({<br>    type: "chapter",<br>    chapterNumber: e.number,<br>    title: `第 ${e.number} 章 ${e.title}`,<br>    autoEdit: !0<br>  });<br>}(e), ["stop"])` | 239 |
| UI-0107 | title | ``第 ${e.number} 章 ${e.title}`` | 243 |
| UI-0108 | onClick | `n[2] \|\| (n[2] = e => W("characters"))` | 253 |
| UI-0109 | onClick | `t => function (e) {<br>  b({<br>    type: "character",<br>    key: e.path.replace(/\.md$/, ""),<br>    title: e.name<br>  });<br>}(e)` | 262 |
| UI-0110 | title | `e.name` | 266 |
| UI-0111 | onClick | `n[3] \|\| (n[3] = e => W("foundation"))` | 279 |
| UI-0112 | onClick | `t => function (e) {<br>  b({<br>    type: "foundation",<br>    key: e.key,<br>    title: e.label<br>  });<br>}(e)` | 287 |
| UI-0113 | title | `e.label` | 291 |
| UI-0114 | title | `"编辑"` | 300 |
| UI-0115 | onClick | `Zi(t => function (e) {<br>  b({<br>    type: "foundation",<br>    key: e.key,<br>    title: e.label,<br>    autoEdit: !0<br>  });<br>}(e), ["stop"])` | 301 |
| UI-0116 | title | `e.label` | 305 |
| UI-0117 | onClick | `n[4] \|\| (n[4] = e => W("covers"))` | 315 |
| UI-0118 | onClick | `n[5] \|\| (n[5] = e => W("worldview"))` | 329 |
| UI-0119 | onClick | `k` | 336 |
| UI-0120 | title | `h.value` | 341 |
| UI-0121 | onBack | `y` | 348 |
| UI-0122 | onSave | `x` | 349 |

### 失败、空态和取消文案

- `:189` "保存失败"
- `:250` "暂无章节"
- `:276` "暂无角色档案"
- `:312` "暂无核心文件"
- `:333` "暂无世界观摘要"

运行状态：未实测；完整语义与剩余缺口见03/04/06/07。

## NovelStudioShortPanel

来源：`source/frontend/NovelStudioShortPanel.readable.js:1`。原bundle范围：[8691913, 8697920]。

### 输入

| 属性 | 类型/默认表达式 | 行号 |
|---|---|---|
| shortStory | `{<br>  type: Object,<br>  default: null<br>}` | 7 |
| shortStoryId | `{<br>  type: String,<br>  default: ""<br>}` | 11 |
| collapsed | `{<br>  type: Boolean,<br>  default: !1<br>}` | 15 |
| isDarkMode | `Boolean` | 19 |

### 输出

`update:collapsed` (:21), `saved` (:21)

### 边界和动作原式

| 编号 | 属性 | 表达式 | 行号 |
|---|---|---|---|
| UI-0123 | title | `e.label \|\| e.title \|\| "文件"` | 161 |
| UI-0124 | title | `"短篇大纲"` | 169 |
| UI-0125 | title | `"推荐"` | 175 |
| UI-0126 | onClick | `n[0] \|\| (n[0] = e => a("update:collapsed", !0))` | 192 |
| UI-0127 | onClick | `n[1] \|\| (n[1] = e => W("chapters"))` | 205 |
| UI-0128 | onClick | `t => x({<br>  key: "final/full.md",<br>  title: e.title<br>})` | 213 |
| UI-0129 | title | `e.title` | 215 |
| UI-0130 | onClick | `n[2] \|\| (n[2] = e => W("outline"))` | 223 |
| UI-0131 | onClick | `w` | 231 |
| UI-0132 | onClick | `n[3] \|\| (n[3] = e => W("artifacts"))` | 235 |
| UI-0133 | onClick | `t => x(e)` | 243 |
| UI-0134 | title | `"编辑"` | 250 |
| UI-0135 | onClick | `Zi(t => function (e) {<br>  b({<br>    key: e.key,<br>    title: e.label \|\| e.title \|\| "文件",<br>    autoEdit: !0<br>  });<br>}(e), ["stop"])` | 251 |
| UI-0136 | title | `e.label \|\| e.title \|\| "文件"` | 254 |
| UI-0137 | onClick | `n[4] \|\| (n[4] = e => W("sales"))` | 264 |
| UI-0138 | onClick | `_` | 271 |
| UI-0139 | title | `g.value` | 276 |
| UI-0140 | onBack | `y` | 283 |
| UI-0141 | onSave | `k` | 284 |

### 失败、空态和取消文案

- `:151` "保存失败"
- `:261` "暂无成品文件"

运行状态：未实测；完整语义与剩余缺口见03/04/06/07。

## NovelStylePanel

来源：`source/frontend/NovelStylePanel.readable.js:1`。原bundle范围：[8698363, 8704788]。

### 输入

| 属性 | 类型/默认表达式 | 行号 |
|---|---|---|
| styleSources | `{<br>  type: Array,<br>  default: () => []<br>}` | 7 |
| selectedStyleIds | `{<br>  type: Array,<br>  default: () => []<br>}` | 11 |
| isDarkMode | `Boolean` | 15 |

### 输出

`update:selectedStyleIds` (:17), `back` (:17), `refresh` (:17), `applied` (:17)

### 边界和动作原式

| 编号 | 属性 | 表达式 | 行号 |
|---|---|---|---|
| UI-0142 | title | `i.value.trim()` | 85 |
| UI-0143 | onClick | `n[0] \|\| (n[0] = e => a("back"))` | 208 |
| UI-0144 | onClick | `p` | 216 |
| UI-0145 | onUpdate:value | `n[1] \|\| (n[1] = e => i.value = e)` | 219 |
| UI-0146 | placeholder | `"例如：金庸武侠、知乎盐选、某作者口语风"` | 220 |
| UI-0147 | onUpdate:value | `n[2] \|\| (n[2] = e => s.value = e)` | 223 |
| UI-0148 | placeholder | `"直接粘贴样章、段落或写作要求……"` | 229 |
| UI-0149 | onUpdate:value | `n[3] \|\| (n[3] = e => c.value = e)` | 232 |
| UI-0150 | placeholder | `"可手写可执行的文风规则；不填则根据参考文本自动生成"` | 238 |
| UI-0151 | onClick | `h` | 241 |
| UI-0152 | onClick | `m` | 248 |
| UI-0153 | onUpdate:value | `n[4] \|\| (n[4] = e => r.value = e)` | 255 |
| UI-0154 | onClick | `Zi(t => f(e), ["prevent"])` | 266 |
| UI-0155 | onClick | `t => g(e)` | 270 |
| UI-0156 | onClick | `W` | 278 |

### 失败、空态和取消文案

- `:58` "加载文风失败"
- `:62` "请填写风格名称"
- `:63` "请粘贴参考文本，或导入文件"
- `:89` "保存失败"
- `:165` "导入失败"
- `:190` "删除失败"
- `:274` "还没有风格。请在上方新建并保存。"

运行状态：未实测；完整语义与剩余缺口见03/04/06/07。

## NovelAgentSettings

来源：`source/frontend/NovelAgentSettings.readable.js:1`。原bundle范围：[8744349, 8756515]。

### 输入

| 属性 | 类型/默认表达式 | 行号 |
|---|---|---|
| modelOptions | `{<br>  type: Array,<br>  default: () => []<br>}` | 7 |
| isDarkMode | `Boolean` | 11 |

### 输出

`back` (:13)

### 边界和动作原式

| 编号 | 属性 | 表达式 | 行号 |
|---|---|---|---|
| UI-0157 | title | `e.title` | 108 |
| UI-0158 | title | `e.title` | 114 |
| UI-0159 | title | `e.title` | 146 |
| UI-0160 | title | `"自定义审稿"` | 188 |
| UI-0161 | title | `"自定义写作要求"` | 211 |
| UI-0162 | placeholder | `""` | 214 |
| UI-0163 | onClick | `n[0] \|\| (n[0] = e => a("back"))` | 279 |
| UI-0164 | onUpdate:modelValue | `e => s.value[t.key] = e` | 291 |
| UI-0165 | placeholder | `"跟随默认"` | 293 |
| UI-0166 | onUpdate:value | `e => s.value[t.key] = e` | 301 |
| UI-0167 | placeholder | `"默认 gemini 出图"` | 303 |
| UI-0168 | onUpdate:value | `n[1] \|\| (n[1] = e => d.value.maxRetries = e)` | 312 |
| UI-0169 | min | `0` | 313 |
| UI-0170 | max | `5` | 314 |
| UI-0171 | onUpdate:value | `n[2] \|\| (n[2] = e => d.value.passScore = e)` | 318 |
| UI-0172 | min | `0` | 319 |
| UI-0173 | max | `100` | 320 |
| UI-0174 | onUpdate:value | `n[3] \|\| (n[3] = e => d.value.temperatureStep = e)` | 324 |
| UI-0175 | min | `0` | 325 |
| UI-0176 | max | `.5` | 326 |
| UI-0177 | onUpdate:value | `n[4] \|\| (n[4] = e => d.value.baseTemperature = e)` | 331 |
| UI-0178 | min | `0` | 332 |
| UI-0179 | max | `2` | 333 |
| UI-0180 | onUpdate:value | `n[5] \|\| (n[5] = e => d.value.enableLengthNorm = e)` | 338 |
| UI-0181 | onUpdate:value | `n[6] \|\| (n[6] = e => d.value.enablePostWrite = e)` | 344 |
| UI-0182 | onUpdate:value | `n[7] \|\| (n[7] = e => d.value.enableAiTells = e)` | 350 |
| UI-0183 | onUpdate:value | `n[8] \|\| (n[8] = e => d.value.requireSecondaryConfirm = e)` | 356 |
| UI-0184 | onClick | `n[9] \|\| (n[9] = e => W(!0))` | 363 |
| UI-0185 | onClick | `n[10] \|\| (n[10] = e => W(!1))` | 370 |
| UI-0186 | onUpdate:checked | `t => function (e, t) {<br>  c.value = c.value.map(n => n.id === e.id ? {<br>    ...n,<br>    enabled: t<br>  } : n);<br>}(e, t)` | 379 |
| UI-0187 | onClick | `k` | 390 |
| UI-0188 | onUpdate:checked | `t => x(e, "enabled", t)` | 399 |
| UI-0189 | placeholder | `"提示词标题"` | 403 |
| UI-0190 | onUpdate:value | `t => x(e, "title", t)` | 405 |
| UI-0191 | onClick | `t => function (e) {<br>  c.value = c.value.filter(t => t.id !== e.id);<br>}(e)` | 409 |
| UI-0192 | placeholder | `"审稿要求说明…"` | 419 |
| UI-0193 | onUpdate:value | `t => x(e, "content", t)` | 420 |
| UI-0194 | onClick | `_` | 426 |
| UI-0195 | onUpdate:checked | `t => w(e, "enabled", t)` | 435 |
| UI-0196 | placeholder | `"提示词标题"` | 439 |
| UI-0197 | onUpdate:value | `t => w(e, "title", t)` | 441 |
| UI-0198 | onUpdate:value | `t => w(e, "category", t)` | 449 |
| UI-0199 | onClick | `t => function (e) {<br>  u.value = u.value.filter(t => t.id !== e.id);<br>}(e)` | 453 |
| UI-0200 | placeholder | `e.placeholder \|\| "写作要求说明……"` | 463 |
| UI-0201 | onUpdate:value | `t => w(e, "content", t)` | 464 |
| UI-0202 | disabled | `r.value` | 472 |
| UI-0203 | onClick | `b` | 473 |

### 失败、空态和取消文案

- `:161` "保存失败"
- `:250` "加载 Agent 设置失败"
- `:394` "暂无自定义项，可点击「新增」添加。"
- `:430` "暂无自定义写作提示词。"

运行状态：未实测；完整语义与剩余缺口见03/04/06/07。

## NovelImportPage

来源：`source/frontend/NovelImportPage.readable.js:1`。原bundle范围：[8809779, 8830298]。

### 输入

| 属性 | 类型/默认表达式 | 行号 |
|---|---|---|
| tab | `{<br>  type: String,<br>  default: "imitation"<br>}` | 7 |
| variant | `{<br>  type: String,<br>  default: "page"<br>}` | 11 |
| sessionId | `{<br>  type: String,<br>  default: ""<br>}` | 15 |
| styleSources | `{<br>  type: Array,<br>  default: () => []<br>}` | 19 |
| books | `{<br>  type: Array,<br>  default: () => []<br>}` | 23 |
| isDarkMode | `Boolean` | 27 |

### 输出

`back` (:29), `close` (:29), `created` (:29), `update:tab` (:29), `submit-start` (:29)

### 边界和动作原式

| 编号 | 属性 | 表达式 | 行号 |
|---|---|---|---|
| UI-0204 | title | `"仿写创作"` | 130 |
| UI-0205 | title | `"同人创作"` | 135 |
| UI-0206 | title | `"番外创作"` | 140 |
| UI-0207 | title | `"续写创作"` | 145 |
| UI-0208 | title | `"导入素材"` | 150 |
| UI-0209 | title | `n` | 208 |
| UI-0210 | title | `"第1章"` | 214 |
| UI-0211 | title | `i.value.trim()` | 317 |
| UI-0212 | onClick | `le` | 467 |
| UI-0213 | onClick | `le` | 471 |
| UI-0214 | onClick | `t => Q.value = e.key` | 478 |
| UI-0215 | onUpdate:value | `n[0] \|\| (n[0] = e => i.value = e)` | 489 |
| UI-0216 | placeholder | `"留空则自动生成"` | 490 |
| UI-0217 | disabled | `l.value` | 491 |
| UI-0218 | onUpdate:value | `n[1] \|\| (n[1] = e => c.value = e)` | 500 |
| UI-0219 | placeholder | `"粘贴参考小说全文（至少 100 字），或上传 txt / docx 自动填入"` | 503 |
| UI-0220 | disabled | `l.value` | 504 |
| UI-0221 | disabled | `l.value \|\| W.value` | 510 |
| UI-0222 | onClick | `X` | 511 |
| UI-0223 | onClick | `e => {<br>  return n = t, y.value.splice(n, 1), void (y.value.length \|\| (b.value = [], k.value = []));<br>  var n;<br>}` | 518 |
| UI-0224 | onClick | `n[2] \|\| (n[2] = e => $(3))` | 528 |
| UI-0225 | onClick | `n[3] \|\| (n[3] = e => $(5))` | 534 |
| UI-0226 | onClick | `n[4] \|\| (n[4] = e => $(10))` | 541 |
| UI-0227 | onClick | `te` | 547 |
| UI-0228 | onUpdate:value | `n[5] \|\| (n[5] = e => _.value = e)` | 564 |
| UI-0229 | min | `1e3` | 565 |
| UI-0230 | max | `5e4` | 566 |
| UI-0231 | placeholder | `"5000"` | 569 |
| UI-0232 | disabled | `l.value` | 570 |
| UI-0233 | disabled | `l.value` | 578 |
| UI-0234 | onClick | `t => {<br>  return n = e.value, void (_.value = n);<br>  var n;<br>}` | 579 |
| UI-0235 | onUpdate:value | `n[6] \|\| (n[6] = e => S.value = e)` | 587 |
| UI-0236 | min | `1` | 588 |
| UI-0237 | max | `500` | 589 |
| UI-0238 | placeholder | `"100"` | 592 |
| UI-0239 | disabled | `l.value` | 593 |
| UI-0240 | disabled | `l.value` | 601 |
| UI-0241 | onClick | `t => S.value = e.value` | 602 |
| UI-0242 | onUpdate:value | `n[7] \|\| (n[7] = e => C.value = e)` | 607 |
| UI-0243 | placeholder | `"填写题材偏好，如：玄幻、都市"` | 608 |
| UI-0244 | disabled | `l.value` | 610 |
| UI-0245 | disabled | `l.value` | 615 |
| UI-0246 | onClick | `t => function (e) {<br>  const t = C.value.trim();<br>  if (!t) return void (C.value = e);<br>  t.split(/[、,，\s]+/).filter(Boolean).includes(e) \|\| (C.value = `${t}、${e}`);<br>}(e)` | 616 |
| UI-0247 | onUpdate:value | `n[8] \|\| (n[8] = e => s.value = e)` | 625 |
| UI-0248 | placeholder | `"剧情、主角、核心冲突（原创剧情）"` | 628 |
| UI-0249 | disabled | `l.value` | 629 |
| UI-0250 | onUpdate:value | `n[9] \|\| (n[9] = e => m.value = e)` | 632 |
| UI-0251 | placeholder | `"留空则默认采用拆书后的方向一；可填「方向二：…」"` | 635 |
| UI-0252 | disabled | `l.value` | 636 |
| UI-0253 | onUpdate:value | `n[10] \|\| (n[10] = e => P.value = e)` | 639 |
| UI-0254 | placeholder | `"填写额外要求，如人设、情节走向、必须出现的桥段、禁忌点等"` | 642 |
| UI-0255 | disabled | `l.value` | 643 |
| UI-0256 | onUpdate:value | `n[11] \|\| (n[11] = e => u.value = e)` | 646 |
| UI-0257 | onUpdate:value | `n[12] \|\| (n[12] = e => h.value = e)` | 657 |
| UI-0258 | placeholder | `"选择备忘录作为参考素材"` | 659 |
| UI-0259 | multiple | `""` | 660 |
| UI-0260 | disabled | `l.value` | 662 |
| UI-0261 | onFocus | `G` | 663 |
| UI-0262 | onUpdate:value | `n[13] \|\| (n[13] = e => d.value = e)` | 668 |
| UI-0263 | placeholder | `"选择已有正传"` | 670 |
| UI-0264 | onUpdate:value | `n[14] \|\| (n[14] = e => v.value = e)` | 674 |
| UI-0265 | placeholder | `"选择已有书籍或留空新建"` | 676 |
| UI-0266 | onUpdate:value | `n[15] \|\| (n[15] = e => i.value = e)` | 680 |
| UI-0267 | placeholder | `"暂定书名"` | 681 |
| UI-0268 | onUpdate:value | `n[16] \|\| (n[16] = e => f.value = e)` | 684 |
| UI-0269 | onUpdate:value | `n[17] \|\| (n[17] = e => s.value = e)` | 700 |
| UI-0270 | placeholder | `"番外要写什么"` | 703 |
| UI-0271 | onUpdate:value | `n[18] \|\| (n[18] = e => s.value = e)` | 706 |
| UI-0272 | placeholder | `"CP、AU、IF 线等"` | 709 |
| UI-0273 | onUpdate:value | `n[19] \|\| (n[19] = e => p.value = e)` | 712 |
| UI-0274 | onUpdate:value | `n[20] \|\| (n[20] = e => c.value = e)` | 728 |
| UI-0275 | placeholder | `"chapters" === Q.value ? "粘贴章节，建议以「第1章」等标题分段" : "粘贴参考原文；也可只选下方文风"` | 731 |
| UI-0276 | onUpdate:value | `n[21] \|\| (n[21] = e => u.value = e)` | 734 |
| UI-0277 | onClick | `le` | 752 |
| UI-0278 | onClick | `ie` | 760 |
| UI-0279 | onUpdate:show | `n[25] \|\| (n[25] = e => x.value = e)` | 766 |
| UI-0280 | onClick | `n[22] \|\| (n[22] = e => x.value = !1)` | 782 |
| UI-0281 | onClick | `ae` | 786 |
| UI-0282 | onClick | `oe` | 790 |
| UI-0283 | onClick | `Zi(e => ne(t), ["prevent"])` | 802 |
| UI-0284 | onClick | `n[23] \|\| (n[23] = Zi(() => {}, ["stop"]))` | 805 |
| UI-0285 | onUpdate:checked | `e => ne(t)` | 806 |
| UI-0286 | onClick | `n[24] \|\| (n[24] = e => x.value = !1)` | 815 |
| UI-0287 | onClick | `re` | 821 |

### 失败、空态和取消文案

- `:238` "当前环境不支持文件选择"
- `:264` "读取文件失败"
- `:303` "请粘贴或上传参考原文"
- `:304` "参考原文至少需要 100 字"
- `:305` "请填写书名"
- `:306` "请粘贴原作参考文本"
- `:307` "请选择正传书籍"
- `:308` "请选择目标书籍或填写新书书名"
- `:309` "请粘贴续写章节正文"
- `:421` "未能解析出章节，请检查格式"
- `:439` "创建失败"
- `:503` "粘贴参考小说全文（至少 100 字），或上传 txt / docx 自动填入"
- `:653` "暂无可用文风"
- `:741` "暂无可用文风"
- `:754` "取消"
- `:817` "取消"

运行状态：未实测；完整语义与剩余缺口见03/04/06/07。

## NovelImportModal

来源：`source/frontend/NovelImportModal.readable.js:1`。原bundle范围：[8830352, 8831305]。

### 输入

| 属性 | 类型/默认表达式 | 行号 |
|---|---|---|
| show | `Boolean` | 7 |
| tab | `{<br>  type: String,<br>  default: "imitation"<br>}` | 8 |
| sessionId | `{<br>  type: String,<br>  default: ""<br>}` | 12 |
| styleSources | `{<br>  type: Array,<br>  default: () => []<br>}` | 16 |
| books | `{<br>  type: Array,<br>  default: () => []<br>}` | 20 |
| isDarkMode | `Boolean` | 24 |

### 输出

`update:show` (:26), `update:tab` (:26), `created` (:26), `submit-start` (:26)

### 边界和动作原式

| 编号 | 属性 | 表达式 | 行号 |
|---|---|---|---|
| UI-0288 | onUpdate:show | `r[2] \|\| (r[2] = e => n("update:show", e))` | 42 |
| UI-0289 | onUpdate:tab | `r[0] \|\| (r[0] = e => n("update:tab", e))` | 55 |
| UI-0290 | onClose | `a` | 56 |
| UI-0291 | onCreated | `o` | 57 |
| UI-0292 | onSubmitStart | `r[1] \|\| (r[1] = e => n("submit-start", e))` | 58 |

### 失败、空态和取消文案

无独立失败文案字面量；父层/动态分支见对应源码，不能编造。

运行状态：未实测；完整语义与剩余缺口见03/04/06/07。

## NovelImitationShortDialog

来源：`source/frontend/NovelImitationShortDialog.readable.js:1`。原bundle范围：[8832348, 8840980]。

### 输入

| 属性 | 类型/默认表达式 | 行号 |
|---|---|---|
| show | `Boolean` | 7 |
| sessionId | `{<br>  type: String,<br>  default: ""<br>}` | 8 |
| styleSources | `{<br>  type: Array,<br>  default: () => []<br>}` | 12 |
| books | `{<br>  type: Array,<br>  default: () => []<br>}` | 16 |
| isDarkMode | `Boolean` | 20 |

### 输出

`update:show` (:22), `submit` (:22)

### 边界和动作原式

| 编号 | 属性 | 表达式 | 行号 |
|---|---|---|---|
| UI-0293 | title | `g.value.trim() \|\| void 0` | 152 |
| UI-0294 | onUpdate:show | `n[9] \|\| (n[9] = e => a("update:show", e))` | 170 |
| UI-0295 | disabled | `r.value` | 184 |
| UI-0296 | onClick | `O` | 185 |
| UI-0297 | onUpdate:value | `n[0] \|\| (n[0] = e => g.value = e)` | 194 |
| UI-0298 | placeholder | `"留空则自动生成"` | 195 |
| UI-0299 | onUpdate:value | `n[1] \|\| (n[1] = e => i.value = e)` | 204 |
| UI-0300 | placeholder | `"粘贴参考小说全文（至少 100 字），或上传 txt / docx 自动填入"` | 207 |
| UI-0301 | disabled | `r.value` | 208 |
| UI-0302 | disabled | `r.value \|\| l.value` | 214 |
| UI-0303 | onClick | `N` | 215 |
| UI-0304 | onClick | `e => {<br>  return n = t, void s.value.splice(n, 1);<br>  var n;<br>}` | 222 |
| UI-0305 | onUpdate:value | `n[2] \|\| (n[2] = e => u.value = e)` | 228 |
| UI-0306 | min | `1e3` | 229 |
| UI-0307 | max | `5e5` | 230 |
| UI-0308 | placeholder | `"如 13000"` | 233 |
| UI-0309 | disabled | `r.value` | 234 |
| UI-0310 | disabled | `r.value` | 242 |
| UI-0311 | onClick | `t => {<br>  return n = e.value, void (u.value = n);<br>  var n;<br>}` | 243 |
| UI-0312 | onUpdate:value | `n[3] \|\| (n[3] = e => p.value = e)` | 251 |
| UI-0313 | placeholder | `"填写读者偏好，如：轻松搞笑、甜宠治愈"` | 252 |
| UI-0314 | disabled | `r.value` | 254 |
| UI-0315 | disabled | `r.value` | 259 |
| UI-0316 | onClick | `t => function (e) {<br>  const t = p.value.trim();<br>  if (!t) return void (p.value = e);<br>  t.split(/[、,，\s]+/).filter(Boolean).includes(e) \|\| (p.value = `${t}、${e}`);<br>}(e)` | 260 |
| UI-0317 | onUpdate:value | `n[4] \|\| (n[4] = e => d.value = e)` | 269 |
| UI-0318 | min | `1` | 270 |
| UI-0319 | max | `50` | 271 |
| UI-0320 | placeholder | `"1"` | 274 |
| UI-0321 | disabled | `r.value` | 275 |
| UI-0322 | onUpdate:value | `n[5] \|\| (n[5] = e => v.value = e)` | 281 |
| UI-0323 | placeholder | `"填写额外要求，如人设、情节走向、必须出现的桥段、禁忌点等"` | 284 |
| UI-0324 | disabled | `r.value` | 285 |
| UI-0325 | disabled | `r.value` | 292 |
| UI-0326 | onClick | `t => function (e) {<br>  const t = f.value.indexOf(e);<br>  t >= 0 ? f.value.splice(t, 1) : f.value.push(e);<br>}(e)` | 293 |
| UI-0327 | onUpdate:value | `n[6] \|\| (n[6] = e => c.value = e)` | 299 |
| UI-0328 | onUpdate:value | `n[7] \|\| (n[7] = e => h.value = e)` | 310 |
| UI-0329 | placeholder | `"选择备忘录作为参考素材"` | 312 |
| UI-0330 | multiple | `""` | 313 |
| UI-0331 | disabled | `r.value` | 315 |
| UI-0332 | onFocus | `y` | 316 |
| UI-0333 | onUpdate:value | `n[8] \|\| (n[8] = e => m.value = e)` | 319 |
| UI-0334 | placeholder | `"关联已有作品"` | 321 |
| UI-0335 | disabled | `r.value` | 323 |
| UI-0336 | disabled | `r.value` | 328 |
| UI-0337 | onClick | `O` | 329 |
| UI-0338 | disabled | `!C.value` | 335 |
| UI-0339 | onClick | `G` | 336 |

### 失败、空态和取消文案

- `:117` "当前环境不支持文件选择"
- `:138` "读取文件失败"
- `:144` "参考原文至少需要 100 字"
- `:207` "粘贴参考小说全文（至少 100 字），或上传 txt / docx 自动填入"
- `:306` "暂无可用文风"
- `:331` "取消"

运行状态：未实测；完整语义与剩余缺口见03/04/06/07。

## NovelBatchDialog

来源：`source/frontend/NovelBatchDialog.readable.js:1`。原bundle范围：[8841139, 8844178]。

### 输入

| 属性 | 类型/默认表达式 | 行号 |
|---|---|---|
| show | `Boolean` | 7 |
| bookId | `String` | 8 |
| sessionId | `String` | 9 |
| model | `String` | 10 |
| isDarkMode | `Boolean` | 11 |

### 输出

`update:show` (:13), `done` (:13)

### 边界和动作原式

| 编号 | 属性 | 表达式 | 行号 |
|---|---|---|---|
| UI-0340 | title | `"批量写作"` | 104 |
| UI-0341 | onUpdate:show | `p` | 110 |
| UI-0342 | onClick | `v` | 130 |
| UI-0343 | onUpdate:value | `n[0] \|\| (n[0] = e => r.value = e)` | 138 |
| UI-0344 | min | `1` | 139 |
| UI-0345 | max | `50` | 140 |
| UI-0346 | onUpdate:value | `n[1] \|\| (n[1] = e => l.value = e)` | 144 |
| UI-0347 | min | `1` | 145 |
| UI-0348 | max | `10` | 146 |
| UI-0349 | onClick | `d` | 151 |

### 失败、空态和取消文案

- `:82` "已取消"
- `:84` "连续失败过多"
- `:88` "批量写作失败"
- `:132` " 取消 "
- `:142` "连续失败停止阈值"

运行状态：未实测；完整语义与剩余缺口见03/04/06/07。

## NovelSchedulerPanel

来源：`source/frontend/NovelSchedulerPanel.readable.js:1`。原bundle范围：[8844545, 8849386]。

### 输入

| 属性 | 类型/默认表达式 | 行号 |
|---|---|---|
| bookId | `String` | 7 |
| sessionId | `String` | 8 |
| model | `String` | 9 |
| isDarkMode | `Boolean` | 10 |

### 输出

`back` (:12)

### 边界和动作原式

| 编号 | 属性 | 表达式 | 行号 |
|---|---|---|---|
| UI-0350 | onClick | `n[0] \|\| (n[0] = e => a("back"))` | 184 |
| UI-0351 | onClick | `h` | 195 |
| UI-0352 | onClick | `m` | 205 |
| UI-0353 | onUpdate:value | `n[1] \|\| (n[1] = e => i.value = e)` | 211 |
| UI-0354 | min | `1` | 212 |
| UI-0355 | max | `1440` | 213 |
| UI-0356 | onUpdate:value | `n[2] \|\| (n[2] = e => s.value = e)` | 217 |
| UI-0357 | min | `1` | 218 |
| UI-0358 | max | `100` | 219 |
| UI-0359 | onUpdate:value | `n[3] \|\| (n[3] = e => c.value = e)` | 223 |
| UI-0360 | min | `1` | 224 |
| UI-0361 | max | `10` | 225 |
| UI-0362 | onClick | `f` | 231 |

### 失败、空态和取消文案

- `:104` "启动失败"
- `:191` " 连续失败 "
- `:221` "连续失败暂停阈值"

运行状态：未实测；完整语义与剩余缺口见03/04/06/07。

## NovelHooksPanel

来源：`source/frontend/NovelHooksPanel.readable.js:1`。原bundle范围：[8849652, 8851209]。

### 输入

| 属性 | 类型/默认表达式 | 行号 |
|---|---|---|
| hooks | `{<br>  type: Array,<br>  default: () => []<br>}` | 7 |
| currentChapter | `{<br>  type: Number,<br>  default: 0<br>}` | 11 |
| isDarkMode | `Boolean` | 15 |

### 输出

没有显式emits声明；不推定有后台写入。

### 边界和动作原式

| 编号 | 属性 | 表达式 | 行号 |
|---|---|---|---|
| UI-0363 | onClick | `t => n.value = e` | 37 |

### 失败、空态和取消文案

- `:60` "暂无伏笔记录"

运行状态：未实测；完整语义与剩余缺口见03/04/06/07。

## NovelDashboard

来源：`source/frontend/NovelDashboard.readable.js:1`。原bundle范围：[8851662, 8853325]。

### 输入

| 属性 | 类型/默认表达式 | 行号 |
|---|---|---|
| chapters | `{<br>  type: Array,<br>  default: () => []<br>}` | 7 |
| isDarkMode | `Boolean` | 11 |

### 输出

没有显式emits声明；不推定有后台写入。

### 边界和动作原式

| 编号 | 属性 | 表达式 | 行号 |
|---|---|---|---|
| UI-0364 | title | ``第 ${t + 1} 章：${e} 字`` | 60 |

### 失败、空态和取消文案

- `:63` "暂无章节数据"

运行状态：未实测；完整语义与剩余缺口见03/04/06/07。

## NovelCheckpointCard

来源：`source/frontend/NovelCheckpointCard.readable.js:1`。原bundle范围：[9221037, 9221822]。

### 输入

| 属性 | 类型/默认表达式 | 行号 |
|---|---|---|
| checkpoint | `{<br>  type: Object,<br>  required: !0<br>}` | 7 |
| loading | `Boolean` | 11 |

### 输出

`confirm` (:13), `dismiss` (:13)

### 边界和动作原式

| 编号 | 属性 | 表达式 | 行号 |
|---|---|---|---|
| UI-0365 | disabled | `e.loading` | 32 |
| UI-0366 | onClick | `n[0] \|\| (n[0] = e => a("confirm"))` | 33 |
| UI-0367 | disabled | `e.loading` | 37 |
| UI-0368 | onClick | `n[1] \|\| (n[1] = e => a("dismiss"))` | 38 |

### 失败、空态和取消文案

- `:39` " 稍后 "

运行状态：未实测；完整语义与剩余缺口见03/04/06/07。

## NovelStudioPage

来源：`source/frontend/NovelStudioPage.readable.js:1`。原bundle范围：[9248264, 9294153]。

### 输入

| 属性 | 类型/默认表达式 | 行号 |
|---|---|---|
| isDarkMode | `{<br>  type: Boolean,<br>  default: !1<br>}` | 7 |
| themeMode | `{<br>  type: String,<br>  default: "light"<br>}` | 11 |

### 输出

没有显式emits声明；不推定有后台写入。

### 边界和动作原式

| 编号 | 属性 | 表达式 | 行号 |
|---|---|---|---|
| UI-0369 | title | `a \|\| K9t[r] \|\| "新会话"` | 400 |
| UI-0370 | title | `t.chapterTitle \|\| t.title` | 944 |
| UI-0371 | title | `null == (i = e.data) ? void 0 : i.title` | 1148 |
| UI-0372 | title | `"选择导出格式"` | 1333 |
| UI-0373 | onClick | `() => async function (e, t) {<br>  a.destroyAll();<br>  try {<br>    const a = await async function (e, t = "md") {<br>      var n, a;<br>      const o = "NnW0",<br>        r = "kdlK",<br>        l = "S*tx";<br>      function i(e, t, n, a) {<br>        return Tan(0, 0, n - -140, a);<br>      }<br>      const s = {};<br>      return s[i(0, 0, 161, "ZU!e") + "ss"] = !1, (null == (a = null == (n = Ran()) ? void 0 : n[i(0, 0, 25, o) + i(0, 0, -84, r) + "ort_b" + (c = -33, u = l, Gan(0, u, c - -492))]) ? void 0 : a.call(n, e, t)) ?? s;<br>      var c, u;<br>    }(e.id, t.value);<br>    if (!(null == a ? void 0 : a.success)) return void n.error((null == a ? void 0 : a.message) \|\| "导出失败");<br>    const o = new Blob([a.data.content \|\| ""], {<br>        type: `${t.mime};charset=utf-8`<br>      }),<br>      r = URL.createObjectURL(o),<br>      l = document.createElement("a");<br>    l.href = r, l.download = a.data.filename \|\| `${e.title \|\| "book"}.${t.value}`, l.click(), URL.revokeObjectURL(r), n.success("导出完成");<br>  } catch {<br>    n.error("导出失败");<br>  }<br>}(e, t)` | 1338 |
| UI-0374 | title | `"删除作品"` | 1375 |
| UI-0375 | onPositiveClick | `async () => {<br>  const t = await async function (e) {<br>    var t, n;<br>    const a = "0UX(",<br>      o = 252,<br>      r = "j^1X",<br>      l = "!jb8";<br>    function i(e, t, n, a) {<br>      return Gan(0, a, e - -443);<br>    }<br>    const s = {};<br>    return s[i(7, 0, 0, "!jb8") + "ss"] = !1, (null == (n = null == (t = Ran()) ? void 0 : t[i(110, 0, 0, a) + i(o, 0, 0, r) + "ete_b" + i(240, 0, 0, l)]) ? void 0 : n.call(t, e)) ?? s;<br>  }(e.id);<br>  (null == t ? void 0 : t.success) ? (delete c[e.id], delete u[e.id], await ye(), W.value === e.id && we(), n.success("作品已删除")) : n.error((null == t ? void 0 : t.message) \|\| "删除失败");<br>}` | 1379 |
| UI-0376 | onLaunchMode | `Re` | 1537 |
| UI-0377 | onSwitchWorkflow | `_e` | 1538 |
| UI-0378 | onCreateSession | `Pe` | 1539 |
| UI-0379 | onCreateBookSession | `n[0] \|\| (n[0] = e => Pe("book", e))` | 1540 |
| UI-0380 | onOpenSession | `Ce` | 1541 |
| UI-0381 | onOpenBook | `Te` | 1542 |
| UI-0382 | onToggleBook | `ze` | 1543 |
| UI-0383 | onRenameSession | `Pt` | 1544 |
| UI-0384 | onDeleteSession | `It` | 1545 |
| UI-0385 | onRenameBook | `Rt` | 1546 |
| UI-0386 | onDeleteBook | `Ot` | 1547 |
| UI-0387 | onExportBook | `Ct` | 1548 |
| UI-0388 | onOpenImport | `Oe` | 1549 |
| UI-0389 | onOpenStyle | `Ie` | 1550 |
| UI-0390 | onOpenImportModal | `Me` | 1551 |
| UI-0391 | onOpenImitationShort | `n[1] \|\| (n[1] = e => z.value = !0)` | 1552 |
| UI-0392 | onOpenAgentSettings | `Ne` | 1553 |
| UI-0393 | onOpenScheduler | `Ge` | 1554 |
| UI-0394 | onOpenDashboard | `qe` | 1555 |
| UI-0395 | onOpenMemo | `Ae` | 1556 |
| UI-0396 | onUpdate:selectedModel | `n[2] \|\| (n[2] = e => P.value = e)` | 1571 |
| UI-0397 | onUpdate:thinkingAutoExpand | `n[3] \|\| (n[3] = e => R.value = e)` | 1574 |
| UI-0398 | onSend | `De` | 1577 |
| UI-0399 | onClearContext | `Qe` | 1578 |
| UI-0400 | onConfirmAction | `Ke` | 1579 |
| UI-0401 | onDismissAction | `He` | 1580 |
| UI-0402 | onCancel | `Je` | 1581 |
| UI-0403 | onQuickWriteNext | `yt` | 1582 |
| UI-0404 | onQuickWriteBatch | `bt` | 1583 |
| UI-0405 | onQuickAudit | `xt` | 1584 |
| UI-0406 | onQuickRevise | `St` | 1585 |
| UI-0407 | onDeleteMessage | `wt` | 1586 |
| UI-0408 | onArchiveMessage | `_t` | 1587 |
| UI-0409 | onRetryAction | `Ze` | 1588 |
| UI-0410 | onClick | `n[4] \|\| (n[4] = e => ee.value = !1)` | 1592 |
| UI-0411 | onMousedown | `ne` | 1597 |
| UI-0412 | onRetry | `Ze` | 1608 |
| UI-0413 | onConfirm | `Ve` | 1613 |
| UI-0414 | onDismiss | `Ue` | 1614 |
| UI-0415 | onUpdate:collapsed | `n[5] \|\| (n[5] = e => ee.value = e)` | 1620 |
| UI-0416 | onSaved | `ht` | 1622 |
| UI-0417 | onUpdate:collapsed | `n[6] \|\| (n[6] = e => ee.value = e)` | 1628 |
| UI-0418 | onSaved | `gt` | 1630 |
| UI-0419 | onUpdate:selectedStyleIds | `n[7] \|\| (n[7] = e => M.value = e)` | 1635 |
| UI-0420 | onBack | `ge` | 1637 |
| UI-0421 | onRefresh | `be` | 1638 |
| UI-0422 | onApplied | `We` | 1639 |
| UI-0423 | onBack | `ge` | 1644 |
| UI-0424 | onClick | `ge` | 1649 |
| UI-0425 | onClick | `ge` | 1659 |
| UI-0426 | onUpdate:show | `n[8] \|\| (n[8] = e => T.value = e)` | 1677 |
| UI-0427 | onUpdate:tab | `n[9] \|\| (n[9] = e => A.value = e)` | 1679 |
| UI-0428 | onCreated | `Xe` | 1684 |
| UI-0429 | onSubmitStart | `Fe` | 1685 |
| UI-0430 | onUpdate:show | `n[10] \|\| (n[10] = e => z.value = e)` | 1688 |
| UI-0431 | onSubmit | `ut` | 1693 |
| UI-0432 | onUpdate:show | `n[12] \|\| (n[12] = e => I.value = e)` | 1696 |
| UI-0433 | title | `"book" === q.value ? "重命名作品" : "重命名会话"` | 1698 |
| UI-0434 | onPositiveClick | `Mt` | 1700 |
| UI-0435 | onUpdate:value | `n[11] \|\| (n[11] = e => N.value = e)` | 1704 |
| UI-0436 | placeholder | `"book" === q.value ? "作品标题" : "会话标题"` | 1705 |
| UI-0437 | onUpdate:show | `n[13] \|\| (n[13] = e => oe.value = e)` | 1710 |
| UI-0438 | onDone | `kt` | 1715 |
| UI-0439 | onClick | `n[15] \|\| (n[15] = Zi(e => re.value = !1, ["self"]))` | 1724 |
| UI-0440 | onUpdate:visible | `n[14] \|\| (n[14] = e => re.value = e)` | 1733 |

### 失败、空态和取消文案

- `:304` "加载会话失败"
- `:402` "创建会话失败"
- `:488` "发送失败"
- `:491` "发送失败"
- `:565` "创建失败"
- `:567` "创建失败"
- `:655` "继续失败"
- `:704` "执行失败"
- `:705` "执行失败"
- `:708` "已打开导入素材，请补充材料"
- `:718` "内容已生成，请查看右侧文件并确认后继续"
- `:722` "审稿未通过，请查看卡片中的问题列表"
- `:877` "无法创建会话，请重试"
- `:913` "仿写短篇失败"
- `:914` "仿写短篇失败"
- `:915` "仿写短篇失败"
- `:921` "创作蓝图已生成，请查看右侧文件并确认后继续"
- `:928` "仿写短篇失败"
- `:929` "仿写短篇失败"
- `:930` "仿写短篇失败"
- `:1137` "写章失败"
- `:1210` "审稿失败"
- `:1226` "审稿未通过，请查看卡片中的问题列表"
- `:1242` "删除失败"
- `:1297` "修订失败"
- `:1353` "导出失败"
- `:1361` "导出失败"
- `:1378` "取消"
- `:1392` "删除失败"
- `:1415` "重命名失败"
- `:1436` "重命名失败"
- `:1459` "删除失败"
- `:1631` " 任务执行中，AI 输出请查看左侧对话框 "

运行状态：未实测；完整语义与剩余缺口见03/04/06/07。
