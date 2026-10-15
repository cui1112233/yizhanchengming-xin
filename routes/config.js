const express = require('express');
const { apiAuth } = require('../middleware/auth');
const { readConfig, writeConfig, publicConfig, normalizeImageConfig, normalizeVideoConfig, normalizeProductionRetentionDays, DEFAULT_CONFIG, requestUpstream, collectResponse } = require('../lib/shared');
const { syncAccountAIConfig } = require('./shuihuo-production');
const { normalizeStorageRoot } = require('../lib/storage-root');
const { normalizePetConfig } = require('../lib/pet-catalog');
const {
  listVisibleModels,
  listManagerModels,
  saveManagerModel,
  updateManagerModel,
  removeManagerModel
} = require('../lib/model-catalog-runtime');
const { createModelReferenceResolver } = require('../lib/model-reference-resolver');
const { createSignedBridgeHeaders } = require('../lib/batch-factory-v11/go-proxy');
const { normalizeModelCatalog } = require('../lib/model-catalog');
const { createTextVerificationCache } = require('../lib/model-catalog-verification');
const { readModelQuotas } = require('../lib/model-quota');

const LOCAL_DOUBAO_MODEL_ID = 'local-doubao-executor-video';

async function defaultExecutorPairingStatus({ username, account, shuihuoGateway, fetchImpl = globalThis.fetch }) {
  if (!fetchImpl) return false;
  const base = String(shuihuoGateway?.targetBaseUrl || process.env.QIANTIE_GO_BASE_URL || 'http://127.0.0.1:4000').replace(/\/$/, '');
  const secret = shuihuoGateway?.bridgeSecret || process.env.QIANTIE_BRIDGE_SECRET || '';
  const pathname = '/api/shuihuo-production/local-executors';
  try {
    const response = await fetchImpl(`${base}${pathname}`, {
      headers: createSignedBridgeHeaders({
        username,
        isOwner: account?.isOwner === true,
        method: 'GET',
        pathname,
        secret
      })
    });
    if (!response.ok) return false;
    const payload = await response.json();
    return Array.isArray(payload?.executors)
      && payload.executors.some(executor => executor?.platform === 'doubao');
  } catch {
    return false;
  }
}

function normalizeTtsConfig(value, fallback = {}) {
  const voice = typeof value?.voice === 'string' && value.voice.startsWith('zh-CN-')
    ? value.voice
    : fallback.voice || 'zh-CN-XiaoxiaoNeural';
  const style = ['general', 'cheerful', 'sad', 'friendly', 'chat'].includes(value?.style)
    ? value.style
    : fallback.style || 'general';
  const speed = Number(value?.speed);
  const pitch = Number(value?.pitch);
  return {
    voice,
    style,
    speed: Number.isFinite(speed) ? Math.min(2, Math.max(0.5, speed)) : (fallback.speed ?? 1.8),
    pitch: Number.isFinite(pitch) ? Math.min(50, Math.max(-50, pitch)) : (fallback.pitch ?? 10)
  };
}

function normalizeNotifications(value, fallback = {}) {
  const volume = Number(value?.soundVolume);
  return {
    soundEnabled: typeof value?.soundEnabled === 'boolean' ? value.soundEnabled : (fallback.soundEnabled ?? true),
    petVisible: typeof value?.petVisible === 'boolean' ? value.petVisible : (fallback.petVisible ?? true),
    soundVolume: Number.isFinite(volume) ? Math.min(100, Math.max(0, Math.round(volume))) : 60
  };
}

function normalizeAvatar(value, fallback = null) {
  const emoji = typeof value?.emoji === 'string' ? value.emoji.trim() : '';
  const background = typeof value?.background === 'string' && /^#[0-9a-fA-F]{3,8}$/.test(value.background.trim())
    ? value.background.trim()
    : '';
  return emoji && background ? { emoji, background } : fallback;
}

