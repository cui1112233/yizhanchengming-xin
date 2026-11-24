import Taro, { useDidHide, useDidShow } from "@tarojs/taro";
import { Button, Image, ScrollView, Text, View } from "@tarojs/components";
import { useEffect, useRef, useState } from "react";
import BrandHeader from "../../../components/BrandHeader";
import AppIcon from "../../../components/AppIcon";
import ComicPageWorkspace from "./ComicPageWorkspace";
import ComicStageHeader from "./ComicStageHeader";
import ComicSettingsStage from "./ComicSettingsStage";
import ComicAssetsStage from "./ComicAssetsStage";
import ComicConfigConfirmStage from "./ComicConfigConfirmStage";

import type { ComicStage } from "./stageState";
import { canEnterPrivateComicStage } from "./privateComicParity";
import { ThemedPage } from "../../../theme";
import type { BatchFactoryAutomationStatus } from "../../../services/batchFactoryV11";
import {
  COMIC_PRESET_EMPTY_HINT, DEFAULT_COMIC_ASPECT_RATIO,
  DEFAULT_COMIC_RUN_MODE, MIN_COMIC_SOURCE_CHARS, buildComicInitialBatchSettings,
  applyComicSettingSelections,
  comicFailureText, comicNewErrorText, comicPresetAvailability,
  comicPresetComicReason, comicProgress, comicSourceLength, createComicBatch, loadComicPrivateSetup,
  readComicAutomation, resolveComicStyleThumb, startComicPolling, submitComicRun, uploadComicReference, type ComicReferenceInfo,
  validateComicSource, type ComicModelOption, type ComicPresetOption, type ComicSettingCatalog, type ComicSettingSelections, type ComicStylePreset,
} from "../../services/comicFlow";
import {
  extractComicAssets,
  generateComicAssetImages, loadComicAssetImagePath, loadComicAssetImages, loadComicAssets,
  readComicAssetImageTask, retryComicBookStage,
  type ComicAssetImageItem, type ComicAssetItem,
} from "../../services/comicAssets";
import {
  buildPreviewRequest, createComicCollection, downloadComicCollectionZip, downloadComicPagePackage, loadComicImagePaths,
  loadComicPages, regenerateComicPage, saveComicImageToAlbum, setPrimaryComicPageImage, type ComicPageItem,
} from "../../services/comicResult";
import {
  COMIC_RUN_STEPS, assetImagesSkipped, comicRunSubmitGate, runStageView, runStateLabel,
} from "../../services/runStage";
import {
  clearComicActiveRun, readComicActiveRun, saveComicActiveRun, shouldForgetComicRun,
} from "../../services/comicRecovery";
import "../../styles/comic.scss";
import "./editor.scss";

/**
 * 私人定制漫画：对标剧本生成的单页工作区。
 * 粘贴正文 → 选画风 / 画幅 / 参考图 → 吸底「开始生成」→ 通过顶部四步流程看进度和页图。
 * 界面上不出现批次、不出现选书；批次由 createComicBatch 在后台建一个单书批次。
 */

type ComicProjectSnapshot = Readonly<{
  sourceLength: number;
  styleId: string;
  presetId: string;
  presetConfig: Readonly<Record<string, unknown>>;
  aspectRatio: string;
  reference: Readonly<ComicReferenceInfo> | null;
}>;

type ComicAssetScope = Readonly<{ batchId: string; bookId: string }>;
type ComicSettingPicker = "" | "prompt" | "derivation" | "model" | "constraints";

type PrivateComicState = {
  stage: ComicStage;
  tab: "create" | "result";
  source: string;
  styleId: string;
  aspectRatio: string;
  referencePath: string;
  projectSnapshot: ComicProjectSnapshot | null;
  styles: ComicStylePreset[];
  thumbs: Record<string, string>;
  /** 缩略图加载失败单独记一位：界面要从骨架切到「重试」，不能停在永远空着。 */
  thumbFailed: Record<string, boolean>;
  presets: ComicPresetOption[];
  settingCatalog: ComicSettingCatalog;
  models: ComicModelOption[];
  settingSelections: ComicSettingSelections;
  selectedPresetId: string;
  advancedOpen: boolean;
  settingPicker: ComicSettingPicker;
  setupLoading: boolean;
  setupError: string;
  submitting: boolean;
  runSubmitted: boolean;
  error: string;
  batchId: string;
  bookId: string;
  status: BatchFactoryAutomationStatus | null;
  progressError: string;
  pages: ComicPageItem[];
  paths: Record<string, string>;
  pageIndex: number;
  resultLoading: boolean;
  resultError: string;
  busy: string;
  referenceOpen: boolean;
  assets: ComicAssetItem[];
  assetImages: Record<string, ComicAssetImageItem[]>;
  assetThumbs: Record<string, string>;
  assetsLoading: boolean;
  assetError: string;
  assetBusy: "" | "extract" | "images" | "retry";
  assetBusyAt: number;
  pollToken: number;
  /** 本页是从本地指针恢复出来的：界面要标出来，别假装是实时进度。 */
  restored: boolean;
};

const INITIAL_STATE: PrivateComicState = {
  stage: "source",
  tab: "create",
  source: "",
  styleId: "",
  aspectRatio: DEFAULT_COMIC_ASPECT_RATIO,
  referencePath: "",
  projectSnapshot: null,
  styles: [],
  thumbs: {},
  thumbFailed: {},
  presets: [],
  settingCatalog: { pagePrompts: [], imageDerivations: [], constraints: { quality: [], restriction: [], negative: [] } },
  models: [],
  settingSelections: {},
  selectedPresetId: "",
  advancedOpen: false,
  settingPicker: "",
  setupLoading: true,
  setupError: "",
  submitting: false,
  runSubmitted: false,
  error: "",
  batchId: "",
  bookId: "",
  status: null,
  progressError: "",
  pages: [],
  paths: {},
  pageIndex: 0,
  resultLoading: false,
  resultError: "",
  busy: "",
  referenceOpen: false,
  assets: [],
  assetImages: {},
  assetThumbs: {},
  assetsLoading: false,
  assetError: "",
  assetBusy: "",
  assetBusyAt: 0,
  pollToken: 0,
  restored: false,
};

