// Inspection only. Original SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Original UTF16 offsets [8809779, 8830298); original line 41
// External dependencies and runtime equivalence remain unresolved.
const recoveredComponent = {
  __name: "NovelImportPage",
  props: {
    tab: {
      type: String,
      default: "imitation"
    },
    variant: {
      type: String,
      default: "page"
    },
    sessionId: {
      type: String,
      default: ""
    },
    styleSources: {
      type: Array,
      default: () => []
    },
    books: {
      type: Array,
      default: () => []
    },
    isDarkMode: Boolean
  },
  emits: ["back", "close", "created", "update:tab", "submit-start"],
  setup(e, {
    expose: t,
    emit: n
  }) {
    const a = e,
      o = n,
      r = $q(),
      l = Nt(!1),
      i = Nt(""),
      s = Nt(""),
      c = Nt(""),
      u = Nt([]),
      d = Nt(null),
      v = Nt(null),
      p = Nt("canon"),
      f = Nt("sequel"),
      m = Nt(""),
      h = Nt([]),
      g = Nt([]),
      W = Nt(!1),
      y = Nt([]),
      b = Nt([]),
      k = Nt([]),
      x = Nt(!1),
      w = Nt([]),
      _ = Nt(5e3),
      S = Nt(100),
      C = Nt(""),
      P = Nt(""),
      R = Nt(null);
    let O = null;
    const M = ["玄幻", "都市", "仙侠", "科幻", "历史", "悬疑", "言情", "武侠", "奇幻", "灵异", "末世", "系统"],
      I = [{
        label: "3000",
        value: 3e3
      }, {
        label: "4000",
        value: 4e3
      }, {
        label: "5000",
        value: 5e3
      }, {
        label: "6000",
        value: 6e3
      }, {
        label: "8000",
        value: 8e3
      }, {
        label: "1万",
        value: 1e4
      }],
      N = [{
        label: "50",
        value: 50
      }, {
        label: "100",
        value: 100
      }, {
        label: "150",
        value: 150
      }, {
        label: "200",
        value: 200
      }, {
        label: "300",
        value: 300
      }];
    async function G() {
      try {
        const e = await din();
        g.value = e.filter(e => e.title && e.plain);
      } catch {
        g.value = [];
      }
    }
    const q = ml(() => g.value.map(e => ({
      label: e.title + (e.plain ? ` (${e.plain.length}字)` : ""),
      value: e.id
    })));
    Ma(G);
    const A = ml(() => "modal" === a.variant),
      T = [{
        key: "imitation",
        label: "仿写",
        icon: "imitation"
      }, {
        key: "fanfic",
        label: "同人",
        icon: "fanfic"
      }, {
        key: "spinoff",
        label: "番外",
        icon: "spinoff"
      }, {
        key: "chapters",
        label: "续写",
        icon: "chapters"
      }],
      z = ml(() => ({
        imitation: {
          title: "仿写创作",
          desc: "参考原文风格，创作全新原创故事",
          icon: "imitation"
        },
        fanfic: {
          title: "同人创作",
          desc: "基于原作世界观和角色进行二次创作",
          icon: "fanfic"
        },
        spinoff: {
          title: "番外创作",
          desc: "为已有作品创作外传、前传或平行篇",
          icon: "spinoff"
        },
        chapters: {
          title: "续写创作",
          desc: "导入已有章节，接续创作后续内容",
          icon: "chapters"
        }
      })[Q.value] || {
        title: "导入素材",
        desc: "快速创建项目",
        icon: "import"
      }),
      Q = ml({
        get: () => a.tab,
        set: e => o("update:tab", e)
      }),
      D = ml(() => (a.books || []).map(e => ({
        label: e.title || e.id,
        value: e.id
      }))),
      L = ml(() => "chapters" === Q.value ? "导入章节" : "创建作品"),
      j = ml(() => c.value.replace(/\s/g, "").length),
      F = ml(() => b.value.length > 1),
      B = ml(() => {
        const e = k.value.map(e => b.value[e]).filter(Boolean);
        return {
          count: e.length,
          totalWords: e.reduce((e, t) => e + (t.wordCount || 0), 0)
        };
      }),
      V = ml(() => {
        const e = b.value.length,
          t = [...k.value].sort((e, t) => e - t);
        if (!t.length || !e) return null;
        for (const n of [3, 5, 10]) {
          const a = Math.min(n, e),
            o = e - a,
            r = Array.from({
              length: a
            }, (e, t) => o + t);
          if (t.length === r.length && t.every((e, t) => e === r[t])) return n;
        }
        return "custom";
      }),
      U = ml(() => {
        const e = w.value.map(e => b.value[e]).filter(Boolean);
        return {
          count: e.length,
          totalWords: e.reduce((e, t) => e + (t.wordCount || 0), 0)
        };
      });
    function H() {
      i.value = "", s.value = "", c.value = "", u.value = [], d.value = null, v.value = null, p.value = "canon", f.value = "sequel", m.value = "", h.value = [], W.value = !1, y.value = [], b.value = [], k.value = [], _.value = 5e3, S.value = 100, C.value = "", P.value = "", R.value = null;
    }
    function J(e) {
      const t = String(e || "").trim();
      if (!t) return [];
      const n = t.split(/(?=第[0-9一二三四五六七八九十百千]+章[^\n]*\n?)/),
        a = [];
      for (const o of n) {
        const e = o.trim();
        if (!e) continue;
        const t = e.match(/^第[0-9一二三四五六七八九十百千]+章[^\n]*/),
          n = t ? t[0].trim() : `第${a.length + 1}章`,
          r = e.replace(/^第[0-9一二三四五六七八九十百千]+章[^\n]*\n?/, "").trim();
        r && a.push({
          title: n,
          content: r,
          wordCount: r.replace(/\s/g, "").length
        });
      }
      return !a.length && t && a.push({
        title: "第1章",
        content: t,
        wordCount: t.replace(/\s/g, "").length
      }), a;
    }
    async function Z(e) {
      var t;
      if (null == e ? void 0 : e.data) return String(e.data);
      const n = (null == e ? void 0 : e.path) || e;
      if (!n) return "";
      const a = null == (t = window.pywebview) ? void 0 : t.api;
      if (null == a ? void 0 : a.read_file_content) {
        const e = await a.read_file_content(n);
        return (null == e ? void 0 : e.success) && e.content ? e.content : ((null == e ? void 0 : e.message) && r.warning(e.message), "");
      }
      if (null == a ? void 0 : a.read_file) {
        return (await a.read_file(n)) || "";
      }
      return "";
    }
    async function X() {
      var e, t;
      if (!W.value && !l.value) try {
        const n = null == (e = window.pywebview) ? void 0 : e.api;
        if (!(null == n ? void 0 : n.novels_pick_style_files)) return void r.warning("当前环境不支持文件选择");
        W.value = !0;
        const a = await n.novels_pick_style_files();
        if (!(null == a ? void 0 : a.success)) return void ((null == a ? void 0 : a.message) && r.error(a.message));
        if (!(null == (t = a.data) ? void 0 : t.length)) return;
        let o = "",
          l = 0;
        for (const e of a.data) {
          const t = e.path || e,
            n = e.fileName || e.name || ("string" == typeof t ? t.split(/[/\\]/).pop() : "文件");
          if (y.value.some(e => e.path === t)) {
            r.info(`「${n}」已导入，跳过重复文件`);
            continue;
          }
          const a = await Z(e);
          (null == a ? void 0 : a.trim()) ? (o += (o ? "\n\n" : "") + a, y.value.push({
            name: n,
            path: t,
            size: e.size || 0
          }), l += 1) : r.warning(`未能读取文件内容：${n}`);
        }
        if (l > 0 && o) {
          const e = J(o);
          e.length > 1 ? (b.value = e, $(5), r.success(`已读取 ${l} 个文件，识别到 ${e.length} 个章节`)) : (b.value = [], k.value = [], c.value = o, r.success(`已读取 ${l} 个文件并填入参考原文`));
        }
      } catch {
        r.error("读取文件失败");
      } finally {
        W.value = !1;
      }
    }
    function $(e) {
      const t = b.value.length;
      if (0 === t) return;
      const n = Math.max(0, t - e);
      k.value = Array.from({
        length: t - n
      }, (e, t) => n + t), ee();
    }
    function ee() {
      const e = k.value.slice().sort((e, t) => e - t).map(e => b.value[e]).filter(Boolean);
      e.length ? c.value = e.map(e => `${e.title}\n\n${e.content}`).join("\n\n") : c.value = "";
    }
    function te() {
      w.value = [...k.value], x.value = !0;
    }
    function ne(e) {
      const t = w.value.indexOf(e);
      t >= 0 ? w.value.splice(t, 1) : w.value.push(e);
    }
    function ae() {
      w.value = b.value.map((e, t) => t);
    }
    function oe() {
      w.value = [];
    }
    function re() {
      k.value = [...w.value], ee(), x.value = !1;
    }
    function le() {
      A.value ? o("close") : o("back");
    }
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
    return Ro(() => a.tab, H), Ro(c, e => {
      O && clearTimeout(O);
      (e || "").replace(/\s/g, "").length < 100 ? R.value = null : O = setTimeout(() => {
        try {
          R.value = ain(e);
        } catch {
          R.value = null;
        }
      }, 600);
    }), t({
      loading: l
    }), (t, n) => (Cr(), Ir(br, null, [zr("div", {
      class: K(["import-shell", [e.variant, {
        dark: e.isDarkMode
      }]])
    }, [A.value ? (Cr(), Ir("header", gin, [zr("div", Win, [zr("div", yin, [Qr(D9t, {
      name: z.value.icon,
      size: 20
    }, null, 8, ["name"])]), zr("div", null, [zr("h2", null, Y(z.value.title), 1), zr("p", null, Y(z.value.desc), 1)]), zr("button", {
      type: "button",
      class: "close-btn",
      "aria-label": "关闭",
      onClick: le
    }, "×")])])) : (Cr(), Ir("header", hin, [zr("button", {
      type: "button",
      class: "back-btn",
      onClick: le
    }, "← 返回对话"), n[26] || (n[26] = zr("h2", null, "导入素材", -1)), n[27] || (n[27] = zr("p", null, "创建同人、番外、仿写或续写项目，可从文风设置多选参考风格。", -1))])), A.value ? Fr("", !0) : (Cr(), Ir("div", bin, [(Cr(), Ir(br, null, Va(T, e => zr("button", {
      key: e.key,
      type: "button",
      class: K(["tab-btn", {
        active: Q.value === e.key
      }]),
      onClick: t => Q.value = e.key
    }, [zr("span", null, Y(e.label), 1)], 10, kin)), 64))])), Qr(Tt(IE), {
      class: "form-scroll",
      style: E(A.value ? {
        maxHeight: "min(58vh, 520px)"
      } : void 0)
    }, {
      default: Sn(() => [zr("div", xin, ["imitation" === Q.value ? (Cr(), Ir(br, {
        key: 0
      }, [zr("div", win, [n[28] || (n[28] = zr("label", null, "暂定标题（可选）", -1)), Qr(Tt(r_), {
        value: i.value,
        "onUpdate:value": n[0] || (n[0] = e => i.value = e),
        placeholder: "留空则自动生成",
        disabled: l.value
      }, null, 8, ["value", "disabled"])]), zr("div", _in, [zr("div", Sin, [n[29] || (n[29] = zr("label", {
        class: "required"
      }, "参考原文", -1)), zr("span", {
        class: K(["ref-char-count", {
          ok: j.value >= 100
        }])
      }, Y(j.value) + " 字 ", 3)]), Qr(Tt(r_), {
        value: c.value,
        "onUpdate:value": n[1] || (n[1] = e => c.value = e),
        type: "textarea",
        rows: 6,
        placeholder: "粘贴参考小说全文（至少 100 字），或上传 txt / docx 自动填入",
        disabled: l.value
      }, null, 8, ["value", "disabled"]), n[30] || (n[30] = zr("div", {
        class: "field-hint"
      }, " 粘贴或上传要仿写的参考原文，系统将自动拆解情绪曲线与核心价值观 ", -1))]), zr("div", Cin, [n[31] || (n[31] = zr("label", null, "上传参考文件（可选）", -1)), zr("div", Pin, [zr("button", {
        type: "button",
        class: "attach-btn",
        disabled: l.value || W.value,
        onClick: X
      }, Y(W.value ? "正在读取…" : "+ 选择文件（docx / txt）"), 9, Rin), y.value.length ? (Cr(), Ir("div", Oin, [(Cr(!0), Ir(br, null, Va(y.value, (e, t) => (Cr(), Ir("div", {
        key: t,
        class: "attach-item"
      }, [zr("span", Min, "已读取 · " + Y(e.name), 1), zr("button", {
        type: "button",
        class: "att-remove",
        onClick: e => {
          return n = t, y.value.splice(n, 1), void (y.value.length || (b.value = [], k.value = []));
          var n;
        }
      }, "×", 8, Iin)]))), 128))])) : Fr("", !0), y.value.length && !F.value ? (Cr(), Ir("div", Nin, " 文件内容已自动填入上方参考原文 ")) : Fr("", !0)])]), F.value ? (Cr(), Ir("div", Gin, [n[34] || (n[34] = zr("label", null, "选择参考章节", -1)), zr("div", qin, [zr("div", Ain, [b.value.length >= 3 ? (Cr(), Ir("button", {
        key: 0,
        type: "button",
        class: K(["chapter-chip", {
          active: 3 === V.value
        }]),
        onClick: n[2] || (n[2] = e => $(3))
      }, "最近3章", 2)) : Fr("", !0), zr("button", {
        type: "button",
        class: K(["chapter-chip", {
          active: 5 === V.value
        }]),
        onClick: n[3] || (n[3] = e => $(5))
      }, "最近5章", 2), b.value.length > 10 ? (Cr(), Ir("button", {
        key: 1,
        type: "button",
        class: K(["chapter-chip", {
          active: 10 === V.value
        }]),
        onClick: n[4] || (n[4] = e => $(10))
      }, "最近10章", 2)) : Fr("", !0), zr("button", {
        type: "button",
        class: K(["chapter-chip outline", {
          active: "custom" === V.value
        }]),
        onClick: te
      }, "选择章节", 2)]), zr("div", Tin, [jr(" 共识别 " + Y(b.value.length) + " 章 · 已选择 ", 1), zr("strong", null, Y(B.value.count), 1), n[32] || (n[32] = jr(" 章，共 ", -1)), zr("strong", null, Y(B.value.totalWords.toLocaleString()), 1), n[33] || (n[33] = jr(" 字 ", -1))])])])) : Fr("", !0), R.value ? (Cr(), Ir("div", zin, [n[41] || (n[41] = zr("label", null, "文风统计指纹", -1)), zr("div", Qin, [zr("div", Din, [zr("span", Lin, [n[35] || (n[35] = zr("span", {
        class: "stat-label"
      }, "平均句长", -1)), zr("span", jin, Y(R.value.avgSentenceLength) + "字", 1)]), zr("span", Ein, [n[36] || (n[36] = zr("span", {
        class: "stat-label"
      }, "句长波动", -1)), zr("span", Fin, Y(R.value.sentenceLengthStdDev), 1)]), zr("span", Bin, [n[37] || (n[37] = zr("span", {
        class: "stat-label"
      }, "段落长度", -1)), zr("span", Vin, Y(R.value.avgParagraphLength) + "字", 1)]), zr("span", Uin, [n[38] || (n[38] = zr("span", {
        class: "stat-label"
      }, "词汇多样性", -1)), zr("span", Kin, Y(Math.round(100 * R.value.vocabularyDiversity)) + "%", 1)])]), R.value.topPatterns.length ? (Cr(), Ir("div", Hin, [n[39] || (n[39] = zr("span", {
        class: "stat-label"
      }, "高频句首：", -1)), zr("span", Jin, Y(R.value.topPatterns.join("、")), 1)])) : Fr("", !0), R.value.rhetoricalFeatures.length ? (Cr(), Ir("div", Zin, [n[40] || (n[40] = zr("span", {
        class: "stat-label"
      }, "修辞特征：", -1)), zr("span", Xin, Y(R.value.rhetoricalFeatures.join("、")), 1)])) : Fr("", !0)]), n[42] || (n[42] = zr("div", {
        class: "field-hint"
      }, "纯算法统计，不消耗 AI 额度；后端将基于 8 维度做深度定性拆解", -1))])) : Fr("", !0), zr("div", $in, [zr("div", Yin, [n[43] || (n[43] = zr("label", null, "每章字数", -1)), Qr(Tt(bL), {
        value: _.value,
        "onUpdate:value": n[5] || (n[5] = e => _.value = e),
        min: 1e3,
        max: 5e4,
        step: 500,
        "show-button": !1,
        placeholder: "5000",
        disabled: l.value,
        class: "word-count-input"
      }, null, 8, ["value", "disabled"]), zr("div", esn, [(Cr(), Ir(br, null, Va(I, e => zr("button", {
        key: e.value,
        type: "button",
        class: K(["tag-chip preset", {
          active: _.value === e.value
        }]),
        disabled: l.value,
        onClick: t => {
          return n = e.value, void (_.value = n);
          var n;
        }
      }, Y(e.label), 11, tsn)), 64))]), n[44] || (n[44] = zr("div", {
        class: "field-hint"
      }, "可自由填写每章字数，也可点选常用字数", -1))]), zr("div", nsn, [n[45] || (n[45] = zr("label", null, "目标章节数", -1)), Qr(Tt(bL), {
        value: S.value,
        "onUpdate:value": n[6] || (n[6] = e => S.value = e),
        min: 1,
        max: 500,
        step: 10,
        "show-button": !1,
        placeholder: "100",
        disabled: l.value,
        class: "word-count-input"
      }, null, 8, ["value", "disabled"]), zr("div", asn, [(Cr(), Ir(br, null, Va(N, e => zr("button", {
        key: e.value,
        type: "button",
        class: K(["tag-chip preset", {
          active: S.value === e.value
        }]),
        disabled: l.value,
        onClick: t => S.value = e.value
      }, Y(e.label), 11, osn)), 64))]), n[46] || (n[46] = zr("div", {
        class: "field-hint"
      }, "长篇默认 100 章，可按需调整", -1))])]), zr("div", rsn, [n[47] || (n[47] = zr("label", null, "读者偏好", -1)), Qr(Tt(r_), {
        value: C.value,
        "onUpdate:value": n[7] || (n[7] = e => C.value = e),
        placeholder: "填写题材偏好，如：玄幻、都市",
        clearable: "",
        disabled: l.value
      }, null, 8, ["value", "disabled"]), zr("div", lsn, [(Cr(), Ir(br, null, Va(M, e => zr("button", {
        key: e,
        type: "button",
        class: "tag-chip preset",
        disabled: l.value,
        onClick: t => function (e) {
          const t = C.value.trim();
          if (!t) return void (C.value = e);
          t.split(/[、,，\s]+/).filter(Boolean).includes(e) || (C.value = `${t}、${e}`);
        }(e)
      }, Y(e), 9, isn)), 64))]), n[48] || (n[48] = zr("div", {
        class: "field-hint"
      }, "可自由填写，点选标签会自动追加", -1))]), zr("div", ssn, [n[49] || (n[49] = zr("label", null, "原创故事想法（可选）", -1)), Qr(Tt(r_), {
        value: s.value,
        "onUpdate:value": n[8] || (n[8] = e => s.value = e),
        type: "textarea",
        rows: 3,
        placeholder: "剧情、主角、核心冲突（原创剧情）",
        disabled: l.value
      }, null, 8, ["value", "disabled"])]), zr("div", csn, [n[50] || (n[50] = zr("label", null, "仿写方向（可选）", -1)), Qr(Tt(r_), {
        value: m.value,
        "onUpdate:value": n[9] || (n[9] = e => m.value = e),
        type: "textarea",
        rows: 2,
        placeholder: "留空则默认采用拆书后的方向一；可填「方向二：…」",
        disabled: l.value
      }, null, 8, ["value", "disabled"])]), zr("div", usn, [n[51] || (n[51] = zr("label", null, "补充信息（可选）", -1)), Qr(Tt(r_), {
        value: P.value,
        "onUpdate:value": n[10] || (n[10] = e => P.value = e),
        type: "textarea",
        rows: 3,
        placeholder: "填写额外要求，如人设、情节走向、必须出现的桥段、禁忌点等",
        disabled: l.value
      }, null, 8, ["value", "disabled"])]), zr("div", dsn, [n[52] || (n[52] = zr("label", null, "文风设置（多选）", -1)), Qr(Tt(eO), {
        value: u.value,
        "onUpdate:value": n[11] || (n[11] = e => u.value = e)
      }, {
        default: Sn(() => [e.styleSources.length ? (Cr(), Ir("div", psn, [(Cr(!0), Ir(br, null, Va(e.styleSources, e => (Cr(), Ir("label", {
          key: e.id,
          class: "kb-card"
        }, [Qr(Tt(aO), {
          value: e.id
        }, null, 8, ["value"]), zr("div", fsn, [zr("span", msn, Y(e.title), 1), zr("span", hsn, Y(e.genre || "未分类"), 1)])]))), 128))])) : (Cr(), Ir("div", vsn, "暂无可用文风"))]),
        _: 1
      }, 8, ["value"])]), zr("div", gsn, [n[53] || (n[53] = zr("label", null, "关联备忘录（可选）", -1)), Qr(Tt(PM), {
        value: h.value,
        "onUpdate:value": n[12] || (n[12] = e => h.value = e),
        options: q.value,
        placeholder: "选择备忘录作为参考素材",
        multiple: "",
        clearable: "",
        disabled: l.value,
        onFocus: G
      }, null, 8, ["value", "options", "disabled"]), h.value.length ? (Cr(), Ir("div", Wsn, " 已选 " + Y(h.value.length) + " 条备忘录，内容将作为参考素材注入 ", 1)) : Fr("", !0)])], 64)) : (Cr(), Ir(br, {
        key: 1
      }, ["spinoff" === Q.value ? (Cr(), Ir("div", ysn, [n[54] || (n[54] = zr("label", null, "正传书籍", -1)), Qr(Tt(PM), {
        value: d.value,
        "onUpdate:value": n[13] || (n[13] = e => d.value = e),
        options: D.value,
        placeholder: "选择已有正传",
        clearable: ""
      }, null, 8, ["value", "options"])])) : Fr("", !0), "chapters" === Q.value ? (Cr(), Ir("div", bsn, [n[55] || (n[55] = zr("label", null, "导入到书籍（可选，不选则新建）", -1)), Qr(Tt(PM), {
        value: v.value,
        "onUpdate:value": n[14] || (n[14] = e => v.value = e),
        options: D.value,
        placeholder: "选择已有书籍或留空新建",
        clearable: ""
      }, null, 8, ["value", "options"])])) : Fr("", !0), "chapters" === Q.value && v.value ? Fr("", !0) : (Cr(), Ir("div", ksn, [zr("label", null, Y("chapters" === Q.value ? "新书书名" : "书名"), 1), Qr(Tt(r_), {
        value: i.value,
        "onUpdate:value": n[15] || (n[15] = e => i.value = e),
        placeholder: "暂定书名"
      }, null, 8, ["value"])])), "spinoff" === Q.value ? (Cr(), Ir("div", xsn, [n[56] || (n[56] = zr("label", null, "番外类型", -1)), Qr(Tt(PM), {
        value: f.value,
        "onUpdate:value": n[16] || (n[16] = e => f.value = e),
        options: [{
          label: "后传",
          value: "sequel"
        }, {
          label: "前传",
          value: "prequel"
        }, {
          label: "平行线",
          value: "parallel"
        }, {
          label: "If 线",
          value: "if"
        }]
      }, null, 8, ["value"])])) : Fr("", !0), "spinoff" === Q.value ? (Cr(), Ir("div", wsn, [n[57] || (n[57] = zr("label", null, "番外方向", -1)), Qr(Tt(r_), {
        value: s.value,
        "onUpdate:value": n[17] || (n[17] = e => s.value = e),
        type: "textarea",
        rows: 2,
        placeholder: "番外要写什么"
      }, null, 8, ["value"])])) : Fr("", !0), "fanfic" === Q.value ? (Cr(), Ir("div", _sn, [n[58] || (n[58] = zr("label", null, "同人方向", -1)), Qr(Tt(r_), {
        value: s.value,
        "onUpdate:value": n[18] || (n[18] = e => s.value = e),
        type: "textarea",
        rows: 2,
        placeholder: "CP、AU、IF 线等"
      }, null, 8, ["value"])])) : Fr("", !0), "fanfic" === Q.value ? (Cr(), Ir("div", Ssn, [n[59] || (n[59] = zr("label", null, "同人模式", -1)), Qr(Tt(PM), {
        value: p.value,
        "onUpdate:value": n[19] || (n[19] = e => p.value = e),
        options: [{
          label: "尊重原作 (canon)",
          value: "canon"
        }, {
          label: "AU / 架空",
          value: "au"
        }, {
          label: "OOC / 性格偏离",
          value: "ooc"
        }, {
          label: "CP / 关系向",
          value: "cp"
        }]
      }, null, 8, ["value"])])) : Fr("", !0), zr("div", Csn, [zr("label", null, Y("chapters" === Q.value ? "章节正文" : "参考文本（可选）"), 1), Qr(Tt(r_), {
        value: c.value,
        "onUpdate:value": n[20] || (n[20] = e => c.value = e),
        type: "textarea",
        rows: "chapters" === Q.value ? 8 : 4,
        placeholder: "chapters" === Q.value ? "粘贴章节，建议以「第1章」等标题分段" : "粘贴参考原文；也可只选下方文风"
      }, null, 8, ["value", "rows", "placeholder"])]), zr("div", Psn, [n[60] || (n[60] = zr("label", null, "文风设置（多选）", -1)), Qr(Tt(eO), {
        value: u.value,
        "onUpdate:value": n[21] || (n[21] = e => u.value = e)
      }, {
        default: Sn(() => [e.styleSources.length ? (Cr(), Ir("div", Osn, [(Cr(!0), Ir(br, null, Va(e.styleSources, e => (Cr(), Ir("label", {
          key: e.id,
          class: "kb-card"
        }, [Qr(Tt(aO), {
          value: e.id
        }, null, 8, ["value"]), zr("div", Msn, [zr("span", Isn, Y(e.title), 1), zr("span", Nsn, Y(e.genre || "未分类"), 1)])]))), 128))])) : (Cr(), Ir("div", Rsn, "暂无可用文风"))]),
        _: 1
      }, 8, ["value"])])], 64))])]),
      _: 1
    }, 8, ["style"]), zr("footer", {
      class: K(["form-footer", {
        modal: A.value
      }])
    }, [A.value ? (Cr(), Nr(Tt(WS), {
      key: 0,
      quaternary: "",
      onClick: le
    }, {
      default: Sn(() => [...(n[61] || (n[61] = [jr("取消", -1)]))]),
      _: 1
    })) : Fr("", !0), Qr(Tt(WS), {
      type: "primary",
      loading: l.value,
      class: "submit-btn",
      onClick: ie
    }, {
      default: Sn(() => [jr(Y(L.value), 1)]),
      _: 1
    }, 8, ["loading"])], 2)], 2), Qr(Tt(wq), {
      show: x.value,
      "onUpdate:show": n[25] || (n[25] = e => x.value = e),
      "auto-focus": !1,
      "mask-closable": !0,
      "transform-origin": "center"
    }, {
      default: Sn(() => [zr("div", {
        class: K(["chapter-dialog", {
          dark: e.isDarkMode
        }])
      }, [zr("header", Gsn, [n[62] || (n[62] = zr("div", {
        class: "chapter-dialog-title"
      }, [zr("div", {
        class: "cd-indicator"
      }), zr("h3", null, "选择关联章节")], -1)), zr("button", {
        type: "button",
        class: "close-btn",
        onClick: n[22] || (n[22] = e => x.value = !1)
      }, "×")]), zr("div", qsn, [zr("button", {
        type: "button",
        class: "cd-action-btn",
        onClick: ae
      }, " 全选 " + Y(b.value.length) + " 章 ", 1), zr("button", {
        type: "button",
        class: "cd-action-btn",
        onClick: oe
      }, "清空")]), Qr(Tt(IE), {
        class: "chapter-dialog-scroll",
        style: {
          maxHeight: "min(50vh, 400px)"
        }
      }, {
        default: Sn(() => [zr("div", Asn, [(Cr(!0), Ir(br, null, Va(b.value, (e, t) => (Cr(), Ir("label", {
          key: t,
          class: K(["chapter-item", {
            selected: w.value.includes(t)
          }]),
          onClick: Zi(e => ne(t), ["prevent"])
        }, [Qr(Tt(aO), {
          checked: w.value.includes(t),
          onClick: n[23] || (n[23] = Zi(() => {}, ["stop"])),
          "onUpdate:checked": e => ne(t)
        }, null, 8, ["checked", "onUpdate:checked"]), zr("span", zsn, Y(e.title), 1), zr("span", Qsn, Y((e.wordCount || 0).toLocaleString()) + "字", 1)], 10, Tsn))), 128))])]),
        _: 1
      }), zr("footer", Dsn, [zr("div", Lsn, [n[63] || (n[63] = zr("span", {
        class: "cd-dot blue"
      }, null, -1)), jr(" 已选择 " + Y(U.value.count) + " 章 ", 1), n[64] || (n[64] = zr("span", {
        class: "cd-dot green"
      }, null, -1)), jr(" 共 " + Y(U.value.totalWords.toLocaleString()) + " 字 ", 1)]), zr("div", jsn, [Qr(Tt(WS), {
        quaternary: "",
        onClick: n[24] || (n[24] = e => x.value = !1)
      }, {
        default: Sn(() => [...(n[65] || (n[65] = [jr("取消", -1)]))]),
        _: 1
      }), Qr(Tt(WS), {
        type: "primary",
        onClick: re
      }, {
        default: Sn(() => [...(n[66] || (n[66] = [jr("确认选择", -1)]))]),
        _: 1
      })])])], 2)]),
      _: 1
    }, 8, ["show"])], 64));
  }
};