function normalizeScriptDefaults(value, fallback = {}) {
  const imageAspectRatio = ['16:9', '9:16', '1:1'].includes(value?.imageAspectRatio)
    ? value.imageAspectRatio : (['16:9', '9:16', '1:1'].includes(fallback?.imageAspectRatio) ? fallback.imageAspectRatio : '9:16');
  const videoAspectRatio = ['16:9', '9:16'].includes(value?.videoAspectRatio)
    ? value.videoAspectRatio : (['16:9', '9:16'].includes(fallback?.videoAspectRatio) ? fallback.videoAspectRatio : '9:16');
  const videoResolution = ['480p', '720p', '1080p'].includes(value?.videoResolution)
    ? value.videoResolution : (['480p', '720p', '1080p'].includes(fallback?.videoResolution) ? fallback.videoResolution : '720p');
  return {
    textModelId: typeof value?.textModelId === 'string' ? value.textModelId.trim() : String(fallback?.textModelId || ''),
    imageModelId: typeof value?.imageModelId === 'string' ? value.imageModelId.trim() : String(fallback?.imageModelId || ''),
    videoModelKey: typeof value?.videoModelKey === 'string' ? value.videoModelKey.trim() : String(fallback?.videoModelKey || ''),
    imageAspectRatio,
    videoAspectRatio,
    videoResolution
  };
}

function apiManagementState(req, memberStore, accountStore) {
  const member = memberStore?.getMember?.(req.username);
  const owner = req.auth?.account?.isOwner === true
    || accountStore?.getInternalAccount?.(req.username)?.isOwner === true;
  return {
    member,
    canManageApi: owner || Boolean(member?.active && ['dev', 'manager'].includes(member.role))
  };
}

function managedPublicConfig(config, member) {
  const safe = publicConfig(config);
  return {
    ...safe,
    provider: 'managed',
    baseUrl: '',
    model: '',
    hasApiKey: false,
    image: { ...safe.image, baseUrl: '', model: '', displayName: '', hasApiKey: false },
    video: { ...safe.video, baseUrl: '', model: '', displayName: '', hasApiKey: false },
    canManageApi: false,
    managedBy: member?.boundTo || null
  };
}

