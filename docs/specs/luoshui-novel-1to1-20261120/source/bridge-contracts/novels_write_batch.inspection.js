// Inspection only; target source is never executed.
// Original frontend SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Native entry 0x18008a790; constructor source/ida/0x180091e60.c:1962

// Frontend candidate 1; original UTF16 [8841854, 8842240]; property UTF8 [9105561, 9105610]; component NovelBatchDialog
async function (e) {
  var t, n;
  const a = 94,
    o = "r[Mo",
    r = 1497,
    l = "4M#3",
    i = 1310,
    s = "XGTO",
    c = 216,
    u = 788;
  function d(e, t, n, a) {
    return Gan(0, n, a - -277);
  }
  const v = {};
  function p(e, t, n, a) {
    return Gan(0, e, a - u);
  }
  return v[p("2CEX", 1404, 1144, 1227) + "ss"] = !1, (null == (n = null == (t = {
    IOYEB: function (e) {
      return e();
    }
  }[d(0, 0, "ASta", a)](Ran)) ? void 0 : t[p(o, 0, 0, r) + d(0, 0, "!jb8", -5) + p(l, 0, 0, i) + d(0, 0, s, c)]) ? void 0 : n.call(t, e)) ?? v;
}

// Caller 1; NovelBatchDialog/d; UTF16 [8841854, 8842340]
// Exact argument expressions: ["{\n  bookId: n.bookId,\n  sessionId: n.sessionId,\n  model: n.model,\n  count: r.value,\n  maxConsecutiveFailures: l.value\n}"]
async function d() {
  i.value = !0, s.value = {
    current: 0,
    total: r.value,
    status: "starting"
  }, window.addEventListener("novel-studio", u);
  try {
    const e = await async function (e) {
      var t, n;
      const a = 94,
        o = "r[Mo",
        r = 1497,
        l = "4M#3",
        i = 1310,
        s = "XGTO",
        c = 216,
        u = 788;
      function d(e, t, n, a) {
        return Gan(0, n, a - -277);
      }
      const v = {};
      function p(e, t, n, a) {
        return Gan(0, e, a - u);
      }
      return v[p("2CEX", 1404, 1144, 1227) + "ss"] = !1, (null == (n = null == (t = {
        IOYEB: function (e) {
          return e();
        }
      }[d(0, 0, "ASta", a)](Ran)) ? void 0 : t[p(o, 0, 0, r) + d(0, 0, "!jb8", -5) + p(l, 0, 0, i) + d(0, 0, s, c)]) ? void 0 : n.call(t, e)) ?? v;
    }({
      bookId: n.bookId,
      sessionId: n.sessionId,
      model: n.model,
      count: r.value,
      maxConsecutiveFailures: l.value
    });
    if (null == e ? void 0 : e.success) {
      const t = e.data,
        n = t.stoppedReason ? ` (${function (e) {
          return {
            cancelled: "已取消",
            target_reached: "已达目标章数",
            consecutive_failures: "连续失败过多"
          }[e] || e;
        }(t.stoppedReason)})` : "";
      o.success(`批量写作完成：${t.completed}/${t.total} 章${n}`), a("done", e.data);
    } else o.error((null == e ? void 0 : e.message) || "批量写作失败");
  } catch (Ofr) {
    o.error("批量写作异常：" + (Ofr.message || Ofr));
  } finally {
    window.removeEventListener("novel-studio", u), i.value = !1;
  }
}
