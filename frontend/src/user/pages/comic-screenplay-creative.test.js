import assert from 'node:assert/strict';
import test from 'node:test';
import { comicImageModelSelection, comicStoryboardEngine, createCreativeCreationFlow, creativeConfigDraft, creativeConfigPayload, creativePromptOptions } from './comic-screenplay-creative.js';
import { loadComicCreativeLibrary, loadComicCreativeConfig, saveComicCreativeConfig, getComicStylePreview, createComicCreativeStyle } from '../../shared/api/comicScreenplay.js';
const config = { script_prompt_id: 'script144', storyboard_prompt_id: 'story307', character_design_prompt_id: 'char213', scene_design_prompt_id: 'scene289', style: { id: 'own-style', name: '我的风格' }, image_core_instruction: '保留服装细节', image_model_id: 'image', image_edit_model_id: 'edit' };

test('恢复四步骤选择和私人画风，不用名称冒充 ID，不取目录首项', () => {
  const draft = creativeConfigDraft(config);
  assert.equal(draft.style_id, 'own-style');
  assert.equal(draft.style_name, '我的风格');
  assert.equal(draft.storyboard_prompt_id, 'story307');
  assert.equal(draft.character_design_prompt_id, 'char213');
  assert.equal(draft.scene_design_prompt_id, 'scene289');
  assert.equal(creativeConfigDraft({ style: { name: '历史名称' } }).style_id, null);
  assert.equal(creativeConfigDraft({}).script_prompt_id, '');
});

test('清空画风可以保存，独立 ID 和修订号不丢失；正文不发到浏览器保存请求', () => {
  const patch = creativeConfigPayload({ ...creativeConfigDraft(config), style_id: null }, 8);
  assert.equal(patch.style_id, null);
  assert.equal(patch.expected_revision, 8);
  assert.equal(patch.script_prompt_id, 'script144');
  assert.equal(patch.image_core_instruction, '保留服装细节');
  assert.equal(patch.style, undefined);
  assert.equal(patch.prompt, undefined);
  assert.throws(() => creativeConfigPayload({ ...creativeConfigDraft(config), scene_design_prompt_id: '' }, 8), /场景设计/);
});

test('按用途保留同名不同 ID，失效选择显式显示且不自动替换', () => {
  const prompts = [{ id: 'a', name: '相同名称', purpose: 'script', version: 'v1' }, { id: 'b', name: '相同名称', purpose: 'script' }, { id: 'c', name: '分镜', purpose: 'storyboard' }];
  const options = creativePromptOptions(prompts, 'script', 'old');
  assert.deepEqual(options.map(item => item.value), ['old', 'a', 'b']);
  assert.equal(options[0].disabled, true);
});

test('创作接口准确使用冻结路径、显式空画风和旧版本检查', async () => {
  const calls = [];
  const request = async (path, options) => { calls.push({ path, options }); return path.endsWith('creative-library') ? { styles: [{ id: 'mine' }], prompts: [] } : { creative_config: config, revision: 9 }; };
  assert.equal((await loadComicCreativeLibrary(request)).styles[0].id, 'mine');
  assert.equal((await loadComicCreativeConfig('p1', request)).revision, 9);
  await saveComicCreativeConfig('p1', { style_id: null, expected_revision: 8 }, request);
  assert.deepEqual(calls.map(item => item.path), ['/api/screenplays/creative-library', '/api/screenplays/p1/creative-config', '/api/screenplays/p1/creative-config']);
  assert.equal(calls[2].options.method, 'PUT');
  assert.deepEqual(JSON.parse(calls[2].options.body), { style_id: null, expected_revision: 8 });
  await assert.rejects(() => loadComicCreativeConfig('../x', request), /无效/);
});

test('私人画风预览使用授权读取，外部地址与批量工厂地址不接收登录信息', async () => {
  const request = async (path, options) => { assert.equal(options.responseType, 'blob'); return path; };
  assert.equal(await getComicStylePreview('/api/reference-assets/file/character/a/original', request), '/api/reference-assets/file/character/a/original');
  await getComicStylePreview('/api/agent/tasks/t/assets/a/content', request);
  await getComicStylePreview('/api/ai-cut/projects/private/assets/style/content', request);
  for (const path of ['https://example.com/a', '//example.com/a', '/api/presets/comic-style/a/preview', '/api/ai-cut/projects/p/assets/a/content?token=x', '/api/ai-cut/projects/p/assets/a/content/extra', 'C:/private/a.png']) await assert.rejects(() => getComicStylePreview(path, () => assert.fail('禁止请求')), /地址无效/);
});


