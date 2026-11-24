import Taro from "@tarojs/taro";
import { apiRequest } from "../../services/api";
import { apiUrl } from "../../config/runtime";
import { BATCH_FACTORY_BASE } from "../../services/batchFactoryV11";
import { loadAuthenticatedMedia } from "../../services/mediaCache";

/** 微信 previewImage 单次最多 9 张，超出需要分批。 */
export const COMIC_PREVIEW_LIMIT = 9;

export type ComicImageItem = {
  id: string;
  url: string;
  isPrimary: boolean;
  mediaType: string;
};

export type ComicPageItem = {
  ordinal: number;
  pageId: string;
  videoId: string;
  status: string;
  aspectRatio: string;
  comicPrompt: string;
  imagePrompt: string;
  images: ComicImageItem[];
  primaryImageId: string;
};

export type ComicCollectionItem = {
  ordinal: number;
  videoId: string;
  comicPageId: string;
  imageId: string;
  url: string;
};

export type ComicCollectionInfo = {
  id: string;
  bookId: string;
  status: string;
  items: ComicCollectionItem[];
};

function text(value: unknown) {
  return typeof value === "string" && value.trim() ? value.trim() : "";
}

function object(value: unknown): Record<string, unknown> {
  return value && typeof value === "object" && !Array.isArray(value) ? value as Record<string, unknown> : {};
}

function arrayFrom(value: unknown, key: string) {
  const source = object(value);
  const candidate = source[key];
  return Array.isArray(candidate) ? candidate : [];
}

function errorText(cause: unknown) {
  const source = cause && typeof cause === "object" ? cause as { errMsg?: unknown; message?: unknown } : {};
  return typeof source.errMsg === "string" ? source.errMsg : typeof source.message === "string" ? source.message : "";
}

/** comic-pages 响应无 ordinal 字段，服务端已按视频 ordinal 排序，这里按数组顺序编号。 */
export function normalizeComicPages(payload: unknown): ComicPageItem[] {
  return arrayFrom(payload, "pages").map((item, index) => {
    const page = object(item);
    const images = (Array.isArray(page.images) ? page.images as unknown[] : [])
      .map((image) => object(image))
      .filter((image) => text(image.id) && text(image.url))
      .map((image) => ({
        id: text(image.id),
        url: text(image.url),
        isPrimary: image.isPrimary === true,
        mediaType: text(image.mediaType) || "image/png",
      }));
    const primary = images.find((image) => image.isPrimary);
    return {
      ordinal: index + 1,
      pageId: text(page.id),
      videoId: text(page.videoId),
      status: text(page.status) || "状态未返回",
      aspectRatio: text(page.aspectRatio),
      comicPrompt: text(page.comicPrompt),
      imagePrompt: text(page.imagePrompt) || text(page.comicImagePrompt),
      images,
      primaryImageId: primary?.id || "",
    };
  });
}

export function normalizeComicCollections(payload: unknown): ComicCollectionInfo[] {
  return arrayFrom(payload, "collections").map((item) => {
    const collection = object(item);
    const items = (Array.isArray(collection.items) ? collection.items as unknown[] : []).map((raw) => {
      const entry = object(raw);
      return {
        ordinal: Number(entry.ordinal) || 0,
        videoId: text(entry.videoId),
        comicPageId: text(entry.comicPageId),
        imageId: text(entry.imageId),
        url: text(entry.url),
      };
    }).filter((entry) => entry.imageId).sort((a, b) => a.ordinal - b.ordinal);
    return {
      id: text(collection.id),
      bookId: text(collection.bookId),
      status: text(collection.status),
      items,
    };
  }).filter((collection) => collection.id);
}

export function imageExtension(mediaType: string, url = "") {
  const mime = String(mediaType || "").toLowerCase();
  if (mime.includes("png")) return "png";
  if (mime.includes("webp")) return "webp";
  if (mime.includes("gif")) return "gif";
  if (mime.includes("jpeg") || mime.includes("jpg")) return "jpg";
  const fromUrl = String(url || "").split("?")[0].split(".").pop()?.toLowerCase() || "";
  return ["png", "jpg", "jpeg", "webp", "gif"].includes(fromUrl) ? fromUrl : "png";
}

export async function loadComicPages(batchId: string, bookId: string): Promise<ComicPageItem[]> {
  if (!batchId || !bookId) throw new Error("缺少批次或书目编号");
  const payload = await apiRequest<unknown>(
    `${BATCH_FACTORY_BASE}/batches/${encodeURIComponent(batchId)}/books/${encodeURIComponent(bookId)}/comic-pages`,
    { timeout: 30000 },
  );
  return normalizeComicPages(payload);
}

/** 与网页版同一条接口：把候选图提升为当前页主图。 */
export async function setPrimaryComicPageImage(batchId: string, bookId: string, pageId: string, imageId: string) {
  await apiRequest(
    `${BATCH_FACTORY_BASE}/batches/${encodeURIComponent(batchId)}/books/${encodeURIComponent(bookId)}/comic-pages/${encodeURIComponent(pageId)}/images/${encodeURIComponent(imageId)}/primary`,
    { method: "PUT", data: {} },
  );
}

export type RegenerateComicPageInput = {
  comicPrompt: string;
  imagePrompt: string;
  modelId: string;
  imageProvider?: string;
  aspectRatio: string;
};

