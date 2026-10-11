const crypto = require('node:crypto');
const http = require('node:http');
const https = require('node:https');
const express = require('express');
const { apiAuth } = require('../middleware/auth');
const { getVideoApiKey, readConfig } = require('../lib/shared');
const { listPublishedForSlot, resolveSystemPresetBody, slotDefinition } = require('../lib/system-preset-catalog');
const { getDefaultVideoModels } = require('../lib/video-model-catalog');

const HOP_BY_HOP_HEADERS = new Set([
  'connection', 'keep-alive', 'proxy-authenticate', 'proxy-authorization',
  'te', 'trailer', 'transfer-encoding', 'upgrade'
]);

function signBridgeRequest(secret, { username, isOwner, issuedAt, method, pathname }) {
  // Shuihuo remains on the established Go gateway, whose platform identity
  // protocol uses newline-delimited fields. V11 has its own upstream and a
  // distinct signature representation, so it must not be reused here.
  const payload = [username, issuedAt, String(isOwner), method, pathname].join('\n');
  return crypto.createHmac('sha256', secret).update(payload).digest('hex');
}

// Shuihuo's production project API is served by the retained compatibility
// service. Go is only the legacy fallback for environments that do not run it.
// Keeping this resolver shared prevents the project list and the workbench
// bridge from silently targeting different upstreams.
function resolveShuihuoBaseUrl(targetBaseUrl) {
  return process.env.QIANTIE_SHUIHUO_COMPAT_BASE_URL
    || process.env.QIANTIE_GO_BASE_URL
    || targetBaseUrl
    || 'http://127.0.0.1:4000';
}

function isTextInferenceRequest(method, pathname) {
  if (method !== 'POST') return false;
  return /^\/api\/shuihuo-production\/projects\/\d+\/(segmentation\/smart|analysis\/(assets|assets-and-bindings))$/.test(pathname)
    || /^\/api\/shuihuo-production\/projects\/\d+\/prompt-candidates\/(image|video)$/.test(pathname);
}

function requiresAccountAIConfigSync(method, pathname) {
  return isTextInferenceRequest(method, pathname)
    || (method === 'POST' && (
      /^\/api\/shuihuo-production\/projects\/\d+\/assets\/generate$/.test(pathname)
      || /^\/api\/shuihuo-production\/projects\/\d+\/tasks(?:\/batch)?$/.test(pathname)
    ));
}

function accountAIConfigPayload(config) {
  const provider = String(config?.provider || '').trim();
  const baseUrl = String(config?.baseUrl || '').trim();
  const model = String(config?.model || '').trim();
  const apiKey = String(config?.apiKey || '').trim();
  const imageProvider = String(config?.image?.provider || '').trim();
  const imageMode = String(config?.image?.mode || '').trim();
  const imageDisplayName = String(config?.image?.displayName || '').trim();
  const imageBaseURL = String(config?.image?.baseUrl || '').trim();
  const imageModel = String(config?.image?.model || '').trim();
  const imageAPIKey = String(config?.image?.apiKey || '').trim();
  const videoAPIKey = getVideoApiKey(config, 'yd');
  const text = baseUrl && model && apiKey
    ? { provider: provider || 'custom', baseUrl, model, apiKey }
    : null;
  const image = imageProvider === 'openai_compatible' && imageBaseURL && imageModel && imageAPIKey
    ? { mode: imageMode === 'custom' ? 'custom' : 'openai_compatible', provider: imageProvider, displayName: imageDisplayName, baseUrl: imageBaseURL, model: imageModel, apiKey: imageAPIKey }
    : null;
  const video = videoAPIKey
    ? { provider: 'yd_video', apiKey: videoAPIKey }
    : null;
  if (!text && !image && !video) {
    return null;
  }
  return { ...(text || {}), ...(image ? { image } : {}), ...(video ? { video } : {}) };
}

