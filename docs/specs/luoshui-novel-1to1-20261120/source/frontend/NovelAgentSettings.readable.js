// Inspection only. Original SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Original UTF16 offsets [8744349, 8756515); original line 41
// External dependencies and runtime equivalence remain unresolved.
const recoveredComponent = {
  __name: "NovelAgentSettings",
  props: {
    modelOptions: {
      type: Array,
      default: () => []
    },
    isDarkMode: Boolean
  },
  emits: ["back"],
  setup(e, {
    emit: t
  }) {
    const n = e,
      a = t,
      o = $q(),
      r = Nt(!0),
      l = Nt(!1),
      i = Nt(!1),
      s = Nt({}),
      c = Nt([]),
      u = Nt([]),
      d = Nt({
        maxRetries: 3,
        passScore: 80,
        temperatureStep: .15,
        baseTemperature: .65,
        enableLengthNorm: !0,
        enablePostWrite: !0,
        enableAiTells: !0,
        requireSecondaryConfirm: !1
      }),
      v = ml(() => [{
        label: "跟随对话默认模型",
        value: "",
        displayName: "跟随对话默认模型"
      }, ...n.modelOptions.map(e => ({
        label: e.label,
        value: e.value,
        displayName: e.displayName,
        channelTag: e.channelTag,
        channelKey: e.channelKey
      }))]),
      p = ml(() => function ({
        codexReady: e = !1
      } = {}) {
        const t = 187,
          n = "DfmN",
          a = "fmRD",
          o = 136,
          r = "^wBz",
          l = 240,
          i = 199,
          s = 127,
          c = "ZOSs",
          u = "g6jQ",
          d = 12,
          v = "3sTD",
          p = 11,
          f = "l0cF",
          m = 214,
          h = {};
        function g(e, t, n, a) {
          return Trn(0, t - -930, 0, e);
        }
        h[g("#spz", 102)] = function (e, t) {
          return e !== t;
        }, h[g("4A&B", -48)] = g("4A&B", t);
        const W = h,
          y = {};
        function b(e, t, n, a) {
          return Trn(0, t - -997, 0, n);
        }
        y[g(n, -28)] = g(a, o) + Vrn + "）", y[b(-82, -133, r)] = "";
        const k = [y, ...Brn];
        if (e) if (W[g("UnKA", l)](W[g("#eoy", i)], W[g("4A&B", -48)])) {
          if (_0x3606c5) {
            const e = _0x1ae87e[b(0, s, c)](_0x42356c, arguments);
            return _0x3a0d2a = null, e;
          }
        } else {
          const e = {};
          e[g(u, d)] = Qrn, e[g(v, p)] = zrn, k[g(f, m) + "e"](1, 0, e);
        }
        return k;
      }({
        codexReady: i.value
      })),
      f = ml(() => c.value.filter(e => "dimension" === e.kind || String(e.id || "").startsWith("dim_"))),
      m = ml(() => c.value.filter(e => "custom" === e.kind || !String(e.id || "").startsWith("dim_") && "dimension" !== e.kind)),
      h = ml(() => f.value.filter(e => e.enabled).length);
    function g(e) {
      return "dimension" === e.kind || String(e.id || "").startsWith("dim_");
    }
    function W(e) {
      c.value = c.value.map(t => g(t) ? {
        ...t,
        enabled: e
      } : t);
    }
    function y(e) {
      return g(e) ? {
        id: e.id,
        dimensionId: e.dimensionId,
        title: e.title,
        content: e.content,
        enabled: Boolean(e.enabled),
        kind: "dimension"
      } : {
        id: e.id,
        title: e.title,
        content: e.content,
        enabled: Boolean(e.enabled),
        kind: "custom"
      };
    }
    async function b() {
      var e, t, n, a;
      l.value = !0;
      try {
        const r = await async function (e) {
          var t, n;
          const a = 823,
            o = 762,
            r = "VQi*",
            l = 193,
            i = "$OLu",
            s = 554,
            c = {};
          function u(e, t, n, a) {
            return Gan(0, a, n - s);
          }
          return c.success = !1, (null == (n = null == (t = Ran()) ? void 0 : t[u(0, 0, a, "$OLu") + u(0, 0, o, r) + "e_age" + (d = l, v = i, Gan(0, v, d - -229) + "ttings")]) ? void 0 : n.call(t, e)) ?? c;
          var d, v;
        }({
          modelRoutes: {
            ...s.value
          },
          reviewPrompts: c.value.map(y),
          customWritePrompts: u.value.map(e => ({
            id: e.id,
            category: e.category,
            title: e.title,
            content: e.content,
            enabled: Boolean(e.enabled)
          })),
          reviewCycle: {
            maxRetries: Math.max(0, Math.min(5, Number(d.value.maxRetries) || 3)),
            passScore: Math.max(0, Math.min(100, Number(d.value.passScore) || 80)),
            temperatureStep: Math.max(0, Math.min(.5, Number(d.value.temperatureStep) || .15)),
            baseTemperature: Math.max(0, Math.min(2, Number(d.value.baseTemperature) || .65)),
            enableLengthNorm: Boolean(d.value.enableLengthNorm),
            enablePostWrite: Boolean(d.value.enablePostWrite),
            enableAiTells: Boolean(d.value.enableAiTells),
            requireSecondaryConfirm: Boolean(d.value.requireSecondaryConfirm)
          }
        });
        if (!(null == r ? void 0 : r.success)) return void o.error((null == r ? void 0 : r.message) || "保存失败");
        s.value = {
          ...((null == (e = r.data) ? void 0 : e.modelRoutes) || {})
        }, c.value = ((null == (t = r.data) ? void 0 : t.reviewPrompts) || []).map(e => ({
          ...e
        })), u.value = ((null == (n = r.data) ? void 0 : n.customWritePrompts) || []).map(e => ({
          ...e
        }));
        const l = (null == (a = r.data) ? void 0 : a.reviewCycle) || {};
        d.value = {
          maxRetries: Number(l.maxRetries ?? 3),
          passScore: Number(l.passScore ?? 80),
          temperatureStep: Number(l.temperatureStep ?? .15),
          baseTemperature: Number(l.baseTemperature ?? .65),
          enableLengthNorm: !1 !== l.enableLengthNorm,
          enablePostWrite: !1 !== l.enablePostWrite,
          enableAiTells: !1 !== l.enableAiTells,
          requireSecondaryConfirm: Boolean(l.requireSecondaryConfirm)
        }, o.success("已保存");
      } finally {
        l.value = !1;
      }
    }
    function k() {
      const e = `custom_${Date.now()}`;
      c.value = [...c.value, {
        id: e,
        title: "自定义审稿",
        content: "",
        enabled: !0,
        kind: "custom"
      }];
    }
    function x(e, t, n) {
      c.value = c.value.map(a => a.id === e.id ? {
        ...a,
        [t]: n
      } : a);
    }
    function w(e, t, n) {
      u.value = u.value.map(a => a.id === e.id ? {
        ...a,
        [t]: n
      } : a);
    }
    function _() {
      const e = `wp_custom_${Date.now()}`;
      u.value = [...u.value, {
        id: e,
        category: "long_fiction",
        title: "自定义写作要求",
        content: "",
        enabled: !0,
        placeholder: ""
      }];
    }
    const S = [{
      label: "长篇正文",
      value: "long_fiction"
    }, {
      label: "短篇小说",
      value: "short_fiction"
    }, {
      label: "仿写创作",
      value: "imitation"
    }];
    return Ro(() => n.modelOptions.length, () => {
      Mrn.forEach(e => {
        e.key in s.value || (s.value[e.key] = "");
      });
    }), Ma(async function () {
      var e, t, n, a, l;
      r.value = !0;
      try {
        const r = null == (e = window.pywebview) ? void 0 : e.api,
          v = await oae(r);
        i.value = Boolean(null == v ? void 0 : v.ready);
        const p = await async function () {
          var e, t;
          const n = "!jb8",
            a = "6S(U",
            o = "VXV*",
            r = {};
          function l(e, t, n, a) {
            return Gan(0, t, e - -706);
          }
          return r[l(-292, "&2iX") + "ss"] = !1, (null == (t = null == (e = Ran()) ? void 0 : e[l(-392, n) + l(-97, "MEJ&") + "_agen" + (i = a, s = 18, Tan(0, 0, s - -64, i)) + l(-290, o)]) ? void 0 : t.call(e)) ?? r;
          var i, s;
        }();
        if (!(null == p ? void 0 : p.success)) return void o.error((null == p ? void 0 : p.message) || "加载 Agent 设置失败");
        s.value = {
          ...((null == (t = p.data) ? void 0 : t.modelRoutes) || {})
        }, c.value = ((null == (n = p.data) ? void 0 : n.reviewPrompts) || []).map(e => ({
          ...e
        })), u.value = ((null == (a = p.data) ? void 0 : a.customWritePrompts) || []).map(e => ({
          ...e
        }));
        const f = (null == (l = p.data) ? void 0 : l.reviewCycle) || {};
        d.value = {
          maxRetries: Number(f.maxRetries ?? 3),
          passScore: Number(f.passScore ?? 80),
          temperatureStep: Number(f.temperatureStep ?? .15),
          baseTemperature: Number(f.baseTemperature ?? .65),
          enableLengthNorm: !1 !== f.enableLengthNorm,
          enablePostWrite: !1 !== f.enablePostWrite,
          enableAiTells: !1 !== f.enableAiTells,
          requireSecondaryConfirm: Boolean(f.requireSecondaryConfirm)
        };
      } finally {
        r.value = !1;
      }
    }), (t, n) => (Cr(), Ir("div", {
      class: K(["agent-settings-page", {
        dark: e.isDarkMode
      }])
    }, [zr("header", Hrn, [zr("button", {
      type: "button",
      class: "back-btn",
      onClick: n[0] || (n[0] = e => a("back"))
    }, "← 返回对话"), n[11] || (n[11] = zr("h2", null, "Agent 设置", -1)), n[12] || (n[12] = zr("p", null, "为各节点指定模型路由（如写作用好模型、审稿用便宜模型）。审稿维度与自定义提示词会注入审稿节点。", -1))]), Qr(Tt(IE), {
      class: "page-body"
    }, {
      default: Sn(() => [zr("section", Jrn, [n[13] || (n[13] = zr("h3", null, "模型路由", -1)), n[14] || (n[14] = zr("p", {
        class: "panel-hint"
      }, " 留空则使用对话区当前选择的默认模型。封面配图使用出图模型，与对话 LLM 无关。 Codex 出图需先在技能页开启 Codex 渠道并完成验证。 ", -1)), r.value ? (Cr(), Ir("div", Zrn, "加载中…")) : (Cr(), Ir("div", Xrn, [(Cr(!0), Ir(br, null, Va(Tt(Mrn), t => (Cr(), Ir("div", {
        key: t.key,
        class: "route-row"
      }, [zr("span", $rn, Y(t.label), 1), "cover" !== t.key ? (Cr(), Nr($re, {
        key: 0,
        modelValue: s.value[t.key],
        "onUpdate:modelValue": e => s.value[t.key] = e,
        options: v.value,
        placeholder: "跟随默认",
        size: "small",
        class: "route-select",
        clearable: "",
        "is-dark-mode": !!e.isDarkMode
      }, null, 8, ["modelValue", "onUpdate:modelValue", "options", "is-dark-mode"])) : (Cr(), Nr(Tt(PM), {
        key: 1,
        value: s.value[t.key],
        "onUpdate:value": e => s.value[t.key] = e,
        options: p.value,
        placeholder: "默认 gemini 出图",
        size: "small",
        class: "route-select",
        filterable: "",
        clearable: ""
      }, null, 8, ["value", "onUpdate:value", "options"]))]))), 128))]))]), zr("section", Yrn, [n[27] || (n[27] = zr("h3", null, "写章审稿循环", -1)), n[28] || (n[28] = zr("p", {
        class: "panel-hint"
      }, "写下一章时自动审稿-修订循环。未达通过分将尝试修订，修订后仍可能回退到更高分版本。", -1)), r.value ? Fr("", !0) : (Cr(), Ir("div", eln, [zr("label", tln, [n[15] || (n[15] = zr("span", null, "最大修订轮次", -1)), Qr(Tt(bL), {
        value: d.value.maxRetries,
        "onUpdate:value": n[1] || (n[1] = e => d.value.maxRetries = e),
        min: 0,
        max: 5,
        size: "small"
      }, null, 8, ["value"])]), zr("label", nln, [n[16] || (n[16] = zr("span", null, "通过分数", -1)), Qr(Tt(bL), {
        value: d.value.passScore,
        "onUpdate:value": n[2] || (n[2] = e => d.value.passScore = e),
        min: 0,
        max: 100,
        size: "small"
      }, null, 8, ["value"])]), zr("label", aln, [n[17] || (n[17] = zr("span", null, "升温步长", -1)), Qr(Tt(bL), {
        value: d.value.temperatureStep,
        "onUpdate:value": n[3] || (n[3] = e => d.value.temperatureStep = e),
        min: 0,
        max: .5,
        step: .05,
        size: "small"
      }, null, 8, ["value"])]), zr("label", oln, [n[18] || (n[18] = zr("span", null, "基础温度", -1)), Qr(Tt(bL), {
        value: d.value.baseTemperature,
        "onUpdate:value": n[4] || (n[4] = e => d.value.baseTemperature = e),
        min: 0,
        max: 2,
        step: .1,
        size: "small"
      }, null, 8, ["value"])])])), r.value ? Fr("", !0) : (Cr(), Ir("div", rln, [zr("label", lln, [Qr(Tt(uF), {
        value: d.value.enableLengthNorm,
        "onUpdate:value": n[5] || (n[5] = e => d.value.enableLengthNorm = e),
        size: "small"
      }, null, 8, ["value"]), n[19] || (n[19] = zr("span", null, "字数归一化", -1)), n[20] || (n[20] = zr("span", {
        class: "switch-hint"
      }, "草稿字数偏离目标 >45% 时自动扩写/压缩", -1))]), zr("label", iln, [Qr(Tt(uF), {
        value: d.value.enablePostWrite,
        "onUpdate:value": n[6] || (n[6] = e => d.value.enablePostWrite = e),
        size: "small"
      }, null, 8, ["value"]), n[21] || (n[21] = zr("span", null, "后写校验", -1)), n[22] || (n[22] = zr("span", {
        class: "switch-hint"
      }, "确定性规则检查（零 LLM 开销）", -1))]), zr("label", sln, [Qr(Tt(uF), {
        value: d.value.enableAiTells,
        "onUpdate:value": n[7] || (n[7] = e => d.value.enableAiTells = e),
        size: "small"
      }, null, 8, ["value"]), n[23] || (n[23] = zr("span", null, "AI 味检测", -1)), n[24] || (n[24] = zr("span", {
        class: "switch-hint"
      }, "段落等长/套话密度/公式化转折（零 LLM 开销）", -1))]), zr("label", cln, [Qr(Tt(uF), {
        value: d.value.requireSecondaryConfirm,
        "onUpdate:value": n[8] || (n[8] = e => d.value.requireSecondaryConfirm = e),
        size: "small"
      }, null, 8, ["value"]), n[25] || (n[25] = zr("span", null, "生产二次确认", -1)), n[26] || (n[26] = zr("span", {
        class: "switch-hint"
      }, "大纲/蓝图/基础设定生成后暂停，可在右侧修改文件后再继续", -1))])]))]), zr("section", uln, [zr("div", dln, [zr("div", null, [n[29] || (n[29] = zr("h3", null, "审稿维度（37 项）", -1)), zr("p", vln, " 勾选后即参与审稿；维度说明为内置文案，不可改字，但可自由开关。当前已选 " + Y(h.value) + " / " + Y(f.value.length) + " 项。 ", 1)]), zr("div", pln, [Qr(Tt(WS), {
        size: "tiny",
        quaternary: "",
        onClick: n[9] || (n[9] = e => W(!0))
      }, {
        default: Sn(() => [...(n[30] || (n[30] = [jr("全选", -1)]))]),
        _: 1
      }), Qr(Tt(WS), {
        size: "tiny",
        quaternary: "",
        onClick: n[10] || (n[10] = e => W(!1))
      }, {
        default: Sn(() => [...(n[31] || (n[31] = [jr("全不选", -1)]))]),
        _: 1
      })])]), r.value ? Fr("", !0) : (Cr(), Ir("div", fln, [(Cr(!0), Ir(br, null, Va(f.value, e => (Cr(), Ir("div", {
        key: e.id,
        class: "dimension-row"
      }, [Qr(Tt(aO), {
        checked: Boolean(e.enabled),
        "onUpdate:checked": t => function (e, t) {
          c.value = c.value.map(n => n.id === e.id ? {
            ...n,
            enabled: t
          } : n);
        }(e, t)
      }, null, 8, ["checked", "onUpdate:checked"]), zr("div", mln, [zr("div", hln, [zr("span", gln, Y(e.dimensionId) + ".", 1), zr("span", null, Y(e.title), 1)]), zr("p", Wln, Y(e.content), 1)])]))), 128))]))]), zr("section", yln, [zr("div", bln, [n[33] || (n[33] = zr("div", null, [zr("h3", null, "自定义审稿"), zr("p", {
        class: "panel-hint"
      }, "勾选后会在审稿时追加到维度列表之后，可自由编写补充要求。")], -1)), Qr(Tt(WS), {
        size: "small",
        quaternary: "",
        onClick: k
      }, {
        default: Sn(() => [...(n[32] || (n[32] = [jr("+ 新增", -1)]))]),
        _: 1
      })]), r.value ? Fr("", !0) : (Cr(), Ir("div", kln, [m.value.length ? Fr("", !0) : (Cr(), Ir("div", xln, "暂无自定义项，可点击「新增」添加。")), (Cr(!0), Ir(br, null, Va(m.value, e => (Cr(), Ir("div", {
        key: e.id,
        class: "prompt-card"
      }, [zr("div", wln, [Qr(Tt(aO), {
        checked: Boolean(e.enabled),
        "onUpdate:checked": t => x(e, "enabled", t)
      }, null, 8, ["checked", "onUpdate:checked"]), Qr(Tt(r_), {
        value: e.title,
        size: "small",
        placeholder: "提示词标题",
        class: "prompt-title-input",
        "onUpdate:value": t => x(e, "title", t)
      }, null, 8, ["value", "onUpdate:value"]), zr("button", {
        type: "button",
        class: "remove-btn",
        onClick: t => function (e) {
          c.value = c.value.filter(t => t.id !== e.id);
        }(e)
      }, "删除", 8, _ln)]), Qr(Tt(r_), {
        value: e.content,
        type: "textarea",
        autosize: {
          minRows: 2,
          maxRows: 6
        },
        placeholder: "审稿要求说明…",
        "onUpdate:value": t => x(e, "content", t)
      }, null, 8, ["value", "onUpdate:value"])]))), 128))]))]), zr("section", Sln, [zr("div", Cln, [n[35] || (n[35] = zr("div", null, [zr("h3", null, "自定义写作提示词"), zr("p", {
        class: "panel-hint"
      }, "针对不同写作模式的自定义要求。勾选后会注入到对应写作 Agent 的系统提示中。")], -1)), Qr(Tt(WS), {
        size: "small",
        quaternary: "",
        onClick: _
      }, {
        default: Sn(() => [...(n[34] || (n[34] = [jr("+ 新增", -1)]))]),
        _: 1
      })]), r.value ? Fr("", !0) : (Cr(), Ir("div", Pln, [u.value.length ? Fr("", !0) : (Cr(), Ir("div", Rln, "暂无自定义写作提示词。")), (Cr(!0), Ir(br, null, Va(u.value, e => (Cr(), Ir("div", {
        key: e.id,
        class: "prompt-card"
      }, [zr("div", Oln, [Qr(Tt(aO), {
        checked: Boolean(e.enabled),
        "onUpdate:checked": t => w(e, "enabled", t)
      }, null, 8, ["checked", "onUpdate:checked"]), Qr(Tt(r_), {
        value: e.title,
        size: "small",
        placeholder: "提示词标题",
        class: "prompt-title-input",
        "onUpdate:value": t => w(e, "title", t)
      }, null, 8, ["value", "onUpdate:value"]), Qr(Tt(PM), {
        value: e.category,
        options: S,
        size: "small",
        style: {
          width: "120px"
        },
        "onUpdate:value": t => w(e, "category", t)
      }, null, 8, ["value", "onUpdate:value"]), zr("button", {
        type: "button",
        class: "remove-btn",
        onClick: t => function (e) {
          u.value = u.value.filter(t => t.id !== e.id);
        }(e)
      }, "删除", 8, Mln)]), Qr(Tt(r_), {
        value: e.content,
        type: "textarea",
        autosize: {
          minRows: 3,
          maxRows: 10
        },
        placeholder: e.placeholder || "写作要求说明……",
        "onUpdate:value": t => w(e, "content", t)
      }, null, 8, ["value", "placeholder", "onUpdate:value"])]))), 128))]))])]),
      _: 1
    }), zr("footer", Iln, [n[37] || (n[37] = zr("span", {
      class: "footer-hint"
    }, "设置保存在本地数据目录，永久生效。", -1)), Qr(Tt(WS), {
      type: "primary",
      loading: l.value,
      disabled: r.value,
      onClick: b
    }, {
      default: Sn(() => [...(n[36] || (n[36] = [jr("保存设置", -1)]))]),
      _: 1
    }, 8, ["loading", "disabled"])])], 2));
  }
};
