import assert from 'node:assert/strict';
import test from 'node:test';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';

import {
  BATCH_ACCESS_MESSAGE_KEY,
  batchAccessErrorMessage,
  batchAccessNotice,
  forgetActiveBatch,
  isStaleBatchAccessError
} from './batchFactoryBatchAccessNotice.js';
import { BATCH_FACTORY_ACTIVE_BATCH_STORAGE_KEY } from './batchFactoryNovelFetchHandoff.js';

const here = dirname(fileURLToPath(import.meta.url));
const workbenchSource = readFileSync(join(here, '..', 'BatchFactoryWorkbenchPage.jsx'), 'utf8');
const productionSource = readFileSync(join(here, '..', 'ShuihuoProductionPage.jsx'), 'utf8');

test('跨账号批次显示服务端可读文案，而不是裸 not found', () => {
  const message = batchAccessErrorMessage({ status: 403, code: 'batch_forbidden', message: '该批次属于其他账号，请用对应账号登录查看' });
  assert.match(message, /该批次属于其他账号/);
  assert.doesNotMatch(message, /not found/i);
});

test('普通账号拿到通用文案时也不会出现英文 not found', () => {
  const message = batchAccessErrorMessage({ status: 403, code: 'batch_forbidden', message: 'not found' });
  assert.match(message, /该批次属于其他账号，当前账号无数据/);
  assert.doesNotMatch(message, /not found/i);
});

test('批次不存在或属于其他账号时明确显示当前账号无数据', () => {
  const message = batchAccessErrorMessage({ status: 404, code: 'batch_not_found', message: 'not found' });
  assert.match(message, /当前账号无数据/);
});

test('媒体库 403/404 不再把空 pages 显示成 0/35', () => {
  const source = readFileSync(join(here, 'BatchFactoryNovelList.jsx'), 'utf8');
  assert.match(source, /pageLoadError && !pages\.length/, '必须区分「无权访问」与「真的 0 页」');
  assert.match(source, /pageAccessMessage \? <div className="batch-factory-comic-media-completion"/, '403/404 时完成度必须显示无数据');
  assert.match(source, /batchAccessErrorMessage\(error\)/, '错误必须翻译成可读的归属文案');
});

test('错误提示固定 key，同一错误只保留一条', () => {
  const notice = batchAccessNotice({ status: 403, code: 'batch_forbidden', message: '该批次属于其他账号' });
  assert.equal(notice.key, BATCH_ACCESS_MESSAGE_KEY);
  assert.equal(notice.duration, 8);
  assert.equal(batchAccessNotice({ status: 404, code: 'batch_not_found' }).key, notice.key);
});

test('只有 403/404 才清理失效批次 id，网络错误保留重试', () => {
  assert.equal(isStaleBatchAccessError({ status: 403, code: 'batch_forbidden' }), true);
  assert.equal(isStaleBatchAccessError({ status: 404, code: 'batch_not_found' }), true);
  assert.equal(isStaleBatchAccessError({ status: 0 }), false);
});

test('清理活跃批次 id 时不会误删其他批次', () => {
  const store = new Map();
  globalThis.localStorage = {
    getItem: key => (store.has(key) ? store.get(key) : null),
    removeItem: key => store.delete(key)
  };
  store.set(BATCH_FACTORY_ACTIVE_BATCH_STORAGE_KEY, 'batch_other');
  forgetActiveBatch('batch_stale');
  assert.equal(store.get(BATCH_FACTORY_ACTIVE_BATCH_STORAGE_KEY), 'batch_other');
  forgetActiveBatch('batch_other');
  assert.equal(store.has(BATCH_FACTORY_ACTIVE_BATCH_STORAGE_KEY), false);
});

test('批次详情请求关掉全局错误条并只走单条可关闭提示', () => {
  for (const [name, source] of [['BatchFactoryWorkbenchPage.jsx', workbenchSource], ['ShuihuoProductionPage.jsx', productionSource]]) {
    assert.match(source, /getBatch\([^)]*suppressGlobalError: true/, `${name} 必须抑制全局错误条，避免同一次失败弹两条`);
    assert.match(source, /notifyBatchAccessFailure\(/, `${name} 必须使用统一提示`);
    assert.match(source, /isStaleBatchAccessError\([^)]*\)[\s\S]{0,160}forgetActiveBatch\(/, `${name} 失败后必须清理失效批次 id`);
    assert.doesNotMatch(source, /message\.error\(requestError\?\.message \|\| '读取批量工程失败'\)/, `${name} 不能继续吞错成裸提示`);
  }
  assert.match(workbenchSource, /message\.destroy\(notice\.key\)/, '提示必须可关闭');
});
