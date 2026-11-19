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
  assert.doesNotMatch(source, /title:\s*'首次完整生成'/);
  assert.doesNotMatch(source, /batch\.comic-bootstrap/);
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
  assert.match(source, /batchFactoryMode === 'comic'[\s\S]*flattenBatchFactoryComicSections\(batchFactorySections\)/);
  assert.match(source, /section\.key !== 'batch-comic'/);
});

test('batch factory comic mode promotes each comic prompt class to its own collapse row', () => {
  assert.match(source, /function flattenBatchFactoryComicSections\(/);
  assert.match(source, /key: `\$\{section\.key\}:\$\{sub\.title\}`/);
  assert.match(source, /return flattenBatchFactoryComicSections\(batchFactorySections\);/);
});
test('shows the comic reference scene library in the batch factory comic preset section', () => {
  assert.match(source, /'batch\.comic-reference-scene'/);
  assert.match(source, /场景参考图库/);
	assert.match(source, /ComicMaterialPreview/);
});

test('comic library previews load through the signed API instead of a bare image request', () => {
  assert.match(source, /loadAdminComicMaterialPreview/);
  assert.match(source, /function ComicMaterialPreview\(/);
  assert.match(source, /URL\.createObjectURL\(blob\)/);
});

test('always explains where a draft can add a comic library image', () => {
  assert.match(source, /图库图片仅用于“漫画画风库”和“漫画场景参考图库”/);
  assert.match(source, /请先选择归属，再上传图片/);
});

test('each comic library section can create a preset with its slot preselected', () => {
  assert.match(source, /title: '画风库', createSlot: 'batch\.comic-style'/);
  assert.match(source, /title: '场景参考图库', createSlot: 'batch\.comic-reference-scene'/);
  assert.match(source, /sectionHeader\(sub\)/);
  assert.match(source, /form\.setFieldsValue\(\{ \.\.\.emptyDraft\(module\), \.\.\.presetSlot/);
});

test('puts a plus row below each slot-bound prompt table', () => {
  assert.match(source, /function presetTable\(items, slotID = ''\)/);
  assert.match(source, /border: '1px dashed #d9d9d9'/);
  assert.match(source, /icon=\{<PlusOutlined \/>\}/);
  assert.match(source, /onClick=\{\(\) => openCreate\(slotID\)\}/);
});

test('uses slot-bound plus rows except for fixed runtime prompt registries', () => {
  assert.match(source, /const librarySections = useMemo/);
  assert.match(source, /createSlot: cmModule \|\| novelModule \? '' : slot\.id/);
  assert.match(source, /module !== 'batch-factory'/);
  assert.match(source, /children: presetTable\(section\.items, section\.createSlot\)/);
});

test('hides the internal preset ID column only for the CM plugin module', () => {
  assert.match(source, /module === 'luoshui-plugin' \? \[\] : \[\{ title: '预设词 ID'/);
});

test('puts comic child-section plus rows below their own prompt tables', () => {
  assert.match(source, /presetTable\(subItems, sub\.createSlot\)/);
});

test('keeps comic constraint ownership category-based when edited from batch factory', () => {
  assert.match(source, /const presetModule = editingPreset\?\.module \|\| module/);
  assert.match(source, /module: presetModule/);
  assert.match(source, /漫画约束使用下面的约束类别作为归属/);
  assert.match(source, /请选择归属，或选择约束类别/);
  assert.match(source, /delete nextProtocolLock\.slot/);
  assert.match(source, /漫画约束 · \$\{comicConstraintLabels/);
});

test('shows an existing draft as the main row instead of hiding it behind the published version', () => {
  assert.match(source, /const draft = sorted\.find\(version => version\.status === 'draft'\)/);
  assert.match(source, /const current = draft \|\| published \|\| sorted\[0\]/);
});

test('allows deleting any preset version, including published versions', () => {
  assert.match(source, /deletePreset/);
  assert.match(source, /历史项目不受影响/);
});

test('maps underscored constraint categories to the existing hyphenated preset IDs', () => {
  assert.match(source, /function constraintIDPrefix\(category\)/);
  assert.match(source, /replace\(\/_\/g, '-'\)/);
  assert.match(source, /const validConstraintID = values\.id === prefix \|\| values\.id\.startsWith\(`\$\{prefix\}-`\)/);
  assert.match(source, /必须是 \$\{prefix\} 或以 \$\{prefix\}- 开头/);
  assert.match(source, /const prefix = constraintIDPrefix\(values\.constraintCategory\)/);
});