/** 单页重生会新增候选图，不删除原图；调用前由页面二次确认额度消耗。 */
export async function regenerateComicPage(batchId: string, bookId: string, videoId: string, input: RegenerateComicPageInput) {
  await apiRequest(
    `${BATCH_FACTORY_BASE}/batches/${encodeURIComponent(batchId)}/books/${encodeURIComponent(bookId)}/comic-pages/${encodeURIComponent(videoId)}/generate`,
    { method: "POST", timeout: 180000, data: input },
  );
}

export async function loadComicCollections(batchId: string, bookId: string): Promise<ComicCollectionInfo[]> {
  const payload = await apiRequest<unknown>(
    `${BATCH_FACTORY_BASE}/batches/${encodeURIComponent(batchId)}/books/${encodeURIComponent(bookId)}/comic-collections`,
    { timeout: 30000 },
  );
  return normalizeComicCollections(payload);
}

/**
 * 逐张下载到本地。图片内容接口受鉴权保护，必须走 downloadFile 链路；
 * 失败的图片保留空值，由页面显示占位，不阻塞其他候选图。
 */
export async function loadComicImagePaths(page: ComicPageItem): Promise<Record<string, string>> {
  const entries = await Promise.all(page.images.map(async (image) => {
    try {
      const media = await loadAuthenticatedMedia(image.url, imageExtension(image.mediaType, image.url));
      return [image.id, media.path] as const;
    } catch {
      return [image.id, ""] as const;
    }
  }));
  return Object.fromEntries(entries.filter(([, path]) => Boolean(path)));
}

/** 把预览窗口切到包含目标下标的 9 张以内批次，防止超过 WeChat 单次上限。 */
export function buildPreviewRequest(paths: string[], index: number, limit = COMIC_PREVIEW_LIMIT) {
  const size = Math.max(1, Math.min(COMIC_PREVIEW_LIMIT, Math.floor(limit) || 0));
  const safe = Math.max(0, Math.min(index, Math.max(0, paths.length - 1)));
  const start = Math.floor(safe / size) * size;
  const urls = paths.slice(start, start + size);
  return { urls, current: paths[safe] || urls[0] || "" };
}

export async function createComicCollection(batchId: string, bookId: string, videoIds: string[]): Promise<ComicCollectionInfo> {
  const payload = await apiRequest<unknown>(
    `${BATCH_FACTORY_BASE}/batches/${encodeURIComponent(batchId)}/books/${encodeURIComponent(bookId)}/comic-collections`,
    { method: "POST", timeout: 60000, data: { videoIds } },
  );
  const collection = normalizeComicCollections({ collections: [object(payload).collection] })[0];
  if (!collection) throw new Error("服务端未返回合集回执");
  return collection;
}

/** 保存单张图片到相册；权限被拒时抛出可引导用户去设置页的提示。 */
export async function saveComicImageToAlbum(filePath: string) {
  if (!filePath) throw new Error("图片尚未下载完成");
  try {
    await Taro.saveImageToPhotosAlbum({ filePath });
  } catch (cause) {
    const message = errorText(cause);
    if (/auth|deny|denied|permission|privacy/i.test(message)) {
      throw new Error("未获得相册权限，请在设置中开启“保存到相册”后重试");
    }
    throw new Error(message || "保存到相册失败");
  }
}

/**
 * 下载合集 ZIP。图片与压缩包统一在 mediaCache 里走 request 拉二进制（不再先试探
 * downloadFile，避免白名单没配主域时刷错误），随后尽力 saveFile 持久化；
 * saveFile 失败时仍返回沙箱内的路径可用。
 */
export async function downloadComicCollectionZip(batchId: string, bookId: string, collectionId: string) {
  const path = `${BATCH_FACTORY_BASE}/batches/${encodeURIComponent(batchId)}/books/${encodeURIComponent(bookId)}/comic-collections/${encodeURIComponent(collectionId)}/download`;
  const target = apiUrl(path);
  let media: { path: string };
  try {
    media = await loadAuthenticatedMedia(path, "zip");
  } catch (cause) {
    throw new Error(`合集下载失败：${cause instanceof Error ? cause.message : "请稍后重试"}`);
  }
  try {
    const saved = await Taro.saveFile({ tempFilePath: media.path });
    return { path: String((saved as { savedFilePath?: string }).savedFilePath || media.path), url: target };
  } catch {
    return { path: media.path, url: target };
  }
}

/** 网页同款提示词包/图文包；只下载，不自动发布。 */
export async function downloadComicPagePackage(batchId: string, bookId: string, kind: 'prompts' | 'douyin') {
  if (!batchId || !bookId) throw new Error('请先打开漫画项目');
  const suffix = kind === 'prompts' ? 'download-zip' : 'douyin-zip';
  const path = `${BATCH_FACTORY_BASE}/batches/${encodeURIComponent(batchId)}/books/${encodeURIComponent(bookId)}/comic-pages/${suffix}`;
  const media = await loadAuthenticatedMedia(path, 'zip');
  try {
    const saved = await Taro.saveFile({ tempFilePath: media.path });
    return { path: ('savedFilePath' in saved && saved.savedFilePath) || media.path, url: apiUrl(path) };
  } catch {
    return { path: media.path, url: apiUrl(path) };
  }
}
