import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const createModal = fs.readFileSync(path.join(here, 'BatchFactoryCreateModal.jsx'), 'utf8');
const unifiedSettings = fs.readFileSync(path.join(here, 'BatchFactoryUnifiedSettingsModal.jsx'), 'utf8');

test('新建批量统一配置：发布页签按当前生产模式只渲染一套内容', () => {
  assert.match(createModal, /const isComicProduction = initialBatchSettings\.productionMode === 'comic';/);
  assert.match(createModal, /key: 'publish', label: '发布统一', children: isComicProduction \? <BatchFactoryComicPublishSettingsForm/);
  assert.match(createModal, /: <BatchFactoryPublishSettingsForm value=\{initialBatchSettings\}/);
  // 旧的硬编码视频内容会让漫画页签渲染视频配置。
  assert.doesNotMatch(createModal, /key: 'publish', label: '发布统一', children: <BatchFactoryPublishSettingsForm/);
});

test('新建批量统一配置：页签与内容区共用同一个生产模式来源', () => {
  assert.match(createModal, /BatchFactoryProductionModeControl value=\{initialBatchSettings\.productionMode\}/);
  assert.match(createModal, /生产模式<Select value=\{initialBatchSettings\.productionMode \|\| 'video'\}/);
});

test('配置类解析失败只发 1 次请求、只留 1 条 banner，其余行保持 pending', () => {
  assert.ok(createModal.includes("new Set(['QINGYU_AUTH_FAILED', 'QINGYU_AUTH_NOT_CONFIGURED', 'GIANT_PLUGIN_NOT_CONFIGURED'])"));
  assert.match(createModal, /if \(giantConfigBlockingCodes\.has\(code\)\) \{\s*setGiantError\(code\);\s*break;\s*\}/);
  assert.equal((createModal.match(/setGiantError\(code\)/g) || []).length, 1);
  assert.equal((createModal.match(/服务器令牌失效，需更新青语 N8 后台登录令牌后重新解析/g) || []).length, 1);
  // 配置类分支里不得把当前行或后续行标成 error。
  const blockingBlock = createModal.slice(
    createModal.indexOf('if (giantConfigBlockingCodes.has(code))'),
    createModal.indexOf('setGiantItems(current => current.map(row => row.id === id', createModal.indexOf('if (giantConfigBlockingCodes.has(code))'))
  );
  assert.doesNotMatch(blockingBlock, /status: 'error'/);
  assert.match(createModal, /giantError \? `等待重新解析 \$\{item\.id\}` : `正在解析 \$\{item\.id\}…`/);
});

test('视频与漫画发布配置分别写入 publishSettings 与 comicPublishSettings，不串写', () => {
  const videoForm = unifiedSettings.slice(
    unifiedSettings.indexOf('export function BatchFactoryPublishSettingsForm'),
    unifiedSettings.indexOf('export function BatchFactoryComicPublishSettingsForm')
  );
  assert.match(videoForm, /patch\(\{ publishSettings:/);
  assert.doesNotMatch(videoForm, /comicPublishSettings:/);

  const comicForm = unifiedSettings.slice(
    unifiedSettings.indexOf('export function BatchFactoryComicPublishSettingsForm'),
    unifiedSettings.indexOf('export function BatchFactoryAiReasoningForm')
  );
  assert.match(comicForm, /comicPublishSettings: \{ \.\.\.comicPublishSettings, \.\.\.next \}/);
  assert.doesNotMatch(comicForm, /publishSettings:/);
});
