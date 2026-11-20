// Inspection only. Original SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Original UTF16 offsets [8841139, 8844178); original line 41
// External dependencies and runtime equivalence remain unresolved.
const recoveredComponent = {
  __name: "NovelBatchDialog",
  props: {
    show: Boolean,
    bookId: String,
    sessionId: String,
    model: String,
    isDarkMode: Boolean
  },
  emits: ["update:show", "done"],
  setup(e, {
    emit: t
  }) {
    const n = e,
      a = t,
      o = $q(),
      r = Nt(5),
      l = Nt(3),
      i = Nt(!1),
      s = Nt(null),
      c = ml(() => {
        if (!s.value) return 0;
        const {
          current: e,
          total: t
        } = s.value;
        return Math.round(e / t * 100);
      });
    function u(e) {
      const t = null == e ? void 0 : e.detail;
      "tool:batch_progress" === (null == t ? void 0 : t.type) && (null == t ? void 0 : t.bookId) === n.bookId && (s.value = {
        current: t.current || 0,
        total: t.total || r.value,
        chapterNumber: t.chapterNumber,
        status: t.status,
        stoppedReason: t.stoppedReason
      });
    }
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
    function v() {
      Aan(n.bookId);
    }
    function p() {
      i.value || a("update:show", !1);
    }
    return (t, n) => (Cr(), Nr(Tt(wq), {
      show: e.show,
      preset: "card",
      title: "批量写作",
      closable: !i.value,
      "mask-closable": !i.value,
      style: {
        width: "420px"
      },
      "onUpdate:show": p
    }, {
      default: Sn(() => {
        var t, a, o, u;
        return [zr("div", {
          class: K(["batch-body", {
            dark: e.isDarkMode
          }])
        }, [i.value ? (Cr(), Ir(br, {
          key: 1
        }, [zr("div", Tcn, [Qr(Tt(pE), {
          type: "line",
          percentage: c.value,
          "show-indicator": !0,
          status: "done" === (null == (t = s.value) ? void 0 : t.status) ? "success" : "default"
        }, null, 8, ["percentage", "status"]), zr("p", zcn, [jr(Y((null == (a = s.value) ? void 0 : a.current) || 0) + " / " + Y((null == (o = s.value) ? void 0 : o.total) || r.value) + " ", 1), (null == (u = s.value) ? void 0 : u.chapterNumber) ? (Cr(), Ir(br, {
          key: 0
        }, [jr(" — 正在写第 " + Y(s.value.chapterNumber) + " 章 ", 1)], 64)) : Fr("", !0)])]), Qr(Tt(WS), {
          type: "warning",
          block: "",
          onClick: v
        }, {
          default: Sn(() => [...(n[5] || (n[5] = [jr(" 取消 ", -1)]))]),
          _: 1
        })], 64)) : (Cr(), Ir(br, {
          key: 0
        }, [zr("label", qcn, [n[2] || (n[2] = zr("span", null, "连续写作章数", -1)), Qr(Tt(bL), {
          value: r.value,
          "onUpdate:value": n[0] || (n[0] = e => r.value = e),
          min: 1,
          max: 50,
          size: "small"
        }, null, 8, ["value"])]), zr("label", Acn, [n[3] || (n[3] = zr("span", null, "连续失败停止阈值", -1)), Qr(Tt(bL), {
          value: l.value,
          "onUpdate:value": n[1] || (n[1] = e => l.value = e),
          min: 1,
          max: 10,
          size: "small"
        }, null, 8, ["value"])]), Qr(Tt(WS), {
          type: "primary",
          block: "",
          onClick: d
        }, {
          default: Sn(() => [...(n[4] || (n[4] = [jr(" 开始批量写作 ", -1)]))]),
          _: 1
        })], 64))], 2)];
      }),
      _: 1
    }, 8, ["show", "closable", "mask-closable"]));
  }
};
