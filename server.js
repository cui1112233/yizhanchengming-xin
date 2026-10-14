const os = require('os');
const express = require('express');

const { hydrateMissingEnvFromFile } = require('./lib/runtime-env');
hydrateMissingEnvFromFile(
  process.env.QIANTIE_BASE_ENV_FILE || '/run/qiantie/base.env',
  ['QIANTIE_BRIDGE_SECRET']
);
// Local development runs the Go service on 4001 and the compatibility Node
// service on 3001. Keep the local bridge aligned even when the user starts
// Node directly without copying a long environment-variable command.
if (!String(process.env.QIANTIE_GO_BASE_URL || '').trim() && String(process.env.QIANTIE_NODE_PORT || '') === '3001') {
  process.env.QIANTIE_GO_BASE_URL = 'http://127.0.0.1:4001';
}
if (process.env.NODE_ENV === 'production' && !String(process.env.QIANTIE_BRIDGE_SECRET || '').trim()) {
  throw new Error('QIANTIE_BRIDGE_SECRET is required in production');
}

const { HOST, PORT } = require('./lib/shared');
const { createApp } = require('./app');
const { apiAuth } = require('./middleware/auth');
const { createNovelFetchV2PageMiddleware } = require('./lib/novel-fetch-workshop/v2-page');
const { attachV78NovelFetchV2 } = require('./lib/novel-fetch-workshop/v2-compose');
const { migrateLegacyScriptPromptPresets } = require('./lib/script-prompt-preset-migration');

// ============================================================
// 启动服务器
// ============================================================
// V78 小说获取 V2 通过一个很薄的前置扩展层增强旧工作台；
// 其余请求继续原样进入现有 V78 Express 应用，避免重写旧兼容页面或 API。
const app = express();
const coreApp = createApp();
const nodeLegacyAPIEnabled = String(process.env.QIANTIE_NODE_LEGACY_API || '') === '1';

// 2026-09-06 剧本提示词收口：只有仍等于旧系统默认正文的后台预设才自动
// 升级为当前“分段开头 / 分镜模式 / 通用规则”元提示词。管理员已经编辑过
// 任意正文时哈希会改变，因此迁移器会保留该版本，不做覆盖。
migrateLegacyScriptPromptPresets(coreApp.locals.presetStore, 'choushiyiguai');

// Shell-level V31 routes must authenticate against exactly the same runtime as
// coreApp; otherwise a valid core session would be rejected before reaching it.
app.locals.authRuntime = coreApp.locals.authRuntime;
app.get(['/batch-rewrite/index.html', '/batch-rewrite/'], createNovelFetchV2PageMiddleware());
attachV78NovelFetchV2({ shellApp: app, coreApp, bodyParser: express.json({ limit: '50mb' }) });

// V88 script smart-unified semantics: when the user explicitly selects
// “画面前缀词 → 智能统一”, run an independent full-source visual analysis first.
// The model may only return the eleven film-level fields; the backend rebuilds
// the authoritative final prefix deterministically and never accepts a model-
// supplied mega prompt with per-shot focal lengths/actions/light positions.
if (nodeLegacyAPIEnabled) {
  const { createScriptSmartUnifiedStyleHandler } = require('./lib/script-smart-unified-route');
  app.post(
    '/api/script/smart-unified-style',
    express.json({ limit: '50mb' }),
    apiAuth,
    createScriptSmartUnifiedStyleHandler()
  );
}

// 快速导演 / 匹配音频保留独立导演事务。普通“分段开头 + 分镜模式”不走
// 该接口，而是在 /api/chat 的一次 script 请求中组合用户当前选择的开头、
// 输出模式和 10s/15s 规则，直接返回最终分镜成品。
if (nodeLegacyAPIEnabled) {
  const { createScriptDirectorPipelineHandler } = require('./lib/script-director-pipeline-route');
  app.post(
    '/api/script/director-pipeline',
    express.json({ limit: '50mb' }),
    apiAuth,
    createScriptDirectorPipelineHandler()
  );
}

// V78.3.0.31 final public identity: preserve the existing build-info payload but
// overlay the verified final-semantics version/build fields before it is emitted.
if (nodeLegacyAPIEnabled) {
  const { createV783031BuildInfoMiddleware } = require('./lib/novel-panel/v783031-build-info');
  app.use('/api/novel-panel/build-info', createV783031BuildInfoMiddleware());
}

// Git-direct release identity is exposed without changing the legacy build-info
// compatibility fields consumed by existing pages and diagnostics.
if (nodeLegacyAPIEnabled) {
  const { readReleaseInfo } = require('./lib/release-info');
  app.get('/api/build-info', (req, res) => {
    res.json({
      app_version: 'v78.3.0.3',
      build_id: 'v78.3.0.3-remote-workbench-20260819-r1',
      ...readReleaseInfo()
    });
  });
}

if (nodeLegacyAPIEnabled) {
  const { createRuntimeBuildInfoHandler } = require('./lib/runtime-build-info');
  app.get('/api/runtime-build-info', createRuntimeBuildInfoHandler({ service: 'v88-node' }));
}

// V78.3.0.20/21 final semantics: full outline generation uses one authoritative
// backend semantic-contract/audit chain. It compiles the user's director rules,
// generates the outline, audits every contract item, repairs only failed source
// rows once while preserving their duration, then re-audits before returning.
if (nodeLegacyAPIEnabled) {
  const { createV783031OutlineHandler } = require('./lib/novel-panel/v783031-outline-route');
  app.post(
    '/api/novel-panel/outline-scenes',
    express.json({ limit: '50mb' }),
    apiAuth,
    createV783031OutlineHandler({ coreApp })
  );
}

// V78.3.0.27/28 final semantics: single-card regeneration is a current-truth Patch.
// Parse/enrich the request before the legacy core route, then hard-validate the
// legacy route response before it can be accepted by the browser writeback path.
// Authentication, model config and usage accounting still execute in coreApp.
if (nodeLegacyAPIEnabled) {
  const { createV783031RegenerationMiddleware } = require('./lib/novel-panel/v783031-regeneration-middleware');
  app.use('/api/novel-panel/regenerate-scene-outline', express.json({ limit: '50mb' }), createV783031RegenerationMiddleware());
}
app.use(coreApp);

app.listen(PORT, HOST, () => {
  const interfaces = os.networkInterfaces();
  console.log('========================================');
  console.log('  Server running on (Express):');
  console.log('  本机访问: http://127.0.0.1:' + PORT);
  for (const [name, addrs] of Object.entries(interfaces)) {
    for (const addr of addrs) {
      if (addr.family === 'IPv4' && !addr.internal) {
        console.log('  局域网访问: http://' + addr.address + ':' + PORT);
      }
    }
  }
  console.log('========================================');
});
