const express = require('express');
const crypto = require('crypto');
const http = require('http');
const https = require('https');
const { apiAuth } = require('../middleware/auth');
const { getVideoApiKey, readConfig } = require('../lib/shared');
const { normalizeModelCatalog } = require('../lib/model-catalog');
const { resolveRuntimeModel } = require('../lib/model-catalog-runtime');
const { H3_MODEL_KEY } = require('../lib/video-model-catalog');
const {
  YFAI_SEEDANCE_MODEL,
  buildYfaiSeedancePayload,
  submitYfaiSeedance,
  requestYfaiSeedanceTask,
  parseYfaiTaskResponse,
  parseYfaiTaskStatus
} = require('../lib/yfai-seedance-adapter');
const { h3ReferenceUrl } = require('../lib/reference-asset-public-url');
const {
  H3_API_BASE_URL,
  H3_DEFAULT_WORKFLOW,
  buildH3Request,
  defaultH3Request,
  defaultH3Submit,
  h3ResultURL,
  h3TaskState,
  readH3TaskID
} = require('../lib/h3-video-adapter');

const YD_CREATE_URL = 'https://ydapi.yadiai.cn/openapi/v1/video/create';
const YD_TASKS_URL = 'https://ydapi.yadiai.cn/openapi/v1/video/tasks';
const DEFAULT_FIRST_FRAME_URL = 'https://tvmao-public.tos-cn-beijing.volces.com/tapnow/empty.png';
const MAX_PROMPT_LENGTH = 12000;
const MAX_OPTIONAL_IMAGES = 3;
const MAX_H3_PROMPT_LENGTH = 10000;
const REFERENCE_TOS_SYNC_TIMEOUT_MS = 3000;
const H3_TASK_PREFIX = 'h3:';
const YFAI_TASK_PREFIX = 'yfai:';
const YD_MODEL_KEY = 'yd2-mini-video';

function acceptsScriptImages(modelKey) {
  return [YD_MODEL_KEY, H3_MODEL_KEY, YFAI_SEEDANCE_MODEL].includes(String(modelKey || '').trim());
}

async function providerAccessibleImageUrls(req, value) {
  const requestedImages = Array.isArray(value) ? value : [];
  const assetStore = req.app?.locals?.novelPanelPremiumStore;
  return Promise.all(requestedImages.map(async imageURL => {
    let tosUrl = null;
    try {
      const sync = assetStore?.syncReferenceAssetUrlToTos?.(req.username, imageURL);
      tosUrl = sync ? await Promise.race([
        Promise.resolve(sync).catch(() => null),
        new Promise(resolve => setTimeout(() => resolve(null), REFERENCE_TOS_SYNC_TIMEOUT_MS))
      ]) : null;
    } catch (_) {
      // A configured-but-unreachable TOS endpoint must not prevent the signed public-link fallback.
    }
    return tosUrl || h3ReferenceUrl(imageURL, req.username);
  }));
}

function readTaskID(payload) {
  const candidates = [payload?.task_id, payload?.taskId, payload?.id, payload?.data?.task_id, payload?.data?.taskId, payload?.data?.id, payload?.result?.task_id, payload?.result?.taskId];
  const value = candidates.find(item => typeof item === 'string' || typeof item === 'number');
  return value === undefined ? '' : String(value).trim();
}

function validOptionalImageURLs(value) {
  if (value === undefined) return [];
  if (!Array.isArray(value) || value.length > MAX_OPTIONAL_IMAGES) throw new Error('可选图片最多 3 张');
  return value.map(item => {
    const imageURL = String(item || '').trim();
    if (!imageURL) throw new Error('可选图片地址不能为空');
    const parsed = new URL(imageURL);
    if (parsed.protocol !== 'https:') throw new Error('可选图片必须使用 HTTPS 地址');
    return parsed.toString();
  });
}

function validH3ReferenceImageURLs(value) {
  if (value === undefined) return [];
  if (!Array.isArray(value) || value.length > 9) throw new Error('H3 最多支持 9 张参考图片');
  return value.map(item => String(item || '').trim()).filter(Boolean).map(imageURL => {
    const parsed = new URL(imageURL);
    if (parsed.protocol !== 'https:') throw new Error('H3 参考图片必须使用 HTTPS 地址');
    return parsed.toString();
  });
}

