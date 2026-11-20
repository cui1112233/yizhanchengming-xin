// Inspection only; target source is never executed.
// Original frontend SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Native entry 0x180081b90; constructor source/ida/0x180091e60.c:1719

// Frontend candidate 1; original UTF16 [8748713, 8748990]; property UTF8 [9008088, 9008158]; component NovelAgentSettings
async function () {
  var e, t;
  const n = "!jb8",
    a = "6S(U",
    o = "VXV*",
    r = {};
  function l(e, t, n, a) {
    return Gan(0, t, e - -706);
  }
  return r[l(-292, "&2iX") + "ss"] = !1, (null == (t = null == (e = Ran()) ? void 0 : e[l(-392, n) + l(-97, "MEJ&") + "_agen" + (i = a, s = 18, Tan(0, 0, s - -64, i)) + l(-290, o)]) ? void 0 : t.call(e)) ?? r;
  var i, s;
}

// Caller 1; NovelAgentSettings/None; UTF16 [8748713, 8748992]
// Exact argument expressions: []
async function () {
  var e, t, n, a, l;
  r.value = !0;
  try {
    const r = null == (e = window.pywebview) ? void 0 : e.api,
      v = await oae(r);
    i.value = Boolean(null == v ? void 0 : v.ready);
    const p = await async function () {
      var e, t;
      const n = "!jb8",
        a = "6S(U",
        o = "VXV*",
        r = {};
      function l(e, t, n, a) {
        return Gan(0, t, e - -706);
      }
      return r[l(-292, "&2iX") + "ss"] = !1, (null == (t = null == (e = Ran()) ? void 0 : e[l(-392, n) + l(-97, "MEJ&") + "_agen" + (i = a, s = 18, Tan(0, 0, s - -64, i)) + l(-290, o)]) ? void 0 : t.call(e)) ?? r;
      var i, s;
    }();
    if (!(null == p ? void 0 : p.success)) return void o.error((null == p ? void 0 : p.message) || "加载 Agent 设置失败");
    s.value = {
      ...((null == (t = p.data) ? void 0 : t.modelRoutes) || {})
    }, c.value = ((null == (n = p.data) ? void 0 : n.reviewPrompts) || []).map(e => ({
      ...e
    })), u.value = ((null == (a = p.data) ? void 0 : a.customWritePrompts) || []).map(e => ({
      ...e
    }));
    const f = (null == (l = p.data) ? void 0 : l.reviewCycle) || {};
    d.value = {
      maxRetries: Number(f.maxRetries ?? 3),
      passScore: Number(f.passScore ?? 80),
      temperatureStep: Number(f.temperatureStep ?? .15),
      baseTemperature: Number(f.baseTemperature ?? .65),
      enableLengthNorm: !1 !== f.enableLengthNorm,
      enablePostWrite: !1 !== f.enablePostWrite,
      enableAiTells: !1 !== f.enableAiTells,
      requireSecondaryConfirm: Boolean(f.requireSecondaryConfirm)
    };
  } finally {
    r.value = !1;
  }
}
