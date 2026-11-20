// Inspection only. Original SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Original UTF16 offsets [8619062, 8625267); original line 41
// External dependencies and runtime equivalence remain unresolved.
const recoveredComponent = {
  __name: "NovelToolSteps",
  props: {
    execution: {
      type: Object,
      default: null
    },
    bookTitleFallback: {
      type: String,
      default: ""
    },
    isDarkMode: {
      type: Boolean,
      default: !1
    },
    embeddedAction: {
      type: Object,
      default: null
    },
    actionLocked: Boolean,
    actionRejected: Boolean,
    confirmLoading: Boolean,
    isStreaming: Boolean,
    streamingText: {
      type: String,
      default: ""
    },
    streamingReasoning: {
      type: String,
      default: ""
    },
    streamStage: {
      type: String,
      default: ""
    },
    hideStreamOutput: Boolean
  },
  emits: ["retry", "confirm", "dismiss"],
  setup(e, {
    emit: t
  }) {
    const n = e,
      a = t,
      o = Nt(!1),
      r = Nt(0);
    let l = null;
    const i = ml(() => Ntn(n.execution)),
      s = ml(() => {
        var e;
        return "propose_action" === ((null == (e = i.value) ? void 0 : e.toolName) || "") || Boolean(n.embeddedAction);
      }),
      c = ml(() => {
        var e;
        const t = null == (e = i.value) ? void 0 : e.status;
        return "running" === t || "processing" === t ? "执行中" : "completed" === t ? "已完成" : "error" === t ? "失败" : "等待中";
      }),
      u = ml(() => {
        var e;
        const t = null == (e = i.value) ? void 0 : e.status;
        return "running" === t || "processing" === t ? "running" : "error" === t ? "error" : "completed" === t ? "completed" : "pending";
      }),
      d = ml(() => {
        const e = i.value;
        return e ? e.label ? e.label : Itn(e.toolName, e.agent) : "任务";
      }),
      v = ml(() => {
        var e;
        const t = (null == (e = i.value) ? void 0 : e.bookTitle) || n.bookTitleFallback;
        return t ? ` · ${t}` : "";
      }),
      p = ml(() => function (e) {
        const t = Math.max(0, Number(e) || 0);
        if (t < 60) return `${t}s`;
        const n = Math.floor(t / 60),
          a = t % 60;
        return a ? `${n}m ${a}s` : `${n}m`;
      }(r.value)),
      f = ml(() => {
        var e;
        return (null == (e = i.value) ? void 0 : e.logs) || [];
      }),
      m = ml(() => {
        var e;
        return n.streamStage || (null == (e = i.value) ? void 0 : e.streamStage) || "";
      }),
      h = ml(() => {
        var e;
        return n.streamingText || (null == (e = i.value) ? void 0 : e.streamText) || "";
      }),
      g = ml(() => {
        var e;
        return n.streamingReasoning || (null == (e = i.value) ? void 0 : e.streamReasoning) || "";
      }),
      W = ml(() => !n.hideStreamOutput && (!!n.isStreaming || Boolean(h.value || g.value))),
      y = ml(() => {
        var e;
        return (null == (e = i.value) ? void 0 : e.stages) || [];
      }),
      b = ml(() => {
        var e;
        const t = n.execution,
          a = (null == t ? void 0 : t.details) && "object" == typeof t.details ? t.details : {};
        return (null == (e = i.value) ? void 0 : e.coverImagePath) || a.coverImagePath || "";
      }),
      k = ml(() => {
        var e, t;
        return ("generate_cover" === (null == (e = i.value) ? void 0 : e.toolName) || "generate_cover" === (null == (t = n.execution) ? void 0 : t.toolName)) && Boolean(b.value);
      });
    function x(e) {
      if (!e || "object" != typeof e) return "";
      const t = Math.round((e.elapsedMs || 0) / 1e3);
      return ["thinking" === e.status ? "思考中" : e.status || "", t ? `${t}s` : "", e.totalChars > 0 ? e.chineseChars > 0 ? `${e.totalChars}字` : `${e.totalChars} chars` : ""].filter(Boolean).join(" · ");
    }
    function w(e) {
      const t = String(e || "");
      return t.startsWith("[error]") || /\[error\]/i.test(t) ? "log-error" : t.startsWith("[warning]") || /\[warning\]|警告/i.test(t) ? "log-warn" : "log-normal";
    }
    return Ro(() => n.execution, function () {
      l && (clearInterval(l), l = null);
      const e = i.value;
      if (e) if ("error" === e.status && (o.value = !1), "running" !== e.status && "processing" !== e.status || !e.startedAt) e.startedAt && e.completedAt ? r.value = Math.max(0, Math.round((e.completedAt - e.startedAt) / 1e3)) : e.startedAt && (r.value = Math.max(0, Math.round((Date.now() - e.startedAt) / 1e3)));else {
        const t = () => {
          r.value = Math.max(0, Math.round((Date.now() - e.startedAt) / 1e3));
        };
        t(), l = setInterval(t, 1e3);
      }
    }, {
      deep: !0,
      immediate: !0
    }), Ro(() => {
      var e;
      return null == (e = i.value) ? void 0 : e.status;
    }, (e, t) => {
      "running" !== e && "processing" !== e ? ("completed" !== e || "completed" === t || s.value || setTimeout(() => {
        o.value = !0;
      }, 500), "error" === e && (o.value = !1)) : o.value = !1;
    }, {
      immediate: !0
    }), Ga(() => {
      l && clearInterval(l);
    }), (t, n) => i.value ? (Cr(), Ir("div", {
      key: 0,
      class: K(["tool-card", [u.value, {
        dark: e.isDarkMode
      }]])
    }, [zr("button", {
      type: "button",
      class: "tool-header",
      onClick: n[0] || (n[0] = e => o.value = !o.value)
    }, [zr("span", Qtn, [jr(Y(d.value), 1), v.value ? (Cr(), Ir("span", Dtn, Y(v.value), 1)) : Fr("", !0)]), zr("div", Ltn, [p.value ? (Cr(), Ir("span", jtn, Y(p.value), 1)) : Fr("", !0), zr("span", {
      class: K(["tool-status-badge", u.value])
    }, ["running" === u.value ? (Cr(), Nr(Tt(VE), {
      key: 0,
      size: 12
    })) : "completed" === u.value ? (Cr(), Nr(Tt($I), {
      key: 1,
      component: Tt(vH),
      size: 12
    }, null, 8, ["component"])) : "error" === u.value ? (Cr(), Nr(Tt($I), {
      key: 2,
      component: Tt(JH),
      size: 12
    }, null, 8, ["component"])) : Fr("", !0), zr("span", null, Y(c.value), 1)], 2), Qr(Tt($I), {
      component: Tt(MH),
      size: 16,
      class: K(["chevron", {
        open: !o.value
      }])
    }, null, 8, ["component", "class"])])]), e.embeddedAction ? (Cr(), Nr(atn, {
      key: 0,
      embedded: "",
      action: e.embeddedAction,
      locked: e.actionLocked,
      rejected: e.actionRejected,
      loading: e.confirmLoading,
      streaming: e.isStreaming,
      onConfirm: n[1] || (n[1] = e => a("confirm")),
      onDismiss: n[2] || (n[2] = e => a("dismiss"))
    }, null, 8, ["action", "locked", "rejected", "loading", "streaming"])) : Fr("", !0), Cn(zr("div", Etn, [W.value ? (Cr(), Ir("div", Ftn, [m.value ? (Cr(), Ir("div", Btn, Y(m.value), 1)) : Fr("", !0), g.value ? (Cr(), Ir("div", Vtn, Y(g.value), 1)) : Fr("", !0), zr("div", Utn, [h.value ? (Cr(), Ir(br, {
      key: 0
    }, [jr(Y(h.value), 1)], 64)) : e.isStreaming ? (Cr(), Ir("span", Ktn, "AI 正在生成…")) : Fr("", !0), e.isStreaming ? (Cr(), Ir("span", Htn, "|")) : Fr("", !0)])])) : Fr("", !0), y.value.length ? (Cr(), Ir("ol", Jtn, [(Cr(!0), Ir(br, null, Va(y.value, (e, t) => (Cr(), Ir("li", {
      key: t,
      class: K(["stage-row", e.status || "pending"])
    }, [zr("span", Ztn, ["active" === e.status ? (Cr(), Nr(Tt(VE), {
      key: 0,
      size: 14
    })) : "completed" === e.status ? (Cr(), Nr(Tt($I), {
      key: 1,
      component: Tt(vH),
      size: 14,
      class: "stage-ok"
    }, null, 8, ["component"])) : (Cr(), Ir("span", Xtn, "○"))]), zr("div", $tn, [zr("div", Ytn, Y(e.name || e.stage || e.label), 1), e.progress ? (Cr(), Ir("div", enn, Y(x(e.progress)), 1)) : e.message ? (Cr(), Ir("div", tnn, Y(e.message), 1)) : Fr("", !0)])], 2))), 128))])) : Fr("", !0), f.value.length ? (Cr(), Ir("ul", nnn, [(Cr(!0), Ir(br, null, Va(f.value, (e, t) => (Cr(), Ir("li", {
      key: t,
      class: K(["log-line", w(e)])
    }, Y(e), 3))), 128))])) : Fr("", !0), i.value.error ? (Cr(), Ir("div", ann, [jr(Y(i.value.error) + " ", 1), "error" === i.value.status ? (Cr(), Ir("button", {
      key: 0,
      type: "button",
      class: "retry-btn",
      onClick: n[3] || (n[3] = Zi(e => a("retry"), ["stop"]))
    }, "重试")) : Fr("", !0)])) : "error" === i.value.status && i.value.summary ? (Cr(), Ir("div", onn, [jr(Y(i.value.summary) + " ", 1), zr("button", {
      type: "button",
      class: "retry-btn",
      onClick: n[4] || (n[4] = Zi(e => a("retry"), ["stop"]))
    }, "重试")])) : Fr("", !0), k.value ? (Cr(), Nr(mtn, {
      key: 5,
      "image-path": b.value,
      alt: "封面预览"
    }, null, 8, ["image-path"])) : Fr("", !0), i.value.coverError && "generate_cover" === i.value.toolName ? (Cr(), Ir("p", rnn, Y(i.value.coverError), 1)) : Fr("", !0)], 512), [[Xl, !o.value]]), i.value.summary && "running" !== i.value.status && !e.embeddedAction ? (Cr(), Ir("p", lnn, Y(i.value.summary), 1)) : Fr("", !0)], 2)) : Fr("", !0);
  }
};
