import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const readSource = relativePath => readFileSync(path.join(here, relativePath), 'utf8').replace(/\r\n/g, '\n');
const boundarySource = readSource('../batch-factory-v11/ProductionMediaBoundary.jsx');
const novelListSource = readSource('BatchFactoryNovelList.jsx');
const helpersSource = readSource('batchFactoryNovelListHelpers.js');
const mediaApiSource = readSource('../../../shared/api/batchFactoryV12.js');

test('合成成片播放改走短期签名地址而不是整片 blob 下载', () => {
  assert.match(mediaApiSource, /export function isMergeMediaURL\(/);
  assert.match(mediaApiSource, /export async function getMergeMediaPlaybackURL\(/);
  assert.match(mediaApiSource, /merge-media\/\$\{id\(parts\.artifactId\)\}\/token/);
  assert.match(mediaApiSource, /method: 'POST'/);

  // merge 分支必须先返回签名地址；只有非 merge 的旧受保护媒体才保留 blob 路径。
  assert.match(boundarySource, /mergeMedia\s*\n\s*\? getMergeMediaPlaybackURL\(source\)/);
  assert.match(boundarySource, /: getProductionMediaBlob\(source\)\.then\(blob =>/);
});

test('首帧加载有明确遮罩、超时兜底与可读失败文案', () => {
  assert.match(boundarySource, /export const MEDIA_FIRST_FRAME_TIMEOUT_MS = 15000;/);
  assert.match(boundarySource, /setTimeout\(\(\) => setTimedOut\(true\), MEDIA_FIRST_FRAME_TIMEOUT_MS\)/);
  assert.match(boundarySource, /视频首帧加载超时，请重试/);
  assert.match(boundarySource, /data-bf-media-state=\{timedOut \? 'timeout' : 'loading'\}/);
  assert.match(boundarySource, /视频加载失败：\{error\}/);
  // 首帧事件必须与调用方已有的 onLoadedData 串联，不能覆盖。
  assert.match(boundarySource, /const originalOnLoadedData = node\.props\.onLoadedData;/);
  assert.match(boundarySource, /onLoadedData: event => \{/);
});

test('列表主预览使用 preload=metadata，缩略图不再下载整段 mp4', () => {
  assert.doesNotMatch(novelListSource, /preload="auto"/);
  assert.match(novelListSource, /preload="metadata"/);
  assert.match(novelListSource, /isBatchFactoryMergeCoverURL\(posterUrl\)/);
  assert.match(novelListSource, /function MergeCoverThumb\(/);
  assert.match(novelListSource, /getBatchFactoryMergeCoverBlob\(url\)/);
  // 合成成片没有封面时只用占位，不回退到视频缩略图。
  assert.match(novelListSource, /: mergeMedia\n[\s\S]{0,120}?<PassiveMediaPoster/);
});

test('merge-cover poster 由 merge-media 地址推导，非 merge 不伪造封面', () => {
  assert.match(helpersSource, /export function batchFactoryIsMergeMedia\(/);
  assert.match(helpersSource, /if \(explicit\) return explicit;/);
  assert.ok(helpersSource.includes('merge-media\\/([^/?#]+)$/'));
  assert.match(helpersSource, /merge-cover\/\$\{match\[3\]\}/);
});

test('后端 merge-media 设置缓存验证器并支持 If-None-Match', async () => {
  const mergeSource = readFileSync(path.resolve(here, '../../../../../backend/internal/httpapi/batch_factory_v11_merge.go'), 'utf8');
  assert.match(mergeSource, /w\.Header\(\)\.Set\("ETag", mergeMediaETag\(job, fileInfo\)\)/);
  assert.match(mergeSource, /w\.Header\(\)\.Set\("Cache-Control", "private, max-age=3600"\)/);
  assert.match(mergeSource, /http\.ServeContent\(w, r, artifactID\+"\.mp4", modTime, seeker\)/);
});
