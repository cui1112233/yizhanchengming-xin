const crypto = require('node:crypto');
const http = require('node:http');
const https = require('node:https');
const {
  normalizeSettings,
  normalizeSourceItem,
  appendActivity
} = require('./store');

function now() { return new Date().toISOString(); }
function createId(prefix) { return `${prefix}_${Date.now().toString(36)}_${crypto.randomBytes(4).toString('hex')}`; }

function sign(secret, { username, isOwner, issuedAt, method, pathname }) {
  return crypto.createHmac('sha256', secret)
    .update([username, issuedAt, String(isOwner), method, pathname].join('\n'))
    .digest('hex');
}

function statusActivityMessage(status) {
  const labels = {
    pending: '已加入待制作队列', queued_hook: '爆款开头已进入排队', hook_generating: '正在生成爆款开头',
    hook_review: '爆款开头待审核', queued_director: 'AI 导演已进入排队', director_generating: 'AI 导演生成中',
    complete: 'AI 导演已完成，等待视频生产', failed: '当前制作阶段失败'
  };
  return labels[status] || `状态更新为 ${status || '未知'}`;
}

function clone(value) { return JSON.parse(JSON.stringify(value)); }

function createMySQLBatchFactoryStore({ targetBaseUrl, bridgeSecret, account } = {}) {
  if (!account?.username) throw new Error('当前登录账号不可用');
  const target = new URL(targetBaseUrl || process.env.QIANTIE_GO_BASE_URL || 'http://127.0.0.1:4000');
  const transport = target.protocol === 'https:' ? https : http;
  const secret = bridgeSecret || process.env.QIANTIE_BRIDGE_SECRET || '';

  function request(method, pathname, payload) {
    const issuedAt = String(Math.floor(Date.now() / 1000));
    const body = payload === undefined ? null : Buffer.from(JSON.stringify(payload));
    return new Promise((resolve, reject) => {
      const upstream = transport.request({
        protocol: target.protocol, hostname: target.hostname, port: target.port || undefined, method, path: pathname, timeout: 30_000,
        headers: {
          Accept: 'application/json',
          'X-Qiantie-Username': account.username,
          'X-Qiantie-Is-Owner': String(account.isOwner === true),
          'X-Qiantie-Issued-At': issuedAt,
          'X-Qiantie-Signature': sign(secret, { username: account.username, isOwner: account.isOwner === true, issuedAt, method, pathname }),
          ...(body ? { 'Content-Type': 'application/json', 'Content-Length': String(body.length) } : {})
        }
      }, response => {
        const chunks = [];
        response.on('data', chunk => chunks.push(chunk));
        response.on('end', () => {
          let parsed = {};
          try { parsed = JSON.parse(Buffer.concat(chunks).toString('utf8') || '{}'); } catch (_) { /* bridge error is handled below */ }
          if ((response.statusCode || 500) >= 400) {
            const error = new Error(parsed.error || `批量工厂服务返回 ${response.statusCode}`);
            error.status = response.statusCode || 503;
            reject(error);
            return;
          }
          resolve(parsed);
        });
      });
      upstream.on('timeout', () => upstream.destroy(new Error('批量工厂服务响应超时')));
      upstream.on('error', () => reject(Object.assign(new Error('批量工厂服务暂不可用，请稍后重试'), { status: 503 })));
      if (body) upstream.write(body);
      upstream.end();
    });
  }

  function makeInitialItem(source, timestamp) {
    return {
      id: createId('opening'), ...source, hookDraft: '', approvedHookScript: '', hookMeta: null,
      directorResult: null, promptVersions: {}, videoPromptErrors: {}, productionSubmissionError: null,
      activityLog: [{ at: timestamp, type: 'created', message: '已加入批量工厂，等待使用批次统一设置开始制作', status: 'pending', videoId: '' }],
      status: 'pending', error: '', createdAt: timestamp, updatedAt: timestamp
    };
  }

  async function createNovelFetchIntake(_username, payload = {}) {
    const input = Array.isArray(payload.items) ? payload.items : [];
    if (!input.length) throw new Error('至少选择一个小说获取任务');
    if (input.length > 200) throw new Error('单次最多转入 200 本小说');
    const items = input.map((item, index) => normalizeSourceItem(item, index, { sourceType: 'novel-fetch' }));
    const books = new Set(); const tasks = new Set();
    for (const item of items) {
      if (books.has(item.bookId)) throw new Error(`书ID ${item.bookId} 在本次选择中重复`);
      if (tasks.has(item.sourceTaskId)) throw new Error(`小说获取任务 ${item.sourceTaskId} 在本次选择中重复`);
      books.add(item.bookId); tasks.add(item.sourceTaskId);
    }
    const intake = { id: createId('intake'), sourceType: 'novel-fetch', name: String(payload.name || `小说获取转入 ${items.length} 本`).trim().slice(0, 80), items, createdAt: now(), consumedAt: '', batchId: '' };
    const result = await request('POST', '/api/batch-factory-data/intakes', intake);
    return result.intake || intake;
  }

  async function getIntake(_username, intakeId) {
    try { return (await request('GET', `/api/batch-factory-data/intakes/${encodeURIComponent(String(intakeId))}`)).intake || null; }
    catch (error) { if (error.status === 404) return null; throw error; }
  }

  async function createBatch(_username, payload = {}) {
    const input = Array.isArray(payload.items) ? payload.items : [];
    if (!input.length) throw new Error('至少需要一篇小说开篇');
    if (input.length > 200) throw new Error('单个批次最多 200 篇');
    const timestamp = now();
    const batch = {
      id: createId('batch'), name: String(payload.name || `批量工厂 ${new Date().toLocaleDateString('zh-CN')}`).trim().slice(0, 80),
      mode: payload.mode === 'viral' ? 'viral' : 'original', settings: normalizeSettings(payload.settings),
      sourceIntakeId: String(payload.sourceIntakeId || '').trim(), createdAt: timestamp, updatedAt: timestamp,
      items: input.map((item, index) => makeInitialItem(normalizeSourceItem(item, index), timestamp))
    };
    const result = await request('POST', '/api/batch-factory-data/batches', {
      ...batch,
      items: batch.items.map(item => ({ id: item.id, status: item.status, payload: item, activityLog: item.activityLog }))
    });
    return result.batch || batch;
  }

  async function listBatches() { return (await request('GET', '/api/batch-factory-data/batches')).batches || []; }
  async function getBatch(_username, batchId) {
    try { return (await request('GET', `/api/batch-factory-data/batches/${encodeURIComponent(String(batchId))}`)).batch || null; }
    catch (error) { if (error.status === 404) return null; throw error; }
  }

  async function updateBatchSettings(username, batchId, settings) {
    const result = await request('PUT', `/api/batch-factory-data/batches/${encodeURIComponent(String(batchId))}/settings`, { settings: normalizeSettings(settings) });
    return result.batch || null;
  }

  async function updateItem(username, batchId, itemId, updater) {
    const batch = await getBatch(username, batchId);
    const current = batch?.items?.find(item => item.id === itemId);
    if (!current) return null;
    const item = clone(current); const priorStatus = item.status;
    const result = updater(item, batch) || item;
    if (item.status !== priorStatus) appendActivity(item, { type: item.status === 'failed' ? 'error' : 'status', message: statusActivityMessage(item.status), status: item.status });
    item.updatedAt = now();
    const activity = Array.isArray(item.activityLog) && item.activityLog.length ? item.activityLog[item.activityLog.length - 1] : undefined;
    const response = await request('PUT', `/api/batch-factory-data/batches/${encodeURIComponent(String(batchId))}/items/${encodeURIComponent(String(itemId))}`, { item, activity });
    return response.item || result;
  }

  async function appendItemActivity(username, batchId, itemId, event) {
    return updateItem(username, batchId, itemId, item => { appendActivity(item, event); return item; });
  }

  return { createNovelFetchIntake, getIntake, createBatch, listBatches, getBatch, updateBatchSettings, updateItem, appendItemActivity, normalizeSettings };
}

function createMySQLBatchFactoryStoreFactory(options = {}) {
  return { forAccount: account => createMySQLBatchFactoryStore({ ...options, account }) };
}

module.exports = { createMySQLBatchFactoryStore, createMySQLBatchFactoryStoreFactory };
