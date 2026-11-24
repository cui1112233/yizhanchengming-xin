import {configureComicRecoveryScope,comicRunStorageKey,comicSelectedBookStorageKey} from '../../services/comicRecovery';
import { beforeEach, describe, expect, test, vi } from "vitest";

const h = vi.hoisted(() => ({
  slots: [] as any[],
  childSlots: new Map<string, any[]>(),
  active: null as any[] | null,
  cursor: 0,
  effects: [] as Array<() => void | (() => void)>,
  seen: new Set<string>(),
  setup: vi.fn(),
  thumb: vi.fn(),
  create: vi.fn(),
  upload: vi.fn(),
  submit: vi.fn(),
  readStatus: vi.fn(),
  pollPublish: null as any,
  pollFail: null as any,
  pollStop: vi.fn(),
  autoFinish: true,
  // 轮询首帧下发的服务端状态，用例可改写它来驱动 runView.stage（例如 comic_assets）。
  pollStatus: null as any,
  pollStarts: 0,
  chooseMedia: vi.fn(),
  getClipboardData: vi.fn(),
  previewImage: vi.fn(),
  showToast: vi.fn(),
  showLoading: vi.fn(),
  hideLoading: vi.fn(),
  showModal: vi.fn(),
  showActionSheet: vi.fn(),
  setClipboardData: vi.fn(),
  openSetting: vi.fn(),
  navigateTo: vi.fn(),
  setPrimary: vi.fn(),
  regeneratePage: vi.fn(),
  pages: vi.fn(),
  loadPaths: vi.fn(),
  saveAlbum: vi.fn(),
  createCollection: vi.fn(),
  zip: vi.fn(),
  assets: vi.fn(),
  assetImages: vi.fn(),
  assetPath: vi.fn(),
  extractAssets: vi.fn(),
  generateImages: vi.fn(),
  assetTask: vi.fn(),
  saveAsset: vi.fn(),
  savePrompts: vi.fn(),
  replaceImage: vi.fn(),
  assetPrimary: vi.fn(),
  retryStage: vi.fn(),
  storage: new Map<string, any>(),
  show: null as any,
  hide: null as any,
}));

vi.mock("react", () => ({
  useState: (initial: any) => {
    const index = h.cursor++;
    const slots = h.active || h.slots;
    if (!(index in slots)) slots[index] = typeof initial === "function" ? initial() : initial;
    return [slots[index], (next: any) => { slots[index] = typeof next === "function" ? next(slots[index]) : next; }];
  },
  useRef: (initial: any) => {
    const index = h.cursor++;
    const slots = h.active || h.slots;
    if (!(index in slots)) slots[index] = { current: initial };
    return slots[index];
  },
  useEffect: (effect: any, deps?: any[]) => {
    const index = h.cursor++;
    const prior = (h.active || h.slots)[index];
    if (prior && deps && JSON.stringify(prior.deps) === JSON.stringify(deps)) return;
    const entry = { deps, cleanup: undefined as any };
    (h.active || h.slots)[index] = entry;
    h.effects.push(() => {
      prior?.cleanup?.();
      entry.cleanup = effect();
    });
  },
}));

vi.mock("@tarojs/taro", () => ({
  default: {
    chooseMedia: h.chooseMedia,
    getClipboardData: h.getClipboardData,
    previewImage: h.previewImage,
    showToast: h.showToast,
    showLoading: h.showLoading,
    hideLoading: h.hideLoading,
    showModal: h.showModal,
    showActionSheet: h.showActionSheet,
    setClipboardData: h.setClipboardData,
    openSetting: h.openSetting,
    navigateTo: h.navigateTo,
    getStorageSync: (key: string) => h.storage.get(key) ?? "",
    setStorageSync: (key: string, value: any) => { h.storage.set(key, value); },
    removeStorageSync: (key: string) => { h.storage.delete(key); },
  },
  useDidShow: (callback: any) => { h.show = callback; },
  useDidHide: (callback: any) => { h.hide = callback; },
}));
vi.mock("@tarojs/components", () => ({
  Button: "Button", Image: "Image", ScrollView: "ScrollView", Text: "Text", Input: "Input", Textarea: "Textarea", View: "View",
}));
vi.mock("../../../components/BrandHeader", () => ({ default: "Header" }));
vi.mock("../../../theme", () => ({ ThemedPage: "Page" }));
vi.mock("../../services/comicFlow", async (importOriginal) => {
  const actual = await importOriginal<typeof import("../../services/comicFlow")>();
  return {
    COMIC_SOURCE_TRUNCATE_HINT_CHARS: 8000,
    COMIC_PRESET_EMPTY_HINT: "当前没有可用于漫画的预设，请先在执行器里配置图片模型",
    COMIC_ASPECT_RATIOS: ["16:9", "9:16", "21:9", "1:1", "4:3", "3:4", "3:2", "2:3"],
    DEFAULT_COMIC_ASPECT_RATIO: "9:16",
    DEFAULT_COMIC_RUN_MODE: "video_no_submit",
    MIN_COMIC_SOURCE_CHARS: 100,
    comicSourceLength: (value: string) => String(value || "").trim().length,
    validateComicSource: (value: string) => (String(value || "").trim().length < 100 ? "正文还不够 100 字" : ""),
    comicNewErrorText: actual.comicNewErrorText,
    buildComicInitialBatchSettings: actual.buildComicInitialBatchSettings,
    comicFailureText: actual.comicFailureText,
    comicPresetComicReason: actual.comicPresetComicReason,
    applyComicSettingSelections: actual.applyComicSettingSelections,
    comicPresetAvailability: (options: any[]) => (options.some((option) => option.usable) ? "" : "当前没有可用于漫画的预设，请先在执行器里配置图片模型"),
  comicProgress: (status: any) => ({
    finished: ["completed", "failed", "partial"].includes(status?.state),
    succeeded: status?.counts?.succeeded || 0,
    failed: status?.counts?.failed || 0,
    books: status?.books || [],
  }),
  loadComicPrivateSetup: h.setup,
  resolveComicStyleThumb: h.thumb,
  createComicBatch: h.create,
  uploadComicReference: h.upload,
  submitComicRun: h.submit,
  readComicAutomation: h.readStatus,
  startComicPolling: (_read: any, publish: any, fail: any) => {
    h.pollStarts += 1;
    h.pollPublish = publish;
    h.pollFail = fail;
    if (h.autoFinish) publish({ counts: { total: 1, succeeded: 1, failed: 0 }, ...h.pollStatus });
    return h.pollStop;
  },
  };
});
vi.mock("../../services/comicAssets", async (importOriginal) => {
  const actual = await importOriginal<typeof import("../../services/comicAssets")>();
  return {
    ...actual,
    loadComicAssets: h.assets,
    loadComicAssetImages: h.assetImages,
    loadComicAssetImagePath: h.assetPath,
    extractComicAssets: h.extractAssets,
    generateComicAssetImages: h.generateImages,
    readComicAssetImageTask: h.assetTask,
    retryComicBookStage: h.retryStage,
  };
});
vi.mock("../../services/comicResult", () => ({
  loadComicPages: h.pages,
  loadComicImagePaths: h.loadPaths,
  saveComicImageToAlbum: h.saveAlbum,
  createComicCollection: h.createCollection,
  downloadComicCollectionZip: h.zip,
  buildPreviewRequest: (urls: string[], index: number) => ({ urls, current: urls[index] }),
  setPrimaryComicPageImage: h.setPrimary,
  regenerateComicPage: h.regeneratePage,
}));
vi.mock("../../services/comicEditing", () => ({
  saveComicAsset: h.saveAsset, saveComicPagePrompts: h.savePrompts,
  replaceComicImage: h.replaceImage, setComicAssetPrimary: h.assetPrimary,
}));

import { LegacyComicPrivatePage as ComicPrivatePage } from "./index";
import ComicStageHeader from "./ComicStageHeader";
import ComicPageWorkspace from "./ComicPageWorkspace";
import ComicSettingsStage from "./ComicSettingsStage";
import ComicAssetsStage from "./ComicAssetsStage";
import ComicConfigConfirmStage from "./ComicConfigConfirmStage";
import ComicStickySpacer, { COMIC_STICKY_SPACER_HEIGHT } from "../../components/ComicStickySpacer";

