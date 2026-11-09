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
  vm.runInContext(`${inlineScript}\n;globalThis.__prototypeTest = { state, navForMode, renderSidebar, renderBrandPage, renderPluginCenter, renderAgentFoundation, renderDuetFoundation };`, context);
  return context.__prototypeTest;
}

test('一战晟铭 keeps its own product navigation', () => {
  const prototype = bootPrototype();
  prototype.state.mode = 'brand';
  const labels = prototype.navForMode().map((item) => item[2]);

  for (const expected of ['首页', '私人定制', '小说获取', '水货生产', '配音', 'AI-CUT', '教程', '插件中心', '我的资产']) {
    assert.ok(labels.includes(expected), `missing product choice: ${expected}`);
  }
  for (const borrowed of ['小说创作', '角色库', '音色库', '作品', '漫剧解说', '剧本创作', 'momo画布']) {
    assert.ok(!labels.includes(borrowed), `borrowed top-level choice remains: ${borrowed}`);
  }
});

test('水货生产 contains the 创作漫剧 entry without renaming the product navigation', () => {
  const prototype = bootPrototype();
  prototype.state.mode = 'brand';
  prototype.state.pages.brand = 'production';
  const page = prototype.renderBrandPage();

  assert.match(page, /水货生产/);
  assert.match(page, /创作漫剧/);
  assert.match(page, /CREATE_COMIC_WORK/);
  assert.doesNotMatch(page, /漫剧解说/);
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

test('插件中心 renders the complete local 29-item catalog', () => {
  const prototype = bootPrototype();
  prototype.state.mode = 'brand';
  prototype.state.pages.brand = 'plugins';
  const pluginNav = prototype.navForMode().find((item) => item[2] === '插件中心');
  const page = prototype.renderPluginCenter();
  const expectedNames = [
    '资产参考', 'AI-分段', 'AI-Subtitle', 'AI-Filler', 'AI-双语字幕',
    'AI-CUT插件版', 'AI-BGM', 'AI-Shot', 'AI-Upscale', 'AI-Split',
    '小说转剧', 'AI改文', '分镜大师', '爆款开头', '女频性转',
    '字幕对轴', '格式化文本', '角色修改', '长篇拆书', '声音克隆',
    '世界观', '字幕合并', '文档合并', '书名生成', '封面生成',
    '角色三视图', '音色设计', '草稿合并', '剪映关键帧'
  ];

  assert.match(page, /插件市场/);
  assert.match(page, /搜索插件名称或功能/);
  assert.equal(pluginNav?.[3], '29');
  assert.equal([...page.matchAll(/data-plugin-id=/g)].length, 29);
  for (const name of expectedNames) assert.match(page, new RegExp(name));
  for (const placeholder of ['剪映草稿助手', '字幕纠错插件', '视频超分插件', '素材上传助手']) {
    assert.doesNotMatch(page, new RegExp(placeholder));
  }
});
