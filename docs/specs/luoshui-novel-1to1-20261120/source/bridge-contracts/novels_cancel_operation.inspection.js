// Inspection only; target source is never executed.
// Original frontend SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Native entry 0x180085ca0; constructor source/ida/0x180091e60.c:1799

// Frontend candidate 1; original UTF16 [8667746, 8668132]; property UTF8 [8926161, 8926225]; component None
async function Aan(e) {
  var t, n;
  const a = "iUeH",
    o = "Weig",
    r = "K)jY",
    l = 1149,
    i = 1429,
    s = "ND3s",
    c = 54,
    u = 1147;
  function d(e, t, n, a) {
    return Tan(e - 374, t - c, n - u, t);
  }
  function v(e, t, n, a) {
    return Gan(0, n, e - -856);
  }
  const p = {
    success: !1
  };
  return (null == (n = null == (t = {
    jkSGp: function (e) {
      return e();
    }
  }[v(-577, 0, a)](Ran)) ? void 0 : t[v(-427, 0, o) + d(1117, r, l) + d(i, a, 1466) + v(-239, 0, s) + v(-622, 0, "VDbS")]) ? void 0 : n.call(t, e)) ?? p;
}

// Caller 1; NovelBatchDialog/v; UTF16 [8842775, 8842788]
// Exact argument expressions: ["n.bookId"]
function v() {
  Aan(n.bookId);
}

// Caller 2; NovelStudioPage/Je; UTF16 [9267396, 9267402]
// Exact argument expressions: ["t"]
async function Je() {
  var e;
  const t = (null == (e = v.value) ? void 0 : e.bookId) || d.value;
  var n;
  await Aan(t), (n = d.value) && (B[n] = {
    isStreaming: !1,
    pendingChat: !1,
    streamingText: "",
    streamingReasoning: "",
    pipelineStreamStage: ""
  });
}
