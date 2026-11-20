// Inspection only; target source is never executed.
// Original frontend SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Native entry 0x1800899f0; constructor source/ida/0x180091e60.c:1942

// Frontend candidate 1; original UTF16 [9277869, 9278163]; property UTF8 [9544107, 9544161]; component NovelStudioPage
async function (e) {
  var t, n;
  const a = 767,
    o = "0UX(",
    r = 635,
    l = "NnW0",
    i = 27,
    s = {};
  function c(e, t, n, a) {
    return Gan(0, t, e - i);
  }
  return s.success = !1, (null == (n = null == (t = {
    kOvFK: function (e) {
      return e();
    }
  }[c(a, o)](Ran)) ? void 0 : t["novel" + (u = "ASta", d = -278, Gan(0, u, d - -842)) + c(r, l) + "xt"]) ? void 0 : n.call(t, e)) ?? s;
  var u, d;
}

// Caller 1; NovelStudioPage/yt; UTF16 [9277869, 9278221]
// Exact argument expressions: ["{\n  bookId: W.value,\n  sessionId: d.value,\n  model: P.value || void 0\n}"]
async function yt() {
  var e, t, a, o, r, l, i, s, c, u, v;
  if (W.value) {
    ae.value = !0, b.value = {
      toolName: "writer",
      bookTitle: null == (e = p.value) ? void 0 : e.title,
      status: "running",
      startedAt: Date.now(),
      stages: [],
      logs: []
    };
    try {
      const e = await async function (e) {
        var t, n;
        const a = 767,
          o = "0UX(",
          r = 635,
          l = "NnW0",
          i = 27,
          s = {};
        function c(e, t, n, a) {
          return Gan(0, t, e - i);
        }
        return s.success = !1, (null == (n = null == (t = {
          kOvFK: function (e) {
            return e();
          }
        }[c(a, o)](Ran)) ? void 0 : t["novel" + (u = "ASta", d = -278, Gan(0, u, d - -842)) + c(r, l) + "xt"]) ? void 0 : n.call(t, e)) ?? s;
        var u, d;
      }({
        bookId: W.value,
        sessionId: d.value,
        model: P.value || void 0
      });
      if (!(null == e ? void 0 : e.success)) return b.value && (b.value = {
        ...b.value,
        status: "error",
        completedAt: Date.now()
      }), void ((null == e ? void 0 : e.cancelled) || n.error((null == e ? void 0 : e.message) || "写章失败"));
      if ((null == (t = e.data) ? void 0 : t.book) && (p.value = e.data.book), "running" === (null == (a = b.value) ? void 0 : a.status)) {
        const t = null == (o = p.value) ? void 0 : o.title,
          n = (null == (r = e.data) ? void 0 : r.status) || "ready-for-review";
        ktn(n);
        b.value = {
          ...b.value,
          status: _tn(n, !0),
          completedAt: Date.now(),
          summary: vt(t, {
            chapterNumber: null == (l = e.data) ? void 0 : l.chapterNumber,
            title: null == (i = e.data) ? void 0 : i.title,
            wordCount: null == (s = e.data) ? void 0 : s.wordCount,
            status: n
          }),
          stages: (b.value.stages || []).map(e => ({
            ...e,
            status: "completed"
          }))
        };
      }
      await ye(), await Ce(d.value);
      ktn((null == (c = e.data) ? void 0 : c.status) || "ready-for-review") ? n.warning(`第 ${null == (u = e.data) ? void 0 : u.chapterNumber} 章已写出，但审稿未通过，请复核后再继续`) : n.success(`第 ${null == (v = e.data) ? void 0 : v.chapterNumber} 章已写入`);
    } finally {
      ae.value = !1;
    }
  }
}
