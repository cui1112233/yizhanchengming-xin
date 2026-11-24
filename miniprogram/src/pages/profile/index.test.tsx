import { beforeEach, expect, test, vi } from "vitest";
import { readFileSync } from "node:fs";

const h = vi.hoisted(() => ({
  slots: [] as any[],
  cursor: 0,
  effects: [] as Array<() => void | (() => void)>,
  shows: [] as Array<() => void>,
  read: vi.fn(),
  loadTasks: vi.fn(),
  getStorageSync: vi.fn(),
  navigate: vi.fn(),
  switchTab: vi.fn(),
  showToast: vi.fn(),
  reLaunch: vi.fn(),
  logout: vi.fn(),
  glassScroll: vi.fn(),
  glassLayout: vi.fn(),
  themeSet: vi.fn(),
  revision: 0,
  hasSession: false,
}));

vi.mock("react", () => ({
  useState: (initial: any) => {
    const index = h.cursor++;
    if (!(index in h.slots)) h.slots[index] = typeof initial === "function" ? initial() : initial;
    return [h.slots[index], (next: any) => { h.slots[index] = typeof next === "function" ? next(h.slots[index]) : next; }];
  },
  useRef: (initial: any) => {
    const index = h.cursor++;
    if (!(index in h.slots)) h.slots[index] = { current: initial };
    return h.slots[index];
  },
  useEffect: (effect: any) => { h.effects.push(effect); },
}));

vi.mock("@tarojs/taro", () => ({
  default: {
    getStorageSync: h.getStorageSync,
    setStorageSync: vi.fn(),
    navigateTo: h.navigate,
    reLaunch: h.reLaunch,
    switchTab: h.switchTab,
    showToast: h.showToast,
    showTabBar: vi.fn(),
    eventCenter: { trigger: vi.fn() },
  },
  useDidShow: (callback: () => void) => { h.shows.push(callback); },
  useDidHide: vi.fn(),
}));
vi.mock("@tarojs/components", () => ({ Button: "Button", Image: "Image", ScrollView: "ScrollView", Switch: "Switch", Text: "Text", View: "View" }));
vi.mock("../../components/BrandHeader", () => ({ default: "Header" }));
vi.mock("../../components/AmbientStarfield", () => ({ default: "Stars" }));
vi.mock("../../components/AppIcon", () => ({ default: "Icon" }));
vi.mock("../../components/Companion", () => ({ default: "Companion" }));
vi.mock("../../theme", () => ({ ThemedPage: "Page", useMobileTheme: () => ({ preference: 'dark', setTheme: h.themeSet }) }));
vi.mock("../../ui/TextureRefraction", () => ({ PageTextureRefraction: "Refraction" }));
vi.mock("../../ui/glassScrollBridge", () => ({
  publishGlassScroll: h.glassScroll,
  publishGlassLayout: h.glassLayout,
}));
vi.mock("../../services/memberProfile", () => ({ readMemberProfile: h.read }));
vi.mock("../../services/taskSummary", async original => ({
  ...await original<any>(),
  loadTaskSummary: h.loadTasks,
}));
vi.mock("../../services/session", () => ({
  hasAccountSession: () => h.hasSession,
  logoutAccount: h.logout,
}));
vi.mock("../../services/sessionBoundary", () => ({ sessionRevision: () => h.revision }));
vi.mock("../../config/localAuth", () => ({ isLocalAccountLogin: () => false }));

import ProfilePage from "./index";

function render() {
  h.cursor = 0;
  h.shows = [];
  return ProfilePage();
}

async function flush() {
  for (let index = 0; index < 12; index += 1) await Promise.resolve();
}

function findByClass(node: any, className: string): any {
  if (Array.isArray(node)) return node.map(child => findByClass(child, className)).find(Boolean);
  if (!node || typeof node !== "object") return undefined;
  if (node.props?.className?.split(" ").includes(className)) return node;
  return findByClass(node.props?.children, className);
}

function findAllByClass(node: any, className: string): any[] {
  if (Array.isArray(node)) return node.flatMap(child => findAllByClass(child, className));
  if (!node || typeof node !== "object") return [];
  const own = node.props?.className?.split(" ").includes(className) ? [node] : [];
  return [...own, ...findAllByClass(node.props?.children, className)];
}

function textContent(node: any): string {
  if (Array.isArray(node)) return node.map(textContent).join("");
  if (node === null || node === undefined || typeof node === "boolean") return "";
  if (typeof node !== "object") return String(node);
  return textContent(node.props?.children);
}

