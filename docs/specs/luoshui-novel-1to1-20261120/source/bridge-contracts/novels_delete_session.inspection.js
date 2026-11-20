// Inspection only; target source is never executed.
// Original frontend SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Native entry 0x1800761f0; constructor source/ida/0x180091e60.c:1459

// Frontend candidate 1; original UTF16 [9285620, 9285920]; property UTF8 [9552214, 9552272]; component NovelStudioPage
async function (e) {
  var t, n;
  const a = "$ex^",
    o = 341,
    r = 159,
    l = {};
  function i(e, t, n, a) {
    return Gan(0, e, n - -357);
  }
  function s(e, t, n, a) {
    return Tan(0, 0, t - -58, a);
  }
  return l[i("rlaI", 0, -24) + "ss"] = !1, (null == (n = null == (t = Ran()) ? void 0 : t[i(a, 0, o) + "s_del" + s(r, 196, 0, "!jb8") + s(-148, 87, 0, "6mzS") + "n"]) ? void 0 : n.call(t, e)) ?? l;
}

// Caller 1; NovelStudioPage/It; UTF16 [9285620, 9285926]
// Exact argument expressions: ["e.id"]
async function It(e) {
  const t = await async function (e) {
    var t, n;
    const a = "$ex^",
      o = 341,
      r = 159,
      l = {};
    function i(e, t, n, a) {
      return Gan(0, e, n - -357);
    }
    function s(e, t, n, a) {
      return Tan(0, 0, t - -58, a);
    }
    return l[i("rlaI", 0, -24) + "ss"] = !1, (null == (n = null == (t = Ran()) ? void 0 : t[i(a, 0, o) + "s_del" + s(r, 196, 0, "!jb8") + s(-148, 87, 0, "6mzS") + "n"]) ? void 0 : n.call(t, e)) ?? l;
  }(e.id);
  if (null == t ? void 0 : t.success) {
    if (await ke(), e.bookId && (await xe(e.bookId)), d.value === e.id) {
      d.value = "", v.value = null, p.value = null, f.value = null, b.value = null, at(), ot(), Q.value = !1, D.value = "", L.value = "";
      const e = fe.value[0];
      e ? await Ce(e.id) : we();
    }
  } else n.error((null == t ? void 0 : t.message) || "删除失败");
}
