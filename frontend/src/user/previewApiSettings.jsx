// 本地视觉预览入口：只用于查看「API 接口」分区的骨架，不参与生产构建。
// 直接渲染设置页，并拦截 /api 请求返回只读假数据，避免依赖登录与运行中的 4001。
import React from 'react';
import { createRoot } from 'react-dom/client';
import { ConfigProvider } from 'antd';
import 'antd/dist/reset.css';
import '../shared/styles/global.css';
import '../shared/styles/v88-tokens.css';
import '../shared/styles/v88-ui.css';
import '../shared/styles/account-center-visual-rebuild.css';
import { createAntTheme } from '../shared/styles/theme';
import { SettingsPage } from './pages/SettingsPage';
import ApiConfigPage from './pages/ApiConfigPage';
import { isCatalogModelAvailable } from './pages/apiConfigCatalogMerge';

const params = new URLSearchParams(window.location.search);
const theme = params.get('theme') === 'light' ? 'light' : 'dark';
// ?view=account 渲染个人中心「API 配置」页（渠道密钥 + 模型目录），默认仍是设置页。
const view = params.get('view');
const width = Number(params.get('width'));

document.documentElement.dataset.theme = theme;
if (Number.isFinite(width) && width > 0) {
  document.documentElement.style.setProperty('--preview-width', `${width}px`);
}

const jsonResponse = (body, status = 200) => new Response(JSON.stringify(body), {
  status,
  headers: { 'content-type': 'application/json' }
});

// 预览数据固定：未记录勾选时预置渠道默认全部展示；是否可用只看 Key 是否存在。
const channelSelection = {};

const integrationState = {
  initialized: false,
  pinned: [],
  channelSelection,
  values: {
    realpersonChannel: 'volcengine',
    subtitleProvider: 'volcano',
    ttsServiceUrl: 'https://tts.example.com',
    ttsSentenceLength: 18
  },
  configured: {
    deepseekOfficialApiKey: true,
    seedance2Api: true,
    minimaxH3Key: true,
    runninghubApi: true,
    volcanoSeedreamApi: true,
    grsaiImage2Key: true,
    geeknowImageApi: true,
    volcanoMediaKitApiKey: true,
    ttsApiKey: true,
    minimaxApiKey: true
  },
  customChannels: [
    {
      id: 'ch_text_demo',
      group: 'text',
      name: '自建中转 A',
      baseUrl: 'https://api.example.com/v1',
      model: 'gpt-4o',
      concurrency: 4,
      enabled: true,
      hasApiKey: true
    },
    {
      // 与目录模型 custom-anti-2-gemini-3-8-flash 同源，用来预览合并后的展示。
      id: 'ch_text_dup',
      group: 'text',
      name: '自建中转 B',
      baseUrl: 'https://tok.sonetto.top/v1/',
      model: '[anti-2]Gemini-3.8-Flash',
      concurrency: 4,
      enabled: true,
      hasApiKey: true
    }
  ]
};

// 预览用模型目录：前 5 条取自本机真实 api-config.json，后 2 条是图片组演示条目。
const previewModelCatalog = [
  { id: 'custom-anti-2-gemini-3-8-flash', kind: 'text', displayName: '[anti-2]gemini-3.8-flash', modelId: '[anti-2]gemini-3.8-flash', baseUrl: 'https://tok.sonetto.top/v1', enabled: true, provider: 'openai_compatible', hasCredential: true },
  { id: 'yd2-mini-video', kind: 'video', displayName: 'YD2.0 Mini（图生）', modelId: 'yd2-mini-video', baseUrl: '', enabled: true, provider: 'yd_video', hasCredential: true },
  { id: 'minimax-h3-video', kind: 'video', displayName: 'MiniMax H3 多图生视频', modelId: 'minimax-h3-video', baseUrl: '', enabled: true, provider: 'minimax_h3', hasCredential: true },
  { id: 'seedance-2-0-official', kind: 'video', displayName: 'Seedance 2.0 官方', modelId: 'seedance-2-0-official', baseUrl: 'https://yf.token6688.com', enabled: true, provider: 'seedance_official', hasCredential: true },
  { id: 'local-doubao-executor-video', kind: 'video', displayName: '本地豆包执行器', modelId: 'local-doubao-executor-video', baseUrl: '', enabled: true, provider: 'local_executor', hasCredential: false },
  { id: 'gpt-image-2', kind: 'image', displayName: 'gpt-image-2', modelId: 'gpt-image-2', baseUrl: 'https://grsai.ai', enabled: true, provider: 'openai_compatible', hasCredential: true },
  { id: 'gpt-image-2-5-off', kind: 'image', displayName: '[off]gpt-image-2.5', modelId: '[off]gpt-image-2.5', baseUrl: 'https://grsai.ai', enabled: false, provider: 'openai_compatible', hasCredential: false }
];

