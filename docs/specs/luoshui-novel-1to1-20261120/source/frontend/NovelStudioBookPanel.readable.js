// Inspection only. Original SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Original UTF16 offsets [8683553, 8691226); original line 41
// External dependencies and runtime equivalence remain unresolved.
const recoveredComponent = {
  __name: "NovelStudioBookPanel",
  props: {
    book: {
      type: Object,
      default: null
    },
    bookId: {
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
        characters: !0,
        foundation: !0,
        covers: !0,
        worldview: !0
      }),
      d = ml(() => {
        var e;
        return (null == (e = n.book) ? void 0 : e.foundationFiles) || [];
      }),
      v = ml(() => {
        var e;
        return (null == (e = n.book) ? void 0 : e.coverImages) || [];
      }),
      p = ml(() => {
        var e;
        return (null == (e = n.book) ? void 0 : e.chapters) || [];
      }),
      f = ml(() => {
        var e;
        return (null == (e = n.book) ? void 0 : e.characters) || [];
      }),
      m = ml(() => {
        var e, t, a;
        const o = (null == (e = n.book) ? void 0 : e.worldviewSummary) || (null == (a = null == (t = n.book) ? void 0 : t.storyFiles) ? void 0 : a.story_frame) || "";
        return String(o).replace(/\s+/g, " ").trim().slice(0, 180);
      }),
      h = ml(() => {
        var e;
        return (null == (e = l.value) ? void 0 : e.title) || "";
      }),
      g = ml(() => {
        var e;
        return Boolean(null == (e = l.value) ? void 0 : e.readOnly);
      });
    function W(e) {
      u.value[e] = !u.value[e];
    }
    function y() {
      r.value = "list", l.value = null, i.value = "";
    }
    async function b(e) {
      var t, a;
      if (n.bookId) {
        r.value = "artifact", l.value = {
          ...e
        }, i.value = "", s.value = !0;
        try {
          let o = "";
          if ("chapter" === e.type) {
            const a = await async function (e, t) {
              var n, a;
              const o = 570,
                r = 666,
                l = "VXV*",
                i = 487,
                s = 472,
                c = "rlaI",
                u = 49,
                d = 23,
                v = 221,
                p = 3;
              function f(e, t, n, a) {
                return Tan(0, 0, n - p, e);
              }
              const m = {};
              function h(e, t, n, a) {
                return Gan(0, n, t - v);
              }
              return m[f("Weig", -107, 70) + "ss"] = !1, (null == (a = null == (n = {
                sQvOS: function (e) {
                  return e();
                }
              }[h(0, o, "VXV*")](Ran)) ? void 0 : n[h(0, r, l) + h(0, 701, "WFL$") + f("WFL$", i, s) + f(c, u, d)]) ? void 0 : a.call(n, e, t)) ?? m;
            }(n.bookId, e.chapterNumber);
            o = (null == (t = null == a ? void 0 : a.data) ? void 0 : t.content) ?? "", (null == a ? void 0 : a.success) || (o = null);
          } else {
            const t = await async function (e, t) {
              var n, a;
              const o = 1115,
                r = "G^6D",
                l = 1428,
                i = 1150,
                s = "6mzS",
                c = 408,
                u = 757,
                d = {};
              function v(e, t, n, a) {
                return Tan(0, 0, n - u, e);
              }
              function p(e, t, n, a) {
                return Tan(0, 0, t - -75, e);
              }
              return d[p("ZU!e", 226) + "ss"] = !1, (null == (a = null == (n = {
                qsKAt: function (e) {
                  return e();
                }
              }[v("OCei", o, 946)](Ran)) ? void 0 : n[p(r, -60) + v("kdlK", l, i) + p(s, 246) + p("4M#3", c) + "le"]) ? void 0 : a.call(n, e, t)) ?? d;
            }(n.bookId, e.key);
            o = (null == (a = null == t ? void 0 : t.data) ? void 0 : a.content) ?? "", (null == t ? void 0 : t.success) || (o = null);
          }
          i.value = o;
        } catch {
          i.value = null;
        } finally {
          s.value = !1;
        }
      }
    }
    function k() {
      b({
        type: "foundation",
        key: "story_frame",
        title: "故事基石 / 世界观"
      });
    }
    async function x(e) {
      var t, r;
      if (l.value && n.bookId) {
        c.value = !0;
        try {
          let s;
          if (s = "chapter" === l.value.type ? await async function (e, t, n, a = {}) {
            var o, r;
            const l = 1247,
              i = 1321,
              s = "ZU!e",
              c = 912,
              u = "OCei",
              d = 656,
              v = {};
            function p(e, t, n, a) {
              return Gan(0, n, e - 361);
            }
            function f(e, t, n, a) {
              return Tan(0, 0, t - 1045, a);
            }
            return v[f(1435, 1584, 1650, "WFL$") + "ss"] = !1, (null == (r = null == (o = Ran()) ? void 0 : o[f(l, i, 0, s) + p(c, 0, u) + "e_cha" + p(d, 0, "K)jY")]) ? void 0 : r.call(o, e, t, n, a)) ?? v;
          }(n.bookId, l.value.chapterNumber, e, {
            syncTruth: !0,
            replayFrom: !0
          }) : await async function (e, t, n) {
            var a, o;
            const r = "ASta",
              l = 682,
              i = "Weig",
              s = 1104,
              c = 43,
              u = 1025;
            function d(e, t, n, a) {
              return Tan(0, 0, n - u, t);
            }
            const v = {};
            return v[d(0, "G^6D", 1496) + "ss"] = !1, (null == (o = null == (a = Ran()) ? void 0 : a[p = r, f = l, Gan(0, p, f - c) + "s_sav" + d(0, i, s) + "ry_file"]) ? void 0 : o.call(a, e, t, n)) ?? v;
            var p, f;
          }(n.bookId, l.value.key, e), !(null == s ? void 0 : s.success)) return void o.error((null == s ? void 0 : s.message) || "保存失败");
          i.value = e, (null == (t = s.data) ? void 0 : t.replayedFrom) ? o.success(`已保存，并从第 ${s.data.replayedFrom} 章回放真相`) : o.success((null == (r = s.data) ? void 0 : r.syncedTruth) ? "已保存并同步真相" : "已保存"), a("saved");
        } finally {
          c.value = !1;
        }
      }
    }
    return Ro(() => n.bookId, () => {
      r.value = "list", l.value = null, i.value = "";
    }), (t, n) => {
      var o;
      return e.book ? (Cr(), Ir("aside", {
        key: 0,
        class: K(["book-side-panel", {
          dark: e.isDarkMode
        }])
      }, ["list" === r.value ? (Cr(), Ir(br, {
        key: 0
      }, [zr("header", Lan, [zr("div", null, [zr("div", jan, Y(e.book.title), 1), zr("div", Ean, "已写 " + Y(e.book.chaptersWritten || p.value.length || 0) + " 章", 1)]), zr("button", {
        type: "button",
        class: "icon-btn",
        onClick: n[0] || (n[0] = e => a("update:collapsed", !0))
      }, "−")]), Qr(Tt(IE), {
        class: "panel-scroll"
      }, {
        default: Sn(() => [zr("section", Fan, [zr("button", {
          type: "button",
          class: "section-head",
          onClick: n[1] || (n[1] = e => W("chapters"))
        }, [Qr(Tt($I), {
          component: u.value.chapters ? Tt(MH) : Tt(qH),
          size: 14
        }, null, 8, ["component"]), n[6] || (n[6] = zr("span", null, "章节", -1))]), Cn(zr("div", Ban, [(Cr(!0), Ir(br, null, Va(p.value, e => (Cr(), Ir("button", {
          key: e.number,
          type: "button",
          class: "list-item",
          onClick: t => function (e) {
            b({
              type: "chapter",
              chapterNumber: e.number,
              title: `第 ${e.number} 章 ${e.title}`
            });
          }(e)
        }, [Qr(Tt($I), {
          component: Tt(BZ),
          size: 14
        }, null, 8, ["component"]), zr("span", Uan, "第 " + Y(e.number) + " 章 " + Y(e.title), 1), zr("button", {
          type: "button",
          class: "row-edit-btn",
          title: "编辑",
          onClick: Zi(t => function (e) {
            b({
              type: "chapter",
              chapterNumber: e.number,
              title: `第 ${e.number} 章 ${e.title}`,
              autoEdit: !0
            });
          }(e), ["stop"])
        }, [Qr(Tt($I), {
          component: Tt($J),
          size: 13
        }, null, 8, ["component"])], 8, Kan)], 8, Van))), 128)), p.value.length ? Fr("", !0) : (Cr(), Ir("div", Han, "暂无章节"))], 512), [[Xl, u.value.chapters]])]), zr("section", Jan, [zr("button", {
          type: "button",
          class: "section-head",
          onClick: n[2] || (n[2] = e => W("characters"))
        }, [Qr(Tt($I), {
          component: u.value.characters ? Tt(MH) : Tt(qH),
          size: 14
        }, null, 8, ["component"]), n[7] || (n[7] = zr("span", null, "角色", -1))]), Cn(zr("div", Zan, [(Cr(!0), Ir(br, null, Va(f.value, e => {
          return Cr(), Ir("button", {
            key: e.path,
            type: "button",
            class: "list-item",
            onClick: t => function (e) {
              b({
                type: "character",
                key: e.path.replace(/\.md$/, ""),
                title: e.name
              });
            }(e)
          }, [Qr(Tt($I), {
            component: Tt(Q2),
            size: 14
          }, null, 8, ["component"]), zr("span", $an, Y(e.name), 1), zr("span", {
            class: K(["tag", e.tier])
          }, Y((t = e.tier, "major" === t ? "主要" : "次要")), 3)], 8, Xan);
          var t;
        }), 128)), f.value.length ? Fr("", !0) : (Cr(), Ir("div", Yan, "暂无角色档案"))], 512), [[Xl, u.value.characters]])]), zr("section", eon, [zr("button", {
          type: "button",
          class: "section-head",
          onClick: n[3] || (n[3] = e => W("foundation"))
        }, [Qr(Tt($I), {
          component: u.value.foundation ? Tt(MH) : Tt(qH),
          size: 14
        }, null, 8, ["component"]), n[8] || (n[8] = zr("span", null, "核心文件", -1))]), Cn(zr("div", ton, [(Cr(!0), Ir(br, null, Va(d.value, e => (Cr(), Ir("button", {
          key: e.key,
          type: "button",
          class: "list-item",
          onClick: t => function (e) {
            b({
              type: "foundation",
              key: e.key,
              title: e.label
            });
          }(e)
        }, [Qr(Tt($I), {
          component: Tt(BZ),
          size: 14
        }, null, 8, ["component"]), zr("span", aon, Y(e.label), 1), zr("button", {
          type: "button",
          class: "row-edit-btn",
          title: "编辑",
          onClick: Zi(t => function (e) {
            b({
              type: "foundation",
              key: e.key,
              title: e.label,
              autoEdit: !0
            });
          }(e), ["stop"])
        }, [Qr(Tt($I), {
          component: Tt($J),
          size: 13
        }, null, 8, ["component"])], 8, oon)], 8, non))), 128)), d.value.length ? Fr("", !0) : (Cr(), Ir("div", ron, "暂无核心文件"))], 512), [[Xl, u.value.foundation]])]), v.value.length ? (Cr(), Ir("section", lon, [zr("button", {
          type: "button",
          class: "section-head",
          onClick: n[4] || (n[4] = e => W("covers"))
        }, [Qr(Tt($I), {
          component: u.value.covers ? Tt(MH) : Tt(qH),
          size: 14
        }, null, 8, ["component"]), n[9] || (n[9] = zr("span", null, "封面资源", -1))]), Cn(zr("div", ion, [(Cr(!0), Ir(br, null, Va(v.value, e => (Cr(), Ir("div", {
          key: e.key,
          class: "cover-card"
        }, [zr("div", son, Y(e.label), 1), Qr(mtn, {
          "image-path": e.path,
          alt: e.label,
          compact: ""
        }, null, 8, ["image-path", "alt"])]))), 128))], 512), [[Xl, u.value.covers]])])) : Fr("", !0), zr("section", con, [zr("button", {
          type: "button",
          class: "section-head",
          onClick: n[5] || (n[5] = e => W("worldview"))
        }, [Qr(Tt($I), {
          component: u.value.worldview ? Tt(MH) : Tt(qH),
          size: 14
        }, null, 8, ["component"]), n[10] || (n[10] = zr("span", null, "世界观", -1))]), Cn(zr("div", uon, [zr("p", don, [jr(Y(m.value || "暂无世界观摘要"), 1), m.value ? (Cr(), Ir("span", von, "...")) : Fr("", !0)]), zr("button", {
          type: "button",
          class: "link-btn",
          onClick: k
        }, "查看完整设定 →")], 512), [[Xl, u.value.worldview]])])]),
        _: 1
      })], 64)) : (Cr(), Nr(xan, {
        key: 1,
        title: h.value,
        content: i.value,
        loading: s.value,
        saving: c.value,
        "read-only": g.value,
        "auto-edit": Boolean(null == (o = l.value) ? void 0 : o.autoEdit),
        "is-dark-mode": e.isDarkMode,
        onBack: y,
        onSave: x
      }, null, 8, ["title", "content", "loading", "saving", "read-only", "auto-edit", "is-dark-mode"]))], 2)) : Fr("", !0);
    };
  }
};
