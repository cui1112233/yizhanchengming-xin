// Inspection only. Original SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Original UTF16 offsets [8851662, 8853325); original line 41
// External dependencies and runtime equivalence remain unresolved.
const recoveredComponent = {
  __name: "NovelDashboard",
  props: {
    chapters: {
      type: Array,
      default: () => []
    },
    isDarkMode: Boolean
  },
  setup(e) {
    const t = e,
      n = ml(() => t.chapters.map(e => e.wordCount || 0)),
      a = ml(() => Math.max(1, ...n.value)),
      o = ml(() => n.value.length ? Math.round(n.value.reduce((e, t) => e + t, 0) / n.value.length) : 0),
      r = ml(() => n.value.reduce((e, t) => e + t, 0)),
      l = ml(() => {
        const e = {};
        for (const n of t.chapters) {
          const t = n.status || "draft";
          e[t] = (e[t] || 0) + 1;
        }
        return e;
      }),
      i = {
        draft: "草稿",
        ready: "通过",
        "ready-for-review": "待审",
        "audit-failed": "未过",
        "state-degraded": "降级"
      };
    return (t, s) => (Cr(), Ir("div", {
      class: K(["dashboard", {
        dark: e.isDarkMode
      }])
    }, [e.chapters.length ? (Cr(), Ir(br, {
      key: 1
    }, [zr("div", uun, [zr("div", dun, [zr("div", vun, Y(e.chapters.length), 1), s[0] || (s[0] = zr("div", {
      class: "stat-label"
    }, "总章数", -1))]), zr("div", pun, [zr("div", fun, Y((r.value / 1e4).toFixed(1)) + "万", 1), s[1] || (s[1] = zr("div", {
      class: "stat-label"
    }, "总字数", -1))]), zr("div", mun, [zr("div", hun, Y(o.value), 1), s[2] || (s[2] = zr("div", {
      class: "stat-label"
    }, "平均字数", -1))])]), zr("div", gun, [s[3] || (s[3] = zr("h4", null, "章节状态分布", -1)), zr("div", Wun, [(Cr(!0), Ir(br, null, Va(l.value, (t, n) => (Cr(), Ir("div", {
      key: n,
      class: "status-row"
    }, [zr("span", yun, Y(i[n] || n), 1), zr("div", bun, [zr("div", {
      class: K(["bar-fill", n]),
      style: E({
        width: t / e.chapters.length * 100 + "%"
      })
    }, null, 6)]), zr("span", kun, Y(t), 1)]))), 128))])]), zr("div", xun, [s[5] || (s[5] = zr("h4", null, "章节字数趋势", -1)), zr("div", wun, [zr("div", _un, [(Cr(!0), Ir(br, null, Va(n.value, (e, t) => (Cr(), Ir("div", {
      key: t,
      class: "chart-bar",
      style: E({
        height: e / a.value * 100 + "%"
      }),
      title: `第 ${t + 1} 章：${e} 字`
    }, null, 12, Sun))), 128))]), zr("div", Cun, [s[4] || (s[4] = zr("span", null, "1", -1)), zr("span", null, Y(e.chapters.length), 1)])])])], 64)) : (Cr(), Nr(Tt(Sx), {
      key: 0,
      description: "暂无章节数据"
    }))], 2));
  }
};
