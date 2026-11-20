// Inspection only; target source is never executed.
// Original frontend SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Native entry 0x1800805d0; constructor source/ida/0x180091e60.c:1699

// Frontend candidate 1; original UTF16 [8817596, 8817860]; property UTF8 [9078878, 9078932]; component NovelImportPage
async function (e) {
  var t, n;
  const a = "EtMe",
    o = "VXV*",
    r = 348,
    l = {};
  function i(e, t, n, a) {
    return Gan(0, a, e - -326);
  }
  return l[i(396, 0, 0, "x5I8") + "ss"] = !1, (null == (n = null == (t = Ran()) ? void 0 : t["novels_imp" + (s = 968, c = a, Gan(0, c, s - r)) + i(-8, 0, 0, o) + "rs"]) ? void 0 : n.call(t, e)) ?? l;
  var s, c;
}

// Caller 1; NovelImportPage/ie; UTF16 [8817596, 8817902]
// Exact argument expressions: ["{\n  ...n,\n  bookId: v.value || void 0,\n  chapters: a\n}"]
async function ie() {
  var e;
  if ("imitation" === Q.value) {
    if (!c.value.trim()) return void r.warning("请粘贴或上传参考原文");
    if (j.value < 100) return void r.warning("参考原文至少需要 100 字");
  } else if ("chapters" !== Q.value && !i.value.trim()) return void r.warning("请填写书名");
  if ("fanfic" === Q.value && !c.value.trim()) return void r.warning("请粘贴原作参考文本");
  if ("spinoff" === Q.value && !d.value) return void r.warning("请选择正传书籍");
  if ("chapters" === Q.value && !v.value && !i.value.trim()) return void r.warning("请选择目标书籍或填写新书书名");
  if ("chapters" === Q.value && !c.value.trim()) return void r.warning("请粘贴续写章节正文");
  let t = c.value.trim();
  if (h.value.length) for (const a of h.value) {
    const e = g.value.find(e => e.id === a);
    (null == e ? void 0 : e.plain) && (t += `\n\n--- 备忘录: ${e.title} ---\n\n${e.plain}`);
  }
  const n = {
    sessionId: a.sessionId || void 0,
    title: i.value.trim(),
    storyIdea: s.value.trim(),
    referenceText: t,
    genre: "other",
    knowledgeBookIds: u.value
  };
  if ("imitation" === Q.value) {
    const e = function (e, t = 5e4) {
        const n = "9!]I",
          a = "O&gf",
          o = "xCrG",
          r = "82ya",
          l = "wvqL",
          i = "A[V&",
          s = {};
        function c(e, t, n, a) {
          return Zln(0, a - -1453, e);
        }
        function u(e, t, n, a) {
          return Zln(0, t - -1087, e);
        }
        s[u("8VZb", 64)] = function (e, t) {
          return e <= t;
        }, s[u("6pAC", -210)] = function (e, t) {
          return e < t;
        }, s[c(n, 0, 0, -368)] = function (e, t) {
          return e + t;
        };
        const d = s;
        if (!e || d[c(a, 0, 0, -635)](e[c(o, 0, 0, -650) + "h"], t)) return [e];
        const v = [];
        for (let p = 0; d.SYdId(p, e[c(r, 0, 0, -584) + "h"]); p += t) v[c(l, 0, 0, -344)](e[u("A[V&", -285)](p, d[c(i, 0, 0, -504)](p, t)));
        return v;
      }(t, 5e4),
      a = R.value || ain(t);
    return o("submit-start", {
      tab: Q.value,
      payload: {
        ...n,
        creationMode: "styleImitation",
        selectedDirection: m.value.trim() || void 0,
        targetWordCount: _.value || void 0,
        targetChapters: S.value || void 0,
        readerPreferences: C.value || void 0,
        supplementaryInfo: P.value.trim() || void 0,
        styleProfile: a,
        styleDimensions: nin,
        referenceChunks: e.length > 1 ? e : void 0
      }
    }), H(), void le();
  }
  l.value = !0, o("submit-start", {
    tab: Q.value
  });
  try {
    let t;
    if ("fanfic" === Q.value) t = await async function (e) {
      var t, n;
      const a = "ce![",
        o = "VXV*",
        r = "!jb8",
        l = "XGTO",
        i = 432,
        s = 422,
        c = {};
      function u(e, t, n, a) {
        return Tan(0, 0, a - s, n);
      }
      function d(e, t, n, a) {
        return Tan(0, 0, a - -698, n);
      }
      return c[u(0, 0, "k*oF", 425) + "ss"] = !1, (null == (n = null == (t = {
        sZwLE: function (e) {
          return e();
        }
      }[u(0, 0, a, 976)](Ran)) ? void 0 : t[d(0, 0, o, -461) + u(0, 0, r, 678) + d(0, 0, l, -169) + u(0, 0, ")H@8", i)]) ? void 0 : n.call(t, e)) ?? c;
    }({
      ...n,
      fanficMode: p.value
    });else if ("spinoff" === Q.value) t = await async function (e) {
      var t, n;
      const a = "j^1X",
        o = 419,
        r = "kdlK",
        l = "NnW0",
        i = 320,
        s = {};
      function c(e, t, n, a) {
        return Tan(0, 0, n - i, t);
      }
      function u(e, t, n, a) {
        return Tan(0, 0, t - -349, a);
      }
      return s[c(0, "4M#3", 466) + "ss"] = !1, (null == (n = null == (t = {
        YCmJP: function (e) {
          return e();
        }
      }[c(0, a, o)](Ran)) ? void 0 : t["novel" + u(0, -337, 0, r) + u(0, -42, 0, l) + "noff"]) ? void 0 : n.call(t, e)) ?? s;
    }({
      ...n,
      parentBookId: d.value,
      spinoffType: f.value
    });else {
      const a = J(c.value);
      if (a.length < 1) return void r.warning("未能解析出章节，请检查格式");
      t = await async function (e) {
        var t, n;
        const a = "EtMe",
          o = "VXV*",
          r = 348,
          l = {};
        function i(e, t, n, a) {
          return Gan(0, a, e - -326);
        }
        return l[i(396, 0, 0, "x5I8") + "ss"] = !1, (null == (n = null == (t = Ran()) ? void 0 : t["novels_imp" + (s = 968, c = a, Gan(0, c, s - r)) + i(-8, 0, 0, o) + "rs"]) ? void 0 : n.call(t, e)) ?? l;
        var s, c;
      }({
        ...n,
        bookId: v.value || void 0,
        chapters: a
      }), (null == t ? void 0 : t.success) && (t.data = (null == (e = t.data) ? void 0 : e.book) || t.data);
    }
    if (!(null == t ? void 0 : t.success)) return void r.error((null == t ? void 0 : t.message) || "创建失败");
    r.success("chapters" === Q.value ? "章节已导入" : "书籍已创建"), o("created", t.data), H(), le();
  } finally {
    l.value = !1;
  }
}
