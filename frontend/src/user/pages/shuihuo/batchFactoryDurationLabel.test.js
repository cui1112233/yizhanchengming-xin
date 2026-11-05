import assert from 'node:assert/strict';
import test from 'node:test';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';
import { formatDurationSeconds } from './batchFactoryVideoMeta.js';

const here = dirname(fileURLToPath(import.meta.url));
const novelList = readFileSync(join(here, 'BatchFactoryNovelList.jsx'), 'utf8');

test('时长缺失或为 0 时显示「未知」，不编造数值', () => {
  for (const seconds of [0, '0.000', -1, null, undefined, NaN, '']) {
    assert.equal(formatDurationSeconds(seconds), '未知', `${String(seconds)} 不应被渲染成数值`);
  }
});

test('有真实时长时按一位小数显示', () => {
  assert.equal(formatDurationSeconds(9.088), '9.1s');
  assert.equal(formatDurationSeconds(7.104), '7.1s');
  assert.equal(formatDurationSeconds('9.088'), '9.1s');
  assert.equal(formatDurationSeconds(15), '15.0s');
  assert.equal(formatDurationSeconds(9.96), '10.0s');
});

test('媒体库主版本徽标：缺失/0 显示未知，有值显示一位小数', () => {
  assert.match(
    novelList,
    /<small>\{formatDurationSeconds\(main\.actualDurationSeconds\)\} · \{state\.list\.length\} 个版本</,
    '主版本徽标必须走 formatDurationSeconds，且不能再按 truthy 静默隐藏'
  );
  assert.doesNotMatch(
    novelList,
    /main\.actualDurationSeconds \? `\$\{Number\(main\.actualDurationSeconds\)\.toFixed\(1\)\}s · ` : ''/,
    '不许再保留「有值才拼时长、0 就隐藏」的旧回落'
  );
});

test('合成成片卡片与合成 viewer 的时长同样缺失即未知', () => {
  assert.match(
    novelList,
    /<small>\{formatDurationSeconds\(currentInfo\.duration\)\} · \{mergeJobSpeed\(current\)\.toFixed\(1\)\}x/,
    '合成成片卡片必须走 formatDurationSeconds'
  );
  assert.match(
    novelList,
    /\$\{formatDurationSeconds\(viewerInfo\.duration\)\} · \$\{mergeJobSpeed\(viewerMerge\)\.toFixed\(1\)\}x/,
    '合成 viewer meta 必须走 formatDurationSeconds'
  );
  assert.doesNotMatch(novelList, /currentInfo\.duration > 0 \?/, '合成卡片不许再隐藏未知时长');
  assert.doesNotMatch(novelList, /viewerInfo\.duration > 0 \?/, '合成 viewer 不许再隐藏未知时长');
});

test('主视频总时长只有全部主版本都读到真实时长才显示合计，否则未知', () => {
  assert.match(
    novelList,
    /const videoTotalDisplay = videoTotalIsExact \? `\$\{videoTotalSeconds\.toFixed\(1\)\}s` : '未知';/,
    '合计必须由 videoTotalIsExact 决定，缺失时显示未知'
  );
  assert.match(novelList, /<span>主视频总时长<\/span><b>\{videoTotalDisplay\}<\/b>/, '合计渲染必须使用 videoTotalDisplay');
  assert.doesNotMatch(novelList, /videoTotalSeconds\.toFixed\(2\)/, '不许再把计划值合计渲染成 ≈ 15.00s 样的数值');
  assert.doesNotMatch(novelList, /主视频总时长<\/span><b>\{videoTotalSeconds > 0 \?/, '不许再回落到 videoTotalSeconds 直接展示');
});
