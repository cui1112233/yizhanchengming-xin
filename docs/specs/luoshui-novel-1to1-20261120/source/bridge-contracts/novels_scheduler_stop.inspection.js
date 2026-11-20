// Inspection only; target source is never executed.
// Original frontend SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Native entry 0x18008dad0; constructor source/ida/0x180091e60.c:2022

// Frontend candidate 1; original UTF16 [8846242, 8846551]; property UTF8 [9110054, 9110091]; component NovelSchedulerPanel
async function (e = {}) {
  var t, n;
  const a = "GI7c",
    o = 1080,
    r = 446,
    l = 322,
    i = {};
  function s(e, t, n, a) {
    return Gan(0, a, t - -283);
  }
  return i[s(0, 439, 0, "x5I8") + "ss"] = !1, (null == (n = null == (t = {
    FBnUL: function (e) {
      return e();
    }
  }[c = a, u = o, Gan(0, c, u - l)](Ran)) ? void 0 : t["novels_sch" + s(0, r, 0, "q0!*") + "r_stop"]) ? void 0 : n.call(t, e)) ?? i;
  var c, u;
}

// Caller 1; NovelSchedulerPanel/m; UTF16 [8846242, 8846553]
// Exact argument expressions: []
async function m() {
  r.value = !0;
  try {
    const e = await async function (e = {}) {
      var t, n;
      const a = "GI7c",
        o = 1080,
        r = 446,
        l = 322,
        i = {};
      function s(e, t, n, a) {
        return Gan(0, a, t - -283);
      }
      return i[s(0, 439, 0, "x5I8") + "ss"] = !1, (null == (n = null == (t = {
        FBnUL: function (e) {
          return e();
        }
      }[c = a, u = o, Gan(0, c, u - l)](Ran)) ? void 0 : t["novels_sch" + s(0, r, 0, "q0!*") + "r_stop"]) ? void 0 : n.call(t, e)) ?? i;
      var c, u;
    }();
    (null == e ? void 0 : e.success) && (l.value = e.data, o.success("调度器已停止"), W());
  } finally {
    r.value = !1;
  }
}
