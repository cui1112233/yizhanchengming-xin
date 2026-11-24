import { apiRequest } from "../../services/api";
import { BATCH_FACTORY_BASE } from "../../services/batchFactoryV11";
import { loadAuthenticatedMedia } from "../../services/mediaCache";

/**
 * 漫画资产（人物 / 场景 / 道具）只读展示与补齐。
 *
 * 接口与网页版 PrivateComicWorkspace 实际在用的完全一致，未新增后端：
 *   GET  {base}/batches/{batch}/books/{book}/assets                    -> { assets: BookAsset[] }
 *        BookAsset 字段：id / kind(character|scene|prop) / name / formalName / facts / appearance / prompt
 *   GET  .../assets/{asset}/images                                     -> { images: BookAssetImage[] }
 *        BookAssetImage 字段：id / url / mediaType / isPrimary；url 即
 *        `/api/batch-factory/v12/batches/{batch}/books/{book}/assets/{asset}/images/{image}/content`
 *   POST .../books/{book}/stages/assets               { requestId, mode:"missing"|"force", textModelId }
 *   POST .../books/{book}/assets/images/generate      { assetIds, modelId, imageAspectRatio }
 *   GET  .../books/{book}/assets/images/generate/{taskId}                -> { status, error }
 *   POST .../books/{book}/stages/retry                { requestId, textModelId }    （网页版「重试漫画资产」入口）
 */

export type ComicAssetItem = {
  id: string;
  kind: string;
  name: string;
  formalName: string;
  facts: string;
  appearance: string;
  prompt: string;
  revision?: number;
  imageAspectRatio?: string;
  imageModelId?: string;
};

export type ComicAssetImageItem = {
  id: string;
  url: string;
  mediaType: string;
  isPrimary: boolean;
};

export const COMIC_ASSET_KIND_LABELS: Record<string, string> = {
  character: "人物",
  scene: "场景",
  prop: "道具",
};

function text(value: unknown) {
  return typeof value === "string" && value.trim() ? value.trim() : "";
}

function object(value: unknown): Record<string, unknown> {
  return value && typeof value === "object" && !Array.isArray(value) ? value as Record<string, unknown> : {};
}

function arrayFrom(value: unknown, key: string) {
  const candidate = object(value)[key];
  return Array.isArray(candidate) ? candidate : [];
}

function bookPath(batchId: string, bookId: string, suffix: string) {
  return `${BATCH_FACTORY_BASE}/batches/${encodeURIComponent(batchId)}/books/${encodeURIComponent(bookId)}/${suffix}`;
}

export function normalizeComicAssets(payload: unknown): ComicAssetItem[] {
  return arrayFrom(payload, "assets").map((raw) => {
    const asset = object(raw);
    return {
      id: text(asset.id),
      kind: text(asset.kind),
      name: text(asset.name),
      formalName: text(asset.formalName),
      facts: text(asset.facts),
      appearance: text(asset.appearance),
      prompt: text(asset.prompt),
      revision: Number(asset.revision || 0),
      imageAspectRatio: text(asset.imageAspectRatio),
      imageModelId: text(asset.imageModelId),
    };
  }).filter((asset) => asset.id);
}

export function normalizeComicAssetImages(payload: unknown): ComicAssetImageItem[] {
  return arrayFrom(payload, "images").map((raw) => {
    const image = object(raw);
    return {
      id: text(image.id),
      url: text(image.url),
      mediaType: text(image.mediaType) || "image/png",
      isPrimary: image.isPrimary === true,
    };
  }).filter((image) => image.id && image.url);
}

/** 名字 + 一句外形描述（网页版列表只有名字，弹层这里补一句 appearance）。 */
export function comicAssetSentence(asset: ComicAssetItem) {
  return asset.appearance || asset.facts || asset.formalName || asset.prompt;
}

export function comicAssetKindCounts(assets: ComicAssetItem[]) {
  return {
    character: assets.filter((asset) => asset.kind === "character").length,
    scene: assets.filter((asset) => asset.kind === "scene").length,
    prop: assets.filter((asset) => asset.kind === "prop").length,
  };
}

/** 四步流程条，文案与网页版 pipelineSteps 一致，最后一步按用户口径显示「出图」。 */
export function comicAssetSteps(input: {
  assetCount: number;
  assetImageCount: number;
  promptReadyCount: number;
  pagesTotal: number;
  imageReadyCount: number;
}) {
  const { assetCount, assetImageCount, promptReadyCount, pagesTotal, imageReadyCount } = input;
  return [
    { key: "assets", label: "提取资产", done: assetCount > 0, detail: assetCount > 0 ? `已完成 ${assetCount} 项` : "未提取" },
    { key: "assetImages", label: "资产图", done: assetCount > 0 && assetImageCount >= assetCount, detail: assetCount > 0 ? `已生成 ${assetImageCount}/${assetCount}` : "待资产" },
    { key: "prompts", label: "分镜提示词", done: promptReadyCount > 0, detail: pagesTotal ? `已生成 ${promptReadyCount}/${pagesTotal} 页` : assetCount > 0 ? "待生成" : "待资产" },
    { key: "images", label: "出图", done: pagesTotal > 0 && imageReadyCount >= pagesTotal, detail: pagesTotal ? `已出图 ${imageReadyCount}/${pagesTotal}` : "待提示词" },
  ];
}

