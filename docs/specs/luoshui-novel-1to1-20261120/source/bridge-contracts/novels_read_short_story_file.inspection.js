// Inspection only; target source is never executed.
// Original frontend SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Native entry 0x1800717e0; constructor source/ida/0x180091e60.c:1336

// Frontend candidate 1; original UTF16 [8693084, 8693548]; property UTF8 [8951777, 8951849]; component NovelStudioShortPanel
async function (e, t) {
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
}

// Caller 1; NovelStudioShortPanel/b; UTF16 [8693084, 8693570]
// Exact argument expressions: ["n.shortStoryId", "e.key"]
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