test('创建部分失败保留同一项目，连续点击合并，修正选择后只补存配置', async () => {
  let creates = 0, saves = 0, createdID = '';
  const flow = createCreativeCreationFlow({ create: async () => { creates++; return { id: 'one' }; }, load: async () => ({ revision: saves + 4 }), save: async (id, patch) => { assert.equal(id, 'one'); assert.equal(patch.expected_revision, saves + 4); saves++; if (saves === 1) throw Error('保存失败'); assert.equal(patch.style_id, null); } });
  const first = flow.run({ title: '作品' }, { style_id: 'old' }, p => { createdID = p.id; });
  const duplicate = flow.run({ title: '重复点击' }, { style_id: 'old' });
  assert.equal(first, duplicate);
  await assert.rejects(first, /保存失败/);
  assert.equal(createdID, 'one');
  assert.equal(flow.getProject().id, 'one');
  assert.equal((await flow.run({ title: '作品' }, { style_id: null })).id, 'one');
  assert.equal(creates, 1); assert.equal(saves, 2);
});

test('空画风库可实际新建私人画风，前端不编造画风正文或图片', async () => {
  const created = await createComicCreativeStyle({ name: ' 水彩 ', prompt: ' 我自己的水彩要求 ' }, async (path, options) => {
    assert.equal(path, '/api/screenplays/styles'); assert.equal(options.method, 'POST');
    assert.deepEqual(JSON.parse(options.body), { name: '水彩', prompt: '我自己的水彩要求', images: [] });
    return { style: { id: 'own', name: '水彩' } };
  });
  assert.equal(created.id, 'own');
  await assert.rejects(() => createComicCreativeStyle({ name: ' ', prompt: '要求' }, () => assert.fail('不应请求')), /名称/);
});


test('媒体图片默认遵循项目B/C模型，参考图切换图片编辑配置，不能被目录首项A覆盖', () => {
  const models = ['A', 'B', 'C', 'D'].map(id => ({ id }));
  const config = { image_model_id: 'B', image_edit_model_id: 'C' };
  assert.equal(comicImageModelSelection(config, { models }).modelId, 'B');
  assert.equal(comicImageModelSelection(config, { models, references: ['ref'] }).modelId, 'C');
  assert.equal(comicImageModelSelection({ image_model_id: 'B' }, { models, references: ['ref'] }).modelId, 'B');
});

test('本次显式模型覆盖可恢复跟随项目，且不会修改项目配置', () => {
  const models = ['A', 'B', 'C', 'D'].map(id => ({ id })), config = { image_model_id: 'B', image_edit_model_id: 'C' };
  const temporary = comicImageModelSelection(config, { models, references: ['ref'], override: 'D' });
  assert.equal(temporary.modelId, 'D'); assert.equal(temporary.source, 'temporary');
  const restored = comicImageModelSelection(config, { models, references: ['ref'], override: '' });
  assert.equal(restored.modelId, 'C'); assert.equal(restored.source, 'project');
  assert.deepEqual(config, { image_model_id: 'B', image_edit_model_id: 'C' });
});

test('失效配置、空配置和目录加载失败均明确阻止媒体生成，不回退目录首项', () => {
  const models = [{ id: 'A' }];
  const invalid = comicImageModelSelection({ image_model_id: 'deleted' }, { models });
  assert.equal(invalid.modelId, 'deleted'); assert.match(invalid.error, /当前不可用/);
  assert.equal(invalid.options[0].disabled, true);
  const missing = comicImageModelSelection({}, { models });
  assert.equal(missing.modelId, ''); assert.match(missing.error, /尚未设置/);
  assert.equal(comicImageModelSelection({ image_model_id: 'A' }, { models, loadError: '读取失败' }).error, '读取失败');
});

test('分镜提示词切换V9到V7再V3时，引擎精确随所选提示词切换', () => {
  let engine = comicStoryboardEngine('luoshui_storyboard_v9_30s_307', 'xiaoyunque_v3');
  assert.equal(engine, 'xiaoluo_v9_30s');
  engine = comicStoryboardEngine('storyboard_xiaoluo_v7_15s', engine); assert.equal(engine, 'xiaoluo_v7_15s');
  engine = comicStoryboardEngine('storyboard_xiaoyunque_v3', engine); assert.equal(engine, 'xiaoyunque_v3');
  assert.equal(comicStoryboardEngine('historical-other', engine), 'xiaoyunque_v3');
});
