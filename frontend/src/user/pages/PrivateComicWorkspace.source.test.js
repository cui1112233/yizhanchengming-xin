import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'PrivateComicWorkspace.jsx'), 'utf8');
const pickerSource = fs.readFileSync(path.join(here, 'shuihuo', 'BatchFactoryComicStylePicker.jsx'), 'utf8');
const constraintsSource = fs.readFileSync(path.join(here, 'private-studio', 'PrivateStudioConstraintsModal.jsx'), 'utf8');

test('private comic workspace accepts pasted text and exposes the comic actions', () => {
  assert.match(source, /name="sourceText"/);
  assert.match(source, /name="title"/);
  assert.match(source, /开始生成漫画/);
  assert.match(source, /刷新状态/);
  assert.match(source, /查看漫画详情/);
  assert.match(source, /不自动发布/);
});

test('private comic workspace reuses the published style library and comic prompt catalog', () => {
  assert.match(source, /BatchFactoryComicStylePicker/);
  assert.match(source, /onReferenceDataUrl=\{onPendingReferenceChange\}/);
  assert.match(source, /batch\.comic-page/);
  assert.match(source, /batch\.comic-image-derivation/);
  assert.match(source, /PRIVATE_STUDIO_CONSTRAINT_FIELDS\.comic\.map/);
  assert.match(constraintsSource, /key: 'comic_quality'/);
  assert.match(constraintsSource, /key: 'comic_restriction'/);
  assert.match(constraintsSource, /key: 'comic_negative'/);
  assert.match(source, /private-customization-reference-note/);
});

test('comic constraints are persisted to the server task with a readback check', () => {
  assert.match(source, /onPersistSettings/);
  assert.match(source, /async function saveSettings\(\)/);
  assert.match(source, /await onPersistSettings\?\.\(\)/);
  assert.match(source, /message\.error\(error\?\.message \|\| '漫画约束保存失败'\)/);
  assert.match(source, /onOk=\{\(\) => void saveSettings\(\)\}/);
  assert.match(source, /okButtonProps=\{\{ loading: savingSettings \}\}/);
  assert.match(source, /okText=\{task\?\.batchId \? '保存到任务' : '完成'\}/);
});

test('private comic workspace blocks submit without pasted source', () => {
  assert.match(source, /const sourceReady = sourceLength >= privateComicMinSourceLength/);
  assert.match(source, /disabled=\{busy \|\| !sourceReady\}/);
});

test('changing the style clears a pre-submit custom reference', () => {
  assert.match(pickerSource, /onReferenceDataUrl\?\.\('\'\)/);
});

test('clearing pasted comic text asks for confirmation first', () => {
  assert.match(source, /<Popconfirm/);
  assert.match(source, /清空已粘贴的 \$\{sourceLength\} 字正文？/);
  assert.match(source, /该操作不可撤销。/);
  assert.match(source, /onConfirm=\{\(\) => patchDraft\(\{ sourceText: '' \}\)\}/);
});
