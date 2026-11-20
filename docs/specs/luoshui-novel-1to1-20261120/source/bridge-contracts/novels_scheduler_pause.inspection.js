// Inspection only; target source is never executed.
// Original frontend SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Native entry 0x18008e1b0; constructor source/ida/0x180091e60.c:2042

// Frontend candidate 1; original UTF16 [8846763, 8847107]; property UTF8 [9110622, 9110667]; component NovelSchedulerPanel
async function (e = {}) {
  var t, n;
  const a = 28,
    o = "vp#)",
    r = "OCei",
    l = "5LFs",
    i = "BnKt";
  function s(e, t, n, a) {
    return Gan(0, n, t - -644);
  }
  const c = {};
  function u(e, t, n, a) {
    return Tan(0, 0, n - -104, a);
  }
  return c.success = !1, (null == (n = null == (t = {
    JoXEm: function (e) {
      return e();
    }
  }[s(0, a, o)](Ran)) ? void 0 : t[u(0, 0, -4, r) + "s_sch" + s(0, 3, l) + u(0, 0, -8, i) + "se"]) ? void 0 : n.call(t, e)) ?? c;
}

// Caller 1; NovelSchedulerPanel/h; UTF16 [8846763, 8847119]
// Exact argument expressions: ["{\n  action: t\n}"]
async function h() {
  var e;
  r.value = !0;
  try {
    const t = "paused" === (null == (e = l.value) ? void 0 : e.status) ? "resume" : "pause",
      n = await async function (e = {}) {
        var t, n;
        const a = 28,
          o = "vp#)",
          r = "OCei",
          l = "5LFs",
          i = "BnKt";
        function s(e, t, n, a) {
          return Gan(0, n, t - -644);
        }
        const c = {};
        function u(e, t, n, a) {
          return Tan(0, 0, n - -104, a);
        }
        return c.success = !1, (null == (n = null == (t = {
          JoXEm: function (e) {
            return e();
          }
        }[s(0, a, o)](Ran)) ? void 0 : t[u(0, 0, -4, r) + "s_sch" + s(0, 3, l) + u(0, 0, -8, i) + "se"]) ? void 0 : n.call(t, e)) ?? c;
      }({
        action: t
      });
    (null == n ? void 0 : n.success) && (l.value = n.data, o.success("resume" === t ? "已恢复" : "已暂停"));
  } finally {
    r.value = !1;
  }
}