function h3ApiKeyFromEnvironment() {
  return String(process.env.QIANTIE_AUTODL_H3_API_KEY || process.env.QIANTIE_H3_API_KEY || '').trim();
}

function h3ApiKeyForRequest(req, configReader, h3ApiKeyReader) {
  const personalKey = getVideoApiKey(configReader?.(req.username), 'h3');
  return personalKey || String(h3ApiKeyReader?.(req) || '').trim();
}

function h3WorkflowForRequest(req, configReader = readConfig) {
  const config = configReader?.(req?.username) || {};
  const catalog = normalizeModelCatalog(config.modelCatalog, config);
  return catalog.find(model => model.id === H3_MODEL_KEY && model.kind === 'video')?.workflowId || H3_DEFAULT_WORKFLOW;
}

function h3TaskID(rawTaskID) {
  const value = String(rawTaskID || '').trim();
  return value.startsWith(H3_TASK_PREFIX) ? value.slice(H3_TASK_PREFIX.length).trim() : '';
}

function publicH3TaskID(rawTaskID) {
  return `${H3_TASK_PREFIX}${String(rawTaskID || '').trim()}`;
}

function yfaiTaskID(rawTaskID) {
  const value = String(rawTaskID || '').trim();
  return value.startsWith(YFAI_TASK_PREFIX) ? value.slice(YFAI_TASK_PREFIX.length).trim() : '';
}

function publicYfaiTaskID(rawTaskID) {
  return `${YFAI_TASK_PREFIX}${String(rawTaskID || '').trim()}`;
}

function h3Duration(value) {
  const duration = value === undefined || value === null || value === '' ? 5 : Number(value);
  if (!Number.isInteger(duration) || duration < 1 || duration > 15) throw new Error('H3 视频时长必须是 1-15 秒的整数');
  return duration;
}

function h3Resolution(value) {
  const resolution = String(value || '480p竖').trim();
  if (!['480p竖', '768p竖', '480p横', '768p横'].includes(resolution)) throw new Error('H3 分辨率不受支持');
  return resolution;
}

function resolveH3VideoMedia({ aspectRatio, resolution } = {}) {
  const legacyResolution = String(resolution || '').trim();
  if (['480p竖', '768p竖', '480p横', '768p横'].includes(legacyResolution)) {
    return {
      aspectRatio: legacyResolution.endsWith('横') ? '16:9' : '9:16',
      resolution: legacyResolution
    };
  }
  const normalizedAspectRatio = String(aspectRatio || '').trim() === '16:9' ? '16:9' : '9:16';
  const resolutionTier = String(resolution || '').trim() === '480p' ? '480p' : '768p';
  return {
    aspectRatio: normalizedAspectRatio,
    resolution: `${resolutionTier}${normalizedAspectRatio === '16:9' ? '横' : '竖'}`
  };
}

function defaultSubmit({ apiKey, payload }) {
  return new Promise((resolve, reject) => {
    const body = JSON.stringify(payload);
    const request = https.request(YD_CREATE_URL, { method: 'POST', timeout: 90000, headers: { 'Content-Type': 'application/json', 'Content-Length': Buffer.byteLength(body), Authorization: `Bearer ${apiKey}` } }, response => {
      const chunks = [];
      response.on('data', chunk => chunks.push(chunk));
      response.on('end', () => resolve({ statusCode: response.statusCode || 500, text: Buffer.concat(chunks).toString('utf8') }));
      response.on('error', reject);
    });
    request.on('timeout', () => request.destroy(new Error('视频服务响应超时')));
    request.on('error', reject);
    request.write(body);
    request.end();
  });
}

function upstreamRequest(url, apiKey) {
  return new Promise((resolve, reject) => {
    const request = https.request(url, { method: 'GET', timeout: 30000, headers: { Authorization: `Bearer ${apiKey}` } }, response => {
      const chunks = [];
      response.on('data', chunk => chunks.push(chunk));
      response.on('end', () => resolve({ statusCode: response.statusCode || 500, headers: response.headers, text: Buffer.concat(chunks).toString('utf8') }));
      response.on('error', reject);
    });
    request.on('timeout', () => request.destroy(new Error('视频服务响应超时')));
    request.on('error', reject);
    request.end();
  });
}

