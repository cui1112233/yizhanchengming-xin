import { LONG_RUNNING_API_TIMEOUT_MS, apiRequest } from './client.js';

// The module name remains for import compatibility while the production
// workbench has moved to the V12 public contract.
const BASE = '/api/batch-factory/v12';
const LOCAL_EXECUTOR_ARTIFACT_PREFIX = '/api/shuihuo-production/local-executor-artifacts/';
const LOCAL_MERGE_MEDIA_PATTERN = /^\/api\/batch-factory\/v1[12]\/batches\/[^/]+\/merge-media\/[^/]+$/;
const LOCAL_MERGE_COVER_PATTERN = /^\/api\/batch-factory\/v1[12]\/batches\/[^/]+\/merge-cover\/[^/]+$/;
const BOOK_ASSET_IMAGE_PATTERN = /^\/api\/batch-factory\/v1[12]\/batches\/[^/]+\/books\/[^/]+\/assets\/[^/]+\/images\/[^/]+\/content$/;
const COMIC_PAGE_IMAGE_PATTERN = /^\/api\/batch-factory\/v12\/batches\/[^/]+\/books\/[^/]+\/comic-pages\/[^/]+\/images\/[^/]+\/(?:content|thumbnail)$/;
const COMIC_STYLE_PREVIEW_PATTERN = /^\/api\/presets\/comic-style\/[^/]+\/preview$/;
const COMIC_REFERENCE_SCENE_PREVIEW_PATTERN = /^\/api\/presets\/comic-reference-scene\/[^/]+\/preview$/;
const COMIC_MATERIAL_PREVIEW_PATTERN = /^\/api\/presets\/comic-material\/[^/?]+\/preview(?:\?[^#]*)?$/;
const COMIC_GLOBAL_REFERENCE_PATTERN = /^\/api\/batch-factory\/v12\/batches\/[^/]+\/comic-references\/[^/]+\/content$/;

function id(value) {
  return encodeURIComponent(String(value ?? ''));
}

function body(payload) {
  return JSON.stringify(payload ?? {});
}

function query(params = {}) {
  const search = new URLSearchParams();
  for (const [key, value] of Object.entries(params)) {
    if (value === undefined || value === null || value === '') continue;
    search.set(key, String(value));
  }
  const suffix = search.toString();
  return suffix ? `?${suffix}` : '';
}

async function productionRetentionDays() {
  const config = await apiRequest('/api/config');
  const days = Number(config?.productionRetentionDays);
  return [3, 7, 14, 30].includes(days) ? days : 7;
}

function localExecutorArtifactRequestPath(value) {
  const raw = String(value || '').trim();
  if (!raw) return '';
  if (raw.startsWith('/')) {
    try {
      const parsed = new URL(raw, 'http://qiantie.local');
      const path = parsed.pathname;
      return path.startsWith(LOCAL_EXECUTOR_ARTIFACT_PREFIX) || LOCAL_MERGE_MEDIA_PATTERN.test(path)
        ? `${path}${parsed.search}`
        : '';
    } catch (_) {
      return '';
    }
  }
  if (typeof window === 'undefined' || !window.location?.origin) return '';
  try {
    const parsed = new URL(raw, window.location.origin);
    if (parsed.origin !== window.location.origin || (!parsed.pathname.startsWith(LOCAL_EXECUTOR_ARTIFACT_PREFIX) && !LOCAL_MERGE_MEDIA_PATTERN.test(parsed.pathname))) return '';
    return `${parsed.pathname}${parsed.search}`;
  } catch (_) {
    return '';
  }
}

function mergeCoverRequestPath(value) {
  const raw = String(value || '').trim();
  if (!raw) return '';
  if (raw.startsWith('/')) {
    try {
      const parsed = new URL(raw, 'http://qiantie.local');
      return LOCAL_MERGE_COVER_PATTERN.test(parsed.pathname) ? `${parsed.pathname}${parsed.search}` : '';
    } catch (_) {
      return '';
    }
  }
  if (typeof window === 'undefined' || !window.location?.origin) return '';
  try {
    const parsed = new URL(raw, window.location.origin);
    if (parsed.origin !== window.location.origin || !LOCAL_MERGE_COVER_PATTERN.test(parsed.pathname)) return '';
    return `${parsed.pathname}${parsed.search}`;
  } catch (_) {
    return '';
  }
}

function bookAssetImageRequestPath(value) {
  const raw = String(value || '').trim();
  if (!raw) return '';
  if (raw.startsWith('/')) {
    try {
      const parsed = new URL(raw, 'http://qiantie.local');
      return BOOK_ASSET_IMAGE_PATTERN.test(parsed.pathname) ? `${parsed.pathname}${parsed.search}` : '';
    } catch (_) {
      return '';
    }
  }
  if (typeof window === 'undefined' || !window.location?.origin) return '';
  try {
    const parsed = new URL(raw, window.location.origin);
    if (parsed.origin !== window.location.origin || !BOOK_ASSET_IMAGE_PATTERN.test(parsed.pathname)) return '';
    return `${parsed.pathname}${parsed.search}`;
  } catch (_) {
    return '';
  }
}

function comicPageImageRequestPath(value) {
  const raw = String(value || '').trim();
  if (!raw) return '';
  try {
    const parsed = new URL(raw, typeof window !== 'undefined' && window.location?.origin ? window.location.origin : 'http://qiantie.local');
    if (!raw.startsWith('/') && typeof window !== 'undefined' && window.location?.origin && parsed.origin !== window.location.origin) return '';
    return COMIC_PAGE_IMAGE_PATTERN.test(parsed.pathname) ? `${parsed.pathname}${parsed.search}` : '';
  } catch (_) {
    return '';
  }
}

export function isProtectedProductionMediaURL(value) {
  return Boolean(localExecutorArtifactRequestPath(value));
}

export function getProductionMediaBlob(mediaUrl) {
  const path = localExecutorArtifactRequestPath(mediaUrl);
  if (!path) return Promise.reject(new Error('Production media URL is not a local executor artifact or local merge artifact'));
  return apiRequest(path, { responseType: 'blob' });
}

const MERGE_MEDIA_PARTS_PATTERN = /^\/api\/batch-factory\/v1[12]\/batches\/([^/]+)\/merge-media\/([^/]+)$/;

function mergeMediaParts(value) {
  const raw = String(value || '').trim();
  if (!raw) return null;
  const origin = typeof window !== 'undefined' && window.location?.origin ? window.location.origin : 'http://qiantie.local';
  try {
    const parsed = new URL(raw, origin);
    if (typeof window !== 'undefined' && window.location?.origin && parsed.origin !== window.location.origin) return null;
    const match = parsed.pathname.match(MERGE_MEDIA_PARTS_PATTERN);
    return match ? { batchId: match[1], artifactId: match[2] } : null;
  } catch (_) {
    return null;
  }
}

export function isMergeMediaURL(value) {
  return Boolean(mergeMediaParts(value));
}

// Native <video> elements cannot attach the Authorization header. Mint a
// short-lived signed URL so the browser can stream the merge artifact with
// Range requests instead of downloading the whole file into a Blob first.
export async function getMergeMediaPlaybackURL(mediaUrl) {
  const parts = mergeMediaParts(mediaUrl);
  if (!parts) throw new Error('合成成片地址无效');
  const payload = await apiRequest(`/api/batch-factory/v12/batches/${id(parts.batchId)}/merge-media/${id(parts.artifactId)}/token`, {
    method: 'POST',
    silent: true,
    suppressGlobalError: true
  });
  const playbackUrl = String(payload?.url || '').trim();
  if (!playbackUrl) throw new Error('服务端未返回可播放的合成成片地址');
  return playbackUrl;
}

export function isBatchFactoryMergeCoverURL(value) {
  return Boolean(mergeCoverRequestPath(value));
}

export function isBookAssetImageURL(value) {
  return Boolean(bookAssetImageRequestPath(value));
}

export function isComicPageImageURL(value) {
  return Boolean(comicPageImageRequestPath(value));
}

// Native <img> requests cannot carry the app bearer token. Merge covers stay
// private, so load the small cached JPG through the authenticated API client
// instead of exposing a public video or image URL.
export function getBatchFactoryMergeCoverBlob(coverUrl) {
  const path = mergeCoverRequestPath(coverUrl);
  if (!path) return Promise.reject(new Error('Merge cover URL is not a local Batch Factory cover artifact'));
  return apiRequest(path, { responseType: 'blob', silent: true, suppressGlobalError: true });
}

// 漫画页图片体积较大，一次请求超时或服务端短暂不可用时不应直接判定为永久失败，
// 否则预览会一直停在“图片加载失败”。仅对网络中断（无状态码）与 5xx 做有限重试，
// 401/403/404 属于确定性失败，重试没有意义。
const IMAGE_BLOB_RETRY_DELAYS = [400, 1200];

async function requestImageBlobWithRetry(path) {
  let lastError;
  for (let attempt = 0; attempt <= IMAGE_BLOB_RETRY_DELAYS.length; attempt += 1) {
    try {
      return await apiRequest(path, { responseType: 'blob', silent: true, suppressGlobalError: true });
    } catch (error) {
      lastError = error;
      const status = typeof error?.status === 'number' ? error.status : 0;
      if ((status !== 0 && status < 500) || attempt === IMAGE_BLOB_RETRY_DELAYS.length) break;
      await new Promise(resolve => setTimeout(resolve, IMAGE_BLOB_RETRY_DELAYS[attempt]));
    }
  }
  throw lastError;
}

export function getBookAssetImageBlob(imageUrl) {
  const path = bookAssetImageRequestPath(imageUrl);
  if (!path) return Promise.reject(new Error('Book asset image URL is not a local Batch Factory asset image'));
  return requestImageBlobWithRetry(path);
}

export function getComicPageImageBlob(imageUrl) {
  const path = comicPageImageRequestPath(imageUrl);
  if (!path) return Promise.reject(new Error('Comic page image URL is not a local Batch Factory comic image'));
  return requestImageBlobWithRetry(path);
}

export function getComicReferencePreviewBlob(imageUrl) {
  const path = String(imageUrl || '').trim();
  if (!COMIC_STYLE_PREVIEW_PATTERN.test(path) && !COMIC_REFERENCE_SCENE_PREVIEW_PATTERN.test(path) && !COMIC_MATERIAL_PREVIEW_PATTERN.test(path) && !COMIC_GLOBAL_REFERENCE_PATTERN.test(path)) {
    return Promise.reject(new Error('Comic reference preview URL is not a protected Batch Factory image'));
  }
  return requestImageBlobWithRetry(path);
}

export function bf11Path(path = '') {
  const clean = String(path || '').replace(/^\/+/, '');
  return clean ? `${BASE}/${clean}` : BASE;
}

export function bf11ScopePath({ scope, batchId, bookId = '', videoId = '' }) {
  if (scope === 'batch') return bf11Path(`batches/${id(batchId)}/settings`);
  if (scope === 'book') return bf11Path(`batches/${id(batchId)}/books/${id(bookId)}/override`);
  if (scope === 'video') return bf11Path(`batches/${id(batchId)}/books/${id(bookId)}/videos/${id(videoId)}/override`);
  throw new Error(`Unsupported batch factory scope: ${scope}`);
}

export function getCapabilities() {
  return apiRequest(bf11Path('capabilities'));
}

// Kept for the legacy Batch Factory page. The Shuihuo workbench uses the
// book-scoped image endpoints below, but Vite still statically imports this
// page and must retain its public API contract.
export function generateConfiguredImage(payload = {}) {
  return apiRequest(bf11Path('image-generation'), { method: 'POST', body: body(payload), timeoutMs: LONG_RUNNING_API_TIMEOUT_MS });
}

export function createNovelFetchIntake(payload) {
  return apiRequest(bf11Path('intakes/novel-fetch'), { method: 'POST', body: body(payload) });
}

export function createManualIntake(payload) {
  return apiRequest(bf11Path('intakes/manual'), { method: 'POST', body: body(payload) });
}

export function createPrivateComicTask(payload = {}) {
  return apiRequest(bf11Path('private-comic/tasks'), { method: 'POST', body: body(payload), timeoutMs: LONG_RUNNING_API_TIMEOUT_MS });
}

export function getIntake(intakeId) {
  return apiRequest(bf11Path(`intakes/${id(intakeId)}`));
}

export function createBatchFromIntake(intakeId, payload = {}) {
  return apiRequest(bf11Path(`intakes/${id(intakeId)}/batches`), { method: 'POST', body: body(payload) });
}

// The novel-fetch source is already frozen in the intake. The service appends
// only source books not already present in the selected batch.
export function appendNovelFetchIntake(batchId, intakeId, options = {}) {
  return apiRequest(bf11Path(`batches/${id(batchId)}/intakes/${id(intakeId)}/books`), {
    method: 'POST',
    body: body({ allowDuplicate: options.allowDuplicate === true })
  });
}

export function listBatches() {
  return apiRequest(bf11Path('batches'));
}

// The project library must not download full originals and storyboard payloads
// for every batch. Detailed data remains available through getBatch(batchId).
export function listBatchSummaries() {
  return apiRequest(bf11Path('batches/summary'));
}

export function createBatch(payload) {
  return apiRequest(bf11Path('batches'), { method: 'POST', body: body(payload) });
}

export function fetchDirectOriginals(payload) {
  const normalized = {
    platform: Number(payload?.platform) || 0,
    bookIds: Array.isArray(payload?.bookIds)
      ? payload.bookIds.map(value => String(value ?? '').trim()).filter(Boolean)
      : [],
    maxTxt: Number(payload?.maxTxt) || 4000
  };
  // Keep the JSON body as the primary contract. The query is a compatibility
  // fallback for the local embedded browser/runtime: older cached fetch
  // wrappers have been observed to drop POST bodies while retaining the URL.
  const params = new URLSearchParams();
  params.set('platform', String(normalized.platform));
  params.set('maxTxt', String(normalized.maxTxt));
  normalized.bookIds.forEach(bookId => params.append('bookId', bookId));
  return apiRequest(`${bf11Path('fetch-originals')}?${params.toString()}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(normalized),
    timeoutMs: LONG_RUNNING_API_TIMEOUT_MS
  });
}

export function fetchBookOriginal(batchId, bookId, options = {}) {
  return apiRequest(bf11Path(`batches/${id(batchId)}/books/${id(bookId)}/fetch-original`), {
    method: 'POST',
    body: body({
      ...(Number(options.maxTxt) > 0 ? { maxTxt: Number(options.maxTxt) } : {}),
      ...(String(options.platform || '').trim() ? { platform: String(options.platform).trim() } : {})
    }),
    timeoutMs: LONG_RUNNING_API_TIMEOUT_MS
  });
}

// Runs after a manual batch has been created.  The API isolates failures per
// book, so classification is useful metadata rather than a creation gate.
export function classifyFetchedBatchMetadata(batchId) {
  return apiRequest(bf11Path(`batches/${id(batchId)}/classify-fetched-metadata`), { method: 'POST', body: body({}) });
}

// Repairs only 121 category/genre-derived metadata for books that were
// imported before direct fetch carried bookinfo through intake creation.
export function backfillBatch121Metadata(batchId) {
  return apiRequest(bf11Path(`batches/${id(batchId)}/backfill-121-metadata`), { method: 'POST', body: body({}) });
}

export function getBatch(batchId, options = {}) {
  return apiRequest(bf11Path(`batches/${id(batchId)}`), options);
}

export function bindBatchFactoryAgentProject(batchId) {
  return apiRequest(`${BASE}/batches/${id(batchId)}/agent-project`, { method: 'POST' });
}

export async function deleteBatchFactoryBook(batchId, bookId) {
  const retentionDays = await productionRetentionDays();
  return apiRequest(`${bf11Path(`batches/${id(batchId)}/books/${id(bookId)}`)}${query({ retentionDays })}`, { method: 'DELETE' });
}

export async function deleteBatchFactoryProject(batchId) {
  const retentionDays = await productionRetentionDays();
  return apiRequest(`${bf11Path(`batches/${id(batchId)}`)}${query({ retentionDays })}`, { method: 'DELETE' });
}

export function updateBookSource(batchId, bookId, input) {
  return apiRequest(bf11Path(`batches/${id(batchId)}/books/${id(bookId)}/source`), { method: 'PUT', body: body(input) });
}

export function saveBatchSettings(batchId, input, options = {}) {
  return apiRequest(bf11ScopePath({ scope: 'batch', batchId }), { ...options, method: 'PUT', body: body(input) });
}

export function updateBookMetadata(batchId, bookId, input) {
  return apiRequest(bf11Path(`batches/${id(batchId)}/books/${id(bookId)}/metadata`), { method: 'PUT', body: body(input) });
}

export function saveBookOverride(batchId, bookId, input) {
  return apiRequest(bf11ScopePath({ scope: 'book', batchId, bookId }), { method: 'PUT', body: body(input) });
}

export function saveVideoOverride(batchId, bookId, videoId, input) {
	return apiRequest(bf11ScopePath({ scope: 'video', batchId, bookId, videoId }), { method: 'PUT', body: body(input) });
}

export function uploadBookVideoMaterial(batchId, bookId, videoId, file) {
  const form = new FormData();
  form.append('file', file);
  return apiRequest(bf11Path(`batches/${id(batchId)}/books/${id(bookId)}/videos/${id(videoId)}/upload`), { method: 'POST', body: form });
}

export function uploadBookMergeMaterial(batchId, bookId, file) {
  const form = new FormData();
  form.append('file', file);
  return apiRequest(bf11Path(`batches/${id(batchId)}/books/${id(bookId)}/merge/upload`), { method: 'POST', body: form });
}

export function deleteProductionTask(batchId, bookId, videoId, taskId) {
	return apiRequest(bf11Path(`batches/${id(batchId)}/books/${id(bookId)}/videos/${id(videoId)}/tasks/${id(taskId)}`), { method: 'DELETE' });
}

function bookAssetsPath(batchId, bookId, assetId = '') {
  const base = bf11Path(`batches/${id(batchId)}/books/${id(bookId)}/assets`);
  return assetId ? `${base}/${id(assetId)}` : base;
}

function bookAssetImagesPath(batchId, bookId, assetId, imageId = '') {
  const base = `${bookAssetsPath(batchId, bookId, assetId)}/images`;
  return imageId ? `${base}/${id(imageId)}` : base;
}

export function listBookAssets(batchId, bookId, options = {}) {
  return apiRequest(bookAssetsPath(batchId, bookId), options);
}

export function createBookAsset(batchId, bookId, payload) {
  return apiRequest(bookAssetsPath(batchId, bookId), { method: 'POST', body: body(payload) });
}

export function updateBookAsset(batchId, bookId, assetId, payload) {
  return apiRequest(bookAssetsPath(batchId, bookId, assetId), { method: 'PATCH', body: body(payload) });
}

export function listBookAssetImages(batchId, bookId, assetId, options = {}) {
  return apiRequest(bookAssetImagesPath(batchId, bookId, assetId), options);
}

// Provider-result registration is intentionally separate from browser upload.
// The UI never invents a provider URL; it only uses uploadBookAssetImage.
export function createBookAssetImage(batchId, bookId, assetId, payload) {
  return apiRequest(bookAssetImagesPath(batchId, bookId, assetId), { method: 'POST', body: body(payload) });
}

export function uploadBookAssetImage(batchId, bookId, assetId, dataUrl) {
  return apiRequest(`${bookAssetImagesPath(batchId, bookId, assetId)}/upload`, { method: 'POST', body: body({ dataUrl }) });
}

export function setPrimaryBookAssetImage(batchId, bookId, assetId, imageId) {
  return apiRequest(`${bookAssetImagesPath(batchId, bookId, assetId, imageId)}/primary`, { method: 'PUT', body: body({}) });
}

// The browser sends only selected asset IDs and the configured model ID. The
// authenticated Node gateway resolves the account credential, calls the model,
// and persists each returned image as a version owned by this book asset.
export function generateBookAssetImages(batchId, bookId, payload) {
  return apiRequest(`${bookAssetsPath(batchId, bookId)}/images/generate`, { method: 'POST', body: body(payload), timeoutMs: LONG_RUNNING_API_TIMEOUT_MS });
}

export function getBookAssetImageGenerationTask(batchId, bookId, taskId) {
  return apiRequest(`${bookAssetsPath(batchId, bookId)}/images/generate/${id(taskId)}`);
}

export function getConfigVersions() {
  return apiRequest(bf11Path('config-versions'));
}

export function createConfigVersion(payload) {
  return apiRequest(bf11Path('config-versions'), { method: 'POST', body: body(payload) });
}

export function renameConfigVersion(versionId, name) {
  return apiRequest(bf11Path(`config-versions/${id(versionId)}`), { method: 'PUT', body: body({ name }) });
}

export function getChangeImpact(batchId, payload) {
  return apiRequest(bf11Path(`batches/${id(batchId)}/change-impact`), { method: 'POST', body: body(payload) });
}

export function listPrompts(params = {}) {
  return apiRequest(`${bf11Path('prompts')}${query(params)}`);
}

export function createPrompt(payload) {
  return apiRequest(bf11Path('prompts'), { method: 'POST', body: body(payload) });
}

// System preset metadata is safe for the browser. The protected body is
// resolved only in the Node-to-Go bridge when a Batch Factory setting is saved.
export function listSystemPresetCatalog(module) {
  const catalogNonce = Date.now();
  return apiRequest(`/api/presets?module=${encodeURIComponent(String(module || ''))}&_=${catalogNonce}`, { cache: 'no-store' });
}

export function listBatchFactorySystemPresets() {
  return listSystemPresetCatalog('batch-factory');
}

// Legacy V11 workbench compatibility. The Shuihuo-integrated Batch Factory
// modal above must use listBatchFactorySystemPresets instead.
export function listPersonalConstraintPrompts(category) {
  return apiRequest(`/api/script-constraint-prompts?category=${encodeURIComponent(String(category || ''))}`);
}

export function savePersonalConstraintPrompt(payload) {
  return apiRequest('/api/script-constraint-prompts', { method: 'POST', body: body(payload) });
}

export function getDraft(params = {}, options = {}) {
  return apiRequest(`${bf11Path('drafts')}${query(params)}`, options);
}

export function saveDraft(payload) {
  return apiRequest(bf11Path('drafts'), { method: 'PUT', body: body(payload) });
}

export function rewriteWorkingFront(batchId, bookId, content, payload = {}) {
  return apiRequest(bf11Path(`batches/${id(batchId)}/books/${id(bookId)}/working-front/viral`), { method: 'POST', body: body({ ...payload, content: String(content || '') }), timeoutMs: LONG_RUNNING_API_TIMEOUT_MS });
}

export function runHook(batchId, bookId, payload = {}) {
  return apiRequest(bf11Path(`batches/${id(batchId)}/books/${id(bookId)}/hook`), { method: 'POST', body: body(payload), timeoutMs: LONG_RUNNING_API_TIMEOUT_MS });
}

export function approveHook(batchId, bookId, hookId) {
  return apiRequest(bf11Path(`batches/${id(batchId)}/books/${id(bookId)}/hooks/${id(hookId)}/approve`), { method: 'POST', body: body({}) });
}

export function runDirector(batchId, bookId) {
  return apiRequest(bf11Path(`batches/${id(batchId)}/books/${id(bookId)}/director`), { method: 'POST', body: body({}), timeoutMs: LONG_RUNNING_API_TIMEOUT_MS });
}

export function uploadComicGlobalReference(batchId, dataUrl) {
  return apiRequest(bf11Path(`batches/${id(batchId)}/comic-reference/upload`), { method: 'POST', body: body({ dataUrl }), timeoutMs: LONG_RUNNING_API_TIMEOUT_MS });
}

function comicPagesPath(batchId, bookId, suffix = '') {
  const base = bf11Path(`batches/${id(batchId)}/books/${id(bookId)}/comic-pages`);
  return suffix ? `${base}/${suffix}` : base;
}

function comicCollectionsPath(batchId, bookId, suffix = '') {
  const base = bf11Path(`batches/${id(batchId)}/books/${id(bookId)}/comic-collections`);
  return suffix ? `${base}/${suffix}` : base;
}

export function listComicPages(batchId, bookId, options = {}) {
  // view=lite 只返回列表行需要的状态与图片地址，去掉提示词正文（单本约 300KB）。
  // 漫画页编辑弹窗需要完整提示词，保持默认全量。
  const { view, ...requestOptions } = options;
  const path = comicPagesPath(batchId, bookId);
  // 生图成功后必须立即读到新候选图；浏览器旧 GET 缓存会让用户误以为图片没回来。
  return apiRequest(view ? `${path}?view=${encodeURIComponent(view)}` : path, { ...requestOptions, cache: 'no-store' });
}

function comicSourceSegmentsPath(batchId, bookId, suffix = '') {
  const base = bf11Path(`batches/${id(batchId)}/books/${id(bookId)}/comic-source-segments`);
  return suffix ? `${base}/${suffix}` : base;
}

export function getComicSourceSegments(batchId, bookId, options = {}) {
  return apiRequest(comicSourceSegmentsPath(batchId, bookId), options);
}

export function segmentComicSource(batchId, bookId) {
  return apiRequest(comicSourceSegmentsPath(batchId, bookId, 'segment'), { method: 'POST', body: body({}), timeoutMs: LONG_RUNNING_API_TIMEOUT_MS });
}

export function splitComicSourceSegment(batchId, bookId, videoId, payload) {
  return apiRequest(comicSourceSegmentsPath(batchId, bookId, `${id(videoId)}/split`), { method: 'POST', body: body(payload) });
}

export function generateComicPage(batchId, bookId, videoId, payload) {
  // 409 是“该页已有在途生成”，由调用方合并成一条提示并启动轮询；
  // 这里不再触发全局顶栏错误，避免同一个失败被全局拦截器和页面各弹一次。
  return apiRequest(comicPagesPath(batchId, bookId, `${id(videoId)}/generate`), { method: 'POST', body: body(payload), suppressGlobalError: true, timeoutMs: LONG_RUNNING_API_TIMEOUT_MS });
}

export function updateComicPage(batchId, bookId, videoId, payload) {
  return apiRequest(comicPagesPath(batchId, bookId, id(videoId)), { method: 'PUT', body: body(payload), suppressGlobalError: true });
}

export function generateComicBook(batchId, bookId) {
  return apiRequest(comicCollectionsPath(batchId, bookId, 'generate'), { method: 'POST', body: body({}), timeoutMs: LONG_RUNNING_API_TIMEOUT_MS });
}

export function getComicPagePromptPreview(batchId, bookId, videoId, payload = {}) {
  return apiRequest(comicPagesPath(batchId, bookId, `${id(videoId)}/prompt-preview`), { method: 'POST', body: body(payload), timeoutMs: LONG_RUNNING_API_TIMEOUT_MS });
}

export function downloadComicPromptZip(batchId, bookId) {
  return apiRequest(comicPagesPath(batchId, bookId, 'download-zip'), { responseType: 'blob' });
}

export function downloadComicPageImagesZip(batchId, bookId) {
  return apiRequest(comicPagesPath(batchId, bookId, 'images-zip'), { responseType: 'blob' });
}

export function downloadDouyinPublishZip(batchId, bookId) {
  return apiRequest(comicPagesPath(batchId, bookId, 'douyin-zip'), { responseType: 'blob' });
}

export function uploadComicPageImage(batchId, bookId, videoId, dataUrl) {
  return apiRequest(comicPagesPath(batchId, bookId, `${id(videoId)}/images`), { method: 'POST', body: body({ dataUrl }) });
}

export function downloadComicCollectionImageZip(batchId, bookId, collectionId) {
  return apiRequest(comicCollectionsPath(batchId, bookId, `${id(collectionId)}/download`), { responseType: 'blob' });
}

export function setComicPageReferenceImages(batchId, bookId, videoId, referenceImageUrls) {
  return apiRequest(comicPagesPath(batchId, bookId, `${id(videoId)}/reference-images`), { method: 'PUT', body: body({ referenceImageUrls }) });
}

export function applyComicPageReferenceImages(batchId, bookId, videoId) {
  return apiRequest(comicPagesPath(batchId, bookId, `${id(videoId)}/reference-images/apply-to-all`), { method: 'POST', body: body({}) });
}

export function setPrimaryComicPageImage(batchId, bookId, pageId, imageId) {
  return apiRequest(comicPagesPath(batchId, bookId, `${id(pageId)}/images/${id(imageId)}/primary`), { method: 'PUT', body: body({}) });
}

export function createComicCollection(batchId, bookId, payload = {}) {
  return apiRequest(comicCollectionsPath(batchId, bookId), { method: 'POST', body: body(payload) });
}

export function listComicCollections(batchId, bookId, options = {}) {
  return apiRequest(comicCollectionsPath(batchId, bookId), options);
}

export function getN8ComicPublishStatus(options = {}) {
  return apiRequest(bf11Path('comic-publish/status'), options);
}

export function getN8ComicPublishAudioCategories(options = {}) {
  return apiRequest(bf11Path('comic-publish/audio-categories'), options);
}

export function getN8ComicPublishAudios({ audioCategoryId = '', organizationId = '' } = {}, options = {}) {
  return apiRequest(`${bf11Path('comic-publish/audios')}${query({ audioCategoryId, organizationId })}`, options);
}

export function searchN8ComicWorks(query, options = {}) {
  const value = String(query || '').trim();
  if (!value) return Promise.resolve({ works: [] });
  return apiRequest(`${bf11Path('comic-publish/works')}?query=${encodeURIComponent(value)}`, options);
}

export function publishComicCollection(batchId, bookId, collectionId, payload) {
  return apiRequest(`${comicCollectionsPath(batchId, bookId)}/${id(collectionId)}/publish`, { method: 'POST', body: body(payload) });
}

export function getComicCollectionPublishRecord(batchId, bookId, collectionId, options = {}) {
  return apiRequest(`${comicCollectionsPath(batchId, bookId)}/${id(collectionId)}/publish`, options);
}

function directorPromptCandidatePath(batchId, bookId) {
  // The public workbench runs on V12. V11 is intentionally hidden in the
  // local/review runtime, so candidate actions must use the same public
  // namespace as the rest of the director workbench.
  return bf11Path(`batches/${id(batchId)}/books/${id(bookId)}/director-prompt-candidate`);
}

export function createDirectorPromptCandidate(batchId, bookId, expectedRevision) {
  return apiRequest(directorPromptCandidatePath(batchId, bookId), {
    method: 'POST',
    body: body({ expectedRevision: Number(expectedRevision || 0) })
  });
}

export function getDirectorPromptCandidate(batchId, bookId) {
  return apiRequest(directorPromptCandidatePath(batchId, bookId), { silent: true, suppressGlobalError: true });
}

export function acceptDirectorPromptCandidate(batchId, bookId, videoId, candidateId, expectedRevision) {
  return apiRequest(bf11Path(`batches/${id(batchId)}/books/${id(bookId)}/videos/${id(videoId)}/director-prompt-candidate/accept`), {
    method: 'POST',
    body: body({ candidateId, expectedRevision: Number(expectedRevision || 0) })
  });
}

export function cancelDirectorPromptCandidate(batchId, bookId, expectedRevision) {
  return apiRequest(directorPromptCandidatePath(batchId, bookId), {
    method: 'DELETE',
    body: body({ expectedRevision: Number(expectedRevision || 0) })
  });
}

export function runH3Director(batchId, bookId, payload) {
  return apiRequest(bf11Path(`batches/${id(batchId)}/books/${id(bookId)}/h3/director`), { method: 'POST', body: body(payload) });
}

export function measureH3Audio(batchId, bookId, audioBase64) {
  return apiRequest(bf11Path(`batches/${id(batchId)}/books/${id(bookId)}/h3/audio-measurement`), {
    method: 'POST',
    body: body(audioBase64 && typeof audioBase64 === 'object' ? audioBase64 : { audio_base64: String(audioBase64 || '') })
  });
}

export function compileH3Video(batchId, bookId, payload) {
  return apiRequest(bf11Path(`batches/${id(batchId)}/books/${id(bookId)}/h3/compile`), { method: 'POST', body: body(payload) });
}

export function getH3Trace(batchId, bookId, compilationId = '') {
  return apiRequest(`${bf11Path(`batches/${id(batchId)}/books/${id(bookId)}/h3/trace`)}${query({ compilationId })}`);
}

export function getBookStageSummary(batchId, bookId) {
  return apiRequest(bf11Path(`batches/${id(batchId)}/books/${id(bookId)}/stages`));
}

export function runBookStage(batchId, bookId, stage, payload = {}) {
  return apiRequest(bf11Path(`batches/${id(batchId)}/books/${id(bookId)}/stages/${id(stage)}`), { method: 'POST', body: body(payload) });
}

export function createBookAudioPlan(batchId, bookId, payload = {}) {
  return apiRequest(bf11Path(`batches/${id(batchId)}/books/${id(bookId)}/audio-plan`), { method: 'POST', body: body(payload) });
}

export function retryBookStage(batchId, bookId, payload = {}) {
  return apiRequest(bf11Path(`batches/${id(batchId)}/books/${id(bookId)}/stages/retry`), { method: 'POST', body: body(payload) });
}

export function refreshBookSmartUnified(batchId, bookId, payload = {}) {
  return apiRequest(bf11Path(`batches/${id(batchId)}/books/${id(bookId)}/smart-unified/refresh`), { method: 'POST', body: body(payload) });
}

export function runBatchDirector(batchId, payload = {}) {
  // 409 表示「全部书都失败」但响应体仍带逐书 reasons；放行状态码让调用方
  // 能列出失败书名，而不是只拿到一句通用错误。
  return apiRequest(bf11Path(`batches/${id(batchId)}/director`), { method: 'POST', body: body(payload), allowStatuses: [409] });
}

export function getEffectiveSettings(batchId, bookId, videoId) {
  return apiRequest(bf11Path(`batches/${id(batchId)}/books/${id(bookId)}/videos/${id(videoId)}/effective-settings`));
}

export function getFinalPrompt(batchId, bookId, videoId, options = {}) {
  return apiRequest(bf11Path(`batches/${id(batchId)}/books/${id(bookId)}/videos/${id(videoId)}/final-prompt`), options);
}

export function submitBookProduction(batchId, bookId, requestId, provider = 'personal_api', videoModelId = '') {
  return apiRequest(bf11Path(`batches/${id(batchId)}/books/${id(bookId)}/production`), {
    method: 'POST',
    body: body({ requestId, provider, ...(videoModelId ? { videoModelId } : {}) })
  });
}

export function submitBatchProduction(batchId, requestId, provider = 'personal_api', videoModelId = '') {
  return apiRequest(bf11Path(`batches/${id(batchId)}/production`), {
    method: 'POST',
    body: body({ requestId, provider, ...(videoModelId ? { videoModelId } : {}) })
  });
}

// The server only marks an active task cancelled after its local executor has
// accepted the cancellation. Unsupported providers are returned as skipped.
export function cancelBatchProduction(batchId, requestId) {
  return apiRequest(bf11Path(`batches/${id(batchId)}/production/cancel`), {
    method: 'POST',
    body: body({ requestId })
  });
}

export function saveVideoProviderConfig(payload = {}) {
  return apiRequest(bf11Path('video-provider/config'), { method: 'PUT', body: body(payload) });
}

export function getVideoProviderStatus(provider = 'personal_api') {
  return apiRequest(`${bf11Path('video-provider/status')}${query({ provider })}`);
}

export function listLocalExecutors() {
  return apiRequest('/api/shuihuo-production/local-executors');
}

export function createLocalExecutorPairing(platform = 'doubao') {
  return apiRequest('/api/shuihuo-production/local-executors/pairings', {
    method: 'POST',
    body: body({ platform })
  });
}

export function getProductionStatus(batchId, options = {}) {
  return apiRequest(bf11Path(`batches/${id(batchId)}/status`), options);
}

export function getBatchAutomationStatus(batchId) {
  return apiRequest(bf11Path(`batches/${id(batchId)}/automation`), { cache: 'no-store' });
}

export function getBatchRuntimeSummary(batchId, options = {}) {
  return apiRequest(bf11Path(`batches/${id(batchId)}/runtime-summary`), { ...options, cache: 'no-store' });
}

export function listSchedules() {
  return apiRequest(bf11Path('schedules'));
}

export function createSchedule(payload = {}) {
  return apiRequest(bf11Path('schedules'), { method: 'POST', body: body(payload) });
}

export function deleteSchedule(scheduleId) {
  return apiRequest(bf11Path(`schedules/${id(scheduleId)}`), { method: 'DELETE' });
}

export function listAutomationPresets() {
  return apiRequest(bf11Path('automation-presets'));
}

export function createAutomationPreset(payload) {
  return apiRequest(bf11Path('automation-presets'), { method: 'POST', body: body(payload) });
}

export function updateAutomationPreset(presetId, payload) {
  return apiRequest(bf11Path(`automation-presets/${id(presetId)}`), { method: 'PUT', body: body(payload) });
}

export function deleteAutomationPreset(presetId) {
  return apiRequest(bf11Path(`automation-presets/${id(presetId)}`), { method: 'DELETE', body: body({}) });
}

export function saveBatchAutomationPreset(batchId, name) {
  return apiRequest(bf11Path(`batches/${id(batchId)}/automation-presets`), { method: 'POST', body: body({ name }) });
}

export function startBatchAutomation(batchId, payload = {}) {
  return apiRequest(bf11Path(`batches/${id(batchId)}/automation/start`), { method: 'POST', body: body(payload) });
}

export function pauseBatchAutomation(batchId) {
  return apiRequest(bf11Path(`batches/${id(batchId)}/automation/pause`), { method: 'POST', body: body({}) });
}

export function resumeBatchAutomation(batchId) {
  return apiRequest(bf11Path(`batches/${id(batchId)}/automation/resume`), { method: 'POST', body: body({}) });
}

export function retryBatchAutomation(batchId, bookIds = []) {
  return apiRequest(bf11Path(`batches/${id(batchId)}/automation/retry`), { method: 'POST', body: body({ bookIds }) });
}

export function cancelBatchAutomation(batchId) {
  return apiRequest(bf11Path(`batches/${id(batchId)}/automation/cancel`), { method: 'POST', body: body({}) });
}

export function submitBatchMerge(batchId, payload = {}) {
  return apiRequest(bf11Path(`batches/${id(batchId)}/merge`), {
    method: 'POST',
    body: body(payload)
  });
}

export function submitBookMerge(batchId, bookId, payload = {}) {
  return apiRequest(bf11Path(`batches/${id(batchId)}/books/${id(bookId)}/merge`), {
    method: 'POST',
    body: body(payload)
  });
}

export function getBookMergeStatus(batchId, bookId, requestId) {
  return apiRequest(`${bf11Path(`batches/${id(batchId)}/books/${id(bookId)}/merge-status`)}${query({ requestId })}`);
}

export function getMergeStatus(batchId, options = {}) {
  return apiRequest(bf11Path(`batches/${id(batchId)}/merge-status`), options);
}

export function getPublishCredential(provider) {
  return apiRequest(bf11Path(`publish/${id(provider)}/credential`));
}

export function savePublishCredential(provider, payload) {
  return apiRequest(bf11Path(`publish/${id(provider)}/credential`), { method: 'PUT', body: body(payload) });
}

export function createPublishIntent(provider, payload) {
  return apiRequest(bf11Path(`publish/${id(provider)}/intents`), { method: 'POST', body: body(payload) });
}

// V11 书级发布走已验证的 121 PHP 会话：每次只提交这一书的
// `{bookId}.txt + {bookId}.mp4`，服务端负责会话、媒体读取和回读。
export function get121OrganizationOptions() {
  return apiRequest(bf11Path('publish-121/organizations'));
}

export function classifyBookPublishMetadata(batchId, bookId, payload = {}) {
  return apiRequest(bf11Path(`batches/${id(batchId)}/books/${id(bookId)}/classify-publish-metadata`), {
    method: 'POST',
    body: body(payload)
  });
}

export function submitBookTo121(batchId, bookId, payload = {}) {
  return apiRequest(bf11Path(`batches/${id(batchId)}/books/${id(bookId)}/publish-121`), {
    method: 'POST',
    body: body(payload)
  });
}

export function confirmPublishIntent(provider, intentId) {
  return apiRequest(bf11Path(`publish/${id(provider)}/intents/${id(intentId)}/confirm`), { method: 'POST', body: body({}) });
}

export function submitPublishIntent(provider, intentId) {
  return apiRequest(bf11Path(`publish/${id(provider)}/intents/${id(intentId)}/submit`), { method: 'POST', body: body({}) });
}

export function getPublishAudits(provider, intentId) {
  return apiRequest(bf11Path(`publish/${id(provider)}/intents/${id(intentId)}/audits`));
}

export default {
  getCapabilities,
  getBookAssetImageBlob,
  getComicPageImageBlob,
  getComicReferencePreviewBlob,
  generateConfiguredImage,
  createNovelFetchIntake,
  getIntake,
  createBatchFromIntake,
  appendNovelFetchIntake,
  listBatches,
  createBatch,
  fetchDirectOriginals,
  fetchBookOriginal,
  backfillBatch121Metadata,
  getBatch,
  bindBatchFactoryAgentProject,
  updateBookSource,
  saveBatchSettings,
  saveBookOverride,
  saveVideoOverride,
  deleteProductionTask,
  listBookAssets,
  createBookAsset,
  updateBookAsset,
  listBookAssetImages,
  createBookAssetImage,
  uploadBookAssetImage,
  uploadComicGlobalReference,
  setPrimaryBookAssetImage,
  getBookAssetImageGenerationTask,
  listComicPages,
  updateComicPage,
  generateComicPage,
  downloadComicPromptZip,
  downloadComicPageImagesZip,
  downloadDouyinPublishZip,
  generateComicBook,
  setComicPageReferenceImages,
  applyComicPageReferenceImages,
  setPrimaryComicPageImage,
  createComicCollection,
  listComicCollections,
	getN8ComicPublishStatus,
	getN8ComicPublishAudioCategories,
	getN8ComicPublishAudios,
	publishComicCollection,
	getComicCollectionPublishRecord,
  getConfigVersions,
  createConfigVersion,
  renameConfigVersion,
  getChangeImpact,
  listPrompts,
  createPrompt,
  listPersonalConstraintPrompts,
  savePersonalConstraintPrompt,
  getDraft,
  saveDraft,
  rewriteWorkingFront,
  runHook,
  approveHook,
  runDirector,
  runBatchDirector,
  getEffectiveSettings,
  getFinalPrompt,
  submitBookProduction,
  submitBatchProduction,
  saveVideoProviderConfig,
  getVideoProviderStatus,
  listLocalExecutors,
  createLocalExecutorPairing,
  getProductionStatus,
  getBatchAutomationStatus,
  getBatchRuntimeSummary,
  listSchedules,
  createSchedule,
  deleteSchedule,
  listAutomationPresets,
  createAutomationPreset,
  updateAutomationPreset,
  deleteAutomationPreset,
  saveBatchAutomationPreset,
  startBatchAutomation,
  pauseBatchAutomation,
  resumeBatchAutomation,
  retryBatchAutomation,
  cancelBatchAutomation,
  getProductionMediaBlob,
  isProtectedProductionMediaURL,
  submitBatchMerge,
  submitBookMerge,
  getBookMergeStatus,
  getMergeStatus,
  getPublishCredential,
  savePublishCredential,
  createPublishIntent,
  submitBookTo121,
  classifyBookPublishMetadata,
  get121OrganizationOptions,
  confirmPublishIntent,
  submitPublishIntent,
  getPublishAudits
};
