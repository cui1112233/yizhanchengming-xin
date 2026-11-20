// Inspection only; target source is never executed.
// Original frontend SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Native entry 0x1800727b0; constructor source/ida/0x180091e60.c:1376

// Frontend candidate 1; original UTF16 [9263180, 9263579]; property UTF8 [9528805, 9528877]; component NovelStudioPage
async function (e) {
  var t, n;
  const a = "G^6D",
    o = "G^6D",
    r = 570,
    l = "6mzS",
    i = 975,
    s = "S*tx",
    c = 742,
    u = "OCei",
    d = "*^FY",
    v = 347;
  function p(e, t, n, a) {
    return Gan(0, n, e - 380);
  }
  const f = {};
  function m(e, t, n, a) {
    return Gan(0, e, n - v);
  }
  return f.success = !1, (null == (n = null == (t = {
    dJmbX: function (e) {
      return e();
    }
  }[p(868, 0, a)](Ran)) ? void 0 : t[m(o, 0, r) + p(624, 0, l) + p(i, 0, "*^FY") + "pipel" + m(s, 0, c) + p(888, 0, u) + m(d, 0, 556)]) ? void 0 : n.call(t, e)) ?? f;
}

// Caller 1; NovelStudioPage/Ve; UTF16 [9263180, 9263770]
// Exact argument expressions: ["{\n  sessionId: r.sessionId || d.value,\n  tool: r.tool,\n  checkpoint: r.checkpoint,\n  storyId: r.storyId,\n  bookId: r.bookId,\n  message: r.message,\n  model: P.value || void 0,\n  codexReasoningEffort: l,\n  claudeReasoningEffort: l\n}"]
async function Ve() {
  var e, t, a, o;
  if (!_.value) return;
  w.value = !0;
  const r = {
      ..._.value
    },
    l = Qre(P.value);
  try {
    const i = await async function (e) {
      var t, n;
      const a = "G^6D",
        o = "G^6D",
        r = 570,
        l = "6mzS",
        i = 975,
        s = "S*tx",
        c = 742,
        u = "OCei",
        d = "*^FY",
        v = 347;
      function p(e, t, n, a) {
        return Gan(0, n, e - 380);
      }
      const f = {};
      function m(e, t, n, a) {
        return Gan(0, e, n - v);
      }
      return f.success = !1, (null == (n = null == (t = {
        dJmbX: function (e) {
          return e();
        }
      }[p(868, 0, a)](Ran)) ? void 0 : t[m(o, 0, r) + p(624, 0, l) + p(i, 0, "*^FY") + "pipel" + m(s, 0, c) + p(888, 0, u) + m(d, 0, 556)]) ? void 0 : n.call(t, e)) ?? f;
    }({
      sessionId: r.sessionId || d.value,
      tool: r.tool,
      checkpoint: r.checkpoint,
      storyId: r.storyId,
      bookId: r.bookId,
      message: r.message,
      model: P.value || void 0,
      codexReasoningEffort: l,
      claudeReasoningEffort: l
    });
    if (!(null == i ? void 0 : i.success)) return void n.error((null == i ? void 0 : i.message) || "继续失败");
    if (null == (e = i.data) ? void 0 : e.awaitingConfirmation) return Be({
      ...i.data.checkpoint,
      tool: r.tool,
      sessionId: r.sessionId || d.value,
      message: r.message
    }), void (await Ce(d.value));
    _.value = null, "architect" === r.tool ? n.success("已确认基础设定") : "imitation_short_run" === r.tool ? (Q.value = !1, at(), (null == (t = i.data) ? void 0 : t.shortFiction) ? await dt(i.data.shortFiction) : n.success("仿写短篇已完成")) : (n.success("已继续生成"), (null == (o = null == (a = i.data) ? void 0 : a.shortFiction) ? void 0 : o.storyId) && (D.value = i.data.shortFiction.storyId, await rt())), await Ce(d.value), await ke();
  } finally {
    w.value = !1;
  }
}
