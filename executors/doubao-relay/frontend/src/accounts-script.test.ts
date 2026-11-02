// @vitest-environment happy-dom
// src-tauri/src/doubao/scripts/accounts.js 的 DOM 行为测试:用 happy-dom 搭最小 DOM,
// 手工 stub getBoundingClientRect(happy-dom 一律返回 0),覆盖账号按钮定位 / 菜单判定 / switchTo 分支。
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import accountsSource from '../../src-tauri/src/doubao/scripts/accounts.js?raw';

type Rect = { left: number; top: number; width: number; height: number };
type AccountsApi = {
  version: string;
  probe(): { buttonFound: boolean; buttonText: string; buttonHasImg: boolean | null; haspopupCount: number; candidates: unknown[] };
  current(): { status: string; name: string; subscription: string; raw: string };
  expandSidebar(): { action: string; count?: number; via?: string };
  closeMenu(): Promise<{ action: string; stillOpen: boolean }>;
  switchTo(name: string): { action: string; count?: number; name?: string; labels?: string[] };
};

const INNER_HEIGHT = 879;
const INNER_WIDTH = 1512;

const stubRect = (el: Element, r: Rect) => {
  const full = { ...r, x: r.left, y: r.top, right: r.left + r.width, bottom: r.top + r.height, toJSON: () => full };
  Object.defineProperty(el, 'getBoundingClientRect', { value: () => full, configurable: true });
};

// 真豆包的"名字 / 套餐"是两个块级 div(innerText 在块之间产生换行;文本节点里的 \n 会被折叠)
const appendLines = (parent: Element, text: string) => {
  for (const line of text.split('\n')) {
    const d = document.createElement('div');
    d.textContent = line;
    parent.appendChild(d);
  }
};

const mkButton = (o: { text: string; rect: Rect; haspopup?: boolean; img?: boolean; attrs?: Record<string, string> }) => {
  const b = document.createElement('button');
  if (o.haspopup !== false) b.setAttribute('aria-haspopup', 'menu');
  b.setAttribute('aria-expanded', 'false');
  if (o.img) b.appendChild(document.createElement('img'));
  appendLines(b, o.text);
  for (const [k, v] of Object.entries(o.attrs ?? {})) b.setAttribute(k, v);
  stubRect(b, o.rect);
  document.body.appendChild(b);
  return b;
};

const mkMenu = (items: { text: string; testid?: string; img?: boolean }[]) => {
  const menu = document.createElement('div');
  menu.setAttribute('role', 'menu');
  menu.setAttribute('data-testid', 'chat_header_menu');
  stubRect(menu, { left: 10, top: 500, width: 240, height: 300 });
  for (const it of items) {
    const el = document.createElement('div');
    el.setAttribute('role', 'menuitem');
    if (it.testid) el.setAttribute('data-testid', it.testid);
    if (it.img) el.appendChild(document.createElement('img'));
    appendLines(el, it.text);
    stubRect(el, { left: 10, top: 500, width: 240, height: 40 });
    menu.appendChild(el);
  }
  document.body.appendChild(menu);
  return menu;
};

// 真机 2.32.5 的账号按钮 rect:(9,823,231,48),innerHeight=879
const CORNER: Rect = { left: 9, top: 823, width: 231, height: 48 };
// 侧栏折叠时整栏平移到屏外:(-299,823,231,48)
const HIDDEN_CORNER: Rect = { left: -299, top: 823, width: 231, height: 48 };

const inject = (): AccountsApi => {
  (0, eval)(accountsSource);
  return (window as unknown as { __doubaoRelayAccounts: AccountsApi }).__doubaoRelayAccounts;
};

// 对话列表每行的"更多"按钮等:aria-haspopup=menu、无 img、不在左下角
const addNoise = () => {
  for (let i = 0; i < 5; i++) mkButton({ text: '', rect: { left: 200, top: 100 + i * 40, width: 24, height: 24 } });
};

