// Inspection only; target source is never executed.
// Original frontend SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Native entry 0x180075870; constructor source/ida/0x180091e60.c:1439

// Frontend candidate 1; original UTF16 [9267750, 9268129]; property UTF8 [9533686, 9533739]; component NovelStudioPage
async function (e, t) {
  var n, a;
  const o = 612,
    r = 374,
    l = 476,
    i = 451,
    s = "VDbS",
    c = 733,
    u = "GI7c",
    d = 398,
    v = "BnKt",
    p = 628,
    f = 520,
    m = "k*oF",
    h = 640,
    g = 682;
  function W(e, t, n, a) {
    return Gan(0, n, t - 8);
  }
  const y = {};
  return y[W(655, 584, "j^1X") + "ss"] = !1, (null == (a = null == (n = {
    sfJbz: function (e) {
      return e();
    }
  }[W(o, r, "iUeH")](Ran)) ? void 0 : n[W(l, i, s) + W(c, 643, u) + W(195, d, v) + W(p, f, m) + W(h, g, "5LFs")]) ? void 0 : a.call(n, e, t)) ?? y;
}

// Caller 1; NovelStudioPage/Xe; UTF16 [9267750, 9268143]
// Exact argument expressions: ["d.value", "e.id"]
async function Xe(e) {
  if (await ye(), null == e ? void 0 : e.id) {
    if (u[e.id] = !0, p.value = e, ee.value = !1, d.value && v.value && !v.value.bookId) {
      const t = await async function (e, t) {
        var n, a;
        const o = 612,
          r = 374,
          l = 476,
          i = 451,
          s = "VDbS",
          c = 733,
          u = "GI7c",
          d = 398,
          v = "BnKt",
          p = 628,
          f = 520,
          m = "k*oF",
          h = 640,
          g = 682;
        function W(e, t, n, a) {
          return Gan(0, n, t - 8);
        }
        const y = {};
        return y[W(655, 584, "j^1X") + "ss"] = !1, (null == (a = null == (n = {
          sfJbz: function (e) {
            return e();
          }
        }[W(o, r, "iUeH")](Ran)) ? void 0 : n[W(l, i, s) + W(c, 643, u) + W(195, d, v) + W(p, f, m) + W(h, g, "5LFs")]) ? void 0 : a.call(n, e, t)) ?? y;
      }(d.value, e.id);
      if (null == t ? void 0 : t.success) return await ke(), await xe(e.id), await Ce(d.value), void n.success("书籍已创建，会话已归档到该书下");
    }
    await Te(e.id);
  }
}