function rawRender() { h.active = h.slots; h.cursor = 0; return ComicPrivatePage() as any; }
function expandSettings(node: any): any {
  if (Array.isArray(node)) return node.map(expandSettings);
  if (!node || typeof node !== "object") return node;
  if ([ComicSettingsStage, ComicAssetsStage, ComicConfigConfirmStage, ComicPageWorkspace, ComicStageHeader].includes(node.type)) {
    const key = node.type.name + ":" + (node.key || "");
    if (!h.childSlots.has(key)) h.childSlots.set(key, []);
    const active = h.active; const cursor = h.cursor;
    h.active = h.childSlots.get(key)!; h.cursor = 0;
    const result = node.type(node.props);
    h.active = active; h.cursor = cursor;
    return expandSettings(result);
  }
  return { ...node, props: { ...node.props, children: expandSettings(node.props?.children) } };
}
function render() { return expandSettings(rawRender()); }
async function flush() { for (let index = 0; index < 40; index += 1) await Promise.resolve(); }
function runEffects() { for (const effect of h.effects.splice(0)) effect(); }
function textContent(node: any): string {
  if (Array.isArray(node)) return node.map(textContent).join("");
  if (node === null || node === undefined || typeof node === "boolean") return "";
  if (typeof node !== "object") return String(node);
  return textContent(node.props?.children);
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
function findSetting(label:string) {
  if((label==="图片描述" || label==="画质与约束") && !findByText(render(),label)) findByText(render(),"提示词").props.onClick();
  return findByText(render(), label === "模型配置" ? "模型" : label === "漫画提示词" ? "提示词" : label);
}
function primaryButton(tree: any) {
  return nodesOfType(tree, "Button").find((node) => node.props?.className === "mp-primary");
}
function typeSource(value: string) {
  const area = nodesOfType(render(), "Textarea")[0];
  area.props.onInput({ detail: { value } });
  return render();
}

const STYLES = [
  { id: "style-1", name: "顶级韩潮风格", previewUrl: "" },
  { id: "style-2", name: "清新日漫", previewUrl: "" },
];
const DUAL_PRESET = {
  preset: {
    id: "p2", name: "双模型",
    config: {
      textModelId: "t1", imageModelId: "i1",
      // automation-presets 实际字段路径：config.aiPromptConfig.comic
      aiPromptConfig: { comic: { enabled: true, body: "分页提示词", assets: { body: "资产提示词" } } },
    },
  },
  name: "双模型", usable: true, reason: "",
};
const TEXT_ONLY_PRESET = {
  preset: { id: "p1", name: "纯文本", config: { textModelId: "t1" } },
  name: "纯文本", usable: false, reason: "这个预设没有配置图片模型",
};
const SETTING_CATALOG = {
  pagePrompts: [{ id: "page-modern", name: "现代条漫分页", slot: "batch.comic-page", version: 2, kind: "", constraintCategory: "" }],
  imageDerivations: [{ id: "derive-cinema", name: "电影感图片推导", slot: "batch.comic-image-derivation", version: 3, kind: "", constraintCategory: "" }],
  constraints: {
    quality: [{ id: "quality-hd", name: "高清细节", slot: "", version: 1, kind: "addon", constraintCategory: "comic_quality" }],
    restriction: [{ id: "restriction-safe", name: "安全构图", slot: "", version: 1, kind: "addon", constraintCategory: "comic_restriction" }],
    negative: [{ id: "negative-clean", name: "无水印无乱码", slot: "", version: 1, kind: "addon", constraintCategory: "comic_negative" }],
  },
};
const MODELS = [
  { id: "text-real", name: "真实文本模型", kind: "text", provider: "personal_api" },
  { id: "image-real", name: "真实图片模型", kind: "image", provider: "doubao" },
];
const SOURCE = `第${"一"}段正文。`.repeat(20);

beforeEach(() => {configureComicRecoveryScope("test-owner");
  h.slots = []; h.active = null; h.childSlots.clear();
  h.effects = [];
  h.seen = new Set();
  h.pollPublish = null;
  h.pollFail = null;
  h.autoFinish = true;
  h.pollStatus = { state: "completed", books: [{ id: "book-1", status: "completed" }] };
  h.pollStarts = 0;
  h.storage.clear();
  h.storage.set("yzcm.mobile.reduceMotion", true);
  h.show = null;
  h.hide = null;
  vi.resetAllMocks();
  h.setup.mockResolvedValue({ styles: STYLES, presets: [DUAL_PRESET], settingCatalog: SETTING_CATALOG, models: MODELS });
  h.thumb.mockResolvedValue("");
  h.create.mockResolvedValue({ batchId: "batch-1", bookId: "book-1", title: "私人定制", bookTitle: "私人定制" });
  h.upload.mockResolvedValue({ assetId: "asset-1", url: "https://cdn/x.png" });
  h.submit.mockResolvedValue({});
  h.pages.mockResolvedValue([
    { pageId: "page-1", videoId: "video-1", ordinal: 1, primaryImageId: "img-1", images: [{ id: "img-1", isPrimary: true }] },
  ]);
  h.loadPaths.mockResolvedValue({ "img-1": "/tmp/page-1.jpg" });
  h.assets.mockResolvedValue([
    { id: "a1", kind: "character", name: "林岚", formalName: "", facts: "", appearance: "短发，穿深色风衣", prompt: "" },
    { id: "a2", kind: "character", name: "沈舟", formalName: "", facts: "", appearance: "高个，戴眼镜", prompt: "" },
    { id: "a3", kind: "scene", name: "雨夜街口", formalName: "", facts: "", appearance: "霓虹反光的路口", prompt: "" },
  ]);
  h.assetImages.mockImplementation(async (_batchId: string, _bookId: string, assetId: string) => (
    assetId === "a1" ? [{ id: "img-a1", url: "/content", mediaType: "image/png", isPrimary: true }] : []
  ));
  // 取图失败（例如白名单/网络问题）时静默返回空串，界面只显示占位。
  h.assetPath.mockResolvedValue("");
  h.extractAssets.mockResolvedValue({});
  h.generateImages.mockResolvedValue({ images: [], taskId: "", warning: "" });
  h.retryStage.mockResolvedValue({});
  h.setPrimary.mockResolvedValue({});
  h.regeneratePage.mockResolvedValue({});
  h.chooseMedia.mockResolvedValue({ tempFiles: [{ tempFilePath: "/tmp/ref.png" }] });
  h.getClipboardData.mockResolvedValue({ data: "" });
  h.showModal.mockResolvedValue({ confirm: true, cancel: false });
  h.showActionSheet.mockResolvedValue({ tapIndex: 0 });
});

describe("旧自动任务兼容页", () => {
  test("资产编辑保存当前选中资产并重新读取当前书籍", async () => {
    h.saveAsset.mockResolvedValue({});
    render();runEffects();await flush();primaryButton(typeSource(SOURCE)).props.onClick();await flush();
    render();runEffects();await flush();
    findByText(render(),"编辑描述").props.onClick();
    nodesOfType(render(),"Textarea")[0].props.onInput({detail:{value:"新的资产名称"}});
    const calls=h.assets.mock.calls.length;
    findByText(render(),"保存资产").props.onClick();await flush();
    expect(h.saveAsset).toHaveBeenCalledWith("batch-1","book-1",expect.objectContaining({id:"a1"}),expect.objectContaining({name:"新的资产名称"}));
    expect(h.assets.mock.calls.length).toBeGreaterThan(calls);
  });

  test("编辑页图提示词使用当前页编号，保存后强制重新读取结果", async () => {
    h.savePrompts.mockResolvedValue({});
    render();runEffects();await flush();primaryButton(typeSource(SOURCE)).props.onClick();await flush();
    findByText(render(),"继续：生成配置确认").props.onClick();findByText(render(),"开始生成漫画").props.onClick();await flush();
    render();runEffects();await flush();
    findByText(render(),"编辑提示词").props.onClick();
    nodesOfType(render(),"Textarea")[0].props.onInput({detail:{value:"新的漫画分页描述"}});
    nodesOfType(render(),"Textarea")[1].props.onInput({detail:{value:"新的图片描述"}});
    const calls=h.pages.mock.calls.length;
    findByText(render(),"保存提示词").props.onClick();await flush();
    expect(h.savePrompts).toHaveBeenCalledWith("batch-1","book-1","video-1","新的漫画分页描述","新的图片描述");
    expect(h.pages.mock.calls.length).toBeGreaterThan(calls);
  });
  test("顶部四步流程是唯一导航，不再重复显示创作和结果标签", async () => {
    render(); runEffects(); await flush();
    const tree = render();
    expect(nodesOfType(tree, "View").some((node) => String(node.props.className || "").split(" ").includes("comic-tabs"))).toBe(false);
    expect(nodesOfType(tree, "View").some((node) => String(node.props.className || "").split(" ").includes("comic-tab"))).toBe(false);
    expect(nodesOfType(rawRender(), ComicStageHeader)).toHaveLength(1);
  });

  test("四项常用创作设置可以直接打开对应选择面板", async () => {
    render(); runEffects(); await flush();
    const titles = [
      ["漫画提示词", "选择漫画提示词"],
      ["图片描述", "选择图片描述"],
      ["模型配置", "选择模型配置"],
      ["画质与约束", "选择画质与约束"],
    ];
    for (const [entry, title] of titles) {
      const card = findSetting(entry);
      expect(card).toBeDefined();
      card.props.onClick();
      expect(textContent(render())).toContain(title);
      findByText(render(), "取消").props.onClick();
    }
  });

  test("四个设置面板各用自己的线上数据，不再重复自动化预设列表", async () => {
    render(); runEffects(); await flush();
    const openCard = (label: string) => {
      const card = findSetting(label);
      card.props.onClick();
      return textContent(render());
    };

    let copy = openCard("漫画提示词");
    expect(copy).toContain("现代条漫分页");
    expect(copy).not.toContain("电影感图片推导");
    expect(copy).not.toContain("双模型");
    findByText(render(), "取消").props.onClick();

    copy = openCard("图片描述");
    expect(copy).toContain("电影感图片推导");
    expect(copy).not.toContain("现代条漫分页");
    findByText(render(), "取消").props.onClick();

    copy = openCard("模型配置");
    expect(copy).toContain("真实文本模型");
    expect(copy).toContain("真实图片模型");
    expect(copy).not.toContain("现代条漫分页");
    findByText(render(), "取消").props.onClick();

    copy = openCard("画质与约束");
    expect(copy).toContain("高清细节");
    expect(copy).toContain("安全构图");
    expect(copy).toContain("无水印无乱码");
    expect(copy).not.toContain("现代条漫分页");
  });

  test("设置阶段使用手机端紧凑布局，继续只创建项目和读取资产", async () => {
    render();
    runEffects();
    await flush();
    let tree = typeSource(SOURCE);
    const findSettings = (node: any): any => Array.isArray(node)
      ? node.map(findSettings).find(Boolean)
      : node && typeof node === "object"
        ? node.type === ComicSettingsStage ? node : findSettings(node.props?.children)
        : undefined;
    const stage = findSettings(rawRender());
    // 抽取后页面通过显式回调连接设置阶段，服务动作仍由页面拥有。
    expect(stage).toBeDefined();
    const copy = textContent(tree);
    expect(copy).toContain("粘贴内容");
    expect(copy).toContain("选择已有书籍");
    expect(copy).toContain("参考图");
    expect(copy).toContain("提示词");
    expect(findByText(tree,"提示词")).toBeDefined();
    expect(copy).toContain("模型");
    expect(findByText(tree,"模型")).toBeDefined();
    expect(copy).not.toContain("上传文件");
    expect(copy).toContain("画幅");
    expect(nodesOfType(tree, "Textarea")[0].props.maxlength).toBe(-1);
    stage.props.onContinue();
    await flush();
    expect(h.create).toHaveBeenCalledTimes(1);
    // 资产改由服务端在 comic_assets 阶段自己提取：创建后不再 POST stages/assets，只读一次列表。
    expect(h.extractAssets).not.toHaveBeenCalled();
    expect(h.assets).toHaveBeenCalledTimes(1);
    expect(h.assets).toHaveBeenCalledWith("batch-1", "book-1");
    expect(h.submit).not.toHaveBeenCalled();
  });

  test("空白正文提供示例但不自动创建或生成，配置完整性独立于字数", async () => {
    render(); runEffects(); await flush();
    let tree = render();
    expect(textContent(tree)).toContain("从一段故事开始");
    expect(textContent(tree)).toContain("创作配置已齐全");
    findByText(tree, "使用示例").props.onClick();
    tree = render();
    expect(nodesOfType(tree, "Textarea")[0].props.value.length).toBeGreaterThanOrEqual(100);
    expect(textContent(tree)).not.toContain("从一段故事开始");
    expect(h.create).not.toHaveBeenCalled(); expect(h.submit).not.toHaveBeenCalled();
    expect(primaryButton(tree).props.disabled).toBe(false);
  });

  test("配置缺失时示例仍能填入，但不会把缺配置误报成可生成", async () => {
    h.setup.mockResolvedValue({styles: STYLES, presets: [], settingCatalog: SETTING_CATALOG, models: MODELS});
    render(); runEffects(); await flush();
    expect(textContent(render())).toContain("创作配置待完善");
    findByText(render(), "使用示例").props.onClick();
    expect(primaryButton(render()).props.disabled).toBe(true);
    expect(h.create).not.toHaveBeenCalled();
  });

  test("手机正文入口可以读取剪贴板，也能进入已有书籍列表", async () => {
    h.getClipboardData.mockResolvedValue({ data: "从剪贴板带入的小说正文" });
    render(); runEffects(); await flush();
    findByText(render(), "粘贴内容").props.onClick();
    await flush();
    expect(nodesOfType(render(), "Textarea")[0].props.value).toBe("从剪贴板带入的小说正文");
    findByText(render(), "选择已有书籍").props.onClick();
    expect(h.navigateTo).toHaveBeenCalledWith({ url: "/comic/pages/comic-config/index?mode=pick" });
    h.storage.set(comicSelectedBookStorageKey()!, { batchId: "old-batch", bookId: "old-book", title: "旧书", sourceText: "从已有书籍带回来的正文" });
    h.show?.();
    await flush();
    expect(nodesOfType(render(), "Textarea")[0].props.value).toBe("从已有书籍带回来的正文");
    expect(h.storage.has(comicSelectedBookStorageKey()!)).toBe(false);
  });

  test("高级设置只切换线上画幅，模型卡只切换对应模型", async () => {
    render(); runEffects(); await flush();
    let tree = typeSource(SOURCE);
    findByText(tree, "画幅").props.onClick();
    tree = render();
    expect(textContent(tree)).toContain("21:9");
    findByText(tree, "3:4").props.onClick();
    findByText(render(), "完成").props.onClick();
    const modelCard = findSetting("模型配置");
    modelCard.props.onClick();
    findByText(render(), "真实图片模型").props.onClick();
    findByText(render(), "使用此配置").props.onClick();
    const selectedModelCard = findSetting("模型配置");
    expect(nodesOfType(rawRender(), ComicSettingsStage)[0].props.settingSelections.imageModelId).toBe("image-real");
    expect(textContent(selectedModelCard)).not.toContain("image-real");
    primaryButton(render()).props.onClick();
    await flush();
    expect(h.create).toHaveBeenCalledWith(expect.objectContaining({
      initialBatchSettings: expect.objectContaining({ imageAspectRatio: "3:4", imageModelId: "image-real", imageProvider: "doubao" }),
    }));
  });

  test("普通动效关闭保留退出阶段，快速重开不会被旧计时器关闭", async () => {
    vi.useFakeTimers();
    try {
      h.storage.set("yzcm.mobile.reduceMotion", false);
      render(); runEffects(); await flush();
      findSetting("漫画提示词").props.onClick();
      findByText(render(), "取消").props.onClick();
      expect(nodesOfType(render(), "View").some(node => String(node.props.className).includes("ce-motion-sheet is-closing"))).toBe(true);
      findSetting("模型配置").props.onClick();
      vi.advanceTimersByTime(300);
      expect(textContent(render())).toContain("选择模型配置");
      findByText(render(), "取消").props.onClick();
      vi.advanceTimersByTime(220);
      expect(textContent(render())).not.toContain("选择模型配置");
    } finally { vi.useRealTimers(); }
  });

  test("选择面板暂存跨分类选择，取消不改变配置，确认一起保存", async () => {
    render(); runEffects(); await flush();
    findSetting("漫画提示词").props.onClick();
    findByText(render(), "现代条漫分页").props.onClick();
    expect(nodesOfType(rawRender(), ComicSettingsStage)[0].props.settingSelections.pagePromptId).toBeUndefined();
    findByText(render(), "取消").props.onClick();
    findSetting("漫画提示词").props.onClick();
    findByText(render(), "现代条漫分页").props.onClick();
    findByText(render(), "图片描述").props.onClick();
    findByText(render(), "电影感图片推导").props.onClick();
    findByText(render(), "使用此提示词").props.onClick();
    expect(nodesOfType(rawRender(), ComicSettingsStage)[0].props.settingSelections).toMatchObject({pagePromptId: "page-modern", imageDerivationId: "derive-cinema"});
  });

  test("四类设置分别写入自己的提交字段，不会整套预设一起替换", async () => {
    render(); runEffects(); await flush();
    typeSource(SOURCE);

    const openSetting = (label: string) => {
      const card = findSetting(label);
      card.props.onClick();
    };
    openSetting("漫画提示词");
    findByText(render(), "现代条漫分页").props.onClick();
    findByText(render(), "使用此提示词").props.onClick();
    openSetting("图片描述");
    findByText(render(), "电影感图片推导").props.onClick();
    findByText(render(), "使用此提示词").props.onClick();
    openSetting("模型配置");
    findByText(render(), "真实文本模型").props.onClick();
    findByText(render(), "真实图片模型").props.onClick();
    findByText(render(), "使用此配置").props.onClick();
    openSetting("画质与约束");
    findByText(render(), "高清细节").props.onClick();
    findByText(render(), "安全构图").props.onClick();
    findByText(render(), "无水印无乱码").props.onClick();
    findByText(render(), "使用此提示词").props.onClick();

    primaryButton(render()).props.onClick();
    await flush();
    const settings = h.create.mock.calls[0][0].initialBatchSettings;
    expect(settings).toMatchObject({
      textModelId: "text-real",
      imageModelId: "image-real",
      imageProvider: "doubao",
      aiPromptConfig: { comic: {
        presetId: "page-modern",
        imageDerivation: { presetId: "derive-cinema" },
      } },
    });
    expect(settings.aiPromptConfig.comic.constraints).toEqual(expect.arrayContaining([
      expect.objectContaining({ presetId: "quality-hd", constraintCategory: "comic_quality" }),
      expect.objectContaining({ presetId: "restriction-safe", constraintCategory: "comic_restriction" }),
      expect.objectContaining({ presetId: "negative-clean", constraintCategory: "comic_negative" }),
    ]));
  });

  test("继续期间停留设置阶段，参考图上传后才进入资产阶段", async () => {
    let resolveCreate: (value: any) => void = () => {};
    h.create.mockReturnValue(new Promise((resolve) => { resolveCreate = resolve; }));
    render();
    runEffects();
    await flush();
    let tree = typeSource(SOURCE);
    findByText(tree, "参考图").props.onClick();
    findByText(render(), "选一张图").props.onClick();
    await flush();
    findByText(render(), "完成").props.onClick();
    primaryButton(render()).props.onClick();
    tree = render();
    expect(textContent(tree)).toContain("正在创建项目…");
    expect(primaryButton(tree).props.disabled).toBe(true);
    expect(textContent(tree)).not.toContain("人物与场景");
    expect(h.upload).not.toHaveBeenCalled();
    resolveCreate({ batchId: "batch-1", bookId: "book-1" });
    await flush();
    expect(h.upload).toHaveBeenCalledWith("batch-1", "/tmp/ref.png");
    // 参考图上传完成后进入资产阶段，只读一次服务端资产，不触发客户端提取。
    expect(h.extractAssets).not.toHaveBeenCalled();
    expect(h.assets).toHaveBeenCalledTimes(1);
    expect(h.assets).toHaveBeenCalledWith("batch-1", "book-1");
    expect(h.submit).not.toHaveBeenCalled();
    expect(textContent(render())).toContain("人物与场景");
  });

  test("入口是漫画生成，三项内容导航保留正确门禁", async () => {
    render();runEffects();await flush();
    const tree=render();expect(nodesOfType(tree,"Header")[0].props.title).toBe("漫画生成");
    const nav=ComicStageHeader(nodesOfType(rawRender(),ComicStageHeader)[0].props);
    expect(nodesOfType(nav,"Button").map(textContent)).toEqual(["正文","人物场景","漫画页图"]);
    expect(nodesOfType(nav,"Button").map(n=>n.props.disabled)).toEqual([false,true,true]);
    expect(textContent(tree)).not.toContain("批次");
  });

  test("字数不足时主按钮禁用并写明原因，点击不发请求", async () => {
    render();
    runEffects();
    await flush();
    let tree = render();
    const button = primaryButton(tree);
    expect(button.props.disabled).toBe(true);
    expect(textContent(tree)).toContain("先粘贴小说正文");
    button.props.onClick();
    expect(h.create).not.toHaveBeenCalled();

    tree = typeSource("只有十个字的样子");
    expect(textContent(tree)).toContain("已输入 8 字，还需 92 字");
    expect(primaryButton(tree).props.disabled).toBe(true);

    tree = typeSource(SOURCE);
    expect(textContent(tree)).toContain("可以开始生成了");
    expect(primaryButton(tree).props.disabled).toBe(false);
  });

  test("只有文本模型的预设被过滤掉，主按钮禁用并提示先配图片模型", async () => {
    h.setup.mockResolvedValue({ styles: STYLES, presets: [TEXT_ONLY_PRESET] });
    let tree = render();
    runEffects();
    await flush();
    tree = typeSource(SOURCE);
    expect(textContent(tree)).toContain("当前没有可用于漫画的预设，请先在执行器里配置图片模型");
    expect(primaryButton(tree).props.disabled).toBe(true);
    primaryButton(tree).props.onClick();
    expect(h.create).not.toHaveBeenCalled();
  });

  test("预设接口 401 时禁用原因是中文兜底，界面上不出现英文原文", async () => {
    h.setup.mockRejectedValue({ statusCode: 401, message: "unauthorized" });
    render();
    runEffects();
    await flush();
    const tree = render();
    const text = textContent(tree);
    expect(text).toContain("登录已失效，请重新登录");
    expect(text).not.toMatch(/unauthorized/i);
    expect(text).not.toContain("401");
    expect(primaryButton(tree).props.disabled).toBe(true);
  });

  test("双模型预设先创建项目并读取资产，确认配置后才启动漫画", async () => {
    let tree = render();
    runEffects();
    await flush();
    tree = typeSource(SOURCE);
    primaryButton(tree).props.onClick();
    await flush();
    expect(h.create).toHaveBeenCalledWith(expect.objectContaining({ sourceText: SOURCE }));
    // 提取由服务端在 comic_assets 阶段做：客户端只读列表，不再 POST stages/assets。
    expect(h.extractAssets).not.toHaveBeenCalled();
    expect(h.assets).toHaveBeenCalledTimes(1);
    expect(h.assets).toHaveBeenCalledWith("batch-1", "book-1");
    expect(h.submit).not.toHaveBeenCalled();
    expect(textContent(render())).toContain("林岚");
    findByText(render(), "继续：生成配置确认").props.onClick();
    await flush();
    findByText(render(), "开始生成漫画").props.onClick();
    await flush();
    expect(h.submit).toHaveBeenCalledWith(expect.objectContaining({
      batchId: "batch-1", bookIds: ["book-1"], presetId: "p2", runMode: "video_no_submit",
    }), expect.objectContaining({ id: "p2" }));
    const after = render();
    expect(textContent(after)).toContain("正在读取生成状态");
    expect(h.navigateTo).not.toHaveBeenCalled();
  });

  test("生成配置确认显示真实只读设置、参考状态与全部资产数量", async () => {
    render(); runEffects(); await flush();
    let tree = typeSource(SOURCE);
    findByText(tree, "参考图").props.onClick();
    findByText(render(), "选一张图").props.onClick(); await flush();
    findByText(render(), "完成").props.onClick();
    primaryButton(render()).props.onClick(); await flush();
    findByText(render(), "继续：生成配置确认").props.onClick();
    tree = render();
    const text = textContent(tree);
    for (const expected of ["生成配置确认", `正文 ${SOURCE.trim().length} 字`, "顶级韩潮风格", "9:16", "参考图已上传", "人物 2", "场景 1", "道具 0", "林岚", "沈舟", "雨夜街口", "逐页编辑提示词"]) expect(text).toContain(expected);
    expect(nodesOfType(tree, "Input")).toHaveLength(0);
    expect(nodesOfType(tree, "Textarea")).toHaveLength(0);
    expect(text).not.toContain("不满意只改这一页");
    expect(text).not.toContain("看分镜");
    expect(h.submit).not.toHaveBeenCalled();
    findByText(tree, "返回资产工作台").props.onClick();
    expect(textContent(render())).toContain("人物与场景");
    expect(h.submit).not.toHaveBeenCalled();
  });

  test("确认配置同步双击只提交一次，成功前停留确认页且保留上传参考元数据", async () => {
    let resolveSubmit: (value: any) => void = () => {};
    h.submit.mockReturnValue(new Promise((resolve) => { resolveSubmit = resolve; }));
    render(); runEffects(); await flush();
    findByText(typeSource(SOURCE), "参考图").props.onClick();
    findByText(render(), "选一张图").props.onClick(); await flush();
    findByText(render(), "完成").props.onClick();
    primaryButton(render()).props.onClick(); await flush();
    findByText(render(), "继续：生成配置确认").props.onClick();
    const confirm = findByText(render(), "开始生成漫画");
    confirm.props.onClick(); confirm.props.onClick();
    await flush();
    expect(h.submit).toHaveBeenCalledTimes(1);
    expect(h.submit).toHaveBeenCalledWith(expect.objectContaining({
      aspectRatio: "9:16", stylePresetId: "style-1", reference: { assetId: "asset-1", url: "https://cdn/x.png" },
    }), expect.objectContaining({ id: "p2" }));
    let tree = render();
    expect(textContent(tree)).toContain("生成配置确认");
    expect(textContent(tree)).not.toContain("正在读取生成状态");
    expect(primaryButton(tree).props.disabled).toBe(true);
    findByText(tree, "返回资产工作台").props.onClick();
    expect(textContent(render())).toContain("开始生成漫画");
    resolveSubmit({}); await flush();
    expect(textContent(render())).toContain("正在读取生成状态");
    expect(h.create).toHaveBeenCalledTimes(1);
    expect(h.upload).toHaveBeenCalledTimes(1);
  });

  test.each([true, false])("创建项目 A 后修改草稿 B，确认与提交仍使用 A（A 参考图：%s）", async (withReference) => {
    const uploadedReference = { assetId: "reference-a", url: "https://cdn/a.png", previewUrl: "https://cdn/a-preview.png" };
    h.upload.mockResolvedValue(uploadedReference);
    render(); runEffects(); await flush();
    typeSource(SOURCE);
    if (withReference) {
      findByText(render(), "参考图").props.onClick();
      findByText(render(), "选一张图").props.onClick(); await flush();
      findByText(render(), "完成").props.onClick();
    }
    primaryButton(render()).props.onClick(); await flush();
    findByText(render(), "正文").props.onClick();
    const sourceB = "另一个故事的正文。".repeat(30);
    let tree = typeSource(sourceB);
    findByText(tree, "画风").props.onClick();
    findByText(render(), "清新日漫").props.onClick();
    findByText(render(), "完成").props.onClick();
    findByText(render(), "参考图").props.onClick();
    findByText(render(), withReference ? "去掉" : "选一张图").props.onClick(); await flush();
    findByText(render(), "完成").props.onClick();
    tree = render();
    expect(nodesOfType(tree, "Textarea")[0].props.value).toBe(sourceB);
    expect(nodesOfType(rawRender(), ComicSettingsStage)[0].props.styleId).toBe("style-2");
    expect(Boolean(nodesOfType(rawRender(), ComicSettingsStage)[0].props.referencePath)).toBe(!withReference);
    nodesOfType(rawRender(), ComicStageHeader)[0].props.onSelectStage("assets");
    findByText(render(), "补齐资产图").props.onClick(); await flush();
    expect(h.generateImages).toHaveBeenCalledWith(expect.objectContaining({ batchId: "batch-1", imageAspectRatio: "9:16" }));
    findByText(render(), "继续：生成配置确认").props.onClick();
    tree = render();
    const summary = textContent(nodesOfType(tree, "View").find((node) => node.props.className === "comic-config-summary"));
    expect(summary).toContain(`正文 ${SOURCE.trim().length} 字`);
    expect(summary).toContain("顶级韩潮风格");
    expect(summary).toContain("9:16");
    expect(summary).toContain(withReference ? "参考图已上传" : "未使用参考图");
    expect(summary).not.toContain(`正文 ${sourceB.length} 字`);
    expect(summary).not.toContain("清新日漫");
    expect(summary).not.toContain("16:9");
    findByText(tree, "开始生成漫画").props.onClick(); await flush();
    expect(h.submit).toHaveBeenCalledWith(expect.objectContaining({
      batchId: "batch-1", bookIds: ["book-1"], aspectRatio: "9:16", stylePresetId: "style-1",
      reference: withReference ? { assetId: "reference-a", url: "https://cdn/a.png", previewUrl: "https://cdn/a-preview.png" } : null,
    }), expect.objectContaining({ id: "p2" }));
    expect(h.create).toHaveBeenCalledTimes(1);
    expect(h.upload).toHaveBeenCalledTimes(withReference ? 1 : 0);
  });

  test("创建项目后再改模型草稿，当前项目提交仍使用创建时的模型", async () => {
    h.setup.mockResolvedValue({
      styles: STYLES,
      presets: [DUAL_PRESET],
      settingCatalog: SETTING_CATALOG,
      models: [...MODELS, { id: "image-next", name: "下一本图片模型", kind: "image", provider: "next-provider" }],
    });
    render(); runEffects(); await flush();
    typeSource(SOURCE);
    const openModel = () => findSetting("模型配置")!.props.onClick();
    openModel();
    findByText(render(), "真实图片模型").props.onClick();
    findByText(render(), "使用此配置").props.onClick();
    primaryButton(render()).props.onClick(); await flush();

    findByText(render(), "正文").props.onClick();
    openModel();
    findByText(render(), "下一本图片模型").props.onClick();
    findByText(render(), "取消").props.onClick();
    nodesOfType(rawRender(), ComicStageHeader)[0].props.onSelectStage("assets");
    findByText(render(), "继续：生成配置确认").props.onClick();
    findByText(render(), "开始生成漫画").props.onClick(); await flush();

    expect(h.submit).toHaveBeenCalledWith(expect.anything(), expect.objectContaining({
      config: expect.objectContaining({ imageModelId: "image-real", imageProvider: "doubao" }),
    }));
  });

  test("配置提交失败保留确认摘要并允许再次确认，未选参考图准确展示", async () => {
    h.submit.mockRejectedValueOnce(Object.assign(new Error("请求失败：500"), { statusCode: 500 }));
    render(); runEffects(); await flush();
    primaryButton(typeSource(SOURCE)).props.onClick(); await flush();
    findByText(render(), "继续：生成配置确认").props.onClick();
    expect(textContent(render())).toContain("未使用参考图");
    findByText(render(), "开始生成漫画").props.onClick(); await flush();
    let tree = render();
    expect(textContent(tree)).toContain("服务端暂时忙不过来，稍后再试一次");
    expect(textContent(tree)).not.toContain("正在读取生成状态");
    expect(primaryButton(tree).props.disabled).toBe(false);
    primaryButton(tree).props.onClick(); await flush();
    expect(h.submit).toHaveBeenCalledTimes(2);
  });

  test("正式任务提交后从 rail 返回配置只能查看进度，旧确认回调也不能重复提交", async () => {
    h.autoFinish = false;
    render(); runEffects(); await flush();
    primaryButton(typeSource(SOURCE)).props.onClick(); await flush();
    findByText(render(), "继续：生成配置确认").props.onClick();
    const originalConfirm = primaryButton(render()).props.onClick;
    originalConfirm(); await flush();
    const header = nodesOfType(rawRender(), ComicStageHeader)[0];
    header.props.onSelectStage("storyboard");
    primaryButton(render()).props.onClick(); await flush();
    originalConfirm(); await flush();
    expect(h.submit).toHaveBeenCalledTimes(1);
    expect(textContent(render())).toContain("正在读取生成状态");
  });

  test.each(["load", "thumbnail", "load-error"])("新项目创建后忽略旧项目 %s 回调，不覆盖新项目资产状态", async (operation) => {
    let resolveOld: (value: any) => void = () => {};
    let rejectOld: (error: Error) => void = () => {};
    const oldRequest = new Promise((resolve, reject) => { resolveOld = resolve; rejectOld = reject; });
    // 旧项目的资产读取挂在半路：load 挂在列表请求上，thumbnail 挂在首张定妆图上。
    if (operation === "thumbnail") h.assetPath.mockReturnValueOnce(oldRequest);
    else h.assets.mockReturnValueOnce(oldRequest);
    render(); runEffects(); await flush();
    primaryButton(typeSource(SOURCE)).props.onClick(); await flush();
    expect(h.slots[0]).toMatchObject({ batchId: "batch-1", assets: [], assetsLoading: true });
    nodesOfType(rawRender(), ComicStageHeader)[0].props.onSelectStage("source");
    h.create.mockResolvedValueOnce({ batchId: "batch-2", bookId: "book-2" });
    primaryButton(render()).props.onClick(); await flush();
    expect(h.slots[0]).toMatchObject({ batchId: "batch-2", bookId: "book-2", assetsLoading: false, assetError: "" });
    const assetReadsBefore = h.assets.mock.calls.length;
    if (operation === "load-error") rejectOld(new Error("旧项目错误"));
    else resolveOld(operation === "thumbnail" ? "/tmp/old-asset.png" : [{ id: "old-asset", kind: "character", name: "旧角色" }]);
    await flush();
    const state = h.slots[0];
    expect(state).toMatchObject({ batchId: "batch-2", bookId: "book-2", assetsLoading: false, assetBusy: "", assetError: "" });
    expect(state.assets.map((asset: any) => asset.id)).toEqual(["a1", "a2", "a3"]);
    expect(h.assets.mock.calls.length).toBe(assetReadsBefore);
    expect(textContent(render())).not.toContain("旧角色");
  });

  test.each(["extract", "images", "retry"])("旧项目手动 %s 操作结束不能影响新项目的资产读取状态", async (operation) => {
    render(); runEffects(); await flush();
    primaryButton(typeSource(SOURCE)).props.onClick(); await flush();
    let rejectOld: (error: Error) => void = () => {};
    const oldRequest = new Promise((_, reject) => { rejectOld = reject; });
    const action = operation === "extract" ? h.extractAssets : operation === "images" ? h.generateImages : h.retryStage;
    action.mockReturnValueOnce(oldRequest);
    if(operation !== "images") findByText(render(), "更多").props.onClick();
    findByText(render(), operation === "extract" ? "重新提取资产" : operation === "images" ? "补齐资产图" : "重试失败阶段").props.onClick();
    expect(h.slots[0].assetBusy).toBe(operation);
    // 切回创作再建一个新项目；新项目的资产读取保持挂起，这就是新的「繁忙」状态。
    nodesOfType(rawRender(), ComicStageHeader)[0].props.onSelectStage("source");
    h.create.mockResolvedValueOnce({ batchId: "batch-2", bookId: "book-2" });
    h.assets.mockReturnValueOnce(new Promise(() => {}));
    primaryButton(render()).props.onClick(); await flush();
    expect(h.slots[0]).toMatchObject({ batchId: "batch-2", assets: [], assetsLoading: true, assetBusy: "", assetError: "" });
    rejectOld(new Error("旧操作失败")); await flush();
    // 旧项目的失败回调既不能把错误写进新项目，也不能清掉新项目仍在进行的资产读取。
    expect(h.slots[0]).toMatchObject({ batchId: "batch-2", assets: [], assetsLoading: true, assetBusy: "", assetError: "" });
  });

  test.each(["startAnother", "unmount"])("旧资产读取在 %s 边界后不能回写", async (boundary) => {
    render(); runEffects(); await flush();
    primaryButton(typeSource(SOURCE)).props.onClick(); await flush();
    findByText(render(), "继续：生成配置确认").props.onClick();
    primaryButton(render()).props.onClick(); await flush();
    let resolveOld: (value: any) => void = () => {};
    h.assets.mockReturnValueOnce(new Promise((resolve) => { resolveOld = resolve; }));
    render(); runEffects(); await flush();
    if (boundary === "startAnother") findByText(render(), "再做一本").props.onClick();
    else for (const entry of h.slots) entry?.cleanup?.();
    const stateBefore = h.slots[0];
    resolveOld([{ id: "old-asset", kind: "character", name: "旧角色" }]); await flush();
    expect(h.slots[0]).toEqual(stateBefore);
    expect(textContent(render())).not.toContain("旧角色");
  });

  test("参考图是一行入口，点开弹层后才选图", async () => {
    let tree = render();
    runEffects();
    await flush();
    expect(findByText(tree,"参考图")).toBeDefined();
    findByText(tree, "参考图").props.onClick();
    tree = render();
    expect(textContent(tree)).toContain("选一张图");
    findByText(tree, "选一张图").props.onClick();
    await flush();
    tree = render();
    expect(h.chooseMedia).toHaveBeenCalled();
    expect(nodesOfType(rawRender(),ComicSettingsStage)[0].props.referencePath).toBeTruthy();
  });

  test("结果在同页内联：轮询完成后直接看页图、保存和打包", async () => {
    let tree = render();
    runEffects();
    await flush();
    tree = typeSource(SOURCE);
    primaryButton(tree).props.onClick();
    await flush();
    findByText(render(), "继续：生成配置确认").props.onClick();
    await flush();
    findByText(render(), "开始生成漫画").props.onClick();
    await flush();
    tree = render();
    runEffects();
    await flush();
    tree = render();
    expect(textContent(tree)).toContain("本次生成已结束");
    expect(textContent(tree)).toContain("已读取 1 页");
    expect(findByText(tree, "导出漫画")).toBeDefined();
    expect(textContent(tree)).toContain("保存到相册");
    expect(textContent(tree)).toContain("导出");
    expect(findByText(tree, "再做一本")).toBeDefined();
    expect(h.navigateTo).not.toHaveBeenCalled();
  });

  test("结果显示真实页图数量并能定位无图失败页", async () => {
    h.pages.mockResolvedValue([
      {pageId:"page-1",videoId:"video-1",ordinal:1,status:"completed",primaryImageId:"img-1",images:[{id:"img-1",isPrimary:true}]},
      {pageId:"page-2",videoId:"video-2",ordinal:2,status:"failed",primaryImageId:"",images:[]},
      {pageId:"page-3",videoId:"video-3",ordinal:3,status:"running",primaryImageId:"",images:[]},
    ]);
    render(); runEffects(); await flush();
    primaryButton(typeSource(SOURCE)).props.onClick(); await flush();
    findByText(render(), "继续：生成配置确认").props.onClick();
    findByText(render(), "开始生成漫画").props.onClick(); await flush();
    render(); runEffects(); await flush();
    const tree=render(); expect(textContent(tree)).toContain("有主图 1 页 · 无图失败 1 页 · 待出图 1 页");
    findByText(tree,"查看失败页").props.onClick(); await flush();
    expect(h.slots[0].pageIndex).toBe(1);
  });

  test("手机端默认竖版画幅贯穿资产图与正式提交", async () => {
    const aspectRatio = "9:16";
    render(); runEffects(); await flush();
    typeSource(SOURCE);
    primaryButton(render()).props.onClick(); await flush();
    findByText(render(), "补齐资产图").props.onClick(); await flush();
    expect(h.generateImages).toHaveBeenCalledWith(expect.objectContaining({ batchId: "batch-1", bookId: "book-1", imageAspectRatio: aspectRatio }));
    findByText(render(), "继续：生成配置确认").props.onClick();
    findByText(render(), "开始生成漫画").props.onClick(); await flush();
    expect(h.submit).toHaveBeenCalledTimes(1);
    expect(h.submit).toHaveBeenCalledWith(expect.objectContaining({ batchId: "batch-1", bookIds: ["book-1"], aspectRatio }), expect.anything());
  });

  test("画风缩略图先出骨架，失败给重试，点击后重新请求", async () => {
    h.setup.mockResolvedValue({
      styles: [{ id: "style-1", name: "顶级韩潮风格", previewUrl: "https://cdn/style-1.png" }],
      presets: [DUAL_PRESET],
    });
    let resolveThumb: (value: string) => void = () => {};
    h.thumb.mockReturnValue(new Promise((resolve) => { resolveThumb = resolve as (value: string) => void; }));
    render(); runEffects(); await flush();
    findByText(render(), "画风").props.onClick();
    // 画风面板打开后，未读取的缩略图有明确骨架。
    expect(nodesOfType(render(), "View").some((node) => node.props.className === "comic-style-thumb is-skeleton")).toBe(true);

    resolveThumb(""); await flush();
    let tree = render();
    expect(nodesOfType(tree, "View").some((node) => node.props.className === "comic-style-thumb is-skeleton")).toBe(false);
    const retry = nodesOfType(tree, "View").find((node) => node.props.className === "comic-style-thumb is-failed");
    expect(retry).toBeDefined();

    h.thumb.mockResolvedValue("/tmp/style-1.jpg");
    retry!.props.onClick({ stopPropagation: () => undefined });
    await flush();
    tree = render();
    expect(nodesOfType(tree, "Image").some((node) => node.props.src === "/tmp/style-1.jpg")).toBe(true);
    expect(h.thumb).toHaveBeenCalledTimes(2);
  });

  test("建批后只读一次资产列表，切页 effect 复用这次在途请求，服务端提取期间给出提示", async () => {
    let releaseAssets: (value: any) => void = () => {};
    h.assets.mockReturnValue(new Promise((resolve) => { releaseAssets = resolve; }));
    // 服务端还在 comic_assets 阶段提取：列表为空是正常的，界面要给出「正在提取」而不是空状态。
    h.pollStatus = { state: "running", books: [{ id: "book-1", stage: "comic_assets" }] };
    render(); runEffects(); await flush();
    primaryButton(typeSource(SOURCE)).props.onClick();
    await flush();
    // 创建后只主动读一次资产，也没有客户端 POST stages/assets。
    expect(h.extractAssets).not.toHaveBeenCalled();
    expect(h.assets).toHaveBeenCalledTimes(1);
    expect(h.assets).toHaveBeenCalledWith("batch-1", "book-1");
    expect(textContent(render())).toContain("正在读取资产…");
    // 切页 effect 会再触发一次非 force 加载；同一批次在途时必须复用，不再空转。
    runEffects(); await flush();
    expect(h.assets).toHaveBeenCalledTimes(1);
    const tree = render();
    expect(textContent(tree)).toContain("正在提取角色与场景");
    expect(textContent(tree)).not.toContain("还没有提取到资产");
    releaseAssets([]); await flush();
    expect(h.assets).toHaveBeenCalledTimes(1);
    expect(h.extractAssets).not.toHaveBeenCalled();
  });

  test("结果保存与整本下载复用当前书籍和页图作用域", async () => {
    h.createCollection.mockResolvedValue({ id: "collection-1", items: [{}] });
    h.zip.mockResolvedValue({ url: "https://cdn/book.zip" });
    render(); runEffects(); await flush();
    primaryButton(typeSource(SOURCE)).props.onClick(); await flush();
    findByText(render(), "继续：生成配置确认").props.onClick();
    findByText(render(), "开始生成漫画").props.onClick(); await flush();
    render(); runEffects(); await flush();
    findByText(render(), "保存到相册").props.onClick(); await flush();
    expect(h.saveAlbum).toHaveBeenCalledWith("/tmp/page-1.jpg");
    findByText(render(), "导出").props.onClick(); await flush();
    expect(h.createCollection).toHaveBeenCalledWith("batch-1", "book-1", ["video-1"]);
    expect(h.zip).toHaveBeenCalledWith("batch-1", "book-1", "collection-1");
  });

  test("结果页可以把候选图设为主图并重新生成当前页", async () => {
    h.pages.mockResolvedValue([
      { pageId: "page-1", videoId: "video-1", ordinal: 1, status: "ready", aspectRatio: "9:16", comicPrompt: "导演描述", imagePrompt: "出图描述", primaryImageId: "img-1", images: [{ id: "img-1", isPrimary: true }, { id: "img-2", isPrimary: false }] },
    ]);
    h.loadPaths.mockResolvedValue({ "img-1": "/tmp/page-1.jpg", "img-2": "/tmp/page-2.jpg" });
    render(); runEffects(); await flush();
    primaryButton(typeSource(SOURCE)).props.onClick(); await flush();
    findByText(render(), "继续：生成配置确认").props.onClick();
    findByText(render(), "开始生成漫画").props.onClick(); await flush();
    render(); runEffects(); await flush();
    findByText(render(), "图片版本").props.onClick();
    findByText(render(), "设为主图").props.onClick({ stopPropagation: () => undefined });
    await flush();
    expect(h.setPrimary).toHaveBeenCalledWith("batch-1", "book-1", "page-1", "img-2");
    findByText(render(), "关闭").props.onClick();
    findByText(render(), "重新生成").props.onClick();
    await flush();
    expect(h.regeneratePage).toHaveBeenCalledWith("batch-1", "book-1", "video-1", expect.objectContaining({
      comicPrompt: "导演描述", imagePrompt: "出图描述", modelId: "i1", aspectRatio: "9:16",
    }));
  });

  test("旧项目预览取图晚于重新创作返回时不会回写或打开旧图", async () => {
    render(); runEffects(); await flush();
    primaryButton(typeSource(SOURCE)).props.onClick(); await flush();
    findByText(render(), "继续：生成配置确认").props.onClick();
    findByText(render(), "开始生成漫画").props.onClick(); await flush();
    render(); runEffects(); await flush();
    let resolvePaths: (paths: any) => void = () => {};
    h.loadPaths.mockReturnValue(new Promise((resolve) => { resolvePaths = resolve; }));
    nodesOfType(render(), "Image").find((node) => node.props.className === "ce-page-main-image")!.props.onClick({ stopPropagation: () => {} });
    findByText(render(), "再做一本").props.onClick();
    resolvePaths({ "img-1": "/tmp/stale-preview.jpg" }); await flush();
    expect(h.slots[0].paths).toEqual({});
    expect(h.previewImage).not.toHaveBeenCalled();
  });

  test("内容导航在建项目后开放资产，提交前页图入口不伪造结果", async () => {
    render();runEffects();await flush();
    const nav=()=>ComicStageHeader(nodesOfType(rawRender(),ComicStageHeader)[0].props);
    expect(nodesOfType(nav(),"Button").map(n=>n.props.disabled)).toEqual([false,true,true]);
    primaryButton(typeSource(SOURCE)).props.onClick();await flush();
    expect(nodesOfType(nav(),"Button").map(n=>n.props.disabled)).toEqual([false,false,true]);
    nodesOfType(rawRender(),ComicStageHeader)[0].props.onSelectStage("storyboard");
    expect(textContent(render())).toContain("开始生成漫画");expect(h.submit).not.toHaveBeenCalled();
  });

  test("没有页图的终态禁用导出，完成后不会重启轮询", async () => {
    h.pages.mockResolvedValue([]);render();runEffects();await flush();
    primaryButton(typeSource(SOURCE)).props.onClick();await flush();
    findByText(render(),"继续：生成配置确认").props.onClick();
    findByText(render(),"开始生成漫画").props.onClick();await flush();
    render();runEffects();await flush();render();runEffects();await flush();
    expect(findByText(render(),"导出").props.disabled).toBe(true);
    expect(h.pollStarts).toBe(1);expect(h.pollStop).toHaveBeenCalled();expect(h.pages).toHaveBeenCalledWith("batch-1","book-1");
  });

  test("终态页图读取失败可以只重读结果，不重启轮询或正式任务", async () => {
    h.pages.mockRejectedValueOnce(new Error("页图暂未落库"));
    render(); runEffects(); await flush();
    primaryButton(typeSource(SOURCE)).props.onClick(); await flush();
    findByText(render(), "继续：生成配置确认").props.onClick();
    findByText(render(), "开始生成漫画").props.onClick(); await flush();
    render(); runEffects(); await flush();
    expect(textContent(render())).toContain("页图暂未落库");
    findByText(render(), "重新读取页图").props.onClick(); await flush();
    expect(textContent(render())).toContain("保存到相册");
    expect(h.pages).toHaveBeenCalledTimes(2);
    expect(h.pollStarts).toBe(1);
    expect(h.submit).toHaveBeenCalledTimes(1);
  });

  test("分页取图晚于重新创作返回时，旧任务图片不能回写新项目", async () => {
    let resolvePaths: (paths: any) => void = () => {};
    h.loadPaths.mockReturnValue(new Promise((resolve) => { resolvePaths = resolve; }));
    render(); runEffects(); await flush();
    primaryButton(typeSource(SOURCE)).props.onClick(); await flush();
    findByText(render(), "继续：生成配置确认").props.onClick();
    findByText(render(), "开始生成漫画").props.onClick(); await flush();
    render(); runEffects(); await flush();
    findByText(render(), "再做一本").props.onClick();
    resolvePaths({ "img-1": "/tmp/stale-page.jpg" }); await flush();
    expect(h.slots[0].paths).toEqual({});
    expect(nodesOfType(render(), "Image").some((node) => node.props.src === "/tmp/stale-page.jpg")).toBe(false);
  });

  test.each(["stage", "error", "unmount"])("轮询在 %s 边界停止且忽略旧回调", async (boundary) => {
    h.autoFinish = false;
    render(); runEffects(); await flush();
    primaryButton(typeSource(SOURCE)).props.onClick(); await flush();
    findByText(render(), "继续：生成配置确认").props.onClick();
    findByText(render(), "开始生成漫画").props.onClick(); await flush();
    render(); runEffects(); await flush();
    const publish = h.pollPublish;
    if (boundary === "stage") {
      findByText(render(), "人物场景").props.onClick();
      render(); runEffects();
    } else if (boundary === "error") {
      h.pollFail(new Error("read failed"));
    } else {
      for (const entry of h.slots) entry?.cleanup?.();
    }
    expect(h.pollStop).toHaveBeenCalled();
    publish({ state: "completed", counts: { succeeded: 1 } }); await flush();
    expect(h.pages).not.toHaveBeenCalled();
  });

  test("资产列表可分类、搜索并打开当前资产编辑表单", async () => {
    render();runEffects();await flush();primaryButton(typeSource(SOURCE)).props.onClick();await flush();
    let tree=render();expect(textContent(tree)).toContain("人物 2");expect(textContent(tree)).toContain("场景 1");
    expect(textContent(tree)).toContain("林岚");
    findByText(tree,"编辑描述").props.onClick();
    expect(textContent(render())).toContain("编辑资产");expect(nodesOfType(render(),"Textarea").length).toBe(5);
    findByText(render(),"关闭").props.onClick();
    nodesOfType(render(),"Input")[0].props.onInput({detail:{value:"眼镜"}});
    expect(textContent(render())).toContain("沈舟");expect(textContent(render())).not.toContain("林岚");
    nodesOfType(render(),"Input")[0].props.onInput({detail:{value:""}});
    findByText(render(),"场景 1").props.onClick();expect(textContent(render())).toContain("雨夜街口");
    expect(h.submit).not.toHaveBeenCalled();
  });

  test("资产为空也不再拦继续，刷新与重试保留为手动兜底", async () => {
    h.assets.mockResolvedValue([]);
    render(); runEffects(); await flush();
    primaryButton(typeSource(SOURCE)).props.onClick(); await flush();
    let tree = render();
    const next = findByText(tree, "继续：生成配置确认");
    expect(next.props.disabled).toBe(false);
    next.props.onClick();
    tree = render();
    expect(textContent(tree)).toContain("生成配置确认");
    expect(textContent(tree)).toContain("当前未读取到人物或场景");
    // 确认按钮也不再因资产为空而禁用：门禁在客户端被移除，改由服务端保证。
    expect(findByText(tree, "开始生成漫画").props.disabled).toBe(false);
    findByText(tree, "返回资产工作台").props.onClick();
    const previousLoads = h.assets.mock.calls.length;
    findByText(render(), "刷新").props.onClick(); await flush();
    expect(h.assets.mock.calls.length).toBeGreaterThan(previousLoads);
    findByText(render(), "更多").props.onClick();
    findByText(render(), "重试失败阶段").props.onClick(); await flush();
    expect(h.retryStage).toHaveBeenCalledWith({ batchId: "batch-1", bookId: "book-1", textModelId: "t1" });
    expect(h.submit).not.toHaveBeenCalled();
  });

  test("补齐资产图只请求缺图资产，重新提取走整本强制重提", async () => {
    let tree = render();
    runEffects();
    await flush();
    tree = typeSource(SOURCE);
    primaryButton(tree).props.onClick();
    await flush();
    tree = render();
    runEffects();
    await flush();
    tree = render();

    findByText(tree, "补齐资产图（缺 2）").props.onClick();
    await flush();
    expect(h.generateImages).toHaveBeenCalledWith(expect.objectContaining({
      assetIds: ["a2", "a3"], modelId: "i1", imageAspectRatio: "9:16",
    }));

    tree = render();
    findByText(tree, "更多").props.onClick();
    findByText(render(), "重新提取资产").props.onClick();
    await flush();
    expect(h.extractAssets).toHaveBeenCalledWith(expect.objectContaining({
      mode: "force", textModelId: "t1",
    }));
  });

  test("资产缩略图渲染失败后刷新相同缓存路径可以恢复", async () => {
    h.assetPath.mockResolvedValue("/tmp/asset-same.png");
    render(); runEffects(); await flush();
    primaryButton(typeSource(SOURCE)).props.onClick(); await flush();
    let tree = render();
    const thumbnail = nodesOfType(tree, "Image").find((node) => node.props.src === "/tmp/asset-same.png");
    expect(thumbnail).toBeDefined();
    thumbnail.props.onError();
    tree = render();
    expect(nodesOfType(tree, "Image").some((node) => node.props.src === "/tmp/asset-same.png")).toBe(false);
    expect(textContent(tree)).toContain("图片暂不可用");
    findByText(tree, "刷新").props.onClick(); await flush();
    tree = render();
    expect(nodesOfType(tree, "Image").some((node) => node.props.src === "/tmp/asset-same.png")).toBe(true);
    expect(textContent(tree)).not.toContain("图片暂不可用");
  });

  test("单张资产媒体失败不阻断其余资产，手动动作繁忙时禁用但继续仍可点", async () => {
    h.assetPath.mockRejectedValue(new Error("single image unavailable"));
    render(); runEffects(); await flush();
    primaryButton(typeSource(SOURCE)).props.onClick(); await flush();
    let tree = render();
    expect(textContent(tree)).toContain("林岚");
    expect(textContent(tree)).toContain("沈舟");
    expect(textContent(tree)).toContain("图片暂不可用");
    expect(textContent(tree)).not.toContain("single image unavailable");
    h.generateImages.mockReturnValue(new Promise(() => {}));
    findByText(tree, "补齐资产图").props.onClick();
    tree = render();
    // 手动兜底入口在动作进行中仍禁用，避免重复请求；但它们不是必经步骤。
    for (const label of ["补齐资产图", "编辑描述", "生成图片", "图片版本"]) {
      expect(nodesOfType(tree,"Button").find(n=>textContent(n).startsWith(label))?.props.disabled).toBe(true);
    }
    expect(h.generateImages).toHaveBeenCalledTimes(1);
    // 门禁去掉后，动作进行中也可以直接继续确认配置。
    const next = findByText(tree, "继续：生成配置确认");
    expect(next.props.disabled).toBe(false);
    next.props.onClick();
    expect(textContent(render())).toContain("生成配置确认");
    // 建批不再触发客户端提取，因此被禁用的「重新提取资产」不会发出任何请求。
    expect(h.extractAssets).not.toHaveBeenCalled();
    expect(h.retryStage).not.toHaveBeenCalled();
    expect(h.submit).not.toHaveBeenCalled();
  });

  test("预设缺 comic.assets 引用时创作页直接说明原因并禁用主按钮", async () => {
    h.setup.mockResolvedValue({
      styles: STYLES,
      presets: [{
        preset: { id: "p3", name: "缺资产提示词", config: { textModelId: "t1", imageModelId: "i1", aiPromptConfig: { comic: { enabled: true, body: "分页" } } } },
        name: "缺资产提示词", usable: true, reason: "",
      }],
    });
    let tree = render();
    runEffects();
    await flush();
    tree = typeSource(SOURCE);
    expect(textContent(tree)).toContain("这个预设没有选择漫画资产提示词");
    expect(primaryButton(tree).props.disabled).toBe(true);
  });

  test("默认用第一条漫画可用的预设，comic=null 的第一条不会顶上去", async () => {
    const comicPreset = {
      preset: { id: "p1", name: "漫画", config: { textModelId: "t1", imageModelId: "i1", aiPromptConfig: { comic: { enabled: true, presetId: "page", assets: { presetId: "assets" } } } } },
      name: "漫画", usable: true, reason: "",
    };
    h.setup.mockResolvedValue({ styles: STYLES, presets: [TEXT_ONLY_PRESET, comicPreset] });
    let tree = render();
    runEffects();
    await flush();
    tree = typeSource(SOURCE);
    expect(primaryButton(tree).props.disabled).toBe(false);
    primaryButton(tree).props.onClick();
    await flush();
    findByText(render(), "继续：生成配置确认").props.onClick();
    await flush();
    findByText(render(), "开始生成漫画").props.onClick();
    await flush();
    expect(h.submit).toHaveBeenCalledWith(expect.objectContaining({ presetId: "p1" }), expect.objectContaining({ id: "p1" }));
  });

  test("建批 500：错误文案可见、按钮恢复可点、且只请求一次", async () => {
    render();
    runEffects();
    await flush();
    h.create.mockRejectedValueOnce(Object.assign(new Error("请求失败：500"), { statusCode: 500 }));
    let tree = typeSource(SOURCE);
    const button = primaryButton(tree);
    expect(button.props.disabled).toBe(false);
    button.props.onClick();
    await flush();
    tree = render();
    expect(textContent(tree)).toContain("服务端暂时忙不过来，稍后再试一次");
    expect(primaryButton(tree).props.disabled).toBe(false);
    expect(h.create).toHaveBeenCalledTimes(1);
  });

  test("错误显示后，后台异步回调（画风缩略图就绪）不会清空错误", async () => {
    let resolveThumb: (value: string) => void = () => {};
    h.thumb.mockReturnValue(new Promise((resolve) => { resolveThumb = resolve; }));
    render();
    runEffects();
    await flush();
    h.create.mockRejectedValueOnce(Object.assign(new Error("请求失败：500"), { statusCode: 500 }));
    let tree = typeSource(SOURCE);
    primaryButton(tree).props.onClick();
    await flush();
    tree = render();
    expect(textContent(tree)).toContain("服务端暂时忙不过来，稍后再试一次");

    resolveThumb("/tmp/style-1.png");
    await flush();
    tree = render();
    expect(textContent(tree)).toContain("服务端暂时忙不过来，稍后再试一次");
  });

  test("提交前阶段检查：状态读不到但仍在 comic_assets 时提示可等可继续，取消不提交", async () => {
    render(); runEffects(); await flush();
    primaryButton(typeSource(SOURCE)).props.onClick(); await flush();
    findByText(render(), "继续：生成配置确认").props.onClick();
    // 没有 state：readComicAutomation 返回 null，走「读不到状态」的兜底提示。
    h.readStatus.mockResolvedValue({ books: [{ id: "book-1", stage: "comic_assets" }] });
    h.showModal.mockResolvedValueOnce({ confirm: false, cancel: true });
    findByText(render(), "开始生成漫画").props.onClick(); await flush();
    expect(h.showModal).toHaveBeenCalledWith(expect.objectContaining({
      content: expect.stringContaining("建议等资产提取完成再提交"),
    }));
    expect(h.submit).not.toHaveBeenCalled();
    expect(textContent(render())).toContain("建议等它完成再提交");

    // 只提示不拦：用户再次点击并选择「继续提交」，必须真的提交。
    h.showModal.mockResolvedValueOnce({ confirm: true, cancel: false });
    findByText(render(), "开始生成漫画").props.onClick(); await flush();
    expect(h.submit).toHaveBeenCalledTimes(1);
  });

  test("提交前阶段检查：状态空闲时已过资产阶段不弹窗、不额外等待", async () => {
    render(); runEffects(); await flush();
    primaryButton(typeSource(SOURCE)).props.onClick(); await flush();
    findByText(render(), "继续：生成配置确认").props.onClick();
    h.readStatus.mockResolvedValue({ state: "idle", books: [{ id: "book-1", stage: "comic_bootstrap" }] });
    findByText(render(), "开始生成漫画").props.onClick(); await flush();
    expect(h.showModal).not.toHaveBeenCalled();
    expect(h.submit).toHaveBeenCalledTimes(1);
  });

  test("已有任务在跑：running 时直接拦下，不提交也不弹确认", async () => {
    render(); runEffects(); await flush();
    primaryButton(typeSource(SOURCE)).props.onClick(); await flush();
    findByText(render(), "继续：生成配置确认").props.onClick();
    h.readStatus.mockResolvedValue({ state: "running", books: [{ id: "book-1", stage: "comic_images" }] });
    findByText(render(), "开始生成漫画").props.onClick(); await flush();
    expect(h.showModal).toHaveBeenCalledWith(expect.objectContaining({
      title: "已有任务在跑",
      content: expect.stringContaining("生成中"),
      showCancel: false,
    }));
    expect(h.submit).not.toHaveBeenCalled();
    expect(textContent(render())).toContain("已有任务在跑（生成中），等它结束再提交");
  });

  test("终态再提交：必须确认会重新消耗额度，取消不提交、确认才提交", async () => {
    render(); runEffects(); await flush();
    primaryButton(typeSource(SOURCE)).props.onClick(); await flush();
    findByText(render(), "继续：生成配置确认").props.onClick();
    h.readStatus.mockResolvedValue({ state: "completed", books: [{ id: "book-1", stage: "ready_for_upload" }] });
    h.showModal.mockResolvedValueOnce({ confirm: false, cancel: true });
    findByText(render(), "开始生成漫画").props.onClick(); await flush();
    expect(h.showModal).toHaveBeenCalledWith(expect.objectContaining({
      title: "重新生成会再扣额度",
      content: expect.stringContaining("出图额度会重新扣一次"),
      confirmText: "重新生成",
      cancelText: "先不提交",
    }));
    expect(h.submit).not.toHaveBeenCalled();
    expect(textContent(render())).toContain("已取消重新生成");

    h.showModal.mockResolvedValueOnce({ confirm: true, cancel: false });
    findByText(render(), "开始生成漫画").props.onClick(); await flush();
    expect(h.submit).toHaveBeenCalledTimes(1);
  });

  test("提交区按字数给预计页数与时长区间，措辞是估算", async () => {
    render(); runEffects(); await flush();
    expect(textContent(typeSource("字".repeat(181)))).toContain("预计 2 页 · 约 6~6.5 分钟（估算）");
    expect(textContent(typeSource("字".repeat(100)))).toContain("预计 1 页 · 约 6~6.5 分钟（估算）");
    expect(textContent(typeSource("字".repeat(4000)))).toContain("预计 35 页 · 约 42~46.5 分钟（估算）");
    expect(textContent(typeSource("字".repeat(20)))).not.toContain("预计 ");
  });

  test("冷启动从本地指针恢复到资产阶段并重新读服务端资产", async () => {
    render(); runEffects(); await flush();
    primaryButton(typeSource(SOURCE)).props.onClick(); await flush();
    expect(h.storage.get(comicRunStorageKey()!)).toMatchObject({
      batchId: "batch-1", bookId: "book-1", submitted: false, source: SOURCE,
    });

    // 模拟小程序被回收后重新打开：组件状态清空，只保留本地存储。
    h.slots = []; h.effects = [];
    render();
    h.show();
    expect(h.slots[0]).toMatchObject({
      batchId: "batch-1", bookId: "book-1", stage: "assets", tab: "result", runSubmitted: false, restored: true,
    });
    render(); runEffects(); await flush();
    expect(h.assets).toHaveBeenCalledWith("batch-1", "book-1");
    expect(textContent(render())).toContain("人物与场景");
  });

  test("提交后冷启动恢复到进度页，轮询立即重启并按服务端校准", async () => {
    h.autoFinish = false;
    render(); runEffects(); await flush();
    primaryButton(typeSource(SOURCE)).props.onClick(); await flush();
    findByText(render(), "继续：生成配置确认").props.onClick();
    findByText(render(), "开始生成漫画").props.onClick(); await flush();
    expect(h.storage.get(comicRunStorageKey()!)).toMatchObject({ batchId: "batch-1", bookId: "book-1", submitted: true });

    h.slots = []; h.effects = []; h.pollStarts = 0;
    render();
    h.show();
    expect(h.slots[0]).toMatchObject({
      batchId: "batch-1", bookId: "book-1", stage: "render", tab: "result", runSubmitted: true, restored: true,
    });
    render(); runEffects(); await flush();
    expect(h.pollStarts).toBeGreaterThan(0);
    expect(textContent(render())).toContain("已恢复上次进度，正在向服务端校准");
  });

  test("退到后台只停表，本地指针与已提交状态都保留", async () => {
    h.autoFinish = false;
    render(); runEffects(); await flush();
    primaryButton(typeSource(SOURCE)).props.onClick(); await flush();
    findByText(render(), "继续：生成配置确认").props.onClick();
    findByText(render(), "开始生成漫画").props.onClick(); await flush();
    render(); runEffects(); await flush();
    h.hide();
    expect(h.pollStop).toHaveBeenCalled();
    expect(h.slots[0].runSubmitted).toBe(true);
    expect(h.storage.get(comicRunStorageKey()!)).toBeTruthy();
  });

  test("读进度遇到 404 清掉本地指针，不再反复撞已不存在的批次", async () => {
    h.autoFinish = false;
    render(); runEffects(); await flush();
    primaryButton(typeSource(SOURCE)).props.onClick(); await flush();
    findByText(render(), "继续：生成配置确认").props.onClick();
    findByText(render(), "开始生成漫画").props.onClick(); await flush();
    render(); runEffects(); await flush();
    h.pollFail(Object.assign(new Error("批次不存在（HTTP 404）"), { statusCode: 404 }));
    expect(h.storage.has(comicRunStorageKey()!)).toBe(false);
  });

  test("本地指针超过 24h 自动失效，不恢复旧批次", async () => {
    h.storage.set(comicRunStorageKey()!, {
      version: 1, batchId: "old-batch", bookId: "old-book", source: "旧正文", sourceLength: 120,
      styleId: "style-1", aspectRatio: "9:16", reference: null, submitted: true,
      startedAt: Date.now() - 25 * 60 * 60 * 1000,
    });
    render();
    h.show();
    expect(h.slots[0].batchId).toBe("");
    expect(h.storage.has(comicRunStorageKey()!)).toBe(false);
  });
});
