// Inspection only; target source is never executed.
// Original frontend SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Native entry 0x18006ed60; constructor source/ida/0x180091e60.c:1276

// Frontend candidate 1; original UTF16 [8666421, 8666854]; property UTF8 [8924895, 8924947]; component None
async function Ian(e = null) {
  var t, n;
  const a = "5LFs",
    o = "c3iw",
    r = 692,
    l = 834,
    i = "!jb8",
    s = 910,
    c = "ce![",
    u = 444;
  function d(e, t, n, a) {
    return Tan(e - 466, t - u, a - 630, e);
  }
  function v(e, t, n, a) {
    return Tan(e - 167, t - 331, n - -567, t);
  }
  const p = {};
  return p[d("MEJ&", 938, 0, 681) + "ss"] = !1, p[d(a, 1040, 0, 1125)] = [], (null == (n = null == (t = {
    dmrKK: function (e) {
      return e();
    }
  }[d(o, r, 0, l)](Ran)) ? void 0 : t["novel" + d(i, s, 0, 737) + v(-8, "A3ag", -284) + v(-143, c, -15)]) ? void 0 : n.call(t, e)) ?? p;
}

// Caller 1; NovelStudioPage/ke; UTF16 [9254739, 9254748]
// Exact argument expressions: ["null"]
async function ke() {
  const e = await Ian(null);
  (null == e ? void 0 : e.success) && (s.value = e.data || []);
}

// Caller 2; NovelStudioPage/xe; UTF16 [9254833, 9254839]
// Exact argument expressions: ["e"]
async function xe(e) {
  const t = await Ian(e);
  (null == t ? void 0 : t.success) && (c[e] = t.data || []);
}
