import { beforeEach, expect, test, vi } from "vitest";

const h = vi.hoisted(() => ({ value: "fetch", reduced: false, slots: [] as any[], cursor: 0, show: null as any, navigate: vi.fn(), switchTab: vi.fn(), save: vi.fn() }));
vi.mock("react", () => ({ useState: (initial: any) => {
  const i = h.cursor++;
  if (!(i in h.slots)) h.slots[i] = typeof initial === "function" ? initial() : initial;
  return [h.slots[i], (value: any) => { h.slots[i] = value; }];
} }));
vi.mock("@tarojs/taro", () => ({ default: { getStorageSync: (key: string) => key === "yzcm.mobile.glass.reduceTransparency" ? h.reduced : h.value, setStorageSync: h.save, navigateTo: h.navigate, switchTab: h.switchTab }, useDidShow: (fn: any) => { h.show = fn; } }));
vi.mock("@tarojs/components", () => ({ Button: "Button", Image: "Image", Switch: "Switch", Text: "Text", View: "View" }));
vi.mock("../../theme", () => ({ ThemedPage: "ThemedPage", applyMobileTheme: vi.fn(), readMobileTheme: () => "light" }));
vi.mock("../../components/BrandHeader", () => ({ default: () => null }));
vi.mock("../../components/AppIcon", () => ({ default: () => null }));
vi.mock("./NativeGlassMaterial", () => ({ default: () => null }));
import FunctionsPage from "./index";
import { MODULES, moduleRoute } from "../../config/modules";
import { readFileSync } from "node:fs";

const galleryStyles = readFileSync(new URL("./gallery.scss", import.meta.url), "utf8").replace(/\s+/g, "").toLowerCase();

const nodes = (tree: any): any[] => !tree || typeof tree !== "object" ? [] : [tree, ...[tree.props?.children].flat(Infinity).flatMap(nodes)];
const render = () => { h.cursor = 0; return nodes(FunctionsPage()); };
beforeEach(() => { h.slots = []; h.value = "fetch"; h.reduced = false; h.navigate.mockReset(); h.switchTab.mockReset(); h.save.mockReset(); });

test("each grouped entry opens its real function and remembers the choice", () => {
  expect(render().filter(n => n.type === "Button" && n.props.className.split(" ").includes('tools-entry'))).toHaveLength(9);
  for (const item of MODULES) {
    render().find(n => n.type === "Button" && n.props.className.split(" ").includes(`tools-entry-${item.key}`)).props.onClick();
    expect(h.save).toHaveBeenLastCalledWith("yzcm.mobile.activeModule", item.key);
    expect(h.navigate).toHaveBeenLastCalledWith({ url: moduleRoute(item.key) });
  }
  expect(h.switchTab).not.toHaveBeenCalled();
});

test("gallery keeps every real module exactly once without duplicate routes", () => {
  const entries = render().filter(n => n.type === "Button" && n.props.className.split(" ").includes("tools-entry"));
  expect(entries).toHaveLength(MODULES.length);
  expect(new Set(entries.map(entry => MODULES.find(item => entry.props.className.includes(`tools-entry-${item.key}`))?.key)).size).toBe(MODULES.length);
  expect(h.navigate).not.toHaveBeenCalled();
  expect(h.save).not.toHaveBeenCalled();
});

test("gallery uses the selected visual hierarchy instead of category filters", () => {
  const allText = render().map(n => typeof n.props?.children === "string" ? n.props.children : "").join(" ");
  expect(allText).toContain("我的创作空间");
  expect(allText).toContain("继续创作");
  expect(allText).toContain("内容创作");
  expect(allText).toContain("生产成片");
  expect(allText).toContain("AI 与工具");
  expect(render().find(n => n.props?.className?.includes("studio-function-filter"))).toBeUndefined();
});

test("matches the approved nine-card copy and visual order", () => {
  const entries = render().filter(n => n.type === "Button" && n.props.className.split(" ").includes("tools-entry"));
  expect(entries.map(entry => MODULES.find(item => entry.props.className.includes(`tools-entry-${item.key}`))?.key)).toEqual([
    "script", "fetch", "batch", "shuihuo", "aicut", "comic", "agent", "voice", "plugins",
  ]);
  const allText = render().map(n => typeof n.props?.children === "string" ? n.props.children : "").join(" ");
  expect(allText).toContain("私人定制");
  expect(allText).toContain("小说创作");
  expect(allText).toContain("批量测书");
  expect(allText).toContain("水货生产");
  expect(allText).toContain("AI-CUT");
  expect(allText).toContain("漫剧创作");
  expect(allText).toContain("CM 画布");
  expect(allText).toContain("配音");
  expect(allText).toContain("插件");
});

test("uses the approved immersive hero and a dedicated illustration on every card", () => {
  const tree = render();
  expect(tree.find(n => n.props?.className === "gallery-hero-shell")).toBeDefined();
  expect(tree.find(n => n.props?.className === "gallery-hero-art")).toBeDefined();
  const entries = tree.filter(n => n.type === "Button" && n.props.className.split(" ").includes("tools-entry"));
  expect(entries.every(entry => nodes(entry).some(n => n.type === "Image"))).toBe(true);
  expect(tree.find(n => n.props?.className === "gallery-resume-art")).toBeUndefined();
});

test("returning to the page refreshes the last-used marker from other workspaces", () => {
  render(); h.value = "batch"; h.show();
  const selected = render().filter(n => n.type === "Button" && n.props.className.includes("is-recent"));
  expect(selected).toHaveLength(1);
  expect(selected[0].props.className).toContain("tools-entry-batch");
});

test("function entry keeps its routes without an appearance preference control", () => {
  expect(render().find(n => n.type === "Switch")).toBeUndefined();
  expect(render().find(n => n.props?.className?.includes("is-recent")).props.className).toContain("tools-entry-fetch");
  expect(h.navigate).not.toHaveBeenCalled();
});

test("function home disables the full-screen native canvas while keeping the liquid shell", () => {
  const page = render().find(n => n.type === "ThemedPage");
  expect(page.props.dynamicGlass).toBe(false);
  expect(page.props.className).toContain("mp-primary-page");
});

test("uses WeChat-safe sizing instead of auto-height images or CSS grid", () => {
  expect(galleryStyles).toContain("height:73.6111vw");
  expect(galleryStyles).not.toContain("height:auto");
  expect(galleryStyles).not.toContain("display:grid");
  expect(galleryStyles).not.toContain("grid-template");
  expect(galleryStyles).toContain("padding:015px");
  expect(galleryStyles).toContain("gap:10px");
  expect(galleryStyles).toContain("height:138px");
  expect(galleryStyles).toContain("height:62px");
  expect(galleryStyles).toContain("font-size:20px");
  expect(galleryStyles).toContain("font-size:13px");
  expect(galleryStyles).toContain("left:104px");
  expect(galleryStyles).toContain("transform:scale(1.25)");
  expect(galleryStyles).toContain(".tools-entry-script.gallery-feature-art{left:auto;right:-10px;bottom:-15px;width:160px;height:160px;transform:none}");
  expect(galleryStyles).toContain(".tools-entry-comic.gallery-feature-art{left:auto;right:-35px;bottom:-42px;width:210px;height:210px;transform:none}");
});