function click(className: string) {
  const node = findByClass(render(), className);
  expect(node, `missing control ${className}`).toBeDefined();
  node.props.onClick();
}

function deferred() {
  let resolve!: (value: any) => void;
  let reject!: (cause: any) => void;
  const promise = new Promise<any>((yes, no) => { resolve = yes; reject = no; });
  return { promise, resolve, reject };
}

function mount() {
  render();
  h.effects.forEach(effect => effect());
}

test("appearance settings stay in a dismissible sheet and keep real theme switching", async () => {
  h.hasSession = true;
  h.read.mockResolvedValue({ member: { username: "creator", displayName: "Creator", role: "member" } });
  mount(); await flush();
  expect(findByClass(render(), "studio-profile-sheet")).toBeUndefined();
  click("studio-appearance-entry");
  const sheet = findByClass(render(), "studio-profile-sheet");
  expect(sheet).toBeDefined();
  const option = findByClass(sheet, "studio-theme-option");
  option.props.onClick();
  expect(h.themeSet).toHaveBeenCalledWith("light");
  click("studio-profile-mask");
  expect(findByClass(render(), "studio-profile-sheet")).toBeUndefined();
});

beforeEach(() => {
  h.slots = [];
  h.effects = [];
  h.shows = [];
  h.hasSession = false;
  h.revision = 0;
  vi.resetAllMocks();
  h.navigate.mockResolvedValue({});
  h.reLaunch.mockResolvedValue({});
  h.logout.mockResolvedValue({ revoked: true });
  h.loadTasks.mockResolvedValue({ entries: [] });
  h.getStorageSync.mockReturnValue(undefined);
});

test("identity atmosphere preserves the real avatar and renders seven reducible petals", async () => {
  h.hasSession = true;
  h.getStorageSync.mockReturnValue(true);
  h.read.mockResolvedValue({
    member: { username: "creator", displayName: "真实用户", avatarUrl: "/media/avatar.webp" },
    usage: { month: { totalTokens: 0 } },
  });
  mount();
  await flush();

  const tree = render();
  const identity = findByClass(tree, "profile-space-identity");
  expect(identity.props.className.split(" ")).toContain("is-reduced-motion");
  expect(findByClass(identity, "profile-space-avatar-image").props.src).toContain("/media/avatar.webp");
  const background = findByClass(identity, "profile-space-identity-background");
  expect(background.props.src).toContain("profile-identity-sunset-v2");
  expect(background.props.src).not.toContain("avatar");
  const petals = findAllByClass(identity, "profile-space-petal");
  expect(petals).toHaveLength(7);
  expect(new Set(petals.map(node => node.props.src)).size).toBe(1);
});

test("approved profile composition uses real account, task, and quota data", async () => {
  h.hasSession = true;
  h.read.mockResolvedValue({
    member: { username: "real-creator", displayName: "真实创作者", role: "member", monthlyTokenLimit: null },
    team: { name: "真实团队" },
    usage: { month: { totalTokens: 1250 } },
  });
  h.loadTasks.mockResolvedValue({ entries: [
    { id: "run", status: "running" },
    { id: "done-a", status: "completed" },
    { id: "done-b", status: "completed" },
  ] });

  mount();
  await flush();
  const tree = render();
  const hero = findByClass(tree, "profile-space-identity");
  expect(textContent(hero)).toContain("真实创作者");
  expect(textContent(hero)).toContain("真实团队");
  expect(textContent(hero)).toContain("1进行中");
  expect(textContent(hero)).toContain("2已完成");
  expect(textContent(findByClass(tree, "profile-space-quota"))).toContain("1,250");
  const listText = textContent(findByClass(tree, "profile-space-list"));
  for (const label of ["我的资产", "创作偏好", "会员与用量", "账号与安全", "帮助与反馈"]) expect(listText).toContain(label);
  const rowIcons = findAllByClass(tree, "profile-space-row-icon").map(node => node.props.src);
  expect(rowIcons).toHaveLength(5);
  expect(new Set(rowIcons).size).toBe(5);
  expect(textContent(tree).match(/我的空间/g)).toHaveLength(1);
});

test("late task summary still updates after the faster account request finishes", async () => {
  h.hasSession = true;
  const tasks = deferred();
  h.read.mockResolvedValue({ member: { username: "creator" } });
  h.loadTasks.mockReturnValue(tasks.promise);
  mount();
  await flush();
  expect(textContent(findByClass(render(), "profile-space-metrics"))).toContain("—进行中");

  tasks.resolve({ entries: [{ id: "running", status: "running" }, { id: "done", status: "completed" }] });
  await flush();
  const metrics = textContent(findByClass(render(), "profile-space-metrics"));
  expect(metrics).toContain("1进行中");
  expect(metrics).toContain("1已完成");
});

