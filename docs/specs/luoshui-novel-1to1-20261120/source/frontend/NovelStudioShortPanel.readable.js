// Inspection only. Original SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Original UTF16 offsets [8691913, 8697920); original line 41
// External dependencies and runtime equivalence remain unresolved.
const recoveredComponent = {
  __name: "NovelStudioShortPanel",
  props: {
    shortStory: {
      type: Object,
      default: null
    },
    shortStoryId: {
      type: String,
      default: ""
    },
    collapsed: {
      type: Boolean,
      default: !1
    },
    isDarkMode: Boolean
  },
  emits: ["update:collapsed", "saved"],
  setup(e, {
    emit: t
  }) {
    const n = e,
      a = t,
      o = $q(),
      r = Nt("list"),
      l = Nt(null),
      i = Nt(""),
      s = Nt(!1),
      c = Nt(!1),
      u = Nt({
        chapters: !0,
        outline: !0,
        artifacts: !0,
        sales: !1
      }),
      d = ml(() => {
        var e;
        return (null == (e = n.shortStory) ? void 0 : e.chapters) || [];
      }),
      v = ml(() => {
        var e;
        return (null == (e = n.shortStory) ? void 0 : e.artifacts) || [];
      }),
      p = ml(() => {
        var e;
        return String((null == (e = n.shortStory) ? void 0 : e.outlinePreview) || "").trim();
      }),
      f = ml(() => {
        var e;
        return String((null == (e = n.shortStory) ? void 0 : e.salesPreview) || "").trim();
      }),
      m = ml(() => {
        var e;
        return String((null == (e = n.shortStory) ? void 0 : e.coverImagePath) || "").trim();
      }),
      h = ml(() => {
        var e;
        const t = null == (e = n.shortStory) ? void 0 : e.status;
        return "completed" === t ? "已完成" : "draft" === t ? "草稿" : t || "进行中";
      }),
      g = ml(() => {
        var e;
        return (null == (e = l.value) ? void 0 : e.title) || "";
      });
    function W(e) {
      u.value[e] = !u.value[e];
    }
    function y() {
      r.value = "list", l.value = null, i.value = "";
    }
    async function b(e) {
      var t;
      if (n.shortStoryId) {
        r.value = "artifact", l.value = {
          ...e,
          readOnly: !1
        }, i.value = "", s.value = !0;
        try {
          const a = await async function (e, t) {
            var n, a;
            const o = "#WGY",
              r = 1466,
              l = 1234,
              i = "K)jY",
              s = 1233,
              c = "A3ag",
              u = 1220,
              d = "rlaI",
              v = 1022,
              p = 308,
              f = "x5I8",
              m = 491,
              h = 1336,
              g = 1081,
              W = 1011,
              y = 71;
            function b(e, t, n, a) {
              return Tan(0, 0, n - y, t);
            }
            function k(e, t, n, a) {
              return Tan(0, 0, n - W, t);
            }
            const x = {};
            return x[k(1328, "4M#3", 1157) + "ss"] = !1, (null == (a = null == (n = {
              BYhuQ: function (e) {
                return e();
              }
            }[k(1565, o, r)](Ran)) ? void 0 : n[k(966, "GI7c", 1057) + k(l, i, s) + b(410, c, 615) + k(u, d, v) + b(p, f, m) + k(h, "!jb8", g)]) ? void 0 : a.call(n, e, t)) ?? x;
          }(n.shortStoryId, e.key);
          i.value = (null == a ? void 0 : a.success) ? (null == (t = a.data) ? void 0 : t.content) ?? "" : null;
        } catch {
          i.value = null;
        } finally {
          s.value = !1;
        }
      }
    }
    async function k(e) {
      if (l.value && n.shortStoryId) {
        c.value = !0;
        try {
          const t = await async function (e, t, n) {
            var a, o;
            const r = 1412,
              l = "XGTO",
              i = 1066,
              s = "BBks",
              c = 1618,
              u = 1408,
              d = "2CEX",
              v = 925,
              p = 909,
              f = "kdlK",
              m = 1208,
              h = "*^FY",
              g = 678,
              W = {};
            function y(e, t, n, a) {
              return Gan(0, n, t - g);
            }
            return W[b = "VQi*", k = 324, Tan(0, 0, k - -184, b) + "ss"] = !1, (null == (o = null == (a = {
              QqStw: function (e) {
                return e();
              }
            }[y(r, 1311, l)](Ran)) ? void 0 : a["novel" + y(i, 1234, s) + y(c, u, d) + y(v, p, f) + y(1437, m, h) + "ile"]) ? void 0 : o.call(a, e, t, n)) ?? W;
            var b, k;
          }(n.shortStoryId, l.value.key, e);
          if (!(null == t ? void 0 : t.success)) return void o.error((null == t ? void 0 : t.message) || "保存失败");
          i.value = e, o.success("已保存"), a("saved");
        } finally {
          c.value = !1;
        }
      }
    }
    function x(e) {
      b({
        key: e.key,
        title: e.label || e.title || "文件"
      });
    }
    function w() {
      var e, t;
      const a = null == (t = null == (e = n.shortStory) ? void 0 : e.artifacts) ? void 0 : t.find(e => String(e.key).startsWith("outline/"));
      a && b({
        key: a.key,
        title: "短篇大纲"
      });
    }
    function _() {
      b({
        key: "final/sales.md",
        title: "推荐"
      });
    }
    return Ro(() => n.shortStoryId, () => {
      r.value = "list", l.value = null, i.value = "";
    }), (t, n) => {
      var o;
      return e.shortStory ? (Cr(), Ir("aside", {
        key: 0,
        class: K(["book-side-panel", {
          dark: e.isDarkMode
        }])
      }, ["list" === r.value ? (Cr(), Ir(br, {
        key: 0
      }, [zr("header", mon, [zr("div", null, [zr("div", hon, Y(e.shortStory.title), 1), zr("div", gon, Y(h.value) + " · 约 " + Y(e.shortStory.wordCount || 0) + " 字 ", 1)]), zr("button", {
        type: "button",
        class: "icon-btn",
        onClick: n[0] || (n[0] = e => a("update:collapsed", !0))
      }, "−")]), Qr(Tt(IE), {
        class: "panel-scroll"
      }, {
        default: Sn(() => [m.value ? (Cr(), Ir("section", Won, [zr("div", yon, [n[5] || (n[5] = zr("div", {
          class: "cover-card-title"
        }, "封面图", -1)), Qr(mtn, {
          "image-path": m.value,
          alt: "封面图",
          compact: ""
        }, null, 8, ["image-path"])])])) : Fr("", !0), zr("section", bon, [zr("button", {
          type: "button",
          class: "section-head",
          onClick: n[1] || (n[1] = e => W("chapters"))
        }, [Qr(Tt($I), {
          component: u.value.chapters ? Tt(MH) : Tt(qH),
          size: 14
        }, null, 8, ["component"]), n[6] || (n[6] = zr("span", null, "章节结构", -1))]), Cn(zr("div", kon, [(Cr(!0), Ir(br, null, Va(d.value, e => (Cr(), Ir("button", {
          key: e.number,
          type: "button",
          class: "list-item",
          onClick: t => x({
            key: "final/full.md",
            title: e.title
          })
        }, [Qr(Tt($I), {
          component: Tt(BZ),
          size: 14
        }, null, 8, ["component"]), zr("span", won, Y(e.title), 1), zr("span", _on, Y(e.wordCount) + " 字", 1)], 8, xon))), 128)), d.value.length ? Fr("", !0) : (Cr(), Ir("div", Son, "成稿生成后可查看章节结构"))], 512), [[Xl, u.value.chapters]])]), zr("section", Con, [zr("button", {
          type: "button",
          class: "section-head",
          onClick: n[2] || (n[2] = e => W("outline"))
        }, [Qr(Tt($I), {
          component: u.value.outline ? Tt(MH) : Tt(qH),
          size: 14
        }, null, 8, ["component"]), n[7] || (n[7] = zr("span", null, "大纲", -1))]), Cn(zr("div", Pon, [zr("p", Ron, [jr(Y(p.value || "大纲生成后显示摘要"), 1), p.value ? (Cr(), Ir("span", Oon, "...")) : Fr("", !0)]), e.shortStory.hasOutline ? (Cr(), Ir("button", {
          key: 0,
          type: "button",
          class: "link-btn",
          onClick: w
        }, " 查看完整大纲 → ")) : Fr("", !0)], 512), [[Xl, u.value.outline]])]), zr("section", Mon, [zr("button", {
          type: "button",
          class: "section-head",
          onClick: n[3] || (n[3] = e => W("artifacts"))
        }, [Qr(Tt($I), {
          component: u.value.artifacts ? Tt(MH) : Tt(qH),
          size: 14
        }, null, 8, ["component"]), n[8] || (n[8] = zr("span", null, "成品文件", -1))]), Cn(zr("div", Ion, [(Cr(!0), Ir(br, null, Va(v.value, e => (Cr(), Ir("button", {
          key: e.key,
          type: "button",
          class: "list-item",
          onClick: t => x(e)
        }, [Qr(Tt($I), {
          component: Tt(BZ),
          size: 14
        }, null, 8, ["component"]), zr("span", Gon, Y(e.label), 1), zr("button", {
          type: "button",
          class: "row-edit-btn",
          title: "编辑",
          onClick: Zi(t => function (e) {
            b({
              key: e.key,
              title: e.label || e.title || "文件",
              autoEdit: !0
            });
          }(e), ["stop"])
        }, [Qr(Tt($I), {
          component: Tt($J),
          size: 13
        }, null, 8, ["component"])], 8, qon)], 8, Non))), 128)), v.value.length ? Fr("", !0) : (Cr(), Ir("div", Aon, "暂无成品文件"))], 512), [[Xl, u.value.artifacts]])]), f.value ? (Cr(), Ir("section", Ton, [zr("button", {
          type: "button",
          class: "section-head",
          onClick: n[4] || (n[4] = e => W("sales"))
        }, [Qr(Tt($I), {
          component: u.value.sales ? Tt(MH) : Tt(qH),
          size: 14
        }, null, 8, ["component"]), n[9] || (n[9] = zr("span", null, "推荐", -1))]), Cn(zr("div", zon, [zr("p", Qon, Y(f.value) + "...", 1), zr("button", {
          type: "button",
          class: "link-btn",
          onClick: _
        }, "查看完整推荐 →")], 512), [[Xl, u.value.sales]])])) : Fr("", !0)]),
        _: 1
      })], 64)) : (Cr(), Nr(xan, {
        key: 1,
        title: g.value,
        content: i.value,
        loading: s.value,
        saving: c.value,
        "read-only": !1,
        "auto-edit": Boolean(null == (o = l.value) ? void 0 : o.autoEdit),
        "is-dark-mode": e.isDarkMode,
        onBack: y,
        onSave: k
      }, null, 8, ["title", "content", "loading", "saving", "auto-edit", "is-dark-mode"]))], 2)) : Fr("", !0);
    };
  }
};
