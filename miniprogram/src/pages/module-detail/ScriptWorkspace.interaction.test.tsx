import { beforeEach, expect, test, vi } from "vitest";

const h = vi.hoisted(() => ({ slots: [] as any[], cursor: 0, show: null as any, api: vi.fn(), modal: vi.fn() }));
vi.mock("react", () => ({
  useState: (initial: any) => { const i = h.cursor++; if (!(i in h.slots)) h.slots[i] = typeof initial === "function" ? initial() : initial; return [h.slots[i], (next: any) => { h.slots[i] = typeof next === "function" ? next(h.slots[i]) : next; }]; },
  useRef: (initial: any) => { const i = h.cursor++; if (!(i in h.slots)) h.slots[i] = { current: initial }; return h.slots[i]; },
  useEffect: () => undefined,
}));
vi.mock("@tarojs/taro", () => ({ default: { showToast: vi.fn(), showModal: h.modal, getFileSystemManager: () => ({ unlink: vi.fn() }) }, useDidShow: (fn: any) => { h.show = fn; } }));
vi.mock("@tarojs/components", () => Object.fromEntries(["Button", "Input", "Picker", "Text", "Textarea", "View", "ScrollView", "Image", "Switch"].map(n => [n, n])));
vi.mock("../../theme", () => ({ withThemedPage: (fn: any) => fn }));
vi.mock("../../services/api", () => ({ apiRequest: h.api }));
vi.mock("../../components/AmbientStarfield", () => ({ default: () => null }));
vi.mock("../../components/BrandHeader", () => ({ default: () => null }));
vi.mock("../../components/AppIcon", () => ({ default: () => null }));
vi.mock("./ScriptVideoProduction", () => ({ default: () => null }));
vi.mock("./ScriptPersonalPrompts", () => ({ default: () => null }));
vi.mock("./ScriptEntityMedia", () => ({ default: () => null }));
import Workspace from "./ScriptWorkspace";

function render() { h.cursor = 0; return Workspace({}); }
function nodes(tree: any): any[] { return !tree || typeof tree !== "object" ? [] : [tree, ...[tree.props?.children].flat(Infinity).flatMap(nodes)]; }
function text(tree: any): string { return typeof tree === "string" || typeof tree === "number" ? String(tree) : tree?.props ? [tree.props.children].flat(Infinity).map(text).join("") : ""; }
function find(label: string) { const found = nodes(render()).find(n => n.type === "Button" && text(n) === label); expect(found, `button: ${label}`).toBeTruthy(); return found; }
async function setup() { render(); h.show(); await Promise.resolve(); await Promise.resolve(); render(); }
function inputNovel(value: string) { const input = nodes(render()).find(n => n.type === "Textarea" && n.props.placeholder === "粘贴小说原文"); expect(input).toBeTruthy(); input.props.onInput({ detail: { value } }); }
beforeEach(() => {
  h.slots = []; h.cursor = 0; h.api.mockReset(); h.modal.mockResolvedValue({confirm:true});
  h.api.mockImplementation(async (path: string, request: any) => {
    if (path === "/api/models?kind=text") return [{ id: "text-real", name: "真实文本模型" }];
    if (path === "/api/models?kind=image") return [];
    if (path === "/api/presets?module=script") return [{ id: "extract-real", name: "真实提取模板", kind: "base", extractionPreset: true }];
    if (path === "/api/chat" && request.data.promptType === "extract") return { characters: [{ id: "c1", name: "林夏", description: "旧描述" }], scenes: [] };
    if (path === "/api/chat") return { output: "真实返回的分镜" };
    if (path === "/api/history" && request?.method === "POST") return {};
    if (path.startsWith("/api/history/")) return { id: decodeURIComponent(path.slice("/api/history/".length)) };
    throw new Error(`unexpected ${path}`);
  });
});