test("profile rows keep their real actions instead of decorative navigation", async () => {
  h.hasSession = true;
  h.read.mockResolvedValue({ member: { username: "creator" }, usage: { month: { totalTokens: 0 } } });
  mount();
  await flush();

  click("profile-space-assets");
  expect(h.switchTab).toHaveBeenCalledWith({ url: "/pages/workspace/index" });
  click("profile-space-preference");
  expect(findByClass(render(), "studio-profile-sheet")).toBeDefined();
  click("profile-space-help");
  expect(h.showToast).toHaveBeenCalledWith({ title: "帮助与反馈暂未开放", icon: "none" });
});

test("forwards internal profile scrolling to the dynamic glass renderer", () => {
  const scroll = findByClass(render(), "mp-profile-scroll");
  expect(scroll.props.onScroll).toBeTypeOf("function");
  scroll.props.onScroll({ detail: { scrollTop: 86 } });
  expect(h.glassScroll).toHaveBeenCalledExactlyOnceWith("profile", 86);
});

test("missing public account session goes to login without starting a profile request", () => {
  render();
  h.shows[1]();
  expect(h.read).not.toHaveBeenCalled();
  expect(h.navigate).toHaveBeenCalledWith({ url: "/pages/bind/index" });
});

test("profile request failure always leaves the loading state", async () => {
  h.hasSession = true;
  h.read.mockRejectedValue(new Error("当前登录状态已失效"));
  render();
  h.shows[1]();
  await flush();
  const tree = render();
  expect(textContent(tree)).toContain("无法读取当前账号资料");
  expect(textContent(tree)).toContain("登录状态已失效");
  expect(textContent(tree)).not.toContain("正在读取账号资料");
});

test("page watchdog leaves loading even when the profile promise never settles", async () => {
  vi.useFakeTimers();
  try {
    h.hasSession = true;
    h.read.mockReturnValue(new Promise(() => {}));
    mount();
    expect(textContent(render())).toContain("正在读取账号资料");

    await vi.advanceTimersByTimeAsync(15000);

    const tree = render();
    expect(textContent(tree)).toContain("无法读取当前账号资料");
    expect(textContent(findByClass(tree, "mp-profile-state"))).toContain("读取超时");
    expect(textContent(tree)).not.toContain("正在读取账号资料");
  } finally {
    vi.useRealTimers();
  }
});

test("mount plus did-show shares one request and renders the account", async () => {
  h.hasSession = true;
  const request = deferred();
  h.read.mockReturnValue(request.promise);
  mount();
  h.shows[1]();
  expect(h.read).toHaveBeenCalledTimes(1);
  expect(textContent(findByClass(render(), "mp-profile-state"))).toContain("正在读取账号资料");
  request.resolve({ member: { username: "real-account" } });
  await flush();
  const tree = render();
  expect(findByClass(tree, "mp-profile-diagnostic")).toBeUndefined();
  expect(textContent(findByClass(tree, "mp-profile-hero"))).toContain("real-account");
});

test("cancel navigates immediately while logout is unresolved and ignores old response", async () => {
  h.hasSession = true;
  const request = deferred();
  h.read.mockReturnValue(request.promise);
  h.logout.mockReturnValue(new Promise(() => {}));
  mount();
  click("mp-profile-cancel");
  expect(h.logout).toHaveBeenCalledTimes(1);
  expect(h.reLaunch).toHaveBeenCalledWith({ url: "/pages/bind/index" });
  request.resolve({ member: { username: "old-account" } });
  await flush();
  expect(findByClass(render(), "mp-profile-hero")).toBeUndefined();
});

test("failure offers retry and login with fixed labels instead of sensitive raw errors", async () => {
  h.hasSession = true;
  h.read.mockRejectedValueOnce(new Error("network secret-token response body"));
  mount();
  await flush();
  const tree = render();
  expect(textContent(tree)).not.toContain("secret-token");
  expect(textContent(findByClass(tree, "mp-profile-state"))).toContain("网络连接失败，请稍后重试");
  expect(textContent(findByClass(tree, "mp-profile-cancel"))).toBe("返回登录");
  h.read.mockReturnValue(new Promise(() => {}));
  click("mp-profile-retry");
  expect(h.read).toHaveBeenCalledTimes(2);
});

