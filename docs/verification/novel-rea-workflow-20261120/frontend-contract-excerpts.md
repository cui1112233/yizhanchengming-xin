# 当前原版前端静态摘录
源bundle SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6。行号是恢复/格式化文件行号；不是原作者源码。静态事件存在不证明运行已成功。

## NovelAgentSettings.readable.js
恢复文件SHA256: e7e5b76ee59a26a294e55d98301beea0a95c5dd48f1a1834d5524f8094b78ad3
```text
6:   props: {
7:     modelOptions: {
8:       type: Array,
9:       default: () => []
10:     },
11:     isDarkMode: Boolean
12:   },
13:   emits: ["back"],
14:   setup(e, {
15:     emit: t
16:   }) {
17:     const n = e,
18:       a = t,
19:       o = $q(),
20:       r = Nt(!0),
128:             r = "VQi*",
129:             l = 193,
130:             i = "$OLu",
131:             s = 554,
132:             c = {};
133:           function u(e, t, n, a) {
134:             return Gan(0, a, n - s);
135:           }
136:           return c.success = !1, (null == (n = null == (t = Ran()) ? void 0 : t[u(0, 0, a, "$OLu") + u(0, 0, o, r) + "e_age" + (d = l, v = i, Gan(0, v, d - -229) + "ttings")]) ? void 0 : n.call(t, e)) ?? c;
137:           var d, v;
138:         }({
139:           modelRoutes: {
140:             ...s.value
141:           },
142:           reviewPrompts: c.value.map(y),
143:           customWritePrompts: u.value.map(e => ({
144:             id: e.id,
145:             category: e.category,
146:             title: e.title,
147:             content: e.content,
148:             enabled: Boolean(e.enabled)
149:           })),
150:           reviewCycle: {
151:             maxRetries: Math.max(0, Math.min(5, Number(d.value.maxRetries) || 3)),
152:             passScore: Math.max(0, Math.min(100, Number(d.value.passScore) || 80)),
153:             temperatureStep: Math.max(0, Math.min(.5, Number(d.value.temperatureStep) || .15)),
154:             baseTemperature: Math.max(0, Math.min(2, Number(d.value.baseTemperature) || .65)),
155:             enableLengthNorm: Boolean(d.value.enableLengthNorm),
156:             enablePostWrite: Boolean(d.value.enablePostWrite),
157:             enableAiTells: Boolean(d.value.enableAiTells),
158:             requireSecondaryConfirm: Boolean(d.value.requireSecondaryConfirm)
159:           }
160:         });
161:         if (!(null == r ? void 0 : r.success)) return void o.error((null == r ? void 0 : r.message) || "保存失败");
162:         s.value = {
163:           ...((null == (e = r.data) ? void 0 : e.modelRoutes) || {})
164:         }, c.value = ((null == (t = r.data) ? void 0 : t.reviewPrompts) || []).map(e => ({
165:           ...e
166:         })), u.value = ((null == (n = r.data) ? void 0 : n.customWritePrompts) || []).map(e => ({
167:           ...e
168:         }));
169:         const l = (null == (a = r.data) ? void 0 : a.reviewCycle) || {};
170:         d.value = {
171:           maxRetries: Number(l.maxRetries ?? 3),
172:           passScore: Number(l.passScore ?? 80),
173:           temperatureStep: Number(l.temperatureStep ?? .15),
174:           baseTemperature: Number(l.baseTemperature ?? .65),
175:           enableLengthNorm: !1 !== l.enableLengthNorm,
176:           enablePostWrite: !1 !== l.enablePostWrite,
177:           enableAiTells: !1 !== l.enableAiTells,
178:           requireSecondaryConfirm: Boolean(l.requireSecondaryConfirm)
179:         }, o.success("已保存");
180:       } finally {
244:           function l(e, t, n, a) {
245:             return Gan(0, t, e - -706);
246:           }
247:           return r[l(-292, "&2iX") + "ss"] = !1, (null == (t = null == (e = Ran()) ? void 0 : e[l(-392, n) + l(-97, "MEJ&") + "_agen" + (i = a, s = 18, Tan(0, 0, s - -64, i)) + l(-290, o)]) ? void 0 : t.call(e)) ?? r;
248:           var i, s;
249:         }();
250:         if (!(null == p ? void 0 : p.success)) return void o.error((null == p ? void 0 : p.message) || "加载 Agent 设置失败");
251:         s.value = {
252:           ...((null == (t = p.data) ? void 0 : t.modelRoutes) || {})
253:         }, c.value = ((null == (n = p.data) ? void 0 : n.reviewPrompts) || []).map(e => ({
254:           ...e
255:         })), u.value = ((null == (a = p.data) ? void 0 : a.customWritePrompts) || []).map(e => ({
256:           ...e
257:         }));
258:         const f = (null == (l = p.data) ? void 0 : l.reviewCycle) || {};
259:         d.value = {
260:           maxRetries: Number(f.maxRetries ?? 3),
261:           passScore: Number(f.passScore ?? 80),
262:           temperatureStep: Number(f.temperatureStep ?? .15),
263:           baseTemperature: Number(f.baseTemperature ?? .65),
264:           enableLengthNorm: !1 !== f.enableLengthNorm,
265:           enablePostWrite: !1 !== f.enablePostWrite,
266:           enableAiTells: !1 !== f.enableAiTells,
267:           requireSecondaryConfirm: Boolean(f.requireSecondaryConfirm)
268:         };
269:       } finally {
270:         r.value = !1;
465:       }, null, 8, ["value", "placeholder", "onUpdate:value"])]))), 128))]))])]),
466:       _: 1
467:     }), zr("footer", Iln, [n[37] || (n[37] = zr("span", {
468:       class: "footer-hint"
469:     }, "设置保存在本地数据目录，永久生效。", -1)), Qr(Tt(WS), {
470:       type: "primary",
471:       loading: l.value,
```

