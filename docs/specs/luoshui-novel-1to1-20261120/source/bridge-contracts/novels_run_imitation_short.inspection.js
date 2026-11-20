// Inspection only; target source is never executed.
// Original frontend SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Native entry 0x18007c840; constructor source/ida/0x180091e60.c:1639

// Frontend candidate 1; original UTF16 [9270923, 9271307]; property UTF8 [9536997, 9537051]; component NovelStudioPage
async function (e) {
  var t, n;
  const a = "$OLu",
    o = "rlaI",
    r = "oP9p",
    l = 1150,
    i = "ZU!e",
    s = "VDbS",
    c = 1316,
    u = 791;
  function d(e, t, n, a) {
    return Gan(0, t, e - -496);
  }
  const v = {};
  function p(e, t, n, a) {
    return Gan(0, t, n - u);
  }
  return v[p(0, "WFL$", 1538) + "ss"] = !1, (null == (n = null == (t = {
    EFVGf: function (e) {
      return e();
    }
  }[d(-104, a)](Ran)) ? void 0 : t[d(-279, "4M#3") + d(188, o) + p(0, r, l) + d(-70, i) + p(0, s, c) + "t"]) ? void 0 : n.call(t, e)) ?? v;
}

// Caller 1; NovelStudioPage/ut; UTF16 [9270923, 9271355]
// Exact argument expressions: ["{\n  ...e,\n  sessionId: d.value,\n  model: P.value || void 0\n}"]
async function ut(e) {
  var t;
  if (z.value = !1, _e("imitation-short"), d.value && "imitation-short" === J9t(v.value) || (await Pe("imitation-short", null, e.title || "仿写短篇"), await vn()), !d.value) return void n.error("无法创建会话，请重试");
  const a = e.title || "仿写短篇";
  ct(a), ge(), ee.value = !1, await et([Ye("user", `开始仿写短篇：${a}`)]), n.info("仿写短篇已开始，可在右侧查看进度");
  let o = !1;
  try {
    const a = await async function (e) {
      var t, n;
      const a = "$OLu",
        o = "rlaI",
        r = "oP9p",
        l = 1150,
        i = "ZU!e",
        s = "VDbS",
        c = 1316,
        u = 791;
      function d(e, t, n, a) {
        return Gan(0, t, e - -496);
      }
      const v = {};
      function p(e, t, n, a) {
        return Gan(0, t, n - u);
      }
      return v[p(0, "WFL$", 1538) + "ss"] = !1, (null == (n = null == (t = {
        EFVGf: function (e) {
          return e();
        }
      }[d(-104, a)](Ran)) ? void 0 : t[d(-279, "4M#3") + d(188, o) + p(0, r, l) + d(-70, i) + p(0, s, c) + "t"]) ? void 0 : n.call(t, e)) ?? v;
    }({
      ...e,
      sessionId: d.value,
      model: P.value || void 0
    });
    if (!(null == a ? void 0 : a.success)) return b.value && (b.value = {
      ...b.value,
      status: "error",
      completedAt: Date.now(),
      error: (null == a ? void 0 : a.message) || "仿写短篇失败",
      summary: (null == a ? void 0 : a.message) || "仿写短篇失败"
    }), void ((null == a ? void 0 : a.cancelled) || n.error((null == a ? void 0 : a.message) || "仿写短篇失败"));
    if (null == (t = a.data) ? void 0 : t.awaitingConfirmation) return o = !0, Be({
      ...a.data,
      tool: "imitation_short_run",
      sessionId: d.value,
      message: a.data.message || "创作蓝图已生成，可在右侧修改后确认继续"
    }), d.value && (await Ce(d.value)), void n.info("创作蓝图已生成，请查看右侧文件并确认后继续");
    await dt(a.data);
  } catch (Ofr) {
    b.value && (b.value = {
      ...b.value,
      status: "error",
      completedAt: Date.now(),
      error: (null == Ofr ? void 0 : Ofr.message) || "仿写短篇失败",
      summary: (null == Ofr ? void 0 : Ofr.message) || "仿写短篇失败"
    }), n.error((null == Ofr ? void 0 : Ofr.message) || "仿写短篇失败");
  } finally {
    o || (Q.value = !1, at());
  }
}
