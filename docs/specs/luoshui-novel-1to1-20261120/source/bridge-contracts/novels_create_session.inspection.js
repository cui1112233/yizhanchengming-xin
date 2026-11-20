// Inspection only; target source is never executed.
// Original frontend SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Native entry 0x180073da0; constructor source/ida/0x180091e60.c:1399

// Frontend candidate 1; original UTF16 [9257841, 9258189]; property UTF8 [9523309, 9523369]; component NovelStudioPage
async function (e = {}) {
  var t, n;
  const a = "XGTO",
    o = 778,
    r = "#WGY",
    l = "iUeH";
  function i(e, t, n, a) {
    return Gan(0, a, e - 111);
  }
  const s = {};
  function c(e, t, n, a) {
    return Gan(0, e, t - -814);
  }
  return s.success = !1, (null == (n = null == (t = {
    FGkdI: function (e) {
      return e();
    }
  }[c(a, -128)](Ran)) ? void 0 : t[i(653, 0, 0, "XGTO") + i(o, 0, 0, r) + i(649, 0, 0, "Z8Hk") + c(l, -587) + "n"]) ? void 0 : n.call(t, e)) ?? s;
}

// Caller 1; NovelStudioPage/Pe; UTF16 [9257841, 9258242]
// Exact argument expressions: ["{\n  kind: o,\n  bookId: t,\n  workflow: r,\n  title: a || K9t[r] || \"新会话\"\n}"]
async function Pe(e = "book-create", t = null, a = "") {
  let o = e,
    r = e;
  t ? (o = "book", r = "book") : ["book-create", "short", "imitation-short", "fanfic", "spinoff", "imitation", "chapters"].includes(e) ? (o = function (e) {
    const t = 1531,
      n = "OrOs",
      a = 1020,
      o = "Dqz^",
      r = 1047,
      l = 1511,
      i = "v(BQ",
      s = 1312,
      c = "knG)",
      u = "Dqz^",
      d = 1530,
      v = 1474,
      p = "Qf1U",
      f = 1395,
      m = 1526,
      h = 1553,
      g = "7l1T",
      W = 1391,
      y = 1684,
      b = 1359,
      k = "knG)",
      x = 1590,
      w = 1639,
      _ = "a7rx",
      S = 1508,
      C = 772,
      P = "3Fep",
      R = "&Jot",
      O = 766,
      M = 1442,
      I = "vqft",
      N = 1357,
      G = "GqS4",
      q = 689,
      A = 1174,
      T = 1342,
      z = 1634,
      Q = 1498,
      D = 742,
      L = 936,
      j = {};
    function E(e, t, n, a) {
      return len(0, n, 0, a - L);
    }
    function F(e, t, n, a) {
      return U9t(0, 0, n - 76, t);
    }
    j[E(1445, 0, "$f@l", 1495)] = function (e, t) {
      return e === t;
    }, j.uLcMQ = E(t, 0, n, 1401), j[F(a, o, r)] = function (e, t) {
      return e === t;
    }, j[E(l, 0, i, 1424)] = E(s, 0, c, 1423) + E(1544, 0, u, d) + E(v, 0, p, f), j[E(m, 0, n, 1342)] = function (e, t) {
      return e === t;
    }, j[E(h, 0, g, W)] = E(y, 0, "C$d9", 1543), j.OqKXn = E(b, 0, k, x) + E(w, 0, _, S) + "e";
    const B = j;
    return B[F(0, "y$*N", C)](e, B[F(0, P, 737)]) || B[F(0, R, O)](e, B[E(M, 0, I, N)]) ? B[F(0, G, q)] : B[E(A, 0, n, T)](e, B[F(0, R, 702)]) ? E(z, 0, "v(BQ", Q) : B[F(0, G, D)];
  }(e), r = e) : (r = m.value || "book-create", o = e);
  const l = await async function (e = {}) {
    var t, n;
    const a = "XGTO",
      o = 778,
      r = "#WGY",
      l = "iUeH";
    function i(e, t, n, a) {
      return Gan(0, a, e - 111);
    }
    const s = {};
    function c(e, t, n, a) {
      return Gan(0, e, t - -814);
    }
    return s.success = !1, (null == (n = null == (t = {
      FGkdI: function (e) {
        return e();
      }
    }[c(a, -128)](Ran)) ? void 0 : t[i(653, 0, 0, "XGTO") + i(o, 0, 0, r) + i(649, 0, 0, "Z8Hk") + c(l, -587) + "n"]) ? void 0 : n.call(t, e)) ?? s;
  }({
    kind: o,
    bookId: t,
    workflow: r,
    title: a || K9t[r] || "新会话"
  });
  (null == l ? void 0 : l.success) ? (t || (m.value = r), await ke(), t && (await xe(t)), await Ce(l.data.id)) : n.error((null == l ? void 0 : l.message) || "创建会话失败");
}
