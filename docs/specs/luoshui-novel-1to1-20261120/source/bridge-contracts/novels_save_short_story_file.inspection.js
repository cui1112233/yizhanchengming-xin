// Inspection only; target source is never executed.
// Original frontend SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Native entry 0x180071f90; constructor source/ida/0x180091e60.c:1356

// Frontend candidate 1; original UTF16 [8693768, 8694173]; property UTF8 [8952409, 8952464]; component NovelStudioShortPanel
async function (e, t, n) {
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
}

// Caller 1; NovelStudioShortPanel/k; UTF16 [8693768, 8694203]
// Exact argument expressions: ["n.shortStoryId", "l.value.key", "e"]
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
