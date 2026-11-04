import assert from 'node:assert/strict';
import test from 'node:test';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const settingsSource = fs.readFileSync(path.join(here, 'SettingsPage.jsx'), 'utf8');
const apiConfigSource = fs.readFileSync(path.join(here, 'ApiConfigPage.jsx'), 'utf8');

test('一战晟铭执行器从设置页提供单 EXE Windows 安装包和 macOS 下载，而不是 API 配置', () => {
  assert.match(settingsSource, /一战晟铭执行器-安装版\.exe/);
  assert.match(settingsSource, /doubao-relay_1\.0\.0_universal\.dmg/);
  assert.match(settingsSource, /localExecutorRelease\.windows/);
  assert.match(settingsSource, /localExecutorRelease\.macos/);
  assert.match(settingsSource, /localExecutorRelease\.windowsUrl/);
  assert.doesNotMatch(settingsSource, /下载 Go Windows 执行器/);
  assert.doesNotMatch(settingsSource, /windows_x64\.zip/);
  assert.doesNotMatch(apiConfigSource, /doubao-relay-image/);
  assert.doesNotMatch(apiConfigSource, /doubao-relay-video/);
  assert.doesNotMatch(apiConfigSource, /local-doubao-executor-video/);
  assert.doesNotMatch(apiConfigSource, /本地豆包执行器/);
});

test('设置页逐台显示执行器名称、版本、在线状态和最后在线时间', () => {
  assert.match(settingsSource, /localExecutors\.map/);
  assert.match(settingsSource, /item\.name/);
  assert.match(settingsSource, /item\.version/);
  assert.match(settingsSource, /item\.online/);
  assert.match(settingsSource, /formatLastSeen\(item\.lastSeenAt\)/);
});

test('设置页通过共享豆包执行器状态读取设备和生成配对码', () => {
  assert.match(settingsSource, /useLocalExecutorStatus/);
  assert.match(settingsSource, /createLocalExecutorPairing/);
  assert.doesNotMatch(settingsSource, /apiRequest\('\/api\/shuihuo-production\/local-executors'/);
});

test('设置页的需配对状态消费后端状态接口字段，不再叠加本地配对码', () => {
  assert.match(settingsSource, /mergeExecutorState\(group\)/);
  assert.match(settingsSource, /executorOverview\?\.local\?\.pairing/);
  assert.match(settingsSource, /executorOverview\?\.giant\?\.pairing/);
  assert.match(settingsSource, /refreshExecutorOverview/);
  assert.doesNotMatch(settingsSource, /executorStatusText\(executorOverview\?\.local\?\.devices/);
});

test('设置页可以测试图片和视频并展示执行器返回结果', () => {
  assert.match(settingsSource, /测试图片生成/);
  assert.match(settingsSource, /测试视频生成/);
  assert.match(settingsSource, /local-executor-tests/);
  assert.match(settingsSource, /<img[^>]+executorTestResult/);
  assert.match(settingsSource, /<video[^>]+executorTestResult/);
  assert.match(settingsSource, /测试成功/);
});