function syncAccountAIConfig({ targetBaseUrl, bridgeSecret, username, isOwner, config }) {
  const payload = accountAIConfigPayload(config);
  if (!payload) {
    return Promise.resolve();
  }
  const target = new URL(resolveShuihuoBaseUrl(targetBaseUrl));
  const secret = bridgeSecret || process.env.QIANTIE_BRIDGE_SECRET || '';
  const transport = target.protocol === 'https:' ? https : http;
  const pathname = '/api/shuihuo-production/account-ai-config';
  const issuedAt = String(Math.floor(Date.now() / 1000));
  const body = Buffer.from(JSON.stringify(payload));
  const signature = signBridgeRequest(secret, { username, isOwner, issuedAt, method: 'PUT', pathname });

  return new Promise((resolve, reject) => {
    const upstream = transport.request({
      protocol: target.protocol,
      hostname: target.hostname,
      port: target.port || undefined,
      method: 'PUT',
      path: pathname,
      headers: {
        'X-Qiantie-Username': username,
        'X-Qiantie-Is-Owner': String(isOwner),
        'X-Qiantie-Issued-At': issuedAt,
        'X-Qiantie-Signature': signature,
        Accept: 'application/json',
        'Content-Type': 'application/json',
        'Content-Length': String(body.length)
      },
      timeout: 15_000
    }, response => {
      response.resume();
      response.on('end', () => {
        if ((response.statusCode || 500) >= 200 && (response.statusCode || 500) < 300) return resolve();
        const error = new Error('水货生产的 AI 配置同步失败，请检查服务是否已启动');
        error.status = response.statusCode === 401 || response.statusCode === 403 ? 503 : response.statusCode || 503;
        reject(error);
      });
    });
    upstream.on('timeout', () => upstream.destroy(new Error('水货生产的 AI 配置同步超时')));
    upstream.on('error', () => {
      const error = new Error('水货生产的 AI 配置同步失败，请检查服务是否已启动');
      error.status = 503;
      reject(error);
    });
    upstream.end(body);
  });
}

function createPublishedPromptBody(body, presetStore, presetIds, label) {
  const presets = presetIds.map(id => presetStore?.getPublished?.(id));
  if (presets.some(preset => !preset)) {
    const error = new Error(`管理后台未发布“${label}”预设词`);
    error.status = 409;
    throw error;
  }
  return {
    ...(body || {}),
    systemPromptId: presetIds.join(','),
    systemPromptVersion: Math.max(...presets.map(preset => Number(preset.version) || 0)),
    systemPrompt: presetIds.map(id => resolveSystemPresetBody(presetStore, id)).join('\n\n---\n\n')
  };
}

function createSmartSegmentationBody(body, presetStore) {
  return createPublishedPromptBody(body, presetStore, ['shuihuo-smart-segmentation'], '智能识别（分镜提示词）');
}

const ASSET_ANALYSIS_SCOPES = Object.freeze(['all', 'character', 'scene', 'prop']);
const ASSET_ANALYSIS_PRESET = Object.freeze({ presetId: 'shuihuo-extract-assets', selectionKey: 'assetPresetId', slot: 'shuihuo.asset.extraction', label: '人物场景、道具提取' });
const CHARACTER_SHEET_PRESET = Object.freeze({ selectionKey: 'characterSheetPresetId', slot: 'shuihuo.asset.character-sheet', label: '人物设定' });
const PROMPT_PRESETS = Object.freeze({
  image: { presetId: 'shuihuo-image-prompt', selectionKey: 'promptPresetId', slot: 'shuihuo.prompt.image', label: '画面提示词' },
  video: { presetId: 'shuihuo-video-prompt', selectionKey: 'promptPresetId', slot: 'shuihuo.prompt.video', label: '视频提示词' }
});

function requirePublishedSlot(presetStore, id, definition) {
  const preset = presetStore?.getPublished?.(id);
  const slot = slotDefinition(definition.slot);
  if (!slot || !preset || preset.module !== slot.module || preset.kind !== 'base' || preset.protocolLock?.slot !== slot.id) {
    const error = new Error(`管理后台未发布“${definition.label}”预设词`);
    error.status = 409;
    throw error;
  }
  return preset;
}

function createScopedAssetAnalysisBody(body, presetStore) {
  const scope = body?.scope || 'all';
  if (!ASSET_ANALYSIS_SCOPES.includes(scope)) {
    const error = new Error('分析范围必须是全部预设、角色、场景或道具');
    error.status = 400;
    throw error;
  }
  const selectedId = body?.[ASSET_ANALYSIS_PRESET.selectionKey] || ASSET_ANALYSIS_PRESET.presetId;
  const preset = requirePublishedSlot(presetStore, selectedId, ASSET_ANALYSIS_PRESET);
  return {
    ...createPublishedPromptBody(body, presetStore, [preset.id], '人物场景、道具提取'),
    scope,
    [ASSET_ANALYSIS_PRESET.selectionKey]: preset.id,
    [`${ASSET_ANALYSIS_PRESET.selectionKey}Version`]: preset.version
  };
}

