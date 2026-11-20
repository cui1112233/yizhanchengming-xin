// Inspection only. Original SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Original UTF16 offsets [9248264, 9294153); original line 56
// External dependencies and runtime equivalence remain unresolved.
const recoveredComponent = {
  __name: "NovelStudioPage",
  props: {
    isDarkMode: {
      type: Boolean,
      default: !1
    },
    themeMode: {
      type: String,
      default: "light"
    }
  },
  setup(e) {
    const t = e,
      n = $q(),
      a = tq(),
      o = function () {
        const e = wo(_p, null);
        return ml(() => {
          if (null === e) return Lk;
          const {
              mergedThemeRef: {
                value: t
              },
              mergedThemeOverridesRef: {
                value: n
              }
            } = e,
            a = (null == t ? void 0 : t.common) || Lk;
          return (null == n ? void 0 : n.common) ? Object.assign({}, a, n.common) : a;
        });
      }();
    function r(e, t) {
      return `rgba(${parseInt(e.slice(1, 3), 16)},${parseInt(e.slice(3, 5), 16)},${parseInt(e.slice(5, 7), 16)},${t})`;
    }
    const l = ml(() => {
        const e = o.value,
          n = t.isDarkMode,
          a = e.primaryColor;
        return {
          "--ns-bg": e.bodyColor,
          "--ns-sidebar-bg": e.cardColor,
          "--ns-text": e.textColor1,
          "--ns-text-2": e.textColor2,
          "--ns-text-3": e.textColor3,
          "--ns-border": e.dividerColor,
          "--ns-hover": n ? "rgba(255,255,255,0.05)" : "rgba(0,0,0,0.04)",
          "--ns-active": n ? "rgba(255,255,255,0.1)" : "rgba(0,0,0,0.07)",
          "--ns-primary": a,
          "--ns-primary-bg": r(a, .08),
          "--ns-primary-bg-hover": r(a, .14),
          "--ns-primary-border": r(a, .4),
          "--ns-panel-bg": n ? e.cardColor : e.bodyColor,
          "--ns-popup-bg": e.popoverColor,
          "--ns-popup-border": e.borderColor,
          "--ns-shadow": n ? "0 12px 32px rgba(0,0,0,.3)" : "0 12px 32px rgba(0,0,0,.12)",
          "--ns-resizer-hover": n ? "rgba(255,255,255,0.08)" : "rgba(0,0,0,0.06)",
          "--ns-btn-text": n ? "rgba(255,255,255,0.72)" : "rgba(0,0,0,0.62)",
          "--ns-btn-text-hover": n ? "#f2f2f2" : "rgba(0,0,0,0.86)",
          "--ns-btn-bg": n ? e.cardColor : e.bodyColor
        };
      }),
      i = Nt([]),
      s = Nt([]),
      c = Wt({}),
      u = Wt({}),
      d = Nt(""),
      v = Nt(null),
      p = Nt(null),
      f = Nt(null),
      m = Nt(localStorage.getItem("novel_studio_active_mode") || "book-create"),
      h = Nt("chat"),
      g = Nt("chat"),
      W = ml(() => {
        var e, t;
        return (null == (e = v.value) ? void 0 : e.bookId) || L.value || (null == (t = p.value) ? void 0 : t.id) || "";
      }),
      y = ml(() => {
        var e, t;
        const n = (null == (e = v.value) ? void 0 : e.shortStoryId) || "";
        return n || (D.value && (null == (t = b.value) ? void 0 : t.sessionId) === d.value ? D.value : "");
      }),
      b = Nt(null),
      k = Nt(null);
    const x = Nt(!1),
      w = Nt(!1),
      _ = Nt(null),
      S = Nt([]),
      C = Nt([]),
      P = Nt(""),
      R = Nt("0" !== localStorage.getItem("novel_studio_thinking_auto_expand")),
      O = Nt([]),
      M = Nt(JSON.parse(localStorage.getItem("novel_studio_selected_style_ids") || "[]")),
      I = Nt(!1),
      N = Nt(""),
      G = Nt(null),
      q = Nt("session"),
      A = Nt("imitation"),
      T = Nt(!1),
      z = Nt(!1),
      Q = Nt(!1),
      D = Nt(""),
      L = Nt("");
    let j = null,
      F = null;
    const B = Wt({});
    function V(e) {
      return e ? (B[e] || (B[e] = {
        isStreaming: !1,
        pendingChat: !1,
        streamingText: "",
        streamingReasoning: "",
        pipelineStreamStage: ""
      }), B[e]) : {
        isStreaming: !1,
        pendingChat: !1,
        streamingText: "",
        streamingReasoning: "",
        pipelineStreamStage: ""
      };
    }
    const U = ml(() => V(d.value)),
      H = ml(() => {
        const e = U.value;
        return e.isStreaming || e.pendingChat;
      }),
      J = ml(() => U.value.streamingText),
      Z = ml(() => U.value.streamingReasoning),
      X = ml(() => U.value.pipelineStreamStage),
      $ = ["拆书分析", "生成仿写方向", "构建创作蓝图", "生成短篇大纲", "审阅大纲", "修订大纲", "撰写正文", "审阅正文", "修订正文", "编辑审核", "打包输出"],
      ee = Nt("1" === localStorage.getItem("novel_book_panel_collapsed"));
    const te = Nt(function () {
      const e = Number(localStorage.getItem("novel_right_panel_width"));
      return e >= 360 && e <= 800 ? e : "undefined" != typeof window ? Math.min(800, Math.max(360, Math.round(.38 * window.innerWidth))) : 560;
    }());
    function ne(e) {
      e.preventDefault();
      const t = e.clientX,
        n = te.value,
        a = e => {
          const a = t - e.clientX;
          te.value = Math.min(800, Math.max(360, n + a));
        },
        o = () => {
          localStorage.setItem("novel_right_panel_width", String(te.value)), document.removeEventListener("mousemove", a), document.removeEventListener("mouseup", o), document.body.style.cursor = "", document.body.style.userSelect = "";
        };
      document.body.style.cursor = "col-resize", document.body.style.userSelect = "none", document.addEventListener("mousemove", a), document.addEventListener("mouseup", o);
    }
    const ae = Nt(!1),
      oe = Nt(!1),
      re = Nt(!1),
      le = Nt(!1),
      ie = Nt(null),
      se = ml(() => {
        var e, t, n, a, o;
        const r = (null == (e = v.value) ? void 0 : e.bookId) || L.value;
        if (!r || "chat" !== h.value || !p.value) return !1;
        if (p.value.id !== r) return !1;
        const l = (null == (t = b.value) ? void 0 : t.sessionId) === d.value && "architect" === (null == (n = b.value) ? void 0 : n.toolName) && "running" === (null == (a = b.value) ? void 0 : a.status);
        return "active" === p.value.status || "incubating" === p.value.status || (p.value.chaptersWritten ?? 0) > 0 || ((null == (o = p.value.foundationFiles) ? void 0 : o.length) ?? 0) > 0 || l;
      }),
      ce = ml(() => {
        var e, t;
        if (!f.value) return !1;
        if ("chat" !== h.value || "short" !== (null == (e = v.value) ? void 0 : e.kind)) return !1;
        const n = (null == (t = v.value) ? void 0 : t.shortStoryId) || "",
          a = D.value || "",
          o = n || a;
        return !o || f.value.id === o;
      }),
      ue = ml(() => se.value || ce.value),
      de = ml(() => {
        const e = he.value;
        return e ? "running" !== e.status && "error" !== e.status ? null : Ntn(e) : null;
      }),
      ve = ml(() => ue.value || Boolean(de.value) || Boolean(_.value)),
      pe = ml(() => !de.value || se.value || ce.value ? se.value ? "书籍资料" : "短篇资料" : "任务进度"),
      fe = ml(() => s.value.filter(e => J9t(e) === m.value)),
      me = ml(() => {
        var e, t;
        const n = {};
        for (const [r, l] of Object.entries(B)) (l.isStreaming || l.pendingChat) && (n[r] = {
          running: !0,
          percent: null
        });
        const a = d.value;
        if (!a) return n;
        const o = (null == (e = b.value) ? void 0 : e.sessionId) === a ? b.value : null;
        if ((x.value || ae.value || o && "running" === o.status) && !n[a]) {
          let e = null;
          if (null == (t = null == o ? void 0 : o.stages) ? void 0 : t.length) {
            const t = o.stages,
              n = t.filter(e => "completed" === e.status).length,
              a = t.some(e => "active" === e.status);
            e = Math.round(n / t.length * 100), a && e < 100 && (e = Math.min(99, Math.round((n + .4) / t.length * 100)));
          }
          n[a] = {
            running: !0,
            percent: e
          };
        }
        return n;
      });
    Ro(m, e => {
      localStorage.setItem("novel_studio_active_mode", e);
    });
    const he = ml(() => {
      const e = b.value;
      return e ? e.sessionId && e.sessionId !== d.value ? null : e : null;
    });
    function ge() {
      h.value = "chat", g.value = "chat";
    }
    function We(e) {
      const t = Array.isArray(e) ? e.length : 0;
      if (t > 0) {
        const a = (O.value || []).filter(t => e.includes(t.id)).map(e => e.title).filter(Boolean);
        n.success(a.length ? `已应用 ${t} 个写作风格：${a.join("、")}` : `已应用 ${t} 个写作风格`);
      } else n.info("已清除写作风格参考");
    }
    async function ye() {
      const e = await async function () {
        var e, t;
        const n = 1136,
          a = 1187,
          o = "VDbS",
          r = 1191,
          l = 1196,
          i = 1402,
          s = "Tv^A",
          c = 261,
          u = 26,
          d = 90,
          v = "c3iw",
          p = 1008;
        function f(e, t, n, a) {
          return Tan(0, 0, n - p, a);
        }
        function m(e, t, n, a) {
          return Gan(0, n, t - -324);
        }
        const h = {
          IEHmU: function (e) {
            return e();
          },
          aylAS: m(601, 382, "XGTO") + "可用"
        };
        return (null == (t = null == (e = h[f(n, 1234, a, o)](Ran)) ? void 0 : e["novel" + f(r, l, i, s) + m(c, u, "q0!*") + "ks"]) ? void 0 : t.call(e)) ?? {
          success: !1,
          message: h[m(d, -33, v)]
        };
      }();
      (null == e ? void 0 : e.success) && (i.value = e.data || []);
    }
    async function be() {
      const e = await async function () {
        var e, t;
        const n = 1539,
          a = "Tv^A",
          o = "OCei",
          r = 1496,
          l = "0UX(",
          i = 19,
          s = {};
        function c(e, t, n, a) {
          return Tan(0, 0, a - 1142, e);
        }
        function u(e, t, n, a) {
          return Gan(0, n, e - i);
        }
        return s.success = !1, s[c("c3iw", 0, 0, n)] = [], (null == (t = null == (e = Ran()) ? void 0 : e["novel" + u(621, 0, a) + "t_sty" + c(o, 0, 0, r) + u(430, 0, l)]) ? void 0 : t.call(e)) ?? s;
      }();
      (null == e ? void 0 : e.success) && (O.value = e.data || []);
    }
    async function ke() {
      const e = await Ian(null);
      (null == e ? void 0 : e.success) && (s.value = e.data || []);
    }
    async function xe(e) {
      const t = await Ian(e);
      (null == t ? void 0 : t.success) && (c[e] = t.data || []);
    }
    function we() {
      d.value = "", v.value = null, p.value = null, f.value = null, k.value = null, b.value = null, S.value = [], at(), ot(), Q.value = !1, D.value = "", L.value = "";
    }
    function _e(e) {
      m.value = e, ge();
      const t = v.value ? J9t(v.value) : "";
      t && t === e || we();
    }
    async function Se(e) {
      var t, n, a;
      if (!e) return null;
      const o = await Nan(e);
      return (null == o ? void 0 : o.success) ? (await ke(), e === d.value && (v.value = o.data, k.value = (null == (t = o.data) ? void 0 : t.pendingAction) || null, (null == (n = o.data) ? void 0 : n.shortStory) ? f.value = o.data.shortStory : (null == (a = o.data) ? void 0 : a.shortStoryId) && (await gt())), o.data) : null;
    }
    async function Ce(e) {
      var t, a, o, r, l, i, s, c, u;
      at(), ot(), Q.value = !1;
      const h = await Nan(e);
      if (!(null == h ? void 0 : h.success)) return void n.error((null == h ? void 0 : h.message) || "加载会话失败");
      d.value = e, v.value = h.data, m.value = J9t(h.data), k.value = (null == (t = h.data) ? void 0 : t.pendingAction) || null, S.value = [], (null == (a = h.data) ? void 0 : a.book) ? (p.value = h.data.book, L.value = h.data.book.id || (null == (o = h.data) ? void 0 : o.bookId) || "") : (null == (r = h.data) ? void 0 : r.bookId) ? (L.value = h.data.bookId, await ht()) : (p.value = null, L.value = ""), (null == (l = h.data) ? void 0 : l.shortStory) ? (f.value = h.data.shortStory, D.value = h.data.shortStory.id || (null == (i = h.data) ? void 0 : i.shortStoryId) || "") : (null == (s = h.data) ? void 0 : s.shortStoryId) ? (D.value = h.data.shortStoryId, await gt()) : (f.value = null, D.value = "");
      const g = function (e) {
        try {
          const t = sessionStorage.getItem("novel_tool_execution");
          if (!t) return null;
          const n = JSON.parse(t);
          if ((null == n ? void 0 : n.sessionId) === e) return n;
        } catch {}
        return null;
      }(e);
      !g || "running" !== g.status && "error" !== g.status ? b.value = null : (b.value = g, "imitation_short_run" === g.toolName && "running" === g.status && (Q.value = !0), "running" === g.status && ("imitation_short_run" !== g.toolName && "short_fiction_run" !== g.toolName || !(null == (c = h.data) ? void 0 : c.shortStoryId) || lt(h.data.shortStoryId), "architect" === g.toolName && (null == (u = h.data) ? void 0 : u.bookId) && st(h.data.bookId))), ge();
    }
    async function Pe(e = "book-create", t = null, a = "") {
      let o = e,
        r = e;
      t ? (o = "book", r = "book") : ["book-create", "short", "imitation-short", "fanfic", "spinoff", "imitation", "chapters"].includes(e) ? (o = function (e) {
        const t = 1531,
          n = "OrOs",
          a = 1020,
          o = "Dqz^",
          r = 1047,
          l = 1511,
          i = "v(BQ",
          s = 1312,
          c = "knG)",
          u = "Dqz^",
          d = 1530,
          v = 1474,
          p = "Qf1U",
          f = 1395,
          m = 1526,
          h = 1553,
          g = "7l1T",
          W = 1391,
          y = 1684,
          b = 1359,
          k = "knG)",
          x = 1590,
          w = 1639,
          _ = "a7rx",
          S = 1508,
          C = 772,
          P = "3Fep",
          R = "&Jot",
          O = 766,
          M = 1442,
          I = "vqft",
          N = 1357,
          G = "GqS4",
          q = 689,
          A = 1174,
          T = 1342,
          z = 1634,
          Q = 1498,
          D = 742,
          L = 936,
          j = {};
        function E(e, t, n, a) {
          return len(0, n, 0, a - L);
        }
        function F(e, t, n, a) {
          return U9t(0, 0, n - 76, t);
        }
        j[E(1445, 0, "$f@l", 1495)] = function (e, t) {
          return e === t;
        }, j.uLcMQ = E(t, 0, n, 1401), j[F(a, o, r)] = function (e, t) {
          return e === t;
        }, j[E(l, 0, i, 1424)] = E(s, 0, c, 1423) + E(1544, 0, u, d) + E(v, 0, p, f), j[E(m, 0, n, 1342)] = function (e, t) {
          return e === t;
        }, j[E(h, 0, g, W)] = E(y, 0, "C$d9", 1543), j.OqKXn = E(b, 0, k, x) + E(w, 0, _, S) + "e";
        const B = j;
        return B[F(0, "y$*N", C)](e, B[F(0, P, 737)]) || B[F(0, R, O)](e, B[E(M, 0, I, N)]) ? B[F(0, G, q)] : B[E(A, 0, n, T)](e, B[F(0, R, 702)]) ? E(z, 0, "v(BQ", Q) : B[F(0, G, D)];
      }(e), r = e) : (r = m.value || "book-create", o = e);
      const l = await async function (e = {}) {
        var t, n;
        const a = "XGTO",
          o = 778,
          r = "#WGY",
          l = "iUeH";
        function i(e, t, n, a) {
          return Gan(0, a, e - 111);
        }
        const s = {};
        function c(e, t, n, a) {
          return Gan(0, e, t - -814);
        }
        return s.success = !1, (null == (n = null == (t = {
          FGkdI: function (e) {
            return e();
          }
        }[c(a, -128)](Ran)) ? void 0 : t[i(653, 0, 0, "XGTO") + i(o, 0, 0, r) + i(649, 0, 0, "Z8Hk") + c(l, -587) + "n"]) ? void 0 : n.call(t, e)) ?? s;
      }({
        kind: o,
        bookId: t,
        workflow: r,
        title: a || K9t[r] || "新会话"
      });
      (null == l ? void 0 : l.success) ? (t || (m.value = r), await ke(), t && (await xe(t)), await Ce(l.data.id)) : n.error((null == l ? void 0 : l.message) || "创建会话失败");
    }
    function Re(e) {
      _e(e);
    }
    function Oe(e) {
      _e(e), A.value = e, T.value = !0, g.value = "import";
    }
    function Me(e = "imitation") {
      ge(), A.value = e, T.value = !0, g.value = "import";
    }
    function Ie() {
      h.value = "style", g.value = "style";
    }
    function Ne() {
      h.value = "agent-settings", g.value = "agent-settings";
    }
    function Ge() {
      h.value = "scheduler", g.value = "scheduler";
    }
    function qe() {
      h.value = "dashboard", g.value = "dashboard";
    }
    function Ae() {
      le.value = !0, re.value = !0;
    }
    async function Te(e) {
      var t;
      u[e] = !0;
      const n = await Oan(e);
      (null == n ? void 0 : n.success) && (p.value = n.data, ee.value = !1), c[e] || (await xe(e));
      const a = c[e] || [];
      ge(), a.length ? await Ce(a[0].id) : await Pe("book", e, (null == (t = p.value) ? void 0 : t.title) || "书籍会话");
    }
    function ze(e) {
      u[e] = !u[e], u[e] && !c[e] && xe(e);
    }
    async function Qe() {
      d.value && v.value && (v.value.messages = [], await Ce(d.value));
    }
    async function De(e, t) {
      var a, o;
      d.value || (await Pe(m.value));
      const r = d.value;
      if (!r) return;
      const l = V(r);
      l.isStreaming = !0, l.pendingChat = !0, l.streamingText = "", l.streamingReasoning = "", l.pipelineStreamStage = "", b.value = null, v.value && r === d.value && (v.value = {
        ...v.value,
        messages: [...(v.value.messages || []), Ye("user", e)]
      });
      const i = Qre(P.value),
        s = {
          sessionId: r,
          message: e,
          model: P.value || void 0,
          knowledgeBookIds: M.value,
          thinkingMode: !0,
          codexReasoningEffort: i,
          claudeReasoningEffort: i
        };
      null != (null == t ? void 0 : t.contextLength) && (s.contextLength = t.contextLength);
      try {
        const e = await async function (e) {
          var t, n;
          const a = 479,
            o = "2CEX",
            r = 555,
            l = 644,
            i = "hjoV",
            s = 948,
            c = "6mzS",
            u = 225,
            d = 115;
          function v(e, t, n, a) {
            return Gan(0, e, a - d);
          }
          const p = {};
          function f(e, t, n, a) {
            return Gan(0, n, t - u);
          }
          return p[v("BBks", 0, 0, 670) + "ss"] = !1, (null == (n = null == (t = {
            EJihI: function (e) {
              return e();
            }
          }[v("VXV*", 0, 0, a)](Ran)) ? void 0 : t[v(o, 0, 0, r) + "s_sen" + f(0, l, i) + f(0, s, c)]) ? void 0 : n.call(t, e)) ?? p;
        }(s);
        if (l.pendingChat = !1, !(null == e ? void 0 : e.success)) return l.isStreaming = !1, l.streamingText = "", l.streamingReasoning = "", void ((null == e ? void 0 : e.cancelled) || n.error((null == e ? void 0 : e.message) || "发送失败"));
        r === d.value ? (k.value = (null == (a = e.data) ? void 0 : a.action) || null, await Ce(r)) : await Se(r);
      } catch (Ofr) {
        return l.pendingChat = !1, l.isStreaming = !1, l.streamingText = "", l.streamingReasoning = "", void n.error((null == Ofr ? void 0 : Ofr.message) || "发送失败");
      } finally {
        await ke(), (null == (o = v.value) ? void 0 : o.bookId) && (await xe(v.value.bookId));
      }
    }
    Ro(ee, e => {
      localStorage.setItem("novel_book_panel_collapsed", e ? "1" : "0");
    }), Ro(de, e => {
      "running" === (null == e ? void 0 : e.status) && (ee.value = !1);
    }), Ro(R, e => {
      localStorage.setItem("novel_studio_thinking_auto_expand", e ? "1" : "0");
    }), Ro(P, e => {
      e && localStorage.setItem("novel_studio_selected_model", e);
    }), Ro(M, e => {
      localStorage.setItem("novel_studio_selected_style_ids", JSON.stringify(e || []));
    }, {
      deep: !0
    }), Ro(b, function () {
      try {
        b.value ? sessionStorage.setItem("novel_tool_execution", JSON.stringify(b.value)) : sessionStorage.removeItem("novel_tool_execution");
      } catch {}
    }, {
      deep: !0
    }), Ro(T, e => {
      e || "import" !== g.value || (g.value = "chat");
    });
    const Le = {
        imitation: "architect",
        fanfic: "fanfic_init",
        spinoff: "spinoff_init",
        chapters: "import_chapters"
      },
      je = {
        write_next: "writer",
        create_book: "architect",
        audit_chapter: "auditor",
        revise_chapter: "reviser",
        short_run: "short_fiction_run"
      };
    function Ee(e, t = "") {
      var n, a;
      b.value = {
        sessionId: d.value,
        toolName: e,
        bookTitle: t || (null == (n = p.value) ? void 0 : n.title) || (null == (a = k.value) ? void 0 : a.title) || "",
        status: "running",
        startedAt: Date.now(),
        stages: [],
        logs: []
      };
    }
    async function Fe({
      tab: e,
      payload: t
    }) {
      if (T.value = !1, _e(e), d.value || (await Pe(e)), ge(), Ee(Le[e] || "architect"), "imitation" === e && t) try {
        t.sessionId = d.value, t.model = P.value;
        const e = await async function (e) {
          var t, n;
          const a = "vp#)",
            o = 544,
            r = "k*oF",
            l = "iUeH",
            i = 282,
            s = 1072,
            c = "2CEX",
            u = 568,
            d = {};
          function v(e, t, n, a) {
            return Gan(0, t, a - -34);
          }
          return d.success = !1, (null == (n = null == (t = Ran()) ? void 0 : t[v(0, a, 0, o) + v(0, r, 0, 391) + v(0, l, 0, i) + (p = s, f = c, Tan(0, 0, p - u, f)) + "n"]) ? void 0 : n.call(t, e)) ?? d;
          var p, f;
        }(t);
        (null == e ? void 0 : e.success) ? await Xe(e.data) : (n.error((null == e ? void 0 : e.message) || "创建失败"), b.value = null);
      } catch (Ofr) {
        n.error("创建失败"), b.value = null;
      }
    }
    function Be(e) {
      const t = ft(e);
      if (_.value = {
        sessionId: t,
        tool: e.tool,
        checkpoint: e.checkpoint,
        storyId: e.storyId,
        bookId: e.bookId,
        message: e.message
      }, ee.value = !1, e.storyId && (D.value = e.storyId, rt()), e.bookId && (L.value = e.bookId, it()), b.value) {
        const t = b.value.stages || [];
        if (t.length) {
          let n = t;
          "outline" === e.checkpoint ? n = t.map((e, t) => ({
            ...e,
            status: t < 3 ? "completed" : "pending"
          })) : "blueprint" === e.checkpoint && (n = t.map((e, t) => ({
            ...e,
            status: t <= 2 ? "completed" : "pending"
          }))), b.value = {
            ...b.value,
            status: "running",
            stages: n
          }, S.value = n;
        }
      } else if ("short_fiction_run" === e.tool) {
        Ee("short_fiction_run");
        const e = (S.value.length ? S.value : ["生成短篇大纲", "审阅大纲", "修订大纲", "撰写正文", "审阅正文", "修订正文", "打包输出"].map(e => ({
          name: e,
          status: "pending"
        }))).map((e, t) => ({
          ...e,
          status: t < 3 ? "completed" : "pending"
        }));
        b.value = {
          ...b.value,
          status: "running",
          stages: e
        }, S.value = e;
      }
    }
    async function Ve() {
      var e, t, a, o;
      if (!_.value) return;
      w.value = !0;
      const r = {
          ..._.value
        },
        l = Qre(P.value);
      try {
        const i = await async function (e) {
          var t, n;
          const a = "G^6D",
            o = "G^6D",
            r = 570,
            l = "6mzS",
            i = 975,
            s = "S*tx",
            c = 742,
            u = "OCei",
            d = "*^FY",
            v = 347;
          function p(e, t, n, a) {
            return Gan(0, n, e - 380);
          }
          const f = {};
          function m(e, t, n, a) {
            return Gan(0, e, n - v);
          }
          return f.success = !1, (null == (n = null == (t = {
            dJmbX: function (e) {
              return e();
            }
          }[p(868, 0, a)](Ran)) ? void 0 : t[m(o, 0, r) + p(624, 0, l) + p(i, 0, "*^FY") + "pipel" + m(s, 0, c) + p(888, 0, u) + m(d, 0, 556)]) ? void 0 : n.call(t, e)) ?? f;
        }({
          sessionId: r.sessionId || d.value,
          tool: r.tool,
          checkpoint: r.checkpoint,
          storyId: r.storyId,
          bookId: r.bookId,
          message: r.message,
          model: P.value || void 0,
          codexReasoningEffort: l,
          claudeReasoningEffort: l
        });
        if (!(null == i ? void 0 : i.success)) return void n.error((null == i ? void 0 : i.message) || "继续失败");
        if (null == (e = i.data) ? void 0 : e.awaitingConfirmation) return Be({
          ...i.data.checkpoint,
          tool: r.tool,
          sessionId: r.sessionId || d.value,
          message: r.message
        }), void (await Ce(d.value));
        _.value = null, "architect" === r.tool ? n.success("已确认基础设定") : "imitation_short_run" === r.tool ? (Q.value = !1, at(), (null == (t = i.data) ? void 0 : t.shortFiction) ? await dt(i.data.shortFiction) : n.success("仿写短篇已完成")) : (n.success("已继续生成"), (null == (o = null == (a = i.data) ? void 0 : a.shortFiction) ? void 0 : o.storyId) && (D.value = i.data.shortFiction.storyId, await rt())), await Ce(d.value), await ke();
      } finally {
        w.value = !1;
      }
    }
    function Ue() {
      _.value = null;
    }
    async function Ke() {
      var e, t, a, o, r, l, i, s, c, f, h, g, W, y, w, _, C, R, O, M, I, N;
      x.value = !0;
      const G = k.value,
        q = null == G ? void 0 : G.action;
      q && je[q] && Ee(je[q]);
      const A = Qre(P.value),
        T = await async function (e) {
          var t, n;
          const a = 262,
            o = 148,
            r = "#WGY",
            l = 224,
            i = "6mzS";
          function s(e, t, n, a) {
            return Tan(0, 0, a - -181, n);
          }
          const c = {};
          return c[s(73, 437, "W^o1", 208) + "ss"] = !1, (null == (n = null == (t = {
            LtASV: function (e) {
              return e();
            }
          }[u = "c3iw", d = -294, Tan(0, 0, d - -677, u)](Ran)) ? void 0 : t["novels_con" + s(a, o, r, l) + s(-112, -58, i, -61) + "n"]) ? void 0 : n.call(t, e)) ?? c;
          var u, d;
        }({
          sessionId: d.value,
          model: P.value || void 0,
          codexReasoningEffort: A,
          claudeReasoningEffort: A
        });
      if (x.value = !1, !(null == T ? void 0 : T.success)) return b.value && (b.value = {
        ...b.value,
        status: "error",
        completedAt: Date.now(),
        error: (null == T ? void 0 : T.message) || "执行失败"
      }), n.error((null == T ? void 0 : T.message) || "执行失败"), void (await Ce(d.value));
      if (null == (e = T.data) ? void 0 : e.assistedAction) {
        const e = T.data.targetRoute || "";
        return ge(), e.includes("imitation") ? Oe("imitation") : e.includes("fanfic") ? Oe("fanfic") : e.includes("spinoff") ? Oe("spinoff") : e.includes("chapters") ? Oe("chapters") : Oe("imitation"), n.info("已打开导入素材，请补充材料"), void (await Ce(d.value));
      }
      if ((null == (t = T.data) ? void 0 : t.book) && (p.value = T.data.book, ee.value = !1, await ye()), (null == (a = T.data) ? void 0 : a.switchToBook) && (null == (o = T.data) ? void 0 : o.session)) return m.value = "book", d.value = T.data.session.id, v.value = T.data.session, k.value = null, (null == (r = T.data.book) ? void 0 : r.id) && (u[T.data.book.id] = !0, await xe(T.data.book.id)), ge(), n.success("书籍已创建，已切换到书籍会话"), void (await ke());
      if (null == (l = T.data) ? void 0 : l.awaitingConfirmation) {
        const e = T.data.checkpoint || T.data.shortFiction || {};
        return Be({
          ...e,
          tool: "short_run" === q ? "short_fiction_run" : e.tool,
          sessionId: d.value,
          message: e.message || "内容已生成，可在右侧修改后确认继续"
        }), T.data.session && (v.value = T.data.session), ee.value = !1, n.info("内容已生成，请查看右侧文件并确认后继续"), void (await ke());
      }
      if ((null == (i = T.data) ? void 0 : i.shortFiction) && (ee.value = !1, n.success(`短篇《${T.data.shortFiction.title || ""}》已生成`)), null == (s = T.data) ? void 0 : s.audit) {
        const e = Boolean((null == (c = T.data.audit) ? void 0 : c.passed) ?? (null == (h = null == (f = T.data.audit) ? void 0 : f.audit) ? void 0 : h.passed));
        n[e ? "success" : "warning"](e ? "审稿通过" : "审稿未通过，请查看卡片中的问题列表");
      } else if ((null == (W = null == (g = T.data) ? void 0 : g.chapter) ? void 0 : W.chapterNumber) && !(null == (w = null == (y = T.data) ? void 0 : y.chapter) ? void 0 : w.mode)) {
        const e = T.data.chapter;
        ktn(e.status || "ready-for-review") ? n.warning(`第 ${e.chapterNumber} 章已写出，但审稿未通过，请复核后再继续`) : n.success(`第 ${e.chapterNumber} 章已写入`);
      } else (null == (C = null == (_ = T.data) ? void 0 : _.chapter) ? void 0 : C.mode) && n.success(`第 ${T.data.chapter.chapterNumber} 章修订完成`);
      await Ce(d.value), await ke(), (null == (R = v.value) ? void 0 : R.bookId) && (await xe(v.value.bookId)), S.value = [], (null == (O = T.data) ? void 0 : O.audit) || (null == (I = null == (M = T.data) ? void 0 : M.chapter) ? void 0 : I.chapterNumber) || (null == (N = T.data) ? void 0 : N.shortFiction) || n.success("执行完成");
    }
    function He() {
      k.value = null;
    }
    async function Je() {
      var e;
      const t = (null == (e = v.value) ? void 0 : e.bookId) || d.value;
      var n;
      await Aan(t), (n = d.value) && (B[n] = {
        isStreaming: !1,
        pendingChat: !1,
        streamingText: "",
        streamingReasoning: "",
        pipelineStreamStage: ""
      });
    }
    async function Ze() {
      b.value = null, k.value ? await Ke() : (await Ce(d.value), k.value && (await Ke()));
    }
    async function Xe(e) {
      if (await ye(), null == e ? void 0 : e.id) {
        if (u[e.id] = !0, p.value = e, ee.value = !1, d.value && v.value && !v.value.bookId) {
          const t = await async function (e, t) {
            var n, a;
            const o = 612,
              r = 374,
              l = 476,
              i = 451,
              s = "VDbS",
              c = 733,
              u = "GI7c",
              d = 398,
              v = "BnKt",
              p = 628,
              f = 520,
              m = "k*oF",
              h = 640,
              g = 682;
            function W(e, t, n, a) {
              return Gan(0, n, t - 8);
            }
            const y = {};
            return y[W(655, 584, "j^1X") + "ss"] = !1, (null == (a = null == (n = {
              sfJbz: function (e) {
                return e();
              }
            }[W(o, r, "iUeH")](Ran)) ? void 0 : n[W(l, i, s) + W(c, 643, u) + W(195, d, v) + W(p, f, m) + W(h, g, "5LFs")]) ? void 0 : a.call(n, e, t)) ?? y;
          }(d.value, e.id);
          if (null == t ? void 0 : t.success) return await ke(), await xe(e.id), await Ce(d.value), void n.success("书籍已创建，会话已归档到该书下");
        }
        await Te(e.id);
      }
    }
    let $e = Promise.resolve();
    function Ye(e, t, n = {}) {
      return {
        id: `msg_${Date.now()}_${Math.random().toString(36).slice(2, 8)}`,
        role: e,
        content: t,
        createdAt: new Date().toISOString(),
        ...n
      };
    }
    async function et(e) {
      var t;
      if (!d.value || !(null == e ? void 0 : e.length)) return;
      const n = [...((null == (t = v.value) ? void 0 : t.messages) || []), ...e];
      v.value = {
        ...v.value,
        messages: n
      };
      const a = await qan(d.value, n);
      (null == a ? void 0 : a.success) && (v.value = a.data);
    }
    async function tt(e, t = "") {
      const n = V(e),
        a = n.streamingText.trim(),
        o = n.streamingReasoning.trim();
      if (!a && !o) return;
      const r = String(t || n.pipelineStreamStage || "").trim(),
        l = Ye("assistant", r ? `【${r}】\n\n${a}` : a, {
          pipelineStage: r || void 0
        });
      o && (l.reasoning = o, l.reasoningExpanded = !1), n.streamingText = "", n.streamingReasoning = "", await async function (e, t) {
        var n;
        if (!e || !(null == t ? void 0 : t.length)) return;
        if (e === d.value) return void (await et(t));
        const a = await Nan(e);
        if (!(null == a ? void 0 : a.success)) return;
        const o = [...((null == (n = a.data) ? void 0 : n.messages) || []), ...t];
        await qan(e, o), await ke();
      }(e, [l]);
    }
    function nt(e, t = "") {
      return $e = $e.then(() => tt(e, t)).catch(() => {}), $e;
    }
    function at() {
      j && (clearInterval(j), j = null);
    }
    function ot() {
      F && (clearInterval(F), F = null);
    }
    async function rt() {
      var e, t;
      const n = D.value || (null == (e = v.value) ? void 0 : e.shortStoryId);
      if (!n) return;
      if ((null == (t = v.value) ? void 0 : t.shortStoryId) && n !== v.value.shortStoryId) return;
      const a = await Man(n);
      (null == a ? void 0 : a.success) && a.data && (f.value = a.data, ee.value = !1);
    }
    function lt(e) {
      var t, n;
      e && ((null == (t = b.value) ? void 0 : t.sessionId) && b.value.sessionId !== d.value || (null == (n = v.value) ? void 0 : n.shortStoryId) && v.value.shortStoryId !== e || (D.value = e, at(), rt(), j = setInterval(() => {
        rt();
      }, 5e3)));
    }
    async function it() {
      var e;
      const t = L.value || (null == (e = v.value) ? void 0 : e.bookId);
      if (!t) return;
      const n = await Oan(t);
      (null == n ? void 0 : n.success) && n.data && (p.value = n.data, ee.value = !1);
    }
    function st(e) {
      var t;
      e && ((null == (t = b.value) ? void 0 : t.sessionId) && b.value.sessionId !== d.value || (L.value = e, ot(), it(), F = setInterval(() => {
        it();
      }, 5e3)));
    }
    function ct(e = "仿写短篇") {
      Q.value = !0, function (e = d.value, t = "") {
        const n = V(e);
        n.pipelineStreamStage = t, n.streamingText = "", n.streamingReasoning = "";
      }(d.value), b.value = {
        sessionId: d.value,
        toolName: "imitation_short_run",
        label: "仿写短篇",
        bookTitle: e,
        status: "running",
        startedAt: Date.now(),
        stages: $.map(e => ({
          name: e,
          status: "pending"
        })),
        logs: []
      }, S.value = b.value.stages;
    }
    async function ut(e) {
      var t;
      if (z.value = !1, _e("imitation-short"), d.value && "imitation-short" === J9t(v.value) || (await Pe("imitation-short", null, e.title || "仿写短篇"), await vn()), !d.value) return void n.error("无法创建会话，请重试");
      const a = e.title || "仿写短篇";
      ct(a), ge(), ee.value = !1, await et([Ye("user", `开始仿写短篇：${a}`)]), n.info("仿写短篇已开始，可在右侧查看进度");
      let o = !1;
      try {
        const a = await async function (e) {
          var t, n;
          const a = "$OLu",
            o = "rlaI",
            r = "oP9p",
            l = 1150,
            i = "ZU!e",
            s = "VDbS",
            c = 1316,
            u = 791;
          function d(e, t, n, a) {
            return Gan(0, t, e - -496);
          }
          const v = {};
          function p(e, t, n, a) {
            return Gan(0, t, n - u);
          }
          return v[p(0, "WFL$", 1538) + "ss"] = !1, (null == (n = null == (t = {
            EFVGf: function (e) {
              return e();
            }
          }[d(-104, a)](Ran)) ? void 0 : t[d(-279, "4M#3") + d(188, o) + p(0, r, l) + d(-70, i) + p(0, s, c) + "t"]) ? void 0 : n.call(t, e)) ?? v;
        }({
          ...e,
          sessionId: d.value,
          model: P.value || void 0
        });
        if (!(null == a ? void 0 : a.success)) return b.value && (b.value = {
          ...b.value,
          status: "error",
          completedAt: Date.now(),
          error: (null == a ? void 0 : a.message) || "仿写短篇失败",
          summary: (null == a ? void 0 : a.message) || "仿写短篇失败"
        }), void ((null == a ? void 0 : a.cancelled) || n.error((null == a ? void 0 : a.message) || "仿写短篇失败"));
        if (null == (t = a.data) ? void 0 : t.awaitingConfirmation) return o = !0, Be({
          ...a.data,
          tool: "imitation_short_run",
          sessionId: d.value,
          message: a.data.message || "创作蓝图已生成，可在右侧修改后确认继续"
        }), d.value && (await Ce(d.value)), void n.info("创作蓝图已生成，请查看右侧文件并确认后继续");
        await dt(a.data);
      } catch (Ofr) {
        b.value && (b.value = {
          ...b.value,
          status: "error",
          completedAt: Date.now(),
          error: (null == Ofr ? void 0 : Ofr.message) || "仿写短篇失败",
          summary: (null == Ofr ? void 0 : Ofr.message) || "仿写短篇失败"
        }), n.error((null == Ofr ? void 0 : Ofr.message) || "仿写短篇失败");
      } finally {
        o || (Q.value = !1, at());
      }
    }
    async function dt(e) {
      if (!(null == e ? void 0 : e.storyId)) return;
      await nt(d.value, U.value.pipelineStreamStage), await ye();
      const t = await Man(e.storyId);
      (null == t ? void 0 : t.success) && (f.value = t.data, ee.value = !1, _e("imitation-short"), ge()), d.value && (await Ce(d.value)), n.success(`仿写短篇「${e.title || e.storyId}」已完成`);
    }
    function vt(e, t = {}) {
      return xtn(e, {
        chapterNumber: t.chapterNumber,
        title: t.chapterTitle || t.title,
        wordCount: t.wordCount,
        status: t.status
      });
    }
    const pt = new Set(["chat.start", "chat.delta", "chat.reasoning.delta", "chat.stream.start", "chat.stream.done", "chat.done"]);
    function ft(e) {
      return e.sessionId || d.value || "";
    }
    function mt(e) {
      var t, n, a, o, r, l, i;
      const s = (null == e ? void 0 : e.detail) || {};
      if (pt.has(s.type)) return void function (e) {
        const t = ft(e);
        if (!t) return;
        const n = V(t),
          a = t === d.value;
        "chat.start" === e.type ? (n.isStreaming = !0, n.pendingChat = !1, Se(t)) : "chat.delta" === e.type ? (n.isStreaming = !0, n.streamingText += e.text || "", a && b.value && (b.value = {
          ...b.value,
          streamStage: e.stage || n.pipelineStreamStage,
          streamText: n.streamingText
        })) : "chat.reasoning.delta" === e.type ? (n.streamingReasoning += e.text || "", a && b.value && (b.value = {
          ...b.value,
          streamReasoning: n.streamingReasoning
        })) : "chat.stream.start" !== e.type ? "chat.stream.done" !== e.type ? "chat.done" === e.type && (n.isStreaming = !1, n.pendingChat = !1, n.streamingText = "", n.streamingReasoning = "", n.pipelineStreamStage = "", t === d.value ? Q.value || Ce(t) : Se(t)) : nt(t, e.stage || n.pipelineStreamStage).then(() => {
          a && b.value && (b.value = {
            ...b.value,
            streamText: "",
            streamStage: e.stage || n.pipelineStreamStage
          });
        }) : nt(t, n.pipelineStreamStage).then(() => {
          n.pipelineStreamStage = e.stage || e.message || "", n.isStreaming = !0, a && b.value && (b.value = {
            ...b.value,
            streamStage: e.stage || "",
            streamText: ""
          });
        });
      }(s);
      if (!function (e) {
        return !e.sessionId || !!d.value && e.sessionId === d.value;
      }(s)) return;
      const c = Q.value && s.tool && "imitation_short_run" !== s.tool;
      if ("tool:start" === s.type) {
        if (c) return;
        const e = (s.stages || []).map(e => ({
            ...("string" == typeof e ? {
              name: e
            } : e),
            status: e.status || "pending"
          })),
          i = s.bookTitle || (null == (t = b.value) ? void 0 : t.bookTitle) || (null == (n = p.value) ? void 0 : n.title) || (null == (a = k.value) ? void 0 : a.title) || "";
        b.value = {
          ...b.value,
          sessionId: s.sessionId || d.value,
          toolName: s.tool,
          agent: s.agent,
          bookTitle: i,
          status: "running",
          startedAt: (null == (o = b.value) ? void 0 : o.startedAt) || Date.now(),
          stages: e,
          logs: (null == (r = b.value) ? void 0 : r.logs) || []
        }, S.value = e, "imitation_short_run" === s.tool && s.storyId && lt(s.storyId), "short_fiction_run" === s.tool && s.storyId && lt(s.storyId), "architect" === s.tool && s.bookId && st(s.bookId), ["writer", "auditor", "reviser"].includes(s.tool) && (null == (l = v.value) ? void 0 : l.bookId) && st(v.value.bookId);
      } else if ("tool:log" === s.type) u = s.message, b.value && u && (b.value = {
        ...b.value,
        logs: [...(b.value.logs || []), u]
      });else if ("tool:progress" === s.type) {
        const e = ft(s);
        if ("active" === s.status && e) {
          const t = V(e),
            n = s.message || s.stage || "";
          n && (t.pipelineStreamStage = n), t.isStreaming = !0;
        }
        !function (e) {
          var t;
          if (!b.value) return;
          const n = "number" == typeof e.stageIndex ? e.stageIndex : -1,
            a = (null == (t = b.value.stages) ? void 0 : t.length) ? b.value.stages : S.value;
          if (!(null == a ? void 0 : a.length) || n < 0) return void (b.value = {
            ...b.value,
            activeStage: e.message || e.stage || b.value.activeStage
          });
          const o = a.map((t, a) => ({
            ...t,
            status: a < n ? "completed" : a === n ? e.status || "active" : t.status || "pending"
          }));
          b.value = {
            ...b.value,
            activeStage: e.message || e.stage,
            stages: o
          }, S.value = o;
        }(s), s.storyId && (D.value || (D.value = s.storyId), rt()), s.bookId ? st(s.bookId) : "architect" === s.tool && L.value && it();
      } else if ("tool:checkpoint" === s.type) Be(s);else if ("tool:end" === s.type) {
        if (c) return void (s.storyId && Man(s.storyId).then(e => {
          (null == e ? void 0 : e.success) && (f.value = e.data, ee.value = !1);
        }));
        if (b.value) {
          const e = b.value.bookTitle || (null == (i = p.value) ? void 0 : i.title);
          let t = s.summary || b.value.summary || "任务完成",
            n = !1 === s.success ? "error" : "completed";
          if ("writer" === s.tool && s.chapterNumber) {
            const a = s.status || "ready-for-review";
            t = vt(e, {
              chapterNumber: s.chapterNumber,
              chapterTitle: s.chapterTitle,
              wordCount: s.wordCount,
              status: a
            }), n = _tn(a, !1 !== s.success);
          } else if ("auditor" === s.tool && s.chapterNumber) {
            const e = !1 !== s.auditPassed && "ready-for-review" === s.status;
            t = s.summary || wtn(s.chapterNumber, e), n = e ? "completed" : "error";
          } else "reviser" === s.tool && s.chapterNumber ? t = `第 ${s.chapterNumber} 章修订完成。` : "imitation_short_run" === s.tool ? t = s.title ? `短篇《${s.title}》已生成` : "仿写短篇已完成" : "generate_cover" === s.tool ? t = s.coverImagePath ? "封面图已生成" : "封面提示词已保存" : s.chapterTitle && (t = `第 ${s.chapterNumber} 章《${s.chapterTitle}》已写入`);
          b.value = {
            ...b.value,
            status: n,
            completedAt: Date.now(),
            summary: t,
            coverImagePath: s.coverImagePath || b.value.coverImagePath || "",
            details: {
              ...(b.value.details || {}),
              coverImagePath: s.coverImagePath,
              coverError: s.coverError
            },
            error: "error" === n && !1 === s.success ? s.message || s.error || t : "",
            logs: !1 === s.success ? [...(b.value.logs || []), s.message || s.error].filter(Boolean) : b.value.logs || [],
            stages: (b.value.stages || []).map(e => ({
              ...e,
              status: "completed"
            }))
          };
        }
        S.value = S.value.map(e => ({
          ...e,
          status: "completed"
        }));
        const e = ft(s);
        nt(e, V(e).pipelineStreamStage).then(() => {
          e === d.value ? Ce(e) : Se(e);
        }), ye(), ht(), s.bookId && Oan(s.bookId).then(e => {
          (null == e ? void 0 : e.success) && (p.value = e.data, L.value = s.bookId, ee.value = !1);
        }), s.storyId && Man(s.storyId).then(e => {
          (null == e ? void 0 : e.success) && (f.value = e.data, D.value = s.storyId, ee.value = !1);
        }), "short_fiction_run" === s.tool && !1 !== s.success && at(), "imitation_short_run" === s.tool && !1 !== s.success && (Q.value = !1, at(), ye());
      }
      var u;
    }
    async function ht() {
      if (!W.value) return;
      const e = await Oan(W.value);
      (null == e ? void 0 : e.success) && (p.value = e.data);
    }
    async function gt() {
      if (!y.value) return;
      const e = await Man(y.value);
      (null == e ? void 0 : e.success) && (f.value = e.data);
    }
    async function yt() {
      var e, t, a, o, r, l, i, s, c, u, v;
      if (W.value) {
        ae.value = !0, b.value = {
          toolName: "writer",
          bookTitle: null == (e = p.value) ? void 0 : e.title,
          status: "running",
          startedAt: Date.now(),
          stages: [],
          logs: []
        };
        try {
          const e = await async function (e) {
            var t, n;
            const a = 767,
              o = "0UX(",
              r = 635,
              l = "NnW0",
              i = 27,
              s = {};
            function c(e, t, n, a) {
              return Gan(0, t, e - i);
            }
            return s.success = !1, (null == (n = null == (t = {
              kOvFK: function (e) {
                return e();
              }
            }[c(a, o)](Ran)) ? void 0 : t["novel" + (u = "ASta", d = -278, Gan(0, u, d - -842)) + c(r, l) + "xt"]) ? void 0 : n.call(t, e)) ?? s;
            var u, d;
          }({
            bookId: W.value,
            sessionId: d.value,
            model: P.value || void 0
          });
          if (!(null == e ? void 0 : e.success)) return b.value && (b.value = {
            ...b.value,
            status: "error",
            completedAt: Date.now()
          }), void ((null == e ? void 0 : e.cancelled) || n.error((null == e ? void 0 : e.message) || "写章失败"));
          if ((null == (t = e.data) ? void 0 : t.book) && (p.value = e.data.book), "running" === (null == (a = b.value) ? void 0 : a.status)) {
            const t = null == (o = p.value) ? void 0 : o.title,
              n = (null == (r = e.data) ? void 0 : r.status) || "ready-for-review";
            ktn(n);
            b.value = {
              ...b.value,
              status: _tn(n, !0),
              completedAt: Date.now(),
              summary: vt(t, {
                chapterNumber: null == (l = e.data) ? void 0 : l.chapterNumber,
                title: null == (i = e.data) ? void 0 : i.title,
                wordCount: null == (s = e.data) ? void 0 : s.wordCount,
                status: n
              }),
              stages: (b.value.stages || []).map(e => ({
                ...e,
                status: "completed"
              }))
            };
          }
          await ye(), await Ce(d.value);
          ktn((null == (c = e.data) ? void 0 : c.status) || "ready-for-review") ? n.warning(`第 ${null == (u = e.data) ? void 0 : u.chapterNumber} 章已写出，但审稿未通过，请复核后再继续`) : n.success(`第 ${null == (v = e.data) ? void 0 : v.chapterNumber} 章已写入`);
        } finally {
          ae.value = !1;
        }
      }
    }
    function bt() {
      W.value && (oe.value = !0);
    }
    async function kt() {
      oe.value = !1, await ye(), d.value && (await Ce(d.value));
    }
    async function xt() {
      var e, t, a, o, r, l, i, s, c, u;
      if (W.value) {
        ae.value = !0, b.value = {
          toolName: "auditor",
          bookTitle: null == (e = p.value) ? void 0 : e.title,
          status: "running",
          startedAt: Date.now(),
          stages: [],
          logs: []
        };
        try {
          const e = await async function (e) {
            var t, n;
            const a = "oP9p",
              o = 401,
              r = "ASta",
              l = 766,
              i = "XGTO",
              s = 526,
              c = "hjoV",
              u = {};
            function d(e, t, n, a) {
              return Tan(0, 0, n - 367, t);
            }
            return u.success = !1, (null == (n = null == (t = {
              nIiaX: function (e) {
                return e();
              }
            }[d(0, a, o)](Ran)) ? void 0 : t["novel" + d(0, r, l) + d(0, i, s) + d(0, c, 785)]) ? void 0 : n.call(t, e)) ?? u;
          }({
            bookId: W.value,
            sessionId: d.value,
            model: P.value || void 0
          });
          if (!(null == e ? void 0 : e.success)) return b.value && (b.value = {
            ...b.value,
            status: "error",
            completedAt: Date.now()
          }), void n.error((null == e ? void 0 : e.message) || "审稿失败");
          if ("running" === (null == (t = b.value) ? void 0 : t.status)) {
            const t = null == (a = e.data) ? void 0 : a.chapterNumber,
              n = Boolean((null == (o = e.data) ? void 0 : o.passed) ?? (null == (l = null == (r = e.data) ? void 0 : r.audit) ? void 0 : l.passed));
            b.value = {
              ...b.value,
              status: n ? "completed" : "error",
              completedAt: Date.now(),
              summary: (null == (i = e.data) ? void 0 : i.summary) || wtn(t, n),
              stages: (b.value.stages || []).map(e => ({
                ...e,
                status: "completed"
              }))
            };
          }
          const v = Boolean((null == (s = e.data) ? void 0 : s.passed) ?? (null == (u = null == (c = e.data) ? void 0 : c.audit) ? void 0 : u.passed));
          n[v ? "success" : "warning"](v ? "审稿通过" : "审稿未通过，请查看卡片中的问题列表"), await Ce(d.value);
        } finally {
          ae.value = !1;
        }
      }
    }
    async function wt({
      index: e
    }) {
      var t, a;
      if (!d.value || e < 0) return;
      const o = [...((null == (t = v.value) ? void 0 : t.messages) || [])];
      if (!o[e]) return;
      let r = 1;
      "user" === o[e].role && "assistant" === (null == (a = o[e + 1]) ? void 0 : a.role) && (r = 2), o.splice(e, r);
      const l = await qan(d.value, o);
      (null == l ? void 0 : l.success) ? v.value = l.data : n.error((null == l ? void 0 : l.message) || "删除失败");
    }
    function _t({
      content: e
    }) {
      var t, a;
      if (!(null == e ? void 0 : e.trim())) return;
      const o = `novel_${d.value || "default"}`,
        r = (null == (t = v.value) ? void 0 : t.title) || (null == (a = p.value) ? void 0 : a.title) || "小说会话";
      le.value = !1, re.value = !0, vn(() => {
        var t;
        null == (t = ie.value) || t.appendOrCreate(o, r, e), n.success("已存档到备忘录");
      });
    }
    async function St() {
      var e, t, a, o, r;
      if (W.value) {
        ae.value = !0, b.value = {
          toolName: "reviser",
          bookTitle: null == (e = p.value) ? void 0 : e.title,
          status: "running",
          startedAt: Date.now(),
          stages: [],
          logs: []
        };
        try {
          const e = await async function (e) {
            var t, n;
            const a = "W^o1",
              o = "ZU!e",
              r = 184,
              l = "5LFs",
              i = "A3ag";
            function s(e, t, n, a) {
              return Tan(0, 0, t - -177, n);
            }
            function c(e, t, n, a) {
              return Tan(0, 0, t - -369, e);
            }
            const u = {};
            return u[s(0, 101, "BnKt") + "ss"] = !1, (null == (n = null == (t = {
              HAqhW: function (e) {
                return e();
              }
            }[c(a, -125)](Ran)) ? void 0 : t[c(o, -93) + s(0, r, l) + c(i, -266) + s(0, -88, "yN0A") + "r"]) ? void 0 : n.call(t, e)) ?? u;
          }({
            bookId: W.value,
            sessionId: d.value,
            model: P.value || void 0,
            mode: "polish"
          });
          if (!(null == e ? void 0 : e.success)) return b.value && (b.value = {
            ...b.value,
            status: "error",
            completedAt: Date.now()
          }), void ((null == e ? void 0 : e.cancelled) || n.error((null == e ? void 0 : e.message) || "修订失败"));
          if ((null == (t = e.data) ? void 0 : t.book) && (p.value = e.data.book), "running" === (null == (a = b.value) ? void 0 : a.status)) {
            const t = null == (o = e.data) ? void 0 : o.chapterNumber;
            b.value = {
              ...b.value,
              status: "completed",
              completedAt: Date.now(),
              summary: t ? `第 ${t} 章修订完成。` : "修订完成",
              stages: (b.value.stages || []).map(e => ({
                ...e,
                status: "completed"
              }))
            };
          }
          await ye(), await Ce(d.value), n.success(`第 ${null == (r = e.data) ? void 0 : r.chapterNumber} 章已修订`);
        } finally {
          ae.value = !1;
        }
      }
    }
    function Ct(e) {
      if (!(null == e ? void 0 : e.id)) return;
      const t = [{
        label: "Markdown (.md)",
        value: "md",
        mime: "text/markdown"
      }, {
        label: "纯文本 (.txt)",
        value: "txt",
        mime: "text/plain"
      }, {
        label: "EPUB (.epub)",
        value: "epub",
        mime: "application/epub+zip"
      }];
      a.create({
        title: "选择导出格式",
        content: () => hl("div", {
          style: "display:flex;flex-direction:column;gap:8px;padding:8px 0"
        }, t.map(t => hl("button", {
          style: "padding:8px 12px;border:1px solid rgba(0,0,0,.1);border-radius:6px;background:transparent;cursor:pointer;text-align:left;font-size:13px",
          onClick: () => async function (e, t) {
            a.destroyAll();
            try {
              const a = await async function (e, t = "md") {
                var n, a;
                const o = "NnW0",
                  r = "kdlK",
                  l = "S*tx";
                function i(e, t, n, a) {
                  return Tan(0, 0, n - -140, a);
                }
                const s = {};
                return s[i(0, 0, 161, "ZU!e") + "ss"] = !1, (null == (a = null == (n = Ran()) ? void 0 : n[i(0, 0, 25, o) + i(0, 0, -84, r) + "ort_b" + (c = -33, u = l, Gan(0, u, c - -492))]) ? void 0 : a.call(n, e, t)) ?? s;
                var c, u;
              }(e.id, t.value);
              if (!(null == a ? void 0 : a.success)) return void n.error((null == a ? void 0 : a.message) || "导出失败");
              const o = new Blob([a.data.content || ""], {
                  type: `${t.mime};charset=utf-8`
                }),
                r = URL.createObjectURL(o),
                l = document.createElement("a");
              l.href = r, l.download = a.data.filename || `${e.title || "book"}.${t.value}`, l.click(), URL.revokeObjectURL(r), n.success("导出完成");
            } catch {
              n.error("导出失败");
            }
          }(e, t)
        }, t.label)))
      });
    }
    function Pt(e) {
      q.value = "session", G.value = e, N.value = e.title || "", I.value = !0;
    }
    function Rt(e) {
      q.value = "book", G.value = e, N.value = e.title || "", I.value = !0;
    }
    function Ot(e) {
      a.warning({
        title: "删除作品",
        content: `确定删除《${e.title || "未命名"}》？作品文件与关联会话将一并删除，且不可恢复。`,
        positiveText: "删除",
        negativeText: "取消",
        onPositiveClick: async () => {
          const t = await async function (e) {
            var t, n;
            const a = "0UX(",
              o = 252,
              r = "j^1X",
              l = "!jb8";
            function i(e, t, n, a) {
              return Gan(0, a, e - -443);
            }
            const s = {};
            return s[i(7, 0, 0, "!jb8") + "ss"] = !1, (null == (n = null == (t = Ran()) ? void 0 : t[i(110, 0, 0, a) + i(o, 0, 0, r) + "ete_b" + i(240, 0, 0, l)]) ? void 0 : n.call(t, e)) ?? s;
          }(e.id);
          (null == t ? void 0 : t.success) ? (delete c[e.id], delete u[e.id], await ye(), W.value === e.id && we(), n.success("作品已删除")) : n.error((null == t ? void 0 : t.message) || "删除失败");
        }
      });
    }
    async function Mt() {
      if (!G.value) return;
      if ("book" === q.value) {
        const e = await async function (e, t) {
          var n, a;
          const o = "Weig",
            r = 322,
            l = 1391,
            i = "$ex^",
            s = 281,
            c = {};
          function u(e, t, n, a) {
            return Tan(0, 0, t - 101, e);
          }
          function d(e, t, n, a) {
            return Gan(0, t, e - 890);
          }
          return c[d(1471, "$OLu") + "ss"] = !1, (null == (a = null == (n = Ran()) ? void 0 : n[u(o, r) + d(l, i) + u("$OLu", s) + "ook"]) ? void 0 : a.call(n, e, t)) ?? c;
        }(G.value.id, N.value);
        return I.value = !1, (null == e ? void 0 : e.success) ? (await ye(), W.value === G.value.id && (await ht()), void n.success("书名已更新")) : void n.error((null == e ? void 0 : e.message) || "重命名失败");
      }
      const e = await async function (e, t) {
        var n, a;
        const o = 210,
          r = "$ex^",
          l = 1405,
          i = "j^1X",
          s = 1201,
          c = 205,
          u = 479,
          d = 223,
          v = "!jb8",
          p = 712;
        function f(e, t, n, a) {
          return Tan(0, 0, n - -280, a);
        }
        const m = {};
        return m[f(-431, -390, -277, "k*oF") + "ss"] = !1, (null == (a = null == (n = Ran()) ? void 0 : n[f(-37, 415, o, r) + (h = i, g = s, Gan(0, h, g - p)) + f(c, u, d, v) + "ession"]) ? void 0 : a.call(n, e, t)) ?? m;
        var h, g;
      }(G.value.id, N.value);
      I.value = !1, (null == e ? void 0 : e.success) ? (await ke(), G.value.bookId && (await xe(G.value.bookId)), d.value === G.value.id && (await Ce(G.value.id))) : n.error((null == e ? void 0 : e.message) || "重命名失败");
    }
    async function It(e) {
      const t = await async function (e) {
        var t, n;
        const a = "$ex^",
          o = 341,
          r = 159,
          l = {};
        function i(e, t, n, a) {
          return Gan(0, e, n - -357);
        }
        function s(e, t, n, a) {
          return Tan(0, 0, t - -58, a);
        }
        return l[i("rlaI", 0, -24) + "ss"] = !1, (null == (n = null == (t = Ran()) ? void 0 : t[i(a, 0, o) + "s_del" + s(r, 196, 0, "!jb8") + s(-148, 87, 0, "6mzS") + "n"]) ? void 0 : n.call(t, e)) ?? l;
      }(e.id);
      if (null == t ? void 0 : t.success) {
        if (await ke(), e.bookId && (await xe(e.bookId)), d.value === e.id) {
          d.value = "", v.value = null, p.value = null, f.value = null, b.value = null, at(), ot(), Q.value = !1, D.value = "", L.value = "";
          const e = fe.value[0];
          e ? await Ce(e.id) : we();
        }
      } else n.error((null == t ? void 0 : t.message) || "删除失败");
    }
    return Ma(async () => {
      window.addEventListener("novel-studio", mt);
      const e = localStorage.getItem("novel_studio_selected_model") || "",
        t = await BWn(e);
      C.value = (null == t ? void 0 : t.options) || [], P.value = function (e, t = "") {
        var n;
        const a = 48,
          o = "Mi8!",
          r = "a!EF",
          l = 890,
          i = "x0bA",
          s = "#lGi",
          c = "cOAw",
          u = 615,
          d = 769,
          v = "nvSI",
          p = "JW7w",
          f = 757,
          m = "9Sms",
          h = 612,
          g = {};
        g[b(786, "f8^g", 588, 680)] = function (e, t) {
          return e !== t;
        }, g[k(-118, a, o, -29)] = "dnoNI";
        const W = g,
          y = Array[k(62, -157, r, -9) + "ay"](e) ? e : [];
        function b(e, t, n, a) {
          return VWn(a - h, t);
        }
        if (t && y[b(0, "]^4M", 0, l)](e => (null == e ? void 0 : e.value) === t)) return t;
        function k(e, t, n, a) {
          return VWn(a - -317, n);
        }
        for (const x of FWn) {
          if (W[k(0, 0, i, -213)](W[b(0, s, 0, 602)], W[b(0, c, 0, u)])) {
            const e = _0x4ae84f ? function () {
              if (_0x3948b1) {
                const n = _0x3e60e8[e = 1195, t = m, k(0, 0, t, e - 1321)](_0x168273, arguments);
                return _0x3dcaa5 = null, n;
              }
              var e, t;
            } : function () {};
            return _0x268ddc = !1, e;
          }
          {
            const e = y[b(0, "PGoq", 0, d)](e => {
              return t = null == e ? void 0 : e[b(0, "[ku)", 0, 776)], String(t || "")[b(0, "*hue", 0, 812)]()["toLow" + b(0, "*hue", 0, 760) + "e"]() === x;
              var t;
            });
            if (null == e ? void 0 : e[k(0, 0, v, -325)]) return e[b(0, p, 0, f)];
          }
        }
        return (null == (n = y[0]) ? void 0 : n[k(0, 0, "xY1%", -211)]) || "";
      }(C.value, e), await Promise.all([ye(), ke(), be()]), _e(m.value);
    }), Ga(() => {
      window.removeEventListener("novel-studio", mt), at(), ot();
    }), (t, n) => {
      var a, o, r, b, S, G, Q, D;
      return Cr(), Ir(br, null, [zr("div", {
        class: K(["novel-studio-page", {
          dark: e.isDarkMode
        }]),
        style: E(l.value)
      }, [Qr(Ben, {
        books: i.value,
        "project-sessions": s.value,
        "mode-sessions": fe.value,
        "session-progress": me.value,
        "active-workflow": m.value,
        "book-sessions": c,
        "expanded-books": u,
        "active-session-id": d.value,
        "active-book-id": W.value,
        "active-mode": m.value,
        "active-tool": g.value,
        "is-dark-mode": e.isDarkMode,
        onLaunchMode: Re,
        onSwitchWorkflow: _e,
        onCreateSession: Pe,
        onCreateBookSession: n[0] || (n[0] = e => Pe("book", e)),
        onOpenSession: Ce,
        onOpenBook: Te,
        onToggleBook: ze,
        onRenameSession: Pt,
        onDeleteSession: It,
        onRenameBook: Rt,
        onDeleteBook: Ot,
        onExportBook: Ct,
        onOpenImport: Oe,
        onOpenStyle: Ie,
        onOpenImportModal: Me,
        onOpenImitationShort: n[1] || (n[1] = e => z.value = !0),
        onOpenAgentSettings: Ne,
        onOpenScheduler: Ge,
        onOpenDashboard: qe,
        onOpenMemo: Ae
      }, null, 8, ["books", "project-sessions", "mode-sessions", "session-progress", "active-workflow", "book-sessions", "expanded-books", "active-session-id", "active-book-id", "active-mode", "active-tool", "is-dark-mode"]), "chat" === h.value ? (Cr(), Ir("div", HWn, [Qr(fan, {
        class: "chat-main-column",
        session: v.value,
        "is-streaming": H.value,
        "streaming-text": J.value,
        "streaming-reasoning": Z.value,
        "pipeline-stream-stage": X.value,
        "tool-execution": he.value,
        "pending-action": k.value,
        "confirm-loading": x.value,
        "show-quick-actions": se.value,
        "quick-action-loading": ae.value,
        "model-options": C.value,
        "selected-model": P.value,
        "onUpdate:selectedModel": n[2] || (n[2] = e => P.value = e),
        "thinking-mode": !0,
        "thinking-auto-expand": R.value,
        "onUpdate:thinkingAutoExpand": n[3] || (n[3] = e => R.value = e),
        "is-dark-mode": e.isDarkMode,
        "book-title-fallback": (null == (a = p.value) ? void 0 : a.title) || (null == (o = f.value) ? void 0 : o.title) || "",
        onSend: De,
        onClearContext: Qe,
        onConfirmAction: Ke,
        onDismissAction: He,
        onCancel: Je,
        onQuickWriteNext: yt,
        onQuickWriteBatch: bt,
        onQuickAudit: xt,
        onQuickRevise: St,
        onDeleteMessage: wt,
        onArchiveMessage: _t,
        onRetryAction: Ze
      }, null, 8, ["session", "is-streaming", "streaming-text", "streaming-reasoning", "pipeline-stream-stage", "tool-execution", "pending-action", "confirm-loading", "show-quick-actions", "quick-action-loading", "model-options", "selected-model", "thinking-auto-expand", "is-dark-mode", "book-title-fallback"]), ve.value && ee.value ? (Cr(), Ir("div", JWn, [zr("button", {
        type: "button",
        class: "panel-reopen-btn",
        onClick: n[4] || (n[4] = e => ee.value = !1)
      }, Y(pe.value), 1)])) : Fr("", !0), ve.value && !ee.value ? (Cr(), Ir(br, {
        key: 1
      }, [zr("div", {
        class: "panel-resizer",
        onMousedown: ne
      }, null, 32), zr("div", {
        class: "reference-panel-shell",
        style: E({
          width: `${te.value}px`
        })
      }, [zr("div", ZWn, [de.value ? (Cr(), Ir("div", XWn, [Qr(snn, {
        execution: de.value,
        "book-title-fallback": (null == (r = p.value) ? void 0 : r.title) || (null == (b = f.value) ? void 0 : b.title) || "",
        "is-dark-mode": e.isDarkMode,
        "hide-stream-output": "",
        onRetry: Ze
      }, null, 8, ["execution", "book-title-fallback", "is-dark-mode"])])) : Fr("", !0), _.value ? (Cr(), Nr(TWn, {
        key: 1,
        checkpoint: _.value,
        loading: w.value,
        onConfirm: Ve,
        onDismiss: Ue
      }, null, 8, ["checkpoint", "loading"])) : Fr("", !0), zr("div", $Wn, [se.value ? (Cr(), Nr(fon, {
        key: 0,
        book: p.value,
        "book-id": W.value,
        collapsed: ee.value,
        "onUpdate:collapsed": n[5] || (n[5] = e => ee.value = e),
        "is-dark-mode": e.isDarkMode,
        onSaved: ht
      }, null, 8, ["book", "book-id", "collapsed", "is-dark-mode"])) : Fr("", !0), ce.value ? (Cr(), Nr(Lon, {
        key: 1,
        "short-story": f.value,
        "short-story-id": y.value,
        collapsed: ee.value,
        "onUpdate:collapsed": n[6] || (n[6] = e => ee.value = e),
        "is-dark-mode": e.isDarkMode,
        onSaved: gt
      }, null, 8, ["short-story", "short-story-id", "collapsed", "is-dark-mode"])) : Fr("", !0), !de.value || se.value || ce.value ? Fr("", !0) : (Cr(), Ir("div", YWn, " 任务执行中，AI 输出请查看左侧对话框 "))])])], 4)], 64)) : Fr("", !0)])) : "style" === h.value ? (Cr(), Nr(rrn, {
        key: 1,
        "style-sources": O.value,
        "selected-style-ids": M.value,
        "onUpdate:selectedStyleIds": n[7] || (n[7] = e => M.value = e),
        "is-dark-mode": e.isDarkMode,
        onBack: ge,
        onRefresh: be,
        onApplied: We
      }, null, 8, ["style-sources", "selected-style-ids", "is-dark-mode"])) : "agent-settings" === h.value ? (Cr(), Nr(Gln, {
        key: 2,
        "model-options": C.value,
        "is-dark-mode": e.isDarkMode,
        onBack: ge
      }, null, 8, ["model-options", "is-dark-mode"])) : "scheduler" === h.value ? (Cr(), Ir("div", eyn, [zr("div", {
        class: "panel-page-header"
      }, [zr("button", {
        class: "back-btn",
        onClick: ge
      }, "← 返回"), n[16] || (n[16] = zr("span", {
        class: "panel-page-title"
      }, "定时写作", -1))]), zr("div", tyn, [Qr($cn, {
        "book-id": W.value,
        "is-dark-mode": e.isDarkMode
      }, null, 8, ["book-id", "is-dark-mode"])])])) : "dashboard" === h.value ? (Cr(), Ir("div", nyn, [zr("div", {
        class: "panel-page-header"
      }, [zr("button", {
        class: "back-btn",
        onClick: ge
      }, "← 返回"), n[17] || (n[17] = zr("span", {
        class: "panel-page-title"
      }, "数据看板", -1))]), zr("div", ayn, [Qr(Run, {
        chapters: (null == (S = p.value) ? void 0 : S.chapters) || [],
        "is-dark-mode": e.isDarkMode
      }, null, 8, ["chapters", "is-dark-mode"]), zr("div", oyn, [n[18] || (n[18] = zr("h4", {
        style: {
          "font-size": "13px",
          "font-weight": "600",
          margin: "0 0 8px"
        }
      }, "伏笔管理", -1)), Qr(cun, {
        hooks: (null == (G = p.value) ? void 0 : G.hooks) || [],
        "current-chapter": (null == (D = null == (Q = p.value) ? void 0 : Q.chapters) ? void 0 : D.length) || 0,
        "is-dark-mode": e.isDarkMode
      }, null, 8, ["hooks", "current-chapter", "is-dark-mode"])])])])) : Fr("", !0)], 6), Qr(Bsn, {
        show: T.value,
        "onUpdate:show": n[8] || (n[8] = e => T.value = e),
        tab: A.value,
        "onUpdate:tab": n[9] || (n[9] = e => A.value = e),
        "session-id": d.value,
        "style-sources": O.value,
        books: i.value,
        "is-dark-mode": e.isDarkMode,
        onCreated: Xe,
        onSubmitStart: Fe
      }, null, 8, ["show", "tab", "session-id", "style-sources", "books", "is-dark-mode"]), Qr(Gcn, {
        show: z.value,
        "onUpdate:show": n[10] || (n[10] = e => z.value = e),
        "session-id": d.value,
        "style-sources": O.value,
        books: i.value,
        "is-dark-mode": e.isDarkMode,
        onSubmit: ut
      }, null, 8, ["show", "session-id", "style-sources", "books", "is-dark-mode"]), Qr(Tt(wq), {
        show: I.value,
        "onUpdate:show": n[12] || (n[12] = e => I.value = e),
        preset: "dialog",
        title: "book" === q.value ? "重命名作品" : "重命名会话",
        "positive-text": "保存",
        onPositiveClick: Mt
      }, {
        default: Sn(() => [Qr(Tt(r_), {
          value: N.value,
          "onUpdate:value": n[11] || (n[11] = e => N.value = e),
          placeholder: "book" === q.value ? "作品标题" : "会话标题"
        }, null, 8, ["value", "placeholder"])]),
        _: 1
      }, 8, ["show", "title"]), Qr(Dcn, {
        show: oe.value,
        "onUpdate:show": n[13] || (n[13] = e => oe.value = e),
        "book-id": W.value,
        "session-id": d.value,
        model: P.value,
        "is-dark-mode": e.isDarkMode,
        onDone: kt
      }, null, 8, ["show", "book-id", "session-id", "model", "is-dark-mode"]), (Cr(), Nr(zn, {
        to: "body"
      }, [Qr(Al, {
        name: "memo-fade"
      }, {
        default: Sn(() => [re.value ? (Cr(), Ir("div", {
          key: 0,
          class: "memo-overlay",
          onClick: n[15] || (n[15] = Zi(e => re.value = !1, ["self"]))
        }, [zr("div", {
          class: K(["memo-popup", {
            "is-dark": e.isDarkMode
          }])
        }, [Qr(RWn, {
          ref_key: "notebookRef",
          ref: ie,
          visible: re.value,
          "onUpdate:visible": n[14] || (n[14] = e => re.value = e),
          "is-dark-mode": e.isDarkMode,
          "session-key": `novel_${d.value || "default"}`,
          "selected-model": P.value,
          "show-all": le.value
        }, null, 8, ["visible", "is-dark-mode", "session-key", "selected-model", "show-all"])], 2)])) : Fr("", !0)]),
        _: 1
      })]))], 64);
    };
  }
};
