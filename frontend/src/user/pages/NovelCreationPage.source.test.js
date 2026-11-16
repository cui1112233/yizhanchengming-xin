import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';
import { MODES, draftKey, mergeSnapshot, modelOptions } from './novel-creation/contracts.js';

test('小说创作拥有独立导航与路由', () => {
  const layout = readFileSync(new URL('../../shared/layouts/UserLayout.jsx', import.meta.url), 'utf8');
  const app = readFileSync(new URL('../App.jsx', import.meta.url), 'utf8');
  assert.match(layout, /href: '\/novel-creation'[\s\S]{0,80}label: '小说创作'/);
  assert.match(app, /'\/novel-creation': NovelCreationPage/);
});

test('六模式使用批准的中文合同', () => {
  assert.deepEqual(MODES.map(([, label]) => label), ['长篇原创', '短篇原创', '仿写长篇', '仿写短篇', '番外创作', '续写创作']);
});

test('草稿键编码分隔符且账号作品会话模式分别隔离', () => {
  const parts = ['owner', 'project', 'session', 'mode'];
  const baseline = draftKey(...parts);
  for (let i = 0; i < parts.length; i++) { const changed = [...parts]; changed[i] += ':other'; assert.notEqual(draftKey(...changed), baseline); }
  assert.notEqual(draftKey('a:b', 'c', 's', 'm'), draftKey('a', 'b:c', 's', 'm'));
});

test('同run增量事件去重按游标合并，换run不混旧事件', () => {
  const old = { run: { id: 'one' }, events: [{ eventId: 2, message: 'old' }], chapters: [{ id: 'saved' }] };
  const result = mergeSnapshot(old, { run: { id: 'one', status: 'writing' }, events: [{ eventId: 3 }, { eventId: 2, message: 'new' }] });
  assert.deepEqual(result.events.map(e => e.eventId), [2, 3]); assert.equal(result.events[0].message, 'new'); assert.equal(result.chapters[0].id, 'saved');
  assert.deepEqual(mergeSnapshot(old, { run: { id: 'two' }, events: [] }).events, []);
});

test('不可用显式模型保留为禁选提示，授权模型不被改写', () => {
  const options = [{ value: 'paid', label: '授权模型' }]; const result = modelOptions(options, 'removed');
  assert.equal(result[0].value, 'removed'); assert.equal(result[0].disabled, true); assert.equal(result[1], options[0]);
});
