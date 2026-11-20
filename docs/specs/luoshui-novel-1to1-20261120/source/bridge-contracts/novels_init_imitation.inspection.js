// Inspection only; target source is never executed.
// Original frontend SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Native entry 0x18007be90; constructor source/ida/0x180091e60.c:1619

// Frontend candidate 1; original UTF16 [9261816, 9262109]; property UTF8 [9527260, 9527323]; component NovelStudioPage
async function (e) {
  var t, n;
  const a = "vp#)",
    o = 544,
    r = "k*oF",
    l = "iUeH",
    i = 282,
    s = 1072,
    c = "2CEX",
    u = 568,
    d = {};
  function v(e, t, n, a) {
    return Gan(0, t, a - -34);
  }
  return d.success = !1, (null == (n = null == (t = Ran()) ? void 0 : t[v(0, a, 0, o) + v(0, r, 0, 391) + v(0, l, 0, i) + (p = s, f = c, Tan(0, 0, p - u, f)) + "n"]) ? void 0 : n.call(t, e)) ?? d;
  var p, f;
}

// Caller 1; NovelStudioPage/Fe; UTF16 [9261816, 9262112]
// Exact argument expressions: ["t"]
async function Fe({
  tab: e,
  payload: t
}) {
  if (T.value = !1, _e(e), d.value || (await Pe(e)), ge(), Ee(Le[e] || "architect"), "imitation" === e && t) try {
    t.sessionId = d.value, t.model = P.value;
    const e = await async function (e) {
      var t, n;
      const a = "vp#)",
        o = 544,
        r = "k*oF",
        l = "iUeH",
        i = 282,
        s = 1072,
        c = "2CEX",
        u = 568,
        d = {};
      function v(e, t, n, a) {
        return Gan(0, t, a - -34);
      }
      return d.success = !1, (null == (n = null == (t = Ran()) ? void 0 : t[v(0, a, 0, o) + v(0, r, 0, 391) + v(0, l, 0, i) + (p = s, f = c, Tan(0, 0, p - u, f)) + "n"]) ? void 0 : n.call(t, e)) ?? d;
      var p, f;
    }(t);
    (null == e ? void 0 : e.success) ? await Xe(e.data) : (n.error((null == e ? void 0 : e.message) || "创建失败"), b.value = null);
  } catch (Ofr) {
    n.error("创建失败"), b.value = null;
  }
}
