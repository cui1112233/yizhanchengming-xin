// Inspection only; target source is never executed.
// Original frontend SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Native entry 0x18006e800; constructor source/ida/0x180091e60.c:1253

// Frontend candidate 1; original UTF16 [9284065, 9284324]; property UTF8 [9550584, 9550628]; component NovelStudioPage
async function (e) {
  var t, n;
  const a = "0UX(",
    o = 252,
    r = "j^1X",
    l = "!jb8";
  function i(e, t, n, a) {
    return Gan(0, a, e - -443);
  }
  const s = {};
  return s[i(7, 0, 0, "!jb8") + "ss"] = !1, (null == (n = null == (t = Ran()) ? void 0 : t[i(110, 0, 0, a) + i(o, 0, 0, r) + "ete_b" + i(240, 0, 0, l)]) ? void 0 : n.call(t, e)) ?? s;
}

// Caller 1; NovelStudioPage/None; UTF16 [9284065, 9284330]
// Exact argument expressions: ["e.id"]
async () => {
  const t = await async function (e) {
    var t, n;
    const a = "0UX(",
      o = 252,
      r = "j^1X",
      l = "!jb8";
    function i(e, t, n, a) {
      return Gan(0, a, e - -443);
    }
    const s = {};
    return s[i(7, 0, 0, "!jb8") + "ss"] = !1, (null == (n = null == (t = Ran()) ? void 0 : t[i(110, 0, 0, a) + i(o, 0, 0, r) + "ete_b" + i(240, 0, 0, l)]) ? void 0 : n.call(t, e)) ?? s;
  }(e.id);
  (null == t ? void 0 : t.success) ? (delete c[e.id], delete u[e.id], await ye(), W.value === e.id && we(), n.success("作品已删除")) : n.error((null == t ? void 0 : t.message) || "删除失败");
}
