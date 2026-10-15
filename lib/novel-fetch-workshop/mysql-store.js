const crypto = require('node:crypto');
const http = require('node:http');
const https = require('node:https');
const { normalizeFetchPolicy, fetchWithPolicy } = require('./fetch-policy');

const DEFAULT_CONFIG = {
  workflow: {
    auto_classify_missing: true,
    auto_fetch_original: true,
    auto_rewrite_after_fetch: true,
    // 对外网站提交必须由用户明确开启，不能因为一次普通改文就自动写入第三方网站。
    auto_submit_after_rewrite: false,
    auto_submit_confirmed: false,
  },
  fetch: { endpoint: 'https://txt.121w.com/api.php', default_max_txt: 4000, timeout_seconds: 30, concurrency: 4, retries: 1, auto_detect_platform: false },
  ai: { base_url: '', api_key: '', model: '', timeout_seconds: 180, max_concurrency: 6, retry_times: 2, max_tokens: 1200, temperature: 0.45, top_p: 0.9, stream: false, json_mode: false, enable_thinking: false, disable_thinking: false, extra_body_json: '' },
  ai_presets: [],
  ai_assignments: { classifier: '__current__', rewrite: '__current__', sensitive_fix: '__current__' },
  rewrite: { default_ai_count: 1, max_ai_count: 5, process_line_count: 5, anchor_line_count: 5, temperature: 0.45, strategy: 'instruction', method_sequence: ['high_imitation', 'opening_instruction', 'instruction'], prompt: '你是短视频小说正文改写师。请只根据原文、风格类型和男女频改写开头部分，保留故事事实、人物关系、时代背景和后续衔接。输出正文，不要解释。' },
  storage: { cleanup_enabled: true, retention_days: 7 },
  parser: { default_parse_mode: 'smart', default_column_preset_id: 'sample_input', custom_column_order: '书籍ID,书名,推荐理由,男女频,标签,评级' }
};

const META_DEFAULTS = {
  bookId: '', bookName: '', paidBookId: '', freeBookId: '', platformId: '', platformName: '', gender: '', genderSource: '', style: '', styleSource: '', tags: '', reason: '', rating: '', sourceLine: '', parseMode: '', parseColumns: [], maxTxt: 4000, aiCount: 1, classifyStatus: '', classifyError: '', classifyConfidence: null, classifyReason: '', classifierModel: '', retryStage: '', status: 'created', originalStatus: '', originalChars: 0, originalRawChars: 0, originalFetchAttempts: 0, originalErrorCode: '', originalUpstreamCode: null, originalErrorMessage: '', originalRefreshErrorMessage: '', originalRefreshErrorCode: '', originalRefreshUpstreamCode: null, originalRefreshAt: '', sensitiveStatus: '', sensitiveHitCount: 0, sensitiveFixedCount: 0, sensitiveFailedCount: 0, aiStatus: '', aiGeneratedCount: 0, aiError: '', aiLastAttemptError: '', siteSubmitStatus: '', siteSubmitError: '', siteSubmitDoneVersions: [], siteSubmitAcceptedVersions: [], siteSubmitFailedVersions: [], error: '', createdAt: '', updatedAt: ''
};

const STYLES = ['古风虐文', '古风甜文', '古风通用', '年代虐文', '年代甜文', '年代通用', '现代虐文', '现代甜文', '现代悬疑', '现代通用', '男频都市', '现代女主', '玄幻', '历史', '爆款BGM', '家庭奇葩', '家庭伤感', '职场打脸'];
const PLATFORMS = [{ id: '1', name: '黑岩付费' }, { id: '2', name: '番茄付费' }, { id: '3', name: '七猫付费' }, { id: '4', name: '点众付费' }, { id: '7', name: '番茄免费' }, { id: '15', name: '知乎付费' }, { id: '20', name: '掌阅付费' }, { id: '26', name: '卓越付费' }, { id: '29', name: '九州书城' }, { id: '31', name: '掌文付费' }];
const BOOK_ID_PATTERN = /^[A-Za-z0-9_.-]+$/;
const DEFAULT_FETCH_ENDPOINT = 'https://txt.121w.com/api.php';