function createConfigRouter({
  shuihuoGateway,
  memberStore,
  accountStore,
  configReader = readConfig,
  configWriter = writeConfig,
  authenticate = apiAuth,
  isModelReferenced,
  batchFactoryStoreFactory,
  accountReader,
  getExecutorPairingStatus = defaultExecutorPairingStatus,
  textVerification = createTextVerificationCache(),
  upstreamRequest = requestUpstream,
  quotaReader = readModelQuotas
} = {}) {
  const router = express.Router();
  router.use(authenticate);
  const resolveModelReference = isModelReferenced || createModelReferenceResolver({
    configReader,
    batchFactoryStoreFactory,
    accountReader
  });

  function requireApiManager(req, res, next) {
    if (!apiManagementState(req, memberStore, accountStore).canManageApi) return res.status(403).json({ error: '仅管理者可以管理模型' });
    next();
  }

  function sendModelError(res, error) {
    return res.status(error?.status || 400).json({ error: error?.message || '模型配置处理失败' });
  }

  function resolvedTextModelForSave(req, input, existingModelId = '') {
    const config = configReader(req.username) || {};
    const catalog = normalizeModelCatalog(config.modelCatalog, config);
    const existing = existingModelId ? catalog.find(model => model.id === existingModelId) : null;
    const patch = { ...(input || {}) };
    if (!String(patch.credential || '').trim()) delete patch.credential;
    return { ...(existing || {}), ...patch };
  }

  function requireTextVerification(req, input, existingModelId = '') {
    const model = resolvedTextModelForSave(req, input, existingModelId);
    const changesRuntimeConfig = ['baseUrl', 'modelId', 'credential'].some(key => Object.hasOwn(input || {}, key) && String(input[key] || '').trim());
    const enablesModel = Object.hasOwn(input || {}, 'enabled') && input.enabled === true;
    if (model.kind === 'text' && model.enabled === true && (!existingModelId || changesRuntimeConfig || enablesModel)) {
      textVerification.assertApproved(req.username, model);
    }
  }

  router.post('/models/test', requireApiManager, async (req, res) => {
    try {
      const model = resolvedTextModelForSave(req, req.body, req.body?.existingModelId);
      if (model.kind !== 'text' || !model.baseUrl || !model.modelId || !model.credential) {
        throw Object.assign(new Error('测试文本模型前请填写 Base URL、模型 ID 和 API Key'), { status: 422 });
      }
      const upstream = await upstreamRequest({ baseUrl: model.baseUrl, model: model.modelId, apiKey: model.credential }, {
        model: model.modelId,
        messages: [{ role: 'user', content: 'Hi' }],
        max_tokens: 5
      }, collectResponse);
      if (upstream.statusCode < 200 || upstream.statusCode >= 300) {
        throw Object.assign(new Error(`文本模型连接失败（状态 ${upstream.statusCode}），请检查接口、模型名或密钥`), { status: 502 });
      }
      const response = JSON.parse(upstream.text);
      if (!response?.choices?.[0]?.message) throw Object.assign(new Error('文本模型未返回有效答复，请检查模型兼容性'), { status: 502 });
      textVerification.approve(req.username, model);
      return res.json({ ok: true, message: '文本模型连接成功，现在可以保存并启用。' });
    } catch (error) {
      return sendModelError(res, error);
    }
  });

  async function readAndPersistExecutorPairing(req) {
    const paired = await getExecutorPairingStatus({
      username: req.username,
      account: req.auth?.account,
      shuihuoGateway
    });
    const config = configReader(req.username) || {};
    const rawCatalog = Array.isArray(config.modelCatalog) ? config.modelCatalog : [];
    const index = rawCatalog.findIndex(model => model?.id === LOCAL_DOUBAO_MODEL_ID);
    if (index < 0) return paired;
    if (rawCatalog[index]?.executorPaired === paired) return paired;
    const nextCatalog = rawCatalog.map((model, position) => position === index ? { ...model, executorPaired: paired } : model);
    configWriter(req.username, { ...config, modelCatalog: nextCatalog, modelCatalogVersion: 1 });
    return paired;
  }

  router.get('/models', (req, res) => {
    if (apiManagementState(req, memberStore, accountStore).canManageApi) {
      return res.json({ models: listManagerModels({
        username: req.username,
        kind: req.query.kind,
        configReader
      }) });
    }
    res.json({ models: listVisibleModels({
      username: req.username,
      kind: req.query.kind,
      memberStore,
      accountStore,
      account: req.auth?.account,
      configReader
    }) });
  });

  router.get('/models/quotas', requireApiManager, async (req, res) => {
    try {
      const config = configReader(req.username) || {};
      const catalog = normalizeModelCatalog(config.modelCatalog, config);
      const quotas = await quotaReader(catalog);
      return res.json({ quotas: Array.isArray(quotas) ? quotas : [] });
    } catch (error) {
      return res.status(502).json({ error: error?.message || '额度查询失败' });
    }
  });

  router.get('/models/local-doubao-executor-video/pairing-status', requireApiManager, async (req, res) => {
    const executorPaired = await readAndPersistExecutorPairing(req);
    return res.json({ executorPaired });
  });

  router.post('/models', requireApiManager, async (req, res) => {
    try {
      const body = { ...(req.body || {}) };
      if (body.id === LOCAL_DOUBAO_MODEL_ID) {
        body.executorPaired = await readAndPersistExecutorPairing(req);
        if (body.enabled && !body.executorPaired) throw Object.assign(new Error('请先完成本地豆包执行器配对'), { status: 422 });
      }
      requireTextVerification(req, body);
      const model = saveManagerModel(req.username, body, { configReader, configWriter });
      return res.status(201).json({ model });
    } catch (error) {
      return sendModelError(res, error);
    }
  });

  router.patch('/models/:modelId', requireApiManager, async (req, res) => {
    try {
      const body = { ...(req.body || {}) };
      if (req.params.modelId === LOCAL_DOUBAO_MODEL_ID) {
        body.executorPaired = await readAndPersistExecutorPairing(req);
        if (body.enabled && !body.executorPaired) throw Object.assign(new Error('请先完成本地豆包执行器配对'), { status: 422 });
      }
      requireTextVerification(req, body, req.params.modelId);
      const model = updateManagerModel(req.username, req.params.modelId, body, { configReader, configWriter });
      return res.json({ model });
    } catch (error) {
      return sendModelError(res, error);
    }
  });

  router.delete('/models/:modelId', requireApiManager, async (req, res) => {
    try {
      await removeManagerModel(req.username, req.params.modelId, { configReader, configWriter, isModelReferenced: resolveModelReference });
      return res.status(204).end();
    } catch (error) {
      return sendModelError(res, error);
    }
  });

  // GET /api/config — 获取配置（不含 apiKey）
  router.get('/', (req, res) => {
    const config = configReader(req.username);
    config.pet = normalizePetConfig(config.pet);
    config.tts = normalizeTtsConfig(config.tts);
    config.notifications = normalizeNotifications(config.notifications);
    config.avatar = normalizeAvatar(config.avatar);
    config.scriptDefaults = normalizeScriptDefaults(config.scriptDefaults);
    config.productionRetentionDays = normalizeProductionRetentionDays(config.productionRetentionDays);
    const { member, canManageApi } = apiManagementState(req, memberStore, accountStore);
    if (!canManageApi) return res.json(managedPublicConfig(config, member));
    return res.json({ ...publicConfig(config), canManageApi: true, managedBy: null });
  });

  // POST /api/config — 保存配置
  router.post('/', async (req, res) => {
    const body = req.body;
    const requestedRetentionDays = Object.prototype.hasOwnProperty.call(body || {}, 'productionRetentionDays')
      ? normalizeProductionRetentionDays(body.productionRetentionDays)
      : null;
    if (typeof body?.storageRoot === 'string') {
      const storageRoot = normalizeStorageRoot(body.storageRoot);
      if (storageRoot.error) return res.status(400).json({ error: storageRoot.error });
      body.storageRoot = storageRoot.value;
    }
    const oldConfig = configReader(req.username);
    const { member, canManageApi } = apiManagementState(req, memberStore, accountStore);
    if (!canManageApi) {
      const nextConfig = {
        ...oldConfig,
        storageRoot: typeof body.storageRoot === 'string' ? body.storageRoot : (oldConfig.storageRoot || ''),
        productionRetentionDays: requestedRetentionDays ?? normalizeProductionRetentionDays(oldConfig.productionRetentionDays),
        pet: normalizePetConfig(body.pet, oldConfig.pet),
        tts: normalizeTtsConfig(body.tts, oldConfig.tts),
        notifications: normalizeNotifications(body.notifications, oldConfig.notifications),
        avatar: normalizeAvatar(body.avatar, oldConfig.avatar),
        scriptDefaults: normalizeScriptDefaults(body.scriptDefaults, oldConfig.scriptDefaults)
      };
      configWriter(req.username, nextConfig);
      return res.json(managedPublicConfig(nextConfig, member));
    }
    const nextConfig = {
      // Generic preference saves must be field-preserving. Model credentials
      // are managed by /models and may be referenced by automation presets;
      // rebuilding this object from a short allow-list used to erase the
      // complete modelCatalog whenever a manager saved an unrelated setting.
      ...oldConfig,
      provider: body.provider || oldConfig.provider || DEFAULT_CONFIG.provider,
      baseUrl: body.baseUrl || oldConfig.baseUrl || DEFAULT_CONFIG.baseUrl,
      model: body.model || oldConfig.model || DEFAULT_CONFIG.model,
      apiKey: body.apiKey ? body.apiKey : oldConfig.apiKey,
      storageRoot: typeof body.storageRoot === 'string' ? body.storageRoot : (oldConfig.storageRoot || ''),
      productionRetentionDays: requestedRetentionDays ?? normalizeProductionRetentionDays(oldConfig.productionRetentionDays),
      image: normalizeImageConfig(body.image, oldConfig.image),
      video: normalizeVideoConfig(body.video, oldConfig.video),
      pet: normalizePetConfig(body.pet, oldConfig.pet),
      tts: normalizeTtsConfig(body.tts, oldConfig.tts),
      notifications: normalizeNotifications(body.notifications, oldConfig.notifications),
      avatar: normalizeAvatar(body.avatar, oldConfig.avatar),
      scriptDefaults: normalizeScriptDefaults(body.scriptDefaults, oldConfig.scriptDefaults)
    };
    if (shuihuoGateway) {
      try {
        await syncAccountAIConfig({
          ...shuihuoGateway,
          username: req.auth.account.username,
          isOwner: req.auth.account.isOwner === true,
          config: nextConfig
        });
      } catch (error) {
        return res.status(error.status || 503).json({ error: error.message || '水货生产的 AI 配置同步失败' });
      }
    }
    configWriter(req.username, nextConfig);
    res.json({ ...publicConfig(nextConfig), canManageApi: true, managedBy: null });
  });

  return router;
}

module.exports = { createConfigRouter, normalizePetConfig, normalizeTtsConfig, normalizeNotifications, normalizeAvatar, normalizeScriptDefaults, managedPublicConfig, defaultExecutorPairingStatus };
