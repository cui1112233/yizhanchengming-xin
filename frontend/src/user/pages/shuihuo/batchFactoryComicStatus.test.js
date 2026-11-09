import assert from 'node:assert/strict';
import test from 'node:test';
import { comicImageLoadPlan, comicInlinePresentation, nextComicOutputVisibleCount, resolveComicImageSources } from './batchFactoryComicStatus.js';

test('已选主图时优先显示图片，不被历史失败状态覆盖', () => {
  const result = comicInlinePresentation({
    page: { generationState: 'failed', imageStatus: 'primary', images: [{ id: 'i1', isPrimary: true }] },
    selectedVideo: { comicPrompt: 'prompt' },
    loading: false,
    error: '读取失败',
    hasPrimary: true
  });
  assert.equal(result.stateText, '已选主图');
  assert.equal(result.placeholder, '');
});

test('生成中不再回落成“暂无漫画页”', () => {
  const result = comicInlinePresentation({
    page: { generationState: 'generating', promptStatus: 'ready', imageStatus: 'none', images: [] },
    selectedVideo: { comicPrompt: 'prompt' },
    loading: false,
    error: '',
    hasPrimary: false
  });
  assert.equal(result.statusText, '生成中');
  assert.equal(result.placeholder, '漫画页生成中…');
});

test('生成失败与尚未创建可区分', () => {
  const failed = comicInlinePresentation({
    page: { generationState: 'failed', promptStatus: 'ready', imageStatus: 'none', images: [] },
    selectedVideo: { comicPrompt: 'prompt' },
    loading: false,
    error: '',
    hasPrimary: false
  });
  assert.equal(failed.placeholder, '漫画页生成失败，点击查看');
  const missingPage = comicInlinePresentation({
    page: null,
    selectedVideo: { comicPrompt: 'prompt' },
    loading: false,
    error: '',
    hasPrimary: false
  });
  assert.equal(missingPage.placeholder, '漫画页尚未生成，点击打开媒体库');
  const noPrompt = comicInlinePresentation({
    page: null,
    selectedVideo: {},
    loading: false,
    error: '',
    hasPrimary: false
  });
  assert.equal(noPrompt.placeholder, '尚未提取漫画提示词');
});

test('请求失败只显示读取失败并允许重试', () => {
  const result = comicInlinePresentation({
    page: null,
    selectedVideo: { comicPrompt: 'prompt' },
    loading: false,
    error: 'network',
    hasPrimary: false
  });
  assert.equal(result.retry, true);
  assert.equal(result.placeholder, '漫画页读取失败，点击重试');
});

test('列表行优先取缩略图，缺缩略图时回落原图', () => {
  const withThumbnail = resolveComicImageSources({ url: '/content', thumbnailUrl: '/thumbnail' });
  assert.equal(withThumbnail.thumbnail, '/thumbnail');
  assert.equal(withThumbnail.original, '/content');
  const withoutThumbnail = resolveComicImageSources({ url: '/content' });
  assert.equal(withoutThumbnail.thumbnail, '/content');
  assert.equal(withoutThumbnail.original, '/content');
  const empty = resolveComicImageSources(null);
  assert.deepEqual(empty, { thumbnail: '', original: '' });
});

test('详情预览先显示缩略图再后台升级原图', () => {
  assert.deepEqual(
    comicImageLoadPlan({ url: '/content', thumbnailUrl: '/thumbnail' }, true),
    ['/thumbnail', '/content']
  );
  assert.deepEqual(
    comicImageLoadPlan({ url: '/content', thumbnailUrl: '/thumbnail' }, false),
    ['/thumbnail']
  );
  assert.deepEqual(comicImageLoadPlan({ url: '/content' }, true), ['/content']);
});

test('漫画输出卡首屏只挂 4 张，滚动按批追加并封顶总页数', () => {
  // 35 页的书首屏只渲染 4 张，避免一次拉 35 张图。
  assert.equal(nextComicOutputVisibleCount(0, 35), 4);
  assert.equal(nextComicOutputVisibleCount(4, 35), 8);
  assert.equal(nextComicOutputVisibleCount(32, 35), 35);
  assert.equal(nextComicOutputVisibleCount(0, 3), 3);
  assert.equal(nextComicOutputVisibleCount(0, 0), 0);
  assert.equal(nextComicOutputVisibleCount(8, 3), 3);
});
