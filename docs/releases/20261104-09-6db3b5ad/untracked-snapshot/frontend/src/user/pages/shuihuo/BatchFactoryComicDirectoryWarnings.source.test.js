import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const unifiedSettings = fs.readFileSync(path.join(here, 'BatchFactoryUnifiedSettingsModal.jsx'), 'utf8');
const publishHandler = fs.readFileSync(
  path.resolve(here, '../../../../../backend/internal/httpapi/batch_factory_n8_comic_publish.go'),
  'utf8'
);

test('comic-publish/status 在组织与音频分类读取失败时返回可展示的 warning', () => {
  assert.match(publishHandler, /organizationsWarning\s*string/);
  assert.match(publishHandler, /audioCategoriesWarning\s*string/);
  assert.match(publishHandler, /"organizationsWarning":\s*organizationsWarning/);
  assert.match(publishHandler, /"audioCategoriesWarning":\s*audioCategoriesWarning/);
  assert.match(publishHandler, /func \(p \*BatchFactoryN8ComicPublisher\) organizationsWithWarning/);
  assert.match(publishHandler, /func \(p \*BatchFactoryN8ComicPublisher\) audioCategoriesWithWarning/);
  assert.match(publishHandler, /N8 组织目录读取失败，请更新后台登录令牌后重试。/);
  assert.match(publishHandler, /N8 音频分类目录读取失败，请更新后台登录令牌后重试。/);
});

test('漫画发布设置读取 configured 与两个 warning，并在对应位置渲染', () => {
  assert.match(unifiedSettings, /setOrganizationsWarning\(String\(response\?\.organizationsWarning \|\| ''\)\)/);
  assert.match(unifiedSettings, /setAudioCategoriesWarning\(String\(response\?\.audioCategoriesWarning \|\| ''\)\)/);
  assert.match(unifiedSettings, /setDirectoryConfigured\(typeof response\?\.configured === 'boolean' \? response\.configured : null\)/);
  assert.match(unifiedSettings, /{organizationsWarning \? <small>{organizationsWarning}<\/small> : null}/);
  assert.match(unifiedSettings, /{audioCategoriesWarning \? <small>{audioCategoriesWarning}<\/small> : null}/);
  assert.match(unifiedSettings, /N8 目录服务未配置或令牌已失效/);
  // 有错误提示时不再同时显示“当前组织没有可用的制作者账号”的假空态文案。
  assert.match(unifiedSettings, /!producerAccountsLoading && !producerAccountsWarning && producerAccountOptions/);
});

test('指定音频读取失败为显式错误态并带重试，只有真空数据才显示没有可用', () => {
  assert.match(unifiedSettings, /const \[audiosError, setAudiosError\] = useState\(''\)/);
  assert.match(unifiedSettings, /setAudiosError\(error\?\.message \|\| '指定音频目录读取失败，请更新后台登录令牌后重试。'\)/);
  assert.match(unifiedSettings, /placeholder="请选择指定音频（可选）"/);
  assert.match(unifiedSettings, /{audiosError \? <small>{audiosError}<Button type="link" size="small" onClick=\{\(\) => setAudiosReloadKey\(key => key \+ 1\)\}>重试<\/Button><\/small> : null}/);
  assert.match(unifiedSettings, /{!audiosError && !audios\.length && comicPublishSettings\.organizationId && comicPublishSettings\.audioCategoryId \? <small>当前分类和组织下没有可用指定音频<\/small> : null}/);
});