test("collapsing the source preserves text for extraction and reopening", async () => {
  await setup(); inputNovel("她推开唱片店的门。");
  find("收起原文").props.onClick();
  expect(nodes(render()).some(n => n.type === "Textarea" && n.props.placeholder === "粘贴小说原文")).toBe(false);
  await find("提取人物场景").props.onClick();
  const request = h.api.mock.calls.find(([path, req]) => path === "/api/chat" && req.data.promptType === "extract");
  expect(request?.[1].data).toMatchObject({ novelText: "她推开唱片店的门。", textModelId: "text-real", extractionPreset: "extract-real" });
  find("展开编辑").props.onClick();
  expect(nodes(render()).find(n => n.type === "Textarea" && n.props.placeholder === "粘贴小说原文")?.props.value).toBe("她推开唱片店的门。");
});

test("the dock generates only after extraction and source edits reset readiness", async () => {
  await setup(); inputNovel("她推开门。"); await find("提取人物场景").props.onClick();
  expect(text(render())).toContain("林夏");
  await find("生成剧本").props.onClick();
  expect(h.api.mock.calls.find(([path, req]) => path === "/api/chat" && req.data.promptType === "script")?.[1].data).toMatchObject({ novelText: "她推开门。", characters: [{ id: "c1", name: "林夏", description: "旧描述" }] });
  find("返回编辑").props.onClick(); find("展开编辑").props.onClick(); inputNovel("换了一篇原文。");
  expect(find("提取人物场景").props.disabled).toBe(false);
  expect(nodes(render()).some(n => n.type === "Button" && text(n) === "生成剧本")).toBe(false);
});

test("entity edits remain drafts until applied and cancelled drafts never change generation", async () => {
  await setup(); inputNovel("她推开门。"); await find("提取人物场景").props.onClick();
  find("查看与编辑").props.onClick();
  const tree = render();
  const editor = nodes(tree).find(n => typeof n.type === "function" && n.props.value?.characters);
  expect(editor).toBeTruthy();
  editor.props.onChange({ characters: [{ id: "c1", name: "改名", description: "新描述" }], scenes: [] });
  find("取消修改").props.onClick();
  await find("生成剧本").props.onClick();
  expect(h.api.mock.calls.find(([path, req]) => path === "/api/chat" && req.data.promptType === "script")?.[1].data.characters[0].name).toBe("林夏");
});

test("applied entity drafts and selected modes reach the real generation payload", async () => {
  await setup(); inputNovel("她推开门。"); await find("提取人物场景").props.onClick();
  find("查看与编辑").props.onClick();
  const editor = nodes(render()).find(n => typeof n.type === "function" && n.props.value?.characters);
  editor.props.onChange({ characters: [{ id: "c1", name: "新名字", description: "新描述" }], scenes: [] });
  find("应用修改").props.onClick();
  nodes(render()).find(n => n.props?.id === 'script-settings').props.onClick();
  const picker = nodes(render()).find(n => n.type === "Picker" && n.props.range.includes("Q版"));
  picker.props.onChange({ detail: { value: "4" } });
  find('完成').props.onClick();
  await find("生成剧本").props.onClick();
  expect(h.api.mock.calls.find(([path, req]) => path === "/api/chat" && req.data.promptType === "script")?.[1].data).toMatchObject({ format: "q版", characters: [{ id: "c1", name: "新名字", description: "新描述" }] });
});

test("audio matching remains unavailable before real audio exists", async () => {
  await setup();
  const audioEntry = nodes(render()).find(n => n.type === "Button" && text(n).startsWith("原文配音"));
  audioEntry.props.onClick();
  expect(find("生成原文配音").props.disabled).toBe(true);
  const matching = nodes(render()).find(n => n.type === "Button" && text(n).startsWith("匹配音频"));
  expect(matching.props.disabled).toBe(true);
  expect(h.api.mock.calls.some(([path]) => path === "/api/tts")).toBe(false);
});

test("synchronous double taps submit extraction only once", async () => {
  await setup(); inputNovel('她推开门。');
  const action = find('提取人物场景').props.onClick;
  await Promise.all([action(), action()]);
  expect(h.api.mock.calls.filter(([path, request]) => path === '/api/chat' && request.data.promptType === 'extract')).toHaveLength(1);
});

