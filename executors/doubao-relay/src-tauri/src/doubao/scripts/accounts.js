// 注入 doubao-chat 页面的账号操作脚本,暴露 window.__doubaoRelayAccounts。
// 约束见 docs/analysis/DECISIONS.md "Phase 0 验证结果":不依赖 aria-label / img[alt]。
// 脚本本身无状态;VERSION 变了就整体覆盖(豆包页面不 reload 也能升级),相同版本重复注入是 no-op。
(() => {
  const VERSION = '1.0.1';
  if (window.__doubaoRelayAccounts && window.__doubaoRelayAccounts.version === VERSION) return;
  const SUB_RE = /^[一-龥A-Za-z0-9]{1,24}套餐$/;
  const LOGIN_RE = /^(登录|立即登录|登录豆包)$/;
  const visible = el => { const r = el.getBoundingClientRect(); return r.width > 0 && r.height > 0; };
  const text = el => (el.innerText || el.textContent || '').trim();
  const lines = s => String(s || '').split(/\r?\n/).map(x => x.trim()).filter(Boolean);
  const sleep = ms => new Promise(r => setTimeout(r, ms));
  function parseLabel(raw) {
    const ls = lines(raw);
    if (!ls.length) return { name: '', subscription: '' };
    const sub = ls.slice(1).find(l => SUB_RE.test(l));
    if (sub) return { name: ls[0], subscription: sub };
    const m = ls[0].match(/^(.+?)\s+([一-龥A-Za-z0-9]{1,24}套餐)$/);
    return m ? { name: m[1].trim(), subscription: m[2] } : { name: ls[0], subscription: '' };
  }
  function allButtons() { return [...document.querySelectorAll('button,[role="button"]')].filter(visible); }
  // 账号按钮定位(macOS 2.32.5 实测:button[aria-haspopup=menu] 共 27–31 个,无 data-testid / aria-label):
  //   候选 = aria-haspopup=menu 且不在菜单内的可见按钮;
  //   1) 左下角位置(left 0–80 / top>innerHeight-120 / 宽 35–360 / 高 20–85)唯一命中即返回——不要求 <img>,
  //      因为默认头像的账号渲染的是占位 div 而不是 <img>(菜单行已实测如此,按钮极可能同一路径);
  //   2) 位置命中多个时依次用"含 <img>"、"文本能解析出套餐"收窄;
  //   3) 侧栏折叠时整栏平移到屏外(实测 rect (-299,823,231,48)),用 left -400–0 的同一高度带再找一轮;
  //   4) 最后退回全页唯一"含 <img>"/"含套餐文本"的候选。
  const inBand = r => r.top > innerHeight - 120 && r.bottom <= innerHeight + 2 && r.width >= 35 && r.width <= 360 && r.height >= 20 && r.height <= 85;
  const inCorner = el => { const r = el.getBoundingClientRect(); return inBand(r) && r.left >= 0 && r.left < 80; };
  const inHiddenCorner = el => { const r = el.getBoundingClientRect(); return inBand(r) && r.left >= -400 && r.left < 0; };
  const hasImg = el => !!el.querySelector('img');
  const hasSub = el => !!parseLabel(text(el)).subscription;
  const pick = list => (list.length === 1 ? list[0] : null);
  const narrow = list => pick(list) || pick(list.filter(hasImg)) || pick(list.filter(hasSub));
  const buttonCandidates = () => allButtons().filter(el => el.getAttribute('aria-haspopup') === 'menu' && !el.closest('[role="menu"]'));
  function findAccountButton() {
    const cands = buttonCandidates();
    return narrow(cands.filter(inCorner)) || narrow(cands.filter(inHiddenCorner)) || pick(cands.filter(hasImg)) || pick(cands.filter(hasSub))
      // 新版桌面端有时去掉 aria-haspopup,但仍把当前账号名放在可点击按钮文本中。
      // 只有旧式菜单候选完全不存在时才启用文字兜底；已有多个模糊候选时不能乱认账号。
      || pick(allButtons().filter(hasSub))
      || (cands.length === 0 ? pick(allButtons().filter(el => /用户\d{3,}/.test(text(el)))) : null);
  }
  function pointer(el, type, buttons) {
    const r = el.getBoundingClientRect();
    el.dispatchEvent(new PointerEvent(type, { bubbles: true, cancelable: true, pointerId: 1, pointerType: 'mouse', isPrimary: true, button: 0, buttons, clientX: r.x + r.width / 2, clientY: r.y + r.height / 2 }));
  }
  const loginVisible = () => allButtons().some(el => LOGIN_RE.test(text(el)));
  // 豆包关闭菜单时会保留子项直到退出动画结束;不能把这些已关闭菜单的子项继续算作可见。
  const MENU_ROOT = '[role="menu"][data-testid="chat_header_menu"]';
  const menuItems = () => [...document.querySelectorAll('[role="menuitem"]')].filter(el => {
    const root = el.closest(MENU_ROOT);
    return visible(el) && (!root || root.getAttribute('data-state') !== 'closed');
  });
  // macOS 2.32.5 实测:账号行带 data-testid="avatar_menu_account_row";无自定义头像的账号行里没有 <img>(只有占位 div),
  // 所以不能用 img 判定账号行。优先 testid,豆包改版去掉 testid 时退回"无子菜单 + 有 <img>"启发式。
  const ACCOUNT_ROW_TESTID = 'avatar_menu_account_row';
  const isAccountRow = el => el.getAttribute('data-testid') === ACCOUNT_ROW_TESTID;
  function accountRows() {
    const items = menuItems();
    const tagged = items.filter(isAccountRow);
    return tagged.length ? tagged : items.filter(el => !el.getAttribute('aria-haspopup') && !!el.querySelector('img'));
  }
  // 账号菜单是否仍处于打开状态:同时检查账号按钮与实际菜单。Windows 真机出现过菜单仍可见、
  // 但账号按钮 aria-expanded 已提前变成 false 的状态,此时仍必须继续关闭菜单。
  // 菜单根 [data-testid=chat_header_menu] 或"退出登录"项可见都算打开——
  // 这两者在任何账号数下都存在(只有 1 个账号时菜单里没有账号行,不能用账号行数兜底)。
  // Radix 会在关闭动画期间保留带尺寸的菜单节点;`data-state=closed` 时它已不再遮挡页面。
  // 页面里还常驻一个 0 尺寸的空 [role="menu"](radix 占位),不能用裸 [role=menu] 判断。
  const menuVisible = () => { const m = document.querySelector(MENU_ROOT); return (!!m && m.getAttribute('data-state') !== 'closed' && visible(m)) || menuItems().some(el => text(el) === '退出登录'); };
  const menuOpen = () => { const b = findAccountButton(); return (b && b.getAttribute('aria-expanded') === 'true') || menuVisible(); };
  const rectOf = el => { const r = el.getBoundingClientRect(); return [Math.round(r.left), Math.round(r.top), Math.round(r.width), Math.round(r.height)]; };
  window.__doubaoRelayAccounts = {
    version: VERSION,
    probe() { const b = findAccountButton(); const cands = buttonCandidates();
      return { version: VERSION, buttonFound: !!b, buttonText: b ? text(b) : '', expanded: b ? b.getAttribute('aria-expanded') : null, buttonHasImg: b ? hasImg(b) : null,
        loginVisible: loginVisible(), hasInput: !!document.querySelector('[contenteditable="true"],textarea,[role="textbox"]'), readyState: document.readyState, visibility: document.visibilityState, innerWidth, innerHeight,
        // 定位失败时的排查信息:aria-haspopup=menu 候选总数 + 落在底部高度带里的候选明细
        haspopupCount: cands.length,
        candidates: cands.filter(el => inCorner(el) || inHiddenCorner(el)).slice(0, 10).map(el => ({ text: text(el).slice(0, 60), hasImg: hasImg(el), hasSub: hasSub(el), rect: rectOf(el) })) }; },
    // macOS 2.32.5 实测:折叠态的"显示侧栏"按钮 data-testid="siderbar_closed_status_btn",位于 (93,15) 26x26,
    // 计划里 left<50 的启发式匹配不到,这里优先用 testid,启发式放宽到 left<120 兜底。不依赖 aria-label(两平台值不一致)。
    expandSidebar() { const tagged = allButtons().filter(el => el.getAttribute('data-testid') === 'siderbar_closed_status_btn');
      const m = tagged.length ? tagged : allButtons().filter(el => { const r = el.getBoundingClientRect();
        return r.width >= 15 && r.width <= 60 && r.height >= 15 && r.height <= 60 && r.left >= 0 && r.left < 120 && r.top >= 0 && r.top < 80 && !!el.querySelector('svg'); });
      if (m.length !== 1) return { action: 'none', count: m.length }; m[0].click(); return { action: 'temporary_expand', via: tagged.length ? 'testid' : 'heuristic' }; },
    // 新版豆包只接受可信鼠标事件打开账号菜单。这里仅返回坐标，由 Rust 侧通过
    // CDP Input.dispatchMouseEvent 点击；已打开时不再反向点击，避免菜单闪回。
    openMenu() { const b = findAccountButton(); if (!b) return { action: 'button_not_found' };
      if (menuOpen()) return { action: 'already_open' };
      const r = b.getBoundingClientRect();
      return { action: 'needs_native_click', x: r.x + r.width / 2, y: r.y + r.height / 2 }; },
    hoverMore() { const more = menuItems().filter(el => text(el) === '更多账号'); if (more.length !== 1) return { action: 'none', count: more.length };
      more[0].dispatchEvent(new PointerEvent('pointermove', { bubbles: true, pointerType: 'mouse', pointerId: 1 }));
      more[0].dispatchEvent(new MouseEvent('mouseover', { bubbles: true })); return { action: 'hover', expanded: more[0].getAttribute('aria-expanded') }; },
    readEntries() { const items = menuItems(); const rows = accountRows(); return { menuItemCount: items.length, logoutVisible: items.some(el => text(el) === '退出登录'),
      entries: items.slice(0, 100).map(el => ({ label: text(el).slice(0, 300), lines: lines(text(el)).slice(0, 6), hasAvatar: !!el.querySelector('img'), hasPopup: !!el.getAttribute('aria-haspopup'), disabled: el.getAttribute('aria-disabled') === 'true',
        isAccountRow: rows.includes(el), testId: el.getAttribute('data-testid') || '' })) }; },
    current() { const b = findAccountButton(); const login = loginVisible();
      if (!b) return { status: login ? 'not_logged_in' : 'unknown', name: '', subscription: '', raw: '' };
      const p = parseLabel(text(b)); return { status: p.name ? 'logged_in' : 'unknown', ...p, raw: text(b) }; },
    // 异步:Escape 由 radix 在 document 上监听,aria-expanded 要等 React 提交后才翻转,
    // 同步读会误判"仍打开"并触发 pointer 兜底(再点一次按钮 = 重新打开)。调用方需 awaitPromise。
    async closeMenu() { if (!menuOpen()) return { action: 'none', stillOpen: false };
      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', code: 'Escape', keyCode: 27, bubbles: true }));
      await sleep(150);
      if (!menuOpen()) return { action: 'escape', stillOpen: false };
      const b = findAccountButton();
      // Windows 上偶发菜单仍可见，但 React 已先把 aria-expanded 改回 false。
      // 只看 aria 会跳过最后一次点击，随后生成按钮被悬浮菜单遮挡。这里不能先
      // 合成 pointer 事件：新版豆包会先关闭菜单，再被末尾 click 重新打开。
      if (b && menuOpen()) {
        const r = b.getBoundingClientRect();
        b.click();
        await sleep(200); return {
          action: 'escape_then_click',
          stillOpen: menuOpen(),
          x: r.x + r.width / 2,
          y: r.y + r.height / 2,
        };
      }
      // 无按钮时只能看菜单根/退出登录项,菜单退出动画约 200ms,再等一轮
      await sleep(150); return { action: 'escape', stillOpen: menuOpen() }; },
    // count===0 → not_found(账号不在列表里),count>1 → ambiguous(同名账号);Task 1.7 按 action 映射错误码。
    switchTo(name) { const items = accountRows().filter(el => parseLabel(text(el)).name === name);
      if (items.length === 0) return { action: 'not_found', count: 0 };
      if (items.length > 1) return { action: 'ambiguous', count: items.length, labels: items.map(el => text(el)) };
      if (items[0].getAttribute('aria-disabled') === 'true') return { action: 'disabled' };
      items[0].click(); return { action: 'clicked', name }; },
  };
})();
