import test from 'node:test';
import assert from 'node:assert/strict';
import vm from 'node:vm';
import { readFile } from 'node:fs/promises';

const html = await readFile(new URL('./index.html', import.meta.url), 'utf8');
const inlineScript = html.match(/<script>([\s\S]*)<\/script>/)?.[1];
assert.ok(inlineScript, 'prototype must contain an inline script');

function bootPrototype() {
  const app = { innerHTML: '' };
  const toast = { textContent: '', classList: { add() {}, remove() {} } };
  const document = {
    body: { dataset: {} },
    documentElement: { dataset: {} },
    querySelector(selector) {
      if (selector === '#app') return app;
      if (selector === '#toast-region') return toast;
      return null;
    },
    addEventListener() {}
  };
  const saved = new Map();
  const context = vm.createContext({
    document,
    localStorage: {
      getItem(key) { return saved.get(key) ?? null; },
      setItem(key, value) { saved.set(key, String(value)); }
    },
    structuredClone,
    setTimeout() { return 1; },
    clearTimeout() {},
    setInterval() { return 1; },
    clearInterval() {},
    console
  });
  context.window = context;
  vm.runInContext(`${inlineScript}\n;globalThis.__prototypeTest = { state, navForMode, renderSidebar, renderBrandPage, renderAgentFoundation, renderDuetFoundation };`, context);
  return context.__prototypeTest;
}

test('一战晟铭 exposes the benchmarked creation choices instead of the old placeholder menu', () => {
  const prototype = bootPrototype();
  prototype.state.mode = 'brand';
  const labels = prototype.navForMode().map((item) => item[2]);

  for (const expected of ['首页', '小说创作', '角色库', '音色库', '作品', '漫剧解说', '剧本创作', 'momo画布', 'AI-CUT', '教程', '插件中心', '我的资产']) {
    assert.ok(labels.includes(expected), `missing creation choice: ${expected}`);
  }
  for (const obsolete of ['私人定制', '小说获取', '水货生产', '配音']) {
    assert.ok(!labels.includes(obsolete), `obsolete placeholder remains: ${obsolete}`);
  }
});

test('Duet uses project channels and a conversation workspace instead of a flat admin menu', () => {
  const prototype = bootPrototype();
  prototype.state.mode = 'duet';
  prototype.state.pages.duet = 'channels';
  const sidebar = prototype.renderSidebar();
  const workspace = prototype.renderDuetFoundation();

  for (const expected of ['任何事', 'q', '剧本生产线', '资产工坊', '分镜台', '＋ 新频道', '＋ 新项目']) {
    assert.match(sidebar, new RegExp(expected));
  }
  for (const obsolete of ['接力看板', '定时任务']) {
    assert.doesNotMatch(sidebar, new RegExp(obsolete));
  }
  for (const expected of ['CM Duet', '员工会自动接力', '跟CM Duet说点什么', '当前频道任务']) {
    assert.match(workspace, new RegExp(expected));
  }
});

test('Agent proxy opens on its command composer and keeps the production workspace secondary', () => {
  const prototype = bootPrototype();
  prototype.state.mode = 'agent';
  prototype.state.pages.agent = 'new-session';
  const workspace = prototype.renderAgentFoundation();

  assert.match(workspace, /新 Agent 会话/);
  assert.match(workspace, /\/ 技能/);
  assert.match(workspace, /@ 引用/);
  assert.doesNotMatch(workspace, /镜头 1/);
});
