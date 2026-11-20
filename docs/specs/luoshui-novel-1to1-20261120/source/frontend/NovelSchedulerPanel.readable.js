// Inspection only. Original SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Original UTF16 offsets [8844545, 8849386); original line 41
// External dependencies and runtime equivalence remain unresolved.
const recoveredComponent = {
  __name: "NovelSchedulerPanel",
  props: {
    bookId: String,
    sessionId: String,
    model: String,
    isDarkMode: Boolean
  },
  emits: ["back"],
  setup(e, {
    emit: t
  }) {
    const n = e,
      a = t,
      o = $q(),
      r = Nt(!1),
      l = Nt(null),
      i = Nt(30),
      s = Nt(10),
      c = Nt(3);
    let u = null;
    const d = ml(() => {
        var e;
        return {
          running: "运行中",
          paused: "已暂停",
          stopped: "已停止"
        }[null == (e = l.value) ? void 0 : e.status] || "未知";
      }),
      v = ml(() => {
        var e;
        return {
          running: "success",
          paused: "warning",
          stopped: "default"
        }[null == (e = l.value) ? void 0 : e.status] || "default";
      });
    async function p() {
      const e = await async function (e = {}) {
        var t, n;
        const a = 677,
          o = "GI7c",
          r = 756,
          l = "ND3s",
          i = 923,
          s = "NnW0",
          c = 122,
          u = 891,
          d = "S*tx",
          v = "Tv^A",
          p = 223;
        function f(e, t, n, a) {
          return Tan(0, 0, a - -330, e);
        }
        const m = {};
        function h(e, t, n, a) {
          return Gan(0, n, e - p);
        }
        return m[h(658, 680, "S*tx") + "ss"] = !1, (null == (n = null == (t = {
          WFpUv: function (e) {
            return e();
          }
        }[h(a, 0, o)](Ran)) ? void 0 : t[h(r, 0, l) + h(i, 0, s) + f("j^1X", 0, 0, c) + h(u, 0, d) + f(v, 0, 0, 22)]) ? void 0 : n.call(t, e)) ?? m;
      }();
      (null == e ? void 0 : e.success) && (l.value = e.data);
    }
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
    function g() {
      W(), u = setInterval(p, 1e4);
    }
    function W() {
      u && (clearInterval(u), u = null);
    }
    return Ma(async () => {
      var e;
      await p(), "running" === (null == (e = l.value) ? void 0 : e.status) && g();
    }), Ga(W), (t, n) => {
      var o, u, p, g, W;
      return Cr(), Ir("div", {
        class: K(["scheduler-panel", {
          dark: e.isDarkMode
        }])
      }, [zr("header", Lcn, [zr("button", {
        type: "button",
        class: "back-btn",
        onClick: n[0] || (n[0] = e => a("back"))
      }, "← 返回"), n[4] || (n[4] = zr("h3", null, "自动写作调度器", -1))]), zr("div", jcn, [zr("div", Ecn, [n[5] || (n[5] = zr("span", null, "状态：", -1)), Qr(Tt(cw), {
        type: v.value,
        size: "small"
      }, {
        default: Sn(() => [jr(Y(d.value), 1)]),
        _: 1
      }, 8, ["type"]), "running" === (null == (o = l.value) ? void 0 : o.status) ? (Cr(), Ir("span", Fcn, " 今日 " + Y(l.value.today_written || 0) + " 章 | 累计 " + Y(l.value.total_written || 0) + " 章 ", 1)) : Fr("", !0)]), (null == (u = l.value) ? void 0 : u.consecutive_failures) > 0 ? (Cr(), Ir("div", Bcn, [jr(" 连续失败 " + Y(l.value.consecutive_failures) + " 次 ", 1), (null == (p = l.value) ? void 0 : p.last_error) ? (Cr(), Ir(br, {
        key: 0
      }, [jr(" — " + Y(l.value.last_error), 1)], 64)) : Fr("", !0)])) : Fr("", !0), "stopped" !== (null == (g = l.value) ? void 0 : g.status) && l.value ? (Cr(), Ir("div", Jcn, [Qr(Tt(WS), {
        loading: r.value,
        onClick: h
      }, {
        default: Sn(() => {
          var e;
          return [jr(Y("paused" === (null == (e = l.value) ? void 0 : e.status) ? "恢复" : "暂停"), 1)];
        }),
        _: 1
      }, 8, ["loading"]), Qr(Tt(WS), {
        type: "error",
        loading: r.value,
        onClick: m
      }, {
        default: Sn(() => [...(n[10] || (n[10] = [jr("停止", -1)]))]),
        _: 1
      }, 8, ["loading"])])) : (Cr(), Ir("div", Vcn, [zr("label", Ucn, [n[6] || (n[6] = zr("span", null, "写作间隔（分钟）", -1)), Qr(Tt(bL), {
        value: i.value,
        "onUpdate:value": n[1] || (n[1] = e => i.value = e),
        min: 1,
        max: 1440,
        size: "small"
      }, null, 8, ["value"])]), zr("label", Kcn, [n[7] || (n[7] = zr("span", null, "每日上限", -1)), Qr(Tt(bL), {
        value: s.value,
        "onUpdate:value": n[2] || (n[2] = e => s.value = e),
        min: 1,
        max: 100,
        size: "small"
      }, null, 8, ["value"])]), zr("label", Hcn, [n[8] || (n[8] = zr("span", null, "连续失败暂停阈值", -1)), Qr(Tt(bL), {
        value: c.value,
        "onUpdate:value": n[3] || (n[3] = e => c.value = e),
        min: 1,
        max: 10,
        size: "small"
      }, null, 8, ["value"])]), Qr(Tt(WS), {
        type: "primary",
        block: "",
        loading: r.value,
        onClick: f
      }, {
        default: Sn(() => [...(n[9] || (n[9] = [jr("启动调度器", -1)]))]),
        _: 1
      }, 8, ["loading"])])), (null == (W = l.value) ? void 0 : W.nextWriteIn) > 0 ? (Cr(), Ir("div", Zcn, " 下次写作：约 " + Y(Math.ceil(l.value.nextWriteIn / 60)) + " 分钟后 ", 1)) : Fr("", !0)])], 2);
    };
  }
};
