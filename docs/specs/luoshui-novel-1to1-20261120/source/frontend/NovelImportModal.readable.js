// Inspection only. Original SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Original UTF16 offsets [8830352, 8831305); original line 41
// External dependencies and runtime equivalence remain unresolved.
const recoveredComponent = {
  __name: "NovelImportModal",
  props: {
    show: Boolean,
    tab: {
      type: String,
      default: "imitation"
    },
    sessionId: {
      type: String,
      default: ""
    },
    styleSources: {
      type: Array,
      default: () => []
    },
    books: {
      type: Array,
      default: () => []
    },
    isDarkMode: Boolean
  },
  emits: ["update:show", "update:tab", "created", "submit-start"],
  setup(e, {
    emit: t
  }) {
    const n = t;
    function a() {
      n("update:show", !1);
    }
    function o(e) {
      n("created", e), a();
    }
    return (t, r) => (Cr(), Nr(Tt(wq), {
      show: e.show,
      "auto-focus": !1,
      "mask-closable": !0,
      "transform-origin": "center",
      "onUpdate:show": r[2] || (r[2] = e => n("update:show", e))
    }, {
      default: Sn(() => [zr("div", {
        class: K(["import-modal-card", {
          dark: e.isDarkMode
        }])
      }, [Qr(Fsn, {
        variant: "modal",
        tab: e.tab,
        "session-id": e.sessionId,
        "style-sources": e.styleSources,
        books: e.books,
        "is-dark-mode": e.isDarkMode,
        "onUpdate:tab": r[0] || (r[0] = e => n("update:tab", e)),
        onClose: a,
        onCreated: o,
        onSubmitStart: r[1] || (r[1] = e => n("submit-start", e))
      }, null, 8, ["tab", "session-id", "style-sources", "books", "is-dark-mode"])], 2)]),
      _: 1
    }, 8, ["show"]));
  }
};
