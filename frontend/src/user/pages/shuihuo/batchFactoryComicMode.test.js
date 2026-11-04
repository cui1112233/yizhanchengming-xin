import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { comicPageErrorMessage, comicPageErrorKind, comicPageGenerationPayload, comicPromptForImageAspect, generateComicPageWithConflict, runComicPagePool } from './batchFactoryComicMode.js';

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
  assert.equal(prompt, '一张完整的9:16竖幅构图条漫页，与 9:16 画幅匹配的高清分辨率');
  assert.equal(comicPageGenerationPayload({ imageAspectRatio: '9:16' }, '一张3:4竖版条漫页').imagePrompt, '一张9:16竖幅构图条漫页');
});

test('landscape and square ratios do not keep portrait-only wording', () => {
  assert.equal(comicPromptForImageAspect({ imageAspectRatio: '4:3' }, '一张3:4竖版条漫页'), '一张4:3横幅构图漫画页');
  assert.equal(comicPromptForImageAspect({ imageAspectRatio: '3:2' }, '竖版构图，条漫分镜'), '横幅构图，漫画分镜');
  assert.equal(comicPromptForImageAspect({ imageAspectRatio: '1:1' }, '一张3:4竖版条漫页'), '一张1:1方形构图漫画页');
});

test('the inline storyboard card applies the current image ratio to saved comic prompts', () => {
  const source = readFileSync(new URL('./BatchFactoryNovelList.jsx', import.meta.url), 'utf8');
  assert.match(source, /const comicPrompt = comicPromptForImageAspect\(effectiveBookSettings\(batch, book\), video\?\.comicPrompt \|\| ''\)/);
});

function httpError(status, message = '') {
  const error = new Error(message || `HTTP ${status}`);
  error.status = status;
  return error;
}

test('409 conflict means “already generating”: one call, no retry, structured in-flight result', async () => {
  let calls = 0;
  const request = async () => { calls += 1; throw httpError(409, 'page is in flight'); };
  const outcome = await generateComicPageWithConflict('b', 'k', 'v', {}, { request, retries: 1, retryDelay: 0, sleep: async () => {} });
  assert.equal(calls, 1);
  assert.equal(outcome.conflict, true);
  assert.equal(outcome.ok, false);
  assert.equal(outcome.error.status, 409);
});

test('401/403 auth failure is not retried so the global login prompt owns it', async () => {
  let calls = 0;
  const request = async () => { calls += 1; throw httpError(401, '登录已失效'); };
  await assert.rejects(async () => generateComicPageWithConflict('b', 'k', 'v', {}, { request, retries: 3, retryDelay: 0, sleep: async () => {} }), /登录已失效/);
  assert.equal(calls, 1);
});

test('5xx is retried once and succeeds on the second attempt', async () => {
  let calls = 0;
  const request = async () => {
    calls += 1;
    if (calls === 1) throw httpError(503, '服务暂不可用');
    return { pages: [] };
  };
  const outcome = await generateComicPageWithConflict('b', 'k', 'v', {}, { request, retries: 1, retryDelay: 0, sleep: async () => {} });
  assert.equal(calls, 2);
  assert.equal(outcome.ok, true);
  assert.equal(outcome.conflict, false);
});

test('404 is not retried and becomes a friendly message without status codes or paths', async () => {
  let calls = 0;
  const request = async () => { calls += 1; throw httpError(404, '漫画页不存在 GET /api/batch/1/comic-pages'); };
  await assert.rejects(
    async () => generateComicPageWithConflict('b', 'k', 'v', {}, { request, retries: 2, retryDelay: 0, sleep: async () => {} }),
    error => {
      assert.match(error.message, /^出图失败：这一页已不存在。请刷新后重试。$/);
      assert.doesNotMatch(error.message, /\/api|\b404\b/);
      return true;
    }
  );
  assert.equal(calls, 1);
});

test('concurrency pool never exceeds its limit and finishes every page', async () => {
  let active = 0;
  let maxActive = 0;
  const items = [1, 2, 3, 4, 5, 6];
  const result = await runComicPagePool(items, {
    limit: 3,
    worker: async () => {
      active += 1;
      maxActive = Math.max(maxActive, active);
      await new Promise(resolve => setTimeout(resolve, 5));
      active -= 1;
    }
  });
  assert.equal(maxActive, 3);
  assert.equal(result.completed.length, items.length);
  assert.equal(result.failed.length, 0);
  assert.equal(result.stopped, false);
});

test('stopping the pool aborts remaining dispatches and leaves pages retryable', async () => {
  const controller = new AbortController();
  const processed = [];
  const items = [1, 2, 3, 4, 5, 6];
  const result = await runComicPagePool(items, {
    limit: 2,
    signal: controller.signal,
    worker: async item => {
      processed.push(item);
      if (processed.length === 2) controller.abort();
    }
  });
  assert.equal(result.stopped, true);
  assert.ok(result.completed.length <= 3);
  assert.equal(result.failed.length, 0);
  assert.ok(result.completed.length + result.failed.length < items.length);
});

test('error helpers classify status codes consistently', () => {
  assert.equal(comicPageErrorKind({ status: 409 }), 'conflict');
  assert.equal(comicPageErrorKind({ status: 401 }), 'auth');
  assert.equal(comicPageErrorKind({ status: 404 }), 'not_found');
  assert.equal(comicPageErrorKind({ status: 503 }), 'retryable');
  assert.equal(comicPageErrorKind(new TypeError('Failed to fetch')), 'retryable');
  assert.equal(comicPageErrorMessage({ status: 409 }), '这一页正在生成中，请稍候…');
});
