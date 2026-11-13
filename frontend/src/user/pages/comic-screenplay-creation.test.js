import assert from 'node:assert/strict';
import test from 'node:test';
import { comicDraftPayload, createSourceImportGate } from './comic-screenplay-creation.js';

const form = { title: '新故事', model: 'text-model', promptId: 'b_director_screenplay', scriptMode: 'b', artStyle: '水墨', selectedStyle: 'style-ink', degree: 'master', aspectRatio: '9:16' };

test('直接输入创建，一次提交原文和所选画风标识，不提交前端提示词', () => {
  const payload = comicDraftPayload(form, '\n窗外是风景。\n下一幕由你创造。\n');
  assert.equal(payload.sourceText, '\n窗外是风景。\n下一幕由你创造。\n');
  assert.equal(payload.selectedStyle, 'style-ink');
  assert.equal(payload.artStyle, '水墨');
  assert.equal(payload.assetDegree, 'master');
  assert.equal(payload.styleDesc, undefined);
});

test('空白或过长原文在创建前被拦截', () => {
  assert.throws(() => comicDraftPayload(form, ' \n\t'), /输入或导入/);
  assert.throws(() => comicDraftPayload(form, '文'.repeat(600001)), /60 万/);
  assert.throws(() => comicDraftPayload({ ...form, title: ' ' }, '正文'), /作品名称/);
  assert.throws(() => comicDraftPayload({ ...form, artStyle: '', selectedStyle: '' }, '正文'), /画风/);
});

test('两次导入乱序完成，仅最新文件可以更新草稿', async () => {
  const resolves = new Map();
  const gate = createSourceImportGate(file => new Promise(resolve => resolves.set(file, resolve)));
  const oldImport = gate.load('旧文件');
  const newImport = gate.load('新文件');
  resolves.get('新文件')('新内容');
  assert.deepEqual(await newImport, { current: true, content: '新内容' });
  resolves.get('旧文件')('旧内容');
  assert.equal((await oldImport).current, false);
});

test('输入、取消或关闭弹窗后，迟到的解析结果和错误都不覆盖草稿', async () => {
  let finish;
  const gate = createSourceImportGate(() => new Promise((resolve, reject) => { finish = reject; }));
  const pending = gate.load('file');
  gate.cancel();
  finish(new Error('已过时的错误'));
  assert.equal((await pending).current, false);
});
