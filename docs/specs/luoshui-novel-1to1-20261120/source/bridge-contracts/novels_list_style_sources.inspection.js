// Inspection only; target source is never executed.
// Original frontend SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Native entry 0x180077180; constructor source/ida/0x180091e60.c:1499

// Frontend candidate 1; original UTF16 [9254337, 9254653]; property UTF8 [9519775, 9519823]; component NovelStudioPage
async function () {
  var e, t;
  const n = 1539,
    a = "Tv^A",
    o = "OCei",
    r = 1496,
    l = "0UX(",
    i = 19,
    s = {};
  function c(e, t, n, a) {
    return Tan(0, 0, a - 1142, e);
  }
  function u(e, t, n, a) {
    return Gan(0, n, e - i);
  }
  return s.success = !1, s[c("c3iw", 0, 0, n)] = [], (null == (t = null == (e = Ran()) ? void 0 : e["novel" + u(621, 0, a) + "t_sty" + c(o, 0, 0, r) + u(430, 0, l)]) ? void 0 : t.call(e)) ?? s;
}

// Caller 1; NovelStudioPage/be; UTF16 [9254337, 9254655]
// Exact argument expressions: []
async function be() {
  const e = await async function () {
    var e, t;
    const n = 1539,
      a = "Tv^A",
      o = "OCei",
      r = 1496,
      l = "0UX(",
      i = 19,
      s = {};
    function c(e, t, n, a) {
      return Tan(0, 0, a - 1142, e);
    }
    function u(e, t, n, a) {
      return Gan(0, n, e - i);
    }
    return s.success = !1, s[c("c3iw", 0, 0, n)] = [], (null == (t = null == (e = Ran()) ? void 0 : e["novel" + u(621, 0, a) + "t_sty" + c(o, 0, 0, r) + u(430, 0, l)]) ? void 0 : t.call(e)) ?? s;
  }();
  (null == e ? void 0 : e.success) && (O.value = e.data || []);
}