function payloadOf(body) {
  const payload = body?.data ?? body?.result ?? body;
  return payload && typeof payload === 'object' ? payload : {};
}

function taskState(body) {
  const payload = payloadOf(body);
  const nested = payload.task && typeof payload.task === 'object' ? payload.task : {};
  return String(payload.status || payload.state || nested.status || nested.state || '').trim().toUpperCase();
}

function taskError(body) {
  const payload = payloadOf(body);
  const nested = payload.task && typeof payload.task === 'object' ? payload.task : {};
  return String(payload.errorMessage || payload.error || payload.message || nested.errorMessage || nested.error || nested.message || '视频生成失败').trim();
}

function resultURL(body) {
  const payload = payloadOf(body);
  const candidates = [
    ...(Array.isArray(payload.urls) ? payload.urls : []),
    ...(Array.isArray(payload.outputs) ? payload.outputs.map(item => item?.url) : []),
    payload.url, payload.video_url, payload.videoUrl
  ];
  for (const candidate of candidates) {
    try {
      const parsed = new URL(String(candidate || '').trim());
      if (parsed.protocol === 'https:') return parsed.toString();
    } catch { /* ignore malformed result URLs */ }
  }
  return '';
}

function bridgeJSON(gateway, account, method, pathname, payload) {
  const target = new URL(gateway?.targetBaseUrl || process.env.QIANTIE_GO_BASE_URL || 'http://127.0.0.1:4000');
  const secret = gateway?.bridgeSecret || process.env.QIANTIE_BRIDGE_SECRET || '';
  const body = payload === undefined ? null : Buffer.from(JSON.stringify(payload));
  const issuedAt = String(Math.floor(Date.now() / 1000));
  const signature = crypto.createHmac('sha256', secret).update([account.username, issuedAt, String(account.isOwner === true), method, pathname].join('\n')).digest('hex');
  const transport = target.protocol === 'https:' ? https : http;
  return new Promise((resolve, reject) => {
    const request = transport.request({ protocol: target.protocol, hostname: target.hostname, port: target.port || undefined, method, path: pathname, headers: { Accept:'application/json','X-Qiantie-Username':account.username,'X-Qiantie-Is-Owner':String(account.isOwner===true),'X-Qiantie-Issued-At':issuedAt,'X-Qiantie-Signature':signature,...(body?{'Content-Type':'application/json','Content-Length':String(body.length)}:{}) } }, response => { const chunks=[]; response.on('data',chunk=>chunks.push(chunk)); response.on('end',()=>{ let data={}; try{data=JSON.parse(Buffer.concat(chunks).toString('utf8')||'{}')}catch{}; if((response.statusCode||500)>=400){const error=new Error(data.error||'本地执行器服务不可用');error.status=response.statusCode;return reject(error)} resolve(data) }) });
    request.on('error',()=>reject(Object.assign(new Error('本地执行器服务暂不可用'),{status:503}))); if(body)request.write(body); request.end();
  });
}

function bridgeDownload(gateway, account, pathname, res) {
  const target = new URL(gateway?.targetBaseUrl || process.env.QIANTIE_GO_BASE_URL || 'http://127.0.0.1:4000'); const secret = gateway?.bridgeSecret || process.env.QIANTIE_BRIDGE_SECRET || ''; const issuedAt=String(Math.floor(Date.now()/1000)); const signature=crypto.createHmac('sha256',secret).update([account.username,issuedAt,String(account.isOwner===true),'GET',pathname].join('\n')).digest('hex'); const transport=target.protocol==='https:'?https:http;
  const upstream=transport.request({protocol:target.protocol,hostname:target.hostname,port:target.port||undefined,method:'GET',path:pathname,headers:{'X-Qiantie-Username':account.username,'X-Qiantie-Is-Owner':String(account.isOwner===true),'X-Qiantie-Issued-At':issuedAt,'X-Qiantie-Signature':signature}},response=>{res.status(response.statusCode||502); if(response.headers['content-type'])res.setHeader('Content-Type',response.headers['content-type']); response.pipe(res)}); upstream.on('error',()=>res.status(503).json({error:'视频下载服务暂不可用'})); upstream.end();
}

