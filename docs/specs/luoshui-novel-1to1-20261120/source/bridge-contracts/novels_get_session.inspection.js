// Inspection only; target source is never executed.
// Original frontend SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Native entry 0x18006f6f0; constructor source/ida/0x180091e60.c:1296

// Frontend candidate 1; original UTF16 [8666854, 8667224]; property UTF8 [8925277, 8925317]; component None
async function Nan(e) {
  var t, n;
  const a = 15,
    o = 1337,
    r = 1413,
    l = "MEJ&",
    i = 1314,
    s = "4M#3",
    c = 480,
    u = 1012,
    d = {};
  function v(e, t, n, a) {
    return Gan(e - 21, e, t - -558);
  }
  function p(e, t, n, a) {
    return Tan(e - 298, t - c, e - u, t);
  }
  return d[v("A3ag", 155) + "ss"] = !1, (null == (n = null == (t = {
    gWgfm: function (e) {
      return e();
    }
  }[v("2CEX", a)](Ran)) ? void 0 : t[p(o, "ND3s") + p(r, l) + p(i, "BBks") + v(s, -211)]) ? void 0 : n.call(t, e)) ?? d;
}

// Caller 1; NovelStudioPage/Se; UTF16 [9255170, 9255176]
// Exact argument expressions: ["e"]
async function Se(e) {
  var t, n, a;
  if (!e) return null;
  const o = await Nan(e);
  return (null == o ? void 0 : o.success) ? (await ke(), e === d.value && (v.value = o.data, k.value = (null == (t = o.data) ? void 0 : t.pendingAction) || null, (null == (n = o.data) ? void 0 : n.shortStory) ? f.value = o.data.shortStory : (null == (a = o.data) ? void 0 : a.shortStoryId) && (await gt())), o.data) : null;
}

// Caller 2; NovelStudioPage/Ce; UTF16 [9255517, 9255523]
// Exact argument expressions: ["e"]
async function Ce(e) {
  var t, a, o, r, l, i, s, c, u;
  at(), ot(), Q.value = !1;
  const h = await Nan(e);
  if (!(null == h ? void 0 : h.success)) return void n.error((null == h ? void 0 : h.message) || "加载会话失败");
  d.value = e, v.value = h.data, m.value = J9t(h.data), k.value = (null == (t = h.data) ? void 0 : t.pendingAction) || null, S.value = [], (null == (a = h.data) ? void 0 : a.book) ? (p.value = h.data.book, L.value = h.data.book.id || (null == (o = h.data) ? void 0 : o.bookId) || "") : (null == (r = h.data) ? void 0 : r.bookId) ? (L.value = h.data.bookId, await ht()) : (p.value = null, L.value = ""), (null == (l = h.data) ? void 0 : l.shortStory) ? (f.value = h.data.shortStory, D.value = h.data.shortStory.id || (null == (i = h.data) ? void 0 : i.shortStoryId) || "") : (null == (s = h.data) ? void 0 : s.shortStoryId) ? (D.value = h.data.shortStoryId, await gt()) : (f.value = null, D.value = "");
  const g = function (e) {
    try {
      const t = sessionStorage.getItem("novel_tool_execution");
      if (!t) return null;
      const n = JSON.parse(t);
      if ((null == n ? void 0 : n.sessionId) === e) return n;
    } catch {}
    return null;
  }(e);
  !g || "running" !== g.status && "error" !== g.status ? b.value = null : (b.value = g, "imitation_short_run" === g.toolName && "running" === g.status && (Q.value = !0), "running" === g.status && ("imitation_short_run" !== g.toolName && "short_fiction_run" !== g.toolName || !(null == (c = h.data) ? void 0 : c.shortStoryId) || lt(h.data.shortStoryId), "architect" === g.toolName && (null == (u = h.data) ? void 0 : u.bookId) && st(h.data.bookId))), ge();
}

// Caller 3; NovelStudioPage/None; UTF16 [9269129, 9269135]
// Exact argument expressions: ["e"]
async function (e, t) {
  var n;
  if (!e || !(null == t ? void 0 : t.length)) return;
  if (e === d.value) return void (await et(t));
  const a = await Nan(e);
  if (!(null == a ? void 0 : a.success)) return;
  const o = [...((null == (n = a.data) ? void 0 : n.messages) || []), ...t];
  await qan(e, o), await ke();
}