function createAssetPlanBody(body, presetStore) {
  const selectedId = body?.[ASSET_ANALYSIS_PRESET.selectionKey] || ASSET_ANALYSIS_PRESET.presetId;
  const assetPreset = requirePublishedSlot(presetStore, selectedId, ASSET_ANALYSIS_PRESET);
  return {
    ...createPublishedPromptBody(body, presetStore, [assetPreset.id, 'shuihuo-asset-binding'], '人物场景、道具提取和分镜资产绑定'),
    [ASSET_ANALYSIS_PRESET.selectionKey]: assetPreset.id,
    [`${ASSET_ANALYSIS_PRESET.selectionKey}Version`]: assetPreset.version
  };
}

function createPromptCandidateBody(kind, body, presetStore) {
  const definition = PROMPT_PRESETS[kind];
  const selectedId = body?.[definition.selectionKey] || definition.presetId;
  const preset = requirePublishedSlot(presetStore, selectedId, definition);
  return {
    ...createPublishedPromptBody(body, presetStore, [preset.id], definition.label),
    [definition.selectionKey]: preset.id,
    [`${definition.selectionKey}Version`]: preset.version
  };
}

function createAssetImageGenerationBody(body, presetStore) {
  const selectedId = body?.[CHARACTER_SHEET_PRESET.selectionKey];
  if (!selectedId) return { ...(body || {}), systemPrompt: '' };
  const preset = requirePublishedSlot(presetStore, selectedId, CHARACTER_SHEET_PRESET);
  return {
    ...(body || {}),
    systemPromptId: preset.id,
    systemPromptVersion: preset.version,
    systemPrompt: resolveSystemPresetBody(presetStore, preset.id)
  };
}

function systemPromptBodyForRequest(pathname, body, presetStore) {
  if (/^\/api\/shuihuo-production\/projects\/\d+\/segmentation\/smart$/.test(pathname)) {
    return createSmartSegmentationBody(body, presetStore);
  }
  if (/^\/api\/shuihuo-production\/projects\/\d+\/analysis\/assets$/.test(pathname)) {
    return createScopedAssetAnalysisBody(body, presetStore);
  }
  if (/^\/api\/shuihuo-production\/projects\/\d+\/analysis\/assets-and-bindings$/.test(pathname)) {
    return createAssetPlanBody(body, presetStore);
  }
  if (/^\/api\/shuihuo-production\/projects\/\d+\/analysis\/bindings$/.test(pathname)) {
    return createPublishedPromptBody(body, presetStore, ['shuihuo-asset-binding'], '分镜资产绑定');
  }
  if (/^\/api\/shuihuo-production\/projects\/\d+\/assets\/generate$/.test(pathname)) {
    return createAssetImageGenerationBody(body, presetStore);
  }
  const promptMatch = pathname.match(/^\/api\/shuihuo-production\/projects\/\d+\/prompt-candidates\/(image|video)$/);
  if (promptMatch) {
    return createPromptCandidateBody(promptMatch[1], body, presetStore);
  }
  return body ?? {};
}

function upstreamTimeoutForRequest(method, pathname) {
  const waitsForTextModel = method === 'POST' && (
    /^\/api\/shuihuo-production\/projects\/\d+\/segmentation\/smart$/.test(pathname) ||
    /^\/api\/shuihuo-production\/projects\/\d+\/analysis\/assets(?:-and-bindings)?$/.test(pathname) ||
    /^\/api\/shuihuo-production\/projects\/\d+\/prompt-candidates\/(image|video)$/.test(pathname)
  );
  return waitsForTextModel ? 100_000 : 15_000;
}

