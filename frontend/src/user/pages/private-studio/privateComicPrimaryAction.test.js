import assert from 'node:assert/strict';
import test from 'node:test';

import { privateComicPrimaryAction } from './privateComicPrimaryAction.js';

test('漫画主按钮在资产完成后先生成提示词，再生成缺失图片', () => {
  const base = {
    sourceReady: true,
    taskReady: true,
    assetCount: 3,
    pagesTotal: 0,
    promptReadyCount: 0,
    imageReadyCount: 0,
    busy: false,
    action: ''
  };

  assert.deepEqual(privateComicPrimaryAction({ ...base, assetCount: 0 }), {
    key: 'prompts',
    label: '生成漫画提示词',
    disabled: true
  });
  assert.deepEqual(privateComicPrimaryAction(base), {
    key: 'prompts',
    label: '生成漫画提示词',
    disabled: false
  });
  assert.deepEqual(privateComicPrimaryAction({ ...base, pagesTotal: 3, promptReadyCount: 1 }), {
    key: 'prompts',
    label: '补齐漫画提示词',
    disabled: false
  });
  assert.deepEqual(privateComicPrimaryAction({ ...base, pagesTotal: 3, promptReadyCount: 3, imageReadyCount: 1 }), {
    key: 'images',
    label: '生成漫画图片',
    disabled: false
  });
  assert.deepEqual(privateComicPrimaryAction({ ...base, pagesTotal: 3, promptReadyCount: 3, imageReadyCount: 3 }), {
    key: 'complete',
    label: '漫画图片已完成',
    disabled: true
  });
});

test('漫画主按钮在运行时显示当前阶段且禁止重复点击', () => {
  const base = {
    sourceReady: true,
    taskReady: true,
    assetCount: 3,
    pagesTotal: 0,
    promptReadyCount: 0,
    imageReadyCount: 0,
    busy: true
  };

  assert.deepEqual(privateComicPrimaryAction({ ...base, action: 'prompts' }), {
    key: 'prompts',
    label: '生成漫画提示词中…',
    disabled: true
  });
  assert.deepEqual(privateComicPrimaryAction({ ...base, pagesTotal: 3, promptReadyCount: 3, action: 'pages' }), {
    key: 'images',
    label: '生成漫画图片中…',
    disabled: true
  });
});
