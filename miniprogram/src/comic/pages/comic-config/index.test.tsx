import {configureComicRecoveryScope,comicRunStorageKey,comicSelectedBookStorageKey} from '../../services/comicRecovery';
import { beforeEach, describe, expect, test, vi } from "vitest";

const h = vi.hoisted(() => ({
  slots: [] as any[],
  cursor: 0,
  effects: [] as Array<() => void | (() => void)>,
  routeParams: {} as Record<string, string>,
  setup: vi.fn(),
  books: vi.fn(),
  upload: vi.fn(),
  thumb: vi.fn(),
  submit: vi.fn(),
  chooseMedia: vi.fn(),
  navigateTo: vi.fn(),
  navigateBack: vi.fn(),
  getStorage: vi.fn(),
  setStorage: vi.fn(),
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
    chooseMedia: h.chooseMedia, navigateTo: h.navigateTo, navigateBack: h.navigateBack,
    getStorageSync: h.getStorage, setStorageSync: h.setStorage,
  },
  useRouter: () => ({ params: h.routeParams }),
}));
vi.mock("@tarojs/components", () => ({ Button: "Button", Image: "Image", ScrollView: "ScrollView", Text: "Text", View: "View" }));
vi.mock("../../../components/BrandHeader", () => ({ default: "Header" }));
vi.mock("../../../theme", () => ({ ThemedPage: "Page" }));
vi.mock("../../services/comicFlow", () => ({
  CLIENT_REFERENCE_LIMIT_BYTES: 3 * 1024 * 1024,
  DEFAULT_COMIC_RUN_MODE: "video_no_submit",
  DEFAULT_COMIC_ASPECT_RATIO: "9:16",
  comicBooksWithoutSource: (books: any[], bookIds: string[]) => books.filter((book) => bookIds.includes(book.id) && !book.hasSource),
  loadComicSetup: h.setup,
  loadComicBooks: h.books,
  resolveComicStyleThumb: h.thumb,
  uploadComicReference: h.upload,
  submitComicRun: h.submit,
}));

import ComicConfigPage from "./index";
import ComicStickySpacer, { COMIC_STICKY_SPACER_HEIGHT } from "../../components/ComicStickySpacer";

function render() { h.cursor = 0; return ComicConfigPage(); }
async function flush() { for (let index = 0; index < 40; index += 1) await Promise.resolve(); }
function runEffects() { for (const effect of h.effects.splice(0)) effect(); }
function textContent(node: any): string {
  if (Array.isArray(node)) return node.map(textContent).join("");
  if (node === null || node === undefined || typeof node === "boolean") return "";
  if (typeof node !== "object") return String(node);
  return textContent(node.props?.children);
}
function findAll(node: any, className: string, found: any[] = []): any[] {
  if (Array.isArray(node)) { node.forEach((child) => findAll(child, className, found)); return found; }
  if (!node || typeof node !== "object") return found;
  if (typeof node.props?.className === "string" && node.props.className.split(" ").includes(className)) found.push(node);
  findAll(node.props?.children, className, found);
  return found;
}
function nodesOfType(tree: any, type: any, found: any[] = []): any[] {
  if (Array.isArray(tree)) { tree.forEach((child) => nodesOfType(child, type, found)); return found; }
  if (!tree || typeof tree !== "object") return found;
  if (tree.type === type) found.push(tree);
  nodesOfType(tree.props?.children, type, found);
  return found;
}
function findByText(node: any, label: string): any {
  if (Array.isArray(node)) return node.map((child) => findByText(child, label)).find(Boolean);
  if (!node || typeof node !== "object") return undefined;
  if (typeof node.props?.onClick === "function" && textContent(node.props?.children).includes(label)) return node;
  return findByText(node.props?.children, label);
}

const SETUP = {
  batches: [{ id: "batch-1", title: "知乎故事", bookCount: 1 }],
  presets: [{ id: "preset-1", name: "默认生成配置", version: 1 }],
  styles: [
    { id: "style-1", name: "顶级韩潮风格", previewUrl: "" },
    { id: "style-2", name: "清新日漫", previewUrl: "" },
  ],
};

