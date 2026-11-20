// Inspection only. Original SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Original UTF16 offsets [8542215, 8550130); original line 41
// External dependencies and runtime equivalence remain unresolved.
const recoveredComponent = {
  __name: "NovelStudioSidebar",
  props: {
    books: {
      type: Array,
      default: () => []
    },
    projectSessions: {
      type: Array,
      default: () => []
    },
    modeSessions: {
      type: Array,
      default: () => []
    },
    sessionProgress: {
      type: Object,
      default: () => ({})
    },
    activeWorkflow: {
      type: String,
      default: "book-create"
    },
    activeSessionId: String,
    activeBookId: String,
    activeMode: String,
    activeTool: String,
    expandedBooks: {
      type: Object,
      default: () => ({})
    },
    bookSessions: {
      type: Object,
      default: () => ({})
    },
    isDarkMode: Boolean
  },
  emits: ["create-session", "open-session", "open-book", "toggle-book", "create-book-session", "rename-session", "delete-session", "rename-book", "delete-book", "export-book", "launch-mode", "switch-workflow", "open-import", "open-style", "open-import-modal", "open-imitation-short", "open-agent-settings", "open-scheduler", "open-dashboard", "open-memo"],
  setup(e, {
    emit: t
  }) {
    const n = e,
      a = t,
      o = Nt(!0),
      r = Nt(!0);
    Nt(!1);
    const l = Nt(localStorage.getItem("novel_book_filter") || "all"),
      i = ml(() => ren(n.books, l.value));
    ml(() => {
      var e;
      return (null == (e = oen.find(e => e.key === l.value)) ? void 0 : e.label) || "全部";
    });
    const s = ml(() => {
      var e;
      const t = (null == (e = n.books) ? void 0 : e.length) || 0,
        a = i.value.length;
      return "all" !== l.value && a !== t ? `${a}/${t}` : String(t);
    });
    function c(e) {
      var t, a;
      return Boolean(null == (a = null == (t = n.sessionProgress) ? void 0 : t[e]) ? void 0 : a.running);
    }
    const u = [{
        key: "book-create",
        label: "长篇原创",
        icon: "novel"
      }, {
        key: "short",
        label: "短篇原创",
        icon: "short"
      }, {
        key: "imitation",
        label: "仿写长篇",
        icon: "imitation"
      }, {
        key: "imitation-short",
        label: "仿写短篇",
        icon: "imitation"
      }, {
        key: "spinoff",
        label: "番外创作",
        icon: "spinoff"
      }, {
        key: "chapters",
        label: "续写创作",
        icon: "chapters"
      }],
      d = [{
        key: "style",
        label: "写作风格",
        icon: "style-settings"
      }, {
        key: "import",
        label: "导入素材",
        icon: "import"
      }, {
        key: "agent-settings",
        label: "Agent 设置",
        icon: "agent"
      }, {
        key: "scheduler",
        label: "定时写作",
        icon: "agent"
      }, {
        key: "dashboard",
        label: "数据看板",
        icon: "agent"
      }, {
        key: "memo",
        label: "备忘录",
        icon: "memo"
      }];
    function v(e) {
      return function (e) {
        const t = "bRb2",
          n = 172,
          a = "C$d9",
          o = 507,
          r = "dzqp",
          l = 349,
          i = 254,
          s = "#Z33",
          c = "TH@d",
          u = 185,
          d = 402,
          v = 551,
          p = 450,
          f = "FjU*",
          m = "&YdT",
          h = 108,
          g = 192,
          W = 371,
          y = "hCkC",
          b = 454,
          k = "grae",
          x = 503,
          w = "SE$K",
          _ = 170,
          S = 202,
          C = 240,
          P = "GP!m",
          R = "a7rx",
          O = 21,
          M = 515,
          I = 539,
          N = "#Z33",
          G = 28,
          q = "bXiq",
          A = 230,
          T = 142,
          z = 544,
          Q = "Iz*G",
          D = "[#!c",
          L = 82,
          j = "Dqz^",
          E = 350,
          F = "$f@l",
          B = 302,
          V = "&Jot",
          U = 450,
          K = 231,
          H = 605,
          J = "o%nT",
          Z = 61,
          X = 330,
          $ = "0z1%",
          Y = 387,
          ee = 595,
          te = "FjU*",
          ne = {};
        function ae(e, t, n, a) {
          return U9t(0, 0, n - -558, e);
        }
        ne[ae("dzqp", 150, 253)] = "6|2|0|5|1|" + ae(t, 281, n), ne[ae(a, o, 382)] = function (e, t) {
          return e === t;
        }, ne[ae(r, l, i)] = ae(s, -134, -20) + "c", ne[ae(c, u, d)] = re(v, p, f) + "tion", ne.rsxvS = function (e, t) {
          return e === t;
        }, ne[ae(m, h, 190)] = re(g, W, y) + re(557, b, k) + re(x, 579, w), ne.CezNr = "chapt" + ae("OrOs", 0, _), ne[re(S, C, P)] = "novel", ne[ae(R, 0, O)] = re(M, I, P) + "ff", ne[ae(N, 0, G)] = ae(q, 0, A);
        const oe = ne;
        function re(e, t, n, a) {
          return U9t(0, 0, t - -285, n);
        }
        const le = oe.nCRuA[ae("1JU(", 0, T)]("|");
        let ie = 0;
        for (;;) {
          switch (le[ie++]) {
            case "0":
              if (oe[re(0, 561, "V1[G")](e, oe[re(0, 283, c)])) return oe[re(0, z, Q)];
              continue;
            case "1":
              if (oe[ae(D, 0, 393)](e, oe[ae(P, 0, L)])) return "imita" + ae(j, 0, 201);
              continue;
            case "2":
              if (oe[re(0, E, F)](e, oe[re(0, B, V)])) return re(0, U, f) + ae("^(CQ", 0, K);
              continue;
            case "3":
              if (oe[re(0, H, "&%$f")](e, oe[ae(J, 0, -27)])) return oe[ae("GqS4", 0, Z)];
              continue;
            case "4":
              return oe[ae("[#!c", 0, X)];
            case "5":
              if (e === oe.PSZIz) return "spinoff";
              continue;
            case "6":
              if (oe[ae($, 0, Y)](e, oe[re(0, ee, $)])) return oe[ae(te, 0, 12)];
              continue;
          }
          break;
        }
      }(J9t(e));
    }
    function p(e) {
      return e.title || "新会话";
    }
    function f(e) {
      if (!e) return "";
      const t = Date.parse(e);
      if (Number.isNaN(t)) return "";
      const n = Date.now() - t,
        a = Math.floor(n / 6e4);
      if (a < 1) return "刚刚";
      if (a < 60) return `${a} 分钟`;
      const o = Math.floor(a / 60);
      if (o < 24) return `${o} 小时`;
      return `${Math.floor(o / 24)} 天`;
    }
    return (t, n) => (Cr(), Ir("aside", sen, [zr("div", cen, [zr("section", uen, [n[3] || (n[3] = zr("div", {
      class: "section-label"
    }, "开始创作", -1)), zr("div", den, [(Cr(), Ir(br, null, Va(u, t => zr("button", {
      key: t.key,
      class: K(["create-item", {
        active: e.activeWorkflow === t.key
      }]),
      onClick: e => function (e) {
        if ("imitation-short" === e.key) return a("switch-workflow", "imitation-short"), void a("open-imitation-short");
        a("switch-workflow", e.key), "short" !== e.key && "book-create" !== e.key && a("open-import", e.key);
      }(t)
    }, [Qr(D9t, {
      name: t.icon,
      size: 18
    }, null, 8, ["name"]), zr("span", null, Y(t.label), 1)], 10, ven)), 64))])]), zr("section", pen, [zr("div", fen, [zr("button", {
      type: "button",
      class: "section-toggle flex-1",
      onClick: n[0] || (n[0] = e => o.value = !o.value)
    }, [Qr(Tt($I), {
      component: Tt(qH),
      size: 12,
      class: K({
        rotated: o.value
      })
    }, null, 8, ["component", "class"]), n[4] || (n[4] = zr("span", null, "我的作品", -1)), zr("span", men, Y(s.value), 1)])]), Cn(zr("div", null, [i.value.length ? Fr("", !0) : (Cr(), Ir("div", hen, Y("all" === l.value ? "还没有作品" : "暂无该类型作品"), 1)), (Cr(!0), Ir(br, null, Va(i.value, t => (Cr(), Ir("div", {
      key: t.id,
      class: "book-block"
    }, [zr("div", gen, [zr("button", {
      class: "expand-btn",
      onClick: e => a("toggle-book", t.id)
    }, [Qr(Tt($I), {
      component: Tt(qH),
      size: 12,
      class: K({
        rotated: e.expandedBooks[t.id]
      })
    }, null, 8, ["component", "class"])], 8, Wen), zr("button", {
      class: K(["book-title-btn", {
        active: e.activeBookId === t.id
      }]),
      onClick: e => a("open-book", t.id)
    }, [Qr(Tt($I), {
      component: Tt(C$),
      size: 14
    }, null, 8, ["component"]), zr("span", ben, Y(t.title), 1)], 10, yen), zr("span", ken, [zr("button", {
      class: "icon-btn",
      title: "重命名",
      onClick: Zi(e => a("rename-book", t), ["stop"])
    }, [Qr(Tt($I), {
      component: Tt(k2),
      size: 12
    }, null, 8, ["component"])], 8, xen), zr("button", {
      class: "icon-btn",
      title: "导出",
      onClick: Zi(e => a("export-book", t), ["stop"])
    }, [Qr(Tt($I), {
      component: Tt(aX),
      size: 12
    }, null, 8, ["component"])], 8, wen), zr("button", {
      class: "icon-btn danger",
      title: "删除",
      onClick: Zi(e => a("delete-book", t), ["stop"])
    }, [Qr(Tt($I), {
      component: Tt(k8),
      size: 12
    }, null, 8, ["component"])], 8, _en)])]), e.expandedBooks[t.id] ? (Cr(), Ir("div", Sen, [(e.bookSessions[t.id] || []).length ? Fr("", !0) : (Cr(), Ir("div", Cen, " 建书后会话归档到这里 ")), (Cr(!0), Ir(br, null, Va(e.bookSessions[t.id] || [], t => (Cr(), Ir("div", {
      key: t.id,
      class: "session-item-wrap"
    }, [zr("button", {
      class: K(["session-item", {
        active: e.activeSessionId === t.id
      }]),
      onClick: e => a("open-session", t.id)
    }, [c(t.id) ? (Cr(), Ir("span", Ren)) : Fr("", !0), Qr(D9t, {
      name: v(t),
      size: 16
    }, null, 8, ["name"]), zr("span", Oen, Y(p(t)), 1)], 10, Pen)]))), 128)), zr("button", {
      class: "session-item muted",
      onClick: e => a("create-book-session", t.id)
    }, [Qr(Tt($I), {
      component: Tt(BV),
      size: 12
    }, null, 8, ["component"]), n[5] || (n[5] = zr("span", null, "新建会话", -1))], 8, Men)])) : Fr("", !0)]))), 128))], 512), [[Xl, o.value]])]), zr("section", Ien, [zr("button", {
      type: "button",
      class: "section-toggle",
      onClick: n[1] || (n[1] = e => r.value = !r.value)
    }, [Qr(Tt($I), {
      component: Tt(qH),
      size: 12,
      class: K({
        rotated: r.value
      })
    }, null, 8, ["component", "class"]), n[6] || (n[6] = zr("span", null, "会话记录", -1))]), Cn(zr("div", Nen, [e.modeSessions.length ? Fr("", !0) : (Cr(), Ir("div", Gen, "当前分类暂无会话，点击下方新建")), (Cr(!0), Ir(br, null, Va(e.modeSessions, t => (Cr(), Ir("div", {
      key: t.id,
      class: "session-item-wrap"
    }, [zr("button", {
      class: K(["session-item", {
        active: e.activeSessionId === t.id && "chat" === e.activeTool
      }]),
      onClick: e => a("open-session", t.id)
    }, [c(t.id) ? (Cr(), Ir("span", Aen)) : Fr("", !0), Qr(D9t, {
      name: v(t),
      size: 16
    }, null, 8, ["name"]), zr("span", Ten, Y(p(t)), 1), zr("span", zen, Y(f(t.updatedAt)), 1), zr("span", Qen, [zr("button", {
      class: "icon-btn",
      onClick: Zi(e => a("rename-session", t), ["stop"])
    }, [Qr(Tt($I), {
      component: Tt(k2),
      size: 12
    }, null, 8, ["component"])], 8, Den), zr("button", {
      class: "icon-btn danger",
      onClick: Zi(e => a("delete-session", t), ["stop"])
    }, [Qr(Tt($I), {
      component: Tt(k8),
      size: 12
    }, null, 8, ["component"])], 8, Len)])], 10, qen)]))), 128)), zr("button", {
      class: "session-item muted",
      onClick: n[2] || (n[2] = t => a("create-session", e.activeWorkflow))
    }, [Qr(Tt($I), {
      component: Tt(BV),
      size: 12
    }, null, 8, ["component"]), n[7] || (n[7] = zr("span", null, "新建会话", -1))])], 512), [[Xl, r.value]])]), zr("section", jen, [(Cr(), Ir(br, null, Va(d, t => zr("button", {
      key: t.key,
      type: "button",
      class: K(["nav-item", {
        active: e.activeTool === t.key
      }]),
      onClick: e => {
        var n;
        "style" === (n = t.key) ? a("open-style") : "import" === n ? a("open-import-modal") : "agent-settings" === n ? a("open-agent-settings") : "scheduler" === n ? a("open-scheduler") : "dashboard" === n ? a("open-dashboard") : "memo" === n && a("open-memo");
      }
    }, [Qr(D9t, {
      name: t.icon,
      size: 16
    }, null, 8, ["name"]), zr("span", null, Y(t.label), 1)], 10, Een)), 64))])])]));
  }
};