export default function ComicPrivatePage() {
  const [state, setState] = useState<PrivateComicState>(INITIAL_STATE);
  const stopPolling = useRef<(() => void) | undefined>(undefined);
  const resultLoaded = useRef(false);
  const imageVersion = useRef(0);
  const assetsLoadedFor = useRef("");
  const assetsInFlight = useRef("");
  const runSubmitting = useRef(false);
  const pageActive = useRef(true);
  const assetScope = useRef<ComicAssetScope | null>(null);
  const submittedScope = useRef<ComicAssetScope | null>(null);
  // 资产阶段自己的 5s 轮询，退后台时和进度轮询一起停表。
  const assetPollStop = useRef<(() => void) | undefined>(undefined);
  // onShow / onHide 是页面级回调，拿不到最新闭包；用 ref 读当前 state。
  const latest = useRef(state);
  latest.current = state;

  const patch = (next: Partial<PrivateComicState>) => setState((prev) => ({ ...prev, ...next }));

  /**
   * 本地活动指针：只记服务端已经存在的批次与提交状态，用于退出重进 / 被回收后恢复。
   * 进行中的提取、补齐、重试没有可持久化句柄，恢复后只能按服务端结果手动重试。
   */
  const rememberActiveRun = (batchId: string, bookId: string, submitted: boolean, project: ComicProjectSnapshot | null) => {
    if (!batchId || !bookId) return;
    const current = latest.current;
    const snapshot = project || current.projectSnapshot;
    const existing = readComicActiveRun();
    const sameRun = existing?.batchId === batchId && existing?.bookId === bookId;
    saveComicActiveRun({
      batchId,
      bookId,
      source: sameRun && existing ? existing.source : current.source,
      sourceLength: snapshot?.sourceLength ?? comicSourceLength(current.source),
      styleId: snapshot?.styleId || current.styleId,
      presetId: snapshot?.presetId || current.selectedPresetId,
      presetConfig: snapshot?.presetConfig ? JSON.parse(JSON.stringify(snapshot.presetConfig)) : {},
      aspectRatio: snapshot?.aspectRatio || current.aspectRatio,
      reference: snapshot?.reference
        ? {
          assetId: snapshot.reference.assetId,
          url: snapshot.reference.url,
          ...(snapshot.reference.previewUrl ? { previewUrl: snapshot.reference.previewUrl } : {}),
        }
        : null,
      submitted,
      startedAt: sameRun && existing ? existing.startedAt : Date.now(),
    });
  };

  /**
   * 资产只读加载：列表 + 每个资产的首张定妆图（优先 isPrimary）。
   * 任何一张图失败都静默落空，界面上显示「未生成」，不弹错误层。
   */
  const loadAssets = async (batchId: string, bookId: string, force = false, scope = assetScope.current) => {
    if (!scope || scope !== assetScope.current || scope.batchId !== batchId || scope.bookId !== bookId) return;
    const key = `${batchId}:${bookId}`;
    if (!force && assetsLoadedFor.current === key) return;
    // 并发窗口：切页 effect 与轮询回调可能同时触发，同一批次在途时直接复用，不发第二次。
    if (assetsInFlight.current === key) return;
    assetsInFlight.current = key;
    assetsLoadedFor.current = key;
    patch({ assetsLoading: true, assetError: "" });
    try {
      const assets = await loadComicAssets(batchId, bookId);
      if (scope !== assetScope.current) return;
      const entries = await Promise.all(assets.map(async (asset) => {
        let images: ComicAssetImageItem[] = [];
        try {
          images = await loadComicAssetImages(batchId, bookId, asset.id);
        } catch {
          images = [];
        }
        const primary = images.find((image) => image.isPrimary) || images[0];
        // 单张媒体失败只影响这一项，不能阻断其余资产列表。
        let path = "";
        try { path = primary ? await loadComicAssetImagePath(primary) : ""; } catch { /* 显示本项占位。 */ }
        return [asset.id, { images, path }] as const;
      }));
      if (scope !== assetScope.current) return;
      patch({
        assets,
        assetsLoading: false,
        assetImages: Object.fromEntries(entries.map(([id, value]) => [id, value.images])),
        assetThumbs: Object.fromEntries(entries.filter(([, value]) => value.path).map(([id, value]) => [id, value.path])),
      });
    } catch (cause) {
      if (scope !== assetScope.current) return;
      assetsLoadedFor.current = "";
      patch({ assetsLoading: false, assetError: comicFailureText(cause, "资产暂时读不到，请刷新重试") });
    } finally {
      if (assetsInFlight.current === key) assetsInFlight.current = "";
    }
  };

  const loadPageImages = async (page: ComicPageItem, version = imageVersion.current) => {
    const loaded = await loadComicImagePaths(page);
    if (version !== imageVersion.current) return;
    setState((prev) => ({ ...prev, paths: { ...prev.paths, ...loaded } }));
  };

  const loadPages = async (batchId: string, bookId: string, refresh = false) => {
    if (resultLoaded.current && !refresh) return;
    resultLoaded.current = true;
    const version = imageVersion.current;
    patch({ resultLoading: true, resultError: "" });
    try {
      const next = await loadComicPages(batchId, bookId);
      if (version !== imageVersion.current) return;
      const hasImageReceipt = next.some((page) => page.images.length > 0 && Boolean(page.primaryImageId));
      const index = refresh ? Math.min(latest.current.pageIndex, Math.max(0, next.length - 1)) : 0;
      patch({ pages: next, resultLoading: false, pageIndex: index, paths: {}, ...(hasImageReceipt ? { stage: "result" as const } : {}) });
      if (next[index]) await loadPageImages(next[index], version);
      // 仅预取其他页主图，候选图在选页时加载；三页一组避免同时请求几十张。
      for (let offset = 0; offset < next.length; offset += 3) {
        if (version !== imageVersion.current) return;
        await Promise.all(next.slice(offset, offset + 3).filter((_, i) => offset + i !== index).map(page => {
          const main = page.images.find(image => image.id === page.primaryImageId);
          return main ? loadPageImages({ ...page, images: [main] }, version) : Promise.resolve();
        }));
      }
    } catch (cause: any) {
      // 终态不再轮询；页图未落库时允许用户只重读结果。
      if (version !== imageVersion.current) return;
      resultLoaded.current = false;
      patch({ resultLoading: false, resultError: cause?.message || "暂时打不开这些图，请稍后再试" });
    }
  };

  /** 画风缩略图：成功进 thumbs，失败单独记一位，界面从骨架切到「重试」。 */
  const loadStyleThumb = async (style: ComicStylePreset) => {
    if (!style.previewUrl) return;
    const path = await resolveComicStyleThumb(style.previewUrl);
    if (!pageActive.current) return;
    setState((prev) => path
      ? { ...prev, thumbs: { ...prev.thumbs, [style.id]: path } }
      : { ...prev, thumbFailed: { ...prev.thumbFailed, [style.id]: true } });
  };

  /** 失败态的重试入口：失败不会写缓存，直接重跑即可重新请求。 */
  const retryStyleThumb = (styleId: string) => {
    const style = state.styles.find((item) => item.id === styleId);
    if (!style?.previewUrl) return;
    setState((prev) => ({ ...prev, thumbFailed: { ...prev.thumbFailed, [styleId]: false } }));
    void loadStyleThumb(style);
  };

  useEffect(() => {
    let active = true;
    void loadComicPrivateSetup().then(({ styles, presets, settingCatalog, models }) => {
      if (!active) return;
      patch({
        styles, presets,
        settingCatalog: settingCatalog || INITIAL_STATE.settingCatalog,
        models: Array.isArray(models) ? models : [],
        styleId: state.styleId || styles[0]?.id || "",
        selectedPresetId: state.selectedPresetId || presets.find((option) => option.usable)?.preset.id || "",
        setupLoading: false, setupError: "",
      });
      styles.forEach((style) => { void loadStyleThumb(style); });
    }).catch((cause: any) => {
      if (active) patch({ setupLoading: false, setupError: comicFailureText(cause) });
    });
    return () => { active = false; };
  }, []);

  // Image requests can outlive the component or a completed run.
  useEffect(() => () => { pageActive.current = false; imageVersion.current += 1; assetScope.current = null; }, []);

  useEffect(() => {
    if (state.tab !== "result" || !state.batchId || !state.bookId || !state.runSubmitted || state.stage !== "render" || comicProgress(state.status).finished) return;
    const batchId = state.batchId;
    const bookId = state.bookId;
    let active = true;
    let stop: (() => void) | undefined;
    const cancel = () => { active = false; stop?.(); };
    stop = startComicPolling(
      () => readComicAutomation(batchId),
      (value) => {
        // 服务端偶发返回空对象：跳过这一帧，保留上一次进度，绝不当成已结束。
        if (!active || !value) return;
        patch({ status: value, progressError: "", restored: false });
        if (!comicProgress(value).finished) return;
        cancel();
        void loadPages(batchId, bookId);
        void loadAssets(batchId, bookId, true);
      },
      (cause) => {
        if (!active) return;
        cancel();
        // 401 / 404 说明本地指针指向的批次已经不可用，必须清掉，避免每次进来都撞同一堵墙。
        if (shouldForgetComicRun(cause)) clearComicActiveRun();
        patch({ progressError: cause instanceof Error ? cause.message : "暂时读不到进度，请重新进入分页出图阶段刷新" });
      },
    );
    // The callback can finish synchronously (for example cached status).
    if (!active) stop();
    stopPolling.current = cancel;
    void loadAssets(batchId, bookId);
    return () => { cancel(); stopPolling.current = undefined; };
  }, [state.tab, state.stage, state.batchId, state.bookId, state.runSubmitted, state.pollToken, comicProgress(state.status).finished]);

  useEffect(() => {
    if (!state.batchId || !state.bookId || (state.stage !== "assets" && state.stage !== "storyboard")) return;
    void loadAssets(state.batchId, state.bookId);
  }, [state.batchId, state.bookId, state.stage]);

  // 资产阶段只读服务端：资产由服务端在 comic_assets 阶段提取，
  // 这里只在阶段变化时重读列表，不再由客户端触发 stages/assets。
  useEffect(() => {
    if (state.stage !== "assets" || !state.batchId || !state.bookId) return;
    if (state.assets.length || state.assetError) return;
    const batchId = state.batchId;
    const bookId = state.bookId;
    const scope = { batchId, bookId };
    let active = true;
    let lastStage = "";
    const stop = startComicPolling(
      () => readComicAutomation(batchId),
      (value) => {
        if (!active || !value) return;
        patch({ status: value });
        const stage = String(value.books?.[0]?.stage || "");
        if (!stage || stage === lastStage) return;
        lastStage = stage;
        void loadAssets(batchId, bookId, true, scope);
      },
      (cause) => {
        // 资产阶段读不到进度不打扰，保留「刷新资产」入口；但 401/404 要清掉本地指针。
        if (shouldForgetComicRun(cause)) clearComicActiveRun();
      },
    );
    assetPollStop.current = stop;
    return () => { active = false; stop(); assetPollStop.current = undefined; };
  }, [state.stage, state.batchId, state.bookId, state.assets.length, state.assetError, state.pollToken]);

  /**
   * 回前台校准：
   * - 有批次：让轮询立即重启（pollToken +1），由服务端状态覆盖本地；
   * - 没批次：读 24h 内的本地活动指针，恢复批次和提交状态。
   * 只恢复「批次 + 提交状态」，退出瞬间正在跑的动作没有句柄，恢复后按服务端结果手动重试。
   */
  useDidShow(() => {
    const current = latest.current;
    const selected = Taro.getStorageSync("comic:selected-existing-book") as { sourceText?: unknown } | "";
    const selectedSource = selected && typeof selected === "object" && typeof selected.sourceText === "string" ? selected.sourceText.trim() : "";
    if (selectedSource) {
      Taro.removeStorageSync("comic:selected-existing-book");
      patch({ source: selectedSource, error: "", stage: "source", tab: "create" });
      return;
    }
    if (current.batchId && current.bookId) {
      if (current.runSubmitted) setState((prev) => ({ ...prev, pollToken: prev.pollToken + 1 }));
      else {
        // 资产阶段：重启被 onHide 停掉的 5s 轮询，并强制重读一次服务端资产。
        setState((prev) => ({ ...prev, pollToken: prev.pollToken + 1 }));
        void loadAssets(current.batchId, current.bookId, true);
      }
      return;
    }
    const run = readComicActiveRun();
    if (!run) return;
    const scope = { batchId: run.batchId, bookId: run.bookId };
    assetScope.current = scope;
    assetsLoadedFor.current = "";
    assetsInFlight.current = "";
    resultLoaded.current = false;
    imageVersion.current += 1;
    patch({
      stage: run.submitted ? "render" : "assets",
      tab: "result",
      batchId: run.batchId,
      bookId: run.bookId,
      source: run.source,
      styleId: run.styleId || current.styleId,
      aspectRatio: run.aspectRatio || DEFAULT_COMIC_ASPECT_RATIO,
      projectSnapshot: Object.freeze({
        sourceLength: run.sourceLength || comicSourceLength(run.source),
        styleId: run.styleId || current.styleId,
        presetId: run.presetId || current.selectedPresetId,
        presetConfig: Object.freeze({ ...run.presetConfig }),
        aspectRatio: run.aspectRatio || DEFAULT_COMIC_ASPECT_RATIO,
        reference: run.reference ? Object.freeze({ ...run.reference }) : null,
      }),
      runSubmitted: run.submitted,
      status: null,
      pages: [], paths: {}, pageIndex: 0,
      progressError: "", resultError: "", error: "",
      assets: [], assetImages: {}, assetThumbs: {}, assetsLoading: !run.submitted, assetError: "",
      assetBusy: "", assetBusyAt: 0,
      restored: true,
    });
  });

  // 退到后台两路轮询都停表：不空转；本地指针与服务端任务都保留，回前台再校准。
  useDidHide(() => {
    stopPolling.current?.();
    assetPollStop.current?.();
  });

  const progress = comicProgress(state.status);
  const finished = Boolean(state.status) && progress.finished;
  // 提交后只读服务端状态：automation.state 与本书 stage 都是后端字段，这里不推测阶段。
  const serverBook = state.status?.books?.find((book) => String(book.bookId || "") === String(state.bookId || ""));
  // 四步条只读：步骤全部由服务端 stage 映射，未在清单里的 stage 原样显示。
  const runView = runStageView(state.status, state.bookId);
  const running = state.runSubmitted && !finished;
  const length = comicSourceLength(state.source);
  const remaining = Math.max(0, MIN_COMIC_SOURCE_CHARS - length);
  const presetHint = comicPresetAvailability(state.presets);
  const usablePreset = state.presets.find((option) => option.usable && option.preset.id === state.selectedPresetId)
    || state.presets.find((option) => option.usable);
  const effectivePreset = usablePreset ? {
    ...usablePreset,
    preset: applyComicSettingSelections(usablePreset.preset, state.settingCatalog, state.models, state.settingSelections),
  } : undefined;
  const projectPresetBase = state.presets.find((option) => option.usable && option.preset.id === state.projectSnapshot?.presetId);
  const projectPreset = state.projectSnapshot && projectPresetBase
    ? {
      ...projectPresetBase,
      preset: {
        ...projectPresetBase.preset,
        config: Object.keys(state.projectSnapshot.presetConfig).length
          ? state.projectSnapshot.presetConfig
          : projectPresetBase.preset.config,
      },
    }
    : effectivePreset;
  // 默认用第一条「模型齐 + 漫画提示词已选」的预设；引用交给后端 hydrate，不在前端兜底。
  const presetConfigHint = effectivePreset ? comicPresetComicReason(effectivePreset.preset) : "";
  const configurationHint = state.setupLoading ? "正在读取画风和预设…"
    : state.setupError || (!state.styles.length ? "还没有可用画风，稍后再试" : !state.styleId ? "请选择漫画画风" : presetHint || presetConfigHint || "");
  const blockReason = state.submitting ? "正在提交…"
    : running ? "上一次还在生成，先去「结果」看进度"
      : state.setupLoading ? "正在读取画风和预设…"
        : state.setupError ? state.setupError
          : !state.styles.length ? "还没有可用画风，稍后再试"
            : presetHint || presetConfigHint
              || (!length ? "先粘贴小说正文"
                : remaining > 0 ? `还需 ${remaining} 字`
                  : "");
  const canSubmit = !blockReason;

  const chooseReference = async () => {
    try {
      const result = await Taro.chooseMedia({ count: 1, mediaType: ["image"], sourceType: ["album", "camera"] });
      const file = result.tempFiles?.[0] as { tempFilePath?: string; path?: string } | undefined;
      const path = String(file?.tempFilePath || file?.path || "");
      if (path) patch({ referencePath: path });
    } catch { /* 取消选择不需要提示。 */ }
  };

  const pasteSource = async () => {
    try {
      const result = await Taro.getClipboardData();
      const source = String(result?.data || "").trim();
      if (source) patch({ source });
      else await Taro.showToast({ title: "剪贴板里还没有文字", icon: "none" });
    } catch {
      await Taro.showToast({ title: "请在输入框里长按粘贴", icon: "none" });
    }
  };

  const prepareProject = async () => {
    if (state.submitting) return;
    try {
      // 预检也放在 try 内：任何一步抛错都必须落到下面的 catch，不能静默吞掉。
      if (!canSubmit) { patch({ error: blockReason || validateComicSource(state.source) }); return; }
      const usable = effectivePreset;
      if (!usable) { patch({ error: COMIC_PRESET_EMPTY_HINT }); return; }
      const presetReason = comicPresetComicReason(usable.preset);
      if (presetReason) { patch({ error: presetReason }); return; }
      // 原生遮罩先于 React 渲染出现，保证点击瞬间就有可见反馈。
      Taro.showLoading({ title: "正在创建项目…", mask: true });
      patch({ submitting: true, error: "", assetsLoading: false, assetBusy: "" });
      const created = await createComicBatch({
        batchTitle: "",
        bookTitle: "",
        sourceText: state.source,
        // 网页版同款：建批就把画面配置带过去，服务端事务内 hydrate style.body。
        // 参考图此时还没上传，跟网页版一样留到 automation/start 的 settingsPatch 里补。
        initialBatchSettings: buildComicInitialBatchSettings(
          { aspectRatio: state.aspectRatio, stylePresetId: state.styleId, reference: null },
          usable.preset,
        ),
      });
      const reference = state.referencePath ? await uploadComicReference(created.batchId, state.referencePath) : null;
      // 当前项目只使用创建时的设置；返回设置修改的草稿属于下一次创作。
      const projectSnapshot: ComicProjectSnapshot = Object.freeze({
        sourceLength: comicSourceLength(state.source),
        styleId: state.styleId,
        presetId: usable.preset.id,
        presetConfig: Object.freeze(JSON.parse(JSON.stringify(usable.preset.config || {})) as Record<string, unknown>),
        aspectRatio: state.aspectRatio,
        reference: reference ? Object.freeze({ ...reference }) : null,
      });
      // 把完整预设配置一起带上：settingsPatch 里的 comic 必须整块，否则后端一层合并会把
      // comic.assets.body / comic.body / constraints 抹掉，资产阶段直接失败。
      resultLoaded.current = false;
      imageVersion.current += 1;
      const scope = { batchId: created.batchId, bookId: created.bookId };
      assetScope.current = scope;
      // 不再手工调 stages/assets：建批时已经带上 initialBatchSettings，服务端在 comic_assets 阶段自己提取。
      assetsLoadedFor.current = "";
      patch({
        submitting: false, batchId: created.batchId, bookId: created.bookId, stage: "assets", tab: "result",
        projectSnapshot, runSubmitted: false,
        status: null, pages: [], paths: {}, pageIndex: 0, progressError: "", resultError: "",
        assets: [], assetImages: {}, assetThumbs: {}, assetsLoading: true, assetError: "",
        assetBusy: "", assetBusyAt: 0,
      });
      Taro.hideLoading();
      // 建完批立刻落本地指针：退出或被回收后能回到这个批次，不必重新建批。
      rememberActiveRun(created.batchId, created.bookId, false, projectSnapshot);
      // 先读一次服务端已落库的资产；服务端还在提取时这里就是空列表，由资产阶段的轮询补齐。
      await loadAssets(created.batchId, created.bookId, true, scope);
    } catch (cause) {
      Taro.hideLoading();
      const message = comicNewErrorText(cause);
      patch({ submitting: false, error: message });
      // 页内错误行可能被滚出屏幕/被吸底条遮住，失败必须再给一次 toast，保证可见。
      void Taro.showToast({ title: message, icon: "none", duration: 2600 });
    }
  };

  const confirmConfigAndStart = async () => {
    const scope = assetScope.current;
    if (!scope || scope.batchId !== state.batchId || scope.bookId !== state.bookId || runSubmitting.current) return;
    // The ref also guards callbacks retained from the pre-submit render.
    if (state.runSubmitted || submittedScope.current === scope) {
      setState((prev) => ({ ...prev, stage: prev.pages.length ? "result" : "render", tab: "result" }));
      return;
    }
    const project = state.projectSnapshot;
    // 资产不再作为提交门禁：服务端在跑提取，客户端等到齐再放行只会让人干等。
    if (runSubmitting.current || state.stage !== "storyboard" || !project || !state.batchId || !state.bookId || state.submitting) return;
    const usable = projectPreset;
    if (!usable) { patch({ error: COMIC_PRESET_EMPTY_HINT }); return; }
    const presetReason = comicPresetComicReason(usable.preset);
    if (presetReason) { patch({ error: presetReason }); return; }
    runSubmitting.current = true;
    try {
      // 提交前只读一次 automation：活动任务直接拦下（否则只会拿 409），
      // 终态必须二次确认（会新建 run 并重新消耗出图额度）。读失败按「无法判断」走原流程。
      let automationState = "";
      let stage = "";
      try {
        const value = await readComicAutomation(state.batchId);
        automationState = String(value?.state || "");
        stage = String(value?.books?.[0]?.stage || "");
      } catch { automationState = ""; stage = ""; }
      const gate = comicRunSubmitGate(automationState);
      const stateLabel = runStateLabel(automationState);
      if (gate === "active") {
        await Taro.showModal({
          title: "已有任务在跑",
          content: `服务端当前状态是「${stateLabel}」，现在提交会被直接拒绝。请等它结束，或在进度页暂停、停止后再提交。`,
          showCancel: false,
          confirmText: "知道了",
        });
        patch({ submitting: false, error: `已有任务在跑（${stateLabel}），等它结束再提交` });
        return;
      }
      if (gate === "finished") {
        const result = await Taro.showModal({
          title: "重新生成会再扣额度",
          content: `当前任务${stateLabel}。继续会新建一个任务，从头重新生成全部漫画页，出图额度会重新扣一次。`,
          confirmText: "重新生成",
          cancelText: "先不提交",
        });
        if (!result?.confirm) {
          patch({ submitting: false, error: `当前任务${stateLabel}，已取消重新生成` });
          return;
        }
      } else if (!automationState && (stage === "comic_assets" || stage === "comic_asset_images")) {
        // 读不到 state 但 stage 显示还在提取资产：保留「只提示不拦」的兜底。
        const result = await Taro.showModal({
          title: "服务端还在提取资产",
          content: "建议等资产提取完成再提交，避免生成中途冲突。你也可以现在继续提交。",
          confirmText: "继续提交",
          cancelText: "再等等",
        });
        if (!result?.confirm) {
          patch({ submitting: false, error: "服务端还在提取资产，建议等它完成再提交" });
          return;
        }
      }
      patch({ submitting: true, error: "" });
      await submitComicRun({
        batchId: state.batchId,
        bookIds: [state.bookId],
        presetId: usable.preset.id,
        runMode: DEFAULT_COMIC_RUN_MODE,
        aspectRatio: project.aspectRatio,
        stylePresetId: project.styleId,
        reference: project.reference,
      }, usable.preset);
      submittedScope.current = scope;
      resultLoaded.current = false;
      imageVersion.current += 1;
      // 提交成功才把指针标成 submitted：恢复时直接回到进度页并由服务端校准。
      rememberActiveRun(state.batchId, state.bookId, true, project);
      patch({ submitting: false, runSubmitted: true, stage: "render", tab: "result", status: null, pages: [], paths: {}, pageIndex: 0, progressError: "", resultError: "" });
    } catch (cause) {
      const message = comicNewErrorText(cause);
      patch({ submitting: false, error: message });
      void Taro.showToast({ title: message, icon: "none", duration: 2600 });
    } finally {
      runSubmitting.current = false;
    }
  };

  const startAnother = () => {
    stopPolling.current?.();
    clearComicActiveRun();
    assetScope.current = null;
    submittedScope.current = null;
    assetsLoadedFor.current = "";
    resultLoaded.current = false;
    imageVersion.current += 1;
    patch({
      stage: "source", tab: "create", batchId: "", bookId: "", status: null, pages: [], paths: {}, pageIndex: 0,
      referencePath: "", projectSnapshot: null, runSubmitted: false, error: "", progressError: "", resultError: "",
      assets: [], assetImages: {}, assetThumbs: {}, assetsLoading: false, assetBusy: "", assetError: "",
    });
  };

  const selectPage = (index: number) => {
    const page = state.pages[index];
    if (!page) return;
    patch({ pageIndex: index });
    void loadPageImages(page);
  };

  const current = state.pages[state.pageIndex];
  const primaryImage = current?.images.find((image) => image.id === current.primaryImageId);
  const stagePath = primaryImage ? state.paths[primaryImage.id] : "";
  const downloadable = state.pages.filter(page => page.primaryImageId && page.images.some(image => image.id === page.primaryImageId)).length;
  const failedBook = state.status?.books?.find((book) => String(book.status || "").toLowerCase() === "failed");

  // ---- 资产只读展示（结果 tab 顶部）----
  const presetConfig = (projectPreset?.preset.config || {}) as Record<string, unknown>;
  const presetTextModelId = typeof presetConfig.textModelId === "string" ? presetConfig.textModelId : "";
  const presetImageModelId = typeof presetConfig.imageModelId === "string" ? presetConfig.imageModelId : "";
  const reloadAssets = async (scope = assetScope.current) => {
    if (!scope || scope !== assetScope.current) return;
    assetsLoadedFor.current = "";
    await loadAssets(scope.batchId, scope.bookId, true, scope);
  };

  /** 重新提取资产：整本强制重提，文案口径与网页版「重新提取资产」一致。 */
  const runExtractAssets = async () => {
    const scope = assetScope.current;
    if (!scope || scope.batchId !== state.batchId || scope.bookId !== state.bookId) return;
    if (!state.batchId || !state.bookId || state.assetBusy) return;
    patch({ assetBusy: "extract", assetBusyAt: Date.now(), assetError: "" });
    try {
      await extractComicAssets({
        batchId: state.batchId, bookId: state.bookId, mode: "force",
        textModelId: presetTextModelId || undefined,
      });
      await reloadAssets(scope);
    } catch (cause) {
      if (scope !== assetScope.current) return;
      patch({ assetError: comicFailureText(cause, "资产提取失败，稍后再试") });
    } finally {
      if (scope === assetScope.current) patch({ assetBusy: "" });
    }
  };

  /** 补齐资产图：只挑还没有图的资产，异步任务复用 3s→10s 的退避轮询。 */
  const runMissingAssetImages = async (onlyAssetId?: string) => {
    const scope = assetScope.current;
    if (!scope || scope.batchId !== state.batchId || scope.bookId !== state.bookId) return;
    const project = state.projectSnapshot;
    const ids = state.assets.filter((asset) => onlyAssetId ? asset.id === onlyAssetId : !(state.assetImages[asset.id] || []).length).map((asset) => asset.id);
    if (!project || !ids.length || !state.batchId || state.assetBusy) return;
    patch({ assetBusy: "images", assetBusyAt: Date.now(), assetError: "" });
    let pendingTask = false;
    try {
      const result = await generateComicAssetImages({
        batchId: state.batchId, bookId: state.bookId, assetIds: ids,
        modelId: presetImageModelId || undefined, imageAspectRatio: project.aspectRatio,
      });
      if (scope !== assetScope.current) return;
      if (!result.images.length && result.taskId) {
        pendingTask = true;
        let stopImages: (() => void) | undefined;
        stopImages = startComicPolling(
          () => readComicAssetImageTask(state.batchId, state.bookId, result.taskId),
          (value) => {
            if (scope !== assetScope.current) { stopImages?.(); return; }
            const status = String((value as { status?: unknown } | null)?.status || "");
            if (!status || status === "processing" || status === "pending") return;
            stopImages?.();
            pendingTask = false;
            patch({ assetBusy: "" });
            if (status === "failed") {
              const message = String((value as { error?: unknown } | null)?.error || "");
              patch({ assetError: message || "资产图生成失败，可再补齐一次" });
            }
            void reloadAssets(scope);
          },
          (cause) => { if (scope === assetScope.current) patch({ assetError: comicFailureText(cause, "资产图生成失败，可再补齐一次") }); },
        );
      }
      await reloadAssets(scope);
    } catch (cause) {
      if (scope !== assetScope.current) return;
      patch({ assetError: comicFailureText(cause, "资产图生成失败，可再补齐一次") });
    } finally {
      if (!pendingTask && scope === assetScope.current) patch({ assetBusy: "" });
    }
  };

  /** 重试这一步：网页版「重试漫画资产」用的就是 stages/retry（只重跑最后失败阶段）。 */
  const runRetryStage = async () => {
    const scope = assetScope.current;
    if (!scope || scope.batchId !== state.batchId || scope.bookId !== state.bookId) return;
    if (!state.batchId || !state.bookId || state.assetBusy) return;
    patch({ assetBusy: "retry", assetBusyAt: Date.now(), assetError: "", pollToken: state.pollToken + 1, status: null });
    try {
      await retryComicBookStage({
        batchId: state.batchId, bookId: state.bookId, textModelId: presetTextModelId || undefined,
      });
      await reloadAssets(scope);
    } catch (cause) {
      if (scope !== assetScope.current) return;
      patch({ assetError: comicFailureText(cause, "重试失败，稍后再试") });
    } finally {
      if (scope === assetScope.current) patch({ assetBusy: "" });
    }
  };

  const preview = async (imageId: string) => {
    if (!current) return;
    const version = imageVersion.current;
    patch({ resultError: "" });
    try {
      const loaded = { ...state.paths, ...(await loadComicImagePaths(current)) };
      if (version !== imageVersion.current) return;
      setState((prev) => ({ ...prev, paths: { ...prev.paths, ...loaded } }));
      const ids = current.images.map((image) => image.id).filter((id) => loaded[id]);
      if (!ids.length) throw new Error("这张图还没准备好，稍等一下再点");
      const request = buildPreviewRequest(ids.map((id) => loaded[id]), Math.max(0, ids.indexOf(imageId)));
      await Taro.previewImage({ urls: request.urls, current: request.current });
    } catch (cause: any) {
      if (version !== imageVersion.current) return;
      patch({ resultError: cause?.message || "暂时打不开大图，请稍后再试" });
    }
  };

  const choosePrimaryImage = async (imageId: string) => {
    if (!current || current.primaryImageId === imageId || state.busy) return;
    const scopeVersion = imageVersion.current;
    const pageId = current.pageId;
    patch({ busy: "primary", resultError: "" });
    try {
      await setPrimaryComicPageImage(state.batchId, state.bookId, current.pageId, imageId);
      if (scopeVersion !== imageVersion.current) return;
      setState((prev) => ({
        ...prev,
        busy: "",
        pages: prev.pages.map((page) => page.pageId === pageId ? {
          ...page,
          primaryImageId: imageId,
          images: page.images.map((image) => ({ ...image, isPrimary: image.id === imageId })),
        } : page),
      }));
      await Taro.showToast({ title: "已设为本页主图", icon: "success" });
    } catch (cause: any) {
      if (scopeVersion !== imageVersion.current) return;
      patch({ busy: "", resultError: cause?.message || "切换主图失败，请稍后再试" });
    }
  };

  const regenerateCurrentPage = async () => {
    if (!current || state.busy) return;
    const scopeVersion = imageVersion.current;
    const config = presetConfig;
    const modelId = typeof config.imageModelId === "string" ? config.imageModelId : "";
    if (!modelId) { patch({ resultError: "当前预设没有图片模型，不能重新生成" }); return; }
    const choice = await Taro.showModal({
      title: "重新生成这一页？",
      content: "会新增一张候选图，并重新消耗一次出图额度；原来的图片不会删除。",
      confirmText: "重新生成",
      cancelText: "先不生成",
    });
    if (!choice?.confirm || scopeVersion !== imageVersion.current) return;
    patch({ busy: "regenerate", resultError: "" });
    try {
      await regenerateComicPage(state.batchId, state.bookId, current.videoId, {
        comicPrompt: current.comicPrompt,
        imagePrompt: current.imagePrompt || current.comicPrompt,
        modelId,
        ...(typeof config.imageProvider === "string" && config.imageProvider ? { imageProvider: config.imageProvider } : {}),
        aspectRatio: current.aspectRatio || state.projectSnapshot?.aspectRatio || state.aspectRatio,
      });
      if (scopeVersion !== imageVersion.current) return;
      resultLoaded.current = false;
      await loadPages(state.batchId, state.bookId, true);
      if (scopeVersion === imageVersion.current) await Taro.showToast({ title: "已提交并刷新页图", icon: "none" });
    } catch (cause: any) {
      if (scopeVersion !== imageVersion.current) return;
      patch({ resultError: cause?.message || "这一页重新生成失败，请稍后再试" });
    } finally {
      if (scopeVersion === imageVersion.current) patch({ busy: "" });
    }
  };

  const savePrimary = async () => {
    if (!current) return;
    const image = current.images.find((item) => item.id === current.primaryImageId);
    if (!image) { patch({ resultError: "这一页还没有可以保存的图" }); return; }
    patch({ busy: "save", resultError: "" });
    try {
      const loaded = state.paths[image.id] ? state.paths : await loadComicImagePaths(current);
      const filePath = loaded[image.id];
      if (!filePath) throw new Error("这张图还没准备好，稍等一下再试");
      await saveComicImageToAlbum(filePath);
      await Taro.showToast({ title: "已保存到相册", icon: "success" });
    } catch (cause: any) {
      const message = String(cause?.message || "保存失败");
      if (/权限/.test(message)) {
        const choice = await Taro.showModal({ title: "需要相册权限", content: message, confirmText: "去设置", cancelText: "取消" });
        if (choice.confirm) void Taro.openSetting({});
      } else {
        patch({ resultError: message });
      }
    } finally {
      patch({ busy: "" });
    }
  };

  const downloadZip = async () => {
    if (state.busy) return;
    const videoIds = state.pages.filter(page => page.primaryImageId && page.images.some(image => image.id === page.primaryImageId)).map(page => page.videoId);
    if (!videoIds.length) { patch({ resultError: "还没有可以打包的图，等生成完成再来" }); return; }
    let kind: "images" | "prompts" | "douyin" = "images";
    try {
      const selected = await Taro.showActionSheet({ itemList: ["导出全部漫画图片", "导出提示词与图片 ZIP", "导出抖音图文包"] });
      kind = (["images", "prompts", "douyin"] as const)[selected.tapIndex] || "images";
    } catch { return; }
    const scopeVersion = imageVersion.current;
    patch({ busy: "zip", resultError: "" });
    try {
      const collection = kind === "images" ? await createComicCollection(state.batchId, state.bookId, videoIds) : null;
      const file = collection ? await downloadComicCollectionZip(state.batchId, state.bookId, collection.id) : await downloadComicPagePackage(state.batchId, state.bookId, kind as "prompts" | "douyin");
      if (scopeVersion !== imageVersion.current) return;
      const choice = await Taro.showModal({
        title: "已打包下载",
        content: `已保存到小程序文件。${collection ? `共 ${collection.items.length} 页。` : ""}可复制下载链接，在已登录的浏览器打开。`,
        confirmText: "知道了",
        cancelText: "复制链接",
      });
      if (choice.cancel) await Taro.setClipboardData({ data: file.url });
    } catch (cause: any) {
      if (scopeVersion !== imageVersion.current) return;
      patch({ resultError: cause?.message || "打包下载失败，请稍后再试" });
    } finally {
      if (scopeVersion === imageVersion.current) patch({ busy: "" });
    }
  };

  const canSelectStage = (stage: ComicStage) => {
    const mapped = stage === "source" ? "settings" : stage === "storyboard" ? "config" : stage;
    // 资产是否齐、是否正在提取都不再拦导航：只挡提交中的重复点击。
    return !runSubmitting.current && !state.submitting && canEnterPrivateComicStage(mapped, {
      sourceLength: state.projectSnapshot?.sourceLength || length,
      minSourceLength: MIN_COMIC_SOURCE_CHARS,
      styleId: state.projectSnapshot?.styleId || state.styleId,
      batchId: state.bookId ? state.batchId : "",
      assetsCount: state.assets.length, pagesCount: state.pages.filter((page) => page.images.length > 0 && Boolean(page.primaryImageId)).length, submitted: state.runSubmitted,
    });
  };
  const selectStage = (stage: ComicStage) => {
    if (!canSelectStage(stage)) return;
    patch({ stage, tab: stage === "source" ? "create" : "result" });
  };


  const content = state.stage === "source" ? "source" : state.stage === "assets" || state.stage === "storyboard" ? "assets" : "pages";
  const title = state.source.trim().split(/\\r?\\n/)[0]?.slice(0, 22) || "未命名漫画";
  return <ThemedPage dynamicGlass={false} className={`mp-page comic-page comic-new comic-editor is-${content}${state.stage === "storyboard" ? " is-confirm" : ""}`}>
    <BrandHeader title="漫画生成" showBack />
    <View className="ce-project-bar"><View><Text className="ce-project-name">{title}</Text><Text className="ce-muted">一战晟铭 · 私人定制</Text></View>{finished || (state.stage === "source" && state.batchId) ? <Button disabled={running || Boolean(state.busy)} onClick={startAnother}>再做一本</Button> : null}<Button disabled={!downloadable || Boolean(state.busy)} onClick={() => { void downloadZip(); }}>导出</Button></View>
    <ComicStageHeader compact stage={state.stage} title="漫画生成" subtitle="" hasResult={Boolean(state.pages.length)} canSelectStage={canSelectStage} onSelectStage={selectStage} />
    <View className="ce-workspace-body">
      {state.tab === "create" ? <ComicSettingsStage
        source={state.source} styleId={state.styleId} aspectRatio={state.aspectRatio}
        referencePath={state.referencePath} styles={state.styles} thumbs={state.thumbs}
        preset={effectivePreset} presets={state.presets} selectedPresetId={effectivePreset?.preset.id || ""} advancedOpen={state.advancedOpen}
        settingCatalog={state.settingCatalog} models={state.models} settingSelections={state.settingSelections}
        settingPicker={state.settingPicker}
        thumbFailed={state.thumbFailed} onRetryThumb={retryStyleThumb}
        setupLoading={state.setupLoading} error={state.error}
        onSourceChange={(source) => patch({ source })}
        onPasteSource={() => { void pasteSource(); }}
        onChooseExisting={() => { void Taro.navigateTo({ url: "/comic/pages/comic-config/index?mode=pick" }); }}
        onStyleChange={(styleId) => patch({ styleId })}
        onAspectChange={(aspectRatio) => patch({ aspectRatio })}
        onPresetChange={(selectedPresetId) => patch({ selectedPresetId })}
        onSettingSelectionChange={(key, value) => setState(prev => ({...prev, settingSelections: {...prev.settingSelections, [key]: value}}))}
        onConfirmSettingSelections={(values) => setState(prev => ({ ...prev, settingSelections: { ...prev.settingSelections, ...values } }))}
        onToggleAdvanced={() => patch({ advancedOpen: !state.advancedOpen })}
        onOpenSettingPicker={(settingPicker) => patch({ settingPicker })}
        onCloseSettingPicker={() => patch({ settingPicker: "" })}
        onOpenReference={() => patch({ referenceOpen: true })}
        onContinue={() => { void prepareProject(); }}
        continueDisabled={!canSubmit} continueLoading={state.submitting}
        continueHint={blockReason} configurationHint={configurationHint}
      /> : null}


      {state.tab === "result" ? <>
        {state.stage === "assets" && state.batchId ? <ComicAssetsStage key={state.batchId + ":" + state.bookId} batchId={state.batchId} bookId={state.bookId} onGenerateImage={(id) => { void runMissingAssetImages(id); }}
          assets={state.assets} assetImages={state.assetImages} assetThumbs={state.assetThumbs}
          loading={state.assetsLoading} busy={state.assetBusy} busyAt={state.assetBusyAt} error={state.assetError}
          serverExtracting={!state.assets.length && runView.stage === "comic_assets"}
          imagesSkipped={assetImagesSkipped(state.status)}
          onExtract={() => { void runExtractAssets(); }}
          onGenerateMissingImages={() => { void runMissingAssetImages(); }}
          onRetry={() => { void runRetryStage(); }} onRefresh={() => { void reloadAssets(); }}
          onContinue={() => patch({ stage: "storyboard" })}
          onBack={() => patch({ stage: "source", tab: "create" })}
        /> : null}


        {state.stage === "storyboard" && state.batchId && state.projectSnapshot ? <ComicConfigConfirmStage
          sourceLength={state.projectSnapshot.sourceLength} styleName={state.styles.find((style) => style.id === state.projectSnapshot?.styleId)?.name || ""}
          aspectRatio={state.projectSnapshot.aspectRatio} referenceReady={Boolean(state.projectSnapshot.reference)} assets={state.assets}
          textModelName={state.models.find(model => model.kind === "text" && model.id === presetTextModelId)?.name || (presetTextModelId ? "已保存的文本模型" : "未配置")}
          imageModelName={state.models.find(model => model.kind === "image" && model.id === presetImageModelId)?.name || (presetImageModelId ? "已保存的图片模型" : "未配置")}
          submitting={state.submitting} submitted={state.runSubmitted} error={state.error}
          onBack={() => { if (!runSubmitting.current) patch({ stage: "assets" }); }}
          onConfirm={() => { void confirmConfigAndStart(); }}
        /> : null}


        {state.stage === "render" || state.stage === "result" ? <>
          <View className="ce-run-status"><Text>{finished ? progress.succeeded ? "本次生成已结束" : "生成未完成，请检查错误" : state.runSubmitted ? runView.label || "正在读取生成状态…" : "尚未提交生成"}</Text>{progress.failed || progress.interrupted ? <Button disabled={Boolean(state.assetBusy)} onClick={() => { void runRetryStage(); }}>重试失败阶段</Button> : null}</View>
          {state.restored ? <Text className="comic-hint">已恢复上次进度，正在向服务端校准…</Text> : null}
          {progress.failed ? <Text className="mp-error">{failedBook?.error || "生成未完成，可重试失败阶段"}</Text> : null}
          {state.progressError || state.error ? <Text className="mp-error">{state.progressError || state.error}</Text> : null}
          {state.resultLoading ? <Text className="comic-hint">正在打开页图…</Text> : null}
          <ComicPageWorkspace key={state.batchId + ":" + state.bookId} batchId={state.batchId} bookId={state.bookId} pages={state.pages} pageIndex={state.pageIndex} paths={state.paths} generating={running} busy={state.busy} error={state.resultError}
            onSelect={selectPage} onPreview={(id) => { void preview(id); }} onRegenerate={() => { void regenerateCurrentPage(); }}
            onPrimary={(id) => { void choosePrimaryImage(id); }} onSave={() => { void savePrimary(); }} onExport={() => { void downloadZip(); }} onReload={() => loadPages(state.batchId,state.bookId,true)} />
          {!state.pages.length ? <View className="ce-empty"><Text>{state.resultError || (running ? "漫画页图生成后会出现在这里，离开后可以返回查看。" : "暂时没有读取到页图")}</Text><Button onClick={() => { void loadPages(state.batchId,state.bookId,true); }}>重新读取页图</Button></View> : null}
        </> : null}
      </> : null}
    </View>
      {state.referenceOpen ? <>
        <View className="comic-sheet-mask" onClick={() => patch({ referenceOpen: false })} />
        <View className="comic-sheet">
          <View className="comic-sheet-handle" />
          <View className="comic-sheet-header">
            <Text>参考图</Text>
            <Text className="comic-sheet-close" onClick={() => patch({ referenceOpen: false })}>完成</Text>
          </View>
          <View className="comic-sheet-body">
            <Text className="comic-hint">选一张和想要画风接近的图，出图会照着它的调子走。</Text>
            {state.referencePath ? <Image src={state.referencePath} mode="aspectFill" className="comic-reference" onClick={() => { void chooseReference(); }} /> : null}
            <View className="comic-sheet-actions">
              <Button className="mp-secondary" onClick={() => { void chooseReference(); }}>{state.referencePath ? "换一张" : "选一张图"}</Button>
              {state.referencePath ? <Button className="mp-secondary" onClick={() => patch({ referencePath: "" })}>去掉</Button> : null}
            </View>
          </View>
        </View>
      </> : null}


  </ThemedPage>;
}