beforeEach(() => {
  document.body.innerHTML = '';
  delete (window as unknown as { __doubaoRelayAccounts?: unknown }).__doubaoRelayAccounts;
  (globalThis as unknown as { innerHeight: number }).innerHeight = INNER_HEIGHT;
  (globalThis as unknown as { innerWidth: number }).innerWidth = INNER_WIDTH;
  (window as unknown as { innerHeight: number }).innerHeight = INNER_HEIGHT;
  (window as unknown as { innerWidth: number }).innerWidth = INNER_WIDTH;
});

afterEach(() => {
  document.body.innerHTML = '';
});

describe('findAccountButton / current()', () => {
  it('有头像账号:左下角 + <img> 唯一命中', () => {
    addNoise();
    mkButton({ text: '她是翻哥\n体验套餐', rect: CORNER, img: true });
    const api = inject();
    expect(api.probe().buttonFound).toBe(true);
    expect(api.probe().buttonHasImg).toBe(true);
    expect(api.current()).toMatchObject({ status: 'logged_in', name: '她是翻哥', subscription: '体验套餐' });
  });

  it('默认头像账号(按钮里没有 <img>,只有占位 div):位置唯一命中即返回', () => {
    addNoise();
    const b = mkButton({ text: '用户225205\n体验套餐', rect: CORNER });
    b.prepend(document.createElement('div'));
    const api = inject();
    const p = api.probe();
    expect(p.buttonFound).toBe(true);
    expect(p.buttonHasImg).toBe(false);
    expect(api.current()).toMatchObject({ status: 'logged_in', name: '用户225205', subscription: '体验套餐' });
  });

  it('位置带内多个候选:先按 <img> 收窄', () => {
    addNoise();
    mkButton({ text: '设置', rect: { left: 9, top: 770, width: 231, height: 48 } });
    mkButton({ text: '她是翻哥\n体验套餐', rect: CORNER, img: true });
    const api = inject();
    expect(api.current().name).toBe('她是翻哥');
  });

  it('位置带内多个候选且都无 <img>:按"能解析出套餐"收窄', () => {
    addNoise();
    mkButton({ text: '设置', rect: { left: 9, top: 770, width: 231, height: 48 } });
    mkButton({ text: '用户225205\n体验套餐', rect: CORNER });
    const api = inject();
    expect(api.current()).toMatchObject({ status: 'logged_in', name: '用户225205' });
  });

  it('位置带内多个候选、无 <img>、都无套餐文本:返回 unknown 而不是乱选', () => {
    addNoise();
    mkButton({ text: '设置', rect: { left: 9, top: 770, width: 231, height: 48 } });
    mkButton({ text: '用户225205', rect: CORNER });
    const api = inject();
    expect(api.probe().buttonFound).toBe(false);
    expect(api.current().status).toBe('unknown');
    expect(api.probe().candidates).toHaveLength(2);
  });

  it('侧栏折叠(按钮在屏外 left<0)+ 无 <img>:仍能定位', () => {
    addNoise();
    mkButton({ text: '用户225205\n体验套餐', rect: HIDDEN_CORNER });
    const api = inject();
    expect(api.current()).toMatchObject({ status: 'logged_in', name: '用户225205' });
  });

  it('菜单内的 menuitem 按钮不算候选;未登录时 status=not_logged_in', () => {
    addNoise();
    mkButton({ text: '登录', rect: { left: 9, top: 823, width: 100, height: 40 }, haspopup: false });
    const api = inject();
    expect(api.current().status).toBe('not_logged_in');
  });

  it('新版单账号按钮没有 aria-haspopup 时按套餐文本定位', () => {
    addNoise();
    mkButton({ text: '体验者\n体验套餐', rect: { left: 220, top: 720, width: 230, height: 48 }, haspopup: false });
    const api = inject();
    expect(api.current()).toMatchObject({ status: 'logged_in', name: '体验者', subscription: '体验套餐' });
  });
});

