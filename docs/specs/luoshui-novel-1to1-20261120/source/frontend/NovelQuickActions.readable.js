// Inspection only. Original SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Original UTF16 offsets [8626774, 8627496); original line 41
// External dependencies and runtime equivalence remain unresolved.
const recoveredComponent = {
  __name: "NovelQuickActions",
  props: {
    disabled: Boolean,
    loading: Boolean,
    isDarkMode: Boolean
  },
  emits: ["write-next", "audit", "revise", "write-batch"],
  setup(e, {
    emit: t
  }) {
    const n = t,
      a = [{
        key: "write-next",
        label: "写下一章",
        primary: !0
      }, {
        key: "write-batch",
        label: "批量写作",
        primary: !1
      }, {
        key: "audit",
        label: "审稿",
        primary: !1
      }, {
        key: "revise",
        label: "修订",
        primary: !1
      }];
    return (t, o) => (Cr(), Ir("div", {
      class: K(["quick-actions", {
        dark: e.isDarkMode
      }])
    }, [(Cr(), Ir(br, null, Va(a, t => zr("button", {
      key: t.key,
      type: "button",
      class: K(["quick-btn", {
        primary: t.primary
      }]),
      disabled: e.disabled || e.loading,
      onClick: e => {
        var a;
        "write-next" === (a = t.key) ? n("write-next") : "write-batch" === a ? n("write-batch") : "audit" === a ? n("audit") : "revise" === a && n("revise");
      }
    }, Y(t.label), 11, mnn)), 64))], 2));
  }
};
