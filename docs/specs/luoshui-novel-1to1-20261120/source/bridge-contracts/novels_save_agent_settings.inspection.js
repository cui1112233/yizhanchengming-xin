// Inspection only; target source is never executed.
// Original frontend SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Native entry 0x1800823b0; constructor source/ida/0x180091e60.c:1739

// Frontend candidate 1; original UTF16 [8746217, 8746495]; property UTF8 [9005529, 9005599]; component NovelAgentSettings
async function (e) {
  var t, n;
  const a = 823,
    o = 762,
    r = "VQi*",
    l = 193,
    i = "$OLu",
    s = 554,
    c = {};
  function u(e, t, n, a) {
    return Gan(0, a, n - s);
  }
  return c.success = !1, (null == (n = null == (t = Ran()) ? void 0 : t[u(0, 0, a, "$OLu") + u(0, 0, o, r) + "e_age" + (d = l, v = i, Gan(0, v, d - -229) + "ttings")]) ? void 0 : n.call(t, e)) ?? c;
  var d, v;
}

// Caller 1; NovelAgentSettings/b; UTF16 [8746217, 8747188]
// Exact argument expressions: ["{\n  modelRoutes: {\n    ...s.value\n  },\n  reviewPrompts: c.value.map(y),\n  customWritePrompts: u.value.map(e => ({\n    id: e.id,\n    category: e.category,\n    title: e.title,\n    content: e.content,\n    enabled: Boolean(e.enabled)\n  })),\n  reviewCycle: {\n    maxRetries: Math.max(0, Math.min(5, Number(d.value.maxRetries) || 3)),\n    passScore: Math.max(0, Math.min(100, Number(d.value.passScore) || 80)),\n    temperatureStep: Math.max(0, Math.min(.5, Number(d.value.temperatureStep) || .15)),\n    baseTemperature: Math.max(0, Math.min(2, Number(d.value.baseTemperature) || .65)),\n    enableLengthNorm: Boolean(d.value.enableLengthNorm),\n    enablePostWrite: Boolean(d.value.enablePostWrite),\n    enableAiTells: Boolean(d.value.enableAiTells),\n    requireSecondaryConfirm: Boolean(d.value.requireSecondaryConfirm)\n  }\n}"]
async function b() {
  var e, t, n, a;
  l.value = !0;
  try {
    const r = await async function (e) {
      var t, n;
      const a = 823,
        o = 762,
        r = "VQi*",
        l = 193,
        i = "$OLu",
        s = 554,
        c = {};
      function u(e, t, n, a) {
        return Gan(0, a, n - s);
      }
      return c.success = !1, (null == (n = null == (t = Ran()) ? void 0 : t[u(0, 0, a, "$OLu") + u(0, 0, o, r) + "e_age" + (d = l, v = i, Gan(0, v, d - -229) + "ttings")]) ? void 0 : n.call(t, e)) ?? c;
      var d, v;
    }({
      modelRoutes: {
        ...s.value
      },
      reviewPrompts: c.value.map(y),
      customWritePrompts: u.value.map(e => ({
        id: e.id,
        category: e.category,
        title: e.title,
        content: e.content,
        enabled: Boolean(e.enabled)
      })),
      reviewCycle: {
        maxRetries: Math.max(0, Math.min(5, Number(d.value.maxRetries) || 3)),
        passScore: Math.max(0, Math.min(100, Number(d.value.passScore) || 80)),
        temperatureStep: Math.max(0, Math.min(.5, Number(d.value.temperatureStep) || .15)),
        baseTemperature: Math.max(0, Math.min(2, Number(d.value.baseTemperature) || .65)),
        enableLengthNorm: Boolean(d.value.enableLengthNorm),
        enablePostWrite: Boolean(d.value.enablePostWrite),
        enableAiTells: Boolean(d.value.enableAiTells),
        requireSecondaryConfirm: Boolean(d.value.requireSecondaryConfirm)
      }
    });
    if (!(null == r ? void 0 : r.success)) return void o.error((null == r ? void 0 : r.message) || "保存失败");
    s.value = {
      ...((null == (e = r.data) ? void 0 : e.modelRoutes) || {})
    }, c.value = ((null == (t = r.data) ? void 0 : t.reviewPrompts) || []).map(e => ({
      ...e
    })), u.value = ((null == (n = r.data) ? void 0 : n.customWritePrompts) || []).map(e => ({
      ...e
    }));
    const l = (null == (a = r.data) ? void 0 : a.reviewCycle) || {};
    d.value = {
      maxRetries: Number(l.maxRetries ?? 3),
      passScore: Number(l.passScore ?? 80),
      temperatureStep: Number(l.temperatureStep ?? .15),
      baseTemperature: Number(l.baseTemperature ?? .65),
      enableLengthNorm: !1 !== l.enableLengthNorm,
      enablePostWrite: !1 !== l.enablePostWrite,
      enableAiTells: !1 !== l.enableAiTells,
      requireSecondaryConfirm: Boolean(l.requireSecondaryConfirm)
    }, o.success("已保存");
  } finally {
    l.value = !1;
  }
}
