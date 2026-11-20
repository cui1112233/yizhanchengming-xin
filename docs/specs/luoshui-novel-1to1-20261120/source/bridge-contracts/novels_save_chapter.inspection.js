// Inspection only; target source is never executed.
// Original frontend SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Native entry 0x180086d40; constructor source/ida/0x180091e60.c:1842

// Frontend candidate 1; original UTF16 [8685985, 8686313]; property UTF8 [8944373, 8944414]; component NovelStudioBookPanel
async function (e, t, n, a = {}) {
  var o, r;
  const l = 1247,
    i = 1321,
    s = "ZU!e",
    c = 912,
    u = "OCei",
    d = 656,
    v = {};
  function p(e, t, n, a) {
    return Gan(0, n, e - 361);
  }
  function f(e, t, n, a) {
    return Tan(0, 0, t - 1045, a);
  }
  return v[f(1435, 1584, 1650, "WFL$") + "ss"] = !1, (null == (r = null == (o = Ran()) ? void 0 : o[f(l, i, 0, s) + p(c, 0, u) + "e_cha" + p(d, 0, "K)jY")]) ? void 0 : r.call(o, e, t, n, a)) ?? v;
}

// Caller 1; NovelStudioBookPanel/x; UTF16 [8685985, 8686376]
// Exact argument expressions: ["n.bookId", "l.value.chapterNumber", "e", "{\n  syncTruth: !0,\n  replayFrom: !0\n}"]
async function x(e) {
  var t, r;
  if (l.value && n.bookId) {
    c.value = !0;
    try {
      let s;
      if (s = "chapter" === l.value.type ? await async function (e, t, n, a = {}) {
        var o, r;
        const l = 1247,
          i = 1321,
          s = "ZU!e",
          c = 912,
          u = "OCei",
          d = 656,
          v = {};
        function p(e, t, n, a) {
          return Gan(0, n, e - 361);
        }
        function f(e, t, n, a) {
          return Tan(0, 0, t - 1045, a);
        }
        return v[f(1435, 1584, 1650, "WFL$") + "ss"] = !1, (null == (r = null == (o = Ran()) ? void 0 : o[f(l, i, 0, s) + p(c, 0, u) + "e_cha" + p(d, 0, "K)jY")]) ? void 0 : r.call(o, e, t, n, a)) ?? v;
      }(n.bookId, l.value.chapterNumber, e, {
        syncTruth: !0,
        replayFrom: !0
      }) : await async function (e, t, n) {
        var a, o;
        const r = "ASta",
          l = 682,
          i = "Weig",
          s = 1104,
          c = 43,
          u = 1025;
        function d(e, t, n, a) {
          return Tan(0, 0, n - u, t);
        }
        const v = {};
        return v[d(0, "G^6D", 1496) + "ss"] = !1, (null == (o = null == (a = Ran()) ? void 0 : a[p = r, f = l, Gan(0, p, f - c) + "s_sav" + d(0, i, s) + "ry_file"]) ? void 0 : o.call(a, e, t, n)) ?? v;
        var p, f;
      }(n.bookId, l.value.key, e), !(null == s ? void 0 : s.success)) return void o.error((null == s ? void 0 : s.message) || "保存失败");
      i.value = e, (null == (t = s.data) ? void 0 : t.replayedFrom) ? o.success(`已保存，并从第 ${s.data.replayedFrom} 章回放真相`) : o.success((null == (r = s.data) ? void 0 : r.syncedTruth) ? "已保存并同步真相" : "已保存"), a("saved");
    } finally {
      c.value = !1;
    }
  }
}