// 预览用额度假数据：只挂在文本目录模型上，用来验证「能拿到额度才显示」。
const previewModelQuotas = {
  quotas: [{
    modelId: 'custom-anti-2-gemini-3-8-flash',
    status: 'available',
    available: 128.5,
    total: 500,
    unit: '次',
    percent: 25.7,
    checkedAt: new Date().toISOString()
  }]
};

let liveIntegrationState = integrationState;

const originalFetch = window.fetch.bind(window);
window.fetch = (input, init) => {
  const url = typeof input === 'string' ? input : String(input?.url || '');
  // reveal 必须先判断，否则会被前面更宽泛的 api-integrations 命中。
  if (url.includes('/api/settings/api-integrations/reveal')) {
    let key = '';
    try {
      key = JSON.parse(init?.body || '{}').key || '';
    } catch {
      key = '';
    }
    return Promise.resolve(jsonResponse({ key, value: `${key || 'secret'}-preview-plaintext` }));
  }
  if (url.includes('/api/settings/api-integrations')) {
    if ((init?.method || 'GET').toUpperCase() === 'POST') {
      let payload = {};
      try {
        payload = JSON.parse(init?.body || '{}');
      } catch {
        payload = {};
      }
      liveIntegrationState = {
        ...liveIntegrationState,
        pinned: Array.isArray(payload.pinned) ? payload.pinned : liveIntegrationState.pinned,
        channelSelection: { ...liveIntegrationState.channelSelection, ...(payload.channelSelection || {}) },
        values: { ...liveIntegrationState.values, ...(payload.values || {}) },
        // 与 Go 侧一致：只有显式提交 customChannels 时才整表替换。
        customChannels: Array.isArray(payload.customChannels)
          ? payload.customChannels.map((channel, index) => {
            const existing = liveIntegrationState.customChannels.find(item => item.id === channel.id);
            return {
              ...channel,
              id: channel.id || `ch_preview_${index + 1}`,
              apiFormat: channel.apiFormat || existing?.apiFormat || 'openai_compatible',
              fullUrl: channel.fullUrl === true,
              hasApiKey: channel.apiKey ? true : (channel.clearApiKey ? false : existing?.hasApiKey === true)
            };
          })
          : liveIntegrationState.customChannels
      };
    }
    return Promise.resolve(jsonResponse(liveIntegrationState));
  }
  if (url.includes('/api/config/models/quotas')) return Promise.resolve(jsonResponse(previewModelQuotas));
  if (url.includes('/api/config/models')) {
    if ((init?.method || 'GET').toUpperCase() === 'PATCH') {
      // 预览里真实写回内存中的目录条目，用来验证「编辑 → 保存 → 复查」。
      const matched = url.match(/\/api\/config\/models\/([^/?#]+)/);
      const modelId = matched ? decodeURIComponent(matched[1]) : '';
      let patch = {};
      try {
        patch = JSON.parse(init?.body || '{}');
      } catch {
        patch = {};
      }
      const index = previewModelCatalog.findIndex(model => model.id === modelId);
      if (index < 0) return Promise.resolve(jsonResponse({ error: '模型不存在' }, 404));
      const next = { ...previewModelCatalog[index] };
      if (patch.displayName) next.displayName = patch.displayName;
      if (patch.modelId) next.modelId = patch.modelId;
      if (patch.baseUrl) next.baseUrl = patch.baseUrl;
      if (patch.credential) next.hasCredential = true;
      if (patch.clearCredential === true) next.hasCredential = false;
      previewModelCatalog[index] = next;
      return Promise.resolve(jsonResponse({ model: next }));
    }
    // 与真实后端一致：只返回可用模型（凭证非空或本地执行器），并支持 kind 过滤。
    const kind = new URL(url, window.location.origin).searchParams.get('kind');
    const models = previewModelCatalog
      .filter(model => isCatalogModelAvailable(model))
      .filter(model => !kind || model.kind === kind);
    return Promise.resolve(jsonResponse({ models }));
  }
  if (url.includes('/api/config')) {
    return Promise.resolve(jsonResponse({
      storageRoot: '',
      productionRetentionDays: 7,
      pet: 'cm-penguin',
      notifications: { soundEnabled: true, soundVolume: 60, petVisible: true },
      canManageApi: true
    }));
  }
  if (url.includes('/api/shuihuo-production/giant-material-executors')) return Promise.resolve(jsonResponse({ executors: [] }));
  if (url.includes('/downloads/')) return Promise.resolve(new Response('', { status: 404 }));
  if (url.includes('/api/')) return Promise.resolve(jsonResponse({}));
  return originalFetch(input, init);
};

createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <ConfigProvider theme={createAntTheme(theme, theme)}>
      {view === 'account'
        ? <ApiConfigPage />
        : <SettingsPage themePreference={theme} onThemeChange={() => {}} />}
    </ConfigProvider>
  </React.StrictMode>
);

// 预览用：?reveal=1 自动点一次眼睛，方便截图对比掩码与明文。
if (params.get('reveal') === '1') {
  window.setTimeout(() => {
    document.querySelectorAll('.api-secret-eye[title*="查看明文"]').forEach(button => button.click());
  }, 1200);
}

// 预览用交互助手：headless 截图时网络加载是真实时间、定时器是虚拟时间，
// 固定 setTimeout 不可靠，改为等元素出现后立即执行。
function whenElement(selector, callback) {
  const existing = document.querySelector(selector);
  if (existing) {
    callback(existing);
    return;
  }
  const observer = new MutationObserver(() => {
    const element = document.querySelector(selector);
    if (!element) return;
    observer.disconnect();
    callback(element);
  });
  observer.observe(document.body, { childList: true, subtree: true });
}

// 预览用：?dialog=1 打开「添加自定义模型」表单；?dialog=edit 打开第一条的编辑表单。
if (params.get('dialog') === '1') {
  whenElement('.api-custom-open', trigger => trigger.click());
}
// 预览用：?dialog=preset 打开「自定义预设」勾选弹窗。
if (params.get('dialog') === 'preset') {
  whenElement('.api-preset-open', trigger => trigger.click());
}
if (params.get('dialog') === 'edit') {
  // 只点「自定义渠道」卡，目录模型卡用自己的 ?dialog=catalog。
  whenElement('.api-channel-card:not(.is-catalog) .api-channel-edit-btn', trigger => trigger.click());
}
// ?dialog=catalog 打开第一条「目录模型」的编辑表单（用户自建模型走同一张表单）。
if (params.get('dialog') === 'catalog') {
  whenElement('.api-channel-card.is-catalog .api-channel-edit-btn', trigger => trigger.click());
}

// 预览用：?merged=1 展开目录卡里的同源渠道编辑区。
if (params.get('merged') === '1') {
  whenElement('.api-catalog-merged-toggle', toggle => toggle.click());
}

// 预览用：?expand=subtitle,tts 展开默认收起的非渠道组。
const expandIds = params.get('expand');
if (expandIds) {
  expandIds.split(',').map(item => item.trim()).filter(Boolean).forEach(groupId => {
    whenElement(`#api-group-${groupId} .api-group-head`, head => head.click());
  });
}
