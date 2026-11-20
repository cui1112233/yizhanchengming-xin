// Inspection only; target source is never executed.
// Original frontend SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Native entry 0x18008c8a0; constructor source/ida/0x180091e60.c:2002

// Frontend candidate 1; original UTF16 [8845530, 8845924]; property UTF8 [9109401, 9109452]; component NovelSchedulerPanel
async function (e) {
  var t, n;
  const a = "yN0A",
    o = "6mzS",
    r = 1150,
    l = "XGTO",
    i = 1338,
    s = "!jb8",
    c = 1543,
    u = "2CEX",
    d = 1015;
  function v(e, t, n, a) {
    return Tan(0, 0, t - -410, n);
  }
  const p = {};
  function f(e, t, n, a) {
    return Tan(0, 0, t - d, a);
  }
  return p[v(0, -343, "Weig") + "ss"] = !1, (null == (n = null == (t = {
    Sjzvs: function (e) {
      return e();
    }
  }[v(0, -50, a)](Ran)) ? void 0 : t[f(0, 1577, 0, o) + f(0, r, 0, l) + f(0, i, 0, s) + f(0, c, 0, u) + "rt"]) ? void 0 : n.call(t, e)) ?? p;
}

// Caller 1; NovelSchedulerPanel/f; UTF16 [8845530, 8846061]
// Exact argument expressions: ["{\n  bookId: n.bookId,\n  sessionId: n.sessionId,\n  model: n.model,\n  intervalSeconds: 60 * i.value,\n  dailyChapterCap: s.value,\n  maxConsecutiveFailures: c.value\n}"]
async function f() {
  r.value = !0;
  try {
    const e = await async function (e) {
      var t, n;
      const a = "yN0A",
        o = "6mzS",
        r = 1150,
        l = "XGTO",
        i = 1338,
        s = "!jb8",
        c = 1543,
        u = "2CEX",
        d = 1015;
      function v(e, t, n, a) {
        return Tan(0, 0, t - -410, n);
      }
      const p = {};
      function f(e, t, n, a) {
        return Tan(0, 0, t - d, a);
      }
      return p[v(0, -343, "Weig") + "ss"] = !1, (null == (n = null == (t = {
        Sjzvs: function (e) {
          return e();
        }
      }[v(0, -50, a)](Ran)) ? void 0 : t[f(0, 1577, 0, o) + f(0, r, 0, l) + f(0, i, 0, s) + f(0, c, 0, u) + "rt"]) ? void 0 : n.call(t, e)) ?? p;
    }({
      bookId: n.bookId,
      sessionId: n.sessionId,
      model: n.model,
      intervalSeconds: 60 * i.value,
      dailyChapterCap: s.value,
      maxConsecutiveFailures: c.value
    });
    (null == e ? void 0 : e.success) ? (l.value = e.data, o.success("调度器已启动"), g()) : o.error((null == e ? void 0 : e.message) || "启动失败");
  } finally {
    r.value = !1;
  }
}