export async function loadComicAssets(batchId: string, bookId: string): Promise<ComicAssetItem[]> {
  if (!batchId || !bookId) throw new Error("缺少批次或书目编号");
  const payload = await apiRequest<unknown>(bookPath(batchId, bookId, "assets"), { timeout: 30000 });
  return normalizeComicAssets(payload);
}

export async function loadComicAssetImages(batchId: string, bookId: string, assetId: string): Promise<ComicAssetImageItem[]> {
  if (!batchId || !bookId || !assetId) return [];
  const payload = await apiRequest<unknown>(
    bookPath(batchId, bookId, `assets/${encodeURIComponent(assetId)}/images`),
    { timeout: 30000 },
  );
  return normalizeComicAssetImages(payload);
}

/**
 * 图片一律走 mediaCache 的 request 通道（loadAuthenticatedMedia），不再试探
 * downloadFile：资产图接口受鉴权保护且主域不在 downloadFile 白名单。
 * 单张失败返回空串，由页面静默显示「未生成」，不产生红色错误层。
 */
export async function loadComicAssetImagePath(image: ComicAssetImageItem): Promise<string> {
  try {
    const media = await loadAuthenticatedMedia(image.url, "png");
    return media.path || "";
  } catch {
    return "";
  }
}

let stageRequestSeq = 0;

/** 后端要求 requestId 非空，且同一阶段每次尝试要能区分，所以拼进程内自增序号。 */
function nextStageRequestId() {
  stageRequestSeq += 1;
  return `miniprogram-${Date.now()}-${stageRequestSeq}`;
}

export async function extractComicAssets(input: {
  batchId: string;
  bookId: string;
  mode: "missing" | "force";
  textModelId?: string;
}) {
  const requestBody = {
    requestId: nextStageRequestId(),
    mode: input.mode,
    ...(input.textModelId ? { textModelId: input.textModelId } : {}),
  };
  try {
    await apiRequest(bookPath(input.batchId, input.bookId, "stages/assets"), {
      method: "POST",
      timeout: 180000,
      data: requestBody,
    });
  } catch (cause) {
    const failure = cause as { statusCode?: number; body?: unknown; message?: string } | null;
    // 只留定位用的最小信息：请求体/响应体不落控制台（响应体已由 comicFailureText 按开关决定是否展示）。
    console.error("[extractComicAssets] 资产提取失败", {
      batchId: input.batchId,
      bookId: input.bookId,
      url: bookPath(input.batchId, input.bookId, "stages/assets"),
      statusCode: failure?.statusCode ?? null,
      message: failure?.message ?? String(cause),
    });
    throw cause;
  }
}

export async function generateComicAssetImages(input: {
  batchId: string;
  bookId: string;
  assetIds: string[];
  modelId?: string;
  imageAspectRatio?: string;
}) {
  const payload = await apiRequest<unknown>(bookPath(input.batchId, input.bookId, "assets/images/generate"), {
    method: "POST",
    timeout: 180000,
    data: {
      assetIds: input.assetIds,
      ...(input.modelId ? { modelId: input.modelId } : {}),
      ...(input.imageAspectRatio ? { imageAspectRatio: input.imageAspectRatio } : {}),
    },
  });
  const source = object(payload);
  return {
    images: Array.isArray(source.images) ? source.images : [],
    taskId: text(source.taskId),
    warning: text(source.warning),
  };
}

export async function readComicAssetImageTask(batchId: string, bookId: string, taskId: string) {
  return apiRequest<unknown>(
    bookPath(batchId, bookId, `assets/images/generate/${encodeURIComponent(taskId)}`),
    { timeout: 30000 },
  );
}

/** 网页版「重试漫画资产」：只重跑这本书最后失败的阶段。 */
export async function retryComicBookStage(input: { batchId: string; bookId: string; textModelId?: string }) {
  const requestBody = {
    requestId: nextStageRequestId(),
    ...(input.textModelId ? { textModelId: input.textModelId } : {}),
  };
  try {
    await apiRequest(bookPath(input.batchId, input.bookId, "stages/retry"), {
      method: "POST",
      timeout: 180000,
      data: requestBody,
    });
  } catch (cause) {
    const failure = cause as { statusCode?: number; body?: unknown; message?: string } | null;
    console.error("[retryComicBookStage] 重试失败", {
      batchId: input.batchId,
      bookId: input.bookId,
      url: bookPath(input.batchId, input.bookId, "stages/retry"),
      statusCode: failure?.statusCode ?? null,
      message: failure?.message ?? String(cause),
    });
    throw cause;
  }
}
