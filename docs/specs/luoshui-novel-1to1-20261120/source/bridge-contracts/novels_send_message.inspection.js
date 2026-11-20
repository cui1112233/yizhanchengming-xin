// Inspection only; target source is never executed.
// Original frontend SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Native entry 0x180082cb0; constructor source/ida/0x180091e60.c:1759

// Frontend candidate 1; original UTF16 [9259707, 9260074]; property UTF8 [9525244, 9525280]; component NovelStudioPage
async function (e) {
  var t, n;
  const a = 479,
    o = "2CEX",
    r = 555,
    l = 644,
    i = "hjoV",
    s = 948,
    c = "6mzS",
    u = 225,
    d = 115;
  function v(e, t, n, a) {
    return Gan(0, e, a - d);
  }
  const p = {};
  function f(e, t, n, a) {
    return Gan(0, n, t - u);
  }
  return p[v("BBks", 0, 0, 670) + "ss"] = !1, (null == (n = null == (t = {
    EJihI: function (e) {
      return e();
    }
  }[v("VXV*", 0, 0, a)](Ran)) ? void 0 : t[v(o, 0, 0, r) + "s_sen" + f(0, l, i) + f(0, s, c)]) ? void 0 : n.call(t, e)) ?? p;
}

// Caller 1; NovelStudioPage/De; UTF16 [9259707, 9260077]
// Exact argument expressions: ["s"]
async function De(e, t) {
  var a, o;
  d.value || (await Pe(m.value));
  const r = d.value;
  if (!r) return;
  const l = V(r);
  l.isStreaming = !0, l.pendingChat = !0, l.streamingText = "", l.streamingReasoning = "", l.pipelineStreamStage = "", b.value = null, v.value && r === d.value && (v.value = {
    ...v.value,
    messages: [...(v.value.messages || []), Ye("user", e)]
  });
  const i = Qre(P.value),
    s = {
      sessionId: r,
      message: e,
      model: P.value || void 0,
      knowledgeBookIds: M.value,
      thinkingMode: !0,
      codexReasoningEffort: i,
      claudeReasoningEffort: i
    };
  null != (null == t ? void 0 : t.contextLength) && (s.contextLength = t.contextLength);
  try {
    const e = await async function (e) {
      var t, n;
      const a = 479,
        o = "2CEX",
        r = 555,
        l = 644,
        i = "hjoV",
        s = 948,
        c = "6mzS",
        u = 225,
        d = 115;
      function v(e, t, n, a) {
        return Gan(0, e, a - d);
      }
      const p = {};
      function f(e, t, n, a) {
        return Gan(0, n, t - u);
      }
      return p[v("BBks", 0, 0, 670) + "ss"] = !1, (null == (n = null == (t = {
        EJihI: function (e) {
          return e();
        }
      }[v("VXV*", 0, 0, a)](Ran)) ? void 0 : t[v(o, 0, 0, r) + "s_sen" + f(0, l, i) + f(0, s, c)]) ? void 0 : n.call(t, e)) ?? p;
    }(s);
    if (l.pendingChat = !1, !(null == e ? void 0 : e.success)) return l.isStreaming = !1, l.streamingText = "", l.streamingReasoning = "", void ((null == e ? void 0 : e.cancelled) || n.error((null == e ? void 0 : e.message) || "发送失败"));
    r === d.value ? (k.value = (null == (a = e.data) ? void 0 : a.action) || null, await Ce(r)) : await Se(r);
  } catch (Ofr) {
    return l.pendingChat = !1, l.isStreaming = !1, l.streamingText = "", l.streamingReasoning = "", void n.error((null == Ofr ? void 0 : Ofr.message) || "发送失败");
  } finally {
    await ke(), (null == (o = v.value) ? void 0 : o.bookId) && (await xe(v.value.bookId));
  }
}