function createShuihuoProductionRouter({ targetBaseUrl, bridgeSecret, presetStore, configReader = readConfig, authenticate = apiAuth } = {}) {
  const target = new URL(resolveShuihuoBaseUrl(targetBaseUrl));
  const secret = bridgeSecret || process.env.QIANTIE_BRIDGE_SECRET || '';
  const transport = target.protocol === 'https:' ? https : http;
  const router = express.Router();

  router.use(authenticate);
  router.get('/models', (req, res) => {
    const accountConfig = configReader(req.auth.account.username);
    const models = getDefaultVideoModels({
      h3Configured: Boolean(getVideoApiKey(accountConfig, 'h3') || process.env.QIANTIE_AUTODL_H3_API_KEY || process.env.QIANTIE_H3_API_KEY)
    }).map(model => {
      if (model.key === 'yd2-mini-video') {
        return { ...model, configured: Boolean(getVideoApiKey(accountConfig, 'yd')) };
      }
      return model;
    });
    return res.json({ models });
  });
  router.get('/preset-slots', (req, res) => {
    const slots = [
      ASSET_ANALYSIS_PRESET.slot,
      PROMPT_PRESETS.image.slot,
      PROMPT_PRESETS.video.slot,
      CHARACTER_SHEET_PRESET.slot
    ];
    return res.json({ presets: slots.flatMap(slot => listPublishedForSlot(presetStore, slot)) });
  });
  router.use(async (req, res) => {
    const requestURL = new URL(req.originalUrl, 'http://qiantie-gateway.local');
    const issuedAt = String(Math.floor(Date.now() / 1000));
    const username = req.auth.account.username;
    const isOwner = req.auth.account.isOwner === true;
    if (requiresAccountAIConfigSync(req.method, requestURL.pathname)) {
      try {
        await syncAccountAIConfig({
          targetBaseUrl: target.toString(), bridgeSecret: secret, username, isOwner,
          config: readConfig(username)
        });
      } catch (error) {
        return res.status(error.status || 503).json({ error: error.message || '水货生产的 AI 配置同步失败' });
      }
    }
    const signature = signBridgeRequest(secret, { username, isOwner, issuedAt, method: req.method, pathname: requestURL.pathname });
    let forwardedBody = req.body ?? {};
    if (req.method === 'POST') {
      try {
        forwardedBody = systemPromptBodyForRequest(requestURL.pathname, req.body, presetStore);
      } catch (error) {
        return res.status(error.status || 500).json({ error: error.message || '读取系统预设词失败' });
      }
    }
    const body = ['GET', 'HEAD'].includes(req.method) ? null : Buffer.from(JSON.stringify(forwardedBody));
    const headers = {
      'X-Qiantie-Username': username,
      'X-Qiantie-Is-Owner': String(isOwner),
      'X-Qiantie-Issued-At': issuedAt,
      'X-Qiantie-Signature': signature,
      Accept: req.get('accept') || 'application/json'
    };
    if (body) {
      headers['Content-Type'] = 'application/json';
      headers['Content-Length'] = String(body.length);
    }

    const upstream = transport.request({
      protocol: target.protocol,
      hostname: target.hostname,
      port: target.port || undefined,
      method: req.method,
      path: requestURL.pathname + requestURL.search,
      headers,
      timeout: upstreamTimeoutForRequest(req.method, requestURL.pathname)
    }, upstreamResponse => {
      if (upstreamResponse.statusCode === 401 || upstreamResponse.statusCode === 403) {
        upstreamResponse.resume();
        upstreamResponse.on('end', () => {
          if (!res.headersSent) res.status(503).json({ error: '水货生产服务鉴权失败，请联系管理员检查服务连接' });
        });
        return;
      }
      for (const [name, value] of Object.entries(upstreamResponse.headers)) {
        if (value !== undefined && !HOP_BY_HOP_HEADERS.has(name.toLowerCase())) res.setHeader(name, value);
      }
      res.status(upstreamResponse.statusCode || 502);
      upstreamResponse.pipe(res);
    });

    upstream.on('timeout', () => upstream.destroy(new Error('水货生产服务响应超时')));
    upstream.on('error', error => {
      if (res.headersSent) return res.destroy(error);
      res.status(503).json({ error: '水货生产服务暂不可用，请稍后重试' });
    });
    if (body) upstream.write(body);
    upstream.end();
  });

  return router;
}

module.exports = {
  createShuihuoProductionRouter,
  createPublishedPromptBody,
  createSmartSegmentationBody,
  createScopedAssetAnalysisBody,
  createAssetPlanBody,
  createPromptCandidateBody,
  createAssetImageGenerationBody,
  requirePublishedSlot,
  signBridgeRequest,
  resolveShuihuoBaseUrl,
  syncAccountAIConfig,
  isTextInferenceRequest,
  systemPromptBodyForRequest,
  upstreamTimeoutForRequest
};