describe('版本守卫', () => {
  it('同版本重复注入是 no-op,旧版本对象会被覆盖', () => {
    const api = inject();
    expect(api.version).toBe('1.0.1');
    expect(inject()).toBe(api);
    (window as unknown as { __doubaoRelayAccounts: unknown }).__doubaoRelayAccounts = { version: '0.0.0' };
    const fresh = inject();
    expect(fresh).not.toBe(api);
    expect(fresh.version).toBe('1.0.1');
    expect(typeof fresh.current).toBe('function');
  });
});

describe('closeMenu() 在找不到账号按钮时的兜底', () => {
  it('账号按钮 aria-expanded 已变 false 但菜单仍可见时继续关闭菜单', async () => {
    mkButton({
      text: '用户945344\n体验套餐',
      rect: { left: 9, top: 699, width: 230, height: 48 },
      attrs: { 'aria-expanded': 'false' },
    });
    const menu = mkMenu([{ text: '设置' }, { text: '退出登录', testid: 'setup_logout' }]);
    let escapes = 0;
    document.addEventListener('keydown', (e) => {
      if ((e as KeyboardEvent).key === 'Escape') {
        escapes++;
        menu.remove();
      }
    });
    const api = inject();
    const r = await api.closeMenu();
    expect(escapes).toBe(1);
    expect(r).toEqual({ action: 'escape', stillOpen: false });
  });

  it('单账号菜单(无账号行,只有退出登录)也判定为打开并派发 Escape', async () => {
    const menu = mkMenu([{ text: '设置' }, { text: '退出登录', testid: 'setup_logout' }]);
    let escapes = 0;
    document.addEventListener('keydown', (e) => {
      if ((e as KeyboardEvent).key === 'Escape') {
        escapes++;
        menu.remove();
      }
    });
    const api = inject();
    const r = await api.closeMenu();
    expect(escapes).toBe(1);
    expect(r).toEqual({ action: 'escape', stillOpen: false });
  });

  it('没有菜单时返回 none,不派发 Escape', async () => {
    let escapes = 0;
    document.addEventListener('keydown', () => escapes++);
    const api = inject();
    expect(await api.closeMenu()).toEqual({ action: 'none', stillOpen: false });
    expect(escapes).toBe(0);
  });
});

describe('switchTo()', () => {
  it('0 匹配 → not_found;≥2 匹配 → ambiguous;唯一匹配 → clicked', () => {
    mkMenu([
      { text: '她是翻哥\n体验套餐', testid: 'avatar_menu_account_row', img: true },
      { text: '用户225205', testid: 'avatar_menu_account_row' },
      { text: '用户225205', testid: 'avatar_menu_account_row' },
      { text: '退出登录', testid: 'setup_logout' },
    ]);
    const api = inject();
    expect(api.switchTo('__nope__')).toEqual({ action: 'not_found', count: 0 });
    expect(api.switchTo('用户225205')).toEqual({ action: 'ambiguous', count: 2, labels: ['用户225205', '用户225205'] });
    expect(api.switchTo('她是翻哥')).toEqual({ action: 'clicked', name: '她是翻哥' });
  });
});

describe('expandSidebar()', () => {
  it('只认 data-testid 与位置启发式,不依赖 aria-label', () => {
    // 只有 aria-label、位置不符、无 svg:不应命中
    mkButton({ text: '', rect: { left: 500, top: 300, width: 26, height: 26 }, haspopup: false, attrs: { 'aria-label': '显示侧栏' } });
    const api = inject();
    expect(api.expandSidebar()).toEqual({ action: 'none', count: 0 });
    const tagged = mkButton({ text: '', rect: { left: 93, top: 15, width: 26, height: 26 }, haspopup: false, attrs: { 'data-testid': 'siderbar_closed_status_btn' } });
    let clicked = 0;
    tagged.addEventListener('click', () => clicked++);
    expect(api.expandSidebar()).toEqual({ action: 'temporary_expand', via: 'testid' });
    expect(clicked).toBe(1);
  });
});
