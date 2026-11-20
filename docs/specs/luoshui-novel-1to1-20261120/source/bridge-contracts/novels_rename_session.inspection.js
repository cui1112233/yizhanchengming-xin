// Inspection only; target source is never executed.
// Original frontend SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Native entry 0x180074f40; constructor source/ida/0x180091e60.c:1419

// Frontend candidate 1; original UTF16 [9285051, 9285384]; property UTF8 [9551659, 9551716]; component NovelStudioPage
async function (e, t) {
  var n, a;
  const o = 210,
    r = "$ex^",
    l = 1405,
    i = "j^1X",
    s = 1201,
    c = 205,
    u = 479,
    d = 223,
    v = "!jb8",
    p = 712;
  function f(e, t, n, a) {
    return Tan(0, 0, n - -280, a);
  }
  const m = {};
  return m[f(-431, -390, -277, "k*oF") + "ss"] = !1, (null == (a = null == (n = Ran()) ? void 0 : n[f(-37, 415, o, r) + (h = i, g = s, Gan(0, h, g - p)) + f(c, u, d, v) + "ession"]) ? void 0 : a.call(n, e, t)) ?? m;
  var h, g;
}

// Caller 1; NovelStudioPage/Mt; UTF16 [9285051, 9285404]
// Exact argument expressions: ["G.value.id", "N.value"]
async function Mt() {
  if (!G.value) return;
  if ("book" === q.value) {
    const e = await async function (e, t) {
      var n, a;
      const o = "Weig",
        r = 322,
        l = 1391,
        i = "$ex^",
        s = 281,
        c = {};
      function u(e, t, n, a) {
        return Tan(0, 0, t - 101, e);
      }
      function d(e, t, n, a) {
        return Gan(0, t, e - 890);
      }
      return c[d(1471, "$OLu") + "ss"] = !1, (null == (a = null == (n = Ran()) ? void 0 : n[u(o, r) + d(l, i) + u("$OLu", s) + "ook"]) ? void 0 : a.call(n, e, t)) ?? c;
    }(G.value.id, N.value);
    return I.value = !1, (null == e ? void 0 : e.success) ? (await ye(), W.value === G.value.id && (await ht()), void n.success("书名已更新")) : void n.error((null == e ? void 0 : e.message) || "重命名失败");
  }
  const e = await async function (e, t) {
    var n, a;
    const o = 210,
      r = "$ex^",
      l = 1405,
      i = "j^1X",
      s = 1201,
      c = 205,
      u = 479,
      d = 223,
      v = "!jb8",
      p = 712;
    function f(e, t, n, a) {
      return Tan(0, 0, n - -280, a);
    }
    const m = {};
    return m[f(-431, -390, -277, "k*oF") + "ss"] = !1, (null == (a = null == (n = Ran()) ? void 0 : n[f(-37, 415, o, r) + (h = i, g = s, Gan(0, h, g - p)) + f(c, u, d, v) + "ession"]) ? void 0 : a.call(n, e, t)) ?? m;
    var h, g;
  }(G.value.id, N.value);
  I.value = !1, (null == e ? void 0 : e.success) ? (await ke(), G.value.bookId && (await xe(G.value.bookId)), d.value === G.value.id && (await Ce(G.value.id))) : n.error((null == e ? void 0 : e.message) || "重命名失败");
}
