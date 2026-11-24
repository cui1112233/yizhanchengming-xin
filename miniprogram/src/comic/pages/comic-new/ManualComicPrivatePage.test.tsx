import {configureComicRecoveryScope,comicRunStorageKey,comicSelectedBookStorageKey} from '../../services/comicRecovery';
import { beforeEach, describe, expect, test, vi } from "vitest";
const h = vi.hoisted(() => ({
    generatePrompts: vi.fn(), saveSettings: vi.fn(), watchPages: vi.fn(), watchStop: vi.fn(), watchCallbacks: null as any,
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
        if (!(index in slots))
            slots[index] = typeof initial === "function" ? initial() : initial;
        return [slots[index], (next: any) => { slots[index] = typeof next === "function" ? next(slots[index]) : next; }];
    },
    useRef: (initial: any) => {
        const index = h.cursor++;
        const slots = h.active || h.slots;
        if (!(index in slots))
            slots[index] = { current: initial };
        return slots[index];
    },
    useEffect: (effect: any, deps?: any[]) => {
        const index = h.cursor++;
        const prior = (h.active || h.slots)[index];
        if (prior && deps && JSON.stringify(prior.deps) === JSON.stringify(deps))
            return;
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
            if (h.autoFinish)
                publish({ counts: { total: 1, succeeded: 1, failed: 0 }, ...h.pollStatus });
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
vi.mock("../../services/comicManual", async (original) => ({
    ...await original<any>(), generateComicPrompts: h.generatePrompts, saveComicManualSettings: h.saveSettings,
    watchComicPageImages: h.watchPages,
}));
import ComicPrivatePage from "./ManualComicPrivatePage";
import {ComicEntryContent as ComicEntry, LegacyComicPrivatePage } from "./index";
import ComicStageHeader from "./ComicStageHeader";
import ComicPageWorkspace from "./ComicPageWorkspace";
import ComicSettingsStage from "./ComicSettingsStage";
import ComicAssetsStage from "./ComicAssetsStage";
import ComicConfigConfirmStage from "./ComicConfigConfirmStage";
import ComicStickySpacer, { COMIC_STICKY_SPACER_HEIGHT } from "../../components/ComicStickySpacer";
function rawRender() { h.active = h.slots; h.cursor = 0; return ComicPrivatePage() as any; }
function expandSettings(node: any): any {
    if (Array.isArray(node))
        return node.map(expandSettings);
    if (!node || typeof node !== "object")
        return node;
    if ([ComicSettingsStage, ComicAssetsStage, ComicConfigConfirmStage, ComicPageWorkspace, ComicStageHeader].includes(node.type)) {
        const key = node.type.name + ":" + (node.key || "");
        if (!h.childSlots.has(key))
            h.childSlots.set(key, []);
        const active = h.active;
        const cursor = h.cursor;
        h.active = h.childSlots.get(key)!;
        h.cursor = 0;
        const result = node.type(node.props);
        h.active = active;
        h.cursor = cursor;
        return expandSettings(result);
    }
    return { ...node, props: { ...node.props, children: expandSettings(node.props?.children) } };
}
function render() { return expandSettings(rawRender()); }
async function flush() { for (let index = 0; index < 40; index += 1)
    await Promise.resolve(); }
function runEffects() { for (const effect of h.effects.splice(0))
    effect(); }
function textContent(node: any): string {
    if (Array.isArray(node))
        return node.map(textContent).join("");
    if (node === null || node === undefined || typeof node === "boolean")
        return "";
    if (typeof node !== "object")
        return String(node);
    return textContent(node.props?.children);
}
function nodesOfType(tree: any, type: any, found: any[] = []): any[] {
    if (Array.isArray(tree)) {
        tree.forEach((child) => nodesOfType(child, type, found));
        return found;
    }
    if (!tree || typeof tree !== "object")
        return found;
    if (tree.type === type)
        found.push(tree);
    nodesOfType(tree.props?.children, type, found);
    return found;
}
function findByText(node: any, label: string): any {
    if (Array.isArray(node))
        return node.map((child) => findByText(child, label)).find(Boolean);
    if (!node || typeof node !== "object")
        return undefined;
    if (typeof node.props?.onClick === "function" && textContent(node.props?.children).includes(label))
        return node;
    return findByText(node.props?.children, label);
}
function findSetting(label: string) {
    if ((label === "图片描述" || label === "画质与约束") && !findByText(render(), label))
        findByText(render(), "提示词").props.onClick();
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
    h.slots = [];
    h.active = null;
    h.childSlots.clear();
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
    h.assetImages.mockImplementation(async (_batchId: string, _bookId: string, assetId: string) => (assetId === "a1" ? [{ id: "img-a1", url: "/content", mediaType: "image/png", isPrimary: true }] : []));
    // 取图失败（例如白名单/网络问题）时静默返回空串，界面只显示占位。
    h.assetPath.mockResolvedValue("");
    h.extractAssets.mockResolvedValue({});
    h.generateImages.mockResolvedValue({ images: [{id:"asset-image-valid"}], taskId: "", warning: "" });
    h.retryStage.mockResolvedValue({});
    h.setPrimary.mockResolvedValue({});
    h.regeneratePage.mockResolvedValue({});
    h.chooseMedia.mockResolvedValue({ tempFiles: [{ tempFilePath: "/tmp/ref.png" }] });
    h.getClipboardData.mockResolvedValue({ data: "" });
    h.showModal.mockResolvedValue({ confirm: true, cancel: false });
    h.showActionSheet.mockResolvedValue({ tapIndex: 0 });
});
const PROMPT_PAGES = [{ pageId: "page-1", videoId: "video-1", ordinal: 1, status: "pending", aspectRatio: "9:16", comicPrompt: "分页描述", imagePrompt: "图片描述", primaryImageId: "", images: [] }];
beforeEach(() => {configureComicRecoveryScope("test-owner");
    h.pages.mockResolvedValue([]);
    h.saveSettings.mockResolvedValue({});
    h.generatePrompts.mockImplementation(async () => { h.pages.mockResolvedValue(PROMPT_PAGES); });
    h.watchPages.mockImplementation((_b: any, _k: any, _baseline: any, callbacks: any) => { h.watchCallbacks = callbacks; return h.watchStop; });
});
async function extractFirst() { render(); runEffects(); await flush(); primaryButton(typeSource(SOURCE)).props.onClick(); await flush(); }
async function generatePromptsFirst() { await extractFirst(); findByText(render(), "继续：生成配置确认").props.onClick(); findByText(render(), "生成漫画提示词").props.onClick(); await flush(); }
describe("当前私人定制手动链路", () => {
    test("首次按钮顺序创建、实际提取并回读，不走自动化", async () => {
        await extractFirst();
        expect(h.create).toHaveBeenCalledTimes(1);
        expect(h.extractAssets).toHaveBeenCalledWith(expect.objectContaining({ batchId: "batch-1", bookId: "book-1", mode: "missing" }));
        expect(h.create.mock.invocationCallOrder[0]).toBeLessThan(h.extractAssets.mock.invocationCallOrder[0]);
        expect(h.extractAssets.mock.invocationCallOrder[0]).toBeLessThan(h.assets.mock.invocationCallOrder[0]);
        expect(textContent(render())).toContain("林岚");
        expect(h.submit).not.toHaveBeenCalled();
        expect(h.generatePrompts).not.toHaveBeenCalled();
    });
    test("提取失败保留项目，重试同正文不重复创建", async () => {
        h.extractAssets.mockRejectedValueOnce(Object.assign(new Error("提取暂时失败"),{statusCode:400}));
        await extractFirst();
        expect(textContent(render())).toContain("提取暂时失败");
        expect(h.storage.get(comicRunStorageKey()!)).toMatchObject({ batchId: "batch-1", mode: "manual" });
        findByText(render(), "更多").props.onClick();
        findByText(render(), "重新提取资产").props.onClick();
        await flush();
        expect(h.create).toHaveBeenCalledTimes(1);
        expect(h.extractAssets).toHaveBeenCalledTimes(2);
        expect(textContent(render())).toContain("林岚");
    });
    test("确认只生成提示词，必须另点图片按钮才会出图", async () => {
        await generatePromptsFirst();
        expect(h.generatePrompts).toHaveBeenCalledWith("batch-1", "book-1", "bootstrap");
        expect(h.regeneratePage).not.toHaveBeenCalled();
        expect(h.submit).not.toHaveBeenCalled();
        expect(textContent(render())).toContain("提示词预览");
        findByText(render(), "生成漫画图片").props.onClick();
        await flush();
        expect(h.regeneratePage).toHaveBeenCalledTimes(1);
        expect(h.regeneratePage).toHaveBeenCalledWith("batch-1", "book-1", "video-1", expect.objectContaining({ imagePrompt: "图片描述" }));
        expect(h.watchPages).toHaveBeenCalled();
    });
    test("同步双击提取与提示词各只提交一次", async () => {
        render();
        runEffects();
        await flush();
        const first = primaryButton(typeSource(SOURCE));
        first.props.onClick();
        first.props.onClick();
        await flush();
        expect(h.create).toHaveBeenCalledTimes(1);
        expect(h.extractAssets).toHaveBeenCalledTimes(1);
        findByText(render(), "继续：生成配置确认").props.onClick();
        const next = findByText(render(), "生成漫画提示词");
        next.props.onClick();
        next.props.onClick();
        await flush();
        expect(h.generatePrompts).toHaveBeenCalledTimes(1);
        expect(h.regeneratePage).not.toHaveBeenCalled();
    });
    test("缺资产图时明确提示且不提交出图", async () => {
        h.assetImages.mockResolvedValue([]);
        await generatePromptsFirst();
        const button = findByText(render(), "生成漫画图片");
        expect(button.props.disabled).toBe(true);
        expect(textContent(render())).toContain("至少一张资产图");
        expect(h.regeneratePage).not.toHaveBeenCalled();
    });
    test("提示词成功且没有图片时也可以导出ZIP", async () => {
        await generatePromptsFirst();
        expect(findByText(render(), "导出漫画").props.disabled).toBe(false);
    });
    test("恢复手动项目只读回页图，不自动重复生成", async () => {
        h.storage.set(comicRunStorageKey()!, { version: 1, batchId: "batch-1", bookId: "book-1", source: SOURCE, sourceLength: SOURCE.length, styleId: "style-1", presetId: "p2", presetConfig: DUAL_PRESET.preset.config, aspectRatio: "9:16", reference: null, submitted: true, startedAt: Date.now(), mode: "manual", phase: "pages", pendingImages: { "video-1": [] } });
        render();
        runEffects();
        await flush();
        expect(h.pages).toHaveBeenCalled();
        expect(h.watchPages).toHaveBeenCalled();
        expect(h.create).not.toHaveBeenCalled();
        expect(h.generatePrompts).not.toHaveBeenCalled();
        expect(h.regeneratePage).not.toHaveBeenCalled();
        expect(h.submit).not.toHaveBeenCalled();
    });
    test("退后台停止轮询，回前台恢复读取而非重发出图", async () => {
        await generatePromptsFirst();
        findByText(render(), "生成漫画图片").props.onClick();
        await flush();
        render();
        h.hide?.();
        expect(h.watchStop).toHaveBeenCalled();
        render();
        h.show?.();
        await flush();
        expect(h.watchPages).toHaveBeenCalledTimes(2);
        expect(h.regeneratePage).toHaveBeenCalledTimes(1);
    });
    test("参考图先写入项目设置，再提取资产", async () => {
        render();
        runEffects();
        await flush();
        h.slots[0] = { ...h.slots[0], referencePath: "/tmp/ref.png" };
        primaryButton(typeSource(SOURCE)).props.onClick();
        await flush();
        expect(h.upload).toHaveBeenCalledTimes(1);
        expect(h.saveSettings).toHaveBeenCalledTimes(1);
        expect(h.saveSettings.mock.invocationCallOrder[0]).toBeLessThan(h.extractAssets.mock.invocationCallOrder[0]);
    });
});
test("入口只对旧已提交自动任务使用兼容页，新建与手动项目都走手动页", () => {
    h.active = h.slots;
    h.cursor = 0;
    expect((ComicEntry() as any).type).toBe(ComicPrivatePage);
    h.slots = [];
    h.active = h.slots;
    h.cursor = 0;
    h.storage.set(comicRunStorageKey()!, { version: 1, batchId: "b", bookId: "k", submitted: true, startedAt: Date.now() });
    const legacy = ComicEntry() as any;
    expect(legacy.type).toBe(LegacyComicPrivatePage);
    legacy.props.onNew();
    h.active = h.slots;
    h.cursor = 0;
    expect((ComicEntry() as any).type).toBe(ComicPrivatePage);
    expect(h.storage.has(comicRunStorageKey()!)).toBe(false);
});
test("新一轮页图回读完成后，旧缩略图请求不能覆盖它", async () => {
    await generatePromptsFirst();
    findByText(render(), "生成漫画图片").props.onClick();
    await flush();
    let resolve: any;
    h.loadPaths.mockReturnValueOnce(new Promise(r => resolve = r));
    h.watchCallbacks.onPages([{ ...PROMPT_PAGES[0], primaryImageId: "old", images: [{ id: "old" }] }]);
    await flush();
    h.loadPaths.mockResolvedValue({ new: "/new.png" });
    h.watchCallbacks.onPages([{ ...PROMPT_PAGES[0], primaryImageId: "new", images: [{ id: "new" }] }]);
    await flush();
    resolve({ old: "/old.png" });
    await flush();
    expect(h.slots[0].paths).toEqual({ new: "/new.png" });
});

test("完成后可开始新作品，清除活动指针但不自动创建",async()=>{
 await generatePromptsFirst();findByText(render(),"新作品").props.onClick();await flush();expect(h.slots[0].project).toBeNull();expect(h.storage.has(comicRunStorageKey()!)).toBe(false);expect(textContent(render())).toContain("从一段故事开始");expect(h.create).toHaveBeenCalledTimes(1);
});

test("离开恢复的旧项目后，迟到的读取错误不能污染新作品",async()=>{
 h.storage.set(comicRunStorageKey()!,{version:1,batchId:"old",bookId:"book",source:SOURCE,sourceLength:SOURCE.length,styleId:"style-1",presetId:"p2",presetConfig:DUAL_PRESET.preset.config,aspectRatio:"9:16",reference:null,submitted:true,startedAt:Date.now(),mode:"manual",phase:"pages"});
 let reject:any;h.pages.mockReturnValueOnce(new Promise((_r,j)=>reject=j));render();runEffects();await flush();findByText(render(),"新作品").props.onClick();await flush();reject(new Error("旧项目读取失败"));await flush();expect(textContent(render())).not.toContain("旧项目读取失败");expect(h.slots[0].project).toBeNull();
});

test('an empty asset generation receipt remains blocked instead of permitting another charge',async()=>{h.generateImages.mockResolvedValue({images:[],taskId:'',warning:''});await extractFirst();const generate=nodesOfType(render(),'Button').find(n=>textContent(n.props.children)==='生成图片');expect(generate?.props.disabled).toBe(false);generate.props.onClick();await flush();expect(h.generateImages).toHaveBeenCalledTimes(1);expect(textContent(render())).toContain('回执未确认');nodesOfType(render(),'Button').find(n=>textContent(n.props.children)==='生成图片').props.onClick();await flush();expect(h.generateImages).toHaveBeenCalledTimes(1);expect(textContent(render())).toContain('待核对');});
