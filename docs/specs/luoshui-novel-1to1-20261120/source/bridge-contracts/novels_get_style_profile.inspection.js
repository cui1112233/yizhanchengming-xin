// Inspection only; target source is never executed.
// Original frontend SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Native entry 0x180077890; constructor source/ida/0x180091e60.c:1519

// Frontend candidate 1; original UTF16 [8698804, 8699181]; property UTF8 [8957588, 8957649]; component NovelStylePanel
async function (e) {
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
}

// Caller 1; NovelStylePanel/f; UTF16 [8698804, 8699187]
// Exact argument expressions: ["e.id"]
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
