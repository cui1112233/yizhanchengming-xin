import { beforeEach, describe, expect, test, vi } from "vitest";

const harness = vi.hoisted(() => ({ slots: [] as any[], cursor: 0, show: () => {}, api: vi.fn(), toast: vi.fn() }));
vi.mock("react", () => ({
  useState: (initial: any) => { const i = harness.cursor++; if (!(i in harness.slots)) harness.slots[i] = typeof initial === "function" ? initial() : initial; return [harness.slots[i], (next: any) => { harness.slots[i] = typeof next === "function" ? next(harness.slots[i]) : next; }]; },
  useRef: (initial: any) => { const i = harness.cursor++; if (!(i in harness.slots)) harness.slots[i] = { current: initial }; return harness.slots[i]; },
  useEffect: () => {},
}));
vi.mock("@tarojs/taro", () => ({ default: { showToast: harness.toast, showModal: () => Promise.resolve({ confirm: true }) }, useDidHide:()=>{}, useDidShow: (callback: () => void) => { harness.show = callback; } }));
vi.mock("@tarojs/components", () => Object.fromEntries(["Button", "Checkbox", "CheckboxGroup", "Input", "Picker", "ScrollView", "Text", "Textarea", "View"].map((name) => [name, name])));
vi.mock("../../theme", () => ({ withThemedPage: (page: any) => page }));
vi.mock("../../components/AmbientStarfield", () => ({ default: "Starfield" }));
vi.mock("../../components/BrandHeader", () => ({ default: "Header" }));
vi.mock("../../components/AppIcon", () => ({ default: "Icon" }));
vi.mock("../../services/api", () => ({ apiRequest: harness.api }));
import FetchWorkspace from "./FetchWorkspace";

const config = { platforms: [{ id: "one", name: "一" }, { id: "two", name: "二" }], app_config: { text_model_id: "A", fetch: { max_txt: 4000 } } };
const models = [{ id: "A", name: "模型 A" }, { id: "B", name: "模型 B" }];
const deferred = () => { let resolve!: (value: any) => void; const promise = new Promise<any>((done) => { resolve = done; }); return { promise, resolve }; };
const flush = async () => { for (let i = 0; i < 12; i++) await Promise.resolve(); };
function render() { harness.cursor = 0; return FetchWorkspace({}); }
function nodes(tree: any): any[] { if (!tree || typeof tree !== "object") return []; return [tree, ...[tree.props?.children].flat(Infinity).flatMap(nodes)]; }
function text(tree: any): string { if (typeof tree === "string" || typeof tree === "number") return String(tree); return tree?.props ? [tree.props.children].flat(Infinity).map(text).join("") : ""; }
function pickers() { return nodes(render()).filter((node) => node.type === "Picker"); }
function click(label: string) { const node = nodes(render()).find((item) => item.props?.onClick && (label === "展开" ? item.props?.className === "fetch-book-row" : text(item) === label)); expect(node, label).toBeTruthy(); return node.props.onClick(); }

beforeEach(() => {
  harness.slots = []; harness.cursor = 0; harness.api.mockReset(); harness.toast.mockReset();
  harness.api.mockImplementation((path: string) => Promise.resolve(path.includes("/models") ? models : path.endsWith("/config") ? config : path.endsWith("/tasks") ? { tasks: [] } : { schedules: [] }));
});

