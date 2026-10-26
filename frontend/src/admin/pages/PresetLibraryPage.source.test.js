import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';

const source = readFileSync(new URL('./PresetLibraryPage.jsx', import.meta.url), 'utf8');

test('batch factory provides video and comic generation as second-level choices', () => {
  assert.match(source, /const batchFactoryGenerationModes = \[/);
  assert.match(source, /label:\s*'视频生成',\s*value:\s*'video'/);
  assert.match(source, /label:\s*'漫画生成',\s*value:\s*'comic'/);
  assert.match(source, /const \[batchFactoryMode, setBatchFactoryMode\] = useState\('video'\)/);
  assert.match(source, /<Segmented options=\{batchFactoryGenerationModes\} value=\{batchFactoryMode\} onChange=\{setBatchFactoryMode\} \/>/);
});

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

test('script preset library separates comic image constraints from shared visual rules', () => {
  assert.match(source, /key: 'comic-constraints'/);
  assert.match(source, /title: '漫画图片约束'/);
  assert.match(source, /只用于漫画图片生成/);
});

test('batch factory comic mode does not show video-only H3 and SD sections', () => {
  assert.match(source, /batchFactoryMode === 'comic'[\s\S]*new Set\(\['batch-comic'\]\)/);
});
