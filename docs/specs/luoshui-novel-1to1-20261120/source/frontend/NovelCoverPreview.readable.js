// Inspection only. Original SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Original UTF16 offsets [8573753, 8575641); original line 41
// External dependencies and runtime equivalence remain unresolved.
const recoveredComponent = {
  __name: "NovelCoverPreview",
  props: {
    imagePath: {
      type: String,
      default: ""
    },
    alt: {
      type: String,
      default: "封面"
    },
    compact: Boolean
  },
  setup(e) {
    const t = e,
      n = Nt(""),
      a = Nt(!1);
    return Ro(() => t.imagePath, async function () {
      const e = String(t.imagePath || "").trim();
      if (e) {
        a.value = !0;
        try {
          n.value = await async function (e, {
            forceRefresh: t = !1
          } = {}) {
            var n;
            const a = 171,
              o = "oooZ",
              r = "NJPA",
              l = 126,
              i = "jVms",
              s = "KtUa",
              c = 184,
              u = 137,
              d = "KtUa",
              v = 334,
              p = 115,
              f = "YvwA",
              m = 69,
              h = "Rj#X",
              g = 282,
              W = 336,
              y = "nlhS",
              b = 276,
              k = 12,
              x = 131,
              w = "^a*6",
              _ = "HfLa",
              S = "rbz$",
              C = "XAaB",
              P = 56,
              R = "y$H%",
              O = 47,
              M = 83,
              I = 78,
              N = "S%5V",
              G = 103,
              q = {
                tkupA: function (e, t) {
                  return e(t);
                },
                meKuG: function (e, t) {
                  return e || t;
                },
                kmLab: function (e, t) {
                  return e !== t;
                },
                BCpLR: "mcOSx",
                szmTz: z(79, 38, 116, "NJPA")
              },
              A = q[T(52, 13, 1, "cjBG")](String, q.meKuG(e, ""))[z(287, 151, a, o)]();
            function T(e, t, n, a) {
              return ltn(e - -275, a);
            }
            if (!A) return "";
            function z(e, t, n, a) {
              return ltn(t - -426, a);
            }
            if (!t && stn.has(A)) return stn[T(197, 0, 0, r)](A);
            const Q = null == (n = window[T(l, 0, 0, i) + z(0, 134, 0, s)]) ? void 0 : n[z(0, c, 0, "Q(Xk")];
            if (!(null == Q ? void 0 : Q[z(0, u, 0, d) + T(v, 0, 0, "XAaB") + T(p, 0, 0, f)])) return "";
            try {
              if (!q[T(m, 0, 0, h)](q[T(g, 0, 0, "7n$e")], q.BCpLR)) {
                const e = await Q[T(b, 0, 0, "*Wc6") + z(0, -k, 0, "Q(Xk") + T(x, 0, 0, w)](A);
                if (!(null == e ? void 0 : e[T(177, 0, 0, _) + "ss"]) || !e[T(225, 0, 0, S)]) return "";
                const t = q[z(0, 177, 0, C)](String, e[T(70, 0, 0, "YvwA")]),
                  n = t[T(P, 0, 0, R) + z(0, 190, 0, "!$gM")](q[T(O, 0, 0, "hsO)")]) ? t : T(M, 0, 0, "p4XL") + T(I, 0, 0, N) + "/png;base64," + t;
                return stn[z(0, -G, 0, "7)GK")](A, n), n;
              }
              if (!_0x152c02) return void _0x20eaad[T(W, 0, 0, y)]();
              _0x58ff52[z(0, a, 0, "prOI") + "e"](_0x12fde2);
            } catch {
              return "";
            }
          }(e);
        } finally {
          a.value = !1;
        }
      } else n.value = "";
    }, {
      immediate: !0
    }), Ga(() => {
      n.value = "";
    }), (t, o) => e.imagePath ? (Cr(), Ir("div", {
      key: 0,
      class: K(["cover-preview", {
        compact: e.compact
      }])
    }, [a.value && !n.value ? (Cr(), Ir("div", dtn, [Qr(Tt(VE), {
      size: 18
    })])) : n.value ? (Cr(), Ir("img", {
      key: 1,
      src: n.value,
      alt: e.alt,
      class: "cover-image"
    }, null, 8, vtn)) : (Cr(), Ir("div", ptn, "图片加载失败"))], 2)) : Fr("", !0);
  }
};
