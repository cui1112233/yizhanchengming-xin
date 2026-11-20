// Inspection only; target source is never executed.
// Original frontend SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Native entry 0x180084690; constructor source/ida/0x180091e60.c:1779

// Frontend candidate 1; original UTF16 [9264608, 9264955]; property UTF8 [9530246, 9530291]; component NovelStudioPage
async function (e) {
  var t, n;
  const a = 262,
    o = 148,
    r = "#WGY",
    l = 224,
    i = "6mzS";
  function s(e, t, n, a) {
    return Tan(0, 0, a - -181, n);
  }
  const c = {};
  return c[s(73, 437, "W^o1", 208) + "ss"] = !1, (null == (n = null == (t = {
    LtASV: function (e) {
      return e();
    }
  }[u = "c3iw", d = -294, Tan(0, 0, d - -677, u)](Ran)) ? void 0 : t["novels_con" + s(a, o, r, l) + s(-112, -58, i, -61) + "n"]) ? void 0 : n.call(t, e)) ?? c;
  var u, d;
}

// Caller 1; NovelStudioPage/Ke; UTF16 [9264608, 9265045]
// Exact argument expressions: ["{\n  sessionId: d.value,\n  model: P.value || void 0,\n  codexReasoningEffort: A,\n  claudeReasoningEffort: A\n}"]
async function Ke() {
  var e, t, a, o, r, l, i, s, c, f, h, g, W, y, w, _, C, R, O, M, I, N;
  x.value = !0;
  const G = k.value,
    q = null == G ? void 0 : G.action;
  q && je[q] && Ee(je[q]);
  const A = Qre(P.value),
    T = await async function (e) {
      var t, n;
      const a = 262,
        o = 148,
        r = "#WGY",
        l = 224,
        i = "6mzS";
      function s(e, t, n, a) {
        return Tan(0, 0, a - -181, n);
      }
      const c = {};
      return c[s(73, 437, "W^o1", 208) + "ss"] = !1, (null == (n = null == (t = {
        LtASV: function (e) {
          return e();
        }
      }[u = "c3iw", d = -294, Tan(0, 0, d - -677, u)](Ran)) ? void 0 : t["novels_con" + s(a, o, r, l) + s(-112, -58, i, -61) + "n"]) ? void 0 : n.call(t, e)) ?? c;
      var u, d;
    }({
      sessionId: d.value,
      model: P.value || void 0,
      codexReasoningEffort: A,
      claudeReasoningEffort: A
    });
  if (x.value = !1, !(null == T ? void 0 : T.success)) return b.value && (b.value = {
    ...b.value,
    status: "error",
    completedAt: Date.now(),
    error: (null == T ? void 0 : T.message) || "执行失败"
  }), n.error((null == T ? void 0 : T.message) || "执行失败"), void (await Ce(d.value));
  if (null == (e = T.data) ? void 0 : e.assistedAction) {
    const e = T.data.targetRoute || "";
    return ge(), e.includes("imitation") ? Oe("imitation") : e.includes("fanfic") ? Oe("fanfic") : e.includes("spinoff") ? Oe("spinoff") : e.includes("chapters") ? Oe("chapters") : Oe("imitation"), n.info("已打开导入素材，请补充材料"), void (await Ce(d.value));
  }
  if ((null == (t = T.data) ? void 0 : t.book) && (p.value = T.data.book, ee.value = !1, await ye()), (null == (a = T.data) ? void 0 : a.switchToBook) && (null == (o = T.data) ? void 0 : o.session)) return m.value = "book", d.value = T.data.session.id, v.value = T.data.session, k.value = null, (null == (r = T.data.book) ? void 0 : r.id) && (u[T.data.book.id] = !0, await xe(T.data.book.id)), ge(), n.success("书籍已创建，已切换到书籍会话"), void (await ke());
  if (null == (l = T.data) ? void 0 : l.awaitingConfirmation) {
    const e = T.data.checkpoint || T.data.shortFiction || {};
    return Be({
      ...e,
      tool: "short_run" === q ? "short_fiction_run" : e.tool,
      sessionId: d.value,
      message: e.message || "内容已生成，可在右侧修改后确认继续"
    }), T.data.session && (v.value = T.data.session), ee.value = !1, n.info("内容已生成，请查看右侧文件并确认后继续"), void (await ke());
  }
  if ((null == (i = T.data) ? void 0 : i.shortFiction) && (ee.value = !1, n.success(`短篇《${T.data.shortFiction.title || ""}》已生成`)), null == (s = T.data) ? void 0 : s.audit) {
    const e = Boolean((null == (c = T.data.audit) ? void 0 : c.passed) ?? (null == (h = null == (f = T.data.audit) ? void 0 : f.audit) ? void 0 : h.passed));
    n[e ? "success" : "warning"](e ? "审稿通过" : "审稿未通过，请查看卡片中的问题列表");
  } else if ((null == (W = null == (g = T.data) ? void 0 : g.chapter) ? void 0 : W.chapterNumber) && !(null == (w = null == (y = T.data) ? void 0 : y.chapter) ? void 0 : w.mode)) {
    const e = T.data.chapter;
    ktn(e.status || "ready-for-review") ? n.warning(`第 ${e.chapterNumber} 章已写出，但审稿未通过，请复核后再继续`) : n.success(`第 ${e.chapterNumber} 章已写入`);
  } else (null == (C = null == (_ = T.data) ? void 0 : _.chapter) ? void 0 : C.mode) && n.success(`第 ${T.data.chapter.chapterNumber} 章修订完成`);
  await Ce(d.value), await ke(), (null == (R = v.value) ? void 0 : R.bookId) && (await xe(v.value.bookId)), S.value = [], (null == (O = T.data) ? void 0 : O.audit) || (null == (I = null == (M = T.data) ? void 0 : M.chapter) ? void 0 : I.chapterNumber) || (null == (N = T.data) ? void 0 : N.shortFiction) || n.success("执行完成");
}