test("model selection uses ids even when display names are identical", async () => {
  const normal = h.api.getMockImplementation()!;
  h.api.mockImplementation((path, options) => path === '/api/models?kind=text' ? Promise.resolve([{ id: 'first', name: '同名模型' }, { id: 'second', name: '同名模型' }]) : normal(path, options));
  await setup(); inputNovel('她推开门。');
  const picker = nodes(render()).find(n => n.type === 'Picker' && n.props.range.includes('同名模型'));
  picker.props.onChange({ detail: { value: '1' } });
  await find('提取人物场景').props.onClick();
  expect(h.api.mock.calls.find(([path, req]) => path === '/api/chat' && req.data.promptType === 'extract')?.[1].data.textModelId).toBe('second');
});

test("disabled models are not silently chosen for first extraction", async () => {
  const normal = h.api.getMockImplementation()!;
  h.api.mockImplementation((path, options) => path === '/api/models?kind=text' ? Promise.resolve([{ id: 'disabled', name: '停用', enabled: false }, { id: 'active', name: '启用', enabled: true }]) : normal(path, options));
  await setup(); inputNovel('她推开门。'); await find('提取人物场景').props.onClick();
  expect(h.api.mock.calls.find(([path, req]) => path === '/api/chat' && req.data.promptType === 'extract')?.[1].data.textModelId).toBe('active');
});

test("manual asset editing is available before paid extraction", async () => {
  await setup(); inputNovel('她推开门。'); find('手动添加人物场景').props.onClick();
  expect(nodes(render()).some(n => typeof n.type === 'function' && n.props.value?.characters)).toBe(true);
  expect(h.api.mock.calls.some(([path]) => path === '/api/chat')).toBe(false);
});

test('generation double taps submit and save one result only', async () => {
  await setup(); inputNovel('她推开门。'); await find('提取人物场景').props.onClick();
  const action=find('生成剧本').props.onClick; await Promise.all([action(),action()]);
  expect(h.api.mock.calls.filter(([path,options])=>path==='/api/chat'&&options.data.promptType==='script')).toHaveLength(1);
  expect(h.api.mock.calls.filter(([path,options])=>path==='/api/history'&&options?.method==='POST')).toHaveLength(1);
});
test('video task progress updates preserve generated text, while creative setting changes invalidate it', async () => {
  await setup();inputNovel('她推开门。');await find('提取人物场景').props.onClick();await find('生成剧本').props.onClick();
  const videos=nodes(render()).find(n=>typeof n.type==='function'&&n.props.output==='真实返回的分镜');expect(videos).toBeTruthy();
  videos.props.onTasksChange({0:{taskId:'video-real',status:'processing'}});expect(text(render())).toContain('真实返回的分镜');
  find('返回编辑').props.onClick();nodes(render()).find(n=>n.props?.id==='script-settings').props.onClick();
  nodes(render()).find(n=>n.type==='Picker'&&n.props.range.includes('Q版')).props.onChange({detail:{value:'4'}});find('完成').props.onClick();
  expect(find('创作结果').props.disabled).toBe(true);
});
test('history button opens private history and restores the explicitly selected record', async () => {
  await setup();nodes(render()).find(n=>n.props?.['aria-label']==='历史记录').props.onClick();
  const history=nodes(render()).find(n=>typeof n.type==='function'&&n.props.onSelect&&!n.props.category);expect(history).toBeTruthy();
  h.api.mockImplementation(async(path:string)=>path==='/api/history/h-real'?{id:'h-real',novelText:'历史原文',output:'历史结果',extractInfo:{characters:[],scenes:[]}}:{});
  history.props.onSelect('h-real');for(let i=0;i<20;i++)await Promise.resolve();
  expect(h.api).toHaveBeenCalledWith('/api/history/h-real');expect(text(render())).toContain('历史结果');
});
