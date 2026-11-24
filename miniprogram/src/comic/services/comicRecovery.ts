import Taro from "@tarojs/taro";
import {authStorageKey} from "../../config/localAuth";
import {onSessionInvalidated} from "../../services/sessionBoundary";
let scopeKey:string|null=null,scopeEpoch=0;
export function configureComicRecoveryScope(username:string){scopeEpoch++;scopeKey=username.trim()?authStorageKey(`${COMIC_ACTIVE_RUN_KEY}:${encodeURIComponent(username.trim())}`):null;}
export function comicRunStorageKey(){return scopeKey;}
export function comicSelectedBookStorageKey(){return scopeKey?`${scopeKey}:selected`:null;}
onSessionInvalidated(()=>configureComicRecoveryScope(""));

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
  try { if(scopeKey)Taro.removeStorageSync(scopeKey); } catch { /* 存储不可用就当作没有指针。 */ }
}

export function saveComicActiveRun(run: Omit<ComicActiveRun, "version">) {
  try { if(scopeKey)Taro.setStorageSync(scopeKey, { ...run, version: 1 }); } catch { /* 存储失败不阻断生成。 */ }
}

/** 读本地指针：损坏或超过 24h 直接清掉，返回 null。 */
export function readComicActiveRun(now = Date.now()): ComicActiveRun | null {
  let raw: unknown = "";
  if(!scopeKey)return null;
  try { raw = Taro.getStorageSync(scopeKey); } catch { return null; }
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

export function saveComicSelectedBook(value:{batchId:string;bookId:string;title?:string;sourceText:string}){const key=comicSelectedBookStorageKey();if(!key)throw Error('漫画账号尚未确认，请先打开漫画生成');Taro.setStorageSync(key,value);}
export function readComicSelectedBook(){const key=comicSelectedBookStorageKey();if(!key)return null;try{const r=Taro.getStorageSync(key);return r&&typeof r==='object'&&typeof r.sourceText==='string'?r:null;}catch{return null;}}
export function clearComicSelectedBook(){const key=comicSelectedBookStorageKey();if(key)Taro.removeStorageSync(key);}
/** Legacy global pointers are never shown before the server proves ownership. */
export async function prepareComicRecovery(request:(path:string)=>Promise<any>,username:string){
 configureComicRecoveryScope(username);const epoch=scopeEpoch;
 const verify=async(batchId:string,bookId:string)=>{try{const value=await request(`/api/batch-factory/v12/batches/${encodeURIComponent(batchId)}`);if(epoch!==scopeEpoch)throw Error('账号已切换，旧恢复记录已隔离');const batch=value?.batch||value;return batch?.id===batchId&&Array.isArray(batch.books)?batch.books.find((b:any)=>b.id===bookId):null;}catch(c:any){if([401,403,404].includes(c.statusCode))return null;throw c;}};
 if(!readComicActiveRun()){let raw:any;try{raw=Taro.getStorageSync(COMIC_ACTIVE_RUN_KEY);}catch{}const legacy=normalize(raw);if(legacy&&Date.now()-legacy.startedAt<=COMIC_ACTIVE_RUN_TTL_MS&&await verify(legacy.batchId,legacy.bookId)){if(epoch!==scopeEpoch)throw Error('账号已切换');saveComicActiveRun(legacy);Taro.removeStorageSync(COMIC_ACTIVE_RUN_KEY);}}
 let selected:any;try{selected=Taro.getStorageSync('comic:selected-existing-book');}catch{}if(selected?.batchId&&selected?.bookId){const book=await verify(String(selected.batchId),String(selected.bookId));if(epoch!==scopeEpoch)throw Error('账号已切换');if(book&&typeof book.sourceText==='string')saveComicSelectedBook({batchId:String(selected.batchId),bookId:String(selected.bookId),title:String(book.title||''),sourceText:book.sourceText});Taro.removeStorageSync('comic:selected-existing-book');}
}

export type ComicPendingOperation={key:string;batchId:string;bookId:string;label:string;taskId?:string;at:number};
function comicOperationsKey(){return scopeKey?`${scopeKey}:operations`:null;}
export function readComicPendingOperations(batchId='',bookId=''):ComicPendingOperation[]{const key=comicOperationsKey();if(!key)return[];try{const r=Taro.getStorageSync(key);return Array.isArray(r)?r.filter(o=>o?.key&&(!batchId||o.batchId===batchId)&&(!bookId||o.bookId===bookId)):[];}catch{return[];}}
export function clearComicPendingOperation(key:string){const storage=comicOperationsKey();if(storage)Taro.setStorageSync(storage,readComicPendingOperations().filter(o=>o.key!==key));}
export async function runComicPaidOperation<T>(operation:Omit<ComicPendingOperation,'at'>,work:()=>Promise<T>,retainTask?:(result:T)=>string){
 const storage=comicOperationsKey();if(!storage)throw Error('漫画账号尚未确认，未发起生成');const rows=readComicPendingOperations();if(rows.some(o=>o.batchId===operation.batchId&&o.bookId===operation.bookId))throw Error('此作品上次生成结果待核对，请先读取原任务，勿重复提交');const record={...operation,at:Date.now()};Taro.setStorageSync(storage,[...rows,record]);
 const remove=()=>{const saved=Taro.getStorageSync(storage);Taro.setStorageSync(storage,Array.isArray(saved)?saved.filter(o=>o?.key!==operation.key):[]);};
 try{const result=await work();const taskId=retainTask?.(result);if(taskId){const saved=Taro.getStorageSync(storage);Taro.setStorageSync(storage,[...(Array.isArray(saved)?saved.filter(o=>o?.key!==operation.key):[]),{...record,taskId}]);}else remove();return result;}catch(c:any){if(c.statusCode>=400&&c.statusCode<500&&![408,409,429].includes(c.statusCode))remove();throw c;}
}
