// Inspection only. Original SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Original UTF16 offsets [8651068, 8652874); original line 41
// External dependencies and runtime equivalence remain unresolved.
const recoveredComponent = {
  __name: "NovelPanelArtifactView",
  props: {
    title: {
      type: String,
      default: ""
    },
    content: {
      type: String,
      default: ""
    },
    loading: Boolean,
    saving: Boolean,
    readOnly: Boolean,
    autoEdit: Boolean,
    isDarkMode: Boolean
  },
  emits: ["back", "save"],
  setup(e, {
    emit: t
  }) {
    const n = e,
      a = t,
      o = Nt(!1),
      r = Nt("");
    function l() {
      r.value = n.content || "", o.value = !0;
    }
    function i() {
      r.value = n.content || "", o.value = !1;
    }
    function s() {
      a("save", r.value);
    }
    return Ro(() => n.content, e => {
      o.value || (r.value = e || ""), !n.autoEdit || null === e || n.loading || n.readOnly || (o.value = !0);
    }, {
      immediate: !0
    }), Ro(() => n.saving, (e, t) => {
      t && !e && (o.value = !1);
    }), (t, n) => (Cr(), Ir("div", {
      class: K(["artifact-view", {
        dark: e.isDarkMode
      }])
    }, [zr("header", man, [zr("button", {
      type: "button",
      class: "icon-btn",
      title: "返回",
      onClick: n[0] || (n[0] = e => a("back"))
    }, [Qr(Tt($I), {
      component: Tt(yU),
      size: 16
    }, null, 8, ["component"])]), zr("span", han, Y(e.title || "资料详情"), 1), zr("div", gan, [e.readOnly || e.loading || null === e.content ? Fr("", !0) : (Cr(), Ir(br, {
      key: 0
    }, [o.value ? (Cr(), Ir(br, {
      key: 1
    }, [zr("button", {
      type: "button",
      class: "icon-btn save",
      title: "保存",
      disabled: e.saving,
      onClick: s
    }, [e.saving ? (Cr(), Nr(Tt(VE), {
      key: 0,
      size: 14
    })) : (Cr(), Nr(Tt($I), {
      key: 1,
      component: Tt(xH),
      size: 15
    }, null, 8, ["component"]))], 8, Wan), zr("button", {
      type: "button",
      class: "icon-btn",
      title: "取消",
      onClick: i
    }, [Qr(Tt($I), {
      component: Tt(eJ),
      size: 15
    }, null, 8, ["component"])])], 64)) : (Cr(), Ir("button", {
      key: 0,
      type: "button",
      class: "icon-btn",
      title: "编辑",
      onClick: l
    }, [Qr(Tt($I), {
      component: Tt($J),
      size: 15
    }, null, 8, ["component"])]))], 64))])]), Qr(Tt(IE), {
      class: "artifact-scroll"
    }, {
      default: Sn(() => [e.loading ? (Cr(), Ir("div", yan, [Qr(Tt(VE), {
        size: 18
      })])) : null === e.content ? (Cr(), Ir("p", ban, "文件不存在")) : o.value ? Cn((Cr(), Ir("textarea", {
        key: 2,
        "onUpdate:modelValue": n[1] || (n[1] = e => r.value = e),
        class: "artifact-editor",
        spellcheck: "false"
      }, null, 512)), [[zi, r.value]]) : (Cr(), Ir("pre", kan, Y(e.content), 1))]),
      _: 1
    })], 2));
  }
};
