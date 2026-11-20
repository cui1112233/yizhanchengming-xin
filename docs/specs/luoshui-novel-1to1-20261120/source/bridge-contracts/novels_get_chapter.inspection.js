// Inspection only; target source is never executed.
// Original frontend SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Native entry 0x1800863b0; constructor source/ida/0x180091e60.c:1819

// Frontend candidate 1; original UTF16 [8684755, 8685136]; property UTF8 [8943180, 8943227]; component NovelStudioBookPanel
async function (e, t) {
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
}

// Caller 1; NovelStudioBookPanel/b; UTF16 [8684755, 8685162]
// Exact argument expressions: ["n.bookId", "e.chapterNumber"]
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
