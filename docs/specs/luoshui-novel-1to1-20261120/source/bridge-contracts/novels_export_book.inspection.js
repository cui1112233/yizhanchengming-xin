// Inspection only; target source is never executed.
// Original frontend SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Native entry 0x180090e60; constructor source/ida/0x180091e60.c:2125

// Frontend candidate 1; original UTF16 [9283066, 9283357]; property UTF8 [9549488, 9549549]; component NovelStudioPage
async function (e, t = "md") {
  var n, a;
  const o = "NnW0",
    r = "kdlK",
    l = "S*tx";
  function i(e, t, n, a) {
    return Tan(0, 0, n - -140, a);
  }
  const s = {};
  return s[i(0, 0, 161, "ZU!e") + "ss"] = !1, (null == (a = null == (n = Ran()) ? void 0 : n[i(0, 0, 25, o) + i(0, 0, -84, r) + "ort_b" + (c = -33, u = l, Gan(0, u, c - -492))]) ? void 0 : a.call(n, e, t)) ?? s;
  var c, u;
}

// Caller 1; NovelStudioPage/None; UTF16 [9283066, 9283371]
// Exact argument expressions: ["e.id", "t.value"]
async function (e, t) {
  a.destroyAll();
  try {
    const a = await async function (e, t = "md") {
      var n, a;
      const o = "NnW0",
        r = "kdlK",
        l = "S*tx";
      function i(e, t, n, a) {
        return Tan(0, 0, n - -140, a);
      }
      const s = {};
      return s[i(0, 0, 161, "ZU!e") + "ss"] = !1, (null == (a = null == (n = Ran()) ? void 0 : n[i(0, 0, 25, o) + i(0, 0, -84, r) + "ort_b" + (c = -33, u = l, Gan(0, u, c - -492))]) ? void 0 : a.call(n, e, t)) ?? s;
      var c, u;
    }(e.id, t.value);
    if (!(null == a ? void 0 : a.success)) return void n.error((null == a ? void 0 : a.message) || "导出失败");
    const o = new Blob([a.data.content || ""], {
        type: `${t.mime};charset=utf-8`
      }),
      r = URL.createObjectURL(o),
      l = document.createElement("a");
    l.href = r, l.download = a.data.filename || `${e.title || "book"}.${t.value}`, l.click(), URL.revokeObjectURL(r), n.success("导出完成");
  } catch {
    n.error("导出失败");
  }
}
