import { generateComicPage } from '../../../shared/api/batchFactoryV12.js';
import { apiErrorMessage } from '../../../shared/apiErrors.js';
import { imageAspectRatioOrientation, isPortraitImageAspectRatio } from '../../../shared/imageAspectRatios.js';

function imageAspectRatio(settings = {}) {
  return String(settings.imageAspectRatio || settings.aspectRatio || '').trim() || '9:16';
}

export function comicPromptForImageAspect(settings = {}, prompt = '') {
  const aspectRatio = imageAspectRatio(settings);
  const orientation = imageAspectRatioOrientation(aspectRatio);
  let text = String(prompt || '')
    .replaceAll('1440×1920 或等比例 3:4', `与 ${aspectRatio} 画幅匹配的高清分辨率`)
    .replaceAll('3:4 竖版', `${aspectRatio}${orientation}`)
    .replaceAll('3:4竖版', `${aspectRatio}${orientation}`)
    .replaceAll('{{image_orientation}}', orientation)
    .replaceAll('{{image_aspect_ratio}}', aspectRatio)
    .replaceAll('3:4', aspectRatio)
    .trim();
  // Preset bodies are authored for portrait pages; landscape and square
  // canvases must not keep portrait-only wording.
  if (!isPortraitImageAspectRatio(aspectRatio)) {
    const landscape = aspectRatio !== '1:1';
    text = text
      .replaceAll('竖版长条', landscape ? '横幅' : '方形')
      .replaceAll('竖版', landscape ? '横幅' : '方形')
      .replaceAll('竖幅', landscape ? '横幅' : '方形')
      .replaceAll('竖屏', landscape ? '横屏' : '方形')
      .replaceAll('条漫', '漫画');
  }
  return text.trim();
}

export function comicPageGenerationPayload(settings = {}, prompt = '', imagePrompt = '') {
  const normalizedPrompt = comicPromptForImageAspect(settings, prompt);
  const normalizedImagePrompt = comicPromptForImageAspect(settings, imagePrompt || prompt);
  return {
    comicPrompt: normalizedPrompt,
    imagePrompt: normalizedImagePrompt,
    provider: String(settings.imageProvider || '').trim(),
    modelId: String(settings.imageModelId || '').trim(),
    aspectRatio: imageAspectRatio(settings)
  };
}

// 出图请求的失败分类：409 是“该页已有在途生成”，不重试也不报错；
// 401/403 是登录问题，交给全局登录失效提示；5xx 与网络中断可有限重试；
// 404 与其它 4xx 是确定性失败，重试没有意义。
export function comicPageErrorKind(error) {
  const status = Number(error?.status) || 0;
  if (status === 409) return 'conflict';
  if (status === 401 || status === 403) return 'auth';
  if (status === 404) return 'not_found';
  if (status === 0 || status >= 500) return 'retryable';
  return 'client';
}

function userFacingDetail(value, maxLength = 80) {
  return String(value || '')
    .replace(/https?:\/\/\S+/gi, '')
    .replace(/\/api\/\S+/gi, '')
    .replace(/\b(?:请求失败|HTTP)\s*[:：]?\s*[45]\d\d\b/gi, '')
    .replace(/[（(]?\b[45]\d\d\b[)）]?/g, '')
    .replace(/\bjob[-_ ]?id\b[：:]?\s*\S+/gi, '')
    .replace(/\s+/g, ' ')
    .trim()
    .slice(0, maxLength);
}

// 统一的用户可懂文案：{动作}失败：{原因}。{下一步}。
// 不暴露接口路径、状态码与任务 ID。
export function comicPageErrorMessage(error, action = '出图') {
  const kind = comicPageErrorKind(error);
  switch (kind) {
    case 'conflict':
      return '这一页正在生成中，请稍候…';
    case 'auth':
      return `${action}失败：登录状态已失效。请重新登录后再试。`;
    case 'not_found':
      return `${action}失败：这一页已不存在。请刷新后重试。`;
    case 'retryable': {
      const reason = userFacingDetail(apiErrorMessage(error, '')).replace(/[。！？!?]+$/, '');
      if (!reason) return `${action}失败：图片服务暂时不可用。请稍后重试。`;
      const nextStep = /请稍后|稍后再试|重试/.test(reason) ? '' : '请稍后重试。';
      return `${action}失败：${reason}。${nextStep}`;
    }
    default: {
      const reason = userFacingDetail(apiErrorMessage(error, '')).replace(/[。！？!?]+$/, '');
      return reason
        ? `${action}失败：${reason}。请调整提示词后重试。`
        : `${action}失败：请调整提示词后重试。`;
    }
  }
}

// 出图统一入口：409 返回结构化“在途”结果，由调用方合并成一条提示并启动轮询；
// 5xx / 网络中断默认一次重试；401/403、404 与其它 4xx 不重试。
export async function generateComicPageWithConflict(batchId, bookId, videoId, payload, options = {}) {
  const request = options.request || generateComicPage;
  const retries = Number.isFinite(Number(options.retries)) ? Math.max(0, Number(options.retries)) : 1;
  const retryDelay = Number.isFinite(Number(options.retryDelay)) ? Math.max(0, Number(options.retryDelay)) : 800;
  const sleep = options.sleep || (ms => new Promise(resolve => setTimeout(resolve, ms)));
  let attempt = 0;
  while (true) {
    try {
      const response = await request(batchId, bookId, videoId, payload);
      return { ok: true, conflict: false, response };
    } catch (error) {
      const kind = comicPageErrorKind(error);
      if (kind === 'conflict') return { ok: false, conflict: true, error };
      // 登录失效由 client.js 的 auth-expired 事件统一提示，这里只把原始错误抛回给调用方。
      if (kind === 'auth') throw error;
      if (kind === 'retryable' && attempt < retries) {
        attempt += 1;
        await sleep(retryDelay);
        continue;
      }
      const failure = new Error(comicPageErrorMessage(error));
      failure.status = Number(error?.status) || 0;
      failure.code = String(error?.code || '');
      failure.cause = error;
      throw failure;
    }
  }
}

// 并发池：只同时跑 limit 个任务；signal 中断后不再派发新任务，已在途任务会自然结束。
// 返回结构让调用方区分“完成 / 失败 / 被停止”，失败页保持可重试。
export async function runComicPagePool(items, { limit = 4, worker, signal, onProgress } = {}) {
  const list = Array.isArray(items) ? items : [];
  const concurrency = Math.max(1, Number(limit) || 4);
  const completed = [];
  const failed = [];
  let stopped = Boolean(signal?.aborted);
  let cursor = 0;
  const report = () => onProgress?.({ completed: completed.length, failed: failed.length, total: list.length, stopped });
  if (stopped || !list.length) {
    report();
    return { completed, failed, stopped };
  }
  const runWorker = async () => {
    while (true) {
      if (signal?.aborted) {
        stopped = true;
        return;
      }
      const index = cursor;
      cursor += 1;
      if (index >= list.length) return;
      const item = list[index];
      try {
        const outcome = await worker(item, index, signal);
        if (outcome?.stopped) {
          stopped = true;
          return;
        }
        completed.push(item);
      } catch (error) {
        if (signal?.aborted) {
          stopped = true;
          return;
        }
        failed.push({ item, index, error });
      }
      report();
    }
  };
  await Promise.all(Array.from({ length: Math.min(concurrency, list.length) }, runWorker));
  if (signal?.aborted) stopped = true;
  report();
  return { completed, failed, stopped };
}
