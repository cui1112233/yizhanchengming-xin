// Inspection only; target source is never executed.
// Original frontend SHA256: 7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6
// Native entry 0x180076750; constructor source/ida/0x180091e60.c:1479

// Frontend candidate 1; original UTF16 [8667267, 8667746]; property UTF8 [8925782, 8925839]; component None
async function qan(e, t) {
  var n, a;
  const o = 1101,
    r = 952,
    l = 1263,
    i = "k*oF",
    s = 761,
    c = "W^o1",
    u = 1195,
    d = "q0!*",
    v = 1032,
    p = "S*tx",
    f = 1095,
    m = 592,
    h = "Weig",
    g = 525,
    W = {};
  function y(e, t, n, a) {
    return Gan(0, e, t - -42);
  }
  function b(e, t, n, a) {
    return Gan(0, e, n - g);
  }
  W["sessi" + y("oP9p", 187)] = e, W[b("k*oF", o, r) + b("c3iw", l, 1088)] = t;
  const k = {
    success: !1
  };
  return (null == (a = null == (n = {
    jSoLr: function (e) {
      return e();
    }
  }[b(i, 0, s)](Ran)) ? void 0 : n[b(c, 0, u) + b(d, 0, v) + "ate_s" + b(p, 0, f) + y("4M#3", m) + b(h, 0, 959)]) ? void 0 : a.call(n, W)) ?? k;
}

// Caller 1; NovelStudioPage/et; UTF16 [9268629, 9268643]
// Exact argument expressions: ["d.value", "n"]
async function et(e) {
  var t;
  if (!d.value || !(null == e ? void 0 : e.length)) return;
  const n = [...((null == (t = v.value) ? void 0 : t.messages) || []), ...e];
  v.value = {
    ...v.value,
    messages: n
  };
  const a = await qan(d.value, n);
  (null == a ? void 0 : a.success) && (v.value = a.data);
}

// Caller 2; NovelStudioPage/None; UTF16 [9269239, 9269247]
// Exact argument expressions: ["e", "o"]
async function (e, t) {
  var n;
  if (!e || !(null == t ? void 0 : t.length)) return;
  if (e === d.value) return void (await et(t));
  const a = await Nan(e);
  if (!(null == a ? void 0 : a.success)) return;
  const o = [...((null == (n = a.data) ? void 0 : n.messages) || []), ...t];
  await qan(e, o), await ke();
}

// Caller 3; NovelStudioPage/wt; UTF16 [9280841, 9280855]
// Exact argument expressions: ["d.value", "o"]
async function wt({
  index: e
}) {
  var t, a;
  if (!d.value || e < 0) return;
  const o = [...((null == (t = v.value) ? void 0 : t.messages) || [])];
  if (!o[e]) return;
  let r = 1;
  "user" === o[e].role && "assistant" === (null == (a = o[e + 1]) ? void 0 : a.role) && (r = 2), o.splice(e, r);
  const l = await qan(d.value, o);
  (null == l ? void 0 : l.success) ? v.value = l.data : n.error((null == l ? void 0 : l.message) || "删除失败");
}
