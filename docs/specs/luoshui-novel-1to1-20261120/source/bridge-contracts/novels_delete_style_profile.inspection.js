// Inspection only; target source is never executed.
// Original frontend SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Native entry 0x180079700; constructor source/ida/0x180091e60.c:1559

// Frontend candidate 1; original UTF16 [8701705, 8702054]; property UTF8 [8960580, 8960648]; component NovelStylePanel
async function (e) {
  var t, n;
  const a = 1129,
    o = "6mzS",
    r = ")H@8",
    l = "VXV*",
    i = 422,
    s = "Tv^A",
    c = 449;
  function u(e, t, n, a) {
    return Gan(0, e, a - -295);
  }
  const d = {};
  function v(e, t, n, a) {
    return Gan(0, a, t - 359);
  }
  return d[u("MEJ&", 0, 0, -36) + "ss"] = !1, (null == (n = null == (t = Ran()) ? void 0 : t[v(0, a, 0, o) + u(r, 0, 0, 438) + v(0, 805, 0, "GI7c") + u(l, 0, 0, i) + u(s, 0, 0, c) + "le"]) ? void 0 : n.call(t, e)) ?? d;
}

// Caller 1; NovelStylePanel/g; UTF16 [8701705, 8702060]
// Exact argument expressions: ["e.id"]
async function g(e) {
  const t = await async function (e) {
    var t, n;
    const a = 1129,
      o = "6mzS",
      r = ")H@8",
      l = "VXV*",
      i = 422,
      s = "Tv^A",
      c = 449;
    function u(e, t, n, a) {
      return Gan(0, e, a - -295);
    }
    const d = {};
    function v(e, t, n, a) {
      return Gan(0, a, t - 359);
    }
    return d[u("MEJ&", 0, 0, -36) + "ss"] = !1, (null == (n = null == (t = Ran()) ? void 0 : t[v(0, a, 0, o) + u(r, 0, 0, 438) + v(0, 805, 0, "GI7c") + u(l, 0, 0, i) + u(s, 0, 0, c) + "le"]) ? void 0 : n.call(t, e)) ?? d;
  }(e.id);
  (null == t ? void 0 : t.success) ? (r.value = r.value.filter(t => t !== e.id), l.value === e.id && p(), o.success("已删除"), a("refresh")) : o.error((null == t ? void 0 : t.message) || "删除失败");
}