## NovelStudioPage.readable.js
恢复文件SHA256: b789cdac45ec085fdb353a0ceadcaa94d54756069f4c2ba169a8f6a97fc2e6d3
```text
404:     function Re(e) {
405:       _e(e);
406:     }
407:     function Oe(e) {
408:       _e(e), A.value = e, T.value = !0, g.value = "import";
409:     }
410:     function Me(e = "imitation") {
411:       ge(), A.value = e, T.value = !0, g.value = "import";
412:     }
413:     function Ie() {
414:       h.value = "style", g.value = "style";
415:     }
416:     function Ne() {
417:       h.value = "agent-settings", g.value = "agent-settings";
418:     }
419:     function Ge() {
420:       h.value = "scheduler", g.value = "scheduler";
421:     }
422:     function qe() {
423:       h.value = "dashboard", g.value = "dashboard";
424:     }
425:     function Ae() {
426:       le.value = !0, re.value = !0;
427:     }
428:     async function Te(e) {
429:       var t;
430:       u[e] = !0;
865:         bookTitle: e,
866:         status: "running",
867:         startedAt: Date.now(),
868:         stages: $.map(e => ({
869:           name: e,
870:           status: "pending"
871:         })),
872:         logs: []
873:       }, S.value = b.value.stages;
874:     }
875:     async function ut(e) {
876:       var t;
877:       if (z.value = !1, _e("imitation-short"), d.value && "imitation-short" === J9t(v.value) || (await Pe("imitation-short", null, e.title || "仿写短篇"), await vn()), !d.value) return void n.error("无法创建会话，请重试");
878:       const a = e.title || "仿写短篇";
879:       ct(a), ge(), ee.value = !1, await et([Ye("user", `开始仿写短篇：${a}`)]), n.info("仿写短篇已开始，可在右侧查看进度");
880:       let o = !1;
881:       try {
882:         const a = await async function (e) {
883:           var t, n;
884:           const a = "$OLu",
885:             o = "rlaI",
886:             r = "oP9p",
887:             l = 1150,
888:             i = "ZU!e",
889:             s = "VDbS",
890:             c = 1316,
891:             u = 791;
892:           function d(e, t, n, a) {
893:             return Gan(0, t, e - -496);
894:           }
895:           const v = {};
896:           function p(e, t, n, a) {
897:             return Gan(0, t, n - u);
898:           }
899:           return v[p(0, "WFL$", 1538) + "ss"] = !1, (null == (n = null == (t = {
900:             EFVGf: function (e) {
901:               return e();
902:             }
903:           }[d(-104, a)](Ran)) ? void 0 : t[d(-279, "4M#3") + d(188, o) + p(0, r, l) + d(-70, i) + p(0, s, c) + "t"]) ? void 0 : n.call(t, e)) ?? v;
904:         }({
905:           ...e,
1635:         "onUpdate:selectedStyleIds": n[7] || (n[7] = e => M.value = e),
1636:         "is-dark-mode": e.isDarkMode,
1637:         onBack: ge,
1638:         onRefresh: be,
1639:         onApplied: We
1640:       }, null, 8, ["style-sources", "selected-style-ids", "is-dark-mode"])) : "agent-settings" === h.value ? (Cr(), Nr(Gln, {
1641:         key: 2,
1642:         "model-options": C.value,
1643:         "is-dark-mode": e.isDarkMode,
1644:         onBack: ge
1645:       }, null, 8, ["model-options", "is-dark-mode"])) : "scheduler" === h.value ? (Cr(), Ir("div", eyn, [zr("div", {
1646:         class: "panel-page-header"
1647:       }, [zr("button", {
1648:         class: "back-btn",
1649:         onClick: ge
1650:       }, "← 返回"), n[16] || (n[16] = zr("span", {
```

