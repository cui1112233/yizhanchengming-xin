import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { comicPageGenerationPayload, comicPromptForImageAspect } from './batchFactoryComicMode.js';

test('comic page generation uses the current prompt, provider and existing image default ratio', () => {
  assert.deepEqual(comicPageGenerationPayload({ imageProvider: 'doubao_local_executor', imageModelId: 'seedream-5-flash-local' }, '完整漫画页提示词'), {
    comicPrompt: '完整漫画页提示词',
    imagePrompt: '完整漫画页提示词',
    provider: 'doubao_local_executor',
    modelId: 'seedream-5-flash-local',
    aspectRatio: '9:16'
  });
});

test('comic page payload keeps the director prompt and uses the derived image prompt for generation', () => {
  const payload = comicPageGenerationPayload(
    { imageAspectRatio: '9:16', imageProvider: 'doubao_local_executor', imageModelId: 'seedream' },
    '漫画导演说明',
    '图片模型执行提示词'
  );
  assert.equal(payload.comicPrompt, '漫画导演说明');
  assert.equal(payload.imagePrompt, '图片模型执行提示词');
});

test('comic page generation respects an explicit image ratio without inheriting video ratio', () => {
  assert.deepEqual(comicPageGenerationPayload({ imageModelId: 'comic-model', imageAspectRatio: '1:1', aspectRatio: '9:16' }, '页面'), {
    comicPrompt: '页面',
    imagePrompt: '页面',
    provider: '',
    modelId: 'comic-model',
    aspectRatio: '1:1'
  });
});

test('comic page generation supports the legacy shared image ratio field', () => {
  assert.equal(comicPageGenerationPayload({ aspectRatio: '16:9' }, '页面').aspectRatio, '16:9');
});

test('saved comic prompts follow the current image ratio instead of retaining legacy 3:4 text', () => {
  const prompt = comicPromptForImageAspect({ imageAspectRatio: '9:16' }, '一张完整的3:4竖版条漫页，1440×1920 或等比例 3:4');
  assert.equal(prompt, '一张完整的9:16画幅条漫页，与 9:16 画幅匹配的高清分辨率');
  assert.equal(comicPageGenerationPayload({ imageAspectRatio: '9:16' }, '一张3:4竖版条漫页').imagePrompt, '一张9:16画幅条漫页');
});

test('the inline storyboard card applies the current image ratio to saved comic prompts', () => {
  const source = readFileSync(new URL('./BatchFactoryNovelList.jsx', import.meta.url), 'utf8');
  assert.match(source, /const comicPrompt = comicPromptForImageAspect\(effectiveBookSettings\(batch, book\), video\?\.comicPrompt \|\| ''\)/);
});
