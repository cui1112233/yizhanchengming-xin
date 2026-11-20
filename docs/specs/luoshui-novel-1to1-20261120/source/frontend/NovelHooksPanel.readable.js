// Inspection only. Original SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Original UTF16 offsets [8849652, 8851209); original line 41
// External dependencies and runtime equivalence remain unresolved.
const recoveredComponent = {
  __name: "NovelHooksPanel",
  props: {
    hooks: {
      type: Array,
      default: () => []
    },
    currentChapter: {
      type: Number,
      default: 0
    },
    isDarkMode: Boolean
  },
  setup(e) {
    const t = e,
      n = Nt("active"),
      a = ml(() => "active" === n.value ? t.hooks.filter(e => "planted" === e.status || "advancing" === e.status) : "resolved" === n.value ? t.hooks.filter(e => "resolved" === e.status) : t.hooks);
    function o(e) {
      if ("resolved" === e.status) return "success";
      const n = e.last_advanced_chapter || e.start_chapter,
        a = t.currentChapter - n;
      return a >= 12 ? "error" : a >= 8 ? "warning" : "success";
    }
    return (t, r) => (Cr(), Ir("div", {
      class: K(["hooks-panel", {
        dark: e.isDarkMode
      }])
    }, [zr("div", Ycn, [(Cr(), Ir(br, null, Va(["active", "resolved", "all"], e => zr("button", {
      key: e,
      type: "button",
      class: K(["filter-btn", {
        active: n.value === e
      }]),
      onClick: t => n.value = e
    }, Y({
      active: "活跃",
      resolved: "已解决",
      all: "全部"
    }[e]), 11, eun)), 64))]), a.value.length ? (Cr(), Ir("div", tun, [(Cr(!0), Ir(br, null, Va(a.value, e => (Cr(), Ir("div", {
      key: e.hook_id,
      class: "hook-card"
    }, [zr("div", nun, [zr("span", aun, Y(e.title || e.hook_id), 1), Qr(Tt(cw), {
      type: o(e),
      size: "small"
    }, {
      default: Sn(() => {
        return [jr(Y((t = e.status, {
          planted: "已植入",
          advancing: "推进中",
          resolved: "已解决"
        }[t] || t)), 1)];
        var t;
      }),
      _: 2
    }, 1032, ["type"])]), zr("div", oun, [zr("span", null, "起始 ch" + Y(e.start_chapter), 1), e.last_advanced_chapter ? (Cr(), Ir("span", run, "最近推进 ch" + Y(e.last_advanced_chapter), 1)) : Fr("", !0), e.advanced_count ? (Cr(), Ir("span", lun, "推进 " + Y(e.advanced_count) + " 次", 1)) : Fr("", !0), "plot" !== e.type ? (Cr(), Ir("span", iun, Y(e.type), 1)) : Fr("", !0)]), e.description ? (Cr(), Ir("p", sun, Y(e.description), 1)) : Fr("", !0)]))), 128))])) : (Cr(), Nr(Tt(Sx), {
      key: 0,
      description: "暂无伏笔记录"
    }))], 2));
  }
};
