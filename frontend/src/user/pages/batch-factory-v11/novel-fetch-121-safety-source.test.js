import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const pagesDir = path.resolve(here, '..');
const frontendDir = path.resolve(here, '../../../..');
const repoDir = path.resolve(frontendDir, '..');
const appSource = fs.readFileSync(path.join(frontendDir, 'src/user/App.jsx'), 'utf8');

function readRequired(relativePath) {
  return fs.readFileSync(path.join(frontendDir, relativePath), 'utf8');
}

test('novel fetch keeps the complete V78 workbench while serving it from Go instead of 3001', () => {
  assert.match(appSource, /const NovelFetchWorkshopPage = lazy\(\(\) => import\('\.\/pages\/NovelFetchWorkshopPage'\)\);/);
  assert.match(appSource, /'\/novel-fetch': NovelFetchPage/);

  const pageSource = fs.readFileSync(path.join(pagesDir, 'NovelFetchWorkshopPage.jsx'), 'utf8');
  assert.match(pageSource, /novelFetchWorkshop/);
  const legacyPageSource = fs.readFileSync(path.join(pagesDir, 'NovelFetchPage.jsx'), 'utf8');
  assert.match(legacyPageSource, /src=\{`\/batch-rewrite\/v2/);
  assert.doesNotMatch(legacyPageSource, /3001/);
});

test('Go-native novel fetch exposes the target upload flow instead of returning to the legacy page', () => {
  const pageSource = fs.readFileSync(path.join(pagesDir, 'NovelFetchWorkshopPage.jsx'), 'utf8');
  const apiSource = readRequired('src/shared/api/novelFetch.js');
  assert.match(pageSource, /uploadLogin/);
  assert.match(pageSource, /getUploadSession/);
  assert.match(pageSource, /uploadBatch/);
  assert.match(pageSource, /label: '上传'/);
  assert.match(pageSource, /uploadBatch/);
  assert.match(apiSource, /\/api\/novel-fetch-upload\/upload-batch/);
  assert.doesNotMatch(pageSource, /window\.location\.assign\(['"]\/novel-fetch['"]\)/);
});

test('novel fetch iframe accepts the parent light-theme sync and applies readable surfaces', () => {
  const appSource = fs.readFileSync(path.join(frontendDir, 'public/batch-rewrite/app.js'), 'utf8');
  const layoutSource = fs.readFileSync(path.join(repoDir, 'public/batch-rewrite/v78-novel-fetch-v2-layout.js'), 'utf8');
  assert.match(appSource, /qiantie-theme-sync/);
  assert.match(appSource, /qiantie-theme-sync/);
  assert.match(appSource, /dataset\.theme/);
  assert.match(layoutSource, /\[data-theme=['\"]light['\"]\]/);
  assert.match(layoutSource, /#v78ImageStatusPanel #processResult[^}]*background\s*:\s*var\(--panel/);
});

test('novel fetch keeps a visible loading experience until the iframe is ready', () => {
  const pageSource = fs.readFileSync(path.join(pagesDir, 'NovelFetchPage.jsx'), 'utf8');
  const styleSource = fs.readFileSync(path.join(pagesDir, 'novel-fetch.css'), 'utf8');
  const appSource = fs.readFileSync(path.join(frontendDir, 'public/batch-rewrite/app.js'), 'utf8');
  assert.match(pageSource, /novel-fetch-loading-overlay/);
  assert.match(pageSource, /正在加载小说获取/);
  assert.match(pageSource, /onError/);
  assert.match(pageSource, /setFrameError/);
  assert.match(pageSource, /qiantie:novel-fetch-ready/);
  assert.match(appSource, /qiantie:novel-fetch-ready/);
  assert.match(styleSource, /novel-fetch-loading-overlay/);
  assert.match(styleSource, /novel-fetch-loader-trace/);
  assert.match(styleSource, /prefers-reduced-motion/);
});

test('121 guard fails closed when batch rewrite config is unavailable without hijacking login submit', () => {
  const source = readRequired('public/batch-rewrite/121-login-hotfix.js');
  assert.match(source, /state\.config\?\.web_submit/);
  assert.match(source, /ensureServerConfig/);
  assert.match(source, /已阻止发送 121 账号密码/);
  assert.match(source, /配置加载失败/);
  assert.doesNotMatch(source, /stopImmediatePropagation/);
  assert.doesNotMatch(source, /qiantieSafeLoginBound/);
});

test('121 guard keeps a 65 second frontend deadline above active worker budgets', () => {
  const source = readRequired('public/batch-rewrite/121-login-hotfix.js');
  assert.match(source, /REQUEST_TIMEOUT_MS\s*=\s*65_?000/);
  assert.match(source, /AbortController/);
  assert.match(source, /REQUEST_TIMEOUT_SECONDS\s*=\s*Math\.ceil\(REQUEST_TIMEOUT_MS\s*\/\s*1000\)/);
});

test('121 backend transport also defaults to a 15 second deadline', () => {
  const source = fs.readFileSync(path.join(repoDir, 'lib/target-upload.js'), 'utf8');
  assert.match(source, /timeoutMs\s*=\s*15000/);
});

test('candidate image injects the guard even when tracked frontend dist is stale', () => {
  const dockerfile = fs.readFileSync(path.join(repoDir, 'Dockerfile'), 'utf8');
  assert.match(dockerfile, /COPY frontend\/public\/batch-rewrite\/121-login-hotfix\.js/);
  assert.match(dockerfile, /qiantie-121-login-hotfix/);
  assert.match(dockerfile, /frontend\/dist\/batch-rewrite\/index\.html/);
});

test('novel status center survives frontend rebuild and direct ECS image packaging', () => {
  const source = readRequired('public/batch-rewrite/v88-novel-status-center.js');
  assert.match(source, /V88NovelStatusCenter/);
  assert.match(source, /处理完成/);
  assert.match(source, /视频管理系统提交/);

  const dockerfile = fs.readFileSync(path.join(repoDir, 'Dockerfile'), 'utf8');
  assert.match(dockerfile, /COPY frontend\/dist\/ \.\/frontend\/dist\//);
  assert.match(dockerfile, /COPY frontend\/public\/batch-rewrite\/121-login-hotfix\.js/);
});

test('121 login delegates to the asynchronous operation flow', () => {
  const source = readRequired('public/batch-rewrite/121-login-hotfix.js');
  assert.doesNotMatch(source, /function bindSafeLoginSubmit/);
  assert.doesNotMatch(source, /stopImmediatePropagation/);
  assert.doesNotMatch(source, /api\('\/api\/web-submit\/config'/);
  const appSource = readRequired('public/batch-rewrite/app.js');
  assert.match(appSource, /api\("\/api\/web-submit\/operations"/);
  assert.match(appSource, /waitWebSubmitOperation\(operation\.id\)/);
  assert.match(appSource, /121 后台操作等待超时/);
  assert.match(appSource, /passwordInput\.value\s*=\s*""/);
});
