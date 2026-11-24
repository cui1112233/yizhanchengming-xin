import Taro from "@tarojs/taro";

/**
 * 进行中的私人定制漫画在本地留一份「活动任务」指针，用于小程序被系统回收、
 * 或用户退出再进入时恢复批次与服务端进度。
 *
 * 能恢复的是「服务端已经存在的东西」：batchId / bookId，以及提交状态。
 * 恢复不了的是「进行中的动作句柄」：退出瞬间正在跑的提取 / 补齐资产图 / 重试，
 * 客户端没有可续接的 task 句柄，只能回前台后按服务端结果手动重试。界面必须如实说明。
 */
export const COMIC_ACTIVE_RUN_KEY = "comic:active-run";
export const COMIC_ACTIVE_RUN_TTL_MS = 24 * 60 * 60 * 1000;

export type ComicActiveRun = {
  version: 1;
  mode?: "manual";
  phase?: "assets" | "pages";
  pendingImages?: Record<string,string[]>;
  batchId: string;
  bookId: string;
  source: string;
  sourceLength: number;
  styleId: string;
  presetId: string;
  /** 创建项目时已经冻结的完整预设配置，防止返回设置后修改下一本草稿影响当前项目。 */
  presetConfig: Record<string, unknown>;
  aspectRatio: string;
  reference: { assetId: string; url: string; previewUrl?: string } | null;
  /** 是否已经调过 automation/start；未提交时恢复回资产阶段。 */
  submitted: boolean;
  /** 创建/提交时间，用于 24h 过期判断。 */
  startedAt: number;
};

function text(value: unknown) {
  return typeof value === "string" && value.trim() ? value.trim() : "";
}

function normalize(raw: unknown): ComicActiveRun | null {
  const value = typeof raw === "string"
    ? (() => { try { return JSON.parse(raw) as unknown; } catch { return null; } })()
    : raw;
  if (!value || typeof value !== "object" || Array.isArray(value)) return null;
  const record = value as Record<string, unknown>;
  const batchId = text(record.batchId);
  const bookId = text(record.bookId);
  const startedAt = Number(record.startedAt);
  if (!batchId || !bookId || !Number.isFinite(startedAt) || startedAt <= 0) return null;
  const reference = record.reference && typeof record.reference === "object"
    ? record.reference as Record<string, unknown>
    : null;
  const assetId = text(reference?.assetId);
  const url = text(reference?.url);
  return {
    version: 1,
    ...(record.mode === "manual" ? { mode: "manual" as const, phase: record.phase === "pages" ? "pages" as const : "assets" as const,
      pendingImages: record.pendingImages && typeof record.pendingImages === "object" && !Array.isArray(record.pendingImages)
        ? Object.fromEntries(Object.entries(record.pendingImages).filter(([key,value])=>Boolean(key)&&Array.isArray(value)&&value.every(id=>typeof id === "string"))) as Record<string,string[]> : {} } : {}),
    batchId,
    bookId,
    source: typeof record.source === "string" ? record.source : "",
    sourceLength: Number.isFinite(Number(record.sourceLength)) ? Math.max(0, Number(record.sourceLength)) : 0,
    styleId: text(record.styleId),
    presetId: text(record.presetId),
    presetConfig: record.presetConfig && typeof record.presetConfig === "object" && !Array.isArray(record.presetConfig)
      ? record.presetConfig as Record<string, unknown>
      : {},
    aspectRatio: text(record.aspectRatio),
    reference: assetId && url ? { assetId, url, ...(text(reference?.previewUrl) ? { previewUrl: text(reference?.previewUrl) } : {}) } : null,
    submitted: record.submitted === true,
    startedAt,
  };
}

/** 删除本地指针；任何一步失败都不影响页面继续。 */
export function clearComicActiveRun() {
  try { Taro.removeStorageSync(COMIC_ACTIVE_RUN_KEY); } catch { /* 存储不可用就当作没有指针。 */ }
}

export function saveComicActiveRun(run: Omit<ComicActiveRun, "version">) {
  try { Taro.setStorageSync(COMIC_ACTIVE_RUN_KEY, { ...run, version: 1 }); } catch { /* 存储失败不阻断生成。 */ }
}

/** 读本地指针：损坏或超过 24h 直接清掉，返回 null。 */
export function readComicActiveRun(now = Date.now()): ComicActiveRun | null {
  let raw: unknown = "";
  try { raw = Taro.getStorageSync(COMIC_ACTIVE_RUN_KEY); } catch { return null; }
  const run = normalize(raw);
  if (!run) {
    if (raw) clearComicActiveRun();
    return null;
  }
  if (now - run.startedAt > COMIC_ACTIVE_RUN_TTL_MS) {
    clearComicActiveRun();
    return null;
  }
  return run;
}

function statusOf(cause: unknown) {
  return Number((cause as { statusCode?: number } | null)?.statusCode || 0);
}

/** 401 / 登录失效：本地指针指向的批次已经不可访问。 */
export function isComicRunSessionLost(cause: unknown) {
  const status = statusOf(cause);
  const message = text((cause as { message?: string } | null)?.message);
  return status === 401 || /unauthorized|401|登录已失效|重新登录/i.test(message);
}

/** 404 / 批次不存在：服务端已经不认这条任务，本地指针必须清掉。 */
export function isComicRunGone(cause: unknown) {
  const status = statusOf(cause);
  const message = text((cause as { message?: string } | null)?.message);
  return status === 404 || /not found|不存在|已删除/i.test(message);
}

/** 这两类失败都要清 key；其它网络抖动保留，下一帧还能继续读。 */
export function shouldForgetComicRun(cause: unknown) {
  return isComicRunSessionLost(cause) || isComicRunGone(cause);
}