function isPlainObject(value) { return Boolean(value) && typeof value === 'object' && !Array.isArray(value); }
function hasValue(value) { return value !== null && value !== undefined && (typeof value !== 'string' || value.trim() !== ''); }
function normalizeNewlines(text) { return String(text || '').replace(/\r\n/g, '\n').replace(/\r/g, '\n'); }
function dropEmptyLines(text) { return String(text || '').split('\n').filter(line => line.trim() !== '').join('\n'); }
function cleanText(text) { return dropEmptyLines(normalizeNewlines(text)); }
function normalizeMaxTxt(value, fallback = META_DEFAULTS.maxTxt) {
  const number = Number(value);
  const fallbackNumber = Number(fallback);
  const safeFallback = Number.isFinite(fallbackNumber) && fallbackNumber > 0 ? Math.floor(fallbackNumber) : META_DEFAULTS.maxTxt;
  return Number.isFinite(number) && number > 0 ? Math.floor(number) : safeFallback;
}
function limitProcessedOriginal(text, maxTxt) {
  return cleanText(text).slice(0, normalizeMaxTxt(maxTxt));
}
function deepMerge(...objects) {
  const result = {};
  for (const source of objects) for (const [key, value] of Object.entries(source || {})) result[key] = isPlainObject(value) && isPlainObject(result[key]) ? deepMerge(result[key], value) : value;
  return result;
}
function parseRecordObject(value) {
  if (isPlainObject(value)) return value;
  if (typeof value !== 'string' || !value.trim()) return {};
  try {
    const parsed = JSON.parse(value);
    return isPlainObject(parsed) ? parsed : {};
  } catch (_) {
    return {};
  }
}
function normalizeWorkshopRecord(record = {}) {
  const source = isPlainObject(record) ? record : {};
  const meta = parseRecordObject(source.meta);
  const versions = parseRecordObject(source.versions);
  const bookId = String(source.bookId || source.book_id || meta.bookId || meta.book_id || meta.id || '').trim();
  const original = source.original ?? source.original_text ?? '';
  const originalRaw = source.originalRaw ?? source.original_raw ?? '';
  const createdAt = source.createdAt || source.created_at || meta.createdAt || meta.created_at || '';
  const updatedAt = source.updatedAt || source.updated_at || meta.updatedAt || meta.updated_at || '';
  return {
    bookId,
    meta: { ...meta, ...(bookId && !meta.bookId ? { bookId } : {}) },
    original: typeof original === 'string' ? original : '',
    originalRaw: typeof originalRaw === 'string' ? originalRaw : '',
    versions,
    logs: Array.isArray(source.logs) ? source.logs : [],
    hasOriginal: source.hasOriginal === true || source.has_original === true || Boolean(original),
    hasOriginalRaw: source.hasOriginalRaw === true || source.has_original_raw === true || Boolean(originalRaw),
    hasAi: source.hasAi === true || source.has_ai === true || Object.keys(versions).some(key => /^ai\d+$/.test(key)),
    createdAt,
    updatedAt,
    status: source.status ?? meta.status ?? '',
    error: source.error ?? meta.error ?? ''
  };
}
function now() { return new Date().toISOString(); }
function assertBookID(bookId) {
  const value = String(bookId || '').trim();
  if (!BOOK_ID_PATTERN.test(value)) throw new Error('非法的书籍ID：' + value);
  return value;
}
function sign(secret, { username, isOwner, issuedAt, method, pathname }) {
  return crypto.createHmac('sha256', secret).update([username, issuedAt, String(isOwner), method, pathname].join('\n')).digest('hex');
}

function buildFetchURL(endpoint, { bookId, platformId, maxTxt }) {
  const template = String(endpoint || DEFAULT_FETCH_ENDPOINT).trim() || DEFAULT_FETCH_ENDPOINT;
  const variables = { bookid: String(bookId), platform: String(platformId), max_txt: String(maxTxt) };
  const substituted = template.replace(/\{(bookid|platform|max_txt)\}/g, (_, key) => encodeURIComponent(variables[key]));
  const url = new URL(substituted);
  // 模板模式让接入方自行决定参数名；普通 URL 则保持既有 121w 参数兼容。
  if (!/\{(?:bookid|platform|max_txt)\}/.test(template)) {
    if (!url.searchParams.has('bookid')) url.searchParams.set('bookid', variables.bookid);
    if (!url.searchParams.has('platform')) url.searchParams.set('platform', variables.platform);
    if (!url.searchParams.has('max_txt')) url.searchParams.set('max_txt', variables.max_txt);
  }
  return url.toString();
}

