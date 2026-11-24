import { beforeEach, expect, test, vi } from "vitest";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

const h = vi.hoisted(() => ({
  slots: [] as any[], cursor: 0, show: undefined as any, hide: undefined as any,
  navigate: vi.fn(), switchTab: vi.fn(), save: vi.fn(), session: vi.fn(), loadSummary: vi.fn(),
}));

vi.mock("react", () => ({
  useState: (initial: any) => { const i = h.cursor++; if (!(i in h.slots)) h.slots[i] = typeof initial === "function" ? initial() : initial; return [h.slots[i], (next: any) => { h.slots[i] = typeof next === "function" ? next(h.slots[i]) : next; }]; },
  useRef: (initial: any) => { const i = h.cursor++; if (!(i in h.slots)) h.slots[i] = { current: initial }; return h.slots[i]; },
  useEffect: vi.fn(),
  memo: (component: any) => component,
}));
vi.mock("@tarojs/taro", () => ({
  default: { getStorageSync: () => "fetch", setStorageSync: h.save, navigateTo: h.navigate, switchTab: h.switchTab },
  useDidShow: (fn: any) => { h.show = fn; },
  useDidHide: (fn:any)=>{h.hide=fn;},
}));
vi.mock("@tarojs/components", () => ({ Button: "Button", Image: "Image", Text: "Text", View: "View" }));
vi.mock("../../theme", () => ({ ThemedPage: "ThemedPage" }));
vi.mock("../../components/BrandHeader", () => ({ default: () => null }));
vi.mock("../../components/AppIcon", () => ({ default: () => null }));
vi.mock("../../services/session", () => ({ hasAccountSession: () => true, readAccountSession: h.session }));
vi.mock("../../services/api", () => ({ apiRequest: vi.fn() }));
vi.mock("../../services/taskSummary", async (loadOriginal) => {
  const original = await loadOriginal<typeof import("../../services/taskSummary")>();
  return { ...original, loadTaskSummary: h.loadSummary, taskDestination: (entry: { id?: string }) => `/resume/${entry.id || ""}` };
});
vi.mock("../../config/localAuth", () => ({ isLocalAccountLogin: () => false }));
vi.mock("../functions/functionVisuals", () => ({ FUNCTION_VISUALS: { fetch: { asset: "fetch-art.webp", label: "小说创作", description: "获取、改文与提交" } } }));

import WorkspacePage from "./index";
import { moduleRoute } from "../../config/modules";

const nodes = (tree: any): any[] => !tree || typeof tree !== "object" ? [] : [tree, ...[tree.props?.children].flat(Infinity).flatMap(nodes)];
const render = () => { h.cursor = 0; return nodes(WorkspacePage()); };
const find = (name: string) => render().find((node) => node.props?.className?.split(" ").includes(name));
const text = (node: any): string => {
  if (typeof node === "string" || typeof node === "number") return String(node);
  if (!node || typeof node !== "object") return "";
  return [node.props?.children].flat(Infinity).map(text).join("");
};
const ready = async () => { render(); h.show(); for (let i = 0; i < 24; i += 1) await Promise.resolve(); };

beforeEach(() => {
  h.slots = []; h.navigate.mockReset(); h.switchTab.mockReset(); h.save.mockReset();
  h.session.mockReset().mockResolvedValue({ state: "authenticated", account: { username: "creator" } });
  h.loadSummary.mockReset().mockResolvedValue({ entries: [], sources: [{ source: "fetch", sourceLabel: "小说获取", state: "ready", error: undefined }] });
});

test("workbench removes the duplicate function catalog and shows the selected workspace", async () => {
  await ready();
  expect(find("studio-entry-grid")).toBeUndefined();
  expect(find("studio-choice-sheet")).toBeUndefined();
  expect(find("workbench-current-card")).toBeDefined();
  expect(text(find("workbench-current-title"))).toBe("小说创作");
});

test('workbench syncs only the selected source and stops reading after hide',async()=>{
 await ready();const options=h.loadSummary.mock.calls[0][1];
 expect(options.sources).toEqual(['fetch']);expect(options.shouldContinue()).toBe(true);
 h.hide();expect(options.shouldContinue()).toBe(false);
});

test("an empty workspace starts configuration through the real module route", async () => {
  await ready();
  expect(text(find("workbench-primary-action"))).toContain("开始配置");
  expect(find("workbench-project-card")).toBeUndefined();
  expect(find("workbench-empty")).toBeDefined();
  find("workbench-primary-action").props.onClick();
  expect(h.navigate).toHaveBeenLastCalledWith({ url: moduleRoute("fetch") });
});

test("the current module resumes only its own newest task", async () => {
  h.loadSummary.mockResolvedValue({
    entries: [
      { source: "script", sourceLabel: "剧本", id: "foreign", title: "其他功能", status: "running" },
      { source: "fetch", sourceLabel: "小说获取", id: "current", title: "夜宴的邀请", status: "running", createdAt: "2026-11-23T20:47:00+08:00", progress: { completed: 12, total: 18 } },
    ],
    sources: [{ source: "fetch", sourceLabel: "小说获取", state: "ready" }],
  });
  await ready();
  expect(text(find("workbench-primary-action"))).toContain("继续执行");
  expect(text(find("workbench-project-title"))).toContain("夜宴的邀请");
  expect(text(find("workbench-progress-value"))).toContain("67%");
  find("workbench-primary-action").props.onClick();
  expect(h.navigate).toHaveBeenLastCalledWith({ url: "/resume/current" });
});

test("missing progress is hidden instead of replaced with example numbers", async () => {
  h.loadSummary.mockResolvedValue({ entries: [{ source: "fetch", sourceLabel: "小说获取", id: "current", title: "真实项目", status: "unknown" }], sources: [{ source: "fetch", sourceLabel: "小说获取", state: "ready" }] });
  await ready();
  expect(find("workbench-progress-value")).toBeUndefined();
  expect(text(find("workbench-project-card"))).not.toContain("68%");
});

test("a failed summary keeps the selected module and offers a retry", async () => {
  h.loadSummary.mockRejectedValue(new Error("最近记录暂不可读取"));
  await ready();
  expect(text(find("workbench-current-title"))).toBe("小说创作");
  expect(find("workbench-error")).toBeDefined();
  find("workbench-retry").props.onClick();
  expect(h.loadSummary).toHaveBeenCalledTimes(2);
});

test("workbench stylesheet locks the mobile card hierarchy and safe area", () => {
  const styles = readFileSync(resolve(__dirname, "studio.scss"), "utf8");
  expect(styles).toContain(".workbench-current-card");
  expect(styles).toContain(".workbench-project-card");
  expect(styles).toContain(".workbench-detail-card");
  expect(styles).toContain("padding-bottom:calc(96PX + env(safe-area-inset-bottom))");
  expect(styles).toContain("min-height:128PX");
});
