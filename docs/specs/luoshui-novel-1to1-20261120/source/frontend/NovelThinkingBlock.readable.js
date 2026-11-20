// Inspection only. Original SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Original UTF16 offsets [8625439, 8626316); original line 41
// External dependencies and runtime equivalence remain unresolved.
const recoveredComponent = {
  __name: "NovelThinkingBlock",
  props: {
    seconds: {
      type: Number,
      default: 0
    },
    reasoning: {
      type: String,
      default: ""
    },
    defaultOpen: {
      type: Boolean,
      default: !1
    },
    thinkingMode: {
      type: Boolean,
      default: !1
    }
  },
  emits: ["update:expanded"],
  setup(e, {
    emit: t
  }) {
    const n = e,
      a = t,
      o = Nt(Boolean(n.defaultOpen));
    Ro(() => n.defaultOpen, e => {
      e && (o.value = !0);
    });
    const r = ml(() => (n.seconds || 0) > 0 || Boolean(String(n.reasoning || "").trim())),
      l = ml(() => String(n.reasoning || "").trim());
    function i() {
      r.value && (o.value = !o.value, a("update:expanded", o.value));
    }
    return (t, n) => r.value ? (Cr(), Ir("div", cnn, [zr("button", {
      type: "button",
      class: "thinking-block",
      onClick: i
    }, [Qr(Tt($I), {
      component: Tt(MH),
      size: 14,
      class: K(["chevron", {
        open: o.value
      }])
    }, null, 8, ["component", "class"]), zr("span", null, "思考了 " + Y(e.seconds || 1) + " 秒", 1)]), o.value ? (Cr(), Ir("div", unn, [l.value ? (Cr(), Ir("pre", dnn, Y(l.value), 1)) : (Cr(), Ir("p", vnn, "暂无思考内容记录"))])) : Fr("", !0)])) : Fr("", !0);
  }
};
