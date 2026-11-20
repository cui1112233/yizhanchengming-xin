// Inspection only; target source is never executed.
// Original frontend SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Native entry 0x18008fb70; constructor source/ida/0x180091e60.c:2102

// Frontend candidate 1; original UTF16 [9281440, 9281801]; property UTF8 [9547898, 9547945]; component NovelStudioPage
async function (e) {
  var t, n;
  const a = "W^o1",
    o = "ZU!e",
    r = 184,
    l = "5LFs",
    i = "A3ag";
  function s(e, t, n, a) {
    return Tan(0, 0, t - -177, n);
  }
  function c(e, t, n, a) {
    return Tan(0, 0, t - -369, e);
  }
  const u = {};
  return u[s(0, 101, "BnKt") + "ss"] = !1, (null == (n = null == (t = {
    HAqhW: function (e) {
      return e();
    }
  }[c(a, -125)](Ran)) ? void 0 : t[c(o, -93) + s(0, r, l) + c(i, -266) + s(0, -88, "yN0A") + "r"]) ? void 0 : n.call(t, e)) ?? u;
}

// Caller 1; NovelStudioPage/St; UTF16 [9281440, 9281873]
// Exact argument expressions: ["{\n  bookId: W.value,\n  sessionId: d.value,\n  model: P.value || void 0,\n  mode: \"polish\"\n}"]
async function St() {
  var e, t, a, o, r;
  if (W.value) {
    ae.value = !0, b.value = {
      toolName: "reviser",
      bookTitle: null == (e = p.value) ? void 0 : e.title,
      status: "running",
      startedAt: Date.now(),
      stages: [],
      logs: []
    };
    try {
      const e = await async function (e) {
        var t, n;
        const a = "W^o1",
          o = "ZU!e",
          r = 184,
          l = "5LFs",
          i = "A3ag";
        function s(e, t, n, a) {
          return Tan(0, 0, t - -177, n);
        }
        function c(e, t, n, a) {
          return Tan(0, 0, t - -369, e);
        }
        const u = {};
        return u[s(0, 101, "BnKt") + "ss"] = !1, (null == (n = null == (t = {
          HAqhW: function (e) {
            return e();
          }
        }[c(a, -125)](Ran)) ? void 0 : t[c(o, -93) + s(0, r, l) + c(i, -266) + s(0, -88, "yN0A") + "r"]) ? void 0 : n.call(t, e)) ?? u;
      }({
        bookId: W.value,
        sessionId: d.value,
        model: P.value || void 0,
        mode: "polish"
      });
      if (!(null == e ? void 0 : e.success)) return b.value && (b.value = {
        ...b.value,
        status: "error",
        completedAt: Date.now()
      }), void ((null == e ? void 0 : e.cancelled) || n.error((null == e ? void 0 : e.message) || "修订失败"));
      if ((null == (t = e.data) ? void 0 : t.book) && (p.value = e.data.book), "running" === (null == (a = b.value) ? void 0 : a.status)) {
        const t = null == (o = e.data) ? void 0 : o.chapterNumber;
        b.value = {
          ...b.value,
          status: "completed",
          completedAt: Date.now(),
          summary: t ? `第 ${t} 章修订完成。` : "修订完成",
          stages: (b.value.stages || []).map(e => ({
            ...e,
            status: "completed"
          }))
        };
      }
      await ye(), await Ce(d.value), n.success(`第 ${null == (r = e.data) ? void 0 : r.chapterNumber} 章已修订`);
    } finally {
      ae.value = !1;
    }
  }
}
