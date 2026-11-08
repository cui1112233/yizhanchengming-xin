import assert from 'node:assert/strict';
import fs from 'node:fs';
import test from 'node:test';

const source = fs.readFileSync(new URL('./shuihuo/BatchFactoryNovelList.jsx', import.meta.url), 'utf8')
  + '\n' + fs.readFileSync(new URL('./shuihuo/batchFactoryNovelListHelpers.js', import.meta.url), 'utf8');

test('guards comic page generation against duplicate clicks for the same page', () => {
  assert.match(source, /const comicGenerationInFlightRef = useRef\(new Set\(\)\);/);
  assert.match(source, /comicGenerationInFlightRef\.current\.has\(requestKey\)/);
  assert.match(source, /comicGenerationInFlightRef\.current\.add\(requestKey\)/);
  assert.match(source, /comicGenerationInFlightRef\.current\.delete\(requestKey\)/);
});

test('任务/日志入口没有任何加载态，打开即用已有快照渲染', () => {
  assert.doesNotMatch(source, /loading=\{logsLoading\}/);
  assert.doesNotMatch(source, /setLogsOpen\(true\); loadRuntimeStatus\(\)/);
  assert.match(source, /data-testid="batch-log-sync-status"/);
  assert.match(source, /已加载 \{automationBooks\.length\} 本任务/);
  assert.match(source, /更早的日志加载失败或超时，已保留当前内容/);
  assert.match(source, /暂无执行日志/);
  assert.match(source, /function mergeAutomationLogSnapshot/);
});

test('refreshes comic page records after generation completes', () => {
  assert.match(source, /const \[comicPageRefreshNonce, setComicPageRefreshNonce\] = useState\(0\);/);
  assert.match(source, /setComicPageRefreshNonce\(value => value \+ 1\)/);
  assert.match(source, /refreshNonce = 0/);
  assert.match(source, /\[batchId, book\?\.id, refreshNonce\]/);
});