function extractFetchedText(response) {
  const payload = isPlainObject(response) ? response : {};
  const candidates = [payload.text, payload.data, payload.content, payload.data?.text, payload.data?.content];
  return candidates.find(value => typeof value === 'string' && value.trim()) || '';
}

function upstreamFetchError(response) {
  const payload = isPlainObject(response) ? response : {};
  if (Number(payload.code) >= 400) {
    const upstreamCode = Number(payload.code);
    const upstreamMessage = payload.msg || `上游抓取失败（${upstreamCode}）`;
    const error = new Error(`上游返回 ${upstreamCode}：${upstreamMessage}`);
    error.code = 'UPSTREAM_FETCH_ERROR';
    error.upstream = true;
    error.upstreamCode = upstreamCode;
    error.upstreamMessage = upstreamMessage;
    return error;
  }
  return null;
}

function createMySQLWorkshopStore({ targetBaseUrl, bridgeSecret, account, fetchUpstream } = {}) {
  if (!account?.username) throw new Error('当前登录账号不可用');
  const target = new URL(targetBaseUrl || process.env.QIANTIE_GO_BASE_URL || 'http://127.0.0.1:4000');
  const secret = bridgeSecret || process.env.QIANTIE_BRIDGE_SECRET || '';
  const transport = target.protocol === 'https:' ? https : http;

  function request(method, pathname, payload) {
    const issuedAt = String(Math.floor(Date.now() / 1000));
    const body = payload === undefined ? null : Buffer.from(JSON.stringify(payload));
    return new Promise((resolve, reject) => {
      const upstream = transport.request({ protocol: target.protocol, hostname: target.hostname, port: target.port || undefined, method, path: pathname, timeout: 30_000, headers: {
        Accept: 'application/json',
        'X-Qiantie-Username': account.username,
        'X-Qiantie-Is-Owner': String(account.isOwner === true),
        'X-Qiantie-Issued-At': issuedAt,
        'X-Qiantie-Signature': sign(secret, { username: account.username, isOwner: account.isOwner === true, issuedAt, method, pathname }),
        ...(body ? { 'Content-Type': 'application/json', 'Content-Length': String(body.length) } : {})
      } }, response => {
        const chunks = [];
        response.on('data', chunk => chunks.push(chunk));
        response.on('end', () => {
          const raw = Buffer.concat(chunks).toString('utf8');
          let parsed = {};
          try { parsed = raw ? JSON.parse(raw) : {}; } catch (_) { parsed = {}; }
          if ((response.statusCode || 500) >= 400) {
            const error = new Error(parsed.error || `小说获取工作台服务返回 ${response.statusCode}`);
            error.status = response.statusCode || 503;
            reject(error);
            return;
          }
          resolve(parsed);
        });
      });
      upstream.on('timeout', () => upstream.destroy(new Error('小说获取工作台服务响应超时')));
      upstream.on('error', () => reject(Object.assign(new Error('小说获取工作台服务暂不可用，请稍后重试'), { status: 503 })));
      if (body) upstream.write(body);
      upstream.end();
    });
  }

  function pathFor(bookId) { return `/api/novel-fetch-workshop/tasks/${encodeURIComponent(assertBookID(bookId))}`; }
  function bodyPathFor(bookId, versionId) {
    return `${pathFor(bookId)}/bodies/${encodeURIComponent(String(versionId || '').trim())}`;
  }
  async function readBody(bookId, versionId) {
    try {
      const body = await request('GET', bodyPathFor(bookId, versionId));
      return typeof body.content === 'string' ? body.content : '';
    } catch (error) {
      if (error.status === 404) return '';
      throw error;
    }
  }
  async function writeBody(bookId, versionId, content, state = 'ready') {
    return request('PUT', bodyPathFor(bookId, versionId), { content: String(content || ''), state });
  }
  async function readDocument(bookId) {
    const record = normalizeWorkshopRecord(await request('GET', pathFor(bookId)));
    const response = await request('GET', `${pathFor(record.bookId)}/bodies`);
    const versionIds = (response.bodies || []).map(body => String(body?.versionId || '').trim()).filter(Boolean);
    const bodyEntries = await Promise.all(versionIds.map(async versionId => [versionId, await readBody(record.bookId, versionId)]));
    const bodyTexts = Object.fromEntries(bodyEntries);
    const original = bodyTexts.original || record.original;
    const versions = { ...record.versions };
    for (const [versionId, content] of bodyEntries) {
      if (versionId !== 'original') versions[versionId] = content;
    }
    return {
      bookId: record.bookId,
      meta: record.meta,
      original,
      originalRaw: record.originalRaw,
      versions,
      logs: record.logs,
      gender: record.gender,
      genderSource: record.genderSource,
      genderReason: record.genderReason,
      bookInfo: record.bookInfo,
      processedText: record.processedText,
      originalRawChars: record.originalRawChars,
      processedChars: record.processedChars,
      maxTxt: record.maxTxt
    };
  }
  async function writeDocument(document) {
    const bookId = assertBookID(document.bookId);
    const original = String(document.original || '');
    const versions = isPlainObject(document.versions) ? document.versions : {};
    // Go stores large text in the body table; task JSON remains metadata-only.
    await request('PUT', pathFor(bookId), {
      meta: document.meta || {},
        ...Object.fromEntries(Object.entries({
          gender: document.gender,
          genderSource: document.genderSource,
          genderReason: document.genderReason,
          bookInfo: document.bookInfo,
          originalRaw: document.originalRaw,
          processedText: document.processedText,
          originalRawChars: document.originalRawChars,
          processedChars: document.processedChars,
          maxTxt: document.maxTxt
        }).filter(([, value]) => value !== undefined && value !== '')),
        versions: Object.fromEntries(Object.keys(versions).map(versionId => [versionId, { versionId, state: 'ready' }])),
      logs: document.logs || []
    });
    if (original) await writeBody(bookId, 'original', original);
    for (const [versionId, content] of Object.entries(versions)) {
      if (typeof content === 'string') await writeBody(bookId, versionId, content);
    }
  }
  async function existingOrNew(bookId) {
    try { return await readDocument(bookId); } catch (error) { if (error.status === 404) return { bookId, meta: {}, original: '', originalRaw: '', versions: {}, logs: [] }; throw error; }
  }
  async function appendLog(username, bookId, event, data) {
    const document = await existingOrNew(assertBookID(bookId));
    document.logs.push({ time: now(), event, data: isPlainObject(data) ? data : {} });
    await writeDocument(document);
  }
  async function updateTaskMeta(username, bookId, patch) {
    const document = await readDocument(assertBookID(bookId));
    document.meta = { ...document.meta, ...(isPlainObject(patch) ? patch : {}), updatedAt: now() };
    await writeDocument(document);
    return document.meta;
  }
  async function getTask(username, bookId) {
    try {
      const document = await readDocument(assertBookID(bookId));
      return { meta: document.meta, hasOriginalRaw: Boolean(document.originalRaw), document };
    } catch (error) { if (error.status === 404) return null; throw error; }
  }
  async function writeRunAudit(record) { return request('POST', '/api/novel-fetch-workshop/runs', record); }
  async function listTasks(username) {
    const response = await request('GET', '/api/novel-fetch-workshop/tasks');
    return (response.tasks || []).map(rawRecord => {
      const record = normalizeWorkshopRecord(rawRecord);
      return {
        ...record.meta,
        bookId: record.bookId,
        hasOriginal: record.hasOriginal,
        hasOriginalRaw: record.hasOriginalRaw,
        hasAi: record.hasAi,
        createdAt: record.createdAt,
        updatedAt: record.updatedAt,
        status: record.status,
        error: record.error,
        maxTxt: normalizeMaxTxt(record.meta.maxTxt, META_DEFAULTS.maxTxt),
        originalChars: record.original ? record.original.length : (Number(record.meta.originalChars) || 0),
        originalRawChars: record.originalRaw ? record.originalRaw.length : (Number(record.meta.originalRawChars) || 0)
      };
    });
  }
  async function saveTasks(username, tasks) {
    let saved = 0;
    const merged = new Map();
    for (const task of Array.isArray(tasks) ? tasks : []) {
      if (!isPlainObject(task) || !String(task.bookId || '').trim()) continue;
      const bookId = assertBookID(task.bookId);
      merged.set(bookId, { ...(merged.get(bookId) || {}), ...Object.fromEntries(Object.entries(task).filter(([, value]) => hasValue(value))) });
    }
    for (const [bookId, task] of merged) {
      const document = await existingOrNew(bookId);
      const meta = { ...META_DEFAULTS, ...document.meta };
      for (const [key, value] of Object.entries(task)) if (key !== 'bookId' && hasValue(value)) meta[key] = value;
      meta.bookId = bookId;
      meta.createdAt = meta.createdAt || now();
      meta.updatedAt = now();
      document.meta = meta;
      document.logs.push({ time: now(), event: 'task_saved', data: { bookId, bookName: meta.bookName || '' } });
      await writeDocument(document);
      saved++;
    }
    return { saved, duplicates: Math.max(0, (Array.isArray(tasks) ? tasks.length : 0) - merged.size) };
  }
  // Batch Factory fetches a saved Book ID before it has a workshop task. Keep
  // this path on the same configured endpoint, retry policy and response
  // validation as fetchOriginal, but do not create a second task or fabricate
  // source text when the upstream cannot return it.
  async function fetchDirectOriginal({ bookId, platformId, maxTxt } = {}) {
    const resolvedBookId = assertBookID(bookId);
    const resolvedPlatformId = String(platformId || '').trim();
    if (!resolvedPlatformId) throw new Error('缺少抓取平台');
    const configured = (await request('GET', '/api/novel-fetch-workshop/config')).settings || {};
    const fetchConfig = isPlainObject(configured.fetch) ? configured.fetch : {};
    const effectiveMaxTxt = normalizeMaxTxt(maxTxt, fetchConfig.default_max_txt || META_DEFAULTS.maxTxt);
    const result = await fetchWithPolicy({
      bookId: resolvedBookId,
      platformId: resolvedPlatformId,
      maxTxt: effectiveMaxTxt,
      fetchConfig,
      fetchUpstream: async (receivedBookId, receivedPlatformId, requestedMaxTxt, options) => {
        let candidate;
        if (fetchUpstream) {
          candidate = await fetchUpstream(receivedBookId, receivedPlatformId, requestedMaxTxt, options);
        } else {
          const controller = new AbortController();
          const timer = setTimeout(() => controller.abort(), options.timeoutMs);
          try {
            const response = await fetch(buildFetchURL(options.endpoint, {
              bookId: receivedBookId,
              platformId: receivedPlatformId,
              maxTxt: requestedMaxTxt
            }), { signal: controller.signal });
            if (!response.ok) {
              const error = new Error(`上游抓取失败（HTTP ${response.status}）`);
              error.code = 'UPSTREAM_HTTP_ERROR';
              error.recoverable = true;
              error.upstreamCode = response.status;
              throw error;
            }
            candidate = await response.json();
          } catch (error) {
            if (error?.name === 'AbortError') {
              const timeoutError = new Error(`原文抓取超时（${Math.floor(options.timeoutMs / 1000)} 秒）`);
              timeoutError.code = 'UPSTREAM_TIMEOUT';
              timeoutError.recoverable = true;
              throw timeoutError;
            }
            throw error;
          } finally {
            clearTimeout(timer);
          }
        }
        const upstreamError = upstreamFetchError(candidate);
        if (upstreamError) {
          upstreamError.recoverable = true;
          throw upstreamError;
        }
        const text = extractFetchedText(candidate);
        if (!text) {
          const error = new Error('上游未返回正文内容');
          error.code = 'EMPTY_ORIGINAL';
          error.recoverable = true;
          throw error;
        }
        return { ...candidate, text };
      }
    });
    const rawText = String(result.text || '');
    return {
      bookId: resolvedBookId,
      platformId: resolvedPlatformId,
      text: limitProcessedOriginal(rawText, effectiveMaxTxt),
      rawText,
      attempts: result.attempts,
      bookinfo: result.bookinfo || {}
    };
  }
  async function fetchOriginal(username, bookId, maxTxt) {
    const document = await readDocument(assertBookID(bookId));
    try {
      const configured = (await request('GET', '/api/novel-fetch-workshop/config')).settings || {};
      const fetchConfig = isPlainObject(configured.fetch) ? configured.fetch : {};
      const effectiveMaxTxt = normalizeMaxTxt(maxTxt, document.meta.maxTxt || fetchConfig.default_max_txt || META_DEFAULTS.maxTxt);
      const policy = normalizeFetchPolicy(fetchConfig);
      const requestedPlatform = String(document.meta.platformId || '').trim();
      if (!requestedPlatform) {
        const error = new Error('任务缺少抓取平台');
        error.recoverable = false;
        throw error;
      }

      const result = await fetchWithPolicy({
        bookId: document.bookId,
        platformId: requestedPlatform,
        maxTxt: effectiveMaxTxt,
        fetchConfig,
        fetchUpstream: async (receivedBookId, platformId, requestedMaxTxt, options) => {
          let candidate;
          if (fetchUpstream) {
            candidate = await fetchUpstream(receivedBookId, platformId, requestedMaxTxt, options);
          } else {
            const controller = new AbortController();
            const timer = setTimeout(() => controller.abort(), options.timeoutMs);
            try {
              const response = await fetch(buildFetchURL(options.endpoint, {
                bookId: receivedBookId,
                platformId,
                maxTxt: requestedMaxTxt
              }), { signal: controller.signal });
              if (!response.ok) {
                const error = new Error(`上游抓取失败（HTTP ${response.status}）`);
                error.code = 'UPSTREAM_HTTP_ERROR';
                error.recoverable = true;
                error.upstreamCode = response.status;
                throw error;
              }
              candidate = await response.json();
            } catch (error) {
              if (error?.name === 'AbortError') {
                const timeoutError = new Error(`原文抓取超时（${Math.floor(options.timeoutMs / 1000)} 秒）`);
                timeoutError.code = 'UPSTREAM_TIMEOUT';
                timeoutError.recoverable = true;
                throw timeoutError;
              }
              throw error;
            } finally {
              clearTimeout(timer);
            }
          }

          const upstreamError = upstreamFetchError(candidate);
          if (upstreamError) {
            upstreamError.recoverable = true;
            throw upstreamError;
          }
          const text = extractFetchedText(candidate);
          if (!text) {
            const error = new Error('上游未返回正文内容');
            error.code = 'EMPTY_ORIGINAL';
            error.recoverable = true;
            throw error;
          }
          return { ...candidate, text };
        }
      });

      const raw = result.text;
      document.originalRaw = raw;
      document.original = limitProcessedOriginal(raw, effectiveMaxTxt);
      document.meta = {
        ...document.meta,
        maxTxt: effectiveMaxTxt,
        platformAutoDetected: false,
        originalFetchAttempts: result.attempts,
        originalErrorCode: '',
        originalUpstreamCode: null,
        originalErrorMessage: '',
        originalRefreshErrorMessage: '',
        originalRefreshErrorCode: '',
        originalRefreshUpstreamCode: null,
        originalRefreshAt: '',
        originalStatus: 'done', originalRawChars: raw.length, originalChars: document.original.length,
        originalFetchedAt: now(), originalProcessedAt: now(), status: 'original_done',
        bookinfo: result?.bookinfo || document.meta.bookinfo, updatedAt: now()
      };
      if (!document.meta.bookName && result?.bookinfo?.book_name) document.meta.bookName = String(result.bookinfo.book_name);
      document.logs.push({
        time: now(),
        event: 'original_fetched',
        data: {
          bookId: document.bookId,
          originalChars: document.original.length,
          requestedPlatform,
          resolvedPlatform: requestedPlatform,
          platformAutoDetected: false,
          timeoutSeconds: Math.floor(policy.timeoutMs / 1000),
          retries: policy.retries,
          attempts: result.attempts
        }
      });
      await writeDocument(document);
      return { status: 'done', attempts: result.attempts };
    } catch (error) {
      const attempts = Math.max(1, Number(error?.attempts) || 1);
      const code = typeof error?.code === 'string' && error.code.trim() ? error.code : 'ORIGINAL_FETCH_FAILED';
      const message = error.message || String(error);
      if (String(document.original || '').trim()) {
        document.meta = {
          ...document.meta,
          platformAutoDetected: false,
          originalStatus: 'done',
          status: 'original_done',
          originalRefreshErrorMessage: message,
          originalRefreshErrorCode: code,
          originalRefreshUpstreamCode: Number.isFinite(Number(error?.upstreamCode)) ? Number(error.upstreamCode) : null,
          originalRefreshAt: now(),
          updatedAt: now()
        };
        document.logs.push({ time: now(), event: 'original_refresh_failed', data: { bookId: document.bookId, error: message, code, upstreamCode: document.meta.originalRefreshUpstreamCode, attempts } });
        await writeDocument(document);
        return { status: 'refresh_failed', attempts, code, preserved: true };
      }
      document.meta = { ...document.meta, platformAutoDetected: false, originalStatus: 'failed', status: 'original_failed', originalFetchAttempts: attempts, originalErrorCode: code, originalUpstreamCode: Number.isFinite(Number(error?.upstreamCode)) ? Number(error.upstreamCode) : null, originalErrorMessage: message, error: message, updatedAt: now() };
      document.logs.push({ time: now(), event: 'original_fetch_failed', data: { bookId: document.bookId, error: message, code, upstreamCode: document.meta.originalUpstreamCode, attempts } });
      await writeDocument(document);
      return { status: 'failed', attempts, code };
    }
  }
  async function readOriginal(username, bookId) { return (await readDocument(assertBookID(bookId))).original; }
  async function saveOriginalText(username, bookId, text) {
    const document = await readDocument(assertBookID(bookId));
    document.original = String(text || '');
    document.meta = { ...document.meta, originalChars: document.original.length, originalProcessedAt: now(), updatedAt: now() };
    document.logs.push({ time: now(), event: 'original_rules_applied', data: { bookId: document.bookId, originalChars: document.original.length } });
    await writeDocument(document);
  }
  async function restoreOriginal(username, bookId) {
    const document = await readDocument(assertBookID(bookId));
    if (!document.originalRaw) throw new Error('没有可恢复的原始备份');
    const effectiveMaxTxt = normalizeMaxTxt(document.meta.maxTxt, META_DEFAULTS.maxTxt);
    document.original = limitProcessedOriginal(document.originalRaw, effectiveMaxTxt);
    document.meta = { ...document.meta, maxTxt: effectiveMaxTxt, originalChars: document.original.length, originalProcessedAt: now(), status: 'original_restored', updatedAt: now() };
    document.logs.push({ time: now(), event: 'original_restored', data: { bookId: document.bookId, originalChars: document.original.length } });
    await writeDocument(document);
    return { status: 'original_restored', originalChars: document.original.length };
  }
  async function readLogs(username, bookId) { return (await readDocument(assertBookID(bookId))).logs; }
  async function readVersionText(username, bookId, version) { const document = await readDocument(assertBookID(bookId)); return version === 'original' || version === 'edited' ? document.original : (document.versions[version] || ''); }
  async function saveVersionText(username, bookId, version, text) { const document = await readDocument(assertBookID(bookId)); document.versions[version] = String(text || ''); await writeDocument(document); }
  async function deleteTasks(username, ids) { const response = await request('DELETE', '/api/novel-fetch-workshop/tasks', { ids: Array.isArray(ids) ? ids.map(assertBookID) : [] }); return response; }
  async function appendSiteSubmitLog(username, bookId, entry) { await appendLog(username, assertBookID(bookId), 'site_submit', entry); return entry; }
  async function readSiteSubmitLog(username, bookId) { return (await readLogs(username, assertBookID(bookId))).filter(item => item && item.event === 'site_submit').map(item => item.data || {}); }
  async function readSiteSubmitResult(username, bookId) { const rows = await readSiteSubmitLog(username, bookId); return rows.length ? rows[rows.length - 1] : {}; }

  return {
    getConfig: async () => deepMerge(DEFAULT_CONFIG, (await request('GET', '/api/novel-fetch-workshop/config')).settings || {}),
    saveConfig: async settings => request('PUT', '/api/novel-fetch-workshop/config', { settings: deepMerge(DEFAULT_CONFIG, settings || {}) }),
    getPlatforms: () => PLATFORMS,
    getStyles: () => STYLES,
    saveTasks, listTasks, getTask, updateTaskMeta, fetchDirectOriginal, fetchOriginal, readOriginal, saveOriginalText, restoreOriginal, readLogs, readVersionText, saveVersionText, appendLog, appendSiteSubmitLog, readSiteSubmitLog, readSiteSubmitResult, deleteTasks, writeRunAudit
  };
}

module.exports = { createMySQLWorkshopStore, DEFAULT_CONFIG, DEFAULT_FETCH_ENDPOINT, buildFetchURL, extractFetchedText, upstreamFetchError, normalizeNewlines, dropEmptyLines, cleanText, normalizeMaxTxt, limitProcessedOriginal };
