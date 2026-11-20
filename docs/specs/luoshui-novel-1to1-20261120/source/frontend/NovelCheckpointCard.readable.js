// Inspection only. Original SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Original UTF16 offsets [9221037, 9221822); original line 56
// External dependencies and runtime equivalence remain unresolved.
const recoveredComponent = {
  __name: "NovelCheckpointCard",
  props: {
    checkpoint: {
      type: Object,
      required: !0
    },
    loading: Boolean
  },
  emits: ["confirm", "dismiss"],
  setup(e, {
    emit: t
  }) {
    const n = e,
      a = t,
      o = ml(() => {
        var e;
        const t = null == (e = n.checkpoint) ? void 0 : e.checkpoint;
        return "blueprint" === t ? "确认创作蓝图" : "foundation" === t ? "确认基础设定" : "确认继续生成";
      }),
      r = ml(() => {
        var e;
        return (null == (e = n.checkpoint) ? void 0 : e.message) || "内容已生成，可在右侧修改本地文件，确认后继续下一步。";
      }),
      l = ml(() => n.loading ? "继续中…" : "确认继续");
    return (t, n) => (Cr(), Ir("div", OWn, [zr("div", MWn, Y(o.value), 1), zr("p", IWn, Y(r.value), 1), zr("div", NWn, [zr("button", {
      type: "button",
      class: "checkpoint-btn primary",
      disabled: e.loading,
      onClick: n[0] || (n[0] = e => a("confirm"))
    }, Y(l.value), 9, GWn), zr("button", {
      type: "button",
      class: "checkpoint-btn secondary",
      disabled: e.loading,
      onClick: n[1] || (n[1] = e => a("dismiss"))
    }, " 稍后 ", 8, qWn)])]));
  }
};
