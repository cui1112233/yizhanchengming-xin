// Inspection only; target source is never executed.
// Original frontend SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Native entry 0x18006cde0; constructor source/ida/0x180091e60.c:1193

// Frontend candidate 1; original UTF16 [9253829, 9254251]; property UTF8 [9519351, 9519388]; component NovelStudioPage
async function () {
  var e, t;
  const n = 1136,
    a = 1187,
    o = "VDbS",
    r = 1191,
    l = 1196,
    i = 1402,
    s = "Tv^A",
    c = 261,
    u = 26,
    d = 90,
    v = "c3iw",
    p = 1008;
  function f(e, t, n, a) {
    return Tan(0, 0, n - p, a);
  }
  function m(e, t, n, a) {
    return Gan(0, n, t - -324);
  }
  const h = {
    IEHmU: function (e) {
      return e();
    },
    aylAS: m(601, 382, "XGTO") + "可用"
  };
  return (null == (t = null == (e = h[f(n, 1234, a, o)](Ran)) ? void 0 : e["novel" + f(r, l, i, s) + m(c, u, "q0!*") + "ks"]) ? void 0 : t.call(e)) ?? {
    success: !1,
    message: h[m(d, -33, v)]
  };
}

// Caller 1; NovelStudioPage/ye; UTF16 [9253829, 9254253]
// Exact argument expressions: []
async function ye() {
  const e = await async function () {
    var e, t;
    const n = 1136,
      a = 1187,
      o = "VDbS",
      r = 1191,
      l = 1196,
      i = 1402,
      s = "Tv^A",
      c = 261,
      u = 26,
      d = 90,
      v = "c3iw",
      p = 1008;
    function f(e, t, n, a) {
      return Tan(0, 0, n - p, a);
    }
    function m(e, t, n, a) {
      return Gan(0, n, t - -324);
    }
    const h = {
      IEHmU: function (e) {
        return e();
      },
      aylAS: m(601, 382, "XGTO") + "可用"
    };
    return (null == (t = null == (e = h[f(n, 1234, a, o)](Ran)) ? void 0 : e["novel" + f(r, l, i, s) + m(c, u, "q0!*") + "ks"]) ? void 0 : t.call(e)) ?? {
      success: !1,
      message: h[m(d, -33, v)]
    };
  }();
  (null == e ? void 0 : e.success) && (i.value = e.data || []);
}