beforeEach(() => {configureComicRecoveryScope("test-owner");
  h.slots = [];
  h.effects = [];
  h.routeParams = {};
  vi.resetAllMocks();
  h.setup.mockResolvedValue(SETUP);
  h.books.mockResolvedValue([{ id: "book-1", title: "第一本", hasSource: true, revision: 1, sourceText: "这是已有书籍的正文" }]);
  h.upload.mockResolvedValue({ assetId: "asset-1", url: "/ref", previewUrl: "/ref-preview" });
  h.thumb.mockResolvedValue("");
  h.submit.mockResolvedValue({});
  h.navigateTo.mockResolvedValue({});
  h.getStorage.mockReturnValue("");
  h.setStorage.mockReturnValue(undefined);
  h.chooseMedia.mockResolvedValue({ tempFiles: [{ tempFilePath: "/tmp/ref.png", size: 1024 }] });
});

describe("漫画配置页", () => {
  test("默认值已选好，且界面不出现技术术语", async () => {
    render();
    runEffects();
    await flush();
    const text = textContent(render());
    expect(text).toContain("知乎故事");
    expect(text).toContain("这个批次只有 1 本书，已自动选中");
    expect(text).toContain("已选中");
    expect(text).toContain("开始生成");
    ["batchId", "runMode", "presetId", "settingsPatch", "设置补丁", "生成模式", "自动化预设", "提交到现有批量工厂"].forEach((word) => {
      expect(text).not.toContain(word);
    });
    const submitButton = findByText(render(), "开始生成");
    expect(submitButton.props.disabled).toBe(false);
  });

  test("点击画风卡片可以切换选中", async () => {
    render();
    runEffects();
    await flush();
    const cards = findAll(render(), "comic-style-card");
    expect(cards).toHaveLength(2);
    expect(cards[0].props.className).toContain("is-active");
    cards[1].props.onClick();
    await flush();
    const after = findAll(render(), "comic-style-card");
    expect(after[1].props.className).toContain("is-active");
    expect(after[0].props.className).not.toContain("is-active");
  });

  test("参考图超过 3 MB 时提示换图", async () => {
    h.chooseMedia.mockResolvedValue({ tempFiles: [{ tempFilePath: "/tmp/big.png", size: 4 * 1024 * 1024 }] });
    render();
    runEffects();
    await flush();
    findByText(render(), "选择图片").props.onClick();
    await flush();
    expect(h.upload).not.toHaveBeenCalled();
    expect(textContent(render())).toContain("参考图太大了");
  });

  test("没有画风时禁用主按钮并给出说明", async () => {
    h.setup.mockResolvedValue({ ...SETUP, styles: [] });
    render();
    runEffects();
    await flush();
    expect(findByText(render(), "开始生成").props.disabled).toBe(true);
    expect(textContent(render())).toContain("还没有可用的画风，先建一个画风再回来");
  });

  test("没有任何批次时给两个动作：新建批次、从已有批次继续", async () => {
    h.setup.mockResolvedValue({ ...SETUP, batches: [] });
    h.books.mockResolvedValue([]);
    render();
    runEffects();
    await flush();
    expect(findByText(render(), "开始生成").props.disabled).toBe(true);
    findAll(render(), "mp-primary").find((node: any) => textContent(node).includes("新建批次（粘贴正文开始）")).props.onClick();
    expect(h.navigateTo).toHaveBeenCalledWith({ url: "/comic/pages/comic-new/index" });
    expect(textContent(render())).toContain("从已有批次继续");
    expect(textContent(render())).not.toContain("网页版");
  });

  test("多本书默认只选第一本，并在按钮旁写明将生成几本", async () => {
    h.books.mockResolvedValue([
      { id: "book-1", title: "第一本", hasSource: true, revision: 1 },
      { id: "book-2", title: "第二本", hasSource: true, revision: 1 },
    ]);
    render();
    runEffects();
    await flush();
    const text = textContent(render());
    expect(text).toContain("已选 1 / 2 本书");
    expect(text).toContain("将生成 1 本书");
    expect(text).not.toContain("已选 2 / 2 本书");
    findByText(render(), "开始生成").props.onClick();
    await flush();
    expect(h.submit).toHaveBeenCalledWith(expect.objectContaining({ bookIds: ["book-1"] }));
  });

  test("点全选后才会生成全部书目", async () => {
    h.books.mockResolvedValue([
      { id: "book-1", title: "第一本", hasSource: true, revision: 1 },
      { id: "book-2", title: "第二本", hasSource: true, revision: 1 },
    ]);
    render();
    runEffects();
    await flush();
    findByText(render(), "全选").props.onClick();
    await flush();
    expect(textContent(render())).toContain("已选 2 / 2 本书");
    expect(textContent(render())).toContain("将生成 2 本书");
  });

  test("页面末尾有与吸底条等高的占位块，避免第 4 步被切掉", async () => {
    render();
    runEffects();
    await flush();
    const spacer = nodesOfType(render(), ComicStickySpacer).length;
    expect(spacer).toBe(1);
    expect(COMIC_STICKY_SPACER_HEIGHT).toContain("200rpx");
  });

  test("缩略图先占位，下载成功才换本地图片，失败保持占位", async () => {
    h.setup.mockResolvedValue({
      ...SETUP,
      styles: [
        { id: "style-1", name: "顶级国漫校园风格", previewUrl: "/api/presets/comic-style/style-1/preview" },
        { id: "style-2", name: "清新日漫", previewUrl: "/api/presets/comic-style/style-2/preview" },
      ],
    });
    h.thumb.mockImplementation((url: string) => Promise.resolve(url.includes("style-1") ? "/tmp/style-1.jpg" : ""));
    render();
    runEffects();
    await flush();
    const images = findAll(render(), "comic-style-thumb");
    expect(images.map((node: any) => node.type)).toEqual(["Image"]);
    expect(images[0].props.src).toBe("/tmp/style-1.jpg");
    // 失败的画风仍然显示首字占位，不再发渲染层网络请求。
    expect(textContent(render())).toContain("清");
    expect(h.thumb).toHaveBeenCalledTimes(2);
  });

  test("从新建页带 batchId 进来时直接定位到新作品", async () => {
    h.setup.mockResolvedValue({
      ...SETUP,
      batches: [{ id: "batch-1", title: "旧作品", bookCount: 1 }, { id: "batch-new", title: "刚建的作品", bookCount: 1 }],
    });
    h.routeParams = { batchId: "batch-new" };
    render();
    runEffects();
    await flush();
    expect(h.books).toHaveBeenCalledWith("batch-new");
    expect(textContent(render())).toContain("刚建的作品");
  });

  test("选择模式只展示已有书籍，并把正文带回上一页", async () => {
    h.routeParams = { mode: "pick" };
    render(); runEffects(); await flush();
    const tree = render();
    expect(textContent(tree)).toContain("选择已有书籍");
    expect(textContent(tree)).not.toContain("选画风");
    findByText(tree, "第一本").props.onClick();
    expect(h.setStorage).toHaveBeenCalledWith(comicSelectedBookStorageKey()!, {
      batchId: "batch-1", bookId: "book-1", title: "第一本", sourceText: "这是已有书籍的正文",
    });
    expect(h.navigateBack).toHaveBeenCalledWith({ delta: 1 });
  });

  test("选中的书没有正文时拦住生成，并给补正文入口", async () => {
    h.books.mockResolvedValue([{ id: "book-1", title: "第一本", hasSource: false, revision: 2 }]);
    render();
    runEffects();
    await flush();
    expect(findByText(render(), "开始生成").props.disabled).toBe(true);
    expect(textContent(render())).toContain("《第一本》还没有正文");
    expect(textContent(render())).toContain("先补上正文，才能开始做漫画");
    findByText(render(), "补正文").props.onClick();
    expect(h.navigateTo).toHaveBeenCalledWith({
      url: "/comic/pages/comic-new/index?batchId=batch-1&bookId=book-1&bookTitle=" + encodeURIComponent("第一本") + "&revision=2",
    });
  });

  test("提交时使用默认设置并跳到任务页", async () => {
    render();
    runEffects();
    await flush();
    findByText(render(), "开始生成").props.onClick();
    await flush();
    expect(h.submit).toHaveBeenCalledWith(expect.objectContaining({
      batchId: "batch-1", bookIds: ["book-1"], presetId: "preset-1",
      runMode: "video_no_submit", aspectRatio: "9:16", stylePresetId: "style-1",
    }));
    expect(h.navigateTo).toHaveBeenCalledWith(expect.objectContaining({ url: expect.stringContaining("/comic/pages/comic-task/index?batchId=batch-1") }));
  });
});
