const fs = require('node:fs');
const path = require('node:path');

const LOGIN_HOTFIX_PATH = '/batch-rewrite/121-login-hotfix.js';
const INTERACTION_FEEDBACK_PATH = '/batch-rewrite/interaction-feedback.js';
const TASK_VISIBILITY_PATH = '/batch-rewrite/task-visibility-hotfix.js';
const RUNTIME_PATH = '/batch-rewrite/v78-novel-fetch-v2-runtime.js';
const APP_PATH = '/batch-rewrite/app.js';
const SELECTED_TASK_SUBMIT_BRIDGE_PATH = '/batch-rewrite/v78-selected-task-submit-bridge.js';
const CLIENT_PATH = '/batch-rewrite/v78-novel-fetch-v2.js';
const DATE_CLIENT_PATH = '/batch-rewrite/v78-novel-fetch-v2-date.js';
const CONFIG_CLIENT_PATH = '/batch-rewrite/v78-novel-fetch-v2-config.js';
const LAYOUT_CLIENT_PATH = '/batch-rewrite/v78-novel-fetch-v2-layout.js';
const QUICK_SUBMIT_SELECTION_HOTFIX_PATH = '/batch-rewrite/v78-quick-submit-selection-hotfix.js';
const SELECTED_SUBMIT_PAYLOAD_GUARD_PATH = '/batch-rewrite/v78-selected-submit-payload-guard.js';
const RUN_CONTROLS_CLIENT_PATH = '/batch-rewrite/v78-novel-fetch-v2-run-controls.js';
const VERSION_CONFIG_AUTHORITY_PATH = '/batch-rewrite/version-config-authority.js';
const LOGIN_HOTFIX_TAG = `<script id="qiantie-121-login-hotfix" src="${LOGIN_HOTFIX_PATH}?v=20260915-session-probe-r2"></script>`;
const INTERACTION_FEEDBACK_TAG = `<script id="qiantie-novel-fetch-interaction-feedback" src="${INTERACTION_FEEDBACK_PATH}?v=20260915-parent-status-r1"></script>`;
const TASK_VISIBILITY_TAG = `<script id="qiantie-novel-fetch-task-visibility" src="${TASK_VISIBILITY_PATH}?v=20260906-task-visibility1"></script>`;
const RUNTIME_TAG = `<script id="qiantie-novel-fetch-runtime" src="${RUNTIME_PATH}?v=20260916-performance-r1"></script>`;
const APP_CLIENT_TAG = `<script src="${APP_PATH}?v=20260922-retry-progress-r1"></script>`;
const SELECTED_TASK_SUBMIT_BRIDGE_TAG = `<script src="${SELECTED_TASK_SUBMIT_BRIDGE_PATH}?v=20260917-task-list-fallback-r1"></script>`;
const CLIENT_TAG = `<script src="${CLIENT_PATH}?v=20260925-processing-flow-r1"></script>`;
const DATE_CLIENT_TAG = `<script src="${DATE_CLIENT_PATH}?v=20260922-rendered-date-r1"></script>`;
const CONFIG_CLIENT_TAG = `<script src="${CONFIG_CLIENT_PATH}?v=20260916-performance-r1"></script>`;
const LAYOUT_CLIENT_TAG = `<script src="${LAYOUT_CLIENT_PATH}?v=20261013-submit-feedback-r3"></script>`;
const QUICK_SUBMIT_SELECTION_HOTFIX_TAG = `<script src="${QUICK_SUBMIT_SELECTION_HOTFIX_PATH}?v=20260917-capture-before-tab-r1"></script>`;
const SELECTED_SUBMIT_PAYLOAD_GUARD_TAG = `<script src="${SELECTED_SUBMIT_PAYLOAD_GUARD_PATH}?v=20260917-payload-guard-r1"></script>`;
const RUN_CONTROLS_CLIENT_TAG = `<script src="${RUN_CONTROLS_CLIENT_PATH}?v=20260925-text-model-layout-r2"></script>`;
const VERSION_CONFIG_AUTHORITY_TAG = `<script src="${VERSION_CONFIG_AUTHORITY_PATH}?v=20260909-single-authority-r1"></script>`;
const LOGIN_HOTFIX_TAG_RE = /<script\b[^>]*\bsrc=["'](?:\.\/|\/batch-rewrite\/)121-login-hotfix\.js(?:\?[^"']*)?["'][^>]*><\/script>/i;
const APP_SCRIPT_TAG_RE = /<script\b[^>]*\bsrc=["'][^"']*\/?app\.js(?:\?[^"']*)?["'][^>]*><\/script>/i;

function escapeRegExp(value) { return String(value).replace(/[.*+?^${}()|[\]\\]/g, '\\$&'); }
function dedupeScriptPath(html, scriptPath) {
  const pattern = new RegExp(`<script\\b[^>]*\\bsrc=["'][^"']*${escapeRegExp(scriptPath)}(?:\\?[^"']*)?["'][^>]*><\\/script>`, 'ig');
  let retained = false;
  return String(html).replace(pattern, tag => {
    if (retained) return '';
    retained = true;
    return tag;
  });
}

function injectNovelFetchV2Script(html) {
  let source = String(html || '');
  if (!source.includes('</body>')) return source;

  for (const scriptPath of [LOGIN_HOTFIX_PATH, INTERACTION_FEEDBACK_PATH, TASK_VISIBILITY_PATH, RUNTIME_PATH, SELECTED_TASK_SUBMIT_BRIDGE_PATH, CLIENT_PATH, DATE_CLIENT_PATH, CONFIG_CLIENT_PATH, LAYOUT_CLIENT_PATH, QUICK_SUBMIT_SELECTION_HOTFIX_PATH, SELECTED_SUBMIT_PAYLOAD_GUARD_PATH, RUN_CONTROLS_CLIENT_PATH, VERSION_CONFIG_AUTHORITY_PATH]) {
    source = dedupeScriptPath(source, scriptPath);
  }

  if (LOGIN_HOTFIX_TAG_RE.test(source)) {
    source = source.replace(LOGIN_HOTFIX_TAG_RE, LOGIN_HOTFIX_TAG);
  }

  const tags = [];
  let runtimeInjected = source.includes(RUNTIME_PATH);
  if (!runtimeInjected) {
    const appScript = source.match(APP_SCRIPT_TAG_RE);
    if (appScript) {
      source = source.replace(APP_SCRIPT_TAG_RE, `${RUNTIME_TAG}\n    ${appScript[0]}`);
      runtimeInjected = true;
    }
  }
  const appScript = source.match(APP_SCRIPT_TAG_RE);
  if (appScript) source = source.replace(APP_SCRIPT_TAG_RE, APP_CLIENT_TAG);
  if (!source.includes(LOGIN_HOTFIX_PATH)) tags.push(LOGIN_HOTFIX_TAG);
  if (!source.includes(INTERACTION_FEEDBACK_PATH)) tags.push(INTERACTION_FEEDBACK_TAG);
  if (!source.includes(TASK_VISIBILITY_PATH)) tags.push(TASK_VISIBILITY_TAG);
  if (!source.includes(SELECTED_TASK_SUBMIT_BRIDGE_PATH)) tags.push(SELECTED_TASK_SUBMIT_BRIDGE_TAG);
  if (!source.includes(CLIENT_PATH)) tags.push(CLIENT_TAG);
  if (!source.includes(DATE_CLIENT_PATH)) tags.push(DATE_CLIENT_TAG);
  if (!source.includes(CONFIG_CLIENT_PATH)) tags.push(CONFIG_CLIENT_TAG);
  if (!source.includes(LAYOUT_CLIENT_PATH)) tags.push(LAYOUT_CLIENT_TAG);
  if (!source.includes(QUICK_SUBMIT_SELECTION_HOTFIX_PATH)) tags.push(QUICK_SUBMIT_SELECTION_HOTFIX_TAG);
  if (!source.includes(SELECTED_SUBMIT_PAYLOAD_GUARD_PATH)) tags.push(SELECTED_SUBMIT_PAYLOAD_GUARD_TAG);
  if (!source.includes(RUN_CONTROLS_CLIENT_PATH)) tags.push(RUN_CONTROLS_CLIENT_TAG);
  if (!source.includes(VERSION_CONFIG_AUTHORITY_PATH)) tags.push(VERSION_CONFIG_AUTHORITY_TAG);
  if (!runtimeInjected) tags.unshift(RUNTIME_TAG);
  if (!tags.length) return source;
  return source.replace('</body>', `    ${tags.join('\n    ')}\n  </body>`);
}

function createNovelFetchV2PageMiddleware({ indexPath } = {}) {
  const trackedIndexPath = path.join(__dirname, '..', '..', 'frontend', 'public', 'batch-rewrite', 'index.html');
  const distIndexPath = path.join(__dirname, '..', '..', 'frontend', 'dist', 'batch-rewrite', 'index.html');
  const resolved = indexPath || (fs.existsSync(trackedIndexPath) ? trackedIndexPath : distIndexPath);
  return (req, res, next) => {
    fs.readFile(resolved, 'utf8', (error, html) => {
      if (error) return next();
      res.setHeader('Content-Type', 'text/html; charset=utf-8');
      res.setHeader('Cache-Control', 'no-store, no-cache, must-revalidate, proxy-revalidate');
      res.setHeader('Pragma', 'no-cache');
      res.setHeader('Expires', '0');
      return res.send(injectNovelFetchV2Script(html));
    });
  };
}

module.exports = {
  LOGIN_HOTFIX_PATH,
  INTERACTION_FEEDBACK_PATH,
  TASK_VISIBILITY_PATH,
  RUNTIME_PATH,
  SELECTED_TASK_SUBMIT_BRIDGE_PATH,
  CLIENT_PATH,
  DATE_CLIENT_PATH,
  CONFIG_CLIENT_PATH,
  LAYOUT_CLIENT_PATH,
  QUICK_SUBMIT_SELECTION_HOTFIX_PATH,
  SELECTED_SUBMIT_PAYLOAD_GUARD_PATH,
  RUN_CONTROLS_CLIENT_PATH,
  VERSION_CONFIG_AUTHORITY_PATH,
  injectNovelFetchV2Script,
  createNovelFetchV2PageMiddleware
};
