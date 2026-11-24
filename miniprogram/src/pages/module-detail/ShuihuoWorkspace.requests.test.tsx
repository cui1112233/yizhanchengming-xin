// @vitest-environment jsdom
import { act } from "react";
import { createRoot, type Root } from "react-dom/client";
import { afterEach, beforeEach, expect, test, vi } from "vitest";
const h = vi.hoisted(() => ({ show: () => {}, hide: () => {}, api: vi.fn(), navigateTo: vi.fn() }));
vi.mock("@tarojs/taro", () => ({ default: { navigateTo: h.navigateTo }, useDidShow: (fn: any) => { h.show = fn; }, useDidHide: (fn: any) => { h.hide = fn; } }));
vi.mock("@tarojs/components", () => ({ View: "div", Text: "span", Button: "button", Input: "input", Textarea: "textarea", Image: "img", Video: "video", Audio: "audio" }));
vi.mock("../../services/api", () => ({ apiRequest: h.api }));
vi.mock("../../services/mediaCache", () => ({ loadAuthenticatedMedia: vi.fn() }));
vi.mock("../../theme", () => ({ withThemedPage: (p: any) => p }));
vi.mock("../../components/AppIcon", () => ({ default: () => null }));
vi.mock("../../components/BrandHeader", () => ({ default: () => null }));
vi.mock("../../components/AmbientStarfield", () => ({ default: () => null }));
import ShuihuoWorkspace from "./ShuihuoWorkspace";
let root: Root, container: HTMLDivElement;
beforeEach(() => { vi.clearAllMocks(); vi.stubGlobal("IS_REACT_ACT_ENVIRONMENT", true); container = document.createElement("div"); root = createRoot(container); });
afterEach(async () => { await act(async () => root.unmount()); vi.unstubAllGlobals(); });
const mount = async () => { await act(async () => root.render(<ShuihuoWorkspace />)); await act(async () => h.show()); };
test("distinguishes pending project reads from the real empty library", async () => {
  let finish!: (v: any) => void;
  h.api.mockImplementation((p: string) => p.endsWith("/projects") ? new Promise(r => { finish = r; }) : Promise.resolve(p.endsWith("/batches/summary") ? { batches: [] } : { models: [] }));
  await mount();
  expect(container.textContent).toContain("正在读取作品库");
  expect(container.textContent).not.toContain("暂无水货项目");
  await act(async () => finish({ projects: [] }));
  expect(container.textContent).toContain("暂无水货项目");
  expect(container.textContent).not.toContain("正在读取作品库");
});
test("a model read failure does not hide successfully loaded projects", async () => {
  h.api.mockImplementation((p: string) => p.endsWith("/projects") ? Promise.resolve({ projects: [{ id: 7, name: "已有水货项目" }] }) : Promise.reject(new Error("模型暂不可用")));
  await mount();
  expect(container.textContent).toContain("已有水货项目");
  expect(container.textContent).toContain("模型读取失败");
  expect(container.textContent).not.toContain("作品库读取失败");
});
test("hidden pages ignore old library results and reload on return", async () => {
  let finish!: (v: any) => void;
  h.api.mockImplementation((p: string) => p.endsWith("/projects") ? new Promise(r => { finish = r; }) : Promise.resolve(p.endsWith("/batches/summary") ? { batches: [] } : { models: [] }));
  await mount();
  await act(async () => { h.hide(); finish({ projects: [{ id: 1, name: "离开后的旧项目" }] }); });
  expect(container.textContent).not.toContain("离开后的旧项目");
  h.api.mockImplementation((p: string) => Promise.resolve(p.endsWith("/projects") ? { projects: [{ id: 2, name: "返回后的项目" }] } : { models: [] }));
  await act(async () => h.show());
  expect(container.textContent).toContain("返回后的项目");
});
test("shows shared batch works when the water-only library is empty and opens the exact batch", async () => {
  h.api.mockImplementation((p: string) => Promise.resolve(p.endsWith('/batches/summary')
    ? { batches: [{ id: 'batch/a', title: '共享作品10151', bookCount: 6, books: [] }] }
    : p.endsWith('/projects') ? { projects: [] } : { models: [] }));
  await mount();
  expect(container.textContent).toContain('共享作品10151');
  const button = [...container.querySelectorAll('button')].find(b => b.textContent?.includes('共享作品10151'))!;
  await act(async () => button.click());
  expect(h.navigateTo).toHaveBeenCalledWith({ url: '/pages/batch/index?batchId=batch%2Fa' });
});
test("a failed batch catalog stays distinct from an empty batch catalog and retains water projects", async () => {
  h.api.mockImplementation((p: string) => p.endsWith('/batches/summary') ? Promise.reject(new Error('作品服务暂不可用'))
    : Promise.resolve(p.endsWith('/projects') ? { projects: [{ id: 7, name: '可用水货项目' }] } : { models: [] }));
  await mount();
  expect(container.textContent).toContain('可用水货项目');
  expect(container.textContent).toContain('批量工厂作品读取失败');
  expect(container.textContent).not.toContain('暂无批量工厂作品');
});
test("hidden pages discard late shared batch results", async () => {
  let finish!: (v: any) => void;
  h.api.mockImplementation((p: string) => p.endsWith('/batches/summary') ? new Promise(r => { finish = r; })
    : Promise.resolve(p.endsWith('/projects') ? { projects: [] } : { models: [] }));
  await mount();
  expect(container.textContent).toContain('正在读取批量工厂作品');
  await act(async () => { h.hide(); finish({ batches: [{ id: 'old', title: '旧批量作品', books: [] }] }); });
  expect(container.textContent).not.toContain('旧批量作品');
});