describe("FetchWorkspace deferred requests", () => {
  test("compact rows open one book detail without showing actions in the list", async () => {
    harness.api.mockImplementation((path: string) => Promise.resolve(path.includes("/models") ? models : path.endsWith("/config") ? config : path.endsWith("/tasks") ? { tasks: [{ book_id: "first", book_name: "小说甲" }, { book_id: "second", book_name: "小说乙" }] } : path.endsWith("/first") ? { original: "第一本正文" } : { schedules: [] }));
    render(); harness.show(); await flush(); click("任务"); await flush();
    expect(nodes(render()).filter(n => n.props?.className === "fetch-book-row")).toHaveLength(2);
    expect(text(render())).not.toContain("预览网站提交");
    expect(text(render())).not.toContain("抓取原文");
    click("展开"); await flush();
    expect(harness.api.mock.calls.filter(([path]) => path.endsWith("/first"))).toHaveLength(1);
    expect(text(render())).toContain("抓取原文");
    expect(nodes(render()).find(n => n.type === "Textarea")?.props.value).toBe("第一本正文");
    click("关闭");
    expect(text(render())).not.toContain("抓取原文");
    expect(nodes(render()).filter(n => n.props?.className === "fetch-book-row")).toHaveLength(2);
  });
  test("collapsing booklist and closing settings preserves the original process payload", async () => {
    render(); harness.show(); await flush();
    const start = () => nodes(render()).find(n => n.type === "Button" && text(n) === "开始处理");
    expect(start().props.disabled).toBe(true);
    const booklist = "123456 小说甲\n789012 小说乙";
    nodes(render()).find(n => n.type === "Textarea").props.onInput({ detail: { value: booklist } });
    pickers()[0].props.onChange({ detail: { value: 1 } });
    click("收起书单");
    expect(nodes(render()).some(n => n.type === "Textarea")).toBe(false);
    nodes(render()).find(n => n.props?.className === "mp-fetch-setting" && text(n).startsWith("处理参数")).props.onClick();
    const before = harness.api.mock.calls.length;
    click("收起面板");
    expect(harness.api.mock.calls).toHaveLength(before);
    click("展开书单");
    expect(nodes(render()).find(n => n.type === "Textarea").props.value).toBe(booklist);
    expect(start().props.disabled).toBe(false);
    await start().props.onClick();
    const write = harness.api.mock.calls.find(([path]) => path.endsWith("/process"));
    expect(write?.[1].data).toMatchObject({ input_text: booklist, platform_id: "two", max_txt: 4000 });
  });
  test("editing a schedule preserves its original input snapshot and can turn autoSubmit off", async () => {
    const schedule = { id: "s1", enabled: false, autoSubmit: true, runAt: "2099-01-01T00:00:00Z", inputSnapshot: { input_text: "saved book" } };
    harness.api.mockImplementation((path: string) => Promise.resolve(path.includes("/models") ? models : path.endsWith("/config") ? config : path.endsWith("/tasks") ? { tasks: [] } : { schedules: [schedule] }));
    render(); harness.show(); await flush();
    nodes(render()).find(n => n.props?.onClick && n.props?.className === "mp-fetch-setting" && text(n).startsWith("定时任务")).props.onClick(); await flush();
    click("编辑");
    nodes(render()).find(n => n.type === "CheckboxGroup").props.onChange({ detail: { value: [] } });
    await click("保存定时修改");
    const write = harness.api.mock.calls.find(([, options]) => options?.method === "PATCH");
    expect(write?.[1].data).toEqual({ runAt: "2099-01-01T00:00:00.000Z", autoSubmit: false });
    expect(write?.[1].data).not.toHaveProperty("inputSnapshot");
  });
  test("delete schedule confirms and reads the resulting list without replay", async () => {
    harness.api.mockImplementation((path: string) => Promise.resolve(path.includes("/models") ? models : path.endsWith("/config") ? config : path.endsWith("/tasks") ? { tasks: [] } : { schedules: [{ id: "s1" }] }));
    render(); harness.show(); await flush();
    nodes(render()).find(n => n.props?.onClick && n.props?.className === "mp-fetch-setting" && text(n).startsWith("定时任务")).props.onClick(); await flush();
    await click("删除定时");
    expect(harness.api.mock.calls.filter(([, o]) => o?.method === "DELETE")).toEqual([["/api/batch-rewrite/schedules/s1", { method: "DELETE" }]]);
    expect(harness.api.mock.calls[harness.api.mock.calls.length - 1][0]).toBe("/api/batch-rewrite/schedules");
  });
  test("schedule maintenance uses PATCH and confirms disabled state by readback", async () => {
    let enabled = true;
    harness.api.mockImplementation((path: string, options: any) => {
      if (options?.method === "PATCH") { enabled = options.data.enabled; return Promise.resolve({ ok: true }); }
      return Promise.resolve(path.includes("/models") ? models : path.endsWith("/config") ? config : path.endsWith("/tasks") ? { tasks: [] } : { schedules: [{ id: "one/2", enabled, autoSubmit: false, runAt: "2027-01-01T00:00:00Z" }] });
    });
    render(); harness.show(); await flush();
    nodes(render()).find(n => n.props?.onClick && n.props?.className === "mp-fetch-setting" && text(n).startsWith("定时任务")).props.onClick(); await flush();
    await click("停用"); await flush();
    expect(harness.api.mock.calls.find(([, options]) => options?.method === "PATCH")).toEqual(["/api/batch-rewrite/schedules/one%2F2", { method: "PATCH", data: { enabled: false } }]);
    expect(text(render())).toContain("已停用");
  });
  test("AI4 and AI5 can be selected and saved using the same version contract as the working web client", async () => {
    render(); harness.show(); await flush();
    nodes(render()).find((item) => item.props?.onClick && item.props?.className === "mp-fetch-setting" && text(item).startsWith("版本配置")).props.onClick();
    expect(nodes(render()).filter((item) => item.type === "Checkbox").map((item) => item.props.value)).toEqual(["original", "ai1", "ai2", "ai3", "ai4", "ai5"]);
    nodes(render()).find((item) => item.type === "CheckboxGroup").props.onChange({ detail: { value: ["original", "ai4", "ai5"] } });
    await click("保存处理规则");
    expect(harness.api.mock.calls.find(([path, options]) => path.endsWith("/work-form") && options?.method === "POST")?.[1].data.state.selected_versions).toEqual(["original", "ai4", "ai5"]);
    expect(harness.api.mock.calls[harness.api.mock.calls.length - 1][0]).toBe("/api/batch-rewrite/config");
  });
  test("a failed schedules read does not hide successfully loaded account tasks and platforms", async () => {
    harness.api.mockImplementation((path: string) => path.endsWith("/schedules") ? Promise.reject(new Error("定时服务暂不可用")) : Promise.resolve(path.includes("/models") ? models : path.endsWith("/config") ? config : { tasks: [{ id: "available", book_name: "可读取的小说" }] }));
    render(); harness.show(); await flush();
    expect(text(nodes(render()).find((item) => item.props?.className === "mp-fetch-meta"))).toContain("1 个任务");
    click("任务"); await flush();
    expect(text(render())).toContain("可读取的小说");
    expect(text(render())).toContain("定时服务暂不可用");
    click("处理"); expect(pickers()[0].props.range).toEqual(["一", "二"]);
  });
  test("existing historical tasks appear by default, and today distinguishes hidden history from empty account", async () => {
    harness.api.mockImplementation((path: string) => Promise.resolve(path.includes("/models") ? models : path.endsWith("/config") ? config : path.endsWith("/tasks") ? { tasks: [{ id: "history", book_name: "历史小说", created_at: "2026-09-01T08:00:00Z" }] } : { schedules: [] }));
    render(); harness.show(); await flush(); click("任务"); await flush();
    expect(text(render())).toContain("历史小说");
    click("今天"); expect(text(render())).not.toContain("历史小说");
    expect(text(render())).toContain("已有 1 个任务被日期筛选隐藏");
    click("全部"); expect(text(render())).toContain("历史小说");
  });

  test("public detail contract displays original and each generated version as actual text", async () => {
    harness.api.mockImplementation((path: string) => Promise.resolve(path.includes("/models") ? models : path.endsWith("/config") ? config : path.endsWith("/tasks") ? { tasks: [{ id: "history", book_name: "历史小说", created_at: "2026-09-01T08:00:00Z" }] } : path.endsWith("/history") ? { meta: { book_id: "history" }, original: "原著正文", ai_texts: [{ version: "ai1", name: "AI1", text: "改写正文一" }, { version: "ai3", name: "AI3", text: "改写正文三" }] } : { schedules: [] }));
    render(); harness.show(); await flush(); click("任务"); await flush(); click("全部"); click("展开"); click("查看"); await flush();
    expect(nodes(render()).filter((node) => node.type === "Textarea").map((node) => node.props.value)).toEqual(["原著正文", "改写正文一", "改写正文三"]);
  });

  test.each(["重试失败", "重试"])("%s uses synchronous deadline, lock and truthful failed counts", async (label) => {
    harness.api.mockImplementation((path: string) => Promise.resolve(path.includes("/models") ? models : path.endsWith("/config") ? config : path.endsWith("/tasks") ? { tasks: [{ book_id: "retry", book_name: "重试任务", created_at: new Date().toISOString() }] } : { schedules: [] }));
    render(); harness.show(); await flush(); click("任务"); await flush(); click("展开");
    const pending = deferred();
    harness.api.mockImplementation((path: string) => path.endsWith("batch-retry") ? pending.promise : Promise.resolve({ tasks: [] }));
    click(label); click(label);
    const writes = harness.api.mock.calls.filter(([path]) => path.endsWith("batch-retry"));
    expect(writes).toHaveLength(1); expect(writes[0][1].timeout).toBe(180000);
    expect(writes[0][1].data.mode).toBe(label === "重试" ? "selected" : "failed");
    pending.resolve({ retried: 0, failed: 1, skipped: 0, tasks: [] }); await flush();
    expect(text(render())).toContain("重试已返回：完成 0，失败 1，跳过 0");
    expect(harness.toast).not.toHaveBeenCalled();
  });

  test("retry timeout blocks another attempt until successful explicit refresh", async () => {
    render(); harness.show(); await flush(); click("任务"); await flush();
    harness.api.mockRejectedValue(new Error("timeout")); click("重试失败"); await flush();
    expect(text(render())).toContain("结果未知");
    const count = harness.api.mock.calls.length; click("重试失败"); await flush(); expect(harness.api.mock.calls).toHaveLength(count);
  });

  test("closing task detail prevents a delayed response from reopening it", async () => {
    harness.api.mockImplementation((path: string) => Promise.resolve(path.includes("/models") ? models : path.endsWith("/config") ? config : path.endsWith("/tasks") ? { tasks: [{ book_id: "detail", book_name: "详情任务", created_at: new Date().toISOString() }] } : { schedules: [] }));
    render(); harness.show(); await flush(); click("任务"); await flush(); click("展开");
    const pending = deferred(); harness.api.mockReturnValue(pending.promise);
    click("查看"); click("关闭"); pending.resolve({ original: "晚到原文" }); await flush();
    expect(nodes(render()).filter((node) => node.props?.className === "mp-fetch-detail-panel")).toHaveLength(0);
  });

  test("synchronous fetch uses a long timeout and blocks retry after unknown outcome", async () => {
    harness.api.mockImplementation((path: string) => Promise.resolve(path.includes("/models") ? models : path.endsWith("/config") ? config : path.endsWith("/tasks") ? { tasks: [{ book_id: "detail", book_name: "详情任务", created_at: new Date().toISOString() }] } : { schedules: [] }));
    render(); harness.show(); await flush(); click("任务"); await flush(); click("展开");
    harness.api.mockRejectedValue(new Error("request timeout"));
    click("抓取原文"); await flush();
    expect(harness.api.mock.calls[harness.api.mock.calls.length - 1]?.[1]?.timeout).toBe(180000);
    expect(text(render())).toContain("结果未知");
    const calls = harness.api.mock.calls.length; click("抓取原文"); await flush();
    expect(harness.api.mock.calls).toHaveLength(calls);
    expect(harness.toast).not.toHaveBeenCalled();
  });

  test("editing platform during show refresh preserves the new selection", async () => {
    render(); harness.show(); await flush();
    const pending = deferred();
    harness.api.mockImplementation((path: string) => path.endsWith("/config") ? pending.promise : Promise.resolve(path.includes("/models") ? models : path.endsWith("/tasks") ? { tasks: [] } : { schedules: [] }));
    harness.show();
    pickers()[0].props.onChange({ detail: { value: 1 } });
    click("处理参数4000 字");
    nodes(render()).find((node) => node.type === "Input").props.onInput({ detail: { value: "8000" } });
    pending.resolve(config); await flush();
    expect(pickers()[0].props.value).toBe(1);
    expect(nodes(render()).find((node) => node.type === "Input").props.value).toBe("8000");
  });

  test("a newer task read wins over a pending combined refresh", async () => {
    render(); harness.show(); await flush();
    const pending = deferred(); let reads = 0;
    harness.api.mockImplementation((path: string) => Promise.resolve(path.includes("/models") ? models : path.endsWith("/config") ? config : path.endsWith("/tasks") ? (++reads === 1 ? pending.promise : { tasks: [{ book_id: "new", book_name: "新任务", created_at: new Date().toISOString() }] }) : { schedules: [] }));
    harness.show(); click("任务"); await flush();
    pending.resolve({ tasks: [{ book_id: "old", book_name: "旧任务", created_at: new Date().toISOString() }] }); await flush();
    expect(text(render())).toContain("新任务"); expect(text(render())).not.toContain("旧任务");
  });

  test("model saves reject rapid competing writes until confirmed readback", async () => {
    render(); harness.show(); await flush();
    const write = deferred(), read = deferred();
    harness.api.mockImplementation((_path: string, options: any) => options?.method === "POST" ? write.promise : read.promise);
    const picker = pickers()[1];
    picker.props.onChange({ detail: { value: 1 } });
    picker.props.onChange({ detail: { value: 0 } });
    expect(harness.api.mock.calls.filter(([, options]) => options?.method === "POST")).toHaveLength(1);
    write.resolve({ ok: true }); await flush();
    expect(pickers()[1].props.disabled).toBe(true);
    read.resolve({ ...config, app_config: { text_model_id: "B" } }); await flush();
    expect(pickers()[1].props.disabled).toBe(false);
    expect(pickers()[1].props.value).toBe(1);
  });
});
