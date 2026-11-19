import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const component = fs.readFileSync(path.join(here, 'UploadSentButton.jsx'), 'utf8');
const stylesheet = fs.readFileSync(path.join(here, 'upload-sent-button.css'), 'utf8');
const pageSource = fs.readFileSync(
  path.join(here, '..', '..', 'user', 'pages', 'shuihuo', 'BatchFactoryNovelList.jsx'),
  'utf8'
);

// 统一上传按钮：点击轻压 → 提交成功后纸飞机飞出并显示「上传成功」→ 自动恢复。
test('provides the shared upload-sent button and its reset hook', () => {
  assert.match(component, /export function useUploadSentState\(/);
  assert.match(component, /export function UploadSentButton\(/);
  assert.match(component, /batch-factory-upload-sent-success/);
  assert.match(component, /batch-factory-upload-sent-plane/);
  assert.match(component, /sentLabel = '上传成功'/);
});

test('stylesheet keeps the press, plane fly-out and reduced-motion fallback', () => {
  assert.match(stylesheet, /\.batch-factory-upload-sent:not\(:disabled\):active/);
  assert.match(stylesheet, /batch-factory-upload-sent-plane-fly/);
  assert.match(stylesheet, /@media \(prefers-reduced-motion: reduce\)/);
  assert.match(stylesheet, /\.batch-factory-upload-network-button:not\(:disabled\):active/);
  assert.match(stylesheet, /\.batch-factory-comic-publish-entry:not\(:disabled\):active/);
});

test('upload network and comic publish confirmations both use the sent animation', () => {
  assert.match(pageSource, /import \{ UploadSentButton, useUploadSentState \} from '\.\.\/\.\.\/\.\.\/shared\/components\/UploadSentButton'/);
  // 上传网络：全部提交成功后才触发纸飞机，失败保持原状。
  assert.match(pageSource, /uploadSent\.markSent\(\)/);
  assert.match(pageSource, /<UploadSentButton danger type="primary" onClick=\{submitTo121\} loading=\{busy\} sent=\{uploadSent\.sent\}/);
  // 发布漫画：确认弹窗改为自定义 footer，成功后延迟收起让动画播完。
  assert.match(pageSource, /comicPublishSent\.markSent\(\)/);
  assert.match(pageSource, /window\.setTimeout\(\(\) => setPublishTarget\(null\), 1700\)/);
  assert.match(pageSource, /<UploadSentButton type="primary" sent=\{comicPublishSent\.sent\}/);
});