## NovelImitationShortDialog.readable.js
恢复文件SHA256: 6b0d791550053430d284b6be4f31e3f5554eba0e789b892e342da3b3f8175b78
```text
88:       const a = null == (t = window.pywebview) ? void 0 : t.api;
89:       if (null == a ? void 0 : a.read_file_content) {
90:         const e = await a.read_file_content(n);
91:         return (null == e ? void 0 : e.success) && e.content ? e.content : ((null == e ? void 0 : e.message) && o.warning(e.message), "");
92:       }
93:       if (null == a ? void 0 : a.read_file) {
94:         return (await a.read_file(n)) || "";
95:       }
96:       return "";
97:     }
98:     function R(e, t) {
99:       const n = String(t || "").trim();
100:       return !!n && (i.value.trim() ? i.value = `${i.value.trim()}\n\n--- ${e} ---\n\n${n}` : i.value = n, !0);
101:     }
102:     function O() {
103:       a("update:show", !1);
104:     }
105:     function M() {
106:       const e = Number(u.value);
107:       return !Number.isFinite(e) || e < 1e3 ? 13e3 : Math.min(5e5, Math.round(e));
108:     }
109:     function I() {
110:       const e = Number(d.value);
111:       return !Number.isFinite(e) || e < 1 ? 1 : Math.min(50, Math.round(e));
112:     }
113:     async function N() {
114:       var e, t;
115:       if (!l.value && !r.value) try {
116:         const n = null == (e = window.pywebview) ? void 0 : e.api;
117:         if (!(null == n ? void 0 : n.novels_pick_style_files)) return void o.warning("当前环境不支持文件选择");
118:         l.value = !0;
119:         const a = await n.novels_pick_style_files();
120:         if (!(null == a ? void 0 : a.success)) return void ((null == a ? void 0 : a.message) && o.error(a.message));
121:         if (!(null == (t = a.data) ? void 0 : t.length)) return;
122:         let r = 0;
123:         for (const e of a.data) {
124:           const t = e.path || e,
125:             n = e.fileName || e.name || ("string" == typeof t ? t.split(/[/\\]/).pop() : "文件");
126:           if (s.value.some(e => e.path === t)) {
127:             o.info(`「${n}」已导入，跳过重复文件`);
128:             continue;
129:           }
130:           R(n, await P(e)) ? (s.value.push({
131:             name: n,
132:             path: t,
133:             size: e.size || 0
134:           }), r += 1) : o.warning(`未能读取文件内容：${n}`);
135:         }
136:         r > 0 && o.success(`已读取 ${r} 个文件并填入参考原文`);
137:       } catch (Ofr) {
138:         o.error("读取文件失败");
139:       } finally {
140:         l.value = !1;
141:       }
142:     }
143:     async function G() {
144:       if (!C.value) return void o.warning("参考原文至少需要 100 字");
145:       let e = i.value.trim();
146:       if (h.value.length) for (const n of h.value) {
147:         const t = W.value.find(e => e.id === n);
148:         (null == t ? void 0 : t.plain) && (e += `\n\n--- 备忘录: ${t.title} ---\n\n${t.plain}`);
149:       }
150:       const t = {
151:         referenceText: e,
152:         title: g.value.trim() || void 0,
153:         targetWordCount: M(),
154:         targetChapters: I(),
155:         chapters: I(),
156:         supplementaryInfo: v.value.trim() || void 0,
157:         readerPreferences: p.value || void 0,
158:         inspirationTags: f.value.join("、") || void 0,
159:         styleName: p.value || void 0,
160:         linkedSaveId: m.value || void 0,
161:         knowledgeBookIds: c.value.length ? c.value : void 0
162:       };
163:       a("submit", t), i.value = "", s.value = [], c.value = [], u.value = 13e3, d.value = 1, v.value = "", p.value = "", f.value = [], m.value = null, h.value = [], g.value = "", O();
164:     }
```

## NovelStudioSidebar.readable.js
恢复文件SHA256: 9c18f2fed68af50d9b513fcd9c9370f96b6b361ad7b849a492e28d8a5ff418b0
```text
4: const recoveredComponent = {
5:   __name: "NovelStudioSidebar",
6:   props: {
7:     books: {
8:       type: Array,
9:       default: () => []
10:     },
11:     projectSessions: {
12:       type: Array,
13:       default: () => []
14:     },
15:     modeSessions: {
16:       type: Array,
17:       default: () => []
18:     },
19:     sessionProgress: {
20:       type: Object,
21:       default: () => ({})
22:     },
23:     activeWorkflow: {
24:       type: String,
25:       default: "book-create"
26:     },
27:     activeSessionId: String,
28:     activeBookId: String,
29:     activeMode: String,
30:     activeTool: String,
31:     expandedBooks: {
32:       type: Object,
33:       default: () => ({})
34:     },
35:     bookSessions: {
36:       type: Object,
37:       default: () => ({})
38:     },
39:     isDarkMode: Boolean
40:   },
41:   emits: ["create-session", "open-session", "open-book", "toggle-book", "create-book-session", "rename-session", "delete-session", "rename-book", "delete-book", "export-book", "launch-mode", "switch-workflow", "open-import", "open-style", "open-import-modal", "open-imitation-short", "open-agent-settings", "open-scheduler", "open-dashboard", "open-memo"],
42:   setup(e, {
43:     emit: t
44:   }) {
45:     const n = e,
46:       a = t,
47:       o = Nt(!0),
48:       r = Nt(!0);
49:     Nt(!1);
50:     const l = Nt(localStorage.getItem("novel_book_filter") || "all"),
```
