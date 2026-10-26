import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';

const source = readFileSync(new URL('./PresetLibraryPage.jsx', import.meta.url), 'utf8');

test('batch factory groups the complete comic prompt chain under manga generation', () => {
  assert.match(source, /title:\s*'漫画生成'/);
  assert.match(source, /title:\s*'资产推导'/);
  assert.match(source, /title:\s*'漫画分镜与页面'/);
  assert.match(source, /title:\s*'图片描述词'/);
  assert.match(source, /title:\s*'画风库'/);
  assert.match(source, /title:\s*'漫画约束'/);
  assert.match(source, /batch\.comic\.assets\.character/);
  assert.match(source, /batch\.comic\.assets\.scene/);
  assert.match(source, /batch\.comic\.assets\.prop/);
  assert.match(source, /batch\.comic\.character-image-template/);

  const mangaSection = source.indexOf("title: '漫画生成'");
  const genericConstraints = source.indexOf("title: '约束设置'", mangaSection);
  assert.ok(mangaSection >= 0 && genericConstraints > mangaSection, '漫画生成必须先接收漫画预设，避免约束提示词重复展示');
});

test('generic constraints exclude the three comic-only constraint categories', () => {
  assert.match(source, /comicConstraintCategories/);
  assert.match(source, /!comicConstraintCategories\.has\(preset\.protocolLock\?\.category\)/);
});
