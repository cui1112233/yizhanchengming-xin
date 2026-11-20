// Inspection only; target source is never executed.
// Original frontend SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Native entry 0x1800780b0; constructor source/ida/0x180091e60.c:1539

// Frontend candidate 1; original UTF16 [8699593, 8699918]; property UTF8 [8958371, 8958438]; component NovelStylePanel
async function (e) {
  var t, n;
  const a = "VXV*",
    o = "ZU!e",
    r = 346,
    l = 727,
    i = 874,
    s = 192,
    c = {};
  function u(e, t, n, a) {
    return Gan(0, t, n - s);
  }
  function d(e, t, n, a) {
    return Tan(0, 0, a - 332, e);
  }
  return c[u(0, "ASta", 509) + "ss"] = !1, (null == (n = null == (t = Ran()) ? void 0 : t[u(0, a, 637) + d(o, 0, 0, r) + d("j^1X", 0, 0, l) + u(0, "0UX(", 561) + u(0, ")H@8", i)]) ? void 0 : n.call(t, e)) ?? c;
}

// Caller 1; NovelStylePanel/m; UTF16 [8699593, 8699999]
// Exact argument expressions: ["{\n  id: l.value || void 0,\n  title: i.value.trim(),\n  sourceText: s.value,\n  styleGuide: c.value\n}"]
async function m(e = {}) {
  const t = Boolean(e.keepEditing);
  if (!i.value.trim()) return o.warning("请填写风格名称"), null;
  if (!s.value.trim() && !c.value.trim()) return o.warning("请粘贴参考文本，或导入文件"), null;
  const n = !l.value;
  u.value = !0;
  try {
    const e = await async function (e) {
      var t, n;
      const a = "VXV*",
        o = "ZU!e",
        r = 346,
        l = 727,
        i = 874,
        s = 192,
        c = {};
      function u(e, t, n, a) {
        return Gan(0, t, n - s);
      }
      function d(e, t, n, a) {
        return Tan(0, 0, a - 332, e);
      }
      return c[u(0, "ASta", 509) + "ss"] = !1, (null == (n = null == (t = Ran()) ? void 0 : t[u(0, a, 637) + d(o, 0, 0, r) + d("j^1X", 0, 0, l) + u(0, "0UX(", 561) + u(0, ")H@8", i)]) ? void 0 : n.call(t, e)) ?? c;
    }({
      id: l.value || void 0,
      title: i.value.trim(),
      sourceText: s.value,
      styleGuide: c.value
    });
    if (!(null == e ? void 0 : e.success)) return o.error((null == e ? void 0 : e.message) || "保存失败"), null;
    const u = e.data.id;
    return c.value = e.data.styleGuide || c.value, o.success(n ? "风格已保存，可继续新建下一条" : "风格已更新"), u && !r.value.includes(u) && (r.value = [...r.value, u]), a("refresh"), n && !t ? p() : l.value = u, u;
  } finally {
    u.value = !1;
  }
}
