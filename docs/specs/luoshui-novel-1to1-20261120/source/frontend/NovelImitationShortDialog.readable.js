// Inspection only. Original SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Original UTF16 offsets [8832348, 8840980); original line 41
// External dependencies and runtime equivalence remain unresolved.
const recoveredComponent = {
  __name: "NovelImitationShortDialog",
  props: {
    show: Boolean,
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
  emits: ["update:show", "submit"],
  setup(e, {
    emit: t
  }) {
    const n = e,
      a = t,
      o = $q(),
      r = Nt(!1),
      l = Nt(!1),
      i = Nt(""),
      s = Nt([]),
      c = Nt([]),
      u = Nt(13e3),
      d = Nt(1),
      v = Nt(""),
      p = Nt(""),
      f = Nt([]),
      m = Nt(null),
      h = Nt([]),
      g = Nt(""),
      W = Nt([]);
    async function y() {
      try {
        const e = await din();
        W.value = e.filter(e => e.title && e.plain);
      } catch {
        W.value = [];
      }
    }
    const b = ml(() => W.value.map(e => ({
      label: e.title + (e.plain ? ` (${e.plain.length}字)` : ""),
      value: e.id
    })));
    Ma(y);
    const k = ["轻松搞笑", "热血爽文", "智斗权谋", "悬疑惊悚", "甜宠治愈"],
      x = [{
        label: "1.0万",
        value: 1e4
      }, {
        label: "1.1万",
        value: 11e3
      }, {
        label: "1.2万",
        value: 12e3
      }, {
        label: "1.3万",
        value: 13e3
      }, {
        label: "1.5万",
        value: 15e3
      }, {
        label: "2.0万",
        value: 2e4
      }],
      w = ["女频", "男频", "丧尸", "萌娃", "萌宠穿越", "通古今", "直播", "历史", "系统", "修仙", "搞笑", "重生", "穿越", "都市", "校园", "豪门", "职场", "追妻火葬场", "大女主", "虐恋", "BE", "HE"],
      _ = ml(() => (n.books || []).map(e => ({
        label: e.title || e.id,
        value: e.id
      }))),
      S = ml(() => i.value.replace(/\s/g, "").length),
      C = ml(() => S.value >= 100);
    async function P(e) {
      var t;
      if (null == e ? void 0 : e.data) return String(e.data);
      const n = (null == e ? void 0 : e.path) || e;
      if (!n) return "";
      const a = null == (t = window.pywebview) ? void 0 : t.api;
      if (null == a ? void 0 : a.read_file_content) {
        const e = await a.read_file_content(n);
        return (null == e ? void 0 : e.success) && e.content ? e.content : ((null == e ? void 0 : e.message) && o.warning(e.message), "");
      }
      if (null == a ? void 0 : a.read_file) {
        return (await a.read_file(n)) || "";
      }
      return "";
    }
    function R(e, t) {
      const n = String(t || "").trim();
      return !!n && (i.value.trim() ? i.value = `${i.value.trim()}\n\n--- ${e} ---\n\n${n}` : i.value = n, !0);
    }
    function O() {
      a("update:show", !1);
    }
    function M() {
      const e = Number(u.value);
      return !Number.isFinite(e) || e < 1e3 ? 13e3 : Math.min(5e5, Math.round(e));
    }
    function I() {
      const e = Number(d.value);
      return !Number.isFinite(e) || e < 1 ? 1 : Math.min(50, Math.round(e));
    }
    async function N() {
      var e, t;
      if (!l.value && !r.value) try {
        const n = null == (e = window.pywebview) ? void 0 : e.api;
        if (!(null == n ? void 0 : n.novels_pick_style_files)) return void o.warning("当前环境不支持文件选择");
        l.value = !0;
        const a = await n.novels_pick_style_files();
        if (!(null == a ? void 0 : a.success)) return void ((null == a ? void 0 : a.message) && o.error(a.message));
        if (!(null == (t = a.data) ? void 0 : t.length)) return;
        let r = 0;
        for (const e of a.data) {
          const t = e.path || e,
            n = e.fileName || e.name || ("string" == typeof t ? t.split(/[/\\]/).pop() : "文件");
          if (s.value.some(e => e.path === t)) {
            o.info(`「${n}」已导入，跳过重复文件`);
            continue;
          }
          R(n, await P(e)) ? (s.value.push({
            name: n,
            path: t,
            size: e.size || 0
          }), r += 1) : o.warning(`未能读取文件内容：${n}`);
        }
        r > 0 && o.success(`已读取 ${r} 个文件并填入参考原文`);
      } catch (Ofr) {
        o.error("读取文件失败");
      } finally {
        l.value = !1;
      }
    }
    async function G() {
      if (!C.value) return void o.warning("参考原文至少需要 100 字");
      let e = i.value.trim();
      if (h.value.length) for (const n of h.value) {
        const t = W.value.find(e => e.id === n);
        (null == t ? void 0 : t.plain) && (e += `\n\n--- 备忘录: ${t.title} ---\n\n${t.plain}`);
      }
      const t = {
        referenceText: e,
        title: g.value.trim() || void 0,
        targetWordCount: M(),
        targetChapters: I(),
        chapters: I(),
        supplementaryInfo: v.value.trim() || void 0,
        readerPreferences: p.value || void 0,
        inspirationTags: f.value.join("、") || void 0,
        styleName: p.value || void 0,
        linkedSaveId: m.value || void 0,
        knowledgeBookIds: c.value.length ? c.value : void 0
      };
      a("submit", t), i.value = "", s.value = [], c.value = [], u.value = 13e3, d.value = 1, v.value = "", p.value = "", f.value = [], m.value = null, h.value = [], g.value = "", O();
    }
    return (t, n) => (Cr(), Nr(Tt(wq), {
      show: e.show,
      "auto-focus": !1,
      "mask-closable": !r.value,
      "transform-origin": "center",
      "onUpdate:show": n[9] || (n[9] = e => a("update:show", e))
    }, {
      default: Sn(() => [zr("div", {
        class: K(["imit-short-card", {
          dark: e.isDarkMode
        }])
      }, [zr("header", Vsn, [zr("div", Usn, [zr("div", Ksn, [Qr(D9t, {
        name: "imitation",
        size: 20
      })]), n[10] || (n[10] = zr("div", {
        class: "header-text"
      }, [zr("h2", null, "仿写短篇"), zr("p", null, "参考原文风格，一键生成知乎盐选风短篇小说")], -1)), zr("button", {
        type: "button",
        class: "close-btn",
        disabled: r.value,
        onClick: O
      }, "×", 8, Hsn)])]), Qr(Tt(IE), {
        class: "card-scroll",
        style: {
          maxHeight: "min(58vh, 520px)"
        }
      }, {
        default: Sn(() => [zr("div", Jsn, [zr("div", Zsn, [n[11] || (n[11] = zr("label", null, "暂定标题（可选）", -1)), Qr(Tt(r_), {
          value: g.value,
          "onUpdate:value": n[0] || (n[0] = e => g.value = e),
          placeholder: "留空则自动生成"
        }, null, 8, ["value"])]), zr("div", Xsn, [zr("div", $sn, [n[12] || (n[12] = zr("label", {
          class: "required"
        }, "参考原文", -1)), zr("span", {
          class: K(["ref-char-count", {
            ok: S.value >= 100
          }])
        }, Y(S.value) + " 字 ", 3)]), Qr(Tt(r_), {
          value: i.value,
          "onUpdate:value": n[1] || (n[1] = e => i.value = e),
          type: "textarea",
          rows: 6,
          placeholder: "粘贴参考小说全文（至少 100 字），或上传 txt / docx 自动填入",
          disabled: r.value
        }, null, 8, ["value", "disabled"]), n[13] || (n[13] = zr("div", {
          class: "field-hint"
        }, " 粘贴或上传要仿写的参考原文，系统将自动拆解情绪曲线与核心价值观 ", -1))]), zr("div", Ysn, [n[14] || (n[14] = zr("label", null, "上传参考文件（可选）", -1)), zr("div", ecn, [zr("button", {
          type: "button",
          class: "attach-btn",
          disabled: r.value || l.value,
          onClick: N
        }, Y(l.value ? "正在读取…" : "+ 选择文件（docx / txt）"), 9, tcn), s.value.length ? (Cr(), Ir("div", ncn, [(Cr(!0), Ir(br, null, Va(s.value, (e, t) => (Cr(), Ir("div", {
          key: t,
          class: "attach-item"
        }, [zr("span", acn, "已读取 · " + Y(e.name), 1), zr("button", {
          type: "button",
          class: "att-remove",
          onClick: e => {
            return n = t, void s.value.splice(n, 1);
            var n;
          }
        }, "×", 8, ocn)]))), 128))])) : Fr("", !0), s.value.length ? (Cr(), Ir("div", rcn, " 文件内容已自动填入上方参考原文，可继续编辑 ")) : Fr("", !0)])]), zr("div", lcn, [zr("div", icn, [n[15] || (n[15] = zr("label", null, "目标字数", -1)), Qr(Tt(bL), {
          value: u.value,
          "onUpdate:value": n[2] || (n[2] = e => u.value = e),
          min: 1e3,
          max: 5e5,
          step: 500,
          "show-button": !1,
          placeholder: "如 13000",
          disabled: r.value,
          class: "word-count-input"
        }, null, 8, ["value", "disabled"]), zr("div", scn, [(Cr(), Ir(br, null, Va(x, e => zr("button", {
          key: e.value,
          type: "button",
          class: K(["tag-chip preset", {
            active: u.value === e.value
          }]),
          disabled: r.value,
          onClick: t => {
            return n = e.value, void (u.value = n);
            var n;
          }
        }, Y(e.label), 11, ccn)), 64))]), n[16] || (n[16] = zr("div", {
          class: "field-hint"
        }, "可自由填写字数（字），也可点选常用字数", -1))]), zr("div", ucn, [n[17] || (n[17] = zr("label", null, "读者偏好", -1)), Qr(Tt(r_), {
          value: p.value,
          "onUpdate:value": n[3] || (n[3] = e => p.value = e),
          placeholder: "填写读者偏好，如：轻松搞笑、甜宠治愈",
          clearable: "",
          disabled: r.value
        }, null, 8, ["value", "disabled"]), zr("div", dcn, [(Cr(), Ir(br, null, Va(k, e => zr("button", {
          key: e,
          type: "button",
          class: "tag-chip preset",
          disabled: r.value,
          onClick: t => function (e) {
            const t = p.value.trim();
            if (!t) return void (p.value = e);
            t.split(/[、,，\s]+/).filter(Boolean).includes(e) || (p.value = `${t}、${e}`);
          }(e)
        }, Y(e), 9, vcn)), 64))]), n[18] || (n[18] = zr("div", {
          class: "field-hint"
        }, "可自由填写，点选标签会自动追加", -1))])]), zr("div", pcn, [zr("div", fcn, [n[19] || (n[19] = zr("label", null, "目标章节", -1)), Qr(Tt(bL), {
          value: d.value,
          "onUpdate:value": n[4] || (n[4] = e => d.value = e),
          min: 1,
          max: 50,
          step: 1,
          "show-button": !1,
          placeholder: "1",
          disabled: r.value,
          class: "word-count-input"
        }, null, 8, ["value", "disabled"]), n[20] || (n[20] = zr("div", {
          class: "field-hint"
        }, "默认 1 章，可按需调整", -1))])]), zr("div", mcn, [n[21] || (n[21] = zr("label", null, "补充信息（可选）", -1)), Qr(Tt(r_), {
          value: v.value,
          "onUpdate:value": n[5] || (n[5] = e => v.value = e),
          type: "textarea",
          rows: 3,
          placeholder: "填写额外要求，如人设、情节走向、必须出现的桥段、禁忌点等",
          disabled: r.value
        }, null, 8, ["value", "disabled"])]), zr("div", hcn, [n[22] || (n[22] = zr("label", null, "灵感标签（可选）", -1)), zr("div", gcn, [(Cr(), Ir(br, null, Va(w, e => zr("button", {
          key: e,
          type: "button",
          class: K(["tag-chip", {
            active: f.value.includes(e)
          }]),
          disabled: r.value,
          onClick: t => function (e) {
            const t = f.value.indexOf(e);
            t >= 0 ? f.value.splice(t, 1) : f.value.push(e);
          }(e)
        }, Y(e), 11, Wcn)), 64))])]), zr("div", ycn, [n[23] || (n[23] = zr("label", null, "文风设置（多选）", -1)), Qr(Tt(eO), {
          value: c.value,
          "onUpdate:value": n[6] || (n[6] = e => c.value = e)
        }, {
          default: Sn(() => [e.styleSources.length ? (Cr(), Ir("div", kcn, [(Cr(!0), Ir(br, null, Va(e.styleSources, e => (Cr(), Ir("label", {
            key: e.id,
            class: "kb-card"
          }, [Qr(Tt(aO), {
            value: e.id
          }, null, 8, ["value"]), zr("div", xcn, [zr("span", wcn, Y(e.title), 1), zr("span", _cn, Y(e.genre || "未分类"), 1)])]))), 128))])) : (Cr(), Ir("div", bcn, "暂无可用文风"))]),
          _: 1
        }, 8, ["value"])]), zr("div", Scn, [n[24] || (n[24] = zr("label", null, "关联备忘录（可选）", -1)), Qr(Tt(PM), {
          value: h.value,
          "onUpdate:value": n[7] || (n[7] = e => h.value = e),
          options: b.value,
          placeholder: "选择备忘录作为参考素材",
          multiple: "",
          clearable: "",
          disabled: r.value,
          onFocus: y
        }, null, 8, ["value", "options", "disabled"]), h.value.length ? (Cr(), Ir("div", Ccn, " 已选 " + Y(h.value.length) + " 条备忘录，内容将作为参考素材注入 ", 1)) : Fr("", !0)]), zr("div", Pcn, [n[25] || (n[25] = zr("label", null, "关联存档（可选）", -1)), Qr(Tt(PM), {
          value: m.value,
          "onUpdate:value": n[8] || (n[8] = e => m.value = e),
          options: _.value,
          placeholder: "关联已有作品",
          clearable: "",
          disabled: r.value
        }, null, 8, ["value", "options", "disabled"])])])]),
        _: 1
      }), zr("footer", Rcn, [zr("div", Ocn, [f.value.length ? (Cr(), Ir("span", Mcn, Y(f.value.length) + " 个标签 ", 1)) : Fr("", !0)]), zr("div", Icn, [Qr(Tt(WS), {
        quaternary: "",
        disabled: r.value,
        onClick: O
      }, {
        default: Sn(() => [...(n[26] || (n[26] = [jr("取消", -1)]))]),
        _: 1
      }, 8, ["disabled"]), Qr(Tt(WS), {
        type: "primary",
        disabled: !C.value,
        onClick: G
      }, {
        default: Sn(() => [...(n[27] || (n[27] = [jr(" 开始仿写 ", -1)]))]),
        _: 1
      }, 8, ["disabled"])])])], 2)]),
      _: 1
    }, 8, ["show", "mask-closable"]));
  }
};