test("a canceled request finally cannot unlock or update a newer attempt", async () => {
  h.hasSession = true;
  const old = deferred();
  const newer = deferred();
  h.read.mockReturnValueOnce(old.promise).mockReturnValueOnce(newer.promise);
  mount();
  click("mp-profile-cancel");
  h.shows[1]();
  expect(h.read).toHaveBeenCalledTimes(2);
  old.reject(new Error("old network error"));
  await flush();
  h.shows[1]();
  expect(h.read).toHaveBeenCalledTimes(2);
  expect(textContent(render())).toContain("正在读取账号资料");
  newer.resolve({ member: { username: "new-account" } });
  await flush();
  expect(textContent(findByClass(render(), "mp-profile-hero"))).toContain("new-account");
});

test.each(['resolve', 'reject'])("a session revision change clears loading and offers safe recovery after %s", async outcome => {
  h.hasSession = true;
  h.read.mockResolvedValueOnce({ member: { username: "previous-account" } });
  mount();
  await flush();
  const request = deferred();
  h.read.mockReturnValue(request.promise);
  h.shows[1]();
  h.revision += 1;
  if (outcome === 'resolve') request.resolve({ member: { username: "old-account" } });
  else request.reject(new Error('network secret-response'));
  await flush();
  const tree = render();
  expect(findByClass(tree, "mp-profile-hero")).toBeUndefined();
  expect(textContent(tree)).not.toContain("正在读取账号资料");
  expect(textContent(tree)).not.toContain('secret-response');
  expect(textContent(findByClass(tree, "mp-profile-state"))).toContain('登录状态已失效');
  expect(textContent(findByClass(tree, "mp-profile-retry"))).toBe('重新读取');
  expect(textContent(findByClass(tree, "mp-profile-cancel"))).toBe('返回登录');
  h.read.mockResolvedValueOnce({ member: { username: "current-account" } });
  click('mp-profile-retry');
  await flush();
  expect(textContent(findByClass(render(), 'mp-profile-hero'))).toContain('current-account');
});

test("unmount invalidates pending request without updating disposed state", async () => {
  h.hasSession = true;
  const request = deferred();
  h.read.mockReturnValue(request.promise);
  render();
  const cleanups = h.effects.map(effect => effect());
  cleanups.forEach(cleanup => cleanup?.());
  const before = JSON.stringify(h.slots);
  request.resolve({ member: { username: "old-account" } });
  await flush();
  expect(JSON.stringify(h.slots)).toBe(before);
});

test("invalid member data never reports rendered success", async () => {
  h.hasSession = true;
  h.read.mockResolvedValue({ member: [] });
  mount();
  await flush();
  const tree = render();
  expect(textContent(findByClass(tree, "mp-profile-state"))).toContain("服务端未返回完整账号资料");
  expect(findByClass(tree, "mp-profile-hero")).toBeUndefined();
});

test("logout navigates even when remote logout fails and ignores pending refresh", async () => {
  h.hasSession = true;
  h.read.mockResolvedValueOnce({ member: { username: "real-account" } });
  mount();
  await flush();
  const request = deferred();
  h.read.mockReturnValueOnce(request.promise);
  h.shows[1]();
  h.logout.mockRejectedValueOnce(new Error("network"));
  click("mp-profile-logout");
  expect(h.reLaunch).toHaveBeenCalledWith({ url: "/pages/bind/index" });
  request.resolve({ member: { username: "late-account" } });
  await flush();
  expect(textContent(render())).not.toContain("late-account");
});

test("profile does not mount per-profile compositor components", () => {
  const source = readFileSync(new URL("./index.tsx", import.meta.url), "utf8");
  expect(source).not.toContain("<PageTextureRefraction");
  expect(source).not.toContain("<AmbientStarfield");
});

test("profile hides the internal version row while login keeps its build label", () => {
  const profileSource = readFileSync(new URL("./index.tsx", import.meta.url), "utf8");
  const loginSource = readFileSync(new URL("../../components/LoginScene.tsx", import.meta.url), "utf8");
  expect(profileSource).not.toContain("mp-profile-diagnostic");
  expect(profileSource).not.toContain("MINIPROGRAM_VERSION_LABEL");
  expect(loginSource).toContain("MINIPROGRAM_VERSION_LABEL");
  expect(loginSource).toContain('className="nl-version"');
});

test("mount starts the profile request when the native did-show callback is missed", () => {
  h.hasSession = true;
  h.read.mockReturnValue(new Promise(() => {}));
  render();
  h.effects.forEach(effect => effect());
  expect(h.read).toHaveBeenCalledTimes(1);
});
