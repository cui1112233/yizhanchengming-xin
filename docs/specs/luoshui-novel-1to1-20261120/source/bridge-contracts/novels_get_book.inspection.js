// Inspection only; target source is never executed.
// Original frontend SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Native entry 0x18006d730; constructor source/ida/0x180091e60.c:1213

// Frontend candidate 1; original UTF16 [8665681, 8666021]; property UTF8 [8924050, 8924106]; component None
async function Oan(e) {
  var t, n;
  const a = "$ex^",
    o = "WFL$",
    r = 530,
    l = ")H@8",
    i = 307;
  function s(e, t, n, a) {
    return Gan(e - 161, a, e - -145);
  }
  const c = {};
  return c[s(368, 0, 0, "EtMe") + "ss"] = !1, (null == (n = null == (t = {
    qLTTo: function (e) {
      return e();
    }
  }[s(479, 0, 0, a)](Ran)) ? void 0 : t[s(94, 0, 0, o) + (u = 707, d = "VDbS", Gan(u - 232, d, u - i)) + s(r, 0, 0, l)]) ? void 0 : n.call(t, e)) ?? c;
  var u, d;
}

// Caller 1; NovelStudioPage/Te; UTF16 [9258834, 9258840]
// Exact argument expressions: ["e"]
async function Te(e) {
  var t;
  u[e] = !0;
  const n = await Oan(e);
  (null == n ? void 0 : n.success) && (p.value = n.data, ee.value = !1), c[e] || (await xe(e));
  const a = c[e] || [];
  ge(), a.length ? await Ce(a[0].id) : await Pe("book", e, (null == (t = p.value) ? void 0 : t.title) || "书籍会话");
}

// Caller 2; NovelStudioPage/it; UTF16 [9270009, 9270015]
// Exact argument expressions: ["t"]
async function it() {
  var e;
  const t = L.value || (null == (e = v.value) ? void 0 : e.bookId);
  if (!t) return;
  const n = await Oan(t);
  (null == n ? void 0 : n.success) && n.data && (p.value = n.data, ee.value = !1);
}

