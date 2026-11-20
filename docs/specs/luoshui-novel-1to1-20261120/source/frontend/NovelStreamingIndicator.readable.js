// Inspection only. Original SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Original UTF16 offsets [8626359, 8626704); original line 41
// External dependencies and runtime equivalence remain unresolved.
const recoveredComponent = {
  __name: "NovelStreamingIndicator",
  props: {
    active: {
      type: Boolean,
      default: !0
    },
    isDarkMode: Boolean
  },
  setup: e => (t, n) => e.active ? (Cr(), Ir("div", {
    key: 0,
    class: K(["thinking-stream", {
      dark: e.isDarkMode
    }])
  }, [...(n[0] || (n[0] = [zr("span", {
    class: "thinking-label"
  }, "思考中……", -1), zr("div", {
    class: "progress-track"
  }, [zr("div", {
    class: "progress-bar"
  })], -1)]))], 2)) : Fr("", !0)
};
