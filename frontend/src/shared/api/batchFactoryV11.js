import { apiRequest } from './client.js';

// The module name remains for import compatibility while the production
// workbench has moved to the V12 public contract.
const BASE = '/api/batch-factory/v12';
const LOCAL_EXECUTOR_ARTIFACT_PREFIX = '/api/shuihuo-production/local-executor-artifacts/';
const LOCAL_MERGE_MEDIA_PATTERN = /^\/api\/batch-factory\/v1[12]\/batches\/[^/]+\/merge-media\/[^/]+$/;
const LOCAL_MERGE_COVER_PATTERN = /^\/api\/batch-factory\/v1[12]\/batches\/[^/]+\/merge-cover\/[^/]+$/;

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

export function isProtectedProductionMediaURL(value) {
  return Boolean(localExecutorArtifactRequestPath(value));
}

export function getProductionMediaBlob(mediaUrl) {
  const path = localExecutorArtifactRequestPath(mediaUrl);
  if (!path) return Promise.reject(new Error('Production media URL is not a local executor artifact or local merge artifact'));
  return apiRequest(path, { responseType: 'blob' });
}

export function isBatchFactoryMergeCoverURL(value) {
  return Boolean(mergeCoverRequestPath(value));
}

// Native <img> requests cannot carry the app bearer token. Merge covers stay
// private, so load the small cached JPG through the authenticated API client
// instead of exposing a public video or image URL.
export function getBatchFactoryMergeCoverBlob(coverUrl) {
  const path = mergeCoverRequestPath(coverUrl);
  if (!path) return Promise.reject(new Error('Merge cover URL is not a local Batch Factory cover artifact'));
  return apiRequest(path, { responseType: 'blob', silent: true, suppressGlobalError: true });
}

export function bf11Path(path = '') {
  const clean = String(path || '').replace(/^\/+/, '');
  return clean ? `${BASE}/${clean}` : BASE;
}

export function bf11ScopePath({ scope, batchId, bookId = '', videoId = '' }) {
  if (scope === 'batch') return bf11Path(`batches/${id(batchId)}/settings`);
  if (scope === 'book') return bf11Path(`batches/${id(batchId)}/books/${id(bookId)}/override`);
  if (scope === 'video') return bf11Path(`batches/${id(batchId)}/books/${id(bookId)}/videos/${id(videoId)}/override`);
  throw new Error(`Unsupported V11 scope: ${scope}`);
}

export function getCapabilities() {
  return apiRequest(bf11Path('capabilities'));
}

// Kept for the legacy Batch Factory page. The Shuihuo workbench uses the
// book-scoped image endpoints below, but Vite still statically imports this
// page and must retain its public API contract.
export function generateConfiguredImage(payload = {}) {
  return apiRequest(bf11Path('image-generation'), { method: 'POST', body: body(payload) });
}

export function createNovelFetchIntake(payload) {
  return apiRequest(bf11Path('intakes/novel-fetch'), { method: 'POST', body: body(payload) });
}

export function createManualIntake(payload) {
  return apiRequest(bf11Path('intakes/manual'), { method: 'POST', body: body(payload) });
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
  return apiRequest(bf11Path('fetch-originals'), { method: 'POST', body: body(payload) });
}

export function fetchBookOriginal(batchId, bookId, options = {}) {
  return apiRequest(bf11Path(`batches/${id(batchId)}/books/${id(bookId)}/fetch-original`), {
    method: 'POST',
    body: body({
      ...(Number(options.maxTxt) > 0 ? { maxTxt: Number(options.maxTxt) } : {}),
      ...(String(options.platform || '').trim() ? { platform: String(options.platform).trim() } : {})
    })
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

export function getBatch(batchId) {
  return apiRequest(bf11Path(`batches/${id(batchId)}`));
}

export function deleteBatchFactoryBook(batchId, bookId) {
  return apiRequest(bf11Path(`batches/${id(batchId)}/books/${id(bookId)}`), { method: 'DELETE' });
}

export function deleteBatchFactoryProject(batchId) {
  return apiRequest(bf11Path(`batches/${id(batchId)}`), { method: 'DELETE' });
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
  return apiRequest(`${bookAssetsPath(batchId, bookId)}/images/generate`, { method: 'POST', body: body(payload) });
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
  return apiRequest(`/api/presets?module=${encodeURIComponent(String(module || ''))}`);
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
  return apiRequest(bf11Path(`batches/${id(batchId)}/books/${id(bookId)}/working-front/viral`), { method: 'POST', body: body({ ...payload, content: String(content || '') }) });
}

export function runHook(batchId, bookId, payload = {}) {
  return apiRequest(bf11Path(`batches/${id(batchId)}/books/${id(bookId)}/hook`), { method: 'POST', body: body(payload) });
}

export function approveHook(batchId, bookId, hookId) {
  return apiRequest(bf11Path(`batches/${id(batchId)}/books/${id(bookId)}/hooks/${id(hookId)}/approve`), { method: 'POST', body: body({}) });
}

export function runDirector(batchId, bookId) {
  return apiRequest(bf11Path(`batches/${id(batchId)}/books/${id(bookId)}/director`), { method: 'POST', body: body({}) });
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

export function retryBookStage(batchId, bookId, payload = {}) {
  return apiRequest(bf11Path(`batches/${id(batchId)}/books/${id(bookId)}/stages/retry`), { method: 'POST', body: body(payload) });
}

export function refreshBookSmartUnified(batchId, bookId, payload = {}) {
  return apiRequest(bf11Path(`batches/${id(batchId)}/books/${id(bookId)}/smart-unified/refresh`), { method: 'POST', body: body(payload) });
}

export function runBatchDirector(batchId, payload = {}) {
  return apiRequest(bf11Path(`batches/${id(batchId)}/director`), { method: 'POST', body: body(payload) });
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
  setPrimaryBookAssetImage,
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
