// Inspection only. Original SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Original UTF16 offsets [8698363, 8704788); original line 41
// External dependencies and runtime equivalence remain unresolved.
const recoveredComponent = {
  __name: "NovelStylePanel",
  props: {
    styleSources: {
      type: Array,
      default: () => []
    },
    selectedStyleIds: {
      type: Array,
      default: () => []
    },
    isDarkMode: Boolean
  },
  emits: ["update:selectedStyleIds", "back", "refresh", "applied"],
  setup(e, {
    emit: t
  }) {
    const n = e,
      a = t,
      o = $q(),
      r = Nt([...n.selectedStyleIds]),
      l = Nt(""),
      i = Nt(""),
      s = Nt(""),
      c = Nt(""),
      u = Nt(!1),
      d = Nt(!1),
      v = ml(() => Boolean(l.value));
    function p() {
      l.value = "", i.value = "", s.value = "", c.value = "";
    }
    async function f(e) {
      const t = await async function (e) {
        var t, n;
        const a = 780,
          o = "VXV*",
          r = 504,
          l = 378,
          i = "EtMe",
          s = 811,
          c = "Weig",
          u = {};
        function d(e, t, n, a) {
          return Tan(0, 0, e - 767, n);
        }
        function v(e, t, n, a) {
          return Gan(0, a, t - 146);
        }
        return u[v(364, 487, 664, "XGTO") + "ss"] = !1, (null == (n = null == (t = {
          xjSLn: function (e) {
            return e();
          }
        }[d(a, 0, o)](Ran)) ? void 0 : t["novel" + v(0, r, 0, "4M#3") + v(0, 413, 0, "Tv^A") + v(0, l, 0, i) + d(s, 0, c)]) ? void 0 : n.call(t, e)) ?? u;
      }(e.id);
      (null == t ? void 0 : t.success) ? (l.value = t.data.id, i.value = t.data.title || "", s.value = t.data.sourceText || "", c.value = t.data.styleGuide || "") : o.error((null == t ? void 0 : t.message) || "加载文风失败");
    }
    async function m(e = {}) {
      const t = Boolean(e.keepEditing);
      if (!i.value.trim()) return o.warning("请填写风格名称"), null;
      if (!s.value.trim() && !c.value.trim()) return o.warning("请粘贴参考文本，或导入文件"), null;
      const n = !l.value;
      u.value = !0;
      try {
        const e = await async function (e) {
          var t, n;
          const a = "VXV*",
            o = "ZU!e",
            r = 346,
            l = 727,
            i = 874,
            s = 192,
            c = {};
          function u(e, t, n, a) {
            return Gan(0, t, n - s);
          }
          function d(e, t, n, a) {
            return Tan(0, 0, a - 332, e);
          }
          return c[u(0, "ASta", 509) + "ss"] = !1, (null == (n = null == (t = Ran()) ? void 0 : t[u(0, a, 637) + d(o, 0, 0, r) + d("j^1X", 0, 0, l) + u(0, "0UX(", 561) + u(0, ")H@8", i)]) ? void 0 : n.call(t, e)) ?? c;
        }({
          id: l.value || void 0,
          title: i.value.trim(),
          sourceText: s.value,
          styleGuide: c.value
        });
        if (!(null == e ? void 0 : e.success)) return o.error((null == e ? void 0 : e.message) || "保存失败"), null;
        const u = e.data.id;
        return c.value = e.data.styleGuide || c.value, o.success(n ? "风格已保存，可继续新建下一条" : "风格已更新"), u && !r.value.includes(u) && (r.value = [...r.value, u]), a("refresh"), n && !t ? p() : l.value = u, u;
      } finally {
        u.value = !1;
      }
    }
    async function h() {
      var e;
      d.value = !0;
      try {
        const t = await async function () {
          var e, t;
          const n = "WFL$",
            a = 468,
            o = 1160,
            r = ")H@8",
            l = 1313,
            i = "K)jY",
            s = 440,
            c = "&2iX",
            u = 855,
            d = 20;
          function v(e, t, n, a) {
            return Gan(0, a, e - d);
          }
          const p = {};
          function f(e, t, n, a) {
            return Gan(0, t, e - u);
          }
          return p.success = !1, p.data = [], (null == (t = null == (e = {
            TRprZ: function (e) {
              return e();
            }
          }[v(773, 0, 0, n)](Ran)) ? void 0 : e[v(a, 0, 0, "A3ag") + f(o, r) + "k_sty" + f(l, i) + v(s, 0, 0, c)]) ? void 0 : t.call(e)) ?? p;
        }();
        if (!(null == t ? void 0 : t.success)) return void ((null == t ? void 0 : t.message) && o.error(t.message));
        const n = t.data || [];
        if (!n.length) return;
        if (!l.value) {
          if (!i.value.trim()) {
            const t = (null == (e = n[0]) ? void 0 : e.fileName) || "导入文件";
            i.value = t.replace(/\.[^.]+$/, "");
          }
          const t = await m({
            keepEditing: !0
          });
          if (!t) return;
          l.value = t;
        }
        const r = await async function (e) {
          var t, n;
          const a = 185,
            o = "2CEX",
            r = 1329,
            l = "K)jY",
            i = 206,
            s = "j^1X",
            c = "c3iw",
            u = 832;
          function d(e, t, n, a) {
            return Tan(0, 0, n - -151, a);
          }
          function v(e, t, n, a) {
            return Gan(0, t, n - u);
          }
          const p = {};
          return p[d(0, 0, -148, "k*oF") + "ss"] = !1, (null == (n = null == (t = {
            vtMfZ: function (e) {
              return e();
            }
          }[v(0, o, r)](Ran)) ? void 0 : t[d(0, 0, -63, l) + "s_imp" + d(0, 0, i, s) + d(0, 0, a, "6S(U") + v(0, c, 1588)]) ? void 0 : n.call(t, e)) ?? p;
        }({
          id: l.value,
          files: n
        });
        if (!(null == r ? void 0 : r.success)) return void o.error((null == r ? void 0 : r.message) || "导入失败");
        s.value = r.data.sourceText || "", c.value = r.data.styleGuide || c.value, o.success(`已导入 ${n.length} 个文件`), a("refresh");
      } finally {
        d.value = !1;
      }
    }
    async function g(e) {
      const t = await async function (e) {
        var t, n;
        const a = 1129,
          o = "6mzS",
          r = ")H@8",
          l = "VXV*",
          i = 422,
          s = "Tv^A",
          c = 449;
        function u(e, t, n, a) {
          return Gan(0, e, a - -295);
        }
        const d = {};
        function v(e, t, n, a) {
          return Gan(0, a, t - 359);
        }
        return d[u("MEJ&", 0, 0, -36) + "ss"] = !1, (null == (n = null == (t = Ran()) ? void 0 : t[v(0, a, 0, o) + u(r, 0, 0, 438) + v(0, 805, 0, "GI7c") + u(l, 0, 0, i) + u(s, 0, 0, c) + "le"]) ? void 0 : n.call(t, e)) ?? d;
      }(e.id);
      (null == t ? void 0 : t.success) ? (r.value = r.value.filter(t => t !== e.id), l.value === e.id && p(), o.success("已删除"), a("refresh")) : o.error((null == t ? void 0 : t.message) || "删除失败");
    }
    function W() {
      a("update:selectedStyleIds", [...r.value]), a("applied", [...r.value]), a("back");
    }
    return Ro(() => n.selectedStyleIds, e => {
      r.value = [...(e || [])];
    }, {
      immediate: !0
    }), Ma(() => {
      n.styleSources.length || p();
    }), (t, n) => (Cr(), Ir("div", {
      class: K(["style-page", {
        dark: e.isDarkMode
      }])
    }, [zr("header", jon, [zr("button", {
      type: "button",
      class: "back-btn",
      onClick: n[0] || (n[0] = e => a("back"))
    }, "← 返回对话"), n[5] || (n[5] = zr("h2", null, "写作风格", -1)), n[6] || (n[6] = zr("p", null, "在此管理写作风格知识库：新建名称、粘贴参考文本，或导入 txt / md / docx 等文件（可多选）。", -1))]), Qr(Tt(IE), {
      class: "page-body"
    }, {
      default: Sn(() => [zr("section", Eon, [zr("div", Fon, [zr("h3", null, Y(v.value ? "编辑风格" : "新建风格"), 1), v.value ? (Cr(), Ir("button", {
        key: 0,
        type: "button",
        class: "link-btn",
        onClick: p
      }, "+ 新建")) : Fr("", !0)]), zr("label", Bon, [n[7] || (n[7] = zr("span", null, "风格名称", -1)), Qr(Tt(r_), {
        value: i.value,
        "onUpdate:value": n[1] || (n[1] = e => i.value = e),
        placeholder: "例如：金庸武侠、知乎盐选、某作者口语风"
      }, null, 8, ["value"])]), zr("label", Von, [n[8] || (n[8] = zr("span", null, "参考文本（可粘贴）", -1)), Qr(Tt(r_), {
        value: s.value,
        "onUpdate:value": n[2] || (n[2] = e => s.value = e),
        type: "textarea",
        autosize: {
          minRows: 6,
          maxRows: 14
        },
        placeholder: "直接粘贴样章、段落或写作要求……"
      }, null, 8, ["value"])]), zr("label", Uon, [n[9] || (n[9] = zr("span", null, "文风指南（可选，留空则保存时自动分析）", -1)), Qr(Tt(r_), {
        value: c.value,
        "onUpdate:value": n[3] || (n[3] = e => c.value = e),
        type: "textarea",
        autosize: {
          minRows: 3,
          maxRows: 8
        },
        placeholder: "可手写可执行的文风规则；不填则根据参考文本自动生成"
      }, null, 8, ["value"])]), zr("div", Kon, [Qr(Tt(WS), {
        loading: d.value,
        onClick: h
      }, {
        default: Sn(() => [...(n[10] || (n[10] = [jr("导入文件", -1)]))]),
        _: 1
      }, 8, ["loading"]), Qr(Tt(WS), {
        type: "primary",
        loading: u.value,
        onClick: m
      }, {
        default: Sn(() => [...(n[11] || (n[11] = [jr("保存", -1)]))]),
        _: 1
      }, 8, ["loading"])])]), zr("section", Hon, [n[12] || (n[12] = zr("h3", null, "已有风格", -1)), e.styleSources.length ? (Cr(), Nr(Tt(eO), {
        key: 1,
        value: r.value,
        "onUpdate:value": n[4] || (n[4] = e => r.value = e)
      }, {
        default: Sn(() => [(Cr(!0), Ir(br, null, Va(e.styleSources, e => {
          var t;
          return Cr(), Ir("div", {
            key: e.id,
            class: "style-card"
          }, [zr("label", Zon, [Qr(Tt(aO), {
            value: e.id
          }, null, 8, ["value"]), zr("div", {
            class: "style-info",
            onClick: Zi(t => f(e), ["prevent"])
          }, [zr("div", $on, Y(e.title), 1), zr("div", Yon, [jr(Y(e.sourceChars ? `${e.sourceChars} 字参考` : "无参考文本") + " ", 1), (null == (t = e.sourceFiles) ? void 0 : t.length) ? (Cr(), Ir("span", ern, " · " + Y(e.sourceFiles.length) + " 个文件", 1)) : Fr("", !0)])], 8, Xon)]), zr("button", {
            type: "button",
            class: "danger-btn",
            onClick: t => g(e)
          }, "删除", 8, trn)]);
        }), 128))]),
        _: 1
      }, 8, ["value"])) : (Cr(), Ir("div", Jon, "还没有风格。请在上方新建并保存。"))])]),
      _: 1
    }), zr("footer", nrn, [zr("span", arn, "已选 " + Y(r.value.length) + " 项用于对话参考", 1), Qr(Tt(WS), {
      type: "primary",
      onClick: W
    }, {
      default: Sn(() => [...(n[13] || (n[13] = [jr("应用", -1)]))]),
      _: 1
    })])], 2));
  }
};