function createScriptVideoRouter({
  configReader = readConfig,
  submit = defaultSubmit,
  request = upstreamRequest,
  h3Submit = defaultH3Submit,
  h3Request = defaultH3Request,
  yfaiSubmit = submitYfaiSeedance,
  yfaiRequest = requestYfaiSeedanceTask,
  h3ApiKeyReader = h3ApiKeyFromEnvironment,
  authenticate = apiAuth,
  shuihuoGateway,
  memberStore,
  accountStore
} = {}) {
  const router = express.Router();
  router.use(authenticate);
  function resolveYfaiModel(req) {
    return resolveRuntimeModel({
      username: req.username,
      kind: 'video',
      modelId: YFAI_SEEDANCE_MODEL,
      memberStore: memberStore || req.app?.locals?.memberStore,
      accountStore,
      account: req.auth?.account,
      configReader
    });
  }
  router.post('/', async (req, res) => {
    const resolvedMemberStore = memberStore || req.app?.locals?.memberStore;
    if (!resolvedMemberStore?.canUseApi(req.auth.username, 'video')) {
      return res.status(403).json({ error: '暂无视频生成权限' });
    }
    const prompt = String(req.body?.prompt || '').trim();
    if (!prompt) return res.status(400).json({ error: '分镜视频提示词不能为空' });
    if (prompt.length > MAX_PROMPT_LENGTH) return res.status(400).json({ error: `分镜视频提示词不能超过 ${MAX_PROMPT_LENGTH} 个字符` });
    const requestedReferenceImages = Array.isArray(req.body?.imageUrls)
      ? req.body.imageUrls.filter(item => String(item || '').trim())
      : [];
    if (requestedReferenceImages.length && !acceptsScriptImages(req.body?.modelKey)) {
      return res.status(409).json({
        ok: false,
        status: 'unsupported_reference_images',
        code: 'UNSUPPORTED_REFERENCE_IMAGES',
        error: '当前视频模型不支持参考图，是否允许无参考图生成'
      });
    }
    if (req.body?.modelKey === 'local-doubao-executor-video') {
      const sourceTaskId = `script-video:${Date.now()}:${crypto.randomBytes(8).toString('hex')}`;
      const localPayload = {
        bookId: String(req.body?.bookId || req.body?.book_id || '').trim(),
        videoId: String(req.body?.videoId || req.body?.video_id || '').trim(),
        model: String(req.body?.model || req.body?.videoModel || 'doubao').trim(),
        prompt,
        duration: Math.max(1, Number(req.body?.duration || 10) || 10),
        aspectRatio: String(req.body?.aspectRatio || req.body?.aspect_ratio || '9:16').trim(),
        resolution: String(req.body?.resolution || '720p').trim()
      };
      try {
        const created = await bridgeJSON(
          shuihuoGateway,
          req.auth.account,
          'POST',
          '/api/shuihuo-production/local-executor-jobs',
          { sourceTaskId, platform: 'doubao', payload: localPayload }
        );
        const createdBody = payloadOf(created);
        const taskId = String(createdBody.id || readTaskID(created) || '').trim();
        if (!taskId) return res.status(502).json({ error: '本地执行器服务未返回任务 ID' });
        return res.status(202).json({ ok: true, taskId, status: 'processing' });
      } catch (error) {
        return res.status(error.status || 503).json({ error: error.message || '本地执行器任务提交失败' });
      }
    }
    if (req.body?.modelKey === H3_MODEL_KEY) {
      const h3Prompt = prompt.slice(0, MAX_H3_PROMPT_LENGTH);
      let runtimeModel = null;
      try {
        runtimeModel = resolveRuntimeModel({
          username: req.username,
          kind: 'video',
          modelId: H3_MODEL_KEY,
          memberStore: resolvedMemberStore,
          accountStore,
          account: req.auth?.account,
          configReader
        });
      } catch { /* retain legacy per-account/environment key fallback */ }
      const apiKey = runtimeModel?.credential || h3ApiKeyForRequest(req, configReader, h3ApiKeyReader);
      if (!apiKey) return res.status(400).json({ error: 'MiniMax H3 尚未配置服务端 Token，请联系管理员配置' });
      let referenceImages;
      let duration;
      let resolution;
      try {
        const requestedImages = req.body?.referenceImages ?? req.body?.imageUrls;
        const normalizedReferences = await providerAccessibleImageUrls(req, requestedImages);
        referenceImages = validH3ReferenceImageURLs(normalizedReferences);
        duration = h3Duration(req.body?.duration);
        resolution = h3Resolution(resolveH3VideoMedia({
          aspectRatio: req.body?.aspectRatio || req.body?.aspect_ratio,
          resolution: req.body?.resolution
        }).resolution);
      } catch (error) {
        return res.status(400).json({ error: error.message || 'H3 参数不正确' });
      }
      let h3;
      try {
        h3 = buildH3Request({ workflowId: runtimeModel?.workflowId || h3WorkflowForRequest(req, configReader), prompt: h3Prompt, duration, resolution, referenceImages });
      } catch (error) {
        return res.status(400).json({ error: error.message || 'H3 工作流配置不合法' });
      }
      try {
        const upstream = await h3Submit({ apiKey, workflow: h3.workflow, payload: h3.body, baseUrl: process.env.QIANTIE_AUTODL_H3_BASE_URL || H3_API_BASE_URL });
        if (upstream.statusCode < 200 || upstream.statusCode >= 300) return res.status(502).json({ error: `MiniMax H3 请求失败（HTTP ${upstream.statusCode}）` });
        let body;
        try { body = JSON.parse(upstream.text); } catch { return res.status(502).json({ error: 'MiniMax H3 返回了无法识别的响应' }); }
        const taskId = readH3TaskID(body);
        if (!taskId) return res.status(502).json({ error: 'MiniMax H3 未返回任务 ID' });
        return res.status(202).json({ ok: true, taskId: publicH3TaskID(taskId), provider: 'autodl_comfyui' });
      } catch (error) {
        return res.status(502).json({ error: error?.message === 'H3 视频服务响应超时' ? error.message : 'MiniMax H3 服务暂不可用，请稍后重试' });
      }
    }
    if (req.body?.modelKey === YFAI_SEEDANCE_MODEL) {
      let model;
      try {
        model = resolveYfaiModel(req);
      } catch (error) {
        return res.status(error.status || 400).json({ error: error.message || 'Seedance 尚未配置或不可用' });
      }
      let payload;
      try {
        const imageUrls = await providerAccessibleImageUrls(req, req.body?.imageUrls);
        payload = buildYfaiSeedancePayload({
          prompt,
          duration: req.body?.duration,
          resolution: req.body?.resolution,
          aspectRatio: req.body?.aspectRatio || req.body?.aspect_ratio,
          quality: req.body?.quality,
          imageUrls
        });
      } catch (error) {
        return res.status(400).json({ error: error.message || 'Seedance 参数不正确' });
      }
      try {
        const upstream = await yfaiSubmit({ apiKey: model.credential, payload, baseUrl: model.baseUrl });
        if (upstream.statusCode < 200 || upstream.statusCode >= 300) return res.status(502).json({ error: `Seedance 请求失败（HTTP ${upstream.statusCode}）` });
        const created = parseYfaiTaskResponse(upstream.text);
        return res.status(202).json({ ok: true, taskId: publicYfaiTaskID(created.taskId), provider: 'yfai_seedance' });
      } catch (error) {
        return res.status(502).json({ error: error?.message === 'Seedance 视频服务响应超时' ? error.message : (error.message || 'Seedance 视频服务暂不可用，请稍后重试') });
      }
    }
    let imageUrls;
    try { imageUrls = validOptionalImageURLs(await providerAccessibleImageUrls(req, req.body?.imageUrls)); } catch (error) { return res.status(400).json({ error: error.message || '可选图片参数不正确' }); }
    let apiKey = '';
    try {
      apiKey = getVideoApiKey(configReader(req.username), 'yd');
    } catch (error) {
      const status = Number.isInteger(error?.status) && error.status >= 400 && error.status < 500 ? error.status : 422;
      return res.status(status).json({ error: error?.message || '请先在设置中保存视频生成 API Key' });
    }
    if (!apiKey) return res.status(400).json({ error: '请先在设置中保存视频生成 API Key' });
    try {
      const aspectRatio = ['16:9', '9:16'].includes(String(req.body?.aspectRatio || req.body?.aspect_ratio || ''))
        ? String(req.body?.aspectRatio || req.body?.aspect_ratio) : '9:16';
      const resolution = ['480p', '720p', '1080p'].includes(String(req.body?.resolution || ''))
        ? String(req.body?.resolution) : '720p';
      const upstream = await submit({ apiKey, payload: { model: 'yd2.0-mini', prompt, image_urls: [DEFAULT_FIRST_FRAME_URL, ...imageUrls], duration: '1', aspect_ratio: aspectRatio, resolution } });
      if (upstream.statusCode < 200 || upstream.statusCode >= 300) return res.status(502).json({ error: `视频服务请求失败（HTTP ${upstream.statusCode}）` });
      let body;
      try { body = JSON.parse(upstream.text); } catch { return res.status(502).json({ error: '视频服务返回了无法识别的响应' }); }
      const taskId = readTaskID(body);
      if (!taskId) return res.status(502).json({ error: '视频服务未返回任务 ID' });
      return res.status(202).json({ ok: true, taskId });
    } catch (error) {
      return res.status(502).json({ error: error?.message === '视频服务响应超时' ? error.message : '视频服务暂不可用，请稍后重试' });
    }
  });
  router.get('/:taskId', async (req, res) => {
    const taskId = String(req.params.taskId || '').trim();
    if (!taskId) return res.status(400).json({ error: '视频任务 ID 不能为空' });
    const rawH3TaskID = h3TaskID(taskId);
    if (rawH3TaskID) {
      const apiKey = h3ApiKeyForRequest(req, configReader, h3ApiKeyReader);
      if (!apiKey) return res.status(400).json({ error: 'MiniMax H3 尚未配置服务端 Token，请联系管理员配置' });
      try {
        const statusReply = await h3Request({ apiKey, taskId: rawH3TaskID, baseUrl: process.env.QIANTIE_AUTODL_H3_BASE_URL || H3_API_BASE_URL });
        if (statusReply.statusCode < 200 || statusReply.statusCode >= 300) return res.status(502).json({ error: 'MiniMax H3 任务状态查询失败' });
        const statusBody = JSON.parse(statusReply.text);
        const state = h3TaskState(statusBody);
        if (['QUEUED', 'SUBMITTED', 'RUNNING', 'PENDING', 'PROCESSING', 'GENERATING'].includes(state)) return res.json({ ok: true, taskId, status: 'processing' });
        if (['FAILED', 'ERROR', 'CANCELLED', 'CANCELED'].includes(state)) return res.json({ ok: true, taskId, status: 'failed', error: taskError(statusBody) });
        if (!['SUCCESS', 'SUCCEEDED', 'COMPLETED', 'DONE'].includes(state)) return res.status(502).json({ error: 'MiniMax H3 返回了无法识别的任务状态' });
        const videoURL = h3ResultURL(statusBody);
        if (!videoURL) return res.status(502).json({ error: 'MiniMax H3 已完成，但没有返回可播放地址' });
        return res.json({ ok: true, taskId, status: 'succeeded', videoUrl: videoURL, provider: 'autodl_comfyui' });
      } catch (error) {
        return res.status(502).json({ error: error?.message === 'H3 视频服务响应超时' ? error.message : 'MiniMax H3 任务状态暂不可用，请稍后重试' });
      }
    }
    const rawYfaiTaskID = yfaiTaskID(taskId);
    if (rawYfaiTaskID) {
      let model;
      try {
        model = resolveYfaiModel(req);
      } catch (error) {
        return res.status(error.status || 400).json({ error: error.message || 'Seedance 尚未配置或不可用' });
      }
      try {
        const statusReply = await yfaiRequest({ apiKey: model.credential, taskId: rawYfaiTaskID, baseUrl: model.baseUrl });
        if (statusReply.statusCode < 200 || statusReply.statusCode >= 300) return res.status(502).json({ error: `Seedance 任务状态查询失败（HTTP ${statusReply.statusCode}）` });
        const status = parseYfaiTaskStatus(statusReply.text);
        if (status.status === 'processing') return res.json({ ok: true, taskId, status: 'processing' });
        if (status.status === 'failed') return res.json({ ok: true, taskId, status: 'failed', error: status.error });
        return res.json({ ok: true, taskId, status: 'succeeded', videoUrl: status.videoUrl, provider: 'yfai_seedance' });
      } catch (error) {
        return res.status(502).json({ error: error?.message === 'Seedance 视频服务响应超时' ? error.message : 'Seedance 任务状态暂不可用，请稍后重试' });
      }
    }
    try {
      const local = await bridgeJSON(shuihuoGateway, req.auth.account, 'GET', `/api/shuihuo-production/local-executor-jobs/${encodeURIComponent(taskId)}`);
      const localBody = payloadOf(local);
      const state = String(localBody.state || '').trim().toLowerCase();
      const artifactId = String(localBody.artifactId || localBody.artifact_id || '').trim();
      if (['queued', 'leased', 'preparing', 'submitting', 'accepted', 'generating', 'downloading', 'uploading'].includes(state)) {
        return res.json({ ok: true, taskId, status: 'processing' });
      }
      if (['failed', 'cancelled', 'canceled'].includes(state)) {
        return res.json({ ok: true, taskId, status: 'failed', error: taskError(local) });
      }
      if (['succeeded', 'success', 'completed'].includes(state)) {
        if (!artifactId) return res.status(502).json({ error: '本地执行器任务已完成，但没有返回视频文件' });
        return res.json({
          ok: true,
          taskId,
          status: 'succeeded',
          videoUrl: `/api/shuihuo-production/local-executor-artifacts/${encodeURIComponent(artifactId)}`
        });
      }
      return res.json({ ok: true, taskId, status: 'processing' });
    } catch (error) {
      if (error.status !== 404) return res.status(error.status || 503).json({ error: error.message || '本地执行器任务状态查询失败' });
    }
    const apiKey = getVideoApiKey(configReader(req.username), 'yd');
    if (!apiKey) return res.status(400).json({ error: '请先在设置中保存视频生成 API Key' });
    try {
      const statusReply = await request(`${YD_TASKS_URL}/${encodeURIComponent(taskId)}`, apiKey);
      if (statusReply.statusCode < 200 || statusReply.statusCode >= 300) return res.status(502).json({ error: '视频任务状态查询失败' });
      const statusBody = JSON.parse(statusReply.text);
      const state = taskState(statusBody);
      if (['QUEUED', 'SUBMITTED', 'RUNNING', 'PENDING'].includes(state)) return res.json({ ok: true, taskId, status: 'processing' });
      if (state === 'FAILED') return res.json({ ok: true, taskId, status: 'failed', error: taskError(statusBody) });
      if (state !== 'SUCCESS') return res.status(502).json({ error: '视频服务返回了无法识别的任务状态' });
      const resultReply = await request(`${YD_TASKS_URL}/${encodeURIComponent(taskId)}/result`, apiKey);
      if (resultReply.statusCode < 200 || resultReply.statusCode >= 300) return res.status(502).json({ error: '视频结果获取失败' });
      const videoUrl = resultURL(JSON.parse(resultReply.text));
      if (!videoUrl) return res.status(502).json({ error: '视频服务未返回可播放地址' });
      return res.json({ ok: true, taskId, status: 'succeeded', videoUrl });
    } catch (error) {
      return res.status(502).json({ error: error?.message === '视频服务响应超时' ? error.message : '视频任务状态暂不可用，请稍后重试' });
    }
  });
  router.get('/:taskId/download', (req, res) => bridgeDownload(shuihuoGateway, req.auth.account, `/api/script-videos/${encodeURIComponent(String(req.params.taskId || ''))}/download`, res));
  return router;
}

module.exports = {
  createScriptVideoRouter,
  DEFAULT_FIRST_FRAME_URL,
  h3ApiKeyFromEnvironment,
  h3Duration,
  h3Resolution,
  resolveH3VideoMedia,
  h3WorkflowForRequest,
  publicH3TaskID,
  publicYfaiTaskID,
  yfaiTaskID,
  validH3ReferenceImageURLs,
  validOptionalImageURLs,
  readTaskID,
  taskState,
  resultURL
};
