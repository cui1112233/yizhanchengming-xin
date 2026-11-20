// Inspection only; target source is never executed.
// Original frontend SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Native entry 0x18008ef80; constructor source/ida/0x180091e60.c:2082

// Frontend candidate 1; original UTF16 [9279484, 9279777]; property UTF8 [9545805, 9545841]; component NovelStudioPage
async function (e) {
  var t, n;
  const a = "oP9p",
    o = 401,
    r = "ASta",
    l = 766,
    i = "XGTO",
    s = 526,
    c = "hjoV",
    u = {};
  function d(e, t, n, a) {
    return Tan(0, 0, n - 367, t);
  }
  return u.success = !1, (null == (n = null == (t = {
    nIiaX: function (e) {
      return e();
    }
  }[d(0, a, o)](Ran)) ? void 0 : t["novel" + d(0, r, l) + d(0, i, s) + d(0, c, 785)]) ? void 0 : n.call(t, e)) ?? u;
}

// Caller 1; NovelStudioPage/xt; UTF16 [9279484, 9279835]
// Exact argument expressions: ["{\n  bookId: W.value,\n  sessionId: d.value,\n  model: P.value || void 0\n}"]
async function xt() {
  var e, t, a, o, r, l, i, s, c, u;
  if (W.value) {
    ae.value = !0, b.value = {
      toolName: "auditor",
      bookTitle: null == (e = p.value) ? void 0 : e.title,
      status: "running",
      startedAt: Date.now(),
      stages: [],
      logs: []
    };
    try {
      const e = await async function (e) {
        var t, n;
        const a = "oP9p",
          o = 401,
          r = "ASta",
          l = 766,
          i = "XGTO",
          s = 526,
          c = "hjoV",
          u = {};
        function d(e, t, n, a) {
          return Tan(0, 0, n - 367, t);
        }
        return u.success = !1, (null == (n = null == (t = {
          nIiaX: function (e) {
            return e();
          }
        }[d(0, a, o)](Ran)) ? void 0 : t["novel" + d(0, r, l) + d(0, i, s) + d(0, c, 785)]) ? void 0 : n.call(t, e)) ?? u;
      }({
        bookId: W.value,
        sessionId: d.value,
        model: P.value || void 0
      });
      if (!(null == e ? void 0 : e.success)) return b.value && (b.value = {
        ...b.value,
        status: "error",
        completedAt: Date.now()
      }), void n.error((null == e ? void 0 : e.message) || "审稿失败");
      if ("running" === (null == (t = b.value) ? void 0 : t.status)) {
        const t = null == (a = e.data) ? void 0 : a.chapterNumber,
          n = Boolean((null == (o = e.data) ? void 0 : o.passed) ?? (null == (l = null == (r = e.data) ? void 0 : r.audit) ? void 0 : l.passed));
        b.value = {
          ...b.value,
          status: n ? "completed" : "error",
          completedAt: Date.now(),
          summary: (null == (i = e.data) ? void 0 : i.summary) || wtn(t, n),
          stages: (b.value.stages || []).map(e => ({
            ...e,
            status: "completed"
          }))
        };
      }
      const v = Boolean((null == (s = e.data) ? void 0 : s.passed) ?? (null == (u = null == (c = e.data) ? void 0 : c.audit) ? void 0 : u.passed));
      n[v ? "success" : "warning"](v ? "审稿通过" : "审稿未通过，请查看卡片中的问题列表"), await Ce(d.value);
    } finally {
      ae.value = !1;
    }
  }
}
