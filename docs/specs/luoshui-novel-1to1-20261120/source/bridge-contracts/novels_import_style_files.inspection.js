// Inspection only; target source is never executed.
// Original frontend SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Native entry 0x180079ef0; constructor source/ida/0x180091e60.c:1579

// Frontend candidate 1; original UTF16 [8701042, 8701430]; property UTF8 [8959945, 8960004]; component NovelStylePanel
async function (e) {
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
}

// Caller 1; NovelStylePanel/h; UTF16 [8701042, 8701452]
// Exact argument expressions: ["{\n  id: l.value,\n  files: n\n}"]
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
