// Inspection only. Original SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Original UTF16 offsets [8629071, 8650803); original line 41
// External dependencies and runtime equivalence remain unresolved.
const recoveredComponent = {
  __name: "NovelStudioChat",
  props: {
    session: {
      type: Object,
      default: null
    },
    isStreaming: Boolean,
    streamingText: String,
    streamingReasoning: String,
    pipelineStreamStage: {
      type: String,
      default: ""
    },
    toolExecution: {
      type: Object,
      default: null
    },
    pendingAction: {
      type: Object,
      default: null
    },
    confirmLoading: Boolean,
    showQuickActions: Boolean,
    quickActionLoading: Boolean,
    modelOptions: {
      type: Array,
      default: () => []
    },
    selectedModel: String,
    thinkingMode: {
      type: Boolean,
      default: !0
    },
    thinkingAutoExpand: {
      type: Boolean,
      default: !0
    },
    isDarkMode: Boolean,
    bookTitleFallback: {
      type: String,
      default: ""
    }
  },
  emits: ["send", "confirm-action", "dismiss-action", "cancel", "quick-write-next", "quick-write-batch", "quick-audit", "quick-revise", "delete-message", "archive-message", "update:selectedModel", "update:thinkingAutoExpand", "clear-context", "retry-action"],
  setup(e, {
    emit: t
  }) {
    const n = e,
      a = t,
      o = $q(),
      r = Nt(""),
      l = Nt(""),
      i = Nt(null),
      s = Nt(null),
      c = Nt(!1),
      u = Nt(!1),
      d = Nt(""),
      v = Nt(""),
      p = Yae,
      f = Nt(!1),
      m = Nt(3e5),
      h = Nt(!1),
      g = [1e5, 2e5, 3e5, 4e5, 1e6, van],
      W = ml(() => m.value >= van ? "无限" : m.value >= 1e6 ? "1m" : `${Math.round(m.value / 1e3)}k`),
      y = ml(() => {
        const e = g.indexOf(m.value);
        return e < 0 ? 0 : g.length <= 1 ? 100 : e / (g.length - 1) * 100;
      }),
      b = Nt(0),
      k = Nt(!1);
    let x = null;
    const w = ml(() => {
        var e;
        return (null == (e = n.session) ? void 0 : e.messages) || [];
      }),
      _ = Nt(!1),
      S = Nt(null),
      C = Nt([]),
      P = Nt(null),
      R = Nt(null);
    let O = null,
      M = null,
      I = null,
      N = null;
    const G = ml(() => w.value.map((e, t) => {
        const n = e.role;
        let a = "";
        return a = "tool" === n ? e.summary || e.toolName || "任务" : (e.content || "").replace(/\n/g, " ").slice(0, 60) || "...", {
          id: e.id || `msg_${t}`,
          role: n,
          preview: a
        };
      })),
      q = ml(() => {
        const e = G.value;
        if (!e.length) return [];
        const t = Math.min(e.length, 20);
        return Array.from({
          length: t
        }, (n, a) => {
          const o = Math.floor(a * e.length / t),
            r = Math.floor((a + 1) * e.length / t),
            l = e.slice(o, r);
          return {
            index: a,
            isVisible: l.some(e => C.value.includes(e.id)),
            isCurrent: l.some(e => e.id === S.value)
          };
        });
      });
    function A() {
      var e;
      const t = i.value;
      if (!t) return;
      const n = null == (e = t.$el) ? void 0 : e.querySelector(".n-scrollbar-container");
      if (!n) return;
      const {
          scrollTop: a,
          clientHeight: o
        } = n,
        r = a + o / 2,
        l = n.querySelectorAll(".msg-row[data-msg-id]"),
        s = [];
      let c = null;
      for (const i of l) {
        const e = i.getAttribute("data-msg-id"),
          t = i.offsetTop,
          n = t + i.offsetHeight;
        n > a && t < a + o && s.push(e), !c && t <= r && n > r && (c = e);
      }
      C.value = s, S.value = c;
    }
    function T() {
      var e;
      z();
      const t = i.value;
      t && (O = null == (e = t.$el) ? void 0 : e.querySelector(".n-scrollbar-container"), O && (M = () => A(), O.addEventListener("scroll", M, {
        passive: !0
      }), A()));
    }
    function z() {
      O && M && O.removeEventListener("scroll", M), O = null, M = null;
    }
    function Q() {
      const e = R.value;
      if (!e || !S.value) return;
      const t = e.querySelector(`[data-msg-id="${S.value}"]`);
      if (!t) return;
      const n = t.offsetTop - (e.clientHeight - t.offsetHeight) / 2;
      e.scrollTop = Math.max(0, n);
    }
    function D() {
      N && (clearTimeout(N), N = null), _.value || I || (I = setTimeout(() => {
        _.value = !0, I = null, vn(Q);
      }, 180));
    }
    function L() {
      I && (clearTimeout(I), I = null), _.value && (N = setTimeout(() => {
        _.value = !1, P.value = null, N = null;
      }, 120));
    }
    Ro([i, () => w.value.length], () => {
      vn(T);
    }, {
      immediate: !0
    });
    const j = ml(() => {
        const e = Atn(n.session, n.pendingAction);
        return !e || Ttn(n.session, e) ? null : e;
      }),
      F = ml(() => {
        const e = Atn(n.session, n.pendingAction);
        return !!e && Ttn(n.session, e);
      });
    function B(e, t) {
      if (!(null == e ? void 0 : e.action) || "propose_action" !== e.toolName) return null;
      const n = qtn(w.value, t);
      return {
        execution: Gtn({
          ...e.action,
          createdAt: e.createdAt
        }, {
          locked: n
        }),
        action: e.action,
        locked: n
      };
    }
    const V = ml(() => j.value ? Gtn(j.value) : null),
      U = ml(() => {
        if (!j.value || H.value) return !1;
        for (let e = w.value.length - 1; e >= 0; e--) {
          const t = w.value[e];
          if ("tool" === (null == t ? void 0 : t.role)) break;
          if ("assistant" === (null == t ? void 0 : t.role) && t.action) return !1;
        }
        return !0;
      }),
      H = ml(() => {
        if (!n.toolExecution) return null;
        const e = n.toolExecution.status;
        return "running" !== e && "error" !== e ? null : Ntn(n.toolExecution);
      }),
      J = ml(() => aoe(n.modelOptions || [])),
      Z = ml(() => J.value.find(e => e.value === n.selectedModel) || null),
      X = ml(() => {
        var e, t;
        return n.selectedModel ? (null == (e = Z.value) ? void 0 : e.displayName) || (null == (t = Z.value) ? void 0 : t.label) || n.selectedModel : "选择模型";
      }),
      $ = ml(() => {
        var e;
        return (null == (e = Z.value) ? void 0 : e.channelTag) || "";
      }),
      ee = ml(() => {
        var e;
        return (null == (e = Z.value) ? void 0 : e.channelKey) || "";
      }),
      te = ml(() => {
        let e = ooe(J.value, v.value);
        const t = d.value.trim().toLowerCase();
        return t ? e.filter(e => (e.displayName || e.label || "").toLowerCase().includes(t) || (e.channelTag || "").toLowerCase().includes(t) || (e.label || "").toLowerCase().includes(t) || (e.value || "").toLowerCase().includes(t)) : e;
      });
    const ne = ml(() => n.isDarkMode ? "#202124" : "#fafafa"),
      ae = ml(() => n.isDarkMode ? "#2a2b2d" : "#e8e8e8"),
      oe = ml(() => "输入指令… 可在模型列表中设置思考强度 (Enter 发送，Shift+Enter 换行)");
    function re() {
      const e = l.value.trim();
      if (!e || n.isStreaming) return;
      const t = {};
      m.value < van && (t.contextLength = m.value), a("send", e, t), l.value = "";
    }
    function le(e) {
      "Enter" !== e.key || e.shiftKey || e.isComposing || (e.preventDefault(), re()), (e.ctrlKey || e.metaKey) && "k" === e.key && (e.preventDefault(), se());
    }
    function ie(e) {
      const t = e.currentTarget.getBoundingClientRect(),
        n = Math.max(0, Math.min(1, (e.clientX - t.left) / t.width)),
        a = Math.round(n * (g.length - 1));
      m.value = g[a];
    }
    function se() {
      a("clear-context"), o.success("上下文已清除");
    }
    function ce() {
      d.value = "", v.value = "", u.value = !0;
    }
    function ue(e, t) {
      return e.id || `${e.createdAt || t}_${e.role}`;
    }
    function de(e) {
      return e ? "tool" === e.role ? String(e.summary || "").trim() : String(e.content || "").trim() : "";
    }
    function ve(e) {
      return Boolean(de(e)) && "tool" !== e.role;
    }
    function pe(e, t) {
      const n = de(e);
      n && navigator.clipboard.writeText(n).then(() => {
        r.value = ue(e, t), setTimeout(() => {
          r.value === ue(e, t) && (r.value = "");
        }, 1800);
      }).catch(() => {
        o.error("复制失败");
      });
    }
    function fe(e) {
      const t = de(e);
      t && a("archive-message", {
        content: t
      });
    }
    function me(e, t) {
      n.isStreaming ? o.warning("请等待当前回复完成") : a("delete-message", {
        index: e,
        msg: t
      });
    }
    function he() {
      const e = String(n.streamingText || "").trim();
      e && navigator.clipboard.writeText(e).then(() => {
        r.value = "__streaming__", setTimeout(() => {
          "__streaming__" === r.value && (r.value = "");
        }, 1800);
      }).catch(() => {
        o.error("复制失败");
      });
    }
    return Ro(() => n.isStreaming, e => {
      if (x && (clearInterval(x), x = null), !e) return b.value = 0, void (k.value = !1);
      n.thinkingAutoExpand && (k.value = !0);
      const t = Date.now();
      b.value = 1, x = setInterval(() => {
        b.value = Math.max(1, Math.round((Date.now() - t) / 1e3));
      }, 1e3);
    }), Ro(() => w.value.length, async () => {
      var e, t;
      await vn(), null == (t = null == (e = i.value) ? void 0 : e.scrollTo) || t.call(e, {
        top: 999999,
        behavior: "smooth"
      });
    }), Ro(() => n.isStreaming, async e => {
      var t, n;
      e && (await vn(), null == (n = null == (t = i.value) ? void 0 : t.scrollTo) || n.call(t, {
        top: 999999,
        behavior: "smooth"
      }));
    }), Ro(() => n.streamingText, async () => {
      var e, t;
      (n.isStreaming || n.toolExecution) && (await vn(), null == (t = null == (e = i.value) ? void 0 : e.scrollTo) || t.call(e, {
        top: 999999,
        behavior: "auto"
      }));
    }), Ma(() => {
      var e;
      null == (e = s.value) || e.focus();
    }), Ga(() => {
      x && clearInterval(x), z(), I && clearTimeout(I), N && clearTimeout(N);
    }), (t, n) => (Cr(), Ir("main", {
      class: K(["novel-chat", {
        "is-dark": e.isDarkMode
      }])
    }, [zr("div", gnn, [Qr(Tt(IE), {
      ref_key: "scrollbarRef",
      ref: i,
      class: "chat-scroller"
    }, {
      default: Sn(() => [zr("div", Wnn, [w.value.length || e.isStreaming ? Fr("", !0) : (Cr(), Ir("div", ynn, [...(n[30] || (n[30] = [zr("p", null, "告诉我你想写什么——题材、世界观、主角、核心冲突。", -1)]))])), (Cr(!0), Ir(br, null, Va(w.value, (t, o) => (Cr(), Ir(br, {
        key: ue(t, o)
      }, ["user" === t.role ? (Cr(), Ir("div", {
        key: 0,
        class: "row user-row msg-row",
        "data-msg-id": t.id || `msg_${o}`
      }, [zr("div", knn, [zr("div", xnn, Y(t.content), 1), ve(t) ? (Cr(), Ir("div", wnn, [Qr(Tt(PI), {
        trigger: "hover",
        placement: "top"
      }, {
        trigger: Sn(() => [zr("button", {
          type: "button",
          class: K(["msg-act-btn", {
            copied: r.value === ue(t, o)
          }]),
          onClick: e => pe(t, o)
        }, [r.value === ue(t, o) ? (Cr(), Nr(Tt($I), {
          key: 0,
          component: Tt(xH),
          size: 14
        }, null, 8, ["component"])) : (Cr(), Nr(Tt($I), {
          key: 1,
          component: Tt(UJ),
          size: 14
        }, null, 8, ["component"]))], 10, _nn)]),
        default: Sn(() => [n[31] || (n[31] = jr(" 复制 ", -1))]),
        _: 2
      }, 1024), Qr(Tt(PI), {
        trigger: "hover",
        placement: "top"
      }, {
        trigger: Sn(() => [zr("button", {
          type: "button",
          class: "msg-act-btn",
          onClick: e => fe(t)
        }, [Qr(Tt($I), {
          component: Tt(L5),
          size: 14
        }, null, 8, ["component"])], 8, Snn)]),
        default: Sn(() => [n[32] || (n[32] = jr(" 存档 ", -1))]),
        _: 2
      }, 1024), Qr(Tt(oE), {
        "positive-text": "删除",
        "negative-text": "取消",
        onPositiveClick: e => me(o, t)
      }, {
        trigger: Sn(() => [Qr(Tt(PI), {
          trigger: "hover",
          placement: "top"
        }, {
          trigger: Sn(() => [zr("button", {
            type: "button",
            class: "msg-act-btn msg-act-del",
            onClick: n[0] || (n[0] = Zi(() => {}, ["stop"]))
          }, [Qr(Tt($I), {
            component: Tt(k8),
            size: 14
          }, null, 8, ["component"])])]),
          default: Sn(() => [n[33] || (n[33] = jr(" 删除 ", -1))]),
          _: 1
        })]),
        default: Sn(() => {
          var e;
          return [jr(" 确定删除此消息" + Y("assistant" === (null == (e = w.value[o + 1]) ? void 0 : e.role) ? "及对应回复" : "") + "？ ", 1)];
        }),
        _: 2
      }, 1032, ["onPositiveClick"])])) : Fr("", !0)])], 8, bnn)) : "assistant" === t.role ? (Cr(), Ir("div", {
        key: 1,
        class: "row assistant-row msg-row",
        "data-msg-id": t.id || `msg_${o}`
      }, [Qr(pnn, {
        seconds: t.thinkingSeconds,
        reasoning: t.reasoning,
        "default-open": t.reasoningExpanded ?? e.thinkingAutoExpand,
        "thinking-mode": e.thinkingMode,
        "onUpdate:expanded": e => t.reasoningExpanded = e
      }, null, 8, ["seconds", "reasoning", "default-open", "thinking-mode", "onUpdate:expanded"]), zr("div", Pnn, [t.content ? (Cr(), Ir("div", Rnn, Y(t.content), 1)) : Fr("", !0), B(t, o) ? (Cr(), Nr(snn, {
        key: 1,
        execution: B(t, o).execution,
        "embedded-action": B(t, o).action,
        "action-locked": B(t, o).locked,
        "confirm-loading": e.confirmLoading && !B(t, o).locked,
        "is-streaming": e.isStreaming,
        "book-title-fallback": e.bookTitleFallback,
        "is-dark-mode": e.isDarkMode,
        onConfirm: n[1] || (n[1] = e => a("confirm-action")),
        onDismiss: n[2] || (n[2] = e => a("dismiss-action"))
      }, null, 8, ["execution", "embedded-action", "action-locked", "confirm-loading", "is-streaming", "book-title-fallback", "is-dark-mode"])) : Fr("", !0), ve(t) ? (Cr(), Ir("div", Onn, [Qr(Tt(PI), {
        trigger: "hover",
        placement: "top"
      }, {
        trigger: Sn(() => [zr("button", {
          type: "button",
          class: K(["msg-act-btn", {
            copied: r.value === ue(t, o)
          }]),
          onClick: e => pe(t, o)
        }, [r.value === ue(t, o) ? (Cr(), Nr(Tt($I), {
          key: 0,
          component: Tt(xH),
          size: 14
        }, null, 8, ["component"])) : (Cr(), Nr(Tt($I), {
          key: 1,
          component: Tt(UJ),
          size: 14
        }, null, 8, ["component"]))], 10, Mnn)]),
        default: Sn(() => [n[34] || (n[34] = jr(" 复制 ", -1))]),
        _: 2
      }, 1024), Qr(Tt(PI), {
        trigger: "hover",
        placement: "top"
      }, {
        trigger: Sn(() => [zr("button", {
          type: "button",
          class: "msg-act-btn",
          onClick: e => fe(t)
        }, [Qr(Tt($I), {
          component: Tt(L5),
          size: 14
        }, null, 8, ["component"])], 8, Inn)]),
        default: Sn(() => [n[35] || (n[35] = jr(" 存档 ", -1))]),
        _: 2
      }, 1024), Qr(Tt(oE), {
        "positive-text": "删除",
        "negative-text": "取消",
        onPositiveClick: e => me(o, t)
      }, {
        trigger: Sn(() => [Qr(Tt(PI), {
          trigger: "hover",
          placement: "top"
        }, {
          trigger: Sn(() => [zr("button", {
            type: "button",
            class: "msg-act-btn msg-act-del",
            onClick: n[3] || (n[3] = Zi(() => {}, ["stop"]))
          }, [Qr(Tt($I), {
            component: Tt(k8),
            size: 14
          }, null, 8, ["component"])])]),
          default: Sn(() => [n[36] || (n[36] = jr(" 删除 ", -1))]),
          _: 1
        })]),
        default: Sn(() => [n[37] || (n[37] = jr(" 确定删除这条 AI 回复？ ", -1))]),
        _: 1
      }, 8, ["onPositiveClick"])])) : Fr("", !0)])], 8, Cnn)) : "tool" === t.role ? (Cr(), Ir("div", {
        key: 2,
        class: "row assistant-row msg-row",
        "data-msg-id": t.id || `msg_${o}`
      }, [Qr(snn, {
        execution: Tt(Ntn)(t),
        "book-title-fallback": e.bookTitleFallback,
        "is-dark-mode": e.isDarkMode
      }, null, 8, ["execution", "book-title-fallback", "is-dark-mode"])], 8, Nnn)) : Fr("", !0)], 64))), 128)), e.isStreaming ? (Cr(), Ir("div", Gnn, [e.streamingReasoning || b.value > 0 ? (Cr(), Nr(pnn, {
        key: 0,
        seconds: b.value,
        reasoning: e.streamingReasoning,
        "default-open": k.value,
        "thinking-mode": e.thinkingMode,
        "onUpdate:expanded": n[4] || (n[4] = e => k.value = e)
      }, null, 8, ["seconds", "reasoning", "default-open", "thinking-mode"])) : Fr("", !0), zr("div", qnn, [e.pipelineStreamStage ? (Cr(), Ir("div", Ann, Y(e.pipelineStreamStage), 1)) : Fr("", !0), e.streamingText ? (Cr(), Ir("div", Tnn, [jr(Y(e.streamingText), 1), n[38] || (n[38] = zr("span", {
        class: "stream-cursor"
      }, "|", -1))])) : (Cr(), Nr(fnn, {
        key: 1,
        "is-dark-mode": e.isDarkMode
      }, null, 8, ["is-dark-mode"])), e.streamingText ? (Cr(), Ir("div", znn, [Qr(Tt(PI), {
        trigger: "hover",
        placement: "top"
      }, {
        trigger: Sn(() => [zr("button", {
          type: "button",
          class: K(["msg-act-btn", {
            copied: "__streaming__" === r.value
          }]),
          onClick: he
        }, ["__streaming__" === r.value ? (Cr(), Nr(Tt($I), {
          key: 0,
          component: Tt(xH),
          size: 14
        }, null, 8, ["component"])) : (Cr(), Nr(Tt($I), {
          key: 1,
          component: Tt(UJ),
          size: 14
        }, null, 8, ["component"]))], 2)]),
        default: Sn(() => [n[39] || (n[39] = jr(" 复制 ", -1))]),
        _: 1
      })])) : Fr("", !0)])])) : Fr("", !0), U.value && V.value ? (Cr(), Nr(snn, {
        key: 2,
        execution: V.value,
        "embedded-action": j.value,
        "action-locked": F.value,
        "confirm-loading": e.confirmLoading,
        "is-streaming": e.isStreaming,
        "book-title-fallback": e.bookTitleFallback,
        "is-dark-mode": e.isDarkMode,
        onConfirm: n[5] || (n[5] = e => a("confirm-action")),
        onDismiss: n[6] || (n[6] = e => a("dismiss-action"))
      }, null, 8, ["execution", "embedded-action", "action-locked", "confirm-loading", "is-streaming", "book-title-fallback", "is-dark-mode"])) : Fr("", !0)])]),
      _: 1
    }, 512), w.value.length >= 3 ? (Cr(), Ir("div", {
      key: 0,
      class: K(["msg-nav-wrap", {
        "is-dark": e.isDarkMode
      }]),
      onMouseenter: D,
      onMouseleave: L
    }, [Qr(Al, {
      name: "msg-nav-pop-fade"
    }, {
      default: Sn(() => [_.value ? (Cr(), Ir("div", {
        key: 0,
        class: "msg-nav-pop",
        onMouseenter: D,
        onMouseleave: L
      }, [zr("div", {
        ref_key: "msgNavListRef",
        ref: R,
        class: "msg-nav-pop-list"
      }, [(Cr(!0), Ir(br, null, Va(G.value, e => (Cr(), Ir("button", {
        key: e.id,
        type: "button",
        class: K(["msg-nav-pop-item", {
          active: e.id === S.value,
          hover: e.id === P.value
        }]),
        "data-msg-id": e.id,
        onMouseenter: t => P.value = e.id,
        onMouseleave: n[7] || (n[7] = e => P.value = null),
        onClick: t => function (e) {
          var t;
          const n = i.value;
          if (!n) return;
          const a = null == (t = n.$el) ? void 0 : t.querySelector(".n-scrollbar-container");
          if (!a) return;
          const o = a.querySelector(`.msg-row[data-msg-id="${e}"]`);
          if (!o) return;
          const r = o.offsetTop - a.clientHeight / 2 + o.offsetHeight / 2;
          n.scrollTo({
            top: Math.max(0, r),
            behavior: "smooth"
          }), o.classList.add("msg-nav-highlight"), setTimeout(() => o.classList.remove("msg-nav-highlight"), 1500), _.value = !1, P.value = null;
        }(e.id)
      }, [n[40] || (n[40] = zr("span", {
        class: "msg-nav-pop-dash"
      }, null, -1)), zr("span", Dnn, Y(e.preview), 1)], 42, Qnn))), 128))], 512)], 32)) : Fr("", !0)]),
      _: 1
    }), zr("div", Lnn, [(Cr(!0), Ir(br, null, Va(q.value, e => (Cr(), Ir("div", {
      key: e.index,
      class: K(["msg-nav-dash", {
        active: e.isCurrent,
        visible: e.isVisible
      }])
    }, null, 2))), 128))])], 34)) : Fr("", !0), zr("div", jnn, [zr("div", Enn, [e.showQuickActions ? (Cr(), Nr(hnn, {
      key: 0,
      disabled: e.isStreaming || e.confirmLoading || !!j.value,
      loading: e.quickActionLoading,
      "is-dark-mode": e.isDarkMode,
      onWriteNext: n[8] || (n[8] = e => a("quick-write-next")),
      onWriteBatch: n[9] || (n[9] = e => a("quick-write-batch")),
      onAudit: n[10] || (n[10] = e => a("quick-audit")),
      onRevise: n[11] || (n[11] = e => a("quick-revise"))
    }, null, 8, ["disabled", "loading", "is-dark-mode"])) : Fr("", !0), zr("div", Fnn, [zr("div", Bnn, [Cn(zr("textarea", {
      ref_key: "textareaRef",
      ref: s,
      "onUpdate:modelValue": n[12] || (n[12] = e => l.value = e),
      class: "float-textarea",
      placeholder: oe.value,
      rows: "2",
      disabled: e.isStreaming,
      onKeydown: le,
      onCompositionstart: n[13] || (n[13] = e => c.value = !0),
      onCompositionend: n[14] || (n[14] = e => c.value = !1)
    }, null, 40, Vnn), [[zi, l.value]])]), zr("div", Unn, [zr("div", Knn, [zr("div", {
      class: "model-trigger",
      onClick: ce
    }, [zr("span", {
      class: "model-trigger-icon",
      innerHTML: Tt(_me)(e.selectedModel)
    }, null, 8, Hnn), zr("span", Jnn, Y(X.value), 1), $.value ? (Cr(), Ir("span", {
      key: 0,
      class: K(["model-channel-tag", "ch-" + ee.value])
    }, Y($.value), 3)) : Fr("", !0), Qr(Tt($I), {
      component: Tt(MH),
      size: 12,
      class: "trigger-chevron"
    }, null, 8, ["component"])]), zr("div", Znn, [zr("button", {
      type: "button",
      class: K(["tb-icon-btn thinking-toggle", {
        active: e.thinkingAutoExpand
      }]),
      title: "展开思考",
      onClick: n[15] || (n[15] = e => f.value = !f.value)
    }, [...(n[41] || (n[41] = [zr("svg", {
      width: "18",
      height: "18",
      viewBox: "0 0 1024 1024",
      fill: "currentColor"
    }, [zr("path", {
      d: "M670.0032 89.6l4.0064 0.0384c104.0512 2.112 188.032 86.1056 192.2176 190.72l0.1408 5.3376 1.7536 1.3184C931.8144 335.2832 972.8 427.1104 972.8 521.088l-0.0512 5.632c-1.152 69.6448-23.808 145.8816-59.4176 194.176l-1.3056 1.728 0.0256 2.4064C912.0512 847.6032 814.0416 947.2 692.8384 947.2c-75.136 0-141.3504-38.272-180.8384-96.5376C472.512 908.928 406.2976 947.2 331.1616 947.2c-121.216 0-219.2128-99.5968-219.2128-222.1568l0.0256-2.4064-1.3056-1.7408c-35.6224-48.2816-58.2528-124.5184-59.4176-194.176L51.2 521.088c0-93.9776 40.9856-185.8048 104.6784-234.0736l1.7536-1.3184 0.0128-1.1904c2.112-106.5344 86.9248-192.7296 192.3456-194.8672L353.9968 89.6c64.8448 0 122.2656 31.808 158.016 80.7936C547.7248 121.408 605.1584 89.6 670.0032 89.6z m-316.0064 76.8c-65.9072 0-119.5904 54.5536-119.5904 122.1632 0 2.9952 0.1024 5.9648 0.32 8.9216l0.384 4.4288c0.128 1.1008 0.192 2.2016 0.2048 3.2896l-0.3968-2.7776 0.512 3.6864a95.488 95.488 0 0 0 25.28 51.8016c9.664 10.2016 20.928 17.408 27.136 19.4176l1.2416 0.3584a38.4 38.4 0 1 1-17.344 74.816c-35.456-8.2176-72.192-38.0928-94.1696-80.1664-29.824 36.608-49.5744 91.904-49.5744 148.7616 0 32.5376 6.336 67.6736 16.6912 98.2016 8.9856-14.4768 19.3536-26.944 30.6048-35.968a38.4 38.4 0 0 1 49.92 58.3168l-1.856 1.5744c-13.184 10.5856-28.4032 41.4208-34.1248 69.76l0.0128-0.256c-0.3328 4.0704-0.512 8.192-0.512 12.3136C188.7488 805.4528 252.6592 870.4 331.1744 870.4 409.6896 870.4 473.6 805.4528 473.6 725.0432L473.6 593.664c-25.4336 21.0048-56.4736 35.712-92.0064 39.7312a38.4 38.4 0 0 1-10.8928-75.9936l2.2272-0.32c55.872-6.336 97.216-67.072 100.6848-116.7488l-0.0256-151.7824C473.6 220.9536 419.904 166.4 353.9968 166.4z m316.0064 0C604.0832 166.4 550.4 220.9536 550.4 288.5632l-0.0256 151.7824c3.4048 48.6912 43.1744 107.9936 97.344 116.3136l3.328 0.4352a38.4 38.4 0 0 1-8.6528 76.3136c-35.5328-4.032-66.56-18.7264-92.0064-39.7312v131.3664C550.4 805.4528 614.3232 870.4 692.8512 870.4c78.5152 0 142.4128-64.9472 142.4128-145.3568 0-4.0448-0.1536-8.064-0.4736-12.0576-5.504-27.1616-19.712-56.6144-32.4864-68.3264l-1.664-1.4336a38.4 38.4 0 1 1 48.0768-59.904c11.264 9.0368 21.6192 21.504 30.6176 35.9552 10.3424-30.5024 16.6784-65.6384 16.6784-98.176 0-56.8448-19.7504-112.1536-49.5872-148.7744-21.952 42.0864-58.7136 71.9616-94.1568 80.1792a38.4 38.4 0 0 1-19.52-74.24l2.176-0.5888c5.824-1.344 18.0224-8.832 28.3776-19.7632 13.9776-14.7456 23.2448-33.4976 25.856-55.9616l-0.4608 3.2512 0.064-1.6384 0.128-1.6512c0.4736-4.4032 0.7168-8.8576 0.7168-13.3504 0-67.6096-53.6832-122.1632-119.5904-122.1632z"
    })], -1)]))], 2), f.value ? (Cr(), Ir("div", {
      key: 0,
      class: "thinking-popover",
      style: E({
        background: ne.value,
        borderColor: ae.value
      })
    }, [zr("div", Xnn, [n[43] || (n[43] = zr("span", {
      class: "thinking-popover-label"
    }, "展开思考", -1)), zr("div", {
      class: K(["thinking-switch", {
        on: e.thinkingAutoExpand
      }]),
      onClick: n[16] || (n[16] = t => {
        a("update:thinkingAutoExpand", !e.thinkingAutoExpand), f.value = !1;
      })
    }, [...(n[42] || (n[42] = [zr("div", {
      class: "thinking-switch-thumb"
    }, null, -1)]))], 2)])], 4)) : Fr("", !0), f.value ? (Cr(), Ir("div", {
      key: 1,
      class: "thinking-popover-mask",
      onClick: n[17] || (n[17] = e => f.value = !1)
    })) : Fr("", !0)]), zr("div", $nn, [Qr(Tt(PI), {
      trigger: "hover",
      placement: "top"
    }, {
      trigger: Sn(() => [zr("button", {
        class: "tb-icon-btn",
        onClick: n[18] || (n[18] = e => h.value = !h.value)
      }, [Qr(Tt($I), {
        component: Tt(e2),
        size: 18
      }, null, 8, ["component"])])]),
      default: Sn(() => [jr(" 上下文长度: " + Y(W.value), 1)]),
      _: 1
    }), h.value ? (Cr(), Ir("div", {
      key: 0,
      class: K(["ctx-popover", {
        "is-dark": e.isDarkMode
      }])
    }, [zr("div", Ynn, [n[44] || (n[44] = zr("span", null, "上下文长度", -1)), zr("span", ean, Y(W.value), 1)]), zr("div", {
      class: "ctx-slider-track",
      onClick: ie
    }, [zr("div", {
      class: "ctx-slider-fill",
      style: E({
        width: y.value + "%"
      })
    }, null, 4), zr("div", {
      class: "ctx-slider-thumb",
      style: E({
        left: y.value + "%"
      })
    }, null, 4)]), zr("div", tan, [zr("span", {
      onClick: n[19] || (n[19] = e => m.value = 1e5)
    }, "100k"), zr("span", {
      onClick: n[20] || (n[20] = e => m.value = 2e5)
    }, "200k"), zr("span", {
      onClick: n[21] || (n[21] = e => m.value = 3e5)
    }, "300k"), zr("span", {
      onClick: n[22] || (n[22] = e => m.value = 4e5)
    }, "400k"), zr("span", {
      onClick: n[23] || (n[23] = e => m.value = 1e6)
    }, "1m"), zr("span", {
      onClick: n[24] || (n[24] = e => m.value = van)
    }, [Qr(Tt($I), {
      component: Tt(JY),
      size: 14
    }, null, 8, ["component"])])])], 2)) : Fr("", !0), h.value ? (Cr(), Ir("div", {
      key: 1,
      class: "ctx-popover-mask",
      onClick: n[25] || (n[25] = e => h.value = !1)
    })) : Fr("", !0)]), Qr(Tt(PI), {
      trigger: "hover",
      placement: "top"
    }, {
      trigger: Sn(() => [zr("button", {
        class: "tb-icon-btn",
        onClick: se
      }, [Qr(Tt($I), {
        component: Tt(V5),
        size: 18
      }, null, 8, ["component"])])]),
      default: Sn(() => [n[45] || (n[45] = jr(" 清除上下文 (Ctrl+K) ", -1))]),
      _: 1
    })]), zr("div", nan, [e.isStreaming ? (Cr(), Ir("button", {
      key: 0,
      type: "button",
      class: "send-btn sending",
      onClick: n[26] || (n[26] = e => a("cancel"))
    }, [...(n[46] || (n[46] = [zr("svg", {
      width: "16",
      height: "16",
      viewBox: "0 0 24 24",
      fill: "currentColor"
    }, [zr("rect", {
      x: "6",
      y: "6",
      width: "12",
      height: "12",
      rx: "2"
    })], -1)]))])) : (Cr(), Ir("button", {
      key: 1,
      type: "button",
      class: K(["send-btn", {
        active: l.value.trim()
      }]),
      disabled: !l.value.trim(),
      onClick: re
    }, [Qr(Tt($I), {
      component: Tt(EU),
      size: 18
    }, null, 8, ["component"])], 10, aan))])])])])])]), Qr(Tt(wq), {
      show: u.value,
      "onUpdate:show": n[29] || (n[29] = e => u.value = e),
      "mask-closable": !0,
      style: {
        width: "480px",
        "max-width": "90vw"
      }
    }, {
      default: Sn(() => [zr("div", {
        class: K(["model-dialog", {
          "is-dark": e.isDarkMode
        }])
      }, [zr("div", oan, [Qr(Tt($I), {
        component: Tt(H4),
        size: 16,
        class: "model-search-icon"
      }, null, 8, ["component"]), Cn(zr("input", {
        "onUpdate:modelValue": n[27] || (n[27] = e => d.value = e),
        class: "model-search-input",
        placeholder: "搜索模型...",
        autofocus: ""
      }, null, 512), [[zi, d.value]]), zr("button", {
        type: "button",
        class: "model-search-close",
        onClick: n[28] || (n[28] = e => u.value = !1)
      }, [Qr(Tt($I), {
        component: Tt(eJ),
        size: 16
      }, null, 8, ["component"])])]), zr("div", ran, [(Cr(!0), Ir(br, null, Va(Tt(p), e => (Cr(), Ir("button", {
        key: e.key,
        type: "button",
        class: K(["model-channel-filter-chip", {
          active: v.value === e.key,
          ["ch-" + e.key]: !0
        }]),
        onClick: t => {
          return n = e.key, void (v.value = v.value === n ? "" : n);
          var n;
        }
      }, Y(e.label), 11, lan))), 128))]), zr("div", ian, [(Cr(!0), Ir(br, null, Va(te.value, t => (Cr(), Ir("div", {
        key: t.value,
        class: K(["model-dialog-item", {
          selected: t.value === e.selectedModel
        }]),
        onClick: e => {
          return n = t.value, a("update:selectedModel", n), void (u.value = !1);
          var n;
        }
      }, [zr("span", {
        class: "model-item-icon",
        innerHTML: Tt(_me)(t.value)
      }, null, 8, can), zr("span", uan, Y(t.displayName || t.label), 1), t.channelTag ? (Cr(), Ir("span", {
        key: 0,
        class: K(["model-channel-tag", "ch-" + (t.channelKey || "")])
      }, Y(t.channelTag), 3)) : Fr("", !0), Qr(Zre, {
        model: t.value,
        "is-dark-mode": e.isDarkMode
      }, null, 8, ["model", "is-dark-mode"]), t.value === e.selectedModel ? (Cr(), Nr(Tt($I), {
        key: 1,
        component: Tt(xH),
        size: 16,
        class: "model-item-check"
      }, null, 8, ["component"])) : Fr("", !0)], 10, san))), 128)), 0 === te.value.length ? (Cr(), Ir("div", dan, "无匹配模型")) : Fr("", !0)])], 2)]),
      _: 1
    }, 8, ["show"])], 2));
  }
};