// Caller 3; NovelStudioPage/mt; UTF16 [9277093, 9277106]
// Exact argument expressions: ["s.bookId"]
function mt(e) {
  var t, n, a, o, r, l, i;
  const s = (null == e ? void 0 : e.detail) || {};
  if (pt.has(s.type)) return void function (e) {
    const t = ft(e);
    if (!t) return;
    const n = V(t),
      a = t === d.value;
    "chat.start" === e.type ? (n.isStreaming = !0, n.pendingChat = !1, Se(t)) : "chat.delta" === e.type ? (n.isStreaming = !0, n.streamingText += e.text || "", a && b.value && (b.value = {
      ...b.value,
      streamStage: e.stage || n.pipelineStreamStage,
      streamText: n.streamingText
    })) : "chat.reasoning.delta" === e.type ? (n.streamingReasoning += e.text || "", a && b.value && (b.value = {
      ...b.value,
      streamReasoning: n.streamingReasoning
    })) : "chat.stream.start" !== e.type ? "chat.stream.done" !== e.type ? "chat.done" === e.type && (n.isStreaming = !1, n.pendingChat = !1, n.streamingText = "", n.streamingReasoning = "", n.pipelineStreamStage = "", t === d.value ? Q.value || Ce(t) : Se(t)) : nt(t, e.stage || n.pipelineStreamStage).then(() => {
      a && b.value && (b.value = {
        ...b.value,
        streamText: "",
        streamStage: e.stage || n.pipelineStreamStage
      });
    }) : nt(t, n.pipelineStreamStage).then(() => {
      n.pipelineStreamStage = e.stage || e.message || "", n.isStreaming = !0, a && b.value && (b.value = {
        ...b.value,
        streamStage: e.stage || "",
        streamText: ""
      });
    });
  }(s);
  if (!function (e) {
    return !e.sessionId || !!d.value && e.sessionId === d.value;
  }(s)) return;
  const c = Q.value && s.tool && "imitation_short_run" !== s.tool;
  if ("tool:start" === s.type) {
    if (c) return;
    const e = (s.stages || []).map(e => ({
        ...("string" == typeof e ? {
          name: e
        } : e),
        status: e.status || "pending"
      })),
      i = s.bookTitle || (null == (t = b.value) ? void 0 : t.bookTitle) || (null == (n = p.value) ? void 0 : n.title) || (null == (a = k.value) ? void 0 : a.title) || "";
    b.value = {
      ...b.value,
      sessionId: s.sessionId || d.value,
      toolName: s.tool,
      agent: s.agent,
      bookTitle: i,
      status: "running",
      startedAt: (null == (o = b.value) ? void 0 : o.startedAt) || Date.now(),
      stages: e,
      logs: (null == (r = b.value) ? void 0 : r.logs) || []
    }, S.value = e, "imitation_short_run" === s.tool && s.storyId && lt(s.storyId), "short_fiction_run" === s.tool && s.storyId && lt(s.storyId), "architect" === s.tool && s.bookId && st(s.bookId), ["writer", "auditor", "reviser"].includes(s.tool) && (null == (l = v.value) ? void 0 : l.bookId) && st(v.value.bookId);
  } else if ("tool:log" === s.type) u = s.message, b.value && u && (b.value = {
    ...b.value,
    logs: [...(b.value.logs || []), u]
  });else if ("tool:progress" === s.type) {
    const e = ft(s);
    if ("active" === s.status && e) {
      const t = V(e),
        n = s.message || s.stage || "";
      n && (t.pipelineStreamStage = n), t.isStreaming = !0;
    }
    !function (e) {
      var t;
      if (!b.value) return;
      const n = "number" == typeof e.stageIndex ? e.stageIndex : -1,
        a = (null == (t = b.value.stages) ? void 0 : t.length) ? b.value.stages : S.value;
      if (!(null == a ? void 0 : a.length) || n < 0) return void (b.value = {
        ...b.value,
        activeStage: e.message || e.stage || b.value.activeStage
      });
      const o = a.map((t, a) => ({
        ...t,
        status: a < n ? "completed" : a === n ? e.status || "active" : t.status || "pending"
      }));
      b.value = {
        ...b.value,
        activeStage: e.message || e.stage,
        stages: o
      }, S.value = o;
    }(s), s.storyId && (D.value || (D.value = s.storyId), rt()), s.bookId ? st(s.bookId) : "architect" === s.tool && L.value && it();
  } else if ("tool:checkpoint" === s.type) Be(s);else if ("tool:end" === s.type) {
    if (c) return void (s.storyId && Man(s.storyId).then(e => {
      (null == e ? void 0 : e.success) && (f.value = e.data, ee.value = !1);
    }));
    if (b.value) {
      const e = b.value.bookTitle || (null == (i = p.value) ? void 0 : i.title);
      let t = s.summary || b.value.summary || "任务完成",
        n = !1 === s.success ? "error" : "completed";
      if ("writer" === s.tool && s.chapterNumber) {
        const a = s.status || "ready-for-review";
        t = vt(e, {
          chapterNumber: s.chapterNumber,
          chapterTitle: s.chapterTitle,
          wordCount: s.wordCount,
          status: a
        }), n = _tn(a, !1 !== s.success);
      } else if ("auditor" === s.tool && s.chapterNumber) {
        const e = !1 !== s.auditPassed && "ready-for-review" === s.status;
        t = s.summary || wtn(s.chapterNumber, e), n = e ? "completed" : "error";
      } else "reviser" === s.tool && s.chapterNumber ? t = `第 ${s.chapterNumber} 章修订完成。` : "imitation_short_run" === s.tool ? t = s.title ? `短篇《${s.title}》已生成` : "仿写短篇已完成" : "generate_cover" === s.tool ? t = s.coverImagePath ? "封面图已生成" : "封面提示词已保存" : s.chapterTitle && (t = `第 ${s.chapterNumber} 章《${s.chapterTitle}》已写入`);
      b.value = {
        ...b.value,
        status: n,
        completedAt: Date.now(),
        summary: t,
        coverImagePath: s.coverImagePath || b.value.coverImagePath || "",
        details: {
          ...(b.value.details || {}),
          coverImagePath: s.coverImagePath,
          coverError: s.coverError
        },
        error: "error" === n && !1 === s.success ? s.message || s.error || t : "",
        logs: !1 === s.success ? [...(b.value.logs || []), s.message || s.error].filter(Boolean) : b.value.logs || [],
        stages: (b.value.stages || []).map(e => ({
          ...e,
          status: "completed"
        }))
      };
    }
    S.value = S.value.map(e => ({
      ...e,
      status: "completed"
    }));
    const e = ft(s);
    nt(e, V(e).pipelineStreamStage).then(() => {
      e === d.value ? Ce(e) : Se(e);
    }), ye(), ht(), s.bookId && Oan(s.bookId).then(e => {
      (null == e ? void 0 : e.success) && (p.value = e.data, L.value = s.bookId, ee.value = !1);
    }), s.storyId && Man(s.storyId).then(e => {
      (null == e ? void 0 : e.success) && (f.value = e.data, D.value = s.storyId, ee.value = !1);
    }), "short_fiction_run" === s.tool && !1 !== s.success && at(), "imitation_short_run" === s.tool && !1 !== s.success && (Q.value = !1, at(), ye());
  }
  var u;
}

// Caller 4; NovelStudioPage/ht; UTF16 [9277485, 9277497]
// Exact argument expressions: ["W.value"]
async function ht() {
  if (!W.value) return;
  const e = await Oan(W.value);
  (null == e ? void 0 : e.success) && (p.value = e.data);
}
