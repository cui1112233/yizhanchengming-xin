// Inspection only. Original SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Original UTF16 offsets [8550523, 8552978); original line 41
// External dependencies and runtime equivalence remain unresolved.
const recoveredComponent = {
  __name: "NovelConfirmCard",
  props: {
    action: {
      type: Object,
      required: !0
    },
    loading: Boolean,
    locked: Boolean,
    embedded: Boolean,
    streaming: Boolean,
    rejected: Boolean
  },
  emits: ["confirm", "dismiss"],
  setup(e, {
    emit: t
  }) {
    const n = e,
      a = t,
      o = new Set(["style_imitation", "fanfic_init", "spinoff_create", "continuation_import"]),
      r = ml(() => {
        var e;
        return o.has(null == (e = n.action) ? void 0 : e.action);
      }),
      l = ml(() => {
        const e = n.action;
        if (!e) return "确认执行";
        if (e.title) return e.title;
        return {
          create_book: "确认创建作品",
          write_next: "确认写下一章",
          audit_chapter: "确认审稿",
          revise_chapter: "确认修订章节",
          short_run: "确认生成短篇",
          play_start: "确认启动互动世界",
          generate_cover: "确认生成封面",
          style_imitation: "打开仿写入口",
          fanfic_init: "打开同人创作",
          spinoff_create: "打开番外创作",
          continuation_import: "打开续写导入"
        }[e.action] || "确认执行";
      }),
      i = ml(() => {
        var e, t;
        const a = String((null == (e = n.action) ? void 0 : e.summary) || "").trim(),
          o = String((null == (t = n.action) ? void 0 : t.instruction) || "").trim();
        return a && a !== o ? a : "";
      }),
      s = ml(() => {
        var e;
        return String((null == (e = n.action) ? void 0 : e.instruction) || "").trim();
      }),
      c = ml(() => {
        var e, t;
        const a = n.action;
        if (!a || "play_start" !== a.action) return [];
        const o = [],
          r = String(a.worldContract || (null == (e = a.playStart) ? void 0 : e.worldContract) || "").trim(),
          l = String(a.visualContract || (null == (t = a.playStart) ? void 0 : t.visualContract) || "").trim();
        return r && o.push({
          label: "世界契约",
          value: r
        }), l && o.push({
          label: "视觉契约",
          value: l
        }), o;
      }),
      u = ml(() => {
        var e;
        return n.streaming || n.loading ? "执行中…" : r.value || (null == (e = n.action) ? void 0 : e.targetRoute) ? "打开入口" : "继续执行";
      }),
      d = ml(() => {
        var e;
        return r.value || (null == (e = n.action) ? void 0 : e.targetRoute) ? "已打开" : "已执行";
      });
    return (t, n) => (Cr(), Ir("div", {
      class: K(["confirm-card", {
        locked: e.locked,
        embedded: e.embedded,
        dark: !1
      }])
    }, [zr("div", Ven, Y(l.value), 1), i.value ? (Cr(), Ir("div", Uen, Y(i.value), 1)) : Fr("", !0), s.value ? (Cr(), Ir("div", Ken, Y(s.value), 1)) : Fr("", !0), c.value.length ? (Cr(), Ir("div", Hen, [(Cr(!0), Ir(br, null, Va(c.value, e => (Cr(), Ir("div", {
      key: e.label,
      class: "contract-row"
    }, [zr("div", Jen, Y(e.label), 1), zr("div", Zen, Y(e.value), 1)]))), 128))])) : Fr("", !0), e.locked ? (Cr(), Ir("div", Xen, [Qr(Tt($I), {
      component: Tt(xH),
      size: 15,
      class: "locked-icon"
    }, null, 8, ["component"]), zr("span", null, Y(d.value), 1)])) : e.rejected ? (Cr(), Ir("div", $en, "已取消")) : (Cr(), Ir("div", Yen, [zr("button", {
      type: "button",
      class: "confirm-btn confirm-btn-primary",
      disabled: e.loading || e.streaming,
      onClick: n[0] || (n[0] = e => a("confirm"))
    }, Y(u.value), 9, etn), zr("button", {
      type: "button",
      class: "confirm-btn confirm-btn-secondary",
      disabled: e.loading || e.streaming,
      onClick: n[1] || (n[1] = e => a("dismiss"))
    }, "取消", 8, ttn)]))], 2));
  }
};
