import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { createHash } from 'node:crypto';
import { originalComicIcons } from './comic-screenplay-icon-data.js';
import { characterDesignInitialSelection, characterDesignOptionValue, toggleCharacterDesignOption, characterDesignSelectionText, applyCharacterDesign } from './comic-character-design.js';

// Minimal fixtures test the observed component rules, not recovered catalog values.
const catalog = [{ key: 'basic', label: 'A', fields: [{ key: 'single', label: 'S', options: ['one', 'two'] }, { key: 'multi', label: 'M', multiple: true, options: ['red', 'blue'] }] }, { key: 'style', label: 'B', fields: [{ key: 'color', label: 'C', type: 'color', options: [{ label: '蓝色', value: '#0000ff' }] }, { key: 'other', label: 'O', options: ['one'] }] }];
test('shipped catalog exactly preserves the recovered six categories, 22 fields and 210 options', () => {
  const raw = readFileSync(new URL('./comic-character-design-catalog.json', import.meta.url));
  assert.equal(createHash('sha256').update(raw.toString().replace(/\r\n/g,'\n')).digest('hex'), '14f0af31f4b4dadad59a6c3667062193534f5d35a0ad87e1d819ddd0a2c72240');
  const recovered = JSON.parse(raw);
  assert.equal(recovered.source_sha256, '2e5c09685ff5bc10160480f963936b56bac38a57a9e9b6bd199492f9437df1e8');
  assert.deepEqual(recovered.catalog.map(item => item.key), ['basic', 'makeup', 'features', 'body', 'hair', 'style']);
  assert.deepEqual(recovered.catalog.map(item => item.label), ['基础', '妆容与穿搭', '五官', '体型', '发型', '风格']);
  assert.equal(recovered.catalog.flatMap(item => item.fields).length, 22);
  assert.equal(recovered.catalog.flatMap(item => item.fields).flatMap(field => field.options).length, 210);
  assert.deepEqual(characterDesignInitialSelection(recovered.catalog), recovered.default_selection);
  assert.ok(!raw.toString().includes('\ufffd'));
});
test('character design PersonOutline preserves the original H2 geometry and source range', () => {
  const icon = originalComicIcons.PersonOutline;
  assert.deepEqual(icon.original, { line: 29, column: 1438771, start: 1547209, end: 1547288 });
  assert.equal(icon.svg.attrs.viewBox, '0 0 512 512');
  assert.deepEqual(icon.svg.children.map(child => child.attrs.d), ['M344 144c-3.92 52.87-44 96-88 96s-84.15-43.12-88-96c-4-55 35-96 88-96s92 42 88 96z', 'M256 304c-87 0-175.3 48-191.64 138.6C62.39 453.52 68.57 464 80 464h352c11.44 0 17.62-10.48 15.65-21.4C431.3 352 343 304 256 304z']);
  assert.ok(icon.svg.children.every(child => child.attrs['stroke-width'] === '32'));
});
test('character design restores isolated selection, label colors, repeated single and multiple toggles', () => {
  const initial = { multi: ['red'], color: '蓝色' };
  const selection = characterDesignInitialSelection(catalog, initial);
  assert.deepEqual(selection, { single: '', multi: ['red'], color: '蓝色', other: '' });
  assert.equal(characterDesignOptionValue({ label: '蓝色', value: '#0000ff' }), '蓝色');
  const first = toggleCharacterDesignOption(selection, catalog[0].fields[0], 'one');
  assert.equal(first.single, 'one');
  assert.equal(toggleCharacterDesignOption(first, catalog[0].fields[0], 'one').single, '');
  const next = toggleCharacterDesignOption(selection, catalog[0].fields[1], 'blue');
  assert.deepEqual(next.multi, ['red', 'blue']);
  assert.deepEqual(toggleCharacterDesignOption(next, catalog[0].fields[1], 'red').multi, ['blue']);
  assert.deepEqual(initial.multi, ['red']);
});
test('selection text follows catalog field order, click order within arrays and global exact deduplication', () => {
  assert.equal(characterDesignSelectionText({ single: ' one ', multi: ['blue', 'red', 'blue'], color: '蓝色', other: 'one' }, catalog), 'one，blue，red，蓝色');
});
test('explicit design save stores full selection without duplicate append and preserves manual prompt edits', () => {
  const selection = { single: 'one', multi: [], color: '蓝色', other: '', preserved: 'extra' };
  const original = { prompt: '手工设定', extra: { alias: '别名' } };
  const saved = applyCharacterDesign(original, { text: 'one，蓝色', selection });
  assert.equal(saved.prompt, '手工设定，one，蓝色');
  assert.deepEqual(JSON.parse(saved.extra.character_design_selection), selection);
  assert.equal(saved.extra.alias, '别名');
  assert.equal(applyCharacterDesign(saved, { text: 'one，蓝色', selection }).prompt, saved.prompt);
  assert.equal(applyCharacterDesign(saved, { text: 'two', selection }).prompt, '手工设定，two');
  const manual = { ...saved, prompt: '手工设定，one，蓝色，用户手动补充' };
  assert.equal(applyCharacterDesign(manual, { text: 'two', selection }).prompt, '手工设定，one，蓝色，用户手动补充，two');
  assert.equal(applyCharacterDesign(manual, { text: 'one，蓝色', selection }).prompt, manual.prompt);
  assert.equal(applyCharacterDesign(saved, { text: '', selection: {} }).prompt, '手工设定');
  assert.equal(original.prompt, '手工设定');
});
test('empty/reset save removes only a provable old design suffix and always clears its anchor', () => {
  const original={prompt:'手写设定，女性，长发',extra:{character_design_text:'女性，长发',character_design_selection:'{"gender":"女性"}',alias:'别名'}};
  const reset=applyCharacterDesign(original,{text:'',selection:{gender:'',hair:''}});
  assert.equal(reset.prompt,'手写设定');assert.equal(reset.extra.character_design_text,undefined);assert.equal(reset.extra.alias,'别名');assert.deepEqual(JSON.parse(reset.extra.character_design_selection),{gender:'',hair:''});
  assert.equal(applyCharacterDesign({prompt:'女性，长发',extra:original.extra},{text:'',selection:{}}).prompt,'');
  const manuallyEdited={...original,prompt:'  手写设定，女性，长发，手工补充内容  '};
  const kept=applyCharacterDesign(manuallyEdited,{text:'',selection:{}});
  assert.equal(kept.prompt,manuallyEdited.prompt);assert.equal(kept.extra.character_design_text,undefined);
  const middle={...original,prompt:'女性，长发，手写结尾'};
  assert.equal(applyCharacterDesign(middle,{text:'',selection:{}}).prompt,middle.prompt);
  const unanchored={prompt:'纯手写内容',extra:{alias:'保持'}};
  assert.equal(applyCharacterDesign(unanchored,{text:'',selection:{}}).prompt,unanchored.prompt);
  assert.equal(applyCharacterDesign(reset,{text:'',selection:{}}).prompt,'手写设定');
});
